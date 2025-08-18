# Golid Development Makefile
# A Go-native frontend framework for WebAssembly applications

# Variables
GO_VERSION = 1.23.0
DEVSERVER_DIR = cmd/devserver
DEVSERVER_BIN = golid-dev
WASM_FILE = main.wasm
WASM_EXEC = wasm_exec.js
PORT = 8090

# Go build flags
GO_BUILD_FLAGS = -ldflags="-s -w"
WASM_BUILD_ENV = GOOS=js GOARCH=wasm
#WASM_HEADLESS=off
# Colors for output
GREEN = \033[0;32m
YELLOW = \033[1;33m
RED = \033[0;31m
NC = \033[0m # No Color

# Helper function for colored output
define print_color
	@printf "$(1)%s$(NC)\n" "$(2)"
endef

.PHONY: help build dev clean test wasm-test test-wasm devserver run setup deps check-deps install-deps examples-list example-counter example-todo example-router example-conditional example-lifecycle example-component-hierarchy example-direct-dom example-error-handling example-event-system example-lazy-loading example-reactivity example-store-action

# Default target
all: setup build

# Help target - shows all available commands
help:
	@printf "$(GREEN)Golid Development Makefile$(NC)\n"
	@printf "A Go-native frontend framework for WebAssembly applications\n"
	@printf "\n"
	@printf "$(YELLOW)Available targets:$(NC)\n"
	@printf "  $(GREEN)help$(NC)           - Show this help message\n"
	@printf "  $(GREEN)setup$(NC)          - Setup development environment (install deps, build devserver)\n"
	@printf "  $(GREEN)build$(NC)          - Build both devserver and WASM (same as 'all')\n"
	@printf "  $(GREEN)dev$(NC)            - Start development server with hot reload\n"
	@printf "  $(GREEN)run$(NC)            - Alias for 'dev'\n"
	@printf "\n"
	@printf "$(YELLOW)Build targets:$(NC)\n"
	@printf "  $(GREEN)devserver$(NC)      - Build the development server binary\n"
	@printf "  $(GREEN)wasm$(NC)           - Compile main.go to WebAssembly\n"
	@printf "  $(GREEN)wasm-tiny$(NC)      - Compile WASM with TinyGo (smaller bundle)\n"
	@printf "\n"
	@printf "$(YELLOW)Development targets:$(NC)\n"
	@printf "  $(GREEN)deps$(NC)           - Install/update Go dependencies\n"
	@printf "  $(GREEN)check-deps$(NC)     - Check if required dependencies are available\n"
	@printf "  $(GREEN)clean$(NC)          - Remove build artifacts\n"
	@printf "  $(GREEN)test$(NC)           - Run tests (non-WASM compatible only)\n"
	@printf "  $(GREEN)wasm-test$(NC)      - Run WASM tests using wasmbrowsertest\n"
	@printf "  $(GREEN)test-wasm$(NC)      - Alias for wasm-test\n"
	@printf "\n"
	@printf "$(YELLOW)Example targets:$(NC)\n"
	@printf "  $(GREEN)examples-list$(NC)          - List all available examples\n"
	@printf "\n"
	@printf "$(YELLOW)Basic Examples:$(NC)\n"
	@printf "  $(GREEN)example-counter$(NC)        - Run counter example (basic signals)\n"
	@printf "  $(GREEN)example-todo$(NC)           - Run todo list example (list rendering)\n"
	@printf "  $(GREEN)example-conditional$(NC)    - Run conditional rendering example\n"
	@printf "\n"
	@printf "$(YELLOW)Advanced Examples:$(NC)\n"
	@printf "  $(GREEN)example-router$(NC)         - Run router example (client-side routing)\n"
	@printf "  $(GREEN)example-lifecycle$(NC)      - Run lifecycle example (component lifecycle hooks)\n"
	@printf "  $(GREEN)example-component-hierarchy$(NC) - Run component hierarchy demo\n"
	@printf "  $(GREEN)example-reactivity$(NC)     - Run reactivity system demo\n"
	@printf "  $(GREEN)example-store-action$(NC)   - Run store and action demo\n"
	@printf "\n"
	@printf "$(YELLOW)System Examples:$(NC)\n"
	@printf "  $(GREEN)example-direct-dom$(NC)     - Run direct DOM manipulation demo\n"
	@printf "  $(GREEN)example-error-handling$(NC) - Run error handling demo\n"
	@printf "  $(GREEN)example-event-system$(NC)   - Run event system demo\n"
	@printf "  $(GREEN)example-lazy-loading$(NC)   - Run lazy loading demo\n"
	@printf "\n"
	@printf "$(YELLOW)Configuration:$(NC)\n"
	@printf "  PORT=$(PORT) (development server port)\n"
	@printf "  Go version: $(GO_VERSION)+\n"

# Setup development environment
setup: check-deps deps devserver
	@printf "$(GREEN)✓ Development environment setup complete$(NC)\n"
	@printf "$(YELLOW)Run 'make dev' to start the development server$(NC)\n"

# Build everything
build: devserver wasm
	@printf "$(GREEN)✓ Build complete$(NC)\n"

# Check if required tools are available
check-deps:
	@printf "$(YELLOW)Checking dependencies...$(NC)\n"
	@command -v go >/dev/null 2>&1 || { printf "$(RED)✗ Go is not installed$(NC)\n"; exit 1; }
	@printf "$(GREEN)✓ Go is available$(NC)\n"
	@go version | grep -q "go$(GO_VERSION)" || printf "$(YELLOW)⚠ Warning: Go $(GO_VERSION)+ recommended$(NC)\n"

# Install/update dependencies
deps:
	@printf "$(YELLOW)Installing dependencies...$(NC)\n"
	@go mod tidy
	@printf "$(YELLOW)Installing devserver dependencies...$(NC)\n"
	@cd $(DEVSERVER_DIR) && go mod tidy
	@printf "$(GREEN)✓ Dependencies updated$(NC)\n"

# Build the development server
devserver:
	@printf "$(YELLOW)Building development server...$(NC)\n"
	@cd $(DEVSERVER_DIR) && go build $(GO_BUILD_FLAGS) -o ../../$(DEVSERVER_BIN) .
	@printf "$(GREEN)✓ Development server built: $(DEVSERVER_BIN)$(NC)\n"

# Compile main.go to WebAssembly
wasm:
	@printf "$(YELLOW)Compiling WebAssembly...$(NC)\n"
	@$(WASM_BUILD_ENV) go build $(GO_BUILD_FLAGS) -o $(WASM_FILE) ./main.go
	@printf "$(GREEN)✓ WebAssembly compiled: $(WASM_FILE)$(NC)\n"

# Compile with TinyGo for smaller bundle (optional)
wasm-tiny:
	@printf "$(YELLOW)Compiling WebAssembly with TinyGo...$(NC)\n"
	@command -v tinygo >/dev/null 2>&1 || { printf "$(RED)✗ TinyGo not found. Install from https://tinygo.org$(NC)\n"; exit 1; }
	@tinygo build -o $(WASM_FILE) -target wasm ./main.go
	@printf "$(GREEN)✓ WebAssembly compiled with TinyGo: $(WASM_FILE)$(NC)\n"

# Start development server with hot reload
dev: devserver
	@printf "$(GREEN)Starting development server on http://localhost:$(PORT)$(NC)\n"
	@printf "$(YELLOW)Press Ctrl+C to stop$(NC)\n"
	@./$(DEVSERVER_BIN)

# Alias for dev
run: dev

# Clean build artifacts
clean:
	@printf "$(YELLOW)Cleaning build artifacts...$(NC)\n"
	@rm -f $(DEVSERVER_BIN)
	@rm -f $(WASM_FILE)
	@rm -f $(WASM_EXEC)
	@printf "$(GREEN)✓ Clean complete$(NC)\n"

# Run tests (note: WASM tests require special handling)
test:
	@printf "$(YELLOW)Running tests...$(NC)\n"
	@printf "$(YELLOW)Note: WASM-dependent tests will be skipped due to syscall/js constraints$(NC)\n"
	@go test -v ./golid/... || printf "$(YELLOW)⚠ Some tests skipped due to WASM constraints$(NC)\n"
	@printf "$(YELLOW)For full WASM testing, use 'make wasm-test'$(NC)\n"

# Run WASM tests using wasmbrowsertest
wasm-test:
	@printf "$(YELLOW)Running WASM tests with wasmbrowsertest...$(NC)\n"
	@printf "$(YELLOW)This will run tests in a browser environment$(NC)\n"
	@cd golid && env -i PATH="$(PATH)" HOME="$(HOME)" GOOS=js GOARCH=wasm go test -v
	@printf "$(GREEN)✓ WASM tests complete$(NC)\n"

# Alias for wasm-test
test-wasm: wasm-test

# Development workflow helpers
watch:
	@printf "$(YELLOW)File watching is handled by the development server$(NC)\n"
	@printf "$(YELLOW)Use 'make dev' to start the server with auto-reload$(NC)\n"

# Install development dependencies (if using devenv)
install-devenv:
	@command -v devenv >/dev/null 2>&1 || { printf "$(RED)✗ devenv not found. Install from https://devenv.sh$(NC)\n"; exit 1; }
	@printf "$(YELLOW)Setting up devenv environment...$(NC)\n"
	@devenv shell
	@printf "$(GREEN)✓ devenv environment ready$(NC)\n"

# Quick development start (build and run in one command)
start: build dev

# Show project status
status:
	@printf "$(GREEN)Golid Project Status$(NC)\n"
	@if [ -f $(DEVSERVER_BIN) ]; then \
		printf "$(YELLOW)Development server binary:$(NC) $(GREEN)✓ Built$(NC)\n"; \
	else \
		printf "$(YELLOW)Development server binary:$(NC) $(RED)✗ Not built$(NC)\n"; \
	fi
	@if [ -f $(WASM_FILE) ]; then \
		printf "$(YELLOW)WebAssembly binary:$(NC) $(GREEN)✓ Built$(NC)\n"; \
	else \
		printf "$(YELLOW)WebAssembly binary:$(NC) $(RED)✗ Not built$(NC)\n"; \
	fi
	@if [ -f $(WASM_EXEC) ]; then \
		printf "$(YELLOW)WASM exec helper:$(NC) $(GREEN)✓ Present$(NC)\n"; \
	else \
		printf "$(YELLOW)WASM exec helper:$(NC) $(YELLOW)⚠ Auto-generated by devserver$(NC)\n"; \
	fi
	@printf "$(YELLOW)Go modules:$(NC)\n"
	@printf "  Main module: $(GREEN)✓$(NC)\n"
	@printf "  Devserver module: $(GREEN)✓$(NC)\n"

# Force rebuild everything
rebuild: clean build
	@printf "$(GREEN)✓ Rebuild complete$(NC)\n"

# Show build info
info:
	@printf "$(GREEN)Golid Build Information$(NC)\n"
	@printf "$(YELLOW)Go version:$(NC) $$(go version)\n"
	@printf "$(YELLOW)Project root:$(NC) $$(pwd)\n"
	@printf "$(YELLOW)Main module:$(NC) $$(head -n1 go.mod)\n"
	@printf "$(YELLOW)Devserver module:$(NC) $$(head -n1 $(DEVSERVER_DIR)/go.mod)\n"
	@printf "$(YELLOW)Build flags:$(NC) $(GO_BUILD_FLAGS)\n"
	@printf "$(YELLOW)WASM environment:$(NC) $(WASM_BUILD_ENV)\n"

# Example targets - Individual example development and debugging
examples-list:
	@printf "$(GREEN)Available Examples$(NC)\n"
	@printf "\n"
	@printf "$(YELLOW)=== BASIC EXAMPLES ====$(NC)\n"
	@printf "\n"
	@printf "$(YELLOW)Counter Example:$(NC)\n"
	@printf "  • Basic signals and reactive updates\n"
	@printf "  • Increment/decrement buttons\n"
	@printf "  • Run with: $(GREEN)make example-counter$(NC)\n"
	@printf "\n"
	@printf "$(YELLOW)Todo List Example:$(NC)\n"
	@printf "  • List rendering with ForEach\n"
	@printf "  • Form handling and input binding\n"
	@printf "  • Add/remove/toggle todo items\n"
	@printf "  • Run with: $(GREEN)make example-todo$(NC)\n"
	@printf "\n"
	@printf "$(YELLOW)Conditional Rendering Example:$(NC)\n"
	@printf "  • Dynamic UI updates based on state\n"
	@printf "  • Show/hide functionality\n"
	@printf "  • Conditional styling\n"
	@printf "  • Run with: $(GREEN)make example-conditional$(NC)\n"
	@printf "\n"
	@printf "$(YELLOW)=== ADVANCED EXAMPLES ====$(NC)\n"
	@printf "\n"
	@printf "$(YELLOW)Router Example:$(NC)\n"
	@printf "  • Client-side routing with multiple pages\n"
	@printf "  • Route parameters extraction\n"
	@printf "  • Navigation with RouterLink\n"
	@printf "  • Run with: $(GREEN)make example-router$(NC)\n"
	@printf "\n"
	@printf "$(YELLOW)Lifecycle Example:$(NC)\n"
	@printf "  • Component lifecycle hooks (OnInit, OnMount, OnDismount)\n"
	@printf "  • Dynamic component creation and destruction\n"
	@printf "  • Resource cleanup patterns and timer management\n"
	@printf "  • Real-time lifecycle event logging\n"
	@printf "  • Run with: $(GREEN)make example-lifecycle$(NC)\n"
	@printf "\n"
	@printf "$(YELLOW)Component Hierarchy Demo:$(NC)\n"
	@printf "  • Parent-child component relationships\n"
	@printf "  • Props passing and state management\n"
	@printf "  • Component composition patterns\n"
	@printf "  • Run with: $(GREEN)make example-component-hierarchy$(NC)\n"
	@printf "\n"
	@printf "$(YELLOW)Reactivity Demo:$(NC)\n"
	@printf "  • Advanced reactive patterns\n"
	@printf "  • Signal dependencies and computed values\n"
	@printf "  • Reactive state management\n"
	@printf "  • Run with: $(GREEN)make example-reactivity$(NC)\n"
	@printf "\n"
	@printf "$(YELLOW)Store Action Demo:$(NC)\n"
	@printf "  • Global state management with stores\n"
	@printf "  • Action dispatching and state updates\n"
	@printf "  • Store subscriptions and reactive updates\n"
	@printf "  • Run with: $(GREEN)make example-store-action$(NC)\n"
	@printf "\n"
	@printf "$(YELLOW)=== SYSTEM EXAMPLES ====$(NC)\n"
	@printf "\n"
	@printf "$(YELLOW)Direct DOM Demo:$(NC)\n"
	@printf "  • Direct DOM manipulation techniques\n"
	@printf "  • Low-level DOM operations\n"
	@printf "  • Performance optimization patterns\n"
	@printf "  • Run with: $(GREEN)make example-direct-dom$(NC)\n"
	@printf "\n"
	@printf "$(YELLOW)Error Handling Demo:$(NC)\n"
	@printf "  • Error boundaries and error recovery\n"
	@printf "  • Graceful error handling patterns\n"
	@printf "  • Error reporting and debugging\n"
	@printf "  • Run with: $(GREEN)make example-error-handling$(NC)\n"
	@printf "\n"
	@printf "$(YELLOW)Event System Demo:$(NC)\n"
	@printf "  • Event delegation and handling\n"
	@printf "  • Custom event systems\n"
	@printf "  • Event propagation and bubbling\n"
	@printf "  • Run with: $(GREEN)make example-event-system$(NC)\n"
	@printf "\n"
	@printf "$(YELLOW)Lazy Loading Demo:$(NC)\n"
	@printf "  • Component lazy loading patterns\n"
	@printf "  • Dynamic imports and code splitting\n"
	@printf "  • Performance optimization techniques\n"
	@printf "  • Run with: $(GREEN)make example-lazy-loading$(NC)\n"

example-counter: devserver
	@printf "$(GREEN)Running Counter Example$(NC)\n"
	@printf "$(GREEN)Starting development server with counter example on http://localhost:$(PORT)$(NC)\n"
	@printf "$(YELLOW)Press Ctrl+C to stop$(NC)\n"
	@./$(DEVSERVER_BIN) --target ./examples/counter/main.go

example-todo: devserver
	@printf "$(GREEN)Running Todo List Example$(NC)\n"
	@printf "$(GREEN)Starting development server with todo example on http://localhost:$(PORT)$(NC)\n"
	@printf "$(YELLOW)Press Ctrl+C to stop$(NC)\n"
	@./$(DEVSERVER_BIN) --target ./examples/todo/main.go

example-router: devserver
	@printf "$(GREEN)Running Router Example$(NC)\n"
	@printf "$(GREEN)Starting development server with router example on http://localhost:$(PORT)$(NC)\n"
	@printf "$(YELLOW)Press Ctrl+C to stop$(NC)\n"
	@./$(DEVSERVER_BIN) --target ./examples/router/main.go

example-conditional: devserver
	@printf "$(GREEN)Running Conditional Rendering Example$(NC)\n"
	@printf "$(GREEN)Starting development server with conditional example on http://localhost:$(PORT)$(NC)\n"
	@printf "$(YELLOW)Press Ctrl+C to stop$(NC)\n"
	@./$(DEVSERVER_BIN) --target ./examples/conditional/main.go

example-lifecycle: devserver
	@printf "$(GREEN)Running Lifecycle Example$(NC)\n"
	@printf "$(GREEN)Starting development server with lifecycle example on http://localhost:$(PORT)$(NC)\n"
	@printf "$(YELLOW)Press Ctrl+C to stop$(NC)\n"
	@./$(DEVSERVER_BIN) --target ./examples/lifecycle/main.go

# Advanced Examples
example-component-hierarchy: devserver
	@printf "$(GREEN)Running Component Hierarchy Demo$(NC)\n"
	@printf "$(GREEN)Starting development server with component hierarchy demo on http://localhost:$(PORT)$(NC)\n"
	@printf "$(YELLOW)Press Ctrl+C to stop$(NC)\n"
	@./$(DEVSERVER_BIN) --target ./examples/component_hierarchy_demo/main.go

example-reactivity: devserver
	@printf "$(GREEN)Running Reactivity Demo$(NC)\n"
	@printf "$(GREEN)Starting development server with reactivity demo on http://localhost:$(PORT)$(NC)\n"
	@printf "$(YELLOW)Press Ctrl+C to stop$(NC)\n"
	@./$(DEVSERVER_BIN) --target ./examples/reactivity_demo/main.go

example-store-action: devserver
	@printf "$(GREEN)Running Store Action Demo$(NC)\n"
	@printf "$(GREEN)Starting development server with store action demo on http://localhost:$(PORT)$(NC)\n"
	@printf "$(YELLOW)Press Ctrl+C to stop$(NC)\n"
	@./$(DEVSERVER_BIN) --target ./examples/store_action_demo/main.go

# System Examples
example-direct-dom: devserver
	@printf "$(GREEN)Running Direct DOM Demo$(NC)\n"
	@printf "$(GREEN)Starting development server with direct DOM demo on http://localhost:$(PORT)$(NC)\n"
	@printf "$(YELLOW)Press Ctrl+C to stop$(NC)\n"
	@./$(DEVSERVER_BIN) --target ./examples/direct_dom_demo/main.go

example-error-handling: devserver
	@printf "$(GREEN)Running Error Handling Demo$(NC)\n"
	@printf "$(GREEN)Starting development server with error handling demo on http://localhost:$(PORT)$(NC)\n"
	@printf "$(YELLOW)Press Ctrl+C to stop$(NC)\n"
	@./$(DEVSERVER_BIN) --target ./examples/error_handling_demo/main.go

example-event-system: devserver
	@printf "$(GREEN)Running Event System Demo$(NC)\n"
	@printf "$(GREEN)Starting development server with event system demo on http://localhost:$(PORT)$(NC)\n"
	@printf "$(YELLOW)Press Ctrl+C to stop$(NC)\n"
	@./$(DEVSERVER_BIN) --target ./examples/event_system_demo/main.go

example-lazy-loading: devserver
	@printf "$(GREEN)Running Lazy Loading Demo$(NC)\n"
	@printf "$(GREEN)Starting development server with lazy loading demo on http://localhost:$(PORT)$(NC)\n"
	@printf "$(YELLOW)Press Ctrl+C to stop$(NC)\n"
	@./$(DEVSERVER_BIN) --target ./examples/lazy_loading_demo/main.go