package metrics

import (
	"context"
	"sync"
	"time"

	"github.com/aws-ssm/pkg/logging"
)

// MetricType represents different types of metrics
type MetricType string

const (
	// MetricCounter represents a cumulative counter metric type
	MetricCounter MetricType = "counter"
	// MetricGauge represents a gauge metric type
	MetricGauge MetricType = "gauge"
	// MetricHistogram represents a histogram metric type
	MetricHistogram MetricType = "histogram"
	// MetricTimer represents a timer metric type
	MetricTimer MetricType = "timer"
)

// Metric represents a single metric
type Metric struct {
	Name        string
	Type        MetricType
	Value       float64
	Labels      map[string]string
	Timestamp   time.Time
	Description string
}

// Counter represents a cumulative counter
type Counter struct {
	name   string
	labels map[string]string
	value  float64
	mu     sync.RWMutex
	logger logging.Logger
}

// NewCounter creates a new counter
func NewCounter(name string, labels map[string]string) *Counter {
	return &Counter{
		name:   name,
		labels: labels,
		logger: logging.With(logging.String("component", "metrics"), logging.String("metric_type", "counter"), logging.String("name", name)),
	}
}

// Inc increments the counter
func (c *Counter) Inc(value float64) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.value += value
	c.logger.Debug("Counter incremented", logging.String("name", c.name), logging.Float64("value", value), logging.Float64("total", c.value))
}

// Add increments the counter by a value
func (c *Counter) Add(value float64) {
	c.Inc(value)
}

// Set sets the counter to a specific value
func (c *Counter) Set(value float64) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.value = value
	c.logger.Debug("Counter set", logging.String("name", c.name), logging.Float64("value", value))
}

// GetValue returns the current counter value
func (c *Counter) GetValue() float64 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.value
}

// ToMetric converts counter to metric
func (c *Counter) ToMetric() *Metric {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return &Metric{
		Name:        c.name,
		Type:        MetricCounter,
		Value:       c.value,
		Labels:      c.labels,
		Timestamp:   time.Now(),
		Description: "Counter metric",
	}
}

// Gauge represents a value that can go up and down
type Gauge struct {
	name   string
	labels map[string]string
	value  float64
	mu     sync.RWMutex
	logger logging.Logger
}

// NewGauge creates a new gauge
func NewGauge(name string, labels map[string]string) *Gauge {
	return &Gauge{
		name:   name,
		labels: labels,
		logger: logging.With(logging.String("component", "metrics"), logging.String("metric_type", "gauge"), logging.String("name", name)),
	}
}

// Inc increments the gauge
func (g *Gauge) Inc(value float64) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.value += value
	g.logger.Debug("Gauge incremented", logging.String("name", g.name), logging.Float64("value", value), logging.Float64("current", g.value))
}

// Dec decrements the gauge
func (g *Gauge) Dec(value float64) {
	g.Inc(-value)
}

// Set sets the gauge to a specific value
func (g *Gauge) Set(value float64) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.value = value
	g.logger.Debug("Gauge set", logging.String("name", g.name), logging.Float64("value", value))
}

// GetValue returns the current gauge value
func (g *Gauge) GetValue() float64 {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.value
}

// ToMetric converts gauge to metric
func (g *Gauge) ToMetric() *Metric {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return &Metric{
		Name:        g.name,
		Type:        MetricGauge,
		Value:       g.value,
		Labels:      g.labels,
		Timestamp:   time.Now(),
		Description: "Gauge metric",
	}
}

// Histogram samples observations
type Histogram struct {
	name    string
	labels  map[string]string
	sum     float64
	count   uint64
	buckets map[float64]uint64
	mu      sync.RWMutex
	logger  logging.Logger
}

// HistogramConfig represents histogram configuration
type HistogramConfig struct {
	Buckets []float64
}

// DefaultHistogramConfig returns default histogram configuration
func DefaultHistogramConfig() *HistogramConfig {
	return &HistogramConfig{
		Buckets: []float64{0.005, 0.01, 0.025, 0.05, 0.075, 0.1, 0.25, 0.5, 0.75, 1.0, 2.5, 5.0, 7.5, 10.0, 1e9},
	}
}

// NewHistogram creates a new histogram
func NewHistogram(name string, labels map[string]string, config *HistogramConfig) *Histogram {
	return &Histogram{
		name:    name,
		labels:  labels,
		buckets: make(map[float64]uint64),
		logger:  logging.With(logging.String("component", "metrics"), logging.String("metric_type", "histogram"), logging.String("name", name)),
	}
}

