# gomponents: HTML Components in Pure Go - LLM Documentation

## Overview

gomponents is a Go library that enables building HTML components using pure Go code. Instead of using traditional HTML templates, developers write HTML as Go functions that compile to type-safe, performant HTML5 output. This approach leverages Go's type system, IDE support, and debugging capabilities while avoiding template language complexity.

## Core Concepts

### Node Interface
The fundamental building block is the `Node` interface:
```go
type Node interface {
    Render(w io.Writer) error
}
```
Everything in gomponents implements this interface - elements, attributes, text, and components.

### Node Types
- **ElementType**: Regular HTML elements (div, span, etc.) and text nodes
- **AttributeType**: HTML attributes (class, href, etc.)

The library automatically handles proper placement during rendering.

## Installation

```bash
go get maragu.dev/gomponents
```

## Import Patterns

### Dot imports (recommended):
Contrary to common idiomatic Go, dot imports are the recommended approach for gomponents as they make the code read like a DSL for HTML:
```go
import (
    . "maragu.dev/gomponents"
    . "maragu.dev/gomponents/html"
    . "maragu.dev/gomponents/components"
)
```

### Standard imports with aliases (alternative):
For those who prefer avoiding dot imports, use single-letter aliases:
```go
import (
    g "maragu.dev/gomponents"
    h "maragu.dev/gomponents/html"
    c "maragu.dev/gomponents/components"
    ghttp "maragu.dev/gomponents/http"
)
```

## Package Structure

### maragu.dev/gomponents (core)
Core interfaces and helper functions:
- `Node` interface
- `El(name string, children ...Node)` - create custom elements
- `Attr(name string, value ...string)` - create custom attributes
- `Text(string)` - HTML-escaped text
- `Textf(format string, args...)` - formatted escaped text
- `Raw(string)` - unescaped HTML
- `Rawf(format string, args...)` - formatted unescaped HTML
- `Group([]Node)` - group multiple nodes
- `Map[T]([]T, func(T) Node)` - transform slices to nodes
- `If(condition bool, node Node)` - conditional rendering
- `Iff(condition bool, func() Node)` - lazy conditional rendering

### maragu.dev/gomponents/html
All HTML5 elements and attributes as Go functions:
- Elements: `Div()`, `Span()`, `A()`, `H1()`, etc.
- Attributes: `Class()`, `ID()`, `Href()`, `Style()`, etc.
- Special: `Doctype()` for HTML5 doctype declaration

### maragu.dev/gomponents/components
Higher-level components:
- `HTML5(HTML5Props)` - complete HTML5 document structure
- `Classes` - dynamic class management map

### maragu.dev/gomponents/http
HTTP handler integration:
- `Handler` type - returns (Node, error)
- `Adapt()` - converts Handler to http.HandlerFunc

## Basic Usage Examples

### Simple Element
```go
import (
    . "maragu.dev/gomponents"
    . "maragu.dev/gomponents/html"
)

// <div class="container">Hello, World!</div>
Div(Class("container"), Text("Hello, World!"))
```

### Nested Structure
```go
import (
    . "maragu.dev/gomponents"
    . "maragu.dev/gomponents/html"
)

// <nav><a href="/">Home</a><a href="/about">About</a></nav>
Nav(
    A(Href("/"), Text("Home")),
    A(Href("/about"), Text("About"))
)
```

### Complete Page
```go
import (
    . "maragu.dev/gomponents"
    . "maragu.dev/gomponents/components"
    . "maragu.dev/gomponents/html"
)

func Page() Node {
    return HTML5(HTML5Props{
        Title: "My Page",
        Language: "en",
        Head: []Node{
            Meta(Name("author"), Content("John Doe")),
        },
        Body: []Node{
            H1(Text("Welcome")),
            P(Text("This is my page")),
        },
    })
}
```

## Advanced Patterns

### Component Functions
Create reusable components as functions:
```go
import (
    . "maragu.dev/gomponents"
    . "maragu.dev/gomponents/html"
)

func Card(title, content string) Node {
    return Div(Class("card"),
        H2(Class("card-title"), Text(title)),
        P(Class("card-content"), Text(content)),
    )
}
```

### Dynamic Rendering
```go
import (
    . "maragu.dev/gomponents"
    . "maragu.dev/gomponents/html"
)

func UserList(users []User) Node {
    return Ul(
        Map(users, func(u User) Node {
            return Li(Text(u.Name))
        }),
    )
}
```

