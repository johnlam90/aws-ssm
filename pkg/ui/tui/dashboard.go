package tui

import (
	"fmt"
	"strings"
)

// renderDashboard renders the main dashboard view with beautiful aesthetics
func (m Model) renderDashboard() string {
	var b strings.Builder

	// ASCII art banner for production-grade feel
	banner := m.renderDashboardBanner()
	b.WriteString(banner)
	b.WriteString("\n")

	// Top header bar with context information
	headerBar := m.renderDashboardHeaderBar()
	b.WriteString(headerBar)
	b.WriteString("\n")

	// Horizontal separator
	separator := m.renderDashboardSeparator()
	b.WriteString(separator)
	b.WriteString("\n")

	// Show loading state with beautiful styling
	if m.loading {
		b.WriteString(m.renderLoading())
		b.WriteString("\n")
		b.WriteString(StatusBarStyle().Width(m.width).Render(m.getStatusBar()))
		return b.String()
	}

	// Show error state with beautiful styling
	if m.err != nil {
		b.WriteString(m.renderError())
		b.WriteString("\n\n")
		b.WriteString(HelpStyle().Render("esc:back"))
		b.WriteString("\n")
		b.WriteString(StatusBarStyle().Width(m.width).Render(m.getStatusBar()))
		return b.String()
	}

	// Welcome message and system info
	welcomeMsg := m.renderWelcomeMessage()
	b.WriteString(welcomeMsg)
	b.WriteString("\n\n")

	// Section title with visual enhancement
	sectionTitle := DashboardSectionTitleStyle().Render("┌─ Services ─────────────────────────────────────────────────┐")
	b.WriteString(sectionTitle)
	b.WriteString("\n")

	// Menu items with beautiful two-column layout
	for i, item := range m.menuItems {
		menuItem := m.renderDashboardMenuItem(i, item, i == m.cursor)
		b.WriteString(menuItem)
		b.WriteString("\n")
	}

	// Close the services box
	boxFooter := DashboardSectionTitleStyle().Render("└────────────────────────────────────────────────────────────┘")
	b.WriteString(boxFooter)
	b.WriteString("\n")

	// Horizontal separator before footer
	b.WriteString("\n")
	b.WriteString(separator)
	b.WriteString("\n")

	// Beautiful footer with concise keyboard hints
	footer := m.renderDashboardFooter()
	b.WriteString(footer)

	// Status bar with consistent styling
	b.WriteString("\n")
	b.WriteString(StatusBarStyle().Width(m.width).Render(m.getStatusBar()))

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
	separator := strings.Repeat("─", m.width)
	return DashboardSeparatorStyle().Render(separator)
}

// renderDashboardBanner renders a minimal, elegant ASCII banner
func (m Model) renderDashboardBanner() string {
	banner := `
   ╔═══════════════════════════════════════════════════════════╗
   ║                                                           ║
   ║           AWS SSM Manager                                 ║
   ║           Terminal User Interface                         ║
   ║                                                           ║
   ╚═══════════════════════════════════════════════════════════╝`

	return DashboardBannerStyle().Render(banner)
}

// renderWelcomeMessage renders a welcoming message with system info
func (m Model) renderWelcomeMessage() string {
	region := m.getRegion()
	profile := m.getProfile()

	var b strings.Builder

	// Welcome message
	welcome := DashboardWelcomeStyle().Render("Welcome to AWS SSM Manager")
	b.WriteString(welcome)
	b.WriteString("\n")

	// System info in minimal format
	info := fmt.Sprintf("Connected to %s using profile '%s'", region, profile)
	infoStyled := DashboardInfoStyle().Render(info)
	b.WriteString(infoStyled)

	return b.String()
}

// renderDashboardMenuItem renders a beautiful two-column menu item
func (m Model) renderDashboardMenuItem(_ int, item MenuItem, isSelected bool) string {
	// Normalize descriptions for consistency
	normalizedDesc := m.normalizeServiceDescription(item.Description)

	if isSelected {
		selectionBar := DashboardSelectionBarStyle().Render("▌")
		name := DashboardSelectedNameStyle().Render(item.Title)
		desc := DashboardSelectedDescStyle().Render(normalizedDesc)
		return fmt.Sprintf("│ %s %s %s │", selectionBar, name, desc)
	}
	name := DashboardServiceNameStyle().Render(item.Title)
	desc := DashboardServiceDescStyle().Render(normalizedDesc)
	return fmt.Sprintf("│   %s %s │", name, desc)
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
		{"↑/k", "Navigate Up"},
		{"↓/j", "Navigate Down"},
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

	var b strings.Builder
	b.WriteString(DashboardFooterStyle().Render("Keyboard Shortcuts"))
	b.WriteString("\n")
	b.WriteString(DashboardFooterDetailStyle().Render(strings.Join(parts, "  •  ")))

	return b.String()
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