// Observe adds an observation to the histogram
func (h *Histogram) Observe(value float64) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.sum += value
	h.count++

	// Find the appropriate bucket
	for _, bucket := range DefaultHistogramConfig().Buckets {
		if value <= bucket {
			h.buckets[bucket]++
		}
	}

	h.logger.Debug("Histogram observation", logging.String("name", h.name), logging.Float64("value", value), logging.Int64("count", int64(h.count)))
}

// GetCount returns the number of observations
func (h *Histogram) GetCount() uint64 {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.count
}

// GetSum returns the sum of all observations
func (h *Histogram) GetSum() float64 {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.sum
}

// GetBuckets returns bucket counts
func (h *Histogram) GetBuckets() map[float64]uint64 {
	h.mu.RLock()
	defer h.mu.RUnlock()

	result := make(map[float64]uint64)
	for k, v := range h.buckets {
		result[k] = v
	}
	return result
}

// ToMetric converts histogram to metric
func (h *Histogram) ToMetric() *Metric {
	h.mu.RLock()
	defer h.mu.RUnlock()

	return &Metric{
		Name:        h.name,
		Type:        MetricHistogram,
		Value:       h.sum,
		Labels:      h.labels,
		Timestamp:   time.Now(),
		Description: "Histogram metric",
	}
}

// Timer measures duration
type Timer struct {
	name   string
	labels map[string]string
	logger logging.Logger
}

// NewTimer creates a new timer
func NewTimer(name string, labels map[string]string) *Timer {
	return &Timer{
		name:   name,
		labels: labels,
		logger: logging.With(logging.String("component", "metrics"), logging.String("metric_type", "timer"), logging.String("name", name)),
	}
}

// Record records a duration
func (t *Timer) Record(duration time.Duration) {
	t.logger.Debug("Timer recorded", logging.String("name", t.name), logging.Duration("duration", duration))
}

// Start starts a timer and returns a TimerContext
func (t *Timer) Start() *TimerContext {
	return &TimerContext{
		timer:  t,
		start:  time.Now(),
		labels: make(map[string]string),
	}
}

// TimerContext manages a timing operation
type TimerContext struct {
	timer  *Timer
	start  time.Time
	labels map[string]string
}

// WithLabel adds a label to the timer context
func (tc *TimerContext) WithLabel(key, value string) *TimerContext {
	tc.labels[key] = value
	return tc
}

// Stop stops the timer and records the duration
func (tc *TimerContext) Stop() {
	duration := time.Since(tc.start)
	tc.timer.Record(duration)
}

// ToMetric converts timer to metric
func (t *Timer) ToMetric() *Metric {
	return &Metric{
		Name:        t.name,
		Type:        MetricTimer,
		Value:       0,
		Labels:      t.labels,
		Timestamp:   time.Now(),
		Description: "Timer metric",
	}
}

// Registry manages all metrics
type Registry struct {
	metrics map[string]MetricCollector
	mu      sync.RWMutex
	logger  logging.Logger
}

// MetricCollector represents a metric that can be collected
type MetricCollector interface {
	ToMetric() *Metric
}

// NewRegistry creates a new registry
func NewRegistry() *Registry {
	return &Registry{
		metrics: make(map[string]MetricCollector),
		logger:  logging.With(logging.String("component", "metrics_registry")),
	}
}

// Register registers a metric
func (r *Registry) Register(name string, metric MetricCollector) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.metrics[name] = metric
	r.logger.Info("Metric registered", logging.String("name", name))
}

// Unregister removes a metric
func (r *Registry) Unregister(name string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.metrics, name)
	r.logger.Info("Metric unregistered", logging.String("name", name))
}

// GetMetric retrieves a metric by name
func (r *Registry) GetMetric(name string) (MetricCollector, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	metric, exists := r.metrics[name]
	return metric, exists
}

// CollectAll collects all metrics
func (r *Registry) CollectAll() []*Metric {
	r.mu.RLock()
	defer r.mu.RUnlock()

	metrics := make([]*Metric, 0, len(r.metrics))
	for _, collector := range r.metrics {
		metrics = append(metrics, collector.ToMetric())
	}

	return metrics
}

// GetCount returns the number of registered metrics
func (r *Registry) GetCount() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.metrics)
}

// Global metrics registry
var globalRegistry = NewRegistry()

// GlobalRegistry returns the global registry
func GlobalRegistry() *Registry {
	return globalRegistry
}

