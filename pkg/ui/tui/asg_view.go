package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// renderASGs renders the Auto Scaling Groups view - minimal design
func (m Model) renderASGs() string {
	var b strings.Builder

	asgs := m.getASGs()

	// Header - simple
	header := m.renderHeader("Auto Scaling Groups", fmt.Sprintf("%d ASGs", len(asgs)))
	b.WriteString(header)
	b.WriteString("\n\n")

	// Show loading or error - minimal
	if m.loading {
		b.WriteString(m.renderLoading())
		b.WriteString("\n")
        b.WriteString(StatusBarStyle().Render(m.getStatusBar()))
		return b.String()
	}

	if m.err != nil {
		b.WriteString(m.renderError())
		b.WriteString("\n\n")
        b.WriteString(HelpStyle().Render("esc:back"))
		b.WriteString("\n")
        b.WriteString(StatusBarStyle().Render(m.getStatusBar()))
		return b.String()
	}

	// No ASGs
	if len(asgs) == 0 {
        b.WriteString(SubtitleStyle().Render("No Auto Scaling Groups found"))
		b.WriteString("\n\n")
        b.WriteString(HelpStyle().Render("esc:back"))
		b.WriteString("\n")
        b.WriteString(StatusBarStyle().Render(m.getStatusBar()))
		return b.String()
	}

	cursor := clampIndex(m.cursor, len(asgs))

	// Table header - clean and aligned
	headerRow := fmt.Sprintf("  %-50s %8s %8s %8s %8s",
		"NAME", "DESIRED", "MIN", "MAX", "CURRENT")
    b.WriteString(TableHeaderStyle().Render(headerRow))
	b.WriteString("\n")

	visibleHeight := m.height - 14
	if visibleHeight < 5 {
		visibleHeight = len(asgs)
	}

	startIdx := 0
	endIdx := len(asgs)
	if len(asgs) > visibleHeight && visibleHeight > 0 {
		startIdx = cursor - visibleHeight/2
		if startIdx < 0 {
			startIdx = 0
		}
		endIdx = startIdx + visibleHeight
		if endIdx > len(asgs) {
			endIdx = len(asgs)
			startIdx = endIdx - visibleHeight
			if startIdx < 0 {
				startIdx = 0
			}
		}
	}

	// Render ASGs with proper alignment
	for i := startIdx; i < endIdx; i++ {
		asg := asgs[i]
		// Truncate name if too long
		name := asg.Name
		if len(name) > 50 {
			name = name[:47] + "..."
		}

		row := fmt.Sprintf("  %-50s %8d %8d %8d %8d",
			name, asg.DesiredCapacity, asg.MinSize, asg.MaxSize, asg.CurrentSize)

		b.WriteString(RenderSelectableRow(row, i == cursor))
		b.WriteString("\n")
	}

	if visibleHeight > 0 && len(asgs) > visibleHeight {
		pageInfo := fmt.Sprintf("Showing %d-%d of %d", startIdx+1, endIdx, len(asgs))
		b.WriteString("\n")
        b.WriteString(SubtitleStyle().Render(pageInfo))
	}

	selected := asgs[cursor]
	b.WriteString("\n")
    b.WriteString(SubtitleStyle().Render(selected.Name))
	b.WriteString("\n")
	b.WriteString(m.renderASGDetails(selected))

	// Scaling prompt / search bar
	b.WriteString("\n")
	if overlay := m.renderScalingPrompt(ViewASGs); overlay != "" {
		b.WriteString(overlay)
		b.WriteString("\n")
	}
	if searchBar := m.renderSearchBar(ViewASGs); searchBar != "" {
		b.WriteString(searchBar)
		b.WriteString("\n")
	}
	if status := m.renderStatusMessage(); status != "" {
		b.WriteString(status)
		b.WriteString("\n")
	}
	b.WriteString(m.renderASGFooter())

	// Status bar
	b.WriteString("\n")
    b.WriteString(StatusBarStyle().Width(m.width).Render(m.getStatusBar()))

	return b.String()
}

