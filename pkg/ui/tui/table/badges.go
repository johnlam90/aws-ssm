// Package table renders adaptive-width tables for resource lists.
// It also owns state-glyph rendering — the small unicode dots and
// rings used to communicate resource health at a glance.
package table

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// State glyphs. Glyph-only style per the foundation design — full
// state text appears in the detail panel, not the table cell.
const (
	GlyphRunning    = "●"
	GlyphPending    = "◐"
	GlyphStopped    = "○"
	GlyphTerminated = "✕"
	GlyphUnknown    = "·"
)

var (
	colorRunning    = lipgloss.Color("#3FB950")
	colorPending    = lipgloss.Color("#D29922")
	colorStopped    = lipgloss.Color("#8B949E")
	colorTerminated = lipgloss.Color("#F85149")
	colorUnknown    = lipgloss.Color("#484F58")
)

// StateBadge returns a single styled glyph for the given state. The
// state string is normalized (trimmed, lowercased) before matching.
// The returned string contains exactly one visible cell.
func StateBadge(state string) string {
	g, c := glyphAndColor(state)
	return lipgloss.NewStyle().Foreground(c).Render(g)
}

// StateBadgeOK returns true if the state corresponds to a healthy
// running resource. Useful for sorting / filtering callers.
func StateBadgeOK(state string) bool {
	g, _ := glyphAndColor(state)
	return g == GlyphRunning
}

func glyphAndColor(state string) (string, lipgloss.Color) {
	switch strings.ToLower(strings.TrimSpace(state)) {
	case "running", "active", "healthy", "ready":
		return GlyphRunning, colorRunning
	case "pending", "creating", "updating", "scaling", "scaling up", "scaling down", "in_progress":
		return GlyphPending, colorPending
	case "stopped", "stopping", "inactive", "idle":
		return GlyphStopped, colorStopped
	case "terminated", "shutting-down", "failed", "deleting", "deleted", "error":
		return GlyphTerminated, colorTerminated
	case "":
		return GlyphUnknown, colorUnknown
	default:
		return GlyphUnknown, colorUnknown
	}
}
