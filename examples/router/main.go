//go:build js && wasm

package main

import (
	"fmt"
	"strconv"

	"github.com/ozanturksever/uiwgo/comps"
	"github.com/ozanturksever/uiwgo/logutil"
	"github.com/ozanturksever/uiwgo/reactivity"
	"github.com/ozanturksever/uiwgo/router"
	g "maragu.dev/gomponents"
	a "maragu.dev/gomponents/html"
)

// Mock user data
type User struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
	Role  string `json:"role"`
}

var mockUsers = []User{
	{1, "John Doe", "john@example.com", "admin"},
	{2, "Jane Smith", "jane@example.com", "user"},
	{3, "Bob Wilson", "bob@example.com", "user"},
	{4, "Alice Brown", "alice@example.com", "editor"},
	{5, "Charlie Davis", "charlie@example.com", "user"},
}

func main() {
	logutil.Log("Starting Router Example App")

	// Initialize navigation system first
	router.InitializeNavigation()

	// Register routes
	setupRoutes()

	// Mount the app with RouterProvider
	comps.Mount("app", func() g.Node {
		return router.RouterProvider(router.RouterProps{
			BasePath: "/",
			BeforeGuard: func(to *router.Location) bool {
				logutil.Log("Navigation guard: going to", to.Pathname)
				
				// Example: block access to admin routes if not admin
				if to.Pathname == "/admin" {
					// In a real app, you'd check user authentication/authorization
					logutil.Log("Admin route access check")
				}
				
				return true // Allow all navigation for demo
			},
			AfterGuard: func(from, to *router.Location) {
				logutil.Log("Navigation completed from", from.Pathname, "to", to.Pathname)
			},
			Children: []g.Node{AppComponent()},
		})
	})

	// Prevent the program from exiting
	select {}
}

func setupRoutes() {
	// Home route
	router.Route(router.RouteProps{
		Path: "/",
		Component: func(match *router.RouteMatch) g.Node {
			return HomePage()
		},
	})

	// About route
	router.Route(router.RouteProps{
		Path: "/about",
		Component: func(match *router.RouteMatch) g.Node {
			return AboutPage()
		},
	})

	// Users list route
	router.Route(router.RouteProps{
		Path: "/users",
		Component: func(match *router.RouteMatch) g.Node {
			return UsersPage()
		},
	})

	// User detail route with parameter
	router.Route(router.RouteProps{
		Path: "/users/:id",
		Component: func(match *router.RouteMatch) g.Node {
			return UserDetailPage(match)
		},
	})

	// Admin route (protected)
	router.ProtectedRoute(router.ProtectedRouteProps{
		Path: "/admin",
		Component: func(match *router.RouteMatch) g.Node {
			return AdminPage()
		},
		CanActivate: func(match *router.RouteMatch) bool {
			// In a real app, check user permissions here
			return true // Allow for demo
		},
		RedirectTo: "/login",
	})

	// Login route
	router.Route(router.RouteProps{
		Path: "/login",
		Component: func(match *router.RouteMatch) g.Node {
			return LoginPage()
		},
	})

	// Wildcard route for 404 (must be last)
	router.Route(router.RouteProps{
		Path: "/*",
		Component: func(match *router.RouteMatch) g.Node {
			return NotFoundPage(match)
		},
	})
}

// AppComponent is the root component with navigation
func AppComponent() g.Node {
	return a.Div(
		a.Class("container"),
		Header(),
		Navigation(),
		Breadcrumbs(),
		a.Div(
			a.Class("main"),
			router.RouterOutlet(),
		),
	)
}

// Header component
func Header() g.Node {
	return a.Div(
		a.Class("header"),
		a.H1(g.Text("Router Example")),
		a.P(g.Text("Demonstrating client-side routing with oiwgo")),
	)
}

// Navigation component with active link highlighting
func Navigation() g.Node {
	return a.Nav(
		a.Class("nav"),
		router.NavLink(router.NavLinkProps{
			To:          "/",
			End:         true, // Exact match for home
			Class:       "",
			ActiveClass: "active",
			Children:    []g.Node{g.Text("Home")},
		}),
		router.NavLink(router.NavLinkProps{
			To:          "/about",
			Class:       "",
			ActiveClass: "active",
			Children:    []g.Node{g.Text("About")},
		}),
		router.NavLink(router.NavLinkProps{
			To:          "/users",
			Class:       "",
			ActiveClass: "active",
			Children:    []g.Node{g.Text("Users")},
		}),
		router.NavLink(router.NavLinkProps{
			To:          "/admin",
			Class:       "",
			ActiveClass: "active",
			Children:    []g.Node{g.Text("Admin")},
		}),
		router.NavLink(router.NavLinkProps{
			To:          "/login",
			Class:       "",
			ActiveClass: "active",
			Children:    []g.Node{g.Text("Login")},
		}),
	)
}

