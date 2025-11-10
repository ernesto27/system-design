package packagecopy

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPackageCopyCopyDirectory(t *testing.T) {
	testCases := []struct {
		name        string
		setup       func(*testing.T) (string, string)
		expectErr   bool
		errContains string
		assertDest  func(*testing.T, string)
	}{
		{
			name: "copies nested directories and files",
			setup: func(t *testing.T) (string, string) {
				t.Helper()
				baseDir := t.TempDir()
				src := filepath.Join(baseDir, "src")
				dst := filepath.Join(baseDir, "dst")

				if err := os.MkdirAll(filepath.Join(src, "nested"), 0o755); err != nil {
					t.Fatalf("mkdir: %v", err)
				}

				if err := os.WriteFile(filepath.Join(src, "root.txt"), []byte("root-content"), 0o644); err != nil {
					t.Fatalf("write root: %v", err)
				}

				if err := os.WriteFile(filepath.Join(src, "nested", "child.txt"), []byte("child-content"), 0o600); err != nil {
					t.Fatalf("write child: %v", err)
				}

				return src, dst
			},
			assertDest: func(t *testing.T, dst string) {
				rootBytes, err := os.ReadFile(filepath.Join(dst, "root.txt"))
				if err != nil {
					t.Fatalf("read root: %v", err)
				}
				assert.Equal(t, []byte("root-content"), rootBytes)

				childBytes, err := os.ReadFile(filepath.Join(dst, "nested", "child.txt"))
				if err != nil {
					t.Fatalf("read child: %v", err)
				}
				assert.Equal(t, []byte("child-content"), childBytes)
			},
		},
		{
			name: "errors when source missing",
			setup: func(t *testing.T) (string, string) {
				t.Helper()
				baseDir := t.TempDir()
				return filepath.Join(baseDir, "missing"), filepath.Join(baseDir, "dst")
			},
			expectErr:   true,
			errContains: "source does not exist",
		},
		{
			name: "errors when source is file",
			setup: func(t *testing.T) (string, string) {
				t.Helper()
				baseDir := t.TempDir()
				src := filepath.Join(baseDir, "file.txt")
				if err := os.WriteFile(src, []byte("content"), 0o644); err != nil {
					t.Fatalf("write file: %v", err)
				}
				return src, filepath.Join(baseDir, "dst")
			},
			expectErr:   true,
			errContains: "source is not a directory",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			src, dst := tc.setup(t)

			pc := NewPackageCopy()
			err := pc.CopyDirectory(src, dst)

			if tc.expectErr {
				assert.Error(t, err)
				if tc.errContains != "" {
					assert.Contains(t, err.Error(), tc.errContains)
				}
				return
			}

			assert.NoError(t, err)
			if tc.assertDest != nil {
				tc.assertDest(t, dst)
			}
		})
	}
}

