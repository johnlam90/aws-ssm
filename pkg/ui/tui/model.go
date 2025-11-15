package tui

import (
	"context"
	"fmt"
	"strings"
	"time"

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

	// Navigation
	navigation *NavigationManager

	// Data
	ec2Instances       []EC2Instance
	filteredEC2        []EC2Instance
	eksClusters        []EKSCluster
	filteredEKS        []EKSCluster
	asgs               []ASG
	filteredASGs       []ASG
	nodeGroups         []NodeGroup
	filteredNodeGroups []NodeGroup
	netInterfaces      []aws.InstanceInterfaces
	filteredNetworks   []aws.InstanceInterfaces

	// Dashboard menu items
	menuItems []MenuItem

	// Post-exit actions
	pendingSSMSession *string // Instance ID to connect to after TUI exits
	pendingEKSCluster *string // Cluster name to show details for after TUI exits

	// Transient UI state
	searchActive     bool
	searchBuffer     string
	searchQueries    map[ViewMode]string
	searchDebounce   *time.Timer
	scaling          *ScalingState
	ltUpdate         *LaunchTemplateUpdateState
	statusMessage    string
	statusAnimation  *StatusAnimation
}

// NewModel creates a new TUI model
func NewModel(ctx context.Context, client *aws.Client, config Config) Model {
	// Initialize theme based on config
	SetTheme(NewModernTheme(!config.NoColor))
	
	// Initialize navigation manager
	navigation := NewNavigationManager()
	
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
			Title:       "EKS Node Groups",
			Description: "Inspect managed node groups",
			View:        ViewNodeGroups,
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
    s.Style = LoadingStyle()

	return Model{
		ctx:           ctx,
		client:        client,
		config:        config,
		currentView:   ViewDashboard,
		viewStack:     []ViewMode{},
		cursor:        0,
		menuItems:     menuItems,
		spinner:       s,
		searchQueries: map[ViewMode]string{},
		navigation:    navigation,
	}
}

// Init initializes the model
func (m Model) Init() tea.Cmd {
	return m.spinner.Tick
}

// Update handles messages and updates the model
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	
	// Update status animation
	m.updateStatusAnimation()

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.ready = true
		return m, nil

	case tea.KeyMsg:
		if m.scaling != nil {
			return m.handleScalingKeys(msg)
		}
		if m.ltUpdate != nil {
			return m.handleLaunchTemplateKeys(msg)
		}
		if m.searchActive {
			var handled bool
			var searchCmd tea.Cmd
			m, searchCmd, handled = m.handleSearchInput(msg)
			if handled {
				return m, searchCmd
			}
		}
		return m.handleKeyPress(msg)

	case DataLoadedMsg:
		return m.handleDataLoaded(msg)

	case ErrorMsg:
		m.err = msg.Err
		m.loading = false
		return m, nil

	case ScalingResultMsg:
		return m.handleScalingResult(msg)

	case LaunchTemplateVersionsMsg:
		return m.handleLaunchTemplateVersions(msg)

	case LaunchTemplateUpdateResultMsg:
		return m.handleLaunchTemplateUpdateResult(msg)

	case SearchDebounceMsg:
		// Apply filters after debounce delay
		m.cursor = 0
		m = m.applyFiltersForView(msg.View)
		return m, nil

	case AnimationMsg:
		// Handle animation messages
		m.startAnimation(msg.AnimationType, msg.Target)
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
	case ViewNodeGroups:
		return m.renderNodeGroups()
	case ViewNetworkInterfaces:
		return m.renderNetworkInterfaces()
	case ViewHelp:
		return m.renderHelp()
	default:
		return "Unknown view"
	}
}

