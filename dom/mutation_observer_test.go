//go:build js && wasm

package dom

import (
	"syscall/js"
	"testing"
	"time"

	"github.com/ozanturksever/uiwgo/reactivity"
)

// TestMutationObserverRegistration tests basic scope registration and unregistration
func TestMutationObserverRegistration(t *testing.T) {
	// Skip if not in browser environment
	if js.Global().Get("document").IsUndefined() {
		t.Skip("Skipping browser-specific test")
	}

	// Create test element
	document := js.Global().Get("document")
	element := document.Call("createElement", "div")

	// Create test scope
	scope := reactivity.NewCleanupScope(nil)

	// Register element with scope
	RegisterElementScope(element, scope)

	// Verify registration
	registeredScope, exists := globalScopeRegistry.GetScope(element)
	if !exists {
		t.Error("Element scope was not registered")
	}
	if registeredScope != scope {
		t.Error("Registered scope does not match original scope")
	}

	// Unregister element
	UnregisterElementScope(element)

	// Verify unregistration
	_, exists = globalScopeRegistry.GetScope(element)
	if exists {
		t.Error("Element scope was not unregistered")
	}
}

// TestMutationObserverCleanup tests automatic cleanup when DOM elements are removed
func TestMutationObserverCleanup(t *testing.T) {
	// Skip if not in browser environment
	if js.Global().Get("document").IsUndefined() {
		t.Skip("Skipping browser-specific test")
	}

	// Check if MutationObserver is available
	if js.Global().Get("MutationObserver").IsUndefined() {
		t.Skip("MutationObserver not available in test environment")
	}

	// Create test container
	document := js.Global().Get("document")
	container := document.Call("createElement", "div")
	container.Set("id", "test-mutation-container")
	document.Get("body").Call("appendChild", container)
	defer document.Get("body").Call("removeChild", container)

	// Create child element
	child := document.Call("createElement", "span")
	container.Call("appendChild", child)

	// Create scope with cleanup tracking
	var cleanupCalled bool
	scope := reactivity.NewCleanupScope(nil)
	scope.RegisterDisposer(func() {
		cleanupCalled = true
	})

	// Register child element with scope
	RegisterElementScope(child, scope)

	// Start observing the container
	StartContainerObserver("test-mutation-container", container)
	defer StopContainerObserver("test-mutation-container")

	// Remove child element to trigger mutation observer
	container.Call("removeChild", child)

	// Allow time for mutation observer to process
	time.Sleep(50 * time.Millisecond)

	// Verify cleanup was called
	if !cleanupCalled {
		t.Error("Cleanup function was not called when element was removed")
	}

	// Verify element was unregistered
	_, exists := globalScopeRegistry.GetScope(child)
	if exists {
		t.Error("Element scope was not automatically unregistered")
	}
}

