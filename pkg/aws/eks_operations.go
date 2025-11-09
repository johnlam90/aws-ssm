package aws

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/eks"
	ekstypes "github.com/aws/aws-sdk-go-v2/service/eks/types"
)

// ListClusters retrieves all EKS clusters in the current region
func (c *Client) ListClusters(ctx context.Context) ([]string, error) {
	eksClient := eks.NewFromConfig(c.Config)

	var clusterNames []string
	paginator := eks.NewListClustersPaginator(eksClient, &eks.ListClustersInput{})

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to list EKS clusters: %w", err)
		}
		clusterNames = append(clusterNames, page.Clusters...)
	}

	return clusterNames, nil
}

// DescribeClusterBasic retrieves basic cluster information (without node groups/Fargate profiles)
// This is faster for initial loading in the fuzzy finder
func (c *Client) DescribeClusterBasic(ctx context.Context, clusterName string) (*Cluster, error) {
	eksClient := eks.NewFromConfig(c.Config)

	output, err := eksClient.DescribeCluster(ctx, &eks.DescribeClusterInput{
		Name: &clusterName,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to describe EKS cluster %s: %w", clusterName, err)
	}

	if output.Cluster == nil {
		return nil, fmt.Errorf("cluster %s not found", clusterName)
	}

	cluster := convertEKSCluster(output.Cluster)
	return cluster, nil
}

// DescribeCluster retrieves detailed information about an EKS cluster
// This includes node groups and Fargate profiles (slower but complete)
func (c *Client) DescribeCluster(ctx context.Context, clusterName string) (*Cluster, error) {
	eksClient := eks.NewFromConfig(c.Config)

	output, err := eksClient.DescribeCluster(ctx, &eks.DescribeClusterInput{
		Name: &clusterName,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to describe EKS cluster %s: %w", clusterName, err)
	}

	if output.Cluster == nil {
		return nil, fmt.Errorf("cluster %s not found", clusterName)
	}

	cluster := convertEKSCluster(output.Cluster)

	// Fetch node groups
	nodeGroups, err := c.listNodeGroups(ctx, eksClient, clusterName)
	if err != nil {
		// Log warning but continue
		fmt.Printf("Warning: failed to fetch node groups for cluster %s: %v\n", clusterName, err)
	}
	cluster.NodeGroups = nodeGroups

	// Fetch Fargate profiles
	fargateProfiles, err := c.listFargateProfiles(ctx, eksClient, clusterName)
	if err != nil {
		// Log warning but continue
		fmt.Printf("Warning: failed to fetch Fargate profiles for cluster %s: %v\n", clusterName, err)
	}
	cluster.FargateProfiles = fargateProfiles

	return cluster, nil
}

// listNodeGroups retrieves all node groups for a cluster
func (c *Client) listNodeGroups(ctx context.Context, eksClient *eks.Client, clusterName string) ([]NodeGroup, error) {
	var nodeGroups []NodeGroup

	input := &eks.ListNodegroupsInput{
		ClusterName: &clusterName,
	}

	paginator := eks.NewListNodegroupsPaginator(eksClient, input)

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to list node groups: %w", err)
		}

		for _, ngName := range page.Nodegroups {
			ng, err := c.describeNodeGroup(ctx, eksClient, clusterName, ngName)
			if err != nil {
				fmt.Printf("Warning: failed to describe node group %s: %v\n", ngName, err)
				continue
			}
			nodeGroups = append(nodeGroups, *ng)
		}
	}

	return nodeGroups, nil
}

// describeNodeGroup retrieves details about a specific node group
func (c *Client) describeNodeGroup(ctx context.Context, eksClient *eks.Client, clusterName, nodeGroupName string) (*NodeGroup, error) {
	output, err := eksClient.DescribeNodegroup(ctx, &eks.DescribeNodegroupInput{
		ClusterName:   &clusterName,
		NodegroupName: &nodeGroupName,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to describe node group %s: %w", nodeGroupName, err)
	}

	if output.Nodegroup == nil {
		return nil, fmt.Errorf("node group %s not found", nodeGroupName)
	}

	return convertNodeGroup(output.Nodegroup), nil
}

// listFargateProfiles retrieves all Fargate profiles for a cluster
func (c *Client) listFargateProfiles(ctx context.Context, eksClient *eks.Client, clusterName string) ([]FargateProfile, error) {
	var fargateProfiles []FargateProfile
	paginator := eks.NewListFargateProfilesPaginator(eksClient, &eks.ListFargateProfilesInput{
		ClusterName: &clusterName,
	})

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to list Fargate profiles: %w", err)
		}

		for _, fpName := range page.FargateProfileNames {
			fp, err := c.describeFargateProfile(ctx, eksClient, clusterName, fpName)
			if err != nil {
				fmt.Printf("Warning: failed to describe Fargate profile %s: %v\n", fpName, err)
				continue
			}
			fargateProfiles = append(fargateProfiles, *fp)
		}
	}

	return fargateProfiles, nil
}

