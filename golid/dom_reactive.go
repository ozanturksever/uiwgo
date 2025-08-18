// dom_reactive.go
// Reactive DOM manipulation primitives for fine-grained reactivity

//go:build js && wasm

package golid

import (
	"sync"
	"sync/atomic"
	"syscall/js"
)

// ------------------------------------
// 🎯 DOM Binding Types
// ------------------------------------

// DOMBinding represents a reactive binding between a signal and a DOM element
type DOMBinding struct {
	id          uint64
	element     js.Value
	property    string
	computation *Computation
	owner       *Owner
	cleanup     func()
	mutex       sync.RWMutex
}

// DOMRenderer manages all DOM bindings and provides batched updates
type DOMRenderer struct {
	bindings  map[uint64]*DOMBinding
	patcher   *DOMPatcher
	scheduler *Scheduler
	mutex     sync.RWMutex
}

// Global DOM renderer instance
var globalDOMRenderer *DOMRenderer
var domRendererOnce sync.Once

// getDOMRenderer returns the global DOM renderer instance
func getDOMRenderer() *DOMRenderer {
	domRendererOnce.Do(func() {
		globalDOMRenderer = &DOMRenderer{
			bindings:  make(map[uint64]*DOMBinding),
			patcher:   newDOMPatcher(),
			scheduler: getScheduler(),
		}
	})
	return globalDOMRenderer
}

// ------------------------------------
// 🔗 Direct DOM Binding API
// ------------------------------------

// BindTextReactive creates a reactive text content binding
func BindTextReactive(element js.Value, textFn func() string) *DOMBinding {
	if !element.Truthy() {
		return nil
	}

	binding := &DOMBinding{
		id:       atomic.AddUint64(&signalIdCounter, 1),
		element:  element,
		property: "textContent",
		owner:    getCurrentOwner(),
	}

	// Create effect that updates text content
	binding.computation = CreateEffect(func() {
		text := textFn()

		// Queue DOM operation for batched execution
		getDOMRenderer().patcher.QueueOperation(DOMOperation{
			type_:    SetText,
			target:   binding.element,
			property: "textContent",
			value:    text,
		})
	}, binding.owner)

	// Register cleanup
	binding.cleanup = func() {
		getDOMRenderer().unregisterBinding(binding.id)
		if binding.computation != nil {
			binding.computation.cleanup()
		}
	}

	// Register with owner for automatic cleanup
	if binding.owner != nil {
		OnCleanup(binding.cleanup)
	}

	// Register with DOM renderer
	getDOMRenderer().registerBinding(binding)

	return binding
}

// BindAttributeReactive creates a reactive attribute binding
func BindAttributeReactive(element js.Value, attr string, valueFn func() string) *DOMBinding {
	if !element.Truthy() {
		return nil
	}

	binding := &DOMBinding{
		id:       atomic.AddUint64(&signalIdCounter, 1),
		element:  element,
		property: attr,
		owner:    getCurrentOwner(),
	}

	// Create effect that updates attribute
	binding.computation = CreateEffect(func() {
		value := valueFn()

		var opType OperationType
		if value == "" {
			opType = RemoveAttribute
		} else {
			opType = SetAttribute
		}

		getDOMRenderer().patcher.QueueOperation(DOMOperation{
			type_:    opType,
			target:   binding.element,
			property: attr,
			value:    value,
		})
	}, binding.owner)

	// Register cleanup
	binding.cleanup = func() {
		getDOMRenderer().unregisterBinding(binding.id)
		if binding.computation != nil {
			binding.computation.cleanup()
		}
	}

	if binding.owner != nil {
		OnCleanup(binding.cleanup)
	}

	getDOMRenderer().registerBinding(binding)
	return binding
}

// BindClassReactive creates a reactive CSS class binding
func BindClassReactive(element js.Value, className string, activeFn func() bool) *DOMBinding {
	if !element.Truthy() {
		return nil
	}

	binding := &DOMBinding{
		id:       atomic.AddUint64(&signalIdCounter, 1),
		element:  element,
		property: "class:" + className,
		owner:    getCurrentOwner(),
	}

	// Create effect that updates CSS class
	binding.computation = CreateEffect(func() {
		active := activeFn()
		classList := binding.element.Get("classList")

		if active {
			classList.Call("add", className)
		} else {
			classList.Call("remove", className)
		}
	}, binding.owner)

	binding.cleanup = func() {
		getDOMRenderer().unregisterBinding(binding.id)
		if binding.computation != nil {
			binding.computation.cleanup()
		}
	}

	if binding.owner != nil {
		OnCleanup(binding.cleanup)
	}

	getDOMRenderer().registerBinding(binding)
	return binding
}

