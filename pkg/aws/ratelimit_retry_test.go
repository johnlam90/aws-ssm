package aws

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"
)

func TestRetryConfig_RetryableError(t *testing.T) {
	cfg := DefaultRetryConfig()

	tests := []struct {
		name      string
		err       error
		retryable bool
	}{
		{"Nil error", nil, false},
		{"Throttling error", errors.New("Throttling: Rate exceeded"), true},
		{"ServiceUnavailable error", errors.New("ServiceUnavailable: Server overload"), true},
		{"Non-retryable error", errors.New("InvalidParameter: Bad input"), false},
		{"Wrapped retryable error", fmt.Errorf("wrapped: %w", errors.New("RequestLimitExceeded")), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := cfg.RetryableError(tt.err); got != tt.retryable {
				t.Errorf("RetryableError() = %v, want %v", got, tt.retryable)
			}
		})
	}
}

func TestRetryConfig_CalculateDelay(t *testing.T) {
	cfg := &RetryConfig{
		BaseDelay:         100 * time.Millisecond,
		MaxDelay:          1 * time.Second,
		BackoffMultiplier: 2.0,
		Jitter:            false, // Disable jitter for deterministic testing
	}

	tests := []struct {
		attempt int
		want    time.Duration
	}{
		{1, 100 * time.Millisecond},
		{2, 200 * time.Millisecond},
		{3, 400 * time.Millisecond},
		{4, 800 * time.Millisecond},
		{5, 1 * time.Second}, // Capped at MaxDelay
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("Attempt %d", tt.attempt), func(t *testing.T) {
			if got := cfg.CalculateDelay(tt.attempt); got != tt.want {
				t.Errorf("CalculateDelay() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRetryOperation_Execute(t *testing.T) {
	cfg := &RetryConfig{
		MaxAttempts:       3,
		BaseDelay:         1 * time.Millisecond,
		MaxDelay:          10 * time.Millisecond,
		BackoffMultiplier: 2.0,
		Jitter:            false,
		RetryableErrors:   []string{"RetryMe"},
	}

	t.Run("Success on first attempt", func(t *testing.T) {
		op := NewRetryOperation(func() error {
			return nil
		}, cfg)

		if err := op.Execute(context.Background()); err != nil {
			t.Errorf("Execute() error = %v", err)
		}
		if total, success, _ := op.GetMetrics().GetStats(); total != 1 || success != 1 {
			t.Errorf("Metrics mismatch: total=%d, success=%d", total, success)
		}
	})

	t.Run("Success after retry", func(t *testing.T) {
		attempts := 0
		op := NewRetryOperation(func() error {
			attempts++
			if attempts < 2 {
				return errors.New("RetryMe: temporary failure")
			}
			return nil
		}, cfg)

		if err := op.Execute(context.Background()); err != nil {
			t.Errorf("Execute() error = %v", err)
		}
		if attempts != 2 {
			t.Errorf("Expected 2 attempts, got %d", attempts)
		}
		if total, success, failures := op.GetMetrics().GetStats(); total != 2 || success != 1 || failures != 1 {
			t.Errorf("Metrics mismatch: total=%d, success=%d, failures=%d", total, success, failures)
		}
	})

	t.Run("Failure after max attempts", func(t *testing.T) {
		attempts := 0
		op := NewRetryOperation(func() error {
			attempts++
			return errors.New("RetryMe: persistent failure")
		}, cfg)

		err := op.Execute(context.Background())
		if err == nil {
			t.Error("Expected error, got nil")
		}
		if attempts != 3 {
			t.Errorf("Expected 3 attempts, got %d", attempts)
		}
	})

	t.Run("Non-retryable error", func(t *testing.T) {
		attempts := 0
		op := NewRetryOperation(func() error {
			attempts++
			return errors.New("FatalError")
		}, cfg)

		err := op.Execute(context.Background())
		if err == nil {
			t.Error("Expected error, got nil")
		}
		if attempts != 1 {
			t.Errorf("Expected 1 attempt, got %d", attempts)
		}
	})

	t.Run("Context cancelled", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		op := NewRetryOperation(func() error {
			return nil
		}, cfg)

		err := op.Execute(ctx)
		if err == nil {
			t.Error("Expected error, got nil")
		}
	})
}
