package packagejson

import (
	"encoding/json"
	"npm-packager/config"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPackageJSONParser_Parse(t *testing.T) {
	testCases := []struct {
		name        string
		setupFile   func(t *testing.T) string
		expectError bool
		validate    func(t *testing.T, result *PackageJSON)
	}{
		{
			name: "Valid basic package.json",
			setupFile: func(t *testing.T) string {
				tmpDir := t.TempDir()
				tmpFile := filepath.Join(tmpDir, "package.json")

				packageData := PackageJSON{
					Name:        "test-package",
					Description: "A test package",
					Version:     "1.2.3",
					Author:      "Test Author",
					License:     "MIT",
					Homepage:    "https://example.com",
					Keywords:    []string{"test", "example"},
					Dependencies: map[string]string{
						"express": "^4.18.0",
						"lodash":  "^4.17.21",
					},
					Scripts: map[string]string{
						"start": "node index.js",
						"test":  "jest",
					},
					Main:    "index.js",
					Types:   "index.d.ts",
					Private: false,
				}

				data, _ := json.MarshalIndent(packageData, "", "  ")
				os.WriteFile(tmpFile, data, 0644)
				return tmpDir
			},
			expectError: false,
			validate: func(t *testing.T, result *PackageJSON) {
				assert.Equal(t, "test-package", result.Name)
				assert.Equal(t, "1.2.3", result.Version)
				assert.Equal(t, "A test package", result.Description)
				assert.Equal(t, "MIT", result.License)
				assert.Equal(t, map[string]string{
					"express": "^4.18.0",
					"lodash":  "^4.17.21",
				}, result.Dependencies)
				assert.Equal(t, map[string]string{
					"start": "node index.js",
					"test":  "jest",
				}, result.Scripts)
			},
		},
		{
			name: "Non-existent file",
			setupFile: func(t *testing.T) string {
				return t.TempDir()
			},
			expectError: true,
			validate: func(t *testing.T, result *PackageJSON) {
				assert.Nil(t, result)
			},
		},
		{
			name: "Invalid JSON",
			setupFile: func(t *testing.T) string {
				tmpDir := t.TempDir()
				tmpFile := filepath.Join(tmpDir, "package.json")

				invalidJSON := []byte(`{
					"name": "test",
					"version": "1.0.0",
					"invalid":
				}`)

				os.WriteFile(tmpFile, invalidJSON, 0644)
				return tmpDir
			},
			expectError: true,
			validate: func(t *testing.T, result *PackageJSON) {
				assert.Nil(t, result)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tmpDir := tc.setupFile(t)

			// Save current directory
			originalDir, err := os.Getwd()
			assert.NoError(t, err)
			defer os.Chdir(originalDir)

			// Change to temp directory
			err = os.Chdir(tmpDir)
			assert.NoError(t, err)

			cfg, err := config.New()
			assert.NoError(t, err)

			parser := NewPackageJSONParser(cfg)
			result, err := parser.Parse("package.json")

			if tc.expectError {
				assert.Error(t, err, "Expected an error")
			} else {
				assert.NoError(t, err, "Expected no error")
			}

			tc.validate(t, result)
		})
	}
}
