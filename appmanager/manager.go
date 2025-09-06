package appmanager

import (
	"context"
	"fmt"

	"github.com/ozanturksever/logutil"
	"github.com/ozanturksever/uiwgo/comps"
	"github.com/ozanturksever/uiwgo/reactivity"
	"github.com/ozanturksever/uiwgo/router"
	"github.com/ozanturksever/uiwgo/wasm"
	"honnef.co/go/js/dom/v2"
	g "maragu.dev/gomponents"
)

// AppManager orchestrates application lifecycle
type AppManager struct {
	config       *AppConfig
	router       *router.Router
	store        *AppStore
	lifecycle    *LifecycleManager
	initialized  reactivity.Signal[bool]
	running      reactivity.Signal[bool]
	cleanupScope *reactivity.CleanupScope
	disposer     func()
}

// NewAppManager constructs a new AppManager with given or default config
func NewAppManager(config *AppConfig) *AppManager {
	if config == nil {
		config = DefaultAppConfig()
	}
	am := &AppManager{
		config:       config,
		lifecycle:    NewLifecycleManager(),
		initialized:  reactivity.CreateSignal(false),
		running:      reactivity.CreateSignal(false),
		cleanupScope: reactivity.NewCleanupScope(nil),
	}
	// Initialize store immediately so tests can verify initial state pre-initialize
	am.store = NewAppStore(config.InitialState, config.PersistenceKey)
	return am
}

// Initialize sets up wasm (if supported), bridge manager (if not set), store and lifecycle
func (am *AppManager) Initialize(ctx context.Context) error { // ctx reserved for future use
	if am.initialized.Get() {
		return fmt.Errorf("app manager already initialized")
	}

	// beforeInit hooks
	if err := am.lifecycle.ExecuteHooks(EventBeforeInit, &LifecycleContext{Event: EventBeforeInit, Manager: am}); err != nil {
		return fmt.Errorf("beforeInit hooks failed: %w", err)
	}

	// Initialize WASM (no-op / ErrNotSupported on non-wasm; treat as non-fatal)
	cfg := wasm.DefaultConfig()
	cfg.Timeout = am.config.Timeout
	if err := wasm.Initialize(cfg); err != nil {
		// On non-wasm platforms Initialize returns an error; we log and continue to allow tests to run.
		logutil.Logf("WASM initialization skipped or failed: %v", err)
	}

	// Defer router creation to Mount where we can target #router-outlet

	// Mark initialized and update lifecycle state
	am.initialized.Set(true)
	am.lifecycle.setState(LifecycleStateInitialized)

	// afterInit hooks
	if err := am.lifecycle.ExecuteHooks(EventAfterInit, &LifecycleContext{Event: EventAfterInit, Manager: am}); err != nil {
		return fmt.Errorf("afterInit hooks failed: %w", err)
	}

	return nil
}

// AddHook registers a lifecycle hook on the internal lifecycle manager
func (am *AppManager) AddHook(event LifecycleEvent, hook LifecycleHook) {
	am.lifecycle.AddHook(event, hook)
}

// Mount renders the root component into MountElementID and initializes the router if enabled
func (am *AppManager) Mount(root func() g.Node) error {
	if !am.initialized.Get() {
		return fmt.Errorf("app manager not initialized")
	}

	// beforeMount hooks
	if err := am.lifecycle.ExecuteHooks(EventBeforeMount, &LifecycleContext{Event: EventBeforeMount, Manager: am}); err != nil {
		return fmt.Errorf("beforeMount hooks failed: %w", err)
	}

	// Mount component via comps
	am.disposer = comps.Mount(am.config.MountElementID, root)

	// Setup router after mount if enabled
	if am.config.EnableRouter && len(am.config.Routes) > 0 {
		// Prefer #router-outlet, fallback to mount element
		outlet := dom.GetWindow().Document().GetElementByID("router-outlet")
		if outlet == nil {
			outlet = dom.GetWindow().Document().GetElementByID(am.config.MountElementID)
		}
		logutil.Log("outlet:", outlet, "id:", am.config.MountElementID)
		if outlet != nil {
			am.router = router.New(am.config.Routes, outlet)
			// Wire router navigation callbacks to lifecycle hooks and store updates
			am.router.OnBeforeNavigate = func(path string, options router.NavigateOptions) {
				if err := am.lifecycle.ExecuteHooks(EventBeforeRoute, &LifecycleContext{Event: EventBeforeRoute, Manager: am, Data: map[string]any{"path": path}}); err != nil {
					logutil.Logf("beforeRoute hooks failed: %v", err)
				}
			}
			am.router.OnAfterNavigate = func(path string, options router.NavigateOptions) {
				// Update router state snapshot
				st := am.store.Get()
				st.Router.PreviousPath = st.Router.CurrentPath
				st.Router.CurrentPath = path
				am.store.Replace(st)
				if err := am.lifecycle.ExecuteHooks(EventAfterRoute, &LifecycleContext{Event: EventAfterRoute, Manager: am, Data: map[string]any{"path": path}}); err != nil {
					logutil.Logf("afterRoute hooks failed: %v", err)
				}
			}
		}
	}

	am.running.Set(true)
	am.lifecycle.setState(LifecycleStateRunning)

	// afterMount hooks
	if err := am.lifecycle.ExecuteHooks(EventAfterMount, &LifecycleContext{Event: EventAfterMount, Manager: am}); err != nil {
		logutil.Logf("afterMount hooks failed: %v", err)
	}
	return nil
}

// Navigate performs navigation via internal router and updates Router state
func (am *AppManager) Navigate(path string, opts ...router.NavigateOptions) error {
	if !am.running.Get() {
		return fmt.Errorf("app manager not running")
	}
	if am.router != nil {
		// Delegate to router; callbacks will handle hooks and store updates
		var options router.NavigateOptions
		if len(opts) > 0 {
			options = opts[0]
		}
		am.router.Navigate(path, options)
		return nil
	}
	// Fallback when no router is present: perform minimal state update and hooks
	if err := am.lifecycle.ExecuteHooks(EventBeforeRoute, &LifecycleContext{Event: EventBeforeRoute, Manager: am, Data: map[string]any{"path": path}}); err != nil {
		return fmt.Errorf("beforeRoute hooks failed: %w", err)
	}
	st := am.store.Get()
	st.Router.PreviousPath = st.Router.CurrentPath
	st.Router.CurrentPath = path
	am.store.Replace(st)
	if err := am.lifecycle.ExecuteHooks(EventAfterRoute, &LifecycleContext{Event: EventAfterRoute, Manager: am, Data: map[string]any{"path": path}}); err != nil {
		logutil.Logf("afterRoute hooks failed: %v", err)
	}
	return nil
}

// GetState returns a snapshot of AppState
func (am *AppManager) GetState() AppState { return am.store.Get() }

// SetState replaces the entire app state
func (am *AppManager) SetState(st AppState) { am.store.Replace(st) }

// Cleanup disposes mounted UI and scope
func (am *AppManager) Cleanup() {
	if am.disposer != nil {
		am.disposer()
		am.disposer = nil
	}
	if am.cleanupScope != nil {
		am.cleanupScope.Dispose()
	}
	am.running.Set(false)
	am.lifecycle.setState(LifecycleStateStopped)
}

// Accessors
func (am *AppManager) IsInitialized() bool       { return am.initialized.Get() }
func (am *AppManager) IsRunning() bool           { return am.running.Get() }
func (am *AppManager) GetRouter() *router.Router { return am.router }
func (am *AppManager) GetAppID() string          { return am.config.AppID }
