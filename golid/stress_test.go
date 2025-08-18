//go:build js && wasm

// stress_test.go
// Stress testing scenarios and load tests for reactive systems

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
// 🎯 Stress Test Configuration
// ------------------------------------

// StressTestConfig defines parameters for stress testing
type StressTestConfig struct {
	Duration            time.Duration
	ConcurrentOps       int
	OperationsPerSecond int
	MemoryThreshold     uint64
	TimeoutThreshold    time.Duration
	MaxCascadeDepth     int
}

// DefaultStressConfig returns default stress test configuration
func DefaultStressConfig() StressTestConfig {
	return StressTestConfig{
		Duration:            30 * time.Second,
		ConcurrentOps:       1000,
		OperationsPerSecond: 10000,
		MemoryThreshold:     100 * 1024 * 1024, // 100MB
		TimeoutThreshold:    1 * time.Second,
		MaxCascadeDepth:     50,
	}
}

// StressTestResult contains stress test execution results
type StressTestResult struct {
	TotalOperations    int64
	SuccessfulOps      int64
	FailedOps          int64
	TimeoutOps         int64
	AverageLatency     time.Duration
	PeakLatency        time.Duration
	MinLatency         time.Duration
	PeakMemoryUsage    uint64
	MemoryLeakDetected bool
	SystemStable       bool
	Errors             []error
	PerformanceMetrics map[string]interface{}
}

// StressTestRunner manages stress test execution
type StressTestRunner struct {
	config  StressTestConfig
	results StressTestResult
	mutex   sync.RWMutex
	running int32
}

// NewStressTestRunner creates a new stress test runner
func NewStressTestRunner(config StressTestConfig) *StressTestRunner {
	return &StressTestRunner{
		config: config,
		results: StressTestResult{
			Errors:             make([]error, 0),
			PerformanceMetrics: make(map[string]interface{}),
		},
	}
}

// RecordOperation records the result of a stress test operation
func (str *StressTestRunner) RecordOperation(success bool, latency time.Duration, err error) {
	str.mutex.Lock()
	defer str.mutex.Unlock()

	str.results.TotalOperations++

	if success {
		str.results.SuccessfulOps++
	} else {
		str.results.FailedOps++
		if err != nil {
			str.results.Errors = append(str.results.Errors, err)
		}
	}

	// Update latency statistics
	if str.results.TotalOperations == 1 {
		str.results.AverageLatency = latency
		str.results.PeakLatency = latency
		str.results.MinLatency = latency
	} else {
		// Update average (simple moving average)
		str.results.AverageLatency = time.Duration(
			(int64(str.results.AverageLatency)*str.results.TotalOperations + int64(latency)) /
				(str.results.TotalOperations + 1))

		if latency > str.results.PeakLatency {
			str.results.PeakLatency = latency
		}
		if latency < str.results.MinLatency {
			str.results.MinLatency = latency
		}
	}

	// Check for timeouts
	if latency > str.config.TimeoutThreshold {
		str.results.TimeoutOps++
	}
}

// GetResults returns the current stress test results
func (str *StressTestRunner) GetResults() StressTestResult {
	str.mutex.RLock()
	defer str.mutex.RUnlock()
	return str.results
}

// ------------------------------------
// 🎯 Signal Stress Tests
// ------------------------------------

