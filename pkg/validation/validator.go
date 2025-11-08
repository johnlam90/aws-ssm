package validation

import (
	"context"
	"fmt"
	"net"
	"regexp"
	"strings"
)

// Result represents the result of validation
type Result struct {
	Valid  bool
	Errors []string
	Fields map[string]interface{}
}

// NewResult creates a new validation result
func NewResult() *Result {
	return &Result{
		Valid:  true,
		Errors: make([]string, 0),
		Fields: make(map[string]interface{}),
	}
}

// AddError adds an error to the result
func (r *Result) AddError(field, message string) {
	r.Valid = false
	r.Errors = append(r.Errors, fmt.Sprintf("%s: %s", field, message))
}

// AddField adds a validated field
func (r *Result) AddField(name string, value interface{}) {
	r.Fields[name] = value
}

// Validator interface for validation
type Validator interface {
	Validate(value interface{}) *Result
}

// Common validation patterns
var (
	// Instance ID pattern: i- followed by 17 hex characters
	instanceIDPattern = regexp.MustCompile(`^i-[0-9a-f]{17}$`)

	// DNS name pattern
	dnsPattern = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9-]*(\.[a-zA-Z0-9][a-zA-Z0-9-]*)*$`)

	// AWS region pattern
	regionPattern = regexp.MustCompile(`^[a-z]{2}-[a-z]+-\d+$`)

	// AWS profile pattern
	profilePattern = regexp.MustCompile(`^[a-zA-Z0-9._-]+$`)

	// Safe command pattern (alphanumeric, common symbols, no dangerous characters)
	safeCommandPattern = regexp.MustCompile(`^[a-zA-Z0-9_\-\s\|\.\/\\:@]+$`)

	// Dangerous command patterns to reject
	dangerousPatterns = []*regexp.Regexp{
		regexp.MustCompile(`[\$\(\)]+`), // Command substitution
		regexp.MustCompile(`;\s*rm\s+`), // rm command injection
		regexp.MustCompile(`\|\|\s*`),   // Command chaining
		regexp.MustCompile(`&&\s*`),     // Command chaining
		regexp.MustCompile(`>\s*\/`),    // Output redirection
		regexp.MustCompile(`<\s*\/`),    // Input redirection
		regexp.MustCompile(`\$\{`),      // Variable expansion
		regexp.MustCompile(`\$\w+`),     // Variable expansion
		regexp.MustCompile(`\` + "`"),   // Backtick command execution
	}
)

// InstanceIDValidator validates AWS instance IDs
type InstanceIDValidator struct{}

// Validate validates an instance ID
func (v *InstanceIDValidator) Validate(value interface{}) *Result {
	result := NewResult()

	if value == nil {
		result.AddError("instance_id", "cannot be nil")
		return result
	}

	instanceID, ok := value.(string)
	if !ok {
		result.AddError("instance_id", "must be a string")
		return result
	}

	instanceID = strings.TrimSpace(instanceID)
	if instanceID == "" {
		result.AddError("instance_id", "cannot be empty")
		return result
	}

	if len(instanceID) > 19 {
		result.AddError("instance_id", "too long (max 19 characters)")
		return result
	}

	if !instanceIDPattern.MatchString(instanceID) {
		result.AddError("instance_id", "invalid format (should be i-xxxxxxxxxxxxxxxxx)")
		return result
	}

	result.AddField("instance_id", instanceID)
	return result
}

// DNSNameValidator validates DNS names
type DNSNameValidator struct{}

// Validate validates a DNS name
func (v *DNSNameValidator) Validate(value interface{}) *Result {
	result := NewResult()

	if value == nil {
		result.AddError("dns_name", "cannot be nil")
		return result
	}

	dnsName, ok := value.(string)
	if !ok {
		result.AddError("dns_name", "must be a string")
		return result
	}

	dnsName = strings.TrimSpace(dnsName)
	if dnsName == "" {
		result.AddError("dns_name", "cannot be empty")
		return result
	}

	if len(dnsName) > 255 {
		result.AddError("dns_name", "too long (max 255 characters)")
		return result
	}

	if !dnsPattern.MatchString(dnsName) {
		result.AddError("dns_name", "invalid DNS name format")
		return result
	}

	result.AddField("dns_name", dnsName)
	return result
}

// IPAddressValidator validates IP addresses
type IPAddressValidator struct{}

