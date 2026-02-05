package layout

import (
	"browser/css"
	"browser/dom"
	"strings"
)

// parseHTML parses an HTML string into a DOM tree
func parseHTML(html string) *dom.Node {
	return dom.Parse(strings.NewReader(html))
}

// emptyStylesheet returns an empty CSS stylesheet
func emptyStylesheet() css.Stylesheet {
	return css.Stylesheet{}
}

// createStylesheet parses CSS text into a Stylesheet
func createStylesheet(cssText string) css.Stylesheet {
	return css.Parse(cssText)
}

// buildTree is a convenience wrapper for building layout from HTML
func buildTree(html string) *LayoutBox {
	root := parseHTML(html)
	return BuildLayoutTree(root, emptyStylesheet(), Viewport{})
}

// buildTreeWithCSS builds tree with stylesheet
func buildTreeWithCSS(html, cssText string) *LayoutBox {
	root := parseHTML(html)
	sheet := createStylesheet(cssText)
	return BuildLayoutTree(root, sheet, Viewport{})
}

// findBoxByTag finds first box with given tag using DFS
func findBoxByTag(root *LayoutBox, tag string) *LayoutBox {
	if root == nil {
		return nil
	}
	if root.Node != nil && root.Node.TagName == tag {
		return root
	}
	for _, child := range root.Children {
		if found := findBoxByTag(child, tag); found != nil {
			return found
		}
	}
	return nil
}

// findBoxByType finds first box with given type using DFS
func findBoxByType(root *LayoutBox, boxType BoxType) *LayoutBox {
	if root == nil {
		return nil
	}
	if root.Type == boxType {
		return root
	}
	for _, child := range root.Children {
		if found := findBoxByType(child, boxType); found != nil {
			return found
		}
	}
	return nil
}

// countBoxes counts total number of boxes in tree
func countBoxes(root *LayoutBox) int {
	if root == nil {
		return 0
	}
	count := 1
	for _, child := range root.Children {
		count += countBoxes(child)
	}
	return count
}

// createBoxWithRect creates a LayoutBox with specific dimensions
func createBoxWithRect(x, y, w, h float64) *LayoutBox {
	return &LayoutBox{
		Rect: Rect{X: x, Y: y, Width: w, Height: h},
	}
}

// createTextBox creates a TextBox with text content
func createTextBox(text string) *LayoutBox {
	return &LayoutBox{
		Type: TextBox,
		Text: text,
		Node: dom.NewText(text),
	}
}

// createElementBox creates a LayoutBox for an element with given tag
func createElementBox(boxType BoxType, tagName string) *LayoutBox {
	return &LayoutBox{
		Type: boxType,
		Node: dom.NewElement(tagName, nil),
	}
}

// createLinkBox creates a LayoutBox with an <a> element
func createLinkBox(href string) *LayoutBox {
	attrs := map[string]string{"href": href}
	return &LayoutBox{
		Type: InlineBox,
		Node: dom.NewElement("a", attrs),
	}
}

// addChild adds a child to parent and sets parent pointer
func addChild(parent, child *LayoutBox) {
	child.Parent = parent
	parent.Children = append(parent.Children, child)
}
