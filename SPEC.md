### Framework Philosophy: "GoSolid" (A working name)

1.  **Declarative UI, Imperative Performance:** You write declarative components using standard Go functions and `gomponents` (see [gomponents-doc.md](gomponents-doc.md)) for structure. The framework's core (and a future compiler) will translate this into the most efficient, direct DOM manipulation calls.
2.  **Fine-Grained Reactivity:** The foundation is a system of `Signals`, `Effects`, and `Memos`. Updates flow through the system automatically, triggering the smallest possible DOM updates. No component re-rendering, no diffing.
3.  **Components are Just Functions:** A component is a Go function that runs *once* to set up state, create effects, and return the initial DOM structure. It does not re-run when its state changes. This is a critical distinction from React's model and is key to our performance.

---

### 1. The Core Reactive API (`pkg/reactivity`)

This is the engine of our framework. It's completely platform-agnostic; it knows nothing about the DOM.

```go
package reactivity

// Signal is the basic reactive primitive. It holds a value and notifies
// observers when that value changes.
type Signal[T any] interface {
    Get() T       // Get the current value and subscribe the current running effect.
    Set(value T)  // Set a new value and notify all subscribers.
}

// CreateSignal creates a new Signal.
func CreateSignal[T any](initialValue T) Signal[T] {
    // ... implementation details ...
    // Internally, it will have a value and a slice of observers (effects).
    // Get() will add the current observer to the list.
    // Set() will iterate over the observers and schedule them to run.
}

// Effect is a function that runs side-effects, automatically re-executing
// when any of its dependent Signals change.
type Effect interface {
    // Dispose manually stops the effect from running again.
    Dispose()
}

// CreateEffect registers a function to be run. The framework tracks which
// signals are read inside this function and re-runs it when they change.
func CreateEffect(fn func()) Effect {
    // ... implementation details ...
    // This function will set a global "current observer", run fn(), and then
    // unset the global. Any signal.Get() called during fn's execution will
    // register itself with this effect.
}

// Memo creates a derived, cached signal. The calculation function is only
// re-executed when its own dependencies change.
func CreateMemo[T any](fn func() T) Signal[T] {
    // ... implementation details ...
    // This is essentially a signal whose value is determined by an effect.
    // When its dependencies change, it re-runs fn, caches the result, and
    // notifies its own subscribers.
}

// OnCleanup registers a function to be called when the current reactive
// scope is disposed of. This is crucial for preventing memory leaks, e.g.,
// by removing event listeners.
func OnCleanup(fn func()) {
    // ... implementation details ...
    // Adds the cleanup function to a stack associated with the current
    // running effect or component scope.
}
```

### 2. The Component & DOM API (`pkg/github.com/ozanturksever/uiwgo`)

This package builds on the reactive core to provide components, lifecycle hooks, and the rendering logic.

#### Component Definition

A component is simply a function.

```go
package github.com/ozanturksever/uiwgo

import "github.com/ozanturksever/gomponents"

// Node is an alias for gomponents.Node for convenience.
type Node = gomponents.Node

// ComponentFunc defines the signature for all components.
// It receives props and returns a DOM node.
type ComponentFunc[P any] func(props P) Node
```

#### Rendering

This is the entry point for the application.

```go
package github.com/ozanturksever/uiwgo

// Mount renders a root component into a specific DOM element identified by its ID.
// It sets up the top-level reactive context.
func Mount(elementID string, rootComponent func() Node) {
    // 1. Get the target element from the DOM via syscall/js.
    // 2. Create a root reactive scope.
    // 3. Execute the rootComponent function *once*.
    // 4. The returned gomponents.Node is rendered into a real DOM node.
    //    We will write a custom renderer for gomponents that uses direct
    //    syscall/js calls (document.createElement, element.SetAttribute, etc.).
    // 5. Append the created DOM node to the target element.
    // 6. The app is now live. The reactive graph will handle all future updates.
}
```

#### Component Lifecycle

Lifecycle management flows naturally from the reactive primitives. We just provide convenient wrappers.

```go
package github.com/ozanturksever/uiwgo

import "github.com/ozanturksever/uiwgo/pkg/reactivity"

// OnMount runs the given function after the component's elements have been
// mounted to the DOM.
// It's simply an alias for a non-tracking effect, but semantically clearer.
func OnMount(fn func()) {
    // Internally, this is just a CreateEffect that is scheduled to run once
    // after the initial render.
    reactivity.CreateEffect(fn)
}

// OnCleanup is a direct pass-through to the core reactivity cleanup function.
// It's used to clean up resources when a component is removed from the DOM
// (e.g., in a conditional <Show> component).
var OnCleanup = reactivity.OnCleanup
```

#### Control Flow Components

This is where the power of this model becomes evident. We cannot use Go's native `if` or `for` for reactive control flow because the component function only runs once. Instead, we provide special components that wrap effects to manage DOM elements.

