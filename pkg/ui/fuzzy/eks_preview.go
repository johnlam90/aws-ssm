package fuzzy

import (
	"fmt"
	"strings"
)

// EKSPreviewRenderer handles rendering EKS cluster preview information
type EKSPreviewRenderer struct {
	colors ColorManager
}

// NewEKSPreviewRenderer creates a new EKS preview renderer
func NewEKSPreviewRenderer(colors ColorManager) *EKSPreviewRenderer {
	return &EKSPreviewRenderer{colors: colors}
}

// Render renders the preview for an EKS cluster
func (r *EKSPreviewRenderer) Render(cluster *EKSCluster, width, height int) string {
	if cluster == nil {
		return ""
	}

	var preview strings.Builder

	// Header
	preview.WriteString(r.colors.HeaderColor("EKS Cluster Details"))
	preview.WriteString("\n")
	preview.WriteString(strings.Repeat("─", min(width, 60)))
	preview.WriteString("\n\n")

	// Basic Information
	preview.WriteString(r.colors.BoldColor("Basic Information:"))
	preview.WriteString("\n")
	preview.WriteString(fmt.Sprintf("  Name:              %s\n", cluster.Name))
	preview.WriteString(fmt.Sprintf("  Status:            %s\n", r.colors.StatusColor(cluster.Status)))
	preview.WriteString(fmt.Sprintf("  Version:           %s\n", cluster.Version))
	preview.WriteString(fmt.Sprintf("  Created:           %s\n", cluster.CreatedAt))
	preview.WriteString(fmt.Sprintf("  ARN:               %s\n", cluster.ARN))
	preview.WriteString("\n")

	// API Server Endpoint
	preview.WriteString(r.colors.BoldColor("API Server:"))
	preview.WriteString("\n")
	preview.WriteString(fmt.Sprintf("  Endpoint:          %s\n", cluster.Endpoint))
	preview.WriteString("\n")

	// Networking Configuration
	preview.WriteString(r.colors.BoldColor("Networking:"))
	preview.WriteString("\n")
	preview.WriteString(fmt.Sprintf("  VPC ID:            %s\n", cluster.VpcID))
	preview.WriteString(fmt.Sprintf("  Subnets:           %d\n", cluster.SubnetCount))
	preview.WriteString(fmt.Sprintf("  Security Groups:   %d\n", cluster.SecurityGroupCount))
	preview.WriteString("\n")

	// Compute Resources
	preview.WriteString(r.colors.BoldColor("Compute Resources:"))
	preview.WriteString("\n")
	preview.WriteString(fmt.Sprintf("  Node Groups:       %d\n", cluster.NodeGroupCount))
	preview.WriteString(fmt.Sprintf("  Fargate Profiles:  %d\n", cluster.FargateProfileCount))
	preview.WriteString("\n")

	// Tags
	if len(cluster.Tags) > 0 {
		preview.WriteString(r.colors.BoldColor("Tags:"))
		preview.WriteString("\n")
		for key, value := range cluster.Tags {
			preview.WriteString(fmt.Sprintf("  %s\n", r.colors.TagColor(key, value)))
		}
	}

	return preview.String()
}

// RenderDetailed renders detailed EKS cluster information (for full cluster details)
func (r *EKSPreviewRenderer) RenderDetailed(cluster *EKSCluster, nodeGroups []string, fargateProfiles []string, width, height int) string {
	if cluster == nil {
		return ""
	}

	var preview strings.Builder

	// Header
	preview.WriteString(r.colors.HeaderColor("EKS Cluster Details"))
	preview.WriteString("\n")
	preview.WriteString(strings.Repeat("─", min(width, 60)))
	preview.WriteString("\n\n")

	// Basic Information
	preview.WriteString(r.colors.BoldColor("Basic Information:"))
	preview.WriteString("\n")
	preview.WriteString(fmt.Sprintf("  Name:              %s\n", cluster.Name))
	preview.WriteString(fmt.Sprintf("  Status:            %s\n", r.colors.StatusColor(cluster.Status)))
	preview.WriteString(fmt.Sprintf("  Version:           %s\n", cluster.Version))
	preview.WriteString(fmt.Sprintf("  Created:           %s\n", cluster.CreatedAt))
	preview.WriteString(fmt.Sprintf("  ARN:               %s\n", cluster.ARN))
	preview.WriteString("\n")

	// API Server Endpoint
	preview.WriteString(r.colors.BoldColor("API Server:"))
	preview.WriteString("\n")
	preview.WriteString(fmt.Sprintf("  Endpoint:          %s\n", cluster.Endpoint))
	preview.WriteString("\n")

	// Networking Configuration
	preview.WriteString(r.colors.BoldColor("Networking:"))
	preview.WriteString("\n")
	preview.WriteString(fmt.Sprintf("  VPC ID:            %s\n", cluster.VpcID))
	preview.WriteString(fmt.Sprintf("  Subnets:           %d\n", cluster.SubnetCount))
	preview.WriteString(fmt.Sprintf("  Security Groups:   %d\n", cluster.SecurityGroupCount))
	preview.WriteString("\n")

	// Node Groups
	if len(nodeGroups) > 0 {
		preview.WriteString(r.colors.BoldColor("Node Groups:"))
		preview.WriteString("\n")
		for _, ng := range nodeGroups {
			preview.WriteString(fmt.Sprintf("  • %s\n", ng))
		}
		preview.WriteString("\n")
	}

	// Fargate Profiles
	if len(fargateProfiles) > 0 {
		preview.WriteString(r.colors.BoldColor("Fargate Profiles:"))
		preview.WriteString("\n")
		for _, fp := range fargateProfiles {
			preview.WriteString(fmt.Sprintf("  • %s\n", fp))
		}
		preview.WriteString("\n")
	}

	// Tags
	if len(cluster.Tags) > 0 {
		preview.WriteString(r.colors.BoldColor("Tags:"))
		preview.WriteString("\n")
		for key, value := range cluster.Tags {
			preview.WriteString(fmt.Sprintf("  %s\n", r.colors.TagColor(key, value)))
		}
	}

	return preview.String()
}
