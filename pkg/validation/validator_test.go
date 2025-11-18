package validation

import (
	"context"
	"testing"
)

func TestNewResult(t *testing.T) {
	result := NewResult()
	if result == nil {
		t.Fatal("NewResult() returned nil")
	}
	if !result.Valid {
		t.Error("NewResult() should create a valid result")
	}
	if len(result.Errors) != 0 {
		t.Errorf("NewResult() should have no errors, got %d", len(result.Errors))
	}
	if result.Fields == nil {
		t.Error("NewResult() should initialize Fields map")
	}
}

func TestResultAddError(t *testing.T) {
	result := NewResult()
	result.AddError("test_field", "test error")

	if result.Valid {
		t.Error("Result should be invalid after adding error")
	}
	if len(result.Errors) != 1 {
		t.Errorf("Expected 1 error, got %d", len(result.Errors))
	}
	if result.Errors[0] != "test_field: test error" {
		t.Errorf("Unexpected error message: %s", result.Errors[0])
	}
}

func TestResultAddField(t *testing.T) {
	result := NewResult()
	result.AddField("test_key", "test_value")

	if result.Fields["test_key"] != "test_value" {
		t.Errorf("Expected field value 'test_value', got %v", result.Fields["test_key"])
	}
}

func TestInstanceIDValidator(t *testing.T) {
	validator := &InstanceIDValidator{}

	tests := []struct {
		name      string
		value     interface{}
		wantValid bool
		wantError string
	}{
		{
			name:      "valid instance ID",
			value:     "i-1234567890abcdef0",
			wantValid: true,
		},
		{
			name:      "nil value",
			value:     nil,
			wantValid: false,
			wantError: "cannot be nil",
		},
		{
			name:      "non-string value",
			value:     12345,
			wantValid: false,
			wantError: "must be a string",
		},
		{
			name:      "empty string",
			value:     "",
			wantValid: false,
			wantError: "cannot be empty",
		},
		{
			name:      "too long",
			value:     "i-12345678901234567890",
			wantValid: false,
			wantError: "too long",
		},
		{
			name:      "invalid format",
			value:     "invalid-id",
			wantValid: false,
			wantError: "invalid format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validator.Validate(tt.value)
			if result.Valid != tt.wantValid {
				t.Errorf("Valid = %v, want %v", result.Valid, tt.wantValid)
			}
			if !tt.wantValid && len(result.Errors) == 0 {
				t.Error("Expected errors but got none")
			}
		})
	}
}

func TestDNSNameValidator(t *testing.T) {
	validator := &DNSNameValidator{}

	tests := []struct {
		name      string
		value     interface{}
		wantValid bool
	}{
		{
			name:      "valid DNS name",
			value:     "example.com",
			wantValid: true,
		},
		{
			name:      "valid subdomain",
			value:     "sub.example.com",
			wantValid: true,
		},
		{
			name:      "nil value",
			value:     nil,
			wantValid: false,
		},
		{
			name:      "non-string value",
			value:     12345,
			wantValid: false,
		},
		{
			name:      "empty string",
			value:     "",
			wantValid: false,
		},
		{
			name:      "too long",
			value:     string(make([]byte, 300)),
			wantValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validator.Validate(tt.value)
			if result.Valid != tt.wantValid {
				t.Errorf("Valid = %v, want %v. Errors: %v", result.Valid, tt.wantValid, result.Errors)
			}
		})
	}
}

func TestIPAddressValidator(t *testing.T) {
	validator := &IPAddressValidator{}

	tests := []struct {
		name      string
		value     interface{}
		wantValid bool
	}{
		{
			name:      "valid IPv4",
			value:     "192.168.1.1",
			wantValid: true,
		},
		{
			name:      "valid IPv6",
			value:     "2001:0db8:85a3:0000:0000:8a2e:0370:7334",
			wantValid: true,
		},
		{
			name:      "nil value",
			value:     nil,
			wantValid: false,
		},
		{
			name:      "non-string value",
			value:     12345,
			wantValid: false,
		},
		{
			name:      "empty string",
			value:     "",
			wantValid: false,
		},
		{
			name:      "invalid IP",
			value:     "999.999.999.999",
			wantValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validator.Validate(tt.value)
			if result.Valid != tt.wantValid {
				t.Errorf("Valid = %v, want %v. Errors: %v", result.Valid, tt.wantValid, result.Errors)
			}
		})
	}
}

