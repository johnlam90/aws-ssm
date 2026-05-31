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
func (r *ASGPreviewRenderer) Render(asg *ASGInfo, width, _ int) string {
	if asg == nil {
		return ""
	}

	var preview strings.Builder

	// Header
	preview.WriteString(r.colors.HeaderColor("Auto Scaling Group Details"))
	preview.WriteString("\n")
	preview.WriteString(strings.Repeat("─", minimum(width, 60)))
	preview.WriteString("\n\n")

	// Basic information
	preview.WriteString(r.colors.BoldColor("Basic Information:"))
	preview.WriteString("\n")
	fmt.Fprintf(&preview, "  Name:              %s\n", asg.Name)
	fmt.Fprintf(&preview, "  Health Check:      %s\n", asg.HealthCheckType)

	if !asg.CreatedTime.IsZero() {
		age := time.Since(asg.CreatedTime)
		fmt.Fprintf(&preview, "  Created:           %s (%s ago)\n",
			asg.CreatedTime.Format("2006-01-02 15:04:05"),
			r.formatDuration(age))
	}

	preview.WriteString("\n")

	// Scaling configuration
	preview.WriteString(r.colors.BoldColor("Scaling Configuration:"))
	preview.WriteString("\n")
	fmt.Fprintf(&preview, "  Desired Capacity:  %d\n", asg.DesiredCapacity)
	fmt.Fprintf(&preview, "  Min Size:          %d\n", asg.MinSize)
	fmt.Fprintf(&preview, "  Max Size:          %d\n", asg.MaxSize)
	fmt.Fprintf(&preview, "  Current Size:      %d\n", asg.CurrentSize)

	// Show scaling status
	switch {
	case asg.CurrentSize < asg.DesiredCapacity:
		fmt.Fprintf(&preview, "  %s Scaling up (%d → %d)\n",
			r.colors.WarningColor("⚠"),
			asg.CurrentSize,
			asg.DesiredCapacity)
	case asg.CurrentSize > asg.DesiredCapacity:
		fmt.Fprintf(&preview, "  %s Scaling down (%d → %d)\n",
			r.colors.WarningColor("⚠"),
			asg.CurrentSize,
			asg.DesiredCapacity)
	case asg.CurrentSize == asg.DesiredCapacity && asg.DesiredCapacity > 0:
		fmt.Fprintf(&preview, "  %s At desired capacity\n",
			r.colors.SuccessColor("✓"))
	case asg.DesiredCapacity == 0:
		fmt.Fprintf(&preview, "  %s Scaled to zero\n",
			r.colors.WarningColor("○"))
	}

	preview.WriteString("\n")

	// Launch configuration
	preview.WriteString(r.colors.BoldColor("Launch Configuration:"))
	preview.WriteString("\n")
	switch {
	case asg.LaunchTemplateName != "":
		fmt.Fprintf(&preview, "  Launch Template:   %s\n", asg.LaunchTemplateName)
	case asg.LaunchConfigurationName != "":
		fmt.Fprintf(&preview, "  Launch Config:     %s\n", asg.LaunchConfigurationName)
	default:
		preview.WriteString("  Launch Config:     Not specified\n")
	}

	// Availability zones
	if len(asg.AvailabilityZones) > 0 {
		preview.WriteString("\n")
		preview.WriteString(r.colors.BoldColor("Availability Zones:"))
		preview.WriteString("\n")
		for _, az := range asg.AvailabilityZones {
			fmt.Fprintf(&preview, "  • %s\n", az)
		}
	}

	// Tags
	if len(asg.Tags) > 0 {
		preview.WriteString("\n")
		preview.WriteString(r.colors.BoldColor("Tags:"))
		preview.WriteString("\n")
		for key, value := range asg.Tags {
			fmt.Fprintf(&preview, "  %s\n", r.colors.TagColor(key, value))
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
