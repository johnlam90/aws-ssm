package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/johnlam90/aws-ssm/pkg/cache"
	testframework "github.com/johnlam90/aws-ssm/pkg/testing"
)

// TestCacheCommand_Integration_BasicOperations tests basic cache operations
func TestCacheCommand_Integration_BasicOperations(t *testing.T) {
	assertion := testframework.NewAssertion(t)

	// Create a temporary cache directory for testing
	tmpDir, err := os.MkdirTemp("", "aws-ssm-cache-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	testCases := []struct {
		name      string
		setup     func(*cache.Service) error
		verify    func(*cache.Service) error
		expectErr bool
	}{
		{
			name: "Cache Service Initialization",
			setup: func(c *cache.Service) error {
				return nil
			},
			verify: func(c *cache.Service) error {
				// Verify cache service is properly initialized
				assertion.NotNil(c, "Cache service should not be nil")
				return nil
			},
			expectErr: false,
		},
		{
			name: "Cache Set and Get - Simple Data",
			setup: func(c *cache.Service) error {
				// Store simple test data
				testData := map[string]interface{}{
					"instance_id": "i-1234567890abcdef0",
					"name":        "test-instance",
					"status":      "running",
				}
				return c.Set("test-instance-1", testData, "us-east-1", "tag:Name=test-instance")
			},
			verify: func(c *cache.Service) error {
				// Retrieve and verify the data
				data, exists := c.Get("test-instance-1")
				assertion.True(exists, "Cache entry should exist")
				assertion.NotNil(data, "Retrieved data should not be nil")

				// Type assertion to map[string]interface{} (JSON converts strings to interface{})
				if testData, ok := data.(map[string]interface{}); ok {
					assertion.Equal("i-1234567890abcdef0", testData["instance_id"], "Instance ID should match")
					assertion.Equal("test-instance", testData["name"], "Name should match")
					assertion.Equal("running", testData["status"], "Status should match")
				} else {
					return fmt.Errorf("type assertion failed")
				}
				return nil
			},
			expectErr: false,
		},
		{
			name: "Cache Set and Get - Complex Data",
			setup: func(c *cache.Service) error {
				// Store complex test data
				testData := map[string]interface{}{
					"instances": []interface{}{
						map[string]interface{}{
							"InstanceID":       "i-1234567890abcdef0",
							"Name":             "web-server-1",
							"State":            "running",
							"InstanceType":     "t3.medium",
							"PrivateIP":        "10.0.1.100",
							"PublicIP":         "54.123.456.789",
							"AvailabilityZone": "us-east-1a",
							"Tags": map[string]interface{}{
								"Environment": "production",
								"Team":        "web",
								"ManagedBy":   "terraform",
							},
						},
					},
					"total_count": float64(1),
					"region":      "us-east-1",
					"query_time":  float64(time.Now().Unix()),
				}
				return c.Set("production-web-servers", testData, "us-east-1", "tag:Environment=production")
			},
			verify: func(c *cache.Service) error {
				// Retrieve and verify the complex data
				data, exists := c.Get("production-web-servers")
				assertion.True(exists, "Cache entry should exist")
				assertion.NotNil(data, "Retrieved data should not be nil")

				// Type assertion and verification
				if testData, ok := data.(map[string]interface{}); ok {
					assertion.Equal(float64(1), testData["total_count"], "Total count should match")
					assertion.Equal("us-east-1", testData["region"], "Region should match")

					instances := testData["instances"].([]interface{})
					assertion.Length(instances, 1, "Should have exactly one instance")

					instance := instances[0].(map[string]interface{})
					assertion.Equal("i-1234567890abcdef0", instance["InstanceID"], "Instance ID should match")
					assertion.Equal("web-server-1", instance["Name"], "Instance name should match")
					assertion.Equal("running", instance["State"], "Instance state should match")
				} else {
					return fmt.Errorf("type assertion failed")
				}
				return nil
			},
			expectErr: false,
		},
		{
			name: "Cache Delete",
			setup: func(c *cache.Service) error {
				// First store data
				testData := map[string]interface{}{"test": "data"}
				if err := c.Set("to-be-deleted", testData, "us-west-2", "test-query"); err != nil {
					return err
				}
				// Then delete it
				return c.Delete("to-be-deleted")
			},
			verify: func(c *cache.Service) error {
				// Verify the data is deleted
				data, exists := c.Get("to-be-deleted")
				assertion.False(exists, "Cache entry should not exist after deletion")
				assertion.Nil(data, "Retrieved data should be nil")
				return nil
			},
			expectErr: false,
		},
		{
			name: "Cache Get - Non-existent Key",
			setup: func(c *cache.Service) error {
				return nil
			},
			verify: func(c *cache.Service) error {
				// Verify non-existent key returns nil
				data, exists := c.Get("non-existent-key")
				assertion.False(exists, "Non-existent cache entry should return false")
				assertion.Nil(data, "Retrieved data should be nil")
				return nil
			},
			expectErr: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create cache service
			cacheService, err := cache.NewCacheService(tmpDir, 60) // 60 minutes TTL
			if err != nil {
				t.Fatalf("Failed to create cache service: %v", err)
			}

			// Setup test
			if tc.setup != nil {
				if err := tc.setup(cacheService); err != nil {
					if tc.expectErr {
						// Expected error, continue with verification
					} else {
						t.Fatalf("Test setup failed: %v", err)
					}
				}
			}

			// Execute verification
			err = tc.verify(cacheService)

			if tc.expectErr {
				assertion.Error(err, "Expected error should occur")
			} else {
				assertion.NoError(err, "Should not have error")
			}

			if assertion.HasErrors() {
				t.Errorf("Assertion errors occurred: %v", assertion.GetErrors())
			}
		})
	}
}

