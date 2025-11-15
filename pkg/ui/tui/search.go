package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/johnlam90/aws-ssm/pkg/aws"
)

// beginSearch starts a search session for the current view
func (m Model) beginSearch() Model {
	if m.searchQueries == nil {
		m.searchQueries = make(map[ViewMode]string)
	}
	m.searchActive = true
	m.searchBuffer = m.searchQueries[m.currentView]
	return m
}

// handleSearchInput processes key events while search is active
func (m Model) handleSearchInput(msg tea.KeyMsg) (Model, bool) {
	if !m.searchActive {
		return m, false
	}

	switch msg.Type {
	case tea.KeyEsc:
		m.searchActive = false
		m.searchBuffer = ""
		m = m.applyFiltersForView(m.currentView)
		return m, true
	case tea.KeyEnter:
		query := strings.TrimSpace(m.searchBuffer)
		if query == "" {
			delete(m.searchQueries, m.currentView)
		} else {
			m.searchQueries[m.currentView] = query
		}
		m.searchActive = false
		m.searchBuffer = ""
		m.cursor = 0
		m = m.applyFiltersForView(m.currentView)
		return m, true
	case tea.KeyBackspace:
		if len(m.searchBuffer) > 0 {
			m.searchBuffer = m.searchBuffer[:len(m.searchBuffer)-1]
			m.cursor = 0
			m = m.applyFiltersForView(m.currentView)
		}
		return m, true
	case tea.KeyCtrlU:
		m.searchBuffer = ""
		m.cursor = 0
		m = m.applyFiltersForView(m.currentView)
		return m, true
	}

	input := msg.String()
	if len(input) == 1 && !msg.Alt {
		m.searchBuffer += input
		m.cursor = 0
		m = m.applyFiltersForView(m.currentView)
		return m, true
	}

	return m, false
}

// applyFiltersForView filters data for a specific view using its search query
func (m Model) applyFiltersForView(view ViewMode) Model {
	query := strings.ToLower(strings.TrimSpace(m.getSearchQuery(view)))

	switch view {
	case ViewEC2Instances:
		if query == "" {
			m.filteredEC2 = nil
		} else {
			var filtered []EC2Instance
			for _, inst := range m.ec2Instances {
				if ec2MatchesQuery(inst, query) {
					filtered = append(filtered, inst)
				}
			}
			m.filteredEC2 = filtered
		}
	case ViewEKSClusters:
		if query == "" {
			m.filteredEKS = nil
		} else {
			var filtered []EKSCluster
			for _, cluster := range m.eksClusters {
				if eksMatchesQuery(cluster, query) {
					filtered = append(filtered, cluster)
				}
			}
			m.filteredEKS = filtered
		}
	case ViewASGs:
		if query == "" {
			m.filteredASGs = nil
		} else {
			var filtered []ASG
			for _, asg := range m.asgs {
				if asgMatchesQuery(asg, query) {
					filtered = append(filtered, asg)
				}
			}
			m.filteredASGs = filtered
		}
	case ViewNodeGroups:
		if query == "" {
			m.filteredNodeGroups = nil
		} else {
			var filtered []NodeGroup
			for _, ng := range m.nodeGroups {
				if nodeGroupMatchesQuery(ng, query) {
					filtered = append(filtered, ng)
				}
			}
			m.filteredNodeGroups = filtered
		}
	case ViewNetworkInterfaces:
		if query == "" {
			m.filteredNetworks = nil
		} else {
			var filtered []aws.InstanceInterfaces
			for _, inst := range m.netInterfaces {
				if networkMatchesQuery(inst, query) {
					filtered = append(filtered, inst)
				}
			}
			m.filteredNetworks = filtered
		}
	}

	return m.ensureCursorInBounds(view)
}

// getSearchQuery returns the active search query for a view
func (m Model) getSearchQuery(view ViewMode) string {
	if m.searchActive && m.currentView == view {
		return m.searchBuffer
	}
	if m.searchQueries == nil {
		return ""
	}
	return m.searchQueries[view]
}

// ensureCursorInBounds enforces cursor limits for a view
func (m Model) ensureCursorInBounds(view ViewMode) Model {
	length := m.viewLength(view)
	if length == 0 {
		m.cursor = 0
		return m
	}
	if m.cursor >= length {
		m.cursor = length - 1
	}
	if m.cursor < 0 {
		m.cursor = 0
	}
	return m
}

// viewLength returns the visible list length for a view
func (m Model) viewLength(view ViewMode) int {
	switch view {
	case ViewEC2Instances:
		return len(m.getEC2Instances())
	case ViewEKSClusters:
		return len(m.getEKSClusters())
	case ViewASGs:
		return len(m.getASGs())
	case ViewNodeGroups:
		return len(m.getNodeGroups())
	case ViewNetworkInterfaces:
		return len(m.getNetworkInterfaces())
	default:
		return 0
	}
}

// renderSearchBar renders the search prompt for a view
func (m Model) renderSearchBar(view ViewMode) string {
	query := strings.TrimSpace(m.getSearchQuery(view))
	if m.searchActive && m.currentView == view {
		query = m.searchBuffer
	}

	if !m.searchActive && query == "" {
		return ""
	}

	status := "Search"
	if m.searchActive && m.currentView == view {
		status = "Search (enter to apply, esc to cancel)"
	}

	display := strings.TrimSpace(query)
	if m.searchActive && m.currentView == view {
		display = query + "▍"
	}
	if display == "" {
		display = "(type to filter)"
	}

	return HelpStyle.Render(fmt.Sprintf("%s › %s", status, display))
}

