package router

import (
	"testing"

	g "maragu.dev/gomponents"
	h "maragu.dev/gomponents/html"
)

// Debug test to understand nested route matching
func TestDebugNestedRouteMatching(t *testing.T) {
	// Create simple nested routes
	childRoute := Route("/profile", func(props ...any) interface{} {
		return h.Div(g.Text("Profile"))
	})

	parentRoute := Route("/user", func(props ...any) interface{} {
		return h.Div(g.Text("User Layout"))
	}, childRoute)

	routes := []*RouteDefinition{parentRoute}
	router := New(routes, nil)

	// Test parent route matching
	t.Logf("Testing parent route /user")
	matchedRoute, params := router.Match("/user")
	if matchedRoute == nil {
		t.Fatalf("Expected to match parent route /user, got nil")
	}
	t.Logf("Matched route: %s", matchedRoute.Path)
	t.Logf("Params: %v", params)

	// Test nested route matching
	t.Logf("\nTesting nested route /user/profile")
	matchedRoute, params = router.Match("/user/profile")
	if matchedRoute == nil {
		t.Fatalf("Expected to match nested route /user/profile, got nil")
	}
	t.Logf("Matched route: %s", matchedRoute.Path)
	t.Logf("Params: %v", params)

	// Test calculateRemainingPath function directly
	t.Logf("\nTesting calculateRemainingPath function")
	remaining := calculateRemainingPath("/user/profile", "/user", make(map[string]string))
	t.Logf("Remaining path for '/user/profile' after '/user': '%s'", remaining)

	// Test splitPath function
	t.Logf("\nTesting splitPath function")
	segments := splitPath("/user/profile")
	t.Logf("Segments for '/user/profile': %v", segments)
	segments = splitPath("/user")
	t.Logf("Segments for '/user': %v", segments)
}