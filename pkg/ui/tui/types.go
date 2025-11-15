package tui

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/johnlam90/aws-ssm/pkg/aws"
	"golang.org/x/sync/errgroup"
)

// ViewMode represents the current view in the TUI
type ViewMode int

const (
	ViewDashboard ViewMode = iota
	ViewEC2Instances
	ViewEKSClusters
	ViewASGs
	ViewNodeGroups
	ViewNetworkInterfaces
	ViewHelp
)

// String returns the string representation of the view mode
func (v ViewMode) String() string {
	switch v {
	case ViewDashboard:
		return "Dashboard"
	case ViewEC2Instances:
		return "EC2 Instances"
	case ViewEKSClusters:
		return "EKS Clusters"
	case ViewASGs:
		return "Auto Scaling Groups"
	case ViewNodeGroups:
		return "EKS Node Groups"
	case ViewNetworkInterfaces:
		return "Network Interfaces"
	case ViewHelp:
		return "Help"
	default:
		return "Unknown"
	}
}

// MenuItem represents an item in the dashboard menu
type MenuItem struct {
	Title       string
	Description string
	View        ViewMode
	Icon        string
}

// EC2Instance represents an EC2 instance in the TUI
type EC2Instance struct {
	InstanceID       string
	Name             string
	State            string
	PrivateIP        string
	PublicIP         string
	PrivateDNS       string
	PublicDNS        string
	InstanceType     string
	AvailabilityZone string
	Tags             map[string]string
	LaunchTime       time.Time
	InstanceProfile  string
	SecurityGroups   []string
	
	// Cached search fields for performance optimization
	cachedNameLower      string
	cachedIDLower        string
	cachedStateLower     string
	cachedTypeLower      string
	cachedPrivateIPLower string
	cachedPublicIPLower  string
	cachedTagsString     string
}

// EKSCluster represents an EKS cluster in the TUI
type EKSCluster struct {
	Name    string
	Status  string
	Version string
	Arn     string
}

// ASG represents an Auto Scaling Group in the TUI
type ASG struct {
	Name                    string
	DesiredCapacity         int32
	MinSize                 int32
	MaxSize                 int32
	CurrentSize             int32
	Status                  string
	HealthCheckType         string
	CreatedAt               time.Time
	AvailabilityZones       []string
	Tags                    map[string]string
	LaunchTemplateName      string
	LaunchTemplateVersion   string
	LaunchConfigurationName string
	LoadBalancerNames       []string
	TargetGroupARNs         []string
}

// NodeGroup represents an EKS node group in the TUI
type NodeGroup struct {
	ClusterName           string
	Name                  string
	Status                string
	Version               string
	InstanceTypes         []string
	DesiredSize           int32
	MinSize               int32
	MaxSize               int32
	CurrentSize           int32
	CreatedAt             string
	LaunchTemplateID      string
	LaunchTemplateName    string
	LaunchTemplateVersion string
	Tags                  map[string]string
	
	// Cached search fields for performance optimization
	cachedClusterLower           string
	cachedNameLower              string
	cachedStatusLower            string
	cachedVersionLower           string
	cachedInstanceTypesJoined    string
	cachedInstanceTypesLower     string
	cachedLaunchTemplateNameLower string
	cachedLaunchTemplateVersionLower string
	cachedLaunchTemplateIDLower  string
}

// LoadingMsg is sent when data is being loaded
type LoadingMsg struct {
	View ViewMode
}

// DataLoadedMsg is sent when data has been loaded
type DataLoadedMsg struct {
	View             ViewMode
	Instances        []EC2Instance
	Clusters         []EKSCluster
	ASGs             []ASG
	NodeGroups       []NodeGroup
	NetworkInstances []aws.InstanceInterfaces
	Error            error
}

// ScalingResultMsg is emitted when a scaling operation finishes
type ScalingResultMsg struct {
	View  ViewMode
	Error error
}

// LaunchTemplateVersionsMsg is sent when launch template versions are loaded
type LaunchTemplateVersionsMsg struct {
	ClusterName   string
	NodeGroupName string
	Versions      []aws.LaunchTemplateVersion
	Error         error
}

// LaunchTemplateUpdateResultMsg is emitted when a launch template update completes
type LaunchTemplateUpdateResultMsg struct {
	ClusterName   string
	NodeGroupName string
	Version       string
	Error         error
}

// SearchDebounceMsg is sent when search debounce completes
type SearchDebounceMsg struct {
	View  ViewMode
	Query string
}

// ErrorMsg represents an error message
type ErrorMsg struct {
	Err error
}

// AWSClientProvider provides access to AWS client
type AWSClientProvider interface {
	GetClient(ctx context.Context) (*aws.Client, error)
}

// Config holds TUI configuration
type Config struct {
	Region     string
	Profile    string
	ConfigPath string
	NoColor    bool
}

