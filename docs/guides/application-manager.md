# Application Manager

The Application Manager provides a centralized system for managing application lifecycle, state, and navigation in UIwGo applications. It handles component mounting/unmounting, state persistence, routing integration, and lifecycle hooks.

## Table of Contents

- [Overview](#overview)
- [Core Concepts](#core-concepts)
- [Getting Started](#getting-started)
- [Lifecycle Management](#lifecycle-management)
- [State Management](#state-management)
- [Navigation and Routing](#navigation-and-routing)
- [Hooks and Events](#hooks-and-events)
- [Advanced Usage](#advanced-usage)
- [Best Practices](#best-practices)

## Overview

The Application Manager serves as the central orchestrator for UIwGo applications, providing:

- **Lifecycle Management**: Automatic component initialization, mounting, and cleanup
- **State Persistence**: Cross-navigation state management and restoration
- **Router Integration**: Seamless integration with UIwGo's routing system
- **Event System**: Lifecycle hooks for custom application logic
- **Resource Management**: Automatic cleanup and memory management

### Key Benefits

- Simplified application structure and organization
- Automatic state management across navigation
- Consistent lifecycle patterns
- Memory leak prevention through automatic cleanup
- Extensible hook system for custom behavior

## Core Concepts

### Application States

The Application Manager tracks several key states:

```go
type LifecycleState int

const (
    StateUninitialized LifecycleState = iota
    StateInitializing
    StateReady
    StateNavigating
    StateCleaningUp
    StateDestroyed
)
```

### Lifecycle Events

Events are triggered at key points in the application lifecycle:

```go
type LifecycleEvent int

const (
    EventBeforeInit LifecycleEvent = iota
    EventAfterInit
    EventBeforeMount
    EventAfterMount
    EventBeforeNavigate
    EventAfterNavigate
    EventBeforeUnmount
    EventAfterUnmount
    EventBeforeCleanup
    EventAfterCleanup
)
```

### State Management

The Application Manager maintains three types of state:

- **UI State**: Component-specific reactive state
- **Router State**: Current route, parameters, and navigation history
- **App State**: Global application state and configuration

## Getting Started

### Basic Setup

```go
package main

import (
    "github.com/ozanturksever/uiwgo/appmanager"
    "github.com/ozanturksever/uiwgo/router"
)

func main() {
    // Create application configuration
    config := appmanager.DefaultAppConfig()
    config.EnableStateRestore = true
    config.EnableAutoCleanup = true
    
    // Initialize the application manager
    manager := appmanager.NewAppManager(config)
    
    // Initialize the application
    if err := manager.Initialize(); err != nil {
        panic(err)
    }
    
    // Set up routes
    setupRoutes(manager)
    
    // Start the application
    manager.Navigate("/")
}
```

### Route Configuration

```go
func setupRoutes(manager *appmanager.AppManager) {
    // Add lifecycle hooks
    manager.AddHook(appmanager.EventBeforeMount, func(ctx *appmanager.LifecycleContext) {
        logutil.Log("Mounting component:", ctx.Route)
    })
    
    manager.AddHook(appmanager.EventAfterMount, func(ctx *appmanager.LifecycleContext) {
        logutil.Log("Component mounted successfully")
    })
    
    // Configure routes with the router
    // The AppManager automatically integrates with the router
}
```

## Lifecycle Management

### Component Lifecycle

The Application Manager automatically manages component lifecycles:

1. **Initialization**: Component creation and setup
2. **Mounting**: DOM attachment and reactive binding
3. **Active**: Normal operation and state updates
4. **Unmounting**: Cleanup and DOM detachment
5. **Destruction**: Final cleanup and memory release

### Lifecycle Hooks

Register hooks to execute custom logic at specific lifecycle points:

```go
// Before component initialization
manager.AddHook(appmanager.EventBeforeInit, func(ctx *appmanager.LifecycleContext) {
    // Prepare component data
    loadComponentData(ctx.Route)
})

// After component mounting
manager.AddHook(appmanager.EventAfterMount, func(ctx *appmanager.LifecycleContext) {
    // Start timers, subscribe to events, etc.
    startPeriodicUpdates()
})

// Before component unmounting
manager.AddHook(appmanager.EventBeforeUnmount, func(ctx *appmanager.LifecycleContext) {
    // Save state, stop timers, etc.
    saveComponentState()
    stopPeriodicUpdates()
})
```

### Automatic Cleanup

The Application Manager automatically handles:

- Signal subscription cleanup
- Event listener removal
- Timer and interval cancellation
- Memory leak prevention

## State Management

### UI State Persistence

The Application Manager can automatically save and restore component state:

```go
type MyComponent struct {
    data *reactivity.Signal[string]
    count *reactivity.Signal[int]
}

func (c *MyComponent) GetState() map[string]interface{} {
    return map[string]interface{}{
        "data": c.data.Get(),
        "count": c.count.Get(),
    }
}

func (c *MyComponent) SetState(state map[string]interface{}) {
    if data, ok := state["data"].(string); ok {
        c.data.Set(data)
    }
    if count, ok := state["count"].(int); ok {
        c.count.Set(count)
    }
}
```

### Global State Access

```go
// Access global application state
appState := manager.GetAppState()

// Update global state
manager.UpdateAppState(func(state *appmanager.AppState) {
    state.UserPreferences["theme"] = "dark"
})
```

## Navigation and Routing

### Programmatic Navigation

```go
// Navigate to a route
manager.Navigate("/users/123")

// Navigate with state preservation
manager.NavigateWithState("/settings", map[string]interface{}{
    "returnTo": "/users/123",
})
```

### Navigation Hooks

```go
// Before navigation
manager.AddHook(appmanager.EventBeforeNavigate, func(ctx *appmanager.LifecycleContext) {
    // Validate navigation, save current state, etc.
    if !validateNavigation(ctx.Route) {
        ctx.Cancel() // Prevent navigation
    }
})

// After navigation
manager.AddHook(appmanager.EventAfterNavigate, func(ctx *appmanager.LifecycleContext) {
    // Update analytics, restore state, etc.
    trackPageView(ctx.Route)
})
```

## Hooks and Events

### Hook Types

The Application Manager supports various hook types:

```go
// Lifecycle hooks
manager.AddHook(appmanager.EventAfterMount, myMountHook)

// Navigation hooks
manager.AddHook(appmanager.EventBeforeNavigate, myNavigationHook)

// Cleanup hooks
manager.AddHook(appmanager.EventBeforeCleanup, myCleanupHook)
```

### Hook Context

Hooks receive a context with relevant information:

```go
func myHook(ctx *appmanager.LifecycleContext) {
    // Access current route
    route := ctx.Route
    
    // Access route parameters
    params := ctx.Params
    
    // Access application state
    appState := ctx.AppState
    
    // Cancel the operation (for "before" hooks)
    if someCondition {
        ctx.Cancel()
    }
}
```

## Advanced Usage

### Custom Configuration

```go
config := &appmanager.AppConfig{
    EnableStateRestore: true,
    EnableAutoCleanup:  true,
    StateStorageKey:    "myapp_state",
    MaxStateHistory:    10,
    CleanupTimeout:     5 * time.Second,
}

manager := appmanager.NewAppManager(config)
```

### Error Handling

```go
// Add error handling hooks
manager.AddHook(appmanager.EventAfterMount, func(ctx *appmanager.LifecycleContext) {
    if ctx.Error != nil {
        logutil.Log("Mount error:", ctx.Error)
        // Handle the error
    }
})
```

### Performance Monitoring

```go
manager.AddHook(appmanager.EventBeforeMount, func(ctx *appmanager.LifecycleContext) {
    ctx.StartTime = time.Now()
})

manager.AddHook(appmanager.EventAfterMount, func(ctx *appmanager.LifecycleContext) {
    duration := time.Since(ctx.StartTime)
    logutil.Logf("Component mounted in %v", duration)
})
```

## Best Practices

### 1. Use Lifecycle Hooks Appropriately

```go
// ✅ Good: Use hooks for cross-cutting concerns
manager.AddHook(appmanager.EventAfterMount, func(ctx *appmanager.LifecycleContext) {
    analytics.TrackPageView(ctx.Route)
})

// ❌ Avoid: Component-specific logic in global hooks
manager.AddHook(appmanager.EventAfterMount, func(ctx *appmanager.LifecycleContext) {
    if ctx.Route == "/specific-page" {
        // This should be in the component itself
    }
})
```

### 2. Implement State Interfaces

```go
// ✅ Good: Implement state management interfaces
type MyComponent struct {
    // ... fields
}

func (c *MyComponent) GetState() map[string]interface{} {
    // Return serializable state
}

func (c *MyComponent) SetState(state map[string]interface{}) {
    // Restore state from serialized data
}
```

### 3. Handle Cleanup Properly

```go
// ✅ Good: Clean up resources in hooks
manager.AddHook(appmanager.EventBeforeUnmount, func(ctx *appmanager.LifecycleContext) {
    // Cancel timers, close connections, etc.
    cancelPeriodicUpdates()
    closeWebSocketConnection()
})
```

### 4. Use Error Boundaries

```go
// ✅ Good: Handle errors gracefully
manager.AddHook(appmanager.EventAfterMount, func(ctx *appmanager.LifecycleContext) {
    if ctx.Error != nil {
        // Log error and show fallback UI
        showErrorPage(ctx.Error)
    }
})
```

### 5. Optimize State Storage

```go
// ✅ Good: Only store necessary state
func (c *MyComponent) GetState() map[string]interface{} {
    return map[string]interface{}{
        "userInput": c.userInput.Get(),
        "selectedTab": c.selectedTab.Get(),
        // Don't store computed values or large objects
    }
}
```

## Integration Examples

### With Router

```go
// The Application Manager integrates seamlessly with the router
router.HandleFunc("/users/{id}", func(w http.ResponseWriter, r *http.Request) {
    // The manager handles component lifecycle automatically
    userComponent := NewUserComponent()
    manager.MountComponent(userComponent)
})
```

### With Reactive State

```go
type AppComponent struct {
    theme *reactivity.Signal[string]
    user  *reactivity.Signal[*User]
}

// State is automatically managed by the Application Manager
func (c *AppComponent) Attach() {
    // Bindings are automatically cleaned up
    comps.BindText("username", c.user)
    comps.BindClass("theme", c.theme)
}
```

The Application Manager provides a robust foundation for building scalable UIwGo applications with proper lifecycle management, state persistence, and clean architecture patterns.