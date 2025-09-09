---
trigger: always_on
alwaysApply: true
---

Following is the guide/rules to run/develop the project

JS/DOM Interop Preference
- Prefer honnef.co/go/js/dom/v2 for browser DOM and Web APIs when relevant.
- Rationale: it is statically typed and provides better safety and discoverability compared to syscall/js.
- Use syscall/js only when an API is unavailable in honnef.co/go/js/dom/v2 or when dynamic JS interop is strictly required.

Logging Guidelines
- Always use the logutil package for logging instead of fmt.Println or console.log.
- Rationale: logutil provides cross-platform logging that works correctly in both standard Go builds and WebAssembly/browser environments.
- Available functions:
    - logutil.Log(args ...any): Logs arguments to console (browser) or stdout (standard Go)
    - logutil.Logf(format string, args ...any): Formatted logging with printf-style formatting
- The logutil package automatically handles JS/WASM vs standard Go builds through build tags.
- Safe to use with any mix of Go values, JS values, and primitive types.


- Following, how you run the dev server, build and test individual examples, execute the full test suite, and add new examples without modifying build scripts.

Prerequisites
- Go toolchain with WebAssembly support (GOOS=js, GOARCH=wasm).
- A local web browser; for browser-driven tests, a Chromium/Chrome installation is recommended for headless automation.
- Port 8080 free for the dev server (use the provided kill command if needed).

Quick Commands
- After changing code in an example, run make build <example> (or make build EX=<example>) to quickly find compile-time errors.
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
- shell is ZSH, so use single quote for arguments like following: go test -tags='!js !wasm' ./router -run TestRouterUpdatesViewOnRouteChange -v


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
    - Run specific test within an example: make test-<example> RUN=<TestName>
    - Example: make test-router_demo RUN=TestRouterDemo_HomePageRender
  - All examples:
    - make test-examples
  - The examples list is discovered automatically from the examples directory; adding a new example folder makes it part of test-examples without any Makefile changes.
  - Debugging failing tests:
    - When you have a failing test, run it individually first to isolate and fix the issue
    - Use the RUN parameter to target specific test functions for faster debugging cycles
- Browser Test Requirements:
    - All examples MUST have browser tests that validate real user interactions
    - Tests should cover component rendering, state changes, user input, and navigation
    - Use chromedp for authentic browser automation (not mocked DOM)
    - Follow the devserver pattern for consistent test infrastructure
    - Test both success and error scenarios where applicable
- Testing Helpers (internal/testhelpers):
    - **ChromedpConfig**: Configurable browser setup with sensible defaults
      - DefaultConfig(): Headless mode with 30s timeout, optimized for CI
      - VisibleConfig(): Visible browser for debugging tests
      - ExtendedTimeoutConfig(): 60s timeout for complex tests
    - **NewChromedpContext()**: Creates properly configured chromedp context
      - Handles context cleanup automatically
      - Supports custom Chrome flags and options
      - Use MustNewChromedpContext() for test setup that should fail fast
    - **CommonTestActions**: Reusable test actions via Actions global
      - WaitForWASMInit(): Waits for WASM initialization with configurable delay
      - NavigateAndWaitForLoad(): Navigate and wait for page load
      - ClickAndWait(): Click element with wait
      - SendKeysAndWait(): Send keys with wait
    - **Usage Pattern**: Import "github.com/ozanturksever/uiwgo/internal/testhelpers"
      ```go
      chromedpCtx := testhelpers.MustNewChromedpContext(testhelpers.DefaultConfig())
      defer chromedpCtx.Cancel()
      err := chromedp.Run(chromedpCtx.Ctx, testhelpers.Actions.NavigateAndWaitForLoad(url, "body"))
      ```
- Full suite:
  - make test-all runs unit tests first, then all example browser tests.

Adding a New Example
- Create a folder under examples/, for example examples/my_feature/.
- Add your Go entry point (e.g., main.go) and required browser-driven tests.
- **Browser Tests are Mandatory**: Every example MUST include comprehensive browser tests in main_test.go:
  - Use the devserver pattern with `//go:build !js && !wasm` constraint
  - Start a local server using `devserver.NewServer("example_name", "localhost:0")`
  - Use chromedp for real browser automation testing
  - Test all interactive functionality, component rendering, and user workflows
  - Follow the pattern from existing examples like counter, todo, resource, etc.
  - Include multiple test functions to cover different scenarios
- Example browser test structure:
  ```go
  //go:build !js && !wasm
  
  package main
  
  import (
      "testing"
      "github.com/chromedp/chromedp"
      "github.com/ozanturksever/uiwgo/internal/devserver"
      "github.com/ozanturksever/uiwgo/internal/testhelpers"
  )
  
  func TestMyFeature(t *testing.T) {
      server := devserver.NewServer("my_feature", "localhost:0")
      if err := server.Start(); err != nil {
          t.Fatalf("Failed to start dev server: %v", err)
      }
      defer server.Stop()
      
      // Use testhelpers for consistent chromedp setup
      chromedpCtx := testhelpers.MustNewChromedpContext(testhelpers.DefaultConfig())
      defer chromedpCtx.Cancel()
      
      err := chromedp.Run(chromedpCtx.Ctx,
          testhelpers.Actions.NavigateAndWaitForLoad(server.URL(), "body"),
          // ... additional test actions
      )
      if err != nil {
          t.Fatalf("Test failed: %v", err)
      }
  }
  ```
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

Documentation Lookup
- Use Context7 to fetch the latest documentation when it is available.
- For Go packages, use Go Docs (pkg.go.dev) to get the latest package documentation.
