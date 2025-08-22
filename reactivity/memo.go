package reactivity

import "reflect"

// memoSignal is a lazily computed derived signal.
type memoSignal[T any] struct {
	base        *baseSignal[T]
	calc        func() T
	initialized bool
	tracker     *effect
}

// CreateMemo creates a derived, cached signal. It defers the initial
// computation until the memo is first read with Get().
func CreateMemo[T any](fn func() T) Signal[T] {
	return &memoSignal[T]{
		base: &baseSignal[T]{deps: make(map[*effect]struct{})},
		calc: fn,
	}
}

func (m *memoSignal[T]) ensureTracker() {
	if m.tracker != nil {
		return
	}
	// tracker effect re-evaluates dependencies and updates value on changes
	m.tracker = CreateEffect(func() {
		newVal := m.calc()
		if !m.initialized {
			// First computation should not trigger dependents re-run immediately.
			m.base.value = newVal
			m.initialized = true
			return
		}
		if !reflect.DeepEqual(m.base.value, newVal) {
			m.base.Set(newVal)
		}
	}).(*effect)
}

func (m *memoSignal[T]) Get() T {
	// Register dependent effect on the base signal
	// But first ensure the value is computed at least once
	if !m.initialized {
		m.ensureTracker()
	}
	// Now normal dependency registration
	return m.base.Get()
}

func (m *memoSignal[T]) Set(v T) { m.base.Set(v) }

// removeEffect satisfies depNode via the embedded base behavior.
func (m *memoSignal[T]) removeEffect(eff *effect) { m.base.removeEffect(eff) }
