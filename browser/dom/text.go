package dom

import "strings"

var blockElements = map[string]bool{
	"html": true, "body": true, "div": true,
	"p": true, "h1": true, "h2": true, "h3": true,
	"h4": true, "h5": true, "h6": true,
	"ul": true, "ol": true, "li": true,
	"table": true, "tr": true,
	"header": true, "footer": true, "main": true,
	"section": true, "article": true, "nav": true,
	"blockquote": true, "pre": true, "hr": true,
}

var skipElements = map[string]bool{
	"script": true, "style": true, "head": true,
	"meta": true, "link": true, "noscript": true,
}

func (n *Node) ExtractText() string {
	var sb strings.Builder
	n.extractTextRecursive(&sb, false)
	return strings.TrimSpace(sb.String())
}

func (n *Node) extractTextRecursive(sb *strings.Builder, prevBlock bool) bool {
	if n.Type == Element && skipElements[n.TagName] {
		return prevBlock
	}

	isBlock := n.Type == Element && blockElements[n.TagName]

	if isBlock && !prevBlock {
		sb.WriteString("\n")
	}

	if n.Type == Text {
		sb.WriteString(n.Text)
		sb.WriteString(" ")
		return false
	}

	lastWasBlock := isBlock
	for _, child := range n.Children {
		lastWasBlock = child.extractTextRecursive(sb, lastWasBlock)
	}

	if isBlock {
		sb.WriteString("\n")
		return true
	}

	return lastWasBlock
}
