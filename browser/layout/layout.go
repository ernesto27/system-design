package layout

import (
	"browser/css"
	"browser/dom"
	"strings"
)

type Viewport struct {
	Width  float64
	Height float64
}

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
	"form":       true,
	"textarea":   true,
	"dl":         true,
	"dt":         true,
	"dd":         true,
	"fieldset":   true,
	"center":     true,
}

var skipElements = map[string]bool{
	"script": true, "style": true, "head": true,
	"meta": true, "link": true, "option": true,
}

var imageElements = map[string]bool{
	"img": true,
}

func BuildLayoutTree(root *dom.Node, stylesheet css.Stylesheet, viewport Viewport) *LayoutBox {
	return BuildBox(root, nil, stylesheet, viewport)
}

func BuildBox(node *dom.Node, parent *LayoutBox, stylesheet css.Stylesheet, viewport Viewport) *LayoutBox {
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

		// Get parent's font-size for em unit resolution
		parentFontSize := 16.0 // Default browser font-size
		if parent != nil && parent.Style.FontSize > 0 {
			parentFontSize = parent.Style.FontSize
		}

		box.Style = css.ApplyStylesheetWithContext(stylesheet, node.TagName, id, classes, parentFontSize, viewport.Width, viewport.Height)

		if align, ok := node.Attributes["align"]; ok {
			switch strings.ToLower(align) {
			case "left":
				box.Style.TextAlign = "left"
			case "right":
				box.Style.TextAlign = "right"
			case "center":
				box.Style.TextAlign = "center"
			case "justify":
				box.Style.TextAlign = "justify"
			}
		}

		// Then apply inline styles (override stylesheet)
		if styleAttr, ok := node.Attributes["style"]; ok {
			inlineStyle := css.ParseInlineStyleWithContext(styleAttr, parentFontSize, viewport.Width, viewport.Height)
			mergeStyles(&box.Style, &inlineStyle)
		}

		if box.Style.Display == "none" {
			return nil
		}

		box.Position = box.Style.Position
		box.Top = box.Style.Top
		box.Left = box.Style.Left
		box.Right = box.Style.Right
		box.Bottom = box.Style.Bottom
		box.Float = box.Style.Float
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
		} else if node.TagName == dom.TagInput {
			inputType := node.Attributes["type"]
			switch strings.ToLower(inputType) {
			case "hidden":
				return nil
			case "radio":
				box.Type = RadioBox
			case "checkbox":
				box.Type = CheckboxBox
			case "file":
				box.Type = FileInputBox
			default:
				box.Type = InputBox
			}
		} else if node.TagName == dom.TagButton {
			box.Type = ButtonBox
		} else if node.TagName == dom.TagTextarea {
			box.Type = TextareaBox
		} else if node.TagName == dom.TagSelect {
			box.Type = SelectBox
		} else if node.TagName == dom.TagFieldSet {
			box.Type = FieldsetBox
		} else if node.TagName == dom.TagLegend {
			box.Type = LegendBox
		} else if blockElements[node.TagName] {
			box.Type = BlockBox
		} else if node.TagName == dom.TagTable || node.TagName == dom.TagTBody || node.TagName == dom.TagTHead || node.TagName == dom.TagTFoot {
			box.Type = TableBox
		} else if node.TagName == dom.TagTR {
			box.Type = TableRowBox
		} else if node.TagName == dom.TagTD || node.TagName == dom.TagTH {
			box.Type = TableCellBox
		} else if node.TagName == dom.TagCaption {
			box.Type = TableCaptionBox
		} else {
			box.Type = InlineBox
		}
	case dom.Text:
		box.Type = TextBox
		box.Text = wrapInlineQuotes(node)
	}

	for _, child := range node.Children {
		childBox := BuildBox(child, box, stylesheet, viewport)
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
	if inline.LineHeight > 0 {
		base.LineHeight = inline.LineHeight
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
	if inline.MarginLeft > 0 || inline.MarginLeftAuto {
		base.MarginLeft = inline.MarginLeft
		base.MarginLeftAuto = inline.MarginLeftAuto
	}
	if inline.MarginRight > 0 || inline.MarginRightAuto {
		base.MarginRight = inline.MarginRight
		base.MarginRightAuto = inline.MarginRightAuto
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
	if inline.TextTransform != "" {
		base.TextTransform = inline.TextTransform
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
	// Sizing properties
	if inline.Width > 0 {
		base.Width = inline.Width
	}

	if inline.MinWidth > 0 {
		base.MinWidth = inline.MinWidth
	}

	if inline.MaxWidth > 0 {
		base.MaxWidth = inline.MaxWidth
	}

	if inline.Height > 0 {
		base.Height = inline.Height
	}

	if inline.MinHeight > 0 {
		base.MinHeight = inline.MinHeight
	}

	if inline.MaxHeight > 0 {
		base.MaxHeight = inline.MaxHeight
	}

	// Position properties
	if inline.Position != "" {
		base.Position = inline.Position
	}
	if inline.Top > 0 {
		base.Top = inline.Top
	}
	if inline.Left > 0 {
		base.Left = inline.Left
	}
	if inline.Right > 0 {
		base.Right = inline.Right
	}
	if inline.Bottom > 0 {
		base.Bottom = inline.Bottom
	}

	if inline.Float != "" {
		base.Float = inline.Float
	}

	if len(inline.FontFamily) > 0 {
		base.FontFamily = inline.FontFamily
	}

	if inline.BorderRadius > 0 {
		base.BorderRadius = inline.BorderRadius
	}

	if inline.BackgroundImage != "" {
		base.BackgroundImage = inline.BackgroundImage
	}
}

// wrapInlineQuotes adds quotation marks for <q> elements
func wrapInlineQuotes(node *dom.Node) string {
	text := node.Text
	if node.Parent != nil && node.Parent.TagName == dom.TagQ {
		text = "\u201C" + text + "\u201D"
	}
	return text
}
