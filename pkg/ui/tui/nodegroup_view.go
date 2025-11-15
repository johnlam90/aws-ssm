package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// renderNodeGroups renders the EKS node groups view
func (m Model) renderNodeGroups() string {
	var b strings.Builder

	groups := m.getNodeGroups()

	header := m.renderHeader("EKS Node Groups", fmt.Sprintf("%d node groups", len(groups)))
	b.WriteString(header)
	b.WriteString("\n\n")

	if m.loading {
		b.WriteString(m.renderLoading())
		b.WriteString("\n")
		b.WriteString(StatusBarStyle.Render(m.getStatusBar()))
		return b.String()
	}

	if m.err != nil {
		b.WriteString(m.renderError())
		b.WriteString("\n\n")
		b.WriteString(HelpStyle.Render("esc:back"))
		b.WriteString("\n")
		b.WriteString(StatusBarStyle.Render(m.getStatusBar()))
		return b.String()
	}

	if len(groups) == 0 {
		b.WriteString(SubtitleStyle.Render("No EKS node groups found"))
		b.WriteString("\n\n")
		b.WriteString(HelpStyle.Render("esc:back"))
		b.WriteString("\n")
		b.WriteString(StatusBarStyle.Render(m.getStatusBar()))
		return b.String()
	}

	cursor := clampIndex(m.cursor, len(groups))
	visibleHeight := m.height - 12
	if visibleHeight < 5 {
		visibleHeight = len(groups)
	}

	startIdx := 0
	endIdx := len(groups)
	if len(groups) > visibleHeight && visibleHeight > 0 {
		startIdx = cursor - visibleHeight/2
		if startIdx < 0 {
			startIdx = 0
		}
		endIdx = startIdx + visibleHeight
		if endIdx > len(groups) {
			endIdx = len(groups)
			startIdx = endIdx - visibleHeight
			if startIdx < 0 {
				startIdx = 0
			}
		}
	}

	headerRow := fmt.Sprintf("  %-24s %-28s %-10s %8s %8s %8s %8s",
		"CLUSTER", "NODE GROUP", "STATUS", "DESIRED", "MIN", "MAX", "CURRENT")
	b.WriteString(TableHeaderStyle.Render(headerRow))
	b.WriteString("\n")

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

	selected := groups[cursor]
	b.WriteString("\n")
	detailTitle := fmt.Sprintf("%s / %s", selected.ClusterName, selected.Name)
	b.WriteString(SubtitleStyle.Render(detailTitle))
	b.WriteString("\n")

	instanceTypes := strings.Join(selected.InstanceTypes, ", ")
	if instanceTypes == "" {
		instanceTypes = "n/a"
	}

	b.WriteString(fmt.Sprintf("  Version: %s\n", normalizeValue(selected.Version, "unknown", 0)))
	b.WriteString(fmt.Sprintf("  Status:  %s\n", StateStyle(strings.ToLower(selected.Status))))
	b.WriteString(fmt.Sprintf("  Scaling: desired %d | min %d | max %d | current %d\n",
		selected.DesiredSize, selected.MinSize, selected.MaxSize, selected.CurrentSize))
	b.WriteString(fmt.Sprintf("  Instances: %s\n", instanceTypes))
	if strings.TrimSpace(selected.LaunchTemplateID) != "" || strings.TrimSpace(selected.LaunchTemplateName) != "" {
		ltName := normalizeValue(selected.LaunchTemplateName, "n/a", 0)
		ltVersion := normalizeValue(selected.LaunchTemplateVersion, "n/a", 0)
		b.WriteString(fmt.Sprintf("  Launch template: %s (version %s)\n", ltName, ltVersion))
		b.WriteString(fmt.Sprintf("  Launch template ID: %s\n", normalizeValue(selected.LaunchTemplateID, "n/a", 0)))
	} else {
		b.WriteString("  Launch template: n/a\n")
	}
	b.WriteString(fmt.Sprintf("  Created:   %s\n", normalizeValue(selected.CreatedAt, "unknown", 0)))

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
	b.WriteString(StatusBarStyle.Width(m.width).Render(m.getStatusBar()))

	return b.String()
}

// handleNodeGroupKeys manages keybindings for node group view
func (m Model) handleNodeGroupKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	groups := m.getNodeGroups()
	switch msg.String() {
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down", "j":
		if m.cursor < len(groups)-1 {
			m.cursor++
		}
	case "g":
		if len(groups) > 0 {
			m.cursor = 0
		}
	case "G":
		if len(groups) > 0 {
			m.cursor = len(groups) - 1
		}
	case "enter", " ":
		if m.cursor < len(groups) {
			ng := groups[m.cursor]
			m = m.startNodeGroupScaling(ng)
		}
	case "u", "U":
		if m.cursor < len(groups) {
			ng := groups[m.cursor]
			var cmd tea.Cmd
			m, cmd = m.startNodeGroupLaunchTemplateUpdate(ng)
			if cmd != nil {
				return m, cmd
			}
		}
	case "r":
		m.loading = true
		m.loadingMsg = "Refreshing node groups..."
		m.err = nil
		m.statusMessage = ""
		return m, LoadNodeGroupsCmd(m.ctx, m.client)
	case "esc":
		return m.navigateBack(), nil
	}

	return m, nil
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
		{"u", "update LT"},
		{"r", "refresh"},
		{"/", "search"},
		{"esc", "back"},
	}

	var parts []string
	for _, k := range keys {
		keyStyle := StatusBarKeyStyle.Render(k.key)
		descStyle := StatusBarValueStyle.Render(k.desc)
		parts = append(parts, fmt.Sprintf("%s %s", keyStyle, descStyle))
	}

	return HelpStyle.Render(strings.Join(parts, " • "))
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
