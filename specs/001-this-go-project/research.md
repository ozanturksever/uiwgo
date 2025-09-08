# Research: React Compatibility Layer

## Decision
We will use a JavaScript-based bridge to render React components and communicate with the Go WASM module. The state will be synchronized using callbacks and Go signals.

## Rationale
Directly manipulating React components from Go (WASM) is complex and not well-supported. A JS bridge provides a clear separation of concerns and leverages the strengths of both environments: React for rendering and Go for application logic. This approach is also more performant than alternatives that involve complex data serialization on every state change.

## Alternatives Considered
- **Using a Go-to-JavaScript transpiler (e.g., GopherJS):** This adds another layer of abstraction and potential performance overhead. It also makes it harder to use the existing Go codebase.
- **Re-implementing the React rendering logic in Go:** This is a massive undertaking and would be difficult to keep up-to-date with React's evolution.
