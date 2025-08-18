//go:build js && wasm

// events.go
// Event binding helpers for DOM interactions with new event system integration

package golid

import (
	"syscall/js"
	"time"

	. "maragu.dev/gomponents"
)

// -------------------------
// 🧩 Legacy Event Binding Helpers (Compatibility)
// -------------------------

// OnClick creates a click event handler using the legacy approach
// Deprecated: Use OnClickV2 for better performance and automatic cleanup
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

// OnInput creates an input event handler using the legacy approach
// Deprecated: Use OnInputV2 for better performance and automatic cleanup
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

// -------------------------
// 🚀 New Event System Integration
// -------------------------

// OnClickV2 creates a click event handler using the new event system with automatic cleanup
func OnClickV2(f func()) Node {
	id := GenID()

	// Register element for event binding after DOM insertion
	globalObserver.RegisterElement(id, func() {
		elem := NodeFromID(id)
		if elem.Truthy() {
			// Use new event system with delegation and automatic cleanup
			Subscribe(elem, "click", func(e js.Value) {
				f()
			}, EventOptions{
				Delegate: true,
				Priority: UserBlocking,
			})
		}
	})

	return Attr("id", id)
}

// OnInputV2 creates an input event handler using the new event system with automatic cleanup
func OnInputV2(handler func(string)) Node {
	id := GenID()

	// Register element for event binding after DOM insertion
	globalObserver.RegisterElement(id, func() {
		elem := NodeFromID(id)
		if elem.Truthy() {
			// Use new event system with delegation and automatic cleanup
			Subscribe(elem, "input", func(e js.Value) {
				value := e.Get("target").Get("value").String()
				handler(value)
			}, EventOptions{
				Delegate: true,
				Priority: UserBlocking,
				Debounce: 0, // No debouncing by default for input
			})
		}
	})

	return Attr("id", id)
}

// OnChangeV2 creates a change event handler using the new event system
func OnChangeV2(handler func(string)) Node {
	id := GenID()

	globalObserver.RegisterElement(id, func() {
		elem := NodeFromID(id)
		if elem.Truthy() {
			Subscribe(elem, "change", func(e js.Value) {
				value := e.Get("target").Get("value").String()
				handler(value)
			}, EventOptions{
				Delegate: true,
				Priority: UserBlocking,
			})
		}
	})

	return Attr("id", id)
}

// OnSubmitV2 creates a form submit event handler using the new event system
func OnSubmitV2(handler func(js.Value)) Node {
	id := GenID()

	globalObserver.RegisterElement(id, func() {
		elem := NodeFromID(id)
		if elem.Truthy() {
			Subscribe(elem, "submit", func(e js.Value) {
				e.Call("preventDefault") // Prevent default form submission
				handler(e)
			}, EventOptions{
				Delegate: true,
				Priority: UserBlocking,
			})
		}
	})

	return Attr("id", id)
}

// OnKeyDownV2 creates a keydown event handler using the new event system
func OnKeyDownV2(handler func(string)) Node {
	id := GenID()

	globalObserver.RegisterElement(id, func() {
		elem := NodeFromID(id)
		if elem.Truthy() {
			Subscribe(elem, "keydown", func(e js.Value) {
				key := e.Get("key").String()
				handler(key)
			}, EventOptions{
				Delegate: true,
				Priority: UserBlocking,
			})
		}
	})

	return Attr("id", id)
}

// OnMouseEnterV2 creates a mouseenter event handler using the new event system
func OnMouseEnterV2(handler func()) Node {
	id := GenID()

	globalObserver.RegisterElement(id, func() {
		elem := NodeFromID(id)
		if elem.Truthy() {
			Subscribe(elem, "mouseenter", func(e js.Value) {
				handler()
			}, EventOptions{
				Delegate: false, // mouseenter doesn't bubble, use direct binding
				Priority: Normal,
			})
		}
	})

	return Attr("id", id)
}

// OnMouseLeaveV2 creates a mouseleave event handler using the new event system
func OnMouseLeaveV2(handler func()) Node {
	id := GenID()

	globalObserver.RegisterElement(id, func() {
		elem := NodeFromID(id)
		if elem.Truthy() {
			Subscribe(elem, "mouseleave", func(e js.Value) {
				handler()
			}, EventOptions{
				Delegate: false, // mouseleave doesn't bubble, use direct binding
				Priority: Normal,
			})
		}
	})

	return Attr("id", id)
}

// -------------------------
// 🎯 Reactive Event Helpers
// -------------------------

// OnClickReactive creates a reactive click event handler that integrates with signals
func OnClickReactive(f func()) Node {
	id := GenID()

	globalObserver.RegisterElement(id, func() {
		elem := NodeFromID(id)
		if elem.Truthy() {
			// Use reactive event subscription
			SubscribeReactive(elem, "click", func(e js.Value) {
				f()
			}, EventOptions{
				Delegate: true,
				Priority: UserBlocking,
			})
		}
	})

	return Attr("id", id)
}

// OnInputReactive creates a reactive input event handler that integrates with signals
func OnInputReactive(handler func(string)) Node {
	id := GenID()

	globalObserver.RegisterElement(id, func() {
		elem := NodeFromID(id)
		if elem.Truthy() {
			SubscribeReactive(elem, "input", func(e js.Value) {
				value := e.Get("target").Get("value").String()
				handler(value)
			}, EventOptions{
				Delegate: true,
				Priority: UserBlocking,
			})
		}
	})

	return Attr("id", id)
}

// -------------------------
// 🔧 Advanced Event Helpers
// -------------------------

// OnEventWithOptions creates a custom event handler with full options control
func OnEventWithOptions(event string, handler func(js.Value), options EventOptions) Node {
	id := GenID()

	globalObserver.RegisterElement(id, func() {
		elem := NodeFromID(id)
		if elem.Truthy() {
			Subscribe(elem, event, handler, options)
		}
	})

	return Attr("id", id)
}

// OnEventDebounced creates a debounced event handler
func OnEventDebounced(event string, handler func(js.Value), debounceMs int) Node {
	return OnEventWithOptions(event, handler, EventOptions{
		Delegate: true,
		Priority: Normal,
		Debounce: time.Duration(debounceMs) * time.Millisecond,
	})
}

// OnEventThrottled creates a throttled event handler
func OnEventThrottled(event string, handler func(js.Value), throttleMs int) Node {
	return OnEventWithOptions(event, handler, EventOptions{
		Delegate: true,
		Priority: Normal,
		Throttle: time.Duration(throttleMs) * time.Millisecond,
	})
}

// -------------------------
// 🧹 Event System Utilities
// -------------------------

// GetEventSystemStats returns statistics about the event system
func GetEventSystemStats() map[string]interface{} {
	manager := GetEventManager()
	return map[string]interface{}{
		"subscriptions": manager.GetSubscriptionCount(),
		"metrics":       manager.GetMetrics().GetStats(),
		"delegator":     manager.delegator.GetStats(),
		"batcher":       manager.batcher.GetStats(),
		"customBus":     manager.customBus.GetStats(),
	}
}

// CleanupEventSystem manually triggers cleanup of the event system
func CleanupEventSystem() {
	GetEventManager().Dispose()
}
