# UIwGo Action System

**Status:** Implementation Documentation

## 1. Overview

The UIwGo Action System provides a robust, in-memory bus for decoupled communication between components. It is designed for applications needing a structured way to handle user interactions, state changes, and other events without creating tight dependencies between different parts of the UI.

The system is built around a central **Bus** that dispatches **Actions** to interested **Subscribers**. It emphasizes developer experience, predictability, and observability, making it a powerful tool for managing application logic in a Go/WASM environment.

### Core Concepts

-   **Action:** A message describing "what happened," identified by a unique string type (e.g., `"counter.increment"`). It carries a payload and rich metadata for tracing and debugging.
-   **Bus:** The central hub for dispatching actions and managing subscriptions. A global bus is provided by default, but local and scoped buses can be created for isolation.
-   **Subscription:** A link between an action type and a handler function. Subscriptions are lifecycle-aware and can be configured with advanced options like filtering, priorities, and conditional execution.
-   **Query:** A request/response pattern built on top of actions, allowing one part of the application to "ask" for data from another.
-   **Observability:** A suite of built-in tools for debugging and monitoring, including a development logger, an analytics tap, a historical action buffer, and enhanced error handling.

## 2. Core API Surface

The Action System's API is designed to be both flexible and type-safe. While the core bus operates on string-based action types and `interface{}` payloads for maximum flexibility, it provides typed wrappers for a safer, more ergonomic developer experience.

### Main Interfaces and Types

-   `action.Bus`: The primary interface for interacting with the action system.
-   `action.Action[T]`: A struct representing a message, containing a `Type` (string), a `Payload` of type `T`, and metadata (`Time`, `Source`, `TraceID`, `Meta`).
-   `action.ActionType[T]`: A helper type for defining a named action with a specific payload type `T`. Generated via `action.DefineAction[T]("action.name")`.
-   `action.QueryType[Req, Res]`: A helper for defining typed queries. Generated via `action.DefineQuery[Req, Res]("query.name")`.
-   `action.Subscription`: An object returned when subscribing, with `Dispose()` and `IsActive()` methods.

### Key `Bus` Methods

```go
// Dispatches an action to all relevant subscribers.
// Can accept an Action struct or just the action type string.
Dispatch(action any, opts ...DispatchOption) error

// Subscribes a handler to a specific action type (string-based).
Subscribe(actionType string, handler func(Action[string]) error, opts ...SubOption) Subscription

// Subscribes a handler to all actions flowing through the bus.
SubscribeAny(handler func(any) error, opts ...SubOption) Subscription

// Creates a new, isolated bus with a specific name in the hierarchy.
Scope(name string) Bus

// Handles a request/response query.
HandleQuery(queryType string, handler func(Action[string]) (any, error), opts ...QueryOption) Subscription

// Sends a query and waits for a response.
Ask(queryType string, query Action[string], opts ...AskOption) (any, error)

// Registers an enhanced error handler for panics or errors from subscribers.
OnError(handler func(ctx Context, err error, recovered any), opts ...SubOption) Subscription
```

### Typed Wrappers (Helpers)

While the core API is string-based, helper functions provide a fully type-safe experience:

```go
// action.OnAction: Subscribes to a typed action with automatic lifecycle management.
// The subscription is automatically created on mount and disposed on unmount.
action.OnAction(bus, IncrementAction, func(ctx action.Context, payload int) {
    // `payload` is strongly typed as int
})

// action.Dispatch: Dispatches a typed action.
action.Dispatch(bus, IncrementAction, 1) // Dispatches with payload 1

// bus.HandleQueryTyped & bus.AskTyped: For strongly-typed queries.
bus.HandleQueryTyped(FetchUserQuery, func(ctx context.Context, id string) (User, error) {
    // ... logic to fetch user ...
})

future := bus.AskTyped(FetchUserQuery, "user-123")
```

## 3. Subscription System

The subscription system is highly configurable, allowing for fine-grained control over how and when handlers are executed.

### Key Subscription Options (`SubOption`)

-   `WithPriority(int)`: Sets the execution order for subscribers to the same action. Higher priority runs first.
-   `Once()`: The subscription is automatically disposed after the first successful delivery.
-   `Filter(func(payload any) bool)`: Only deliver the action if the predicate returns `true`.
-   `When(reactivity.Signal[bool])`: The subscription is only active when the provided signal is `true`. This is excellent for tying subscriptions to component visibility or application state.
-   `DistinctUntilChanged(equal func(a, b any) bool)`: Prevents delivery if the payload has not changed since the last delivery. An optional equality function can be provided.

**Example: Conditional Subscription**

```go
// This handler only runs when the featureEnabled signal is true.
bus.Subscribe(
    MyAction.Name,
    myHandler,
    action.When(featureEnabled),
)
```

## 4. Observability and Debugging

