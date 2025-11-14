// Package main demonstrates production-grade integrations of aws-ssm components.
package main

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/johnlam90/aws-ssm/pkg/aws"
	"github.com/johnlam90/aws-ssm/pkg/health"
	"github.com/johnlam90/aws-ssm/pkg/logging"
	"github.com/johnlam90/aws-ssm/pkg/metrics"
	"github.com/johnlam90/aws-ssm/pkg/security"
	testframework "github.com/johnlam90/aws-ssm/pkg/testing"
	"github.com/johnlam90/aws-ssm/pkg/validation"
)

// ProductionIntegration demonstrates how to use all production-grade features together
func main() {
	// Initialize all production-grade components

	// 1. Initialize structured logging
	logging.Init(
		logging.WithLevel("info"),
		logging.WithOutput(os.Stdout),
		logging.WithEnvironment("production"),
	)
	logger := logging.Default()
	logger.Info("Starting AWS SSM production mode",
		logging.String("version", "1.0.0"),
		logging.String("environment", "production"))

	// 2. Initialize security
	securityManager := security.InitializeSecurity()

	// 3. Initialize metrics service
	ctx := context.Background()
	metricsService := metrics.NewService(ctx)
	metricsService.AddReporter(metrics.NewConsoleReporter())
	metricsService.Start()

	// 4. Initialize health checks
	healthChecker := health.NewDefaultChecker(
		// AWS connectivity test
		func(ctx context.Context) error {
			if err := ctx.Err(); err != nil {
				return err
			}
			// Simulate AWS API call
			return nil
		},
		// Cache test
		func(ctx context.Context) error {
			if err := ctx.Err(); err != nil {
				return err
			}
			// Simulate cache check
			return nil
		},
		// Configuration validation
		func() error {
			// Validate configuration
			return nil
		},
	)

	// 5. Initialize testing framework
	testFramework := testframework.NewTestFramework()

	// Run example operations demonstrating production features
	runExampleSession(logger, securityManager, metricsService, healthChecker, testFramework)
}

