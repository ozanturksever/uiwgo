// lifecycle_v2.go
// Enhanced component lifecycle system with reactivity integration and cascade prevention

//go:build !js && !wasm

package golid

import (
	"sync"
	"sync/atomic"
	"time"

	"maragu.dev/gomponents"
)

// ------------------------------------
// 🔄 Component State Machine
// ------------------------------------

// ComponentStateV2 represents the lifecycle state of a component
type ComponentStateV2 int

const (
	ComponentUnmountedV2 ComponentStateV2 = iota
	ComponentMountingV2
	ComponentMountedV2
	ComponentUpdatingV2
	ComponentUnmountingV2
	ComponentErrorV2
)

// String returns the string representation of ComponentStateV2
func (s ComponentStateV2) String() string {
	switch s {
	case ComponentUnmountedV2:
		return "Unmounted"
	case ComponentMountingV2:
		return "Mounting"
	case ComponentMountedV2:
		return "Mounted"
	case ComponentUpdatingV2:
		return "Updating"
	case ComponentUnmountingV2:
		return "Unmounting"
	case ComponentErrorV2:
		return "Error"
	default:
		return "Unknown"
	}
}

// ------------------------------------
// 🧱 Enhanced Component System
// ------------------------------------

var componentV2IdCounter uint64

// ComponentV2 represents a component with enhanced lifecycle management
type ComponentV2 struct {
	id        uint64
	owner     *Owner
	parent    *ComponentV2
	children  []*ComponentV2
	state     ComponentStateV2
	render    func() gomponents.Node
	context   *ComponentContextV2
	hooks     *LifecycleHooksV2
	resources *ResourceTracker
	guard     *LifecycleGuardV2
	mutex     sync.RWMutex
	disposed  bool
}

// ComponentContextV2 provides context for component execution
type ComponentContextV2 struct {
	scheduler     *Scheduler
	errorBoundary *ErrorBoundaryV2
	depth         int
	maxDepth      int
}

// LifecycleHooksV2 manages component lifecycle callbacks
type LifecycleHooksV2 struct {
	onMount   []func()
	onCleanup []func()
	onError   []func(error)
	onUpdate  []func()
	mutex     sync.RWMutex
}

// ------------------------------------
// 🛡️ Lifecycle Guard System
// ------------------------------------

// LifecycleGuardV2 prevents recursive lifecycle execution
type LifecycleGuardV2 struct {
	depth     int
	maxDepth  int
	visited   map[uint64]bool
	executing map[uint64]bool
	mutex     sync.RWMutex
}

// NewLifecycleGuardV2 creates a new lifecycle guard
func NewLifecycleGuardV2(maxDepth int) *LifecycleGuardV2 {
	return &LifecycleGuardV2{
		maxDepth:  maxDepth,
		visited:   make(map[uint64]bool),
		executing: make(map[uint64]bool),
	}
}

// Enter attempts to enter a lifecycle operation
func (g *LifecycleGuardV2) Enter(componentId uint64) bool {
	g.mutex.Lock()
	defer g.mutex.Unlock()

	// Check depth limit
	if g.depth >= g.maxDepth {
		return false
	}

	// Check if already executing
	if g.executing[componentId] {
		return false
	}

	// Check if visited in this cycle
	if g.visited[componentId] {
		return false
	}

	g.depth++
	g.executing[componentId] = true
	g.visited[componentId] = true
	return true
}

// Exit exits a lifecycle operation
func (g *LifecycleGuardV2) Exit(componentId uint64) {
	g.mutex.Lock()
	defer g.mutex.Unlock()

	g.depth--
	delete(g.executing, componentId)
}

// Reset resets the guard state
func (g *LifecycleGuardV2) Reset() {
	g.mutex.Lock()
	defer g.mutex.Unlock()

	g.depth = 0
	g.visited = make(map[uint64]bool)
	g.executing = make(map[uint64]bool)
}

// ------------------------------------
// 🏗️ Component Creation
// ------------------------------------

// CreateComponentV2 creates a new component with automatic cleanup
func CreateComponentV2(fn func() gomponents.Node) *ComponentV2 {
	owner := getCurrentOwner()
	if owner == nil {
		// Create a root owner if none exists
		_, cleanup := CreateRoot(func() interface{} { return nil })
		defer cleanup()
		owner = getCurrentOwner()
	}

	comp := &ComponentV2{
		id:     atomic.AddUint64(&componentV2IdCounter, 1),
		owner:  owner,
		state:  ComponentUnmountedV2,
		render: fn,
		context: &ComponentContextV2{
			scheduler: getScheduler(),
			depth:     0,
			maxDepth:  10, // Configurable cascade depth limit
		},
		hooks: &LifecycleHooksV2{
			onMount:   make([]func(), 0),
			onCleanup: make([]func(), 0),
			onError:   make([]func(error), 0),
			onUpdate:  make([]func(), 0),
		},
		resources: NewResourceTracker(),
		guard:     NewLifecycleGuardV2(5), // Prevent deep recursion
		children:  make([]*ComponentV2, 0),
	}

	// Register with owner for cleanup
	if owner != nil {
		owner.addCleanup(func() {
			comp.UnmountV2()
		})
	}

	return comp
}

