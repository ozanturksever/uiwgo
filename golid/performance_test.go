//go:build js && wasm

// performance_test.go
// Performance regression tests and benchmarks for reactive systems

package golid

import (
	"fmt"
	"runtime"
	"sync"
	"sync/atomic"
	"syscall/js"
	"testing"
	"time"
)

// ------------------------------------
// 🎯 Performance Test Configuration
// ------------------------------------

// PerformanceTargets defines the performance targets from the roadmap
type PerformanceTargets struct {
	SignalUpdateTime     time.Duration // Target: 5μs
	DOMUpdateBatch       time.Duration // Target: 10ms
	MemoryPerSignal      uint64        // Target: 200B
	MaxConcurrentEffects int           // Target: 10,000
	MaxCascadeDepth      int           // Target: 10
}

// GetPerformanceTargets returns the performance targets from the implementation roadmap
func GetPerformanceTargets() PerformanceTargets {
	return PerformanceTargets{
		SignalUpdateTime:     5 * time.Microsecond,
		DOMUpdateBatch:       10 * time.Millisecond,
		MemoryPerSignal:      200, // 200 bytes
		MaxConcurrentEffects: 10000,
		MaxCascadeDepth:      10,
	}
}

// PerformanceResult contains the results of a performance test
type PerformanceResult struct {
	Operation       string
	Duration        time.Duration
	MemoryUsed      uint64
	OperationsCount int
	Target          time.Duration
	Passed          bool
	Improvement     float64 // Improvement factor vs target
}

// PerformanceBenchmark manages performance testing
type PerformanceBenchmark struct {
	targets PerformanceTargets
	results []PerformanceResult
	mutex   sync.Mutex
}

// NewPerformanceBenchmark creates a new performance benchmark
func NewPerformanceBenchmark() *PerformanceBenchmark {
	return &PerformanceBenchmark{
		targets: GetPerformanceTargets(),
		results: make([]PerformanceResult, 0),
	}
}

// RecordResult records a performance test result
func (pb *PerformanceBenchmark) RecordResult(result PerformanceResult) {
	pb.mutex.Lock()
	defer pb.mutex.Unlock()
	pb.results = append(pb.results, result)
}

// ValidateResults validates all recorded results against targets
func (pb *PerformanceBenchmark) ValidateResults(t *testing.T) {
	pb.mutex.Lock()
	defer pb.mutex.Unlock()

	for _, result := range pb.results {
		if !result.Passed {
			t.Errorf("Performance regression in %s: %v > %v (%.2fx slower than target)",
				result.Operation, result.Duration, result.Target, 1.0/result.Improvement)
		} else {
			t.Logf("Performance target met for %s: %v < %v (%.2fx faster than target)",
				result.Operation, result.Duration, result.Target, result.Improvement)
		}
	}
}

// ------------------------------------
// 🎯 Signal Performance Tests
// ------------------------------------

func TestSignalUpdatePerformance(t *testing.T) {
	benchmark := NewPerformanceBenchmark()
	targets := benchmark.targets

	ResetReactiveContext()
	ResetScheduler()

	// Test single signal update performance
	t.Run("SingleSignalUpdate", func(t *testing.T) {
		_, setter := CreateSignal(0)

		// Warm up
		for i := 0; i < 100; i++ {
			setter(i)
			FlushScheduler()
		}

		// Measure performance
		start := time.Now()
		iterations := 1000

		for i := 0; i < iterations; i++ {
			setter(i)
			FlushScheduler()
		}

		duration := time.Since(start)
		avgDuration := duration / time.Duration(iterations)

		result := PerformanceResult{
			Operation:       "SingleSignalUpdate",
			Duration:        avgDuration,
			OperationsCount: iterations,
			Target:          targets.SignalUpdateTime,
			Passed:          avgDuration <= targets.SignalUpdateTime,
			Improvement:     float64(targets.SignalUpdateTime) / float64(avgDuration),
		}

		benchmark.RecordResult(result)
	})

	// Test batched signal updates
	t.Run("BatchedSignalUpdates", func(t *testing.T) {
		signals := make([]func(int), 100)
		for i := 0; i < 100; i++ {
			_, setter := CreateSignal(0)
			signals[i] = setter
		}

		// Warm up
		Batch(func() {
			for i, setter := range signals {
				setter(i)
			}
		})
		FlushScheduler()

		// Measure batched update performance
		start := time.Now()
		iterations := 100

		for iter := 0; iter < iterations; iter++ {
			Batch(func() {
				for i, setter := range signals {
					setter(iter*100 + i)
				}
			})
			FlushScheduler()
		}

		duration := time.Since(start)
		avgDuration := duration / time.Duration(iterations)

		result := PerformanceResult{
			Operation:       "BatchedSignalUpdates",
			Duration:        avgDuration,
			OperationsCount: iterations * 100,
			Target:          targets.SignalUpdateTime * 100, // Allow for batch overhead
			Passed:          avgDuration <= targets.SignalUpdateTime*100,
			Improvement:     float64(targets.SignalUpdateTime*100) / float64(avgDuration),
		}

		benchmark.RecordResult(result)
	})

	benchmark.ValidateResults(t)
}

