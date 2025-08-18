// signals.go
// Reactive signals and state management system

package golid

// ------------------------------------
// 📦 Reactive Signals (State Handling)
// ------------------------------------

// Global effect tracking
var currentEffect *effect

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
	for e := range s.watchers {
		go runEffect(e)
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
	// Clean up old dependencies
	for dep := range e.deps {
		if s, ok := dep.(hasWatchers); ok {
			s.removeWatcher(e)
		}
	}
	e.deps = make(map[any]struct{})

	// Run the effect function and track new dependencies
	currentEffect = e
	e.fn()
	currentEffect = nil
}
