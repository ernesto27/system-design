package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
)

type PackageCopy struct {
	sourcePath string
	targetPath string
	packages   Packages
}

func newPackageCopy(sourcePath, targetPath string, packages Packages) *PackageCopy {
	return &PackageCopy{
		sourcePath: sourcePath,
		targetPath: targetPath,
		packages:   packages,
	}
}

func (pc *PackageCopy) copyPackages() error {
	if err := createDir(pc.targetPath); err != nil {
		return fmt.Errorf("failed to create node_modules directory: %v", err)
	}

	var wg sync.WaitGroup
	errChan := make(chan error, len(pc.packages))

	for pkgName, pkg := range pc.packages {
		wg.Add(1)
		go func(name string, p Package) {
			defer wg.Done()

			sourcePkgPath := filepath.Join(pc.sourcePath, fmt.Sprintf("%s@%s", name, p.Version))
			targetPkgPath := filepath.Join(pc.targetPath, name)

			if err := pc.copyDirectory(sourcePkgPath, targetPkgPath); err != nil {
				errChan <- fmt.Errorf("failed to copy package %s: %v", name, err)
				return
			}
			fmt.Printf("  ✓ Copied: %s@%s\n", name, p.Version)

			if err := pc.handleNestedDependencies(name, p, targetPkgPath); err != nil {
				errChan <- fmt.Errorf("failed to handle nested dependencies for %s: %v", name, err)
				return
			}
		}(pkgName, pkg)
	}

	// Close errChan when all goroutines complete
	go func() {
		wg.Wait()
		close(errChan)
	}()

	// Return immediately on first error
	for err := range errChan {
		if err != nil {
			return err
		}
	}

	fmt.Println("All packages copied successfully!")
	return nil
}

func (pc *PackageCopy) handleNestedDependencies(pkgName string, pkg Package, targetPkgPath string) error {
	for _, dep := range pkg.Dependencies {
		if dep.Nested {
			nestedNodeModules := filepath.Join(targetPkgPath, "node_modules")
			if err := createDir(nestedNodeModules); err != nil {
				return fmt.Errorf("failed to create nested node_modules: %v", err)
			}

			depSourcePath := filepath.Join(pc.sourcePath, fmt.Sprintf("%s@%s", dep.Name, dep.Version))
			depTargetPath := filepath.Join(nestedNodeModules, dep.Name)

			if err := pc.copyDirectory(depSourcePath, depTargetPath); err != nil {
				return fmt.Errorf("failed to copy nested dependency %s: %v", dep.Name, err)
			}
			fmt.Printf("    ↳ Nested: %s@%s -> %s/node_modules/%s\n", dep.Name, dep.Version, pkgName, dep.Name)
		}
	}
	return nil
}

func (pc *PackageCopy) copyDirectory(src, dst string) error {
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
			if err := pc.copyDirectory(srcPath, dstPath); err != nil {
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
