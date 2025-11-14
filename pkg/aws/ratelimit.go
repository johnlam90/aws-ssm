package aws

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"strings"
	"sync"
	"time"

	"github.com/johnlam90/aws-ssm/pkg/logging"
)

// Note: math/rand is automatically seeded in Go 1.20+ with a random value on first use.
// No explicit seeding is required.

// RateLimiter implements token bucket rate limiting
type RateLimiter struct {
	tokens     float64
	capacity   float64
	refill     float64
	lastRefill time.Time
	mu         sync.Mutex
	logger     logging.Logger
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(ratePerSecond float64, burst int) *RateLimiter {
	return &RateLimiter{
		tokens:     float64(burst),
		capacity:   float64(burst),
		refill:     ratePerSecond,
		lastRefill: time.Now(),
		logger:     logging.With(logging.String("component", "ratelimiter")),
	}
}

// Acquire attempts to acquire tokens for a request, waiting if necessary
// Returns an error if the context is cancelled before tokens are available
func (r *RateLimiter) Acquire(ctx context.Context, tokens float64) error {
	for {
		r.mu.Lock()

		// Calculate tokens to add based on time passed
		now := time.Now()
		elapsed := now.Sub(r.lastRefill)
		tokensToAdd := elapsed.Seconds() * r.refill

		// Don't let tokens exceed capacity
		r.tokens = math.Min(r.tokens+tokensToAdd, r.capacity)
		r.lastRefill = now

		if r.tokens >= tokens {
			r.tokens -= tokens
			r.logger.Debug("Token acquired", logging.Float64("tokens", tokens), logging.Float64("remaining", r.tokens))
			r.mu.Unlock()
			return nil
		}

		// Calculate wait time until tokens are available
		waitTime := time.Duration((tokens - r.tokens) / r.rate() * float64(time.Second))

		r.logger.Debug("Rate limit: waiting for tokens",
			logging.Float64("requested", tokens),
			logging.Float64("available", r.tokens),
			logging.Duration("wait_time", waitTime))

		r.mu.Unlock()

		// Wait for either the wait time to elapse or context to be cancelled
		select {
		case <-time.After(waitTime):
			// Try again after waiting
			continue
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

// AcquireNonBlocking attempts to acquire tokens without blocking
// Returns RateLimitError if tokens are not immediately available
func (r *RateLimiter) AcquireNonBlocking(tokens float64) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Calculate tokens to add based on time passed
	now := time.Now()
	elapsed := now.Sub(r.lastRefill)
	tokensToAdd := elapsed.Seconds() * r.refill

	// Don't let tokens exceed capacity
	r.tokens = math.Min(r.tokens+tokensToAdd, r.capacity)
	r.lastRefill = now

	if r.tokens >= tokens {
		r.tokens -= tokens
		r.logger.Debug("Token acquired (non-blocking)", logging.Float64("tokens", tokens), logging.Float64("remaining", r.tokens))
		return nil
	}

	// Calculate wait time until tokens are available
	waitTime := time.Duration((tokens - r.tokens) / r.rate() * float64(time.Second))

	r.logger.Debug("Rate limit: insufficient tokens (non-blocking)",
		logging.Float64("requested", tokens),
		logging.Float64("available", r.tokens),
		logging.Duration("wait_time", waitTime))

	return &RateLimitError{
		WaitTime: waitTime,
		Reason:   fmt.Sprintf("insufficient tokens: have %.2f, need %.2f", r.tokens, tokens),
	}
}

// rate returns the current refill rate
func (r *RateLimiter) rate() float64 {
	return r.refill
}

// RateLimitError indicates rate limiting
type RateLimitError struct {
	WaitTime time.Duration
	Reason   string
}

func (e *RateLimitError) Error() string {
	return fmt.Sprintf("rate limit: %s, wait %v", e.Reason, e.WaitTime)
}

// RetryConfig holds retry configuration
type RetryConfig struct {
	MaxAttempts       int
	BaseDelay         time.Duration
	MaxDelay          time.Duration
	BackoffMultiplier float64
	Jitter            bool
	RetryableErrors   []string
}

// DefaultRetryConfig returns default retry configuration
func DefaultRetryConfig() *RetryConfig {
	return &RetryConfig{
		MaxAttempts:       3,
		BaseDelay:         100 * time.Millisecond,
		MaxDelay:          30 * time.Second,
		BackoffMultiplier: 2.0,
		Jitter:            true,
		RetryableErrors: []string{
			"Throttling",
			"ThrottlingException",
			"RequestLimitExceeded",
			"ServiceUnavailable",
			"RequestTimeout",
			"RequestTimeoutException",
		},
	}
}

// RetryableError checks if an error is retryable
func (cfg *RetryConfig) RetryableError(err error) bool {
	if err == nil {
		return false
	}

	errMsg := err.Error()
	for _, retryableErr := range cfg.RetryableErrors {
		if strings.Contains(errMsg, retryableErr) {
			return true
		}
	}

	// Check for AWS API error codes
	if awsErr, ok := err.(interface{ ErrorCode() string }); ok {
		for _, retryableErr := range cfg.RetryableErrors {
			if awsErr.ErrorCode() == retryableErr {
				return true
			}
		}
	}

	return false
}

// CalculateDelay calculates the delay for attempt number
func (cfg *RetryConfig) CalculateDelay(attempt int) time.Duration {
	// Exponential backoff: baseDelay * (multiplier ^ attempt)
	delay := time.Duration(float64(cfg.BaseDelay) * math.Pow(cfg.BackoffMultiplier, float64(attempt-1)))

	// Cap at max delay
	if delay > cfg.MaxDelay {
		delay = cfg.MaxDelay
	}

	// Add jitter to avoid thundering herd
	if cfg.Jitter {
		// Add ±25% jitter using math/rand (sufficient for rate limiting)
		jitterRange := int64(delay / 4)
		randomJitter := rand.Int63n(jitterRange*2) - jitterRange
		delay += time.Duration(randomJitter)
	}

	return delay
}

// RetryOperation retries an operation with exponential backoff
type RetryOperation struct {
	operation func() error
	config    *RetryConfig
	logger    logging.Logger
	metrics   *RetryMetrics
}

// NewRetryOperation creates a new retry operation
func NewRetryOperation(op func() error, cfg *RetryConfig) *RetryOperation {
	return &RetryOperation{
		operation: op,
		config:    cfg,
		logger:    logging.With(logging.String("component", "retry")),
		metrics:   &RetryMetrics{},
	}
}

// Execute executes the operation with retries
func (ro *RetryOperation) Execute(ctx context.Context) (err error) {
	for attempt := 1; attempt <= ro.config.MaxAttempts; attempt++ {
		select {
		case <-ctx.Done():
			return fmt.Errorf("operation cancelled: %w", ctx.Err())
		default:
		}

		ro.logger.Debug("Executing operation attempt",
			logging.Int("attempt", attempt),
			logging.Int("max_attempts", ro.config.MaxAttempts))

		err = ro.operation()
		if err == nil {
			ro.metrics.Success()
			ro.logger.Debug("Operation succeeded", logging.Int("attempts", attempt))
			return nil
		}

		ro.metrics.Failure()

		// Check if error is retryable
		if !ro.config.RetryableError(err) {
			ro.logger.Error("Operation failed with non-retryable error",
				logging.String("error", err.Error()),
				logging.Int("attempt", attempt))
			return err
		}

		// Check if this was the last attempt
		if attempt == ro.config.MaxAttempts {
			ro.logger.Error("Operation failed after all retry attempts",
				logging.String("error", err.Error()),
				logging.Int("attempts", attempt))
			return fmt.Errorf("operation failed after %d attempts: %w", attempt, err)
		}

		// Calculate delay
		delay := ro.config.CalculateDelay(attempt)

		ro.logger.Warn("Operation failed, retrying",
			logging.String("error", err.Error()),
			logging.Int("attempt", attempt),
			logging.Duration("delay", delay))

		// Wait with context cancellation
		select {
		case <-ctx.Done():
			return fmt.Errorf("operation cancelled during retry delay: %w", ctx.Err())
		case <-time.After(delay):
		}
	}

	return fmt.Errorf("exhausted all retry attempts")
}

// GetMetrics returns retry operation metrics
func (ro *RetryOperation) GetMetrics() *RetryMetrics {
	return ro.metrics
}

// RetryMetrics tracks retry operation statistics
type RetryMetrics struct {
	totalAttempts int
	successes     int
	failures      int
	mu            sync.Mutex
}

// NewRetryMetrics creates a new retry metrics instance
func NewRetryMetrics() *RetryMetrics {
	return &RetryMetrics{}
}

// Success records a successful retry operation
func (rm *RetryMetrics) Success() {
	rm.mu.Lock()
	defer rm.mu.Unlock()
	rm.totalAttempts++
	rm.successes++
}

// Failure records a failed retry operation
func (rm *RetryMetrics) Failure() {
	rm.mu.Lock()
	defer rm.mu.Unlock()
	rm.totalAttempts++
	rm.failures++
}

// GetStats returns retry metrics statistics
func (rm *RetryMetrics) GetStats() (total, successes, failures int) {
	rm.mu.Lock()
	defer rm.mu.Unlock()
	return rm.totalAttempts, rm.successes, rm.failures
}

// CircuitBreakerState represents circuit breaker states
type CircuitBreakerState int

const (
	// CircuitClosed represents the closed state of circuit breaker
	CircuitClosed CircuitBreakerState = iota
	// CircuitOpen represents the open state of circuit breaker
	CircuitOpen
	// CircuitHalfOpen represents the half-open state of circuit breaker
	CircuitHalfOpen
)

func (s CircuitBreakerState) String() string {
	switch s {
	case CircuitClosed:
		return "closed"
	case CircuitOpen:
		return "open"
	case CircuitHalfOpen:
		return "half_open"
	default:
		return "unknown"
	}
}

// CircuitBreaker implements circuit breaker pattern
type CircuitBreaker struct {
	state             CircuitBreakerState
	failureCount      int
	successCount      int
	lastFailureTime   time.Time
	halfOpenMaxCalls  int
	halfOpenCallCount int
	failureThreshold  int
	resetTimeout      time.Duration
	mu                sync.RWMutex
	logger            logging.Logger
	clock             Clock
}

// CircuitBreakerConfig holds circuit breaker configuration
type CircuitBreakerConfig struct {
	FailureThreshold int
	ResetTimeout     time.Duration
	HalfOpenMaxCalls int
}

// DefaultCircuitBreakerConfig returns default configuration
func DefaultCircuitBreakerConfig() *CircuitBreakerConfig {
	return &CircuitBreakerConfig{
		FailureThreshold: 5,
		ResetTimeout:     60 * time.Second,
		HalfOpenMaxCalls: 3,
	}
}

// Clock abstracts access to the current time for deterministic testing.
type Clock interface {
	Now() time.Time
}

type realClock struct{}

func (realClock) Now() time.Time { return time.Now() }

// NewCircuitBreaker creates a new circuit breaker that uses the real clock.
func NewCircuitBreaker(cfg *CircuitBreakerConfig) *CircuitBreaker {
	return NewCircuitBreakerWithClock(cfg, realClock{})
}

// NewCircuitBreakerWithClock creates a new circuit breaker with an injected clock.
func NewCircuitBreakerWithClock(cfg *CircuitBreakerConfig, clk Clock) *CircuitBreaker {
	return &CircuitBreaker{
		state:            CircuitClosed,
		failureThreshold: cfg.FailureThreshold,
		resetTimeout:     cfg.ResetTimeout,
		halfOpenMaxCalls: cfg.HalfOpenMaxCalls,
		logger:           logging.With(logging.String("component", "circuit_breaker")),
		clock:            clk,
	}
}

// Allow checks if an operation should be allowed
func (cb *CircuitBreaker) Allow() error {
	for {
		cb.mu.Lock()
		switch cb.state {
		case CircuitClosed:
			cb.mu.Unlock()
			return nil
		case CircuitOpen:
			if cb.clock.Now().Sub(cb.lastFailureTime) > cb.resetTimeout {
				cb.state = CircuitHalfOpen
				cb.halfOpenCallCount = 0
				cb.logger.Info("Circuit breaker half-open",
					logging.String("previous_state", "open"),
					logging.String("new_state", "half_open"))
				cb.mu.Unlock()
				continue
			}
			cb.mu.Unlock()
			return fmt.Errorf("circuit breaker is open")
		case CircuitHalfOpen:
			if cb.halfOpenCallCount < cb.halfOpenMaxCalls {
				cb.halfOpenCallCount++
				cb.mu.Unlock()
				return nil
			}
			cb.mu.Unlock()
			return fmt.Errorf("circuit breaker half-open limit exceeded")
		default:
			cb.mu.Unlock()
			return fmt.Errorf("unknown circuit breaker state")
		}
	}
}

// RecordSuccess records a successful operation
func (cb *CircuitBreaker) RecordSuccess() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	switch cb.state {
	case CircuitHalfOpen:
		cb.successCount++
		// Close circuit breaker if we have enough consecutive successes
		if cb.successCount >= cb.halfOpenMaxCalls {
			cb.state = CircuitClosed
			cb.successCount = 0
			cb.failureCount = 0
			cb.logger.Info("Circuit breaker closed",
				logging.String("previous_state", "half_open"),
				logging.String("new_state", "closed"))
		}
	case CircuitClosed:
		// Reset failure count on success
		cb.failureCount = 0
	}
}

// RecordFailure records a failed operation
func (cb *CircuitBreaker) RecordFailure() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	switch cb.state {
	case CircuitClosed, CircuitHalfOpen:
		cb.failureCount++
		cb.lastFailureTime = cb.clock.Now()

		if cb.state == CircuitClosed && cb.failureCount >= cb.failureThreshold {
			// Open circuit breaker
			cb.state = CircuitOpen
			cb.logger.Warn("Circuit breaker opened",
				logging.Int("failure_count", cb.failureCount),
				logging.Int("threshold", cb.failureThreshold))
		} else if cb.state == CircuitHalfOpen {
			// Go back to open on failure in half-open state
			cb.state = CircuitOpen
			cb.logger.Warn("Circuit breaker reopened",
				logging.String("previous_state", "half_open"),
				logging.String("new_state", "open"),
				logging.Int("failure_count", cb.failureCount))
		}
	}
}

// GetState returns the current state
func (cb *CircuitBreaker) GetState() CircuitBreakerState {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.state
}

// GetMetrics returns circuit breaker metrics
func (cb *CircuitBreaker) GetMetrics() map[string]interface{} {
	cb.mu.RLock()
	defer cb.mu.RUnlock()

	return map[string]interface{}{
		"state":             cb.state.String(),
		"failure_count":     cb.failureCount,
		"success_count":     cb.successCount,
		"last_failure_time": cb.lastFailureTime,
		"half_open_calls":   cb.halfOpenCallCount,
		"reset_timeout":     cb.resetTimeout.String(),
		"failure_threshold": cb.failureThreshold,
	}
}
