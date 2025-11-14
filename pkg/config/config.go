// Package config loads and persists CLI configuration files and defaults.
package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	yaml "gopkg.in/yaml.v3"
)

// Config represents the application configuration
type Config struct {
	Default struct {
		Region  string            `yaml:"region"`
		Profile string            `yaml:"profile"`
		Filters map[string]string `yaml:"filters"`
		Columns []string          `yaml:"columns"`
		Weights map[string]int    `yaml:"weights"`
	} `yaml:"default"`
	Interactive struct {
		Columns      []string `yaml:"columns"`
		NoColor      bool     `yaml:"no_color"`
		Width        int      `yaml:"width"`
		CacheTTL     int      `yaml:"cache_ttl_minutes"`
		MaxInstances int      `yaml:"max_instances"`
	} `yaml:"interactive"`
	Keybindings map[string]string `yaml:"keybindings"`
	Cache       struct {
		Enabled    bool   `yaml:"enabled"`
		TTLMinutes int    `yaml:"ttl_minutes"`
		CacheDir   string `yaml:"cache_dir"`
	} `yaml:"cache"`
	Bookmarks struct {
		File string `yaml:"file"`
	} `yaml:"bookmarks"`
	Plugins struct {
		Dir string `yaml:"dir"`
	} `yaml:"plugins"`
}

// LoadConfig loads configuration from file
func LoadConfig(configPath string) (*Config, error) {
	// Resolve and validate config path
	validatedPath, err := resolveConfigPath(configPath)
	if err != nil {
		return nil, err
	}

	// Create default configuration
	config := createDefaultConfig()

	// Load configuration from file if it exists
	if err := loadConfigFile(config, validatedPath); err != nil {
		return nil, err
	}

	// Set default paths if not specified
	if err := setDefaultPaths(config); err != nil {
		return nil, err
	}

	return config, nil
}

// resolveConfigPath resolves and validates the config file path
func resolveConfigPath(configPath string) (string, error) {
	if configPath == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("failed to get user home directory: %w", err)
		}
		return filepath.Join(homeDir, ".aws-ssm", "config.yaml"), nil
	}

	// Validate and clean the path to prevent directory traversal
	cleanPath := filepath.Clean(configPath)

	// Get absolute path to ensure we're comparing canonical paths
	absPath, err := filepath.Abs(cleanPath)
	if err != nil {
		return "", fmt.Errorf("invalid config path: %w", err)
	}

	// Get the home directory as the base for config files
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to determine home directory: %w", err)
	}

	// Verify the absolute path is within the home directory or /etc
	// This prevents directory traversal attacks
	if !strings.HasPrefix(absPath, homeDir) && !strings.HasPrefix(absPath, "/etc") {
		return "", fmt.Errorf("invalid config path: must be in home directory (~/.aws-ssm/) or /etc/aws-ssm/")
	}

	return cleanPath, nil
}

// createDefaultConfig creates the default configuration
func createDefaultConfig() *Config {
	return &Config{
		Default: struct {
			Region  string            `yaml:"region"`
			Profile string            `yaml:"profile"`
			Filters map[string]string `yaml:"filters"`
			Columns []string          `yaml:"columns"`
			Weights map[string]int    `yaml:"weights"`
		}{
			Filters: make(map[string]string),
			Columns: []string{"name", "instance-id", "private-ip", "state"},
			Weights: map[string]int{
				"name":        5,
				"instance-id": 4,
				"tag":         3,
				"ip":          2,
				"dns":         1,
			},
		},
		Interactive: struct {
			Columns      []string `yaml:"columns"`
			NoColor      bool     `yaml:"no_color"`
			Width        int      `yaml:"width"`
			CacheTTL     int      `yaml:"cache_ttl_minutes"`
			MaxInstances int      `yaml:"max_instances"`
		}{
			Columns:      []string{"name", "instance-id", "private-ip", "state"},
			NoColor:      false,
			Width:        0,
			CacheTTL:     5,
			MaxInstances: 10000,
		},
		Keybindings: map[string]string{
			"enter":  "connect",
			"c":      "command",
			"p":      "port-forward",
			"i":      "interfaces",
			"b":      "bookmark",
			"r":      "refresh",
			"R":      "region",
			"f2":     "sort",
			"j":      "json-view",
			"t":      "toggle-tags",
			"s":      "toggle-stopped",
			"escape": "exit",
			"ctrl+c": "exit",
			"space":  "select",
			":":      "palette",
		},
		Cache: struct {
			Enabled    bool   `yaml:"enabled"`
			TTLMinutes int    `yaml:"ttl_minutes"`
			CacheDir   string `yaml:"cache_dir"`
		}{
			Enabled:    true,
			TTLMinutes: 5,
			CacheDir:   "",
		},
		Bookmarks: struct {
			File string `yaml:"file"`
		}{
			File: "",
		},
		Plugins: struct {
			Dir string `yaml:"dir"`
		}{
			Dir: "",
		},
	}
}

// loadConfigFile loads configuration from file if it exists
func loadConfigFile(config *Config, configPath string) error {
	if _, statErr := os.Stat(configPath); statErr == nil {
		data, err := os.ReadFile(configPath)
		if err != nil {
			return fmt.Errorf("failed to read config file: %w", err)
		}

		if err := yaml.Unmarshal(data, config); err != nil {
			return fmt.Errorf("failed to parse config file: %w", err)
		}
	} else if !os.IsNotExist(statErr) {
		return fmt.Errorf("failed to access config file: %w", statErr)
	}
	return nil
}

// setDefaultPaths sets default paths for directories if not specified
func setDefaultPaths(config *Config) error {
	if config.Bookmarks.File == "" || config.Cache.CacheDir == "" || config.Plugins.Dir == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to get user home directory: %w", err)
		}

		if config.Bookmarks.File == "" {
			config.Bookmarks.File = filepath.Join(homeDir, ".aws-ssm", "favorites.json")
		}

		if config.Cache.CacheDir == "" {
			config.Cache.CacheDir = filepath.Join(homeDir, ".aws-ssm", "cache")
		}

		if config.Plugins.Dir == "" {
			config.Plugins.Dir = filepath.Join(homeDir, ".aws-ssm", "actions")
		}
	}
	return nil
}

// SaveConfig saves configuration to file
func SaveConfig(config *Config, configPath string) error {
	if configPath == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to get user home directory: %w", err)
		}
		configPath = filepath.Join(homeDir, ".aws-ssm", "config.yaml")

		// Ensure config directory exists with restricted permissions
		if err := os.MkdirAll(filepath.Dir(configPath), 0700); err != nil {
			return fmt.Errorf("failed to create config directory: %w", err)
		}
	}

	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	return os.WriteFile(configPath, data, 0600)
}
