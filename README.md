# Golid

**Golid** is a simple, solid, Go-native frontend framework for WebAssembly applications.  
It focuses on clarity, modularity, and reactivity — without the complexity of heavy Virtual DOM systems.

A minimal, Go-native frontend framework with signals, components, and WebAssembly — no Node.js, no npm, no JSX, no bundlers.


## ✨ What is Golid?

**Golid** (short for Go + Solid) is a lightweight frontend framework written entirely in Go, compiled to WebAssembly. It’s inspired by frameworks like Solid.js, but built for Go developers who want simplicity, control, and zero JS toolchain pain.

With Golid, you can build reactive web apps using:
- ✅ Pure Go
- ✅ Signals and reactive components  
- ✅ Built-in router with SPA navigation
- ✅ Tiny `.wasm` bundles (TinyGo optional)
- ✅ No Node.js, no npm, no React, no JSX, no bundlers
- Command line ""golid-dev" (plus auto-compile and hot-reload (client-side))
- Self-sufficient (no external tools needed (no external server, no bash, no Make))

---

## 🚀 Quick Start

1. Clone the repo:
   ```bash
   git clone https://github.com/serge-hulne/Golid.git
   cd Golid
   ```

2. Build the CLI:
    ```bash
    cd cmd/devserver
    go build
    mv golid-dev ../..
	```

3. Run the CLI (development server) :
    ```bash
    ./golid-dev
	```

4. Watch the app in a browser:
    ```bash
	http://localhost:8090
	```

## 💡 Example: Counter Component

```

func CounterComponent() Node {
	// Observable (represents the state of the app)
	count := golid.NewSignal(0)

	return Div(
		Style("border: 1px solid orange; padding: 10px; margin: 10px;"),

		// Bind text Element to the reactive count signal (observable)
		golid.Bind(func() Node {
			return Div(Text(fmt.Sprintf("Count = %d", count.Get())))
		}),

		// [+] Button element
		Button(
			Style("margin: 3px;"),
			Text("+"),
			golid.OnClick(func() {
				count.Set(count.Get() + 1)
			}),
		),

		// [-] Button element
		Button(
			Style("margin: 3px;"),
			Text("-"),
			golid.OnClick(func() {
				count.Set(count.Get() - 1)
			}),
		),
	)
}
    
```

## 🧭 Router System

Golid includes a powerful, SolidJS-inspired router system for building Single Page Applications with client-side navigation:

### Basic Router Setup

```go
func main() {
    // Initialize the router
    router := golid.NewRouter()
    golid.SetGlobalRouter(router)

    // Define routes
    router.AddRoute("/", func(params golid.RouteParams) Node {
        return HomePage()
    })
    
    router.AddRoute("/about", func(params golid.RouteParams) Node {
        return AboutPage()
    })
    
    // Route with parameters
    router.AddRoute("/user/:id", func(params golid.RouteParams) Node {
        userID := params["id"]
        return UserPage(userID)
    })
    
    // Multi-parameter routes
    router.AddRoute("/posts/:category/:slug", func(params golid.RouteParams) Node {
        category := params["category"]
        slug := params["slug"]
        return PostPage(category, slug)
    })

    // Render the app with router
    golid.Render(RouterApp())
    golid.Run()
}

// Main app component with navigation
func RouterApp() Node {
    return Div(
        NavigationBar(),
        golid.RouterOutlet(), // Renders the current route
    )
}
```

### Navigation Components

```go
func NavigationBar() Node {
    return Nav(
        // RouterLink creates navigatable links
        golid.RouterLink("/", Text("Home")),
        golid.RouterLink("/about", Text("About")),
        golid.RouterLink("/user/123", Text("User Profile")),
    )
}
```

### Route Parameters and Query Strings

```go
func UserPage(userID string) Node {
    // Access route parameters
    params := golid.UseParams()
    
    // Access query parameters (?name=john&age=30)
    query := golid.UseQuery()
    
    // Access current path
    location := golid.UseLocation()
    
    return Div(
        H1(Text("User Profile")),
        P(golid.BindText(func() string {
            return fmt.Sprintf("User ID: %s", params.Get()["id"])
        })),
        P(golid.BindText(func() string {
            return fmt.Sprintf("Current path: %s", location.Get())
        })),
    )
}
```

### Programmatic Navigation

```go
func LoginPage() Node {
    return Div(
        H1(Text("Login")),
        Button(
            Text("Login"),
            golid.OnClick(func() {
                // Navigate programmatically
                if globalRouter != nil {
                    globalRouter.Navigate("/dashboard")
                }
            }),
        ),
    )
}
```

### Route Guards

```go
// Add protected routes with guards
router.AddRouteWithGuard("/admin", 
    func(params golid.RouteParams) Node {
        return AdminPage()
    },
    func(params golid.RouteParams) bool {
        // Return true if user is authorized
        return isUserAdmin()
    },
)
```

### Features

- ✅ **Declarative routing** - Define routes with simple patterns
- ✅ **Route parameters** - Extract dynamic segments from URLs (`/user/:id`)
- ✅ **Query string parsing** - Access URL query parameters
- ✅ **Programmatic navigation** - Navigate with `Navigate()` and `Replace()`
- ✅ **Route guards** - Protect routes with authorization logic
- ✅ **Browser history integration** - Full back/forward button support
- ✅ **Reactive route state** - Route changes trigger component updates
- ✅ **404 handling** - Automatic fallback for unmatched routes


## ❌ What Golid Does Not Require

- No Node.js
- No npm or yarn
- No Parcel, Webpack, Vite, or other bundlers
- No React, Vue, Svelte, Solid.js, or JSX
- No go:generate or code generation

- Just:
✅ Go

## 🛣 Roadmap

- [x] ✅ **Add routing system** - Complete SPA router with navigation, parameters, and guards
- [] Add built-in UI components (e.g., Toggle, Input, Form)
- [] Provide example apps and templates
- [] Optional CSS helper system


## 📜 License

Golid is open source under the GNU General Public License v3.

