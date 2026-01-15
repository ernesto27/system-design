package js

import (
	"browser/dom"
	"strings"

	"github.com/dop251/goja"
)

type Document struct {
	rt   *JSRuntime
	root *dom.Node
}

func newDocument(rt *JSRuntime, root *dom.Node) *Document {
	return &Document{
		rt:   rt,
		root: root,
	}
}

func (d *Document) GetElementById(id string) goja.Value {
	node := findNodeById(d.root, id)
	if node == nil {
		return goja.Null()
	}

	return d.rt.wrapElement(node)
}

func (d *Document) QuerySelector(selector string) goja.Value {
	selector = strings.TrimSpace(selector)
	if selector == "" {
		return goja.Null()
	}

	node := findNodeBySelector(d.root, selector)
	if node == nil {
		return goja.Null()
	}

	return d.rt.wrapElement(node)
}

func findNodeById(node *dom.Node, id string) *dom.Node {
	if node == nil {
		return nil
	}

	if node.Type == dom.Element && node.Attributes["id"] == id {
		return node
	}

	for _, child := range node.Children {
		if found := findNodeById(child, id); found != nil {
			return found
		}
	}

	return nil
}

func findNodeBySelector(node *dom.Node, selector string) *dom.Node {
	if node == nil {
		return nil
	}

	if matchesSelector(node, selector) {
		return node
	}

	for _, child := range node.Children {
		if found := findNodeBySelector(child, selector); found != nil {
			return found
		}
	}

	return nil
}

func matchesSelector(node *dom.Node, selector string) bool {
	if node.Type != dom.Element {
		return false
	}

	switch {
	case strings.HasPrefix(selector, "#"):
		id := strings.TrimPrefix(selector, "#")
		return id != "" && node.Attributes["id"] == id
	case strings.HasPrefix(selector, "."):
		className := strings.TrimPrefix(selector, ".")
		return className != "" && hasClass(node.Attributes["class"], className)
	default:
		return strings.EqualFold(node.TagName, selector)
	}
}

func hasClass(classAttr, className string) bool {
	for _, class := range strings.Fields(classAttr) {
		if class == className {
			return true
		}
	}

	return false
}
