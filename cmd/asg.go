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
	asgMinSize         int32
	asgMaxSize         int32
	asgDesiredCapacity int32
	asgSkipConfirm     bool
)

var asgCmd = &cobra.Command{
	Use:     "asg",
	Aliases: []string{"autoscaling"},
	Short:   "Manage AWS Auto Scaling Groups",
	Long: `Manage AWS Auto Scaling Groups including scaling operations.

Auto Scaling Groups allow scaling down to 0 instances, unlike EKS node groups.

Examples:
  # Interactive ASG selection for scaling
  aws-ssm asg scale

  # Scale specific ASG
  aws-ssm asg scale my-asg --desired 3

  # Scale with custom min/max/desired
  aws-ssm asg scale my-asg --min 0 --max 10 --desired 3

  # Scale down to 0 instances
  aws-ssm asg scale my-asg --desired 0

  # Using 'autoscaling' alias
  aws-ssm autoscaling scale my-asg --desired 2`,
}

var asgScaleCmd = &cobra.Command{
	Use:   "scale [asg-name]",
	Short: "Scale an Auto Scaling Group",
	Long: `Scale an Auto Scaling Group by updating its min, max, and desired capacity.

If the ASG name is not provided, an interactive fuzzy finder will be displayed
to select the Auto Scaling Group.

Unlike EKS node groups, Auto Scaling Groups can be scaled down to 0 instances.

Examples:
  # Interactive selection
  aws-ssm asg scale

  # Scale specific ASG to desired capacity
  aws-ssm asg scale my-asg --desired 3

  # Scale with custom min/max/desired
  aws-ssm asg scale my-asg --min 0 --max 10 --desired 3

  # Scale down to 0 instances
  aws-ssm asg scale my-asg --desired 0

  # Skip confirmation prompt
  aws-ssm asg scale my-asg --desired 5 --skip-confirm`,
	Args: cobra.MaximumNArgs(1),
	RunE: runASGScale,
}

func init() {
	// Add asg command to root command
	rootCmd.AddCommand(asgCmd)

	// Add scale subcommand to asg command
	asgCmd.AddCommand(asgScaleCmd)

	// Scale command flags
	asgScaleCmd.Flags().Int32Var(&asgMinSize, "min", -1, "Minimum size (optional - defaults to current or desired)")
	asgScaleCmd.Flags().Int32Var(&asgMaxSize, "max", -1, "Maximum size (optional - defaults to current or desired)")
	asgScaleCmd.Flags().Int32Var(&asgDesiredCapacity, "desired", -1, "Desired capacity (required when ASG name is specified)")
	asgScaleCmd.Flags().BoolVar(&asgSkipConfirm, "skip-confirm", false, "Skip confirmation prompt")
}

func runASGScale(_ *cobra.Command, args []string) error {
	// Create a context that can be cancelled with Ctrl+C
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	// Create AWS client
	client, err := aws.NewClient(ctx, region, profile)
	if err != nil {
		return fmt.Errorf("failed to create AWS client: %w", err)
	}

	// For CLI mode (non-interactive), run once without loop
	if len(args) > 0 {
		_, err := runASGScaleOnce(ctx, client, args)
		return err
	}

	// For interactive mode, allow retry on cancel
	for {
		shouldRetry, err := runASGScaleOnce(ctx, client, args)
		if err != nil {
			return err
		}
		if !shouldRetry {
			// User completed the operation or cancelled at fuzzy finder
			return nil
		}
		// User cancelled after selection, loop back to fuzzy finder
		fmt.Println("\nReturning to ASG selection...")
	}
}

// runASGScaleOnce executes one iteration of the ASG scale workflow
// Returns (shouldRetry, error) where shouldRetry indicates if user wants to select a different ASG
func runASGScaleOnce(ctx context.Context, client *aws.Client, args []string) (bool, error) {
	// Resolve ASG and parameters
	selectedASG, asgInfo, err := resolveASGAndParameters(ctx, client, args)
	if err != nil {
		return false, err
	}

	// Check if user cancelled selection (pressed ESC at fuzzy finder)
	if selectedASG == "" {
		return false, nil
	}

	// Get current ASG details
	asg, err := client.DescribeAutoScalingGroup(ctx, selectedASG)
	if err != nil {
		return false, fmt.Errorf("failed to get ASG details: %w", err)
	}

	// Calculate final scaling parameters
	finalParams := calculateScalingParameters(asg, asgInfo)

	// Display configuration and confirm
	shouldRetry, confirmed := confirmASGScalingActionWithRetry(selectedASG, asg, finalParams)
	if !confirmed {
		if shouldRetry {
			// User wants to select a different ASG
			return true, nil
		}
		// User cancelled
		return false, nil
	}

	// Perform the scaling operation
	err = executeASGScaling(ctx, client, selectedASG, finalParams)
	return false, err
}

