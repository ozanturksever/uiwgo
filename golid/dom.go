// dom.go
// DOM utilities and reactive binding functionality

//go:build js && wasm

package golid

import (
	"fmt"
	"strings"
	"sync/atomic"
	"syscall/js"

	"github.com/google/uuid"
	"maragu.dev/gomponents"
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

// ------------------
// 🧱 DOM Utilities
// ------------------

// GenID generates a unique ID for DOM elements
func GenID() string {
	return fmt.Sprintf("e_%s", uuid.NewString())
}

// Append appends HTML content to a DOM element
func Append(html string, Element js.Value) {
	Element.Call("insertAdjacentHTML", "beforeend", html)
}

// NodeFromID retrieves a DOM node by its ID
func NodeFromID(id string) js.Value {
	return doc.Call("getElementById", id)
}

// BodyElement returns the document body element
func BodyElement() js.Value {
	return doc.Get("body")
}

// ------------------------
// 🖼  Reactive DOM Binding
// ------------------------

// Bind creates a reactive binding that updates when dependencies change using direct DOM manipulation
func Bind(fn func() Node) Node {
	id := GenID()
	placeholder := Span(Attr("id", id))

	// Register with global observer for initial setup
	globalObserver.RegisterElement(id, func() {
		elem := NodeFromID(id)
		if elem.Truthy() {
			// Create root for scoped reactivity
			_, cleanup := CreateRoot(func() interface{} {
				// Create effect that directly manipulates DOM
				CreateEffect(func() {
					newContent := fn()
					html := RenderHTML(Div(Attr("id", id), newContent))

					// Use DOM patcher for efficient updates
					getDOMRenderer().patcher.QueueOperation(DOMOperation{
						type_:   ReplaceNode,
						target:  elem.Get("parentNode"),
						newNode: createElementFromHTML(html),
						refNode: elem,
					})
				}, nil)
				return nil
			})

			// Register cleanup with observer
			globalObserver.RegisterDismountCallback(id, cleanup)
		}
	})

	return placeholder
}

// BindText creates a reactive text binding using direct DOM manipulation
func BindText(fn func() string) Node {
	id := GenID()
	span := Span(Attr("id", id), Text(fn()))

	// Register with global observer for initial setup
	globalObserver.RegisterElement(id, func() {
		elem := NodeFromID(id)
		if elem.Truthy() {
			// Create root for scoped reactivity
			_, cleanup := CreateRoot(func() interface{} {
				// Use the new reactive text binding
				BindTextReactive(elem, fn)
				return nil
			})

			// Register cleanup with observer
			globalObserver.RegisterDismountCallback(id, cleanup)
		}
	})

	return span
}

// BindReactive creates a reactive binding using the new DOM manipulation system
func BindReactive(element js.Value, fn func()) *DOMBinding {
	if !element.Truthy() {
		return nil
	}

	binding := &DOMBinding{
		id:       atomic.AddUint64(&signalIdCounter, 1),
		element:  element,
		property: "reactive",
		owner:    getCurrentOwner(),
	}

	// Create effect that runs the binding function
	binding.computation = CreateEffect(fn, binding.owner)

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

// ------------------
// Lists (Foreach())
// -------------------

// ForEach renders a list of items using a render function
func ForEach[T any](items []T, render func(T) Node) Node {
	var children []Node
	for _, item := range items {
		children = append(children, render(item))
	}
	return Group(children)
}

// ForEachSignal renders a reactive list using direct DOM manipulation
func ForEachSignal[T any](sig *Signal[[]T], render func(T) Node) Node {
	id := GenID()
	container := Span(Attr("id", id))

	// Register with global observer for initial setup
	globalObserver.RegisterElement(id, func() {
		elem := NodeFromID(id)
		if elem.Truthy() {
			// Create root for scoped reactivity
			_, cleanup := CreateRoot(func() interface{} {
				// Use the new list rendering system
				ListRender(elem, func() []T {
					return sig.Get()
				}, func(item T) string {
					// Generate a simple key - in practice, this should be more sophisticated
					return fmt.Sprintf("%v", item)
				}, func(item T) js.Value {
					html := RenderHTML(render(item))
					return createElementFromHTML(html)
				})
				return nil
			})

			// Register cleanup with observer
			globalObserver.RegisterDismountCallback(id, cleanup)
		}
	})

	return container
}

// ForEachReactive creates a reactive list binding with efficient DOM updates
func ForEachReactive[T any](container js.Value, items func() []T, keyFn func(T) string, renderFn func(T) js.Value) *DOMBinding {
	return ListRender(container, items, keyFn, renderFn)
}

// ----------------------
// 🧪 Rendering Utilities
// ----------------------

// RenderHTML renders a gomponents Node to HTML string
func RenderHTML(n gomponents.Node) string {
	var b strings.Builder
	err := n.Render(&b)
	if err != nil {
		return "<div>render error</div>"
	}
	return b.String()
}

// Render renders a Node to the document body using optimized DOM operations
func Render(n Node) {
	html := RenderHTML(n)

	// Use DOM patcher for efficient rendering
	getDOMRenderer().patcher.QueueOperation(DOMOperation{
		type_:   InsertNode,
		target:  BodyElement(),
		newNode: createElementFromHTML(html),
	})

	// Flush immediately for initial render
	getDOMRenderer().patcher.Flush()
}

// RenderTo renders a Node to a specific container element
func RenderTo(n Node, container js.Value) {
	if !container.Truthy() {
		return
	}

	html := RenderHTML(n)

	// Use DOM patcher for efficient rendering
	getDOMRenderer().patcher.QueueOperation(DOMOperation{
		type_:   InsertNode,
		target:  container,
		newNode: createElementFromHTML(html),
	})

	// Flush immediately for initial render
	getDOMRenderer().patcher.Flush()
}

// ------------------
// 🛠 Callback Helper
// ------------------

// Callback creates a JavaScript callback from a Go function
func Callback(f func()) JsCallback {
	return func(this js.Value, args []js.Value) interface{} {
		f()
		return nil
	}
}

// --------------
// 🧭 App Entrypoint
// --------------

// Run starts the application main loop with reactive system initialization
func Run() {
	// Initialize the reactive DOM system
	getDOMRenderer()

	// Start the scheduler
	getScheduler()

	// Keep the application running
	select {}
}

// ------------------------------------
// 🔧 DOM Utilities for Reactive System
// ------------------------------------

// createElementFromHTML creates a DOM element from HTML string
func createElementFromHTML(html string) js.Value {
	container := doc.Call("createElement", "div")
	container.Set("innerHTML", html)

	children := container.Get("children")
	if children.Get("length").Int() == 1 {
		return children.Index(0)
	}

	// Multiple elements - return document fragment
	fragment := doc.Call("createDocumentFragment")
	length := children.Get("length").Int()

	for i := 0; i < length; i++ {
		fragment.Call("appendChild", children.Index(i))
	}

	return fragment
}

// GetElementBinder creates an element binder for reactive DOM manipulation
func GetElementBinder(elementId string) *ElementBinder {
	return BindElement(elementId)
}

// QueryBinder creates an element binder using CSS selector
func QueryBinder(selector string) *ElementBinder {
	return BindQuery(selector)
}

// CreateReactiveElement creates a new element with reactive bindings
func CreateReactiveElement(tagName string, attributes map[string]string) *ElementBinder {
	return CreateElement(tagName, attributes)
}
