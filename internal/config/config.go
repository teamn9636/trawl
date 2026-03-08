package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

const defaultModel = "claude-sonnet-4-6"

// Config holds runtime configuration for trawl.
type Config struct {
	APIKey  string `yaml:"api_key"`
	Model   string `yaml:"model"`
	Verbose bool   `yaml:"-"`
}

// Load reads configuration from the environment and optional config file.
func Load() (*Config, error) {
	cfg := &Config{
		Model: defaultModel,
	}

	cfgPath := filepath.Join(os.Getenv("HOME"), ".trawl", "config.yaml")
	if data, err := os.ReadFile(cfgPath); err == nil {
		if err := yaml.Unmarshal(data, cfg); err != nil {
			return nil, fmt.Errorf("parsing config file %s: %w", cfgPath, err)
		}
	}

	if key := os.Getenv("ANTHROPIC_API_KEY"); key != "" {
		cfg.APIKey = key
	}

	if cfg.Model == "" {
		cfg.Model = defaultModel
	}

	return cfg, nil
}

// Validate returns an error if the configuration is unusable.
func (c *Config) Validate() error {
	if c.APIKey == "" {
		return fmt.Errorf("ANTHROPIC_API_KEY is not set; export it or add api_key to ~/.trawl/config.yaml")
	}
	return nil
}

// CacheDir returns the path to the strategy cache directory, creating it if needed.
func CacheDir() (string, error) {
	dir := filepath.Join(os.Getenv("HOME"), ".trawl", "strategies")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", fmt.Errorf("creating cache dir: %w", err)
	}
	return dir, nil
}
