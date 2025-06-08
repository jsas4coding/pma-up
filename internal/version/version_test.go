package version

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestFetchLatestVersion(t *testing.T) {
	mockContent := "5.2.2\n2025-01-21\nhttps://files.phpmyadmin.net/phpMyAdmin/5.2.2/phpMyAdmin-5.2.2-all-languages.zip\n"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		if _, err := fmt.Fprint(w, mockContent); err != nil {
			t.Fatalf("failed to write mockContent: %v", err)
		}
	}))
	defer server.Close()

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

func TestFetchLatestVersion_FailureScenarios(t *testing.T) {
	t.Run("unreachable version server", func(t *testing.T) {
		originalURL := VersionURL
		VersionURL = "http://127.0.0.1:1/non-existent"
		defer func() { VersionURL = originalURL }()

		_, err := FetchLatestVersion()
		if err == nil {
			t.Errorf("expected error from unreachable version URL, got nil")
		}
	})

	t.Run("invalid response format - incomplete lines", func(t *testing.T) {
		mockContent := "5.2.2\n2025-01-21\n"

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			if _, err := fmt.Fprint(w, mockContent); err != nil {
				t.Fatalf("failed to write mockContent: %v", err)
			}
		}))
		defer server.Close()

		originalURL := VersionURL
		VersionURL = server.URL
		defer func() { VersionURL = originalURL }()

		_, err := FetchLatestVersion()
		if err == nil {
			t.Errorf("expected error for incomplete response, got nil")
		}
	})

	t.Run("invalid response format - empty body", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
			// empty body
		}))
		defer server.Close()

		originalURL := VersionURL
		VersionURL = server.URL
		defer func() { VersionURL = originalURL }()

		_, err := FetchLatestVersion()
		if err == nil {
			t.Errorf("expected error for empty response, got nil")
		}
	})

	t.Run("invalid URL syntax triggers NewRequest failure", func(t *testing.T) {
		originalURL := VersionURL
		VersionURL = ":/invalid-url"
		defer func() { VersionURL = originalURL }()

		_, err := FetchLatestVersion()
		if err == nil || !strings.Contains(err.Error(), "failed to create HTTP request") {
			t.Errorf("expected HTTP request creation error, got %v", err)
		}
	})

	// t.Run("scanner failure triggers scanner.Err()", func(t *testing.T) {
	// 	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
	// 		// Simulate incomplete write and close connection to cause scanner error
	// 		hj, ok := w.(http.Hijacker)
	// 		if !ok {
	// 			t.Fatalf("server does not support hijacking")
	// 		}
	// 		conn, _, err := hj.Hijack()
	// 		if err != nil {
	// 			t.Fatalf("failed to hijack connection: %v", err)
	// 		}
	// 		if cerr := conn.Close(); cerr != nil {
	// 			t.Errorf("failed to close hijacked connection: %v", cerr)
	// 		}
	// 	}))
	// 	defer server.Close()
	//
	// 	originalURL := VersionURL
	// 	VersionURL = server.URL
	// 	defer func() { VersionURL = originalURL }()
	//
	// 	_, err := FetchLatestVersion()
	// 	if err == nil || !strings.Contains(err.Error(), "failed to scan response body") {
	// 		t.Errorf("expected scanner failure, got %v", err)
	// 	}
	// })
}