func TestSignalStress(t *testing.T) {
	config := DefaultStressConfig()
	config.Duration = 10 * time.Second // Shorter for unit tests
	runner := NewStressTestRunner(config)

	ResetReactiveContext()
	ResetScheduler()

	t.Run("MassiveSignalCreation", func(t *testing.T) {
		// Test creating and using thousands of signals
		signalCount := 5000
		signals := make([]func() int, signalCount)
		setters := make([]func(int), signalCount)

		start := time.Now()

		// Create signals
		for i := 0; i < signalCount; i++ {
			opStart := time.Now()
			getter, setter := CreateSignal(i)
			signals[i] = getter
			setters[i] = setter

			latency := time.Since(opStart)
			runner.RecordOperation(true, latency, nil)
		}

		FlushScheduler()

		// Test signal access performance
		for i := 0; i < signalCount; i++ {
			opStart := time.Now()
			value := signals[i]()
			latency := time.Since(opStart)

			success := value == i
			runner.RecordOperation(success, latency, nil)
		}

		// Test signal updates
		for i := 0; i < signalCount; i++ {
			opStart := time.Now()
			setters[i](i * 2)
			latency := time.Since(opStart)
			runner.RecordOperation(true, latency, nil)
		}

		FlushScheduler()

		duration := time.Since(start)
		t.Logf("Massive signal creation test completed in %v", duration)

		// Validate results
		results := runner.GetResults()
		if results.FailedOps > 0 {
			t.Errorf("Signal stress test had %d failed operations", results.FailedOps)
		}

		if results.AverageLatency > 100*time.Microsecond {
			t.Errorf("Average latency %v exceeds threshold", results.AverageLatency)
		}
	})

	t.Run("ConcurrentSignalUpdates", func(t *testing.T) {
		signalCount := 100
		signals := make([]func() int, signalCount)
		setters := make([]func(int), signalCount)

		// Create signals
		for i := 0; i < signalCount; i++ {
			getter, setter := CreateSignal(i)
			signals[i] = getter
			setters[i] = setter
		}

		FlushScheduler()

		// Concurrent updates
		var wg sync.WaitGroup
		goroutineCount := 50
		updatesPerGoroutine := 100

		for g := 0; g < goroutineCount; g++ {
			wg.Add(1)
			go func(goroutineID int) {
				defer wg.Done()

				for u := 0; u < updatesPerGoroutine; u++ {
					signalIndex := (goroutineID*updatesPerGoroutine + u) % signalCount
					newValue := goroutineID*1000 + u

					opStart := time.Now()
					setters[signalIndex](newValue)
					latency := time.Since(opStart)

					runner.RecordOperation(true, latency, nil)
				}
			}(g)
		}

		wg.Wait()
		FlushScheduler()

		// Verify final state
		for i := 0; i < signalCount; i++ {
			_ = signals[i]() // Just ensure no panics
		}

		results := runner.GetResults()
		expectedOps := int64(goroutineCount * updatesPerGoroutine)
		if results.TotalOperations < expectedOps {
			t.Errorf("Expected at least %d operations, got %d", expectedOps, results.TotalOperations)
		}
	})
}

func TestEffectStress(t *testing.T) {
	config := DefaultStressConfig()
	config.Duration = 10 * time.Second
	runner := NewStressTestRunner(config)

	ResetReactiveContext()
	ResetScheduler()

	t.Run("MassiveEffectCreation", func(t *testing.T) {
		effectCount := 2000
		var executionCount int64

		getter, setter := CreateSignal(0)

		// Create many effects
		for i := 0; i < effectCount; i++ {
			opStart := time.Now()

			CreateEffect(func() {
				_ = getter()
				atomic.AddInt64(&executionCount, 1)
			}, nil)

			latency := time.Since(opStart)
			runner.RecordOperation(true, latency, nil)
		}

		FlushScheduler()

		initialExecutions := atomic.LoadInt64(&executionCount)
		if initialExecutions != int64(effectCount) {
			t.Errorf("Expected %d initial effect executions, got %d", effectCount, initialExecutions)
		}

		// Trigger all effects
		opStart := time.Now()
		setter(1)
		FlushScheduler()
		latency := time.Since(opStart)

		finalExecutions := atomic.LoadInt64(&executionCount)
		expectedTotal := int64(effectCount * 2) // Initial + triggered

		if finalExecutions != expectedTotal {
			t.Errorf("Expected %d total effect executions, got %d", expectedTotal, finalExecutions)
		}

		runner.RecordOperation(true, latency, nil)

		t.Logf("Massive effect creation: %d effects, trigger latency: %v", effectCount, latency)
	})

	t.Run("EffectCascadeStress", func(t *testing.T) {
		cascadeDepth := 20
		var cascadeCount int64

		// Create cascade chain
		signals := make([]func() int, cascadeDepth)
		setters := make([]func(int), cascadeDepth)

		for i := 0; i < cascadeDepth; i++ {
			getter, setter := CreateSignal(i)
			signals[i] = getter
			setters[i] = setter

			if i > 0 {
				// Create effect that depends on previous signal and updates current
				prevIndex := i - 1
				currentIndex := i
				CreateEffect(func() {
					prevValue := signals[prevIndex]()
					setters[currentIndex](prevValue + 1)
					atomic.AddInt64(&cascadeCount, 1)
				}, nil)
			}
		}

		FlushScheduler()

		// Trigger cascade
		opStart := time.Now()
		setters[0](100)
		FlushScheduler()
		latency := time.Since(opStart)

		runner.RecordOperation(true, latency, nil)

		// Verify cascade propagated
		finalValue := signals[cascadeDepth-1]()
		expectedValue := 100 + cascadeDepth - 1

		if finalValue != expectedValue {
			t.Errorf("Cascade failed: expected final value %d, got %d", expectedValue, finalValue)
		}

		t.Logf("Effect cascade stress: depth %d, latency: %v, cascades: %d",
			cascadeDepth, latency, atomic.LoadInt64(&cascadeCount))
	})
}

