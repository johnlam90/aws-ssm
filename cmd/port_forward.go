package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

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
	if err := portForwardCmd.MarkFlagRequired("remote-port"); err != nil {
		panic(fmt.Sprintf("failed to mark remote-port as required: %v", err))
	}
	if err := portForwardCmd.MarkFlagRequired("local-port"); err != nil {
		panic(fmt.Sprintf("failed to mark local-port as required: %v", err))
	}
}

func runPortForward(cmd *cobra.Command, args []string) error {
	identifier := args[0]

	// Create a context that can be cancelled with Ctrl+C
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

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

    // Resolve instance with interactive fallback
	fmt.Printf("Searching for instance: %s\n", identifier)
	instance, err := client.ResolveSingleInstance(ctx, identifier)
	if err != nil {
		if multiErr, ok := err.(*aws.MultipleInstancesError); ok && multiErr.AllowInteractive {
			fmt.Print(multiErr.FormatInstanceList())
			selected, selErr := client.SelectInstanceFromProvided(ctx, multiErr.Instances)
			if selErr != nil {
				return fmt.Errorf("instance selection cancelled or failed: %w", selErr)
			}
			instance = selected
		} else {
			return err
		}
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