func TestRegionValidator(t *testing.T) {
	validator := &RegionValidator{}

	tests := []struct {
		name      string
		value     interface{}
		wantValid bool
	}{
		{
			name:      "valid region us-east-1",
			value:     "us-east-1",
			wantValid: true,
		},
		{
			name:      "valid region eu-west-1",
			value:     "eu-west-1",
			wantValid: true,
		},
		{
			name:      "nil value",
			value:     nil,
			wantValid: false,
		},
		{
			name:      "non-string value",
			value:     12345,
			wantValid: false,
		},
		{
			name:      "empty string",
			value:     "",
			wantValid: false,
		},
		{
			name:      "too long",
			value:     string(make([]byte, 40)),
			wantValid: false,
		},
		{
			name:      "invalid format",
			value:     "invalid-region",
			wantValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validator.Validate(tt.value)
			if result.Valid != tt.wantValid {
				t.Errorf("Valid = %v, want %v. Errors: %v", result.Valid, tt.wantValid, result.Errors)
			}
		})
	}
}

func TestProfileValidator(t *testing.T) {
	validator := &ProfileValidator{}

	tests := []struct {
		name      string
		value     interface{}
		wantValid bool
	}{
		{
			name:      "valid profile",
			value:     "default",
			wantValid: true,
		},
		{
			name:      "valid profile with hyphen",
			value:     "my-profile",
			wantValid: true,
		},
		{
			name:      "nil value",
			value:     nil,
			wantValid: false,
		},
		{
			name:      "non-string value",
			value:     12345,
			wantValid: false,
		},
		{
			name:      "empty string",
			value:     "",
			wantValid: false,
		},
		{
			name:      "too long",
			value:     string(make([]byte, 130)),
			wantValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validator.Validate(tt.value)
			if result.Valid != tt.wantValid {
				t.Errorf("Valid = %v, want %v. Errors: %v", result.Valid, tt.wantValid, result.Errors)
			}
		})
	}
}

func TestCommandValidator(t *testing.T) {
	validator := NewCommandValidator()

	tests := []struct {
		name      string
		value     interface{}
		wantValid bool
	}{
		{
			name:      "valid command",
			value:     "ls -la",
			wantValid: true,
		},
		{
			name:      "nil value",
			value:     nil,
			wantValid: false,
		},
		{
			name:      "non-string value",
			value:     12345,
			wantValid: false,
		},
		{
			name:      "empty string",
			value:     "",
			wantValid: false,
		},
		{
			name:      "too long",
			value:     string(make([]byte, 2000)),
			wantValid: false,
		},
		{
			name:      "dangerous pattern - command substitution",
			value:     "echo $(whoami)",
			wantValid: false,
		},
		{
			name:      "dangerous pattern - semicolon",
			value:     "ls; rm -rf /",
			wantValid: false,
		},
		{
			name:      "dangerous pattern - pipe with OR",
			value:     "ls || rm file",
			wantValid: false,
		},
		{
			name:      "dangerous pattern - pipe with AND",
			value:     "ls && rm file",
			wantValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validator.Validate(tt.value)
			if result.Valid != tt.wantValid {
				t.Errorf("Valid = %v, want %v. Errors: %v", result.Valid, tt.wantValid, result.Errors)
			}
		})
	}
}

func TestTagValidator(t *testing.T) {
	validator := &TagValidator{}

	tests := []struct {
		name      string
		key       interface{}
		value     interface{}
		wantValid bool
	}{
		{
			name:      "valid tag",
			key:       "Environment",
			value:     "production",
			wantValid: true,
		},
		{
			name:      "nil key",
			key:       nil,
			value:     "value",
			wantValid: false,
		},
		{
			name:      "non-string key",
			key:       12345,
			value:     "value",
			wantValid: false,
		},
		{
			name:      "empty key",
			key:       "",
			value:     "value",
			wantValid: false,
		},
		{
			name:      "key too long",
			key:       string(make([]byte, 130)),
			value:     "value",
			wantValid: false,
		},
		{
			name:      "nil value",
			key:       "key",
			value:     nil,
			wantValid: false,
		},
		{
			name:      "non-string value",
			key:       "key",
			value:     12345,
			wantValid: false,
		},
		{
			name:      "value too long",
			key:       "key",
			value:     string(make([]byte, 300)),
			wantValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validator.Validate(tt.key, tt.value)
			if result.Valid != tt.wantValid {
				t.Errorf("Valid = %v, want %v. Errors: %v", result.Valid, tt.wantValid, result.Errors)
			}
		})
	}
}

