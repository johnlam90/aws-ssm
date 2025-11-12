package fuzzy

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	fuzzyfinder "github.com/ktr0731/go-fuzzyfinder"
)

// NodeGroupInfo represents a node group for fuzzy finder display
type NodeGroupInfo struct {
	Name               string
	ClusterName        string
	Status             string
	Version            string
	InstanceTypes      []string
	DesiredSize        int32
	MinSize            int32
	MaxSize            int32
	CurrentSize        int32
	CreatedAt          time.Time
	Tags               map[string]string
	LaunchTemplateName string
}

// NodeGroupLoader interface for loading node groups
type NodeGroupLoader interface {
	LoadNodeGroups(ctx context.Context, clusterName string) ([]NodeGroupInfo, error)
	GetNodeGroupDetails(ctx context.Context, clusterName, nodeGroupName string) (*NodeGroupInfo, error)
}

// AWSNodeGroupClientInterface defines the interface for AWS EKS node group operations
type AWSNodeGroupClientInterface interface {
	ListNodeGroupsForCluster(ctx context.Context, clusterName string) ([]string, error)
	DescribeNodeGroupPublic(ctx context.Context, clusterName, nodeGroupName string) (*NodeGroupDetail, error)
	GetConfig() aws.Config
}

// NodeGroupDetail is an interface to avoid circular imports
type NodeGroupDetail interface {
	GetName() string
	GetStatus() string
	GetVersion() string
	GetInstanceTypes() []string
	GetDesiredSize() int32
	GetMinSize() int32
	GetMaxSize() int32
	GetCurrentSize() int32
	GetCreatedAt() time.Time
	GetTags() map[string]string
	GetLaunchTemplateName() string
}

// AWSNodeGroupLoader implements NodeGroupLoader interface using the AWS client
type AWSNodeGroupLoader struct {
	client      AWSNodeGroupClientInterface
	clusterName string
}

// NewAWSNodeGroupLoader creates a new AWS node group loader
func NewAWSNodeGroupLoader(client AWSNodeGroupClientInterface, clusterName string) *AWSNodeGroupLoader {
	return &AWSNodeGroupLoader{
		client:      client,
		clusterName: clusterName,
	}
}

// LoadNodeGroups loads node groups for a cluster
func (l *AWSNodeGroupLoader) LoadNodeGroups(ctx context.Context, clusterName string) ([]NodeGroupInfo, error) {
	nodeGroupNames, err := l.client.ListNodeGroupsForCluster(ctx, clusterName)
	if err != nil {
		return nil, fmt.Errorf("failed to list node groups: %w", err)
	}

	var nodeGroups []NodeGroupInfo
	for _, name := range nodeGroupNames {
		ngDetail, err := l.client.DescribeNodeGroupPublic(ctx, clusterName, name)
		if err != nil {
			fmt.Printf("Warning: failed to describe node group %s: %v\n", name, err)
			// Still add a basic entry if describe fails
			nodeGroups = append(nodeGroups, NodeGroupInfo{
				Name:        name,
				ClusterName: clusterName,
			})
			continue
		}

		// Convert to NodeGroupInfo
		ngInfo := l.convertToNodeGroupInfo(ngDetail, clusterName)
		nodeGroups = append(nodeGroups, ngInfo)
	}

	return nodeGroups, nil
}

// GetNodeGroupDetails retrieves details about a specific node group
func (l *AWSNodeGroupLoader) GetNodeGroupDetails(ctx context.Context, clusterName, nodeGroupName string) (*NodeGroupInfo, error) {
	ngDetail, err := l.client.DescribeNodeGroupPublic(ctx, clusterName, nodeGroupName)
	if err != nil {
		return nil, fmt.Errorf("failed to describe node group: %w", err)
	}

	ngInfo := l.convertToNodeGroupInfo(ngDetail, clusterName)
	return &ngInfo, nil
}

// convertToNodeGroupInfo converts a node group detail to NodeGroupInfo for display
func (l *AWSNodeGroupLoader) convertToNodeGroupInfo(ngDetail *NodeGroupDetail, clusterName string) NodeGroupInfo {
	ngInfo := NodeGroupInfo{
		ClusterName: clusterName,
	}

	if ngDetail == nil {
		return ngInfo
	}

	// Use the interface methods to extract fields
	ng := *ngDetail
	ngInfo.Name = ng.GetName()
	ngInfo.Status = ng.GetStatus()
	ngInfo.Version = ng.GetVersion()
	ngInfo.InstanceTypes = ng.GetInstanceTypes()
	ngInfo.DesiredSize = ng.GetDesiredSize()
	ngInfo.MinSize = ng.GetMinSize()
	ngInfo.MaxSize = ng.GetMaxSize()
	ngInfo.CurrentSize = ng.GetCurrentSize()
	ngInfo.CreatedAt = ng.GetCreatedAt()
	ngInfo.Tags = ng.GetTags()
	ngInfo.LaunchTemplateName = ng.GetLaunchTemplateName()

	return ngInfo
}