// handleKeyPress handles keyboard input using the navigation manager
func (m Model) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Use navigation manager to handle the key press
	navAction := m.navigation.HandleKey(msg, m.currentView)
	
	switch navAction {
	case NavQuit:
		// Handle quit with confirmation
		if m.currentView == ViewDashboard {
			// From dashboard, quit immediately
			return m, tea.Quit
		} else {
			// From other views, show confirmation on first 'q'
			if strings.HasPrefix(m.statusMessage, "press q again") {
				// Second 'q' quits
				return m, tea.Quit
			} else {
				// First 'q' shows confirmation
				m.statusMessage = "press q again to quit, esc to stay"
				return m, nil
			}
		}

	case NavHelp:
		// Toggle help
		if m.currentView == ViewHelp {
			return m.navigateBack(), nil
		}
		m.pushView(ViewHelp)
		return m, nil

	case NavSearch:
		if !m.searchActive {
			m = m.beginSearch()
			return m, nil
		}
		
	case NavBack:
		return m.navigateBack(), nil
		
	case NavRefresh:
		// Handle refresh based on current view
		return m.handleRefresh()
	}

	// Handle navigation actions based on current view
	switch m.currentView {
	case ViewDashboard:
		return m.handleDashboardNavigation(navAction)
	case ViewEC2Instances:
		return m.handleEC2Navigation(navAction)
	case ViewEKSClusters:
		return m.handleEKSNavigation(navAction)
	case ViewASGs:
		return m.handleASGNavigation(navAction)
	case ViewNodeGroups:
		return m.handleNodeGroupNavigation(navAction)
	case ViewNetworkInterfaces:
		return m.handleNetworkInterfaceNavigation(navAction)
	case ViewHelp:
		return m.handleHelpNavigation(navAction)
	}

	return m, nil
}

// pushView pushes the current view onto the stack and switches to a new view
func (m *Model) pushView(view ViewMode) {
	m.viewStack = append(m.viewStack, m.currentView)
	m.currentView = view
	m.cursor = 0 // Reset cursor when changing views
	m.scaling = nil
	m.statusMessage = ""
}

// navigateBack pops the previous view from the stack
func (m Model) navigateBack() Model {
	if len(m.viewStack) > 0 {
		m.currentView = m.viewStack[len(m.viewStack)-1]
		m.viewStack = m.viewStack[:len(m.viewStack)-1]
		m.cursor = 0
		m.scaling = nil
		m.statusMessage = ""
	}
	return m
}

// startAnimation starts a new animation
func (m *Model) startAnimation(animationType AnimationType, target string) {
	// For now, we'll focus on status message animations
	if target == "status" && m.statusAnimation != nil {
		m.statusAnimation.Animation = NewAnimation(animationType, 300*time.Millisecond)
	}
}

// setStatusMessage sets a status message with optional animation
func (m *Model) setStatusMessage(message string, messageType string) {
	m.statusMessage = message
	if message != "" {
		m.statusAnimation = NewStatusAnimation(message, messageType, 2*time.Second)
	} else {
		m.statusAnimation = nil
	}
}

// updateStatusAnimation updates the status animation state
func (m *Model) updateStatusAnimation() {
	if m.statusAnimation != nil {
		completed, message := m.statusAnimation.Update()
		if completed {
			m.statusMessage = ""
			m.statusAnimation = nil
		} else {
			m.statusMessage = message
		}
	}
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
	case ViewNodeGroups:
		m.nodeGroups = msg.NodeGroups
	case ViewNetworkInterfaces:
		m.netInterfaces = msg.NetworkInstances
	}

	m = m.applyFiltersForView(msg.View)

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

// scheduleSSMSession schedules an SSM session to start after the TUI exits
func (m *Model) scheduleSSMSession(instanceID string) tea.Cmd {
	m.pendingSSMSession = &instanceID
	return tea.Quit
}

// scheduleClusterDetails schedules cluster details display after the TUI exits
func (m *Model) scheduleClusterDetails(clusterName string) tea.Cmd {
	m.pendingEKSCluster = &clusterName
	return tea.Quit
}

// Navigation handler methods

// handleDashboardNavigation handles navigation actions for dashboard
func (m Model) handleDashboardNavigation(action NavigationKey) (tea.Model, tea.Cmd) {
	switch action {
	case NavUp:
		if m.cursor > 0 {
			m.cursor--
		}
	case NavDown:
		if m.cursor < len(m.menuItems)-1 {
			m.cursor++
		}
	case NavSelect:
		// Navigate to selected view
		selectedItem := m.menuItems[m.cursor]
		m.pushView(selectedItem.View)

		// Load data for the selected view
		var cmd tea.Cmd
		switch selectedItem.View {
		case ViewEC2Instances:
			m.loading = true
			m.loadingMsg = "Loading EC2 instances..."
			cmd = LoadEC2InstancesCmd(m.ctx, m.client)
		case ViewEKSClusters:
			m.loading = true
			m.loadingMsg = "Loading EKS clusters..."
			cmd = LoadEKSClustersCmd(m.ctx, m.client)
		case ViewASGs:
			m.loading = true
			m.loadingMsg = "Loading Auto Scaling Groups..."
			cmd = LoadASGsCmd(m.ctx, m.client)
		case ViewNodeGroups:
			m.loading = true
			m.loadingMsg = "Loading EKS node groups..."
			cmd = LoadNodeGroupsCmd(m.ctx, m.client)
		case ViewNetworkInterfaces:
			m.loading = true
			m.loadingMsg = "Loading network interfaces..."
			cmd = LoadNetworkInterfacesCmd(m.ctx, m.client)
		}
		return m, cmd
	}
	return m, nil
}

