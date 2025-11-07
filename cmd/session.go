package cmd

import (
	"context"
	"fmt"

	"github.com/aws-ssm/pkg/aws"
	"github.com/spf13/cobra"
)

var sessionCmd = &cobra.Command{
	Use:   "session [instance-identifier]",
	Short: "Start an SSM session with an EC2 instance",
	Long: `Start an interactive SSM session with an EC2 instance.

The instance can be identified by:
  - Instance ID (e.g., i-1234567890abcdef0)
  - DNS name (e.g., ec2-1-2-3-4.compute.amazonaws.com)
  - Private DNS name (e.g., ip-10-0-0-1.ec2.internal)
  - IP address (public or private)
  - Tag (format: Key:Value, e.g., Name:web-server)
  - Name (uses Name tag, e.g., web-server)

Examples:
  # Connect by instance ID
  aws-ssm session i-1234567890abcdef0

  # Connect by name
  aws-ssm session web-server

  # Connect by tag
  aws-ssm session Environment:production

  # Connect by IP address
  aws-ssm session 10.0.1.100

  # Connect by DNS name
  aws-ssm session ec2-1-2-3-4.us-west-2.compute.amazonaws.com

  # Connect with specific region and profile
  aws-ssm session web-server --region us-west-2 --profile production`,
	Args: cobra.ExactArgs(1),
	RunE: runSession,
}

func init() {
	rootCmd.AddCommand(sessionCmd)
}

func runSession(cmd *cobra.Command, args []string) error {
	identifier := args[0]
	ctx := context.Background()

	// Create AWS client
	client, err := aws.NewClient(ctx, region, profile)
	if err != nil {
		return fmt.Errorf("failed to create AWS client: %w", err)
	}

	// Find the instance
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
		for i, instance := range instances {
			name := instance.Name
			if name == "" {
				name = "(no name)"
			}
			fmt.Printf("%d. %s - %s [%s] - %s\n", i+1, instance.InstanceID, name, instance.State, instance.PrivateIP)
		}
		return fmt.Errorf("multiple instances found, please use a more specific identifier")
	}

	instance := instances[0]

	// Check if instance is running
	if instance.State != "running" {
		return fmt.Errorf("instance %s is not running (current state: %s)", instance.InstanceID, instance.State)
	}

	// Display instance information
	name := instance.Name
	if name == "" {
		name = "(no name)"
	}
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

	// Start SSM session
	if err := client.StartSession(ctx, instance.InstanceID); err != nil {
		return fmt.Errorf("failed to start session: %w", err)
	}

	return nil
}

