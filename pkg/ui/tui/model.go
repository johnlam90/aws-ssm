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
	selectedItems      map[ViewMode]string

	// Dashboard menu items
	menuItems []MenuItem

	// Post-exit actions
	pendingSSMSession *string // Instance ID to connect to after TUI exits
	pendingEKSCluster *string // Cluster name to show details for after TUI exits

	// Transient UI state
	searchActive    bool
	searchBuffer    string
	searchQueries   map[ViewMode]string
	searchDebounce  *time.Timer
	scaling         *ScalingState
	ltUpdate        *LaunchTemplateUpdateState
	statusMessage   string
	statusAnimation *StatusAnimation
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
		selectedItems: map[ViewMode]string{},
	}
}

// Init initializes the model
func (m Model) Init() tea.Cmd {
	return m.spinner.Tick
}

// Update handles messages and updates the model
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	m.updateStatusAnimation()
	switch x := msg.(type) {
	case tea.KeyMsg:
		return m.updateKeyMsg(x)
	default:
		return m.updateNonKeyMsg(x)
	}
}

func (m Model) updateKeyMsg(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.scaling != nil {
		return m.handleScalingKeys(msg)
	}
	if m.ltUpdate != nil {
		return m.handleLaunchTemplateKeys(msg)
	}
	if m.searchActive {
		updated, searchCmd, handled := m.handleSearchInput(msg)
		if handled {
			return updated, searchCmd
		}
	}
	if msg.Type == tea.KeyEsc && !m.searchActive && strings.TrimSpace(m.getSearchQuery(m.currentView)) != "" {
		m = m.clearSearchQuery(m.currentView)
		return m, nil
	}
	if updated, cmd, handled := m.handleNodeGroupLaunchTemplateShortcut(msg); handled {
		return updated, cmd
	}
	return m.handleKeyPress(msg)
}

func (m Model) updateNonKeyMsg(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch v := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = v.Width
		m.height = v.Height
		m.ready = true
		return m, nil
	case DataLoadedMsg:
		return m.handleDataLoaded(v)
	case ErrorMsg:
		m.err = v.Err
		m.loading = false
		return m, nil
	case ScalingResultMsg:
		return m.handleScalingResult(v)
	case LaunchTemplateVersionsMsg:
		return m.handleLaunchTemplateVersions(v)
	case LaunchTemplateUpdateResultMsg:
		return m.handleLaunchTemplateUpdateResult(v)
	case SearchDebounceMsg:
		m.cursor = 0
		m = m.applyFiltersForView(v.View)
		return m, nil
	case AnimationMsg:
		m.startAnimation(v.AnimationType, v.Target)
		return m, nil
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(v)
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
		if m.currentView == ViewDashboard {
			return m, tea.Quit
		}
		if strings.HasPrefix(m.statusMessage, "press q again") {
			return m, tea.Quit
		}
		m.statusMessage = "press q again to quit, esc to stay"
		return m, nil

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

	return m.handleNavigation(navAction)
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
	if msg.View == m.currentView {
		m = m.restoreSelection(msg.View)
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

// scheduleSSMSession schedules an SSM session to start after the TUI exits
func (m *Model) scheduleSSMSession(instanceID string) tea.Cmd {
	m.pendingSSMSession = &instanceID
	return tea.Quit
}

// scheduleClusterDetails schedules cluster details display after the TUI exits

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
	m.captureSelection(m.currentView)

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
	case NavUp, NavDown, NavHome, NavEnd, NavPageUp, NavPageDown:
		return m.applyCursorMovement(len(instances), action), nil
	case NavSSH:
		return m.ec2Connect(instances)
	case NavDetails:
		return m.ec2Details(instances), nil
	case NavScale:
		return m.ec2ScaleNotice(), nil
	case NavFilter:
		return m.ec2FilterHint(), nil
	}
	return m, nil
}

func (m Model) applyCursorMovement(length int, action NavigationKey) Model {
	switch action {
	case NavUp:
		if m.cursor > 0 {
			m.cursor--
		}
	case NavDown:
		if m.cursor < length-1 {
			m.cursor++
		}
	case NavHome:
		m.cursor = 0
	case NavEnd:
		m.cursor = length - 1
	case NavPageUp:
		m.cursor = max(0, m.cursor-10)
	case NavPageDown:
		m.cursor = min(length-1, m.cursor+10)
	}
	return m
}

func (m Model) ec2Connect(instances []EC2Instance) (tea.Model, tea.Cmd) {
	if m.cursor >= len(instances) {
		return m, nil
	}
	inst := instances[m.cursor]
	if inst.State != "running" {
		m.err = fmt.Errorf("instance %s is not running (state: %s)", inst.Name, inst.State)
		return m, nil
	}
	return m, m.scheduleSSMSession(inst.InstanceID)
}

func (m Model) ec2Details(instances []EC2Instance) Model {
	if m.cursor < len(instances) {
		m.statusMessage = fmt.Sprintf("Instance %s details: %s", instances[m.cursor].InstanceID, instances[m.cursor].State)
	}
	return m
}

func (m Model) ec2ScaleNotice() Model {
	if m.cursor >= 0 {
		m.statusMessage = "EC2 instance scaling not implemented yet"
	}
	return m
}

