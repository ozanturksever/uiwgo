// debug.go
// Debugging utilities for development and troubleshooting

package golid

import (
	"fmt"
	"syscall/js"
)

// ------------------
// 🪵 Debugging
// ------------------

// Log outputs values to the browser console
func Log(args ...interface{}) {
	js.Global().Get("console").Call("log", args...)
}

// Logf outputs formatted messages to the browser console
func Logf(format string, args ...interface{}) {
	js.Global().Get("console").Call("log", fmt.Sprintf(format, args...))
}
