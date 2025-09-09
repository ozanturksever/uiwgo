//go:build js && wasm

package react

import (
	"syscall/js"
	"testing"
)

func TestCallbackRegistry_Register(t *testing.T) {
	registry := NewCallbackRegistry()
	
	// Test callback registration
	callback := func(args ...interface{}) interface{} {
		return "test result"
	}
	
	id := registry.Register(callback)
	if id == "" {
		t.Error("Expected non-empty callback ID")
	}
	
	if registry.GetCallbackCount() != 1 {
		t.Errorf("Expected 1 callback, got %d", registry.GetCallbackCount())
	}
}

func TestCallbackRegistry_Unregister(t *testing.T) {
	registry := NewCallbackRegistry()
	
	// Register a callback
	callback := func(args ...interface{}) interface{} {
		return "test result"
	}
	id := registry.Register(callback)
	
	// Test successful unregistration
	if !registry.Unregister(id) {
		t.Error("Expected successful unregistration")
	}
	
	if registry.GetCallbackCount() != 0 {
		t.Errorf("Expected 0 callbacks after unregistration, got %d", registry.GetCallbackCount())
	}
	
	// Test unregistering non-existent callback
	if registry.Unregister("non-existent") {
		t.Error("Expected failed unregistration for non-existent callback")
	}
}

func TestCallbackRegistry_Invoke(t *testing.T) {
	registry := NewCallbackRegistry()
	
	// Register a callback that returns the first argument
	callback := func(args ...interface{}) interface{} {
		if len(args) > 0 {
			return args[0]
		}
		return "no args"
	}
	id := registry.Register(callback)
	
	// Test successful invocation
	result, err := registry.Invoke(id, "test input")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if result != "test input" {
		t.Errorf("Expected 'test input', got %v", result)
	}
	
	// Test invocation with no arguments
	result, err = registry.Invoke(id)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if result != "no args" {
		t.Errorf("Expected 'no args', got %v", result)
	}
	
	// Test invocation of non-existent callback
	_, err = registry.Invoke("non-existent")
	if err == nil {
		t.Error("Expected error for non-existent callback")
	}
	
	// Check error type
	if bridgeErr, ok := err.(*ReactBridgeError); ok {
		if bridgeErr.Type != ErrorTypeCallbackNotFound {
			t.Errorf("Expected ErrorTypeCallbackNotFound, got %s", bridgeErr.Type)
		}
	} else {
		t.Error("Expected ReactBridgeError")
	}
}

func TestCallbackRegistry_Clear(t *testing.T) {
	registry := NewCallbackRegistry()
	
	// Register multiple callbacks
	for i := 0; i < 3; i++ {
		callback := func(args ...interface{}) interface{} {
			return "test"
		}
		registry.Register(callback)
	}
	
	if registry.GetCallbackCount() != 3 {
		t.Errorf("Expected 3 callbacks, got %d", registry.GetCallbackCount())
	}
	
	// Clear all callbacks
	registry.Clear()
	
	if registry.GetCallbackCount() != 0 {
		t.Errorf("Expected 0 callbacks after clear, got %d", registry.GetCallbackCount())
	}
}

func TestGlobalCallbackFunctions(t *testing.T) {
	// Clear any existing callbacks
	ClearGlobalCallbacks()
	
	// Test global registration
	callback := func(args ...interface{}) interface{} {
		return "global test"
	}
	id := RegisterCallback(callback)
	
	if GetGlobalCallbackCount() != 1 {
		t.Errorf("Expected 1 global callback, got %d", GetGlobalCallbackCount())
	}
	
	// Test global invocation
	result, err := InvokeCallback(id)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if result != "global test" {
		t.Errorf("Expected 'global test', got %v", result)
	}
	
	// Test global unregistration
	if !UnregisterCallback(id) {
		t.Error("Expected successful global unregistration")
	}
	
	if GetGlobalCallbackCount() != 0 {
		t.Errorf("Expected 0 global callbacks after unregistration, got %d", GetGlobalCallbackCount())
	}
}

