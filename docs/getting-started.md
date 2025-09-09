# Getting Started

Get up and running with UIwGo in minutes. This guide will walk you through setting up your development environment and creating your first interactive component.

## Prerequisites

### Required
- **Go 1.21+** with WebAssembly support
- **Modern web browser** (Chrome, Firefox, Safari, Edge)
- **HTTP server** (development server included)

### Recommended
- **Make** for build automation
- **Git** for version control
- **VS Code** with Go extension for development

### Verify WebAssembly Support

```bash
# Check Go version
go version

# Verify WebAssembly target support
GOOS=js GOARCH=wasm go version
```

## Quick Start

### 1. Clone and Setup

```bash
# Clone the repository
git clone <repository-url>
cd uiwgo

# Install dependencies (if using npm for React compatibility)
npm install
```

### 2. Run Your First Example

```bash
# Start the development server with the counter example
make run counter

# Or specify explicitly
make run EX=counter
```

Open your browser to `http://localhost:8080` and you should see a working counter with increment/decrement buttons.

### 3. Explore the Code

Let's look at the counter example to understand the basic functional structure:

```go
// examples/counter/main.go
package main

import (
	"fmt"
	"github.com/ozanturksever/uiwgo/comps"
	"github.com/ozanturksever/uiwgo/dom"
	"github.com/ozanturksever/uiwgo/reactivity"
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

// A component is a function that returns a gomponents.Node
func Counter() Node {
	// Create a reactive signal to hold the count.
	count := reactivity.NewSignal(0)

	// Use OnMount to set up event listeners after the DOM is created.
	comps.OnMount(func() {
		// Get elements by ID.
		incrementBtn := dom.GetElementByID("increment-btn")
		decrementBtn := dom.GetElementByID("decrement-btn")

		// Bind click handlers directly to the DOM elements.
		dom.BindClickToCallback(incrementBtn, func() {
			count.Set(count.Get() + 1
		})
		dom.BindClickToCallback(decrementBtn, func() {
			count.Set(count.Get() - 1
		})
	})

	// Return the UI structure using gomponents.
	return Div(
		H1(Text("Counter Example")),
		Div(Class("controls"),
			Button(ID("decrement-btn"), Text("-")),
			Span(
				Class("count"),
				// Bind the text content directly to the count signal.
				comps.BindText(func() string {
					return fmt.Sprintf("%d", count.Get())
				}),
			),
			Button(ID("increment-btn"), Text("+")),
		),
	)
}

func main() {
	// Mount the component function to the DOM element with the ID "app".
	comps.Mount("app", Counter)

	// Prevent the Go program from exiting.
	select {}
}
```

## Understanding the Flow

### 1. Functional Components with gomponents

UIwGo uses a **functional** approach where components are simply Go functions that return a `gomponents.Node`.

1.  **Define State**: Reactive state is created using signals (`reactivity.NewSignal`).
2.  **Define View**: The UI structure is defined using `gomponents` functions. Reactive content is embedded directly using helpers like `comps.BindText`.
3.  **Define Behavior**: Side effects, like event handling, are set up within lifecycle hooks like `comps.OnMount`. This ensures the DOM elements exist before you try to access them.
4.  **Mount**: The top-level component function is mounted to a DOM element, which renders the UI and activates the reactive bindings.

```go
// A component is just a function.
func MyComponent() g.Node {
    // 1. State
    message := reactivity.NewSignal("Hello!")

    // 2. Behavior (runs after render)
    comps.OnMount(func() {
        // ... event binding ...
    })

    // 3. View
    return h.Div(
        comps.BindText(message.Get), // Reactive binding
    )
}
```

### 2. Signals and Reactivity

**Signals** are the foundation of UIwGo's reactivity:

```go
// Create a signal
count := reactivity.NewSignal(0)

// Read the value
value := count.Get()

// Update the value (triggers reactive updates)
count.Set(value + 1)

// Signals automatically track dependencies
effect := reactivity.NewEffect(func() {
    fmt.Printf("Count is: %d\n", count.Get())
})
// Effect runs immediately and re-runs when count changes
```

### 3. Direct DOM Interaction

Instead of using abstract binding markers, you interact with the DOM directly when needed, primarily for attaching event listeners.

-   Give elements an `ID` in your `gomponents` structure.
-   Use `dom.GetElementByID` within `comps.OnMount` to get a reference to the element.
-   Use `dom.BindClickToCallback` and other `dom.Bind*` functions to attach handlers.

## Creating Your First Component

Let's create a simple "Hello, Signals" component from scratch:

### 1. Create the Project Structure

```bash
# Create a new example directory
mkdir -p examples/hello_signals
cd examples/hello_signals
```

### 2. Write the Component

