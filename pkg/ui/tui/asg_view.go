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

	// Get responsive column widths based on terminal width
	nameW, desiredW, minW, maxW, currentW := ASGColumnWidths(m.width)

	b.WriteString(TableHeaderStyle().Render(fmt.Sprintf("  %-*s %*s %*s %*s %*s",
		nameW, "NAME", desiredW, "DESIRED", minW, "MIN", maxW, "MAX", currentW, "CURRENT")))
	b.WriteString("\n")
	cursor := clampIndex(m.cursor, len(asgs))

	// Calculate responsive vertical layout
	layout := ASGLayout(m.height)
	startIdx, endIdx := m.calculateVisibleRange(len(asgs), cursor, layout.TableHeight)

	for i := startIdx; i < endIdx; i++ {
		asg := asgs[i]
		name := asg.Name
		if len(name) > nameW {
			name = TruncateWithEllipsis(name, nameW)
		}
		row := fmt.Sprintf("  %-*s %*d %*d %*d %*d",
			nameW, name, desiredW, asg.DesiredCapacity, minW, asg.MinSize, maxW, asg.MaxSize, currentW, asg.CurrentSize)
		b.WriteString(RenderSelectableRow(row, i == cursor))
		b.WriteString("\n")
	}
	if len(asgs) > endIdx-startIdx {
		b.WriteString("\n")
		b.WriteString(SubtitleStyle().Render(fmt.Sprintf("Showing %d-%d of %d", startIdx+1, endIdx, len(asgs))))
	}
	selected := asgs[cursor]
	b.WriteString("\n")
	b.WriteString(SubtitleStyle().Render(selected.Name))
	b.WriteString("\n")
	b.WriteString(m.renderASGDetailsResponsive(selected, layout.DetailHeight))
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
		b.WriteString(StatusBarStyle().Width(m.width).Render(m.getStatusBar()))
		return b.String()
	}
	if m.err != nil {
		b.WriteString(m.renderError())
		b.WriteString("\n\n")
		b.WriteString(HelpStyle().Render("esc:back"))
		b.WriteString("\n")
		b.WriteString(StatusBarStyle().Width(m.width).Render(m.getStatusBar()))
		return b.String()
	}
	if len(asgs) == 0 {
		b.WriteString(SubtitleStyle().Render("No Auto Scaling Groups found"))
		b.WriteString("\n\n")
		b.WriteString(HelpStyle().Render("esc:back"))
		b.WriteString("\n")
		b.WriteString(StatusBarStyle().Width(m.width).Render(m.getStatusBar()))
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

// renderASGDetailsResponsive renders ASG details with responsive height
func (m Model) renderASGDetailsResponsive(asg ASG, maxLines int) string {
	var lines []string

	// Scaling info (always show)
	lines = append(lines, "  Scaling:")
	lines = append(lines, fmt.Sprintf("    Desired: %d  Min: %d  Max: %d  Current: %d",
		asg.DesiredCapacity, asg.MinSize, asg.MaxSize, asg.CurrentSize))

	if maxLines > 3 && asg.Status != "" {
		lines = append(lines, fmt.Sprintf("    Status:  %s", StateStyle(strings.ToLower(asg.Status))))
	}
	if maxLines > 4 && asg.HealthCheckType != "" {
		lines = append(lines, fmt.Sprintf("    Health Check: %s", asg.HealthCheckType))
	}
	if maxLines > 5 && !asg.CreatedAt.IsZero() {
		lines = append(lines, fmt.Sprintf("    Created: %s", formatRelativeTimestamp(asg.CreatedAt)))
	}

	// Launch Configuration (show if space available)
	if maxLines > 7 {
		lines = append(lines, "")
		lines = append(lines, "  Launch Configuration:")
		switch {
		case strings.TrimSpace(asg.LaunchTemplateName) != "":
			ltLine := fmt.Sprintf("    Template: %s", asg.LaunchTemplateName)
			if strings.TrimSpace(asg.LaunchTemplateVersion) != "" {
				ltLine += fmt.Sprintf(" (version %s)", asg.LaunchTemplateVersion)
			}
			lines = append(lines, ltLine)
		case strings.TrimSpace(asg.LaunchConfigurationName) != "":
			lines = append(lines, fmt.Sprintf("    Configuration: %s", asg.LaunchConfigurationName))
		default:
			lines = append(lines, "    Configuration: n/a")
		}
	}

	// Availability Zones (show if space available)
	if maxLines > 11 && len(asg.AvailabilityZones) > 0 {
		lines = append(lines, "")
		lines = append(lines, "  Availability Zones:")
		remainingLines := maxLines - len(lines) - 4 // Reserve space for other sections
		for i, az := range asg.AvailabilityZones {
			if i >= remainingLines {
				lines = append(lines, "    ...")
				break
			}
			lines = append(lines, fmt.Sprintf("    • %s", az))
		}
	}

	// Load Balancers (show if space available)
	if maxLines > 15 && len(asg.LoadBalancerNames) > 0 {
		lines = append(lines, "")
		lines = append(lines, "  Load Balancers:")
		remainingLines := maxLines - len(lines) - 3
		for i, lb := range asg.LoadBalancerNames {
			if i >= remainingLines {
				lines = append(lines, "    ...")
				break
			}
			lines = append(lines, fmt.Sprintf("    • %s", lb))
		}
	}

	// Target Groups (show if space available)
	if maxLines > 18 && len(asg.TargetGroupARNs) > 0 {
		lines = append(lines, "")
		lines = append(lines, "  Target Groups:")
		remainingLines := maxLines - len(lines) - 2
		for i, tg := range asg.TargetGroupARNs {
			if i >= remainingLines {
				lines = append(lines, "    ...")
				break
			}
			lines = append(lines, fmt.Sprintf("    • %s", tg))
		}
	}

	// Tags (show if space available)
	if maxLines > 20 {
		tagLines := renderTagLines(asg.Tags)
		if len(tagLines) > 0 {
			lines = append(lines, "")
			lines = append(lines, "  Tags:")
			remainingLines := maxLines - len(lines)
			for i, line := range tagLines {
				if i >= remainingLines-1 {
					lines = append(lines, "    ...")
					break
				}
				lines = append(lines, line)
			}
		}
	}

	return strings.Join(lines, "\n")
}
