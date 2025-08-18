//go:build js && wasm

// event_system_test.go
// Comprehensive tests for the new event system

package golid

import (
	"syscall/js"
	"testing"
	"time"
)

// TestEventManagerCreation tests event manager creation and initialization
func TestEventManagerCreation(t *testing.T) {
	manager := NewEventManager(nil)

	if manager == nil {
		t.Fatal("Event manager should not be nil")
	}

	if manager.subscriptions == nil {
		t.Error("Subscriptions map should be initialized")
	}

	if manager.delegator == nil {
		t.Error("Delegator should be initialized")
	}

	if manager.batcher == nil {
		t.Error("Batcher should be initialized")
	}

	if manager.metrics == nil {
		t.Error("Metrics should be initialized")
	}

	if manager.customBus == nil {
		t.Error("Custom bus should be initialized")
	}
}

// TestEventSubscription tests basic event subscription functionality
func TestEventSubscription(t *testing.T) {
	manager := NewEventManager(nil)

	// Create a mock element
	mockElement := js.Global().Get("document").Call("createElement", "div")

	var eventFired bool
	cleanup := manager.Subscribe(mockElement, "click", func(e js.Value) {
		eventFired = true
	})

	if cleanup == nil {
		t.Error("Cleanup function should not be nil")
	}

	// Check subscription count
	if manager.GetSubscriptionCount() != 1 {
		t.Errorf("Expected 1 subscription, got %d", manager.GetSubscriptionCount())
	}

	// Use eventFired to avoid unused variable error
	_ = eventFired

	// Cleanup
	cleanup()

	// Check subscription count after cleanup
	if manager.GetSubscriptionCount() != 0 {
		t.Errorf("Expected 0 subscriptions after cleanup, got %d", manager.GetSubscriptionCount())
	}
}

// TestEventDelegation tests event delegation functionality
func TestEventDelegation(t *testing.T) {
	delegator := NewEventDelegator()

	if delegator == nil {
		t.Fatal("Event delegator should not be nil")
	}

	// Test delegation capability
	if !delegator.CanDelegate("click") {
		t.Error("Click events should be delegatable")
	}

	if !delegator.CanDelegate("input") {
		t.Error("Input events should be delegatable")
	}

	if delegator.CanDelegate("load") {
		t.Error("Load events should not be delegatable")
	}
}

// TestEventBatcher tests event batching functionality
func TestEventBatcher(t *testing.T) {
	batcher := NewEventBatcher()

	if batcher == nil {
		t.Fatal("Event batcher should not be nil")
	}

	executed := false
	batcher.Schedule(func() {
		executed = true
	}, Normal)

	// Give some time for batch processing
	time.Sleep(50 * time.Millisecond)

	if !executed {
		t.Error("Batched event should have been executed")
	}

	batcher.Dispose()
}

// TestCustomEventBus tests custom event bus functionality
func TestCustomEventBus(t *testing.T) {
	bus := NewCustomEventBus()

	if bus == nil {
		t.Fatal("Custom event bus should not be nil")
	}

	eventReceived := false
	var receivedData interface{}

	// Subscribe to custom event
	cleanup := bus.On("test-event", func(data interface{}) {
		eventReceived = true
		receivedData = data
	})

	// Emit event
	testData := "test data"
	bus.Emit("test-event", testData)

	if !eventReceived {
		t.Error("Custom event should have been received")
	}

	if receivedData != testData {
		t.Errorf("Expected data '%s', got '%v'", testData, receivedData)
	}

	cleanup()
	bus.Dispose()
}