func (m Model) ec2FilterHint() Model {
	m.statusMessage = "Filter by: running, stopped, terminated"
	return m
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

// handleNodeGroupLaunchTemplateShortcut catches the legacy 'u' shortcut directly
func (m Model) handleNodeGroupLaunchTemplateShortcut(msg tea.KeyMsg) (tea.Model, tea.Cmd, bool) {
	if m.currentView != ViewNodeGroups {
		return m, nil, false
	}

	switch msg.String() {
	case "u", "U":
		groups := m.getNodeGroups()
		if len(groups) == 0 || m.cursor >= len(groups) {
			return m, nil, true
		}
		var cmd tea.Cmd
		m, cmd = m.startNodeGroupLaunchTemplateUpdate(groups[m.cursor])
		return m, cmd, true
	default:
		return m, nil, false
	}
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
func (m Model) handleNavigation(action NavigationKey) (tea.Model, tea.Cmd) {
	handlers := map[ViewMode]func(Model, NavigationKey) (tea.Model, tea.Cmd){
		ViewDashboard:         Model.handleDashboardNavigation,
		ViewEC2Instances:      Model.handleEC2Navigation,
		ViewEKSClusters:       Model.handleEKSNavigation,
		ViewASGs:              Model.handleASGNavigation,
		ViewNodeGroups:        Model.handleNodeGroupNavigation,
		ViewNetworkInterfaces: Model.handleNetworkInterfaceNavigation,
		ViewHelp:              Model.handleHelpNavigation,
	}
	if h, ok := handlers[m.currentView]; ok {
		return h(m, action)
	}
	return m, nil
}

// Selection helpers
func (m Model) currentSelectionKey(view ViewMode) string {
	switch view {
	case ViewEC2Instances:
		return m.getEC2SelectionKey()
	case ViewEKSClusters:
		return m.getEKSSelectionKey()
	case ViewASGs:
		return m.getASGSelectionKey()
	case ViewNodeGroups:
		return m.getNodeGroupSelectionKey()
	case ViewNetworkInterfaces:
		return m.getNISelectionKey()
	default:
		return ""
	}
}

func (m Model) getEC2SelectionKey() string {
	items := m.getEC2Instances()
	if m.cursor < 0 || m.cursor >= len(items) {
		return ""
	}
	return items[m.cursor].InstanceID
}

func (m Model) getEKSSelectionKey() string {
	items := m.getEKSClusters()
	if m.cursor < 0 || m.cursor >= len(items) {
		return ""
	}
	return items[m.cursor].Name
}

func (m Model) getASGSelectionKey() string {
	items := m.getASGs()
	if m.cursor < 0 || m.cursor >= len(items) {
		return ""
	}
	return items[m.cursor].Name
}

func (m Model) getNodeGroupSelectionKey() string {
	items := m.getNodeGroups()
	if m.cursor < 0 || m.cursor >= len(items) {
		return ""
	}
	return fmt.Sprintf("%s|%s", items[m.cursor].ClusterName, items[m.cursor].Name)
}

func (m Model) getNISelectionKey() string {
	items := m.getNetworkInterfaces()
	if m.cursor < 0 || m.cursor >= len(items) {
		return ""
	}
	return items[m.cursor].InstanceID
}

func (m Model) findSelectionIndex(view ViewMode, key string) int {
	key = strings.TrimSpace(key)
	if key == "" {
		return -1
	}
	switch view {
	case ViewEC2Instances:
		return m.findEC2Index(key)
	case ViewEKSClusters:
		return m.findEKSIndex(key)
	case ViewASGs:
		return m.findASGIndex(key)
	case ViewNodeGroups:
		return m.findNodeGroupIndex(key)
	case ViewNetworkInterfaces:
		return m.findNIIndex(key)
	default:
		return -1
	}
}

func (m Model) findEC2Index(key string) int {
	for i, inst := range m.getEC2Instances() {
		if inst.InstanceID == key {
			return i
		}
	}
	return -1
}

func (m Model) findEKSIndex(key string) int {
	for i, cluster := range m.getEKSClusters() {
		if cluster.Name == key {
			return i
		}
	}
	return -1
}

func (m Model) findASGIndex(key string) int {
	for i, asg := range m.getASGs() {
		if asg.Name == key {
			return i
		}
	}
	return -1
}

func (m Model) findNodeGroupIndex(key string) int {
	for i, ng := range m.getNodeGroups() {
		if fmt.Sprintf("%s|%s", ng.ClusterName, ng.Name) == key {
			return i
		}
	}
	return -1
}

func (m Model) findNIIndex(key string) int {
	for i, ni := range m.getNetworkInterfaces() {
		if ni.InstanceID == key {
			return i
		}
	}
	return -1
}

func (m *Model) captureSelection(view ViewMode) {
	if m.selectedItems == nil {
		m.selectedItems = map[ViewMode]string{}
	}
	key := strings.TrimSpace(m.currentSelectionKey(view))
	if key == "" {
		delete(m.selectedItems, view)
		return
	}
	m.selectedItems[view] = key
}

func (m Model) restoreSelection(view ViewMode) Model {
	if m.currentView != view {
		return m
	}
	if m.selectedItems == nil {
		return m.ensureCursorInBounds(view)
	}
	key := strings.TrimSpace(m.selectedItems[view])
	if key == "" {
		return m.ensureCursorInBounds(view)
	}
	if idx := m.findSelectionIndex(view, key); idx >= 0 {
		m.cursor = idx
		return m
	}
	return m.ensureCursorInBounds(view)
}
