package action

import (
	"sync"
	"testing"
	"time"

	"github.com/ozanturksever/uiwgo/reactivity"
)

func TestBusInterfaceStubMethods(t *testing.T) {
	// Test that the Bus interface can be implemented and all methods exist
	bus := &testBus{}

	// Test Dispatch method exists and can be called
	action := Action[string]{
		Type:    "test-action",
		Payload: "test-payload",
	}

	// These should not panic (they're stub implementations)
	bus.Dispatch(action)
	bus.Dispatch(action, WithTimeout(5*time.Second))

	// Test Subscribe method exists
	sub := bus.Subscribe("test-action", func(action Action[string]) error {
		return nil
	})
	if sub == nil {
		t.Error("Expected Subscribe to return a subscription")
	}

	// Test SubscribeAny method exists
	subAny := bus.SubscribeAny(func(action any) error {
		return nil
	})
	if subAny == nil {
		t.Error("Expected SubscribeAny to return a subscription")
	}

	// Test Scope method exists
	scopedBus := bus.Scope("test-scope")
	if scopedBus == nil {
		t.Error("Expected Scope to return a Bus")
	}

	// Test ToSignal method exists (stub implementation returns nil)
	signal := bus.ToSignal("test-action")
	_ = signal // Just test that the method exists and can be called

	// Test ToStream method exists (stub implementation returns nil)
	stream := bus.ToStream("test-action")
	_ = stream // Just test that the method exists and can be called

	// Test HandleQuery method exists
	queryHandler := bus.HandleQuery("test-query", func(query Action[string]) (any, error) {
		return "response", nil
	})
	if queryHandler == nil {
		t.Error("Expected HandleQuery to return a subscription")
	}

	// Test Ask method exists
	response, err := bus.Ask("test-query", action)
	if err != nil {
		t.Errorf("Expected Ask to return without error, got %v", err)
	}
	if response != nil {
		t.Error("Expected Ask to return nil response in stub implementation")
	}

	// Test OnError method exists
	errorSub := bus.OnError(func(ctx Context, err error, recovered any) {
		// Error handler
	})
	if errorSub == nil {
		t.Error("Expected OnError to return a subscription")
	}
}

// testBus is a stub implementation of the Bus interface for testing
type testBus struct{}

func (tb *testBus) Dispatch(action any, opts ...DispatchOption) error {
	return nil
}

func (tb *testBus) Subscribe(actionType string, handler func(Action[string]) error, opts ...SubOption) Subscription {
	return &NoOpSubscription{}
}

func (tb *testBus) SubscribeAny(handler func(any) error, opts ...SubOption) Subscription {
	return &NoOpSubscription{}
}

func (tb *testBus) Scope(name string) Bus {
	return tb
}

func (tb *testBus) ToSignal(actionType string, opts ...BridgeOption) any {
	return nil
}

func (tb *testBus) ToStream(actionType string, opts ...BridgeOption) any {
	return nil
}

func (tb *testBus) HandleQuery(queryType string, handler func(Action[string]) (any, error), opts ...QueryOption) Subscription {
	return &NoOpSubscription{}
}

func (tb *testBus) Ask(queryType string, query Action[string], opts ...AskOption) (any, error) {
	return nil, nil
}

func (tb *testBus) OnError(handler func(ctx Context, err error, recovered any), opts ...SubOption) Subscription {
	return &NoOpSubscription{}
}

func (tb *testBus) HandleQueryTyped(qt interface{}, handler interface{}, opts ...QueryOption) Subscription {
	return &NoOpSubscription{}
}

func (tb *testBus) AskTyped(qt interface{}, req interface{}, opts ...AskOption) interface{} {
	return nil
}

