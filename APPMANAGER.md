# WASM Application Manager Design & Implementation Plan

## Overview

This document outlines the design and implementation of a comprehensive WASM Application Manager that streamlines app lifecycle management, state persistence, and component orchestration for Go WebAssembly applications using the existing reactive framework, bridge system, and router.

## Architecture Goals

1. **Unified Lifecycle Management**: Centralized control over application initialization, mounting, routing, and cleanup
2. **State Persistence**: Hydration/dehydration hooks for seamless state management across sessions
3. **Bridge Integration**: Seamless integration with the existing bridge system for DOM/JS interop
4. **Router Integration**: First-class router support with navigation lifecycle hooks
5. **Developer Experience**: Simple, intuitive API that reduces boilerplate while maintaining flexibility
6. **Test-Driven Development**: Comprehensive test coverage with clear testing patterns

## Core Components

### 1. Application Manager (`AppManager`)

The central orchestrator that manages the entire application lifecycle.

```go
type AppManager struct {
    config       *AppConfig
    bridge       bridge.Manager
    router       *router.Router
    store        *AppStore
    lifecycle    *LifecycleManager
    persistence  *PersistenceManager
    initialized  bool
    running      bool
    cleanupScope *reactivity.CleanupScope
}

type AppConfig struct {
    AppID           string
    MountElementID  string
    Routes          []*router.RouteDefinition
    InitialState    interface{}
    PersistenceKey  string
    EnableRouter    bool
    EnablePersistence bool
    Timeout         time.Duration
    OnReady         func(*AppManager) error
    OnError         func(error)
}
```

### 2. Lifecycle Manager (`LifecycleManager`)

Manages application lifecycle events and hooks.

```go
type LifecycleManager struct {
    hooks map[LifecycleEvent][]LifecycleHook
    state LifecycleState
}

type LifecycleEvent string

const (
    EventBeforeInit    LifecycleEvent = "beforeInit"
    EventAfterInit     LifecycleEvent = "afterInit"
    EventBeforeMount   LifecycleEvent = "beforeMount"
    EventAfterMount    LifecycleEvent = "afterMount"
    EventBeforeRoute   LifecycleEvent = "beforeRoute"
    EventAfterRoute    LifecycleEvent = "afterRoute"
    EventBeforeUnmount LifecycleEvent = "beforeUnmount"
    EventAfterUnmount  LifecycleEvent = "afterUnmount"
    EventAppReady      LifecycleEvent = "appReady"
    EventError         LifecycleEvent = "error"
)

type LifecycleHook func(ctx *LifecycleContext) error

type LifecycleContext struct {
    Event     LifecycleEvent
    Manager   *AppManager
    Data      interface{}
    Cancel    context.CancelFunc
}
```

### 3. Persistence Manager (`PersistenceManager`)

Handles state hydration and dehydration using the reactive framework.

```go
type PersistenceManager struct {
    key       string
    store     *AppStore
    enabled   bool
    hydrators map[string]HydratorFunc
}

type HydratorFunc func(data interface{}) error
type DehydratorFunc func() (interface{}, error)

type PersistenceConfig struct {
    Key         string
    AutoSave    bool
    SaveInterval time.Duration
    Hydrators   map[string]HydratorFunc
    Dehydrators map[string]DehydratorFunc
}
```

### 4. Application Store (`AppStore`)

Centralized reactive state management with persistence hooks.

```go
type AppStore struct {
    state        reactivity.Store[AppState]
    setState     func(...interface{})
    persistence  *PersistenceManager
    subscribers  map[string][]StateSubscriber
}

type AppState struct {
    Router      RouterState
    User        interface{}
    UI          UIState
    Custom      map[string]interface{}
}

type StateSubscriber func(state AppState, path string)
```

## Implementation Plan

### Phase 1: Core Infrastructure (TDD)

#### 1.1 Basic Application Manager Structure

**Test File**: `appmanager/manager_test.go`

