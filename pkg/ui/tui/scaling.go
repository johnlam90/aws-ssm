package tui

import (
	"fmt"
	"strconv"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// ScalingState holds context for inline scaling prompts
type ScalingState struct {
	TargetView       ViewMode
	ASGName          string
	ClusterName      string
	NodeGroupName    string
	CurrentMin       int32
	CurrentMax       int32
	CurrentDesired   int32
	CurrentSize      int32
	Input            string
	Submitting       bool
	RequestedDesired int32
	Error            error
}

func newASGScalingState(asg ASG) *ScalingState {
	return &ScalingState{
		TargetView:     ViewASGs,
		ASGName:        asg.Name,
		CurrentMin:     asg.MinSize,
		CurrentMax:     asg.MaxSize,
		CurrentDesired: asg.DesiredCapacity,
		CurrentSize:    asg.CurrentSize,
		Input:          fmt.Sprintf("%d", asg.DesiredCapacity),
	}
}

func newNodeGroupScalingState(ng NodeGroup) *ScalingState {
	return &ScalingState{
		TargetView:     ViewNodeGroups,
		ClusterName:    ng.ClusterName,
		NodeGroupName:  ng.Name,
		CurrentMin:     ng.MinSize,
		CurrentMax:     ng.MaxSize,
		CurrentDesired: ng.DesiredSize,
		CurrentSize:    ng.CurrentSize,
		Input:          fmt.Sprintf("%d", ng.DesiredSize),
	}
}

func (s *ScalingState) displayName() string {
	switch s.TargetView {
	case ViewASGs:
		return fmt.Sprintf("ASG %s", s.ASGName)
	case ViewNodeGroups:
		return fmt.Sprintf("Node group %s/%s", s.ClusterName, s.NodeGroupName)
	default:
		return "resource"
	}
}

// startASGScaling opens the scaling prompt for an ASG
func (m Model) startASGScaling(asg ASG) Model {
	m.scaling = newASGScalingState(asg)
	m.ltUpdate = nil
	m.statusMessage = ""
	m.searchActive = false
	return m
}

// startNodeGroupScaling opens the scaling prompt for a node group
func (m Model) startNodeGroupScaling(ng NodeGroup) Model {
	m.scaling = newNodeGroupScalingState(ng)
	m.ltUpdate = nil
	m.statusMessage = ""
	m.searchActive = false
	return m
}

// handleScalingKeys processes input while the scaling overlay is active
func (m Model) handleScalingKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.scaling == nil {
		return m, nil
	}

	if m.scaling.Submitting {
		// Allow cancelling while request is in-flight
		if msg.Type == tea.KeyEsc {
			m.scaling = nil
		}
		return m, nil
	}

	switch msg.Type {
	case tea.KeyEsc:
		m.scaling = nil
		return m, nil
	case tea.KeyCtrlU:
		m.scaling.Input = ""
		m.scaling.Error = nil
		return m, nil
	case tea.KeyBackspace:
		if len(m.scaling.Input) > 0 {
			m.scaling.Input = m.scaling.Input[:len(m.scaling.Input)-1]
		}
		m.scaling.Error = nil
		return m, nil
	case tea.KeyEnter:
		if strings.TrimSpace(m.scaling.Input) == "" {
			m.scaling.Error = fmt.Errorf("enter a desired capacity")
			return m, nil
		}
		val, err := strconv.ParseInt(m.scaling.Input, 10, 32)
		if err != nil || val < 0 {
			m.scaling.Error = fmt.Errorf("invalid capacity")
			return m, nil
		}

		desired := int32(val)
		m.scaling.Submitting = true
		m.scaling.RequestedDesired = desired
		m.scaling.Error = nil

		switch m.scaling.TargetView {
		case ViewASGs:
			return m, ScaleASGCmd(m.ctx, m.client, m.scaling.ASGName, desired, m.scaling.CurrentMin, m.scaling.CurrentMax)
		case ViewNodeGroups:
			return m, ScaleNodeGroupCmd(m.ctx, m.client, m.scaling.ClusterName, m.scaling.NodeGroupName, desired, m.scaling.CurrentMin, m.scaling.CurrentMax)
		default:
			m.scaling = nil
			return m, nil
		}
	}

	input := msg.String()
	if len(input) == 1 && input[0] >= '0' && input[0] <= '9' {
		if m.scaling.Input == "0" {
			m.scaling.Input = input
		} else {
			m.scaling.Input += input
		}
		m.scaling.Error = nil
	}

	return m, nil
}