### Conditional Rendering
```go
import (
    . "maragu.dev/gomponents"
    . "maragu.dev/gomponents/html"
)

func NavBar(isLoggedIn bool, username string) Node {
    return Nav(
        A(Href("/"), Text("Home")),
        If(isLoggedIn, 
            Span(Text("Welcome, " + username))),
        If(!isLoggedIn,
            A(Href("/login"), Text("Login"))),
    )
}
```

### Dynamic Classes
```go
import (
    . "maragu.dev/gomponents"
    . "maragu.dev/gomponents/components"
    . "maragu.dev/gomponents/html"
)

Div(
    Classes{
        "active": isActive,
        "disabled": isDisabled,
        "primary": isPrimary,
    },
    Text("Dynamic styling"),
)
```

## Special Elements and Attributes

### Name Conflicts
Some HTML names conflict in Go. The library provides both variants:
- `Style()` (attribute) vs `StyleEl()` (element)
- `Title()` (attribute) vs `TitleEl()` (element)
- `Form()` (element) vs `FormAttr()` (attribute)
- `Label()` (element) vs `LabelAttr()` (attribute)
- `Data()` (attribute) vs `DataEl()` (element)
- `Cite()` (element) vs `CiteAttr()` (attribute)

### Void Elements
Self-closing elements (br, img, input, etc.) are handled automatically. Child nodes that aren't attributes are ignored:
```go
// Correct: <img src="pic.jpg" alt="Picture">
Img(Src("pic.jpg"), Alt("Picture"))

// Text("ignored") won't render for void elements
Img(Src("pic.jpg"), Text("ignored"))
```

## HTTP Integration

### Basic Handler
```go
import (
    "net/http"
    
    . "maragu.dev/gomponents"
    . "maragu.dev/gomponents/html"
    ghttp "maragu.dev/gomponents/http"
)

func HomeHandler(w http.ResponseWriter, r *http.Request) (Node, error) {
    return Page("Welcome!"), nil
}

// In main:
http.HandleFunc("/", ghttp.Adapt(HomeHandler))
```

### Error Handling
```go
import (
    "net/http"
    
    . "maragu.dev/gomponents"
    ghttp "maragu.dev/gomponents/http"
)

type HTTPError struct {
    Code int
    Message string
}

func (e HTTPError) Error() string { return e.Message }
func (e HTTPError) StatusCode() int { return e.Code }

func Handler(w http.ResponseWriter, r *http.Request) (Node, error) {
    if unauthorized {
        return ErrorPage(), HTTPError{Code: 401, Message: "Unauthorized"}
    }
    return SuccessPage(), nil
}
```

## Best Practices

### 1. Component Composition
Build complex UIs from simple, reusable components:
```go
import (
    . "maragu.dev/gomponents"
    . "maragu.dev/gomponents/components"
    . "maragu.dev/gomponents/html"
)

func Layout(title string, content Node) Node {
    return HTML5(HTML5Props{
        Title: title,
        Body: []Node{
            Header(),
            Main(content),
            Footer(),
        },
    })
}
```

### 2. Type Safety
Leverage Go's type system for compile-time guarantees:
```go
import (
    . "maragu.dev/gomponents"
    . "maragu.dev/gomponents/html"
)

type ButtonVariant string

const (
    ButtonPrimary   ButtonVariant = "btn-primary"
    ButtonSecondary ButtonVariant = "btn-secondary"
)

func Button(variant ButtonVariant, text string) Node {
    return Button(Class(string(variant)), Type("button"), Text(text))
}
```

### 3. Performance
Nodes render directly to io.Writer for efficiency:
```go
import (
    "net/http"
)

// Efficient - streams directly to response
func (h Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    node := BuildPage()
    node.Render(w)
}
```

### 4. Testing
Components are pure functions, making testing straightforward:
```go
import (
    "bytes"
    "testing"
)

func TestButton(t *testing.T) {
    btn := Button("Click me")
    
    var buf bytes.Buffer
    btn.Render(&buf)
    
    expected := `<button>Click me</button>`
    if buf.String() != expected {
        t.Errorf("got %q, want %q", buf.String(), expected)
    }
}
```

## Common Patterns

