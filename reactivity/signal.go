package reactivity

import (
	"fmt"
	"reflect"
	"syscall/js"
)

// Signal is the basic reactive primitive. It holds a value and notifies
// observers when that value changes.
type Signal[T any] interface {
	// Get returns the current value and registers the current running effect
	// (if any) as a dependent of this signal.
	Get() T
	// Set updates the value. If the value hasn't changed (DeepEqual), it's a no-op.
	// Otherwise all dependent effects are re-executed.
	Set(value T)
}

// baseSignal implements Signal and tracks dependent effects.
// It's intentionally minimal and not concurrency-safe for MVP.
type baseSignal[T any] struct {
	value T
	// deps tracks effects depending on this signal
	deps map[*effect]struct{}
}

// removeEffect detaches the given effect from this signal's dependency list.
func (s *baseSignal[T]) removeEffect(eff *effect) {
	delete(s.deps, eff)
}

func CreateSignal[T any](initial T) Signal[T] {
	return &baseSignal[T]{
		value: initial,
		deps:  make(map[*effect]struct{}),
	}
}

func (s *baseSignal[T]) Get() T {
	if currentEffect != nil && !currentEffect.disposed {
		// Debug: Log dependency registration
		if js.Global().Truthy() {
			js.Global().Get("console").Call("log", "[Signal] Registering effect dependency")
		}
		// Register dependency both ways
		s.deps[currentEffect] = struct{}{}
		currentEffect.deps[s] = struct{}{}
	} else {
		// Debug: Log when no current effect
		if js.Global().Truthy() {
			js.Global().Get("console").Call("log", "[Signal] No current effect to register")
		}
	}
	return s.value
}

func (s *baseSignal[T]) Set(v T) {
	isEqual := reflect.DeepEqual(s.value, v)
	// Debug: Log the comparison for slices
	if js.Global().Truthy() {
		js.Global().Get("console").Call("log", fmt.Sprintf("[Signal] Set called. Equal: %v", isEqual))
	}
	if isEqual {
		return
	}
	s.value = v
	// Re-run all dependent effects (iterate over a snapshot to avoid mutation issues)
	effects := make([]*effect, 0, len(s.deps))
	for e := range s.deps {
		effects = append(effects, e)
	}
	if js.Global().Truthy() {
		js.Global().Get("console").Call("log", fmt.Sprintf("[Signal] Triggering %d effects", len(effects)))
	}
	for _, e := range effects {
		if e.disposed {
			delete(s.deps, e)
			continue
		}
		e.run()
	}
}

// sAny is a tiny helper to coerce generic signal to any for effect deps map.
func sAny[T any](s *baseSignal[T]) any { return any(s) }
