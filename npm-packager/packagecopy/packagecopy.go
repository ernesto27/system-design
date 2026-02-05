package packagecopy

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

type PackageCopy struct {
}

func NewPackageCopy() *PackageCopy {
	return &PackageCopy{}
}

func (pc *PackageCopy) CopyDirectory(src, dst string) error {
	srcInfo, err := os.Stat(src)
	if err != nil {
		return fmt.Errorf("source does not exist: %v", err)
	}

	if !srcInfo.IsDir() {
		return fmt.Errorf("source is not a directory")
	}

	if err := os.MkdirAll(dst, srcInfo.Mode()); err != nil {
		return fmt.Errorf("failed to create destination directory: %v", err)
	}

	entries, err := os.ReadDir(src)
	if err != nil {
		return fmt.Errorf("failed to read source directory: %v", err)
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			if err := pc.CopyDirectory(srcPath, dstPath); err != nil {
				return err
			}
		} else {
			if err := pc.copyFile(srcPath, dstPath); err != nil {
				return err
			}
		}
	}

	return nil
}

func (pc *PackageCopy) copyFile(src, dst string) error {
	// Try hardlink first (fast, no copy, works with Node.js resolution)
	err := os.Link(src, dst)
	if err == nil {
		return nil
	}

	// Fallback to regular copy if hardlink fails (e.g., cross-device)
	srcFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open source file: %v", err)
	}
	defer srcFile.Close()

	srcInfo, err := srcFile.Stat()
	if err != nil {
		return fmt.Errorf("failed to stat source file: %v", err)
	}

	dstFile, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, srcInfo.Mode())
	if err != nil {
		return fmt.Errorf("failed to create destination file: %v", err)
	}
	defer dstFile.Close()

	if _, err := io.Copy(dstFile, srcFile); err != nil {
		return fmt.Errorf("failed to copy file contents: %v", err)
	}

	return nil
}
