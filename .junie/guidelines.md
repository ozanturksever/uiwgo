# Golid Development Guidelines

## Project Overview

Golid is a Go-native frontend framework for WebAssembly applications that provides reactive components, signals, and client-side routing without requiring Node.js, npm, or JavaScript toolchain dependencies.

**Key Technologies:**
- Go 1.23.0+ with WebAssembly compilation
- gomponents for HTML generation
- Reactive signals system
- Built-in SPA router
- Development server with hot reload

## Build & Configuration Instructions

### Prerequisites

1. **Go 1.23.0+** - Required for generics support in Signal system
2. **Optional: TinyGo** - For smaller WASM bundles (configured in devenv.nix)
3. **Optional: devenv** - For reproducible development environment

### Development Setup

#### Method 1: Using devenv (Recommended)
```bash
# If you have devenv installed
devenv shell
```

#### Method 2: Manual Setup
```bash
# Ensure Go 1.23.0+ is installed
go version

# Clone and setup
git clone <repository-url>
cd Golid
```

### Building the Development Server

```bash
# Build the development CLI tool
cd cmd/devserver
go build -o ../../golid-dev .
cd ../..
```

### Running the Development Server

```bash
# Start development server with hot reload
./golid-dev

# Server will be available at http://localhost:8090
# Automatically rebuilds WASM when Go files change
# Serves SPA routes correctly
```

### Manual WASM Compilation

```bash
# Compile main.go to WASM
GOOS=js GOARCH=wasm go build -o main.wasm ./main.go

# The development server handles this automatically
```

### Project Structure

```
Golid/
├── main.go              # Main application entry point
├── golid/              # Core framework package
│   └── golid.go        # Signals, components, router, DOM bindings
├── cmd/devserver/      # Development server
│   ├── main.go         # Server implementation with file watching
│   ├── go.mod          # Separate module for dev dependencies
│   └── index.html      # HTML template
├── devenv.nix          # Development environment config
├── devenv.yaml         # devenv configuration
└── go.mod              # Main project dependencies
```

## Testing Information

### Testing Challenges with WASM

**Important:** Golid is designed for WebAssembly environments and uses `syscall/js` for DOM manipulation. This creates unique testing challenges:

1. **Standard `go test` limitations:** Tests fail on native platforms due to `syscall/js` import constraints
2. **WASM test execution:** Running tests with `GOOS=js GOARCH=wasm` has environment limitations
3. **DOM dependency:** Most reactive features require a browser DOM environment

### Testing Strategies

#### 1. Unit Testing Non-DOM Logic

Focus on testing pure Go logic that doesn't depend on DOM APIs:

```go
// Test Signal operations (basic functionality)
func TestSignalOperations(t *testing.T) {
    signal := golid.NewSignal(42)
    if signal.Get() != 42 {
        t.Errorf("Expected 42, got %d", signal.Get())
    }
    
    signal.Set(100)
    if signal.Get() != 100 {
        t.Errorf("Expected 100, got %d", signal.Get())
    }
}
```

#### 2. Component Structure Testing

Test component creation and HTML generation:

```go
func TestComponentStructure(t *testing.T) {
    component := MyComponent()
    
    if component == nil {
        t.Error("Component should not be nil")
    }
    
    // Test HTML generation (works without DOM)
    html := golid.RenderHTML(component)
    if !strings.Contains(html, "expected-content") {
        t.Error("Component should contain expected content")
    }
}
```

#### 3. Integration Testing with Browser

For full testing, use browser-based approaches:

1. **Manual browser testing** using the development server
2. **E2E testing tools** like Playwright or Selenium
3. **WASM test runners** in headless browsers

#### 4. Example Test File

```go
package main

import (
    "app/golid"
    "fmt"
    "strings"
    "testing"
    
    . "maragu.dev/gomponents"
    . "maragu.dev/gomponents/html"
)

// Test Signal with different types
func TestSignalTypes(t *testing.T) {
    // String signals
    strSignal := golid.NewSignal("hello")
    strSignal.Set("world")
    if strSignal.Get() != "world" {
        t.Errorf("Expected 'world', got %s", strSignal.Get())
    }
    
    // Boolean signals
    boolSignal := golid.NewSignal(true)
    boolSignal.Set(false)
    if boolSignal.Get() {
        t.Error("Expected false")
    }
}

// Test component creation
func TestComponentCreation(t *testing.T) {
    component := Div(Text("test"))
    html := golid.RenderHTML(component)
    
    if !strings.Contains(html, "test") {
        t.Error("Component should render text content")
    }
}
```

### Running Tests

Due to WASM constraints, testing approaches:

```bash
# This will fail due to syscall/js constraints
go test ./...

# Alternative: Test specific non-DOM functions
# Create separate test files for non-WASM logic

# Integration testing: Use the development server
./golid-dev
# Then test manually in browser at http://localhost:8090
```

## Development Best Practices

### Code Style Guidelines

#### 1. Signal Usage

```go
// Good: Descriptive signal names
userCount := golid.NewSignal(0)
isLoggedIn := golid.NewSignal(false)

// Good: Initialize signals with appropriate types
userName := golid.NewSignal("")  // string
userAge := golid.NewSignal(0)    // int

// Good: Use signals in reactive contexts
golid.BindText(func() string {
    return fmt.Sprintf("Users: %d", userCount.Get())
})
```

