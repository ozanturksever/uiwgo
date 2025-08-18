// lifecycle.go
// Component lifecycle system with hooks for initialization, mounting, and dismounting

package golid

import (
	"maragu.dev/gomponents"
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

// ----------------------------------
// 🧱 Component Lifecycle System
// ----------------------------------

// NewComponent creates a new component with lifecycle support
func NewComponent(render func() gomponents.Node) *Component {
	return &Component{
		id:     GenID(),
		render: render,
		lifecycle: &ComponentLifecycle{
			onInit:      make([]LifecycleHook, 0),
			onMount:     make([]LifecycleHook, 0),
			onDismount:  make([]LifecycleHook, 0),
			initialized: false,
			mounted:     false,
		},
	}
}

// OnInit registers a callback to be executed when the component is initialized
func (c *Component) OnInit(hook LifecycleHook) *Component {
	c.lifecycle.onInit = append(c.lifecycle.onInit, hook)
	return c
}

// OnMount registers a callback to be executed when the component is mounted to the DOM
func (c *Component) OnMount(hook LifecycleHook) *Component {
	c.lifecycle.onMount = append(c.lifecycle.onMount, hook)
	return c
}

// OnDismount registers a callback to be executed when the component is removed from the DOM
func (c *Component) OnDismount(hook LifecycleHook) *Component {
	c.lifecycle.onDismount = append(c.lifecycle.onDismount, hook)
	return c
}

// Render creates the DOM node for this component and sets up lifecycle hooks
func (c *Component) Render() Node {
	// Execute onInit hooks only once per component instance
	c.lifecycle.mutex.Lock()
	if !c.lifecycle.initialized && len(c.lifecycle.onInit) > 0 {
		for _, hook := range c.lifecycle.onInit {
			hook()
		}
		c.lifecycle.initialized = true
	}
	c.lifecycle.mutex.Unlock()

	// Create wrapper div with unique ID for tracking
	wrapperID := c.id
	content := c.render()
	wrapper := Div(Attr("id", wrapperID), content)

	// Register mount and dismount hooks with observer only once
	c.lifecycle.mutex.Lock()
	if !c.lifecycle.mounted {
		if len(c.lifecycle.onMount) > 0 {
			globalObserver.RegisterElement(wrapperID, func() {
				c.lifecycle.mutex.Lock()
				if !c.lifecycle.mounted {
					for _, hook := range c.lifecycle.onMount {
						hook()
					}
					c.lifecycle.mounted = true
				}
				c.lifecycle.mutex.Unlock()
			})
		}

		if len(c.lifecycle.onDismount) > 0 {
			for _, hook := range c.lifecycle.onDismount {
				globalObserver.RegisterDismountCallback(wrapperID, func() {
					c.lifecycle.mutex.Lock()
					c.lifecycle.mounted = false
					c.lifecycle.mutex.Unlock()
					hook()
				})
			}
		}
	}
	c.lifecycle.mutex.Unlock()

	return wrapper
}

// WithLifecycle is a convenience function to create a component with lifecycle hooks from a render function
func WithLifecycle(render func() gomponents.Node) *Component {
	return NewComponent(render)
}
