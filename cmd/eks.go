package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/aws-ssm/pkg/aws"
	"github.com/spf13/cobra"
)

var eksCmd = &cobra.Command{
	Use:     "eks [cluster-name]",
	Aliases: []string{"e"},
	Short:   "Manage EKS clusters",
	Long: `Manage and interact with EKS (Elastic Kubernetes Service) clusters.

Display comprehensive information about EKS clusters including:
- Cluster status and version
- Node groups and Fargate profiles
- Networking configuration (VPC, subnets, security groups)
- API server endpoint details
- Tags and metadata

Examples:
  # Interactive fuzzy finder (recommended - no argument)
  aws-ssm eks

  # Get specific cluster information
  aws-ssm eks my-cluster

  # Specific region
  aws-ssm eks --region us-west-2

  # Specific profile
  aws-ssm eks --profile production`,
	Args: cobra.MaximumNArgs(1),
	RunE: runEKS,
}

func init() {
	rootCmd.AddCommand(eksCmd)
}

func runEKS(cmd *cobra.Command, args []string) error {
	// Create a context that can be cancelled with Ctrl+C
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	// Create AWS client
	client, err := aws.NewClient(ctx, region, profile)
	if err != nil {
		return fmt.Errorf("failed to create AWS client: %w", err)
	}

	var clusterName string

	// If cluster name is provided as argument, use it directly
	if len(args) > 0 {
		clusterName = args[0]
	} else {
		// Use interactive fuzzy finder to select cluster
		cluster, err := selectEKSClusterInteractive(ctx, client)
		if err != nil {
			return fmt.Errorf("failed to select EKS cluster: %w", err)
		}

		if cluster == nil {
			fmt.Println("No cluster selected")
			return nil
		}

		clusterName = cluster.Name
	}

	// Describe the cluster
	cluster, err := client.DescribeCluster(ctx, clusterName)
	if err != nil {
		return fmt.Errorf("failed to describe EKS cluster: %w", err)
	}

	// Display cluster information
	displayClusterInfo(cluster)

	return nil
}

// selectEKSClusterInteractive displays an interactive fuzzy finder to select an EKS cluster
func selectEKSClusterInteractive(ctx context.Context, client *aws.Client) (*aws.Cluster, error) {
	fmt.Println("Opening interactive EKS cluster selector...")
	fmt.Println("(Use arrow keys to navigate, type to filter, Enter to select, Esc to cancel)")
	fmt.Println()

	// Use the client's SelectEKSClusterInteractive method
	cluster, err := client.SelectEKSClusterInteractive(ctx)
	if err != nil {
		// Check if it's a context cancellation (Ctrl+C)
		if err == context.Canceled {
			fmt.Println("\nSelection cancelled.")
			return nil, nil
		}
		return nil, fmt.Errorf("failed to select EKS cluster: %w", err)
	}

	if cluster == nil {
		// User cancelled the selection (Esc)
		fmt.Println("\nNo cluster selected.")
		return nil, nil
	}

	return cluster, nil
}

