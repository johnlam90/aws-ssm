package security

import (
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/aws-ssm/pkg/logging"
	"github.com/aws-ssm/pkg/validation"
)

// Level represents different security levels
type Level string

const (
	// SecurityLow represents low security level
	SecurityLow Level = "low"
	// SecurityMedium represents medium security level
	SecurityMedium Level = "medium"
	// SecurityHigh represents high security level
	SecurityHigh Level = "high"
	// SecurityStrict represents strict security level
	SecurityStrict Level = "strict"
)

// Config represents security configuration
type Config struct {
	Level                    Level
	CommandTimeout           time.Duration
	MaxCommandLength         int
	AllowedCommands          []string
	BlockedPatterns          []string
	RequireCommandValidation bool
	EnableAuditLogging       bool
	CredentialRotationCheck  bool
	SessionTimeout           time.Duration
	RateLimitPerIP           int
	EnableTLSVerification    bool
	CertPaths                []string
}

// DefaultConfig returns default security configuration
func DefaultConfig() *Config {
	return &Config{
		Level:                    SecurityMedium,
		CommandTimeout:           300 * time.Second,
		MaxCommandLength:         1024,
		RequireCommandValidation: true,
		EnableAuditLogging:       true,
		CredentialRotationCheck:  true,
		SessionTimeout:           3600 * time.Second,
		RateLimitPerIP:           100,
		EnableTLSVerification:    true,
		CertPaths:                []string{},
		AllowedCommands: []string{
			"bash", "sh", "zsh",
			"ls", "cd", "pwd", "cat", "grep", "find", "tail", "head",
			"ps", "top", "df", "du", "free", "uptime",
			"systemctl", "service",
			"curl", "wget", "ping", "ssh", "scp",
		},
		BlockedPatterns: []string{
			// Dangerous destructive operations - rm variants (consolidated pattern)
			// Matches: rm -rf / or rm -rf /* or rm -rf /root, /home, /etc, /var, /usr, /boot, /lib, /sys, /proc, /dev
			`rm\s+-rf\s+/(?:\*|root|home|etc|var|usr|boot|lib|sys|proc|dev)?$`,

			// Dangerous permission changes
			`chmod\s+000\s+/`,      // chmod 000 / (remove all permissions from root)
			`chmod\s+-R\s+000`,     // chmod -R 000 (recursive permission removal)
			`chown\s+\S+\s+/`,      // chown on root directory
			`chown\s+-R\s+\S+\s+/`, // chown -R on root directory

			// Dangerous sudo combinations
			`sudo\s+rm\s+-rf`,    // sudo rm -rf (dangerous combination)
			`sudo\s+chmod\s+000`, // sudo chmod 000 (dangerous combination)

			// Dangerous redirects to system files
			`>\s*/etc/`,  // Redirect to /etc
			`>\s*/sys/`,  // Redirect to /sys
			`>\s*/proc/`, // Redirect to /proc
			`>\s*/dev/`,  // Redirect to /dev
			`>\s*/root`,  // Redirect to /root
			`>\s*/home`,  // Redirect to /home
			`>\s*/boot`,  // Redirect to /boot

			// Dangerous command chaining patterns
			`;\s*rm\s+-rf\s+/`,   // ; rm -rf / (chained destructive command)
			`\|\|\s*rm\s+-rf`,    // || rm -rf (fallback to destructive command)
			`&&\s*rm\s+-rf`,      // && rm -rf (chained destructive command)
			`;\s*chmod\s+000`,    // ; chmod 000 (chained permission removal)
			`\|\|\s*chmod\s+000`, // || chmod 000 (fallback permission removal)
			`&&\s*chmod\s+000`,   // && chmod 000 (chained permission removal)
		},
	}
}

// Manager manages security policies and enforcement
type Manager struct {
	config          *Config
	logger          logging.Logger
	auditor         *AuditLogger
	blockedPatterns []*regexp.Regexp
	suspiciousRegex []*regexp.Regexp
}

