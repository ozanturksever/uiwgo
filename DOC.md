# Developer Documentation

This repository provides a small, reactive UI toolkit for Go (compiled to WebAssembly) that lets you:
- Build UI as composable components.
- Use Signals, Effects, and Memos for fine‑grained reactivity.
- Render to HTML first, then activate behavior by scanning and “attaching” binders.
- Use higher‑level control‑flow components like Show, For, Index, Switch/Match, and Dynamic.
- Integrate cleanly with DOM operations in the browser.

The mental model is similar to libraries like SolidJS, but adapted to Go and gomponents. You render static HTML up front, then reactive binders update the DOM efficiently when state changes.

---

## 1) Start Here: The Big Picture

- You write components that return gomponents Nodes.
- At startup, you Mount a root component into a DOM element by id.
- Mount renders HTML, inserts it into the page, then runs a DOM scan to “attach” reactive behaviors.
- Reactive primitives (Signals, Effects, Memos) drive updates without rerendering the whole page.
- Control‑flow helpers (Show/For/Index/Switch/Dynamic) declaratively manage visibility, lists, branching, and dynamic component swapping.

Why this design?
- Initial HTML render is fast and simple.
- Reactive updates only touch exactly what changed.
- Binders keep DOM operations local and efficient.

---

## 2) Quick Start