func runExampleSession(
	logger logging.Logger,
	securityManager *security.Manager,
	metricsService *metrics.Service,
	healthChecker *health.Checker,
	testFramework *testframework.TestFramework,
) {
	logger.Info("Running example production session")

	// Demonstrate security validation
	command := "ls -la /var/log"
	if err := securityManager.ValidateCommand(command); err != nil {
		logger.Error("Command validation failed", logging.String("error", err.Error()))
		return
	}

	logger.Info("Command validated successfully", logging.String("command", command))

	// Demonstrate metrics collection
	metrics.AWSSSMRequestsTotal.Inc(1.0)

	// Create a timer for operation measurement
	performanceMonitor := metrics.NewPerformanceMonitor()
	timer := performanceMonitor.StartOperation("session_start")

	// Simulate session operation
	time.Sleep(100 * time.Millisecond)

	timer.Stop()
	metrics.CommandExecutionTime.Observe(0.1)
	performanceMonitor.RecordOperation("session_start", 100*time.Millisecond)

	logger.Info("Session operation completed",
		logging.Duration("duration", 100*time.Millisecond))

	// Demonstrate health checking
	healthResult := healthChecker.CheckAll(context.Background())
	logger.Info("Health check completed",
		logging.String("overall_status", healthResult.Overall.String()),
		logging.Int("total_checks", len(healthResult.Checks)))

	// Demonstrate input validation
	validator := validation.InstanceID
	result := validator.Validate("i-1234567890abcdef0")
	if !result.Valid {
		logger.Error("Input validation failed", logging.Any("errors", result.Errors))
		return
	}

	logger.Info("Input validated successfully", logging.Any("fields", result.Fields))

	// Demonstrate security event handling
	eventHandler := security.NewEventHandler()
	event := security.CreateSecurityEvent(
		"session_start",
		"info",
		"production_integration",
		"aws-ssm-cli",
		"Session started successfully",
		map[string]interface{}{"session_id": "test-session-123"},
	)
	eventHandler.HandleEvent(event)

	// Generate security report
	report, err := securityManager.GenerateSecurityReport()
	if err != nil {
		logger.Error("Failed to generate security report", logging.String("error", err.Error()))
	} else {
		logger.Info("Security report generated", logging.String("report", report))
	}

	// Demonstrate testing framework with a simple test
	testSuite := &testframework.TestSuite{
		Name: "Production Integration Tests",
		Tests: []testframework.TestCase{
			{
				Name: "Test Security Validation",
				Function: func(_ *testframework.TestFramework) error {
					// Create a mock testing.T for the assertion
					t := &testing.T{}
					assertion := testframework.NewAssertion(t)

					// Test command validation
					validCommand := "ls -la"
					if err := securityManager.ValidateCommand(validCommand); err != nil {
						return fmt.Errorf("expected valid command to pass: %w", err)
					}

					// Test invalid command
					invalidCommand := "rm -rf /"
					if err := securityManager.ValidateCommand(invalidCommand); err == nil {
						return fmt.Errorf("expected invalid command to fail")
					}

					assertion.Equal(true, true, "Basic assertion test")

					if assertion.HasErrors() {
						return fmt.Errorf("assertion errors: %v", assertion.GetErrors())
					}

					return nil
				},
			},
		},
	}

	// Run the test suite
	results := testFramework.RunTestSuite(testSuite)

	// Log test results
	passed := 0
	failed := 0
	for _, result := range results {
		if result.Passed {
			passed++
		} else if result.Failed {
			failed++
		}
	}

	logger.Info("Test suite completed",
		logging.Int("total", len(results)),
		logging.Int("passed", passed),
		logging.Int("failed", failed))

	// Demonstrate circuit breaker pattern
	circuitBreaker := aws.NewCircuitBreaker(aws.DefaultCircuitBreakerConfig())

	// Simulate some operations
	for i := 0; i < 10; i++ {
		if err := circuitBreaker.Allow(); err != nil {
			logger.Warn("Circuit breaker blocked operation",
				logging.Int("attempt", i),
				logging.String("error", err.Error()))
		} else {
			// Simulate success/failure
			if i < 3 {
				circuitBreaker.RecordSuccess()
				logger.Info("Circuit breaker operation successful", logging.Int("attempt", i))
			} else {
				circuitBreaker.RecordFailure()
				logger.Warn("Circuit breaker operation failed", logging.Int("attempt", i))
			}
		}
	}

	// Log circuit breaker state
	state := circuitBreaker.GetState()
	logger.Info("Circuit breaker final state",
		logging.String("state", state.String()),
		logging.Any("metrics", circuitBreaker.GetMetrics()))

	// Demonstrate rate limiting
	rateLimiter := aws.NewRateLimiter(10.0, 5) // 10 requests per second, burst of 5

	// Simulate request rate limiting
	for i := 0; i < 15; i++ {
		if err := rateLimiter.Acquire(context.Background(), 1.0); err != nil {
			logger.Warn("Request rate limited",
				logging.Int("request", i),
				logging.String("error", err.Error()))
		} else {
			logger.Info("Request allowed", logging.Int("request", i))
		}
	}

	// Create comprehensive production configuration
	productionConfig := &security.Config{
		Level:                    security.SecurityHigh,
		CommandTimeout:           300 * time.Second,
		MaxCommandLength:         512,
		RequireCommandValidation: true,
		EnableAuditLogging:       true,
		CredentialRotationCheck:  true,
		SessionTimeout:           3600 * time.Second,
		RateLimitPerIP:           50,
		EnableTLSVerification:    true,
	}

	logger.Info("Production configuration loaded",
		logging.String("security_level", string(productionConfig.Level)),
		logging.Duration("command_timeout", productionConfig.CommandTimeout),
		logging.Duration("session_timeout", productionConfig.SessionTimeout),
		logging.Int("max_command_length", productionConfig.MaxCommandLength),
		logging.Bool("audit_logging", productionConfig.EnableAuditLogging))

	// Final status summary
	metricsService.Stop()

	logger.Info("Production integration example completed successfully",
		logging.String("status", "success"),
		logging.Duration("total_duration", 5*time.Second))
}

// ExampleAWSClient demonstrates production-grade AWS client usage
func ExampleAWSClient() {
	// Initialize context with cancellation
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Initialize production-grade AWS client (using default aws-ssm config path)
	awsClient, err := aws.NewClient(ctx, "us-east-1", "production", "")
	if err != nil {
		logging.Error("Failed to create AWS client", logging.String("error", err.Error()))
		return
	}

	_ = awsClient // Use the client to avoid unused variable warning

	// Use the client with all production features
	// The client would automatically:
	// - Use structured logging
	// - Apply rate limiting
	// - Use circuit breaker patterns
	// - Collect metrics
	// - Validate inputs
	// - Handle errors gracefully

	fmt.Println("Example production AWS client initialized")
}

// This example demonstrates the complete integration of all production-grade features:
// - Structured logging for observability
// - Security validation and hardening
// - Metrics collection and monitoring
// - Health checking and monitoring
// - Comprehensive testing framework
// - Rate limiting and circuit breakers
// - Input validation and sanitization
// - Error handling and resilience
