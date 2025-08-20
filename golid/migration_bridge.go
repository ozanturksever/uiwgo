// migration_bridge.go
// V1 to V2 compatibility bridge for safe migration
package golid

import (
	"sync"
	"sync/atomic"

	"maragu.dev/gomponents"
)

// ------------------------------------
// 🌉 V1/V2 Compatibility Bridge
// ------------------------------------

var (
	migrationMetrics = &MigrationMetrics{
		V1CallsRemaining: make(map[string]int64),
		mutex:            sync.RWMutex{},
	}
	migrationEnabled = int32(1) // 1 = enabled, 0 = disabled
)

// MigrationMetrics tracks the progress of V1 to V2 migration
type MigrationMetrics struct {
	SignalOperations  int64
	EffectExecutions  int64
	ComponentMounts   int64
	MemoryAllocations int64
	CascadePrevented  int64
	V1CallsRemaining  map[string]int64
	mutex             sync.RWMutex
}

// GetMigrationMetrics returns current migration progress metrics
func GetMigrationMetrics() *MigrationMetrics {
	migrationMetrics.mutex.RLock()
	defer migrationMetrics.mutex.RUnlock()

	// Create a copy to avoid race conditions
	metrics := &MigrationMetrics{
		SignalOperations:  atomic.LoadInt64(&migrationMetrics.SignalOperations),
		EffectExecutions:  atomic.LoadInt64(&migrationMetrics.EffectExecutions),
		ComponentMounts:   atomic.LoadInt64(&migrationMetrics.ComponentMounts),
		MemoryAllocations: atomic.LoadInt64(&migrationMetrics.MemoryAllocations),
		CascadePrevented:  atomic.LoadInt64(&migrationMetrics.CascadePrevented),
		V1CallsRemaining:  make(map[string]int64),
	}

	for k, v := range migrationMetrics.V1CallsRemaining {
		metrics.V1CallsRemaining[k] = v
	}

	return metrics
}

// trackV1Call increments the counter for a specific V1 API call
func trackV1Call(apiName string) {
	if atomic.LoadInt32(&migrationEnabled) == 0 {
		return
	}

	migrationMetrics.mutex.Lock()
	migrationMetrics.V1CallsRemaining[apiName]++
	migrationMetrics.mutex.Unlock()
}

// ------------------------------------
// 🔄 Signal Bridge
// ------------------------------------

// LegacySignal provides V1 API compatibility while using V2 internals
type LegacySignal[T any] struct {
	getter   func() T
	setter   func(T)
	v2Signal bool
	apiName  string
}

// NewSignal creates a V1-compatible signal that uses V2 internals
// This maintains API compatibility while providing V2 performance benefits
func NewSignal[T any](initial T) *LegacySignal[T] {
	trackV1Call("NewSignal")

	// Use V2 CreateSignal internally but wrap with V1 API
	getter, setter := CreateSignal(initial)

	return &LegacySignal[T]{
		getter:   getter,
		setter:   setter,
		v2Signal: true,
		apiName:  "LegacySignal",
	}
}

// Get retrieves the current value (V1 API compatibility)
func (s *LegacySignal[T]) Get() T {
	trackV1Call("Signal.Get")
	atomic.AddInt64(&migrationMetrics.SignalOperations, 1)
	return s.getter()
}

// Set updates the signal value (V1 API compatibility)
func (s *LegacySignal[T]) Set(value T) {
	trackV1Call("Signal.Set")
	atomic.AddInt64(&migrationMetrics.SignalOperations, 1)
	s.setter(value)
}

// Subscribe provides V1-style manual subscription (deprecated)
func (s *LegacySignal[T]) Subscribe(handler func(T)) func() {
	trackV1Call("Signal.Subscribe")

	// Convert to V2 effect-based subscription
	var cleanup func()
	CreateOwner(func() {
		CreateEffect(func() {
			value := s.getter()
			handler(value)
		}, nil)
		cleanup = func() {
			// V2 automatic cleanup through owner disposal
		}
	})

	return cleanup
}

// removeWatcher provides V1 API compatibility (no-op in V2)
func (s *LegacySignal[T]) removeWatcher(e *effect) {
	trackV1Call("Signal.removeWatcher")
	// V2 handles this automatically through owner context
}

// ------------------------------------
// 🧱 Component Bridge
// ------------------------------------

// LegacyComponent provides V1 API compatibility with V2 internals
type LegacyComponent struct {
	renderV1      func() gomponents.Node
	v2Owner       *Owner
	hooks         []LifecycleHook
	mountHooks    []LifecycleHook
	dismountHooks []LifecycleHook
	initialized   bool
	mounted       bool
	id            string
}

// NewComponent creates a V1-compatible component using V2 owner context
func NewComponent(render func() gomponents.Node) *LegacyComponent {
	trackV1Call("NewComponent")
	atomic.AddInt64(&migrationMetrics.ComponentMounts, 1)

	// Create V2 owner context for automatic cleanup
	var owner *Owner
	CreateOwner(func() {
		owner = getCurrentOwner()
	})

	return &LegacyComponent{
		renderV1: render,
		v2Owner:  owner,
		hooks:    make([]LifecycleHook, 0),
		id:       GenID(),
	}
}

