// store_bindings.go
// Store-component binding utilities for reactive UI updates

package golid

import (
	"reflect"
	"sync"
	"sync/atomic"
)

// ------------------------------------
// 🔗 Store-Component Binding Types
// ------------------------------------

// ComponentBinding represents a binding between a store and a component
type ComponentBinding struct {
	id        uint64
	store     interface{}
	component interface{}
	selector  func(interface{}) interface{}
	updater   func(interface{}, interface{})
	cleanup   func()
	owner     *Owner
	active    bool
	mutex     sync.RWMutex
}

// BindingOptions configures store-component bindings
type BindingOptions struct {
	Owner     *Owner
	Selector  func(interface{}) interface{}
	Updater   func(interface{}, interface{})
	Immediate bool
}

// ComponentConnector manages store-component connections
type ComponentConnector struct {
	bindings map[uint64]*ComponentBinding
	mutex    sync.RWMutex
}

var globalConnector = &ComponentConnector{
	bindings: make(map[uint64]*ComponentBinding),
}

// ------------------------------------
// 🔗 Store-Component Binding Functions
// ------------------------------------

// ConnectStore connects a store to a component with automatic updates
func ConnectStore[T any](store *Store[T], component interface{}, options ...BindingOptions) func() {
	var opts BindingOptions
	if len(options) > 0 {
		opts = options[0]
	}

	// Default selector (identity function)
	if opts.Selector == nil {
		opts.Selector = func(value interface{}) interface{} {
			return value
		}
	}

	// Default updater (calls Update method if available)
	if opts.Updater == nil {
		opts.Updater = func(comp interface{}, value interface{}) {
			updateComponent(comp, value)
		}
	}

	// Create binding
	binding := &ComponentBinding{
		id:        generateBindingId(),
		store:     store,
		component: component,
		selector:  opts.Selector,
		updater:   opts.Updater,
		owner:     opts.Owner,
		active:    true,
	}

	// Subscribe to store changes
	unsubscribe := store.Subscribe(func(value T) {
		binding.mutex.RLock()
		if !binding.active {
			binding.mutex.RUnlock()
			return
		}

		selectedValue := binding.selector(value)
		component := binding.component
		updater := binding.updater
		binding.mutex.RUnlock()

		// Update component with selected value
		updater(component, selectedValue)
	})

	// Set up cleanup
	binding.cleanup = func() {
		binding.mutex.Lock()
		binding.active = false
		binding.mutex.Unlock()
		unsubscribe()
		globalConnector.removeBinding(binding.id)
	}

	// Register binding
	globalConnector.addBinding(binding)

	// Register cleanup with owner if available
	if opts.Owner != nil {
		opts.Owner.addCleanup(binding.cleanup)
	}

	// Immediate update if requested
	if opts.Immediate {
		currentValue := store.Get()
		selectedValue := opts.Selector(currentValue)
		opts.Updater(component, selectedValue)
	}

	return binding.cleanup
}

// ConnectDerivedStore connects a derived store to a component
func ConnectDerivedStore[T any](derivedStore *Store[T], component interface{}, options ...BindingOptions) func() {
	return ConnectStore(derivedStore, component, options...)
}

// ConnectMultipleStores connects multiple stores to a component with a combiner function
func ConnectMultipleStores(stores []interface{}, component interface{}, combiner func([]interface{}) interface{}, options ...BindingOptions) func() {
	var opts BindingOptions
	if len(options) > 0 {
		opts = options[0]
	}

	// Create a combined store
	combinedStore := CreateDerivedStore(func() interface{} {
		values := make([]interface{}, len(stores))
		for i, storeInterface := range stores {
			// Use reflection to get store value
			storeValue := reflect.ValueOf(storeInterface)
			if storeValue.Kind() == reflect.Ptr {
				storeValue = storeValue.Elem()
			}

			// Look for Get method
			getMethod := storeValue.MethodByName("Get")
			if getMethod.IsValid() {
				results := getMethod.Call(nil)
				if len(results) > 0 {
					values[i] = results[0].Interface()
				}
			}
		}
		return combiner(values)
	}, StoreOptions[interface{}]{
		Owner: opts.Owner,
	})

	return ConnectStore(combinedStore, component, options...)
}

