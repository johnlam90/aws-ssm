package fuzzy

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// DefaultPreviewRenderer implements PreviewRenderer interface
type DefaultPreviewRenderer struct {
	colors ColorManager
}

// NewDefaultPreviewRenderer creates a new preview renderer
func NewDefaultPreviewRenderer(colors ColorManager) *DefaultPreviewRenderer {
	return &DefaultPreviewRenderer{colors: colors}
}

// Render renders the preview for an instance
func (r *DefaultPreviewRenderer) Render(instance *Instance, width, _ int) string {
	if instance == nil {
		return ""
	}

	var preview strings.Builder

	// Header
	preview.WriteString(r.colors.HeaderColor("Instance Details"))
	preview.WriteString("\n")
	preview.WriteString(strings.Repeat("─", minimum(width, 50)))
	preview.WriteString("\n\n")

	// Basic information
	preview.WriteString(r.colors.BoldColor("Basic Info:"))
	preview.WriteString("\n")
	fmt.Fprintf(&preview, "  Name:           %s\n", instance.Name)
	if instance.Name == "" {
		preview.WriteString("  (no name tag)\n")
	}
	fmt.Fprintf(&preview, "  Instance ID:    %s\n", instance.InstanceID)
	fmt.Fprintf(&preview, "  State:          %s\n", r.colors.StateColor(instance.State))
	fmt.Fprintf(&preview, "  Instance Type:  %s\n", instance.InstanceType)
	fmt.Fprintf(&preview, "  Availability:   %s\n", instance.AvailabilityZone)

	// Launch time and uptime
	if !instance.LaunchTime.IsZero() {
		uptime := time.Since(instance.LaunchTime)
		fmt.Fprintf(&preview, "  Launch Time:    %s\n", instance.LaunchTime.Format("2006-01-02 15:04:05"))
		fmt.Fprintf(&preview, "  Uptime:         %s\n", r.formatUptime(uptime))
	}

	preview.WriteString("\n")

	// Network information
	preview.WriteString(r.colors.BoldColor("Network:"))
	preview.WriteString("\n")
	if instance.PrivateIP != "" {
		fmt.Fprintf(&preview, "  Private IP:     %s\n", instance.PrivateIP)
	}
	if instance.PublicIP != "" {
		fmt.Fprintf(&preview, "  Public IP:      %s\n", instance.PublicIP)
	}
	if instance.PrivateDNS != "" {
		fmt.Fprintf(&preview, "  Private DNS:    %s\n", instance.PrivateDNS)
	}
	if instance.PublicDNS != "" {
		fmt.Fprintf(&preview, "  Public DNS:     %s\n", instance.PublicDNS)
	}

	preview.WriteString("\n")

	// Security information
	preview.WriteString(r.colors.BoldColor("Security:"))
	preview.WriteString("\n")
	if instance.InstanceProfile != "" {
		fmt.Fprintf(&preview, "  Instance Profile: %s\n", instance.InstanceProfile)
	}
	if len(instance.SecurityGroups) > 0 {
		preview.WriteString("  Security Groups:\n")
		for _, sg := range instance.SecurityGroups {
			fmt.Fprintf(&preview, "    • %s\n", sg)
		}
	}

	// Tags
	if len(instance.Tags) > 0 {
		preview.WriteString("\n")
		preview.WriteString(r.colors.BoldColor("Tags:"))
		preview.WriteString("\n")
		for key, value := range instance.Tags {
			if key != "Name" { // Name already shown in basic info
				fmt.Fprintf(&preview, "  %s\n", r.colors.TagColor(key, value))
			}
		}
	}

	return preview.String()
}

// RenderJSON renders the instance as JSON
func (r *DefaultPreviewRenderer) RenderJSON(instance *Instance) string {
	if instance == nil {
		return ""
	}

	jsonData, err := json.MarshalIndent(instance, "", "  ")
	if err != nil {
		return fmt.Sprintf("Error rendering JSON: %v", err)
	}

	return string(jsonData)
}

// formatUptime formats uptime duration in a human-readable way
func (r *DefaultPreviewRenderer) formatUptime(uptime time.Duration) string {
	days := int(uptime.Hours()) / 24
	hours := int(uptime.Hours()) % 24
	minutes := int(uptime.Minutes()) % 60

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

// min returns the minimum of two integers
func minimum(a, b int) int {
	if a < b {
		return a
	}
	return b
}
