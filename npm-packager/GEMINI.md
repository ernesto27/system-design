# Gemini Project Context: Go NPM Packager

## Project Overview

This project is a high-performance, custom-built npm package manager written in Go. Its primary purpose is to handle the download and installation of npm packages, including full dependency resolution. The Go application reads a `package.json` file, resolves version constraints for all dependencies and devDependencies, fetches the package tarballs from the npm registry, and extracts them into the `node_modules` directory.

The project also contains a sample Node.js/Express application, defined by `package.json` and `index.js`, which serves as a test case for the Go package manager.

### Key Technologies
*   **Core Application:** Go
*   **Testing/Example App:** Node.js, Express
*   **Go Modules:** `golang.org/x/mod/semver` for versioning, `github.com/tidwall/gjson` for JSON parsing.

### Architecture
The Go application is structured into several packages:
*   `main.go`: The entry point and CLI handler for commands like `i`, `add`, and `rm`.
*   `manager`: Orchestrates the dependency resolution and installation process.
*   `manifest`: Handles fetching and parsing package manifests from the npm registry.
*   `tarball`: Manages downloading package tarballs (`.tgz` files).
*   `extractor`: Responsible for securely extracting tarball contents.
*   `packagejson`: Parses the initial `package.json` file.
*   `utils`: Contains shared utility functions.

The dependency resolution uses a Breadth-First Search (BFS) approach to build the dependency tree and avoid duplicate processing.

## Building and Running

### Core Go Application
To build and run the Go package manager:

```bash
# Build the executable
go build -o npm-packager

# Run the packager to install dependencies from package.json
./npm-packager i
```

### Sample Node.js Application
To run the sample Node.js application (after installing dependencies with the Go packager):

```bash
# Start the server
node index.js
```

## Development

### Testing
Run the Go test suite using the following command:

```bash
go test -v ./...
```

### Dependencies
To manage Go dependencies:

```bash
# Download Go dependencies
go mod download

# Tidy up the go.mod and go.sum files
go mod tidy
```
