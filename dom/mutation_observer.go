//go:build js && wasm

package dom

import (
	"fmt"
	"sync"
	"sync/atomic"
	"syscall/js"

	reactivity "github.com/ozanturksever/uiwgo/reactivity"
)

// ScopeRegistry manages the association between DOM elements and their cleanup scopes
type ScopeRegistry struct {
	mu     sync.RWMutex
	scopes map[string]*reactivity.CleanupScope // maps DOM element ID to its scope
	elements map[string]js.Value // maps DOM element ID to js.Value
}

// MutationObserverManager handles automatic cleanup when DOM nodes are removed
type MutationObserverManager struct {
	mu        sync.RWMutex
	observers map[string]*MutationObserver // maps container ID to observer
	registry  *ScopeRegistry
}

// MutationObserver wraps a JavaScript MutationObserver
type MutationObserver struct {
	observer   js.Value
	callback   js.Func
	container  js.Value
	registry   *ScopeRegistry
	isObserving bool
}

// Global instances
var (
	globalScopeRegistry   = &ScopeRegistry{
		scopes: make(map[string]*reactivity.CleanupScope),
		elements: make(map[string]js.Value),
	}
	globalObserverManager = &MutationObserverManager{
		observers: make(map[string]*MutationObserver),
		registry:  globalScopeRegistry,
	}
	elementIDCounter int64
)

// RegisterScope associates a DOM element with a cleanup scope
func (sr *ScopeRegistry) RegisterScope(element js.Value, scope *reactivity.CleanupScope) {
	sr.mu.Lock()
	defer sr.mu.Unlock()
	elementID := element.Get("id").String()
	if elementID == "" {
		// Generate a unique ID if element doesn't have one
		counter := atomic.AddInt64(&elementIDCounter, 1)
		elementID = fmt.Sprintf("elem_%s_%d", element.Get("tagName").String(), counter)
		// Set the ID on the element so we can find it later
		element.Set("id", elementID)
	}
	sr.scopes[elementID] = scope
	sr.elements[elementID] = element
}

// UnregisterScope removes the association between a DOM element and its scope
func (sr *ScopeRegistry) UnregisterScope(element js.Value) {
	sr.mu.Lock()
	defer sr.mu.Unlock()
	elementID := element.Get("id").String()
	if elementID == "" {
		// Try to find by element reference
		for id, elem := range sr.elements {
			if elem.Equal(element) {
				elementID = id
				break
			}
		}
	}
	if elementID != "" {
		delete(sr.scopes, elementID)
		delete(sr.elements, elementID)
	}
}

// GetScope retrieves the cleanup scope associated with a DOM element
func (sr *ScopeRegistry) GetScope(element js.Value) (*reactivity.CleanupScope, bool) {
	sr.mu.RLock()
	defer sr.mu.RUnlock()
	elementID := element.Get("id").String()
	if elementID == "" {
		// Try to find by element reference
		for id, elem := range sr.elements {
			if elem.Equal(element) {
				elementID = id
				break
			}
		}
	}
	if elementID != "" {
		scope, exists := sr.scopes[elementID]
		return scope, exists
	}
	return nil, false
}

// GetAllScopes returns a copy of all registered scopes
func (sr *ScopeRegistry) GetAllScopes() map[string]*reactivity.CleanupScope {
	sr.mu.RLock()
	defer sr.mu.RUnlock()
	result := make(map[string]*reactivity.CleanupScope)
	for k, v := range sr.scopes {
		result[k] = v
	}
	return result
}

// NewMutationObserver creates a new MutationObserver for a container
func NewMutationObserver(container js.Value, registry *ScopeRegistry) *MutationObserver {
	mo := &MutationObserver{
		container: container,
		registry:  registry,
	}

	// Create the callback function that will be called when mutations occur
	mo.callback = js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		if len(args) > 0 {
			mutations := args[0]
			mo.handleMutations(mutations)
		}
		return nil
	})

	// Create the JavaScript MutationObserver
	mo.observer = js.Global().Get("MutationObserver").New(mo.callback)

	return mo
}

