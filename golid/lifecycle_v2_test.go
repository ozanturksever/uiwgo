// lifecycle_v2_test.go
// Tests for enhanced component lifecycle system with cascade prevention

//go:build !js && !wasm

package golid

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"maragu.dev/gomponents"
)

// ------------------------------------
// 🧪 Component Hierarchy Tests
// ------------------------------------

func TestComponentHierarchyCreation(t *testing.T) {
	hierarchy := NewComponentHierarchy(10)

	if hierarchy == nil {
		t.Fatal("Failed to create component hierarchy")
	}

	stats := hierarchy.GetHierarchyStats()
	if stats["totalNodes"].(int) != 0 {
		t.Errorf("Expected 0 nodes, got %d", stats["totalNodes"].(int))
	}

	if stats["maxDepth"].(int) != 10 {
		t.Errorf("Expected maxDepth 10, got %d", stats["maxDepth"].(int))
	}
}

func TestHierarchyNodeOperations(t *testing.T) {
	hierarchy := NewComponentHierarchy(5)

	// Create root owner
	_, cleanup := CreateRoot(func() interface{} { return nil })
	defer cleanup()

	owner := getCurrentOwner()
	if owner == nil {
		t.Fatal("Failed to get current owner")
	}

	// Create nodes
	rootNode := hierarchy.CreateHierarchyNode(owner)
	childNode := hierarchy.CreateHierarchyNode(owner)

	if rootNode == nil || childNode == nil {
		t.Fatal("Failed to create hierarchy nodes")
	}

	// Test attachment
	err := hierarchy.AttachChild(rootNode, childNode)
	if err != nil {
		t.Fatalf("Failed to attach child: %v", err)
	}

	// Verify hierarchy
	children := hierarchy.GetChildren(rootNode)
	if len(children) != 1 {
		t.Errorf("Expected 1 child, got %d", len(children))
	}

	parent := hierarchy.GetParent(childNode)
	if parent != rootNode {
		t.Error("Child's parent is not the root node")
	}

	// Test depth calculation
	depth := hierarchy.GetNodeDepth(childNode)
	if depth != 1 {
		t.Errorf("Expected depth 1, got %d", depth)
	}
}

func TestHierarchyDepthLimit(t *testing.T) {
	hierarchy := NewComponentHierarchy(2) // Very shallow limit

	_, cleanup := CreateRoot(func() interface{} { return nil })
	defer cleanup()

	owner := getCurrentOwner()

	// Create nodes
	rootNode := hierarchy.CreateHierarchyNode(owner)
	child1Node := hierarchy.CreateHierarchyNode(owner)
	child2Node := hierarchy.CreateHierarchyNode(owner)
	child3Node := hierarchy.CreateHierarchyNode(owner)

	// Attach first level (should succeed)
	err := hierarchy.AttachChild(rootNode, child1Node)
	if err != nil {
		t.Fatalf("Failed to attach first level: %v", err)
	}

	// Attach second level (should succeed)
	err = hierarchy.AttachChild(child1Node, child2Node)
	if err != nil {
		t.Fatalf("Failed to attach second level: %v", err)
	}

	// Attach third level (should fail due to depth limit)
	err = hierarchy.AttachChild(child2Node, child3Node)
	if err == nil {
		t.Error("Expected depth limit error, but attachment succeeded")
	}

	if err != ErrHierarchyDepthExceeded {
		t.Errorf("Expected ErrHierarchyDepthExceeded, got %v", err)
	}
}

func TestCircularHierarchyPrevention(t *testing.T) {
	hierarchy := NewComponentHierarchy(10)

	_, cleanup := CreateRoot(func() interface{} { return nil })
	defer cleanup()

	owner := getCurrentOwner()

	// Create nodes
	node1 := hierarchy.CreateHierarchyNode(owner)
	node2 := hierarchy.CreateHierarchyNode(owner)
	node3 := hierarchy.CreateHierarchyNode(owner)

	// Create chain: node1 -> node2 -> node3
	err := hierarchy.AttachChild(node1, node2)
	if err != nil {
		t.Fatalf("Failed to attach node2 to node1: %v", err)
	}

	err = hierarchy.AttachChild(node2, node3)
	if err != nil {
		t.Fatalf("Failed to attach node3 to node2: %v", err)
	}

	// Try to create circular reference: node3 -> node1 (should fail)
	err = hierarchy.AttachChild(node3, node1)
	if err == nil {
		t.Error("Expected circular reference error, but attachment succeeded")
	}

	if err != ErrCircularHierarchy {
		t.Errorf("Expected ErrCircularHierarchy, got %v", err)
	}
}

// ------------------------------------
// 🧪 Resource Management Tests
// ------------------------------------

