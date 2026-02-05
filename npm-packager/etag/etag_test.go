package etag

import (
	"encoding/json"
	"npm-packager/packagejson"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

// setupTestEtagDir creates temporary config directory for testing
// Note: NewEtag will create an "etag" subdirectory inside this path
func setupTestEtagDir(t *testing.T) string {
	tmpDir := t.TempDir()
	return tmpDir
}

func TestEtagSetPackages(t *testing.T) {
	testCases := []struct {
		name      string
		setupFunc func(t *testing.T) (*Etag, map[string]packagejson.Dependency)
		validate  func(t *testing.T, etag *Etag)
	}{
		{
			name: "Set packages successfully",
			setupFunc: func(t *testing.T) (*Etag, map[string]packagejson.Dependency) {
				configDir := setupTestEtagDir(t)
				etag, err := NewEtag(configDir)
				assert.NoError(t, err)
				packages := map[string]packagejson.Dependency{
					"express": {Name: "express", Version: "4.18.0", Etag: "W/\"abc123\""},
					"lodash":  {Name: "lodash", Version: "4.17.21", Etag: "W/\"def456\""},
				}
				return etag, packages
			},
			validate: func(t *testing.T, etag *Etag) {
				assert.NotNil(t, etag.packages)
				assert.Equal(t, 2, len(etag.packages))
				assert.Equal(t, "4.18.0", etag.packages["express"].Version)
				assert.Equal(t, "W/\"abc123\"", etag.packages["express"].Etag)
			},
		},
		{
			name: "Overwrite existing packages",
			setupFunc: func(t *testing.T) (*Etag, map[string]packagejson.Dependency) {
				configDir := setupTestEtagDir(t)
				etag, err := NewEtag(configDir)
				assert.NoError(t, err)
				etag.packages = map[string]packagejson.Dependency{
					"old-package": {Name: "old-package", Version: "1.0.0", Etag: "W/\"old\""},
				}
				packages := map[string]packagejson.Dependency{
					"new-package": {Name: "new-package", Version: "2.0.0", Etag: "W/\"new\""},
				}
				return etag, packages
			},
			validate: func(t *testing.T, etag *Etag) {
				assert.Equal(t, 1, len(etag.packages))
				assert.Contains(t, etag.packages, "new-package")
				assert.NotContains(t, etag.packages, "old-package")
			},
		},
		{
			name: "Handle empty packages map",
			setupFunc: func(t *testing.T) (*Etag, map[string]packagejson.Dependency) {
				configDir := setupTestEtagDir(t)
				etag, err := NewEtag(configDir)
				assert.NoError(t, err)
				packages := map[string]packagejson.Dependency{}
				return etag, packages
			},
			validate: func(t *testing.T, etag *Etag) {
				assert.NotNil(t, etag.packages)
				assert.Equal(t, 0, len(etag.packages))
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			etag, packages := tc.setupFunc(t)
			etag.setPackages(packages)
			tc.validate(t, etag)
		})
	}
}

func TestEtagGet(t *testing.T) {
	testCases := []struct {
		name        string
		setupFunc   func(t *testing.T) (*Etag, string)
		expectedVal string
	}{
		{
			name: "Retrieve existing etag successfully",
			setupFunc: func(t *testing.T) (*Etag, string) {
				configDir := setupTestEtagDir(t)
				etag, err := NewEtag(configDir)
				assert.NoError(t, err)
				etag.etagData = map[string]EtagEntry{
					"express": {Etag: "W/\"abc123\""},
					"lodash":  {Etag: "W/\"def456\""},
				}
				return etag, "express"
			},
			expectedVal: "W/\"abc123\"",
		},
		{
			name: "Return empty string for non-existent package",
			setupFunc: func(t *testing.T) (*Etag, string) {
				configDir := setupTestEtagDir(t)
				etag, err := NewEtag(configDir)
				assert.NoError(t, err)
				etag.etagData = map[string]EtagEntry{
					"express": {Etag: "W/\"abc123\""},
				}
				return etag, "non-existent"
			},
			expectedVal: "",
		},
		{
			name: "Return empty string from empty etagData",
			setupFunc: func(t *testing.T) (*Etag, string) {
				configDir := setupTestEtagDir(t)
				etag, err := NewEtag(configDir)
				assert.NoError(t, err)
				return etag, "any-package"
			},
			expectedVal: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			etag, packageName := tc.setupFunc(t)
			result := etag.Get(packageName)
			assert.Equal(t, tc.expectedVal, result)
		})
	}
}

