package tui

import (
	"testing"

	"github.com/charmbracelet/lipgloss"
)

func TestNoColorMode(t *testing.T) {
	// Test that RenderSelectableRow works correctly
	row := "test row"
	selectedRow := RenderSelectableRow(row, true)
	if !contains(selectedRow, row) {
		t.Errorf("Expected selected row to contain '%s', got '%s'", row, selectedRow)
	}

	// Test RenderSelectableRow without selection
	normalRow := RenderSelectableRow(row, false)
	if normalRow != row {
		t.Errorf("Expected '%s', got '%s'", row, normalRow)
	}
}

func TestGetStateColor(t *testing.T) {
	tests := []struct {
		state    string
		expected lipgloss.Color
	}{
		{"running", ColorRunning},
		{"stopped", ColorStopped},
		{"pending", ColorPending},
		{"stopping", ColorPending},
		{"terminated", ColorTerminated},
		{"shutting-down", ColorTerminated},
		{"unknown", ColorMuted},
	}

	for _, tt := range tests {
		t.Run(tt.state, func(t *testing.T) {
			result := GetStateColor(tt.state)
			if result != tt.expected {
				t.Errorf("GetStateColor(%s) = %v, want %v", tt.state, result, tt.expected)
			}
		})
	}
}

func TestStateStyle(t *testing.T) {
	tests := []struct {
		state    string
		contains string
	}{
		{"running", "running"},
		{"stopped", "stopped"},
		{"pending", "pending"},
		{"terminated", "terminated"},
	}

	for _, tt := range tests {
		t.Run(tt.state, func(t *testing.T) {
			result := StateStyle(tt.state)
			if !contains(result, tt.contains) {
				t.Errorf("StateStyle(%s) should contain '%s'", tt.state, tt.contains)
			}
		})
	}
}

func TestRenderStatusMessage(t *testing.T) {
	tests := []struct {
		message     string
		messageType string
		contains    string
	}{
		{"Success message", "success", "Success message"},
		{"Error message", "error", "Error message"},
		{"Warning message", "warning", "Warning message"},
		{"Info message", "info", "Info message"},
		{"Default message", "default", "Default message"},
	}

	for _, tt := range tests {
		t.Run(tt.messageType, func(t *testing.T) {
			result := RenderStatusMessage(tt.message, tt.messageType)
			if !contains(result, tt.contains) {
				t.Errorf("RenderStatusMessage(%s, %s) should contain '%s'", tt.message, tt.messageType, tt.contains)
			}
		})
	}
}

func TestRenderMetric(t *testing.T) {
	label := "CPU"
	value := "75%"
	
	// Test normal metric
	result := RenderMetric(label, value, false)
	if !contains(result, label) || !contains(result, value) {
		t.Errorf("RenderMetric should contain both label and value")
	}
	
	// Test highlighted metric
	highlighted := RenderMetric(label, value, true)
	if !contains(highlighted, label) || !contains(highlighted, value) {
		t.Errorf("Highlighted RenderMetric should contain both label and value")
	}
}