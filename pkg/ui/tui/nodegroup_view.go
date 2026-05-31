package tui

import (
	"fmt"
	"strings"
)

// renderNodeGroups renders the EKS node groups view
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

	header := m.renderHeader("EKS Node Groups", fmt.Sprintf("%d node groups", len(groups)))
	b.WriteString(header)
	b.WriteString("\n\n")
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
	b.WriteString("\n")

	b.WriteString("\n")
	if overlay := m.renderScalingPrompt(ViewNodeGroups); overlay != "" {
		b.WriteString(overlay)
		b.WriteString("\n")
	}
	if ltOverlay := m.renderLaunchTemplatePrompt(); ltOverlay != "" {
		b.WriteString(ltOverlay)
		b.WriteString("\n")
	}
	if searchBar := m.renderSearchBar(ViewNodeGroups); searchBar != "" {
		b.WriteString(searchBar)
		b.WriteString("\n")
	}
	if status := m.renderStatusMessage(); status != "" {
		b.WriteString(status)
		b.WriteString("\n")
	}
	b.WriteString(m.renderNodeGroupFooter())
	b.WriteString("\n")
	b.WriteString(StatusBarStyle().Width(m.width).Render(m.getStatusBar()))
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
	const fixedLines = 9 // title, blank, table header, detail title, spacing, footer, status bar
	detailLines := countRenderedLines(details)
	rows := terminalHeight - fixedLines - detailLines
	if rows < 1 {
		return 1
	}
	return rows
}

// renderNodeGroupState renders loading/error/empty states
func (m Model) renderNodeGroupState(groups []NodeGroup) string {
	var b strings.Builder
	header := m.renderHeader("EKS Node Groups", fmt.Sprintf("%d node groups", len(groups)))
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
	if len(groups) == 0 {
		b.WriteString(SubtitleStyle().Render("No EKS node groups found"))
		b.WriteString("\n\n")
		b.WriteString(HelpStyle().Render("esc:back"))
		b.WriteString("\n")
		b.WriteString(StatusBarStyle().Render(m.getStatusBar()))
		return b.String()
	}
	return ""
}

func calculateNodeGroupVisibleRange(total, cursor, visibleHeight int) (int, int) {
	return calculateBoundedVisibleRange(total, cursor, visibleHeight)
}

// renderNodeGroupFooter renders footer controls for node group view
func (m Model) renderNodeGroupFooter() string {
	keys := []struct {
		key  string
		desc string
	}{
		{"↑/k", "up"},
		{"↓/j", "down"},
		{"g/G", "top/bottom"},
		{"enter", "scale"},
		{"u/U", "update LT"},
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
