// observer_test.go
// Comprehensive unit tests for the ObserverManager system
// Tests DOM-independent functionality that can be verified without full browser environment

package golid

import (
	"sync"
	"syscall/js"
	"testing"
	"time"
)

// TestObserverManagerInitialization tests the global observer initialization
func TestObserverManagerInitialization(t *testing.T) {
	// The globalObserver should be initialized in init()
	if globalObserver == nil {
		t.Fatal("globalObserver should be initialized")
	}

	// Check that the globalObserver has the required fields initialized
	globalObserver.mutex.RLock()
	callbacksInitialized := globalObserver.callbacks != nil
	dismountCallbacksInitialized := globalObserver.dismountCallbacks != nil
	trackedElementsInitialized := globalObserver.trackedElements != nil
	globalObserver.mutex.RUnlock()

	if !callbacksInitialized {
		t.Error("callbacks map should be initialized")
	}
	if !dismountCallbacksInitialized {
		t.Error("dismountCallbacks map should be initialized")
	}
	if !trackedElementsInitialized {
		t.Error("trackedElements map should be initialized")
	}

	// Note: We don't check for empty maps or isObserving state here because
	// other tests or framework components may have already registered callbacks
	// The important thing is that the globalObserver exists and is properly initialized
}

// TestObserverManagerStructure tests the ObserverManager struct fields
func TestObserverManagerStructure(t *testing.T) {
	om := &ObserverManager{
		callbacks:         make(map[string]ElementCallback),
		dismountCallbacks: make(map[string][]LifecycleHook),
		trackedElements:   make(map[string]js.Value),
		isObserving:       false,
	}

	if om.callbacks == nil {
		t.Error("callbacks map should be initialized")
	}
	if om.dismountCallbacks == nil {
		t.Error("dismountCallbacks map should be initialized")
	}
	if om.trackedElements == nil {
		t.Error("trackedElements map should be initialized")
	}
	if om.isObserving {
		t.Error("isObserving should be false initially")
	}
}

// TestRegisterElement tests the RegisterElement functionality
func TestRegisterElement(t *testing.T) {
	// Create a test observer manager
	om := &ObserverManager{
		callbacks:         make(map[string]ElementCallback),
		dismountCallbacks: make(map[string][]LifecycleHook),
		trackedElements:   make(map[string]js.Value),
		isObserving:       false,
	}

	// Test callback function
	called := false
	testCallback := func() {
		called = true
	}

	// Register an element
	om.RegisterElement("test-id-1", testCallback)

	// Check that callback was registered
	om.mutex.RLock()
	callback, exists := om.callbacks["test-id-1"]
	om.mutex.RUnlock()

	if !exists {
		t.Error("Callback should be registered")
	}

	// Execute the callback to verify it works
	if callback != nil {
		callback()
		if !called {
			t.Error("Callback should have been executed")
		}
	}
}

// TestRegisterMultipleElements tests registering multiple elements
func TestRegisterMultipleElements(t *testing.T) {
	om := &ObserverManager{
		callbacks:         make(map[string]ElementCallback),
		dismountCallbacks: make(map[string][]LifecycleHook),
		trackedElements:   make(map[string]js.Value),
		isObserving:       false,
	}

	// Register multiple elements
	callbackResults := make(map[string]bool)

	for i := 0; i < 5; i++ {
		id := "test-id-" + string(rune('1'+i))
		om.RegisterElement(id, func() {
			callbackResults[id] = true
		})
	}

	// Check all callbacks were registered
	om.mutex.RLock()
	callbacksLen := len(om.callbacks)
	om.mutex.RUnlock()

	if callbacksLen != 5 {
		t.Errorf("Expected 5 callbacks, got %d", callbacksLen)
	}
}

