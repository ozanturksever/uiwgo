//go:build js && wasm

package dom

import (
	"runtime"
	"testing"
	"time"

	reactivity "github.com/ozanturksever/uiwgo/reactivity"
	"honnef.co/go/js/dom/v2"
)

// TestMemoryLeakRepeatedMountUnmount tests that repeated mount/unmount cycles
// don't cause memory leaks by properly disposing of scopes and effects
func TestMemoryLeakRepeatedMountUnmount(t *testing.T) {
	// Force garbage collection before starting
	runtime.GC()
	runtime.GC()
	time.Sleep(10 * time.Millisecond)
	
	initialMemStats := &runtime.MemStats{}
	runtime.ReadMemStats(initialMemStats)
	
	// Perform many mount/unmount cycles
	const cycles = 100
	for i := 0; i < cycles; i++ {
		// Create a signal that will be used in reactive bindings
		signal := reactivity.CreateSignal("test-value")
		
		// Create element with reactive bindings
		eb := NewElement("div")
		eb.BindReactiveText(func() string {
			return signal.Get()
		})
		eb.BindReactiveAttribute("class", func() string {
			return "test-class-" + signal.Get()
		})
		
		// Add event listeners
		eb.OnClick(func(event dom.Event) {
			// Do nothing, just test cleanup
		})
		
		// Build with cleanup
		element, cleanup := eb.BuildWithCleanup()
		
		// Trigger some updates to ensure effects are active
		signal.Set("updated-value")
		
		// Verify the element is working
		if element.TextContent() != "updated-value" {
			t.Errorf("Cycle %d: Expected 'updated-value', got '%s'", i, element.TextContent())
		}
		
		// Cleanup (unmount)
		cleanup()
		
		// Verify cleanup worked by trying to update signal
		signal.Set("should-not-update")
		time.Sleep(1 * time.Millisecond) // Give time for any potential updates
		
		if element.TextContent() != "updated-value" {
			t.Errorf("Cycle %d: Element should not update after cleanup, got '%s'", i, element.TextContent())
		}
		
		// Force garbage collection every 10 cycles
		if i%10 == 9 {
			runtime.GC()
			runtime.GC()
			time.Sleep(1 * time.Millisecond)
		}
	}
	
	// Final garbage collection
	runtime.GC()
	runtime.GC()
	time.Sleep(10 * time.Millisecond)
	
	finalMemStats := &runtime.MemStats{}
	runtime.ReadMemStats(finalMemStats)
	
	// Check for significant memory growth
	// We allow some growth but not excessive (more than 1MB)
	memoryGrowth := finalMemStats.Alloc - initialMemStats.Alloc
	const maxAllowedGrowth = 1024 * 1024 // 1MB
	
	if memoryGrowth > maxAllowedGrowth {
		t.Errorf("Potential memory leak detected: memory grew by %d bytes (max allowed: %d)", memoryGrowth, maxAllowedGrowth)
	}
	
	t.Logf("Memory growth after %d cycles: %d bytes", cycles, memoryGrowth)
}