func TestCallbackPanicRecovery(t *testing.T) {
	registry := NewCallbackRegistry()
	
	// Register a callback that panics
	callback := func(args ...interface{}) interface{} {
		panic("test panic")
	}
	id := registry.Register(callback)
	
	// Invoke the panicking callback - should not crash the test
	result, err := registry.Invoke(id)
	if err != nil {
		t.Errorf("Expected no error even with panic, got %v", err)
	}
	
	// Result should be nil since the callback panicked
	if result != nil {
		t.Errorf("Expected nil result from panicked callback, got %v", result)
	}
}

func TestJSValueConversion(t *testing.T) {
	// Test convertJSValueToGo
	tests := []struct {
		name     string
		jsValue  js.Value
		expected interface{}
	}{
		{"boolean true", js.ValueOf(true), true},
		{"boolean false", js.ValueOf(false), false},
		{"number", js.ValueOf(42.5), 42.5},
		{"string", js.ValueOf("hello"), "hello"},
		{"null", js.Null(), nil},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := convertJSValueToGo(tt.jsValue)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestGoValueConversion(t *testing.T) {
	// Test convertGoValueToJS
	tests := []struct {
		name     string
		goValue  interface{}
		expected js.Value
	}{
		{"nil", nil, js.Null()},
		{"boolean", true, js.ValueOf(true)},
		{"int", 42, js.ValueOf(42)},
		{"float", 3.14, js.ValueOf(3.14)},
		{"string", "test", js.ValueOf("test")},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := convertGoValueToJS(tt.goValue)
			
			// Compare types and values
			if result.Type() != tt.expected.Type() {
				t.Errorf("Expected type %v, got type %v", tt.expected.Type(), result.Type())
				return
			}
			
			// For non-null values, compare the actual values
			if !tt.expected.IsNull() {
				switch tt.expected.Type() {
				case js.TypeBoolean:
					if result.Bool() != tt.expected.Bool() {
						t.Errorf("Expected %v, got %v", tt.expected.Bool(), result.Bool())
					}
				case js.TypeNumber:
					if result.Float() != tt.expected.Float() {
						t.Errorf("Expected %v, got %v", tt.expected.Float(), result.Float())
					}
				case js.TypeString:
					if result.String() != tt.expected.String() {
						t.Errorf("Expected %v, got %v", tt.expected.String(), result.String())
					}
				}
			}
		})
	}
}

func TestCallbackIntegrationSetup(t *testing.T) {
	// Test that ReactCompatCallbacks object exists
	reactCallbacks := js.Global().Get("ReactCompatCallbacks")
	if reactCallbacks.IsUndefined() {
		t.Error("Expected ReactCompatCallbacks to be defined")
	}
	
	// Test that invoke function exists
	invokeFunc := reactCallbacks.Get("invoke")
	if invokeFunc.IsUndefined() {
		t.Error("Expected ReactCompatCallbacks.invoke to be defined")
	}
	
	if invokeFunc.Type() != js.TypeFunction {
		t.Error("Expected ReactCompatCallbacks.invoke to be a function")
	}
}

func TestJavaScriptCallbackInvocation(t *testing.T) {
	// Clear any existing callbacks
	ClearGlobalCallbacks()
	
	// Register a test callback
	callback := func(args ...interface{}) interface{} {
		if len(args) > 0 {
			return "received: " + args[0].(string)
		}
		return "no args received"
	}
	id := RegisterCallback(callback)
	
	// Test invoking the callback through JavaScript
	reactCallbacks := js.Global().Get("ReactCompatCallbacks")
	invokeFunc := reactCallbacks.Get("invoke")
	
	// Call with arguments
	result := invokeFunc.Invoke(string(id), "test message")
	if result.IsNull() {
		t.Error("Expected non-null result from JavaScript callback invocation")
	} else {
		resultStr := result.String()
		expected := "received: test message"
		if resultStr != expected {
			t.Errorf("Expected '%s', got '%s'", expected, resultStr)
		}
	}
	
	// Call without arguments
	result = invokeFunc.Invoke(string(id))
	if result.IsNull() {
		t.Error("Expected non-null result from JavaScript callback invocation")
	} else {
		resultStr := result.String()
		expected := "no args received"
		if resultStr != expected {
			t.Errorf("Expected '%s', got '%s'", expected, resultStr)
		}
	}
	
	// Test with non-existent callback ID
	result = invokeFunc.Invoke("non-existent-id")
	if !result.IsNull() {
		t.Error("Expected null result for non-existent callback")
	}
}