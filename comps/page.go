//go:build js && wasm

package comps

import (
	"crypto/md5"
	"fmt"
	"reflect"
	"sync"

	"github.com/ozanturksever/uiwgo/reactivity"
	g "maragu.dev/gomponents"
	h "maragu.dev/gomponents/html"
)

type ExampleComp struct {
	Title string
}

func (p *ExampleComp) OnMount() {
	// server queries
	// set some default state
}

func (p *ExampleComp) OnUnMount() {
	// clean up
}

func (p *ExampleComp) Render() g.Node {
	return h.Div(g.Text("test:"), g.Text(p.Title), ComponentFactory(&OtherComp{}))
}

type OtherComp struct {
}

func (p *OtherComp) OnMount() {
	// server queries
	// set some default state
}

func (p *OtherComp) OnUnMount() {
	// clean up
}

func (p *OtherComp) Render() g.Node {
	return h.Div(g.Text("from other comp"))
}

type Comp interface {
	OnMount()
	OnUnMount()
	Render() g.Node
}

// ComponentWithProps defines the interface for components that accept props
type ComponentWithProps[P any] interface {
	OnMount()
	OnUnMount()
	Render(props P) g.Node
}

// StatefulComponent defines the interface for components with internal state
type StatefulComponent interface {
	Comp
	GetState() any
	SetState(state any)
}

// ComponentInstance holds the state and lifecycle information for a component
type ComponentInstance struct {
	ID             string
	Component      Comp
	MountContainer string
	CleanupScope   *reactivity.CleanupScope
	Mounted        bool
	mutex          sync.RWMutex
}

var (
	// componentRegistry stores active component instances
	componentRegistry = make(map[string]*ComponentInstance)
	// registryMutex protects the component registry
	registryMutex sync.RWMutex
)

// Mount sets up the component for mounting
func (ci *ComponentInstance) Mount() {
	ci.mutex.Lock()
	defer ci.mutex.Unlock()

	if ci.Mounted {
		return
	}

	// Create cleanup scope for this component
	ci.CleanupScope = reactivity.NewCleanupScope(reactivity.GetCurrentCleanupScope())

	// Schedule OnMount to be called after DOM attachment
	enqueueOnMount(func() {
		ci.mutex.Lock()
		defer ci.mutex.Unlock()

		// Set cleanup scope context during OnMount execution
		previous := reactivity.GetCurrentCleanupScope()
		reactivity.SetCurrentCleanupScope(ci.CleanupScope)
		defer reactivity.SetCurrentCleanupScope(previous)

		ci.Component.OnMount()
		ci.Mounted = true
	})

	// Register cleanup handler in parent scope
	reactivity.OnCleanup(func() {
		ci.Unmount()
	})
}

// Unmount cleans up the component
func (ci *ComponentInstance) Unmount() {
	ci.mutex.Lock()
	defer ci.mutex.Unlock()

	if !ci.Mounted {
		return
	}

	// Call component OnUnMount
	ci.Component.OnUnMount()

	// Dispose cleanup scope
	if ci.CleanupScope != nil {
		ci.CleanupScope.Dispose()
	}

	// Remove from registry
	registryMutex.Lock()
	delete(componentRegistry, ci.ID)
	registryMutex.Unlock()

	ci.Mounted = false
}

// generateComponentID creates a unique ID for a component instance
func generateComponentID(c Comp) string {
	// Use reflection to get component type
	t := reflect.TypeOf(c)
	typeName := t.String()

	// Include mount container for scoping
	container := getCurrentMountContainer()

	// Generate hash for uniqueness based on type, container, and pointer
	hash := md5.Sum([]byte(fmt.Sprintf("%s:%s:%p", typeName, container, c)))

	return fmt.Sprintf("%s-%x", typeName, hash[:8])
}

// ComponentFactory creates a new component instance
// Ensures component state is preserved across renders and manages lifecycle
func ComponentFactory(c Comp) g.Node {
	// Generate unique ID for this component instance
	id := generateComponentID(c)

	registryMutex.Lock()
	instance, exists := componentRegistry[id]
	if !exists {
		// Create new component instance
		instance = &ComponentInstance{
			ID:             id,
			Component:      c,
			MountContainer: getCurrentMountContainer(),
			Mounted:        false,
		}
		componentRegistry[id] = instance
	}
	registryMutex.Unlock()

	// Mount the component (sets up lifecycle)
	instance.Mount()

	// Return the rendered component
	return instance.Component.Render()
}

// cleanupComponentsForContainer removes all components for a specific container
func cleanupComponentsForContainer(containerID string) {
	registryMutex.Lock()
	// Collect instances to unmount without holding the lock during unmount
	instancesToUnmount := make([]*ComponentInstance, 0)
	for id, instance := range componentRegistry {
		if instance.MountContainer == containerID {
			instancesToUnmount = append(instancesToUnmount, instance)
			delete(componentRegistry, id)
		}
	}
	registryMutex.Unlock()

	// Unmount each instance without holding the registry lock
	for _, instance := range instancesToUnmount {
		instance.Unmount()
	}
}

