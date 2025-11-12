package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/johnlam90/aws-ssm/pkg/aws"
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

func runEKS(_ *cobra.Command, args []string) error {
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
		cluster, selectionErr := selectEKSClusterInteractive(ctx, client)
		if selectionErr != nil {
			return fmt.Errorf("failed to select EKS cluster: %w", selectionErr)
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
	// Show loading message with spinner
	s := createLoadingSpinner("Loading EKS clusters...")
	s.Start()

	// Load clusters first (this is the slow part)
	clusterNames, err := client.ListClusters(ctx)
	s.Stop()

	if err != nil {
		return nil, fmt.Errorf("failed to list EKS clusters: %w", err)
	}

	if len(clusterNames) == 0 {
		return nil, fmt.Errorf("no EKS clusters found")
	}

	// Now show the interactive prompt
	printInteractivePrompt("EKS cluster selector")
	fmt.Println()

	// Use the client's SelectEKSClusterInteractive method
	cluster, err := client.SelectEKSClusterInteractive(ctx)

	if err != nil {
		// Check if it's a context cancellation (Ctrl+C)
		if err == context.Canceled {
			printSelectionCancelled()
			return nil, nil
		}
		return nil, fmt.Errorf("failed to select EKS cluster: %w", err)
	}

	if cluster == nil {
		// User cancelled the selection (Esc)
		printNoSelection("cluster")
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

	// Display each section
	displayBasicInfo(cluster)
	displayAPIEndpoint(cluster)
	displayNetworking(cluster)
	displayComputeResources(cluster)
	displayLogging(cluster)
	displayEncryption(cluster)
	displayIdentityProvider(cluster)
	displayTags(cluster)

	fmt.Println("\n" + separator)
}

// displayBasicInfo displays basic cluster information
func displayBasicInfo(cluster *aws.Cluster) {
	fmt.Println("\n📋 Basic Information:")
	fmt.Printf("  Status:              %s\n", cluster.Status)
	fmt.Printf("  Version:             %s\n", cluster.Version)
	fmt.Printf("  Platform Version:    %s\n", cluster.PlatformVersion)
	fmt.Printf("  Created:             %s\n", cluster.CreatedAt.Format("2006-01-02 15:04:05"))
	fmt.Printf("  ARN:                 %s\n", cluster.ARN)
	fmt.Printf("  Role ARN:            %s\n", cluster.RoleARN)
}

// displayAPIEndpoint displays API server information
func displayAPIEndpoint(cluster *aws.Cluster) {
	fmt.Println("\n🔗 API Server:")
	fmt.Printf("  Endpoint:            %s\n", cluster.Endpoint)
}

// displayNetworking displays networking configuration
func displayNetworking(cluster *aws.Cluster) {
	fmt.Println("\n🌐 Networking:")
	fmt.Printf("  VPC ID:              %s\n", cluster.VPC.VpcID)

	displaySubnets(cluster.VPC.SubnetIDs)
	displaySecurityGroups(cluster.VPC.SecurityGroupIDs)
	displayEndpointAccess(cluster.VPC)
}

// displaySubnets displays subnet information
func displaySubnets(subnets []string) {
	fmt.Printf("  Subnets:             %d\n", len(subnets))
	if len(subnets) > 0 {
		for i, subnet := range subnets {
			if i < 3 { // Show first 3
				fmt.Printf("    • %s\n", subnet)
			}
		}
		if len(subnets) > 3 {
			fmt.Printf("    • ... and %d more\n", len(subnets)-3)
		}
	}
}

// displaySecurityGroups displays security group information
func displaySecurityGroups(sgs []string) {
	fmt.Printf("  Security Groups:     %d\n", len(sgs))
	if len(sgs) > 0 {
		for i, sg := range sgs {
			if i < 3 { // Show first 3
				fmt.Printf("    • %s\n", sg)
			}
		}
		if len(sgs) > 3 {
			fmt.Printf("    • ... and %d more\n", len(sgs)-3)
		}
	}
}

// displayEndpointAccess displays endpoint access information
func displayEndpointAccess(vpc aws.VPCInfo) {
	fmt.Printf("  Endpoint Private Access: %v\n", vpc.EndpointPrivateAccess)
	fmt.Printf("  Endpoint Public Access:  %v\n", vpc.EndpointPublicAccess)
	if len(vpc.PublicAccessCIDRs) > 0 {
		fmt.Printf("  Public Access CIDRs:     %v\n", vpc.PublicAccessCIDRs)
	}
}

// displayComputeResources displays node groups and Fargate profiles
func displayComputeResources(cluster *aws.Cluster) {
	fmt.Println("\n⚙️  Compute Resources:")

	displayNodeGroups(cluster.NodeGroups)
	displayFargateProfiles(cluster.FargateProfiles)
}

// displayNodeGroups displays node group information
func displayNodeGroups(ngs []aws.NodeGroup) {
	fmt.Printf("  Node Groups:         %d\n", len(ngs))
	if len(ngs) > 0 {
		for _, ng := range ngs {
			fmt.Printf("    • %s (%s) - %d/%d nodes\n", ng.Name, ng.Status, ng.CurrentSize, ng.DesiredSize)
		}
	}
}

// displayFargateProfiles displays Fargate profile information
func displayFargateProfiles(fps []aws.FargateProfile) {
	fmt.Printf("  Fargate Profiles:    %d\n", len(fps))
	if len(fps) > 0 {
		for _, fp := range fps {
			fmt.Printf("    • %s (%s)\n", fp.Name, fp.Status)
		}
	}
}

// displayLogging displays logging configuration
func displayLogging(cluster *aws.Cluster) {
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
}

// displayEncryption displays encryption configuration
func displayEncryption(cluster *aws.Cluster) {
	if len(cluster.EncryptionConfig) > 0 {
		fmt.Println("\n🔐 Encryption:")
		for _, enc := range cluster.EncryptionConfig {
			fmt.Printf("  Resources: %v\n", enc.Resources)
			fmt.Printf("  Key ARN:   %s\n", enc.Provider.KeyARN)
		}
	}
}

// displayIdentityProvider displays identity provider information
func displayIdentityProvider(cluster *aws.Cluster) {
	if cluster.Identity.OIDC.Issuer != "" {
		fmt.Println("\n🔑 Identity Provider:")
		fmt.Printf("  OIDC Issuer: %s\n", cluster.Identity.OIDC.Issuer)
	}
}

// displayTags displays cluster tags
func displayTags(cluster *aws.Cluster) {
	if len(cluster.Tags) > 0 {
		fmt.Println("\n🏷️  Tags:")
		for key, value := range cluster.Tags {
			fmt.Printf("  %s: %s\n", key, value)
		}
	}
}
