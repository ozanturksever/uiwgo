//go:build js && wasm

package router

import (
	"github.com/ozanturksever/uiwgo/comps"
	"github.com/ozanturksever/uiwgo/reactivity"
	g "maragu.dev/gomponents"
	a "maragu.dev/gomponents/html"
)

// RouterProps defines properties for the Router component
type RouterProps struct {
	BasePath    string                        // Base path for the router
	BeforeGuard func(to *Location) bool       // Global before navigation guard
	AfterGuard  func(from, to *Location)      // Global after navigation guard
	Children    []Node                        // Child Route components
}

// RouteProps defines properties for individual Route components
type RouteProps struct {
	Path      string                           // Route pattern
	Component func(match *RouteMatch) Node     // Component to render
	Guard     func(match *RouteMatch) bool     // Route-specific guard
	Redirect  string                           // Redirect path
	Children  []Node                           // Nested routes
}

// LinkProps defines properties for Link components
type LinkProps struct {
	To       string                     // Target path
	Replace  bool                       // Whether to replace history entry
	State    any                        // State to pass with navigation
	Class    string                     // CSS class
	Style    string                     // Inline styles
	Children []Node                     // Link content
	OnClick  func()                     // Additional click handler
}

// NavigateProps defines properties for programmatic navigation
type NavigateProps struct {
	To      string  // Target path
	Replace bool    // Whether to replace history entry
	State   any     // State to pass
	When    func() bool // Conditional navigation
}

// RouterProvider component creates and manages the routing context
func RouterProvider(props RouterProps) Node {
	// Get the existing global router instance
	router := GetRouter()
	
	// Set the base path if provided
	if props.BasePath != "" {
		router.basePath = props.BasePath
	}
	
	// Set up guards
	if props.BeforeGuard != nil {
		router.SetBeforeGuard(props.BeforeGuard)
	}
	if props.AfterGuard != nil {
		router.SetAfterGuard(props.AfterGuard)
	}
	
	// Process child Route components to extract route configurations
	// In a real implementation, we would need to extract RouteProps
	// from the child nodes. For now, this is a placeholder.
	// Routes should be added via router.AddRoute() calls
	
	// Return the children wrapped in a router context
	// The children will contain RouterOutlet() which will render matched routes
	return g.Group(props.Children)
}

// RouterOutlet renders the currently matched route
func RouterOutlet() Node {
	router := GetRouter()
	
	// Create a reactive outlet that updates when location changes
	return comps.BindHTML(func() Node {
		// Get current location and render matched route
		loc := router.GetLocation().Get()
		
		// Find matching route
		route, match := router.MatchRoute(loc.Pathname)
		
		var content Node
		
		if route == nil {
			// No route matched - render 404
			content = g.Text("404 - Route not found: " + loc.Pathname)
		} else if route.Redirect != "" {
			// Handle redirects
			router.Navigate(route.Redirect, &NavigateOptions{Replace: true})
			content = g.Text("")
		} else if route.Guard != nil && !route.Guard(match) {
			// Check route guard
			content = g.Text("Access Denied")
		} else if route.Component != nil {
			// Render the route component
			content = route.Component(match)
		} else {
			content = g.Text("No component defined for route: " + route.Path)
		}
		
		// Wrap in router outlet div
		return a.Div(
			a.DataAttr("router-outlet", "true"),
			content,
		)
	})
}

// Route is a helper function to register routes with the global router
func Route(props RouteProps) {
	router := GetRouter()
	
	config := &RouteConfig{
		Path:      props.Path,
		Component: props.Component,
		Guard:     props.Guard,
		Redirect:  props.Redirect,
	}
	
	router.AddRoute(config)
}

// Link creates a navigation link component
func Link(props LinkProps) Node {
	return a.A(
		a.Href(props.To),
		g.If(props.Class != "", a.Class(props.Class)),
		g.If(props.Style != "", a.Style(props.Style)),
		// Note: Real click handling would be implemented in LinkWithHandler
		g.Group(props.Children),
	)
}

// createLinkClickHandler creates a click handler for Link components
func createLinkClickHandler(props LinkProps) string {
	// This would need to be implemented as a global JS function
	// For now, return a placeholder
	return "event.preventDefault(); uiwgo_navigate('" + props.To + "', " + 
		boolToString(props.Replace) + ");"
}

// Navigate component for programmatic navigation
func Navigate(props NavigateProps) Node {
	// This component triggers navigation when mounted
	comps.OnMount(func() {
		// Check condition if provided
		if props.When != nil && !props.When() {
			return
		}
		
		router := GetRouter()
		router.Navigate(props.To, &NavigateOptions{
			Replace: props.Replace,
			State:   props.State,
		})
	})
	
	// Return empty node since this is just for side effects
	return g.Text("")
}

// Hook functions for accessing router state

// UseLocation returns the current location signal
func UseLocation() reactivity.Signal[*Location] {
	router := GetRouter()
	return router.GetLocation()
}

// UseParams returns a signal with the current route parameters
func UseParams() reactivity.Signal[map[string]string] {
	router := GetRouter()
	
	// Create a computed signal that extracts params from current route match
	return reactivity.CreateMemo(func() map[string]string {
		loc := router.GetLocation().Get()
		_, match := router.MatchRoute(loc.Pathname)
		
		if match != nil {
			return match.Params
		}
		
		return make(map[string]string)
	})
}

// UseNavigate returns a navigation function
func UseNavigate() func(to string, options ...*NavigateOptions) {
	router := GetRouter()
	return router.Navigate
}

// UseQuery returns a signal with the current query parameters
func UseQuery() reactivity.Signal[map[string]string] {
	router := GetRouter()
	
	return reactivity.CreateMemo(func() map[string]string {
		loc := router.GetLocation().Get()
		return loc.Query
	})
}

// UseRouteMatch returns the current route match information
func UseRouteMatch() reactivity.Signal[*RouteMatch] {
	router := GetRouter()
	
	return reactivity.CreateMemo(func() *RouteMatch {
		loc := router.GetLocation().Get()
		_, match := router.MatchRoute(loc.Pathname)
		return match
	})
}

// Helper functions

func boolToString(b bool) string {
	if b {
		return "true"
	}
	return "false"
}

// SetupGlobalNavigationFunction sets up the global JavaScript navigation function
func SetupGlobalNavigationFunction() {
	// This should be called during app initialization to set up the global
	// uiwgo_navigate function that can be called from onclick handlers
	
	// Implementation would use dom.FunctionManager to create a global JS function
	// that calls router.Navigate()
}