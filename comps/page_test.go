//go:build js && wasm

package comps

import (
	"fmt"
	"testing"

	g "maragu.dev/gomponents"
	h "maragu.dev/gomponents/html"
)

// TestComponentFactory_CreatesNewInstance tests that ComponentFactory creates a new instance when none exists
func TestComponentFactory_CreatesNewInstance(t *testing.T) {
	defer cleanupComponentRegistry()

	comp := &TestLifecycleComp{}
	node := ComponentFactory(comp)

	// Verify the node is rendered
	if node == nil {
		t.Error("ComponentFactory should return a non-nil node")
	}

	// Verify the instance is in the registry
	registry := GetComponentRegistry()
	if len(registry) != 1 {
		t.Errorf("Expected 1 component in registry, got %d", len(registry))
	}

	// Verify the instance properties
	for _, instance := range registry {
		if instance.Component != comp {
			t.Error("Registry should contain the same component instance")
		}
		if instance.Mounted {
			t.Error("Component should not be mounted yet")
		}
	}
}

// TestComponentFactory_ReusesExistingInstance tests that ComponentFactory reuses existing instances
func TestComponentFactory_ReusesExistingInstance(t *testing.T) {
	defer cleanupComponentRegistry()

	comp := &TestLifecycleComp{}
	node1 := ComponentFactory(comp)
	node2 := ComponentFactory(comp)

	// Should not panic on node comparison - nodes are functions and can't be compared directly
	// Instead, verify that both calls return non-nil nodes
	if node1 == nil {
		t.Error("First ComponentFactory call should return a non-nil node")
	}
	if node2 == nil {
		t.Error("Second ComponentFactory call should return a non-nil node")
	}

	// Only one instance should be in the registry
	registry := GetComponentRegistry()
	if len(registry) != 1 {
		t.Errorf("Expected 1 component in registry, got %d", len(registry))
	}

	// Verify the same component instance is reused
	var foundComp *TestLifecycleComp
	for _, instance := range registry {
		if tc, ok := instance.Component.(*TestLifecycleComp); ok {
			if foundComp != nil {
				t.Error("Should only have one TestLifecycleComp instance in registry")
			}
			foundComp = tc
		}
	}

	if foundComp != comp {
		t.Error("Registry should contain the same component instance that was passed in")
	}
}

// TestComponentInstance_Mount tests the Mount method
func TestComponentInstance_Mount(t *testing.T) {
	defer cleanupComponentRegistry()

	comp := &ExampleComp{Title: "Test"}
	instance := &ComponentInstance{
		ID:        "test-instance",
		Component: comp,
		Mounted:   false,
	}

	// In WASM environment, we can't easily mock package-level functions
	// Instead, we test the behavior through the public ComponentFactory API
	// This test verifies that Mount sets the basic state correctly
	instance.Mount()

	// Execute the mount queue to simulate the actual mounting process
	executeMountQueue()

	if !instance.Mounted {
		t.Error("Mount should set Mounted to true after mount queue execution")
	}

	// The cleanup scope creation depends on the current reactive context
	// which is complex to mock in WASM, so we focus on state changes
}

// TestComponentInstance_Unmount tests the Unmount method
func TestComponentInstance_Unmount(t *testing.T) {
	defer cleanupComponentRegistry()

	comp := &ExampleComp{Title: "Test"}
	instance := &ComponentInstance{
		ID:        "test-instance",
		Component: comp,
		Mounted:   true,
	}

	// Add to registry for cleanup test
	componentRegistry[instance.ID] = instance

	instance.Unmount()

	if instance.Mounted {
		t.Error("Unmount should set Mounted to false")
	}

	// Verify removed from registry
	if _, exists := componentRegistry[instance.ID]; exists {
		t.Error("Unmount should remove instance from registry")
	}
}

// TestGenerateComponentID_Unique tests that component IDs are unique
func TestGenerateComponentID_Unique(t *testing.T) {
	comp1 := &ExampleComp{Title: "Test1"}
	comp2 := &ExampleComp{Title: "Test2"}

	id1 := generateComponentID(comp1)
	id2 := generateComponentID(comp2)

	if id1 == id2 {
		t.Error("generateComponentID should generate unique IDs for different components")
	}

	// Same component should have same ID
	id1Again := generateComponentID(comp1)
	if id1 != id1Again {
		t.Error("generateComponentID should generate same ID for same component")
	}
}

