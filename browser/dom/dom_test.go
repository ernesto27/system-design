package dom

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewElement(t *testing.T) {
	tests := []struct {
		name       string
		tagName    string
		attributes map[string]string
	}{
		{
			name:       "div with class attribute",
			tagName:    "div",
			attributes: map[string]string{"class": "container"},
		},
		{
			name:       "a with href",
			tagName:    "a",
			attributes: map[string]string{"href": "https://example.com", "target": "_blank"},
		},
		{
			name:       "empty attributes",
			tagName:    "span",
			attributes: map[string]string{},
		},
		{
			name:       "nil attributes",
			tagName:    "p",
			attributes: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node := NewElement(tt.tagName, tt.attributes)

			assert.Equal(t, Element, node.Type, "node Type")
			assert.Equal(t, tt.tagName, node.TagName, "node TagName")
			assert.NotNil(t, node.Children, "Children should not be nil")
			assert.Empty(t, node.Children, "Children should be empty")

			// Check attributes
			for k, v := range tt.attributes {
				assert.Equal(t, v, node.Attributes[k], "Attributes[%q]", k)
			}
		})
	}
}

func TestNewText(t *testing.T) {
	tests := []struct {
		name string
		text string
	}{
		{"simple text", "Hello, World!"},
		{"empty text", ""},
		{"text with whitespace", "  some text  "},
		{"multiline text", "line1\nline2"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node := NewText(tt.text)

			assert.Equal(t, Text, node.Type, "node Type")
			assert.Equal(t, tt.text, node.Text, "node Text")
		})
	}
}

func TestAppendChild(t *testing.T) {
	t.Run("single child", func(t *testing.T) {
		parent := NewElement("div", nil)
		child := NewElement("span", nil)

		parent.AppendChild(child)

		assert.Len(t, parent.Children, 1)
		assert.Same(t, child, parent.Children[0], "appended child should be the same reference")
		assert.Same(t, parent, child.Parent, "child.Parent should be parent")
	})

	t.Run("multiple children", func(t *testing.T) {
		parent := NewElement("ul", nil)
		child1 := NewElement("li", nil)
		child2 := NewElement("li", nil)
		child3 := NewElement("li", nil)

		parent.AppendChild(child1)
		parent.AppendChild(child2)
		parent.AppendChild(child3)

		assert.Len(t, parent.Children, 3)
		assert.Same(t, child1, parent.Children[0])
		assert.Same(t, child2, parent.Children[1])
		assert.Same(t, child3, parent.Children[2])
		assert.Same(t, parent, child1.Parent)
		assert.Same(t, parent, child2.Parent)
		assert.Same(t, parent, child3.Parent)
	})

	t.Run("text child", func(t *testing.T) {
		parent := NewElement("p", nil)
		textChild := NewText("Hello")

		parent.AppendChild(textChild)

		assert.Len(t, parent.Children, 1)
		assert.Equal(t, Text, parent.Children[0].Type)
		assert.Equal(t, "Hello", parent.Children[0].Text)
	})
}

