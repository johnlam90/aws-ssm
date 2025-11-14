package aws

import (
	"context"
	"fmt"

	awsSdk "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/mmmorris1975/ssm-session-client/ssmclient"
)

// StartNativeSession initiates an SSM session using pure Go implementation (no plugin required)
func (c *Client) StartNativeSession(ctx context.Context, instanceID string) error {
	fmt.Printf("Starting native SSM session with instance %s...\n", instanceID)
	fmt.Println("(Using pure Go implementation - no session-manager-plugin required)")
	fmt.Println()

	// Check circuit breaker before making API call
	if err := c.CircuitBreaker.Allow(); err != nil {
		return fmt.Errorf("circuit breaker open (too many recent failures), please retry in a few moments: %w", err)
	}

	// First, call StartSession API to get session credentials
	// DocumentName: nil uses the default AWS-StartInteractiveCommand document for shell sessions
	input := &ssm.StartSessionInput{
		Target:       &instanceID,
		DocumentName: nil,
	}

	result, err := c.SSMClient.StartSession(ctx, input)
	if err != nil {
		c.CircuitBreaker.RecordFailure()
		return fmt.Errorf("failed to start SSM session: %w", err)
	}

	// Record success
	c.CircuitBreaker.RecordSuccess()

	sessionID := *result.SessionId

	// Use the ssm-session-client library for shell session
	// It accepts AWS SDK v2 config directly
	if sessionErr := ssmclient.ShellSession(c.Config, instanceID); sessionErr != nil {
		// Attempt to terminate the session even if it failed
		terminateErr := c.terminateSessionSilently(ctx, sessionID)
		if terminateErr != nil {
			fmt.Printf("Warning: failed to terminate session after error: %v\n", terminateErr)
		}
		return fmt.Errorf("failed to start native session: %w", sessionErr)
	}

	// Terminate the session after it completes
	if terminateErr := c.terminateSessionSilently(ctx, sessionID); terminateErr != nil {
		// Log but don't fail - session might already be terminated
		fmt.Printf("Warning: failed to terminate session: %v\n", terminateErr)
	}

	return nil
}

// StartPortForwardingSession starts a port forwarding session using pure Go
func (c *Client) StartPortForwardingSession(ctx context.Context, instanceID string, remotePort, localPort int) error {
	fmt.Printf("Starting port forwarding session...\n")
	fmt.Printf("  Instance:    %s\n", instanceID)
	fmt.Printf("  Remote Port: %d\n", remotePort)
	fmt.Printf("  Local Port:  %d\n", localPort)
	fmt.Println("(Using pure Go implementation - no session-manager-plugin required)")
	fmt.Println()

	// Check circuit breaker before making API call
	if err := c.CircuitBreaker.Allow(); err != nil {
		return fmt.Errorf("circuit breaker open (too many recent failures), please retry in a few moments: %w", err)
	}

	// First, call StartSession API to get session credentials
	// Use AWS-StartPortForwardingSession document explicitly for port forwarding
	input := &ssm.StartSessionInput{
		Target:       &instanceID,
		DocumentName: awsSdk.String("AWS-StartPortForwardingSession"),
	}

	result, err := c.SSMClient.StartSession(ctx, input)
	if err != nil {
		c.CircuitBreaker.RecordFailure()
		return fmt.Errorf("failed to start SSM session: %w", err)
	}

	// Record success
	c.CircuitBreaker.RecordSuccess()

	sessionID := *result.SessionId

	portForwardingInput := &ssmclient.PortForwardingInput{
		Target:     instanceID,
		RemotePort: remotePort,
		LocalPort:  localPort,
	}

	if sessionErr := ssmclient.PortForwardingSession(c.Config, portForwardingInput); sessionErr != nil {
		// Attempt to terminate the session even if it failed
		terminateErr := c.terminateSessionSilently(ctx, sessionID)
		if terminateErr != nil {
			fmt.Printf("Warning: failed to terminate session after error: %v\n", terminateErr)
		}
		return fmt.Errorf("failed to start port forwarding session: %w", sessionErr)
	}

	// Terminate the session after it completes
	if terminateErr := c.terminateSessionSilently(ctx, sessionID); terminateErr != nil {
		// Log but don't fail - session might already be terminated
		fmt.Printf("Warning: failed to terminate session: %v\n", terminateErr)
	}

	return nil
}

// terminateSessionSilently terminates an SSM session without failing
func (c *Client) terminateSessionSilently(ctx context.Context, sessionID string) error {
	terminateInput := &ssm.TerminateSessionInput{
		SessionId: &sessionID,
	}
	_, err := c.SSMClient.TerminateSession(ctx, terminateInput)
	return err
}
