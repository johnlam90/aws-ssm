package tui

import (
	"strings"
)

// renderHelp renders the help main-panel content.
func (m Model) renderHelp() string {
	var b strings.Builder

	b.WriteString(TitleStyle().Render("Help & Keybindings"))
	b.WriteString("\n")
	b.WriteString(SubtitleStyle().Render("Navigate and manage AWS resources"))
	b.WriteString("\n\n")

	b.WriteString(GetQuickReference())
	b.WriteString("\n")

	if m.currentView != ViewHelp {
		bindings := GetKeyBindings(m.currentView)
		if len(bindings) > 0 {
			b.WriteString(TitleStyle().Render("Current View Keybindings"))
			b.WriteString("\n\n")
			b.WriteString(FormatKeyBindings(bindings))
			b.WriteString("\n")
		}
	}

	b.WriteString(TitleStyle().Render("Global Keybindings"))
	b.WriteString("\n\n")
	b.WriteString(FormatKeyBindings(globalKeyBindings))
	b.WriteString("\n")

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

	return b.String()
}
