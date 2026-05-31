package tui

import (
	"fmt"
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
	selected := clusters[cursor]
	details := limitRenderedLines(m.renderEKSDetails(selected), max(1, m.height-10))
	visibleRows := calculateTableRows(m.height, 9, details)
	startIdx, endIdx := calculateBoundedVisibleRange(len(clusters), cursor, visibleRows)

	cols := table.Allocate(eksColumns(), m.mainWidth())
	b.WriteString(TableHeaderStyle().Render(table.FormatHeader(cols)))
	b.WriteString("\n")

	for i := startIdx; i < endIdx; i++ {
		row := table.FormatRow(eksRowValues(clusters[i]), cols)
		b.WriteString(RenderSelectableRow(row, i == cursor))
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(SubtitleStyle().Render(selected.Name))
	b.WriteString("\n")
	b.WriteString(details)

	if searchBar := m.renderSearchBar(ViewEKSClusters); searchBar != "" {
		b.WriteString("\n")
		b.WriteString(searchBar)
	}

	return b.String()
}

func (m Model) renderEKSDetails(c EKSCluster) string {
	var b strings.Builder

	b.WriteString("  Cluster:\n")
	fmt.Fprintf(&b, "    Status:  %s\n", StateStyle(strings.ToLower(c.Status)))
	fmt.Fprintf(&b, "    Version: %s\n", normalizeValue(c.Version, "unknown", 0))

	if strings.TrimSpace(c.Arn) != "" {
		b.WriteString("\n  Identity:\n")
		fmt.Fprintf(&b, "    ARN: %s\n", c.Arn)
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
