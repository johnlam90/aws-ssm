# TUI Foundation — Phase 1: Layout Pipeline Skeleton

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Introduce a `layout` package and rewire `Model.View()` to compose through it, without changing visual output. Lays the foundation for all subsequent phases (sidebar, top bar, bottom hint bar, detail block) to slot in cleanly.

**Architecture:** A new sub-package `pkg/ui/tui/layout` exposes `layout.Compute(width, height int) Layout` returning four `Rect`s (TopBar, Sidebar, Main, BottomBar). In Phase 1, only `Main` is non-empty — the other regions return zero-sized rectangles so today's screen output is byte-identical. `Model.View()` extracts its current per-view dispatch into `Model.renderMainPanel()` and composes through `layout.Compute`.

**Tech Stack:** Go 1.24+, Charm Bubble Tea, Lipgloss, project's existing `pkg/testing` helpers.

**Spec reference:** `docs/superpowers/specs/2026-05-31-tui-foundation-design.md` §6.1, §6.2, §6.6 phase 1.

---

## File Structure

**Created:**
- `pkg/ui/tui/layout/layout.go` — `Rect`, `Layout` types + `Compute` function (single source of truth for region dimensions).
- `pkg/ui/tui/layout/layout_test.go` — table-driven tests for `Compute`.

**Modified:**
- `pkg/ui/tui/model.go` — `View()` rewired to call `layout.Compute`; per-view dispatch lifted into `renderMainPanel()`.

**Unchanged:** every other file. No per-view renderer changes. No tests outside the new package added or modified.

---

## Task 1: Layout package — types

**Files:**
- Create: `pkg/ui/tui/layout/layout.go`

- [ ] **Step 1: Create the layout package with Rect and Layout types only**

```go
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
```

Write the file exactly as above.

- [ ] **Step 2: Verify it compiles**

Run: `go build ./pkg/ui/tui/layout/...`
Expected: no output, exit 0.

- [ ] **Step 3: Verify the package vets cleanly**

Run: `go vet ./pkg/ui/tui/layout/...`
Expected: no output, exit 0.

---

## Task 2: Layout package — Compute (TDD)

**Files:**
- Create: `pkg/ui/tui/layout/layout_test.go`
- Modify: `pkg/ui/tui/layout/layout.go` (append `Compute`)

- [ ] **Step 1: Write the failing test file**

```go
package layout

import "testing"

func TestCompute_PreservesPhase1Behavior(t *testing.T) {
	tests := []struct {
		name           string
		width, height  int
		wantTop        Rect
		wantSidebar    Rect
		wantMain       Rect
		wantBottom     Rect
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
```

- [ ] **Step 2: Run the test, expect compile failure (Compute undefined)**

Run: `go test ./pkg/ui/tui/layout/... -run TestCompute -v`
Expected: build error containing `undefined: Compute`.

- [ ] **Step 3: Implement Compute by appending to `pkg/ui/tui/layout/layout.go`**

```go
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
```

- [ ] **Step 4: Run the tests, expect all pass**

Run: `go test ./pkg/ui/tui/layout/... -v`
Expected: `PASS` for both `TestCompute_PreservesPhase1Behavior` (5 sub-tests) and `TestRect_IsEmpty` (7 sub-tests).

- [ ] **Step 5: Race + coverage scoped to the new package**

Run: `go test -race -coverprofile=/tmp/layout.cov ./pkg/ui/tui/layout/...`
Expected: PASS, `coverage:` line printed.

- [ ] **Step 6: Confirm coverage is 100% for the new package**

Run: `go tool cover -func=/tmp/layout.cov`
Expected: every function shows `100.0%`. If anything is below 100%, add a test case that hits the uncovered branch and re-run from Step 4.

- [ ] **Step 7: Commit**

```bash
git add pkg/ui/tui/layout/
git commit -m "$(cat <<'EOF'
✨ feat(tui/layout): add Compute pipeline foundation

New package pkg/ui/tui/layout owns region dimension arithmetic for the
TUI. Compute returns a Layout describing the four screen regions
(TopBar, Sidebar, Main, BottomBar). Phase 1 reserves the entire screen
for Main; later phases populate the other regions without changing the
API.

Refs: docs/superpowers/specs/2026-05-31-tui-foundation-design.md
EOF
)"
```

---

## Task 3: Wire Model.View() through layout.Compute

This is the behavior-preserving refactor. The visible output before and after this task must be byte-identical for every code path.

**Files:**
- Modify: `pkg/ui/tui/model.go`

- [ ] **Step 1: Read the current `View` method to understand what we're replacing**

Run: `rg -n "^func \(m Model\) View\(\)" pkg/ui/tui/model.go`
Expected: a single match on a line in `pkg/ui/tui/model.go`. Note the line number; the body that follows is what we replace.

- [ ] **Step 2: Add the layout import and replace `View` + add `renderMainPanel`**

Open `pkg/ui/tui/model.go`. Locate the existing import block (lines 3–12) and add the new import in the third paren-group:

```go
import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/johnlam90/aws-ssm/pkg/aws"
	"github.com/johnlam90/aws-ssm/pkg/ui/tui/layout"
)
```

Replace the existing `View` method (the one currently containing the `switch m.currentView` block) with:

