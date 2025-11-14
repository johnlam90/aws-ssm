package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// renderEC2Instances renders the EC2 instances view
func (m Model) renderEC2Instances() string {
	var b strings.Builder

	// Header
	header := m.renderHeader("EC2 Instances", fmt.Sprintf("%d instances", len(m.ec2Instances)))
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

	// No instances
	if len(m.ec2Instances) == 0 {
		b.WriteString(SubtitleStyle.Render("No EC2 instances found"))
		b.WriteString("\n\n")
		b.WriteString(HelpStyle.Render("esc:back"))
		b.WriteString("\n")
		b.WriteString(StatusBarStyle.Render(m.getStatusBar()))
		return b.String()
	}

	// Table header
	headerRow := fmt.Sprintf("%-30s %-20s %-15s %-15s %-12s",
		"NAME", "INSTANCE ID", "PRIVATE IP", "STATE", "TYPE")
	b.WriteString(TableHeaderStyle.Render(headerRow))
	b.WriteString("\n")

	// Calculate visible range for pagination
	visibleHeight := m.height - 10 // Reserve space for header, footer, status
	startIdx := 0
	endIdx := len(m.ec2Instances)

	if len(m.ec2Instances) > visibleHeight {
		// Center the cursor in the visible area
		startIdx = m.cursor - visibleHeight/2
		if startIdx < 0 {
			startIdx = 0
		}
		endIdx = startIdx + visibleHeight
		if endIdx > len(m.ec2Instances) {
			endIdx = len(m.ec2Instances)
			startIdx = endIdx - visibleHeight
			if startIdx < 0 {
				startIdx = 0
			}
		}
	}

	// Render instances
	for i := startIdx; i < endIdx; i++ {
		inst := m.ec2Instances[i]
		name := inst.Name
		if name == "" {
			name = "(no name)"
		}
		if len(name) > 28 {
			name = name[:25] + "..."
		}

		state := StateStyle(inst.State)
		row := fmt.Sprintf("%-30s %-20s %-15s %-15s %-12s",
			name, inst.InstanceID, inst.PrivateIP, state, inst.InstanceType)

		if i == m.cursor {
			b.WriteString(ListItemSelectedStyle.Render("▶ " + row))
		} else {
			b.WriteString(ListItemStyle.Render("  " + row))
		}
		b.WriteString("\n")
	}

	// Pagination indicator
	if len(m.ec2Instances) > visibleHeight {
		pageInfo := fmt.Sprintf("Showing %d-%d of %d", startIdx+1, endIdx, len(m.ec2Instances))
		b.WriteString("\n")
		b.WriteString(SubtitleStyle.Render(pageInfo))
	}

	// Footer
	b.WriteString("\n")
	b.WriteString(m.renderEC2Footer())

	// Status bar
	b.WriteString("\n")
	b.WriteString(StatusBarStyle.Width(m.width).Render(m.getStatusBar()))

	return b.String()
}

// handleEC2Keys handles keyboard input for EC2 instances view
func (m Model) handleEC2Keys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}

	case "down", "j":
		if m.cursor < len(m.ec2Instances)-1 {
			m.cursor++
		}

	case "g":
		// Go to top
		m.cursor = 0

	case "G":
		// Go to bottom
		m.cursor = len(m.ec2Instances) - 1

	case "enter", " ":
		// Connect to selected instance via SSM
		if m.cursor < len(m.ec2Instances) {
			inst := m.ec2Instances[m.cursor]

			// Check if instance is running
			if inst.State != "running" {
				m.err = fmt.Errorf("instance %s is not running (state: %s)", inst.Name, inst.State)
				return m, nil
			}

			// Schedule SSM session to start after TUI exits
			return m, m.scheduleSSMSession(inst.InstanceID)
		}

	case "r":
		// Refresh instances
		m.loading = true
		m.loadingMsg = "Refreshing EC2 instances..."
		m.err = nil
		return m, LoadEC2InstancesCmd(m.ctx, m.client)

	case "esc":
		return m.navigateBack(), nil
	}

	return m, nil
}

// renderEC2Footer renders the footer for EC2 view
func (m Model) renderEC2Footer() string {
	keys := []struct {
		key  string
		desc string
	}{
		{"↑/k", "up"},
		{"↓/j", "down"},
		{"g/G", "top/bottom"},
		{"enter", "connect"},
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
