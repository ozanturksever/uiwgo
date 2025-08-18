//go:build js && wasm

// reactive_test_utils.go
// Testing utilities specifically for reactive systems and signal testing

package test_utils

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// ------------------------------------
// 🎯 Reactive System Test Utilities
// ------------------------------------

// ReactiveTestContext provides isolated testing environment for reactive systems
type ReactiveTestContext struct {
	signalCount   int32
	effectCount   int32
	computeCount  int32
	cleanupCount  int32
	errors        []error
	mutex         sync.RWMutex
	startTime     time.Time
	memoryTracker *MemoryTracker
	perfTimer     *PerformanceTimer
}

// NewReactiveTestContext creates a new reactive test context
func NewReactiveTestContext() *ReactiveTestContext {
	return &ReactiveTestContext{
		errors:        make([]error, 0),
		startTime:     time.Now(),
		memoryTracker: NewMemoryTracker(),
		perfTimer:     NewPerformanceTimer(),
	}
}

// TrackSignalCreation increments signal creation counter
func (ctx *ReactiveTestContext) TrackSignalCreation() {
	atomic.AddInt32(&ctx.signalCount, 1)
}

// TrackEffectExecution increments effect execution counter
func (ctx *ReactiveTestContext) TrackEffectExecution() {
	atomic.AddInt32(&ctx.effectCount, 1)
}

// TrackComputation increments computation counter
func (ctx *ReactiveTestContext) TrackComputation() {
	atomic.AddInt32(&ctx.computeCount, 1)
}

// TrackCleanup increments cleanup counter
func (ctx *ReactiveTestContext) TrackCleanup() {
	atomic.AddInt32(&ctx.cleanupCount, 1)
}

// AddError records an error during testing
func (ctx *ReactiveTestContext) AddError(err error) {
	ctx.mutex.Lock()
	defer ctx.mutex.Unlock()
	ctx.errors = append(ctx.errors, err)
}

// GetStats returns current test statistics
func (ctx *ReactiveTestContext) GetStats() ReactiveTestStats {
	ctx.mutex.RLock()
	defer ctx.mutex.RUnlock()

	return ReactiveTestStats{
		SignalCount:     atomic.LoadInt32(&ctx.signalCount),
		EffectCount:     atomic.LoadInt32(&ctx.effectCount),
		ComputeCount:    atomic.LoadInt32(&ctx.computeCount),
		CleanupCount:    atomic.LoadInt32(&ctx.cleanupCount),
		ErrorCount:      int32(len(ctx.errors)),
		Duration:        time.Since(ctx.startTime),
		MemoryUsage:     ctx.memoryTracker.GetCurrentAllocation(),
		PeakMemoryUsage: ctx.memoryTracker.GetPeakAllocation(),
	}
}

// ReactiveTestStats contains test execution statistics
type ReactiveTestStats struct {
	SignalCount     int32
	EffectCount     int32
	ComputeCount    int32
	CleanupCount    int32
	ErrorCount      int32
	Duration        time.Duration
	MemoryUsage     uint64
	PeakMemoryUsage uint64
}

// ------------------------------------
// 🎯 Signal Testing Utilities
// ------------------------------------

// SignalTestCase represents a test case for signal behavior
type SignalTestCase struct {
	Name            string
	InitialValue    interface{}
	Updates         []interface{}
	ExpectedValues  []interface{}
	ExpectedEffects int
	Timeout         time.Duration
}

// EffectTestCase represents a test case for effect behavior
type EffectTestCase struct {
	Name               string
	Setup              func() (interface{}, func())
	TriggerUpdates     func()
	ExpectedExecutions int
	MaxExecutionTime   time.Duration
	Cleanup            func()
}

// MemoTestCase represents a test case for memo behavior
type MemoTestCase struct {
	Name                 string
	Dependencies         []interface{}
	ComputeFunction      func() interface{}
	Updates              []func()
	ExpectedComputations int
	ExpectedValue        interface{}
}

// ------------------------------------
// 🎯 Performance Testing Utilities
// ------------------------------------

// PerformanceTestConfig configures performance test parameters
type PerformanceTestConfig struct {
	SignalUpdateTarget    time.Duration // Target: 5μs
	DOMUpdateTarget       time.Duration // Target: 10ms
	MemoryPerSignalTarget uint64        // Target: 200B
	MaxConcurrentEffects  int           // Target: 10,000
	MaxCascadeDepth       int           // Target: 10
}

// DefaultPerformanceTargets returns the performance targets from the roadmap
func DefaultPerformanceTargets() PerformanceTestConfig {
	return PerformanceTestConfig{
		SignalUpdateTarget:    5 * time.Microsecond,
		DOMUpdateTarget:       10 * time.Millisecond,
		MemoryPerSignalTarget: 200, // 200 bytes
		MaxConcurrentEffects:  10000,
		MaxCascadeDepth:       10,
	}
}

