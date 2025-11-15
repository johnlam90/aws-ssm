package tui

import (
    "fmt"
    "strings"
    "time"

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
func (m Model) handleSearchInput(msg tea.KeyMsg) (Model, tea.Cmd, bool) {
	if !m.searchActive {
		return m, nil, false
	}

	switch msg.Type {
	case tea.KeyEsc:
		m.searchActive = false
		m.searchBuffer = ""
		m.cancelSearchDebounce()
		m = m.applyFiltersForView(m.currentView)
		return m, nil, true
	case tea.KeyEnter:
		query := strings.TrimSpace(m.searchBuffer)
		if query == "" {
			delete(m.searchQueries, m.currentView)
		} else {
			m.searchQueries[m.currentView] = query
		}
		m.searchActive = false
		m.searchBuffer = ""
		m.cancelSearchDebounce()
		m.cursor = 0
		m = m.applyFiltersForView(m.currentView)
		return m, nil, true
	case tea.KeyBackspace:
		if len(m.searchBuffer) > 0 {
			m.searchBuffer = m.searchBuffer[:len(m.searchBuffer)-1]
			m.cursor = 0
			return m, m.scheduleSearchDebounce(), true
		}
		return m, nil, true
	case tea.KeyCtrlU:
		m.searchBuffer = ""
		m.cursor = 0
		return m, m.scheduleSearchDebounce(), true
	}

	input := msg.String()
	if len(input) == 1 && !msg.Alt {
		m.searchBuffer += input
		m.cursor = 0
		return m, m.scheduleSearchDebounce(), true
	}

	return m, nil, false
}

// scheduleSearchDebounce schedules a debounced search operation
func (m *Model) scheduleSearchDebounce() tea.Cmd {
	// Cancel any existing debounce timer
	m.cancelSearchDebounce()
	
	// Return a command that will send the debounce message after delay
	return func() tea.Msg {
		time.Sleep(150 * time.Millisecond)
		return SearchDebounceMsg{
			View:  m.currentView,
			Query: m.searchBuffer,
		}
	}
}

// cancelSearchDebounce cancels any pending search debounce
func (m *Model) cancelSearchDebounce() {
	if m.searchDebounce != nil {
		m.searchDebounce.Stop()
		m.searchDebounce = nil
	}
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

    return HelpStyle().Render(fmt.Sprintf("%s › %s", status, display))
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
	query = strings.ToLower(strings.TrimSpace(query))
	
	// Parse tokens for key:value filtering
	tokens := parseTokens(query)
	if len(tokens) > 0 {
		// Token-based filtering
		for _, token := range tokens {
			key := strings.ToLower(token[0])
			value := strings.ToLower(token[1])
			
			switch key {
			case "name":
				if !strings.Contains(strings.ToLower(inst.Name), value) {
					return false
				}
			case "id", "instance":
				if !strings.Contains(strings.ToLower(inst.InstanceID), value) {
					return false
				}
			case "privateip", "private":
				if !strings.Contains(strings.ToLower(inst.PrivateIP), value) {
					return false
				}
			case "publicip", "public", "ip", "pip":
				if !strings.Contains(strings.ToLower(inst.PublicIP), value) && !strings.Contains(strings.ToLower(inst.PrivateIP), value) {
					return false
				}
			case "type":
				if !strings.Contains(strings.ToLower(inst.InstanceType), value) {
					return false
				}
			case "state":
				if !strings.Contains(strings.ToLower(inst.State), value) {
					return false
				}
            case "tag":
                // Use cached tags string for efficient searching
                tags := inst.cachedTagsString
                if tags == "" {
                    var pairs []string
                    for k, v := range inst.Tags {
                        pairs = append(pairs, strings.ToLower(fmt.Sprintf("%s:%s", k, v)))
                        pairs = append(pairs, strings.ToLower(fmt.Sprintf("%s=%s", k, v)))
                    }
                    tags = strings.Join(pairs, " ")
                }
                if strings.Contains(value, "=") || strings.Contains(value, ":") {
                    if !strings.Contains(tags, value) {
                        return false
                    }
                } else {
                    // No separator: treat as tag key presence
                    found := false
                    for k := range inst.Tags {
                        if strings.Contains(strings.ToLower(k), value) {
                            found = true
                            break
                        }
                    }
                    if !found {
                        return false
                    }
                }
			default:
				// Unknown key, treat as regular search
				if !strings.Contains(inst.cachedNameLower, key+":"+value) {
					return false
				}
			}
		}
		return true
	}
	
    // Fallback to regular text search using cached fields (compute if cache missing)
    nameLower := inst.cachedNameLower
    if nameLower == "" { nameLower = strings.ToLower(inst.Name) }
    idLower := inst.cachedIDLower
    if idLower == "" { idLower = strings.ToLower(inst.InstanceID) }
    privLower := inst.cachedPrivateIPLower
    if privLower == "" { privLower = strings.ToLower(inst.PrivateIP) }
    pubLower := inst.cachedPublicIPLower
    if pubLower == "" { pubLower = strings.ToLower(inst.PublicIP) }
    typeLower := inst.cachedTypeLower
    if typeLower == "" { typeLower = strings.ToLower(inst.InstanceType) }
    stateLower := inst.cachedStateLower
    if stateLower == "" { stateLower = strings.ToLower(inst.State) }
    tagsLower := inst.cachedTagsString
    if tagsLower == "" {
        var pairs []string
        for k, v := range inst.Tags {
            pairs = append(pairs, strings.ToLower(fmt.Sprintf("%s:%s", k, v)))
        }
        tagsLower = strings.Join(pairs, " ")
    }

    if strings.Contains(nameLower, query) ||
        strings.Contains(idLower, query) ||
        strings.Contains(privLower, query) ||
        strings.Contains(pubLower, query) ||
        strings.Contains(typeLower, query) ||
        strings.Contains(stateLower, query) ||
        strings.Contains(tagsLower, query) {
        return true
    }

	return false
}

// parseTokens parses key:value pairs from a search query
func parseTokens(query string) [][2]string {
	var tokens [][2]string
	parts := strings.Fields(query)
	
	for _, part := range parts {
		if strings.Contains(part, ":") {
			pair := strings.SplitN(part, ":", 2)
			if len(pair) == 2 && pair[0] != "" && pair[1] != "" {
				tokens = append(tokens, [2]string{pair[0], pair[1]})
			}
		}
	}
	
	return tokens
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
	query = strings.ToLower(strings.TrimSpace(query))
	
	// Parse tokens for key:value filtering
	tokens := parseTokens(query)
	if len(tokens) > 0 {
		// Token-based filtering
		for _, token := range tokens {
			key := strings.ToLower(token[0])
			value := strings.ToLower(token[1])
			
			switch key {
			case "cluster":
				if !strings.Contains(strings.ToLower(ng.ClusterName), value) {
					return false
				}
			case "name":
				if !strings.Contains(strings.ToLower(ng.Name), value) {
					return false
				}
			case "status":
				if !strings.Contains(strings.ToLower(ng.Status), value) {
					return false
				}
			case "version":
				if !strings.Contains(strings.ToLower(ng.Version), value) {
					return false
				}
			case "instancetype", "type":
				instanceTypes := strings.ToLower(strings.Join(ng.InstanceTypes, ","))
				if !strings.Contains(instanceTypes, value) {
					return false
				}
			case "ltname", "launchtemplatename":
				if !strings.Contains(strings.ToLower(ng.LaunchTemplateName), value) {
					return false
				}
			case "ltid", "launchtemplateid":
				if !strings.Contains(strings.ToLower(ng.LaunchTemplateID), value) {
					return false
				}
			case "ltversion", "launchtemplateversion":
				if !strings.Contains(strings.ToLower(ng.LaunchTemplateVersion), value) {
					return false
				}
			default:
				// Unknown key, treat as regular search
				if !strings.Contains(strings.ToLower(ng.Name), key+":"+value) {
					return false
				}
			}
		}
		return true
	}
	
    // Fallback to regular text search using cached fields (compute if cache missing)
    clusterLower := ng.cachedClusterLower
    if clusterLower == "" { clusterLower = strings.ToLower(ng.ClusterName) }
    nameLower := ng.cachedNameLower
    if nameLower == "" { nameLower = strings.ToLower(ng.Name) }
    statusLower := ng.cachedStatusLower
    if statusLower == "" { statusLower = strings.ToLower(ng.Status) }
    versionLower := ng.cachedVersionLower
    if versionLower == "" { versionLower = strings.ToLower(ng.Version) }
    instTypesLower := ng.cachedInstanceTypesLower
    if instTypesLower == "" { instTypesLower = strings.ToLower(strings.Join(ng.InstanceTypes, ",")) }
    ltNameLower := ng.cachedLaunchTemplateNameLower
    if ltNameLower == "" { ltNameLower = strings.ToLower(ng.LaunchTemplateName) }
    ltVersionLower := ng.cachedLaunchTemplateVersionLower
    if ltVersionLower == "" { ltVersionLower = strings.ToLower(ng.LaunchTemplateVersion) }
    ltIDLower := ng.cachedLaunchTemplateIDLower
    if ltIDLower == "" { ltIDLower = strings.ToLower(ng.LaunchTemplateID) }

    if strings.Contains(clusterLower, query) ||
        strings.Contains(nameLower, query) ||
        strings.Contains(statusLower, query) ||
        strings.Contains(versionLower, query) ||
        strings.Contains(instTypesLower, query) ||
        strings.Contains(ltNameLower, query) ||
        strings.Contains(ltVersionLower, query) ||
        strings.Contains(ltIDLower, query) {
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
