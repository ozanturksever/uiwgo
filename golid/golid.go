// golid.go
// A minimal reactive UI toolkit for Go+WASM using gomponents

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

// ---------------------------
// 🔧 Core Types & Interfaces
// ---------------------------

type JsCallback func(this js.Value, args []js.Value) interface{}

var doc = js.Global().Get("document")

// ---------------------------
// 🔍 Global MutationObserver System
// ---------------------------

type ElementCallback func()

type ObserverManager struct {
	observer    js.Value
	callbacks   map[string]ElementCallback
	isObserving bool
}

var globalObserver *ObserverManager

func init() {
	globalObserver = &ObserverManager{
		callbacks: make(map[string]ElementCallback),
	}
}

// RegisterElement registers an element ID with a callback to be executed when the element is found
func (om *ObserverManager) RegisterElement(id string, callback ElementCallback) {
	// Check if element already exists
	if NodeFromID(id).Truthy() {
		callback()
		return
	}

	om.callbacks[id] = callback

	// Start observing if not already observing
	if !om.isObserving {
		om.startObserving()
	}
}

// UnregisterElement removes an element from tracking
func (om *ObserverManager) UnregisterElement(id string) {
	delete(om.callbacks, id)

	// Stop observing if no more callbacks
	if len(om.callbacks) == 0 && om.isObserving {
		om.stopObserving()
	}
}

// startObserving initializes the MutationObserver
func (om *ObserverManager) startObserving() {
	if om.isObserving {
		return
	}

	observerCallback := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		mutations := args[0]
		mutationsLength := mutations.Get("length").Int()

		for i := 0; i < mutationsLength; i++ {
			mutation := mutations.Index(i)
			if mutation.Get("type").String() == "childList" {
				addedNodes := mutation.Get("addedNodes")
				addedNodesLength := addedNodes.Get("length").Int()

				for j := 0; j < addedNodesLength; j++ {
					node := addedNodes.Index(j)
					om.checkNodeForTargets(node)
				}
			}
		}
		return nil
	})

	om.observer = js.Global().Get("MutationObserver").New(observerCallback)

	config := js.Global().Get("Object").New()
	config.Set("childList", true)
	config.Set("subtree", true)

	om.observer.Call("observe", doc.Get("body"), config)
	om.isObserving = true
}

// stopObserving disconnects the MutationObserver
func (om *ObserverManager) stopObserving() {
	if om.isObserving && om.observer.Truthy() {
		om.observer.Call("disconnect")
		om.isObserving = false
	}
}

// checkNodeForTargets checks if any registered elements are found in the added node
func (om *ObserverManager) checkNodeForTargets(node js.Value) {
	if node.Get("nodeType").Int() != 1 { // Not an element node
		return
	}

	var foundIDs []string

	for id, callback := range om.callbacks {
		if node.Get("id").String() == id {
			foundIDs = append(foundIDs, id)
			callback()
		} else {
			// Check descendants using getElementById instead of querySelector
			// to avoid CSS selector syntax issues with UUIDs starting with digits
			found := doc.Call("getElementById", id)
			if found.Truthy() {
				// Verify that the found element is actually a descendant of the node
				if isDescendantOf(found, node) {
					foundIDs = append(foundIDs, id)
					callback()
				}
			}
		}
	}

	// Remove found elements from callbacks
	for _, id := range foundIDs {
		om.UnregisterElement(id)
	}
}

// Helper function to check if element is a descendant of node
func isDescendantOf(element js.Value, ancestor js.Value) bool {
	current := element.Get("parentNode")
	for current.Truthy() {
		if current.Equal(ancestor) {
			return true
		}
		current = current.Get("parentNode")
	}
	return false
}

// ------------------------------------
// 📦 Reactive Signals (State Handling)
// ------------------------------------

type effect struct {
	fn   func()
	deps map[any]struct{}
}

var currentEffect *effect

type hasWatchers interface {
	removeWatcher(*effect)
}

type Signal[T any] struct {
	value    T
	watchers map[*effect]struct{}
}

func NewSignal[T any](initial T) *Signal[T] {
	return &Signal[T]{
		value:    initial,
		watchers: make(map[*effect]struct{}),
	}
}

func (s *Signal[T]) Get() T {
	if currentEffect != nil {
		s.watchers[currentEffect] = struct{}{}
		currentEffect.deps[s] = struct{}{}
	}
	return s.value
}

func (s *Signal[T]) Set(val T) {
	s.value = val
	for e := range s.watchers {
		go runEffect(e)
	}
}

func (s *Signal[T]) removeWatcher(e *effect) {
	delete(s.watchers, e)
}

