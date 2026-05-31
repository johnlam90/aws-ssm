# TUI Foundation Redesign — Design Spec

**Date:** 2026-05-31
**Status:** Design (pre-implementation)
**Scope:** Sub-project #1 of a phased TUI overhaul. This spec covers the foundation only: layout skeleton, sidebar nav, top/bottom chrome, command palette, search bar, state badges, view-state preservation, narrow-terminal behavior, and the ENI→EC2 merge. Later sub-projects (action menu, multi-select handlers, favorites, live refresh, plugins) get their own specs.

## 1. Goals

- Replace the current top-to-bottom text-flow layout with a four-region panel layout.
- Eliminate per-view rendering inconsistencies (separators, header, footer, empty/loading/error states all differ today).
- Make navigation persistent and discoverable: sidebar always visible (when room), no need to bounce through the dashboard to switch views.
- Add a `:` command palette for fast keyboard-driven actions and view jumps.
- Promote search to a permanent, discoverable region of every view.
- Reduce the count of mental concepts: merge the ENI view into EC2, where the data is already keyed.
- Prepare the codebase for the follow-on sub-projects (action menu, multi-select, favorites, live refresh, plugins) by introducing clean component boundaries.

## 2. Non-Goals

- Bulk operations / multi-select handlers (visual gutter only this sub-project).
- Inline action menu on `Enter` (lazygit-style context menu) — sub-project #2.
- Per-resource verbs beyond what exists today plus `y` (yank-id) and `:scale <n>`.
- Open-in-AWS-console, copy-ARN, custom plugin commands.
- Theme switching, ASCII-only fallback, mouse support.
- Persisting favorites or session state to disk.
- Live auto-refresh / polling.
- Surfacing **orphan ENIs** (interfaces not attached to any instance). Today's view doesn't do this either; explicit non-goal so the ENI merge isn't read as closing that door.

## 3. Layout

Every view renders the same four-region skeleton.

```
╭─ aws-ssm ─ ⬡ Home ▸ EC2 Instances ────────── ● us-east-1 │ prod │ acct 1234… │ 14:32:05 ─╮
│                                                                                          │
│  ┃ ⬡ Home   ┃   /  filter (⌘/ to focus)                                                   │
│  ┃ ▣ EC2 14 ┃                                                                              │
│  ┃ ☸ EKS  3 ┃   NAME              INSTANCE ID         PRIV IP   STATE  TYPE     ENIs AGE │
│  ┃ ⚖ ASG  5 ┃   web-prod-01       i-0abc1234567890ab  10.0.1.42 ●      t3.large 2    3d  │
│  ┃ ⛁ NG   8 ┃   api-prod-02       i-0def4567890abcde  10.0.1.55 ●      m5.xl    1    3d  │
│  ┃ ?  Help  ┃ ▌ analytics-staging i-09a8b7c6d5e4f3g2  10.0.2.13 ●      c6i.4xl  4    12h │
│  ┃          ┃   batch-runner-04   i-01234567abcdef    10.0.2.99 ◐      r5.2xl   1    1m  │
│  ┃          ┃   archive-svc-01    i-fedcba9876543210  —         ○      t2.med   1    30d │
│  ┃          ┃                                                                              │
│  ┃          ┠──────────────────────────────────────────────────────────────────────────── │
│  ┃          ┃  analytics-staging · i-09a8b7c6d5e4f3g2                                      │
│  ┃          ┃  Type c6i.4xlarge · AZ us-east-1b · Up 12h · IAM ssm-role                    │
│  ┃          ┃  IPs 10.0.2.13 (priv) · DNS ip-10-0-2-13.ec2.internal                        │
│  ┃          ┃  SGs sg-prod-web, sg-shared-ssm                                               │
│  ┃          ┃  Interfaces (4)                                                               │
│  ┃          ┃    IFACE  CARD DEVICE SUBNET            CIDR           SECURITY GROUP        │
│  ┃          ┃    eni-1  0    0      subnet-0a1b2c3d   10.0.2.0/24    sg-prod-web           │
│  ┃          ┃    eni-2  0    1      subnet-0a1b2c3d   10.0.2.0/24    sg-shared-ssm         │
├──────────────┴───────────────────────────────────────────────────────────────────────────┤
│  /  search   ↵  connect    s  scale    y  yank-id    r  refresh    :  cmd    ?  help     │
╰─ Showing 3–7 of 14 · 1 selected ─────────────────────────────────────────────────────────╯
```

