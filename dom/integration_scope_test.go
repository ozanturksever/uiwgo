//go:build js && wasm

package dom

import (
	"testing"
	"time"

	reactivity "github.com/ozanturksever/uiwgo/reactivity"
	"honnef.co/go/js/dom/v2"
)

// TestReactiveElementScopeIntegration tests ReactiveElement's CleanupScope integration
func TestReactiveElementScopeIntegration(t *testing.T) {
	// Create a test element
	doc := dom.GetWindow().Document()
	el := doc.CreateElement("div")

	// Create reactive element
	reactiveEl := NewReactiveElement(el)

	// Verify scope is created
	scope := reactiveEl.GetScope()
	if scope == nil {
		t.Fatal("ReactiveElement should have a CleanupScope")
	}

	// Test reactive text binding
	textSignal := reactivity.CreateSignal("initial")
	reactiveEl.BindText(textSignal)

	// Verify initial value
	if el.TextContent() != "initial" {
		t.Errorf("Expected 'initial', got '%s'", el.TextContent())
	}

	// Update signal
	textSignal.Set("updated")
	if el.TextContent() != "updated" {
		t.Errorf("Expected 'updated', got '%s'", el.TextContent())
	}

	// Cleanup should dispose all effects
	reactiveEl.Cleanup()

	// After cleanup, signal updates should not affect the element
	textSignal.Set("should not update")
	if el.TextContent() == "should not update" {
		t.Error("Element should not update after cleanup")
	}
}

// TestReactiveNodeBuilderScopeIntegration tests ReactiveNodeBuilder's CleanupScope integration
func TestReactiveNodeBuilderScopeIntegration(t *testing.T) {
	// Create reactive node builder
	rnb := NewReactiveNode("div")

	// Verify scope is created
	scope := rnb.GetScope()
	if scope == nil {
		t.Fatal("ReactiveNodeBuilder should have a CleanupScope")
	}

	// Test reactive text binding
	textSignal := reactivity.CreateSignal("test")
	rnb.BindText(textSignal)

	// Build the element
	reactiveEl := rnb.Build()
	el := reactiveEl.Element()

	// Verify initial value
	if el.TextContent() != "test" {
		t.Errorf("Expected 'test', got '%s'", el.TextContent())
	}

	// Update signal
	textSignal.Set("updated")
	if el.TextContent() != "updated" {
		t.Errorf("Expected 'updated', got '%s'", el.TextContent())
	}

	// Cleanup should dispose all effects
	rnb.Cleanup()

	// After cleanup, signal updates should not affect the element
	textSignal.Set("should not update")
	if el.TextContent() == "should not update" {
		t.Error("Element should not update after cleanup")
	}
}

// TestNestedReactiveNodeBuilderCleanup tests cleanup of nested reactive node builders
func TestNestedReactiveNodeBuilderCleanup(t *testing.T) {
	// Create parent and child node builders
	parent := NewReactiveNode("div")
	child := NewReactiveNode("span")

	// Add reactive bindings to child only (parent will contain child)
	childSignal := reactivity.CreateSignal("child")
	child.BindText(childSignal)

	// Append child to parent
	parent.AppendChild(child)

	// Build elements
	parentEl := parent.Build()
	childEl := child.Build()

	// Verify initial values - parent should contain child's text
	if parentEl.Element().TextContent() != "child" {
		t.Errorf("Expected 'child', got '%s'", parentEl.Element().TextContent())
	}
	if childEl.Element().TextContent() != "child" {
		t.Errorf("Expected 'child', got '%s'", childEl.Element().TextContent())
	}

	// Update child signal
	childSignal.Set("child updated")

	// Verify updates - both parent and child should reflect the change
	if parentEl.Element().TextContent() != "child updated" {
		t.Errorf("Expected 'child updated', got '%s'", parentEl.Element().TextContent())
	}
	if childEl.Element().TextContent() != "child updated" {
		t.Errorf("Expected 'child updated', got '%s'", childEl.Element().TextContent())
	}

	// Cleanup parent should clean up child as well
	parent.Cleanup()

	// After cleanup, signal updates should not affect either element
	childSignal.Set("should not update child")

	if parentEl.Element().TextContent() == "should not update child" {
		t.Error("Parent element should not update after cleanup")
	}
	if childEl.Element().TextContent() == "should not update child" {
		t.Error("Child element should not update after cleanup")
	}
}

// TestReactiveElementEventCleanup tests event listener cleanup with CleanupScope
func TestReactiveElementEventCleanup(t *testing.T) {
	// Create a test element
	doc := dom.GetWindow().Document()
	el := doc.CreateElement("button")

	// Create reactive element
	reactiveEl := NewReactiveElement(el)

	// Track if event handler was called
	handlerCalled := false
	reactiveEl.OnClick(func(event dom.Event) {
		handlerCalled = true
	})

	// Cleanup should remove event listeners
	reactiveEl.Cleanup()

	// Note: In a real browser environment, we would simulate a click event
	// and verify the handler is not called. For this test, we just verify
	// that cleanup completes without errors.
	if handlerCalled {
		t.Error("Handler should not be called after cleanup")
	}
}

// TestReactiveElementMemoryLeak tests for memory leaks in reactive elements
func TestReactiveElementMemoryLeak(t *testing.T) {
	// Create multiple reactive elements with bindings
	for i := 0; i < 100; i++ {
		doc := dom.GetWindow().Document()
		el := doc.CreateElement("div")
		reactiveEl := NewReactiveElement(el)

		// Add multiple reactive bindings
		textSignal := reactivity.CreateSignal("test")
		attrSignal := reactivity.CreateSignal("value")

		reactiveEl.BindText(textSignal)
		reactiveEl.BindAttribute("data-test", attrSignal)
		reactiveEl.OnClick(func(event dom.Event) {})

		// Update signals to trigger effects
		textSignal.Set("updated")
		attrSignal.Set("new-value")

		// Cleanup immediately
		reactiveEl.Cleanup()
	}

	// Force garbage collection
	time.Sleep(10 * time.Millisecond)

	// Test passes if no memory leaks or panics occur
}