// PerformanceBenchmark measures performance of reactive operations
type PerformanceBenchmark struct {
	config  PerformanceTestConfig
	results map[string]time.Duration
	memory  map[string]uint64
	mutex   sync.RWMutex
}

// NewPerformanceBenchmark creates a new performance benchmark
func NewPerformanceBenchmark(config PerformanceTestConfig) *PerformanceBenchmark {
	return &PerformanceBenchmark{
		config:  config,
		results: make(map[string]time.Duration),
		memory:  make(map[string]uint64),
	}
}

// MeasureSignalUpdate measures signal update performance
func (pb *PerformanceBenchmark) MeasureSignalUpdate(name string, operation func()) {
	start := time.Now()
	operation()
	duration := time.Since(start)

	pb.mutex.Lock()
	pb.results[name] = duration
	pb.mutex.Unlock()
}

// MeasureMemoryUsage measures memory usage of an operation
func (pb *PerformanceBenchmark) MeasureMemoryUsage(name string, operation func()) {
	tracker := NewMemoryTracker()
	operation()
	tracker.Sample()
	usage := tracker.GetCurrentAllocation()

	pb.mutex.Lock()
	pb.memory[name] = usage
	pb.mutex.Unlock()
}

// ValidatePerformance checks if results meet targets
func (pb *PerformanceBenchmark) ValidatePerformance(t *testing.T) {
	pb.mutex.RLock()
	defer pb.mutex.RUnlock()

	for name, duration := range pb.results {
		if duration > pb.config.SignalUpdateTarget {
			t.Errorf("Performance target missed for %s: %v > %v",
				name, duration, pb.config.SignalUpdateTarget)
		}
	}

	for name, memory := range pb.memory {
		if memory > pb.config.MemoryPerSignalTarget {
			t.Errorf("Memory target missed for %s: %d bytes > %d bytes",
				name, memory, pb.config.MemoryPerSignalTarget)
		}
	}
}

// ------------------------------------
// 🎯 Concurrency Testing Utilities
// ------------------------------------

// ConcurrencyTestConfig configures concurrent testing parameters
type ConcurrencyTestConfig struct {
	NumGoroutines          int
	OperationsPerGoroutine int
	TestDuration           time.Duration
	ExpectedRaceConditions bool
}

// ConcurrencyTestResult contains results of concurrency testing
type ConcurrencyTestResult struct {
	TotalOperations int
	SuccessfulOps   int
	RaceConditions  int
	Deadlocks       int
	PanicRecoveries int
	AverageLatency  time.Duration
	MaxLatency      time.Duration
}

// RunConcurrencyTest executes a concurrent test scenario
func RunConcurrencyTest(config ConcurrencyTestConfig, operation func(int) error) *ConcurrencyTestResult {
	result := &ConcurrencyTestResult{}
	var wg sync.WaitGroup
	var mutex sync.Mutex

	latencies := make([]time.Duration, 0)

	for i := 0; i < config.NumGoroutines; i++ {
		wg.Add(1)
		go func(goroutineID int) {
			defer wg.Done()
			defer func() {
				if r := recover(); r != nil {
					mutex.Lock()
					result.PanicRecoveries++
					mutex.Unlock()
				}
			}()

			for j := 0; j < config.OperationsPerGoroutine; j++ {
				start := time.Now()
				err := operation(goroutineID)
				latency := time.Since(start)

				mutex.Lock()
				result.TotalOperations++
				latencies = append(latencies, latency)

				if err != nil {
					result.RaceConditions++
				} else {
					result.SuccessfulOps++
				}
				mutex.Unlock()
			}
		}(i)
	}

	wg.Wait()

	// Calculate latency statistics
	if len(latencies) > 0 {
		var total time.Duration
		var max time.Duration

		for _, latency := range latencies {
			total += latency
			if latency > max {
				max = latency
			}
		}

		result.AverageLatency = total / time.Duration(len(latencies))
		result.MaxLatency = max
	}

	return result
}

// ------------------------------------
// 🎯 Lifecycle Testing Utilities
// ------------------------------------

// LifecycleTestCase represents a component lifecycle test scenario
type LifecycleTestCase struct {
	Name                 string
	ComponentFactory     func() interface{}
	MountOperations      []func()
	UpdateOperations     []func()
	UnmountOperations    []func()
	ExpectedMountCalls   int
	ExpectedUpdateCalls  int
	ExpectedCleanupCalls int
	MaxCascadeDepth      int
}

