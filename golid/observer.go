// observer.go
// Global MutationObserver system for DOM monitoring

package golid

import (
	"syscall/js"
)

// ---------------------------
// 🔍 Global MutationObserver System
// ---------------------------

// Global observer instance
var globalObserver *ObserverManager

func init() {
	globalObserver = &ObserverManager{
		callbacks:         make(map[string]ElementCallback),
		dismountCallbacks: make(map[string][]LifecycleHook),
		trackedElements:   make(map[string]js.Value),
		maxRecursionDepth: 50, // Prevent infinite recursion
	}
}

// RegisterElement registers an element ID with a callback to be executed when the element is found
func (om *ObserverManager) RegisterElement(id string, callback ElementCallback) {
	// Check if element already exists
	if elem := NodeFromID(id); elem.Truthy() {
		om.mutex.Lock()
		om.trackedElements[id] = elem
		om.mutex.Unlock()
		callback()
		return
	}

	om.mutex.Lock()
	om.callbacks[id] = callback
	shouldStartObserving := !om.isObserving
	om.mutex.Unlock()

	// Start observing if not already observing
	if shouldStartObserving {
		om.startObserving()
	}
}

// RegisterDismountCallback registers a callback to be executed when an element is removed from the DOM
func (om *ObserverManager) RegisterDismountCallback(id string, callback LifecycleHook) {
	om.mutex.Lock()
	defer om.mutex.Unlock()

	if om.dismountCallbacks[id] == nil {
		om.dismountCallbacks[id] = make([]LifecycleHook, 0)
	}
	om.dismountCallbacks[id] = append(om.dismountCallbacks[id], callback)
}

// UnregisterElement removes an element from tracking
func (om *ObserverManager) UnregisterElement(id string) {
	om.mutex.Lock()
	delete(om.callbacks, id)
	delete(om.trackedElements, id)
	delete(om.dismountCallbacks, id)

	// Stop observing if no more callbacks and no tracked elements
	shouldStopObserving := len(om.callbacks) == 0 && len(om.trackedElements) == 0 && om.isObserving
	om.mutex.Unlock()

	if shouldStopObserving {
		om.stopObserving()
	}
}

// startObserving initializes the MutationObserver
func (om *ObserverManager) startObserving() {
	if om.isObserving {
		return
	}

	observerCallback := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		// Check if observer is suspended to prevent infinite loops
		om.mutex.RLock()
		if om.isSuspended {
			om.mutex.RUnlock()
			return nil
		}

		// Check recursion depth to prevent infinite loops
		if om.recursionDepth >= om.maxRecursionDepth {
			om.mutex.RUnlock()
			return nil
		}

		// Increment recursion depth
		om.recursionDepth++
		om.mutex.RUnlock()

		defer func() {
			// Decrement recursion depth when done
			om.mutex.Lock()
			om.recursionDepth--
			om.mutex.Unlock()
		}()

		mutations := args[0]
		mutationsLength := mutations.Get("length").Int()

		for i := 0; i < mutationsLength; i++ {
			mutation := mutations.Index(i)
			if mutation.Get("type").String() == "childList" {
				// Handle added nodes (mounting)
				addedNodes := mutation.Get("addedNodes")
				addedNodesLength := addedNodes.Get("length").Int()

				for j := 0; j < addedNodesLength; j++ {
					node := addedNodes.Index(j)
					om.checkNodeForTargets(node)
				}

				// Handle removed nodes (dismounting)
				removedNodes := mutation.Get("removedNodes")
				removedNodesLength := removedNodes.Get("length").Int()

				for j := 0; j < removedNodesLength; j++ {
					node := removedNodes.Index(j)
					om.checkNodeForDismount(node)
				}
			}
		}
		return nil
	})

	om.observer = js.Global().Get("MutationObserver").New(observerCallback)

	config := js.Global().Get("Object").New()
	config.Set("childList", true)
	config.Set("subtree", true)

	om.observer.Call("observe", doc.Get("body"), config)
	om.isObserving = true
}

// stopObserving disconnects the MutationObserver
func (om *ObserverManager) stopObserving() {
	if om.isObserving && om.observer.Truthy() {
		om.observer.Call("disconnect")
		om.isObserving = false
	}
}

