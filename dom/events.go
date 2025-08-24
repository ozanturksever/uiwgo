//go:build js && wasm

package dom

import (
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
	jsFunc := js.FuncOf(func(this js.Value, args []js.Value) any {
		if len(args) > 0 {
			event := dom.WrapEvent(args[0])
			handler(event)
		}
		return nil
	})

	// Use addEventListener instead of inline attribute
	element.Underlying().Call("addEventListener", "click", jsFunc)

	binding := &EventBinding{
		element:   element,
		eventType: "click",
		funcName:  "", // not used with addEventListener
		cleanupFn: func() {
			element.Underlying().Call("removeEventListener", "click", jsFunc)
			jsFunc.Release()
		},
	}

	GlobalEventManager.AddBinding(binding)
	return binding
}

// BindClickInline binds a click event using inline onclick attribute
func BindClickInline(element dom.Element, handler func(event dom.Event)) *EventBinding {
	// Keep API but implement with addEventListener for reliability
	jsFunc := js.FuncOf(func(this js.Value, args []js.Value) any {
		if len(args) > 0 {
			event := dom.WrapEvent(args[0])
			handler(event)
		}
		return nil
	})

	element.Underlying().Call("addEventListener", "click", jsFunc)

	binding := &EventBinding{
		element:   element,
		eventType: "click",
		funcName:  "",
		cleanupFn: func() {
			element.Underlying().Call("removeEventListener", "click", jsFunc)
			jsFunc.Release()
		},
	}

	GlobalEventManager.AddBinding(binding)
	return binding
}

// BindChange binds a change event handler to an element
func BindChange(element dom.Element, handler func(event dom.Event)) *EventBinding {
	jsFunc := js.FuncOf(func(this js.Value, args []js.Value) any {
		if len(args) > 0 {
			event := dom.WrapEvent(args[0])
			handler(event)
		}
		return nil
	})

	element.Underlying().Call("addEventListener", "change", jsFunc)

	binding := &EventBinding{
		element:   element,
		eventType: "change",
		funcName:  "",
		cleanupFn: func() {
			element.Underlying().Call("removeEventListener", "change", jsFunc)
			jsFunc.Release()
		},
	}

	GlobalEventManager.AddBinding(binding)
	return binding
}

// BindInput binds an input event handler to an element
func BindInput(element dom.Element, handler func(event dom.Event)) *EventBinding {
	jsFunc := js.FuncOf(func(this js.Value, args []js.Value) any {
		if len(args) > 0 {
			event := dom.WrapEvent(args[0])
			handler(event)
		}
		return nil
	})

	element.Underlying().Call("addEventListener", "input", jsFunc)

	binding := &EventBinding{
		element:   element,
		eventType: "input",
		funcName:  "",
		cleanupFn: func() {
			element.Underlying().Call("removeEventListener", "input", jsFunc)
			jsFunc.Release()
		},
	}

	GlobalEventManager.AddBinding(binding)
	return binding
}

// BindKeyDown binds a keydown event handler to an element
func BindKeyDown(element dom.Element, handler func(event dom.Event)) *EventBinding {
	jsFunc := js.FuncOf(func(this js.Value, args []js.Value) any {
		if len(args) > 0 {
			event := dom.WrapEvent(args[0])
			handler(event)
		}
		return nil
	})

	element.Underlying().Call("addEventListener", "keydown", jsFunc)

	binding := &EventBinding{
		element:   element,
		eventType: "keydown",
		funcName:  "",
		cleanupFn: func() {
			element.Underlying().Call("removeEventListener", "keydown", jsFunc)
			jsFunc.Release()
		},
	}

	GlobalEventManager.AddBinding(binding)
	return binding
}

