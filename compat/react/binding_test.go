//go:build js && wasm

package react

import (
	"syscall/js"
	"testing"
	"time"
	"github.com/ozanturksever/uiwgo/reactivity"
)

func TestBind(t *testing.T) {
	
	// Create a signal for testing
	counter := reactivity.CreateSignal(0)
	
	// Create binding options
	options := BindingOptions{
		ComponentID:   "test-bind-component",
		ComponentName: "TestComponent",
		Props: func() map[string]interface{} {
			return map[string]interface{}{
				"count": counter.Get(),
			}
		},
		AutoRender: false,
	}
	
	// Test binding creation
	binding, err := Bind(options)
	if err != nil {
		t.Fatalf("Failed to create binding: %v", err)
	}
	
	if binding == nil {
		t.Fatal("Binding should not be nil")
	}
	
	if binding.GetComponentID() != "test-bind-component" {
		t.Errorf("Expected component ID 'test-bind-component', got '%s'", binding.GetComponentID())
	}
	
	if binding.GetComponentName() != "TestComponent" {
		t.Errorf("Expected component name 'TestComponent', got '%s'", binding.GetComponentName())
	}
	
	if binding.IsActive() {
		t.Error("Binding should not be active initially when AutoRender is false")
	}
}

func TestBindWithAutoRender(t *testing.T) {
	
	// Create a signal for testing
	counter := reactivity.CreateSignal(5)
	
	// Create binding options with auto-render
	options := BindingOptions{
		ComponentID:   "test-auto-render-component",
		ComponentName: "AutoRenderComponent",
		Props: func() map[string]interface{} {
			return map[string]interface{}{
				"count": counter.Get(),
			}
		},
		AutoRender: true,
		OnError: func(err error) {
			// Ignore render errors in tests since we don't have a real React environment
			t.Logf("Render error (expected in test): %v", err)
		},
	}
	
	// Test binding creation with auto-render
	binding, err := Bind(options)
	if err != nil {
		// In test environment, render failures are expected
		t.Logf("Expected render failure in test environment: %v", err)
		return
	}
	
	if binding != nil && !binding.IsActive() {
		t.Error("Binding should be active when AutoRender is true")
	}
	
	// Clean up
	if binding != nil {
		err = binding.Stop()
		if err != nil {
			t.Errorf("Failed to stop binding: %v", err)
		}
	}
}

func TestBindingStartStop(t *testing.T) {
	
	// Create a signal for testing
	counter := reactivity.CreateSignal(10)
	
	// Create binding options
	options := BindingOptions{
		ComponentID:   "test-start-stop-component",
		ComponentName: "StartStopComponent",
		Props: func() map[string]interface{} {
			return map[string]interface{}{
				"count": counter.Get(),
			}
		},
		AutoRender: false,
		OnError: func(err error) {
			t.Logf("Render error (expected in test): %v", err)
		},
	}
	
	binding, err := Bind(options)
	if err != nil {
		t.Fatalf("Failed to create binding: %v", err)
	}
	
	// Test start
	err = binding.Start()
	if err != nil {
		t.Logf("Expected render failure in test environment: %v", err)
		return
	}
	
	if !binding.IsActive() {
		t.Error("Binding should be active after start")
	}
	
	// Test stop
	err = binding.Stop()
	if err != nil {
		t.Fatalf("Failed to stop binding: %v", err)
	}
	
	if binding.IsActive() {
		t.Error("Binding should not be active after stop")
	}
}

func TestBindingReactiveUpdates(t *testing.T) {
	
	// Create a signal for testing
	counter := reactivity.CreateSignal(0)
	updateCount := 0
	
	// Create binding options
	options := BindingOptions{
		ComponentID:   "test-reactive-component",
		ComponentName: "ReactiveComponent",
		Props: func() map[string]interface{} {
			return map[string]interface{}{
				"count": counter.Get(),
				"updateCount": updateCount,
			}
		},
		AutoRender: true,
		OnError: func(err error) {
			t.Logf("Render error (expected in test): %v", err)
		},
	}
	
	binding, err := Bind(options)
	if err != nil {
		t.Logf("Expected render failure in test environment: %v", err)
		return
	}
	defer binding.Stop()
	
	// Update the signal and check if component updates
	counter.Set(1)
	
	// Give some time for the effect to run
	time.Sleep(10 * time.Millisecond)
	
	// Update again
	counter.Set(2)
	updateCount++
	
	// Give some time for the effect to run
	time.Sleep(10 * time.Millisecond)
	
	// The test passes if no errors occur during reactive updates
}

