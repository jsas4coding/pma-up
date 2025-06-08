package updater

import (
	"archive/zip"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/jsas4coding/pma-up/internal/version"
)

func TestRunUpdate(t *testing.T) {
	tempDir := t.TempDir()

	// Create a simulated existing phpMyAdmin directory
	existingPmaDir := filepath.Join(tempDir, "phpmyadmin")
	if err := os.MkdirAll(existingPmaDir, os.ModePerm); err != nil {
		t.Fatalf("failed to create existing phpMyAdmin directory: %v", err)
	}

	// Create a simulated config.inc.php file
	existingConfigPath := filepath.Join(existingPmaDir, "config.inc.php")
	if err := os.WriteFile(existingConfigPath, []byte("existing config"), 0644); err != nil {
		t.Fatalf("failed to create existing config file: %v", err)
	}

	// Create the update zip archive
	mockZipPath := filepath.Join(tempDir, "mock_update.zip")
	files := map[string]string{
		"phpMyAdmin-5.2.2-all-languages/file.txt":       "new version",
		"phpMyAdmin-5.2.2-all-languages/config.inc.php": "should be replaced",
	}
	if err := createTestZip(t, mockZipPath, files); err != nil {
		t.Fatalf("failed to create test zip: %v", err)
	}

	// Mock the download server
	downloadServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		f, err := os.Open(mockZipPath)
		if err != nil {
			t.Fatalf("failed to open mock zip: %v", err)
		}
		defer func() {
			if cerr := f.Close(); cerr != nil {
				t.Errorf("failed to close mock zip: %v", cerr)
			}
		}()
		if _, err := io.Copy(w, f); err != nil {
			t.Fatalf("failed to copy zip to response: %v", err)
		}
	}))
	defer downloadServer.Close()

	// Mock the version.txt server
	versionTxt := fmt.Sprintf("5.2.2\n2025-01-21\n%s/phpMyAdmin-5.2.2-all-languages.zip\n", downloadServer.URL)
	versionServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		if _, err := fmt.Fprint(w, versionTxt); err != nil {
			t.Fatalf("failed to write versionTxt: %v", err)
		}
	}))
	defer versionServer.Close()

	// Override VersionURL to point to mock
	originalVersionURL := version.VersionURL
	version.VersionURL = versionServer.URL
	defer func() { version.VersionURL = originalVersionURL }()

	// Run the real update
	if err := RunUpdate(existingPmaDir, existingConfigPath); err != nil {
		t.Fatalf("RunUpdate failed: %v", err)
	}

	// Verify extracted file
	newFilePath := filepath.Join(existingPmaDir, "file.txt")
	data, err := os.ReadFile(newFilePath)
	if err != nil {
		t.Fatalf("failed to read extracted file: %v", err)
	}
	if string(data) != "new version" {
		t.Errorf("extracted file content mismatch: expected 'new version', got '%s'", string(data))
	}

	// Verify config file was preserved
	finalConfigPath := filepath.Join(existingPmaDir, "config.inc.php")
	configData, err := os.ReadFile(finalConfigPath)
	if err != nil {
		t.Fatalf("failed to read restored config file: %v", err)
	}
	if string(configData) != "existing config" {
		t.Errorf("config file not restored correctly: got '%s'", string(configData))
	}
}

// createTestZip creates a zip archive for testing purposes.
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
