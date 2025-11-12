package binlink

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Helper functions for creating test fixtures

func createTestPackageJSON(t *testing.T, dir, name string, binField interface{}) {
	pkgPath := filepath.Join(dir, "package.json")

	binJSON, err := json.Marshal(binField)
	assert.NoError(t, err)

	pkg := map[string]json.RawMessage{
		"name": json.RawMessage(`"` + name + `"`),
		"bin":  binJSON,
	}

	data, err := json.Marshal(pkg)
	assert.NoError(t, err)

	err = os.WriteFile(pkgPath, data, 0644)
	assert.NoError(t, err)
}

func createTestPackage(t *testing.T, nodeModulesPath, pkgName string, binField interface{}) string {
	pkgPath := filepath.Join(nodeModulesPath, pkgName)
	err := os.MkdirAll(pkgPath, 0755)
	assert.NoError(t, err)

	createTestPackageJSON(t, pkgPath, pkgName, binField)
	return pkgPath
}

func createScopedPackage(t *testing.T, nodeModulesPath, scope, pkgName string, binField interface{}) string {
	scopePath := filepath.Join(nodeModulesPath, scope)
	err := os.MkdirAll(scopePath, 0755)
	assert.NoError(t, err)

	pkgPath := filepath.Join(scopePath, pkgName)
	err = os.MkdirAll(pkgPath, 0755)
	assert.NoError(t, err)

	fullName := scope + "/" + pkgName
	createTestPackageJSON(t, pkgPath, fullName, binField)
	return pkgPath
}

func verifySymlink(t *testing.T, linkPath, expectedTarget string) {
	target, err := os.Readlink(linkPath)
	assert.NoError(t, err, "Failed to read symlink at %s", linkPath)
	assert.Equal(t, expectedTarget, target, "Symlink target mismatch")
}

// Tests for NewBinLinker

func TestNewBinLinker(t *testing.T) {
	testCases := []struct {
		name           string
		nodeModules    string
		expectedBinDir string
	}{
		{
			name:           "Local installation",
			nodeModules:    "/home/user/project/node_modules",
			expectedBinDir: "/home/user/project/node_modules/.bin",
		},
		{
			name:           "Empty node modules path",
			nodeModules:    "",
			expectedBinDir: ".bin",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			bl := NewBinLinker(tc.nodeModules)

			assert.NotNil(t, bl)
			assert.Equal(t, tc.nodeModules, bl.nodeModulesPath)
			assert.False(t, bl.isGlobal)
			assert.Equal(t, tc.expectedBinDir, bl.binPath)
		})
	}
}

func TestSetGlobalMode(t *testing.T) {
	testCases := []struct {
		name               string
		initialNodeModules string
		globalNodeModules  string
		globalBinPath      string
	}{
		{
			name:               "Switch to global mode",
			initialNodeModules: "/home/user/project/node_modules",
			globalNodeModules:  "/usr/local/lib/node_modules",
			globalBinPath:      "/usr/local/bin",
		},
		{
			name:               "Switch to global mode with same path",
			initialNodeModules: "/usr/local/lib/node_modules",
			globalNodeModules:  "/usr/local/lib/node_modules",
			globalBinPath:      "/usr/local/bin",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			bl := NewBinLinker(tc.initialNodeModules)

			// Verify it starts in local mode
			assert.False(t, bl.isGlobal)
			assert.Equal(t, tc.initialNodeModules, bl.nodeModulesPath)

			// Switch to global mode
			bl.SetGlobalMode(tc.globalNodeModules, tc.globalBinPath)

			// Verify it's now in global mode
			assert.True(t, bl.isGlobal)
			assert.Equal(t, tc.globalNodeModules, bl.nodeModulesPath)
			assert.Equal(t, tc.globalBinPath, bl.binPath)
		})
	}
}

// Tests for CreateBinDirectory

