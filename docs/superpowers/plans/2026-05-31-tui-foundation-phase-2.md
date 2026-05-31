# TUI Foundation — Phase 2: Chrome & Sidebar Live

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Make the four-region skeleton visible. Top chrome bar, left sidebar, and bottom hint bar render real content. Strip per-view footers, headers, and status bars (their content moves into the chrome). Visible UI changes substantially in this phase.

**Architecture:** Two new packages — `pkg/ui/tui/chrome` (top bar + bottom hint bar) and `pkg/ui/tui/sidebar` (left rail). `layout.Compute` extended to reserve rows/cols for the chrome and sidebar. `Model.View()` composes regions via `lipgloss.JoinHorizontal`/`JoinVertical`. Per-view renderers are slimmed: they only emit table + detail + search-bar; their old header/footer/status-bar code is deleted (chrome owns that now).

**Spec ref:** §3, §4.1, §4.2, §4.4, §6.6 phase 2.

---

## File Structure

**Created:**
- `pkg/ui/tui/chrome/topbar.go` — top chrome bar renderer
- `pkg/ui/tui/chrome/bottombar.go` — bottom hint bar + status footer
- `pkg/ui/tui/chrome/chrome_test.go` — tests
- `pkg/ui/tui/sidebar/sidebar.go` — sidebar component
- `pkg/ui/tui/sidebar/items.go` — sidebar entry registry
- `pkg/ui/tui/sidebar/sidebar_test.go` — tests

**Modified:**
- `pkg/ui/tui/layout/layout.go` — reserve regions for chrome/sidebar
- `pkg/ui/tui/layout/layout_test.go` — updated expectations
- `pkg/ui/tui/model.go` — `View()` composes all four regions
- `pkg/ui/tui/dashboard.go` — strip header bar, separator, footer
- `pkg/ui/tui/ec2_view.go` — strip header / footer / status bar calls
- `pkg/ui/tui/eks_view.go` — same
- `pkg/ui/tui/asg_view.go` — same
- `pkg/ui/tui/nodegroup_view.go` — same
- `pkg/ui/tui/network_view.go` — same
- `pkg/ui/tui/help_view.go` — same

**Removed (functions, not files):** `renderHeader`, `renderDashboardHeaderBar`, `renderDashboardSeparator`, `renderDashboardFooter`, `renderEC2Footer`, `renderEKSFooter`, `renderASGFooter`, `renderNodeGroupFooter`, `renderNetworkFooter`, `getStatusBar`, `renderLoading` (top-level loading; per-view loading stays inside main).

---

## Task 1: Layout — reserve chrome and sidebar regions

Update `layout.Compute` to return non-zero rectangles for TopBar (1 row), Sidebar (14 cols when terminal ≥90 cols, 0 otherwise — Phase 2 covers full width only; compact/hidden modes land in Phase 11), and BottomBar (2 rows).

**Files:**
- Modify: `pkg/ui/tui/layout/layout.go`
- Modify: `pkg/ui/tui/layout/layout_test.go`

Constants for sidebar widths and chrome heights live in this file as exported package-level consts so chrome/sidebar packages can import them.

---

## Task 2: Chrome — top bar

Single-line top bar: `aws-ssm · <breadcrumb> · <region> · <profile>`. Constructs from `(appName, breadcrumb, region, profile, width)`. Phase 2 omits clock and account suffix (those need STS + Tick — added in Phase 11). Width-truncates from the right.

**Files:**
- Create: `pkg/ui/tui/chrome/topbar.go`
- Create: `pkg/ui/tui/chrome/chrome_test.go`

---

## Task 3: Chrome — bottom hint bar

Two lines: line 1 = key hints (per-view registry of hint pairs), line 2 = status footer (pagination + selection + toast). The hint registry is a function `HintsForView(view, focus) []Hint` exported from `chrome`.

**Files:**
- Modify: `pkg/ui/tui/chrome/bottombar.go` (new)
- Modify: `pkg/ui/tui/chrome/chrome_test.go`

---

## Task 4: Sidebar component

`Render(rect, items, selected) string` returns a 14-col vertical block. Items are `(icon, label, count, viewMode)`. Selected entry uses `┃` left-edge accent.

**Files:**
- Create: `pkg/ui/tui/sidebar/sidebar.go`
- Create: `pkg/ui/tui/sidebar/items.go`
- Create: `pkg/ui/tui/sidebar/sidebar_test.go`

---

## Task 5: Strip per-view headers/footers/status-bars

Remove `m.renderHeader`, footer functions, and `getStatusBar` calls from every per-view renderer. The chrome owns those now. Per-view content is now: search bar (if active) → table → detail block.

**Files (each modified):**
- `pkg/ui/tui/dashboard.go`
- `pkg/ui/tui/ec2_view.go`
- `pkg/ui/tui/eks_view.go`
- `pkg/ui/tui/asg_view.go`
- `pkg/ui/tui/nodegroup_view.go`
- `pkg/ui/tui/network_view.go`
- `pkg/ui/tui/help_view.go`

---

## Task 6: Compose View() through all four regions

`View()` now calls `chrome.RenderTopBar`, `sidebar.Render`, `m.renderMainPanel`, `chrome.RenderBottomBar` and joins them with lipgloss.

**Files:**
- Modify: `pkg/ui/tui/model.go`

---

## Task 7: Verify + commit + push

`make verify`, `make build`, push.

---

## Self-review notes

- **Per-view test files** (`dashboard_test.go`, `model_test.go`, etc.) may assert against the old output that included the header/footer/status-bar. Those tests are updated to assert against just the per-view content (no chrome).
- **Compact/hidden sidebar modes** are deliberately deferred to Phase 11 (narrow-terminal handling). Phase 2 always renders full sidebar at 14 cols.
- **Clock and account suffix** in the top bar deferred to Phase 11 too — they require additional infrastructure (`tea.Tick`, STS call).
- **Single-commit phase** (vs. multiple intermediate commits): Phase 2 is atomic because intermediate states either have duplicate footers (chrome bottom bar + per-view footer) or missing chrome with stripped per-view footer. Atomic landing avoids a broken intermediate.
