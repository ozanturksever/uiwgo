# Core Concepts

UIwGo introduces a unique approach to building web UIs that combines type-safe HTML generation through gomponents with client-side reactivity. This guide explains the fundamental concepts and mental models you need to master UIwGo.

## Table of Contents

- [The UIwGo Philosophy](#the-uiwgo-philosophy)
- [Gomponents-Based Architecture](#gomponents-based-architecture)
- [Signals-Based Reactivity](#signals-based-reactivity)
- [Action System Architecture](#action-system-architecture)
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

**5. Decoupled Communication**
- Event-driven architecture through actions
- Type-safe message passing
- Built-in observability and tracing

### Design Goals

```
Traditional SPA          UIwGo Approach
┌─────────────────┐     ┌─────────────────┐
│ JS-Heavy Client │ ──► │ Gomponents      │
│ Virtual DOM     │     │ Real DOM        │
│ Runtime Diffing │     │ Targeted Updates│
│ Bundle Size     │     │ WASM Efficiency │
│ Props Drilling  │     │ Action Bus      │
└─────────────────┘     └─────────────────┘
```

The action system adds another layer to this philosophy: **structured communication**. Rather than tightly coupling components through direct calls or prop drilling, UIwGo encourages event-driven patterns where components communicate through typed actions on a central bus.

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

#### Phase 2: Mount Effects (Reactive Enhancement)

```go
func Counter() g.Node {
    count := reactivity.CreateSignal(0)
    
    // Set up mount effects for DOM interactions
    comps.OnMount(func() {
        // Set up any DOM interactions that need direct element access
        if btn := dom.GetElementByID("special-btn"); btn != nil {
            dom.BindClickToCallback(btn, func() {
                // Handle special interactions
            })
        }
    })
    
    return Div(
        // Reactive text binding
        comps.BindText(func() string {
            return fmt.Sprintf("Count: %d", count.Get())
        }),
        // Inline event handlers
        Button(
            Text("Increment"),
            dom.OnClickInline(func(el dom.Element) {
                count.Set(count.Get() + 1)
            }),
        ),
        Button(
            Text("Decrement"),
            dom.OnClickInline(func(el dom.Element) {
                count.Set(count.Get() - 1)
            }),
        ),
    )
}
```

**Key Points:**
- Uses reactive bindings for dynamic content
- Inline event handlers for user interactions
- Mount effects for DOM setup when needed
- Functional, declarative component structure

### Reactive Bindings as Component Contract

Reactive bindings establish the **contract** between static HTML structure and dynamic behavior:

| Binding Type | Purpose | Usage Pattern |
|--------------|---------|---------------|
| `comps.BindText()` | Dynamic text content | `comps.BindText(func() string { return signal.Get() })` |
| `comps.BindHTML()` | Dynamic HTML content | `comps.BindHTML(func() string { return htmlSignal.Get() })` |
| `dom.OnClickInline()` | Click events | `dom.OnClickInline(func(el dom.Element) { ... })` |
| `dom.OnInputInline()` | Input handling | `dom.OnInputInline(func(el dom.Element) { ... })` |
| `comps.BindShow()` | Conditional visibility | `comps.BindShow(func() bool { return signal.Get() })` |
| `comps.BindFor()` | List rendering | `comps.BindFor(items, func(item T) g.Node { ... })` |
| `comps.BindClass()` | Dynamic CSS classes | `comps.BindClass("active", func() bool { return isActive.Get() })` |
| `comps.BindAttr()` | Dynamic attributes | `comps.BindAttr("src", func() string { return url.Get() })` |

### Benefits of Functional Components

**SEO and Accessibility**
```go
func Article(title, summary string, isVisible *reactivity.Signal[bool]) g.Node {
    return g.If(isVisible.Get(),
        g.El("article",
            g.El("h1", g.Text(title)),
            g.El("p", comps.BindText(func() string { return summary })),
            g.El("button",
                g.Text("Read More"),
                dom.OnClickInline(func(el dom.Element) {
                    // Handle read more action
                }),
            ),
        ),
    )
}
```

**Reactive Enhancement**
- Type-safe component definitions
- Automatic reactivity through signals
- Declarative event handling
- Compile-time validation

**Developer Understanding**
- HTML structure is immediately visible
- Behavior is explicitly marked
- Easy to reason about and debug

## Signals-Based Reactivity

### Understanding Signals

Signals are the foundation of UIwGo's reactivity system:

```go
// Create a signal
count := reactivity.CreateSignal(0)

// Read the current value
value := count.Get() // 0

// Update the value (triggers reactivity)
count.Set(42)

// The DOM automatically updates wherever this signal is bound
```

### Signal Types

#### 1. Basic Signals
```go
func Counter() g.Node {
    // Create signals within the component function
    count := reactivity.CreateSignal(0)
    name := reactivity.CreateSignal("Counter")
    items := reactivity.CreateSignal([]Item{})
    
    return Div(
        comps.BindText(func() string {
            return fmt.Sprintf("%s: %d", name.Get(), count.Get())
        }),
        // ... rest of component
    )
}
```

#### 2. Computed Signals (Memos)
```go
func ShoppingCart() g.Node {
    items := reactivity.CreateSignal([]Item{})
    
    // Memo automatically recomputes when items change
    total := reactivity.CreateMemo(func() float64 {
        total := 0.0
        for _, item := range items.Get() {
            total += item.Price * float64(item.Quantity)
        }
        return total
    })
    
    return Div(
        comps.BindText(func() string {
            return fmt.Sprintf("Total: $%.2f", total.Get())
        }),
        // ... rest of component
    )
}
```

#### 3. Effects (Side Effects)
```go
func Counter() g.Node {
    count := reactivity.CreateSignal(0)
    
    comps.OnMount(func() {
        // Effect runs when count changes
        effect := reactivity.CreateEffect(func() {
            if count.Get() > 10 {
                // Side effect: show notification
                showNotification("Count is getting high!")
            }
        })
        
        // Clean up effect when component unmounts
        comps.OnCleanup(func() {
            effect.Dispose()
        })
    })
    
    return Div(
        // ... component content
    )
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

## Action System Architecture

### Understanding Actions

Actions provide a structured way for components to communicate without tight coupling. Think of actions as **typed messages** that describe "what happened" in your application:

```go
// Define typed action types
var (
    IncrementAction    = action.DefineAction[int]("counter.increment")
    UserLoginAction    = action.DefineAction[User]("auth.user_login")
    NotificationAction = action.DefineAction[string]("ui.notification")
)
```

### The Action Bus

The **Bus** is the central hub that routes actions to interested subscribers:

```go
// Get the global bus (or create a local one)
bus := action.Global()

// Dispatch an action
action.Dispatch(bus, IncrementAction, 5)

// Subscribe to actions with lifecycle management
action.OnAction(bus, IncrementAction, func(ctx action.Context, payload int) {
    count.Set(count.Get() + payload)
    logutil.Logf("Incremented by %d", payload)
})
```

### Action-Signal Bridge

Actions integrate seamlessly with the reactive system. You can convert action streams into signals:

```go
type User struct {
    ID   int    `json:"id"`
    Name string `json:"name"`
}

type UserProfile struct {
    // Signal automatically updates when UserLoginAction is dispatched
    currentUser *reactivity.Signal[User]
}

func NewUserProfile(bus action.Bus) *UserProfile {
    return &UserProfile{
        currentUser: action.UseActionSignal(bus, UserLoginAction),
    }
}

func (up *UserProfile) Render() g.Node {
    return h.Div(
        h.H2(g.Text("Welcome!")),
        h.P(
            g.Attr("data-text", "username"),
            g.Text("Loading..."),
        ),
    )
}

func (up *UserProfile) Attach() {
    // Automatically updates when currentUser signal changes
    up.BindText("username", reactivity.CreateMemo(func() string {
        user := up.currentUser.Get()
        if user.Name == "" {
            return "Not logged in"
        }
        return user.Name
    }))
}
```

### Practical Example: Todo App with Actions

```go
// Define action types for todo operations
var (
    AddTodoAction    = action.DefineAction[string]("todo.add")
    ToggleTodoAction = action.DefineAction[int]("todo.toggle")
    DeleteTodoAction = action.DefineAction[int]("todo.delete")
)

type Todo struct {
    ID   int    `json:"id"`
    Text string `json:"text"`
    Done bool   `json:"done"`
}

type TodoApp struct {
    bus   action.Bus
    todos *reactivity.Signal[[]Todo]
}

func NewTodoApp() *TodoApp {
    bus := action.New()
    app := &TodoApp{
        bus:   bus,
        todos: reactivity.CreateSignal([]Todo{}),
    }
    
    // Set up action handlers with automatic lifecycle management
    action.OnAction(bus, AddTodoAction, func(ctx action.Context, text string) {
        current := app.todos.Get()
        newTodo := Todo{
            ID:   len(current) + 1,
            Text: text,
            Done: false,
        }
        app.todos.Set(append(current, newTodo))
    })
    
    action.OnAction(bus, ToggleTodoAction, func(ctx action.Context, id int) {
        current := app.todos.Get()
        for i, todo := range current {
            if todo.ID == id {
                current[i].Done = !current[i].Done
                break
            }
        }
        app.todos.Set(current)
    })
    
    return app
}

func (app *TodoApp) Render() g.Node {
    return h.Div(
        h.H1(g.Text("Todo List")),
        
        // Add todo form
        h.Form(
            h.Input(
                g.Attr("data-input", "todoText"),
                g.Attr("placeholder", "Add a todo..."),
            ),
            h.Button(
                g.Attr("data-click", "addTodo"),
                g.Text("Add"),
            ),
        ),
        
        // Todo list
        h.Ul(g.Attr("data-for", "todos")),
    )
}

func (app *TodoApp) Attach() {
    todoText := reactivity.CreateSignal("")
    
    app.BindInput("todoText", todoText)
    
    app.BindClick("addTodo", func() {
        text := todoText.Get()
        if text != "" {
            // Dispatch action instead of direct manipulation
            action.Dispatch(app.bus, AddTodoAction, text)
            todoText.Set("")
        }
    })
    
    app.BindFor("todos", app.todos, func(todo Todo, index int) g.Node {
        return h.Li(
            h.Input(
                g.Attr("type", "checkbox"),
                g.Attr("data-click", fmt.Sprintf("toggle-%d", todo.ID)),
                g.If(todo.Done, g.Attr("checked", "true")),
            ),
            h.Span(g.Text(todo.Text)),
            h.Button(
                g.Attr("data-click", fmt.Sprintf("delete-%d", todo.ID)),
                g.Text("Delete"),
            ),
        )
    })
    
    // Dynamic event binding for todo items
    app.BindDynamic("toggle-", func(id string) {
        todoID, _ := strconv.Atoi(id)
        action.Dispatch(app.bus, ToggleTodoAction, todoID)
    })
    
    app.BindDynamic("delete-", func(id string) {
        todoID, _ := strconv.Atoi(id)
        action.Dispatch(app.bus, DeleteTodoAction, todoID)
    })
}
```

### Action System Benefits

#### 1. **Decoupling**
```go
// Components don't need to know about each other
// UserList dispatches an action when user is selected
action.Dispatch(bus, UserSelectedAction, user.ID)

// UserDetails subscribes and reacts independently
action.OnAction(bus, UserSelectedAction, func(ctx action.Context, userID int) {
    // Load user details
})
```

#### 2. **Tracing and Debugging**
```go
// Actions carry rich metadata for debugging
bus.Dispatch(action.Action[string]{
    Type:    NotificationAction.Name,
    Payload: "User saved successfully",
    TraceID: "trace-12345",
    Source:  "user-form",
    Meta:    map[string]any{"userID": 42},
})

// Enhanced error handling with full context
bus.OnError(func(ctx action.Context, err error, recovered any) {
    logutil.Logf("Error in action %s from %s (TraceID: %s): %v",
        ctx.ActionType, ctx.Source, ctx.TraceID, err)
})
```

#### 3. **Testing**
```go
func TestTodoActions(t *testing.T) {
    // Create isolated bus for testing
    bus := action.New()
    app := NewTodoApp(bus)
    
    // Test action dispatch
    action.Dispatch(bus, AddTodoAction, "Test todo")
    
    // Verify state change
    todos := app.todos.Get()
    assert.Equal(t, 1, len(todos))
    assert.Equal(t, "Test todo", todos[0].Text)
}
```

### Advanced Features

#### Query Pattern
For request-response communication:

```go
var FetchUserQuery = action.DefineQuery[int, User]("user.fetch")

// Handler
bus.HandleQueryTyped(FetchUserQuery, func(ctx action.Context, userID int) (User, error) {
    return userService.GetUser(userID)
})

// Caller
user, err := bus.AskTyped(FetchUserQuery, 42)
```

#### Scoped Buses
For component isolation:

```go
// Create a scoped bus for a modal dialog
modalBus := bus.Scope("user-modal")

// Actions in this scope don't affect the parent
action.Dispatch(modalBus, CloseModalAction, nil)
```

#### Observability
Built-in development tools:

```go
// Enable development logger
action.EnableDevLogger(bus, func(entry action.DevLogEntry) {
    logutil.Logf("[%s] %s -> %d subscribers",
        entry.Timestamp.Format("15:04:05"),
        entry.ActionType,
        entry.SubscriberCount)
})

// Enable debug ring buffer
action.EnableDebugRingBuffer(bus, 50)

// View action history
entries := action.GetDebugRingBufferEntries(bus, IncrementAction.Name)
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
        data: reactivity.CreateSignal("initial"),
    }
    
    c.computed = reactivity.CreateMemo(func() string {
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

UIwGo supports multiple data flow patterns:

#### 1. Signal-Based Flow
```
User Input → Event Handler → Signal Update → DOM Update
    ↑                                            ↓
    └────────── User sees change ←──────────────┘
```

#### 2. Action-Based Flow
```
User Input → Action Dispatch → Action Handlers → Signal Updates → DOM Update
    ↑                                                                ↓
    └──────────────────── User sees change ←──────────────────────┘
```

#### 3. Hybrid Flow (Recommended)
```
User Input → Action Dispatch → Multiple Handlers → Signals/Effects → DOM Update
    ↑             ↓                    ↓               ↓              ↓
    └─────── Analytics ──── Logging ─── State ─── Side Effects ──────┘
```

### Component Communication Patterns

#### 1. Direct Signal Sharing (Simple Cases)
```go
type Parent struct {
    message *reactivity.Signal[string]
    child   *Child
}

func NewParent() *Parent {
    p := &Parent{
        message: reactivity.CreateSignal("Hello"),
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

#### 2. Action-Based Communication (Recommended)
```go
// Define domain actions
var (
    UserSelectedAction = action.DefineAction[User]("app.user_selected")
    NotificationAction = action.DefineAction[string]("ui.notification")
)

type UserList struct {
    bus action.Bus
}

func (ul *UserList) Attach() {
    ul.BindClick("user-item", func(userID string) {
        user := ul.findUser(userID)
        // Dispatch action - any component can listen
        action.Dispatch(ul.bus, UserSelectedAction, user)
    })
}

type UserDetails struct {
    bus         action.Bus
    selectedUser *reactivity.Signal[User]
}

func NewUserDetails(bus action.Bus) *UserDetails {
    ud := &UserDetails{
        bus: bus,
        selectedUser: reactivity.CreateSignal(User{}),
    }
    
    // Subscribe to user selection events
    action.OnAction(bus, UserSelectedAction, func(ctx action.Context, user User) {
        ud.selectedUser.Set(user)
        
        // Show notification
        action.Dispatch(bus, NotificationAction, "User details loaded")
    })
    
    return ud
}
```

#### 3. Event Callbacks (Legacy/Integration)
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
    User:     reactivity.CreateSignal(User{}),
    Settings: reactivity.CreateSignal(Settings{}),
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
        Theme: reactivity.CreateSignal("light"),
        User:  reactivity.CreateSignal(User{}),
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
| **Communication** | Props/Context/Redux | Action Bus |
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
// UIwGo - Signal-based approach
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

```go
// UIwGo - Action-based approach
var IncrementAction = action.DefineAction[int]("counter.increment")

type Counter struct {
    bus   action.Bus
    count *reactivity.Signal[int]
}

func (c *Counter) Attach() {
    // Subscribe to increment actions
    action.OnAction(c.bus, IncrementAction, func(ctx action.Context, amount int) {
        c.count.Set(c.count.Get() + amount)
    })
    
    c.BindText("count", c.count)
    c.BindClick("increment", func() {
        // Dispatch action instead of direct state change
        action.Dispatch(c.bus, IncrementAction, 1)
    })
}
```

### vs. Vue

| Aspect | Vue | UIwGo |
|--------|-----|-------|
| **Templates** | Vue templates | HTML strings |
| **Reactivity** | Proxy-based | Signal-based |
| **Communication** | Events/Vuex/Pinia | Action Bus |
| **Compilation** | Vue compiler | Go compiler |
| **Directives** | v-if, v-for | data-show, data-for |

### vs. Svelte

| Aspect | Svelte | UIwGo |
|--------|--------|-------|
| **Compilation** | Svelte compiler | Go → WASM |
| **Reactivity** | Compiler magic | Explicit signals |
| **Communication** | Events/Stores | Action Bus |
| **Runtime** | Minimal JS | WASM runtime |
| **Syntax** | Svelte syntax | Go + HTML |

### vs. HTMX

| Aspect | HTMX | UIwGo |
|--------|------|-------|
| **Approach** | Server-driven | Client-side reactive |
| **State** | Server state | Client signals |
| **Communication** | HTTP/SSE | Action Bus |
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
count := reactivity.CreateSignal(0)
doubled := reactivity.CreateMemo(func() int {
    return count.Get() * 2
})
reactivity.CreateEffect(func() {
    dom.GetWindow().Document().SetTitle(fmt.Sprintf("Count: %d", count.Get()))
})
```

### From Props Drilling to Action Bus

**React Props Drilling:**
```jsx
function App() {
    const [user, setUser] = useState(null);
    
    return (
        <Layout user={user}>
            <Header user={user} onLogout={() => setUser(null)} />
            <Content user={user} />
            <Sidebar user={user} />
        </Layout>
    );
}

// Every component needs user prop even if it doesn't use it
function Layout({ user, children }) {
    return <div>{children}</div>;
}
```

**UIwGo Action Bus:**
```go
// Define actions once
var (
    UserLoginAction  = action.DefineAction[User]("auth.login")
    UserLogoutAction = action.DefineAction[struct{}]("auth.logout")
)

// Components subscribe to what they need
type Header struct {
    bus  action.Bus
    user *reactivity.Signal[User]
}

func NewHeader(bus action.Bus) *Header {
    h := &Header{
        bus:  bus,
        user: action.UseActionSignal(bus, UserLoginAction),
    }
    
    // Subscribe to logout events
    action.OnAction(bus, UserLogoutAction, func(ctx action.Context, _ struct{}) {
        h.user.Set(User{})
    })
    
    return h
}

// No props needed - components get what they need from the bus
func (h *Header) Attach() {
    h.BindText("username", h.user)
    h.BindClick("logout", func() {
        action.Dispatch(h.bus, UserLogoutAction, struct{}{})
    })
}
```

### From Event Callbacks to Typed Actions

**Traditional Event Handling:**
```go
type Modal struct {
    onClose   func()
    onSubmit  func(data FormData)
    onCancel  func()
}

func NewModal(onClose, onSubmit, onCancel func(...)) *Modal {
    // Callback hell - hard to track, test, and debug
}
```

**Action-Based Events:**
```go
// Clear, discoverable event types
var (
    ModalCloseAction  = action.DefineAction[string]("modal.close")
    ModalSubmitAction = action.DefineAction[FormData]("modal.submit")
    ModalCancelAction = action.DefineAction[string]("modal.cancel")
)

type Modal struct {
    bus action.Bus
}

func NewModal(bus action.Bus) *Modal {
    return &Modal{bus: bus}
}

func (m *Modal) Attach() {
    m.BindClick("close", func() {
        // Discoverable, traceable, testable
        action.Dispatch(m.bus, ModalCloseAction, "user-clicked-x")
    })
    
    m.BindClick("submit", func() {
        data := m.collectFormData()
        action.Dispatch(m.bus, ModalSubmitAction, data)
    })
}

// Other components can listen without tight coupling
action.OnAction(bus, ModalSubmitAction, func(ctx action.Context, data FormData) {
    // Save data
    saveFormData(data)
    
    // Show notification
    action.Dispatch(bus, NotificationAction, "Data saved successfully")
    
    // Analytics
    action.Dispatch(bus, AnalyticsAction, "form_submitted")
})
```

### From Redux Complexity to Action Simplicity

**Redux Pattern:**
```jsx
// Action creators
const increment = (amount) => ({
    type: 'INCREMENT',
    payload: amount
});

// Reducers
const counterReducer = (state = { count: 0 }, action) => {
    switch (action.type) {
        case 'INCREMENT':
            return { ...state, count: state.count + action.payload };
        default:
            return state;
    }
};

// Store setup
const store = createStore(counterReducer);

// Component connection
const mapStateToProps = (state) => ({ count: state.count });
const mapDispatchToProps = { increment };
export default connect(mapStateToProps, mapDispatchToProps)(Counter);
```

**UIwGo Action Pattern:**
```go
// Define action type
var IncrementAction = action.DefineAction[int]("counter.increment")

// Component with direct state and action handling
type Counter struct {
    bus   action.Bus
    count *reactivity.Signal[int]
}

func NewCounter(bus action.Bus) *Counter {
    c := &Counter{
        bus:   bus,
        count: reactivity.CreateSignal(0),
    }
    
    // Direct action handling - no boilerplate
    action.OnAction(bus, IncrementAction, func(ctx action.Context, amount int) {
        c.count.Set(c.count.Get() + amount)
    })
    
    return c
}

func (c *Counter) Attach() {
    c.BindText("count", c.count)
    c.BindClick("increment", func() {
        action.Dispatch(c.bus, IncrementAction, 1)
    })
}
```

## Key Takeaways

### Mental Model Summary

1. **Think HTML-First**: Start with semantic HTML, enhance with behavior
2. **Signals are State**: All reactive state lives in signals
3. **Actions are Events**: Use typed actions for all communication
4. **Data Attributes are Contracts**: They connect HTML to Go behavior
5. **Fine-Grained Updates**: Only what changes gets updated
6. **Explicit Binding**: No magic, clear connection points
7. **Bus for Communication**: Decouple components through the action bus

### Best Practices

1. **Design HTML structure first**
2. **Use semantic data attribute names**
3. **Keep components focused and small**
4. **Leverage Go's type system**
5. **Think in terms of data flow**
6. **Use actions for component communication**
7. **Define action types at package level**
8. **Enable observability for debugging**
9. **Use scoped buses for component isolation**
10. **Leverage action metadata for tracing**

### Common Pitfalls

1. **Over-engineering HTML**: Keep it simple and semantic
2. **Forgetting to bind**: Data attributes without corresponding bindings
3. **Complex render logic**: Move complexity to computed signals
4. **Memory leaks**: Clean up effects and subscriptions
5. **Action type conflicts**: Use descriptive, namespaced action names
6. **Overusing global bus**: Create scoped buses when appropriate
7. **Ignoring action context**: Use TraceID and metadata for debugging
8. **Missing action handlers**: Actions without subscribers fail silently

---

Now that you understand the core concepts, you're ready to dive deeper into:
- [Action System](../action/API_REFERENCE.md) - Complete action system documentation
- [Reactivity & State](./guides/reactivity-state.md) - Advanced reactive patterns
- [Forms & Events](./guides/forms-events.md) - User interaction handling
- [Performance Optimization](./guides/performance-optimization.md) - Best practices for performance