// PrecomputeSearchFields precomputes searchable fields for performance
func (e *EC2Instance) PrecomputeSearchFields() {
	e.cachedNameLower = strings.ToLower(e.Name)
	e.cachedIDLower = strings.ToLower(e.InstanceID)
	e.cachedStateLower = strings.ToLower(e.State)
	e.cachedTypeLower = strings.ToLower(e.InstanceType)
	e.cachedPrivateIPLower = strings.ToLower(e.PrivateIP)
	e.cachedPublicIPLower = strings.ToLower(e.PublicIP)
	
	// Precompute tags string
	tagPairs := make([]string, 0, len(e.Tags))
	for k, v := range e.Tags {
		tagPairs = append(tagPairs, fmt.Sprintf("%s:%s", strings.ToLower(k), strings.ToLower(v)))
	}
	e.cachedTagsString = strings.Join(tagPairs, " ")
}

// PrecomputeSearchFields precomputes searchable fields for performance
func (ng *NodeGroup) PrecomputeSearchFields() {
	ng.cachedClusterLower = strings.ToLower(ng.ClusterName)
	ng.cachedNameLower = strings.ToLower(ng.Name)
	ng.cachedStatusLower = strings.ToLower(ng.Status)
	ng.cachedVersionLower = strings.ToLower(ng.Version)
	ng.cachedInstanceTypesJoined = strings.Join(ng.InstanceTypes, ",")
	ng.cachedInstanceTypesLower = strings.ToLower(ng.cachedInstanceTypesJoined)
	ng.cachedLaunchTemplateNameLower = strings.ToLower(ng.LaunchTemplateName)
	ng.cachedLaunchTemplateVersionLower = strings.ToLower(ng.LaunchTemplateVersion)
	ng.cachedLaunchTemplateIDLower = strings.ToLower(ng.LaunchTemplateID)
}

const (
	asgDescribeWorkerLimit       = 6
	nodeGroupDescribeWorkerLimit = 8
)

// LoadEC2InstancesCmd loads EC2 instances asynchronously
func LoadEC2InstancesCmd(ctx context.Context, client *aws.Client) tea.Cmd {
	return func() tea.Msg {
		instances, err := client.ListInstances(ctx, nil)
		if err != nil {
			return DataLoadedMsg{
				View:  ViewEC2Instances,
				Error: err,
			}
		}

		// Convert to TUI instances
		tuiInstances := make([]EC2Instance, len(instances))
		for i, inst := range instances {
			tuiInstances[i] = EC2Instance{
				InstanceID:       inst.InstanceID,
				Name:             inst.Name,
				State:            inst.State,
				PrivateIP:        inst.PrivateIP,
				PublicIP:         inst.PublicIP,
				PrivateDNS:       inst.PrivateDNS,
				PublicDNS:        inst.PublicDNS,
				InstanceType:     inst.InstanceType,
				AvailabilityZone: inst.AvailabilityZone,
				Tags:             inst.Tags,
				LaunchTime:       inst.LaunchTime,
				InstanceProfile:  inst.InstanceProfile,
				SecurityGroups:   append([]string{}, inst.SecurityGroups...),
			}
			// Precompute search fields for performance
			tuiInstances[i].PrecomputeSearchFields()
		}

		return DataLoadedMsg{
			View:      ViewEC2Instances,
			Instances: tuiInstances,
		}
	}
}

// LoadEKSClustersCmd loads EKS clusters asynchronously
func LoadEKSClustersCmd(ctx context.Context, client *aws.Client) tea.Cmd {
	return func() tea.Msg {
		clusterNames, err := client.ListClusters(ctx)
		if err != nil {
			return DataLoadedMsg{
				View:  ViewEKSClusters,
				Error: err,
			}
		}

		// Convert to TUI clusters - fetch basic details for each
		tuiClusters := make([]EKSCluster, 0, len(clusterNames))
		for _, name := range clusterNames {
			// Fetch basic cluster details
			cluster, err := client.DescribeClusterBasic(ctx, name)
			if err != nil {
				// Skip clusters that fail to describe
				continue
			}

			tuiClusters = append(tuiClusters, EKSCluster{
				Name:    cluster.Name,
				Status:  cluster.Status,
				Version: cluster.Version,
				Arn:     cluster.ARN,
			})
		}

		return DataLoadedMsg{
			View:     ViewEKSClusters,
			Clusters: tuiClusters,
		}
	}
}

