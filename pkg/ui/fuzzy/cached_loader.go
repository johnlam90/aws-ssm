package fuzzy

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/johnlam90/aws-ssm/pkg/cache"
)

// CachedInstanceLoader wraps an InstanceLoader with caching support
type CachedInstanceLoader struct {
	loader       InstanceLoader
	cacheService *cache.Service
	region       string
	enabled      bool
}

// NewCachedInstanceLoader creates a new cached instance loader
func NewCachedInstanceLoader(loader InstanceLoader, cacheService *cache.Service, region string, enabled bool) *CachedInstanceLoader {
	return &CachedInstanceLoader{
		loader:       loader,
		cacheService: cacheService,
		region:       region,
		enabled:      enabled,
	}
}

// LoadInstances loads instances with caching support
func (c *CachedInstanceLoader) LoadInstances(ctx context.Context, query *SearchQuery) ([]Instance, error) {
	if !c.enabled || c.cacheService == nil {
		// Cache disabled, load directly
		return c.loader.LoadInstances(ctx, query)
	}

	// Generate cache key from query
	cacheKey := c.generateCacheKey(query)

	// Try to get from cache
	if cached, ok := c.cacheService.Get(cacheKey); ok {
		if instances, ok := cached.([]Instance); ok {
			return instances, nil
		}
		// If cache entry is corrupted, continue to load fresh data
	}

	// Load from source
	instances, err := c.loader.LoadInstances(ctx, query)
	if err != nil {
		return nil, err
	}

	// Store in cache
	if err := c.cacheService.Set(cacheKey, instances, c.region, c.queryToString(query)); err != nil {
		// Log error but don't fail - caching is optional
		fmt.Printf("Warning: failed to cache instances: %v\n", err)
	}

	return instances, nil
}

// LoadInstance loads a single instance with caching support
func (c *CachedInstanceLoader) LoadInstance(ctx context.Context, instanceID string) (*Instance, error) {
	if !c.enabled || c.cacheService == nil {
		// Cache disabled, load directly
		return c.loader.LoadInstance(ctx, instanceID)
	}

	// Generate cache key for single instance
	cacheKey := fmt.Sprintf("instance_%s_%s", c.region, instanceID)

	// Try to get from cache
	if cached, ok := c.cacheService.Get(cacheKey); ok {
		if instance, ok := cached.(*Instance); ok {
			return instance, nil
		}
		// If cache entry is corrupted, continue to load fresh data
	}

	// Load from source
	instance, err := c.loader.LoadInstance(ctx, instanceID)
	if err != nil {
		return nil, err
	}

	// Store in cache
	if err := c.cacheService.Set(cacheKey, instance, c.region, instanceID); err != nil {
		// Log error but don't fail - caching is optional
		fmt.Printf("Warning: failed to cache instance: %v\n", err)
	}

	return instance, nil
}

// generateCacheKey generates a cache key from a search query
func (c *CachedInstanceLoader) generateCacheKey(query *SearchQuery) string {
	// Create a deterministic key from the query
	queryStr := c.queryToString(query)
	hash := md5.Sum([]byte(queryStr))
	return fmt.Sprintf("instances_%s_%s", c.region, hex.EncodeToString(hash[:]))
}

// queryToString converts a search query to a string for cache key generation
func (c *CachedInstanceLoader) queryToString(query *SearchQuery) string {
	if query == nil {
		return "all"
	}

	// Create a JSON representation for consistent hashing
	data, err := json.Marshal(query)
	if err != nil {
		return "all"
	}
	return string(data)
}

// GetRegions returns available regions from the underlying loader
func (c *CachedInstanceLoader) GetRegions() []string {
	if loader, ok := c.loader.(interface{ GetRegions() []string }); ok {
		return loader.GetRegions()
	}
	return []string{}
}

// GetCurrentRegion returns the current region
func (c *CachedInstanceLoader) GetCurrentRegion() string {
	return c.region
}
