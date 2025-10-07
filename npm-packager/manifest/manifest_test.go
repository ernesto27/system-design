package manifest

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

// setupTestDirs creates temporary config directory for testing
// Note: NewManifest will create a "manifest" subdirectory inside this path
func setupTestDirs(t *testing.T) string {
	tmpDir := t.TempDir()
	return tmpDir
}

func TestDownloadManifest_Download(t *testing.T) {
	packageName := "express"

	testCases := []struct {
		name        string
		setupFunc   func(t *testing.T) (string, string, string)
		expectError bool
		validate    func(t *testing.T, m *Manifest, packageName string, statusCode int)
	}{
		{
			name: "Download express manifest without Etag",
			setupFunc: func(t *testing.T) (string, string, string) {
				configDir := setupTestDirs(t)
				return configDir, packageName, ""
			},
			expectError: false,
			validate: func(t *testing.T, m *Manifest, packageName string, statusCode int) {
				assert.Equal(t, http.StatusOK, statusCode, "Expected status code 200")

				expectedFile := filepath.Join(m.Path, packageName+".json")
				_, err := os.Stat(expectedFile)
				assert.NoError(t, err, "Manifest file should exist")

				info, err := os.Stat(expectedFile)
				assert.NoError(t, err)
				assert.Greater(t, info.Size(), int64(0), "File should not be empty")
			},
		},
		{
			name: "Download express manifest with Etag (not modified)",
			setupFunc: func(t *testing.T) (string, string, string) {
				configDir := setupTestDirs(t)

				manifest, err := NewManifest(configDir, "https://registry.npmjs.org/")
				assert.NoError(t, err)
				etag, _, err := manifest.Download(packageName, "")
				assert.NoError(t, err)
				fmt.Println(etag)

				return configDir, packageName, etag
			},
			expectError: false,
			validate: func(t *testing.T, m *Manifest, packageName string, statusCode int) {
				assert.Equal(t, http.StatusNotModified, statusCode, "Expected status code 304")

				expectedFile := filepath.Join(m.Path, packageName+".json")
				info, err := os.Stat(expectedFile)
				assert.NoError(t, err)
				assert.Greater(t, info.Size(), int64(0), "File should not be empty")
			},
		},
		{
			name: "Error with invalid package name",
			setupFunc: func(t *testing.T) (string, string, string) {
				configDir := setupTestDirs(t)
				return configDir, "this-package-does-not-exist-12345678", ""
			},
			expectError: true,
			validate: func(t *testing.T, m *Manifest, packageName string, statusCode int) {
				expectedFile := filepath.Join(m.Path, packageName+".json")
				info, err := os.Stat(expectedFile)
				if err == nil {
					assert.Equal(t, int64(0), info.Size(), "File should be empty or not exist")
				}
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			configDir, packageName, etag := tc.setupFunc(t)
			manifest, err := NewManifest(configDir, "https://registry.npmjs.org/")
			assert.NoError(t, err)
			etag, statusCode, err := manifest.Download(packageName, etag)

			if tc.expectError {
				assert.Error(t, err, "Expected an error")
			} else {
				assert.NoError(t, err, "Expected no error")
			}

			fmt.Println(etag)
			tc.validate(t, manifest, packageName, statusCode)
		})
	}
}
