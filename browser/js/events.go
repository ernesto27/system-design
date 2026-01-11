package js

import (
	"browser/dom"

	"github.com/dop251/goja"
)

type EventListener struct {
	eventType string
	callback  goja.Callable
}

type EventManager struct {
	listeners map[string][]EventListener
}

func NewEventManager() *EventManager {
	return &EventManager{
		listeners: make(map[string][]EventListener),
	}
}

func (em *EventManager) AddEventListener(eventType string, callback goja.Callable) {
	em.listeners[eventType] = append(em.listeners[eventType], EventListener{
		eventType: eventType,
		callback:  callback,
	})
}

func (em *EventManager) Dispatch(rt *JSRuntime, node *dom.Node, eventType string) {
	listeners := em.listeners[eventType]

	for _, l := range listeners {
		if l.eventType == eventType {
			event := rt.vm.NewObject()
			event.Set("type", eventType)
			event.Set("target", rt.wrapElement(node))

			l.callback(goja.Undefined(), event)
		}
	}
}
