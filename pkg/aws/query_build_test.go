package aws

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

func TestAddStateFilter(t *testing.T) {
	f1 := addStateFilter(false)
	if len(f1) != 1 || len(f1[0].Values) != 1 || f1[0].Values[0] != "running" {
		t.Fatalf("state")
	}
	f2 := addStateFilter(true)
	have := map[string]bool{}
	for _, v := range f2[0].Values {
		have[v] = true
	}
	for _, s := range []string{"running", "stopped", "stopping", "pending"} {
		if !have[s] {
			t.Fatalf("states")
		}
	}
}

func TestAddTagFilters(t *testing.T) {
	filters := addTagFilters(nil, []string{"Env:prod", "Team:ops"})
	if len(filters) != 2 {
		t.Fatalf("len")
	}
}

func TestBuildInstanceQueryInstanceIDs(t *testing.T) {
	c := &Client{}
	input, err := c.buildInstanceQuery(context.Background(), InterfacesOptions{InstanceIDs: []string{"i-1", "i-2"}})
	if err != nil {
		t.Fatalf("err")
	}
	if len(input.InstanceIds) != 2 {
		t.Fatalf("ids")
	}
}

func TestResolveSingleIdentifier(t *testing.T) {
	c := &Client{}
	c.describeInstancesHook = func(_ context.Context, _ []types.Filter) ([]Instance, error) {
		return []Instance{{InstanceID: "i-11", State: "running"}, {InstanceID: "i-22", State: "running"}}, nil
	}
	input, err := c.resolveSingleIdentifier(context.Background(), "Name:web", false)
	if err != nil {
		t.Fatalf("err")
	}
	if len(input.InstanceIds) != 2 {
		t.Fatalf("ids")
	}
}

func TestCollectInstanceInterfacesTerminatesSkipped(t *testing.T) {
	c := &Client{}
	out := &ec2.DescribeInstancesOutput{Reservations: []types.Reservation{{Instances: []types.Instance{{State: &types.InstanceState{Name: types.InstanceStateNameTerminated}}}}}}
	res, err := c.collectInstanceInterfaces(context.Background(), out)
	if err != nil {
		t.Fatalf("err")
	}
	if res != nil {
		t.Fatalf("nil")
	}
}
