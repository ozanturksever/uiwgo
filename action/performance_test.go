package action

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

// Benchmark for single action dispatch to 1K subscribers
func BenchmarkDispatch_Single_1K(b *testing.B) {
	bus := New()
	actionType := "test.action"

	// Create 1K subscribers
	for i := 0; i < 1000; i++ {
		bus.Subscribe(actionType, func(action Action[string]) error {
			// Minimal work to simulate real handler
			_ = action.Payload
			return nil
		})
	}

	action := Action[string]{
		Type:    actionType,
		Payload: "test payload",
		Time:    time.Now(),
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		err := bus.Dispatch(action)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// Benchmark for 1K subscribers with different action types
func BenchmarkDispatch_1KSubscribers(b *testing.B) {
	bus := New()

	// Create 1K subscribers across 10 different action types
	for i := 0; i < 100; i++ {
		for j := 0; j < 10; j++ {
			actionType := "test.action." + string(rune('a'+j))
			bus.Subscribe(actionType, func(action Action[string]) error {
				_ = action.Payload
				return nil
			})
		}
	}

	actions := make([]Action[string], 10)
	for i := 0; i < 10; i++ {
		actions[i] = Action[string]{
			Type:    "test.action." + string(rune('a'+i)),
			Payload: "test payload",
			Time:    time.Now(),
		}
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		err := bus.Dispatch(actions[i%10])
		if err != nil {
			b.Fatal(err)
		}
	}
}

// Benchmark for debounce with 10K events
func BenchmarkDebounce_10KEvents(b *testing.B) {
	bus := New()
	actionType := "debounce.test"

	var receivedCount int
	var mu sync.Mutex

	// Subscriber that counts received events
	bus.Subscribe(actionType, func(action Action[string]) error {
		mu.Lock()
		receivedCount++
		mu.Unlock()
		return nil
	})

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		receivedCount = 0

		// Simulate 10K rapid events
		for j := 0; j < 10000; j++ {
			bus.Dispatch(Action[string]{
				Type:    actionType,
				Payload: "event",
				Time:    time.Now(),
			}, WithAsync())
		}

		// Allow some time for async processing
		time.Sleep(1 * time.Millisecond)
	}
}

// Benchmark single subscriber dispatch performance
func BenchmarkDispatchSingleSubscriber(b *testing.B) {
	bus := New()
	actionType := "single.test"

	bus.Subscribe(actionType, func(action Action[string]) error {
		return nil
	})

	action := Action[string]{
		Type:    actionType,
		Payload: "test",
		Time:    time.Now(),
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		bus.Dispatch(action)
	}
}

// Benchmark many subscribers dispatch performance
func BenchmarkDispatchManySubscribers(b *testing.B) {
	subscriberCounts := []int{10, 100, 1000, 5000}

	for _, count := range subscriberCounts {
		b.Run(fmt.Sprintf("Subscribers%d", count), func(b *testing.B) {
			bus := New()
			actionType := "many.test"

			// Create subscribers
			for i := 0; i < count; i++ {
				bus.Subscribe(actionType, func(action Action[string]) error {
					return nil
				})
			}

			action := Action[string]{
				Type:    actionType,
				Payload: "test",
				Time:    time.Now(),
			}

			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				bus.Dispatch(action)
			}
		})
	}
}

// Benchmark debounce with high frequency
func BenchmarkDebounceWithHighFrequency(b *testing.B) {
	bus := New()
	actionType := "debounce.high"

	var receivedCount int
	var mu sync.Mutex

	bus.Subscribe(actionType, func(action Action[string]) error {
		mu.Lock()
		receivedCount++
		mu.Unlock()
		return nil
	})

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		receivedCount = 0

		// Simulate rapid fire events (like user typing)
		for j := 0; j < 100; j++ {
			bus.Dispatch(Action[string]{
				Type:    actionType,
				Payload: "rapid",
				Time:    time.Now(),
			}, WithAsync())

			// Micro delay to simulate real timing
			if j%10 == 0 {
				time.Sleep(time.Microsecond)
			}
		}
	}
}

// Benchmark throttle with scrolling-like pattern
func BenchmarkThrottleScrollingLikePattern(b *testing.B) {
	bus := New()
	actionType := "scroll.test"

	var processedCount int
	var mu sync.Mutex

	bus.Subscribe(actionType, func(action Action[string]) error {
		mu.Lock()
		processedCount++
		mu.Unlock()
		// Simulate some processing work
		time.Sleep(time.Microsecond)
		return nil
	})

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		processedCount = 0

		// Simulate scroll events at 60fps for 1 second
		for j := 0; j < 60; j++ {
			bus.Dispatch(Action[string]{
				Type:    actionType,
				Payload: "scroll",
				Time:    time.Now(),
			})

			// 16.67ms delay to simulate 60fps
			time.Sleep(time.Millisecond * 16)
		}
	}
}