```go
func TestAppManager_Creation(t *testing.T) {
    config := &AppConfig{
        AppID:          "test-app",
        MountElementID: "app",
        InitialState: AppState{
            User: map[string]interface{}{"name": "test"},
            UI:   UIState{Theme: "light"},
        },
    }

    manager := NewAppManager(config)

    if manager == nil {
        t.Fatal("NewAppManager returned nil")
    }
    if manager.config.AppID != "test-app" {
        t.Errorf("Expected AppID 'test-app', got '%s'", manager.config.AppID)
    }
    if manager.initialized.Get() {
        t.Error("Expected initialized to be false")
    }
    if manager.running.Get() {
        t.Error("Expected running to be false")
    }

    // Test reactive state
    state := manager.store.Get()
    if state.User.(map[string]interface{})["name"] != "test" {
        t.Error("Initial state not properly set")
    }
}

func TestAppManager_Initialize(t *testing.T) {
    tests := []struct {
        name        string
        config      *AppConfig
        expectError bool
        errorMsg    string
    }{
        {
            name: "successful_initialization",
            config: &AppConfig{
                AppID:          "test-app",
                MountElementID: "app",
                InitialState:   AppState{},
            },
            expectError: false,
        },
        {
            name: "double_initialization",
            config: &AppConfig{
                AppID:          "test-app",
                MountElementID: "app",
                InitialState:   AppState{},
            },
            expectError: true,
            errorMsg:    "already initialized",
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Use mockdom for testing
            cleanup := mockdom.SetupMockDOM()
            defer cleanup()

            manager := NewAppManager(tt.config)

            // First initialization
            ctx := context.Background()
            err := manager.Initialize(ctx)

            if tt.name == "double_initialization" {
                // Try to initialize again
                err = manager.Initialize(ctx)
            }

            if tt.expectError {
                if err == nil {
                    t.Error("Expected error but got none")
                }
                if tt.errorMsg != "" && !strings.Contains(err.Error(), tt.errorMsg) {
                    t.Errorf("Expected error containing '%s', got '%s'", tt.errorMsg, err.Error())
                }
            } else {
                if err != nil {
                    t.Errorf("Unexpected error: %v", err)
                }
                if !manager.initialized.Get() {
                    t.Error("Expected manager to be initialized")
                }
                if manager.lifecycle.GetState() != LifecycleStateInitialized {
                    t.Error("Expected lifecycle state to be initialized")
                }
            }
        })
    }
}

func TestAppManager_InitializeWithBridge(t *testing.T) {
    // Use mockdom for bridge testing
    cleanup := mockdom.SetupMockDOM()
    defer cleanup()

    config := &AppConfig{
        AppID:          "test-app",
        MountElementID: "app",
        InitialState:   AppState{},
    }

    manager := NewAppManager(config)

    ctx := context.Background()
    err := manager.Initialize(ctx)

    if err != nil {
        t.Fatalf("Initialize failed: %v", err)
    }

    // Test bridge availability
    bridgeManager := bridge.GetManager()
    if bridgeManager == nil {
        t.Error("Bridge manager not available after initialization")
    }
}
```

**Implementation File**: `appmanager/manager.go`

```go
package appmanager

import (
    "context"
    "fmt"
    "time"
    
    "github.com/ozanturksever/uiwgo/bridge"
    "github.com/ozanturksever/uiwgo/reactivity"
    "github.com/ozanturksever/uiwgo/router"
    "github.com/ozanturksever/uiwgo/wasm"
    "github.com/ozanturksever/logutil"
)

func NewAppManager(config *AppConfig) *AppManager {
    if config == nil {
        config = DefaultAppConfig()
    }
    
    return &AppManager{
        config:       config,
        lifecycle:    NewLifecycleManager(),
        cleanupScope: reactivity.CreateCleanupScope(),
    }
}

func (am *AppManager) Initialize(ctx context.Context) error {
    if am.initialized {
        return fmt.Errorf("app manager already initialized")
    }
    
    // Execute beforeInit hooks
    if err := am.lifecycle.ExecuteHooks(EventBeforeInit, &LifecycleContext{
        Event:   EventBeforeInit,
        Manager: am,
    }); err != nil {
        return fmt.Errorf("beforeInit hooks failed: %w", err)
    }
    
    // Initialize WASM
    wasmConfig := wasm.DefaultConfig()
    wasmConfig.Timeout = am.config.Timeout
    if err := wasm.Initialize(wasmConfig); err != nil {
        return fmt.Errorf("WASM initialization failed: %w", err)
    }
    
    // Initialize bridge
    am.bridge = bridge.NewRealManager()
    bridge.SetManager(am.bridge)
    
    // Initialize store
    am.store = NewAppStore(am.config.InitialState, am.config.PersistenceKey)
    
    // Initialize persistence if enabled
    if am.config.EnablePersistence {
        am.persistence = NewPersistenceManager(am.config.PersistenceKey, am.store)
        if err := am.persistence.Hydrate(); err != nil {
            logutil.Logf("Failed to hydrate state: %v", err)
        }
    }
    
    // Initialize router if enabled
    if am.config.EnableRouter && len(am.config.Routes) > 0 {
        outlet, err := bridge.GetElementByID(am.config.MountElementID)
        if err != nil {
            return fmt.Errorf("failed to get mount element: %w", err)
        }
        am.router = router.New(am.config.Routes, outlet.Raw())
    }
    
    am.initialized = true
    
    // Execute afterInit hooks
    if err := am.lifecycle.ExecuteHooks(EventAfterInit, &LifecycleContext{
        Event:   EventAfterInit,
        Manager: am,
    }); err != nil {
        return fmt.Errorf("afterInit hooks failed: %w", err)
    }
    
    return nil
}
```

