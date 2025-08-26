package router

import (
	"testing"

	g "maragu.dev/gomponents"
	h "maragu.dev/gomponents/html"
)

func TestDebugCatchAllMatching(t *testing.T) {
	// Create routes similar to router demo
	routes := []*RouteDefinition{
		// Static routes
		Route("/", func(props ...any) interface{} {
			return h.Div(g.Text("Home"))
		}),
		Route("/about", func(props ...any) interface{} {
			return h.Div(g.Text("About"))
		}),
		
		// Catch-all route for 404
		Route("*", func(props ...any) interface{} {
			return h.Div(g.Text("404 - Page Not Found"))
		}),
	}

	// Create router
	router := New(routes, nil)

	// Test catch-all matching with non-existent route
	t.Logf("Testing catch-all route * with /non-existent-route")
	route, params := router.Match("/non-existent-route")
	if route == nil {
		t.Errorf("No match found for /non-existent-route")
	} else {
		t.Logf("Matched route: %s", route.Path)
		t.Logf("Params: %v", params)
	}

	// Test the catch-all matcher directly
	catchAllRoute := routes[2] // The * route
	if catchAllRoute.matcher == nil {
		catchAllRoute.matcher = compileMatcher(catchAllRoute)
	}
	isMatch, directParams := catchAllRoute.matcher("/non-existent-route")
	t.Logf("Direct catch-all matcher result: isMatch=%v, params=%v", isMatch, directParams)

	// Test home route to make sure it works
	t.Logf("Testing home route / with /")
	route, params = router.Match("/")
	if route == nil {
		t.Errorf("No match found for /")
	} else {
		t.Logf("Matched route: %s", route.Path)
		t.Logf("Params: %v", params)
	}
}