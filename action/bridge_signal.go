package action

import (
	"sync"

	"github.com/ozanturksever/uiwgo/reactivity"
)

// bridgeSignal wraps a reactivity.Signal and adds disposal functionality for the subscription.
type bridgeSignal[T any] struct {
	signal       reactivity.Signal[T]
	subscription Subscription
	mu           sync.Mutex
	disposed     bool
}

// Get returns the current value and registers the current running effect
// (if any) as a dependent of this signal.
func (bs *bridgeSignal[T]) Get() T {
	bs.mu.Lock()
	defer bs.mu.Unlock()

	if bs.disposed {
		var zero T
		return zero
	}

	return bs.signal.Get()
}

// Set updates the value. If the value hasn't changed (DeepEqual), it's a no-op.
// Otherwise all dependent effects are re-executed.
func (bs *bridgeSignal[T]) Set(value T) {
	bs.mu.Lock()
	defer bs.mu.Unlock()

	if bs.disposed {
		return
	}

	bs.signal.Set(value)
}

// Dispose disposes the bridge signal and its underlying subscription.
func (bs *bridgeSignal[T]) Dispose() {
	bs.mu.Lock()
	defer bs.mu.Unlock()

	if bs.disposed {
		return
	}

	bs.disposed = true

	// Dispose the subscription if it exists
	if bs.subscription != nil {
		bs.subscription.Dispose()
	}
}

// NewBridgeSignal creates a new bridge signal with the given signal and subscription.
func NewBridgeSignal[T any](signal reactivity.Signal[T], subscription Subscription) reactivity.Signal[T] {
	// Type assert to see if it's already a bridgeSignal
	if bridgeSig, ok := signal.(*bridgeSignal[T]); ok {
		// If it's already a bridge signal, just update the subscription
		bridgeSig.mu.Lock()
		bridgeSig.subscription = subscription
		bridgeSig.mu.Unlock()
		return bridgeSig
	}

	// Otherwise create a new bridge signal
	return &bridgeSignal[T]{
		signal:       signal,
		subscription: subscription,
	}
}
