# Browser TODO

## In Progress
- [ ] Word wrapping for long text that exceeds container width

## Known Issues
- [ ] Whitespace between inline elements is missing (e.g., "Here is**bold**and" instead of "Here is **bold** and")
  - Spaces between inline elements like `<strong>`, `<em>`, `<small>` are not rendering
  - Need to debug DOM parser to see if whitespace text nodes are preserved
  - Options: debug DOM output, add space normalization, or fix parser whitespace handling
- [ ] `position: absolute` - text/color inside positioned elements not rendering
  - Background colors and borders of positioned elements work
  - Text inside positioned elements (like divs with `position: absolute`) is missing
  - Children of positioned elements are not being painted correctly
  - This affects z-index testing since z-index requires positioned elements

## Missing HTML Tags

### Text Formatting
- [ ] `<code>` - inline code
- [ ] `<kbd>` - keyboard input
- [ ] `<samp>` - sample output
- [ ] `<var>` - variable
- [ ] `<abbr>` - abbreviation
- [ ] `<cite>` - citation
- [x] `<q>` - inline quotation
- [ ] `<mark>` - highlighted text
- [ ] `<sub>` - subscript
- [ ] `<sup>` - superscript
- [x] `<del>` - deleted text (strikethrough)
- [x] `<ins>` - inserted text (underline)
- [ ] `<time>` - date/time
- [ ] `<dfn>` - definition term

### Lists
- [ ] `<dl>` - description list
- [ ] `<dt>` - description term
- [ ] `<dd>` - description details

### Media
- [ ] `<video>` - video player
- [ ] `<audio>` - audio player
- [ ] `<source>` - media source
- [ ] `<picture>` - responsive images
- [ ] `<figure>` - figure container
- [ ] `<figcaption>` - figure caption
- [ ] `<canvas>` - drawing canvas
- [ ] `<svg>` - vector graphics
- [ ] `<iframe>` - embedded frame

### Tables
- [ ] `<caption>` - table caption
- [ ] `<colgroup>` - column group
- [ ] `<col>` - column properties

### Forms
- [ ] `<fieldset>` - form group
- [ ] `<legend>` - fieldset caption
- [ ] `<datalist>` - input suggestions
- [ ] `<output>` - calculation result
- [ ] `<progress>` - progress bar
- [ ] `<meter>` - gauge/meter
- [ ] `<optgroup>` - option group

### Interactive
- [ ] `<details>` - collapsible content
- [ ] `<summary>` - details summary
- [ ] `<dialog>` - modal dialog

### Semantic
- [ ] `<aside>` - sidebar content
- [ ] `<address>` - contact info
- [ ] `<hgroup>` - heading group

## Missing CSS Properties

### Sizing
- [x] `width` - element width
- [x] `height` - element height
- [x] `min-width` - minimum width
- [x] `max-width` - maximum width
- [x] `min-height` - minimum height
- [x] `max-height` - maximum height

### Positioning
- [x] `position` - static/relative/absolute/fixed/sticky (partially: absolute works)
- [x] `top` - top offset (works for absolute)
- [x] `left` - left offset (works for absolute)
- [x] `right` - right offset
- [x] `bottom` - bottom offset
- [x] `z-index` - stacking order (needs sorting in paint.go)
- [x] `float` - float left/right
- [ ] `clear` - clear floats

### Flexbox
- [ ] `display: flex` - flex container
- [ ] `flex-direction` - row/column
- [ ] `justify-content` - main axis alignment
- [ ] `align-items` - cross axis alignment
- [ ] `align-content` - multi-line alignment
- [ ] `flex-wrap` - wrap items
- [ ] `gap` - spacing between items
- [ ] `flex-grow` - grow factor
- [ ] `flex-shrink` - shrink factor
- [ ] `flex-basis` - initial size
- [ ] `order` - item order
- [ ] `align-self` - individual alignment

### Grid
- [ ] `display: grid` - grid container
- [ ] `grid-template-columns` - column definitions
- [ ] `grid-template-rows` - row definitions
- [ ] `grid-gap` / `gap` - grid spacing
- [ ] `grid-column` - column span
- [ ] `grid-row` - row span

### Box Model
- [ ] `box-sizing` - border-box/content-box
- [ ] `overflow` - visible/hidden/scroll/auto
- [ ] `overflow-x` - horizontal overflow
- [ ] `overflow-y` - vertical overflow

### Typography
- [x] `font-family` - font stack
- [ ] `line-height` - line spacing
- [ ] `letter-spacing` - character spacing
- [ ] `word-spacing` - word spacing
- [x] `text-transform` - uppercase/lowercase/capitalize
- [ ] `text-shadow` - text shadow
- [ ] `white-space` - whitespace handling
- [ ] `text-overflow` - ellipsis/clip
- [ ] `text-indent` - first line indent
- [ ] `vertical-align` - inline alignment

### Background
- [ ] `background` - shorthand
- [ ] `background-image` - image/gradient
- [ ] `background-position` - position
- [ ] `background-size` - cover/contain/size
- [ ] `background-repeat` - repeat/no-repeat

### Border
- [x] `border-radius` - rounded corners
- [ ] `border-top-left-radius` - individual corner
- [ ] `border-top-right-radius` - individual corner
- [ ] `border-bottom-left-radius` - individual corner
- [ ] `border-bottom-right-radius` - individual corner

### Effects
- [ ] `box-shadow` - drop shadow
- [ ] `transform` - rotate/scale/translate
- [ ] `transform-origin` - transform center
- [ ] `transition` - animated changes
- [ ] `animation` - keyframe animations
- [ ] `filter` - blur/brightness/etc

### List
- [ ] `list-style` - shorthand
- [ ] `list-style-type` - disc/circle/square/decimal/none
- [ ] `list-style-position` - inside/outside
- [ ] `list-style-image` - custom marker

### Table
- [ ] `border-collapse` - collapse/separate
- [ ] `border-spacing` - cell spacing
- [ ] `table-layout` - auto/fixed

### Othernd related tests
- [ ] `outline` - focus outline
- [ ] `content` - generated content
- [ ] `pointer-events` - click behavior
- [ ] `user-select` - text selection

### Units (parsing)
- [ ] `em` - relative to font size
- [ ] `rem` - relative to root font size
- [ ] `%` - percentage
- [ ] `vw` - viewport width
- [ ] `vh` - viewport height
- [ ] `calc()` - calculations
- [ ] `rgb()` / `rgba()` - color functions
- [ ] `hsl()` / `hsla()` - color functions

## Future Features
- [ ] Forward navigation button
- [ ] Keyboard shortcuts (Ctrl+R refresh, Alt+Left back)
- [ ] JavaScript support

## Refactoring
- [ ] Refactor Rect usage pattern:
  ```go
  X:     box.Rect.X,
  Y:     box.Rect.Y,
  Width: box.Rect.Width,
  ```

## Testing
- [ ] http://acid1.acidtests.org
- [ ] http://acid2.acidtests.org


https://github.com/web-platform-tests/wpt?tab=readme-ov-file

