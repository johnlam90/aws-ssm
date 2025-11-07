package cmd

import (
	"context"
	"fmt"

	"github.com/aws-ssm/pkg/aws"
	"github.com/spf13/cobra"
)

var (
	useNative bool
)

var sessionCmd = &cobra.Command{
	Use:   "session [instance-identifier] [command]",
	Short: "Start an SSM session with an EC2 instance or execute a command",
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
	ctx := context.Background()

	// Create AWS client
	client, err := aws.NewClient(ctx, region, profile)
	if err != nil {
		return fmt.Errorf("failed to create AWS client: %w", err)
	}

	var instance *aws.Instance
	var command string

	// Parse arguments
	if len(args) == 0 {
		// No arguments - use interactive fuzzy finder
		fmt.Println("Opening interactive instance selector...")
		fmt.Println("(Use arrow keys to navigate, type to filter, Enter to select, Esc to cancel)")
		fmt.Println()

		selectedInstance, err := client.SelectInstanceInteractive(ctx)
		if err != nil {
			return fmt.Errorf("failed to select instance: %w", err)
		}
		instance = selectedInstance
	} else if len(args) == 1 {
		// One argument - instance identifier only (interactive shell)
		identifier := args[0]
		fmt.Printf("Searching for instance: %s\n", identifier)
		instances, err := client.FindInstances(ctx, identifier)
		if err != nil {
			return fmt.Errorf("failed to find instance: %w", err)
		}

		if len(instances) == 0 {
			return fmt.Errorf("no instances found matching: %s", identifier)
		}

		if len(instances) > 1 {
			fmt.Printf("Found %d instances matching '%s':\n\n", len(instances), identifier)
			for i, inst := range instances {
				name := inst.Name
				if name == "" {
					name = "(no name)"
				}
				fmt.Printf("%d. %s - %s [%s] - %s\n", i+1, inst.InstanceID, name, inst.State, inst.PrivateIP)
			}
			return fmt.Errorf("multiple instances found, please use a more specific identifier")
		}

		instance = &instances[0]
	} else {
		// Two arguments - instance identifier and command
		identifier := args[0]
		command = args[1]

		fmt.Printf("Searching for instance: %s\n", identifier)
		instances, err := client.FindInstances(ctx, identifier)
		if err != nil {
			return fmt.Errorf("failed to find instance: %w", err)
		}

		if len(instances) == 0 {
			return fmt.Errorf("no instances found matching: %s", identifier)
		}

		if len(instances) > 1 {
			fmt.Printf("Found %d instances matching '%s':\n\n", len(instances), identifier)
			for i, inst := range instances {
				name := inst.Name
				if name == "" {
					name = "(no name)"
				}
				fmt.Printf("%d. %s - %s [%s] - %s\n", i+1, inst.InstanceID, name, inst.State, inst.PrivateIP)
			}
			return fmt.Errorf("multiple instances found, please use a more specific identifier")
		}

		instance = &instances[0]
	}

	// Check if instance is running
	if instance.State != "running" {
		return fmt.Errorf("instance %s is not running (current state: %s)", instance.InstanceID, instance.State)
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
