package cache

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/johnlam90/aws-ssm/pkg/logging"
)

// EnhancedEntry represents a cached item with refresh metadata
type EnhancedEntry struct {
	Data           interface{} `json:"data"`
	Timestamp      time.Time   `json:"timestamp"`
	Region         string      `json:"region"`
	Query          string      `json:"query"`
	LastRefresh    time.Time   `json:"last_refresh"`
	RefreshAttempt int         `json:"refresh_attempt"`
	IsStale        bool        `json:"is_stale"`
}

// RefreshFunc is a function type for background cache refresh
type RefreshFunc func(ctx context.Context, key, region, query string) (interface{}, error)

// EnhancedService provides caching with background refresh capabilities
type EnhancedService struct {
	*Service // Embed base cache service

	mu                sync.RWMutex
	refreshFuncs      map[string]RefreshFunc
	refreshQueue      chan refreshRequest
	refreshWorkers    int
	refreshInterval   time.Duration
	staleThreshold    time.Duration
	stopRefresh       chan struct{}
	refreshInProgress map[string]bool
	logger            logging.Logger
	metrics           *EnhancedMetrics
}

// EnhancedMetrics tracks enhanced cache statistics
type EnhancedMetrics struct {
	mu                   sync.RWMutex
	hits                 int64
	misses               int64
	staleHits            int64
	refreshes            int64
	refreshSuccesses     int64
	refreshFailures      int64
	backgroundRefreshes  int64
	lastRefreshTime      time.Time
	avgRefreshDuration   time.Duration
	totalRefreshDuration time.Duration
}

// refreshRequest represents a background refresh request
type refreshRequest struct {
	key    string
	region string
	query  string
	fn     RefreshFunc
}

// EnhancedCacheConfig holds configuration for enhanced cache
type EnhancedCacheConfig struct {
	CacheDir        string
	TTLMinutes      int
	RefreshWorkers  int
	RefreshInterval time.Duration
	StaleThreshold  time.Duration
}

// DefaultEnhancedCacheConfig returns default enhanced cache configuration
func DefaultEnhancedCacheConfig() *EnhancedCacheConfig {
	return &EnhancedCacheConfig{
		CacheDir:        "",
		TTLMinutes:      5,
		RefreshWorkers:  3,
		RefreshInterval: 4 * time.Minute, // Refresh 1 minute before TTL expiration
		StaleThreshold:  6 * time.Minute, // Consider stale after 6 minutes
	}
}

// NewEnhancedCacheService creates a new enhanced cache service
func NewEnhancedCacheService(cfg *EnhancedCacheConfig) (*EnhancedService, error) {
	if cfg == nil {
		cfg = DefaultEnhancedCacheConfig()
	}

	// Create base cache service
	baseService, err := NewCacheService(cfg.CacheDir, cfg.TTLMinutes)
	if err != nil {
		return nil, err
	}

	enhanced := &EnhancedService{
		Service:           baseService,
		refreshFuncs:      make(map[string]RefreshFunc),
		refreshQueue:      make(chan refreshRequest, 100),
		refreshWorkers:    cfg.RefreshWorkers,
		refreshInterval:   cfg.RefreshInterval,
		staleThreshold:    cfg.StaleThreshold,
		stopRefresh:       make(chan struct{}),
		refreshInProgress: make(map[string]bool),
		logger:            logging.With(logging.String("component", "enhanced_cache")),
		metrics:           &EnhancedMetrics{},
	}

	// Start background refresh workers
	for i := 0; i < cfg.RefreshWorkers; i++ {
		go enhanced.refreshWorker(i)
	}

	enhanced.logger.Info("Enhanced cache service initialized",
		logging.Int("refresh_workers", cfg.RefreshWorkers),
		logging.Duration("refresh_interval", cfg.RefreshInterval),
		logging.Duration("stale_threshold", cfg.StaleThreshold))

	return enhanced, nil
}

