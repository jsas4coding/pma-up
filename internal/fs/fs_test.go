package fs

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCopyFile(t *testing.T) {
	tempDir := t.TempDir()

	srcFile := filepath.Join(tempDir, "source.txt")
	dstFile := filepath.Join(tempDir, "dest.txt")

	content := []byte("test file content")

	// Create source file
	if err := os.WriteFile(srcFile, content, 0644); err != nil {
		t.Fatalf("failed to write source file: %v", err)
	}

	// Perform copy operation
	if err := CopyFile(srcFile, dstFile); err != nil {
		t.Fatalf("CopyFile failed: %v", err)
	}

	// Verify destination file content
	read, err := os.ReadFile(dstFile)
	if err != nil {
		t.Fatalf("failed to read destination file: %v", err)
	}

	if string(read) != string(content) {
		t.Errorf("content mismatch: expected %q, got %q", content, read)
	}
}

func TestMoveDir(t *testing.T) {
	tempDir := t.TempDir()

	sourceDir := filepath.Join(tempDir, "source")
	destDir := filepath.Join(tempDir, "dest")

	// Create source directory
	if err := os.MkdirAll(sourceDir, 0755); err != nil {
		t.Fatalf("failed to create source directory: %v", err)
	}

	testFile := filepath.Join(sourceDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("move test"), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	// Perform move operation
	if err := MoveDir(sourceDir, destDir); err != nil {
		t.Fatalf("MoveDir failed: %v", err)
	}

	// Verify source directory was removed
	if _, err := os.Stat(sourceDir); !os.IsNotExist(err) {
		t.Errorf("source directory still exists after move")
	}

	// Verify file was moved correctly
	read, err := os.ReadFile(filepath.Join(destDir, "test.txt"))
	if err != nil {
		t.Fatalf("failed to read moved file: %v", err)
	}

	if string(read) != "move test" {
		t.Errorf("content mismatch: expected 'move test', got %q", string(read))
	}
}
