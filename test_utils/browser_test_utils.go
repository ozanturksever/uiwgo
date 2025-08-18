//go:build js && wasm

// browser_test_utils.go
// Comprehensive testing utilities for browser-based WASM testing with wasmbrowsertest

package test_utils

import (
	"fmt"
	"runtime"
	"syscall/js"
	"testing"
	"time"
)

// ------------------------------------
// 🎯 Browser Test Environment Setup
// ------------------------------------

// BrowserTestEnvironment provides a controlled testing environment for browser tests
type BrowserTestEnvironment struct {
	document    js.Value
	testRoot    js.Value
	cleanup     []func()
	startTime   time.Time
	memBaseline runtime.MemStats
}

// NewBrowserTestEnvironment creates a new browser test environment
func NewBrowserTestEnvironment(t *testing.T) *BrowserTestEnvironment {
	env := &BrowserTestEnvironment{
		document:  js.Global().Get("document"),
		cleanup:   make([]func(), 0),
		startTime: time.Now(),
	}

	// Get memory baseline
	runtime.GC()
	runtime.ReadMemStats(&env.memBaseline)

	// Create test root container
	env.testRoot = env.document.Call("createElement", "div")
	env.testRoot.Set("id", fmt.Sprintf("test-root-%d", time.Now().UnixNano()))
	env.testRoot.Get("style").Set("display", "none") // Hidden during tests
	env.document.Get("body").Call("appendChild", env.testRoot)

	// Register cleanup
	env.AddCleanup(func() {
		if env.testRoot.Truthy() {
			env.document.Get("body").Call("removeChild", env.testRoot)
		}
	})

	return env
}

// CreateTestElement creates a DOM element for testing
func (env *BrowserTestEnvironment) CreateTestElement(tag string) js.Value {
	element := env.document.Call("createElement", tag)
	env.testRoot.Call("appendChild", element)

	// Auto-cleanup
	env.AddCleanup(func() {
		if element.Truthy() && env.testRoot.Truthy() {
			env.testRoot.Call("removeChild", element)
		}
	})

	return element
}

// AddCleanup registers a cleanup function
func (env *BrowserTestEnvironment) AddCleanup(fn func()) {
	env.cleanup = append(env.cleanup, fn)
}

// Cleanup runs all registered cleanup functions
func (env *BrowserTestEnvironment) Cleanup() {
	for i := len(env.cleanup) - 1; i >= 0; i-- {
		env.cleanup[i]()
	}
	env.cleanup = nil
}

// GetMemoryUsage returns current memory usage compared to baseline
func (env *BrowserTestEnvironment) GetMemoryUsage() (allocated, freed uint64) {
	var current runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&current)

	allocated = current.Alloc - env.memBaseline.Alloc
	freed = current.Frees - env.memBaseline.Frees
	return
}

// GetTestDuration returns the duration since test start
func (env *BrowserTestEnvironment) GetTestDuration() time.Duration {
	return time.Since(env.startTime)
}

// ------------------------------------
// 🎯 DOM Testing Utilities
// ------------------------------------

// WaitForElement waits for an element to appear in the DOM
func WaitForElement(selector string, timeout time.Duration) (js.Value, error) {
	start := time.Now()
	for time.Since(start) < timeout {
		element := js.Global().Get("document").Call("querySelector", selector)
		if element.Truthy() {
			return element, nil
		}
		time.Sleep(10 * time.Millisecond)
	}
	return js.Null(), fmt.Errorf("element %s not found within timeout", selector)
}

// WaitForCondition waits for a condition to become true
func WaitForCondition(condition func() bool, timeout time.Duration) error {
	start := time.Now()
	for time.Since(start) < timeout {
		if condition() {
			return nil
		}
		time.Sleep(10 * time.Millisecond)
	}
	return fmt.Errorf("condition not met within timeout")
}

// TriggerEvent triggers a DOM event on an element
func TriggerEvent(element js.Value, eventType string) {
	event := js.Global().Get("document").Call("createEvent", "Event")
	event.Call("initEvent", eventType, true, true)
	element.Call("dispatchEvent", event)
}

// TriggerMouseEvent triggers a mouse event with coordinates
func TriggerMouseEvent(element js.Value, eventType string, x, y int) {
	event := js.Global().Get("document").Call("createEvent", "MouseEvent")
	event.Call("initMouseEvent", eventType, true, true, js.Global().Get("window"),
		0, x, y, x, y, false, false, false, false, 0, js.Null())
	element.Call("dispatchEvent", event)
}

// ------------------------------------
// 🎯 Performance Testing Utilities
// ------------------------------------

// PerformanceTimer measures execution time with high precision
type PerformanceTimer struct {
	start time.Time
	marks map[string]time.Time
}

// NewPerformanceTimer creates a new performance timer
func NewPerformanceTimer() *PerformanceTimer {
	return &PerformanceTimer{
		start: time.Now(),
		marks: make(map[string]time.Time),
	}
}

// Mark records a timestamp with a label
func (pt *PerformanceTimer) Mark(label string) {
	pt.marks[label] = time.Now()
}

