package version

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

// replace the VersionURL for test purposes
var originalVersionURL = VersionURL

func TestFetchLatestVersion(t *testing.T) {
	mockContent := "5.2.2\n2025-01-21\nhttps://files.phpmyadmin.net/phpMyAdmin/5.2.2/phpMyAdmin-5.2.2-all-languages.zip\n"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, mockContent)
	}))
	defer server.Close()

	// Override the constant versionURL temporarily
	VersionURL = server.URL
	defer func() { VersionURL = originalVersionURL }()

	got, err := FetchLatestVersion()
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if got.Version != "5.2.2" {
		t.Errorf("Expected version '5.2.2', got %s", got.Version)
	}

	if got.Date != "2025-01-21" {
		t.Errorf("Expected date '2025-01-21', got %s", got.Date)
	}

	expectedURL := "https://files.phpmyadmin.net/phpMyAdmin/5.2.2/phpMyAdmin-5.2.2-all-languages.zip"
	if got.URL != expectedURL {
		t.Errorf("Expected URL '%s', got '%s'", expectedURL, got.URL)
	}
}
