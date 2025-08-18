// store_test.go
// Tests for store and action functionality

package golid

import (
	"context"
	"testing"
	"time"
)

// ------------------------------------
// 🧪 Store Tests
// ------------------------------------

func TestCreateStore(t *testing.T) {
	store := CreateStore(42)

	if store.Get() != 42 {
		t.Errorf("Expected initial value 42, got %d", store.Get())
	}
}

func TestStoreSet(t *testing.T) {
	store := CreateStore(0)
	store.Set(100)

	if store.Get() != 100 {
		t.Errorf("Expected value 100, got %d", store.Get())
	}
}

func TestStoreUpdate(t *testing.T) {
	store := CreateStore(10)
	store.Update(func(val int) int {
		return val * 2
	})

	if store.Get() != 20 {
		t.Errorf("Expected value 20, got %d", store.Get())
	}
}

func TestStoreSubscription(t *testing.T) {
	store := CreateStore(0)

	var receivedValue int
	var callCount int

	unsubscribe := store.Subscribe(func(value int) {
		receivedValue = value
		callCount++
	})

	store.Set(42)

	// Give some time for subscription to process
	time.Sleep(10 * time.Millisecond)

	if receivedValue != 42 {
		t.Errorf("Expected received value 42, got %d", receivedValue)
	}

	if callCount != 1 {
		t.Errorf("Expected call count 1, got %d", callCount)
	}

	// Test unsubscribe
	unsubscribe()
	store.Set(100)

	time.Sleep(10 * time.Millisecond)

	if callCount != 1 {
		t.Errorf("Expected call count to remain 1 after unsubscribe, got %d", callCount)
	}
}

func TestStoreWithMiddleware(t *testing.T) {
	middleware := NewTestMiddleware[int]()

	store := CreateStore(0, StoreOptions[int]{
		Name:       "TestStore",
		Middleware: []StoreMiddleware[int]{middleware},
	})

	store.Set(42)

	before, after, _ := middleware.GetCallStatus()
	if !before {
		t.Error("Expected before middleware to be called")
	}
	if !after {
		t.Error("Expected after middleware to be called")
	}
}

func TestDerivedStore(t *testing.T) {
	baseStore := CreateStore(10)

	derivedStore := CreateDerivedStore(func() int {
		return baseStore.Get() * 2
	})

	if derivedStore.Get() != 20 {
		t.Errorf("Expected derived value 20, got %d", derivedStore.Get())
	}

	baseStore.Set(15)

	// Give some time for derived store to update
	time.Sleep(10 * time.Millisecond)

	if derivedStore.Get() != 30 {
		t.Errorf("Expected derived value 30, got %d", derivedStore.Get())
	}
}

func TestMapStore(t *testing.T) {
	baseStore := CreateStore(5)

	mappedStore := MapStore(baseStore, func(val int) string {
		return "Value: " + string(rune(val+'0'))
	})

	expected := "Value: 5"
	if mappedStore.Get() != expected {
		t.Errorf("Expected mapped value %s, got %s", expected, mappedStore.Get())
	}
}

func TestFilterStore(t *testing.T) {
	baseStore := CreateStore(5)

	filteredStore := FilterStore(baseStore, func(val int) bool {
		return val > 10
	})

	// Set a value that passes the filter
	baseStore.Set(15)
	time.Sleep(10 * time.Millisecond)

	if filteredStore.Get() != 15 {
		t.Errorf("Expected filtered value 15, got %d", filteredStore.Get())
	}

	// Set a value that doesn't pass the filter
	baseStore.Set(5)
	time.Sleep(10 * time.Millisecond)

	// Should still have the previous value
	if filteredStore.Get() != 15 {
		t.Errorf("Expected filtered value to remain 15, got %d", filteredStore.Get())
	}
}

func TestStoreDispose(t *testing.T) {
	store := CreateStore(0)

	var callCount int
	store.Subscribe(func(int) {
		callCount++
	})

	store.Set(1)
	time.Sleep(10 * time.Millisecond)

	if callCount != 1 {
		t.Errorf("Expected call count 1, got %d", callCount)
	}

	store.Dispose()
	store.Set(2)
	time.Sleep(10 * time.Millisecond)

	// Should not increase after dispose
	if callCount != 1 {
		t.Errorf("Expected call count to remain 1 after dispose, got %d", callCount)
	}
}

// ------------------------------------
// 🎯 Action Tests
// ------------------------------------

func TestCreateAction(t *testing.T) {
	action := CreateAction(func(x int) int {
		return x * 2
	})

	result := action.Execute(5)

	if result.Error != nil {
		t.Errorf("Unexpected error: %v", result.Error)
	}

	if result.Value != 10 {
		t.Errorf("Expected result 10, got %d", result.Value)
	}
}

