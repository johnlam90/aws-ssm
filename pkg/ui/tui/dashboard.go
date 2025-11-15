package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// renderDashboard renders the main dashboard view
func (m Model) renderDashboard() string {
	var b strings.Builder

	// Header
	header := m.renderHeader("AWS SSM Manager", "Select a service to manage")
	b.WriteString(header)
	b.WriteString("\n\n")

	// Menu items - minimal, no icons
    for i, item := range m.menuItems {
        var style lipgloss.Style
        if i == m.cursor {
            style = MenuItemSelectedStyle()
        } else {
            style = MenuItemStyle()
        }

		// Simple format: title and description on same line
		title := style.Render(item.Title)
        desc := SubtitleStyle().Render(" - " + item.Description)

		b.WriteString("  " + title + desc)
		b.WriteString("\n")
	}

	// Footer with keybindings
	b.WriteString("\n")
	b.WriteString(m.renderFooter())

	// Status bar
	b.WriteString("\n")
    b.WriteString(StatusBarStyle().Width(m.width).Render(m.getStatusBar()))

	return b.String()
}

// handleDashboardKeys handles keyboard input for the dashboard
func (m Model) handleDashboardKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}

	case "down", "j":
		if m.cursor < len(m.menuItems)-1 {
			m.cursor++
		}

	case "enter", " ":
		// Navigate to selected view
		selectedItem := m.menuItems[m.cursor]
		m.pushView(selectedItem.View)

		// Load data for the selected view
		var cmd tea.Cmd
		switch selectedItem.View {
		case ViewEC2Instances:
			m.loading = true
			m.loadingMsg = "Loading EC2 instances..."
			cmd = LoadEC2InstancesCmd(m.ctx, m.client)
		case ViewEKSClusters:
			m.loading = true
			m.loadingMsg = "Loading EKS clusters..."
			cmd = LoadEKSClustersCmd(m.ctx, m.client)
		case ViewASGs:
			m.loading = true
			m.loadingMsg = "Loading Auto Scaling Groups..."
			cmd = LoadASGsCmd(m.ctx, m.client)
		case ViewNodeGroups:
			m.loading = true
			m.loadingMsg = "Loading EKS node groups..."
			cmd = LoadNodeGroupsCmd(m.ctx, m.client)
		case ViewNetworkInterfaces:
			m.loading = true
			m.loadingMsg = "Loading network interfaces..."
			cmd = LoadNetworkInterfacesCmd(m.ctx, m.client)
		}

		return m, cmd

	case "esc":
		return m, tea.Quit
	}

	return m, nil
}

// renderHeader renders a minimal header
func (m Model) renderHeader(title, subtitle string) string {
	var b strings.Builder

	// Title - simple
    titleText := TitleStyle().Render(title)
	b.WriteString(titleText)

	// Subtitle on same line
	if subtitle != "" {
        subtitleText := SubtitleStyle().Render(" - " + subtitle)
		b.WriteString(subtitleText)
	}

	return b.String()
}

// renderFooter renders a minimal footer with keybindings
func (m Model) renderFooter() string {
	keys := []struct {
		key  string
		desc string
	}{
		{"↑/k", "up"},
		{"↓/j", "down"},
		{"enter", "select"},
		{"?", "help"},
		{"q", "quit"},
	}

	var parts []string
	for _, k := range keys {
		parts = append(parts, fmt.Sprintf("%s:%s", k.key, k.desc))
	}

    return SubtitleStyle().Render(strings.Join(parts, "  "))
}

// renderLoading renders a loading message
func (m Model) renderLoading() string {
	if !m.loading {
		return ""
	}

	spinner := "⠋⠙⠹⠸⠼⠴⠦⠧⠇⠏" // Braille spinner
	// In a real implementation, we'd animate this
	frame := string(spinner[0])

	msg := fmt.Sprintf("%s %s", frame, m.loadingMsg)
    return LoadingStyle().Render(msg)
}

// renderError renders an error message
func (m Model) renderError() string {
	if m.err == nil {
		return ""
	}

    return ErrorStyle().Render(fmt.Sprintf("Error: %v", m.err))
}
