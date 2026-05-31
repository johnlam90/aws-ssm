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
	colorPrimary     = lipgloss.Color("#FFFFFF")
	colorSecondary   = lipgloss.Color("#C9D1D9")
	colorMuted       = lipgloss.Color("#484F58")
	colorHighlightBg = lipgloss.Color("#264F78")
)

// Item is one entry in the sidebar.
type Item struct {
	Icon  string // single rune / glyph; may be empty
	Label string
	Count int  // -1 means hide count (e.g. Home, Help)
	Focus bool // whether this entry is highlighted
}

// Render returns a vertical block of the requested width × height.
// The rightmost column is reserved for a thin vertical separator
// (` │ `) so the sidebar visually demarcates from the main panel.
// Width must equal layout.SidebarFullWidth (14) for full mode;
// smaller widths simply pad with spaces. The returned string contains
// exactly height lines separated by newlines; each line is exactly
// width cells wide.
func Render(width, height int, items []Item) string {
	if width <= 0 || height <= 0 {
		return ""
	}

	// Reserve the rightmost cell for a vertical separator.
	itemWidth := width - 2
	if itemWidth < 1 {
		itemWidth = 1
	}

	separator := lipgloss.NewStyle().Foreground(colorMuted).Render(" │")
	lines := make([]string, 0, height)
	for i, item := range items {
		if i >= height {
			break
		}
		lines = append(lines, renderItem(item, itemWidth)+separator)
	}

	for len(lines) < height {
		lines = append(lines, padRight("", itemWidth)+separator)
	}

	return strings.Join(lines, "\n")
}

func renderItem(item Item, width int) string {
	icon := item.Icon
	if icon == "" {
		icon = " "
	}

	countStr := ""
	if item.Count >= 0 {
		countStr = fmt.Sprintf("%d", item.Count)
	}

	// Layout (full mode, width=14):
	//   " ┃ <icon> <label>     <count> "
	// indicator(1) + space(1) + icon(1) + space(1) + label(N) + space + count(M) + trailing space
	const fixed = 4                // indicator + space + icon + space
	available := width - fixed - 1 // -1 trailing space
	if available < 0 {
		available = 0
	}

	labelW := available - lipgloss.Width(countStr) - 1
	if countStr == "" {
		labelW = available
	}
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

	// Build the body without per-segment styles, then apply a single
	// row-wide style. This eliminates the visual seam between the
	// indicator and the rest of the row when the entry is focused.
	indicatorChar := " "
	if item.Focus {
		indicatorChar = "┃"
	}

	body := indicatorChar + " " + icon + " " + label + strings.Repeat(" ", pad)
	if countStr != "" {
		body += " " + countStr
	}
	body = padRight(body, width)

	if item.Focus {
		return lipgloss.NewStyle().
			Foreground(colorPrimary).
			Background(colorHighlightBg).
			Bold(true).
			Render(body)
	}

	return lipgloss.NewStyle().
		Foreground(colorSecondary).
		Render(body)
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