// GetWithRefresh retrieves cached data with stale-while-revalidate pattern
// Returns stale data immediately if available, triggers background refresh if needed
// If no cache exists, performs synchronous refresh and caches the result
func (ec *EnhancedService) GetWithRefresh(key, region, query string, refreshFn RefreshFunc) (interface{}, bool, bool) {
	cacheFile := filepath.Join(ec.cacheDir, key+".json")
	cleanPath := filepath.Clean(cacheFile)

	if !strings.HasPrefix(cleanPath, filepath.Clean(ec.cacheDir)) {
		ec.recordMiss()
		return nil, false, false
	}

	// Check file size before reading
	const maxCacheFileSize = 10 * 1024 * 1024 // 10 MB limit
	fileInfo, err := os.Stat(cleanPath)
	if err != nil {
		if os.IsNotExist(err) {
			// Cache miss - perform synchronous refresh
			ec.recordMiss()
			data, refreshErr := ec.performSynchronousRefresh(key, region, query, refreshFn)
			if refreshErr != nil {
				return nil, false, false
			}
			return data, false, false // found (just refreshed), not stale
		}
		ec.logger.Warn("Failed to stat cache file", logging.String("file", cacheFile), logging.String("error", err.Error()))
		ec.recordMiss()
		return nil, false, false
	}

	if fileInfo.Size() > maxCacheFileSize {
		ec.logger.Warn("Cache file exceeds size limit",
			logging.String("file", cacheFile),
			logging.Int64("size", fileInfo.Size()))
		ec.recordMiss()
		return nil, false, false
	}

	data, err := os.ReadFile(cleanPath)
	if err != nil {
		ec.logger.Warn("Failed to read cache file", logging.String("file", cacheFile), logging.String("error", err.Error()))
		ec.recordMiss()
		return nil, false, false
	}

	var entry EnhancedEntry
	if unmarshalErr := json.Unmarshal(data, &entry); unmarshalErr != nil {
		ec.logger.Warn("Failed to unmarshal cache entry", logging.String("file", cacheFile), logging.String("error", unmarshalErr.Error()))
		ec.recordMiss()
		return nil, false, false
	}

	age := time.Since(entry.Timestamp)
	isStale := age > ec.staleThreshold

	// Check if data is completely expired (beyond stale threshold)
	if age > ec.ttl+ec.staleThreshold {
		ec.logger.Debug("Cache entry expired",
			logging.String("key", key),
			logging.Duration("age", age))
		ec.recordMiss()
		// Trigger background refresh
		ec.triggerBackgroundRefresh(key, region, query, refreshFn)
		return nil, false, false
	}

	// Data is valid or stale but usable
	if isStale {
		ec.recordStaleHit()
		ec.logger.Debug("Returning stale data and triggering refresh",
			logging.String("key", key),
			logging.Duration("age", age))
		// Trigger background refresh for stale data
		ec.triggerBackgroundRefresh(key, region, query, refreshFn)
		return entry.Data, true, true // found, isStale
	}

	// Check if we should proactively refresh (before becoming stale)
	if age > ec.refreshInterval {
		ec.logger.Debug("Proactively refreshing cache",
			logging.String("key", key),
			logging.Duration("age", age))
		ec.triggerBackgroundRefresh(key, region, query, refreshFn)
	}

	ec.recordHit()
	return entry.Data, true, false // found, not stale
}

// performSynchronousRefresh performs an immediate synchronous refresh
func (ec *EnhancedService) performSynchronousRefresh(key, region, query string, refreshFn RefreshFunc) (interface{}, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Execute refresh function
	data, err := refreshFn(ctx, key, region, query)
	if err != nil {
		ec.recordRefreshFailure()
		return nil, err
	}

	// Store refreshed data
	if err := ec.Set(key, data, region, query); err != nil {
		ec.recordRefreshFailure()
		return nil, err
	}

	ec.recordRefreshSuccess(0) // Duration not tracked for sync refresh
	return data, nil
}

// triggerBackgroundRefresh queues a background refresh request
func (ec *EnhancedService) triggerBackgroundRefresh(key, region, query string, refreshFn RefreshFunc) {
	ec.mu.Lock()
	// Check if refresh is already in progress
	if ec.refreshInProgress[key] {
		ec.mu.Unlock()
		ec.logger.Debug("Refresh already in progress", logging.String("key", key))
		return
	}
	ec.refreshInProgress[key] = true
	ec.mu.Unlock()

	// Queue refresh request (non-blocking)
	select {
	case ec.refreshQueue <- refreshRequest{
		key:    key,
		region: region,
		query:  query,
		fn:     refreshFn,
	}:
		ec.logger.Debug("Queued background refresh", logging.String("key", key))
	default:
		// Queue is full, skip this refresh
		ec.mu.Lock()
		delete(ec.refreshInProgress, key)
		ec.mu.Unlock()
		ec.logger.Warn("Refresh queue full, skipping refresh", logging.String("key", key))
	}
}

