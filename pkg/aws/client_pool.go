package aws

import (
	"context"
	"fmt"
	"sync"
	"time"

	awssdk "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	appconfig "github.com/johnlam90/aws-ssm/pkg/config"
	"github.com/johnlam90/aws-ssm/pkg/logging"
)

// ClientPool manages a pool of AWS clients for connection reuse
type ClientPool struct {
	mu            sync.RWMutex
	ec2Clients    map[string]*ec2.Client
	ssmClients    map[string]*ssm.Client
	configs       map[string]awssdk.Config
	maxPoolSize   int
	clientTTL     time.Duration
	lastAccess    map[string]time.Time
	cleanupTicker *time.Ticker
	cleanupStop   chan struct{}
	cleanupDone   chan struct{}
	logger        logging.Logger
	metrics       *PoolMetrics
}

// PoolMetrics tracks client pool statistics
type PoolMetrics struct {
	mu              sync.RWMutex
	hits            int64
	misses          int64
	evictions       int64
	creations       int64
	poolSize        int
	ec2PoolSize     int
	ssmPoolSize     int
	lastCleanupTime time.Time
}

// ClientPoolConfig holds configuration for the client pool
type ClientPoolConfig struct {
	MaxPoolSize     int
	ClientTTL       time.Duration
	CleanupInterval time.Duration
	EnableMetrics   bool
}

// DefaultClientPoolConfig returns default client pool configuration
func DefaultClientPoolConfig() *ClientPoolConfig {
	return &ClientPoolConfig{
		MaxPoolSize:     50,               // Maximum 50 clients in pool
		ClientTTL:       30 * time.Minute, // Clients expire after 30 minutes
		CleanupInterval: 5 * time.Minute,  // Cleanup every 5 minutes
		EnableMetrics:   true,
	}
}

// NewClientPool creates a new client pool
func NewClientPool(cfg *ClientPoolConfig) *ClientPool {
	if cfg == nil {
		cfg = DefaultClientPoolConfig()
	}

	pool := &ClientPool{
		ec2Clients:  make(map[string]*ec2.Client),
		ssmClients:  make(map[string]*ssm.Client),
		configs:     make(map[string]awssdk.Config),
		lastAccess:  make(map[string]time.Time),
		maxPoolSize: cfg.MaxPoolSize,
		clientTTL:   cfg.ClientTTL,
		cleanupStop: make(chan struct{}),
		cleanupDone: make(chan struct{}),
		logger:      logging.With(logging.String("component", "client_pool")),
		metrics:     &PoolMetrics{},
	}

	// Start background cleanup goroutine
	pool.cleanupTicker = time.NewTicker(cfg.CleanupInterval)
	go pool.cleanupLoop()

	pool.logger.Info("Client pool initialized",
		logging.Int("max_pool_size", cfg.MaxPoolSize),
		logging.Duration("client_ttl", cfg.ClientTTL),
		logging.Duration("cleanup_interval", cfg.CleanupInterval))

	return pool
}

// generateKey creates a unique key for client caching
func (cp *ClientPool) generateKey(region, profile string) string {
	return fmt.Sprintf("%s:%s", region, profile)
}

// GetOrCreateClient gets an existing client from the pool or creates a new one
func (cp *ClientPool) GetOrCreateClient(ctx context.Context, region, profile, configPath string) (*Client, error) {
	key := cp.generateKey(region, profile)

	// Try to get from pool first (read lock)
	cp.mu.RLock()
	if ec2Client, exists := cp.ec2Clients[key]; exists {
		ssmClient := cp.ssmClients[key]
		config := cp.configs[key]
		cp.mu.RUnlock()

		// Update last access time
		cp.mu.Lock()
		cp.lastAccess[key] = time.Now()
		cp.mu.Unlock()

		cp.recordHit()
		cp.logger.Debug("Client pool hit", logging.String("key", key))

		// Return existing client
		client := &Client{
			EC2Client:      ec2Client,
			SSMClient:      ssmClient,
			Config:         config,
			CircuitBreaker: NewCircuitBreaker(DefaultCircuitBreakerConfig()),
		}

		// Load app config (not pooled as it's lightweight)
		appCfg, err := loadAppConfig(configPath)
		if err != nil {
			return nil, err
		}
		client.AppConfig = appCfg

		return client, nil
	}
	cp.mu.RUnlock()

	// Not in pool, create new client
	cp.recordMiss()
	cp.logger.Debug("Client pool miss, creating new client", logging.String("key", key))

	// Check pool size before creating
	if err := cp.ensurePoolCapacity(); err != nil {
		return nil, err
	}

	// Create new AWS client
	client, err := NewClient(ctx, region, profile, configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create AWS client: %w", err)
	}

	// Store in pool
	cp.mu.Lock()
	cp.ec2Clients[key] = client.EC2Client
	cp.ssmClients[key] = client.SSMClient
	cp.configs[key] = client.Config
	cp.lastAccess[key] = time.Now()
	cp.mu.Unlock()

	cp.recordCreation()
	cp.updatePoolSize()

	cp.logger.Info("New client added to pool",
		logging.String("key", key),
		logging.Int("pool_size", cp.getPoolSize()))

	return client, nil
}

