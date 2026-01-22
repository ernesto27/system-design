package js

import (
	"browser/dom"
	"slices"
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

func (e *Element) GetInnerHTML() string {
	var result strings.Builder
	for _, child := range e.node.Children {
		serializeNode(&result, child)
	}
	return result.String()
}

func (e *Element) SetInnerHTML(htmlContent string) {
	e.node.Children = []*dom.Node{}

	parsed := dom.ParseFragment(htmlContent)

	for _, child := range parsed {
		e.node.AppendChild(child)
	}

	if e.rt != nil && e.rt.onReflow != nil {
		e.rt.onReflow()
	}
}

func (e *Element) getClasses() []string {
	classAttr := e.node.Attributes["class"]
	if classAttr == "" {
		return []string{}
	}
	return strings.Fields(classAttr)
}

func (e *Element) setClasses(classes []string) {
	if e.node.Attributes == nil {
		e.node.Attributes = make(map[string]string)
	}
	e.node.Attributes["class"] = strings.Join(classes, " ")
}

func (e *Element) ClassListAdd(className string) {
	classes := e.getClasses()
	if slices.Contains(classes, className) {
		return
	}

	classes = append(classes, className)
	e.setClasses(classes)

	if e.rt.onReflow != nil {
		e.rt.onReflow()
	}
}

func (e *Element) ClassListRemove(className string) {
	classes := e.getClasses()
	classes = slices.DeleteFunc(classes, func(c string) bool {
		return c == className
	})
	e.setClasses(classes)

	if e.rt.onReflow != nil {
		e.rt.onReflow()
	}
}

// serializeNode converts a DOM node back to HTML string
func serializeNode(sb *strings.Builder, node *dom.Node) {
	// Handle text nodes - just write the text
	if node.Type == dom.Text {
		sb.WriteString(node.Text)
		return
	}

	// Handle element nodes - write opening tag
	sb.WriteString("<")
	sb.WriteString(node.TagName)

	// Write attributes
	for name, value := range node.Attributes {
		sb.WriteString(" ")
		sb.WriteString(name)
		sb.WriteString(`="`)
		sb.WriteString(value)
		sb.WriteString(`"`)
	}
	sb.WriteString(">")

	// Recursively serialize children
	for _, child := range node.Children {
		serializeNode(sb, child)
	}

	// Write closing tag
	sb.WriteString("</")
	sb.WriteString(node.TagName)
	sb.WriteString(">")
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

func unwrapNode(rt *JSRuntime, val goja.Value) *dom.Node {
	if val == nil || goja.IsNull(val) || goja.IsUndefined(val) {
		return nil
	}

	obj := val.ToObject(rt.vm)
	if obj == nil {
		return nil
	}

	elemval := obj.Get("_elem")
	if elemval == nil || goja.IsUndefined(elemval) {
		return nil
	}

	elem, ok := elemval.Export().(*Element)
	if !ok || elem == nil {
		return nil
	}

	return elem.node
}
