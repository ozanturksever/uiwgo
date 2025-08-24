//go:build js && wasm

package router

import (
	"bytes"
	"syscall/js"
	"testing"
	"time"

	"github.com/ozanturksever/uiwgo/comps"
	g "maragu.dev/gomponents"
)

func TestRouterProvider(t *testing.T) {
	// Test RouterProvider component creation
	props := RouterProps{
		BasePath: "/app",
		BeforeGuard: func(to *Location) bool {
			return true
		},
		AfterGuard: func(from, to *Location) {
			// Test callback
		},
		Children: []Node{},
	}
	
	component := RouterProvider(props)
	if component == nil {
		t.Fatal("RouterProvider returned nil component")
	}
	
	// Verify router was created with correct base path
	router := GetRouter()
	if router == nil {
		t.Fatal("Router was not created")
	}
	
	if router.basePath != "/app" {
		t.Errorf("Expected basePath '/app', got '%s'", router.basePath)
	}
	
	// Clean up
	router.Dispose()
}

func TestRouterOutlet(t *testing.T) {
	// Setup router with test route
	router := CreateRouter("")
	defer router.Dispose()
	
	router.AddRoute(&RouteConfig{
		Path: "/test",
		Component: func(match *RouteMatch) Node {
			return g.Text("Test Component")
		},
	})
	
	// Navigate to test route
	router.Navigate("/test")
	
	// Create router outlet
	outlet := RouterOutlet()
	if outlet == nil {
		t.Fatal("RouterOutlet returned nil")
	}
	
	// Test rendering (this would normally trigger the component)
	var buf bytes.Buffer
	err := outlet.Render(&buf)
	if err != nil {
		t.Errorf("RouterOutlet render failed: %v", err)
	}
	
	output := buf.String()
	if output == "" {
		t.Error("RouterOutlet rendered empty content")
	}
}

func TestRoute(t *testing.T) {
	// Setup router
	router := CreateRouter("")
	defer router.Dispose()
	
	initialRouteCount := len(router.routes)
	
	// Test Route function
	props := RouteProps{
		Path: "/test-route",
		Component: func(match *RouteMatch) Node {
			return g.Text("Test Route Component")
		},
		Guard: func(match *RouteMatch) bool {
			return true
		},
		Redirect: "",
	}
	
	Route(props)
	
	// Verify route was added
	if len(router.routes) != initialRouteCount+1 {
		t.Errorf("Expected %d routes, got %d", initialRouteCount+1, len(router.routes))
	}
	
	// Verify route configuration
	addedRoute := router.routes[len(router.routes)-1]
	if addedRoute.Path != "/test-route" {
		t.Errorf("Expected route path '/test-route', got '%s'", addedRoute.Path)
	}
}

func TestNavigateComponent(t *testing.T) {
	// Setup router
	router := CreateRouter("")
	defer router.Dispose()
	
	initialPath := router.GetLocation().Get().Pathname
	
	// Test Navigate component
	props := NavigateProps{
		To:      "/navigate-test",
		Replace: false,
		State:   "test-state",
		When:    func() bool { return true },
	}
	
	component := Navigate(props)
	if component == nil {
		t.Fatal("Navigate component returned nil")
	}
	
	// Test conditional navigation (When returns false)
	propsConditional := NavigateProps{
		To:      "/should-not-navigate",
		Replace: false,
		When:    func() bool { return false },
	}
	
	Navigate(propsConditional)
	
	// Path should not have changed since When() returned false
	currentPath := router.GetLocation().Get().Pathname
	if currentPath != initialPath {
		t.Error("Navigation should not have occurred when When() returns false")
	}
}

func TestUseLocation(t *testing.T) {
	// Setup router
	router := CreateRouter("")
	defer router.Dispose()
	
	// Test UseLocation hook
	locationSignal := UseLocation()
	if locationSignal == nil {
		t.Fatal("UseLocation returned nil signal")
	}
	
	// Test that it returns the same signal as router
	routerLocation := router.GetLocation()
	if locationSignal != routerLocation {
		t.Error("UseLocation should return the same signal as router.GetLocation()")
	}
	
	// Test reactivity
	initialLocation := locationSignal.Get()
	if initialLocation == nil {
		t.Fatal("Location signal returned nil location")
	}
}

func TestUseParams(t *testing.T) {
	// Setup router with parameterized route
	router := CreateRouter("")
	defer router.Dispose()
	
	router.AddRoute(&RouteConfig{
		Path: "/users/:id/posts/:postId",
		Component: func(match *RouteMatch) Node {
			return g.Text("User Post")
		},
	})
	
	// Navigate to route with parameters
	router.Navigate("/users/123/posts/456")
	
	// Test UseParams hook
	paramsSignal := UseParams()
	if paramsSignal == nil {
		t.Fatal("UseParams returned nil signal")
	}
	
	params := paramsSignal.Get()
	if params == nil {
		t.Fatal("Params signal returned nil")
	}
	
	// Check parameter values
	if params["id"] != "123" {
		t.Errorf("Expected id param '123', got '%s'", params["id"])
	}
	
	if params["postId"] != "456" {
		t.Errorf("Expected postId param '456', got '%s'", params["postId"])
	}
}