// handleRefresh handles refresh based on current view
func (m Model) handleRefresh() (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	
	switch m.currentView {
	case ViewEC2Instances:
		m.loading = true
		m.loadingMsg = "Refreshing EC2 instances..."
		cmd = LoadEC2InstancesCmd(m.ctx, m.client)
	case ViewEKSClusters:
		m.loading = true
		m.loadingMsg = "Refreshing EKS clusters..."
		cmd = LoadEKSClustersCmd(m.ctx, m.client)
	case ViewASGs:
		m.loading = true
		m.loadingMsg = "Refreshing Auto Scaling Groups..."
		cmd = LoadASGsCmd(m.ctx, m.client)
	case ViewNodeGroups:
		m.loading = true
		m.loadingMsg = "Refreshing EKS node groups..."
		cmd = LoadNodeGroupsCmd(m.ctx, m.client)
	case ViewNetworkInterfaces:
		m.loading = true
		m.loadingMsg = "Refreshing network interfaces..."
		cmd = LoadNetworkInterfacesCmd(m.ctx, m.client)
	default:
		m.statusMessage = "Refresh not available for this view"
	}
	
	return m, cmd
}

// handleEC2Navigation handles navigation actions for EC2 instances
func (m Model) handleEC2Navigation(action NavigationKey) (tea.Model, tea.Cmd) {
	instances := m.getEC2Instances()
	
	switch action {
	case NavUp:
		if m.cursor > 0 {
			m.cursor--
		}
	case NavDown:
		if m.cursor < len(instances)-1 {
			m.cursor++
		}
	case NavHome:
		m.cursor = 0
	case NavEnd:
		m.cursor = len(instances) - 1
	case NavPageUp:
		m.cursor = max(0, m.cursor-10)
	case NavPageDown:
		m.cursor = min(len(instances)-1, m.cursor+10)
	case NavSSH:
		if m.cursor < len(instances) {
			inst := instances[m.cursor]
			// Check if instance is running
			if inst.State != "running" {
				m.err = fmt.Errorf("instance %s is not running (state: %s)", inst.Name, inst.State)
				return m, nil
			}
			// Schedule SSM session to start after TUI exits
			return m, m.scheduleSSMSession(inst.InstanceID)
		}
	case NavDetails:
		if m.cursor < len(instances) {
			m.statusMessage = fmt.Sprintf("Instance %s details: %s", 
				instances[m.cursor].InstanceID, instances[m.cursor].State)
		}
	case NavScale:
		if m.cursor < len(instances) {
			m.statusMessage = "EC2 instance scaling not implemented yet"
		}
	case NavFilter:
		m.statusMessage = "Filter by: running, stopped, terminated"
	}
	return m, nil
}

// handleEKSNavigation handles navigation actions for EKS clusters
func (m Model) handleEKSNavigation(action NavigationKey) (tea.Model, tea.Cmd) {
	clusters := m.getEKSClusters()
	
	switch action {
	case NavUp:
		if m.cursor > 0 {
			m.cursor--
		}
	case NavDown:
		if m.cursor < len(clusters)-1 {
			m.cursor++
		}
	case NavHome:
		m.cursor = 0
	case NavEnd:
		m.cursor = len(clusters) - 1
	case NavPageUp:
		m.cursor = max(0, m.cursor-10)
	case NavPageDown:
		m.cursor = min(len(clusters)-1, m.cursor+10)
	case NavSelect:
		if m.cursor < len(clusters) {
			clusterName := clusters[m.cursor].Name
			m.pushView(ViewNodeGroups)
			m.loading = true
			m.loadingMsg = fmt.Sprintf("Loading node groups for %s...", clusterName)
			return m, LoadNodeGroupsCmd(m.ctx, m.client)
		}
	case NavDetails:
		if m.cursor < len(clusters) {
			m.statusMessage = fmt.Sprintf("Cluster %s details: %s", 
				clusters[m.cursor].Name, clusters[m.cursor].Status)
		}
	}
	return m, nil
}

