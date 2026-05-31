package cache

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

func TestEnhancedService_StaleWhileRevalidate(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "enhanced-cache-test-*")
	if err != nil {
		t.Fatalf("should create temp directory: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	cfg := DefaultEnhancedCacheConfig()
	cfg.CacheDir = tmpDir
	cfg.StaleThreshold = 100 * time.Millisecond

	cache, err := NewEnhancedCacheService(cfg)
	if err != nil {
		t.Fatalf("should create enhanced cache service: %v", err)
	}
	defer cache.Close()

	var mu sync.Mutex
	refreshCalled := false
	refreshFn := func(context.Context, string, string, string) (interface{}, error) {
		mu.Lock()
		refreshCalled = true
		mu.Unlock()
		time.Sleep(50 * time.Millisecond) // Simulate some work
		return map[string]string{"refreshed": "data"}, nil
	}

	// First access - populate cache
	_, found, isStale := cache.GetWithRefresh("test-key", "us-east-1", "test", refreshFn)
	if found {
		t.Error("should not find data on first access")
	}
	if isStale {
		t.Error("data should not be stale on first access")
	}

	// Wait for data to become stale
	time.Sleep(150 * time.Millisecond)

	// Reset flag
	refreshCalled = false

	// Second access - should return stale data immediately and trigger background refresh
	data, found, isStale := cache.GetWithRefresh("test-key", "us-east-1", "test", refreshFn)
	if !found {
		t.Error("should find data on second access")
	}
	if !isStale {
		t.Error("data should be stale on second access")
	}
	if data == nil {
		t.Error("data should not be nil")
	}

	// Wait for background refresh to complete
	time.Sleep(200 * time.Millisecond)

	// Refresh should have been triggered
	mu.Lock()
	called := refreshCalled
	mu.Unlock()
	if !called {
		t.Error("refresh function should have been called for stale data")
	}

	t.Log("Stale-while-revalidate test completed successfully")
}

func TestEnhancedService_BackgroundRefresh(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "enhanced-cache-refresh-test-*")
	if err != nil {
		t.Fatalf("should create temp directory: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	cfg := DefaultEnhancedCacheConfig()
	cfg.CacheDir = tmpDir
	cfg.StaleThreshold = 100 * time.Millisecond
	cfg.RefreshInterval = 50 * time.Millisecond

	cache, err := NewEnhancedCacheService(cfg)
	if err != nil {
		t.Fatalf("should create enhanced cache service: %v", err)
	}
	defer cache.Close()

	var mu sync.Mutex
	refreshCallCount := 0
	refreshFn := func(context.Context, string, string, string) (interface{}, error) {
		mu.Lock()
		refreshCallCount++
		count := refreshCallCount
		mu.Unlock()
		return map[string]string{"refreshed": "data", "count": string(rune(count))}, nil
	}

	// First access - populate cache
	cached, found, _ := cache.GetWithRefresh("refresh-test", "us-east-1", "test", refreshFn)
	if found {
		t.Error("should not find data on first access")
	}
	if cached == nil {
		t.Error("data should not be nil after first refresh")
	}

	// Wait for data to become stale
	time.Sleep(150 * time.Millisecond)

	// Access again - should trigger background refresh
	_, found, isStale := cache.GetWithRefresh("refresh-test", "us-east-1", "test", refreshFn)
	if !found {
		t.Error("should find data on second access")
	}
	if !isStale {
		t.Error("data should be stale on second access")
	}

	// Wait for background refresh to complete
	time.Sleep(200 * time.Millisecond)

	// Check that refresh was called (initial + background)
	mu.Lock()
	count := refreshCallCount
	mu.Unlock()
	if count < 2 {
		t.Errorf("refresh should have been called at least 2 times, got %d", count)
	}

	// Check metrics
	metrics := cache.GetMetrics()
	t.Logf("Background refresh test completed: refreshes=%d, stale_hits=%d",
		metrics.backgroundRefreshes, metrics.staleHits)
}

func TestEnhancedCache_Metrics(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "enhanced-cache-metrics-test-*")
	if err != nil {
		t.Fatalf("should create temp directory: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	cfg := DefaultEnhancedCacheConfig()
	cfg.CacheDir = tmpDir

	cache, err := NewEnhancedCacheService(cfg)
	if err != nil {
		t.Fatalf("should create enhanced cache service: %v", err)
	}
	defer cache.Close()

	refreshFn := func(context.Context, string, string, string) (interface{}, error) {
		return map[string]string{"test": "data"}, nil
	}

	// Generate some cache activity
	for i := 0; i < 10; i++ {
		cache.GetWithRefresh("test-key", "us-east-1", "test", refreshFn)
	}

	// Store some data
	_ = cache.Set("stored-key", map[string]string{"stored": "value"}, "us-east-1", "test")

	// Access stored data
	for i := 0; i < 5; i++ {
		cache.GetWithRefresh("stored-key", "us-east-1", "test", refreshFn)
	}

	metrics := cache.GetMetrics()
	hitRate := cache.GetHitRate()

	total := metrics.hits + metrics.misses + metrics.staleHits
	if total == 0 {
		t.Error("should have recorded cache accesses")
	}
	if hitRate < 0 || hitRate > 100 {
		t.Errorf("hit rate should be between 0 and 100, got %.2f", hitRate)
	}

	t.Logf("Metrics: hits=%d, misses=%d, stale_hits=%d, hit_rate=%.2f%%",
		metrics.hits, metrics.misses, metrics.staleHits, hitRate)
}

func TestEnhancedCacheRejectsTraversalKeys(t *testing.T) {
	parent := t.TempDir()
	cacheDir := filepath.Join(parent, "cache")
	outsideDir := filepath.Join(parent, "cache_evil")
	if err := os.MkdirAll(outsideDir, 0700); err != nil {
		t.Fatalf("create outside cache dir: %v", err)
	}

	outsideEntry := EnhancedEntry{
		Data:        map[string]string{"outside": "value"},
		Timestamp:   time.Now(),
		Region:      "us-west-2",
		Query:       "query",
		LastRefresh: time.Now(),
	}
	payload, err := json.Marshal(outsideEntry)
	if err != nil {
		t.Fatalf("marshal outside entry: %v", err)
	}
	if err := os.WriteFile(filepath.Join(outsideDir, "pwn.json"), payload, 0600); err != nil {
		t.Fatalf("write outside cache entry: %v", err)
	}

	cfg := DefaultEnhancedCacheConfig()
	cfg.CacheDir = cacheDir
	enhanced, err := NewEnhancedCacheService(cfg)
	if err != nil {
		t.Fatalf("create enhanced cache service: %v", err)
	}
	defer enhanced.Close()

	refreshFn := func(context.Context, string, string, string) (interface{}, error) {
		return map[string]string{"refreshed": "data"}, nil
	}
	data, found, stale := enhanced.GetWithRefresh("../cache_evil/pwn", "us-west-2", "query", refreshFn)
	if found || stale || data != nil {
		t.Fatalf("expected traversal key to miss, got data=%v found=%v stale=%v", data, found, stale)
	}
}
