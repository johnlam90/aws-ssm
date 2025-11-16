package tui

import (
	"context"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/johnlam90/aws-ssm/pkg/aws"
)

func TestNodeGroupUpdateKeyTriggersLaunchTemplateState(t *testing.T) {
	model := NewModel(context.Background(), &aws.Client{}, Config{})
	model.currentView = ViewNodeGroups
	model.nodeGroups = []NodeGroup{{
		ClusterName:           "cluster",
		Name:                  "nodegroup",
		LaunchTemplateID:      "lt-123",
		LaunchTemplateName:    "lt-name",
		LaunchTemplateVersion: "1",
	}}

	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'u'}})
	m, ok := updated.(Model)
	if !ok {
		t.Fatalf("expected Model after update, got %T", updated)
	}

	if m.ltUpdate == nil {
		t.Fatalf("expected launch template update state to be initialized")
	}

	if m.ltUpdate.LaunchTemplateID != "lt-123" {
		t.Fatalf("expected launch template ID to be preserved, got %s", m.ltUpdate.LaunchTemplateID)
	}
}

func TestNodeGroupShortcutHandler(t *testing.T) {
	model := NewModel(context.Background(), &aws.Client{}, Config{})
	model.currentView = ViewNodeGroups
	model.nodeGroups = []NodeGroup{{
		ClusterName:           "cluster",
		Name:                  "nodegroup",
		LaunchTemplateID:      "lt-123",
		LaunchTemplateName:    "lt-name",
		LaunchTemplateVersion: "1",
	}}

	handledModel, _, handled := model.handleNodeGroupLaunchTemplateShortcut(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'u'}})
	if !handled {
		t.Fatalf("expected shortcut handler to consume 'u' key")
	}

	m, ok := handledModel.(Model)
	if !ok {
		t.Fatalf("expected model to be returned from shortcut handler, got %T", handledModel)
	}

	if m.ltUpdate == nil {
		t.Fatalf("expected shortcut handler to initialize launch template state")
	}
}

func TestNavigationBindingForLaunchTemplateUpdate(t *testing.T) {
	nm := NewNavigationManager()
	tests := []struct {
		name string
		msg  tea.KeyMsg
	}{
		{
			name: "lowercase u",
			msg:  tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'u'}},
		},
		{
			name: "uppercase U",
			msg:  tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'U'}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			action := nm.HandleKey(tt.msg, ViewNodeGroups)
			if action != NavSelect {
				t.Fatalf("expected %q to map to NavSelect, got %v", tt.msg.String(), action)
			}
		})
	}
}