// refreshWorker processes background refresh requests
func (ec *EnhancedService) refreshWorker(id int) {
	ec.logger.Info("Started refresh worker", logging.Int("worker_id", id))

	for {
		select {
		case req := <-ec.refreshQueue:
			ec.processRefresh(req)
		case <-ec.stopRefresh:
			ec.logger.Info("Stopping refresh worker", logging.Int("worker_id", id))
			return
		}
	}
}

// processRefresh executes a background refresh
func (ec *EnhancedService) processRefresh(req refreshRequest) {
	startTime := time.Now()

	defer func() {
		ec.mu.Lock()
		delete(ec.refreshInProgress, req.key)
		ec.mu.Unlock()
	}()

	ec.logger.Debug("Processing background refresh", logging.String("key", req.key))

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Execute refresh function
	data, err := req.fn(ctx, req.key, req.region, req.query)
	duration := time.Since(startTime)

	if err != nil {
		ec.recordRefreshFailure()
		ec.logger.Warn("Background refresh failed",
			logging.String("key", req.key),
			logging.String("error", err.Error()),
			logging.Duration("duration", duration))
		return
	}

	// Store refreshed data
	if err := ec.Set(req.key, data, req.region, req.query); err != nil {
		ec.recordRefreshFailure()
		ec.logger.Error("Failed to store refreshed data",
			logging.String("key", req.key),
			logging.String("error", err.Error()))
		return
	}

	ec.recordRefreshSuccess(duration)
	ec.logger.Info("Background refresh completed",
		logging.String("key", req.key),
		logging.Duration("duration", duration))
}

// Close stops the background refresh workers
func (ec *EnhancedService) Close() {
	close(ec.stopRefresh)
	ec.logger.Info("Enhanced cache service stopped")
}

// GetMetrics returns enhanced cache metrics
func (ec *EnhancedService) GetMetrics() EnhancedMetrics {
	ec.metrics.mu.RLock()
	defer ec.metrics.mu.RUnlock()

	return EnhancedMetrics{
		hits:                 ec.metrics.hits,
		misses:               ec.metrics.misses,
		staleHits:            ec.metrics.staleHits,
		refreshes:            ec.metrics.refreshes,
		refreshSuccesses:     ec.metrics.refreshSuccesses,
		refreshFailures:      ec.metrics.refreshFailures,
		backgroundRefreshes:  ec.metrics.backgroundRefreshes,
		lastRefreshTime:      ec.metrics.lastRefreshTime,
		avgRefreshDuration:   ec.metrics.avgRefreshDuration,
		totalRefreshDuration: ec.metrics.totalRefreshDuration,
	}
}

// GetHitRate returns the cache hit rate as a percentage
func (ec *EnhancedService) GetHitRate() float64 {
	ec.metrics.mu.RLock()
	defer ec.metrics.mu.RUnlock()

	total := ec.metrics.hits + ec.metrics.misses + ec.metrics.staleHits
	if total == 0 {
		return 0
	}
	return float64(ec.metrics.hits+ec.metrics.staleHits) / float64(total) * 100
}

// recordHit records a cache hit
func (ec *EnhancedService) recordHit() {
	ec.metrics.mu.Lock()
	ec.metrics.hits++
	ec.metrics.mu.Unlock()
}

// recordMiss records a cache miss
func (ec *EnhancedService) recordMiss() {
	ec.metrics.mu.Lock()
	ec.metrics.misses++
	ec.metrics.mu.Unlock()
}

// recordStaleHit records a stale cache hit
func (ec *EnhancedService) recordStaleHit() {
	ec.metrics.mu.Lock()
	ec.metrics.staleHits++
	ec.metrics.mu.Unlock()
}

// recordRefreshSuccess records a successful refresh
func (ec *EnhancedService) recordRefreshSuccess(duration time.Duration) {
	ec.metrics.mu.Lock()
	ec.metrics.refreshes++
	ec.metrics.refreshSuccesses++
	ec.metrics.backgroundRefreshes++
	ec.metrics.lastRefreshTime = time.Now()
	ec.metrics.totalRefreshDuration += duration
	ec.metrics.avgRefreshDuration = ec.metrics.totalRefreshDuration / time.Duration(ec.metrics.refreshSuccesses)
	ec.metrics.mu.Unlock()
}

// recordRefreshFailure records a failed refresh
func (ec *EnhancedService) recordRefreshFailure() {
	ec.metrics.mu.Lock()
	ec.metrics.refreshes++
	ec.metrics.refreshFailures++
	ec.metrics.mu.Unlock()
}
