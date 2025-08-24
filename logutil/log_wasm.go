//go:build js && wasm

package logutil

import (
	"fmt"
	"syscall/js"
)

// toJSArg converts a Go value into something safe to pass to JS console methods.
// It prefers passing JS values through as-is, and falls back to strings for unsupported types.
func toJSArg(v any) any {
	switch x := v.(type) {
	case nil:
		return nil
	case js.Value:
		return x
	default:
		// Support wrapper-like values without depending on js.Wrapper symbol directly.
		type jsWrapper interface{ JSValue() js.Value }
		if w, ok := any(x).(jsWrapper); ok {
			return w.JSValue()
		}
		switch x := v.(type) {
		case string, bool,
			int, int8, int16, int32, int64,
			uint, uint8, uint16, uint32, uint64,
			float32, float64:
			// Primitive types are fine to pass directly; JS bridge will convert them.
			return x
		default:
			// Avoid js.ValueOf on arbitrary structs, maps, funcs, etc., which may panic.
			// Provide a readable representation instead.
			return fmt.Sprintf("%v", x)
		}
	}
}

// Log prints the given arguments to the browser console when running under js/wasm.
// It is safe to call with any mix of Go and JS values (js.Value, js.Wrapper, primitives, etc.).
// If the console is not available, it falls back to fmt.Println.
func Log(args ...any) {
	g := js.Global()
	if g.Truthy() {
		c := g.Get("console")
		if c.Truthy() {
			converted := make([]any, 0, len(args))
			for _, a := range args {
				converted = append(converted, toJSArg(a))
			}
			c.Call("log", converted...)
			return
		}
	}
	// Fallback
	fmt.Println(args...)
}

// Logf formats according to a format specifier and logs to the browser console when available.
// When console is not present, it falls back to fmt.Printf.
func Logf(format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	g := js.Global()
	if g.Truthy() {
		c := g.Get("console")
		if c.Truthy() {
			c.Call("log", msg)
			return
		}
	}
	fmt.Print(msg)
}
