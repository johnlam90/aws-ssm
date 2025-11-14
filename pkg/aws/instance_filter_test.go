package aws

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

func TestFindInstancesDefaultsToRunningState(t *testing.T) {
	t.Parallel()

	var captured []types.Filter

	client := &Client{
		describeInstancesHook: func(_ context.Context, filters []types.Filter) ([]Instance, error) {
			captured = filters
			return []Instance{
				{InstanceID: "i-1234567890", State: "running"},
			}, nil
		},
	}

	if _, err := client.FindInstances(context.Background(), "web-server"); err != nil {
		t.Fatalf("FindInstances returned error: %v", err)
	}

	stateFilter := findFilterByName(t, captured, "instance-state-name")
	if len(stateFilter.Values) != 1 || stateFilter.Values[0] != "running" {
		t.Fatalf("expected running state filter, got %#v", stateFilter.Values)
	}
}

func TestFindInstancesWithStatesAllowsCustomStates(t *testing.T) {
	t.Parallel()

	var captured []types.Filter

	client := &Client{
		describeInstancesHook: func(_ context.Context, filters []types.Filter) ([]Instance, error) {
			captured = filters
			return []Instance{
				{InstanceID: "i-0987654321", State: "stopped"},
			}, nil
		},
	}

	customStates := []string{"running", "stopped"}
	if _, err := client.FindInstancesWithStates(context.Background(), "web-server", customStates); err != nil {
		t.Fatalf("FindInstancesWithStates returned error: %v", err)
	}

	stateFilter := findFilterByName(t, captured, "instance-state-name")
	if len(stateFilter.Values) != len(customStates) {
		t.Fatalf("expected %d state values, got %d", len(customStates), len(stateFilter.Values))
	}
	for i, state := range customStates {
		if stateFilter.Values[i] != state {
			t.Fatalf("state filter mismatch at position %d: got %s want %s", i, stateFilter.Values[i], state)
		}
	}
}

func TestFindInstancesWithStatesFallsBackForIPLookups(t *testing.T) {
	t.Parallel()

	var (
		calls       int
		privateSeen bool
		publicSeen  bool
	)

	client := &Client{
		describeInstancesHook: func(_ context.Context, filters []types.Filter) ([]Instance, error) {
			calls++
			stateFilter := findFilterByName(t, filters, "instance-state-name")
			if len(stateFilter.Values) != 1 || stateFilter.Values[0] != "running" {
				t.Fatalf("expected running state filter for IP lookup, got %#v", stateFilter.Values)
			}

			if containsFilter(filters, "private-ip-address") {
				privateSeen = true
				return nil, nil
			}
			if containsFilter(filters, "ip-address") {
				publicSeen = true
				return []Instance{{InstanceID: "i-abcdef", State: "running"}}, nil
			}

			t.Fatalf("unexpected filters: %#v", filters)
			return nil, nil
		},
	}

	result, err := client.FindInstancesWithStates(context.Background(), "10.0.0.1", []string{"running"})
	if err != nil {
		t.Fatalf("FindInstancesWithStates returned error: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("expected result from public IP fallback, got %v", result)
	}
	if calls != 2 || !privateSeen || !publicSeen {
		t.Fatalf("expected private and public lookups (calls=%d private=%v public=%v)", calls, privateSeen, publicSeen)
	}
}

func containsFilter(filters []types.Filter, name string) bool {
	for _, f := range filters {
		if f.Name != nil && *f.Name == name {
			return true
		}
	}
	return false
}

func findFilterByName(t *testing.T, filters []types.Filter, name string) types.Filter {
	t.Helper()

	for _, f := range filters {
		if f.Name != nil && *f.Name == name {
			return f
		}
	}

	t.Fatalf("filter %s not found", name)
	return types.Filter{}
}
