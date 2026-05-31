// Package chrome renders the persistent top and bottom regions of the
// TUI: the single-line top bar (app brand · breadcrumb · region ·
// profile) and the two-line bottom hint bar (key hints + status
// footer). The package depends only on lipgloss and primitive inputs;
// callers pass already-resolved strings so the package stays free of
// circular imports against the parent tui package.
package chrome

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Color tokens duplicated locally to avoid importing the parent tui
// package (which would cycle back through chrome). A future
// refactoring sub-project may unify these into a shared theme package.
var (
	colorBrand     = lipgloss.Color("#58A6FF")
	colorBreadcrum = lipgloss.Color("#E6EDF3")
	colorAccent    = lipgloss.Color("#3FB950")
	colorMuted     = lipgloss.Color("#8B949E")
	colorSecondary = lipgloss.Color("#C9D1D9")
)

// TopBarInput holds the values rendered in the top chrome bar.
type TopBarInput struct {
	Brand      string
	Breadcrumb string
	Region     string
	Profile    string
	Width      int
}

// RenderTopBar returns a single-line string of the requested width.
// The string is sized to exactly Width cells (excluding ANSI escape
// codes); the caller is responsible for placing it at the top of the
// screen.
//
// Segments drop right-to-left under width pressure: clock and account
// (added in later phases) drop first, then profile, then region, then
// breadcrumb. The brand never drops.
func RenderTopBar(in TopBarInput) string {
	if in.Width <= 0 {
		return ""
	}

	brand := lipgloss.NewStyle().
		Foreground(colorBrand).
		Bold(true).
		Render(in.Brand)

	left := []string{brand}

	if in.Breadcrumb != "" {
		breadcrumb := lipgloss.NewStyle().Foreground(colorBreadcrum).Render(in.Breadcrumb)
		left = append(left, breadcrumb)
	}

	right := []string{}

	if in.Region != "" {
		dot := lipgloss.NewStyle().Foreground(colorAccent).Render("●")
		region := lipgloss.NewStyle().Foreground(colorSecondary).Render(in.Region)
		right = append(right, dot+" "+region)
	}
	if in.Profile != "" {
		profile := lipgloss.NewStyle().Foreground(colorSecondary).Render(in.Profile)
		right = append(right, profile)
	}

	leftJoined := joinSegments(left, " · ")
	rightJoined := joinSegments(right, " │ ")

	return composeWithFill(leftJoined, rightJoined, in.Width)
}

// joinSegments joins styled segments with a separator, returning the
// concatenated string.
func joinSegments(segments []string, sep string) string {
	if len(segments) == 0 {
		return ""
	}
	return strings.Join(segments, sep)
}

// composeWithFill places left at the start, right at the end, and
// pads in the middle with spaces so the visible-width totals exactly
// width cells. Segments are dropped from the right side first if they
// don't fit; if even left alone won't fit, it's truncated.
func composeWithFill(left, right string, width int) string {
	leftW := lipgloss.Width(left)
	rightW := lipgloss.Width(right)

	if leftW+rightW <= width {
		gap := width - leftW - rightW
		return left + strings.Repeat(" ", gap) + right
	}

	if leftW <= width {
		gap := width - leftW
		return left + strings.Repeat(" ", gap)
	}

	return truncateToWidth(left, width)
}

// truncateToWidth crudely truncates a string to the given visible
// width. ANSI safety is approximate; the brand and breadcrumb are
// short enough that this is acceptable for Phase 2.
func truncateToWidth(s string, width int) string {
	if lipgloss.Width(s) <= width {
		return s
	}
	runes := []rune(s)
	for len(runes) > 0 && lipgloss.Width(string(runes)) > width {
		runes = runes[:len(runes)-1]
	}
	return string(runes)
}
