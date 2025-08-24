//go:build js && wasm

package router

import (
	"fmt"
	"strconv"
	"syscall/js"
	"time"

	"github.com/ozanturksever/uiwgo/comps"
	"github.com/ozanturksever/uiwgo/logutil"
	g "maragu.dev/gomponents"
	a "maragu.dev/gomponents/html"
)

// NavigationManager handles global navigation functions and Link event handling
type NavigationManager struct {
	initialized bool
	router      *Router
	linkCounter int
}

var navigationManager = &NavigationManager{}

// InitializeNavigation sets up global navigation functions
func InitializeNavigation() {
	if navigationManager.initialized {
		logutil.Log("Navigation already initialized")
		return
	}
	
	router := GetRouter()
	navigationManager.router = router
	navigationManager.initialized = true
	
	// Create global navigation function for Link components using syscall/js directly
	navFunc := js.FuncOf(func(this js.Value, args []js.Value) any {
		logutil.Log("uiwgo_navigate called with args:", len(args))
		if len(args) < 1 {
			logutil.Log("uiwgo_navigate: insufficient arguments")
			return nil
		}
		
		to := args[0].String()
		replace := false
		if len(args) > 1 && args[1].Type() == js.TypeBoolean {
			replace = args[1].Bool()
		}
		
		logutil.Log("Navigating to:", to, "replace:", replace)
		router.Navigate(to, &NavigateOptions{Replace: replace})
		return nil
	})
	
	// Set as global function
	js.Global().Set("uiwgo_navigate", navFunc)
	
	logutil.Log("Navigation manager initialized with global function")
}

// LinkWithHandler creates a Link component with proper event handling
func LinkWithHandler(props LinkProps) Node {
	// Ensure navigation is initialized
	InitializeNavigation()
	
	// Create a simple HTML link with onclick handler
	onClickScript := fmt.Sprintf("event.preventDefault(); uiwgo_navigate('%s', %s); return false;", 
		props.To, boolToString(props.Replace))
	
	logutil.Log("Creating link with onclick:", onClickScript)
	
	return a.A(
		a.Href(props.To),
		g.If(props.Class != "", a.Class(props.Class)),
		g.If(props.Style != "", a.Style(props.Style)),
		g.Attr("onclick", onClickScript),
		g.Group(props.Children),
	)
}





// NavLink creates a Link with active state styling
type NavLinkProps struct {
	To           string    // Target path
	Replace      bool      // Whether to replace history entry  
	State        any       // State to pass with navigation
	Class        string    // Base CSS class
	ActiveClass  string    // CSS class when route is active
	Style        string    // Inline styles
	Children     []Node    // Link content
	OnClick      func()    // Additional click handler
	End          bool      // Exact match for active state
}

// NavLink creates a navigation link with active state indication
func NavLink(props NavLinkProps) Node {
	router := GetRouter()
	
	// Create a reactive class that includes active state
	activeClass := ""
	
	// Check if current route matches
	loc := router.GetLocation().Get()
	isActive := false
	
	if props.End {
		// Exact match
		isActive = loc.Pathname == props.To
	} else {
		// Prefix match
		isActive = loc.Pathname == props.To || 
		          (props.To != "/" && loc.Pathname != "/" && 
		           len(props.To) > 1 && 
		           loc.Pathname[:len(props.To)] == props.To)
	}
	
	if isActive && props.ActiveClass != "" {
		activeClass = props.ActiveClass
	}
	
	// Combine classes
	finalClass := props.Class
	if activeClass != "" {
		if finalClass != "" {
			finalClass += " " + activeClass
		} else {
			finalClass = activeClass
		}
	}
	
	// Create a simple HTML link with onclick handler
	onClickScript := fmt.Sprintf("event.preventDefault(); uiwgo_navigate('%s', %s); return false;", 
		props.To, boolToString(props.Replace))
	
	return a.A(
		a.Href(props.To),
		g.If(finalClass != "", a.Class(finalClass)),
		g.If(props.Style != "", a.Style(props.Style)),
		g.Attr("onclick", onClickScript),
		g.Group(props.Children),
	)
}

// Redirect component for immediate navigation
func Redirect(to string, replace bool) Node {
	comps.OnMount(func() {
		router := GetRouter()
		router.Navigate(to, &NavigateOptions{Replace: replace})
	})
	
	return g.Text("")
}

// ProtectedRoute creates a route that requires authentication/authorization
type ProtectedRouteProps struct {
	Path        string                           // Route pattern
	Component   func(match *RouteMatch) Node     // Component to render
	CanActivate func(match *RouteMatch) bool     // Authorization check
	Fallback    func() Node                      // Component to render if not authorized
	RedirectTo  string                           // Path to redirect if not authorized
}

// ProtectedRoute creates a route with authorization checking
func ProtectedRoute(props ProtectedRouteProps) {
	router := GetRouter()
	
	config := &RouteConfig{
		Path: props.Path,
		Component: func(match *RouteMatch) Node {
			// Check authorization
			if props.CanActivate != nil && !props.CanActivate(match) {
				// Not authorized
				if props.RedirectTo != "" {
					// Redirect to specified path
					router.Navigate(props.RedirectTo, &NavigateOptions{Replace: true})
					return g.Text("")
				} else if props.Fallback != nil {
					// Render fallback component
					return props.Fallback()
				} else {
					// Default unauthorized message
					return g.Text("Unauthorized")
				}
			}
			
			// Authorized - render the component
			return props.Component(match)
		},
	}
	
	router.AddRoute(config)
}

// WithRouter HOC (Higher Order Component) pattern
func WithRouter(component func(router *Router, match *RouteMatch) Node) func(match *RouteMatch) Node {
	return func(match *RouteMatch) Node {
		router := GetRouter()
		return component(router, match)
	}
}

// RouteGuard creates a global route guard
func RouteGuard(guard func(from, to *Location) bool) {
	router := GetRouter()
	
	router.SetBeforeGuard(func(to *Location) bool {
		from := router.GetLocation().Get()
		return guard(from, to)
	})
}

// generateLinkID generates a unique ID for link components
func generateLinkID() string {
	// Use current time in nanoseconds to ensure uniqueness
	timestamp := time.Now().UnixNano()
	return "uiwgo-link-" + strconv.FormatInt(timestamp, 10)
}

// getLinkCounter returns the current link counter
func (nm *NavigationManager) getLinkCounter() int {
	nm.linkCounter++
	return nm.linkCounter
}