//go:build js && wasm

package router

import (
	"bytes"
	"syscall/js"

	dom "honnef.co/go/js/dom/v2"
	. "maragu.dev/gomponents"
)

// setupWASM initializes WASM-specific functionality for the router.
func setupWASM(router *Router) {
	// Expose the router's location to JavaScript for testing
	exposeRouterToJS(router)

	// Add popstate event listener to handle browser history navigation
	addPopstateEventListener(router)

	// Subscribe to location state changes for reactive rendering
	router.locationState.Subscribe(func(newLocation Location) {
		renderLocation(router, newLocation)
		// Update the JavaScript global variable after rendering
		updateJSLocation(newLocation)
	})

	// Perform initial render based on current URL
	performInitialRender(router)
}

// exposeRouterToJS sets up a global JavaScript variable to access the router's location and navigation for testing.
func exposeRouterToJS(router *Router) {
	// Create a JavaScript object to represent the router
	js.Global().Set("__router", map[string]interface{}{
		"location": map[string]interface{}{
			"pathname": router.locationState.Get().Pathname,
			"search":   router.locationState.Get().Search,
			"hash":     router.locationState.Get().Hash,
			"state":    router.locationState.Get().State,
		},
		"Navigate": js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			if len(args) < 1 {
				return js.ValueOf("Error: path argument required")
			}
			path := args[0].String()
			router.Navigate(path, NavigateOptions{})
			return js.ValueOf(nil)
		}),
	})
}

// renderLocation renders the appropriate component for the given location.
func renderLocation(router *Router, location Location) {
	currentPath := location.Pathname

	// Match the route for the current path
	matchedRoute, params := router.Match(currentPath)
	if matchedRoute == nil {
		// No route matched, could render a 404 or default component
		return
	}

	// Call the component function with captured parameters
	componentResult := matchedRoute.Component(params)
	if componentResult == nil {
		return
	}

	// Convert the component result to a gomponents Node
	componentNode, ok := componentResult.(Node)
	if !ok {
		// Handle error: component did not return a Node
		return
	}

	// Render the node to HTML string using bytes.Buffer
	var buf bytes.Buffer
	err := componentNode.Render(&buf)
	if err != nil {
		// Handle rendering error
		return
	}
	htmlString := buf.String()

	// Get the outlet element and set its innerHTML
	outlet, ok := router.outlet.(dom.Element)
	if !ok {
		// outlet is not a DOM element, cannot render
		return
	}

	outlet.SetInnerHTML(htmlString)
}

// performInitialRender renders the initial component based on the current browser URL.
func performInitialRender(router *Router) {
	window := dom.GetWindow()
	currentLocation := window.Location()
	location := Location{
		Pathname: currentLocation.Pathname(),
		Search:   currentLocation.Search(),
		Hash:     currentLocation.Hash(),
		State:    nil, // Initial state is nil
	}
	renderLocation(router, location)
}

// updateJSLocation updates the global JavaScript variable with the current location.
func updateJSLocation(location Location) {
	js.Global().Set("__router", map[string]interface{}{
		"location": map[string]interface{}{
			"pathname": location.Pathname,
			"search":   location.Search,
			"hash":     location.Hash,
			"state":    location.State,
		},
	})
}

// addPopstateEventListener adds an event listener for popstate events to handle browser history navigation.
func addPopstateEventListener(router *Router) {
	window := dom.GetWindow()
	window.AddEventListener("popstate", false, func(event dom.Event) {
		// Get the current location from the browser
		currentLocation := window.Location()
		// Create a Location object from the browser's location
		newLocation := Location{
			Pathname: currentLocation.Pathname(),
			Search:   currentLocation.Search(),
			Hash:     currentLocation.Hash(),
			State:    nil, // popstate event doesn't carry state, it's in the history state
		}
		// Update the router's location state
		router.locationState.Set(newLocation)
		// Also update the JavaScript global variable
		updateJSLocation(newLocation)
	})
}