func TestActionWithMiddleware(t *testing.T) {
	middleware := NewActionLoggingMiddleware[int, int]("TestAction")

	action := CreateAction(func(x int) int {
		return x + 1
	}, ActionOptions[int, int]{
		Name:       "TestAction",
		Middleware: []ActionMiddleware[int, int]{middleware},
	})

	result := action.Execute(5)

	if result.Error != nil {
		t.Errorf("Unexpected error: %v", result.Error)
	}

	if result.Value != 6 {
		t.Errorf("Expected result 6, got %d", result.Value)
	}
}

func TestAsyncAction(t *testing.T) {
	action := CreateAsyncAction(func(ctx context.Context, x int) (int, error) {
		time.Sleep(10 * time.Millisecond)
		return x * 3, nil
	})

	ctx := context.Background()
	result := action.Execute(ctx, 4)

	if result.Error != nil {
		t.Errorf("Unexpected error: %v", result.Error)
	}

	if result.Value != 12 {
		t.Errorf("Expected result 12, got %d", result.Value)
	}

	if result.Duration < 10*time.Millisecond {
		t.Errorf("Expected duration >= 10ms, got %v", result.Duration)
	}
}

func TestAsyncActionAsync(t *testing.T) {
	action := CreateAsyncAction(func(ctx context.Context, x int) (int, error) {
		time.Sleep(10 * time.Millisecond)
		return x * 3, nil
	})

	ctx := context.Background()
	resultChan := action.ExecuteAsync(ctx, 4)

	select {
	case result := <-resultChan:
		if result.Error != nil {
			t.Errorf("Unexpected error: %v", result.Error)
		}

		if result.Value != 12 {
			t.Errorf("Expected result 12, got %d", result.Value)
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("Async action timed out")
	}
}

func TestChainActions(t *testing.T) {
	first := CreateAction(func(x int) int {
		return x + 10
	})

	second := CreateAction(func(x int) int {
		return x * 2
	})

	chained := ChainActions(first, second)
	result := chained.Execute(5)

	if result.Error != nil {
		t.Errorf("Unexpected error: %v", result.Error)
	}

	// (5 + 10) * 2 = 30
	if result.Value != 30 {
		t.Errorf("Expected result 30, got %d", result.Value)
	}
}

// ------------------------------------
// 🚀 Dispatcher Tests
// ------------------------------------

func TestActionDispatcher(t *testing.T) {
	dispatcher := CreateDispatcher()

	action := CreateAction(func(x int) int {
		return x * 2
	})

	dispatcher.RegisterAction("double", action)

	result, err := dispatcher.Dispatch("double", 5)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if result != 10 {
		t.Errorf("Expected result 10, got %v", result)
	}
}

func TestDispatcherWithMiddleware(t *testing.T) {
	middleware := NewDispatcherLoggingMiddleware()
	dispatcher := CreateDispatcher(middleware)

	action := CreateAction(func(x int) int {
		return x + 1
	})

	dispatcher.RegisterAction("increment", action)

	result, err := dispatcher.Dispatch("increment", 5)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if result != 6 {
		t.Errorf("Expected result 6, got %v", result)
	}
}

func TestDispatcherNonExistentAction(t *testing.T) {
	dispatcher := CreateDispatcher()

	_, err := dispatcher.Dispatch("nonexistent", 5)
	if err == nil {
		t.Error("Expected error for non-existent action")
	}
}

// ------------------------------------
// 💾 Persistence Tests
// ------------------------------------

func TestMemoryAdapter(t *testing.T) {
	adapter := NewMemoryAdapter()

	// Test save and load
	data := []byte("test data")
	err := adapter.Save("test_key", data)
	if err != nil {
		t.Errorf("Unexpected error saving: %v", err)
	}

	loaded, err := adapter.Load("test_key")
	if err != nil {
		t.Errorf("Unexpected error loading: %v", err)
	}

	if string(loaded) != string(data) {
		t.Errorf("Expected loaded data %s, got %s", string(data), string(loaded))
	}

	// Test exists
	if !adapter.Exists("test_key") {
		t.Error("Expected key to exist")
	}

	if adapter.Exists("nonexistent_key") {
		t.Error("Expected nonexistent key to not exist")
	}

	// Test delete
	err = adapter.Delete("test_key")
	if err != nil {
		t.Errorf("Unexpected error deleting: %v", err)
	}

	if adapter.Exists("test_key") {
		t.Error("Expected key to not exist after delete")
	}
}

func TestJSONSerializer(t *testing.T) {
	serializer := NewJSONSerializer()

	original := map[string]interface{}{
		"name":  "test",
		"value": 42,
	}

	data, err := serializer.Serialize(original)
	if err != nil {
		t.Errorf("Unexpected error serializing: %v", err)
	}

	var deserialized map[string]interface{}
	err = serializer.Deserialize(data, &deserialized)
	if err != nil {
		t.Errorf("Unexpected error deserializing: %v", err)
	}

	if deserialized["name"] != "test" {
		t.Errorf("Expected name 'test', got %v", deserialized["name"])
	}

	// JSON numbers are float64
	if deserialized["value"] != float64(42) {
		t.Errorf("Expected value 42, got %v", deserialized["value"])
	}
}

func TestPersistentStore(t *testing.T) {
	store := CreateStore(42)
	adapter := NewMemoryAdapter()

	options := PersistenceOptions{
		Key:     "test_store",
		Adapter: adapter,
	}

	persistentStore, err := PersistStore(store, options)
	if err != nil {
		t.Errorf("Unexpected error creating persistent store: %v", err)
	}

	// Save current state
	err = persistentStore.Save()
	if err != nil {
		t.Errorf("Unexpected error saving: %v", err)
	}

	// Change store value
	store.Set(100)

	// Load from persistence (should restore to 42)
	err = persistentStore.Load()
	if err != nil {
		t.Errorf("Unexpected error loading: %v", err)
	}

	if store.Get() != 42 {
		t.Errorf("Expected restored value 42, got %d", store.Get())
	}
}

// ------------------------------------
// 🔄 Store Hydration Tests
// ------------------------------------

func TestStoreHydrator(t *testing.T) {
	hydrator := NewStoreHydrator()

	store1 := CreateStore(42)
	store2 := CreateStore("hello")

	hydrator.RegisterStore("store1", store1)
	hydrator.RegisterStore("store2", store2)

	// Dehydrate
	data, err := hydrator.Dehydrate()
	if err != nil {
		t.Errorf("Unexpected error dehydrating: %v", err)
	}

	if len(data.Stores) != 2 {
		t.Errorf("Expected 2 stores in hydration data, got %d", len(data.Stores))
	}

	// Change store values
	store1.Set(100)
	store2.Set("world")

	// Hydrate (should restore original values)
	err = hydrator.Hydrate(data)
	if err != nil {
		t.Errorf("Unexpected error hydrating: %v", err)
	}

	if store1.Get() != 42 {
		t.Errorf("Expected store1 value 42, got %d", store1.Get())
	}

	if store2.Get() != "hello" {
		t.Errorf("Expected store2 value 'hello', got %s", store2.Get())
	}
}

// ------------------------------------
// 🧪 Integration Tests
// ------------------------------------

func TestStoreActionIntegration(t *testing.T) {
	// Create a counter store
	store := CreateStore(0)

	// Create increment action
	incrementAction := CreateAction(func(amount int) int {
		current := store.Get()
		return current + amount
	})

	// Create dispatcher
	dispatcher := CreateDispatcher()
	dispatcher.RegisterAction("increment", incrementAction)

	// Subscribe to store changes
	var lastValue int
	store.Subscribe(func(value int) {
		lastValue = value
	})

	// Dispatch increment action
	result, err := dispatcher.Dispatch("increment", 5)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Update store with result
	store.Set(result.(int))

	time.Sleep(10 * time.Millisecond)

	if lastValue != 5 {
		t.Errorf("Expected store value 5, got %d", lastValue)
	}

	// Dispatch another increment
	result, err = dispatcher.Dispatch("increment", 3)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	store.Set(result.(int))
	time.Sleep(10 * time.Millisecond)

	if lastValue != 8 {
		t.Errorf("Expected store value 8, got %d", lastValue)
	}
}

// ------------------------------------
// 🧹 Test Utilities
// ------------------------------------

func TestCreateTestStore(t *testing.T) {
	store, cleanup := CreateTestStore(42)
	defer cleanup()

	if store.Get() != 42 {
		t.Errorf("Expected test store value 42, got %d", store.Get())
	}
}

func TestCreateTestAction(t *testing.T) {
	action, cleanup := CreateTestAction(func(x int) int {
		return x * 2
	})
	defer cleanup()

	result := action.Execute(5)
	if result.Value != 10 {
		t.Errorf("Expected test action result 10, got %d", result.Value)
	}
}

func TestCreateTestAsyncAction(t *testing.T) {
	action, cleanup := CreateTestAsyncAction(func(ctx context.Context, x int) (int, error) {
		return x * 3, nil
	})
	defer cleanup()

	result := action.Execute(context.Background(), 4)
	if result.Value != 12 {
		t.Errorf("Expected test async action result 12, got %d", result.Value)
	}
}
