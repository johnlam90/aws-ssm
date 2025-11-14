package features

import "testing"

func TestDefaultGates(t *testing.T) {
	gates := DefaultGates()

	if gates == nil {
		t.Fatal("DefaultGates returned nil")
	}

	if !gates.Metrics.IsEnabled() {
		t.Error("Metrics gate should be enabled by default")
	}

	if !gates.HealthChecks.IsEnabled() {
		t.Error("HealthChecks gate should be enabled by default")
	}

	if !gates.Security.IsEnabled() {
		t.Error("Security gate should be enabled by default")
	}

	if !gates.Validation.IsEnabled() {
		t.Error("Validation gate should be enabled by default")
	}
}

func TestGateIsEnabled(t *testing.T) {
	gate := &Gate{name: "test", enabled: true}

	if !gate.IsEnabled() {
		t.Error("Gate should be enabled")
	}

	gate.SetEnabled(false)
	if gate.IsEnabled() {
		t.Error("Gate should be disabled after SetEnabled(false)")
	}

	gate.SetEnabled(true)
	if !gate.IsEnabled() {
		t.Error("Gate should be enabled after SetEnabled(true)")
	}
}

func TestNilGateIsEnabled(t *testing.T) {
	var gate *Gate
	if gate.IsEnabled() {
		t.Error("Nil gate should return false for IsEnabled()")
	}
}

func TestLoadFromEnvironment(t *testing.T) {
	t.Setenv("AWS_SSM_FEATURE_METRICS", "false")
	t.Setenv("AWS_SSM_FEATURE_HEALTH_CHECKS", "false")
	t.Setenv("AWS_SSM_FEATURE_SECURITY", "true")
	t.Setenv("AWS_SSM_FEATURE_VALIDATION", "true")

	gates := DefaultGates()
	gates.LoadFromEnvironment()

	if gates.Metrics.IsEnabled() {
		t.Error("Metrics gate should be disabled")
	}

	if gates.HealthChecks.IsEnabled() {
		t.Error("HealthChecks gate should be disabled")
	}

	if !gates.Security.IsEnabled() {
		t.Error("Security gate should be enabled")
	}

	if !gates.Validation.IsEnabled() {
		t.Error("Validation gate should be enabled")
	}
}

func TestLoadFromString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected map[string]bool
	}{
		{
			name:  "all enabled",
			input: "metrics=true,health_checks=true,security=true,validation=true",
			expected: map[string]bool{
				"metrics":       true,
				"health_checks": true,
				"security":      true,
				"validation":    true,
			},
		},
		{
			name:  "all disabled",
			input: "metrics=false,health_checks=false,security=false,validation=false",
			expected: map[string]bool{
				"metrics":       false,
				"health_checks": false,
				"security":      false,
				"validation":    false,
			},
		},
		{
			name:  "mixed",
			input: "metrics=true,health_checks=false,security=true,validation=false",
			expected: map[string]bool{
				"metrics":       true,
				"health_checks": false,
				"security":      true,
				"validation":    false,
			},
		},
		{
			name:  "with spaces",
			input: "metrics = true , health_checks = false , security = true , validation = false",
			expected: map[string]bool{
				"metrics":       true,
				"health_checks": false,
				"security":      true,
				"validation":    false,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gates := DefaultGates()
			err := gates.LoadFromString(tt.input)

			if err != nil {
				t.Errorf("LoadFromString failed: %v", err)
			}

			if gates.Metrics.IsEnabled() != tt.expected["metrics"] {
				t.Errorf("Metrics gate mismatch: got %v, want %v", gates.Metrics.IsEnabled(), tt.expected["metrics"])
			}

			if gates.HealthChecks.IsEnabled() != tt.expected["health_checks"] {
				t.Errorf("HealthChecks gate mismatch: got %v, want %v", gates.HealthChecks.IsEnabled(), tt.expected["health_checks"])
			}

			if gates.Security.IsEnabled() != tt.expected["security"] {
				t.Errorf("Security gate mismatch: got %v, want %v", gates.Security.IsEnabled(), tt.expected["security"])
			}

			if gates.Validation.IsEnabled() != tt.expected["validation"] {
				t.Errorf("Validation gate mismatch: got %v, want %v", gates.Validation.IsEnabled(), tt.expected["validation"])
			}
		})
	}
}

func TestSummary(t *testing.T) {
	gates := DefaultGates()
	summary := gates.Summary()

	if summary == "" {
		t.Error("Summary should not be empty")
	}

	if !contains(summary, "Feature Gates") {
		t.Error("Summary should contain 'Feature Gates'")
	}

	if !contains(summary, "Metrics") {
		t.Error("Summary should contain 'Metrics'")
	}

	if !contains(summary, "Health Checks") {
		t.Error("Summary should contain 'Health Checks'")
	}

	if !contains(summary, "Security") {
		t.Error("Summary should contain 'Security'")
	}

	if !contains(summary, "Validation") {
		t.Error("Summary should contain 'Validation'")
	}
}

func contains(s, substr string) bool {
	for i := 0; i < len(s)-len(substr)+1; i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
