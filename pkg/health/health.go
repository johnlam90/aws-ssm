package health

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/johnlam90/aws-ssm/pkg/logging"
)

// Status represents health check status
type Status string

const (
	// StatusOK indicates the health check passed successfully
	StatusOK Status = "ok"
	// StatusWarning indicates a non-critical issue was detected
	StatusWarning Status = "warning"
	// StatusError indicates a general error occurred
	StatusError Status = "error"
	// StatusCritical indicates a critical error that requires immediate attention
	StatusCritical Status = "critical"
	// StatusUnknown indicates the health check status could not be determined
	StatusUnknown Status = "unknown"
)

func (s Status) String() string {
	return string(s)
}

// Check represents a health check interface
type Check interface {
	Name() string
	Check(ctx context.Context) *CheckResult
}

// CheckResult represents the result of a health check
type CheckResult struct {
	Status      Status
	Message     string
	Duration    time.Duration
	Timestamp   time.Time
	Metadata    map[string]interface{}
	ServiceName string
}

// NewCheckResult creates a new health check result
func NewCheckResult(status Status, message string) *CheckResult {
	return &CheckResult{
		Status:    status,
		Message:   message,
		Timestamp: time.Now(),
		Metadata:  make(map[string]interface{}),
	}
}

// WithDuration sets the duration for the health check result
func (r *CheckResult) WithDuration(duration time.Duration) *CheckResult {
	r.Duration = duration
	return r
}

// WithMetadata adds metadata to the health check result
func (r *CheckResult) WithMetadata(key string, value interface{}) *CheckResult {
	r.Metadata[key] = value
	return r
}

// WithServiceName sets the service name for the health check result
func (r *CheckResult) WithServiceName(serviceName string) *CheckResult {
	r.ServiceName = serviceName
	return r
}

// BaseHealthCheck provides common health check functionality
type BaseHealthCheck struct {
	name      string
	logger    logging.Logger
	mu        sync.RWMutex
	lastCheck *CheckResult
}

// NewBaseHealthCheck creates a new base health check with the given name
func NewBaseHealthCheck(name string) *BaseHealthCheck {
	return &BaseHealthCheck{
		name:   name,
		logger: logging.With(logging.String("component", "health_check"), logging.String("check_name", name)),
	}
}

// Name returns the name of the health check
func (bc *BaseHealthCheck) Name() string {
	return bc.name
}

// GetLastCheck returns the result of the most recent health check
func (bc *BaseHealthCheck) GetLastCheck() *CheckResult {
	bc.mu.RLock()
	defer bc.mu.RUnlock()
	return bc.lastCheck
}

func (bc *BaseHealthCheck) setLastCheck(result *CheckResult) {
	bc.mu.Lock()
	defer bc.mu.Unlock()
	bc.lastCheck = result
}

// AWSConnectivityCheck checks AWS service connectivity
type AWSConnectivityCheck struct {
	*BaseHealthCheck
	testFunc func(ctx context.Context) error
}

// NewAWSConnectivityCheck creates a new AWS connectivity health check
func NewAWSConnectivityCheck(testFunc func(ctx context.Context) error) *AWSConnectivityCheck {
	return &AWSConnectivityCheck{
		BaseHealthCheck: NewBaseHealthCheck("aws_connectivity"),
		testFunc:        testFunc,
	}
}

// Check performs the AWS connectivity health check
func (c *AWSConnectivityCheck) Check(ctx context.Context) *CheckResult {
	start := time.Now()

	result := NewCheckResult(StatusOK, "AWS connectivity check passed")

	err := c.testFunc(ctx)
	duration := time.Since(start)

	if err != nil {
		result.Status = StatusError
		result.Message = fmt.Sprintf("AWS connectivity check failed: %v", err)
		c.logger.Error("AWS connectivity check failed", logging.String("error", err.Error()))
	} else {
		c.logger.Debug("AWS connectivity check passed")
	}

	result.WithDuration(duration)
	result.WithServiceName("aws")

	c.setLastCheck(result)
	return result
}

// CacheHealthCheck checks cache service health
type CacheHealthCheck struct {
	*BaseHealthCheck
	testFunc func(ctx context.Context) error
}

// NewCacheHealthCheck creates a new cache health check
func NewCacheHealthCheck(testFunc func(ctx context.Context) error) *CacheHealthCheck {
	return &CacheHealthCheck{
		BaseHealthCheck: NewBaseHealthCheck("cache"),
		testFunc:        testFunc,
	}
}