// handleASGNavigation handles navigation actions for ASGs
func (m Model) handleASGNavigation(action NavigationKey) (tea.Model, tea.Cmd) {
	asgs := m.getASGs()
	
	switch action {
	case NavUp:
		if m.cursor > 0 {
			m.cursor--
		}
	case NavDown:
		if m.cursor < len(asgs)-1 {
			m.cursor++
		}
	case NavHome:
		m.cursor = 0
	case NavEnd:
		m.cursor = len(asgs) - 1
	case NavPageUp:
		m.cursor = max(0, m.cursor-10)
	case NavPageDown:
		m.cursor = min(len(asgs)-1, m.cursor+10)
	case NavScale:
		if m.cursor < len(asgs) {
			asgs := m.getASGs()
			m = m.startASGScaling(asgs[m.cursor])
		}
	case NavDetails:
		if m.cursor < len(asgs) {
			m.statusMessage = fmt.Sprintf("ASG %s details: %d instances", 
				asgs[m.cursor].Name, asgs[m.cursor].DesiredCapacity)
		}
	}
	return m, nil
}

// handleNodeGroupNavigation handles navigation actions for node groups
func (m Model) handleNodeGroupNavigation(action NavigationKey) (tea.Model, tea.Cmd) {
	groups := m.getNodeGroups()
	
	switch action {
	case NavUp:
		if m.cursor > 0 {
			m.cursor--
		}
	case NavDown:
		if m.cursor < len(groups)-1 {
			m.cursor++
		}
	case NavHome:
		m.cursor = 0
	case NavEnd:
		m.cursor = len(groups) - 1
	case NavPageUp:
		m.cursor = max(0, m.cursor-10)
	case NavPageDown:
		m.cursor = min(len(groups)-1, m.cursor+10)
	case NavScale:
		if m.cursor < len(groups) {
			groups := m.getNodeGroups()
			m = m.startNodeGroupScaling(groups[m.cursor])
		}
	case NavSelect:
		if m.cursor < len(groups) {
			groups := m.getNodeGroups()
			var cmd tea.Cmd
			m, cmd = m.startNodeGroupLaunchTemplateUpdate(groups[m.cursor])
			return m, cmd
		}
	case NavDetails:
		if m.cursor < len(groups) {
			m.statusMessage = fmt.Sprintf("Node group %s details: %s", 
				groups[m.cursor].Name, groups[m.cursor].Status)
		}
	}
	return m, nil
}

// handleNetworkInterfaceNavigation handles navigation actions for network interfaces
func (m Model) handleNetworkInterfaceNavigation(action NavigationKey) (tea.Model, tea.Cmd) {
	interfaces := m.getNetworkInterfaces()
	
	switch action {
	case NavUp:
		if m.cursor > 0 {
			m.cursor--
		}
	case NavDown:
		if m.cursor < len(interfaces)-1 {
			m.cursor++
		}
	case NavHome:
		m.cursor = 0
	case NavEnd:
		m.cursor = len(interfaces) - 1
	case NavPageUp:
		m.cursor = max(0, m.cursor-10)
	case NavPageDown:
		m.cursor = min(len(interfaces)-1, m.cursor+10)
	case NavDetails:
		if m.cursor < len(interfaces) {
			iface := interfaces[m.cursor]
			m.statusMessage = fmt.Sprintf("Instance %s interfaces: %d total", 
				iface.InstanceID, len(iface.Interfaces))
		}
	}
	return m, nil
}

// handleHelpNavigation handles navigation actions for help view
func (m Model) handleHelpNavigation(action NavigationKey) (tea.Model, tea.Cmd) {
	switch action {
	case NavBack:
		return m.navigateBack(), nil
	case NavUp:
		// Future: scroll help content up
	case NavDown:
		// Future: scroll help content down
	case NavHome:
		// Future: go to top of help
	case NavEnd:
		// Future: go to bottom of help
	}
	return m, nil
}