func TestEffectPerformance(t *testing.T) {
	benchmark := NewPerformanceBenchmark()
	targets := benchmark.targets

	ResetReactiveContext()
	ResetScheduler()

	t.Run("EffectExecutionPerformance", func(t *testing.T) {
		getter, setter := CreateSignal(0)
		var effectExecutions int32

		// Create effect
		CreateEffect(func() {
			_ = getter()
			atomic.AddInt32(&effectExecutions, 1)
		}, nil)

		FlushScheduler()

		// Warm up
		for i := 0; i < 100; i++ {
			setter(i)
			FlushScheduler()
		}

		// Reset counter
		atomic.StoreInt32(&effectExecutions, 0)

		// Measure effect execution performance
		start := time.Now()
		iterations := 1000

		for i := 0; i < iterations; i++ {
			setter(i)
			FlushScheduler()
		}

		duration := time.Since(start)
		avgDuration := duration / time.Duration(iterations)
		finalExecutions := atomic.LoadInt32(&effectExecutions)

		if int(finalExecutions) != iterations {
			t.Errorf("Expected %d effect executions, got %d", iterations, finalExecutions)
		}

		result := PerformanceResult{
			Operation:       "EffectExecution",
			Duration:        avgDuration,
			OperationsCount: iterations,
			Target:          targets.SignalUpdateTime * 2, // Effects should be ~2x signal update time
			Passed:          avgDuration <= targets.SignalUpdateTime*2,
			Improvement:     float64(targets.SignalUpdateTime*2) / float64(avgDuration),
		}

		benchmark.RecordResult(result)
	})

	benchmark.ValidateResults(t)
}

func TestMemoPerformance(t *testing.T) {
	benchmark := NewPerformanceBenchmark()
	targets := benchmark.targets

	ResetReactiveContext()
	ResetScheduler()

	t.Run("MemoComputationPerformance", func(t *testing.T) {
		getter, setter := CreateSignal(0)
		var computations int32

		// Create memo with expensive computation
		memo := CreateMemo(func() int {
			atomic.AddInt32(&computations, 1)
			value := getter()
			// Simulate some computation
			result := value
			for i := 0; i < 10; i++ {
				result = result*2 + 1
			}
			return result
		}, nil)

		FlushScheduler()

		// Warm up
		for i := 0; i < 100; i++ {
			setter(i)
			_ = memo()
			FlushScheduler()
		}

		// Reset counter
		atomic.StoreInt32(&computations, 0)

		// Measure memo performance
		start := time.Now()
		iterations := 1000

		for i := 0; i < iterations; i++ {
			setter(i)
			_ = memo()
			FlushScheduler()
		}

		duration := time.Since(start)
		avgDuration := duration / time.Duration(iterations)
		finalComputations := atomic.LoadInt32(&computations)

		if int(finalComputations) != iterations {
			t.Errorf("Expected %d memo computations, got %d", iterations, finalComputations)
		}

		result := PerformanceResult{
			Operation:       "MemoComputation",
			Duration:        avgDuration,
			OperationsCount: iterations,
			Target:          targets.SignalUpdateTime * 5, // Memos can be slower due to computation
			Passed:          avgDuration <= targets.SignalUpdateTime*5,
			Improvement:     float64(targets.SignalUpdateTime*5) / float64(avgDuration),
		}

		benchmark.RecordResult(result)
	})

	benchmark.ValidateResults(t)
}