func TestCreateBinDirectory(t *testing.T) {
	testCases := []struct {
		name        string
		setupFunc   func(t *testing.T) *BinLinker
		expectError bool
		validate    func(t *testing.T, bl *BinLinker)
	}{
		{
			name: "Create .bin directory when it doesn't exist",
			setupFunc: func(t *testing.T) *BinLinker {
				tmpDir := t.TempDir()
				return NewBinLinker(tmpDir)
			},
			expectError: false,
			validate: func(t *testing.T, bl *BinLinker) {
				info, err := os.Stat(bl.binPath)
				assert.NoError(t, err)
				assert.True(t, info.IsDir())
				assert.Equal(t, os.FileMode(0755), info.Mode().Perm())
			},
		},
		{
			name: "No error when .bin directory already exists",
			setupFunc: func(t *testing.T) *BinLinker {
				tmpDir := t.TempDir()
				binDir := filepath.Join(tmpDir, ".bin")
				err := os.Mkdir(binDir, 0755)
				assert.NoError(t, err)
				return NewBinLinker(tmpDir)
			},
			expectError: false,
			validate: func(t *testing.T, bl *BinLinker) {
				info, err := os.Stat(bl.binPath)
				assert.NoError(t, err)
				assert.True(t, info.IsDir())
			},
		},
		{
			name: "Create nested directory structure",
			setupFunc: func(t *testing.T) *BinLinker {
				tmpDir := t.TempDir()
				nodeModules := filepath.Join(tmpDir, "nested", "path", "node_modules")
				return NewBinLinker(nodeModules)
			},
			expectError: false,
			validate: func(t *testing.T, bl *BinLinker) {
				info, err := os.Stat(bl.binPath)
				assert.NoError(t, err)
				assert.True(t, info.IsDir())
			},
		},
		{
			name: "Error when path is read-only",
			setupFunc: func(t *testing.T) *BinLinker {
				tmpDir := t.TempDir()
				readOnlyDir := filepath.Join(tmpDir, "readonly")
				err := os.Mkdir(readOnlyDir, 0444)
				assert.NoError(t, err)
				return NewBinLinker(readOnlyDir)
			},
			expectError: true,
			validate:    func(t *testing.T, bl *BinLinker) {},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			bl := tc.setupFunc(t)
			err := bl.CreateBinDirectory()

			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			tc.validate(t, bl)
		})
	}
}

// Tests for parseBinField

func TestParseBinField(t *testing.T) {
	testCases := []struct {
		name        string
		pkgName     string
		binField    json.RawMessage
		expected    map[string]string
		expectError bool
	}{
		{
			name:     "String format - regular package",
			pkgName:  "express",
			binField: json.RawMessage(`"./bin/cli.js"`),
			expected: map[string]string{
				"express": "./bin/cli.js",
			},
			expectError: false,
		},
		{
			name:     "String format - scoped package",
			pkgName:  "@babel/cli",
			binField: json.RawMessage(`"./bin/babel.js"`),
			expected: map[string]string{
				"cli": "./bin/babel.js",
			},
			expectError: false,
		},
		{
			name:     "Object format - single bin",
			pkgName:  "mycli",
			binField: json.RawMessage(`{"mycli": "./bin/cli.js"}`),
			expected: map[string]string{
				"mycli": "./bin/cli.js",
			},
			expectError: false,
		},
		{
			name:    "Object format - multiple bins",
			pkgName: "multi-tool",
			binField: json.RawMessage(`{
				"cmd1": "./bin/cmd1.js",
				"cmd2": "./bin/cmd2.js",
				"cmd3": "./bin/cmd3.js"
			}`),
			expected: map[string]string{
				"cmd1": "./bin/cmd1.js",
				"cmd2": "./bin/cmd2.js",
				"cmd3": "./bin/cmd3.js",
			},
			expectError: false,
		},
		{
			name:        "Empty string",
			pkgName:     "empty",
			binField:    json.RawMessage(`""`),
			expected:    map[string]string{"empty": ""},
			expectError: false,
		},
		{
			name:        "Invalid JSON - number",
			pkgName:     "invalid",
			binField:    json.RawMessage(`123`),
			expected:    nil,
			expectError: true,
		},
		{
			name:        "Invalid JSON - array",
			pkgName:     "invalid",
			binField:    json.RawMessage(`["./bin/cli.js"]`),
			expected:    nil,
			expectError: true,
		},
		{
			name:        "Invalid JSON - boolean",
			pkgName:     "invalid",
			binField:    json.RawMessage(`true`),
			expected:    nil,
			expectError: true,
		},
		{
			name:     "Scoped package with multiple scopes",
			pkgName:  "@types/node",
			binField: json.RawMessage(`"./bin/types.js"`),
			expected: map[string]string{
				"node": "./bin/types.js",
			},
			expectError: false,
		},
		{
			name:        "Empty object",
			pkgName:     "empty-obj",
			binField:    json.RawMessage(`{}`),
			expected:    map[string]string{},
			expectError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			bl := NewBinLinker(tmpDir)

			result, err := bl.parseBinField(tc.pkgName, tc.binField)

			if tc.expectError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expected, result)
			}
		})
	}
}

