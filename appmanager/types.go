package appmanager

import (
    "time"

    "github.com/ozanturksever/uiwgo/reactivity"
)

// LifecycleState represents the high-level state of the application lifecycle.
type LifecycleState int

const (
    LifecycleStateUninitialized LifecycleState = iota
    LifecycleStateInitialized
    LifecycleStateRunning
    LifecycleStateStopped
)

// LifecycleEvent represents hook points in the lifecycle.
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

// LifecycleContext provides contextual information to hooks.
type LifecycleContext struct {
    Event   LifecycleEvent
    Manager *AppManager
    Data    any
    Cancel  func()
}

// LifecycleHook is a hook function signature.
type LifecycleHook func(ctx *LifecycleContext) error

// UIState is a sample UI-related state used by tests and examples.
type UIState struct {
    Theme   string
    Loading bool
}

// RouterState tracks minimal router state for tests.
type RouterState struct {
    CurrentPath  string
    PreviousPath string
}

// AppState is the root application state stored in AppStore.
type AppState struct {
    Router RouterState
    User   any
    UI     UIState
    Custom map[string]any
}

// AppConfig configures the AppManager.
type AppConfig struct {
    AppID             string
    MountElementID    string
    Routes            []*RouteDefinitionAlias // placeholder alias to avoid importing router in this file
    InitialState      AppState
    PersistenceKey    string
    EnableRouter      bool
    EnablePersistence bool
    Timeout           time.Duration
    OnReady           func(*AppManager) error
    OnError           func(error)
}

// DefaultAppConfig returns a safe default config.
func DefaultAppConfig() *AppConfig {
    st := AppState{
        UI:     UIState{Theme: "light"},
        Custom: map[string]any{},
    }
    return &AppConfig{
        AppID:          "app",
        MountElementID: "app",
        InitialState:   st,
        Timeout:        30 * time.Second,
    }
}

// LifecycleManager manages lifecycle hooks and state.
type LifecycleManager struct {
    hooks map[LifecycleEvent][]LifecycleHook
    state reactivity.Signal[LifecycleState]
}
