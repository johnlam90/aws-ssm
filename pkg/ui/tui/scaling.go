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
		Input:          "",
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
		Input:          "",
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
		if msg.Type == tea.KeyEsc {
			m.scaling = nil
		}
		return m, nil
	}
	switch msg.Type {
	case tea.KeyEsc:
		return m.clearScaling(), nil
	case tea.KeyCtrlU:
		return m.clearScalingInput(), nil
	case tea.KeyBackspace:
		return m.backspaceScalingInput(), nil
	case tea.KeyEnter:
		return m.submitScaling()
	}
	return m.appendScalingDigit(msg), nil
}

func (m Model) clearScaling() Model {
	m.scaling = nil
	return m
}

func (m Model) clearScalingInput() Model {
	m.scaling.Input = ""
	m.scaling.Error = nil
	return m
}

func (m Model) backspaceScalingInput() Model {
	if len(m.scaling.Input) > 0 {
		m.scaling.Input = m.scaling.Input[:len(m.scaling.Input)-1]
	}
	m.scaling.Error = nil
	return m
}

func (m Model) submitScaling() (tea.Model, tea.Cmd) {
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

	// Check if scaling to 0 requires confirmation
	if requiresScalingConfirmation(desired) {
		// Create confirmation dialog and close scaling modal
		s := m.scaling
		m.confirmation = newScaleToZeroConfirmation(
			s.TargetView,
			s.ASGName,
			s.ClusterName,
			s.NodeGroupName,
			s.CurrentMin,
			s.CurrentMax,
		)
		m.scaling = nil
		return m, nil
	}

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

func (m Model) appendScalingDigit(msg tea.KeyMsg) Model {
	s := msg.String()
	if len(s) == 1 && s[0] >= '0' && s[0] <= '9' {
		if m.scaling.Input == "0" {
			m.scaling.Input = s
		} else {
			m.scaling.Input += s
		}
		m.scaling.Error = nil
	}
	return m
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
		m.setStatusMessage(fmt.Sprintf("Scaled %s to %d", name, desired), "success")
	} else if msg.Error != nil {
		m.setStatusMessage(fmt.Sprintf("Scaling failed: %v", msg.Error), "error")
	}

	switch msg.View {
	case ViewASGs:
		if m.currentView == ViewASGs {
			m.captureSelection(ViewASGs)
		}
		m.loading = true
		m.loadingMsg = "Refreshing Auto Scaling Groups..."
		return m, LoadASGsCmd(m.ctx, m.client)
	case ViewNodeGroups:
		if m.currentView == ViewNodeGroups {
			m.captureSelection(ViewNodeGroups)
		}
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
		subtitle = s.ASGName
	case ViewNodeGroups:
		title = "Scale Node Group"
		subtitle = fmt.Sprintf("%s / %s", s.ClusterName, s.NodeGroupName)
	default:
		title = "Scale Resource"
		subtitle = s.displayName()
	}

	var b strings.Builder
	b.WriteString(ModalTitleStyle().Render(title))
	b.WriteString("\n")
	b.WriteString(SubtitleStyle().Render(subtitle))
	b.WriteString("\n\n")

	b.WriteString(ModalLabelStyle().Render("Current capacity"))
	b.WriteString("\n")
	b.WriteString(fmt.Sprintf("  Desired %d  |  Min %d  |  Max %d  |  Actual %d\n\n",
		s.CurrentDesired, s.CurrentMin, s.CurrentMax, s.CurrentSize))

	b.WriteString(ModalLabelStyle().Render("New desired capacity"))
	b.WriteString("\n")

	currentValue := fmt.Sprintf("%d", s.CurrentDesired)
	target := s.Input
	if s.Submitting {
		display := strings.TrimSpace(target)
		if display == "" {
			display = currentValue
		}
		b.WriteString(LoadingStyle().Render(fmt.Sprintf("  Scaling to %s ...", display)))
		b.WriteString("\n")
	} else {
		var inputField string
		if target == "" {
			inputField = ModalPlaceholderStyle().Render(currentValue)
		} else {
			inputField = ModalInputStyle().Render(target)
		}
		b.WriteString("  ")
		b.WriteString(inputField)
		b.WriteString("\n")
		b.WriteString(ModalHelpStyle().Render("enter:apply   esc:cancel   digits:edit   backspace:delete   ctrl+u:clear"))
		b.WriteString("\n")
	}

	if s.Error != nil {
		b.WriteString("\n")
		b.WriteString(ErrorStyle().Render(fmt.Sprintf("Error: %v", s.Error)))
		b.WriteString("\n")
	}

	modalWidth := calculateModalWidth(m.width)
	modal := ModalStyle().Width(modalWidth).Render(b.String())
	return centerModal(modal, m.width)
}

// renderStatusMessage renders the transient status message with appropriate semantic color
func (m Model) renderStatusMessage() string {
	if strings.TrimSpace(m.statusMessage) == "" {
		return ""
	}

	// Determine message type from status animation if available
	messageType := "success" // Default to success
	if m.statusAnimation != nil && m.statusAnimation.MessageType != "" {
		messageType = m.statusAnimation.MessageType
	}

	return RenderStatusMessage(m.statusMessage, messageType)
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