func TestUseQuery(t *testing.T) {
	// Setup router
	router := CreateRouter("")
	defer router.Dispose()
	
	// Navigate with query parameters
	router.Navigate("/search?q=test&page=2")
	
	// Test UseQuery hook
	querySignal := UseQuery()
	if querySignal == nil {
		t.Fatal("UseQuery returned nil signal")
	}
	
	query := querySignal.Get()
	if query == nil {
		t.Fatal("Query signal returned nil")
	}
	
	// Check query values
	if query["q"] != "test" {
		t.Errorf("Expected query param 'q' to be 'test', got '%s'", query["q"])
	}
	
	if query["page"] != "2" {
		t.Errorf("Expected query param 'page' to be '2', got '%s'", query["page"])
	}
}

func TestUseNavigate(t *testing.T) {
	// Setup router
	router := CreateRouter("")
	defer router.Dispose()
	
	// Test UseNavigate hook
	navigate := UseNavigate()
	if navigate == nil {
		t.Fatal("UseNavigate returned nil function")
	}
	
	// Test navigation using the hook
	initialPath := router.GetLocation().Get().Pathname
	navigate("/hook-test")
	
	newPath := router.GetLocation().Get().Pathname
	if newPath == initialPath {
		t.Error("Navigation via UseNavigate hook did not work")
	}
	
	if newPath != "/hook-test" {
		t.Errorf("Expected path '/hook-test', got '%s'", newPath)
	}
}

func TestUseRouteMatch(t *testing.T) {
	// Setup router with test route
	router := CreateRouter("")
	defer router.Dispose()
	
	router.AddRoute(&RouteConfig{
		Path: "/match-test/:id",
		Component: func(match *RouteMatch) Node {
			return g.Text("Match Test")
		},
	})
	
	// Navigate to route
	router.Navigate("/match-test/789")
	
	// Test UseRouteMatch hook
	matchSignal := UseRouteMatch()
	if matchSignal == nil {
		t.Fatal("UseRouteMatch returned nil signal")
	}
	
	match := matchSignal.Get()
	if match == nil {
		t.Fatal("RouteMatch signal returned nil")
	}
	
	// Check match details
	if match.Path != "/match-test/:id" {
		t.Errorf("Expected match path '/match-test/:id', got '%s'", match.Path)
	}
	
	if match.Params["id"] != "789" {
		t.Errorf("Expected id param '789', got '%s'", match.Params["id"])
	}
}

func TestBoolToString(t *testing.T) {
	// Test helper function
	if boolToString(true) != "true" {
		t.Errorf("Expected 'true', got '%s'", boolToString(true))
	}
	
	if boolToString(false) != "false" {
		t.Errorf("Expected 'false', got '%s'", boolToString(false))
	}
}

func TestComponentReactivity(t *testing.T) {
	// Setup router
	router := CreateRouter("")
	defer router.Dispose()
	
	// Add route with parameter
	router.AddRoute(&RouteConfig{
		Path: "/reactive/:value",
		Component: func(match *RouteMatch) Node {
			return g.Text("Value: " + match.Params["value"])
		},
	})
	
	// Create a test container element
	document := js.Global().Get("document")
	container := document.Call("createElement", "div")
	container.Set("id", "test-container")
	document.Get("body").Call("appendChild", container)
	defer document.Get("body").Call("removeChild", container)
	
	// Mount the RouterOutlet to the DOM
	disposer := comps.Mount("test-container", func() comps.Node {
		return RouterOutlet()
	})
	defer disposer()
	
	// Navigate to first route
	router.Navigate("/reactive/first")
	
	// Allow time for reactive updates
	time.Sleep(10 * time.Millisecond)
	
	// Get rendered content
	content1 := container.Get("innerHTML").String()
	
	// Navigate to second route
	router.Navigate("/reactive/second")
	
	// Allow time for reactive updates
	time.Sleep(10 * time.Millisecond)
	
	// Get rendered content
	content2 := container.Get("innerHTML").String()
	
	// Debug: Print the rendered content
	t.Logf("First render: %s", content1)
	t.Logf("Second render: %s", content2)
	
	// Components should be different for different parameters
	if content1 == content2 {
		t.Error("Router outlet should render different content for different routes")
	}
}

func TestRouterHooksReactivity(t *testing.T) {
	// Setup router
	router := CreateRouter("")
	defer router.Dispose()
	
	// Setup route with parameters
	router.AddRoute(&RouteConfig{
		Path: "/hooks/:id",
		Component: func(match *RouteMatch) Node {
			return g.Text("Hooks Test")
		},
	})
	
	// Get hooks
	locationSignal := UseLocation()
	paramsSignal := UseParams()
	querySignal := UseQuery()
	
	// Navigate to initial route
	router.Navigate("/hooks/initial?test=1")
	
	initialLocation := locationSignal.Get()
	initialParams := paramsSignal.Get()
	initialQuery := querySignal.Get()
	
	// Navigate to new route
	router.Navigate("/hooks/updated?test=2")
	
	newLocation := locationSignal.Get()
	newParams := paramsSignal.Get()
	newQuery := querySignal.Get()
	
	// Verify signals updated
	if initialLocation.Pathname == newLocation.Pathname {
		t.Error("Location signal should update when route changes")
	}
	
	if initialParams["id"] == newParams["id"] {
		t.Error("Params signal should update when parameters change")
	}
	
	if initialQuery["test"] == newQuery["test"] {
		t.Error("Query signal should update when query parameters change")
	}
	
	// Verify new values
	if newParams["id"] != "updated" {
		t.Errorf("Expected param 'id' to be 'updated', got '%s'", newParams["id"])
	}
	
	if newQuery["test"] != "2" {
		t.Errorf("Expected query 'test' to be '2', got '%s'", newQuery["test"])
	}
}