// TestCleanupComponentsForContainer tests container-based cleanup
func TestCleanupComponentsForContainer(t *testing.T) {
	defer cleanupComponentRegistry()

	// Create instances for different containers
	comp1 := &ExampleComp{Title: "Test1"}
	comp2 := &ExampleComp{Title: "Test2"}

	instance1 := &ComponentInstance{
		ID:             "test1",
		Component:      comp1,
		MountContainer: "container1",
		Mounted:        true,
	}

	instance2 := &ComponentInstance{
		ID:             "test2",
		Component:      comp2,
		MountContainer: "container2",
		Mounted:        true,
	}

	componentRegistry[instance1.ID] = instance1
	componentRegistry[instance2.ID] = instance2

	// Cleanup only container1
	cleanupComponentsForContainer("container1")

	// Verify only container1 instances are removed
	if _, exists := componentRegistry[instance1.ID]; exists {
		t.Error("cleanupComponentsForContainer should remove instances for specified container")
	}

	if _, exists := componentRegistry[instance2.ID]; !exists {
		t.Error("cleanupComponentsForContainer should not remove instances for other containers")
	}
}

// SimpleTestComp is a simple component for testing without nested components
type SimpleTestComp struct {
	Title string
}

func (p *SimpleTestComp) OnMount() {
	// Test implementation
}

func (p *SimpleTestComp) OnUnMount() {
	// Test implementation
}

func (p *SimpleTestComp) Render() g.Node {
	return h.Div(g.Text("Simple:"), g.Text(p.Title))
}

// TestComponentFactory_WithDifferentContainers tests component isolation across containers
func TestComponentFactory_WithDifferentContainers(t *testing.T) {
	cleanupComponentRegistry() // Clean up before test
	defer cleanupComponentRegistry()
	defer setCurrentMountContainer("") // Reset mount container after test

	comp := &SimpleTestComp{Title: "Test"}

	// Set different mount containers
	setCurrentMountContainer("container1")
	node1 := ComponentFactory(comp)

	setCurrentMountContainer("container2")
	node2 := ComponentFactory(comp)

	// Should create different instances for different containers
	registry := GetComponentRegistry()
	if len(registry) != 2 {
		t.Errorf("Expected 2 components in registry for different containers, got %d: %v", len(registry), registry)
	}

	// Nodes should be non-nil
	if node1 == nil {
		t.Error("First ComponentFactory call should return a non-nil node")
	}
	if node2 == nil {
		t.Error("Second ComponentFactory call should return a non-nil node")
	}
}

// TestLifecycleComp is a test component that tracks mount and unmount calls
type TestLifecycleComp struct {
	MountCalled   bool
	UnmountCalled bool
}

func (c *TestLifecycleComp) OnMount() {
	c.MountCalled = true
}

func (c *TestLifecycleComp) OnUnMount() {
	c.UnmountCalled = true
}

func (c *TestLifecycleComp) Render() g.Node {
	return h.Div(g.Text("Lifecycle Test"))
}

// TestComponentLifecycle_Integration tests full lifecycle integration
func TestComponentLifecycle_Integration(t *testing.T) {
	defer cleanupComponentRegistry()

	// Create a test component that tracks calls
	comp := &TestLifecycleComp{}

	// Create and mount component
	node := ComponentFactory(comp)
	if node == nil {
		t.Error("ComponentFactory should return a node")
	}

	// Force execution of mount callbacks
	executeMountQueue()

	if !comp.MountCalled {
		t.Error("Component OnMount should be called")
	}

	// Cleanup
	cleanupComponentsForContainer(getCurrentMountContainer())

	if !comp.UnmountCalled {
		t.Error("Component OnUnMount should be called")
	}
}

// Helper function to clean up the registry after tests
func cleanupComponentRegistry() {
	registryMutex.Lock()
	defer registryMutex.Unlock()
	for id := range componentRegistry {
		delete(componentRegistry, id)
	}
}

