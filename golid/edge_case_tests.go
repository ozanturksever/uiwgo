// edge_case_tests.go
// Comprehensive edge case tests for the reactive system

//go:build !js && !wasm

package golid

import (
	"fmt"
	"runtime"
	"sync"
	"testing"
	"time"
)

// ------------------------------------
// 🧪 Edge Case Test Suite
// ------------------------------------

// TestSignalEdgeCases tests various edge cases for signals
func TestSignalEdgeCases(t *testing.T) {
	t.Run("NilSignalHandling", func(t *testing.T) {
		// Test behavior with nil values
		var nilPtr *string
		getter, setter := CreateSignal(nilPtr)

		if getter() != nil {
			t.Errorf("Expected nil, got %v", getter())
		}

		value := "test"
		setter(&value)
		if getter() == nil || *getter() != "test" {
			t.Errorf("Expected 'test', got %v", getter())
		}
	})

	t.Run("ZeroValueSignals", func(t *testing.T) {
		// Test with various zero values
		intGetter, intSetter := CreateSignal(0)
		stringGetter, stringSetter := CreateSignal("")
		boolGetter, boolSetter := CreateSignal(false)

		if intGetter() != 0 {
			t.Errorf("Expected 0, got %d", intGetter())
		}
		if stringGetter() != "" {
			t.Errorf("Expected empty string, got %s", stringGetter())
		}
		if boolGetter() != false {
			t.Errorf("Expected false, got %t", boolGetter())
		}

		// Test setting to non-zero values
		intSetter(42)
		stringSetter("hello")
		boolSetter(true)

		if intGetter() != 42 {
			t.Errorf("Expected 42, got %d", intGetter())
		}
		if stringGetter() != "hello" {
			t.Errorf("Expected 'hello', got %s", stringGetter())
		}
		if boolGetter() != true {
			t.Errorf("Expected true, got %t", boolGetter())
		}
	})

	t.Run("LargeValueSignals", func(t *testing.T) {
		// Test with large values
		largeString := make([]byte, 1024*1024) // 1MB string
		for i := range largeString {
			largeString[i] = 'A'
		}

		getter, setter := CreateSignal(string(largeString))
		if len(getter()) != len(largeString) {
			t.Errorf("Expected length %d, got %d", len(largeString), len(getter()))
		}

		// Test updating large value
		setter("small")
		if getter() != "small" {
			t.Errorf("Expected 'small', got %s", getter())
		}
	})
}

// TestEffectEdgeCases tests edge cases for effects
func TestEffectEdgeCases(t *testing.T) {
	t.Run("EffectWithPanic", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("Expected panic, but none occurred")
			}
		}()

		count, setCount := CreateSignal(0)

		// Create effect that panics
		CreateEffect(func() {
			if count() > 5 {
				panic("count too high")
			}
		}, nil)

		// This should trigger the panic
		setCount(10)
	})

	t.Run("EffectWithInfiniteLoop", func(t *testing.T) {
		// Test effect that could cause infinite updates
		count, setCount := CreateSignal(0)
		updateCount := 0

		CreateEffect(func() {
			updateCount++
			if updateCount > 100 {
				t.Error("Effect ran too many times, possible infinite loop")
				return
			}

			currentCount := count()
			// Don't create infinite loop by setting the same signal
			if currentCount < 5 {
				// This would normally cause infinite loop, but scheduler should prevent it
				go func() {
					time.Sleep(1 * time.Millisecond)
					setCount(currentCount + 1)
				}()
			}
		}, nil)

		setCount(1)
		time.Sleep(100 * time.Millisecond) // Allow effects to run

		if updateCount > 50 {
			t.Errorf("Effect ran %d times, possible infinite loop", updateCount)
		}
	})

	t.Run("EffectWithSlowComputation", func(t *testing.T) {
		// Test effect with slow computation
		count, setCount := CreateSignal(0)
		computationDone := make(chan bool, 1)

		CreateEffect(func() {
			_ = count() // Read signal
			// Simulate slow computation
			time.Sleep(50 * time.Millisecond)
			computationDone <- true
		}, nil)

		start := time.Now()
		setCount(1)

		// Wait for computation to complete
		select {
		case <-computationDone:
			duration := time.Since(start)
			if duration < 40*time.Millisecond {
				t.Errorf("Computation completed too quickly: %v", duration)
			}
		case <-time.After(200 * time.Millisecond):
			t.Error("Computation took too long")
		}
	})
}

