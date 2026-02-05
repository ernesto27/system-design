# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a Go implementation of an npm package manager that downloads and installs npm packages and their dependencies. It reads a `package.json` file, resolves dependencies recursively, downloads package manifests and tarballs from the npm registry, and extracts them into a local `node_modules` directory.

## Commands

### Build and Run
```bash
go run main.go              # Run the package manager (uses package.json in current directory)
go build -o npm-packager    # Build binary
```

### Testing
```bash
go test                     # Run all tests
go test -v                  # Run tests with verbose output
go test -run TestName       # Run specific test
go test -v -run TestDownloadManifest  # Run specific test with verbose output
```

### Dependencies
```bash
go mod tidy                 # Clean up dependencies
go mod download             # Download dependencies
```

## Architecture

### Core Components

**PackageManager** (`main.go:19-69`)
- Central orchestrator that manages the entire download process
- Initializes configuration paths: `~/.config/go-npm/{manifest,tarball}`
- Uses breadth-first search (queue-based) to resolve dependencies recursively
- Tracks processed packages to avoid duplicate downloads

**Dependency Resolution Flow**
1. `PackageJSONParser` parses local `package.json`
2. `PackageManager.downloadDependencies()` processes each dependency:
   - `DownloadManifest` fetches package metadata from npm registryvalidate
   - `ParseJsonManifest` parses the manifest to get available versions
   - `VersionInfo` resolves version constraints (^, ~, ranges, exact)
   - `DownloadTarball` downloads the .tgz file
   - `TGZExtractor` extracts to `node_modules/<package>`
   - Sub-dependencies are added to queue and processed recursively

### Key Modules

**downloadmanifest.go**
- Downloads package manifests from `https://registry.npmjs.org/<package>`
- Caches manifests in `~/.config/go-npm/manifest/`
- Skips re-download if manifest already exists

**version.go**
- Resolves npm version ranges to specific versions
- Supports: `^` (caret), `~` (tilde), complex ranges (e.g., `>= 2.1.2 < 3.0.0`)
- Uses `golang.org/x/mod/semver` for semantic version comparison
- Falls back to `dist-tags.latest` for unrecognized patterns

**parseJson.go**
- Parses npm registry manifest JSON (different from package.json)
- Contains full package metadata including all versions and their metadata
- Key structure: `NPMPackage.Versions` maps version strings to `Version` structs

**extractor.go**
- Extracts .tgz files with security checks against path traversal
- Strips `package/` prefix from tar entries
- Uses buffered I/O (32KB buffer) for performance
- Only extracts regular files (skips directories, symlinks)

**utils.go**
- `downloadFile()`: HTTP downloader with error handling
- `createDir()`: Directory creation helper

### Data Structures

**PackageJSON** vs **NPMPackage**
- `PackageJSON`: Local package.json format (simple dependencies map)
- `NPMPackage`: Registry manifest format (contains all versions, dist-tags, full metadata)

**Version Resolution**
- Version strings from package.json are resolved against `NPMPackage.Versions` map
- Result is an exact version string used to construct tarball URL

## Testing Conventions

Tests follow table-driven test pattern with three key functions:
- `setupFunc`: Creates test environment (uses `t.TempDir()` for isolation)
- `expectError`: Boolean flag for error assertions
- `validate`: Post-execution validation logic

Example test structure (see `packagejson_test.go` and `downloadmanifest_test.go`):
```go
testCases := []struct {
    name        string
    setupFunc   func(t *testing.T) (params...)
    expectError bool
    validate    func(t *testing.T, params...)
}{...}
```

Use `github.com/stretchr/testify/assert` for assertions.

## Important Paths

- **Config directory**: `~/.config/go-npm/`
- **Manifest cache**: `~/.config/go-npm/manifest/`
- **Tarball cache**: `~/.config/go-npm/tarball/`
- **Extraction target**: `./node_modules/`
- **Input file**: `./package.json`

## NPM Registry URL

All package downloads use: `https://registry.npmjs.org/`
- Manifest: `https://registry.npmjs.org/<package>`
- Tarball: `https://registry.npmjs.org/<package>/-/<tarball-name>-<version>.tgz`

Note: Scoped packages (e.g., `@types/node`) require special tarball name handling - extract the part after `/` for the tarball filename.
- add comments only if logic is too complex