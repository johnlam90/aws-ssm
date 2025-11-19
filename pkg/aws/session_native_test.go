package aws

import (
	"context"
	"errors"
	"io"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/mmmorris1975/ssm-session-client/ssmclient"
)

func TestStartNativeSession(t *testing.T) {
	// Save original functions
	origShellSession := ssmShellSession
	defer func() { ssmShellSession = origShellSession }()

	mockAPI := &MockSSMAPI{
		StartSessionFunc: func(_ context.Context, _ *ssm.StartSessionInput, _ ...func(*ssm.Options)) (*ssm.StartSessionOutput, error) {
			return &ssm.StartSessionOutput{
				SessionId: aws.String("session-123"),
			}, nil
		},
		TerminateSessionFunc: func(_ context.Context, _ *ssm.TerminateSessionInput, _ ...func(*ssm.Options)) (*ssm.TerminateSessionOutput, error) {
			return &ssm.TerminateSessionOutput{}, nil
		},
	}

	cb := NewCircuitBreaker(DefaultCircuitBreakerConfig())

	t.Run("Success", func(t *testing.T) {
		ssmShellSession = func(_ aws.Config, _ string, _ ...io.Reader) error {
			return nil
		}

		err := startNativeSession(context.Background(), mockAPI, aws.Config{}, "i-123", cb)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("SSM StartSession Failure", func(t *testing.T) {
		failAPI := &MockSSMAPI{
			StartSessionFunc: func(_ context.Context, _ *ssm.StartSessionInput, _ ...func(*ssm.Options)) (*ssm.StartSessionOutput, error) {
				return nil, errors.New("ssm error")
			},
		}
		err := startNativeSession(context.Background(), failAPI, aws.Config{}, "i-123", cb)
		if err == nil {
			t.Error("expected error, got nil")
		}
	})

	t.Run("Shell Session Failure", func(t *testing.T) {
		ssmShellSession = func(_ aws.Config, _ string, _ ...io.Reader) error {
			return errors.New("shell error")
		}

		err := startNativeSession(context.Background(), mockAPI, aws.Config{}, "i-123", cb)
		if err == nil {
			t.Error("expected error, got nil")
		}
	})
}

func TestStartPortForwardingSession(t *testing.T) {
	// Save original functions
	origPortForwardingSession := ssmPortForwardingSession
	defer func() { ssmPortForwardingSession = origPortForwardingSession }()

	mockAPI := &MockSSMAPI{
		StartSessionFunc: func(_ context.Context, _ *ssm.StartSessionInput, _ ...func(*ssm.Options)) (*ssm.StartSessionOutput, error) {
			return &ssm.StartSessionOutput{
				SessionId: aws.String("session-123"),
			}, nil
		},
		TerminateSessionFunc: func(_ context.Context, _ *ssm.TerminateSessionInput, _ ...func(*ssm.Options)) (*ssm.TerminateSessionOutput, error) {
			return &ssm.TerminateSessionOutput{}, nil
		},
	}

	cb := NewCircuitBreaker(DefaultCircuitBreakerConfig())

	t.Run("Success", func(t *testing.T) {
		ssmPortForwardingSession = func(_ aws.Config, _ *ssmclient.PortForwardingInput) error {
			return nil
		}

		err := startPortForwardingSession(context.Background(), mockAPI, aws.Config{}, "i-123", 8080, 8080, cb)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("SSM StartSession Failure", func(t *testing.T) {
		failAPI := &MockSSMAPI{
			StartSessionFunc: func(_ context.Context, _ *ssm.StartSessionInput, _ ...func(*ssm.Options)) (*ssm.StartSessionOutput, error) {
				return nil, errors.New("ssm error")
			},
		}
		err := startPortForwardingSession(context.Background(), failAPI, aws.Config{}, "i-123", 8080, 8080, cb)
		if err == nil {
			t.Error("expected error, got nil")
		}
	})

	t.Run("Port Forwarding Session Failure", func(t *testing.T) {
		ssmPortForwardingSession = func(_ aws.Config, _ *ssmclient.PortForwardingInput) error {
			return errors.New("port forwarding error")
		}

		err := startPortForwardingSession(context.Background(), mockAPI, aws.Config{}, "i-123", 8080, 8080, cb)
		if err == nil {
			t.Error("expected error, got nil")
		}
	})
}