// BindKeyUp binds a keyup event handler to an element
func BindKeyUp(element dom.Element, handler func(event dom.Event)) *EventBinding {
	jsFunc := js.FuncOf(func(this js.Value, args []js.Value) any {
		if len(args) > 0 {
			event := dom.WrapEvent(args[0])
			handler(event)
		}
		return nil
	})

	element.Underlying().Call("addEventListener", "keyup", jsFunc)

	binding := &EventBinding{
		element:   element,
		eventType: "keyup",
		funcName:  "",
		cleanupFn: func() {
			element.Underlying().Call("removeEventListener", "keyup", jsFunc)
			jsFunc.Release()
		},
	}

	GlobalEventManager.AddBinding(binding)
	return binding
}

// BindSubmit binds a submit event handler to a form element
func BindSubmit(element dom.Element, handler func(event dom.Event)) *EventBinding {
	jsFunc := js.FuncOf(func(this js.Value, args []js.Value) any {
		if len(args) > 0 {
			event := dom.WrapEvent(args[0])
			handler(event)
		}
		return nil
	})

	element.Underlying().Call("addEventListener", "submit", jsFunc)

	binding := &EventBinding{
		element:   element,
		eventType: "submit",
		funcName:  "",
		cleanupFn: func() {
			element.Underlying().Call("removeEventListener", "submit", jsFunc)
			jsFunc.Release()
		},
	}

	GlobalEventManager.AddBinding(binding)
	return binding
}

// BindMouseOver binds a mouseover event handler to an element
func BindMouseOver(element dom.Element, handler func(event dom.Event)) *EventBinding {
	jsFunc := js.FuncOf(func(this js.Value, args []js.Value) any {
		if len(args) > 0 {
			event := dom.WrapEvent(args[0])
			handler(event)
		}
		return nil
	})

	element.Underlying().Call("addEventListener", "mouseover", jsFunc)

	binding := &EventBinding{
		element:   element,
		eventType: "mouseover",
		funcName:  "",
		cleanupFn: func() {
			element.Underlying().Call("removeEventListener", "mouseover", jsFunc)
			jsFunc.Release()
		},
	}

	GlobalEventManager.AddBinding(binding)
	return binding
}

// BindMouseOut binds a mouseout event handler to an element
func BindMouseOut(element dom.Element, handler func(event dom.Event)) *EventBinding {
	jsFunc := js.FuncOf(func(this js.Value, args []js.Value) any {
		if len(args) > 0 {
			event := dom.WrapEvent(args[0])
			handler(event)
		}
		return nil
	})

	element.Underlying().Call("addEventListener", "mouseout", jsFunc)

	binding := &EventBinding{
		element:   element,
		eventType: "mouseout",
		funcName:  "",
		cleanupFn: func() {
			element.Underlying().Call("removeEventListener", "mouseout", jsFunc)
			jsFunc.Release()
		},
	}

	GlobalEventManager.AddBinding(binding)
	return binding
}

// BindFocus binds a focus event handler to an element
func BindFocus(element dom.Element, handler func(event dom.Event)) *EventBinding {
	jsFunc := js.FuncOf(func(this js.Value, args []js.Value) any {
		if len(args) > 0 {
			event := dom.WrapEvent(args[0])
			handler(event)
		}
		return nil
	})

	element.Underlying().Call("addEventListener", "focus", jsFunc)

	binding := &EventBinding{
		element:   element,
		eventType: "focus",
		funcName:  "",
		cleanupFn: func() {
			element.Underlying().Call("removeEventListener", "focus", jsFunc)
			jsFunc.Release()
		},
	}

	GlobalEventManager.AddBinding(binding)
	return binding
}

// BindBlur binds a blur event handler to an element
func BindBlur(element dom.Element, handler func(event dom.Event)) *EventBinding {
	jsFunc := js.FuncOf(func(this js.Value, args []js.Value) any {
		if len(args) > 0 {
			event := dom.WrapEvent(args[0])
			handler(event)
		}
		return nil
	})

	element.Underlying().Call("addEventListener", "blur", jsFunc)

	binding := &EventBinding{
		element:   element,
		eventType: "blur",
		funcName:  "",
		cleanupFn: func() {
			element.Underlying().Call("removeEventListener", "blur", jsFunc)
			jsFunc.Release()
		},
	}

	GlobalEventManager.AddBinding(binding)
	return binding
}

