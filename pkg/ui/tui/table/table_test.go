package table

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
)

func TestStateBadge(t *testing.T) {
	cases := []struct {
		state string
		glyph string
	}{
		{"running", GlyphRunning},
		{"RUNNING", GlyphRunning},
		{"active", GlyphRunning},
		{"healthy", GlyphRunning},
		{"pending", GlyphPending},
		{"creating", GlyphPending},
		{"scaling up", GlyphPending},
		{"stopped", GlyphStopped},
		{"stopping", GlyphStopped},
		{"terminated", GlyphTerminated},
		{"failed", GlyphTerminated},
		{"unknown", GlyphUnknown},
		{"", GlyphUnknown},
	}
	for _, tc := range cases {
		t.Run(tc.state, func(t *testing.T) {
			got := StateBadge(tc.state)
			if !strings.Contains(got, tc.glyph) {
				t.Errorf("StateBadge(%q) = %q, want glyph %q", tc.state, got, tc.glyph)
			}
			if w := lipgloss.Width(got); w != 1 {
				t.Errorf("StateBadge(%q) width = %d, want 1", tc.state, w)
			}
		})
	}
}

func TestStateBadgeOK(t *testing.T) {
	cases := []struct {
		state string
		want  bool
	}{
		{"running", true},
		{"active", true},
		{"pending", false},
		{"stopped", false},
		{"terminated", false},
		{"", false},
	}
	for _, tc := range cases {
		t.Run(tc.state, func(t *testing.T) {
			if got := StateBadgeOK(tc.state); got != tc.want {
				t.Errorf("StateBadgeOK(%q) = %v, want %v", tc.state, got, tc.want)
			}
		})
	}
}

func TestAllocate_FitsAllAtPrefWidth(t *testing.T) {
	specs := []ColumnSpec{
		{Header: "A", MinWidth: 4, PrefWidth: 10, MaxWidth: 20},
		{Header: "B", MinWidth: 4, PrefWidth: 10, MaxWidth: 20},
		{Header: "C", MinWidth: 4, PrefWidth: 10, MaxWidth: 20},
	}
	got := Allocate(specs, 50) // 30 prefs + 2 gaps = 32; 50 has slack to spread
	for i, c := range got {
		if c.Dropped {
			t.Errorf("column %d unexpectedly dropped", i)
		}
		if c.Width < c.Spec.PrefWidth {
			t.Errorf("column %d width %d below PrefWidth %d", i, c.Width, c.Spec.PrefWidth)
		}
	}
	total := 0
	for _, c := range got {
		total += c.Width
	}
	gaps := len(got) - 1
	if total+gaps > 50 {
		t.Errorf("total %d + gaps %d exceeds 50", total, gaps)
	}
}

func TestAllocate_DropsRightmostWhenTooNarrow(t *testing.T) {
	specs := []ColumnSpec{
		{Header: "A", MinWidth: 10, PrefWidth: 10, MaxWidth: 10},
		{Header: "B", MinWidth: 10, PrefWidth: 10, MaxWidth: 10},
		{Header: "C", MinWidth: 10, PrefWidth: 10, MaxWidth: 10},
	}
	got := Allocate(specs, 22)
	if !got[2].Dropped {
		t.Errorf("expected column C to be dropped (terminal too narrow)")
	}
	if got[0].Dropped || got[1].Dropped {
		t.Errorf("first two columns should not be dropped")
	}
}

func TestAllocate_RespectsMaxWidth(t *testing.T) {
	specs := []ColumnSpec{
		{Header: "A", MinWidth: 4, PrefWidth: 8, MaxWidth: 8},
		{Header: "B", MinWidth: 4, PrefWidth: 8, MaxWidth: 8},
	}
	got := Allocate(specs, 100) // way more slack than MaxWidth allows
	for i, c := range got {
		if c.Width > c.Spec.MaxWidth {
			t.Errorf("column %d width %d exceeds MaxWidth %d", i, c.Width, c.Spec.MaxWidth)
		}
	}
}

func TestFitCell_LeftRightCenter(t *testing.T) {
	cases := []struct {
		value, align string
		width        int
		want         string
	}{
		{"foo", "left", 5, "foo  "},
		{"foo", "right", 5, "  foo"},
		{"foo", "center", 5, " foo "},
		{"longvalue", "left", 5, "long…"},
		{"already5", "left", 5, "alre…"},
		{"x", "center", 4, " x  "},
	}
	for _, tc := range cases {
		t.Run(tc.value+"_"+tc.align, func(t *testing.T) {
			got := FitCell(tc.value, tc.width, tc.align)
			if got != tc.want {
				t.Errorf("FitCell(%q, %d, %q) = %q, want %q", tc.value, tc.width, tc.align, got, tc.want)
			}
			if w := lipgloss.Width(got); w != tc.width {
				t.Errorf("FitCell width = %d, want %d", w, tc.width)
			}
		})
	}
}

func TestFitCell_ZeroAndNegativeWidth(t *testing.T) {
	if got := FitCell("foo", 0, "left"); got != "" {
		t.Errorf("FitCell width 0 = %q, want empty", got)
	}
	if got := FitCell("foo", -1, "left"); got != "" {
		t.Errorf("FitCell width -1 = %q, want empty", got)
	}
}
