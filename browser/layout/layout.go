package layout

import (
	"browser/css"
	"browser/dom"
	"strings"
)

var blockElements = map[string]bool{
	"html":       true,
	"body":       true,
	"div":        true,
	"p":          true,
	"h1":         true,
	"h2":         true,
	"h3":         true,
	"h4":         true,
	"h5":         true,
	"h6":         true,
	"ul":         true,
	"ol":         true,
	"li":         true,
	"header":     true,
	"footer":     true,
	"blockquote": true,
	"pre":        true,
}

var skipElements = map[string]bool{
	"script": true, "style": true, "head": true,
	"meta": true, "link": true,
}

var imageElements = map[string]bool{
	"img": true,
}

func BuildLayoutTree(root *dom.Node, stylesheet css.Stylesheet) *LayoutBox {
	return BuildBox(root, nil, stylesheet)
}

func BuildBox(node *dom.Node, parent *LayoutBox, stylesheet css.Stylesheet) *LayoutBox {
	if node.Type == dom.Element && skipElements[node.TagName] {
		return nil
	}

	box := &LayoutBox{Node: node, Parent: parent}

	if node.Type == dom.Element {
		id := node.Attributes["id"]
		classAttr := node.Attributes["class"]
		var classes []string
		if classAttr != "" {
			classes = strings.Fields(classAttr)
		}

		box.Style = css.ApplyStylesheet(stylesheet, node.TagName, id, classes)

		// Then apply inline styles (override stylesheet)
		if styleAttr, ok := node.Attributes["style"]; ok {
			inlineStyle := css.ParseInlineStyle(styleAttr)
			mergeStyles(&box.Style, &inlineStyle)
		}

		if box.Style.Display == "none" {
			return nil
		}

	}

	switch node.Type {
	case dom.Document:
		box.Type = BlockBox
	case dom.Element:
		if node.TagName == dom.TagHR {
			box.Type = HRBox
		} else if node.TagName == dom.TagBR {
			box.Type = BRBox
		} else if imageElements[node.TagName] {
			box.Type = ImageBox
		} else if blockElements[node.TagName] {
			box.Type = BlockBox
		} else if node.TagName == dom.TagTable || node.TagName == dom.TagTBody || node.TagName == dom.TagTHead || node.TagName == dom.TagTFoot {
			box.Type = TableBox
		} else if node.TagName == dom.TagTR {
			box.Type = TableRowBox
		} else if node.TagName == dom.TagTD || node.TagName == dom.TagTH {
			box.Type = TableCellBox
		} else {
			box.Type = InlineBox
		}
	case dom.Text:
		box.Type = TextBox
		box.Text = node.Text
	}

	for _, child := range node.Children {
		childBox := BuildBox(child, box, stylesheet)
		if childBox != nil {
			box.Children = append(box.Children, childBox)
		}
	}

	return box
}

// mergeStyles merges inline style into base style (inline wins)
func mergeStyles(base *css.Style, inline *css.Style) {
	if inline.Color != nil {
		base.Color = inline.Color
	}
	if inline.BackgroundColor != nil {
		base.BackgroundColor = inline.BackgroundColor
	}
	if inline.FontSize > 0 {
		base.FontSize = inline.FontSize
	}
	if inline.Bold {
		base.Bold = true
	}
	if inline.Italic {
		base.Italic = true
	}
	if inline.MarginTop > 0 {
		base.MarginTop = inline.MarginTop
	}
	if inline.MarginBottom > 0 {
		base.MarginBottom = inline.MarginBottom
	}
	if inline.MarginLeft > 0 {
		base.MarginLeft = inline.MarginLeft
	}
	if inline.MarginRight > 0 {
		base.MarginRight = inline.MarginRight
	}
	if inline.PaddingTop > 0 {
		base.PaddingTop = inline.PaddingTop
	}
	if inline.PaddingBottom > 0 {
		base.PaddingBottom = inline.PaddingBottom
	}
	if inline.PaddingLeft > 0 {
		base.PaddingLeft = inline.PaddingLeft
	}
	if inline.PaddingRight > 0 {
		base.PaddingRight = inline.PaddingRight
	}
	if inline.TextAlign != "" {
		base.TextAlign = inline.TextAlign
	}
	if inline.Display != "" {
		base.Display = inline.Display
	}
	if inline.TextDecoration != "" {
		base.TextDecoration = inline.TextDecoration
	}
	if inline.Opacity != 1.0 {
		base.Opacity = inline.Opacity
	}
	if inline.Visibility != "" {
		base.Visibility = inline.Visibility
	}
	if inline.Cursor != "" {
		base.Cursor = inline.Cursor
	}
	// Border properties
	if inline.BorderTopWidth > 0 {
		base.BorderTopWidth = inline.BorderTopWidth
	}
	if inline.BorderRightWidth > 0 {
		base.BorderRightWidth = inline.BorderRightWidth
	}
	if inline.BorderBottomWidth > 0 {
		base.BorderBottomWidth = inline.BorderBottomWidth
	}
	if inline.BorderLeftWidth > 0 {
		base.BorderLeftWidth = inline.BorderLeftWidth
	}
	if inline.BorderTopColor != nil {
		base.BorderTopColor = inline.BorderTopColor
	}
	if inline.BorderRightColor != nil {
		base.BorderRightColor = inline.BorderRightColor
	}
	if inline.BorderBottomColor != nil {
		base.BorderBottomColor = inline.BorderBottomColor
	}
	if inline.BorderLeftColor != nil {
		base.BorderLeftColor = inline.BorderLeftColor
	}
	if inline.BorderTopStyle != "" {
		base.BorderTopStyle = inline.BorderTopStyle
	}
	if inline.BorderRightStyle != "" {
		base.BorderRightStyle = inline.BorderRightStyle
	}
	if inline.BorderBottomStyle != "" {
		base.BorderBottomStyle = inline.BorderBottomStyle
	}
	if inline.BorderLeftStyle != "" {
		base.BorderLeftStyle = inline.BorderLeftStyle
	}
}
