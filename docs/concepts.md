# Core Concepts

UIwGo introduces a unique approach to building web UIs that combines type-safe HTML generation through gomponents with client-side reactivity. This guide explains the fundamental concepts and mental models you need to master UIwGo.

## Table of Contents

- [The UIwGo Philosophy](#the-uiwgo-philosophy)
- [Gomponents-Based Architecture](#gomponents-based-architecture)
- [Signals-Based Reactivity](#signals-based-reactivity)
- [Component Lifecycle](#component-lifecycle)
- [Data Flow Patterns](#data-flow-patterns)
- [Comparison with Other Frameworks](#comparison-with-other-frameworks)
- [Mental Model Shifts](#mental-model-shifts)

## The UIwGo Philosophy

### Core Principles

**1. Gomponents-First, Type-Safe HTML**
- Start with type-safe gomponents structure
- Enhance with reactive behavior through data attributes
- Compile-time HTML validation and Go tooling support

**2. Explicit Over Implicit**
- Clear separation between rendering and behavior
- Explicit binding points marked in HTML
- Predictable update patterns

**3. Performance by Design**
- Fine-grained reactivity (no virtual DOM diffing)
- Compile-time optimizations through Go
- Minimal runtime overhead

**4. Developer Experience**
- Type safety through Go's type system
- Familiar patterns for Go developers
- Hot reload and excellent debugging

### Design Goals

```
Traditional SPA          UIwGo Approach
┌─────────────────┐     ┌─────────────────┐
│ JS-Heavy Client │ ──► │ Gomponents      │
│ Virtual DOM     │     │ Real DOM        │
│ Runtime Diffing │     │ Targeted Updates│
│ Bundle Size     │     │ WASM Efficiency │
└─────────────────┘     └─────────────────┘
```

## Gomponents-Based Architecture

### The Two-Phase Approach

UIwGo components work in two distinct phases:

#### Phase 1: Render (Gomponents Node Generation)

```go
func (c *Counter) Render() g.Node {
    return h.Div(g.Class("counter"),
        h.H2(
            g.Text("Count: "),
            h.Span(
                g.Attr("data-text", "count"),
                g.Text("0"),
            ),
        ),
        h.Button(
            g.Attr("data-click", "increment"),
            g.Text("+"),
        ),
        h.Button(
            g.Attr("data-click", "decrement"),
            g.Text("-"),
        ),
    )
}
```

**Key Points:**
- Generates type-safe HTML through gomponents
- Uses data attributes as "binding markers"
- Compile-time validation and IDE support
- Can be server-side rendered or generated client-side

#### Phase 2: Attach (Reactive Enhancement)

```go
func (c *Counter) Attach() {
    // Scan DOM for markers and bind reactive behavior
    c.BindText("count", c.count)        // data-text="count"
    c.BindClick("increment", c.increment) // data-click="increment"
    c.BindClick("decrement", c.decrement) // data-click="decrement"
}
```

**Key Points:**
- Scans rendered HTML for data attributes
- Establishes reactive connections
- Sets up event listeners
- Creates the "living" component

### Data Attributes as Binding Contract

Data attributes serve as a **contract** between the gomponents-generated HTML structure and Go behavior:

| Attribute Pattern | Purpose | Go Binding |
|------------------|---------|------------|
| `data-text="key"` | Text content | `BindText("key", signal)` |
| `data-html="key"` | HTML content | `BindHTML("key", signal)` |
| `data-click="key"` | Click events | `BindClick("key", handler)` |
| `data-input="key"` | Input binding | `BindInput("key", signal)` |
| `data-show="key"` | Visibility | `BindShow("key", signal)` |
| `data-for="key"` | List rendering | `BindFor("key", signal)` |
| `data-class="key"` | CSS classes | `BindClass("key", signal)` |
| `data-attr="key"` | Attributes | `BindAttr("key", signal)` |

### Benefits of HTML-First

**SEO and Accessibility**
```html
<!-- Semantic HTML structure -->
<article data-show="isVisible">
    <h1 data-text="title">Article Title</h1>
    <p data-text="summary">Article summary...</p>
    <button data-click="readMore">Read More</button>
</article>
```

**Progressive Enhancement**
- Works without JavaScript (basic functionality)
- Enhanced with JavaScript (full interactivity)
- Graceful degradation

**Developer Understanding**
- HTML structure is immediately visible
- Behavior is explicitly marked
- Easy to reason about and debug

## Signals-Based Reactivity

### Understanding Signals

Signals are the foundation of UIwGo's reactivity system:

```go
// Create a signal
count := reactivity.NewSignal(0)

// Read the current value
value := count.Get() // 0

// Update the value (triggers reactivity)
count.Set(42)

// The DOM automatically updates wherever this signal is bound
```

### Signal Types

#### 1. Basic Signals
```go
type Counter struct {
    count *reactivity.Signal[int]
    name  *reactivity.Signal[string]
    items *reactivity.Signal[[]Item]
}

func NewCounter() *Counter {
    return &Counter{
        count: reactivity.NewSignal(0),
        name:  reactivity.NewSignal("Counter"),
        items: reactivity.NewSignal([]Item{}),
    }
}
```

#### 2. Computed Signals (Memos)
```go
type ShoppingCart struct {
    items *reactivity.Signal[[]Item]
    total *reactivity.Memo[float64]
}

func NewShoppingCart() *ShoppingCart {
    cart := &ShoppingCart{
        items: reactivity.NewSignal([]Item{}),
    }
    
    // Memo automatically recomputes when items change
    cart.total = reactivity.NewMemo(func() float64 {
        total := 0.0
        for _, item := range cart.items.Get() {
            total += item.Price * float64(item.Quantity)
        }
        return total
    })
    
    return cart
}
```

#### 3. Effects (Side Effects)
```go
func (c *Counter) Attach() {
    // Effect runs when count changes
    reactivity.NewEffect(func() {
        count := c.count.Get()
        if count > 10 {
            // Side effect: show notification
            showNotification("Count is getting high!")
        }
    })
}
```

### Reactivity Graph

UIwGo automatically tracks dependencies:

```
Signals          Memos           Effects/DOM
┌─────────┐     ┌─────────┐     ┌─────────┐
│ items   │────►│ total   │────►│ DOM     │
│ taxRate │────►│ taxAmt  │────►│ Update  │
│ discount│────►│ final   │────►│         │
└─────────┘     └─────────┘     └─────────┘
```

When `items` changes:
1. `total` automatically recomputes
2. `final` automatically recomputes (depends on `total`)
3. DOM automatically updates
4. All in a single synchronous update cycle

### Fine-Grained Updates

Unlike virtual DOM frameworks, UIwGo updates only what actually changed:

```go
// Only the specific <span data-text="count"> updates
c.count.Set(c.count.Get() + 1)

// Only elements bound to "total" update
c.items.Set(append(c.items.Get(), newItem))
```

## Component Lifecycle

### Component Structure

Every UIwGo component follows this pattern:

```go
type MyComponent struct {
    // 1. State (signals)
    data *reactivity.Signal[string]
    
    // 2. Computed state (memos)
    computed *reactivity.Memo[string]
    
    // 3. Child components (optional)
    child *ChildComponent
}

// 4. Constructor
func NewMyComponent() *MyComponent {
    c := &MyComponent{
        data: reactivity.NewSignal("initial"),
    }
    
    c.computed = reactivity.NewMemo(func() string {
        return "Computed: " + c.data.Get()
    })
    
    return c
}

// 5. Render method (required)
func (c *MyComponent) Render() g.Node {
    return h.Div(
        g.Attr("data-text", "computed"),
        g.Text("Loading..."),
    )
}

// 6. Attach method (required)
func (c *MyComponent) Attach() {
    c.BindText("computed", c.computed)
}

// 7. Cleanup method (optional)
func (c *MyComponent) Cleanup() {
    // Clean up resources, cancel effects, etc.
}
```

### Lifecycle Flow

```
1. Constructor    │ NewMyComponent()
   ↓              │ - Initialize signals
2. Render         │ - Set up memos
   ↓              │ - Create child components
3. Mount          │
   ↓              │ component.Render()
4. Attach         │ - Generate HTML string
   ↓              │ - Insert into DOM
5. Active         │
   ↓              │ component.Attach()
6. Cleanup        │ - Scan for data attributes
                  │ - Bind reactive behavior
                  │ - Set up event listeners
                  │
                  │ [Component is now live]
                  │ - Signals trigger updates
                  │ - User interactions work
                  │ - Effects run
                  │
                  │ component.Cleanup() (optional)
                  │ - Clean up resources
                  │ - Cancel subscriptions
```

### Mounting Components

```go
func main() {
    // Create component instance
    app := NewMyApp()
    
    // Mount to DOM element with id="app"
    comps.Mount("app", app)
    
    // Component is now live and reactive
}
```

## Data Flow Patterns

### Unidirectional Data Flow

UIwGo follows a unidirectional data flow pattern:

```
User Input → Event Handler → Signal Update → DOM Update
    ↑                                            ↓
    └────────── User sees change ←──────────────┘
```

### Parent-Child Communication

#### Props Down
```go
type Parent struct {
    message *reactivity.Signal[string]
    child   *Child
}

func NewParent() *Parent {
    p := &Parent{
        message: reactivity.NewSignal("Hello"),
    }
    
    // Pass signal to child
    p.child = NewChild(p.message)
    return p
}

type Child struct {
    message *reactivity.Signal[string] // Shared signal
}

func NewChild(message *reactivity.Signal[string]) *Child {
    return &Child{message: message}
}
```

#### Events Up
```go
type Child struct {
    onEvent func(data string) // Callback function
}

func (c *Child) Attach() {
    c.BindClick("button", func() {
        if c.onEvent != nil {
            c.onEvent("button clicked")
        }
    })
}

// Parent provides callback
child := NewChild()
child.onEvent = func(data string) {
    parent.handleChildEvent(data)
}
```

### State Management Patterns

#### Local State
```go
type Component struct {
    localState *reactivity.Signal[string]
}
```

#### Shared State
```go
// Global store
var AppStore = struct {
    User     *reactivity.Signal[User]
    Settings *reactivity.Signal[Settings]
}{
    User:     reactivity.NewSignal(User{}),
    Settings: reactivity.NewSignal(Settings{}),
}

// Components access shared state
func (c *UserProfile) Attach() {
    c.BindText("username", AppStore.User)
}
```

#### Context Pattern
```go
type AppContext struct {
    Theme *reactivity.Signal[string]
    User  *reactivity.Signal[User]
}

func NewAppWithContext() *App {
    ctx := &AppContext{
        Theme: reactivity.NewSignal("light"),
        User:  reactivity.NewSignal(User{}),
    }
    
    return &App{
        context: ctx,
        header:  NewHeader(ctx),
        content: NewContent(ctx),
    }
}
```

## Comparison with Other Frameworks

### vs. React

| Aspect | React | UIwGo |
|--------|-------|-------|
| **Rendering** | Virtual DOM | Direct DOM |
| **State** | useState/useReducer | Signals |
| **Updates** | Reconciliation | Fine-grained |
| **Language** | JavaScript/TypeScript | Go |
| **Bundle** | JavaScript bundle | WebAssembly |
| **Learning** | JSX, hooks, lifecycle | HTML + Go patterns |

```jsx
// React
function Counter() {
    const [count, setCount] = useState(0);
    return (
        <div>
            <span>{count}</span>
            <button onClick={() => setCount(count + 1)}>+</button>
        </div>
    );
}
```

```go
// UIwGo
type Counter struct {
    count *reactivity.Signal[int]
}

func (c *Counter) Render() g.Node {
    return h.Div(
        h.Span(
            g.Attr("data-text", "count"),
            g.Text("0"),
        ),
        h.Button(
            g.Attr("data-click", "increment"),
            g.Text("+"),
        ),
    )
}

func (c *Counter) Attach() {
    c.BindText("count", c.count)
    c.BindClick("increment", func() {
        c.count.Set(c.count.Get() + 1)
    })
}
```

### vs. Vue

| Aspect | Vue | UIwGo |
|--------|-----|-------|
| **Templates** | Vue templates | HTML strings |
| **Reactivity** | Proxy-based | Signal-based |
| **Compilation** | Vue compiler | Go compiler |
| **Directives** | v-if, v-for | data-show, data-for |

### vs. Svelte

| Aspect | Svelte | UIwGo |
|--------|--------|-------|
| **Compilation** | Svelte compiler | Go → WASM |
| **Reactivity** | Compiler magic | Explicit signals |
| **Runtime** | Minimal JS | WASM runtime |
| **Syntax** | Svelte syntax | Go + HTML |

### vs. HTMX

| Aspect | HTMX | UIwGo |
|--------|------|-------|
| **Approach** | Server-driven | Client-side reactive |
| **State** | Server state | Client signals |
| **Interactivity** | HTTP requests | Direct DOM updates |
| **Complexity** | Simple | Moderate |

## Mental Model Shifts

### From Virtual DOM to Direct DOM

**Old Mental Model (React/Vue):**
```
State Change → Re-render → Virtual DOM → Diff → DOM Update
```

**New Mental Model (UIwGo):**
```
Signal Change → Direct DOM Update
```

### From Components as Functions to Components as Objects

**React Pattern:**
```jsx
function MyComponent(props) {
    const [state, setState] = useState(initial);
    return <div>{state}</div>;
}
```

**UIwGo Pattern:**
```go
type MyComponent struct {
    state *reactivity.Signal[string]
}

func (c *MyComponent) Render() g.Node {
    return h.Div(
        g.Attr("data-text", "state"),
        g.Text("initial"),
    )
}

func (c *MyComponent) Attach() {
    c.BindText("state", c.state)
}
```

### From JSX to HTML-First

**JSX Mindset:**
- JavaScript expressions in markup
- Dynamic structure through conditionals
- Event handlers as props

**HTML-First Mindset:**
- Static HTML structure with markers
- Dynamic behavior through bindings
- Event handlers bound by name

### From Hooks to Signals

**React Hooks:**
```jsx
const [count, setCount] = useState(0);
const doubled = useMemo(() => count * 2, [count]);
useEffect(() => {
    document.title = `Count: ${count}`;
}, [count]);
```

**UIwGo Signals:**
```go
count := reactivity.NewSignal(0)
doubled := reactivity.NewMemo(func() int {
    return count.Get() * 2
})
reactivity.NewEffect(func() {
    dom.GetWindow().Document().SetTitle(fmt.Sprintf("Count: %d", count.Get()))
})
```

## Key Takeaways

### Mental Model Summary

1. **Think HTML-First**: Start with semantic HTML, enhance with behavior
2. **Signals are State**: All reactive state lives in signals
3. **Data Attributes are Contracts**: They connect HTML to Go behavior
4. **Fine-Grained Updates**: Only what changes gets updated
5. **Explicit Binding**: No magic, clear connection points

### Best Practices

1. **Design HTML structure first**
2. **Use semantic data attribute names**
3. **Keep components focused and small**
4. **Leverage Go's type system**
5. **Think in terms of data flow**

### Common Pitfalls

1. **Over-engineering HTML**: Keep it simple and semantic
2. **Forgetting to bind**: Data attributes without corresponding bindings
3. **Complex render logic**: Move complexity to computed signals
4. **Memory leaks**: Clean up effects and subscriptions

---

Now that you understand the core concepts, you're ready to dive deeper into [Reactivity & State](./guides/reactivity-state.md) or explore [Forms & Events](./guides/forms-events.md).