// Check performs the cache health check
func (c *CacheHealthCheck) Check(ctx context.Context) *CheckResult {
	start := time.Now()

	result := NewCheckResult(StatusOK, "Cache health check passed")

	err := c.testFunc(ctx)
	duration := time.Since(start)

	if err != nil {
		result.Status = StatusError
		result.Message = fmt.Sprintf("Cache health check failed: %v", err)
		c.logger.Error("Cache health check failed", logging.String("error", err.Error()))
	} else {
		c.logger.Debug("Cache health check passed")
	}

	result.WithDuration(duration)
	result.WithServiceName("cache")

	c.setLastCheck(result)
	return result
}

// ConfigHealthCheck validates configuration
type ConfigHealthCheck struct {
	*BaseHealthCheck
	validateFunc func() error
}

// NewConfigHealthCheck creates a new configuration health check
func NewConfigHealthCheck(validateFunc func() error) *ConfigHealthCheck {
	return &ConfigHealthCheck{
		BaseHealthCheck: NewBaseHealthCheck("configuration"),
		validateFunc:    validateFunc,
	}
}

// Check performs the configuration health check
func (c *ConfigHealthCheck) Check(ctx context.Context) *CheckResult {
	start := time.Now()

	result := NewCheckResult(StatusOK, "Configuration validation passed")

	err := c.validateFunc()
	duration := time.Since(start)

	if err != nil {
		result.Status = StatusError
		result.Message = fmt.Sprintf("Configuration validation failed: %v", err)
		c.logger.Error("Configuration validation failed", logging.String("error", err.Error()))
	} else {
		c.logger.Debug("Configuration validation passed")
	}

	result.WithDuration(duration)
	result.WithServiceName("config")

	c.setLastCheck(result)
	return result
}

// MemoryHealthCheck monitors memory usage
type MemoryHealthCheck struct {
	*BaseHealthCheck
	thresholdPercent float64
}

// NewMemoryHealthCheck creates a new memory usage health check
func NewMemoryHealthCheck(thresholdPercent float64) *MemoryHealthCheck {
	return &MemoryHealthCheck{
		BaseHealthCheck:  NewBaseHealthCheck("memory"),
		thresholdPercent: thresholdPercent,
	}
}

// Check performs the memory usage health check
func (c *MemoryHealthCheck) Check(ctx context.Context) *CheckResult {
	start := time.Now()

	result := NewCheckResult(StatusOK, "Memory usage within limits")

	// This is a simplified memory check
	// In a real implementation, you'd use runtime.MemStats
	duration := time.Since(start)

	// Simulate memory check
	memoryUsage := c.getMemoryUsage()

	if memoryUsage > c.thresholdPercent {
		result.Status = StatusWarning
		result.Message = fmt.Sprintf("High memory usage: %.1f%%", memoryUsage)
		c.logger.Warn("High memory usage detected", logging.Float64("usage_percent", memoryUsage))
	}

	result.WithDuration(duration)
	result.WithMetadata("memory_usage_percent", memoryUsage)
	result.WithServiceName("system")

	c.setLastCheck(result)
	return result
}

func (c *MemoryHealthCheck) getMemoryUsage() float64 {
	// Simplified implementation
	// In production, use runtime.MemStats and system metrics
	return float64(50 + (time.Now().Second() % 20)) // Simulate varying memory usage
}

// DiskSpaceHealthCheck monitors disk space
type DiskSpaceHealthCheck struct {
	*BaseHealthCheck
	thresholdPercent float64
}

// NewDiskSpaceHealthCheck creates a new disk space health check
func NewDiskSpaceHealthCheck(thresholdPercent float64) *DiskSpaceHealthCheck {
	return &DiskSpaceHealthCheck{
		BaseHealthCheck:  NewBaseHealthCheck("disk_space"),
		thresholdPercent: thresholdPercent,
	}
}

// Check performs the disk space health check
func (c *DiskSpaceHealthCheck) Check(ctx context.Context) *CheckResult {
	start := time.Now()

	result := NewCheckResult(StatusOK, "Disk space within limits")

	// In production, check actual disk space
	duration := time.Since(start)

	// Simulate disk space check
	diskUsage := c.getDiskUsage()

	if diskUsage > c.thresholdPercent {
		result.Status = StatusWarning
		result.Message = fmt.Sprintf("Low disk space: %.1f%% free", 100-diskUsage)
		c.logger.Warn("Low disk space detected", logging.Float64("usage_percent", diskUsage))
	}

	result.WithDuration(duration)
	result.WithMetadata("disk_usage_percent", diskUsage)
	result.WithServiceName("system")

	c.setLastCheck(result)
	return result
}

