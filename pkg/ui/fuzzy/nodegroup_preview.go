package fuzzy

import (
	"fmt"
	"strings"
	"time"
)

// NodeGroupPreviewRenderer renders node group preview
type NodeGroupPreviewRenderer struct {
	colors ColorManager
}

// NewNodeGroupPreviewRenderer creates a new node group preview renderer
func NewNodeGroupPreviewRenderer(colors ColorManager) *NodeGroupPreviewRenderer {
	return &NodeGroupPreviewRenderer{colors: colors}
}

// Render renders the preview for a node group
func (r *NodeGroupPreviewRenderer) Render(ng *NodeGroupInfo, width, height int) string {
	if ng == nil {
		return ""
	}

	var preview strings.Builder

	// Header
	preview.WriteString(r.colors.HeaderColor("Node Group Details"))
	preview.WriteString("\n")
	preview.WriteString(strings.Repeat("─", min(width, 60)))
	preview.WriteString("\n\n")

	// Basic information
	preview.WriteString(r.colors.BoldColor("Basic Information:"))
	preview.WriteString("\n")
	preview.WriteString(fmt.Sprintf("  Name:              %s\n", ng.Name))
	preview.WriteString(fmt.Sprintf("  Cluster:           %s\n", ng.ClusterName))
	preview.WriteString(fmt.Sprintf("  Status:            %s\n", r.formatStatus(ng.Status)))
	preview.WriteString(fmt.Sprintf("  Version:           %s\n", ng.Version))

	if !ng.CreatedAt.IsZero() {
		age := time.Since(ng.CreatedAt)
		preview.WriteString(fmt.Sprintf("  Created:           %s (%s ago)\n",
			ng.CreatedAt.Format("2006-01-02 15:04:05"),
			r.formatDuration(age)))
	}

	preview.WriteString("\n")

	// Scaling configuration
	preview.WriteString(r.colors.BoldColor("Scaling Configuration:"))
	preview.WriteString("\n")
	preview.WriteString(fmt.Sprintf("  Desired Size:      %d\n", ng.DesiredSize))
	preview.WriteString(fmt.Sprintf("  Min Size:          %d\n", ng.MinSize))
	preview.WriteString(fmt.Sprintf("  Max Size:          %d\n", ng.MaxSize))
	preview.WriteString(fmt.Sprintf("  Current Size:      %d\n", ng.CurrentSize))

	// Show scaling status
	switch {
	case ng.CurrentSize < ng.DesiredSize:
		preview.WriteString(fmt.Sprintf("  %s Scaling up (%d → %d)\n",
			r.colors.WarningColor("⚠"),
			ng.CurrentSize,
			ng.DesiredSize))
	case ng.CurrentSize > ng.DesiredSize:
		preview.WriteString(fmt.Sprintf("  %s Scaling down (%d → %d)\n",
			r.colors.WarningColor("⚠"),
			ng.CurrentSize,
			ng.DesiredSize))
	case ng.CurrentSize == ng.DesiredSize && ng.DesiredSize > 0:
		preview.WriteString(fmt.Sprintf("  %s At desired capacity\n",
			r.colors.SuccessColor("✓")))
	}

	preview.WriteString("\n")

	// Instance types
	preview.WriteString(r.colors.BoldColor("Instance Configuration:"))
	preview.WriteString("\n")
	switch {
	case len(ng.InstanceTypes) > 0:
		preview.WriteString("  Instance Types:\n")
		for _, instanceType := range ng.InstanceTypes {
			preview.WriteString(fmt.Sprintf("    • %s\n", instanceType))
		}
	case ng.LaunchTemplateName != "":
		preview.WriteString(fmt.Sprintf("  Launch Template:   %s\n", ng.LaunchTemplateName))
		preview.WriteString("  Instance Types:    (defined in launch template)\n")
	default:
		preview.WriteString("  Instance Types:    Not specified\n")
	}

	// Tags
	if len(ng.Tags) > 0 {
		preview.WriteString("\n")
		preview.WriteString(r.colors.BoldColor("Tags:"))
		preview.WriteString("\n")
		for key, value := range ng.Tags {
			preview.WriteString(fmt.Sprintf("  %s\n", r.colors.TagColor(key, value)))
		}
	}

	return preview.String()
}

// formatStatus formats the status with color
func (r *NodeGroupPreviewRenderer) formatStatus(status string) string {
	switch status {
	case "ACTIVE":
		return r.colors.SuccessColor(status)
	case "CREATING", "UPDATING":
		return r.colors.WarningColor(status)
	case "DELETING", "DELETE_FAILED", "DEGRADED":
		return r.colors.ErrorColor(status)
	default:
		return status
	}
}

// formatDuration formats a duration in a human-readable way
func (r *NodeGroupPreviewRenderer) formatDuration(d time.Duration) string {
	days := int(d.Hours()) / 24
	hours := int(d.Hours()) % 24
	minutes := int(d.Minutes()) % 60

	var parts []string
	if days > 0 {
		parts = append(parts, fmt.Sprintf("%dd", days))
	}
	if hours > 0 {
		parts = append(parts, fmt.Sprintf("%dh", hours))
	}
	if minutes > 0 {
		parts = append(parts, fmt.Sprintf("%dm", minutes))
	}

	if len(parts) == 0 {
		return "less than 1 minute"
	}

	return strings.Join(parts, " ")
}
