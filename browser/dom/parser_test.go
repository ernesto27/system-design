package dom

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// mapsEqual compares two string maps for equality
func mapsEqual(m1, m2 map[string]string) bool {
	if len(m1) != len(m2) {
		return false
	}
	for k, v := range m1 {
		if m2[k] != v {
			return false
		}
	}
	return true
}

// nodeEqual recursively compares two DOM trees for equality
func nodeEqual(n1, n2 *Node) bool {
	if n1 == nil && n2 == nil {
		return true
	}
	if n1 == nil || n2 == nil {
		return false
	}
	if n1.Type != n2.Type {
		return false
	}
	if n1.TagName != n2.TagName {
		return false
	}
	if n1.Text != n2.Text {
		return false
	}
	if !mapsEqual(n1.Attributes, n2.Attributes) {
		return false
	}
	if len(n1.Children) != len(n2.Children) {
		return false
	}
	for i := range n1.Children {
		if !nodeEqual(n1.Children[i], n2.Children[i]) {
			return false
		}
	}
	return true
}

// nodeDiff returns a detailed description of the first difference found between two trees
func nodeDiff(n1, n2 *Node, path string) string {
	if path == "" {
		path = "root"
	}

	if n1 == nil && n2 == nil {
		return ""
	}
	if n1 == nil {
		return fmt.Sprintf("%s: got nil, want node (Type=%v, TagName=%q)", path, n2.Type, n2.TagName)
	}
	if n2 == nil {
		return fmt.Sprintf("%s: got node (Type=%v, TagName=%q), want nil", path, n1.Type, n1.TagName)
	}
	if n1.Type != n2.Type {
		return fmt.Sprintf("%s: Type = %v, want %v", path, n1.Type, n2.Type)
	}
	if n1.TagName != n2.TagName {
		return fmt.Sprintf("%s: TagName = %q, want %q", path, n1.TagName, n2.TagName)
	}
	if n1.Text != n2.Text {
		return fmt.Sprintf("%s: Text = %q, want %q", path, n1.Text, n2.Text)
	}
	if !mapsEqual(n1.Attributes, n2.Attributes) {
		return fmt.Sprintf("%s: Attributes = %v, want %v", path, n1.Attributes, n2.Attributes)
	}
	if len(n1.Children) != len(n2.Children) {
		return fmt.Sprintf("%s: len(Children) = %d, want %d", path, len(n1.Children), len(n2.Children))
	}
	for i := range n1.Children {
		childPath := fmt.Sprintf("%s.Children[%d]", path, i)
		if n1.Children[i].TagName != "" {
			childPath = fmt.Sprintf("%s/%s", path, n1.Children[i].TagName)
		} else if n1.Children[i].Type == Text {
			childPath = fmt.Sprintf("%s/#text", path)
		}
		if diff := nodeDiff(n1.Children[i], n2.Children[i], childPath); diff != "" {
			return diff
		}
	}
	return ""
}

// printTree returns a string representation of a DOM tree for debugging
func printTree(n *Node, indent string) string {
	if n == nil {
		return indent + "<nil>\n"
	}

	var sb strings.Builder
	switch n.Type {
	case Document:
		sb.WriteString(indent + "[Document]\n")
	case Element:
		sb.WriteString(fmt.Sprintf("%s<%s", indent, n.TagName))
		for k, v := range n.Attributes {
			sb.WriteString(fmt.Sprintf(" %s=%q", k, v))
		}
		sb.WriteString(">\n")
	case Text:
		sb.WriteString(fmt.Sprintf("%s#text: %q\n", indent, n.Text))
	}

	for _, child := range n.Children {
		sb.WriteString(printTree(child, indent+"  "))
	}

	return sb.String()
}

func TestNormalizeWhitespace(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		// Empty input
		{"empty string", "", ""},

		// All whitespace with newlines (HTML formatting between blocks)
		{"newlines only", "\n\n", ""},
		{"newlines with spaces", "\n    \n", ""},
		{"tabs and newlines", "\t\n\t", ""},

		// All whitespace without newlines (inline separator)
		{"single space", " ", " "},
		{"multiple spaces", "   ", " "},
		{"tabs only", "\t\t", " "},

		// Text with leading space (inline element after text)
		{"leading space", " hello", " hello"},
		{"leading tab", "\thello", " hello"},

		// Text with trailing space (text before inline element)
		{"trailing space", "hello ", "hello "},
		{"trailing tab", "hello\t", "hello "},

		// Text with both leading and trailing space
		{"both spaces", " hello ", " hello "},

		// Text with internal whitespace (should collapse)
		{"internal spaces", "hello    world", "hello world"},
		{"internal newline", "hello\nworld", "hello world"},
		{"internal mixed", "hello  \n  world", "hello world"},

		// Text starting/ending with newline (no boundary space)
		{"starts with newline", "\nhello", "hello"},
		{"ends with newline", "hello\n", "hello"},
		{"both newlines", "\nhello\n", "hello"},

		// Real-world HTML scenarios
		{"before inline tag", "This is ", "This is "},
		{"after inline tag", " using strong tag.", " using strong tag."},
		{"between block tags", "\n    \n    ", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := normalizeWhitespace(tt.input)
			assert.Equal(t, tt.expected, result, "normalizeWhitespace(%q)", tt.input)
		})
	}
}

