// router.go
// Client-side routing system for single-page applications

package golid

import (
	"strings"
	"syscall/js"

	"maragu.dev/gomponents"
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

// -------------------------
// 🧭 Router System
// -------------------------

// Global router instance
var globalRouter *Router

// NewRouter creates a new router instance
func NewRouter() *Router {
	r := &Router{
		routes:        make([]Route, 0),
		currentPath:   NewSignal(""),
		currentParams: NewSignal(make(RouteParams)),
		currentQuery:  NewSignal(make(map[string]string)),
		basePath:      "",
		history:       js.Global().Get("window").Get("history"),
	}

	// Set up popstate event listener
	js.Global().Get("window").Call("addEventListener", "popstate", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		r.handleURLChange()
		return nil
	}))

	// Initialize with current URL
	r.handleURLChange()

	return r
}

// SetGlobalRouter sets the global router instance
func SetGlobalRouter(r *Router) {
	globalRouter = r
}

// GetCurrentPath returns the current path signal
func (r *Router) GetCurrentPath() *Signal[string] {
	return r.currentPath
}

// GetCurrentParams returns the current params signal
func (r *Router) GetCurrentParams() *Signal[RouteParams] {
	return r.currentParams
}

// GetCurrentQuery returns the current query signal
func (r *Router) GetCurrentQuery() *Signal[map[string]string] {
	return r.currentQuery
}

// AddRoute adds a route to the router
func (r *Router) AddRoute(path string, handler func(RouteParams) gomponents.Node) {
	r.routes = append(r.routes, Route{
		Path:    path,
		Handler: handler,
	})
}

// AddRouteWithGuard adds a route with a guard function
func (r *Router) AddRouteWithGuard(path string, handler func(RouteParams) gomponents.Node, guard func(RouteParams) bool) {
	r.routes = append(r.routes, Route{
		Path:    path,
		Handler: handler,
		Guard:   guard,
	})
}

// Navigate programmatically navigates to a path
func (r *Router) Navigate(path string) {
	fullPath := r.basePath + path
	r.history.Call("pushState", nil, "", fullPath)
	r.handleURLChange()
}

// Replace replaces current history entry
func (r *Router) Replace(path string) {
	fullPath := r.basePath + path
	r.history.Call("replaceState", nil, "", fullPath)
	r.handleURLChange()
}

// handleURLChange processes URL changes
func (r *Router) handleURLChange() {
	location := js.Global().Get("window").Get("location")
	pathname := location.Get("pathname").String()
	search := location.Get("search").String()

	// Remove base path if present
	if r.basePath != "" && strings.HasPrefix(pathname, r.basePath) {
		pathname = strings.TrimPrefix(pathname, r.basePath)
	}

	// Parse query parameters
	query := parseQueryString(search)

	// Find matching route
	match := r.matchRoute(pathname)

	// Update signals
	r.currentPath.Set(pathname)
	r.currentQuery.Set(query)

	if match != nil {
		r.currentParams.Set(match.Params)
	} else {
		r.currentParams.Set(make(RouteParams))
	}
}

// matchRoute finds a matching route for the given path
func (r *Router) matchRoute(path string) *RouteMatch {
	for _, route := range r.routes {
		if params := r.matchPath(route.Path, path); params != nil {
			// Check guard if present
			if route.Guard != nil && !route.Guard(*params) {
				continue
			}

			return &RouteMatch{
				Path:    path,
				Params:  *params,
				Handler: route.Handler,
			}
		}
	}
	return nil
}

// matchPath checks if a path matches a route pattern and extracts parameters
func (r *Router) matchPath(pattern, path string) *RouteParams {
	patternParts := strings.Split(strings.Trim(pattern, "/"), "/")
	pathParts := strings.Split(strings.Trim(path, "/"), "/")

	// Handle root path
	if pattern == "/" && path == "/" {
		params := make(RouteParams)
		return &params
	}

	if len(patternParts) != len(pathParts) {
		return nil
	}

	params := make(RouteParams)

	for i, patternPart := range patternParts {
		if patternPart == "" && pathParts[i] == "" {
			continue // Both empty, skip
		}

		if strings.HasPrefix(patternPart, ":") {
			// Parameter
			paramName := strings.TrimPrefix(patternPart, ":")
			params[paramName] = pathParts[i]
		} else if patternPart != pathParts[i] {
			// Literal match failed
			return nil
		}
	}

	return &params
}

// parseQueryString parses URL query string into a map
func parseQueryString(query string) map[string]string {
	params := make(map[string]string)
	if query == "" {
		return params
	}

	// Remove leading ?
	query = strings.TrimPrefix(query, "?")

	pairs := strings.Split(query, "&")
	for _, pair := range pairs {
		if pair == "" {
			continue
		}

		parts := strings.SplitN(pair, "=", 2)
		key := parts[0]
		value := ""
		if len(parts) > 1 {
			value = parts[1]
		}

		params[key] = value
	}

	return params
}

// RouterOutlet renders the current route
func RouterOutlet() Node {
	if globalRouter == nil {
		return Div(Text("Router not initialized"))
	}

	return Bind(func() Node {
		path := globalRouter.currentPath.Get()
		params := globalRouter.currentParams.Get()

		match := globalRouter.matchRoute(path)
		if match != nil {
			return match.Handler(params)
		}

		// Default 404 page
		return Div(
			H1(Text("404 - Page Not Found")),
			P(Text("The page you're looking for doesn't exist.")),
			A(Href("/"), Text("Go Home")),
		)
	})
}

// RouterLink creates a navigation link that uses the router
func RouterLink(href string, children ...Node) Node {
	id := GenID()

	// Register click handler with global observer
	globalObserver.RegisterElement(id, func() {
		elem := NodeFromID(id)
		if elem.Truthy() {
			elem.Call("addEventListener", "click", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
				event := args[0]
				// Prevent default browser navigation
				event.Call("preventDefault")

				// Use router for client-side navigation
				if globalRouter != nil {
					globalRouter.Navigate(href)
				}
				return nil
			}))
		}
	})

	return A(
		Attr("id", id),
		Href(href), // Keep href for accessibility/SEO
		Style("text-decoration: none; color: inherit; cursor: pointer;"),
		Group(children),
	)
}

// UseParams returns the current route parameters
func UseParams() *Signal[RouteParams] {
	if globalRouter == nil {
		return NewSignal(make(RouteParams))
	}
	return globalRouter.currentParams
}

// UseQuery returns the current query parameters
func UseQuery() *Signal[map[string]string] {
	if globalRouter == nil {
		return NewSignal(make(map[string]string))
	}
	return globalRouter.currentQuery
}

// UseLocation returns the current path
func UseLocation() *Signal[string] {
	if globalRouter == nil {
		return NewSignal("")
	}
	return globalRouter.currentPath
}
