//go:build js && wasm

package router

import (
	"bytes"
	"syscall/js"

	"github.com/ozanturksever/uiwgo/logutil"
	dom "honnef.co/go/js/dom/v2"
	. "maragu.dev/gomponents"
)

// setupWASM initializes WASM-specific functionality for the router.
func setupWASM(router *Router) {
	// Set up WASM-specific navigation function
	router.navigateWASM = router.navigateWASMImpl

	// Expose the router's location to JavaScript for testing
	exposeRouterToJS(router)

	// Add popstate event listener to handle browser history navigation
	addPopstateEventListener(router)

	// Set up navigation link handling
	setupNavigationWASM(router)

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
	logutil.Log("Exposing router to JS");
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
		logutil.Log("__router set with location and Navigate");
	}

// renderLocation renders the appropriate component for the given location.
// This implements the destructive-and-replace rendering strategy as specified in the design.
// For nested routes, it composes parent and child components using the layout pattern.
func renderLocation(router *Router, location Location) {
	currentPath := location.Pathname
	logutil.Logf("Rendering location: %s", currentPath)

	// Match the route for the current path
	matchedRoute, params := router.Match(currentPath)
	if matchedRoute == nil {
		logutil.Logf("No route matched for path: %s, trying catch-all route", currentPath)
		// Try to find a catch-all route (wildcard route with "*")
		for _, route := range router.routes {
			if route.Path == "*" {
				matchedRoute = route
				params = make(map[string]string)
				logutil.Logf("Using catch-all route for path: %s", currentPath)
				break
			}
		}
		if matchedRoute == nil {
			logutil.Logf("No catch-all route found for path: %s", currentPath)
			return
		}
	}

	// Update router's current state
	router.currentRoute = matchedRoute
	router.currentParams = params

	// Build the component hierarchy for nested routes
	componentNode := buildComponentHierarchy(router, currentPath, matchedRoute, params)
	if componentNode == nil {
		logutil.Log("Failed to build component hierarchy")
		return
	}

	// Render the node to HTML string using bytes.Buffer (destructive-and-replace strategy)
	var buf bytes.Buffer
	err := componentNode.Render(&buf)
	if err != nil {
		logutil.Logf("Error rendering component to HTML: %v", err)
		return
	}
	htmlString := buf.String()
	logutil.Logf("Generated HTML (%d chars)", len(htmlString))

	// Get the outlet element and set its innerHTML
	outlet, ok := router.outlet.(dom.Element)
	if !ok {
		logutil.Logf("Outlet is not a DOM element, got: %T", router.outlet)
		return
	}

	// Perform destructive-and-replace DOM update
	outlet.SetInnerHTML(htmlString)
	logutil.Log("DOM updated successfully")
}

// buildComponentHierarchy constructs the component hierarchy for nested routes.
// It traverses the route tree from root to the matched route, composing parent
// components with their child content according to the layout pattern.
func buildComponentHierarchy(router *Router, originalPath string, matchedRoute *RouteDefinition, params map[string]string) Node {
	// Find the route hierarchy from root to the matched route
	routeHierarchy := findRouteHierarchy(router.routes, originalPath, matchedRoute)
	if len(routeHierarchy) == 0 {
		logutil.Log("No route hierarchy found")
		return nil
	}

	logutil.Logf("Building component hierarchy with %d levels", len(routeHierarchy))

	// Start from the deepest (matched) route and work backwards
	var currentNode Node

	// Render the deepest route first
	deepestRoute := routeHierarchy[len(routeHierarchy)-1]
	logutil.Logf("Rendering deepest component for route: %s", deepestRoute.Path)
	componentResult := deepestRoute.Component(params)
	if componentResult == nil {
		logutil.Log("Deepest component function returned nil")
		return nil
	}

	var ok bool
	currentNode, ok = componentResult.(Node)
	if !ok {
		logutil.Logf("Deepest component did not return a gomponents.Node, got: %T", componentResult)
		return nil
	}

	// Work backwards through parent routes, passing child content as props
	for i := len(routeHierarchy) - 2; i >= 0; i-- {
		parentRoute := routeHierarchy[i]
		logutil.Logf("Composing parent component for route: %s", parentRoute.Path)
		
		// Call parent component with child node as first argument
		parentResult := parentRoute.Component(currentNode, params)
		if parentResult == nil {
			logutil.Logf("Parent component for route %s returned nil", parentRoute.Path)
			return nil
		}

		parentNode, ok := parentResult.(Node)
		if !ok {
			logutil.Logf("Parent component for route %s did not return a gomponents.Node, got: %T", parentRoute.Path, parentResult)
			return nil
		}

		currentNode = parentNode
	}

	return currentNode
}

// findRouteHierarchy finds the complete route hierarchy from root to the matched route.
// It returns a slice of RouteDefinition pointers representing the path from root to leaf.
func findRouteHierarchy(routes []*RouteDefinition, originalPath string, targetRoute *RouteDefinition) []*RouteDefinition {
	for _, route := range routes {
		if route == targetRoute {
			// Found the target route at this level
			return []*RouteDefinition{route}
		}

		// Check if this route matches the beginning of the path
		if route.matcher == nil {
			route.matcher = compileMatcher(route)
		}
		isMatch, params := route.matcher(originalPath)
		if isMatch {
			// Calculate remaining path and search in children
			remainingPath := calculateRemainingPath(originalPath, route.Path, params)
			if len(route.Children) > 0 {
				childHierarchy := findRouteHierarchy(route.Children, remainingPath, targetRoute)
				if len(childHierarchy) > 0 {
					// Found target in children, prepend this route
					return append([]*RouteDefinition{route}, childHierarchy...)
				}
			}
		}
	}
	return nil
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
	// Update the router's location state to match the current URL
	router.locationState.Set(location)
	renderLocation(router, location)
}

// updateJSLocation updates the global JavaScript variable with the current location.
func updateJSLocation(location Location) {
	routerObj := js.Global().Get("__router")
	if !routerObj.Truthy() {
		logutil.Log("Error: __router not found for update")
		return
	}
	locationObj := js.ValueOf(map[string]interface{}{
		"pathname": location.Pathname,
		"search":   location.Search,
		"hash":     location.Hash,
		"state":    location.State,
	})
	routerObj.Set("location", locationObj)
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
