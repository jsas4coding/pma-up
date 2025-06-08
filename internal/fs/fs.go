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

// Injected functions for deterministic fault injection
type injFunc struct {
	Rename    func(string, string) error
	RemoveAll func(string) error
	MkdirAll  func(string, os.FileMode) error
	Open      func(string) (*os.File, error)
	OpenFile  func(string, int, os.FileMode) (*os.File, error)
	Copy      func(io.Writer, io.Reader) (int64, error)
	Rel       func(string, string) (string, error)
	Walk      func(string, filepath.WalkFunc) error
	Stat      func(string) (os.FileInfo, error)
}

var inj = injFunc{
	Rename:    os.Rename,
	RemoveAll: os.RemoveAll,
	MkdirAll:  os.MkdirAll,
	Open:      os.Open,
	OpenFile:  os.OpenFile,
	Copy:      io.Copy,
	Rel:       filepath.Rel,
	Walk:      filepath.Walk,
}

// MoveDir moves a directory from source to destination.
//
// If the source and destination are on different devices, MoveDir transparently falls back
// to a recursive copy followed by removal of the source directory.
func MoveDir(source, dest string) error {
	err := inj.Rename(source, dest)
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

	if err := inj.RemoveAll(source); err != nil {
		return wrap("failed to cleanup source after copy", err)
	}

	return nil
}

// CopyFile copies a single file from src to dst.
func CopyFile(src, dst string) error {
	sourceFileStat, err := os.Stat(src)
	if err != nil {
		return wrap("failed to stat source file", err)
	}

	if !sourceFileStat.Mode().IsRegular() {
		return wrap("source file is not regular", nil)
	}

	source, err := inj.Open(src)
	if err != nil {
		return wrap("failed to open source file", err)
	}
	defer safeClose("source", source)

	destination, err := inj.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, sourceFileStat.Mode())
	if err != nil {
		return wrap("failed to create destination file", err)
	}
	defer safeClose("destination", destination)

	_, err = inj.Copy(destination, source)
	if err != nil {
		return wrap("failed to copy file content", err)
	}

	return nil
}

func copyDir(source, dest string) error {
	return inj.Walk(source, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := inj.Rel(source, path)
		if err != nil {
			return err
		}
		targetPath := filepath.Join(dest, relPath)

		if info.IsDir() {
			return inj.MkdirAll(targetPath, info.Mode())
		}

		srcFile, err := inj.Open(path)
		if err != nil {
			return err
		}
		defer safeClose("source file", srcFile)

		destFile, err := inj.OpenFile(targetPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, info.Mode())
		if err != nil {
			return err
		}
		defer safeClose("dest file", destFile)

		_, err = inj.Copy(destFile, srcFile)
		return err
	})
}

func safeClose(label string, c io.Closer) {
	if err := c.Close(); err != nil {
		log.Printf("warning: failed to close %s: %v", label, err)
	}
}

func wrap(msg string, err error) error {
	if err == nil {
		return errors.New(msg)
	}
	return errors.Join(errors.New(msg), err)
}