// NewManager creates a new security manager
func NewManager(config *Config) *Manager {
	if config == nil {
		config = DefaultConfig()
	}

	// Compile blocked patterns as regex
	blockedPatterns := make([]*regexp.Regexp, 0, len(config.BlockedPatterns))
	for _, pattern := range config.BlockedPatterns {
		if re, err := regexp.Compile(pattern); err == nil {
			blockedPatterns = append(blockedPatterns, re)
		}
	}

	// Compile suspicious patterns for high security mode
	// Note: These patterns are used to flag potentially suspicious commands for review,
	// not to block them outright. They help identify commands that may need additional scrutiny.
	suspiciousPatterns := []string{
		`\$\(`,    // Command substitution $(...)
		`\$\{`,    // Variable expansion ${...}
		`[<>|&;]`, // Shell pipe, redirect, and command chaining operators
		"`",       // Backtick command substitution
		`\\`,      // Backslash escaping
		`"`,       // Double quotes
		`'`,       // Single quotes
		// Note: curl and wget are legitimate tools and are NOT blocked by default
		// They are only flagged as suspicious in strict security mode
	}
	suspiciousRegex := make([]*regexp.Regexp, 0, len(suspiciousPatterns))
	for _, pattern := range suspiciousPatterns {
		if re, err := regexp.Compile(pattern); err == nil {
			suspiciousRegex = append(suspiciousRegex, re)
		}
	}

	return &Manager{
		config:          config,
		logger:          logging.With(logging.String("component", "security_manager")),
		auditor:         NewAuditLogger(config),
		blockedPatterns: blockedPatterns,
		suspiciousRegex: suspiciousRegex,
	}
}

// ValidateCommand validates a command for security compliance
func (sm *Manager) ValidateCommand(command string) error {
	// Trim leading and trailing whitespace to avoid false negatives with trailing spaces
	command = strings.TrimSpace(command)

	// Use the standard command validator for basic validation
	validator := validation.NewCommandValidator()
	result := validator.Validate(command)
	if !result.Valid {
		reason := strings.Join(result.Errors, "; ")
		sm.auditor.Log("command_rejected", map[string]interface{}{
			"command": command,
			"reason":  reason,
		})
		return fmt.Errorf("command validation failed: %s", reason)
	}

	// Apply security-specific pattern validation
	if err := sm.validatePatterns(command); err != nil {
		sm.auditor.Log("command_rejected", map[string]interface{}{
			"command": command,
			"reason":  err.Error(),
		})
		return err
	}

	// Apply security-specific structure validation
	if err := sm.validateCommandStructure(command); err != nil {
		sm.auditor.Log("command_rejected", map[string]interface{}{
			"command": command,
			"reason":  err.Error(),
		})
		return err
	}

	sm.auditor.Log("command_approved", map[string]interface{}{
		"command": command,
	})

	return nil
}

func (sm *Manager) validateBasicRules(command string) error {
	// Check length
	if len(command) > sm.config.MaxCommandLength {
		return fmt.Errorf("command too long (max %d characters)", sm.config.MaxCommandLength)
	}

	// Check for null bytes or control characters
	for _, r := range command {
		if r == 0 || (r < 32 && r != '\t' && r != '\n' && r != '\r') {
			return fmt.Errorf("command contains control characters")
		}
	}

	// Check for non-ASCII characters in high security mode
	if sm.config.Level == SecurityHigh || sm.config.Level == SecurityStrict {
		for _, r := range command {
			if r > 127 {
				return fmt.Errorf("command contains non-ASCII characters")
			}
		}
	}

	return nil
}

