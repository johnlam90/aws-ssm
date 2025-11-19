package aws

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
)

// MockSSMAPI is a mock implementation of SSMAPI
type MockSSMAPI struct {
	StartSessionFunc     func(ctx context.Context, params *ssm.StartSessionInput, optFns ...func(*ssm.Options)) (*ssm.StartSessionOutput, error)
	TerminateSessionFunc func(ctx context.Context, params *ssm.TerminateSessionInput, optFns ...func(*ssm.Options)) (*ssm.TerminateSessionOutput, error)
}

func (m *MockSSMAPI) StartSession(ctx context.Context, params *ssm.StartSessionInput, optFns ...func(*ssm.Options)) (*ssm.StartSessionOutput, error) {
	if m.StartSessionFunc != nil {
		return m.StartSessionFunc(ctx, params, optFns...)
	}
	return nil, nil
}

func (m *MockSSMAPI) TerminateSession(ctx context.Context, params *ssm.TerminateSessionInput, optFns ...func(*ssm.Options)) (*ssm.TerminateSessionOutput, error) {
	if m.TerminateSessionFunc != nil {
		return m.TerminateSessionFunc(ctx, params, optFns...)
	}
	return nil, nil
}

// TestHelperProcess is used to mock exec.Command
func TestHelperProcess(_ *testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}
	// We can check arguments here if needed, or just exit
	// For failure test, we can check another env var
	if os.Getenv("GO_HELPER_PROCESS_FAIL") == "1" {
		os.Exit(1)
	}
	os.Exit(0)
}

func TestIsValidAWSRegion(t *testing.T) {
	tests := []struct {
		region string
		want   bool
	}{
		{"us-east-1", true},
		{"eu-west-1", true},
		{"ap-southeast-2", true},
		{"invalid", false},                   // Too short/no hyphens
		{"us-east-1-", false},                // Ends with hyphen
		{"-us-east-1", false},                // Starts with hyphen
		{"us-east-1a", true},                 // Valid availability zone-like (though region validation is loose)
		{"us_east_1", false},                 // Underscores not allowed
		{"US-EAST-1", false},                 // Uppercase not allowed
		{"very-long-region-name-123", false}, // Too long
		{"us-gov-east-1", true},              // GovCloud
		{"cn-north-1", true},                 // China
		{"a-b-1", false},                     // Invalid start
		{"aa-b-1", true},                     // Minimal valid
	}

	for _, tt := range tests {
		t.Run(tt.region, func(t *testing.T) {
			if got := isValidAWSRegion(tt.region); got != tt.want {
				t.Errorf("isValidAWSRegion(%q) = %v, want %v", tt.region, got, tt.want)
			}
		})
	}
}

func TestCheckSessionManagerPlugin(t *testing.T) {
	// Save original execLookPath and restore after test
	origLookPath := execLookPath
	defer func() { execLookPath = origLookPath }()

	t.Run("Plugin Found", func(t *testing.T) {
		execLookPath = func(_ string) (string, error) {
			return "/usr/local/bin/session-manager-plugin", nil
		}
		if err := checkSessionManagerPlugin(); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("Plugin Not Found", func(t *testing.T) {
		execLookPath = func(_ string) (string, error) {
			return "", errors.New("not found")
		}
		if err := checkSessionManagerPlugin(); err == nil {
			t.Error("expected error, got nil")
		}
	})
}

func TestStartSession(t *testing.T) {
	// Save original exec functions and restore after test
	origExecCommand := execCommand
	origLookPath := execLookPath
	defer func() {
		execCommand = origExecCommand
		execLookPath = origLookPath
	}()

	// Mock LookPath to always succeed
	execLookPath = func(_ string) (string, error) {
		return "/usr/local/bin/session-manager-plugin", nil
	}

	// Mock execCommand to call TestHelperProcess
	execCommand = func(name string, arg ...string) *exec.Cmd {
		cs := []string{"-test.run=TestHelperProcess", "--", name}
		cs = append(cs, arg...)
		cmd := exec.Command(os.Args[0], cs...)
		cmd.Env = []string{"GO_WANT_HELPER_PROCESS=1"}
		return cmd
	}

	mockAPI := &MockSSMAPI{
		StartSessionFunc: func(_ context.Context, _ *ssm.StartSessionInput, _ ...func(*ssm.Options)) (*ssm.StartSessionOutput, error) {
			return &ssm.StartSessionOutput{
				SessionId:  aws.String("session-123"),
				TokenValue: aws.String("token"),
				StreamUrl:  aws.String("url"),
			}, nil
		},
		TerminateSessionFunc: func(_ context.Context, _ *ssm.TerminateSessionInput, _ ...func(*ssm.Options)) (*ssm.TerminateSessionOutput, error) {
			return &ssm.TerminateSessionOutput{}, nil
		},
	}

	cb := NewCircuitBreaker(DefaultCircuitBreakerConfig())

	t.Run("Success", func(t *testing.T) {
		err := startSession(context.Background(), mockAPI, "us-east-1", "i-1234567890abcdef0", cb)
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
		err := startSession(context.Background(), failAPI, "us-east-1", "i-123", cb)
		if err == nil {
			t.Error("expected error, got nil")
		}
	})

	t.Run("Invalid Region", func(t *testing.T) {
		err := startSession(context.Background(), mockAPI, "invalid_region", "i-123", cb)
		if err == nil {
			t.Error("expected error, got nil")
		}
	})

	t.Run("Plugin Execution Failure", func(t *testing.T) {
		// Mock execCommand to fail
		execCommand = func(name string, arg ...string) *exec.Cmd {
			cs := []string{"-test.run=TestHelperProcess", "--", name}
			cs = append(cs, arg...)
			cmd := exec.Command(os.Args[0], cs...)
			cmd.Env = []string{"GO_WANT_HELPER_PROCESS=1", "GO_HELPER_PROCESS_FAIL=1"}
			return cmd
		}

		err := startSession(context.Background(), mockAPI, "us-east-1", "i-123", cb)
		if err == nil {
			t.Error("expected error, got nil")
		}
	})
}
