package js

import (
	"browser/dom"
	"fmt"
	"strings"

	"github.com/dop251/goja"
)

type JSRuntime struct {
	vm       *goja.Runtime
	document *dom.Node
	onReflow func()
	onAlert  func(message string)
	Events   *EventManager
}

func NewJSRuntime(document *dom.Node, onReflow func()) *JSRuntime {
	rt := &JSRuntime{
		vm:       goja.New(),
		document: document,
		onReflow: onReflow,
		Events:   NewEventManager(),
	}
	rt.setupGlobals()
	return rt
}

func (rt *JSRuntime) setupGlobals() {
	console := rt.vm.NewObject()
	console.Set("log", func(call goja.FunctionCall) goja.Value {
		for _, arg := range call.Arguments {
			fmt.Print(arg.String(), " ")
		}
		fmt.Println()
		return goja.Undefined()
	})
	rt.vm.Set("console", console)

	doc := newDocument(rt, rt.document)
	docObj := rt.vm.NewObject()
	docObj.Set("getElementById", doc.GetElementById)
	rt.vm.Set("document", docObj)

	rt.vm.Set("alert", func(call goja.FunctionCall) goja.Value {
		message := ""
		if len(call.Arguments) > 0 {
			message = call.Arguments[0].String()
		}
		if rt.onAlert != nil {
			rt.onAlert(message)
		}

		return goja.Undefined()
	})

}

func (rt *JSRuntime) Execute(code string) error {
	_, err := rt.vm.RunString(code)
	if err != nil {
		fmt.Println("JS error: ", err)
	}
	return err
}

// FindScripts extracts JavaScript code from <script> tags
func FindScripts(node *dom.Node) []string {
	var scripts []string
	findScriptsRecursive(node, &scripts)
	return scripts
}

func findScriptsRecursive(node *dom.Node, scripts *[]string) {
	if node == nil {
		return
	}

	if node.Type == dom.Element && node.TagName == "script" {
		// Get inline script content
		for _, child := range node.Children {
			if child.Type == dom.Text && child.Text != "" {
				*scripts = append(*scripts, child.Text)
			}
		}
	}

	for _, child := range node.Children {
		findScriptsRecursive(child, scripts)
	}
}

func (rt *JSRuntime) wrapElement(node *dom.Node) goja.Value {
	if node == nil {
		return goja.Null()
	}

	elem := newElement(rt, node)
	obj := rt.vm.NewObject()

	// Static properties
	obj.Set("tagName", strings.ToUpper(node.TagName))
	obj.Set("id", node.Attributes["id"])
	obj.Set("className", node.Attributes["class"])

	// Methods
	obj.Set("getAttribute", elem.GetAttribute)
	obj.Set("setAttribute", elem.SetAttribute)

	// Dynamic property: textContent (getter/setter)
	obj.DefineAccessorProperty("textContent",
		rt.vm.ToValue(func(call goja.FunctionCall) goja.Value {
			return rt.vm.ToValue(elem.GetTextContent())
		}),
		rt.vm.ToValue(func(call goja.FunctionCall) goja.Value {
			if len(call.Arguments) > 0 {
				elem.SetTextContent(call.Arguments[0].String())
			}
			return goja.Undefined()
		}),
		goja.FLAG_FALSE, goja.FLAG_TRUE)

	// parentElement
	obj.DefineAccessorProperty("parentElement",
		rt.vm.ToValue(func(call goja.FunctionCall) goja.Value {
			if node.Parent == nil {
				return goja.Null()
			}
			return rt.wrapElement(node.Parent)
		}),
		nil,
		goja.FLAG_FALSE, goja.FLAG_TRUE)

	obj.Set("addEventListener", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 2 {
			return goja.Undefined()
		}

		eventType := call.Arguments[0].String()

		callback, ok := goja.AssertFunction(call.Arguments[1])
		if !ok {
			return goja.Undefined()
		}

		rt.Events.AddEventListener(eventType, callback)
		return goja.Undefined()
	})

	obj.DefineAccessorProperty("innerHTML",
		rt.vm.ToValue(func(call goja.FunctionCall) goja.Value {
			return rt.vm.ToValue(elem.GetInnerHTML())
		}),
		rt.vm.ToValue(func(call goja.FunctionCall) goja.Value {
			if len(call.Arguments) > 0 {
				elem.SetInnerHTML(call.Arguments[0].String())
			}
			return goja.Undefined()
		}),
		goja.FLAG_FALSE, goja.FLAG_TRUE)

	return obj
}

func (rt *JSRuntime) DispatchClick(node *dom.Node) {
	rt.Events.Dispatch(rt, node, "click")
}

func (rt *JSRuntime) SetAlertHandler(handler func(message string)) {
	rt.onAlert = handler
}