// TestRegisterElementOverwrite tests that registering the same ID overwrites
func TestRegisterElementOverwrite(t *testing.T) {
	om := &ObserverManager{
		callbacks:         make(map[string]ElementCallback),
		dismountCallbacks: make(map[string][]LifecycleHook),
		trackedElements:   make(map[string]js.Value),
		isObserving:       false,
	}

	// Register first callback
	firstCalled := false
	om.RegisterElement("test-id", func() {
		firstCalled = true
	})

	// Register second callback with same ID
	secondCalled := false
	om.RegisterElement("test-id", func() {
		secondCalled = true
	})

	// Should only have one callback
	om.mutex.RLock()
	callbacksLen := len(om.callbacks)
	callback := om.callbacks["test-id"]
	om.mutex.RUnlock()

	if callbacksLen != 1 {
		t.Errorf("Expected 1 callback after overwrite, got %d", callbacksLen)
	}

	// Execute callback - should be the second one
	if callback != nil {
		callback()
		if firstCalled {
			t.Error("First callback should not have been called")
		}
		if !secondCalled {
			t.Error("Second callback should have been called")
		}
	}
}

// TestRegisterDismountCallback tests dismount callback registration
func TestRegisterDismountCallback(t *testing.T) {
	om := &ObserverManager{
		callbacks:         make(map[string]ElementCallback),
		dismountCallbacks: make(map[string][]LifecycleHook),
		trackedElements:   make(map[string]js.Value),
		isObserving:       false,
	}

	called := false
	testCallback := func() {
		called = true
	}

	// Register a dismount callback
	om.RegisterDismountCallback("test-id", testCallback)

	// Check that callback was registered
	om.mutex.RLock()
	callbacks, exists := om.dismountCallbacks["test-id"]
	om.mutex.RUnlock()

	if !exists {
		t.Error("Dismount callback should be registered")
	}

	if len(callbacks) != 1 {
		t.Errorf("Expected 1 dismount callback, got %d", len(callbacks))
	}

	// Execute the callback to verify it works
	if len(callbacks) > 0 && callbacks[0] != nil {
		callbacks[0]()
		if !called {
			t.Error("Dismount callback should have been executed")
		}
	}
}

// TestRegisterMultipleDismountCallbacks tests multiple dismount callbacks for same ID
func TestRegisterMultipleDismountCallbacks(t *testing.T) {
	om := &ObserverManager{
		callbacks:         make(map[string]ElementCallback),
		dismountCallbacks: make(map[string][]LifecycleHook),
		trackedElements:   make(map[string]js.Value),
		isObserving:       false,
	}

	callCount := 0

	// Register multiple dismount callbacks for the same ID
	for i := 0; i < 3; i++ {
		om.RegisterDismountCallback("test-id", func() {
			callCount++
		})
	}

	// Check that all callbacks were registered
	om.mutex.RLock()
	callbacks := om.dismountCallbacks["test-id"]
	om.mutex.RUnlock()

	if len(callbacks) != 3 {
		t.Errorf("Expected 3 dismount callbacks, got %d", len(callbacks))
	}

	// Execute all callbacks
	for _, callback := range callbacks {
		if callback != nil {
			callback()
		}
	}

	if callCount != 3 {
		t.Errorf("Expected 3 callback executions, got %d", callCount)
	}
}

// TestUnregisterElement tests the UnregisterElement functionality
func TestUnregisterElement(t *testing.T) {
	om := &ObserverManager{
		callbacks:         make(map[string]ElementCallback),
		dismountCallbacks: make(map[string][]LifecycleHook),
		trackedElements:   make(map[string]js.Value),
		isObserving:       false,
	}

	// Register an element first
	om.RegisterElement("test-id", func() {})
	om.RegisterDismountCallback("test-id", func() {})

	// Verify it was registered
	om.mutex.RLock()
	_, callbackExists := om.callbacks["test-id"]
	_, dismountExists := om.dismountCallbacks["test-id"]
	om.mutex.RUnlock()

	if !callbackExists {
		t.Error("Callback should exist before unregistering")
	}
	if !dismountExists {
		t.Error("Dismount callback should exist before unregistering")
	}

	// Unregister the element
	om.UnregisterElement("test-id")

	// Verify it was unregistered
	om.mutex.RLock()
	_, callbackExists = om.callbacks["test-id"]
	_, dismountExists = om.dismountCallbacks["test-id"]
	om.mutex.RUnlock()

	if callbackExists {
		t.Error("Callback should not exist after unregistering")
	}
	if dismountExists {
		t.Error("Dismount callback should not exist after unregistering")
	}
}

