# AGENTS.md

Instructions for AI coding agents working in this Go browser codebase.

## Project Overview

A web browser built from scratch in Go for educational purposes. It implements a 5-stage rendering pipeline:

```
URL -> HTTP Fetch -> DOM Tree -> Layout Tree -> Display Commands -> GUI
        (main.go)    (dom/)       (layout/)      (render/)         (render/)
```

### Package Structure

| Package   | Purpose                                    |
|-----------|--------------------------------------------|
| `dom/`    | HTML parsing, DOM tree representation      |
| `css/`    | CSS parsing (parsed but not yet applied)   |
| `layout/` | Box model, position/size computation       |
| `render/` | Fyne GUI, painting, click handling         |
| `main.go` | Pipeline orchestration, HTTP fetching      |

---

## Build Commands

```bash
# Build binary
go build -o browser

# Run with URL
go run . https://example.com

# Run against local test server
./serve.sh                           # Terminal 1: starts server on :8080
go run . http://localhost:8080       # Terminal 2: open browser
```

## Test Commands

```bash
# Run all tests
go test ./...

# Run tests in a specific package
go test ./dom

# Run a single test by name
go test ./dom -run TestNormalizeWhitespace

# Run a specific subtest (table-driven tests)
go test ./dom -run TestNormalizeWhitespace/empty_string

# Verbose output
go test -v ./...

# With coverage
go test -cover ./...
go test -coverprofile=coverage.out ./... && go tool cover -html=coverage.out
```

## Lint/Format Commands

```bash
# Format code (required before commit)
gofmt -w .
go fmt ./...

# Vet for common issues
go vet ./...

# Optional: install and run staticcheck
go install honnef.co/go/tools/cmd/staticcheck@latest
staticcheck ./...
```

---

## Code Style Guidelines

### Import Organization

Group imports in this order, separated by blank lines:
1. Standard library
2. External packages
3. Local project packages

```go
import (
    "fmt"
    "io"
    "net/http"

    "fyne.io/fyne/v2"

    "browser/dom"
    "browser/layout"
)
```

### Naming Conventions

| Element            | Convention   | Example                          |
|--------------------|--------------|----------------------------------|
| Exported types     | PascalCase   | `LayoutBox`, `NodeType`          |
| Unexported types   | camelCase    | `blockElements`                  |
| Exported functions | PascalCase   | `BuildLayoutTree`, `Parse`       |
| Unexported funcs   | camelCase    | `convertNode`, `paintLayoutBox`  |
| Constructors       | `New` prefix | `NewElement`, `NewBrowser`       |
| Constants          | PascalCase   | `TagBody`, `BlockBox`            |
| Method receivers   | Single letter| `(n *Node)`, `(box *LayoutBox)`  |
| Local variables    | Short names  | `n`, `err`, `sb`, `child`        |

### Constants

Use `iota` for enumerations. Group related constants:

```go
const (
    BlockBox BoxType = iota
    InlineBox
    TextBox
)
```

### Error Handling

```go
// Pattern 1: Return nil on error (for parsers)
if err != nil {
    return nil
}

// Pattern 2: Log and continue (for non-critical operations)
if err != nil {
    fmt.Println("Error:", err)
    return
}

// Pattern 3: Early return on nil
if node == nil {
    return ""
}
```

### Methods

Use pointer receivers with single-letter receiver names:

```go
func (n *Node) AppendChild(child *Node) {
    child.Parent = n
    n.Children = append(n.Children, child)
}
```

---

## Testing Patterns

Use table-driven tests with `t.Run()` subtests:

```go
func TestSomething(t *testing.T) {
    tests := []struct {
        name     string
        input    string
        expected string
    }{
        {"empty input", "", ""},
        {"basic case", "hello", "hello"},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := functionUnderTest(tt.input)
            if result != tt.expected {
                t.Errorf("got %q, want %q", result, tt.expected)
            }
        })
    }
}
```

- Use `%q` for strings in error messages (shows quotes and escapes)
- Test files go in the same package (e.g., `dom/parser_test.go`)

---

## Dependencies

- `golang.org/x/net/html` - HTML tokenizer/parser
- `fyne.io/fyne/v2` - Cross-platform GUI framework

---

## Key Files Reference

| File                  | Purpose                                  |
|-----------------------|------------------------------------------|
| `main.go`             | Entry point, HTTP fetch, navigation      |
| `dom/dom.go`          | Node type, tree structure                |
| `dom/parser.go`       | HTML parsing via x/net/html              |
| `layout/box.go`       | LayoutBox type, box model                |
| `layout/compute.go`   | Layout algorithm, positioning            |
| `render/paint.go`     | Display commands, colors, font sizes     |
| `render/window.go`    | Fyne window, browser chrome              |