func TestParse(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected *Node
	}{
		{
			name:  "simple paragraph",
			input: "<p>Hello</p>",
			expected: func() *Node {
				doc := &Node{Type: Document, Children: []*Node{}}
				html := NewElement("html", nil)
				head := NewElement("head", nil)
				body := NewElement("body", nil)
				p := NewElement("p", nil)
				text := NewText("Hello")

				p.AppendChild(text)
				body.AppendChild(p)
				html.AppendChild(head)
				html.AppendChild(body)
				doc.AppendChild(html)
				return doc
			}(),
		},
		{
			name:  "nested elements",
			input: "<div><p>Text</p></div>",
			expected: func() *Node {
				doc := &Node{Type: Document, Children: []*Node{}}
				html := NewElement("html", nil)
				head := NewElement("head", nil)
				body := NewElement("body", nil)
				div := NewElement("div", nil)
				p := NewElement("p", nil)
				text := NewText("Text")

				p.AppendChild(text)
				div.AppendChild(p)
				body.AppendChild(div)
				html.AppendChild(head)
				html.AppendChild(body)
				doc.AppendChild(html)
				return doc
			}(),
		},
		{
			name:  "element with attributes",
			input: `<a href="https://example.com" target="_blank">Link</a>`,
			expected: func() *Node {
				doc := &Node{Type: Document, Children: []*Node{}}
				html := NewElement("html", nil)
				head := NewElement("head", nil)
				body := NewElement("body", nil)
				a := NewElement("a", map[string]string{
					"href":   "https://example.com",
					"target": "_blank",
				})
				text := NewText("Link")

				a.AppendChild(text)
				body.AppendChild(a)
				html.AppendChild(head)
				html.AppendChild(body)
				doc.AppendChild(html)
				return doc
			}(),
		},
		{
			name:  "multiple siblings",
			input: "<ul><li>One</li><li>Two</li><li>Three</li></ul>",
			expected: func() *Node {
				doc := &Node{Type: Document, Children: []*Node{}}
				html := NewElement("html", nil)
				head := NewElement("head", nil)
				body := NewElement("body", nil)
				ul := NewElement("ul", nil)
				li1 := NewElement("li", nil)
				li2 := NewElement("li", nil)
				li3 := NewElement("li", nil)
				text1 := NewText("One")
				text2 := NewText("Two")
				text3 := NewText("Three")

				li1.AppendChild(text1)
				li2.AppendChild(text2)
				li3.AppendChild(text3)
				ul.AppendChild(li1)
				ul.AppendChild(li2)
				ul.AppendChild(li3)
				body.AppendChild(ul)
				html.AppendChild(head)
				html.AppendChild(body)
				doc.AppendChild(html)
				return doc
			}(),
		},
		{
			name:  "text with whitespace normalization",
			input: "<p>  hello   world  </p>",
			expected: func() *Node {
				doc := &Node{Type: Document, Children: []*Node{}}
				html := NewElement("html", nil)
				head := NewElement("head", nil)
				body := NewElement("body", nil)
				p := NewElement("p", nil)
				// normalizeWhitespace preserves boundary spaces for inline element separation
				text := NewText(" hello world ")

				p.AppendChild(text)
				body.AppendChild(p)
				html.AppendChild(head)
				html.AppendChild(body)
				doc.AppendChild(html)
				return doc
			}(),
		},
		{
			name:  "full html document",
			input: "<html><head><title>Test</title></head><body><h1>Hello</h1></body></html>",
			expected: func() *Node {
				doc := &Node{Type: Document, Children: []*Node{}}
				html := NewElement("html", nil)
				head := NewElement("head", nil)
				title := NewElement("title", nil)
				titleText := NewText("Test")
				body := NewElement("body", nil)
				h1 := NewElement("h1", nil)
				h1Text := NewText("Hello")

				title.AppendChild(titleText)
				head.AppendChild(title)
				h1.AppendChild(h1Text)
				body.AppendChild(h1)
				html.AppendChild(head)
				html.AppendChild(body)
				doc.AppendChild(html)
				return doc
			}(),
		},
		{
			name:  "mixed inline and text",
			input: "<p>Hello <strong>World</strong>!</p>",
			expected: func() *Node {
				doc := &Node{Type: Document, Children: []*Node{}}
				html := NewElement("html", nil)
				head := NewElement("head", nil)
				body := NewElement("body", nil)
				p := NewElement("p", nil)
				text1 := NewText("Hello ")
				strong := NewElement("strong", nil)
				strongText := NewText("World")
				text2 := NewText("!")

				strong.AppendChild(strongText)
				p.AppendChild(text1)
				p.AppendChild(strong)
				p.AppendChild(text2)
				body.AppendChild(p)
				html.AppendChild(head)
				html.AppendChild(body)
				doc.AppendChild(html)
				return doc
			}(),
		},
		{
			name:  "element with class attribute",
			input: `<div class="container"><span class="text">Content</span></div>`,
			expected: func() *Node {
				doc := &Node{Type: Document, Children: []*Node{}}
				html := NewElement("html", nil)
				head := NewElement("head", nil)
				body := NewElement("body", nil)
				div := NewElement("div", map[string]string{"class": "container"})
				span := NewElement("span", map[string]string{"class": "text"})
				text := NewText("Content")

				span.AppendChild(text)
				div.AppendChild(span)
				body.AppendChild(div)
				html.AppendChild(head)
				html.AppendChild(body)
				doc.AppendChild(html)
				return doc
			}(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Parse(strings.NewReader(tt.input))

			if !nodeEqual(result, tt.expected) {
				diff := nodeDiff(result, tt.expected, "")
				assert.Fail(t, "Parse() tree mismatch",
					"%s\n\nGot:\n%s\nWant:\n%s",
					diff, printTree(result, ""), printTree(tt.expected, ""))
			}
		})
	}
}

