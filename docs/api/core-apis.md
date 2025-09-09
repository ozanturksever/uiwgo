# Core APIs Reference

This reference covers the essential APIs for building UIwGo applications. The APIs are organized by functionality and include practical examples for each method.

## Table of Contents

- [Component APIs](#component-apis)
- [Reactivity APIs](#reactivity-apis)
- [DOM Binding APIs](#dom-binding-apis)
- [Mounting & Lifecycle](#mounting--lifecycle)
- [Utility Functions](#utility-functions)
- [Type Definitions](#type-definitions)

## Component APIs

### Component Interface

All UIwGo components must implement the `Component` interface:

```go
import g "maragu.dev/gomponents"

type Component interface {
    Render() g.Node
    Attach()
}
```

#### Optional Interfaces

```go
// For components that need cleanup
type CleanupComponent interface {
    Component
    Cleanup()
}

// For components with children
type ParentComponent interface {
    Component
    GetChildren() []Component
}
```

### Component Binding

UIwGo provides standalone binding functions in the `comps` package for creating reactive components. These functions are used within a component's `Attach()` method to bind reactive values to DOM elements.

```go
// Binding functions available in the comps package:
func BindText(selector string, signal reactivity.Readable[string])
func Show(selector string, signal reactivity.Readable[bool])
func BindClass(selector string, className string, signal reactivity.Readable[bool])
func BindAttr(selector string, attr string, signal reactivity.Readable[string])
func BindClick(selector string, handler func())
func BindInput(selector string, signal *reactivity.Signal[string])
func BindEvent(selector string, eventType string, handler func(dom.Event))
```

#### Usage Example

```go
import (
    g "maragu.dev/gomponents"
    h "maragu.dev/gomponents/html"
    "github.com/ozanturksever/uiwgo/reactivity"
    "github.com/ozanturksever/uiwgo/comps"
)

type MyComponent struct {
    title   *reactivity.Signal[string]
    visible *reactivity.Signal[bool]
}

func NewMyComponent() *MyComponent {
    return &MyComponent{
        title:   reactivity.NewSignal("Hello"),
        visible: reactivity.NewSignal(true),
    }
}

func (c *MyComponent) Render() g.Node {
    return h.Div(g.Class("my-component"),
        h.H1(
            g.Attr("data-text", "title"),
            g.Text("Loading..."),
        ),
        h.Button(
            g.Attr("data-click", "toggle"),
            g.Attr("data-show", "visible"),
            g.Text("Toggle"),
        ),
    )
}

func (c *MyComponent) Attach() {
    comps.BindText("title", c.title)
    comps.Show("visible", c.visible)
    comps.BindClick("toggle", c.toggle)
}

func (c *MyComponent) toggle() {
    c.visible.Set(!c.visible.Get())
}
```

## Reactivity APIs

### Signals

Signals are the foundation of UIwGo's reactivity system.

#### Creating Signals

```go
// Create a signal with initial value
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
    // Get current value
    Get() T
    
    // Set new value (triggers reactivity)
    Set(value T)
    
    // Update value using a function
    Update(fn func(T) T)
    
    // Subscribe to changes
    Subscribe(fn func(T)) func()
    
    // Transform to different type
    Map[U any](fn func(T) U) Memo[U]
    
    // Filter changes
    Filter(fn func(T) bool) Memo[T]
}
```

#### Signal Examples

```go
// Basic usage
count := reactivity.NewSignal(0)
logutil.Logf("Count: %d", count.Get()) // Count: 0

count.Set(5)
logutil.Logf("Count: %d", count.Get()) // Count: 5

// Update with function
count.Update(func(current int) int {
    return current + 1
})
logutil.Logf("Count: %d", count.Get()) // Count: 6

// Subscribe to changes
unsubscribe := count.Subscribe(func(value int) {
    logutil.Logf("Count changed to: %d", value)
})
defer unsubscribe()

count.Set(10) // Logs: "Count changed to: 10"

// Transform signal
name := reactivity.NewSignal("john")
uppercaseName := name.Map(func(n string) string {
    return strings.ToUpper(n)
})

logutil.Logf("Uppercase: %s", uppercaseName.Get()) // Uppercase: JOHN

// Filter signal
positive := count.Filter(func(n int) bool {
    return n > 0
})
```

### Memos

Memos are computed values that automatically update when their dependencies change.

#### Creating Memos

```go
// Create a memo from a function
func NewMemo[T any](fn func() T) *Memo[T]

// Examples
firstName := reactivity.NewSignal("John")
lastName := reactivity.NewSignal("Doe")

fullName := reactivity.NewMemo(func() string {
    return firstName.Get() + " " + lastName.Get()
})

count := reactivity.NewSignal(0)
isEven := reactivity.NewMemo(func() bool {
    return count.Get()%2 == 0
})
```

#### Memo Methods

```go
type Memo[T any] interface {
    // Get current computed value
    Get() T
    
    // Subscribe to changes
    Subscribe(fn func(T)) func()
    
    // Transform to different type
    Map[U any](fn func(T) U) Memo[U]
    
    // Filter changes
    Filter(fn func(T) bool) Memo[T]
    
    // Dispose the memo
    Dispose()
}
```

#### Memo Examples

```go
// Complex computation
type User struct {
    FirstName string
    LastName  string
    Age       int
}

user := reactivity.NewSignal(User{
    FirstName: "John",
    LastName:  "Doe",
    Age:       25,
})

// Computed display name
displayName := reactivity.NewMemo(func() string {
    u := user.Get()
    return fmt.Sprintf("%s %s (%d)", u.FirstName, u.LastName, u.Age)
})

// Computed validation
isAdult := reactivity.NewMemo(func() bool {
    return user.Get().Age >= 18
})

// Chained memos
canVote := reactivity.NewMemo(func() bool {
    u := user.Get()
    return u.Age >= 18 && u.FirstName != "" && u.LastName != ""
})

votingStatus := reactivity.NewMemo(func() string {
    if canVote.Get() {
        return "Eligible to vote"
    }
    return "Not eligible to vote"
})

logutil.Logf("Status: %s", votingStatus.Get())
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
    // Stop the effect from running
    Dispose()
    
    // Check if effect is disposed
    IsDisposed() bool
}
```

#### Effect Examples

```go
// Document title effect
title := reactivity.NewSignal("Home")

titleEffect := reactivity.NewEffect(func() {
    document := dom.GetWindow().Document()
    document.SetTitle("MyApp - " + title.Get())
})

// Local storage effect
settings := reactivity.NewSignal(map[string]string{
    "theme": "dark",
    "lang":  "en",
})

storageEffect := reactivity.NewEffect(func() {
    settings := settings.Get()
    localStorage := dom.GetWindow().LocalStorage()
    
    for key, value := range settings {
        localStorage.SetItem(key, value)
    }
})

// Cleanup effects
defer func() {
    titleEffect.Dispose()
    storageEffect.Dispose()
}()
```

### Batching

Batch multiple signal updates to trigger effects only once.

```go
// Batch function
func Batch(fn func())

// Example
firstName := reactivity.NewSignal("John")
lastName := reactivity.NewSignal("Doe")

fullName := reactivity.NewMemo(func() string {
    return firstName.Get() + " " + lastName.Get()
})

reactivity.NewEffect(func() {
    logutil.Logf("Full name: %s", fullName.Get())
})

// Without batching: effect runs twice
firstName.Set("Jane")
lastName.Set("Smith")

// With batching: effect runs once
reactivity.Batch(func() {
    firstName.Set("Alice")
    lastName.Set("Johnson")
})
```

## DOM Binding APIs

### Text Binding

Bind reactive values to element text content.

```go
// Bind text content
func BindText(selector string, signal reactivity.Readable[string])

// Example
name := reactivity.NewSignal("World")
comps.BindText("greeting", name.Map(func(n string) string {
    return "Hello, " + n + "!"
}))

// HTML: <span data-text="greeting">Loading...</span>
// Result: <span>Hello, World!</span>
```

### Visibility Binding

Bind reactive boolean values to element visibility.

```go
// Show/hide elements
func Show(selector string, signal reactivity.Readable[bool])

// Example
visible := reactivity.NewSignal(true)
comps.Show("content", visible)

// HTML: <div data-show="content">Content</div>
// When visible=true: <div style="display: block">Content</div>
// When visible=false: <div style="display: none">Content</div>
```

### Class Binding

Bind reactive boolean values to CSS classes.

```go
// Toggle CSS classes
func BindClass(selector string, className string, signal reactivity.Readable[bool])

// Example
active := reactivity.NewSignal(false)
comps.BindClass("button", "active", active)

// HTML: <button data-class-active="button">Click me</button>
// When active=true: <button class="active">Click me</button>
// When active=false: <button class="">Click me</button>
```

### Attribute Binding

Bind reactive values to element attributes.

```go
// Bind attributes
func BindAttr(selector string, attr string, signal reactivity.Readable[string])

// Example
imageURL := reactivity.NewSignal("/default.jpg")
comps.BindAttr("avatar", "src", imageURL)

// HTML: <img data-attr-src="avatar" alt="Avatar" />
// Result: <img src="/default.jpg" alt="Avatar" />
```

### Event Binding

Bind event handlers to elements.

```go
// Click events
func BindClick(selector string, handler func())

// Input events
func BindInput(selector string, signal *reactivity.Signal[string])

// Custom events
func BindEvent(selector string, eventType string, handler func(dom.Event))

// Examples
count := reactivity.NewSignal(0)

// Click handler
comps.BindClick("increment", func() {
    count.Update(func(c int) int { return c + 1 })
})

// Input binding (two-way)
name := reactivity.NewSignal("")
comps.BindInput("nameInput", name)

// Custom event
comps.BindEvent("customButton", "mouseenter", func(e dom.Event) {
    logutil.Log("Mouse entered button")
})
```

### Advanced Binding

#### Multiple Class Binding

```go
// Bind multiple classes
type ButtonState struct {
    Loading  bool
    Disabled bool
    Primary  bool
}

buttonState := reactivity.NewSignal(ButtonState{
    Loading:  false,
    Disabled: false,
    Primary:  true,
})

comps.BindClass("button", "loading", buttonState.Map(func(s ButtonState) bool {
    return s.Loading
}))
comps.BindClass("button", "disabled", buttonState.Map(func(s ButtonState) bool {
    return s.Disabled
}))
comps.BindClass("button", "primary", buttonState.Map(func(s ButtonState) bool {
    return s.Primary
}))
```

#### Conditional Attributes

```go
// Conditional href attribute
link := reactivity.NewSignal("")
enabled := reactivity.NewSignal(true)

comps.BindAttr("link", "href", reactivity.NewMemo(func() string {
    if enabled.Get() && link.Get() != "" {
        return link.Get()
    }
    return "#"
}))
```

## Mounting & Lifecycle

### Mounting Components

```go
// Mount a component to a DOM element
func Mount(elementID string, component Component) error

// Mount multiple components
func MountAll(mounts map[string]Component) error

// Examples
app := NewApp()
err := comps.Mount("app", app)
if err != nil {
    logutil.Logf("Failed to mount: %v", err)
}

// Mount multiple
err = comps.MountAll(map[string]Component{
    "header": NewHeader(),
    "main":   NewMainContent(),
    "footer": NewFooter(),
})
```

### Unmounting Components

```go
// Unmount a component
func Unmount(elementID string) error

// Unmount all components
func UnmountAll() error

// Example
err := comps.Unmount("app")
if err != nil {
    logutil.Logf("Failed to unmount: %v", err)
}
```

### Component Registry

```go
// Get mounted component
func GetComponent(elementID string) (Component, bool)

// List all mounted components
func GetAllComponents() map[string]Component

// Check if component is mounted
func IsMounted(elementID string) bool

// Examples
if comp, exists := comps.GetComponent("app"); exists {
    if appComp, ok := comp.(*App); ok {
        appComp.UpdateTitle("New Title")
    }
}

if comps.IsMounted("modal") {
    comps.Unmount("modal")
}
```



## Type Definitions

### Core Types

```go
// Component interface
type Component interface {
    Render() g.Node
    Attach()
}

// Cleanup interface
type CleanupComponent interface {
    Component
    Cleanup()
}

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

### Event Types

```go
// Event handler types
type ClickHandler func()
type InputHandler func(string)
type ChangeHandler func(string)
type SubmitHandler func(dom.Event)
type CustomEventHandler func(dom.Event)

// Event binding options
type EventOptions struct {
    PreventDefault bool
    StopPropagation bool
    Once           bool
    Passive        bool
}
```

### Component Options

```go
// Mount options
type MountOptions struct {
    ReplaceContent bool
    AutoCleanup    bool
    ErrorHandler   func(error)
}

// Render options
type RenderOptions struct {
    Minify     bool
    Indent     string
    Attributes map[string]string
}
```

### Error Types

```go
// UIwGo specific errors
type ComponentError struct {
    Component string
    Operation string
    Err       error
}

func (e *ComponentError) Error() string {
    return fmt.Sprintf("component %s: %s: %v", e.Component, e.Operation, e.Err)
}

// Common error variables
var (
    ErrComponentNotFound = errors.New("component not found")
    ErrElementNotFound   = errors.New("element not found")
    ErrInvalidSelector   = errors.New("invalid selector")
    ErrMountFailed       = errors.New("mount failed")
    ErrAlreadyMounted    = errors.New("already mounted")
)
```

---

## Quick Reference

### Common Patterns

```go
// 1. Basic component
type MyComponent struct {
    data *reactivity.Signal[string]
}

func (c *MyComponent) Render() g.Node {
    return h.Div(
        g.Attr("data-text", "data"),
        g.Text("Loading..."),
    )
}

func (c *MyComponent) Attach() {
    comps.BindText("data", c.data)
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
comps.BindClick("button", func() {
    count.Update(func(c int) int { return c + 1 })
})

// 5. Cleanup
func (c *MyComponent) Cleanup() {
    for _, effect := range c.effects {
        effect.Dispose()
    }
}
```

Next: Explore [Control Flow](../guides/control-flow.md) or check out [Troubleshooting](../troubleshooting.md).