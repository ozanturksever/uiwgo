//go:build js && wasm

package logutil

import (
	"syscall/js"
	"testing"
)

// mockConsole creates a mock console object for testing
func mockConsole() (js.Value, []string) {
	var logs []string
	
	// Create a mock log function that captures calls
	logFunc := js.FuncOf(func(this js.Value, args []js.Value) any {
		for _, arg := range args {
			logs = append(logs, arg.String())
		}
		return nil
	})
	
	// Create mock console object
	console := js.Global().Get("Object").New()
	console.Set("log", logFunc)
	
	return console, logs
}

func TestToJSArg(t *testing.T) {
	tests := []struct {
		name     string
		input    any
		expected string // We'll check the string representation
	}{
		{
			name:     "nil value",
			input:    nil,
			expected: "null", // JS null becomes "null" when converted to string
		},
		{
			name:     "string value",
			input:    "hello",
			expected: "hello",
		},
		{
			name:     "int value",
			input:    42,
			expected: "42",
		},
		{
			name:     "bool value",
			input:    true,
			expected: "true",
		},
		{
			name:     "float value",
			input:    3.14,
			expected: "3.14",
		},
		{
			name:     "js.Value",
			input:    js.ValueOf("test"),
			expected: "test",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := toJSArg(tt.input)
			
			// Convert result to string for comparison
			var resultStr string
			if result == nil {
				resultStr = "null"
			} else if jsVal, ok := result.(js.Value); ok {
				resultStr = jsVal.String()
			} else {
				resultStr = js.ValueOf(result).String()
			}
			
			if resultStr != tt.expected {
				t.Errorf("toJSArg(%v) = %q, expected %q", tt.input, resultStr, tt.expected)
			}
		})
	}
}

func TestToJSArgWithComplexTypes(t *testing.T) {
	// Test with struct (should be converted to string representation)
	type testStruct struct {
		Name  string
		Value int
	}
	
	s := testStruct{Name: "test", Value: 42}
	result := toJSArg(s)
	
	// Should be converted to string representation
	resultStr := js.ValueOf(result).String()
	if resultStr != "{test 42}" {
		t.Errorf("toJSArg(struct) = %q, expected %q", resultStr, "{test 42}")
	}
	
	// Test with slice
	slice := []int{1, 2, 3}
	result = toJSArg(slice)
	resultStr = js.ValueOf(result).String()
	if resultStr != "[1 2 3]" {
		t.Errorf("toJSArg(slice) = %q, expected %q", resultStr, "[1 2 3]")
	}
}

func TestLogWithMockConsole(t *testing.T) {
	// Save original console
	origConsole := js.Global().Get("console")
	defer js.Global().Set("console", origConsole)
	
	// Set up mock console
	mockConsole, logs := mockConsole()
	js.Global().Set("console", mockConsole)
	
	// Test logging
	Log("hello", "world", 42)
	
	// Check that log was called with correct arguments
	if len(logs) != 3 {
		t.Errorf("Expected 3 log entries, got %d", len(logs))
	}
	
	expected := []string{"hello", "world", "42"}
	for i, expected := range expected {
		if i >= len(logs) || logs[i] != expected {
			t.Errorf("Log entry %d = %q, expected %q", i, logs[i], expected)
		}
	}
}

func TestLogfWithMockConsole(t *testing.T) {
	// Save original console
	origConsole := js.Global().Get("console")
	defer js.Global().Set("console", origConsole)
	
	// Set up mock console
	mockConsole, logs := mockConsole()
	js.Global().Set("console", mockConsole)
	
	// Test formatted logging
	Logf("count: %d, name: %s", 42, "test")
	
	// Check that log was called with formatted message
	if len(logs) != 1 {
		t.Errorf("Expected 1 log entry, got %d", len(logs))
	}
	
	expected := "count: 42, name: test"
	if logs[0] != expected {
		t.Errorf("Logf() = %q, expected %q", logs[0], expected)
	}
}

func TestLogWithoutConsole(t *testing.T) {
	// Save original console
	origConsole := js.Global().Get("console")
	defer js.Global().Set("console", origConsole)
	
	// Remove console
	js.Global().Set("console", js.Undefined())
	
	// This should not panic and should fall back to fmt.Println
	// We can't easily test the fallback output in WASM, but we can ensure no panic
	Log("test message")
	Logf("formatted %s", "message")
	
	// If we reach here without panic, the test passes
}

func TestLogWithFalsyConsole(t *testing.T) {
	// Save original console
	origConsole := js.Global().Get("console")
	defer js.Global().Set("console", origConsole)
	
	// Set console to falsy value
	js.Global().Set("console", js.ValueOf(false))
	
	// This should not panic and should fall back to fmt.Println
	Log("test message")
	Logf("formatted %s", "message")
	
	// If we reach here without panic, the test passes
}

func TestLogWithJSValues(t *testing.T) {
	// Save original console
	origConsole := js.Global().Get("console")
	defer js.Global().Set("console", origConsole)
	
	// Set up mock console
	mockConsole, logs := mockConsole()
	js.Global().Set("console", mockConsole)
	
	// Test with js.Value arguments
	jsString := js.ValueOf("js string")
	jsNumber := js.ValueOf(123)
	jsBool := js.ValueOf(true)
	
	Log("mixed:", jsString, jsNumber, jsBool)
	
	// Check that JS values are passed through correctly
	if len(logs) != 4 {
		t.Errorf("Expected 4 log entries, got %d", len(logs))
	}
	
	expected := []string{"mixed:", "js string", "123", "true"}
	for i, expected := range expected {
		if i >= len(logs) || logs[i] != expected {
			t.Errorf("Log entry %d = %q, expected %q", i, logs[i], expected)
		}
	}
}

// Test wrapper-like interface (simulated)
type jsWrapper struct {
	value js.Value
}

func (w jsWrapper) JSValue() js.Value {
	return w.value
}

func TestLogWithJSWrapper(t *testing.T) {
	// Save original console
	origConsole := js.Global().Get("console")
	defer js.Global().Set("console", origConsole)
	
	// Set up mock console
	mockConsole, logs := mockConsole()
	js.Global().Set("console", mockConsole)
	
	// Test with wrapper that implements JSValue() method
	wrapper := jsWrapper{value: js.ValueOf("wrapped value")}
	Log("wrapper:", wrapper)
	
	// Check that wrapper is unwrapped correctly
	if len(logs) != 2 {
		t.Errorf("Expected 2 log entries, got %d", len(logs))
	}
	
	expected := []string{"wrapper:", "wrapped value"}
	for i, expected := range expected {
		if i >= len(logs) || logs[i] != expected {
			t.Errorf("Log entry %d = %q, expected %q", i, logs[i], expected)
		}
	}
}