func Watch(fn func()) {
	e := &effect{
		fn:   fn,
		deps: make(map[any]struct{}),
	}
	runEffect(e)
}

func runEffect(e *effect) {
	for dep := range e.deps {
		if s, ok := dep.(hasWatchers); ok {
			s.removeWatcher(e)
		}
	}
	e.deps = make(map[any]struct{})
	currentEffect = e
	e.fn()
	currentEffect = nil
}

// ------------------------
// 🖼  Reactive DOM Binding
// ------------------------

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

// -------------------------
// 🧩 Event Binding Helpers
// -------------------------

func OnClick(f func()) Node {
	id := GenID()
	go func() {
		js.Global().Call("setTimeout", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			elem := NodeFromID(id)
			if elem.Truthy() {
				elem.Call("addEventListener", "click", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
					f()
					return nil
				}))
			}
			return nil
		}), 0)
	}()

	return Attr("id", id)
}

// ------------------
// 🧱 DOM Utilities
// ------------------

func GenID() string {
	return fmt.Sprintf("e_%s", uuid.NewString())
}

func Append(html string, Element js.Value) {
	Element.Call("insertAdjacentHTML", "beforeend", html)
}

func NodeFromID(id string) js.Value {
	return doc.Call("getElementById", id)
}

func BodyElement() js.Value {
	return doc.Get("body")
}

// ----------------------
// 🧪 Rendering Utilities
// ----------------------

func RenderHTML(n gomponents.Node) string {
	var b strings.Builder
	err := n.Render(&b)
	if err != nil {
		return "<div>render error</div>"
	}
	return b.String()
}

func Render(n Node) {
	Append(RenderHTML(n), BodyElement())
}

// ------------------
// 🛠 Callback Helper
// ------------------

func Callback(f func()) JsCallback {
	return func(this js.Value, args []js.Value) interface{} {
		f()
		return nil
	}
}

// --------------
// 🧭 App Entrypoint
// --------------

func Run() {
	select {}
}

// ------------------
// 🪵 Debugging
// ------------------

func Log(args ...interface{}) {
	js.Global().Get("console").Call("log", args...)
}

func Logf(format string, args ...interface{}) {
	js.Global().Get("console").Call("log", fmt.Sprintf(format, args...))
}

// ------------------
// Lists (Foreach())
// -------------------

func ForEach[T any](items []T, render func(T) Node) Node {
	var children []Node
	for _, item := range items {
		children = append(children, render(item))
	}
	return Group(children)
}

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

// -----------
// text inputs
// -----------

func OnInput(handler func(string)) Node {
	id := GenID()
	go func() {
		js.Global().Call("setTimeout", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			elem := NodeFromID(id)
			if elem.Truthy() {
				elem.Call("addEventListener", "input", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
					value := this.Get("value").String()
					handler(value)
					return nil
				}))
			}
			return nil
		}), 0)
	}()
	return Attr("id", id)
}

func BindInput(sig *Signal[string], placeholder string) Node {
	id := GenID()
	isComposing := false

	// Register with global observer instead of polling
	globalObserver.RegisterElement(id, func() {
		elem := NodeFromID(id)
		if elem.Truthy() {
			// attach listeners
			elem.Call("addEventListener", "input", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
				if !isComposing {
					val := this.Get("value").String()
					if val != sig.Get() {
						sig.Set(val)
					}
				}
				return nil
			}))

			elem.Call("addEventListener", "compositionstart", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
				isComposing = true
				return nil
			}))

			elem.Call("addEventListener", "compositionend", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
				isComposing = false
				val := this.Get("value").String()
				if val != sig.Get() {
					sig.Set(val)
				}
				return nil
			}))

			elem.Call("addEventListener", "paste", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
				js.Global().Call("setTimeout", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
					val := elem.Get("value").String()
					if val != sig.Get() {
						sig.Set(val)
					}
					return nil
				}), 0)
				return nil
			}))

			// Now set up the Watch effect after the element is ready
			Watch(func() {
				elem := NodeFromID(id)
				if elem.Truthy() {
					signalVal := sig.Get()
					elemVal := elem.Get("value").String()
					if elemVal != signalVal {
						selectionStart := elem.Get("selectionStart")
						selectionEnd := elem.Get("selectionEnd")

						elem.Set("value", signalVal)

						if doc.Get("activeElement").Equal(elem) {
							if selectionStart.Truthy() && selectionEnd.Truthy() {
								start := selectionStart.Int()
								end := selectionEnd.Int()
								maxPos := len(signalVal)

								if start > maxPos {
									start = maxPos
								}
								if end > maxPos {
									end = maxPos
								}

								elem.Call("setSelectionRange", start, end)
							}
						}
					}
				}
			})
		}
	})

	return Input(
		Attr("id", id),
		Type("text"),
		Placeholder(placeholder),
		Value(sig.Get()), // initial value
	)
}

