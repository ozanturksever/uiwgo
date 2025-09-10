# Action System Performance Guide

This document provides a realistic overview of the performance characteristics of the UIwGo Action System. It clarifies which features are available in different build environments (standard Go vs. WebAssembly) and offers guidance on writing efficient, responsive code.

**Key Takeaway**: The advanced performance optimization system (including object pooling, reactive batching, and a microtask scheduler) is **only available in standard Go builds (`!js && !wasm`)**. For the primary target of WebAssembly (`js/wasm`), the system relies on a simpler, more lightweight dispatch mechanism.

---

## 1. Performance Model by Build Target

### WebAssembly (`js/wasm`) Performance
- **Dispatch**: Uses a standard, synchronous dispatch mechanism (`bus.go`). Asynchronous dispatch (`WithAsync()`) is handled by spawning a new goroutine for each async action.
- **No Advanced Optimizations**: The features described below are **NOT** available in WASM builds due to the `//go:build !js && !wasm` constraint on `action/performance.go`:
    - **No Object Pooling**: `Action` and `Context` objects are allocated for each dispatch.
    - **No Reactive Batching**: Signal updates are immediate and not batched.
    - **No Microtask Scheduler**: Async dispatch uses `go dispatchSync(...)`, which lacks sophisticated queueing or concurrency management.
- **Performance Focus**: In WASM, performance relies on the efficiency of the Go runtime's scheduler and garbage collector, along with careful application-level design (e.g., avoiding high-frequency dispatches on critical paths).

### Standard Go (`!js && !wasm`) Performance
- **Optimized Dispatch**: Can use a highly optimized dispatch path (`OptimizedDispatch`) that leverages a full suite of performance features.
- **All Features Available**:
    - **Object Pooling**: Reuses `Action`, `Context`, and subscriber slice objects to reduce GC pressure. Enabled via `PerformanceConfig`.
    - **Reactive Batching**: Batches reactive signal updates to coalesce UI re-renders, preventing layout thrashing.
    - **Microtask Scheduler**: Manages async operations through a fixed-size worker pool, providing predictable concurrency and backpressure.
    - **Profiling**: Includes detailed, built-in profiling hooks to measure dispatch time, allocation counts, and more.

---

## 2. Configuration and Usage Reality

The `PerformanceConfig` struct and related functions (`DefaultPerformanceConfig`, `EnablePerformanceOptimizations`) are defined in `action/performance.go` and are therefore **not accessible in WASM builds**. Any code attempting to use them will fail to compile under `GOOS=js GOARCH=wasm`.

#### Incorrect Usage Example (Will Not Compile in WASM)

```go
// DO NOT DO THIS in a WASM project. This code will only compile in a standard Go environment.
import "github.com/ozanturksever/uiwgo/action"

// This will cause a compile error: undefined: action.DefaultPerformanceConfig
config := action.DefaultPerformanceConfig()
config.EnableObjectPooling = true

// This will also cause a compile error: undefined: action.EnablePerformanceOptimizations
action.EnablePerformanceOptimizations(config)
```

---

## 3. Practical Performance Considerations for WASM

Since the advanced features are unavailable, developers targeting WebAssembly must focus on application-level patterns:

- **Avoid High-Frequency Actions**: Be cautious with actions tied to frequent events like `onmousemove` or `onscroll`. If necessary, implement manual debouncing or throttling within your application logic.
- **Use `distinctUntilChanged`**: For subscriptions that drive UI updates, use the `WithDistinctUntilChanged()` option to prevent handlers from running if the action payload has not changed. This is a crucial tool for preventing unnecessary re-renders.

    ```go
    // This subscription handler will only run if the payload is different from the last one.
    bus.Subscribe("ui.update", func(a action.Action[string]) error {
        // ... update UI
        return nil
    }, action.WithDistinctUntilChanged())
    ```

- **Leverage Asynchronous Dispatch Wisely**: `WithAsync()` is available in WASM, but it simply runs the dispatch in a new goroutine. It is useful for offloading non-critical tasks from the main UI thread but does not provide queueing or cancellation.

- **Mind Your Allocations**: Since object pooling is not active, be mindful of creating large or complex payloads in performance-critical paths.

---

## 4. Benchmarks and Their Relevance

The benchmark results previously cited in this document were generated in a standard Go environment and reflect the capabilities of the optimized, non-WASM system. **They are not representative of performance in a WebAssembly environment.**

New benchmarks specific to the `js/wasm` target are needed to accurately measure performance and identify bottlenecks relevant to browser execution.

### Running Benchmarks
To run the existing benchmarks (in a non-WASM environment):

```bash
# Run all action system benchmarks
go test -bench=. -benchmem ./action

# Run a specific benchmark
go test -bench=BenchmarkDispatchSingleSubscriber -benchmem ./action
```

---

## 5. Summary and Recommendations

| Feature                  | WASM (`js/wasm`) Status | Standard Go Status | Notes for WASM Developers                                       |
|--------------------------|-------------------------|--------------------|-----------------------------------------------------------------|
| **Object Pooling**       | 🔴 **Not Available**    | ✅ **Available**    | Be mindful of allocation hotspots in your own code.             |
| **Reactive Batching**    | 🔴 **Not Available**    | ✅ **Available**    | Signal updates are immediate; use `distinctUntilChanged`.       |
| **Microtask Scheduler**  | 🔴 **Not Available**    | ✅ **Available**    | `WithAsync()` uses a simple `go` call.                          |
| **Profiling Hooks**      | 🔴 **Not Available**    | ✅ **Available**    | Use browser dev-tools for profiling.                            |
| **OptimizedDispatch**    | 🔴 **Not Available**    | ✅ **Available**    | The standard `bus.Dispatch` is the only option.                 |
| **Observability/Logging**| ✅ **Available**        | ✅ **Available**    | The core observability hooks (`instrumentDispatch`) are active. |

**Conclusion**: The Action System's documentation previously described a highly advanced performance system that is not available for the primary WebAssembly target. This guide corrects that by clarifying the feature gap. For WASM development, focus on standard Go performance practices and leverage subscription options like `WithDistinctUntilChanged` to build efficient applications.