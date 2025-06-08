package extractor

import (
	"archive/zip"
	"os"
	"path/filepath"
	"testing"
)

func TestExtractZip(t *testing.T) {
	tempDir := t.TempDir()

	// Create a test zip file with sample files
	zipPath := filepath.Join(tempDir, "test.zip")
	testFiles := map[string]string{
		"folder1/file1.txt": "content1",
		"folder2/file2.txt": "content2",
	}

	if err := createTestZip(t, zipPath, testFiles); err != nil {
		t.Fatalf("failed to create test zip: %v", err)
	}

	extractDir := filepath.Join(tempDir, "extracted")

	// Execute extraction
	if err := ExtractZip(zipPath, extractDir); err != nil {
		t.Fatalf("ExtractZip failed: %v", err)
	}

	// Validate that extracted files match expected contents
	for name, content := range testFiles {
		extractedPath := filepath.Join(extractDir, name)
		data, err := os.ReadFile(extractedPath)
		if err != nil {
			t.Errorf("failed to read extracted file %s: %v", name, err)
			continue
		}
		if string(data) != content {
			t.Errorf("file content mismatch for %s: expected '%s', got '%s'", name, content, string(data))
		}
	}
}

// createTestZip creates a zip file at the given path with provided files and contents.
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
		if _, err = writer.Write([]byte(content)); err != nil {
			return err
		}
	}

	return nil
}
