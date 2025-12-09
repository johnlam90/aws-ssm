package tui

// renderHelp renders the help view with comprehensive keybindings using pooled string builder
func (m Model) renderHelp() string {
	rb := NewRenderBuffer()

	// Header
	header := m.renderHeader("Help & Keybindings", "Navigate and manage AWS resources")
	rb.WriteLine(header).Newline()

	// Show quick reference first
	rb.WriteLine(GetQuickReference())

	// View-specific keybindings
	if m.currentView != ViewHelp {
		bindings := GetKeyBindings(m.currentView)
		if len(bindings) > 0 {
			rb.WriteLine(TitleStyle().Render("Current View Keybindings")).Newline()
			rb.WriteLine(FormatKeyBindings(bindings))
		}
	}

	// Global keybindings
	rb.WriteLine(TitleStyle().Render("Global Keybindings")).Newline()
	rb.WriteLine(FormatKeyBindings(globalKeyBindings))

	// Navigation tips
	rb.WriteLine(TitleStyle().Render("Navigation Tips")).Newline()
	rb.WriteLine(SubtitleStyle().Render("Vim-style navigation:"))
	rb.WriteString("  • Use j/k for up/down movement\n")
	rb.WriteString("  • Use g g to jump to top, G to jump to bottom\n")
	rb.WriteString("  • Use ctrl+d/ctrl+u for page navigation\n")
	rb.Newline()
	rb.WriteLine(SubtitleStyle().Render("Command chaining:"))
	rb.WriteString("  • Press keys in sequence for advanced commands\n")
	rb.WriteString("  • Press esc to cancel current command\n")
	rb.Newline()

	// Footer
	rb.WriteLine(HelpStyle().Render("Press ESC or ? to close help"))

	// Status bar
	rb.WriteString(StatusBarStyle().Width(m.width).Render(m.getStatusBar()))

	return rb.String()
}

// helpItem represents a help item
