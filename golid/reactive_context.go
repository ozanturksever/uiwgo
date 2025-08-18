// reactive_context.go
// Global reactive context and computation management

package golid

import (
	"sync"
	"sync/atomic"
)

// ------------------------------------
// 🌐 Global Reactive Context
// ------------------------------------

var (
	// Global context for tracking current computation and owner
	reactiveContext = &ReactiveContext{
		computationStack: make([]*Computation, 0),
		ownerStack:       make([]*Owner, 0),
	}
	reactiveContextMutex sync.RWMutex
)

// ReactiveContext manages the global reactive state
type ReactiveContext struct {
	currentComputation *Computation
	currentOwner       *Owner
	computationStack   []*Computation
	ownerStack         []*Owner
	mutex              sync.RWMutex
}

// ------------------------------------
// 🔄 Computation Context Management
// ------------------------------------

// getCurrentComputation returns the currently executing computation
func getCurrentComputation() *Computation {
	reactiveContextMutex.RLock()
	defer reactiveContextMutex.RUnlock()
	return reactiveContext.currentComputation
}

// setCurrentComputation sets the currently executing computation
func setCurrentComputation(comp *Computation) {
	reactiveContextMutex.Lock()
	defer reactiveContextMutex.Unlock()

	if comp != nil {
		// Push current computation to stack
		if reactiveContext.currentComputation != nil {
			reactiveContext.computationStack = append(reactiveContext.computationStack, reactiveContext.currentComputation)
		}
		reactiveContext.currentComputation = comp
	} else {
		// Pop from stack
		if len(reactiveContext.computationStack) > 0 {
			reactiveContext.currentComputation = reactiveContext.computationStack[len(reactiveContext.computationStack)-1]
			reactiveContext.computationStack = reactiveContext.computationStack[:len(reactiveContext.computationStack)-1]
		} else {
			reactiveContext.currentComputation = nil
		}
	}
}

// ------------------------------------
// 🏠 Owner Context Management
// ------------------------------------

// getCurrentOwner returns the current owner context
func getCurrentOwner() *Owner {
	reactiveContextMutex.RLock()
	defer reactiveContextMutex.RUnlock()
	return reactiveContext.currentOwner
}

// setCurrentOwner sets the current owner context
func setCurrentOwner(owner *Owner) {
	reactiveContextMutex.Lock()
	defer reactiveContextMutex.Unlock()

	if owner != nil {
		// Push current owner to stack
		if reactiveContext.currentOwner != nil {
			reactiveContext.ownerStack = append(reactiveContext.ownerStack, reactiveContext.currentOwner)
		}
		reactiveContext.currentOwner = owner
	} else {
		// Pop from stack
		if len(reactiveContext.ownerStack) > 0 {
			reactiveContext.currentOwner = reactiveContext.ownerStack[len(reactiveContext.ownerStack)-1]
			reactiveContext.ownerStack = reactiveContext.ownerStack[:len(reactiveContext.ownerStack)-1]
		} else {
			reactiveContext.currentOwner = nil
		}
	}
}

// ------------------------------------
// 🔄 Tracking Utilities
// ------------------------------------

// Track runs a function and tracks all signal accesses within it
func Track[T any](fn func() T) T {
	// Create a temporary computation for tracking
	comp := &Computation{
		id:           atomic.AddUint64(&computationIdCounter, 1),
		dependencies: make(map[uint64]Dependency),
		state:        Clean,
		context:      make(map[string]interface{}),
	}

	prevComputation := getCurrentComputation()
	setCurrentComputation(comp)
	defer func() {
		setCurrentComputation(prevComputation)
		// Cleanup the temporary computation
		comp.cleanup()
	}()

	return fn()
}

// Batch runs a function within a batched update context
func Batch(fn func()) {
	getScheduler().batch(fn)
}

// ------------------------------------
// 🔄 Transition Management
// ------------------------------------

var transitionIdCounter uint64

// Transition represents a batched update transition
type Transition struct {
	id        uint64
	pending   []*Update
	effects   []*Computation
	completed chan bool
	mutex     sync.RWMutex
}

// Update represents a pending signal update
type Update struct {
	signalId uint64
	value    interface{}
}

// StartTransition creates a new transition for batched updates
func StartTransition() *Transition {
	return &Transition{
		id:        atomic.AddUint64(&transitionIdCounter, 1),
		pending:   make([]*Update, 0),
		effects:   make([]*Computation, 0),
		completed: make(chan bool, 1),
	}
}

// Complete finalizes the transition and applies all updates
func (t *Transition) Complete() {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	// Apply all pending updates
	for _, update := range t.pending {
		// Apply update logic here
		_ = update // TODO: Implement update application
	}

	// Schedule all effects
	for _, effect := range t.effects {
		effect.markDirty()
	}

	// Signal completion
	select {
	case t.completed <- true:
	default:
	}
}

// ------------------------------------
// 🧹 Context Cleanup
// ------------------------------------

// ResetReactiveContext resets the global reactive context (for testing)
func ResetReactiveContext() {
	reactiveContextMutex.Lock()
	defer reactiveContextMutex.Unlock()

	reactiveContext.currentComputation = nil
	reactiveContext.currentOwner = nil
	reactiveContext.computationStack = reactiveContext.computationStack[:0]
	reactiveContext.ownerStack = reactiveContext.ownerStack[:0]
}

// GetReactiveStats returns statistics about the current reactive context
func GetReactiveStats() map[string]interface{} {
	reactiveContextMutex.RLock()
	defer reactiveContextMutex.RUnlock()

	return map[string]interface{}{
		"currentComputation":   reactiveContext.currentComputation != nil,
		"currentOwner":         reactiveContext.currentOwner != nil,
		"computationStackSize": len(reactiveContext.computationStack),
		"ownerStackSize":       len(reactiveContext.ownerStack),
	}
}