// Measure returns the duration between two marks
func (pt *PerformanceTimer) Measure(startMark, endMark string) time.Duration {
	start, ok1 := pt.marks[startMark]
	end, ok2 := pt.marks[endMark]

	if !ok1 {
		start = pt.start
	}
	if !ok2 {
		end = time.Now()
	}

	return end.Sub(start)
}

// GetTotalDuration returns total duration since timer creation
func (pt *PerformanceTimer) GetTotalDuration() time.Duration {
	return time.Since(pt.start)
}

// ------------------------------------
// 🎯 Memory Testing Utilities
// ------------------------------------

// MemoryTracker tracks memory allocations during tests
type MemoryTracker struct {
	baseline runtime.MemStats
	samples  []runtime.MemStats
}

// NewMemoryTracker creates a new memory tracker
func NewMemoryTracker() *MemoryTracker {
	tracker := &MemoryTracker{
		samples: make([]runtime.MemStats, 0),
	}

	runtime.GC()
	runtime.ReadMemStats(&tracker.baseline)

	return tracker
}

// Sample takes a memory snapshot
func (mt *MemoryTracker) Sample() {
	var stats runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&stats)
	mt.samples = append(mt.samples, stats)
}

// GetPeakAllocation returns the peak memory allocation
func (mt *MemoryTracker) GetPeakAllocation() uint64 {
	peak := mt.baseline.Alloc
	for _, sample := range mt.samples {
		if sample.Alloc > peak {
			peak = sample.Alloc
		}
	}
	return peak - mt.baseline.Alloc
}

// GetCurrentAllocation returns current allocation vs baseline
func (mt *MemoryTracker) GetCurrentAllocation() uint64 {
	var current runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&current)
	return current.Alloc - mt.baseline.Alloc
}

// DetectLeaks checks for memory leaks by comparing allocations
func (mt *MemoryTracker) DetectLeaks(threshold uint64) bool {
	return mt.GetCurrentAllocation() > threshold
}

// ------------------------------------
// 🎯 Stress Testing Utilities
// ------------------------------------

// StressTestConfig configures stress test parameters
type StressTestConfig struct {
	Duration         time.Duration
	ConcurrentOps    int
	OperationsPerSec int
	MemoryThreshold  uint64
	TimeoutThreshold time.Duration
}

// DefaultStressConfig returns default stress test configuration
func DefaultStressConfig() StressTestConfig {
	return StressTestConfig{
		Duration:         30 * time.Second,
		ConcurrentOps:    100,
		OperationsPerSec: 1000,
		MemoryThreshold:  50 * 1024 * 1024, // 50MB
		TimeoutThreshold: 100 * time.Millisecond,
	}
}

// StressTestResult contains stress test results
type StressTestResult struct {
	TotalOperations    int
	SuccessfulOps      int
	FailedOps          int
	AverageLatency     time.Duration
	PeakLatency        time.Duration
	PeakMemoryUsage    uint64
	MemoryLeakDetected bool
	Errors             []error
}

// ------------------------------------
// 🎯 Assertion Utilities
// ------------------------------------

// AssertElementExists checks if an element exists in the DOM
func AssertElementExists(t *testing.T, selector string) js.Value {
	element := js.Global().Get("document").Call("querySelector", selector)
	if !element.Truthy() {
		t.Fatalf("Element %s does not exist", selector)
	}
	return element
}

// AssertElementText checks if an element has expected text content
func AssertElementText(t *testing.T, element js.Value, expected string) {
	actual := element.Get("textContent").String()
	if actual != expected {
		t.Errorf("Expected text %q, got %q", expected, actual)
	}
}

// AssertElementAttribute checks if an element has expected attribute value
func AssertElementAttribute(t *testing.T, element js.Value, attr, expected string) {
	actual := element.Call("getAttribute", attr)
	if !actual.Truthy() && expected != "" {
		t.Errorf("Expected attribute %s to be %q, but attribute is missing", attr, expected)
		return
	}

	actualStr := actual.String()
	if actualStr != expected {
		t.Errorf("Expected attribute %s to be %q, got %q", attr, expected, actualStr)
	}
}

// AssertPerformance checks if operation meets performance requirements
func AssertPerformance(t *testing.T, duration time.Duration, threshold time.Duration, operation string) {
	if duration > threshold {
		t.Errorf("%s took %v, expected under %v", operation, duration, threshold)
	}
}

// AssertMemoryUsage checks if memory usage is within limits
func AssertMemoryUsage(t *testing.T, usage uint64, limit uint64, operation string) {
	if usage > limit {
		t.Errorf("%s used %d bytes, expected under %d bytes", operation, usage, limit)
	}
}

// AssertNoMemoryLeaks checks for memory leaks using the tracker
func AssertNoMemoryLeaks(t *testing.T, tracker *MemoryTracker, threshold uint64) {
	if tracker.DetectLeaks(threshold) {
		current := tracker.GetCurrentAllocation()
		t.Errorf("Memory leak detected: %d bytes allocated, threshold %d bytes", current, threshold)
	}
}
