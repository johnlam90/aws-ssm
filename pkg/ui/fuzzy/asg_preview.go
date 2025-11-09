package fuzzy

import (
	"fmt"
	"strings"
	"time"
)

// ASGPreviewRenderer renders ASG preview
type ASGPreviewRenderer struct {
	colors ColorManager
}

// NewASGPreviewRenderer creates a new ASG preview renderer
func NewASGPreviewRenderer(colors ColorManager) *ASGPreviewRenderer {
	return &ASGPreviewRenderer{colors: colors}
}

// Render renders the preview for an Auto Scaling Group
func (r *ASGPreviewRenderer) Render(asg *ASGInfo, width, height int) string {
	if asg == nil {
		return ""
	}

	var preview strings.Builder

	// Header
	preview.WriteString(r.colors.HeaderColor("Auto Scaling Group Details"))
	preview.WriteString("\n")
	preview.WriteString(strings.Repeat("─", min(width, 60)))
	preview.WriteString("\n\n")

	// Basic information
	preview.WriteString(r.colors.BoldColor("Basic Information:"))
	preview.WriteString("\n")
	preview.WriteString(fmt.Sprintf("  Name:              %s\n", asg.Name))
	preview.WriteString(fmt.Sprintf("  Health Check:      %s\n", asg.HealthCheckType))

	if !asg.CreatedTime.IsZero() {
		age := time.Since(asg.CreatedTime)
		preview.WriteString(fmt.Sprintf("  Created:           %s (%s ago)\n",
			asg.CreatedTime.Format("2006-01-02 15:04:05"),
			r.formatDuration(age)))
	}

	preview.WriteString("\n")

	// Scaling configuration
	preview.WriteString(r.colors.BoldColor("Scaling Configuration:"))
	preview.WriteString("\n")
	preview.WriteString(fmt.Sprintf("  Desired Capacity:  %d\n", asg.DesiredCapacity))
	preview.WriteString(fmt.Sprintf("  Min Size:          %d\n", asg.MinSize))
	preview.WriteString(fmt.Sprintf("  Max Size:          %d\n", asg.MaxSize))
	preview.WriteString(fmt.Sprintf("  Current Size:      %d\n", asg.CurrentSize))

	// Show scaling status
	switch {
	case asg.CurrentSize < asg.DesiredCapacity:
		preview.WriteString(fmt.Sprintf("  %s Scaling up (%d → %d)\n",
			r.colors.WarningColor("⚠"),
			asg.CurrentSize,
			asg.DesiredCapacity))
	case asg.CurrentSize > asg.DesiredCapacity:
		preview.WriteString(fmt.Sprintf("  %s Scaling down (%d → %d)\n",
			r.colors.WarningColor("⚠"),
			asg.CurrentSize,
			asg.DesiredCapacity))
	case asg.CurrentSize == asg.DesiredCapacity && asg.DesiredCapacity > 0:
		preview.WriteString(fmt.Sprintf("  %s At desired capacity\n",
			r.colors.SuccessColor("✓")))
	case asg.DesiredCapacity == 0:
		preview.WriteString(fmt.Sprintf("  %s Scaled to zero\n",
			r.colors.WarningColor("○")))
	}

	preview.WriteString("\n")

	// Launch configuration
	preview.WriteString(r.colors.BoldColor("Launch Configuration:"))
	preview.WriteString("\n")
	switch {
	case asg.LaunchTemplateName != "":
		preview.WriteString(fmt.Sprintf("  Launch Template:   %s\n", asg.LaunchTemplateName))
	case asg.LaunchConfigurationName != "":
		preview.WriteString(fmt.Sprintf("  Launch Config:     %s\n", asg.LaunchConfigurationName))
	default:
		preview.WriteString("  Launch Config:     Not specified\n")
	}

	// Availability zones
	if len(asg.AvailabilityZones) > 0 {
		preview.WriteString("\n")
		preview.WriteString(r.colors.BoldColor("Availability Zones:"))
		preview.WriteString("\n")
		for _, az := range asg.AvailabilityZones {
			preview.WriteString(fmt.Sprintf("  • %s\n", az))
		}
	}

	// Tags
	if len(asg.Tags) > 0 {
		preview.WriteString("\n")
		preview.WriteString(r.colors.BoldColor("Tags:"))
		preview.WriteString("\n")
		for key, value := range asg.Tags {
			preview.WriteString(fmt.Sprintf("  %s\n", r.colors.TagColor(key, value)))
		}
	}

	return preview.String()
}

// formatDuration formats a duration in a human-readable way
func (r *ASGPreviewRenderer) formatDuration(d time.Duration) string {
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
