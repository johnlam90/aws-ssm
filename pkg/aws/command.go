package aws

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/aws/aws-sdk-go-v2/service/ssm/types"
)

// ExecuteCommand executes a command on an EC2 instance via SSM and returns the output
func (c *Client) ExecuteCommand(ctx context.Context, instanceID, command string) (string, error) {
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
	fmt.Printf("Waiting for command to complete...\n\n")

	// Poll for command completion
	maxAttempts := 60 // 60 attempts * 2 seconds = 2 minutes max wait
	for attempt := 0; attempt < maxAttempts; attempt++ {
		// Get command invocation status
		invocationInput := &ssm.GetCommandInvocationInput{
			CommandId:  aws.String(commandID),
			InstanceId: aws.String(instanceID),
		}

		invocationResult, err := c.SSMClient.GetCommandInvocation(ctx, invocationInput)
		if err != nil {
			// Command might not be ready yet, continue polling
			time.Sleep(2 * time.Second)
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
			// Still running, continue polling
			time.Sleep(2 * time.Second)
			continue

		default:
			// Unknown status
			return "", fmt.Errorf("unknown command status: %s", status)
		}
	}

	return "", fmt.Errorf("command execution timed out after waiting for 2 minutes")
}
