//go:build js && wasm

// counter_test.go
// Integration tests for the counter example to validate complete reactive workflow

package main

import (
	"fmt"
	"syscall/js"
	"testing"
	"time"

	"app/golid"
)

func TestCounterReactiveSystem(t *testing.T) {
	// Reset reactive context for clean test
	golid.ResetReactiveContext()
	golid.ResetScheduler()

	// Create a signal like in the counter example
	count, setCount := golid.CreateSignal(0)

	// Test initial value
	if count() != 0 {
		t.Errorf("Expected initial count 0, got %d", count())
	}

	// Test increment
	setCount(count() + 1)
	golid.FlushScheduler()

	if count() != 1 {
		t.Errorf("Expected count 1 after increment, got %d", count())
	}

	// Test decrement
	setCount(count() - 1)
	golid.FlushScheduler()

	if count() != 0 {
		t.Errorf("Expected count 0 after decrement, got %d", count())
	}

	// Test reset
	setCount(42)
	golid.FlushScheduler()
	setCount(0)
	golid.FlushScheduler()

	if count() != 0 {
		t.Errorf("Expected count 0 after reset, got %d", count())
	}
}

func TestCounterEffectSystem(t *testing.T) {
	golid.ResetReactiveContext()
	golid.ResetScheduler()

	count, setCount := golid.CreateSignal(0)
	var effectValue int
	var executionCount int

	// Create effect that tracks count changes
	golid.CreateEffect(func() {
		effectValue = count()
		executionCount++
	}, nil)

	golid.FlushScheduler()

	// Effect should run immediately
	if executionCount != 1 {
		t.Errorf("Expected effect to run once initially, got %d executions", executionCount)
	}
	if effectValue != 0 {
		t.Errorf("Expected effect value 0, got %d", effectValue)
	}

	// Update count and verify effect runs
	setCount(5)
	golid.FlushScheduler()

	if executionCount != 2 {
		t.Errorf("Expected effect to run twice after update, got %d executions", executionCount)
	}
	if effectValue != 5 {
		t.Errorf("Expected effect value 5, got %d", effectValue)
	}
}

func TestCounterDOMIntegration(t *testing.T) {
	golid.ResetReactiveContext()
	golid.ResetScheduler()

	// Create DOM elements for testing
	doc := js.Global().Get("document")
	testDiv := doc.Call("createElement", "div")
	testDiv.Set("id", "test-counter")
	doc.Get("body").Call("appendChild", testDiv)

	// Cleanup function
	defer func() {
		testDiv.Call("remove")
	}()

	count, setCount := golid.CreateSignal(0)

	// Create reactive text binding like in counter example
	_, cleanup := golid.CreateRoot(func() interface{} {
		golid.BindTextReactive(testDiv, func() string {
			return fmt.Sprintf("Count: %d", count())
		})
		return nil
	})
	defer cleanup()

	golid.FlushScheduler()

	// Allow DOM updates to process
	time.Sleep(10 * time.Millisecond)

	// Check initial text content
	initialText := testDiv.Get("textContent").String()
	if initialText != "Count: 0" {
		t.Errorf("Expected initial text 'Count: 0', got '%s'", initialText)
	}

	// Update count and check DOM updates
	setCount(42)
	golid.FlushScheduler()
	time.Sleep(10 * time.Millisecond)

	updatedText := testDiv.Get("textContent").String()
	if updatedText != "Count: 42" {
		t.Errorf("Expected updated text 'Count: 42', got '%s'", updatedText)
	}
}

func TestCounterEventHandlers(t *testing.T) {
	golid.ResetReactiveContext()
	golid.ResetScheduler()

	// Create DOM button for testing
	doc := js.Global().Get("document")
	testButton := doc.Call("createElement", "button")
	testButton.Set("id", "test-button")
	testButton.Set("textContent", "Click me")
	doc.Get("body").Call("appendChild", testButton)

	// Cleanup function
	defer func() {
		testButton.Call("remove")
	}()

	count, setCount := golid.CreateSignal(0)
	var clickCount int

	// Create click handler like in counter example
	_, cleanup := golid.CreateRoot(func() interface{} {
		handler := func() {
			clickCount++
			setCount(count() + 1)
		}
		golid.OnClickV2(handler)
		return nil
	})
	defer cleanup()

	golid.FlushScheduler()
	time.Sleep(10 * time.Millisecond)

	// Simulate click event
	clickEvent := doc.Call("createEvent", "MouseEvent")
	clickEvent.Call("initEvent", "click", true, true)
	testButton.Call("dispatchEvent", clickEvent)

	// Allow event processing
	time.Sleep(10 * time.Millisecond)
	golid.FlushScheduler()

	// Verify click was handled
	if clickCount != 1 {
		t.Errorf("Expected 1 click, got %d", clickCount)
	}
	if count() != 1 {
		t.Errorf("Expected count 1 after click, got %d", count())
	}
}

