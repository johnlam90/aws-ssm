package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// renderHelp renders the help main-panel content for the legacy
// ViewHelp screen. The Phase 10 help overlay is preferred and
// reachable via `?` from anywhere; this kept-around screen renders
// when ViewHelp is reached via the sidebar's Help entry.
func (m Model) renderHelp() string {
	var b strings.Builder

	b.WriteString(TitleStyle().Render("Help & Keybindings"))
	b.WriteString("\n")
	b.WriteString(SubtitleStyle().Render("Navigate and manage AWS resources"))
	b.WriteString("\n\n")

	b.WriteString(GetQuickReference())
	b.WriteString("\n")

	b.WriteString(TitleStyle().Render("Global Keybindings"))
	b.WriteString("\n\n")
	b.WriteString(FormatKeyBindings(globalKeyBindings))
	b.WriteString("\n")

	b.WriteString(SubtitleStyle().Render("Press ? from any view to open the help overlay."))

	return b.String()
}

// renderHelpOverlay renders a centered Help overlay that can be
// toggled from any view via `?`. Phase 10 of the foundation redesign.
func (m Model) renderHelpOverlay() string {
	titleView := m.currentView.String()
	keyStyle := lipgloss.NewStyle().Foreground(ColorAccentIndigo).Bold(true)
	descStyle := lipgloss.NewStyle().Foreground(ColorSecondary)
	sectionStyle := lipgloss.NewStyle().Foreground(ColorAccentBlue).Bold(true).Underline(true)

	var b strings.Builder
	b.WriteString(sectionStyle.Render(fmt.Sprintf("Help · %s", titleView)))
	b.WriteString("\n\n")

	groups := []struct {
		title string
		keys  []struct{ key, desc string }
	}{
		{
			"Navigation",
			[]struct{ key, desc string }{
				{"j  k  ↑  ↓", "move cursor"},
				{"g  G", "top / bottom"},
				{"Ctrl+D / Ctrl+U", "page down / up"},
				{"1  2  3  4  5", "jump to view by index"},
				{"esc", "back / close overlay"},
			},
		},
		{
			"This view",
			helpHintsForView(m.currentView),
		},
		{
			"Global",
			[]struct{ key, desc string }{
				{"/", "filter / search"},
				{":", "open command palette"},
				{"r", "refresh"},
				{"y", "copy focused resource ID"},
				{"?", "toggle this help"},
				{"q", "quit"},
			},
		},
	}

	for _, g := range groups {
		b.WriteString(sectionStyle.Render(g.title))
		b.WriteString("\n")
		for _, k := range g.keys {
			fmt.Fprintf(&b, "  %s  %s\n",
				keyStyle.Width(18).Render(k.key),
				descStyle.Render(k.desc),
			)
		}
		b.WriteString("\n")
	}

	b.WriteString(descStyle.Italic(true).Render("Press ? or esc to close"))
	return b.String()
}

func helpHintsForView(v ViewMode) []struct{ key, desc string } {
	switch v {
	case ViewEC2Instances:
		return []struct{ key, desc string }{
			{"↵", "connect via SSM"},
			{"y", "yank instance ID"},
			{"/", "filter (state:, type:, subnet:, sg:, etc.)"},
		}
	case ViewEKSClusters:
		return []struct{ key, desc string }{
			{"↵", "open node groups"},
			{"y", "yank cluster name"},
		}
	case ViewASGs:
		return []struct{ key, desc string }{
			{"↵ / s", "scale ASG"},
			{"y", "yank ASG name"},
		}
	case ViewNodeGroups:
		return []struct{ key, desc string }{
			{"↵ / s", "scale node group"},
			{"u", "update launch template"},
			{"y", "yank node group name"},
		}
	default:
		return []struct{ key, desc string }{
			{"↵", "select / activate"},
		}
	}
}
