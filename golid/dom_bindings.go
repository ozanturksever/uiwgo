// dom_bindings.go
// Fine-grained DOM binding system for reactive UI components

//go:build js && wasm

package golid

import (
	"strings"
	"sync"
	"sync/atomic"
	"syscall/js"
)

// ------------------------------------
// 🔗 High-Level Binding API
// ------------------------------------

// ElementBinder provides a fluent API for binding reactive values to DOM elements
type ElementBinder struct {
	element  js.Value
	bindings []*DOMBinding
	owner    *Owner
	mutex    sync.RWMutex
}

// NewElementBinder creates a new element binder for the given DOM element
func NewElementBinder(element js.Value) *ElementBinder {
	if !element.Truthy() {
		return nil
	}

	return &ElementBinder{
		element:  element,
		bindings: make([]*DOMBinding, 0),
		owner:    getCurrentOwner(),
	}
}

// Text binds reactive text content to the element
func (b *ElementBinder) Text(textFn func() string) *ElementBinder {
	if b == nil {
		return b
	}

	binding := BindTextReactive(b.element, textFn)
	if binding != nil {
		b.mutex.Lock()
		b.bindings = append(b.bindings, binding)
		b.mutex.Unlock()
	}

	return b
}

// Attr binds a reactive attribute to the element
func (b *ElementBinder) Attr(name string, valueFn func() string) *ElementBinder {
	if b == nil {
		return b
	}

	binding := BindAttributeReactive(b.element, name, valueFn)
	if binding != nil {
		b.mutex.Lock()
		b.bindings = append(b.bindings, binding)
		b.mutex.Unlock()
	}

	return b
}

// Class binds a reactive CSS class to the element
func (b *ElementBinder) Class(className string, activeFn func() bool) *ElementBinder {
	if b == nil {
		return b
	}

	binding := BindClassReactive(b.element, className, activeFn)
	if binding != nil {
		b.mutex.Lock()
		b.bindings = append(b.bindings, binding)
		b.mutex.Unlock()
	}

	return b
}

// Classes binds multiple reactive CSS classes to the element
func (b *ElementBinder) Classes(classesFn func() map[string]bool) *ElementBinder {
	if b == nil {
		return b
	}

	// Create a single effect that manages all classes
	binding := &DOMBinding{
		id:       atomic.AddUint64(&signalIdCounter, 1),
		element:  b.element,
		property: "classes",
		owner:    b.owner,
	}

	binding.computation = CreateEffect(func() {
		classes := classesFn()
		classList := b.element.Get("classList")

		// Get current classes to determine what to remove
		currentClasses := make(map[string]bool)
		length := classList.Get("length").Int()
		for i := 0; i < length; i++ {
			className := classList.Call("item", i).String()
			currentClasses[className] = true
		}

		// Apply new classes
		for className, active := range classes {
			if active {
				classList.Call("add", className)
			} else {
				classList.Call("remove", className)
			}
		}
	}, b.owner)

	binding.cleanup = func() {
		getDOMRenderer().unregisterBinding(binding.id)
		if binding.computation != nil {
			binding.computation.cleanup()
		}
	}

	if b.owner != nil {
		OnCleanup(binding.cleanup)
	}

	getDOMRenderer().registerBinding(binding)

	b.mutex.Lock()
	b.bindings = append(b.bindings, binding)
	b.mutex.Unlock()

	return b
}

// Style binds a reactive CSS style property to the element
func (b *ElementBinder) Style(property string, valueFn func() string) *ElementBinder {
	if b == nil {
		return b
	}

	binding := BindStyleReactive(b.element, property, valueFn)
	if binding != nil {
		b.mutex.Lock()
		b.bindings = append(b.bindings, binding)
		b.mutex.Unlock()
	}

	return b
}

// Styles binds multiple reactive CSS styles to the element
func (b *ElementBinder) Styles(stylesFn func() map[string]string) *ElementBinder {
	if b == nil {
		return b
	}

	// Create a single effect that manages all styles
	binding := &DOMBinding{
		id:       atomic.AddUint64(&signalIdCounter, 1),
		element:  b.element,
		property: "styles",
		owner:    b.owner,
	}

	binding.computation = CreateEffect(func() {
		styles := stylesFn()
		elementStyle := b.element.Get("style")

		for property, value := range styles {
			if value == "" {
				elementStyle.Call("removeProperty", property)
			} else {
				elementStyle.Call("setProperty", property, value)
			}
		}
	}, b.owner)

	binding.cleanup = func() {
		getDOMRenderer().unregisterBinding(binding.id)
		if binding.computation != nil {
			binding.computation.cleanup()
		}
	}

	if b.owner != nil {
		OnCleanup(binding.cleanup)
	}

	getDOMRenderer().registerBinding(binding)

	b.mutex.Lock()
	b.bindings = append(b.bindings, binding)
	b.mutex.Unlock()

	return b
}