// LoadASGsCmd loads Auto Scaling Groups asynchronously
func LoadASGsCmd(ctx context.Context, client *aws.Client) tea.Cmd {
	return func() tea.Msg {
		asgNames, err := client.ListAutoScalingGroups(ctx)
		if err != nil {
			return DataLoadedMsg{
				View:  ViewASGs,
				Error: err,
			}
		}

		if len(asgNames) == 0 {
			return DataLoadedMsg{
				View: ViewASGs,
				ASGs: []ASG{},
			}
		}

		var (
			mu      sync.Mutex
			tuiASGs = make([]ASG, 0, len(asgNames))
		)
		g, gCtx := errgroup.WithContext(ctx)
		concurrency := asgDescribeWorkerLimit
		if len(asgNames) < concurrency {
			concurrency = len(asgNames)
		}
		if concurrency == 0 {
			concurrency = 1
		}
		sem := make(chan struct{}, concurrency)

		for _, name := range asgNames {
			name := name
			g.Go(func() error {
				select {
				case <-gCtx.Done():
					return gCtx.Err()
				default:
				}

				sem <- struct{}{}
				defer func() { <-sem }()

				asg, err := client.DescribeAutoScalingGroup(gCtx, name)
				if err != nil {
					// Skip ASGs that fail to describe but keep loading others
					return nil
				}

				mu.Lock()
				tuiASGs = append(tuiASGs, convertToTUIASG(asg))
				mu.Unlock()
				return nil
			})
		}

		if err := g.Wait(); err != nil {
			return DataLoadedMsg{
				View:  ViewASGs,
				Error: err,
			}
		}

		return DataLoadedMsg{
			View: ViewASGs,
			ASGs: tuiASGs,
		}
	}
}

// LoadNodeGroupsCmd loads EKS node groups asynchronously
func LoadNodeGroupsCmd(ctx context.Context, client *aws.Client) tea.Cmd {
	return func() tea.Msg {
		clusterNames, err := client.ListClusters(ctx)
		if err != nil {
			return DataLoadedMsg{
				View:  ViewNodeGroups,
				Error: err,
			}
		}

		type nodeGroupTarget struct {
			clusterName   string
			nodeGroupName string
		}

		var targets []nodeGroupTarget
		for _, clusterName := range clusterNames {
			ngNames, err := client.ListNodeGroupsForCluster(ctx, clusterName)
			if err != nil {
				// Skip problematic clusters but continue loading others
				continue
			}

			for _, ngName := range ngNames {
				targets = append(targets, nodeGroupTarget{
					clusterName:   clusterName,
					nodeGroupName: ngName,
				})
			}
		}

		if len(targets) == 0 {
			return DataLoadedMsg{
				View:       ViewNodeGroups,
				NodeGroups: []NodeGroup{},
			}
		}

		var (
			nodeGroups = make([]NodeGroup, 0, len(targets))
			mu         sync.Mutex
		)

		g, gCtx := errgroup.WithContext(ctx)
		concurrency := nodeGroupDescribeWorkerLimit
		if len(targets) < concurrency {
			concurrency = len(targets)
		}
		if concurrency == 0 {
			concurrency = 1
		}
		sem := make(chan struct{}, concurrency)

		for _, target := range targets {
			target := target
			g.Go(func() error {
				select {
				case <-gCtx.Done():
					return gCtx.Err()
				default:
				}

				sem <- struct{}{}
				defer func() { <-sem }()

				ng, err := client.DescribeNodeGroupPublic(gCtx, target.clusterName, target.nodeGroupName)
				if err != nil {
					return nil
				}

				mu.Lock()
				nodeGroups = append(nodeGroups, convertToTUINodeGroup(target.clusterName, ng))
				mu.Unlock()
				return nil
			})
		}

		if err := g.Wait(); err != nil {
			return DataLoadedMsg{
				View:  ViewNodeGroups,
				Error: err,
			}
		}

		return DataLoadedMsg{
			View:       ViewNodeGroups,
			NodeGroups: nodeGroups,
		}
	}
}

// LoadNetworkInterfacesCmd loads network interfaces asynchronously
func LoadNetworkInterfacesCmd(ctx context.Context, client *aws.Client) tea.Cmd {
	return func() tea.Msg {
		interfaces, err := client.FetchNetworkInterfaces(ctx, aws.InterfacesOptions{})
		if err != nil {
			return DataLoadedMsg{
				View:  ViewNetworkInterfaces,
				Error: err,
			}
		}

		return DataLoadedMsg{
			View:             ViewNetworkInterfaces,
			NetworkInstances: interfaces,
		}
	}
}

