# Browser TODO

## In Progress
- [ ] Word wrapping for long text that exceeds container width

## Known Issues
- [ ] Whitespace between inline elements is missing (e.g., "Here is**bold**and" instead of "Here is **bold** and")
  - Spaces between inline elements like `<strong>`, `<em>`, `<small>` are not rendering
  - Need to debug DOM parser to see if whitespace text nodes are preserved
  - Options: debug DOM output, add space normalization, or fix parser whitespace handling

## Future Features
- [ ] CSS styling support
- [ ] Forward navigation button
- [ ] Keyboard shortcuts (Ctrl+R refresh, Alt+Left back)
- [ ] Form support (input fields, buttons)
- [ ] `<pre>` / `<code>` tags (monospace text)
- [ ] `<mark>` tag (highlighted background)
