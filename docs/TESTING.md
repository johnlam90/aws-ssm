# Testing Guide for AWS SSM CLI

This document describes the testing approach, test framework, and integration test harness for the AWS SSM CLI.

## Overview

The AWS SSM CLI includes comprehensive testing infrastructure:

- **Unit Tests**: Test individual components in isolation
- **Integration Tests**: Test AWS interactions with mocked services
- **Test Framework**: Comprehensive testing utilities in `pkg/testing`
- **Dry-Run Mode**: Validate selection workflows without executing commands

## Running Tests

### Run All Tests

```bash
# Run all tests with coverage
make test

# Run tests with coverage report
make test-coverage

# Run tests for specific package
go test -v ./pkg/aws/...

# Run tests with race detector
go test -race ./...
```

### Run Specific Tests

```bash
# Run tests matching a pattern
go test -v -run TestRateLimiter ./pkg/aws

# Run tests with timeout
go test -v --timeout=60s ./...

# Run tests with verbose output
go test -v ./pkg/aws
```

## Test Framework

The `pkg/testing` package provides a comprehensive testing framework:

### Basic Usage

```go
import "github.com/johnlam90/aws-ssm/pkg/testing"

// Create a test framework instance
tf := testing.NewTestFramework()

// Create test cases
testCases := []testing.TestCase{
    {
        Name: "Test instance resolution",
        Function: func(tf *testing.TestFramework) error {
            // Test logic here
            return nil
        },
    },
}

// Run tests
results := tf.RunTests(testCases)
```

### Assertions

```go
// Create assertions
assertion := testing.NewAssertion(t)

// Use assertion methods
assertion.True(condition, "message")
assertion.False(condition, "message")
assertion.Equal(expected, actual, "message")
assertion.NotEqual(expected, actual, "message")
assertion.Nil(value, "message")
assertion.NotNil(value, "message")
assertion.Error(err, "message")
assertion.NoError(err, "message")
```

## Integration Test Harness

### Dry-Run Mode

The CLI supports a dry-run mode for testing selection workflows without executing commands:

```bash
# List instances (dry-run)
aws-ssm list --dry-run

# Start session (dry-run)
aws-ssm session i-1234567890abcdef0 --dry-run

# Port forward (dry-run)
aws-ssm port-forward i-1234567890abcdef0 --dry-run
```

### Mocked AWS Services

For integration testing, use mocked AWS services:

```go
import (
    "context"
    "github.com/johnlam90/aws-ssm/pkg/aws"
)

// Create a client with mocked services
ctx := context.Background()
client, err := aws.NewClient(ctx, "us-east-1", "test")
if err != nil {
    t.Fatalf("failed to create client: %v", err)
}

// Use the client for testing
instances, err := client.ListInstances(ctx, nil)
if err != nil {
    t.Fatalf("failed to list instances: %v", err)
}
```

## Test Coverage

### Current Test Coverage

- **Identifier Parsing**: 100% coverage (40+ test cases)
- **Instance Resolution**: Comprehensive tests for multi-instance scenarios
- **Rate Limiter**: Tests for token acquisition, context cancellation, and capacity
- **Circuit Breaker**: Tests for state transitions and recovery
- **Security Patterns**: Tests for command validation and pattern matching
- **Metrics**: Tests for histogram bucketing and counter operations

### Running Coverage Reports

```bash
# Generate coverage report
make test-coverage

# View coverage in browser
open coverage.html

# Check coverage for specific package
go test -coverprofile=coverage.out ./pkg/aws
go tool cover -html=coverage.out
```

## Writing Tests

### Unit Test Example

```go
func TestInstanceResolution(t *testing.T) {
    tests := []struct {
        name        string
        identifier  string
        expected    string
        shouldError bool
    }{
        {
            name:        "valid instance ID",
            identifier:  "i-1234567890abcdef0",
            expected:    "i-1234567890abcdef0",
            shouldError: false,
        },
        {
            name:        "invalid instance ID",
            identifier:  "invalid",
            expected:    "",
            shouldError: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Test logic
        })
    }
}
```

### Integration Test Example

```go
func TestSessionCreation(t *testing.T) {
    ctx := context.Background()
    client, err := aws.NewClient(ctx, "us-east-1", "test")
    if err != nil {
        t.Fatalf("failed to create client: %v", err)
    }

    // Test session creation
    instance, err := client.ResolveSingleInstance(ctx, "i-1234567890abcdef0")
    if err != nil {
        t.Fatalf("failed to resolve instance: %v", err)
    }

    if instance.InstanceID != "i-1234567890abcdef0" {
        t.Errorf("expected instance ID i-1234567890abcdef0, got %s", instance.InstanceID)
    }
}
```

## Best Practices

### Test Organization

- Use table-driven tests for multiple scenarios
- Group related tests in the same file
- Use descriptive test names
- Include setup and teardown when needed

### Test Isolation

- Each test should be independent
- Use temporary files/directories for file operations
- Mock external dependencies
- Clean up resources in defer statements

### Test Performance

- Keep tests fast (< 100ms per test)
- Use context timeouts to prevent hanging
- Avoid unnecessary sleeps
- Use parallel testing when possible

### Test Coverage

- Aim for > 80% code coverage
- Test error paths and edge cases
- Test concurrent operations
- Test timeout and cancellation scenarios

## Continuous Integration

Tests are automatically run on:

- Every commit to the `hardening` branch
- Pull requests to `main` and `develop` branches
- Release branches (`release/*`)

See `.github/workflows/` for CI configuration.

## Troubleshooting

### Tests Hanging

If tests hang, check for:
- Missing context cancellation
- Deadlocks in concurrent code
- Infinite loops in retry logic

Use `go test -timeout=30s` to set a timeout.

### Flaky Tests

If tests are flaky, check for:
- Race conditions (use `go test -race`)
- Timing-dependent assertions
- Uncontrolled randomness

### Coverage Gaps

To find untested code:

```bash
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

## Additional Resources

- [Go Testing Package](https://golang.org/pkg/testing/)
- [Table-Driven Tests](https://github.com/golang/go/wiki/TableDrivenTests)
- [Go Test Best Practices](https://golang.org/doc/effective_go#testing)
