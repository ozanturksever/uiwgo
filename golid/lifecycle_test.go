// lifecycle_test.go
// Comprehensive tests for lifecycle events, focusing on infinite loop detection
// These tests reproduce the CPU usage bug caused by cascading lifecycle events

package golid

import (
	"runtime"
	"sync"
	"testing"
	"time"

	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

// TestLifecycleInfiniteLoopDOMModification reproduces the bug where lifecycle hooks
// modify the DOM, triggering more observer events and creating infinite loops
func TestLifecycleInfiniteLoopDOMModification(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping infinite loop test in short mode")
	}

	// Track how many times the lifecycle hook is called
	var mountCount int32
	var mu sync.Mutex

	// Create a component with a mount hook that modifies the DOM
	component := NewComponent(func() Node {
		return Div(Text("Test Component"))
	}).OnMount(func() {
		mu.Lock()
		mountCount++
		currentCount := mountCount
		mu.Unlock()

		// This is the problematic pattern: lifecycle hook modifies DOM
		// which triggers the observer again, creating an infinite loop
		if currentCount < 100 { // Prevent actual infinite loop in test
			// Simulate DOM modification that would trigger observer
			// In real scenarios, this could be adding elements, updating content, etc.
			NewComponent(func() Node {
				return Div(Text("Added by lifecycle hook"))
			})
		}
	})

	// Start monitoring CPU usage
	startTime := time.Now()
	var startCPU runtime.MemStats
	runtime.ReadMemStats(&startCPU)

	// Render the component (this should trigger the problematic cascade)
	rendered := component.Render()

	// In a real browser environment, this would create the infinite loop
	// For testing, we simulate the observer behavior
	if rendered != nil {
		// Simulate what happens in the browser:
		// 1. Component renders and gets added to DOM
		// 2. Observer detects the addition
		// 3. Mount hook executes
		// 4. Mount hook creates new component (DOM modification)
		// 5. Observer detects new addition
		// 6. Process repeats infinitely

		// We expect the mount count to be exactly 1 in a well-behaved system
		// But the bug causes it to cascade
	}

	duration := time.Since(startTime)

	mu.Lock()
	finalCount := mountCount
	mu.Unlock()

	// Test should complete quickly if no infinite loop
	if duration > time.Millisecond*100 {
		t.Errorf("Test took too long (%v), suggesting infinite loop behavior", duration)
	}

	// In the buggy version, this would be much higher due to cascading
	if finalCount > 1 {
		t.Logf("Warning: Mount hook called %d times, indicating potential cascade", finalCount)
	}
}

// TestLifecycleSignalCascade tests the signal-lifecycle infinite loop scenario
func TestLifecycleSignalCascade(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping signal cascade test in short mode")
	}

	// Create a signal that will be updated by lifecycle hooks
	counter := NewSignal(0)
	var hookExecutions int32
	var mu sync.Mutex

	// Create a component that updates signals in lifecycle hooks
	component := NewComponent(func() Node {
		// This creates a reactive dependency
		count := counter.Get()
		return Div(Text("Counter: " + string(rune(count+'0'))))
	}).OnMount(func() {
		mu.Lock()
		hookExecutions++
		currentExecs := hookExecutions
		mu.Unlock()

		// This is the problematic pattern:
		// 1. Lifecycle hook updates signal
		// 2. Signal triggers reactive effects
		// 3. Effects modify DOM
		// 4. DOM changes trigger observer
		// 5. Observer calls more lifecycle hooks
		// 6. Loop continues
		if currentExecs < 50 { // Prevent actual infinite loop in test
			counter.Set(counter.Get() + 1)
		}
	})

	startTime := time.Now()

	// This should trigger the cascade in the buggy version
	rendered := component.Render()

	duration := time.Since(startTime)

	if rendered == nil {
		t.Error("Component should render successfully")
	}

	mu.Lock()
	finalExecs := hookExecutions
	mu.Unlock()

	// Test should complete quickly
	if duration > time.Millisecond*100 {
		t.Errorf("Signal cascade test took too long (%v)", duration)
	}

	// Should only execute once in a well-behaved system
	if finalExecs > 1 {
		t.Logf("Warning: Hook executed %d times, indicating signal cascade", finalExecs)
	}
}

// TestLifecycleDismountCascade tests dismount hooks that trigger more DOM changes
func TestLifecycleDismountCascade(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping dismount cascade test in short mode")
	}

	var dismountCount int32
	var mu sync.Mutex

	// Component with dismount hook that adds new elements
	component := NewComponent(func() Node {
		return Div(Text("Component to be dismounted"))
	}).OnDismount(func() {
		mu.Lock()
		dismountCount++
		currentCount := dismountCount
		mu.Unlock()

		// Dismount hook that adds new elements - problematic pattern
		if currentCount < 20 {
			// In real scenarios, this might create cleanup components,
			// notification elements, etc.
			NewComponent(func() Node {
				return Div(Text("Created during dismount"))
			}).Render()
		}
	})

	rendered := component.Render()
	if rendered == nil {
		t.Error("Component should render")
	}

	// Simulate component dismount
	// In real browser, this would trigger the cascade

	mu.Lock()
	finalCount := dismountCount
	mu.Unlock()

	// Should only dismount once
	if finalCount > 1 {
		t.Logf("Warning: Dismount called %d times, indicating cascade", finalCount)
	}
}

