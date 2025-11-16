package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// NavigationKey represents different types of navigation actions
type NavigationKey int

const (
	NavUp NavigationKey = iota
	NavDown
	NavLeft
	NavRight
	NavPageUp
	NavPageDown
	NavHome
	NavEnd
	NavBack
	NavSelect
	NavSearch
	NavRefresh
	NavHelp
	NavQuit
	NavScale
	NavSSH
	NavDetails
	NavFilter
)

// KeyBinding represents a keyboard shortcut
type KeyBinding struct {
	Key         string
	Description string
	Action      NavigationKey
}

// CommandMode represents different command modes for chaining
type CommandMode int

const (
	CmdModeNormal CommandMode = iota
	CmdModeVisual
	CmdModeSearch
	CmdModeFilter
	CmdModeScale
)

// NavigationManager handles keyboard shortcuts and command chaining
type NavigationManager struct {
	mode           CommandMode
	commandBuffer  string
	lastNavigation NavigationKey
}

// Global keybindings for all views
var globalKeyBindings = []KeyBinding{
	{Key: "ctrl+c", Description: "Quit immediately", Action: NavQuit},
	{Key: "q", Description: "Quit (with confirmation)", Action: NavQuit},
	{Key: "?", Description: "Toggle help", Action: NavHelp},
	{Key: "/", Description: "Search", Action: NavSearch},
	{Key: "r, ctrl+r", Description: "Refresh data", Action: NavRefresh},
	{Key: "esc", Description: "Back/Cancel", Action: NavBack},
}

// View-specific keybindings
var viewKeyBindings = map[ViewMode][]KeyBinding{
	ViewDashboard: {
		{Key: "up, k", Description: "Move up", Action: NavUp},
		{Key: "down, j", Description: "Move down", Action: NavDown},
		{Key: "enter, space", Description: "Select", Action: NavSelect},
	},
	ViewEC2Instances: {
		{Key: "up, k", Description: "Move up", Action: NavUp},
		{Key: "down, j", Description: "Move down", Action: NavDown},
		{Key: "g g", Description: "Go to top", Action: NavHome},
		{Key: "G", Description: "Go to bottom", Action: NavEnd},
		{Key: "ctrl+u", Description: "Page up", Action: NavPageUp},
		{Key: "ctrl+d", Description: "Page down", Action: NavPageDown},
		{Key: "enter, space", Description: "Connect via SSM", Action: NavSSH},
		{Key: "d", Description: "Show details", Action: NavDetails},
		{Key: "s", Description: "Scale instance", Action: NavScale},
		{Key: "f", Description: "Filter by state", Action: NavFilter},
	},
	ViewEKSClusters: {
		{Key: "up, k", Description: "Move up", Action: NavUp},
		{Key: "down, j", Description: "Move down", Action: NavDown},
		{Key: "g g", Description: "Go to top", Action: NavHome},
		{Key: "G", Description: "Go to bottom", Action: NavEnd},
		{Key: "enter, space", Description: "Show node groups", Action: NavSelect},
		{Key: "d", Description: "Show details", Action: NavDetails},
	},
	ViewASGs: {
		{Key: "up, k", Description: "Move up", Action: NavUp},
		{Key: "down, j", Description: "Move down", Action: NavDown},
		{Key: "g g", Description: "Go to top", Action: NavHome},
		{Key: "G", Description: "Go to bottom", Action: NavEnd},
		{Key: "enter, space", Description: "Scale ASG", Action: NavScale},
		{Key: "d", Description: "Show details", Action: NavDetails},
	},
	ViewNodeGroups: {
		{Key: "up, k", Description: "Move up", Action: NavUp},
		{Key: "down, j", Description: "Move down", Action: NavDown},
		{Key: "g g", Description: "Go to top", Action: NavHome},
		{Key: "G", Description: "Go to bottom", Action: NavEnd},
		{Key: "enter, space", Description: "Scale node group", Action: NavScale},
		{Key: "u, U", Description: "Update launch template", Action: NavSelect},
		{Key: "d", Description: "Show details", Action: NavDetails},
	},
	ViewNetworkInterfaces: {
		{Key: "up, k", Description: "Move up", Action: NavUp},
		{Key: "down, j", Description: "Move down", Action: NavDown},
		{Key: "g g", Description: "Go to top", Action: NavHome},
		{Key: "G", Description: "Go to bottom", Action: NavEnd},
		{Key: "d", Description: "Show details", Action: NavDetails},
	},
}

// NewNavigationManager creates a new navigation manager
func NewNavigationManager() *NavigationManager {
	return &NavigationManager{
		mode:          CmdModeNormal,
		commandBuffer: "",
	}
}