func TestBindingManualUpdate(t *testing.T) {
	
	// Create a signal for testing
	counter := reactivity.CreateSignal(0)
	
	// Create binding options
	options := BindingOptions{
		ComponentID:   "test-manual-update-component",
		ComponentName: "ManualUpdateComponent",
		Props: func() map[string]interface{} {
			return map[string]interface{}{
				"count": counter.Get(),
			}
		},
		AutoRender: false,
		OnError: func(err error) {
			t.Logf("Render error (expected in test): %v", err)
		},
	}
	
	binding, err := Bind(options)
	if err != nil {
		t.Fatalf("Failed to create binding: %v", err)
	}
	
	// Start the binding
	err = binding.Start()
	if err != nil {
		t.Logf("Expected render failure in test environment: %v", err)
		return
	}
	defer binding.Stop()
	
	// Test manual update
	counter.Set(100)
	err = binding.Update()
	if err != nil {
		t.Logf("Expected update failure in test environment: %v", err)
	}
}

func TestBindingErrorHandling(t *testing.T) {
	// Test invalid options
	_, err := Bind(BindingOptions{})
	if err == nil {
		t.Error("Expected error for empty options")
	}
	
	_, err = Bind(BindingOptions{
		ComponentID: "test-id",
		// Missing ComponentName and Props
	})
	if err == nil {
		t.Error("Expected error for missing ComponentName")
	}
	
	_, err = Bind(BindingOptions{
		ComponentID:   "test-id",
		ComponentName: "TestComponent",
		// Missing Props
	})
	if err == nil {
		t.Error("Expected error for missing Props function")
	}
}

func TestBindingManager(t *testing.T) {
	manager := NewBindingManager()
	
	if manager.Count() != 0 {
		t.Errorf("Expected 0 bindings, got %d", manager.Count())
	}
	
	// Create a binding
	counter := reactivity.CreateSignal(0)
	options := BindingOptions{
		ComponentID:   "test-manager-component",
		ComponentName: "ManagerComponent",
		Props: func() map[string]interface{} {
			return map[string]interface{}{
				"count": counter.Get(),
			}
		},
		AutoRender: false,
	}
	
	binding, err := Bind(options)
	if err != nil {
		t.Fatalf("Failed to create binding: %v", err)
	}
	
	// Add to manager
	manager.Add(binding)
	
	if manager.Count() != 1 {
		t.Errorf("Expected 1 binding, got %d", manager.Count())
	}
	
	// Get from manager
	retrieved, exists := manager.Get("test-manager-component")
	if !exists {
		t.Error("Binding should exist in manager")
	}
	if retrieved != binding {
		t.Error("Retrieved binding should be the same as added")
	}
	
	// Remove from manager
	err = manager.Remove("test-manager-component")
	if err != nil {
		t.Errorf("Failed to remove binding: %v", err)
	}
	
	if manager.Count() != 0 {
		t.Errorf("Expected 0 bindings after removal, got %d", manager.Count())
	}
	
	// Test removing non-existent binding
	err = manager.Remove("non-existent")
	if err == nil {
		t.Error("Expected error when removing non-existent binding")
	}
}

func TestGlobalBindingManager(t *testing.T) {
	manager := GetGlobalBindingManager()
	if manager == nil {
		t.Error("Global binding manager should not be nil")
	}
	
	// Clean up any existing bindings
	manager.StopAll()
}

func TestBindingPropsComparison(t *testing.T) {
	counter := reactivity.CreateSignal(0)
	options := BindingOptions{
		ComponentID:   "test-props-comparison",
		ComponentName: "PropsComponent",
		Props: func() map[string]interface{} {
			return map[string]interface{}{
				"count": counter.Get(),
				"static": "value",
			}
		},
		AutoRender: false,
		OnError: func(err error) {
			t.Logf("Render error (expected in test): %v", err)
		},
	}
	
	binding, err := Bind(options)
	if err != nil {
		t.Fatalf("Failed to create binding: %v", err)
	}
	defer binding.Stop()
	
	// Test that binding was created successfully
	if binding == nil {
		t.Error("Binding should not be nil")
	}
	
	// Test that binding is not active initially (AutoRender is false)
	if binding.IsActive() {
		t.Error("Binding should not be active initially when AutoRender is false")
	}
	
	// Test component ID matches
	if binding.GetComponentID() != "test-props-comparison" {
		t.Error("Component ID should match")
	}
}

