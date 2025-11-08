package cache

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Entry represents a cached item
type Entry struct {
	Data      interface{} `json:"data"`
	Timestamp time.Time   `json:"timestamp"`
	Region    string      `json:"region"`
	Query     string      `json:"query"`
}

// Service handles caching of instance data
type Service struct {
	cacheDir string
	ttl      time.Duration
}

// NewCacheService creates a new cache service
func NewCacheService(cacheDir string, ttlMinutes int) (*Service, error) {
	if cacheDir == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get user home directory: %w", err)
		}
		cacheDir = filepath.Join(homeDir, ".aws-ssm", "cache")
	}

	// Ensure cache directory exists with restricted permissions
	if err := os.MkdirAll(cacheDir, 0700); err != nil {
		return nil, fmt.Errorf("failed to create cache directory: %w", err)
	}

	return &Service{
		cacheDir: cacheDir,
		ttl:      time.Duration(ttlMinutes) * time.Minute,
	}, nil
}

// Get retrieves cached data for the given key
func (c *Service) Get(key string) (interface{}, bool) {
	cacheFile := filepath.Join(c.cacheDir, key+".json")
	// Validate path to prevent directory traversal
	cleanPath := filepath.Clean(cacheFile)
	if !strings.HasPrefix(cleanPath, filepath.Clean(c.cacheDir)) {
		return nil, false
	}

	// Check file size before reading to prevent reading excessively large files
	const maxCacheFileSize = 10 * 1024 * 1024 // 10 MB limit
	fileInfo, err := os.Stat(cleanPath)
	if err != nil {
		if !os.IsNotExist(err) {
			// Log error but don't fail
			fmt.Fprintf(os.Stderr, "Warning: failed to stat cache file %s: %v\n", cacheFile, err)
		}
		return nil, false
	}

	if fileInfo.Size() > maxCacheFileSize {
		fmt.Fprintf(os.Stderr, "Warning: cache file %s exceeds size limit (%d > %d bytes)\n", cacheFile, fileInfo.Size(), maxCacheFileSize)
		return nil, false
	}

	data, err := os.ReadFile(cleanPath)
	if err != nil {
		if !os.IsNotExist(err) {
			// Log error but don't fail
			fmt.Fprintf(os.Stderr, "Warning: failed to read cache file %s: %v\n", cacheFile, err)
		}
		return nil, false
	}

	var entry Entry
	if err := json.Unmarshal(data, &entry); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to unmarshal cache entry %s: %v\n", cacheFile, err)
		return nil, false
	}

	// Check if cache entry is expired
	if time.Since(entry.Timestamp) > c.ttl {
		// Remove expired cache file (ignore error as it's cleanup)
		//nolint:errcheck // Cleanup operation, error is not critical
		_ = os.Remove(cacheFile)
		return nil, false
	}

	return entry.Data, true
}

// Set stores data in cache with the given key
func (c *Service) Set(key string, data interface{}, region, query string) error {
	cacheFile := filepath.Join(c.cacheDir, key+".json")

	entry := Entry{
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
	if err := os.WriteFile(tempFile, jsonData, 0600); err != nil {
		return fmt.Errorf("failed to write cache file: %w", err)
	}

	return os.Rename(tempFile, cacheFile)
}

// Delete removes cached data for the given key
func (c *Service) Delete(key string) error {
	cacheFile := filepath.Join(c.cacheDir, key+".json")
	return os.Remove(cacheFile)
}

// Clear removes all cache files
func (c *Service) Clear() error {
	files, err := os.ReadDir(c.cacheDir)
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
func (c *Service) Cleanup() error {
	files, err := os.ReadDir(c.cacheDir)
	if err != nil {
		return fmt.Errorf("failed to read cache directory: %w", err)
	}

	for _, file := range files {
		if filepath.Ext(file.Name()) != ".json" {
			continue
		}

		cacheFile := filepath.Join(c.cacheDir, file.Name())
		// Validate path to prevent directory traversal
		cleanPath := filepath.Clean(cacheFile)
		if !strings.HasPrefix(cleanPath, filepath.Clean(c.cacheDir)) {
			fmt.Fprintf(os.Stderr, "Warning: invalid cache file path %s\n", file.Name())
			continue
		}
		data, err := os.ReadFile(cleanPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to read cache file %s: %v\n", file.Name(), err)
			continue
		}

		var entry Entry
		if err := json.Unmarshal(data, &entry); err != nil {
			// Invalid cache file, remove it (ignore error as it's cleanup)
			//nolint:errcheck // Cleanup operation, error is not critical
			_ = os.Remove(cacheFile)
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
func (c *Service) GetCacheStats() (totalFiles, expiredFiles int, totalSize int64, err error) {
	files, err := os.ReadDir(c.cacheDir)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("failed to read cache directory: %w", err)
	}

	for _, file := range files {
		if filepath.Ext(file.Name()) != ".json" {
			continue
		}

		totalFiles++
		info, err := file.Info()
		if err != nil {
			continue
		}
		totalSize += info.Size()

		cacheFile := filepath.Join(c.cacheDir, file.Name())
		// Validate path to prevent directory traversal
		cleanPath := filepath.Clean(cacheFile)
		if !strings.HasPrefix(cleanPath, filepath.Clean(c.cacheDir)) {
			continue
		}
		data, err := os.ReadFile(cleanPath)
		if err != nil {
			continue
		}

		var entry Entry
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
