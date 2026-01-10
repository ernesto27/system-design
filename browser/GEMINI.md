# Browser Engine Project

## Overview
This project is a custom web browser engine written in Go from scratch. It is designed for educational purposes to demonstrate the core components of a browser: fetching, parsing, layout, and rendering. It uses `fyne` for the GUI window and `golang.org/x/net/html` for HTML tokenization.

## Architecture
The browser follows a classic 5-stage rendering pipeline:
1.  **Network**: Fetches content via HTTP (`main.go`, `http.Get`).
2.  **HTML Parsing**: Converts HTML text into a DOM tree (`dom/` package).
3.  **Layout**: Converts the DOM tree into a Layout tree (Box Model) with computed positions and sizes (`layout/` package).
4.  **Painting**: Converts the Layout tree into a list of display commands (`render/` package).
5.  **Rasterization/Display**: Draws the display commands to a Fyne canvas.

### Directory Structure
*   `dom/`: Defines the Document Object Model. Nodes, attributes, and tree traversal.
*   `layout/`: The layout engine. Handles the Box Model, block formatting contexts, and dimension calculations.
*   `render/`: Interaction with the GUI framework (Fyne). Handles painting and window management.
*   `css/`: CSS parsing logic. *Note: Full CSS integration is currently in planning/progress (see `CSS_INTEGRATION_PLAN.md`).*
*   `testpage/`: Contains `index.html` for manual testing.
*   `main.go`: Entry point. Orchestrates the pipeline.

## Build and Run

### Prerequisites
*   Go 1.21+
*   Fyne dependencies (typically OpenGL libraries for your OS)

### Commands
To run the browser with a specific URL:
```bash
go run . https://example.com
```

To run with the local test page (requires serving the file or passing a file URI if supported, though the code uses `http.Get`):
```bash
# Start a simple server in another terminal if needed, or pass a live URL
go run . http://localhost:8080/index.html
```

To build a binary:
```bash
go build -o browser
./browser https://google.com
```

## Development Conventions
*   **Language**: Go (Idiomatic).
*   **Error Handling**: Standard Go error returns.
*   **Testing**: Currently, there are no unit tests (`*_test.go`). Testing is primarily manual using the provided `testpage/index.html`.
*   **Style**: Standard `gofmt`.
*   **Architecture**: strict separation of concerns between DOM (data), Layout (geometry), and Render (GUI).

## Current Status & Roadmap
*   **HTML**: Basic parsing and DOM tree construction are implemented.
*   **Layout**: Basic block layout is working.
*   **CSS**: In active development. The goal is to move to `tdewolff/parse` for robust CSS support. See `CSS_INTEGRATION_PLAN.md` for details.


## Testing Conventions

Use table-driven tests with `testify/assert`:

```go
func TestFunctionName(t *testing.T) {
    tests := []struct {
        name     string
        input    string
        expected string
    }{
        {"case one", "input1", "expected1"},
        {"case two", "input2", "expected2"},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := FunctionName(tt.input)
            assert.Equal(t, tt.expected, result)
        })
    }
}
```

Run tests:
```bash
go test ./... -v           # All tests
go test ./layout/... -v    # Package tests
go test ./... -cover       # With coverage
```