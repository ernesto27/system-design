package js

import (
	"browser/dom"

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
	nodeListeners := em.listeners[node]
	if nodeListeners == nil {
		return
	}

	listeners := nodeListeners[eventType]
	for _, l := range listeners {
		event := rt.vm.NewObject()
		event.Set("type", eventType)
		event.Set("target", rt.wrapElement(node))

		l.callback(goja.Undefined(), event)
	}
}