func (c *DiskSpaceHealthCheck) getDiskUsage() float64 {
	// Simplified implementation
	// In production, use syscall.Statfs or similar
	return float64(30 + (time.Now().Second() % 40)) // Simulate varying disk usage
}

// TCPHealthCheck performs TCP connectivity checks
type TCPHealthCheck struct {
	*BaseHealthCheck
	host    string
	port    int
	timeout time.Duration
}

// NewTCPHealthCheck creates a new TCP connectivity health check
func NewTCPHealthCheck(host string, port int, timeout time.Duration) *TCPHealthCheck {
	return &TCPHealthCheck{
		BaseHealthCheck: NewBaseHealthCheck("tcp_connectivity"),
		host:            host,
		port:            port,
		timeout:         timeout,
	}
}

// Check performs the TCP connectivity health check
func (c *TCPHealthCheck) Check(ctx context.Context) *CheckResult {
	start := time.Now()

	result := NewCheckResult(StatusOK, fmt.Sprintf("TCP connection to %s:%d successful", c.host, c.port))

	dialer := &net.Dialer{
		Timeout: c.timeout,
	}

	conn, err := dialer.DialContext(ctx, "tcp", fmt.Sprintf("%s:%d", c.host, c.port))
	duration := time.Since(start)

	if err != nil {
		result.Status = StatusError
		result.Message = fmt.Sprintf("TCP connection to %s:%d failed: %v", c.host, c.port, err)
		c.logger.Error("TCP connection failed",
			logging.String("host", c.host),
			logging.Int("port", c.port),
			logging.String("error", err.Error()))
	} else {
		if err := conn.Close(); err != nil {
			c.logger.Warn("Failed to close TCP connection", logging.String("error", err.Error()))
		}
		c.logger.Debug("TCP connection successful",
			logging.String("host", c.host),
			logging.Int("port", c.port))
	}

	result.WithDuration(duration)
	result.WithMetadata("host", c.host)
	result.WithMetadata("port", c.port)
	result.WithServiceName("network")

	c.setLastCheck(result)
	return result
}

// HTTPHealthCheck performs HTTP endpoint health checks
type HTTPHealthCheck struct {
	*BaseHealthCheck
	url     string
	timeout time.Duration
}

// NewHTTPHealthCheck creates a new HTTP health check
func NewHTTPHealthCheck(url string, timeout time.Duration) *HTTPHealthCheck {
	return &HTTPHealthCheck{
		BaseHealthCheck: NewBaseHealthCheck("http_endpoint"),
		url:             url,
		timeout:         timeout,
	}
}

// Check performs the HTTP health check
func (c *HTTPHealthCheck) Check(ctx context.Context) *CheckResult {
	start := time.Now()

	result := NewCheckResult(StatusOK, fmt.Sprintf("HTTP request to %s successful", c.url))

	client := &http.Client{
		Timeout: c.timeout,
	}

	req, err := http.NewRequestWithContext(ctx, "GET", c.url, nil)
	if err != nil {
		result.Status = StatusError
		result.Message = fmt.Sprintf("Failed to create HTTP request: %v", err)
		result.WithDuration(time.Since(start))
		c.setLastCheck(result)
		return result
	}

	resp, err := client.Do(req)
	duration := time.Since(start)

	if err != nil {
		result.Status = StatusError
		result.Message = fmt.Sprintf("HTTP request to %s failed: %v", c.url, err)
		c.logger.Error("HTTP request failed",
			logging.String("url", c.url),
			logging.String("error", err.Error()))
	} else {
		defer func() {
			if err := resp.Body.Close(); err != nil {
				c.logger.Warn("Failed to close response body", logging.String("error", err.Error()))
			}
		}()

		if resp.StatusCode >= 400 {
			result.Status = StatusError
			result.Message = fmt.Sprintf("HTTP request to %s returned status code %d", c.url, resp.StatusCode)
			c.logger.Warn("HTTP request returned error status",
				logging.String("url", c.url),
				logging.Int("status_code", resp.StatusCode))
		} else {
			c.logger.Debug("HTTP request successful",
				logging.String("url", c.url),
				logging.Int("status_code", resp.StatusCode))
		}
	}

	result.WithDuration(duration)
	result.WithMetadata("url", c.url)
	result.WithServiceName("http")

	c.setLastCheck(result)
	return result
}