// ------------------------------------
// 🎯 DOM Stress Tests
// ------------------------------------

func TestDOMStress(t *testing.T) {
	config := DefaultStressConfig()
	runner := NewStressTestRunner(config)

	ResetReactiveContext()
	ResetScheduler()

	// Create test container
	document := js.Global().Get("document")
	container := document.Call("createElement", "div")
	document.Get("body").Call("appendChild", container)

	defer func() {
		document.Get("body").Call("removeChild", container)
	}()

	t.Run("MassiveDOMUpdates", func(t *testing.T) {
		elementCount := 500
		elements := make([]js.Value, elementCount)
		signals := make([]func() string, elementCount)
		setters := make([]func(string), elementCount)

		// Create elements and bind them
		for i := 0; i < elementCount; i++ {
			opStart := time.Now()

			element := document.Call("createElement", "div")
			container.Call("appendChild", element)
			elements[i] = element

			getter, setter := CreateSignal(fmt.Sprintf("initial-%d", i))
			signals[i] = getter
			setters[i] = setter

			BindTextReactive(element, func() string {
				return getter()
			})

			latency := time.Since(opStart)
			runner.RecordOperation(true, latency, nil)
		}

		FlushScheduler()

		// Perform rapid updates
		updateRounds := 50
		for round := 0; round < updateRounds; round++ {
			opStart := time.Now()

			// Update all elements
			for i := 0; i < elementCount; i++ {
				setters[i](fmt.Sprintf("round-%d-element-%d", round, i))
			}

			FlushScheduler()
			latency := time.Since(opStart)
			runner.RecordOperation(true, latency, nil)
		}

		// Verify final state
		for i := 0; i < elementCount; i++ {
			expectedText := fmt.Sprintf("round-%d-element-%d", updateRounds-1, i)
			actualText := elements[i].Get("textContent").String()

			if actualText != expectedText {
				t.Errorf("Element %d text mismatch: expected %s, got %s", i, expectedText, actualText)
			}
		}

		// Cleanup
		for _, element := range elements {
			container.Call("removeChild", element)
		}

		results := runner.GetResults()
		t.Logf("Massive DOM updates: %d operations, avg latency: %v",
			results.TotalOperations, results.AverageLatency)
	})

	t.Run("ComplexDOMBindings", func(t *testing.T) {
		elementCount := 200
		elements := make([]js.Value, elementCount)

		for i := 0; i < elementCount; i++ {
			opStart := time.Now()

			element := document.Call("createElement", "div")
			container.Call("appendChild", element)
			elements[i] = element

			// Multiple signal bindings per element
			text, setText := CreateSignal(fmt.Sprintf("text-%d", i))
			active, setActive := CreateSignal(i%2 == 0)
			count, setCount := CreateSignal(i)

			// Multiple bindings
			BindTextReactive(element, func() string {
				return text()
			})

			BindAttributeReactive(element, "data-count", func() string {
				return fmt.Sprintf("%d", count())
			})

			BindClassReactive(element, "active", func() bool {
				return active()
			})

			BindStyleReactive(element, "opacity", func() string {
				if active() {
					return "1.0"
				}
				return "0.5"
			})

			// Event binding
			SubscribeReactive(element, "click", func(e js.Value) {
				setCount(count() + 1)
				setActive(!active())
			})

			// Update signals
			setText(fmt.Sprintf("updated-text-%d", i))
			setActive(!active())
			setCount(i * 2)

			latency := time.Since(opStart)
			runner.RecordOperation(true, latency, nil)
		}

		FlushScheduler()

		// Cleanup
		for _, element := range elements {
			container.Call("removeChild", element)
		}

		results := runner.GetResults()
		t.Logf("Complex DOM bindings: %d elements, avg latency: %v",
			elementCount, results.AverageLatency)
	})
}

// ------------------------------------
// 🎯 Memory Stress Tests
// ------------------------------------

