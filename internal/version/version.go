// Package version handles fetching and parsing the latest phpMyAdmin release information.
// It retrieves the current release version, release date, and download URL from the phpMyAdmin version endpoint.
package version

import (
	"bufio"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// VersionURL defines the URL endpoint where phpMyAdmin publishes the latest version information.
// This file is fetched and parsed to retrieve release version, release date, and download URL.
var VersionURL = "https://www.phpmyadmin.net/home_page/version.txt"

// PhpMyAdminVersion represents the phpMyAdmin release information
// fetched from the version.txt endpoint.
type PhpMyAdminVersion struct {
	Version string // Release version (e.g. "5.2.2")
	Date    string // Release date (e.g. "2025-01-21")
	URL     string // Full URL to the downloadable archive
}

// FetchLatestVersion retrieves the latest phpMyAdmin version information.
//
// It downloads and parses the version.txt file from the phpMyAdmin site,
// returning the release version, release date, and download URL.
//
// Returns:
//   - *PhpMyAdminVersion: parsed release information
//   - error: non-nil if fetching or parsing fails.
func FetchLatestVersion() (*PhpMyAdminVersion, error) {
	client := &http.Client{
		Timeout: 15 * time.Second,
	}

	req, err := http.NewRequest("GET", VersionURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to perform HTTP request: %w", err)
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			fmt.Printf("warning: failed to close response body: %v\n", closeErr)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected HTTP status: %d", resp.StatusCode)
	}

	scanner := bufio.NewScanner(resp.Body)
	lines := []string{}
	for scanner.Scan() {
		lines = append(lines, strings.TrimSpace(scanner.Text()))
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to scan response body: %w", err)
	}

	if len(lines) < 3 {
		return nil, errors.New("unexpected version.txt format")
	}

	version := &PhpMyAdminVersion{
		Version: lines[0],
		Date:    lines[1],
		URL:     lines[2],
	}

	return version, nil
}
