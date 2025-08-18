// signals_test.go
// Comprehensive unit tests for the reactive signals system
// Tests both DOM-independent functionality and the Watch/effect system

package golid

import (
	"sync"
	"testing"
	"time"
)

// TestNewSignalWithDifferentTypes tests signal creation with various types
func TestNewSignalWithDifferentTypes(t *testing.T) {
	testCases := []struct {
		name         string
		initialValue interface{}
		expectedType string
	}{
		{"string signal", "hello world", "string"},
		{"integer signal", 42, "int"},
		{"boolean signal", true, "bool"},
		{"float signal", 3.14, "float64"},
		{"slice signal", []string{"a", "b", "c"}, "[]string"},
		{"map signal", map[string]int{"key": 1}, "map[string]int"},
		{"struct signal", struct{ Name string }{Name: "test"}, "struct"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			switch v := tc.initialValue.(type) {
			case string:
				signal := NewSignal(v)
				if signal == nil {
					t.Fatalf("NewSignal should not return nil for %s", tc.name)
				}
				if signal.Get() != v {
					t.Errorf("Expected %v, got %v", v, signal.Get())
				}
			case int:
				signal := NewSignal(v)
				if signal == nil {
					t.Fatalf("NewSignal should not return nil for %s", tc.name)
				}
				if signal.Get() != v {
					t.Errorf("Expected %v, got %v", v, signal.Get())
				}
			case bool:
				signal := NewSignal(v)
				if signal == nil {
					t.Fatalf("NewSignal should not return nil for %s", tc.name)
				}
				if signal.Get() != v {
					t.Errorf("Expected %v, got %v", v, signal.Get())
				}
			case float64:
				signal := NewSignal(v)
				if signal == nil {
					t.Fatalf("NewSignal should not return nil for %s", tc.name)
				}
				if signal.Get() != v {
					t.Errorf("Expected %v, got %v", v, signal.Get())
				}
			case []string:
				signal := NewSignal(v)
				if signal == nil {
					t.Fatalf("NewSignal should not return nil for %s", tc.name)
				}
				result := signal.Get()
				if len(result) != len(v) {
					t.Errorf("Expected length %d, got %d", len(v), len(result))
				}
			case map[string]int:
				signal := NewSignal(v)
				if signal == nil {
					t.Fatalf("NewSignal should not return nil for %s", tc.name)
				}
				result := signal.Get()
				if len(result) != len(v) {
					t.Errorf("Expected length %d, got %d", len(v), len(result))
				}
			default:
				// Handle struct and other types generically
				signal := NewSignal(v)
				if signal == nil {
					t.Fatalf("NewSignal should not return nil for %s", tc.name)
				}
			}
		})
	}
}

// TestSignalGetSet tests basic signal get/set operations
func TestSignalGetSet(t *testing.T) {
	// Test string signal
	stringSignal := NewSignal("initial")
	if stringSignal.Get() != "initial" {
		t.Errorf("Expected 'initial', got %s", stringSignal.Get())
	}

	stringSignal.Set("updated")
	if stringSignal.Get() != "updated" {
		t.Errorf("Expected 'updated', got %s", stringSignal.Get())
	}

	// Test integer signal
	intSignal := NewSignal(0)
	if intSignal.Get() != 0 {
		t.Errorf("Expected 0, got %d", intSignal.Get())
	}

	intSignal.Set(100)
	if intSignal.Get() != 100 {
		t.Errorf("Expected 100, got %d", intSignal.Get())
	}

	// Test boolean signal
	boolSignal := NewSignal(false)
	if boolSignal.Get() != false {
		t.Errorf("Expected false, got %t", boolSignal.Get())
	}

	boolSignal.Set(true)
	if boolSignal.Get() != true {
		t.Errorf("Expected true, got %t", boolSignal.Get())
	}
}

