package reactivity

import (
	"reflect"
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
		// Register dependency both ways
		s.deps[currentEffect] = struct{}{}
		currentEffect.deps[s] = struct{}{}
	}
	return s.value
}

func (s *baseSignal[T]) Set(v T) {
	if reflect.DeepEqual(s.value, v) {
		return
	}
	s.value = v
	// Re-run all dependent effects (iterate over a snapshot to avoid mutation issues)
	effects := make([]*effect, 0, len(s.deps))
	for e := range s.deps {
		effects = append(effects, e)
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