### 3.1 Regions

1. **Top chrome bar** — single line: app brand · breadcrumb · region health dot · profile · account suffix · clock.
2. **Left sidebar** — six entries (Home + 4 resource views + Help). Three width modes: full (14 cols), compact (4 cols, icons only), hidden.
3. **Main panel** — search bar (always docked) → table → divider → detail block.
4. **Bottom hint bar** — context-aware key hints + status footer (pagination, selection count, transient toast).

### 3.2 Visual language

- Outer frame uses rounded unicode corners (`╭╮╰╯`); inner separators use box drawing (`┃ ┠ ┴ │`).
- State glyphs: `●` running/healthy/active, `◐` pending/scaling, `○` stopped/idle, `✕` failed/terminated. Color-coded with the existing semantic palette.
- Selected row: `▌` left accent + selected-row background spanning full main width.
- Region health dot in top bar reflects AWS API call success: green ok, amber rate-limited, red persistent error.

## 4. Components

### 4.1 Top chrome bar

Single line, left and right segments separated by an em-dash fill.

| Segment | Source | Drop order |
|---|---|---|
| `aws-ssm` | static brand | never drops |
| `⬡ Home ▸ <view>` | breadcrumb | never drops |
| `● us-east-1` | `client.GetRegion()` + health | 4th (last) to drop |
| `prod` | `config.Profile` | 3rd to drop |
| `acct 1234…` | STS `GetCallerIdentity`, last 4 of account, lazy + cached | 2nd to drop |
| `14:32:05` | local clock, 1Hz `tea.Tick` | 1st to drop |

Approximate widths at which segments drop: clock at <120 cols, account at <100, profile at <90, region at <80. Below 80 cols only `aws-ssm · view-name` remains. Implementation tunes thresholds based on actual rendered width.

### 4.2 Left sidebar

Three width modes adapt to terminal width:

| Mode | Width | Trigger | Contents |
|---|---|---|---|
| Full | 14 | ≥90 cols | icon + label + count |
| Compact | 4 | 70–89 cols | icon only |
| Hidden | 0 | <70 cols, or `Ctrl+B` toggle | — |

- Six entries (post ENI merge): Home, EC2, EKS, ASG, NodeGroups, Help.
- Help selects the help overlay (does not navigate to a separate view).
- Selected entry: bright accent + filled left edge `┃`.
- Counts dim until data is loaded for that view; Home and Help show no count.
- Sidebar focus toggle: `Ctrl+B`. With focus, `j/k`/`↑↓` move within the sidebar; `Enter`/`l`/`→` switches to the chosen view's main panel. `1`–`6` jump directly without entering sidebar focus.
- Manual hide overrides the auto-collapse breakpoint and is sticky for the session.

### 4.3 Main panel

Vertical stack inside the main region:

1. **Search bar (always docked, 1 row).**
   - Empty (unfocused): `/  filter (⌘/ to focus)` in placeholder color, with the live refresh indicator right-aligned (`◐ refreshing…` during loads, `↻ updated 2s ago` for ~3s after success, then collapses).
   - Focused: highlighted border, blinking caret, claims input. Same syntax as today (`name:`, `state:`, `tag:k=v`, etc.).
   - Applied (filtered, unfocused): query is rendered with token coloring; right side shows `N → M matches` and an `✕ clear` chip (Esc to clear).

