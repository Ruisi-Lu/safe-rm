package config

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

// Config represents the safe-rm configuration
type Config struct {
	TrashDir          string   `yaml:"trash_dir"`
	RetentionDays     int      `yaml:"retention_days"`
	ProtectedPaths    []string `yaml:"protected_paths"`
	ProtectedBehavior string   `yaml:"protected_behavior"` // "block" or "confirm"
	VerboseWarnings   bool     `yaml:"verbose_warnings"`
}

// Default returns a Config with default values
func Default() *Config {
	homeDir, _ := os.UserHomeDir()
	return &Config{
		TrashDir:          filepath.Join(homeDir, ".local", "share", "safe-rm", "trash"),
		RetentionDays:     30,
		ProtectedPaths:    []string{},
		ProtectedBehavior: "confirm",
		VerboseWarnings:   true,
	}
}

// Load loads configuration from file and environment variables
func Load() (*Config, error) {
	cfg := Default()

	// Try to load from config file
	configPath := getConfigPath()
	if data, err := os.ReadFile(configPath); err == nil {
		if err := yaml.Unmarshal(data, cfg); err != nil {
			return nil, err
		}
	}

	// Expand ~ in trash_dir
	if strings.HasPrefix(cfg.TrashDir, "~") {
		homeDir, _ := os.UserHomeDir()
		cfg.TrashDir = filepath.Join(homeDir, cfg.TrashDir[1:])
	}

	// Override with environment variables
	if envTrash := os.Getenv("SAFERM_TRASH"); envTrash != "" {
		cfg.TrashDir = envTrash
	}

	if envProtected := os.Getenv("SAFERM_PROTECTED_PATHS"); envProtected != "" {
		paths := strings.Split(envProtected, string(os.PathListSeparator))
		cfg.ProtectedPaths = append(cfg.ProtectedPaths, paths...)
	}

	if envRetention := os.Getenv("SAFERM_RETENTION_DAYS"); envRetention != "" {
		if days, err := strconv.Atoi(envRetention); err == nil {
			cfg.RetentionDays = days
		}
	}

	if envBehavior := os.Getenv("SAFERM_PROTECTED_BEHAVIOR"); envBehavior != "" {
		cfg.ProtectedBehavior = envBehavior
	}

	return cfg, nil
}

func getConfigPath() string {
	// Check XDG_CONFIG_HOME first
	if xdgConfig := os.Getenv("XDG_CONFIG_HOME"); xdgConfig != "" {
		return filepath.Join(xdgConfig, "safe-rm", "config.yml")
	}

	// Fall back to ~/.config
	homeDir, _ := os.UserHomeDir()
	return filepath.Join(homeDir, ".config", "safe-rm", "config.yml")
}

// GetTrashDir returns the resolved trash directory path
func (c *Config) GetTrashDir() string {
	return c.TrashDir
}