func TestMemoryStress(t *testing.T) {
	config := DefaultStressConfig()
	runner := NewStressTestRunner(config)

	t.Run("MemoryPressureTest", func(t *testing.T) {
		// Get baseline memory
		runtime.GC()
		var baseline runtime.MemStats
		runtime.ReadMemStats(&baseline)

		ResetReactiveContext()
		ResetScheduler()

		// Create memory pressure
		iterations := 1000

		for iter := 0; iter < iterations; iter++ {
			opStart := time.Now()

			// Create temporary reactive structures
			_, cleanup := CreateRoot(func() interface{} {
				// Multiple signals
				signals := make([]func() int, 10)
				setters := make([]func(int), 10)

				for i := 0; i < 10; i++ {
					getter, setter := CreateSignal(iter*10 + i)
					signals[i] = getter
					setters[i] = setter

					// Effects for each signal
					CreateEffect(func() {
						_ = getter()
					}, nil)
				}

				// Memos
				for i := 0; i < 5; i++ {
					CreateMemo(func() int {
						sum := 0
						for _, getter := range signals {
							sum += getter()
						}
						return sum
					}, nil)
				}

				// Update signals
				for i, setter := range setters {
					setter(iter*100 + i)
				}

				return nil
			})

			FlushScheduler()

			// Immediate cleanup
			cleanup()

			latency := time.Since(opStart)
			runner.RecordOperation(true, latency, nil)

			// Check memory every 100 iterations
			if iter%100 == 0 {
				runtime.GC()
				var current runtime.MemStats
				runtime.ReadMemStats(&current)

				memoryGrowth := current.Alloc - baseline.Alloc
				if memoryGrowth > config.MemoryThreshold {
					t.Errorf("Memory growth %d bytes exceeds threshold %d bytes at iteration %d",
						memoryGrowth, config.MemoryThreshold, iter)
					break
				}
			}
		}

		// Final memory check
		runtime.GC()
		var final runtime.MemStats
		runtime.ReadMemStats(&final)

		finalGrowth := final.Alloc - baseline.Alloc
		t.Logf("Memory stress test: %d iterations, final growth: %d bytes",
			iterations, finalGrowth)

		results := runner.GetResults()
		if results.FailedOps > 0 {
			t.Errorf("Memory stress test had %d failed operations", results.FailedOps)
		}
	})
}

// ------------------------------------
// 🎯 System Stability Tests
// ------------------------------------

func TestSystemStability(t *testing.T) {
	config := DefaultStressConfig()
	config.Duration = 15 * time.Second
	runner := NewStressTestRunner(config)

	t.Run("LongRunningStabilityTest", func(t *testing.T) {
		ResetReactiveContext()
		ResetScheduler()

		// Create persistent reactive system
		mainSignal, setMainSignal := CreateSignal(0)
		var operationCount int64

		// Create multiple effect chains
		for chain := 0; chain < 10; chain++ {
			CreateEffect(func() {
				value := mainSignal()
				atomic.AddInt64(&operationCount, 1)

				// Simulate work
				for i := 0; i < 100; i++ {
					_ = value * i
				}
			}, nil)
		}

		FlushScheduler()

		// Run for duration
		start := time.Now()
		updateCount := 0

		for time.Since(start) < config.Duration {
			opStart := time.Now()

			setMainSignal(updateCount)
			FlushScheduler()

			latency := time.Since(opStart)
			runner.RecordOperation(true, latency, nil)

			updateCount++

			// Brief pause to prevent overwhelming
			time.Sleep(10 * time.Millisecond)
		}

		duration := time.Since(start)
		finalOperationCount := atomic.LoadInt64(&operationCount)

		t.Logf("Stability test: ran for %v, %d updates, %d effect operations",
			duration, updateCount, finalOperationCount)

		results := runner.GetResults()

		// System should remain stable
		if results.FailedOps > 0 {
			t.Errorf("System stability test had %d failed operations", results.FailedOps)
		}

		// Performance should not degrade significantly
		if results.PeakLatency > 100*time.Millisecond {
			t.Errorf("Peak latency %v indicates performance degradation", results.PeakLatency)
		}

		// Should handle reasonable operation rate
		opsPerSecond := float64(results.TotalOperations) / duration.Seconds()
		if opsPerSecond < 10 { // Very conservative threshold
			t.Errorf("Operation rate %.2f ops/sec too low, system may be unstable", opsPerSecond)
		}
	})

	t.Run("ErrorRecoveryTest", func(t *testing.T) {
		ResetReactiveContext()
		ResetScheduler()

		getter, setter := CreateSignal(0)
		var successfulEffects int64
		var errorEffects int64

		// Create effects that may error
		for i := 0; i < 100; i++ {
			CreateEffect(func() {
				value := getter()

				// Simulate occasional errors
				if value%10 == 7 { // Error on specific values
					atomic.AddInt64(&errorEffects, 1)
					panic("simulated effect error")
				}

				atomic.AddInt64(&successfulEffects, 1)
			}, nil)
		}

		FlushScheduler()

		// Trigger values that will cause errors
		for i := 0; i < 50; i++ {
			opStart := time.Now()

			func() {
				defer func() {
					if r := recover(); r != nil {
						// Expected for error cases
					}
				}()

				setter(i)
				FlushScheduler()
			}()

			latency := time.Since(opStart)
			runner.RecordOperation(true, latency, nil)
		}

		t.Logf("Error recovery test: %d successful effects, %d error effects",
			atomic.LoadInt64(&successfulEffects), atomic.LoadInt64(&errorEffects))

		// System should continue functioning despite errors
		if atomic.LoadInt64(&successfulEffects) == 0 {
			t.Error("No successful effects executed - system may have failed")
		}
	})
}

