# UiwGo Implementation Risks and Recommendations

This document provides an actionable checklist of recommendations to mitigate risks and guide a safe implementation. It is written for developers and reviewers to use during planning, implementation, and review.

---

## Executive Summary (Action Items)
- Feasibility: Implementable in pure Go (via WASM) — proceed with caution.
- Difficulty: Moderate–High — requires careful memory management, DOM interop discipline, and robust testing.
- Critical focus areas:
  - Memory and resource cleanup (effects, DOM nodes, JS callbacks)
  - Minimizing WASM↔JS boundary calls and batching DOM operations
  - Correctness and determinism of effect execution and dependency tracking
  - Keyed reconciliation for lists without state loss
  - Observability and debugging for WASM

---

## 1) Project Setup and Scope
- Define a minimal “walking skeleton” milestone:
  - Static render to DOM, one reactive text binding, one event handler, predictable teardown.
- Freeze initial scope for the first milestone: signals, effects, memos, static mount, BindText, OnClick, and simple Show.
- Create a “compatibility matrix” early:
  - Target browsers and versions, WASM runtime constraints, and supported features.
- Establish coding guidelines:
  - No DOM creation without registered cleanup.
  - No js.FuncOf without release/teardown path.
  - No long-running synchronous work in effects.

---

## 2) Core Reactivity (Signals, Effects, Memos)
- Signals
  - Ensure Set() is no-op on equal values (avoid redundant scheduling).
  - Provide value equality strategy per type (comparable vs custom comparator).
  - Maintain a stable observer list with safe iteration during mutation (copy-on-write or deferring).
- Effects
  - Use an effect stack (not a single global) to safely handle nested effects.
  - Track dependencies per effect; on re-run, re-subscribe exactly to the newly read signals.
  - Guard against infinite loops (e.g., iteration cap; warn in dev mode).
  - Support effect disposal and idempotent cleanup (cleanup can run multiple times without side effects).
- Memos
  - Recompute only on dependency changes; cache and avoid extra emissions when result didn’t change.
  - Ensure memo reads don’t create accidental cycles; document best practices.

---

## 3) Cleanup and Lifecycle
- Always run cleanup before an effect re-executes and when disposing.
- Cleanup responsibilities must include:
  - Removing DOM nodes created by that scope.
  - Releasing js.Func callbacks.
  - Detaching event listeners.
  - Breaking references to parent/children to aid GC.
- Provide a small “cleanup registry” utility:
  - Add(onDispose func()) returns a handle; Dispose() runs all once, ignores duplicates.
- Test matrix for lifecycle:
  - Effect re-run, conditional unmount (Show false), component removal, page navigation or host element replacement.

---

## 4) DOM Interop (syscall/js)
- Minimize boundary crossings:
  - Batch node creation/updates per tick; prefer local variables and set attributes in one go.
  - Avoid frequent textContent updates; coalesce updates within a micro-batch.
- DOM creation discipline:
  - Create elements/text nodes once; store references in the component scope.
  - Never “lose” a reference without cleanup; maintain a simple owner tree.
- Attribute and property setting:
  - Differentiate attributes vs properties (e.g., value vs setAttribute).
  - Normalize event handler attachment (add/removeEventListener) consistently.
- Error handling:
  - Wrap DOM calls with safe access patterns; add dev-mode assertions with descriptive messages.

---

## 5) Event Handling
- All js.FuncOf callbacks must be released during teardown.
- Centralize event binding:
  - Provide helpers that return a cleanup function that is auto-registered with OnCleanup.
- Avoid accidental multiple bindings:
  - Keep one binding per element per event unless explicitly needed.
  - On reactivity-driven changes, rebind only if needed (e.g., when handler identity changed).
- Document event propagation:
  - Default to bubbling; expose options and document implications.

---

## 6) Control Flow Components (Show, For)
- Show
  - Use a stable placeholder (e.g., comment node) and insert/remove sibling ranges.
  - Keep a list of currently inserted nodes and guarantee full teardown on hide.
  - Ensure nested cleanups run when toggling states.
- For (simple)
  - Start with non-keyed: full clear and rebuild within an effect.
  - Ensure teardown correctly removes all event handlers and js.Funcs for removed nodes.
- For (keyed)
  - Introduce only after correctness of simple version is proven.
  - Use a stable key map to preserve identity and state.
  - Implement minimal DOM operations (insert/move/remove) and verify state retention with interactive nodes (inputs).

---

## 7) Performance Strategy
- Scheduling
  - Defer effect executions within the same tick when multiple signals change; batch re-runs.
  - Prevent starvation: limit cascading updates per frame; surface warnings in dev builds.
