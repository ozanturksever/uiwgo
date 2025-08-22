package reactivity

// effect is the internal implementation of a reactive effect.
// Not concurrency-safe; designed for single-threaded JS/WASM and tests.
type effect struct {
	fn       func()
	disposed bool
	// deps holds the set of signals this effect currently depends on
	deps map[depNode]struct{}
	// cleanups are run before re-execution and at dispose
	cleanups []func()
}

// Effect represents a running reactive computation that can be disposed.
type Effect interface {
	Dispose()
}

// depNode is implemented by signals that can remove a dependent effect.
type depNode interface {
	removeEffect(eff *effect)
}

var currentEffect *effect

// CreateEffect registers a reactive effect that runs immediately and then
// re-runs whenever any of its dependent signals change.
func CreateEffect(fn func()) Effect {
	e := &effect{fn: fn, deps: make(map[depNode]struct{})}
	e.run()
	return e
}

func (e *effect) run() {
	if e.disposed {
		return
	}
	// Cleanup previous run
	for _, c := range e.cleanups {
		c()
	}
	e.cleanups = nil
	// Detach from previous dependencies
	for d := range e.deps {
		d.removeEffect(e)
	}
	e.deps = make(map[depNode]struct{})
	// Run with this effect set as current
	prev := currentEffect
	currentEffect = e
	e.fn()

	currentEffect = prev
}

// Dispose stops the effect: runs final cleanups and detaches from dependencies.
func (e *effect) Dispose() {
	if e.disposed {
		return
	}
	e.disposed = true
	for _, c := range e.cleanups {
		c()
	}
	e.cleanups = nil
	for d := range e.deps {
		d.removeEffect(e)
	}
	e.deps = nil
}