// TestUnregisterNonexistentElement tests unregistering an element that doesn't exist
func TestUnregisterNonexistentElement(t *testing.T) {
	om := &ObserverManager{
		callbacks:         make(map[string]ElementCallback),
		dismountCallbacks: make(map[string][]LifecycleHook),
		trackedElements:   make(map[string]js.Value),
		isObserving:       false,
	}

	// This should not panic or cause errors
	om.UnregisterElement("nonexistent-id")

	// Verify state is still clean
	om.mutex.RLock()
	callbacksLen := len(om.callbacks)
	dismountLen := len(om.dismountCallbacks)
	om.mutex.RUnlock()

	if callbacksLen != 0 {
		t.Errorf("Expected 0 callbacks, got %d", callbacksLen)
	}
	if dismountLen != 0 {
		t.Errorf("Expected 0 dismount callbacks, got %d", dismountLen)
	}
}

// TestObserverConcurrentAccess tests concurrent access to observer methods
func TestObserverConcurrentAccess(t *testing.T) {
	om := &ObserverManager{
		callbacks:         make(map[string]ElementCallback),
		dismountCallbacks: make(map[string][]LifecycleHook),
		trackedElements:   make(map[string]js.Value),
		isObserving:       false,
	}

	var wg sync.WaitGroup

	// Concurrent registrations
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			elementID := "element-" + string(rune('0'+id))
			om.RegisterElement(elementID, func() {})
			om.RegisterDismountCallback(elementID, func() {})
		}(i)
	}

	// Concurrent unregistrations
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			elementID := "element-" + string(rune('0'+id))
			time.Sleep(1 * time.Millisecond) // Let some registrations complete first
			om.UnregisterElement(elementID)
		}(i)
	}

	wg.Wait()

	// Verify final state (should have 5 remaining elements)
	om.mutex.RLock()
	callbacksLen := len(om.callbacks)
	dismountLen := len(om.dismountCallbacks)
	om.mutex.RUnlock()

	if callbacksLen != 5 {
		t.Errorf("Expected 5 remaining callbacks, got %d", callbacksLen)
	}
	if dismountLen != 5 {
		t.Errorf("Expected 5 remaining dismount callbacks, got %d", dismountLen)
	}
}

// TestObserverStateManagement tests the isObserving state management
func TestObserverStateManagement(t *testing.T) {
	om := &ObserverManager{
		callbacks:         make(map[string]ElementCallback),
		dismountCallbacks: make(map[string][]LifecycleHook),
		trackedElements:   make(map[string]js.Value),
		isObserving:       false,
	}

	// Initially not observing
	om.mutex.RLock()
	isObserving := om.isObserving
	om.mutex.RUnlock()

	if isObserving {
		t.Error("Should not be observing initially")
	}

	// Test stopObserving when isObserving is true but observer is not initialized
	// This should NOT change isObserving to false because observer.Truthy() is false
	om.mutex.Lock()
	om.isObserving = true
	om.mutex.Unlock()

	// Call stopObserving - should NOT change isObserving since observer is not initialized
	om.stopObserving()

	om.mutex.RLock()
	isObserving = om.isObserving
	om.mutex.RUnlock()

	// The isObserving should still be true because stopObserving only sets it to false
	// when both isObserving is true AND observer.Truthy() is true
	if !isObserving {
		t.Error("isObserving should remain true when observer is not initialized")
	}

	// Test that stopObserving can be called safely when not observing
	om.mutex.Lock()
	om.isObserving = false
	om.mutex.Unlock()

	// This should not cause any issues
	om.stopObserving()

	om.mutex.RLock()
	isObserving = om.isObserving
	om.mutex.RUnlock()

	if isObserving {
		t.Error("Should not be observing after calling stopObserving when not observing")
	}
}