// Common metrics
var (
	// AWS API metrics
	AWSSSMRequestsTotal   = NewCounter("aws_ssm_requests_total", nil)
	AWSSSMRequestDuration = NewHistogram("aws_ssm_request_duration_seconds", nil, nil)
	AWSSSMRequestErrors   = NewCounter("aws_ssm_request_errors_total", nil)

	// Session metrics
	SessionStartTotal = NewCounter("session_start_total", map[string]string{"status": "unknown"})
	SessionDuration   = NewHistogram("session_duration_seconds", nil, nil)
	SessionActive     = NewGauge("session_active_total", nil)

	// Cache metrics
	CacheHits   = NewCounter("cache_hits_total", nil)
	CacheMisses = NewCounter("cache_misses_total", nil)
	CacheSize   = NewGauge("cache_size_bytes", nil)

	// Error metrics
	ErrorsTotal = NewCounter("errors_total", map[string]string{"type": "unknown"})

	// Performance metrics
	CommandExecutionTime = NewHistogram("command_execution_time_seconds", nil, nil)
	InstanceSearchTime   = NewHistogram("instance_search_time_seconds", nil, nil)

	// Health metrics
	HealthCheckDuration = NewHistogram("health_check_duration_seconds", map[string]string{"check": "unknown"}, nil)
	HealthCheckStatus   = NewGauge("health_check_status", map[string]string{"check": "unknown"})
)

// Service provides metrics collection and reporting
type Service struct {
	registry  *Registry
	ctx       context.Context
	mu        sync.RWMutex
	logger    logging.Logger
	reporters []Reporter
	stopCh    chan struct{}
}

// Reporter interface for metrics reporting
type Reporter interface {
	Report(ctx context.Context, metrics []*Metric) error
	Name() string
}

// NewService creates a new metrics service
func NewService(ctx context.Context) *Service {
	service := &Service{
		registry:  NewRegistry(),
		ctx:       ctx,
		logger:    logging.With(logging.String("component", "metrics_service")),
		reporters: make([]Reporter, 0),
		stopCh:    make(chan struct{}),
	}

	// Register common metrics
	service.registry.Register("aws_ssm_requests_total", AWSSSMRequestsTotal)
	service.registry.Register("aws_ssm_request_duration_seconds", AWSSSMRequestDuration)
	service.registry.Register("aws_ssm_request_errors_total", AWSSSMRequestErrors)
	service.registry.Register("session_start_total", SessionStartTotal)
	service.registry.Register("session_duration_seconds", SessionDuration)
	service.registry.Register("session_active_total", SessionActive)
	service.registry.Register("cache_hits_total", CacheHits)
	service.registry.Register("cache_misses_total", CacheMisses)
	service.registry.Register("cache_size_bytes", CacheSize)
	service.registry.Register("errors_total", ErrorsTotal)
	service.registry.Register("command_execution_time_seconds", CommandExecutionTime)
	service.registry.Register("instance_search_time_seconds", InstanceSearchTime)
	service.registry.Register("health_check_duration_seconds", HealthCheckDuration)
	service.registry.Register("health_check_status", HealthCheckStatus)

	return service
}

// AddReporter adds a metrics reporter
func (s *Service) AddReporter(reporter Reporter) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.reporters = append(s.reporters, reporter)
	s.logger.Info("Metrics reporter added", logging.String("name", reporter.Name()))
}

// Start starts the metrics service
func (s *Service) Start() {
	go s.reportLoop()
	s.logger.Info("Metrics service started")
}

// Stop stops the metrics service
func (s *Service) Stop() {
	close(s.stopCh)
	s.logger.Info("Metrics service stopped")
}

func (s *Service) reportLoop() {
	ticker := time.NewTicker(30 * time.Second) // Report every 30 seconds
	defer ticker.Stop()

	for {
		select {
		case <-s.ctx.Done():
			return
		case <-s.stopCh:
			return
		case <-ticker.C:
			s.report()
		}
	}
}

func (s *Service) report() {
	metrics := s.registry.CollectAll()

	s.mu.RLock()
	reporters := make([]Reporter, len(s.reporters))
	copy(reporters, s.reporters)
	s.mu.RUnlock()

	for _, reporter := range reporters {
		if err := reporter.Report(s.ctx, metrics); err != nil {
			s.logger.Error("Metrics reporting failed",
				logging.String("reporter", reporter.Name()),
				logging.String("error", err.Error()))
		}
	}
}

// ConsoleReporter reports metrics to console
type ConsoleReporter struct {
	logger logging.Logger
}

// NewConsoleReporter creates a new console reporter for metrics output
func NewConsoleReporter() *ConsoleReporter {
	return &ConsoleReporter{
		logger: logging.With(logging.String("component", "console_reporter")),
	}
}

