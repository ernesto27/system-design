# Browser TODO


csswg-drafts

tc39-ecma262 for JavaScript

whatwg-dom

whatwg-html
https://html.spec.whatwg.org/


# HTML SPECS

[x] - <p>
[x] - <html>
[x] - <head> (WHATWG 4.2.1 compliance)
[x] - <title> (WHATWG 4.2.2 compliance - document.title getter/setter)
[x] - <style> (WHATWG 4.2.6 compliance - disabled property)
[x] - <body onbeforeunload> - Navigation warning event

### HTMLBodyElement (WHATWG 4.3.1)
- [ ] `document.body` setter - Allow setting body element
- [ ] `onload` event - Window load event on body (High priority)
- [ ] `ononline` / `onoffline` - Network status events
- [ ] `onhashchange` - URL hash navigation
- [ ] `onpopstate` - History API
- [ ] `onmessage` - postMessage API
- [ ] `onstorage` - localStorage events
- [ ] `onpagehide` / `onpageshow` - Page visibility events
- [ ] `HTMLBodyElement` interface - Proper DOM interface
### HTMLStyleElement (WHATWG 4.2.6)
- [x] `styleElement.disabled` - Getter/setter to enable/disable stylesheet
- [ ] `styleElement.sheet` - Get associated CSSStyleSheet object (LinkStyle interface)
- [ ] `styleElement.media` - Get/set media query string
- [ ] `styleElement.type` - Validate type attribute (only "text/css" or empty)
- [ ] `styleElement.title` - Style sheet set name for alternate stylesheets
- [ ] `load` event - Fire when style processing completes
- [ ] `error` event - Fire when style loading fails

### CSSStyleSheet (CSSOM)
- [ ] `sheet.cssRules` - Get list of CSS rules
- [ ] `sheet.insertRule(rule, index)` - Add a CSS rule
- [ ] `sheet.deleteRule(index)` - Remove a CSS rule
- [ ] `sheet.disabled` - Enable/disable the stylesheet




## Related TODO Files

| File | Purpose |
|------|---------|
| `TODO-HTML.md` | HTML tags implementation |
| `TODO-CSS.md` | CSS properties implementation |
| `TODO-JS.md` | JavaScript implementation |
| `test-todo.md` | Test coverage tracking |

---


# BUGS 

[] - Text overlapp with bold and normal text
[x] - allow select text
[] - Partial text selection (select characters within a line, not entire text boxes)
[] - show indication of CTRL-C copied text 

## In Progress
- [ ] Word wrapping for long text that exceeds container width

---

## Known Issues
- [ ] Whitespace between inline elements is missing (e.g., "Here is**bold**and" instead of "Here is **bold** and")
  - Spaces between inline elements like `<strong>`, `<em>`, `<small>` are not rendering
  - Need to debug DOM parser to see if whitespace text nodes are preserved
- [ ] `position: absolute` - text/color inside positioned elements not rendering
  - Background colors and borders of positioned elements work
  - Text inside positioned elements is missing
  - Children of positioned elements are not being painted correctly

---

## Future Features
- [ ] Forward navigation button
- [ ] Keyboard shortcuts (Ctrl+R refresh, Alt+Left back)
- [ ] Browser history (back/forward)
- [ ] Bookmarks
- [ ] Multiple tabs

---

## Refactoring
- [ ] Refactor Rect usage pattern:
  ```go
  X:     box.Rect.X,
  Y:     box.Rect.Y,
  Width: box.Rect.Width,
  ```

---

## Testing Resources
- [ ] http://acid1.acidtests.org
- [ ] http://acid2.acidtests.org
- https://github.com/web-platform-tests/wpt
  