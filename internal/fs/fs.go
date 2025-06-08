// Package fs provides file system operations used in the update process.
//
// It offers utilities to move directories (handling cross-device moves),
// copy entire directory trees, and copy individual files with full error handling.
//
// The functions in this package ensure safe and reliable file system manipulations,
// preserving file permissions and handling edge cases across different platforms.
package fs

import (
	"errors"
	"io"
	"log"
	"os"
	"path/filepath"
	"syscall"
)

// MoveDir moves a directory from source to destination.
//
// If the source and destination are on different devices, MoveDir transparently falls back
// to a recursive copy followed by removal of the source directory.
//
// Parameters:
//   - source: full path of the source directory to move.
//   - dest: full path of the destination directory.
//
// Returns:
//   - error: non-nil if the move or copy operation fails.
func MoveDir(source, dest string) error {
	err := os.Rename(source, dest)
	if err == nil {
		return nil
	}

	linkErr, ok := err.(*os.LinkError)
	if !ok || linkErr.Err != syscall.EXDEV {
		return wrap("rename failed", err)
	}

	if err := copyDir(source, dest); err != nil {
		return wrap("copyDir failed", err)
	}

	if err := os.RemoveAll(source); err != nil {
		return wrap("failed to cleanup source after copy", err)
	}

	return nil
}

// CopyFile copies a single file from src to dst.
//
// It fully preserves the file contents and permissions. Errors are returned
// if any part of the copy operation fails.
//
// Parameters:
//   - src: full path to the source file.
//   - dst: full path to the destination file.
//
// Returns:
//   - error: non-nil if the copy operation fails.
func CopyFile(src, dst string) error {
	sourceFileStat, err := os.Stat(src)
	if err != nil {
		return wrap("failed to stat source file", err)
	}

	if !sourceFileStat.Mode().IsRegular() {
		return wrap("source file is not regular", nil)
	}

	source, err := os.Open(src)
	if err != nil {
		return wrap("failed to open source file", err)
	}
	defer func() {
		if closeErr := source.Close(); closeErr != nil {
			log.Printf("warning: failed to close source file: %v", closeErr)
		}
	}()

	destination, err := os.Create(dst)
	if err != nil {
		return wrap("failed to create destination file", err)
	}
	defer func() {
		if closeErr := destination.Close(); closeErr != nil {
			log.Printf("warning: failed to close destination file: %v", closeErr)
		}
	}()

	_, err = io.Copy(destination, source)
	if err != nil {
		return wrap("failed to copy file content", err)
	}

	return nil
}

func copyDir(source, dest string) error {
	return filepath.Walk(source, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(source, path)
		if err != nil {
			return err
		}
		targetPath := filepath.Join(dest, relPath)

		if info.IsDir() {
			return os.MkdirAll(targetPath, info.Mode())
		}

		srcFile, err := os.Open(path)
		if err != nil {
			return err
		}
		defer func() {
			if closeErr := srcFile.Close(); closeErr != nil {
				log.Printf("warning: failed to close source file: %v", closeErr)
			}
		}()

		destFile, err := os.OpenFile(targetPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, info.Mode())
		if err != nil {
			return err
		}
		defer func() {
			if closeErr := destFile.Close(); closeErr != nil {
				log.Printf("warning: failed to close destination file: %v", closeErr)
			}
		}()

		_, err = io.Copy(destFile, srcFile)
		return err
	})
}

func wrap(msg string, err error) error {
	if err == nil {
		return errors.New(msg)
	}
	return errors.Join(errors.New(msg), err)
}