```go
// examples/hello_signals/main.go
package main

import (
    "fmt"
    "github.com/ozanturksever/uiwgo/comps"
    "github.com/ozanturksever/uiwgo/dom"
    "github.com/ozanturksever/uiwgo/reactivity"
    g "maragu.dev/gomponents"
    h "maragu.dev/gomponents/html"
)

func HelloSignals() g.Node {
    // Signals for state
    name := reactivity.NewSignal("World")

    // Memo for computed state
    greeting := reactivity.NewMemo(func() string {
        return fmt.Sprintf("Hello, %s!", name.Get())
    })

    // Lifecycle hook for event binding
    comps.OnMount(func() {
        nameInput := dom.GetElementByID("name-input")
        resetBtn := dom.GetElementByID("reset-btn")

        // Two-way binding for the input field
        dom.BindValue(nameInput, name)

        // Click handler for the reset button
        dom.BindClickToCallback(resetBtn, func() {
            name.Set("World")
        })
    })

    // Return the UI tree
    return h.Div(g.Class("hello-signals"),
        h.H1(
            // Bind the computed greeting to the heading's text
            comps.BindText(greeting.Get),
        ),
        h.Div(g.Class("input-group"),
            h.Label(For("name-input"), g.Text("Enter your name:")),
            h.Input(
                ID("name-input"),
                Type("text"),
                Placeholder("Your name"),
            ),
        ),
        h.Button(
            ID("reset-btn"),
            g.Text("Reset"),
        ),
    )
}

func main() {
    comps.Mount("app", HelloSignals)
    select {}
}
```

### 3. Create the HTML File

```html
<!-- examples/hello_signals/index.html -->
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Hello Signals - UIwGo</title>
    <style>
        body {
            font-family: system-ui, sans-serif;
            max-width: 600px;
            margin: 2rem auto;
            padding: 1rem;
        }
        .hello-signals {
            text-align: center;
        }
        .input-group {
            margin: 1rem 0;
        }
        input {
            padding: 0.5rem;
            margin: 0.5rem;
            border: 1px solid #ccc;
            border-radius: 4px;
        }
        button {
            padding: 0.5rem 1rem;
            background: #007bff;
            color: white;
            border: none;
            border-radius: 4px;
            cursor: pointer;
        }
        button:hover {
            background: #0056b3;
        }
    </style>
</head>
<body>
    <div id="app"></div>
    <script src="wasm_exec.js"></script>
    <script>
        const go = new Go();
        WebAssembly.instantiateStreaming(fetch("main.wasm"), go.importObject)
            .then((result) => {
                go.run(result.instance);
            });
    </script>
</body>
</html>
```

### 4. Run Your Component

```bash
# From the project root
make run hello_signals

# Open http://localhost:8080
```

You should see:
- A heading that says "Hello, World!"
- An input field
- A reset button
- The heading updates automatically as you type in the input

## Key Concepts Demonstrated

### Signals
```go
name := reactivity.NewSignal("World") // Reactive state
name.Set("Alice")                     // Update triggers reactivity
value := name.Get()                   // Read current value
```

### Memos (Computed Values)
```go
greeting := reactivity.NewMemo(func() string {
    return fmt.Sprintf("Hello, %s!", name.Get())
})
// Automatically recomputes when name changes
```

### Two-Way Data Binding
```go
// In component function
name := reactivity.NewSignal("")
comps.OnMount(func() {
    nameInput := dom.GetElementByID("name-input")
    dom.BindValue(nameInput, name) // Binds signal to input value
})

// In gomponents tree
Input(ID("name-input"))
```

### Event Handling
```go
// In component function
comps.OnMount(func() {
    resetBtn := dom.GetElementByID("reset-btn")
    dom.BindClickToCallback(resetBtn, func() {
        name.Set("World")
    })
})

// In gomponents tree
Button(ID("reset-btn"), Text("Reset"))
```

## Development Workflow

### Available Commands

```bash
# Development
make run <example>          # Start dev server with hot reload
make build <example>        # Build to WebAssembly
make clean                  # Clean built artifacts

# Testing
make test                   # Run unit tests
make test-example <example> # Run browser tests for example
make test-examples          # Run all example tests
make test-all              # Run everything

# Utilities
make kill                   # Free port 8080
```

### Hot Reload

The development server automatically:
- Compiles your Go code to WebAssembly
- Serves the required `wasm_exec.js` file
- Reloads the browser when you save changes

### Project Structure

```
examples/
├── counter/           # Basic counter example
├── todo/             # Todo list with persistence
├── router_demo/      # Client-side routing
├── hello_signals/    # Your new component
└── ...

docs/                 # Documentation
src/                  # Core UIwGo source
├── reactivity/       # Signals, effects, memos
├── comps/           # Component system
├── router/          # Client-side routing
└── dom/             # DOM integration
```

## Next Steps

### Learn Core Concepts
1. **[Concepts](./concepts.md)** - Understand the mental model
2. **[Reactivity & State](./guides/reactivity-state.md)** - Master signals and effects
3. **[Control Flow](./guides/control-flow.md)** - Conditional rendering and lists

### Build Something Real
1. **[Forms & Events](./guides/forms-events.md)** - Handle user input
2. **[Lifecycle & Effects](./guides/lifecycle-effects.md)** - Component lifecycle management
3. **[Application Manager](./guides/application-manager.md)** - Application lifecycle and management

## Troubleshooting

### Common Issues

**Empty page or "loading" forever**
- Ensure you're serving over HTTP, not `file://`
- Check browser console for WebAssembly errors
- Verify Go WebAssembly support: `GOOS=js GOARCH=wasm go version`

**Changes not reflecting**
- Make sure the dev server is running with hot reload
- Check for Go compilation errors in the terminal
- Hard refresh the browser (Ctrl+F5 / Cmd+Shift+R)

**Port 8080 in use**
```bash
make kill  # Free the port
```

**Build errors**
- Ensure you're in the project root directory
- Check that all imports are correct
- Verify the example directory structure

For more help, see [Troubleshooting](./troubleshooting.md).

---

**Congratulations!** You've created your first UIwGo component. Ready to dive deeper? Continue with [Concepts](./concepts.md) to understand the core mental models.