// Breadcrumbs component showing current route path
func Breadcrumbs() g.Node {
	location := router.UseLocation()
	
	return a.Div(
		a.Class("breadcrumbs"),
		g.Text("Current Path: "),
		a.Code(comps.BindText(func() string {
			return location.Get().Pathname
		})),
	)
}

// HomePage component
func HomePage() g.Node {
	return a.Div(
		RouteInfo("Home Page", "This is the landing page of the router example."),
		a.Div(
			a.Class("card"),
			a.H3(g.Text("Welcome to the Router Demo")),
			a.P(g.Text("This example demonstrates:")),
			a.Ul(
				a.Li(g.Text("Basic routing with path matching")),
				a.Li(g.Text("Route parameters and wildcards")),
				a.Li(g.Text("Navigation guards and protection")),
				a.Li(g.Text("Programmatic navigation")),
				a.Li(g.Text("Active link highlighting")),
				a.Li(g.Text("404 error handling")),
			),
			a.P(g.Text("Try navigating through the different sections using the navigation above.")),
		),
	)
}

// AboutPage component
func AboutPage() g.Node {
	return a.Div(
		RouteInfo("About Page", "Information about the router system."),
		a.Div(
			a.Class("card"),
			a.H3(g.Text("About This Router")),
			a.P(g.Text("This is a client-side router implementation for the oiwgo framework, inspired by SolidJS Router.")),
			a.H4(g.Text("Features:")),
			a.Ul(
				a.Li(g.Text("🔄 Reactive routing with signals")),
				a.Li(g.Text("🎯 Path parameters and wildcards")),
				a.Li(g.Text("🛡️ Route guards and protection")),
				a.Li(g.Text("🔗 Navigation components (Link, NavLink)")),
				a.Li(g.Text("🪝 Router hooks (useLocation, useParams, useNavigate)")),
				a.Li(g.Text("📱 Browser history integration")),
				a.Li(g.Text("⚡ Live navigation without page reloads")),
			),
		),
	)
}

// UsersPage component showing list of users
func UsersPage() g.Node {
	query := router.UseQuery()
	
	// Get current search query
	searchQuery := query.Get()["search"]
	
	return a.Div(
		RouteInfo("Users Page", "List of all users with search functionality."),
		a.Div(
			a.Class("card"),
			a.H3(g.Text("Users Directory")),
			g.If(searchQuery != "", 
				a.P(g.Text("Searching for: "), a.Strong(g.Text(searchQuery))),
			),
			a.Div(
				a.Class("user-list"),
				g.Group(renderUserCards(searchQuery)),
			),
		),
	)
}

// UserDetailPage component showing individual user details
func UserDetailPage(match *router.RouteMatch) g.Node {
	userID := match.Params["id"]
	
	// Find user by ID
	var user *User
	if id, err := strconv.Atoi(userID); err == nil {
		for _, u := range mockUsers {
			if u.ID == id {
				user = &u
				break
			}
		}
	}
	
	if user == nil {
		return a.Div(
			RouteInfo("User Not Found", "The requested user does not exist."),
			a.Div(
				a.Class("error"),
				a.H3(g.Text("User Not Found")),
				a.P(g.Text("No user found with ID: "), a.Code(g.Text(userID))),
				router.LinkWithHandler(router.LinkProps{
					To:       "/users",
					Class:    "btn",
					Children: []g.Node{g.Text("← Back to Users")},
				}),
			),
		)
	}
	
	return a.Div(
		RouteInfo("User Detail", fmt.Sprintf("Details for user %s (ID: %s)", user.Name, userID)),
		a.Div(
			a.Class("card"),
			a.H3(g.Text("User Details")),
			a.P(a.Strong(g.Text("ID: ")), g.Text(fmt.Sprintf("%d", user.ID))),
			a.P(a.Strong(g.Text("Name: ")), g.Text(user.Name)),
			a.P(a.Strong(g.Text("Email: ")), g.Text(user.Email)),
			a.P(a.Strong(g.Text("Role: ")), g.Text(user.Role)),
			a.Div(
				router.LinkWithHandler(router.LinkProps{
					To:       "/users",
					Class:    "btn btn-secondary",
					Children: []g.Node{g.Text("← Back to Users")},
				}),
			),
		),
	)
}

