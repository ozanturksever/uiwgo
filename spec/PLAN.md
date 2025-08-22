# UiwGo Implementation Plan

This document outlines the step-by-step implementation plan for the UiwGo framework, based on `SPEC.md`. The plan is divided into two main parts: the core reactivity engine and the DOM component layer. Each step is designed to be small, incremental, and testable.

---

## Part 1: Core Reactivity API (`github.com/ozanturksever/uiwgo/reactivity`)

This is the foundational, platform-agnostic engine of the framework. All tests in this section can be standard Go tests (`go test`).

### Step 1.1: Project Setup and Basic Signal

*   **Goal:** Create the basic structure for the reactivity package and a non-reactive `Signal`.
*   **Actions:**
    1.  Create the package directory `reactivity` (module path: `github.com/ozanturksever/uiwgo/reactivity`).
    2.  Inside `reactivity`, create `signal.go`.
    3.  Define the `Signal[T]` interface with `Get() T` and `Set(T)`.
    4.  Implement `CreateSignal[T]` which returns a struct holding a value and an empty list of observers (effects).
    5.  `Get()` should simply return the current value.
    6.  `Set()` should simply update the value.
*   **Testing (`TestSignalInitial`):**
    *   Verify `Get()` returns the initial value.
    *   Verify `Set()` updates the value and a subsequent `Get()` returns the new value.

### Step 1.2: Basic Effect

*   **Goal:** Create an `Effect` that runs once but doesn't yet react to changes.
*   **Actions:**
    1.  Create `effect.go` in `pkg/reactivity`.
    2.  Define a global variable to hold the current `*effect` context.
    3.  Implement `CreateEffect(fn func())`. It should:
        *   Create an `effect` struct.
        *   Set the global context to this new effect.
        *   Run the provided function `fn` immediately.
        *   Clear the global context.
*   **Testing (`TestEffectInitialRun`):**
    *   Verify that the function passed to `CreateEffect` is executed once upon creation.

### Step 1.3: Wire Up Reactivity

*   **Goal:** Connect `Signal` and `Effect` to create a reactive data graph.
*   **Actions:**
    1.  Modify `Signal.Get()`: If there is a current effect in the global context, add that effect to the signal's list of observers.
    2.  Modify `Signal.Set()`:
        *   If the new value is the same as the old value, do nothing.
        *   If the value changes, iterate over the signal's observers and call their `execute()` method.
*   **Testing (`TestReactivity`):**
    *   Create a signal and an effect that reads from it.
    *   Verify the effect runs initially.
    *   Call `Set()` on the signal with a new value and verify the effect runs again.
    *   Call `Set()` with the *same* value and verify the effect does *not* run again.

### Step 1.4: Cleanup and Dispose Logic

*   **Goal:** Implement manual memory management for effects.
*   **Actions:**
    1.  Add a `disposed` flag to the `effect` struct.
    2.  Implement `Effect.Dispose()`, which sets the `disposed` flag.
    3.  Modify the effect execution logic to not run if it's disposed.
    4.  Implement `OnCleanup(fn func())`. It should add the cleanup function to a list within the current effect context.
    5.  When an effect is about to re-execute or is disposed, run all of its registered cleanup functions.
*   **Testing (`TestCleanupAndDispose`):**
    *   Verify `Dispose()` prevents an effect from re-running.
    *   Verify `OnCleanup` functions are called when an effect re-runs.
    *   Verify `OnCleanup` functions are called when an effect is disposed.

### Step 1.5: Memoization (`CreateMemo`)

*   **Goal:** Implement derived, cached signals.
*   **Actions:**
    1.  Implement `CreateMemo[T](fn func() T) Signal[T]`.
    2.  Internally, `CreateMemo` will:
        *   Create a new `Signal` to store the memoized value.
        *   Create an `Effect` that runs the user's calculation function `fn` and uses `Set()` to update the storage signal.
    3.  Return the storage signal.
*   **Testing (`TestMemo`):**
    *   Verify the calculation runs only once if `Get()` is called multiple times.
    *   Verify the memo updates when one of its dependencies changes.
    *   Verify an effect that depends on the memo updates when the memo's value changes.

---

## Part 2: Component & DOM API (`github.com/ozanturksever/uiwgo/comps`)

This layer builds on the reactive core. Tests in this section will require a browser environment (WASM).

### Step 2.1: Project Setup and Static Rendering

*   **Goal:** Render a static, non-reactive component to the DOM.
*   **Actions:**
    1.  Create the package directory `comps` (module path: `github.com/ozanturksever/uiwgo/comps`).
    2.  Create `comps.go` (or similar) and define `Node` and `ComponentFunc` types.
    3.  Implement a basic `Mount(elementID, rootComponent)` function.
    4.  Inside `Mount`, write a simple `gomponents` renderer using `syscall/js` that handles `createElement`, `setAttribute`, and `appendChild`. It does not need to handle updates yet.
    5.  Create an `examples/counter/main.go` and `index.html`.
*   **Testing (Manual):**
    *   Write a simple "Hello World" component.
    *   Compile to WASM and verify it renders correctly in the browser.

### Step 2.2: Reactive Text Node: BindText

