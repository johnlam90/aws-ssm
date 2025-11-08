package cache

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"
)

// CacheEntry represents a cached item
type CacheEntry struct {
	Data      interface{} `json:"data"`
	Timestamp time.Time   `json:"timestamp"`
	Region    string      `json:"region"`
	Query     string      `json:"query"`
}

// CacheService handles caching of instance data
type CacheService struct {
	cacheDir string
	ttl      time.Duration
}

// NewCacheService creates a new cache service
func NewCacheService(cacheDir string, ttlMinutes int) (*CacheService, error) {
	if cacheDir == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get user home directory: %w", err)
		}
		cacheDir = filepath.Join(homeDir, ".aws-ssm", "cache")
	}

	// Ensure cache directory exists
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create cache directory: %w", err)
	}

	return &CacheService{
		cacheDir: cacheDir,
		ttl:      time.Duration(ttlMinutes) * time.Minute,
	}, nil
}

// Get retrieves cached data for the given key
func (c *CacheService) Get(key string) (interface{}, bool) {
	cacheFile := filepath.Join(c.cacheDir, key+".json")
	
	data, err := ioutil.ReadFile(cacheFile)
	if err != nil {
		if !os.IsNotExist(err) {
			// Log error but don't fail
			fmt.Fprintf(os.Stderr, "Warning: failed to read cache file %s: %v\n", cacheFile, err)
		}
		return nil, false
	}

	var entry CacheEntry
	if err := json.Unmarshal(data, &entry); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to unmarshal cache entry %s: %v\n", cacheFile, err)
		return nil, false
	}

	// Check if cache entry is expired
	if time.Since(entry.Timestamp) > c.ttl {
		// Remove expired cache file
		os.Remove(cacheFile)
		return nil, false
	}

	return entry.Data, true
}

// Set stores data in cache with the given key
func (c *CacheService) Set(key string, data interface{}, region, query string) error {
	cacheFile := filepath.Join(c.cacheDir, key+".json")
	
	entry := CacheEntry{
		Data:      data,
		Timestamp: time.Now(),
		Region:    region,
		Query:     query,
	}

	jsonData, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("failed to marshal cache entry: %w", err)
	}

	// Write to temporary file first, then rename to avoid corruption
	tempFile := cacheFile + ".tmp"
	if err := ioutil.WriteFile(tempFile, jsonData, 0644); err != nil {
		return fmt.Errorf("failed to write cache file: %w", err)
	}

	return os.Rename(tempFile, cacheFile)
}

// Delete removes cached data for the given key
func (c *CacheService) Delete(key string) error {
	cacheFile := filepath.Join(c.cacheDir, key+".json")
	return os.Remove(cacheFile)
}

// Clear removes all cache files
func (c *CacheService) Clear() error {
	files, err := ioutil.ReadDir(c.cacheDir)
	if err != nil {
		return fmt.Errorf("failed to read cache directory: %w", err)
	}

	for _, file := range files {
		if filepath.Ext(file.Name()) == ".json" {
			if err := os.Remove(filepath.Join(c.cacheDir, file.Name())); err != nil {
				fmt.Fprintf(os.Stderr, "Warning: failed to remove cache file %s: %v\n", file.Name(), err)
			}
		}
	}

	return nil
}

// Cleanup removes expired cache files
func (c *CacheService) Cleanup() error {
	files, err := ioutil.ReadDir(c.cacheDir)
	if err != nil {
		return fmt.Errorf("failed to read cache directory: %w", err)
	}

	for _, file := range files {
		if filepath.Ext(file.Name()) != ".json" {
			continue
		}

		cacheFile := filepath.Join(c.cacheDir, file.Name())
		data, err := ioutil.ReadFile(cacheFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to read cache file %s: %v\n", file.Name(), err)
			continue
		}

		var entry CacheEntry
		if err := json.Unmarshal(data, &entry); err != nil {
			// Invalid cache file, remove it
			os.Remove(cacheFile)
			continue
		}

		if time.Since(entry.Timestamp) > c.ttl {
			// Remove expired cache file
			if err := os.Remove(cacheFile); err != nil {
				fmt.Fprintf(os.Stderr, "Warning: failed to remove expired cache file %s: %v\n", file.Name(), err)
			}
		}
	}

	return nil
}

// GetCacheStats returns cache statistics
func (c *CacheService) GetCacheStats() (totalFiles, expiredFiles int, totalSize int64, err error) {
	files, err := ioutil.ReadDir(c.cacheDir)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("failed to read cache directory: %w", err)
	}

	for _, file := range files {
		if filepath.Ext(file.Name()) != ".json" {
			continue
		}

		totalFiles++
		totalSize += file.Size()

		cacheFile := filepath.Join(c.cacheDir, file.Name())
		data, err := ioutil.ReadFile(cacheFile)
		if err != nil {
			continue
		}

		var entry CacheEntry
		if err := json.Unmarshal(data, &entry); err != nil {
			continue
		}

		if time.Since(entry.Timestamp) > c.ttl {
			expiredFiles++
		}
	}

	return totalFiles, expiredFiles, totalSize, nil
}

// GenerateCacheKey generates a cache key based on region and query
func GenerateCacheKey(region, query string) string {
	return fmt.Sprintf("%s_%s", region, query)
}