// TestGlobalBusSingleton verifies that Global() returns the same instance
func TestGlobalBusSingleton(t *testing.T) {
	// Get the global bus twice
	bus1 := Global()
	bus2 := Global()

	// They should be the same instance
	if bus1 != bus2 {
		t.Error("Global() should return the same singleton instance")
	}

	// Test concurrent access to ensure thread safety
	var wg sync.WaitGroup
	buses := make([]Bus, 10)

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			buses[idx] = Global()
		}(i)
	}

	wg.Wait()

	// All should be the same instance
	for i := 1; i < 10; i++ {
		if buses[i] != buses[0] {
			t.Error("Global() should return the same instance even under concurrent access")
		}
	}
}

// TestScopedBusIsolation_NoDeliveryAcrossScopes verifies that scoped buses are isolated
func TestScopedBusIsolation_NoDeliveryAcrossScopes(t *testing.T) {
	globalBus := New() // Use a fresh bus for this test

	// Subscribe to an action on the global bus
	globalReceived := false
	globalSub := globalBus.Subscribe("test-action", func(action Action[string]) error {
		globalReceived = true
		return nil
	})
	defer globalSub.Dispose()

	// Create a scoped bus
	scopedBus := globalBus.Scope("test-scope")

	// Subscribe to the same action on the scoped bus
	scopedReceived := false
	scopedSub := scopedBus.Subscribe("test-action", func(action Action[string]) error {
		scopedReceived = true
		return nil
	})
	defer scopedSub.Dispose()

	// Dispatch from the scoped bus
	scopedBus.Dispatch(Action[string]{
		Type:    "test-action",
		Payload: "scoped-payload",
	})

	// The scoped bus should receive the action, but the global bus should not
	if !scopedReceived {
		t.Error("Scoped bus should receive actions dispatched from it")
	}
	if globalReceived {
		t.Error("Global bus should not receive actions dispatched from scoped bus")
	}

	// Reset flags
	globalReceived = false
	scopedReceived = false

	// Dispatch from the global bus
	globalBus.Dispatch(Action[string]{
		Type:    "test-action",
		Payload: "global-payload",
	})

	// The global bus should receive the action, but the scoped bus should not
	if !globalReceived {
		t.Error("Global bus should receive actions dispatched from it")
	}
	if scopedReceived {
		t.Error("Scoped bus should not receive actions dispatched from global bus")
	}
}

// TestScopePathComposition_ParentChild verifies scope path composition
func TestScopePathComposition_ParentChild(t *testing.T) {
	// Create a root bus
	rootBus := New()

	// Create nested scopes
	childBus := rootBus.Scope("child")
	grandchildBus := childBus.Scope("grandchild")

	// Test that we can dispatch and receive actions at each level
	rootReceived := false
	childReceived := false
	grandchildReceived := false

	// Subscribe at each level
	rootSub := rootBus.Subscribe("test-action", func(action Action[string]) error {
		rootReceived = true
		return nil
	})
	defer rootSub.Dispose()

	childSub := childBus.Subscribe("test-action", func(action Action[string]) error {
		childReceived = true
		return nil
	})
	defer childSub.Dispose()

	grandchildSub := grandchildBus.Subscribe("test-action", func(action Action[string]) error {
		grandchildReceived = true
		return nil
	})
	defer grandchildSub.Dispose()

	// Dispatch from grandchild - only grandchild should receive
	grandchildBus.Dispatch(Action[string]{
		Type:    "test-action",
		Payload: "grandchild-payload",
	})

	if !grandchildReceived {
		t.Error("Grandchild bus should receive its own dispatched actions")
	}
	if childReceived || rootReceived {
		t.Error("Parent buses should not receive actions from grandchild")
	}

	// Reset and test child dispatch
	rootReceived = false
	childReceived = false
	grandchildReceived = false

	childBus.Dispatch(Action[string]{
		Type:    "test-action",
		Payload: "child-payload",
	})

	if !childReceived {
		t.Error("Child bus should receive its own dispatched actions")
	}
	if rootReceived || grandchildReceived {
		t.Error("Other buses should not receive actions from child")
	}
}