// TestEventMetrics tests event metrics tracking
func TestEventMetrics(t *testing.T) {
	metrics := NewEventMetrics()

	if metrics == nil {
		t.Fatal("Event metrics should not be nil")
	}

	// Test metric increments
	metrics.incrementTotal()
	metrics.incrementDelegated()
	metrics.incrementDirect()
	metrics.incrementCleanup()

	stats := metrics.GetStats()

	if stats["totalSubscriptions"].(uint64) != 1 {
		t.Error("Total subscriptions should be 1")
	}

	if stats["delegatedEvents"].(uint64) != 1 {
		t.Error("Delegated events should be 1")
	}

	if stats["directEvents"].(uint64) != 1 {
		t.Error("Direct events should be 1")
	}

	if stats["cleanupOperations"].(uint64) != 1 {
		t.Error("Cleanup operations should be 1")
	}
}

// TestEventOptions tests event options functionality
func TestEventOptions(t *testing.T) {
	manager := NewEventManager(nil)
	mockElement := js.Global().Get("document").Call("createElement", "div")

	// Test with custom options
	options := EventOptions{
		Delegate: false,
		Priority: UserBlocking,
		Debounce: 100 * time.Millisecond,
	}

	cleanup := manager.Subscribe(mockElement, "click", func(e js.Value) {
		// Test handler
	}, options)

	if cleanup == nil {
		t.Error("Cleanup function should not be nil")
	}

	cleanup()
}

// TestEventCleanup tests automatic cleanup functionality
func TestEventCleanup(t *testing.T) {
	manager := NewEventManager(nil)

	// Create multiple subscriptions
	mockElement := js.Global().Get("document").Call("createElement", "div")

	cleanup1 := manager.Subscribe(mockElement, "click", func(e js.Value) {})
	cleanup2 := manager.Subscribe(mockElement, "input", func(e js.Value) {})
	cleanup3 := manager.Subscribe(mockElement, "change", func(e js.Value) {})

	if manager.GetSubscriptionCount() != 3 {
		t.Errorf("Expected 3 subscriptions, got %d", manager.GetSubscriptionCount())
	}

	// Cleanup individual subscriptions
	cleanup1()
	if manager.GetSubscriptionCount() != 2 {
		t.Errorf("Expected 2 subscriptions after cleanup1, got %d", manager.GetSubscriptionCount())
	}

	cleanup2()
	if manager.GetSubscriptionCount() != 1 {
		t.Errorf("Expected 1 subscription after cleanup2, got %d", manager.GetSubscriptionCount())
	}

	cleanup3()
	if manager.GetSubscriptionCount() != 0 {
		t.Errorf("Expected 0 subscriptions after cleanup3, got %d", manager.GetSubscriptionCount())
	}
}

// TestEventManagerDispose tests complete event manager disposal
func TestEventManagerDispose(t *testing.T) {
	manager := NewEventManager(nil)
	mockElement := js.Global().Get("document").Call("createElement", "div")

	// Create subscriptions
	manager.Subscribe(mockElement, "click", func(e js.Value) {})
	manager.Subscribe(mockElement, "input", func(e js.Value) {})

	if manager.GetSubscriptionCount() != 2 {
		t.Errorf("Expected 2 subscriptions, got %d", manager.GetSubscriptionCount())
	}

	// Dispose manager
	manager.Dispose()

	if manager.GetSubscriptionCount() != 0 {
		t.Errorf("Expected 0 subscriptions after dispose, got %d", manager.GetSubscriptionCount())
	}
}

// TestReactiveEventSubscription tests reactive event subscription
func TestReactiveEventSubscription(t *testing.T) {
	manager := NewEventManager(nil)
	mockElement := js.Global().Get("document").Call("createElement", "div")

	var eventFired bool
	cleanup := manager.SubscribeReactive(mockElement, "click", func(e js.Value) {
		eventFired = true
	})

	if cleanup == nil {
		t.Error("Cleanup function should not be nil")
	}

	// Use eventFired to avoid unused variable error
	_ = eventFired

	cleanup()
}