#### 2. Component Structure

```go
// Good: Components return gomponents.Node
func UserProfile(userID string) Node {
    user := golid.NewSignal(User{})
    
    return Div(
        Class("user-profile"),
        H1(golid.BindText(func() string {
            return user.Get().Name
        })),
        Button(
            Text("Refresh"),
            golid.OnClick(func() {
                // Update user data
                fetchUser(userID, user)
            }),
        ),
    )
}

// Good: Separate data fetching logic
func fetchUser(id string, signal *golid.Signal[User]) {
    // API call or data fetching logic
    // signal.Set(fetchedUser)
}
```

#### 3. Router Configuration

```go
// Good: Organize routes logically
func setupRoutes(router *golid.Router) {
    // Static routes first
    router.AddRoute("/", HomePage)
    router.AddRoute("/about", AboutPage)
    
    // Parameterized routes
    router.AddRoute("/user/:id", UserPage)
    router.AddRoute("/posts/:category/:slug", PostPage)
    
    // Protected routes with guards
    router.AddRouteWithGuard("/admin", AdminPage, isAdminGuard)
}
```

#### 4. Event Handling

```go
// Good: Clear, specific event handlers
Button(
    Text("Save"),
    golid.OnClick(func() {
        if validateForm() {
            saveData()
        }
    }),
)

// Good: Extract complex logic to separate functions
func handleFormSubmit(form *FormData) func() {
    return func() {
        if err := validateForm(form); err != nil {
            showError(err)
            return
        }
        
        if err := submitForm(form); err != nil {
            showError(err)
            return
        }
        
        showSuccess()
        resetForm(form)
    }
}
```

### Development Workflow

#### 1. Standard Development

```bash
# Start development server
./golid-dev

# Edit Go files - server auto-rebuilds WASM
# Refresh browser to see changes
# Check browser console for errors
```

#### 2. Debugging

```bash
# Check server logs
tail -f cmd/devserver/server.log

# Browser DevTools:
# - Console for Go panic messages
# - Network tab for WASM loading issues
# - Elements tab for DOM inspection
```

#### 3. Adding New Components

```go
// 1. Create component function
func NewComponent() Node {
    signal := golid.NewSignal(initialValue)
    
    return Div(
        // Component structure
    )
}

// 2. Add to router if needed
router.AddRoute("/new-component", func(params golid.RouteParams) Node {
    return NewComponent()
})

// 3. Test in browser
// 4. Add to navigation if needed
```

### Common Patterns

#### 1. Form Handling

```go
func ContactForm() Node {
    name := golid.NewSignal("")
    email := golid.NewSignal("")
    message := golid.NewSignal("")
    
    return Form(
        Div(
            Label(Text("Name:")),
            Input(
                Type("text"),
                golid.BindInput(name),
            ),
        ),
        Div(
            Label(Text("Email:")),
            Input(
                Type("email"),
                golid.BindInput(email),
            ),
        ),
        Div(
            Label(Text("Message:")),
            Textarea(
                golid.BindTextarea(message),
            ),
        ),
        Button(
            Type("submit"),
            Text("Send"),
            golid.OnClick(func() {
                submitForm(name.Get(), email.Get(), message.Get())
            }),
        ),
    )
}
```

#### 2. Conditional Rendering

```go
func ConditionalComponent() Node {
    showDetails := golid.NewSignal(false)
    
    return Div(
        Button(
            Text("Toggle Details"),
            golid.OnClick(func() {
                showDetails.Set(!showDetails.Get())
            }),
        ),
        golid.Bind(func() Node {
            if showDetails.Get() {
                return Div(
                    Class("details"),
                    Text("Detailed information here"),
                )
            }
            return Text("")
        }),
    )
}
```

#### 3. List Rendering

```go
func TodoList() Node {
    todos := golid.NewSignal([]Todo{})
    
    return Div(
        H2(Text("Todo List")),
        golid.Bind(func() Node {
            items := []Node{}
            for _, todo := range todos.Get() {
                items = append(items, renderTodoItem(todo))
            }
            return Ul(items...)
        }),
    )
}
```

### Performance Considerations

1. **Signal Updates:** Minimize unnecessary signal updates to reduce reactive re-renders
2. **Component Nesting:** Avoid deeply nested reactive components
3. **Memory Management:** Be aware that WASM memory management differs from native Go
4. **Bundle Size:** Consider TinyGo for production builds to reduce WASM size

### Troubleshooting

#### Common Issues

1. **"Build failed" errors:**
   - Check Go syntax errors
   - Verify import paths
   - Ensure WASM-compatible code

2. **Router not working:**
   - Verify `golid.SetGlobalRouter()` called
   - Check route patterns match URLs
   - Ensure `RouterOutlet()` in component tree

3. **Signals not updating UI:**
   - Verify `golid.Bind()` or `golid.BindText()` usage
   - Check that signal `.Set()` is called
   - Ensure reactive context is properly established

4. **Development server issues:**
   - Check if port 8090 is available
   - Verify `wasm_exec.js` is present
   - Check file permissions on `golid-dev`

---

**Last Updated:** 2025-08-18

This document provides project-specific guidelines for advanced developers working with the Golid framework. For basic Go development practices, refer to standard Go documentation.