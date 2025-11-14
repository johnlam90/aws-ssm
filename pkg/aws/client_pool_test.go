package aws

import (
	"context"
	"testing"
	"time"
)

func TestClientPool_GetOrCreateClient(t *testing.T) {
	cfg := &ClientPoolConfig{
		MaxPoolSize:     5,
		ClientTTL:       1 * time.Minute,
		CleanupInterval: 10 * time.Second,
		EnableMetrics:   true,
	}

	pool := NewClientPool(cfg)
	defer pool.Close()

	ctx := context.Background()

	// Test: First client creation
	client1, err := pool.GetOrCreateClient(ctx, "us-east-1", "default", "")
	if err != nil {
		t.Fatalf("should create client without error: %v", err)
	}
	if client1 == nil {
		t.Fatal("client should not be nil")
	}

	// Test: Pool hit - same region and profile
	client2, err := pool.GetOrCreateClient(ctx, "us-east-1", "default", "")
	if err != nil {
		t.Fatalf("should get cached client without error: %v", err)
	}
	if client2 == nil {
		t.Fatal("cached client should not be nil")
	}

	// Verify metrics show a hit
	metrics := pool.GetMetrics()
	if metrics.hits == 0 {
		t.Error("should have at least one cache hit")
	}

	// Test: Pool miss - different region
	client3, err := pool.GetOrCreateClient(ctx, "us-west-2", "default", "")
	if err != nil {
		t.Fatalf("should create new client for different region: %v", err)
	}
	if client3 == nil {
		t.Fatal("new client should not be nil")
	}

	// Verify pool size
	if pool.getPoolSize() > cfg.MaxPoolSize {
		t.Errorf("pool size should not exceed maximum: got %d, max %d", pool.getPoolSize(), cfg.MaxPoolSize)
	}

	t.Logf("Pool metrics: hits=%d, misses=%d, pool_size=%d", metrics.hits, metrics.misses, pool.getPoolSize())
}

func TestClientPool_EvictionOnFullPool(t *testing.T) {
	cfg := &ClientPoolConfig{
		MaxPoolSize:     3, // Small pool for testing
		ClientTTL:       1 * time.Minute,
		CleanupInterval: 10 * time.Second,
		EnableMetrics:   true,
	}

	pool := NewClientPool(cfg)
	defer pool.Close()

	ctx := context.Background()

	// Fill the pool
	regions := []string{"us-east-1", "us-west-2", "eu-west-1", "ap-southeast-1"}
	for _, region := range regions {
		_, err := pool.GetOrCreateClient(ctx, region, "default", "")
		if err != nil {
			t.Fatalf("should create client for region %s: %v", region, err)
		}
	}

	// Pool should be at max size (oldest evicted)
	if pool.getPoolSize() > cfg.MaxPoolSize {
		t.Errorf("pool size should not exceed max size: got %d, max %d", pool.getPoolSize(), cfg.MaxPoolSize)
	}

	metrics := pool.GetMetrics()
	if metrics.evictions == 0 {
		t.Error("should have evicted at least one client")
	}

	t.Logf("Evictions: %d, final pool size: %d", metrics.evictions, pool.getPoolSize())
}

func TestClientPool_HitRate(t *testing.T) {
	cfg := DefaultClientPoolConfig()
	pool := NewClientPool(cfg)
	defer pool.Close()

	ctx := context.Background()

	// Create client - miss
	_, err := pool.GetOrCreateClient(ctx, "us-east-1", "default", "")
	if err != nil {
		t.Fatalf("first client creation should succeed: %v", err)
	}

	// Access same client multiple times - hits
	for i := 0; i < 5; i++ {
		_, err := pool.GetOrCreateClient(ctx, "us-east-1", "default", "")
		if err != nil {
			t.Fatalf("cached client access %d should succeed: %v", i, err)
		}
	}

	hitRate := pool.GetHitRate()
	if hitRate <= 80.0 {
		t.Errorf("hit rate should be > 80%%, got %.2f%%", hitRate)
	}

	t.Logf("Hit rate: %.2f%%", hitRate)
}

func TestClientPool_Cleanup(t *testing.T) {
	cfg := &ClientPoolConfig{
		MaxPoolSize:     10,
		ClientTTL:       100 * time.Millisecond, // Very short TTL for testing
		CleanupInterval: 50 * time.Millisecond,  // Frequent cleanup
		EnableMetrics:   true,
	}

	pool := NewClientPool(cfg)
	ctx := context.Background()

	// Create some clients
	for i := 0; i < 3; i++ {
		region := []string{"us-east-1", "us-west-2", "eu-west-1"}[i]
		_, err := pool.GetOrCreateClient(ctx, region, "default", "")
		if err != nil {
			t.Fatalf("should create client for region %s: %v", region, err)
		}
	}

	initialSize := pool.getPoolSize()
	if initialSize != 3 {
		t.Errorf("should have 3 clients in pool, got %d", initialSize)
	}

	// Wait for TTL expiration and automatic cleanup cycles
	// The background cleanup goroutine will run multiple times
	time.Sleep(300 * time.Millisecond)

	// Close pool and check final size
	pool.Close()

	finalSize := pool.getPoolSize()
	if finalSize >= initialSize {
		t.Errorf("pool size should decrease after cleanup: initial=%d, final=%d", initialSize, finalSize)
	}

	t.Logf("Pool size after cleanup: initial=%d, final=%d", initialSize, finalSize)
}
