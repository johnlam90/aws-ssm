package layout

import "testing"

func TestCompute(t *testing.T) {
	tests := []struct {
		name          string
		width, height int
		wantTop       Rect
		wantSidebar   Rect
		wantMain      Rect
		wantBottom    Rect
	}{
		{
			name:        "zero size returns empty layout",
			width:       0,
			height:      0,
			wantTop:     Rect{0, 0},
			wantSidebar: Rect{0, 0},
			wantMain:    Rect{0, 0},
			wantBottom:  Rect{0, 0},
		},
		{
			name:        "negative size returns empty layout",
			width:       -10,
			height:      -5,
			wantTop:     Rect{0, 0},
			wantSidebar: Rect{0, 0},
			wantMain:    Rect{0, 0},
			wantBottom:  Rect{0, 0},
		},
		{
			name:        "below minimum returns empty layout",
			width:       49,
			height:      24,
			wantTop:     Rect{0, 0},
			wantSidebar: Rect{0, 0},
			wantMain:    Rect{0, 0},
			wantBottom:  Rect{0, 0},
		},
		{
			name:        "below minimum height returns empty layout",
			width:       80,
			height:      5,
			wantTop:     Rect{0, 0},
			wantSidebar: Rect{0, 0},
			wantMain:    Rect{0, 0},
			wantBottom:  Rect{0, 0},
		},
		{
			name:        "narrow terminal hides sidebar",
			width:       80,
			height:      24,
			wantTop:     Rect{80, 1},
			wantSidebar: Rect{0, 21},
			wantMain:    Rect{80, 21},
			wantBottom:  Rect{80, 2},
		},
		{
			name:        "edge: 89 cols still hides sidebar",
			width:       89,
			height:      24,
			wantTop:     Rect{89, 1},
			wantSidebar: Rect{0, 21},
			wantMain:    Rect{89, 21},
			wantBottom:  Rect{89, 2},
		},
		{
			name:        "edge: 90 cols shows full sidebar",
			width:       90,
			height:      24,
			wantTop:     Rect{90, 1},
			wantSidebar: Rect{14, 21},
			wantMain:    Rect{76, 21},
			wantBottom:  Rect{90, 2},
		},
		{
			name:        "standard 120x40 terminal",
			width:       120,
			height:      40,
			wantTop:     Rect{120, 1},
			wantSidebar: Rect{14, 37},
			wantMain:    Rect{106, 37},
			wantBottom:  Rect{120, 2},
		},
		{
			name:        "wide 200x60 terminal",
			width:       200,
			height:      60,
			wantTop:     Rect{200, 1},
			wantSidebar: Rect{14, 57},
			wantMain:    Rect{186, 57},
			wantBottom:  Rect{200, 2},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Compute(tt.width, tt.height)

			if got.TopBar != tt.wantTop {
				t.Errorf("TopBar = %+v, want %+v", got.TopBar, tt.wantTop)
			}
			if got.Sidebar != tt.wantSidebar {
				t.Errorf("Sidebar = %+v, want %+v", got.Sidebar, tt.wantSidebar)
			}
			if got.Main != tt.wantMain {
				t.Errorf("Main = %+v, want %+v", got.Main, tt.wantMain)
			}
			if got.BottomBar != tt.wantBottom {
				t.Errorf("BottomBar = %+v, want %+v", got.BottomBar, tt.wantBottom)
			}
		})
	}
}

func TestCompute_RegionsTileFullScreen(t *testing.T) {
	cases := []struct {
		width, height int
	}{
		{80, 24},
		{120, 40},
		{200, 60},
	}
	for _, c := range cases {
		layout := Compute(c.width, c.height)
		// TopBar + (Sidebar | Main) + BottomBar must equal width × height.
		if layout.TopBar.Width != c.width {
			t.Errorf("TopBar width %d, want %d", layout.TopBar.Width, c.width)
		}
		if layout.BottomBar.Width != c.width {
			t.Errorf("BottomBar width %d, want %d", layout.BottomBar.Width, c.width)
		}
		bodyWidth := layout.Sidebar.Width + layout.Main.Width
		if bodyWidth != c.width {
			t.Errorf("Sidebar(%d) + Main(%d) = %d, want %d", layout.Sidebar.Width, layout.Main.Width, bodyWidth, c.width)
		}
		totalHeight := layout.TopBar.Height + layout.Main.Height + layout.BottomBar.Height
		if totalHeight != c.height {
			t.Errorf("TopBar(%d) + Main(%d) + BottomBar(%d) = %d, want %d",
				layout.TopBar.Height, layout.Main.Height, layout.BottomBar.Height, totalHeight, c.height)
		}
		if layout.Sidebar.Height != layout.Main.Height {
			t.Errorf("Sidebar.Height %d, Main.Height %d, must match", layout.Sidebar.Height, layout.Main.Height)
		}
	}
}

func TestRect_IsEmpty(t *testing.T) {
	tests := []struct {
		name string
		r    Rect
		want bool
	}{
		{"zero", Rect{0, 0}, true},
		{"zero width", Rect{0, 10}, true},
		{"zero height", Rect{10, 0}, true},
		{"negative width", Rect{-1, 10}, true},
		{"negative height", Rect{10, -1}, true},
		{"non-empty", Rect{1, 1}, false},
		{"large", Rect{200, 60}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.r.IsEmpty(); got != tt.want {
				t.Errorf("IsEmpty() = %v, want %v", got, tt.want)
			}
		})
	}
}
