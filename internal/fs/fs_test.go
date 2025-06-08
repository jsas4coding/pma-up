package fs

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
)

func resetInjection() {
	inj = injFunc{
		Rename:    os.Rename,
		RemoveAll: os.RemoveAll,
		MkdirAll:  os.MkdirAll,
		Open:      os.Open,
		OpenFile:  os.OpenFile,
		Copy:      io.Copy,
		Rel:       filepath.Rel,
		Walk:      filepath.Walk,
		Stat:      os.Stat,
	}
}

func TestCopyFile_Success(t *testing.T) {
	resetInjection()
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

func TestCopyFile_Errors(t *testing.T) {
	resetInjection()
	tempDir := t.TempDir()

	t.Run("stat fails", func(t *testing.T) {
		inj.Stat = func(string) (os.FileInfo, error) { return nil, fmt.Errorf("stat failed") }
		err := CopyFile("nonexistent", "out")
		if err == nil {
			t.Errorf("expected error, got nil")
		}
		if !strings.Contains(err.Error(), "failed to stat source file") {
			t.Errorf("unexpected error message: %v", err)
		}
	})

	t.Run("non-regular file", func(t *testing.T) {
		dirPath := filepath.Join(tempDir, "dir")
		if err := os.Mkdir(dirPath, 0755); err != nil {
			t.Fatalf("failed to create dir: %v", err)
		}
		err := CopyFile(dirPath, "out")
		if err == nil {
			t.Errorf("expected non-regular file error")
		}
	})
}

func TestMoveDir_Success(t *testing.T) {
	resetInjection()
	tempDir := t.TempDir()
	sourceDir := filepath.Join(tempDir, "source")
	destDir := filepath.Join(tempDir, "dest")
	if err := os.MkdirAll(sourceDir, 0755); err != nil {
		t.Fatalf("failed to create source dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(sourceDir, "file.txt"), []byte("data"), 0644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}

	if err := MoveDir(sourceDir, destDir); err != nil {
		t.Fatalf("MoveDir failed: %v", err)
	}
}

func TestMoveDir_EXDEV(t *testing.T) {
	resetInjection()
	tempDir := t.TempDir()
	sourceDir := filepath.Join(tempDir, "source")
	destDir := filepath.Join(tempDir, "dest")
	if err := os.MkdirAll(sourceDir, 0755); err != nil {
		t.Fatalf("failed to create source dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(sourceDir, "file.txt"), []byte("data"), 0644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}

	inj.Rename = func(_, _ string) error {
		return &os.LinkError{
			Op:  "rename",
			Old: sourceDir,
			New: destDir,
			Err: syscall.EXDEV,
		}
	}

	if err := MoveDir(sourceDir, destDir); err != nil {
		t.Fatalf("MoveDir EXDEV failed: %v", err)
	}
}

func TestFaultInjection_CopyDirFailures(t *testing.T) {
	resetInjection()
	tempDir := t.TempDir()
	sourceDir := filepath.Join(tempDir, "source")
	destDir := filepath.Join(tempDir, "dest")
	if err := os.MkdirAll(sourceDir, 0755); err != nil {
		t.Fatalf("failed to create source dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(sourceDir, "file.txt"), []byte("data"), 0644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}

	t.Run("Walk fails", func(t *testing.T) {
		inj.Walk = func(string, filepath.WalkFunc) error { return fmt.Errorf("walk failed") }
		err := copyDir(sourceDir, destDir)
		if err == nil {
			t.Errorf("expected walk failure")
		}
	})

	t.Run("Rel fails", func(t *testing.T) {
		inj.Rel = func(string, string) (string, error) { return "", fmt.Errorf("rel failed") }
		err := copyDir(sourceDir, destDir)
		if err == nil {
			t.Errorf("expected rel failure")
		}
	})

	t.Run("MkdirAll fails", func(t *testing.T) {
		inj.MkdirAll = func(string, os.FileMode) error { return fmt.Errorf("mkdir failed") }
		err := copyDir(sourceDir, destDir)
		if err == nil {
			t.Errorf("expected mkdir failure")
		}
	})

	t.Run("Open fails", func(t *testing.T) {
		inj.Open = func(string) (*os.File, error) { return nil, fmt.Errorf("open failed") }
		err := copyDir(sourceDir, destDir)
		if err == nil {
			t.Errorf("expected open failure")
		}
	})

	t.Run("OpenFile fails", func(t *testing.T) {
		inj.OpenFile = func(string, int, os.FileMode) (*os.File, error) { return nil, fmt.Errorf("openfile failed") }
		err := copyDir(sourceDir, destDir)
		if err == nil {
			t.Errorf("expected openfile failure")
		}
	})

	t.Run("Copy fails", func(t *testing.T) {
		inj.Copy = func(io.Writer, io.Reader) (int64, error) { return 0, fmt.Errorf("copy failed") }
		err := copyDir(sourceDir, destDir)
		if err == nil {
			t.Errorf("expected copy failure")
		}
	})

	t.Run("RemoveAll fails", func(t *testing.T) {
		inj.RemoveAll = func(string) error { return fmt.Errorf("removeall failed") }
		inj.Rename = func(_, _ string) error {
			return &os.LinkError{Op: "rename", Err: syscall.EXDEV}
		}
		err := MoveDir(sourceDir, destDir)
		if err == nil {
			t.Errorf("expected removeall failure")
		}
	})
}
