package aws

import (
	"context"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/aws/aws-sdk-go-v2/service/ssm/types"
)

const (
	// DefaultCommandTimeout is the default timeout for command execution
	DefaultCommandTimeout = 2 * time.Minute
	// MaxCommandTimeout is the maximum allowed timeout
	MaxCommandTimeout = 10 * time.Minute
)

// ExecuteCommand executes a command on an EC2 instance via SSM and returns the output
// Note: Commands with spaces should be properly quoted when passed from the shell.
// Example: aws-ssm session i-1234567890abcdef0 "echo 'hello world'"
func (c *Client) ExecuteCommand(ctx context.Context, instanceID, command string) (string, error) {
	return c.ExecuteCommandWithTimeout(ctx, instanceID, command, DefaultCommandTimeout)
}

// ExecuteCommandWithTimeout executes a command with a configurable timeout
// The command parameter should be a complete shell command string.
// Multi-word commands must be properly quoted to be treated as a single command.
func (c *Client) ExecuteCommandWithTimeout(ctx context.Context, instanceID, command string, timeout time.Duration) (string, error) {
	// Validate and normalize timeout
	normalizedTimeout := normalizeTimeout(timeout)

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(ctx, normalizedTimeout)
	defer cancel()

	// Check circuit breaker before making API call
	if err := c.CircuitBreaker.Allow(); err != nil {
		return "", fmt.Errorf("circuit breaker open: %w", err)
	}

	// Send the command using SSM SendCommand API
	commandID, err := c.sendCommand(ctx, instanceID, command)
	if err != nil {
		return "", err
	}

	// Wait for the command to complete
	fmt.Printf("Command ID: %s\n", commandID)
	fmt.Printf("Waiting for command to complete (timeout: %v)...\n\n", normalizedTimeout)

	// Poll for command completion
	return c.pollCommandCompletion(ctx, instanceID, commandID)
}

// normalizeTimeout validates and normalizes the timeout value
func normalizeTimeout(timeout time.Duration) time.Duration {
	if timeout > MaxCommandTimeout {
		timeout = MaxCommandTimeout
	}
	if timeout < 10*time.Second {
		timeout = 10 * time.Second
	}
	return timeout
}

// sendCommand sends the command to the instance via SSM
func (c *Client) sendCommand(ctx context.Context, instanceID, command string) (string, error) {
	sendInput := &ssm.SendCommandInput{
		InstanceIds:  []string{instanceID},
		DocumentName: aws.String("AWS-RunShellScript"),
		Parameters: map[string][]string{
			"commands": {command},
		},
		Comment: aws.String("Executed via aws-ssm CLI"),
	}

	sendResult, err := c.SSMClient.SendCommand(ctx, sendInput)
	if err != nil {
		c.CircuitBreaker.RecordFailure()
		return "", fmt.Errorf("failed to send command: %w", err)
	}

	commandID := aws.ToString(sendResult.Command.CommandId)
	return commandID, nil
}

// pollCommandCompletion polls for command completion and returns the result
func (c *Client) pollCommandCompletion(ctx context.Context, instanceID, commandID string) (string, error) {
	backoff := 500 * time.Millisecond
	maxBackoff := 5 * time.Second

	for {
		select {
		case <-ctx.Done():
			return "", fmt.Errorf("command execution cancelled or timed out: %w", ctx.Err())
		default:
		}

		// Get command invocation status
		invocationResult, err := c.getCommandInvocation(ctx, instanceID, commandID)
		if err != nil {
			// Command might not be ready yet, continue polling with backoff
			time.Sleep(backoff)
			// Exponential backoff with max cap
			backoff = time.Duration(math.Min(float64(backoff*2), float64(maxBackoff)))
			continue
		}

		// Record success for the API call
		c.CircuitBreaker.RecordSuccess()

		// Handle the command status
		result, err := c.handleCommandStatus(invocationResult, backoff)
		if err == nil {
			// Command completed successfully
			return result, nil
		}
		if err.Error() == "command still in progress" {
			// Update backoff for next iteration
			backoff = time.Duration(math.Min(float64(backoff*2), float64(maxBackoff)))
			continue
		}
		// Return other errors
		return "", err
	}
}

// getCommandInvocation gets the command invocation result
func (c *Client) getCommandInvocation(ctx context.Context, instanceID, commandID string) (*ssm.GetCommandInvocationOutput, error) {
	invocationInput := &ssm.GetCommandInvocationInput{
		CommandId:  aws.String(commandID),
		InstanceId: aws.String(instanceID),
	}

	return c.SSMClient.GetCommandInvocation(ctx, invocationInput)
}

// handleCommandStatus handles different command status outcomes
func (c *Client) handleCommandStatus(invocationResult *ssm.GetCommandInvocationOutput, backoff time.Duration) (string, error) {
	status := invocationResult.Status

	switch status {
	case types.CommandInvocationStatusSuccess:
		// Command completed successfully
		return c.formatCommandOutput(invocationResult), nil

	case types.CommandInvocationStatusFailed:
		// Command failed
		stderr := aws.ToString(invocationResult.StandardErrorContent)
		return "", fmt.Errorf("command failed: %s", stderr)

	case types.CommandInvocationStatusCancelled:
		return "", fmt.Errorf("command was cancelled")

	case types.CommandInvocationStatusTimedOut:
		return "", fmt.Errorf("command timed out")

	case types.CommandInvocationStatusPending, types.CommandInvocationStatusInProgress:
		// Still running, continue polling with backoff
		time.Sleep(backoff)
		return "", fmt.Errorf("command still in progress")

	default:
		// Unknown status
		return "", fmt.Errorf("unknown command status: %s", status)
	}
}

// formatCommandOutput formats the command output with clear separation between stdout and stderr
func (c *Client) formatCommandOutput(invocationResult *ssm.GetCommandInvocationOutput) string {
	output := aws.ToString(invocationResult.StandardOutputContent)
	stderr := aws.ToString(invocationResult.StandardErrorContent)

	// Format output with clear separation between stdout and stderr
	var result strings.Builder
	if output != "" {
		result.WriteString(output)
	}
	if stderr != "" {
		if output != "" {
			result.WriteString("\n")
		}
		result.WriteString("[STDERR]\n")
		result.WriteString(stderr)
	}
	return result.String()
}
