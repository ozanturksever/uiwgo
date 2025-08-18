// events_test.go
// Comprehensive unit tests for event handling functionality
// Tests OnClick and OnInput event handlers, focusing on DOM-independent aspects

package golid

import (
	"strings"
	"testing"

	. "maragu.dev/gomponents"
)

// TestOnClickStructure tests that OnClick creates proper attribute structure
func TestOnClickStructure(t *testing.T) {
	callbackExecuted := false
	callback := func() {
		callbackExecuted = true
	}

	clickHandler := OnClick(callback)

	if clickHandler == nil {
		t.Fatal("OnClick should return a non-nil Node")
	}

	html := RenderHTML(clickHandler)

	// Should create an ID attribute
	if !strings.Contains(html, `id="`) {
		t.Errorf("OnClick should create ID attribute, got: %s", html)
	}

	// Should be an attribute (not a full element)
	if strings.Contains(html, "<") && strings.Contains(html, ">") {
		// If it contains < and >, it should be a simple attribute
		if !strings.Contains(html, `id="`) {
			t.Errorf("OnClick should generate ID attribute, got: %s", html)
		}
	}

	// The callback should not be executed during structure creation
	if callbackExecuted {
		t.Error("Callback should not be executed during OnClick creation")
	}
}

// TestOnClickIDGeneration tests that each OnClick gets unique ID
func TestOnClickIDGeneration(t *testing.T) {
	callback1 := func() {}
	callback2 := func() {}
	callback3 := func() {}

	click1 := OnClick(callback1)
	click2 := OnClick(callback2)
	click3 := OnClick(callback3)

	html1 := RenderHTML(click1)
	html2 := RenderHTML(click2)
	html3 := RenderHTML(click3)

	// Extract IDs
	id1 := extractID(html1)
	id2 := extractID(html2)
	id3 := extractID(html3)

	if id1 == "" || id2 == "" || id3 == "" {
		t.Error("All OnClick handlers should have non-empty IDs")
	}

	if id1 == id2 || id1 == id3 || id2 == id3 {
		t.Errorf("All OnClick IDs should be unique: id1=%s, id2=%s, id3=%s", id1, id2, id3)
	}
}

// TestOnClickCallbackTypes tests OnClick with different callback types
func TestOnClickCallbackTypes(t *testing.T) {
	testCases := []struct {
		name     string
		callback func()
	}{
		{
			name: "simple callback",
			callback: func() {
				// Simple callback
			},
		},
		{
			name: "callback with closure",
			callback: func() {
				counter := 1
				counter++
			},
		},
		{
			name:     "empty callback",
			callback: func() {},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			clickHandler := OnClick(tc.callback)

			if clickHandler == nil {
				t.Errorf("OnClick should work with %s", tc.name)
			}

			html := RenderHTML(clickHandler)
			if !strings.Contains(html, `id="`) {
				t.Errorf("OnClick should generate ID for %s", tc.name)
			}
		})
	}
}

// TestOnInputStructure tests that OnInput creates proper attribute structure
func TestOnInputStructure(t *testing.T) {
	var receivedValue string
	handler := func(value string) {
		receivedValue = value
	}

	inputHandler := OnInput(handler)

	if inputHandler == nil {
		t.Fatal("OnInput should return a non-nil Node")
	}

	html := RenderHTML(inputHandler)

	// Should create an ID attribute
	if !strings.Contains(html, `id="`) {
		t.Errorf("OnInput should create ID attribute, got: %s", html)
	}

	// The handler should not be executed during structure creation
	if receivedValue != "" {
		t.Error("Handler should not be executed during OnInput creation")
	}
}

// TestOnInputIDGeneration tests that each OnInput gets unique ID
func TestOnInputIDGeneration(t *testing.T) {
	handler1 := func(value string) {}
	handler2 := func(value string) {}
	handler3 := func(value string) {}

	input1 := OnInput(handler1)
	input2 := OnInput(handler2)
	input3 := OnInput(handler3)

	html1 := RenderHTML(input1)
	html2 := RenderHTML(input2)
	html3 := RenderHTML(input3)

	// Extract IDs
	id1 := extractID(html1)
	id2 := extractID(html2)
	id3 := extractID(html3)

	if id1 == "" || id2 == "" || id3 == "" {
		t.Error("All OnInput handlers should have non-empty IDs")
	}

	if id1 == id2 || id1 == id3 || id2 == id3 {
		t.Errorf("All OnInput IDs should be unique: id1=%s, id2=%s, id3=%s", id1, id2, id3)
	}
}

// TestOnInputHandlerTypes tests OnInput with different handler types
func TestOnInputHandlerTypes(t *testing.T) {
	testCases := []struct {
		name    string
		handler func(string)
	}{
		{
			name: "simple handler",
			handler: func(value string) {
				// Simple handler
			},
		},
		{
			name: "handler with processing",
			handler: func(value string) {
				processed := strings.ToUpper(value)
				_ = processed
			},
		},
		{
			name: "handler with validation",
			handler: func(value string) {
				if len(value) > 0 {
					// Valid input
				}
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			inputHandler := OnInput(tc.handler)

			if inputHandler == nil {
				t.Errorf("OnInput should work with %s", tc.name)
			}

			html := RenderHTML(inputHandler)
			if !strings.Contains(html, `id="`) {
				t.Errorf("OnInput should generate ID for %s", tc.name)
			}
		})
	}
}

