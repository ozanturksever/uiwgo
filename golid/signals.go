// signals.go
// Reactive signals and state management system

package golid

import (
	"sync/atomic"
)

// ------------------------------------
// 📦 Reactive Signals (State Handling)
// ------------------------------------

// Global effect tracking
var currentEffect *effect
var effectDepth int
var maxEffectDepth = 50 // Prevent infinite effect cascades

// Effect execution limits to prevent goroutine explosion
var (
	activeEffects    int32
	maxActiveEffects int32 = 100 // Maximum concurrent effect goroutines
)

// NewSignal creates a new reactive signal with an initial value
func NewSignal[T any](initial T) *Signal[T] {
	return &Signal[T]{
		value:    initial,
		watchers: make(map[*effect]struct{}),
	}
}

// Get retrieves the current value of the signal and tracks dependencies
func (s *Signal[T]) Get() T {
	if currentEffect != nil {
		s.watchers[currentEffect] = struct{}{}
		currentEffect.deps[s] = struct{}{}
	}
	return s.value
}

// Set updates the signal value and triggers reactive effects
func (s *Signal[T]) Set(val T) {
	s.value = val

	// Trigger effects with controlled concurrency
	for e := range s.watchers {
		// Check if we're under the concurrent effect limit
		if atomic.LoadInt32(&activeEffects) < maxActiveEffects {
			atomic.AddInt32(&activeEffects, 1)
			go func(effect *effect) {
				defer atomic.AddInt32(&activeEffects, -1)
				runEffect(effect)
			}(e)
		} else {
			// Execute synchronously if we're at the limit
			runEffect(e)
		}
	}
}

// removeWatcher removes an effect from this signal's watchers
func (s *Signal[T]) removeWatcher(e *effect) {
	delete(s.watchers, e)
}

// Watch creates a reactive effect that runs when its dependencies change
func Watch(fn func()) {
	e := &effect{
		fn:   fn,
		deps: make(map[any]struct{}),
	}
	runEffect(e)
}

// runEffect executes an effect and tracks its dependencies
func runEffect(e *effect) {
	// Prevent infinite effect cascades
	if effectDepth >= maxEffectDepth {
		return
	}

	// Increment depth
	effectDepth++
	defer func() {
		effectDepth--
	}()

	// Store old dependencies for cleanup
	oldDeps := e.deps
	e.deps = make(map[any]struct{})

	// Run the effect function and track new dependencies
	currentEffect = e
	e.fn()
	currentEffect = nil

	// Clean up old dependencies that are no longer needed
	for dep := range oldDeps {
		// Only remove if this dependency is not in the new dependencies
		if _, stillNeeded := e.deps[dep]; !stillNeeded {
			if s, ok := dep.(hasWatchers); ok {
				s.removeWatcher(e)
			}
		}
	}
}
