package render

import (
	"browser/css"
	"browser/dom"
	"browser/layout"
	"bytes"
	"fmt"
	"image/color"
	"mime/multipart"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

type NavigationRequest struct {
	URL         string
	Method      string
	Data        url.Values
	Body        []byte
	ContentType string
}

type Browser struct {
	App          fyne.App
	Window       fyne.Window
	Width        float32
	Height       float32
	layoutTree   *layout.LayoutBox
	currentURL   *url.URL
	currentStyle css.Stylesheet
	OnNavigate   func(req NavigationRequest)

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
	fileInputValues  map[*dom.Node]string

	onJSClick func(node *dom.Node)
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
		App:             a,
		Window:          w,
		Width:           width,
		Height:          height,
		history:         []string{},
		historyPos:      -1,
		inputValues:     make(map[*dom.Node]string),
		radioValues:     make(map[string]*dom.Node),
		checkboxValue:   make(map[*dom.Node]bool),
		fileInputValues: make(map[*dom.Node]string),
	}
	// Create URL entry
	b.urlEntry = widget.NewEntry()
	b.urlEntry.SetPlaceHolder("Enter URL...")

	b.urlEntry.OnSubmitted = func(text string) {
		if text != "" {
			if b.OnNavigate != nil {
				b.OnNavigate(NavigationRequest{URL: text, Method: "GET"})
			}
		}
	}

	// Create buttons
	goBtn := widget.NewButton("Go", func() {
		url := b.urlEntry.Text
		if url != "" {
			if b.OnNavigate != nil {
				b.OnNavigate(NavigationRequest{URL: url, Method: "GET"})
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

	objects := RenderToCanvas(commands, baseURL, false, b.triggerRepaint)
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
			b.OnNavigate(NavigationRequest{URL: prevURL, Method: "GET"})
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

	if b.onJSClick != nil && hit.Node != nil {
		// Run JS in goroutine so dialogs (confirm/prompt) don't block UI
		go b.onJSClick(hit.Node)
	}

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

	if hit.Type == layout.FileInputBox && hit.Node != nil {
		if isNodeDisabled(hit.Node) {
			return
		}
		fmt.Println("click file input")
		dialog.ShowFileOpen(func(reader fyne.URIReadCloser, err error) {
			if err != nil {
				fmt.Println("File dialog error:", err)
				return
			}
			if reader == nil {
				return // User cancelled
			}
			b.fileInputValues[hit.Node] = reader.URI().Path()
			reader.Close()
			b.repaint()
		}, b.Window)
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

	// Check if this is a form submission
	if hit.Node != nil && (hit.Node.TagName == "button" ||
		(hit.Node.TagName == "input" && hit.Node.Attributes["type"] == "submit")) {
		if isNodeDisabled(hit.Node) {
			return
		}
		buttonType := hit.Node.Attributes["type"]
		// Default button type inside form is "submit"
		if buttonType == "" || buttonType == "submit" {
			if formNode := findParentForm(hit.Node); formNode != nil {
				b.submitForm(formNode)
				return
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
		b.OnNavigate(NavigationRequest{URL: fullURL, Method: "GET"})
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
		InputValues:     b.inputValues,
		FocusedNode:     b.focusedInputNode,
		OpenSelectNode:  b.openSelectNode,
		RadioValues:     b.radioValues,
		CheckboxValues:  b.checkboxValue,
		FileInputValues: b.fileInputValues,
	})

	baseURL := ""
	if b.currentURL != nil {
		baseURL = b.currentURL.Scheme + "://" + b.currentURL.Host
	}

	// Use cached images on reflow (don't re-fetch)
	objects := RenderToCanvas(commands, baseURL, true, b.triggerRepaint) // true = use cache

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

func (b *Browser) ShowAlert(message string) {
	fyne.Do(func() {
		dialog.ShowInformation("Alert", message, b.Window)
	})
}

func (b *Browser) Refresh() {
	if b.currentURL != nil && b.OnNavigate != nil {
		b.OnNavigate(NavigationRequest{URL: b.currentURL.String(), Method: "GET"})
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
		InputValues:     b.inputValues,
		FocusedNode:     b.focusedInputNode,
		OpenSelectNode:  b.openSelectNode,
		RadioValues:     b.radioValues,
		CheckboxValues:  b.checkboxValue,
		FileInputValues: b.fileInputValues,
	})

	baseURL := ""
	if b.currentURL != nil {
		baseURL = b.currentURL.Scheme + "://" + b.currentURL.Host
	}

	objects := RenderToCanvas(commands, baseURL, true, nil)

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

// collectFormData gathers all input name/value pairs from a form
func (b *Browser) collectFormData(formNode *dom.Node) url.Values {
	data := url.Values{}
	b.collectInputs(formNode, data)
	return data
}

// collectInputs recursively collects inputs from the DOM tree
func (b *Browser) collectInputs(node *dom.Node, data url.Values) {
	if node == nil {
		return
	}

	if node.Type == dom.Element {
		name := node.Attributes["name"]
		if name != "" {
			switch node.TagName {
			case "input":
				inputType := node.Attributes["type"]
				switch inputType {
				case "checkbox", "radio":
					// Only include if checked
					if node.Attributes["checked"] != "" || b.isChecked(node) {
						value := node.Attributes["value"]
						if value == "" {
							value = "on" // Default value for checkboxes
						}
						data.Add(name, value)
					}
				case "file":
					// Skip file inputs - they require multipart encoding
				case "submit", "button":
					// Don't include submit buttons in data
				default:
					// text, password, email, number, hidden, etc.
					value := b.inputValues[node]
					if value == "" {
						value = node.Attributes["value"]
					}
					data.Add(name, value)
				}
			case "textarea":
				value := b.inputValues[node]
				data.Add(name, value)
			case "select":
				value := b.getSelectedValue(node)
				data.Add(name, value)
			}
		}
	}

	// Recurse into children
	for _, child := range node.Children {
		b.collectInputs(child, data)
	}
}

func (b *Browser) ShowConfirm(message string) bool {
	result := make(chan bool)

	fyne.Do(func() {
		dialog.ShowConfirm("Confirm", message, func(ok bool) {
			result <- ok
		}, b.Window)
	})

	return <-result
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

// findParentForm walks up the DOM tree to find the parent <form> element
func findParentForm(node *dom.Node) *dom.Node {
	current := node
	for current != nil {
		if current.Type == dom.Element && current.TagName == "form" {
			return current
		}
		current = current.Parent
	}
	return nil
}

// submitForm handles form submission
func (b *Browser) submitForm(formNode *dom.Node) {
	// Get form attributes
	action := formNode.Attributes["action"]
	method := strings.ToUpper(formNode.Attributes["method"])
	enctype := formNode.Attributes["enctype"]

	if method == "" {
		method = "GET" // Default method
	}

	// Collect form data
	data := b.collectFormData(formNode)

	// Build target URL
	var targetURL string
	if action == "" {
		// Submit to current page
		if b.currentURL != nil {
			targetURL = b.currentURL.String()
		}
	} else {
		// Resolve relative URL
		targetURL = b.resolveURL(action)
	}

	switch method {
	case "GET":
		// Append query string to URL
		if len(data) > 0 {
			if strings.Contains(targetURL, "?") {
				targetURL += "&" + data.Encode()
			} else {
				targetURL += "?" + data.Encode()
			}
		}

		// Navigate to the URL
		if b.OnNavigate != nil {
			b.OnNavigate(NavigationRequest{URL: targetURL, Method: "GET"})
		}
	case "POST":
		if enctype == "multipart/form-data" {
			body, contentType, err := b.buildMultipartBody(formNode)
			if err != nil {
				fmt.Println("Error building multipart body:", err)
				return
			}
			if b.OnNavigate != nil {
				b.OnNavigate(NavigationRequest{
					URL:         targetURL,
					Method:      "POST",
					Body:        body,
					ContentType: contentType,
				})
			}
		} else {
			data := b.collectFormData(formNode)
			if b.OnNavigate != nil {
				b.OnNavigate(NavigationRequest{
					URL:    targetURL,
					Method: "POST",
					Data:   data,
				})
			}
		}
	}
}

// isChecked checks if a checkbox/radio is currently checked
func (b *Browser) isChecked(node *dom.Node) bool {
	inputType := node.Attributes["type"]
	if inputType == "checkbox" {
		return b.checkboxValue[node]
	}
	if inputType == "radio" {
		name := node.Attributes["name"]
		return b.radioValues[name] == node
	}
	return false
}

// getSelectedValue gets the selected option value from a select element
func (b *Browser) getSelectedValue(selectNode *dom.Node) string {
	// Check if we have a stored value
	if val, ok := b.inputValues[selectNode]; ok {
		return val
	}

	// Otherwise check DOM for selected attribute
	for _, child := range selectNode.Children {
		if child.TagName == "option" {
			if child.Attributes["selected"] != "" {
				value := child.Attributes["value"]
				if value == "" {
					// Use text content if no value attribute
					for _, textNode := range child.Children {
						if textNode.Type == dom.Text {
							return textNode.Text
						}
					}
				}
				return value
			}
		}
	}

	// Return first option if none selected
	for _, child := range selectNode.Children {
		if child.TagName == "option" {
			value := child.Attributes["value"]
			if value == "" {
				for _, textNode := range child.Children {
					if textNode.Type == dom.Text {
						return textNode.Text
					}
				}
			}
			return value
		}
	}
	return ""
}

func (b *Browser) buildMultipartBody(formNode *dom.Node) ([]byte, string, error) {
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)

	b.addMultipartFields(writer, formNode)
	b.addmultipartFiles(writer, formNode)

	err := writer.Close()
	if err != nil {
		return nil, "", err
	}

	return body.Bytes(), writer.FormDataContentType(), nil
}

// addMultipartFields adds non-file form fields to the multipart writer
func (b *Browser) addMultipartFields(writer *multipart.Writer, node *dom.Node) {
	if node == nil {
		return
	}

	if node.Type == dom.Element {
		name := node.Attributes["name"]
		if name != "" {
			switch node.TagName {
			case "input":
				inputType := node.Attributes["type"]
				switch inputType {
				case "file":
					// Skip - handled separately
				case "checkbox", "radio":
					if b.isChecked(node) {
						value := node.Attributes["value"]
						if value == "" {
							value = "on"
						}
						field, _ := writer.CreateFormField(name)
						field.Write([]byte(value))
					}
				case "submit", "button":
					// Skip submit buttons
				default:
					value := b.inputValues[node]
					if value == "" {
						value = node.Attributes["value"]
					}
					field, _ := writer.CreateFormField(name)
					field.Write([]byte(value))
				}
			case "textarea":
				value := b.inputValues[node]
				field, _ := writer.CreateFormField(name)
				field.Write([]byte(value))
			case "select":
				value := b.getSelectedValue(node)
				field, _ := writer.CreateFormField(name)
				field.Write([]byte(value))
			}
		}
	}

	for _, child := range node.Children {
		b.addMultipartFields(writer, child)
	}
}

func (b *Browser) addmultipartFiles(writer *multipart.Writer, node *dom.Node) {
	if node == nil {
		return
	}

	if node.Type == dom.Element && node.TagName == "input" {
		inputType := node.Attributes["type"]
		name := node.Attributes["name"]

		if inputType == "file" && name != "" {
			filePath := b.fileInputValues[node]
			if filePath != "" {
				fileData, err := os.ReadFile(filePath)
				if err != nil {
					fmt.Println("Error reading file for upload:", err)
					return
				}

				filename := filepath.Base(filePath)

				fileWriter, err := writer.CreateFormFile(name, filename)
				if err != nil {
					fmt.Println("Error creating form file field:", err)
					return
				}

				fileWriter.Write(fileData)
			}
		}
	}

	for _, child := range node.Children {
		b.addmultipartFiles(writer, child)
	}
}

func (b *Browser) SetJSClickHandler(handler func(node *dom.Node)) {
	b.onJSClick = handler
}

func (b *Browser) triggerRepaint() {
	b.repaint()
}
