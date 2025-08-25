package router

import (
	"testing"
)

// TestRouteMatcher_StaticPathMatches verifies that a static path matches exactly.
func TestRouteMatcher_StaticPathMatches(t *testing.T) {
	// Create a route definition with a static path using the constructor
	route := NewRouteDefinition("/about", nil)

	// Test matching path
	isMatch, params := route.matcher("/about")
	if !isMatch {
		t.Error("Expected path /about to match route /about")
	}
	if len(params) != 0 {
		t.Errorf("Expected no parameters for static path, got %v", params)
	}

	// Test non-matching path
	isMatch, params = route.matcher("/contact")
	if isMatch {
		t.Error("Expected path /contact not to match route /about")
	}
}

// TestRouteMatcher_DynamicSegmentCapturesParam verifies that dynamic segments
// (starting with :) capture parameters correctly.
func TestRouteMatcher_DynamicSegmentCapturesParam(t *testing.T) {
	// Create a route definition with a dynamic segment
	route := NewRouteDefinition("/users/:id", nil)

	// Test matching path with parameter
	isMatch, params := route.matcher("/users/123")
	if !isMatch {
		t.Error("Expected path /users/123 to match route /users/:id")
	}
	if len(params) != 1 {
		t.Errorf("Expected 1 parameter, got %d", len(params))
	}
	if params["id"] != "123" {
		t.Errorf("Expected parameter 'id' to be '123', got %q", params["id"])
	}

	// Test non-matching path
	isMatch, params = route.matcher("/users")
	if isMatch {
		t.Error("Expected path /users not to match route /users/:id (missing segment)")
	}

	// Test another non-matching path
	isMatch, params = route.matcher("/users/123/posts")
	if isMatch {
		t.Error("Expected path /users/123/posts not to match route /users/:id (extra segment)")
	}
}

// TestRouteMatcher_OptionalSegment verifies that optional segments (ending with ?)
// match both when present and when absent.
func TestRouteMatcher_OptionalSegment(t *testing.T) {
	// Create a route definition with an optional segment
	route := NewRouteDefinition("/users/:id?/profile", nil)

	// Test matching path with parameter present
	isMatch, params := route.matcher("/users/123/profile")
	if !isMatch {
		t.Error("Expected path /users/123/profile to match route /users/:id?/profile")
	}
	if len(params) != 1 {
		t.Errorf("Expected 1 parameter, got %d", len(params))
	}
	if params["id"] != "123" {
		t.Errorf("Expected parameter 'id' to be '123', got %q", params["id"])
	}

	// Test matching path with parameter absent
	isMatch, params = route.matcher("/users//profile") // Double slash indicates empty segment
	if !isMatch {
		t.Error("Expected path /users//profile to match route /users/:id?/profile (optional segment absent)")
	}
	if len(params) != 1 {
		t.Errorf("Expected 1 parameter (even if empty), got %d", len(params))
	}
	if params["id"] != "" {
		t.Errorf("Expected parameter 'id' to be empty, got %q", params["id"])
	}

	// Test non-matching path (wrong structure)
	isMatch, params = route.matcher("/users/123/posts")
	if isMatch {
		t.Error("Expected path /users/123/posts not to match route /users/:id?/profile")
	}
}

// TestRouteMatcher_WildcardSegment verifies that wildcard segments (starting with *)
// capture all remaining path segments greedily.
func TestRouteMatcher_WildcardSegment(t *testing.T) {
	// Create a route definition with a wildcard segment
	route := NewRouteDefinition("/static/*filepath", nil)

	// Test matching path with single segment after wildcard
	isMatch, params := route.matcher("/static/css")
	if !isMatch {
		t.Error("Expected path /static/css to match route /static/*filepath")
	}
	if len(params) != 1 {
		t.Errorf("Expected 1 parameter, got %d", len(params))
	}
	if params["filepath"] != "css" {
		t.Errorf("Expected parameter 'filepath' to be 'css', got %q", params["filepath"])
	}

	// Test matching path with multiple segments after wildcard
	isMatch, params = route.matcher("/static/css/styles.css")
	if !isMatch {
		t.Error("Expected path /static/css/styles.css to match route /static/*filepath")
	}
	if len(params) != 1 {
		t.Errorf("Expected 1 parameter, got %d", len(params))
	}
	if params["filepath"] != "css/styles.css" {
		t.Errorf("Expected parameter 'filepath' to be 'css/styles.css', got %q", params["filepath"])
	}

	// Test matching path with nested segments
	isMatch, params = route.matcher("/static/images/icons/logo.png")
	if !isMatch {
		t.Error("Expected path /static/images/icons/logo.png to match route /static/*filepath")
	}
	if len(params) != 1 {
		t.Errorf("Expected 1 parameter, got %d", len(params))
	}
	if params["filepath"] != "images/icons/logo.png" {
		t.Errorf("Expected parameter 'filepath' to be 'images/icons/logo.png', got %q", params["filepath"])
	}

	// Test non-matching path (doesn't start with /static)
	isMatch, params = route.matcher("/assets/css/styles.css")
	if isMatch {
		t.Error("Expected path /assets/css/styles.css not to match route /static/*filepath")
	}

	// Test non-matching path (exact /static without wildcard capture)
	isMatch, params = route.matcher("/static")
	if isMatch {
		t.Error("Expected path /static not to match route /static/*filepath (wildcard requires at least one segment)")
	}
}

