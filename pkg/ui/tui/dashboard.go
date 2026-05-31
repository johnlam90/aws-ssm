package tui

import (
	"fmt"
	"sort"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// renderDashboard renders the Home view — a summary roll-up of the
// current AWS context plus per-resource counts. Phase 9 of the
// foundation redesign replaces the previous menu-style dashboard
// (sidebar handles navigation now; the home becomes a real
// dashboard).
func (m Model) renderDashboard() string {
	var b strings.Builder

	if m.loading {
		b.WriteString(m.renderLoading())
		return b.String()
	}
	if m.err != nil {
		b.WriteString(m.renderError())
		return b.String()
	}

	contextStyle := SubtitleStyle()
	b.WriteString(contextStyle.Render(fmt.Sprintf(
		"Region %s   ·   Profile %s",
		m.getRegion(), m.getProfile(),
	)))
	b.WriteString("\n\n")

	b.WriteString(DashboardSectionTitleStyle().Render("Resources in this region"))
	b.WriteString("\n\n")

	rollups := []struct {
		icon  string
		label string
		count int
		view  ViewMode
		state string
	}{
		{"▣", "EC2 Instances", len(m.ec2Instances), ViewEC2Instances, m.ec2StateSummary()},
		{"☸", "EKS Clusters", len(m.eksClusters), ViewEKSClusters, m.eksStateSummary()},
		{"⚖", "Auto Scaling Groups", len(m.asgs), ViewASGs, m.asgStateSummary()},
		{"⛁", "EKS Node Groups", len(m.nodeGroups), ViewNodeGroups, m.nodeGroupStateSummary()},
	}

	iconStyle := lipgloss.NewStyle().Foreground(ColorAccentBlue).Bold(true)
	labelStyle := lipgloss.NewStyle().Foreground(ColorPrimary).Bold(true)
	dimStyle := lipgloss.NewStyle().Foreground(ColorSecondary)
	countStyle := lipgloss.NewStyle().Foreground(ColorAccentIndigo).Bold(true)

	for i, r := range rollups {
		marker := "  "
		if i == m.cursor {
			marker = DashboardSelectionBarStyle().Render("▌ ")
		}
		count := dimStyle.Render("(not loaded)")
		if r.count > 0 || r.state != "" {
			count = fmt.Sprintf("%s   %s", countStyle.Render(fmt.Sprintf("%d", r.count)), dimStyle.Render(r.state))
		} else if r.count == 0 {
			count = dimStyle.Render("0")
		}
		fmt.Fprintf(&b, "%s%s  %s   %s\n",
			marker,
			iconStyle.Render(r.icon),
			labelStyle.Width(22).Render(r.label),
			count,
		)
	}

	b.WriteString("\n")
	b.WriteString(SubtitleStyle().Render("Quick start"))
	b.WriteString("\n")
	hintKey := lipgloss.NewStyle().Foreground(ColorAccentIndigo).Bold(true)
	hintDesc := lipgloss.NewStyle().Foreground(ColorSecondary)
	for _, h := range []struct{ key, desc string }{
		{":", "open command palette"},
		{"/", "filter the focused list"},
		{"↵", "select / open the highlighted resource"},
		{"r", "refresh"},
		{"?", "help"},
	} {
		fmt.Fprintf(&b, "  %s  %s\n", hintKey.Render(h.key), hintDesc.Render(h.desc))
	}

	return b.String()
}

// State summaries — short human-readable strings showing per-state
// breakdowns. Empty when no data is loaded.

func (m Model) ec2StateSummary() string {
	if len(m.ec2Instances) == 0 {
		return ""
	}
	counts := map[string]int{}
	for _, inst := range m.ec2Instances {
		counts[strings.ToLower(strings.TrimSpace(inst.State))]++
	}
	return formatStateCounts(counts)
}

func (m Model) eksStateSummary() string {
	if len(m.eksClusters) == 0 {
		return ""
	}
	counts := map[string]int{}
	for _, c := range m.eksClusters {
		counts[strings.ToLower(strings.TrimSpace(c.Status))]++
	}
	return formatStateCounts(counts)
}

func (m Model) asgStateSummary() string {
	if len(m.asgs) == 0 {
		return ""
	}
	counts := map[string]int{}
	for _, a := range m.asgs {
		counts[strings.ToLower(strings.TrimSpace(a.Status))]++
	}
	return formatStateCounts(counts)
}

func (m Model) nodeGroupStateSummary() string {
	if len(m.nodeGroups) == 0 {
		return ""
	}
	counts := map[string]int{}
	for _, ng := range m.nodeGroups {
		counts[strings.ToLower(strings.TrimSpace(ng.Status))]++
	}
	return formatStateCounts(counts)
}

func formatStateCounts(counts map[string]int) string {
	if len(counts) == 0 {
		return ""
	}
	keys := make([]string, 0, len(counts))
	for k := range counts {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	parts := make([]string, 0, len(keys))
	for _, k := range keys {
		display := k
		if display == "" {
			display = "unknown"
		}
		parts = append(parts, fmt.Sprintf("%d %s", counts[k], display))
	}
	return strings.Join(parts, " · ")
}

// renderLoading renders an inline loading message used by per-view
// renderers when their data is still loading.
func (m Model) renderLoading() string {
	if !m.loading {
		return ""
	}
	frame := "⠋"
	msg := fmt.Sprintf("%s %s", frame, m.loadingMsg)
	return LoadingStyle().Render(msg)
}

// renderError renders an inline error message.
func (m Model) renderError() string {
	if m.err == nil {
		return ""
	}
	return ErrorStyle().Render(fmt.Sprintf("Error: %v", m.err))
}

// normalizeServiceDescription normalizes service descriptions for consistency.
// Retained as a small text helper used by tests; the Home view's
// roll-up renderer no longer needs descriptions inline.
func (m Model) normalizeServiceDescription(desc string) string {
	normalizedMap := map[string]string{
		"View and manage EC2 instances":        "Manage EC2 instances",
		"Manage EKS clusters and node groups":  "Manage EKS clusters & node groups",
		"View and scale ASGs":                  "Scale and monitor ASGs",
		"Inspect managed node groups":          "Inspect managed node groups",
		"View EC2 network interfaces and ENIs": "View EC2 ENIs",
		"View keybindings and help":            "Keybindings and documentation",
	}

	if normalized, exists := normalizedMap[desc]; exists {
		return normalized
	}
	return desc
}

// getRegion returns the current AWS region.
func (m Model) getRegion() string {
	if m.config.Region != "" {
		return m.config.Region
	}
	return "us-east-1"
}

// getProfile returns the current AWS profile.
func (m Model) getProfile() string {
	if m.config.Profile != "" {
		return m.config.Profile
	}
	return "default"
}
