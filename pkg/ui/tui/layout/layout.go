// Package layout owns region dimension arithmetic for the TUI.
//
// Compute returns a Layout describing the four screen regions (top bar,
// sidebar, main, bottom bar) for a given terminal size. Each region is
// expressed as a Rect; consumers render their content into those Rects.
package layout

// Chrome dimensions for the flat three-region layout: top chrome bar
// (1 row), main panel (fills the middle), bottom hint bar (2 rows).
// The sidebar is gone — view switching happens via the command
// palette and Mod+1..N hotkeys. Sidebar/SidebarMinTerminalWidth
// constants remain at zero for backwards-compat imports.
const (
	TopBarHeight            = 1
	BottomBarHeight         = 2
	TopRuleHeight           = 1
	BottomRuleHeight        = 1
	SidebarFullWidth        = 0
	SidebarMinTerminalWidth = 0
	MinTerminalWidth        = 50
	MinTerminalHeight       = 8
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

// Options carry user-toggleable layout preferences. The HideSidebar
// option is retained for source compatibility; the flat layout has
// no sidebar so the option is currently a no-op.
type Options struct {
	HideSidebar bool
}

// Compute returns a Layout describing how to subdivide a terminal of
// the given width and height into the screen regions.
//
// Flat three-region layout:
//
//	row 0:           top chrome bar (brand · breadcrumb / region · profile)
//	row 1:           horizontal rule
//	rows 2..H-4:     main panel (full width)
//	row H-3:         horizontal rule
//	rows H-2..H-1:   bottom hint bar (key hints / status footer)
//
// Sidebar rect is always zero in the flat layout but remains in
// Layout for backwards-compat with imports that read it.
func Compute(width, height int) Layout {
	return ComputeWith(width, height, Options{})
}

// ComputeWith is the option-aware version of Compute.
func ComputeWith(width, height int, _ Options) Layout {
	if width < MinTerminalWidth || height < MinTerminalHeight {
		return Layout{}
	}

	mainHeight := height - TopBarHeight - TopRuleHeight - BottomRuleHeight - BottomBarHeight
	if mainHeight < 1 {
		mainHeight = 1
	}

	return Layout{
		TopBar:    Rect{Width: width, Height: TopBarHeight},
		Sidebar:   Rect{Width: 0, Height: 0},
		Main:      Rect{Width: width, Height: mainHeight},
		BottomBar: Rect{Width: width, Height: BottomBarHeight},
	}
}
