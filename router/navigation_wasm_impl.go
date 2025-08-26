//go:build js && wasm

package router

import (
	"syscall/js"

	dom "honnef.co/go/js/dom/v2"
)

// navigateWASMImpl handles navigation in WASM builds with proper history API integration.
func (r *Router) navigateWASMImpl(path string, options NavigateOptions) {
	window := dom.GetWindow()
	if window == nil {
		// Fallback for test environments where DOM is not available
		newLocation := Location{
			Pathname: path,
			Search:   "",
			Hash:     "",
			State:    options.State,
		}
		r.locationState.Set(newLocation)
		return
	}

	history := window.History()
	if history == nil {
		// Fallback for test environments where History API is not available
		newLocation := Location{
			Pathname: path,
			Search:   "",
			Hash:     "",
			State:    options.State,
		}
		r.locationState.Set(newLocation)
		return
	}

	// Create a new Location with the given path
	newLocation := Location{
		Pathname: path,
		Search:   "", // TODO: Parse search from path if needed
		Hash:     "", // TODO: Parse hash from path if needed
		State:    options.State,
	}

	// Push the new state to browser history
	var stateValue js.Value
	if options.State != nil {
		// Convert Go state to JS value
		stateValue = js.ValueOf(options.State)
	} else {
		stateValue = js.Null()
	}

	// Use pushState to add the new location to browser history
	history.PushState(stateValue, "", path)

	// Update the router's location state (this will trigger rendering)
	r.locationState.Set(newLocation)
}

// setupNavigationWASM sets up WASM-specific navigation functionality.
func setupNavigationWASM(router *Router) {
	// Add click event delegation for navigation links
	setupLinkClickHandling(router)
}

// setupLinkClickHandling sets up event delegation for navigation links.
func setupLinkClickHandling(router *Router) {
	document := dom.GetWindow().Document()

	// Add click event listener to document for event delegation
	document.AddEventListener("click", true, func(event dom.Event) {
		// Find closest ancestor <a data-router-link>
		target := event.Target()
		if target == nil {
			return
		}

		// Start from the target node, ascend until we find an anchor
		var curr dom.Node
		if tn, ok := target.(dom.Node); ok {
			curr = tn
		} else {
			return
		}

		var anchor dom.Element
		for curr != nil {
			if el, ok := curr.(dom.Element); ok {
				if el.TagName() == "A" && (el.HasAttribute("data-router-link") || el.GetAttribute("data-router-link") == "true") {
					anchor = el
					break
				}
			}
			curr = curr.ParentNode()
		}

		if anchor == nil {
			return
		}

		// Prevent default browser navigation
		event.PreventDefault()
		// Stop propagation to avoid other handlers triggering full navigation
		event.StopPropagation()

		// Get the href attribute
		href := anchor.GetAttribute("href")
		if href == "" {
			return
		}
		// Navigate using the router
		router.Navigate(href, NavigateOptions{})
	})
}