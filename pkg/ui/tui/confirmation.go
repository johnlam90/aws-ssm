package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// ConfirmationType represents the type of confirmation
type ConfirmationType int

const (
	ConfirmScaleToZero ConfirmationType = iota
	ConfirmScaleDown
)

// ConfirmationState tracks inline confirmation dialogs
type ConfirmationState struct {
	Type            ConfirmationType
	Message         string
	TargetView      ViewMode
	ASGName         string
	ClusterName     string
	NodeGroupName   string
	RequestedDesired int32
	CurrentMin      int32
	CurrentMax      int32
}

// newScaleToZeroConfirmation creates a confirmation dialog for scaling to zero
func newScaleToZeroConfirmation(view ViewMode, asgName, clusterName, nodeGroupName string, currentMin, currentMax int32) *ConfirmationState {
	var message string
	switch view {
	case ViewASGs:
		message = fmt.Sprintf("Scale ASG '%s' to 0 instances?", asgName)
	case ViewNodeGroups:
		message = fmt.Sprintf("Scale node group '%s/%s' to 0 instances?", clusterName, nodeGroupName)
	default:
		message = "Scale to 0 instances?"
	}

	return &ConfirmationState{
		Type:            ConfirmScaleToZero,
		Message:         message,
		TargetView:      view,
		ASGName:         asgName,
		ClusterName:     clusterName,
		NodeGroupName:   nodeGroupName,
		RequestedDesired: 0,
		CurrentMin:      currentMin,
		CurrentMax:      currentMax,
	}
}

// handleConfirmationKeys processes input while confirmation dialog is active
func (m Model) handleConfirmationKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.confirmation == nil {
		return m, nil
	}

	switch msg.String() {
	case "y", "Y", "enter":
		return m.confirmAction()
	case "n", "N", "esc":
		return m.cancelConfirmation(), nil
	}

	return m, nil
}

// confirmAction executes the confirmed action
func (m Model) confirmAction() (tea.Model, tea.Cmd) {
	if m.confirmation == nil {
		return m, nil
	}

	conf := m.confirmation
	m.confirmation = nil

	switch conf.Type {
	case ConfirmScaleToZero:
		switch conf.TargetView {
		case ViewASGs:
			return m, ScaleASGCmd(m.ctx, m.client, conf.ASGName, 0, conf.CurrentMin, conf.CurrentMax)
		case ViewNodeGroups:
			return m, ScaleNodeGroupCmd(m.ctx, m.client, conf.ClusterName, conf.NodeGroupName, 0, conf.CurrentMin, conf.CurrentMax)
		}
	}

	return m, nil
}

// cancelConfirmation cancels the current confirmation dialog
func (m Model) cancelConfirmation() Model {
	m.confirmation = nil
	m.setStatusMessage("Operation cancelled", "info")
	return m
}

// requiresConfirmation checks if a scaling operation needs confirmation
func requiresScalingConfirmation(desired int32) bool {
	return desired == 0
}

// renderConfirmationDialog renders the confirmation modal
func (m Model) renderConfirmationDialog() string {
	if m.confirmation == nil {
		return ""
	}

	var b strings.Builder

	// Title with warning styling
	b.WriteString(ModalTitleStyle().Render("⚠️  Confirmation Required"))
	b.WriteString("\n\n")

	// Message
	b.WriteString(m.confirmation.Message)
	b.WriteString("\n\n")

	// Warning text
	b.WriteString(ErrorStyle().Render("This will scale all instances down to zero!"))
	b.WriteString("\n\n")

	// Help text
	b.WriteString(ModalHelpStyle().Render("y/enter: confirm   n/esc: cancel"))
	b.WriteString("\n")

	modalWidth := calculateModalWidth(m.width)
	modal := ModalStyle().Width(modalWidth).Render(b.String())
	return centerModal(modal, m.width)
}

// hasActiveConfirmation returns true if a confirmation dialog is showing
func (m Model) hasActiveConfirmation() bool {
	return m.confirmation != nil
}
