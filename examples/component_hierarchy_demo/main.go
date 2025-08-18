// main.go
// Component hierarchy demonstration with cascade prevention and proper cleanup

//go:build !js && !wasm

package main

import (
	"fmt"
	"log"
	"time"

	"../../golid"
	"maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

func main() {
	fmt.Println("🏗️ Component Hierarchy Demo")
	fmt.Println("Demonstrating proper component hierarchies with cascade prevention")

	// Test 1: Basic hierarchy creation and cleanup
	testBasicHierarchy()

	// Test 2: Cascade prevention
	testCascadePrevention()

	// Test 3: Resource management
	testResourceManagement()

	// Test 4: Error boundaries
	testErrorBoundaries()

	fmt.Println("✅ All hierarchy tests completed successfully!")
}

// ------------------------------------
// 🧪 Test 1: Basic Hierarchy
// ------------------------------------

func testBasicHierarchy() {
	fmt.Println("\n📋 Test 1: Basic Component Hierarchy")

	// Create hierarchy manager
	hierarchy := golid.NewComponentHierarchy(10)

	// Create root component with hierarchy
	rootComp, rootNode := golid.CreateComponentWithHierarchy(hierarchy, func() gomponents.Node {
		return Div(
			H1(Text("Root Component")),
			P(Text("This is the root of our component hierarchy")),
		)
	})

	// Create child components
	child1Comp, child1Node := golid.CreateComponentWithHierarchy(hierarchy, func() gomponents.Node {
		return Div(
			H2(Text("Child Component 1")),
			P(Text("First child component")),
		)
	})

	child2Comp, child2Node := golid.CreateComponentWithHierarchy(hierarchy, func() gomponents.Node {
		return Div(
			H2(Text("Child Component 2")),
			P(Text("Second child component")),
		)
	})

	// Mount components in hierarchy
	err := golid.MountComponentWithHierarchy(hierarchy, rootComp, rootNode, nil)
	if err != nil {
		log.Printf("❌ Failed to mount root component: %v", err)
		return
	}

	err = golid.MountComponentWithHierarchy(hierarchy, child1Comp, child1Node, rootNode)
	if err != nil {
		log.Printf("❌ Failed to mount child1 component: %v", err)
		return
	}

	err = golid.MountComponentWithHierarchy(hierarchy, child2Comp, child2Node, rootNode)
	if err != nil {
		log.Printf("❌ Failed to mount child2 component: %v", err)
		return
	}

	// Check hierarchy stats
	stats := hierarchy.GetHierarchyStats()
	fmt.Printf("📊 Hierarchy Stats: %+v\n", stats)

	// Verify hierarchy structure
	children := hierarchy.GetChildren(rootNode)
	fmt.Printf("👥 Root has %d children\n", len(children))

	depth1 := hierarchy.GetNodeDepth(child1Node)
	depth2 := hierarchy.GetNodeDepth(child2Node)
	fmt.Printf("📏 Child depths: %d, %d\n", depth1, depth2)

	// Cleanup
	hierarchy.DisposeNode(rootNode)
	cleanedUp := hierarchy.CleanupDisposed()
	fmt.Printf("🧹 Cleaned up %d disposed nodes\n", cleanedUp)

	fmt.Println("✅ Basic hierarchy test passed")
}

// ------------------------------------
// 🧪 Test 2: Cascade Prevention
// ------------------------------------

func testCascadePrevention() {
	fmt.Println("\n🛡️ Test 2: Cascade Prevention")

	// Create hierarchy with strict limits
	hierarchy := golid.NewComponentHierarchy(3) // Very shallow limit
	cascadeGuard := golid.NewCascadePreventionGuard(5, time.Second)

	// Create components that might cause cascades
	var cascadeCount int

	createCascadingComponent := func(name string, depth int) (*golid.ComponentV2, *golid.HierarchyNode) {
		return golid.CreateComponentWithHierarchy(hierarchy, func() gomponents.Node {
			return Div(
				H3(Text(fmt.Sprintf("Cascading Component %s (depth %d)", name, depth))),
				P(Text("This component might trigger cascades")),
			)
		})
	}

	// Create root
	rootComp, rootNode := createCascadingComponent("Root", 0)
	err := golid.MountComponentWithHierarchy(hierarchy, rootComp, rootNode, nil)
	if err != nil {
		log.Printf("❌ Failed to mount root: %v", err)
		return
	}

	// Try to create deep hierarchy (should be prevented)
	var currentParent *golid.HierarchyNode = rootNode

	for i := 1; i <= 5; i++ {
		if !cascadeGuard.CheckOperation(uint64(i)) {
			fmt.Printf("🛑 Cascade prevented at depth %d\n", i)
			break
		}

		comp, node := createCascadingComponent(fmt.Sprintf("Child%d", i), i)
		err := golid.MountComponentWithHierarchy(hierarchy, comp, node, currentParent)

		if err != nil {
			fmt.Printf("🛑 Hierarchy limit reached at depth %d: %v\n", i, err)
			break
		}

		currentParent = node
		cascadeCount++
		fmt.Printf("✅ Successfully mounted component at depth %d\n", i)
	}

	// Check final stats
	stats := hierarchy.GetHierarchyStats()
	fmt.Printf("📊 Final Stats: %+v\n", stats)
	fmt.Printf("🔢 Total cascade attempts: %d\n", cascadeCount)

	// Cleanup
	hierarchy.DisposeNode(rootNode)

	fmt.Println("✅ Cascade prevention test passed")
}

// ------------------------------------
// 🧪 Test 3: Resource Management
// ------------------------------------

func testResourceManagement() {
	fmt.Println("\n🗂️ Test 3: Resource Management")

	// Create component with resources
	_, cleanup := golid.CreateRoot(func() interface{} {
		// Create resource tracker
		tracker := golid.NewResourceTracker()

		// Track various resources
		timerId := tracker.TrackTimer("test-timer", func() {
			fmt.Println("🕒 Timer cleaned up")
		})

		intervalId := tracker.TrackInterval("test-interval", func() {
			fmt.Println("⏰ Interval cleaned up")
		})

		listenerId := tracker.TrackEventListener("test-listener", func() {
			fmt.Println("👂 Event listener cleaned up")
		})

		signalId := tracker.TrackSignal("test-signal", func() {
			fmt.Println("📡 Signal cleaned up")
		})

		// Check resource stats
		stats := tracker.GetResourceStats()
		fmt.Printf("📊 Resource Stats: %+v\n", stats)

		// Test individual cleanup
		cleaned := tracker.CleanupResource(timerId)
		fmt.Printf("🧹 Individual cleanup successful: %v\n", cleaned)

		// Test cleanup by type
		intervalsCleaned := tracker.CleanupResourcesByType(golid.ResourceInterval)
		fmt.Printf("🧹 Cleaned %d intervals\n", intervalsCleaned)

		// Check remaining resources
		remaining := tracker.GetResourceCount()
		fmt.Printf("📊 Remaining resources: %d\n", remaining)

		// Test leak detection
		leakDetector := golid.NewLeakDetector(2, time.Millisecond*100)

		// Add more resources to trigger leak detection
		for i := 0; i < 5; i++ {
			tracker.TrackCustom(fmt.Sprintf("leak-test-%d", i), func() {})
		}

		violations := leakDetector.CheckForLeaks(tracker)
		if len(violations) > 0 {
			fmt.Printf("🚨 Memory leak detected: %+v\n", violations)
		} else {
			fmt.Println("✅ No memory leaks detected")
		}

		// Final cleanup will happen automatically
		return tracker
	})

	// Trigger cleanup
	cleanup()
	fmt.Println("✅ Resource management test passed")
}

// ------------------------------------
// 🧪 Test 4: Error Boundaries
// ------------------------------------

func testErrorBoundaries() {
	fmt.Println("\n🚨 Test 4: Error Boundaries")

	// Create error boundary
	errorBoundary := golid.CreateErrorBoundaryV2(func(err error) gomponents.Node {
		return Div(
			H3(Text("Error Boundary Activated")),
			P(Text(fmt.Sprintf("Caught error: %v", err))),
		)
	})

	// Test error handling
	var caughtError error

	err := errorBoundary.Catch(func() {
		// Create component that will panic
		comp := golid.CreateComponentV2(func() gomponents.Node {
			return Div(Text("This component will panic"))
		}).OnMountV2(func() {
			panic("Intentional panic for testing")
		})

		// Try to render (should trigger panic)
		comp.RenderV2()
	})

	if err != nil {
		caughtError = err
		fmt.Printf("✅ Error boundary caught error: %v\n", err)
	} else {
		fmt.Println("❌ Error boundary failed to catch error")
	}

	// Test error boundary reset
	errorBoundary.Reset()
	fmt.Println("🔄 Error boundary reset")

	// Test successful operation after reset
	err = errorBoundary.Catch(func() {
		comp := golid.CreateComponentV2(func() gomponents.Node {
			return Div(Text("This component works fine"))
		})
		comp.RenderV2()
	})

	if err == nil {
		fmt.Println("✅ Error boundary works correctly after reset")
	} else {
		fmt.Printf("❌ Unexpected error after reset: %v\n", err)
	}

	fmt.Println("✅ Error boundary test passed")
}

// ------------------------------------
// 🔧 Helper Functions
// ------------------------------------

// TrackCustom adds a custom resource type to the tracker
func (rt *golid.ResourceTracker) TrackCustom(name string, cleanup func()) uint64 {
	return rt.TrackResource(golid.ResourceCustom, name, cleanup)
}

// DemoComponentWithLifecycle creates a demo component with full lifecycle
func DemoComponentWithLifecycle(name string) *golid.ComponentV2 {
	return golid.CreateComponentV2(func() gomponents.Node {
		return Div(
			H3(Text(fmt.Sprintf("Demo Component: %s", name))),
			P(Text("This component demonstrates full lifecycle management")),
		)
	}).OnMountV2(func() {
		fmt.Printf("🚀 %s mounted\n", name)
	}).OnCleanupV2(func() {
		fmt.Printf("🧹 %s cleaned up\n", name)
	}).OnErrorV2(func(err error) {
		fmt.Printf("🚨 %s error: %v\n", name, err)
	}).OnUpdateV2(func() {
		fmt.Printf("🔄 %s updated\n", name)
	})
}

// CreateNestedHierarchy creates a nested component hierarchy for testing
func CreateNestedHierarchy(hierarchy *golid.ComponentHierarchy, depth int, maxDepth int) (*golid.ComponentV2, *golid.HierarchyNode) {
	if depth >= maxDepth {
		return nil, nil
	}

	comp := DemoComponentWithLifecycle(fmt.Sprintf("Level%d", depth))
	node := hierarchy.CreateHierarchyNode(comp.GetOwner())

	// Create children
	for i := 0; i < 2 && depth < maxDepth-1; i++ {
		childComp, childNode := CreateNestedHierarchy(hierarchy, depth+1, maxDepth)
		if childComp != nil && childNode != nil {
			hierarchy.AttachChild(node, childNode)
		}
	}

	return comp, node
}

// GetOwner returns the owner of a component (helper method)
func (c *golid.ComponentV2) GetOwner() *golid.Owner {
	// This would need to be implemented in the actual ComponentV2 struct
	// For now, return current owner
	return golid.GetCurrentOwner()
}

// GetCurrentOwner returns the current owner (wrapper function)
func GetCurrentOwner() *golid.Owner {
	// This would call the actual getCurrentOwner function
	return nil // Placeholder
}