// BindStyleReactive creates a reactive CSS style binding
func BindStyleReactive(element js.Value, property string, valueFn func() string) *DOMBinding {
	if !element.Truthy() {
		return nil
	}

	binding := &DOMBinding{
		id:       atomic.AddUint64(&signalIdCounter, 1),
		element:  element,
		property: "style:" + property,
		owner:    getCurrentOwner(),
	}

	// Create effect that updates CSS style
	binding.computation = CreateEffect(func() {
		value := valueFn()
		style := binding.element.Get("style")
		style.Call("setProperty", property, value)
	}, binding.owner)

	binding.cleanup = func() {
		getDOMRenderer().unregisterBinding(binding.id)
		if binding.computation != nil {
			binding.computation.cleanup()
		}
	}

	if binding.owner != nil {
		OnCleanup(binding.cleanup)
	}

	getDOMRenderer().registerBinding(binding)
	return binding
}

// BindEventReactive creates an event binding with automatic cleanup
func BindEventReactive(element js.Value, event string, handler func(js.Value)) *DOMBinding {
	if !element.Truthy() {
		return nil
	}

	binding := &DOMBinding{
		id:       atomic.AddUint64(&signalIdCounter, 1),
		element:  element,
		property: "event:" + event,
		owner:    getCurrentOwner(),
	}

	// Create JavaScript callback
	jsCallback := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		if len(args) > 0 {
			handler(args[0])
		}
		return nil
	})

	// Add event listener
	binding.element.Call("addEventListener", event, jsCallback)

	// Setup cleanup to remove event listener
	binding.cleanup = func() {
		binding.element.Call("removeEventListener", event, jsCallback)
		jsCallback.Release()
		getDOMRenderer().unregisterBinding(binding.id)
	}

	if binding.owner != nil {
		OnCleanup(binding.cleanup)
	}

	getDOMRenderer().registerBinding(binding)
	return binding
}

// ------------------------------------
// 🔄 Conditional & List Rendering
// ------------------------------------

// ConditionalRender creates a reactive conditional rendering binding
func ConditionalRender(element js.Value, conditionFn func() bool, templateFn func() js.Value) *DOMBinding {
	if !element.Truthy() {
		return nil
	}

	binding := &DOMBinding{
		id:       atomic.AddUint64(&signalIdCounter, 1),
		element:  element,
		property: "conditional",
		owner:    getCurrentOwner(),
	}

	var currentChild js.Value
	var placeholder js.Value

	// Create effect that manages conditional rendering
	binding.computation = CreateEffect(func() {
		condition := conditionFn()

		if condition {
			// Show content
			if !currentChild.Truthy() {
				newChild := templateFn()
				if newChild.Truthy() {
					if placeholder.Truthy() {
						// Replace placeholder with actual content - add null check
						parentNode := placeholder.Get("parentNode")
						if parentNode.Truthy() {
							parentNode.Call("replaceChild", newChild, placeholder)
						}
						placeholder = js.Undefined()
					} else {
						// Append to element
						if binding.element.Truthy() {
							binding.element.Call("appendChild", newChild)
						}
					}
					currentChild = newChild
				}
			}
		} else {
			// Hide content
			if currentChild.Truthy() {
				// Create placeholder comment
				if js.Global().Get("document").Truthy() {
					placeholder = js.Global().Get("document").Call("createComment", "conditional")
					parentNode := currentChild.Get("parentNode")
					if parentNode.Truthy() {
						parentNode.Call("replaceChild", placeholder, currentChild)
					}
				}
				currentChild = js.Undefined()
			}
		}
	}, binding.owner)

	binding.cleanup = func() {
		if currentChild.Truthy() {
			currentChild.Call("remove")
		}
		if placeholder.Truthy() {
			placeholder.Call("remove")
		}
		getDOMRenderer().unregisterBinding(binding.id)
		if binding.computation != nil {
			binding.computation.cleanup()
		}
	}

	if binding.owner != nil {
		OnCleanup(binding.cleanup)
	}

	getDOMRenderer().registerBinding(binding)
	return binding
}