```go
package github.com/ozanturksever/uiwgo

// ShowProps defines the properties for the Show component.
type ShowProps struct {
    When     reactivity.Signal[bool]
    Children Node
}

// Show conditionally renders its children.
func Show(props ShowProps) Node {
    // This is a special component. It doesn't just return a static node.
    // 1. It creates a placeholder DOM element (like a comment node or empty div).
    // 2. It creates an Effect that depends on props.When.
    // 3. Inside the effect:
    //    - If props.When.Get() is true, it renders the Children and inserts them
    //      into the DOM after the placeholder. It also sets up a cleanup function
    //      to remove them.
    //    - If props.When.Get() is false, it runs the cleanup from the previous
    //      state, removing the children's nodes from the DOM.
    // This is vastly more efficient than re-rendering a whole subtree and diffing it.
    // We are surgically adding or removing a small set of nodes.
}

// ... Similarly, we would create a `For` component for rendering lists.
// It would take a Signal[[]T] and a render function, and it would perform
// efficient, keyed reconciliation to add, remove, or move items without
// re-creating existing ones.
```

---

### 3. Putting It All Together: A Simple Counter Example

Let's see how a developer would use this API.

**`main.go`**

```go
package main

import (
	"fmt"
	"syscall/js"

	"github.com/ozanturksever/gomponents"
	"maragu.dev/gomponents/html"

	"github.com/ozanturksever/uiwgo"
	"github.com/ozanturksever/uiwgo/pkg/reactivity"
)

// Our simple Counter component.
// It doesn't take any props, so we use an empty struct.
func Counter(props struct{}) github.com/ozanturksever/uiwgo.Node {
	// 1. Create a reactive state variable (a signal).
	count := reactivity.CreateSignal(0)

	// A derived signal (a memo). This will only re-calculate when count changes.
	doubleCount := reactivity.CreateMemo(func() int {
		return count.Get() * 2
	})

	// 2. Set up a side-effect. This will run whenever 'count' changes.
	reactivity.CreateEffect(func() {
		fmt.Printf("The current count is: %d\n", count.Get())
	})

	// This is a lifecycle hook.
	github.com/ozanturksever/uiwgo.OnMount(func() {
		fmt.Println("Counter component has been mounted to the DOM!")
	})

	// 3. Return the DOM structure using gomponents.
	// This function body is executed ONLY ONCE.
	return html.Div(
		html.Class("p-4 border rounded-lg"), // Styling with Tailwind CSS classes

		html.H1(gomponents.Textf("Current Count: %d", count.Get())),

		// IMPORTANT: This text node needs to be reactive.
		// We'll wrap it in an effect to update it directly.
		html.P(github.com/ozanturksever/uiwgo.Text(func() string {
			return fmt.Sprintf("Double Count: %d", doubleCount.Get())
		})),

		html.Button(
			html.Class("px-4 py-2 bg-blue-500 text-white rounded hover:bg-blue-700"),
			// Event handlers are set up once. The closure captures the signal.
			gomponents.Attr("onclick", js.FuncOf(func(this js.Value, args []js.Value) any {
				count.Set(count.Get() + 1)
				return nil
			}).String()),
			gomponents.Text("Increment"),
		),
	)
}

// github.com/ozanturksever/uiwgo.Text is a helper we would create. It takes a function and creates
// a text node, then wraps it in an effect that updates the text node's
// `data` property whenever any signals inside the function change.
/*
func Text(fn func() string) gomponents.Node {
    // Creates a placeholder text node.
    // Creates an effect:
    //   reactivity.CreateEffect(func() {
    //       textNode.Set("data", fn())
    //   })
    // Returns the placeholder node.
}
*/

func main() {
	// Mount the root component into the <div id="app"></div> in our index.html
	github.com/ozanturksever/uiwgo.Mount("app", func() github.com/ozanturksever/uiwgo.Node {
		return Counter(struct{}{})
	})

	// Prevent the Go WASM program from exiting.
	select {}
}
```

### Summary of the API Design

*   **Reactivity (`reactivity` package):** `CreateSignal`, `CreateEffect`, `CreateMemo`. The core, pure logic.
*   **Components (`github.com/ozanturksever/uiwgo` package):**
    *   `ComponentFunc[P]`: The standard signature `func(props P) Node`.
    *   `Mount(id, component)`: The application entry point.
    *   `OnMount(fn)`, `OnCleanup(fn)`: Lifecycle helpers.
    *   `Show(props)`, `For(props)`: Reactive control flow primitives.
    *   `Text(fn)`: A helper for creating reactive text content.

This design provides a powerful, performant, and Go-idiomatic foundation for building web UIs. It directly reflects the SolidJS philosophy, prioritizing performance by creating a direct connection between state and the DOM, entirely sidestepping the overhead of a VDOM.
