# Missing Tests TODO

## Overview
Current test coverage analysis for the browser project.

---

## Layout Package (5% coverage - CRITICAL)

### `hittest.go` - DONE
- [x] `TestContains` - Point-in-box detection (24 test cases)
  - Point inside/outside box
  - Boundary conditions (edges, corners)
  - Zero-dimension boxes
- [x] `TestHitTest` - Finding box at coordinates (9 test cases)
  - Point outside returns nil
  - Nested boxes return deepest child
  - Sibling overlap handling
- [x] `TestFindLink` - Walking up tree for `<a>` tags (10 test cases)
  - Direct `<a>` element with href
  - Parent/grandparent is `<a>`
  - No link ancestor returns ""

### `measure.go` - DONE
- [x] `TestMeasureText` - Text width measurement (14 test cases)
  - Empty string returns 0
  - Estimation formula: `len(text) * fontSize * 0.5`
  - Custom TextMeasurer callback
- [x] `TestMeasureTextFormula` - Formula verification (5 test cases)
- [x] `TestWrapText` - Text wrapping (9 test cases)
  - Empty text, fits in one line, word too long
  - Wrapping to multiple lines
  - Zero/negative maxWidth handling

### `layout.go` - DONE
- [x] `TestBuildLayoutTreeBoxTypes` - Box type assignments (52 test cases)
- [x] `TestBuildLayoutTreeDisplayNone` - display:none handling (2 test cases)
- [x] `TestBuildLayoutTreeStructure` - Parent/children linking (2 test cases)
- [x] `TestBuildLayoutTreeStyles` - Stylesheet application (4 test cases)
- [x] `TestMergeStyles` - Style merging (16 test cases)

### `compute.go` - DONE
- [x] `TestGetLineHeight` - Tag to line height mapping (12 test cases)
- [x] `TestGetFontSize` - Tag to font size mapping (12 test cases)
- [x] `TestGetImageSize` - Image dimension parsing (8 test cases)
- [x] `TestIsInsidePre` - Pre element detection (4 test cases)
- [x] `TestGetButtonText` - Button text extraction (4 test cases)
- [x] `TestApplyLineAlignment` - Text alignment (6 test cases)
- [x] `TestComputeLayout` - Main layout computation (9 test cases)

---

## CSS Package (41% coverage)

### `css.go`
- [ ] `TestApplyStylesheet` - Core styling function
  - Applies CSS rules to nodes
  - Selector matching (tag, id, class)
  - Specificity handling
- [ ] `TestDefaultStyle` - Default style factory

---

## Render Package

### `canvas.go` - PARTIAL
- [x] `TestIsLocalFile` - Local file detection (10 test cases)
  - file:// protocol
  - Absolute paths
  - HTTP URLs (should return false)
  - Protocol-relative URLs (should return false)
- [x] `TestToLocalPath` - file:// URL to path conversion (5 test cases)
- [x] `TestResolveImageURL` - URL resolution (8 test cases)
  - HTTP/HTTPS absolute URLs
  - Protocol-relative URLs
  - Local file paths
  - Relative paths with base URL

### `display.go`
- [ ] `TestBuildDisplayList` - Creates drawing commands
- [ ] `TestPaintLayoutBox` - Paints layout box to commands

### `render.go`
- [ ] `TestRenderToCanvas` - Converts commands to Fyne objects

### `browser.go`
- [ ] `TestHandleClick` - Click event handling
- [ ] `TestHandleTypedRune` - Text input handling
- [ ] `TestSubmitForm` - Form submission

### `utils.go`
- [ ] `TestApplyOpacity` - Opacity calculation
- [ ] `TestIsValidEmail` - Email validation

---

## DOM Package

### `dom.go`
- [x] `TestFindElementsByTagName` - Find element by tag name (7 test cases)
  - Find head/body elements
  - Find nested elements
  - Element not found returns nil
  - Nil node returns nil
  - Find first of multiple matching elements
  - Find root element itself

---

## JS Package

### `runtime.go`
- [ ] `TestWrapElementCaching` - Same DOM node returns same JS object
- [ ] `TestDocumentHead` - document.head returns head element
- [ ] `TestDocumentBody` - document.body returns body element
- [ ] `TestDocumentTitle` - document.title getter/setter
- [ ] `TestParentElementNull` - Root element parentElement is null

---

## Compliance Test Pages

| File | Tests |
|------|-------|
| `testpage/html_compliance.html` | Root element (WHATWG 4.1.1) |
| `testpage/head_compliance.html` | Head element (WHATWG 4.2.1) |
| `testpage/title_compliance.html` | Title element (WHATWG 4.2.2) |

---

## Priority Order

1. **Layout package** - Core rendering pipeline
2. **CSS ApplyStylesheet** - Style application
3. **Render package** - Display and interaction
4. **JS package** - DOM bindings and compliance

---

## Testing Patterns to Follow

Based on existing tests in `box_test.go`:
- Table-driven tests with `t.Run()`
- `github.com/stretchr/testify/assert`
- Descriptive test case names
