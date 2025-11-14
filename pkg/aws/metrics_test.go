package aws

import (
	"testing"
	"time"
)

func TestPerformanceMetrics_APIMetrics(t *testing.T) {
	metrics := NewPerformanceMetrics(true)

	// Record some API calls
	metrics.RecordAPICall("DescribeInstances", 100*time.Millisecond, nil, 0)
	metrics.RecordAPICall("DescribeInstances", 150*time.Millisecond, nil, 0)
	metrics.RecordAPICall("DescribeInstances", 200*time.Millisecond, nil, 0)

	apiMetrics, exists := metrics.GetAPIMetrics("DescribeInstances")
	if !exists {
		t.Fatal("API metrics should exist")
	}
	if apiMetrics.Calls != 3 {
		t.Errorf("expected 3 total calls, got %d", apiMetrics.Calls)
	}
	if apiMetrics.Successes != 3 {
		t.Errorf("expected 3 successful calls, got %d", apiMetrics.Successes)
	}
	if apiMetrics.Failures != 0 {
		t.Errorf("expected 0 failed calls, got %d", apiMetrics.Failures)
	}
	if apiMetrics.AvgDuration <= 0 {
		t.Error("average latency should be > 0")
	}

	t.Logf("API metrics: calls=%d, avg_latency=%v",
		apiMetrics.Calls, apiMetrics.AvgDuration)
}

func TestPerformanceMetrics_CacheMetrics(t *testing.T) {
	metrics := NewPerformanceMetrics(true)

	// Record cache hits and misses
	for i := 0; i < 8; i++ {
		metrics.RecordCacheHit()
	}
	for i := 0; i < 2; i++ {
		metrics.RecordCacheMiss()
	}

	summary := metrics.GetSummary()
	if summary.CacheHits != 8 {
		t.Errorf("expected 8 cache hits, got %d", summary.CacheHits)
	}
	if summary.CacheMisses != 2 {
		t.Errorf("expected 2 cache misses, got %d", summary.CacheMisses)
	}
	if summary.CacheHitRate < 79.0 || summary.CacheHitRate > 81.0 {
		t.Errorf("expected hit rate around 80%%, got %.2f%%", summary.CacheHitRate)
	}

	t.Logf("Cache metrics: hits=%d, misses=%d, hit_rate=%.2f%%",
		summary.CacheHits, summary.CacheMisses, summary.CacheHitRate)
}

func TestPerformanceMetrics_MemoryTracking(t *testing.T) {
	metrics := NewPerformanceMetrics(true)

	// Record memory usage
	metrics.RecordMemoryUsage(1024 * 1024 * 10) // 10MB
	metrics.RecordMemoryUsage(1024 * 1024 * 15) // 15MB
	metrics.RecordMemoryUsage(1024 * 1024 * 12) // 12MB

	summary := metrics.GetSummary()
	if summary.MemoryUsageBytes == 0 {
		t.Error("memory usage should not be 0")
	}

	t.Logf("Memory metrics: current=%dMB",
		summary.MemoryUsageBytes/(1024*1024))
}

func TestPerformanceMetrics_SuccessRate(t *testing.T) {
	metrics := NewPerformanceMetrics(true)

	// Record successful calls
	for i := 0; i < 9; i++ {
		metrics.RecordAPICall("TestOp", 100*time.Millisecond, nil, 0)
	}

	// Record one failed call
	metrics.RecordAPICall("TestOp", 100*time.Millisecond, &TestError{msg: "test error"}, 0)

	summary := metrics.GetSummary()
	expectedRate := 90.0
	if summary.SuccessRate < expectedRate-1 || summary.SuccessRate > expectedRate+1 {
		t.Errorf("expected success rate around %.2f%%, got %.2f%%", expectedRate, summary.SuccessRate)
	}

	t.Logf("Success rate: %.2f%% (%d/%d successful)",
		summary.SuccessRate, summary.TotalSuccesses, summary.TotalCalls)
}

func TestMetricsCollector_Creation(t *testing.T) {
	metrics := NewPerformanceMetrics(true)
	collector := NewMetricsCollector(metrics, 100*time.Millisecond)

	if collector == nil {
		t.Fatal("collector should not be nil")
	}
	if collector.metrics != metrics {
		t.Error("collector should reference the correct metrics")
	}

	t.Log("Metrics collector created successfully")
}

func TestInstrumentedClient_MetricsTracking(t *testing.T) {
	// Create a mock client (nil is acceptable for this test)
	var baseClient *Client

	// Create instrumented client with metrics enabled
	client := NewInstrumentedClient(baseClient, true)

	if client.metrics == nil {
		t.Fatal("instrumented client should have metrics")
	}

	// Record some operations through the client's metrics
	client.metrics.RecordAPICall("TestOperation", 100*time.Millisecond, nil, 0)

	apiMetrics, exists := client.metrics.GetAPIMetrics("TestOperation")
	if !exists {
		t.Fatal("API metrics should exist")
	}
	if apiMetrics.Calls != 1 {
		t.Errorf("expected 1 call, got %d", apiMetrics.Calls)
	}

	t.Log("Instrumented client metrics working correctly")
}

func TestPerformanceMetrics_Disabled(t *testing.T) {
	// Create metrics with disabled flag
	metrics := NewPerformanceMetrics(false)

	// Try to record metrics - should be no-ops
	metrics.RecordAPICall("TestOp", 100*time.Millisecond, nil, 0)
	metrics.RecordCacheHit()
	metrics.RecordCacheMiss()

	summary := metrics.GetSummary()
	if summary.TotalCalls != 0 {
		t.Error("disabled metrics should not record calls")
	}
	if summary.CacheHits != 0 {
		t.Error("disabled metrics should not record cache hits")
	}

	t.Log("Disabled metrics working correctly")
}

// TestError is a simple error type for testing
type TestError struct {
	msg string
}

func (e *TestError) Error() string {
	return e.msg
}