// displayClusterInfo displays comprehensive EKS cluster information
func displayClusterInfo(cluster *aws.Cluster) {
	separator := strings.Repeat("═", 80)
	fmt.Println("\n" + separator)
	fmt.Printf("EKS Cluster: %s\n", cluster.Name)
	fmt.Println(separator)

	// Basic Information
	fmt.Println("\n📋 Basic Information:")
	fmt.Printf("  Status:              %s\n", cluster.Status)
	fmt.Printf("  Version:             %s\n", cluster.Version)
	fmt.Printf("  Platform Version:    %s\n", cluster.PlatformVersion)
	fmt.Printf("  Created:             %s\n", cluster.CreatedAt.Format("2006-01-02 15:04:05"))
	fmt.Printf("  ARN:                 %s\n", cluster.ARN)
	fmt.Printf("  Role ARN:            %s\n", cluster.RoleARN)

	// API Server Endpoint
	fmt.Println("\n🔗 API Server:")
	fmt.Printf("  Endpoint:            %s\n", cluster.Endpoint)

	// Networking Configuration
	fmt.Println("\n🌐 Networking:")
	fmt.Printf("  VPC ID:              %s\n", cluster.VPC.VpcID)
	fmt.Printf("  Subnets:             %d\n", len(cluster.VPC.SubnetIDs))
	if len(cluster.VPC.SubnetIDs) > 0 {
		for i, subnet := range cluster.VPC.SubnetIDs {
			if i < 3 { // Show first 3
				fmt.Printf("    • %s\n", subnet)
			}
		}
		if len(cluster.VPC.SubnetIDs) > 3 {
			fmt.Printf("    • ... and %d more\n", len(cluster.VPC.SubnetIDs)-3)
		}
	}
	fmt.Printf("  Security Groups:     %d\n", len(cluster.VPC.SecurityGroupIDs))
	if len(cluster.VPC.SecurityGroupIDs) > 0 {
		for i, sg := range cluster.VPC.SecurityGroupIDs {
			if i < 3 { // Show first 3
				fmt.Printf("    • %s\n", sg)
			}
		}
		if len(cluster.VPC.SecurityGroupIDs) > 3 {
			fmt.Printf("    • ... and %d more\n", len(cluster.VPC.SecurityGroupIDs)-3)
		}
	}
	fmt.Printf("  Endpoint Private Access: %v\n", cluster.VPC.EndpointPrivateAccess)
	fmt.Printf("  Endpoint Public Access:  %v\n", cluster.VPC.EndpointPublicAccess)
	if len(cluster.VPC.PublicAccessCIDRs) > 0 {
		fmt.Printf("  Public Access CIDRs:     %v\n", cluster.VPC.PublicAccessCIDRs)
	}

	// Compute Resources
	fmt.Println("\n⚙️  Compute Resources:")
	fmt.Printf("  Node Groups:         %d\n", len(cluster.NodeGroups))
	if len(cluster.NodeGroups) > 0 {
		for _, ng := range cluster.NodeGroups {
			fmt.Printf("    • %s (%s) - %d/%d nodes\n", ng.Name, ng.Status, ng.CurrentSize, ng.DesiredSize)
		}
	}
	fmt.Printf("  Fargate Profiles:    %d\n", len(cluster.FargateProfiles))
	if len(cluster.FargateProfiles) > 0 {
		for _, fp := range cluster.FargateProfiles {
			fmt.Printf("    • %s (%s)\n", fp.Name, fp.Status)
		}
	}

	// Logging Configuration
	if len(cluster.Logging.ClusterLogging) > 0 {
		fmt.Println("\n📝 Logging:")
		for _, log := range cluster.Logging.ClusterLogging {
			status := "disabled"
			if log.Enabled {
				status = "enabled"
			}
			fmt.Printf("  • %s: %s\n", log.Type, status)
		}
	}

	// Encryption Configuration
	if len(cluster.EncryptionConfig) > 0 {
		fmt.Println("\n🔐 Encryption:")
		for _, enc := range cluster.EncryptionConfig {
			fmt.Printf("  Resources: %v\n", enc.Resources)
			fmt.Printf("  Key ARN:   %s\n", enc.Provider.KeyARN)
		}
	}

	// Identity Provider
	if cluster.Identity.OIDC.Issuer != "" {
		fmt.Println("\n🔑 Identity Provider:")
		fmt.Printf("  OIDC Issuer: %s\n", cluster.Identity.OIDC.Issuer)
	}

	// Tags
	if len(cluster.Tags) > 0 {
		fmt.Println("\n🏷️  Tags:")
		for key, value := range cluster.Tags {
			fmt.Printf("  %s: %s\n", key, value)
		}
	}

	fmt.Println("\n" + separator)
}
