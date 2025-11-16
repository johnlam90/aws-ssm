package aws

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

func TestResolveSingleIdentifierNoMatchError(t *testing.T) {
	c := &Client{}
	c.describeInstancesHook = func(_ context.Context, _ []types.Filter) ([]Instance, error) {
		return nil, nil
	}
	_, err := c.resolveSingleIdentifier(context.Background(), "Name:web", false)
	if err == nil {
		t.Fatalf("expected error")
	}
}