// handleMutations processes mutation records and disposes scopes for removed nodes
func (mo *MutationObserver) handleMutations(mutations js.Value) {
	length := mutations.Get("length").Int()
	for i := 0; i < length; i++ {
		mutation := mutations.Index(i)
		mutationType := mutation.Get("type").String()

		// We're only interested in childList mutations (node additions/removals)
		if mutationType == "childList" {
			removedNodes := mutation.Get("removedNodes")
			mo.processRemovedNodes(removedNodes)
		}
	}
}

// processRemovedNodes handles cleanup for removed DOM nodes
func (mo *MutationObserver) processRemovedNodes(removedNodes js.Value) {
	length := removedNodes.Get("length").Int()
	for i := 0; i < length; i++ {
		node := removedNodes.Index(i)
		mo.cleanupNodeAndDescendants(node)
	}
}

// cleanupNodeAndDescendants recursively cleans up a node and all its descendants
func (mo *MutationObserver) cleanupNodeAndDescendants(node js.Value) {
	// Check if this node has an associated scope
	if scope, exists := mo.registry.GetScope(node); exists {
		// Dispose the scope (this will clean up all effects, listeners, etc.)
		scope.Dispose()
		// Unregister the scope
		mo.registry.UnregisterScope(node)
	}

	// Recursively check all child nodes
	if node.Get("nodeType").Int() == 1 { // Element node
		children := node.Get("children")
		length := children.Get("length").Int()
		for i := 0; i < length; i++ {
			child := children.Index(i)
			mo.cleanupNodeAndDescendants(child)
		}
	}
}

// StartObserving begins observing mutations on the container
func (mo *MutationObserver) StartObserving() {
	if mo.isObserving {
		return
	}

	// Configure the observer to watch for child list changes and subtree changes
	config := js.Global().Get("Object").New()
	config.Set("childList", true)
	config.Set("subtree", true)

	// Start observing
	mo.observer.Call("observe", mo.container, config)
	mo.isObserving = true
}

// StopObserving stops observing mutations
func (mo *MutationObserver) StopObserving() {
	if !mo.isObserving {
		return
	}

	mo.observer.Call("disconnect")
	mo.isObserving = false
}

// Dispose cleans up the observer and releases resources
func (mo *MutationObserver) Dispose() {
	mo.StopObserving()
	// Always try to release the callback - js.Func.Release() is safe to call multiple times
	mo.callback.Release()
}

// StartObservingContainer starts observing mutations for a specific container
func (mom *MutationObserverManager) StartObservingContainer(containerID string, container js.Value) {
	mom.mu.Lock()
	defer mom.mu.Unlock()

	// Don't create duplicate observers
	if _, exists := mom.observers[containerID]; exists {
		return
	}

	// Create and start the observer
	observer := NewMutationObserver(container, mom.registry)
	observer.StartObserving()
	mom.observers[containerID] = observer
}

// StopObservingContainer stops observing mutations for a specific container
func (mom *MutationObserverManager) StopObservingContainer(containerID string) {
	mom.mu.Lock()
	defer mom.mu.Unlock()

	if observer, exists := mom.observers[containerID]; exists {
		observer.Dispose()
		delete(mom.observers, containerID)
	}
}

// RegisterElementScope registers a DOM element with its cleanup scope
func RegisterElementScope(element js.Value, scope *reactivity.CleanupScope) {
	globalScopeRegistry.RegisterScope(element, scope)
}

// UnregisterElementScope unregisters a DOM element's cleanup scope
func UnregisterElementScope(element js.Value) {
	globalScopeRegistry.UnregisterScope(element)
}

// StartContainerObserver starts observing a container for mutations
func StartContainerObserver(containerID string, container js.Value) {
	globalObserverManager.StartObservingContainer(containerID, container)
}

// StopContainerObserver stops observing a container for mutations
func StopContainerObserver(containerID string) {
	globalObserverManager.StopObservingContainer(containerID)
}