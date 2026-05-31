package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/johnlam90/aws-ssm/pkg/ui/tui/table"
)

// renderEC2Instances renders the EC2 instances main-panel content.
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

	cols := table.Allocate(ec2Columns(), m.mainWidth())
	b.WriteString(TableHeaderStyle().Render(table.FormatHeader(cols)))
	b.WriteString("\n")

	visibleHeight := calculateTableRows(m.height, 9, details)
	startIdx, endIdx := calculateBoundedVisibleRange(len(instances), cursor, visibleHeight)

	for i := startIdx; i < endIdx; i++ {
		inst := instances[i]
		row := table.FormatRow(ec2RowValues(inst), cols)
		b.WriteString(RenderSelectableRow(row, i == cursor))
		b.WriteString("\n")
	}

	b.WriteString("\n")
	detailTitle := fmt.Sprintf("%s · %s", normalizeValue(selected.Name, "(no name)", 0), selected.InstanceID)
	b.WriteString(SubtitleStyle().Render(detailTitle))
	b.WriteString("\n")
	b.WriteString(details)

	if searchBar := m.renderSearchBar(ViewEC2Instances); searchBar != "" {
		b.WriteString("\n")
		b.WriteString(searchBar)
	}

	return b.String()
}

func ec2Columns() []table.ColumnSpec {
	return []table.ColumnSpec{
		{Header: "NAME", MinWidth: 12, PrefWidth: 28, MaxWidth: 40, Align: "left"},
		{Header: "INSTANCE ID", MinWidth: 19, PrefWidth: 19, MaxWidth: 19, Align: "left"},
		{Header: "PRIV IP", MinWidth: 9, PrefWidth: 13, MaxWidth: 15, Align: "left"},
		{Header: "STATE", MinWidth: 5, PrefWidth: 5, MaxWidth: 5, Align: "center"},
		{Header: "TYPE", MinWidth: 6, PrefWidth: 10, MaxWidth: 16, Align: "left"},
		{Header: "AGE", MinWidth: 3, PrefWidth: 5, MaxWidth: 8, Align: "right"},
	}
}

func ec2RowValues(inst EC2Instance) []string {
	name := inst.Name
	if name == "" {
		name = "(no name)"
	}
	age := ""
	if !inst.LaunchTime.IsZero() {
		age = humanDurationShort(time.Since(inst.LaunchTime))
	}
	return []string{
		name,
		inst.InstanceID,
		inst.PrivateIP,
		"  " + table.StateBadge(inst.State) + "  ",
		inst.InstanceType,
		age,
	}
}

// humanDurationShort returns a single-token age like "3d", "12h",
// "5m", or "<1m". Used in tabular AGE columns.
func humanDurationShort(d time.Duration) string {
	if d < 0 {
		d = -d
	}
	if d < time.Minute {
		return "<1m"
	}
	if d < time.Hour {
		return fmt.Sprintf("%dm", int(d.Minutes()))
	}
	if d < 24*time.Hour {
		return fmt.Sprintf("%dh", int(d.Hours()))
	}
	days := int(d.Hours()) / 24
	if days < 30 {
		return fmt.Sprintf("%dd", days)
	}
	if days < 365 {
		return fmt.Sprintf("%dmo", days/30)
	}
	return fmt.Sprintf("%dy", days/365)
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
