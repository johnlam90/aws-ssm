package aws

import (
	"testing"
)

func TestIsInstanceID(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"valid instance ID", "i-1234567890abcdef0", true},
		{"valid short instance ID", "i-12345678", true},
		{"invalid - no prefix", "1234567890abcdef0", false},
		{"invalid - too short", "i-123", false},
		{"invalid - wrong prefix", "j-1234567890abcdef0", false},
		{"empty string", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsInstanceID(tt.input)
			if result != tt.expected {
				t.Errorf("IsInstanceID(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestIsTag(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"valid tag", "Environment:production", true},
		{"valid tag with spaces", "Environment: production", true},
		{"valid tag with multiple colons", "Key:Value:Extra", true},
		{"invalid - no colon", "Environment", false},
		{"invalid - empty key", ":production", false},
		{"invalid - empty value", "Environment:", false},
		{"invalid - only colon", ":", false},
		{"empty string", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsTag(tt.input)
			if result != tt.expected {
				t.Errorf("IsTag(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestIsIPAddress(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"valid IPv4", "192.168.1.1", true},
		{"valid IPv4 - private", "10.0.1.100", true},
		{"valid IPv4 - public", "54.123.45.67", true},
		{"valid IPv6", "2001:0db8:85a3:0000:0000:8a2e:0370:7334", true},
		{"valid IPv6 - short", "2001:db8::1", true},
		{"invalid - not an IP", "not-an-ip", false},
		{"invalid - partial IP", "192.168.1", false},
		{"invalid - too many octets", "192.168.1.1.1", false},
		{"empty string", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsIPAddress(tt.input)
			if result != tt.expected {
				t.Errorf("IsIPAddress(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestIsDNSName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"valid AWS public DNS", "ec2-54-123-45-67.us-west-2.compute.amazonaws.com", true},
		{"valid AWS private DNS", "ip-10-0-1-100.us-west-2.compute.internal", true},
		{"valid EC2 internal DNS", "ip-100-64-149-165.ec2.internal", true},
		{"valid generic DNS", "server.example.com", true},
		{"valid subdomain", "api.prod.example.com", true},
		{"invalid - single dot", "server.local", false},
		{"invalid - no dots", "localhost", false},
		{"invalid - IP address", "192.168.1.1", false},
		{"empty string", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsDNSName(tt.input)
			if result != tt.expected {
				t.Errorf("IsDNSName(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestParseIdentifier(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		expectedType IdentifierType
		expectedKey  string
		expectedVal  string
	}{
		{
			name:         "instance ID",
			input:        "i-1234567890abcdef0",
			expectedType: IdentifierTypeInstanceID,
		},
		{
			name:         "tag",
			input:        "Environment:production",
			expectedType: IdentifierTypeTag,
			expectedKey:  "Environment",
			expectedVal:  "production",
		},
		{
			name:         "tag with spaces",
			input:        "Environment: production ",
			expectedType: IdentifierTypeTag,
			expectedKey:  "Environment",
			expectedVal:  "production",
		},
		{
			name:         "IPv4 address",
			input:        "10.0.1.100",
			expectedType: IdentifierTypeIPAddress,
		},
		{
			name:         "AWS public DNS",
			input:        "ec2-54-123-45-67.us-west-2.compute.amazonaws.com",
			expectedType: IdentifierTypeDNSName,
		},
		{
			name:         "AWS private DNS",
			input:        "ip-10-0-1-100.us-west-2.compute.internal",
			expectedType: IdentifierTypeDNSName,
		},
		{
			name:         "instance name",
			input:        "web-server",
			expectedType: IdentifierTypeName,
		},
		{
			name:         "instance name with spaces",
			input:        " web-server ",
			expectedType: IdentifierTypeName,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseIdentifier(tt.input)
			if result.Type != tt.expectedType {
				t.Errorf("ParseIdentifier(%q).Type = %v, want %v", tt.input, result.Type, tt.expectedType)
			}
			if tt.expectedKey != "" && result.TagKey != tt.expectedKey {
				t.Errorf("ParseIdentifier(%q).TagKey = %q, want %q", tt.input, result.TagKey, tt.expectedKey)
			}
			if tt.expectedVal != "" && result.TagValue != tt.expectedVal {
				t.Errorf("ParseIdentifier(%q).TagValue = %q, want %q", tt.input, result.TagValue, tt.expectedVal)
			}
		})
	}
}

func TestIdentifierTypeString(t *testing.T) {
	tests := []struct {
		idType   IdentifierType
		expected string
	}{
		{IdentifierTypeInstanceID, "instance ID"},
		{IdentifierTypeTag, "tag"},
		{IdentifierTypeIPAddress, "IP address"},
		{IdentifierTypeDNSName, "DNS name"},
		{IdentifierTypeName, "instance name"},
		{IdentifierTypeUnknown, "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := tt.idType.String()
			if result != tt.expected {
				t.Errorf("IdentifierType(%d).String() = %q, want %q", tt.idType, result, tt.expected)
			}
		})
	}
}