#### 1.2 Lifecycle Management

**Test File**: `appmanager/lifecycle_test.go`

```go
func TestLifecycleManager_AddHook(t *testing.T) {
    lm := NewLifecycleManager()
    
    called := false
    hook := func(ctx *LifecycleContext) error {
        called = true
        return nil
    }
    
    lm.AddHook(EventBeforeInit, hook)
    
    err := lm.ExecuteHooks(EventBeforeInit, &LifecycleContext{
        Event: EventBeforeInit,
    })
    
    assert.NoError(t, err)
    assert.True(t, called)
}

func TestLifecycleManager_HookError(t *testing.T) {
    // Test error propagation
    // Test hook cancellation
}

func TestLifecycleManager_MultipleHooks(t *testing.T) {
    // Test execution order
    // Test partial failure handling
}
```

**Implementation File**: `appmanager/lifecycle.go`

```go
func NewLifecycleManager() *LifecycleManager {
    state := reactivity.CreateSignal(LifecycleStateUninitialized)
    eventLog := reactivity.CreateSignal([]LifecycleEvent{})

    lm := &LifecycleManager{
        hooks:      make(map[LifecycleEvent][]LifecycleHook),
        state:      state,
        eventLog:   eventLog,
        cleanupFns: make([]func(), 0),
    }

    // Set up reactive effect to log state changes
    lm.cleanupFns = append(lm.cleanupFns, reactivity.CreateEffect(func() {
        currentState := lm.state.Get()
        logutil.Logf("Lifecycle state changed to: %v", currentState)
    }))

    return lm
}

func (lm *LifecycleManager) AddHook(event LifecycleEvent, hook LifecycleHook) {
    lm.hooks[event] = append(lm.hooks[event], hook)
}

func (lm *LifecycleManager) ExecuteHooks(event LifecycleEvent, ctx *LifecycleContext) error {
    hooks, exists := lm.hooks[event]
    if !exists {
        return nil
    }

    // Log the event to reactive event log
    currentLog := lm.eventLog.Get()
    lm.eventLog.Set(append(currentLog, event))

    // Collect all errors instead of stopping on first error
    var hookErrors []error

    for i, hook := range hooks {
        if err := hook(ctx); err != nil {
            hookErrors = append(hookErrors, fmt.Errorf("hook %d for event %s: %w", i, event, err))
            // Continue executing other hooks instead of stopping
        }
    }

    // Return combined error if any hooks failed
    if len(hookErrors) > 0 {
        errorMessages := make([]string, len(hookErrors))
        for i, err := range hookErrors {
            errorMessages[i] = err.Error()
        }
        return fmt.Errorf("lifecycle hook failures: %s", strings.Join(errorMessages, "; "))
    }

    return nil
}

func (lm *LifecycleManager) GetState() LifecycleState {
    return lm.state.Get()
}

func (lm *LifecycleManager) GetEventHistory() []LifecycleEvent {
    return lm.eventLog.Get()
}

func (lm *LifecycleManager) Dispose() {
    // Clean up all reactive effects
    for _, cleanup := range lm.cleanupFns {
        cleanup()
    }
    lm.cleanupFns = nil
    lm.hooks = nil
}
```

#### 1.3 State Management Integration

**Test File**: `appmanager/store_test.go`

```go
func TestAppStore_Creation(t *testing.T) {
    initialState := AppState{
        User: map[string]interface{}{"name": "test"},
    }
    
    store := NewAppStore(initialState, "test-key")
    
    assert.NotNil(t, store)
    assert.Equal(t, "test", store.state.Get().User.(map[string]interface{})["name"])
}

func TestAppStore_Subscribe(t *testing.T) {
    // Test state subscription
    // Test reactive updates
}

func TestAppStore_NestedUpdates(t *testing.T) {
    // Test nested state updates
    // Test path-based subscriptions
}
```

### Phase 2: Router Integration

#### 2.1 Router Lifecycle Hooks

**Test File**: `appmanager/router_integration_test.go`

