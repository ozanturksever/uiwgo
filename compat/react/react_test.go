//go:build js && wasm

package react

import (
	"syscall/js"
	"testing"
)

// Mock ReactCompat object for testing
func setupMockReactCompat() {
	// Create individual function mocks
	renderFunc := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		if len(args) < 1 {
			return nil
		}
		componentName := args[0].String()
		instanceID := "test-" + componentName + "-123"
		// Add to mock instances storage
		js.Global().Get("__mockInstances").Call("add", instanceID)
		return instanceID
	})

	updateFunc := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		return len(args) >= 2
	})

	unmountFunc := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		if len(args) >= 1 {
			instanceID := args[0].String()
			// Remove from mock instances storage
			js.Global().Get("__mockInstances").Call("delete", instanceID)
			return true
		}
		return false
	})

	registerFunc := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		return len(args) >= 2
	})

	resolveFunc := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		if len(args) < 1 {
			return nil
		}
		componentName := args[0].String()
		if componentName == "NotFound" {
			return nil
		}
		// Return a simple truthy value instead of complex object
		return js.ValueOf("mock-component")
	})

	diagnosticsFunc := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		// Create a proper JavaScript object
		diagObj := js.Global().Get("Object").New()
		diagObj.Set("instanceCount", 5)
		
		// Create array for registeredComponents
		components := js.Global().Get("Array").New()
		components.SetIndex(0, "Button")
		components.SetIndex(1, "Input")
		diagObj.Set("registeredComponents", components)
		
		return diagObj
	})

	setThemeFunc := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		if len(args) < 1 {
			return "Theme parameter required"
		}
		theme := args[0].String()
		if theme == "invalid" {
			return "Invalid theme"
		}
		// Success case - return undefined
		return js.Undefined()
	})

	// Create a global instances storage for the mock
	if js.Global().Get("__mockInstances").IsUndefined() {
		js.Global().Set("__mockInstances", js.Global().Get("Set").New())
	}
	
	getInstancesMapFunc := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		// Return a mock instances map with realistic behavior
		instancesMap := js.Global().Get("Object").New()
		hasFunc := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			if len(args) < 1 {
				return false
			}
			instanceID := args[0].String()
			return js.Global().Get("__mockInstances").Call("has", instanceID).Bool()
		})
		instancesMap.Set("has", hasFunc)
		return instancesMap
	})

	// Create the mock bridge object step by step
	mockBridge := js.Global().Get("Object").New()
	mockBridge.Set("renderComponent", renderFunc)
	mockBridge.Set("updateComponent", updateFunc)
	mockBridge.Set("unmountComponent", unmountFunc)
	mockBridge.Set("register", registerFunc)
	mockBridge.Set("resolve", resolveFunc)
	mockBridge.Set("getDiagnostics", diagnosticsFunc)
	mockBridge.Set("setTheme", setThemeFunc)
	mockBridge.Set("getInstancesMap", getInstancesMapFunc)

	// Set up MutationObserver mock for auto-unmount
	if js.Global().Get("__mockObserver").IsUndefined() {
		// Create a simple mock that simulates auto-unmount when containers are removed
		observerCallback := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			// This would normally be called by MutationObserver
			// For testing, we'll trigger it manually when containers are removed
			return nil
		})
		js.Global().Set("__mockObserver", observerCallback)
		
		// Override removeChild to trigger auto-unmount
		originalRemoveChild := js.Global().Get("Node").Get("prototype").Get("removeChild")
		js.Global().Get("Node").Get("prototype").Set("removeChild", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			if len(args) > 0 {
				child := args[0]
				// Check if this container has any rendered components
				if !child.Get("id").IsUndefined() {
					containerID := child.Get("id").String()
					// Remove any instances associated with this container
					mockInstances := js.Global().Get("__mockInstances")
					// For simplicity, remove all instances when any container is removed
					// In a real implementation, this would be more sophisticated
					if containerID == "auto-unmount-test-container" || containerID == "parent-container" {
						mockInstances.Call("clear")
					}
				}
			}
			// Call original removeChild
			return originalRemoveChild.Call("call", this, args[0])
		}))
	}
	
	// Set up the mock on the global window object
	js.Global().Get("window").Set("ReactCompat", mockBridge)
}

