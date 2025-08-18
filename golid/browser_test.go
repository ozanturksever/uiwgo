//go:build js && wasm

// browser_test.go
// Comprehensive browser-based tests for reactive systems with real DOM integration

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

// ------------------------------------
// 🎯 Reactive Test Context
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

// GetCurrentAllocation returns current allocation vs baseline
func (mt *MemoryTracker) GetCurrentAllocation() uint64 {
	var current runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&current)
	return current.Alloc - mt.baseline.Alloc
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

// ------------------------------------
// 🎯 Test Utilities
// ------------------------------------

// TriggerEvent triggers a DOM event on an element
func TriggerEvent(element js.Value, eventType string) {
	if !element.Truthy() {
		return
	}

	// Check if document is available
	document := js.Global().Get("document")
	if !document.Truthy() {
		return
	}

	// Create event based on type for better compatibility
	var event js.Value
	switch eventType {
	case "click", "mousedown", "mouseup":
		event = document.Call("createEvent", "MouseEvent")
		event.Call("initMouseEvent", eventType, true, true, js.Global().Get("window"), 0, 0, 0, 0, 0, false, false, false, false, 0, js.Null())
	case "input", "change":
		event = document.Call("createEvent", "Event")
		event.Call("initEvent", eventType, true, true)
	case "resize":
		// Special handling for resize events
		event = document.Call("createEvent", "UIEvent")
		event.Call("initUIEvent", eventType, true, true, js.Global().Get("window"), 0)
	default:
		event = document.Call("createEvent", "Event")
		event.Call("initEvent", eventType, true, true)
	}

	// Dispatch the event
	element.Call("dispatchEvent", event)
}

