package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
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
type helpItem struct {
	key  string
	desc string
}

// renderHelpSection renders a section of help items
func (m Model) renderHelpSection(items []helpItem) string {
	var b strings.Builder

	for _, item := range items {
        key := HelpKeyStyle().Render(fmt.Sprintf("%-15s", item.key))
        desc := HelpDescStyle().Render(item.desc)
		b.WriteString(fmt.Sprintf("  %s  %s\n", key, desc))
	}

	return b.String()
}

// handleHelpKeys handles keyboard input for help view
func (m Model) handleHelpKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "?":
		return m.navigateBack(), nil

	case "up", "k":
		// Scroll help content (future)

	case "down", "j":
		// Scroll help content (future)
	}

	return m, nil
}
