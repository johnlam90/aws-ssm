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

	if !strings.Contains(view, "Services") {
		t.Error("Dashboard should contain 'Services' section title")
	}

	if !strings.Contains(view, "us-west-2") {
		t.Error("View should contain region information (chrome top bar)")
	}

	if !strings.Contains(view, "test-profile") {
		t.Error("View should contain profile information (chrome top bar)")
	}

	if !strings.Contains(view, "aws-ssm") {
		t.Error("View should contain app brand (chrome top bar)")
	}

	if !strings.Contains(view, "Home") {
		t.Error("View should contain breadcrumb 'Home' (chrome top bar)")
	}

	// Service tiles are rendered by the dashboard's main panel.
	services := []string{"EC2 Instances", "EKS Clusters", "Auto Scaling Groups", "EKS Node Groups", "Network Interfaces", "Help"}
	for _, service := range services {
		if !strings.Contains(view, service) {
			t.Errorf("View should contain service '%s'", service)
		}
	}

	// Bottom hint bar carries navigation hints.
	for _, hint := range []string{"navigate", "select", "search", "help", "quit"} {
		if !strings.Contains(view, hint) {
			t.Errorf("View should contain hint label %q (bottom hint bar)", hint)
		}
	}
}

func TestDashboardSelectionRendering(t *testing.T) {
	// Create a test model with selection
	ctx := context.Background()
	config := Config{NoColor: false}

	model := NewModel(ctx, nil, config)
	model.width = 120
	model.height = 30
	model.ready = true
	model.cursor = 2 // Select third item (Auto Scaling Groups)

	view := model.renderDashboard()

	// Verify selection indicator is present
	if !strings.Contains(view, "▌") {
		t.Error("Selected item should show vertical bar indicator")
	}
}

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
