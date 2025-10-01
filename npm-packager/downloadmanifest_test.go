package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDownloadManifest_Download(t *testing.T) {
	testCases := []struct {
		name        string
		setupFunc   func(t *testing.T) (string, string, string)
		expectError bool
		validate    func(t *testing.T, manifestPath string, packageName string)
	}{
		{
			name: "Download express manifest",
			setupFunc: func(t *testing.T) (string, string, string) {
				tmpDir := t.TempDir()
				manifestDir := filepath.Join(tmpDir, "manifests")
				etagDir := filepath.Join(tmpDir, "etag")
				os.MkdirAll(manifestDir, 0755)
				os.MkdirAll(etagDir, 0755)
				return manifestDir, etagDir, "express"
			},
			expectError: false,
			validate: func(t *testing.T, manifestPath string, packageName string) {
				expectedFile := filepath.Join(manifestPath, packageName+".json")
				_, err := os.Stat(expectedFile)
				assert.NoError(t, err, "Manifest file should exist")

				info, err := os.Stat(expectedFile)
				assert.NoError(t, err)
				assert.Greater(t, info.Size(), int64(0), "File should not be empty")
			},
		},
		{
			name: "Skip download if express manifest already exists",
			setupFunc: func(t *testing.T) (string, string, string) {
				tmpDir := t.TempDir()
				manifestDir := filepath.Join(tmpDir, "manifests")
				etagDir := filepath.Join(tmpDir, "etag")
				os.MkdirAll(manifestDir, 0755)
				os.MkdirAll(etagDir, 0755)

				manifestFile := filepath.Join(manifestDir, "express.json")
				os.WriteFile(manifestFile, []byte(`{"name":"express"}`), 0644)

				return manifestDir, etagDir, "express"
			},
			expectError: false,
			validate: func(t *testing.T, manifestPath string, packageName string) {
				expectedFile := filepath.Join(manifestPath, packageName+".json")
				content, err := os.ReadFile(expectedFile)
				assert.NoError(t, err)
				assert.Contains(t, string(content), `"name":"express"`)
			},
		},
		{
			name: "Error with invalid package name",
			setupFunc: func(t *testing.T) (string, string, string) {
				tmpDir := t.TempDir()
				manifestDir := filepath.Join(tmpDir, "manifests")
				etagDir := filepath.Join(tmpDir, "etag")
				os.MkdirAll(manifestDir, 0755)
				os.MkdirAll(etagDir, 0755)
				return manifestDir, etagDir, "this-package-does-not-exist-12345678"
			},
			expectError: true,
			validate: func(t *testing.T, manifestPath string, packageName string) {
				expectedFile := filepath.Join(manifestPath, packageName+".json")
				info, err := os.Stat(expectedFile)
				if err == nil {
					assert.Equal(t, int64(0), info.Size(), "File should be empty or not exist")
				}
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			manifestPath, etagPath, packageName := tc.setupFunc(t)
			manifest := newDownloadManifest(packageName, manifestPath, etagPath)
			_, err := manifest.download("")

			if tc.expectError {
				assert.Error(t, err, "Expected an error")
			} else {
				assert.NoError(t, err, "Expected no error")
			}

			tc.validate(t, manifestPath, packageName)
		})
	}
}
