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
			name:        "below minimum returns empty layout",
			width:       49,
			height:      24,
			wantTop:     Rect{0, 0},
			wantSidebar: Rect{0, 0},
			wantMain:    Rect{0, 0},
			wantBottom:  Rect{0, 0},
		},
		{
			name:        "standard 80x24 terminal",
			width:       80,
			height:      24,
			wantTop:     Rect{80, 1},
			wantSidebar: Rect{0, 0},
			wantMain:    Rect{80, 19},
			wantBottom:  Rect{80, 2},
		},
		{
			name:        "wide 200x60 terminal",
			width:       200,
			height:      60,
			wantTop:     Rect{200, 1},
			wantSidebar: Rect{0, 0},
			wantMain:    Rect{200, 55},
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

func TestRect_IsEmpty(t *testing.T) {
	if !(Rect{0, 0}).IsEmpty() {
		t.Error("zero rect should be empty")
	}
	if (Rect{1, 1}).IsEmpty() {
		t.Error("1x1 rect should not be empty")
	}
}
