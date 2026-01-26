package config

import (
	"os"
	"path/filepath"
)

type Config struct {
	// Base directories
	BaseDir     string
	ManifestDir string
	TarballDir  string
	PackagesDir string

	// Local installation paths
	LocalNodeModules string
	LocalBinDir      string

	// Global installation paths
	GlobalDir         string
	GlobalNodeModules string
	GlobalBinDir      string
	GlobalPackageJSON string
	GlobalLockFile    string
}

func New() (*Config, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	baseDir := filepath.Join(homeDir, ".config", "go-npm")
	globalDir := filepath.Join(baseDir, "global")

	return &Config{
		BaseDir:     baseDir,
		ManifestDir: filepath.Join(baseDir, "manifest"),
		TarballDir:  filepath.Join(baseDir, "tarball"),
		PackagesDir: filepath.Join(baseDir, "packages"),

		LocalNodeModules: "./node_modules",
		LocalBinDir:      "./node_modules/.bin",

		GlobalDir:         globalDir,
		GlobalNodeModules: filepath.Join(globalDir, "node_modules"),
		GlobalBinDir:      filepath.Join(globalDir, "bin"),
		GlobalPackageJSON: filepath.Join(globalDir, "package.json"),
		GlobalLockFile:    filepath.Join(globalDir, "go-package-lock.json"),
	}, nil
}