// OnInit registers initialization hooks (V1 compatibility)
func (c *LegacyComponent) OnInit(hook LifecycleHook) *LegacyComponent {
	trackV1Call("Component.OnInit")
	c.hooks = append(c.hooks, hook)
	return c
}

// OnMount registers mount hooks using V2 owner context
func (c *LegacyComponent) OnMount(hook LifecycleHook) *LegacyComponent {
	trackV1Call("Component.OnMount")

	// Convert V1 hook to V2 OnMount within owner context
	if c.v2Owner != nil {
		// Use V2 OnMount which provides automatic cleanup
		OnMount(hook)
	}

	c.mountHooks = append(c.mountHooks, hook)
	return c
}

// OnDismount registers dismount hooks using V2 cleanup system
func (c *LegacyComponent) OnDismount(hook LifecycleHook) *LegacyComponent {
	trackV1Call("Component.OnDismount")

	// Convert V1 dismount to V2 cleanup within owner context
	if c.v2Owner != nil {
		// Use V2 OnCleanup which provides automatic scheduling
		OnCleanup(hook)
	}

	c.dismountHooks = append(c.dismountHooks, hook)
	return c
}

// Render provides V1 API compatibility
func (c *LegacyComponent) Render() gomponents.Node {
	trackV1Call("Component.Render")

	// Execute init hooks if not already initialized
	if !c.initialized {
		for _, hook := range c.hooks {
			hook()
		}
		c.initialized = true
	}

	return c.renderV1()
}

// ------------------------------------
// 🔍 Effect Bridge
// ------------------------------------

// Watch creates a V1-compatible effect using V2 internals
func Watch(fn func()) {
	trackV1Call("Watch")
	atomic.AddInt64(&migrationMetrics.EffectExecutions, 1)

	// Use V2 CreateEffect with current owner context
	CreateEffect(fn, getCurrentOwner())
}

// ------------------------------------
// 🎛️ Migration Control
// ------------------------------------

// EnableMigrationBridge enables V1/V2 compatibility tracking
func EnableMigrationBridge() {
	atomic.StoreInt32(&migrationEnabled, 1)
}

// DisableMigrationBridge disables migration tracking (for pure V2 mode)
func DisableMigrationBridge() {
	atomic.StoreInt32(&migrationEnabled, 0)
}

// IsMigrationEnabled returns whether migration tracking is active
func IsMigrationEnabled() bool {
	return atomic.LoadInt32(&migrationEnabled) == 1
}

// ResetMigrationMetrics clears all migration tracking data
func ResetMigrationMetrics() {
	migrationMetrics.mutex.Lock()
	defer migrationMetrics.mutex.Unlock()

	atomic.StoreInt64(&migrationMetrics.SignalOperations, 0)
	atomic.StoreInt64(&migrationMetrics.EffectExecutions, 0)
	atomic.StoreInt64(&migrationMetrics.ComponentMounts, 0)
	atomic.StoreInt64(&migrationMetrics.MemoryAllocations, 0)
	atomic.StoreInt64(&migrationMetrics.CascadePrevented, 0)

	migrationMetrics.V1CallsRemaining = make(map[string]int64)
}

// GetMigrationProgress returns migration completion percentage
func GetMigrationProgress() float64 {
	metrics := GetMigrationMetrics()

	totalV1Calls := int64(0)
	for _, count := range metrics.V1CallsRemaining {
		totalV1Calls += count
	}

	if totalV1Calls == 0 {
		return 100.0 // Migration complete
	}

	// Calculate progress based on V2 vs V1 API usage ratio
	v2Operations := metrics.SignalOperations + metrics.EffectExecutions + metrics.ComponentMounts
	if v2Operations == 0 {
		return 0.0
	}

	progress := float64(v2Operations) / float64(v2Operations+totalV1Calls) * 100.0
	if progress > 100.0 {
		progress = 100.0
	}

	return progress
}

// ------------------------------------
// 🔧 Migration Utilities
// ------------------------------------

// WithLegacyMode executes a function with V1 compatibility enabled
func WithLegacyMode(fn func()) {
	wasEnabled := IsMigrationEnabled()
	EnableMigrationBridge()
	defer func() {
		if !wasEnabled {
			DisableMigrationBridge()
		}
	}()
	fn()
}

// WithV2Mode executes a function with pure V2 mode
func WithV2Mode(fn func()) {
	wasEnabled := IsMigrationEnabled()
	DisableMigrationBridge()
	defer func() {
		if wasEnabled {
			EnableMigrationBridge()
		}
	}()
	fn()
}

// ValidateMigrationCompatibility ensures V1 and V2 APIs produce equivalent results
func ValidateMigrationCompatibility[T comparable](
	v1Fn func() T,
	v2Fn func() T,
) bool {
	v1Result := v1Fn()
	v2Result := v2Fn()
	return v1Result == v2Result
}
