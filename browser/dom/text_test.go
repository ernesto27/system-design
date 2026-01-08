package dom

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExtractText(t *testing.T) {
	tests := []struct {
		name     string
		build    func() *Node
		expected string
	}{
		{
			name: "simple text in paragraph",
			build: func() *Node {
				p := NewElement("p", nil)
				text := NewText("Hello World")
				p.AppendChild(text)
				return p
			},
			expected: "Hello World",
		},
		{
			name: "nested text in divs",
			build: func() *Node {
				div := NewElement("div", nil)
				p1 := NewElement("p", nil)
				p2 := NewElement("p", nil)
				text1 := NewText("First")
				text2 := NewText("Second")

				p1.AppendChild(text1)
				p2.AppendChild(text2)
				div.AppendChild(p1)
				div.AppendChild(p2)
				return div
			},
			// ExtractText adds space after text nodes
			expected: "First \nSecond",
		},
		{
			name: "inline elements preserve flow",
			build: func() *Node {
				p := NewElement("p", nil)
				text1 := NewText("Hello")
				strong := NewElement("strong", nil)
				strongText := NewText("World")
				text2 := NewText("!")

				strong.AppendChild(strongText)
				p.AppendChild(text1)
				p.AppendChild(strong)
				p.AppendChild(text2)
				return p
			},
			expected: "Hello World !",
		},
		{
			name: "skip script content",
			build: func() *Node {
				div := NewElement("div", nil)
				text1 := NewText("Before")
				script := NewElement("script", nil)
				scriptText := NewText("var x = 1;")
				text2 := NewText("After")

				script.AppendChild(scriptText)
				div.AppendChild(text1)
				div.AppendChild(script)
				div.AppendChild(text2)
				return div
			},
			expected: "Before After",
		},
		{
			name: "skip style content",
			build: func() *Node {
				div := NewElement("div", nil)
				text1 := NewText("Before")
				style := NewElement("style", nil)
				styleText := NewText("body { color: red; }")
				text2 := NewText("After")

				style.AppendChild(styleText)
				div.AppendChild(text1)
				div.AppendChild(style)
				div.AppendChild(text2)
				return div
			},
			expected: "Before After",
		},
		{
			name: "block elements add newlines",
			build: func() *Node {
				div := NewElement("div", nil)
				h1 := NewElement("h1", nil)
				h1Text := NewText("Title")
				p := NewElement("p", nil)
				pText := NewText("Paragraph")

				h1.AppendChild(h1Text)
				p.AppendChild(pText)
				div.AppendChild(h1)
				div.AppendChild(p)
				return div
			},
			// ExtractText adds space after text nodes
			expected: "Title \nParagraph",
		},
		{
			name: "empty element",
			build: func() *Node {
				return NewElement("div", nil)
			},
			expected: "",
		},
		{
			name: "deeply nested text",
			build: func() *Node {
				div := NewElement("div", nil)
				section := NewElement("section", nil)
				article := NewElement("article", nil)
				p := NewElement("p", nil)
				text := NewText("Deep content")

				p.AppendChild(text)
				article.AppendChild(p)
				section.AppendChild(article)
				div.AppendChild(section)
				return div
			},
			expected: "Deep content",
		},
		{
			name: "mixed block and inline",
			build: func() *Node {
				body := NewElement("body", nil)
				h1 := NewElement("h1", nil)
				h1Text := NewText("Welcome")
				p := NewElement("p", nil)
				text1 := NewText("This is")
				em := NewElement("em", nil)
				emText := NewText("important")
				text2 := NewText("text.")

				h1.AppendChild(h1Text)
				em.AppendChild(emText)
				p.AppendChild(text1)
				p.AppendChild(em)
				p.AppendChild(text2)
				body.AppendChild(h1)
				body.AppendChild(p)
				return body
			},
			// ExtractText adds space after text nodes
			expected: "Welcome \nThis is important text.",
		},
		{
			name: "list items",
			build: func() *Node {
				ul := NewElement("ul", nil)
				li1 := NewElement("li", nil)
				li2 := NewElement("li", nil)
				li3 := NewElement("li", nil)
				text1 := NewText("Item 1")
				text2 := NewText("Item 2")
				text3 := NewText("Item 3")

				li1.AppendChild(text1)
				li2.AppendChild(text2)
				li3.AppendChild(text3)
				ul.AppendChild(li1)
				ul.AppendChild(li2)
				ul.AppendChild(li3)
				return ul
			},
			// ExtractText adds space after text nodes
			expected: "Item 1 \nItem 2 \nItem 3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node := tt.build()
			result := node.ExtractText()

			// Normalize expected for comparison (trim and normalize newlines)
			expected := strings.TrimSpace(tt.expected)
			result = strings.TrimSpace(result)

			assert.Equal(t, expected, result)
		})
	}
}
