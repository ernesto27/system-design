package js

import (
	"browser/dom"
	"fmt"
	"strings"

	"github.com/dop251/goja"
)

type JSRuntime struct {
	vm            *goja.Runtime
	document      *dom.Node
	onReflow      func()
	onAlert       func(message string)
	Events        *EventManager
	onConfirm     func(string) bool
	currentURL    string
	onReload      func()
	onPrompt      func(message, defaultValue string) *string
	elementCache  map[*dom.Node]*goja.Object
	onTitleChange func(string)
}

func NewJSRuntime(document *dom.Node, onReflow func()) *JSRuntime {
	rt := &JSRuntime{
		vm:           goja.New(),
		document:     document,
		onReflow:     onReflow,
		Events:       NewEventManager(),
		elementCache: make(map[*dom.Node]*goja.Object),
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

	// document.documentElement
	docObj.DefineAccessorProperty("documentElement",
		rt.vm.ToValue(func(call goja.FunctionCall) goja.Value {
			for _, child := range rt.document.Children {
				if child.Type == dom.Element {
					return rt.wrapElement(child)
				}
			}
			return goja.Null()
		}),
		nil,
		goja.FLAG_FALSE, goja.FLAG_TRUE)

	docObj.DefineAccessorProperty("head", rt.vm.ToValue(func(call goja.FunctionCall) goja.Value {
		headNode := dom.FindElementsByTagName(rt.document, dom.TagHead)
		if headNode == nil {
			return goja.Null()
		}
		return rt.wrapElement(headNode)
	}),
		nil,
		goja.FLAG_FALSE, goja.FLAG_TRUE)

	docObj.DefineAccessorProperty("title",
		rt.vm.ToValue(func(call goja.FunctionCall) goja.Value {
			return rt.vm.ToValue(dom.FindTitle(rt.document))
		}),
		rt.vm.ToValue(func(call goja.FunctionCall) goja.Value {
			if len(call.Arguments) > 0 {
				newTitle := call.Arguments[0].String()
				titleNode := dom.FindElementsByTagName(rt.document, dom.TagTitle)
				if titleNode != nil {
					titleNode.Children = nil
					titleNode.AppendChild(dom.NewText(newTitle))
				}
				if rt.onTitleChange != nil {
					rt.onTitleChange(newTitle)
				}
			}
			return goja.Undefined()
		}),
		goja.FLAG_FALSE, goja.FLAG_TRUE)

	docObj.DefineAccessorProperty("body", rt.vm.ToValue(func(call goja.FunctionCall) goja.Value {
		bodyNode := dom.FindElementsByTagName(rt.document, dom.TagBody)
		if bodyNode == nil {
			return goja.Null()
		}
		return rt.wrapElement(bodyNode)
	}),
		nil,
		goja.FLAG_FALSE, goja.FLAG_TRUE)

	docObj.DefineAccessorProperty("baseURI",
		rt.vm.ToValue(func(call goja.FunctionCall) goja.Value {
			baseHref := dom.FindBaseHref(rt.document)
			if baseHref != "" {
				return rt.vm.ToValue(baseHref)
			}
			return rt.vm.ToValue(rt.currentURL)
		}),
		nil,
		goja.FLAG_FALSE, goja.FLAG_TRUE)

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

	rt.vm.Set("confirm", func(call goja.FunctionCall) goja.Value {
		message := ""
		if len(call.Arguments) > 0 {
			message = call.Arguments[0].String()
		}

		result := false
		if rt.onConfirm != nil {
			result = rt.onConfirm(message)
		}

		return rt.vm.ToValue(result)
	})

	rt.vm.Set("prompt", func(call goja.FunctionCall) goja.Value {
		message := ""
		defaultValue := ""

		if len(call.Arguments) > 0 {
			message = call.Arguments[0].String()
		}

		if len(call.Arguments) > 1 {
			defaultValue = call.Arguments[1].String()
		}

		if rt.onPrompt != nil {
			result := rt.onPrompt(message, defaultValue)
			if result == nil {
				return goja.Null()
			}
			return rt.vm.ToValue(*result)
		}

		return goja.Null()
	})

	docObj.Set("createElement", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 1 {
			return goja.Null()
		}

		tagName := call.Arguments[0].String()
		newNode := dom.NewElement(tagName, nil)
		return rt.wrapElement(newNode)
	})

	docObj.Set("createTextNode", func(call goja.FunctionCall) goja.Value {
		text := ""
		if len(call.Arguments) > 0 {
			text = call.Arguments[0].String()
		}

		newNode := dom.NewText(text)
		return rt.wrapElement(newNode)
	})

	window := rt.vm.NewObject()
	location := rt.vm.NewObject()

	location.Set("href", rt.currentURL)

	location.Set("reload", func(call goja.FunctionCall) goja.Value {
		if rt.onReload != nil {
			rt.onReload()
		}
		return goja.Undefined()
	})

	window.Set("location", location)
	rt.vm.Set("window", window)

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

	// Check cache first
	if cached, ok := rt.elementCache[node]; ok {
		return cached
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

	// parentElement - only returns Element nodes, not Document
	obj.DefineAccessorProperty("parentElement",
		rt.vm.ToValue(func(call goja.FunctionCall) goja.Value {
			if node.Parent == nil || node.Parent.Type != dom.Element {
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

		rt.Events.AddEventListener(node, eventType, callback)
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

	obj.DefineAccessorProperty("innerText",
		rt.vm.ToValue(func(call goja.FunctionCall) goja.Value {
			return rt.vm.ToValue(node.InnerText())
		}),
		rt.vm.ToValue(func(call goja.FunctionCall) goja.Value {
			if len(call.Arguments) > 0 {
				node.SetInnerText(call.Arguments[0].String())
				if rt.onReflow != nil {
					rt.onReflow()
				}
			}
			return goja.Undefined()
		}),
		goja.FLAG_FALSE, goja.FLAG_TRUE)

	obj.Set("appendChild", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 1 {
			return goja.Undefined()
		}

		childNode := unwrapNode(rt, call.Arguments[0])
		if childNode == nil {
			return goja.Undefined()
		}

		node.AppendChild(childNode)

		if rt.onReflow != nil {
			rt.onReflow()
		}

		return call.Arguments[0]
	})

	obj.Set("removeChild", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 1 {
			return goja.Undefined()
		}

		childNode := unwrapNode(rt, call.Arguments[0])
		if childNode == nil {
			return goja.Undefined()
		}

		node.RemoveChild(childNode)

		if rt.onReflow != nil {
			rt.onReflow()
		}

		return call.Arguments[0]
	})

	obj.Set("remove", func(call goja.FunctionCall) goja.Value {
		node.Remove()

		if rt.onReflow != nil {
			rt.onReflow()
		}
		return goja.Undefined()
	})

	classList := rt.vm.NewObject()
	classList.Set("add", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) > 0 {
			elem.ClassListAdd(call.Arguments[0].String())
		}
		return goja.Undefined()
	})

	classList.Set("remove", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) > 0 {
			elem.ClassListRemove(call.Arguments[0].String())
		}
		return goja.Undefined()
	})

	obj.Set("classList", classList)

	obj.Set("_elem", elem)

	// HTMLStyleElement.disabled property (spec 4.2.6)
	if node.TagName == "style" {
		obj.DefineAccessorProperty("disabled",
			rt.vm.ToValue(func(call goja.FunctionCall) goja.Value {
				return rt.vm.ToValue(node.Disabled)
			}),
			rt.vm.ToValue(func(call goja.FunctionCall) goja.Value {
				if len(call.Arguments) > 0 {
					node.Disabled = call.Arguments[0].ToBoolean()
					if rt.onReflow != nil {
						rt.onReflow()
					}
				}
				return goja.Undefined()
			}),
			goja.FLAG_FALSE, goja.FLAG_TRUE)
	}

	// Cache before returning
	rt.elementCache[node] = obj

	return obj
}

func (rt *JSRuntime) DispatchClick(node *dom.Node) {
	rt.Events.Dispatch(rt, node, "click")
}

func (rt *JSRuntime) SetAlertHandler(handler func(message string)) {
	rt.onAlert = handler
}

func (rt *JSRuntime) SetConfirmHandler(handler func(string) bool) {
	rt.onConfirm = handler
}

func (rt *JSRuntime) SetCurrentURL(url string) {
	rt.currentURL = url
}

func (rt *JSRuntime) SetReloadHandler(handler func()) {
	rt.onReload = handler
}

func (rt *JSRuntime) SetPromptHandler(handler func(message, defaultValue string) *string) {
	rt.onPrompt = handler
}

func (rt *JSRuntime) SetTitleChangeHandler(handler func(string)) {
	rt.onTitleChange = handler
}
