package cmd

import (
	"context"
	"testing"

	"github.com/johnlam90/aws-ssm/pkg/aws"
)

// TestStartPendingSSMSession_UsesNativeImplementation verifies that launching a
// session from the TUI uses the native (pure Go) implementation rather than the
// plugin-based path. The plugin path requires session-manager-plugin to be
// installed, but native mode is the documented default for `aws-ssm session`,
// so the TUI must behave the same way.
//
// Regression test for the TUI calling client.StartSession (plugin) instead of
// client.StartNativeSession, which surfaced as:
//
//	"The session-manager-plugin is required for plugin-based mode (--native=false)."
func TestStartPendingSSMSession_UsesNativeImplementation(t *testing.T) {
	var nativeCalled, pluginCalled bool
	var gotInstanceID string

	mock := &aws.MockSSMOperations{
		StartNativeSessionFunc: func(_ context.Context, instanceID string) error {
			nativeCalled = true
			gotInstanceID = instanceID
			return nil
		},
		StartSessionFunc: func(_ context.Context, _ string) error {
			pluginCalled = true
			return nil
		},
	}

	const wantInstanceID = "i-1234567890abcdef0"
	if err := startPendingSSMSession(context.Background(), mock, wantInstanceID); err != nil {
		t.Fatalf("startPendingSSMSession returned error: %v", err)
	}

	if pluginCalled {
		t.Error("TUI launched the plugin-based session (StartSession); it must use the " +
			"native implementation so session-manager-plugin is not required")
	}
	if !nativeCalled {
		t.Fatal("TUI did not launch the native session (StartNativeSession)")
	}
	if gotInstanceID != wantInstanceID {
		t.Errorf("native session launched with instance %q, want %q", gotInstanceID, wantInstanceID)
	}
}
