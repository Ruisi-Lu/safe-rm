package protect

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/user/safe-rm/internal/config"
)

func TestCheckBuiltinProtectedPaths(t *testing.T) {
	cfg := config.Default()

	tests := []struct {
		path      string
		recursive bool
		want      bool
		desc      string
	}{
		{"/", false, true, "root directory"},
		{"/home", false, true, "home directory"},
		{"/usr", false, true, "usr directory"},
		{"/bin", false, true, "bin directory"},
		{"/etc", false, true, "etc directory"},
		{"/var", false, true, "var directory"},
		{"/tmp/testfile", false, false, "regular file in tmp"},
		{"/home/user/documents", false, false, "user documents"},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			status := Check(cfg, tt.path, tt.recursive)
			if status.Protected != tt.want {
				t.Errorf("Check(%q) = %v, want %v", tt.path, status.Protected, tt.want)
			}
		})
	}
}

func TestCheckGitDirectory(t *testing.T) {
	cfg := config.Default()

	// Create a temp directory with .git
	tempDir, err := os.MkdirTemp("", "saferm-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	gitDir := filepath.Join(tempDir, ".git")
	if err := os.Mkdir(gitDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Test .git directory detection
	status := Check(cfg, gitDir, false)
	if !status.Protected {
		t.Errorf("Check(%q) should be protected (is .git directory)", gitDir)
	}
}

func TestCheckCustomProtectedPaths(t *testing.T) {
	cfg := config.Default()
	cfg.ProtectedPaths = []string{
		"/custom/protected/*",
		"/important/file.txt",
	}

	tests := []struct {
		path string
		want bool
		desc string
	}{
		{"/important/file.txt", true, "exactly matching protected file"},
		{"/custom/protected/something", true, "glob pattern match"},
		{"/custom/other/file", false, "non-matching path"},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			status := Check(cfg, tt.path, false)
			if status.Protected != tt.want {
				t.Errorf("Check(%q) = %v, want %v (reason: %s)", tt.path, status.Protected, tt.want, status.Reason)
			}
		})
	}
}

func TestIsProtectedByDefault(t *testing.T) {
	tests := []struct {
		path string
		want bool
	}{
		{"/", true},
		{"/home", true},
		{"/tmp/file", false},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			got := IsProtectedByDefault(tt.path)
			if got != tt.want {
				t.Errorf("IsProtectedByDefault(%q) = %v, want %v", tt.path, got, tt.want)
			}
		})
	}
}