// ensurePoolCapacity evicts oldest clients if pool is full
func (cp *ClientPool) ensurePoolCapacity() error {
	cp.mu.Lock()
	defer cp.mu.Unlock()

	if len(cp.ec2Clients) < cp.maxPoolSize {
		return nil
	}

	// Find oldest client
	var oldestKey string
	var oldestTime time.Time
	for key, lastAccess := range cp.lastAccess {
		if oldestKey == "" || lastAccess.Before(oldestTime) {
			oldestKey = key
			oldestTime = lastAccess
		}
	}

	if oldestKey != "" {
		// Evict oldest client
		delete(cp.ec2Clients, oldestKey)
		delete(cp.ssmClients, oldestKey)
		delete(cp.configs, oldestKey)
		delete(cp.lastAccess, oldestKey)

		cp.logger.Info("Evicted oldest client from pool",
			logging.String("key", oldestKey),
			logging.Time("last_access", oldestTime))

		cp.recordEviction()
	}

	return nil
}

// cleanupLoop periodically removes expired clients
func (cp *ClientPool) cleanupLoop() {
	defer close(cp.cleanupDone)
	for {
		select {
		case <-cp.cleanupTicker.C:
			cp.cleanup()
		case <-cp.cleanupStop:
			cp.cleanupTicker.Stop()
			return
		}
	}
}

// cleanup removes expired clients from the pool
func (cp *ClientPool) cleanup() {
	cp.mu.Lock()

	now := time.Now()
	expiredKeys := []string{}

	for key, lastAccess := range cp.lastAccess {
		if now.Sub(lastAccess) > cp.clientTTL {
			expiredKeys = append(expiredKeys, key)
		}
	}

	if len(expiredKeys) > 0 {
		for _, key := range expiredKeys {
			delete(cp.ec2Clients, key)
			delete(cp.ssmClients, key)
			delete(cp.configs, key)
			delete(cp.lastAccess, key)
			cp.recordEviction()
		}

		cp.logger.Info("Cleaned up expired clients",
			logging.Int("count", len(expiredKeys)),
			logging.Int("remaining", len(cp.ec2Clients)))
	}

	// Get pool size while we still hold the lock
	poolSize := len(cp.ec2Clients)
	cp.mu.Unlock()

	// Update metrics after releasing the lock
	cp.metrics.mu.Lock()
	cp.metrics.lastCleanupTime = now
	cp.metrics.poolSize = poolSize
	cp.metrics.ec2PoolSize = poolSize
	cp.metrics.ssmPoolSize = poolSize
	cp.metrics.mu.Unlock()
}

// Close stops the cleanup goroutine and clears the pool
func (cp *ClientPool) Close() {
	// Signal cleanup goroutine to stop
	close(cp.cleanupStop)

	// Wait for cleanup goroutine to finish
	<-cp.cleanupDone

	// Now safe to clear the pool
	cp.mu.Lock()
	defer cp.mu.Unlock()

	cp.ec2Clients = make(map[string]*ec2.Client)
	cp.ssmClients = make(map[string]*ssm.Client)
	cp.configs = make(map[string]awssdk.Config)
	cp.lastAccess = make(map[string]time.Time)

	cp.logger.Info("Client pool closed")
}

// GetMetrics returns pool metrics
func (cp *ClientPool) GetMetrics() PoolMetrics {
	cp.metrics.mu.RLock()
	defer cp.metrics.mu.RUnlock()

	return PoolMetrics{
		hits:            cp.metrics.hits,
		misses:          cp.metrics.misses,
		evictions:       cp.metrics.evictions,
		creations:       cp.metrics.creations,
		poolSize:        cp.metrics.poolSize,
		ec2PoolSize:     cp.metrics.ec2PoolSize,
		ssmPoolSize:     cp.metrics.ssmPoolSize,
		lastCleanupTime: cp.metrics.lastCleanupTime,
	}
}

// GetHitRate returns the cache hit rate as a percentage
func (cp *ClientPool) GetHitRate() float64 {
	cp.metrics.mu.RLock()
	defer cp.metrics.mu.RUnlock()

	total := cp.metrics.hits + cp.metrics.misses
	if total == 0 {
		return 0
	}
	return float64(cp.metrics.hits) / float64(total) * 100
}

// recordHit records a cache hit
func (cp *ClientPool) recordHit() {
	cp.metrics.mu.Lock()
	cp.metrics.hits++
	cp.metrics.mu.Unlock()
}

// recordMiss records a cache miss
func (cp *ClientPool) recordMiss() {
	cp.metrics.mu.Lock()
	cp.metrics.misses++
	cp.metrics.mu.Unlock()
}

// recordEviction records a client eviction
func (cp *ClientPool) recordEviction() {
	cp.metrics.mu.Lock()
	cp.metrics.evictions++
	cp.metrics.mu.Unlock()
}

// recordCreation records a client creation
func (cp *ClientPool) recordCreation() {
	cp.metrics.mu.Lock()
	cp.metrics.creations++
	cp.metrics.mu.Unlock()
}

// updatePoolSize updates the pool size metrics
func (cp *ClientPool) updatePoolSize() {
	cp.mu.RLock()
	ec2Size := len(cp.ec2Clients)
	ssmSize := len(cp.ssmClients)
	cp.mu.RUnlock()

	cp.metrics.mu.Lock()
	cp.metrics.poolSize = ec2Size // ec2 and ssm should be equal
	cp.metrics.ec2PoolSize = ec2Size
	cp.metrics.ssmPoolSize = ssmSize
	cp.metrics.mu.Unlock()
}

// getPoolSize returns the current pool size
func (cp *ClientPool) getPoolSize() int {
	cp.mu.RLock()
	defer cp.mu.RUnlock()
	return len(cp.ec2Clients)
}

// loadAppConfig is a helper to load app config (extracted to avoid duplication)
func loadAppConfig(configPath string) (*appconfig.Config, error) {
	return appconfig.LoadConfig(configPath)
}