// Tests for createSymlink

func TestCreateSymlink(t *testing.T) {
	testCases := []struct {
		name        string
		setupFunc   func(t *testing.T) (bl *BinLinker, pkgPath, binName, binRelPath string)
		expectError bool
		validate    func(t *testing.T, bl *BinLinker, binName string)
	}{
		{
			name: "Local installation - regular package",
			setupFunc: func(t *testing.T) (*BinLinker, string, string, string) {
				tmpDir := t.TempDir()
				nodeModules := filepath.Join(tmpDir, "node_modules")
				os.MkdirAll(nodeModules, 0755)

				bl := NewBinLinker(nodeModules)
				bl.CreateBinDirectory()

				pkgPath := filepath.Join(nodeModules, "express")
				os.MkdirAll(filepath.Join(pkgPath, "bin"), 0755)

				return bl, pkgPath, "express", "./bin/cli.js"
			},
			expectError: false,
			validate: func(t *testing.T, bl *BinLinker, binName string) {
				verifySymlink(t, filepath.Join(bl.binPath, binName), "../express/bin/cli.js")
			},
		},
		{
			name: "Local installation - scoped package",
			setupFunc: func(t *testing.T) (*BinLinker, string, string, string) {
				tmpDir := t.TempDir()
				nodeModules := filepath.Join(tmpDir, "node_modules")
				os.MkdirAll(nodeModules, 0755)

				bl := NewBinLinker(nodeModules)
				bl.CreateBinDirectory()

				pkgPath := filepath.Join(nodeModules, "@babel", "cli")
				os.MkdirAll(filepath.Join(pkgPath, "bin"), 0755)

				return bl, pkgPath, "babel", "./bin/babel.js"
			},
			expectError: false,
			validate: func(t *testing.T, bl *BinLinker, binName string) {
				verifySymlink(t, filepath.Join(bl.binPath, binName), "../@babel/cli/bin/babel.js")
			},
		},
		{
			name: "Global installation - absolute path",
			setupFunc: func(t *testing.T) (*BinLinker, string, string, string) {
				tmpDir := t.TempDir()
				nodeModules := filepath.Join(tmpDir, "node_modules")
				binPath := filepath.Join(tmpDir, "bin")
				os.MkdirAll(nodeModules, 0755)
				os.MkdirAll(binPath, 0755)

				bl := NewBinLinker(nodeModules)
				bl.SetGlobalMode(nodeModules, binPath)

				pkgPath := filepath.Join(nodeModules, "nodemon")
				os.MkdirAll(filepath.Join(pkgPath, "bin"), 0755)

				return bl, pkgPath, "nodemon", "./bin/nodemon.js"
			},
			expectError: false,
			validate: func(t *testing.T, bl *BinLinker, binName string) {
				target, err := os.Readlink(filepath.Join(bl.binPath, binName))
				assert.NoError(t, err)
				assert.True(t, filepath.IsAbs(target), "Global installation should use absolute path")
				assert.Contains(t, target, "nodemon/bin/nodemon.js")
			},
		},
		{
			name: "Symlink already exists with correct target - skip",
			setupFunc: func(t *testing.T) (*BinLinker, string, string, string) {
				tmpDir := t.TempDir()
				nodeModules := filepath.Join(tmpDir, "node_modules")
				os.MkdirAll(nodeModules, 0755)

				bl := NewBinLinker(nodeModules)
				bl.CreateBinDirectory()

				pkgPath := filepath.Join(nodeModules, "jest")
				os.MkdirAll(filepath.Join(pkgPath, "bin"), 0755)

				// Create existing symlink
				linkPath := filepath.Join(bl.binPath, "jest")
				targetPath := "../jest/bin/jest.js"
				os.Symlink(targetPath, linkPath)

				return bl, pkgPath, "jest", "./bin/jest.js"
			},
			expectError: false,
			validate: func(t *testing.T, bl *BinLinker, binName string) {
				verifySymlink(t, filepath.Join(bl.binPath, binName), "../jest/bin/jest.js")
			},
		},
		{
			name: "Symlink exists with wrong target - update",
			setupFunc: func(t *testing.T) (*BinLinker, string, string, string) {
				tmpDir := t.TempDir()
				nodeModules := filepath.Join(tmpDir, "node_modules")
				os.MkdirAll(nodeModules, 0755)

				bl := NewBinLinker(nodeModules)
				bl.CreateBinDirectory()

				pkgPath := filepath.Join(nodeModules, "webpack")
				os.MkdirAll(filepath.Join(pkgPath, "bin"), 0755)

				// Create symlink with wrong target
				linkPath := filepath.Join(bl.binPath, "webpack")
				os.Symlink("../old/path.js", linkPath)

				return bl, pkgPath, "webpack", "./bin/webpack.js"
			},
			expectError: false,
			validate: func(t *testing.T, bl *BinLinker, binName string) {
				verifySymlink(t, filepath.Join(bl.binPath, binName), "../webpack/bin/webpack.js")
			},
		},
		{
			name: "Clean relative paths with ./ prefix",
			setupFunc: func(t *testing.T) (*BinLinker, string, string, string) {
				tmpDir := t.TempDir()
				nodeModules := filepath.Join(tmpDir, "node_modules")
				os.MkdirAll(nodeModules, 0755)

				bl := NewBinLinker(nodeModules)
				bl.CreateBinDirectory()

				pkgPath := filepath.Join(nodeModules, "test-pkg")
				os.MkdirAll(pkgPath, 0755)

				return bl, pkgPath, "test", "./cli.js"
			},
			expectError: false,
			validate: func(t *testing.T, bl *BinLinker, binName string) {
				verifySymlink(t, filepath.Join(bl.binPath, binName), "../test-pkg/cli.js")
			},
		},
		{
			name: "Deeply nested bin path",
			setupFunc: func(t *testing.T) (*BinLinker, string, string, string) {
				tmpDir := t.TempDir()
				nodeModules := filepath.Join(tmpDir, "node_modules")
				os.MkdirAll(nodeModules, 0755)

				bl := NewBinLinker(nodeModules)
				bl.CreateBinDirectory()

				pkgPath := filepath.Join(nodeModules, "deep-pkg")
				os.MkdirAll(filepath.Join(pkgPath, "dist", "bin", "commands"), 0755)

				return bl, pkgPath, "deep", "./dist/bin/commands/cli.js"
			},
			expectError: false,
			validate: func(t *testing.T, bl *BinLinker, binName string) {
				verifySymlink(t, filepath.Join(bl.binPath, binName), "../deep-pkg/dist/bin/commands/cli.js")
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			bl, pkgPath, binName, binRelPath := tc.setupFunc(t)
			err := bl.createSymlink(pkgPath, binName, binRelPath)

			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			tc.validate(t, bl, binName)
		})
	}
}