```go
func TestAppManager_RouterIntegration(t *testing.T) {
    routes := []*router.RouteDefinition{
        router.Route("/", func(props ...any) interface{} {
            return h.Div(g.Text("Home"))
        }),
        router.Route("/about", func(props ...any) interface{} {
            return h.Div(g.Text("About"))
        }),
    }
    
    config := &AppConfig{
        AppID:          "test-app",
        MountElementID: "app",
        Routes:         routes,
        EnableRouter:   true,
    }
    
    manager := NewAppManager(config)
    
    // Test router initialization
    // Test navigation hooks
    // Test route state persistence
}

func TestAppManager_NavigationHooks(t *testing.T) {
    // Test beforeRoute hooks
    // Test afterRoute hooks
    // Test route cancellation
}
```

#### 2.2 Enhanced Router State Management

```go
type RouterState struct {
    CurrentPath   string
    PreviousPath  string
    Params        map[string]string
    Query         map[string]string
    NavigationID  string
    Timestamp     time.Time
}

func (am *AppManager) Navigate(path string, options ...router.NavigateOptions) error {
    if !am.running {
        return fmt.Errorf("app manager not running")
    }
    
    // Execute beforeRoute hooks
    ctx := &LifecycleContext{
        Event:   EventBeforeRoute,
        Manager: am,
        Data: map[string]interface{}{
            "path":    path,
            "options": options,
        },
    }
    
    if err := am.lifecycle.ExecuteHooks(EventBeforeRoute, ctx); err != nil {
        return fmt.Errorf("beforeRoute hooks failed: %w", err)
    }
    
    // Perform navigation
    if am.router != nil {
        am.router.Navigate(path, options...)
    }
    
    // Update router state in store
    am.updateRouterState(path)
    
    // Execute afterRoute hooks
    ctx.Event = EventAfterRoute
    if err := am.lifecycle.ExecuteHooks(EventAfterRoute, ctx); err != nil {
        logutil.Logf("afterRoute hooks failed: %v", err)
    }
    
    return nil
}
```

### Phase 3: State Persistence

#### 3.1 Hydration/Dehydration System

**Test File**: `appmanager/persistence_test.go`

```go
func TestPersistenceManager_Hydrate(t *testing.T) {
    // Mock localStorage data
    // Test state restoration
    // Test hydration hooks
}

func TestPersistenceManager_Dehydrate(t *testing.T) {
    // Test state serialization
    // Test selective persistence
    // Test dehydration hooks
}

func TestPersistenceManager_AutoSave(t *testing.T) {
    // Test automatic state saving
    // Test save intervals
    // Test error handling
}
```

**Implementation File**: `appmanager/persistence.go`

```go
func NewPersistenceManager(key string, store *AppStore) *PersistenceManager {
    return &PersistenceManager{
        key:       key,
        store:     store,
        enabled:   true,
        hydrators: make(map[string]HydratorFunc),
    }
}

func (pm *PersistenceManager) Hydrate() error {
    if !pm.enabled {
        return nil
    }
    
    // Get data from localStorage via bridge
    manager := bridge.GetManager()
    if manager == nil {
        return fmt.Errorf("bridge manager not available")
    }
    
    localStorage := manager.JS().Global().Get("localStorage")
    if localStorage.IsUndefined() {
        return fmt.Errorf("localStorage not available")
    }
    
    dataStr := localStorage.Call("getItem", pm.key).String()
    if dataStr == "" {
        return nil // No saved state
    }
    
    // Parse JSON data
    var data map[string]interface{}
    if err := json.Unmarshal([]byte(dataStr), &data); err != nil {
        return fmt.Errorf("failed to parse saved state: %w", err)
    }
    
    // Apply hydrators
    for path, hydrator := range pm.hydrators {
        if value, exists := data[path]; exists {
            if err := hydrator(value); err != nil {
                logutil.Logf("Hydration failed for path %s: %v", path, err)
            }
        }
    }
    
    return nil
}

func (pm *PersistenceManager) Dehydrate() error {
    if !pm.enabled {
        return nil
    }
    
    // Collect state data
    data := make(map[string]interface{})
    
    // Get current state snapshot
    currentState := pm.store.state.Get()
    
    // Serialize relevant parts
    data["router"] = currentState.Router
    data["user"] = currentState.User
    data["custom"] = currentState.Custom
    
    // Convert to JSON
    jsonData, err := json.Marshal(data)
    if err != nil {
        return fmt.Errorf("failed to serialize state: %w", err)
    }
    
    // Save to localStorage
    manager := bridge.GetManager()
    if manager == nil {
        return fmt.Errorf("bridge manager not available")
    }
    
    localStorage := manager.JS().Global().Get("localStorage")
    localStorage.Call("setItem", pm.key, string(jsonData))
    
    return nil
}
```