// TestMemoryLeakNestedScopes tests that nested scopes are properly cleaned up
func TestMemoryLeakNestedScopes(t *testing.T) {
	// Force garbage collection before starting
	runtime.GC()
	runtime.GC()
	time.Sleep(10 * time.Millisecond)
	
	initialMemStats := &runtime.MemStats{}
	runtime.ReadMemStats(initialMemStats)
	
	const cycles = 50
	for i := 0; i < cycles; i++ {
		// Create parent scope
		parentScope := reactivity.NewCleanupScope(nil)
		reactivity.SetCurrentCleanupScope(parentScope)
		
		// Create parent element
		parentSignal := reactivity.CreateSignal("parent")
		parentEB := NewElement("div")
		parentEB.BindReactiveText(func() string {
			return parentSignal.Get()
		})
		
		// Create multiple child elements within parent scope
		reactivity.SetCurrentCleanupScope(parentEB.GetScope())
		for j := 0; j < 5; j++ {
			childSignal := reactivity.CreateSignal("child")
			childEB := NewElement("span")
			childEB.BindReactiveText(func() string {
				return childSignal.Get()
			})
			childEB.OnClick(func(event dom.Event) {
				// Do nothing
			})
			
			// Build child
			childEB.Build()
			
			// Update child signal
			childSignal.Set("child-updated")
		}
		
		// Build parent
		parentEB.Build()
		
		// Update parent signal
		parentSignal.Set("parent-updated")
		
		// Dispose parent scope (should dispose all children)
		parentScope.Dispose()
		
		// Clean up current scope
		reactivity.SetCurrentCleanupScope(nil)
		
		// Force garbage collection every 10 cycles
		if i%10 == 9 {
			runtime.GC()
			runtime.GC()
			time.Sleep(1 * time.Millisecond)
		}
	}
	
	// Final garbage collection
	runtime.GC()
	runtime.GC()
	time.Sleep(10 * time.Millisecond)
	
	finalMemStats := &runtime.MemStats{}
	runtime.ReadMemStats(finalMemStats)
	
	// Check for significant memory growth
	memoryGrowth := finalMemStats.Alloc - initialMemStats.Alloc
	const maxAllowedGrowth = 1024 * 1024 // 1MB
	
	if memoryGrowth > maxAllowedGrowth {
		t.Errorf("Potential memory leak detected in nested scopes: memory grew by %d bytes (max allowed: %d)", memoryGrowth, maxAllowedGrowth)
	}
	
	t.Logf("Memory growth after %d nested scope cycles: %d bytes", cycles, memoryGrowth)
}

// TestMemoryLeakEventListeners tests that event listeners are properly cleaned up
func TestMemoryLeakEventListeners(t *testing.T) {
	// Force garbage collection before starting
	runtime.GC()
	runtime.GC()
	time.Sleep(10 * time.Millisecond)
	
	initialMemStats := &runtime.MemStats{}
	runtime.ReadMemStats(initialMemStats)
	
	const cycles = 100
	for i := 0; i < cycles; i++ {
		// Create element with multiple event listeners
		eb := NewElement("button")
		
		// Add multiple event listeners
		eb.OnClick(func(event dom.Event) {
			// Handler 1
		})
		eb.OnEvent("mouseenter", func(event dom.Event) {
			// Handler 2
		})
		eb.OnEvent("mouseleave", func(event dom.Event) {
			// Handler 3
		})
		eb.OnEvent("focus", func(event dom.Event) {
			// Handler 4
		})
		eb.OnEvent("blur", func(event dom.Event) {
			// Handler 5
		})
		
		// Build with cleanup
		element, cleanup := eb.BuildWithCleanup()
		
		// Verify element was created
		if element.TagName() != "BUTTON" {
			t.Errorf("Expected BUTTON, got %s", element.TagName())
		}
		
		// Cleanup
		cleanup()
		
		// Force garbage collection every 20 cycles
		if i%20 == 19 {
			runtime.GC()
			runtime.GC()
			time.Sleep(1 * time.Millisecond)
		}
	}
	
	// Final garbage collection
	runtime.GC()
	runtime.GC()
	time.Sleep(10 * time.Millisecond)
	
	finalMemStats := &runtime.MemStats{}
	runtime.ReadMemStats(finalMemStats)
	
	// Check for significant memory growth
	memoryGrowth := finalMemStats.Alloc - initialMemStats.Alloc
	const maxAllowedGrowth = 512 * 1024 // 512KB (event listeners should be lightweight)
	
	if memoryGrowth > maxAllowedGrowth {
		t.Errorf("Potential memory leak detected in event listeners: memory grew by %d bytes (max allowed: %d)", memoryGrowth, maxAllowedGrowth)
	}
	
	t.Logf("Memory growth after %d event listener cycles: %d bytes", cycles, memoryGrowth)
}