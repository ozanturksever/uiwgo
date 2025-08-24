//go:build !js || !wasm

package logutil

import (
	"fmt"
)

// Log prints the given arguments to the standard output in non-JS builds.
// It is safe to call with any mix of Go values.
func Log(args ...any) {
	fmt.Println(args...)
}

// Logf formats according to a format specifier and prints to the standard output
// in non-JS builds. It is safe to call with any mix of Go values.
func Logf(format string, args ...any) {
	fmt.Printf(format, args...)
}
