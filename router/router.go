//go:build js && wasm

package router

import (
	"net/url"
	"path"
	"regexp"
	"strings"
	"syscall/js"

	"github.com/ozanturksever/uiwgo/logutil"
	"github.com/ozanturksever/uiwgo/reactivity"
	g "maragu.dev/gomponents"
	"honnef.co/go/js/dom/v2"
)

// Node is an alias for gomponents.Node for convenience.
type Node = g.Node

// Location represents the current location information
type Location struct {
	Pathname string            // The path portion of the URL
	Search   string            // The query string portion
	Hash     string            // The hash portion
	State    any               // Associated state data
	Query    map[string]string // Parsed query parameters
}

// RouteMatch contains information about a matched route
type RouteMatch struct {
	Path     string            // The original route pattern
	Params   map[string]string // Extracted path parameters
	Query    map[string]string // Query parameters
	Wildcard string            // Wildcard match if present
}

// Router manages the application's routing state and navigation
type Router struct {
	location    reactivity.Signal[*Location]
	routes      []*RouteConfig
	basePath    string
	beforeGuard func(to *Location) bool
	afterGuard  func(from, to *Location)
	disposed    bool
	popstateHandler js.Func
}

// RouteConfig defines a route configuration
type RouteConfig struct {
	Path      string                           // Route pattern (e.g., "/users/:id", "/posts/*")
	Component func(match *RouteMatch) Node     // Component to render
	Guard     func(match *RouteMatch) bool     // Route guard function
	Children  []*RouteConfig                   // Nested routes
	Redirect  string                           // Redirect path
	regexp    *regexp.Regexp                   // Compiled pattern
	paramKeys []string                         // Parameter names
}

// NavigateOptions contains options for navigation
type NavigateOptions struct {
	Replace bool // Whether to replace the current history entry
	State   any  // State to associate with the navigation
}

// Global router instance
var globalRouter *Router

// CreateRouter creates a new router instance and sets it as the global router
func CreateRouter(basePath string) *Router {
	if globalRouter != nil && !globalRouter.disposed {
		globalRouter.Dispose()
	}

	// Get current location from browser
	currentLoc := getCurrentLocation()
	
	router := &Router{
		location:    reactivity.CreateSignal(currentLoc),
		routes:      make([]*RouteConfig, 0),
		basePath:    strings.TrimSuffix(basePath, "/"),
	}

	// Set up popstate event listener
	router.setupPopstateListener()
	
	globalRouter = router
	logutil.Log("Router created with base path:", basePath)
	
	return router
}

// GetRouter returns the global router instance
func GetRouter() *Router {
	if globalRouter == nil {
		return CreateRouter("")
	}
	return globalRouter
}

// setupPopstateListener sets up the browser popstate event listener
func (r *Router) setupPopstateListener() {
	r.popstateHandler = js.FuncOf(func(this js.Value, args []js.Value) any {
		if r.disposed {
			return nil
		}
		
		// Update location signal when browser navigation occurs
		newLoc := getCurrentLocation()
		r.location.Set(newLoc)
		logutil.Log("Popstate navigation to:", newLoc.Pathname)
		
		// Call after guard if present
		if r.afterGuard != nil {
			oldLoc := r.location.Get()
			r.afterGuard(oldLoc, newLoc)
		}
		
		return nil
	})
	
	// Use syscall/js directly for event listener since dom/v2 doesn't have Underlying
	js.Global().Get("window").Call("addEventListener", "popstate", r.popstateHandler)
}

// getCurrentLocation gets the current browser location
func getCurrentLocation() *Location {
	location := dom.GetWindow().Location()
	
	// Parse query parameters
	query := make(map[string]string)
	searchQuery := location.Search()
	if searchQuery != "" {
		if parsed, err := url.ParseQuery(strings.TrimPrefix(searchQuery, "?")); err == nil {
			for key, values := range parsed {
				if len(values) > 0 {
					query[key] = values[0]
				}
			}
		}
	}
	
	return &Location{
		Pathname: location.Pathname(),
		Search:   location.Search(),
		Hash:     location.Hash(),
		Query:    query,
		State:    nil, // Browser state not directly accessible
	}
}

// AddRoute adds a route configuration to the router
func (r *Router) AddRoute(config *RouteConfig) {
	if config.Path == "" {
		logutil.Log("Warning: Empty route path provided")
		return
	}
	
	// Compile the route pattern
	r.compileRoute(config)
	r.routes = append(r.routes, config)
	logutil.Log("Route added:", config.Path)
}

