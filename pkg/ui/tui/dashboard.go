package tui

import (
	"fmt"
	"sort"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// renderDashboard renders the Home view — a clean welcome screen
// with current AWS context, brief usage tips, and any per-resource
// counts that have been loaded this session. It deliberately does
// NOT duplicate the resource-type list (which appeared in the prior
// dashboard); navigation now goes through the command palette and
// Mod+1..N hotkeys.
func (m Model) renderDashboard() string {
	if m.loading {
		return m.renderLoading()
	}
	if m.err != nil {
		return m.renderError()
	}

	var b strings.Builder

	titleStyle := lipgloss.NewStyle().Foreground(ColorAccentBlue).Bold(true)
	primaryStyle := lipgloss.NewStyle().Foreground(ColorPrimary).Bold(true)
	secondaryStyle := lipgloss.NewStyle().Foreground(ColorSecondary)
	mutedStyle := lipgloss.NewStyle().Foreground(ColorMuted).Italic(true)
	keyStyle := lipgloss.NewStyle().Foreground(ColorAccentIndigo).Bold(true)
	countStyle := lipgloss.NewStyle().Foreground(ColorAccentGreen).Bold(true)

	// Welcome line — large brand + version-ish subtitle.
	b.WriteString(titleStyle.Render("aws-ssm"))
	b.WriteString(secondaryStyle.Render("  ·  AWS Systems Manager TUI"))
	b.WriteString("\n\n")

	// Context block.
	b.WriteString(secondaryStyle.Render("Region   "))
	b.WriteString(primaryStyle.Render(m.getRegion()))
	b.WriteString("\n")
	b.WriteString(secondaryStyle.Render("Profile  "))
	b.WriteString(primaryStyle.Render(m.getProfile()))
	b.WriteString("\n\n")

	// Loaded-resource summary, only if anything is actually loaded.
	if m.anythingLoaded() {
		b.WriteString(secondaryStyle.Render("Loaded this session"))
		b.WriteString("\n")
		summary := []struct {
			label string
			count int
			state string
		}{
			{"EC2 Instances", len(m.ec2Instances), m.ec2StateSummary()},
			{"EKS Clusters", len(m.eksClusters), m.eksStateSummary()},
			{"Auto Scaling Groups", len(m.asgs), m.asgStateSummary()},
			{"EKS Node Groups", len(m.nodeGroups), m.nodeGroupStateSummary()},
		}
		for _, s := range summary {
			if s.count == 0 && s.state == "" {
				continue
			}
			fmt.Fprintf(&b, "  %s   %s",
				secondaryStyle.Width(22).Render(s.label),
				countStyle.Render(fmt.Sprintf("%d", s.count)),
			)
			if s.state != "" {
				fmt.Fprintf(&b, "   %s", mutedStyle.Render(s.state))
			}
			b.WriteString("\n")
		}
		b.WriteString("\n")
	}

	// Tips. Single section, terse.
	b.WriteString(secondaryStyle.Render("Get started"))
	b.WriteString("\n")
	tips := []struct {
		key, desc string
	}{
		{":", "open the command palette · try :ec2, :eks, :asg, :ng"},
		{"/", "filter the focused list"},
		{"y", "copy the focused resource ID to clipboard"},
		{"r", "refresh the current view"},
		{"?", "show keybindings overlay"},
	}
	for _, t := range tips {
		fmt.Fprintf(&b, "  %s   %s\n", keyStyle.Render(t.key), secondaryStyle.Render(t.desc))
	}

	return b.String()
}

func (m Model) anythingLoaded() bool {
	return len(m.ec2Instances)+len(m.eksClusters)+len(m.asgs)+len(m.nodeGroups) > 0
}

// normalizeServiceDescription preserved for legacy tests; the Home
// view's roll-up is gone.
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

// renderLoading renders an inline loading message.
func (m Model) renderLoading() string {
	if !m.loading {
		return ""
	}
	frame := "⠋"
	return LoadingStyle().Render(fmt.Sprintf("%s %s", frame, m.loadingMsg))
}

// renderError renders an inline error message.
func (m Model) renderError() string {
	if m.err == nil {
		return ""
	}
	return ErrorStyle().Render(fmt.Sprintf("Error: %v", m.err))
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
