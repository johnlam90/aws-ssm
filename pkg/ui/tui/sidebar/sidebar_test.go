package sidebar

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
)

func TestRender_ZeroDimensionsReturnEmpty(t *testing.T) {
	if got := Render(0, 10, nil); got != "" {
		t.Errorf("expected empty for zero width, got %q", got)
	}
	if got := Render(14, 0, nil); got != "" {
		t.Errorf("expected empty for zero height, got %q", got)
	}
}

func TestRender_LineCountMatchesHeight(t *testing.T) {
	items := []Item{
		{Icon: "⬡", Label: "Home", Count: -1},
		{Icon: "▣", Label: "EC2", Count: 14, Focus: true},
		{Icon: "☸", Label: "EKS", Count: 3},
	}
	got := Render(14, 8, items)
	lines := strings.Split(got, "\n")
	if len(lines) != 8 {
		t.Errorf("line count = %d, want 8", len(lines))
	}
}

func TestRender_LineWidthsExact(t *testing.T) {
	items := []Item{
		{Icon: "⬡", Label: "Home", Count: -1},
		{Icon: "▣", Label: "EC2", Count: 14, Focus: true},
	}
	got := Render(14, 4, items)
	for i, line := range strings.Split(got, "\n") {
		if w := lipgloss.Width(line); w != 14 {
			t.Errorf("line %d width = %d, want 14 (line: %q)", i, w, line)
		}
	}
}

func TestRender_FocusItemContainsIndicator(t *testing.T) {
	items := []Item{
		{Icon: "⬡", Label: "Home", Count: -1},
		{Icon: "▣", Label: "EC2", Count: 14, Focus: true},
	}
	got := Render(14, 2, items)
	if !strings.Contains(got, "┃") {
		t.Errorf("expected focus indicator '┃' in output, got %q", got)
	}
}

func TestRender_ContainsAllLabelsAndCounts(t *testing.T) {
	items := []Item{
		{Icon: "⬡", Label: "Home", Count: -1},
		{Icon: "▣", Label: "EC2", Count: 14},
		{Icon: "☸", Label: "EKS", Count: 3},
	}
	got := Render(14, 6, items)
	for _, want := range []string{"Home", "EC2", "EKS", "14", "3"} {
		if !strings.Contains(got, want) {
			t.Errorf("output missing %q (got: %q)", want, got)
		}
	}
}

func TestRender_PadsExtraLines(t *testing.T) {
	items := []Item{{Icon: "⬡", Label: "Home", Count: -1}}
	got := Render(14, 5, items)
	lines := strings.Split(got, "\n")
	if len(lines) != 5 {
		t.Fatalf("line count = %d, want 5", len(lines))
	}
	// Each line carries the right-side separator (" │"), so blank
	// rows still contain the separator. Strip it before checking.
	for i := 1; i < 5; i++ {
		stripped := strings.TrimRight(strings.TrimSuffix(lines[i], " │"), " ")
		if stripped != "" {
			t.Errorf("line %d expected blank (excluding separator), got %q", i, lines[i])
		}
	}
}
