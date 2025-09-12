# AI Repo Info

This document provides information for AI code generation.

## File/Package Structure

*   `/action`: Core package for managing asynchronous operations, side effects, and state with lifecycle-aware components. Includes concepts like actions, queries, streams, and subscriptions.
*   `/appmanager`: Provides a high-level manager for bootstrapping and running the entire application.
*   `/compat`: Contains compatibility helpers, particularly for integrating with React components.
*   `/comps`: Home to the core component model, page management, and mounting logic.
*   `/debug`: Utilities and examples for debugging applications.
*   `/docs`: Project documentation.
*   `/dom`: Handles DOM manipulation, event handling, and reactive rendering logic.
*   `/examples`: A directory containing various standalone examples demonstrating different features of the framework. Each subdirectory is a runnable application.
*   `/form`: Utilities and components for handling form state and validation (likely, based on name).
*   `/internal`: Internal packages used by the framework, not intended for public use. Includes test helpers.
*   `/js`: (Likely) Low-level JavaScript interop helpers.
*   `/reactivity`: The heart of the reactive system, containing signals, effects, memos, and stores for state management.
*   `/router`: Provides client-side routing capabilities.
*   `/scripts`: Shell scripts for development, building, and testing.
*   `/spec`: (Likely) Specification documents.
*   `/templates`: Go templates for code generation or other purposes.
*   `/viteassets`: Assets related to the Vite development server.
*   `/wasm`: (Likely) WASM-related utilities or build artifacts.

## `maragu.dev/gomponents` API

`gomponents` is a library for writing declarative, type-safe HTML in pure Go. It allows building reusable UI components without a separate templating language.

### Core Concepts

- **`Node`**: The central interface representing anything that can be rendered to an `io.Writer`. This includes elements, attributes, and text.
- **Declarative HTML**: Write HTML structures using Go functions (e.g., `Div()`, `P()`, `A()`).
- **Type-Safety**: HTML structure and attributes are checked at compile-time.
- **Component-Based**: Create reusable components by defining functions that return a `Node`.

### Key Packages and Usage

- `import . "maragu.dev/gomponents"`: Core types and helpers.
- `import . "maragu.dev/gomponents/html"`: HTML element and attribute functions (e.g., `Div`, `ID`, `Class`).
- `import . "maragu.dev/gomponents/components"`: Higher-level, pre-built components (e.g., `HTML5`).

### Core API (`gomponents` package)

- **`Node` interface**:
  ```go
  type Node interface {
      Render(w io.Writer) error
  }
  ```

- **`El(name string, children ...Node) Node`**: Creates an HTML element.
  - Example: `El("div", Attr("id", "main"), Text("Hello"))`

- **`Attr(name string, value ...string) Node`**: Creates an HTML attribute.
  - Boolean attribute: `Attr("disabled")`
  - Key-value attribute: `Attr("href", "/")`

- **Text Nodes**:
  - `Text(t string) Node`: Renders an HTML-escaped string.
  - `Textf(format string, a ...any) Node`: Renders a formatted, HTML-escaped string.
  - `Raw(t string) Node`: Renders a raw, unescaped string. Use with care.
  - `Rawf(format string, a ...any) Node`: Renders a formatted, raw string.

- **Control Flow &amp; Structuring**:
  - `If(condition bool, n Node) Node`: Conditionally renders a `Node`. The node is always evaluated.
  - `Iff(condition bool, f func() Node) Node`: Conditionally renders a `Node` from a function. The function is only called if the condition is true. Useful for avoiding nil panics.
  - `Map[T any](ts []T, cb func(T) Node) Group`: Maps a slice of data to a slice of `Node`s.
  - `Group([]Node) Node`: Groups multiple nodes into a single `Node`.

### `html` Package

This package provides convenient functions for all standard HTML5 elements and attributes.

- **Elements**: `Div(...)`, `Span(...)`, `H1(...)`, `P(...)`, `A(...)`, `Input(...)`, etc.
- **Attributes**: `ID(...)`, `Class(...)`, `Href(...)`, `Src(...)`, `Type(...)`, `Style(...)`, etc.

**Example:**
```go
import . "maragu.dev/gomponents"
import . "maragu.dev/gomponents/html"

func SimpleComponent() Node {
    return Div(
        ID("container"),
        Class("p-4"),
        H1(Text("Welcome")),
        P(Text("This is a gomponents component.")),
    )
}
```

### `components` Package

Provides higher-level, pre-built components.

- **`HTML5(props HTML5Props) Node`**: Creates a full HTML5 document structure, including `<!DOCTYPE>`, `<html>`, `<head>`, and `<body>`.
  ```go
  HTML5(HTML5Props{
      Title: "My Page",
      Head: []Node{
          Link(Rel("stylesheet"), Href("/styles.css")),
      },
      Body: []Node{
          // ... page content
      },
  })
  ```