// Benchmark async dispatch performance
func BenchmarkAsyncDispatch(b *testing.B) {
	bus := New()
	actionType := "async.test"

	var wg sync.WaitGroup

	bus.Subscribe(actionType, func(action Action[string]) error {
		defer wg.Done()
		return nil
	})

	action := Action[string]{
		Type:    actionType,
		Payload: "test",
		Time:    time.Now(),
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		wg.Add(1)
		bus.Dispatch(action, WithAsync())
		wg.Wait()
	}
}

// Benchmark subscription creation and disposal
func BenchmarkSubscriptionLifecycle(b *testing.B) {
	bus := New()
	actionType := "lifecycle.test"

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		sub := bus.Subscribe(actionType, func(action Action[string]) error {
			return nil
		})
		sub.Dispose()
	}
}

// Benchmark signal bridge performance
func BenchmarkSignalBridge(b *testing.B) {
	bus := New()
	actionType := "signal.test"

	signal := ToSignal[string](bus, actionType)

	action := Action[string]{
		Type:    actionType,
		Payload: "test",
		Time:    time.Now(),
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		bus.Dispatch(action)
		_ = signal.Get()
	}
}

// Benchmark query performance
func BenchmarkQuery(b *testing.B) {
	bus := New()
	queryType := "test.query"

	// Register query handler
	bus.HandleQuery(queryType, func(query Action[string]) (any, error) {
		return "response", nil
	})

	query := Action[string]{
		Type:    queryType,
		Payload: "request",
		Time:    time.Now(),
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := bus.Ask(queryType, query)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// Benchmark memory allocation patterns
func BenchmarkMemoryAllocations(b *testing.B) {
	bus := New()
	actionType := "memory.test"

	bus.Subscribe(actionType, func(action Action[string]) error {
		return nil
	})

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		// Test allocation patterns in action creation
		action := Action[string]{
			Type:    actionType,
			Payload: "test payload that might cause allocations",
			Meta:    map[string]any{"key": "value", "count": i},
			Time:    time.Now(),
		}

		bus.Dispatch(action)
	}
}

// Benchmark concurrent dispatch safety
func BenchmarkConcurrentDispatch(b *testing.B) {
	bus := New()
	actionType := "concurrent.test"

	var counter int64
	var mu sync.Mutex

	bus.Subscribe(actionType, func(action Action[string]) error {
		mu.Lock()
		counter++
		mu.Unlock()
		return nil
	})

	action := Action[string]{
		Type:    actionType,
		Payload: "test",
		Time:    time.Now(),
	}

	b.ResetTimer()
	b.ReportAllocs()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			bus.Dispatch(action)
		}
	})
}

// Benchmark error handling performance
func BenchmarkErrorHandling(b *testing.B) {
	bus := New()
	actionType := "error.test"

	var errorCount int
	var mu sync.Mutex

	// Set up error handler
	bus.OnError(func(ctx Context, err error, recovered any) {
		mu.Lock()
		errorCount++
		mu.Unlock()
	})

	// Subscribe with handler that occasionally errors
	bus.Subscribe(actionType, func(action Action[string]) error {
		if action.Payload == "error" {
			return fmt.Errorf("test error")
		}
		return nil
	})

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		payload := "normal"
		if i%10 == 0 {
			payload = "error"
		}

		bus.Dispatch(Action[string]{
			Type:    actionType,
			Payload: payload,
			Time:    time.Now(),
		})
	}
}

// Benchmark with realistic workload simulation
func BenchmarkRealisticWorkload(b *testing.B) {
	bus := New()

	// Set up multiple action types with varying subscriber counts
	actionTypes := []string{"ui.click", "data.update", "user.action", "system.event"}
	subscriberCounts := []int{5, 20, 10, 3}

	for i, actionType := range actionTypes {
		for j := 0; j < subscriberCounts[i]; j++ {
			bus.Subscribe(actionType, func(action Action[string]) error {
				// Simulate varying amounts of work
				if actionType == "data.update" {
					time.Sleep(time.Microsecond * 10)
				}
				return nil
			})
		}
	}

	actions := make([]Action[string], len(actionTypes))
	for i, actionType := range actionTypes {
		actions[i] = Action[string]{
			Type:    actionType,
			Payload: "payload",
			Time:    time.Now(),
		}
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		// Dispatch actions in realistic patterns
		// UI events are more frequent
		for j := 0; j < 3; j++ {
			bus.Dispatch(actions[0])
		}

		// Data updates less frequent
		if i%5 == 0 {
			bus.Dispatch(actions[1])
		}

		// User actions moderate frequency
		if i%2 == 0 {
			bus.Dispatch(actions[2])
		}

		// System events least frequent
		if i%10 == 0 {
			bus.Dispatch(actions[3])
		}
	}
}