### Phase 4: DOM Management & AppReady Events

#### 4.1 Enhanced DOM Integration

**Test File**: `appmanager/dom_integration_test.go`

```go
func TestAppManager_DOMReady(t *testing.T) {
    // Test DOM readiness detection
    // Test element mounting
    // Test cleanup on unmount
}

func TestAppManager_AppReadyEvent(t *testing.T) {
    // Test AppReady event emission
    // Test JavaScript integration
    // Test timing of ready events
}
```

**Implementation**: Enhanced DOM management with AppReady events

```go
func (am *AppManager) Mount(rootComponent func() g.Node) error {
    if !am.initialized {
        return fmt.Errorf("app manager not initialized")
    }
    
    // Execute beforeMount hooks
    ctx := &LifecycleContext{
        Event:   EventBeforeMount,
        Manager: am,
    }
    
    if err := am.lifecycle.ExecuteHooks(EventBeforeMount, ctx); err != nil {
        return fmt.Errorf("beforeMount hooks failed: %w", err)
    }
    
    // Mount the component
    disposer := comps.Mount(am.config.MountElementID, rootComponent)
    
    // Store disposer for cleanup
    am.cleanupScope.RegisterDisposer(disposer)
    
    am.running = true
    
    // Execute afterMount hooks
    ctx.Event = EventAfterMount
    if err := am.lifecycle.ExecuteHooks(EventAfterMount, ctx); err != nil {
        logutil.Logf("afterMount hooks failed: %v", err)
    }
    
    // Emit AppReady event
    am.emitAppReady()
    
    // Execute appReady hooks
    ctx.Event = EventAppReady
    if err := am.lifecycle.ExecuteHooks(EventAppReady, ctx); err != nil {
        logutil.Logf("appReady hooks failed: %v", err)
    }
    
    return nil
}

func (am *AppManager) emitAppReady() {
    if am.bridge == nil {
        return
    }
    
    // Emit to JavaScript
    global := am.bridge.JS().Global()
    event := global.Get("CustomEvent").New("appReady", map[string]interface{}{
        "detail": map[string]interface{}{
            "appId":     am.config.AppID,
            "timestamp": time.Now().Unix(),
        },
    })
    
    global.Get("window").Call("dispatchEvent", event)
    
    // Also emit wasmReady for compatibility
    wasmReadyEvent := global.Get("CustomEvent").New("wasmReady", map[string]interface{}{
        "detail": map[string]interface{}{
            "appId": am.config.AppID,
        },
    })
    
    global.Get("window").Call("dispatchEvent", wasmReadyEvent)
}
```

### Phase 5: Integration Tests & Examples

#### 5.1 Browser Integration Tests

**Test File**: `appmanager/integration_test.go`

```go
//go:build !js && !wasm

package appmanager

import (
    "context"
    "fmt"
    "testing"
    "time"

    "github.com/chromedp/chromedp"
    "github.com/chromedp/cdproto/cdp"
    "github.com/ozanturksever/uiwgo/devserver"
    "github.com/ozanturksever/uiwgo/internal/testhelpers"
)

func TestAppManager_FullLifecycle(t *testing.T) {
    // Create test server using existing patterns
    server := devserver.NewServer("appmanager_test", "localhost:0")
    if err := server.Start(); err != nil {
        t.Fatalf("Failed to start dev server: %v", err)
    }
    defer server.Stop()

    // Use existing chromedp test helpers
    chromedpCtx := testhelpers.MustNewChromedpContext(testhelpers.DefaultConfig())
    defer chromedpCtx.Cancel()

    err := chromedp.Run(chromedpCtx.Ctx,
        testhelpers.Actions.NavigateAndWaitForLoad(server.URL(), "body"),

        // Wait for WASM initialization using existing patterns
        testhelpers.Actions.WaitForWASMInit(),

        // Wait for AppReady event with timeout
        chromedp.WaitFunc(func(ctx context.Context, cur *cdp.Frame) error {
            var ready bool
            err := chromedp.Evaluate(`
                (function() {
                    return window.appReady === true && 
                           window.appManager !== undefined;
                })()
            `, &ready).Do(ctx)
            if err != nil {
                return err
            }
            if !ready {
                return fmt.Errorf("app not ready")
            }
            return nil
        }),

        // Test reactive state access
        chromedp.Evaluate(`
            (function() {
                const state = window.appManager.getState();
                return state && state.router && state.ui;
            })()
        `, nil),

        // Test navigation functionality
        chromedp.Click(`a[href="/about"]`),
        chromedp.WaitVisible(`#about-page`),

        // Verify router state update
        chromedp.WaitFunc(func(ctx context.Context, cur *cdp.Frame) error {
            var currentPath string
            err := chromedp.Evaluate(`window.appManager.getState().router.currentPath`, &currentPath).Do(ctx)
            if err != nil {
                return err
            }
            if currentPath != "/about" {
                return fmt.Errorf("expected current path '/about', got '%s'", currentPath)
            }
            return nil
        }),

        // Test state persistence
        chromedp.Evaluate(`
            (function() {
                // Trigger state change
                const state = window.appManager.getState();
                state.user = {name: 'test-user', theme: 'dark'};
                window.appManager.setState(state);

                // Check if localStorage has the data
                const saved = localStorage.getItem('demo-app-state');
                return saved !== null && saved.includes('test-user');
            })()
        `, nil),
    )

    if err != nil {
        t.Fatalf("Integration test failed: %v", err)
    }
}

