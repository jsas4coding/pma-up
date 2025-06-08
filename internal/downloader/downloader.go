// Package downloader handles downloading the phpMyAdmin distribution archive.
// It downloads the zip file from a provided URL and stores it in a given destination directory.
package downloader

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

// DownloadPhpMyAdmin downloads the phpMyAdmin zip file from the provided URL,
// saves it into the given destination directory using the specified version,
// and returns the full path to the downloaded file.
//
// Parameters:
//   - downloadURL: complete URL to download the phpMyAdmin zip file.
//   - destinationDir: directory where the file should be stored.
//   - version: version string, used for naming the output file.
//
// Returns:
//   - string: full path to the downloaded zip file.
//   - error: non-nil if the download or file creation fails.
func DownloadPhpMyAdmin(downloadURL, destinationDir, version string) (string, error) {
	if downloadURL == "" {
		return "", errors.New("empty download URL")
	}

	if destinationDir == "" {
		return "", errors.New("empty destination directory")
	}

	if version == "" {
		return "", errors.New("empty version string")
	}

	filename := fmt.Sprintf("phpMyAdmin-%s-all-languages.zip", version)
	filePath := filepath.Join(destinationDir, filename)

	client := &http.Client{
		Timeout: 60 * time.Second,
	}

	req, err := http.NewRequest("GET", downloadURL, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create HTTP request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to perform HTTP request: %w", err)
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			fmt.Printf("warning: failed to close response body: %v\n", closeErr)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected HTTP status: %d", resp.StatusCode)
	}

	outFile, err := os.Create(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to create file: %w", err)
	}
	defer func() {
		if closeErr := outFile.Close(); closeErr != nil {
			fmt.Printf("warning: failed to close file: %v\n", closeErr)
		}
	}()

	_, err = io.Copy(outFile, resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to write to file: %w", err)
	}

	return filePath, nil
}
