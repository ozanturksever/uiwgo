.PHONY: build test serve run clean test-counter test-todo test-todo-store test-resource test-examples test-all

# Extract example name from command or variable
# Usage:
#   make run                 -> uses default 'counter'
#   make run counter         -> selects 'counter' (or any other example)
#   make run EX=counter      -> selects via variable
EX ?=
EXAMPLE_RAW := $(firstword $(filter-out run,$(MAKECMDGOALS)))
EXAMPLE := $(or $(EXAMPLE_RAW),$(EX),counter)

# Swallow extra goal (the example name) so make doesn't try to build it
ifeq ($(firstword $(MAKECMDGOALS)),run)
  OTHER_GOALS := $(filter-out run,$(MAKECMDGOALS))
  .PHONY: $(OTHER_GOALS)
  $(OTHER_GOALS):
	@:
endif

# Default target
build:
	@echo "==> Building WASM binary for counter example..."
	GOOS=js GOARCH=wasm go build -o examples/counter/main.wasm examples/counter/main.go

serve:
	@echo "==> Serving http://localhost:8080"
	go run ./server.go

kill:
	lsof -ti:8080 | xargs kill -9 || true

run: kill
	@echo "==> Starting dev server with live reload for example: $(EXAMPLE) ..."
	go run ./spec/dev.go --example $(EXAMPLE)


clean:
	@echo "==> Cleaning up WASM binary..."
	rm -f examples/counter/main.wasm
# Test configuration for js/wasm
PKG ?= ./...
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
	@$(TEST_ENV) go test $(PKG) $(if $(RUN),-run $(RUN),)

# Browser tests for individual examples
test-counter:
	@echo "==> Running browser tests for counter example..."
	go test ./examples/counter -v

test-todo:
	@echo "==> Running browser tests for todo example..."
	go test ./examples/todo -v

test-todo-store:
	@echo "==> Running browser tests for todo_store example..."
	go test ./examples/todo_store -v

test-resource:
	@echo "==> Running browser tests for resource example..."
	go test ./examples/resource -v

test-dom-integration:
	@echo "==> Running browser tests for dom-integration example..."
	go test ./examples/dom_integration -v

# Run all browser tests for examples
test-examples:
	@echo "==> Running browser tests for all examples..."
	@$(MAKE) test-counter
	@$(MAKE) test-todo
	@$(MAKE) test-todo-store
	@$(MAKE) test-resource
	@$(MAKE) test-dom-integration

# Run all tests (unit tests + browser tests)
test-all:
	@echo "==> Running all tests (unit + browser)..."
	@echo "==> Running WASM unit tests (excluding examples)..."
	@$(TEST_ENV) go test ./reactivity ./comps $(if $(RUN),-run $(RUN),)
	@$(MAKE) test-examples