func TestAppManager_ReactiveStateUpdates(t *testing.T) {
    server := devserver.NewServer("appmanager_reactive_test", "localhost:0")
    if err := server.Start(); err != nil {
        t.Fatalf("Failed to start dev server: %v", err)
    }
    defer server.Stop()

    chromedpCtx := testhelpers.MustNewChromedpContext(testhelpers.DefaultConfig())
    defer chromedpCtx.Cancel()

    err := chromedp.Run(chromedpCtx.Ctx,
        testhelpers.Actions.NavigateAndWaitForLoad(server.URL(), "body"),
        testhelpers.Actions.WaitForWASMInit(),

        // Wait for app ready
        chromedp.WaitFunc(func(ctx context.Context, cur *cdp.Frame) error {
            var ready bool
            err := chromedp.Evaluate(`window.appReady === true`, &ready).Do(ctx)
            if err != nil || !ready {
                return fmt.Errorf("app not ready")
            }
            return nil
        }),

        // Test reactive updates
        chromedp.Evaluate(`
            (function() {
                let updateCount = 0;

                // Subscribe to state changes
                window.appManager.subscribe('ui.theme', function(theme) {
                    updateCount++;
                    window.testUpdateCount = updateCount;
                    window.testLastTheme = theme;
                });

                // Trigger state change
                const state = window.appManager.getState();
                state.ui.theme = 'dark';
                window.appManager.setState(state);

                return true;
            })()
        `, nil),

        // Wait for reactive update
        chromedp.Sleep(100 * time.Millisecond),

        // Verify reactive update occurred
        chromedp.WaitFunc(func(ctx context.Context, cur *cdp.Frame) error {
            var updateCount int
            var lastTheme string

            err := chromedp.Evaluate(`window.testUpdateCount`, &updateCount).Do(ctx)
            if err != nil {
                return err
            }

            err = chromedp.Evaluate(`window.testLastTheme`, &lastTheme).Do(ctx)
            if err != nil {
                return err
            }

            if updateCount < 1 {
                return fmt.Errorf("expected at least 1 update, got %d", updateCount)
            }

            if lastTheme != "dark" {
                return fmt.Errorf("expected theme 'dark', got '%s'", lastTheme)
            }

            return nil
        }),
    )

    if err != nil {
        t.Fatalf("Reactive state test failed: %v", err)
    }
}
```

#### 5.2 Example Application

**File**: `examples/appmanager_demo/main.go`

```go
//go:build js && wasm

package main

import (
    "context"
    "time"
    
    "github.com/ozanturksever/uiwgo/appmanager"
    "github.com/ozanturksever/uiwgo/router"
    "github.com/ozanturksever/logutil"
    . "maragu.dev/gomponents"
    . "maragu.dev/gomponents/html"
)

