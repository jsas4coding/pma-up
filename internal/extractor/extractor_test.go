package extractor

import (
	"archive/zip"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestExtractZip_Success(t *testing.T) {
	tempDir := t.TempDir()
	zipPath := filepath.Join(tempDir, "test.zip")
	testFiles := map[string]string{
		"folder1/file1.txt": "content1",
		"folder2/file2.txt": "content2",
	}

	if err := createTestZip(t, zipPath, testFiles); err != nil {
		t.Fatalf("failed to create test zip: %v", err)
	}

	extractDir := filepath.Join(tempDir, "extracted")
	if err := ExtractZip(zipPath, extractDir); err != nil {
		t.Fatalf("ExtractZip failed: %v", err)
	}

	for name, content := range testFiles {
		path := filepath.Join(extractDir, name)
		data, err := os.ReadFile(path)
		if err != nil {
			t.Errorf("failed to read extracted file %s: %v", name, err)
		}
		if string(data) != content {
			t.Errorf("content mismatch for %s", name)
		}
	}
}

func TestExtractZip_FailureScenarios(t *testing.T) {
	tempDir := t.TempDir()

	t.Run("file not found", func(t *testing.T) {
		extractDir := filepath.Join(tempDir, "extracted1")
		err := ExtractZip(filepath.Join(tempDir, "nonexistent.zip"), extractDir)
		if err == nil {
			t.Errorf("expected error for nonexistent file")
		}
	})

	t.Run("corrupted zip file", func(t *testing.T) {
		badZipPath := filepath.Join(tempDir, "bad.zip")
		os.WriteFile(badZipPath, []byte("not a real zip content"), 0644)
		extractDir := filepath.Join(tempDir, "extracted2")
		err := ExtractZip(badZipPath, extractDir)
		if err == nil {
			t.Errorf("expected error for corrupted zip")
		}
	})

	t.Run("empty zip path", func(t *testing.T) {
		extractDir := filepath.Join(tempDir, "extracted3")
		err := ExtractZip("", extractDir)
		if err == nil || !strings.Contains(err.Error(), "empty zip path") {
			t.Errorf("expected empty zip path error")
		}
	})

	t.Run("empty destination", func(t *testing.T) {
		zipPath := filepath.Join(tempDir, "test.zip")
		testFiles := map[string]string{"file.txt": "data"}
		createTestZip(t, zipPath, testFiles)
		err := ExtractZip(zipPath, "")
		if err == nil || !strings.Contains(err.Error(), "empty destination path") {
			t.Errorf("expected empty destination path error")
		}
	})

	t.Run("permission denied on destination", func(t *testing.T) {
		zipPath := filepath.Join(tempDir, "test2.zip")
		testFiles := map[string]string{"file.txt": "data"}
		createTestZip(t, zipPath, testFiles)

		extractDir := filepath.Join(tempDir, "extracted3")
		os.MkdirAll(extractDir, 0500)
		defer os.Chmod(extractDir, 0700)

		err := ExtractZip(zipPath, extractDir)
		if err == nil {
			t.Errorf("expected permission error")
		}
	})

	t.Run("invalid path traversal", func(t *testing.T) {
		zipPath := filepath.Join(tempDir, "evil.zip")
		createTestZip(t, zipPath, map[string]string{"../evil.txt": "attack"})
		extractDir := filepath.Join(tempDir, "extracted4")
		err := ExtractZip(zipPath, extractDir)
		if err == nil || !strings.Contains(err.Error(), "invalid file path detected") {
			t.Errorf("expected invalid path detection")
		}
	})
}

func createTestZip(t *testing.T, zipPath string, files map[string]string) error {
	zipFile, err := os.Create(zipPath)
	if err != nil {
		t.Fatalf("failed to create zip file: %v", err)
	}
	defer zipFile.Close()

	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	for name, content := range files {
		writer, err := zipWriter.Create(name)
		if err != nil {
			t.Fatalf("failed to create entry in zip: %v", err)
		}
		_, err = writer.Write([]byte(content))
		if err != nil {
			t.Fatalf("failed to write zip content: %v", err)
		}
	}
	return nil
}
