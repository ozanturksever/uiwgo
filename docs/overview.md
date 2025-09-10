# Overview

UIwGo is a browser-first UI runtime for Go that targets WebAssembly, providing fine-grained reactivity and a declarative component model for building modern web applications.

## Value Proposition

**Write interactive web UIs in Go.** UIwGo brings Go's type safety, tooling, and ecosystem to frontend development while maintaining the performance and user experience expectations of modern web applications.

### Key Benefits

- **Type-Safe HTML** - Use gomponents for compile-time HTML validation and Go's type system
- **No Template Languages** - Write HTML as Go code with full IDE support and refactoring
- **Familiar Tooling** - Use Go's testing, debugging, and development tools for frontend code
- **Performance** - WebAssembly execution with minimal DOM updates through fine-grained reactivity
- **Ecosystem Integration** - Bridge to React/shadcn/ui components when needed
- **Developer Experience** - Hot reload, comprehensive testing, and Make-based workflows

## Target Use Cases

### Single Page Applications (SPAs)
Build complete web applications with client-side routing, state management, and complex user interactions.

### Interactive Components
Add reactive behavior to existing pages or create reusable interactive widgets.

### Data-Heavy Interfaces
Leverage Go's excellent data processing capabilities for dashboards, admin panels, and analytical tools.

### Progressive Enhancement
Start with server-rendered HTML and progressively enhance with client-side interactivity.

## Key Features at a Glance

### 🎯 Fine-Grained Reactivity
```go
// Signals automatically track dependencies
count := reactivity.NewSignal(0)
effect := reactivity.NewEffect(func() {
    fmt.Printf("Count is: %d\n", count.Get())
})
count.Set(1) // Effect automatically re-runs
```

### 🏗️ Functional Components with gomponents
```go
import (
    "fmt"
    "github.com/ozanturksever/uiwgo/comps"
    "github.com/ozanturksever/uiwgo/dom"
    "github.com/ozanturksever/uiwgo/reactivity"
    . "maragu.dev/gomponents"
    . "maragu.dev/gomponents/html"
)

// A component is a Go function that returns a gomponents.Node.
func Counter() Node {
    count := reactivity.NewSignal(0)

    // Event handlers are set up in OnMount, after the DOM exists.
    comps.OnMount(func() {
        incrementBtn := dom.GetElementByID("increment-btn")
        dom.BindClickToCallback(incrementBtn, func() {
            count.Set(count.Get() + 1)
        })
    })

    // The UI tree is returned using type-safe gomponents.
    return Div(
        Span(comps.BindText(func() string { return fmt.Sprintf("%d", count.Get()) })),
        Button(ID("increment-btn"), Text("+")),
    )
}
```

### 🧭 Client-Side Routing
```go
// Define routes with parameters and nesting
router.Route("/users/:id", UserDetailView)
router.Route("/users/:id/posts/*", UserPostsView)

// Navigate programmatically
router.Push("/users/123")
```

### 🚌 Action System for Decoupled Communication
```go
// Define typed actions for different events
var UserCreated = action.DefineAction[User]("user.created")
var NotificationShown = action.DefineAction[string]("notification.shown")

// Create a global bus for component communication
bus := action.New()

// Components subscribe to actions with lifecycle management
action.OnAction(bus, UserCreated, func(ctx action.Context, user User) {
    // Handler automatically disposed on component unmount
    logutil.Logf("New user: %s (TraceID: %s)", user.Name, ctx.TraceID)
})

// Dispatch actions from anywhere in the application
action.Dispatch(bus, UserCreated, User{Name: "Alice"})

// Bridge actions to reactive signals for UI updates
userSignal := action.ToSignal(bus, UserCreated)
```

### ⚛️ React Compatibility
```go
// Use React/shadcn/ui components as leaf widgets
react.Button(react.Props{
    "variant": "primary",
    "onClick": "handleClick",
}, "Click me")
```

### 🧪 Comprehensive Testing
```go
// Browser E2E tests with real user interactions
func TestCounter(t *testing.T) {
    chromedp.Run(ctx,
        chromedp.Navigate("http://localhost:8080"),
        chromedp.Click("button"),
        chromedp.WaitVisible("span[data-text='1']"),
    )
}
```

