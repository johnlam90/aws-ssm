package tui

import (
	"fmt"
	"strings"
)

// renderEKSClusters renders the EKS clusters main-panel content.
func (m Model) renderEKSClusters() string {
	var b strings.Builder

	clusters := m.getEKSClusters()

	if m.loading {
		b.WriteString(m.renderLoading())
		return b.String()
	}

	if m.err != nil {
		b.WriteString(m.renderError())
		return b.String()
	}

	if len(clusters) == 0 {
		b.WriteString(SubtitleStyle().Render("No EKS clusters found"))
		return b.String()
	}

	cursor := clampIndex(m.cursor, len(clusters))
	visibleRows := calculateTableRows(m.height, 7, "")
	startIdx, endIdx := calculateBoundedVisibleRange(len(clusters), cursor, visibleRows)

	headerRow := fmt.Sprintf("  %-45s %-15s %-10s", "NAME", "STATUS", "VERSION")
	b.WriteString(TableHeaderStyle().Render(headerRow))
	b.WriteString("\n")

	for i := startIdx; i < endIdx; i++ {
		cluster := clusters[i]
		name := cluster.Name
		if len(name) > 45 {
			name = name[:42] + "..."
		}

		status := StateStyle(cluster.Status)
		row := fmt.Sprintf("  %-45s %-15s %-10s", name, status, cluster.Version)

		b.WriteString(RenderSelectableRow(row, i == cursor))
		b.WriteString("\n")
	}

	if searchBar := m.renderSearchBar(ViewEKSClusters); searchBar != "" {
		b.WriteString("\n")
		b.WriteString(searchBar)
	}

	return b.String()
}
