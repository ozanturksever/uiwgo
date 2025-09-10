package action

import (
	"encoding/json"
	"fmt"

	"github.com/ozanturksever/uiwgo/comps"
	"github.com/ozanturksever/uiwgo/reactivity"
)

// Disposable represents a resource that can be disposed to clean up.
type Disposable interface {
	Dispose() error
}

// autoDisposable manages multiple disposables that should be disposed together.
type autoDisposable struct {
	disposables []Disposable
	disposed    bool
}

// newAutoDisposable creates a new autoDisposable.
func newAutoDisposable() *autoDisposable {
	return &autoDisposable{
		disposables: make([]Disposable, 0),
	}
}

// Add adds a disposable to be managed.
func (ad *autoDisposable) Add(disposable Disposable) {
	if ad.disposed {
		// If already disposed, dispose immediately
		disposable.Dispose()
		return
	}
	ad.disposables = append(ad.disposables, disposable)
}

// Dispose disposes all managed disposables.
func (ad *autoDisposable) Dispose() error {
	if ad.disposed {
		return nil
	}
	ad.disposed = true

	for _, disposable := range ad.disposables {
		disposable.Dispose()
	}
	ad.disposables = nil
	return nil
}

// AutoSubscribe creates a subscription that is automatically disposed when the component is unmounted.
// The subscribeFn function is called during mount to create subscriptions that will be tracked.
func AutoSubscribe(subscribeFn func() Disposable) Disposable {
	autoDisp := newAutoDisposable()

	// Check if we're in a reactive context, if not defer to OnMount
	if reactivity.GetCurrentCleanupScope() != nil {
		// We're in a reactive context, create the effect immediately
		reactivity.CreateEffect(func() {
			disposable := subscribeFn()
			autoDisp.Add(disposable)
			reactivity.OnCleanup(func() {
				autoDisp.Dispose()
			})
		})
	} else {
		// Defer to OnMount when we have proper cleanup scope
		comps.OnMount(func() {
			reactivity.CreateEffect(func() {
				disposable := subscribeFn()
				autoDisp.Add(disposable)
				reactivity.OnCleanup(func() {
					autoDisp.Dispose()
				})
			})
		})
	}

	return autoDisp
}

// WithLifecycleDispose registers a subscription for automatic disposal when the component is unmounted.
func WithLifecycleDispose(sub Subscription) Subscription {
	reactivity.OnCleanup(func() {
		sub.Dispose()
	})
	return sub
}

// OnAction registers an action handler that is automatically disposed when the component is unmounted.
// This is a high-level helper that combines subscription creation with lifecycle management.
func OnAction[T any](bus Bus, actionType ActionType[T], handler func(Context, T), opts ...SubOption) Subscription {
	var sub Subscription

	// Check if we're in a reactive context, if not defer to OnMount
	if reactivity.GetCurrentCleanupScope() != nil {
		// We're in a reactive context, create the effect immediately
		reactivity.CreateEffect(func() {
			sub = bus.Subscribe(actionType.Name, func(action Action[string]) error {
				var payload T
				if err := json.Unmarshal([]byte(action.Payload), &payload); err != nil {
					return fmt.Errorf("failed to unmarshal payload for action %s: %w", actionType.Name, err)
				}

				handler(Context{}, payload)
				return nil
			}, opts...)
			reactivity.OnCleanup(func() {
				sub.Dispose()
			})
		})
	} else {
		// Create a stub subscription for return value
		sub = &stubSubscription{}

		// Defer to OnMount when we have proper cleanup scope
		comps.OnMount(func() {
			reactivity.CreateEffect(func() {
				sub = bus.Subscribe(actionType.Name, func(action Action[string]) error {
					var payload T
					if err := json.Unmarshal([]byte(action.Payload), &payload); err != nil {
						return fmt.Errorf("failed to unmarshal payload for action %s: %w", actionType.Name, err)
					}

					handler(Context{}, payload)
					return nil
				}, opts...)
				reactivity.OnCleanup(func() {
					sub.Dispose()
				})
			})
		})
	}

	return sub
}

// UseActionSignal creates a signal that reflects the latest value of an action type and ensures proper cleanup.
// This is a high-level helper that combines signal creation with lifecycle management.
func UseActionSignal[T any](bus Bus, actionType ActionType[T], opts ...BridgeOption) reactivity.Signal[T] {
	var signal reactivity.Signal[T]

	// Check if we're in a reactive context, if not defer to OnMount
	if reactivity.GetCurrentCleanupScope() != nil {
		// We're in a reactive context, create the effect immediately
		reactivity.CreateEffect(func() {
			signal = ToSignal[T](bus, actionType.Name, opts...)
			reactivity.OnCleanup(func() {
				if bridgeSig, ok := signal.(Disposable); ok {
					bridgeSig.Dispose()
				}
			})
		})
	} else {
		// Create a default signal for return value
		signal = reactivity.CreateSignal[T](getZeroValue[T]())

		// Defer to OnMount when we have proper cleanup scope
		comps.OnMount(func() {
			reactivity.CreateEffect(func() {
				signal = ToSignal[T](bus, actionType.Name, opts...)
				reactivity.OnCleanup(func() {
					if bridgeSig, ok := signal.(Disposable); ok {
						bridgeSig.Dispose()
					}
				})
			})
		})
	}

	return signal
}

// stubSubscription is a placeholder subscription for cases where we return before actual creation
type stubSubscription struct{}

func (s *stubSubscription) Dispose() error { return nil }
func (s *stubSubscription) IsActive() bool { return false }

// getZeroValue returns the zero value for type T
func getZeroValue[T any]() T {
	var zero T
	return zero
}
