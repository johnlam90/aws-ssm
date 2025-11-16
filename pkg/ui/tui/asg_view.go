package tui

import (
	"fmt"
	"strings"
)

// renderASGs renders the Auto Scaling Groups view - minimal design
func (m Model) renderASGs() string {
	asgs := m.getASGs()
	if s := m.renderASGState(asgs); s != "" {
		return s
	}
	var b strings.Builder
	header := m.renderHeader("Auto Scaling Groups", fmt.Sprintf("%d ASGs", len(asgs)))
	b.WriteString(header)
	b.WriteString("\n\n")
	b.WriteString(TableHeaderStyle().Render(fmt.Sprintf("  %-50s %8s %8s %8s %8s",
		"NAME", "DESIRED", "MIN", "MAX", "CURRENT")))
	b.WriteString("\n")
	cursor := clampIndex(m.cursor, len(asgs))
	startIdx, endIdx := m.calculateVisibleRange(len(asgs), cursor, m.height-14)
	for i := startIdx; i < endIdx; i++ {
		asg := asgs[i]
		name := asg.Name
		if len(name) > 50 {
			name = name[:47] + "..."
		}
		row := fmt.Sprintf("  %-50s %8d %8d %8d %8d",
			name, asg.DesiredCapacity, asg.MinSize, asg.MaxSize, asg.CurrentSize)
		b.WriteString(RenderSelectableRow(row, i == cursor))
		b.WriteString("\n")
	}
	if endIdx-startIdx > 0 && len(asgs) > endIdx-startIdx {
		b.WriteString("\n")
		b.WriteString(SubtitleStyle().Render(fmt.Sprintf("Showing %d-%d of %d", startIdx+1, endIdx, len(asgs))))
	}
	selected := asgs[cursor]
	b.WriteString("\n")
	b.WriteString(SubtitleStyle().Render(selected.Name))
	b.WriteString("\n")
	b.WriteString(m.renderASGDetails(selected))
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
	b.WriteString("\n")
	b.WriteString(StatusBarStyle().Width(m.width).Render(m.getStatusBar()))
	return b.String()
}

func (m Model) renderASGState(asgs []ASG) string {
	var b strings.Builder
	header := m.renderHeader("Auto Scaling Groups", fmt.Sprintf("%d ASGs", len(asgs)))
	b.WriteString(header)
	b.WriteString("\n\n")
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
	if len(asgs) == 0 {
		b.WriteString(SubtitleStyle().Render("No Auto Scaling Groups found"))
		b.WriteString("\n\n")
		b.WriteString(HelpStyle().Render("esc:back"))
		b.WriteString("\n")
		b.WriteString(StatusBarStyle().Render(m.getStatusBar()))
		return b.String()
	}
	return ""
}

func (m Model) calculateVisibleRange(total, cursor, visibleHeight int) (int, int) {
	if visibleHeight < 5 {
		return 0, total
	}
	start := cursor - visibleHeight/2
	if start < 0 {
		start = 0
	}
	end := start + visibleHeight
	if end > total {
		end = total
		start = end - visibleHeight
		if start < 0 {
			start = 0
		}
	}
	return start, end
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
