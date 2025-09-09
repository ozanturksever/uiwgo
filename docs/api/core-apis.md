# Core APIs Reference

This reference covers the essential APIs for building UIwGo applications. The APIs are organized by functionality and include practical examples for each method.

## Table of Contents

- [Component Model](#component-model)
- [Reactivity APIs](#reactivity-apis)
- [DOM & Binding APIs](#dom--binding-apis)
- [Mounting & Lifecycle](#mounting--lifecycle)
- [Type Definitions](#type-definitions)

## Component Model

### Functional Components

In UIwGo, a component is a Go function that returns a `gomponents.Node`. This functional approach allows for simple, composable, and type-safe UI development.

State is managed using signals, and side effects (like event binding) are handled using lifecycle hooks such as `comps.OnMount`.

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
    title := reactivity.NewSignal("Hello, UIwGo!")
    visible := reactivity.NewSignal(true)

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
        comps.Show(visible.Get,
            h.P(g.Text("This paragraph can be toggled.")),
        ),
        h.Button(
            h.ID("toggle-btn"),
            g.Text("Toggle Visibility"),
        ),
    )
}
```

## Reactivity APIs

### Signals

Signals are the foundation of UIwGo's reactivity system.

#### Creating Signals

```go
// Create a signal with an initial value
func NewSignal[T any](initial T) *Signal[T]

// Examples
name := reactivity.NewSignal("John")
count := reactivity.NewSignal(0)
user := reactivity.NewSignal(User{ID: 1, Name: "Alice"})
visible := reactivity.NewSignal(true)
```

#### Signal Methods

```go
type Signal[T any] interface {
    Get() T          // Get current value
    Set(value T)     // Set new value (triggers reactivity)
    Update(fn func(T) T) // Update value using a function
    Subscribe(fn func(T)) func() // Subscribe to changes
    Map[U any](fn func(T) U) Memo[U] // Transform to a different type
    Filter(fn func(T) bool) Memo[T] // Filter changes
}
```

### Memos

Memos are computed values that automatically update when their dependencies change.

#### Creating Memos

```go
// Create a memo from a function
func NewMemo[T any](fn func() T) *Memo[T]

// Example
firstName := reactivity.NewSignal("John")
lastName := reactivity.NewSignal("Doe")

fullName := reactivity.NewMemo(func() string {
    return firstName.Get() + " " + lastName.Get()
})
```

#### Memo Methods

```go
type Memo[T any] interface {
    Get() T
    Subscribe(fn func(T)) func()
    Map[U any](fn func(T) U) Memo[U]
    Filter(fn func(T) bool) Memo[T]
    Dispose()
}
```

### Effects

Effects run side effects in response to reactive changes.

#### Creating Effects

```go
// Create an effect that runs immediately and on dependency changes
func NewEffect(fn func()) Effect

// Example
name := reactivity.NewSignal("World")

effect := reactivity.NewEffect(func() {
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

### Batching

Batch multiple signal updates to trigger effects only once.

```go
// Batch function
func Batch(fn func())

// Example
reactivity.Batch(func() {
    firstName.Set("Alice")
    lastName.Set("Johnson")
}) // Effects depending on both signals run only once.
```

## DOM & Binding APIs

These helpers, mostly from the `comps` and `dom` packages, connect your reactive state to the DOM.

### Reactive Content & Visibility

These are `gomponents.Node` helpers used directly in your component's return tree.

```go
// comps.BindText creates a reactive text node.
func BindText(fn func() string) g.Node

// comps.Show conditionally renders its children nodes.
func Show(condition bool, children ...g.Node) g.Node
```

#### Example
```go
func Component() g.Node {
    name := reactivity.NewSignal("World")
    isVisible := reactivity.NewSignal(true)

    return Div(
        H1(
            BindText(func() string { return "Hello, " + name.Get() }),
        ),
        Show(isVisible.Get(),
            P(Text("You can see me!")),
        ),
    )
}
```

### Event Binding

Event binding is performed on `dom.Element` objects, typically within a `comps.OnMount` hook.

```go
// dom.BindClickToCallback binds a simple click handler.
func BindClickToCallback(element dom.Element, handler func()) *EventBinding

// dom.BindClickToSignal sets a signal to a specific value on click.
func BindClickToSignal[T any](element dom.Element, signal reactivity.Signal[T], value T) *EventBinding

// dom.BindEvent is a generic event binder.
func BindEvent(element dom.Element, eventType string, handler func(event dom.Event)) *EventBinding
```

#### Example
```go
func Component() g.Node {
    count := reactivity.NewSignal(0)

    comps.OnMount(func() {
        incrementBtn := dom.GetElementByID("increment-btn")
        resetBtn := dom.GetElementByID("reset-btn")

        dom.BindClickToCallback(incrementBtn, func() {
            count.Set(count.Get() + 1)
        })
        
        dom.BindClickToSignal(resetBtn, count, 0)
    })

    return Div(
        Button(ID("increment-btn"), Text("+")),
        Button(ID("reset-btn"), Text("Reset")),
    )
}
```

### Input Handling

The `dom` package provides helpers for two-way binding on input elements.

```go
// dom.BindValue binds a string signal to an input's value.
func BindValue(element dom.Element, signal reactivity.Signal[string]) *EventBinding

// dom.BindChecked binds a boolean signal to a checkbox's checked state.
func BindChecked(element dom.Element, signal reactivity.Signal[bool]) *EventBinding
```

#### Example
```go
func FormComponent() g.Node {
    name := reactivity.NewSignal("")
    agreed := reactivity.NewSignal(false)

    comps.OnMount(func() {
        nameInput := dom.GetElementByID("name-input")
        agreeCheckbox := dom.GetElementByID("agree-cb")

        dom.BindValue(nameInput, name)
        dom.BindChecked(agreeCheckbox, agreed)
    })
    
    return Div(
        Input(ID("name-input"), Type("text")),
        Input(ID("agree-cb"), Type("checkbox")),
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

// Readable interface (for signals and memos)
type Readable[T any] interface {
    Get() T
    Subscribe(fn func(T)) func()
    Map[U any](fn func(T) U) Memo[U]
    Filter(fn func(T) bool) Memo[T]
}

// Writable interface (for signals)
type Writable[T any] interface {
    Set(value T)
    Update(fn func(T) T)
}

// Signal interface
type Signal[T any] interface {
    Readable[T]
    Writable[T]
}

// Memo interface
type Memo[T any] interface {
    Readable[T]
    Dispose()
}

// Effect interface
type Effect interface {
    Dispose()
    IsDisposed() bool
}
```

---

## Quick Reference

### Common Patterns

```go
// 1. A basic functional component
func MyComponent() g.Node {
    // State
    message := reactivity.NewSignal("Hello")

    // View
    return h.Div(
        comps.BindText(message.Get),
    )
}

// 2. Computed state
fullName := reactivity.NewMemo(func() string {
    return firstName.Get() + " " + lastName.Get()
})

// 3. Side effects
reactivity.NewEffect(func() {
    title := pageTitle.Get()
    dom.GetWindow().Document().SetTitle(title)
})

// 4. Event handling
comps.OnMount(func() {
    btn := dom.GetElementByID("my-btn")
    dom.BindClickToCallback(btn, func() {
        count.Update(func(c int) int { return c + 1 })
    })
})

// 5. Cleanup
comps.OnCleanup(func() {
    myEffect.Dispose()
    myTimer.Stop()
})
```

Next: Explore [Control Flow](../guides/control-flow.md) or check out [Troubleshooting](../troubleshooting.md).