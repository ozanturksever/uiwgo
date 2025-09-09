//go:build js && wasm

package react

import (
	"sync"
	"sync/atomic"
	"syscall/js"

	"github.com/ozanturksever/logutil"
)

// CallbackID represents a unique identifier for a callback
type CallbackID string

// CallbackFunc represents a Go callback function that can be invoked from JavaScript
type CallbackFunc func(args ...interface{}) interface{}

// CallbackRegistry manages Go callbacks that can be invoked from JavaScript
type CallbackRegistry struct {
	mu        sync.RWMutex
	callbacks map[CallbackID]CallbackFunc
	counter   int64
}

// Global callback registry instance
var globalCallbackRegistry = &CallbackRegistry{
	callbacks: make(map[CallbackID]CallbackFunc),
}

// NewCallbackRegistry creates a new callback registry
func NewCallbackRegistry() *CallbackRegistry {
	return &CallbackRegistry{
		callbacks: make(map[CallbackID]CallbackFunc),
	}
}

// Register adds a callback function and returns its unique ID
func (cr *CallbackRegistry) Register(fn CallbackFunc) CallbackID {
	cr.mu.Lock()
	defer cr.mu.Unlock()

	// Generate unique callback ID
	id := CallbackID("cb_" + string(rune(atomic.AddInt64(&cr.counter, 1))))
	cr.callbacks[id] = fn

	logutil.Logf("Registered callback with ID: %s", id)
	return id
}

// Unregister removes a callback by its ID
func (cr *CallbackRegistry) Unregister(id CallbackID) bool {
	cr.mu.Lock()
	defer cr.mu.Unlock()

	if _, exists := cr.callbacks[id]; exists {
		delete(cr.callbacks, id)
		logutil.Logf("Unregistered callback with ID: %s", id)
		return true
	}
	return false
}

// Invoke calls a callback by its ID with the provided arguments
func (cr *CallbackRegistry) Invoke(id CallbackID, args ...interface{}) (interface{}, error) {
	cr.mu.RLock()
	fn, exists := cr.callbacks[id]
	cr.mu.RUnlock()

	if !exists {
		return nil, &ReactBridgeError{
			Type:    ErrorTypeCallbackNotFound,
			Message: "callback not found: " + string(id),
		}
	}

	// Execute callback with error recovery
	defer func() {
		if r := recover(); r != nil {
			logutil.Logf("Callback %s panicked: %v", id, r)
		}
	}()

	result := fn(args...)
	logutil.Logf("Invoked callback %s successfully", id)
	return result, nil
}

// GetCallbackCount returns the number of registered callbacks
func (cr *CallbackRegistry) GetCallbackCount() int {
	cr.mu.RLock()
	defer cr.mu.RUnlock()
	return len(cr.callbacks)
}

// Clear removes all registered callbacks
func (cr *CallbackRegistry) Clear() {
	cr.mu.Lock()
	defer cr.mu.Unlock()
	cr.callbacks = make(map[CallbackID]CallbackFunc)
	logutil.Log("Cleared all callbacks from registry")
}

// Global callback registry functions

// RegisterCallback registers a callback in the global registry
func RegisterCallback(fn CallbackFunc) CallbackID {
	return globalCallbackRegistry.Register(fn)
}

// UnregisterCallback removes a callback from the global registry
func UnregisterCallback(id CallbackID) bool {
	return globalCallbackRegistry.Unregister(id)
}

// InvokeCallback invokes a callback from the global registry
func InvokeCallback(id CallbackID, args ...interface{}) (interface{}, error) {
	return globalCallbackRegistry.Invoke(id, args...)
}

// GetGlobalCallbackCount returns the number of callbacks in the global registry
func GetGlobalCallbackCount() int {
	return globalCallbackRegistry.GetCallbackCount()
}

// ClearGlobalCallbacks clears all callbacks from the global registry
func ClearGlobalCallbacks() {
	globalCallbackRegistry.Clear()
}

// JavaScript integration functions

// setupCallbackIntegration sets up the JavaScript integration for callbacks
func setupCallbackIntegration() {
	// Create the global ReactCompatCallbacks object if it doesn't exist
	if js.Global().Get("ReactCompatCallbacks").IsUndefined() {
		js.Global().Set("ReactCompatCallbacks", js.Global().Get("Object").New())
	}

	// Set up the invoke function that JavaScript can call
	invokeFunc := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		if len(args) < 1 {
			logutil.Log("ReactCompatCallbacks.invoke called with insufficient arguments")
			return nil
		}

		callbackID := CallbackID(args[0].String())

		// Convert JS arguments to Go interface{} slice
		goArgs := make([]interface{}, len(args)-1)
		for i := 1; i < len(args); i++ {
			goArgs[i-1] = convertJSValueToGo(args[i])
		}

		// Invoke the callback
		result, err := InvokeCallback(callbackID, goArgs...)
		if err != nil {
			logutil.Logf("Failed to invoke callback %s: %v", callbackID, err)
			return nil
		}

		// Convert result back to JS value
		return convertGoValueToJS(result)
	})

	js.Global().Get("ReactCompatCallbacks").Set("invoke", invokeFunc)
	logutil.Log("Callback integration setup completed")
}

// Helper functions for JS/Go value conversion

// convertJSValueToGo converts a JavaScript value to a Go interface{}
func convertJSValueToGo(val js.Value) interface{} {
	// Handle null values first
	if val.IsNull() {
		return nil
	}

	switch val.Type() {
	case js.TypeBoolean:
		return val.Bool()
	case js.TypeNumber:
		return val.Float()
	case js.TypeString:
		return val.String()
	case js.TypeObject:
		// For objects, return the JS value itself for now
		// More sophisticated conversion could be added later
		return val
	default:
		return val
	}
}

// convertGoValueToJS converts a Go value to a JavaScript value
func convertGoValueToJS(val interface{}) js.Value {
	if val == nil {
		return js.Null()
	}

	switch v := val.(type) {
	case bool:
		return js.ValueOf(v)
	case int, int8, int16, int32, int64:
		return js.ValueOf(v)
	case uint, uint8, uint16, uint32, uint64:
		return js.ValueOf(v)
	case float32, float64:
		return js.ValueOf(v)
	case string:
		return js.ValueOf(v)
	case js.Value:
		return v
	default:
		// For other types, try to convert using js.ValueOf
		return js.ValueOf(v)
	}
}

// Initialize callback integration when the package is imported
func init() {
	setupCallbackIntegration()
}