func convertToTUIASG(asg *aws.AutoScalingGroup) ASG {
	if asg == nil {
		return ASG{}
	}

	status := "Healthy"
	if asg.CurrentSize < asg.DesiredCapacity {
		status = "Scaling Up"
	} else if asg.CurrentSize > asg.DesiredCapacity {
		status = "Scaling Down"
	}

	availabilityZones := append([]string{}, asg.AvailabilityZones...)
	loadBalancers := append([]string{}, asg.LoadBalancerNames...)
	targetGroups := append([]string{}, asg.TargetGroupARNs...)
	tags := make(map[string]string, len(asg.Tags))
	for k, v := range asg.Tags {
		tags[k] = v
	}

	return ASG{
		Name:                    asg.Name,
		DesiredCapacity:         asg.DesiredCapacity,
		MinSize:                 asg.MinSize,
		MaxSize:                 asg.MaxSize,
		CurrentSize:             asg.CurrentSize,
		Status:                  status,
		HealthCheckType:         asg.HealthCheckType,
		CreatedAt:               asg.CreatedTime,
		AvailabilityZones:       availabilityZones,
		Tags:                    tags,
		LaunchTemplateName:      asg.LaunchTemplateName,
		LaunchTemplateVersion:   asg.LaunchTemplateVersion,
		LaunchConfigurationName: asg.LaunchConfigurationName,
		LoadBalancerNames:       loadBalancers,
		TargetGroupARNs:         targetGroups,
	}
}

func convertToTUINodeGroup(clusterName string, ng *aws.NodeGroup) NodeGroup {
	if ng == nil {
		return NodeGroup{ClusterName: clusterName}
	}

	tags := make(map[string]string, len(ng.Tags))
	for k, v := range ng.Tags {
		tags[k] = v
	}

	nodeGroup := NodeGroup{
		ClusterName:           clusterName,
		Name:                  ng.Name,
		Status:                ng.Status,
		Version:               ng.Version,
		InstanceTypes:         ng.InstanceTypes,
		DesiredSize:           ng.DesiredSize,
		MinSize:               ng.MinSize,
		MaxSize:               ng.MaxSize,
		CurrentSize:           ng.CurrentSize,
		CreatedAt:             ng.CreatedAt.Format("2006-01-02 15:04:05"),
		LaunchTemplateID:      ng.LaunchTemplate.ID,
		LaunchTemplateName:    ng.LaunchTemplate.Name,
		LaunchTemplateVersion: ng.LaunchTemplate.Version,
		Tags:                  tags,
	}
	
	// Precompute search fields for performance
	nodeGroup.PrecomputeSearchFields()
	
	return nodeGroup
}

// ScaleASGCmd scales an Auto Scaling Group without leaving the TUI
func ScaleASGCmd(ctx context.Context, client *aws.Client, asgName string, desired, currentMin, currentMax int32) tea.Cmd {
	return func() tea.Msg {
		min := currentMin
		max := currentMax

		if desired < min {
			min = desired
		}
		if desired > max {
			max = desired
		}

		err := client.UpdateAutoScalingGroupCapacity(ctx, asgName, min, max, desired)
		if err != nil {
			return ScalingResultMsg{View: ViewASGs, Error: err}
		}

		return ScalingResultMsg{View: ViewASGs}
	}
}

// ScaleNodeGroupCmd scales an EKS node group inline
func ScaleNodeGroupCmd(ctx context.Context, client *aws.Client, clusterName, nodeGroupName string, desired, currentMin, currentMax int32) tea.Cmd {
	return func() tea.Msg {
		min := currentMin
		max := currentMax

		if desired < min {
			min = desired
		}
		if desired > max {
			max = desired
		}

		err := client.UpdateNodeGroupScaling(ctx, clusterName, nodeGroupName, min, max, desired)
		if err != nil {
			return ScalingResultMsg{View: ViewNodeGroups, Error: err}
		}

		return ScalingResultMsg{View: ViewNodeGroups}
	}
}

// LoadLaunchTemplateVersionsCmd retrieves launch template versions for a node group
func LoadLaunchTemplateVersionsCmd(ctx context.Context, client *aws.Client, launchTemplateID, clusterName, nodeGroupName string) tea.Cmd {
	return func() tea.Msg {
		versions, err := client.ListLaunchTemplateVersions(ctx, launchTemplateID)
		if err != nil {
			return LaunchTemplateVersionsMsg{
				ClusterName:   clusterName,
				NodeGroupName: nodeGroupName,
				Error:         err,
			}
		}

		return LaunchTemplateVersionsMsg{
			ClusterName:   clusterName,
			NodeGroupName: nodeGroupName,
			Versions:      versions,
		}
	}
}

// UpdateNodeGroupLaunchTemplateCmd updates the launch template version for a node group
func UpdateNodeGroupLaunchTemplateCmd(ctx context.Context, client *aws.Client, clusterName, nodeGroupName, launchTemplateID, version string) tea.Cmd {
	return func() tea.Msg {
		err := client.UpdateNodeGroupLaunchTemplate(ctx, clusterName, nodeGroupName, launchTemplateID, version)
		return LaunchTemplateUpdateResultMsg{
			ClusterName:   clusterName,
			NodeGroupName: nodeGroupName,
			Version:       version,
			Error:         err,
		}
	}
}
