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

	// Test dashboard rendering
	view := model.renderDashboard()

	// Verify key components are present
	if !strings.Contains(view, "AWS SSM Manager") {
		t.Error("Dashboard should contain title 'AWS SSM Manager'")
	}

	if !strings.Contains(view, "Region: us-west-2") {
		t.Error("Dashboard should contain region information")
	}

	if !strings.Contains(view, "Profile: test-profile") {
		t.Error("Dashboard should contain profile information")
	}

	if !strings.Contains(view, "Services") {
		t.Error("Dashboard should contain 'Services' section title")
	}

	// Verify services are present
	services := []string{"EC2 Instances", "EKS Clusters", "Auto Scaling Groups", "EKS Node Groups", "Network Interfaces", "Help"}
	for _, service := range services {
		if !strings.Contains(view, service) {
			t.Errorf("Dashboard should contain service '%s'", service)
		}
	}

	// Verify navigation hints
	if !strings.Contains(view, "Navigation:") {
		t.Error("Dashboard should contain navigation hints")
	}

	if !strings.Contains(view, "↑/k Up") {
		t.Error("Dashboard should contain up navigation hint")
	}

	if !strings.Contains(view, "↓/j Down") {
		t.Error("Dashboard should contain down navigation hint")
	}

	// Verify separator lines are present
	if !strings.Contains(view, "────────────────────────────────────────────────────────────────────────────────") {
		t.Error("Dashboard should contain horizontal separator lines")
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