// TestMutationObserverNestedCleanup tests cleanup of nested elements
func TestMutationObserverNestedCleanup(t *testing.T) {
	// Skip if not in browser environment
	if js.Global().Get("document").IsUndefined() {
		t.Skip("Skipping browser-specific test")
	}

	// Check if MutationObserver is available
	if js.Global().Get("MutationObserver").IsUndefined() {
		t.Skip("MutationObserver not available in test environment")
	}

	// Create test container
	document := js.Global().Get("document")
	container := document.Call("createElement", "div")
	container.Set("id", "test-nested-container")
	document.Get("body").Call("appendChild", container)
	defer document.Get("body").Call("removeChild", container)

	// Create nested structure: parent -> child -> grandchild
	parent := document.Call("createElement", "div")
	child := document.Call("createElement", "span")
	grandchild := document.Call("createElement", "em")

	child.Call("appendChild", grandchild)
	parent.Call("appendChild", child)
	container.Call("appendChild", parent)

	// Create scopes with cleanup tracking
	var parentCleanupCalled, childCleanupCalled, grandchildCleanupCalled bool

	parentScope := reactivity.NewCleanupScope(nil)
	parentScope.RegisterDisposer(func() {
		parentCleanupCalled = true
	})

	childScope := reactivity.NewCleanupScope(nil)
	childScope.RegisterDisposer(func() {
		childCleanupCalled = true
	})

	grandchildScope := reactivity.NewCleanupScope(nil)
	grandchildScope.RegisterDisposer(func() {
		grandchildCleanupCalled = true
	})

	// Register all elements with their scopes
	RegisterElementScope(parent, parentScope)
	RegisterElementScope(child, childScope)
	RegisterElementScope(grandchild, grandchildScope)

	// Start observing the container
	StartContainerObserver("test-nested-container", container)
	defer StopContainerObserver("test-nested-container")

	// Remove parent element (should trigger cleanup for all nested elements)
	container.Call("removeChild", parent)

	// Allow time for mutation observer to process
	time.Sleep(50 * time.Millisecond)

	// Verify all cleanups were called
	if !parentCleanupCalled {
		t.Error("Parent cleanup function was not called")
	}
	if !childCleanupCalled {
		t.Error("Child cleanup function was not called")
	}
	if !grandchildCleanupCalled {
		t.Error("Grandchild cleanup function was not called")
	}

	// Verify all elements were unregistered
	if _, exists := globalScopeRegistry.GetScope(parent); exists {
		t.Error("Parent element scope was not automatically unregistered")
	}
	if _, exists := globalScopeRegistry.GetScope(child); exists {
		t.Error("Child element scope was not automatically unregistered")
	}
	if _, exists := globalScopeRegistry.GetScope(grandchild); exists {
		t.Error("Grandchild element scope was not automatically unregistered")
	}
}

// TestMutationObserverMultipleContainers tests that multiple containers can be observed independently
func TestMutationObserverMultipleContainers(t *testing.T) {
	// Skip if not in browser environment
	if js.Global().Get("document").IsUndefined() {
		t.Skip("Skipping browser-specific test")
	}

	// Check if MutationObserver is available
	if js.Global().Get("MutationObserver").IsUndefined() {
		t.Skip("MutationObserver not available in test environment")
	}

	// Create test containers
	document := js.Global().Get("document")
	container1 := document.Call("createElement", "div")
	container1.Set("id", "test-container-1")
	container2 := document.Call("createElement", "div")
	container2.Set("id", "test-container-2")

	document.Get("body").Call("appendChild", container1)
	document.Get("body").Call("appendChild", container2)
	defer func() {
		document.Get("body").Call("removeChild", container1)
		document.Get("body").Call("removeChild", container2)
	}()

	// Create child elements
	child1 := document.Call("createElement", "span")
	child2 := document.Call("createElement", "span")
	container1.Call("appendChild", child1)
	container2.Call("appendChild", child2)

	// Create scopes with cleanup tracking
	var cleanup1Called, cleanup2Called bool

	scope1 := reactivity.NewCleanupScope(nil)
	scope1.RegisterDisposer(func() {
		cleanup1Called = true
	})

	scope2 := reactivity.NewCleanupScope(nil)
	scope2.RegisterDisposer(func() {
		cleanup2Called = true
	})

	// Register elements with their scopes
	RegisterElementScope(child1, scope1)
	RegisterElementScope(child2, scope2)

	// Start observing both containers
	StartContainerObserver("test-container-1", container1)
	StartContainerObserver("test-container-2", container2)
	defer func() {
		StopContainerObserver("test-container-1")
		StopContainerObserver("test-container-2")
	}()

	// Remove child from first container only
	container1.Call("removeChild", child1)

	// Allow time for mutation observer to process
	time.Sleep(50 * time.Millisecond)

	// Verify only first cleanup was called
	if !cleanup1Called {
		t.Error("Cleanup1 function was not called when child1 was removed")
	}
	if cleanup2Called {
		t.Error("Cleanup2 function was called when child2 was not removed")
	}

	// Verify only first element was unregistered
	if _, exists := globalScopeRegistry.GetScope(child1); exists {
		t.Error("Child1 element scope was not automatically unregistered")
	}
	if _, exists := globalScopeRegistry.GetScope(child2); !exists {
		t.Error("Child2 element scope was incorrectly unregistered")
	}
}