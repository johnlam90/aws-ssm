package table

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// ColumnSpec describes one column in an adaptive table. Each column
// declares a minimum width, a preferred width, and a maximum width.
// Allocate distributes the available row width across columns: every
// column gets at least its MinWidth (or is dropped if even that won't
// fit), then slack is distributed toward each column's PrefWidth, and
// any remaining slack is bounded by MaxWidth.
type ColumnSpec struct {
	Header    string
	MinWidth  int
	PrefWidth int
	MaxWidth  int
	// Align is "left" (default), "right", or "center".
	Align string
}

// AllocatedColumn is the runtime resolution of a ColumnSpec for a
// given total row width. Width is the number of cells the column
// occupies, including padding. If Dropped is true the column is not
// rendered (typically because the available row width was below the
// summed MinWidths).
type AllocatedColumn struct {
	Spec    ColumnSpec
	Width   int
	Dropped bool
}

// Allocate distributes totalWidth across the supplied column specs.
// It accounts for one space of padding between each pair of columns.
func Allocate(specs []ColumnSpec, totalWidth int) []AllocatedColumn {
	n := len(specs)
	out := make([]AllocatedColumn, n)
	for i, s := range specs {
		out[i] = AllocatedColumn{Spec: s, Width: s.MinWidth}
	}

	if n == 0 || totalWidth <= 0 {
		return out
	}

	gap := n - 1 // one space between adjacent columns
	usable := totalWidth - gap

	// Drop columns from the right if even MinWidths won't fit.
	required := 0
	for _, c := range out {
		required += c.Width
	}
	for required > usable && n > 0 {
		out[n-1].Dropped = true
		required -= out[n-1].Width
		out[n-1].Width = 0
		n--
		gap = n - 1
		if gap < 0 {
			gap = 0
		}
		usable = totalWidth - gap
	}

	// Distribute slack toward PrefWidth.
	slack := usable - required
	if slack > 0 {
		for i := 0; i < n && slack > 0; i++ {
			want := out[i].Spec.PrefWidth - out[i].Width
			if want <= 0 {
				continue
			}
			grant := want
			if grant > slack {
				grant = slack
			}
			out[i].Width += grant
			slack -= grant
		}
	}

	// Distribute remaining slack up to MaxWidth.
	if slack > 0 {
		for slack > 0 {
			progressed := false
			for i := 0; i < n && slack > 0; i++ {
				if out[i].Spec.MaxWidth <= 0 {
					out[i].Width++
					slack--
					progressed = true
					continue
				}
				if out[i].Width < out[i].Spec.MaxWidth {
					out[i].Width++
					slack--
					progressed = true
				}
			}
			if !progressed {
				break
			}
		}
	}

	return out
}

// FitCell formats a value into exactly width cells using the given
// alignment, truncating with `…` when needed. Width is measured in
// terminal cells (lipgloss-aware).
func FitCell(value string, width int, align string) string {
	if width <= 0 {
		return ""
	}
	w := lipgloss.Width(value)
	if w == width {
		return value
	}
	if w > width {
		return truncateWithEllipsis(value, width)
	}
	pad := width - w
	switch align {
	case "right":
		return strings.Repeat(" ", pad) + value
	case "center":
		left := pad / 2
		return strings.Repeat(" ", left) + value + strings.Repeat(" ", pad-left)
	default:
		return value + strings.Repeat(" ", pad)
	}
}

func truncateWithEllipsis(s string, width int) string {
	if width <= 0 {
		return ""
	}
	if width == 1 {
		return "…"
	}
	runes := []rune(s)
	for len(runes) > 0 && lipgloss.Width(string(runes)+"…") > width {
		runes = runes[:len(runes)-1]
	}
	return string(runes) + "…"
}
