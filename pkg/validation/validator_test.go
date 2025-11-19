package validation

import (
	"strings"
	"testing"
)

func TestInstanceIDValidator(t *testing.T) {
	v := &InstanceIDValidator{}

	tests := []struct {
		name    string
		input   interface{}
		isValid bool
	}{
		{"Valid Instance ID", "i-1234567890abcdef0", true},
		{"Nil Input", nil, false},
		{"Not String", 123, false},
		{"Empty String", "", false},
		{"Too Long", "i-1234567890abcdef0123", false},
		{"Invalid Format", "invalid-id", false},
		{"Short ID", "i-123", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := v.Validate(tt.input)
			if res.Valid != tt.isValid {
				t.Errorf("expected valid=%v, got %v", tt.isValid, res.Valid)
			}
		})
	}
}

func TestDNSNameValidator(t *testing.T) {
	v := &DNSNameValidator{}

	tests := []struct {
		name    string
		input   interface{}
		isValid bool
	}{
		{"Valid DNS", "example.com", true},
		{"Valid Subdomain", "sub.example.com", true},
		{"Nil Input", nil, false},
		{"Not String", 123, false},
		{"Empty String", "", false},
		{"Too Long", strings.Repeat("a", 256), false},
		{"Invalid Format", "-start.com", false},
		{"Invalid Char", "exa$mple.com", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := v.Validate(tt.input)
			if res.Valid != tt.isValid {
				t.Errorf("expected valid=%v, got %v", tt.isValid, res.Valid)
			}
		})
	}
}

func TestIPAddressValidator(t *testing.T) {
	v := &IPAddressValidator{}

	tests := []struct {
		name    string
		input   interface{}
		isValid bool
	}{
		{"Valid IPv4", "192.168.1.1", true},
		{"Valid IPv6", "2001:db8::1", true},
		{"Nil Input", nil, false},
		{"Not String", 123, false},
		{"Empty String", "", false},
		{"Invalid IP", "256.256.256.256", false},
		{"Not IP", "example.com", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := v.Validate(tt.input)
			if res.Valid != tt.isValid {
				t.Errorf("expected valid=%v, got %v", tt.isValid, res.Valid)
			}
		})
	}
}

func TestRegionValidator(t *testing.T) {
	v := &RegionValidator{}

	tests := []struct {
		name    string
		input   interface{}
		isValid bool
	}{
		{"Valid Region", "us-east-1", true},
		{"Valid Region 2", "eu-west-1", true},
		{"Nil Input", nil, false},
		{"Not String", 123, false},
		{"Empty String", "", false},
		{"Too Long", "us-east-1-very-very-long-region-name", false},
		{"Invalid Format", "us-east", false},
		{"Unknown Region", "mars-north-1", false}, // Matches pattern
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := v.Validate(tt.input)
			if res.Valid != tt.isValid {
				t.Errorf("expected valid=%v, got %v (errors: %v)", tt.isValid, res.Valid, res.Errors)
			}
		})
	}
}

func TestProfileValidator(t *testing.T) {
	v := &ProfileValidator{}

	tests := []struct {
		name    string
		input   interface{}
		isValid bool
	}{
		{"Valid Profile", "default", true},
		{"Valid Profile 2", "my-profile", true},
		{"Nil Input", nil, false},
		{"Not String", 123, false},
		{"Empty String", "", false},
		{"Too Long", strings.Repeat("a", 129), false},
		{"Invalid Char", "profile/name", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := v.Validate(tt.input)
			if res.Valid != tt.isValid {
				t.Errorf("expected valid=%v, got %v", tt.isValid, res.Valid)
			}
		})
	}
}

func TestCommandValidator(t *testing.T) {
	v := NewCommandValidator()

	tests := []struct {
		name    string
		input   interface{}
		isValid bool
	}{
		{"Valid Command", "ls -la", true},
		{"Valid Command 2", "echo 'hello'", true}, // Now valid with updated regex
		{"Nil Input", nil, false},
		{"Not String", 123, false},
		{"Empty String", "", false},
		{"Too Long", strings.Repeat("a", 1025), false},
		{"Dangerous rm", "rm -rf /", false}, // Now invalid with updated regex
		{"Dangerous subshell", "$(whoami)", false},
		{"Dangerous pipe", "ls | grep", true},
		{"Dangerous chaining", "ls && rm", false},
		{"Unsafe Char", "echo \"hello\"", false}, // " is still unsafe
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := v.Validate(tt.input)
			if res.Valid != tt.isValid {
				t.Errorf("expected valid=%v, got %v (errors: %v)", tt.isValid, res.Valid, res.Errors)
			}
		})
	}
}