// Helper function to execute all queued mount callbacks
func executeMountQueue() {
	for len(mountQueue) > 0 {
		callback := mountQueue[0]
		mountQueue = mountQueue[1:]
		callback()
	}
}

// TestComponentWithProps tests the props-based component system
func TestComponentWithProps(t *testing.T) {
	defer cleanupComponentRegistry()

	// Create a test component that uses props
	comp := &TestPropsComp{}
	props := TestProps{Name: "Test", Count: 42}

	// Create component with props
	node := ComponentFactoryWithProps(comp, props)

	// Verify the node is rendered
	if node == nil {
		t.Error("ComponentFactoryWithProps should return a non-nil node")
	}

	// Verify the instance is in the registry with props
	registry := GetComponentRegistry()
	if len(registry) != 1 {
		t.Errorf("Expected 1 component in registry, got %d", len(registry))
	}

	// Verify props are stored correctly (this would need to be implemented in the factory)
	for _, instance := range registry {
		// Check if the component is wrapped for props
		if _, ok := instance.Component.(*propsComponentWrapper[TestProps]); !ok {
			t.Error("Registry should contain a props component wrapper")
		}
	}
}

// TestFragment tests the Fragment helper
func TestFragment(t *testing.T) {
	// Create some test nodes
	child1 := h.Div(g.Text("Child 1"))
	child2 := h.Div(g.Text("Child 2"))
	child3 := h.Div(g.Text("Child 3"))

	// Create a fragment with multiple children
	fragment := Fragment(child1, child2, child3)

	// Verify that fragment is not nil
	if fragment == nil {
		t.Error("Fragment should return a non-nil node")
	}

	// In gomponents, Group nodes don't render as a single element,
	// so we can't easily test the output without rendering to string.
	// But we can verify that it accepts multiple children without error.
}

// TestPortal tests the Portal helper
func TestPortal(t *testing.T) {
	// Create a test child node
	child := h.Div(g.Text("Portal Content"))

	// Create a portal with a target and child
	portal := Portal("#modal-target", child)

	// Verify that portal returns the child node (basic implementation)
	if portal == nil {
		t.Error("Portal should return a non-nil node")
	}

	// In the current basic implementation, Portal just returns the children
	// A full implementation would require DOM manipulation during mount
}

// TestMemo tests the Memo helper
func TestMemo(t *testing.T) {
	// Create a simple component function
	counter := 0
	componentFunc := func() g.Node {
		counter++
		return h.Div(g.Text(fmt.Sprintf("Rendered %d times", counter)))
	}

	// Call Memo multiple times with same dependencies
	node1 := Memo(componentFunc, "dep1")
	node2 := Memo(componentFunc, "dep1")

	// Both should return non-nil nodes
	if node1 == nil {
		t.Error("First Memo call should return a non-nil node")
	}
	if node2 == nil {
		t.Error("Second Memo call should return a non-nil node")
	}

	// In the current basic implementation, Memo doesn't actually cache
	// A full implementation would need to track dependencies
}

// TestLazy tests the Lazy helper
func TestLazy(t *testing.T) {
	// Create a loader function that returns a component
	loader := func() func() g.Node {
		return func() g.Node {
			return h.Div(g.Text("Lazy Loaded Component"))
		}
	}

	// Load the component lazily
	node := Lazy(loader)

	// Should return a non-nil node
	if node == nil {
		t.Error("Lazy should return a non-nil node")
	}

	// In the current basic implementation, Lazy loads synchronously
	// A full implementation would handle asynchronous loading
}

// TestErrorBoundary tests the ErrorBoundary helper
func TestErrorBoundary(t *testing.T) {
	// Create a fallback function
	fallback := func(err error) g.Node {
		return h.Div(g.Text(fmt.Sprintf("Error: %v", err)))
	}

	// Create normal children
	children := h.Div(g.Text("Normal Content"))

	// Create error boundary
	node := ErrorBoundary(ErrorBoundaryProps{
		Fallback: fallback,
		Children: children,
	})

	// Should return a non-nil node
	if node == nil {
		t.Error("ErrorBoundary should return a non-nil node")
	}

	// In the current basic implementation, ErrorBoundary doesn't actually catch errors
	// A full implementation would need to handle error catching
}

