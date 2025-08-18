// signals_goroutine_test.go
// Tests for goroutine management and resource conservation in Signal.Set

package golid

import (
	"runtime"
	"sync"
	"testing"
	"time"
)

// TestSignalGoroutineManagement tests that Signal.Set doesn't create goroutine explosion
func TestSignalGoroutineManagement(t *testing.T) {
	// Record initial goroutine count
	initialGoroutines := runtime.NumGoroutine()

	// Create a signal with multiple watchers
	signal := NewSignal(0)
	var executionCounts []int
	var mutexes []sync.Mutex

	numWatchers := 100
	executionCounts = make([]int, numWatchers)
	mutexes = make([]sync.Mutex, numWatchers)

	// Create many watchers
	for i := 0; i < numWatchers; i++ {
		watcherIndex := i
		Watch(func() {
			_ = signal.Get()
			mutexes[watcherIndex].Lock()
			executionCounts[watcherIndex]++
			mutexes[watcherIndex].Unlock()
		})
	}

	// Allow initial effects to complete
	time.Sleep(10 * time.Millisecond)

	// Record goroutine count after watchers are set up
	afterWatchersGoroutines := runtime.NumGoroutine()

	// Perform many rapid signal updates
	numUpdates := 50
	var wg sync.WaitGroup

	for i := 0; i < numUpdates; i++ {
		wg.Add(1)
		go func(val int) {
			defer wg.Done()
			signal.Set(val)
		}(i)
	}

	wg.Wait()

	// Allow all effects to complete
	time.Sleep(100 * time.Millisecond)

	// Record final goroutine count
	finalGoroutines := runtime.NumGoroutine()

	// With the old implementation, we would have created numUpdates * numWatchers goroutines
	// With the new implementation, we should have a bounded number of worker goroutines

	t.Logf("Initial goroutines: %d", initialGoroutines)
	t.Logf("After watchers setup: %d", afterWatchersGoroutines)
	t.Logf("After %d updates with %d watchers: %d", numUpdates, numWatchers, finalGoroutines)

	// The increase in goroutines should be bounded (worker pool + reasonable overhead)
	// Not the explosive numUpdates * numWatchers that would happen with the old approach
	maxExpectedIncrease := 50 // Worker goroutines + some reasonable overhead
	actualIncrease := finalGoroutines - initialGoroutines

	if actualIncrease > maxExpectedIncrease {
		t.Errorf("Goroutine increase too high: %d (expected <= %d). Possible goroutine explosion.",
			actualIncrease, maxExpectedIncrease)
	}

	// Verify that effects were still executed
	totalExecutions := 0
	for i := 0; i < numWatchers; i++ {
		mutexes[i].Lock()
		count := executionCounts[i]
		mutexes[i].Unlock()
		totalExecutions += count
	}

	if totalExecutions == 0 {
		t.Error("No effects were executed - worker pool might be broken")
	}

	t.Logf("Total effect executions: %d", totalExecutions)
}

// TestSignalWorkerPoolQueueHandling tests queue overflow handling
func TestSignalWorkerPoolQueueHandling(t *testing.T) {
	signal := NewSignal(0)
	var totalExecutions int32
	var mutex sync.Mutex

	// Create a watcher with a slow effect to fill up the queue
	Watch(func() {
		_ = signal.Get()
		time.Sleep(1 * time.Millisecond) // Simulate slow effect
		mutex.Lock()
		totalExecutions++
		mutex.Unlock()
	})

	// Perform rapid updates to potentially overflow the queue
	numUpdates := 1500 // More than the queue buffer size (1000)
	var wg sync.WaitGroup

	start := time.Now()
	for i := 0; i < numUpdates; i++ {
		wg.Add(1)
		go func(val int) {
			defer wg.Done()
			signal.Set(val)
		}(i)
	}

	wg.Wait()
	duration := time.Since(start)

	// Allow effects to complete
	time.Sleep(200 * time.Millisecond)

	mutex.Lock()
	finalExecutions := totalExecutions
	mutex.Unlock()

	t.Logf("Updates completed in: %v", duration)
	t.Logf("Effect executions: %d", finalExecutions)

	// Should complete reasonably quickly (not hang due to blocking)
	if duration > 5*time.Second {
		t.Errorf("Updates took too long (%v), suggesting blocking behavior", duration)
	}

	// Should have executed effects (maybe not all due to queue overflow, but some)
	if finalExecutions == 0 {
		t.Error("No effects executed - worker pool might be completely broken")
	}

	// With the fallback mechanism, we should get some effects even with queue overflow
	if finalExecutions < int32(numUpdates/10) { // At least 10% should execute
		t.Logf("Warning: Low execution rate (%d/%d) - queue might be overflowing frequently",
			finalExecutions, numUpdates)
	}
}

// TestSignalResourceConservation tests memory and resource usage
func TestSignalResourceConservation(t *testing.T) {
	var m1, m2 runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&m1)

	// Create signals and perform operations
	signals := make([]*Signal[int], 10)
	for i := range signals {
		signals[i] = NewSignal(i)

		// Add watchers
		Watch(func() {
			_ = signals[i].Get()
		})
	}

	// Perform many operations
	for round := 0; round < 100; round++ {
		for i, signal := range signals {
			signal.Set(round * i)
		}
	}

	// Allow operations to complete
	time.Sleep(50 * time.Millisecond)

	runtime.GC()
	runtime.ReadMemStats(&m2)

	allocDiff := m2.TotalAlloc - m1.TotalAlloc
	t.Logf("Memory allocated during test: %d bytes", allocDiff)

	// Should not allocate excessive memory (this is a rough check)
	// The exact threshold depends on the system, but we're looking for obvious leaks
	maxExpectedAlloc := uint64(10 * 1024 * 1024) // 10MB seems reasonable for this test
	if allocDiff > maxExpectedAlloc {
		t.Errorf("Excessive memory allocation: %d bytes (expected < %d)",
			allocDiff, maxExpectedAlloc)
	}
}
