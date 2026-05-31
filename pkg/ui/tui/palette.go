package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// PaletteState holds the command palette's transient input and
// selection state. The palette is opened with `:`, accepts free-text
// command names, fuzzy-matches them against a static command catalog,
// and dispatches the chosen command on Enter.
type PaletteState struct {
	Input  string
	Cursor int // index into the filtered command list
}

// PaletteCommand describes one palette entry.
type PaletteCommand struct {
	Name        string   // canonical name (e.g. "ec2")
	Aliases     []string // alternate names (e.g. "instances")
	Description string   // shown in the palette list
}

// paletteCatalog returns the static command catalog. Each model
// session reuses the same catalog; per-call resolution pulls live
// state (focused row, filter, etc.) from the model when executing.
func paletteCatalog() []PaletteCommand {
	return []PaletteCommand{
		{Name: "home", Description: "Go to home / dashboard"},
		{Name: "ec2", Aliases: []string{"instances"}, Description: "Show EC2 instances"},
		{Name: "eks", Aliases: []string{"clusters"}, Description: "Show EKS clusters"},
		{Name: "asg", Aliases: []string{"asgs", "autoscaling"}, Description: "Show Auto Scaling Groups"},
		{Name: "ng", Aliases: []string{"nodegroups"}, Description: "Show EKS node groups"},
		{Name: "eni", Aliases: []string{"network", "interfaces"}, Description: "Show network interfaces (alias for EC2)"},
		{Name: "help", Aliases: []string{"?"}, Description: "Toggle help"},
		{Name: "refresh", Aliases: []string{"r"}, Description: "Refresh current view"},
		{Name: "clear", Description: "Clear current view's filter"},
		{Name: "yank-id", Aliases: []string{"y", "yank"}, Description: "Copy focused resource ID to clipboard"},
		{Name: "quit", Aliases: []string{"q", "exit"}, Description: "Quit aws-ssm"},
	}
}

// openPalette starts a palette session.
func (m Model) openPalette() Model {
	m.palette = &PaletteState{}
	m.searchActive = false
	return m
}

// closePalette ends a palette session without executing a command.
func (m Model) closePalette() Model {
	m.palette = nil
	return m
}

// handlePaletteKeys processes keyboard input while the palette is open.
func (m Model) handlePaletteKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.palette == nil {
		return m, nil
	}

	matches := palettematches(m.palette.Input)
	maxIdx := len(matches) - 1

	switch msg.Type {
	case tea.KeyEsc:
		return m.closePalette(), nil
	case tea.KeyEnter:
		if len(matches) == 0 {
			return m.closePalette(), nil
		}
		chosen := matches[clampInt(m.palette.Cursor, 0, maxIdx)]
		return m.executePaletteCommand(chosen.Name)
	case tea.KeyUp:
		if m.palette.Cursor > 0 {
			m.palette.Cursor--
		}
		return m, nil
	case tea.KeyDown:
		if m.palette.Cursor < maxIdx {
			m.palette.Cursor++
		}
		return m, nil
	case tea.KeyBackspace:
		if len(m.palette.Input) > 0 {
			m.palette.Input = m.palette.Input[:len(m.palette.Input)-1]
		}
		m.palette.Cursor = 0
		return m, nil
	case tea.KeyCtrlU:
		m.palette.Input = ""
		m.palette.Cursor = 0
		return m, nil
	}

	if s := msg.String(); len(s) == 1 && !msg.Alt {
		m.palette.Input += s
		m.palette.Cursor = 0
	}
	return m, nil
}

// executePaletteCommand runs the named command and closes the palette.
func (m Model) executePaletteCommand(name string) (tea.Model, tea.Cmd) {
	m = m.closePalette()

	switch name {
	case "home":
		m.currentView = ViewDashboard
		m.cursor = 0
		return m, nil
	case "ec2", "instances", "eni", "network", "interfaces":
		m.currentView = ViewEC2Instances
		m.loading = true
		m.loadingMsg = "Loading EC2 instances..."
		return m, tea.Batch(
			LoadEC2InstancesCmd(m.ctx, m.client),
			LoadNetworkInterfacesCmd(m.ctx, m.client),
		)
	case "eks", "clusters":
		m.currentView = ViewEKSClusters
		m.loading = true
		m.loadingMsg = "Loading EKS clusters..."
		return m, LoadEKSClustersCmd(m.ctx, m.client)
	case "asg", "asgs", "autoscaling":
		m.currentView = ViewASGs
		m.loading = true
		m.loadingMsg = "Loading Auto Scaling Groups..."
		return m, LoadASGsCmd(m.ctx, m.client)
	case "ng", "nodegroups":
		m.currentView = ViewNodeGroups
		m.loading = true
		m.loadingMsg = "Loading EKS node groups..."
		return m, LoadNodeGroupsCmd(m.ctx, m.client)
	case "help", "?":
		if m.currentView == ViewHelp {
			return m.navigateBack(), nil
		}
		m.pushView(ViewHelp)
		return m, nil
	case "refresh", "r":
		return m.handleRefresh()
	case "clear":
		return m.clearSearchQuery(m.currentView), nil
	case "yank-id", "y", "yank":
		return m.yankFocusedID(), nil
	case "quit", "q", "exit":
		return m, tea.Quit
	}

	m.statusMessage = fmt.Sprintf("Unknown command: %s", name)
	return m, nil
}

