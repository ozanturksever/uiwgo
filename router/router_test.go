//go:build js && wasm

package router

import (
	"testing"
)

func TestCreateRouter(t *testing.T) {
	// Test router creation with base path
	router := CreateRouter("/app")
	
	if router == nil {
		t.Fatal("CreateRouter returned nil")
	}
	
	if router.basePath != "/app" {
		t.Errorf("Expected basePath '/app', got '%s'", router.basePath)
	}
	
	if router.location == nil {
		t.Fatal("Router location signal is nil")
	}
	
	if router.routes == nil {
		t.Fatal("Router routes slice is nil")
	}
	
	// Clean up
	router.Dispose()
}

func TestRouteCompilation(t *testing.T) {
	router := CreateRouter("")
	defer router.Dispose()
	
	testCases := []struct {
		pattern    string
		testPath   string
		shouldMatch bool
		expectedParams map[string]string
	}{
		{"/users", "/users", true, map[string]string{}},
		{"/users", "/users/", false, nil},
		{"/users/:id", "/users/123", true, map[string]string{"id": "123"}},
		{"/users/:id", "/users/", false, nil},
		{"/users/:id/posts/:postId", "/users/123/posts/456", true, map[string]string{"id": "123", "postId": "456"}},
		{"/api/*", "/api/v1/users", true, map[string]string{}},
		{"/api/*", "/api", false, nil},
		{"/*", "/anything/goes/here", true, map[string]string{}},
	}
	
	for _, tc := range testCases {
		t.Run(tc.pattern+"->"+tc.testPath, func(t *testing.T) {
			config := &RouteConfig{
				Path: tc.pattern,
				Component: func(match *RouteMatch) Node {
					return nil
				},
			}
			
			router.compileRoute(config)
			
			if config.regexp == nil {
				t.Fatal("Route compilation failed: regexp is nil")
			}
			
			match := router.matchSingleRoute(config, tc.testPath)
			
			if tc.shouldMatch {
				if match == nil {
					t.Errorf("Expected pattern '%s' to match path '%s', but it didn't", tc.pattern, tc.testPath)
					return
				}
				
				// Check parameters
				for expectedKey, expectedValue := range tc.expectedParams {
					if actualValue, exists := match.Params[expectedKey]; !exists {
						t.Errorf("Expected parameter '%s' not found in match", expectedKey)
					} else if actualValue != expectedValue {
						t.Errorf("Expected parameter '%s' to be '%s', got '%s'", expectedKey, expectedValue, actualValue)
					}
				}
				
				// Check no extra parameters
				if len(match.Params) != len(tc.expectedParams) {
					t.Errorf("Expected %d parameters, got %d: %v", len(tc.expectedParams), len(match.Params), match.Params)
				}
			} else {
				if match != nil {
					t.Errorf("Expected pattern '%s' to NOT match path '%s', but it did", tc.pattern, tc.testPath)
				}
			}
		})
	}
}

func TestRouteMatching(t *testing.T) {
	router := CreateRouter("")
	defer router.Dispose()
	
	// Add test routes
	routes := []*RouteConfig{
		{Path: "/", Component: func(match *RouteMatch) Node { return nil }},
		{Path: "/about", Component: func(match *RouteMatch) Node { return nil }},
		{Path: "/users/:id", Component: func(match *RouteMatch) Node { return nil }},
		{Path: "/api/*", Component: func(match *RouteMatch) Node { return nil }},
	}
	
	for _, route := range routes {
		router.AddRoute(route)
	}
	
	testCases := []struct {
		path           string
		expectedRoute  string
		expectedParams map[string]string
	}{
		{"/", "/", map[string]string{}},
		{"/about", "/about", map[string]string{}},
		{"/users/123", "/users/:id", map[string]string{"id": "123"}},
		{"/api/v1/test", "/api/*", map[string]string{}},
	}
	
	for _, tc := range testCases {
		t.Run(tc.path, func(t *testing.T) {
			route, match := router.MatchRoute(tc.path)
			
			if route == nil {
				t.Fatalf("No route matched for path '%s'", tc.path)
			}
			
			if route.Path != tc.expectedRoute {
				t.Errorf("Expected route '%s', got '%s'", tc.expectedRoute, route.Path)
			}
			
			if match == nil {
				t.Fatal("Match is nil")
			}
			
			// Check parameters
			for expectedKey, expectedValue := range tc.expectedParams {
				if actualValue, exists := match.Params[expectedKey]; !exists {
					t.Errorf("Expected parameter '%s' not found", expectedKey)
				} else if actualValue != expectedValue {
					t.Errorf("Expected parameter '%s' to be '%s', got '%s'", expectedKey, expectedValue, actualValue)
				}
			}
		})
	}
}