2. **Table.**
   - Header row: dim secondary color, bottom border, sticky.
   - Adaptive column widths via a column-spec allocator (see §6.4). No more hard-coded `%-32s` formats.
   - All views gain an **AGE** column (relative time since launch/creation). EC2 also gains an **ENIs** count column (post-merge).
   - Selected row: `▌` accent + full-width highlight.
   - Multi-select gutter (1 col) on the leftmost edge: `☐`/`☑`. Visual only this sub-project — handlers come in sub-project #3. The gutter is reserved now to avoid a column shift later.
   - Pagination indicator moves out of the table (it lives in the bottom hint bar's status footer).

3. **Detail block.**
   - Separator `─` divides table from detail.
   - Detail height: `min(content_height, 40% of main_height)`, at least 4 rows when there's room. Hidden entirely when terminal is <18 rows tall.
   - Renders via a unified `renderDetail(view, item)` that delegates to per-view detail builders.
   - Format: title line (`name · id`), then 1–3 dense info lines using `·` separators, then optional groups (SGs, AZs, LBs, target groups, tags, **interfaces**) only when present. Empty fields are skipped, not rendered as `n/a` filler.
   - Long values wrap to next line indented; truncation only as last resort.
   - When focus is empty (mid-load, empty list), block shows a one-liner hint instead of an empty box.
   - **EC2 detail block specifically** gets a new "Interfaces" group (table sub-render: IFACE · CARD · DEVICE · SUBNET · CIDR · SG) — lifted from today's `network_view.go`. Lazy-loaded per focused instance, cached, invalidated on `r`.

### 4.4 Bottom hint bar (2 lines)

- **Line 1 — key hints:** dynamic, based on `(currentView, focus)`. Always shows `/`, `↵`, view-specific verbs, `r`, `:`, `?`. Sidebar-focused mode swaps to: `↑↓` switch · `↵` open · `Esc` back.
- **Line 2 — status footer:** pagination (`Showing N–M of T`) + filter chip if active + selection count + transient toast (`✓ Scaled web-asg to 4` for 2s with fade).
- Replaces today's per-view `renderEC2Footer`, `renderEKSFooter`, etc.

## 5. Interaction model

### 5.1 Focus model

Single focus enum drives key dispatch:

```go
type Focus int
const (
    FocusMain Focus = iota
    FocusSidebar
    FocusSearch
    FocusPalette
    FocusModal
)
```

- Visual cue: focused region's border uses `theme.AccentBlue()`; others use `theme.Border()`.
- Default: `FocusMain`. Returns to `FocusMain` on `Esc`.
- Replaces today's mix of `searchActive`, `scaling != nil`, `ltUpdate != nil` flags. The structs stay; the *who-claims-keys* question is answered once.

### 5.2 Hotkey table

| Key | Action | Where |
|---|---|---|
| `Ctrl+B` | Toggle sidebar focus / visibility | Anywhere |
| `1`–`6` | Jump to view by index (Home, EC2, EKS, ASG, NG, Help) | When not in input |
| `Tab` / `Shift+Tab` | Cycle focus regions (Main → Sidebar → Search → Main); skipped regions (hidden sidebar) drop out of the cycle | Anywhere |
| `:` | Open command palette | Anywhere |
| `/` | Focus search | Any list view |
| `Esc` | Drop focus / close overlay / clear filter | Anywhere |
| `g g` / `G` | Top / bottom | Lists |
| `Ctrl+D` / `Ctrl+U` | Page down / up | Lists |
| `j` `k` `↓` `↑` | Move | Lists, sidebar |
| `Enter` | Primary action (connect / details / scale) | Lists |
| `r` | Refresh current view | Lists |
| `?` | Toggle help overlay | Anywhere |
| `q` | Back / quit (with double-press confirmation, kept) | Anywhere |
| `Space` | Toggle multi-select mark (visual only) | Lists |
| `y` | Yank focused resource ID to clipboard | Lists |
| `i` | Toggle detail block visibility | Lists |
| `s` | Open scaling modal | ASG, NG |
| `u` | Open launch-template update modal (kept) | NG |

### 5.3 Command palette (`:`)

Centered overlay (~60 cols × ~12 rows, capped at 80% of screen).

- Opens on `:` from any non-input context. Only one overlay is active at a time: when a modal (scaling / launch-template) is up, `:` is ignored; conversely, opening a modal closes the palette. Search auto-blurs when palette opens.
- Fuzzy-filtered scrolling list; Esc closes; Enter runs.
- Sources, in priority order: exact-prefix → fuzzy match → recent history.

**Foundation command catalog:**

| Command | Args | Behavior | Where |
|---|---|---|---|
| `:home` `:ec2` `:eks` `:asg` `:ng` `:help` | — | Switch view | Anywhere |
| `:eni` `:network` | — | Alias for `:ec2` (with detail focused on Interfaces) | Anywhere |
| `:refresh` `:r` | — | Force-reload current view | List views |
| `:filter <token>` | e.g. `state:running` | Set search query | List views |
| `:clear` | — | Clear current view's filter | List views |
| `:scale <n>` | non-negative int | Apply scaling to focused ASG/NG (skips modal) | ASG, NG |
| `:yank-id` | — | Copy focused resource ID to clipboard | List views |
| `:quit` `:q` | — | Exit TUI | Anywhere |

History (last 50) in-memory only this sub-project. On-disk persistence is a separate sub-project's concern.

### 5.4 View state preservation

Per-view memory survives view switches:

```go
type ViewState struct {
    Cursor      int
    ScrollTop   int
    SearchQuery string
    Selected    string // selection key
}

m.viewStates map[ViewMode]ViewState
```

- Snapshot on view switch; restore on entry.
- Cursor restoration is best-effort: prefer matching `Selected` to a current row index; fall back to `Cursor` if still in bounds; otherwise reset to 0.
- Existing `m.searchQueries` and `m.selectedItems` fold into `ViewState`.
- In-memory only (does not persist across runs in this sub-project).

### 5.5 Empty / loading / error states

Unified across all views, rendered *inside* the main panel (top bar, sidebar, hint bar stay visible — today the loading state replaces the entire screen):

- `⠋  Loading EC2 instances…` (uses existing `spinner.Model`).
- `⊘  No EC2 instances in us-east-1.   r  refresh    /  filter`.
- `✕  Couldn't load EC2 instances: <message>                           r  retry`.

## 6. Implementation architecture

### 6.1 Code organization

```
pkg/ui/tui/
├── model.go                 # Model + Update/View dispatch only (~250 lines target)
├── focus.go                 # Focus enum + transitions
├── viewstate.go             # ViewState + per-view memory
├── loaders.go               # Async data-load tea.Cmds (extracted from today's types.go)
├── types.go                 # Data types only (EC2Instance, ASG, …)
├── styles.go                # + new tokens: AccentSelectionBg, AccentInputBorder, AccentSubtle
├── animations.go            # unchanged
├── navigation.go            # NavigationKey + key bindings (kept)
├── layout/
│   ├── layout.go            # Region budget computation (sidebar/main/detail dimensions)
│   ├── breakpoints.go       # Width/height → layout-mode classification
│   └── chrome.go            # Top bar + bottom hint bar renderers
├── sidebar/
│   ├── sidebar.go           # Sidebar component (full / compact / hidden)
│   └── items.go             # Sidebar item registry (icon, label, count source)
├── palette/
│   ├── palette.go           # Command palette overlay + input handling
│   ├── commands.go          # Command registry + execution dispatch
│   └── fuzzy.go             # Thin wrapper over pkg/ui/fuzzy/parser
├── views/
│   ├── home.go              # Home renderer (replaces today's dashboard.go)
│   ├── ec2.go               # slimmed; gains ENIs column + interfaces detail group
│   ├── eks.go
│   ├── asg.go
│   ├── nodegroup.go
│   └── help.go              # overlay-style
├── table/
│   ├── columns.go           # ColumnSpec + adaptive width allocator
│   ├── render.go            # Generic table renderer
│   └── badges.go            # State glyphs + colors
├── detail/
│   ├── detail.go            # Generic detail panel renderer
│   └── builders.go          # Per-view detail builders (incl. EC2 interfaces)
├── search/
│   └── searchbar.go         # Visual layer; matchers stay in their current files
└── overlays/
    ├── help.go              # Help overlay
    ├── scaling.go           # was scaling.go
    └── nodegroup_lt.go      # was nodegroup_lt.go
```

**Files removed:** `network_view.go` (lifted into `views/ec2.go` + `detail/builders.go`), `dashboard.go` (replaced by `views/home.go`), `help_view.go` (replaced by `overlays/help.go`), `eks_display.go` (post-exit display kept in `cmd/tui.go` directly).

### 6.2 Render pipeline

A single composition pipeline replaces per-view `renderXxx` from-scratch builds:

```go
func (m Model) View() string {
    if !m.ready { return "Initializing…" }

    layout := layout.Compute(m.width, m.height, m.sidebar.Mode(), m.detailVisible)

    top    := chrome.RenderTopBar(m, layout.TopBar)
    side   := sidebar.Render(m, layout.Sidebar)
    main   := m.renderMainPanel(layout.Main)
    bottom := chrome.RenderBottomBar(m, layout.BottomBar)

    body   := lipgloss.JoinHorizontal(lipgloss.Top, side, main)
    screen := lipgloss.JoinVertical(lipgloss.Left, top, body, bottom)

    return overlays.Apply(screen, m)
}
```

`layout.Compute` is the single source of truth for region dimensions. `overlays.Apply` is the centered-overlay compositor reused by help, palette, scaling, lt-update.

### 6.3 Update path

Focus-dispatched switch:

```go
switch m.focus {
case FocusPalette: return m.palette.Update(msg)
case FocusModal:   return m.activeModal.Update(msg)
case FocusSearch:  return m.searchbar.Update(msg)
case FocusSidebar: return m.sidebar.Update(msg)
default:           return m.handleMainKey(msg)
}
```

### 6.4 Adaptive column widths

Each view declares `[]ColumnSpec{Header, MinWidth, PrefWidth, MaxWidth, Render(item) string}`. The allocator runs once per frame:

1. Sum `MinWidth`s. If > available width, columns drop right-to-left until they fit.
2. Distribute slack proportionally toward `PrefWidth`.
3. Apply `MaxWidth` ceilings; redistribute remainder.
4. Truncate row values exceeding their column's allocated width with `…`.

Deterministic. The single biggest source of regression risk in the refactor — golden tests cover a wide grid of widths.

### 6.5 ENI merge specifics

- `network_view.go` deleted. Its detail-table renderer (`renderInterfaceColumn*`) lifts into `detail/builders.go` as `detailEC2InterfacesGroup`.
- `aws.InstanceInterfaces` type and loader stay (`pkg/aws`).
- New per-instance lazy loader: `LoadEC2InterfacesCmd(ctx, client, instanceID)` triggered on focus change. Result cached in `Model` keyed by `InstanceID`; cache invalidated on `r`.
- EC2 column spec gains `ENIs` (count) — populated from cached interfaces if available, otherwise blank with a faint `…` placeholder.
- EC2 search matcher gains tokens: `subnet:`, `cidr:`, `sg:`, `eni:`. Matchers reuse logic from today's network filter.
- `ViewMode.ViewNetworkInterfaces` is removed from the enum. Migration: any persisted state referencing the old enum is ignored on load (in-memory only, so trivial).
- `:eni` / `:network` palette commands resolve to `:ec2`. Documented in help overlay.

### 6.6 Phased migration

The implementation plan (next step, via the `writing-plans` skill) will sequence these as ordered, each-shippable phases:

1. Skeleton without behavior change. New layout pipeline; sidebar disabled (width 0); existing per-view renderers wrapped.
2. Sidebar + top bar + bottom hint bar live. Per-view footers removed.
3. Adaptive columns + state glyphs.
4. Detail-block unification (EKS / NG gain detail blocks; EC2 gains Interfaces group).
5. ENI merge (delete network_view.go; route tokens; lazy interface loader; column).
6. Search bar always docked + filter chip + match counter.
7. Focus model + view-state preservation.
8. Command palette.
9. Home view (replaces dashboard).
10. Help overlay (remove ViewHelp from enum).
11. Narrow-terminal breakpoints + height handling + min-size guard.
12. Polish: `:scale`, `:yank-id`, sidebar collapse hotkey, `i` toggle.

Each phase compiles, passes lint and tests, and ships independently.

## 7. Testing

Existing tests are kept and updated in place:
- `model_test.go`, `nodegroup_lt_test.go`, `search_test.go`, `styles_test.go`, `benchmark_test.go` — kept; updated to reflect the new model fields and removed `ViewNetworkInterfaces` enum value.
- `dashboard_test.go` → renamed to `home_test.go` and rewritten against the new Home view.
- `network_view`-related tests (any) — folded into `views/ec2_test.go` and `detail/builders_test.go` for the Interfaces group.

New test files paired with new components:

- `layout/layout_test.go` — table-driven `(width, height, sidebarMode, detailVisible)` → expected region rectangles.
- `layout/breakpoints_test.go` — width transitions across thresholds (89↔90, 69↔70, etc.).
- `palette/commands_test.go` — palette parsing + dispatch table.
- `palette/fuzzy_test.go` — command-name fuzzy matching sanity.
- `sidebar/sidebar_test.go` — render snapshots in full / compact / hidden modes.
- `table/columns_test.go` — adaptive width allocator (extensive table tests).
- `table/render_test.go` — golden output for known column specs and row sets.
- `detail/builders_test.go` — golden detail blocks per view, including the new EC2 Interfaces group.
- `viewstate_test.go` — round-trip cursor + selection + search across switches.

Coverage target: post-refactor must not drop below the current `make test-coverage` baseline.

## 8. Risks

| Risk | Mitigation |
|---|---|
| Adaptive column allocator surprises in edge cases (3-col terminal, all-empty columns, very long IDs). | Deterministic allocator; golden tests over wide grid; minimum-size guard renders a friendly "resize" message below 50 cols. |
| `lipgloss.JoinHorizontal` on bordered children with mismatched heights produces visual artifacts. | Pad regions to consistent heights before joining; focused unit tests. |
| `tea.Tick` for the clock leaks on shutdown. | Use the Tea-idiomatic Tick pattern; cancel via `tea.Quit`'s natural lifecycle. |
| `GetCallerIdentity` fails or is slow. | Best-effort, single-shot, cached; show `acct …` until it resolves; degrade silently if it errors. |
| ENI merge regresses search behavior. | Migration phase #5 has a paired test sweep over each token class (`subnet:`, `cidr:`, `sg:`, `eni:`) on representative fixtures. |
| `pkg/ui/fuzzy` parser doesn't fit palette use case. | Fall back to a tiny standalone matcher in `palette/fuzzy.go`; the corpus is ~10 commands so this is cheap if needed. |
| Big refactor lands as one giant PR and regresses something subtle. | 12-phase plan; each phase is an independent PR; tests on each. |

## 9. Out of scope (explicit)

- Bulk-action handlers (sub-project #3).
- Inline action menu / right-click equivalent (sub-project #2).
- `:open-console`, `:copy-arn`, custom plugin commands (sub-project #2 / #6).
- Persisted favorites (sub-project #4).
- Live auto-refresh tick (sub-project #5).
- Plugin registry (sub-project #6).
- Theme switching, ASCII-only fallback, mouse support.
- **Orphan ENI listing** (interfaces not attached to any instance). Today's behavior preserved. Future: a "Networking" view as a separate sub-project.
- Tabs / multiple instances of the same view in parallel.

## 10. Success criteria

- Every view renders inside the same four-region skeleton.
- Switching from EC2 to EKS to ASG and back preserves cursor, scroll, search, and selection.
- `:` opens a working command palette with the documented commands.
- Search bar is always docked; `/` focuses it; cleared with `Esc`.
- Sidebar auto-collapses at the documented breakpoints; `Ctrl+B` overrides.
- ENI sidebar entry is gone; EC2 view shows ENIs count column and Interfaces group in detail block; `subnet:` / `cidr:` / `sg:` filters work on EC2.
- Help is an overlay reachable from any view via `?`.
- `make verify` passes; coverage holds.