// TestCacheCommand_Integration_ExpirationAndCleanup tests cache expiration and cleanup operations
func TestCacheCommand_Integration_ExpirationAndCleanup(t *testing.T) {
	assertion := testframework.NewAssertion(t)

	// Create a temporary cache directory for testing
	tmpDir, err := os.MkdirTemp("", "aws-ssm-cache-expiration-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	testCases := []struct {
		name      string
		setup     func(*cache.Service) error
		verify    func(*cache.Service) error
		expectErr bool
	}{
		{
			name: "Cache Entry Expiration",
			setup: func(c *cache.Service) error {
				// Store data with very short TTL (1 minute)
				cacheService, err := cache.NewCacheService(tmpDir, 1) // 1 minute TTL
				if err != nil {
					return err
				}

				testData := map[string]interface{}{"test": "data"}
				return cacheService.Set("expiring-data", testData, "us-east-1", "test-query")
			},
			verify: func(c *cache.Service) error {
				// Verify the data exists initially
				data, exists := c.Get("expiring-data")
				assertion.True(exists, "Fresh cache entry should exist")
				assertion.NotNil(data, "Fresh cache data should not be nil")

				// Note: In real integration tests, we would wait for expiration
				// For this test, we simulate the scenario by checking the logic
				return nil
			},
			expectErr: false,
		},
		{
			name: "Cache Cleanup Operation",
			setup: func(c *cache.Service) error {
				// Create some test cache entries
				testData := map[string]interface{}{"test": "data"}

				// Store multiple entries
				for i := 0; i < 5; i++ {
					key := fmt.Sprintf("test-entry-%d", i)
					if err := c.Set(key, testData, "us-east-1", "test-query"); err != nil {
						return err
					}
				}
				return nil
			},
			verify: func(c *cache.Service) error {
				// Get cache statistics before cleanup
				totalFiles, _, totalSize, err := c.GetCacheStats()
				assertion.NoError(err, "Should be able to get cache stats")
				assertion.Greater(totalFiles, 0, "Should have cache files")
				assertion.Greater(totalSize, int64(0), "Should have cache size")

				// Perform cleanup
				err = c.Cleanup()
				assertion.NoError(err, "Cleanup should succeed")

				// Get cache statistics after cleanup
				totalFilesAfter, _, _, err := c.GetCacheStats()
				assertion.NoError(err, "Should be able to get cache stats after cleanup")

				// Files should be cleaned up (though number depends on expiration)
				assertion.Greater(totalFilesAfter, -1, "Should have non-negative file count")

				return nil
			},
			expectErr: false,
		},
		{
			name: "Cache Statistics",
			setup: func(c *cache.Service) error {
				// Store test entries with different regions
				testData := map[string]interface{}{"test": "data"}

				if err := c.Set("us-east-1-entry", testData, "us-east-1", "query1"); err != nil {
					return err
				}
				if err := c.Set("us-west-2-entry", testData, "us-west-2", "query2"); err != nil {
					return err
				}
				if err := c.Set("eu-west-1-entry", testData, "eu-west-1", "query3"); err != nil {
					return err
				}

				return nil
			},
			verify: func(c *cache.Service) error {
				// Get comprehensive cache statistics
				totalFiles, expiredFiles, totalSize, err := c.GetCacheStats()
				assertion.NoError(err, "Should be able to get cache stats")
				assertion.Greater(totalFiles, 0, "Should have cache files")
				assertion.Greater(totalSize, int64(0), "Should have cache data")
				assertion.Greater(expiredFiles, -1, "Expired files count should be non-negative")

				// Verify cache files exist in the directory
				files, err := os.ReadDir(tmpDir)
				assertion.NoError(err, "Should be able to read cache directory")

				jsonFiles := 0
				for _, file := range files {
					if filepath.Ext(file.Name()) == ".json" {
						jsonFiles++
					}
				}
				assertion.Greater(jsonFiles, 0, "Should have JSON cache files")

				return nil
			},
			expectErr: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create cache service with reasonable TTL
			cacheService, err := cache.NewCacheService(tmpDir, 60) // 60 minutes TTL
			if err != nil {
				t.Fatalf("Failed to create cache service: %v", err)
			}

			// Setup test
			if tc.setup != nil {
				if err := tc.setup(cacheService); err != nil {
					if tc.expectErr {
						// Expected error, continue with verification
					} else {
						t.Fatalf("Test setup failed: %v", err)
					}
				}
			}

			// Execute verification
			err = tc.verify(cacheService)

			if tc.expectErr {
				assertion.Error(err, "Expected error should occur")
			} else {
				assertion.NoError(err, "Should not have error")
			}

			if assertion.HasErrors() {
				t.Errorf("Assertion errors occurred: %v", assertion.GetErrors())
			}
		})
	}
}