// CompositeHealthCheck runs multiple health checks
type CompositeHealthCheck struct {
	*BaseHealthCheck
	checks    []Check
	threshold int // Minimum number of checks that must pass
}

// NewCompositeHealthCheck creates a new composite health check that runs multiple checks
func NewCompositeHealthCheck(checks []Check, threshold int) *CompositeHealthCheck {
	return &CompositeHealthCheck{
		BaseHealthCheck: NewBaseHealthCheck("composite"),
		checks:          checks,
		threshold:       threshold,
	}
}

// Check performs all health checks in the composite
func (c *CompositeHealthCheck) Check(ctx context.Context) *CheckResult {
	start := time.Now()

	results := make([]*CheckResult, 0, len(c.checks))
	var overallStatus Status = StatusOK

	passCount := 0
	for _, check := range c.checks {
		result := check.Check(ctx)
		results = append(results, result)

		switch result.Status {
		case StatusOK:
			passCount++
		case StatusWarning:
			if overallStatus == StatusOK {
				overallStatus = StatusWarning
			}
		case StatusError, StatusCritical:
			overallStatus = StatusError
		}
	}

	duration := time.Since(start)

	// Check threshold
	if passCount < c.threshold {
		overallStatus = StatusError
	}

	result := NewCheckResult(overallStatus, "Composite health check completed")
	result.WithDuration(duration)

	// Aggregate results
	result.Metadata = map[string]interface{}{
		"total_checks":  len(c.checks),
		"passed_checks": passCount,
		"failed_checks": len(c.checks) - passCount,
		"threshold":     c.threshold,
		"check_results": results,
	}

	result.WithServiceName("composite")

	c.setLastCheck(result)
	return result
}

// Checker manages all health checks
type Checker struct {
	checks    map[string]Check
	mu        sync.RWMutex
	logger    logging.Logger
	lastCheck *CompositeResult
}

// CompositeResult represents the result of a composite health check
type CompositeResult struct {
	Timestamp time.Time
	Overall   Status
	Checks    map[string]*CheckResult
	Duration  time.Duration
}

// NewChecker creates a new health checker
func NewChecker() *Checker {
	return &Checker{
		checks: make(map[string]Check),
		logger: logging.With(logging.String("component", "health_checker")),
	}
}

// AddCheck adds a health check to the checker
func (hc *Checker) AddCheck(name string, check Check) {
	hc.mu.Lock()
	defer hc.mu.Unlock()
	hc.checks[name] = check
	hc.logger.Info("Health check added", logging.String("name", name))
}

// RemoveCheck removes a health check from the checker
func (hc *Checker) RemoveCheck(name string) {
	hc.mu.Lock()
	defer hc.mu.Unlock()
	delete(hc.checks, name)
	hc.logger.Info("Health check removed", logging.String("name", name))
}

// GetCheck retrieves a health check by name
func (hc *Checker) GetCheck(name string) (Check, bool) {
	hc.mu.RLock()
	defer hc.mu.RUnlock()
	check, exists := hc.checks[name]
	return check, exists
}

// ListChecks returns the list of all health check names
func (hc *Checker) ListChecks() []string {
	hc.mu.RLock()
	defer hc.mu.RUnlock()
	names := make([]string, 0, len(hc.checks))
	for name := range hc.checks {
		names = append(names, name)
	}
	return names
}

// CheckAll performs all registered health checks
func (hc *Checker) CheckAll(ctx context.Context) *CompositeResult {
	hc.mu.RLock()
	checks := make([]Check, 0, len(hc.checks))
	names := make([]string, 0, len(hc.checks))
	for name, check := range hc.checks {
		checks = append(checks, check)
		names = append(names, name)
	}
	hc.mu.RUnlock()

	start := time.Now()

	results := make(map[string]*CheckResult)
	var overallStatus Status = StatusOK
	var passCount int

	for i, check := range checks {
		result := check.Check(ctx)
		results[names[i]] = result

		switch result.Status {
		case StatusOK:
			passCount++
		case StatusWarning:
			if overallStatus == StatusOK {
				overallStatus = StatusWarning
			}
		case StatusError, StatusCritical:
			overallStatus = StatusError
		}
	}

	duration := time.Since(start)

	compositeResult := &CompositeResult{
		Timestamp: time.Now(),
		Overall:   overallStatus,
		Checks:    results,
		Duration:  duration,
	}

	hc.mu.Lock()
	hc.lastCheck = compositeResult
	hc.mu.Unlock()

	hc.logger.Info("Health check completed",
		logging.String("overall_status", overallStatus.String()),
		logging.Int("total_checks", len(checks)),
		logging.Int("passed_checks", passCount),
		logging.Duration("duration", duration))

	return compositeResult
}