- DOM update minimization
  - Memoize computed strings and attributes; avoid re-setting unchanged values.
  - For text bindings, only update when content changed.
- Measurement
  - Add counters for:
    - Effects executed per tick
    - DOM nodes created/removed
    - WASM↔JS calls per second
  - Provide dev tooling/logging to trace heavy updates.

---

## 8) Memory Management and Leak Prevention
- Standardize a “resource ownership” model:
  - Every created DOM node, JS function, and listener must have an owning scope.
  - Owners register disposers; on scope cleanup, execute all disposers.
- Watch for cycles:
  - Avoid capturing long-lived scopes in closures.
  - Break references in cleanup functions (set slices/maps to nil).
- Leak checks
  - Provide dev-only counters for live elements and callbacks.
  - Add a “leak test” page that mounts/unmounts components repeatedly and reports deltas.

---

## 9) Concurrency and Scheduling Considerations
- WASM in browsers is single-threaded by default:
  - Avoid blocking operations in effects; yield to the event loop for long work.
  - Consider microtask-style debouncing (using timeouts) for heavy recomputations.
- Maintain deterministic execution order:
  - Run effects in creation order or provide a priority model; document the behavior.
- Prevent re-entrancy issues:
  - Avoid running effects synchronously while mutating dependencies; queue executions.

---

## 10) Error Handling and Debuggability
- Developer mode
  - Verbose logs for effect registration, dependency changes, and cleanup.
  - Assertions for illegal states (e.g., updating disposed effects).
- Error boundaries (conceptual)
  - Wrap effect bodies in recover; surface meaningful diagnostics in console.
- Tracing
  - Optional global debug flag to print reactive graph changes and DOM operation summaries.

---

## 11) Testing Strategy
- Unit tests (pure Go)
  - Signals: get/set semantics, equality short-circuit.
  - Effects: dependency tracking, re-exec, disposal, cleanup order.
  - Memos: cache behavior, propagation, stabilization on unchanged results.
- WASM integration tests (browser)
  - Static render consistency.
  - Text bindings update on signal changes.
  - Event handlers fire, then release on teardown.
  - Show: toggling correctness and cleanup.
  - For: add/remove correctness; for keyed, reorder without losing state.
- Regression tests
  - Reproduction harness for bug reports; snapshot DOM structure or behaviors.
- Manual test pages
  - Performance stressors: many signals/effects, deep lists, rapid updates.

---

## 12) Browser Compatibility
- Define a supported set of browsers; confirm baseline WASM support.
- Guard against missing APIs; polyfill or fail with clear messages in dev builds.
- Validate behavior under throttled CPU and low-memory conditions.

---

## 13) Security Considerations
- Sanitize or strictly control any raw HTML injection helpers.
- Avoid exposing internal mutable state via global JS objects in production builds.
- Ensure event handlers don’t inadvertently leak sensitive data via logs.

---

## 14) Documentation and Developer Experience
- Provide a “How to reason about reactivity” guide:
  - One-time component run, effect-driven updates, and cleanup expectations.
- Include patterns and anti-patterns:
  - Do: compute inside memos/effects; Don’t: mutate and read same signal synchronously in one effect without guards.
- Offer migration/starter templates:
  - Minimal example, counter, list with keys, and teardown demo.
- Add a Troubleshooting section:
  - Common symptoms (double handlers, stale text, no updates) and fixes.

---

## 15) Release and Rollout Plan
- Phase 1: Core reactivity validated with unit tests.
- Phase 2: Static DOM mount and reactive text binding.
- Phase 3: Event handling with strict cleanup.
- Phase 4: Show and simple For; memory and performance baseline.
- Phase 5: Keyed For with state retention; optimize and harden.
- Phase 6: Documentation and example gallery; developer tooling.

---

## Quick Checklists

### Implementation PR Checklist
- No js.FuncOf without a matching Release() in cleanup.
- All created DOM nodes have a registered disposer.
- Effects dispose cleanly; cleanup is idempotent.
- Equality checks prevent unnecessary Set() and DOM updates.
- Logs and counters are behind a dev flag and off by default.

### Review Checklist
- Tests cover new reactive patterns and cleanup paths.
- Boundary crossing minimized and reasonably batched.
- No synchronous long-running work inside effects.
- Keyed list operations preserve state on reorders.
- Browser smoke tests pass on the target matrix.

---

Adhering to these recommendations will minimize risks in correctness, memory usage, and performance, and will improve developer productivity and maintainability over time.
