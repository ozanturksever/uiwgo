//go:build js && wasm

package dom

import (
	"context"
	"testing"
	"time"

	reactivity "github.com/ozanturksever/uiwgo/reactivity"
	"honnef.co/go/js/dom/v2"
)

// TestBrowserSubtreeRemoval tests cleanup when DOM subtrees are removed
func TestBrowserSubtreeRemoval(t *testing.T) {
	doc := dom.GetWindow().Document()
	body := doc.QuerySelector("body")
	if body == nil {
		t.Skip("No body element found - skipping browser test")
	}
	container := doc.CreateElement("div")
	body.AppendChild(container)
	defer body.RemoveChild(container)

	// Create a structure with sibling elements (no nested text elements)
	parent := NewReactiveNode("div")
	child1 := NewReactiveNode("span")
	child2 := NewReactiveNode("p")
	child3 := NewReactiveNode("em")

	// Add reactive bindings to leaf elements only
	child1Signal := reactivity.CreateSignal("child1")
	child2Signal := reactivity.CreateSignal("child2")
	child3Signal := reactivity.CreateSignal("child3")

	child1.BindText(child1Signal)
	child2.BindText(child2Signal)
	child3.BindText(child3Signal)

	// Build flat structure (all children are siblings)
	parent.AppendChild(child1)
	parent.AppendChild(child2)
	parent.AppendChild(child3)

	// Mount to DOM
	parentEl := parent.Build()
	container.AppendChild(parentEl.Element())

	// Verify initial state
	if parentEl.Element().TextContent() != "child1child2child3" {
		t.Errorf("Unexpected initial content: %s", parentEl.Element().TextContent())
	}

	// Update signals to verify reactivity
	child1Signal.Set("updated child1")
	child3Signal.Set("updated child3")

	// Allow time for DOM updates
	time.Sleep(10 * time.Millisecond)

	expected := "updated child1child2updated child3"
	if parentEl.Element().TextContent() != expected {
		t.Errorf("Expected '%s', got '%s'", expected, parentEl.Element().TextContent())
	}

	// Remove subtree from DOM
	container.RemoveChild(parentEl.Element())

	// Cleanup the reactive structure
	parent.Cleanup()

	// Verify that signal updates no longer affect the removed elements
	child1Signal.Set("should not update")
	child3Signal.Set("should not update")

	// Allow time for any potential updates
	time.Sleep(10 * time.Millisecond)

	// The elements should retain their last values since they're disconnected
	if parentEl.Element().TextContent() == "should not updatechild2should not update" {
		t.Error("Elements should not update after cleanup")
	}
}

// TestBrowserDynamicMountUnmount tests repeated mount/unmount cycles
func TestBrowserDynamicMountUnmount(t *testing.T) {
	doc := dom.GetWindow().Document()
	body := doc.QuerySelector("body")
	if body == nil {
		t.Skip("No body element found - skipping browser test")
	}
	container := doc.CreateElement("div")
	body.AppendChild(container)
	defer body.RemoveChild(container)

	for i := 0; i < 5; i++ {
		// Create reactive element
		node := NewReactiveNode("div")
		signal := reactivity.CreateSignal("initial")
		node.BindText(signal)

		// Mount to DOM
		el := node.Build()
		container.AppendChild(el.Element())

		// Verify initial state
		if el.Element().TextContent() != "initial" {
			t.Errorf("Iteration %d: Expected 'initial', got '%s'", i, el.Element().TextContent())
		}

		// Update signal
		signal.Set("updated")
		time.Sleep(5 * time.Millisecond)

		if el.Element().TextContent() != "updated" {
			t.Errorf("Iteration %d: Expected 'updated', got '%s'", i, el.Element().TextContent())
		}

		// Unmount and cleanup
		container.RemoveChild(el.Element())
		node.Cleanup()

		// Verify cleanup worked
		signal.Set("should not update")
		time.Sleep(5 * time.Millisecond)

		if el.Element().TextContent() == "should not update" {
			t.Errorf("Iteration %d: Element should not update after cleanup", i)
		}
	}
}

