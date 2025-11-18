package health

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestStatus_String(t *testing.T) {
	tests := []struct {
		status   Status
		expected string
	}{
		{StatusOK, "ok"},
		{StatusWarning, "warning"},
		{StatusError, "error"},
		{StatusCritical, "critical"},
		{StatusUnknown, "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := tt.status.String()
			if result != tt.expected {
				t.Errorf("Status.String() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestNewCheckResult(t *testing.T) {
	result := NewCheckResult(StatusOK, "test message")

	if result == nil {
		t.Fatal("NewCheckResult() returned nil")
	}
	if result.Status != StatusOK {
		t.Errorf("Status = %v, want %v", result.Status, StatusOK)
	}
	if result.Message != "test message" {
		t.Errorf("Message = %q, want %q", result.Message, "test message")
	}
	if result.Metadata == nil {
		t.Error("Metadata should be initialized")
	}
}

func TestCheckResult_WithDuration(t *testing.T) {
	result := NewCheckResult(StatusOK, "test")
	duration := time.Second

	result.WithDuration(duration)

	if result.Duration != duration {
		t.Errorf("Duration = %v, want %v", result.Duration, duration)
	}
}

func TestCheckResult_WithMetadata(t *testing.T) {
	result := NewCheckResult(StatusOK, "test")

	result.WithMetadata("key", "value")

	if result.Metadata["key"] != "value" {
		t.Errorf("Metadata[\"key\"] = %v, want \"value\"", result.Metadata["key"])
	}
}

func TestCheckResult_WithServiceName(t *testing.T) {
	result := NewCheckResult(StatusOK, "test")

	result.WithServiceName("test-service")

	if result.ServiceName != "test-service" {
		t.Errorf("ServiceName = %q, want \"test-service\"", result.ServiceName)
	}
}

func TestNewBaseHealthCheck(t *testing.T) {
	check := NewBaseHealthCheck("test-check")

	if check == nil {
		t.Fatal("NewBaseHealthCheck() returned nil")
	}
	if check.name != "test-check" {
		t.Errorf("name = %q, want \"test-check\"", check.name)
	}
	if check.logger == nil {
		t.Error("logger should be initialized")
	}
}

func TestBaseHealthCheck_Name(t *testing.T) {
	check := NewBaseHealthCheck("test-check")

	name := check.Name()
	if name != "test-check" {
		t.Errorf("Name() = %q, want \"test-check\"", name)
	}
}

func TestBaseHealthCheck_GetLastCheck(t *testing.T) {
	check := NewBaseHealthCheck("test-check")

	// Initially should be nil
	if check.GetLastCheck() != nil {
		t.Error("GetLastCheck() should initially return nil")
	}

	// Set a result
	result := NewCheckResult(StatusOK, "test")
	check.setLastCheck(result)

	lastCheck := check.GetLastCheck()
	if lastCheck == nil {
		t.Fatal("GetLastCheck() should return the last check result")
	}
	if lastCheck.Status != StatusOK {
		t.Errorf("Last check status = %v, want %v", lastCheck.Status, StatusOK)
	}
}

func TestAWSConnectivityCheck(t *testing.T) {
	t.Run("successful check", func(t *testing.T) {
		testFunc := func(ctx context.Context) error {
			return nil
		}

		check := NewAWSConnectivityCheck(testFunc)
		result := check.Check(context.Background())

		if result.Status != StatusOK {
			t.Errorf("Status = %v, want %v", result.Status, StatusOK)
		}
		if result.ServiceName != "aws" {
			t.Errorf("ServiceName = %q, want \"aws\"", result.ServiceName)
		}
	})

	t.Run("failed check", func(t *testing.T) {
		testFunc := func(ctx context.Context) error {
			return errors.New("connection failed")
		}

		check := NewAWSConnectivityCheck(testFunc)
		result := check.Check(context.Background())

		if result.Status != StatusError {
			t.Errorf("Status = %v, want %v", result.Status, StatusError)
		}
		if !strings.Contains(result.Message, "failed") {
			t.Errorf("Message should contain 'failed', got %q", result.Message)
		}
	})
}

func TestCacheHealthCheck(t *testing.T) {
	t.Run("successful check", func(t *testing.T) {
		testFunc := func(ctx context.Context) error {
			return nil
		}

		check := NewCacheHealthCheck(testFunc)
		result := check.Check(context.Background())

		if result.Status != StatusOK {
			t.Errorf("Status = %v, want %v", result.Status, StatusOK)
		}
		if result.ServiceName != "cache" {
			t.Errorf("ServiceName = %q, want \"cache\"", result.ServiceName)
		}
	})

	t.Run("failed check", func(t *testing.T) {
		testFunc := func(ctx context.Context) error {
			return errors.New("cache error")
		}

		check := NewCacheHealthCheck(testFunc)
		result := check.Check(context.Background())

		if result.Status != StatusError {
			t.Errorf("Status = %v, want %v", result.Status, StatusError)
		}
	})
}

func TestConfigHealthCheck(t *testing.T) {
	t.Run("successful check", func(t *testing.T) {
		validateFunc := func() error {
			return nil
		}

		check := NewConfigHealthCheck(validateFunc)
		result := check.Check(context.Background())

		if result.Status != StatusOK {
			t.Errorf("Status = %v, want %v", result.Status, StatusOK)
		}
		if result.ServiceName != "config" {
			t.Errorf("ServiceName = %q, want \"config\"", result.ServiceName)
		}
	})

	t.Run("failed check", func(t *testing.T) {
		validateFunc := func() error {
			return errors.New("invalid config")
		}

		check := NewConfigHealthCheck(validateFunc)
		result := check.Check(context.Background())

		if result.Status != StatusError {
			t.Errorf("Status = %v, want %v", result.Status, StatusError)
		}
	})
}

func TestMemoryHealthCheck(t *testing.T) {
	check := NewMemoryHealthCheck(90.0)

	result := check.Check(context.Background())

	if result == nil {
		t.Fatal("Check() returned nil")
	}
	if result.ServiceName != "system" {
		t.Errorf("ServiceName = %q, want \"system\"", result.ServiceName)
	}
	if result.Metadata["memory_usage_percent"] == nil {
		t.Error("Metadata should contain memory_usage_percent")
	}
}

func TestDiskSpaceHealthCheck(t *testing.T) {
	check := NewDiskSpaceHealthCheck(85.0)

	result := check.Check(context.Background())

	if result == nil {
		t.Fatal("Check() returned nil")
	}
	if result.ServiceName != "system" {
		t.Errorf("ServiceName = %q, want \"system\"", result.ServiceName)
	}
	if result.Metadata["disk_usage_percent"] == nil {
		t.Error("Metadata should contain disk_usage_percent")
	}
}

func TestTCPHealthCheck(t *testing.T) {
	t.Run("failed check - unreachable host", func(t *testing.T) {
		check := NewTCPHealthCheck("localhost", 9999, time.Second)
		result := check.Check(context.Background())

		if result.Status != StatusError {
			t.Errorf("Status = %v, want %v for unreachable host", result.Status, StatusError)
		}
		if result.ServiceName != "network" {
			t.Errorf("ServiceName = %q, want \"network\"", result.ServiceName)
		}
	})

	t.Run("check name", func(t *testing.T) {
		check := NewTCPHealthCheck("localhost", 80, time.Second)
		if check.Name() != "tcp_connectivity" {
			t.Errorf("Name() = %q, want \"tcp_connectivity\"", check.Name())
		}
	})
}

func TestHTTPHealthCheck(t *testing.T) {
	t.Run("successful check", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		check := NewHTTPHealthCheck(server.URL, time.Second)
		result := check.Check(context.Background())

		if result.Status != StatusOK {
			t.Errorf("Status = %v, want %v", result.Status, StatusOK)
		}
		if result.ServiceName != "http" {
			t.Errorf("ServiceName = %q, want \"http\"", result.ServiceName)
		}
	})

	t.Run("failed check - error status", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer server.Close()

		check := NewHTTPHealthCheck(server.URL, time.Second)
		result := check.Check(context.Background())

		if result.Status != StatusError {
			t.Errorf("Status = %v, want %v", result.Status, StatusError)
		}
	})

	t.Run("failed check - invalid URL", func(t *testing.T) {
		check := NewHTTPHealthCheck("http://invalid-host-that-does-not-exist-12345.com", time.Second)
		result := check.Check(context.Background())

		if result.Status != StatusError {
			t.Errorf("Status = %v, want %v", result.Status, StatusError)
		}
	})

	t.Run("check name", func(t *testing.T) {
		check := NewHTTPHealthCheck("http://example.com", time.Second)
		if check.Name() != "http_endpoint" {
			t.Errorf("Name() = %q, want \"http_endpoint\"", check.Name())
		}
	})
}

func TestCompositeHealthCheck(t *testing.T) {
	t.Run("all checks pass", func(t *testing.T) {
		checks := []Check{
			NewAWSConnectivityCheck(func(ctx context.Context) error { return nil }),
			NewCacheHealthCheck(func(ctx context.Context) error { return nil }),
		}

		composite := NewCompositeHealthCheck(checks, 2)
		result := composite.Check(context.Background())

		if result.Status != StatusOK {
			t.Errorf("Status = %v, want %v", result.Status, StatusOK)
		}
	})

	t.Run("some checks fail", func(t *testing.T) {
		checks := []Check{
			NewAWSConnectivityCheck(func(ctx context.Context) error { return nil }),
			NewCacheHealthCheck(func(ctx context.Context) error { return errors.New("error") }),
		}

		composite := NewCompositeHealthCheck(checks, 2)
		result := composite.Check(context.Background())

		if result.Status != StatusError {
			t.Errorf("Status = %v, want %v", result.Status, StatusError)
		}
	})

	t.Run("below threshold", func(t *testing.T) {
		checks := []Check{
			NewAWSConnectivityCheck(func(ctx context.Context) error { return nil }),
			NewCacheHealthCheck(func(ctx context.Context) error { return errors.New("error") }),
		}

		composite := NewCompositeHealthCheck(checks, 2)
		result := composite.Check(context.Background())

		if result.Status != StatusError {
			t.Errorf("Status = %v, want %v when below threshold", result.Status, StatusError)
		}
	})
}

func TestNewChecker(t *testing.T) {
	checker := NewChecker()

	if checker == nil {
		t.Fatal("NewChecker() returned nil")
	}
	if checker.checks == nil {
		t.Error("checks map should be initialized")
	}
	if checker.logger == nil {
		t.Error("logger should be initialized")
	}
}

func TestChecker_AddCheck(t *testing.T) {
	checker := NewChecker()
	check := NewAWSConnectivityCheck(func(ctx context.Context) error { return nil })

	checker.AddCheck("test", check)

	if len(checker.checks) != 1 {
		t.Errorf("Expected 1 check, got %d", len(checker.checks))
	}
}

func TestChecker_RemoveCheck(t *testing.T) {
	checker := NewChecker()
	check := NewAWSConnectivityCheck(func(ctx context.Context) error { return nil })

	checker.AddCheck("test", check)
	checker.RemoveCheck("test")

	if len(checker.checks) != 0 {
		t.Errorf("Expected 0 checks after removal, got %d", len(checker.checks))
	}
}

func TestChecker_GetCheck(t *testing.T) {
	checker := NewChecker()
	check := NewAWSConnectivityCheck(func(ctx context.Context) error { return nil })

	checker.AddCheck("test", check)

	retrievedCheck, exists := checker.GetCheck("test")
	if !exists {
		t.Error("Check should exist")
	}
	if retrievedCheck == nil {
		t.Error("Retrieved check should not be nil")
	}

	_, exists = checker.GetCheck("nonexistent")
	if exists {
		t.Error("Nonexistent check should not exist")
	}
}

func TestChecker_ListChecks(t *testing.T) {
	checker := NewChecker()
	checker.AddCheck("check1", NewAWSConnectivityCheck(func(ctx context.Context) error { return nil }))
	checker.AddCheck("check2", NewCacheHealthCheck(func(ctx context.Context) error { return nil }))

	names := checker.ListChecks()
	if len(names) != 2 {
		t.Errorf("Expected 2 check names, got %d", len(names))
	}
}

func TestChecker_CheckAll(t *testing.T) {
	checker := NewChecker()
	checker.AddCheck("aws", NewAWSConnectivityCheck(func(ctx context.Context) error { return nil }))
	checker.AddCheck("cache", NewCacheHealthCheck(func(ctx context.Context) error { return nil }))

	result := checker.CheckAll(context.Background())

	if result == nil {
		t.Fatal("CheckAll() returned nil")
	}
	if result.Overall != StatusOK {
		t.Errorf("Overall status = %v, want %v", result.Overall, StatusOK)
	}
	if len(result.Checks) != 2 {
		t.Errorf("Expected 2 check results, got %d", len(result.Checks))
	}
}

func TestChecker_GetLastCheck(t *testing.T) {
	checker := NewChecker()
	checker.AddCheck("test", NewAWSConnectivityCheck(func(ctx context.Context) error { return nil }))

	// Initially should be nil
	if checker.GetLastCheck() != nil {
		t.Error("GetLastCheck() should initially return nil")
	}

	// Run a check
	checker.CheckAll(context.Background())

	lastCheck := checker.GetLastCheck()
	if lastCheck == nil {
		t.Error("GetLastCheck() should return result after CheckAll()")
	}
}

func TestChecker_HealthHandler(t *testing.T) {
	checker := NewChecker()
	checker.AddCheck("test", NewAWSConnectivityCheck(func(ctx context.Context) error { return nil }))

	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	checker.HealthHandler(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Status code = %d, want %d", resp.StatusCode, http.StatusOK)
	}
	if contentType := resp.Header.Get("Content-Type"); contentType != "application/json" {
		t.Errorf("Content-Type = %q, want \"application/json\"", contentType)
	}
}

func TestChecker_HealthHandler_Error(t *testing.T) {
	checker := NewChecker()
	checker.AddCheck("test", NewAWSConnectivityCheck(func(ctx context.Context) error {
		return errors.New("test error")
	}))

	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	checker.HealthHandler(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusServiceUnavailable {
		t.Errorf("Status code = %d, want %d", resp.StatusCode, http.StatusServiceUnavailable)
	}
}

func TestChecker_ReadinessHandler(t *testing.T) {
	t.Run("ready", func(t *testing.T) {
		checker := NewChecker()
		checker.AddCheck("test", NewAWSConnectivityCheck(func(ctx context.Context) error { return nil }))

		req := httptest.NewRequest("GET", "/ready", nil)
		w := httptest.NewRecorder()

		checker.ReadinessHandler(w, req)

		resp := w.Result()
		if resp.StatusCode != http.StatusOK {
			t.Errorf("Status code = %d, want %d", resp.StatusCode, http.StatusOK)
		}
	})

	t.Run("not ready", func(t *testing.T) {
		checker := NewChecker()
		checker.AddCheck("test", NewAWSConnectivityCheck(func(ctx context.Context) error {
			return errors.New("not ready")
		}))

		req := httptest.NewRequest("GET", "/ready", nil)
		w := httptest.NewRecorder()

		checker.ReadinessHandler(w, req)

		resp := w.Result()
		if resp.StatusCode != http.StatusServiceUnavailable {
			t.Errorf("Status code = %d, want %d", resp.StatusCode, http.StatusServiceUnavailable)
		}
	})
}

func TestChecker_LivenessHandler(t *testing.T) {
	checker := NewChecker()

	req := httptest.NewRequest("GET", "/live", nil)
	w := httptest.NewRecorder()

	checker.LivenessHandler(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Status code = %d, want %d", resp.StatusCode, http.StatusOK)
	}
	if contentType := resp.Header.Get("Content-Type"); contentType != "application/json" {
		t.Errorf("Content-Type = %q, want \"application/json\"", contentType)
	}
}

func TestNewDefaultChecker(t *testing.T) {
	awsTestFunc := func(ctx context.Context) error { return nil }
	cacheTestFunc := func(ctx context.Context) error { return nil }
	configValidationFunc := func() error { return nil }

	checker := NewDefaultChecker(awsTestFunc, cacheTestFunc, configValidationFunc)

	if checker == nil {
		t.Fatal("NewDefaultChecker() returned nil")
	}

	// Check that expected checks were added
	expectedChecks := []string{"aws_connectivity", "cache", "configuration", "memory", "disk_space"}
	for _, name := range expectedChecks {
		if _, exists := checker.GetCheck(name); !exists {
			t.Errorf("Expected check %q to be added", name)
		}
	}
}

func TestCheckAllWithWarnings(t *testing.T) {
	checker := NewChecker()

	// Add a check that returns a warning
	checker.AddCheck("memory", NewMemoryHealthCheck(10.0)) // Very low threshold, likely to trigger warning

	result := checker.CheckAll(context.Background())

	// We can't guarantee the exact status without knowing the system state,
	// but we can verify the check ran and returned a result
	if result == nil {
		t.Error("CheckAll() should return a result")
	}
	if len(result.Checks) != 1 {
		t.Errorf("Expected 1 check result, got %d", len(result.Checks))
	}
}

func TestHTTPHealthCheckWithInvalidRequest(t *testing.T) {
	// Test with malformed URL that causes request creation to fail
	check := &HTTPHealthCheck{
		BaseHealthCheck: NewBaseHealthCheck("http_test"),
		url:             "ht!tp://invalid\nurl",
		timeout:         time.Second,
	}

	result := check.Check(context.Background())
	if result.Status != StatusError {
		t.Errorf("Status = %v, want %v for invalid URL", result.Status, StatusError)
	}
}

// Prevent unused variable warning
var _ = fmt.Sprint("")