// resolveASGAndParameters resolves ASG selection and scaling parameters
func resolveASGAndParameters(ctx context.Context, client *aws.Client, args []string) (string, *fuzzy.ASGInfo, error) {
	if len(args) > 0 {
		// Command-line mode
		selectedASG := args[0]
		if asgDesiredCapacity == -1 {
			return "", nil, fmt.Errorf("--desired flag is required when ASG name is provided")
		}
		return selectedASG, &fuzzy.ASGInfo{}, nil
	}

	// Interactive mode
	return selectASGInteractively(ctx, client)
}

// selectASGInteractively selects ASG using fuzzy finder
func selectASGInteractively(ctx context.Context, client *aws.Client) (string, *fuzzy.ASGInfo, error) {
	// Now show the interactive prompt
	printInteractivePrompt("Auto Scaling Group selector")
	fmt.Println()

	colors := fuzzy.NewDefaultColorManager(noColor)
	adapter := &asgClientAdapter{client: client}
	loader := fuzzy.NewAWSASGLoader(adapter)
	finder := fuzzy.NewASGFinder(loader, colors)

	asgInfo, findErr := finder.SelectASGInteractive(ctx)
	if findErr != nil {
		// Check if it's a context cancellation (Ctrl+C)
		if findErr == context.Canceled {
			printSelectionCancelled()
			return "", nil, nil
		}
		return "", nil, fmt.Errorf("failed to select ASG: %w", findErr)
	}
	if asgInfo == nil {
		printNoSelection("Auto Scaling Group")
		return "", nil, nil
	}

	// Prompt for desired capacity if not provided
	if asgDesiredCapacity == -1 {
		promptForDesiredCapacity(asgInfo)
	}

	return asgInfo.Name, asgInfo, nil
}

// promptForDesiredCapacity prompts user for desired capacity in interactive mode
func promptForDesiredCapacity(asgInfo *fuzzy.ASGInfo) {
	fmt.Printf("\nCurrent ASG configuration:\n")
	fmt.Printf("  Min Size:          %d\n", asgInfo.MinSize)
	fmt.Printf("  Max Size:          %d\n", asgInfo.MaxSize)
	fmt.Printf("  Desired Capacity:  %d\n", asgInfo.DesiredCapacity)
	fmt.Printf("  Current Size:      %d\n\n", asgInfo.CurrentSize)
	fmt.Printf("Enter desired capacity: ")
	if _, scanErr := fmt.Scanln(&asgDesiredCapacity); scanErr != nil {
		fmt.Printf("Error reading desired capacity: %v\n", scanErr)
		// Set a default value to avoid further issues
		asgDesiredCapacity = 0
	}
}

// promptForDesiredCapacityWithRetry prompts user for desired capacity with retry support
// Returns true if user wants to retry (select different ASG), false otherwise
//
//nolint:unused // Kept for backward compatibility
func promptForDesiredCapacityWithRetry(asgInfo *fuzzy.ASGInfo) bool {
	fmt.Printf("\nCurrent ASG configuration:\n")
	fmt.Printf("  Min Size:          %d\n", asgInfo.MinSize)
	fmt.Printf("  Max Size:          %d\n", asgInfo.MaxSize)
	fmt.Printf("  Desired Capacity:  %d\n", asgInfo.DesiredCapacity)
	fmt.Printf("  Current Size:      %d\n\n", asgInfo.CurrentSize)
	fmt.Printf("Enter desired capacity (or 'back' to select a different ASG): ")

	var input string
	if _, scanErr := fmt.Scanln(&input); scanErr != nil {
		fmt.Printf("Error reading desired capacity: %v\n", scanErr)
		return false
	}

	// Check if user wants to go back
	if input == "back" || input == "b" {
		return true
	}

	// Try to parse as integer
	var capacity int32
	_, parseErr := fmt.Sscanf(input, "%d", &capacity)
	if parseErr != nil {
		fmt.Printf("Invalid input. Please enter a number or 'back'.\n")
		return false
	}

	asgDesiredCapacity = capacity
	return false
}

// ASGScalingParameters holds the final scaling configuration
type ASGScalingParameters struct {
	Min     int32
	Max     int32
	Desired int32
}

// calculateScalingParameters calculates final min, max, and desired values
func calculateScalingParameters(asg *aws.AutoScalingGroup, _ *fuzzy.ASGInfo) ASGScalingParameters {
	finalMin := asgMinSize
	finalMax := asgMaxSize
	finalDesired := asgDesiredCapacity

	// If min and max are not provided, set them to desired capacity
	if asgMinSize == -1 && asgMaxSize == -1 {
		finalMin = asgDesiredCapacity
		finalMax = asgDesiredCapacity
	} else {
		if asgMinSize == -1 {
			finalMin = asg.MinSize
		}
		if asgMaxSize == -1 {
			finalMax = asg.MaxSize
		}
	}

	return ASGScalingParameters{
		Min:     finalMin,
		Max:     finalMax,
		Desired: finalDesired,
	}
}

