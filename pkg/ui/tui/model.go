package tui

import (
	"context"
	"fmt"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/johnlam90/aws-ssm/pkg/aws"
)

// Model represents the main TUI model
type Model struct {
	ctx         context.Context
	client      *aws.Client
	config      Config
	currentView ViewMode
	viewStack   []ViewMode // For navigation history
	cursor      int        // Current cursor position in lists
	width       int
	height      int
	ready       bool
	err         error
	loading     bool
	loadingMsg  string
	spinner     spinner.Model

	// Data
	ec2Instances []EC2Instance
	eksClusters  []EKSCluster
	asgs         []ASG

	// Dashboard menu items
	menuItems []MenuItem

	// Post-exit actions
	pendingSSMSession *string // Instance ID to connect to after TUI exits
	pendingEKSCluster *string // Cluster name to show details for after TUI exits
	pendingASGScale   *string // ASG name to scale after TUI exits
	shouldQuit        bool    // Flag to indicate we should quit
}

// NewModel creates a new TUI model
func NewModel(ctx context.Context, client *aws.Client, config Config) Model {
	menuItems := []MenuItem{
		{
			Title:       "EC2 Instances",
			Description: "View and manage EC2 instances",
			View:        ViewEC2Instances,
			Icon:        "",
		},
		{
			Title:       "EKS Clusters",
			Description: "Manage EKS clusters and node groups",
			View:        ViewEKSClusters,
			Icon:        "",
		},
		{
			Title:       "Auto Scaling Groups",
			Description: "View and scale ASGs",
			View:        ViewASGs,
			Icon:        "",
		},
		{
			Title:       "Network Interfaces",
			Description: "View EC2 network interfaces and ENIs",
			View:        ViewNetworkInterfaces,
			Icon:        "",
		},
		{
			Title:       "Help",
			Description: "View keybindings and help",
			View:        ViewHelp,
			Icon:        "",
		},
	}

	// Create spinner
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = LoadingStyle

	return Model{
		ctx:         ctx,
		client:      client,
		config:      config,
		currentView: ViewDashboard,
		viewStack:   []ViewMode{},
		cursor:      0,
		menuItems:   menuItems,
		spinner:     s,
	}
}

// Init initializes the model
func (m Model) Init() tea.Cmd {
	return m.spinner.Tick
}

// Update handles messages and updates the model
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.ready = true
		return m, nil

	case tea.KeyMsg:
		return m.handleKeyPress(msg)

	case DataLoadedMsg:
		return m.handleDataLoaded(msg)

	case ErrorMsg:
		m.err = msg.Err
		m.loading = false
		return m, nil

	case spinner.TickMsg:
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}

	return m, nil
}

// View renders the current view
func (m Model) View() string {
	if !m.ready {
		return "Initializing..."
	}

	// Show loading spinner if loading
	if m.loading {
		return fmt.Sprintf("\n\n   %s %s\n\n", m.spinner.View(), m.loadingMsg)
	}

	// Render based on current view
	switch m.currentView {
	case ViewDashboard:
		return m.renderDashboard()
	case ViewEC2Instances:
		return m.renderEC2Instances()
	case ViewEKSClusters:
		return m.renderEKSClusters()
	case ViewASGs:
		return m.renderASGs()
	case ViewHelp:
		return m.renderHelp()
	default:
		return "Unknown view"
	}
}

// handleKeyPress handles keyboard input
func (m Model) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Global keybindings
	switch msg.String() {
	case "ctrl+c", "q":
		// Only quit from dashboard, otherwise go back
		if m.currentView == ViewDashboard {
			return m, tea.Quit
		}
		return m.navigateBack(), nil

	case "?":
		// Toggle help
		if m.currentView == ViewHelp {
			return m.navigateBack(), nil
		}
		m.pushView(ViewHelp)
		return m, nil
	}

	// View-specific keybindings
	switch m.currentView {
	case ViewDashboard:
		return m.handleDashboardKeys(msg)
	case ViewEC2Instances:
		return m.handleEC2Keys(msg)
	case ViewEKSClusters:
		return m.handleEKSKeys(msg)
	case ViewASGs:
		return m.handleASGKeys(msg)
	case ViewHelp:
		return m.handleHelpKeys(msg)
	}

	return m, nil
}

// pushView pushes the current view onto the stack and switches to a new view
func (m *Model) pushView(view ViewMode) {
	m.viewStack = append(m.viewStack, m.currentView)
	m.currentView = view
	m.cursor = 0 // Reset cursor when changing views
}

// navigateBack pops the previous view from the stack
func (m Model) navigateBack() Model {
	if len(m.viewStack) > 0 {
		m.currentView = m.viewStack[len(m.viewStack)-1]
		m.viewStack = m.viewStack[:len(m.viewStack)-1]
		m.cursor = 0
	}
	return m
}

// handleDataLoaded handles data loaded messages
func (m Model) handleDataLoaded(msg DataLoadedMsg) (tea.Model, tea.Cmd) {
	m.loading = false

	if msg.Error != nil {
		m.err = msg.Error
		return m, nil
	}

	switch msg.View {
	case ViewEC2Instances:
		m.ec2Instances = msg.Instances
	case ViewEKSClusters:
		m.eksClusters = msg.Clusters
	case ViewASGs:
		m.asgs = msg.ASGs
	}

	return m, nil
}

// getStatusBar returns the status bar content - minimal
func (m Model) getStatusBar() string {
	region := m.config.Region
	if region == "" {
		region = "default"
	}

	profile := m.config.Profile
	if profile == "" {
		profile = "default"
	}

	return fmt.Sprintf("%s | %s | %s",
		region, profile, m.currentView.String())
}

// GetError returns the current error (for external access)
func (m Model) GetError() error {
	return m.err
}

// GetRegion returns the current region (for external access)
func (m Model) GetRegion() string {
	return m.config.Region
}

// GetProfile returns the current profile (for external access)
func (m Model) GetProfile() string {
	return m.config.Profile
}

// GetPendingSSMSession returns the instance ID to connect to after TUI exits
func (m Model) GetPendingSSMSession() *string {
	return m.pendingSSMSession
}

// GetPendingEKSCluster returns the cluster name to show details for after TUI exits
func (m Model) GetPendingEKSCluster() *string {
	return m.pendingEKSCluster
}

// GetPendingASGScale returns the ASG name to scale after TUI exits
func (m Model) GetPendingASGScale() *string {
	return m.pendingASGScale
}

// scheduleSSMSession schedules an SSM session to start after the TUI exits
func (m *Model) scheduleSSMSession(instanceID string) tea.Cmd {
	m.pendingSSMSession = &instanceID
	m.shouldQuit = true
	return tea.Quit
}

// scheduleClusterDetails schedules cluster details display after the TUI exits
func (m *Model) scheduleClusterDetails(clusterName string) tea.Cmd {
	m.pendingEKSCluster = &clusterName
	m.shouldQuit = true
	return tea.Quit
}

// scheduleASGScale schedules ASG scaling after the TUI exits
func (m *Model) scheduleASGScale(asgName string) tea.Cmd {
	m.pendingASGScale = &asgName
	m.shouldQuit = true
	return tea.Quit
}
