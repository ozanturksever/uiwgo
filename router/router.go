package router

import (
	"github.com/ozanturksever/uiwgo/logutil"
	"strings"
)

// currentRouter holds the most recently created router instance.
// This is used by the A function to navigate.
var currentRouter *Router

// NavigateOptions contains options for programmatic navigation.
type NavigateOptions struct {
	State any // State data to associate with the navigation
}

// Router holds a collection of route definitions and provides matching functionality.
type Router struct {
	routes        []*RouteDefinition
	outlet        any
	locationState *LocationState
	currentRoute  *RouteDefinition
	currentParams map[string]string
	// Optional navigation callbacks for integration (e.g., AppManager)
	OnBeforeNavigate func(path string, options NavigateOptions)
	OnAfterNavigate  func(path string, options NavigateOptions)
	// WASM-specific navigation function
	navigateWASM func(path string, options NavigateOptions)
}

// New creates a new Router instance with the provided routes and outlet.
// The outlet parameter is any to accommodate mocks for testing.
func New(routes []*RouteDefinition, outlet any) *Router {
	router := &Router{
		routes:        routes,
		outlet:        outlet,
		locationState: NewLocationState(),
	}
	// Set this as the current router for navigation
	currentRouter = router

	// Set initial location to root path
	router.locationState.Set(Location{
		Pathname: "/",
		Search:   "",
		Hash:     "",
		State:    nil,
	})

	// Setup WASM-specific functionality if available
	setupWASM(router)

	return router
}

// Match iterates through the router's routes in order and returns the first route
// that matches the given path, along with any captured parameters.
// For nested routes, it returns the deepest matching child route and accumulates
// parameters from all parent routes in the hierarchy.
// If no route matches, it returns (nil, nil).
func (r *Router) Match(path string) (*RouteDefinition, map[string]string) {
	matchedRoute, params := r.matchRecursive(path, r.routes, make(map[string]string))
	
	// Store the matched route and parameters for later access via Params()
	r.currentRoute = matchedRoute
	r.currentParams = params
	
	return matchedRoute, params
}

// matchRecursive performs recursive route matching for nested routes.
// It tries to match the path against routes at the current level, and if a route matches,
// it attempts to match the remaining path against the route's children.
// Returns the deepest matching route and accumulated parameters from the entire hierarchy.
func (r *Router) matchRecursive(path string, routes []*RouteDefinition, accumulatedParams map[string]string) (*RouteDefinition, map[string]string) {
	for _, route := range routes {
		logutil.Logf("Trying route: %s for path: %s", route.Path, path)
		if route.matcher == nil {
			route.matcher = compileMatcher(route)
		}
		isMatch, params := route.matcher(path)
		if isMatch {
			logutil.Logf("Matched route: %s with params: %v", route.Path, params)
			// Merge parameters from this route with accumulated parameters
			mergedParams := make(map[string]string)
			for k, v := range accumulatedParams {
				mergedParams[k] = v
			}
			for k, v := range params {
				mergedParams[k] = v
			}
			
			// Calculate remaining path after this route match
			remainingPath := calculateRemainingPath(path, route.Path, params)
			
			// Always check children first, even if remaining path is empty
			// This allows child routes with path "/" to match empty remaining paths
			if len(route.Children) > 0 {
			logutil.Logf("Checking children for route: %s with remaining: %s", route.Path, remainingPath)
			childRoute, childParams := r.matchRecursive(remainingPath, route.Children, mergedParams)
			if childRoute != nil {
				logutil.Logf("Found child match for %s, returning %s", route.Path, childRoute.Path)
				// Found a matching child, return it
				return childRoute, childParams
			}
			logutil.Logf("No child match for %s", route.Path)
		}
			
			// If no children matched and remaining path is empty, this route is the match
			if remainingPath == "" {
			logutil.Logf("Exact match for route: %s", route.Path)
			return route, mergedParams
		}
			
			// There's remaining path but no matching children - this is not a valid match
			// Continue to try other routes at this level
		}
	}
	
	logutil.Log("No match found at this level")
	// No match found at this level
	return nil, nil
}

