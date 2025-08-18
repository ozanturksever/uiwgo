// events.go
// Event binding helpers for DOM interactions

package golid

import (
	"syscall/js"

	. "maragu.dev/gomponents"
)

// -------------------------
// 🧩 Event Binding Helpers
// -------------------------

// OnClick creates a click event handler
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

// OnInput creates an input event handler
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
