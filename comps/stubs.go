//go:build !js || !wasm

package comps

import (
	"github.com/ozanturksever/uiwgo/reactivity"
	g "maragu.dev/gomponents"
)

// Stub implementations for non-browser environments

var (
	// queue of functions to execute after Mount completes
	mountQueue []func()
	// tracks the current mount container during binding
	currentMountContainer string
)

// enqueueOnMount adds a function to be executed after Mount completes
func enqueueOnMount(fn func()) {
	mountQueue = append(mountQueue, fn)
}

// executeMountQueue executes all queued mount callbacks
func executeMountQueue() {
	for len(mountQueue) > 0 {
		callback := mountQueue[0]
		mountQueue = mountQueue[1:]
		callback()
	}
}

// getCurrentMountContainer returns the current mount container ID
func getCurrentMountContainer() string {
	return currentMountContainer
}

// setCurrentMountContainer sets the current mount container ID
func setCurrentMountContainer(containerID string) {
	currentMountContainer = containerID
}

// cleanupRegistriesForContainer is a stub for testing
func cleanupRegistriesForContainer(containerID string) {
	// Stub implementation for testing
}

// Mount is a stub implementation for testing
func Mount(elementID string, root func() g.Node) func() {
	// Set current mount container during component rendering
	setCurrentMountContainer(elementID)
	
	// Render the component (but don't actually mount to DOM in tests)
	_ = root()
	
	// Execute queued OnMount callbacks
	executeMountQueue()
	
	// Reset current mount container
	setCurrentMountContainer("")
	
	// Return a disposer function
	return func() {
		// Stub cleanup
		cleanupRegistriesForContainer(elementID)
	}
}

// OnMount schedules a function to run after Mount has attached the DOM.
func OnMount(fn func()) g.Node {
	// We return a no-op node so it can be used in gomponents trees.
	enqueueOnMount(fn)
	return g.Group([]g.Node{})
}

// OnCleanup is re-exported from reactivity.
var OnCleanup = reactivity.OnCleanup

// Fragment creates a group of nodes without a wrapper element
func Fragment(children ...g.Node) g.Node {
	return g.Group(children)
}

// ComponentFactory creates a component instance from a component struct
func ComponentFactory(component interface{}) g.Node {
	// Stub implementation - just return empty group
	return g.Group([]g.Node{})
}

// ComponentFactoryWithProps creates a component instance with props
func ComponentFactoryWithProps(component interface{}, props interface{}) g.Node {
	// Stub implementation - just return empty group
	return g.Group([]g.Node{})
}