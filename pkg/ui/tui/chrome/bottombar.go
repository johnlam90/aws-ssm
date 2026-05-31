package chrome

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Hint is one key/label pair shown on the bottom hint bar.
type Hint struct {
	Key   string
	Label string
}

// BottomBarInput holds the values rendered in the bottom hint bar.
type BottomBarInput struct {
	Hints  []Hint
	Status string
	Width  int
}

// RenderBottomBar returns a two-line string sized to width × 2.
// Line 1 is the key hints; line 2 is the status footer.
func RenderBottomBar(in BottomBarInput) string {
	if in.Width <= 0 {
		return ""
	}

	hintsLine := renderHintsLine(in.Hints, in.Width)
	statusLine := renderStatusLine(in.Status, in.Width)

	return hintsLine + "\n" + statusLine
}

func renderHintsLine(hints []Hint, width int) string {
	if len(hints) == 0 {
		return strings.Repeat(" ", width)
	}

	keyStyle := lipgloss.NewStyle().Foreground(colorBrand).Bold(true)
	labelStyle := lipgloss.NewStyle().Foreground(colorSecondary)
	sepStyle := lipgloss.NewStyle().Foreground(colorMuted)

	parts := make([]string, 0, len(hints))
	for _, h := range hints {
		parts = append(parts, keyStyle.Render(h.Key)+" "+labelStyle.Render(h.Label))
	}
	separator := sepStyle.Render("  ·  ")
	joined := strings.Join(parts, separator)

	return padOrTruncateRight(joined, width)
}

func renderStatusLine(status string, width int) string {
	if status == "" {
		return strings.Repeat(" ", width)
	}
	statusStyle := lipgloss.NewStyle().Foreground(colorMuted)
	rendered := statusStyle.Render(status)
	return padOrTruncateRight(rendered, width)
}

func padOrTruncateRight(s string, width int) string {
	w := lipgloss.Width(s)
	if w == width {
		return s
	}
	if w < width {
		return s + strings.Repeat(" ", width-w)
	}
	return truncateToWidth(s, width)
}
