package aws

import (
	"context"
	"fmt"

	awsSdk "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/mmmorris1975/ssm-session-client/ssmclient"
)

// Variables for mocking in tests
var (
	ssmShellSession          = ssmclient.ShellSession
	ssmPortForwardingSession = ssmclient.PortForwardingSession
)

// StartNativeSession initiates an SSM session using pure Go implementation (no plugin required)
func (c *Client) StartNativeSession(ctx context.Context, instanceID string) error {
	// Use the existing client if available, otherwise create a new one (fallback)
	var api SSMAPI
	if c.SSMClient != nil {
		api = c.SSMClient
	} else {
		api = ssm.NewFromConfig(c.Config)
	}

	return startNativeSession(ctx, api, c.Config, instanceID, c.CircuitBreaker)
}

func startNativeSession(ctx context.Context, api SSMAPI, config awsSdk.Config, instanceID string, cb *CircuitBreaker) error {
	fmt.Printf("Starting native SSM session with instance %s...\n", instanceID)
	fmt.Println("(Using pure Go implementation - no session-manager-plugin required)")
	fmt.Println()

	// Check circuit breaker before making API call
	if err := cb.Allow(); err != nil {
		return fmt.Errorf("circuit breaker open (too many recent failures), please retry in a few moments: %w", err)
	}

	// First, call StartSession API to get session credentials
	// DocumentName: nil uses the default AWS-StartInteractiveCommand document for shell sessions
	input := &ssm.StartSessionInput{
		Target:       &instanceID,
		DocumentName: nil,
	}

	result, err := api.StartSession(ctx, input)
	if err != nil {
		cb.RecordFailure()
		return fmt.Errorf("failed to start SSM session: %w", err)
	}

	// Record success
	cb.RecordSuccess()

	sessionID := *result.SessionId

	// Use the ssm-session-client library for shell session
	// It accepts AWS SDK v2 config directly
	if sessionErr := ssmShellSession(config, instanceID); sessionErr != nil {
		// Attempt to terminate the session even if it failed
		terminateErr := terminateSessionSilently(ctx, api, sessionID)
		if terminateErr != nil {
			fmt.Printf("Warning: failed to terminate session after error: %v\n", terminateErr)
		}
		return fmt.Errorf("failed to start native session: %w", sessionErr)
	}

	// Terminate the session after it completes
	if terminateErr := terminateSessionSilently(ctx, api, sessionID); terminateErr != nil {
		// Log but don't fail - session might already be terminated
		fmt.Printf("Warning: failed to terminate session: %v\n", terminateErr)
	}

	return nil
}

// StartPortForwardingSession starts a port forwarding session using pure Go
func (c *Client) StartPortForwardingSession(ctx context.Context, instanceID string, remotePort, localPort int) error {
	// Use the existing client if available, otherwise create a new one (fallback)
	var api SSMAPI
	if c.SSMClient != nil {
		api = c.SSMClient
	} else {
		api = ssm.NewFromConfig(c.Config)
	}

	return startPortForwardingSession(ctx, api, c.Config, instanceID, remotePort, localPort, c.CircuitBreaker)
}

func startPortForwardingSession(ctx context.Context, api SSMAPI, config awsSdk.Config, instanceID string, remotePort, localPort int, cb *CircuitBreaker) error {
	fmt.Printf("Starting port forwarding session...\n")
	fmt.Printf("  Instance:    %s\n", instanceID)
	fmt.Printf("  Remote Port: %d\n", remotePort)
	fmt.Printf("  Local Port:  %d\n", localPort)
	fmt.Println("(Using pure Go implementation - no session-manager-plugin required)")
	fmt.Println()

	// Check circuit breaker before making API call
	if err := cb.Allow(); err != nil {
		return fmt.Errorf("circuit breaker open (too many recent failures), please retry in a few moments: %w", err)
	}

	// First, call StartSession API to get session credentials
	// Use AWS-StartPortForwardingSession document explicitly for port forwarding
	input := &ssm.StartSessionInput{
		Target:       &instanceID,
		DocumentName: awsSdk.String("AWS-StartPortForwardingSession"),
	}

	result, err := api.StartSession(ctx, input)
	if err != nil {
		cb.RecordFailure()
		return fmt.Errorf("failed to start SSM session: %w", err)
	}

	// Record success
	cb.RecordSuccess()

	sessionID := *result.SessionId

	portForwardingInput := &ssmclient.PortForwardingInput{
		Target:     instanceID,
		RemotePort: remotePort,
		LocalPort:  localPort,
	}

	if sessionErr := ssmPortForwardingSession(config, portForwardingInput); sessionErr != nil {
		// Attempt to terminate the session even if it failed
		terminateErr := terminateSessionSilently(ctx, api, sessionID)
		if terminateErr != nil {
			fmt.Printf("Warning: failed to terminate session after error: %v\n", terminateErr)
		}
		return fmt.Errorf("failed to start port forwarding session: %w", sessionErr)
	}

	// Terminate the session after it completes
	if terminateErr := terminateSessionSilently(ctx, api, sessionID); terminateErr != nil {
		// Log but don't fail - session might already be terminated
		fmt.Printf("Warning: failed to terminate session: %v\n", terminateErr)
	}

	return nil
}

func terminateSessionSilently(ctx context.Context, api SSMAPI, sessionID string) error {
	terminateInput := &ssm.TerminateSessionInput{
		SessionId: &sessionID,
	}
	_, err := api.TerminateSession(ctx, terminateInput)
	return err
}
