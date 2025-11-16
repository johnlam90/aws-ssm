package aws

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

func TestListInstancesAppliesTagAndRunningFilters(t *testing.T) {
	var captured []types.Filter
	c := &Client{}
	c.describeInstancesHook = func(_ context.Context, f []types.Filter) ([]Instance, error) {
		captured = f
		return []Instance{{InstanceID: "i-1", State: "running"}}, nil
	}
	_, err := c.ListInstances(context.Background(), map[string]string{"Env": "prod", "Team": "ops"})
	if err != nil {
		t.Fatalf("err")
	}
	if !containsFilter(captured, "tag:Env") || !containsFilter(captured, "tag:Team") {
		t.Fatalf("tags")
	}
	sf := findFilterByName(t, captured, "instance-state-name")
	if len(sf.Values) != 1 || sf.Values[0] != "running" {
		t.Fatalf("state")
	}
}
