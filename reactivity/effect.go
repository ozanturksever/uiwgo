package reactivity

import "syscall/js"

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
	IsDisposed() bool
}

// depNode is implemented by signals that can remove a dependent effect.
type depNode interface {
	removeEffect(eff *effect)
}

var currentEffect *effect

// CreateEffect registers a reactive effect that runs immediately and then
// re-runs whenever any of its dependent signals change.
// If there's a current cleanup scope, the effect will be automatically
// disposed when the scope is disposed.
func CreateEffect(fn func()) Effect {
	e := &effect{fn: fn, deps: make(map[depNode]struct{})}
	
	// Debug: Log effect creation
	if js.Global().Truthy() {
		js.Global().Get("console").Call("log", "[Effect] Creating effect")
	}
	
	// Register with current cleanup scope if available
	RegisterCleanup(func() {
		e.Dispose()
	})
	
	e.run()
	return e
}

// CreatePersistentEffect creates a reactive effect that persists across cleanup scopes.
// This is useful for components that need to remain active even when parent scopes are cleaned up.
// The caller is responsible for manually disposing this effect.
func CreatePersistentEffect(fn func()) Effect {
	e := &effect{fn: fn, deps: make(map[depNode]struct{})}
	
	// Debug: Log persistent effect creation
	if js.Global().Truthy() {
		js.Global().Get("console").Call("log", "[Effect] Creating persistent effect (no auto-cleanup)")
	}
	
	// Do NOT register with cleanup scope - this effect persists
	
	e.run()
	return e
}

func (e *effect) run() {
	if e.disposed {
		return
	}
	// Debug: Log effect execution
	if js.Global().Truthy() {
		js.Global().Get("console").Call("log", "[Effect] Effect running")
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
	if js.Global().Truthy() {
		js.Global().Get("console").Call("log", "[Effect] Setting current effect")
	}
	currentEffect = e
	e.fn()
	if js.Global().Truthy() {
		js.Global().Get("console").Call("log", "[Effect] Restoring previous effect")
	}
	currentEffect = prev
}

// Dispose stops the effect: runs final cleanups and detaches from dependencies.
func (e *effect) Dispose() {
	if e.disposed {
		return
	}
	// Debug: Log effect disposal
	if js.Global().Truthy() {
		js.Global().Get("console").Call("log", "[Effect] Effect being disposed")
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

// IsDisposed returns true if the effect has been disposed.
func (e *effect) IsDisposed() bool {
	return e.disposed
}
