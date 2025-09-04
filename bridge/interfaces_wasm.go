//go:build js && wasm

package bridge

import (
	"syscall/js"
)

// JSValue represents a JavaScript value with type-safe operations
type JSValue interface {
	// Type returns the JavaScript type of the value
	Type() js.Type
	// String returns the string representation
	String() string
	// Int returns the integer value
	Int() int
	// Float returns the float64 value
	Float() float64
	// Bool returns the boolean value
	Bool() bool
	// IsNull returns true if the value is null
	IsNull() bool
	// IsUndefined returns true if the value is undefined
	IsUndefined() bool
	// Get retrieves a property by name
	Get(key string) JSValue
	// Set sets a property by name
	Set(key string, value interface{})
	// Call calls the value as a function
	Call(method string, args ...interface{}) JSValue
	// Index retrieves an array element by index
	Index(i int) JSValue
	// Length returns the length property
	Length() int
	// Raw returns the underlying js.Value (for advanced use)
	Raw() js.Value
}

// JSBridge provides typed wrapper over JavaScript values
type JSBridge interface {
	// Global returns the global object (window in browsers)
	Global() JSValue
	// Undefined returns an undefined value
	Undefined() JSValue
	// Null returns a null value
	Null() JSValue
	// ValueOf converts a Go value to JSValue
	ValueOf(x interface{}) JSValue
	// FuncOf creates a JavaScript function from a Go function
	FuncOf(fn func(this JSValue, args []JSValue) interface{}) JSValue
}