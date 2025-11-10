# AGENTS.md

## Build/Lint/Test Commands

### Go Commands
```bash
go run main.go              # Run the package manager (uses package.json in current directory)
go run main.go i            # Install dependencies from package.json
go run main.go add <pkg>@<version>  # Add a package dependency
go build -o npm-packager    # Build binary
go test                     # Run all tests
go test -v                  # Run tests with verbose output
go test -run TestName       # Run specific test
go test -v -run TestDownloadManifest  # Run specific test with verbose output
go mod tidy                 # Clean up dependencies
go mod download             # Download dependencies
```

### Node.js Commands
```bash
npm start                   # Start production server
npm run dev                 # Start development server with nodemon
npm run build               # Build TypeScript to JavaScript
npm run lint                # Run ESLint with auto-fix
npm run format              # Format code with Prettier
npm test                    # Run Jest tests
npm run test:watch          # Run tests in watch mode
```

## Architecture and Codebase Structure

### Core Components
- **PackageManager** (`main.go`): Central orchestrator managing downloads, caching, and dependency resolution
- **Concurrent Architecture**: Uses goroutines with worker pools, mutexes for thread safety
- **Caching Strategy**: Manifests and tarballs cached in `~/.config/go-npm/`
- **Dependency Resolution**: Breadth-first search with queue-based processing

### Key Modules
- **manifest/**: Downloads and caches npm package manifests from registry
- **tarball/**: Downloads .tgz files from npm registry
- **extractor/**: Extracts tarballs with path traversal security checks
- **packagejson/**: Parses package.json and manages lock files
- **version/**: Resolves npm version ranges (^, ~, ranges) to specific versions
- **etag/**: HTTP caching using ETags for manifest downloads
- **utils/**: Common utilities for file operations and downloads

### Data Flow
1. Parse local package.json
2. Download package manifests from npm registry
3. Resolve version constraints to specific versions
4. Download and extract tarballs
5. Recursively process dependencies with deduplication
6. Generate go-package-lock.json

## Code Style Guidelines

### Go Code Style
- **Imports**: Standard library first, then third-party, then local modules
- **Error Handling**: Use `fmt.Errorf` with `%w` verb for error wrapping
- **Naming**: PascalCase for exported, camelCase for unexported
- **Structs**: Use JSON tags for serialization, omitempty for optional fields
- **Concurrency**: Use sync.Mutex/RWMutex, channels for coordination
- **Testing**: Table-driven tests with testify/assert, use `t.TempDir()` for isolation

### JavaScript/TypeScript Style
- **ESLint**: Recommended rules with TypeScript plugin
- **Prettier**: Semi-colons enabled, double quotes, 100 char width, trailing commas
- **Testing**: Jest with ts-jest preset, tests in `/tests` directory
- **Node Version**: >= 18.0.0

### Testing Conventions
- **Go Tests**: Table-driven pattern with setupFunc, expectError, validate functions
- **Assertions**: Use `github.com/stretchr/testify/assert`
- **Mocking**: Use testify's mock package where needed
- **Coverage**: Aim for comprehensive test coverage

### File Organization
- **Go**: One package per directory, main package in root
- **Node.js**: Source in `src/`, compiled to `dist/`
- **Config**: package.json for Node.js, go.mod for Go dependencies

### Include CLAUDE.md Rules
- Add comments only if logic is too complex
- Use buffered I/O for performance (32KB buffers)
- Security: Strip `package/` prefix to prevent path traversal
- Scoped packages: Extract name after `/` for tarball URLs
