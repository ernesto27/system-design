package js

import (
	"browser/dom"
	"testing"

	"github.com/dop251/goja"
	"github.com/stretchr/testify/assert"
)

func TestDocumentQuerySelector(t *testing.T) {
	root := &dom.Node{Type: dom.Document, Children: []*dom.Node{}}
	body := dom.NewElement("body", nil)
	root.AppendChild(body)

	div := dom.NewElement("div", map[string]string{
		"id":    "main",
		"class": "container",
	})
	body.AppendChild(div)

	paragraph := dom.NewElement("p", map[string]string{
		"id": "primary",
	})
	div.AppendChild(paragraph)

	span := dom.NewElement("span", map[string]string{
		"class": "item active",
	})
	body.AppendChild(span)

	rt := NewJSRuntime(root, nil)
	doc := newDocument(rt, root)

	tests := []struct {
		name      string
		selector  string
		wantTag   string
		wantID    string
		wantClass string
	}{
		{
			name:     "tag selector",
			selector: "div",
			wantTag:  "DIV",
			wantID:   "main",
		},
		{
			name:     "id selector",
			selector: "#primary",
			wantTag:  "P",
			wantID:   "primary",
		},
		{
			name:      "class selector",
			selector:  ".item",
			wantTag:   "SPAN",
			wantClass: "item active",
		},
		{
			name:     "missing selector",
			selector: ".missing",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			value := doc.QuerySelector(tt.selector)
			if tt.wantTag == "" {
				assert.True(t, goja.IsNull(value))
				return
			}

			obj := value.ToObject(rt.vm)
			assert.Equal(t, tt.wantTag, obj.Get("tagName").String())
			if tt.wantID != "" {
				assert.Equal(t, tt.wantID, obj.Get("id").String())
			}
			if tt.wantClass != "" {
				assert.Equal(t, tt.wantClass, obj.Get("className").String())
			}
		})
	}
}
