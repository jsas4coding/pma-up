package fs

import (
	"os"
	"path/filepath"
	"syscall"
	"testing"
)

var renameFunc = os.Rename

func TestCopyFile_Success(t *testing.T) {
	tempDir := t.TempDir()

	srcFile := filepath.Join(tempDir, "source.txt")
	dstFile := filepath.Join(tempDir, "dest.txt")

	content := []byte("test file content")

	if err := os.WriteFile(srcFile, content, 0644); err != nil {
		t.Fatalf("failed to write source file: %v", err)
	}

	if err := CopyFile(srcFile, dstFile); err != nil {
		t.Fatalf("CopyFile failed: %v", err)
	}

	read, err := os.ReadFile(dstFile)
	if err != nil {
		t.Fatalf("failed to read dest file: %v", err)
	}

	if string(read) != string(content) {
		t.Errorf("content mismatch: expected %q, got %q", content, read)
	}
}

func TestCopyFile_FailureScenarios(t *testing.T) {
	tempDir := t.TempDir()

	t.Run("source does not exist", func(t *testing.T) {
		err := CopyFile(filepath.Join(tempDir, "no-source.txt"), filepath.Join(tempDir, "dest.txt"))
		if err == nil {
			t.Errorf("expected error for missing source file, got nil")
		}
	})

	t.Run("source is not regular file", func(t *testing.T) {
		dirPath := filepath.Join(tempDir, "some-dir")
		if err := os.Mkdir(dirPath, 0755); err != nil {
			t.Fatalf("failed to create dir: %v", err)
		}
		err := CopyFile(dirPath, filepath.Join(tempDir, "dest.txt"))
		if err == nil {
			t.Errorf("expected error for non-regular source file, got nil")
		}
	})
}

func TestMoveDir_Success(t *testing.T) {
	tempDir := t.TempDir()

	sourceDir := filepath.Join(tempDir, "source")
	destDir := filepath.Join(tempDir, "dest")

	if err := os.MkdirAll(sourceDir, 0755); err != nil {
		t.Fatalf("failed to create source dir: %v", err)
	}

	testFile := filepath.Join(sourceDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("move test"), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	if err := MoveDir(sourceDir, destDir); err != nil {
		t.Fatalf("MoveDir failed: %v", err)
	}

	if _, err := os.Stat(sourceDir); !os.IsNotExist(err) {
		t.Errorf("source dir still exists after move")
	}

	read, err := os.ReadFile(filepath.Join(destDir, "test.txt"))
	if err != nil {
		t.Fatalf("failed to read moved file: %v", err)
	}

	if string(read) != "move test" {
		t.Errorf("content mismatch: expected 'move test', got %q", string(read))
	}
}

// simulate copyDir failure by mocking filepath.Walk (advanced scenario - optional in real pipelines)

func TestMoveDir_FallbackCrossDevice(t *testing.T) {
	// here we simulate EXDEV manually to trigger the fallback
	tempDir := t.TempDir()

	sourceDir := filepath.Join(tempDir, "source")
	destDir := filepath.Join(tempDir, "dest")

	if err := os.MkdirAll(sourceDir, 0755); err != nil {
		t.Fatalf("failed to create source dir: %v", err)
	}

	testFile := filepath.Join(sourceDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("move test"), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	// replace os.Rename temporarily to simulate EXDEV
	originalRename := renameFunc
	defer func() { renameFunc = originalRename }()
	renameFunc = func(oldpath, newpath string) error {
		return &os.LinkError{Err: syscall.EXDEV}
	}

	if err := MoveDir(sourceDir, destDir); err != nil {
		t.Fatalf("MoveDir fallback failed: %v", err)
	}

	if _, err := os.Stat(sourceDir); !os.IsNotExist(err) {
		t.Errorf("source dir still exists after fallback move")
	}
}