// ListRender creates a reactive list rendering binding with efficient keying
func ListRender[T any](element js.Value, itemsFn func() []T, keyFn func(T) string, templateFn func(T) js.Value) *DOMBinding {
	if !element.Truthy() {
		return nil
	}

	binding := &DOMBinding{
		id:       atomic.AddUint64(&signalIdCounter, 1),
		element:  element,
		property: "list",
		owner:    getCurrentOwner(),
	}

	// Track rendered items by key
	renderedItems := make(map[string]js.Value)
	var currentOrder []string

	// Create effect that manages list rendering
	binding.computation = CreateEffect(func() {
		items := itemsFn()
		newOrder := make([]string, 0, len(items))
		newItems := make(map[string]js.Value)

		// Create DocumentFragment for efficient DOM manipulation - add null check
		var fragment js.Value
		if js.Global().Get("document").Truthy() {
			fragment = js.Global().Get("document").Call("createDocumentFragment")
		} else {
			return // Skip if document is not available
		}

		// Process new items
		for _, item := range items {
			key := keyFn(item)
			newOrder = append(newOrder, key)

			if existingElement, exists := renderedItems[key]; exists && existingElement.Truthy() {
				// Reuse existing element
				newItems[key] = existingElement
				if fragment.Truthy() {
					fragment.Call("appendChild", existingElement)
				}
			} else {
				// Create new element
				newElement := templateFn(item)
				if newElement.Truthy() {
					newItems[key] = newElement
					if fragment.Truthy() {
						fragment.Call("appendChild", newElement)
					}
				}
			}
		}

		// Remove old items that are no longer needed
		for _, key := range currentOrder {
			if _, exists := newItems[key]; !exists {
				if element, exists := renderedItems[key]; exists && element.Truthy() {
					element.Call("remove")
				}
			}
		}

		// Clear and append new content - add null checks
		if binding.element.Truthy() {
			binding.element.Set("innerHTML", "")
			if fragment.Truthy() {
				binding.element.Call("appendChild", fragment)
			}
		}

		// Update tracking
		renderedItems = newItems
		currentOrder = newOrder
	}, binding.owner)

	binding.cleanup = func() {
		// Clean up all rendered items
		for _, element := range renderedItems {
			if element.Truthy() {
				element.Call("remove")
			}
		}
		getDOMRenderer().unregisterBinding(binding.id)
		if binding.computation != nil {
			binding.computation.cleanup()
		}
	}

	if binding.owner != nil {
		OnCleanup(binding.cleanup)
	}

	getDOMRenderer().registerBinding(binding)
	return binding
}

// ------------------------------------
// 🔧 DOM Renderer Management
// ------------------------------------

// registerBinding registers a DOM binding with the renderer
func (r *DOMRenderer) registerBinding(binding *DOMBinding) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	r.bindings[binding.id] = binding
}

// unregisterBinding removes a DOM binding from the renderer
func (r *DOMRenderer) unregisterBinding(id uint64) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	delete(r.bindings, id)
}

// GetBindingCount returns the number of active DOM bindings
func (r *DOMRenderer) GetBindingCount() int {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	return len(r.bindings)
}

// CleanupAllBindings removes all DOM bindings (for testing)
func (r *DOMRenderer) CleanupAllBindings() {
	r.mutex.Lock()
	bindings := make([]*DOMBinding, 0, len(r.bindings))
	for _, binding := range r.bindings {
		bindings = append(bindings, binding)
	}
	r.mutex.Unlock()

	// Cleanup all bindings
	for _, binding := range bindings {
		if binding.cleanup != nil {
			binding.cleanup()
		}
	}
}

// ------------------------------------
// 🧪 Testing & Debugging
// ------------------------------------

// GetDOMRendererStats returns statistics about the DOM renderer
func GetDOMRendererStats() map[string]interface{} {
	renderer := getDOMRenderer()
	renderer.mutex.RLock()
	defer renderer.mutex.RUnlock()

	return map[string]interface{}{
		"activeBindings": len(renderer.bindings),
		"patcherStats":   renderer.patcher.GetStats(),
	}
}

// ResetDOMRenderer resets the global DOM renderer (for testing)
func ResetDOMRenderer() {
	if globalDOMRenderer != nil {
		globalDOMRenderer.CleanupAllBindings()
	}
	domRendererOnce = sync.Once{}
	globalDOMRenderer = nil
}
