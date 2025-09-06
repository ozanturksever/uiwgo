//go:build js && wasm

package main

import (
	"github.com/ozanturksever/logutil"
	"github.com/ozanturksever/uiwgo/router"
	"github.com/ozanturksever/uiwgo/wasm"
	"honnef.co/go/js/dom/v2"
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

// Global router instance for accessing params and location
var appRouter *router.Router

// HomeComponent renders the home page with navigation links
func HomeComponent(props ...any) interface{} {
	return Div(
		Class("p-6 max-w-4xl mx-auto"),
		H1(Class("text-3xl font-bold mb-6"), Text("Router Demo - Home")),
		P(Class("mb-4"), Text("Welcome to the UIWGo Router Demo! This example showcases all router features:")),
		Ul(Class("list-disc list-inside space-y-2 mb-6"),
			Li(Text("Static routes")),
			Li(Text("Dynamic segments with parameters")),
			Li(Text("Optional parameters")),
			Li(Text("Wildcard routes")),
			Li(Text("Nested routes")),
			Li(Text("Programmatic navigation")),
		),
		Nav(Class("space-y-2"),
			H2(Class("text-xl font-semibold mb-3"), Text("Navigation:")),
			Div(Class("flex flex-wrap gap-2"),
				router.A("/about", Class("bg-blue-500 text-white px-4 py-2 rounded hover:bg-blue-600"), Text("About")),
				router.A("/users", Class("bg-green-500 text-white px-4 py-2 rounded hover:bg-green-600"), Text("Users")),
				router.A("/users/123", Class("bg-purple-500 text-white px-4 py-2 rounded hover:bg-purple-600"), Text("User 123")),
				router.A("/users/456/profile", Class("bg-indigo-500 text-white px-4 py-2 rounded hover:bg-indigo-600"), Text("User 456 Profile")),
				router.A("/files/docs/readme.txt", Class("bg-yellow-500 text-white px-4 py-2 rounded hover:bg-yellow-600"), Text("File Browser")),
				router.A("/admin", Class("bg-red-500 text-white px-4 py-2 rounded hover:bg-red-600"), Text("Admin")),
				router.A("/admin/settings", Class("bg-pink-500 text-white px-4 py-2 rounded hover:bg-pink-600"), Text("Admin Settings")),
			),
		),
	)
}

// AboutComponent renders the about page
func AboutComponent(props ...any) interface{} {
	return Div(
		Class("p-6 max-w-4xl mx-auto"),
		H1(Class("text-3xl font-bold mb-6"), Text("About UIWGo Router")),
		P(Class("mb-4"), Text("UIWGo Router is a high-performance, reactive routing solution built in pure Go for WebAssembly applications.")),
		P(Class("mb-4"), Text("Key features:")),
		Ul(Class("list-disc list-inside space-y-1 mb-6"),
			Li(Text("Fine-grained reactivity inspired by SolidJS")),
			Li(Text("Zero JavaScript dependencies")),
			Li(Text("Type-safe route definitions")),
			Li(Text("Declarative component-based architecture")),
			Li(Text("Browser history integration")),
		),
		Div(Class("space-x-2"),
			router.A("/users", Class("bg-green-500 text-white px-4 py-2 rounded hover:bg-green-600"), Text("Users")),
			router.A("/", Class("bg-blue-500 text-white px-4 py-2 rounded hover:bg-blue-600"), Text("← Back to Home")),
		),
	)
}

// UsersListComponent renders the users list page
func UsersListComponent(props ...any) interface{} {
	return Div(
		Class("p-6 max-w-4xl mx-auto"),
		H1(Class("text-3xl font-bold mb-6"), Text("Users List")),
		P(Class("mb-4"), Text("This is a static route that lists all users.")),
		Div(Class("grid grid-cols-1 md:grid-cols-2 gap-4 mb-6"),
			Div(Class("border p-4 rounded"),
				H3(Class("font-semibold"), Text("User 123")),
				P(Class("text-gray-600"), Text("john@example.com")),
				router.A("/users/123", Class("text-blue-500 hover:underline"), Text("View Profile")),
			),
			Div(Class("border p-4 rounded"),
				H3(Class("font-semibold"), Text("User 456")),
				P(Class("text-gray-600"), Text("jane@example.com")),
				router.A("/users/456", Class("text-blue-500 hover:underline"), Text("View Profile")),
			),
		),
		router.A("/", Class("bg-blue-500 text-white px-4 py-2 rounded hover:bg-blue-600"), Text("← Back to Home")),
	)
}

// UserProfileComponent renders a user profile page with dynamic ID
func UserProfileComponent(props ...any) interface{} {
	params := appRouter.Params()
	userID := params["id"]

	return Div(
		Class("p-6 max-w-4xl mx-auto"),
		H1(Class("text-3xl font-bold mb-6"), Text("User Profile")),
		Div(Class("bg-gray-100 p-4 rounded mb-4"),
			P(Class("font-semibold"), Text("User ID: "), Text(userID)),
			P(Class("text-gray-600"), Text("This demonstrates dynamic route parameters.")),
		),
		P(Class("mb-4"), Text("Profile information for user "), Strong(Text(userID)), Text(" would be displayed here.")),
		Div(Class("space-x-2"),
			router.A("/users/"+userID+"/profile", Class("bg-green-500 text-white px-4 py-2 rounded hover:bg-green-600"), Text("View Extended Profile")),
			router.A("/users", Class("bg-gray-500 text-white px-4 py-2 rounded hover:bg-gray-600"), Text("← Back to Users")),
			router.A("/", Class("bg-blue-500 text-white px-4 py-2 rounded hover:bg-blue-600"), Text("← Home")),
		),
	)
}

// UserExtendedProfileComponent renders extended user profile with optional segments
func UserExtendedProfileComponent(props ...any) interface{} {
	params := appRouter.Params()
	userID := params["id"]
	section := params["section"] // This could be empty for optional segments

	var sectionContent Node
	if section == "" {
		sectionContent = P(Class("mb-4"), Text("Default profile view for user "), Strong(Text(userID)), Text("."))
	} else {
		sectionContent = P(Class("mb-4"), Text("Showing "), Strong(Text(section)), Text(" section for user "), Strong(Text(userID)), Text("."))
	}

	return Div(
		Class("p-6 max-w-4xl mx-auto"),
		H1(Class("text-3xl font-bold mb-6"), Text("Extended User Profile")),
		Div(Class("bg-gray-100 p-4 rounded mb-4"),
			P(Class("font-semibold"), Text("User ID: "), Text(userID)),
			P(Class("font-semibold"), Text("Section: "), Text(func() string {
				if section == "" {
					return "(default profile)"
				}
				return section
			}())),
			P(Class("text-gray-600"), Text("This demonstrates optional route parameters.")),
		),
		sectionContent,
		Div(Class("space-x-2"),
			router.A("/users/"+userID, Class("bg-gray-500 text-white px-4 py-2 rounded hover:bg-gray-600"), Text("← Back to Profile")),
			router.A("/users", Class("bg-gray-500 text-white px-4 py-2 rounded hover:bg-gray-600"), Text("← Back to Users")),
			router.A("/", Class("bg-blue-500 text-white px-4 py-2 rounded hover:bg-blue-600"), Text("← Home")),
		),
	)
}

// FileBrowserComponent demonstrates wildcard routes
func FileBrowserComponent(props ...any) interface{} {
	params := appRouter.Params()
	filepath := params["filepath"]

	return Div(
		Class("p-6 max-w-4xl mx-auto"),
		H1(Class("text-3xl font-bold mb-6"), Text("File Browser")),
		Div(Class("bg-gray-100 p-4 rounded mb-4"),
			P(Class("font-semibold"), Text("File Path: "), Code(Class("bg-gray-200 px-2 py-1 rounded"), Text(filepath))),
			P(Class("text-gray-600"), Text("This demonstrates wildcard route matching.")),
		),
		P(Class("mb-4"), Text("File content for "), Strong(Text(filepath)), Text(" would be displayed here.")),
		Div(Class("bg-white border p-4 rounded mb-4"),
			Pre(Class("text-sm text-gray-700"), Text("# Sample file content\n\nThis is a demonstration of how wildcard routes\ncan capture file paths and display content.\n\nPath: "+filepath)),
		),
		Div(Class("space-x-2"),
			router.A("/files/docs", Class("bg-yellow-500 text-white px-4 py-2 rounded hover:bg-yellow-600"), Text("Browse Docs")),
			router.A("/files/src/main.go", Class("bg-yellow-500 text-white px-4 py-2 rounded hover:bg-yellow-600"), Text("Browse Source")),
			router.A("/", Class("bg-blue-500 text-white px-4 py-2 rounded hover:bg-blue-600"), Text("← Home")),
		),
	)
}

// AdminLayoutComponent demonstrates nested routes
func AdminLayoutComponent(props ...any) interface{} {
	// Get child content from props (for nested routing)
	var content Node
	if len(props) > 0 {
		if childNode, ok := props[0].(Node); ok {
			content = childNode
		} else {
			// Fallback content if no child is provided
			content = Div(Class("bg-white border p-4 rounded"),
				P(Text("Select an option from the navigation above.")),
			)
		}
	} else {
		// Default content when no child is matched
		content = Div(Class("bg-white border p-4 rounded"),
			P(Text("Select an option from the navigation above.")),
		)
	}

	return Div(
		Class("p-6 max-w-4xl mx-auto"),
		H1(Class("text-3xl font-bold mb-6"), Text("Admin Panel")),
		Div(Class("bg-red-50 border border-red-200 p-4 rounded mb-6"),
			P(Class("text-red-700"), Text("⚠️ This is the admin area. Access restricted.")),
		),
		Nav(Class("mb-6"),
			Div(Class("flex space-x-2"),
				router.A("/admin", Class("bg-red-500 text-white px-4 py-2 rounded hover:bg-red-600"), Text("Dashboard")),
				router.A("/admin/settings", Class("bg-red-500 text-white px-4 py-2 rounded hover:bg-red-600"), Text("Settings")),
				router.A("/admin/users", Class("bg-red-500 text-white px-4 py-2 rounded hover:bg-red-600"), Text("User Management")),
			),
		),
		// Render the child content
		content,
		Div(Class("mt-6 pt-4 border-t"),
			router.A("/", Class("bg-blue-500 text-white px-4 py-2 rounded hover:bg-blue-600"), Text("← Back to Home")),
		),
	)
}

// AdminDashboardComponent renders the admin dashboard
func AdminDashboardComponent(props ...any) interface{} {
	return Div(
		Class("bg-white border p-4 rounded"),
		H2(Class("text-xl font-semibold mb-4"), Text("Admin Dashboard")),
		P(Class("mb-4"), Text("Welcome to the admin dashboard. This demonstrates nested routing.")),
		Div(Class("grid grid-cols-1 md:grid-cols-3 gap-4"),
			Div(Class("bg-blue-100 p-4 rounded text-center"),
				H3(Class("font-semibold"), Text("Total Users")),
				P(Class("text-2xl font-bold text-blue-600"), Text("1,234")),
			),
			Div(Class("bg-green-100 p-4 rounded text-center"),
				H3(Class("font-semibold"), Text("Active Sessions")),
				P(Class("text-2xl font-bold text-green-600"), Text("89")),
			),
			Div(Class("bg-yellow-100 p-4 rounded text-center"),
				H3(Class("font-semibold"), Text("Pending Tasks")),
				P(Class("text-2xl font-bold text-yellow-600"), Text("12")),
			),
		),
	)
}

// AdminSettingsComponent renders the admin settings page
func AdminSettingsComponent(props ...any) interface{} {
	return Div(
		Class("bg-white border p-4 rounded"),
		H2(Class("text-xl font-semibold mb-4"), Text("Admin Settings")),
		P(Class("mb-4"), Text("Configure system settings here. This is a nested route under /admin.")),
		Form(Class("space-y-4"),
			Div(
				Label(Class("block text-sm font-medium text-gray-700"), Text("Site Name")),
				Input(Type("text"), Value("UIWGo Router Demo"), Class("mt-1 block w-full border border-gray-300 rounded-md px-3 py-2")),
			),
			Div(
				Label(Class("block text-sm font-medium text-gray-700"), Text("Max Users")),
				Input(Type("number"), Value("1000"), Class("mt-1 block w-full border border-gray-300 rounded-md px-3 py-2")),
			),
			Div(
				Button(Type("button"), Class("bg-green-500 text-white px-4 py-2 rounded hover:bg-green-600"), Text("Save Settings")),
			),
		),
	)
}

// NotFoundComponent renders a 404 page
func NotFoundComponent(props ...any) interface{} {
	location := appRouter.Location()
	return Div(
		Class("p-6 max-w-4xl mx-auto text-center"),
		H1(Class("text-4xl font-bold mb-6 text-red-600"), Text("404 - Page Not Found")),
		P(Class("mb-4 text-gray-600"), Text("The page you're looking for doesn't exist.")),
		Div(Class("bg-gray-100 p-4 rounded mb-6"),
			P(Class("font-mono text-sm"), Text("Requested path: "), Code(Text(location.Pathname))),
		),
		router.A("/", Class("bg-blue-500 text-white px-4 py-2 rounded hover:bg-blue-600"), Text("← Go Home")),
	)
}

func main() {
	logutil.Log("Starting main function")

	// Initialize WASM and bridge
	if err := wasm.QuickInit(); err != nil {
		logutil.Logf("Failed to initialize WASM: %v", err)
		return
	}


	// Define comprehensive routes showcasing all router features
	routes := []*router.RouteDefinition{
		// Static routes
		router.Route("/", HomeComponent),
		router.Route("/about", AboutComponent),
		router.Route("/users", UsersListComponent),

		// Dynamic routes with parameters
		router.Route("/users/:id", UserProfileComponent),

		// Optional parameters
		router.Route("/users/:id/profile/:section?", UserExtendedProfileComponent),

		// Wildcard routes
		router.Route("/files/*filepath", FileBrowserComponent),

		// Nested routes - proper nested structure
		router.Route("/admin", AdminLayoutComponent,
			// Child routes for admin section
			router.Route("/", AdminDashboardComponent),        // matches /admin exactly
			router.Route("/settings", AdminSettingsComponent), // matches /admin/settings
		),

		// Catch-all route for 404
		router.Route("/*", NotFoundComponent),
	}

	// Get the app element to use as outlet
	outlet := dom.GetWindow().Document().GetElementByID("app")
	if outlet == nil {
		panic("Could not find #app element")
	}

	// Create router with the outlet element
	appRouter = router.New(routes, outlet)
	logutil.Log("Router created successfully")

	// Keep the program running
	select {}
}