func TestEtagSave(t *testing.T) {
	testCases := []struct {
		name        string
		setupFunc   func(t *testing.T) (*Etag, string)
		expectError bool
		validate    func(t *testing.T, etag *Etag, err error)
	}{
		{
			name: "Save new etags to file successfully",
			setupFunc: func(t *testing.T) (*Etag, string) {
				configDir := setupTestEtagDir(t)
				etag, err := NewEtag(configDir)
				assert.NoError(t, err)
				etag.packages = map[string]packagejson.Dependency{
					"express": {Name: "express", Version: "4.18.0", Etag: "W/\"abc123\""},
					"lodash":  {Name: "lodash", Version: "4.17.21", Etag: "W/\"def456\""},
				}
				return etag, configDir
			},
			expectError: false,
			validate: func(t *testing.T, etag *Etag, err error) {
				assert.NoError(t, err, "Save should succeed")

				etagFilePath := filepath.Join(etag.etagPath, "etag.json")
				assert.FileExists(t, etagFilePath)

				data, readErr := os.ReadFile(etagFilePath)
				assert.NoError(t, readErr)

				var savedData map[string]EtagEntry
				jsonErr := json.Unmarshal(data, &savedData)
				assert.NoError(t, jsonErr)

				assert.Equal(t, 2, len(savedData))
				assert.Equal(t, "W/\"abc123\"", savedData["express"].Etag)
				assert.Equal(t, "W/\"def456\"", savedData["lodash"].Etag)
			},
		},
		{
			name: "Merge with existing etag data",
			setupFunc: func(t *testing.T) (*Etag, string) {
				configDir := setupTestEtagDir(t)
				etagDir := filepath.Join(configDir, "etag")
				os.MkdirAll(etagDir, 0755)
				etagFilePath := filepath.Join(etagDir, "etag.json")

				existingData := map[string]EtagEntry{
					"old-package": {Etag: "W/\"old123\""},
				}
				jsonData, _ := json.Marshal(existingData)
				os.WriteFile(etagFilePath, jsonData, 0644)

				etag, err := NewEtag(configDir)
				assert.NoError(t, err)
				etag.packages = map[string]packagejson.Dependency{
					"express": {Name: "express", Version: "4.18.0", Etag: "W/\"abc123\""},
				}
				return etag, configDir
			},
			expectError: false,
			validate: func(t *testing.T, etag *Etag, err error) {
				assert.NoError(t, err)

				etagFilePath := filepath.Join(etag.etagPath, "etag.json")
				data, readErr := os.ReadFile(etagFilePath)
				assert.NoError(t, readErr)

				var savedData map[string]EtagEntry
				jsonErr := json.Unmarshal(data, &savedData)
				assert.NoError(t, jsonErr)

				assert.Equal(t, 2, len(savedData))
				assert.Equal(t, "W/\"old123\"", savedData["old-package"].Etag)
				assert.Equal(t, "W/\"abc123\"", savedData["express"].Etag)
			},
		},
		{
			name: "Skip packages with empty etags",
			setupFunc: func(t *testing.T) (*Etag, string) {
				configDir := setupTestEtagDir(t)
				etag, err := NewEtag(configDir)
				assert.NoError(t, err)
				etag.packages = map[string]packagejson.Dependency{
					"express":     {Name: "express", Version: "4.18.0", Etag: "W/\"abc123\""},
					"no-etag-pkg": {Name: "no-etag-pkg", Version: "1.0.0", Etag: ""},
					"lodash":      {Name: "lodash", Version: "4.17.21", Etag: "W/\"def456\""},
				}
				return etag, configDir
			},
			expectError: false,
			validate: func(t *testing.T, etag *Etag, err error) {
				assert.NoError(t, err)

				etagFilePath := filepath.Join(etag.etagPath, "etag.json")
				data, readErr := os.ReadFile(etagFilePath)
				assert.NoError(t, readErr)

				var savedData map[string]EtagEntry
				jsonErr := json.Unmarshal(data, &savedData)
				assert.NoError(t, jsonErr)

				assert.Equal(t, 2, len(savedData))
				assert.Contains(t, savedData, "express")
				assert.Contains(t, savedData, "lodash")
				assert.NotContains(t, savedData, "no-etag-pkg")
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			etag, _ := tc.setupFunc(t)
			err := etag.Save()

			if tc.expectError {
				assert.Error(t, err, "Expected an error")
			} else {
				assert.NoError(t, err, "Expected no error")
			}

			tc.validate(t, etag, err)
		})
	}
}
