package cmd

import (
	"context"
	"fmt"

	"github.com/aws-ssm/pkg/aws"
	"github.com/spf13/cobra"
)

var (
	remotePort int
	localPort  int
)

var portForwardCmd = &cobra.Command{
	Use:   "port-forward [instance-identifier]",
	Short: "Forward a local port to a remote port on an EC2 instance",
	Long: `Forward a local port to a remote port on an EC2 instance via SSM.

This allows you to access services running on the instance without exposing them to the internet.
For example, you can forward a local port to a database running on the instance.

The instance can be identified by:
  - Instance ID (e.g., i-1234567890abcdef0)
  - DNS name (e.g., ec2-1-2-3-4.compute.amazonaws.com)
  - Private DNS name (e.g., ip-10-0-0-1.ec2.internal)
  - IP address (public or private)
  - Tag (format: Key:Value, e.g., Name:web-server)
  - Name (uses Name tag, e.g., web-server)

Examples:
  # Forward local port 3306 to remote MySQL port 3306
  aws-ssm port-forward db-server --remote-port 3306 --local-port 3306

  # Forward local port 8080 to remote port 80
  aws-ssm port-forward web-server --remote-port 80 --local-port 8080

  # Access RDS through a bastion instance
  aws-ssm port-forward bastion --remote-port 5432 --local-port 5432`,
	Args: cobra.ExactArgs(1),
	RunE: runPortForward,
}

func init() {
	rootCmd.AddCommand(portForwardCmd)
	portForwardCmd.Flags().IntVarP(&remotePort, "remote-port", "R", 0, "Remote port on the instance (required)")
	portForwardCmd.Flags().IntVarP(&localPort, "local-port", "L", 0, "Local port to listen on (required)")
	portForwardCmd.MarkFlagRequired("remote-port")
	portForwardCmd.MarkFlagRequired("local-port")
}

func runPortForward(cmd *cobra.Command, args []string) error {
	identifier := args[0]
	ctx := context.Background()

	// Validate ports
	if remotePort < 1 || remotePort > 65535 {
		return fmt.Errorf("invalid remote port: %d (must be between 1 and 65535)", remotePort)
	}
	if localPort < 1 || localPort > 65535 {
		return fmt.Errorf("invalid local port: %d (must be between 1 and 65535)", localPort)
	}

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
	fmt.Printf("Port forwarding to instance:\n")
	fmt.Printf("  ID:          %s\n", instance.InstanceID)
	fmt.Printf("  Name:        %s\n", name)
	fmt.Printf("  State:       %s\n", instance.State)
	fmt.Printf("  Private IP:  %s\n\n", instance.PrivateIP)

	// Start port forwarding session
	if err := client.StartPortForwardingSession(ctx, instance.InstanceID, remotePort, localPort); err != nil {
		return fmt.Errorf("failed to start port forwarding: %w", err)
	}

	return nil
}

