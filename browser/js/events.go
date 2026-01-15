package js

import (
	"browser/dom"
	"fmt"

	"github.com/dop251/goja"
)

type EventListener struct {
	eventType string
	callback  goja.Callable
}

// EventManager manages event listeners per DOM node
type EventManager struct {
	// Map from DOM node -> event type -> list of listeners
	listeners map[*dom.Node]map[string][]EventListener
}

func NewEventManager() *EventManager {
	return &EventManager{
		listeners: make(map[*dom.Node]map[string][]EventListener),
	}
}

// AddEventListener registers a callback for a specific node and event type
func (em *EventManager) AddEventListener(node *dom.Node, eventType string, callback goja.Callable) {
	if em.listeners[node] == nil {
		em.listeners[node] = make(map[string][]EventListener)
	}
	em.listeners[node][eventType] = append(em.listeners[node][eventType], EventListener{
		eventType: eventType,
		callback:  callback,
	})
}

// Dispatch fires all listeners for the given node and event type
func (em *EventManager) Dispatch(rt *JSRuntime, node *dom.Node, eventType string) {
	fmt.Printf("Dispatch: eventType=%s, node=%p, tagName=%s\n", eventType, node, node.TagName)
	fmt.Printf("  Total registered nodes: %d\n", len(em.listeners))
	for n := range em.listeners {
		fmt.Printf("    Registered node: %p tagName=%s id=%s\n", n, n.TagName, n.Attributes["id"])
	}

	// Bubble up through the DOM tree
	current := node
	for current != nil {
		fmt.Printf("  Checking node: %p tagName=%s\n", current, current.TagName)
		nodeListeners := em.listeners[current]
		if nodeListeners != nil {
			listeners := nodeListeners[eventType]
			fmt.Printf("    Found %d listeners for %s\n", len(listeners), eventType)
			for _, l := range listeners {
				event := rt.vm.NewObject()
				event.Set("type", eventType)
				event.Set("target", rt.wrapElement(node)) // original target
				event.Set("currentTarget", rt.wrapElement(current))
				l.callback(goja.Undefined(), event)
			}
		}
		current = current.Parent
	}
}