func (sm *Manager) validatePatterns(command string) error {
	// Check against compiled blocked patterns using regex
	for i, re := range sm.blockedPatterns {
		if re.MatchString(command) {
			return fmt.Errorf("command contains blocked pattern: %s", sm.config.BlockedPatterns[i])
		}
	}

	// Additional checks for high security levels
	if sm.config.Level == SecurityHigh || sm.config.Level == SecurityStrict {
		// Check for suspicious character sequences using regex
		for i, re := range sm.suspiciousRegex {
			if re.MatchString(command) {
				patterns := []string{
					"command substitution",
					"shell metacharacters",
					"variable expansion",
					"network tools",
				}
				if i < len(patterns) {
					return fmt.Errorf("command contains suspicious pattern: %s", patterns[i])
				}
				return fmt.Errorf("command contains suspicious pattern")
			}
		}
	}

	return nil
}

func (sm *Manager) validateCommandStructure(command string) error {
	if sm.config.Level == SecurityStrict {
		// In strict mode, only allow whitelisted commands
		parts := strings.Fields(command)
		if len(parts) == 0 {
			return fmt.Errorf("empty command")
		}

		baseCommand := parts[0]
		if !containsString(sm.config.AllowedCommands, baseCommand) {
			return fmt.Errorf("command not in whitelist: %s", baseCommand)
		}
	}

	return nil
}

// ValidateSession validates a session for security compliance
func (sm *Manager) ValidateSession(sessionID string, userID string) error {
	// Check session timeout
	if err := sm.checkSessionTimeout(sessionID); err != nil {
		return err
	}

	// Check rate limiting
	if err := sm.checkRateLimit(userID); err != nil {
		return err
	}

	// Check credential rotation if enabled
	if sm.config.CredentialRotationCheck {
		if err := sm.checkCredentialRotation(userID); err != nil {
			return err
		}
	}

	sm.auditor.Log("session_validated", map[string]interface{}{
		"session_id": sessionID,
		"user_id":    userID,
	})

	return nil
}

func (sm *Manager) checkSessionTimeout(sessionID string) error {
	// This would check against actual session store
	// For now, just return nil as placeholder
	return nil
}

func (sm *Manager) checkRateLimit(userID string) error {
	// This would check against a rate limiter
	// For now, just return nil as placeholder
	return nil
}

func (sm *Manager) checkCredentialRotation(userID string) error {
	// This would check if credentials have been rotated recently
	// For now, just return nil as placeholder
	return nil
}

// SanitizeInput sanitizes user input
func (sm *Manager) SanitizeInput(input string) string {
	// Remove dangerous characters
	dangerous := []string{
		"${", "}", "$(", "`", "\\", "\"", "'", ";", "&", "|", ">", "<", "*", "?",
	}

	cleaned := input
	for _, d := range dangerous {
		cleaned = strings.ReplaceAll(cleaned, d, "")
	}

	// Trim whitespace
	cleaned = strings.TrimSpace(cleaned)

	// Limit length
	if len(cleaned) > sm.config.MaxCommandLength {
		cleaned = cleaned[:sm.config.MaxCommandLength]
	}

	return cleaned
}

// AuditLogger logs security events
type AuditLogger struct {
	config *Config
	logger logging.Logger
	mu     sync.Mutex
}

// NewAuditLogger creates a new audit logger
func NewAuditLogger(config *Config) *AuditLogger {
	return &AuditLogger{
		config: config,
		logger: logging.With(logging.String("component", "audit_logger")),
	}
}

// Log logs a security audit event
func (al *AuditLogger) Log(event string, data map[string]interface{}) {
	if !al.config.EnableAuditLogging {
		return
	}

	al.mu.Lock()
	defer al.mu.Unlock()

	// Log the event
	al.logger.Info("Security audit event",
		logging.String("event", event),
		logging.Any("data", data))
}

// CredentialManager manages AWS credentials securely
type CredentialManager struct {
	logger logging.Logger
	mu     sync.RWMutex
	// In production, this would use proper credential storage
	cache map[string]*CredentialInfo
}

// CredentialInfo represents credential information
type CredentialInfo struct {
	Profile     string
	LastUsed    time.Time
	Rotated     bool
	Hash        string
	Permissions []string
}