// Tests for LinkPackage

func TestLinkPackage(t *testing.T) {
	testCases := []struct {
		name        string
		setupFunc   func(t *testing.T) (bl *BinLinker, pkgPath string)
		expectError bool
		validate    func(t *testing.T, bl *BinLinker)
	}{
		{
			name: "Link package with string bin field",
			setupFunc: func(t *testing.T) (*BinLinker, string) {
				tmpDir := t.TempDir()
				nodeModules := filepath.Join(tmpDir, "node_modules")
				os.MkdirAll(nodeModules, 0755)

				bl := NewBinLinker(nodeModules)
				bl.CreateBinDirectory()

				pkgPath := createTestPackage(t, nodeModules, "express", "./bin/express.js")
				return bl, pkgPath
			},
			expectError: false,
			validate: func(t *testing.T, bl *BinLinker) {
				verifySymlink(t, filepath.Join(bl.binPath, "express"), "../express/bin/express.js")
			},
		},
		{
			name: "Link package with object bin field",
			setupFunc: func(t *testing.T) (*BinLinker, string) {
				tmpDir := t.TempDir()
				nodeModules := filepath.Join(tmpDir, "node_modules")
				os.MkdirAll(nodeModules, 0755)

				bl := NewBinLinker(nodeModules)
				bl.CreateBinDirectory()

				bins := map[string]string{
					"cmd1": "./bin/cmd1.js",
					"cmd2": "./bin/cmd2.js",
				}
				pkgPath := createTestPackage(t, nodeModules, "multi-tool", bins)
				return bl, pkgPath
			},
			expectError: false,
			validate: func(t *testing.T, bl *BinLinker) {
				verifySymlink(t, filepath.Join(bl.binPath, "cmd1"), "../multi-tool/bin/cmd1.js")
				verifySymlink(t, filepath.Join(bl.binPath, "cmd2"), "../multi-tool/bin/cmd2.js")
			},
		},
		{
			name: "Link scoped package",
			setupFunc: func(t *testing.T) (*BinLinker, string) {
				tmpDir := t.TempDir()
				nodeModules := filepath.Join(tmpDir, "node_modules")
				os.MkdirAll(nodeModules, 0755)

				bl := NewBinLinker(nodeModules)
				bl.CreateBinDirectory()

				pkgPath := createScopedPackage(t, nodeModules, "@babel", "cli", "./bin/babel.js")
				return bl, pkgPath
			},
			expectError: false,
			validate: func(t *testing.T, bl *BinLinker) {
				verifySymlink(t, filepath.Join(bl.binPath, "cli"), "../@babel/cli/bin/babel.js")
			},
		},
		{
			name: "Skip - package.json doesn't exist",
			setupFunc: func(t *testing.T) (*BinLinker, string) {
				tmpDir := t.TempDir()
				nodeModules := filepath.Join(tmpDir, "node_modules")
				os.MkdirAll(nodeModules, 0755)

				bl := NewBinLinker(nodeModules)
				bl.CreateBinDirectory()

				pkgPath := filepath.Join(nodeModules, "no-package-json")
				os.MkdirAll(pkgPath, 0755)

				return bl, pkgPath
			},
			expectError: false,
			validate: func(t *testing.T, bl *BinLinker) {
				// No symlinks should be created
				entries, _ := os.ReadDir(bl.binPath)
				assert.Equal(t, 0, len(entries))
			},
		},
		{
			name: "Skip - invalid package.json",
			setupFunc: func(t *testing.T) (*BinLinker, string) {
				tmpDir := t.TempDir()
				nodeModules := filepath.Join(tmpDir, "node_modules")
				os.MkdirAll(nodeModules, 0755)

				bl := NewBinLinker(nodeModules)
				bl.CreateBinDirectory()

				pkgPath := filepath.Join(nodeModules, "invalid-json")
				os.MkdirAll(pkgPath, 0755)

				// Write invalid JSON
				os.WriteFile(filepath.Join(pkgPath, "package.json"), []byte(`{invalid json`), 0644)

				return bl, pkgPath
			},
			expectError: false,
			validate: func(t *testing.T, bl *BinLinker) {
				entries, _ := os.ReadDir(bl.binPath)
				assert.Equal(t, 0, len(entries))
			},
		},
		{
			name: "Skip - no bin field",
			setupFunc: func(t *testing.T) (*BinLinker, string) {
				tmpDir := t.TempDir()
				nodeModules := filepath.Join(tmpDir, "node_modules")
				os.MkdirAll(nodeModules, 0755)

				bl := NewBinLinker(nodeModules)
				bl.CreateBinDirectory()

				pkgPath := filepath.Join(nodeModules, "no-bin")
				os.MkdirAll(pkgPath, 0755)

				// Create package.json without bin field
				pkgJSON := map[string]string{"name": "no-bin", "version": "1.0.0"}
				data, _ := json.Marshal(pkgJSON)
				os.WriteFile(filepath.Join(pkgPath, "package.json"), data, 0644)

				return bl, pkgPath
			},
			expectError: false,
			validate: func(t *testing.T, bl *BinLinker) {
				entries, _ := os.ReadDir(bl.binPath)
				assert.Equal(t, 0, len(entries))
			},
		},
		{
			name: "Error - invalid bin field format",
			setupFunc: func(t *testing.T) (*BinLinker, string) {
				tmpDir := t.TempDir()
				nodeModules := filepath.Join(tmpDir, "node_modules")
				os.MkdirAll(nodeModules, 0755)

				bl := NewBinLinker(nodeModules)
				bl.CreateBinDirectory()

				pkgPath := filepath.Join(nodeModules, "invalid-bin")
				os.MkdirAll(pkgPath, 0755)

				// Create package.json with invalid bin field (number)
				pkgJSON := `{"name": "invalid-bin", "bin": 123}`
				os.WriteFile(filepath.Join(pkgPath, "package.json"), []byte(pkgJSON), 0644)

				return bl, pkgPath
			},
			expectError: true,
			validate:    func(t *testing.T, bl *BinLinker) {},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			bl, pkgPath := tc.setupFunc(t)
			err := bl.LinkPackage(pkgPath)

			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			tc.validate(t, bl)
		})
	}
}