func TestRouteGuards(t *testing.T) {
	router := CreateRouter("")
	defer router.Dispose()
	
	// Test before guard
	guardCalled := false
	router.SetBeforeGuard(func(to *Location) bool {
		guardCalled = true
		return to.Pathname != "/blocked"
	})
	
	// Test navigation to allowed path
	router.Navigate("/allowed")
	if !guardCalled {
		t.Error("Before guard was not called")
	}
	
	// Reset guard call flag
	guardCalled = false
	
	// Test navigation to blocked path
	currentPath := router.GetLocation().Get().Pathname
	router.Navigate("/blocked")
	
	// Should still be on the same path since navigation was blocked
	newPath := router.GetLocation().Get().Pathname
	if newPath != currentPath {
		t.Error("Navigation should have been blocked by guard")
	}
	
	if !guardCalled {
		t.Error("Before guard was not called for blocked navigation")
	}
}

func TestRouteWithRedirect(t *testing.T) {
	router := CreateRouter("")
	defer router.Dispose()
	
	// Add a redirect route
	router.AddRoute(&RouteConfig{
		Path:     "/old-path",
		Redirect: "/new-path",
	})
	
	// Add the target route
	router.AddRoute(&RouteConfig{
		Path: "/new-path",
		Component: func(match *RouteMatch) Node {
			return nil
		},
	})
	
	// Test that matching the redirect route returns the redirect config
	route, match := router.MatchRoute("/old-path")
	
	if route == nil {
		t.Fatal("No route matched for redirect path")
	}
	
	if route.Redirect != "/new-path" {
		t.Errorf("Expected redirect to '/new-path', got '%s'", route.Redirect)
	}
	
	if match == nil {
		t.Fatal("Match is nil for redirect route")
	}
}

func TestLocationSignalReactivity(t *testing.T) {
	router := CreateRouter("")
	defer router.Dispose()
	
	// Test that location signal is reactive
	locationSignal := router.GetLocation()
	
	if locationSignal == nil {
		t.Fatal("Location signal is nil")
	}
	
	// Get initial location
	initialLocation := locationSignal.Get()
	if initialLocation == nil {
		t.Fatal("Initial location is nil")
	}
	
	// Test that we can access location properties
	if initialLocation.Pathname == "" {
		t.Error("Initial pathname is empty")
	}
	
	if initialLocation.Query == nil {
		t.Error("Initial query map is nil")
	}
}

func TestParameterExtraction(t *testing.T) {
	router := CreateRouter("")
	defer router.Dispose()
	
	// Test complex parameter extraction
	config := &RouteConfig{
		Path: "/users/:userId/posts/:postId/comments/:commentId",
		Component: func(match *RouteMatch) Node {
			return nil
		},
	}
	
	router.compileRoute(config)
	router.AddRoute(config)
	
	testPath := "/users/123/posts/456/comments/789"
	route, match := router.MatchRoute(testPath)
	
	if route == nil {
		t.Fatal("Route not matched")
	}
	
	if match == nil {
		t.Fatal("Match is nil")
	}
	
	expectedParams := map[string]string{
		"userId":    "123",
		"postId":    "456",
		"commentId": "789",
	}
	
	for key, expected := range expectedParams {
		if actual, exists := match.Params[key]; !exists {
			t.Errorf("Parameter '%s' not found", key)
		} else if actual != expected {
			t.Errorf("Parameter '%s': expected '%s', got '%s'", key, expected, actual)
		}
	}
}

