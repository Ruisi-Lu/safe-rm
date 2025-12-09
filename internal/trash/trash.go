package trash

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/user/safe-rm/internal/config"
)

// Metadata stores information about a trashed item
type Metadata struct {
	OriginalPath string    `json:"original_path"`
	DeletedAt    time.Time `json:"deleted_at"`
	Hostname     string    `json:"hostname"`
	IsDirectory  bool      `json:"is_directory"`
}

// Move moves a file or directory to the trash
func Move(cfg *config.Config, absPath string) (string, error) {
	// Get file info
	info, err := os.Lstat(absPath)
	if err != nil {
		return "", err
	}

	// Get hostname
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "unknown"
	}

	// Create trash path preserving original structure
	// Format: $TRASH/<hostname>/<original-path>
	trashBase := cfg.GetTrashDir()
	relativePath := absPath
	if filepath.IsAbs(absPath) {
		// Remove drive letter on Windows or leading / on Unix
		relativePath = absPath
		if len(absPath) > 0 && absPath[0] == '/' {
			relativePath = absPath[1:]
		} else if len(absPath) > 2 && absPath[1] == ':' {
			// Windows: C:\path -> C/path
			relativePath = string(absPath[0]) + absPath[2:]
		}
	}

	trashPath := filepath.Join(trashBase, hostname, relativePath)

	// Handle conflicts by adding timestamp suffix
	if _, err := os.Stat(trashPath); err == nil {
		timestamp := time.Now().Format("20060102-150405")
		trashPath = trashPath + "." + timestamp
	}

	// Create parent directories in trash
	trashDir := filepath.Dir(trashPath)
	if err := os.MkdirAll(trashDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create trash directory: %v", err)
	}

	// Move the file/directory
	if err := os.Rename(absPath, trashPath); err != nil {
		// If rename fails (cross-device), fall back to copy+delete
		if err := copyAndDelete(absPath, trashPath, info.IsDir()); err != nil {
			return "", err
		}
	}

	// Write metadata file
	metadata := Metadata{
		OriginalPath: absPath,
		DeletedAt:    time.Now(),
		Hostname:     hostname,
		IsDirectory:  info.IsDir(),
	}

	metadataPath := trashPath + ".saferm-meta"
	if err := writeMetadata(metadataPath, &metadata); err != nil {
		// Non-fatal: log warning but don't fail the operation
		fmt.Fprintf(os.Stderr, "warning: failed to write metadata: %v\n", err)
	}

	return trashPath, nil
}

func writeMetadata(path string, meta *Metadata) error {
	data, err := json.MarshalIndent(meta, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

func copyAndDelete(src, dst string, isDir bool) error {
	if isDir {
		return copyDirAndDelete(src, dst)
	}
	return copyFileAndDelete(src, dst)
}

func copyFileAndDelete(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}

	info, err := os.Stat(src)
	if err != nil {
		return err
	}

	if err := os.WriteFile(dst, data, info.Mode()); err != nil {
		return err
	}

	return os.Remove(src)
}

func copyDirAndDelete(src, dst string) error {
	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(dst, srcInfo.Mode()); err != nil {
		return err
	}

	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			if err := copyDirAndDelete(srcPath, dstPath); err != nil {
				return err
			}
		} else {
			if err := copyFileAndDelete(srcPath, dstPath); err != nil {
				return err
			}
		}
	}

	return os.RemoveAll(src)
}

// GetMetadata reads metadata for a trashed item
func GetMetadata(trashPath string) (*Metadata, error) {
	metadataPath := trashPath + ".saferm-meta"
	data, err := os.ReadFile(metadataPath)
	if err != nil {
		return nil, err
	}

	var meta Metadata
	if err := json.Unmarshal(data, &meta); err != nil {
		return nil, err
	}

	return &meta, nil
}