// On binds an event handler to the element
func (b *ElementBinder) On(event string, handler func(js.Value)) *ElementBinder {
	if b == nil {
		return b
	}

	binding := BindEventReactive(b.element, event, handler)
	if binding != nil {
		b.mutex.Lock()
		b.bindings = append(b.bindings, binding)
		b.mutex.Unlock()
	}

	return b
}

// Show conditionally shows/hides the element
func (b *ElementBinder) Show(conditionFn func() bool) *ElementBinder {
	if b == nil {
		return b
	}

	// Store original display style
	var originalDisplay string
	if b.element.Get("style").Get("display").String() != "" {
		originalDisplay = b.element.Get("style").Get("display").String()
	} else {
		originalDisplay = "block" // Default display
	}

	binding := &DOMBinding{
		id:       atomic.AddUint64(&signalIdCounter, 1),
		element:  b.element,
		property: "show",
		owner:    b.owner,
	}

	binding.computation = CreateEffect(func() {
		show := conditionFn()
		style := b.element.Get("style")

		if show {
			style.Set("display", originalDisplay)
		} else {
			style.Set("display", "none")
		}
	}, b.owner)

	binding.cleanup = func() {
		getDOMRenderer().unregisterBinding(binding.id)
		if binding.computation != nil {
			binding.computation.cleanup()
		}
	}

	if b.owner != nil {
		OnCleanup(binding.cleanup)
	}

	getDOMRenderer().registerBinding(binding)

	b.mutex.Lock()
	b.bindings = append(b.bindings, binding)
	b.mutex.Unlock()

	return b
}

// Value binds reactive value to form inputs
func (b *ElementBinder) Value(valueFn func() string, onInput func(string)) *ElementBinder {
	if b == nil {
		return b
	}

	// Bind value property
	valueBinding := &DOMBinding{
		id:       atomic.AddUint64(&signalIdCounter, 1),
		element:  b.element,
		property: "value",
		owner:    b.owner,
	}

	valueBinding.computation = CreateEffect(func() {
		value := valueFn()
		currentValue := b.element.Get("value").String()
		if currentValue != value {
			b.element.Set("value", value)
		}
	}, b.owner)

	valueBinding.cleanup = func() {
		getDOMRenderer().unregisterBinding(valueBinding.id)
		if valueBinding.computation != nil {
			valueBinding.computation.cleanup()
		}
	}

	if b.owner != nil {
		OnCleanup(valueBinding.cleanup)
	}

	getDOMRenderer().registerBinding(valueBinding)

	// Bind input event if handler provided
	if onInput != nil {
		inputBinding := BindEventReactive(b.element, "input", func(event js.Value) {
			value := event.Get("target").Get("value").String()
			onInput(value)
		})

		b.mutex.Lock()
		b.bindings = append(b.bindings, valueBinding)
		if inputBinding != nil {
			b.bindings = append(b.bindings, inputBinding)
		}
		b.mutex.Unlock()
	} else {
		b.mutex.Lock()
		b.bindings = append(b.bindings, valueBinding)
		b.mutex.Unlock()
	}

	return b
}

// Cleanup removes all bindings from the element
func (b *ElementBinder) Cleanup() {
	if b == nil {
		return
	}

	b.mutex.Lock()
	defer b.mutex.Unlock()

	for _, binding := range b.bindings {
		if binding.cleanup != nil {
			binding.cleanup()
		}
	}

	b.bindings = b.bindings[:0]
}

// GetBindingCount returns the number of active bindings
func (b *ElementBinder) GetBindingCount() int {
	if b == nil {
		return 0
	}

	b.mutex.RLock()
	defer b.mutex.RUnlock()
	return len(b.bindings)
}

// ------------------------------------
// 🎯 Specialized Binding Functions
// ------------------------------------

// BindFormInput creates a two-way binding for form inputs
func BindFormInput(element js.Value, getValue func() string, setValue func(string)) *ElementBinder {
	binder := NewElementBinder(element)
	if binder == nil {
		return nil
	}

	return binder.Value(getValue, setValue)
}