// HandleKey processes keyboard input and returns navigation action
func (nm *NavigationManager) HandleKey(msg tea.KeyMsg, currentView ViewMode) NavigationKey {
	keyStr := msg.String()

	// Handle command mode transitions
	switch nm.mode {
	case CmdModeSearch:
		if keyStr == "esc" {
			nm.mode = CmdModeNormal
			return NavBack
		}
		return NavSearch
	case CmdModeFilter:
		if keyStr == "esc" {
			nm.mode = CmdModeNormal
			return NavBack
		}
		return NavFilter
	case CmdModeScale:
		if keyStr == "esc" {
			nm.mode = CmdModeNormal
			return NavBack
		}
		return NavScale
	}

	// Handle vim-style command chaining
	nm.commandBuffer += keyStr

	// Check for vim-style commands
	switch nm.commandBuffer {
	case "g":
		// Wait for next key
		return nm.lastNavigation
	case "g g":
		nm.commandBuffer = ""
		return NavHome
	case "G":
		nm.commandBuffer = ""
		return NavEnd
	}

	// Single key commands
	nm.commandBuffer = ""

	// Check global bindings first
	for _, binding := range globalKeyBindings {
		if nm.matchesKey(keyStr, binding.Key) {
			return binding.Action
		}
	}

	// Check view-specific bindings
	if bindings, exists := viewKeyBindings[currentView]; exists {
		for _, binding := range bindings {
			if nm.matchesKey(keyStr, binding.Key) {
				return binding.Action
			}
		}
	}

	return nm.lastNavigation
}

// matchesKey checks if a key matches a binding (handles comma-separated keys)
func (nm *NavigationManager) matchesKey(input, binding string) bool {
	// Handle comma-separated alternatives like "up, k"
	if strings.Contains(binding, ",") {
		keys := strings.Split(binding, ",")
		for _, key := range keys {
			key = strings.TrimSpace(key)
			if input == key {
				return true
			}
		}
		return false
	}
	return input == binding
}

// SetMode sets the current command mode
func (nm *NavigationManager) SetMode(mode CommandMode) {
	nm.mode = mode
}

// GetMode returns the current command mode
func (nm *NavigationManager) GetMode() CommandMode {
	return nm.mode
}

// GetKeyBindings returns keybindings for a specific view
func GetKeyBindings(view ViewMode) []KeyBinding {
	var bindings []KeyBinding

	// Add view-specific bindings
	if viewBindings, exists := viewKeyBindings[view]; exists {
		bindings = append(bindings, viewBindings...)
	}

	// Add global bindings
	bindings = append(bindings, globalKeyBindings...)

	return bindings
}

// FormatKeyBindings formats keybindings for display
func FormatKeyBindings(bindings []KeyBinding) string {
	var result strings.Builder

	for _, binding := range bindings {
		key := HelpKeyStyle().Render(binding.Key)
		desc := binding.Description
		result.WriteString(fmt.Sprintf("  %-20s %s\n", key, desc))
	}

	return result.String()
}

// GetQuickReference returns a quick reference for common actions
func GetQuickReference() string {
	var result strings.Builder

	result.WriteString(TitleStyle().Render("Quick Reference") + "\n\n")

	sections := []struct {
		title    string
		bindings []KeyBinding
	}{
		{
			title: "Navigation",
			bindings: []KeyBinding{
				{Key: "j/↓", Description: "Move down"},
				{Key: "k/↑", Description: "Move up"},
				{Key: "g g", Description: "Go to top"},
				{Key: "G", Description: "Go to bottom"},
				{Key: "ctrl+d", Description: "Page down"},
				{Key: "ctrl+u", Description: "Page up"},
			},
		},
		{
			title: "Actions",
			bindings: []KeyBinding{
				{Key: "enter", Description: "Select/Connect"},
				{Key: "d", Description: "Show details"},
				{Key: "s", Description: "Scale"},
				{Key: "/", Description: "Search"},
				{Key: "f", Description: "Filter"},
			},
		},
		{
			title: "General",
			bindings: []KeyBinding{
				{Key: "?", Description: "Help"},
				{Key: "esc", Description: "Back/Cancel"},
				{Key: "r, ctrl+r", Description: "Refresh"},
				{Key: "q", Description: "Quit"},
			},
		},
	}

	for _, section := range sections {
		result.WriteString(SubtitleStyle().Render(section.title) + "\n")
		for _, binding := range section.bindings {
			key := HelpKeyStyle().Render(binding.Key)
			result.WriteString(fmt.Sprintf("  %-12s %s\n", key, binding.Description))
		}
		result.WriteString("\n")
	}

	return result.String()
}