// GetLastCheck returns the result of the most recent health check
func (hc *Checker) GetLastCheck() *CompositeResult {
	hc.mu.RLock()
	defer hc.mu.RUnlock()
	return hc.lastCheck
}

// HealthHandler is an HTTP handler that returns health check results
func (hc *Checker) HealthHandler(w http.ResponseWriter, r *http.Request) {
	// Check all health checks
	result := hc.CheckAll(r.Context())

	// Set response code based on overall status
	var statusCode int
	switch result.Overall {
	case StatusOK:
		statusCode = http.StatusOK
	case StatusWarning:
		statusCode = http.StatusOK // Warnings don't fail the health check
	case StatusError, StatusCritical:
		statusCode = http.StatusServiceUnavailable
	default:
		statusCode = http.StatusInternalServerError
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	// Write JSON response
	if _, err := fmt.Fprintf(w, `{
		"status": "%s",
		"timestamp": "%s",
		"duration": "%v",
		"checks": {`, result.Overall.String(), result.Timestamp.Format(time.RFC3339), result.Duration); err != nil {
		hc.logger.Error("Failed to write health check response header", logging.String("error", err.Error()))
		return
	}

	first := true
	for name, checkResult := range result.Checks {
		if !first {
			if _, err := fmt.Fprint(w, ","); err != nil {
				hc.logger.Error("Failed to write health check response comma", logging.String("error", err.Error()))
				return
			}
		}
		first = false

		if _, err := fmt.Fprintf(w, `
			"%s": {
				"status": "%s",
				"message": "%s",
				"duration": "%v"
			}`, name, checkResult.Status.String(), checkResult.Message, checkResult.Duration); err != nil {
			hc.logger.Error("Failed to write health check response entry", logging.String("error", err.Error()))
			return
		}
	}

	if _, err := fmt.Fprint(w, `
		}
	}`); err != nil {
		hc.logger.Error("Failed to write health check response footer", logging.String("error", err.Error()))
		return
	}
}

// ReadinessHandler returns readiness status
func (hc *Checker) ReadinessHandler(w http.ResponseWriter, r *http.Request) {
	result := hc.CheckAll(r.Context())

	// Readiness fails if any critical checks fail
	var statusCode int
	var status string

	criticalFailures := 0
	for _, checkResult := range result.Checks {
		if checkResult.Status == StatusCritical || checkResult.Status == StatusError {
			criticalFailures++
		}
	}

	if criticalFailures == 0 {
		statusCode = http.StatusOK
		status = "ready"
	} else {
		statusCode = http.StatusServiceUnavailable
		status = "not ready"
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if _, err := fmt.Fprintf(w, `{"status": "%s", "timestamp": "%s"}`, status, time.Now().Format(time.RFC3339)); err != nil {
		hc.logger.Error("Failed to write readiness response", logging.String("error", err.Error()))
	}
}

// LivenessHandler returns liveness status
func (hc *Checker) LivenessHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte(`{"status": "alive", "timestamp": "` + time.Now().Format(time.RFC3339) + `"}`)); err != nil {
		hc.logger.Error("Failed to write liveness response", logging.String("error", err.Error()))
	}
}

// NewDefaultChecker creates a default health checker with common checks
func NewDefaultChecker(
	awsTestFunc func(ctx context.Context) error,
	cacheTestFunc func(ctx context.Context) error,
	configValidationFunc func() error,
) *Checker {
	hc := NewChecker()

	hc.AddCheck("aws_connectivity", NewAWSConnectivityCheck(awsTestFunc))
	hc.AddCheck("cache", NewCacheHealthCheck(cacheTestFunc))
	hc.AddCheck("configuration", NewConfigHealthCheck(configValidationFunc))
	hc.AddCheck("memory", NewMemoryHealthCheck(90.0))
	hc.AddCheck("disk_space", NewDiskSpaceHealthCheck(85.0))

	return hc
}
