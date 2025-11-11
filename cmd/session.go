package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/johnlam90/aws-ssm/pkg/aws"
	"github.com/spf13/cobra"
)

var (
	useNative bool
)

var sessionCmd = &cobra.Command{
	Use:     "session [instance-identifier] [command]",
	Aliases: []string{"s"},
	Short:   "Start an SSM session with an EC2 instance or execute a command",
	Long: `Start an interactive SSM session with an EC2 instance or execute a remote command.

If no instance identifier is provided, an interactive fuzzy finder will be displayed
to select from all running instances.

If a command is provided as the second argument, it will be executed on the remote
instance and the output will be displayed, instead of starting an interactive shell.

The instance can be identified by:
  - Instance ID (e.g., i-1234567890abcdef0)
  - DNS name (e.g., ec2-1-2-3-4.compute.amazonaws.com)
  - Private DNS name (e.g., ip-10-0-0-1.ec2.internal)
  - IP address (public or private)
  - Tag (format: Key:Value, e.g., Name:web-server)
  - Name (uses Name tag, e.g., web-server)

Examples:
  # Interactive fuzzy finder (no argument)
  aws-ssm session

  # Connect by instance ID (interactive shell)
  aws-ssm session i-1234567890abcdef0

  # Execute a command on instance
  aws-ssm session i-1234567890abcdef0 "df -h"

  # Execute command by instance name
  aws-ssm session web-server "uptime"

  # Execute multi-word command
  aws-ssm session web-server "ps aux | grep nginx"

  # Connect by tag
  aws-ssm session Environment:production

  # Connect by IP address
  aws-ssm session 10.0.1.100

  # Connect by DNS name
  aws-ssm session ec2-1-2-3-4.us-west-2.compute.amazonaws.com

  # Execute command with specific region and profile
  aws-ssm session web-server "systemctl status nginx" --region us-west-2 --profile production`,
	Args: cobra.MaximumNArgs(2),
	RunE: runSession,
}

func init() {
	rootCmd.AddCommand(sessionCmd)
	sessionCmd.Flags().BoolVarP(&useNative, "native", "n", true, "Use native Go implementation (no plugin required)")
}

func runSession(_ *cobra.Command, args []string) error {
	// Create a context that can be cancelled with Ctrl+C
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	// Create AWS client with interactive flags
	client, err := aws.NewClientWithFlags(ctx, region, profile, interactive, interactiveCols, noColor, width, favorites)
	if err != nil {
		return fmt.Errorf("failed to create AWS client: %w", err)
	}

	// Parse and resolve arguments
	instance, command, err := parseAndResolveArgs(ctx, client, args)
	if err != nil {
		return err
	}

	// If no instance resolved, use interactive selection
	if instance == nil {
		instance, err = selectInstanceInteractive(ctx, client)
		if err != nil {
			return err
		}
		if instance == nil {
			return nil // User cancelled
		}
	}

	// Execute based on command presence
	if command != "" {
		return executeRemoteCommand(ctx, client, instance, command)
	}

	return startInteractiveSession(ctx, client, instance)
}

// parseAndResolveArgs parses command line arguments and resolves instance
func parseAndResolveArgs(ctx context.Context, client *aws.Client, args []string) (*aws.Instance, string, error) {
	switch len(args) {
	case 0:
		return nil, "", nil // Interactive selection
	case 1:
		// One argument - instance identifier only
		instance, err := resolveInstance(ctx, client, args[0])
		if err != nil {
			err = handleInstanceResolutionError(ctx, client, err)
			return nil, "", err
		}
		return instance, "", nil
	default:
		// Two or more arguments - instance identifier and command
		instance, err := resolveInstance(ctx, client, args[0])
		if err != nil {
			err = handleInstanceResolutionError(ctx, client, err)
			return nil, "", err
		}
		return instance, args[1], nil
	}
}

// resolveInstance resolves an instance from an identifier
func resolveInstance(ctx context.Context, client *aws.Client, identifier string) (*aws.Instance, error) {
	fmt.Printf("Searching for instance: %s\n", identifier)
	return client.ResolveSingleInstance(ctx, identifier)
}

