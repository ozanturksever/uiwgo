### Goal
Create the initial implementation of “UiwGo” in this repository: a Go/WASM reactive UI micro-framework inspired by SolidJS principles, using gomponents for declarative structure and a fine‑grained reactivity core. Implement the reactivity engine and a minimal DOM/component layer sufficient to run a Counter example in the browser.

### Tech & Constraints
- Language: Go (go 1.24.3)
- Target: WebAssembly (GOOS=js, GOARCH=wasm)
- Module: github.com/ozanturksever/uiwgo
- UI markup: maragu.dev/gomponents (see spec/gomponents-doc.md)
- Philosophy: No VDOM, components run once, fine‑grained signals/effects/memos, direct DOM updates via syscall/js
- Event handling: Define Go functions and bind via js.FuncOf / js.Global(), or provide thin helpers to attach events with proper cleanup

### Deliverables (Incremental)
Implement according to the spec and plan in the spec/ directory:
- Core reactivity package: github.com/ozanturksever/uiwgo/reactivity
    - Signal[T], CreateSignal, CreateEffect, Effect.Dispose, CreateMemo, OnCleanup
- Component/DOM package: github.com/ozanturksever/uiwgo/comps
    - Node alias, ComponentFunc[P], Mount(elementID string, root func() Node)
    - Lifecycle helpers: OnMount(fn), OnCleanup passthrough
    - Reactive text helper: BindText(fn func() string)
    - Control flow: Show(props)
    - (Optional initial) OnClick helper for events (with cleanup)
- Example: examples/counter with main.go and index.html that mounts a Counter demo and uses BindText + event wiring
- Basic tests for reactivity per SPEC.md test scenarios

### High‑Level Architecture
- reactivity: platform agnostic graph of Signals and Effects (and Memos). Effects track dependencies via a global current effect context set while running.
- comps: one‑time component function produces initial gomponents tree, rendered to real DOM nodes via custom renderer using syscall/js. Updates are applied surgically by effects that specifically target nodes/attributes/text, not by re‑rendering components.

### Implementation Plan (follow these steps)
1) reactivity (pkg path: github.com/ozanturksever/uiwgo/reactivity)
- Files:
    - reactivity/signal.go
    - reactivity/effect.go
    - reactivity/memo.go
    - reactivity/cleanup.go
    - reactivity/signal_test.go, effect_test.go, memo_test.go, cleanup_test.go
- API:
    - type Signal[T any] interface { Get() T; Set(T) }
    - func CreateSignal[T any](initial T) Signal[T]
    - type Effect interface { Dispose() }
    - func CreateEffect(fn func()) Effect
    - func CreateMemo[T any](fn func() T) Signal[T]
    - func OnCleanup(fn func())
- Behavior details:
    - CreateEffect sets a global current effect, runs fn, unsets. While fn runs, any Signal.Get registers dependency on that effect.
    - Signal.Set: if value unchanged, no-op; else schedule/run dependent effects (synchronously is fine for MVP). Ensure disposed effects don’t run.
    - OnCleanup: within current effect scope, register cleanup functions; run them before re-executing an effect and when disposing.
    - CreateMemo: returns a Signal backed by an effect that recomputes value when dependencies change, caching the value; only recompute when needed.
- Tests (based on SPEC.md 4.1): initial Get/Set, dependency tracking, no re-run on same value, Dispose(), OnCleanup timing, memo caching, chained memos.

2) comps (pkg path: github.com/ozanturksever/uiwgo/comps)
- Files:
    - comps/comps.go (types, Mount)
    - comps/render.go (gomponents -> DOM renderer)
    - comps/helpers.go (BindText, OnMount alias, OnCleanup passthrough)
    - comps/controlflow.go (Show component, scaffold For for later)
- Types:
    - type Node = gomponents.Node
    - type ComponentFunc[P any] func(props P) Node
- Mount(elementID string, root func() Node):
    - Lookup document.getElementById(elementID)
    - Build gomponents tree via root()
    - Render to DOM: implement minimal renderer for gomponents elements/attributes/text using document.createElement, setAttribute, appendChild, and text nodes. MVP can assume Nodes are elements/attributes/text as produced by gomponents.
- BindText(fn func() string) Node:
    - Create a DOM text node and a small wrapper Node to attach it
    - CreateEffect that sets node.nodeValue = fn() whenever dependencies change
- Lifecycle helpers:
    - func OnMount(fn func()) { reactivity.CreateEffect(fn) } (MVP semantics: run after initial render; if needed, schedule to run after first DOM insert)
    - var OnCleanup = reactivity.OnCleanup
