package version

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestFetchLatestVersion(t *testing.T) {
	mockContent := "5.2.2\n2025-01-21\nhttps://files.phpmyadmin.net/phpMyAdmin/5.2.2/phpMyAdmin-5.2.2-all-languages.zip\n"

	// Setup mock server to simulate version.txt endpoint
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		if _, err := fmt.Fprint(w, mockContent); err != nil {
			t.Fatalf("failed to write mockContent: %v", err)
		}
	}))
	defer server.Close()

	// Backup and override VersionURL for test
	originalVersionURL := VersionURL
	VersionURL = server.URL
	defer func() { VersionURL = originalVersionURL }()

	got, err := FetchLatestVersion()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if got.Version != "5.2.2" {
		t.Errorf("expected version '5.2.2', got %s", got.Version)
	}

	if got.Date != "2025-01-21" {
		t.Errorf("expected date '2025-01-21', got %s", got.Date)
	}

	expectedURL := "https://files.phpmyadmin.net/phpMyAdmin/5.2.2/phpMyAdmin-5.2.2-all-languages.zip"
	if got.URL != expectedURL {
		t.Errorf("expected URL '%s', got '%s'", expectedURL, got.URL)
	}
}
