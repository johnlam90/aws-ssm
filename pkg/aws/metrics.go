package aws

import (
	"sync"
	"time"

	"github.com/johnlam90/aws-ssm/pkg/logging"
)

// PerformanceMetrics tracks AWS API call performance metrics
type PerformanceMetrics struct {
	mu               sync.RWMutex
	apiCalls         map[string]*APIMetrics
	totalCalls       int64
	totalSuccesses   int64
	totalFailures    int64
	totalRetries     int64
	totalDuration    time.Duration
	cacheHits        int64
	cacheMisses      int64
	memoryUsageBytes int64
	lastResetTime    time.Time
	enabled          bool
	logger           logging.Logger
}

// APIMetrics tracks metrics for a specific API operation
type APIMetrics struct {
	mu            sync.RWMutex
	calls         int64
	successes     int64
	failures      int64
	retries       int64
	totalDuration time.Duration
	avgDuration   time.Duration
	minDuration   time.Duration
	maxDuration   time.Duration
	lastCallTime  time.Time
	lastError     string
	circuitOpen   bool
}

// NewPerformanceMetrics creates a new performance metrics tracker
func NewPerformanceMetrics(enabled bool) *PerformanceMetrics {
	return &PerformanceMetrics{
		apiCalls:      make(map[string]*APIMetrics),
		lastResetTime: time.Now(),
		enabled:       enabled,
		logger:        logging.With(logging.String("component", "metrics")),
	}
}

// RecordAPICall records metrics for an API call
func (pm *PerformanceMetrics) RecordAPICall(operation string, duration time.Duration, err error, retryCount int) {
	if !pm.enabled {
		return
	}

	pm.mu.Lock()
	defer pm.mu.Unlock()

	// Get or create API metrics
	apiMetrics, exists := pm.apiCalls[operation]
	if !exists {
		apiMetrics = &APIMetrics{
			minDuration: duration,
			maxDuration: duration,
		}
		pm.apiCalls[operation] = apiMetrics
	}

	// Update API-specific metrics
	apiMetrics.mu.Lock()
	apiMetrics.calls++
	apiMetrics.totalDuration += duration
	apiMetrics.avgDuration = apiMetrics.totalDuration / time.Duration(apiMetrics.calls)
	apiMetrics.lastCallTime = time.Now()

	if duration < apiMetrics.minDuration {
		apiMetrics.minDuration = duration
	}
	if duration > apiMetrics.maxDuration {
		apiMetrics.maxDuration = duration
	}

	if err != nil {
		apiMetrics.failures++
		apiMetrics.lastError = err.Error()
	} else {
		apiMetrics.successes++
		apiMetrics.lastError = ""
	}

	if retryCount > 0 {
		apiMetrics.retries += int64(retryCount)
	}
	apiMetrics.mu.Unlock()

	// Update global metrics
	pm.totalCalls++
	pm.totalDuration += duration
	if err != nil {
		pm.totalFailures++
	} else {
		pm.totalSuccesses++
	}
	if retryCount > 0 {
		pm.totalRetries += int64(retryCount)
	}
}

// RecordCacheHit records a cache hit
func (pm *PerformanceMetrics) RecordCacheHit() {
	if !pm.enabled {
		return
	}

	pm.mu.Lock()
	pm.cacheHits++
	pm.mu.Unlock()
}

// RecordCacheMiss records a cache miss
func (pm *PerformanceMetrics) RecordCacheMiss() {
	if !pm.enabled {
		return
	}

	pm.mu.Lock()
	pm.cacheMisses++
	pm.mu.Unlock()
}

// RecordMemoryUsage records current memory usage
func (pm *PerformanceMetrics) RecordMemoryUsage(bytes int64) {
	if !pm.enabled {
		return
	}

	pm.mu.Lock()
	pm.memoryUsageBytes = bytes
	pm.mu.Unlock()
}

// GetSummary returns a summary of all metrics
func (pm *PerformanceMetrics) GetSummary() MetricsSummary {
	if !pm.enabled {
		return MetricsSummary{}
	}

	pm.mu.RLock()
	defer pm.mu.RUnlock()

	summary := MetricsSummary{
		TotalCalls:       pm.totalCalls,
		TotalSuccesses:   pm.totalSuccesses,
		TotalFailures:    pm.totalFailures,
		TotalRetries:     pm.totalRetries,
		TotalDuration:    pm.totalDuration,
		CacheHits:        pm.cacheHits,
		CacheMisses:      pm.cacheMisses,
		MemoryUsageBytes: pm.memoryUsageBytes,
		LastResetTime:    pm.lastResetTime,
		APIMetrics:       make(map[string]APIMetricsSummary),
	}

	if pm.totalCalls > 0 {
		summary.AvgDuration = pm.totalDuration / time.Duration(pm.totalCalls)
		summary.SuccessRate = float64(pm.totalSuccesses) / float64(pm.totalCalls) * 100
	}

	totalCacheAccess := pm.cacheHits + pm.cacheMisses
	if totalCacheAccess > 0 {
		summary.CacheHitRate = float64(pm.cacheHits) / float64(totalCacheAccess) * 100
	}

	// Copy API metrics
	for op, metrics := range pm.apiCalls {
		metrics.mu.RLock()
		summary.APIMetrics[op] = APIMetricsSummary{
			Calls:        metrics.calls,
			Successes:    metrics.successes,
			Failures:     metrics.failures,
			Retries:      metrics.retries,
			AvgDuration:  metrics.avgDuration,
			MinDuration:  metrics.minDuration,
			MaxDuration:  metrics.maxDuration,
			LastCallTime: metrics.lastCallTime,
			LastError:    metrics.lastError,
			CircuitOpen:  metrics.circuitOpen,
		}
		metrics.mu.RUnlock()
	}

	return summary
}

