package router

import (
	"testing"

	g "maragu.dev/gomponents"
	h "maragu.dev/gomponents/html"
)

// Test nested route matching functionality
func TestNestedRoutes_BasicMatching(t *testing.T) {
	// Create child routes
	childRoute1 := Route("/profile", func(props ...any) interface{} {
		return h.Div(g.Text("Profile"))
	})
	childRoute2 := Route("/settings", func(props ...any) interface{} {
		return h.Div(g.Text("Settings"))
	})

	// Create parent route with children
	parentRoute := Route("/user", func(props ...any) interface{} {
		// Layout component that accepts child content
		if len(props) > 0 {
			if childNode, ok := props[0].(g.Node); ok {
				return h.Div(
					h.Class("user-layout"),
					h.H1(g.Text("User Dashboard")),
					childNode, // Render child content
				)
			}
		}
		// Default content when no child is matched
		return h.Div(
			h.Class("user-layout"),
			h.H1(g.Text("User Dashboard")),
			h.P(g.Text("Select an option")),
		)
	}, childRoute1, childRoute2)

	routes := []*RouteDefinition{parentRoute}
	router := New(routes, nil)

	// Test matching parent route
	matchedRoute, params := router.Match("/user")
	if matchedRoute == nil {
		t.Fatal("Expected to match parent route /user")
	}
	if matchedRoute.Path != "/user" {
		t.Errorf("Expected path '/user', got '%s'", matchedRoute.Path)
	}
	if params == nil {
		params = make(map[string]string)
	}

	// Test matching nested route
	matchedRoute, params = router.Match("/user/profile")
	if matchedRoute == nil {
		t.Fatal("Expected to match nested route /user/profile")
	}
	// Should return the child route, not the parent
	if matchedRoute.Path != "/profile" {
		t.Errorf("Expected matched route path '/profile', got '%s'", matchedRoute.Path)
	}
	if params == nil {
		params = make(map[string]string)
	}
}

func TestNestedRoutes_WithDynamicSegments(t *testing.T) {
	// Create child route with dynamic segment
	childRoute := Route("/posts/:postId", func(props ...any) interface{} {
		return h.Div(g.Text("Post Detail"))
	})

	// Create parent route with dynamic segment and child
	parentRoute := Route("/users/:userId", func(props ...any) interface{} {
		if len(props) > 0 {
			if childNode, ok := props[0].(g.Node); ok {
				return h.Div(
					h.Class("user-layout"),
					childNode,
				)
			}
		}
		return h.Div(h.Class("user-layout"), h.P(g.Text("User Home")))
	}, childRoute)

	routes := []*RouteDefinition{parentRoute}
	router := New(routes, nil)

	// Test matching nested route with parameters
	matchedRoute, params := router.Match("/users/123/posts/456")
	if matchedRoute == nil {
		t.Fatal("Expected to match nested route /users/123/posts/456")
	}

	// Should return the child route
	if matchedRoute.Path != "/posts/:postId" {
		t.Errorf("Expected matched route path '/posts/:postId', got '%s'", matchedRoute.Path)
	}

	// Should capture parameters from both parent and child
	if params["userId"] != "123" {
		t.Errorf("Expected userId parameter '123', got '%s'", params["userId"])
	}
	if params["postId"] != "456" {
		t.Errorf("Expected postId parameter '456', got '%s'", params["postId"])
	}
}

func TestNestedRoutes_DeepNesting(t *testing.T) {
	// Create deeply nested routes: /admin/users/:userId/posts/:postId/comments/:commentId
	commentRoute := Route("/comments/:commentId", func(props ...any) interface{} {
		return h.Div(g.Text("Comment Detail"))
	})

	postRoute := Route("/posts/:postId", func(props ...any) interface{} {
		if len(props) > 0 {
			if childNode, ok := props[0].(g.Node); ok {
				return h.Div(h.Class("post-layout"), childNode)
			}
		}
		return h.Div(h.Class("post-layout"), h.P(g.Text("Post Home")))
	}, commentRoute)

	userRoute := Route("/users/:userId", func(props ...any) interface{} {
		if len(props) > 0 {
			if childNode, ok := props[0].(g.Node); ok {
				return h.Div(h.Class("user-layout"), childNode)
			}
		}
		return h.Div(h.Class("user-layout"), h.P(g.Text("User Home")))
	}, postRoute)

	adminRoute := Route("/admin", func(props ...any) interface{} {
		if len(props) > 0 {
			if childNode, ok := props[0].(g.Node); ok {
				return h.Div(h.Class("admin-layout"), childNode)
			}
		}
		return h.Div(h.Class("admin-layout"), h.P(g.Text("Admin Home")))
	}, userRoute)

	routes := []*RouteDefinition{adminRoute}
	router := New(routes, nil)

	// Test matching deeply nested route
	matchedRoute, params := router.Match("/admin/users/123/posts/456/comments/789")
	if matchedRoute == nil {
		t.Fatal("Expected to match deeply nested route")
	}

	// Should return the deepest child route
	if matchedRoute.Path != "/comments/:commentId" {
		t.Errorf("Expected matched route path '/comments/:commentId', got '%s'", matchedRoute.Path)
	}

	// Should capture all parameters from the route hierarchy
	if params["userId"] != "123" {
		t.Errorf("Expected userId parameter '123', got '%s'", params["userId"])
	}
	if params["postId"] != "456" {
		t.Errorf("Expected postId parameter '456', got '%s'", params["postId"])
	}
	if params["commentId"] != "789" {
		t.Errorf("Expected commentId parameter '789', got '%s'", params["commentId"])
	}
}

func TestNestedRoutes_PartialMatching(t *testing.T) {
	// Test that partial matches work correctly
	childRoute := Route("/settings", func(props ...any) interface{} {
		return h.Div(g.Text("Settings"))
	})

	parentRoute := Route("/user", func(props ...any) interface{} {
		if len(props) > 0 {
			if childNode, ok := props[0].(g.Node); ok {
				return h.Div(h.Class("user-layout"), childNode)
			}
		}
		return h.Div(h.Class("user-layout"), h.P(g.Text("User Home")))
	}, childRoute)

	routes := []*RouteDefinition{parentRoute}
	router := New(routes, nil)

	// Test matching parent route only
	matchedRoute, params := router.Match("/user")
	if matchedRoute == nil {
		t.Fatal("Expected to match parent route /user")
	}
	if matchedRoute.Path != "/user" {
		t.Errorf("Expected matched route path '/user', got '%s'", matchedRoute.Path)
	}

	// Test non-existent child route
	matchedRoute, params = router.Match("/user/nonexistent")
	if matchedRoute != nil {
		t.Errorf("Expected no match for /user/nonexistent, but got route: %s", matchedRoute.Path)
	}
	if params != nil {
		t.Error("Expected params to be nil for non-matching route")
	}
}