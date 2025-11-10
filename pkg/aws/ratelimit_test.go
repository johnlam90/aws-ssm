package aws

import (
	"context"
	"testing"
)

func TestRateLimiterAcquire(t *testing.T) {
	rl := NewRateLimiter(10, 2)
	ctx := context.Background()
	if err := rl.Acquire(ctx, 1); err != nil {
		t.Fatalf("unexpected acquire err: %v", err)
	}
	if err := rl.Acquire(ctx, 1); err != nil {
		t.Fatalf("second acquire err: %v", err)
	}
}

func TestRateLimiterNonBlocking(t *testing.T) {
	rl := NewRateLimiter(1, 1)
	if err := rl.AcquireNonBlocking(1); err != nil {
		t.Fatalf("expected success: %v", err)
	}
	if err := rl.AcquireNonBlocking(1); err == nil {
		t.Fatalf("expected failure when tokens exhausted")
	}
}

func TestRetryMetrics(t *testing.T) {
	rm := NewRetryMetrics()
	rm.Success()
	rm.Failure()
	rm.Failure()
	total, succ, fail := rm.GetStats()
	if total != 3 || succ != 1 || fail != 2 {
		t.Fatalf("unexpected stats: %d %d %d", total, succ, fail)
	}
}
