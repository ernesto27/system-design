package utils

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDownloadFile(t *testing.T) {
	testCases := []struct {
		name          string
		setupFunc     func(t *testing.T) (server *httptest.Server, filename string, etag string)
		expectError   bool
		validateError func(t *testing.T, err error)
		validate      func(t *testing.T, filename string, returnedEtag string, statusCode int)
	}{
		{
			name: "Successful download without etag",
			setupFunc: func(t *testing.T) (*httptest.Server, string, string) {
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					assert.Equal(t, "", r.Header.Get("If-None-Match"))
					w.Header().Set("ETag", `"test-etag-123"`)
					w.WriteHeader(http.StatusOK)
					w.Write([]byte("test content"))
				}))

				tmpDir := t.TempDir()
				filename := filepath.Join(tmpDir, "test.json")
				return server, filename, ""
			},
			expectError: false,
			validate: func(t *testing.T, filename string, returnedEtag string, statusCode int) {
				assert.Equal(t, http.StatusOK, statusCode)
				assert.Equal(t, `"test-etag-123"`, returnedEtag)

				content, err := os.ReadFile(filename)
				assert.NoError(t, err)
				assert.Equal(t, "test content", string(content))
			},
		},
		{
			name: "Download with etag - not modified (304)",
			setupFunc: func(t *testing.T) (*httptest.Server, string, string) {
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					assert.Equal(t, `"existing-etag"`, r.Header.Get("If-None-Match"))
					w.WriteHeader(http.StatusNotModified)
				}))

				tmpDir := t.TempDir()
				filename := filepath.Join(tmpDir, "test.json")
				return server, filename, `"existing-etag"`
			},
			expectError: false,
			validate: func(t *testing.T, filename string, returnedEtag string, statusCode int) {
				assert.Equal(t, http.StatusNotModified, statusCode)
				assert.Equal(t, `"existing-etag"`, returnedEtag)

				// File should not exist since download was skipped
				_, err := os.Stat(filename)
				assert.True(t, os.IsNotExist(err))
			},
		},
		{
			name: "Download with etag - modified (200)",
			setupFunc: func(t *testing.T) (*httptest.Server, string, string) {
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					assert.Equal(t, `"old-etag"`, r.Header.Get("If-None-Match"))
					w.Header().Set("ETag", `"new-etag"`)
					w.WriteHeader(http.StatusOK)
					w.Write([]byte("updated content"))
				}))

				tmpDir := t.TempDir()
				filename := filepath.Join(tmpDir, "test.json")
				return server, filename, `"old-etag"`
			},
			expectError: false,
			validate: func(t *testing.T, filename string, returnedEtag string, statusCode int) {
				assert.Equal(t, http.StatusOK, statusCode)
				assert.Equal(t, `"new-etag"`, returnedEtag)

				content, err := os.ReadFile(filename)
				assert.NoError(t, err)
				assert.Equal(t, "updated content", string(content))
			},
		},
		{
			name: "HTTP 404 error",
			setupFunc: func(t *testing.T) (*httptest.Server, string, string) {
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusNotFound)
				}))

				tmpDir := t.TempDir()
				filename := filepath.Join(tmpDir, "test.json")
				return server, filename, ""
			},
			expectError: true,
			validateError: func(t *testing.T, err error) {
				assert.Contains(t, err.Error(), "HTTP error")
				assert.Contains(t, err.Error(), "404")
			},
			validate: func(t *testing.T, filename string, returnedEtag string, statusCode int) {
				assert.Equal(t, http.StatusNotFound, statusCode)
			},
		},
		{
			name: "HTTP 500 server error",
			setupFunc: func(t *testing.T) (*httptest.Server, string, string) {
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusInternalServerError)
				}))

				tmpDir := t.TempDir()
				filename := filepath.Join(tmpDir, "test.json")
				return server, filename, ""
			},
			expectError: true,
			validateError: func(t *testing.T, err error) {
				assert.Contains(t, err.Error(), "HTTP error")
				assert.Contains(t, err.Error(), "500")
			},
			validate: func(t *testing.T, filename string, returnedEtag string, statusCode int) {
				assert.Equal(t, http.StatusInternalServerError, statusCode)
			},
		},
		{
			name: "Download creates nested directories",
			setupFunc: func(t *testing.T) (*httptest.Server, string, string) {
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusOK)
					w.Write([]byte("content"))
				}))

				tmpDir := t.TempDir()
				filename := filepath.Join(tmpDir, "nested", "deep", "file.json")
				return server, filename, ""
			},
			expectError: false,
			validate: func(t *testing.T, filename string, returnedEtag string, statusCode int) {
				assert.Equal(t, http.StatusOK, statusCode)

				content, err := os.ReadFile(filename)
				assert.NoError(t, err)
				assert.Equal(t, "content", string(content))

				// Verify directories were created
				dir := filepath.Dir(filename)
				info, err := os.Stat(dir)
				assert.NoError(t, err)
				assert.True(t, info.IsDir())
			},
		},
		{
			name: "Invalid URL request creation",
			setupFunc: func(t *testing.T) (*httptest.Server, string, string) {
				// Return nil server to test URL handling
				tmpDir := t.TempDir()
				filename := filepath.Join(tmpDir, "test.json")
				return nil, filename, ""
			},
			expectError: false, // Will test separately with invalid URL
			validate:    func(t *testing.T, filename string, returnedEtag string, statusCode int) {},
		},
		{
			name: "Empty file content",
			setupFunc: func(t *testing.T) (*httptest.Server, string, string) {
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusOK)
					// Don't write anything
				}))

				tmpDir := t.TempDir()
				filename := filepath.Join(tmpDir, "empty.json")
				return server, filename, ""
			},
			expectError: false,
			validate: func(t *testing.T, filename string, returnedEtag string, statusCode int) {
				assert.Equal(t, http.StatusOK, statusCode)

				content, err := os.ReadFile(filename)
				assert.NoError(t, err)
				assert.Equal(t, "", string(content))
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			server, filename, etag := tc.setupFunc(t)
			if server != nil {
				defer server.Close()
			}

			var url string
			if server != nil {
				url = server.URL
			} else if tc.name == "Invalid URL request creation" {
				// Skip this test case as it would be handled separately
				t.Skip("Testing invalid URL separately")
				return
			}

			returnedEtag, statusCode, err := DownloadFile(url, filename, etag)

			if tc.expectError {
				assert.Error(t, err)
				if tc.validateError != nil {
					tc.validateError(t, err)
				}
			} else {
				assert.NoError(t, err)
			}

			tc.validate(t, filename, returnedEtag, statusCode)
		})
	}
}