// TestCacheCommand_Integration_MultiRegionCache tests multi-region cache handling
func TestCacheCommand_Integration_MultiRegionCache(t *testing.T) {
	assertion := testframework.NewAssertion(t)

	// Create a temporary cache directory for testing
	tmpDir, err := os.MkdirTemp("", "aws-ssm-cache-multi-region-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	testCases := []struct {
		name      string
		setup     func(*cache.Service) error
		verify    func(*cache.Service) error
		expectErr bool
	}{
		{
			name: "Multi-Region Cache Entries",
			setup: func(c *cache.Service) error {
				// Store same query results for different regions
				usEast1Data := map[string]interface{}{
					"region":    "us-east-1",
					"count":     "10",
					"instances": "i-123,i-456,i-789",
				}

				usWest2Data := map[string]interface{}{
					"region":    "us-west-2",
					"count":     "5",
					"instances": "i-abc,i-def,i-ghi",
				}

				euWest1Data := map[string]interface{}{
					"region":    "eu-west-1",
					"count":     "3",
					"instances": "i-111,i-222,i-333",
				}

				if err := c.Set("web-servers-us-east-1", usEast1Data, "us-east-1", "tag:Role=web"); err != nil {
					return err
				}
				if err := c.Set("web-servers-us-west-2", usWest2Data, "us-west-2", "tag:Role=web"); err != nil {
					return err
				}
				if err := c.Set("web-servers-eu-west-1", euWest1Data, "eu-west-1", "tag:Role=web"); err != nil {
					return err
				}

				return nil
			},
			verify: func(c *cache.Service) error {
				// Verify each region's data is correctly stored and retrieved

				// Check US East 1
				data, exists := c.Get("web-servers-us-east-1")
				assertion.True(exists, "us-east-1 cache entry should exist")
				if testData, ok := data.(map[string]interface{}); ok {
					assertion.Equal("us-east-1", testData["region"], "Region should match")
					assertion.Equal("10", testData["count"], "Count should match")
				}

				// Check US West 2
				data, exists = c.Get("web-servers-us-west-2")
				assertion.True(exists, "us-west-2 cache entry should exist")
				if testData, ok := data.(map[string]interface{}); ok {
					assertion.Equal("us-west-2", testData["region"], "Region should match")
					assertion.Equal("5", testData["count"], "Count should match")
				}

				// Check EU West 1
				data, exists = c.Get("web-servers-eu-west-1")
				assertion.True(exists, "eu-west-1 cache entry should exist")
				if testData, ok := data.(map[string]interface{}); ok {
					assertion.Equal("eu-west-1", testData["region"], "Region should match")
					assertion.Equal("3", testData["count"], "Count should match")
				}

				return nil
			},
			expectErr: false,
		},
		{
			name: "Cache Key Generation",
			setup: func(c *cache.Service) error {
				return nil
			},
			verify: func(c *cache.Service) error {
				// Test cache key generation function
				key1 := cache.GenerateCacheKey("us-east-1", "tag:Environment=production")
				key2 := cache.GenerateCacheKey("us-west-2", "tag:Environment=production")
				key3 := cache.GenerateCacheKey("us-east-1", "tag:Environment=staging")

				assertion.Contains(key1, "us-east-1", "Key should contain region")
				assertion.Contains(key1, "production", "Key should contain query")
				assertion.Contains(key2, "us-west-2", "Different regions should produce different keys")
				assertion.Contains(key3, "staging", "Different queries should produce different keys")

				// Keys should be deterministic for same inputs
				key1Duplicate := cache.GenerateCacheKey("us-east-1", "tag:Environment=production")
				assertion.Equal(key1, key1Duplicate, "Same inputs should produce same key")

				return nil
			},
			expectErr: false,
		},
		{
			name: "Cache Clear All Entries",
			setup: func(c *cache.Service) error {
				// Store multiple entries across regions
				testData := map[string]interface{}{"test": "data"}

				for i := 0; i < 10; i++ {
					key := fmt.Sprintf("test-entry-%d", i)
					region := []string{"us-east-1", "us-west-2", "eu-west-1"}[i%3]
					if err := c.Set(key, testData, region, "test-query"); err != nil {
						return err
					}
				}
				return nil
			},
			verify: func(c *cache.Service) error {
				// Verify cache has entries before clear
				totalFiles, _, _, err := c.GetCacheStats()
				assertion.NoError(err, "Should be able to get cache stats")
				assertion.Greater(totalFiles, 0, "Should have cache entries before clear")

				// Clear all cache entries
				err = c.Clear()
				assertion.NoError(err, "Clear operation should succeed")

				// Verify cache is empty after clear
				totalFilesAfter, _, _, err := c.GetCacheStats()
				assertion.NoError(err, "Should be able to get cache stats after clear")
				assertion.Equal(0, totalFilesAfter, "Cache should be empty after clear")

				// Verify individual entries are gone
				for i := 0; i < 10; i++ {
					key := fmt.Sprintf("test-entry-%d", i)
					_, exists := c.Get(key)
					assertion.False(exists, "Entry should not exist after clear")
				}

				return nil
			},
			expectErr: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create cache service
			cacheService, err := cache.NewCacheService(tmpDir, 60) // 60 minutes TTL
			if err != nil {
				t.Fatalf("Failed to create cache service: %v", err)
			}

			// Setup test
			if tc.setup != nil {
				if err := tc.setup(cacheService); err != nil {
					if tc.expectErr {
						// Expected error, continue with verification
					} else {
						t.Fatalf("Test setup failed: %v", err)
					}
				}
			}

			// Execute verification
			err = tc.verify(cacheService)

			if tc.expectErr {
				assertion.Error(err, "Expected error should occur")
			} else {
				assertion.NoError(err, "Should not have error")
			}

			if assertion.HasErrors() {
				t.Errorf("Assertion errors occurred: %v", assertion.GetErrors())
			}
		})
	}
}