### Forms
```go
import (
    . "maragu.dev/gomponents"
    . "maragu.dev/gomponents/html"
)

func LoginForm() Node {
    return Form(Method("post"), Action("/login"),
        Label(For("email"), Text("Email:")),
        Input(Type("email"), ID("email"), Name("email"), Required()),
        
        Label(For("password"), Text("Password:")),
        Input(Type("password"), ID("password"), Name("password"), Required()),
        
        Button(Type("submit"), Text("Login")),
    )
}
```

### Tables
```go
import (
    . "maragu.dev/gomponents"
    . "maragu.dev/gomponents/html"
)

func DataTable(headers []string, rows [][]string) Node {
    return Table(
        Thead(
            Tr(Map(headers, func(h string) Node {
                return Th(Text(h))
            })),
        ),
        Tbody(
            Map(rows, func(row []string) Node {
                return Tr(Map(row, func(cell string) Node {
                    return Td(Text(cell))
                }))
            }),
        ),
    )
}
```

### Lists
```go
import (
    . "maragu.dev/gomponents"
    . "maragu.dev/gomponents/html"
)

func NavMenu(items []MenuItem) Node {
    return Nav(
        Ul(Class("nav-menu"),
            Map(items, func(item MenuItem) Node {
                return Li(
                    A(Href(item.URL), Text(item.Label)),
                )
            }),
        ),
    )
}
```

## Integration Tips

### With CSS Frameworks
Works seamlessly with Tailwind, Bootstrap, etc.:
```go
import (
    . "maragu.dev/gomponents"
    . "maragu.dev/gomponents/html"
)

// Tailwind CSS
Div(Class("flex items-center justify-between p-4 bg-blue-500"))

// Bootstrap
Div(Class("container-fluid"),
    Div(Class("row"),
        Div(Class("col-md-6"), Text("Column 1")),
        Div(Class("col-md-6"), Text("Column 2")),
    ),
)
```

### With JavaScript
Include scripts and handle interactions:
```go
import (
    . "maragu.dev/gomponents"
    . "maragu.dev/gomponents/html"
)

Button(
    Class("interactive-btn"),
    ID("myButton"),
    Text("Click me"),
)

Script(Raw(`
    document.getElementById('myButton').addEventListener('click', () => {
        alert('Clicked!');
    });
`))
```

### Custom Elements
For web components or non-standard elements:
```go
import (
    . "maragu.dev/gomponents"
)

// <my-component attr="value">Content</my-component>
El("my-component", 
    Attr("attr", "value"),
    Text("Content"),
)
```

## Event Handling

gomponents is designed for server-side HTML generation and does not provide `OnClick`-style event handling directly within Go code. The library focuses on generating static HTML markup that can be enhanced with client-side JavaScript for interactivity.

### JavaScript-Based Event Handling

For any interactivity or event handling, you should use pure JavaScript to attach event listeners to the rendered DOM elements. This separation of concerns keeps the Go code focused on HTML structure while delegating dynamic behavior to the client side.

### Example: Adding Click Handlers

Here's how to create a button with gomponents and attach a click event handler using JavaScript:

```go
import (
    . "maragu.dev/gomponents"
    . "maragu.dev/gomponents/html"
)

// Create a button with gomponents
func InteractiveButton() Node {
    return Div(
        // Button with an ID for JavaScript targeting
        Button(
            ID("my-button"),
            Class("btn btn-primary"),
            Text("Click Me!"),
        ),
        
        // JavaScript to handle the click event
        Script(Raw(`
            document.addEventListener('DOMContentLoaded', function() {
                const button = document.getElementById('my-button');
                button.addEventListener('click', function(event) {
                    alert('Button was clicked!');
                    console.log('Click event triggered');
                });
            });
        `)),
    )
}
```

### Preferred Pattern: Go Logic with JavaScript Binding

When client-side logic is required, the preferred method is to define the logic in Go and expose it to JavaScript using `js.Global().Set`. This approach keeps the core logic within Go, maintaining better organization and type safety, while still using standard JavaScript for DOM event binding.

#### Exposing Go Functions to JavaScript

