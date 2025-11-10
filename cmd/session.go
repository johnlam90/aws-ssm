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

func runSession(cmd *cobra.Command, args []string) error {
	// Create a context that can be cancelled with Ctrl+C
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	// Create AWS client with interactive flags
	client, err := aws.NewClientWithFlags(ctx, region, profile, interactive, interactiveCols, noColor, width, favorites)
	if err != nil {
		return fmt.Errorf("failed to create AWS client: %w", err)
	}

	var instance *aws.Instance
	var command string

	// Parse arguments
	switch len(args) {
	case 0:
		// No arguments - use interactive fuzzy finder
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
			// Check if it's a context cancellation (Ctrl+C)
			if err == context.Canceled {
				fmt.Println("\nSelection cancelled.")
				return nil
			}
			return fmt.Errorf("failed to select instance: %w", err)
		}
		if selectedInstance == nil {
			// User cancelled the selection (Esc)
			fmt.Println("\nNo instance selected.")
			return nil
		}
		instance = selectedInstance
	case 1:
		// One argument - instance identifier only (interactive shell)
		identifier := args[0]
		fmt.Printf("Searching for instance: %s\n", identifier)

		resolvedInstance, err := client.ResolveSingleInstance(ctx, identifier)
		if err != nil {
			if multiErr, ok := err.(*aws.MultipleInstancesError); ok && multiErr.AllowInteractive {
				fmt.Print(multiErr.FormatInstanceList())
				selected, selErr := client.SelectInstanceFromProvided(ctx, multiErr.Instances)
				if selErr != nil {
					// Check if it's a context cancellation (Ctrl+C)
					if selErr == context.Canceled {
						fmt.Println("\nSelection cancelled.")
						return nil
					}
					return fmt.Errorf("instance selection cancelled or failed: %w", selErr)
				}
				if selected == nil {
					// User cancelled the selection (Esc)
					fmt.Println("\nNo instance selected.")
					return nil
				}
				instance = selected
				break
			}
			return err
		}
		instance = resolvedInstance
	default:
		// Two or more arguments - instance identifier and command
		identifier := args[0]
		command = args[1]

		fmt.Printf("Searching for instance: %s\n", identifier)

		resolvedInstance, err := client.ResolveSingleInstance(ctx, identifier)
		if err != nil {
			if multiErr, ok := err.(*aws.MultipleInstancesError); ok && multiErr.AllowInteractive {
				fmt.Print(multiErr.FormatInstanceList())
				selected, selErr := client.SelectInstanceFromProvided(ctx, multiErr.Instances)
				if selErr != nil {
					// Check if it's a context cancellation (Ctrl+C)
					if selErr == context.Canceled {
						fmt.Println("\nSelection cancelled.")
						return nil
					}
					return fmt.Errorf("instance selection cancelled or failed: %w", selErr)
				}
				if selected == nil {
					// User cancelled the selection (Esc)
					fmt.Println("\nNo instance selected.")
					return nil
				}
				instance = selected
				break
			}
			return err
		}
		instance = resolvedInstance
	}

	// Display instance information
	name := instance.Name
	if name == "" {
		name = "(no name)"
	}

	// If command is provided, execute it and return
	if command != "" {
		fmt.Printf("Executing command on instance:\n")
		fmt.Printf("  ID:          %s\n", instance.InstanceID)
		fmt.Printf("  Name:        %s\n", name)
		fmt.Printf("  Command:     %s\n\n", command)

		output, err := client.ExecuteCommand(ctx, instance.InstanceID, command)
		if err != nil {
			return fmt.Errorf("failed to execute command: %w", err)
		}

		// Print the output
		fmt.Print(output)
		return nil
	}

	// Otherwise, start an interactive session
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

	// Start SSM session - choose between native or plugin-based
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
