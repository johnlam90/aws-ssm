package tui

import (
	"fmt"
	"strings"
	"time"
)

// renderEC2Instances renders the EC2 instances main-panel content.
// Chrome (header, footer, status bar) is owned by the chrome package
// and is composed at the View() level.
func (m Model) renderEC2Instances() string {
	var b strings.Builder

	instances := m.getEC2Instances()

	if m.loading {
		b.WriteString(m.renderLoading())
		return b.String()
	}

	if m.err != nil {
		b.WriteString(m.renderError())
		return b.String()
	}

	if len(instances) == 0 {
		b.WriteString(SubtitleStyle().Render("No EC2 instances found"))
		return b.String()
	}

	cursor := clampIndex(m.cursor, len(instances))
	selected := instances[cursor]
	details := limitRenderedLines(m.renderEC2Details(selected), max(1, m.height-10))

	headerRow := fmt.Sprintf("  %-32s %-20s %-15s %-12s %-15s",
		"NAME", "INSTANCE ID", "PRIVATE IP", "STATE", "TYPE")
	b.WriteString(TableHeaderStyle().Render(headerRow))
	b.WriteString("\n")

	visibleHeight := calculateTableRows(m.height, 9, details)
	startIdx, endIdx := calculateBoundedVisibleRange(len(instances), cursor, visibleHeight)

	for i := startIdx; i < endIdx; i++ {
		inst := instances[i]
		name := inst.Name
		if name == "" {
			name = "(no name)"
		}
		if len(name) > 32 {
			name = name[:29] + "..."
		}

		state := RenderStateCell(inst.State, 12)
		row := fmt.Sprintf("  %-32s %-20s %-15s %s %-15s",
			name, inst.InstanceID, inst.PrivateIP, state, inst.InstanceType)

		b.WriteString(RenderSelectableRow(row, i == cursor))
		b.WriteString("\n")
	}

	if visibleHeight > 0 && len(instances) > visibleHeight {
		pageInfo := fmt.Sprintf("Showing %d-%d of %d", startIdx+1, endIdx, len(instances))
		b.WriteString("\n")
		b.WriteString(SubtitleStyle().Render(pageInfo))
	}

	b.WriteString("\n")
	detailTitle := fmt.Sprintf("%s (%s)", normalizeValue(selected.Name, "(no name)", 0), selected.InstanceID)
	b.WriteString(SubtitleStyle().Render(detailTitle))
	b.WriteString("\n")
	b.WriteString(details)

	if searchBar := m.renderSearchBar(ViewEC2Instances); searchBar != "" {
		b.WriteString("\n")
		b.WriteString(searchBar)
	}

	return b.String()
}

func (m Model) renderEC2Details(inst EC2Instance) string {
	var b strings.Builder

	b.WriteString("  Basic Info:\n")
	fmt.Fprintf(&b, "    State:       %s\n", StateStyle(strings.ToLower(inst.State)))
	fmt.Fprintf(&b, "    Type:        %s\n", normalizeValue(inst.InstanceType, "unknown", 0))
	fmt.Fprintf(&b, "    AZ:          %s\n", normalizeValue(inst.AvailabilityZone, "unknown", 0))
	if !inst.LaunchTime.IsZero() {
		fmt.Fprintf(&b, "    Launch:      %s\n", formatRelativeTimestamp(inst.LaunchTime))
		fmt.Fprintf(&b, "    Uptime:      %s\n", humanDuration(time.Since(inst.LaunchTime)))
	}

	b.WriteString("\n  Network:\n")
	fmt.Fprintf(&b, "    Private IP:  %s\n", normalizeValue(inst.PrivateIP, "n/a", 0))
	fmt.Fprintf(&b, "    Private DNS: %s\n", normalizeValue(inst.PrivateDNS, "n/a", 0))
	fmt.Fprintf(&b, "    Public IP:   %s\n", normalizeValue(inst.PublicIP, "n/a", 0))
	fmt.Fprintf(&b, "    Public DNS:  %s\n", normalizeValue(inst.PublicDNS, "n/a", 0))

	b.WriteString("\n  Security:\n")
	if inst.InstanceProfile != "" {
		fmt.Fprintf(&b, "    IAM Role:    %s\n", inst.InstanceProfile)
	} else {
		b.WriteString("    IAM Role:    n/a\n")
	}
	if len(inst.SecurityGroups) > 0 {
		for _, sg := range inst.SecurityGroups {
			fmt.Fprintf(&b, "    • %s\n", sg)
		}
	} else {
		b.WriteString("    • no security groups detected\n")
	}

	if lines := renderTagLines(inst.Tags, "Name"); len(lines) > 0 {
		b.WriteString("\n  Tags:\n")
		for _, line := range lines {
			b.WriteString(line)
			b.WriteString("\n")
		}
	}

	return b.String()
}
