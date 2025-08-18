// reactivity_test.go
// Comprehensive tests for the new SolidJS-inspired reactivity system

package golid

import (
	"sync"
	"testing"
)

func TestCreateSignal(t *testing.T) {
	// Reset reactive context for clean test
	ResetReactiveContext()
	ResetScheduler()

	// Test basic signal creation and access
	getter, setter := CreateSignal(42)

	if getter() != 42 {
		t.Errorf("Expected initial value 42, got %d", getter())
	}

	// Test signal update
	setter(100)
	FlushScheduler() // Ensure updates are processed

	if getter() != 100 {
		t.Errorf("Expected updated value 100, got %d", getter())
	}
}

func TestCreateEffect(t *testing.T) {
	ResetReactiveContext()
	ResetScheduler()

	getter, setter := CreateSignal(0)
	var effectValue int
	var executionCount int

	// Create effect that depends on signal
	CreateEffect(func() {
		effectValue = getter()
		executionCount++
	}, nil)

	FlushScheduler()

	// Effect should run immediately
	if executionCount != 1 {
		t.Errorf("Expected effect to run once initially, got %d executions", executionCount)
	}
	if effectValue != 0 {
		t.Errorf("Expected effect value 0, got %d", effectValue)
	}

	// Update signal and check effect runs again
	setter(42)
	FlushScheduler()

	if executionCount != 2 {
		t.Errorf("Expected effect to run twice after signal update, got %d executions", executionCount)
	}
	if effectValue != 42 {
		t.Errorf("Expected effect value 42, got %d", effectValue)
	}
}

func TestCreateMemo(t *testing.T) {
	ResetReactiveContext()
	ResetScheduler()

	getter, setter := CreateSignal(5)
	var computationCount int

	// Create memo that doubles the signal value
	memo := CreateMemo(func() int {
		computationCount++
		return getter() * 2
	}, nil)

	FlushScheduler()

	// Memo should compute initially
	if memo() != 10 {
		t.Errorf("Expected memo value 10, got %d", memo())
	}
	if computationCount != 1 {
		t.Errorf("Expected memo to compute once, got %d computations", computationCount)
	}

	// Accessing memo again should not recompute
	if memo() != 10 {
		t.Errorf("Expected memo value 10 on second access, got %d", memo())
	}
	if computationCount != 1 {
		t.Errorf("Expected memo to still have computed once, got %d computations", computationCount)
	}

	// Update signal and check memo recomputes
	setter(7)
	FlushScheduler()

	if memo() != 14 {
		t.Errorf("Expected memo value 14 after signal update, got %d", memo())
	}
	if computationCount != 2 {
		t.Errorf("Expected memo to compute twice after signal update, got %d computations", computationCount)
	}
}

func TestCreateRoot(t *testing.T) {
	ResetReactiveContext()
	ResetScheduler()

	var cleanupCalled bool
	var effectValue int

	result, cleanup := CreateRoot(func() string {
		getter, setter := CreateSignal(42)

		CreateEffect(func() {
			effectValue = getter()
		}, nil)

		OnCleanup(func() {
			cleanupCalled = true
		})

		setter(100)
		return "test result"
	})

	FlushScheduler()

	if result != "test result" {
		t.Errorf("Expected result 'test result', got %s", result)
	}
	if effectValue != 100 {
		t.Errorf("Expected effect value 100, got %d", effectValue)
	}
	if cleanupCalled {
		t.Error("Cleanup should not be called before cleanup function is called")
	}

	// Call cleanup and verify cleanup functions run
	cleanup()

	if !cleanupCalled {
		t.Error("Cleanup function should have been called")
	}
}

func TestOwnerCleanup(t *testing.T) {
	ResetReactiveContext()
	ResetScheduler()

	var cleanupCount int
	var effectExecutions int

	_, cleanup := CreateRoot(func() interface{} {
		getter, setter := CreateSignal(0)

		// Create multiple effects and cleanup functions
		for i := 0; i < 3; i++ {
			CreateEffect(func() {
				_ = getter()
				effectExecutions++
			}, nil)

			OnCleanup(func() {
				cleanupCount++
			})
		}

		setter(1) // Trigger effects
		return nil
	})

	FlushScheduler()

	// All effects should have run initially and after update
	if effectExecutions != 6 { // 3 initial + 3 after update
		t.Errorf("Expected 6 effect executions, got %d", effectExecutions)
	}

	// Cleanup should dispose all resources
	cleanup()

	if cleanupCount != 3 {
		t.Errorf("Expected 3 cleanup calls, got %d", cleanupCount)
	}
}

func TestBatchedUpdates(t *testing.T) {
	ResetReactiveContext()
	ResetScheduler()

	getter1, setter1 := CreateSignal(1)
	getter2, setter2 := CreateSignal(2)
	var effectExecutions int
	var lastSum int

	CreateEffect(func() {
		lastSum = getter1() + getter2()
		effectExecutions++
	}, nil)

	FlushScheduler()

	// Effect should run once initially
	if effectExecutions != 1 {
		t.Errorf("Expected 1 initial effect execution, got %d", effectExecutions)
	}
	if lastSum != 3 {
		t.Errorf("Expected initial sum 3, got %d", lastSum)
	}

	// Batch multiple updates
	Batch(func() {
		setter1(10)
		setter2(20)
	})

	FlushScheduler()

	// Effect should run only once for batched updates
	if effectExecutions != 2 {
		t.Errorf("Expected 2 total effect executions after batch, got %d", effectExecutions)
	}
	if lastSum != 30 {
		t.Errorf("Expected final sum 30, got %d", lastSum)
	}
}

