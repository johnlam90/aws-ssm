package aws

import (
	"context"
	"fmt"
	"strings"

	fuzzyfinder "github.com/ktr0731/go-fuzzyfinder"
)

// SelectInstanceInteractive displays an interactive fuzzy finder to select an EC2 instance
func (c *Client) SelectInstanceInteractive(ctx context.Context) (*Instance, error) {
	// Get all running instances (no tag filters)
	instances, err := c.ListInstances(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to list instances: %w", err)
	}

	if len(instances) == 0 {
		return nil, fmt.Errorf("no instances found in region %s", c.Config.Region)
	}

	// Filter to only running instances
	var runningInstances []Instance
	for _, inst := range instances {
		if inst.State == "running" {
			runningInstances = append(runningInstances, inst)
		}
	}

	if len(runningInstances) == 0 {
		return nil, fmt.Errorf("no running instances found in region %s (found %d instances in other states)", c.Config.Region, len(instances))
	}

	instances = runningInstances

	// Use fuzzy finder to select an instance
	idx, err := fuzzyfinder.Find(
		instances,
		func(i int) string {
			// Display format: "Name | Instance ID | Private IP | State"
			name := instances[i].Name
			if name == "" {
				name = "(no name)"
			}
			return fmt.Sprintf("%-30s | %-19s | %-15s | %s",
				truncate(name, 30),
				instances[i].InstanceID,
				instances[i].PrivateIP,
				instances[i].State,
			)
		},
		fuzzyfinder.WithPreviewWindow(func(i, w, h int) string {
			if i == -1 {
				return ""
			}
			return formatInstancePreview(instances[i])
		}),
	)

	if err != nil {
		return nil, err
	}

	return &instances[idx], nil
}

// SelectInstanceFromProvided displays an interactive fuzzy finder for a provided instance slice.
// It does not refetch instances and assumes the slice is non-empty.
func (c *Client) SelectInstanceFromProvided(ctx context.Context, instances []Instance) (*Instance, error) {
	if len(instances) == 0 {
		return nil, fmt.Errorf("no instances provided for interactive selection")
	}

	// Filter running instances first to reduce noise, but if that empties the list, fall back.
	running := make([]Instance, 0, len(instances))
	for _, inst := range instances {
		if inst.State == "running" {
			running = append(running, inst)
		}
	}
	if len(running) > 0 {
		instances = running
	}

	idx, err := fuzzyfinder.Find(
		instances,
		func(i int) string {
			name := instances[i].Name
			if name == "" {
				name = "(no name)"
			}
			return fmt.Sprintf("%-30s | %-19s | %-15s | %s",
				truncate(name, 30),
				instances[i].InstanceID,
				instances[i].PrivateIP,
				instances[i].State,
			)
		},
		fuzzyfinder.WithPreviewWindow(func(i, w, h int) string {
			if i == -1 {
				return ""
			}
			return formatInstancePreview(instances[i])
		}),
	)
	if err != nil {
		return nil, err
	}
	return &instances[idx], nil
}

// formatInstancePreview formats instance details for the preview window
func formatInstancePreview(instance Instance) string {
	var preview strings.Builder

	name := instance.Name
	if name == "" {
		name = "(no name)"
	}

	preview.WriteString("Instance Details\n")
	preview.WriteString("================\n\n")
	preview.WriteString(fmt.Sprintf("Name:          %s\n", name))
	preview.WriteString(fmt.Sprintf("Instance ID:   %s\n", instance.InstanceID))
	preview.WriteString(fmt.Sprintf("State:         %s\n", instance.State))
	preview.WriteString(fmt.Sprintf("Instance Type: %s\n", instance.InstanceType))
	preview.WriteString(fmt.Sprintf("Private IP:    %s\n", instance.PrivateIP))

	if instance.PublicIP != "" {
		preview.WriteString(fmt.Sprintf("Public IP:     %s\n", instance.PublicIP))
	}

	if instance.PrivateDNS != "" {
		preview.WriteString(fmt.Sprintf("Private DNS:   %s\n", instance.PrivateDNS))
	}

	if instance.PublicDNS != "" {
		preview.WriteString(fmt.Sprintf("Public DNS:    %s\n", instance.PublicDNS))
	}

	preview.WriteString(fmt.Sprintf("AZ:            %s\n", instance.AvailabilityZone))

	// Display tags
	if len(instance.Tags) > 0 {
		preview.WriteString("\nTags:\n")
		for key, value := range instance.Tags {
			if key != "Name" { // Skip Name tag as it's already displayed
				preview.WriteString(fmt.Sprintf("  %s: %s\n", key, value))
			}
		}
	}

	return preview.String()
}

// truncate truncates a string to the specified length
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return s[:maxLen]
	}
	return s[:maxLen-3] + "..."
}
