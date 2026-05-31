// Package sidebar renders the left navigation rail.
package sidebar

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Visual tokens. The sidebar leans on a single accent color for the
// selected entry's left bar; the rest is muted so the user's eye is
// drawn to the focused row, not to icon clutter.
var (
	colorAccent      = lipgloss.Color("#58A6FF")
	colorPrimary     = lipgloss.Color("#E6EDF3")
	colorMutedText   = lipgloss.Color("#8B949E")
	colorBorder      = lipgloss.Color("#30363D")
	colorBorderFocus = lipgloss.Color("#388BFD")
)

// Item is one entry in the sidebar.
type Item struct {
	Label string
	Count int  // -1 means hide count (e.g. Home, Help)
	Focus bool // currently selected
}

// Render returns a bordered panel of exactly width × height cells.
// The panel includes a rounded border on all four sides plus a "Views"
// title in the top border. Content area is (width-2) × (height-2).
func Render(width, height int, items []Item) string {
	if width < 4 || height < 4 {
		return ""
	}

	// Border eats 1 cell on each side. The body is sized to fit
	// exactly inside the border (width-2 cells × height-2 lines).
	// Item rendering bakes in 1 cell of inset on each side itself
	// so we don't fight lipgloss's Padding/Width interaction.
	contentWidth := width - 2
	contentHeight := height - 2

	if contentWidth < 4 {
		return ""
	}

	lines := make([]string, 0, contentHeight)
	for i, item := range items {
		if i >= contentHeight {
			break
		}
		lines = append(lines, renderItem(item, contentWidth))
	}
	for len(lines) < contentHeight {
		lines = append(lines, strings.Repeat(" ", contentWidth))
	}
	body := strings.Join(lines, "\n")

	panel := lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(colorBorder).
		Width(contentWidth).
		Height(contentHeight)

	return panel.Render(body)
}

func renderItem(item Item, contentWidth int) string {
	// Layout (contentWidth ≈ 18 typical):
	//   " ▎ <label>           <count> "
	// 1 leading space, 1 indicator, 1 space, label fills, count
	// right-aligned with 1 trailing space.

	indicator := " "
	indicatorStyle := lipgloss.NewStyle().Foreground(colorMutedText)
	labelStyle := lipgloss.NewStyle().Foreground(colorMutedText)

	if item.Focus {
		indicator = "▎"
		indicatorStyle = lipgloss.NewStyle().Foreground(colorAccent).Bold(true)
		labelStyle = lipgloss.NewStyle().Foreground(colorPrimary).Bold(true)
	}

	countStr := ""
	if item.Count >= 0 {
		countStr = fmt.Sprintf("%d", item.Count)
	}
	countStyle := lipgloss.NewStyle().Foreground(colorMutedText)

	// Compose: " <indicator> <label>...<count> "
	const leadPad = 1
	const tailPad = 1
	gap := contentWidth - leadPad - 1 - 1 - lipgloss.Width(item.Label) - lipgloss.Width(countStr) - tailPad
	if gap < 1 {
		gap = 1
	}

	return strings.Repeat(" ", leadPad) +
		indicatorStyle.Render(indicator) + " " +
		labelStyle.Render(item.Label) +
		strings.Repeat(" ", gap) +
		countStyle.Render(countStr) +
		strings.Repeat(" ", tailPad)
}

// FocusedBorder returns a border foreground color suitable for a
// panel that should be visually distinguished as focused. Exported so
// the parent package can apply matching color treatments to the
// active region.
func FocusedBorder() lipgloss.Color { return colorBorderFocus }
