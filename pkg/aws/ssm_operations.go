package aws

import (
	"context"
	"time"
)

// SSMOperations defines the interface for SSM operations
// This interface allows for mocking and testing of SSM operations
type SSMOperations interface {
	// StartSession initiates an SSM session with the specified instance using the plugin
	StartSession(ctx context.Context, instanceID string) error

	// StartNativeSession initiates an SSM session using pure Go implementation (no plugin required)
	StartNativeSession(ctx context.Context, instanceID string) error

	// StartPortForwardingSession starts a port forwarding session
	StartPortForwardingSession(ctx context.Context, instanceID string, remotePort, localPort int) error

	// ExecuteCommand executes a command on an EC2 instance via SSM and returns the output
	ExecuteCommand(ctx context.Context, instanceID, command string) (string, error)

	// ExecuteCommandWithTimeout executes a command with a specific timeout
	ExecuteCommandWithTimeout(ctx context.Context, instanceID, command string, timeout time.Duration) (string, error)
}

// Ensure Client implements SSMOperations
var _ SSMOperations = (*Client)(nil)

// MockSSMOperations provides a mock implementation of SSMOperations for testing
type MockSSMOperations struct {
	StartSessionFunc               func(ctx context.Context, instanceID string) error
	StartNativeSessionFunc         func(ctx context.Context, instanceID string) error
	StartPortForwardingSessionFunc func(ctx context.Context, instanceID string, remotePort, localPort int) error
	ExecuteCommandFunc             func(ctx context.Context, instanceID, command string) (string, error)
	ExecuteCommandWithTimeoutFunc  func(ctx context.Context, instanceID, command string, timeout time.Duration) (string, error)
}

// StartSession calls the mock function
func (m *MockSSMOperations) StartSession(ctx context.Context, instanceID string) error {
	if m.StartSessionFunc != nil {
		return m.StartSessionFunc(ctx, instanceID)
	}
	return nil
}

// StartNativeSession calls the mock function
func (m *MockSSMOperations) StartNativeSession(ctx context.Context, instanceID string) error {
	if m.StartNativeSessionFunc != nil {
		return m.StartNativeSessionFunc(ctx, instanceID)
	}
	return nil
}

// StartPortForwardingSession calls the mock function
func (m *MockSSMOperations) StartPortForwardingSession(ctx context.Context, instanceID string, remotePort, localPort int) error {
	if m.StartPortForwardingSessionFunc != nil {
		return m.StartPortForwardingSessionFunc(ctx, instanceID, remotePort, localPort)
	}
	return nil
}

// ExecuteCommand calls the mock function
func (m *MockSSMOperations) ExecuteCommand(ctx context.Context, instanceID, command string) (string, error) {
	if m.ExecuteCommandFunc != nil {
		return m.ExecuteCommandFunc(ctx, instanceID, command)
	}
	return "", nil
}

// ExecuteCommandWithTimeout calls the mock function
func (m *MockSSMOperations) ExecuteCommandWithTimeout(ctx context.Context, instanceID, command string, timeout time.Duration) (string, error) {
	if m.ExecuteCommandWithTimeoutFunc != nil {
		return m.ExecuteCommandWithTimeoutFunc(ctx, instanceID, command, timeout)
	}
	return "", nil
}
