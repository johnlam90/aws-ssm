package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// renderDashboard renders the main dashboard view with beautiful aesthetics
func (m Model) renderDashboard() string {
	var b strings.Builder

	// Top header bar with context information
	headerBar := m.renderDashboardHeaderBar()
	b.WriteString(headerBar)
	b.WriteString("\n")

	// Main title
	title := DashboardTitleStyle().Render("AWS SSM Manager")
	b.WriteString(title)
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

	// Section title
	sectionTitle := DashboardSectionTitleStyle().Render("Services")
	b.WriteString(sectionTitle)
	b.WriteString("\n")

	// Menu items with beautiful two-column layout
	for i, item := range m.menuItems {
		menuItem := m.renderDashboardMenuItem(i, item, i == m.cursor)
		b.WriteString(menuItem)
		b.WriteString("\n")
	}

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

// renderFooter renders an enhanced footer with sophisticated keybinding styling
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
		// Use consistent styling with other views (StatusBarKeyStyle for keys)
		keyStyle := StatusBarKeyStyle().Render(k.key)
		descStyle := StatusBarValueStyle().Render(k.desc)
		parts = append(parts, fmt.Sprintf("%s %s", keyStyle, descStyle))
	}

	// Use the same separator style as other views (bullet separator)
	return HelpStyle().Render(strings.Join(parts, " • "))
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

// renderMenuItemWithEffect renders a menu item with subtle visual effects
func (m Model) renderMenuItemWithEffect(index int, item MenuItem, isSelected bool) string {
	var titleStyle lipgloss.Style
	var descStyle lipgloss.Style
	var prefix string
	
	if isSelected {
		// Enhanced selected state with subtle animation-like effects
		titleStyle = MenuItemSelectedStyle().
			Background(lipgloss.Color("#1E3A8A")). // Deep blue background
			Foreground(lipgloss.Color("#60A5FA")). // Bright blue text
			Bold(true)
		descStyle = MenuItemSelectedStyle().
			Background(lipgloss.Color("#1E3A8A")). // Consistent background
			Foreground(lipgloss.Color("#93C5FD"))  // Lighter blue for description
		prefix = "▸ " // Visual indicator
	} else {
		// Unselected state with subtle hover potential
		titleStyle = MenuItemStyle().
			Foreground(lipgloss.Color("#E6EDF3")) // Primary text color
		descStyle = SubtitleStyle().
			Foreground(lipgloss.Color("#8B949E")) // Secondary text color
		prefix = "  " // Standard indentation
	}
	
	// Apply consistent spacing and alignment
	title := titleStyle.Render(item.Title)
	desc := descStyle.Render(" – " + item.Description)
	
	return prefix + title + desc
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

// renderDashboardMenuItem renders a beautiful two-column menu item
func (m Model) renderDashboardMenuItem(index int, item MenuItem, isSelected bool) string {
	// Normalize descriptions for consistency
	normalizedDesc := m.normalizeServiceDescription(item.Description)
	
	if isSelected {
		// Selected state with vertical bar and highlight
		selectionBar := DashboardSelectionBarStyle().Render("▌")
		name := DashboardSelectedNameStyle().Render(item.Title)
		desc := DashboardSelectedDescStyle().Render(normalizedDesc)
		return fmt.Sprintf("%s %s %s", selectionBar, name, desc)
	} else {
		// Normal state with proper spacing
		name := DashboardServiceNameStyle().Render(item.Title)
		desc := DashboardServiceDescStyle().Render(normalizedDesc)
		return fmt.Sprintf("  %s %s", name, desc)
	}
}

// normalizeServiceDescription normalizes service descriptions for consistency
func (m Model) normalizeServiceDescription(desc string) string {
	// Map of original descriptions to normalized ones
	normalizedMap := map[string]string{
		"View and manage EC2 instances":                    "Manage EC2 instances",
		"Manage EKS clusters and node groups":              "Manage EKS clusters & node groups",
		"View and scale ASGs":                              "Scale and monitor ASGs",
		"Inspect managed node groups":                      "Inspect managed node groups",
		"View EC2 network interfaces and ENIs":           "View EC2 ENIs",
		"View keybindings and help":                        "Keybindings and documentation",
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
