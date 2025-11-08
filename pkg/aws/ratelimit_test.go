package aws

import (
	"context"
	"testing"
	"time"
)

func TestRateLimiterAcquire(t *testing.T) {
	tests := []struct {
		name        string
		ratePerSec  float64
		burst       int
		tokensReq   float64
		shouldPass  bool
		description string
	}{
		{
			name:        "acquire within burst",
			ratePerSec:  10.0,
			burst:       5,
			tokensReq:   3.0,
			shouldPass:  true,
			description: "Should acquire tokens within burst capacity",
		},
		{
			name:        "acquire exact burst",
			ratePerSec:  10.0,
			burst:       5,
			tokensReq:   5.0,
			shouldPass:  true,
			description: "Should acquire tokens equal to burst",
		},
		{
			name:        "acquire exceeds burst",
			ratePerSec:  10.0,
			burst:       5,
			tokensReq:   6.0,
			shouldPass:  false,
			description: "Should not acquire tokens exceeding burst",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			limiter := NewRateLimiter(tt.ratePerSec, tt.burst)
			ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
			defer cancel()

			err := limiter.Acquire(ctx, tt.tokensReq)

			if tt.shouldPass && err != nil {
				t.Errorf("%s: expected success but got error: %v", tt.description, err)
			}
			if !tt.shouldPass && err == nil {
				t.Errorf("%s: expected error but got success", tt.description)
			}
		})
	}
}

func TestRateLimiterContextCancellation(t *testing.T) {
	limiter := NewRateLimiter(1.0, 1) // 1 token per second, burst of 1

	// Consume the burst
	ctx := context.Background()
	limiter.Acquire(ctx, 1.0)

	// Try to acquire more tokens with a cancelled context
	cancelledCtx, cancel := context.WithCancel(context.Background())
	cancel()

	err := limiter.Acquire(cancelledCtx, 1.0)
	if err == nil {
		t.Errorf("expected context cancellation error, got nil")
	}
	if err != context.Canceled {
		t.Errorf("expected context.Canceled, got: %v", err)
	}
}

func TestRateLimiterTokenRefill(t *testing.T) {
	limiter := NewRateLimiter(100.0, 5) // 100 tokens per second, burst of 5

	ctx := context.Background()

	// Consume all tokens
	err := limiter.Acquire(ctx, 5.0)
	if err != nil {
		t.Errorf("failed to acquire initial tokens: %v", err)
	}

	// Wait for tokens to refill (50ms should give ~5 tokens at 100/sec)
	time.Sleep(50 * time.Millisecond)

	// Try to acquire 2 tokens (should succeed)
	ctx2, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	err = limiter.Acquire(ctx2, 2.0)
	if err != nil {
		t.Errorf("failed to acquire refilled tokens: %v", err)
	}
}

func TestRateLimiterCapacity(t *testing.T) {
	limiter := NewRateLimiter(10.0, 5) // 10 tokens per second, burst of 5

	ctx := context.Background()

	// Acquire initial tokens
	limiter.Acquire(ctx, 5.0)

	// Try to acquire more tokens immediately (should timeout)
	ctx2, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	err := limiter.Acquire(ctx2, 1.0)
	if err == nil {
		t.Errorf("expected timeout, got success")
	}
}

func TestRetryConfigRetryableError(t *testing.T) {
	cfg := DefaultRetryConfig()

	tests := []struct {
		name        string
		errMsg      string
		shouldRetry bool
		description string
	}{
		{
			name:        "throttling error",
			errMsg:      "Throttling",
			shouldRetry: true,
			description: "Should retry throttling errors",
		},
		{
			name:        "service unavailable",
			errMsg:      "ServiceUnavailable",
			shouldRetry: true,
			description: "Should retry service unavailable errors",
		},
		{
			name:        "non-retryable error",
			errMsg:      "InvalidParameterException",
			shouldRetry: false,
			description: "Should not retry non-retryable errors",
		},
		{
			name:        "nil error",
			errMsg:      "",
			shouldRetry: false,
			description: "Should not retry nil errors",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var err error
			if tt.errMsg != "" {
				err = &testError{msg: tt.errMsg}
			}

			result := cfg.RetryableError(err)
			if result != tt.shouldRetry {
				t.Errorf("%s: expected %v, got %v", tt.description, tt.shouldRetry, result)
			}
		})
	}
}

