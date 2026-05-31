// Package sidebar renders the left navigation rail. The package is
// agnostic to the parent tui package's view enum — callers pass a
// slice of Item values plus an index of which item is currently
// selected.
package sidebar

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Color tokens duplicated locally; see chrome package for rationale.
var (
	colorAccent      = lipgloss.Color("#58A6FF")
	colorPrimary     = lipgloss.Color("#E6EDF3")
	colorSecondary   = lipgloss.Color("#C9D1D9")
	colorMuted       = lipgloss.Color("#8B949E")
	colorBackground  = lipgloss.Color("#161B22")
	colorHighlightBg = lipgloss.Color("#1F2A3D")
)

// Item is one entry in the sidebar.
type Item struct {
	Icon  string // single rune / glyph; may be empty
	Label string
	Count int  // -1 means hide count (e.g. Home, Help)
	Focus bool // whether this entry is highlighted
}

// Render returns a vertical block of the requested width × height.
// Width must equal layout.SidebarFullWidth (14) for Phase 2; smaller
// widths simply pad with spaces. The returned string contains exactly
// height lines separated by newlines; each line is exactly width
// cells wide.
func Render(width, height int, items []Item) string {
	if width <= 0 || height <= 0 {
		return ""
	}

	lines := make([]string, 0, height)
	for i, item := range items {
		if i >= height {
			break
		}
		lines = append(lines, renderItem(item, width))
	}

	for len(lines) < height {
		lines = append(lines, padRight("", width))
	}

	return strings.Join(lines, "\n")
}

func renderItem(item Item, width int) string {
	indicator := " "
	labelStyle := lipgloss.NewStyle().Foreground(colorSecondary)
	countStyle := lipgloss.NewStyle().Foreground(colorMuted)

	if item.Focus {
		indicator = lipgloss.NewStyle().Foreground(colorAccent).Bold(true).Render("┃")
		labelStyle = lipgloss.NewStyle().Foreground(colorPrimary).Background(colorHighlightBg).Bold(true)
		countStyle = lipgloss.NewStyle().Foreground(colorPrimary).Background(colorHighlightBg)
	} else {
		_ = colorBackground // reserved for future inactive-rail backdrop
	}

	icon := item.Icon
	if icon == "" {
		icon = " "
	}

	countStr := ""
	if item.Count >= 0 {
		countStr = fmt.Sprintf("%d", item.Count)
	}

	// Layout: " ┃ <icon> <label>     <count> "
	// width is typically 14. Indicator (1) + space (1) + icon (1) +
	// space (1) + label (variable) + count (right-aligned) + trailing space.
	const fixed = 5 // indicator + 2 spaces + icon + space
	available := width - fixed
	if available < 0 {
		available = 0
	}

	labelW := available - lipgloss.Width(countStr) - 1
	if labelW < 0 {
		labelW = 0
	}

	label := item.Label
	if lipgloss.Width(label) > labelW {
		label = truncateRunes(label, labelW)
	}

	pad := labelW - lipgloss.Width(label)
	if pad < 0 {
		pad = 0
	}

	body := indicator + " " + icon + " " + labelStyle.Render(label) + strings.Repeat(" ", pad)
	if countStr != "" {
		body += " " + countStyle.Render(countStr)
	} else if available > 0 {
		body += strings.Repeat(" ", lipgloss.Width(countStr)+1)
	}

	return padRight(body, width)
}

func padRight(s string, width int) string {
	w := lipgloss.Width(s)
	if w >= width {
		if w == width {
			return s
		}
		return truncateRunes(s, width)
	}
	return s + strings.Repeat(" ", width-w)
}

func truncateRunes(s string, width int) string {
	runes := []rune(s)
	for len(runes) > 0 && lipgloss.Width(string(runes)) > width {
		runes = runes[:len(runes)-1]
	}
	return string(runes)
}
