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

func TestRunUpdate_Success(t *testing.T) {
	tempDir := t.TempDir()

	// Simulated phpMyAdmin existing directory
	existingPmaDir := filepath.Join(tempDir, "phpmyadmin")
	if err := os.MkdirAll(existingPmaDir, os.ModePerm); err != nil {
		t.Fatalf("failed to create existing phpMyAdmin dir: %v", err)
	}

	// Simulated existing config.inc.php
	existingConfigPath := filepath.Join(existingPmaDir, "config.inc.php")
	if err := os.WriteFile(existingConfigPath, []byte("existing config"), 0644); err != nil {
		t.Fatalf("failed to create existing config: %v", err)
	}

	// Create mock zip archive
	mockZipPath := filepath.Join(tempDir, "mock_update.zip")
	files := map[string]string{
		"phpMyAdmin-5.2.2-all-languages/file.txt":       "new version",
		"phpMyAdmin-5.2.2-all-languages/config.inc.php": "should be replaced",
	}
	if err := createTestZip(t, mockZipPath, files); err != nil {
		t.Fatalf("failed to create test zip: %v", err)
	}

	// Mock download server
	downloadServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		f, openErr := os.Open(mockZipPath)
		if openErr != nil {
			t.Fatalf("failed to open mock zip: %v", openErr)
		}
		defer func() {
			if cerr := f.Close(); cerr != nil {
				t.Errorf("failed to close mock zip: %v", cerr)
			}
		}()
		if _, copyErr := io.Copy(w, f); copyErr != nil {
			t.Fatalf("failed to copy zip: %v", copyErr)
		}
	}))
	defer downloadServer.Close()

	// Mock version.txt server
	versionTxt := fmt.Sprintf("5.2.2\n2025-01-21\n%s/phpMyAdmin-5.2.2-all-languages.zip\n", downloadServer.URL)
	versionServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		if _, writeErr := fmt.Fprint(w, versionTxt); writeErr != nil {
			t.Fatalf("failed to write versionTxt: %v", writeErr)
		}
	}))
	defer versionServer.Close()

	// Override VersionURL for controlled environment
	originalVersionURL := version.VersionURL
	version.VersionURL = versionServer.URL
	defer func() { version.VersionURL = originalVersionURL }()

	// Execute real update
	if err := RunUpdate(existingPmaDir, existingConfigPath); err != nil {
		t.Fatalf("RunUpdate failed: %v", err)
	}

	// Verify file restored
	newFilePath := filepath.Join(existingPmaDir, "file.txt")
	data, readErr := os.ReadFile(newFilePath)
	if readErr != nil {
		t.Fatalf("failed to read extracted file: %v", readErr)
	}
	if string(data) != "new version" {
		t.Errorf("file content mismatch: expected 'new version', got '%s'", string(data))
	}

	// Verify config preserved
	finalConfigPath := filepath.Join(existingPmaDir, "config.inc.php")
	configData, configReadErr := os.ReadFile(finalConfigPath)
	if configReadErr != nil {
		t.Fatalf("failed to read restored config: %v", configReadErr)
	}
	if string(configData) != "existing config" {
		t.Errorf("config file not restored correctly: got '%s'", string(configData))
	}
}

func TestRunUpdate_FailureScenarios(t *testing.T) {
	tempDir := t.TempDir()

	t.Run("destination path invalid", func(t *testing.T) {
		invalidPath := filepath.Join(tempDir, "invalid/does/not/exist/phpmyadmin")
		configPath := filepath.Join(tempDir, "dummy-config.php")
		_ = os.WriteFile(configPath, []byte("dummy"), 0644)

		err := RunUpdate(invalidPath, configPath)
		if err == nil {
			t.Errorf("expected error for invalid destination path, got nil")
		}
	})

	t.Run("cannot create temp directory", func(t *testing.T) {
		// simulate by setting TMPDIR to invalid path
		invalidTmp := filepath.Join(tempDir, "no-permission")
		_ = os.Mkdir(invalidTmp, 0400)
		t.Setenv("TMPDIR", invalidTmp)

		destination := filepath.Join(tempDir, "phpmyadmin")
		configPath := filepath.Join(tempDir, "dummy-config.php")
		_ = os.WriteFile(configPath, []byte("dummy"), 0644)

		err := RunUpdate(destination, configPath)
		if err == nil {
			t.Errorf("expected error for tempdir failure, got nil")
		}
	})

	t.Run("missing version server", func(t *testing.T) {
		originalURL := version.VersionURL
		version.VersionURL = "http://127.0.0.1:1/non-existent"
		defer func() { version.VersionURL = originalURL }()

		destination := filepath.Join(tempDir, "phpmyadmin")
		_ = os.MkdirAll(destination, 0755)
		configPath := filepath.Join(destination, "config.inc.php")
		_ = os.WriteFile(configPath, []byte("dummy"), 0644)

		err := RunUpdate(destination, configPath)
		if err == nil {
			t.Errorf("expected error from unreachable version URL, got nil")
		}
	})
}