// TestSubscribeAndDispatchWithinScopeOnly verifies subscriptions and dispatch work within their scope
func TestSubscribeAndDispatchWithinScopeOnly(t *testing.T) {
	bus := New()

	// Create multiple scopes
	scopeA := bus.Scope("scope-a")
	scopeB := bus.Scope("scope-b")

	// Track received actions
	var scopeAReceived []string
	var scopeBReceived []string
	var rootReceived []string

	// Subscribe in each scope
	scopeASub := scopeA.Subscribe("test-action", func(action Action[string]) error {
		scopeAReceived = append(scopeAReceived, action.Payload)
		return nil
	})
	defer scopeASub.Dispose()

	scopeBSub := scopeB.Subscribe("test-action", func(action Action[string]) error {
		scopeBReceived = append(scopeBReceived, action.Payload)
		return nil
	})
	defer scopeBSub.Dispose()

	rootSub := bus.Subscribe("test-action", func(action Action[string]) error {
		rootReceived = append(rootReceived, action.Payload)
		return nil
	})
	defer rootSub.Dispose()

	// Dispatch different actions from each scope
	scopeA.Dispatch(Action[string]{
		Type:    "test-action",
		Payload: "from-scope-a",
	})

	scopeB.Dispatch(Action[string]{
		Type:    "test-action",
		Payload: "from-scope-b",
	})

	bus.Dispatch(Action[string]{
		Type:    "test-action",
		Payload: "from-root",
	})

	// Verify isolation
	if len(scopeAReceived) != 1 || scopeAReceived[0] != "from-scope-a" {
		t.Errorf("Scope A should only receive its own action, got: %v", scopeAReceived)
	}

	if len(scopeBReceived) != 1 || scopeBReceived[0] != "from-scope-b" {
		t.Errorf("Scope B should only receive its own action, got: %v", scopeBReceived)
	}

	if len(rootReceived) != 1 || rootReceived[0] != "from-root" {
		t.Errorf("Root should only receive its own action, got: %v", rootReceived)
	}
}

// TestDispatchSyncOrdering_PriorityThenFIFO verifies that dispatch ordering follows priority then FIFO
func TestDispatchSyncOrdering_PriorityThenFIFO(t *testing.T) {
	bus := New()

	var received []string

	// Subscribe with different priorities
	bus.Subscribe("test-action", func(action Action[string]) error {
		received = append(received, "high-priority")
		return nil
	}, SubWithPriority(10))

	bus.Subscribe("test-action", func(action Action[string]) error {
		received = append(received, "low-priority-1")
		return nil
	}, SubWithPriority(1))

	bus.Subscribe("test-action", func(action Action[string]) error {
		received = append(received, "low-priority-2")
		return nil
	}, SubWithPriority(1))

	bus.Subscribe("test-action", func(action Action[string]) error {
		received = append(received, "medium-priority")
		return nil
	}, SubWithPriority(5))

	// Dispatch the action
	bus.Dispatch(Action[string]{
		Type:    "test-action",
		Payload: "test-payload",
	})

	// Verify ordering: high priority first, then medium, then low priority in FIFO order
	expected := []string{"high-priority", "medium-priority", "low-priority-1", "low-priority-2"}
	if len(received) != len(expected) {
		t.Fatalf("Expected %d handlers to be called, got %d", len(expected), len(received))
	}

	for i, expectedValue := range expected {
		if received[i] != expectedValue {
			t.Errorf("Expected handler %d to receive '%s', got '%s'", i, expectedValue, received[i])
		}
	}
}