- Control flow Show:
    - type ShowProps struct { When reactivity.Signal[bool]; Children Node }
    - Show renders a placeholder (e.g., comment or empty span) and an effect:
        - On change: cleanup previous children; if When true, render Children after placeholder and register cleanup to remove them
- Optional events MVP:
    - func OnClick(fn func()) Node: attach a js.Func callback to the element’s onclick; release via OnCleanup to avoid leaks

3) Example app (examples/counter)
- Files:
    - examples/counter/main.go
    - examples/counter/index.html
- Behavior:
    - Create a signal count, a memo for doubleCount, an effect logging count
    - Expose increment function via js.FuncOf + js.Global().Set OR use comps.OnClick helper on a button
    - Use comps.BindText to update text reactively
    - Mount("app", root)
- index.html loads wasm_exec.js from repo root and main.wasm from same dir; include live reload snippet if desired

### Signatures and Samples to Follow (from spec/SPEC.md)
- Reactivity API:
    - Signal[T] { Get() T; Set(T) }
    - CreateSignal[T](initial T) Signal[T]
    - CreateEffect(fn func()) Effect, Effect.Dispose()
    - CreateMemo[T](fn func() T) Signal[T]
    - OnCleanup(fn func())
- Component API:
    - type ComponentFunc[P any] func(props P) Node
    - func Mount(elementID string, rootComponent func() Node)
    - func OnMount(fn func())
    - var OnCleanup = reactivity.OnCleanup
    - func BindText(fn func() string) Node
    - Show(props ShowProps) Node
- Example structure in SPEC.md shows Counter with BindText updating text node and onclick calling a function published on window.

### Testing Requirements (minimum)
Implement table-driven or simple unit tests for reactivity:
- CreateSignal: initial value, update value, no-trigger on same value
- CreateEffect: runs once at creation; re-runs when dependencies change; not when unrelated signals change; Dispose prevents re-run; OnCleanup called on re-run and dispose
- CreateMemo: lazy computation, caching, recompute on deps change, chained memos propagate

### Rendering Notes
- gomponents produces a Node tree that normally writes HTML to io.Writer. For WASM DOM, implement a renderer that interprets gomponents.Node instances:
    - Create elements for html.El, set attributes for html.Attr, create text nodes for Text
    - Append children in order
    - For MVP, support common subset needed by examples (Div, H1, P, Button, Class, Style, generic Attr, Text, Textf)
- For reactive updates, do not re-render components; rely on effects like BindText to target specific DOM nodes.

### Event Handling Notes
- Prefer Go-defined functions exposed to JS:
    - incFn := js.FuncOf(func(this js.Value, args []js.Value) any { count.Set(count.Get()+1); return nil })
    - js.Global().Set("incrementCounter", incFn)
    - In markup: Attr("onclick", "window.incrementCounter()")
- Or implement comps.OnClick(fn) to attach and release a per-element js.Func via OnCleanup

### File/Directory Layout to Create
- reactivity/
    - signal.go, effect.go, memo.go, cleanup.go, *_test.go
- comps/
    - comps.go, render.go, helpers.go, controlflow.go
- examples/counter/
    - main.go, index.html
- wasm_exec.js already present at repo root (use it in example)

### Build & Run (for example)
- Compile: GOOS=js GOARCH=wasm go build -o examples/counter/main.wasm ./examples/counter
- Serve a static file server rooted at examples/counter (or use the provided spec/dev.go tool if adapted later)
- Open http://localhost:8080/index.html

### Acceptance Criteria
- go test ./... passes for reactivity package tests
- Opening examples/counter/index.html shows a counter app that:
    - Displays count via BindText and updates reactively without component re-run
    - Button increments the count via Go-defined function
    - Console logs demonstrate Effect and OnMount behavior
- No panics, memory leaks prevented via OnCleanup for attached js.Func handlers

### References
- spec/SPEC.md: Philosophy, API, counter code sample, test scenarios
- spec/PLAN.md: Step-by-step actions and milestones
- spec/gomponents-doc.md: How gomponents Nodes, elements, and attributes work
- spec/example.go: Illustrative composition, BindText, OnClick usage and structure

### Please proceed to implement
- Follow the steps above in order.
- Prefer small, composable functions with clear comments.
- Add GoDoc for public APIs.
- If any ambiguity arises, align with SPEC.md priorities: components run once, rely on signals/effects for updates, keep reactivity core platform-agnostic, and favor direct DOM ops over VDOM.
