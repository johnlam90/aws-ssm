package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	awsconfig "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/johnlam90/aws-ssm/pkg/aws"
	"github.com/johnlam90/aws-ssm/pkg/ui/fuzzy"
	"github.com/spf13/cobra"
)

var (
	nodeGroupName string
	minSize       int32
	maxSize       int32
	desiredSize   int32
	skipConfirm   bool
)

var eksNodeGroupCmd = &cobra.Command{
	Use:     "nodegroup",
	Aliases: []string{"ng"},
	Short:   "Manage EKS node groups",
	Long: `Manage EKS node groups including scaling operations.

Examples:
  # Interactive node group selection for scaling
  aws-ssm eks nodegroup scale my-cluster

  # Scale specific node group
  aws-ssm eks nodegroup scale my-cluster --nodegroup my-ng --desired 3

  # Scale with custom min/max/desired
  aws-ssm eks nodegroup scale my-cluster --nodegroup my-ng --min 1 --max 5 --desired 3

  # Using 'ng' alias
  aws-ssm eks ng scale my-cluster --desired 2`,
}

var scaleCmd = &cobra.Command{
	Use:   "scale [cluster-name]",
	Short: "Scale an EKS node group",
	Long: `Scale an EKS node group by updating its min, max, and desired capacity.

If the node group name is not provided, an interactive fuzzy finder will be displayed
to select the node group.

Examples:
  # Interactive selection
  aws-ssm eks nodegroup scale my-cluster

  # Scale specific node group to desired size
  aws-ssm eks nodegroup scale my-cluster --nodegroup my-ng --desired 3

  # Scale with custom min/max/desired
  aws-ssm eks nodegroup scale my-cluster --nodegroup my-ng --min 1 --max 5 --desired 3

  # Skip confirmation prompt
  aws-ssm eks nodegroup scale my-cluster --nodegroup my-ng --desired 0 --skip-confirm`,
	Args: cobra.MaximumNArgs(1),
	RunE: runScale,
}

func init() {
	// Add nodegroup command to eks command
	eksCmd.AddCommand(eksNodeGroupCmd)

	// Add scale subcommand to nodegroup command
	eksNodeGroupCmd.AddCommand(scaleCmd)

	// Scale command flags
	scaleCmd.Flags().StringVar(&nodeGroupName, "nodegroup", "", "Node group name (if not provided, interactive selection will be used)")
	scaleCmd.Flags().Int32Var(&minSize, "min", -1, "Minimum size (optional - defaults to current or desired)")
	scaleCmd.Flags().Int32Var(&maxSize, "max", -1, "Maximum size (optional - defaults to current or desired)")
	scaleCmd.Flags().Int32Var(&desiredSize, "desired", -1, "Desired size (required when cluster is specified)")
	scaleCmd.Flags().BoolVar(&skipConfirm, "skip-confirm", false, "Skip confirmation prompt")
}

