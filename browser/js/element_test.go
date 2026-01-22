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

func TestGetClasses(t *testing.T) {
	tests := []struct {
		name     string
		class    string
		expected []string
	}{
		{"empty string", "", []string{}},
		{"single class", "container", []string{"container"}},
		{"multiple classes", "foo bar baz", []string{"foo", "bar", "baz"}},
		{"extra whitespace", "  foo   bar  ", []string{"foo", "bar"}},
		{"tabs and newlines", "foo\tbar\nbaz", []string{"foo", "bar", "baz"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node := dom.NewElement("div", map[string]string{"class": tt.class})
			elem := &Element{node: node}
			result := elem.getClasses()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSetClasses(t *testing.T) {
	tests := []struct {
		name     string
		classes  []string
		expected string
	}{
		{"empty slice", []string{}, ""},
		{"single class", []string{"container"}, "container"},
		{"multiple classes", []string{"foo", "bar", "baz"}, "foo bar baz"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node := dom.NewElement("div", nil)
			elem := &Element{node: node}
			elem.setClasses(tt.classes)
			assert.Equal(t, tt.expected, node.Attributes["class"])
		})
	}

	t.Run("nil attributes map", func(t *testing.T) {
		node := &dom.Node{Type: dom.Element, TagName: "div", Attributes: nil}
		elem := &Element{node: node}
		elem.setClasses([]string{"test"})
		assert.Equal(t, "test", node.Attributes["class"])
	})
}

func TestClassListAdd(t *testing.T) {
	tests := []struct {
		name          string
		initialClass  string
		addClass      string
		expectedClass string
	}{
		{"add to empty", "", "highlight", "highlight"},
		{"add new class", "box", "highlight", "box highlight"},
		{"add duplicate", "box highlight", "highlight", "box highlight"},
		{"add to multiple", "a b c", "d", "a b c d"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node := dom.NewElement("div", map[string]string{"class": tt.initialClass})
			rt := &JSRuntime{}
			elem := &Element{node: node, rt: rt}
			elem.ClassListAdd(tt.addClass)
			assert.Equal(t, tt.expectedClass, node.Attributes["class"])
		})
	}
}

func TestClassListRemove(t *testing.T) {
	tests := []struct {
		name          string
		initialClass  string
		removeClass   string
		expectedClass string
	}{
		{"remove existing class", "box blue", "blue", "box"},
		{"remove only class", "highlight", "highlight", ""},
		{"remove non-existent class", "box", "blue", "box"},
		{"remove from multiple", "a b c", "b", "a c"},
		{"remove first class", "first second", "first", "second"},
		{"remove from empty", "", "anything", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node := dom.NewElement("div", map[string]string{"class": tt.initialClass})
			rt := &JSRuntime{}
			elem := &Element{node: node, rt: rt}
			elem.ClassListRemove(tt.removeClass)
			assert.Equal(t, tt.expectedClass, node.Attributes["class"])
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
