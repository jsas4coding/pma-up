// Package extractor handles extracting the phpMyAdmin zip archive.
// It unpacks the provided zip file into the specified destination directory while preserving directory structure.
package extractor

import (
	"archive/zip"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// ExtractZip extracts the contents of a phpMyAdmin zip archive into the given destination directory.
//
// Parameters:
//   - zipPath: full path to the zip archive to extract.
//   - destination: target directory where the contents will be extracted.
//
// Returns:
//   - error: non-nil if extraction fails.
func ExtractZip(zipPath, destination string) error {
	if zipPath == "" {
		return errors.New("empty zip path")
	}
	if destination == "" {
		return errors.New("empty destination path")
	}

	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return fmt.Errorf("failed to open zip file: %w", err)
	}
	defer func() {
		if closeErr := r.Close(); closeErr != nil {
			fmt.Printf("warning: failed to close zip reader: %v\n", closeErr)
		}
	}()

	for _, file := range r.File {
		filePath := filepath.Join(destination, file.Name)

		if !strings.HasPrefix(filePath, filepath.Clean(destination)+string(os.PathSeparator)) {
			return fmt.Errorf("invalid file path detected: %s", filePath)
		}

		if file.FileInfo().IsDir() {
			if err := os.MkdirAll(filePath, os.ModePerm); err != nil {
				return fmt.Errorf("failed to create directory: %w", err)
			}
			continue
		}

		srcFile, err := file.Open()
		if err != nil {
			return fmt.Errorf("failed to open file inside zip: %w", err)
		}
		defer func(f io.ReadCloser) {
			if closeErr := f.Close(); closeErr != nil {
				fmt.Printf("warning: failed to close file inside zip: %v\n", closeErr)
			}
		}(srcFile)

		if err := os.MkdirAll(filepath.Dir(filePath), os.ModePerm); err != nil {
			return fmt.Errorf("failed to create parent directories: %w", err)
		}

		dstFile, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, file.Mode())
		if err != nil {
			return fmt.Errorf("failed to create destination file: %w", err)
		}
		defer func(f *os.File) {
			if closeErr := f.Close(); closeErr != nil {
				fmt.Printf("warning: failed to close destination file: %v\n", closeErr)
			}
		}(dstFile)

		if _, err := io.Copy(dstFile, srcFile); err != nil {
			return fmt.Errorf("failed to copy file data: %w", err)
		}
	}

	return nil
}
