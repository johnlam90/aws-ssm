package tui

import (
	"fmt"
	"strings"
)

// renderEKSClusters renders the EKS clusters view - minimal design
func (m Model) renderEKSClusters() string {
	var b strings.Builder

	clusters := m.getEKSClusters()

	// Header - simple
	header := m.renderHeader("EKS Clusters", fmt.Sprintf("%d clusters", len(clusters)))
	b.WriteString(header)
	b.WriteString("\n\n")

	// Show loading or error - minimal
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

	// No clusters
	if len(clusters) == 0 {
		b.WriteString(SubtitleStyle().Render("No EKS clusters found"))
		b.WriteString("\n\n")
		b.WriteString(HelpStyle().Render("esc:back"))
		b.WriteString("\n")
		b.WriteString(StatusBarStyle().Width(m.width).Render(m.getStatusBar()))
		return b.String()
	}

	// Get responsive column widths based on terminal width
	nameW, statusW, versionW := EKSColumnWidths(m.width)

	// Table header - clean and aligned with responsive widths
	headerRow := fmt.Sprintf("  %-*s %-*s %-*s", nameW, "NAME", statusW, "STATUS", versionW, "VERSION")
	b.WriteString(TableHeaderStyle().Render(headerRow))
	b.WriteString("\n")

	// Calculate responsive vertical layout
	layout := EKSLayout(m.height)
	cursor := clampIndex(m.cursor, len(clusters))
	startIdx, endIdx := calculateEKSVisibleRange(len(clusters), cursor, layout.TableHeight)

	// Render clusters with proper alignment using responsive widths
	for i := startIdx; i < endIdx; i++ {
		cluster := clusters[i]
		// Truncate name if too long
		name := cluster.Name
		if len(name) > nameW {
			name = TruncateWithEllipsis(name, nameW)
		}

		status := StateStyle(cluster.Status)
		row := fmt.Sprintf("  %-*s %-*s %-*s", nameW, name, statusW, status, versionW, cluster.Version)

		b.WriteString(RenderSelectableRow(row, i == cursor))
		b.WriteString("\n")
	}

	// Pagination indicator
	if len(clusters) > endIdx-startIdx {
		b.WriteString("\n")
		b.WriteString(SubtitleStyle().Render(fmt.Sprintf("Showing %d-%d of %d", startIdx+1, endIdx, len(clusters))))
	}

	// Footer
	b.WriteString("\n")
	if searchBar := m.renderSearchBar(ViewEKSClusters); searchBar != "" {
		b.WriteString(searchBar)
		b.WriteString("\n")
	}
	b.WriteString(m.renderEKSFooter())

	// Status bar
	b.WriteString("\n")
	b.WriteString(StatusBarStyle().Width(m.width).Render(m.getStatusBar()))

	return b.String()
}

// calculateEKSVisibleRange calculates the visible range for EKS clusters
func calculateEKSVisibleRange(total, cursor, visibleHeight int) (int, int) {
	if visibleHeight < 3 || total <= visibleHeight {
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

// renderEKSFooter renders the footer for EKS view
func (m Model) renderEKSFooter() string {
	keys := []struct {
		key  string
		desc string
	}{
		{"↑/k", "up"},
		{"↓/j", "down"},
		{"g/G", "top/bottom"},
		{"enter", "details"},
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
