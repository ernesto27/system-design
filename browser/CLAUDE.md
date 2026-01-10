# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

A web browser built from scratch in Go for educational purposes. It fetches HTML pages, parses them into a DOM tree, computes layout, and renders to a GUI window using Fyne.

## Build and Run

```bash
# Build
go build -o browser

# Run
./browser <url>
./browser https://example.com

# Or directly
go run . https://example.com
```

## Architecture

The browser follows a 5-stage rendering pipeline:

```
URL → HTTP Fetch → DOM Tree → Layout Tree → Display Commands → GUI
```

### Package Structure

| Package | Purpose |
|---------|---------|
| `dom/` | HTML parsing, DOM tree representation |
| `css/` | CSS parsing (parsed but not yet applied) |
| `layout/` | Box model, position/size computation, hit testing |
| `render/` | Fyne GUI, painting, click handling |
| `main.go` | Pipeline orchestration |

### Key Data Flow

1. **dom.Parse()** - HTML string → `*dom.Node` tree
2. **layout.BuildLayoutTree()** - DOM tree → `*layout.LayoutBox` tree
3. **layout.ComputeLayout()** - Calculates X, Y, Width, Height for each box
4. **render.BuildDisplayList()** - Layout tree → `[]DisplayCommand` (DrawRect, DrawText)
5. **render.RenderToCanvas()** - Display commands → Fyne canvas objects

### Layout Model

- Block elements stack vertically (div, p, h1-h6)
- Text boxes get height based on parent tag (h1=40px, h2=32px, default=24px)
- Body has 8px margin
- Width flows down (parent → child), height flows up (children → parent)

### Click Handling

1. `ClickableContainer` captures tap events with (x, y) coordinates
2. `LayoutBox.HitTest(x, y)` finds deepest box containing the point
3. `LayoutBox.FindLink()` walks up parent chain looking for `<a>` tags
4. URL resolved relative to current page, then `OnNavigate` callback fired

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

## Dependencies

- `golang.org/x/net/html` - HTML tokenizer/parser
- `fyne.io/fyne/v2` - Cross-platform GUI framework
