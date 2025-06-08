// Package updater performs the full phpMyAdmin update process.
// It handles version detection, download, extraction, backup, replacement
// and configuration preservation for phpMyAdmin installations.
package updater

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/jsas4coding/pma-up/internal/downloader"
	"github.com/jsas4coding/pma-up/internal/extractor"
	"github.com/jsas4coding/pma-up/internal/fs"
	"github.com/jsas4coding/pma-up/internal/version"
)

// RunUpdate performs the phpMyAdmin update process.
//
// It downloads the latest phpMyAdmin release, extracts its content, backs up
// the current installation, replaces the old version with the new one, and
// restores the existing configuration file.
//
// Parameters:
//   - destinationPath: absolute path where phpMyAdmin is installed.
//   - configFilePath: absolute path to the phpMyAdmin configuration file.
//
// Returns:
//   - error: non-nil if any operation fails during the update process.
func RunUpdate(destinationPath, configFilePath string) error {
	fmt.Println("Starting phpMyAdmin update process...")

	latestVersion, err := version.FetchLatestVersion()
	if err != nil {
		return fmt.Errorf("failed to fetch latest version: %w", err)
	}
	fmt.Printf("Latest version: %s (%s)\n", latestVersion.Version, latestVersion.Date)

	tempDir, err := os.MkdirTemp("", "pma-up-*")
	if err != nil {
		return fmt.Errorf("failed to create temporary directory: %w", err)
	}
	defer func() {
		if removeErr := os.RemoveAll(tempDir); removeErr != nil {
			fmt.Printf("warning: failed to remove temp directory: %v\n", removeErr)
		}
	}()

	zipFilePath, err := downloader.DownloadPhpMyAdmin(latestVersion.URL, tempDir, latestVersion.Version)
	if err != nil {
		return fmt.Errorf("failed to download phpMyAdmin: %w", err)
	}

	extractDir := filepath.Join(tempDir, "extracted")
	if err := os.MkdirAll(extractDir, os.ModePerm); err != nil {
		return fmt.Errorf("failed to create extraction directory: %w", err)
	}

	if err := extractor.ExtractZip(zipFilePath, extractDir); err != nil {
		return fmt.Errorf("failed to extract phpMyAdmin: %w", err)
	}

	subDirs, err := os.ReadDir(extractDir)
	if err != nil {
		return fmt.Errorf("failed to read extraction directory: %w", err)
	}
	if len(subDirs) != 1 {
		return fmt.Errorf("unexpected extracted directory structure")
	}

	extractedContentPath := filepath.Join(extractDir, subDirs[0].Name())

	backupPath := fmt.Sprintf("%s_backup_%d", destinationPath, time.Now().Unix())
	if err := fs.MoveDir(destinationPath, backupPath); err != nil {
		return fmt.Errorf("failed to backup existing phpMyAdmin: %w", err)
	}

	if err := fs.MoveDir(extractedContentPath, destinationPath); err != nil {
		return fmt.Errorf("failed to move new phpMyAdmin to destination: %w", err)
	}

	originalConfigPath := filepath.Join(backupPath, filepath.Base(configFilePath))
	newConfigPath := filepath.Join(destinationPath, filepath.Base(configFilePath))

	if err := fs.CopyFile(originalConfigPath, newConfigPath); err != nil {
		return fmt.Errorf("failed to restore config file: %w", err)
	}

	fmt.Println("phpMyAdmin update process completed successfully.")
	return nil
}
