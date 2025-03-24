package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

// Config represents the plugin configuration
type Config struct {
	Meshery  MesheryConfig  `yaml:"meshery"`
	Defaults DefaultsConfig `yaml:"defaults"`
}

// MesheryConfig represents Meshery server configuration
type MesheryConfig struct {
	URL              string `yaml:"url"`
	SnapshotEndpoint string `yaml:"snapshot_endpoint"`
}

// DefaultsConfig represents default settings
type DefaultsConfig struct {
	SnapshotName       string `yaml:"snapshot_name"`
	TimeoutSeconds     int    `yaml:"timeout_seconds"`
	NotifyOnCompletion bool   `yaml:"notify_on_completion"`
}

// GetConfigFilePath returns the path to the config file
func GetConfigFilePath() string {
	// Check in current directory
	if _, err := os.Stat("config/config.yaml"); err == nil {
		return "config/config.yaml"
	}

	// Check in user's home directory
	homeDir, err := os.UserHomeDir()
	if err == nil {
		configPath := filepath.Join(homeDir, ".meshery", "kubectl-kanvas-snapshot", "config.yaml")
		if _, err := os.Stat(configPath); err == nil {
			return configPath
		}
	}

	// Return default location
	return "config/config.yaml"
}

// LoadConfig loads the configuration from the config file
func LoadConfig() (*Config, error) {
	configPath := GetConfigFilePath()

	// Check if config file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return DefaultConfig(), nil
	}

	// Read config file
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("error reading config file: %w", err)
	}

	// Parse YAML
	config := &Config{}
	err = yaml.Unmarshal(data, config)
	if err != nil {
		return nil, fmt.Errorf("error parsing config file: %w", err)
	}

	// Validate config
	if config.Meshery.URL == "" {
		config.Meshery.URL = "http://localhost:9081"
	}

	if config.Meshery.SnapshotEndpoint == "" {
		config.Meshery.SnapshotEndpoint = "/api/k8scontext/manifest"
	}

	return config, nil
}

// DefaultConfig returns the default configuration
func DefaultConfig() *Config {
	return &Config{
		Meshery: MesheryConfig{
			URL:              "http://localhost:9081",
			SnapshotEndpoint: "/api/k8scontext/manifest",
		},
		Defaults: DefaultsConfig{
			SnapshotName:       "kubectl-snapshot",
			TimeoutSeconds:     30,
			NotifyOnCompletion: true,
		},
	}
}
