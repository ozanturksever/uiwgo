package router

import (
	"testing"

	g "maragu.dev/gomponents"
	h "maragu.dev/gomponents/html"
)

func TestDebugWildcardMatching(t *testing.T) {
	// Create a simple wildcard route
	routes := []*RouteDefinition{
		Route("/files/*filepath", func(props ...any) interface{} {
			return h.Div(g.Text("File"))
		}),
	}

	// Create router
	router := New(routes, nil)

	// Test wildcard matching
	t.Logf("Testing wildcard route /files/*filepath")
	route, params := router.Match("/files/docs/readme.txt")
	if route == nil {
		t.Logf("No match found for /files/docs/readme.txt")
	} else {
		t.Logf("Matched route: %s", route.Path)
		t.Logf("Params: %v", params)
	}

	// Test the matcher directly
	wildcardRoute := routes[0]
	if wildcardRoute.matcher == nil {
		wildcardRoute.matcher = compileMatcher(wildcardRoute)
	}
	isMatch, directParams := wildcardRoute.matcher("/files/docs/readme.txt")
	t.Logf("Direct matcher result: isMatch=%v, params=%v", isMatch, directParams)
}