// compileRoute compiles a route pattern into a regular expression
func (r *Router) compileRoute(config *RouteConfig) {
	pattern := config.Path
	config.paramKeys = make([]string, 0)
	
	// Escape special regex characters
	pattern = regexp.MustCompile(`[.+?^${}()|[\]\\]`).ReplaceAllString(pattern, `\$0`)
	
	// Replace parameters (:param) with regex groups
	paramRegex := regexp.MustCompile(`:([a-zA-Z_][a-zA-Z0-9_]*)`)
	pattern = paramRegex.ReplaceAllStringFunc(pattern, func(match string) string {
		// Extract parameter name from the submatch
		submatches := paramRegex.FindStringSubmatch(match)
		if len(submatches) > 1 {
			paramName := submatches[1]
			config.paramKeys = append(config.paramKeys, paramName)
		}
		return `([^/]+)`
	})
	
	// Replace wildcards (*) with regex groups - require at least one character
	// Handle both /* and * patterns
	wildcardRegex := regexp.MustCompile(`\*`)
	pattern = wildcardRegex.ReplaceAllStringFunc(pattern, func(match string) string {
		// For wildcard patterns, we want to match one or more characters
		// but only if there's content after the base path
		return `(.+)`
	})
	
	// Anchor the pattern
	pattern = `^` + pattern + `$`
	
	config.regexp = regexp.MustCompile(pattern)
}

// Navigate navigates to a new path
func (r *Router) Navigate(to string, options ...*NavigateOptions) {
	opts := &NavigateOptions{}
	if len(options) > 0 && options[0] != nil {
		opts = options[0]
	}
	
	// Resolve relative to base path
	fullPath := r.resolvePath(to)
	
	// Create new location
	newLoc := &Location{
		Pathname: fullPath,
		Search:   "",
		Hash:     "",
		Query:    make(map[string]string),
		State:    opts.State,
	}
	
	// Parse URL to extract search and hash
	if parsed, err := url.Parse(fullPath); err == nil {
		newLoc.Pathname = parsed.Path
		newLoc.Search = parsed.RawQuery
		newLoc.Hash = parsed.Fragment
		
		// Parse query parameters
		if parsed.RawQuery != "" {
			if queryValues, err := url.ParseQuery(parsed.RawQuery); err == nil {
				for key, values := range queryValues {
					if len(values) > 0 {
						newLoc.Query[key] = values[0]
					}
				}
			}
		}
	}
	
	// Check before guard
	if r.beforeGuard != nil && !r.beforeGuard(newLoc) {
		logutil.Log("Navigation blocked by guard:", to)
		return
	}
	
	// Update browser history
	history := js.Global().Get("history")
	
	if opts.Replace {
		history.Call("replaceState", opts.State, "", fullPath)
	} else {
		history.Call("pushState", opts.State, "", fullPath)
	}
	
	// Update location signal
	oldLoc := r.location.Get()
	r.location.Set(newLoc)
	
	// Call after guard
	if r.afterGuard != nil {
		r.afterGuard(oldLoc, newLoc)
	}
}

// resolvePath resolves a path relative to the router's base path
func (r *Router) resolvePath(to string) string {
	if strings.HasPrefix(to, "/") {
		// Absolute path - just prepend base path if not already present
		if r.basePath == "" || strings.HasPrefix(to, r.basePath) {
			return to
		}
		return r.basePath + to
	}
	// Relative path - join with current location
	currentPath := r.location.Get().Pathname
	return path.Join(currentPath, to)
}

// GetLocation returns the current location signal
func (r *Router) GetLocation() reactivity.Signal[*Location] {
	return r.location
}

// MatchRoute finds the first route that matches the current location
func (r *Router) MatchRoute(pathname string) (*RouteConfig, *RouteMatch) {
	// Remove base path
	cleanPath := strings.TrimPrefix(pathname, r.basePath)
	if cleanPath == "" {
		cleanPath = "/"
	}
	
	for _, route := range r.routes {
		if match := r.matchSingleRoute(route, cleanPath); match != nil {
			return route, match
		}
	}
	
	return nil, nil
}

// matchSingleRoute checks if a single route matches the given path
func (r *Router) matchSingleRoute(route *RouteConfig, pathname string) *RouteMatch {
	if route.regexp == nil {
		return nil
	}
	
	matches := route.regexp.FindStringSubmatch(pathname)
	if matches == nil {
		return nil
	}
	
	// Extract parameters
	params := make(map[string]string)
	for i, paramName := range route.paramKeys {
		if i+1 < len(matches) {
			params[paramName] = matches[i+1]
		}
	}
	
	// Handle wildcard - check if route has wildcard pattern
	wildcard := ""
	if strings.Contains(route.Path, "*") {
		// Wildcard is the last capture group
		if len(matches) > len(route.paramKeys)+1 {
			wildcard = matches[len(matches)-1]
		}
	}
	
	// Get current query parameters
	currentLoc := r.location.Get()
	
	return &RouteMatch{
		Path:     route.Path,
		Params:   params,
		Query:    currentLoc.Query,
		Wildcard: wildcard,
	}
}

// SetBeforeGuard sets a global navigation guard that runs before navigation
func (r *Router) SetBeforeGuard(guard func(to *Location) bool) {
	r.beforeGuard = guard
}

// SetAfterGuard sets a global navigation guard that runs after navigation
func (r *Router) SetAfterGuard(guard func(from, to *Location)) {
	r.afterGuard = guard
}

// Dispose cleans up the router and removes event listeners
func (r *Router) Dispose() {
	if r.disposed {
		return
	}
	
	r.disposed = true
	
	// Remove popstate listener
	if !r.popstateHandler.IsUndefined() {
		js.Global().Get("window").Call("removeEventListener", "popstate", r.popstateHandler)
		r.popstateHandler.Release()
	}
	
	logutil.Log("Router disposed")
}