package aws

import (
	"context"
	"fmt"
	"math"
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
func (c *Client) ExecuteCommand(ctx context.Context, instanceID, command string) (string, error) {
	return c.ExecuteCommandWithTimeout(ctx, instanceID, command, DefaultCommandTimeout)
}

// ExecuteCommandWithTimeout executes a command with a configurable timeout
func (c *Client) ExecuteCommandWithTimeout(ctx context.Context, instanceID, command string, timeout time.Duration) (string, error) {
	// Validate timeout
	if timeout > MaxCommandTimeout {
		timeout = MaxCommandTimeout
	}
	if timeout < 10*time.Second {
		timeout = 10 * time.Second
	}

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Send the command using SSM SendCommand API
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
		return "", fmt.Errorf("failed to send command: %w", err)
	}

	commandID := aws.ToString(sendResult.Command.CommandId)

	// Wait for the command to complete
	fmt.Printf("Command ID: %s\n", commandID)
	fmt.Printf("Waiting for command to complete (timeout: %v)...\n\n", timeout)

	// Poll for command completion with exponential backoff
	attempt := 0
	backoff := 500 * time.Millisecond
	maxBackoff := 5 * time.Second

	for {
		select {
		case <-ctx.Done():
			return "", fmt.Errorf("command execution cancelled or timed out: %w", ctx.Err())
		default:
		}

		// Get command invocation status
		invocationInput := &ssm.GetCommandInvocationInput{
			CommandId:  aws.String(commandID),
			InstanceId: aws.String(instanceID),
		}

		invocationResult, err := c.SSMClient.GetCommandInvocation(ctx, invocationInput)
		if err != nil {
			// Command might not be ready yet, continue polling with backoff
			attempt++
			time.Sleep(backoff)
			// Exponential backoff with max cap
			backoff = time.Duration(math.Min(float64(backoff*2), float64(maxBackoff)))
			continue
		}

		status := invocationResult.Status

		switch status {
		case types.CommandInvocationStatusSuccess:
			// Command completed successfully
			output := aws.ToString(invocationResult.StandardOutputContent)
			stderr := aws.ToString(invocationResult.StandardErrorContent)

			if stderr != "" {
				return output + "\nStderr:\n" + stderr, nil
			}
			return output, nil

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
			attempt++
			time.Sleep(backoff)
			// Exponential backoff with max cap
			backoff = time.Duration(math.Min(float64(backoff*2), float64(maxBackoff)))
			continue

		default:
			// Unknown status
			return "", fmt.Errorf("unknown command status: %s", status)
		}
	}
}