- **`Classes` helper**: A map-based helper to conditionally apply CSS classes.
  ```go
  // "active" class is applied only if isActive is true.
  A(Href("/"), Classes{"active": isActive}, Text("Home"))
  ```
## `honnef.co/go/js/dom/v2` API

This package provides statically-typed Go bindings for the browser's JavaScript DOM APIs. It is the preferred way to interact with the DOM over `syscall/js` because it offers type safety and better autocompletion.

### Core Concepts

- **Statically-Typed Wrappers**: Provides Go types (structs and interfaces) that wrap native JavaScript DOM objects.
- **Idiomatic Go**: Aims to provide an idiomatic Go experience while closely mirroring the JavaScript DOM API structure.
- **Interfaces and Type Assertions**: Generic elements are returned as interfaces like `dom.Element` or `dom.Node`. You must use type assertions to access specific element types (e.g., `*dom.HTMLInputElement`).
- **Pointer Semantics**: DOM objects are pointers (`*dom.HTMLInputElement`). Changes made through one pointer are reflected in others pointing to the same underlying JS object.
- **Static Collections**: Methods like `QuerySelectorAll` return static Go slices (`[]dom.Element`), not live `NodeList`s. They are a snapshot and do not update automatically if the DOM changes.

### Getting Started

The primary entry point is `dom.GetWindow()`, which provides access to the global `window` object. From there, you can access the `document` and other global browser APIs.

```go
import "honnef.co/go/js/dom/v2"

func main() {
    // Get the window and document objects
    win := dom.GetWindow()
    doc := win.Document()

    // Find an element by its ID
    el := doc.GetElementByID("my-element")
    if el == nil {
        // Handle element not found
        return
    }

    // Manipulate the element
    el.SetInnerHTML("Hello from Go!")
}
```

### Key Interfaces &amp; Types

- **`Window`**: Represents the browser window.
  - `Document() Document`: Access the document object.
  - `Alert(string)`: Show an alert dialog.
  - `SetTimeout(fn func(), delay int) int`: Execute a function after a delay.
  - `RequestAnimationFrame(callback func(time.Duration)) int`: Schedule an animation frame callback.

- **`Document`**: Represents the HTML document.
  - `QuerySelector(sel string) Element`: Finds the first element matching a CSS selector.
  - `QuerySelectorAll(sel string) []Element`: Finds all elements matching a CSS selector.
  - `GetElementByID(id string) Element`: Gets an element by its ID.
  - `CreateElement(name string) Element`: Creates a new element.
  - `CreateTextNode(s string) *Text`: Creates a new text node.
  - `Body() HTMLElement`: Returns the `<body>` element.

- **`Element`**: A generic interface for all elements.
  - `SetInnerHTML(string)` / `InnerHTML() string`: Get or set the inner HTML.
  - `SetAttribute(name, value string)` / `GetAttribute(name string) string`: Manipulate attributes.
  - `Style() *CSSStyleDeclaration`: Access the element's inline style.
  - `Class() *TokenList`: Manipulate the element's CSS classes.
  - `AddEventListener(type string, useCapture bool, listener func(Event))`: Attach an event listener.
  - `Remove()`: Removes the element from the DOM.
  - `AppendChild(Node)`: Appends a child node.

- **Element-Specific Types**: You must cast from `Element` to access specialized properties.
  - `*HTMLInputElement`: `Value()`, `SetValue()`, `Checked()`
  - `*HTMLSelectElement`: `SelectedIndex()`, `Options()`
  - `*HTMLButtonElement`: `Disabled()`, `SetDisabled()`
  - `*HTMLAnchorElement`: `Href()`, `SetHref()`

- **`Event`**: Base interface for all events. Cast to specific event types for more details.
  - `PreventDefault()`: Prevents the browser's default action.
  - `Target() Element`: The element that triggered the event.
  - `*MouseEvent`: `ClientX()`, `ClientY()`
  - `*KeyboardEvent`: `Key()`, `Code()`

### Example: Handling an Input Event

```go
import "honnef.co/go/js/dom/v2"

func main() {
    doc := dom.GetWindow().Document()
    input := doc.GetElementByID("my-input")

    // Type assertion is required to access input-specific methods
    inputEl, ok := input.(*dom.HTMLInputElement)
    if !ok {
        return // Or handle error
    }

    inputEl.AddEventListener("input", false, func(e dom.Event) {
        // Here, e.Target() is the same as inputEl
        newValue := inputEl.Value()
        doc.GetElementByID("output").SetTextContent("You typed: " + newValue)
    })
}

```