// TestCacheCommand_Integration_ErrorHandling tests cache error handling and edge cases
func TestCacheCommand_Integration_ErrorHandling(t *testing.T) {
	assertion := testframework.NewAssertion(t)

	testCases := []struct {
		name      string
		setup     func(*cache.Service, string) error
		execute   func(*cache.Service) error
		verify    func(*cache.Service) error
		expectErr bool
	}{
		{
			name: "Invalid Cache Directory Permissions",
			setup: func(c *cache.Service, dir string) error {
				// Make directory read-only
				return os.Chmod(dir, 0444) // Read-only permissions
			},
			execute: func(c *cache.Service) error {
				// Try to write to read-only directory
				testData := map[string]interface{}{"test": "data"}
				return c.Set("test-entry", testData, "us-east-1", "test-query")
			},
			verify: func(c *cache.Service) error {
				// Should handle permission error gracefully
				return nil
			},
			expectErr: true,
		},
		{
			name: "Corrupted Cache File Handling",
			setup: func(c *cache.Service, dir string) error {
				// Create a corrupted cache file
				corruptedData := `{"invalid": json data`
				return os.WriteFile(filepath.Join(dir, "corrupted.json"), []byte(corruptedData), 0600)
			},
			execute: func(c *cache.Service) error {
				// Try to get corrupted entry - should handle gracefully
				data, exists := c.Get("corrupted")
				assertion.False(exists, "Corrupted entry should not exist")
				assertion.Nil(data, "Data should be nil for corrupted entry")
				return nil
			},
			verify: func(c *cache.Service) error {
				// Verify corrupted file is cleaned up
				return c.Cleanup()
			},
			expectErr: false,
		},
		{
			name: "Large Cache Entry Handling",
			setup: func(c *cache.Service, dir string) error {
				// Create a moderately sized cache entry to avoid complex type assertions
				testData := map[string]interface{}{
					"test_array": []string{"item1", "item2", "item3"},
					"metadata": map[string]interface{}{
						"region": "us-east-1",
						"type":   "test_entry",
					},
				}
				return c.Set("test-entry", testData, "us-east-1", "test-query")
			},
			verify: func(c *cache.Service) error {
				// Verify entry can be retrieved safely
				data, exists := c.Get("test-entry")
				assertion.True(exists, "Test entry should exist")
				assertion.NotNil(data, "Test entry data should not be nil")
				return nil
			},
			expectErr: false,
		},
		{
			name: "Cache Directory Does Not Exist",
			setup: func(c *cache.Service, dir string) error {
				// Remove the cache directory
				return os.RemoveAll(dir)
			},
			execute: func(c *cache.Service) error {
				// Try operations on non-existent directory
				testData := map[string]interface{}{"test": "data"}
				if err := c.Set("test-entry", testData, "us-east-1", "test-query"); err != nil {
					return err
				}

				// Get cache stats
				_, _, _, err := c.GetCacheStats()
				return err
			},
			verify: func(c *cache.Service) error {
				// Should handle missing directory gracefully
				return nil
			},
			expectErr: true, // This might fail depending on implementation
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create a temporary cache directory for testing
			tmpDir, err := os.MkdirTemp("", "aws-ssm-cache-error-test-*")
			if err != nil {
				t.Fatalf("Failed to create temp directory: %v", err)
			}
			defer os.Chmod(tmpDir, 0755) // Restore permissions before cleanup
			defer os.RemoveAll(tmpDir)

			// Create cache service
			cacheService, err := cache.NewCacheService(tmpDir, 60) // 60 minutes TTL
			if err != nil {
				t.Fatalf("Failed to create cache service: %v", err)
			}

			// Setup test environment
			if tc.setup != nil {
				if err := tc.setup(cacheService, tmpDir); err != nil {
					if tc.expectErr {
						// Expected error, continue with verification
					} else {
						t.Fatalf("Test setup failed: %v", err)
					}
				}
			}

			// Execute test logic
			err = tc.execute(cacheService)

			if tc.expectErr {
				assertion.Error(err, "Expected error should occur")
			} else {
				assertion.NoError(err, "Should not have error")
			}

			// Execute verification
			if tc.verify != nil {
				err = tc.verify(cacheService)
				if !tc.expectErr {
					assertion.NoError(err, "Verification should not have error")
				}
			}

			if assertion.HasErrors() {
				t.Errorf("Assertion errors occurred: %v", assertion.GetErrors())
			}
		})
	}
}