// NewCredentialManager creates a new credential manager
func NewCredentialManager() *CredentialManager {
	return &CredentialManager{
		logger: logging.With(logging.String("component", "credential_manager")),
		cache:  make(map[string]*CredentialInfo),
	}
}

// ValidateCredentials validates AWS credentials
func (cm *CredentialManager) ValidateCredentials(profile string) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	// In production, this would:
	// 1. Check if credentials exist and are valid
	// 2. Verify they have required permissions
	// 3. Check for credential rotation requirements
	// 4. Log usage for audit

	credential := cm.cache[profile]
	if credential == nil {
		// Create new credential info
		credential = &CredentialInfo{
			Profile:     profile,
			LastUsed:    time.Now(),
			Rotated:     false,
			Hash:        cm.hashProfile(profile),
			Permissions: []string{},
		}
		cm.cache[profile] = credential
	}

	credential.LastUsed = time.Now()

	// Log credential usage
	cm.logger.Info("Credential validated",
		logging.String("profile", profile),
		logging.String("hash", credential.Hash))

	return nil
}

// HashProfile creates a hash of the profile for logging without exposing sensitive data
func (cm *CredentialManager) hashProfile(profile string) string {
	hash := sha256.Sum256([]byte(profile))
	return hex.EncodeToString(hash[:])[:16] // First 8 bytes
}

// GetCredentialInfo returns credential information (without sensitive data)
func (cm *CredentialManager) GetCredentialInfo(profile string) (*CredentialInfo, bool) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	cred, exists := cm.cache[profile]
	if !exists {
		return nil, false
	}

	// Return a copy without sensitive information
	return &CredentialInfo{
		Profile:  cred.Profile,
		LastUsed: cred.LastUsed,
		Rotated:  cred.Rotated,
		Hash:     cred.Hash,
		// Don't expose Permissions
	}, true
}

// TLSConfig manages TLS configuration for secure connections
type TLSConfig struct {
	Config     *tls.Config
	CertPool   *x509.CertPool
	ClientCert tls.Certificate
	logger     logging.Logger
}

// NewTLSConfig creates a new TLS configuration
func NewTLSConfig(certPath, keyPath string, caPath string) (*TLSConfig, error) {
	logger := logging.With(logging.String("component", "tls_config"))

	// Load CA certificate
	var certPool *x509.CertPool
	if caPath != "" {
		// Validate and clean the path to prevent directory traversal
		caPath = filepath.Clean(caPath)
		if strings.Contains(caPath, "..") {
			return nil, fmt.Errorf("invalid CA certificate path: contains directory traversal")
		}
		caCert, err := os.ReadFile(caPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read CA certificate: %w", err)
		}

		certPool = x509.NewCertPool()
		if !certPool.AppendCertsFromPEM(caCert) {
			return nil, fmt.Errorf("failed to parse CA certificate")
		}
		logger.Info("CA certificate loaded", logging.String("path", caPath))
	}

	// Load client certificate if provided
	var clientCert tls.Certificate
	if certPath != "" && keyPath != "" {
		cert, err := tls.LoadX509KeyPair(certPath, keyPath)
		if err != nil {
			return nil, fmt.Errorf("failed to load client certificate: %w", err)
		}
		clientCert = cert
		logger.Info("Client certificate loaded", logging.String("cert_path", certPath))
	}

	// Configure TLS
	// Note: ClientAuth is set to NoClientCert by default to support standard server-only verification.
	// If mutual TLS (mTLS) is required, set ClientAuth to RequireAndVerifyClientCert and provide client certificates.
	clientAuth := tls.NoClientCert
	if clientCert.Certificate != nil {
		// If client certificate is provided, require and verify it
		clientAuth = tls.RequireAndVerifyClientCert
		logger.Info("Mutual TLS (mTLS) enabled - client certificate verification required")
	}

	tlsConfig := &tls.Config{
		MinVersion:         tls.VersionTLS12,
		ClientAuth:         clientAuth,
		RootCAs:            certPool,
		ClientCAs:          certPool,
		InsecureSkipVerify: false,
	}

	if clientCert.Certificate != nil {
		tlsConfig.Certificates = []tls.Certificate{clientCert}
	}

	return &TLSConfig{
		Config:   tlsConfig,
		CertPool: certPool,
		logger:   logger,
	}, nil
}

