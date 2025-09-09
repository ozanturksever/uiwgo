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

#### Inline Event Binding (Preferred)

**Inline event binding is the recommended approach** for handling DOM events in uiwgo applications. This method allows you to attach event handlers directly during element creation without requiring DOM queries or lifecycle hooks.

```go
// Inline event handlers - attach directly to elements
dom.OnClickInline(func(el dom.Element) { /* handler */ })     // Click events
dom.OnInputInline(func(el dom.Element) { /* handler */ })     // Input events
dom.OnChangeInline(func(el dom.Element) { /* handler */ })    // Change events
dom.OnEnterInline(func(el dom.Element) { /* handler */ })     // Enter key
dom.OnEscapeInline(func(el dom.Element) { /* handler */ })    // Escape key
dom.OnKeyDownInline(func(el dom.Element) { /* handler */ }, "key") // Custom keys
```

#### Inline Events Example (Recommended)
```go
func Component() g.Node {
    count := reactivity.NewSignal(0)

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

#### Traditional Event Binding (Legacy)

> **Note**: Traditional event binding is still supported but **not recommended** for new code. Use inline events instead.

Event binding is performed on `dom.Element` objects, typically within a `comps.OnMount` hook.

```go
// dom.BindClickToCallback binds a simple click handler.
func BindClickToCallback(element dom.Element, handler func()) *EventBinding

// dom.BindClickToSignal sets a signal to a specific value on click.
func BindClickToSignal[T any](element dom.Element, signal reactivity.Signal[T], value T) *EventBinding

// dom.BindEvent is a generic event binder.
func BindEvent(element dom.Element, eventType string, handler func(event dom.Event)) *EventBinding
```

#### Traditional Example (Not Recommended)
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

## Control Flow Helpers

UIwGo provides powerful helper functions for conditional rendering, loops, and dynamic content.

### Show

Conditionally render content based on a boolean signal.

```go
func Show(p ShowProps) g.Node

type ShowProps struct {
    When     reactivity.Signal[bool]
    Children g.Node
}
```

**Example:**
```go
func ConditionalContent() g.Node {
    visible := reactivity.NewSignal(true)
    
    return Div(
        Button(
            Text("Toggle"),
            OnClickInline(func(el dom.Element) {
                visible.Set(!visible.Get())
            }),
        ),
        Show(ShowProps{
            When: visible,
            Children: P(Text("This content is conditionally visible!")),
        }),
    )
}
```

### For

Render a list of items with efficient reconciliation and keying.

```go
func For[T any](p ForProps[T]) g.Node

type ForProps[T any] struct {
    Items    any // reactivity.Signal[[]T] or func() []T
    Key      func(T) string
    Children func(item T, index int) g.Node
}
```

**Example:**
```go
type Todo struct {
    ID   string
    Text string
    Done bool
}

func TodoList() g.Node {
    todos := reactivity.NewSignal([]Todo{
        {ID: "1", Text: "Learn UIwGo", Done: false},
        {ID: "2", Text: "Build an app", Done: true},
    })
    
    return Ul(
        For(ForProps[Todo]{
            Items: todos,
            Key: func(todo Todo) string { return todo.ID },
            Children: func(todo Todo, index int) g.Node {
                return Li(
                    Input(
                        Type("checkbox"),
                        If(todo.Done, Checked()),
                    ),
                    Text(todo.Text),
                )
            },
        }),
    )
}
```

### Index

Render a list with index-based reconciliation (useful when items don't have stable keys).

```go
func Index[T any](p IndexProps[T]) g.Node

type IndexProps[T any] struct {
    Items    any // reactivity.Signal[[]T] or func() []T
    Children func(getItem func() T, index int) g.Node
}
```

**Example:**
```go
func NumberList() g.Node {
    numbers := reactivity.NewSignal([]int{1, 2, 3, 4, 5})
    
    return Ul(
        Index(IndexProps[int]{
            Items: numbers,
            Children: func(getItem func() int, index int) g.Node {
                return Li(
                    Text(fmt.Sprintf("Item %d: %d", index, getItem())),
                )
            },
        }),
    )
}
```

### Switch and Match

Render different content based on a value (like a switch statement).

```go
func Switch(p SwitchProps) g.Node
func Match(p MatchProps) g.Node

type SwitchProps struct {
    When     any // reactivity.Signal[any] or func() any
    Fallback g.Node
    Children []g.Node // Array of Match nodes
}

