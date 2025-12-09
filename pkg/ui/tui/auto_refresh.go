package tui

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// AutoRefreshMsg is sent when auto-refresh timer fires
type AutoRefreshMsg struct {
	View ViewMode
}

// Default auto-refresh interval
const defaultAutoRefreshInterval = 30 * time.Second

// autoRefreshTick creates a command that sends AutoRefreshMsg after the specified interval
func autoRefreshTick(interval time.Duration, view ViewMode) tea.Cmd {
	return tea.Tick(interval, func(t time.Time) tea.Msg {
		return AutoRefreshMsg{View: view}
	})
}

// toggleAutoRefresh toggles auto-refresh for the current view and returns the updated model and command
func (m Model) toggleAutoRefresh() (Model, tea.Cmd) {
	// Toggle the auto-refresh state
	m.autoRefreshEnabled = !m.autoRefreshEnabled

	if m.autoRefreshEnabled {
		// Set default interval if not set
		if m.autoRefreshInterval == 0 {
			m.autoRefreshInterval = defaultAutoRefreshInterval
		}
		m.setStatusMessage("Auto-refresh enabled (30s)", "success")
		return m, autoRefreshTick(m.autoRefreshInterval, m.currentView)
	}

	m.setStatusMessage("Auto-refresh disabled", "info")
	return m, nil
}

// handleAutoRefresh processes auto-refresh messages
func (m Model) handleAutoRefresh(msg AutoRefreshMsg) (Model, tea.Cmd) {
	// Only refresh if:
	// 1. Auto-refresh is still enabled
	// 2. We're on the same view that triggered the refresh
	// 3. Not currently loading
	// 4. No modal is open (scaling, launch template update)
	if !m.autoRefreshEnabled {
		return m, nil
	}

	if m.loading {
		// Reschedule if currently loading
		return m, autoRefreshTick(m.autoRefreshInterval, m.currentView)
	}

	if m.scaling != nil || m.ltUpdate != nil {
		// Reschedule if modal is open
		return m, autoRefreshTick(m.autoRefreshInterval, m.currentView)
	}

	// Only refresh data views, not dashboard or help
	switch m.currentView {
	case ViewDashboard, ViewHelp:
		// Don't auto-refresh these views, but keep timer going for when user navigates
		return m, autoRefreshTick(m.autoRefreshInterval, m.currentView)
	}

	// Perform the actual refresh
	var cmd tea.Cmd
	m.captureSelection(m.currentView)

	switch m.currentView {
	case ViewEC2Instances:
		m.loading = true
		m.loadingMsg = "Auto-refreshing EC2 instances..."
		cmd = LoadEC2InstancesCmd(m.ctx, m.client)
	case ViewEKSClusters:
		m.loading = true
		m.loadingMsg = "Auto-refreshing EKS clusters..."
		cmd = LoadEKSClustersCmd(m.ctx, m.client)
	case ViewASGs:
		m.loading = true
		m.loadingMsg = "Auto-refreshing ASGs..."
		cmd = LoadASGsCmd(m.ctx, m.client)
	case ViewNodeGroups:
		m.loading = true
		m.loadingMsg = "Auto-refreshing node groups..."
		cmd = LoadNodeGroupsCmd(m.ctx, m.client)
	case ViewNetworkInterfaces:
		m.loading = true
		m.loadingMsg = "Auto-refreshing network interfaces..."
		cmd = LoadNetworkInterfacesCmd(m.ctx, m.client)
	}

	// Chain: load data + schedule next tick
	return m, tea.Batch(cmd, autoRefreshTick(m.autoRefreshInterval, m.currentView))
}

// isAutoRefreshEnabled returns whether auto-refresh is currently enabled
func (m Model) isAutoRefreshEnabled() bool {
	return m.autoRefreshEnabled
}