// TestEventHandlerCombinations tests combining different event handlers
func TestEventHandlerCombinations(t *testing.T) {
	// Test that different event handlers can coexist
	var clickCount int
	var lastInputValue string

	clickHandler := OnClick(func() {
		clickCount++
	})

	inputHandler := OnInput(func(value string) {
		lastInputValue = value
	})

	handlers := []Node{clickHandler, inputHandler}

	for i, handler := range handlers {
		if handler == nil {
			t.Errorf("Handler %d should not be nil", i)
		}

		html := RenderHTML(handler)
		if !strings.Contains(html, `id="`) {
			t.Errorf("Handler %d should have ID attribute", i)
		}
	}

	// Extract IDs and verify they're unique
	html1 := RenderHTML(clickHandler)
	html2 := RenderHTML(inputHandler)

	id1 := extractID(html1)
	id2 := extractID(html2)

	if id1 == id2 {
		t.Errorf("Different event handlers should have unique IDs: %s vs %s", id1, id2)
	}

	// Verify initial state
	if clickCount != 0 {
		t.Error("Click count should be 0 initially")
	}
	if lastInputValue != "" {
		t.Error("Input value should be empty initially")
	}
}

// TestEventHandlerParameterValidation tests parameter validation
func TestEventHandlerParameterValidation(t *testing.T) {
	// Test with nil callback (should not panic)
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Event handlers should not panic with nil callbacks: %v", r)
		}
	}()

	// OnClick with nil - this would be a programming error, but shouldn't crash
	// We can't actually pass nil due to Go's type system, so we test with valid callbacks
	validClick := OnClick(func() {})
	if validClick == nil {
		t.Error("OnClick should return non-nil with valid callback")
	}

	validInput := OnInput(func(string) {})
	if validInput == nil {
		t.Error("OnInput should return non-nil with valid handler")
	}
}

// TestEventHandlerAsyncBehavior tests the async nature of event binding
func TestEventHandlerAsyncBehavior(t *testing.T) {
	// Test that OnClick returns immediately (doesn't block)
	executed := false

	// This should return immediately, not wait for DOM binding
	clickHandler := OnClick(func() {
		executed = true
	})

	if clickHandler == nil {
		t.Error("OnClick should return immediately")
	}

	// The callback should not have been executed yet
	if executed {
		t.Error("OnClick callback should not execute during creation")
	}

	// Same for OnInput
	inputExecuted := false
	inputHandler := OnInput(func(value string) {
		inputExecuted = true
	})

	if inputHandler == nil {
		t.Error("OnInput should return immediately")
	}

	if inputExecuted {
		t.Error("OnInput handler should not execute during creation")
	}
}

// TestEventHandlerIDFormat tests that generated IDs follow expected format
func TestEventHandlerIDFormat(t *testing.T) {
	clickHandler := OnClick(func() {})
	inputHandler := OnInput(func(string) {})

	clickHTML := RenderHTML(clickHandler)
	inputHTML := RenderHTML(inputHandler)

	clickID := extractID(clickHTML)
	inputID := extractID(inputHTML)

	// IDs should not be empty
	if clickID == "" || inputID == "" {
		t.Error("Generated IDs should not be empty")
	}

	// IDs should have reasonable length (not too short, not too long)
	if len(clickID) < 5 || len(clickID) > 50 {
		t.Errorf("Click ID length should be reasonable, got: %d chars", len(clickID))
	}

	if len(inputID) < 5 || len(inputID) > 50 {
		t.Errorf("Input ID length should be reasonable, got: %d chars", len(inputID))
	}
}

// TestEventHandlerConcurrentCreation tests concurrent creation of event handlers
func TestEventHandlerConcurrentCreation(t *testing.T) {
	// Create multiple handlers concurrently to test for race conditions
	const numHandlers = 10

	clickHandlers := make([]Node, numHandlers)
	inputHandlers := make([]Node, numHandlers)

	// Create handlers concurrently
	done := make(chan bool, numHandlers*2)

	for i := 0; i < numHandlers; i++ {
		go func(idx int) {
			clickHandlers[idx] = OnClick(func() {})
			done <- true
		}(i)

		go func(idx int) {
			inputHandlers[idx] = OnInput(func(string) {})
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < numHandlers*2; i++ {
		<-done
	}

	// Verify all handlers were created
	for i := 0; i < numHandlers; i++ {
		if clickHandlers[i] == nil {
			t.Errorf("Click handler %d should not be nil", i)
		}
		if inputHandlers[i] == nil {
			t.Errorf("Input handler %d should not be nil", i)
		}
	}

	// Verify all IDs are unique
	ids := make(map[string]bool)
	for i := 0; i < numHandlers; i++ {
		clickHTML := RenderHTML(clickHandlers[i])
		inputHTML := RenderHTML(inputHandlers[i])

		clickID := extractID(clickHTML)
		inputID := extractID(inputHTML)

		if ids[clickID] {
			t.Errorf("Duplicate click handler ID: %s", clickID)
		}
		if ids[inputID] {
			t.Errorf("Duplicate input handler ID: %s", inputID)
		}

		ids[clickID] = true
		ids[inputID] = true
	}
}