func main() {
    // Configure the application using existing patterns
    config := &appmanager.AppConfig{
        AppID:             "demo-app",
        MountElementID:    "app",
        EnableRouter:      true,
        EnablePersistence: true,
        PersistenceKey:    "demo-app-state",
        Timeout:           30 * time.Second,
        Routes: []*router.RouteDefinition{
            router.Route("/", HomeComponent),
            router.Route("/about", AboutComponent),
            router.Route("/users/:id", UserComponent),
        },
        InitialState: appmanager.AppState{
            User: map[string]interface{}{
                "name":  "Guest",
                "theme": "light",
            },
            UI: appmanager.UIState{
                Theme:   "light",
                Loading: false,
                Sidebar: false,
                Modals:  []string{},
            },
            Custom: map[string]interface{}{
                "counter": 0,
            },
        },
        OnReady: func(am *appmanager.AppManager) error {
            logutil.Log("App is ready!")

            // Set up reactive effects for theme changes
            reactivity.CreateEffect(func() {
                state := am.GetState()
                theme := state.UI.Theme

                // Update document class using bridge
                manager := bridge.GetManager()
                if manager != nil {
                    doc := manager.DOM().Document()
                    body := doc.Body()
                    if body != nil {
                        body.SetClassName("theme-" + theme)
                    }
                }
            })

            return nil
        },
        OnError: func(err error) {
            logutil.Logf("App error: %v", err)
        },
    }

    // Create and initialize the app manager
    manager := appmanager.NewAppManager(config)

    // Add lifecycle hooks using reactive patterns
    manager.AddHook(appmanager.EventBeforeRoute, func(ctx *appmanager.LifecycleContext) error {
        if ctx.Data != nil {
            if data, ok := ctx.Data.(map[string]interface{}); ok {
                if path, exists := data["path"].(string); exists {
                    logutil.Logf("Navigating to: %s", path)

                    // Update UI state to show loading during navigation
                    state := manager.GetState()
                    state.UI.Loading = true
                    manager.SetState(state)
                }
            }
        }
        return nil
    })

    manager.AddHook(appmanager.EventAfterRoute, func(ctx *appmanager.LifecycleContext) error {
        // Clear loading state after navigation
        state := manager.GetState()
        state.UI.Loading = false
        manager.SetState(state)
        return nil
    })

    // Initialize the application with proper error handling
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    if err := manager.Initialize(ctx); err != nil {
        logutil.Logf("Failed to initialize app: %v", err)
        return
    }

    // Mount the root component using existing comps patterns
    if err := manager.Mount(func() g.Node {
        return RootComponent(manager)
    }); err != nil {
        logutil.Logf("Failed to mount app: %v", err)
        return
    }

    // Set up cleanup on page unload using existing patterns
    reactivity.OnCleanup(func() {
        logutil.Log("Cleaning up app manager...")
        manager.Cleanup()
    })

    // Keep running
    select {}
}

func RootComponent(manager *appmanager.AppManager) g.Node {
    return h.Div(
        g.Attr("class", "app min-h-screen bg-base-100"),

        // Header with navigation
        h.Header(
            g.Attr("class", "navbar bg-base-200 shadow-lg"),
            h.Div(
                g.Attr("class", "navbar-start"),
                h.A(
                    g.Attr("class", "btn btn-ghost text-xl"),
                    g.Attr("href", "/"),
                    g.Text("Demo App"),
                ),
            ),
            h.Div(
                g.Attr("class", "navbar-end"),
                h.Nav(
                    g.Attr("class", "menu menu-horizontal px-1"),
                    h.Li(h.A(g.Attr("href", "/"), g.Text("Home"))),
                    h.Li(h.A(g.Attr("href", "/about"), g.Text("About"))),
                    h.Li(h.A(g.Attr("href", "/users/123"), g.Text("Profile"))),
                ),
            ),
        ),

        // Main content area with router outlet
        h.Main(
            g.Attr("class", "container mx-auto p-4"),
            g.Attr("id", "router-outlet"),
            // Router content will be rendered here
        ),

        // Loading indicator using reactive state
        LoadingIndicator(manager),

        // Footer
        h.Footer(
            g.Attr("class", "footer footer-center p-4 bg-base-300 text-base-content mt-auto"),
            h.P(g.Text("© 2024 Demo App - Built with UIwGo")),
        ),
    )
}

func LoadingIndicator(manager *appmanager.AppManager) g.Node {
    // This would use reactive state in a real implementation
    return h.Div(
        g.Attr("class", "loading loading-spinner loading-lg fixed top-4 right-4 hidden"),
        g.Attr("id", "loading-indicator"),
    )
}

func RootComponent(manager *appmanager.AppManager) Node {
    return Div(
        Class("app"),
        Header(
            Nav(
                A(Href("/"), Text("Home")),
                A(Href("/about"), Text("About")),
                A(Href("/users/123"), Text("User Profile")),
            ),
        ),
        Main(
            ID("router-outlet"),
            // Router content will be rendered here
        ),
    )
}

func HomeComponent(props ...any) interface{} {
    return Div(
        ID("home-page"),
        H1(Text("Welcome Home")),
        P(Text("This is the home page")),
    )
}

func AboutComponent(props ...any) interface{} {
    return Div(
        ID("about-page"),
        H1(Text("About Us")),
        P(Text("This is the about page")),
    )
}

func UserComponent(props ...any) interface{} {
    return Div(
        ID("user-page"),
        H1(Text("User Profile")),
        P(Text("User details will be shown here")),
    )
}
```

## API Reference

### Core API

```go
// Create a new app manager with default config
manager := appmanager.NewAppManager(config)