// ------------------------------------
// 🔗 Component Hierarchy Management
// ------------------------------------

// MountComponentV2 safely mounts a component with hierarchy tracking
func MountComponentV2(component *ComponentV2, parent *ComponentV2) error {
	if component == nil {
		return nil
	}

	component.mutex.Lock()
	defer component.mutex.Unlock()

	// Check if already mounted
	if component.state != ComponentUnmountedV2 {
		return nil
	}

	// Set parent relationship
	if parent != nil {
		component.parent = parent
		parent.addChildV2(component)
	}

	// Transition to mounting state
	component.state = ComponentMountingV2

	// Check cascade guard
	if !component.guard.Enter(component.id) {
		component.state = ComponentUnmountedV2
		return ErrCascadeLimitV2
	}

	defer component.guard.Exit(component.id)

	// Execute mount hooks with error handling
	if err := component.executeMountHooksV2(); err != nil {
		component.state = ComponentErrorV2
		component.handleErrorV2(err)
		return err
	}

	// Transition to mounted state
	component.state = ComponentMountedV2
	return nil
}

// UnmountV2 guarantees cleanup with cascade prevention
func (c *ComponentV2) UnmountV2() error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	// Check if already unmounted
	if c.state == ComponentUnmountedV2 || c.disposed {
		return nil
	}

	// Transition to unmounting state
	c.state = ComponentUnmountingV2

	// Check cascade guard
	if !c.guard.Enter(c.id) {
		return ErrCascadeLimitV2
	}

	defer c.guard.Exit(c.id)

	// Unmount children first
	for _, child := range c.children {
		child.UnmountV2()
	}

	// Execute cleanup hooks
	c.executeCleanupHooksV2()

	// Clean up resources
	c.resources.CleanupAll()

	// Remove from parent
	if c.parent != nil {
		c.parent.removeChildV2(c)
	}

	// Transition to unmounted state
	c.state = ComponentUnmountedV2
	c.disposed = true

	return nil
}

// ------------------------------------
// 🎣 Lifecycle Hooks
// ------------------------------------

// OnMountV2 registers a mount hook
func (c *ComponentV2) OnMountV2(fn func()) *ComponentV2 {
	c.hooks.mutex.Lock()
	defer c.hooks.mutex.Unlock()
	c.hooks.onMount = append(c.hooks.onMount, fn)
	return c
}

// OnCleanupV2 registers a cleanup hook
func (c *ComponentV2) OnCleanupV2(fn func()) *ComponentV2 {
	c.hooks.mutex.Lock()
	defer c.hooks.mutex.Unlock()
	c.hooks.onCleanup = append(c.hooks.onCleanup, fn)
	return c
}

// OnErrorV2 registers an error hook
func (c *ComponentV2) OnErrorV2(fn func(error)) *ComponentV2 {
	c.hooks.mutex.Lock()
	defer c.hooks.mutex.Unlock()
	c.hooks.onError = append(c.hooks.onError, fn)
	return c
}

// OnUpdateV2 registers an update hook
func (c *ComponentV2) OnUpdateV2(fn func()) *ComponentV2 {
	c.hooks.mutex.Lock()
	defer c.hooks.mutex.Unlock()
	c.hooks.onUpdate = append(c.hooks.onUpdate, fn)
	return c
}

// ------------------------------------
// 🔄 Component Rendering
// ------------------------------------

// RenderV2 renders the component with lifecycle management
func (c *ComponentV2) RenderV2() gomponents.Node {
	c.mutex.RLock()
	if c.disposed {
		c.mutex.RUnlock()
		return nil
	}
	c.mutex.RUnlock()

	// Mount component if not already mounted
	if c.state == ComponentUnmountedV2 {
		if err := MountComponentV2(c, nil); err != nil {
			c.handleErrorV2(err)
			return nil
		}
	}

	// Execute render function within owner context
	var result gomponents.Node
	if c.owner != nil {
		result = RunWithOwner(c.owner, func() gomponents.Node {
			return c.render()
		})
	} else {
		result = c.render()
	}

	return result
}

// ------------------------------------
// 🛠️ Internal Methods
// ------------------------------------

// executeMountHooksV2 executes all mount hooks with error handling
func (c *ComponentV2) executeMountHooksV2() error {
	c.hooks.mutex.RLock()
	hooks := make([]func(), len(c.hooks.onMount))
	copy(hooks, c.hooks.onMount)
	c.hooks.mutex.RUnlock()

	for _, hook := range hooks {
		if err := c.safeExecuteHookV2(hook); err != nil {
			return err
		}
	}
	return nil
}