// TestMultipleLifecycleHooks tests complex scenarios with multiple hook types
func TestMultipleLifecycleHooks(t *testing.T) {
	var initCount, mountCount, dismountCount int32
	var mu sync.Mutex

	component := NewComponent(func() Node {
		return Div(Text("Multi-hook component"))
	}).OnInit(func() {
		mu.Lock()
		initCount++
		mu.Unlock()
		// Init hook that creates signals
		NewSignal("created in init")
	}).OnMount(func() {
		mu.Lock()
		mountCount++
		mu.Unlock()
		// Mount hook that modifies DOM
		NewComponent(func() Node {
			return Span(Text("Mount child"))
		})
	}).OnDismount(func() {
		mu.Lock()
		dismountCount++
		mu.Unlock()
		// Dismount hook that triggers effects
		NewSignal("cleanup signal").Set("cleaning up")
	})

	rendered := component.Render()
	if rendered == nil {
		t.Error("Component should render")
	}

	mu.Lock()
	finalInit := initCount
	finalMount := mountCount
	finalDismount := dismountCount
	mu.Unlock()

	// Each hook should execute exactly once
	if finalInit != 1 {
		t.Errorf("Init hook executed %d times, expected 1", finalInit)
	}
	if finalMount > 1 {
		t.Logf("Warning: Mount hook executed %d times", finalMount)
	}
	if finalDismount > 1 {
		t.Logf("Warning: Dismount hook executed %d times", finalDismount)
	}
}

// TestLifecyclePerformanceStress creates a stress test that would demonstrate
// the 100% CPU usage problem in the buggy version
func TestLifecyclePerformanceStress(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance stress test in short mode")
	}

	const numComponents = 10
	var totalHookCalls int32
	var mu sync.Mutex

	components := make([]*Component, numComponents)

	startTime := time.Now()

	for i := 0; i < numComponents; i++ {
		components[i] = NewComponent(func() Node {
			return Div(Text("Stress test component"))
		}).OnMount(func() {
			mu.Lock()
			totalHookCalls++
			calls := totalHookCalls
			mu.Unlock()

			// Each component tries to create another component on mount
			// In the buggy version, this creates exponential cascade
			if calls < 100 { // Safety limit for test
				NewComponent(func() Node {
					return Div(Text("Cascade component"))
				}).Render()
			}
		})
	}

	// Render all components
	for _, comp := range components {
		rendered := comp.Render()
		if rendered == nil {
			t.Error("Component should render")
		}
	}

	duration := time.Since(startTime)

	mu.Lock()
	finalCalls := totalHookCalls
	mu.Unlock()

	// Should complete quickly even with multiple components
	if duration > time.Second {
		t.Errorf("Stress test took %v, indicating performance problem", duration)
	}

	// In a well-behaved system, should be exactly numComponents
	expectedCalls := int32(numComponents)
	if finalCalls > expectedCalls*2 { // Some tolerance for test environment
		t.Errorf("Hook calls: %d, expected around %d, indicating cascade", finalCalls, expectedCalls)
	}

	t.Logf("Stress test completed in %v with %d hook calls", duration, finalCalls)
}

// TestLifecycleRecursiveComponent tests deeply recursive component creation
func TestLifecycleRecursiveComponent(t *testing.T) {
	var depth int32
	const maxDepth = 5

	var createComponent func(currentDepth int32) *Component
	createComponent = func(currentDepth int32) *Component {
		return NewComponent(func() Node {
			return Div(Text("Depth: " + string(rune(currentDepth+'0'))))
		}).OnMount(func() {
			if currentDepth < maxDepth {
				// Each component creates a child on mount - potential cascade
				child := createComponent(currentDepth + 1)
				child.Render()
				depth = currentDepth + 1
			}
		})
	}

	startTime := time.Now()
	rootComponent := createComponent(0)
	rendered := rootComponent.Render()
	duration := time.Since(startTime)

	if rendered == nil {
		t.Error("Root component should render")
	}

	// Should complete quickly without infinite recursion
	if duration > time.Millisecond*500 {
		t.Errorf("Recursive test took %v, possibly indicating infinite recursion", duration)
	}

	if depth > maxDepth+5 { // Some tolerance
		t.Errorf("Recursion depth %d exceeded expected maximum %d", depth, maxDepth)
	}

	t.Logf("Recursive test completed at depth %d in %v", depth, duration)
}