type MatchProps struct {
    When     any // value or func() bool for matching
    Children g.Node
}
```

**Example:**
```go
func StatusDisplay() g.Node {
    status := reactivity.NewSignal("loading")
    
    return Div(
        Switch(SwitchProps{
            When: status,
            Fallback: P(Text("Unknown status")),
            Children: []g.Node{
                Match(MatchProps{
                    When: "loading",
                    Children: P(Text("Loading...")),
                }),
                Match(MatchProps{
                    When: "success",
                    Children: P(Text("✅ Success!")),
                }),
                Match(MatchProps{
                    When: "error",
                    Children: P(Text("❌ Error occurred")),
                }),
            },
        }),
    )
}
```

### Dynamic

Render components dynamically based on a signal.

```go
func Dynamic(p DynamicProps) g.Node

type DynamicProps struct {
    Component any // reactivity.Signal[ComponentFunc] or func() ComponentFunc
}
```

**Example:**
```go
func DynamicView() g.Node {
    currentView := reactivity.NewSignal(func() g.Node {
        return P(Text("Home View"))
    })
    
    return Div(
        Button(
            Text("Switch to Profile"),
            OnClickInline(func(el dom.Element) {
                currentView.Set(func() g.Node {
                    return P(Text("Profile View"))
                })
            }),
        ),
        Dynamic(DynamicProps{
            Component: currentView,
        }),
    )
}
```

## Utility Helpers

### Fragment

Group multiple nodes without creating a wrapper element.

```go
func Fragment(children ...g.Node) g.Node
```

**Example:**
```go
func MultipleElements() g.Node {
    return Fragment(
        H1(Text("Title")),
        P(Text("Description")),
        Button(Text("Action")),
    )
}
```

### Portal

Render content into a different DOM location.

```go
func Portal(target string, children g.Node) g.Node
```

**Example:**
```go
func ModalExample() g.Node {
    showModal := reactivity.NewSignal(false)
    
    return Div(
        Button(
            Text("Open Modal"),
            OnClickInline(func(el dom.Element) {
                showModal.Set(true)
            }),
        ),
        Show(ShowProps{
            When: showModal,
            Children: Portal("body", Div(
                Class("modal-overlay"),
                Div(
                    Class("modal"),
                    H2(Text("Modal Title")),
                    P(Text("Modal content goes here")),
                    Button(
                        Text("Close"),
                        OnClickInline(func(el dom.Element) {
                            showModal.Set(false)
                        }),
                    ),
                ),
            )),
        }),
    )
}
```

### Memo

Memoize component rendering based on dependencies.

```go
func Memo(component func() g.Node, dependencies ...any) g.Node

type MemoProps struct {
    Component    func() g.Node
    Dependencies []any
}
```

**Example:**
```go
func ExpensiveComponent(data []string) g.Node {
    return Memo(func() g.Node {
        // Expensive rendering logic
        var items []g.Node
        for _, item := range data {
            items = append(items, Li(Text(item)))
        }
        return Ul(items...)
    }, data) // Re-render only when data changes
}
```

### Lazy

Lazy-load components for code splitting.

```go
func Lazy(loader func() func() g.Node) g.Node

type LazyProps struct {
    Loader func() func() g.Node
}
```

**Example:**
```go
func LazyRoute() g.Node {
    return Lazy(func() func() g.Node {
        // This could load from a different module
        return func() g.Node {
            return Div(
                H1(Text("Lazy Loaded Component")),
                P(Text("This component was loaded on demand")),
            )
        }
    })
}
```

### ErrorBoundary

Catch and handle errors in component trees.

```go
func ErrorBoundary(props ErrorBoundaryProps) g.Node

