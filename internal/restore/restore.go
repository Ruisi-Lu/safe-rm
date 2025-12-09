package restore

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/user/safe-rm/internal/config"
	"github.com/user/safe-rm/internal/trash"
)

// List displays all items in the trash
func List(cfg *config.Config) error {
	trashDir := cfg.GetTrashDir()

	if _, err := os.Stat(trashDir); os.IsNotExist(err) {
		fmt.Println("Trash is empty.")
		return nil
	}

	items, err := findTrashItems(trashDir)
	if err != nil {
		return err
	}

	if len(items) == 0 {
		fmt.Println("Trash is empty.")
		return nil
	}

	fmt.Printf("Items in trash (%s):\n\n", trashDir)
	fmt.Printf("%-30s %-50s %s\n", "DELETED AT", "ORIGINAL PATH", "TRASH PATH")
	fmt.Println(strings.Repeat("-", 120))

	for _, item := range items {
		meta, err := trash.GetMetadata(item)
		if err != nil {
			// If no metadata, show what we can
			fmt.Printf("%-30s %-50s %s\n", "unknown", "unknown", item)
			continue
		}
		fmt.Printf("%-30s %-50s %s\n",
			meta.DeletedAt.Format("2006-01-02 15:04:05"),
			meta.OriginalPath,
			item)
	}

	return nil
}

// Restore restores a file from trash to its original location
func Restore(cfg *config.Config, originalPath string) error {
	trashDir := cfg.GetTrashDir()

	// Find the item in trash
	items, err := findTrashItems(trashDir)
	if err != nil {
		return err
	}

	var matchedItem string
	var matchedMeta *trash.Metadata

	for _, item := range items {
		meta, err := trash.GetMetadata(item)
		if err != nil {
			continue
		}

		if meta.OriginalPath == originalPath {
			// If multiple matches, prefer the most recent
			if matchedMeta == nil || meta.DeletedAt.After(matchedMeta.DeletedAt) {
				matchedItem = item
				matchedMeta = meta
			}
		}
	}

	if matchedItem == "" {
		return fmt.Errorf("no item found in trash with original path: %s", originalPath)
	}

	// Check if destination exists
	if _, err := os.Stat(originalPath); err == nil {
		return fmt.Errorf("destination already exists: %s", originalPath)
	}

	// Create parent directory if needed
	parentDir := filepath.Dir(originalPath)
	if err := os.MkdirAll(parentDir, 0755); err != nil {
		return fmt.Errorf("failed to create parent directory: %v", err)
	}

	// Move the item back
	if err := os.Rename(matchedItem, originalPath); err != nil {
		return fmt.Errorf("failed to restore: %v", err)
	}

	// Remove metadata file
	metadataPath := matchedItem + ".saferm-meta"
	os.Remove(metadataPath) // Ignore error

	fmt.Printf("Restored: %s -> %s\n", matchedItem, originalPath)
	return nil
}

// Purge removes items older than the specified number of days
func Purge(cfg *config.Config, days int) error {
	trashDir := cfg.GetTrashDir()

	if _, err := os.Stat(trashDir); os.IsNotExist(err) {
		fmt.Println("Trash is empty, nothing to purge.")
		return nil
	}

	items, err := findTrashItems(trashDir)
	if err != nil {
		return err
	}

	cutoff := time.Now().AddDate(0, 0, -days)
	purged := 0

	for _, item := range items {
		meta, err := trash.GetMetadata(item)
		if err != nil {
			// If no metadata, check file modification time
			info, err := os.Stat(item)
			if err != nil {
				continue
			}
			if info.ModTime().Before(cutoff) {
				if err := os.RemoveAll(item); err == nil {
					purged++
					fmt.Printf("Purged: %s\n", item)
				}
			}
			continue
		}

		if meta.DeletedAt.Before(cutoff) {
			if err := os.RemoveAll(item); err == nil {
				os.Remove(item + ".saferm-meta")
				purged++
				fmt.Printf("Purged: %s (deleted at %s)\n", meta.OriginalPath, meta.DeletedAt.Format("2006-01-02"))
			}
		}
	}

	if purged == 0 {
		fmt.Printf("No items older than %d days found.\n", days)
	} else {
		fmt.Printf("\nPurged %d item(s).\n", purged)
	}

	return nil
}

// Empty permanently deletes all items in the trash
func Empty(cfg *config.Config) error {
	trashDir := cfg.GetTrashDir()

	if _, err := os.Stat(trashDir); os.IsNotExist(err) {
		fmt.Println("Trash is already empty.")
		return nil
	}

	items, err := findTrashItems(trashDir)
	if err != nil {
		return err
	}

	if len(items) == 0 {
		fmt.Println("Trash is already empty.")
		return nil
	}

	// Require confirmation
	fmt.Printf("WARNING: This will PERMANENTLY DELETE %d item(s) from trash!\n", len(items))
	fmt.Printf("This action cannot be undone.\n")
	fmt.Printf("Type 'yes I am sure' to confirm: ")

	var response string
	fmt.Scanln(&response)
	if response != "yes I am sure" {
		fmt.Println("Aborted.")
		return nil
	}

	// Delete all items
	deleted := 0
	for _, item := range items {
		if err := os.RemoveAll(item); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to delete %s: %v\n", item, err)
			continue
		}
		// Also remove metadata file
		os.Remove(item + ".saferm-meta")
		deleted++
	}

	// Clean up empty directories in trash
	cleanEmptyDirs(trashDir)

	fmt.Printf("\nPermanently deleted %d item(s).\n", deleted)
	return nil
}

// cleanEmptyDirs removes empty directories in the trash
func cleanEmptyDirs(dir string) {
	filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil || !info.IsDir() || path == dir {
			return nil
		}
		entries, err := os.ReadDir(path)
		if err == nil && len(entries) == 0 {
			os.Remove(path)
		}
		return nil
	})
}

// findTrashItems finds all trashed items (files without .saferm-meta extension)
func findTrashItems(trashDir string) ([]string, error) {
	var items []string

	err := filepath.Walk(trashDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip errors
		}

		// Skip metadata files
		if strings.HasSuffix(path, ".saferm-meta") {
			return nil
		}

		// Skip the root trash directory itself
		if path == trashDir {
			return nil
		}

		// Skip directories that contain other items (we only want leaf items)
		if info.IsDir() {
			entries, err := os.ReadDir(path)
			if err != nil {
				return nil
			}
			// If directory has entries, skip it (we'll get the contents)
			for _, entry := range entries {
				if !strings.HasSuffix(entry.Name(), ".saferm-meta") {
					return nil
				}
			}
		}

		// Check if there's a metadata file for this item
		if _, err := os.Stat(path + ".saferm-meta"); err == nil {
			items = append(items, path)
		}

		return nil
	})

	return items, err
}
