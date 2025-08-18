//go:build js && wasm

// memory_leak_test.go
// Memory leak detection and validation tests for reactive systems

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
// 🎯 Memory Leak Detection Framework
// ------------------------------------

// MemoryLeakDetector tracks memory allocations and detects leaks
type MemoryLeakDetector struct {
	baseline      runtime.MemStats
	samples       []runtime.MemStats
	thresholds    MemoryThresholds
	activeObjects map[string]int32
	mutex         sync.RWMutex
}

// MemoryThresholds defines acceptable memory usage limits
type MemoryThresholds struct {
	MaxLeakBytes        uint64        // Maximum acceptable memory leak
	MaxGrowthPercent    float64       // Maximum acceptable growth percentage
	SampleInterval      time.Duration // Interval between memory samples
	StabilizationPeriod time.Duration // Time to wait for memory to stabilize
}

// DefaultMemoryThresholds returns conservative memory leak thresholds
func DefaultMemoryThresholds() MemoryThresholds {
	return MemoryThresholds{
		MaxLeakBytes:        1024 * 1024, // 1MB
		MaxGrowthPercent:    10.0,        // 10%
		SampleInterval:      100 * time.Millisecond,
		StabilizationPeriod: 2 * time.Second,
	}
}

// NewMemoryLeakDetector creates a new memory leak detector
func NewMemoryLeakDetector() *MemoryLeakDetector {
	detector := &MemoryLeakDetector{
		samples:       make([]runtime.MemStats, 0),
		thresholds:    DefaultMemoryThresholds(),
		activeObjects: make(map[string]int32),
	}

	// Establish baseline
	runtime.GC()
	runtime.ReadMemStats(&detector.baseline)

	return detector
}

// TrackObject increments the count of active objects of a given type
func (mld *MemoryLeakDetector) TrackObject(objectType string) {
	mld.mutex.Lock()
	defer mld.mutex.Unlock()
	mld.activeObjects[objectType]++
}

// UntrackObject decrements the count of active objects of a given type
func (mld *MemoryLeakDetector) UntrackObject(objectType string) {
	mld.mutex.Lock()
	defer mld.mutex.Unlock()
	if count, exists := mld.activeObjects[objectType]; exists && count > 0 {
		mld.activeObjects[objectType]--
	}
}

// Sample takes a memory snapshot
func (mld *MemoryLeakDetector) Sample() {
	var stats runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&stats)

	mld.mutex.Lock()
	mld.samples = append(mld.samples, stats)
	mld.mutex.Unlock()
}

// StartMonitoring begins continuous memory monitoring
func (mld *MemoryLeakDetector) StartMonitoring() chan struct{} {
	stop := make(chan struct{})

	go func() {
		ticker := time.NewTicker(mld.thresholds.SampleInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				mld.Sample()
			case <-stop:
				return
			}
		}
	}()

	return stop
}

// DetectLeaks analyzes memory samples and detects potential leaks
func (mld *MemoryLeakDetector) DetectLeaks() MemoryLeakReport {
	mld.mutex.RLock()
	defer mld.mutex.RUnlock()

	report := MemoryLeakReport{
		BaselineAlloc:   mld.baseline.Alloc,
		ActiveObjects:   make(map[string]int32),
		LeaksDetected:   false,
		Recommendations: make([]string, 0),
	}

	// Copy active objects
	for objType, count := range mld.activeObjects {
		report.ActiveObjects[objType] = count
	}

	if len(mld.samples) == 0 {
		return report
	}

	// Get current memory stats
	current := mld.samples[len(mld.samples)-1]
	report.CurrentAlloc = current.Alloc
	report.MemoryGrowth = current.Alloc - mld.baseline.Alloc

	// Check for memory leaks
	if report.MemoryGrowth > mld.thresholds.MaxLeakBytes {
		report.LeaksDetected = true
		report.Recommendations = append(report.Recommendations,
			"Memory usage exceeds threshold - potential memory leak detected")
	}

	// Check growth percentage
	if mld.baseline.Alloc > 0 {
		growthPercent := float64(report.MemoryGrowth) / float64(mld.baseline.Alloc) * 100
		report.GrowthPercent = growthPercent

		if growthPercent > mld.thresholds.MaxGrowthPercent {
			report.LeaksDetected = true
			report.Recommendations = append(report.Recommendations,
				"Memory growth percentage exceeds threshold")
		}
	}

	// Check for orphaned objects
	for objType, count := range report.ActiveObjects {
		if count > 0 {
			report.LeaksDetected = true
			report.Recommendations = append(report.Recommendations,
				"Orphaned objects detected: "+objType)
		}
	}

	// Analyze memory trend
	if len(mld.samples) >= 3 {
		trend := mld.analyzeMemoryTrend()
		report.MemoryTrend = trend

		if trend == "increasing" {
			report.Recommendations = append(report.Recommendations,
				"Memory usage shows increasing trend - monitor for leaks")
		}
	}

	return report
}