func TestCreateDir(t *testing.T) {
	testCases := []struct {
		name        string
		setupFunc   func(t *testing.T) string
		expectError bool
		validate    func(t *testing.T, dirPath string)
	}{
		{
			name: "Create new directory",
			setupFunc: func(t *testing.T) string {
				tmpDir := t.TempDir()
				return filepath.Join(tmpDir, "newdir")
			},
			expectError: false,
			validate: func(t *testing.T, dirPath string) {
				info, err := os.Stat(dirPath)
				assert.NoError(t, err)
				assert.True(t, info.IsDir())
			},
		},
		{
			name: "Directory already exists",
			setupFunc: func(t *testing.T) string {
				tmpDir := t.TempDir()
				existingDir := filepath.Join(tmpDir, "existing")
				err := os.Mkdir(existingDir, 0755)
				assert.NoError(t, err)
				return existingDir
			},
			expectError: false,
			validate: func(t *testing.T, dirPath string) {
				info, err := os.Stat(dirPath)
				assert.NoError(t, err)
				assert.True(t, info.IsDir())
			},
		},
		{
			name: "Create directory in nested path",
			setupFunc: func(t *testing.T) string {
				tmpDir := t.TempDir()
				// Parent directory exists
				parentDir := filepath.Join(tmpDir, "parent")
				err := os.Mkdir(parentDir, 0755)
				assert.NoError(t, err)
				return filepath.Join(parentDir, "child")
			},
			expectError: false,
			validate: func(t *testing.T, dirPath string) {
				info, err := os.Stat(dirPath)
				assert.NoError(t, err)
				assert.True(t, info.IsDir())
			},
		},
		{
			name: "Cannot create directory (parent doesn't exist)",
			setupFunc: func(t *testing.T) string {
				tmpDir := t.TempDir()
				// Multiple levels deep without creating parents
				return filepath.Join(tmpDir, "nonexistent", "parent", "child")
			},
			expectError: true,
			validate: func(t *testing.T, dirPath string) {
				_, err := os.Stat(dirPath)
				assert.True(t, os.IsNotExist(err))
			},
		},
		{
			name: "Path is a file, not directory",
			setupFunc: func(t *testing.T) string {
				tmpDir := t.TempDir()
				filePath := filepath.Join(tmpDir, "file.txt")
				err := os.WriteFile(filePath, []byte("content"), 0644)
				assert.NoError(t, err)
				return filePath
			},
			expectError: false, // CreateDir checks IsNotExist, so it will skip if file exists
			validate: func(t *testing.T, dirPath string) {
				info, err := os.Stat(dirPath)
				assert.NoError(t, err)
				assert.False(t, info.IsDir()) // Should still be a file
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			dirPath := tc.setupFunc(t)
			err := CreateDir(dirPath)

			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			tc.validate(t, dirPath)
		})
	}
}

