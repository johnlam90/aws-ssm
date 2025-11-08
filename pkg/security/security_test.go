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

func TestBlockedPatternsExtended(t *testing.T) {
	tests := []struct {
		name        string
		command     string
		shouldBlock bool
		description string
	}{
		{
			name:        "rm -rf with wildcard",
			command:     "rm -rf /*",
			shouldBlock: true,
			description: "Should block rm -rf /* (recursive delete with wildcard)",
		},
		{
			name:        "rm -rf /root",
			command:     "rm -rf /root",
			shouldBlock: true,
			description: "Should block rm -rf /root",
		},
		{
			name:        "rm -rf /home",
			command:     "rm -rf /home",
			shouldBlock: true,
			description: "Should block rm -rf /home",
		},
		{
			name:        "rm -rf /etc",
			command:     "rm -rf /etc",
			shouldBlock: true,
			description: "Should block rm -rf /etc",
		},
		{
			name:        "rm -rf /var",
			command:     "rm -rf /var",
			shouldBlock: true,
			description: "Should block rm -rf /var",
		},
		{
			name:        "chmod -R 000",
			command:     "chmod -R 000 /root",
			shouldBlock: true,
			description: "Should block chmod -R 000",
		},
		{
			name:        "chown -R on root",
			command:     "chown -R root:root /",
			shouldBlock: true,
			description: "Should block chown -R on root directory",
		},
		{
			name:        "redirect to /root",
			command:     "echo test > /root/file",
			shouldBlock: true,
			description: "Should block redirect to /root",
		},
		{
			name:        "redirect to /home",
			command:     "echo test > /home/user/file",
			shouldBlock: true,
			description: "Should block redirect to /home",
		},
		{
			name:        "chained chmod with semicolon",
			command:     "ls /tmp; chmod 000 /root",
			shouldBlock: true,
			description: "Should block chained chmod 000",
		},
		{
			name:        "safe rm on /tmp",
			command:     "rm -rf /tmp/cache",
			shouldBlock: false,
			description: "Should allow rm -rf on /tmp",
		},
		{
			name:        "safe chmod on /tmp",
			command:     "chmod 755 /tmp/file",
			shouldBlock: false,
			description: "Should allow chmod on /tmp",
		},
		{
			name:        "rm -rf /etc with extra spaces",
			command:     "rm  -rf  /etc",
			shouldBlock: true,
			description: "Should block rm -rf /etc even with extra spaces",
		},
		{
			name:        "rm -rf / with extra spaces",
			command:     "rm  -rf  /",
			shouldBlock: true,
			description: "Should block rm -rf / even with extra spaces",
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

// TestSuspiciousPatternsNonBlocking tests that benign complex commands with suspicious patterns
// are not blocked, ensuring the security system doesn't over-restrict legitimate operations
func TestSuspiciousPatternsNonBlocking(t *testing.T) {
	config := DefaultConfig()
	manager := NewManager(config)

	tests := []struct {
		name        string
		command     string
		description string
	}{
		{
			name:        "curl with pipe to grep",
			command:     "curl https://example.com | grep 'pattern'",
			description: "Should allow curl with pipe to grep (legitimate data processing)",
		},
		{
			name:        "find with command substitution",
			command:     "find /tmp -name '*.log' -exec rm {} \\;",
			description: "Should allow find with exec (legitimate cleanup in /tmp)",
		},
		{
			name:        "echo with variable expansion",
			command:     "echo $HOME",
			description: "Should allow echo with variable expansion (legitimate shell usage)",
		},
		{
			name:        "grep with regex",
			command:     "grep -E '^[a-z]+$' /var/log/syslog",
			description: "Should allow grep with regex (legitimate log analysis)",
		},
		{
			name:        "sed with backreferences",
			command:     "sed 's/\\([0-9]\\+\\)/[\\1]/g' file.txt",
			description: "Should allow sed with backreferences (legitimate text processing)",
		},
		{
			name:        "awk with complex expression",
			command:     "awk '{print $1, $NF}' data.txt",
			description: "Should allow awk with field references (legitimate data extraction)",
		},
		{
			name:        "ls with pipe to wc",
			command:     "ls -la /home | wc -l",
			description: "Should allow ls with pipe to wc (legitimate counting)",
		},
		{
			name:        "cat with multiple files",
			command:     "cat file1.txt file2.txt | sort",
			description: "Should allow cat with pipe to sort (legitimate file processing)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := manager.validatePatterns(tt.command)
			if err != nil {
				t.Errorf("%s: expected no error but got: %v", tt.description, err)
			}
		})
	}
}