func TestParseFragment(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []*Node
	}{
		{
			name:  "single text node",
			input: "Hello World",
			expected: []*Node{
				NewText("Hello World"),
			},
		},
		{
			name:  "single element",
			input: "<p>Hello</p>",
			expected: func() []*Node {
				p := NewElement("p", nil)
				p.AppendChild(NewText("Hello"))
				return []*Node{p}
			}(),
		},
		{
			name:  "multiple elements",
			input: "<p>First</p><p>Second</p>",
			expected: func() []*Node {
				p1 := NewElement("p", nil)
				p1.AppendChild(NewText("First"))
				p2 := NewElement("p", nil)
				p2.AppendChild(NewText("Second"))
				return []*Node{p1, p2}
			}(),
		},
		{
			name:  "nested elements",
			input: "<div><span>Nested</span></div>",
			expected: func() []*Node {
				div := NewElement("div", nil)
				span := NewElement("span", nil)
				span.AppendChild(NewText("Nested"))
				div.AppendChild(span)
				return []*Node{div}
			}(),
		},
		{
			name:  "element with attributes",
			input: `<a href="https://example.com">Link</a>`,
			expected: func() []*Node {
				a := NewElement("a", map[string]string{"href": "https://example.com"})
				a.AppendChild(NewText("Link"))
				return []*Node{a}
			}(),
		},
		{
			name:  "mixed text and elements",
			input: "Start <strong>bold</strong> end",
			expected: func() []*Node {
				text1 := NewText("Start ")
				strong := NewElement("strong", nil)
				strong.AppendChild(NewText("bold"))
				text2 := NewText(" end")
				return []*Node{text1, strong, text2}
			}(),
		},
		{
			name:     "empty string",
			input:    "",
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseFragment(tt.input)

			if tt.expected == nil {
				assert.Nil(t, result)
				return
			}

			assert.Equal(t, len(tt.expected), len(result), "number of nodes")

			for i := range tt.expected {
				if !nodeEqual(result[i], tt.expected[i]) {
					diff := nodeDiff(result[i], tt.expected[i], fmt.Sprintf("node[%d]", i))
					assert.Fail(t, "ParseFragment() node mismatch",
						"%s\n\nGot:\n%s\nWant:\n%s",
						diff, printTree(result[i], ""), printTree(tt.expected[i], ""))
				}
			}
		})
	}
}
