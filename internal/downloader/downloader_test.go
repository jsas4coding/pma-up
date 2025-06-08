package downloader

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestDownloadPhpMyAdmin(t *testing.T) {
	mockZipContent := []byte("PK\x03\x04 dummy zip content")

	// Setup mock HTTP server to simulate phpMyAdmin download endpoint
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

	// Execute the download function
	downloadedFilePath, err := DownloadPhpMyAdmin(mockDownloadURL, tempDir, version)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Verify the file exists
	if _, err := os.Stat(downloadedFilePath); err != nil {
		t.Fatalf("expected file to exist, but got error: %v", err)
	}

	// Verify the file content matches mock data
	data, err := os.ReadFile(downloadedFilePath)
	if err != nil {
		t.Fatalf("failed to read downloaded file: %v", err)
	}

	if string(data) != string(mockZipContent) {
		t.Errorf("downloaded file content does not match expected mock content")
	}
}