// confirmASGScalingAction displays configuration and gets user confirmation
//
//nolint:unused // Kept for backward compatibility
func confirmASGScalingAction(selectedASG string, asg *aws.AutoScalingGroup, params ASGScalingParameters) bool {
	// Display current and new configuration
	fmt.Printf("\nAuto Scaling Group: %s\n", selectedASG)
	fmt.Printf("\nCurrent configuration:\n")
	fmt.Printf("  Min Size:          %d\n", asg.MinSize)
	fmt.Printf("  Max Size:          %d\n", asg.MaxSize)
	fmt.Printf("  Desired Capacity:  %d\n", asg.DesiredCapacity)
	fmt.Printf("  Current Size:      %d\n", asg.CurrentSize)

	fmt.Printf("\nNew configuration:\n")
	fmt.Printf("  Min Size:          %d\n", params.Min)
	fmt.Printf("  Max Size:          %d\n", params.Max)
	fmt.Printf("  Desired Capacity:  %d\n", params.Desired)

	// Confirm before scaling (unless skip-confirm is set)
	if asgSkipConfirm {
		return true
	}

	fmt.Printf("\nDo you want to proceed with scaling? (yes/no): ")
	var confirmation string
	if _, scanErr := fmt.Scanln(&confirmation); scanErr != nil {
		fmt.Printf("Error reading confirmation: %v\n", scanErr)
		return false
	}

	if confirmation != "yes" && confirmation != "y" {
		fmt.Println("Scaling cancelled")
		return false
	}

	return true
}

// confirmASGScalingActionWithRetry displays configuration and gets user confirmation with retry support
// Returns (shouldRetry, confirmed) where shouldRetry indicates if user wants to select a different ASG
func confirmASGScalingActionWithRetry(selectedASG string, asg *aws.AutoScalingGroup, params ASGScalingParameters) (bool, bool) {
	// Display current and new configuration
	fmt.Printf("\nAuto Scaling Group: %s\n", selectedASG)
	fmt.Printf("\nCurrent configuration:\n")
	fmt.Printf("  Min Size:          %d\n", asg.MinSize)
	fmt.Printf("  Max Size:          %d\n", asg.MaxSize)
	fmt.Printf("  Desired Capacity:  %d\n", asg.DesiredCapacity)
	fmt.Printf("  Current Size:      %d\n", asg.CurrentSize)

	fmt.Printf("\nNew configuration:\n")
	fmt.Printf("  Min Size:          %d\n", params.Min)
	fmt.Printf("  Max Size:          %d\n", params.Max)
	fmt.Printf("  Desired Capacity:  %d\n", params.Desired)

	// Confirm before scaling (unless skip-confirm is set)
	if asgSkipConfirm {
		return false, true
	}

	fmt.Printf("\nDo you want to proceed with scaling? (yes/no/back): ")
	var confirmation string
	if _, scanErr := fmt.Scanln(&confirmation); scanErr != nil {
		fmt.Printf("Error reading confirmation: %v\n", scanErr)
		return false, false
	}

	// Check if user wants to go back
	if confirmation == "back" || confirmation == "b" {
		return true, false
	}

	if confirmation != "yes" && confirmation != "y" {
		fmt.Println("Scaling cancelled")
		return false, false
	}

	return false, true
}

// executeASGScaling performs the actual scaling operation
func executeASGScaling(ctx context.Context, client *aws.Client, selectedASG string, params ASGScalingParameters) error {
	fmt.Printf("\nScaling Auto Scaling Group %s...\n", selectedASG)

	err := client.UpdateAutoScalingGroupCapacity(ctx, selectedASG, params.Min, params.Max, params.Desired)
	if err != nil {
		return fmt.Errorf("failed to scale ASG: %w", err)
	}

	fmt.Printf("✓ Successfully initiated scaling for Auto Scaling Group %s\n", selectedASG)
	fmt.Printf("\n")
	fmt.Printf("Note: The scaling operation may take several minutes to complete.\n")
	fmt.Printf("You can check the status with: aws autoscaling describe-auto-scaling-groups --auto-scaling-group-names %s\n", selectedASG)

	return nil
}

// asgClientAdapter adapts aws.Client to fuzzy.AWSASGClientInterface
type asgClientAdapter struct {
	client *aws.Client
}

// ListAutoScalingGroups implements fuzzy.AWSASGClientInterface
func (a *asgClientAdapter) ListAutoScalingGroups(ctx context.Context) ([]string, error) {
	return a.client.ListAutoScalingGroups(ctx)
}

// DescribeAutoScalingGroup implements fuzzy.AWSASGClientInterface
func (a *asgClientAdapter) DescribeAutoScalingGroup(ctx context.Context, asgName string) (*fuzzy.ASGDetail, error) {
	asg, err := a.client.DescribeAutoScalingGroup(ctx, asgName)
	if err != nil {
		return nil, err
	}
	// Convert *aws.AutoScalingGroup to fuzzy.ASGDetail interface
	var detail fuzzy.ASGDetail = asg
	return &detail, nil
}

// GetConfig implements fuzzy.AWSASGClientInterface
func (a *asgClientAdapter) GetConfig() awsconfig.Config {
	return a.client.GetConfig()
}