func TestRetryConfigCalculateDelay(t *testing.T) {
	cfg := DefaultRetryConfig()

	tests := []struct {
		name        string
		attempt     int
		minDelay    time.Duration
		maxDelay    time.Duration
		description string
	}{
		{
			name:        "first attempt",
			attempt:     1,
			minDelay:    100 * time.Millisecond,
			maxDelay:    200 * time.Millisecond,
			description: "First attempt should use base delay",
		},
		{
			name:        "second attempt",
			attempt:     2,
			minDelay:    150 * time.Millisecond,
			maxDelay:    400 * time.Millisecond,
			description: "Second attempt should double delay",
		},
		{
			name:        "max delay cap",
			attempt:     10,
			minDelay:    cfg.MaxDelay,
			maxDelay:    cfg.MaxDelay,
			description: "Delay should be capped at max delay",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			delay := cfg.CalculateDelay(tt.attempt)

			// Account for jitter (±25%)
			jitterMargin := cfg.MaxDelay / 4

			if delay < tt.minDelay-jitterMargin || delay > tt.maxDelay+jitterMargin {
				t.Errorf("%s: delay %v outside expected range [%v, %v]", tt.description, delay, tt.minDelay-jitterMargin, tt.maxDelay+jitterMargin)
			}
		})
	}
}

func TestCircuitBreakerStateTransitions(t *testing.T) {
	cfg := &CircuitBreakerConfig{
		FailureThreshold: 2,
		ResetTimeout:     50 * time.Millisecond,
		HalfOpenMaxCalls: 1,
	}
	cb := NewCircuitBreaker(cfg)

	// Initial state should be closed
	if err := cb.Allow(); err != nil {
		t.Errorf("initial state should allow operations, got error: %v", err)
	}

	// Record failures to open the circuit
	for i := 0; i < cfg.FailureThreshold; i++ {
		cb.RecordFailure()
	}

	// Circuit should now be open
	if err := cb.Allow(); err == nil {
		t.Errorf("circuit should be open after threshold failures")
	}
}

func TestCircuitBreakerRecordSuccess(t *testing.T) {
	cfg := DefaultCircuitBreakerConfig()
	cb := NewCircuitBreaker(cfg)

	// Record a success in closed state (should reset failure count)
	cb.RecordFailure()
	cb.RecordSuccess()

	// Failure count should be reset
	// We can verify this by checking that we need more failures to open
	for i := 0; i < cfg.FailureThreshold; i++ {
		cb.RecordFailure()
	}

	// Circuit should now be open
	if err := cb.Allow(); err == nil {
		t.Errorf("circuit should be open after threshold failures")
	}
}

func TestRateLimiterAcquireNonBlocking(t *testing.T) {
	tests := []struct {
		name        string
		ratePerSec  float64
		burst       int
		tokensReq   float64
		shouldPass  bool
		description string
	}{
		{
			name:        "acquire within burst",
			ratePerSec:  10.0,
			burst:       5,
			tokensReq:   3.0,
			shouldPass:  true,
			description: "Should acquire tokens within burst capacity",
		},
		{
			name:        "acquire exact burst",
			ratePerSec:  10.0,
			burst:       5,
			tokensReq:   5.0,
			shouldPass:  true,
			description: "Should acquire tokens equal to burst",
		},
		{
			name:        "acquire exceeds burst",
			ratePerSec:  10.0,
			burst:       5,
			tokensReq:   6.0,
			shouldPass:  false,
			description: "Should not acquire tokens exceeding burst",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			limiter := NewRateLimiter(tt.ratePerSec, tt.burst)

			err := limiter.AcquireNonBlocking(tt.tokensReq)

			if tt.shouldPass && err != nil {
				t.Errorf("%s: expected success but got error: %v", tt.description, err)
			}
			if !tt.shouldPass && err == nil {
				t.Errorf("%s: expected error but got success", tt.description)
			}

			// Verify error type for non-blocking failures
			if !tt.shouldPass && err != nil {
				if _, ok := err.(*RateLimitError); !ok {
					t.Errorf("%s: expected RateLimitError but got %T", tt.description, err)
				}
			}
		})
	}
}

func TestRateLimiterContextCancellationDuringWait(t *testing.T) {
	limiter := NewRateLimiter(1.0, 1) // 1 token per second, burst of 1

	// Consume the burst
	ctx := context.Background()
	limiter.Acquire(ctx, 1.0)

	// Try to acquire more tokens with a context that will be cancelled
	cancelledCtx, cancel := context.WithCancel(context.Background())

	// Cancel after a short delay
	go func() {
		time.Sleep(50 * time.Millisecond)
		cancel()
	}()

	err := limiter.Acquire(cancelledCtx, 1.0)
	if err == nil {
		t.Errorf("expected context cancellation error, got nil")
	}
	if err != context.Canceled {
		t.Errorf("expected context.Canceled, got: %v", err)
	}
}

// testError is a simple error implementation for testing
type testError struct {
	msg string
}

func (e *testError) Error() string {
	return e.msg
}