// ------------------------------------
// 🎯 DOM Performance Tests
// ------------------------------------

func TestDOMUpdatePerformance(t *testing.T) {
	benchmark := NewPerformanceBenchmark()
	targets := benchmark.targets

	ResetReactiveContext()
	ResetScheduler()

	// Create test environment
	document := js.Global().Get("document")
	container := document.Call("createElement", "div")
	document.Get("body").Call("appendChild", container)

	defer func() {
		document.Get("body").Call("removeChild", container)
	}()

	t.Run("DOMTextUpdatePerformance", func(t *testing.T) {
		elements := make([]js.Value, 100)
		signals := make([]func(string), 100)

		// Create elements and bind them
		for i := 0; i < 100; i++ {
			element := document.Call("createElement", "div")
			container.Call("appendChild", element)
			elements[i] = element

			getter, setter := CreateSignal(fmt.Sprintf("initial-%d", i))
			signals[i] = setter

			BindTextReactive(element, func() string {
				return getter()
			})
		}

		FlushScheduler()

		// Warm up
		for i := 0; i < 10; i++ {
			for j, setter := range signals {
				setter(fmt.Sprintf("warmup-%d-%d", i, j))
			}
			FlushScheduler()
		}

		// Measure DOM update performance
		start := time.Now()
		iterations := 100

		for iter := 0; iter < iterations; iter++ {
			for j, setter := range signals {
				setter(fmt.Sprintf("test-%d-%d", iter, j))
			}
			FlushScheduler()
		}

		duration := time.Since(start)
		avgDuration := duration / time.Duration(iterations)

		result := PerformanceResult{
			Operation:       "DOMTextUpdate",
			Duration:        avgDuration,
			OperationsCount: iterations * 100,
			Target:          targets.DOMUpdateBatch,
			Passed:          avgDuration <= targets.DOMUpdateBatch,
			Improvement:     float64(targets.DOMUpdateBatch) / float64(avgDuration),
		}

		benchmark.RecordResult(result)

		// Cleanup
		for _, element := range elements {
			container.Call("removeChild", element)
		}
	})

	t.Run("DOMAttributeUpdatePerformance", func(t *testing.T) {
		elements := make([]js.Value, 50)
		signals := make([]func(bool), 50)

		// Create elements and bind attributes
		for i := 0; i < 50; i++ {
			element := document.Call("createElement", "div")
			container.Call("appendChild", element)
			elements[i] = element

			getter, setter := CreateSignal(false)
			signals[i] = setter

			BindAttributeReactive(element, "data-active", func() string {
				if getter() {
					return "true"
				}
				return "false"
			})
		}

		FlushScheduler()

		// Measure attribute update performance
		start := time.Now()
		iterations := 200

		for iter := 0; iter < iterations; iter++ {
			for _, setter := range signals {
				setter(iter%2 == 0)
			}
			FlushScheduler()
		}

		duration := time.Since(start)
		avgDuration := duration / time.Duration(iterations)

		result := PerformanceResult{
			Operation:       "DOMAttributeUpdate",
			Duration:        avgDuration,
			OperationsCount: iterations * 50,
			Target:          targets.DOMUpdateBatch,
			Passed:          avgDuration <= targets.DOMUpdateBatch,
			Improvement:     float64(targets.DOMUpdateBatch) / float64(avgDuration),
		}

		benchmark.RecordResult(result)

		// Cleanup
		for _, element := range elements {
			container.Call("removeChild", element)
		}
	})

	benchmark.ValidateResults(t)
}

// ------------------------------------
// 🎯 Memory Performance Tests
// ------------------------------------

