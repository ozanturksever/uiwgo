AGENTS.md — Running, Testing, and Dev Server Features

Audience and Purpose
- This guide enables automation agents to run the dev server, build and test individual examples, execute the full test suite, and add new examples without modifying build scripts.

Prerequisites
- Go toolchain with WebAssembly support (GOOS=js, GOARCH=wasm).
- A local web browser; for browser-driven tests, a Chromium/Chrome installation is recommended for headless automation.
- Port 8080 free for the dev server (use the provided kill command if needed).

Quick Commands
- Run dev server for an example:
  - make run                 # defaults to 'counter'
  - make run <example>       # e.g., make run todo
  - make run EX=<example>    # e.g., make run EX=todo
- Build a specific example to WebAssembly:
  - make build <example>
  - make build EX=<example>
- Unit tests under js/wasm (core packages):
  - make test
  - make test PKG=./reactivity
  - make test RUN=Signal
  - make test PKG=./reactivity RUN=Signal
- Browser tests for examples:
  - Single example:
    - make test-example <example>
    - make test-example EX=<example>
    - or pattern form: make test-<example>   # e.g., make test-counter
  - All examples (auto-discovered from examples/):
    - make test-examples
- Run everything (unit + all example browser tests):
  - make test-all
- Clean built WASM artifacts:
  - make clean
- Free port 8080 if in use:
  - make kill

Dev Server Capabilities
- Embedded wasm_exec.js:
  - Served automatically; no manual inclusion required during development.
- Auto-compile to WASM:
  - When the server starts for a chosen example, it compiles that example for WebAssembly (GOOS=js, GOARCH=wasm).
- Live reload:
  - Source changes trigger rebuilds and automatic browser reloads to reflect changes instantly.
- One-command workflow:
  - make run <example> launches the dev server for that example and prepares all assets.

Testing Strategy
- Unit tests (js/wasm):
  - Use make test to run core package tests under the js/wasm target.
  - Narrow scope with PKG and filter test names with RUN.
  - Examples:
    - make test PKG=./dom
    - make test RUN=MyTestName
- Browser-driven example tests:
  - Single example:
    - make test-example <example> (or make test-<example>)
  - All examples:
    - make test-examples
  - The examples list is discovered automatically from the examples directory; adding a new example folder makes it part of test-examples without any Makefile changes.
- Full suite:
  - make test-all runs unit tests first, then all example browser tests.

Adding a New Example
- Create a folder under examples/, for example examples/my_feature/.
- Add your Go entry point (e.g., main.go) and optional browser-driven tests.
- You can then:
  - Run it: make run my_feature
  - Build it: make build my_feature (outputs examples/my_feature/main.wasm)
  - Test it: make test-example my_feature
  - Include it in all-example runs automatically: make test-examples and make test-all (no edits required).

Operational Notes
- Example selection:
  - Commands accept the example name as a positional argument (make run todo) or via EX=todo.
  - If not provided, run defaults to the counter example.
- js/wasm test environment:
  - Unit tests use a minimized environment to avoid command-line length issues while preserving PATH and HOME.
- Cleaning:
  - make clean removes compiled WASM files for all examples.

Recommended Automation Flows
- Develop a single example:
  - make run <example>
- Quick unit test pass:
  - make test
- Validate one example’s browser tests:
  - make test-example <example>  (or make test-<example>)
- CI-like thoroughness:
  - make test-all

Examples (replace <example> with a folder under examples/)
- Start dev server: make run <example>
- Build just the WASM: make build <example>
- Test only this example: make test-example <example>
- Test everything: make test-all
