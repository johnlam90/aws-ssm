package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// renderHelp renders the help view
func (m Model) renderHelp() string {
	var b strings.Builder

	// Header
	header := m.renderHeader("Help & Keybindings", "Navigate and manage AWS resources")
	b.WriteString(header)
	b.WriteString("\n\n")

	// Global keybindings - minimal
	b.WriteString(TitleStyle.Render("Global Keybindings"))
	b.WriteString("\n\n")
	b.WriteString(m.renderHelpSection([]helpItem{
		{"q", "Quit application (from dashboard)"},
		{"ctrl+c", "Force quit"},
		{"?", "Toggle help"},
		{"esc", "Go back / Cancel"},
	}))
	b.WriteString("\n")

	// Navigation keybindings - minimal
	b.WriteString(TitleStyle.Render("Navigation"))
	b.WriteString("\n\n")
	b.WriteString(m.renderHelpSection([]helpItem{
		{"↑ / k", "Move up"},
		{"↓ / j", "Move down"},
		{"← / h", "Move left (future use)"},
		{"→ / l", "Move right (future use)"},
		{"g", "Go to top"},
		{"G", "Go to bottom"},
		{"enter / space", "Select item"},
	}))
	b.WriteString("\n")

	// Resource-specific keybindings - minimal
	b.WriteString(TitleStyle.Render("Resource Actions"))
	b.WriteString("\n\n")
	b.WriteString(m.renderHelpSection([]helpItem{
		{"r", "Refresh current view"},
		{"enter", "Connect to instance / View details"},
		{"/", "Search/filter current view"},
		{"ctrl+u", "Clear search input"},
	}))
	b.WriteString("\n")

	// Clipboard - minimal
	b.WriteString(TitleStyle.Render("Clipboard"))
	b.WriteString("\n\n")
	b.WriteString(m.renderHelpSection([]helpItem{
		{"mouse select", "Select any visible text with mouse"},
		{"cmd+c", "Copy selected text (macOS)"},
		{"ctrl+shift+c", "Copy selected text (Linux/Windows)"},
	}))
	b.WriteString("\n")

	// About - minimal
	b.WriteString(TitleStyle.Render("About"))
	b.WriteString("\n\n")
	b.WriteString(SubtitleStyle.Render("AWS SSM Manager TUI"))
	b.WriteString("\n")
	b.WriteString(SubtitleStyle.Render("A terminal user interface for managing AWS resources"))
	b.WriteString("\n\n")

	// Footer
	b.WriteString(HelpStyle.Render("Press ESC or ? to close help"))
	b.WriteString("\n")

	// Status bar
	b.WriteString(StatusBarStyle.Width(m.width).Render(m.getStatusBar()))

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
		key := HelpKeyStyle.Render(fmt.Sprintf("%-15s", item.key))
		desc := HelpDescStyle.Render(item.desc)
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
