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
  