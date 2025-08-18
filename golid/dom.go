// dom.go
// DOM utilities and reactive binding functionality

package golid

import (
	"fmt"
	"strings"
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

// Bind creates a reactive binding that updates when dependencies change
func Bind(fn func() Node) Node {
	id := GenID()
	placeholder := Span(Attr("id", id))

	// Register with global observer instead of polling
	globalObserver.RegisterElement(id, func() {
		elem := NodeFromID(id)
		if elem.Truthy() {
			Watch(func() {
				html := RenderHTML(Div(Attr("id", id), fn()))
				elem := NodeFromID(id)
				if elem.Truthy() {
					elem.Set("outerHTML", html)
				}
			})
		}
	})

	return placeholder
}

// BindText creates a reactive text binding that updates when dependencies change
func BindText(fn func() string) Node {
	id := GenID()
	span := Span(Attr("id", id), Text(fn()))

	// Register with global observer instead of individual MutationObserver
	globalObserver.RegisterElement(id, func() {
		elem := NodeFromID(id)
		if elem.Truthy() {
			// Set up reactive Watch effect
			Watch(func() {
				elem := NodeFromID(id)
				if elem.Truthy() {
					currentVal := elem.Get("textContent").String()
					newVal := fn()
					if currentVal != newVal {
						elem.Set("textContent", newVal)
					}
				}
			})
		}
	})

	return span
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

// ForEachSignal renders a reactive list based on a signal
func ForEachSignal[T any](sig *Signal[[]T], render func(T) Node) Node {
	return Bind(func() Node {
		items := sig.Get()
		var children []Node
		for _, item := range items {
			children = append(children, render(item))
		}
		return Group(children)
	})
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

// Render renders a Node to the document body
func Render(n Node) {
	Append(RenderHTML(n), BodyElement())
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

// Run starts the application main loop
func Run() {
	select {}
}