The Action System includes powerful, built-in tools for development, debugging, and monitoring. These are designed to be low-overhead and easily toggled.

### Features

1.  **Enhanced Error Handling**: The `bus.OnError` handler captures panics and returned errors from any subscriber. It receives a rich `action.Context` object containing the `TraceID`, `Source`, and other metadata from the original dispatch, preserving the full context of the failure.

    ```go
    bus.OnError(func(ctx action.Context, err error, recovered any) {
        logutil.Logf("🚨 ERROR: %v (recovered: %v) TraceID: %s", err, recovered, ctx.TraceID)
    })
    ```

2.  **Development Logger**: `action.EnableDevLogger(bus, handler)` installs a tap that reports every action dispatch, including its type, duration, and the number of subscribers. This is invaluable for understanding the flow of events in the application.

    ```go
to-do-list
- [ ] Implement new feature
- [ ] Implement new feature
- [ ] Implement new feature
- [ ] Implement new feature
- [ ] Implement new feature
- [ ] Implement new feature
- [ ] Implement new feature
- [ ] Implement new feature
- [ ] Implement new feature
- [ ] Implement new feature
- [ ] Implement new feature
- [ ] Implement new feature
```

3.  **Analytics Tap**: `action.NewAnalyticsTap(bus, handler)` provides a way to instrument the application for analytics. It can be filtered to only capture specific actions, making it efficient for sending data to analytics backends.

4.  **Debug Ring Buffer**: `action.EnableDebugRingBuffer(bus, size)` creates a historical buffer of the last `N` actions of each type. This is extremely useful for debugging, as you can inspect the sequence of actions that led to a specific state.

    ```go
    // In a button click handler, for example:
    entries := action.GetDebugRingBufferEntries(bus, IncrementAction.Name)
    logutil.Logf("🔍 Debug buffer for %s:", IncrementAction.Name)
    for i, entry := range entries {
        logutil.Logf("  %d: Payload=%v at %s", i, entry.Payload, entry.Timestamp)
    }
    ```

5.  **Tracing and Metadata**: Every action dispatch can carry a `TraceID`, `Source`, and a `Meta` map. This allows you to follow a causal chain of events across different parts of your application, from UI interaction to final state update.

## 5. Reactive Integration

The Action System is designed to work seamlessly with UIwGo's reactive primitives (`Signal`, `Effect`).

### Action Bridges

-   `action.ToSignal[T](bus, actionType)`: Creates a reactive `reactivity.Signal[T]` that automatically updates its value whenever a matching action is dispatched. This is perfect for displaying the last-seen value of an event stream in the UI.

-   `action.ToStream[T](bus, actionType)`: Creates a reactive stream (channel-based) from an action type. This allows for more complex processing, such as debouncing, throttling, or buffering, using stream operators.

**Example: Action to Signal**

```go
// Define an action for user status changes
var UserStatusChanged = action.DefineAction[string]("user.status_changed")

// Create a signal that reflects the latest user status
userStatusSignal := action.ToSignal[string](bus, UserStatusChanged.Name)

// In a component, you can now react to this signal
comps.BindText(func() string {
    return "Current Status: " + userStatusSignal.Get()
})

// Dispatching an action will automatically update the UI
action.Dispatch(bus, UserStatusChanged, "Online")
```

## 6. Testing

The Action System is designed for testability.

-   **Local Buses**: In tests, create a new local bus (`action.New()`) for each test case to ensure isolation.
-   **Synchronous by Default**: Actions are dispatched synchronously, making tests deterministic and easy to write.
-   **Test Helpers**: The framework provides helpers like `action.TestBus` and `action.FakeClock` to mock time-dependent features like debouncing and throttling.
-   **Observability Hooks**: Use the observability features within tests to make assertions about which actions were fired and in what order.

## 7. Practical Example: Action Lifecycle Demo

The `examples/action_lifecycle_demo` provides a comprehensive example of these features in action.

### Defining and Dispatching a Traced Action

```go
// Define a typed action
var IncrementAction = action.DefineAction[int]("counter.increment")

// Dispatch with rich metadata from a UI event handler
bus.Dispatch(IncrementAction, 1,
    action.WithTraceID(fmt.Sprintf("trace-%d", time.Now().UnixNano())),
    action.WithSource("counter-ui"),
    action.WithMeta(map[string]any{"user": "demo"}),
)
```

### Subscribing with Lifecycle Management and Context

The `action.OnAction` helper ties the subscription's lifecycle to the component it's defined in, automatically disposing it on unmount.

```go
// In a component function:
action.OnAction(bus, IncrementAction, func(ctx action.Context, payload int) {
    // The handler receives the strongly-typed payload and the rich context
    count.Set(count.Get() + payload)
    logutil.Logf("✅ Increment executed. TraceID: %s", ctx.TraceID)
})
```

This ensures that event handlers are automatically cleaned up, preventing memory leaks and unintended side effects.