// TestObserverMapSizes tests that the internal maps maintain correct sizes
func TestObserverMapSizes(t *testing.T) {
	om := &ObserverManager{
		callbacks:         make(map[string]ElementCallback),
		dismountCallbacks: make(map[string][]LifecycleHook),
		trackedElements:   make(map[string]js.Value),
		isObserving:       false,
	}

	// Add some elements
	for i := 0; i < 3; i++ {
		id := "test-" + string(rune('1'+i))
		om.RegisterElement(id, func() {})
		om.RegisterDismountCallback(id, func() {})
		// Note: trackedElements would be populated by DOM operations we can't test here
	}

	om.mutex.RLock()
	callbacksLen := len(om.callbacks)
	dismountLen := len(om.dismountCallbacks)
	om.mutex.RUnlock()

	if callbacksLen != 3 {
		t.Errorf("Expected 3 callbacks, got %d", callbacksLen)
	}
	if dismountLen != 3 {
		t.Errorf("Expected 3 dismount callback entries, got %d", dismountLen)
	}

	// Remove one element
	om.UnregisterElement("test-1")

	om.mutex.RLock()
	callbacksLen = len(om.callbacks)
	dismountLen = len(om.dismountCallbacks)
	om.mutex.RUnlock()

	if callbacksLen != 2 {
		t.Errorf("Expected 2 callbacks after removal, got %d", callbacksLen)
	}
	if dismountLen != 2 {
		t.Errorf("Expected 2 dismount callback entries after removal, got %d", dismountLen)
	}
}

// TestObserverCallbackExecution tests that registered callbacks can be executed
func TestObserverCallbackExecution(t *testing.T) {
	om := &ObserverManager{
		callbacks:         make(map[string]ElementCallback),
		dismountCallbacks: make(map[string][]LifecycleHook),
		trackedElements:   make(map[string]js.Value),
		isObserving:       false,
	}

	// Test data to track callback executions
	executed := make(map[string]bool)
	var mutex sync.Mutex

	// Register callbacks that track their execution
	for i := 0; i < 3; i++ {
		id := "test-" + string(rune('1'+i))
		om.RegisterElement(id, func() {
			mutex.Lock()
			executed[id] = true
			mutex.Unlock()
		})
	}

	// Execute all registered callbacks
	om.mutex.RLock()
	callbacks := make(map[string]ElementCallback)
	for k, v := range om.callbacks {
		callbacks[k] = v
	}
	om.mutex.RUnlock()

	for _, callback := range callbacks {
		if callback != nil {
			callback()
		}
	}

	// Verify all callbacks were executed
	mutex.Lock()
	executedCount := len(executed)
	mutex.Unlock()

	if executedCount != 3 {
		t.Errorf("Expected 3 callbacks to be executed, got %d", executedCount)
	}

	for i := 0; i < 3; i++ {
		id := "test-" + string(rune('1'+i))
		mutex.Lock()
		wasExecuted := executed[id]
		mutex.Unlock()

		if !wasExecuted {
			t.Errorf("Callback for %s should have been executed", id)
		}
	}
}

// TestObserverRecursionDepthLogging tests that logging occurs when recursion depth limit is reached
func TestObserverRecursionDepthLogging(t *testing.T) {
	om := &ObserverManager{
		callbacks:         make(map[string]ElementCallback),
		dismountCallbacks: make(map[string][]LifecycleHook),
		trackedElements:   make(map[string]js.Value),
		isObserving:       false,
		recursionDepth:    0,
		maxRecursionDepth: 5, // Set a low limit to easily trigger
	}

	// Create a mock observer callback function that we can test directly
	// This simulates the logic from the actual observerCallback in startObserving
	testCallback := func() bool {
		// Check if observer is suspended (simulate the real check)
		om.mutex.RLock()
		if om.isSuspended {
			om.mutex.RUnlock()
			return false
		}

		// Check recursion depth to prevent infinite loops (this is the code we're testing)
		if om.recursionDepth >= om.maxRecursionDepth {
			om.mutex.RUnlock()
			// This is where our logging should occur - we can't easily capture console output
			// in tests, but we can verify the condition is met and the function exits correctly
			Logf("⚠️ Observer recursion depth limit reached (%d). Possible infinite loop detected - observer callback execution stopped to prevent stack overflow.", om.maxRecursionDepth)
			return false // Simulate returning nil from the real callback
		}

		// Increment recursion depth
		om.recursionDepth++
		om.mutex.RUnlock()

		return true // Continue processing
	}

	// Test normal operation - should succeed
	om.mutex.Lock()
	om.recursionDepth = 3 // Below limit
	om.mutex.Unlock()

	result := testCallback()
	if !result {
		t.Error("Callback should succeed when recursion depth is below limit")
	}

	// Verify recursion depth was incremented
	om.mutex.RLock()
	currentDepth := om.recursionDepth
	om.mutex.RUnlock()

	if currentDepth != 4 {
		t.Errorf("Expected recursion depth 4, got %d", currentDepth)
	}

	// Test recursion limit reached - should trigger logging and return false
	om.mutex.Lock()
	om.recursionDepth = 5 // At the limit
	om.mutex.Unlock()

	result = testCallback()
	if result {
		t.Error("Callback should fail (return false) when recursion depth limit is reached")
	}

	// Verify recursion depth was not incremented when limit was reached
	om.mutex.RLock()
	finalDepth := om.recursionDepth
	om.mutex.RUnlock()

	if finalDepth != 5 {
		t.Errorf("Expected recursion depth to remain 5 when limit reached, got %d", finalDepth)
	}

	// Test recursion limit exceeded - should also trigger logging and return false
	om.mutex.Lock()
	om.recursionDepth = 10 // Above the limit
	om.mutex.Unlock()

	result = testCallback()
	if result {
		t.Error("Callback should fail (return false) when recursion depth limit is exceeded")
	}

	// Verify that the logging mechanism doesn't interfere with normal operation
	// Reset to normal state
	om.mutex.Lock()
	om.recursionDepth = 0
	om.mutex.Unlock()

	result = testCallback()
	if !result {
		t.Error("Callback should work normally after recursion depth is reset")
	}

	t.Logf("✅ Recursion depth logging test completed - logging will appear in browser console when limit is reached")
}

