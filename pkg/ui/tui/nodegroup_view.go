package tui

import (
	"fmt"
	"strings"
)

// renderNodeGroups renders the EKS node groups main-panel content.
func (m Model) renderNodeGroups() string {
	groups := m.getNodeGroups()
	if s := m.renderNodeGroupState(groups); s != "" {
		return s
	}
	var b strings.Builder
	cursor := clampIndex(m.cursor, len(groups))
	selected := groups[cursor]
	details := renderNodeGroupDetails(selected)
	visibleRows := calculateNodeGroupTableRows(m.height, details)

	b.WriteString(TableHeaderStyle().Render(fmt.Sprintf("  %-24s %-28s %-10s %8s %8s %8s %8s",
		"CLUSTER", "NODE GROUP", "STATUS", "DESIRED", "MIN", "MAX", "CURRENT")))
	b.WriteString("\n")
	startIdx, endIdx := calculateNodeGroupVisibleRange(len(groups), cursor, visibleRows)
	for i := startIdx; i < endIdx; i++ {
		ng := groups[i]
		cluster := normalizeValue(ng.ClusterName, "unknown", 24)
		name := normalizeValue(ng.Name, "n/a", 28)
		status := RenderStateCell(ng.Status, 10)
		row := fmt.Sprintf("  %-24s %-28s %s %8d %8d %8d %8d",
			cluster, name, status, ng.DesiredSize, ng.MinSize, ng.MaxSize, ng.CurrentSize)
		b.WriteString(RenderSelectableRow(row, i == cursor))
		b.WriteString("\n")
	}
	b.WriteString("\n")
	b.WriteString(SubtitleStyle().Render(fmt.Sprintf("%s / %s", selected.ClusterName, selected.Name)))
	b.WriteString("\n")
	b.WriteString(details)

	if overlay := m.renderScalingPrompt(ViewNodeGroups); overlay != "" {
		b.WriteString("\n")
		b.WriteString(overlay)
	}
	if ltOverlay := m.renderLaunchTemplatePrompt(); ltOverlay != "" {
		b.WriteString("\n")
		b.WriteString(ltOverlay)
	}
	if searchBar := m.renderSearchBar(ViewNodeGroups); searchBar != "" {
		b.WriteString("\n")
		b.WriteString(searchBar)
	}
	if status := m.renderStatusMessage(); status != "" {
		b.WriteString("\n")
		b.WriteString(status)
	}
	return b.String()
}

func renderNodeGroupDetails(selected NodeGroup) string {
	var b strings.Builder
	instanceTypes := strings.Join(selected.InstanceTypes, ", ")
	if instanceTypes == "" {
		instanceTypes = "n/a"
	}

	fmt.Fprintf(&b, "  Version: %s\n", normalizeValue(selected.Version, "unknown", 0))
	fmt.Fprintf(&b, "  Status:  %s\n", StateStyle(strings.ToLower(selected.Status)))
	fmt.Fprintf(&b, "  Scaling: desired %d | min %d | max %d | current %d\n",
		selected.DesiredSize, selected.MinSize, selected.MaxSize, selected.CurrentSize)
	fmt.Fprintf(&b, "  Instances: %s\n", instanceTypes)
	if strings.TrimSpace(selected.LaunchTemplateID) != "" || strings.TrimSpace(selected.LaunchTemplateName) != "" {
		ltName := normalizeValue(selected.LaunchTemplateName, "n/a", 0)
		ltVersion := normalizeValue(selected.LaunchTemplateVersion, "n/a", 0)
		fmt.Fprintf(&b, "  Launch template: %s (version %s)\n", ltName, ltVersion)
		fmt.Fprintf(&b, "  Launch template ID: %s\n", normalizeValue(selected.LaunchTemplateID, "n/a", 0))
	} else {
		b.WriteString("  Launch template: n/a\n")
	}
	fmt.Fprintf(&b, "  Created:   %s\n", normalizeValue(selected.CreatedAt, "unknown", 0))
	if lines := renderTagLines(selected.Tags); len(lines) > 0 {
		b.WriteString("  Tags:\n")
		for _, line := range lines {
			b.WriteString(line)
			b.WriteString("\n")
		}
	}
	return strings.TrimSuffix(b.String(), "\n")
}

func calculateNodeGroupTableRows(terminalHeight int, details string) int {
	if terminalHeight <= 0 {
		return 5
	}
	const fixedLines = 6 // table header, detail title, spacing
	detailLines := countRenderedLines(details)
	rows := terminalHeight - fixedLines - detailLines
	if rows < 1 {
		return 1
	}
	return rows
}

// renderNodeGroupState renders loading/error/empty states inside the
// main panel; chrome remains visible above and below.
func (m Model) renderNodeGroupState(groups []NodeGroup) string {
	if m.loading {
		return m.renderLoading()
	}
	if m.err != nil {
		return m.renderError()
	}
	if len(groups) == 0 {
		return SubtitleStyle().Render("No EKS node groups found")
	}
	return ""
}

func calculateNodeGroupVisibleRange(total, cursor, visibleHeight int) (int, int) {
	return calculateBoundedVisibleRange(total, cursor, visibleHeight)
}

// clampIndex ensures cursor stays within list bounds
func clampIndex(idx, length int) int {
	switch {
	case length == 0:
		return 0
	case idx < 0:
		return 0
	case idx >= length:
		return length - 1
	default:
		return idx
	}
}
