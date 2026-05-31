package fuzzy

import (
	"context"
	"fmt"
	"strings"
)

// EKSPreviewRenderer handles rendering EKS cluster preview information
type EKSPreviewRenderer struct {
	colors ColorManager
	loader EKSClusterLoader // For lazy loading full cluster details
}

// NewEKSPreviewRenderer creates a new EKS preview renderer
func NewEKSPreviewRenderer(colors ColorManager, loader EKSClusterLoader) *EKSPreviewRenderer {
	return &EKSPreviewRenderer{
		colors: colors,
		loader: loader,
	}
}

// RenderWithLazyLoad renders the preview with lazy loading of full cluster details
func (r *EKSPreviewRenderer) RenderWithLazyLoad(ctx context.Context, cluster *EKSCluster, width, height int) string {
	if cluster == nil {
		return ""
	}

	// If cluster has no details yet (only name), fetch them lazily
	if cluster.Status == "" && cluster.Version == "" {
		// Try to get full details from loader
		if awsLoader, ok := r.loader.(*AWSEKSLoader); ok {
			fullCluster, err := awsLoader.GetClusterDetails(ctx, cluster.Name)
			if err != nil {
				// If fetch fails, show basic info with error
				return r.renderBasicWithError(cluster, width, err)
			}
			// Use the full cluster details for rendering
			cluster = fullCluster
		}
	}

	// Render full cluster details
	return r.Render(cluster, width, height)
}

// renderBasicWithError renders basic cluster info when full details can't be loaded
func (r *EKSPreviewRenderer) renderBasicWithError(cluster *EKSCluster, width int, err error) string {
	var preview strings.Builder

	preview.WriteString(r.colors.HeaderColor("EKS Cluster Details"))
	preview.WriteString("\n")
	preview.WriteString(strings.Repeat("─", minimum(width, 60)))
	preview.WriteString("\n\n")

	preview.WriteString(r.colors.BoldColor("Basic Information:"))
	preview.WriteString("\n")
	fmt.Fprintf(&preview, "  Name:              %s\n", cluster.Name)
	preview.WriteString("\n")

	preview.WriteString("⚠ Failed to load full cluster details\n")
	fmt.Fprintf(&preview, "  Error: %v\n", err)

	return preview.String()
}

// Render renders the preview for an EKS cluster
func (r *EKSPreviewRenderer) Render(cluster *EKSCluster, width, _ int) string {
	if cluster == nil {
		return ""
	}

	var preview strings.Builder

	// Header
	preview.WriteString(r.colors.HeaderColor("EKS Cluster Details"))
	preview.WriteString("\n")
	preview.WriteString(strings.Repeat("─", minimum(width, 60)))
	preview.WriteString("\n\n")

	// Basic Information
	preview.WriteString(r.colors.BoldColor("Basic Information:"))
	preview.WriteString("\n")
	fmt.Fprintf(&preview, "  Name:              %s\n", cluster.Name)
	fmt.Fprintf(&preview, "  Status:            %s\n", r.colors.StatusColor(cluster.Status))
	fmt.Fprintf(&preview, "  Version:           %s\n", cluster.Version)
	fmt.Fprintf(&preview, "  Created:           %s\n", cluster.CreatedAt)
	fmt.Fprintf(&preview, "  ARN:               %s\n", cluster.ARN)
	preview.WriteString("\n")

	// API Server Endpoint
	preview.WriteString(r.colors.BoldColor("API Server:"))
	preview.WriteString("\n")
	fmt.Fprintf(&preview, "  Endpoint:          %s\n", cluster.Endpoint)
	preview.WriteString("\n")

	// Networking Configuration
	preview.WriteString(r.colors.BoldColor("Networking:"))
	preview.WriteString("\n")
	fmt.Fprintf(&preview, "  VPC ID:            %s\n", cluster.VpcID)
	fmt.Fprintf(&preview, "  Subnets:           %d\n", cluster.SubnetCount)
	fmt.Fprintf(&preview, "  Security Groups:   %d\n", cluster.SecurityGroupCount)
	preview.WriteString("\n")

	// Compute Resources
	preview.WriteString(r.colors.BoldColor("Compute Resources:"))
	preview.WriteString("\n")
	fmt.Fprintf(&preview, "  Node Groups:       %d\n", cluster.NodeGroupCount)
	fmt.Fprintf(&preview, "  Fargate Profiles:  %d\n", cluster.FargateProfileCount)
	preview.WriteString("\n")

	// Tags
	if len(cluster.Tags) > 0 {
		preview.WriteString(r.colors.BoldColor("Tags:"))
		preview.WriteString("\n")
		for key, value := range cluster.Tags {
			fmt.Fprintf(&preview, "  %s\n", r.colors.TagColor(key, value))
		}
	}

	return preview.String()
}

// RenderDetailed renders detailed EKS cluster information (for full cluster details)
func (r *EKSPreviewRenderer) RenderDetailed(cluster *EKSCluster, nodeGroups []string, fargateProfiles []string, width, _ int) string {
	if cluster == nil {
		return ""
	}

	var preview strings.Builder

	// Header
	preview.WriteString(r.colors.HeaderColor("EKS Cluster Details"))
	preview.WriteString("\n")
	preview.WriteString(strings.Repeat("─", minimum(width, 60)))
	preview.WriteString("\n\n")

	// Basic Information
	preview.WriteString(r.colors.BoldColor("Basic Information:"))
	preview.WriteString("\n")
	fmt.Fprintf(&preview, "  Name:              %s\n", cluster.Name)
	fmt.Fprintf(&preview, "  Status:            %s\n", r.colors.StatusColor(cluster.Status))
	fmt.Fprintf(&preview, "  Version:           %s\n", cluster.Version)
	fmt.Fprintf(&preview, "  Created:           %s\n", cluster.CreatedAt)
	fmt.Fprintf(&preview, "  ARN:               %s\n", cluster.ARN)
	preview.WriteString("\n")

	// API Server Endpoint
	preview.WriteString(r.colors.BoldColor("API Server:"))
	preview.WriteString("\n")
	fmt.Fprintf(&preview, "  Endpoint:          %s\n", cluster.Endpoint)
	preview.WriteString("\n")

	// Networking Configuration
	preview.WriteString(r.colors.BoldColor("Networking:"))
	preview.WriteString("\n")
	fmt.Fprintf(&preview, "  VPC ID:            %s\n", cluster.VpcID)
	fmt.Fprintf(&preview, "  Subnets:           %d\n", cluster.SubnetCount)
	fmt.Fprintf(&preview, "  Security Groups:   %d\n", cluster.SecurityGroupCount)
	preview.WriteString("\n")

	// Node Groups
	if len(nodeGroups) > 0 {
		preview.WriteString(r.colors.BoldColor("Node Groups:"))
		preview.WriteString("\n")
		for _, ng := range nodeGroups {
			fmt.Fprintf(&preview, "  • %s\n", ng)
		}
		preview.WriteString("\n")
	}

	// Fargate Profiles
	if len(fargateProfiles) > 0 {
		preview.WriteString(r.colors.BoldColor("Fargate Profiles:"))
		preview.WriteString("\n")
		for _, fp := range fargateProfiles {
			fmt.Fprintf(&preview, "  • %s\n", fp)
		}
		preview.WriteString("\n")
	}

	// Tags
	if len(cluster.Tags) > 0 {
		preview.WriteString(r.colors.BoldColor("Tags:"))
		preview.WriteString("\n")
		for key, value := range cluster.Tags {
			fmt.Fprintf(&preview, "  %s\n", r.colors.TagColor(key, value))
		}
	}

	return preview.String()
}