// calculateRemainingPath determines what part of the original path remains
// after a route has been matched. This is used for nested route matching.
// For example:
//   - path: "/users/123/posts/456", routePath: "/users/:userId" -> remaining: "/posts/456"
//   - path: "/admin", routePath: "/admin" -> remaining: ""
//   - path: "/users/123", routePath: "/users/:userId" -> remaining: ""
func calculateRemainingPath(originalPath, routePath string, params map[string]string) string {
	// Split paths into segments
	originalSegments := splitPath(originalPath)
	routeSegments := splitPath(routePath)
	
	// Check if the route has a wildcard segment (starts with *)
	for _, segment := range routeSegments {
		if strings.HasPrefix(segment, "*") {
			// Wildcard routes consume all remaining path
			return ""
		}
	}
	
	// If route has more segments than original path, no remaining path
	if len(routeSegments) >= len(originalSegments) {
		return ""
	}
	
	// Calculate remaining segments
	remainingSegments := originalSegments[len(routeSegments):]
	
	// Join remaining segments back into a path
	if len(remainingSegments) == 0 {
		return ""
	}
	
	remainingPath := "/" + joinSegments(remainingSegments)
	return remainingPath
}

// splitPath splits a path into segments, removing empty segments
func splitPath(path string) []string {
	if path == "/" || path == "" {
		return []string{}
	}
	
	// Remove leading slash and split
	if path[0] == '/' {
		path = path[1:]
	}
	
	segments := []string{}
	for _, segment := range strings.Split(path, "/") {
		if segment != "" {
			segments = append(segments, segment)
		}
	}
	return segments
}

// joinSegments joins path segments with "/" separator
func joinSegments(segments []string) string {
	if len(segments) == 0 {
		return ""
	}
	return strings.Join(segments, "/")
}

// Route creates a new RouteDefinition with the specified path, component, and children.
// This is a builder function for defining routes in a declarative way.
func Route(path string, component func(props ...any) interface{}, children ...*RouteDefinition) *RouteDefinition {
	rd := &RouteDefinition{
		Path:         path,
		Component:    component,
		Children:     children,
		MatchFilters: make(map[string]any),
	}
	rd.matcher = compileMatcher(rd)
	return rd
}

// Location returns the current Location from the router's internal LocationState.
// This provides access to the current routing state including pathname, search, hash, and state.
func (r *Router) Location() Location {
	return r.locationState.Get()
}

// Params returns the parameters captured from the most recent successful route match.
// If no route has been matched or no parameters were captured, it returns an empty map.
func (r *Router) Params() map[string]string {
	if r.currentParams == nil {
		return make(map[string]string)
	}
	return r.currentParams
}

// Navigate performs programmatic navigation to the specified path.
// It calls the un-exported navigate method on the router instance.
func (r *Router) Navigate(path string, opts ...NavigateOptions) {
	options := NavigateOptions{}
	if len(opts) > 0 {
		options = opts[0]
	}
	r.navigate(path, options)
}

// navigate updates the router's location state to the new path.
// This is an unexported method that will be called by the A component's OnClick handler.
func (r *Router) navigate(path string, options NavigateOptions) {
	// Notify before navigation
	if r.OnBeforeNavigate != nil {
		r.OnBeforeNavigate(path, options)
	}

	// Use WASM-specific navigation if available
	if r.navigateWASM != nil {
		r.navigateWASM(path, options)
	} else {
		// Fallback for non-WASM builds
		// Create a new Location with the given path
		newLocation := Location{
			Pathname: path,
			Search:   "", // You might want to handle search and hash later
			Hash:     "",
			State:    options.State,
		}
		// Update the location state
		r.locationState.Set(newLocation)
	}

	// Notify after navigation
	if r.OnAfterNavigate != nil {
		r.OnAfterNavigate(path, options)
	}
}