// TestComponentWithProps_ReuseInstance tests instance reuse with same props
func TestComponentWithProps_ReuseInstance(t *testing.T) {
	defer cleanupComponentRegistry()

	comp := &TestPropsComp{}
	props := TestProps{Name: "Test", Count: 42}

	// Create two instances with same props
	node1 := ComponentFactoryWithProps(comp, props)
	node2 := ComponentFactoryWithProps(comp, props)

	// Should return non-nil nodes
	if node1 == nil {
		t.Error("First ComponentFactoryWithProps call should return a non-nil node")
	}
	if node2 == nil {
		t.Error("Second ComponentFactoryWithProps call should return a non-nil node")
	}

	// Only one instance should be in the registry
	registry := GetComponentRegistry()
	if len(registry) != 1 {
		t.Errorf("Expected 1 component in registry, got %d", len(registry))
	}
}

// TestComponentWithProps_DifferentProps tests different instances with different props
func TestComponentWithProps_DifferentProps(t *testing.T) {
	defer cleanupComponentRegistry()

	comp := &TestPropsComp{}
	props1 := TestProps{Name: "Test1", Count: 42}
	props2 := TestProps{Name: "Test2", Count: 100}

	// Create instances with different props
	node1 := ComponentFactoryWithProps(comp, props1)
	node2 := ComponentFactoryWithProps(comp, props2)

	// Should create different instances for different props
	registry := GetComponentRegistry()
	if len(registry) != 2 {
		t.Errorf("Expected 2 components in registry for different props, got %d", len(registry))
	}

	// Nodes should be non-nil
	if node1 == nil {
		t.Error("First ComponentFactoryWithProps call should return a non-nil node")
	}
	if node2 == nil {
		t.Error("Second ComponentFactoryWithProps call should return a non-nil node")
	}
}

// TestPropsLifecycleComp is a test component that tracks mount and unmount calls for props
type TestPropsLifecycleComp struct {
	MountCalled   bool
	UnmountCalled bool
}

func (c *TestPropsLifecycleComp) OnMount() {
	c.MountCalled = true
}

func (c *TestPropsLifecycleComp) OnUnMount() {
	c.UnmountCalled = true
}

func (c *TestPropsLifecycleComp) Render(props TestProps) g.Node {
	return h.Div(g.Text(fmt.Sprintf("Props Lifecycle: %s, %d", props.Name, props.Count)))
}

// TestComponentWithProps_Lifecycle tests lifecycle integration with props
func TestComponentWithProps_Lifecycle(t *testing.T) {
	defer cleanupComponentRegistry()

	comp := &TestPropsLifecycleComp{}
	props := TestProps{Name: "Test", Count: 42}

	// Create and mount component with props
	node := ComponentFactoryWithProps(comp, props)
	if node == nil {
		t.Error("ComponentFactoryWithProps should return a node")
	}

	// Force execution of mount callbacks
	executeMountQueue()

	if !comp.MountCalled {
		t.Error("Component OnMount should be called for props-based component")
	}

	// Cleanup
	cleanupComponentsForContainer(getCurrentMountContainer())

	if !comp.UnmountCalled {
		t.Error("Component OnUnMount should be called for props-based component")
	}
}

// TestPropsComp is a test component that uses props
type TestPropsComp struct {
}

func (p *TestPropsComp) OnMount() {
	// Test implementation
}

func (p *TestPropsComp) OnUnMount() {
	// Test implementation
}

func (p *TestPropsComp) Render(props TestProps) g.Node {
	return h.Div(g.Text(fmt.Sprintf("Name: %s, Count: %d", props.Name, props.Count)))
}

// TestProps represents test props
type TestProps struct {
	Name  string
	Count int
}

