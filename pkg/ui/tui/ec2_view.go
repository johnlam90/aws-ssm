package tui

import (
	"fmt"
	"strings"
	"time"
)

// renderEC2Instances renders the EC2 instances view using pooled string builder
func (m Model) renderEC2Instances() string {
	rb := NewRenderBuffer()

	instances := m.getEC2Instances()

	// Header
	header := m.renderHeader("EC2 Instances", fmt.Sprintf("%d instances", len(instances)))
	rb.WriteLine(header).Newline()

	// Show loading or error - minimal
	if m.loading {
		rb.WriteLine(m.renderLoading())
		rb.WriteString(StatusBarStyle().Width(m.width).Render(m.getStatusBar()))
		return rb.String()
	}

	if m.err != nil {
		rb.WriteLine(m.renderError()).Newline()
		rb.WriteLine(HelpStyle().Render("esc:back"))
		rb.WriteString(StatusBarStyle().Width(m.width).Render(m.getStatusBar()))
		return rb.String()
	}

	// No instances
	if len(instances) == 0 {
		rb.WriteLine(SubtitleStyle().Render("No EC2 instances found")).Newline()
		rb.WriteLine(HelpStyle().Render("esc:back"))
		rb.WriteString(StatusBarStyle().Width(m.width).Render(m.getStatusBar()))
		return rb.String()
	}

	cursor := clampIndex(m.cursor, len(instances))

	// Get responsive column widths based on terminal width
	nameW, idW, ipW, stateW, typeW := EC2ColumnWidths(m.width)

	// Table header - clean and aligned with responsive widths
	headerRow := fmt.Sprintf("  %-*s %-*s %-*s %-*s %-*s",
		nameW, "NAME", idW, "INSTANCE ID", ipW, "PRIVATE IP", stateW, "STATE", typeW, "TYPE")
	rb.WriteLine(TableHeaderStyle().Render(headerRow))

	// Calculate responsive vertical layout
	layout := EC2Layout(m.height)
	startIdx, endIdx := CalculateVisibleRange(len(instances), cursor, layout.TableHeight)

	// Render instances with proper alignment using responsive widths
	for i := startIdx; i < endIdx; i++ {
		inst := instances[i]
		name := inst.Name
		if name == "" {
			name = "(no name)"
		}
		if len(name) > nameW {
			name = TruncateWithEllipsis(name, nameW)
		}

		state := RenderStateCell(inst.State, stateW)
		row := fmt.Sprintf("  %-*s %-*s %-*s %s %-*s",
			nameW, name, idW, inst.InstanceID, ipW, inst.PrivateIP, state, typeW, inst.InstanceType)

		rb.WriteLine(RenderSelectableRow(row, i == cursor))
	}

	// Pagination indicator
	if len(instances) > endIdx-startIdx {
		pageInfo := fmt.Sprintf("Showing %d-%d of %d", startIdx+1, endIdx, len(instances))
		rb.Newline().WriteString(SubtitleStyle().Render(pageInfo))
	}

	selected := instances[cursor]
	rb.Newline()
	detailTitle := fmt.Sprintf("%s (%s)", normalizeValue(selected.Name, "(no name)", 0), selected.InstanceID)
	rb.WriteLine(SubtitleStyle().Render(detailTitle))
	rb.WriteString(m.renderEC2DetailsResponsive(selected, layout.DetailHeight))

	if searchBar := m.renderSearchBar(ViewEC2Instances); searchBar != "" {
		rb.Newline().WriteLine(searchBar)
	}

	// Footer
	rb.Newline().WriteLine(m.renderEC2Footer())

	// Status bar
	rb.WriteString(StatusBarStyle().Width(m.width).Render(m.getStatusBar()))

	return rb.String()
}

// renderEC2Footer renders the footer for EC2 view
func (m Model) renderEC2Footer() string {
	return RenderFooter(CommonFooterKeys("connect"))
}


// renderEC2DetailsResponsive renders EC2 details with responsive height
func (m Model) renderEC2DetailsResponsive(inst EC2Instance, maxLines int) string {
	var lines []string

	// Basic info section (always show core info)
	lines = append(lines, "  Basic Info:")
	lines = append(lines, fmt.Sprintf("    State:       %s", StateStyle(strings.ToLower(inst.State))))
	lines = append(lines, fmt.Sprintf("    Type:        %s", normalizeValue(inst.InstanceType, "unknown", 0)))
	lines = append(lines, fmt.Sprintf("    AZ:          %s", normalizeValue(inst.AvailabilityZone, "unknown", 0)))

	if maxLines > 5 && !inst.LaunchTime.IsZero() {
		lines = append(lines, fmt.Sprintf("    Launch:      %s", formatRelativeTimestamp(inst.LaunchTime)))
		if maxLines > 6 {
			lines = append(lines, fmt.Sprintf("    Uptime:      %s", humanDuration(time.Since(inst.LaunchTime))))
		}
	}

	// Network section (show if space available)
	if maxLines > 8 {
		lines = append(lines, "")
		lines = append(lines, "  Network:")
		lines = append(lines, fmt.Sprintf("    Private IP:  %s", normalizeValue(inst.PrivateIP, "n/a", 0)))
		if maxLines > 10 {
			lines = append(lines, fmt.Sprintf("    Private DNS: %s", normalizeValue(inst.PrivateDNS, "n/a", 0)))
		}
		if maxLines > 11 {
			lines = append(lines, fmt.Sprintf("    Public IP:   %s", normalizeValue(inst.PublicIP, "n/a", 0)))
		}
		if maxLines > 12 {
			lines = append(lines, fmt.Sprintf("    Public DNS:  %s", normalizeValue(inst.PublicDNS, "n/a", 0)))
		}
	}

	// Security section (show if space available)
	if maxLines > 14 {
		lines = append(lines, "")
		lines = append(lines, "  Security:")
		if inst.InstanceProfile != "" {
			lines = append(lines, fmt.Sprintf("    IAM Role:    %s", inst.InstanceProfile))
		} else {
			lines = append(lines, "    IAM Role:    n/a")
		}
		// Security groups
		remainingLines := maxLines - len(lines) - 2 // Reserve space for tags header if needed
		if len(inst.SecurityGroups) > 0 {
			for i, sg := range inst.SecurityGroups {
				if i >= remainingLines {
					lines = append(lines, "    ...")
					break
				}
				lines = append(lines, fmt.Sprintf("    • %s", sg))
			}
		} else {
			lines = append(lines, "    • no security groups detected")
		}
	}

	// Tags section (show if space available)
	if maxLines > 18 {
		tagLines := renderTagLines(inst.Tags, "Name")
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
