package tui

import (
	"strings"
)

// renderHelp renders the help view with comprehensive keybindings
func (m Model) renderHelp() string {
	var b strings.Builder

	// Header
	header := m.renderHeader("Help & Keybindings", "Navigate and manage AWS resources")
	b.WriteString(header)
	b.WriteString("\n\n")

	// Show quick reference first
	b.WriteString(GetQuickReference())
	b.WriteString("\n")

	// View-specific keybindings
	if m.currentView != ViewHelp {
		bindings := GetKeyBindings(m.currentView)
		if len(bindings) > 0 {
			b.WriteString(TitleStyle().Render("Current View Keybindings"))
			b.WriteString("\n\n")
			b.WriteString(FormatKeyBindings(bindings))
			b.WriteString("\n")
		}
	}

	// Global keybindings
	b.WriteString(TitleStyle().Render("Global Keybindings"))
	b.WriteString("\n\n")
	b.WriteString(FormatKeyBindings(globalKeyBindings))
	b.WriteString("\n")

	// Navigation tips
	b.WriteString(TitleStyle().Render("Navigation Tips"))
	b.WriteString("\n\n")
	b.WriteString(SubtitleStyle().Render("Vim-style navigation:"))
	b.WriteString("\n")
	b.WriteString("  • Use j/k for up/down movement\n")
	b.WriteString("  • Use g g to jump to top, G to jump to bottom\n")
	b.WriteString("  • Use ctrl+d/ctrl+u for page navigation\n")
	b.WriteString("\n")
	b.WriteString(SubtitleStyle().Render("Command chaining:"))
	b.WriteString("\n")
	b.WriteString("  • Press keys in sequence for advanced commands\n")
	b.WriteString("  • Press esc to cancel current command\n")
	b.WriteString("\n")

	// Footer
	b.WriteString(HelpStyle().Render("Press ESC or ? to close help"))
	b.WriteString("\n")

	// Status bar
	b.WriteString(StatusBarStyle().Width(m.width).Render(m.getStatusBar()))

	return b.String()
}

// helpItem represents a help item
