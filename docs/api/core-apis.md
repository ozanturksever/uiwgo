# Core APIs Reference

This reference covers the essential APIs for building UIwGo applications. The APIs are organized by functionality and include practical examples for each method.

**Note:** This documentation reflects the current, simplified API. Some previously documented features are under review for future implementation.

## Table of Contents

- [Component Model](#component-model)
- [Reactivity APIs](#reactivity-apis)
- [DOM & Binding APIs](#dom--binding-apis)
- [Mounting & Lifecycle](#mounting--lifecycle)
- [Type Definitions](#type-definitions)

## Component Model

### Functional Components

In UIwGo, a component is a Go function that returns a `gomponents.Node`. This functional approach allows for simple, composable, and type-safe UI development.

State is managed using signals, and interactions are handled primarily through inline event bindings, which is the recommended approach.

#### Usage Example

```go
import (
    "fmt"
    g "maragu.dev/gomponents"
    h "maragu.dev/gomponents/html"
    "github.com/ozanturksever/uiwgo/reactivity"
    "github.com/ozanturksever/uiwgo/comps"
    "github.com/ozanturksever/uiwgo/dom"
)

// MyComponent is a function that returns a gomponents Node.
func MyComponent() g.Node {
    // 1. Define reactive state using signals.
    title := reactivity.CreateSignal("Hello, UIwGo!")
    visible := reactivity.CreateSignal(true)

    // 2. Use OnMount to perform actions after the component's DOM is rendered.
    comps.OnMount(func() {
        // Find the button element by its ID.
        toggleBtn := dom.GetElementByID("toggle-btn")
        if toggleBtn != nil {
            // Bind a click event handler directly to the element.
            dom.BindClickToCallback(toggleBtn, func() {
                // Update the signal's value.
                visible.Set(!visible.Get())
            })
        }
    })

    // 3. Return the component's UI structure using gomponents.
    return h.Div(
        h.H1(
            // Bind the text content of this H1 to the title signal.
            comps.BindText(title.Get),
        ),
        // Use a helper for conditional rendering.
        comps.Show(comps.ShowProps{
            When: visible,
            Children: h.P(g.Text("This paragraph can be toggled.")),
        }),
        h.Button(
            g.Text("Toggle Visibility"),
            // Bind a click event handler directly to the element.
            dom.OnClick(func(evt dom.Event) {
                // Update the signal's value.
                visible.Set(!visible.Get())
            }),
        ),
    )
}
```

## Reactivity APIs

### Signals

Signals are the foundation of UIwGo's reactivity system. They hold a value and trigger updates when that value changes.

#### Creating Signals

```go
// Create a signal with an initial value
func CreateSignal[T any](initial T) reactivity.Signal[T]

// Examples
name := reactivity.CreateSignal("John")
count := reactivity.CreateSignal(0)
user := reactivity.CreateSignal(User{ID: 1, Name: "Alice"})
visible := reactivity.CreateSignal(true)
```

#### Signal Methods

A signal is a simple interface with just two methods.

```go
type Signal[T any] interface {
    Get() T      // Get current value
    Set(value T) // Set new value (triggers reactivity)
}
```

### Memos

Memos are computed, read-only signals that automatically update when their dependencies change.

#### Creating Memos

```go
// Create a memo from a function
func CreateMemo[T any](fn func() T) reactivity.Signal[T]

// Example
firstName := reactivity.CreateSignal("John")
lastName := reactivity.CreateSignal("Doe")

// fullName is a read-only signal that depends on firstName and lastName.
fullName := reactivity.CreateMemo(func() string {
    return firstName.Get() + " " + lastName.Get()
})
```
*Note: `CreateMemo` returns a `Signal[T]`, which is read-only in practice because it's derived from other signals.*

### Effects

Effects run side effects in response to reactive changes.

#### Creating Effects

```go
// Create an effect that runs immediately and on dependency changes
func CreateEffect(fn func()) Effect

// Example
name := reactivity.CreateSignal("World")

effect := reactivity.CreateEffect(func() {
    logutil.Logf("Hello, %s!", name.Get())
})
// Immediately logs: "Hello, World!"

name.Set("UIwGo")
// Logs: "Hello, UIwGo!"
```

#### Effect Methods

```go
type Effect interface {
    Dispose()
    IsDisposed() bool
}
```

## DOM & Binding APIs

These helpers, from the `comps` and `dom` packages, connect your reactive state to the DOM.

### Reactive Content & Visibility

These are `gomponents.Node` helpers used directly in your component's return tree.

```go
// comps.BindText creates a reactive text node.
func BindText(fn func() string) g.Node

// comps.Show conditionally renders its children nodes.
func Show(p ShowProps) g.Node

type ShowProps struct {
    When     reactivity.Signal[bool]
    Children g.Node
}
```

#### Example
```go
func Component() g.Node {
    name := reactivity.CreateSignal("World")
    isVisible := reactivity.CreateSignal(true)

    return h.Div(
        h.H1(
            comps.BindText(func() string { return "Hello, " + name.Get() }),
        ),
        comps.Show(comps.ShowProps{
            When: isVisible,
            Children: h.P(g.Text("You can see me!")),
        }),
    )
}
```

### Event Binding

**Inline event binding is the standard approach** for handling DOM events in UIwGo. This method allows you to attach event handlers directly during element creation. All inline event handlers return a `gomponents.Node`.

```go
// Inline event handlers - attach directly to elements
dom.OnClickInline(func(el dom.Element) { /* handler */ })     // Click events
dom.OnInputInline(func(el dom.Element) { /* handler */ })     // Input events
dom.OnChangeInline(func(el dom.Element) { /* handler */ })    // Change events
dom.OnEnterInline(func(el dom.Element) { /* handler */ })     // Enter key
dom.OnEscapeInline(func(el dom.Element) { /* handler */ })    // Escape key
dom.OnKeyDownInline(func(el dom.Element) { /* handler */ }, "key") // Custom keys
```

#### Inline Events Example
```go
func Component() g.Node {
    count := reactivity.CreateSignal(0)

    return Div(
        Button(
            Text("+"),
            dom.OnClickInline(func(el dom.Element) {
                count.Set(count.Get() + 1)
            }),
        ),
        Button(
            Text("Reset"),
            dom.OnClickInline(func(el dom.Element) {
                count.Set(0)
            }),
        ),
        P(comps.BindText(func() string {
            return fmt.Sprintf("Count: %d", count.Get())
        })),
    )
}
```

### Input Handling

The `dom` package provides helpers for two-way binding on input elements. These are also used as inline attributes.

```go
// dom.BindInputToSignal binds a string signal to an input's value.
func BindInputToSignal(s reactivity.Signal[string]) g.Node

// dom.BindChangeToSignal binds a boolean signal to a checkbox's checked state.
func BindChangeToSignal(s reactivity.Signal[bool]) g.Node
```

#### Example
```go
func FormComponent() g.Node {
    name := reactivity.CreateSignal("")
    agreed := reactivity.CreateSignal(false)

    return h.Div(
        h.Input(h.Type("text"), dom.BindInputToSignal(name)),
        h.Input(h.Type("checkbox"), dom.BindChangeToSignal(agreed)),
        h.P(comps.BindText(func() string {
            return fmt.Sprintf("Name: %s, Agreed: %t", name.Get(), agreed.Get())
        })),
    )
}
```

## Mounting & Lifecycle

### Mounting Components

You start your application by mounting a top-level component function to a DOM element.

```go
// Mounts a component function to a DOM element by its ID.
// Returns a "disposer" function to unmount and clean up the component.
func Mount(elementID string, component func() g.Node) func()

// Example
func main() {
    // This code runs in the browser via WASM
    disposer := comps.Mount("app", MyComponent)
    
    // To unmount later:
    // disposer()

    // Prevent the Go program from exiting.
    select {}
}
```

### Lifecycle Hooks

Lifecycle hooks allow you to run code at specific points in a component's life.

```go
// comps.OnMount registers a function to run after the component is rendered to the DOM.
func OnMount(fn func())

// comps.OnCleanup registers a function to run when the component is unmounted.
func OnCleanup(fn func())
```

#### Example
```go
func MyComponent() g.Node {
    OnMount(func() {
        logutil.Log("Component is now in the DOM.")
        // Perfect for setting up event listeners, timers, or fetching data.
    })

    OnCleanup(func() {
        logutil.Log("Component is being removed.")
        // Perfect for stopping timers, closing connections, or disposing effects.
    })

    return P(Text("Hello"))
}
```

## Type Definitions

### Core Types

```go
// A Component is a function that returns a gomponents Node.
type Component func() g.Node

// Signal interface
type Signal[T any] interface {
    Get() T
    Set(value T)
}

// Effect interface
type Effect interface {
    Dispose()
    IsDisposed() bool
}

// ShowProps for conditional rendering
type ShowProps struct {
    When     reactivity.Signal[bool]
    Children g.Node
}