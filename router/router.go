package router

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

	// Setup WASM-specific functionality if available
	setupWASM(router)

	return router
}

// Match iterates through the router's routes in order and returns the first route
// that matches the given path, along with any captured parameters.
// If no route matches, it returns (nil, nil).
func (r *Router) Match(path string) (*RouteDefinition, map[string]string) {
	for _, route := range r.routes {
		if route.matcher == nil {
			route.matcher = buildMatcher(route.Path)
		}
		isMatch, params := route.matcher(path)
		if isMatch {
			// Store the matched route and parameters for later access via Params()
			r.currentRoute = route
			r.currentParams = params
			return route, params
		}
	}
	// No match found, clear the current route and params
	r.currentRoute = nil
	r.currentParams = nil
	return nil, nil
}

// buildMatcher creates a matcher function for a given route path.
func buildMatcher(path string) func(string) (bool, map[string]string) {
	// This is a simplified matcher that only handles static paths for now.
	return func(url string) (bool, map[string]string) {
		return path == url, nil
	}
}

// Route creates a new RouteDefinition with the specified path, component, and children.
// This is a builder function for defining routes in a declarative way.
func Route(path string, component func(props ...any) interface{}, children ...*RouteDefinition) *RouteDefinition {
	return &RouteDefinition{
		Path:      path,
		Component: component,
		Children:  children,
		matcher:   buildMatcher(path),
	}
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
