package tui

import (
	"context"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/johnlam90/aws-ssm/pkg/aws"
)

func TestNewModel(t *testing.T) {
	ctx := context.Background()
	client := &aws.Client{}
	config := Config{
		Region:  "us-west-2",
		Profile: "test",
		NoColor: true,
	}

	model := NewModel(ctx, client, config)

	if model.ctx != ctx {
		t.Error("Model context should match input context")
	}
	if model.client != client {
		t.Error("Model client should match input client")
	}
	if model.config.Region != config.Region {
		t.Error("Model config region should match input config")
	}
	if model.currentView != ViewDashboard {
		t.Error("Model should start in dashboard view")
	}
	if len(model.menuItems) != 6 {
		t.Error("Model should have 6 menu items")
	}
	if model.cursor != 0 {
		t.Error("Model cursor should start at 0")
	}
}

func TestModelInit(t *testing.T) {
	ctx := context.Background()
	client := &aws.Client{}
	config := Config{}
	model := NewModel(ctx, client, config)

	cmd := model.Init()
	if cmd == nil {
		t.Error("Init should return a command")
	}
}

func TestNavigateBack(t *testing.T) {
	ctx := context.Background()
	client := &aws.Client{}
	config := Config{}
	model := NewModel(ctx, client, config)

	// Push a view onto the stack
	model.pushView(ViewEC2Instances)
	if model.currentView != ViewEC2Instances {
		t.Error("Current view should be EC2Instances after push")
	}
	if len(model.viewStack) != 1 {
		t.Error("View stack should have 1 item")
	}

	// Navigate back
	model = model.navigateBack()
	if model.currentView != ViewDashboard {
		t.Error("Current view should be Dashboard after navigateBack")
	}
	if len(model.viewStack) != 0 {
		t.Error("View stack should be empty after navigateBack")
	}

	// Test navigateBack when stack is empty
	model = model.navigateBack()
	if model.currentView != ViewDashboard {
		t.Error("Current view should remain Dashboard when stack is empty")
	}
}

func TestPushView(t *testing.T) {
	ctx := context.Background()
	client := &aws.Client{}
	config := Config{}
	model := NewModel(ctx, client, config)

	originalView := model.currentView
	model.pushView(ViewEKSClusters)

	if model.currentView != ViewEKSClusters {
		t.Error("Current view should be EKSClusters after push")
	}
	if len(model.viewStack) != 1 {
		t.Error("View stack should have 1 item")
	}
	if model.viewStack[0] != originalView {
		t.Error("View stack should contain original view")
	}
	if model.cursor != 0 {
		t.Error("Cursor should be reset to 0 after pushView")
	}
}

func TestHandleKeyPressQuit(t *testing.T) {
	ctx := context.Background()
	client := &aws.Client{}
	config := Config{}
	model := NewModel(ctx, client, config)

	// Test quit from dashboard
	msg := tea.KeyMsg{Type: tea.KeyCtrlC}
	_, cmd := model.Update(msg)
	if cmd == nil {
		t.Error("Ctrl+C should return a quit command")
	}

	// Test quit with 'q' from dashboard
	msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}}
	_, cmd = model.Update(msg)
	if cmd == nil {
		t.Error("'q' should return a quit command from dashboard")
	}

	// Test quit confirmation from sub-view
	model.pushView(ViewEC2Instances)
	msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}}
	newModel, _ := model.Update(msg)
	if newModel.(Model).statusMessage != "press q again to quit, esc to stay" {
		t.Error("First 'q' should show quit confirmation")
	}

	// Test second 'q' quits
	_, cmd = newModel.Update(msg)
	if cmd == nil {
		t.Error("Second 'q' should return a quit command")
	}
}

func TestHandleKeyPressHelp(t *testing.T) {
	ctx := context.Background()
	client := &aws.Client{}
	config := Config{}
	model := NewModel(ctx, client, config)

	// Test help toggle
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}}
	newModel, _ := model.Update(msg)
	if newModel.(Model).currentView != ViewHelp {
		t.Error("'?' should switch to help view")
	}

	// Test help toggle back
	_, cmd := newModel.Update(msg)
	if cmd != nil {
		t.Error("Second '?' should return to previous view")
	}
}

func TestHandleKeyPressSearch(t *testing.T) {
	ctx := context.Background()
	client := &aws.Client{}
	config := Config{}
	model := NewModel(ctx, client, config)

	// Test search activation
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}}
	newModel, _ := model.Update(msg)
	if !newModel.(Model).searchActive {
		t.Error("'/' should activate search")
	}
}

func TestGetStatusBar(t *testing.T) {
	ctx := context.Background()
	client := &aws.Client{}
	config := Config{
		Region:  "us-west-2",
		Profile: "test",
	}
	model := NewModel(ctx, client, config)
	model.currentView = ViewEC2Instances

	statusBar := model.getStatusBar()
	if statusBar == "" {
		t.Error("Status bar should not be empty")
	}
	if !contains(statusBar, "us-west-2") {
		t.Error("Status bar should contain region")
	}
	if !contains(statusBar, "test") {
		t.Error("Status bar should contain profile")
	}
	if !contains(statusBar, "EC2 Instances") {
		t.Error("Status bar should contain current view")
	}
}

func TestModelNavigation(t *testing.T) {
	ctx := context.Background()
	client := &aws.Client{}
	config := Config{}
	model := NewModel(ctx, client, config)

	// Test that model can navigate between views
	if model.currentView != ViewDashboard {
		t.Error("Model should start in dashboard view")
	}

	// Test navigation
	model.pushView(ViewEC2Instances)
	if model.currentView != ViewEC2Instances {
		t.Error("Model should navigate to EC2 instances view")
	}

	// Test navigation back
	model = model.navigateBack()
	if model.currentView != ViewDashboard {
		t.Error("Model should navigate back to dashboard")
	}
}

// Helper function
func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