## Mental Model Comparison

### Traditional Frontend Frameworks
```javascript
// Virtual DOM approach
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

### UIwGo Approach
```go
// Gomponents + fine-grained reactivity
import (
    "fmt"
    "github.com/ozanturksever/uiwgo/comps"
    "github.com/ozanturksever/uiwgo/dom"
    "github.com/ozanturksever/uiwgo/reactivity"
    . "maragu.dev/gomponents"
    . "maragu.dev/gomponents/html"
)

func Counter() Node {
    count := reactivity.NewSignal(0)

    comps.OnMount(func() {
        incrementBtn := dom.GetElementByID("increment-btn")
        dom.BindClickToCallback(incrementBtn, func() {
            count.Set(count.Get() + 1)
        })
    })

    return Div(
        Span(comps.BindText(func() string { return fmt.Sprintf("%d", count.Get()) })),
        Button(ID("increment-btn"), Text("+")),
    )
}
```

### Key Differences

| Aspect | Traditional | UIwGo |
|--------|-------------|-------|
| **Rendering** | Virtual DOM diffing | Direct DOM updates |
| **Reactivity** | Component re-renders | Fine-grained signals |
| **Language** | JavaScript/TypeScript | Go + WebAssembly |
| **Mental Model** | Functional components | Gomponents + binding |
| **Performance** | Reconciliation overhead | Minimal update targeting |
| **Communication** | Props/callbacks/context | Action system + signals |

## Architecture Overview

```
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│   Go Components │    │   Reactivity     │    │   DOM Binding   │
│                 │    │                  │    │                 │
│ • Render HTML   │───▶│ • Signals        │───▶│ • Attach Phase  │
│ • Define State  │    │ • Effects        │    │ • Event Binding │
│ • Handle Events │    │ • Memos          │    │ • Text Updates  │
└─────────────────┘    │ • Cleanup        │    │ • Lifecycle     │
         │              └──────────────────┘    └─────────────────┘
         ▼                       │                       │
┌─────────────────┐              │                       ▼
│   Action System │              │              ┌──────────────────┐
│                 │              │              │   Browser DOM    │
│ • Event Bus     │──────────────┼─────────────▶│                  │
│ • Typed Actions │              │              │ • Minimal Updates│
│ • Subscriptions │              ▼              │ • Event Handling │
│ • Observability │     ┌──────────────────┐    │ • User Interaction│
└─────────────────┘     │   Cross-Component│    └──────────────────┘
                        │   Communication  │
                        │                  │
                        │ • Decoupled      │
                        │ • Traceable      │
                        │ • Debuggable     │
                        └──────────────────┘
```

## When to Choose UIwGo

### ✅ Great Fit
- **Go Teams** - Leverage existing Go expertise for frontend development
- **Type Safety Priority** - Need compile-time guarantees for UI logic
- **Performance Critical** - Require minimal DOM updates and efficient reactivity
- **Complex State** - Benefit from Go's data structures and algorithms
- **Testing Focus** - Want comprehensive unit and E2E testing workflows

### ⚠️ Consider Alternatives
- **Large Teams** - Existing JavaScript/React expertise might be more valuable
- **Rapid Prototyping** - JavaScript ecosystem might offer faster iteration
- **SEO Critical** - Server-side rendering requirements (though UIwGo can complement SSR)
- **Bundle Size** - WebAssembly overhead might not suit very simple interactions

## Next Steps

### Quick Start
1. **[Getting Started](./getting-started.md)** - Set up your first UIwGo project
2. **[Concepts](./concepts.md)** - Understand the core mental models
3. **[Core APIs](./api/core-apis.md)** - Complete API documentation

### Deep Dive Guides
1. **[Reactivity & State](./guides/reactivity-state.md)** - Master signals, effects, and memos
2. **[Control Flow](./guides/control-flow.md)** - Learn conditional rendering and lists
3. **[Forms & Events](./guides/forms-events.md)** - Handle user input and events
4. **[Lifecycle & Effects](./guides/lifecycle-effects.md)** - Component lifecycle management
5. **[Application Manager](./guides/application-manager.md)** - Application lifecycle and management

---

**Ready to build?** Start with the [Getting Started guide](./getting-started.md) to create your first interactive component in minutes.