// AdminPage component (protected route)
func AdminPage() g.Node {
	return a.Div(
		RouteInfo("Admin Page", "Protected administrative area."),
		a.Div(
			a.Class("card"),
			a.H3(g.Text("Admin Dashboard")),
			a.P(g.Text("This is a protected route that requires authorization.")),
			a.P(g.Text("In a real application, access would be controlled by authentication and authorization checks.")),
		),
	)
}

// LoginPage component
func LoginPage() g.Node {
	return a.Div(
		RouteInfo("Login Page", "User authentication form."),
		a.Div(
			a.Class("card"),
			a.H3(g.Text("Login")),
			a.P(g.Text("This is a mock login page.")),
		),
	)
}

// NotFoundPage component for 404 errors
func NotFoundPage(match *router.RouteMatch) g.Node {
	location := router.UseLocation()
	
	return a.Div(
		RouteInfo("404 Not Found", "The requested page could not be found."),
		a.Div(
			a.Class("error"),
			a.H3(g.Text("404 - Page Not Found")),
			a.P(g.Text("The page you're looking for doesn't exist.")),
			a.P(
				a.Strong(g.Text("Requested path: ")),
				a.Code(comps.BindText(func() string {
					return location.Get().Pathname
				})),
			),
			g.If(match.Wildcard != "",
				a.P(
					a.Strong(g.Text("Wildcard match: ")),
					a.Code(g.Text(match.Wildcard)),
				),
			),
			router.LinkWithHandler(router.LinkProps{
				To:       "/",
				Class:    "btn",
				Children: []g.Node{g.Text("← Go Home")},
			}),
		),
	)
}

// Helper function to render route information
func RouteInfo(title, description string) g.Node {
	location := router.UseLocation()
	params := router.UseParams()
	query := router.UseQuery()
	
	// Create computed signals for conditional rendering
	hasParams := reactivity.CreateMemo(func() bool {
		return len(params.Get()) > 0
	})
	hasQuery := reactivity.CreateMemo(func() bool {
		return len(query.Get()) > 0
	})
	
	return a.Div(
		a.Class("route-info"),
		a.H3(g.Text(title)),
		a.P(g.Text(description)),
		a.P(
			a.Strong(g.Text("Path: ")),
			a.Code(comps.BindText(func() string {
				return location.Get().Pathname
			})),
		),
		comps.Show(comps.ShowProps{
			When: hasParams,
			Children: a.P(
				a.Strong(g.Text("Parameters: ")),
				a.Code(comps.BindText(func() string {
					return fmt.Sprintf("%+v", params.Get())
				})),
			),
		}),
		comps.Show(comps.ShowProps{
			When: hasQuery,
			Children: a.P(
				a.Strong(g.Text("Query: ")),
				a.Code(comps.BindText(func() string {
					return fmt.Sprintf("%+v", query.Get())
				})),
			),
		}),
	)
}

// Helper function to render user cards
func renderUserCards(searchQuery string) []g.Node {
	var cards []g.Node
	
	for _, user := range mockUsers {
		// Simple search filter
		if searchQuery != "" {
			if !contains(user.Name, searchQuery) && !contains(user.Email, searchQuery) {
				continue
			}
		}
		
		card := a.Div(
			a.Class("user-card"),
			a.H4(g.Text(user.Name)),
			a.P(g.Text(user.Email)),
			a.P(g.Text("Role: "+user.Role)),
			router.LinkWithHandler(router.LinkProps{
				To:       fmt.Sprintf("/users/%d", user.ID),
				Class:    "btn",
				Children: []g.Node{g.Text("View Details")},
			}),
		)
		
		cards = append(cards, card)
	}
	
	if len(cards) == 0 {
		return []g.Node{
			a.P(g.Text("No users found matching the search criteria.")),
		}
	}
	
	return cards
}

// Helper function for case-insensitive string contains
func contains(str, substr string) bool {
	return len(str) >= len(substr) && 
		   (str == substr || 
		    (len(str) > len(substr) && 
		     someSubstring(str, substr)))
}

func someSubstring(str, substr string) bool {
	for i := 0; i <= len(str)-len(substr); i++ {
		if str[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
