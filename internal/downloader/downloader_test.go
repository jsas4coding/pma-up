package downloader

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func TestDownloadPhpMyAdmin_Success(t *testing.T) {
	mockZipContent := []byte("PK\x03\x04 dummy zip content")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/zip")
		if _, err := w.Write(mockZipContent); err != nil {
			t.Fatalf("failed to write mock zip content: %v", err)
		}
	}))
	defer server.Close()

	tempDir := t.TempDir()
	version := "5.2.2"
	mockDownloadURL := fmt.Sprintf("%s/phpMyAdmin-%s-all-languages.zip", server.URL, version)

	filePath, err := DownloadPhpMyAdmin(mockDownloadURL, tempDir, version)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, err := os.Stat(filePath); err != nil {
		t.Fatalf("expected file, got stat error: %v", err)
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}

	if string(data) != string(mockZipContent) {
		t.Errorf("file content mismatch")
	}
}

func TestDownloadPhpMyAdmin_InputValidation(t *testing.T) {
	tests := []struct {
		name     string
		url      string
		dest     string
		ver      string
		experror string
	}{
		{"empty url", "", "/tmp", "5.2.2", "empty download URL"},
		{"empty dest", "http://example.com/file.zip", "", "5.2.2", "empty destination directory"},
		{"empty version", "http://example.com/file.zip", "/tmp", "", "empty version string"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := DownloadPhpMyAdmin(tc.url, tc.dest, tc.ver)
			if err == nil || !strings.Contains(err.Error(), tc.experror) {
				t.Errorf("expected error '%s', got '%v'", tc.experror, err)
			}
		})
	}
}

func TestDownloadPhpMyAdmin_RequestCreationFailure(t *testing.T) {
	tempDir := t.TempDir()
	_, err := DownloadPhpMyAdmin(":/invalid-url", tempDir, "5.2.2")
	if err == nil || !strings.Contains(err.Error(), "failed to create HTTP request") {
		t.Errorf("expected HTTP request creation error, got %v", err)
	}
}

func TestDownloadPhpMyAdmin_ClientFailure(t *testing.T) {
	tempDir := t.TempDir()
	_, err := DownloadPhpMyAdmin("http://nonexistent.invalid/file.zip", tempDir, "5.2.2")
	if err == nil || !strings.Contains(err.Error(), "failed to perform HTTP request") {
		t.Errorf("expected client failure, got %v", err)
	}
}

func TestDownloadPhpMyAdmin_ServerReturns500(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "internal error", http.StatusInternalServerError)
	}))
	defer server.Close()

	tempDir := t.TempDir()
	version := "5.2.2"
	mockDownloadURL := fmt.Sprintf("%s/phpMyAdmin-%s-all-languages.zip", server.URL, version)

	_, err := DownloadPhpMyAdmin(mockDownloadURL, tempDir, version)
	if err == nil || !strings.Contains(err.Error(), "unexpected HTTP status") {
		t.Errorf("expected HTTP status error, got %v", err)
	}
}

func TestDownloadPhpMyAdmin_DirectoryNotWritable(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/zip")
		w.Write([]byte("PK\x03\x04 dummy zip content"))
	}))
	defer server.Close()

	tempDir := t.TempDir()
	if err := os.Chmod(tempDir, 0500); err != nil {
		t.Fatalf("failed to chmod: %v", err)
	}
	defer os.Chmod(tempDir, 0700)

	version := "5.2.2"
	mockDownloadURL := fmt.Sprintf("%s/phpMyAdmin-%s-all-languages.zip", server.URL, version)

	_, err := DownloadPhpMyAdmin(mockDownloadURL, tempDir, version)
	if err == nil {
		t.Errorf("expected permission error, got none")
	}
}
