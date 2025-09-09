//go:build js && wasm

package react

import (
	"github.com/ozanturksever/logutil"
	"syscall/js"
	"testing"
	"time"
)

func TestAutoUnmount(t *testing.T) {
	setupMockReactCompat()
	defer teardownMockReactCompat()

	// Initialize React bridge
	if err := InitializeBridge(); err != nil {
		t.Fatalf("Failed to initialize React bridge: %v", err)
	}

	// Register a simple test component
	componentName := "TestAutoUnmountComponent"
	componentJS := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		return js.Global().Get("React").Call("createElement", "div", js.ValueOf(map[string]interface{}{
			"children": "Auto-unmount test component",
		}))
	})
	defer componentJS.Release()

	if err := Register(componentName, componentJS.Value); err != nil {
		t.Fatalf("Failed to register component: %v", err)
	}

	// Create a container element
	container := js.Global().Get("document").Call("createElement", "div")
	container.Set("id", "auto-unmount-test-container")
	js.Global().Get("document").Get("body").Call("appendChild", container)

	// Render the component
	instanceID, err := Render(componentName, map[string]interface{}{}, &RenderOptions{
		ContainerID: "auto-unmount-test-container",
	})
	if err != nil {
		t.Fatalf("Failed to render component: %v", err)
	}

	// Verify the component was rendered
	instancesMap := js.Global().Get("window").Get("ReactCompat").Call("getInstancesMap")
	if !instancesMap.Call("has", js.ValueOf(string(instanceID))).Bool() {
		t.Fatalf("Component instance not found after render")
	}

	logutil.Logf("Component rendered with instance ID: %s", instanceID)

	// Remove the container from the DOM to trigger auto-unmount
	container.Get("parentNode").Call("removeChild", container)

	// Wait a bit for the MutationObserver to process the change
	time.Sleep(100 * time.Millisecond)

	// Verify the component was auto-unmounted
	if instancesMap.Call("has", js.ValueOf(string(instanceID))).Bool() {
		t.Errorf("Component instance still exists after container removal - auto-unmount failed")
	}

	logutil.Logf("Auto-unmount test completed successfully")
}

func TestAutoUnmountMultipleInstances(t *testing.T) {
	setupMockReactCompat()
	defer teardownMockReactCompat()

	// Initialize React bridge
	if err := InitializeBridge(); err != nil {
		t.Fatalf("Failed to initialize React bridge: %v", err)
	}

	// Register a simple test component
	componentName := "TestMultiAutoUnmountComponent"
	componentJS := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		return js.Global().Get("React").Call("createElement", "div", js.ValueOf(map[string]interface{}{
			"children": "Multi auto-unmount test component",
		}))
	})
	defer componentJS.Release()

	if err := Register(componentName, componentJS.Value); err != nil {
		t.Fatalf("Failed to register component: %v", err)
	}

	// Create parent container
	parentContainer := js.Global().Get("document").Call("createElement", "div")
	parentContainer.Set("id", "parent-container")
	js.Global().Get("document").Get("body").Call("appendChild", parentContainer)

	// Create multiple child containers
	container1 := js.Global().Get("document").Call("createElement", "div")
	container1.Set("id", "child-container-1")
	parentContainer.Call("appendChild", container1)

	container2 := js.Global().Get("document").Call("createElement", "div")
	container2.Set("id", "child-container-2")
	parentContainer.Call("appendChild", container2)

	// Render components in both containers
	instanceID1, err := Render(componentName, map[string]interface{}{}, &RenderOptions{
		ContainerID: "child-container-1",
	})
	if err != nil {
		t.Fatalf("Failed to render component 1: %v", err)
	}

	instanceID2, err := Render(componentName, map[string]interface{}{}, &RenderOptions{
		ContainerID: "child-container-2",
	})
	if err != nil {
		t.Fatalf("Failed to render component 2: %v", err)
	}

	// Verify both components were rendered
	instancesMap := js.Global().Get("window").Get("ReactCompat").Call("getInstancesMap")
	if !instancesMap.Call("has", js.ValueOf(string(instanceID1))).Bool() {
		t.Fatalf("Component instance 1 not found after render")
	}
	if !instancesMap.Call("has", js.ValueOf(string(instanceID2))).Bool() {
		t.Fatalf("Component instance 2 not found after render")
	}

	logutil.Logf("Both components rendered with instance IDs: %s, %s", instanceID1, instanceID2)

	// Remove the parent container to trigger auto-unmount of both children
	parentContainer.Get("parentNode").Call("removeChild", parentContainer)

	// Wait a bit for the MutationObserver to process the change
	time.Sleep(100 * time.Millisecond)

	// Verify both components were auto-unmounted
	if instancesMap.Call("has", js.ValueOf(string(instanceID1))).Bool() {
		t.Errorf("Component instance 1 still exists after parent container removal - auto-unmount failed")
	}
	if instancesMap.Call("has", js.ValueOf(string(instanceID2))).Bool() {
		t.Errorf("Component instance 2 still exists after parent container removal - auto-unmount failed")
	}

	logutil.Logf("Multi auto-unmount test completed successfully")
}

func TestManualUnmountStopsObserving(t *testing.T) {
	setupMockReactCompat()
	defer teardownMockReactCompat()

	// Initialize React bridge
	if err := InitializeBridge(); err != nil {
		t.Fatalf("Failed to initialize React bridge: %v", err)
	}

	// Register a simple test component
	componentName := "TestManualUnmountComponent"
	componentJS := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		return js.Global().Get("React").Call("createElement", "div", js.ValueOf(map[string]interface{}{
			"children": "Manual unmount test component",
		}))
	})
	defer componentJS.Release()

	if err := Register(componentName, componentJS.Value); err != nil {
		t.Fatalf("Failed to register component: %v", err)
	}

	// Create a container element
	container := js.Global().Get("document").Call("createElement", "div")
	container.Set("id", "manual-unmount-test-container")
	js.Global().Get("document").Get("body").Call("appendChild", container)

	// Render the component
	instanceID, err := Render(componentName, map[string]interface{}{}, &RenderOptions{
		ContainerID: "manual-unmount-test-container",
	})
	if err != nil {
		t.Fatalf("Failed to render component: %v", err)
	}

	// Verify the component was rendered
	instancesMap := js.Global().Get("window").Get("ReactCompat").Call("getInstancesMap")
	if !instancesMap.Call("has", js.ValueOf(string(instanceID))).Bool() {
		t.Fatalf("Component instance not found after render")
	}

	logutil.Logf("Component rendered with instance ID: %s", instanceID)

	// Manually unmount the component
	if err := Unmount(instanceID); err != nil {
		t.Fatalf("Failed to manually unmount component: %v", err)
	}

	// Verify the component was unmounted
	if instancesMap.Call("has", js.ValueOf(string(instanceID))).Bool() {
		t.Errorf("Component instance still exists after manual unmount")
	}

	// Now remove the container from the DOM - this should not cause any issues
	// since the component was already manually unmounted
	container.Get("parentNode").Call("removeChild", container)

	// Wait a bit for any potential MutationObserver processing
	time.Sleep(100 * time.Millisecond)

	// The test passes if no errors occurred
	logutil.Logf("Manual unmount test completed successfully")
}