// Or create with default config
manager := appmanager.NewAppManagerWithDefaults("my-app", "app")

// Initialize the application
ctx := context.Background()
err := manager.Initialize(ctx)

// Mount the root component using existing patterns
err := manager.Mount(func() g.Node {
    return MyRootComponent()
})

// Navigate programmatically
err := manager.Navigate("/path")

// Navigate with options
err := manager.Navigate("/path", router.NavigateOptions{
    Replace: true,
    State:   map[string]interface{}{"from": "app"},
})

// Get current reactive state
state := manager.GetState()

// Set state (triggers reactive updates)
newState := state
newState.User = map[string]interface{}{"name": "John"}
manager.SetState(newState)

// Subscribe to state changes using reactive patterns
unsubscribe := manager.SubscribeToPath("user.name", func(value interface{}) {
    logutil.Logf("User name changed: %v", value)
})

// Subscribe to lifecycle state changes
lifecycleUnsub := manager.SubscribeToLifecycle(func(state LifecycleState) {
    logutil.Logf("Lifecycle state: %v", state)
})

// Add lifecycle hooks
manager.AddHook(appmanager.EventBeforeRoute, func(ctx *LifecycleContext) error {
    // Handle before route change
    return nil
})

// Get lifecycle state
currentLifecycle := manager.GetLifecycleState()

// Get router instance
router := manager.GetRouter()

// Get mount context
mountCtx := manager.GetMountContext()

// Enable/disable persistence
manager.SetPersistenceEnabled(true)

// Manually trigger save
err := manager.SaveState()

// Manually trigger load
err := manager.LoadState()

// Clean up (disposes all reactive effects and cleanup scopes)
manager.Cleanup()

// Additional utility methods
isInitialized := manager.IsInitialized()
isRunning := manager.IsRunning()
appID := manager.GetAppID()
```

### JavaScript Integration

```javascript
// Listen for app ready event
window.addEventListener('appReady', (event) => {
    console.log('App is ready:', event.detail.appId);
    
    // Access app manager from JavaScript
    const manager = window.appManager;
    
    // Navigate from JavaScript
    manager.navigate('/about');
    
    // Get current state
    const state = manager.getState();
    
    // Subscribe to state changes
    manager.subscribe('user.theme', (theme) => {
        document.body.className = `theme-${theme}`;
    });
});
```

## Testing Strategy

### Unit Tests
- Test each component in isolation
- Mock dependencies using the existing mockdom package
- Test error conditions and edge cases
- Verify lifecycle hook execution order

### Integration Tests
- Test full application lifecycle in browser
- Test router integration with real navigation
- Test state persistence across page reloads
- Test JavaScript interop

### Performance Tests
- Test initialization time
- Test memory usage during lifecycle
- Test cleanup effectiveness
- Test state update performance

## Migration Guide

### From Current Pattern

**Before:**
```go
func main() {
    if err := wasm.QuickInit(); err != nil {
        logutil.Logf("Failed to initialize WASM: %v", err)
        return
    }
    
    bridge.InitializeManager(bridge.NewRealManager())
    
    routes := []*router.RouteDefinition{
        router.Route("/", HomeComponent),
    }
    
    outlet, _ := bridge.GetElementByID("app")
    router := router.New(routes, outlet.Raw())
    
    select {}
}
```

**After:**
```go
func main() {
    config := &appmanager.AppConfig{
        AppID:          "my-app",
        MountElementID: "app",
        EnableRouter:   true,
        Routes: []*router.RouteDefinition{
            router.Route("/", HomeComponent),
        },
    }
    
    manager := appmanager.NewAppManager(config)
    
    ctx := context.Background()
    if err := manager.Initialize(ctx); err != nil {
        logutil.Logf("Failed to initialize: %v", err)
        return
    }
    
    if err := manager.Mount(RootComponent); err != nil {
        logutil.Logf("Failed to mount: %v", err)
        return
    }
    
    select {}
}
```

## Utility Functions and Helpers

### Default Configurations

## Implementation Timeline

- **Week 1**: Phase 1 - Core Infrastructure (TDD)
- **Week 2**: Phase 2 - Router Integration
- **Week 3**: Phase 3 - State Persistence
- **Week 4**: Phase 4 - DOM Management & AppReady Events
- **Week 5**: Phase 5 - Integration Tests & Examples
- **Week 6**: Documentation, Migration Guide, and Polish

This design provides a comprehensive, test-driven approach to WASM application management that integrates seamlessly with the existing architecture while providing significant improvements in developer experience and functionality.