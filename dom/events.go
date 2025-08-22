//go:build js && wasm

package dom

import (
	"fmt"
	"sync"
	"syscall/js"

	reactivity "github.com/ozanturksever/uiwgo/reactivity"
	"honnef.co/go/js/dom/v2"
)

// EventHandler represents a typed event handler
type EventHandler[T dom.Event] func(event T)

// EventBinding represents an active event binding with cleanup capability
type EventBinding struct {
	element   dom.Element
	eventType string
	funcName  string
	cleanupFn func()
	mu        sync.Mutex
	disposed  bool
}

// Dispose removes the event listener and cleans up the JavaScript function
func (eb *EventBinding) Dispose() {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	if eb.disposed {
		return
	}

	if eb.cleanupFn != nil {
		eb.cleanupFn()
	}

	if eb.funcName != "" {
		ReleaseJSFunction(eb.funcName)
	}

	eb.disposed = true
}

// IsDisposed returns whether the event binding has been disposed
func (eb *EventBinding) IsDisposed() bool {
	eb.mu.Lock()
	defer eb.mu.Unlock()
	return eb.disposed
}

// EventManager manages multiple event bindings with automatic cleanup
type EventManager struct {
	bindings []*EventBinding
	mu       sync.RWMutex
}

// NewEventManager creates a new event manager
func NewEventManager() *EventManager {
	return &EventManager{
		bindings: make([]*EventBinding, 0),
	}
}

// AddBinding adds an event binding to the manager
func (em *EventManager) AddBinding(binding *EventBinding) {
	em.mu.Lock()
	defer em.mu.Unlock()
	em.bindings = append(em.bindings, binding)
}

// DisposeAll disposes all managed event bindings
func (em *EventManager) DisposeAll() {
	em.mu.Lock()
	defer em.mu.Unlock()

	for _, binding := range em.bindings {
		binding.Dispose()
	}
	em.bindings = em.bindings[:0]
}

// RemoveDisposed removes disposed bindings from the manager
func (em *EventManager) RemoveDisposed() {
	em.mu.Lock()
	defer em.mu.Unlock()

	activeBindings := make([]*EventBinding, 0, len(em.bindings))
	for _, binding := range em.bindings {
		if !binding.IsDisposed() {
			activeBindings = append(activeBindings, binding)
		}
	}
	em.bindings = activeBindings
}

// Global event manager
var GlobalEventManager = NewEventManager()

// Event binding functions

// BindClick binds a click event handler to an element
func BindClick(element dom.Element, handler func(event dom.Event)) *EventBinding {
	funcName := CreateJSFunctionOnTheFly(func(this js.Value, args []js.Value) any {
		if len(args) > 0 {
			event := dom.WrapEvent(args[0])
			handler(event)
		}
		return nil
	})

	// Use onclick attribute for better compatibility
	element.SetAttribute("onclick", fmt.Sprintf("%s(event)", funcName))

	binding := &EventBinding{
		element:   element,
		eventType: "click",
		funcName:  funcName,
		cleanupFn: func() {
			element.RemoveAttribute("onclick")
		},
	}

	GlobalEventManager.AddBinding(binding)
	return binding
}

// BindClickInline binds a click event using inline onclick attribute
func BindClickInline(element dom.Element, handler func(event dom.Event)) *EventBinding {
	funcName := CreateJSFunctionOnTheFly(func(this js.Value, args []js.Value) any {
		if len(args) > 0 {
			event := dom.WrapEvent(args[0])
			handler(event)
		}
		return nil
	})

	element.SetAttribute("onclick", fmt.Sprintf("%s(event)", funcName))

	binding := &EventBinding{
		element:   element,
		eventType: "click",
		funcName:  funcName,
		cleanupFn: func() {
			element.RemoveAttribute("onclick")
		},
	}

	GlobalEventManager.AddBinding(binding)
	return binding
}

// BindChange binds a change event handler to an element
func BindChange(element dom.Element, handler func(event dom.Event)) *EventBinding {
	funcName := CreateJSFunctionOnTheFly(func(this js.Value, args []js.Value) any {
		if len(args) > 0 {
			event := dom.WrapEvent(args[0])
			handler(event)
		}
		return nil
	})

	element.SetAttribute("onchange", fmt.Sprintf("%s(event)", funcName))

	binding := &EventBinding{
		element:   element,
		eventType: "change",
		funcName:  funcName,
		cleanupFn: func() {
			element.RemoveAttribute("onchange")
		},
	}

	GlobalEventManager.AddBinding(binding)
	return binding
}

// BindInput binds an input event handler to an element
func BindInput(element dom.Element, handler func(event dom.Event)) *EventBinding {
	funcName := CreateJSFunctionOnTheFly(func(this js.Value, args []js.Value) any {
		if len(args) > 0 {
			event := dom.WrapEvent(args[0])
			handler(event)
		}
		return nil
	})

	element.SetAttribute("oninput", fmt.Sprintf("%s(event)", funcName))

	binding := &EventBinding{
		element:   element,
		eventType: "input",
		funcName:  funcName,
		cleanupFn: func() {
			element.RemoveAttribute("oninput")
		},
	}

	GlobalEventManager.AddBinding(binding)
	return binding
}