```go
// View renders the current screen by composing through the layout pipeline.
//
// Phase 1 of the foundation redesign: layout.Compute is consulted for
// the screen rectangles, but only the Main region is non-empty, so the
// visible output is byte-identical to the prior implementation. The
// indirection is the seam that lets later phases populate sidebar, top
// bar, and bottom hint bar without rewriting View again.
func (m Model) View() string {
	if !m.ready {
		return "Initializing..."
	}

	rects := layout.Compute(m.width, m.height)
	if rects.Main.IsEmpty() {
		return "Initializing..."
	}

	return m.renderMainPanel()
}

// renderMainPanel renders the main region's contents by dispatching to
// the per-view renderer for the current view, mirroring the previous
// switch in View() exactly. Phase 1 deliberately keeps each per-view
// renderer untouched; Phase 2 onward will pass region dimensions in.
func (m Model) renderMainPanel() string {
	if m.loading {
		return fmt.Sprintf("\n\n   %s %s\n\n", m.spinner.View(), m.loadingMsg)
	}

	switch m.currentView {
	case ViewDashboard:
		return m.renderDashboard()
	case ViewEC2Instances:
		return m.renderEC2Instances()
	case ViewEKSClusters:
		return m.renderEKSClusters()
	case ViewASGs:
		return m.renderASGs()
	case ViewNodeGroups:
		return m.renderNodeGroups()
	case ViewNetworkInterfaces:
		return m.renderNetworkInterfaces()
	case ViewHelp:
		return m.renderHelp()
	default:
		return "Unknown view"
	}
}
```

`layout.Compute` has a real job in Phase 1: it is the gate that decides whether the terminal is large enough to render. `IsEmpty()` short-circuits to the same "Initializing..." path the existing `!m.ready` guard uses, so a 0×0 terminal at startup behaves the same as before. Later phases replace this scaffolding with proper region composition.

- [ ] **Step 3: Verify the file compiles**

Run: `go build ./pkg/ui/tui/...`
Expected: no output, exit 0.

- [ ] **Step 4: Verify `go vet` is clean**

Run: `go vet ./pkg/ui/tui/...`
Expected: no output, exit 0.

- [ ] **Step 5: Run the existing TUI test suite — every test must still pass**

Run: `go test ./pkg/ui/tui/...`
Expected: PASS for all existing tests (`model_test.go`, `dashboard_test.go`, `nodegroup_lt_test.go`, `search_test.go`, `styles_test.go`, plus the newly added `layout_test.go`). If anything fails, the refactor changed observable behavior — revert the `View` rewrite and inspect the diff for any output difference.

- [ ] **Step 6: Run the full project verify gate**

Run: `make verify`
Expected: PASS. This runs `go fmt`, `go vet`, `golangci-lint run`, and `go test -race -coverprofile=coverage.out ./...` per the project Makefile. If lint flags the new code (e.g. naming, package comments), fix in place and re-run.

- [ ] **Step 7: Commit**

```bash
git add pkg/ui/tui/model.go
git commit -m "$(cat <<'EOF'
♻️ refactor(tui): compose View through layout.Compute pipeline

Model.View now calls layout.Compute and delegates to renderMainPanel,
mirroring the existing per-view dispatch exactly. Visible output is
unchanged in Phase 1; the indirection is the seam later phases use to
slot in the sidebar, top bar, and bottom hint bar without rewriting
View again.

Refs: docs/superpowers/specs/2026-05-31-tui-foundation-design.md §6.2
EOF
)"
```

---

## Task 4: Smoke-test the binary

A fast manual sanity check that the TUI still renders on a live terminal. Optional if the test suite already covers it; included because a behavior-preserving refactor of `View()` deserves a real eyeball.

**Files:** none (build + run only)

- [ ] **Step 1: Build the binary**

Run: `make build`
Expected: PASS; `./aws-ssm` (or `dist/aws-ssm`) exists and is freshly built.

- [ ] **Step 2: Open the TUI**

Run: `./aws-ssm tui` (or the path printed by `make build`).
Expected: dashboard renders identically to the pre-refactor version. Press `j`/`k` to confirm cursor movement; `Enter` on EC2 to confirm view switching loads instances; `Esc` returns to dashboard; `q` quits cleanly.

If anything looks different from before, capture the diff (region affected, key sequence to reproduce) and revert Task 3.

No commit for this task — it's a manual gate.

---

## Task 5: Wrap-up — final coverage gate

**Files:** none (verification only)

- [ ] **Step 1: Generate coverage report**

Run: `make test-coverage`
Expected: PASS; coverage report at `coverage.html`.

- [ ] **Step 2: Confirm package-level coverage held**

Run: `go test -coverprofile=/tmp/tui.cov ./pkg/ui/tui/... && go tool cover -func=/tmp/tui.cov | tail -1`
Expected: aggregate coverage line. Compare against the pre-Phase-1 baseline (`git stash` + run + `git stash pop` if you need a clean baseline). Coverage must not drop. The new `layout` package adds coverage; the modified `model.go` should not lose any since the refactor only restructures dispatch.

If coverage dropped, find the regression with `go tool cover -html=/tmp/tui.cov` and add the missing test before declaring Phase 1 done.

- [ ] **Step 3: Push the branch (optional, only if requested)**

The branch (`feature/tui-foundation-redesign`) holds the design spec commit and the Phase 1 commits. Do **not** push unless the user has explicitly asked. The repository's git rules require explicit user request before pushing.

---

## Phase 1 done — what you have

- New `pkg/ui/tui/layout` package with `Rect`, `Layout`, `Compute`. 100% covered.
- `Model.View()` composes through `layout.Compute` and `renderMainPanel`. Visible output unchanged.
- All existing tests pass; `make verify` clean.
- Three commits on `feature/tui-foundation-redesign`:
  1. Design spec.
  2. `feat(tui/layout): add Compute pipeline foundation`.
  3. `refactor(tui): compose View through layout.Compute pipeline`.

Phase 2 (sidebar + top bar + bottom hint bar live) is the next plan to write — re-brainstorm or jump straight into a new plan once Phase 1 is merged.
