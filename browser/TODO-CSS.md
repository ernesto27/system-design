# CSS Implementation TODO

## Completed
- [x] `color` - text color
- [x] `background-color` - background color
- [x] `background-image` - url() images (remote)
- [x] `background-image` - local files (file:// and absolute paths)
- [x] `font-size` - text size
- [x] `font-weight` - bold
- [x] `font-style` - italic
- [x] `font-family` - font stack
- [x] `margin` - all sides
- [x] `margin-top/right/bottom/left` - individual margins
- [x] `padding` - all sides
- [x] `padding-top/right/bottom/left` - individual padding
- [x] `text-align` - left/center/right
- [x] `text-decoration` - underline/line-through
- [x] `text-transform` - uppercase/lowercase/capitalize
- [x] `border` - shorthand
- [x] `border-width/color/style` - border properties
- [x] `border-top/right/bottom/left` - individual borders
- [x] `border-radius` - rounded corners
- [x] `width` - element width
- [x] `height` - element height
- [x] `min-width` - minimum width
- [x] `max-width` - maximum width
- [x] `min-height` - minimum height
- [x] `max-height` - maximum height
- [x] `display` - block/inline/none
- [x] `position` - static/relative/absolute
- [x] `top/left/right/bottom` - position offsets
- [x] `z-index` - stacking order
- [x] `float` - left/right
- [x] `opacity` - transparency
- [x] `visibility` - visible/hidden
- [x] `cursor` - pointer/text/crosshair
- [x] `em` unit - relative to parent font size
- [x] User-agent default styles (margins for p, h1-h6, ul, ol, blockquote, hr)

---

## In Progress
- [ ] Word wrapping for long text

---

## Missing Properties

### Background
- [ ] `background` - shorthand
- [ ] `background-position` - position
- [ ] `background-size` - cover/contain/size
- [ ] `background-repeat` - repeat/no-repeat

### Border
- [ ] `border-top-left-radius` - individual corner
- [ ] `border-top-right-radius` - individual corner
- [ ] `border-bottom-left-radius` - individual corner
- [ ] `border-bottom-right-radius` - individual corner

### Box Model
- [ ] `box-sizing` - border-box/content-box
- [ ] `overflow` - visible/hidden/scroll/auto
- [ ] `overflow-x` - horizontal overflow
- [ ] `overflow-y` - vertical overflow
- [ ] `clear` - clear floats

### Typography
- [ ] `line-height` - line spacing
- [ ] `letter-spacing` - character spacing
- [ ] `word-spacing` - word spacing
- [ ] `text-shadow` - text shadow
- [ ] `white-space` - whitespace handling
- [ ] `text-overflow` - ellipsis/clip
- [ ] `text-indent` - first line indent
- [ ] `vertical-align` - inline alignment

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

### Other
- [ ] `outline` - focus outline
- [ ] `content` - generated content
- [ ] `pointer-events` - click behavior
- [ ] `user-select` - text selection

### Units (parsing)
- [x] `em` - relative to font size
- [ ] `rem` - relative to root font size
- [ ] `%` - percentage
- [ ] `vw` - viewport width
- [ ] `vh` - viewport height
- [ ] `calc()` - calculations
- [ ] `rgb()` / `rgba()` - color functions
- [ ] `hsl()` / `hsla()` - color functions

### Selectors
- [ ] Descendant selectors - `div p`, `ul li`
- [ ] Child selectors - `ul > li`
- [ ] Pseudo-classes - `:hover`, `:focus`, `:active`, `:first-child`, `:last-child`
- [ ] Pseudo-elements - `::before`, `::after`
- [ ] Attribute selectors - `[type="text"]`, `[href^="https"]`
- [ ] Sibling selectors - `h1 + p`, `h1 ~ p`

### Cascade & Specificity
- [ ] Specificity calculation - proper weighting (inline > id > class > tag)
- [ ] `!important` - override rules
- [ ] Inheritance - properties inheriting from parent elements

### Shorthand Expansion
- [ ] `margin` multi-value - `margin: 10px 20px`, `margin: 10px 20px 30px 40px`
- [ ] `padding` multi-value - `padding: 10px 20px`
- [ ] `border-radius` multi-value - per-corner values

---

## Parsed But Not Applied
- [ ] `cursor` - parsed but not applied in render
- [ ] `display: block/inline` - only `none` actually works
- [ ] `position: relative/fixed/sticky` - only `absolute` works
- [ ] `z-index` - parsed but stacking may not work correctly

---

## Known Issues
- [ ] `position: absolute` - text/color inside positioned elements not rendering