// BindKeyDown binds a keydown event handler to an element
func BindKeyDown(element dom.Element, handler func(event dom.Event)) *EventBinding {
	funcName := CreateJSFunctionOnTheFly(func(this js.Value, args []js.Value) any {
		if len(args) > 0 {
			event := dom.WrapEvent(args[0])
			handler(event)
		}
		return nil
	})

	element.SetAttribute("onkeydown", fmt.Sprintf("%s(event)", funcName))

	binding := &EventBinding{
		element:   element,
		eventType: "keydown",
		funcName:  funcName,
		cleanupFn: func() {
			element.RemoveAttribute("onkeydown")
		},
	}

	GlobalEventManager.AddBinding(binding)
	return binding
}

// BindKeyUp binds a keyup event handler to an element
func BindKeyUp(element dom.Element, handler func(event dom.Event)) *EventBinding {
	funcName := CreateJSFunctionOnTheFly(func(this js.Value, args []js.Value) any {
		if len(args) > 0 {
			event := dom.WrapEvent(args[0])
			handler(event)
		}
		return nil
	})

	element.SetAttribute("onkeyup", fmt.Sprintf("%s(event)", funcName))

	binding := &EventBinding{
		element:   element,
		eventType: "keyup",
		funcName:  funcName,
		cleanupFn: func() {
			element.RemoveAttribute("onkeyup")
		},
	}

	GlobalEventManager.AddBinding(binding)
	return binding
}

// BindSubmit binds a submit event handler to a form element
func BindSubmit(element dom.Element, handler func(event dom.Event)) *EventBinding {
	funcName := CreateJSFunctionOnTheFly(func(this js.Value, args []js.Value) any {
		if len(args) > 0 {
			event := dom.WrapEvent(args[0])
			handler(event)
		}
		return nil
	})

	element.SetAttribute("onsubmit", fmt.Sprintf("%s(event)", funcName))

	binding := &EventBinding{
		element:   element,
		eventType: "submit",
		funcName:  funcName,
		cleanupFn: func() {
			element.RemoveAttribute("onsubmit")
		},
	}

	GlobalEventManager.AddBinding(binding)
	return binding
}

// BindMouseOver binds a mouseover event handler to an element
func BindMouseOver(element dom.Element, handler func(event dom.Event)) *EventBinding {
	funcName := CreateJSFunctionOnTheFly(func(this js.Value, args []js.Value) any {
		if len(args) > 0 {
			event := dom.WrapEvent(args[0])
			handler(event)
		}
		return nil
	})

	element.SetAttribute("onmouseover", fmt.Sprintf("%s(event)", funcName))

	binding := &EventBinding{
		element:   element,
		eventType: "mouseover",
		funcName:  funcName,
		cleanupFn: func() {
			element.RemoveAttribute("onmouseover")
		},
	}

	GlobalEventManager.AddBinding(binding)
	return binding
}

// BindMouseOut binds a mouseout event handler to an element
func BindMouseOut(element dom.Element, handler func(event dom.Event)) *EventBinding {
	funcName := CreateJSFunctionOnTheFly(func(this js.Value, args []js.Value) any {
		if len(args) > 0 {
			event := dom.WrapEvent(args[0])
			handler(event)
		}
		return nil
	})

	element.SetAttribute("onmouseout", fmt.Sprintf("%s(event)", funcName))

	binding := &EventBinding{
		element:   element,
		eventType: "mouseout",
		funcName:  funcName,
		cleanupFn: func() {
			element.RemoveAttribute("onmouseout")
		},
	}

	GlobalEventManager.AddBinding(binding)
	return binding
}

// BindFocus binds a focus event handler to an element
func BindFocus(element dom.Element, handler func(event dom.Event)) *EventBinding {
	funcName := CreateJSFunctionOnTheFly(func(this js.Value, args []js.Value) any {
		if len(args) > 0 {
			event := dom.WrapEvent(args[0])
			handler(event)
		}
		return nil
	})

	element.SetAttribute("onfocus", fmt.Sprintf("%s(event)", funcName))

	binding := &EventBinding{
		element:   element,
		eventType: "focus",
		funcName:  funcName,
		cleanupFn: func() {
			element.RemoveAttribute("onfocus")
		},
	}

	GlobalEventManager.AddBinding(binding)
	return binding
}

// BindBlur binds a blur event handler to an element
func BindBlur(element dom.Element, handler func(event dom.Event)) *EventBinding {
	funcName := CreateJSFunctionOnTheFly(func(this js.Value, args []js.Value) any {
		if len(args) > 0 {
			event := dom.WrapEvent(args[0])
			handler(event)
		}
		return nil
	})

	element.SetAttribute("onblur", fmt.Sprintf("%s(event)", funcName))

	binding := &EventBinding{
		element:   element,
		eventType: "blur",
		funcName:  funcName,
		cleanupFn: func() {
			element.RemoveAttribute("onblur")
		},
	}

	GlobalEventManager.AddBinding(binding)
	return binding
}

