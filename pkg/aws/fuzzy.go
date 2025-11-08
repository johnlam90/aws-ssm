package aws

import (
	"context"
	"fmt"

	"github.com/aws-ssm/pkg/ui/fuzzy"
)

// SelectInstanceInteractive displays an enhanced interactive fuzzy finder to select an EC2 instance
func (c *Client) SelectInstanceInteractive(ctx context.Context) (*Instance, error) {
	// Use cached configuration for performance (loaded once in NewClient)
	cfg := c.AppConfig

	// Create fuzzy config
	fuzzyConfig := fuzzy.Config{
		Columns:      fuzzy.DefaultColumnConfig(),
		Weights:      fuzzy.DefaultWeightConfig(),
		Cache:        fuzzy.CacheConfig{Enabled: cfg.Cache.Enabled, TTLMinutes: cfg.Cache.TTLMinutes, CacheDir: cfg.Cache.CacheDir},
		MaxInstances: cfg.Interactive.MaxInstances,
		NoColor:      cfg.Interactive.NoColor,
		Width:        cfg.Interactive.Width,
		Favorites:    false,
		ConfigPath:   "", // Using default
	}

	// Create instance loader
	loader := fuzzy.NewAWSInstanceLoader(c)

	// Create enhanced fuzzy finder
	finder := fuzzy.NewEnhancedFinder(loader, fuzzyConfig)

	// Select instances
	selectedInstances, err := finder.SelectInstanceInteractive(ctx)
	if err != nil {
		return nil, err
	}

	if len(selectedInstances) == 0 {
		return nil, nil // No selection made
	}

	// Convert to AWS instance format
	awsInstance := &Instance{
		InstanceID:       selectedInstances[0].InstanceID,
		Name:             selectedInstances[0].Name,
		State:            selectedInstances[0].State,
		PrivateIP:        selectedInstances[0].PrivateIP,
		PublicIP:         selectedInstances[0].PublicIP,
		PrivateDNS:       selectedInstances[0].PrivateDNS,
		PublicDNS:        selectedInstances[0].PublicDNS,
		InstanceType:     selectedInstances[0].InstanceType,
		AvailabilityZone: selectedInstances[0].AvailabilityZone,
		Tags:             selectedInstances[0].Tags,
	}

	return awsInstance, nil
}

// SelectInstancesInteractive displays an enhanced interactive fuzzy finder to select multiple EC2 instances
func (c *Client) SelectInstancesInteractive(ctx context.Context) ([]Instance, error) {
	// Use cached configuration for performance (loaded once in NewClient)
	cfg := c.AppConfig

	// Create fuzzy config
	fuzzyConfig := fuzzy.Config{
		Columns:      fuzzy.DefaultColumnConfig(),
		Weights:      fuzzy.DefaultWeightConfig(),
		Cache:        fuzzy.CacheConfig{Enabled: cfg.Cache.Enabled, TTLMinutes: cfg.Cache.TTLMinutes, CacheDir: cfg.Cache.CacheDir},
		MaxInstances: cfg.Interactive.MaxInstances,
		NoColor:      cfg.Interactive.NoColor,
		Width:        cfg.Interactive.Width,
		Favorites:    false,
		ConfigPath:   "", // Using default
	}

	// Create instance loader
	loader := fuzzy.NewAWSInstanceLoader(c)

	// Create enhanced fuzzy finder
	finder := fuzzy.NewEnhancedFinder(loader, fuzzyConfig)

	// Select instances
	selectedInstances, err := finder.SelectInstanceInteractive(ctx)
	if err != nil {
		return nil, err
	}

	// Convert to AWS instance format
	var awsInstances []Instance
	for _, inst := range selectedInstances {
		awsInstance := Instance{
			InstanceID:       inst.InstanceID,
			Name:             inst.Name,
			State:            inst.State,
			PrivateIP:        inst.PrivateIP,
			PublicIP:         inst.PublicIP,
			PrivateDNS:       inst.PrivateDNS,
			PublicDNS:        inst.PublicDNS,
			InstanceType:     inst.InstanceType,
			AvailabilityZone: inst.AvailabilityZone,
			Tags:             inst.Tags,
		}
		awsInstances = append(awsInstances, awsInstance)
	}

	return awsInstances, nil
}

// SelectInstanceFromProvided displays an interactive fuzzy finder for a provided instance slice.
// It does not refetch instances and assumes the slice is non-empty.
func (c *Client) SelectInstanceFromProvided(ctx context.Context, instances []Instance) (*Instance, error) {
	if len(instances) == 0 {
		return nil, fmt.Errorf("no instances provided for interactive selection")
	}

	// For now, keep the original implementation for this method
	// TODO: Convert to use enhanced fuzzy finder

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

	// Create fuzzy config
	fuzzyConfig := fuzzy.DefaultConfig()

	// Convert AWS instances to fuzzy instances
	var fuzzyInstances []fuzzy.Instance
	for _, inst := range instances {
		fuzzyInst := fuzzy.Instance{
			InstanceID:       inst.InstanceID,
			Name:             inst.Name,
			State:            inst.State,
			PrivateIP:        inst.PrivateIP,
			PublicIP:         inst.PublicIP,
			PrivateDNS:       inst.PrivateDNS,
			PublicDNS:        inst.PublicDNS,
			InstanceType:     inst.InstanceType,
			AvailabilityZone: inst.AvailabilityZone,
			Tags:             inst.Tags,
		}
		fuzzyInstances = append(fuzzyInstances, fuzzyInst)
	}

	// Create loader for provided instances
	loader := fuzzy.NewProvidedInstanceLoader(fuzzyInstances)

	// Create enhanced fuzzy finder
	finder := fuzzy.NewEnhancedFinder(loader, fuzzyConfig)

	// Select instance
	selectedInstances, err := finder.SelectInstanceInteractive(ctx)
	if err != nil {
		return nil, err
	}

	if len(selectedInstances) == 0 {
		return nil, nil // No selection made
	}

	// Convert back to AWS instance format
	for i, inst := range instances {
		if inst.InstanceID == selectedInstances[0].InstanceID {
			return &instances[i], nil
		}
	}

	return nil, fmt.Errorf("selected instance not found in provided list")
}
