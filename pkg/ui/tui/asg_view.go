package tui

import (
	"fmt"
	"strings"
)

// renderASGs renders the Auto Scaling Groups view using pooled string builder
func (m Model) renderASGs() string {
	asgs := m.getASGs()
	if s := m.renderASGState(asgs); s != "" {
		return s
	}
	rb := NewRenderBuffer()
	header := m.renderHeader("Auto Scaling Groups", fmt.Sprintf("%d ASGs", len(asgs)))
	rb.WriteLine(header).Newline()

	// Get responsive column widths based on terminal width
	nameW, desiredW, minW, maxW, currentW := ASGColumnWidths(m.width)

	rb.WriteLine(TableHeaderStyle().Render(fmt.Sprintf("  %-*s %*s %*s %*s %*s",
		nameW, "NAME", desiredW, "DESIRED", minW, "MIN", maxW, "MAX", currentW, "CURRENT")))
	cursor := clampIndex(m.cursor, len(asgs))

	// Calculate responsive vertical layout
	layout := ASGLayout(m.height)
	startIdx, endIdx := CalculateVisibleRangeWithThreshold(len(asgs), cursor, layout.TableHeight, 5)

	for i := startIdx; i < endIdx; i++ {
		asg := asgs[i]
		name := asg.Name
		if len(name) > nameW {
			name = TruncateWithEllipsis(name, nameW)
		}
		row := fmt.Sprintf("  %-*s %*d %*d %*d %*d",
			nameW, name, desiredW, asg.DesiredCapacity, minW, asg.MinSize, maxW, asg.MaxSize, currentW, asg.CurrentSize)
		rb.WriteLine(RenderSelectableRow(row, i == cursor))
	}
	if len(asgs) > endIdx-startIdx {
		rb.Newline()
		rb.WriteString(SubtitleStyle().Render(fmt.Sprintf("Showing %d-%d of %d", startIdx+1, endIdx, len(asgs))))
	}
	selected := asgs[cursor]
	rb.Newline()
	rb.WriteLine(SubtitleStyle().Render(selected.Name))
	rb.WriteLine(m.renderASGDetailsResponsive(selected, layout.DetailHeight))
	if overlay := m.renderScalingPrompt(ViewASGs); overlay != "" {
		rb.WriteLine(overlay)
	}
	if searchBar := m.renderSearchBar(ViewASGs); searchBar != "" {
		rb.WriteLine(searchBar)
	}
	if status := m.renderStatusMessage(); status != "" {
		rb.WriteLine(status)
	}
	rb.WriteLine(m.renderASGFooter())
	rb.WriteString(StatusBarStyle().Width(m.width).Render(m.getStatusBar()))
	return rb.String()
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
// renderASGFooter renders the footer for ASG view
func (m Model) renderASGFooter() string {
	return RenderFooter(CommonFooterKeys("scale"))
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
