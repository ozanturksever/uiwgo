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

	var check js.Func
	check = js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		elem := NodeFromID(id)
		if elem.Truthy() {
			Watch(func() {
				html := RenderHTML(Div(Attr("id", id), fn()))
				elem := NodeFromID(id)
				if elem.Truthy() {
					elem.Set("outerHTML", html)
				}
			})
			return nil
		}
		js.Global().Call("setTimeout", check, 10)
		return nil
	})
	js.Global().Call("setTimeout", check, 10)

	return placeholder
}

func BindText(fn func() string) Node {
	id := GenID()
	span := Span(Attr("id", id), Text(fn()))

	// MutationObserver-based approach for better performance
	go func() {
		// Create a callback function for when the element is found
		var onElementAdded js.Func
		onElementAdded = js.FuncOf(func(this js.Value, args []js.Value) interface{} {
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
				onElementAdded.Release()
			}
			return nil
		})

		// Create MutationObserver to watch for DOM additions
		var observer js.Value
		var observerCallback js.Func
		observerCallback = js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			mutations := args[0]
			mutationsLength := mutations.Get("length").Int()

			for i := 0; i < mutationsLength; i++ {
				mutation := mutations.Index(i)
				if mutation.Get("type").String() == "childList" {
					addedNodes := mutation.Get("addedNodes")
					addedNodesLength := addedNodes.Get("length").Int()

					for j := 0; j < addedNodesLength; j++ {
						node := addedNodes.Index(j)
						// Check if this node or any of its descendants is our target element
						if checkNodeForTarget(node, id) {
							onElementAdded.Invoke()
							observer.Call("disconnect")
							observerCallback.Release()
							return nil
						}
					}
				}
			}
			return nil
		})

		// Check if element already exists (immediate check)
		if NodeFromID(id).Truthy() {
			onElementAdded.Invoke()
			observerCallback.Release()
			return
		}

		// Create and configure the MutationObserver
		observer = js.Global().Get("MutationObserver").New(observerCallback)

		// Start observing the document body for child additions
		config := js.Global().Get("Object").New()
		config.Set("childList", true)
		config.Set("subtree", true)

		observer.Call("observe", doc.Get("body"), config)
	}()

	return span
}

// Helper function to check if a node or its descendants contains our target element
func checkNodeForTarget(node js.Value, targetID string) bool {
	// Check if the node itself has the target ID
	if node.Get("nodeType").Int() == 1 { // Element node
		if node.Get("id").String() == targetID {
			return true
		}

		// Check descendants using querySelector
		found := node.Call("querySelector", "#"+targetID)
		return found.Truthy()
	}
	return false
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
	return uuid.NewString()
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

	// Poll until the element exists, then attach listeners and Watch
	var check js.Func
	check = js.FuncOf(func(this js.Value, args []js.Value) interface{} {
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

			// release the check function to avoid leaking resources
			check.Release()
			return nil
		}

		// not present yet, retry shortly
		js.Global().Call("setTimeout", check, 10)
		return nil
	})

	js.Global().Call("setTimeout", check, 10)

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

	var check js.Func
	check = js.FuncOf(func(this js.Value, args []js.Value) interface{} {
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

			// release the check function to avoid leaking resources
			check.Release()
			return nil
		}

		// not present yet, retry shortly
		js.Global().Call("setTimeout", check, 10)
		return nil
	})
	js.Global().Call("setTimeout", check, 10)

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

	go func() {
		js.Global().Call("setTimeout", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
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
			return nil
		}), 0)
	}()

	return Input(
		Attr("id", id),
		Type("text"),
		Placeholder(placeholder),
		Value(sig.Get()),
	)
}
