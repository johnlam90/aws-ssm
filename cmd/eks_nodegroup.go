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
	nodeGroupName         string
	minSize               int32
	maxSize               int32
	desiredSize           int32
	skipConfirm           bool
	launchTemplateVersion string
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

var updateLTCmd = &cobra.Command{
	Use:     "update-lt [cluster-name]",
	Aliases: []string{"update-launch-template"},
	Short:   "Update the launch template version of an EKS node group",
	Long: `Update the launch template version of an EKS node group.

If the cluster name or node group name is not provided, an interactive fuzzy finder
will be displayed to select them. If the launch template version is not provided,
an interactive selection will be displayed.

Examples:
  # Interactive selection (recommended)
  aws-ssm eks nodegroup update-lt

  # Interactive selection for specific cluster
  aws-ssm eks nodegroup update-lt my-cluster

  # Update specific node group to specific version
  aws-ssm eks nodegroup update-lt my-cluster --nodegroup my-ng --version 5

  # Update to $Latest version
  aws-ssm eks nodegroup update-lt my-cluster --nodegroup my-ng --version '$Latest'

  # Update to $Default version
  aws-ssm eks nodegroup update-lt my-cluster --nodegroup my-ng --version '$Default'

  # Skip confirmation prompt
  aws-ssm eks ng update-lt my-cluster --nodegroup my-ng --version 5 --skip-confirm`,
	Args: cobra.MaximumNArgs(1),
	RunE: runUpdateLT,
}

func init() {
	// Add nodegroup command to eks command
	eksCmd.AddCommand(eksNodeGroupCmd)

	// Add subcommands to nodegroup command
	eksNodeGroupCmd.AddCommand(scaleCmd)
	eksNodeGroupCmd.AddCommand(updateLTCmd)

	// Scale command flags
	scaleCmd.Flags().StringVar(&nodeGroupName, "nodegroup", "", "Node group name (if not provided, interactive selection will be used)")
	scaleCmd.Flags().Int32Var(&minSize, "min", -1, "Minimum size (optional - defaults to current or desired)")
	scaleCmd.Flags().Int32Var(&maxSize, "max", -1, "Maximum size (optional - defaults to current or desired)")
	scaleCmd.Flags().Int32Var(&desiredSize, "desired", -1, "Desired size (required when cluster is specified)")
	scaleCmd.Flags().BoolVar(&skipConfirm, "skip-confirm", false, "Skip confirmation prompt")

	// Update launch template command flags
	updateLTCmd.Flags().StringVar(&nodeGroupName, "nodegroup", "", "Node group name (if not provided, interactive selection will be used)")
	updateLTCmd.Flags().StringVar(&launchTemplateVersion, "version", "", "Launch template version (if not provided, interactive selection will be used)")
	updateLTCmd.Flags().BoolVar(&skipConfirm, "skip-confirm", false, "Skip confirmation prompt")
}

func runScale(_ *cobra.Command, args []string) error {
	// Create a context that can be cancelled with Ctrl+C
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	// Create AWS client
	client, err := aws.NewClient(ctx, region, profile)
	if err != nil {
		return fmt.Errorf("failed to create AWS client: %w", err)
	}

	// Validate required flags for CLI mode
	if len(args) > 0 && desiredSize == -1 {
		return fmt.Errorf("--desired flag is required when cluster name is provided")
	}

	// Resolve cluster and node group
	clusterName, resolvedNodeGroupName, err := resolveClusterAndNodeGroup(ctx, client, args)
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

// resolveClusterAndNodeGroup resolves cluster and node group from args or interactive selection
// This is a generic function used by both scale and update-lt commands
func resolveClusterAndNodeGroup(ctx context.Context, client *aws.Client, args []string) (string, string, error) {
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
	// Show loading message with spinner
	s := createLoadingSpinner("Loading node groups...")
	s.Start()

	// Load node groups first (this is the slow part)
	nodeGroupNames, err := client.ListNodeGroupsForCluster(ctx, clusterName)
	s.Stop()

	if err != nil {
		return nil, fmt.Errorf("failed to list node groups: %w", err)
	}

	if len(nodeGroupNames) == 0 {
		return nil, fmt.Errorf("no node groups found in cluster %s", clusterName)
	}

	// Now show the interactive prompt
	printInteractivePrompt("node group selector")
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
			printSelectionCancelled()
			return nil, nil
		}
		return nil, fmt.Errorf("failed to select node group: %w", err)
	}

	return selectedNodeGroup, nil
}

// launchTemplateClientAdapter adapts aws.Client to fuzzy.AWSLaunchTemplateClientInterface
type launchTemplateClientAdapter struct {
	client *aws.Client
}

// ListLaunchTemplateVersions implements fuzzy.AWSLaunchTemplateClientInterface
func (a *launchTemplateClientAdapter) ListLaunchTemplateVersions(ctx context.Context, launchTemplateID string) ([]fuzzy.LaunchTemplateVersionDetail, error) {
	versions, err := a.client.ListLaunchTemplateVersions(ctx, launchTemplateID)
	if err != nil {
		return nil, err
	}

	// Convert to interface slice
	var details []fuzzy.LaunchTemplateVersionDetail
	for i := range versions {
		var detail fuzzy.LaunchTemplateVersionDetail = &versions[i]
		details = append(details, detail)
	}
	return details, nil
}

// GetConfig implements fuzzy.AWSLaunchTemplateClientInterface
func (a *launchTemplateClientAdapter) GetConfig() awsconfig.Config {
	return a.client.GetConfig()
}

