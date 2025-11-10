//go:build unit

package aws

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"testing"
)

// mockDescribeClient minimal interface wrapper
type mockDescribeClient struct{ pages [][]Instance }

func (m *mockDescribeClient) nextPage() []Instance {
	if len(m.pages) == 0 {
		return nil
	}
	p := m.pages[0]
	m.pages = m.pages[1:]
	return p
}

// mockClient embeds Client but replaces describeInstances behavior
func newTestClient(pages [][]Instance) *Client {
	c := &Client{CircuitBreaker: NewCircuitBreaker(DefaultCircuitBreakerConfig())}
	c.describeInstancesHook = func(ctx context.Context, _ []types.Filter) ([]Instance, error) {
		var all []Instance
		for _, p := range pages {
			all = append(all, p...)
		}
		return all, nil
	}
	return c
}

func TestUnitResolveSingleSuccess(t *testing.T) {
	inst := Instance{InstanceID: "i-123", State: "running", Name: "web"}
	c := newTestClient([][]Instance{{inst}})
	got, err := c.ResolveSingleInstance(context.Background(), "i-123")
	if err != nil || got.InstanceID != "i-123" {
		t.Fatalf("expected instance, got=%v err=%v", got, err)
	}
}

func TestUnitResolveSingleMultiple(t *testing.T) {
	insts := []Instance{{InstanceID: "i-1", State: "running"}, {InstanceID: "i-2", State: "running"}}
	c := newTestClient([][]Instance{{insts[0], insts[1]}})
	_, err := c.ResolveSingleInstance(context.Background(), "i-1")
	if err == nil {
		t.Fatalf("expected multiple instances error")
	}
}

func TestUnitResolveSingleNotRunning(t *testing.T) {
	inst := Instance{InstanceID: "i-3", State: "stopped"}
	c := newTestClient([][]Instance{{inst}})
	_, err := c.ResolveSingleInstance(context.Background(), "i-3")
	if err == nil {
		t.Fatalf("expected not running error")
	}
}