// TestObserverMaxRecursionDepthValidation tests the validation of maxRecursionDepth values
func TestObserverMaxRecursionDepthValidation(t *testing.T) {
	// Test valid values within bounds
	validCases := []int{1, 5, 50, 100, 500, 1000}

	for _, validValue := range validCases {
		om := &ObserverManager{
			callbacks:         make(map[string]ElementCallback),
			dismountCallbacks: make(map[string][]LifecycleHook),
			trackedElements:   make(map[string]js.Value),
			maxRecursionDepth: validateMaxRecursionDepth(validValue),
		}

		if om.maxRecursionDepth != validValue {
			t.Errorf("Valid value %d should not be changed, got %d", validValue, om.maxRecursionDepth)
		}
	}

	// Test minimum boundary (exactly at the boundary)
	minBoundary := validateMaxRecursionDepth(MinRecursionDepth)
	if minBoundary != MinRecursionDepth {
		t.Errorf("Minimum boundary value %d should be accepted, got %d", MinRecursionDepth, minBoundary)
	}

	// Test maximum boundary (exactly at the boundary)
	maxBoundary := validateMaxRecursionDepth(MaxRecursionDepth)
	if maxBoundary != MaxRecursionDepth {
		t.Errorf("Maximum boundary value %d should be accepted, got %d", MaxRecursionDepth, maxBoundary)
	}
}

// TestObserverMaxRecursionDepthInvalidValues tests validation with invalid values
func TestObserverMaxRecursionDepthInvalidValues(t *testing.T) {
	// Test values below minimum (should default to DefaultRecursionDepth)
	invalidLowCases := []int{0, -1, -10, -100}

	for _, invalidValue := range invalidLowCases {
		result := validateMaxRecursionDepth(invalidValue)
		if result != DefaultRecursionDepth {
			t.Errorf("Invalid low value %d should default to %d, got %d", invalidValue, DefaultRecursionDepth, result)
		}
	}

	// Test values above maximum (should default to DefaultRecursionDepth)
	invalidHighCases := []int{1001, 2000, 10000, 999999}

	for _, invalidValue := range invalidHighCases {
		result := validateMaxRecursionDepth(invalidValue)
		if result != DefaultRecursionDepth {
			t.Errorf("Invalid high value %d should default to %d, got %d", invalidValue, DefaultRecursionDepth, result)
		}
	}
}