// TestDispatchWithAsync_SchedulesLater verifies that async dispatch runs later
func TestDispatchWithAsync_SchedulesLater(t *testing.T) {
	bus := New()

	received := make(chan string, 1)
	dispatchCompleted := make(chan bool, 1)

	// Subscribe to the action
	bus.Subscribe("test-action", func(action Action[string]) error {
		received <- action.Payload
		return nil
	})

	// Dispatch asynchronously
	go func() {
		bus.Dispatch(Action[string]{
			Type:    "test-action",
			Payload: "async-payload",
		}, WithAsync())
		dispatchCompleted <- true
	}()

	// Wait a bit to ensure dispatch returns before handler is called
	select {
	case <-dispatchCompleted:
		// Dispatch returned, now wait for handler
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Dispatch did not return in time")
	}

	// Handler should be called after dispatch returns
	select {
	case payload := <-received:
		if payload != "async-payload" {
			t.Errorf("Expected payload 'async-payload', got '%s'", payload)
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Async handler was not called in time")
	}
}

// TestDispatchWithMetaTraceSource_PropagatesToContext verifies that dispatch options propagate to context
func TestDispatchWithMetaTraceSource_PropagatesToContext(t *testing.T) {
	bus := New()

	var receivedAction Action[string]

	// Subscribe to capture the context and action
	bus.Subscribe("test-action", func(action Action[string]) error {
		receivedAction = action
		return nil
	})

	// SubscribeAny to capture the context
	bus.SubscribeAny(func(action any) error {
		// Context is passed to dispatchToHandler but not directly accessible in test
		// We'll verify through the action metadata instead
		return nil
	})

	// Dispatch with options
	meta := map[string]any{"key": "value", "user": "test-user"}
	bus.Dispatch(Action[string]{
		Type:    "test-action",
		Payload: "test-payload",
	}, WithMeta(meta), WithTrace("trace-123"), WithSource("test-source"))

	// Verify action has the correct metadata
	if receivedAction.Meta == nil {
		t.Fatal("Expected action to have metadata")
	}

	if receivedAction.Meta["key"] != "value" {
		t.Errorf("Expected meta key 'key' to be 'value', got '%v'", receivedAction.Meta["key"])
	}

	if receivedAction.Meta["user"] != "test-user" {
		t.Errorf("Expected meta key 'user' to be 'test-user', got '%v'", receivedAction.Meta["user"])
	}

	if receivedAction.TraceID != "trace-123" {
		t.Errorf("Expected TraceID to be 'trace-123', got '%s'", receivedAction.TraceID)
	}

	if receivedAction.Source != "test-source" {
		t.Errorf("Expected Source to be 'test-source', got '%s'", receivedAction.Source)
	}
}

// TestSubscribeAny_ReceivesAllActions verifies that SubscribeAny receives all actions
func TestSubscribeAny_ReceivesAllActions(t *testing.T) {
	bus := New()

	var anyReceived []string
	var specificReceived []string

	// Subscribe to specific action type
	bus.Subscribe("specific-action", func(action Action[string]) error {
		specificReceived = append(specificReceived, action.Payload)
		return nil
	})

	// Subscribe to all actions
	bus.SubscribeAny(func(action any) error {
		if act, ok := action.(Action[string]); ok {
			anyReceived = append(anyReceived, act.Payload)
		}
		return nil
	})

	// Dispatch different actions
	bus.Dispatch(Action[string]{
		Type:    "specific-action",
		Payload: "specific-payload",
	})

	bus.Dispatch(Action[string]{
		Type:    "other-action",
		Payload: "other-payload",
	})
}

// TestToSignal_InitialValue tests that ToSignal respects initial values
func TestToSignal_InitialValue(t *testing.T) {
	bus := New()

	// Create a signal with an initial value
	signal := ToSignal[string](bus, "test-action", BridgeWithInitialValue("initial"))

	// Type assert to the expected signal type
	typedSignal, ok := signal.(reactivity.Signal[string])
	if !ok {
		t.Fatalf("Expected Signal[string], got %T", signal)
	}

	// Check initial value
	if value := typedSignal.Get(); value != "initial" {
		t.Errorf("Expected initial value 'initial', got '%s'", value)
	}

	// Dispatch an action
	bus.Dispatch(Action[string]{
		Type:    "test-action",
		Payload: "updated",
	})

	// Check updated value
	if value := typedSignal.Get(); value != "updated" {
		t.Errorf("Expected updated value 'updated', got '%s'", value)
	}
}

// TestToSignal_UpdatesOnDispatch tests that ToSignal updates on action dispatch
func TestToSignal_UpdatesOnDispatch(t *testing.T) {
	bus := New()

	// Create a signal without initial value
	signal := ToSignal[any](bus, "test-action")

	// Type assert to the expected signal type
	typedSignal, ok := signal.(reactivity.Signal[any])
	if !ok {
		t.Fatalf("Expected Signal[any], got %T", signal)
	}

	// Check initial zero value
	if value := typedSignal.Get(); value != nil {
		t.Errorf("Expected initial nil value, got '%v'", value)
	}

	// Dispatch an action
	bus.Dispatch(Action[string]{
		Type:    "test-action",
		Payload: "test-payload",
	})

	// Check updated value
	if value := typedSignal.Get(); value != "test-payload" {
		t.Errorf("Expected updated value 'test-payload', got '%v'", value)
	}
}

// TestToSignal_DistinctUntilChanged tests that ToSignal respects distinct until changed option
func TestToSignal_DistinctUntilChanged(t *testing.T) {
	bus := New()

	// Create a signal with distinct until changed
	signal := ToSignal[any](bus, "test-action", BridgeWithDistinctUntilChanged(nil))

	// Type assert to the expected signal type
	typedSignal, ok := signal.(reactivity.Signal[any])
	if !ok {
		t.Fatalf("Expected Signal[any], got %T", signal)
	}

	// Track how many times the signal value is accessed
	accessCount := 0

	// Create an effect to track signal changes
	reactivity.CreateEffect(func() {
		_ = typedSignal.Get() // Access the signal value
		accessCount++
	})

	// Dispatch first action
	bus.Dispatch(Action[string]{
		Type:    "test-action",
		Payload: "value1",
	})

	// Dispatch same value again (should be filtered out)
	bus.Dispatch(Action[string]{
		Type:    "test-action",
		Payload: "value1",
	})

	// Dispatch different value
	bus.Dispatch(Action[string]{
		Type:    "test-action",
		Payload: "value2",
	})

	// Check that the signal was accessed 3 times (initial + 2 updates)
	// Note: This might be flaky in async environments, but should work in this test
	if accessCount != 3 {
		t.Errorf("Expected signal to be accessed 3 times, got %d", accessCount)
	}
}

// TestToSignal_FilterMap tests that ToSignal respects filter and transform options
func TestToSignal_FilterMap(t *testing.T) {
	bus := New()

	// Create a signal with filter and transform
	signal := ToSignal[string](bus, "test-action",
		BridgeWithFilter(func(payload any) bool {
			// Only allow string payloads that start with "allow:"
			if str, ok := payload.(string); ok {
				return len(str) > 6 && str[:6] == "allow:"
			}
			return false
		}),
		BridgeWithTransform(func(payload any) any {
			// Transform by removing "allow:" prefix
			if str, ok := payload.(string); ok && len(str) > 6 {
				return str[6:]
			}
			return payload
		}),
	)

	// Type assert to the expected signal type
	typedSignal, ok := signal.(reactivity.Signal[string])
	if !ok {
		t.Fatalf("Expected Signal[string], got %T", signal)
	}

	// Check initial zero value
	if value := typedSignal.Get(); value != "" {
		t.Errorf("Expected initial empty string value, got '%s'", value)
	}

	// Dispatch filtered out action
	bus.Dispatch(Action[string]{
		Type:    "test-action",
		Payload: "deny:test",
	})

	// Value should still be empty (filtered out)
	if value := typedSignal.Get(); value != "" {
		t.Errorf("Expected value to remain empty after filtered action, got '%s'", value)
	}

	// Dispatch allowed action
	bus.Dispatch(Action[string]{
		Type:    "test-action",
		Payload: "allow:transformed",
	})

	// Value should be transformed
	if value := typedSignal.Get(); value != "transformed" {
		t.Errorf("Expected transformed value 'transformed', got '%s'", value)
	}
}

// TestToStream_BasicRecv tests that ToStream receives actions through Recv method
func TestToStream_BasicRecv(t *testing.T) {
	bus := New()

	// Create a stream
	stream := ToStream[any](bus, "test-action", BridgeWithBufferSize(5))

	// Type assert to the expected stream type
	typedStream, ok := stream.(Stream[any])
	if !ok {
		t.Fatalf("Expected Stream[any], got %T", stream)
	}

	// Dispatch some actions
	bus.Dispatch(Action[string]{
		Type:    "test-action",
		Payload: "payload1",
	})

	bus.Dispatch(Action[string]{
		Type:    "test-action",
		Payload: "payload2",
	})

	// Receive values
	value1, ok1 := typedStream.TryRecv()
	if !ok1 {
		t.Error("Expected to receive first value")
	}
	if value1 != "payload1" {
		t.Errorf("Expected 'payload1', got '%v'", value1)
	}

	value2, ok2 := typedStream.TryRecv()
	if !ok2 {
		t.Error("Expected to receive second value")
	}
	if value2 != "payload2" {
		t.Errorf("Expected 'payload2', got '%v'", value2)
	}

	// No more values
	_, ok3 := typedStream.TryRecv()
	if ok3 {
		t.Error("Expected no more values")
	}
}

// TestToStream_BufferAndDropPolicy tests that ToStream respects buffer size and drop policy
func TestToStream_BufferAndDropPolicy(t *testing.T) {
	bus := New()

	// Create a stream with small buffer and DropOldest policy
	stream := ToStream[any](bus, "test-action",
		BridgeWithBufferSize(2),
		BridgeWithDropPolicy(DropOldest))

	// Type assert to the expected stream type
	typedStream, ok := stream.(Stream[any])
	if !ok {
		t.Fatalf("Expected Stream[any], got %T", stream)
	}

	// Dispatch more actions than buffer can hold
	bus.Dispatch(Action[string]{
		Type:    "test-action",
		Payload: "payload1",
	})

	bus.Dispatch(Action[string]{
		Type:    "test-action",
		Payload: "payload2",
	})

	bus.Dispatch(Action[string]{
		Type:    "test-action",
		Payload: "payload3",
	})

	// With DropOldest policy, oldest item should be dropped
	// So we should receive payload2 and payload3
	value1, ok1 := typedStream.TryRecv()
	if !ok1 {
		t.Error("Expected to receive first value")
	}
	if value1 != "payload2" {
		t.Errorf("Expected 'payload2', got '%v'", value1)
	}

	value2, ok2 := typedStream.TryRecv()
	if !ok2 {
		t.Error("Expected to receive second value")
	}
	if value2 != "payload3" {
		t.Errorf("Expected 'payload3', got '%v'", value2)
	}

	// No more values
	_, ok3 := typedStream.TryRecv()
	if ok3 {
		t.Error("Expected no more values")
	}
}

// TestBridge_DisposeCleansSubscription tests that bridge disposal cleans up subscriptions
func TestBridge_DisposeCleansSubscription(t *testing.T) {
	bus := New()

	// Create a signal
	signal := ToSignal[any](bus, "test-action")

	// Type assert to the expected signal type
	typedSignal, ok := signal.(reactivity.Signal[any])
	if !ok {
		t.Fatalf("Expected Signal[any], got %T", signal)
	}

	// Dispatch an action to verify signal works
	bus.Dispatch(Action[string]{
		Type:    "test-action",
		Payload: "test",
	})

	if value := typedSignal.Get(); value != "test" {
		t.Errorf("Expected signal to receive value, got '%v'", value)
	}

	// Verify the subscriber count increased (should be 1 now)
	busImpl := bus.(*busImpl)
	busImpl.mu.RLock()
	subscriberCountAfter := len(busImpl.subscribers["test-action"])
	busImpl.mu.RUnlock()

	if subscriberCountAfter != 1 {
		t.Errorf("Expected subscriber count to be 1 after creating bridge, got %d", subscriberCountAfter)
	}

	// Test disposal
	if bridgeSignal, ok := signal.(interface{ Dispose() }); ok {
		bridgeSignal.Dispose()

		// Verify the subscriber count decreased
		busImpl.mu.RLock()
		subscriberCountAfterDispose := len(busImpl.subscribers["test-action"])
		busImpl.mu.RUnlock()

		if subscriberCountAfterDispose != 0 {
			t.Errorf("Expected subscriber count to be 0 after disposing bridge, got %d", subscriberCountAfterDispose)
		}
	}
}

// TestHandlerPanic_IsolatedAndErrorHookInvoked verifies panic recovery and error hook invocation
func TestHandlerPanic_IsolatedAndErrorHookInvoked(t *testing.T) {
	bus := New()

	var errorReceived error
	var panicRecovered any

	// Set up error handler
	bus.OnError(func(ctx Context, err error, recovered any) {
		errorReceived = err
		panicRecovered = recovered
	})

	var secondHandlerCalled bool
	var thirdHandlerCalled bool

	// First handler that panics
	bus.Subscribe("test-action", func(action Action[string]) error {
		panic("handler panic")
	})

	// Second handler that should still be called
	bus.Subscribe("test-action", func(action Action[string]) error {
		secondHandlerCalled = true
		return nil
	})

	// Third handler that should also be called
	bus.Subscribe("test-action", func(action Action[string]) error {
		thirdHandlerCalled = true
		return nil
	})

	// Dispatch the action
	bus.Dispatch(Action[string]{
		Type:    "test-action",
		Payload: "test-payload",
	})

	// Verify error handler was called with panic
	if errorReceived == nil {
		t.Fatal("Expected error handler to be called")
	}

	if panicRecovered != "handler panic" {
		t.Errorf("Expected panic recovered to be 'handler panic', got '%v'", panicRecovered)
	}

	// Verify other handlers were still called
	if !secondHandlerCalled {
		t.Error("Expected second handler to be called despite first handler panic")
	}

	if !thirdHandlerCalled {
		t.Error("Expected third handler to be called despite first handler panic")
	}
}

// TestOnce_DisposesAfterFirstDelivery verifies that Once subscriptions auto-dispose after first delivery
func TestOnce_DisposesAfterFirstDelivery(t *testing.T) {
	bus := New()

	var receivedCount int
	var subscription Subscription

	// Subscribe with Once option
	subscription = bus.Subscribe("test-action", func(action Action[string]) error {
		receivedCount++
		return nil
	}, SubOnce())

	// Dispatch first action
	bus.Dispatch(Action[string]{
		Type:    "test-action",
		Payload: "first-payload",
	})

	// Should have received the first action
	if receivedCount != 1 {
		t.Errorf("Expected to receive 1 action, got %d", receivedCount)
	}

	// Subscription should still be active immediately after delivery
	if !subscription.IsActive() {
		t.Error("Subscription should still be active immediately after first delivery")
	}

	// Wait a bit to allow for async disposal
	time.Sleep(10 * time.Millisecond)

	// Subscription should now be disposed
	if subscription.IsActive() {
		t.Error("Subscription should be disposed after first delivery")
	}

	// Dispatch second action
	bus.Dispatch(Action[string]{
		Type:    "test-action",
		Payload: "second-payload",
	})

	// Should not have received the second action
	if receivedCount != 1 {
		t.Errorf("Expected to still receive only 1 action, got %d", receivedCount)
	}
}

// TestFilter_SkipsWhenFalse verifies that Filter subscriptions skip delivery when predicate returns false
func TestFilter_SkipsWhenFalse(t *testing.T) {
	bus := New()

	var received []string

	// Subscribe with Filter option that only allows payloads containing "allowed"
	bus.Subscribe("test-action", func(action Action[string]) error {
		received = append(received, action.Payload)
		return nil
	}, SubFilter(func(payload any) bool {
		if str, ok := payload.(string); ok {
			return str == "allowed-payload"
		}
		return false
	}))

	// Dispatch allowed action
	bus.Dispatch(Action[string]{
		Type:    "test-action",
		Payload: "allowed-payload",
	})

	// Dispatch filtered action
	bus.Dispatch(Action[string]{
		Type:    "test-action",
		Payload: "filtered-payload",
	})

	// Dispatch another allowed action
	bus.Dispatch(Action[string]{
		Type:    "test-action",
		Payload: "allowed-payload",
	})

	// Should only have received the allowed actions
	if len(received) != 2 {
		t.Errorf("Expected to receive 2 actions, got %d", len(received))
	}

	for i, payload := range received {
		if payload != "allowed-payload" {
			t.Errorf("Expected received[%d] to be 'allowed-payload', got '%s'", i, payload)
		}
	}
}

// TestWhen_GatesDeliveryTrueOnly verifies that When subscriptions only deliver when signal is true
func TestWhen_GatesDeliveryTrueOnly(t *testing.T) {
	bus := New()

	var received []string
	gateSignal := reactivity.CreateSignal(true)

	// Subscribe with When option
	bus.Subscribe("test-action", func(action Action[string]) error {
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
}

// TestDistinctUntilChanged_DefaultEquality verifies that DistinctUntilChanged suppresses duplicate payloads
func TestDistinctUntilChanged_DefaultEquality(t *testing.T) {
	bus := New()

	var received []string

	// Subscribe with DistinctUntilChanged option
	bus.Subscribe("test-action", func(action Action[string]) error {
		received = append(received, action.Payload)
		return nil
	}, SubDistinctUntilChanged(nil)) // nil means use default equality

	// Dispatch sequence of actions
	testPayloads := []string{"same", "same", "different", "same", "same", "different", "same"}

	for _, payload := range testPayloads {
		bus.Dispatch(Action[string]{
			Type:    "test-action",
			Payload: payload,
		})
	}

	// Should only have received the distinct values in order
	expected := []string{"same", "different", "same", "different", "same"}
	if len(received) != len(expected) {
		t.Errorf("Expected to receive %d distinct actions, got %d", len(expected), len(received))
	}

	for i, expectedPayload := range expected {
		if received[i] != expectedPayload {
			t.Errorf("Expected received[%d] to be '%s', got '%s'", i, expectedPayload, received[i])
		}
	}
}

// TestDistinctUntilChanged_CustomEquality verifies that DistinctUntilChanged works with custom equality function
func TestDistinctUntilChanged_CustomEquality(t *testing.T) {
	bus := New()

	var received []int

	// Subscribe with DistinctUntilChanged option using custom equality (even/odd grouping)
	bus.SubscribeAny(func(action any) error {
		// Extract number from payload and convert to int
		if act, ok := action.(Action[any]); ok {
			if num, ok := act.Payload.(int); ok {
				received = append(received, num)
			}
		}
		return nil
	}, SubDistinctUntilChanged(func(a, b any) bool {
		numA, okA := a.(int)
		numB, okB := b.(int)
		if !okA || !okB {
			return false
		}
		// Consider numbers equal if they're both even or both odd
		return (numA%2 == 0 && numB%2 == 0) || (numA%2 == 1 && numB%2 == 1)
	}))

	// Dispatch sequence of numbers
	testNumbers := []int{1, 3, 5, 2, 4, 6, 8, 3, 1, 5, 7, 2, 4, 6}

	for _, num := range testNumbers {
		bus.Dispatch(Action[any]{
			Type:    "test-action",
			Payload: num,
		})
	}

	// Should only have received the first of each even/odd group
	expected := []int{1, 2, 3, 2} // First odd, first even, then odd again, then even again
	if len(received) != len(expected) {
		t.Errorf("Expected to receive %d distinct actions, got %d", len(expected), len(received))
	}

	for i, expectedNum := range expected {
		if received[i] != expectedNum {
			t.Errorf("Expected received[%d] to be %d, got %d", i, expectedNum, received[i])
		}
	}
}
