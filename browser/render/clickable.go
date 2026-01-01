package render

import (
	"browser/layout"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"
)

// ClickableContainer is a container that responds to clicks
type ClickableContainer struct {
	widget.BaseWidget
	objects       []fyne.CanvasObject
	onTapped      func(x, y float32)
	layoutTree    *layout.LayoutBox
	currentCursor desktop.Cursor
}

// NewClickableContainer creates a clickable container
func NewClickableContainer(objects []fyne.CanvasObject, onTapped func(x, y float32), layoutTree *layout.LayoutBox) *ClickableContainer {
	c := &ClickableContainer{
		objects:       objects,
		onTapped:      onTapped,
		layoutTree:    layoutTree,
		currentCursor: desktop.DefaultCursor,
	}
	c.ExtendBaseWidget(c)
	return c
}

// Tapped is called when user clicks
func (c *ClickableContainer) Tapped(event *fyne.PointEvent) {
	if c.onTapped != nil {
		c.onTapped(event.Position.X, event.Position.Y)
	}
}

// TappedSecondary is called for right-click (we ignore it)
func (c *ClickableContainer) TappedSecondary(event *fyne.PointEvent) {}

func (c *ClickableContainer) MouseIn(event *desktop.MouseEvent) {}

func (c *ClickableContainer) MouseOut() {}

func (c *ClickableContainer) MouseMoved(event *desktop.MouseEvent) {
	if c.layoutTree == nil {
		return
	}

	hit := c.layoutTree.HitTest(float64(event.Position.X), float64(event.Position.Y))

	cursor := desktop.DefaultCursor
	for box := hit; box != nil; box = box.Parent {
		if box.Style.Cursor != "" {
			switch box.Style.Cursor {
			case "pointer":
				cursor = desktop.PointerCursor
			case "text":
				cursor = desktop.TextCursor
			case "crosshair":
				cursor = desktop.CrosshairCursor
			}
			break
		}
		if box.Node != nil && box.Node.TagName == "a" {
			cursor = desktop.PointerCursor
			break
		}
	}

	if c.currentCursor != cursor {
		c.currentCursor = cursor
		c.Refresh()
	}
}

// Cursor implements desktop.Cursorable
func (c *ClickableContainer) Cursor() desktop.Cursor {
	return c.currentCursor
}

// CreateRenderer returns the renderer for this widget
func (c *ClickableContainer) CreateRenderer() fyne.WidgetRenderer {
	return &clickableRenderer{
		container: c,
		objects:   c.objects,
	}
}

// clickableRenderer handles drawing
type clickableRenderer struct {
	container *ClickableContainer
	objects   []fyne.CanvasObject
}

func (r *clickableRenderer) Layout(size fyne.Size) {
	// Objects are already positioned, no need to layout
}

func (r *clickableRenderer) MinSize() fyne.Size {
	// Calculate actual content bounds
	var maxX, maxY float32
	for _, obj := range r.objects {
		pos := obj.Position()
		size := obj.Size()
		right := pos.X + size.Width
		bottom := pos.Y + size.Height
		if right > maxX {
			maxX = right
		}
		if bottom > maxY {
			maxY = bottom
		}
	}
	return fyne.NewSize(maxX, maxY)
}

func (r *clickableRenderer) Refresh() {
	for _, obj := range r.objects {
		obj.Refresh()
	}
}

func (r *clickableRenderer) Objects() []fyne.CanvasObject {
	return r.objects
}

func (r *clickableRenderer) Destroy() {}