// Validate validates an IP address
func (v *IPAddressValidator) Validate(value interface{}) *Result {
	result := NewResult()

	if value == nil {
		result.AddError("ip_address", "cannot be nil")
		return result
	}

	ipStr, ok := value.(string)
	if !ok {
		result.AddError("ip_address", "must be a string")
		return result
	}

	ipStr = strings.TrimSpace(ipStr)
	if ipStr == "" {
		result.AddError("ip_address", "cannot be empty")
		return result
	}

	ip := net.ParseIP(ipStr)
	if ip == nil {
		result.AddError("ip_address", "invalid IP address format")
		return result
	}

	result.AddField("ip_address", ipStr)
	return result
}

// RegionValidator validates AWS region names
type RegionValidator struct{}

// Validate validates a region
func (v *RegionValidator) Validate(value interface{}) *Result {
	result := NewResult()

	if value == nil {
		result.AddError("region", "cannot be nil")
		return result
	}

	region, ok := value.(string)
	if !ok {
		result.AddError("region", "must be a string")
		return result
	}

	region = strings.TrimSpace(region)
	if region == "" {
		result.AddError("region", "cannot be empty")
		return result
	}

	if len(region) > 32 {
		result.AddError("region", "too long")
		return result
	}

	// Common valid regions (subset for validation)
	validRegions := map[string]bool{
		"us-east-1": true, "us-east-2": true, "us-west-1": true, "us-west-2": true,
		"eu-west-1": true, "eu-west-2": true, "eu-west-3": true, "eu-central-1": true,
		"ap-southeast-1": true, "ap-southeast-2": true, "ap-northeast-1": true,
		"ap-northeast-2": true, "ap-south-1": true, "ca-central-1": true,
		"sa-east-1": true, "us-gov-west-1": true, "us-gov-east-1": true,
	}

	if !regionPattern.MatchString(region) && !validRegions[region] {
		result.AddError("region", "invalid AWS region format")
		return result
	}

	result.AddField("region", region)
	return result
}

// ProfileValidator validates AWS profile names
type ProfileValidator struct{}

// Validate validates a profile
func (v *ProfileValidator) Validate(value interface{}) *Result {
	result := NewResult()

	if value == nil {
		result.AddError("profile", "cannot be nil")
		return result
	}

	profile, ok := value.(string)
	if !ok {
		result.AddError("profile", "must be a string")
		return result
	}

	profile = strings.TrimSpace(profile)
	if profile == "" {
		result.AddError("profile", "cannot be empty")
		return result
	}

	if len(profile) > 128 {
		result.AddError("profile", "too long (max 128 characters)")
		return result
	}

	if !profilePattern.MatchString(profile) {
		result.AddError("profile", "contains invalid characters")
		return result
	}

	result.AddField("profile", profile)
	return result
}

// CommandValidator validates remote commands
type CommandValidator struct {
	MaxLength int
}

// NewCommandValidator creates a new command validator
func NewCommandValidator() *CommandValidator {
	return &CommandValidator{
		MaxLength: 1024, // Reasonable limit for commands
	}
}

// Validate validates a command
func (v *CommandValidator) Validate(value interface{}) *Result {
	result := NewResult()

	if value == nil {
		result.AddError("command", "cannot be nil")
		return result
	}

	command, ok := value.(string)
	if !ok {
		result.AddError("command", "must be a string")
		return result
	}

	command = strings.TrimSpace(command)
	if command == "" {
		result.AddError("command", "cannot be empty")
		return result
	}

	if len(command) > v.MaxLength {
		result.AddError("command", fmt.Sprintf("too long (max %d characters)", v.MaxLength))
		return result
	}

	// Check for dangerous patterns
	for _, pattern := range dangerousPatterns {
		if pattern.MatchString(command) {
			result.AddError("command", "contains dangerous characters or patterns")
			return result
		}
	}

	// Verify command is safe
	if !safeCommandPattern.MatchString(command) {
		result.AddError("command", "contains unsafe characters")
		return result
	}

	// Additional check: ensure no null bytes or control characters
	for _, r := range command {
		if r == 0 || (r < 32 && r != '\t' && r != '\n' && r != '\r') {
			result.AddError("command", "contains control characters")
			return result
		}
		if r > 127 {
			result.AddError("command", "contains non-ASCII characters")
			return result
		}
	}

	result.AddField("command", command)
	return result
}

