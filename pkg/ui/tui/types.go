package tui

import (
	"context"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/johnlam90/aws-ssm/pkg/aws"
)

// ViewMode represents the current view in the TUI
type ViewMode int

const (
	ViewDashboard ViewMode = iota
	ViewEC2Instances
	ViewEKSClusters
	ViewASGs
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
	InstanceType     string
	AvailabilityZone string
	Tags             map[string]string
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
	Name            string
	DesiredCapacity int32
	MinSize         int32
	MaxSize         int32
	CurrentSize     int32
	Status          string
}

// LoadingMsg is sent when data is being loaded
type LoadingMsg struct {
	View ViewMode
}

// DataLoadedMsg is sent when data has been loaded
type DataLoadedMsg struct {
	View      ViewMode
	Instances []EC2Instance
	Clusters  []EKSCluster
	ASGs      []ASG
	Error     error
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
				InstanceType:     inst.InstanceType,
				AvailabilityZone: inst.AvailabilityZone,
				Tags:             inst.Tags,
			}
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

		// Fetch details for each ASG
		tuiASGs := make([]ASG, 0, len(asgNames))
		for _, name := range asgNames {
			asg, err := client.DescribeAutoScalingGroup(ctx, name)
			if err != nil {
				// Skip ASGs that fail to describe
				continue
			}

			// Derive status from current vs desired capacity
			status := "Healthy"
			if asg.CurrentSize < asg.DesiredCapacity {
				status = "Scaling Up"
			} else if asg.CurrentSize > asg.DesiredCapacity {
				status = "Scaling Down"
			}

			tuiASGs = append(tuiASGs, ASG{
				Name:            asg.Name,
				DesiredCapacity: asg.DesiredCapacity,
				MinSize:         asg.MinSize,
				MaxSize:         asg.MaxSize,
				CurrentSize:     asg.CurrentSize,
				Status:          status,
			})
		}

		return DataLoadedMsg{
			View: ViewASGs,
			ASGs: tuiASGs,
		}
	}
}
