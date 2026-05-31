package tui

import (
	"strings"

	"github.com/johnlam90/aws-ssm/pkg/ui/tui/table"
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

	cols := table.Allocate(eksColumns(), m.mainWidth())
	b.WriteString(TableHeaderStyle().Render(table.FormatHeader(cols)))
	b.WriteString("\n")

	for i := startIdx; i < endIdx; i++ {
		row := table.FormatRow(eksRowValues(clusters[i]), cols)
		b.WriteString(RenderSelectableRow(row, i == cursor))
		b.WriteString("\n")
	}

	if searchBar := m.renderSearchBar(ViewEKSClusters); searchBar != "" {
		b.WriteString("\n")
		b.WriteString(searchBar)
	}

	return b.String()
}

func eksColumns() []table.ColumnSpec {
	return []table.ColumnSpec{
		{Header: "NAME", MinWidth: 12, PrefWidth: 32, MaxWidth: 50, Align: "left"},
		{Header: "STATE", MinWidth: 5, PrefWidth: 5, MaxWidth: 5, Align: "center"},
		{Header: "VERSION", MinWidth: 4, PrefWidth: 8, MaxWidth: 10, Align: "left"},
	}
}

func eksRowValues(c EKSCluster) []string {
	return []string{
		c.Name,
		"  " + table.StateBadge(c.Status) + "  ",
		c.Version,
	}
}
