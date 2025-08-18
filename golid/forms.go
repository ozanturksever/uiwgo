// forms.go
// Input binding and form handling functionality

package golid

import (
	"syscall/js"

	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

// -----------
// Form Inputs
// -----------

// BindInput creates a two-way binding between a signal and an input element
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

// BindInputWithType creates a two-way binding with custom input type
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

// BindInputWithFocus creates a two-way binding with focus state tracking
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
