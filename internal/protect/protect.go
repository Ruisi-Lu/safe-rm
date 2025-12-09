package protect

import (
	"path/filepath"
	"strings"

	"github.com/user/safe-rm/internal/config"
)

// Status represents the protection status of a path
type Status struct {
	Protected bool
	Reason    string
}

// Built-in protected paths (absolute paths on Unix-like systems)
var builtinProtectedPaths = []string{
	"/",
	"/bin",
	"/boot",
	"/dev",
	"/etc",
	"/home",
	"/lib",
	"/lib64",
	"/opt",
	"/proc",
	"/root",
	"/run",
	"/sbin",
	"/srv",
	"/sys",
	"/tmp",
	"/usr",
	"/var",
}

// Check checks if a path is protected
func Check(cfg *config.Config, absPath string, recursive bool) Status {
	// Normalize path
	absPath = filepath.Clean(absPath)

	// Check for root directory
	if absPath == "/" || absPath == "\\" {
		return Status{
			Protected: true,
			Reason:    "Root directory is always protected",
		}
	}

	// Check for dangerous patterns like /* (all top-level dirs)
	if isWildcardRoot(absPath) {
		return Status{
			Protected: true,
			Reason:    "Wildcard patterns targeting root level are blocked",
		}
	}

	// Check built-in protected paths
	for _, protected := range builtinProtectedPaths {
		if absPath == protected || absPath == protected+"/" {
			return Status{
				Protected: true,
				Reason:    "System directory is protected: " + protected,
			}
		}
		// Also protect if trying to recursively delete parent of protected path
		if recursive && strings.HasPrefix(protected, absPath+"/") {
			return Status{
				Protected: true,
				Reason:    "Path contains protected system directory: " + protected,
			}
		}
	}

	// Check for .git directories
	if isGitPath(absPath) {
		return Status{
			Protected: true,
			Reason:    ".git directory or repository root is protected",
		}
	}

	// Check user-defined protected paths from config
	for _, pattern := range cfg.ProtectedPaths {
		// Expand ~ in pattern
		if strings.HasPrefix(pattern, "~") {
			homeDir, _ := filepath.Abs(filepath.Join("~"))
			pattern = strings.Replace(pattern, "~", homeDir, 1)
		}

		matched, err := filepath.Match(pattern, absPath)
		if err == nil && matched {
			return Status{
				Protected: true,
				Reason:    "Path matches protected pattern: " + pattern,
			}
		}

		// Also check if absPath is under a protected directory pattern
		if strings.HasSuffix(pattern, "/**") {
			dirPattern := strings.TrimSuffix(pattern, "/**")
			if strings.HasPrefix(absPath, dirPattern) {
				return Status{
					Protected: true,
					Reason:    "Path is under protected directory: " + dirPattern,
				}
			}
		}
	}

	return Status{Protected: false}
}

// isWildcardRoot checks if the path looks like a dangerous wildcard operation
func isWildcardRoot(path string) bool {
	// On Unix, /* expands to all top-level directories
	// The shell will expand this before rm sees it, but we check anyway
	return path == "/*" || path == "\\*"
}

// isGitPath checks if the path is a .git directory or contains one
func isGitPath(absPath string) bool {
	// Check if path ends with .git
	if filepath.Base(absPath) == ".git" {
		return true
	}

	// Check if .git exists in this directory (repository root)
	gitPath := filepath.Join(absPath, ".git")
	if _, err := filepath.Abs(gitPath); err == nil {
		// Check if .git directory actually exists
		if info, err := filepath.Glob(gitPath); err == nil && len(info) > 0 {
			return true
		}
	}

	return false
}

// IsProtectedByDefault returns true if the path is in the built-in protected list
func IsProtectedByDefault(absPath string) bool {
	absPath = filepath.Clean(absPath)
	for _, protected := range builtinProtectedPaths {
		if absPath == protected {
			return true
		}
	}
	return false
}
