### Framework Philosophy: "UiwGo" (A working name)

1.  **Declarative UI, Imperative Performance:** You write declarative components using standard Go functions and `gomponents` (see [gomponents-doc.md](gomponents-doc.md)) for structure. The framework's core (and a future compiler) will translate this into the most efficient, direct DOM manipulation calls.
2.  **Fine-Grained Reactivity:** The foundation is a system of `Signals`, `Effects`, and `Memos`. Updates flow through the system automatically, triggering the smallest possible DOM updates. No component re-rendering, no diffing.
3.  **Components are Just Functions:** A component is a Go function that runs *once* to set up state, create effects, and return the initial DOM structure. It does not re-run when its state changes. This is a critical distinction from React's model and is key to our performance.
4.  **Fallback Reactivity Mechanisms:** If you exhaust the simple ways to do things, you can revert to use DOM diff or node dirty algorithms to handle reactivity. While fine-grained reactivity is preferred for optimal performance, these fallback mechanisms provide flexibility for complex scenarios where direct signal-based updates are insufficient.

---

### Event Handling Philosophy

The `gomponents` library follows a strict separation of concerns between server-side HTML generation and client-side interactivity:

1. **No Server-Side Event Binding:** The `gomponents` library does not, by design, include support for `OnClick`-style or other server-side event binding mechanisms. This is an intentional architectural decision that maintains clarity between what happens on the server versus what happens in the browser.

2. **Recommended Pattern - Go-Defined Client Logic:** The preferred architectural pattern is to define client-side logic within Go and expose it to JavaScript using `js.Global().Set` and `js.FuncOf`. This approach maintains type safety and centralizes application logic within Go while providing seamless interoperability with the browser environment.

   **Example:**
   ```go
   js.Global().Set("myFunction", js.FuncOf(myGoFunction))
   ```

   This pattern is recommended because it:
   - **Maintains Type Safety:** All application logic remains in Go with compile-time type checking
   - **Centralizes Logic:** Business logic and state management stay within the Go application
   - **Preserves Performance:** Direct function calls without marshaling overhead
   - **Enables Testing:** Go functions can be unit tested independently of the DOM

3. **Clear Separation of Concerns:** This design choice enforces a clear separation of concerns between:
   - **Go backend:** Responsible for generating HTML structure, managing application state, defining business logic, and exposing functions to the JavaScript runtime
   - **Frontend JavaScript:** Responsible primarily for binding Go-defined functions to DOM events and handling browser-specific interactions

4. **Framework Integration:** When using this reactive framework (UiwGo), event handlers are attached directly to DOM elements by binding Go functions to JavaScript. The reactive signals provide the bridge between user interactions and state updates, maintaining the unidirectional data flow that is central to the framework's design. This pattern reinforces the separation of concerns where Go manages the application's state and logic, while JavaScript's role is primarily to bind these Go functions to DOM events.