// yankFocusedID copies the focused resource's primary identifier to
// the system clipboard. Best-effort; failure shows a status message.
func (m Model) yankFocusedID() Model {
	id := m.currentSelectionKey(m.currentView)
	if id == "" {
		m.statusMessage = "Nothing to yank"
		return m
	}
	if err := writeToClipboard(id); err != nil {
		m.statusMessage = fmt.Sprintf("Yank failed: %v", err)
		return m
	}
	m.statusMessage = "Copied: " + id
	return m
}

// palettematches returns the catalog entries that fuzzy-match input.
func palettematches(input string) []PaletteCommand {
	input = strings.TrimSpace(strings.ToLower(input))
	cat := paletteCatalog()
	if input == "" {
		return cat
	}
	out := make([]PaletteCommand, 0, len(cat))
	// Exact-prefix first.
	for _, c := range cat {
		if strings.HasPrefix(c.Name, input) {
			out = append(out, c)
			continue
		}
		for _, a := range c.Aliases {
			if strings.HasPrefix(a, input) {
				out = append(out, c)
				break
			}
		}
	}
	// Then substring matches not already added.
	seen := make(map[string]bool, len(out))
	for _, c := range out {
		seen[c.Name] = true
	}
	for _, c := range cat {
		if seen[c.Name] {
			continue
		}
		if strings.Contains(c.Name, input) || strings.Contains(c.Description, input) {
			out = append(out, c)
			continue
		}
		for _, a := range c.Aliases {
			if strings.Contains(a, input) {
				out = append(out, c)
				break
			}
		}
	}
	return out
}

// renderPalette returns the centered palette overlay string.
func (m Model) renderPalette() string {
	if m.palette == nil {
		return ""
	}

	matches := palettematches(m.palette.Input)
	maxIdx := len(matches) - 1
	cursor := clampInt(m.palette.Cursor, 0, maxIdx)

	titleStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#58A6FF")).Bold(true)
	inputStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF")).Bold(true)
	keyStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#A371F7"))
	descStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#8B949E"))
	selStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF")).Background(lipgloss.Color("#264F78")).Bold(true)
	hintStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#8B949E")).Italic(true)

	var b strings.Builder
	b.WriteString(titleStyle.Render(": " + m.palette.Input))
	b.WriteString(inputStyle.Render("▍"))
	b.WriteString("\n\n")

	if len(matches) == 0 {
		b.WriteString(descStyle.Render("  no matching commands"))
		b.WriteString("\n")
	} else {
		const maxRows = 8
		shown := matches
		if len(shown) > maxRows {
			shown = shown[:maxRows]
		}
		for i, c := range shown {
			line := fmt.Sprintf("  %s  %s", keyStyle.Render(":"+c.Name), descStyle.Render(c.Description))
			if i == cursor {
				line = selStyle.Render(fmt.Sprintf("▌ :%s  %s", c.Name, c.Description))
			}
			b.WriteString(line)
			b.WriteString("\n")
		}
		if len(matches) > maxRows {
			b.WriteString(descStyle.Render(fmt.Sprintf("  ... and %d more", len(matches)-maxRows)))
			b.WriteString("\n")
		}
	}

	b.WriteString("\n")
	b.WriteString(hintStyle.Render("  ↑↓ select   ↵ run   esc cancel"))

	width := calculateModalWidth(m.width)
	overlay := ModalStyle().Width(width).Render(b.String())
	return centerModal(overlay, m.width)
}

// clampInt clamps n to the closed interval [lo, hi].
func clampInt(n, lo, hi int) int {
	if hi < lo {
		return lo
	}
	if n < lo {
		return lo
	}
	if n > hi {
		return hi
	}
	return n
}