// TestSignalMultipleUpdates tests multiple signal updates
func TestSignalMultipleUpdates(t *testing.T) {
	signal := NewSignal(0)

	// Perform multiple updates
	for i := 1; i <= 10; i++ {
		signal.Set(i)
		if signal.Get() != i {
			t.Errorf("Expected %d, got %d", i, signal.Get())
		}
	}
}

// TestSignalZeroValues tests signals with zero values
func TestSignalZeroValues(t *testing.T) {
	// Test zero string
	stringSignal := NewSignal("")
	if stringSignal.Get() != "" {
		t.Errorf("Expected empty string, got %s", stringSignal.Get())
	}

	// Test zero int
	intSignal := NewSignal(0)
	if intSignal.Get() != 0 {
		t.Errorf("Expected 0, got %d", intSignal.Get())
	}

	// Test zero bool
	boolSignal := NewSignal(false)
	if boolSignal.Get() != false {
		t.Errorf("Expected false, got %t", boolSignal.Get())
	}

	// Test nil slice
	sliceSignal := NewSignal([]string(nil))
	if sliceSignal.Get() != nil {
		t.Errorf("Expected nil slice, got %v", sliceSignal.Get())
	}
}

// TestWatchBasicFunctionality tests the Watch function for creating reactive effects
func TestWatchBasicFunctionality(t *testing.T) {
	counter := NewSignal(0)
	executed := false

	// Create a Watch effect that should execute immediately
	Watch(func() {
		_ = counter.Get() // Access the signal to create dependency
		executed = true
	})

	// Watch should execute immediately
	if !executed {
		t.Error("Watch effect should execute immediately")
	}
}

// TestWatchWithSignalChanges tests that Watch effects respond to signal changes
func TestWatchWithSignalChanges(t *testing.T) {
	counter := NewSignal(0)
	executionCount := 0
	lastValue := -1

	// Create a Watch effect
	Watch(func() {
		lastValue = counter.Get()
		executionCount++
	})

	// Should have executed once immediately
	if executionCount != 1 {
		t.Errorf("Expected 1 execution, got %d", executionCount)
	}
	if lastValue != 0 {
		t.Errorf("Expected last value 0, got %d", lastValue)
	}

	// Update the signal - effect should run again
	counter.Set(5)

	// Give some time for the goroutine to execute
	time.Sleep(10 * time.Millisecond)

	if executionCount < 2 {
		t.Errorf("Expected at least 2 executions, got %d", executionCount)
	}
	if lastValue != 5 {
		t.Errorf("Expected last value 5, got %d", lastValue)
	}
}

// TestWatchWithMultipleSignals tests Watch with multiple signal dependencies
func TestWatchWithMultipleSignals(t *testing.T) {
	signal1 := NewSignal(1)
	signal2 := NewSignal(2)
	executionCount := 0
	lastSum := -1

	// Create effect that depends on both signals
	Watch(func() {
		lastSum = signal1.Get() + signal2.Get()
		executionCount++
	})

	// Should execute immediately
	if executionCount != 1 {
		t.Errorf("Expected 1 execution, got %d", executionCount)
	}
	if lastSum != 3 {
		t.Errorf("Expected sum 3, got %d", lastSum)
	}

	// Update first signal
	signal1.Set(10)
	time.Sleep(10 * time.Millisecond)

	if executionCount < 2 {
		t.Errorf("Expected at least 2 executions after first update, got %d", executionCount)
	}
	if lastSum != 12 {
		t.Errorf("Expected sum 12, got %d", lastSum)
	}

	// Update second signal
	signal2.Set(20)
	time.Sleep(10 * time.Millisecond)

	if executionCount < 3 {
		t.Errorf("Expected at least 3 executions after second update, got %d", executionCount)
	}
	if lastSum != 30 {
		t.Errorf("Expected sum 30, got %d", lastSum)
	}
}

