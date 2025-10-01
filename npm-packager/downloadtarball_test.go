package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

// setupTestTarballDirs creates temporary tarball directory for testing
func setupTestTarballDirs(t *testing.T) string {
	tmpDir := t.TempDir()
	tarballDir := filepath.Join(tmpDir, "tarballs")
	os.MkdirAll(tarballDir, 0755)
	return tarballDir
}

func TestDownloadTarball_Download(t *testing.T) {
	testCases := []struct {
		name        string
		setupFunc   func(t *testing.T) (string, string)
		expectError bool
		validate    func(t *testing.T, tarballPath string, url string, err error)
	}{
		{
			name: "Download express tarball successfully",
			setupFunc: func(t *testing.T) (string, string) {
				tarballDir := setupTestTarballDirs(t)
				url := "https://registry.npmjs.org/express/-/express-4.18.2.tgz"
				return tarballDir, url
			},
			expectError: false,
			validate: func(t *testing.T, tarballPath string, url string, err error) {
				assert.NoError(t, err, "Download should succeed")

				expectedFile := filepath.Join(tarballPath, "express-4.18.2.tgz")
				info, statErr := os.Stat(expectedFile)
				assert.NoError(t, statErr, "Tarball file should exist")
				assert.Greater(t, info.Size(), int64(0), "File should not be empty")
			},
		},
		{
			name: "Error with invalid tarball URL",
			setupFunc: func(t *testing.T) (string, string) {
				tarballDir := setupTestTarballDirs(t)
				url := "https://registry.npmjs.org/invalid-package-12345678/-/invalid-package-12345678-1.0.0.tgz"
				return tarballDir, url
			},
			expectError: true,
			validate: func(t *testing.T, tarballPath string, url string, err error) {
				assert.Error(t, err, "Should return error for non-existent package")
				assert.Contains(t, err.Error(), "HTTP error", "Error should indicate HTTP status problem")

				expectedFile := filepath.Join(tarballPath, "invalid-package-12345678-1.0.0.tgz")
				info, statErr := os.Stat(expectedFile)
				if statErr == nil {
					assert.Equal(t, int64(0), info.Size(), "File should be empty or not exist")
				}
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tarballPath, url := tc.setupFunc(t)
			tarball := newDownloadTarball(url, tarballPath)
			err := tarball.download()

			if tc.expectError {
				assert.Error(t, err, "Expected an error")
			} else {
				assert.NoError(t, err, "Expected no error")
			}

			tc.validate(t, tarballPath, url, err)
		})
	}
}
