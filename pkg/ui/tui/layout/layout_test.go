package layout

import "testing"

func TestCompute_PreservesPhase1Behavior(t *testing.T) {
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
			name:        "tiny terminal collapses to main only",
			width:       10,
			height:      3,
			wantTop:     Rect{0, 0},
			wantSidebar: Rect{0, 0},
			wantMain:    Rect{10, 3},
			wantBottom:  Rect{0, 0},
		},
		{
			name:        "standard 80x24 terminal",
			width:       80,
			height:      24,
			wantTop:     Rect{0, 0},
			wantSidebar: Rect{0, 0},
			wantMain:    Rect{80, 24},
			wantBottom:  Rect{0, 0},
		},
		{
			name:        "wide 200x60 terminal",
			width:       200,
			height:      60,
			wantTop:     Rect{0, 0},
			wantSidebar: Rect{0, 0},
			wantMain:    Rect{200, 60},
			wantBottom:  Rect{0, 0},
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