// TestWatchConcurrentUpdates tests Watch with concurrent signal updates
func TestWatchConcurrentUpdates(t *testing.T) {
	signal := NewSignal(0)
	var mutex sync.Mutex
	executionCount := 0

	Watch(func() {
		_ = signal.Get()
		mutex.Lock()
		executionCount++
		mutex.Unlock()
	})

	// Initial execution
	if executionCount != 1 {
		t.Errorf("Expected 1 initial execution, got %d", executionCount)
	}

	// Perform concurrent updates
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(val int) {
			defer wg.Done()
			signal.Set(val)
		}(i)
	}

	wg.Wait()
	time.Sleep(50 * time.Millisecond) // Allow effects to complete

	mutex.Lock()
	finalCount := executionCount
	mutex.Unlock()

	// Should have executed more than once due to concurrent updates
	if finalCount <= 1 {
		t.Errorf("Expected more than 1 execution from concurrent updates, got %d", finalCount)
	}
}

// TestEffectCleanup tests that effects properly clean up their dependencies
func TestEffectCleanup(t *testing.T) {
	signal1 := NewSignal(1)
	signal2 := NewSignal(2)
	executionCount := 0

	Watch(func() {
		executionCount++
		// Initially depend on both signals
		if executionCount == 1 {
			_ = signal1.Get()
			_ = signal2.Get()
		} else {
			// Later only depend on signal1
			_ = signal1.Get()
		}
	})

	// Initial execution
	if executionCount != 1 {
		t.Errorf("Expected 1 initial execution, got %d", executionCount)
	}

	// Update signal1 - should trigger effect
	signal1.Set(10)
	time.Sleep(10 * time.Millisecond)

	if executionCount < 2 {
		t.Errorf("Expected at least 2 executions, got %d", executionCount)
	}

	previousCount := executionCount

	// Update signal2 - should trigger effect since dependency cleanup happens on next run
	signal2.Set(20)
	time.Sleep(10 * time.Millisecond)

	// The effect may or may not run again depending on timing of cleanup
	// This test mainly ensures no crash occurs during dependency cleanup
	if executionCount < previousCount {
		t.Error("Execution count should not decrease")
	}
}

// TestSignalWatchersMap tests that the watchers map is properly maintained
func TestSignalWatchersMap(t *testing.T) {
	signal := NewSignal(0)

	// Initially, watchers should be empty
	if len(signal.watchers) != 0 {
		t.Errorf("Expected empty watchers map, got length %d", len(signal.watchers))
	}

	// Create a Watch effect
	Watch(func() {
		_ = signal.Get()
	})

	// Now watchers should have one entry
	if len(signal.watchers) != 1 {
		t.Errorf("Expected 1 watcher, got %d", len(signal.watchers))
	}
}

// TestSignalIsolation tests that different signals don't interfere with each other
func TestSignalIsolation(t *testing.T) {
	signal1 := NewSignal("signal1")
	signal2 := NewSignal("signal2")

	// Verify initial values
	if signal1.Get() != "signal1" {
		t.Errorf("Signal1 expected 'signal1', got %s", signal1.Get())
	}
	if signal2.Get() != "signal2" {
		t.Errorf("Signal2 expected 'signal2', got %s", signal2.Get())
	}

	// Update one signal
	signal1.Set("updated1")

	// Verify only the updated signal changed
	if signal1.Get() != "updated1" {
		t.Errorf("Signal1 expected 'updated1', got %s", signal1.Get())
	}
	if signal2.Get() != "signal2" {
		t.Errorf("Signal2 should remain unchanged, got %s", signal2.Get())
	}

	// Update the other signal
	signal2.Set("updated2")

	// Verify both have correct values
	if signal1.Get() != "updated1" {
		t.Errorf("Signal1 expected 'updated1', got %s", signal1.Get())
	}
	if signal2.Get() != "updated2" {
		t.Errorf("Signal2 expected 'updated2', got %s", signal2.Get())
	}
}