```go
import (
    "syscall/js"
    . "maragu.dev/gomponents"
    . "maragu.dev/gomponents/html"
)

// Define your logic in Go
func increment(this js.Value, args []js.Value) interface{} {
    // Your Go logic here
    counter := args[0].Int()
    newValue := counter + 1
    
    // Update the DOM element
    document := js.Global().Get("document")
    element := document.Call("getElementById", "counter")
    element.Set("textContent", newValue)
    
    return newValue
}

// In your Go code, expose the function to JavaScript
js.Global().Set("increment", js.FuncOf(increment))

// Create the HTML structure
func CounterComponent() Node {
    return Div(
        P(ID("counter"), Text("0")),
        Button(
            ID("increment-btn"),
            Text("Increment"),
        ),
        
        // JavaScript to bind the event
        Script(Raw(`
            document.addEventListener('DOMContentLoaded', function() {
                const button = document.getElementById('increment-btn');
                const counter = document.getElementById('counter');
                
                button.addEventListener('click', function() {
                    const currentValue = parseInt(counter.textContent);
                    // Call the Go function exposed to JavaScript
                    window.increment(currentValue);
                });
            });
        `)),
    )
}
```

#### Benefits of This Pattern

1. **Type Safety**: Core logic remains in Go with compile-time type checking
2. **Code Organization**: Business logic stays centralized in Go code
3. **Debugging**: Use Go's debugging tools for complex logic
4. **Testing**: Go functions can be unit tested independently
5. **Performance**: Go logic can be more performant than JavaScript for complex operations

This pattern is particularly valuable for applications that need to share logic between server and client, or when you want to leverage Go's strengths (type safety, performance, tooling) for client-side behavior.

### Best Practices for Event Handling

1. **Use IDs or classes** to target elements from JavaScript
2. **Wrap event listeners** in `DOMContentLoaded` to ensure elements exist
3. **Keep JavaScript separate** from Go logic when possible
4. **Consider using external JavaScript files** for complex interactions
5. **Use data attributes** to pass data from Go to JavaScript:

```go
Button(
    ID("submit-btn"),
    DataAttr("user-id", "123"),
    DataAttr("action", "submit"),
    Text("Submit"),
)

Script(Raw(`
    document.getElementById('submit-btn').addEventListener('click', function(e) {
        const userId = e.target.dataset.userId;
        const action = e.target.dataset.action;
        // Use the data in your JavaScript logic
    });
`))
```

This approach maintains the clean separation between server-side HTML generation (Go) and client-side behavior (JavaScript), which is a core principle of gomponents' design philosophy.

## Best Practice: Interacting with Go from JavaScript

When building interactive components with gomponents, it's important to follow best practices for structuring your code. A key principle is to avoid writing complex, multi-line JavaScript directly inside `Script(Raw(...))` tags, and instead structure your code to keep business logic in Go while using minimal JavaScript for DOM event binding.

### ❌ What to Avoid

Don't write complex logic directly in JavaScript within `Script(Raw(...))` tags:

```go
import (
    . "maragu.dev/gomponents"
    . "maragu.dev/gomponents/html"
)

func CounterComponent() Node {
    return Div(
        P(ID("counter"), Text("0")),
        Button(ID("increment-btn"), Text("Increment")),
        
        // AVOID: Complex logic embedded in JavaScript
        Script(Raw(`
            document.addEventListener('DOMContentLoaded', function() {
                const button = document.getElementById('increment-btn');
                const counter = document.getElementById('counter');
                
                button.addEventListener('click', function() {
                    const currentValue = parseInt(counter.textContent);
                    
                    // Complex business logic mixed with DOM manipulation
                    let newValue;
                    if (currentValue < 10) {
                        newValue = currentValue + 1;
                    } else if (currentValue < 100) {
                        newValue = currentValue + 5;
                    } else {
                        newValue = currentValue + 10;
                    }
                    
                    // Validation logic
                    if (newValue > 1000) {
                        alert('Maximum value reached!');
                        newValue = 1000;
                    }
                    
                    counter.textContent = newValue;
                    
                    // Additional side effects
                    if (newValue % 10 === 0) {
                        console.log('Milestone reached:', newValue);
                    }
                });
            });
        `)),
    )
}
```

### ✅ Recommended Approach

Instead, write your core logic in Go and expose it to JavaScript using `js.Global().Set()`, then use minimal JavaScript for event binding:

