# JavaScript Implementation TODO

## Current Progress

- [x] **Step 1:** Install goja and create basic JS runtime
- [x] **Step 2:** Create document object with getElementById
- [x] **Step 3:** Build Element wrapper with properties
- [x] **Step 4:** Implement event system (addEventListener)
- [ ] **Step 5:** Add DOM mutation (innerHTML, appendChild)
- [x] **Step 6:** Execute script tags in rendering pipeline
- [x] **Step 7:** Trigger reflow on DOM changes

---

## Step 5: DOM Mutation (Next Up)

### Element Creation
- [ ] `document.createElement(tagName)` - Create new elements
- [ ] `document.createTextNode(text)` - Create text nodes

### Element Modification
- [x] `element.innerHTML` getter - Get HTML content as string
- [x] `element.innerHTML` setter - Parse and replace children
- [ ] `element.innerText` getter/setter - Text without HTML tags
- [ ] `element.appendChild(child)` - Add child at end
- [ ] `element.removeChild(child)` - Remove a child
- [ ] `element.insertBefore(new, ref)` - Insert before reference node
- [ ] `element.replaceChild(new, old)` - Replace a child
- [ ] `element.remove()` - Remove self from parent

### Attribute Manipulation
- [ ] `element.classList.add(class)` - Add CSS class
- [ ] `element.classList.remove(class)` - Remove CSS class
- [ ] `element.classList.toggle(class)` - Toggle CSS class
- [ ] `element.classList.contains(class)` - Check for class
- [ ] `element.style.property` - Get/set inline styles

---

## Step 7: Reflow on DOM Changes

- [ ] Detect when DOM is modified (appendChild, innerHTML, etc.)
- [ ] Rebuild layout tree from modified DOM
- [ ] Recompute layout positions
- [ ] Repaint to screen
- [ ] Debounce rapid changes (batch updates)

---

## Future: Document Methods

### Query Methods
- [x] `document.querySelector(selector)` - Find first match
- [ ] `document.querySelectorAll(selector)` - Find all matches
- [ ] `document.getElementsByClassName(class)` - Find by class
- [ ] `document.getElementsByTagName(tag)` - Find by tag
- [ ] `element.querySelector(selector)` - Scoped query
- [ ] `element.querySelectorAll(selector)` - Scoped query all

### Document Properties
- [ ] `document.body` - Get body element
- [ ] `document.head` - Get head element
- [ ] `document.title` - Get/set page title
- [ ] `document.URL` - Get current URL

---

## Future: More Events

### Mouse Events
- [ ] `mousedown` / `mouseup`
- [ ] `mouseover` / `mouseout`
- [ ] `mousemove`
- [ ] `dblclick`

### Keyboard Events
- [ ] `keydown` / `keyup`
- [ ] `keypress`
- [ ] `input` (for form fields)

### Form Events
- [ ] `submit`
- [ ] `change`
- [ ] `focus` / `blur`

### Event Features
- [ ] `event.preventDefault()`
- [ ] `event.stopPropagation()`
- [ ] Event bubbling (child → parent)
- [ ] Event capturing (parent → child)

---

## Future: Window Object

### Timers
- [ ] `setTimeout(fn, ms)` - Run once after delay
- [ ] `setInterval(fn, ms)` - Run repeatedly
- [ ] `clearTimeout(id)` - Cancel timeout
- [ ] `clearInterval(id)` - Cancel interval

### Dialogs
- [x] `alert(message)` - Show alert dialog (non-blocking)
- [x] `confirm(message)` - Yes/No dialog (blocking, returns boolean)
- [ ] `prompt(message)` - Input dialog

### Navigation
- [ ] `window.location.href` - Get/set URL
- [ ] `window.location.reload()` - Refresh page
- [ ] `window.history.back()` - Go back
- [ ] `window.history.forward()` - Go forward

### Window Properties
- [ ] `window.innerWidth` / `innerHeight`
- [ ] `window.scrollX` / `scrollY`
- [ ] `window.scroll(x, y)` / `scrollTo(x, y)`

---

## Future: Storage

- [ ] `localStorage.getItem(key)`
- [ ] `localStorage.setItem(key, value)`
- [ ] `localStorage.removeItem(key)`
- [ ] `localStorage.clear()`
- [ ] `sessionStorage` (same API)

---

## Future: Network

- [ ] `fetch(url)` - Basic GET requests
- [ ] `fetch(url, options)` - POST, headers, etc.
- [ ] `Response.json()` - Parse JSON response
- [ ] `Response.text()` - Get text response
- [ ] `XMLHttpRequest` (legacy support)

---

## Future: Advanced

### Promises & Async
- [ ] Promise support (goja has this)
- [ ] async/await syntax
- [ ] `Promise.all()`, `Promise.race()`

### JSON
- [ ] `JSON.parse(string)` - Parse JSON
- [ ] `JSON.stringify(obj)` - Convert to JSON

### Console
- [x] `console.log()` - Basic logging
- [ ] `console.error()` - Error logging (red)
- [ ] `console.warn()` - Warning logging (yellow)
- [ ] `console.table()` - Table format

---

## Architecture Notes

### Files Structure
```
js/
├── runtime.go      # Goja setup, script execution
├── document.go     # document object bindings
├── element.go      # Element wrapper with methods
├── events.go       # Event system (addEventListener)
├── window.go       # window object (TODO)
├── storage.go      # localStorage (TODO)
└── fetch.go        # fetch API (TODO)
```

### Key Patterns

1. **Wrap dom.Node** - Never expose raw Go pointers to JS
2. **Use DefineAccessorProperty** - For live bindings (textContent, innerHTML)
3. **Store callbacks with *dom.Node key** - Unique identity for event matching
4. **Trigger reflow on mutation** - Keep visual in sync with DOM

---

## Testing Checklist

- [ ] console.log with multiple arguments
- [ ] document.getElementById finds elements
- [ ] document.getElementById returns null for missing
- [ ] element.textContent reads correctly
- [ ] element.textContent writes and triggers reflow
- [ ] addEventListener registers callback
- [ ] Click dispatches to correct element
- [ ] Multiple listeners on same element
- [ ] DOM mutation updates visual display
