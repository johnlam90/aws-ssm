package tui

import "testing"

func TestPaletteMatches_EmptyReturnsFullCatalog(t *testing.T) {
	got := palettematches("")
	if len(got) != len(paletteCatalog()) {
		t.Errorf("empty input should return full catalog (%d), got %d", len(paletteCatalog()), len(got))
	}
}

func TestPaletteMatches_PrefixWins(t *testing.T) {
	got := palettematches("ec")
	if len(got) == 0 {
		t.Fatal("expected matches for 'ec'")
	}
	if got[0].Name != "ec2" {
		t.Errorf("first match for 'ec' should be 'ec2', got %q", got[0].Name)
	}
}

func TestPaletteMatches_AliasMatches(t *testing.T) {
	got := palettematches("network")
	if len(got) == 0 {
		t.Fatal("expected matches for 'network'")
	}
	if got[0].Name != "eni" {
		t.Errorf("first match for 'network' should be 'eni' (alias), got %q", got[0].Name)
	}
}

func TestPaletteMatches_NoMatchReturnsEmpty(t *testing.T) {
	got := palettematches("xyzunknown")
	if len(got) != 0 {
		t.Errorf("expected zero matches for 'xyzunknown', got %d", len(got))
	}
}

func TestPaletteMatches_CaseInsensitive(t *testing.T) {
	gotLower := palettematches("ec")
	gotUpper := palettematches("EC")
	if len(gotLower) != len(gotUpper) {
		t.Errorf("case sensitivity mismatch: lower=%d upper=%d", len(gotLower), len(gotUpper))
	}
}

func TestClampInt(t *testing.T) {
	cases := []struct {
		n, lo, hi, want int
	}{
		{5, 0, 10, 5},
		{-1, 0, 10, 0},
		{15, 0, 10, 10},
		{5, 10, 5, 10}, // hi < lo: degenerate, returns lo
	}
	for _, tc := range cases {
		if got := clampInt(tc.n, tc.lo, tc.hi); got != tc.want {
			t.Errorf("clampInt(%d, %d, %d) = %d, want %d", tc.n, tc.lo, tc.hi, got, tc.want)
		}
	}
}
