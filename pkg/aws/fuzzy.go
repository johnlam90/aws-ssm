package aws

import (
	"context"
	"fmt"

	"github.com/aws-ssm/pkg/cache"
	"github.com/aws-ssm/pkg/ui/fuzzy"
)

// columnNamesToConfig converts CLI column names to ColumnConfig
func columnNamesToConfig(columnNames []string) fuzzy.ColumnConfig {
	config := fuzzy.ColumnConfig{}
	for _, col := range columnNames {
		switch col {
		case "name":
			config.Name = true
		case "instance-id":
			config.InstanceID = true
		case "private-ip":
			config.PrivateIP = true
		case "state":
			config.State = true
		case "type":
			config.Type = true
		case "az":
			config.AZ = true
		}
	}
	return config
}

// SelectInstanceInteractive displays an enhanced interactive fuzzy finder to select an EC2 instance
func (c *Client) SelectInstanceInteractive(ctx context.Context) (*Instance, error) {
	// Use cached configuration for performance (loaded once in NewClient)
	cfg := c.AppConfig

	// Determine columns to display - use CLI flags if provided, otherwise use config
	columns := fuzzy.DefaultColumnConfig()
	if len(c.InteractiveCols) > 0 {
		columns = columnNamesToConfig(c.InteractiveCols)
	}

	// Determine NoColor setting - CLI flag takes precedence
	noColor := cfg.Interactive.NoColor
	if c.NoColor {
		noColor = true
	}

	// Determine Width setting - CLI flag takes precedence if non-zero
	width := cfg.Interactive.Width
	if c.Width > 0 {
		width = c.Width
	}

	// Create fuzzy config
	fuzzyConfig := fuzzy.Config{
		Columns:      columns,
		Weights:      fuzzy.DefaultWeightConfig(),
		Cache:        fuzzy.CacheConfig{Enabled: cfg.Cache.Enabled, TTLMinutes: cfg.Cache.TTLMinutes, CacheDir: cfg.Cache.CacheDir},
		MaxInstances: cfg.Interactive.MaxInstances,
		NoColor:      noColor,
		Width:        width,
		Favorites:    c.Favorites,
		ConfigPath:   "", // Using default
	}

	// Create instance loader
	baseLoader := fuzzy.NewAWSInstanceLoader(c)

	// Wrap with cache if enabled
	var loader fuzzy.InstanceLoader = baseLoader
	if fuzzyConfig.Cache.Enabled {
		cacheService, err := cache.NewCacheService(fuzzyConfig.Cache.CacheDir, fuzzyConfig.Cache.TTLMinutes)
		if err != nil {
			// Log warning but continue without cache
			fmt.Printf("Warning: failed to initialize cache: %v\n", err)
		} else {
			loader = fuzzy.NewCachedInstanceLoader(baseLoader, cacheService, c.Config.Region, true)
		}
	}

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

	// Determine columns to display - use CLI flags if provided, otherwise use config
	columns := fuzzy.DefaultColumnConfig()
	if len(c.InteractiveCols) > 0 {
		columns = columnNamesToConfig(c.InteractiveCols)
	}

	// Determine NoColor setting - CLI flag takes precedence
	noColor := cfg.Interactive.NoColor
	if c.NoColor {
		noColor = true
	}

	// Determine Width setting - CLI flag takes precedence if non-zero
	width := cfg.Interactive.Width
	if c.Width > 0 {
		width = c.Width
	}

	// Create fuzzy config
	fuzzyConfig := fuzzy.Config{
		Columns:      columns,
		Weights:      fuzzy.DefaultWeightConfig(),
		Cache:        fuzzy.CacheConfig{Enabled: cfg.Cache.Enabled, TTLMinutes: cfg.Cache.TTLMinutes, CacheDir: cfg.Cache.CacheDir},
		MaxInstances: cfg.Interactive.MaxInstances,
		NoColor:      noColor,
		Width:        width,
		Favorites:    c.Favorites,
		ConfigPath:   "", // Using default
	}

	// Create instance loader
	baseLoader := fuzzy.NewAWSInstanceLoader(c)

	// Wrap with cache if enabled
	var loader fuzzy.InstanceLoader = baseLoader
	if fuzzyConfig.Cache.Enabled {
		cacheService, err := cache.NewCacheService(fuzzyConfig.Cache.CacheDir, fuzzyConfig.Cache.TTLMinutes)
		if err != nil {
			// Log warning but continue without cache
			fmt.Printf("Warning: failed to initialize cache: %v\n", err)
		} else {
			loader = fuzzy.NewCachedInstanceLoader(baseLoader, cacheService, c.Config.Region, true)
		}
	}

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

// EKSDescriber wraps the Client to provide the ClusterDescriber interface
type EKSDescriber struct {
	client *Client
}

// DescribeCluster implements the ClusterDescriber interface
func (e *EKSDescriber) DescribeCluster(ctx context.Context, clusterName string) (any, error) {
	return e.client.DescribeCluster(ctx, clusterName)
}

// SelectEKSClusterInteractive displays an interactive fuzzy finder to select an EKS cluster
func (c *Client) SelectEKSClusterInteractive(ctx context.Context) (*Cluster, error) {
	// Create EKS loader with both list and describe capabilities
	describer := &EKSDescriber{client: c}
	loader := fuzzy.NewAWSEKSLoader(c, describer)

	// Create EKS finder
	finder := fuzzy.NewEKSFinder(loader, fuzzy.DefaultConfig())

	// Select cluster
	selectedCluster, err := finder.SelectClusterInteractive(ctx)
	if err != nil {
		return nil, err
	}

	if selectedCluster == nil {
		return nil, nil // No selection made
	}

	// Fetch full cluster details
	cluster, err := c.DescribeCluster(ctx, selectedCluster.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to describe selected cluster: %w", err)
	}

	return cluster, nil
}