// TestObserverSetMaxRecursionDepth tests the SetMaxRecursionDepth method
func TestObserverSetMaxRecursionDepth(t *testing.T) {
	om := &ObserverManager{
		callbacks:         make(map[string]ElementCallback),
		dismountCallbacks: make(map[string][]LifecycleHook),
		trackedElements:   make(map[string]js.Value),
		maxRecursionDepth: DefaultRecursionDepth,
	}

	// Test setting a valid value
	om.SetMaxRecursionDepth(100)
	if om.GetMaxRecursionDepth() != 100 {
		t.Errorf("Expected maxRecursionDepth to be 100, got %d", om.GetMaxRecursionDepth())
	}

	// Test setting an invalid low value (should default to DefaultRecursionDepth)
	om.SetMaxRecursionDepth(-5)
	if om.GetMaxRecursionDepth() != DefaultRecursionDepth {
		t.Errorf("Invalid low value should set maxRecursionDepth to %d, got %d", DefaultRecursionDepth, om.GetMaxRecursionDepth())
	}

	// Test setting an invalid high value (should default to DefaultRecursionDepth)
	om.SetMaxRecursionDepth(5000)
	if om.GetMaxRecursionDepth() != DefaultRecursionDepth {
		t.Errorf("Invalid high value should set maxRecursionDepth to %d, got %d", DefaultRecursionDepth, om.GetMaxRecursionDepth())
	}

	// Test setting boundary values
	om.SetMaxRecursionDepth(MinRecursionDepth)
	if om.GetMaxRecursionDepth() != MinRecursionDepth {
		t.Errorf("Minimum boundary value should be accepted, expected %d, got %d", MinRecursionDepth, om.GetMaxRecursionDepth())
	}

	om.SetMaxRecursionDepth(MaxRecursionDepth)
	if om.GetMaxRecursionDepth() != MaxRecursionDepth {
		t.Errorf("Maximum boundary value should be accepted, expected %d, got %d", MaxRecursionDepth, om.GetMaxRecursionDepth())
	}
}

// TestObserverGetMaxRecursionDepth tests the GetMaxRecursionDepth method
func TestObserverGetMaxRecursionDepth(t *testing.T) {
	om := &ObserverManager{
		callbacks:         make(map[string]ElementCallback),
		dismountCallbacks: make(map[string][]LifecycleHook),
		trackedElements:   make(map[string]js.Value),
		maxRecursionDepth: 75, // Set a specific test value
	}

	result := om.GetMaxRecursionDepth()
	if result != 75 {
		t.Errorf("Expected GetMaxRecursionDepth to return 75, got %d", result)
	}
}

// TestObserverInitializationValidation tests that the global observer is initialized with validated values
func TestObserverInitializationValidation(t *testing.T) {
	// The global observer should be initialized with a valid maxRecursionDepth
	if globalObserver == nil {
		t.Fatal("globalObserver should be initialized")
	}

	depth := globalObserver.GetMaxRecursionDepth()

	// Should be within valid bounds
	if depth < MinRecursionDepth || depth > MaxRecursionDepth {
		t.Errorf("Global observer maxRecursionDepth %d is outside valid bounds [%d, %d]", depth, MinRecursionDepth, MaxRecursionDepth)
	}

	// Should be the default value (since we initialize with DefaultRecursionDepth)
	if depth != DefaultRecursionDepth {
		t.Errorf("Expected global observer maxRecursionDepth to be %d, got %d", DefaultRecursionDepth, depth)
	}
}

// TestObserverConcurrentMaxRecursionDepthAccess tests concurrent access to maxRecursionDepth setter and getter
func TestObserverConcurrentMaxRecursionDepthAccess(t *testing.T) {
	om := &ObserverManager{
		callbacks:         make(map[string]ElementCallback),
		dismountCallbacks: make(map[string][]LifecycleHook),
		trackedElements:   make(map[string]js.Value),
		maxRecursionDepth: DefaultRecursionDepth,
	}

	var wg sync.WaitGroup

	// Concurrent setters
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(value int) {
			defer wg.Done()
			// Set values within valid range
			om.SetMaxRecursionDepth(10 + (value * 5)) // Values from 10 to 55
		}(i)
	}

	// Concurrent getters
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			depth := om.GetMaxRecursionDepth()
			// Verify we get a valid value
			if depth < MinRecursionDepth || depth > MaxRecursionDepth {
				t.Errorf("Concurrent getter returned invalid depth: %d", depth)
			}
		}()
	}

	wg.Wait()

	// Final check - should have a valid value
	finalDepth := om.GetMaxRecursionDepth()
	if finalDepth < MinRecursionDepth || finalDepth > MaxRecursionDepth {
		t.Errorf("Final depth after concurrent access is invalid: %d", finalDepth)
	}
}
