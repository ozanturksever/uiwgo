.PHONY: build test serve run clean test-examples test-all test-example install dev-counter dev-todo dev-todo_store dev-component_demo dev-dom_integration dev-resource dev-router_demo dev-router_test

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
	lsof -ti:$(PORT) | xargs kill -9 || true

run: kill
	@echo "==> Starting Vite dev server for example: $(EXAMPLE) ..."
	@npm run dev:$(EXAMPLE)

# Install npm dependencies
install:
	@echo "==> Installing npm dependencies..."
	npm install

# Run Vite dev server for specific examples
dev-counter:
	@echo "==> Starting Vite dev server for counter example..."
	npm run dev:counter

dev-todo:
	@echo "==> Starting Vite dev server for todo example..."
	npm run dev:todo

dev-todo_store:
	@echo "==> Starting Vite dev server for todo_store example..."
	npm run dev:todo_store

dev-component_demo:
	@echo "==> Starting Vite dev server for component_demo example..."
	npm run dev:component_demo

dev-dom_integration:
	@echo "==> Starting Vite dev server for dom_integration example..."
	npm run dev:dom_integration

dev-resource:
	@echo "==> Starting Vite dev server for resource example..."
	npm run dev:resource

dev-router_demo:
	@echo "==> Starting Vite dev server for router_demo example..."
	npm run dev:router_demo

dev-router_test:
	@echo "==> Starting Vite dev server for router_test example..."
	npm run dev:router_test

clean:
	@rm -f examples/*/main.wasm || true

# Test configuration for js/wasm
# By default, auto-discover unit-testable packages under wasm (exclude examples and internal dev tooling)
# You can still override: make test PKG="./mypkg ./other" or filter names with RUN
PKG ?= $(shell go list ./... | grep -v '/examples/' | grep -v '/internal/' | tr '\n' ' ')
RUN ?=
# Minimized environment avoids wasm_exec.js command line/env length limits
TEST_ENV := env -i PATH="$(PATH)" HOME="$(HOME)" GOOS=js GOARCH=wasm

# Run tests under js/wasm
# Usage examples:
#   make test                         # tests all packages under wasm
#   make test PKG=./reactivity        # tests one package
#   make test RUN=TestName            # filter tests by name
#   make test PKG=./reactivity RUN=Signal
test:
	@echo "==> Running WASM tests for $(PKG) ..."
	@set -e; trap '$(MAKE) clean' EXIT INT TERM; $(TEST_ENV) go test $(PKG) $(if $(RUN),-run $(RUN),)

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