// LifecycleTracker tracks component lifecycle events
type LifecycleTracker struct {
	mountCalls   int32
	updateCalls  int32
	cleanupCalls int32
	cascadeDepth int32
	maxDepth     int32
	errors       []error
	mutex        sync.RWMutex
}

// NewLifecycleTracker creates a new lifecycle tracker
func NewLifecycleTracker() *LifecycleTracker {
	return &LifecycleTracker{
		errors: make([]error, 0),
	}
}

// TrackMount records a component mount event
func (lt *LifecycleTracker) TrackMount() {
	atomic.AddInt32(&lt.mountCalls, 1)
}

// TrackUpdate records a component update event
func (lt *LifecycleTracker) TrackUpdate() {
	atomic.AddInt32(&lt.updateCalls, 1)
}

// TrackCleanup records a cleanup event
func (lt *LifecycleTracker) TrackCleanup() {
	atomic.AddInt32(&lt.cleanupCalls, 1)
}

// TrackCascadeDepth records cascade depth
func (lt *LifecycleTracker) TrackCascadeDepth(depth int32) {
	atomic.StoreInt32(&lt.cascadeDepth, depth)

	// Update max depth
	for {
		current := atomic.LoadInt32(&lt.maxDepth)
		if depth <= current || atomic.CompareAndSwapInt32(&lt.maxDepth, current, depth) {
			break
		}
	}
}

// GetLifecycleStats returns current lifecycle statistics
func (lt *LifecycleTracker) GetLifecycleStats() LifecycleStats {
	lt.mutex.RLock()
	defer lt.mutex.RUnlock()

	return LifecycleStats{
		MountCalls:   atomic.LoadInt32(&lt.mountCalls),
		UpdateCalls:  atomic.LoadInt32(&lt.updateCalls),
		CleanupCalls: atomic.LoadInt32(&lt.cleanupCalls),
		CascadeDepth: atomic.LoadInt32(&lt.cascadeDepth),
		MaxDepth:     atomic.LoadInt32(&lt.maxDepth),
		ErrorCount:   int32(len(lt.errors)),
	}
}

// LifecycleStats contains lifecycle execution statistics
type LifecycleStats struct {
	MountCalls   int32
	UpdateCalls  int32
	CleanupCalls int32
	CascadeDepth int32
	MaxDepth     int32
	ErrorCount   int32
}

// ------------------------------------
// 🎯 Validation Utilities
// ------------------------------------

// ValidateReactiveSystem performs comprehensive validation of reactive system
func ValidateReactiveSystem(t *testing.T, ctx *ReactiveTestContext, targets PerformanceTestConfig) {
	stats := ctx.GetStats()

	// Validate performance targets
	if stats.Duration > targets.DOMUpdateTarget {
		t.Errorf("Test duration %v exceeds DOM update target %v",
			stats.Duration, targets.DOMUpdateTarget)
	}

	// Validate memory usage
	memoryPerSignal := uint64(0)
	if stats.SignalCount > 0 {
		memoryPerSignal = stats.MemoryUsage / uint64(stats.SignalCount)
	}

	if memoryPerSignal > targets.MemoryPerSignalTarget {
		t.Errorf("Memory per signal %d bytes exceeds target %d bytes",
			memoryPerSignal, targets.MemoryPerSignalTarget)
	}

	// Validate cleanup
	if stats.CleanupCount < stats.SignalCount {
		t.Errorf("Insufficient cleanup: %d cleanups for %d signals",
			stats.CleanupCount, stats.SignalCount)
	}

	// Validate error count
	if stats.ErrorCount > 0 {
		t.Errorf("Reactive system generated %d errors", stats.ErrorCount)
	}
}

// ValidateLifecycleSystem validates component lifecycle behavior
func ValidateLifecycleSystem(t *testing.T, tracker *LifecycleTracker, testCase LifecycleTestCase) {
	stats := tracker.GetLifecycleStats()

	if stats.MountCalls != int32(testCase.ExpectedMountCalls) {
		t.Errorf("Expected %d mount calls, got %d",
			testCase.ExpectedMountCalls, stats.MountCalls)
	}

	if stats.UpdateCalls != int32(testCase.ExpectedUpdateCalls) {
		t.Errorf("Expected %d update calls, got %d",
			testCase.ExpectedUpdateCalls, stats.UpdateCalls)
	}

	if stats.CleanupCalls != int32(testCase.ExpectedCleanupCalls) {
		t.Errorf("Expected %d cleanup calls, got %d",
			testCase.ExpectedCleanupCalls, stats.CleanupCalls)
	}

	if stats.MaxDepth > int32(testCase.MaxCascadeDepth) {
		t.Errorf("Cascade depth %d exceeds maximum %d",
			stats.MaxDepth, testCase.MaxCascadeDepth)
	}
}
