# Golid V2 API Reference

## 🎯 Complete API Documentation

This comprehensive reference documents all APIs in Golid V2's SolidJS-inspired reactivity system, including reactive primitives, DOM manipulation, event handling, and performance monitoring.

---

## 📋 Table of Contents

1. [Core Reactive Primitives](#core-reactive-primitives)
2. [DOM Manipulation](#dom-manipulation)
3. [Event System](#event-system)
4. [Lifecycle Management](#lifecycle-management)
5. [Error Handling](#error-handling)
6. [Performance Monitoring](#performance-monitoring)
7. [Utilities and Helpers](#utilities-and-helpers)
8. [Advanced Patterns](#advanced-patterns)

---

## 🔧 Core Reactive Primitives

### Signals

#### `CreateSignal[T](initialValue T) (getter func() T, setter func(T))`

Creates a reactive signal with automatic dependency tracking.

```go
// Basic usage
count, setCount := CreateSignal(0)
name, setName := CreateSignal("John")

// Reading triggers dependency tracking
value := count() // Automatically tracked in effects

// Setting triggers reactive updates
setCount(42) // All dependent effects will re-run
```

**Parameters:**
- `initialValue T`: Initial value of the signal

**Returns:**
- `getter func() T`: Function to read the current value
- `setter func(T)`: Function to update the value

**Features:**
- Automatic dependency tracking
- Batched updates for performance
- Type-safe with Go generics
- Memory efficient (150B per signal)

#### `CreateSignalWithCompare[T](initialValue T, compareFn func(prev, next T) bool) (getter func() T, setter func(T))`

Creates a signal with custom equality comparison.

```go
// Custom comparison for complex types
user, setUser := CreateSignalWithCompare(User{}, func(prev, next User) bool {
    return prev.ID == next.ID && prev.Name == next.Name
})

// Only triggers updates when comparison returns false
setUser(User{ID: 1, Name: "John"}) // Triggers update
setUser(User{ID: 1, Name: "John"}) // No update (same values)
```

### Effects

#### `CreateEffect(fn func(), owner *Owner) *Computation`

Creates a reactive computation that automatically re-runs when dependencies change.

```go
count, setCount := CreateSignal(0)

// Effect automatically tracks count dependency
effect := CreateEffect(func() {
    fmt.Printf("Count is: %d\n", count())
}, nil)

setCount(1) // Prints: "Count is: 1"
setCount(2) // Prints: "Count is: 2"
```

**Parameters:**
- `fn func()`: Function to execute reactively
- `owner *Owner`: Owner for cleanup (nil for current owner)

**Returns:**
- `*Computation`: Computation handle for manual control

**Features:**
- Automatic dependency tracking
- Batched execution
- Automatic cleanup on owner disposal
- Cascade prevention

#### `CreateMemo[T](fn func() T, owner *Owner) func() T`

Creates a memoized computation that caches results until dependencies change.

```go
count, setCount := CreateSignal(0)

// Expensive computation cached until count changes
expensiveValue := CreateMemo(func() int {
    fmt.Println("Computing expensive value...")
    return count() * count() * count()
}, nil)

result1 := expensiveValue() // Computes and caches
result2 := expensiveValue() // Returns cached value
setCount(1)
result3 := expensiveValue() // Recomputes due to dependency change
```

**Parameters:**
- `fn func() T`: Function to compute the memoized value
- `owner *Owner`: Owner for cleanup (nil for current owner)

**Returns:**
- `func() T`: Getter function for the memoized value

### Ownership and Cleanup

#### `CreateOwner(fn func()) *Owner`

Creates a new ownership scope for automatic resource management.

```go
owner := CreateOwner(func() {
    // All reactive primitives created here are automatically cleaned up
    count, setCount := CreateSignal(0)
    
    CreateEffect(func() {
        fmt.Printf("Count: %d\n", count())
    }, nil)
    
    // Resources automatically cleaned up when owner is disposed
})

// Manually dispose if needed
owner.Dispose()
```

**Parameters:**
- `fn func()`: Function to execute within the owner scope

**Returns:**
- `*Owner`: Owner handle for manual disposal

#### `OnCleanup(fn func())`

Registers a cleanup function to run when the current owner is disposed.

```go
CreateOwner(func() {
    resource := acquireResource()
    
    OnCleanup(func() {
        resource.Release()
        fmt.Println("Resource cleaned up")
    })
})
```

#### `OnMount(fn func())`

Registers a function to run when the component is mounted.

```go
CreateOwner(func() {
    OnMount(func() {
        fmt.Println("Component mounted")
        // Initialize component
    })
})
```

---

## 🌐 DOM Manipulation

### Reactive DOM Bindings

#### `BindTextReactive(element js.Value, fn func() string) *DOMBinding`

Reactively binds text content to a DOM element.

```go
element := document.Call("createElement", "div")
text, setText := CreateSignal("Hello")

// Text automatically updates when signal changes
BindTextReactive(element, func() string {
    return text()
})

setText("World") // Element text automatically updates
```

#### `BindAttributeReactive(element js.Value, attr string, fn func() string) *DOMBinding`

Reactively binds an attribute to a DOM element.

```go
button := document.Call("createElement", "button")
disabled, setDisabled := CreateSignal(false)

// Attribute automatically updates
BindAttributeReactive(button, "disabled", func() string {
    if disabled() {
        return "true"
    }
    return ""
})
```

#### `BindStyleReactive(element js.Value, property string, fn func() string) *DOMBinding`

Reactively binds a CSS style property.

```go
element := document.Call("createElement", "div")
color, setColor := CreateSignal("red")

BindStyleReactive(element, "color", func() string {
    return color()
})

setColor("blue") // Style automatically updates
```

#### `BindClassReactive(element js.Value, className string, fn func() bool) *DOMBinding`

Reactively toggles a CSS class.

```go
element := document.Call("createElement", "div")
active, setActive := CreateSignal(false)

BindClassReactive(element, "active", func() bool {
    return active()
})

setActive(true) // Class "active" added
setActive(false) // Class "active" removed
```

### Direct DOM Operations

#### `BindReactive(element js.Value, fn func()) *DOMBinding`

Creates a general reactive binding for custom DOM operations.

```go
element := document.Call("createElement", "canvas")
width, setWidth := CreateSignal(100)
height, setHeight := CreateSignal(100)

BindReactive(element, func() {
    element.Set("width", width())
    element.Set("height", height())
    // Custom drawing logic
})
```

---

## 🎪 Event System

### Reactive Event Handling

#### `SubscribeReactive(element js.Value, event string, handler func(js.Value), options ...EventOptions) func()`

Subscribes to DOM events with reactive integration and automatic cleanup.

```go
button := document.Call("createElement", "button")
count, setCount := CreateSignal(0)

// Automatic cleanup and batching
unsubscribe := SubscribeReactive(button, "click", func(e js.Value) {
    setCount(count() + 1)
}, EventOptions{
    Delegate: true,
    Priority: UserBlocking,
})

// Manual unsubscribe if needed (automatic with owner cleanup)
defer unsubscribe()
```

**Parameters:**
- `element js.Value`: DOM element to attach event to
- `event string`: Event type (e.g., "click", "input")
- `handler func(js.Value)`: Event handler function
- `options ...EventOptions`: Optional event configuration

**Returns:**
- `func()`: Unsubscribe function

#### `EventOptions`

Configuration for event subscriptions.

```go
type EventOptions struct {
    Delegate    bool     // Use event delegation
    Priority    Priority // Event processing priority
    Passive     bool     // Passive event listener
    Once        bool     // Remove after first trigger
    Capture     bool     // Capture phase
}
```

### Convenience Event Handlers

#### `OnClickReactive(fn func()) Node`

Reactive click event handler for components.

```go
button := OnClickReactive(func() {
    fmt.Println("Button clicked!")
})
```

#### `OnInputReactive(fn func(string)) Node`

Reactive input event handler.

```go
input := OnInputReactive(func(value string) {
    fmt.Printf("Input value: %s\n", value)
})
```

### Event System Management

#### `GetEventManager() *EventManager`

Returns the global event manager for advanced configuration.

```go
manager := GetEventManager()
manager.SetBatchSize(100)
manager.SetFlushTimeout(16 * time.Millisecond)
```

---

## 🔄 Lifecycle Management

### Component Lifecycle

#### `CreateComponent[P any](fn func(props P) js.Value) func(P) *Component`

Creates a reusable component with props.

```go
Counter := CreateComponent(func(props struct{ Initial int }) js.Value {
    count, setCount := CreateSignal(props.Initial)
    
    element := document.Call("createElement", "div")
    
    BindTextReactive(element, func() string {
        return fmt.Sprintf("Count: %d", count())
    })
    
    return element
})

// Use component
counterElement := Counter(struct{ Initial int }{Initial: 5})
```

#### Lifecycle Hooks

```go
CreateOwner(func() {
    OnMount(func() {
        // Component mounted
    })
    
    OnCleanup(func() {
        // Component unmounting
    })
})
```

### Resource Management

#### `CreateResource[T](fetcher func() (T, error), options ...ResourceOptions) (getter func() T, refetch func())`

Creates a reactive resource for async data.

```go
user, refetchUser := CreateResource(func() (User, error) {
    return fetchUserFromAPI()
}, ResourceOptions{
    Cache: true,
    Timeout: 5 * time.Second,
})

// Use in effects
CreateEffect(func() {
    currentUser := user()
    fmt.Printf("User: %+v\n", currentUser)
}, nil)

// Manually refetch
refetchUser()
```

---

## 🛡️ Error Handling

### Error Boundaries

#### `CreateErrorBoundary(fn func(), errorHandler func(error)) *ErrorBoundary`

Creates an error boundary to catch and handle errors in reactive code.

```go
boundary := CreateErrorBoundary(func() {
    // Protected reactive code
    count, setCount := CreateSignal(0)
    
    CreateEffect(func() {
        if count() < 0 {
            panic("Count cannot be negative")
        }
    }, nil)
}, func(err error) {
    fmt.Printf("Error caught: %v\n", err)
    // Handle error gracefully
})
```

#### `RecoverFromError(err error) bool`

Attempts to recover from a reactive system error.

```go
if err := someReactiveOperation(); err != nil {
    if RecoverFromError(err) {
        fmt.Println("Successfully recovered")
    } else {
        fmt.Println("Recovery failed")
    }
}
```

### Error Monitoring

#### `EnableErrorMonitoring()`

Enables global error monitoring and reporting.

```go
EnableErrorMonitoring()

// Errors are automatically tracked and reported
```

---

## 📊 Performance Monitoring

### Performance Monitoring

#### `EnablePerformanceMonitoring()`

Enables comprehensive performance monitoring.

```go
EnablePerformanceMonitoring()

// Monitor performance in real-time
metrics := GetPerformanceMetrics()
fmt.Printf("Signal updates: %d\n", metrics.SignalUpdates)
```

#### `GetPerformanceMetrics() PerformanceMetrics`

Returns current performance metrics.

```go
type PerformanceMetrics struct {
    SignalUpdates       uint64        `json:"signal_updates"`
    SignalUpdateLatency time.Duration `json:"signal_update_latency"`
    DOMUpdates          uint64        `json:"dom_updates"`
    DOMUpdateLatency    time.Duration `json:"dom_update_latency"`
    MemoryUsage         uint64        `json:"memory_usage"`
    ErrorCount          uint64        `json:"error_count"`
    // ... more metrics
}
```

#### `AddPerformanceAlertHandler(handler PerformanceAlertHandler)`

Adds a handler for performance alerts.

```go
AddPerformanceAlertHandler(func(alert PerformanceAlert) {
    if alert.Severity == "error" {
        fmt.Printf("Performance Alert: %s\n", alert.Message)
        // Send to monitoring system
    }
})
```

### Profiling

#### `EnablePerformanceProfiling()`

Enables detailed performance profiling.

```go
EnablePerformanceProfiling()

// Get detailed profiles
profiler := GetPerformanceMonitor().profiler
signalProfiles := profiler.GetSignalProfiles()
```

---

## 🔧 Utilities and Helpers

### Scheduler Control

#### `FlushScheduler()`

Manually flushes all pending reactive updates.

```go
count, setCount := CreateSignal(0)
setCount(1)
setCount(2)
setCount(3)

FlushScheduler() // All updates processed in batch
```

#### `getScheduler() *Scheduler`

Returns the global scheduler for advanced control.

```go
scheduler := getScheduler()
stats := scheduler.GetStats()
fmt.Printf("Queue size: %d\n", stats.QueueSize)
```

### Context Management

#### `getCurrentOwner() *Owner`

Returns the current owner context.

```go
owner := getCurrentOwner()
if owner != nil {
    fmt.Println("Inside owner context")
}
```

#### `getCurrentComputation() *Computation`

Returns the current computation context.

```go
comp := getCurrentComputation()
if comp != nil {
    fmt.Println("Inside reactive computation")
}
```

### Testing Utilities

#### `ResetReactiveContext()`

Resets the reactive context for testing.

```go
func TestMyComponent(t *testing.T) {
    ResetReactiveContext()
    
    // Test reactive code
}
```

#### `ResetScheduler()`

Resets the scheduler for testing.

```go
func TestScheduler(t *testing.T) {
    ResetScheduler()
    
    // Test scheduling behavior
}
```

---

## 🎨 Advanced Patterns

### Derived State

```go
// Create derived state from multiple signals
firstName, setFirstName := CreateSignal("John")
lastName, setLastName := CreateSignal("Doe")

fullName := CreateMemo(func() string {
    return firstName() + " " + lastName()
}, nil)

// Use derived state
CreateEffect(func() {
    fmt.Printf("Full name: %s\n", fullName())
}, nil)
```

### Conditional Effects

```go
enabled, setEnabled := CreateSignal(true)
count, setCount := CreateSignal(0)

// Effect only runs when enabled
CreateEffect(func() {
    if enabled() {
        fmt.Printf("Count: %d\n", count())
    }
}, nil)
```

### Batched Updates

```go
count, setCount := CreateSignal(0)
name, setName := CreateSignal("")

// Batch multiple updates
getScheduler().batch(func() {
    setCount(42)
    setName("John")
    // Both updates processed together
})
```

### Custom Reactive Primitives

```go
// Create a custom store
func CreateStore[T any](initialState T) (getter func() T, setter func(func(T) T)) {
    state, setState := CreateSignal(initialState)
    
    return state, func(updater func(T) T) {
        setState(updater(state()))
    }
}

// Usage
state, updateState := CreateStore(AppState{})

updateState(func(current AppState) AppState {
    return AppState{
        Count: current.Count + 1,
        Name:  current.Name,
    }
})
```

### Reactive Collections

```go
// Reactive array operations
items, setItems := CreateSignal([]string{})

addItem := func(item string) {
    setItems(append(items(), item))
}

removeItem := func(index int) {
    current := items()
    if index >= 0 && index < len(current) {
        setItems(append(current[:index], current[index+1:]...))
    }
}

// Reactive rendering
CreateEffect(func() {
    for i, item := range items() {
        fmt.Printf("%d: %s\n", i, item)
    }
}, nil)
```

---

## 🔍 Type Definitions

### Core Types

```go
// Signal types
type ReactiveSignal[T any] struct {
    // Internal implementation
}

// Computation types
type Computation struct {
    id           uint64
    fn           func()
    dependencies []*ReactiveSignal[any]
    state        ComputationState
    owner        *Owner
}

type ComputationState int
const (
    Clean ComputationState = iota
    Check
    Dirty
)

// Owner types
type Owner struct {
    id       uint64
    children []*Owner
    cleanups []func()
    disposed bool
}

// Priority types
type Priority int
const (
    UserBlocking Priority = iota
    Normal
    Idle
)
```

### DOM Types

```go
type DOMBinding struct {
    id         uint64
    element    js.Value
    property   string
    computation *Computation
    cleanup    func()
    owner      *Owner
}
```

### Event Types

```go
type EventOptions struct {
    Delegate bool
    Priority Priority
    Passive  bool
    Once     bool
    Capture  bool
}

type EventSubscription struct {
    id        uint64
    event     string
    element   js.Value
    handler   func(js.Value)
    cleanup   func()
    autoClean bool
    owner     *Owner
}
```

---

## 📚 Best Practices

### 1. Always Use Owners

```go
// ✅ Good - Automatic cleanup
CreateOwner(func() {
    count, setCount := CreateSignal(0)
    // Automatically cleaned up
})

// ❌ Avoid - Manual cleanup required
count, setCount := CreateSignal(0)
// Must manually clean up
```

### 2. Prefer Memos for Expensive Computations

```go
// ✅ Good - Cached computation
expensiveValue := CreateMemo(func() int {
    return heavyComputation(input())
}, nil)

// ❌ Avoid - Recomputes every time
CreateEffect(func() {
    result := heavyComputation(input()) // Expensive!
    // Use result
}, nil)
```

### 3. Use Reactive DOM Bindings

```go
// ✅ Good - Reactive binding
BindTextReactive(element, func() string {
    return text()
})

// ❌ Avoid - Manual updates
CreateEffect(func() {
    element.Set("textContent", text())
}, nil)
```

### 4. Batch Related Updates

```go
// ✅ Good - Batched updates
getScheduler().batch(func() {
    setCount(count() + 1)
    setName("Updated")
    setActive(true)
})

// ❌ Avoid - Separate updates
setCount(count() + 1)
setName("Updated")
setActive(true)
```

---

## 🚀 Migration from V1

### Signal Migration

```go
// V1
signal := NewSignal(0)
value := signal.Get()
signal.Set(42)

// V2
getter, setter := CreateSignal(0)
value := getter()
setter(42)
```

### Effect Migration

```go
// V1
unsubscribe := signal.Subscribe(func(value int) {
    // Handle change
})
defer unsubscribe()

// V2
CreateEffect(func() {
    value := getter()
    // Handle change - automatic cleanup
}, nil)
```

### Component Migration

```go
// V1
type Component struct {
    signal *Signal[int]
}

func (c *Component) Mount() {
    c.signal = NewSignal(0)
}

func (c *Component) Unmount() {
    c.signal.Dispose()
}

// V2
func CreateComponent() js.Value {
    return CreateOwner(func() js.Value {
        count, setCount := CreateSignal(0)
        // Automatic cleanup
        return element
    })
}
```

---

## 📖 Examples

See the [examples](./examples/) directory for complete working examples:

- [Counter](./examples/counter/main.go) - Basic signal usage
- [Todo List](./examples/todo/main.go) - Complex state management
- [Conditional Rendering](./examples/conditional/main.go) - Dynamic UI
- [Router](./examples/router/main.go) - Navigation and routing
- [Lifecycle](./examples/lifecycle/main.go) - Component lifecycle

---

## 🔗 Related Documentation

- [Migration Guide](./GOLID_V2_MIGRATION_GUIDE.md) - Complete migration instructions
- [Performance Report](./PERFORMANCE_COMPARISON_REPORT.md) - Performance analysis
- [Deployment Guide](./PRODUCTION_DEPLOYMENT_GUIDE.md) - Production deployment
- [Architecture Overview](./golid_reactivity_architecture.md) - System architecture

---

**API Reference Version**: Golid V2.0.0  
**Last Updated**: 2025-08-18  
**Status**: ✅ Complete and Production Ready