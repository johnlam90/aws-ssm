package chrome

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
)

func TestRenderTopBar_ReturnsEmptyForZeroWidth(t *testing.T) {
	got := RenderTopBar(TopBarInput{Brand: "aws-ssm", Width: 0})
	if got != "" {
		t.Errorf("expected empty string for zero width, got %q", got)
	}
}

func TestRenderTopBar_FitsToWidth(t *testing.T) {
	cases := []struct {
		name  string
		in    TopBarInput
		width int
	}{
		{"all segments wide", TopBarInput{Brand: "aws-ssm", Breadcrumb: "Home ▸ EC2 Instances", Region: "us-east-1", Profile: "prod", Width: 120}, 120},
		{"all segments tight", TopBarInput{Brand: "aws-ssm", Breadcrumb: "Home", Region: "us-east-1", Profile: "prod", Width: 80}, 80},
		{"brand only", TopBarInput{Brand: "aws-ssm", Width: 50}, 50},
		{"truncated", TopBarInput{Brand: "aws-ssm-very-long-brand-name", Width: 10}, 10},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := RenderTopBar(tc.in)
			gotWidth := lipgloss.Width(got)
			if gotWidth != tc.width {
				t.Errorf("RenderTopBar width = %d, want %d (output: %q)", gotWidth, tc.width, got)
			}
		})
	}
}

func TestRenderTopBar_ContainsExpectedSegments(t *testing.T) {
	got := RenderTopBar(TopBarInput{
		Brand:      "aws-ssm",
		Breadcrumb: "Home ▸ EC2 Instances",
		Region:     "us-east-1",
		Profile:    "prod",
		Width:      120,
	})

	for _, want := range []string{"aws-ssm", "Home ▸ EC2 Instances", "us-east-1", "prod"} {
		if !strings.Contains(got, want) {
			t.Errorf("RenderTopBar output missing %q (got: %q)", want, got)
		}
	}
}

func TestRenderBottomBar_TwoLinesAtExactWidth(t *testing.T) {
	got := RenderBottomBar(BottomBarInput{
		Hints: []Hint{
			{Key: "/", Label: "search"},
			{Key: "↵", Label: "select"},
			{Key: "?", Label: "help"},
		},
		Status: "Showing 1–5 of 14",
		Width:  100,
	})

	lines := strings.Split(got, "\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 lines, got %d (output: %q)", len(lines), got)
	}
	for i, line := range lines {
		if w := lipgloss.Width(line); w != 100 {
			t.Errorf("line %d width = %d, want 100", i+1, w)
		}
	}
}

func TestRenderBottomBar_ContainsHintsAndStatus(t *testing.T) {
	got := RenderBottomBar(BottomBarInput{
		Hints:  []Hint{{Key: "/", Label: "search"}, {Key: "?", Label: "help"}},
		Status: "Showing 1–5 of 14",
		Width:  120,
	})

	for _, want := range []string{"/", "search", "?", "help", "Showing 1–5 of 14"} {
		if !strings.Contains(got, want) {
			t.Errorf("RenderBottomBar missing %q", want)
		}
	}
}

func TestRenderBottomBar_EmptyHintsAndStatus(t *testing.T) {
	got := RenderBottomBar(BottomBarInput{Width: 80})
	lines := strings.Split(got, "\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 lines, got %d", len(lines))
	}
	for i, line := range lines {
		if lipgloss.Width(line) != 80 {
			t.Errorf("line %d width = %d, want 80", i+1, lipgloss.Width(line))
		}
		if strings.TrimSpace(line) != "" {
			t.Errorf("line %d expected blank, got %q", i+1, line)
		}
	}
}

func TestRenderBottomBar_ZeroWidthReturnsEmpty(t *testing.T) {
	got := RenderBottomBar(BottomBarInput{Width: 0})
	if got != "" {
		t.Errorf("expected empty string, got %q", got)
	}
}