// BindGenericEvent binds a generic event handler to an element
func BindGenericEvent(element dom.Element, eventType string, handler func(event dom.Event)) *EventBinding {
	jsFunc := js.FuncOf(func(this js.Value, args []js.Value) any {
		if len(args) > 0 {
			event := dom.WrapEvent(args[0])
			handler(event)
		}
		return nil
	})

	element.Underlying().Call("addEventListener", eventType, jsFunc)

	binding := &EventBinding{
		element:   element,
		eventType: eventType,
		funcName:  "",
		cleanupFn: func() {
			element.Underlying().Call("removeEventListener", eventType, jsFunc)
			jsFunc.Release()
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
			// Use the underlying JavaScript value to get the input value
			value := target.Underlying().Get("value").String()
			signal.Set(value)
		}
	})
}

// BindChangeToSignal binds a change event that updates a string signal with the element value
func BindChangeToSignal(element dom.Element, signal reactivity.Signal[string]) *EventBinding {
	return BindChange(element, func(event dom.Event) {
		if target := event.Target(); target != nil {
			// Use the underlying JavaScript value to get the element value
			value := target.Underlying().Get("value").String()
			signal.Set(value)
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

// DelegateEvent attaches a delegated handler on parent for events bubbling from descendants
// matching the CSS selector. Uses Element.closest() so selectors like
// [data-action='toggle'] and complex selectors are supported.
func DelegateEvent(parent dom.Element, eventType string, selector string, handler func(e dom.Event, target dom.Element)) *EventBinding {
	jsFunc := js.FuncOf(func(this js.Value, args []js.Value) any {
		if len(args) == 0 {
			return nil
		}
		rawEvent := args[0]
		ev := dom.WrapEvent(rawEvent)

		target := rawEvent.Get("target")
		if target.IsUndefined() || target.IsNull() {
			return nil
		}

		// closest() supports full CSS selectors, including attribute selectors with quotes
		matched := target.Call("closest", selector)
		if !matched.IsUndefined() && !matched.IsNull() {
			handler(ev, dom.WrapElement(matched))
		}
		return nil
	})

	// Use addEventListener for delegation on the parent
	parent.Underlying().Call("addEventListener", eventType, jsFunc)

	binding := &EventBinding{
		element:   parent,
		eventType: eventType,
		funcName:  "",
		cleanupFn: func() {
			parent.Underlying().Call("removeEventListener", eventType, jsFunc)
			jsFunc.Release()
		},
	}

	GlobalEventManager.AddBinding(binding)
	return binding

}

// matchesSelector checks if the given element matches the provided CSS selector.
// It uses the native Element.matches() with vendor-prefixed fallbacks, so it
// supports full CSS selectors (including attribute and complex selectors).
// Note: For delegated scenarios where the listener is on an ancestor, consider
// using Element.closest(selector) to both find and verify matches up the tree.
func matchesSelector(element dom.Element, selector string) bool {
	u := element.Underlying()
	if !u.Truthy() {
		return false
	}

	// Standard method
	if u.Get("matches").Truthy() {
		return u.Call("matches", selector).Bool()
	}

	// Vendor-prefixed fallbacks
	if m := u.Get("webkitMatchesSelector"); m.Truthy() {
		return m.Invoke(selector).Bool()
	}
	if m := u.Get("msMatchesSelector"); m.Truthy() {
		return m.Invoke(selector).Bool()
	}

	// Last-resort fallback: use closest() and compare
	closest := u.Call("closest", selector)
	return !closest.IsUndefined() && !closest.IsNull() && closest.Equal(u)
}