// Tests for LinkAllPackages

func TestLinkAllPackages(t *testing.T) {
	testCases := []struct {
		name        string
		setupFunc   func(t *testing.T) *BinLinker
		expectError bool
		validate    func(t *testing.T, bl *BinLinker)
	}{
		{
			name: "Link multiple regular packages",
			setupFunc: func(t *testing.T) *BinLinker {
				tmpDir := t.TempDir()
				nodeModules := filepath.Join(tmpDir, "node_modules")
				os.MkdirAll(nodeModules, 0755)

				createTestPackage(t, nodeModules, "express", "./bin/express.js")
				createTestPackage(t, nodeModules, "jest", "./bin/jest.js")
				createTestPackage(t, nodeModules, "webpack", "./bin/webpack.js")

				return NewBinLinker(nodeModules)
			},
			expectError: false,
			validate: func(t *testing.T, bl *BinLinker) {
				verifySymlink(t, filepath.Join(bl.binPath, "express"), "../express/bin/express.js")
				verifySymlink(t, filepath.Join(bl.binPath, "jest"), "../jest/bin/jest.js")
				verifySymlink(t, filepath.Join(bl.binPath, "webpack"), "../webpack/bin/webpack.js")
			},
		},
		{
			name: "Link scoped packages",
			setupFunc: func(t *testing.T) *BinLinker {
				tmpDir := t.TempDir()
				nodeModules := filepath.Join(tmpDir, "node_modules")
				os.MkdirAll(nodeModules, 0755)

				createScopedPackage(t, nodeModules, "@babel", "cli", "./bin/babel.js")
				createScopedPackage(t, nodeModules, "@types", "node", "./bin/types.js")

				return NewBinLinker(nodeModules)
			},
			expectError: false,
			validate: func(t *testing.T, bl *BinLinker) {
				verifySymlink(t, filepath.Join(bl.binPath, "cli"), "../@babel/cli/bin/babel.js")
				verifySymlink(t, filepath.Join(bl.binPath, "node"), "../@types/node/bin/types.js")
			},
		},
		{
			name: "Mix of scoped and regular packages",
			setupFunc: func(t *testing.T) *BinLinker {
				tmpDir := t.TempDir()
				nodeModules := filepath.Join(tmpDir, "node_modules")
				os.MkdirAll(nodeModules, 0755)

				createTestPackage(t, nodeModules, "lodash", "./lodash.js")
				createScopedPackage(t, nodeModules, "@babel", "core", "./lib/index.js")
				createTestPackage(t, nodeModules, "react", "./index.js")

				return NewBinLinker(nodeModules)
			},
			expectError: false,
			validate: func(t *testing.T, bl *BinLinker) {
				verifySymlink(t, filepath.Join(bl.binPath, "lodash"), "../lodash/lodash.js")
				verifySymlink(t, filepath.Join(bl.binPath, "core"), "../@babel/core/lib/index.js")
				verifySymlink(t, filepath.Join(bl.binPath, "react"), "../react/index.js")
			},
		},
		{
			name: "Skip non-directory entries",
			setupFunc: func(t *testing.T) *BinLinker {
				tmpDir := t.TempDir()
				nodeModules := filepath.Join(tmpDir, "node_modules")
				os.MkdirAll(nodeModules, 0755)

				createTestPackage(t, nodeModules, "express", "./bin/cli.js")

				// Add a file in node_modules (should be skipped)
				os.WriteFile(filepath.Join(nodeModules, "README.md"), []byte("readme"), 0644)

				return NewBinLinker(nodeModules)
			},
			expectError: false,
			validate: func(t *testing.T, bl *BinLinker) {
				verifySymlink(t, filepath.Join(bl.binPath, "express"), "../express/bin/cli.js")

				// Only one symlink should exist
				entries, _ := os.ReadDir(bl.binPath)
				assert.Equal(t, 1, len(entries))
			},
		},
		{
			name: "Skip .bin directory itself",
			setupFunc: func(t *testing.T) *BinLinker {
				tmpDir := t.TempDir()
				nodeModules := filepath.Join(tmpDir, "node_modules")
				os.MkdirAll(nodeModules, 0755)

				// Create .bin directory first
				os.MkdirAll(filepath.Join(nodeModules, ".bin"), 0755)

				createTestPackage(t, nodeModules, "jest", "./bin/jest.js")

				return NewBinLinker(nodeModules)
			},
			expectError: false,
			validate: func(t *testing.T, bl *BinLinker) {
				// Should only have jest, not try to link .bin
				entries, _ := os.ReadDir(bl.binPath)
				assert.Equal(t, 1, len(entries))
				verifySymlink(t, filepath.Join(bl.binPath, "jest"), "../jest/bin/jest.js")
			},
		},
		{
			name: "Empty node_modules directory",
			setupFunc: func(t *testing.T) *BinLinker {
				tmpDir := t.TempDir()
				nodeModules := filepath.Join(tmpDir, "node_modules")
				os.MkdirAll(nodeModules, 0755)

				return NewBinLinker(nodeModules)
			},
			expectError: false,
			validate: func(t *testing.T, bl *BinLinker) {
				// .bin should be created but empty
				entries, err := os.ReadDir(bl.binPath)
				assert.NoError(t, err)
				assert.Equal(t, 0, len(entries))
			},
		},
		{
			name: "Success - node_modules doesn't exist but gets created",
			setupFunc: func(t *testing.T) *BinLinker {
				tmpDir := t.TempDir()
				nodeModules := filepath.Join(tmpDir, "nonexistent")

				return NewBinLinker(nodeModules)
			},
			expectError: false,
			validate: func(t *testing.T, bl *BinLinker) {
				// MkdirAll in CreateBinDirectory creates the parent
				info, err := os.Stat(bl.nodeModulesPath)
				assert.NoError(t, err)
				assert.True(t, info.IsDir())
			},
		},
		{
			name: "Mix of packages with and without bins",
			setupFunc: func(t *testing.T) *BinLinker {
				tmpDir := t.TempDir()
				nodeModules := filepath.Join(tmpDir, "node_modules")
				os.MkdirAll(nodeModules, 0755)

				createTestPackage(t, nodeModules, "with-bin", "./cli.js")

				// Create package without bin
				noBinPath := filepath.Join(nodeModules, "no-bin")
				os.MkdirAll(noBinPath, 0755)
				pkgJSON := map[string]string{"name": "no-bin"}
				data, _ := json.Marshal(pkgJSON)
				os.WriteFile(filepath.Join(noBinPath, "package.json"), data, 0644)

				return NewBinLinker(nodeModules)
			},
			expectError: false,
			validate: func(t *testing.T, bl *BinLinker) {
				// Only package with bin should be linked
				entries, _ := os.ReadDir(bl.binPath)
				assert.Equal(t, 1, len(entries))
				verifySymlink(t, filepath.Join(bl.binPath, "with-bin"), "../with-bin/cli.js")
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			bl := tc.setupFunc(t)
			err := bl.LinkAllPackages()

			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			tc.validate(t, bl)
		})
	}
}
