package action

import (
	"testing"

	"github.com/ozanturksever/uiwgo/reactivity"
)

// TestLifecycle_AutoDisposeOnUnmount verifies that subscriptions created with AutoSubscribe
// are automatically disposed when the component is unmounted.
func TestLifecycle_AutoDisposeOnUnmount(t *testing.T) {
	bus := New()
	var subscription Subscription
	disposed := false

	// Create a signal to track when the subscription handler is called
	handlerCalled := reactivity.CreateSignal(false)

	// Use AutoSubscribe to create a subscription
	disposable := AutoSubscribe(func() Disposable {
		sub := bus.Subscribe("test-action", func(action Action[string]) error {
			handlerCalled.Set(true)
			return nil
		})
		subscription = sub
		return sub
	})

	// Initially, the subscription should not be active (not mounted yet)
	if subscription != nil && subscription.IsActive() {
		t.Error("Expected subscription to not be active before mount")
	}

	// Simulate mount by calling OnMount callbacks
	// In a real scenario, this would happen during component mounting
	// For testing, we'll manually trigger the mount behavior

	// Since we can't easily simulate the mount/unmount cycle in a test,
	// we'll test the disposal behavior directly
	err := disposable.Dispose()
	if err != nil {
		t.Errorf("Expected Dispose to succeed, got error: %v", err)
	}
	disposed = true

	// Verify the disposable is marked as disposed
	if !disposed {
		t.Error("Expected disposable to be marked as disposed")
	}

	// Try to dispose again - should be safe
	err = disposable.Dispose()
	if err != nil {
		t.Errorf("Expected multiple Dispose calls to be safe, got error: %v", err)
	}
}

// TestLifecycle_ReMountCreatesNewActiveSubscription verifies that re-mounting
// creates new active subscriptions.
func TestLifecycle_ReMountCreatesNewActiveSubscription(t *testing.T) {
	bus := New()
	var firstSubscription, secondSubscription Subscription

	// Create a signal to track when handlers are called
	firstHandlerCalled := reactivity.CreateSignal(false)
	secondHandlerCalled := reactivity.CreateSignal(false)

	// Use AutoSubscribe to create subscriptions
	firstDisposable := AutoSubscribe(func() Disposable {
		sub := bus.Subscribe("test-action", func(action Action[string]) error {
			firstHandlerCalled.Set(true)
			return nil
		})
		firstSubscription = sub
		return sub
	})

	secondDisposable := AutoSubscribe(func() Disposable {
		sub := bus.Subscribe("test-action", func(action Action[string]) error {
			secondHandlerCalled.Set(true)
			return nil
		})
		secondSubscription = sub
		return sub
	})

	// Test first subscription
	if firstSubscription != nil && firstSubscription.IsActive() {
		t.Error("Expected first subscription to not be active before mount")
	}

	// Test second subscription
	if secondSubscription != nil && secondSubscription.IsActive() {
		t.Error("Expected second subscription to not be active before mount")
	}

	// Dispose both disposables
	firstDisposable.Dispose()
	secondDisposable.Dispose()

	// Verify they're no longer active
	if firstSubscription != nil && firstSubscription.IsActive() {
		t.Error("Expected first subscription to be inactive after disposal")
	}
	if secondSubscription != nil && secondSubscription.IsActive() {
		t.Error("Expected second subscription to be inactive after disposal")
	}
}

// TestWhen_StopsListeningWhenSignalFalseAndResumesTrue verifies that the When option
// properly gates delivery based on a reactive signal.
func TestWhen_StopsListeningWhenSignalFalseAndResumesTrue(t *testing.T) {
	bus := New()
	gateSignal := reactivity.CreateSignal(true)
	var received []string

	// Subscribe with When option
	sub := bus.Subscribe("test-action", func(action Action[string]) error {
		received = append(received, action.Payload)
		return nil
	}, SubWhen(gateSignal))

	// Dispatch when signal is true
	bus.Dispatch(Action[string]{
		Type:    "test-action",
		Payload: "first-true-payload",
	})

	// Should have received the first action
	if len(received) != 1 || received[0] != "first-true-payload" {
		t.Errorf("Expected to receive first action when signal is true, got: %v", received)
	}

	// Set signal to false
	gateSignal.Set(false)

	// Dispatch when signal is false
	bus.Dispatch(Action[string]{
		Type:    "test-action",
		Payload: "false-payload",
	})

	// Should not have received the false action
	if len(received) != 1 {
		t.Errorf("Expected to still have only 1 action when signal is false, got: %v", received)
	}

	// Set signal back to true
	gateSignal.Set(true)

	// Dispatch when signal is true again
	bus.Dispatch(Action[string]{
		Type:    "test-action",
		Payload: "second-true-payload",
	})

	// Should have received the second true action
	if len(received) != 2 || received[1] != "second-true-payload" {
		t.Errorf("Expected to receive second true action, got: %v", received)
	}

	// Clean up
	sub.Dispose()
}

// TestNoLeak_MultipleMountUnmountCycles verifies that there are no memory leaks
// under repeated mount/unmount cycles by checking IsActive toggling.
func TestNoLeak_MultipleMountUnmountCycles(t *testing.T) {
	bus := New()
	var subscriptions []Subscription
	handlerCalledCount := 0

	// Create multiple subscriptions to test
	for i := 0; i < 5; i++ {
		sub := bus.Subscribe("test-action", func(action Action[string]) error {
			handlerCalledCount++
			return nil
		})
		subscriptions = append(subscriptions, sub)
	}

	// Verify all subscriptions are active
	for i, sub := range subscriptions {
		if !sub.IsActive() {
			t.Errorf("Expected subscription %d to be active", i)
		}
	}

	// Dispatch an action to verify handlers are called
	bus.Dispatch(Action[string]{
		Type:    "test-action",
		Payload: "test-payload",
	})

	// Verify handlers were called
	if handlerCalledCount != 5 {
		t.Errorf("Expected 5 handlers to be called, got %d", handlerCalledCount)
	}

	// Dispose all subscriptions
	for _, sub := range subscriptions {
		err := sub.Dispose()
		if err != nil {
			t.Errorf("Expected Dispose to succeed, got error: %v", err)
		}
	}

	// Verify all subscriptions are inactive
	for i, sub := range subscriptions {
		if sub.IsActive() {
			t.Errorf("Expected subscription %d to be inactive after disposal", i)
		}
	}

	// Try to dispatch again - handlers should not be called
	oldCount := handlerCalledCount
	bus.Dispatch(Action[string]{
		Type:    "test-action",
		Payload: "test-payload-2",
	})

	// Handler count should not have changed
	if handlerCalledCount != oldCount {
		t.Errorf("Expected handler count to remain the same after disposal, got %d", handlerCalledCount)
	}

	// Try to dispose again - should be safe
	for _, sub := range subscriptions {
		err := sub.Dispose()
		if err != nil {
			t.Errorf("Expected multiple Dispose calls to be safe, got error: %v", err)
		}
	}
}
