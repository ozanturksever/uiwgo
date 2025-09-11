.PHONY: build test serve run clean test-examples test-all test-example install

MAKEFLAGS += --no-print-directory

# Extract example name from command or variable
# Usage:
#   make run                    -> uses default 'counter'
#   make run counter            -> selects 'counter' (or any other example)
#   make run EX=counter         -> selects via variable
#   make build todo             -> builds 'todo' example
#   make test-example todo      -> tests 'todo' example
EX ?=
PORT ?= 8080
EXAMPLE_RAW := $(firstword $(filter-out run build test-example,$(MAKECMDGOALS)))
EXAMPLE := $(or $(EXAMPLE_RAW),$(EX),counter)

# Swallow extra goal (the example name) so make doesn't try to build it
ifneq (,$(filter run build test-example,$(firstword $(MAKECMDGOALS))))
  OTHER_GOALS := $(filter-out run build test-example,$(MAKECMDGOALS))
  .PHONY: $(OTHER_GOALS)
  $(OTHER_GOALS):
	@:
endif

# Auto-discover example directories under ./examples
EXAMPLE_DIRS := $(shell test -d examples && find examples -mindepth 1 -maxdepth 1 -type d -print 2>/dev/null | sort)
EXAMPLES := $(patsubst %/,%,$(notdir $(EXAMPLE_DIRS)))

# Default target: build selected example
build:
	@echo "==> Building WASM binary for example: $(EXAMPLE) ..."
	GOOS=js GOARCH=wasm go build -o examples/$(EXAMPLE)/main.wasm examples/$(EXAMPLE)/main.go

serve:
	@echo "==> Serving http://localhost:8080"
	go run ./server.go

kill:
	@lsof -ti:$(PORT) 2>/dev/null | xargs -r kill -9 2>/dev/null || true

run: kill
	@echo "==> Starting Vite dev server for example: $(EXAMPLE) ..."
	@if [ -f "examples/$(EXAMPLE)/vite.config.js" ]; then \
		npx vite -c examples/$(EXAMPLE)/vite.config.js; \
	else \
		echo "Error: vite.config.js not found for example '$(EXAMPLE)'"; \
		echo "Available examples: $(EXAMPLES)"; \
		exit 1; \
	fi

# Install npm dependencies
install:
	@echo "==> Installing npm dependencies..."
	npm install

# Note: Individual dev-* targets are no longer needed.
# Use 'make run <example>' instead, which auto-discovers examples.

clean:
	@rm -f examples/*/main.wasm || true

# Test configuration for js/wasm
# By default, auto-discover unit-testable packages under wasm (exclude examples and internal dev tooling)
# You can still override: make test PKG="./mypkg ./other" or filter names with RUN
PKG ?= $(shell GOOS=js GOARCH=wasm go list ./... | grep -v '/examples/' | grep -v '/internal/' | tr '\n' ' ')
RUN ?=
# Minimized environment avoids wasm_exec.js command line/env length limits
GO_BIN := $(shell command -v go)
GO_DIR := $(dir $(GO_BIN))
TEST_ENV := env -i PATH="$(GO_DIR):/usr/bin:/bin" GOOS=js GOARCH=wasm

# Run tests under js/wasm
# Usage examples:
#   make test                         # tests all packages under wasm
#   make test PKG=./reactivity        # tests one package
#   make test RUN=TestName            # filter tests by name
#   make test PKG=./reactivity RUN=Signal
test:
	@echo "==> Running WASM tests for $(PKG) $(GO_DIR)..."
	@set -e; trap '$(MAKE) clean' EXIT INT TERM; \
	for pkg in $(PKG); do \
	  echo "==> Testing $$pkg"; \
	  $(TEST_ENV) go test $$pkg $(if $(RUN),-run $(RUN),) || exit $$?; \
	done

# Generic browser test for a single example (accepts positional arg or EX variable)
test-example:
	@echo "==> Running browser tests for example: $(EXAMPLE) ..."
	@set -e; trap '$(MAKE) clean' EXIT INT TERM; go test ./examples/$(EXAMPLE) -v $(if $(RUN),-run $(RUN),)

# Pattern target to run browser tests for a named example: make test-foo
test-%:
	@echo "==> Running browser tests for $* example..."
	@set -e; trap '$(MAKE) clean' EXIT INT TERM; go test ./examples/$* -v $(if $(RUN),-run $(RUN),)

# Run all browser tests for discovered examples
test-examples:
	@echo "==> Running browser tests for all examples ($(EXAMPLES))..."
	@set -e; trap '$(MAKE) clean' EXIT INT TERM; \
	for ex in $(EXAMPLES); do \
	  echo "==> Running browser tests for $$ex example..."; \
	  go test ./examples/$$ex -v $(if $(RUN),-run $(RUN),); \
	done

# Run all tests (unit tests + browser tests)
test-all:
	@echo "==> Running all tests (unit + browser)..."
	@echo "==> Running WASM unit tests (excluding examples)..."
	@set -e; trap '$(MAKE) clean' EXIT INT TERM; $(TEST_ENV) go test $(PKG) $(if $(RUN),-run $(RUN),)
	@$(MAKE) test-examples