func runScale(cmd *cobra.Command, args []string) error {
	// Setup context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle Ctrl+C gracefully
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigChan
		fmt.Println("\nOperation cancelled by user")
		cancel()
	}()

	// Create AWS client
	client, err := aws.NewClient(ctx, region, profile)
	if err != nil {
		return fmt.Errorf("failed to create AWS client: %w", err)
	}

	var clusterName string

	// Get cluster name from argument or use fuzzy finder
	if len(args) > 0 {
		clusterName = args[0]
		// When cluster is provided as argument, desired size is required
		if desiredSize == -1 {
			return fmt.Errorf("--desired flag is required when cluster name is provided")
		}
	} else {
		// Use interactive fuzzy finder to select cluster
		cluster, selectionErr := selectEKSClusterInteractive(ctx, client)
		if selectionErr != nil {
			return fmt.Errorf("failed to select EKS cluster: %w", selectionErr)
		}

		if cluster == nil {
			fmt.Println("No cluster selected")
			return nil
		}

		clusterName = cluster.Name
	}

	// Get node group name from flag or use fuzzy finder
	var selectedNodeGroup string
	if nodeGroupName != "" {
		selectedNodeGroup = nodeGroupName
	} else {
		// Use interactive fuzzy finder to select node group
		ng, selectionErr := selectNodeGroupInteractive(ctx, client, clusterName)
		if selectionErr != nil {
			return fmt.Errorf("failed to select node group: %w", selectionErr)
		}

		if ng == nil {
			fmt.Println("No node group selected")
			return nil
		}

		selectedNodeGroup = ng.Name
	}

	// Get current node group details
	ng, err := client.DescribeNodeGroupPublic(ctx, clusterName, selectedNodeGroup)
	if err != nil {
		return fmt.Errorf("failed to describe node group: %w", err)
	}

	// If in interactive mode and desired size not provided, prompt for it
	if len(args) == 0 && desiredSize == -1 {
		fmt.Printf("\nCurrent node group configuration:\n")
		fmt.Printf("  Min Size:     %d\n", ng.MinSize)
		fmt.Printf("  Max Size:     %d\n", ng.MaxSize)
		fmt.Printf("  Desired Size: %d\n", ng.DesiredSize)
		fmt.Printf("  Current Size: %d\n\n", ng.CurrentSize)
		fmt.Printf("Enter desired size: ")
		if _, scanErr := fmt.Scanln(&desiredSize); scanErr != nil {
			return fmt.Errorf("failed to read desired size: %w", scanErr)
		}
	}

	// Determine final min/max/desired values
	finalMin := minSize
	finalMax := maxSize
	finalDesired := desiredSize

	// If min/max not provided, use current values or calculate from desired
	if finalMin == -1 {
		switch {
		case finalDesired == 0:
			finalMin = 0
		default:
			finalMin = ng.MinSize
		}
	}

	if finalMax == -1 {
		switch {
		case finalDesired == 0:
			finalMax = 0
		case finalDesired > ng.MaxSize:
			finalMax = finalDesired
		default:
			finalMax = ng.MaxSize
		}
	}

	// Validate scaling parameters
	if finalMin < 0 {
		return fmt.Errorf("min size cannot be negative")
	}
	if finalMax < finalMin {
		return fmt.Errorf("max size (%d) cannot be less than min size (%d)", finalMax, finalMin)
	}
	if finalDesired < finalMin || finalDesired > finalMax {
		return fmt.Errorf("desired size (%d) must be between min size (%d) and max size (%d)", finalDesired, finalMin, finalMax)
	}

	// Display current and target configuration
	fmt.Printf("\n")
	fmt.Printf("Cluster:       %s\n", clusterName)
	fmt.Printf("Node Group:    %s\n", selectedNodeGroup)
	fmt.Printf("\n")
	fmt.Printf("Current Configuration:\n")
	fmt.Printf("  Min:         %d\n", ng.MinSize)
	fmt.Printf("  Max:         %d\n", ng.MaxSize)
	fmt.Printf("  Desired:     %d\n", ng.DesiredSize)
	fmt.Printf("  Current:     %d\n", ng.CurrentSize)
	fmt.Printf("\n")
	fmt.Printf("Target Configuration:\n")
	fmt.Printf("  Min:         %d\n", finalMin)
	fmt.Printf("  Max:         %d\n", finalMax)
	fmt.Printf("  Desired:     %d\n", finalDesired)
	fmt.Printf("\n")

	// Confirm action unless skip-confirm is set
	if !skipConfirm {
		fmt.Printf("⚠️  Are you sure you want to scale this node group? (yes/no): ")
		var response string
		if _, scanErr := fmt.Scanln(&response); scanErr != nil {
			return fmt.Errorf("failed to read response: %w", scanErr)
		}
		if response != "yes" && response != "y" {
			fmt.Println("Operation cancelled.")
			return nil
		}
		fmt.Printf("\n")
	}

	// Perform the scaling operation
	fmt.Printf("Scaling node group %s...\n", selectedNodeGroup)
	err = client.UpdateNodeGroupScaling(ctx, clusterName, selectedNodeGroup, finalMin, finalMax, finalDesired)
	if err != nil {
		return fmt.Errorf("failed to scale node group: %w", err)
	}

	fmt.Printf("✓ Successfully initiated scaling for node group %s\n", selectedNodeGroup)
	fmt.Printf("\n")
	fmt.Printf("Note: The scaling operation may take several minutes to complete.\n")
	fmt.Printf("You can check the status with: aws-ssm eks %s\n", clusterName)

	return nil
}

// nodeGroupClientAdapter adapts aws.Client to fuzzy.AWSNodeGroupClientInterface
type nodeGroupClientAdapter struct {
	client *aws.Client
}

// ListNodeGroupsForCluster implements fuzzy.AWSNodeGroupClientInterface
func (a *nodeGroupClientAdapter) ListNodeGroupsForCluster(ctx context.Context, clusterName string) ([]string, error) {
	return a.client.ListNodeGroupsForCluster(ctx, clusterName)
}

// DescribeNodeGroupPublic implements fuzzy.AWSNodeGroupClientInterface
func (a *nodeGroupClientAdapter) DescribeNodeGroupPublic(ctx context.Context, clusterName, nodeGroupName string) (*fuzzy.NodeGroupDetail, error) {
	ng, err := a.client.DescribeNodeGroupPublic(ctx, clusterName, nodeGroupName)
	if err != nil {
		return nil, err
	}
	// Convert *aws.NodeGroup to fuzzy.NodeGroupDetail interface
	var detail fuzzy.NodeGroupDetail = ng
	return &detail, nil
}

// GetConfig implements fuzzy.AWSNodeGroupClientInterface
func (a *nodeGroupClientAdapter) GetConfig() awsconfig.Config {
	return a.client.GetConfig()
}

// selectNodeGroupInteractive displays an interactive fuzzy finder to select a node group
func selectNodeGroupInteractive(ctx context.Context, client *aws.Client, clusterName string) (*fuzzy.NodeGroupInfo, error) {
	fmt.Println("Opening interactive node group selector...")
	fmt.Println("(Use arrow keys to navigate, type to filter, Enter to select, Esc to cancel)")
	fmt.Println()

	// Create adapter
	adapter := &nodeGroupClientAdapter{client: client}

	// Create node group loader
	loader := fuzzy.NewAWSNodeGroupLoader(adapter, clusterName)

	// Create color manager
	colors := fuzzy.NewDefaultColorManager(noColor)

	// Create node group finder
	finder := fuzzy.NewNodeGroupFinder(loader, colors)

	// Select node group
	selectedNodeGroup, err := finder.SelectNodeGroupInteractive(ctx, clusterName)
	if err != nil {
		// Check if it's a context cancellation (Ctrl+C)
		if err == context.Canceled {
			fmt.Println("\nSelection cancelled.")
			return nil, nil
		}
		return nil, fmt.Errorf("failed to select node group: %w", err)
	}

	return selectedNodeGroup, nil
}