// BindCheckbox creates a two-way binding for checkboxes
func BindCheckbox(element js.Value, getValue func() bool, setValue func(bool)) *ElementBinder {
	binder := NewElementBinder(element)
	if binder == nil {
		return nil
	}

	// Bind checked property
	checkedBinding := &DOMBinding{
		id:       atomic.AddUint64(&signalIdCounter, 1),
		element:  element,
		property: "checked",
		owner:    getCurrentOwner(),
	}

	checkedBinding.computation = CreateEffect(func() {
		checked := getValue()
		element.Set("checked", checked)
	}, checkedBinding.owner)

	checkedBinding.cleanup = func() {
		getDOMRenderer().unregisterBinding(checkedBinding.id)
		if checkedBinding.computation != nil {
			checkedBinding.computation.cleanup()
		}
	}

	if checkedBinding.owner != nil {
		OnCleanup(checkedBinding.cleanup)
	}

	getDOMRenderer().registerBinding(checkedBinding)

	// Bind change event
	changeBinding := BindEventReactive(element, "change", func(event js.Value) {
		checked := event.Get("target").Get("checked").Bool()
		setValue(checked)
	})

	binder.mutex.Lock()
	binder.bindings = append(binder.bindings, checkedBinding)
	if changeBinding != nil {
		binder.bindings = append(binder.bindings, changeBinding)
	}
	binder.mutex.Unlock()

	return binder
}

// BindSelect creates a two-way binding for select elements
func BindSelect(element js.Value, getValue func() string, setValue func(string)) *ElementBinder {
	binder := NewElementBinder(element)
	if binder == nil {
		return nil
	}

	return binder.Value(getValue, setValue).On("change", func(event js.Value) {
		value := event.Get("target").Get("value").String()
		setValue(value)
	})
}

// ------------------------------------
// 🔄 List Binding Utilities
// ------------------------------------

// ListBinder provides utilities for binding reactive lists to DOM elements
type ListBinder[T any] struct {
	container js.Value
	items     func() []T
	keyFn     func(T) string
	renderFn  func(T) js.Value
	binding   *DOMBinding
}

// NewListBinder creates a new list binder
func NewListBinder[T any](container js.Value, items func() []T, keyFn func(T) string, renderFn func(T) js.Value) *ListBinder[T] {
	if !container.Truthy() {
		return nil
	}

	binder := &ListBinder[T]{
		container: container,
		items:     items,
		keyFn:     keyFn,
		renderFn:  renderFn,
	}

	// Create the list rendering binding
	binder.binding = ListRender(container, items, keyFn, renderFn)

	return binder
}

// Update forces an update of the list
func (lb *ListBinder[T]) Update() {
	if lb == nil || lb.binding == nil || lb.binding.computation == nil {
		return
	}

	lb.binding.computation.markDirty()
}

// Cleanup removes the list binding
func (lb *ListBinder[T]) Cleanup() {
	if lb == nil || lb.binding == nil {
		return
	}

	if lb.binding.cleanup != nil {
		lb.binding.cleanup()
	}
}

// ------------------------------------
// 🧪 Utility Functions
// ------------------------------------

// BindElement creates an element binder for a DOM element by ID
func BindElement(elementId string) *ElementBinder {
	element := doc.Call("getElementById", elementId)
	return NewElementBinder(element)
}

// BindQuery creates an element binder for the first element matching a CSS selector
func BindQuery(selector string) *ElementBinder {
	element := doc.Call("querySelector", selector)
	return NewElementBinder(element)
}

// BindQueryAll creates element binders for all elements matching a CSS selector
func BindQueryAll(selector string) []*ElementBinder {
	elements := doc.Call("querySelectorAll", selector)
	length := elements.Get("length").Int()

	binders := make([]*ElementBinder, 0, length)
	for i := 0; i < length; i++ {
		element := elements.Index(i)
		if binder := NewElementBinder(element); binder != nil {
			binders = append(binders, binder)
		}
	}

	return binders
}

// CreateElement creates a new DOM element with optional attributes and returns a binder
func CreateElement(tagName string, attributes map[string]string) *ElementBinder {
	element := doc.Call("createElement", tagName)

	// Set attributes
	for name, value := range attributes {
		element.Call("setAttribute", name, value)
	}

	return NewElementBinder(element)
}

// ------------------------------------
// 📊 Binding Statistics
// ------------------------------------

// GetBindingStats returns statistics about active DOM bindings
func GetBindingStats() map[string]interface{} {
	renderer := getDOMRenderer()
	renderer.mutex.RLock()
	defer renderer.mutex.RUnlock()

	bindingTypes := make(map[string]int)
	for _, binding := range renderer.bindings {
		property := binding.property
		if strings.Contains(property, ":") {
			parts := strings.SplitN(property, ":", 2)
			bindingTypes[parts[0]]++
		} else {
			bindingTypes[property]++
		}
	}

	return map[string]interface{}{
		"totalBindings": len(renderer.bindings),
		"bindingTypes":  bindingTypes,
		"rendererStats": renderer.patcher.GetStats(),
	}
}