// handleScalingResult processes the result of a scaling operation
func (m Model) handleScalingResult(msg ScalingResultMsg) (tea.Model, tea.Cmd) {
	if m.scaling != nil && m.scaling.TargetView == msg.View {
		if msg.Error != nil {
			m.scaling.Submitting = false
			m.scaling.Error = msg.Error
			return m, nil
		}
		name := m.scaling.displayName()
		desired := m.scaling.RequestedDesired
		m.scaling = nil
		m.statusMessage = fmt.Sprintf("Scaled %s to %d", name, desired)
	} else if msg.Error != nil {
		m.statusMessage = fmt.Sprintf("Scaling failed: %v", msg.Error)
	}

	switch msg.View {
	case ViewASGs:
		m.loading = true
		m.loadingMsg = "Refreshing Auto Scaling Groups..."
		return m, LoadASGsCmd(m.ctx, m.client)
	case ViewNodeGroups:
		m.loading = true
		m.loadingMsg = "Refreshing node groups..."
		return m, LoadNodeGroupsCmd(m.ctx, m.client)
	default:
		return m, nil
	}
}

// renderScalingPrompt returns the scaling overlay for the provided view
func (m Model) renderScalingPrompt(view ViewMode) string {
	if m.scaling == nil || m.scaling.TargetView != view {
		return ""
	}

	s := m.scaling
	var title, subtitle string
	switch view {
	case ViewASGs:
		title = "Scale Auto Scaling Group"
		subtitle = fmt.Sprintf("%s", s.ASGName)
	case ViewNodeGroups:
		title = "Scale Node Group"
		subtitle = fmt.Sprintf("%s / %s", s.ClusterName, s.NodeGroupName)
	default:
		title = "Scale Resource"
		subtitle = s.displayName()
	}

	input := strings.TrimSpace(s.Input)
	if input == "" {
		input = fmt.Sprintf("%d", s.CurrentDesired)
	}

	var b strings.Builder
	b.WriteString(ModalTitleStyle.Render(title))
	b.WriteString("\n")
	b.WriteString(SubtitleStyle.Render(subtitle))
	b.WriteString("\n\n")

	b.WriteString(ModalLabelStyle.Render("Current capacity"))
	b.WriteString("\n")
	b.WriteString(fmt.Sprintf("  Desired %d  |  Min %d  |  Max %d  |  Actual %d\n\n",
		s.CurrentDesired, s.CurrentMin, s.CurrentMax, s.CurrentSize))

	b.WriteString(ModalLabelStyle.Render("New desired capacity"))
	b.WriteString("\n")

	if s.Submitting {
		b.WriteString(LoadingStyle.Render(fmt.Sprintf("  Scaling to %s ...", input)))
		b.WriteString("\n")
	} else {
		b.WriteString("  ")
		b.WriteString(ModalInputStyle.Render(input))
		b.WriteString("\n")
		b.WriteString(ModalHelpStyle.Render("enter:apply   esc:cancel   digits:edit   ctrl+u:clear"))
		b.WriteString("\n")
	}

	if s.Error != nil {
		b.WriteString("\n")
		b.WriteString(ErrorStyle.Render(fmt.Sprintf("Error: %v", s.Error)))
		b.WriteString("\n")
	}

	modalWidth := calculateModalWidth(m.width)
	modal := ModalStyle.Width(modalWidth).Render(b.String())
	return centerModal(modal, m.width)
}

// renderStatusMessage renders the transient status message, if present
func (m Model) renderStatusMessage() string {
	if strings.TrimSpace(m.statusMessage) == "" {
		return ""
	}
	return SuccessStyle.Render(m.statusMessage)
}

func calculateModalWidth(totalWidth int) int {
	switch {
	case totalWidth >= 100:
		return 80
	case totalWidth >= 80:
		return 70
	case totalWidth >= 60:
		return 55
	default:
		if w := totalWidth - 4; w > 30 {
			return w
		}
		return totalWidth
	}
}

func centerModal(modal string, totalWidth int) string {
	if totalWidth <= 0 {
		return modal
	}
	lines := strings.Split(modal, "\n")
	modalWidth := 0
	for _, line := range lines {
		if len(line) > modalWidth {
			modalWidth = len(line)
		}
	}

	padding := 0
	if totalWidth > modalWidth {
		padding = (totalWidth - modalWidth) / 2
		if padding < 0 {
			padding = 0
		}
	}

	if padding == 0 {
		return modal
	}

	pad := strings.Repeat(" ", padding)
	for i, line := range lines {
		if strings.TrimSpace(line) == "" {
			lines[i] = line
			continue
		}
		lines[i] = pad + line
	}
	return "\n" + strings.Join(lines, "\n") + "\n"
}
