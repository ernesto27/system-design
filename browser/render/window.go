package render

import (
	"browser/css"
	"browser/dom"
	"browser/layout"
	"fmt"
	"image/color"
	"net/url"
	"strconv"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

type Browser struct {
	App          fyne.App
	Window       fyne.Window
	Width        float32
	Height       float32
	layoutTree   *layout.LayoutBox
	currentURL   *url.URL
	currentStyle css.Stylesheet
	OnNavigate   func(newURL string)

	urlEntry   *widget.Entry
	content    *fyne.Container
	history    []string
	historyPos int

	document *dom.Node

	// Input state - keyed by DOM node (stable across reflow)
	focusedInputNode *dom.Node
	inputValues      map[*dom.Node]string
	openSelectNode   *dom.Node // Which select dropdown is open
	radioValues      map[string]*dom.Node
	checkboxValue    map[*dom.Node]bool
}

func NewBrowser(width, height float32) *Browser {
	a := app.New()
	w := a.NewWindow("Go Browser")
	w.Resize(fyne.NewSize(width, height))

	// Set up accurate text measurement using Fyne
	layout.TextMeasurer = func(text string, fontSize float64, bold bool, italic bool) float64 {
		style := fyne.TextStyle{Bold: bold, Italic: italic}
		size := fyne.MeasureText(text, float32(fontSize), style)
		return float64(size.Width)
	}

	b := &Browser{
		App:           a,
		Window:        w,
		Width:         width,
		Height:        height,
		history:       []string{},
		historyPos:    -1,
		inputValues:   make(map[*dom.Node]string),
		radioValues:   make(map[string]*dom.Node),
		checkboxValue: make(map[*dom.Node]bool),
	}
	// Create URL entry
	b.urlEntry = widget.NewEntry()
	b.urlEntry.SetPlaceHolder("Enter URL...")

	b.urlEntry.OnSubmitted = func(text string) {
		if text != "" {
			if b.OnNavigate != nil {
				b.OnNavigate(text)
			}
		}
	}

	// Create buttons
	goBtn := widget.NewButton("Go", func() {
		url := b.urlEntry.Text
		if url != "" {
			if b.OnNavigate != nil {
				b.OnNavigate(url)
			}
		}
	})

	backBtn := widget.NewButton("←", func() {
		b.GoBack()
	})

	refreshBtn := widget.NewButton("↻", func() {
		b.Refresh()
	})

	// Toolbar: [Back] [URL Entry] [Go]
	toolbar := container.NewBorder(
		nil, nil, // top, bottom
		container.NewHBox(backBtn, refreshBtn), goBtn, // left, right
		b.urlEntry, // center (fills remaining space)
	)

	// Content area (NewMax makes children fill available space)
	b.content = container.NewMax()

	// Main layout: toolbar on top, content below
	main := container.NewBorder(
		toolbar, nil, nil, nil, // top, bottom, left, right
		b.content, // center
	)

	w.Canvas().SetOnTypedRune(func(r rune) {
		b.handleTypedRune(r)
	})
	w.Canvas().SetOnTypedKey(func(key *fyne.KeyEvent) {
		b.handleTypedKey(key)
	})

	go func() {
		var lastWidth float32
		for {
			size := w.Canvas().Size()
			if size.Width != lastWidth && size.Width > 0 {
				lastWidth = size.Width
				b.Reflow(size.Width)
			}
			// Check every 100ms
			time.Sleep(100 * time.Millisecond)
		}
	}()

	w.SetContent(main)

	return b
}

func (b *Browser) SetContent(layoutTree *layout.LayoutBox) {
	b.layoutTree = layoutTree // Save it so handleClick can use it

	commands := BuildDisplayList(layoutTree)

	// Get base URL for resolving relative image URLs
	baseURL := ""
	if b.currentURL != nil {
		baseURL = b.currentURL.Scheme + "://" + b.currentURL.Host
	}

	objects := RenderToCanvas(commands, baseURL, false) // false = fetch fresh

	// content := container.NewWithoutLayout(objects...
	// Create clickable container
	clickable := NewClickableContainer(objects, func(x, y float32) {
		b.handleClick(float64(x), float64(y))
	}, b.layoutTree)

	scroll := container.NewScroll(clickable)

	b.content.Objects = []fyne.CanvasObject{scroll}
	b.content.Refresh()
}

func (b *Browser) AddToHistory(url string) {
	if b.historyPos < len(b.history)-1 {
		b.history = b.history[:b.historyPos+1]
	}

	b.history = append(b.history, url)
	b.historyPos = len(b.history) - 1
}

func (b *Browser) GoBack() {
	if b.historyPos > 0 {
		b.historyPos--
		prevURL := b.history[b.historyPos]
		b.urlEntry.SetText(prevURL)

		if b.OnNavigate != nil {
			b.OnNavigate(prevURL)
		}
	}
}

func (b *Browser) SetCurrentURL(rawURL string) {
	parsed, err := url.Parse(rawURL)
	if err == nil {
		b.currentURL = parsed
	}
}

func (b *Browser) handleClick(x, y float64) {
	fmt.Printf("Click at (%.0f, %.0f)\n", x, y)

	if b.layoutTree == nil {
		fmt.Println("  layoutTree is nil!")
		return
	}

	// Hit test: find what was clicked
	hit := b.layoutTree.HitTest(x, y)
	if hit == nil {
		fmt.Println("  No hit found")
		if b.focusedInputNode != nil {
			b.focusedInputNode = nil
			b.repaint()
		}
		return
	}
	fmt.Printf("  Hit: %+v\n", hit.Text)

	if hit.Type == layout.InputBox && hit.Node != nil {
		if isNodeDisabled(hit.Node) {
			return
		}

		inputType := strings.ToLower(hit.Node.Attributes["type"])

		// Handle number input spin buttons
		if inputType == "number" {
			buttonWidth := 24.0
			btnX := hit.Rect.X + hit.Rect.Width - buttonWidth

			// Check if click is in spin button area
			if x >= btnX {
				midY := hit.Rect.Y + hit.Rect.Height/2
				currentVal := b.inputValues[hit.Node]
				num := parseNumber(currentVal)

				if y < midY {
					// Up button clicked - increment
					num++
				} else {
					// Down button clicked - decrement
					num--
				}

				b.inputValues[hit.Node] = formatNumber(num)
				b.focusedInputNode = hit.Node
				b.repaint()
				return
			}
		}

		fmt.Print("click input box")
		b.focusedInputNode = hit.Node // Store DOM node, not LayoutBox
		b.repaint()
		return
	}

	if hit.Type == layout.RadioBox && hit.Node != nil {
		if isNodeDisabled(hit.Node) {
			return
		}
		fmt.Println("click radio button")
		name := hit.Node.Attributes["name"]
		if name != "" {
			b.radioValues[name] = hit.Node
		}
		b.repaint()
		return
	}

	if hit.Type == layout.CheckboxBox && hit.Node != nil {
		if isNodeDisabled(hit.Node) {
			return
		}
		fmt.Println("click checkbox")
		b.checkboxValue[hit.Node] = !b.checkboxValue[hit.Node]
		b.repaint()
		return
	}

	if hit.Type == layout.TextareaBox && hit.Node != nil {
		if isNodeDisabled(hit.Node) {
			return
		}
		fmt.Println("click textarea")
		b.focusedInputNode = hit.Node
		b.openSelectNode = nil // Close any open select
		b.repaint()
		return
	}

	if hit.Type == layout.SelectBox && hit.Node != nil {
		if isNodeDisabled(hit.Node) {
			return
		}
		fmt.Println("click select")
		if b.openSelectNode == hit.Node {
			// Already open - close it
			b.openSelectNode = nil
		} else {
			// Open this select
			b.openSelectNode = hit.Node
			b.focusedInputNode = nil // Unfocus any input
		}
		b.repaint()
		return
	}

	// Check if clicked on a select option (when dropdown is open)
	if b.openSelectNode != nil {
		// Check if we clicked an option by checking y position
		selectBox := b.findSelectBox(b.openSelectNode)
		if selectBox != nil {
			optionHeight := 28.0
			optionY := selectBox.Rect.Y + selectBox.Rect.Height
			numOptions := b.countSelectOptions(b.openSelectNode)

			// Check if click is in dropdown area
			if y >= optionY && y < optionY+float64(numOptions)*optionHeight &&
				x >= selectBox.Rect.X && x < selectBox.Rect.X+selectBox.Rect.Width {
				// Calculate which option was clicked
				optionIndex := int((y - optionY) / optionHeight)
				optionValue := b.getSelectOptionByIndex(b.openSelectNode, optionIndex)
				if optionValue != "" {
					fmt.Println("  Selected option:", optionValue)
					b.inputValues[b.openSelectNode] = optionValue
					b.openSelectNode = nil // Close dropdown
					b.repaint()
					return
				}
			}
		}
	}

	// Check if it's a link
	href := hit.FindLink()
	if href == "" {
		fmt.Println("  Not a link")
		if b.focusedInputNode != nil || b.openSelectNode != nil {
			b.focusedInputNode = nil
			b.openSelectNode = nil
			b.repaint()
		}
		return
	}
	fmt.Println("  Found link:", href)

	// Resolve relative URL
	fullURL := b.resolveURL(href)
	fmt.Println("Link clicked:", fullURL)

	// Call navigation callback
	if b.OnNavigate != nil {
		b.OnNavigate(fullURL)
	}
}

func (b *Browser) resolveURL(href string) string {
	if b.currentURL == nil {
		return href
	}

	// Parse the href
	parsed, err := url.Parse(href)
	if err != nil {
		return href
	}

	// Resolve relative to current URL
	resolved := b.currentURL.ResolveReference(parsed)
	return resolved.String()
}

// UpdateURLBar sets the text in the URL entry field
func (b *Browser) UpdateURLBar(url string) {
	b.urlEntry.SetText(url)
}

func (b *Browser) ShowLoading() {
	// White background
	bg := canvas.NewRectangle(ColorWhite)
	bg.Resize(fyne.NewSize(b.Width, b.Height))
	bg.Move(fyne.NewPos(0, 0))

	// Loading text - centered
	loading := canvas.NewText("Loading...", ColorBlack)
	loading.TextSize = 18
	loading.Alignment = fyne.TextAlignCenter

	// Use a center container to position the text
	centered := container.NewCenter(loading)

	// Stack background and centered text
	stack := container.NewStack(bg, centered)

	b.content.Objects = []fyne.CanvasObject{stack}
	b.content.Refresh()
}

func (b *Browser) Run() {
	b.Window.ShowAndRun()
}

func (b *Browser) SetTitle(title string) {
	if title == "" {
		title = "Go Browser"
	}
	b.Window.SetTitle(title)
}

func (b *Browser) SetDocument(doc *dom.Node) {
	b.document = doc
}

func (b *Browser) SetStylesheet(stylesheet css.Stylesheet) {
	b.currentStyle = stylesheet
}

// Reflow re-computes layout with new width and repaints
func (b *Browser) Reflow(width float32) {
	if b.document == nil {
		return
	}

	// Re-build layout tree with new width
	layoutTree := layout.BuildLayoutTree(b.document, b.currentStyle)
	layout.ComputeLayout(layoutTree, float64(width))

	// Update stored values
	b.Width = width
	b.layoutTree = layoutTree

	// Repaint with input state preserved (uses DOM node keys, stable across reflow)
	commands := BuildDisplayListWithInputs(layoutTree, InputState{
		InputValues:    b.inputValues,
		FocusedNode:    b.focusedInputNode,
		OpenSelectNode: b.openSelectNode,
		RadioValues:    b.radioValues,
		CheckboxValues: b.checkboxValue,
	})

	baseURL := ""
	if b.currentURL != nil {
		baseURL = b.currentURL.Scheme + "://" + b.currentURL.Host
	}

	// Use cached images on reflow (don't re-fetch)
	objects := RenderToCanvas(commands, baseURL, true) // true = use cache

	// UI updates must be on main thread
	fyne.Do(func() {
		// Preserve scroll position
		var scrollOffset fyne.Position
		if len(b.content.Objects) > 0 {
			if oldScroll, ok := b.content.Objects[0].(*container.Scroll); ok {
				scrollOffset = oldScroll.Offset
			}
		}

		clickable := NewClickableContainer(objects, func(x, y float32) {
			b.handleClick(float64(x), float64(y))
		}, b.layoutTree)

		scroll := container.NewScroll(clickable)
		scroll.Offset = scrollOffset // Restore scroll position

		b.content.Objects = []fyne.CanvasObject{scroll}
		b.content.Refresh()
	})
}

func (b *Browser) ShowError(message string) {
	fyne.Do(func() {
		bg := canvas.NewRectangle(ColorWhite)
		bg.Resize(fyne.NewSize(b.Width, b.Height))

		// Error title
		title := canvas.NewText("Error", color.RGBA{200, 0, 0, 255})
		title.TextSize = 24
		title.TextStyle = fyne.TextStyle{Bold: true}

		// Error message
		msg := canvas.NewText(message, ColorBlack)
		msg.TextSize = 16

		// Stack title and message vertically
		content := container.NewVBox(title, msg)
		centered := container.NewCenter(content)
		stack := container.NewStack(bg, centered)

		b.content.Objects = []fyne.CanvasObject{stack}
		b.content.Refresh()
	})
}

func (b *Browser) Refresh() {
	if b.currentURL != nil && b.OnNavigate != nil {
		b.OnNavigate(b.currentURL.String())
	}
}

func (b *Browser) handleTypedRune(r rune) {
	if b.focusedInputNode == nil {
		return
	}

	if isNodeDisabled(b.focusedInputNode) || isNodeReadonly(b.focusedInputNode) {
		return
	}

	// Filter input for number type
	inputType := strings.ToLower(b.focusedInputNode.Attributes["type"])
	if inputType == "number" && !isNumericRune(r) {
		return // Ignore non-numeric input
	}

	// Add character to input value
	current := b.inputValues[b.focusedInputNode]
	b.inputValues[b.focusedInputNode] = current + string(r)

	// Re-render to show new text
	b.refreshContent()
}

func (b *Browser) handleTypedKey(key *fyne.KeyEvent) {
	if b.focusedInputNode == nil {
		return
	}

	switch key.Name {
	case fyne.KeyBackspace:
		if isNodeDisabled(b.focusedInputNode) || isNodeReadonly(b.focusedInputNode) {
			return
		}
		current := b.inputValues[b.focusedInputNode]
		if len(current) > 0 {
			// Remove last character (handle UTF-8)
			runes := []rune(current)
			b.inputValues[b.focusedInputNode] = string(runes[:len(runes)-1])
			b.repaint()
		}
	case fyne.KeyReturn, fyne.KeyEnter:
		if isNodeDisabled(b.focusedInputNode) || isNodeReadonly(b.focusedInputNode) {
			return
		}
		if b.focusedInputNode.TagName == "textarea" {
			current := b.inputValues[b.focusedInputNode]
			b.inputValues[b.focusedInputNode] = current + "\n"
			b.repaint()
		}
	case fyne.KeyEscape:
		// Unfocus on escape
		b.focusedInputNode = nil
		b.openSelectNode = nil
		b.repaint()
	}
}

// findSelectBox finds the LayoutBox for a given select DOM node
func (b *Browser) findSelectBox(node *dom.Node) *layout.LayoutBox {
	return findBoxByNode(b.layoutTree, node)
}

func findBoxByNode(box *layout.LayoutBox, node *dom.Node) *layout.LayoutBox {
	if box == nil {
		return nil
	}
	if box.Node == node {
		return box
	}
	for _, child := range box.Children {
		if found := findBoxByNode(child, node); found != nil {
			return found
		}
	}
	return nil
}

// countSelectOptions counts the number of <option> children
func (b *Browser) countSelectOptions(selectNode *dom.Node) int {
	count := 0
	for _, child := range selectNode.Children {
		if child.TagName == "option" {
			count++
		}
	}
	return count
}

// getSelectOptionByIndex returns the text of the option at the given index
func (b *Browser) getSelectOptionByIndex(selectNode *dom.Node, index int) string {
	current := 0
	for _, child := range selectNode.Children {
		if child.TagName == "option" {
			if current == index {
				// Get text content
				for _, textNode := range child.Children {
					if textNode.Type == dom.Text {
						return textNode.Text
					}
				}
				return ""
			}
			current++
		}
	}
	return ""
}

// repaint re-renders the current layout tree without recalculating layout
func (b *Browser) repaint() {
	if b.layoutTree == nil {
		return
	}

	commands := BuildDisplayListWithInputs(b.layoutTree, InputState{
		InputValues:    b.inputValues,
		FocusedNode:    b.focusedInputNode,
		OpenSelectNode: b.openSelectNode,
		RadioValues:    b.radioValues,
		CheckboxValues: b.checkboxValue,
	})

	baseURL := ""
	if b.currentURL != nil {
		baseURL = b.currentURL.Scheme + "://" + b.currentURL.Host
	}

	objects := RenderToCanvas(commands, baseURL, true)

	fyne.Do(func() {
		// Preserve scroll position
		var scrollOffset fyne.Position
		if len(b.content.Objects) > 0 {
			if oldScroll, ok := b.content.Objects[0].(*container.Scroll); ok {
				scrollOffset = oldScroll.Offset
			}
		}

		clickable := NewClickableContainer(objects, func(x, y float32) {
			b.handleClick(float64(x), float64(y))
		}, b.layoutTree)

		scroll := container.NewScroll(clickable)
		scroll.Offset = scrollOffset // Restore scroll position
		b.content.Objects = []fyne.CanvasObject{scroll}
		b.content.Refresh()
	})
}

// refreshContent is an alias for repaint (called by keyboard handlers)
func (b *Browser) refreshContent() {
	b.repaint()
}

// isNodeDisabled checks if a DOM node has the disabled attribute
func isNodeDisabled(node *dom.Node) bool {
	if node == nil {
		return false
	}
	_, disabled := node.Attributes["disabled"]
	return disabled
}

// isNodeReadonly checks if a DOM node has the readonly attribute
func isNodeReadonly(node *dom.Node) bool {
	if node == nil {
		return false
	}
	_, readonly := node.Attributes["readonly"]
	return readonly
}

// parseNumber parses a string to int, returns 0 if invalid
func parseNumber(s string) int {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0
	}
	n, err := strconv.Atoi(s)
	if err != nil {
		return 0
	}
	return n
}

// formatNumber formats an int to string
func formatNumber(n int) string {
	return strconv.Itoa(n)
}

// isNumericRune checks if a rune is valid for number input
func isNumericRune(r rune) bool {
	return (r >= '0' && r <= '9') || r == '-'
}
