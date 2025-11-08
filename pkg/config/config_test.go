package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfigPathTraversal(t *testing.T) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("failed to get home directory: %v", err)
	}

	// Create a temporary directory for testing
	tmpDir := filepath.Join(homeDir, ".aws-ssm-test")
	os.MkdirAll(tmpDir, 0755)
	defer os.RemoveAll(tmpDir)

	tests := []struct {
		name        string
		configPath  string
		createFile  bool
		shouldFail  bool
		description string
	}{
		{
			name:        "valid home directory path",
			configPath:  filepath.Join(tmpDir, "config.yaml"),
			createFile:  true,
			shouldFail:  false,
			description: "Should allow config in home directory",
		},
		{
			name:        "path traversal attempt with ..",
			configPath:  filepath.Join(tmpDir, "..", "..", "etc", "passwd"),
			createFile:  false,
			shouldFail:  true,
			description: "Should reject path traversal with .. (file doesn't exist)",
		},
		{
			name:        "absolute path outside home and etc",
			configPath:  "/tmp/config.yaml",
			createFile:  false,
			shouldFail:  true,
			description: "Should reject paths outside home and /etc (file doesn't exist)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create config file if needed
			if tt.createFile {
				dir := filepath.Dir(tt.configPath)
				os.MkdirAll(dir, 0755)
				f, err := os.Create(tt.configPath)
				if err != nil {
					t.Fatalf("failed to create test config file: %v", err)
				}
				f.Close()
				defer os.Remove(tt.configPath)
			}

			_, err := LoadConfig(tt.configPath)
			if tt.shouldFail && err == nil {
				t.Errorf("%s: expected error but got none", tt.description)
			}
			if !tt.shouldFail && err != nil {
				t.Errorf("%s: expected no error but got: %v", tt.description, err)
			}
		})
	}
}

func TestLoadConfigDefaultPath(t *testing.T) {
	// Test that default path works
	config, err := LoadConfig("")
	if err != nil {
		t.Errorf("failed to load default config: %v", err)
	}
	if config == nil {
		t.Errorf("expected config but got nil")
	}
}