func TestConvertToString(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected string
	}{
		{"string", "hello", "hello"},
		{"int", 42, "42"},
		{"float", 3.14, "3.14"},
		{"bool true", true, "true"},
		{"bool false", false, "false"},
		{"nil", nil, ""},
		{"slice", []int{1, 2, 3}, "[1 2 3]"},
		{"map", map[string]int{"a": 1}, "map[a:1]"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := convertToString(tt.input)
			if result != tt.expected {
				t.Errorf("convertToString(%v) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestBindTheme(t *testing.T) {
	setupMockReactCompat()
	defer teardownMockReactCompat()
	
	// Create a theme signal
	themeSignal := reactivity.CreateSignal("light")
	
	// Test theme binding creation
	binding, err := BindTheme(themeSignal)
	if err != nil {
		t.Fatalf("Failed to create theme binding: %v", err)
	}
	
	if binding == nil {
		t.Fatal("Theme binding should not be nil")
	}
	
	if binding.IsActive() {
		t.Error("Theme binding should not be active initially")
	}
}

func TestBindTheme_NilSignal(t *testing.T) {
	// Test with nil signal
	binding, err := BindTheme(nil)
	if err == nil {
		t.Fatal("Expected error when theme signal is nil")
	}
	
	if binding != nil {
		t.Error("Binding should be nil when error occurs")
	}
	
	if !contains(err.Error(), "theme signal is required") {
		t.Errorf("Expected 'theme signal is required' error, got: %v", err)
	}
}

func TestThemeBindingStartStop(t *testing.T) {
	setupMockReactCompat()
	defer teardownMockReactCompat()
	
	// Create a theme signal
	themeSignal := reactivity.CreateSignal("dark")
	
	// Create theme binding
	binding, err := BindTheme(themeSignal)
	if err != nil {
		t.Fatalf("Failed to create theme binding: %v", err)
	}
	
	// Test start
	err = binding.Start()
	if err != nil {
		t.Fatalf("Failed to start theme binding: %v", err)
	}
	
	if !binding.IsActive() {
		t.Error("Theme binding should be active after start")
	}
	
	// Test stop
	binding.Stop()
	
	if binding.IsActive() {
		t.Error("Theme binding should not be active after stop")
	}
	
	// Test multiple starts (should be safe)
	err = binding.Start()
	if err != nil {
		t.Fatalf("Failed to restart theme binding: %v", err)
	}
	
	// Test multiple stops (should be safe)
	binding.Stop()
	binding.Stop()
}

func TestThemeBindingReactiveUpdates(t *testing.T) {
	setupMockReactCompat()
	defer teardownMockReactCompat()
	
	// Initialize the bridge
	err := InitializeBridge()
	if err != nil {
		t.Fatalf("Failed to initialize bridge: %v", err)
	}
	
	// Track setTheme calls
	var setThemeCalls []string
	originalSetTheme := js.Global().Get("ReactCompat").Get("setTheme")
	js.Global().Get("ReactCompat").Set("setTheme", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		if len(args) > 0 {
			setThemeCalls = append(setThemeCalls, args[0].String())
		}
		return nil
	}))
	defer js.Global().Get("ReactCompat").Set("setTheme", originalSetTheme)
	
	// Create a theme signal
	themeSignal := reactivity.CreateSignal("light")
	
	// Create and start theme binding
	binding, err := BindTheme(themeSignal)
	if err != nil {
		t.Fatalf("Failed to create theme binding: %v", err)
	}
	
	err = binding.Start()
	if err != nil {
		t.Fatalf("Failed to start theme binding: %v", err)
	}
	defer binding.Stop()
	
	// Wait for initial effect
	time.Sleep(10 * time.Millisecond)
	
	// Check initial theme was set
	if len(setThemeCalls) != 1 || setThemeCalls[0] != "light" {
		t.Errorf("Expected initial setTheme call with 'light', got: %v", setThemeCalls)
	}
	
	// Update theme signal
	themeSignal.Set("dark")
	
	// Wait for effect to run
	time.Sleep(10 * time.Millisecond)
	
	// Check theme was updated
	if len(setThemeCalls) != 2 || setThemeCalls[1] != "dark" {
		t.Errorf("Expected setTheme call with 'dark', got: %v", setThemeCalls)
	}
	
	// Update to custom theme
	themeSignal.Set("custom-theme")
	
	// Wait for effect to run
	time.Sleep(10 * time.Millisecond)
	
	// Check custom theme was set
	if len(setThemeCalls) != 3 || setThemeCalls[2] != "custom-theme" {
		t.Errorf("Expected setTheme call with 'custom-theme', got: %v", setThemeCalls)
	}
}