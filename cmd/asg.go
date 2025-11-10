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

	var selectedASG string

	// Determine if we're in interactive mode or command-line mode
	if len(args) > 0 {
		selectedASG = args[0]
		// When ASG is provided as argument, desired capacity is required
		if asgDesiredCapacity == -1 {
			return fmt.Errorf("--desired flag is required when ASG name is provided")
		}
	} else {
		// Interactive mode - use fuzzy finder to select ASG
		colors := fuzzy.NewDefaultColorManager(noColor)
		adapter := &asgClientAdapter{client: client}
		loader := fuzzy.NewAWSASGLoader(adapter)
		finder := fuzzy.NewASGFinder(loader, colors)

		asgInfo, findErr := finder.SelectASGInteractive(ctx)
		if findErr != nil {
			return fmt.Errorf("failed to select ASG: %w", findErr)
		}
		if asgInfo == nil {
			return fmt.Errorf("no ASG selected")
		}

		selectedASG = asgInfo.Name

		// If in interactive mode and desired capacity not provided, prompt for it
		if asgDesiredCapacity == -1 {
			fmt.Printf("\nCurrent ASG configuration:\n")
			fmt.Printf("  Min Size:          %d\n", asgInfo.MinSize)
			fmt.Printf("  Max Size:          %d\n", asgInfo.MaxSize)
			fmt.Printf("  Desired Capacity:  %d\n", asgInfo.DesiredCapacity)
			fmt.Printf("  Current Size:      %d\n\n", asgInfo.CurrentSize)
			fmt.Printf("Enter desired capacity: ")
			if _, scanErr := fmt.Scanln(&asgDesiredCapacity); scanErr != nil {
				return fmt.Errorf("failed to read desired capacity: %w", scanErr)
			}
		}
	}

	// Get current ASG details
	asg, err := client.DescribeAutoScalingGroup(ctx, selectedASG)
	if err != nil {
		return fmt.Errorf("failed to get ASG details: %w", err)
	}

	// Determine final min/max/desired values
	finalMin := asgMinSize
	finalMax := asgMaxSize
	finalDesired := asgDesiredCapacity

	// If min and max are not provided, set them to desired capacity
	switch {
	case asgMinSize == -1 && asgMaxSize == -1:
		finalMin = asgDesiredCapacity
		finalMax = asgDesiredCapacity
	case asgMinSize == -1:
		finalMin = asg.MinSize
	case asgMaxSize == -1:
		finalMax = asg.MaxSize
	}

	// Display current and new configuration
	fmt.Printf("\nAuto Scaling Group: %s\n", selectedASG)
	fmt.Printf("\nCurrent configuration:\n")
	fmt.Printf("  Min Size:          %d\n", asg.MinSize)
	fmt.Printf("  Max Size:          %d\n", asg.MaxSize)
	fmt.Printf("  Desired Capacity:  %d\n", asg.DesiredCapacity)
	fmt.Printf("  Current Size:      %d\n", asg.CurrentSize)

	fmt.Printf("\nNew configuration:\n")
	fmt.Printf("  Min Size:          %d\n", finalMin)
	fmt.Printf("  Max Size:          %d\n", finalMax)
	fmt.Printf("  Desired Capacity:  %d\n", finalDesired)

	// Confirm before scaling (unless skip-confirm is set)
	if !asgSkipConfirm {
		fmt.Printf("\nDo you want to proceed with scaling? (yes/no): ")
		var confirmation string
		if _, scanErr := fmt.Scanln(&confirmation); scanErr != nil {
			return fmt.Errorf("failed to read confirmation: %w", scanErr)
		}
		if confirmation != "yes" && confirmation != "y" {
			fmt.Println("Scaling cancelled")
			return nil
		}
	}

	// Perform the scaling operation
	fmt.Printf("\nScaling Auto Scaling Group %s...\n", selectedASG)
	err = client.UpdateAutoScalingGroupCapacity(ctx, selectedASG, finalMin, finalMax, finalDesired)
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
