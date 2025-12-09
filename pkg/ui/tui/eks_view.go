package tui

import (
	"fmt"
	"strings"
)

// renderEKSClusters renders the EKS clusters view using pooled string builder
func (m Model) renderEKSClusters() string {
	rb := NewRenderBuffer()

	clusters := m.getEKSClusters()

	// Header - simple
	header := m.renderHeader("EKS Clusters", fmt.Sprintf("%d clusters", len(clusters)))
	rb.WriteLine(header).Newline()

	// Show loading or error - minimal
	if m.loading {
		rb.WriteLine(m.renderLoading())
		rb.WriteString(StatusBarStyle().Width(m.width).Render(m.getStatusBar()))
		return rb.String()
	}

	if m.err != nil {
		rb.WriteLine(m.renderError()).Newline()
		rb.WriteLine(HelpStyle().Render("esc:back"))
		rb.WriteString(StatusBarStyle().Width(m.width).Render(m.getStatusBar()))
		return rb.String()
	}

	// No clusters
	if len(clusters) == 0 {
		rb.WriteLine(SubtitleStyle().Render("No EKS clusters found")).Newline()
		rb.WriteLine(HelpStyle().Render("esc:back"))
		rb.WriteString(StatusBarStyle().Width(m.width).Render(m.getStatusBar()))
		return rb.String()
	}

	// Get responsive column widths based on terminal width
	nameW, statusW, versionW := EKSColumnWidths(m.width)

	// Table header - clean and aligned with responsive widths
	headerRow := fmt.Sprintf("  %-*s %-*s %-*s", nameW, "NAME", statusW, "STATUS", versionW, "VERSION")
	rb.WriteLine(TableHeaderStyle().Render(headerRow))

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

		rb.WriteLine(RenderSelectableRow(row, i == cursor))
	}

	// Pagination indicator
	if len(clusters) > endIdx-startIdx {
		rb.Newline()
		rb.WriteString(SubtitleStyle().Render(fmt.Sprintf("Showing %d-%d of %d", startIdx+1, endIdx, len(clusters))))
	}

	// Footer
	rb.Newline()
	if searchBar := m.renderSearchBar(ViewEKSClusters); searchBar != "" {
		rb.WriteLine(searchBar)
	}
	rb.WriteString(m.renderEKSFooter())

	// Status bar
	rb.Newline()
	rb.WriteString(StatusBarStyle().Width(m.width).Render(m.getStatusBar()))

	return rb.String()
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
