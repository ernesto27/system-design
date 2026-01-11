package js

import (
	"browser/dom"
	"strings"

	"github.com/dop251/goja"
)

type Element struct {
	rt   *JSRuntime
	node *dom.Node
}

func newElement(rt *JSRuntime, node *dom.Node) *Element {
	return &Element{rt: rt, node: node}
}

// GetAttribute returns an attribute value
func (e *Element) GetAttribute(name string) goja.Value {
	val, ok := e.node.Attributes[name]
	if !ok {
		return goja.Null()
	}
	return e.rt.vm.ToValue(val)
}

// SetAttribute sets an attribute value
func (e *Element) SetAttribute(name, value string) {
	if e.node.Attributes == nil {
		e.node.Attributes = make(map[string]string)
	}
	e.node.Attributes[name] = value
}

// GetTextContent returns all text content
func (e *Element) GetTextContent() string {
	return collectText(e.node)
}

func (e *Element) SetTextContent(text string) {
	e.node.Children = []*dom.Node{}
	if text != "" {
		textNode := dom.NewText(text)
		e.node.AppendChild(textNode)
	}

	if e.rt.onReflow != nil {
		e.rt.onReflow()
	}
}

func collectText(node *dom.Node) string {
	if node == nil {
		return ""
	}

	if node.Type == dom.Text {
		return node.Text
	}

	var sb strings.Builder
	for _, child := range node.Children {
		sb.WriteString(collectText(child))
	}
	return sb.String()
}