// executeCleanupHooksV2 executes all cleanup hooks
func (c *ComponentV2) executeCleanupHooksV2() {
	c.hooks.mutex.RLock()
	hooks := make([]func(), len(c.hooks.onCleanup))
	copy(hooks, c.hooks.onCleanup)
	c.hooks.mutex.RUnlock()

	for _, hook := range hooks {
		c.safeExecuteHookV2(hook)
	}
}

// safeExecuteHookV2 executes a hook with panic recovery
func (c *ComponentV2) safeExecuteHookV2(hook func()) (err error) {
	defer func() {
		if r := recover(); r != nil {
			if e, ok := r.(error); ok {
				err = e
			} else {
				err = ErrHookPanicV2
			}
		}
	}()

	hook()
	return nil
}

// handleErrorV2 handles component errors
func (c *ComponentV2) handleErrorV2(err error) {
	c.hooks.mutex.RLock()
	errorHooks := make([]func(error), len(c.hooks.onError))
	copy(errorHooks, c.hooks.onError)
	c.hooks.mutex.RUnlock()

	for _, hook := range errorHooks {
		func() {
			defer func() {
				recover() // Prevent error hooks from panicking
			}()
			hook(err)
		}()
	}
}

// addChildV2 adds a child component
func (c *ComponentV2) addChildV2(child *ComponentV2) {
	c.children = append(c.children, child)
}

// removeChildV2 removes a child component
func (c *ComponentV2) removeChildV2(child *ComponentV2) {
	for i, c := range c.children {
		if c == child {
			c.children = append(c.children[:i], c.children[i+1:]...)
			break
		}
	}
}

// ------------------------------------
// 🧹 Resource Management Wrapper
// ------------------------------------

// WithCleanupV2 wraps a function with automatic cleanup
func WithCleanupV2(fn func()) func() {
	return func() {
		if owner := getCurrentOwner(); owner != nil {
			OnCleanup(fn)
		} else {
			// Execute immediately if no owner context
			fn()
		}
	}
}

// ComponentBoundaryV2 creates an error and cleanup isolation boundary
func ComponentBoundaryV2(fn func()) gomponents.Node {
	_, cleanup := CreateRoot(func() interface{} {
		fn()
		return nil
	})

	// Register cleanup with current owner if available
	if owner := getCurrentOwner(); owner != nil {
		OnCleanup(cleanup)
	}

	return nil
}

// ------------------------------------
// 📊 Component Statistics
// ------------------------------------

// GetComponentStatsV2 returns statistics about the component
func (c *ComponentV2) GetComponentStatsV2() map[string]interface{} {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	return map[string]interface{}{
		"id":            c.id,
		"state":         c.state.String(),
		"childCount":    len(c.children),
		"disposed":      c.disposed,
		"hasParent":     c.parent != nil,
		"resourceCount": c.resources.GetResourceCount(),
		"guardDepth":    c.guard.depth,
	}
}

// ------------------------------------
// 🚨 Error Types
// ------------------------------------

var (
	ErrCascadeLimitV2 = &ComponentErrorV2{Type: "CascadeLimit", Message: "Component lifecycle cascade limit exceeded"}
	ErrHookPanicV2    = &ComponentErrorV2{Type: "HookPanic", Message: "Component hook panicked"}
)

// ComponentErrorV2 represents a component-specific error
type ComponentErrorV2 struct {
	Type    string
	Message string
}

func (e *ComponentErrorV2) Error() string {
	return e.Type + ": " + e.Message
}

// ------------------------------------
// 🔧 Error Boundary System
// ------------------------------------

// ErrorBoundaryV2 provides error isolation for components
type ErrorBoundaryV2 struct {
	owner     *Owner
	fallback  func(error) gomponents.Node
	onError   func(error, *ErrorInfoV2)
	recovered bool
	mutex     sync.RWMutex
}

// ErrorInfoV2 contains information about an error
type ErrorInfoV2 struct {
	Error     error
	Component string
	Timestamp time.Time
}

// CreateErrorBoundaryV2 creates a new error boundary
func CreateErrorBoundaryV2(fallback func(error) gomponents.Node) *ErrorBoundaryV2 {
	return &ErrorBoundaryV2{
		owner:    getCurrentOwner(),
		fallback: fallback,
	}
}

// Catch catches and handles errors within the boundary
func (e *ErrorBoundaryV2) Catch(fn func()) (err error) {
	defer func() {
		if r := recover(); r != nil {
			e.mutex.Lock()
			defer e.mutex.Unlock()

			if er, ok := r.(error); ok {
				err = er
			} else {
				err = ErrHookPanicV2
			}

			e.recovered = true

			if e.onError != nil {
				info := &ErrorInfoV2{
					Error:     err,
					Component: "unknown",
					Timestamp: time.Now(),
				}
				e.onError(err, info)
			}
		}
	}()

	fn()
	return nil
}

// Reset resets the error boundary state
func (e *ErrorBoundaryV2) Reset() {
	e.mutex.Lock()
	defer e.mutex.Unlock()
	e.recovered = false
}