type ErrorBoundaryProps struct {
    Fallback func(error) g.Node
    Children g.Node
}
```

**Example:**
```go
func SafeComponent() g.Node {
    return ErrorBoundary(ErrorBoundaryProps{
        Fallback: func(err error) g.Node {
            return Div(
                Class("error-boundary"),
                H2(Text("Something went wrong")),
                P(Text(err.Error())),
                Button(
                    Text("Retry"),
                    OnClickInline(func(el dom.Element) {
                        // Retry logic
                    }),
                ),
            )
        },
        Children: RiskyComponent(),
    })
}
```

## Helper Functions Integration

UIwGo provides powerful helper functions that work seamlessly with the core APIs to simplify common UI patterns.

### Conditional Rendering with Show

```go
// Basic conditional rendering
func LoginStatus() g.Node {
    isLoggedIn := reactivity.NewSignal(false)
    
    return comps.Show(comps.ShowProps{
        When: isLoggedIn,
        Children: h.Div(
            h.Text("Welcome back!"),
            h.Button(
                h.Text("Logout"),
                dom.OnClick(func() { isLoggedIn.Set(false) }),
            ),
        ),
        Fallback: h.Div(
            h.Text("Please log in"),
            h.Button(
                h.Text("Login"),
                dom.OnClick(func() { isLoggedIn.Set(true) }),
            ),
        ),
    })
}
```

### List Rendering with For

```go
// Dynamic list with reactive data
func TodoList() g.Node {
    todos := reactivity.NewSignal([]Todo{
        {ID: "1", Text: "Learn UIwGo", Done: false},
        {ID: "2", Text: "Build an app", Done: false},
    })
    
    return h.Div(
        h.H2(h.Text("Todo List")),
        h.Ul(
            comps.For(comps.ForProps[Todo]{
                Items: todos,
                Key: func(todo Todo) string { return todo.ID },
                Children: func(todo Todo, index int) g.Node {
                    return h.Li(
                        h.Input(
                            h.Type("checkbox"),
                            h.Checked(todo.Done),
                            dom.OnChange(func(checked bool) {
                                // Update todo status
                                current := todos.Get()
                                current[index].Done = checked
                                todos.Set(current)
                            }),
                        ),
                        h.Span(h.Text(todo.Text)),
                    )
                },
            }),
        ),
    )
}
```

### Multi-State Switching

```go
// Complex state management with Switch/Match
func DataView() g.Node {
    loadingState := reactivity.NewSignal("idle") // idle, loading, success, error
    data := reactivity.NewSignal([]string{})
    
    return h.Div(
        h.Button(
            h.Text("Load Data"),
            dom.OnClick(func() {
                loadingState.Set("loading")
                // Simulate async data loading
                go func() {
                    time.Sleep(2 * time.Second)
                    data.Set([]string{"Item 1", "Item 2", "Item 3"})
                    loadingState.Set("success")
                }()
            }),
        ),
        comps.Switch(comps.SwitchProps{
            When: loadingState,
            Children: []g.Node{
                comps.Match(comps.MatchProps{
                    When: "idle",
                    Children: h.P(h.Text("Click to load data")),
                }),
                comps.Match(comps.MatchProps{
                    When: "loading",
                    Children: h.P(h.Text("Loading...")),
                }),
                comps.Match(comps.MatchProps{
                    When: "success",
                    Children: h.Ul(
                        comps.For(comps.ForProps[string]{
                            Items: data,
                            Key: func(item string) string { return item },
                            Children: func(item string, index int) g.Node {
                                return h.Li(h.Text(item))
                            },
                        }),
                    ),
                }),
                comps.Match(comps.MatchProps{
                    When: "error",
                    Children: h.P(
                        h.Class("error"),
                        h.Text("Failed to load data"),
                    ),
                }),
            },
        }),
    )
}
```

### Reactive Text Binding

```go
// Dynamic text content with BindText
func UserProfile() g.Node {
    user := reactivity.NewSignal(User{Name: "John", Email: "john@example.com"})
    
    return h.Div(
        h.H1(
            h.Text("Welcome, "),
            comps.BindText(reactivity.NewMemo(func() string {
                return user.Get().Name
            })),
        ),
        h.P(
            h.Text("Email: "),
            comps.BindText(reactivity.NewMemo(func() string {
                return user.Get().Email
            })),
        ),
        h.Input(
            h.Placeholder("Update name"),
            dom.OnInput(func(value string) {
                current := user.Get()
                current.Name = value
                user.Set(current)
            }),
        ),
    )
}
```

For comprehensive helper function documentation, see:
- **[Helper Functions Guide](../guides/helper-functions.md)** - Complete guide with all patterns
- **[Quick Reference](../guides/quick-reference.md)** - Syntax reference
- **[Real-World Examples](../guides/real-world-examples.md)** - Practical applications

## Quick Reference

### Common Patterns

```go
// Functional component with state
func Counter() g.Node {
    count := reactivity.NewSignal(0)
    return Div(
        P(Text("Count: "), BindText(func() string { return strconv.Itoa(count.Get()) })),
        Button(Text("Increment"), OnClickInline(func(el dom.Element) {
            count.Set(count.Get() + 1)
        })),
    )
}

// Computed state
fullName := reactivity.NewMemo(func() string {
    return firstName.Get() + " " + lastName.Get()
})

// Side effects
reactivity.NewEffect(func() {
    logutil.Log("Count changed:", count.Get())
})

// Event handling
OnClickInline(func(el dom.Element) {
    // Handle click
})

// Cleanup
comps.OnCleanup(func() {
    // Cleanup resources
})

// Conditional rendering
Show(ShowProps{
    When: visible,
    Children: P(Text("Conditional content")),
})

// List rendering
For(ForProps[Item]{
    Items: items,
    Key: func(item Item) string { return item.ID },
    Children: func(item Item, index int) g.Node {
        return Li(Text(item.Name))
    },
})
```

Next: Explore [Control Flow](../guides/control-flow.md) or check out [Troubleshooting](../troubleshooting.md).