package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// renderDashboard renders the main dashboard view.
func (m Model) renderDashboard() string {
	var b strings.Builder

	b.WriteString(m.renderDashboardTopBar())
	b.WriteString("\n")
	b.WriteString(m.renderDashboardSeparator())
	b.WriteString("\n\n")
	b.WriteString(DashboardTitleStyle().Render("AWS SSM Manager"))
	b.WriteString("\n")
	b.WriteString(SubtitleStyle().Render("Choose a resource group to inspect or operate."))
	b.WriteString("\n\n")

	if m.loading {
		b.WriteString(m.renderLoading())
		return b.String()
	}

	if m.err != nil {
		b.WriteString(m.renderError())
		return b.String()
	}

	b.WriteString(DashboardTableHeaderStyle().Render(fmt.Sprintf("  %-4s %-24s %s", "KEY", "SERVICE", "ACTION")))
	b.WriteString("\n")
	b.WriteString(DashboardSubtleRule(m.dashboardContentWidth()).Render(strings.Repeat("─", m.dashboardContentWidth())))
	b.WriteString("\n")

	for i, item := range m.menuItems {
		menuItem := m.renderDashboardMenuItem(i, item, i == m.cursor)
		b.WriteString(menuItem)
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(m.renderDashboardFooter())

	b.WriteString("\n")
	b.WriteString(m.renderDashboardSeparator())

	return b.String()
}

// renderHeader renders an enhanced header with modern styling
func (m Model) renderHeader(title, subtitle string) string {
	var b strings.Builder

	// Title with sophisticated styling
	titleText := TitleStyle().Render(title)
	b.WriteString(titleText)

	// Subtitle with en-dash separator for visual consistency
	if subtitle != "" {
		subtitleText := SubtitleStyle().Render(" – " + subtitle)
		b.WriteString(subtitleText)
	}

	return b.String()
}

// renderLoading renders an enhanced loading message with modern styling
func (m Model) renderLoading() string {
	if !m.loading {
		return ""
	}

	// Use the same sophisticated spinner as other views
	spinner := "⠋⠙⠹⠸⠼⠴⠦⠧⠇⠏" // Braille spinner
	// In a real implementation, we'd animate this
	frame := string(spinner[0])

	msg := fmt.Sprintf("%s %s", frame, m.loadingMsg)
	return LoadingStyle().Render(msg)
}

// renderError renders an enhanced error message with consistent styling
func (m Model) renderError() string {
	if m.err == nil {
		return ""
	}

	return ErrorStyle().Render(fmt.Sprintf("Error: %v", m.err))
}

// renderDashboardTopBar renders the single-line app/context chrome.
func (m Model) renderDashboardTopBar() string {
	width := m.dashboardContentWidth()
	left := DashboardBrandStyle().Render("aws-ssm") + DashboardMutedStyle().Render("  dashboard")
	right := DashboardContextStyle().Render(fmt.Sprintf("%s  %s", m.getRegion(), m.getProfile()))

	gap := width - lipgloss.Width(left) - lipgloss.Width(right)
	if gap < 1 {
		return left
	}

	return left + strings.Repeat(" ", gap) + right
}

// renderDashboardSeparator renders a horizontal separator line
func (m Model) renderDashboardSeparator() string {
	separator := strings.Repeat("─", m.dashboardContentWidth())
	return DashboardSeparatorStyle().Render(separator)
}

// renderDashboardMenuItem renders one service row.
func (m Model) renderDashboardMenuItem(index int, item MenuItem, isSelected bool) string {
	normalizedDesc := m.normalizeServiceDescription(item.Description)
	key := fmt.Sprintf("%d", index+1)

	if isSelected {
		return fmt.Sprintf("%s %s %s %s",
			DashboardSelectionBarStyle().Render("›"),
			DashboardSelectedKeyStyle().Render(fmt.Sprintf("%-4s", key)),
			DashboardSelectedNameStyle().Render(fmt.Sprintf("%-24s", item.Title)),
			DashboardSelectedDescStyle().Render(normalizedDesc),
		)
	}

	name := DashboardServiceNameStyle().Render(fmt.Sprintf("%-24s", item.Title))
	desc := DashboardServiceDescStyle().Render(normalizedDesc)
	return fmt.Sprintf("  %s %s %s", DashboardFooterKeyStyle().Render(fmt.Sprintf("%-4s", key)), name, desc)
}

// normalizeServiceDescription normalizes service descriptions for consistency
func (m Model) normalizeServiceDescription(desc string) string {
	// Map of original descriptions to normalized ones
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

// renderDashboardFooter renders concise keyboard hints.
func (m Model) renderDashboardFooter() string {
	keys := []struct {
		key  string
		desc string
	}{
		{"↑/k", "Up"},
		{"↓/j", "Down"},
		{"enter", "Select"},
		{"?", "Help"},
		{"q", "Quit"},
	}

	var parts []string
	for _, k := range keys {
		keyStyle := DashboardFooterKeyStyle().Render(k.key)
		descStyle := DashboardFooterActionStyle().Render(k.desc)
		parts = append(parts, fmt.Sprintf("%s %s", keyStyle, descStyle))
	}

	return DashboardFooterStyle().Render(strings.Join(parts, "   "))
}

func (m Model) dashboardContentWidth() int {
	if m.width <= 0 {
		return 80
	}
	if m.width > 100 {
		return 100
	}
	return m.width
}

// getRegion returns the current AWS region
func (m Model) getRegion() string {
	if m.config.Region != "" {
		return m.config.Region
	}
	return "us-east-1" // default
}

// getProfile returns the current AWS profile
func (m Model) getProfile() string {
	if m.config.Profile != "" {
		return m.config.Profile
	}
	return "default" // default
}
