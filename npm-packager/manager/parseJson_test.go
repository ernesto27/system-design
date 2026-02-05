package manager

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseJsonManifestParse(t *testing.T) {
	testCases := []struct {
		name        string
		setup       func(*testing.T) string
		expectErr   bool
		errContains string
		assertPkg   func(*testing.T, *NPMPackage)
	}{
		{
			name: "valid manifest",
			setup: func(t *testing.T) string {
				t.Helper()
				dir := t.TempDir()
				filePath := filepath.Join(dir, "package.json")
				content := `{"_id":"example","name":"example","dist-tags":{"latest":"1.0.0"},"versions":{"1.0.0":{"name":"example","version":"1.0.0","_id":"example@1.0.0","dist":{"tarball":"https://example.com/example-1.0.0.tgz","shasum":"abc123"}}}}`

				if err := os.WriteFile(filePath, []byte(content), 0o600); err != nil {
					t.Fatalf("write manifest: %v", err)
				}

				return filePath
			},
			assertPkg: func(t *testing.T, pkg *NPMPackage) {
				assert.Equal(t, "example", pkg.Name)
				assert.Equal(t, "1.0.0", pkg.DistTags.Latest)

				if assert.Contains(t, pkg.Versions, "1.0.0") {
					assert.Equal(t, "https://example.com/example-1.0.0.tgz", pkg.Versions["1.0.0"].Dist.Tarball)
				}
			},
		},
		{
			name: "file missing",
			setup: func(t *testing.T) string {
				t.Helper()
				return filepath.Join(t.TempDir(), "missing.json")
			},
			expectErr:   true,
			errContains: "failed to open file",
		},
		{
			name: "invalid json",
			setup: func(t *testing.T) string {
				t.Helper()
				dir := t.TempDir()
				filePath := filepath.Join(dir, "package.json")

				if err := os.WriteFile(filePath, []byte("not-json"), 0o600); err != nil {
					t.Fatalf("write manifest: %v", err)
				}

				return filePath
			},
			expectErr:   true,
			errContains: "failed to parse JSON",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			parser := newParseJsonManifest()
			filePath := tc.setup(t)

			pkg, err := parser.parse(filePath)

			if tc.expectErr {
				assert.Error(t, err)
				if tc.errContains != "" {
					assert.Contains(t, err.Error(), tc.errContains)
				}
				assert.Nil(t, pkg)
				return
			}

			assert.NoError(t, err)
			if assert.NotNil(t, pkg) && tc.assertPkg != nil {
				tc.assertPkg(t, pkg)
			}
		})
	}
}

