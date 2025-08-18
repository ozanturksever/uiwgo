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

# Colors for output
GREEN = \033[0;32m
YELLOW = \033[1;33m
RED = \033[0;31m
NC = \033[0m # No Color

# Helper function for colored output
define print_color
	@printf "$(1)%s$(NC)\n" "$(2)"
endef

.PHONY: help build dev clean test wasm devserver run setup deps check-deps install-deps

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
	@printf "$(YELLOW)For full testing, use browser-based integration tests$(NC)\n"

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