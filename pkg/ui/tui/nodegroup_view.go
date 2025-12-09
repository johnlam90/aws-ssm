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
	header := m.renderHeader("EKS Node Groups", fmt.Sprintf("%d node groups", len(groups)))
	b.WriteString(header)
	b.WriteString("\n\n")

	// Get responsive column widths based on terminal width
	clusterW, nodeGroupW, statusW, desiredW, minW, maxW, currentW := NodeGroupColumnWidths(m.width)

	b.WriteString(TableHeaderStyle().Render(fmt.Sprintf("  %-*s %-*s %-*s %*s %*s %*s %*s",
		clusterW, "CLUSTER", nodeGroupW, "NODE GROUP", statusW, "STATUS",
		desiredW, "DESIRED", minW, "MIN", maxW, "MAX", currentW, "CURRENT")))
	b.WriteString("\n")
	cursor := clampIndex(m.cursor, len(groups))

	// Calculate responsive vertical layout
	layout := NodeGroupLayout(m.height)
	startIdx, endIdx := calculateNodeGroupVisibleRange(len(groups), cursor, layout.TableHeight)

	for i := startIdx; i < endIdx; i++ {
		ng := groups[i]
		cluster := normalizeValue(ng.ClusterName, "unknown", clusterW)
		name := normalizeValue(ng.Name, "n/a", nodeGroupW)
		status := RenderStateCell(ng.Status, statusW)
		row := fmt.Sprintf("  %-*s %-*s %s %*d %*d %*d %*d",
			clusterW, cluster, nodeGroupW, name, status,
			desiredW, ng.DesiredSize, minW, ng.MinSize, maxW, ng.MaxSize, currentW, ng.CurrentSize)
		b.WriteString(RenderSelectableRow(row, i == cursor))
		b.WriteString("\n")
	}

	// Pagination indicator when not showing all items
	if len(groups) > endIdx-startIdx {
		b.WriteString(SubtitleStyle().Render(fmt.Sprintf("Showing %d-%d of %d", startIdx+1, endIdx, len(groups))))
		b.WriteString("\n")
	}

	selected := groups[cursor]
	b.WriteString("\n")
	b.WriteString(SubtitleStyle().Render(fmt.Sprintf("%s / %s", selected.ClusterName, selected.Name)))
	b.WriteString("\n")

	// Render detail section with responsive height
	detailLines := m.renderNodeGroupDetails(selected, layout.DetailHeight)
	b.WriteString(detailLines)

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

// renderNodeGroupDetails renders the detail section with responsive height
func (m Model) renderNodeGroupDetails(selected NodeGroup, maxLines int) string {
	var lines []string

	instanceTypes := strings.Join(selected.InstanceTypes, ", ")
	if instanceTypes == "" {
		instanceTypes = "n/a"
	}

	// Core details (always shown)
	lines = append(lines, fmt.Sprintf("  Version: %s", normalizeValue(selected.Version, "unknown", 0)))
	lines = append(lines, fmt.Sprintf("  Status:  %s", StateStyle(strings.ToLower(selected.Status))))
	lines = append(lines, fmt.Sprintf("  Scaling: desired %d | min %d | max %d | current %d",
		selected.DesiredSize, selected.MinSize, selected.MaxSize, selected.CurrentSize))
	lines = append(lines, fmt.Sprintf("  Instances: %s", instanceTypes))

	// Launch template info (show if space available)
	if maxLines > 5 {
		if strings.TrimSpace(selected.LaunchTemplateID) != "" || strings.TrimSpace(selected.LaunchTemplateName) != "" {
			ltName := normalizeValue(selected.LaunchTemplateName, "n/a", 0)
			ltVersion := normalizeValue(selected.LaunchTemplateVersion, "n/a", 0)
			lines = append(lines, fmt.Sprintf("  Launch template: %s (version %s)", ltName, ltVersion))
			if maxLines > 6 {
				lines = append(lines, fmt.Sprintf("  Launch template ID: %s", normalizeValue(selected.LaunchTemplateID, "n/a", 0)))
			}
		} else {
			lines = append(lines, "  Launch template: n/a")
		}
	}

	// Created timestamp (show if space available)
	if maxLines > 7 {
		lines = append(lines, fmt.Sprintf("  Created:   %s", normalizeValue(selected.CreatedAt, "unknown", 0)))
	}

	// Tags (show if space available and we have room for at least header + 1 tag)
	if maxLines > 9 {
		tagLines := renderTagLines(selected.Tags)
		if len(tagLines) > 0 {
			lines = append(lines, "  Tags:")
			remainingLines := maxLines - len(lines) - 1 // Reserve 1 for potential "..."
			for i, line := range tagLines {
				if i >= remainingLines {
					lines = append(lines, "    ...")
					break
				}
				lines = append(lines, line)
			}
		}
	}

	return strings.Join(lines, "\n")
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
	if len(groups) == 0 {
		b.WriteString(SubtitleStyle().Render("No EKS node groups found"))
		b.WriteString("\n\n")
		b.WriteString(HelpStyle().Render("esc:back"))
		b.WriteString("\n")
		b.WriteString(StatusBarStyle().Width(m.width).Render(m.getStatusBar()))
		return b.String()
	}
	return ""
}

func calculateNodeGroupVisibleRange(total, cursor, visibleHeight int) (int, int) {
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