// BindGenericEvent binds a generic event handler to an element
func BindGenericEvent(element dom.Element, eventType string, handler func(event dom.Event)) *EventBinding {
	funcName := CreateJSFunctionOnTheFly(func(this js.Value, args []js.Value) any {
		if len(args) > 0 {
			event := dom.WrapEvent(args[0])
			handler(event)
		}
		return nil
	})

	// Convert event type to attribute name (e.g., "click" -> "onclick")
	attrName := "on" + eventType
	element.SetAttribute(attrName, fmt.Sprintf("%s(event)", funcName))

	binding := &EventBinding{
		element:   element,
		eventType: eventType,
		funcName:  funcName,
		cleanupFn: func() {
			element.RemoveAttribute(attrName)
		},
	}

	GlobalEventManager.AddBinding(binding)
	return binding
}

// Reactive event binding functions

// BindClickToSignal binds a click event that updates a signal
func BindClickToSignal[T any](element dom.Element, signal reactivity.Signal[T], value T) *EventBinding {
	return BindClick(element, func(event dom.Event) {
		signal.Set(value)
	})
}

// BindClickToCallback binds a click event that calls a callback function
func BindClickToCallback(element dom.Element, callback func()) *EventBinding {
	return BindClick(element, func(event dom.Event) {
		callback()
	})
}

// BindInputToSignal binds an input event that updates a string signal with the input value
func BindInputToSignal(element dom.Element, signal reactivity.Signal[string]) *EventBinding {
	return BindInput(element, func(event dom.Event) {
		if target := event.Target(); target != nil {
			if input, ok := target.(dom.HTMLInputElement); ok {
				signal.Set(input.Value())
			}
		}
	})
}

// BindChangeToSignal binds a change event that updates a string signal with the element value
func BindChangeToSignal(element dom.Element, signal reactivity.Signal[string]) *EventBinding {
	return BindChange(element, func(event dom.Event) {
		if target := event.Target(); target != nil {
			if input, ok := target.(dom.HTMLInputElement); ok {
				signal.Set(input.Value())
			} else if textarea, ok := target.(dom.HTMLTextAreaElement); ok {
				signal.Set(textarea.Value())
			} else if selectEl, ok := target.(dom.HTMLSelectElement); ok {
				signal.Set(selectEl.Value())
			}
		}
	})
}

// BindKeyDownToCallback binds a keydown event for a specific key
func BindKeyDownToCallback(element dom.Element, key string, callback func()) *EventBinding {
	return BindKeyDown(element, func(event dom.Event) {
		eventKey := event.Underlying().Get("key").String()
		if eventKey == key {
			callback()
		}
	})
}

// BindEnterKeyToCallback binds the Enter key to a callback
func BindEnterKeyToCallback(element dom.Element, callback func()) *EventBinding {
	return BindKeyDownToCallback(element, "Enter", callback)
}

// BindEscapeKeyToCallback binds the Escape key to a callback
func BindEscapeKeyToCallback(element dom.Element, callback func()) *EventBinding {
	return BindKeyDownToCallback(element, "Escape", callback)
}

// Cleanup functions

// CleanupAllEvents disposes all globally managed event bindings
func CleanupAllEvents() {
	GlobalEventManager.DisposeAll()
}

// CleanupDisposedEvents removes disposed event bindings from the global manager
func CleanupDisposedEvents() {
	GlobalEventManager.RemoveDisposed()
}

// Event delegation helpers

// DelegateEvent sets up event delegation for dynamically added elements
func DelegateEvent(container dom.Element, eventType string, selector string, handler func(event dom.Event, target dom.Element)) *EventBinding {
	funcName := CreateJSFunctionOnTheFly(func(this js.Value, args []js.Value) any {
		if len(args) > 0 {
			event := dom.WrapEvent(args[0])
			target := event.Target()

			// Check if target matches selector
			if target != nil {
				// Simple selector matching - can be enhanced
				if element, ok := target.(dom.Element); ok {
					if matchesSelector(element, selector) {
						handler(event, element)
					}
				}
			}
		}
		return nil
	})

	// Convert event type to attribute name and set up delegation
	attrName := "on" + eventType
	container.SetAttribute(attrName, fmt.Sprintf("%s(event)", funcName))

	binding := &EventBinding{
		element:   container,
		eventType: eventType,
		funcName:  funcName,
		cleanupFn: func() {
			container.RemoveAttribute(attrName)
		},
	}

	GlobalEventManager.AddBinding(binding)
	return binding
}

// Simple selector matching helper (basic implementation)
func matchesSelector(element dom.Element, selector string) bool {
	// Basic implementation - can be enhanced with more complex selectors
	switch {
	case selector[0] == '#': // ID selector
		return element.ID() == selector[1:]
	case selector[0] == '.': // Class selector
		return element.Class().Contains(selector[1:])
	default: // Tag selector
		return element.TagName() == selector
	}
}