func TestSimpleNameValidator(t *testing.T) {
	validator := &SimpleNameValidator{}

	tests := []struct {
		name      string
		value     interface{}
		wantValid bool
	}{
		{
			name:      "valid name",
			value:     "my-server",
			wantValid: true,
		},
		{
			name:      "nil value",
			value:     nil,
			wantValid: false,
		},
		{
			name:      "non-string value",
			value:     12345,
			wantValid: false,
		},
		{
			name:      "empty string",
			value:     "",
			wantValid: false,
		},
		{
			name:      "too long",
			value:     string(make([]byte, 130)),
			wantValid: false,
		},
		{
			name:      "contains slash",
			value:     "my/server",
			wantValid: false,
		},
		{
			name:      "contains backslash",
			value:     "my\\server",
			wantValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validator.Validate(tt.value)
			if result.Valid != tt.wantValid {
				t.Errorf("Valid = %v, want %v. Errors: %v", result.Valid, tt.wantValid, result.Errors)
			}
		})
	}
}

func TestTagKeyValueValidator(t *testing.T) {
	validator := &TagKeyValueValidator{}

	tests := []struct {
		name      string
		value     interface{}
		wantValid bool
	}{
		{
			name:      "valid tag identifier",
			value:     "Environment:production",
			wantValid: true,
		},
		{
			name:      "nil value",
			value:     nil,
			wantValid: false,
		},
		{
			name:      "non-string value",
			value:     12345,
			wantValid: false,
		},
		{
			name:      "empty string",
			value:     "",
			wantValid: false,
		},
		{
			name:      "missing colon",
			value:     "Environment",
			wantValid: false,
		},
		{
			name:      "multiple colons",
			value:     "Env:prod:test",
			wantValid: true, // Should split on first colon only
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validator.Validate(tt.value)
			if result.Valid != tt.wantValid {
				t.Errorf("Valid = %v, want %v. Errors: %v", result.Valid, tt.wantValid, result.Errors)
			}
		})
	}
}

