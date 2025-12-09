package trash

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/user/safe-rm/internal/config"
)

func TestMove(t *testing.T) {
	// Create a temp directory for testing
	tempDir, err := os.MkdirTemp("", "saferm-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	// Create a test file
	testFile := filepath.Join(tempDir, "testfile.txt")
	if err := os.WriteFile(testFile, []byte("test content"), 0644); err != nil {
		t.Fatal(err)
	}

	// Create config with trash in temp directory
	cfg := &config.Config{
		TrashDir: filepath.Join(tempDir, "trash"),
	}

	// Move the file to trash
	trashPath, err := Move(cfg, testFile)
	if err != nil {
		t.Fatalf("Move() error = %v", err)
	}

	// Verify original file is gone
	if _, err := os.Stat(testFile); !os.IsNotExist(err) {
		t.Error("Original file should not exist after Move()")
	}

	// Verify file is in trash
	if _, err := os.Stat(trashPath); err != nil {
		t.Errorf("Trashed file should exist at %s: %v", trashPath, err)
	}

	// Verify metadata file exists
	metaPath := trashPath + ".saferm-meta"
	if _, err := os.Stat(metaPath); err != nil {
		t.Errorf("Metadata file should exist at %s: %v", metaPath, err)
	}

	// Verify metadata content
	meta, err := GetMetadata(trashPath)
	if err != nil {
		t.Fatalf("GetMetadata() error = %v", err)
	}

	if meta.OriginalPath != testFile {
		t.Errorf("Metadata.OriginalPath = %q, want %q", meta.OriginalPath, testFile)
	}

	if meta.IsDirectory {
		t.Error("Metadata.IsDirectory should be false for a file")
	}
}

func TestMoveDirectory(t *testing.T) {
	// Create a temp directory for testing
	tempDir, err := os.MkdirTemp("", "saferm-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	// Create a test directory with files
	testDir := filepath.Join(tempDir, "testdir")
	if err := os.Mkdir(testDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(testDir, "file1.txt"), []byte("content1"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(testDir, "file2.txt"), []byte("content2"), 0644); err != nil {
		t.Fatal(err)
	}

	// Create config with trash in temp directory
	cfg := &config.Config{
		TrashDir: filepath.Join(tempDir, "trash"),
	}

	// Move the directory to trash
	trashPath, err := Move(cfg, testDir)
	if err != nil {
		t.Fatalf("Move() error = %v", err)
	}

	// Verify original directory is gone
	if _, err := os.Stat(testDir); !os.IsNotExist(err) {
		t.Error("Original directory should not exist after Move()")
	}

	// Verify directory is in trash
	if _, err := os.Stat(trashPath); err != nil {
		t.Errorf("Trashed directory should exist at %s: %v", trashPath, err)
	}

	// Verify files are in trashed directory
	if _, err := os.Stat(filepath.Join(trashPath, "file1.txt")); err != nil {
		t.Error("file1.txt should exist in trashed directory")
	}
	if _, err := os.Stat(filepath.Join(trashPath, "file2.txt")); err != nil {
		t.Error("file2.txt should exist in trashed directory")
	}
}

func TestMoveConflict(t *testing.T) {
	// Create a temp directory for testing
	tempDir, err := os.MkdirTemp("", "saferm-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	cfg := &config.Config{
		TrashDir: filepath.Join(tempDir, "trash"),
	}

	// Create and move first file
	testFile1 := filepath.Join(tempDir, "testfile.txt")
	if err := os.WriteFile(testFile1, []byte("content1"), 0644); err != nil {
		t.Fatal(err)
	}
	trashPath1, err := Move(cfg, testFile1)
	if err != nil {
		t.Fatalf("Move() first file error = %v", err)
	}

	// Create another file with the same name
	testFile2 := filepath.Join(tempDir, "testfile.txt")
	if err := os.WriteFile(testFile2, []byte("content2"), 0644); err != nil {
		t.Fatal(err)
	}
	trashPath2, err := Move(cfg, testFile2)
	if err != nil {
		t.Fatalf("Move() second file error = %v", err)
	}

	// Paths should be different due to conflict handling
	if trashPath1 == trashPath2 {
		t.Error("Trash paths should be different for conflicting names")
	}
}