func teardownMockReactCompat() {
	// Clean up the mock
	js.Global().Set("ReactCompat", js.Undefined())
	// Reset global bridge
	globalBridge = nil
}

func TestNewReactBridge(t *testing.T) {
	setupMockReactCompat()
	defer teardownMockReactCompat()

	bridge, err := NewReactBridge()
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if bridge == nil {
		t.Fatal("Expected bridge to be non-nil")
	}

	if !bridge.bridge.Truthy() {
		t.Fatal("Expected bridge.bridge to be truthy")
	}
}

func TestNewReactBridge_NoReactCompat(t *testing.T) {
	// Ensure ReactCompat is not available
	js.Global().Set("ReactCompat", js.Undefined())
	defer teardownMockReactCompat()

	_, err := NewReactBridge()
	if err == nil {
		t.Fatal("Expected error when ReactCompat is not available")
	}

	expected := "ReactCompat not found on window object"
	if err.Error() != expected {
		t.Fatalf("Expected error %q, got %q", expected, err.Error())
	}
}

func TestInitializeBridge(t *testing.T) {
	setupMockReactCompat()
	defer teardownMockReactCompat()

	// Ensure global bridge is nil
	globalBridge = nil

	err := InitializeBridge()
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if globalBridge == nil {
		t.Fatal("Expected globalBridge to be initialized")
	}

	// Test that calling InitializeBridge again doesn't error
	err = InitializeBridge()
	if err != nil {
		t.Fatalf("Expected no error on second initialization, got %v", err)
	}
}

func TestGetBridge(t *testing.T) {
	setupMockReactCompat()
	defer teardownMockReactCompat()

	// Test when bridge is not initialized
	globalBridge = nil
	_, err := GetBridge()
	if err == nil {
		t.Fatal("Expected error when bridge is not initialized")
	}

	// Initialize bridge
	err = InitializeBridge()
	if err != nil {
		t.Fatalf("Failed to initialize bridge: %v", err)
	}

	// Test when bridge is initialized
	bridge, err := GetBridge()
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if bridge == globalBridge {
		t.Log("GetBridge returned the correct global bridge instance")
	} else {
		t.Fatal("GetBridge did not return the global bridge instance")
	}
}

func TestRender(t *testing.T) {
	setupMockReactCompat()
	defer teardownMockReactCompat()

	bridge, err := NewReactBridge()
	if err != nil {
		t.Fatalf("Failed to create bridge: %v", err)
	}

	props := Props{
		"text":    "Hello World",
		"onClick": map[string]interface{}{"__cbid": "callback-123"},
	}

	options := &RenderOptions{
		ContainerID: "root",
		Replace:     true,
	}

	componentID, err := bridge.Render("Button", props, options)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	expected := ComponentID("test-Button-123")
	if componentID != expected {
		t.Fatalf("Expected componentID %q, got %q", expected, componentID)
	}
}

func TestRender_WithoutOptions(t *testing.T) {
	setupMockReactCompat()
	defer teardownMockReactCompat()

	bridge, err := NewReactBridge()
	if err != nil {
		t.Fatalf("Failed to create bridge: %v", err)
	}

	props := Props{"text": "Hello"}

	componentID, err := bridge.Render("Input", props, nil)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	expected := ComponentID("test-Input-123")
	if componentID != expected {
		t.Fatalf("Expected componentID %q, got %q", expected, componentID)
	}
}

func TestRender_InvalidProps(t *testing.T) {
	setupMockReactCompat()
	defer teardownMockReactCompat()

	bridge, err := NewReactBridge()
	if err != nil {
		t.Fatalf("Failed to create bridge: %v", err)
	}

	// Create props that can't be serialized to JSON
	props := Props{
		"invalid": make(chan int), // channels can't be serialized to JSON
	}

	_, err = bridge.Render("Button", props, nil)
	if err == nil {
		t.Fatal("Expected error for invalid props")
	}

	if !contains(err.Error(), "failed to serialize props") {
		t.Fatalf("Expected serialization error, got %v", err)
	}
}