// ------------------------------------
// 🎯 Comprehensive Stress Test
// ------------------------------------

func TestComprehensiveStress(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping comprehensive stress test in short mode")
	}

	config := DefaultStressConfig()
	config.Duration = 30 * time.Second
	runner := NewStressTestRunner(config)

	t.Run("FullSystemStress", func(t *testing.T) {
		ResetReactiveContext()
		ResetScheduler()

		// Create test environment
		document := js.Global().Get("document")
		container := document.Call("createElement", "div")
		document.Get("body").Call("appendChild", container)

		defer func() {
			document.Get("body").Call("removeChild", container)
		}()

		// Create complex reactive application
		var totalOperations int64
		start := time.Now()

		for time.Since(start) < config.Duration {
			opStart := time.Now()

			// Create mini-application
			_, cleanup := CreateRoot(func() interface{} {
				// Application state
				count, setCount := CreateSignal(0)
				name, setName := CreateSignal("test")
				items, setItems := CreateSignal([]string{})

				// DOM elements
				element := document.Call("createElement", "div")
				container.Call("appendChild", element)

				// Reactive bindings
				BindTextReactive(element, func() string {
					return fmt.Sprintf("%s: %d", name(), count())
				})

				// Effects
				CreateEffect(func() {
					currentItems := items()
					if len(currentItems) > 10 {
						setItems(currentItems[:5]) // Trim list
					}
				}, nil)

				// Simulate user interactions
				for i := 0; i < 10; i++ {
					setCount(count() + 1)
					setName(fmt.Sprintf("test-%d", i))

					currentItems := items()
					newItems := append(currentItems, fmt.Sprintf("item-%d", i))
					setItems(newItems)
				}

				// Cleanup registration
				OnCleanup(func() {
					container.Call("removeChild", element)
				})

				return element
			})

			FlushScheduler()
			cleanup()

			latency := time.Since(opStart)
			runner.RecordOperation(true, latency, nil)

			atomic.AddInt64(&totalOperations, 1)

			// Brief pause
			time.Sleep(50 * time.Millisecond)
		}

		duration := time.Since(start)
		results := runner.GetResults()

		t.Logf("Comprehensive stress test completed:")
		t.Logf("  Duration: %v", duration)
		t.Logf("  Total operations: %d", results.TotalOperations)
		t.Logf("  Successful: %d", results.SuccessfulOps)
		t.Logf("  Failed: %d", results.FailedOps)
		t.Logf("  Average latency: %v", results.AverageLatency)
		t.Logf("  Peak latency: %v", results.PeakLatency)

		// Validate system remained stable
		if results.FailedOps > results.TotalOperations/10 { // Allow 10% failure rate
			t.Errorf("Too many failed operations: %d/%d", results.FailedOps, results.TotalOperations)
		}

		if results.AverageLatency > 500*time.Millisecond {
			t.Errorf("Average latency too high: %v", results.AverageLatency)
		}

		// Check memory usage
		runtime.GC()
		var memStats runtime.MemStats
		runtime.ReadMemStats(&memStats)

		if memStats.Alloc > config.MemoryThreshold {
			t.Errorf("Memory usage %d bytes exceeds threshold %d bytes",
				memStats.Alloc, config.MemoryThreshold)
		}
	})
}
