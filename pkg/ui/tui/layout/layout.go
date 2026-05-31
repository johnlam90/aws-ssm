// Package layout owns region dimension arithmetic for the TUI.
//
// Compute returns a Layout describing the four screen regions (top bar,
// sidebar, main, bottom bar) for a given terminal size. Each region is
// expressed as a Rect; consumers render their content into those Rects.
package layout

// Chrome and sidebar dimensions. Exported so chrome/sidebar packages
// can import them and stay in sync with the layout's reservations.
const (
	// TopBarHeight is the vertical reservation for the top chrome bar.
	TopBarHeight = 1
	// BottomBarHeight is the vertical reservation for the bottom hint
	// bar (line 1: key hints; line 2: status footer).
	BottomBarHeight = 2
	// SidebarFullWidth is the column reservation for the full sidebar
	// (including its rounded panel border on each side).
	SidebarFullWidth = 18
	// SidebarMinTerminalWidth is the minimum terminal width at which
	// the full sidebar is shown.
	SidebarMinTerminalWidth = 90
	// MinTerminalWidth is the minimum width below which Compute returns
	// an empty Layout — the screen cannot show the four-region skeleton
	// at any reasonable density.
	MinTerminalWidth = 50
	// MinTerminalHeight is the minimum height for the same reason.
	MinTerminalHeight = 6
)

// Rect describes a region's outer dimensions in cells.
type Rect struct {
	Width  int
	Height int
}

// IsEmpty reports whether the rect has no drawable area.
func (r Rect) IsEmpty() bool {
	return r.Width <= 0 || r.Height <= 0
}

// Layout groups the four screen regions returned by Compute.
type Layout struct {
	TopBar    Rect
	Sidebar   Rect
	Main      Rect
	BottomBar Rect
}

// Options carry user-toggleable layout preferences (e.g. the sidebar
// has been manually hidden via Ctrl+B). They override the auto-fit
// behavior in Compute.
type Options struct {
	HideSidebar bool
}

// Compute returns a Layout describing how to subdivide a terminal of
// the given width and height into the four screen regions.
//
// The top bar reserves 1 row across the full width. The bottom hint
// bar reserves 2 rows across the full width. The sidebar reserves 14
// columns on the left side of the inner region when the terminal is
// at least SidebarMinTerminalWidth wide; below that the sidebar is
// auto-hidden. Options.HideSidebar forces the sidebar hidden at any
// width.
//
// Inputs below MinTerminalWidth × MinTerminalHeight return an empty
// Layout — the caller is expected to render a "terminal too small"
// fallback message.
func Compute(width, height int) Layout {
	return ComputeWith(width, height, Options{})
}

// ComputeWith is the option-aware version of Compute.
func ComputeWith(width, height int, opts Options) Layout {
	if width < MinTerminalWidth || height < MinTerminalHeight {
		return Layout{}
	}

	sidebarWidth := SidebarFullWidth
	if width < SidebarMinTerminalWidth || opts.HideSidebar {
		sidebarWidth = 0
	}

	innerHeight := height - TopBarHeight - BottomBarHeight
	if innerHeight < 1 {
		innerHeight = 1
	}

	mainWidth := width - sidebarWidth
	if mainWidth < 1 {
		mainWidth = 1
	}

	return Layout{
		TopBar:    Rect{Width: width, Height: TopBarHeight},
		Sidebar:   Rect{Width: sidebarWidth, Height: innerHeight},
		Main:      Rect{Width: mainWidth, Height: innerHeight},
		BottomBar: Rect{Width: width, Height: BottomBarHeight},
	}
}