func TestUpdate(t *testing.T) {
	setupMockReactCompat()
	defer teardownMockReactCompat()

	bridge, err := NewReactBridge()
	if err != nil {
		t.Fatalf("Failed to create bridge: %v", err)
	}

	props := Props{"text": "Updated Text"}
	componentID := ComponentID("test-component-123")

	err = bridge.Update(componentID, props)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
}

func TestUpdate_InvalidProps(t *testing.T) {
	setupMockReactCompat()
	defer teardownMockReactCompat()

	bridge, err := NewReactBridge()
	if err != nil {
		t.Fatalf("Failed to create bridge: %v", err)
	}

	// Create props that can't be serialized to JSON
	props := Props{
		"invalid": make(chan int),
	}
	componentID := ComponentID("test-component-123")

	err = bridge.Update(componentID, props)
	if err == nil {
		t.Fatal("Expected error for invalid props")
	}

	if !contains(err.Error(), "failed to serialize props") {
		t.Fatalf("Expected serialization error, got %v", err)
	}
}

func TestUnmount(t *testing.T) {
	setupMockReactCompat()
	defer teardownMockReactCompat()

	bridge, err := NewReactBridge()
	if err != nil {
		t.Fatalf("Failed to create bridge: %v", err)
	}

	componentID := ComponentID("test-component-123")

	err = bridge.Unmount(componentID)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
}

func TestRegister(t *testing.T) {
	setupMockReactCompat()
	defer teardownMockReactCompat()

	bridge, err := NewReactBridge()
	if err != nil {
		t.Fatalf("Failed to create bridge: %v", err)
	}

	// Create a mock React component
	component := js.ValueOf(map[string]interface{}{"name": "TestComponent"})

	err = bridge.Register("TestComponent", component)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
}

func TestResolve(t *testing.T) {
	setupMockReactCompat()
	defer teardownMockReactCompat()

	bridge, err := NewReactBridge()
	if err != nil {
		t.Fatalf("Failed to create bridge: %v", err)
	}

	// Test resolving an existing component
	component, err := bridge.Resolve("Button")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if !component.Truthy() {
		t.Fatal("Expected component to be truthy")
	}

	// Test resolving a non-existing component
	_, err = bridge.Resolve("NotFound")
	if err == nil {
		t.Fatal("Expected error for non-existing component")
	}

	if !contains(err.Error(), "component NotFound not found") {
		t.Fatalf("Expected 'not found' error, got %v", err)
	}
}

func TestGetDiagnostics(t *testing.T) {
	setupMockReactCompat()
	defer teardownMockReactCompat()

	bridge, err := NewReactBridge()
	if err != nil {
		t.Fatalf("Failed to create bridge: %v", err)
	}

	diagnostics, err := bridge.GetDiagnostics()
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if diagnostics == nil {
		t.Fatal("Expected diagnostics to be non-nil")
	}

	// Check expected fields
	if instanceCount, ok := diagnostics["instanceCount"]; !ok {
		t.Fatal("Expected instanceCount in diagnostics")
	} else if instanceCount != float64(5) { // JSON numbers are float64
		t.Fatalf("Expected instanceCount to be 5, got %v", instanceCount)
	}

	if registeredComponents, ok := diagnostics["registeredComponents"]; !ok {
		t.Fatal("Expected registeredComponents in diagnostics")
	} else {
		components, ok := registeredComponents.([]interface{})
		if !ok {
			t.Fatal("Expected registeredComponents to be an array")
		}
		if len(components) != 2 {
			t.Fatalf("Expected 2 registered components, got %d", len(components))
		}
	}
}