// analyzeMemoryTrend analyzes the trend in memory usage
func (mld *MemoryLeakDetector) analyzeMemoryTrend() string {
	if len(mld.samples) < 3 {
		return "insufficient_data"
	}

	// Look at last 3 samples
	recent := mld.samples[len(mld.samples)-3:]

	if recent[2].Alloc > recent[1].Alloc && recent[1].Alloc > recent[0].Alloc {
		return "increasing"
	} else if recent[2].Alloc < recent[1].Alloc && recent[1].Alloc < recent[0].Alloc {
		return "decreasing"
	}

	return "stable"
}

// MemoryLeakReport contains the results of memory leak detection
type MemoryLeakReport struct {
	BaselineAlloc   uint64
	CurrentAlloc    uint64
	MemoryGrowth    uint64
	GrowthPercent   float64
	MemoryTrend     string
	ActiveObjects   map[string]int32
	LeaksDetected   bool
	Recommendations []string
}

// ------------------------------------
// 🎯 Signal Memory Leak Tests
// ------------------------------------

func TestSignalMemoryLeaks(t *testing.T) {
	detector := NewMemoryLeakDetector()
	stopMonitoring := detector.StartMonitoring()
	defer close(stopMonitoring)

	t.Run("SignalCreationAndCleanup", func(t *testing.T) {
		ResetReactiveContext()
		ResetScheduler()

		// Create and destroy signals in a loop
		iterations := 1000

		for i := 0; i < iterations; i++ {
			detector.TrackObject("signal")

			getter, setter := CreateSignal(i)

			// Use the signal
			setter(i * 2)
			_ = getter()

			// Signal should be eligible for GC when out of scope
			detector.UntrackObject("signal")
		}

		FlushScheduler()

		// Wait for memory to stabilize
		time.Sleep(detector.thresholds.StabilizationPeriod)
		detector.Sample()

		// Check for leaks
		report := detector.DetectLeaks()

		if report.LeaksDetected {
			t.Errorf("Memory leaks detected in signal creation/cleanup: %+v", report)
		}

		t.Logf("Signal memory test: %d bytes growth, %.2f%% increase",
			report.MemoryGrowth, report.GrowthPercent)
	})

	t.Run("SignalWithEffectsCleanup", func(t *testing.T) {
		ResetReactiveContext()
		ResetScheduler()

		iterations := 500

		for i := 0; i < iterations; i++ {
			detector.TrackObject("signal")
			detector.TrackObject("effect")

			getter, setter := CreateSignal(i)

			// Create effect that depends on signal
			CreateEffect(func() {
				_ = getter()
			}, nil)

			// Trigger effect
			setter(i * 2)
			FlushScheduler()

			detector.UntrackObject("signal")
			detector.UntrackObject("effect")
		}

		// Wait for memory to stabilize
		time.Sleep(detector.thresholds.StabilizationPeriod)
		detector.Sample()

		report := detector.DetectLeaks()

		if report.LeaksDetected {
			t.Errorf("Memory leaks detected in signal+effect cleanup: %+v", report)
		}

		t.Logf("Signal+Effect memory test: %d bytes growth, %.2f%% increase",
			report.MemoryGrowth, report.GrowthPercent)
	})
}

func TestEffectMemoryLeaks(t *testing.T) {
	detector := NewMemoryLeakDetector()
	stopMonitoring := detector.StartMonitoring()
	defer close(stopMonitoring)

	t.Run("EffectCleanupOnOwnerDisposal", func(t *testing.T) {
		ResetReactiveContext()
		ResetScheduler()

		iterations := 200

		for i := 0; i < iterations; i++ {
			detector.TrackObject("root")
			detector.TrackObject("effect")

			// Create root with effect
			_, cleanup := CreateRoot(func() interface{} {
				getter, setter := CreateSignal(i)

				CreateEffect(func() {
					_ = getter()
				}, nil)

				setter(i * 2)
				return nil
			})

			FlushScheduler()

			// Cleanup should dispose all effects
			cleanup()

			detector.UntrackObject("root")
			detector.UntrackObject("effect")
		}

		// Wait for memory to stabilize
		time.Sleep(detector.thresholds.StabilizationPeriod)
		detector.Sample()

		report := detector.DetectLeaks()

		if report.LeaksDetected {
			t.Errorf("Memory leaks detected in effect cleanup: %+v", report)
		}

		t.Logf("Effect cleanup memory test: %d bytes growth, %.2f%% increase",
			report.MemoryGrowth, report.GrowthPercent)
	})

	t.Run("NestedEffectCleanup", func(t *testing.T) {
		ResetReactiveContext()
		ResetScheduler()

		iterations := 100

		for i := 0; i < iterations; i++ {
			detector.TrackObject("root")

			_, cleanup := CreateRoot(func() interface{} {
				getter, setter := CreateSignal(i)

				// Create nested effects
				for j := 0; j < 5; j++ {
					detector.TrackObject("effect")
					CreateEffect(func() {
						_ = getter()
					}, nil)
				}

				setter(i * 2)
				return nil
			})

			FlushScheduler()
			cleanup()

			// All nested effects should be cleaned up
			for j := 0; j < 5; j++ {
				detector.UntrackObject("effect")
			}
			detector.UntrackObject("root")
		}

		time.Sleep(detector.thresholds.StabilizationPeriod)
		detector.Sample()

		report := detector.DetectLeaks()

		if report.LeaksDetected {
			t.Errorf("Memory leaks detected in nested effect cleanup: %+v", report)
		}

		t.Logf("Nested effect cleanup memory test: %d bytes growth, %.2f%% increase",
			report.MemoryGrowth, report.GrowthPercent)
	})
}

