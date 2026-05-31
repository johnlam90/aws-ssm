// Package layout owns region dimension arithmetic for the TUI.
//
// Compute returns a Layout describing the four screen regions (top bar,
// sidebar, main, bottom bar) for a given terminal size. Each region is
// expressed as a Rect; consumers render their content into those Rects.
//
// Phase 1 returns only a non-empty Main region — sidebar/top/bottom are
// reserved for later phases and are zero-sized in Phase 1, preserving
// today's visual output exactly.
package layout

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

// Compute returns a Layout describing how to subdivide a terminal of
// the given width and height into the four screen regions.
//
// Phase 1 reserves the entire screen for the Main region; TopBar,
// Sidebar, and BottomBar are deliberately zero-sized so today's
// per-view rendering is preserved byte-for-byte. Later phases populate
// the other regions; the API is stable so callers do not change.
//
// Non-positive inputs return an empty Layout (all rects zero) so
// callers do not have to guard for unrealistic terminal sizes.
func Compute(width, height int) Layout {
	if width <= 0 || height <= 0 {
		return Layout{}
	}

	return Layout{
		TopBar:    Rect{Width: 0, Height: 0},
		Sidebar:   Rect{Width: 0, Height: 0},
		Main:      Rect{Width: width, Height: height},
		BottomBar: Rect{Width: 0, Height: 0},
	}
}
