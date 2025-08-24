//go:build js && wasm

package comps

import (
	"bytes"
	"fmt"
	"syscall/js"

	"github.com/ozanturksever/uiwgo/dom"
	"github.com/ozanturksever/uiwgo/reactivity"
	g "maragu.dev/gomponents"
)

// Node is an alias for gomponents.Node for convenience.
type Node = g.Node

// ComponentFunc defines the signature for components.
type ComponentFunc[P any] func(props P) Node

var (
	// queue of functions to execute after Mount has attached DOM
	mountQueue []func()
	// registry of mounted containers and their cleanup scopes
	mountedContainers = make(map[string]*MountContext)
)

// MountContext holds the cleanup scope and disposer for a mounted container
type MountContext struct {
	ElementID    string
	Container    js.Value
	CleanupScope *reactivity.CleanupScope
	Disposer     func()
}

// Mount renders a root component into a specific DOM element identified by its ID.
// It runs any OnMount functions and attaches reactive binders.
// Returns a disposer function that cleans up all effects, listeners, and registry entries.
func Mount(elementID string, root func() Node) func() {
	doc := js.Global().Get("document")
	if doc.IsUndefined() || doc.IsNull() {
		panic("document is not available (not running in a browser)")
	}
	container := doc.Call("getElementById", elementID)
	if container.IsUndefined() || container.IsNull() {
		panic(fmt.Sprintf("Mount: element with id '%s' not found", elementID))
	}

	// Set current mount container during component rendering and mounting
	setCurrentMountContainer(elementID)
	
	// Render the component to HTML
	var buf bytes.Buffer
	_ = root().Render(&buf)
	container.Set("innerHTML", buf.String())

	// Create a cleanup scope for this mount
	cleanupScope := reactivity.NewCleanupScope(nil)

	// Set cleanup scope as current during binder attachment and OnMount execution
	previous := reactivity.GetCurrentCleanupScope()
	reactivity.SetCurrentCleanupScope(cleanupScope)
	
	attachBinders(container)
	
	// Reset current mount container after binders are attached
	setCurrentMountContainer("")
	// Execute queued OnMount callbacks
	for len(mountQueue) > 0 {
		callback := mountQueue[0]
		mountQueue = mountQueue[1:]
		callback()
	}
	
	// Start MutationObserver for automatic cleanup
	dom.StartContainerObserver(elementID, container)
	
	// Restore previous cleanup scope
	reactivity.SetCurrentCleanupScope(previous)

	// Create disposer function
	disposer := func() {
		// Stop MutationObserver for this container
		dom.StopContainerObserver(elementID)
		// Remove from mounted containers registry
		delete(mountedContainers, elementID)
		// Dispose the cleanup scope (this will clean up all effects and listeners)
		cleanupScope.Dispose()
		// Clear the container's innerHTML
		container.Set("innerHTML", "")
		// Clean up registries for this container
		cleanupRegistriesForContainer(elementID)
	}

	// Store mount context
	mountedContainers[elementID] = &MountContext{
		ElementID:    elementID,
		Container:    container,
		CleanupScope: cleanupScope,
		Disposer:     disposer,
	}

	return disposer
}

// enqueueOnMount stores a function to be executed after Mount completes DOM attachment.
func enqueueOnMount(fn func()) { mountQueue = append(mountQueue, fn) }
