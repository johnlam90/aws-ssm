// Package features manages runtime feature gate configuration.
package features

import (
	"os"
	"strconv"
	"strings"
)

// Gate represents a feature gate that can be enabled or disabled
type Gate struct {
	name    string
	enabled bool
}

// Gates contains all available feature gates
type Gates struct {
	Metrics      *Gate
	HealthChecks *Gate
	Security     *Gate
	Validation   *Gate
}

// DefaultGates returns the default feature gates with all subsystems enabled
func DefaultGates() *Gates {
	return &Gates{
		Metrics:      &Gate{name: "metrics", enabled: true},
		HealthChecks: &Gate{name: "health_checks", enabled: true},
		Security:     &Gate{name: "security", enabled: true},
		Validation:   &Gate{name: "validation", enabled: true},
	}
}

// IsEnabled returns whether the gate is enabled
func (g *Gate) IsEnabled() bool {
	return g != nil && g.enabled
}

// SetEnabled sets the gate's enabled state
func (g *Gate) SetEnabled(enabled bool) {
	if g != nil {
		g.enabled = enabled
	}
}

// LoadFromEnvironment loads feature gates from environment variables
// Environment variables should be in the format: AWS_SSM_FEATURE_<GATE_NAME>=true|false
// Example: AWS_SSM_FEATURE_METRICS=false
func (gates *Gates) LoadFromEnvironment() {
	if gates == nil {
		return
	}

	// Load metrics gate
	if val := os.Getenv("AWS_SSM_FEATURE_METRICS"); val != "" {
		if enabled, err := strconv.ParseBool(val); err == nil {
			gates.Metrics.SetEnabled(enabled)
		}
	}

	// Load health checks gate
	if val := os.Getenv("AWS_SSM_FEATURE_HEALTH_CHECKS"); val != "" {
		if enabled, err := strconv.ParseBool(val); err == nil {
			gates.HealthChecks.SetEnabled(enabled)
		}
	}

	// Load security gate
	if val := os.Getenv("AWS_SSM_FEATURE_SECURITY"); val != "" {
		if enabled, err := strconv.ParseBool(val); err == nil {
			gates.Security.SetEnabled(enabled)
		}
	}

	// Load validation gate
	if val := os.Getenv("AWS_SSM_FEATURE_VALIDATION"); val != "" {
		if enabled, err := strconv.ParseBool(val); err == nil {
			gates.Validation.SetEnabled(enabled)
		}
	}
}

// LoadFromString loads feature gates from a comma-separated string
// Format: "metrics=true,health_checks=false,security=true,validation=true"
func (gates *Gates) LoadFromString(gateString string) error {
	if gates == nil || gateString == "" {
		return nil
	}

	pairs := strings.Split(gateString, ",")
	for _, pair := range pairs {
		parts := strings.Split(strings.TrimSpace(pair), "=")
		if len(parts) != 2 {
			continue
		}

		gateName := strings.TrimSpace(parts[0])
		gateValue := strings.TrimSpace(parts[1])

		enabled, err := strconv.ParseBool(gateValue)
		if err != nil {
			continue
		}

		switch gateName {
		case "metrics":
			gates.Metrics.SetEnabled(enabled)
		case "health_checks":
			gates.HealthChecks.SetEnabled(enabled)
		case "security":
			gates.Security.SetEnabled(enabled)
		case "validation":
			gates.Validation.SetEnabled(enabled)
		}
	}

	return nil
}

// Summary returns a summary of all feature gates and their states
func (gates *Gates) Summary() string {
	if gates == nil {
		return "No feature gates configured"
	}

	var sb strings.Builder
	sb.WriteString("Feature Gates:\n")
	sb.WriteString("  Metrics: ")
	sb.WriteString(boolToString(gates.Metrics.IsEnabled()))
	sb.WriteString("\n")
	sb.WriteString("  Health Checks: ")
	sb.WriteString(boolToString(gates.HealthChecks.IsEnabled()))
	sb.WriteString("\n")
	sb.WriteString("  Security: ")
	sb.WriteString(boolToString(gates.Security.IsEnabled()))
	sb.WriteString("\n")
	sb.WriteString("  Validation: ")
	sb.WriteString(boolToString(gates.Validation.IsEnabled()))
	sb.WriteString("\n")

	return sb.String()
}

func boolToString(b bool) string {
	if b {
		return "enabled"
	}
	return "disabled"
}
