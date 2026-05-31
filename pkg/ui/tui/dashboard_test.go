package tui

import (
	"context"
	"strings"
	"testing"
)

func TestBeautifulDashboardRendering(t *testing.T) {
	// Create a test model
	ctx := context.Background()
	config := Config{
		Region:  "us-west-2",
		Profile: "test-profile",
		NoColor: false,
	}

	model := NewModel(ctx, nil, config)
	model.width = 120
	model.height = 30
	model.ready = true

	// Phase 2: chrome (top bar, sidebar, bottom hint bar) is composed
	// at the View() level. Assert on the full screen so we capture
	// chrome contributions like region, profile, navigation hints.
	view := model.View()

	if !strings.Contains(view, "us-west-2") {
		t.Error("View should contain region information")
	}
	if !strings.Contains(view, "test-profile") {
		t.Error("View should contain profile information")
	}
	if !strings.Contains(view, "aws-ssm") {
		t.Error("View should contain app brand")
	}

	// Tips block surfaces the palette + filter + help hotkeys.
	for _, hint := range []string{"command palette", "filter the focused list", "show keybindings", "refresh"} {
		if !strings.Contains(view, hint) {
			t.Errorf("Home view should contain tip %q", hint)
		}
	}
}

// Selection rendering test removed — Home no longer has a rollup, so
// there's nothing to "select" via cursor on the Home view. Selection
// is exercised by the per-resource list view tests.

func TestDashboardDescriptionNormalization(t *testing.T) {
	ctx := context.Background()
	config := Config{}
	model := NewModel(ctx, nil, config)

	// Test description normalization
	tests := []struct {
		input    string
		expected string
	}{
		{"View and manage EC2 instances", "Manage EC2 instances"},
		{"View and scale ASGs", "Scale and monitor ASGs"},
		{"View EC2 network interfaces and ENIs", "View EC2 ENIs"},
		{"Unknown description", "Unknown description"},
	}

	for _, test := range tests {
		result := model.normalizeServiceDescription(test.input)
		if result != test.expected {
			t.Errorf("Description normalization failed: input='%s', expected='%s', got='%s'",
				test.input, test.expected, result)
		}
	}
}