// GetAPIMetrics returns metrics for a specific API operation
func (pm *PerformanceMetrics) GetAPIMetrics(operation string) (APIMetricsSummary, bool) {
	if !pm.enabled {
		return APIMetricsSummary{}, false
	}

	pm.mu.RLock()
	metrics, exists := pm.apiCalls[operation]
	pm.mu.RUnlock()

	if !exists {
		return APIMetricsSummary{}, false
	}

	metrics.mu.RLock()
	defer metrics.mu.RUnlock()

	return APIMetricsSummary{
		Calls:        metrics.calls,
		Successes:    metrics.successes,
		Failures:     metrics.failures,
		Retries:      metrics.retries,
		AvgDuration:  metrics.avgDuration,
		MinDuration:  metrics.minDuration,
		MaxDuration:  metrics.maxDuration,
		LastCallTime: metrics.lastCallTime,
		LastError:    metrics.lastError,
		CircuitOpen:  metrics.circuitOpen,
	}, true
}

// Reset resets all metrics
func (pm *PerformanceMetrics) Reset() {
	if !pm.enabled {
		return
	}

	pm.mu.Lock()
	defer pm.mu.Unlock()

	pm.apiCalls = make(map[string]*APIMetrics)
	pm.totalCalls = 0
	pm.totalSuccesses = 0
	pm.totalFailures = 0
	pm.totalRetries = 0
	pm.totalDuration = 0
	pm.cacheHits = 0
	pm.cacheMisses = 0
	pm.memoryUsageBytes = 0
	pm.lastResetTime = time.Now()

	pm.logger.Info("Performance metrics reset")
}

// LogSummary logs a summary of all metrics
func (pm *PerformanceMetrics) LogSummary() {
	if !pm.enabled {
		return
	}

	summary := pm.GetSummary()

	pm.logger.Info("Performance Metrics Summary",
		logging.Int64("total_calls", summary.TotalCalls),
		logging.Int64("successes", summary.TotalSuccesses),
		logging.Int64("failures", summary.TotalFailures),
		logging.Float64("success_rate", summary.SuccessRate),
		logging.Duration("avg_duration", summary.AvgDuration),
		logging.Int64("cache_hits", summary.CacheHits),
		logging.Int64("cache_misses", summary.CacheMisses),
		logging.Float64("cache_hit_rate", summary.CacheHitRate),
		logging.Int64("memory_usage_mb", summary.MemoryUsageBytes/(1024*1024)))

	// Log top API operations by call count
	pm.logger.Debug("API Operations Summary")
	for op, metrics := range summary.APIMetrics {
		pm.logger.Debug("API Operation",
			logging.String("operation", op),
			logging.Int64("calls", metrics.Calls),
			logging.Int64("successes", metrics.Successes),
			logging.Int64("failures", metrics.Failures),
			logging.Duration("avg_duration", metrics.AvgDuration),
			logging.Duration("min_duration", metrics.MinDuration),
			logging.Duration("max_duration", metrics.MaxDuration))
	}
}

// MetricsSummary contains a summary of all metrics
type MetricsSummary struct {
	TotalCalls       int64
	TotalSuccesses   int64
	TotalFailures    int64
	TotalRetries     int64
	TotalDuration    time.Duration
	AvgDuration      time.Duration
	SuccessRate      float64
	CacheHits        int64
	CacheMisses      int64
	CacheHitRate     float64
	MemoryUsageBytes int64
	LastResetTime    time.Time
	APIMetrics       map[string]APIMetricsSummary
}

// APIMetricsSummary contains metrics for a specific API operation
type APIMetricsSummary struct {
	Calls        int64
	Successes    int64
	Failures     int64
	Retries      int64
	AvgDuration  time.Duration
	MinDuration  time.Duration
	MaxDuration  time.Duration
	LastCallTime time.Time
	LastError    string
	CircuitOpen  bool
}

// MetricsCollector periodically collects and logs metrics
type MetricsCollector struct {
	metrics  *PerformanceMetrics
	interval time.Duration
	stopChan chan struct{}
	logger   logging.Logger
}

// NewMetricsCollector creates a new metrics collector
func NewMetricsCollector(metrics *PerformanceMetrics, interval time.Duration) *MetricsCollector {
	return &MetricsCollector{
		metrics:  metrics,
		interval: interval,
		stopChan: make(chan struct{}),
		logger:   logging.With(logging.String("component", "metrics_collector")),
	}
}

// Start starts the metrics collection loop
func (mc *MetricsCollector) Start() {
	mc.logger.Info("Starting metrics collector", logging.Duration("interval", mc.interval))

	go func() {
		ticker := time.NewTicker(mc.interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				mc.metrics.LogSummary()
			case <-mc.stopChan:
				mc.logger.Info("Stopping metrics collector")
				return
			}
		}
	}()
}

// Stop stops the metrics collection loop
func (mc *MetricsCollector) Stop() {
	close(mc.stopChan)
}

// InstrumentedClient wraps a Client with performance metrics
type InstrumentedClient struct {
	*Client
	metrics *PerformanceMetrics
}

// NewInstrumentedClient creates a new instrumented client
func NewInstrumentedClient(client *Client, enableMetrics bool) *InstrumentedClient {
	return &InstrumentedClient{
		Client:  client,
		metrics: NewPerformanceMetrics(enableMetrics),
	}
}

// GetMetrics returns the performance metrics
func (ic *InstrumentedClient) GetMetrics() *PerformanceMetrics {
	return ic.metrics
}