// AssertElementText checks if an element has expected text content
func AssertElementText(t *testing.T, element js.Value, expected string) {
	if !element.Truthy() {
		t.Errorf("Expected text %q, but element is null/undefined", expected)
		return
	}

	textContent := element.Get("textContent")
	if !textContent.Truthy() {
		t.Errorf("Expected text %q, but textContent is null/undefined", expected)
		return
	}

	actual := textContent.String()
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

// Performance targets from roadmap
const (
	SignalUpdateTarget    = 5 * time.Microsecond  // Target: 5μs
	DOMUpdateTarget       = 10 * time.Millisecond // Target: 10ms
	MemoryPerSignalTarget = 200                   // Target: 200B
)

// ------------------------------------
// 🎯 Signal System Browser Tests
// ------------------------------------

func TestSignalWithRealDOM(t *testing.T) {
	env := NewBrowserTestEnvironment(t)
	defer env.Cleanup()

	ctx := NewReactiveTestContext()
	ResetReactiveContext()
	ResetScheduler()

	// Create test element
	element := env.CreateTestElement("div")

	// Create signal and bind to DOM
	getter, setter := CreateSignal("initial")
	ctx.TrackSignalCreation()

	// Create reactive DOM binding
	binding := BindTextReactive(element, func() string {
		return getter()
	})

	if binding == nil {
		t.Fatal("BindTextReactive should return non-nil binding")
	}

	FlushScheduler()

	// Verify initial DOM content
	AssertElementText(t, element, "initial")

	// Update signal and verify DOM updates
	ctx.perfTimer.Mark("signal_update_start")
	setter("updated")
	FlushScheduler()
	ctx.perfTimer.Mark("signal_update_end")

	// Verify DOM was updated
	AssertElementText(t, element, "updated")

	// Validate performance
	updateDuration := ctx.perfTimer.Measure("signal_update_start", "signal_update_end")
	AssertPerformance(t, updateDuration, SignalUpdateTarget, "signal update")

	// Validate memory usage
	stats := ctx.GetStats()
	AssertMemoryUsage(t, stats.MemoryUsage, MemoryPerSignalTarget, "signal creation")
}

func TestEffectWithRealDOM(t *testing.T) {
	env := NewBrowserTestEnvironment(t)
	defer env.Cleanup()

	ctx := NewReactiveTestContext()
	ResetReactiveContext()
	ResetScheduler()

	// Create test elements
	input := env.CreateTestElement("input")
	output := env.CreateTestElement("div")

	// Create signals
	value, setValue := CreateSignal("")
	ctx.TrackSignalCreation()

	// Create effect that updates DOM
	var effectExecutions int
	CreateEffect(func() {
		ctx.TrackEffectExecution()
		effectExecutions++
		currentValue := value()
		output.Set("textContent", fmt.Sprintf("Value: %s", currentValue))
	}, nil)

	FlushScheduler()

	// Verify initial effect execution
	if effectExecutions != 1 {
		t.Errorf("Expected 1 initial effect execution, got %d", effectExecutions)
	}

	AssertElementText(t, output, "Value: ")

	// Simulate user input
	input.Set("value", "test input")
	setValue("test input")
	FlushScheduler()

	// Verify effect re-execution
	if effectExecutions != 2 {
		t.Errorf("Expected 2 effect executions after update, got %d", effectExecutions)
	}

	AssertElementText(t, output, "Value: test input")

	// Validate performance
	stats := ctx.GetStats()
	AssertPerformance(t, stats.Duration, DOMUpdateTarget, "effect execution")
}

func TestMemoWithRealDOM(t *testing.T) {
	env := NewBrowserTestEnvironment(t)
	defer env.Cleanup()

	ctx := NewReactiveTestContext()
	ResetReactiveContext()
	ResetScheduler()

	// Create test element
	element := env.CreateTestElement("div")

	// Create signals
	firstName, setFirstName := CreateSignal("John")
	lastName, setLastName := CreateSignal("Doe")
	ctx.TrackSignalCreation()
	ctx.TrackSignalCreation()

	// Create memo
	var computations int
	fullName := CreateMemo(func() string {
		ctx.TrackComputation()
		computations++
		return firstName() + " " + lastName()
	}, nil)

	// Bind memo to DOM
	BindTextReactive(element, func() string {
		return fullName()
	})

	FlushScheduler()

	// Verify initial computation
	if computations != 1 {
		t.Errorf("Expected 1 initial computation, got %d", computations)
	}

	AssertElementText(t, element, "John Doe")

	// Update first name
	setFirstName("Jane")
	FlushScheduler()

	// Verify recomputation
	if computations != 2 {
		t.Errorf("Expected 2 computations after first name update, got %d", computations)
	}

	AssertElementText(t, element, "Jane Doe")

	// Batch update both names
	Batch(func() {
		setFirstName("Alice")
		setLastName("Smith")
	})
	FlushScheduler()

	// Verify single recomputation for batched updates
	if computations != 3 {
		t.Errorf("Expected 3 computations after batch update, got %d", computations)
	}

	AssertElementText(t, element, "Alice Smith")
}

// ------------------------------------
// 🎯 DOM Manipulation Tests
// ------------------------------------

func TestReactiveDOMBindings(t *testing.T) {
	env := NewBrowserTestEnvironment(t)
	defer env.Cleanup()

	ResetReactiveContext()
	ResetScheduler()

	// Create test element
	element := env.CreateTestElement("div")

	// Test attribute binding
	isActive, setActive := CreateSignal(false)
	BindAttributeReactive(element, "data-active", func() string {
		if isActive() {
			return "true"
		}
		return "false"
	})

	FlushScheduler()
	AssertElementAttribute(t, element, "data-active", "false")

	// Update signal
	setActive(true)
	FlushScheduler()
	AssertElementAttribute(t, element, "data-active", "true")

	// Test class binding
	BindClassReactive(element, "active", func() bool {
		return isActive()
	})

	FlushScheduler()

	// Verify class was added
	classList := element.Get("classList")
	if !classList.Call("contains", "active").Bool() {
		t.Error("Element should have 'active' class")
	}

	// Update signal to false
	setActive(false)
	FlushScheduler()

	// Verify class was removed
	if classList.Call("contains", "active").Bool() {
		t.Error("Element should not have 'active' class")
	}

	// Test style binding
	color, setColor := CreateSignal("red")
	BindStyleReactive(element, "color", func() string {
		return color()
	})

	FlushScheduler()

	// Verify style was set
	style := element.Get("style").Call("getPropertyValue", "color").String()
	if style != "red" {
		t.Errorf("Expected color 'red', got '%s'", style)
	}

	// Update color
	setColor("blue")
	FlushScheduler()

	style = element.Get("style").Call("getPropertyValue", "color").String()
	if style != "blue" {
		t.Errorf("Expected color 'blue', got '%s'", style)
	}
}

func TestEventSystemIntegration(t *testing.T) {
	env := NewBrowserTestEnvironment(t)
	defer env.Cleanup()

	ResetReactiveContext()
	ResetScheduler()

	// Create test elements
	button := env.CreateTestElement("button")
	counter := env.CreateTestElement("div")

	// Create reactive state
	count, setCount := CreateSignal(0)

	// Bind counter display
	BindTextReactive(counter, func() string {
		return fmt.Sprintf("Count: %d", count())
	})

	// Set up reactive event handling
	var clickCount int
	cleanup := SubscribeReactive(button, "click", func(e js.Value) {
		clickCount++
		setCount(count() + 1)
	})

	FlushScheduler()

	// Verify initial state
	AssertElementText(t, counter, "Count: 0")

	// Simulate click events
	TriggerEvent(button, "click")
	FlushScheduler()

	// Verify reactive update
	if clickCount != 1 {
		t.Errorf("Expected 1 click, got %d", clickCount)
	}
	AssertElementText(t, counter, "Count: 1")

	// Trigger multiple clicks
	for i := 0; i < 5; i++ {
		TriggerEvent(button, "click")
	}
	FlushScheduler()

	// Verify final state
	if clickCount != 6 {
		t.Errorf("Expected 6 clicks, got %d", clickCount)
	}
	AssertElementText(t, counter, "Count: 6")

	// Cleanup
	cleanup()
}

// ------------------------------------
// 🎯 Component Lifecycle Tests
// ------------------------------------

func TestComponentLifecycleWithDOM(t *testing.T) {
	env := NewBrowserTestEnvironment(t)
	defer env.Cleanup()

	ResetReactiveContext()
	ResetScheduler()

	// Create component with lifecycle hooks
	var cleanupCalled bool
	var mountCalls, updateCalls int

	result, cleanup := CreateRoot(func() interface{} {
		mountCalls++

		// Create reactive state
		value, setValue := CreateSignal("initial")

		// Create DOM element
		element := env.CreateTestElement("div")
		BindTextReactive(element, func() string {
			updateCalls++
			return value()
		})

		// Register cleanup
		OnCleanup(func() {
			cleanupCalled = true
		})

		// Simulate component updates
		setValue("updated")
		setValue("final")

		return element
	})

	FlushScheduler()

	// Verify component was created
	if result == nil {
		t.Fatal("Component creation should return non-nil result")
	}

	// Verify lifecycle calls
	if mountCalls != 1 {
		t.Errorf("Expected 1 mount call, got %d", mountCalls)
	}

	if updateCalls < 2 {
		t.Errorf("Expected at least 2 update calls, got %d", updateCalls)
	}

	// Cleanup component
	cleanup()

	// Verify cleanup was called
	if !cleanupCalled {
		t.Error("Cleanup function should have been called")
	}
}

// ------------------------------------
// 🎯 Integration Tests
// ------------------------------------

func TestFullReactiveApplication(t *testing.T) {
	env := NewBrowserTestEnvironment(t)
	defer env.Cleanup()

	ctx := NewReactiveTestContext()
	ResetReactiveContext()
	ResetScheduler()

	// Create a mini todo application
	input := env.CreateTestElement("input")
	addButton := env.CreateTestElement("button")
	todoList := env.CreateTestElement("ul")

	// Application state
	todos, setTodos := CreateSignal([]string{})
	inputValue, setInputValue := CreateSignal("")
	ctx.TrackSignalCreation()
	ctx.TrackSignalCreation()

	// Bind input value
	BindAttributeReactive(input, "value", func() string {
		return inputValue()
	})

	// Handle input changes
	SubscribeReactive(input, "input", func(e js.Value) {
		value := e.Get("target").Get("value").String()
		setInputValue(value)
	})

	// Handle add button clicks
	SubscribeReactive(addButton, "click", func(e js.Value) {
		current := inputValue()
		if current != "" {
			currentTodos := todos()
			newTodos := append(currentTodos, current)
			setTodos(newTodos)
			setInputValue("")
		}
	})

	// Render todo list
	CreateEffect(func() {
		ctx.TrackEffectExecution()
		currentTodos := todos()

		// Clear existing todos
		todoList.Set("innerHTML", "")

		// Add each todo
		for _, todo := range currentTodos {
			li := js.Global().Get("document").Call("createElement", "li")
			li.Set("textContent", todo)
			todoList.Call("appendChild", li)
		}
	}, nil)

	FlushScheduler()

	// Test adding todos
	input.Set("value", "Test todo 1")
	TriggerEvent(input, "input")
	FlushScheduler()

	TriggerEvent(addButton, "click")
	FlushScheduler()

	// Verify todo was added
	todoItems := todoList.Get("children")
	if todoItems.Get("length").Int() != 1 {
		t.Errorf("Expected 1 todo item, got %d", todoItems.Get("length").Int())
	}

	firstTodo := todoItems.Call("item", 0)
	AssertElementText(t, firstTodo, "Test todo 1")

	// Add more todos
	for i := 2; i <= 5; i++ {
		input.Set("value", fmt.Sprintf("Test todo %d", i))
		TriggerEvent(input, "input")
		TriggerEvent(addButton, "click")
		FlushScheduler()
	}

	// Verify all todos were added
	if todoList.Get("children").Get("length").Int() != 5 {
		t.Errorf("Expected 5 todo items, got %d", todoList.Get("children").Get("length").Int())
	}

	// Validate overall system performance
	stats := ctx.GetStats()
	maxMemoryUsage := uint64(1024 * 1024) // 1MB limit for this test
	AssertMemoryUsage(t, stats.MemoryUsage, maxMemoryUsage, "todo application")
}

func TestBrowserAPIIntegration(t *testing.T) {
	env := NewBrowserTestEnvironment(t)
	defer env.Cleanup()

	ResetReactiveContext()
	ResetScheduler()

	// Test localStorage integration
	element := env.CreateTestElement("div")

	// Create signal that syncs with localStorage
	stored := js.Global().Get("localStorage").Call("getItem", "test-value")
	initialValue := "default"
	if stored.Truthy() {
		initialValue = stored.String()
	}

	value, setValue := CreateSignal(initialValue)

	// Effect to sync with localStorage
	CreateEffect(func() {
		current := value()
		js.Global().Get("localStorage").Call("setItem", "test-value", current)
	}, nil)

	// Bind to DOM
	BindTextReactive(element, func() string {
		return fmt.Sprintf("Stored: %s", value())
	})

	FlushScheduler()

	// Test initial state
	AssertElementText(t, element, "Stored: default")

	// Update value
	setValue("updated value")
	FlushScheduler()

	// Verify DOM update
	AssertElementText(t, element, "Stored: updated value")

	// Verify localStorage was updated
	storedValue := js.Global().Get("localStorage").Call("getItem", "test-value").String()
	if storedValue != "updated value" {
		t.Errorf("Expected localStorage value 'updated value', got '%s'", storedValue)
	}

	// Test window resize integration
	resizeCount := 0
	SubscribeReactive(js.Global().Get("window"), "resize", func(e js.Value) {
		resizeCount++
	})

	// Simulate resize event
	TriggerEvent(js.Global().Get("window"), "resize")

	if resizeCount != 1 {
		t.Errorf("Expected 1 resize event, got %d", resizeCount)
	}

	// Cleanup localStorage
	js.Global().Get("localStorage").Call("removeItem", "test-value")
}
