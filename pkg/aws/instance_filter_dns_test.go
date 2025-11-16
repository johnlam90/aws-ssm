package aws

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

func TestFindInstancesWithStatesFallsBackForDNSLookups(t *testing.T) {
	t.Parallel()
	var calls int
	var privateSeen, publicSeen bool
	client := &Client{
		describeInstancesHook: func(_ context.Context, filters []types.Filter) ([]Instance, error) {
			calls++
			sf := findFilterByName(t, filters, "instance-state-name")
			if len(sf.Values) != 1 || sf.Values[0] != "running" {
				t.Fatalf("state")
			}
			if containsFilter(filters, "private-dns-name") {
				privateSeen = true
				return nil, nil
			}
			if containsFilter(filters, "dns-name") {
				publicSeen = true
				return []Instance{{InstanceID: "i-abcdef", State: "running"}}, nil
			}
			t.Fatalf("unexpected filters: %#v", filters)
			return nil, nil
		},
	}
	result, err := client.FindInstancesWithStates(context.Background(), "ip-10-0-0-1.ec2.internal", []string{"running"})
	if err != nil {
		t.Fatalf("err")
	}
	if len(result) != 1 {
		t.Fatalf("result")
	}
	if calls != 2 || !privateSeen || !publicSeen {
		t.Fatalf("fallback")
	}
}
