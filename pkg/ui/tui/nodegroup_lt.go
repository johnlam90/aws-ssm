package tui

import (
	"fmt"
	"sort"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/johnlam90/aws-ssm/pkg/aws"
)

// LaunchTemplateUpdateState tracks inline launch template updates
type LaunchTemplateUpdateState struct {
	ClusterName        string
	NodeGroupName      string
	LaunchTemplateID   string
	LaunchTemplateName string
	CurrentVersion     string
	Options            []launchTemplateVersionOption
	Cursor             int
	Loading            bool
	Submitting         bool
	RequestedVersion   string
	Error              error
}

type launchTemplateVersionOption struct {
	Value  string
	Label  string
	Detail string
}

// startNodeGroupLaunchTemplateUpdate opens the update prompt for a node group
func (m Model) startNodeGroupLaunchTemplateUpdate(ng NodeGroup) (Model, tea.Cmd) {
	if strings.TrimSpace(ng.LaunchTemplateID) == "" {
		m.statusMessage = "Selected node group has no launch template configured"
		return m, nil
	}

	m.scaling = nil
	state := &LaunchTemplateUpdateState{
		ClusterName:        ng.ClusterName,
		NodeGroupName:      ng.Name,
		LaunchTemplateID:   ng.LaunchTemplateID,
		LaunchTemplateName: ng.LaunchTemplateName,
		CurrentVersion:     ng.LaunchTemplateVersion,
		Loading:            true,
	}
	m.ltUpdate = state
	m.statusMessage = ""
	m.searchActive = false

	return m, LoadLaunchTemplateVersionsCmd(m.ctx, m.client, ng.LaunchTemplateID, ng.ClusterName, ng.Name)
}

// handleLaunchTemplateKeys processes keybindings for the launch template overlay
func (m Model) handleLaunchTemplateKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.ltUpdate == nil {
		return m, nil
	}

	state := m.ltUpdate

	if state.Submitting {
		if msg.Type == tea.KeyEsc {
			m.ltUpdate = nil
		}
		return m, nil
	}

	switch msg.String() {
	case "esc":
		m.ltUpdate = nil
	case "up", "k":
		if state.Loading || len(state.Options) == 0 {
			return m, nil
		}
		if state.Cursor > 0 {
			state.Cursor--
		}
	case "down", "j":
		if state.Loading || len(state.Options) == 0 {
			return m, nil
		}
		if state.Cursor < len(state.Options)-1 {
			state.Cursor++
		}
	case "r":
		state.Loading = true
		state.Error = nil
		return m, LoadLaunchTemplateVersionsCmd(m.ctx, m.client, state.LaunchTemplateID, state.ClusterName, state.NodeGroupName)
	case "enter":
		if state.Loading || len(state.Options) == 0 {
			return m, nil
		}
		selected := state.Options[state.Cursor]
		state.Submitting = true
		state.RequestedVersion = selected.Value
		state.Error = nil
		return m, UpdateNodeGroupLaunchTemplateCmd(
			m.ctx,
			m.client,
			state.ClusterName,
			state.NodeGroupName,
			state.LaunchTemplateID,
			selected.Value,
		)
	}

	return m, nil
}

// handleLaunchTemplateVersions processes launch template version load results
func (m Model) handleLaunchTemplateVersions(msg LaunchTemplateVersionsMsg) (tea.Model, tea.Cmd) {
	if m.ltUpdate == nil {
		return m, nil
	}

	state := m.ltUpdate
	if state.ClusterName != msg.ClusterName || state.NodeGroupName != msg.NodeGroupName {
		return m, nil
	}

	state.Loading = false
	if msg.Error != nil {
		state.Options = nil
		state.Error = msg.Error
		return m, nil
	}

	state.Options = buildLaunchTemplateOptions(state.CurrentVersion, msg.Versions)
	if len(state.Options) == 0 {
		state.Cursor = 0
		return m, nil
	}

	state.Cursor = findLaunchTemplateCursor(state.Options, state.CurrentVersion)
	return m, nil
}

