// types.go
// Core types and interfaces for the Golid framework

package golid

import (
	"sync"
	"syscall/js"

	"maragu.dev/gomponents"
)

// ---------------------------
// 🔧 Core Types & Interfaces
// ---------------------------

// JsCallback defines a JavaScript callback function type
type JsCallback func(this js.Value, args []js.Value) interface{}

// Global document reference
var doc = js.Global().Get("document")

// ---------------------------
// 🔄 Lifecycle Types
// ---------------------------

// ElementCallback defines a callback function for element operations
type ElementCallback func()

// LifecycleHook defines a function type for lifecycle events
type LifecycleHook func()

// ComponentLifecycle manages the lifecycle state of a component
type ComponentLifecycle struct {
	onInit      []LifecycleHook
	onMount     []LifecycleHook
	onDismount  []LifecycleHook
	initialized bool
	mounted     bool
	mutex       sync.RWMutex
}

// ---------------------------
// 🔍 Observer Types
// ---------------------------

// ObserverManager manages DOM mutation observation
type ObserverManager struct {
	observer          js.Value
	callbacks         map[string]ElementCallback
	dismountCallbacks map[string][]LifecycleHook
	trackedElements   map[string]js.Value
	isObserving       bool
	mutex             sync.RWMutex
}

// ---------------------------
// 📦 Signal Types
// ---------------------------

// effect represents a reactive effect with dependencies
type effect struct {
	fn   func()
	deps map[any]struct{}
}

// hasWatchers interface for objects that can have watchers removed
type hasWatchers interface {
	removeWatcher(*effect)
}

// Signal represents a reactive value that can be watched
type Signal[T any] struct {
	value    T
	watchers map[*effect]struct{}
}

// ---------------------------
// 🧭 Router Types
// ---------------------------

// RouteParams holds parameters extracted from the URL path
type RouteParams map[string]string

// RouteMatch represents a matched route
type RouteMatch struct {
	Path    string
	Params  RouteParams
	Query   map[string]string
	Handler func(RouteParams) gomponents.Node
}

// Route defines a single route configuration
type Route struct {
	Path     string
	Handler  func(RouteParams) gomponents.Node
	Guard    func(RouteParams) bool // Optional route guard
	Children []Route                // For nested routing
}

// Router manages the routing system
type Router struct {
	routes        []Route
	currentPath   *Signal[string]
	currentParams *Signal[RouteParams]
	currentQuery  *Signal[map[string]string]
	basePath      string
	history       js.Value
}

// ---------------------------
// 🧱 Component Types
// ---------------------------

// Component represents a component with lifecycle hooks
type Component struct {
	id        string
	render    func() gomponents.Node
	lifecycle *ComponentLifecycle
}