// handleASGKeys handles keyboard input for ASG view
func (m Model) handleASGKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	asgs := m.getASGs()

	switch msg.String() {
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}

	case "down", "j":
		if m.cursor < len(asgs)-1 {
			m.cursor++
		}

	case "g":
		if len(asgs) > 0 {
			m.cursor = 0
		}

	case "G":
		if len(asgs) > 0 {
			m.cursor = len(asgs) - 1
		}

	case "enter", " ":
		if m.cursor < len(asgs) {
			asg := asgs[m.cursor]
			m = m.startASGScaling(asg)
		}

	case "r":
		// Refresh ASGs
		m.loading = true
		m.loadingMsg = "Refreshing Auto Scaling Groups..."
		m.err = nil
		m.statusMessage = ""
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
		{"/", "search"},
		{"esc", "back"},
	}

	var parts []string
	for _, k := range keys {
        keyStyle := StatusBarKeyStyle().Render(k.key)
        descStyle := StatusBarValueStyle().Render(k.desc)
		parts = append(parts, fmt.Sprintf("%s %s", keyStyle, descStyle))
	}

    return HelpStyle().Render(strings.Join(parts, " • "))
}

func (m Model) renderASGDetails(asg ASG) string {
	var b strings.Builder

	b.WriteString("  Scaling:\n")
	b.WriteString(fmt.Sprintf("    Desired: %d  Min: %d  Max: %d  Current: %d\n",
		asg.DesiredCapacity, asg.MinSize, asg.MaxSize, asg.CurrentSize))
	if asg.Status != "" {
		b.WriteString(fmt.Sprintf("    Status:  %s\n", StateStyle(strings.ToLower(asg.Status))))
	}
	if asg.HealthCheckType != "" {
		b.WriteString(fmt.Sprintf("    Health Check: %s\n", asg.HealthCheckType))
	}
	if !asg.CreatedAt.IsZero() {
		b.WriteString(fmt.Sprintf("    Created: %s\n", formatRelativeTimestamp(asg.CreatedAt)))
	}

	b.WriteString("\n  Launch Configuration:\n")
	switch {
	case strings.TrimSpace(asg.LaunchTemplateName) != "":
		b.WriteString(fmt.Sprintf("    Template: %s", asg.LaunchTemplateName))
		if strings.TrimSpace(asg.LaunchTemplateVersion) != "" {
			b.WriteString(fmt.Sprintf(" (version %s)", asg.LaunchTemplateVersion))
		}
		b.WriteString("\n")
	case strings.TrimSpace(asg.LaunchConfigurationName) != "":
		b.WriteString(fmt.Sprintf("    Configuration: %s\n", asg.LaunchConfigurationName))
	default:
		b.WriteString("    Configuration: n/a\n")
	}

	if len(asg.AvailabilityZones) > 0 {
		b.WriteString("\n  Availability Zones:\n")
		for _, az := range asg.AvailabilityZones {
			b.WriteString(fmt.Sprintf("    • %s\n", az))
		}
	}

	if len(asg.LoadBalancerNames) > 0 {
		b.WriteString("\n  Load Balancers:\n")
		for _, lb := range asg.LoadBalancerNames {
			b.WriteString(fmt.Sprintf("    • %s\n", lb))
		}
	}

	if len(asg.TargetGroupARNs) > 0 {
		b.WriteString("\n  Target Groups:\n")
		for _, tg := range asg.TargetGroupARNs {
			b.WriteString(fmt.Sprintf("    • %s\n", tg))
		}
	}

	if lines := renderTagLines(asg.Tags); len(lines) > 0 {
		b.WriteString("\n  Tags:\n")
		for _, line := range lines {
			b.WriteString(line)
			b.WriteString("\n")
		}
	}

	return b.String()
}
