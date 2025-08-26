//go:build js && wasm

package comps

import (
	"fmt"
	"testing"

	g "maragu.dev/gomponents"
	h "maragu.dev/gomponents/html"
)

// TestExampleComponent tests the functional ExampleComponent
func TestExampleComponent(t *testing.T) {
	title := "Test Title"
	node := ExampleComponent(title)

	if node == nil {
		t.Error("ExampleComponent should return a non-nil node")
	}
}

// TestOtherComponent tests the functional OtherComponent
func TestOtherComponent(t *testing.T) {
	node := OtherComponent()

	if node == nil {
		t.Error("OtherComponent should return a non-nil node")
	}
}

// TestFragment tests the Fragment helper
func TestFragment(t *testing.T) {
	// Create some test nodes
	child1 := h.Div(g.Text("Child 1"))
	child2 := h.Div(g.Text("Child 2"))
	child3 := h.Div(g.Text("Child 3"))

	// Create a fragment with multiple children
	fragment := Fragment(child1, child2, child3)

	// Verify that fragment is not nil
	if fragment == nil {
		t.Error("Fragment should return a non-nil node")
	}

	// In gomponents, Group nodes don't render as a single element,
	// so we can't easily test the output without rendering to string.
	// But we can verify that it accepts multiple children without error.
}

// TestPortal tests the Portal helper
func TestPortal(t *testing.T) {
	// Create a test child node
	child := h.Div(g.Text("Portal Content"))

	// Create a portal with a target and child
	portal := Portal("#modal-target", child)

	// Verify that portal returns the child node (basic implementation)
	if portal == nil {
		t.Error("Portal should return a non-nil node")
	}

	// In the current basic implementation, Portal just returns the children
	// A full implementation would require DOM manipulation during mount
}

// TestMemo tests the Memo helper
func TestMemo(t *testing.T) {
	// Create a simple component function
	counter := 0
	componentFunc := func() g.Node {
		counter++
		return h.Div(g.Text(fmt.Sprintf("Rendered %d times", counter)))
	}

	// Call Memo multiple times with same dependencies
	node1 := Memo(componentFunc, "dep1")
	node2 := Memo(componentFunc, "dep1")

	// Both should return non-nil nodes
	if node1 == nil {
		t.Error("First Memo call should return a non-nil node")
	}
	if node2 == nil {
		t.Error("Second Memo call should return a non-nil node")
	}

	// In the current basic implementation, Memo doesn't actually cache
	// A full implementation would need to track dependencies
}

// TestLazy tests the Lazy helper
func TestLazy(t *testing.T) {
	// Create a loader function that returns a component
	loader := func() func() g.Node {
		return func() g.Node {
			return h.Div(g.Text("Lazy Loaded Component"))
		}
	}

	// Load the component lazily
	node := Lazy(loader)

	// Should return a non-nil node
	if node == nil {
		t.Error("Lazy should return a non-nil node")
	}

	// In the current basic implementation, Lazy loads synchronously
	// A full implementation would handle asynchronous loading
}

// TestErrorBoundary tests the ErrorBoundary helper
func TestErrorBoundary(t *testing.T) {
	// Create a fallback function
	fallback := func(err error) g.Node {
		return h.Div(g.Text(fmt.Sprintf("Error: %v", err)))
	}

	// Create normal children
	children := h.Div(g.Text("Normal Content"))

	// Create error boundary
	node := ErrorBoundary(ErrorBoundaryProps{
		Fallback: fallback,
		Children: children,
	})

	// Should return a non-nil node
	if node == nil {
		t.Error("ErrorBoundary should return a non-nil node")
	}

	// In the current basic implementation, ErrorBoundary doesn't actually catch errors
	// A full implementation would need to handle error catching
}

// TestMountContainerHelpers tests the mount container helper functions
func TestMountContainerHelpers(t *testing.T) {
	// Test setting and getting current mount container
	original := getCurrentMountContainer()
	defer setCurrentMountContainer(original) // Restore original

	testContainer := "test-container"
	setCurrentMountContainer(testContainer)

	current := getCurrentMountContainer()
	if current != testContainer {
		t.Errorf("Expected current mount container to be '%s', got '%s'", testContainer, current)
	}

	// Test resetting to empty
	setCurrentMountContainer("")
	current = getCurrentMountContainer()
	if current != "" {
		t.Errorf("Expected current mount container to be empty, got '%s'", current)
	}
}

// TestEnqueueOnMount tests the enqueueOnMount function
func TestEnqueueOnMount(t *testing.T) {
	// Track if callback was executed
	callbackExecuted := false
	callback := func() {
		callbackExecuted = true
	}

	// Enqueue the callback
	enqueueOnMount(callback)

	// Execute the mount queue
	testExecuteMountQueue()

	// Verify callback was executed
	if !callbackExecuted {
		t.Error("Enqueued callback should be executed when mount queue is processed")
	}
}

// TestMultipleOnMountCallbacks tests multiple callbacks in the mount queue
func TestMultipleOnMountCallbacks(t *testing.T) {
	callback1Executed := false
	callback2Executed := false
	callback3Executed := false

	callback1 := func() { callback1Executed = true }
	callback2 := func() { callback2Executed = true }
	callback3 := func() { callback3Executed = true }

	// Enqueue multiple callbacks
	enqueueOnMount(callback1)
	enqueueOnMount(callback2)
	enqueueOnMount(callback3)

	// Execute the mount queue
	testExecuteMountQueue()

	// Verify all callbacks were executed
	if !callback1Executed {
		t.Error("First callback should be executed")
	}
	if !callback2Executed {
		t.Error("Second callback should be executed")
	}
	if !callback3Executed {
		t.Error("Third callback should be executed")
	}
}

// TestFunctionalComponentComposition tests composing functional components
func TestFunctionalComponentComposition(t *testing.T) {
	// Create a composed component using the functional components
	composedComponent := func() g.Node {
		return h.Div(
			ExampleComponent("Composed Title"),
			OtherComponent(),
			Fragment(
				h.P(g.Text("Additional content")),
				h.Span(g.Text("More content")),
			),
		)
	}

	node := composedComponent()
	if node == nil {
		t.Error("Composed component should return a non-nil node")
	}
}

// TestFunctionalComponentWithState tests functional components with external state
func TestFunctionalComponentWithState(t *testing.T) {
	// Simulate external state
	state := struct {
		title   string
		content string
	}{
		title:   "Dynamic Title",
		content: "Dynamic Content",
	}

	// Create components using the state
	titleNode := ExampleComponent(state.title)
	contentNode := OtherComponent()

	if titleNode == nil {
		t.Error("Title component should return a non-nil node")
	}
	if contentNode == nil {
		t.Error("Content component should return a non-nil node")
	}
}

// Helper function to execute all queued mount callbacks
// Note: This function is already defined in comps.go, but we need it for testing
func testExecuteMountQueue() {
	for len(mountQueue) > 0 {
		callback := mountQueue[0]
		mountQueue = mountQueue[1:]
		callback()
	}
}