func TestMemoryPerformance(t *testing.T) {
	benchmark := NewPerformanceBenchmark()
	targets := benchmark.targets

	t.Run("SignalMemoryUsage", func(t *testing.T) {
		// Get baseline memory
		runtime.GC()
		var baseline runtime.MemStats
		runtime.ReadMemStats(&baseline)

		ResetReactiveContext()
		ResetScheduler()

		// Create signals
		signalCount := 1000
		signals := make([]func() int, signalCount)

		for i := 0; i < signalCount; i++ {
			getter, _ := CreateSignal(i)
			signals[i] = getter
		}

		FlushScheduler()

		// Force garbage collection and measure memory
		runtime.GC()
		var current runtime.MemStats
		runtime.ReadMemStats(&current)

		memoryUsed := current.Alloc - baseline.Alloc
		memoryPerSignal := memoryUsed / uint64(signalCount)

		t.Logf("Memory usage: %d bytes total, %d bytes per signal", memoryUsed, memoryPerSignal)

		if memoryPerSignal > targets.MemoryPerSignal {
			t.Errorf("Memory per signal %d bytes exceeds target %d bytes",
				memoryPerSignal, targets.MemoryPerSignal)
		}

		// Verify signals still work
		for i, getter := range signals {
			if getter() != i {
				t.Errorf("Signal %d returned wrong value", i)
			}
		}
	})

	t.Run("EffectMemoryUsage", func(t *testing.T) {
		// Get baseline memory
		runtime.GC()
		var baseline runtime.MemStats
		runtime.ReadMemStats(&baseline)

		ResetReactiveContext()
		ResetScheduler()

		// Create signals and effects
		effectCount := 500
		var executionCount int32

		getter, setter := CreateSignal(0)

		for i := 0; i < effectCount; i++ {
			CreateEffect(func() {
				_ = getter()
				atomic.AddInt32(&executionCount, 1)
			}, nil)
		}

		FlushScheduler()

		// Trigger effects
		setter(1)
		FlushScheduler()

		// Force garbage collection and measure memory
		runtime.GC()
		var current runtime.MemStats
		runtime.ReadMemStats(&current)

		memoryUsed := current.Alloc - baseline.Alloc
		memoryPerEffect := memoryUsed / uint64(effectCount)

		t.Logf("Effect memory usage: %d bytes total, %d bytes per effect", memoryUsed, memoryPerEffect)

		// Effects should use reasonable memory (allow 2x signal memory)
		maxMemoryPerEffect := targets.MemoryPerSignal * 2
		if memoryPerEffect > maxMemoryPerEffect {
			t.Errorf("Memory per effect %d bytes exceeds target %d bytes",
				memoryPerEffect, maxMemoryPerEffect)
		}

		// Verify effects executed
		expectedExecutions := int32(effectCount * 2) // Initial + update
		if atomic.LoadInt32(&executionCount) != expectedExecutions {
			t.Errorf("Expected %d effect executions, got %d",
				expectedExecutions, atomic.LoadInt32(&executionCount))
		}
	})
}

// ------------------------------------
// 🎯 Concurrency Performance Tests
// ------------------------------------

func TestConcurrencyPerformance(t *testing.T) {
	benchmark := NewPerformanceBenchmark()
	targets := benchmark.targets

	t.Run("ConcurrentEffectExecution", func(t *testing.T) {
		ResetReactiveContext()
		ResetScheduler()

		// Test concurrent effect limit
		getter, setter := CreateSignal(0)
		var activeEffects int32
		var maxActiveEffects int32
		var totalExecutions int32

		// Create many effects
		effectCount := targets.MaxConcurrentEffects

		for i := 0; i < effectCount; i++ {
			CreateEffect(func() {
				current := atomic.AddInt32(&activeEffects, 1)

				// Track maximum concurrent effects
				for {
					max := atomic.LoadInt32(&maxActiveEffects)
					if current <= max || atomic.CompareAndSwapInt32(&maxActiveEffects, max, current) {
						break
					}
				}

				_ = getter()
				atomic.AddInt32(&totalExecutions, 1)

				// Simulate some work
				time.Sleep(time.Microsecond)

				atomic.AddInt32(&activeEffects, -1)
			}, nil)
		}

		FlushScheduler()

		// Measure concurrent execution performance
		start := time.Now()

		// Trigger all effects
		setter(1)
		FlushScheduler()

		duration := time.Since(start)

		t.Logf("Concurrent effects: %d created, %d executed, max concurrent: %d, duration: %v",
			effectCount, atomic.LoadInt32(&totalExecutions),
			atomic.LoadInt32(&maxActiveEffects), duration)

		// Validate that we can handle the target number of concurrent effects
		if atomic.LoadInt32(&totalExecutions) != int32(effectCount) {
			t.Errorf("Expected %d effect executions, got %d",
				effectCount, atomic.LoadInt32(&totalExecutions))
		}

		// Performance should be reasonable even with many effects
		maxDuration := 100 * time.Millisecond
		if duration > maxDuration {
			t.Errorf("Concurrent effect execution took %v, expected under %v",
				duration, maxDuration)
		}

		result := PerformanceResult{
			Operation:       "ConcurrentEffects",
			Duration:        duration,
			OperationsCount: effectCount,
			Target:          maxDuration,
			Passed:          duration <= maxDuration,
			Improvement:     float64(maxDuration) / float64(duration),
		}

		benchmark.RecordResult(result)
	})

	benchmark.ValidateResults(t)
}

