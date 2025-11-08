package aws

import (
	"testing"
)

func TestMultipleInstancesErrorFormatting(t *testing.T) {
	tests := []struct {
		name        string
		identifier  string
		instances   []Instance
		expected    string
		description string
	}{
		{
			name:       "single instance",
			identifier: "web-server",
			instances: []Instance{
				{
					InstanceID: "i-1234567890abcdef0",
					Name:       "web-server",
					State:      "running",
					PrivateIP:  "10.0.1.100",
				},
			},
			expected:    "Found 1 instances matching 'web-server'",
			description: "Should format single instance correctly",
		},
		{
			name:       "multiple instances",
			identifier: "app-server",
			instances: []Instance{
				{
					InstanceID: "i-1111111111111111",
					Name:       "app-server-1",
					State:      "running",
					PrivateIP:  "10.0.1.10",
				},
				{
					InstanceID: "i-2222222222222222",
					Name:       "app-server-2",
					State:      "running",
					PrivateIP:  "10.0.1.20",
				},
				{
					InstanceID: "i-3333333333333333",
					Name:       "app-server-3",
					State:      "stopped",
					PrivateIP:  "10.0.1.30",
				},
			},
			expected:    "Found 3 instances matching 'app-server'",
			description: "Should format multiple instances correctly",
		},
		{
			name:       "instance without name",
			identifier: "i-1234567890abcdef0",
			instances: []Instance{
				{
					InstanceID: "i-1234567890abcdef0",
					Name:       "",
					State:      "running",
					PrivateIP:  "10.0.1.100",
				},
			},
			expected:    "(no name)",
			description: "Should show (no name) for instances without name tag",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := &MultipleInstancesError{
				Identifier:       tt.identifier,
				Instances:        tt.instances,
				AllowInteractive: true,
			}

			formatted := err.FormatInstanceList()
			if !contains(formatted, tt.expected) {
				t.Errorf("%s: expected '%s' in output, got: %s", tt.description, tt.expected, formatted)
			}
		})
	}
}

func TestMultipleInstancesErrorMessage(t *testing.T) {
	err := &MultipleInstancesError{
		Identifier: "test-instance",
		Instances: []Instance{
			{InstanceID: "i-1111111111111111"},
			{InstanceID: "i-2222222222222222"},
		},
		AllowInteractive: true,
	}

	expected := "multiple instances found matching 'test-instance'"
	if err.Error() != expected {
		t.Errorf("expected error message '%s', got '%s'", expected, err.Error())
	}
}

func TestMultipleInstancesErrorFormatting_WithInteractiveFlag(t *testing.T) {
	err := &MultipleInstancesError{
		Identifier: "web-server",
		Instances: []Instance{
			{
				InstanceID: "i-1234567890abcdef0",
				Name:       "web-server-1",
				State:      "running",
				PrivateIP:  "10.0.1.100",
			},
			{
				InstanceID: "i-0987654321fedcba0",
				Name:       "web-server-2",
				State:      "running",
				PrivateIP:  "10.0.1.101",
			},
		},
		AllowInteractive: true,
	}

	formatted := err.FormatInstanceList()
	if !contains(formatted, "Opening interactive selector") {
		t.Errorf("expected 'Opening interactive selector' in output when AllowInteractive=true, got: %s", formatted)
	}
}

func TestMultipleInstancesErrorFormatting_WithoutInteractiveFlag(t *testing.T) {
	err := &MultipleInstancesError{
		Identifier: "web-server",
		Instances: []Instance{
			{
				InstanceID: "i-1234567890abcdef0",
				Name:       "web-server-1",
				State:      "running",
				PrivateIP:  "10.0.1.100",
			},
		},
		AllowInteractive: false,
	}

	formatted := err.FormatInstanceList()
	if contains(formatted, "Opening interactive selector") {
		t.Errorf("expected no 'Opening interactive selector' in output when AllowInteractive=false, got: %s", formatted)
	}
}

func TestMultipleInstancesErrorFormatting_InstanceDetails(t *testing.T) {
	instance := Instance{
		InstanceID: "i-1234567890abcdef0",
		Name:       "web-server",
		State:      "running",
		PrivateIP:  "10.0.1.100",
	}

	err := &MultipleInstancesError{
		Identifier:       "web-server",
		Instances:        []Instance{instance},
		AllowInteractive: true,
	}

	formatted := err.FormatInstanceList()

	// Check that all instance details are included
	if !contains(formatted, instance.InstanceID) {
		t.Errorf("expected instance ID '%s' in output", instance.InstanceID)
	}
	if !contains(formatted, instance.Name) {
		t.Errorf("expected instance name '%s' in output", instance.Name)
	}
	if !contains(formatted, instance.State) {
		t.Errorf("expected instance state '%s' in output", instance.State)
	}
	if !contains(formatted, instance.PrivateIP) {
		t.Errorf("expected instance IP '%s' in output", instance.PrivateIP)
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