// TestStatefulComponent tests the StatefulComponent interface
func TestStatefulComponent(t *testing.T) {
	defer cleanupComponentRegistry()

	comp := &TestStatefulComp{}
	comp.SetState(TestState{Value: "initial"})

	// Create component instance
	node := ComponentFactory(comp)

	// Verify the node is rendered
	if node == nil {
		t.Error("ComponentFactory should return a non-nil node for stateful component")
	}

	// Verify the instance is in the registry
	registry := GetComponentRegistry()
	if len(registry) != 1 {
		t.Errorf("Expected 1 component in registry, got %d", len(registry))
	}

	// Verify state is preserved
	for _, instance := range registry {
		if statefulComp, ok := instance.Component.(StatefulComponent); ok {
			state := statefulComp.GetState().(TestState)
			if state.Value != "initial" {
				t.Errorf("Expected state value 'initial', got '%s'", state.Value)
			}
		} else {
			t.Error("Component should implement StatefulComponent interface")
		}
	}
}

// TestStatefulComponent_StateUpdate tests state updates
func TestStatefulComponent_StateUpdate(t *testing.T) {
	defer cleanupComponentRegistry()

	comp := &TestStatefulComp{}
	comp.SetState(TestState{Value: "initial"})

	// Create component instance
	node1 := ComponentFactory(comp)

	// Update state
	comp.SetState(TestState{Value: "updated"})

	// Create another instance - should reuse and preserve updated state
	node2 := ComponentFactory(comp)

	// Should return non-nil nodes
	if node1 == nil {
		t.Error("First ComponentFactory call should return a non-nil node")
	}
	if node2 == nil {
		t.Error("Second ComponentFactory call should return a non-nil node")
	}

	// Verify state is updated
	registry := GetComponentRegistry()
	for _, instance := range registry {
		if statefulComp, ok := instance.Component.(StatefulComponent); ok {
			state := statefulComp.GetState().(TestState)
			if state.Value != "updated" {
				t.Errorf("Expected state value 'updated', got '%s'", state.Value)
			}
		}
	}
}

// TestStatefulLifecycleComp is a test component that tracks mount and unmount calls for stateful components
type TestStatefulLifecycleComp struct {
	MountCalled   bool
	UnmountCalled bool
	state         TestState
}

func (c *TestStatefulLifecycleComp) OnMount() {
	c.MountCalled = true
}

func (c *TestStatefulLifecycleComp) OnUnMount() {
	c.UnmountCalled = true
}

func (c *TestStatefulLifecycleComp) Render() g.Node {
	return h.Div(g.Text(fmt.Sprintf("Stateful Lifecycle: %s", c.state.Value)))
}

func (c *TestStatefulLifecycleComp) GetState() any {
	return c.state
}

func (c *TestStatefulLifecycleComp) SetState(state any) {
	c.state = state.(TestState)
}

// TestStatefulComponent_Lifecycle tests lifecycle with state
func TestStatefulComponent_Lifecycle(t *testing.T) {
	defer cleanupComponentRegistry()

	comp := &TestStatefulLifecycleComp{}
	comp.SetState(TestState{Value: "test"})

	// Create and mount component
	node := ComponentFactory(comp)
	if node == nil {
		t.Error("ComponentFactory should return a node for stateful component")
	}

	// Force execution of mount callbacks
	executeMountQueue()

	if !comp.MountCalled {
		t.Error("Stateful component OnMount should be called")
	}

	// Cleanup
	cleanupComponentsForContainer(getCurrentMountContainer())

	if !comp.UnmountCalled {
		t.Error("Stateful component OnUnMount should be called")
	}
}

// TestStatefulComp is a test component that implements StatefulComponent
type TestStatefulComp struct {
	state TestState
}

func (p *TestStatefulComp) OnMount() {
	// Test implementation
}

func (p *TestStatefulComp) OnUnMount() {
	// Test implementation
}

func (p *TestStatefulComp) Render() g.Node {
	return h.Div(g.Text(fmt.Sprintf("Value: %s", p.state.Value)))
}

func (p *TestStatefulComp) GetState() any {
	return p.state
}

func (p *TestStatefulComp) SetState(state any) {
	p.state = state.(TestState)
}

// TestState represents test state
type TestState struct {
	Value string
}
