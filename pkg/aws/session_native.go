package aws

import (
	"context"
	"fmt"

	"github.com/mmmorris1975/ssm-session-client/ssmclient"
)

// StartNativeSession initiates an SSM session using pure Go implementation (no plugin required)
func (c *Client) StartNativeSession(ctx context.Context, instanceID string) error {
	fmt.Printf("Starting native SSM session with instance %s...\n", instanceID)
	fmt.Println("(Using pure Go implementation - no session-manager-plugin required)")
	fmt.Println()

	// Use the ssm-session-client library for shell session
	// It accepts AWS SDK v2 config directly
	if err := ssmclient.ShellSession(c.Config, instanceID); err != nil {
		return fmt.Errorf("failed to start native session: %w", err)
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

	input := &ssmclient.PortForwardingInput{
		Target:     instanceID,
		RemotePort: remotePort,
		LocalPort:  localPort,
	}

	if err := ssmclient.PortForwardingSession(c.Config, input); err != nil {
		return fmt.Errorf("failed to start port forwarding session: %w", err)
	}

	return nil
}