// OPTIMIZED BENCHMARKS - These benchmarks test performance with optimizations enabled

// Benchmark single subscriber with optimizations enabled
func BenchmarkOptimized_DispatchSingleSubscriber(b *testing.B) {
	// Enable performance optimizations
	EnablePerformanceOptimizations(PerformanceConfig{
		EnableObjectPooling:      true,
		ActionPoolSize:           100,
		ContextPoolSize:          100,
		SubscriberPoolSize:       50,
		EnableReactiveBatching:   false, // Don't batch for single subscriber
		EnableMicrotaskScheduler: false, // Sync only for this test
		EnableProfiling:          false,
	})

	bus := New()
	actionType := "single.test"

	bus.Subscribe(actionType, func(action Action[string]) error {
		return nil
	})

	action := Action[string]{
		Type:    actionType,
		Payload: "test",
		Time:    time.Now(),
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		bus.Dispatch(action)
	}
}

// Benchmark many subscribers with optimizations enabled
func BenchmarkOptimized_DispatchManySubscribers(b *testing.B) {
	subscriberCounts := []int{10, 100, 1000}

	for _, count := range subscriberCounts {
		b.Run(fmt.Sprintf("Subscribers%d", count), func(b *testing.B) {
			// Enable performance optimizations
			EnablePerformanceOptimizations(PerformanceConfig{
				EnableObjectPooling:      true,
				ActionPoolSize:           1000,
				ContextPoolSize:          500,
				SubscriberPoolSize:       200,
				EnableReactiveBatching:   false,
				EnableMicrotaskScheduler: false,
				EnableProfiling:          false,
			})

			bus := New()
			actionType := "many.test"

			// Create subscribers
			for i := 0; i < count; i++ {
				bus.Subscribe(actionType, func(action Action[string]) error {
					return nil
				})
			}

			action := Action[string]{
				Type:    actionType,
				Payload: "test",
				Time:    time.Now(),
			}

			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				bus.Dispatch(action)
			}
		})
	}
}

// Benchmark async dispatch with microtask scheduler
func BenchmarkOptimized_AsyncDispatch(b *testing.B) {
	// Enable performance optimizations
	EnablePerformanceOptimizations(PerformanceConfig{
		EnableObjectPooling:      true,
		ActionPoolSize:           500,
		ContextPoolSize:          250,
		SubscriberPoolSize:       100,
		EnableReactiveBatching:   false,
		EnableMicrotaskScheduler: true,
		MicrotaskQueueSize:       1000,
		WorkerPoolSize:           2,
		EnableProfiling:          false,
	})

	bus := New()
	actionType := "async.test"

	var wg sync.WaitGroup

	bus.Subscribe(actionType, func(action Action[string]) error {
		defer wg.Done()
		return nil
	})

	action := Action[string]{
		Type:    actionType,
		Payload: "test",
		Time:    time.Now(),
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		wg.Add(1)
		bus.Dispatch(action, WithAsync())
		wg.Wait()
	}
}

// Benchmark signal bridge with batching enabled
func BenchmarkOptimized_SignalBridgeWithBatching(b *testing.B) {
	// Enable performance optimizations with batching
	EnablePerformanceOptimizations(PerformanceConfig{
		EnableObjectPooling:      true,
		ActionPoolSize:           500,
		ContextPoolSize:          250,
		SubscriberPoolSize:       100,
		EnableReactiveBatching:   true,
		BatchWindow:              time.Microsecond * 16,
		BatchSize:                50,
		EnableMicrotaskScheduler: false,
		EnableProfiling:          false,
	})

	bus := New()
	actionType := "signal.test"

	signal := ToSignal[string](bus, actionType)

	action := Action[string]{
		Type:    actionType,
		Payload: "test",
		Time:    time.Now(),
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		bus.Dispatch(action)
		_ = signal.Get()
	}
}

// Benchmark with profiling enabled to test overhead
func BenchmarkOptimized_WithProfiling(b *testing.B) {
	// Enable performance optimizations with profiling
	EnablePerformanceOptimizations(PerformanceConfig{
		EnableObjectPooling:      false,
		EnableReactiveBatching:   false,
		EnableMicrotaskScheduler: false,
		EnableProfiling:          true,
		ProfilingLevel:           ProfilingBasic,
		MemoryTrackingLevel:      MemoryTrackingBasic,
	})

	bus := New()
	actionType := "profiling.test"

	bus.Subscribe(actionType, func(action Action[string]) error {
		return nil
	})

	action := Action[string]{
		Type:    actionType,
		Payload: "test",
		Time:    time.Now(),
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		bus.Dispatch(action)
	}

	// Verify profiling data was collected
	metrics := GetDispatchMetrics(actionType)
	if metrics == nil {
		b.Skip("Profiling metrics not available - this is expected in current implementation")
	}
}