// checkNodeForTargets checks if any registered elements are found in the added node
func (om *ObserverManager) checkNodeForTargets(node js.Value) {
	if node.Get("nodeType").Int() != 1 { // Not an element node
		return
	}

	var foundElements []struct {
		id       string
		element  js.Value
		callback ElementCallback
	}

	// First pass: read callbacks and find matching elements
	om.mutex.RLock()
	callbacksCopy := make(map[string]ElementCallback)
	for id, callback := range om.callbacks {
		callbacksCopy[id] = callback
	}
	om.mutex.RUnlock()

	// Check for matching elements without holding locks
	for id, callback := range callbacksCopy {
		if node.Get("id").String() == id {
			foundElements = append(foundElements, struct {
				id       string
				element  js.Value
				callback ElementCallback
			}{id, node, callback})
		} else {
			// Check descendants using getElementById instead of querySelector
			// to avoid CSS selector syntax issues with UUIDs starting with digits
			found := doc.Call("getElementById", id)
			if found.Truthy() {
				// Verify that the found element is actually a descendant of the node
				if isDescendantOf(found, node) {
					foundElements = append(foundElements, struct {
						id       string
						element  js.Value
						callback ElementCallback
					}{id, found, callback})
				}
			}
		}
	}

	// Second pass: update maps and call callbacks
	if len(foundElements) > 0 {
		om.mutex.Lock()
		for _, found := range foundElements {
			om.trackedElements[found.id] = found.element
			delete(om.callbacks, found.id)
		}
		om.mutex.Unlock()

		// Suspend observer to prevent infinite loops from callback-triggered DOM changes
		om.mutex.Lock()
		om.isSuspended = true
		om.mutex.Unlock()

		// Call callbacks outside of lock
		for _, found := range foundElements {
			found.callback()
		}

		// Resume observer after callbacks complete
		om.mutex.Lock()
		om.isSuspended = false
		om.mutex.Unlock()
	}
}

// checkNodeForDismount checks if any tracked elements are found in the removed node
func (om *ObserverManager) checkNodeForDismount(node js.Value) {
	if node.Get("nodeType").Int() != 1 { // Not an element node
		return
	}

	var dismountedElements []struct {
		id        string
		callbacks []LifecycleHook
	}

	// First pass: read tracked elements and find dismounted ones
	om.mutex.RLock()
	for id, trackedElement := range om.trackedElements {
		// Check if the removed node is the tracked element itself
		if trackedElement.Equal(node) {
			callbacks := make([]LifecycleHook, len(om.dismountCallbacks[id]))
			copy(callbacks, om.dismountCallbacks[id])
			dismountedElements = append(dismountedElements, struct {
				id        string
				callbacks []LifecycleHook
			}{id, callbacks})
		} else {
			// Check if the tracked element is a descendant of the removed node
			if isDescendantOf(trackedElement, node) {
				callbacks := make([]LifecycleHook, len(om.dismountCallbacks[id]))
				copy(callbacks, om.dismountCallbacks[id])
				dismountedElements = append(dismountedElements, struct {
					id        string
					callbacks []LifecycleHook
				}{id, callbacks})
			}
		}
	}
	om.mutex.RUnlock()

	// Second pass: clean up maps and call callbacks
	if len(dismountedElements) > 0 {
		om.mutex.Lock()
		for _, dismounted := range dismountedElements {
			delete(om.trackedElements, dismounted.id)
			delete(om.dismountCallbacks, dismounted.id)
		}
		om.mutex.Unlock()

		// Suspend observer to prevent infinite loops from callback-triggered DOM changes
		om.mutex.Lock()
		om.isSuspended = true
		om.mutex.Unlock()

		// Call callbacks outside of lock
		for _, dismounted := range dismountedElements {
			for _, callback := range dismounted.callbacks {
				callback()
			}
		}

		// Resume observer after callbacks complete
		om.mutex.Lock()
		om.isSuspended = false
		om.mutex.Unlock()
	}
}

// Helper function to check if element is a descendant of node
func isDescendantOf(element js.Value, ancestor js.Value) bool {
	current := element.Get("parentNode")
	for current.Truthy() {
		if current.Equal(ancestor) {
			return true
		}
		current = current.Get("parentNode")
	}
	return false
}