This philosophy ensures that developers have explicit control over where and how events are handled, while maintaining the performance benefits of fine-grained reactivity for state updates and keeping the majority of application logic within the type-safe Go environment.

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
func Counter(props struct{}) uiwgo.Node {
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
	uiwgo.OnMount(func() {
		fmt.Println("Counter component has been mounted to the DOM!")
	})

	// 3. Return the DOM structure using gomponents.
	// This function body is executed ONLY ONCE.
	return html.Div(
		html.Class("p-4 border rounded-lg"), // Styling with Tailwind CSS classes

		html.H1(gomponents.Textf("Current Count: %d", count.Get())),

		// IMPORTANT: This text node needs to be reactive.
		// We'll wrap it in an effect to update it directly.
		html.P(uiwgo.Text(func() string {
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

// uiwgo.Text is a helper we would create. It takes a function and creates
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
	uiwgo.Mount("app", func() uiwgo.Node {
		return Counter(struct{}{})
	})

	// Prevent the Go WASM program from exiting.
	select {}
}
```

---

**Note on Execution:**

To run the compiled WASM binary, you will need an `index.html` file and the `wasm_exec.js` script. The `wasm_exec.js` file is provided by the Go standard library and should be placed in the root of your project. It handles the loading and execution of the WASM module in the browser.
### Summary of the API Design

*   **Reactivity (`reactivity` package):** `CreateSignal`, `CreateEffect`, `CreateMemo`. The core, pure logic.
*   **Components (`github.com/ozanturksever/uiwgo` package):**
    *   `ComponentFunc[P]`: The standard signature `func(props P) Node`.
    *   `Mount(id, component)`: The application entry point.
    *   `OnMount(fn)`, `OnCleanup(fn)`: Lifecycle helpers.
    *   `Show(props)`, `For(props)`: Reactive control flow primitives.
    *   `Text(fn)`: A helper for creating reactive text content.

This design provides a powerful, performant, and Go-idiomatic foundation for building web UIs. It directly reflects the SolidJS philosophy, prioritizing performance by creating a direct connection between state and the DOM, entirely sidestepping the overhead of a VDOM.


### 4. Test Scenarios

This section outlines test scenarios for the core features of the framework.

#### 4.1. Core Reactivity API (`pkg/reactivity`)

##### 4.1.1. `CreateSignal[T]`

*   **Initial Value:**
    *   Test that `Get()` returns the initial value provided during creation (e.g., for integers, strings, structs).
    *   Test with a nil or zero value as the initial value.
*   **Value Update:**
    *   Test that `Set()` correctly updates the signal's value.
    *   Test that a subsequent `Get()` returns the new value.
*   **Dependency Tracking:**
    *   Test that calling `Set()` with a new value triggers a dependent `CreateEffect`.
    *   Test that calling `Set()` with the *same* value (e.g., `count.Set(5)` when the value is already 5) does *not* trigger a dependent `CreateEffect`.
    *   Test that multiple effects depending on the same signal are all triggered on change.
*   **Type Safety:**
    *   (Compile-time) Ensure `Set()` only accepts values of the type the signal was created with.

##### 4.1.2. `CreateEffect(fn func())`

*   **Initial Execution:**
    *   Test that the effect function runs once immediately upon creation.
*   **Dependency-based Re-execution:**
    *   Test that the effect re-runs when a single signal it depends on (`Get()` was called inside) is updated.
    *   Test that the effect re-runs when any of multiple signals it depends on are updated.
    *   Test that the effect does *not* re-run if a signal it does *not* depend on is updated.
*   **Cleanup:**
    *   Test that calling `Dispose()` on the effect prevents it from re-running when its dependencies change.
    *   Test that any cleanup functions registered via `OnCleanup` within the effect are executed when the effect is disposed.
*   **Nested Effects:**
    *   Test that an effect created inside another effect functions correctly and is re-created/re-run when the outer effect re-runs.
    *   Test that the inner effect's cleanup is called when the outer effect re-runs.

##### 4.1.3. `CreateMemo[T](fn func() T)`

*   **Lazy Evaluation & Caching:**
    *   Test that the memo's calculation function does not run if its value is never requested (`Get()`).
    *   Test that the calculation function runs only once when `Get()` is called multiple times and dependencies have not changed.
*   **Dependency Tracking:**
    *   Test that the memo re-calculates its value when a signal it depends on changes.
    *   Test that an effect depending on the memo is triggered when the memo's value changes.
    *   Test that the memo does *not* re-calculate if a signal it does not depend on changes.
*   **Chained Memos:**
    *   Test that a memo (`memo2`) that depends on another memo (`memo1`) updates correctly when `memo1`'s dependencies change.

##### 4.1.4. `OnCleanup(fn func())`

*   **Execution Timing:**
    *   Test that a cleanup function registered inside an effect is called when that effect is re-executed (just before the new execution).
    *   Test that the cleanup function is called when the effect is manually disposed.
*   **Scope:**
    *   Test that `OnCleanup` within a nested scope (e.g., inner effect) does not affect the cleanup of an outer scope.
    *   Test that multiple cleanup functions in the same scope are all executed.

#### 4.2. Component & DOM API (`pkg/uiwgo`)

##### 4.2.1. `Mount(elementID string, rootComponent func() Node)`

*   **DOM Attachment:**
    *   Test that the component's root node is successfully appended to the DOM element specified by `elementID`.
    *   Test that `Mount` panics or returns an error if the `elementID` does not exist in the DOM.
*   **Execution:**
    *   Test that the `rootComponent` function is executed exactly once during the mount process.

##### 4.2.2. `OnMount(fn func())`

*   **Execution Timing:**
    *   Test that the `OnMount` function is executed after the component's elements are attached to the main DOM.
    *   Test that it is executed only once for a given component instance.
*   **DOM Access:**
    *   Test that code inside `OnMount` can successfully access the component's rendered DOM elements.

##### 4.2.3. Control Flow

*   **`Show(props ShowProps)`:**
    *   **Initial State (True):** Test that if `When` signal is initially `true`, the `Children` are rendered immediately.
    *   **Initial State (False):** Test that if `When` signal is initially `false`, the `Children` are not rendered.
    *   **Transition False -> True:** Test that `Children` are rendered when the `When` signal changes from `false` to `true`.
    *   **Transition True -> False:** Test that `Children` are removed from the DOM when the `When` signal changes from `true` to `false`.
    *   **Cleanup:** Test that `OnCleanup` hooks within the `Children` are called when they are removed from the DOM.
*   **`For(...)` (Conceptual):**
    *   **Initial Render:** Test that the component correctly renders a list of items from a signal slice.
    *   **Adding Items:** Test that adding an item to the slice results in a new DOM element being added.
    *   **Removing Items:** Test that removing an item from the slice removes the corresponding DOM element.
    *   **Reordering Items (Keyed):** Test that reordering items in the slice reorders the DOM elements without re-creating them.
    *   **Cleanup:** Test that `OnCleanup` is called for components of items that are removed.

##### 4.2.4. Reactive Helpers

*   **`Text(fn func() string)`:**
    *   **Initial Render:** Test that the helper renders a text node with the initial string returned by `fn`.
    *   **Reactivity:** Test that the text node's content updates automatically when a signal used inside `fn` changes.
    *   **Non-Reactive:** Test that the text node does not update if no signals are used within `fn`.
