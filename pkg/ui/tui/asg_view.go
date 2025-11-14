package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// renderASGs renders the Auto Scaling Groups view - minimal design
func (m Model) renderASGs() string {
	var b strings.Builder

	// Header - simple
	header := m.renderHeader("Auto Scaling Groups", fmt.Sprintf("%d ASGs", len(m.asgs)))
	b.WriteString(header)
	b.WriteString("\n\n")

	// Show loading or error - minimal
	if m.loading {
		b.WriteString(m.renderLoading())
		b.WriteString("\n")
		b.WriteString(StatusBarStyle.Render(m.getStatusBar()))
		return b.String()
	}

	if m.err != nil {
		b.WriteString(m.renderError())
		b.WriteString("\n\n")
		b.WriteString(HelpStyle.Render("esc:back"))
		b.WriteString("\n")
		b.WriteString(StatusBarStyle.Render(m.getStatusBar()))
		return b.String()
	}

	// No ASGs
	if len(m.asgs) == 0 {
		b.WriteString(SubtitleStyle.Render("No Auto Scaling Groups found"))
		b.WriteString("\n\n")
		b.WriteString(HelpStyle.Render("esc:back"))
		b.WriteString("\n")
		b.WriteString(StatusBarStyle.Render(m.getStatusBar()))
		return b.String()
	}

	// Table header - clean and aligned
	headerRow := fmt.Sprintf("  %-50s %8s %8s %8s %8s",
		"NAME", "DESIRED", "MIN", "MAX", "CURRENT")
	b.WriteString(TableHeaderStyle.Render(headerRow))
	b.WriteString("\n")

	// Render ASGs with proper alignment
	for i, asg := range m.asgs {
		// Truncate name if too long
		name := asg.Name
		if len(name) > 50 {
			name = name[:47] + "..."
		}

		row := fmt.Sprintf("  %-50s %8d %8d %8d %8d",
			name, asg.DesiredCapacity, asg.MinSize, asg.MaxSize, asg.CurrentSize)

		if i == m.cursor {
			b.WriteString(ListItemSelectedStyle.Render("▶" + row[1:]))
		} else {
			b.WriteString(ListItemStyle.Render(row))
		}
		b.WriteString("\n")
	}

	// Footer
	b.WriteString("\n")
	b.WriteString(m.renderASGFooter())

	// Status bar
	b.WriteString("\n")
	b.WriteString(StatusBarStyle.Width(m.width).Render(m.getStatusBar()))

	return b.String()
}

// handleASGKeys handles keyboard input for ASG view
func (m Model) handleASGKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}

	case "down", "j":
		if m.cursor < len(m.asgs)-1 {
			m.cursor++
		}

	case "g":
		m.cursor = 0

	case "G":
		m.cursor = len(m.asgs) - 1

	case "enter", " ":
		// Scale ASG - exit TUI and prompt for new capacity
		if m.cursor < len(m.asgs) {
			asg := m.asgs[m.cursor]
			return m, m.scheduleASGScale(asg.Name)
		}

	case "r":
		// Refresh ASGs
		m.loading = true
		m.loadingMsg = "Refreshing Auto Scaling Groups..."
		m.err = nil
		return m, LoadASGsCmd(m.ctx, m.client)

	case "esc":
		return m.navigateBack(), nil
	}

	return m, nil
}

// renderASGFooter renders the footer for ASG view
func (m Model) renderASGFooter() string {
	keys := []struct {
		key  string
		desc string
	}{
		{"↑/k", "up"},
		{"↓/j", "down"},
		{"g/G", "top/bottom"},
		{"enter", "scale"},
		{"r", "refresh"},
		{"esc", "back"},
	}

	var parts []string
	for _, k := range keys {
		keyStyle := StatusBarKeyStyle.Render(k.key)
		descStyle := StatusBarValueStyle.Render(k.desc)
		parts = append(parts, fmt.Sprintf("%s %s", keyStyle, descStyle))
	}

	return HelpStyle.Render(strings.Join(parts, " • "))
}
