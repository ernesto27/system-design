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
	var node *Node

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
		text := strings.TrimSpace(n.Data)
		if text == "" {
			return nil
		}
		node = NewText(text)
	default:
		return nil
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		child := convertNode(c)
		if child != nil {
			node.AppendChild(child)
		}
	}

	return node
}
