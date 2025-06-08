package e2e

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/jsas4coding/pma-up/internal/updater"
)

func TestRunUpdateE2E(t *testing.T) {
	// WARNING: this test runs against the official phpMyAdmin servers.

	sandboxDir := "./e2e-sandbox"
	pmaDir := filepath.Join(sandboxDir, "phpmyadmin")
	configPath := filepath.Join(pmaDir, "config.inc.php")

	// Clean sandbox to ensure a fresh test environment
	if err := os.RemoveAll(sandboxDir); err != nil {
		t.Fatalf("failed to clean up sandbox: %v", err)
	}

	// Prepare initial phpMyAdmin directory
	if err := os.MkdirAll(pmaDir, os.ModePerm); err != nil {
		t.Fatalf("failed to create phpMyAdmin directory: %v", err)
	}

	// Create a simulated config.inc.php file
	if err := os.WriteFile(configPath, []byte("dummy config"), 0644); err != nil {
		t.Fatalf("failed to create config file: %v", err)
	}

	// Execute real update
	if err := updater.RunUpdate(pmaDir, configPath); err != nil {
		t.Fatalf("RunUpdate failed: %v", err)
	}

	// Validate that config file was correctly restored
	finalConfig, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("failed to read restored config file: %v", err)
	}
	if string(finalConfig) != "dummy config" {
		t.Errorf("config file content mismatch: got '%s'", string(finalConfig))
	}

	// Validate that exactly one backup was created
	backups, err := filepath.Glob(filepath.Join(sandboxDir, "phpmyadmin_backup_*"))
	if err != nil {
		t.Fatalf("failed to list backups: %v", err)
	}
	if len(backups) != 1 {
		t.Errorf("expected 1 backup, found %d", len(backups))
	}
}