// describeFargateProfile retrieves details about a specific Fargate profile
func (c *Client) describeFargateProfile(ctx context.Context, eksClient *eks.Client, clusterName, fargateProfileName string) (*FargateProfile, error) {
	output, err := eksClient.DescribeFargateProfile(ctx, &eks.DescribeFargateProfileInput{
		ClusterName:        &clusterName,
		FargateProfileName: &fargateProfileName,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to describe Fargate profile %s: %w", fargateProfileName, err)
	}

	if output.FargateProfile == nil {
		return nil, fmt.Errorf("Fargate profile %s not found", fargateProfileName)
	}

	return convertFargateProfile(output.FargateProfile), nil
}

// Conversion helper functions

func convertEKSCluster(eksCluster *ekstypes.Cluster) *Cluster {
	cluster := &Cluster{
		Name:            *eksCluster.Name,
		Status:          string(eksCluster.Status),
		Version:         *eksCluster.Version,
		Endpoint:        *eksCluster.Endpoint,
		RoleARN:         *eksCluster.RoleArn,
		CreatedAt:       *eksCluster.CreatedAt,
		Tags:            eksCluster.Tags,
		PlatformVersion: *eksCluster.PlatformVersion,
	}

	if eksCluster.Arn != nil {
		cluster.ARN = *eksCluster.Arn
	}

	// Convert VPC info
	if eksCluster.ResourcesVpcConfig != nil {
		cluster.VPC = VPCInfo{
			VpcID:                 *eksCluster.ResourcesVpcConfig.VpcId,
			SubnetIDs:             eksCluster.ResourcesVpcConfig.SubnetIds,
			SecurityGroupIDs:      eksCluster.ResourcesVpcConfig.SecurityGroupIds,
			PublicAccessCIDRs:     eksCluster.ResourcesVpcConfig.PublicAccessCidrs,
			EndpointPrivateAccess: eksCluster.ResourcesVpcConfig.EndpointPrivateAccess,
			EndpointPublicAccess:  eksCluster.ResourcesVpcConfig.EndpointPublicAccess,
		}
	}

	// Convert logging info
	if eksCluster.Logging != nil && eksCluster.Logging.ClusterLogging != nil {
		for _, cl := range eksCluster.Logging.ClusterLogging {
			if len(cl.Types) > 0 {
				enabled := false
				if cl.Enabled != nil {
					enabled = *cl.Enabled
				}
				for _, t := range cl.Types {
					cluster.Logging.ClusterLogging = append(cluster.Logging.ClusterLogging, LoggingType{
						Type:    string(t),
						Enabled: enabled,
					})
				}
			}
		}
	}

	// Convert encryption config
	if eksCluster.EncryptionConfig != nil {
		for _, ec := range eksCluster.EncryptionConfig {
			encConfig := EncryptionConfig{}
			if ec.Resources != nil {
				encConfig.Resources = ec.Resources
			}
			if ec.Provider != nil && ec.Provider.KeyArn != nil {
				encConfig.Provider = EncryptionProvider{
					KeyARN: *ec.Provider.KeyArn,
				}
			}
			cluster.EncryptionConfig = append(cluster.EncryptionConfig, encConfig)
		}
	}

	// Convert identity info
	if eksCluster.Identity != nil && eksCluster.Identity.Oidc != nil && eksCluster.Identity.Oidc.Issuer != nil {
		cluster.Identity = IdentityInfo{
			OIDC: OIDCInfo{
				Issuer: *eksCluster.Identity.Oidc.Issuer,
			},
		}
	}

	// Convert certificate authority
	if eksCluster.CertificateAuthority != nil && eksCluster.CertificateAuthority.Data != nil {
		cluster.CertificateAuthority = CertificateAuthority{
			Data: *eksCluster.CertificateAuthority.Data,
		}
	}

	return cluster
}

func convertNodeGroup(ng *ekstypes.Nodegroup) *NodeGroup {
	if ng == nil {
		return nil
	}

	nodeGroup := &NodeGroup{
		Status: string(ng.Status),
		Tags:   ng.Tags,
		Labels: ng.Labels,
	}

	// Safe pointer dereferences with nil checks
	if ng.NodegroupName != nil {
		nodeGroup.Name = *ng.NodegroupName
	}
	if ng.Version != nil {
		nodeGroup.Version = *ng.Version
	}
	if ng.DiskSize != nil {
		nodeGroup.DiskSize = *ng.DiskSize
	}
	if ng.CreatedAt != nil {
		nodeGroup.CreatedAt = *ng.CreatedAt
	}

	// Handle InstanceTypes
	nodeGroup.InstanceTypes = ng.InstanceTypes

	// Handle ScalingConfig
	if ng.ScalingConfig != nil {
		if ng.ScalingConfig.DesiredSize != nil {
			nodeGroup.DesiredSize = *ng.ScalingConfig.DesiredSize
			// Use DesiredSize as CurrentSize (actual current count would require EC2 ASG API call)
			nodeGroup.CurrentSize = *ng.ScalingConfig.DesiredSize
		}
		if ng.ScalingConfig.MinSize != nil {
			nodeGroup.MinSize = *ng.ScalingConfig.MinSize
		}
		if ng.ScalingConfig.MaxSize != nil {
			nodeGroup.MaxSize = *ng.ScalingConfig.MaxSize
		}
	}

	if ng.NodegroupArn != nil {
		nodeGroup.NodeGroupARN = *ng.NodegroupArn
	}

	if ng.LaunchTemplate != nil {
		launchTemplate := LaunchTemplateInfo{}
		if ng.LaunchTemplate.Id != nil {
			launchTemplate.ID = *ng.LaunchTemplate.Id
		}
		if ng.LaunchTemplate.Name != nil {
			launchTemplate.Name = *ng.LaunchTemplate.Name
		}
		if ng.LaunchTemplate.Version != nil {
			launchTemplate.Version = *ng.LaunchTemplate.Version
		}
		nodeGroup.LaunchTemplate = launchTemplate
	}

	if ng.Taints != nil {
		for _, t := range ng.Taints {
			taint := Taint{Effect: string(t.Effect)}
			if t.Key != nil {
				taint.Key = *t.Key
			}
			if t.Value != nil {
				taint.Value = *t.Value
			}
			nodeGroup.Taints = append(nodeGroup.Taints, taint)
		}
	}

	return nodeGroup
}

// ListNodeGroupsForCluster retrieves all node group names for a cluster (public method)
func (c *Client) ListNodeGroupsForCluster(ctx context.Context, clusterName string) ([]string, error) {
	eksClient := eks.NewFromConfig(c.Config)

	var nodeGroupNames []string
	input := &eks.ListNodegroupsInput{
		ClusterName: &clusterName,
	}

	paginator := eks.NewListNodegroupsPaginator(eksClient, input)

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to list node groups: %w", err)
		}
		nodeGroupNames = append(nodeGroupNames, page.Nodegroups...)
	}

	return nodeGroupNames, nil
}