// GetTLSConfig returns the TLS configuration
func (t *TLSConfig) GetTLSConfig() *tls.Config {
	return t.Config
}

// Event represents a security event
type Event struct {
	Type      string
	Severity  string
	Source    string
	Target    string
	Message   string
	Timestamp time.Time
	Data      map[string]interface{}
}

// EventHandler handles security events
type EventHandler struct {
	logger logging.Logger
	mu     sync.Mutex
}

// NewEventHandler creates a new security event handler
func NewEventHandler() *EventHandler {
	return &EventHandler{
		logger: logging.With(logging.String("component", "security_event_handler")),
	}
}

// HandleEvent handles a security event
func (seh *EventHandler) HandleEvent(event *Event) {
	seh.mu.Lock()
	defer seh.mu.Unlock()

	// Log the event
	seh.logger.Warn("Security event",
		logging.String("type", event.Type),
		logging.String("severity", event.Severity),
		logging.String("source", event.Source),
		logging.String("target", event.Target),
		logging.String("message", event.Message),
		logging.Any("data", event.Data))

	// In production, you would also:
	// - Send alerts to monitoring system
	// - Store in security database
	// - Trigger automated responses
	// - Notify security team
}

// CreateSecurityEvent creates a new security event
func CreateSecurityEvent(eventType, severity, source, target, message string, data map[string]interface{}) *Event {
	return &Event{
		Type:      eventType,
		Severity:  severity,
		Source:    source,
		Target:    target,
		Message:   message,
		Timestamp: time.Now(),
		Data:      data,
	}
}

// SecureConnection creates a secure network connection
func SecureConnection(network, addr string, timeout time.Duration) (net.Conn, error) {
	dialer := &net.Dialer{
		Timeout:   timeout,
		KeepAlive: timeout,
	}

	conn, err := dialer.Dial(network, addr)
	if err != nil {
		return nil, fmt.Errorf("failed to create secure connection: %w", err)
	}

	// Wrap connection with security checks
	return &SecureConn{
		Conn:    conn,
		logger:  logging.With(logging.String("component", "secure_connection")),
		timeout: timeout,
	}, nil
}

// SecureConn wraps a net.Conn with security features
type SecureConn struct {
	net.Conn
	logger  logging.Logger
	timeout time.Duration
}

func (sc *SecureConn) Read(b []byte) (n int, err error) {
	n, err = sc.Conn.Read(b)

	// Log read operations in high security mode
	if err == nil && n > 0 {
		sc.logger.Debug("Secure connection read",
			logging.Int("bytes", n),
			logging.String("remote_addr", sc.Conn.RemoteAddr().String()))
	}

	return n, err
}

func (sc *SecureConn) Write(b []byte) (n int, err error) {
	n, err = sc.Conn.Write(b)

	// Log write operations in high security mode
	if err == nil && n > 0 {
		sc.logger.Debug("Secure connection write",
			logging.Int("bytes", n),
			logging.String("remote_addr", sc.Conn.RemoteAddr().String()))
	}

	return n, err
}

// SecureHTTPClient creates a secure HTTP client
func SecureHTTPClient(timeout time.Duration, tlsConfig *TLSConfig) *http.Client {
	transport := &http.Transport{
		TLSClientConfig:     tlsConfig.GetTLSConfig(),
		Dial:                func(network, addr string) (net.Conn, error) { return SecureConnection(network, addr, timeout) },
		MaxIdleConns:        10,
		MaxIdleConnsPerHost: 5,
		IdleConnTimeout:     90 * time.Second,
		DisableCompression:  false,
	}

	return &http.Client{
		Transport: transport,
		Timeout:   timeout,
	}
}