*   **Goal:** Create a helper to render reactive text content directly in the DOM.
*   **Actions:**
    1.  Implement `BindText(fn func() string) Node`.
    2.  This function will:
        *   Create a DOM text node via `syscall/js`.
        *   Create a `gomponents.Node` that wraps this text node.
        *   Use `reactivity.CreateEffect` to run `fn` and update the text node's `nodeValue` whenever its dependencies change.
    3.  Return the wrapper node.
*   **Testing (Manual):**
    *   Update the counter example to use `comps.BindText` for the count display.
    *   Verify the number on the screen updates when the "Increment" button is clicked.

### Step 2.3: Lifecycle Hooks (OnMount, Effect, OnCleanup)

*   **Goal:** Provide lifecycle hooks for components to run side-effects and clean up.
*   **Actions:**
    1.  Implement `OnMount(fn func())`. For now, this can be a simple alias for `reactivity.CreateEffect` that runs once when the component is initialized.
    2.  Implement `Effect(fn func())` as a direct pass-through to `reactivity.CreateEffect` for tracking and running whenever its dependencies change.
    3.  Expose `reactivity.OnCleanup` by declaring `var OnCleanup = reactivity.OnCleanup`.
*   **Testing (Manual):**
    *   In the counter example, add an `OnMount` hook that prints a message to the console.
    *   Add an `Effect` that logs the current count whenever it changes.
    *   Verify cleanup is called when the component is removed (e.g., via a conditional).

### Step 2.4: Control Flow - `Show` Component

*   **Goal:** Implement conditional rendering.
*   **Actions:**
    1.  Implement `Show(props ShowProps) Node`.
    2.  The function will:
        *   Create a placeholder DOM node (e.g., a comment node).
        *   Create an effect that tracks `props.When`.
        *   Inside the effect:
            *   Run cleanup for the previous state.
            *   If `props.When.Get()` is true, render the `Children` and insert them into the DOM. Register a cleanup function to remove them.
            *   If false, do nothing (the cleanup has already removed them).
*   **Testing (Manual):**
    *   Create an example with a checkbox that toggles a signal.
    *   Use `Show` to conditionally display an element based on the checkbox's state.
    *   Verify the element appears and disappears correctly.
    *   Verify any `OnCleanup` hooks inside the `Show` component are fired.

### Step 2.5: Control Flow - `For` Component (Simple)

*   **Goal:** Implement a basic, non-keyed list rendering component.
*   **Actions:**
    1.  Define the `For` component signature. It will take a signal of a slice (`Signal[[]T]`) and a render function `func(T) Node`.
    2.  Implement a simple version that, inside an effect:
        *   Clears all previously rendered DOM nodes.
        *   Iterates over the current slice and renders a new DOM node for each item.
*   **Testing (Manual):**
    *   Create an example with a signal holding a slice of strings.
    *   Use `For` to render the list.
    *   Add buttons to add and remove items from the slice.
    *   Verify the DOM updates correctly.

### Step 2.7: Event Handling - `OnClick`

*   **Goal:** Provide a helper to attach Go functions as click event handlers on DOM elements.
*   **Actions:**
    1.  Implement `OnClick(fn func()) Node` that attaches a single-use JS function to the element's onclick.
    2.  Ensure the created JS callback is cleaned up using `OnCleanup` to avoid leaks.
    3.  Consider a generic `On(event string, fn func())` for future extensibility, but start with `OnClick`.
*   **Testing (Manual):**
    *   Use `OnClick` in the counter example to increment, decrement, and reset.
    *   Verify the handlers are called and that no memory leaks occur on teardown.

### Step 2.6: Control Flow - `For` Component (Keyed)

*   **Goal:** Improve the `For` component with efficient, keyed reconciliation.
*   **Actions:**
    1.  Modify the `For` component to accept a key for each item.
    2.  Implement a reconciliation algorithm (e.g., a variation of a simple block-diffing or mapping algorithm) that compares the previous list of items with the new one.
    3.  The algorithm should be able to identify items that are added, removed, or moved, and perform the minimum number of DOM operations.
*   **Testing (Manual):**
    *   Create an example where list items have state (e.g., an input field).
    *   Re-order the items in the source slice.
    *   Verify that the corresponding DOM elements are re-ordered *without* losing their state, proving they were not re-created.

---

## Part 3: Finalization

### Step 2.8: Component Composition Patterns

*   **Goal:** Demonstrate composition by splitting a feature into smaller components.
*   **Actions:**
    1.  Create a presentational header component (e.g., `AppHeader(title, subtitle string) Node`).
    2.  Split a feature component into display and controls (e.g., `CounterDisplay(getCount func() int)`, `CounterControls(onInc, onDec, onReset func())`).
    3.  Wire callbacks from the parent to children; keep state in the parent via `NewSignal/CreateSignal`.
*   **Testing (Manual):**
    *   Replace inline markup with composed components in the counter example and verify behavior.

### Step 3.1: Documentation and Final Example

*   **Goal:** Clean up the code and provide a comprehensive example.
*   **Actions:**
    1.  Add GoDoc comments to all public functions and types.
    2.  Create a `README.md` with build and run instructions.
    3.  Ensure the final `Counter` example in `examples/counter/main.go` is clean and demonstrates all major features (`Signal`, `Memo`, `Effect`, `OnMount`, `OnCleanup`, `BindText`, `OnClick`, composition).