// ------------------------------------
// 🎯 DOM Memory Leak Tests
// ------------------------------------

func TestDOMMemoryLeaks(t *testing.T) {
	detector := NewMemoryLeakDetector()
	stopMonitoring := detector.StartMonitoring()
	defer close(stopMonitoring)

	// Create test container
	document := js.Global().Get("document")
	container := document.Call("createElement", "div")
	document.Get("body").Call("appendChild", container)

	defer func() {
		document.Get("body").Call("removeChild", container)
	}()

	t.Run("DOMBindingCleanup", func(t *testing.T) {
		ResetReactiveContext()
		ResetScheduler()

		iterations := 200

		for i := 0; i < iterations; i++ {
			detector.TrackObject("element")
			detector.TrackObject("binding")

			// Create element and binding
			element := document.Call("createElement", "div")
			container.Call("appendChild", element)

			getter, setter := CreateSignal("test")

			binding := BindTextReactive(element, func() string {
				return getter()
			})

			// Use the binding
			setter("updated")
			FlushScheduler()

			// Cleanup
			if binding != nil && binding.cleanup != nil {
				binding.cleanup()
			}
			container.Call("removeChild", element)

			detector.UntrackObject("element")
			detector.UntrackObject("binding")
		}

		time.Sleep(detector.thresholds.StabilizationPeriod)
		detector.Sample()

		report := detector.DetectLeaks()

		if report.LeaksDetected {
			t.Errorf("Memory leaks detected in DOM binding cleanup: %+v", report)
		}

		t.Logf("DOM binding cleanup memory test: %d bytes growth, %.2f%% increase",
			report.MemoryGrowth, report.GrowthPercent)
	})

	t.Run("EventSubscriptionCleanup", func(t *testing.T) {
		ResetReactiveContext()
		ResetScheduler()

		iterations := 100

		for i := 0; i < iterations; i++ {
			detector.TrackObject("element")
			detector.TrackObject("subscription")

			element := document.Call("createElement", "button")
			container.Call("appendChild", element)

			var clickCount int32
			cleanup := SubscribeReactive(element, "click", func(e js.Value) {
				atomic.AddInt32(&clickCount, 1)
			})

			// Trigger event
			event := document.Call("createEvent", "Event")
			event.Call("initEvent", "click", true, true)
			element.Call("dispatchEvent", event)

			// Cleanup subscription
			cleanup()
			container.Call("removeChild", element)

			detector.UntrackObject("element")
			detector.UntrackObject("subscription")
		}

		time.Sleep(detector.thresholds.StabilizationPeriod)
		detector.Sample()

		report := detector.DetectLeaks()

		if report.LeaksDetected {
			t.Errorf("Memory leaks detected in event subscription cleanup: %+v", report)
		}

		t.Logf("Event subscription cleanup memory test: %d bytes growth, %.2f%% increase",
			report.MemoryGrowth, report.GrowthPercent)
	})
}

// ------------------------------------
// 🎯 Component Memory Leak Tests
// ------------------------------------

