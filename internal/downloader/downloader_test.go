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
	if _, statErr := os.Stat(downloadedFilePath); statErr != nil {
		t.Fatalf("expected file to exist, but got error: %v", statErr)
	}

	// Verify the file content matches mock data
	data, readErr := os.ReadFile(downloadedFilePath)
	if readErr != nil {
		t.Fatalf("failed to read downloaded file: %v", readErr)
	}

	if string(data) != string(mockZipContent) {
		t.Errorf("downloaded file content does not match expected mock content")
	}
}

func TestDownloadPhpMyAdmin_FailureScenarios(t *testing.T) {
	tests := []struct {
		name       string
		serverFunc func() *httptest.Server
		setupDir   func() (string, error)
		expectErr  bool
	}{
		{
			name: "http 500 error",
			serverFunc: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
					http.Error(w, "internal server error", http.StatusInternalServerError)
				}))
			},
			setupDir: func() (string, error) {
				return t.TempDir(), nil
			},
			expectErr: true,
		},
		{
			name: "directory not writable",
			serverFunc: func() *httptest.Server {
				content := []byte("PK\x03\x04 dummy zip content")
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
					w.Header().Set("Content-Type", "application/zip")
					if _, err := w.Write(content); err != nil {
						t.Fatalf("failed to write mock zip content: %v", err)
					}
				}))
			},
			setupDir: func() (string, error) {
				dir := t.TempDir()
				if chmodErr := os.Chmod(dir, 0500); chmodErr != nil {
					return "", chmodErr
				}
				return dir, nil
			},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		tt := tt // capture range variable
		t.Run(tt.name, func(t *testing.T) {
			server := tt.serverFunc()
			defer server.Close()

			tempDir, dirErr := tt.setupDir()
			if dirErr != nil {
				t.Fatalf("failed to setup dir: %v", dirErr)
			}
			defer func() {
				// restore permission so tempdir can be cleaned up
				_ = os.Chmod(tempDir, 0700)
			}()

			version := "5.2.2"
			mockDownloadURL := fmt.Sprintf("%s/phpMyAdmin-%s-all-languages.zip", server.URL, version)

			_, downloadErr := DownloadPhpMyAdmin(mockDownloadURL, tempDir, version)
			if tt.expectErr && downloadErr == nil {
				t.Errorf("expected error but got none")
			}
			if !tt.expectErr && downloadErr != nil {
				t.Errorf("unexpected error: %v", downloadErr)
			}
		})
	}
}