// Report outputs metrics to the console
func (r *ConsoleReporter) Report(ctx context.Context, metrics []*Metric) error {
	// Log metrics summary
	r.logger.Info("Metrics report",
		logging.Int("metric_count", len(metrics)),
		logging.String("timestamp", time.Now().Format(time.RFC3339)))

	// In a real implementation, you would format and output metrics
	// For now, just log them
	for _, metric := range metrics {
		r.logger.Debug("Metric",
			logging.String("name", metric.Name),
			logging.String("type", string(metric.Type)),
			logging.Float64("value", metric.Value))
	}

	return nil
}

// Name returns the name of the console reporter
func (r *ConsoleReporter) Name() string {
	return "console"
}

// PrometheusReporter would implement Prometheus format reporting
// This is a placeholder for a real implementation
type PrometheusReporter struct {
	logger logging.Logger
}

// NewPrometheusReporter creates a new Prometheus reporter for metrics
func NewPrometheusReporter() *PrometheusReporter {
	return &PrometheusReporter{
		logger: logging.With(logging.String("component", "prometheus_reporter")),
	}
}

// Report formats and outputs metrics in Prometheus format
func (r *PrometheusReporter) Report(ctx context.Context, metrics []*Metric) error {
	// In a real implementation, you would format metrics in Prometheus format
	// and expose them via HTTP endpoint or write to a file
	r.logger.Debug("Prometheus metrics reported", logging.Int("metric_count", len(metrics)))
	return nil
}

// Name returns the name of the Prometheus reporter
func (r *PrometheusReporter) Name() string {
	return "prometheus"
}

// Inf returns infinity value for metric usage
func Inf() float64 {
	return 1 << 53 // Safe integer precision for floating point infinity representation
}

// ExportToJSON exports metrics to JSON format
func ExportToJSON(metrics []*Metric) (string, error) {
	// This would implement JSON export
	// For now, return empty string
	return "", nil
}

// ResetAllMetrics resets all registered metrics
func ResetAllMetrics() {
	metrics := globalRegistry.CollectAll()
	for _, metric := range metrics {
		// Reset counters and gauges
		switch metric.Type {
		case MetricCounter:
			// Would reset counter to 0
		case MetricGauge:
			// Would reset gauge to 0
		case MetricHistogram:
			// Would reset histogram
		case MetricTimer:
			// Would reset timer
		}
	}
}

// Labels creates a map of key-value pairs for metric labels
func Labels(pairs ...string) map[string]string {
	if len(pairs)%2 != 0 {
		panic("labels must have even number of arguments")
	}

	labels := make(map[string]string, len(pairs)/2)
	for i := 0; i < len(pairs); i += 2 {
		labels[pairs[i]] = pairs[i+1]
	}
	return labels
}

// PerformanceMonitor provides high-level performance monitoring
type PerformanceMonitor struct {
	timers map[string]*Timer
	mu     sync.RWMutex
	logger logging.Logger
}

// NewPerformanceMonitor creates a new performance monitor
func NewPerformanceMonitor() *PerformanceMonitor {
	return &PerformanceMonitor{
		timers: make(map[string]*Timer),
		logger: logging.With(logging.String("component", "performance_monitor")),
	}
}

// StartTimer starts a timer with the given name and returns a timer context
func (pm *PerformanceMonitor) StartTimer(name string) *TimerContext {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	timer, exists := pm.timers[name]
	if !exists {
		timer = NewTimer(name, nil)
		pm.timers[name] = timer
	}

	return timer.Start()
}

// RecordOperation records the duration of an operation
func (pm *PerformanceMonitor) RecordOperation(operation string, duration time.Duration) {
	pm.logger.Info("Operation performance",
		logging.String("operation", operation),
		logging.Duration("duration", duration))

	// Update relevant metrics
	CommandExecutionTime.Observe(duration.Seconds())
}

// StartOperation starts timing an operation and returns a timer context
func (pm *PerformanceMonitor) StartOperation(operation string) *TimerContext {
	return pm.StartTimer("operation_" + operation)
}

// Start the global metrics service
func init() {
	ctx := context.Background()
	service := NewService(ctx)
	GlobalRegistry() // Initialize global registry
	GlobalMetricsService = service
	service.AddReporter(NewConsoleReporter())
	service.Start()
}

// GlobalMetricsService is the global metrics service instance
var GlobalMetricsService *Service

// GetGlobalMetricsService returns the global metrics service
func GetGlobalMetricsService() *Service {
	return GlobalMetricsService
}