// TestBrowserEventCleanupOnRemoval tests event listener cleanup when elements are removed
func TestBrowserEventCleanupOnRemoval(t *testing.T) {
	doc := dom.GetWindow().Document()
	body := doc.QuerySelector("body")
	if body == nil {
		t.Skip("No body element found - skipping browser test")
	}
	container := doc.CreateElement("div")
	body.AppendChild(container)
	defer body.RemoveChild(container)

	// Track event handler calls
	var clickCount int

	// Create reactive button with click handler
	button := NewReactiveNode("button")
	button.SetText("Click me")
	button.OnClick(func(event dom.Event) {
		clickCount++
	})

	// Mount to DOM
	buttonEl := button.Build()
	container.AppendChild(buttonEl.Element())

	// Simulate click event using JavaScript
	buttonEl.Element().Underlying().Call("click")

	if clickCount != 1 {
		t.Errorf("Expected 1 click, got %d", clickCount)
	}

	// Remove from DOM and cleanup
	container.RemoveChild(buttonEl.Element())
	button.Cleanup()

	// Simulate another click event (should not increment counter)
	buttonEl.Element().Underlying().Call("click")

	// Allow time for any potential handler execution
	time.Sleep(10 * time.Millisecond)

	if clickCount != 1 {
		t.Errorf("Expected click count to remain 1 after cleanup, got %d", clickCount)
	}
}

// TestBrowserScopeHierarchyCleanup tests that parent scope cleanup affects child scopes
func TestBrowserScopeHierarchyCleanup(t *testing.T) {
	doc := dom.GetWindow().Document()
	body := doc.QuerySelector("body")
	if body == nil {
		t.Skip("No body element found - skipping browser test")
	}
	container := doc.CreateElement("div")
	body.AppendChild(container)
	defer body.RemoveChild(container)

	// Create parent scope
	parentScope := reactivity.NewCleanupScope(nil)

	// Create child elements with parent scope
	prevScope := reactivity.GetCurrentCleanupScope()
	reactivity.SetCurrentCleanupScope(parentScope)

	// Create multiple reactive elements
	var elements []*ReactiveElement
	var signals []reactivity.Signal[string]

	for i := 0; i < 3; i++ {
		el := CreateReactiveDiv()
		signal := reactivity.CreateSignal("initial")
		el.BindText(signal)
		container.AppendChild(el.Element())

		elements = append(elements, el)
		signals = append(signals, signal)
	}

	reactivity.SetCurrentCleanupScope(prevScope)

	// Verify all elements are reactive
	for i, signal := range signals {
		signal.Set("updated")
		time.Sleep(5 * time.Millisecond)
		if elements[i].Element().TextContent() != "updated" {
			t.Errorf("Element %d not reactive", i)
		}
	}

	// Dispose parent scope (should cleanup all child elements)
	parentScope.Dispose()

	// Verify all elements are no longer reactive
	for i, signal := range signals {
		signal.Set("should not update")
		time.Sleep(5 * time.Millisecond)
		if elements[i].Element().TextContent() == "should not update" {
			t.Errorf("Element %d should not update after parent scope disposal", i)
		}
	}

	// Cleanup DOM
	for _, el := range elements {
		container.RemoveChild(el.Element())
	}
}

// TestBrowserAsyncCleanup tests cleanup behavior with async operations
func TestBrowserAsyncCleanup(t *testing.T) {
	doc := dom.GetWindow().Document()
	body := doc.QuerySelector("body")
	if body == nil {
		t.Skip("No body element found - skipping browser test")
	}
	container := doc.CreateElement("div")
	body.AppendChild(container)
	defer body.RemoveChild(container)

	// Create reactive element with async updates
	node := NewReactiveNode("div")
	signal := reactivity.CreateSignal("initial")
	node.BindText(signal)

	el := node.Build()
	container.AppendChild(el.Element())

	// Start async updates
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		ticker := time.NewTicker(10 * time.Millisecond)
		defer ticker.Stop()
		counter := 0
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				counter++
				signal.Set("async update")
				if counter > 10 {
					return
				}
			}
		}
	}()

	// Let it run for a bit
	time.Sleep(50 * time.Millisecond)

	// Cleanup while async updates are happening
	node.Cleanup()
	cancel()

	// Wait a bit more
	time.Sleep(50 * time.Millisecond)

	// Element should not be updating anymore
	lastContent := el.Element().TextContent()
	signal.Set("final update")
	time.Sleep(20 * time.Millisecond)

	if el.Element().TextContent() != lastContent {
		t.Error("Element should not update after cleanup during async operations")
	}

	container.RemoveChild(el.Element())
}