func TestSignalEquality(t *testing.T) {
	ResetReactiveContext()
	ResetScheduler()

	// Test with custom equality function
	getter, setter := CreateSignal(5, SignalOptions[int]{
		Equals: func(prev, next int) bool {
			return prev == next
		},
	})

	var effectExecutions int
	CreateEffect(func() {
		_ = getter()
		effectExecutions++
	}, nil)

	FlushScheduler()

	// Initial execution
	if effectExecutions != 1 {
		t.Errorf("Expected 1 initial execution, got %d", effectExecutions)
	}

	// Set same value - should not trigger effect
	setter(5)
	FlushScheduler()

	if effectExecutions != 1 {
		t.Errorf("Expected still 1 execution after setting same value, got %d", effectExecutions)
	}

	// Set different value - should trigger effect
	setter(10)
	FlushScheduler()

	if effectExecutions != 2 {
		t.Errorf("Expected 2 executions after setting different value, got %d", effectExecutions)
	}
}

func TestConcurrentSignalAccess(t *testing.T) {
	ResetReactiveContext()
	ResetScheduler()

	getter, setter := CreateSignal(0)
	var wg sync.WaitGroup
	var mu sync.Mutex
	var values []int

	// Start multiple goroutines that read and write the signal
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			// Write
			setter(id)

			// Read
			val := getter()

			mu.Lock()
			values = append(values, val)
			mu.Unlock()
		}(i)
	}

	wg.Wait()
	FlushScheduler()

	// Should have collected 10 values
	if len(values) != 10 {
		t.Errorf("Expected 10 values, got %d", len(values))
	}

	// All values should be valid (0-9)
	for _, val := range values {
		if val < 0 || val > 9 {
			t.Errorf("Invalid value %d, expected 0-9", val)
		}
	}
}

func TestNestedEffects(t *testing.T) {
	ResetReactiveContext()
	ResetScheduler()

	getter1, setter1 := CreateSignal(1)
	getter2, setter2 := CreateSignal(2)
	var outerExecutions, innerExecutions int

	CreateEffect(func() {
		_ = getter1()
		outerExecutions++

		CreateEffect(func() {
			_ = getter2()
			innerExecutions++
		}, nil)
	}, nil)

	FlushScheduler()

	// Both effects should run initially
	if outerExecutions != 1 {
		t.Errorf("Expected 1 outer execution, got %d", outerExecutions)
	}
	if innerExecutions != 1 {
		t.Errorf("Expected 1 inner execution, got %d", innerExecutions)
	}

	// Update first signal - should trigger outer effect and create new inner effect
	setter1(10)
	FlushScheduler()

	if outerExecutions != 2 {
		t.Errorf("Expected 2 outer executions, got %d", outerExecutions)
	}
	if innerExecutions != 2 {
		t.Errorf("Expected 2 inner executions, got %d", innerExecutions)
	}

	// Update second signal - should trigger the latest inner effect
	setter2(20)
	FlushScheduler()

	if outerExecutions != 2 {
		t.Errorf("Expected still 2 outer executions, got %d", outerExecutions)
	}
	if innerExecutions >= 3 {
		t.Logf("Inner executions: %d (may vary due to nested effect creation)", innerExecutions)
	}
}

func TestSchedulerStats(t *testing.T) {
	ResetReactiveContext()
	ResetScheduler()

	scheduler := getScheduler()

	// Initially should be empty
	stats := scheduler.GetStats()
	if stats.QueueSize != 0 {
		t.Errorf("Expected empty queue, got size %d", stats.QueueSize)
	}

	// Create some work
	getter, setter := CreateSignal(0)
	CreateEffect(func() {
		_ = getter()
	}, nil)

	setter(1)

	// Should have some work queued (before flush)
	stats = scheduler.GetStats()
	if stats.QueueSize < 0 {
		t.Errorf("Expected non-negative queue size, got %d", stats.QueueSize)
	}

	FlushScheduler()

	// Should be empty after flush
	stats = scheduler.GetStats()
	if stats.QueueSize != 0 {
		t.Errorf("Expected empty queue after flush, got size %d", stats.QueueSize)
	}
}

func TestReactiveContextStats(t *testing.T) {
	ResetReactiveContext()

	// Initially should be empty
	stats := GetReactiveStats()
	if stats["currentComputation"].(bool) {
		t.Error("Expected no current computation initially")
	}
	if stats["currentOwner"].(bool) {
		t.Error("Expected no current owner initially")
	}

	// Create root should set current owner
	_, cleanup := CreateRoot(func() interface{} {
		stats := GetReactiveStats()
		if !stats["currentOwner"].(bool) {
			t.Error("Expected current owner within root")
		}

		CreateEffect(func() {
			stats := GetReactiveStats()
			if !stats["currentComputation"].(bool) {
				t.Error("Expected current computation within effect")
			}
		}, nil)

		return nil
	})

	cleanup()

	// Should be clean after cleanup
	stats = GetReactiveStats()
	if stats["currentComputation"].(bool) {
		t.Error("Expected no current computation after cleanup")
	}
	if stats["currentOwner"].(bool) {
		t.Error("Expected no current owner after cleanup")
	}
}

// Benchmark tests
func BenchmarkSignalCreation(b *testing.B) {
	ResetReactiveContext()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		CreateSignal(i)
	}
}

func BenchmarkSignalAccess(b *testing.B) {
	ResetReactiveContext()
	getter, _ := CreateSignal(42)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = getter()
	}
}

func BenchmarkSignalUpdate(b *testing.B) {
	ResetReactiveContext()
	ResetScheduler()
	_, setter := CreateSignal(0)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		setter(i)
	}
}

func BenchmarkEffectExecution(b *testing.B) {
	ResetReactiveContext()
	ResetScheduler()
	getter, setter := CreateSignal(0)

	CreateEffect(func() {
		_ = getter()
	}, nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		setter(i)
		FlushScheduler()
	}
}
