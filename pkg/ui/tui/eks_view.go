package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
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

	// No clusters
	if len(clusters) == 0 {
		b.WriteString(SubtitleStyle.Render("No EKS clusters found"))
		b.WriteString("\n\n")
		b.WriteString(HelpStyle.Render("esc:back"))
		b.WriteString("\n")
		b.WriteString(StatusBarStyle.Render(m.getStatusBar()))
		return b.String()
	}

	// Table header - clean and aligned
	headerRow := fmt.Sprintf("  %-45s %-15s %-10s", "NAME", "STATUS", "VERSION")
	b.WriteString(TableHeaderStyle.Render(headerRow))
	b.WriteString("\n")

	// Render clusters with proper alignment
	for i, cluster := range clusters {
		// Truncate name if too long
		name := cluster.Name
		if len(name) > 45 {
			name = name[:42] + "..."
		}

		status := StateStyle(cluster.Status)
		row := fmt.Sprintf("  %-45s %-15s %-10s", name, status, cluster.Version)

		b.WriteString(RenderSelectableRow(row, i == m.cursor))
		b.WriteString("\n")
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
	b.WriteString(StatusBarStyle.Width(m.width).Render(m.getStatusBar()))

	return b.String()
}

// handleEKSKeys handles keyboard input for EKS clusters view
func (m Model) handleEKSKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	clusters := m.getEKSClusters()

	switch msg.String() {
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}

	case "down", "j":
		if m.cursor < len(clusters)-1 {
			m.cursor++
		}

	case "g":
		if len(clusters) > 0 {
			m.cursor = 0
		}

	case "G":
		if len(clusters) > 0 {
			m.cursor = len(clusters) - 1
		}

	case "enter", " ":
		// Show cluster details (exit TUI and display cluster info)
		if m.cursor < len(clusters) {
			cluster := clusters[m.cursor]

			// Schedule cluster details display after TUI exits
			return m, m.scheduleClusterDetails(cluster.Name)
		}

	case "r":
		// Refresh clusters
		m.loading = true
		m.loadingMsg = "Refreshing EKS clusters..."
		m.err = nil
		return m, LoadEKSClustersCmd(m.ctx, m.client)

	case "esc":
		return m.navigateBack(), nil
	}

	return m, nil
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
		keyStyle := StatusBarKeyStyle.Render(k.key)
		descStyle := StatusBarValueStyle.Render(k.desc)
		parts = append(parts, fmt.Sprintf("%s %s", keyStyle, descStyle))
	}

	return HelpStyle.Render(strings.Join(parts, " • "))
}
