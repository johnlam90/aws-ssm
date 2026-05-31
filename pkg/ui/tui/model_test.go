package tui

import (
	"context"
	"fmt"
	"strings"
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
	// Phase 9: Home view's roll-up is the four resource types.
	// Help moved to overlay/sidebar and is no longer a menu entry.
	if len(model.menuItems) != 4 {
		t.Errorf("Model should have 4 menu items, got %d", len(model.menuItems))
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

func TestListNavigation_EmptyListsKeepCursorValid(t *testing.T) {
	ctx := context.Background()
	client := &aws.Client{}
	config := Config{}

	tests := []struct {
		name    string
		view    ViewMode
		actions []NavigationKey
	}{
		{name: "ec2", view: ViewEC2Instances, actions: []NavigationKey{NavEnd, NavPageDown, NavSSH, NavDetails}},
		{name: "eks", view: ViewEKSClusters, actions: []NavigationKey{NavEnd, NavPageDown, NavSelect, NavDetails}},
		{name: "asg", view: ViewASGs, actions: []NavigationKey{NavEnd, NavPageDown, NavScale, NavDetails}},
		{name: "node_groups", view: ViewNodeGroups, actions: []NavigationKey{NavEnd, NavPageDown, NavScale, NavSelect, NavDetails}},
		{name: "network", view: ViewNetworkInterfaces, actions: []NavigationKey{NavEnd, NavPageDown, NavDetails}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := NewModel(ctx, client, config)
			model.currentView = tt.view

			defer func() {
				if r := recover(); r != nil {
					t.Fatalf("empty %s navigation panicked: %v", tt.view, r)
				}
			}()

			for _, action := range tt.actions {
				updated, _ := model.handleNavigation(action)
				model = updated.(Model)
				if model.cursor < 0 {
					t.Fatalf("cursor should stay non-negative after %v, got %d", action, model.cursor)
				}
			}
		})
	}
}

func TestRenderNodeGroups_FitsTerminalHeightWhenScrolled(t *testing.T) {
	model := NewModel(context.Background(), &aws.Client{}, Config{})
	model.currentView = ViewNodeGroups
	model.ready = true
	model.width = 120
	model.height = 18
	model.cursor = 15
	model.nodeGroups = make([]NodeGroup, 23)
	for i := range model.nodeGroups {
		model.nodeGroups[i] = NodeGroup{
			ClusterName: fmt.Sprintf("cluster-%02d", i),
			Name:        fmt.Sprintf("node-group-%02d", i),
			Status:      "ACTIVE",
			Version:     "1.29",
			DesiredSize: 3,
			MinSize:     1,
			MaxSize:     5,
			CurrentSize: 3,
		}
	}

	view := model.renderNodeGroups()
	lines := strings.Split(strings.TrimSuffix(view, "\n"), "\n")
	if len(lines) > model.height {
		t.Fatalf("node-group view rendered %d lines for height %d", len(lines), model.height)
	}
	if !strings.Contains(view, "CLUSTER") || !strings.Contains(view, "NODE GROUP") {
		t.Fatalf("node-group table header should remain rendered")
	}
}

func TestTableViews_FitTerminalHeightWhenScrolled(t *testing.T) {
	tests := []struct {
		name       string
		setup      func() Model
		render     func(Model) string
		headerText []string
	}{
		{
			name: "ec2",
			setup: func() Model {
				model := newScrolledRenderModel(ViewEC2Instances)
				model.ec2Instances = make([]EC2Instance, 23)
				for i := range model.ec2Instances {
					model.ec2Instances[i] = EC2Instance{
						Name:             fmt.Sprintf("instance-%02d", i),
						InstanceID:       fmt.Sprintf("i-%017d", i),
						State:            "running",
						PrivateIP:        fmt.Sprintf("10.0.0.%d", i+1),
						InstanceType:     "t3.micro",
						AvailabilityZone: "us-east-1a",
					}
				}
				return model
			},
			render:     Model.renderEC2Instances,
			headerText: []string{"NAME", "INSTANCE ID"},
		},
		{
			name: "eks",
			setup: func() Model {
				model := newScrolledRenderModel(ViewEKSClusters)
				model.eksClusters = make([]EKSCluster, 23)
				for i := range model.eksClusters {
					model.eksClusters[i] = EKSCluster{
						Name:    fmt.Sprintf("cluster-%02d", i),
						Status:  "ACTIVE",
						Version: "1.29",
					}
				}
				return model
			},
			render:     Model.renderEKSClusters,
			headerText: []string{"NAME", "STATE", "VERSION"},
		},
		{
			name: "asg",
			setup: func() Model {
				model := newScrolledRenderModel(ViewASGs)
				model.asgs = make([]ASG, 23)
				for i := range model.asgs {
					model.asgs[i] = ASG{
						Name:            fmt.Sprintf("asg-%02d", i),
						DesiredCapacity: 3,
						MinSize:         1,
						MaxSize:         5,
						CurrentSize:     3,
					}
				}
				return model
			},
			render:     Model.renderASGs,
			headerText: []string{"NAME", "DES", "CUR"},
		},
		{
			name: "network",
			setup: func() Model {
				model := newScrolledRenderModel(ViewNetworkInterfaces)
				model.netInterfaces = make([]aws.InstanceInterfaces, 23)
				for i := range model.netInterfaces {
					model.netInterfaces[i] = aws.InstanceInterfaces{
						InstanceID:   fmt.Sprintf("i-%017d", i),
						InstanceName: fmt.Sprintf("instance-%02d", i),
						DNSName:      fmt.Sprintf("ip-10-0-0-%d.ec2.internal", i+1),
					}
				}
				return model
			},
			render:     Model.renderNetworkInterfaces,
			headerText: []string{"NAME", "INSTANCE ID", "IFACES"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := tt.setup()
			view := tt.render(model)
			lines := strings.Split(strings.TrimSuffix(view, "\n"), "\n")
			if len(lines) > model.height {
				t.Fatalf("%s view rendered %d lines for height %d", tt.name, len(lines), model.height)
			}
			for _, header := range tt.headerText {
				if !strings.Contains(view, header) {
					t.Fatalf("%s table header should contain %q", tt.name, header)
				}
			}
		})
	}
}

func newScrolledRenderModel(view ViewMode) Model {
	model := NewModel(context.Background(), &aws.Client{}, Config{})
	model.currentView = view
	model.ready = true
	model.width = 120
	model.height = 18
	model.cursor = 15
	return model
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