func TestResourceTracker(t *testing.T) {
	tracker := NewResourceTracker()

	if tracker == nil {
		t.Fatal("Failed to create resource tracker")
	}

	// Test resource tracking
	var cleaned1, cleaned2, cleaned3 bool

	id1 := tracker.TrackTimer("timer1", func() { cleaned1 = true })
	id2 := tracker.TrackInterval("interval1", func() { cleaned2 = true })
	id3 := tracker.TrackEventListener("listener1", func() { cleaned3 = true })

	if id1 == 0 || id2 == 0 || id3 == 0 {
		t.Error("Failed to track resources")
	}

	// Check resource count
	count := tracker.GetResourceCount()
	if count != 3 {
		t.Errorf("Expected 3 resources, got %d", count)
	}

	// Test individual cleanup
	success := tracker.CleanupResource(id1)
	if !success {
		t.Error("Failed to cleanup individual resource")
	}

	if !cleaned1 {
		t.Error("Resource cleanup function was not called")
	}

	// Test cleanup by type
	cleanedCount := tracker.CleanupResourcesByType(ResourceInterval)
	if cleanedCount != 1 {
		t.Errorf("Expected to clean 1 interval, got %d", cleanedCount)
	}

	if !cleaned2 {
		t.Error("Interval cleanup function was not called")
	}

	// Test cleanup all
	remainingCount := tracker.CleanupAll()
	if remainingCount != 1 {
		t.Errorf("Expected to clean 1 remaining resource, got %d", remainingCount)
	}

	if !cleaned3 {
		t.Error("Event listener cleanup function was not called")
	}
}

func TestResourceLeakDetection(t *testing.T) {
	tracker := NewResourceTracker()
	detector := NewLeakDetector(2, time.Millisecond*10)

	// Add resources below threshold
	tracker.TrackTimer("timer1", func() {})
	tracker.TrackTimer("timer2", func() {})

	violations := detector.CheckForLeaks(tracker)
	if len(violations) != 0 {
		t.Errorf("Expected no violations, got %d", len(violations))
	}

	// Add more resources to exceed threshold
	tracker.TrackTimer("timer3", func() {})
	tracker.TrackTimer("timer4", func() {})

	violations = detector.CheckForLeaks(tracker)
	if len(violations) == 0 {
		t.Error("Expected leak violations, got none")
	}

	if violations[0].ResourceType != ResourceTimer {
		t.Errorf("Expected ResourceTimer violation, got %v", violations[0].ResourceType)
	}

	if violations[0].Count <= violations[0].Threshold {
		t.Errorf("Expected count > threshold, got count=%d, threshold=%d",
			violations[0].Count, violations[0].Threshold)
	}
}

// ------------------------------------
// 🧪 Lifecycle Guard Tests
// ------------------------------------

func TestLifecycleGuard(t *testing.T) {
	guard := NewLifecycleGuardV2(3)

	// Test normal operation
	if !guard.Enter(1) {
		t.Error("First enter should succeed")
	}

	if !guard.Enter(2) {
		t.Error("Second enter should succeed")
	}

	if !guard.Enter(3) {
		t.Error("Third enter should succeed")
	}

	// Test depth limit
	if guard.Enter(4) {
		t.Error("Fourth enter should fail due to depth limit")
	}

	// Test exit
	guard.Exit(1)
	guard.Exit(2)
	guard.Exit(3)

	// Should be able to enter again after exits
	if !guard.Enter(5) {
		t.Error("Enter should succeed after exits")
	}

	guard.Exit(5)
}

func TestLifecycleGuardCircularPrevention(t *testing.T) {
	guard := NewLifecycleGuardV2(10)

	// Enter same component twice (should fail second time)
	if !guard.Enter(1) {
		t.Error("First enter should succeed")
	}

	if guard.Enter(1) {
		t.Error("Second enter with same ID should fail")
	}

	guard.Exit(1)

	// Should be able to enter again after exit
	if !guard.Enter(1) {
		t.Error("Enter should succeed after exit")
	}

	guard.Exit(1)
}

// ------------------------------------
// 🧪 Cascade Prevention Tests
// ------------------------------------

func TestCascadePrevention(t *testing.T) {
	guard := NewCascadePreventionGuard(3, time.Millisecond*100)

	// Test normal operations
	for i := 0; i < 3; i++ {
		if !guard.CheckOperation(1) {
			t.Errorf("Operation %d should be allowed", i)
		}
	}

	// Test cascade prevention
	if guard.CheckOperation(1) {
		t.Error("Fourth operation should be prevented")
	}

	// Test time window reset
	time.Sleep(time.Millisecond * 150)

	if !guard.CheckOperation(1) {
		t.Error("Operation should be allowed after time window")
	}
}

