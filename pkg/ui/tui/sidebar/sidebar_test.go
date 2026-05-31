package sidebar

import (
	"strings"
	"testing"
)

func TestRender_TooSmallReturnsEmpty(t *testing.T) {
	if got := Render(2, 10, nil); got != "" {
		t.Errorf("expected empty for tiny width, got %q", got)
	}
	if got := Render(20, 2, nil); got != "" {
		t.Errorf("expected empty for tiny height, got %q", got)
	}
}

func TestRender_ContainsLabelsAndCounts(t *testing.T) {
	items := []Item{
		{Label: "Home", Count: -1},
		{Label: "EC2", Count: 14, Focus: true},
		{Label: "EKS", Count: 3},
	}
	got := Render(20, 10, items)
	for _, want := range []string{"Home", "EC2", "EKS", "14", "3"} {
		if !strings.Contains(got, want) {
			t.Errorf("rendered sidebar missing %q (got: %q)", want, got)
		}
	}
}

func TestRender_FocusIndicator(t *testing.T) {
	items := []Item{
		{Label: "Home", Count: -1},
		{Label: "EC2", Count: 14, Focus: true},
	}
	got := Render(20, 10, items)
	if !strings.Contains(got, "▎") {
		t.Errorf("focused entry should render the '▎' indicator, got: %q", got)
	}
}

func TestRender_RoundedBorderCorners(t *testing.T) {
	got := Render(20, 8, []Item{{Label: "Home", Count: -1}})
	for _, corner := range []string{"╭", "╮", "╰", "╯"} {
		if !strings.Contains(got, corner) {
			t.Errorf("expected rounded-border corner %q in panel, got: %q", corner, got)
		}
	}
}