// TestHandlerPool tests handler pool functionality
func TestHandlerPool(t *testing.T) {
	pool := NewHandlerPool(5)

	if pool == nil {
		t.Fatal("Handler pool should not be nil")
	}

	// Get handler from empty pool
	handler := pool.Get()
	if handler != nil {
		t.Error("Should get nil from empty pool")
	}

	// Create and put handler
	testHandler := &DelegatedEventHandler{
		id:       1,
		selector: "test",
	}

	pool.Put(testHandler)

	// Get handler from pool
	retrieved := pool.Get()
	if retrieved == nil {
		t.Error("Should get handler from pool")
	}

	if retrieved.selector != "" {
		t.Error("Handler should be reset when retrieved from pool")
	}

	stats := pool.GetStats()
	if stats["reused"] != 1 {
		t.Errorf("Expected reused count to be 1, got %d", stats["reused"])
	}
}

// TestEventRouter tests event routing functionality
func TestEventRouter(t *testing.T) {
	router := NewEventRouter()

	if router == nil {
		t.Fatal("Event router should not be nil")
	}

	// Create mock element
	mockElement := js.Global().Get("document").Call("createElement", "div")
	mockElement.Set("tagName", "DIV")
	mockElement.Set("className", "test-class")

	// Get event path
	path := router.GetEventPath(mockElement)

	if path == nil {
		t.Error("Event path should not be nil")
	}

	if len(path.elements) == 0 {
		t.Error("Event path should contain elements")
	}

	router.Dispose()
}

// TestSubscriptionPool tests subscription pool functionality
func TestSubscriptionPool(t *testing.T) {
	pool := NewSubscriptionPool(5)

	if pool == nil {
		t.Fatal("Subscription pool should not be nil")
	}

	// Get from empty pool
	sub := pool.Get()
	if sub == nil {
		t.Error("Should create new subscription when pool is empty")
	}

	// Put back to pool
	pool.Put(sub)

	// Get from pool
	retrieved := pool.Get()
	if retrieved == nil {
		t.Error("Should get subscription from pool")
	}

	stats := pool.GetStats()
	if stats["reused"].(uint64) != 1 {
		t.Error("Reused count should be 1")
	}
}

// TestGlobalEventManager tests global event manager functionality
func TestGlobalEventManager(t *testing.T) {
	manager1 := GetEventManager()
	manager2 := GetEventManager()

	if manager1 != manager2 {
		t.Error("Global event manager should be singleton")
	}

	if manager1 == nil {
		t.Error("Global event manager should not be nil")
	}
}

// TestConvenienceFunctions tests convenience functions
func TestConvenienceFunctions(t *testing.T) {
	mockElement := js.Global().Get("document").Call("createElement", "div")

	// Test Subscribe convenience function
	cleanup := Subscribe(mockElement, "click", func(e js.Value) {})
	if cleanup == nil {
		t.Error("Subscribe convenience function should return cleanup")
	}
	cleanup()

	// Test SubscribeReactive convenience function
	cleanup2 := SubscribeReactive(mockElement, "input", func(e js.Value) {})
	if cleanup2 == nil {
		t.Error("SubscribeReactive convenience function should return cleanup")
	}
	cleanup2()
}

// BenchmarkEventSubscription benchmarks event subscription performance
func BenchmarkEventSubscription(b *testing.B) {
	manager := NewEventManager(nil)
	mockElement := js.Global().Get("document").Call("createElement", "div")

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		cleanup := manager.Subscribe(mockElement, "click", func(e js.Value) {})
		cleanup()
	}
}

// BenchmarkEventDelegation benchmarks event delegation performance
func BenchmarkEventDelegation(b *testing.B) {
	delegator := NewEventDelegator()
	mockElement := js.Global().Get("document").Call("createElement", "div")

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		cleanup := delegator.Subscribe(mockElement, "click", func(e js.Value) {}, EventOptions{
			Delegate: true,
		})
		cleanup()
	}
}

// BenchmarkHandlerPool benchmarks handler pool performance
func BenchmarkHandlerPool(b *testing.B) {
	pool := NewHandlerPool(100)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		handler := pool.Get()
		if handler == nil {
			handler = &DelegatedEventHandler{}
		}
		pool.Put(handler)
	}
}
