package tui

import (
	"fmt"
	"strings"

	"github.com/johnlam90/aws-ssm/pkg/aws"
)

// DisplayClusterInfo displays comprehensive EKS cluster information
func DisplayClusterInfo(cluster *aws.Cluster) {
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
	fmt.Println("\nBasic Information:")
	fmt.Printf("  Status:              %s\n", cluster.Status)
	fmt.Printf("  Version:             %s\n", cluster.Version)
	fmt.Printf("  Platform Version:    %s\n", cluster.PlatformVersion)
	fmt.Printf("  Created:             %s\n", cluster.CreatedAt.Format("2006-01-02 15:04:05"))
	fmt.Printf("  ARN:                 %s\n", cluster.ARN)
	fmt.Printf("  Role ARN:            %s\n", cluster.RoleARN)
}

// displayAPIEndpoint displays API server information
func displayAPIEndpoint(cluster *aws.Cluster) {
	fmt.Println("\nAPI Server:")
	fmt.Printf("  Endpoint:            %s\n", cluster.Endpoint)
}

// displayNetworking displays networking configuration
func displayNetworking(cluster *aws.Cluster) {
	fmt.Println("\nNetworking:")
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
	fmt.Println("\nCompute Resources:")

	// Node Groups
	if len(cluster.NodeGroups) > 0 {
		fmt.Printf("  Node Groups:         %d\n", len(cluster.NodeGroups))
		for _, ng := range cluster.NodeGroups {
			fmt.Printf("    • %s (Status: %s, Desired: %d, Min: %d, Max: %d)\n",
				ng.Name, ng.Status, ng.DesiredSize, ng.MinSize, ng.MaxSize)
			if len(ng.InstanceTypes) > 0 {
				fmt.Printf("      Instance Types: %v\n", ng.InstanceTypes)
			}
		}
	} else {
		fmt.Println("  Node Groups:         None")
	}

	// Fargate Profiles
	if len(cluster.FargateProfiles) > 0 {
		fmt.Printf("  Fargate Profiles:    %d\n", len(cluster.FargateProfiles))
		for _, fp := range cluster.FargateProfiles {
			fmt.Printf("    • %s (Status: %s)\n", fp.Name, fp.Status)
		}
	} else {
		fmt.Println("  Fargate Profiles:    None")
	}
}

// displayLogging displays logging configuration
func displayLogging(cluster *aws.Cluster) {
	if len(cluster.Logging.ClusterLogging) == 0 {
		return
	}

	fmt.Println("\nLogging:")
	for _, log := range cluster.Logging.ClusterLogging {
		if log.Enabled {
			fmt.Printf("  %s: Enabled\n", log.Type)
		}
	}
}

// displayEncryption displays encryption configuration
func displayEncryption(cluster *aws.Cluster) {
	if len(cluster.EncryptionConfig) == 0 {
		return
	}

	fmt.Println("\nEncryption:")
	for _, enc := range cluster.EncryptionConfig {
		fmt.Printf("  Resources: %v\n", enc.Resources)
		fmt.Printf("  Key ARN:   %s\n", enc.Provider.KeyARN)
	}
}

// displayIdentityProvider displays identity provider configuration
func displayIdentityProvider(cluster *aws.Cluster) {
	if len(cluster.Identity.OIDC.Issuer) == 0 {
		return
	}

	fmt.Println("\nIdentity Provider:")
	fmt.Printf("  OIDC Issuer: %s\n", cluster.Identity.OIDC.Issuer)
}

// displayTags displays cluster tags
func displayTags(cluster *aws.Cluster) {
	if len(cluster.Tags) == 0 {
		return
	}

	fmt.Println("\nTags:")
	for key, value := range cluster.Tags {
		fmt.Printf("  %s: %s\n", key, value)
	}
}
