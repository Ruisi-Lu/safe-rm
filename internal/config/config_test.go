package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefault(t *testing.T) {
	cfg := Default()

	if cfg.TrashDir == "" {
		t.Error("Default TrashDir should not be empty")
	}

	if cfg.RetentionDays != 30 {
		t.Errorf("Default RetentionDays = %d, want 30", cfg.RetentionDays)
	}

	if cfg.ProtectedBehavior != "confirm" {
		t.Errorf("Default ProtectedBehavior = %q, want 'confirm'", cfg.ProtectedBehavior)
	}

	if !cfg.VerboseWarnings {
		t.Error("Default VerboseWarnings should be true")
	}
}

func TestLoadWithEnvVars(t *testing.T) {
	// Save and restore environment
	oldTrash := os.Getenv("SAFERM_TRASH")
	oldPaths := os.Getenv("SAFERM_PROTECTED_PATHS")
	oldRetention := os.Getenv("SAFERM_RETENTION_DAYS")
	oldBehavior := os.Getenv("SAFERM_PROTECTED_BEHAVIOR")
	defer func() {
		os.Setenv("SAFERM_TRASH", oldTrash)
		os.Setenv("SAFERM_PROTECTED_PATHS", oldPaths)
		os.Setenv("SAFERM_RETENTION_DAYS", oldRetention)
		os.Setenv("SAFERM_PROTECTED_BEHAVIOR", oldBehavior)
	}()

	// Set test environment variables
	os.Setenv("SAFERM_TRASH", "/custom/trash")
	os.Setenv("SAFERM_PROTECTED_PATHS", "/path1:/path2")
	os.Setenv("SAFERM_RETENTION_DAYS", "7")
	os.Setenv("SAFERM_PROTECTED_BEHAVIOR", "block")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.TrashDir != "/custom/trash" {
		t.Errorf("TrashDir = %q, want '/custom/trash'", cfg.TrashDir)
	}

	if cfg.RetentionDays != 7 {
		t.Errorf("RetentionDays = %d, want 7", cfg.RetentionDays)
	}

	if cfg.ProtectedBehavior != "block" {
		t.Errorf("ProtectedBehavior = %q, want 'block'", cfg.ProtectedBehavior)
	}

	// Check protected paths (note: separator is OS-dependent)
	if len(cfg.ProtectedPaths) < 2 {
		t.Error("ProtectedPaths should have at least 2 entries from env var")
	}
}

func TestLoadWithConfigFile(t *testing.T) {
	// Create a temp config directory
	tempDir, err := os.MkdirTemp("", "saferm-config-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	// Set XDG_CONFIG_HOME to temp directory
	oldXDG := os.Getenv("XDG_CONFIG_HOME")
	os.Setenv("XDG_CONFIG_HOME", tempDir)
	defer os.Setenv("XDG_CONFIG_HOME", oldXDG)

	// Create config file
	configDir := filepath.Join(tempDir, "safe-rm")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatal(err)
	}

	configContent := `trash_dir: /from/config
retention_days: 14
protected_behavior: block
protected_paths:
  - /protected/one
  - /protected/two
`
	if err := os.WriteFile(filepath.Join(configDir, "config.yml"), []byte(configContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Clear env vars that might override
	os.Unsetenv("SAFERM_TRASH")
	os.Unsetenv("SAFERM_RETENTION_DAYS")
	os.Unsetenv("SAFERM_PROTECTED_BEHAVIOR")
	os.Unsetenv("SAFERM_PROTECTED_PATHS")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.TrashDir != "/from/config" {
		t.Errorf("TrashDir = %q, want '/from/config'", cfg.TrashDir)
	}

	if cfg.RetentionDays != 14 {
		t.Errorf("RetentionDays = %d, want 14", cfg.RetentionDays)
	}

	if len(cfg.ProtectedPaths) != 2 {
		t.Errorf("ProtectedPaths count = %d, want 2", len(cfg.ProtectedPaths))
	}
}

func TestGetTrashDir(t *testing.T) {
	cfg := &Config{
		TrashDir: "/test/trash",
	}

	if cfg.GetTrashDir() != "/test/trash" {
		t.Errorf("GetTrashDir() = %q, want '/test/trash'", cfg.GetTrashDir())
	}
}
