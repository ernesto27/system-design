package layout

import (
	"browser/css"
	"browser/dom"
)

var blockElements = map[string]bool{
	"html": true, "body": true, "div": true,
	"p": true, "h1": true, "h2": true, "h3": true, "h4": true, "h5": true, "h6": true,
	"ul": true, "ol": true, "li": true,
	"header": true, "footer": true,
	"blockquote": true,
}

var skipElements = map[string]bool{
	"script": true, "style": true, "head": true,
	"meta": true, "link": true,
}

var imageElements = map[string]bool{
	"img": true,
}

func BuildLayoutTree(root *dom.Node) *LayoutBox {
	return BuildBox(root, nil)
}

func BuildBox(node *dom.Node, parent *LayoutBox) *LayoutBox {
	if node.Type == dom.Element && skipElements[node.TagName] {
		return nil
	}

	box := &LayoutBox{Node: node, Parent: parent}

	if node.Type == dom.Element {
		if styleAttr, ok := node.Attributes["style"]; ok {
			box.Style = css.ParseInlineStyle(styleAttr)
		} else {
			box.Style = css.DefaultStyle()
		}
	} else {

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
		childBox := BuildBox(child, box)
		if childBox != nil {
			box.Children = append(box.Children, childBox)
		}
	}

	return box
}
