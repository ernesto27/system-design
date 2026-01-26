package js

import (
	"browser/dom"

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