// NodeGroupFinder handles node group selection
type NodeGroupFinder struct {
	loader NodeGroupLoader
	colors ColorManager
}

// NewNodeGroupFinder creates a new node group finder
func NewNodeGroupFinder(loader NodeGroupLoader, colors ColorManager) *NodeGroupFinder {
	return &NodeGroupFinder{
		loader: loader,
		colors: colors,
	}
}

// SelectNodeGroupInteractive displays the fuzzy finder for node group selection
func (f *NodeGroupFinder) SelectNodeGroupInteractive(ctx context.Context, clusterName string) (*NodeGroupInfo, error) {
	// Load node groups
	nodeGroups, err := f.loader.LoadNodeGroups(ctx, clusterName)
	if err != nil {
		return nil, fmt.Errorf("failed to load node groups: %w", err)
	}

	if len(nodeGroups) == 0 {
		return nil, fmt.Errorf("no node groups found in cluster %s", clusterName)
	}

	// Use fuzzyfinder to select
	fuzzyfinder := NewNodeGroupFuzzyFinder(nodeGroups, f.colors)
	selectedIndex, err := fuzzyfinder.Select(ctx)
	if err != nil {
		return nil, err
	}

	if selectedIndex < 0 || selectedIndex >= len(nodeGroups) {
		return nil, fmt.Errorf("invalid node group selection")
	}

	return &nodeGroups[selectedIndex], nil
}

// NodeGroupFuzzyFinder handles the actual fuzzy finding for node groups
type NodeGroupFuzzyFinder struct {
	nodeGroups []NodeGroupInfo
	colors     ColorManager
}

// NewNodeGroupFuzzyFinder creates a new node group fuzzy finder
func NewNodeGroupFuzzyFinder(nodeGroups []NodeGroupInfo, colors ColorManager) *NodeGroupFuzzyFinder {
	return &NodeGroupFuzzyFinder{
		nodeGroups: nodeGroups,
		colors:     colors,
	}
}

// Select displays the fuzzy finder and returns the selected node group index
func (f *NodeGroupFuzzyFinder) Select(ctx context.Context) (int, error) {
	// Create preview renderer
	renderer := NewNodeGroupPreviewRenderer(f.colors)

	// Use fuzzyfinder to select with context support for Ctrl+C handling
	selectedIndex, err := fuzzyfinder.Find(
		f.nodeGroups,
		func(i int) string {
			return f.formatNodeGroupRow(f.nodeGroups[i])
		},
		fuzzyfinder.WithPreviewWindow(func(i, width, height int) string {
			if i < 0 || i >= len(f.nodeGroups) {
				return "Select a node group to view details"
			}
			return renderer.Render(&f.nodeGroups[i], width, height)
		}),
		fuzzyfinder.WithPromptString("Node Group > "),
		fuzzyfinder.WithContext(ctx),
	)

	if err != nil {
		if err == fuzzyfinder.ErrAbort {
			return -1, nil // User cancelled
		}
		return -1, err
	}

	return selectedIndex, nil
}

// formatNodeGroupRow formats a node group for display in the fuzzy finder
// Format matches v0.2.0: name | status | desired/min/max | instance-types
func (f *NodeGroupFuzzyFinder) formatNodeGroupRow(ng NodeGroupInfo) string {
	name := ng.Name
	if name == "" {
		name = "(no name)"
	}

	// Truncate name to fit nicely
	if len(name) > 25 {
		name = name[:22] + "..."
	}

	// Format: Name | Status | Desired/Min/Max | Instance Types
	status := ng.Status
	if status == "" {
		status = "UNKNOWN"
	}

	scaling := fmt.Sprintf("%d/%d/%d", ng.DesiredSize, ng.MinSize, ng.MaxSize)

	instanceTypes := "Launch Template"
	if len(ng.InstanceTypes) > 0 {
		instanceTypes = ng.InstanceTypes[0]
		if len(ng.InstanceTypes) > 1 {
			instanceTypes += fmt.Sprintf(" +%d", len(ng.InstanceTypes)-1)
		}
	}

	return fmt.Sprintf("%-25s | %-12s | %-11s | %s",
		name,
		status,
		scaling,
		instanceTypes,
	)
}
