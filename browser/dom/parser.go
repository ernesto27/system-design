package dom

import (
	"io"
	"strings"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
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
	if n.Type == html.ElementNode && (n.Data == "pre" || n.Data == "script" || n.Data == "style") {
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
			text = normalizeWhitespace(n.Data)
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

// normalizeWhitespace collapses whitespace sequences to single spaces
// while preserving space boundaries for inline element separation
func normalizeWhitespace(s string) string {
	if len(s) == 0 {
		return ""
	}

	// Check boundaries - only actual spaces (not newlines) indicate inline separation
	startsWithSpace := s[0] == ' ' || s[0] == '\t'
	endsWithSpace := s[len(s)-1] == ' ' || s[len(s)-1] == '\t'

	// Collapse all whitespace to single spaces
	words := strings.Fields(s)
	if len(words) == 0 {
		// Text was all whitespace
		// Only keep if it's purely spaces/tabs (inline separator)
		// Discard if it contains newlines (HTML formatting)
		if strings.ContainsAny(s, "\n\r") {
			return ""
		}
		return " "
	}

	result := strings.Join(words, " ")

	// Preserve boundary spaces for inline element separation
	if startsWithSpace {
		result = " " + result
	}
	if endsWithSpace {
		result = result + " "
	}

	return result
}

// ParseFragment parses an HTML fragment (not a full document)
// Returns a slice of nodes that were parsed
func ParseFragment(htmlContent string) []*Node {
	context := &html.Node{
		Type:     html.ElementNode,
		Data:     "div",
		DataAtom: atom.Div,
	}

	nodes, err := html.ParseFragment(strings.NewReader(htmlContent), context)
	if err != nil {
		return nil
	}

	var result []*Node
	for _, n := range nodes {
		converted := convertNode(n)
		if converted != nil {
			result = append(result, converted)
		}
	}
	return result
}