func TestComponentMemoryLeaks(t *testing.T) {
	detector := NewMemoryLeakDetector()
	stopMonitoring := detector.StartMonitoring()
	defer close(stopMonitoring)

	t.Run("ComponentLifecycleCleanup", func(t *testing.T) {
		ResetReactiveContext()
		ResetScheduler()

		iterations := 100

		for i := 0; i < iterations; i++ {
			detector.TrackObject("component")

			var cleanupCalled bool

			_, cleanup := CreateRoot(func() interface{} {
				// Component state
				count, setCount := CreateSignal(0)

				// Component effects
				CreateEffect(func() {
					_ = count()
				}, nil)

				// Component cleanup
				OnCleanup(func() {
					cleanupCalled = true
				})

				// Simulate component updates
				setCount(i)

				return nil
			})

			FlushScheduler()

			// Cleanup component
			cleanup()

			if !cleanupCalled {
				t.Errorf("Component cleanup not called for iteration %d", i)
			}

			detector.UntrackObject("component")
		}

		time.Sleep(detector.thresholds.StabilizationPeriod)
		detector.Sample()

		report := detector.DetectLeaks()

		if report.LeaksDetected {
			t.Errorf("Memory leaks detected in component lifecycle cleanup: %+v", report)
		}

		t.Logf("Component lifecycle cleanup memory test: %d bytes growth, %.2f%% increase",
			report.MemoryGrowth, report.GrowthPercent)
	})

	t.Run("ComplexComponentCleanup", func(t *testing.T) {
		ResetReactiveContext()
		ResetScheduler()

		document := js.Global().Get("document")
		container := document.Call("createElement", "div")
		document.Get("body").Call("appendChild", container)

		defer func() {
			document.Get("body").Call("removeChild", container)
		}()

		iterations := 50

		for i := 0; i < iterations; i++ {
			detector.TrackObject("complex_component")

			_, cleanup := CreateRoot(func() interface{} {
				// Multiple signals
				name, setName := CreateSignal("test")
				count, setCount := CreateSignal(0)
				active, setActive := CreateSignal(true)

				// DOM elements
				element := document.Call("createElement", "div")
				container.Call("appendChild", element)

				// Multiple bindings
				BindTextReactive(element, func() string {
					return name()
				})

				BindAttributeReactive(element, "data-count", func() string {
					return fmt.Sprintf("%d", count())
				})

				BindClassReactive(element, "active", func() bool {
					return active()
				})

				// Event subscription
				eventCleanup := SubscribeReactive(element, "click", func(e js.Value) {
					setCount(count() + 1)
				})

				// Multiple effects
				CreateEffect(func() {
					_ = name()
					_ = count()
				}, nil)

				CreateEffect(func() {
					if active() {
						_ = count()
					}
				}, nil)

				// Cleanup registration
				OnCleanup(func() {
					eventCleanup()
					container.Call("removeChild", element)
				})

				// Simulate usage
				setName(fmt.Sprintf("test-%d", i))
				setCount(i)
				setActive(i%2 == 0)

				return element
			})

			FlushScheduler()
			cleanup()
			detector.UntrackObject("complex_component")
		}

		time.Sleep(detector.thresholds.StabilizationPeriod)
		detector.Sample()

		report := detector.DetectLeaks()

		if report.LeaksDetected {
			t.Errorf("Memory leaks detected in complex component cleanup: %+v", report)
		}

		t.Logf("Complex component cleanup memory test: %d bytes growth, %.2f%% increase",
			report.MemoryGrowth, report.GrowthPercent)
	})
}

// ------------------------------------
// 🎯 Stress Test Memory Leaks
// ------------------------------------

func TestStressMemoryLeaks(t *testing.T) {
	detector := NewMemoryLeakDetector()
	stopMonitoring := detector.StartMonitoring()
	defer close(stopMonitoring)

	t.Run("HighVolumeSignalOperations", func(t *testing.T) {
		ResetReactiveContext()
		ResetScheduler()

		// Create many signals and effects
		signalCount := 1000
		signals := make([]func(int), signalCount)

		for i := 0; i < signalCount; i++ {
			detector.TrackObject("stress_signal")

			getter, setter := CreateSignal(i)
			signals[i] = setter

			// Create effect for each signal
			detector.TrackObject("stress_effect")
			CreateEffect(func() {
				_ = getter()
			}, nil)
		}

		FlushScheduler()

		// Perform many updates
		updateRounds := 100
		for round := 0; round < updateRounds; round++ {
			for i, setter := range signals {
				setter(round*signalCount + i)
			}
			FlushScheduler()
		}

		// Cleanup tracking
		for i := 0; i < signalCount; i++ {
			detector.UntrackObject("stress_signal")
			detector.UntrackObject("stress_effect")
		}

		time.Sleep(detector.thresholds.StabilizationPeriod)
		detector.Sample()

		report := detector.DetectLeaks()

		// Allow higher thresholds for stress tests
		if report.MemoryGrowth > 10*1024*1024 { // 10MB threshold
			t.Errorf("Excessive memory growth in stress test: %d bytes", report.MemoryGrowth)
		}

		t.Logf("Stress test memory usage: %d bytes growth, %.2f%% increase",
			report.MemoryGrowth, report.GrowthPercent)
	})
}
