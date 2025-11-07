package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/aws-ssm/pkg/aws"
	"github.com/spf13/cobra"
)

var (
	tagFilter []string
	allStates bool
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List EC2 instances",
	Long: `List EC2 instances with optional filtering by tags.
	
Examples:
  # List all running instances
  aws-ssm list

  # List instances with specific tag
  aws-ssm list --tag Name=web-server

  # List instances with multiple tags
  aws-ssm list --tag Environment=production --tag Team=backend

  # List instances in a specific region
  aws-ssm list --region us-west-2`,
	RunE: runList,
}

func init() {
	rootCmd.AddCommand(listCmd)
	listCmd.Flags().StringSliceVarP(&tagFilter, "tag", "t", []string{}, "Filter by tags (format: Key=Value)")
	listCmd.Flags().BoolVarP(&allStates, "all", "a", false, "Show instances in all states (not just running)")
}

func runList(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	// Create AWS client
	client, err := aws.NewClient(ctx, region, profile)
	if err != nil {
		return fmt.Errorf("failed to create AWS client: %w", err)
	}

	// Parse tag filters
	tagFilters := make(map[string]string)
	for _, tag := range tagFilter {
		parts := strings.SplitN(tag, "=", 2)
		if len(parts) != 2 {
			return fmt.Errorf("invalid tag format: %s (expected Key=Value)", tag)
		}
		tagFilters[parts[0]] = parts[1]
	}

	// List instances
	instances, err := client.ListInstances(ctx, tagFilters)
	if err != nil {
		return fmt.Errorf("failed to list instances: %w", err)
	}

	if len(instances) == 0 {
		fmt.Println("No instances found")
		return nil
	}

	// Display instances in a table
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	if _, err := fmt.Fprintln(w, "INSTANCE ID\tNAME\tSTATE\tINSTANCE TYPE\tPRIVATE IP\tPUBLIC IP\tAVAILABILITY ZONE"); err != nil {
		return fmt.Errorf("failed to write table header: %w", err)
	}
	if _, err := fmt.Fprintln(w, strings.Repeat("-", 11)+"\t"+strings.Repeat("-", 4)+"\t"+strings.Repeat("-", 5)+"\t"+strings.Repeat("-", 13)+"\t"+strings.Repeat("-", 10)+"\t"+strings.Repeat("-", 9)+"\t"+strings.Repeat("-", 17)); err != nil {
		return fmt.Errorf("failed to write table separator: %w", err)
	}

	for _, instance := range instances {
		// Skip non-running instances unless --all flag is set
		if !allStates && instance.State != "running" {
			continue
		}

		name := instance.Name
		if name == "" {
			name = "-"
		}

		publicIP := instance.PublicIP
		if publicIP == "" {
			publicIP = "-"
		}

		if _, err := fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
			instance.InstanceID,
			name,
			instance.State,
			instance.InstanceType,
			instance.PrivateIP,
			publicIP,
			instance.AvailabilityZone,
		); err != nil {
			return fmt.Errorf("failed to write table row: %w", err)
		}
	}

	if err := w.Flush(); err != nil {
		return fmt.Errorf("failed to flush table writer: %w", err)
	}

	return nil
}
