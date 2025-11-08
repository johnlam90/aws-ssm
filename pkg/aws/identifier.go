package aws

import (
	"net"
	"strings"
)

// IdentifierType represents the type of instance identifier
type IdentifierType int

const (
	IdentifierTypeUnknown IdentifierType = iota
	IdentifierTypeInstanceID
	IdentifierTypeTag
	IdentifierTypeIPAddress
	IdentifierTypeDNSName
	IdentifierTypeName
)

// IdentifierInfo contains information about a parsed identifier
type IdentifierInfo struct {
	Type     IdentifierType
	Value    string
	TagKey   string
	TagValue string
}

// ParseIdentifier determines the type of identifier and extracts relevant information
func ParseIdentifier(identifier string) IdentifierInfo {
	// Trim whitespace
	identifier = strings.TrimSpace(identifier)

	// Check for instance ID (starts with i-)
	if IsInstanceID(identifier) {
		return IdentifierInfo{
			Type:  IdentifierTypeInstanceID,
			Value: identifier,
		}
	}

	// Check for tag format (Key:Value)
	if IsTag(identifier) {
		parts := strings.SplitN(identifier, ":", 2)
		return IdentifierInfo{
			Type:     IdentifierTypeTag,
			Value:    identifier,
			TagKey:   strings.TrimSpace(parts[0]),
			TagValue: strings.TrimSpace(parts[1]),
		}
	}

	// Check for IP address
	if IsIPAddress(identifier) {
		return IdentifierInfo{
			Type:  IdentifierTypeIPAddress,
			Value: identifier,
		}
	}

	// Check for DNS name
	if IsDNSName(identifier) {
		return IdentifierInfo{
			Type:  IdentifierTypeDNSName,
			Value: identifier,
		}
	}

	// Default to name tag
	return IdentifierInfo{
		Type:  IdentifierTypeName,
		Value: identifier,
	}
}

// IsInstanceID checks if the string is an EC2 instance ID
func IsInstanceID(s string) bool {
	return strings.HasPrefix(s, "i-") && len(s) >= 10
}

// IsTag checks if the string is in tag format (Key:Value)
func IsTag(s string) bool {
	if !strings.Contains(s, ":") {
		return false
	}
	parts := strings.SplitN(s, ":", 2)
	return len(parts) == 2 && len(strings.TrimSpace(parts[0])) > 0 && len(strings.TrimSpace(parts[1])) > 0
}

// IsIPAddress checks if the string is a valid IP address
func IsIPAddress(s string) bool {
	return net.ParseIP(s) != nil
}

// IsDNSName checks if the string looks like a DNS name
func IsDNSName(s string) bool {
	// Must contain at least one dot
	if !strings.Contains(s, ".") {
		return false
	}

	// Exclude IP addresses (they also contain dots)
	if IsIPAddress(s) {
		return false
	}

	// Check for AWS-specific DNS patterns
	if strings.Contains(s, "compute.amazonaws.com") ||
		strings.Contains(s, "compute.internal") ||
		strings.Contains(s, ".ec2.internal") {
		return true
	}

	// Generic DNS name check: at least 2 dots (e.g., host.domain.com)
	return strings.Count(s, ".") >= 2
}

// String returns a human-readable description of the identifier type
func (t IdentifierType) String() string {
	switch t {
	case IdentifierTypeInstanceID:
		return "instance ID"
	case IdentifierTypeTag:
		return "tag"
	case IdentifierTypeIPAddress:
		return "IP address"
	case IdentifierTypeDNSName:
		return "DNS name"
	case IdentifierTypeName:
		return "instance name"
	default:
		return "unknown"
	}
}
