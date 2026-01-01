package render

import (
	"browser/css"
	"browser/dom"
	"browser/layout"
	"fmt"
	"image/color"
	"net/url"
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
		App:        a,
		Window:     w,
		Width:      width,
		Height:     height,
		history:    []string{},
		historyPos: -1,
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

	w.Canvas().SetOnTypedKey(nil) // Ensure canvas is initialized
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
		return
	}
	fmt.Printf("  Hit: %+v\n", hit.Text)

	// Check if it's a link
	href := hit.FindLink()
	if href == "" {
		fmt.Println("  Not a link")
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

	// Repaint
	commands := BuildDisplayList(layoutTree)

	baseURL := ""
	if b.currentURL != nil {
		baseURL = b.currentURL.Scheme + "://" + b.currentURL.Host
	}

	// Use cached images on reflow (don't re-fetch)
	objects := RenderToCanvas(commands, baseURL, true) // true = use cache

	// UI updates must be on main thread
	fyne.Do(func() {
		clickable := NewClickableContainer(objects, func(x, y float32) {
			b.handleClick(float64(x), float64(y))
		}, b.layoutTree)

		scroll := container.NewScroll(clickable)

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