// ------------------------------------
// 🎯 Cascade Depth Performance Tests
// ------------------------------------

func TestCascadeDepthPerformance(t *testing.T) {
	targets := GetPerformanceTargets()

	t.Run("CascadeDepthLimit", func(t *testing.T) {
		ResetReactiveContext()
		ResetScheduler()

		// Create a chain of effects that could cause cascades
		var depth int32
		var maxDepth int32

		getter, setter := CreateSignal(0)

		// Create nested effects
		for i := 0; i < targets.MaxCascadeDepth+5; i++ {
			CreateEffect(func() {
				current := atomic.AddInt32(&depth, 1)

				// Track maximum depth
				for {
					max := atomic.LoadInt32(&maxDepth)
					if current <= max || atomic.CompareAndSwapInt32(&maxDepth, max, current) {
						break
					}
				}

				_ = getter()

				atomic.AddInt32(&depth, -1)
			}, nil)
		}

		FlushScheduler()

		// Trigger cascade
		setter(1)
		FlushScheduler()

		finalMaxDepth := atomic.LoadInt32(&maxDepth)
		t.Logf("Maximum cascade depth reached: %d (target limit: %d)",
			finalMaxDepth, targets.MaxCascadeDepth)

		// The system should handle reasonable cascade depths
		if finalMaxDepth > int32(targets.MaxCascadeDepth*2) {
			t.Errorf("Cascade depth %d exceeds reasonable limit %d",
				finalMaxDepth, targets.MaxCascadeDepth*2)
		}
	})
}

// ------------------------------------
// 🎯 Benchmark Tests
// ------------------------------------

func BenchmarkSignalUpdatesPerf(b *testing.B) {
	ResetReactiveContext()
	ResetScheduler()

	_, setter := CreateSignal(0)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		setter(i)
		FlushScheduler()
	}
}

func BenchmarkEffectExecutionPerf(b *testing.B) {
	ResetReactiveContext()
	ResetScheduler()

	getter, setter := CreateSignal(0)

	CreateEffect(func() {
		_ = getter()
	}, nil)

	FlushScheduler()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		setter(i)
		FlushScheduler()
	}
}

func BenchmarkMemoComputation(b *testing.B) {
	ResetReactiveContext()
	ResetScheduler()

	getter, setter := CreateSignal(0)

	memo := CreateMemo(func() int {
		return getter() * 2
	}, nil)

	FlushScheduler()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		setter(i)
		_ = memo()
		FlushScheduler()
	}
}

func BenchmarkDOMUpdates(b *testing.B) {
	ResetReactiveContext()
	ResetScheduler()

	document := js.Global().Get("document")
	element := document.Call("createElement", "div")
	document.Get("body").Call("appendChild", element)

	defer func() {
		document.Get("body").Call("removeChild", element)
	}()

	getter, setter := CreateSignal("initial")

	BindTextReactive(element, func() string {
		return getter()
	})

	FlushScheduler()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		setter(fmt.Sprintf("value-%d", i))
		FlushScheduler()
	}
}