// getEC2Instances returns the visible EC2 instances (filtered or not)
func (m Model) getEC2Instances() []EC2Instance {
	if strings.TrimSpace(m.getSearchQuery(ViewEC2Instances)) == "" {
		return m.ec2Instances
	}
	return m.filteredEC2
}

// getEKSClusters returns the visible EKS clusters
func (m Model) getEKSClusters() []EKSCluster {
	if strings.TrimSpace(m.getSearchQuery(ViewEKSClusters)) == "" {
		return m.eksClusters
	}
	return m.filteredEKS
}

// getASGs returns the visible ASGs
func (m Model) getASGs() []ASG {
	if strings.TrimSpace(m.getSearchQuery(ViewASGs)) == "" {
		return m.asgs
	}
	return m.filteredASGs
}

// getNodeGroups returns the visible node groups
func (m Model) getNodeGroups() []NodeGroup {
	if strings.TrimSpace(m.getSearchQuery(ViewNodeGroups)) == "" {
		return m.nodeGroups
	}
	return m.filteredNodeGroups
}

// getNetworkInterfaces returns the visible network interface entries
func (m Model) getNetworkInterfaces() []aws.InstanceInterfaces {
	if strings.TrimSpace(m.getSearchQuery(ViewNetworkInterfaces)) == "" {
		return m.netInterfaces
	}
	return m.filteredNetworks
}

func ec2MatchesQuery(inst EC2Instance, query string) bool {
	name := strings.ToLower(inst.Name)
	instanceID := strings.ToLower(inst.InstanceID)
	privateIP := strings.ToLower(inst.PrivateIP)
	publicIP := strings.ToLower(inst.PublicIP)
	instanceType := strings.ToLower(inst.InstanceType)
	state := strings.ToLower(inst.State)

	if strings.Contains(name, query) ||
		strings.Contains(instanceID, query) ||
		strings.Contains(privateIP, query) ||
		strings.Contains(publicIP, query) ||
		strings.Contains(instanceType, query) ||
		strings.Contains(state, query) {
		return true
	}

	for k, v := range inst.Tags {
		tag := strings.ToLower(fmt.Sprintf("%s:%s", k, v))
		if strings.Contains(tag, query) {
			return true
		}
	}

	return false
}

func eksMatchesQuery(cluster EKSCluster, query string) bool {
	name := strings.ToLower(cluster.Name)
	status := strings.ToLower(cluster.Status)
	version := strings.ToLower(cluster.Version)
	arn := strings.ToLower(cluster.Arn)

	return strings.Contains(name, query) ||
		strings.Contains(status, query) ||
		strings.Contains(version, query) ||
		strings.Contains(arn, query)
}

func asgMatchesQuery(asg ASG, query string) bool {
	name := strings.ToLower(asg.Name)
	status := strings.ToLower(asg.Status)

	if strings.Contains(name, query) || strings.Contains(status, query) {
		return true
	}

	capacityFields := []int32{asg.DesiredCapacity, asg.MinSize, asg.MaxSize, asg.CurrentSize}
	for _, val := range capacityFields {
		if strings.Contains(fmt.Sprint(val), query) {
			return true
		}
	}

	return false
}

func nodeGroupMatchesQuery(ng NodeGroup, query string) bool {
	cluster := strings.ToLower(ng.ClusterName)
	name := strings.ToLower(ng.Name)
	status := strings.ToLower(ng.Status)
	version := strings.ToLower(ng.Version)
	instanceTypes := strings.ToLower(strings.Join(ng.InstanceTypes, ","))
	launchTemplateName := strings.ToLower(ng.LaunchTemplateName)
	launchTemplateVersion := strings.ToLower(ng.LaunchTemplateVersion)
	launchTemplateID := strings.ToLower(ng.LaunchTemplateID)

	if strings.Contains(cluster, query) ||
		strings.Contains(name, query) ||
		strings.Contains(status, query) ||
		strings.Contains(version, query) ||
		strings.Contains(instanceTypes, query) ||
		strings.Contains(launchTemplateName, query) ||
		strings.Contains(launchTemplateVersion, query) ||
		strings.Contains(launchTemplateID, query) {
		return true
	}

	sizeFields := []int32{ng.DesiredSize, ng.MinSize, ng.MaxSize, ng.CurrentSize}
	for _, val := range sizeFields {
		if strings.Contains(fmt.Sprint(val), query) {
			return true
		}
	}

	return false
}

func networkMatchesQuery(inst aws.InstanceInterfaces, query string) bool {
	instanceFields := []string{
		strings.ToLower(inst.InstanceName),
		strings.ToLower(inst.InstanceID),
		strings.ToLower(inst.DNSName),
	}

	for _, field := range instanceFields {
		if strings.Contains(field, query) {
			return true
		}
	}

	for _, iface := range inst.Interfaces {
		if strings.Contains(strings.ToLower(iface.InterfaceName), query) ||
			strings.Contains(strings.ToLower(iface.SubnetID), query) ||
			strings.Contains(strings.ToLower(iface.CIDR), query) ||
			strings.Contains(strings.ToLower(iface.SecurityGroup), query) {
			return true
		}
	}

	return false
}
