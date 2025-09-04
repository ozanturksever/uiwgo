//go:build !js || !wasm

package logutil

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"testing"
)

// captureOutput captures stdout during function execution
func captureOutput(fn func()) string {
	// Save original stdout
	origStdout := os.Stdout
	
	// Create a pipe to capture output
	r, w, _ := os.Pipe()
	os.Stdout = w
	
	// Buffer to store output
	var buf bytes.Buffer
	done := make(chan bool)
	
	// Start goroutine to read from pipe
	go func() {
		buf.ReadFrom(r)
		done <- true
	}()
	
	// Execute function
	fn()
	
	// Close writer and restore stdout
	w.Close()
	os.Stdout = origStdout
	
	// Wait for reading to complete
	<-done
	r.Close()
	
	return buf.String()
}

func TestLog(t *testing.T) {
	tests := []struct {
		name     string
		args     []any
		expected string
	}{
		{
			name:     "single string",
			args:     []any{"hello"},
			expected: "hello\n",
		},
		{
			name:     "multiple strings",
			args:     []any{"hello", "world"},
			expected: "hello world\n",
		},
		{
			name:     "mixed types",
			args:     []any{"count:", 42, "active:", true},
			expected: "count: 42 active: true\n",
		},
		{
			name:     "empty args",
			args:     []any{},
			expected: "\n",
		},
		{
			name:     "nil value",
			args:     []any{nil},
			expected: "<nil>\n",
		},
		{
			name:     "numbers",
			args:     []any{123, 45.67, -89},
			expected: "123 45.67 -89\n",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := captureOutput(func() {
				Log(tt.args...)
			})
			
			if output != tt.expected {
				t.Errorf("Log() output = %q, expected %q", output, tt.expected)
			}
		})
	}
}

func TestLogf(t *testing.T) {
	tests := []struct {
		name     string
		format   string
		args     []any
		expected string
	}{
		{
			name:     "simple format",
			format:   "hello %s",
			args:     []any{"world"},
			expected: "hello world",
		},
		{
			name:     "multiple placeholders",
			format:   "count: %d, active: %t, name: %s",
			args:     []any{42, true, "test"},
			expected: "count: 42, active: true, name: test",
		},
		{
			name:     "no placeholders",
			format:   "static message",
			args:     []any{},
			expected: "static message",
		},
		{
			name:     "float formatting",
			format:   "value: %.2f",
			args:     []any{3.14159},
			expected: "value: 3.14",
		},
		{
			name:     "with newline",
			format:   "line 1\nline 2\n",
			args:     []any{},
			expected: "line 1\nline 2\n",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := captureOutput(func() {
				Logf(tt.format, tt.args...)
			})
			
			if output != tt.expected {
				t.Errorf("Logf() output = %q, expected %q", output, tt.expected)
			}
		})
	}
}

func TestLogfWithComplexTypes(t *testing.T) {
	// Test with struct
	type testStruct struct {
		Name  string
		Value int
	}
	
	s := testStruct{Name: "test", Value: 42}
	output := captureOutput(func() {
		Logf("struct: %+v", s)
	})
	
	expected := "struct: {Name:test Value:42}"
	if output != expected {
		t.Errorf("Logf() with struct output = %q, expected %q", output, expected)
	}
}

func TestLogWithComplexTypes(t *testing.T) {
	// Test with slice
	slice := []int{1, 2, 3}
	output := captureOutput(func() {
		Log("slice:", slice)
	})
	
	if !strings.Contains(output, "slice: [1 2 3]") {
		t.Errorf("Log() with slice should contain slice representation, got: %q", output)
	}
	
	// Test with map
	m := map[string]int{"a": 1, "b": 2}
	output = captureOutput(func() {
		Log("map:", m)
	})
	
	if !strings.Contains(output, "map:") {
		t.Errorf("Log() with map should contain map representation, got: %q", output)
	}
}

func TestLogConcurrency(t *testing.T) {
	// Test that Log is safe for concurrent use
	done := make(chan bool, 10)
	
	for i := 0; i < 10; i++ {
		go func(id int) {
			Log(fmt.Sprintf("goroutine %d", id))
			done <- true
		}(i)
	}
	
	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}
	
	// If we reach here without panic, the test passes
}

func TestLogfConcurrency(t *testing.T) {
	// Test that Logf is safe for concurrent use
	done := make(chan bool, 10)
	
	for i := 0; i < 10; i++ {
		go func(id int) {
			Logf("goroutine %d\n", id)
			done <- true
		}(i)
	}
	
	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}
	
	// If we reach here without panic, the test passes
}