// ------------------------------------
// 🎯 Component Update Utilities
// ------------------------------------

// updateComponent attempts to update a component using various strategies
func updateComponent(component interface{}, value interface{}) {
	componentValue := reflect.ValueOf(component)

	// Strategy 1: Call Update method if available
	if updateMethod := componentValue.MethodByName("Update"); updateMethod.IsValid() {
		args := []reflect.Value{reflect.ValueOf(value)}
		updateMethod.Call(args)
		return
	}

	// Strategy 2: Call SetState method if available
	if setStateMethod := componentValue.MethodByName("SetState"); setStateMethod.IsValid() {
		args := []reflect.Value{reflect.ValueOf(value)}
		setStateMethod.Call(args)
		return
	}

	// Strategy 3: Call Render method if available (for re-rendering)
	if renderMethod := componentValue.MethodByName("Render"); renderMethod.IsValid() {
		renderMethod.Call(nil)
		return
	}

	// Strategy 4: Set State field if available
	if componentValue.Kind() == reflect.Ptr {
		componentValue = componentValue.Elem()
	}

	if stateField := componentValue.FieldByName("State"); stateField.IsValid() && stateField.CanSet() {
		stateField.Set(reflect.ValueOf(value))
		return
	}

	// Strategy 5: Set Props field if available
	if propsField := componentValue.FieldByName("Props"); propsField.IsValid() && propsField.CanSet() {
		propsField.Set(reflect.ValueOf(value))
		return
	}
}

// ------------------------------------
// 🔧 Connector Management
// ------------------------------------

var bindingIdCounter uint64

func generateBindingId() uint64 {
	return atomic.AddUint64(&bindingIdCounter, 1)
}

func (c *ComponentConnector) addBinding(binding *ComponentBinding) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.bindings[binding.id] = binding
}

func (c *ComponentConnector) removeBinding(id uint64) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	delete(c.bindings, id)
}

// GetActiveBindings returns the number of active bindings
func GetActiveBindings() int {
	globalConnector.mutex.RLock()
	defer globalConnector.mutex.RUnlock()
	return len(globalConnector.bindings)
}

// CleanupAllBindings cleans up all active bindings
func CleanupAllBindings() {
	globalConnector.mutex.Lock()
	bindings := make([]*ComponentBinding, 0, len(globalConnector.bindings))
	for _, binding := range globalConnector.bindings {
		bindings = append(bindings, binding)
	}
	globalConnector.mutex.Unlock()

	for _, binding := range bindings {
		if binding.cleanup != nil {
			binding.cleanup()
		}
	}
}

// ------------------------------------
// 🧪 Testing Utilities
// ------------------------------------

// MockComponent represents a mock component for testing
type MockComponent struct {
	State       interface{}
	UpdateCount int
	LastValue   interface{}
	mutex       sync.RWMutex
}

// NewMockComponent creates a new mock component
func NewMockComponent() *MockComponent {
	return &MockComponent{}
}

// Update updates the mock component state
func (m *MockComponent) Update(value interface{}) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.State = value
	m.LastValue = value
	m.UpdateCount++
}

// GetState returns the current state thread-safely
func (m *MockComponent) GetState() interface{} {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return m.State
}

// GetUpdateCount returns the update count thread-safely
func (m *MockComponent) GetUpdateCount() int {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return m.UpdateCount
}

// CreateTestBinding creates a binding for testing
func CreateTestBinding[T any](store *Store[T], component *MockComponent) func() {
	return ConnectStore(store, component, BindingOptions{
		Immediate: true,
	})
}
