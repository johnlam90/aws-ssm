package tui

import (
	"fmt"

	"github.com/atotto/clipboard"
)

// CopyResult represents the type of data that was copied
type CopyResult struct {
	Type    string // "instance_id", "private_ip", "name", "arn", etc.
	Value   string
	Success bool
	Error   error
}

// copyToClipboard writes text to the system clipboard and returns the result
func copyToClipboard(text string) error {
	return clipboard.WriteAll(text)
}

// clipboardAvailable checks if clipboard is available on this system
func clipboardAvailable() bool {
	// Attempt a read - if it fails, clipboard is likely not available
	_, err := clipboard.ReadAll()
	return err == nil
}

// getClipboardContent returns what should be copied based on the current view and selection
func (m Model) getClipboardContent() CopyResult {
	switch m.currentView {
	case ViewEC2Instances:
		return m.getEC2ClipboardContent()
	case ViewEKSClusters:
		return m.getEKSClipboardContent()
	case ViewASGs:
		return m.getASGClipboardContent()
	case ViewNodeGroups:
		return m.getNodeGroupClipboardContent()
	case ViewNetworkInterfaces:
		return m.getNetworkClipboardContent()
	default:
		return CopyResult{Success: false, Error: fmt.Errorf("copying not supported for this view")}
	}
}

func (m Model) getEC2ClipboardContent() CopyResult {
	instances := m.getEC2Instances()
	if m.cursor < 0 || m.cursor >= len(instances) {
		return CopyResult{Success: false, Error: fmt.Errorf("no instance selected")}
	}

	inst := instances[m.cursor]
	// Primary copy target is Instance ID
	return CopyResult{
		Type:    "Instance ID",
		Value:   inst.InstanceID,
		Success: true,
	}
}

func (m Model) getEKSClipboardContent() CopyResult {
	clusters := m.getEKSClusters()
	if m.cursor < 0 || m.cursor >= len(clusters) {
		return CopyResult{Success: false, Error: fmt.Errorf("no cluster selected")}
	}

	cluster := clusters[m.cursor]
	return CopyResult{
		Type:    "Cluster Name",
		Value:   cluster.Name,
		Success: true,
	}
}

func (m Model) getASGClipboardContent() CopyResult {
	asgs := m.getASGs()
	if m.cursor < 0 || m.cursor >= len(asgs) {
		return CopyResult{Success: false, Error: fmt.Errorf("no ASG selected")}
	}

	asg := asgs[m.cursor]
	return CopyResult{
		Type:    "ASG Name",
		Value:   asg.Name,
		Success: true,
	}
}

func (m Model) getNodeGroupClipboardContent() CopyResult {
	nodeGroups := m.getNodeGroups()
	if m.cursor < 0 || m.cursor >= len(nodeGroups) {
		return CopyResult{Success: false, Error: fmt.Errorf("no node group selected")}
	}

	ng := nodeGroups[m.cursor]
	// Format as cluster/nodegroup for easy identification
	return CopyResult{
		Type:    "Node Group",
		Value:   fmt.Sprintf("%s/%s", ng.ClusterName, ng.Name),
		Success: true,
	}
}

func (m Model) getNetworkClipboardContent() CopyResult {
	interfaces := m.getNetworkInterfaces()
	if m.cursor < 0 || m.cursor >= len(interfaces) {
		return CopyResult{Success: false, Error: fmt.Errorf("no instance selected")}
	}

	iface := interfaces[m.cursor]
	return CopyResult{
		Type:    "Instance ID",
		Value:   iface.InstanceID,
		Success: true,
	}
}

// handleCopy handles the 'y' (yank) keypress to copy selection to clipboard
func (m Model) handleCopy() Model {
	content := m.getClipboardContent()
	if !content.Success {
		m.setStatusMessage(fmt.Sprintf("Copy failed: %v", content.Error), "error")
		return m
	}

	if err := copyToClipboard(content.Value); err != nil {
		m.setStatusMessage(fmt.Sprintf("Copy failed: %v", err), "error")
		return m
	}

	m.setStatusMessage(fmt.Sprintf("Copied %s: %s", content.Type, content.Value), "success")
	return m
}

// handleCopyIP copies the private IP address for EC2 instances
func (m Model) handleCopyIP() Model {
	if m.currentView != ViewEC2Instances {
		m.setStatusMessage("IP copy only available for EC2 instances", "error")
		return m
	}

	instances := m.getEC2Instances()
	if m.cursor < 0 || m.cursor >= len(instances) {
		m.setStatusMessage("No instance selected", "error")
		return m
	}

	inst := instances[m.cursor]
	ip := inst.PrivateIP
	if ip == "" {
		m.setStatusMessage("No private IP available", "error")
		return m
	}

	if err := copyToClipboard(ip); err != nil {
		m.setStatusMessage(fmt.Sprintf("Copy failed: %v", err), "error")
		return m
	}

	m.setStatusMessage(fmt.Sprintf("Copied IP: %s", ip), "success")
	return m
}