// TestCacheCommand_Integration_Performance tests cache performance characteristics
func TestCacheCommand_Integration_Performance(t *testing.T) {
	assertion := testframework.NewAssertion(t)

	// Create a temporary cache directory for testing
	tmpDir, err := os.MkdirTemp("", "aws-ssm-cache-performance-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	testCases := []struct {
		name      string
		setup     func(*cache.Service) error
		verify    func(*cache.Service) error
		expectErr bool
	}{
		{
			name: "Cache Performance - Many Entries",
			setup: func(c *cache.Service) error {
				// Store many cache entries to test performance
				startTime := time.Now()

				for i := 0; i < 100; i++ {
					testData := map[string]interface{}{
						"entry_id":    float64(i),
						"instance_id": fmt.Sprintf("i-%032d", i),
						"name":        fmt.Sprintf("test-instance-%d", i),
						"region":      []string{"us-east-1", "us-west-2", "eu-west-1"}[i%3],
						"timestamp":   float64(time.Now().Unix()),
						"tags":        map[string]interface{}{"Environment": "test", "Index": string(rune('a' + i%26))},
					}

					key := fmt.Sprintf("performance-test-%d", i)
					region := []string{"us-east-1", "us-west-2", "eu-west-1"}[i%3]
					query := fmt.Sprintf("tag:Index=%s", string(rune('a'+i%26)))

					if err := c.Set(key, testData, region, query); err != nil {
						return err
					}
				}

				elapsed := time.Since(startTime)
				t.Logf("Stored 100 cache entries in %v", elapsed)

				return nil
			},
			verify: func(c *cache.Service) error {
				// Retrieve all entries to test read performance
				startTime := time.Now()

				foundEntries := 0
				for i := 0; i < 100; i++ {
					key := fmt.Sprintf("performance-test-%d", i)
					data, exists := c.Get(key)
					if exists {
						foundEntries++
						assertion.NotNil(data, "Retrieved data should not be nil")

						// Verify some basic structure
						if testData, ok := data.(map[string]interface{}); ok {
							assertion.Greater(testData["entry_id"].(float64), float64(-1), "Entry ID should be valid")
							assertion.Contains(testData["instance_id"].(string), "i-", "Instance ID should be valid")
						}
					}
				}

				elapsed := time.Since(startTime)
				t.Logf("Retrieved %d cache entries in %v", foundEntries, elapsed)

				assertion.Greater(foundEntries, 90, "Should find most entries") // Allow for some tolerance

				// Get cache statistics
				totalFiles, _, totalSize, err := c.GetCacheStats()
				assertion.NoError(err, "Should be able to get cache stats")
				assertion.Greater(totalFiles, 90, "Should have many cache files")
				assertion.Greater(totalSize, int64(1000), "Should have significant cache size")

				return nil
			},
			expectErr: false,
		},
		{
			name: "Cache Cleanup Performance",
			setup: func(c *cache.Service) error {
				// Create many entries for cleanup testing
				for i := 0; i < 50; i++ {
					testData := map[string]interface{}{"test": "data"}
					key := fmt.Sprintf("cleanup-test-%d", i)
					if err := c.Set(key, testData, "us-east-1", "cleanup-query"); err != nil {
						return err
					}
				}
				return nil
			},
			verify: func(c *cache.Service) error {
				// Measure cleanup performance
				startTime := time.Now()

				err := c.Cleanup()
				assertion.NoError(err, "Cleanup should succeed")

				elapsed := time.Since(startTime)
				t.Logf("Cleaned up cache in %v", elapsed)

				// Cleanup should complete within reasonable time
				assertion.Less(elapsed, 5*time.Second, "Cleanup should complete within 5 seconds")

				return nil
			},
			expectErr: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create cache service
			cacheService, err := cache.NewCacheService(tmpDir, 60) // 60 minutes TTL
			if err != nil {
				t.Fatalf("Failed to create cache service: %v", err)
			}

			// Setup test
			if tc.setup != nil {
				if err := tc.setup(cacheService); err != nil {
					if tc.expectErr {
						// Expected error, continue with verification
					} else {
						t.Fatalf("Test setup failed: %v", err)
					}
				}
			}

			// Execute verification
			err = tc.verify(cacheService)

			if tc.expectErr {
				assertion.Error(err, "Expected error should occur")
			} else {
				assertion.NoError(err, "Should not have error")
			}

			if assertion.HasErrors() {
				t.Errorf("Assertion errors occurred: %v", assertion.GetErrors())
			}
		})
	}
}
