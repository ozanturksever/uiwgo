package router

import "testing"

func TestRouterNewInitializesWithRoutesAndOutlet(t *testing.T) {
	// Create mock routes
	routes := []*RouteDefinition{
		{Path: "/home"},
		{Path: "/about"},
	}

	// Use nil as a mock for dom.Element outlet
	outlet := any(nil)

	// Call the New constructor
	router := New(routes, outlet)

	// Verify that the router was initialized with the correct routes
	if len(router.routes) != len(routes) {
		t.Errorf("Expected %d routes, got %d", len(routes), len(router.routes))
	}

	// Verify that each route is correctly stored
	for i, route := range routes {
		if router.routes[i] != route {
			t.Errorf("Route at index %d does not match expected route", i)
		}
	}

	// Note: The outlet is stored as any type, so we can't directly assert on it in this test
	// but we can verify that the router was created successfully
	if router == nil {
		t.Error("Router should not be nil")
	}
}

func TestRouteBuilderCreatesDefinition(t *testing.T) {
	// Define a mock component function
	mockComponent := func(props ...any) interface{} {
		return nil
	}

	// Define child routes
	childRoute := &RouteDefinition{Path: "/child"}
	children := []*RouteDefinition{childRoute}

	// Call the Route builder function
	routeDef := Route("/test", mockComponent, children...)

	// Verify the returned RouteDefinition has the correct path
	if routeDef.Path != "/test" {
		t.Errorf("Expected path '/test', got '%s'", routeDef.Path)
	}

	// Verify the component function is set correctly
	if routeDef.Component == nil {
		t.Error("Component function should not be nil")
	} else {
		// Call the component function to ensure it's the same
		result := routeDef.Component()
		if result != nil {
			t.Error("Mock component should return nil")
		}
	}

	// Verify children are set correctly
	if len(routeDef.Children) != 1 {
		t.Errorf("Expected 1 child, got %d", len(routeDef.Children))
	} else if routeDef.Children[0] != childRoute {
		t.Error("Child route does not match expected")
	}

	// Verify MatchFilters is initialized (set by Route builder)
	if routeDef.MatchFilters == nil {
		t.Error("MatchFilters should be initialized as empty map")
	}

	// Verify matcher is compiled (set by Route builder)
	if routeDef.matcher == nil {
		t.Error("matcher should be compiled by Route builder")
	}
}

func TestRouterLocationReturnsCurrentState(t *testing.T) {
	// Create a new Router instance with empty routes and nil outlet
	router := New([]*RouteDefinition{}, nil)

	// Verify that calling Location() initially returns a root path Location struct
	initialLocation := router.Location()
	expectedInitialLocation := Location{Pathname: "/", Search: "", Hash: "", State: nil}
	if initialLocation != expectedInitialLocation {
		t.Errorf("Expected initial location to be %+v, got %+v", expectedInitialLocation, initialLocation)
	}

	// Set a new location on the router's internal LocationState
	newLocation := Location{
		Pathname: "/test",
		Search:   "?query=1",
		Hash:     "#section",
		State:    "test-state",
	}
	router.locationState.Set(newLocation)

	// Verify that a subsequent call to Location() returns the updated location
	updatedLocation := router.Location()
	if updatedLocation != newLocation {
		t.Errorf("Expected updated location %+v, got %+v", newLocation, updatedLocation)
	}
}

func TestRouterParamsReturnsMatchedParams(t *testing.T) {
	// Create a dynamic route definition
	dynamicRoute := &RouteDefinition{
		Path: "/users/:id",
		// matcher will be set by the route processing, but for test we need to simulate it
		matcher: func(path string) (bool, map[string]string) {
			if path == "/users/123" {
				return true, map[string]string{"id": "123"}
			}
			return false, nil
		},
	}

	// Create router with the dynamic route
	router := New([]*RouteDefinition{dynamicRoute}, nil)

	// Verify that calling Params() initially returns an empty map
	initialParams := router.Params()
	if initialParams == nil {
		t.Error("Params() should return an empty map, not nil")
	} else if len(initialParams) != 0 {
		t.Errorf("Expected empty params map initially, got %v", initialParams)
	}

	// Manually call Match to simulate a route match and populate internal state
	matchedRoute, params := router.Match("/users/123")
	if matchedRoute == nil {
		t.Fatal("Expected route to match /users/123")
	}
	if params == nil || params["id"] != "123" {
		t.Errorf("Expected params to contain id=123, got %v", params)
	}

	// Verify that a subsequent call to Params() returns the correct captured parameters
	updatedParams := router.Params()
	if updatedParams == nil {
		t.Error("Params() should return the matched params, not nil")
	} else if len(updatedParams) != 1 || updatedParams["id"] != "123" {
		t.Errorf("Expected params map with id=123, got %v", updatedParams)
	}
}