// handleLaunchTemplateUpdateResult handles the result of a launch template update
func (m Model) handleLaunchTemplateUpdateResult(msg LaunchTemplateUpdateResultMsg) (tea.Model, tea.Cmd) {
	if m.ltUpdate != nil &&
		m.ltUpdate.ClusterName == msg.ClusterName &&
		m.ltUpdate.NodeGroupName == msg.NodeGroupName {
		if msg.Error != nil {
			m.ltUpdate.Submitting = false
			m.ltUpdate.Error = msg.Error
			return m, nil
		}

		name := fmt.Sprintf("%s/%s", m.ltUpdate.ClusterName, m.ltUpdate.NodeGroupName)
		version := msg.Version
		m.ltUpdate = nil
		m.statusMessage = fmt.Sprintf("Updated launch template for %s to %s", name, version)
	} else if msg.Error != nil {
		m.statusMessage = fmt.Sprintf("Launch template update failed: %v", msg.Error)
	}

	m.loading = true
	m.loadingMsg = "Refreshing node groups..."
	return m, LoadNodeGroupsCmd(m.ctx, m.client)
}

func findLaunchTemplateCursor(options []launchTemplateVersionOption, current string) int {
	if len(options) == 0 {
		return 0
	}

	for i, opt := range options {
		if opt.Value == current {
			return i
		}
	}
	return 0
}

func buildLaunchTemplateOptions(current string, versions []aws.LaunchTemplateVersion) []launchTemplateVersionOption {
	var opts []launchTemplateVersionOption

	// Special entries
	special := []launchTemplateVersionOption{
		{
			Value:  "$Latest",
			Label:  "$Latest (latest)",
			Detail: "Always use the newest launch template version",
		},
		{
			Value:  "$Default",
			Label:  "$Default (default)",
			Detail: "Use the template's default version",
		},
	}

	for i := range special {
		if special[i].Value == current {
			special[i].Label += " • current"
		}
	}

	opts = append(opts, special...)

	sort.Slice(versions, func(i, j int) bool {
		return versions[i].VersionNumber > versions[j].VersionNumber
	})

	for _, v := range versions {
		value := fmt.Sprintf("%d", v.VersionNumber)
		label := fmt.Sprintf("Version %s", value)
		var details []string
		if v.DefaultVersion {
			label += " • default"
		}
		if value == current {
			label += " • current"
		}
		if desc := strings.TrimSpace(v.VersionDescription); desc != "" {
			details = append(details, desc)
		}
		if created := strings.TrimSpace(v.CreateTime); created != "" {
			details = append(details, created)
		}

		opts = append(opts, launchTemplateVersionOption{
			Value:  value,
			Label:  label,
			Detail: strings.Join(details, " — "),
		})
	}

	return opts
}

// renderLaunchTemplatePrompt renders the inline update overlay
func (m Model) renderLaunchTemplatePrompt() string {
	if m.ltUpdate == nil {
		return ""
	}

	state := m.ltUpdate
	var b strings.Builder

	title := fmt.Sprintf("Update Launch Template • %s / %s", state.ClusterName, state.NodeGroupName)
	b.WriteString(TitleStyle.Render(title))
	b.WriteString("\n")

	ltName := normalizeValue(state.LaunchTemplateName, "n/a", 0)
	b.WriteString(fmt.Sprintf("Template: %s\n", ltName))
	b.WriteString(fmt.Sprintf("Current version: %s\n", normalizeValue(state.CurrentVersion, "n/a", 0)))
	b.WriteString("\n")

	switch {
	case state.Submitting:
		b.WriteString(LoadingStyle.Render(fmt.Sprintf("Updating to %s ...", state.RequestedVersion)))
		b.WriteString("\n")
	case state.Loading:
		b.WriteString(LoadingStyle.Render("Loading launch template versions..."))
		b.WriteString("\n")
	case len(state.Options) == 0:
		b.WriteString(ErrorStyle.Render("No launch template versions available"))
		b.WriteString("\n")
	default:
		start := state.Cursor - 3
		if start < 0 {
			start = 0
		}
		end := start + 7
		if end > len(state.Options) {
			end = len(state.Options)
			start = end - 7
			if start < 0 {
				start = 0
			}
		}

		for i := start; i < end; i++ {
			option := state.Options[i]
			line := fmt.Sprintf("  %s", option.Label)
			if option.Detail != "" {
				line = fmt.Sprintf("%s — %s", line, option.Detail)
			}
			b.WriteString(RenderSelectableRow(line, i == state.Cursor))
			b.WriteString("\n")
		}

		b.WriteString("\n")
		b.WriteString(HelpStyle.Render("↑/k ↓/j select  enter:apply  r:reload  esc:cancel"))
		b.WriteString("\n")
	}

	if state.Error != nil {
		b.WriteString(ErrorStyle.Render(fmt.Sprintf("Error: %v", state.Error)))
		b.WriteString("\n")
	}

	return PanelStyle.Width(m.width).Render(b.String())
}
