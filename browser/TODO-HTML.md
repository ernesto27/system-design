# HTML Implementation TODO

## Completed Tags

### Document Structure
- [x] `<html>` - root element
- [x] `<head>` - document head
- [x] `<body>` - document body
- [x] `<title>` - page title
- [x] `<link>` - external resources (stylesheets)
- [x] `<style>` - embedded CSS (WHATWG 4.2.6: disabled property support)
- [x] `<script>` - JavaScript

### Text Content
- [x] `<h1>` - `<h6>` - headings
- [x] `<p>` - paragraph
- [x] `<br>` - line break
- [x] `<hr>` - horizontal rule
- [x] `<pre>` - preformatted text
- [x] `<blockquote>` - block quotation
- [x] `<div>` - division
- [x] `<span>` - inline container

### Text Formatting
- [x] `<strong>` / `<b>` - bold
- [x] `<em>` / `<i>` - italic
- [x] `<u>` - underline
- [x] `<small>` - smaller text
- [x] `<del>` - deleted text (strikethrough)
- [x] `<ins>` - inserted text (underline)
- [x] `<q>` - inline quotation

### Links & Media
- [x] `<a>` - hyperlink
- [x] `<img>` - image

### Lists
- [x] `<ul>` - unordered list
- [x] `<ol>` - ordered list
- [x] `<li>` - list item
- [x] `<dl>` - description list
- [x] `<dt>` - description term
- [x] `<dd>` - description details

### Tables
- [x] `<table>` - table
- [x] `<thead>` - table header group
- [x] `<tbody>` - table body group
- [x] `<tfoot>` - table footer group
- [x] `<tr>` - table row
- [x] `<th>` - table header cell
- [x] `<td>` - table data cell

### Forms
- [x] `<form>` - form container
- [x] `<input>` - text/password/email/number/checkbox/radio/file
- [x] `<button>` - button
- [x] `<textarea>` - multiline text
- [x] `<select>` - dropdown
- [x] `<option>` - select option
- [x] `<label>` - form label
- [x] `<fieldset>` - form group (basic implementation, see Known Issues)
- [x] `<legend>` - fieldset caption (basic implementation, see Known Issues)

### Semantic
- [x] `<header>` - header section
- [x] `<footer>` - footer section
- [x] `<main>` - main content
- [x] `<nav>` - navigation
- [x] `<section>` - section
- [x] `<article>` - article

---

## Missing Tags

### Text Formatting
- [ ] `<code>` - inline code
- [ ] `<kbd>` - keyboard input
- [ ] `<samp>` - sample output
- [ ] `<var>` - variable
- [ ] `<abbr>` - abbreviation
- [ ] `<cite>` - citation
- [ ] `<mark>` - highlighted text
- [ ] `<sub>` - subscript
- [ ] `<sup>` - superscript
- [ ] `<time>` - date/time
- [ ] `<dfn>` - definition term

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
- [x] `<caption>` - table caption (centered text)
- [ ] `<colgroup>` - column group
- [ ] `<col>` - column properties

### Forms
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

### Document Metadata
- [x] `<base>` - base URL for relative links
- [ ] `<meta>` - document metadata (viewport, charset, etc.)
- [ ] `<noscript>` - fallback for no JavaScript

### Embedded Content
- [ ] `<embed>` - external content plugin
- [ ] `<object>` - embedded object
- [ ] `<param>` - object parameter
- [ ] `<map>` - image map container
- [ ] `<area>` - image map clickable area

### Ruby Annotations (East Asian text)
- [ ] `<ruby>` - ruby annotation container
- [ ] `<rt>` - ruby text (pronunciation)
- [ ] `<rp>` - ruby fallback parenthesis

### Text Direction & Breaks
- [ ] `<wbr>` - word break opportunity
- [ ] `<bdi>` - bidirectional isolation
- [ ] `<bdo>` - bidirectional override

### Web Components
- [ ] `<template>` - content template (not rendered)
- [ ] `<slot>` - web component slot

---

## Missing Input Types
- [ ] `type="date"` - date picker
- [ ] `type="time"` - time picker
- [ ] `type="datetime-local"` - date and time picker
- [ ] `type="month"` - month picker
- [ ] `type="week"` - week picker
- [ ] `type="color"` - color picker
- [ ] `type="range"` - slider control
- [ ] `type="search"` - search field
- [ ] `type="tel"` - telephone input
- [ ] `type="url"` - URL input
- [ ] `type="hidden"` - hidden field

---

## Missing Attributes & Features

### Form Validation
- [x] `required` - required field validation (red border + prevents submit)
- [ ] `pattern` - regex validation
- [ ] `min` / `max` - number range validation
- [ ] `minlength` / `maxlength` - text length validation
- [ ] `step` - number increment
- [ ] `:valid` / `:invalid` pseudo-classes

### Table Features
- [ ] `colspan` - cell column span
- [ ] `rowspan` - cell row span
- [ ] `scope` - header cell scope

### Link Features
- [x] `target="_blank"` - open in new window (basic implementation)
- [ ] `rel="noopener"` - security for external links (parsed but not enforced)
- [ ] `download` - download link

### Image Features
- [ ] `srcset` - responsive image sources
- [ ] `sizes` - responsive image sizes
- [ ] `loading="lazy"` - lazy loading
- [ ] `alt` text display on error

### Accessibility (ARIA)
- [ ] `role` - element role
- [ ] `aria-label` - accessible label
- [ ] `aria-hidden` - hide from screen readers
- [ ] `aria-expanded` - expandable state
- [ ] `aria-describedby` - description reference
- [ ] `tabindex` - keyboard navigation order

### Global Attributes
- [ ] `contenteditable` - editable content
- [ ] `draggable` - drag and drop
- [ ] `hidden` - hide element
- [ ] `title` - tooltip on hover
- [ ] `lang` - language specification
- [ ] `data-*` - custom data attributes

---

## Known Issues
- [ ] Whitespace between inline elements missing (e.g., `<strong>`, `<em>`)
- [ ] Text inside `position: absolute` elements not rendering
- [x] `<main>`, `<nav>`, `<section>`, `<article>` added to blockElements map
- [ ] No keyboard navigation between form elements (Tab key)
- [x] No form validation feedback UI (implemented red border for required fields)
- [ ] Images don't show alt text on load failure
- [ ] `<fieldset>` legend spacing needs fine-tuning (gap between legend text and border)
- [ ] `<fieldset>` without legend shows no top border (basic fieldset case)

---

## Future Enhancements

### Performance
- [ ] Incremental layout (don't recompute entire tree)
- [ ] Virtual scrolling for long pages
- [ ] Image caching to disk
- [ ] Lazy image loading

### User Experience
- [ ] Text selection and copy
- [ ] Find in page (Ctrl+F)
- [ ] Zoom in/out
- [ ] Print page
- [ ] View page source
- [ ] Developer tools panel

### Standards Compliance
- [ ] DOCTYPE handling
- [ ] Character encoding detection
- [ ] Quirks mode vs standards mode
- [ ] HTML entity decoding (`&nbsp;`, `&amp;`, etc.)