func TestRunUpdate_InvalidExtractedStructure(t *testing.T) {
	tempDir := t.TempDir()

	// Simulated phpMyAdmin existing directory
	existingPmaDir := filepath.Join(tempDir, "phpmyadmin")
	if err := os.MkdirAll(existingPmaDir, os.ModePerm); err != nil {
		t.Fatalf("failed to create existing phpMyAdmin dir: %v", err)
	}

	existingConfigPath := filepath.Join(existingPmaDir, "config.inc.php")
	if err := os.WriteFile(existingConfigPath, []byte("existing config"), 0644); err != nil {
		t.Fatalf("failed to create existing config: %v", err)
	}

	// Create zip with multiple subdirectories
	mockZipPath := filepath.Join(tempDir, "mock_bad_structure.zip")
	files := map[string]string{
		"phpMyAdmin-5.2.2-all-languages/file.txt": "new version",
		"another-folder/file2.txt":                "extra file",
	}
	if err := createTestZip(t, mockZipPath, files); err != nil {
		t.Fatalf("failed to create test zip: %v", err)
	}

	// Mock download server
	downloadServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		f, openErr := os.Open(mockZipPath)
		if openErr != nil {
			t.Fatalf("failed to open mock zip: %v", openErr)
		}
		defer func() {
			if cerr := f.Close(); cerr != nil {
				t.Errorf("failed to close mock zip: %v", cerr)
			}
		}()
		if _, copyErr := io.Copy(w, f); copyErr != nil {
			t.Fatalf("failed to copy zip: %v", copyErr)
		}
	}))
	defer downloadServer.Close()

	// Mock version.txt server
	versionTxt := fmt.Sprintf("5.2.2\n2025-01-21\n%s/phpMyAdmin-5.2.2-all-languages.zip\n", downloadServer.URL)
	versionServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		if _, writeErr := fmt.Fprint(w, versionTxt); writeErr != nil {
			t.Fatalf("failed to write versionTxt: %v", writeErr)
		}
	}))
	defer versionServer.Close()

	// Override VersionURL
	originalVersionURL := version.VersionURL
	version.VersionURL = versionServer.URL
	defer func() { version.VersionURL = originalVersionURL }()

	err := RunUpdate(existingPmaDir, existingConfigPath)
	if err == nil {
		t.Errorf("expected error for invalid extracted structure, got nil")
	}
}

func TestRunUpdate_ConfigRestoreFailure(t *testing.T) {
	tempDir := t.TempDir()

	// Prepare existing phpMyAdmin dir without config.inc.php to simulate restore failure
	existingPmaDir := filepath.Join(tempDir, "phpmyadmin")
	if err := os.MkdirAll(existingPmaDir, os.ModePerm); err != nil {
		t.Fatalf("failed to create existing phpMyAdmin dir: %v", err)
	}

	// Create valid zip
	mockZipPath := filepath.Join(tempDir, "mock_update.zip")
	files := map[string]string{
		"phpMyAdmin-5.2.2-all-languages/file.txt":       "new version",
		"phpMyAdmin-5.2.2-all-languages/config.inc.php": "should be replaced",
	}
	if err := createTestZip(t, mockZipPath, files); err != nil {
		t.Fatalf("failed to create test zip: %v", err)
	}

	// Mock download server
	downloadServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		f, openErr := os.Open(mockZipPath)
		if openErr != nil {
			t.Fatalf("failed to open mock zip: %v", openErr)
		}
		defer func() {
			if cerr := f.Close(); cerr != nil {
				t.Errorf("failed to close mock zip: %v", cerr)
			}
		}()
		if _, copyErr := io.Copy(w, f); copyErr != nil {
			t.Fatalf("failed to copy zip: %v", copyErr)
		}
	}))
	defer downloadServer.Close()

	// Mock version.txt server
	versionTxt := fmt.Sprintf("5.2.2\n2025-01-21\n%s/phpMyAdmin-5.2.2-all-languages.zip\n", downloadServer.URL)
	versionServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		if _, writeErr := fmt.Fprint(w, versionTxt); writeErr != nil {
			t.Fatalf("failed to write versionTxt: %v", writeErr)
		}
	}))
	defer versionServer.Close()

	originalVersionURL := version.VersionURL
	version.VersionURL = versionServer.URL
	defer func() { version.VersionURL = originalVersionURL }()

	// Intencionalmente não criaremos o config.inc.php original → vai falhar ao restaurar
	err := RunUpdate(existingPmaDir, filepath.Join(existingPmaDir, "config.inc.php"))
	if err == nil {
		t.Errorf("expected error when restoring missing config, got nil")
	}
}

// Hardening helper — fully linter safe
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
