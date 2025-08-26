//go:build !js && !wasm

package router

import (
	"testing"
)

// TestNavigationSystem_Navigate tests the Navigate method functionality.
func TestNavigationSystem_Navigate(t *testing.T) {
	// Create a simple component function
	component := func(props ...any) interface{} {
		return "test component"
	}

	// Create route definitions
	routes := []*RouteDefinition{
		Route("/", component),
		Route("/about", component),
		Route("/users/:id", component),
	}

	// Create router
	router := New(routes, nil)

	// Test initial location
	initialLocation := router.Location()
	if initialLocation.Pathname != "/" {
		t.Errorf("Expected initial pathname to be '/', got %s", initialLocation.Pathname)
	}

	// Test navigation to /about
	router.Navigate("/about", NavigateOptions{})
	location := router.Location()
	if location.Pathname != "/about" {
		t.Errorf("Expected pathname to be '/about', got %s", location.Pathname)
	}

	// Test navigation with state
	testState := "test-state-value"
	router.Navigate("/users/123", NavigateOptions{State: testState})
	location = router.Location()
	if location.Pathname != "/users/123" {
		t.Errorf("Expected pathname to be '/users/123', got %s", location.Pathname)
	}
	if location.State != testState {
		t.Errorf("Expected state to be %v, got %v", testState, location.State)
	}
}

// TestNavigationSystem_RouteMatching tests that navigation properly matches routes.
func TestNavigationSystem_RouteMatching(t *testing.T) {
	// Create a component function
	component := func(props ...any) interface{} {
		return "test component"
	}

	// Create route definitions with parameters
	routes := []*RouteDefinition{
		Route("/", component),
		Route("/users/:id", component),
		Route("/posts/:id/comments/:commentId", component),
	}

	// Create router
	router := New(routes, nil)

	// Test navigation to route with single parameter
	router.Navigate("/users/456", NavigateOptions{})
	matchedRoute, params := router.Match(router.Location().Pathname)
	if matchedRoute == nil {
		t.Error("Expected route to match /users/:id")
	}
	if params["id"] != "456" {
		t.Errorf("Expected id parameter to be '456', got %s", params["id"])
	}

	// Test navigation to route with multiple parameters
	router.Navigate("/posts/123/comments/789", NavigateOptions{})
	matchedRoute, params = router.Match(router.Location().Pathname)
	if matchedRoute == nil {
		t.Error("Expected route to match /posts/:id/comments/:commentId")
	}
	if params["id"] != "123" {
		t.Errorf("Expected id parameter to be '123', got %s", params["id"])
	}
	if params["commentId"] != "789" {
		t.Errorf("Expected commentId parameter to be '789', got %s", params["commentId"])
	}
}

// TestNavigationSystem_LocationStateReactivity tests that location state changes trigger subscribers.
func TestNavigationSystem_LocationStateReactivity(t *testing.T) {
	// Create a simple component function
	component := func(props ...any) interface{} {
		return "test component"
	}

	// Create route definitions
	routes := []*RouteDefinition{
		Route("/", component),
		Route("/page1", component),
		Route("/page2", component),
	}

	// Create router
	router := New(routes, nil)

	// Track location changes
	var locationChanges []Location
	router.locationState.Subscribe(func(newLocation Location) {
		locationChanges = append(locationChanges, newLocation)
	})

	// Navigate to different pages
	router.Navigate("/page1", NavigateOptions{})
	router.Navigate("/page2", NavigateOptions{})

	// Verify that subscribers were notified
	if len(locationChanges) != 2 {
		t.Errorf("Expected 2 location changes, got %d", len(locationChanges))
	}

	if locationChanges[0].Pathname != "/page1" {
		t.Errorf("Expected first change to be '/page1', got %s", locationChanges[0].Pathname)
	}

	if locationChanges[1].Pathname != "/page2" {
		t.Errorf("Expected second change to be '/page2', got %s", locationChanges[1].Pathname)
	}
}