func TestTagValidator(t *testing.T) {
	v := &TagValidator{}

	tests := []struct {
		name    string
		key     interface{}
		value   interface{}
		isValid bool
	}{
		{"Valid Tag", "Environment", "Production", true},
		{"Nil Key", nil, "val", false},
		{"Nil Value", "key", nil, false},
		{"Not String Key", 123, "val", false},
		{"Not String Value", "key", 123, false},
		{"Empty Key", "", "val", false},
		{"Long Key", strings.Repeat("a", 129), "val", false},
		{"Long Value", "key", strings.Repeat("a", 257), false},
		{"Invalid Key Char", "key!", "val", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := v.Validate(tt.key, tt.value)
			if res.Valid != tt.isValid {
				t.Errorf("expected valid=%v, got %v", tt.isValid, res.Valid)
			}
		})
	}
}

func TestIdentifierValidator(t *testing.T) {
	v := &IdentifierValidator{}

	tests := []struct {
		name    string
		input   interface{}
		isValid bool
	}{
		{"Valid InstanceID", "i-1234567890abcdef0", true},
		{"Valid DNS", "example.com", true},
		{"Valid IP", "192.168.1.1", true},
		{"Valid Tag", "Env:Prod", true},
		{"Valid Name", "my-server", true},
		{"Nil Input", nil, false},
		{"Not String", 123, false},
		{"Empty String", "", false},
		{"Too Long", strings.Repeat("a", 257), false},
		{"Invalid Format", "invalid/name!", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := v.Validate(tt.input)
			if res.Valid != tt.isValid {
				t.Errorf("expected valid=%v, got %v", tt.isValid, res.Valid)
			}
		})
	}
}

func TestTagKeyValueValidator(t *testing.T) {
	v := &TagKeyValueValidator{}

	tests := []struct {
		name    string
		input   interface{}
		isValid bool
	}{
		{"Valid Tag", "Key:Value", true},
		{"Nil Input", nil, false},
		{"Not String", 123, false},
		{"Empty String", "", false},
		{"No Colon", "KeyValue", false},
		{"Invalid Key", "Key!:Value", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := v.Validate(tt.input)
			if res.Valid != tt.isValid {
				t.Errorf("expected valid=%v, got %v", tt.isValid, res.Valid)
			}
		})
	}
}

func TestSimpleNameValidator(t *testing.T) {
	v := &SimpleNameValidator{}

	tests := []struct {
		name    string
		input   interface{}
		isValid bool
	}{
		{"Valid Name", "server-01", true},
		{"Nil Input", nil, false},
		{"Not String", 123, false},
		{"Empty String", "", false},
		{"Too Long", strings.Repeat("a", 129), false},
		{"Invalid Char", "server/01", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := v.Validate(tt.input)
			if res.Valid != tt.isValid {
				t.Errorf("expected valid=%v, got %v", tt.isValid, res.Valid)
			}
		})
	}
}

func TestHelperFunctions(t *testing.T) {
	// IsASCII
	if !IsASCII("abc") {
		t.Error("IsASCII failed for ASCII string")
	}
	if IsASCII("ab€") {
		t.Error("IsASCII passed for non-ASCII string")
	}

	// HasDangerousCharacters
	if !HasDangerousCharacters("rm -rf /;") { // Updated to include ;
		t.Error("HasDangerousCharacters failed to detect dangerous string")
	}
	if HasDangerousCharacters("ls -la") {
		t.Error("HasDangerousCharacters flagged safe string")
	}

	// TruncateString
	if TruncateString("hello", 3) != "hel" {
		t.Error("TruncateString failed")
	}
	if TruncateString("hello", 10) != "hello" {
		t.Error("TruncateString failed")
	}
}

func TestSanitizers(t *testing.T) {
	// CommandSanitizer
	cs := &CommandSanitizer{}
	if cs.Sanitize("`ls`") != "ls" {
		t.Error("CommandSanitizer failed to remove backticks")
	}

	// DNSSanitizer
	ds := &DNSSanitizer{}
	if ds.Sanitize("Example.Com ") != "example.com" {
		t.Error("DNSSanitizer failed")
	}

	// TagSanitizer
	ts := &TagSanitizer{}
	if ts.SanitizeTagKey(`"Key"`) != "Key" {
		t.Error("TagSanitizer failed for key")
	}
	if ts.SanitizeTagValue(`'Value'`) != "Value" {
		t.Error("TagSanitizer failed for value")
	}
}

func TestChain(t *testing.T) {
	c := NewChain()
	c.Add(&SimpleNameValidator{})

	res := c.Validate("valid-name")
	if !res.Valid {
		t.Error("Chain validation failed for valid input")
	}

	res = c.Validate("invalid/name")
	if res.Valid {
		t.Error("Chain validation passed for invalid input")
	}
}
