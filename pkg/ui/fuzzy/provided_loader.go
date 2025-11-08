package fuzzy

import (
	"context"
	"fmt"
)

// ProvidedInstanceLoader implements InstanceLoader for a pre-provided slice of instances
type ProvidedInstanceLoader struct {
	instances []Instance
	region    string
}

// NewProvidedInstanceLoader creates a new loader for provided instances
func NewProvidedInstanceLoader(instances []Instance) *ProvidedInstanceLoader {
	return &ProvidedInstanceLoader{
		instances: instances,
		region:    "unknown", // Could be enhanced to extract from instances
	}
}

// LoadInstances returns the provided instances directly (no channel overhead)
func (l *ProvidedInstanceLoader) LoadInstances(ctx context.Context, query *SearchQuery) ([]Instance, error) {
	// Pre-allocate with capacity for better performance
	filtered := make([]Instance, 0, len(l.instances))

	// Filter instances based on query
	for _, inst := range l.instances {
		if inst.MatchesQuery(query) {
			filtered = append(filtered, inst)
		}
	}

	return filtered, nil
}

// LoadInstance finds a specific instance by ID
func (l *ProvidedInstanceLoader) LoadInstance(ctx context.Context, instanceID string) (*Instance, error) {
	for _, inst := range l.instances {
		if inst.InstanceID == instanceID {
			return &inst, nil
		}
	}
	return nil, fmt.Errorf("instance not found: %s", instanceID)
}

// GetRegions returns available regions (just the one we know)
func (l *ProvidedInstanceLoader) GetRegions() []string {
	return []string{l.region}
}

// GetCurrentRegion returns the current region
func (l *ProvidedInstanceLoader) GetCurrentRegion() string {
	return l.region
}
