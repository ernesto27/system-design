package js

import (
	"browser/dom"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSerializeNode(t *testing.T) {
	tests := []struct {
		name     string
		build    func() *dom.Node
		expected string
	}{
		{
			name: "text node only",
			build: func() *dom.Node {
				return dom.NewText("Hello World")
			},
			expected: "Hello World",
		},
		{
			name: "empty element",
			build: func() *dom.Node {
				return dom.NewElement("div", nil)
			},
			expected: "<div></div>",
		},
		{
			name: "element with text child",
			build: func() *dom.Node {
				p := dom.NewElement("p", nil)
				p.AppendChild(dom.NewText("Hello"))
				return p
			},
			expected: "<p>Hello</p>",
		},
		{
			name: "element with attributes",
			build: func() *dom.Node {
				return dom.NewElement("div", map[string]string{
					"id":    "main",
					"class": "container",
				})
			},
			expected: `<div id="main" class="container"></div>`,
		},
		{
			name: "nested elements",
			build: func() *dom.Node {
				div := dom.NewElement("div", nil)
				p := dom.NewElement("p", nil)
				p.AppendChild(dom.NewText("Hello"))
				div.AppendChild(p)
				return div
			},
			expected: "<div><p>Hello</p></div>",
		},
		{
			name: "multiple children",
			build: func() *dom.Node {
				div := dom.NewElement("div", nil)
				p := dom.NewElement("p", nil)
				p.AppendChild(dom.NewText("Hello"))
				span := dom.NewElement("span", nil)
				span.AppendChild(dom.NewText("World"))
				div.AppendChild(p)
				div.AppendChild(span)
				return div
			},
			expected: "<div><p>Hello</p><span>World</span></div>",
		},
		{
			name: "deeply nested structure",
			build: func() *dom.Node {
				outer := dom.NewElement("div", nil)
				inner := dom.NewElement("div", nil)
				p := dom.NewElement("p", nil)
				p.AppendChild(dom.NewText("Deep"))
				inner.AppendChild(p)
				outer.AppendChild(inner)
				return outer
			},
			expected: "<div><div><p>Deep</p></div></div>",
		},
		{
			name: "mixed text and elements",
			build: func() *dom.Node {
				div := dom.NewElement("div", nil)
				div.AppendChild(dom.NewText("Start "))
				span := dom.NewElement("span", nil)
				span.AppendChild(dom.NewText("middle"))
				div.AppendChild(span)
				div.AppendChild(dom.NewText(" end"))
				return div
			},
			expected: "<div>Start <span>middle</span> end</div>",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node := tt.build()
			var sb strings.Builder
			serializeNode(&sb, node)
			result := sb.String()

			if strings.Contains(tt.name, "attributes") {
				assert.Contains(t, result, `id="main"`)
				assert.Contains(t, result, `class="container"`)
				assert.True(t, strings.HasPrefix(result, "<div "))
				assert.True(t, strings.HasSuffix(result, "></div>"))
			} else {
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestGetInnerHTML(t *testing.T) {
	tests := []struct {
		name     string
		build    func() *dom.Node
		expected string
	}{
		{
			name: "element with text only",
			build: func() *dom.Node {
				div := dom.NewElement("div", nil)
				div.AppendChild(dom.NewText("Hello"))
				return div
			},
			expected: "Hello",
		},
		{
			name: "element with no children",
			build: func() *dom.Node {
				return dom.NewElement("div", nil)
			},
			expected: "",
		},
		{
			name: "element with child elements",
			build: func() *dom.Node {
				div := dom.NewElement("div", nil)
				p := dom.NewElement("p", nil)
				p.AppendChild(dom.NewText("Hello"))
				span := dom.NewElement("span", nil)
				span.AppendChild(dom.NewText("World"))
				div.AppendChild(p)
				div.AppendChild(span)
				return div
			},
			expected: "<p>Hello</p><span>World</span>",
		},
		{
			name: "deeply nested children",
			build: func() *dom.Node {
				outer := dom.NewElement("div", nil)
				inner := dom.NewElement("div", nil)
				p := dom.NewElement("p", nil)
				p.AppendChild(dom.NewText("Nested"))
				inner.AppendChild(p)
				outer.AppendChild(inner)
				return outer
			},
			expected: "<div><p>Nested</p></div>",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node := tt.build()

			elem := &Element{node: node}
			result := elem.GetInnerHTML()
			assert.Equal(t, tt.expected, result)
		})
	}
}
