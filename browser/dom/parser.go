package dom

import (
	"io"
	"strings"

	"golang.org/x/net/html"
)

func Parse(r io.Reader) *Node {
	doc, err := html.Parse(r)
	if err != nil {
		return nil
	}

	return convertNode(doc)
}

func convertNode(n *html.Node) *Node {
	return convertNodeWithContext(n, false)
}

func convertNodeWithContext(n *html.Node, preserveWhitespace bool) *Node {
	var node *Node

	// Check if this element preserves whitespace
	isPreserving := preserveWhitespace
	if n.Type == html.ElementNode && n.Data == "pre" {
		isPreserving = true
	}

	switch n.Type {
	case html.DocumentNode:
		node = &Node{Type: Document, Children: []*Node{}}
	case html.ElementNode:
		attrs := make(map[string]string)
		for _, attr := range n.Attr {
			attrs[attr.Key] = attr.Val
		}
		node = NewElement(n.Data, attrs)
	case html.TextNode:
		var text string
		if preserveWhitespace {
			// Keep all whitespace for <pre> content
			text = n.Data
		} else {
			text = strings.TrimSpace(n.Data)
		}
		if text == "" {
			return nil
		}
		node = NewText(text)
	default:
		return nil
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		child := convertNodeWithContext(c, isPreserving)
		if child != nil {
			node.AppendChild(child)
		}
	}

	return node
}

