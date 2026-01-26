package layout

import (
	"browser/dom"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestContains(t *testing.T) {
	tests := []struct {
		name     string
		rect     Rect
		x, y     float64
		expected bool
	}{
		// Basic cases
		{"point clearly inside", Rect{10, 10, 100, 50}, 50, 30, true},
		{"point clearly outside left", Rect{10, 10, 100, 50}, 5, 30, false},
		{"point clearly outside right", Rect{10, 10, 100, 50}, 120, 30, false},
		{"point clearly outside above", Rect{10, 10, 100, 50}, 50, 5, false},
		{"point clearly outside below", Rect{10, 10, 100, 50}, 50, 70, false},

		// Boundary conditions (inclusive)
		{"point on left boundary", Rect{10, 10, 100, 50}, 10, 30, true},
		{"point on right boundary", Rect{10, 10, 100, 50}, 110, 30, true},
		{"point on top boundary", Rect{10, 10, 100, 50}, 50, 10, true},
		{"point on bottom boundary", Rect{10, 10, 100, 50}, 50, 60, true},

		// Corner cases
		{"point at top-left corner", Rect{10, 10, 100, 50}, 10, 10, true},
		{"point at top-right corner", Rect{10, 10, 100, 50}, 110, 10, true},
		{"point at bottom-left corner", Rect{10, 10, 100, 50}, 10, 60, true},
		{"point at bottom-right corner", Rect{10, 10, 100, 50}, 110, 60, true},

		// Just outside boundaries
		{"point just outside left", Rect{10, 10, 100, 50}, 9.99, 30, false},
		{"point just outside right", Rect{10, 10, 100, 50}, 110.01, 30, false},
		{"point just outside top", Rect{10, 10, 100, 50}, 50, 9.99, false},
		{"point just outside bottom", Rect{10, 10, 100, 50}, 50, 60.01, false},

		// Edge cases
		{"box at origin", Rect{0, 0, 100, 100}, 50, 50, true},
		{"point at origin in origin box", Rect{0, 0, 100, 100}, 0, 0, true},
		{"zero width box", Rect{10, 10, 0, 50}, 10, 30, true},
		{"zero height box", Rect{10, 10, 100, 0}, 50, 10, true},
		{"zero dimension box point inside", Rect{10, 10, 0, 0}, 10, 10, true},
		{"negative coordinates box", Rect{-50, -50, 100, 100}, -25, -25, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			box := &LayoutBox{Rect: tt.rect}
			result := box.Contains(tt.x, tt.y)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestHitTest(t *testing.T) {
	t.Run("returns nil when point outside box", func(t *testing.T) {
		box := createBoxWithRect(10, 10, 100, 50)
		result := box.HitTest(0, 0)
		assert.Nil(t, result)
	})

	t.Run("returns self when point inside box with no children", func(t *testing.T) {
		box := createBoxWithRect(10, 10, 100, 50)
		result := box.HitTest(50, 30)
		assert.Same(t, box, result)
	})

	t.Run("returns child when point inside child", func(t *testing.T) {
		parent := createBoxWithRect(0, 0, 200, 200)
		child := createBoxWithRect(50, 50, 100, 100)
		addChild(parent, child)

		result := parent.HitTest(75, 75)
		assert.Same(t, child, result)
	})

	t.Run("returns deepest nested child", func(t *testing.T) {
		// Create 3-level hierarchy
		grandparent := createBoxWithRect(0, 0, 300, 300)
		parent := createBoxWithRect(50, 50, 200, 200)
		child := createBoxWithRect(100, 100, 100, 100)

		addChild(grandparent, parent)
		addChild(parent, child)

		// Point inside all three
		result := grandparent.HitTest(150, 150)
		assert.Same(t, child, result)
	})

	t.Run("returns parent when point outside child but inside parent", func(t *testing.T) {
		parent := createBoxWithRect(0, 0, 200, 200)
		child := createBoxWithRect(100, 100, 50, 50)
		addChild(parent, child)

		// Point inside parent but outside child
		result := parent.HitTest(25, 25)
		assert.Same(t, parent, result)
	})

	t.Run("returns last child when siblings overlap (z-index simulation)", func(t *testing.T) {
		parent := createBoxWithRect(0, 0, 200, 200)
		child1 := createBoxWithRect(50, 50, 100, 100)
		child2 := createBoxWithRect(75, 75, 100, 100) // overlaps with child1
		addChild(parent, child1)
		addChild(parent, child2)

		// Point in overlap region - should return child2 (last added, rendered on top)
		result := parent.HitTest(100, 100)
		assert.Same(t, child2, result)
	})

	t.Run("returns first matching child when point only in first child", func(t *testing.T) {
		parent := createBoxWithRect(0, 0, 300, 100)
		child1 := createBoxWithRect(0, 0, 100, 100)
		child2 := createBoxWithRect(200, 0, 100, 100)
		addChild(parent, child1)
		addChild(parent, child2)

		// Point only in child1
		result := parent.HitTest(50, 50)
		assert.Same(t, child1, result)
	})

	t.Run("handles multiple levels of nesting correctly", func(t *testing.T) {
		// Create a 5-level deep hierarchy
		boxes := make([]*LayoutBox, 5)
		for i := 0; i < 5; i++ {
			offset := float64(i * 20)
			boxes[i] = createBoxWithRect(offset, offset, 200-offset*2, 200-offset*2)
			if i > 0 {
				addChild(boxes[i-1], boxes[i])
			}
		}

		// Point in the deepest box
		result := boxes[0].HitTest(100, 100)
		assert.Same(t, boxes[4], result)
	})

	t.Run("returns nil when box has no dimensions", func(t *testing.T) {
		box := createBoxWithRect(10, 10, 0, 0)
		// Point not exactly on the zero-dimension box
		result := box.HitTest(15, 15)
		assert.Nil(t, result)
	})
}

func TestFindLink(t *testing.T) {
	tests := []struct {
		name     string
		setup    func() *LayoutBox
		expected string
	}{
		{
			name: "direct link element with href",
			setup: func() *LayoutBox {
				return createLinkBox("https://example.com")
			},
			expected: "https://example.com",
		},
		{
			name: "parent is link element",
			setup: func() *LayoutBox {
				link := createLinkBox("https://parent.com")
				text := createTextBox("Click me")
				addChild(link, text)
				return text
			},
			expected: "https://parent.com",
		},
		{
			name: "grandparent is link element",
			setup: func() *LayoutBox {
				link := createLinkBox("https://grandparent.com")
				span := createElementBox(InlineBox, "span")
				text := createTextBox("Nested text")
				addChild(link, span)
				addChild(span, text)
				return text
			},
			expected: "https://grandparent.com",
		},
		{
			name: "no link ancestor returns empty string",
			setup: func() *LayoutBox {
				div := createElementBox(BlockBox, "div")
				text := createTextBox("No link here")
				addChild(div, text)
				return text
			},
			expected: "",
		},
		{
			name: "link element without href returns empty string",
			setup: func() *LayoutBox {
				// Create <a> without href attribute
				return &LayoutBox{
					Type: InlineBox,
					Node: createElementNode("a", nil),
				}
			},
			expected: "",
		},
		{
			name: "link with empty href returns empty string",
			setup: func() *LayoutBox {
				return createLinkBox("")
			},
			expected: "",
		},
		{
			name: "deep nesting finds ancestor link",
			setup: func() *LayoutBox {
				link := createLinkBox("https://deep.com")
				boxes := make([]*LayoutBox, 5)
				boxes[0] = link
				for i := 1; i < 5; i++ {
					boxes[i] = createElementBox(InlineBox, "span")
					addChild(boxes[i-1], boxes[i])
				}
				return boxes[4] // Return deepest box
			},
			expected: "https://deep.com",
		},
		{
			name: "box with nil node",
			setup: func() *LayoutBox {
				return &LayoutBox{Type: BlockBox, Node: nil}
			},
			expected: "",
		},
		{
			name: "relative URL preserved",
			setup: func() *LayoutBox {
				return createLinkBox("/about")
			},
			expected: "/about",
		},
		{
			name: "finds closest link when multiple exist",
			setup: func() *LayoutBox {
				outerLink := createLinkBox("https://outer.com")
				innerLink := createLinkBox("https://inner.com")
				text := createTextBox("Click")
				addChild(outerLink, innerLink)
				addChild(innerLink, text)
				return text
			},
			expected: "https://inner.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			box := tt.setup()
			result := box.FindLink()
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Helper to create a DOM element node
func createElementNode(tagName string, attrs map[string]string) *dom.Node {
	return dom.NewElement(tagName, attrs)
}