// TagValidator validates AWS tags
type TagValidator struct{}

// Validate validates a tag key and value
func (v *TagValidator) Validate(key, value interface{}) *Result {
	result := NewResult()

	// Validate key
	if key == nil {
		result.AddError("tag_key", "cannot be nil")
		return result
	}

	keyStr, ok := key.(string)
	if !ok {
		result.AddError("tag_key", "must be a string")
		return result
	}

	keyStr = strings.TrimSpace(keyStr)
	if keyStr == "" {
		result.AddError("tag_key", "cannot be empty")
		return result
	}

	if len(keyStr) > 128 {
		result.AddError("tag_key", "too long (max 128 characters)")
		return result
	}

	// Tag key must match pattern: [a-zA-Z0-9_.\-\/]+
	if !regexp.MustCompile(`^[a-zA-Z0-9_.\-\/]+$`).MatchString(keyStr) {
		result.AddError("tag_key", "contains invalid characters")
		return result
	}

	// Validate value
	if value == nil {
		result.AddError("tag_value", "cannot be nil")
		return result
	}

	valueStr, ok := value.(string)
	if !ok {
		result.AddError("tag_value", "must be a string")
		return result
	}

	valueStr = strings.TrimSpace(valueStr)
	if len(valueStr) > 256 {
		result.AddError("tag_value", "too long (max 256 characters)")
		return result
	}

	result.AddField("tag_key", keyStr)
	result.AddField("tag_value", valueStr)
	return result
}

// IdentifierValidator validates instance identifiers
type IdentifierValidator struct{}

// Validate validates an identifier
func (v *IdentifierValidator) Validate(value interface{}) *Result {
	result := NewResult()

	if value == nil {
		result.AddError("identifier", "cannot be nil")
		return result
	}

	identifier, ok := value.(string)
	if !ok {
		result.AddError("identifier", "must be a string")
		return result
	}

	identifier = strings.TrimSpace(identifier)
	if identifier == "" {
		result.AddError("identifier", "cannot be empty")
		return result
	}

	if len(identifier) > 256 {
		result.AddError("identifier", "too long (max 256 characters)")
		return result
	}

	// Try to determine the type and validate accordingly
	identifierTypes := []Validator{
		&InstanceIDValidator{},
		&DNSNameValidator{},
		&IPAddressValidator{},
		&TagKeyValueValidator{},
		&SimpleNameValidator{},
	}

	for _, validator := range identifierTypes {
		if validationResult := validator.Validate(identifier); validationResult.Valid {
			return validationResult
		}
	}

	result.AddError("identifier", "invalid format")
	return result
}

// TagKeyValueValidator validates tag-style identifiers
type TagKeyValueValidator struct{}

// Validate validates a tag key-value pair
func (v *TagKeyValueValidator) Validate(value interface{}) *Result {
	result := NewResult()

	if value == nil {
		result.AddError("tag_identifier", "cannot be nil")
		return result
	}

	tagStr, ok := value.(string)
	if !ok {
		result.AddError("tag_identifier", "must be a string")
		return result
	}

	tagStr = strings.TrimSpace(tagStr)
	if tagStr == "" {
		result.AddError("tag_identifier", "cannot be empty")
		return result
	}

	// Must contain exactly one colon
	parts := strings.SplitN(tagStr, ":", 2)
	if len(parts) != 2 {
		result.AddError("tag_identifier", "must be in format Key:Value")
		return result
	}

	key, value := parts[0], parts[1]

	// Validate key and value
	tagValidator := &TagValidator{}
	if validationResult := tagValidator.Validate(key, value); !validationResult.Valid {
		return validationResult
	}

	result.AddField("tag_key", key)
	result.AddField("tag_value", value)
	result.AddField("identifier_type", "tag")
	return result
}

// SimpleNameValidator validates simple names
type SimpleNameValidator struct{}

// Validate validates a simple name
func (v *SimpleNameValidator) Validate(value interface{}) *Result {
	result := NewResult()

	if value == nil {
		result.AddError("name", "cannot be nil")
		return result
	}

	name, ok := value.(string)
	if !ok {
		result.AddError("name", "must be a string")
		return result
	}

	name = strings.TrimSpace(name)
	if name == "" {
		result.AddError("name", "cannot be empty")
		return result
	}

	if len(name) > 128 {
		result.AddError("name", "too long (max 128 characters)")
		return result
	}

	// Check for basic invalid characters
	invalidChars := []rune{'/', '\\', '?', '*', '<', '>', '|', ':', '"'}
	for _, char := range invalidChars {
		if strings.ContainsRune(name, char) {
			result.AddError("name", fmt.Sprintf("contains invalid character: %c", char))
			return result
		}
	}

	result.AddField("name", name)
	result.AddField("identifier_type", "name")
	return result
}