- Use Go (js/wasm) toolchain.
- Include wasm_exec.js (from your Go distribution) in your host HTML.
- Build your wasm module and serve it over HTTP (file:// won’t work due to browser restrictions).

Typical steps:
1. Write your Go code to define a root component and call Mount("app", root).
2. Build for WebAssembly:
   - GOOS=js GOARCH=wasm go build -o main.wasm ./...
3. Reference wasm_exec.js and main.wasm in index.html, and start a local server.

Tip: Make sure the element with the id you pass to Mount (e.g., “app”) exists in the HTML.

---

## 3) Core Concepts

- Components: Functions that return gomponents Nodes. They compose naturally.
- Mount: Renders the component tree to HTML, injects it into the DOM, attaches binders, and then runs OnMount callbacks.
- Signals: Reactive variables with Get/Set. Effects re-run when signals they read change.
- Memos: Derived, cached values recomputed only when dependencies change.
- Cleanup: Register cleanups with OnCleanup to dispose resources tied to an effect or a mounted subtree.
- Store: Optional reactive store for structured state; lets you Select nested properties as signals for fine-grained subscriptions.

The key idea: Effects run your reactive logic. When a signal changes, only dependent effects re-run, producing minimal DOM updates via the attached binders.

---

## 4) Essential APIs at a Glance

- Mount(elementID string, root func() Node)
- OnMount(func())  // schedule a callback right after DOM is attached
- OnCleanup(func()) // register cleanup within an effect or reactive context

Reactive primitives:
- CreateSignal[T](initial T) -> Signal[T]
- CreateEffect(fn func()) -> Effect
- CreateMemo[T](calc func() T) -> Signal[T]
- Store: CreateStore[T](initial T) -> (Store[T], setState func(...any))

DOM-facing helpers:
- BindText(func() string) -> Node     // reactive text content
- BindHTML(func() Node) -> Node       // reactive HTML inside a container

Control flow:
- Show(ShowProps) -> Node
- For(ForProps[T]) -> Node            // keyed lists
- Index(IndexProps[T]) -> Node        // index-based lists
- Switch(SwitchProps) -> Node, Match(MatchProps) -> Node
- Dynamic(DynamicProps) -> Node       // reactively render a chosen component

You combine these to build UI that stays in sync with reactive state.

---

## 5) Common Patterns

- Displaying reactive text:
  - BindText(() => mySignal.Get()) updates text when the signal changes.
- Conditionally rendering UI:
  - Show({ When: someBoolSignal, Children: ... })
- Rendering lists:
  - For with a Key function for stable keyed reconciliation.
  - Index for index‑based reconciliation; pass getItem so each child reads the current item reactively.
- Branching UI:
  - Switch with Match cases to select a branch reactively.
- Swapping components:
  - Dynamic to render whichever component a signal or function returns.

---

## 6) Lifecycle Overview

- Initial render:
  - Mount renders gomponents Nodes to an HTML string.
  - The string is injected into the container element.
- Binder attachment:
  - The DOM is scanned for binder placeholders (data‑attributes).
  - Each binder sets up the necessary reactive Effects and event hooks.
  - Nested content is recursively scanned as it’s inserted.
- OnMount callbacks:
  - After DOM is attached and binders are active, OnMount callbacks run.
- Cleanup:
  - When effects re-run or elements are removed, OnCleanup callbacks fire, disposing resources and listeners.
  - Control‑flow components propagate cleanups to their subtrees.

---

## 7) Control Flow Components (How They Behave)

Show
- Renders children inside a placeholder when When is true; clears them when false.
- Uses a simple reactive Effect to toggle innerHTML and re‑attach binders for newly inserted content.

For (keyed lists)
- You provide Items (reactive slice or function), a Key(item) string, and Children(item, index).
- The binder keeps a map of key -> child record and updates minimally:
  - Insert new items at the right position.
  - Move existing DOM segments for reordered keys.
  - Remove DOM (and run cleanup) for dropped keys.
- This enables efficient list updates even for large data sets.

Index (index‑based lists)
- You provide Items and Children(getItem, index).
- Instead of passing the item value, Index passes a getter getItem() so each child reads the current item reactively, tracking changes by position.
- Useful when order is stable and you want per‑item reactivity by index.

Switch/Match
- Switch chooses one Match branch based on When.
- If the active branch changes, the previous branch DOM is cleared (with cleanup) and the new branch is rendered and activated.

Dynamic
- Renders a component chosen by a signal or function.
- When the component changes, the previous subtree is disposed (cleanup) and the new one is mounted and activated.

---

## 8) Reactivity Deep Dive

Signals
- Get registers the current Effect (if any) as a dependent.
- Set updates the value; if changed, all dependent Effects re‑run.

Effects
- CreateEffect(fn) runs fn immediately, tracking dependencies read during that run.
- When any dependency changes, the effect cleanly re-runs:
  - Prior cleanups execute.
  - Dependencies are recalculated.
  - New cleanups can be registered.

Memos
- A memo is a lazily computed, cached Signal.
- The first read computes the value; subsequent updates propagate if and only if the computed value changed.

Cleanup propagation
- Use OnCleanup within Effects or binders to dispose event listeners, timers, or DOM segments.
- Control‑flow components keep cleanups per subtree, so removing list items or switching branches disposes correctly.

---

## 9) How DOM Binding Works Internally

- HTML first: Components render to an HTML string (gomponents), inserted into the DOM container.
- Scan: A single attach step queries for data attributes (e.g., data‑uiwgo‑txt, data‑uiwgo‑html, data‑uiwgo‑show, list/switch/dynamic markers).
- Register: For each placeholder, the relevant binder sets up Effects and keeps references needed for updates.
- Update: When an Effect re‑runs (due to Signal changes), it updates DOM minimally:
  - Text: set textContent.
  - HTML containers: update innerHTML and re‑attach binders in the updated subtree.
  - Lists/branches/dynamic: perform insert/move/remove operations and re‑attach binders for newly inserted DOM.

Key guarantees
- Avoid duplicate attachment by marking nodes as “bound”.
- Maintain per‑subtree cleanup so removals are safe and leak‑free.
- Re‑scan only the subtree that changed to keep work bounded.

---

## 10) Project Layout

- comps/: Component and binder-facing APIs (Mount, Bind*, Show/For/Index/Switch/Dynamic, lifecycle hooks).
- reactivity/: Core reactive engine (Signal, Effect, Memo, Store, Cleanup).
- dom/: Interop and helpers for DOM operations and event integration.
- examples/: Example applications or snippets (if present).
- spec/ and tests: Validation of reactive behavior, list reconciliation, and memory/cleanup correctness.

You’ll mostly write application code against comps and reactivity. The dom package is used internally and can be extended for lower-level DOM needs.

---

## 11) Building and Running

Requirements
- Go toolchain with WebAssembly support.
- A static file server to serve index.html and the compiled wasm.

Typical integration
- Include wasm_exec.js from your Go toolchain in your index.html.
- Load main.wasm via a small bootstrapping script.
- Ensure the root element (e.g., <div id="app"></div>) is in your HTML.
- In Go, call Mount("app", RootComponent) from main.

Troubleshooting
- “document is not available”: Ensure code runs in a browser environment (not plain Go without js/wasm).
- Element not found: Verify the id you pass to Mount exists in HTML.
- Empty UI: Confirm your root component returns Nodes and that you serve over http(s), not file://.

---

## 12) Best Practices

- Prefer For with stable keys for dynamic lists; only use Index when order is stable and identity is by position.
- Use Memos for derived state that’s expensive to compute.
- Keep Effects small and focused; let Signals model state.
- Register OnCleanup for timers, subscriptions, and event listeners.
- Keep HTML structure valid:
  - When inserting list items under table elements or lists, ensure wrapper tags match the parent (e.g., use an “As” variant once available).

---

## 13) Performance Notes

- Minimal updates:
  - Signals limit effect re-execution to exactly what read them.
  - For uses keyed reconciliation to avoid wholesale rerenders.
- Avoid excessive innerHTML churn:
  - Prefer fine‑grained binders (text updates, item inserts/moves/removes) over full container rewrites.
- Batch updates if you trigger many Set calls in a tight loop (consider a batching helper if you add one later).

---

## 14) Extensibility: Adding a New Binder

High‑level approach
- Define a function that emits a placeholder element with a unique data attribute and an id.
- Store a registry entry keyed by that id with whatever context you need (signals, handlers).
- On attachBinders, query for your attribute, mark as bound, and create Effects to perform updates.
- On insertions, call the attach routine for newly added subtrees so nested binders activate.
- Register cleanups appropriately to avoid memory leaks when nodes are removed.

This pattern is consistent across BindText, BindHTML, Show, list controls, Switch, and Dynamic.

---

## 15) Testing Strategy

- Reactivity:
  - Signals trigger exactly dependent Effects.
  - Memos recompute only on dependency changes.
  - Cleanup is called before re-runs and on dispose.
- Lists:
  - Insertion, removal, middle inserts, moves, replacements, empty lists.
- Branching/Dynamic:
  - Switch between cases; fallback behavior; nested combinations.
  - Dynamic component swaps dispose the previous subtree correctly.
- Memory:
  - No leaks after removing items or swapping branches.
- DOM structure:
  - Ensure valid markup under tables and lists; provide “As” variants when necessary.

---

## 16) Roadmap (Highlights)

- Polished For/Index with item wrappers now; explore comment marker ranges later for finer control.
- Switch/Match stability, nested switching inside lists.
- Dynamic component improvements for typed props and memoization of heavy subtrees.
- “As” helpers for list controls to ensure valid HTML structure (e.g., ForAs/IndexAs).
- Performance pass:
  - Reduce innerHTML operations.
  - Improve move detection (e.g., LIS-based algorithms) if lists are very large.
- Developer docs and example apps: Todos, tables, chat lists.

---

## 17) FAQ

Q: Do I need a virtual DOM?
- No. The system uses direct DOM updates targeted by reactive effects and keyed diffing for lists.

Q: How do I debug what re-runs?
- Add logging inside CreateEffect and signal setters during development. Because dependencies are tracked automatically on Get, reactivity remains explicit and predictable.

Q: When should I choose Index over For?
- Use Index for stable, position-based lists where item identity does not matter. Prefer For with Key for dynamic lists where you insert/remove/move items frequently.

Q: Is server-side rendering supported?
- The current flow is browser-first with js/wasm. You render HTML and attach behaviors client-side.

---

## 18) Summary

- Think: render once, then bind.
- Use Signals and Effects for reactivity.
- Let control‑flow helpers manage conditional UI, lists, and dynamic swapping.
- Trust the attach-and-update pipeline to update only what changed.
- Use cleanups to keep resources tidy.

With these patterns, you can build responsive, fast UIs in Go that feel as ergonomic as modern JS frameworks while staying idiomatic to Go.