func TestFindTitle(t *testing.T) {
	tests := []struct {
		name     string
		build    func() *Node
		expected string
	}{
		{
			name: "title in head",
			build: func() *Node {
				html := NewElement("html", nil)
				head := NewElement("head", nil)
				title := NewElement("title", nil)
				titleText := NewText("My Page Title")

				title.AppendChild(titleText)
				head.AppendChild(title)
				html.AppendChild(head)
				return html
			},
			expected: "My Page Title",
		},
		{
			name: "nested title deep in tree",
			build: func() *Node {
				html := NewElement("html", nil)
				head := NewElement("head", nil)
				meta := NewElement("meta", nil)
				title := NewElement("title", nil)
				titleText := NewText("Deep Title")

				title.AppendChild(titleText)
				head.AppendChild(meta)
				head.AppendChild(title)
				html.AppendChild(head)
				return html
			},
			expected: "Deep Title",
		},
		{
			name: "no title element",
			build: func() *Node {
				html := NewElement("html", nil)
				body := NewElement("body", nil)
				html.AppendChild(body)
				return html
			},
			expected: "",
		},
		{
			name: "empty title element",
			build: func() *Node {
				html := NewElement("html", nil)
				head := NewElement("head", nil)
				title := NewElement("title", nil)

				head.AppendChild(title)
				html.AppendChild(head)
				return html
			},
			expected: "",
		},
		{
			name: "nil node",
			build: func() *Node {
				return nil
			},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node := tt.build()
			result := FindTitle(node)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFindStyleContent(t *testing.T) {
	tests := []struct {
		name     string
		build    func() *Node
		expected string
	}{
		{
			name: "single style block",
			build: func() *Node {
				html := NewElement("html", nil)
				head := NewElement("head", nil)
				style := NewElement("style", nil)
				styleText := NewText("body { color: red; }")

				style.AppendChild(styleText)
				head.AppendChild(style)
				html.AppendChild(head)
				return html
			},
			expected: "body { color: red; }\n",
		},
		{
			name: "multiple style blocks",
			build: func() *Node {
				html := NewElement("html", nil)
				head := NewElement("head", nil)
				style1 := NewElement("style", nil)
				style1Text := NewText("body { color: red; }")
				style2 := NewElement("style", nil)
				style2Text := NewText("p { margin: 10px; }")

				style1.AppendChild(style1Text)
				style2.AppendChild(style2Text)
				head.AppendChild(style1)
				head.AppendChild(style2)
				html.AppendChild(head)
				return html
			},
			expected: "body { color: red; }\np { margin: 10px; }\n",
		},
		{
			name: "no style element",
			build: func() *Node {
				html := NewElement("html", nil)
				body := NewElement("body", nil)
				html.AppendChild(body)
				return html
			},
			expected: "",
		},
		{
			name: "nil node",
			build: func() *Node {
				return nil
			},
			expected: "",
		},
		{
			name: "style in body",
			build: func() *Node {
				html := NewElement("html", nil)
				body := NewElement("body", nil)
				style := NewElement("style", nil)
				styleText := NewText("div { padding: 5px; }")

				style.AppendChild(styleText)
				body.AppendChild(style)
				html.AppendChild(body)
				return html
			},
			expected: "div { padding: 5px; }\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node := tt.build()
			result := FindStyleContent(node)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFindStylesheetLinks(t *testing.T) {
	tests := []struct {
		name     string
		build    func() *Node
		expected []string
	}{
		{
			name: "single stylesheet link",
			build: func() *Node {
				html := NewElement("html", nil)
				head := NewElement("head", nil)
				link := NewElement("link", map[string]string{
					"rel":  "stylesheet",
					"href": "styles.css",
				})

				head.AppendChild(link)
				html.AppendChild(head)
				return html
			},
			expected: []string{"styles.css"},
		},
		{
			name: "multiple stylesheet links",
			build: func() *Node {
				html := NewElement("html", nil)
				head := NewElement("head", nil)
				link1 := NewElement("link", map[string]string{
					"rel":  "stylesheet",
					"href": "main.css",
				})
				link2 := NewElement("link", map[string]string{
					"rel":  "stylesheet",
					"href": "theme.css",
				})

				head.AppendChild(link1)
				head.AppendChild(link2)
				html.AppendChild(head)
				return html
			},
			expected: []string{"main.css", "theme.css"},
		},
		{
			name: "non-stylesheet link ignored",
			build: func() *Node {
				html := NewElement("html", nil)
				head := NewElement("head", nil)
				link := NewElement("link", map[string]string{
					"rel":  "icon",
					"href": "favicon.ico",
				})

				head.AppendChild(link)
				html.AppendChild(head)
				return html
			},
			expected: nil,
		},
		{
			name: "link without href",
			build: func() *Node {
				html := NewElement("html", nil)
				head := NewElement("head", nil)
				link := NewElement("link", map[string]string{
					"rel": "stylesheet",
				})

				head.AppendChild(link)
				html.AppendChild(head)
				return html
			},
			expected: nil,
		},
		{
			name: "mixed links",
			build: func() *Node {
				html := NewElement("html", nil)
				head := NewElement("head", nil)
				link1 := NewElement("link", map[string]string{
					"rel":  "stylesheet",
					"href": "styles.css",
				})
				link2 := NewElement("link", map[string]string{
					"rel":  "icon",
					"href": "favicon.ico",
				})
				link3 := NewElement("link", map[string]string{
					"rel":  "stylesheet",
					"href": "print.css",
				})

				head.AppendChild(link1)
				head.AppendChild(link2)
				head.AppendChild(link3)
				html.AppendChild(head)
				return html
			},
			expected: []string{"styles.css", "print.css"},
		},
		{
			name: "no links",
			build: func() *Node {
				html := NewElement("html", nil)
				body := NewElement("body", nil)
				html.AppendChild(body)
				return html
			},
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node := tt.build()
			result := FindStylesheetLinks(node)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFindElementsByTagName(t *testing.T) {
	tests := []struct {
		name        string
		build       func() *Node
		tagName     string
		expectNil   bool
		expectedTag string
	}{
		{
			name: "find head element",
			build: func() *Node {
				html := NewElement("html", nil)
				head := NewElement("head", nil)
				body := NewElement("body", nil)
				html.AppendChild(head)
				html.AppendChild(body)
				return html
			},
			tagName:     "head",
			expectNil:   false,
			expectedTag: "head",
		},
		{
			name: "find body element",
			build: func() *Node {
				html := NewElement("html", nil)
				head := NewElement("head", nil)
				body := NewElement("body", nil)
				html.AppendChild(head)
				html.AppendChild(body)
				return html
			},
			tagName:     "body",
			expectNil:   false,
			expectedTag: "body",
		},
		{
			name: "find nested element",
			build: func() *Node {
				html := NewElement("html", nil)
				body := NewElement("body", nil)
				div := NewElement("div", nil)
				p := NewElement("p", nil)
				div.AppendChild(p)
				body.AppendChild(div)
				html.AppendChild(body)
				return html
			},
			tagName:     "p",
			expectNil:   false,
			expectedTag: "p",
		},
		{
			name: "element not found",
			build: func() *Node {
				html := NewElement("html", nil)
				body := NewElement("body", nil)
				html.AppendChild(body)
				return html
			},
			tagName:   "head",
			expectNil: true,
		},
		{
			name: "nil node",
			build: func() *Node {
				return nil
			},
			tagName:   "body",
			expectNil: true,
		},
		{
			name: "find first of multiple",
			build: func() *Node {
				html := NewElement("html", nil)
				body := NewElement("body", nil)
				p1 := NewElement("p", map[string]string{"id": "first"})
				p2 := NewElement("p", map[string]string{"id": "second"})
				body.AppendChild(p1)
				body.AppendChild(p2)
				html.AppendChild(body)
				return html
			},
			tagName:     "p",
			expectNil:   false,
			expectedTag: "p",
		},
		{
			name: "find root element itself",
			build: func() *Node {
				html := NewElement("html", nil)
				return html
			},
			tagName:     "html",
			expectNil:   false,
			expectedTag: "html",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node := tt.build()
			result := FindElementsByTagName(node, tt.tagName)

			if tt.expectNil {
				assert.Nil(t, result)
			} else {
				assert.NotNil(t, result)
				assert.Equal(t, tt.expectedTag, result.TagName)
			}
		})
	}
}
