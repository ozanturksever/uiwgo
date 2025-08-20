// v2_signals.go
// V2 signal implementation with clean API separation

//go:build v2
// +build v2

package golid

import (
	"sync/atomic"
)

// ------------------------------------
// 🔧 V2 Signal Implementation
// ------------------------------------

// V2Signal represents the new SolidJS-inspired signal implementation
type V2Signal[T any] struct {
	getter func() T
	setter func(T)
	id     uint64
}

// CreateV2Signal creates a new V2 signal with fine-grained reactivity
func CreateV2Signal[T any](initial T) (func() T, func(T)) {
	// Use the existing CreateSignal from reactivity_core.go
	return CreateSignal(initial)
}

// ------------------------------------
// 🔄 V2 Effect Implementation
// ------------------------------------

// CreateV2Effect creates a V2 effect using owner context
func CreateV2Effect(fn func()) {
	CreateEffect(fn, getCurrentOwner())
}

// ------------------------------------
// 🧱 V2 Component Implementation
// ------------------------------------

// V2Component represents a component with automatic lifecycle management
type V2Component struct {
	render func() any
	owner  *Owner
}

// CreateV2Component creates a component with V2 patterns
func CreateV2Component(render func() any) *V2Component {
	var owner *Owner
	CreateOwner(func() {
		owner = getCurrentOwner()
	})

	return &V2Component{
		render: render,
		owner:  owner,
	}
}

// Render executes the component render function
func (c *V2Component) Render() any {
	return c.render()
}

// ------------------------------------
// 🔧 V2 Migration Helpers
// ------------------------------------

var v2MigrationActive = int32(0)

// EnableV2Migration activates V2 migration mode
func EnableV2Migration() {
	atomic.StoreInt32(&v2MigrationActive, 1)
}

// DisableV2Migration deactivates V2 migration mode
func DisableV2Migration() {
	atomic.StoreInt32(&v2MigrationActive, 0)
}

// IsV2MigrationActive returns whether V2 migration is active
func IsV2MigrationActive() bool {
	return atomic.LoadInt32(&v2MigrationActive) == 1
}
