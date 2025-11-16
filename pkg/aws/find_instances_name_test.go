package aws

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

func TestFindInstancesUsesNameTagFilter(t *testing.T) {
	var captured []types.Filter
	client := &Client{
		describeInstancesHook: func(_ context.Context, filters []types.Filter) ([]Instance, error) {
			captured = filters
			return []Instance{{InstanceID: "i-1", State: "running"}}, nil
		},
	}
	_, err := client.FindInstances(context.Background(), "web-server")
	if err != nil {
		t.Fatalf("err")
	}
	if !containsFilter(captured, "tag:Name") {
		t.Fatalf("name filter")
	}
}
