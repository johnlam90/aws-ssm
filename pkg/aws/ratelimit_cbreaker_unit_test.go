//go:build unit

package aws

import (
	"testing"
	"time"
	"github.com/johnlam90/aws-ssm/pkg/logging"
)

type testClock struct{ t time.Time }
func (c *testClock) Now() time.Time { return c.t }
func (c *testClock) Advance(d time.Duration) { c.t = c.t.Add(d) }
type discardWriter struct{}
func (discardWriter) Write(p []byte) (int, error) { return len(p), nil }

func TestUnitCircuitBreakerTransitions(t *testing.T) {
	logging.Init(logging.WithOutput(discardWriter{}), logging.WithLevel("error"))
	clk := &testClock{t: time.Unix(0, 0)}
	cfg := &CircuitBreakerConfig{FailureThreshold: 2, ResetTimeout: 10 * time.Millisecond, HalfOpenMaxCalls: 2}
	cb := NewCircuitBreakerWithClock(cfg, clk)
	cb.RecordFailure()
	cb.RecordFailure()
	if cb.GetState() != CircuitOpen {
		t.Fatalf("expected open")
	}
	clk.Advance(11 * time.Millisecond)
	if err := cb.Allow(); err != nil {
		t.Fatalf("allow error: %v", err)
	}
	if cb.GetState() != CircuitHalfOpen {
		t.Fatalf("expected half-open")
	}
	cb.RecordSuccess()
	cb.RecordSuccess()
	if cb.GetState() != CircuitClosed {
		t.Fatalf("expected closed")
	}
}

func TestUnitCircuitBreakerReopen(t *testing.T) {
	logging.Init(logging.WithOutput(discardWriter{}), logging.WithLevel("error"))
	clk := &testClock{t: time.Unix(0, 0)}
	cfg := &CircuitBreakerConfig{FailureThreshold: 2, ResetTimeout: 5 * time.Millisecond, HalfOpenMaxCalls: 1}
	cb := NewCircuitBreakerWithClock(cfg, clk)
	cb.RecordFailure()
	cb.RecordFailure()
	clk.Advance(6 * time.Millisecond)
	_ = cb.Allow()
	if cb.GetState() != CircuitHalfOpen {
		t.Fatalf("expected half-open")
	}
	cb.RecordFailure()
	if cb.GetState() != CircuitOpen {
		t.Fatalf("expected reopen")
	}
}