// handleInstanceResolutionError handles errors from instance resolution
func handleInstanceResolutionError(ctx context.Context, client *aws.Client, err error) error {
	if multiErr, ok := err.(*aws.MultipleInstancesError); ok && multiErr.AllowInteractive {
		return selectFromMultipleInstances(ctx, client, multiErr.Instances)
	}
	return err
}

// selectFromMultipleInstances handles interactive selection from multiple matching instances
func selectFromMultipleInstances(ctx context.Context, client *aws.Client, instances []aws.Instance) error {
	fmt.Print("Multiple instances found:\n\n")

	selected, selErr := client.SelectInstanceFromProvided(ctx, instances)
	if selErr != nil {
		if selErr == context.Canceled {
			fmt.Println("\nSelection cancelled.")
			return nil
		}
		return fmt.Errorf("instance selection cancelled or failed: %w", selErr)
	}

	if selected == nil {
		fmt.Println("\nNo instance selected.")
		return nil
	}

	return nil
}

// selectInstanceInteractive handles interactive instance selection
func selectInstanceInteractive(ctx context.Context, client *aws.Client) (*aws.Instance, error) {
	if interactive {
		fmt.Println("Opening enhanced interactive instance selector...")
		fmt.Println("(Use arrow keys to navigate, type to filter, Space to multi-select, Enter to confirm)")
		fmt.Println()
	} else {
		fmt.Println("Opening interactive instance selector...")
		fmt.Println("(Use arrow keys to navigate, type to filter, Enter to select, Esc to cancel)")
		fmt.Println()
	}

	selectedInstance, err := client.SelectInstanceInteractive(ctx)
	if err != nil {
		if err == context.Canceled {
			fmt.Println("\nSelection cancelled.")
			return nil, nil
		}
		return nil, fmt.Errorf("failed to select instance: %w", err)
	}

	if selectedInstance == nil {
		fmt.Println("\nNo instance selected.")
		return nil, nil
	}

	return selectedInstance, nil
}

// executeRemoteCommand executes a command on a remote instance
func executeRemoteCommand(ctx context.Context, client *aws.Client, instance *aws.Instance, command string) error {
	name := getInstanceDisplayName(instance)

	fmt.Printf("Executing command on instance:\n")
	fmt.Printf("  ID:          %s\n", instance.InstanceID)
	fmt.Printf("  Name:        %s\n", name)
	fmt.Printf("  Command:     %s\n\n", command)

	output, err := client.ExecuteCommand(ctx, instance.InstanceID, command)
	if err != nil {
		return fmt.Errorf("failed to execute command: %w", err)
	}

	fmt.Print(output)
	return nil
}

// startInteractiveSession starts an interactive SSM session
func startInteractiveSession(ctx context.Context, client *aws.Client, instance *aws.Instance) error {
	name := getInstanceDisplayName(instance)

	fmt.Printf("Connecting to instance:\n")
	fmt.Printf("  ID:          %s\n", instance.InstanceID)
	fmt.Printf("  Name:        %s\n", name)
	fmt.Printf("  State:       %s\n", instance.State)
	fmt.Printf("  Type:        %s\n", instance.InstanceType)
	fmt.Printf("  Private IP:  %s\n", instance.PrivateIP)
	if instance.PublicIP != "" {
		fmt.Printf("  Public IP:   %s\n", instance.PublicIP)
	}
	fmt.Printf("  AZ:          %s\n\n", instance.AvailabilityZone)

	if useNative {
		if err := client.StartNativeSession(ctx, instance.InstanceID); err != nil {
			return fmt.Errorf("failed to start native session: %w", err)
		}
	} else {
		if err := client.StartSession(ctx, instance.InstanceID); err != nil {
			return fmt.Errorf("failed to start session: %w", err)
		}
	}

	return nil
}

// getInstanceDisplayName returns a display name for an instance
func getInstanceDisplayName(instance *aws.Instance) string {
	if name := instance.Name; name != "" {
		return name
	}
	return "(no name)"
}
