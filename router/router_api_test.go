package router

import (
	"testing"

	g "maragu.dev/gomponents"
	h "maragu.dev/gomponents/html"
)

// TestRouterAPI_RouteBuilderWithDynamicPaths tests that the Route builder function
// correctly creates RouteDefinitions that can match dynamic paths.
func TestRouterAPI_RouteBuilderWithDynamicPaths(t *testing.T) {
	// Define a simple component function
	userComponent := func(props ...any) interface{} {
		return "User Component"
	}

	// Create routes using the Route builder function
	routes := []*RouteDefinition{
		Route("/", func(props ...any) interface{} { return "Home" }),
		Route("/users/:id", userComponent),
		Route("/posts/:id/comments/:commentId", func(props ...any) interface{} { return "Comment" }),
		Route("/files/*filepath", func(props ...any) interface{} { return "File" }),
	}

	// Create router with mock outlet
	router := New(routes, nil)

	// Test static route matching
	route, params := router.Match("/")
	if route == nil {
		t.Fatal("Expected route to match '/', got nil")
	}
	if route.Path != "/" {
		t.Errorf("Expected route path '/', got %q", route.Path)
	}
	if len(params) != 0 {
		t.Errorf("Expected no parameters for static route, got %v", params)
	}

	// Test dynamic route matching with single parameter
	route, params = router.Match("/users/123")
	if route == nil {
		t.Fatal("Expected route to match '/users/123', got nil")
	}
	if route.Path != "/users/:id" {
		t.Errorf("Expected route path '/users/:id', got %q", route.Path)
	}
	if params["id"] != "123" {
		t.Errorf("Expected parameter 'id' to be '123', got %q", params["id"])
	}

	// Test router.Params() method
	routerParams := router.Params()
	if routerParams["id"] != "123" {
		t.Errorf("Expected router.Params()['id'] to be '123', got %q", routerParams["id"])
	}

	// Test dynamic route with multiple parameters
	route, params = router.Match("/posts/456/comments/789")
	if route == nil {
		t.Fatal("Expected route to match '/posts/456/comments/789', got nil")
	}
	if params["id"] != "456" {
		t.Errorf("Expected parameter 'id' to be '456', got %q", params["id"])
	}
	if params["commentId"] != "789" {
		t.Errorf("Expected parameter 'commentId' to be '789', got %q", params["commentId"])
	}

	// Test wildcard route
	route, params = router.Match("/files/docs/readme.txt")
	if route == nil {
		t.Fatal("Expected route to match '/files/docs/readme.txt', got nil")
	}
	if params["filepath"] != "docs/readme.txt" {
		t.Errorf("Expected parameter 'filepath' to be 'docs/readme.txt', got %q", params["filepath"])
	}
}

// TestRouterAPI_LocationAndNavigation tests the Location() method and Navigate() functionality.
func TestRouterAPI_LocationAndNavigation(t *testing.T) {
	routes := []*RouteDefinition{
		Route("/", func(props ...any) interface{} { return h.Div(g.Text("Home")) }),
		Route("/about", func(props ...any) interface{} { return h.Div(g.Text("About")) }),
	}

	router := New(routes, nil)

	// Test initial location
	location := router.Location()
	if location.Pathname != "/" {
		t.Errorf("Expected initial pathname to be '/', got %q", location.Pathname)
	}

	// Test navigation
	router.Navigate("/about")
	location = router.Location()
	if location.Pathname != "/about" {
		t.Errorf("Expected pathname after navigation to be '/about', got %q", location.Pathname)
	}

	// Test navigation with state
	router.Navigate("/", NavigateOptions{State: "test-state"})
	location = router.Location()
	if location.Pathname != "/" {
		t.Errorf("Expected pathname after navigation to be '/', got %q", location.Pathname)
	}
	if location.State != "test-state" {
		t.Errorf("Expected state to be 'test-state', got %v", location.State)
	}
}

// TestRouterAPI_ParamsAfterNoMatch tests that Params() returns empty map when no route matches.
func TestRouterAPI_ParamsAfterNoMatch(t *testing.T) {
	routes := []*RouteDefinition{
		Route("/users/:id", func(props ...any) interface{} { return "User" }),
	}

	router := New(routes, nil)

	// First match a route to set parameters
	router.Match("/users/123")
	params := router.Params()
	if params["id"] != "123" {
		t.Errorf("Expected parameter 'id' to be '123', got %q", params["id"])
	}

	// Now try to match a non-existent route
	route, _ := router.Match("/nonexistent")
	if route != nil {
		t.Errorf("Expected no route to match '/nonexistent', got %v", route)
	}

	// Params should now be empty
	params = router.Params()
	if len(params) != 0 {
		t.Errorf("Expected empty params after no match, got %v", params)
	}
}