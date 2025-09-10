# Action System API Reference

This document provides a comprehensive reference for the public API of the Action System. For conceptual overviews and usage patterns, please refer to [`ACTIONSYSTEM.md`](/action/ACTIONSYSTEM.md).

---

## Table of Contents

1.  [Core Types](#1-core-types)
    -   [`Action[T]`](#actiont)
    -   [`Context`](#context)
    -   [`Bus`](#bus)
    -   [`Subscription`](#subscription)
    -   [`ActionType[T]`](#actiontypet)
    -   [`QueryType[Req, Res]`](#querytypereq-res)
2.  [Bus Methods](#2-bus-methods)
    -   [Dispatch & Subscription](#dispatch--subscription)
    -   [Querying](#querying)
    -   [Lifecycle & Error Handling](#lifecycle--error-handling)
3.  [Helper Functions](#3-helper-functions)
    -   [Typed Wrappers](#typed-wrappers)
    -   [Reactive Bridges](#reactive-bridges)
    -   [Lifecycle Helpers](#lifecycle-helpers)
4.  [Configuration Options](#4-configuration-options)
    -   [`SubOption`](#suboption)
    -   [`DispatchOption`](#dispatchoption)
    -   [`AskOption`](#askoption)
5.  [Observability & Debugging](#5-observability--debugging)
6.  [Testing Utilities](#6-testing-utilities)

---

## 1. Core Types

### `Action[T]`

A generic struct representing a message flowing through the bus.

```go
type Action[T any] struct {
    Type    string
    Payload T
    Time    time.Time
    ID      string
    TraceID string
    Source  string
    Meta    map[string]any
}
```
- **`Type`**: The mandatory string identifier for the action (e.g., `"user.created"`).
- **`Payload`**: The data associated with the action, strongly-typed.
- **`Time`**: The timestamp of when the action was dispatched.
- **`ID`**: A unique identifier for this specific action instance.
- **`TraceID`**: An identifier to correlate a chain of actions.
- **`Source`**: A string indicating the origin of the action (e.g., a component name).
- **`Meta`**: A map for arbitrary metadata.

### `Context`

Provides metadata to subscribers and error handlers about the dispatched action.

```go
type Context struct {
    Bus      Bus
    TraceID  string
    Source   string
    ActionID string
    Meta     map[string]any
}
```

### `Bus`

The central interface for dispatching actions and managing subscriptions.

```go
type Bus interface {
    // See Bus Methods section for details
    Dispatch(action any, opts ...DispatchOption) error
    Subscribe(actionType string, handler func(Action[any]) error, opts ...SubOption) Subscription
    SubscribeAny(handler func(any) error, opts ...SubOption) Subscription
    HandleQuery(queryType string, handler func(Action[any]) (any, error), opts ...SubOption) Subscription
    Ask(queryType string, query Action[any], opts ...AskOption) Future[any]
    OnError(handler func(ctx Context, err error, recovered any), opts ...SubOption) Subscription
    Scope(name string) Bus
    Dispose()
}
```

### `Subscription`

An object representing a handler's registration with the bus.

```go
type Subscription interface {
    Dispose()
    IsActive() bool
}
```
- **`Dispose()`**: Unregisters the handler from the bus.
- **`IsActive()`**: Returns `true` if the subscription is currently active.

### `ActionType[T]`

A type-safe helper for defining an action and its payload type.

```go
// Creation
var UserCreated = action.DefineAction[User]("user.created")

// Usage
action.Dispatch(bus, UserCreated, User{...})
action.OnAction(bus, UserCreated, func(ctx Context, user User) { ... })
```

### `QueryType[Req, Res]`

A type-safe helper for defining a request/response query.

```go
// Creation
var FetchUser = action.DefineQuery[string, User]("user.fetch")

// Usage
bus.HandleQueryTyped(FetchUser, func(ctx Context, id string) (User, error) { ... })
future := bus.AskTyped(FetchUser, "user-123")
```

---

## 2. Bus Methods

### Dispatch & Subscription

- **`Dispatch(action any, opts ...DispatchOption)`**: Sends an action to all subscribers.
- **`Subscribe(actionType string, handler func(Action[any]) error, opts ...SubOption)`**: Registers a handler for a specific action type.
- **`SubscribeAny(handler func(any) error, opts ...SubOption)`**: Registers a handler for all actions.

### Querying

- **`HandleQuery(queryType string, handler func(Action[any]) (any, error), opts ...SubOption)`**: Registers a handler to respond to a query. The handler must return a value and an error.
- **`Ask(queryType string, query Action[any], opts ...AskOption)`**: Sends a query and returns a `Future[any]` that will resolve with the response.
- **`HandleQueryTyped[Req, Res](queryType, handler)`**: Type-safe version of `HandleQuery`.
- **`AskTyped[Req, Res](queryType, payload)`**: Type-safe version of `Ask`.

### Lifecycle & Error Handling

- **`OnError(handler func(ctx Context, err error, recovered any), opts ...SubOption)`**: Registers a global error handler for the bus. It catches panics and returned errors from subscribers.
- **`Scope(name string) Bus`**: Creates a new, isolated child bus.
- **`Dispose()`**: Shuts down the bus and all its subscriptions.

---

## 3. Helper Functions

Located in the `action` package.

### Typed Wrappers

- **`DefineAction[T](name string) ActionType[T]`**: Creates a type-safe action definition.
- **`DefineQuery[Req, Res](name string) QueryType[Req, Res]`**: Creates a type-safe query definition.
- **`Dispatch[T](bus Bus, actionType ActionType[T], payload T, opts ...DispatchOption)`**: Dispatches a type-safe action.

### Reactive Bridges

- **`ToSignal[T](bus Bus, actionType ActionType[T], opts ...SubOption) reactivity.Signal[T]`**: Creates a reactive signal that updates its value with the payload of the most recent action of the given type.
- **`ToStream[T](bus Bus, actionType ActionType[T], capacity int, opts ...SubOption) <-chan T`**: Creates a read-only channel that receives payloads from all actions of the given type.

### Lifecycle Helpers

- **`OnAction[T](bus Bus, actionType ActionType[T], handler func(ctx Context, payload T), opts ...SubOption)`**: A lifecycle-aware subscriber that automatically registers on component mount and disposes on unmount.

---

## 4. Configuration Options

### `SubOption`

Functions that configure a subscription's behavior.

- **`WithPriority(level int)`**: Sets execution order (higher runs first).
- **`Once()`**: Subscription is disposed after one execution.
- **`Filter[T](predicate func(payload T) bool)`**: Only execute if the predicate returns `true`.
- **`When(signal reactivity.Signal[bool])`**: Subscription is only active when the signal is `true`.
- **`DistinctUntilChanged[T](equal func(a, b T) bool)`**: Prevents execution if the payload is unchanged. `equal` is optional.
- **`WithBackpressure(strategy BackpressureStrategy, bufferSize int)`**: (WASM-only) Manages high-frequency actions.
- **`WithScheduler(scheduler Scheduler)`**: (Non-WASM) Assigns a custom scheduler.

### `DispatchOption`

Functions that configure a dispatch operation.

- **`WithTraceID(id string)`**: Sets the trace ID for the action.
- **`WithSource(source string)`**: Sets the source for the action.
- **`WithMeta(m map[string]any)`**: Attaches metadata to the action.
- **`WithAsync()`**: Dispatches the action asynchronously in a new goroutine.

### `AskOption`

Functions that configure a query operation.

- **`WithTimeout(d time.Duration)`**: Sets a timeout for the query.

---

## 5. Observability & Debugging

- **`EnableDevLogger(bus Bus, logger func(msg string))`**: Logs every action dispatch, including type, duration, and subscriber count.
- **`NewAnalyticsTap(bus Bus, handler func(action any))`**: Provides a hook for analytics instrumentation.
- **`EnableDebugRingBuffer(bus Bus, size int)`**: Creates a historical buffer of the last `N` actions for each type.
- **`GetDebugRingBufferEntries(bus Bus, actionType string) []DebugEntry`**: Retrieves the buffered entries for an action type.

---

## 6. Testing Utilities

Refer to [`action/TESTING.md`](/action/TESTING.md) for a detailed guide. Key components include:

- **`TestBus`**: An isolated bus for testing with an integrated `FakeClock`.
- **`FakeClock`**: A controllable time source for deterministic testing of time-based logic.
- **`TestFuture`**: An enhanced `Future` with synchronous `Await` and timeout methods.
- **`MockSubscriber`**: A mock for verifying subscriber interactions.
- **`ActionsEqual(a, b any) bool`**: Compares two actions for equality, ignoring timestamps.