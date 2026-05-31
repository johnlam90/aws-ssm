package tui

import (
	"fmt"
	"strings"
)

// renderDashboard renders the dashboard's main panel content.
//
// Phase 2 of the foundation redesign: the dashboard no longer renders
// its own header bar, separator, footer, or status bar — the chrome
// owns those. This renderer only emits the section title and the menu
// items.
func (m Model) renderDashboard() string {
	var b strings.Builder

	if m.loading {
		b.WriteString(m.renderLoading())
		return b.String()
	}
	if m.err != nil {
		b.WriteString(m.renderError())
		return b.String()
	}

	sectionTitle := DashboardSectionTitleStyle().Render("Services")
	b.WriteString(sectionTitle)
	b.WriteString("\n\n")

	for i, item := range m.menuItems {
		menuItem := m.renderDashboardMenuItem(i, item, i == m.cursor)
		b.WriteString(menuItem)
		b.WriteString("\n")
	}

	return b.String()
}

// renderLoading renders an inline loading message used by per-view
// renderers when their data is still loading.
func (m Model) renderLoading() string {
	if !m.loading {
		return ""
	}
	frame := "⠋"
	msg := fmt.Sprintf("%s %s", frame, m.loadingMsg)
	return LoadingStyle().Render(msg)
}

// renderError renders an inline error message.
func (m Model) renderError() string {
	if m.err == nil {
		return ""
	}
	return ErrorStyle().Render(fmt.Sprintf("Error: %v", m.err))
}

// renderDashboardMenuItem renders one menu row in the dashboard.
func (m Model) renderDashboardMenuItem(_ int, item MenuItem, isSelected bool) string {
	normalizedDesc := m.normalizeServiceDescription(item.Description)

	if isSelected {
		selectionBar := DashboardSelectionBarStyle().Render("▌")
		name := DashboardSelectedNameStyle().Render(item.Title)
		desc := DashboardSelectedDescStyle().Render(normalizedDesc)
		return fmt.Sprintf("%s %s %s", selectionBar, name, desc)
	}
	name := DashboardServiceNameStyle().Render(item.Title)
	desc := DashboardServiceDescStyle().Render(normalizedDesc)
	return fmt.Sprintf("  %s %s", name, desc)
}

// normalizeServiceDescription normalizes service descriptions for consistency.
func (m Model) normalizeServiceDescription(desc string) string {
	normalizedMap := map[string]string{
		"View and manage EC2 instances":        "Manage EC2 instances",
		"Manage EKS clusters and node groups":  "Manage EKS clusters & node groups",
		"View and scale ASGs":                  "Scale and monitor ASGs",
		"Inspect managed node groups":          "Inspect managed node groups",
		"View EC2 network interfaces and ENIs": "View EC2 ENIs",
		"View keybindings and help":            "Keybindings and documentation",
	}

	if normalized, exists := normalizedMap[desc]; exists {
		return normalized
	}
	return desc
}

// getRegion returns the current AWS region.
func (m Model) getRegion() string {
	if m.config.Region != "" {
		return m.config.Region
	}
	return "us-east-1"
}

// getProfile returns the current AWS profile.
func (m Model) getProfile() string {
	if m.config.Profile != "" {
		return m.config.Profile
	}
	return "default"
}