```go
import (
    "syscall/js"
    . "maragu.dev/gomponents"
    . "maragu.dev/gomponents/html"
)

// Core business logic in Go - type-safe and testable
func calculateIncrement(current int) int {
    if current < 10 {
        return current + 1
    } else if current < 100 {
        return current + 5
    } else {
        return current + 10
    }
}

func validateValue(value int) (int, bool) {
    if value > 1000 {
        return 1000, true // Return max value and indicate limit reached
    }
    return value, false
}

func logMilestone(value int) {
    if value%10 == 0 {
        js.Global().Get("console").Call("log", "Milestone reached:", value)
    }
}

// Go function exposed to JavaScript
func increment(this js.Value, args []js.Value) interface{} {
    currentValue := args[0].Int()
    
    // Use Go functions for business logic
    newValue := calculateIncrement(currentValue)
    finalValue, limitReached := validateValue(newValue)
    
    if limitReached {
        js.Global().Call("alert", "Maximum value reached!")
    }
    
    // Update DOM
    document := js.Global().Get("document")
    element := document.Call("getElementById", "counter")
    element.Set("textContent", finalValue)
    
    // Log milestone
    logMilestone(finalValue)
    
    return finalValue
}

// Expose the Go function to JavaScript (typically in your main function)
func init() {
    js.Global().Set("increment", js.FuncOf(increment))
}

// Component with minimal JavaScript
func CounterComponent() Node {
    return Div(
        P(ID("counter"), Text("0")),
        Button(ID("increment-btn"), Text("Increment")),
        
        // Minimal JavaScript - just event binding
        Script(Raw(`
            document.addEventListener('DOMContentLoaded', function() {
                const button = document.getElementById('increment-btn');
                const counter = document.getElementById('counter');
                
                button.addEventListener('click', function() {
                    const currentValue = parseInt(counter.textContent);
                    // Call the Go function exposed to JavaScript
                    window.increment(currentValue);
                });
            });
        `)),
    )
}
```

### Why This Approach is Better

1. **Type Safety**: Business logic remains in Go with compile-time type checking, reducing runtime errors.

2. **Code Organization**: Core logic is centralized in Go functions that can be easily found, understood, and maintained.

3. **Testability**: Go functions can be unit tested independently using standard Go testing tools:
   ```go
   func TestCalculateIncrement(t *testing.T) {
       tests := []struct {
           input    int
           expected int
       }{
           {5, 6},
           {15, 20},
           {150, 160},
       }
       
       for _, test := range tests {
           result := calculateIncrement(test.input)
           if result != test.expected {
               t.Errorf("calculateIncrement(%d) = %d; want %d", test.input, result, test.expected)
           }
       }
   }
   ```

4. **Debugging**: Use Go's excellent debugging tools and error handling for complex logic instead of browser developer tools.

5. **Performance**: Go logic can be more performant than JavaScript for computationally intensive operations.

6. **Code Reuse**: Business logic written in Go can be shared between server-side and client-side code.

7. **Maintainability**: Separating concerns makes the code easier to read and modify. JavaScript is limited to simple DOM manipulation and event handling.

This pattern keeps the JavaScript minimal and focused solely on DOM event binding, while all business logic remains in Go where it benefits from strong typing, better tooling, and easier testing.

## Debugging

### String() Method
All nodes implement String() for debugging:
```go
import (
    "fmt"
    
    . "maragu.dev/gomponents"
    . "maragu.dev/gomponents/html"
)

node := Div(Class("test"), Text("Hello"))
fmt.Println(node) // <div class="test">Hello</div>
```

### Rendering to Buffer
Test component output:
```go
import (
    "bytes"
)

var buf bytes.Buffer
err := node.Render(&buf)
html := buf.String()
```

## Performance Considerations

1. **Direct Rendering**: Nodes render directly to io.Writer without intermediate string allocation
2. **No Reflection**: Pure function calls, no runtime reflection overhead
3. **Compile-Time Safety**: Errors caught at compile time, not runtime
4. **Zero Dependencies**: Core library has no external dependencies

## Common Gotchas

1. **Nil Nodes**: Nil nodes are safely ignored during rendering
2. **Attribute Order**: Attributes render in the order they're specified
3. **Escaping**: Use Text() for escaped content, Raw() for unescaped HTML
4. **Void Elements**: Children (except attributes) are ignored for void elements

## Summary

gomponents provides a type-safe, performant way to generate HTML in Go applications. It's particularly well-suited for:
- Server-side rendered web applications
- API servers that return HTML
- Static site generators
- Email template generation
- Any scenario where you need programmatic HTML generation with Go's type safety

The library's philosophy emphasizes simplicity, type safety, and Go idioms over template languages, making it an excellent choice for Go developers who prefer staying within the Go ecosystem.