func runUpdateLT(_ *cobra.Command, args []string) error {
	// Create a context that can be cancelled with Ctrl+C
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	// Create AWS client
	client, err := aws.NewClient(ctx, region, profile)
	if err != nil {
		return fmt.Errorf("failed to create AWS client: %w", err)
	}

	// Resolve cluster and node group
	clusterName, resolvedNodeGroupName, err := resolveClusterAndNodeGroup(ctx, client, args)
	if err != nil {
		return err
	}

	// Get current node group details
	ng, err := client.DescribeNodeGroupPublic(ctx, clusterName, resolvedNodeGroupName)
	if err != nil {
		return fmt.Errorf("failed to describe node group: %w", err)
	}

	// Check if node group has a launch template
	if ng.LaunchTemplate.ID == "" {
		return fmt.Errorf("node group %s does not use a launch template", resolvedNodeGroupName)
	}

	// Resolve launch template version
	version, err := resolveLaunchTemplateVersion(ctx, client, ng)
	if err != nil {
		return err
	}

	// Display configuration
	displayLTUpdateConfiguration(clusterName, resolvedNodeGroupName, ng, version)

	// Confirm action
	if !confirmLTUpdateAction() {
		return nil
	}

	// Perform update
	return executeLTUpdate(ctx, client, clusterName, resolvedNodeGroupName, ng.LaunchTemplate.ID, version)
}

func resolveLaunchTemplateVersion(ctx context.Context, client *aws.Client, ng *aws.NodeGroup) (string, error) {
	if launchTemplateVersion != "" {
		return launchTemplateVersion, nil
	}

	// Interactive selection
	return selectLaunchTemplateVersionInteractive(ctx, client, ng)
}

func selectLaunchTemplateVersionInteractive(ctx context.Context, client *aws.Client, ng *aws.NodeGroup) (string, error) {
	fmt.Println()

	// Show loading message with spinner
	s := createLoadingSpinner("Loading launch template versions...")
	s.Start()

	// Load versions first (this is the slow part)
	versions, err := client.ListLaunchTemplateVersions(ctx, ng.LaunchTemplate.ID)
	s.Stop()

	if err != nil {
		return "", fmt.Errorf("failed to list launch template versions: %w", err)
	}

	if len(versions) == 0 {
		return "", fmt.Errorf("no launch template versions found")
	}

	// Now show the interactive prompt
	printInteractivePrompt("launch template version selector")

	// Display current version with color
	if noColor {
		fmt.Printf("\nCurrent version: %s\n", ng.LaunchTemplate.Version)
	} else {
		label := fuzzy.ColorDim + "Current version:" + fuzzy.ColorReset
		version := fuzzy.ColorCyan + ng.LaunchTemplate.Version + fuzzy.ColorReset
		fmt.Printf("\n%s %s\n", label, version)
	}
	fmt.Println()

	// Create adapter
	adapter := &launchTemplateClientAdapter{client: client}

	// Create launch template version loader
	loader := fuzzy.NewAWSLaunchTemplateLoader(adapter, ng.LaunchTemplate.ID, ng.LaunchTemplate.Name)

	// Create color manager
	colors := fuzzy.NewDefaultColorManager(noColor)

	// Create launch template version finder
	finder := fuzzy.NewLaunchTemplateVersionFinder(loader, colors)

	// Select version
	selectedVersion, err := finder.SelectVersionInteractive(ctx)
	if err != nil {
		// Check if it's a context cancellation (Ctrl+C)
		if err == context.Canceled {
			printSelectionCancelled()
			return "", nil
		}
		return "", fmt.Errorf("failed to select launch template version: %w", err)
	}

	if selectedVersion == nil {
		return "", fmt.Errorf("no version selected")
	}

	// Convert version to string
	return fuzzy.GetVersionString(selectedVersion), nil
}

func displayLTUpdateConfiguration(clusterName, nodeGroupName string, ng *aws.NodeGroup, version string) {
	fmt.Printf("\n")
	fmt.Printf("Cluster:                  %s\n", clusterName)
	fmt.Printf("Node Group:               %s\n", nodeGroupName)
	fmt.Printf("\n")
	fmt.Printf("Current Configuration:\n")
	fmt.Printf("  Launch Template:        %s (%s)\n", ng.LaunchTemplate.Name, ng.LaunchTemplate.ID)
	fmt.Printf("  Current Version:        %s\n", ng.LaunchTemplate.Version)
	fmt.Printf("\n")
	fmt.Printf("Target Configuration:\n")
	fmt.Printf("  New Version:            %s\n", version)
	fmt.Printf("\n")
}

func confirmLTUpdateAction() bool {
	if skipConfirm {
		return true
	}

	fmt.Printf("⚠️  Are you sure you want to update the launch template version? (yes/no): ")
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

func executeLTUpdate(ctx context.Context, client *aws.Client, clusterName, nodeGroupName, launchTemplateID, version string) error {
	fmt.Printf("Updating launch template version for node group %s...\n", nodeGroupName)

	err := client.UpdateNodeGroupLaunchTemplate(ctx, clusterName, nodeGroupName, launchTemplateID, version)
	if err != nil {
		return fmt.Errorf("failed to update launch template: %w", err)
	}

	fmt.Printf("✓ Successfully initiated launch template update for node group %s\n", nodeGroupName)
	fmt.Printf("\n")
	fmt.Printf("Note: The update operation may take several minutes to complete.\n")
	fmt.Printf("      Nodes will be replaced with the new launch template version.\n")
	fmt.Printf("You can check the status with: aws-ssm eks %s\n", clusterName)

	return nil
}
