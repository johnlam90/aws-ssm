package cmd

import (
	"context"
	"fmt"

	"github.com/aws-ssm/pkg/aws"
	"github.com/spf13/cobra"
)

var (
	nodeNames   []string
	instanceIDs []string
	filterTags  []string
	showAll     bool
)

var interfacesCmd = &cobra.Command{
	Use:   "interfaces [instance-identifier]",
	Short: "List network interfaces for EC2 instances",
	Long: `Display all network interfaces attached to EC2 instances in a formatted table.
Shows interface names (ens5, ens6, etc.), subnet IDs, CIDRs, and security groups.

This is useful for instances with multiple network interfaces (Multus, EKS, etc.).

If no instance identifier is provided, an interactive fuzzy finder will be displayed
to select from all running instances.

The instance can be identified by:
  - Instance ID (e.g., i-1234567890abcdef0)
  - DNS name (e.g., ip-100-64-149-165.ec2.internal)
  - Private DNS name (e.g., ip-10-0-0-1.ec2.internal)
  - IP address (public or private)
  - Tag (format: Key:Value, e.g., Name:web-server)
  - Name (uses Name tag, e.g., web-server)

Examples:
  # Interactive fuzzy finder (no argument)
  aws-ssm interfaces

  # List interfaces for specific instance by ID
  aws-ssm interfaces i-1234567890abcdef0

  # List interfaces by instance name
  aws-ssm interfaces web-server

  # List interfaces by Kubernetes node name
  aws-ssm interfaces ip-100-64-149-165.ec2.internal

  # List interfaces for multiple nodes
  aws-ssm interfaces --node-name ip-100-64-149-165.ec2.internal --node-name ip-100-64-87-43.ec2.internal

  # List interfaces with tag filter
  aws-ssm interfaces --tag Environment:production

  # List interfaces for all instances (including stopped)
  aws-ssm interfaces --all

  # List with specific region and profile
  aws-ssm interfaces web-server --region us-west-2 --profile production`,
	Args: cobra.MaximumNArgs(1),
	RunE: runInterfaces,
}

func init() {
	rootCmd.AddCommand(interfacesCmd)

	interfacesCmd.Flags().StringSliceVarP(&nodeNames, "node-name", "n", nil, "Kubernetes node DNS name(s) (e.g., ip-100-64-149-165.ec2.internal)")
	interfacesCmd.Flags().StringSliceVar(&instanceIDs, "instance-id", nil, "Specific instance ID(s) to query")
	interfacesCmd.Flags().StringSliceVarP(&filterTags, "tag", "t", nil, "Filter by tag (format: Key:Value, can be used multiple times)")
	interfacesCmd.Flags().BoolVar(&showAll, "all", false, "Show all instances including stopped ones (default: running only)")
}

func runInterfaces(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	// Create AWS client
	client, err := aws.NewClient(ctx, region, profile)
	if err != nil {
		return fmt.Errorf("failed to create AWS client: %w", err)
	}

	var identifier string

	// If no arguments and no filters provided, use interactive fuzzy finder
	if len(args) == 0 && len(nodeNames) == 0 && len(instanceIDs) == 0 && len(filterTags) == 0 {
		fmt.Println("Opening interactive instance selector...")
		fmt.Println("(Use arrow keys to navigate, type to filter, Enter to select, Esc to cancel)")
		fmt.Println()

		selectedInstance, err := client.SelectInstanceInteractive(ctx)
		if err != nil {
			return fmt.Errorf("failed to select instance: %w", err)
		}

		// Use the selected instance ID as identifier
		identifier = selectedInstance.InstanceID
	} else if len(args) > 0 {
		identifier = args[0]
	}

	// Build options
	opts := aws.InterfacesOptions{
		Identifier:  identifier,
		NodeNames:   nodeNames,
		InstanceIDs: instanceIDs,
		FilterTags:  filterTags,
		ShowAll:     showAll,
	}

	// List interfaces
	if err := client.ListNetworkInterfaces(ctx, opts); err != nil {
		return fmt.Errorf("failed to list network interfaces: %w", err)
	}

	return nil
}