// Test convenience functions
func TestConvenienceFunctions(t *testing.T) {
	setupMockReactCompat()
	defer teardownMockReactCompat()

	// Initialize global bridge
	err := InitializeBridge()
	if err != nil {
		t.Fatalf("Failed to initialize bridge: %v", err)
	}

	// Test Render convenience function
	props := Props{"text": "Hello"}
	componentID, err := Render("Button", props, nil)
	if err != nil {
		t.Fatalf("Render convenience function failed: %v", err)
	}
	if componentID == "" {
		t.Fatal("Expected non-empty component ID")
	}

	// Test Update convenience function
	updatedProps := Props{"text": "Updated"}
	err = Update(componentID, updatedProps)
	if err != nil {
		t.Fatalf("Update convenience function failed: %v", err)
	}

	// Test Register convenience function
	component := js.ValueOf(map[string]interface{}{"name": "TestComponent"})
	err = Register("TestComponent", component)
	if err != nil {
		t.Fatalf("Register convenience function failed: %v", err)
	}

	// Test Resolve convenience function
	resolvedComponent, err := Resolve("TestComponent")
	if err != nil {
		t.Fatalf("Resolve convenience function failed: %v", err)
	}
	if !resolvedComponent.Truthy() {
		t.Fatal("Expected resolved component to be truthy")
	}

	// Test GetDiagnostics convenience function
	diagnostics, err := GetDiagnostics()
	if err != nil {
		t.Fatalf("GetDiagnostics convenience function failed: %v", err)
	}
	if diagnostics == nil {
		t.Fatal("Expected diagnostics to be non-nil")
	}

	// Test Unmount convenience function
	err = Unmount(componentID)
	if err != nil {
		t.Fatalf("Unmount convenience function failed: %v", err)
	}
}

func TestConvenienceFunctions_NotInitialized(t *testing.T) {
	// Ensure global bridge is not initialized
	globalBridge = nil

	// Test that convenience functions return appropriate errors
	_, err := Render("Button", Props{}, nil)
	if err == nil {
		t.Fatal("Expected error when bridge not initialized")
	}

	err = Update("test-id", Props{})
	if err == nil {
		t.Fatal("Expected error when bridge not initialized")
	}

	err = Unmount("test-id")
	if err == nil {
		t.Fatal("Expected error when bridge not initialized")
	}

	err = Register("Test", js.Undefined())
	if err == nil {
		t.Fatal("Expected error when bridge not initialized")
	}

	_, err = Resolve("Test")
	if err == nil {
		t.Fatal("Expected error when bridge not initialized")
	}

	_, err = GetDiagnostics()
	if err == nil {
		t.Fatal("Expected error when bridge not initialized")
	}

	err = SetTheme("dark")
	if err == nil {
		t.Fatal("Expected error when bridge not initialized")
	}
}

func TestSetTheme(t *testing.T) {
	setupMockReactCompat()
	defer teardownMockReactCompat()

	bridge, err := NewReactBridge()
	if err != nil {
		t.Fatalf("Failed to create bridge: %v", err)
	}

	// Test successful theme setting
	err = bridge.SetTheme("dark")
	if err != nil {
		t.Fatalf("SetTheme failed: %v", err)
	}

	err = bridge.SetTheme("light")
	if err != nil {
		t.Fatalf("SetTheme failed: %v", err)
	}

	// Test error case
	err = bridge.SetTheme("invalid")
	if err == nil {
		t.Fatal("Expected error for invalid theme")
	}
	if !contains(err.Error(), "Invalid theme") {
		t.Fatalf("Expected 'Invalid theme' error, got: %v", err)
	}
}

func TestSetTheme_NoSetThemeFunction(t *testing.T) {
	setupMockReactCompat()
	defer teardownMockReactCompat()

	// Create a bridge without setTheme function
	mockBridge := js.Global().Get("Object").New()
	js.Global().Set("ReactCompat", mockBridge)

	bridge, err := NewReactBridge()
	if err != nil {
		t.Fatalf("Failed to create bridge: %v", err)
	}

	err = bridge.SetTheme("dark")
	if err == nil {
		t.Fatal("Expected error when setTheme function not found")
	}
	if !contains(err.Error(), "setTheme function not found") {
		t.Fatalf("Expected 'setTheme function not found' error, got: %v", err)
	}
}

func TestSetTheme_ConvenienceFunction(t *testing.T) {
	setupMockReactCompat()
	defer teardownMockReactCompat()

	err := InitializeBridge()
	if err != nil {
		t.Fatalf("Failed to initialize bridge: %v", err)
	}

	// Test convenience function
	err = SetTheme("dark")
	if err != nil {
		t.Fatalf("SetTheme convenience function failed: %v", err)
	}

	err = SetTheme("light")
	if err != nil {
		t.Fatalf("SetTheme convenience function failed: %v", err)
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 || 
		(len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || 
			containsHelper(s, substr))))
}

func containsHelper(s, substr string) bool {
	for i := 1; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}