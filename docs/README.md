# UIwGo Documentation

A browser-first UI runtime for Go targeting WebAssembly with gomponents-based type-safe HTML and fine-grained reactivity.

## Quick Navigation

### Getting Started
- [Overview](./overview.md) - Introduction and key concepts
- [Getting Started](./getting-started.md) - Quick setup and first example
- [Concepts](./concepts.md) - Core mental models and architecture

### Guides
- [Reactivity & State](./guides/reactivity-state.md) - Signals, Effects, Memos, Store, Resources
- [Control Flow](./guides/control-flow.md) - Show, For, Index, Switch, Dynamic components
- [Forms & Events](./guides/forms-events.md) - Event handling and form patterns
- [Lifecycle & Effects](./guides/lifecycle-effects.md) - Component lifecycle and effect management
- [Application Manager](./guides/application-manager.md) - Application lifecycle and management

### Reference
- [Core APIs](./api/core-apis.md) - Complete API documentation

### Additional Resources
- [Troubleshooting](./troubleshooting.md) - Common issues and solutions

## Project Structure

This documentation follows a task-oriented approach designed for three main audiences:

1. **Application Developers** - Building SPAs and interactive pages with Go + WebAssembly
2. **Integrators** - Using React/shadcn/ui components within the Go runtime
3. **Contributors** - Extending the runtime and maintaining the codebase

## Quick Start

For the fastest path to a working example:

```bash
# Clone and setup
git clone <repository>
cd uiwgo

# Run the counter example
make run counter

# Open http://localhost:8080 in your browser
```

See [Getting Started](./getting-started.md) for detailed setup instructions.

## Key Features

- **Fine-grained Reactivity** - Signals, Effects, Memos with automatic dependency tracking
- **Gomponents-based Rendering** - Type-safe HTML generation with post-render binding
- **WebAssembly Target** - Go code running efficiently in the browser
- **React Compatibility** - Bridge for using React/shadcn/ui components
- **Client-side Routing** - Full navigation API with nested routes
- **Comprehensive Testing** - Unit tests and browser E2E testing patterns
- **Developer Experience** - Hot reload dev server and Make-based workflows

## Documentation Phases

This documentation is being delivered in phases:

- **Phase 1** ✅ - Foundations (Overview, Getting Started, Concepts, Core Guides)
- **Phase 2** 🚧 - Application Building (Advanced Guides, Router, Examples)
- **Phase 3** 📋 - Ecosystem & Quality (React Compatibility, Performance, Testing)

Contributions and feedback are welcome!