// TestRouteMatcher_PrecedenceOrderMatters verifies that when multiple routes could match a URL,
// the one defined first is chosen (static routes take precedence over dynamic ones).
func TestRouteMatcher_PrecedenceOrderMatters(t *testing.T) {
	// Create a router with multiple routes that could potentially match the same path
	routes := []*RouteDefinition{
		NewRouteDefinition("/users/new", nil), // Static route - should match first
		NewRouteDefinition("/users/:id", nil), // Dynamic route - should not match if static matches
	}

	router := &Router{
		routes: routes,
	}

	// Test that /users/new matches the static route, not the dynamic one
	matchedRoute, params := router.Match("/users/new")
	if matchedRoute == nil {
		t.Fatal("Expected a route to match /users/new")
	}
	if matchedRoute.Path != "/users/new" {
		t.Errorf("Expected to match static route /users/new, but matched %s", matchedRoute.Path)
	}
	if len(params) != 0 {
		t.Errorf("Expected no parameters for static route, got %v", params)
	}

	// Test that /users/123 matches the dynamic route when no static route matches
	matchedRoute, params = router.Match("/users/123")
	if matchedRoute == nil {
		t.Fatal("Expected a route to match /users/123")
	}
	if matchedRoute.Path != "/users/:id" {
		t.Errorf("Expected to match dynamic route /users/:id, but matched %s", matchedRoute.Path)
	}
	if len(params) != 1 || params["id"] != "123" {
		t.Errorf("Expected parameter 'id' to be '123', got %v", params)
	}
}

// TestRouteMatcher_MatchFiltersRegexpAndFunc verifies that MatchFilters can validate parameters
// using both regular expressions and validation functions.
func TestRouteMatcher_MatchFiltersRegexpAndFunc(t *testing.T) {
	// Test regex filter: ensure :id is a number (\d+)
	t.Run("regex filter", func(t *testing.T) {
		route := NewRouteDefinition("/users/:id", nil)
		// Add regex filter for :id parameter
		route.MatchFilters = map[string]any{
			"id": `^\d+$`, // Regex to match digits only
		}

		// Test matching path with numeric id
		isMatch, params := route.matcher("/users/123")
		if !isMatch {
			t.Error("Expected path /users/123 to match route /users/:id with regex filter")
		}
		if len(params) != 1 || params["id"] != "123" {
			t.Errorf("Expected parameter 'id' to be '123', got %v", params)
		}

		// Test non-matching path with non-numeric id
		isMatch, params = route.matcher("/users/abc")
		if isMatch {
			t.Error("Expected path /users/abc not to match route /users/:id with regex filter")
		}
	})

	// Test function filter: ensure :name is "apple" or "banana"
	t.Run("function filter", func(t *testing.T) {
		route := NewRouteDefinition("/products/:name", nil)
		// Add function filter for :name parameter
		route.MatchFilters = map[string]any{
			"name": func(value string) bool {
				return value == "apple" || value == "banana"
			},
		}

		// Test matching path with valid name
		isMatch, params := route.matcher("/products/apple")
		if !isMatch {
			t.Error("Expected path /products/apple to match route /products/:name with function filter")
		}
		if len(params) != 1 || params["name"] != "apple" {
			t.Errorf("Expected parameter 'name' to be 'apple', got %v", params)
		}

		// Test matching path with another valid name
		isMatch, params = route.matcher("/products/banana")
		if !isMatch {
			t.Error("Expected path /products/banana to match route /products/:name with function filter")
		}
		if len(params) != 1 || params["name"] != "banana" {
			t.Errorf("Expected parameter 'name' to be 'banana', got %v", params)
		}

		// Test non-matching path with invalid name
		isMatch, params = route.matcher("/products/orange")
		if isMatch {
			t.Error("Expected path /products/orange not to match route /products/:name with function filter")
		}
	})
}