// Sanitizer interface for input sanitization
type Sanitizer interface {
	Sanitize(value string) string
}

// CommandSanitizer sanitizes commands
type CommandSanitizer struct{}

// Sanitize sanitizes a command string
func (s *CommandSanitizer) Sanitize(command string) string {
	// Remove dangerous characters
	command = strings.ReplaceAll(command, "`", "")
	command = strings.ReplaceAll(command, "\\", "")

	// Remove leading/trailing whitespace
	command = strings.TrimSpace(command)

	// Limit length
	if len(command) > 1024 {
		command = command[:1024]
	}

	return command
}

// DNSSanitizer sanitizes DNS names
type DNSSanitizer struct{}

// Sanitize sanitizes a DNS name
func (s *DNSSanitizer) Sanitize(dnsName string) string {
	// Convert to lowercase
	dnsName = strings.ToLower(dnsName)

	// Remove whitespace
	dnsName = strings.ReplaceAll(dnsName, " ", "")

	return dnsName
}

// TagSanitizer sanitizes tag keys and values
type TagSanitizer struct{}

// SanitizeTagKey sanitizes a tag key
func (s *TagSanitizer) SanitizeTagKey(key string) string {
	// Trim whitespace
	key = strings.TrimSpace(key)

	// Remove quotes
	key = strings.Trim(key, `"'`)

	return key
}

// SanitizeTagValue sanitizes a tag value
func (s *TagSanitizer) SanitizeTagValue(value string) string {
	// Trim whitespace
	value = strings.TrimSpace(value)

	// Remove quotes
	value = strings.Trim(value, `"'`)

	return value
}

// Chain allows chaining multiple validators
type Chain struct {
	validators []Validator
}

// NewChain creates a new validation chain
func NewChain() *Chain {
	return &Chain{
		validators: make([]Validator, 0),
	}
}

// Add adds a validator to the chain
func (vc *Chain) Add(validator Validator) *Chain {
	vc.validators = append(vc.validators, validator)
	return vc
}

// Validate validates using all validators in the chain
func (vc *Chain) Validate(value interface{}) *Result {
	result := NewResult()

	for _, validator := range vc.validators {
		validationResult := validator.Validate(value)
		if !validationResult.Valid {
			result.Errors = append(result.Errors, validationResult.Errors...)
		}
		result.Fields = mergeMaps(result.Fields, validationResult.Fields)
	}

	return result
}

func mergeMaps(dest, src map[string]interface{}) map[string]interface{} {
	if dest == nil {
		dest = make(map[string]interface{})
	}
	for k, v := range src {
		dest[k] = v
	}
	return dest
}

// ValidateWithContext validates with context awareness
func ValidateWithContext(ctx context.Context, value interface{}, validator Validator) *Result {
	// Add validation to context for logging/monitoring
	// This can be extended to include request IDs, user context, etc.
	return validator.Validate(value)
}

// IsASCII checks if a string contains only ASCII characters
func IsASCII(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] > 127 {
			return false
		}
	}
	return true
}

// HasDangerousCharacters checks if a string contains dangerous characters
func HasDangerousCharacters(s string) bool {
	dangerous := []rune{';', '&', '|', '<', '>', '$', '`', '\\', '"', '\'', '(', ')', '[', ']'}
	for _, char := range dangerous {
		if strings.ContainsRune(s, char) {
			return true
		}
	}
	return false
}

// TruncateString truncates a string to the maximum length
func TruncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen]
}

// Default validators
var (
	InstanceID = &InstanceIDValidator{}
	DNSName    = &DNSNameValidator{}
	IPAddress  = &IPAddressValidator{}
	Region     = &RegionValidator{}
	Profile    = &ProfileValidator{}
	Command    = NewCommandValidator()
	Tag        = &TagValidator{}
	Identifier = &IdentifierValidator{}

	CommandSanit = &CommandSanitizer{}
	DNSSanit     = &DNSSanitizer{}
	TagSanit     = &TagSanitizer{}
)
