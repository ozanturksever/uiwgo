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
2. **Optional: TinyGo** - For smaller WASM bundles (recommended for production)
3. **Optional: devenv** - For reproducible development environment

### Quick Start (Recommended)

The easiest way to get started with Golid development:

```bash
# Clone and setup (one-time)
git clone <repository-url>
cd Golid

# Setup development environment (installs deps, builds devserver)
make setup

# Start development server with hot reload
make dev
```

Server will be available at **http://localhost:8090** with automatic WASM rebuilding on file changes.

### Using the Makefile

Golid includes a comprehensive Makefile for streamlined development. View all available commands:

```bash
make help
```

#### Essential Commands

```bash
# Setup everything for first-time development
make setup              # Install dependencies and build devserver

# Development workflow
make dev               # Start development server (with hot reload)
make run               # Alias for 'make dev'

# Building
make build             # Build devserver and WASM
make devserver         # Build only the development server
make wasm              # Compile only WebAssembly
make wasm-tiny         # Compile WASM with TinyGo (smaller bundle)

# Maintenance
make clean             # Remove build artifacts
make rebuild           # Clean and rebuild everything
make deps              # Update dependencies
make status            # Check project build status
```

#### Advanced Commands

```bash
make test              # Run tests (uses wasmbrowsertest)
make info              # Show build information
make check-deps        # Verify required tools are available
```

### Alternative Setup Methods

#### Method 1: Using devenv (Advanced Users)
```bash
# If you have devenv installed
devenv shell
# Then run: make setup
```

#### Method 2: Manual Setup (Not Recommended)
```bash
# Ensure Go 1.23.0+ is installed
go version

# Install dependencies manually
go mod tidy
cd cmd/devserver && go mod tidy && cd ../..

# Build development server manually
cd cmd/devserver
go build -ldflags="-s -w" -o ../../golid-dev .
cd ../..

# Start development server
./golid-dev
```

### Manual WASM Compilation

While the development server handles WASM compilation automatically, you can compile manually:

```bash
# Standard Go compilation
GOOS=js GOARCH=wasm go build -ldflags="-s -w" -o main.wasm ./main.go

# Or use the Makefile
make wasm

# For production (smaller bundle with TinyGo)
make wasm-tiny
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

**Test-Driven Development (TDD) is strongly recommended for Golid development.** Always write tests before implementing new features to ensure robust and reliable code.

### Testing Infrastructure (Updated)

**Great News:** Golid now has comprehensive testing infrastructure with **wasmbrowsertest** support that resolves previous WASM testing challenges!

- ✅ **Working WASM Tests:** Tests run in actual browser environment using wasmbrowsertest
- ✅ **Comprehensive Test Suite:** 12+ test functions covering Bind functionality and DOM utilities
- ✅ **Proper CI/CD Support:** Tests can be automated in development workflow
- ✅ **Detailed Documentation:** See `golid/TESTING_README.md` for complete testing guide

### Test-First Development Workflow

**Always follow this workflow for new features:**

1. **Write tests first** - Define expected behavior through tests
2. **Run tests** - Verify they fail initially (red phase)
3. **Implement code** - Write minimal code to make tests pass (green phase)
4. **Refactor** - Improve code while keeping tests passing
5. **Ensure all tests pass** - Never commit code with failing tests

### Testing Challenges with WASM (Legacy Information)

While we now have working test infrastructure, understanding WASM constraints helps with test design:

1. **DOM-dependent vs DOM-independent:** Structure tests work everywhere, reactive behavior needs browser
2. **Test environment:** Use `wasmbrowsertest` for full WASM functionality testing
3. **Native vs WASM:** Some tests require browser environment for `syscall/js` APIs

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

#### ✅ WASM Tests (Recommended - Now Working!)

```bash
# Run all tests with WASM/browser environment
cd golid && env -i PATH="$PATH" HOME="$HOME" GOOS=js GOARCH=wasm go test -v

# Run specific test
cd golid && env -i PATH="$PATH" HOME="$HOME" GOOS=js GOARCH=wasm go test -v -run TestBindStructureGeneration

# Add to Makefile for convenience
make test              # Run tests (uses wasmbrowsertest)
```

**Important:** Use `env -i` to clear environment variables to avoid command line limit errors in the browser environment.

#### ❌ Standard Go Tests (Expected to Fail)

```bash
# This will still fail due to syscall/js constraints in native environment
go test ./golid -v
```

#### Integration Testing

```bash
# Use the development server for manual testing
make dev
# Then test manually in browser at http://localhost:8090
```

**For detailed testing information, see `golid/TESTING_README.md`**

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
make dev

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

#### 3. Adding New Components (Test-First Approach)

```go
// 1. WRITE TESTS FIRST - Define expected behavior
func TestNewComponent(t *testing.T) {
    component := NewComponent()
    
    if component == nil {
        t.Error("Component should not be nil")
    }
    
    html := golid.RenderHTML(component)
    if !strings.Contains(html, "expected-content") {
        t.Error("Component should contain expected content")
    }
}

// 2. RUN TESTS - Verify they fail initially
// cd golid && env -i PATH="$PATH" HOME="$HOME" GOOS=js GOARCH=wasm go test -v

// 3. IMPLEMENT - Create component function
func NewComponent() Node {
    signal := golid.NewSignal(initialValue)
    
    return Div(
        // Component structure
    )
}

// 4. RUN TESTS AGAIN - Ensure they pass
// 5. Add to router if needed
router.AddRoute("/new-component", func(params golid.RouteParams) Node {
    return NewComponent()
})

// 6. Test in browser for integration
// 7. Add to navigation if needed
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