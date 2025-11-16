package aws

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

func TestResolveSingleIdentifierIncludeStoppedStates(t *testing.T) {
	var captured []types.Filter
	c := &Client{}
	c.describeInstancesHook = func(_ context.Context, f []types.Filter) ([]Instance, error) {
		captured = f
		return []Instance{{InstanceID: "i-1", State: "stopped"}}, nil
	}
	_, err := c.resolveSingleIdentifier(context.Background(), "Name:web", true)
	if err != nil {
		t.Fatalf("err")
	}
	sf := findFilterByName(t, captured, "instance-state-name")
	have := map[string]bool{}
	for _, v := range sf.Values {
		have[v] = true
	}
	for _, s := range []string{"running", "stopped", "stopping", "pending"} {
		if !have[s] {
			t.Fatalf("state")
		}
	}
}
