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

func runScale(_ *cobra.Command, args []string) error {
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

	// Resolve cluster and node group parameters
	clusterName, resolvedNodeGroupName, err := resolveScalingParameters(ctx, client, args)
	if err != nil {
		return err
	}

	// Get current node group details
	ng, err := client.DescribeNodeGroupPublic(ctx, clusterName, resolvedNodeGroupName)
	if err != nil {
		return fmt.Errorf("failed to describe node group: %w", err)
	}

	// Calculate final scaling parameters
	finalParams := calculateFinalParameters(ng, minSize, maxSize, desiredSize, len(args))

	// Validate parameters
	if err := validateScalingParameters(finalParams); err != nil {
		return err
	}

	// Display configuration
	displayScalingConfiguration(clusterName, resolvedNodeGroupName, ng, finalParams)

	// Confirm action
	if !confirmScalingAction() {
		return nil
	}

	// Perform scaling
	return executeScaling(ctx, client, clusterName, resolvedNodeGroupName, finalParams)
}

// resolveScalingParameters resolves cluster and node group from args or interactive selection
func resolveScalingParameters(ctx context.Context, client *aws.Client, args []string) (string, string, error) {
	// Get cluster name
	clusterName, err := resolveClusterName(ctx, client, args)
	if err != nil {
		return "", "", err
	}

	// Get node group name
	ngName, err := resolveNodeGroupName(ctx, client, clusterName)
	if err != nil {
		return "", "", err
	}

	return clusterName, ngName, nil
}

// resolveClusterName gets cluster name from args or interactive selection
func resolveClusterName(ctx context.Context, client *aws.Client, args []string) (string, error) {
	if len(args) > 0 {
		// When cluster is provided as argument, desired size is required
		if desiredSize == -1 {
			return "", fmt.Errorf("--desired flag is required when cluster name is provided")
		}
		return args[0], nil
	}

	// Interactive selection
	cluster, err := selectEKSClusterInteractive(ctx, client)
	if err != nil {
		return "", fmt.Errorf("failed to select EKS cluster: %w", err)
	}
	if cluster == nil {
		return "", fmt.Errorf("no cluster selected")
	}
	return cluster.Name, nil
}

// resolveNodeGroupName gets node group name from flag or interactive selection
func resolveNodeGroupName(ctx context.Context, client *aws.Client, clusterName string) (string, error) {
	if nodeGroupName != "" {
		return nodeGroupName, nil
	}

	// Interactive selection
	ng, err := selectNodeGroupInteractive(ctx, client, clusterName)
	if err != nil {
		return "", fmt.Errorf("failed to select node group: %w", err)
	}
	if ng == nil {
		return "", fmt.Errorf("no node group selected")
	}
	return ng.Name, nil
}

// ScalingParameters holds the final scaling configuration
type ScalingParameters struct {
	Min     int32
	Max     int32
	Desired int32
}

// calculateFinalParameters calculates the final min, max, and desired values
func calculateFinalParameters(ng *aws.NodeGroup, minSize, maxSize, desiredSize int32, argCount int) ScalingParameters {
	// Prompt for desired size in interactive mode if not provided
	if argCount == 0 && desiredSize == -1 {
		promptForDesiredSize(ng)
	}

	// Calculate final values
	finalMin := resolveMinSize(ng, minSize, desiredSize)
	finalMax := resolveMaxSize(ng, maxSize, desiredSize)
	finalDesired := desiredSize

	return ScalingParameters{
		Min:     finalMin,
		Max:     finalMax,
		Desired: finalDesired,
	}
}

// promptForDesiredSize prompts user for desired size in interactive mode
func promptForDesiredSize(ng *aws.NodeGroup) {
	fmt.Printf("\nCurrent node group configuration:\n")
	fmt.Printf("  Min Size:     %d\n", ng.MinSize)
	fmt.Printf("  Max Size:     %d\n", ng.MaxSize)
	fmt.Printf("  Desired Size: %d\n", ng.DesiredSize)
	fmt.Printf("  Current Size: %d\n\n", ng.CurrentSize)
	fmt.Printf("Enter desired size: ")
	_, err := fmt.Scanln(&desiredSize)
	if err != nil {
		fmt.Printf("Error reading input: %v\n", err)
	}
}

// resolveMinSize determines the final minimum size
func resolveMinSize(ng *aws.NodeGroup, minSize, desiredSize int32) int32 {
	if minSize != -1 {
		return minSize
	}
	if desiredSize == 0 {
		return 0
	}
	return ng.MinSize
}

// resolveMaxSize determines the final maximum size
func resolveMaxSize(ng *aws.NodeGroup, maxSize, desiredSize int32) int32 {
	if maxSize != -1 {
		return maxSize
	}
	if desiredSize == 0 {
		return 0
	}
	if desiredSize > ng.MaxSize {
		return desiredSize
	}
	return ng.MaxSize
}

// validateScalingParameters validates the scaling parameters
func validateScalingParameters(params ScalingParameters) error {
	if params.Min < 0 {
		return fmt.Errorf("min size cannot be negative")
	}
	if params.Max < params.Min {
		return fmt.Errorf("max size (%d) cannot be less than min size (%d)", params.Max, params.Min)
	}
	if params.Desired < params.Min || params.Desired > params.Max {
		return fmt.Errorf("desired size (%d) must be between min size (%d) and max size (%d)", params.Desired, params.Min, params.Max)
	}
	return nil
}

// displayScalingConfiguration displays the current and target configuration
func displayScalingConfiguration(clusterName, nodeGroupName string, ng *aws.NodeGroup, params ScalingParameters) {
	fmt.Printf("\n")
	fmt.Printf("Cluster:       %s\n", clusterName)
	fmt.Printf("Node Group:    %s\n", nodeGroupName)
	fmt.Printf("\n")
	fmt.Printf("Current Configuration:\n")
	fmt.Printf("  Min:         %d\n", ng.MinSize)
	fmt.Printf("  Max:         %d\n", ng.MaxSize)
	fmt.Printf("  Desired:     %d\n", ng.DesiredSize)
	fmt.Printf("  Current:     %d\n", ng.CurrentSize)
	fmt.Printf("\n")
	fmt.Printf("Target Configuration:\n")
	fmt.Printf("  Min:         %d\n", params.Min)
	fmt.Printf("  Max:         %d\n", params.Max)
	fmt.Printf("  Desired:     %d\n", params.Desired)
	fmt.Printf("\n")
}

// confirmScalingAction prompts for user confirmation
func confirmScalingAction() bool {
	if skipConfirm {
		return true
	}

	fmt.Printf("⚠️  Are you sure you want to scale this node group? (yes/no): ")
	var response string
	_, err := fmt.Scanln(&response)
	if err != nil {
		fmt.Printf("Error reading input: %v\n", err)
		return false
	}

	if response != "yes" && response != "y" {
		fmt.Println("Operation cancelled.")
		return false
	}

	fmt.Printf("\n")
	return true
}

// executeScaling performs the actual scaling operation
func executeScaling(ctx context.Context, client *aws.Client, clusterName, nodeGroupName string, params ScalingParameters) error {
	fmt.Printf("Scaling node group %s...\n", nodeGroupName)

	err := client.UpdateNodeGroupScaling(ctx, clusterName, nodeGroupName, params.Min, params.Max, params.Desired)
	if err != nil {
		return fmt.Errorf("failed to scale node group: %w", err)
	}

	fmt.Printf("✓ Successfully initiated scaling for node group %s\n", nodeGroupName)
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