func TestWildcardMatching(t *testing.T) {
	// Test /files/* route in isolation first
	t.Run("files-wildcard", func(t *testing.T) {
		router := CreateRouter("")
		defer router.Dispose()
		
		config := &RouteConfig{
			Path: "/files/*",
			Component: func(match *RouteMatch) Node {
				return nil
			},
		}
		
		router.compileRoute(config)
		t.Logf("Compiled /files/* to regex: %s", config.regexp.String())
		router.AddRoute(config)
		
		testCases := []struct {
			path             string
			shouldMatch      bool
			expectedWildcard string
		}{
			{"/files/document.pdf", true, "document.pdf"},
			{"/files/images/photo.jpg", true, "images/photo.jpg"},
			{"/files/nested/deep/path.txt", true, "nested/deep/path.txt"},
			{"/files", false, ""},
			{"/other", false, ""},
		}
		
		for _, tc := range testCases {
			t.Run(tc.path, func(t *testing.T) {
				route, match := router.MatchRoute(tc.path)
				
				if tc.shouldMatch {
					if route == nil {
						t.Errorf("Expected path '%s' to match wildcard route", tc.path)
						return
					}
					
					if match == nil {
						t.Fatal("Match is nil")
					}
					
					if match.Wildcard != tc.expectedWildcard {
						t.Errorf("Expected wildcard '%s', got '%s'", tc.expectedWildcard, match.Wildcard)
					}
				} else {
					if route != nil {
						t.Errorf("Expected path '%s' to NOT match wildcard route, but matched route '%s'", tc.path, route.Path)
					}
				}
			})
		}
	})
	
	// Test root wildcard pattern with multiple routes
	t.Run("root-wildcard-with-other-routes", func(t *testing.T) {
		router := CreateRouter("")
		defer router.Dispose()
		
		// Add specific routes first
		router.AddRoute(&RouteConfig{
			Path: "/",
			Component: func(match *RouteMatch) Node { return nil },
		})
		
		router.AddRoute(&RouteConfig{
			Path: "/about",
			Component: func(match *RouteMatch) Node { return nil },
		})
		
		router.AddRoute(&RouteConfig{
			Path: "/files/*",
			Component: func(match *RouteMatch) Node { return nil },
		})
		
		// Add catch-all route last
		rootWildcardConfig := &RouteConfig{
			Path: "/*",
			Component: func(match *RouteMatch) Node {
				return nil
			},
		}
		
		router.compileRoute(rootWildcardConfig)
		t.Logf("Compiled /* to regex: %s", rootWildcardConfig.regexp.String())
		router.AddRoute(rootWildcardConfig)
	
		rootWildcardCases := []struct {
			path             string
			expectedRoute    string
			expectedWildcard string
		}{
			{"/", "/", ""}, // Should match exact root route
			{"/about", "/about", ""}, // Should match exact about route
			{"/files/doc.pdf", "/files/*", "doc.pdf"}, // Should match files wildcard
			{"/nonexistent-page", "/*", "nonexistent-page"}, // Should match catch-all
			{"/some/deep/path", "/*", "some/deep/path"}, // Should match catch-all
			{"/files", "/*", "files"}, // Should match catch-all (not /files/*)
			{"/other", "/*", "other"}, // Should match catch-all
		}
		
		for _, tc := range rootWildcardCases {
			t.Run(tc.path, func(t *testing.T) {
				route, match := router.MatchRoute(tc.path)
				
				if route == nil {
					t.Errorf("Expected path '%s' to match a route, got no match", tc.path)
					return
				}
				
				if route.Path != tc.expectedRoute {
					t.Errorf("Expected path '%s' to match route '%s', got '%s'", tc.path, tc.expectedRoute, route.Path)
				}
				
				if match == nil {
					t.Fatal("Match is nil")
				}
				
				if match.Wildcard != tc.expectedWildcard {
					t.Errorf("Expected wildcard '%s', got '%s'", tc.expectedWildcard, match.Wildcard)
				}
				
				t.Logf("Path '%s' correctly matched route '%s' with wildcard '%s'", tc.path, route.Path, match.Wildcard)
			})
		}
	})
}