// ------------------------------------
// 🧪 Error Boundary Tests
// ------------------------------------

func TestErrorBoundary(t *testing.T) {
	var caughtError error

	boundary := CreateErrorBoundaryV2(func(err error) gomponents.Node {
		caughtError = err
		return nil
	})

	// Test error catching
	err := boundary.Catch(func() {
		panic("test panic")
	})

	if err == nil {
		t.Error("Expected error to be caught")
	}

	// Test reset
	boundary.Reset()

	// Test normal operation after reset
	err = boundary.Catch(func() {
		// Normal operation
	})

	if err != nil {
		t.Errorf("Expected no error after reset, got %v", err)
	}
}

// ------------------------------------
// 🧪 Concurrent Safety Tests
// ------------------------------------

func TestHierarchyConcurrentSafety(t *testing.T) {
	hierarchy := NewComponentHierarchy(10)

	_, cleanup := CreateRoot(func() interface{} { return nil })
	defer cleanup()

	owner := getCurrentOwner()

	var wg sync.WaitGroup
	const numGoroutines = 10
	const nodesPerGoroutine = 5

	// Create nodes concurrently
	wg.Add(numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < nodesPerGoroutine; j++ {
				node := hierarchy.CreateHierarchyNode(owner)
				if node == nil {
					t.Errorf("Failed to create node in goroutine")
				}
			}
		}()
	}

	wg.Wait()

	stats := hierarchy.GetHierarchyStats()
	expectedNodes := numGoroutines * nodesPerGoroutine
	if stats["totalNodes"].(int) != expectedNodes {
		t.Errorf("Expected %d nodes, got %d", expectedNodes, stats["totalNodes"].(int))
	}
}

func TestResourceTrackerConcurrentSafety(t *testing.T) {
	tracker := NewResourceTracker()

	var wg sync.WaitGroup
	const numGoroutines = 10
	const resourcesPerGoroutine = 5

	// Track resources concurrently
	wg.Add(numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < resourcesPerGoroutine; j++ {
				tracker.TrackTimer(fmt.Sprintf("timer-%d-%d", id, j), func() {})
			}
		}(i)
	}

	wg.Wait()

	expectedResources := numGoroutines * resourcesPerGoroutine
	if tracker.GetResourceCount() != expectedResources {
		t.Errorf("Expected %d resources, got %d", expectedResources, tracker.GetResourceCount())
	}

	// Cleanup concurrently
	wg.Add(numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			tracker.CleanupResourcesByType(ResourceTimer)
		}()
	}

	wg.Wait()

	// All resources should be cleaned up
	if tracker.GetResourceCount() != 0 {
		t.Errorf("Expected 0 resources after cleanup, got %d", tracker.GetResourceCount())
	}
}

// ------------------------------------
// 🧪 Integration Tests
// ------------------------------------

func TestComponentLifecycleIntegration(t *testing.T) {
	hierarchy := NewComponentHierarchy(5)

	// Create root context
	_, cleanup := CreateRoot(func() interface{} {
		// Create component with hierarchy
		comp, node := CreateComponentWithHierarchy(hierarchy, func() gomponents.Node {
			return nil // Simple test component
		})

		if comp == nil || node == nil {
			t.Fatal("Failed to create component with hierarchy")
		}

		// Test mounting
		err := MountComponentWithHierarchy(hierarchy, comp, node, nil)
		if err != nil {
			t.Fatalf("Failed to mount component: %v", err)
		}

		// Test stats
		stats := comp.GetComponentStatsV2()
		if stats["state"].(string) != "Mounted" {
			t.Errorf("Expected Mounted state, got %s", stats["state"].(string))
		}

		// Test unmounting
		err = comp.UnmountV2()
		if err != nil {
			t.Fatalf("Failed to unmount component: %v", err)
		}

		stats = comp.GetComponentStatsV2()
		if stats["state"].(string) != "Unmounted" {
			t.Errorf("Expected Unmounted state, got %s", stats["state"].(string))
		}

		return nil
	})

	cleanup()
}

// ------------------------------------
// 🧪 Performance Tests
// ------------------------------------

func BenchmarkHierarchyNodeCreation(b *testing.B) {
	hierarchy := NewComponentHierarchy(100)

	_, cleanup := CreateRoot(func() interface{} { return nil })
	defer cleanup()

	owner := getCurrentOwner()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		hierarchy.CreateHierarchyNode(owner)
	}
}

func BenchmarkResourceTracking(b *testing.B) {
	tracker := NewResourceTracker()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tracker.TrackTimer("benchmark-timer", func() {})
	}
}

func BenchmarkCascadeGuardCheck(b *testing.B) {
	guard := NewCascadePreventionGuard(1000, time.Second)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		guard.CheckOperation(uint64(i % 100))
	}
}