func TestCreateDepKey(t *testing.T) {
	testCases := []struct {
		name       string
		pkgName    string
		version    string
		parentName string
		expected   string
	}{
		{
			name:       "Simple dependency key",
			pkgName:    "express",
			version:    "4.18.0",
			parentName: "myapp",
			expected:   "express@4.18.0@myapp",
		},
		{
			name:       "Scoped package",
			pkgName:    "@types/node",
			version:    "18.0.0",
			parentName: "typescript-app",
			expected:   "@types/node@18.0.0@typescript-app",
		},
		{
			name:       "Empty parent name",
			pkgName:    "lodash",
			version:    "4.17.21",
			parentName: "",
			expected:   "lodash@4.17.21@",
		},
		{
			name:       "Empty version",
			pkgName:    "react",
			version:    "",
			parentName: "my-react-app",
			expected:   "react@@my-react-app",
		},
		{
			name:       "All empty",
			pkgName:    "",
			version:    "",
			parentName: "",
			expected:   "@@",
		},
		{
			name:       "Special characters in names",
			pkgName:    "pkg-with-dash",
			version:    "1.0.0-beta.1",
			parentName: "parent_with_underscore",
			expected:   "pkg-with-dash@1.0.0-beta.1@parent_with_underscore",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := CreateDepKey(tc.pkgName, tc.version, tc.parentName)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestFolderExists(t *testing.T) {
	testCases := []struct {
		name      string
		setupFunc func(t *testing.T) string
		expected  bool
	}{
		{
			name: "Folder exists",
			setupFunc: func(t *testing.T) string {
				tmpDir := t.TempDir()
				dirPath := filepath.Join(tmpDir, "existing")
				err := os.Mkdir(dirPath, 0755)
				assert.NoError(t, err)
				return dirPath
			},
			expected: true,
		},
		{
			name: "Folder does not exist",
			setupFunc: func(t *testing.T) string {
				tmpDir := t.TempDir()
				return filepath.Join(tmpDir, "nonexistent")
			},
			expected: false,
		},
		{
			name: "Path is a file, not folder",
			setupFunc: func(t *testing.T) string {
				tmpDir := t.TempDir()
				filePath := filepath.Join(tmpDir, "file.txt")
				err := os.WriteFile(filePath, []byte("content"), 0644)
				assert.NoError(t, err)
				return filePath
			},
			expected: false,
		},
		{
			name: "Empty path",
			setupFunc: func(t *testing.T) string {
				return ""
			},
			expected: false,
		},
		{
			name: "Nested folder exists",
			setupFunc: func(t *testing.T) string {
				tmpDir := t.TempDir()
				nestedPath := filepath.Join(tmpDir, "level1", "level2", "level3")
				err := os.MkdirAll(nestedPath, 0755)
				assert.NoError(t, err)
				return nestedPath
			},
			expected: true,
		},
		{
			name: "Parent exists but child doesn't",
			setupFunc: func(t *testing.T) string {
				tmpDir := t.TempDir()
				parentPath := filepath.Join(tmpDir, "parent")
				err := os.Mkdir(parentPath, 0755)
				assert.NoError(t, err)
				return filepath.Join(parentPath, "nonexistent_child")
			},
			expected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			dirPath := tc.setupFunc(t)
			result := FolderExists(dirPath)
			assert.Equal(t, tc.expected, result, fmt.Sprintf("FolderExists(%q) should return %v", dirPath, tc.expected))
		})
	}
}