func TestBasePath(t *testing.T) {
	router := CreateRouter("/app")
	defer router.Dispose()
	
	// Test that base path is correctly handled
	if router.basePath != "/app" {
		t.Errorf("Expected basePath '/app', got '%s'", router.basePath)
	}
	
	// Test path resolution
	resolved := router.resolvePath("/users")
	expected := "/app/users"
	if resolved != expected {
		t.Errorf("Expected resolved path '%s', got '%s'", expected, resolved)
	}
	
	// Test relative path resolution
	// Set current location for relative resolution test
	router.location.Set(&Location{Pathname: "/app/users"})
	relativeResolved := router.resolvePath("123")
	expectedRelative := "/app/users/123"
	if relativeResolved != expectedRelative {
		t.Errorf("Expected relative resolved path '%s', got '%s'", expectedRelative, relativeResolved)
	}
}

func TestRouterDisposal(t *testing.T) {
	router := CreateRouter("")
	
	// Test that router can be disposed without errors
	router.Dispose()
	
	if !router.disposed {
		t.Error("Router should be marked as disposed")
	}
	
	// Test that double disposal doesn't cause issues
	router.Dispose() // Should not panic
}

func TestRouteConfigValidation(t *testing.T) {
	router := CreateRouter("")
	defer router.Dispose()
	
	// Test empty route path validation
	config := &RouteConfig{
		Path: "",
		Component: func(match *RouteMatch) Node {
			return nil
		},
	}
	
	initialCount := len(router.routes)
	router.AddRoute(config)
	
	// Should not add empty path routes
	if len(router.routes) != initialCount {
		t.Error("Empty path route should not be added")
	}
}

func TestNavigationOptions(t *testing.T) {
	router := CreateRouter("")
	defer router.Dispose()
	
	// Test navigation with options
	opts := &NavigateOptions{
		Replace: true,
		State:   "test-state",
	}
	
	router.Navigate("/test", opts)
	
	// Verify location was updated
	location := router.GetLocation().Get()
	if location.Pathname != "/test" {
		t.Errorf("Expected pathname '/test', got '%s'", location.Pathname)
	}
	
	if location.State != "test-state" {
		t.Errorf("Expected state 'test-state', got '%v'", location.State)
	}
}

func TestQueryParameterParsing(t *testing.T) {
	router := CreateRouter("")
	defer router.Dispose()
	
	// Test navigation with query parameters
	router.Navigate("/search?q=test&page=1")
	
	location := router.GetLocation().Get()
	
	if location.Query["q"] != "test" {
		t.Errorf("Expected query param 'q' to be 'test', got '%s'", location.Query["q"])
	}
	
	if location.Query["page"] != "1" {
		t.Errorf("Expected query param 'page' to be '1', got '%s'", location.Query["page"])
	}
}

// Benchmark tests
func BenchmarkRouteMatching(b *testing.B) {
	router := CreateRouter("")
	defer router.Dispose()
	
	// Add many routes
	for i := 0; i < 100; i++ {
		router.AddRoute(&RouteConfig{
			Path: "/route" + string(rune(48+i%10)),
			Component: func(match *RouteMatch) Node {
				return nil
			},
		})
	}
	
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		router.MatchRoute("/route5")
	}
}

func BenchmarkParameterExtraction(b *testing.B) {
	router := CreateRouter("")
	defer router.Dispose()
	
	router.AddRoute(&RouteConfig{
		Path: "/users/:id/posts/:postId",
		Component: func(match *RouteMatch) Node {
			return nil
		},
	})
	
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		router.MatchRoute("/users/123/posts/456")
	}
}