package security

import (
	"testing"
)

func TestValidatePatternsBlocked(t *testing.T) {
	tests := []struct {
		name        string
		command     string
		shouldBlock bool
		description string
	}{
		{
			name:        "rm -rf command",
			command:     "rm -rf /",
			shouldBlock: true,
			description: "Should block rm -rf / (destructive root operation)",
		},
		{
			name:        "chmod with digits",
			command:     "chmod 000 /",
			shouldBlock: true,
			description: "Should block chmod 000 / (destructive root operation)",
		},
		{
			name:        "chown command",
			command:     "chown root:root /",
			shouldBlock: true,
			description: "Should block chown on root directory",
		},
		{
			name:        "sudo rm -rf",
			command:     "sudo rm -rf /",
			shouldBlock: true,
			description: "Should block sudo rm -rf (dangerous combination)",
		},
		{
			name:        "redirect to /etc/",
			command:     "echo test > /etc/config",
			shouldBlock: true,
			description: "Should block redirect to /etc/",
		},
		{
			name:        "redirect to /sys/",
			command:     "echo test > /sys/config",
			shouldBlock: true,
			description: "Should block redirect to /sys/",
		},
		{
			name:        "semicolon with rm -rf",
			command:     "find /tmp -name '*.log'; rm -rf /",
			shouldBlock: true,
			description: "Should block semicolon with rm -rf /",
		},
		{
			name:        "safe ls command",
			command:     "ls -la /tmp",
			shouldBlock: false,
			description: "Should allow safe ls command",
		},
		{
			name:        "safe cat command",
			command:     "cat /var/log/syslog",
			shouldBlock: false,
			description: "Should allow safe cat command",
		},
		{
			name:        "safe grep command",
			command:     "grep error /var/log/app.log",
			shouldBlock: false,
			description: "Should allow safe grep command",
		},
		{
			name:        "safe chmod on /tmp",
			command:     "chmod 755 /tmp/file",
			shouldBlock: false,
			description: "Should allow chmod on non-root directories",
		},
		{
			name:        "safe variable expansion",
			command:     "echo ${HOME}",
			shouldBlock: false,
			description: "Should allow variable expansion (not blocked by default)",
		},
		{
			name:        "safe curl command",
			command:     "curl https://example.com",
			shouldBlock: false,
			description: "Should allow curl (legitimate tool)",
		},
	}

	config := DefaultConfig()
	manager := NewManager(config)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := manager.validatePatterns(tt.command)
			if tt.shouldBlock && err == nil {
				t.Errorf("%s: expected error but got none", tt.description)
			}
			if !tt.shouldBlock && err != nil {
				t.Errorf("%s: expected no error but got: %v", tt.description, err)
			}
		})
	}
}

func TestValidatePatternsHighSecurity(t *testing.T) {
	tests := []struct {
		name        string
		command     string
		shouldBlock bool
		description string
	}{
		{
			name:        "command substitution",
			command:     "echo $(whoami)",
			shouldBlock: true,
			description: "Should block command substitution in high security",
		},
		{
			name:        "backtick substitution",
			command:     "echo `whoami`",
			shouldBlock: true,
			description: "Should block backtick substitution in high security",
		},
		{
			name:        "pipe operator",
			command:     "cat file | grep pattern",
			shouldBlock: true,
			description: "Should block pipe in high security",
		},
		{
			name:        "curl command",
			command:     "curl https://example.com",
			shouldBlock: false,
			description: "Should allow curl (legitimate tool, not blocked by default)",
		},
		{
			name:        "wget command",
			command:     "wget https://example.com/file",
			shouldBlock: false,
			description: "Should allow wget (legitimate tool, not blocked by default)",
		},
		{
			name:        "safe echo command",
			command:     "echo hello world",
			shouldBlock: false,
			description: "Should allow safe echo in high security",
		},
	}

	config := DefaultConfig()
	config.Level = SecurityHigh
	manager := NewManager(config)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := manager.validatePatterns(tt.command)
			if tt.shouldBlock && err == nil {
				t.Errorf("%s: expected error but got none", tt.description)
			}
			if !tt.shouldBlock && err != nil {
				t.Errorf("%s: expected no error but got: %v", tt.description, err)
			}
		})
	}
}

func TestValidateCommandIntegration(t *testing.T) {
	tests := []struct {
		name        string
		command     string
		shouldBlock bool
		description string
	}{
		{
			name:        "dangerous rm command",
			command:     "rm -rf /",
			shouldBlock: true,
			description: "Should block dangerous rm command",
		},
		{
			name:        "safe ls command",
			command:     "ls -la /home",
			shouldBlock: false,
			description: "Should allow safe ls command",
		},
		{
			name:        "command too long",
			command:     string(make([]byte, 2000)),
			shouldBlock: true,
			description: "Should block command exceeding max length",
		},
	}

	config := DefaultConfig()
	manager := NewManager(config)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := manager.ValidateCommand(tt.command)
			if tt.shouldBlock && err == nil {
				t.Errorf("%s: expected error but got none", tt.description)
			}
			if !tt.shouldBlock && err != nil {
				t.Errorf("%s: expected no error but got: %v", tt.description, err)
			}
		})
	}
}
