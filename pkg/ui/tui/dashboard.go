package tui

import (
	"fmt"
	"strings"
)

// renderDashboard renders the main dashboard view with beautiful aesthetics using pooled string builder
func (m Model) renderDashboard() string {
	rb := NewRenderBuffer()

	// Top header bar with context information
	headerBar := m.renderDashboardHeaderBar()
	rb.WriteLine(headerBar)

	// Main title
	title := DashboardTitleStyle().Render("AWS SSM Manager")
	rb.WriteLine(title)

	// Horizontal separator
	separator := m.renderDashboardSeparator()
	rb.WriteLine(separator)

	// Show loading state with beautiful styling
	if m.loading {
		rb.WriteLine(m.renderLoading())
		rb.WriteString(StatusBarStyle().Width(m.width).Render(m.getStatusBar()))
		return rb.String()
	}

	// Show error state with beautiful styling
	if m.err != nil {
		rb.WriteLine(m.renderError()).Newline()
		rb.WriteLine(HelpStyle().Render("esc:back"))
		rb.WriteString(StatusBarStyle().Width(m.width).Render(m.getStatusBar()))
		return rb.String()
	}

	// Section title
	sectionTitle := DashboardSectionTitleStyle().Render("Services")
	rb.WriteLine(sectionTitle)

	// Menu items with beautiful two-column layout
	for i, item := range m.menuItems {
		menuItem := m.renderDashboardMenuItem(i, item, i == m.cursor)
		rb.WriteLine(menuItem)
	}

	// Horizontal separator before footer
	rb.Newline().WriteLine(separator)

	// Beautiful footer with concise keyboard hints
	footer := m.renderDashboardFooter()
	rb.WriteString(footer)

	// Status bar with consistent styling
	rb.Newline()
	rb.WriteString(StatusBarStyle().Width(m.width).Render(m.getStatusBar()))

	return rb.String()
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

// renderLoading renders an enhanced loading message with animated spinner
func (m Model) renderLoading() string {
	if !m.loading {
		return ""
	}

	// Use the animated spinner from the model
	msg := fmt.Sprintf("%s %s", m.spinner.View(), m.loadingMsg)
	return LoadingStyle().Render(msg)
}

// renderError renders an enhanced error message with consistent styling
func (m Model) renderError() string {
	if m.err == nil {
		return ""
	}

	return ErrorStyle().Render(fmt.Sprintf("Error: %v", m.err))
}

// renderDashboardHeaderBar renders the top header bar with context information
func (m Model) renderDashboardHeaderBar() string {
	region := m.getRegion()
	profile := m.getProfile()

	contextParts := []string{}
	if region != "" {
		contextParts = append(contextParts, fmt.Sprintf("Region: %s", region))
	}
	if profile != "" {
		contextParts = append(contextParts, fmt.Sprintf("Profile: %s", profile))
	}
	contextParts = append(contextParts, fmt.Sprintf("View: %s", "Dashboard"))

	return DashboardHeaderBarStyle().Render(strings.Join(contextParts, "   "))
}

// renderDashboardSeparator renders a horizontal separator line
func (m Model) renderDashboardSeparator() string {
	// Use available width, but clamp to reasonable max for readability
	separatorWidth := m.width
	if separatorWidth <= 0 {
		separatorWidth = 80
	}
	separator := strings.Repeat("─", separatorWidth)
	return DashboardSeparatorStyle().Render(separator)
}

// renderDashboardMenuItem renders a beautiful two-column menu item
func (m Model) renderDashboardMenuItem(_ int, item MenuItem, isSelected bool) string {
	// Normalize descriptions for consistency
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

// renderDashboardFooter renders the beautiful footer with concise keyboard hints
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

	return DashboardFooterStyle().Render(fmt.Sprintf("Navigation: %s", strings.Join(parts, "  ")))
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