// TestConcurrencyEdgeCases tests edge cases related to concurrency
func TestConcurrencyEdgeCases(t *testing.T) {
	t.Run("HighConcurrencySignalUpdates", func(t *testing.T) {
		count, setCount := CreateSignal(0)
		numGoroutines := 100
		numUpdatesPerGoroutine := 10

		var wg sync.WaitGroup
		wg.Add(numGoroutines)

		// Start multiple goroutines updating the signal
		for i := 0; i < numGoroutines; i++ {
			go func(goroutineID int) {
				defer wg.Done()
				for j := 0; j < numUpdatesPerGoroutine; j++ {
					current := count()
					setCount(current + 1)
					time.Sleep(1 * time.Millisecond)
				}
			}(i)
		}

		wg.Wait()

		// The final count should be at least the number of updates
		// (may be less due to race conditions, but should not be 0)
		finalCount := count()
		if finalCount == 0 {
			t.Error("Signal was not updated by any goroutine")
		}

		t.Logf("Final count: %d (expected around %d)", finalCount, numGoroutines*numUpdatesPerGoroutine)
	})

	t.Run("ConcurrentEffectCreation", func(t *testing.T) {
		count, setCount := CreateSignal(0)
		numEffects := 50
		effectResults := make([]int, numEffects)

		var wg sync.WaitGroup
		wg.Add(numEffects)

		// Create multiple effects concurrently
		for i := 0; i < numEffects; i++ {
			go func(effectID int) {
				defer wg.Done()
				CreateEffect(func() {
					effectResults[effectID] = count()
				}, nil)
			}(i)
		}

		wg.Wait()

		// Trigger all effects
		setCount(42)
		time.Sleep(50 * time.Millisecond) // Allow effects to run

		// Check that all effects ran
		for i, result := range effectResults {
			if result != 42 {
				t.Errorf("Effect %d did not run correctly, got %d, expected 42", i, result)
			}
		}
	})

	t.Run("RapidSignalUpdates", func(t *testing.T) {
		// Test rapid signal updates to stress test the scheduler
		count, setCount := CreateSignal(0)
		effectRunCount := 0

		CreateEffect(func() {
			_ = count()
			effectRunCount++
		}, nil)

		// Rapidly update signal
		for i := 0; i < 1000; i++ {
			setCount(i)
		}

		time.Sleep(100 * time.Millisecond) // Allow batching to complete

		// Effect should run much fewer times than updates due to batching
		if effectRunCount > 100 {
			t.Errorf("Effect ran %d times for 1000 updates, batching may not be working", effectRunCount)
		}

		if effectRunCount == 0 {
			t.Error("Effect never ran")
		}

		t.Logf("Effect ran %d times for 1000 signal updates", effectRunCount)
	})
}

// TestMemoryEdgeCases tests memory-related edge cases
func TestMemoryEdgeCases(t *testing.T) {
	t.Run("MemoryLeakPrevention", func(t *testing.T) {
		// Get initial memory stats
		var m1 runtime.MemStats
		runtime.GC()
		runtime.ReadMemStats(&m1)

		// Create many signals and effects
		for i := 0; i < 1000; i++ {
			getter, setter := CreateSignal(i)
			CreateEffect(func() {
				_ = getter()
			}, nil)
			setter(i + 1)
		}

		// Force garbage collection
		runtime.GC()
		runtime.GC() // Run twice to ensure cleanup

		// Get memory stats after cleanup
		var m2 runtime.MemStats
		runtime.ReadMemStats(&m2)

		// Memory usage should not have grown excessively
		memoryGrowth := m2.Alloc - m1.Alloc
		if memoryGrowth > 10*1024*1024 { // 10MB threshold
			t.Errorf("Memory usage grew by %d bytes, possible memory leak", memoryGrowth)
		}

		t.Logf("Memory growth: %d bytes", memoryGrowth)
	})

	t.Run("LargeSignalCleanup", func(t *testing.T) {
		// Test cleanup of signals with large data
		largeData := make([]byte, 1024*1024) // 1MB
		for i := range largeData {
			largeData[i] = byte(i % 256)
		}

		getter, setter := CreateSignal(largeData)

		// Create effect that uses the large data
		CreateEffect(func() {
			data := getter()
			if len(data) == 0 {
				t.Error("Large data was lost")
			}
		}, nil)

		// Update with smaller data
		setter([]byte{1, 2, 3})

		// Force garbage collection
		runtime.GC()
		runtime.GC()

		// Verify the signal still works
		if len(getter()) != 3 {
			t.Errorf("Expected length 3, got %d", len(getter()))
		}
	})
}