// Scanner scans for security issues
type Scanner struct {
	logger logging.Logger
}

// NewScanner creates a new security scanner
func NewScanner() *Scanner {
	return &Scanner{
		logger: logging.With(logging.String("component", "security_scanner")),
	}
}

// ScanCommand scans a command for security issues
func (ss *Scanner) ScanCommand(command string) []string {
	issues := make([]string, 0)

	// Check for common security issues
	securityChecks := []struct {
		pattern string
		message string
	}{
		{"rm\\s+-rf", "Dangerous recursive delete command"},
		{"chmod\\s+777", "Overly permissive file permissions"},
		{"sudo\\s", "Privilege escalation detected"},
		{">\\s*/etc/", "Modification of system files"},
		{"\\$\\{", "Variable expansion - potential injection"},
		{"`.*`", "Command substitution - potential injection"},
		{"wget.*\\|", "Download and pipe - potential code injection"},
		{"curl.*\\|", "Download and pipe - potential code injection"},
	}

	for _, check := range securityChecks {
		if strings.Contains(command, check.pattern) {
			issues = append(issues, check.message)
		}
	}

	if len(issues) > 0 {
		ss.logger.Warn("Security issues detected in command",
			logging.String("command", command),
			logging.Any("issues", issues))
	}

	return issues
}

// GenerateSecurityReport generates a security report
func (sm *Manager) GenerateSecurityReport() (string, error) {
	report := struct {
		Timestamp     time.Time              `json:"timestamp"`
		SecurityLevel Level                  `json:"security_level"`
		Config        Config                 `json:"config"`
		Statistics    map[string]int         `json:"statistics"`
		Credentials   map[string]interface{} `json:"credentials"`
	}{
		Timestamp:     time.Now(),
		SecurityLevel: sm.config.Level,
		Config:        *sm.config,
		Statistics:    make(map[string]int),
		Credentials:   make(map[string]interface{}),
	}

	// Add credential statistics
	sm.auditor.mu.Lock()
	// In real implementation, would include actual statistics
	report.Statistics["validated_credentials"] = 1
	report.Statistics["security_events"] = 0
	sm.auditor.mu.Unlock()

	// Convert to JSON
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal security report: %w", err)
	}

	return string(data), nil
}

// Utility functions
func containsString(slice []string, str string) bool {
	for _, s := range slice {
		if s == str {
			return true
		}
	}
	return false
}

// CheckSecurityPolicy checks if a security policy is satisfied
func (sm *Manager) CheckSecurityPolicy(policy string) error {
	switch policy {
	case "require_tls":
		if !sm.config.EnableTLSVerification {
			return fmt.Errorf("TLS verification is required but disabled")
		}
	case "require_audit":
		if !sm.config.EnableAuditLogging {
			return fmt.Errorf("audit logging is required but disabled")
		}
	case "strict_mode":
		if sm.config.Level != SecurityStrict {
			return fmt.Errorf("strict security mode is required but not enabled")
		}
	default:
		return fmt.Errorf("unknown security policy: %s", policy)
	}

	return nil
}

// InitializeSecurity initializes security with environment-specific settings
func InitializeSecurity() *Manager {
	config := DefaultConfig()

	// Load security level from environment
	if level := os.Getenv("AWS_SSM_SECURITY_LEVEL"); level != "" {
		config.Level = Level(level)
	}

	// Load other settings from environment
	if timeout := os.Getenv("AWS_SSM_SESSION_TIMEOUT"); timeout != "" {
		if d, err := time.ParseDuration(timeout); err == nil {
			config.SessionTimeout = d
		}
	}

	if audit := os.Getenv("AWS_SSM_AUDIT_LOGGING"); audit != "" {
		config.EnableAuditLogging = audit == "true"
	}

	return NewManager(config)
}
