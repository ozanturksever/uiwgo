.PHONY: build test serve run clean

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

test:
	@echo "==> Running tests..."
	go test $(shell go list ./... | grep -v '/examples/' | grep -v '/pkg/uiwgo')

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