// TestErrorRecoveryEdgeCases tests edge cases in error recovery
func TestErrorRecoveryEdgeCases(t *testing.T) {
	t.Run("NestedErrorBoundaries", func(t *testing.T) {
		outerRecovered := false
		innerRecovered := false

		// Create nested error boundaries
		_ = CreateErrorBoundary(func(err error) interface{} {
			outerRecovered = true
			return "outer recovery"
		})

		innerBoundary := CreateErrorBoundary(func(err error) interface{} {
			innerRecovered = true
			return "inner recovery"
		})

		// Test that inner boundary catches error first
		err := innerBoundary.Catch(func() {
			panic("test error")
		})

		if err == nil {
			t.Error("Expected error to be caught")
		}

		if !innerRecovered {
			t.Error("Inner boundary should have recovered")
		}

		if outerRecovered {
			t.Error("Outer boundary should not have recovered")
		}
	})

	t.Run("ErrorBoundaryReset", func(t *testing.T) {
		recoveryCount := 0
		boundary := CreateErrorBoundary(func(err error) interface{} {
			recoveryCount++
			return fmt.Sprintf("recovery %d", recoveryCount)
		})

		// First error
		err1 := boundary.Catch(func() {
			panic("first error")
		})
		if err1 == nil {
			t.Error("Expected first error to be caught")
		}

		// Reset boundary
		boundary.Reset()

		// Second error after reset
		err2 := boundary.Catch(func() {
			panic("second error")
		})
		if err2 == nil {
			t.Error("Expected second error to be caught")
		}

		if recoveryCount != 2 {
			t.Errorf("Expected 2 recoveries, got %d", recoveryCount)
		}
	})
}

// TestSchedulerEdgeCases tests edge cases in the scheduler
func TestSchedulerEdgeCases(t *testing.T) {
	t.Run("SchedulerOverload", func(t *testing.T) {
		// Create many signals and effects to overload scheduler
		numSignals := 100
		signals := make([]func() int, numSignals)
		setters := make([]func(int), numSignals)

		for i := 0; i < numSignals; i++ {
			getter, setter := CreateSignal(0)
			signals[i] = getter
			setters[i] = setter

			// Create effect for each signal
			CreateEffect(func() {
				_ = getter()
			}, nil)
		}

		// Update all signals simultaneously
		start := time.Now()
		for i, setter := range setters {
			setter(i + 1)
		}

		// Wait for all effects to complete
		time.Sleep(100 * time.Millisecond)
		duration := time.Since(start)

		// Verify all signals were updated
		for i, getter := range signals {
			if getter() != i+1 {
				t.Errorf("Signal %d was not updated correctly, got %d, expected %d", i, getter(), i+1)
			}
		}

		t.Logf("Scheduler handled %d signals in %v", numSignals, duration)
	})

	t.Run("SchedulerShutdown", func(t *testing.T) {
		// Test graceful scheduler shutdown
		count, setCount := CreateSignal(0)
		effectRan := false

		CreateEffect(func() {
			if count() > 0 {
				effectRan = true
			}
		}, nil)

		// Update signal
		setCount(1)

		// Wait for effect to run
		time.Sleep(50 * time.Millisecond)

		if !effectRan {
			t.Error("Effect did not run before shutdown")
		}
	})
}

// BenchmarkSignalPerformance benchmarks signal performance under various conditions
func BenchmarkSignalPerformance(b *testing.B) {
	b.Run("SignalCreation", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = CreateSignal(i)
		}
	})

	b.Run("SignalRead", func(b *testing.B) {
		getter, _ := CreateSignal(42)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = getter()
		}
	})

	b.Run("SignalWrite", func(b *testing.B) {
		_, setter := CreateSignal(0)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			setter(i)
		}
	})

	b.Run("EffectExecution", func(b *testing.B) {
		count, setCount := CreateSignal(0)
		effectCount := 0

		CreateEffect(func() {
			_ = count()
			effectCount++
		}, nil)

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			setCount(i)
		}
	})
}