func TestCounterCompleteWorkflow(t *testing.T) {
	golid.ResetReactiveContext()
	golid.ResetScheduler()

	// Create complete counter setup
	doc := js.Global().Get("document")
	testContainer := doc.Call("createElement", "div")
	testContainer.Set("id", "test-container")
	doc.Get("body").Call("appendChild", testContainer)

	// Cleanup function
	defer func() {
		testContainer.Call("remove")
	}()

	count, setCount := golid.CreateSignal(0)

	// Create complete reactive setup
	_, cleanup := golid.CreateRoot(func() interface{} {
		// Create display element
		display := doc.Call("createElement", "div")
		display.Set("id", "count-display")
		testContainer.Call("appendChild", display)

		// Create buttons
		incrementBtn := doc.Call("createElement", "button")
		incrementBtn.Set("id", "increment-btn")
		incrementBtn.Set("textContent", "+")
		testContainer.Call("appendChild", incrementBtn)

		decrementBtn := doc.Call("createElement", "button")
		decrementBtn.Set("id", "decrement-btn")
		decrementBtn.Set("textContent", "-")
		testContainer.Call("appendChild", decrementBtn)

		resetBtn := doc.Call("createElement", "button")
		resetBtn.Set("id", "reset-btn")
		resetBtn.Set("textContent", "Reset")
		testContainer.Call("appendChild", resetBtn)

		// Bind reactive text
		golid.BindTextReactive(display, func() string {
			return fmt.Sprintf("Count: %d", count())
		})

		// Bind event handlers using direct DOM event listeners for testing
		incrementBtn.Call("addEventListener", "click", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			setCount(count() + 1)
			return nil
		}))

		decrementBtn.Call("addEventListener", "click", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			setCount(count() - 1)
			return nil
		}))

		resetBtn.Call("addEventListener", "click", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			setCount(0)
			return nil
		}))

		return nil
	})

	defer cleanup()

	golid.FlushScheduler()
	time.Sleep(10 * time.Millisecond)

	// Test initial state
	display := doc.Call("getElementById", "count-display")
	if display.Get("textContent").String() != "Count: 0" {
		t.Errorf("Expected initial display 'Count: 0', got '%s'", display.Get("textContent").String())
	}

	// Test increment
	incrementBtn := doc.Call("getElementById", "increment-btn")
	clickEvent := doc.Call("createEvent", "MouseEvent")
	clickEvent.Call("initEvent", "click", true, true)
	incrementBtn.Call("dispatchEvent", clickEvent)

	time.Sleep(10 * time.Millisecond)
	golid.FlushScheduler()
	time.Sleep(10 * time.Millisecond)

	if display.Get("textContent").String() != "Count: 1" {
		t.Errorf("Expected display 'Count: 1' after increment, got '%s'", display.Get("textContent").String())
	}

	// Test decrement
	decrementBtn := doc.Call("getElementById", "decrement-btn")
	decrementBtn.Call("dispatchEvent", clickEvent)

	time.Sleep(10 * time.Millisecond)
	golid.FlushScheduler()
	time.Sleep(10 * time.Millisecond)

	if display.Get("textContent").String() != "Count: 0" {
		t.Errorf("Expected display 'Count: 0' after decrement, got '%s'", display.Get("textContent").String())
	}

	// Test reset (first increment to non-zero)
	incrementBtn.Call("dispatchEvent", clickEvent)
	incrementBtn.Call("dispatchEvent", clickEvent)
	incrementBtn.Call("dispatchEvent", clickEvent)

	time.Sleep(10 * time.Millisecond)
	golid.FlushScheduler()
	time.Sleep(10 * time.Millisecond)

	if display.Get("textContent").String() != "Count: 3" {
		t.Errorf("Expected display 'Count: 3' after multiple increments, got '%s'", display.Get("textContent").String())
	}

	// Now test reset
	resetBtn := doc.Call("getElementById", "reset-btn")
	resetBtn.Call("dispatchEvent", clickEvent)

	time.Sleep(10 * time.Millisecond)
	golid.FlushScheduler()
	time.Sleep(10 * time.Millisecond)

	if display.Get("textContent").String() != "Count: 0" {
		t.Errorf("Expected display 'Count: 0' after reset, got '%s'", display.Get("textContent").String())
	}
}
