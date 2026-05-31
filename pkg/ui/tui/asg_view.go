package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/johnlam90/aws-ssm/pkg/ui/tui/table"
)

// renderASGs renders the Auto Scaling Groups main-panel content.
func (m Model) renderASGs() string {
	asgs := m.getASGs()

	var search strings.Builder
	search.WriteString(m.renderSearchBar(ViewASGs))
	search.WriteString("\n")

	if m.loading {
		return search.String() + m.renderLoading()
	}
	if m.err != nil {
		return search.String() + m.renderError()
	}
	if len(asgs) == 0 {
		return search.String() + SubtitleStyle().Render("No Auto Scaling Groups found")
	}

	var b strings.Builder
	b.WriteString(search.String())
	cursor := clampIndex(m.cursor, len(asgs))
	selected := asgs[cursor]
	details := limitRenderedLines(m.renderASGDetails(selected), max(1, m.height-10))
	visibleRows := calculateTableRows(m.height, 9, details)

	cols := table.Allocate(asgColumns(), m.mainWidth())
	b.WriteString(TableHeaderStyle().Render(table.FormatHeader(cols)))
	b.WriteString("\n")

	startIdx, endIdx := m.calculateVisibleRange(len(asgs), cursor, visibleRows)
	for i := startIdx; i < endIdx; i++ {
		row := table.FormatRow(asgRowValues(asgs[i]), cols)
		b.WriteString(RenderSelectableRow(row, i == cursor))
		b.WriteString("\n")
	}
	if endIdx-startIdx > 0 && len(asgs) > endIdx-startIdx {
		b.WriteString("\n")
		b.WriteString(SubtitleStyle().Render(fmt.Sprintf("Showing %d-%d of %d", startIdx+1, endIdx, len(asgs))))
	}
	if !m.hideDetail {
		b.WriteString("\n")
		b.WriteString(SubtitleStyle().Render(selected.Name))
		b.WriteString("\n")
		b.WriteString(details)
	}

	if overlay := m.renderScalingPrompt(ViewASGs); overlay != "" {
		b.WriteString("\n")
		b.WriteString(overlay)
	}
	if status := m.renderStatusMessage(); status != "" {
		b.WriteString("\n")
		b.WriteString(status)
	}

	return b.String()
}

func asgColumns() []table.ColumnSpec {
	return []table.ColumnSpec{
		{Header: "NAME", MinWidth: 12, PrefWidth: 36, MaxWidth: 60, Align: "left"},
		{Header: "STATE", MinWidth: 5, PrefWidth: 5, MaxWidth: 5, Align: "center"},
		{Header: "DES", MinWidth: 3, PrefWidth: 5, MaxWidth: 6, Align: "right"},
		{Header: "MIN", MinWidth: 3, PrefWidth: 5, MaxWidth: 6, Align: "right"},
		{Header: "MAX", MinWidth: 3, PrefWidth: 5, MaxWidth: 6, Align: "right"},
		{Header: "CUR", MinWidth: 3, PrefWidth: 5, MaxWidth: 6, Align: "right"},
		{Header: "AGE", MinWidth: 3, PrefWidth: 5, MaxWidth: 8, Align: "right"},
	}
}

func asgRowValues(a ASG) []string {
	age := ""
	if !a.CreatedAt.IsZero() {
		age = humanDurationShort(time.Since(a.CreatedAt))
	}
	return []string{
		a.Name,
		"  " + table.StateBadge(a.Status) + "  ",
		fmt.Sprintf("%d", a.DesiredCapacity),
		fmt.Sprintf("%d", a.MinSize),
		fmt.Sprintf("%d", a.MaxSize),
		fmt.Sprintf("%d", a.CurrentSize),
		age,
	}
}

func (m Model) calculateVisibleRange(total, cursor, visibleHeight int) (int, int) {
	return calculateBoundedVisibleRange(total, cursor, visibleHeight)
}

func (m Model) renderASGDetails(asg ASG) string {
	var b strings.Builder

	b.WriteString("  Scaling:\n")
	fmt.Fprintf(&b, "    Desired: %d  Min: %d  Max: %d  Current: %d\n",
		asg.DesiredCapacity, asg.MinSize, asg.MaxSize, asg.CurrentSize)
	if asg.Status != "" {
		fmt.Fprintf(&b, "    Status:  %s\n", StateStyle(strings.ToLower(asg.Status)))
	}
	if asg.HealthCheckType != "" {
		fmt.Fprintf(&b, "    Health Check: %s\n", asg.HealthCheckType)
	}
	if !asg.CreatedAt.IsZero() {
		fmt.Fprintf(&b, "    Created: %s\n", formatRelativeTimestamp(asg.CreatedAt))
	}

	b.WriteString("\n  Launch Configuration:\n")
	switch {
	case strings.TrimSpace(asg.LaunchTemplateName) != "":
		fmt.Fprintf(&b, "    Template: %s", asg.LaunchTemplateName)
		if strings.TrimSpace(asg.LaunchTemplateVersion) != "" {
			fmt.Fprintf(&b, " (version %s)", asg.LaunchTemplateVersion)
		}
		b.WriteString("\n")
	case strings.TrimSpace(asg.LaunchConfigurationName) != "":
		fmt.Fprintf(&b, "    Configuration: %s\n", asg.LaunchConfigurationName)
	default:
		b.WriteString("    Configuration: n/a\n")
	}

	if len(asg.AvailabilityZones) > 0 {
		b.WriteString("\n  Availability Zones:\n")
		for _, az := range asg.AvailabilityZones {
			fmt.Fprintf(&b, "    • %s\n", az)
		}
	}

	if len(asg.LoadBalancerNames) > 0 {
		b.WriteString("\n  Load Balancers:\n")
		for _, lb := range asg.LoadBalancerNames {
			fmt.Fprintf(&b, "    • %s\n", lb)
		}
	}

	if len(asg.TargetGroupARNs) > 0 {
		b.WriteString("\n  Target Groups:\n")
		for _, tg := range asg.TargetGroupARNs {
			fmt.Fprintf(&b, "    • %s\n", tg)
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
