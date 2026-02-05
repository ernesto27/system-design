# NPM Package Manager (Go)

A high-performance npm package manager written in Go that downloads and installs npm packages with full dependency resolution.



## Features

- **Recursive Dependency Resolution**: BFS queue-based approach prevents duplicate processing
- **Version Range Support**: Handles `^1.0.0`, `~2.3.4`, `>=1.2.3 <2.0.0`, and exact versions
- **Intelligent Caching**: Manifests and tarballs cached in `~/.config/go-npm/`
- **Security**: Path traversal protection during tarball extraction
- **Scoped Packages**: Full support for `@types/node`, `@babel/core`, etc.
- **DevDependencies**: Installs both dependencies and devDependencies

## Quick Start

```bash
# Build
go build -o npm-packager

# Run (requires package.json in current directory)
./npm-packager
```

## Example package.json

```json
{
  "name": "my-project",
  "version": "1.0.0",
  "dependencies": {
    "express": "^4.18.0",
    "lodash": "~4.17.21",
    "@types/node": "^18.0.0"
  },
  "devDependencies": {
    "jest": "^29.5.0"
  }
}
```

## How It Works

1. **Parse**: Read `package.json` and extract dependencies
2. **Queue**: Initialize BFS queue with all dependencies
3. **Resolve**: For each dependency:
   - Download manifest from `https://registry.npmjs.org/<package>`
   - Resolve version constraint to exact version using semver
   - Download tarball (`.tgz` file)
   - Extract to `node_modules/<package>` with security checks
   - Add sub-dependencies to queue
4. **Repeat**: Process queue until empty (all transitive dependencies installed)

## Core Components

| Component | File | Purpose |
|-----------|------|---------|
| PackageManager | `main.go` | Orchestrates dependency resolution (BFS) |
| DownloadManifest | `downloadmanifest.go` | Fetches package metadata from npm registry |
| VersionInfo | `version.go` | Resolves `^`, `~`, ranges to exact versions |
| ParseJsonManifest | `parseJson.go` | Parses npm manifest and package.json |
| TGZExtractor | `extractor.go` | Extracts tarballs with path traversal protection |

## Testing

```bash
go test -v                          # All tests
go test -run TestDownloadManifest   # Specific test
go test -v -run TestVersion         # Verbose output
```

Tests use table-driven patterns with `setupFunc`, `expectError`, and `validate` functions for comprehensive coverage.

## Cache Structure

```
~/.config/go-npm/
├── manifest/              # Package metadata cache
│   ├── express
│   └── lodash
└── tarball/               # Downloaded .tgz files
    ├── express-4.18.2.tgz
    └── lodash-4.17.21.tgz
```

## Technical Details

- **NPM Registry**: `https://registry.npmjs.org/`
- **Semver Library**: `golang.org/x/mod/semver`
- **Extraction Buffer**: 32KB for optimal I/O performance
- **Version Resolution**: Falls back to `dist-tags.latest` for unrecognized patterns
- **Concurrent Safety**: Tracks processed packages to avoid duplicates

## Dependencies

```bash
go mod download  # Download Go dependencies
go mod tidy      # Clean up go.mod
```

## License

MIT
