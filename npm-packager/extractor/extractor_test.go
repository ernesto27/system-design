package extractor

import (
	"archive/tar"
	"compress/gzip"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

// setupTestExtractorDirs creates temporary directories for testing
func setupTestExtractorDirs(t *testing.T) (string, string) {
	tmpDir := t.TempDir()
	srcDir := filepath.Join(tmpDir, "src")
	destDir := filepath.Join(tmpDir, "dest")
	os.MkdirAll(srcDir, 0755)
	os.MkdirAll(destDir, 0755)
	return srcDir, destDir
}

// createTestTarball creates a test .tgz file with specified entries
func createTestTarball(t *testing.T, path string, entries map[string]string) {
	file, err := os.Create(path)
	assert.NoError(t, err)
	defer file.Close()

	gzw := gzip.NewWriter(file)
	defer gzw.Close()

	tw := tar.NewWriter(gzw)
	defer tw.Close()

	for name, content := range entries {
		header := &tar.Header{
			Name:     name,
			Mode:     0644,
			Size:     int64(len(content)),
			Typeflag: tar.TypeReg,
		}
		err := tw.WriteHeader(header)
		assert.NoError(t, err)

		_, err = tw.Write([]byte(content))
		assert.NoError(t, err)
	}
}

func TestTGZExtractorStripPackagePrefix(t *testing.T) {
	testCases := []struct {
		name        string
		inputPath   string
		expectedVal string
	}{
		{
			name:        "Strip package prefix successfully",
			inputPath:   "package/index.js",
			expectedVal: "index.js",
		},
		{
			name:        "Strip package prefix from nested path",
			inputPath:   "package/lib/utils.js",
			expectedVal: "lib/utils.js",
		},
		{
			name:        "No package prefix - return empty string",
			inputPath:   "index.js",
			expectedVal: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			extractor := NewTGZExtractor()
			result := extractor.stripPackagePrefix(tc.inputPath)
			assert.Equal(t, tc.expectedVal, result)
		})
	}
}

func TestTGZExtractorExtract(t *testing.T) {
	testCases := []struct {
		name        string
		setupFunc   func(t *testing.T) (string, string)
		expectError bool
		validate    func(t *testing.T, destDir string, err error)
	}{
		{
			name: "Extract tarball with package prefix successfully",
			setupFunc: func(t *testing.T) (string, string) {
				srcDir, destDir := setupTestExtractorDirs(t)
				tarballPath := filepath.Join(srcDir, "test.tgz")

				entries := map[string]string{
					"package/index.js":     "console.log('hello');",
					"package/package.json": "{\"name\":\"test\"}",
					"package/lib/utils.js": "module.exports = {};",
				}
				createTestTarball(t, tarballPath, entries)

				return tarballPath, destDir
			},
			expectError: false,
			validate: func(t *testing.T, destDir string, err error) {
				assert.NoError(t, err, "Extract should succeed")

				indexPath := filepath.Join(destDir, "index.js")
				assert.FileExists(t, indexPath)

				packageJsonPath := filepath.Join(destDir, "package.json")
				assert.FileExists(t, packageJsonPath)

				utilsPath := filepath.Join(destDir, "lib", "utils.js")
				assert.FileExists(t, utilsPath)

				content, readErr := os.ReadFile(indexPath)
				assert.NoError(t, readErr)
				assert.Equal(t, "console.log('hello');", string(content))
			},
		},
		{
			name: "Skip files without directory prefix",
			setupFunc: func(t *testing.T) (string, string) {
				srcDir, destDir := setupTestExtractorDirs(t)
				tarballPath := filepath.Join(srcDir, "test.tgz")

				entries := map[string]string{
					"index.js":  "console.log('no prefix');",
					"README.md": "# Test Package",
				}
				createTestTarball(t, tarballPath, entries)

				return tarballPath, destDir
			},
			expectError: false,
			validate: func(t *testing.T, destDir string, err error) {
				assert.NoError(t, err)

				indexPath := filepath.Join(destDir, "index.js")
				assert.NoFileExists(t, indexPath, "Files without directory prefix should be skipped")

				readmePath := filepath.Join(destDir, "README.md")
				assert.NoFileExists(t, readmePath, "Files without directory prefix should be skipped")
			},
		},
		{
			name: "Error with non-existent tarball file",
			setupFunc: func(t *testing.T) (string, string) {
				srcDir, destDir := setupTestExtractorDirs(t)
				tarballPath := filepath.Join(srcDir, "nonexistent.tgz")
				return tarballPath, destDir
			},
			expectError: true,
			validate: func(t *testing.T, destDir string, err error) {
				assert.Error(t, err, "Should return error for non-existent file")
				assert.Contains(t, err.Error(), "failed to open file")
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tarballPath, destDir := tc.setupFunc(t)
			extractor := NewTGZExtractor()
			err := extractor.Extract(tarballPath, destDir)

			if tc.expectError {
				assert.Error(t, err, "Expected an error")
			} else {
				assert.NoError(t, err, "Expected no error")
			}

			tc.validate(t, destDir, err)
		})
	}
}
