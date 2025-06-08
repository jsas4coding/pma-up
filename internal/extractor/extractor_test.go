package extractor

import (
	"archive/zip"
	"os"
	"path/filepath"
	"testing"
)

func TestExtractZip(t *testing.T) {
	tempDir := t.TempDir()

	// Create test zip file
	zipPath := filepath.Join(tempDir, "test.zip")
	testFiles := map[string]string{
		"folder1/file1.txt": "content1",
		"folder2/file2.txt": "content2",
	}

	createErr := createTestZip(t, zipPath, testFiles)
	if createErr != nil {
		t.Fatalf("failed to create test zip: %v", createErr)
	}

	extractDir := filepath.Join(tempDir, "extracted")

	extractErr := ExtractZip(zipPath, extractDir)
	if extractErr != nil {
		t.Fatalf("ExtractZip failed: %v", extractErr)
	}

	for name, content := range testFiles {
		extractedPath := filepath.Join(extractDir, name)
		data, readErr := os.ReadFile(extractedPath)
		if readErr != nil {
			t.Errorf("failed to read extracted file %s: %v", name, readErr)
			continue
		}
		if string(data) != content {
			t.Errorf("file content mismatch for %s: expected '%s', got '%s'", name, content, string(data))
		}
	}
}

func TestExtractZip_FailureScenarios(t *testing.T) {
	tempDir := t.TempDir()

	t.Run("file not found", func(t *testing.T) {
		invalidPath := filepath.Join(tempDir, "nonexistent.zip")
		extractDir := filepath.Join(tempDir, "extracted1")

		err := ExtractZip(invalidPath, extractDir)
		if err == nil {
			t.Errorf("expected error when opening nonexistent file, got nil")
		}
	})

	t.Run("corrupted zip file", func(t *testing.T) {
		badZipPath := filepath.Join(tempDir, "bad.zip")
		writeErr := os.WriteFile(badZipPath, []byte("not a real zip content"), 0644)
		if writeErr != nil {
			t.Fatalf("failed to write corrupted zip: %v", writeErr)
		}

		extractDir := filepath.Join(tempDir, "extracted2")
		err := ExtractZip(badZipPath, extractDir)
		if err == nil {
			t.Errorf("expected error when extracting corrupted zip, got nil")
		}
	})
}

// Hardening helper - fully linter safe
func createTestZip(t *testing.T, zipPath string, files map[string]string) error {
	zipFile, err := os.Create(zipPath)
	if err != nil {
		return err
	}
	defer func() {
		if cerr := zipFile.Close(); cerr != nil {
			t.Errorf("failed to close zipFile: %v", cerr)
		}
	}()

	zipWriter := zip.NewWriter(zipFile)
	defer func() {
		if cerr := zipWriter.Close(); cerr != nil {
			t.Errorf("failed to close zipWriter: %v", cerr)
		}
	}()

	for name, content := range files {
		writer, err := zipWriter.Create(name)
		if err != nil {
			return err
		}
		if _, err := writer.Write([]byte(content)); err != nil {
			return err
		}
	}

	return nil
}
