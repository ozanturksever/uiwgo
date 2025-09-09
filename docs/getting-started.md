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

Let's look at the counter example to understand the basic structure:

```go
// examples/counter/main.go
package main

import (
    "github.com/ozanturksever/uiwgo/comps"
    "github.com/ozanturksever/uiwgo/reactivity"
    g "maragu.dev/gomponents"
    h "maragu.dev/gomponents/html"
)

type Counter struct {
    count *reactivity.Signal[int]
}

func NewCounter() *Counter {
    return &Counter{
        count: reactivity.NewSignal(0),
    }
}

// Render returns the gomponents Node structure
func (c *Counter) Render() g.Node {
    return h.Div(g.Class("counter"),
        h.H1(g.Text("Counter Example")),
        h.Div(g.Class("controls"),
            h.Button(g.Attr("data-click", "decrement"), g.Text("-")),
            h.Span(
                g.Attr("data-text", "count"),
                g.Class("count"),
                g.Text("0"),
            ),
            h.Button(g.Attr("data-click", "increment"), g.Text("+")),
        ),
    )
}

// Attach wires up the reactive behavior
func (c *Counter) Attach() {
    // Bind the count signal to the text content
    c.BindText("count", c.count)
    
    // Bind click handlers
    c.BindClick("increment", func() {
        c.count.Set(c.count.Get() + 1)
    })
    
    c.BindClick("decrement", func() {
        c.count.Set(c.count.Get() - 1)
    })
}

func main() {
    counter := NewCounter()
    
    // Mount the component to the DOM
    comps.Mount("app", counter)
}
```

## Understanding the Flow

### 1. Gomponents-Based Rendering

UIwGo follows a **gomponents-first** approach for type-safe HTML generation:

1. **Render** - Generate gomponents Node with data attributes as markers
2. **Attach** - Scan the DOM and bind reactive behavior to marked elements
3. **Update** - Fine-grained updates when signals change

```go
// Step 1: Render gomponents Node with markers
func (c *Counter) Render() g.Node {
    return h.Span(
        g.Attr("data-text", "count"),
        g.Text("0"), // Initial value
    )
}

// Step 2: Attach reactive behavior
func (c *Counter) Attach() {
    c.BindText("count", c.count) // Bind signal to marker
}

// Step 3: Updates happen automatically
c.count.Set(42) // DOM automatically updates to show "42"
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

### 3. Data Attributes as Binding Markers

UIwGo uses data attributes to mark elements for binding:

| Attribute | Purpose | Example |
|-----------|---------|----------|
| `data-text="key"` | Text content binding | `<span data-text="count">0</span>` |
| `data-click="key"` | Click event binding | `<button data-click="increment">+</button>` |
| `data-show="key"` | Conditional visibility | `<div data-show="isVisible">Content</div>` |
| `data-for="key"` | List rendering | `<ul data-for="items">...</ul>` |

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
    "github.com/ozanturksever/uiwgo/reactivity"
    g "maragu.dev/gomponents"
    h "maragu.dev/gomponents/html"
)

type HelloSignals struct {
    name     *reactivity.Signal[string]
    greeting *reactivity.Memo[string]
}

func NewHelloSignals() *HelloSignals {
    h := &HelloSignals{
        name: reactivity.NewSignal("World"),
    }
    
    // Memo automatically recomputes when name changes
    h.greeting = reactivity.NewMemo(func() string {
        return fmt.Sprintf("Hello, %s!", h.name.Get())
    })
    
    return h
}

func (h *HelloSignals) Render() g.Node {
    return h.Div(g.Class("hello-signals"),
        h.H1(
            g.Attr("data-text", "greeting"),
            g.Text("Hello, World!"),
        ),
        h.Div(g.Class("input-group"),
            h.Label(
                g.Attr("for", "name-input"),
                g.Text("Enter your name:"),
            ),
            h.Input(
                g.Attr("id", "name-input"),
                g.Attr("type", "text"),
                g.Attr("data-input", "name"),
                g.Attr("placeholder", "Your name"),
                g.Attr("value", "World"),
            ),
        ),
        h.Button(
            g.Attr("data-click", "reset"),
            g.Text("Reset"),
        ),
    )
}

func (h *HelloSignals) Attach() {
    // Bind the computed greeting to the heading
    h.BindText("greeting", h.greeting)
    
    // Bind the input to the name signal (two-way binding)
    h.BindInput("name", h.name)
    
    // Bind reset button
    h.BindClick("reset", func() {
        h.name.Set("World")
    })
}

func main() {
    hello := NewHelloSignals()
    comps.Mount("app", hello)
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
// HTML
`<input data-input="name" value="World" />`

// Go
h.BindInput("name", h.name) // Input updates signal, signal updates input
```

### Event Handling
```go
// HTML
`<button data-click="reset">Reset</button>`

// Go
h.BindClick("reset", func() {
    h.name.Set("World")
})
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