func TestCommandSanitizer(t *testing.T) {
	sanitizer := &CommandSanitizer{}

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "remove backticks",
			input:    "echo `whoami`",
			expected: "echo whoami",
		},
		{
			name:     "remove backslashes",
			input:    "echo test\\",
			expected: "echo test",
		},
		{
			name:     "trim whitespace",
			input:    "  echo test  ",
			expected: "echo test",
		},
		{
			name:     "truncate long command",
			input:    string(make([]byte, 2000)),
			expected: string(make([]byte, 1024)),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sanitizer.Sanitize(tt.input)
			if result != tt.expected {
				t.Errorf("Sanitize() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestDNSSanitizer(t *testing.T) {
	sanitizer := &DNSSanitizer{}

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "convert to lowercase",
			input:    "EXAMPLE.COM",
			expected: "example.com",
		},
		{
			name:     "remove whitespace",
			input:    "example .com",
			expected: "example.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sanitizer.Sanitize(tt.input)
			if result != tt.expected {
				t.Errorf("Sanitize() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestTagSanitizer(t *testing.T) {
	sanitizer := &TagSanitizer{}

	t.Run("SanitizeTagKey", func(t *testing.T) {
		tests := []struct {
			name     string
			input    string
			expected string
		}{
			{
				name:     "trim whitespace",
				input:    "  Environment  ",
				expected: "Environment",
			},
			{
				name:     "remove quotes",
				input:    `"Environment"`,
				expected: "Environment",
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result := sanitizer.SanitizeTagKey(tt.input)
				if result != tt.expected {
					t.Errorf("SanitizeTagKey() = %q, want %q", result, tt.expected)
				}
			})
		}
	})

	t.Run("SanitizeTagValue", func(t *testing.T) {
		tests := []struct {
			name     string
			input    string
			expected string
		}{
			{
				name:     "trim whitespace",
				input:    "  production  ",
				expected: "production",
			},
			{
				name:     "remove quotes",
				input:    `'production'`,
				expected: "production",
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result := sanitizer.SanitizeTagValue(tt.input)
				if result != tt.expected {
					t.Errorf("SanitizeTagValue() = %q, want %q", result, tt.expected)
				}
			})
		}
	})
}

func TestChain(t *testing.T) {
	t.Run("NewChain", func(t *testing.T) {
		chain := NewChain()
		if chain == nil {
			t.Fatal("NewChain() returned nil")
		}
		if chain.validators == nil {
			t.Error("Chain validators not initialized")
		}
	})

	t.Run("Add and Validate", func(t *testing.T) {
		chain := NewChain()
		chain.Add(&InstanceIDValidator{})

		result := chain.Validate("i-1234567890abcdef0")
		if !result.Valid {
			t.Errorf("Expected valid result, got errors: %v", result.Errors)
		}
	})
}

func TestValidateWithContext(t *testing.T) {
	ctx := context.Background()
	validator := &InstanceIDValidator{}
	result := ValidateWithContext(ctx, "i-1234567890abcdef0", validator)

	if !result.Valid {
		t.Errorf("Expected valid result, got errors: %v", result.Errors)
	}
}

func TestIsASCII(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "ASCII string",
			input:    "Hello World",
			expected: true,
		},
		{
			name:     "non-ASCII string",
			input:    "Hello 世界",
			expected: false,
		},
		{
			name:     "empty string",
			input:    "",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsASCII(tt.input)
			if result != tt.expected {
				t.Errorf("IsASCII(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestHasDangerousCharacters(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "safe string",
			input:    "hello world",
			expected: false,
		},
		{
			name:     "contains semicolon",
			input:    "hello; world",
			expected: true,
		},
		{
			name:     "contains dollar sign",
			input:    "hello $world",
			expected: true,
		},
		{
			name:     "contains backtick",
			input:    "hello `world`",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := HasDangerousCharacters(tt.input)
			if result != tt.expected {
				t.Errorf("HasDangerousCharacters(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestTruncateString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		maxLen   int
		expected string
	}{
		{
			name:     "no truncation needed",
			input:    "hello",
			maxLen:   10,
			expected: "hello",
		},
		{
			name:     "truncate",
			input:    "hello world",
			maxLen:   5,
			expected: "hello",
		},
		{
			name:     "empty string",
			input:    "",
			maxLen:   5,
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := TruncateString(tt.input, tt.maxLen)
			if result != tt.expected {
				t.Errorf("TruncateString(%q, %d) = %q, want %q", tt.input, tt.maxLen, result, tt.expected)
			}
		})
	}
}

func TestDefaultValidators(t *testing.T) {
	// Test that default validators are initialized
	if InstanceID == nil {
		t.Error("InstanceID validator is nil")
	}
	if DNSName == nil {
		t.Error("DNSName validator is nil")
	}
	if IPAddress == nil {
		t.Error("IPAddress validator is nil")
	}
	if Region == nil {
		t.Error("Region validator is nil")
	}
	if Profile == nil {
		t.Error("Profile validator is nil")
	}
	if Command == nil {
		t.Error("Command validator is nil")
	}
	if Tag == nil {
		t.Error("Tag validator is nil")
	}
	if Identifier == nil {
		t.Error("Identifier validator is nil")
	}
	if CommandSanit == nil {
		t.Error("CommandSanit sanitizer is nil")
	}
	if DNSSanit == nil {
		t.Error("DNSSanit sanitizer is nil")
	}
	if TagSanit == nil {
		t.Error("TagSanit sanitizer is nil")
	}
}

func TestIdentifierValidator(t *testing.T) {
	validator := &IdentifierValidator{}

	tests := []struct {
		name      string
		value     interface{}
		wantValid bool
	}{
		{
			name:      "valid instance ID",
			value:     "i-1234567890abcdef0",
			wantValid: true,
		},
		{
			name:      "valid IP",
			value:     "192.168.1.1",
			wantValid: true,
		},
		{
			name:      "valid DNS",
			value:     "example.com",
			wantValid: true,
		},
		{
			name:      "nil value",
			value:     nil,
			wantValid: false,
		},
		{
			name:      "empty string",
			value:     "",
			wantValid: false,
		},
		{
			name:      "too long",
			value:     string(make([]byte, 300)),
			wantValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validator.Validate(tt.value)
			if result.Valid != tt.wantValid {
				t.Errorf("Valid = %v, want %v. Errors: %v", result.Valid, tt.wantValid, result.Errors)
			}
		})
	}
}
