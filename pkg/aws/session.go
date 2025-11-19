package aws

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"regexp"
	"syscall"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
)

// SSMAPI defines the interface for SSM operations
type SSMAPI interface {
	StartSession(ctx context.Context, params *ssm.StartSessionInput, optFns ...func(*ssm.Options)) (*ssm.StartSessionOutput, error)
	TerminateSession(ctx context.Context, params *ssm.TerminateSessionInput, optFns ...func(*ssm.Options)) (*ssm.TerminateSessionOutput, error)
}

// Variables for mocking in tests
var (
	execCommand  = exec.Command
	execLookPath = exec.LookPath
)

// StartSession initiates an SSM session with the specified instance
func (c *Client) StartSession(ctx context.Context, instanceID string) error {
	// Check if session-manager-plugin is installed
	if err := checkSessionManagerPlugin(); err != nil {
		return err
	}

	// Check circuit breaker before making API call
	if err := c.CircuitBreaker.Allow(); err != nil {
		return fmt.Errorf("circuit breaker open: %w", err)
	}

	// Use the existing client if available, otherwise create a new one (fallback)
	var api SSMAPI
	if c.SSMClient != nil {
		api = c.SSMClient
	} else {
		api = ssm.NewFromConfig(c.Config)
	}

	return startSession(ctx, api, c.Config.Region, instanceID, c.CircuitBreaker)
}

func startSession(ctx context.Context, api SSMAPI, region, instanceID string, cb *CircuitBreaker) error {
	// Start SSM session
	input := &ssm.StartSessionInput{
		Target: aws.String(instanceID),
	}

	result, err := api.StartSession(ctx, input)
	if err != nil {
		cb.RecordFailure()
		return fmt.Errorf("failed to start SSM session: %w", err)
	}

	// Record success
	cb.RecordSuccess()

	sessionID := aws.ToString(result.SessionId)
	defer func() {
		if sessionID == "" {
			return
		}
		terminateInput := &ssm.TerminateSessionInput{
			SessionId: aws.String(sessionID),
		}
		if _, terminateErr := api.TerminateSession(ctx, terminateInput); terminateErr != nil {
			// Log but don't fail - session might already be terminated
			fmt.Printf("Warning: failed to terminate session: %v\n", terminateErr)
		}
	}()

	// Prepare session data for the plugin
	sessionData := map[string]interface{}{
		"SessionId":  sessionID,
		"TokenValue": aws.ToString(result.TokenValue),
		"StreamUrl":  aws.ToString(result.StreamUrl),
		"Target":     instanceID,
	}

	sessionJSON, marshalErr := json.Marshal(sessionData)
	if marshalErr != nil {
		return fmt.Errorf("failed to marshal session data: %w", marshalErr)
	}

	// Validate region to prevent command injection
	if !isValidAWSRegion(region) {
		return fmt.Errorf("invalid AWS region: %s", region)
	}

	// Prepare parameters for session-manager-plugin
	params := map[string]interface{}{
		"Target": instanceID,
	}
	paramsJSON, err := json.Marshal(params)
	if err != nil {
		return fmt.Errorf("failed to marshal parameters: %w", err)
	}

	// Execute session-manager-plugin
	// #nosec G204 - region is validated above to prevent command injection
	cmd := execCommand(
		"session-manager-plugin",
		string(sessionJSON),
		region,
		"StartSession",
		"",
		string(paramsJSON),
		fmt.Sprintf("https://ssm.%s.amazonaws.com", region),
	)

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Handle interrupt signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigChan
		if cmd.Process != nil {
			if signalErr := cmd.Process.Signal(os.Interrupt); signalErr != nil {
				// Log the error but don't fail - we're already handling a signal
				fmt.Printf("Warning: failed to send interrupt signal: %v\n", signalErr)
			}
		}
	}()

	fmt.Printf("Starting session with instance %s...\n", instanceID)
	if runErr := cmd.Run(); runErr != nil {
		return fmt.Errorf("session-manager-plugin failed: %w", runErr)
	}

	return nil
}

// checkSessionManagerPlugin verifies that the session-manager-plugin is installed
func checkSessionManagerPlugin() error {
	_, err := execLookPath("session-manager-plugin")
	if err != nil {
		return fmt.Errorf("session-manager-plugin not found in PATH\n\n" +
			"The session-manager-plugin is required for plugin-based mode (--native=false).\n" +
			"By default, aws-ssm uses native mode which does NOT require the plugin.\n\n" +
			"To install the plugin, see:\n" +
			"https://docs.aws.amazon.com/systems-manager/latest/userguide/session-manager-working-with-install-plugin.html\n\n" +
			"Or use native mode (default): aws-ssm session <instance>")
	}
	return nil
}

// isValidAWSRegion validates that a region string is a valid AWS region format
func isValidAWSRegion(region string) bool {
	// AWS regions follow the pattern: [a-z]{2}(-[a-z]+)+-\d+[a-z]?
	// Examples: us-east-1, eu-west-1, ap-southeast-2, us-gov-east-1
	match, _ := regexp.MatchString(`^[a-z]{2}(-[a-z]+)+-\d+[a-z]?$`, region)
	return match
}