// DescribeNodeGroupPublic retrieves details about a specific node group (public method)
func (c *Client) DescribeNodeGroupPublic(ctx context.Context, clusterName, nodeGroupName string) (*NodeGroup, error) {
	eksClient := eks.NewFromConfig(c.Config)
	return c.describeNodeGroup(ctx, eksClient, clusterName, nodeGroupName)
}

// UpdateNodeGroupScaling updates the scaling configuration of a node group
func (c *Client) UpdateNodeGroupScaling(ctx context.Context, clusterName, nodeGroupName string, minSize, maxSize, desiredSize int32) error {
	eksClient := eks.NewFromConfig(c.Config)

	// Validate scaling parameters
	if minSize < 0 {
		return fmt.Errorf("min size cannot be negative")
	}
	if maxSize < minSize {
		return fmt.Errorf("max size (%d) cannot be less than min size (%d)", maxSize, minSize)
	}
	if desiredSize < minSize || desiredSize > maxSize {
		return fmt.Errorf("desired size (%d) must be between min size (%d) and max size (%d)", desiredSize, minSize, maxSize)
	}

	input := &eks.UpdateNodegroupConfigInput{
		ClusterName:   &clusterName,
		NodegroupName: &nodeGroupName,
		ScalingConfig: &ekstypes.NodegroupScalingConfig{
			MinSize:     &minSize,
			MaxSize:     &maxSize,
			DesiredSize: &desiredSize,
		},
	}

	_, err := eksClient.UpdateNodegroupConfig(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to update node group scaling: %w", err)
	}

	return nil
}

func convertFargateProfile(fp *ekstypes.FargateProfile) *FargateProfile {
	if fp == nil {
		return nil
	}

	fargateProfile := &FargateProfile{
		Status: string(fp.Status),
		Tags:   fp.Tags,
	}

	// Safe pointer dereferences with nil checks
	if fp.FargateProfileName != nil {
		fargateProfile.Name = *fp.FargateProfileName
	}
	if fp.CreatedAt != nil {
		fargateProfile.CreatedAt = *fp.CreatedAt
	}

	if fp.FargateProfileArn != nil {
		fargateProfile.FargateProfileARN = *fp.FargateProfileArn
	}

	if fp.PodExecutionRoleArn != nil {
		fargateProfile.PodExecutionRoleARN = *fp.PodExecutionRoleArn
	}

	if fp.Selectors != nil {
		for _, sel := range fp.Selectors {
			selector := FargateSelector{Labels: sel.Labels}
			if sel.Namespace != nil {
				selector.Namespace = *sel.Namespace
			}
			fargateProfile.Selectors = append(fargateProfile.Selectors, selector)
		}
	}

	return fargateProfile
}