// Enhanced input binding with type support
func BindInputWithType(sig *Signal[string], inputType, placeholder string) Node {
	id := GenID()
	isComposing := false

	// Register with global observer instead of polling
	globalObserver.RegisterElement(id, func() {
		elem := NodeFromID(id)
		if elem.Truthy() {
			// attach listeners
			elem.Call("addEventListener", "input", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
				if !isComposing {
					val := this.Get("value").String()
					if val != sig.Get() {
						sig.Set(val)
					}
				}
				return nil
			}))

			elem.Call("addEventListener", "compositionstart", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
				isComposing = true
				return nil
			}))

			elem.Call("addEventListener", "compositionend", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
				isComposing = false
				val := this.Get("value").String()
				if val != sig.Get() {
					sig.Set(val)
				}
				return nil
			}))

			elem.Call("addEventListener", "paste", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
				js.Global().Call("setTimeout", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
					val := elem.Get("value").String()
					if val != sig.Get() {
						sig.Set(val)
					}
					return nil
				}), 0)
				return nil
			}))

			// Now set up the Watch effect after the element is ready
			Watch(func() {
				elem := NodeFromID(id)
				if elem.Truthy() {
					signalVal := sig.Get()
					elemVal := elem.Get("value").String()
					if elemVal != signalVal {
						selectionStart := elem.Get("selectionStart")
						selectionEnd := elem.Get("selectionEnd")

						elem.Set("value", signalVal)

						if doc.Get("activeElement").Equal(elem) {
							if selectionStart.Truthy() && selectionEnd.Truthy() {
								start := selectionStart.Int()
								end := selectionEnd.Int()
								maxPos := len(signalVal)

								if start > maxPos {
									start = maxPos
								}
								if end > maxPos {
									end = maxPos
								}

								elem.Call("setSelectionRange", start, end)
							}
						}
					}
				}
			})
		}
	})

	return Input(
		Attr("id", id),
		Type(inputType),
		Placeholder(placeholder),
		Value(sig.Get()),
	)
}

// Bind input with focus state tracking
func BindInputWithFocus(sig *Signal[string], focusSig *Signal[bool], placeholder string) Node {
	id := GenID()
	isComposing := false

	// Register with global observer instead of polling
	globalObserver.RegisterElement(id, func() {
		elem := NodeFromID(id)
		if elem.Truthy() {
			// Input handling
			elem.Call("addEventListener", "input", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
				if !isComposing {
					val := elem.Get("value").String()
					if val != sig.Get() {
						sig.Set(val)
					}
				}
				return nil
			}))

			// Composition handling
			elem.Call("addEventListener", "compositionstart", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
				isComposing = true
				return nil
			}))

			elem.Call("addEventListener", "compositionend", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
				isComposing = false
				val := elem.Get("value").String()
				if val != sig.Get() {
					sig.Set(val)
				}
				return nil
			}))

			// Focus/blur handling
			elem.Call("addEventListener", "focus", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
				focusSig.Set(true)
				return nil
			}))

			elem.Call("addEventListener", "blur", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
				focusSig.Set(false)
				return nil
			}))

			// Paste handling
			elem.Call("addEventListener", "paste", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
				js.Global().Call("setTimeout", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
					val := elem.Get("value").String()
					if val != sig.Get() {
						sig.Set(val)
					}
					return nil
				}), 0)
				return nil
			}))

			// Now set up the Watch effect after the element is ready
			Watch(func() {
				elem := NodeFromID(id)
				if elem.Truthy() {
					signalVal := sig.Get()
					elemVal := elem.Get("value").String()
					if elemVal != signalVal {
						selectionStart := elem.Get("selectionStart")
						selectionEnd := elem.Get("selectionEnd")

						elem.Set("value", signalVal)

						if doc.Get("activeElement").Equal(elem) {
							if selectionStart.Truthy() && selectionEnd.Truthy() {
								start := selectionStart.Int()
								end := selectionEnd.Int()
								maxPos := len(signalVal)

								if start > maxPos {
									start = maxPos
								}
								if end > maxPos {
									end = maxPos
								}

								elem.Call("setSelectionRange", start, end)
							}
						}
					}
				}
			})
		}
	})

	return Input(
		Attr("id", id),
		Type("text"),
		Placeholder(placeholder),
		Value(sig.Get()),
	)
}