// GetComponentRegistry returns a copy of the current component registry for testing
func GetComponentRegistry() map[string]*ComponentInstance {
	registryMutex.RLock()
	defer registryMutex.RUnlock()

	copy := make(map[string]*ComponentInstance)
	for k, v := range componentRegistry {
		copy[k] = v
	}
	return copy
}

// ComponentInstanceWithProps holds the state and lifecycle information for a component with props
type ComponentInstanceWithProps[P any] struct {
	ID             string
	Component      ComponentWithProps[P]
	Props          P
	MountContainer string
	CleanupScope   *reactivity.CleanupScope
	Mounted        bool
	mutex          sync.RWMutex
}

// Mount sets up the component for mounting
func (ci *ComponentInstanceWithProps[P]) Mount() {
	ci.mutex.Lock()
	defer ci.mutex.Unlock()

	if ci.Mounted {
		return
	}

	// Create cleanup scope for this component
	ci.CleanupScope = reactivity.NewCleanupScope(reactivity.GetCurrentCleanupScope())

	// Schedule OnMount to be called after DOM attachment
	enqueueOnMount(func() {
		// Set cleanup scope context during OnMount execution
		previous := reactivity.GetCurrentCleanupScope()
		reactivity.SetCurrentCleanupScope(ci.CleanupScope)
		defer reactivity.SetCurrentCleanupScope(previous)

		ci.Component.OnMount()
	})

	// Register cleanup handler in parent scope
	reactivity.OnCleanup(func() {
		ci.Unmount()
	})

	ci.Mounted = true
}

// Unmount cleans up the component
func (ci *ComponentInstanceWithProps[P]) Unmount() {
	ci.mutex.Lock()
	defer ci.mutex.Unlock()

	if !ci.Mounted {
		return
	}

	// Call component OnUnMount
	ci.Component.OnUnMount()

	// Dispose cleanup scope
	if ci.CleanupScope != nil {
		ci.CleanupScope.Dispose()
	}

	// Remove from registry
	registryMutex.Lock()
	delete(componentRegistry, ci.ID)
	registryMutex.Unlock()

	ci.Mounted = false
}

// generateComponentIDWithProps creates a unique ID for a component instance with props
func generateComponentIDWithProps[P any](c ComponentWithProps[P], props P) string {
	// Use reflection to get component type
	t := reflect.TypeOf(c)
	typeName := t.String()

	// Include mount container for scoping
	container := getCurrentMountContainer()

	// Include props in the hash for uniqueness
	propsHash := md5.Sum([]byte(fmt.Sprintf("%v", props)))

	// Generate hash for uniqueness based on type, container, pointer, and props
	hash := md5.Sum([]byte(fmt.Sprintf("%s:%s:%p:%x", typeName, container, c, propsHash[:8])))

	return fmt.Sprintf("%s-%x", typeName, hash[:8])
}

// ComponentFactoryWithProps creates a new component instance with props
// Ensures component state is preserved across renders and manages lifecycle
func ComponentFactoryWithProps[P any](c ComponentWithProps[P], props P) g.Node {
	// Generate unique ID for this component instance with props
	id := generateComponentIDWithProps(c, props)

	registryMutex.Lock()
	instance, exists := componentRegistry[id]
	if !exists {
		// Create new component instance with props
		propsInstance := &ComponentInstanceWithProps[P]{
			ID:             id,
			Component:      c,
			Props:          props,
			MountContainer: getCurrentMountContainer(),
			Mounted:        false,
		}
		componentRegistry[id] = &ComponentInstance{
			ID:             id,
			Component:      &propsComponentWrapper[P]{instance: propsInstance},
			MountContainer: propsInstance.MountContainer,
			Mounted:        false,
		}
		instance = componentRegistry[id]
	}
	registryMutex.Unlock()

	// Mount the component (sets up lifecycle)
	instance.Mount()

	// Return the rendered component with props
	if propsWrapper, ok := instance.Component.(*propsComponentWrapper[P]); ok {
		return propsWrapper.Render()
	}
	return g.Text("") // Fallback
}

// propsComponentWrapper wraps a ComponentWithProps to implement the Comp interface
type propsComponentWrapper[P any] struct {
	instance *ComponentInstanceWithProps[P]
}

func (w *propsComponentWrapper[P]) OnMount() {
	w.instance.Component.OnMount()
}

func (w *propsComponentWrapper[P]) OnUnMount() {
	w.instance.Component.OnUnMount()
}

func (w *propsComponentWrapper[P]) Render() g.Node {
	return w.instance.Component.Render(w.instance.Props)
}

// statefulComponentWrapper wraps a StatefulComponent to ensure proper state handling
type statefulComponentWrapper struct {
	component StatefulComponent
}

func (w *statefulComponentWrapper) OnMount() {
	w.component.OnMount()
}

func (w *statefulComponentWrapper) OnUnMount() {
	w.component.OnUnMount()
}

func (w *statefulComponentWrapper) Render() g.Node {
	return w.component.Render()
}
