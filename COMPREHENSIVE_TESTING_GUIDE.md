# Comprehensive Testing Guide for Golid Reactive Systems

## Overview

This guide covers the complete testing infrastructure for the Golid reactive systems, including browser-based testing with wasmbrowsertest, performance regression tests, memory leak detection, and stress testing scenarios.

## Table of Contents

1. [Testing Infrastructure](#testing-infrastructure)
2. [Test Categories](#test-categories)
3. [Running Tests](#running-tests)
4. [Performance Targets](#performance-targets)
5. [CI/CD Integration](#cicd-integration)
6. [Troubleshooting](#troubleshooting)

## Testing Infrastructure

### wasmbrowsertest Integration

The project uses `wasmbrowsertest` to run Go/WASM tests in a real browser environment, enabling testing of:

- DOM manipulation and reactive bindings
- Browser API integration
- Event system with real browser events
- Component lifecycle with actual DOM elements
- Performance measurements in browser context

### Test Files Structure

```
golid/
├── browser_test.go           # Browser-specific integration tests
├── performance_test.go       # Performance regression tests
├── memory_leak_test.go       # Memory leak detection tests
├── stress_test.go           # Stress testing scenarios
├── reactivity_test.go       # Core reactive system tests
├── bind_test.go            # Legacy bind function tests
├── dom_test.go             # DOM utility tests
└── TESTING_README.md       # Basic testing documentation

test_utils/
├── browser_test_utils.go    # Browser testing utilities
└── reactive_test_utils.go   # Reactive system testing utilities

.github/workflows/
└── test.yml                # CI/CD pipeline configuration

test_runner.sh              # Comprehensive test runner script
```

## Test Categories

### 1. Core Reactive System Tests

**File:** `golid/reactivity_test.go`

Tests the fundamental reactive primitives:
- Signal creation, updates, and dependency tracking
- Effect execution and cleanup
- Memo computation and caching
- Owner context and cleanup validation
- Batched updates and scheduling

**Run Command:**
```bash
cd golid && env -i PATH="$PATH" HOME="$HOME" GOOS=js GOARCH=wasm go test -v -run="TestCreateSignal|TestCreateEffect|TestCreateMemo"
```

### 2. Browser Integration Tests

**File:** `golid/browser_test.go`

Tests reactive systems with real DOM integration:
- Signal updates triggering DOM changes
- Effect execution with browser APIs
- Memo computation with DOM bindings
- Event system integration
- Component lifecycle with DOM elements

**Key Tests:**
- `TestSignalWithRealDOM` - Signal-DOM integration
- `TestEffectWithRealDOM` - Effect-DOM integration
- `TestMemoWithRealDOM` - Memo-DOM integration
- `TestReactiveDOMBindings` - Comprehensive DOM bindings
- `TestEventSystemIntegration` - Event handling
- `TestComponentLifecycleWithDOM` - Component lifecycle
- `TestFullReactiveApplication` - Complete application scenario
- `TestBrowserAPIIntegration` - Browser API usage

**Run Command:**
```bash
cd golid && env -i PATH="$PATH" HOME="$HOME" GOOS=js GOARCH=wasm go test -v -run="TestSignalWithRealDOM"
```

### 3. Performance Regression Tests

**File:** `golid/performance_test.go`

Validates performance targets from the implementation roadmap:

| Metric | Target | Test |
|--------|--------|------|
| Signal Update Time | < 5μs | `TestSignalUpdatePerformance` |
| DOM Update Batch | < 10ms | `TestDOMUpdatePerformance` |
| Memory per Signal | < 200B | `TestMemoryPerformance` |
| Max Concurrent Effects | > 10,000 | `TestConcurrencyPerformance` |
| Cascade Depth | < 10 levels | `TestCascadeDepthPerformance` |

**Key Tests:**
- `TestSignalUpdatePerformance` - Signal update latency
- `TestEffectPerformance` - Effect execution performance
- `TestMemoPerformance` - Memo computation performance
- `TestDOMUpdatePerformance` - DOM manipulation performance
- `TestMemoryPerformance` - Memory usage validation
- `TestConcurrencyPerformance` - Concurrent operation handling

**Run Command:**
```bash
cd golid && env -i PATH="$PATH" HOME="$HOME" GOOS=js GOARCH=wasm go test -v -run="TestSignalUpdatePerformance"
```

### 4. Memory Leak Detection Tests

**File:** `golid/memory_leak_test.go`

Detects memory leaks and validates cleanup:
- Signal subscription cleanup
- Effect disposal on owner cleanup
- DOM binding cleanup
- Event subscription cleanup
- Component lifecycle cleanup

**Key Tests:**
- `TestSignalMemoryLeaks` - Signal cleanup validation
- `TestEffectMemoryLeaks` - Effect cleanup validation
- `TestDOMMemoryLeaks` - DOM binding cleanup
- `TestComponentMemoryLeaks` - Component cleanup

**Run Command:**
```bash
cd golid && env -i PATH="$PATH" HOME="$HOME" GOOS=js GOARCH=wasm go test -v -run="TestSignalMemoryLeaks"
```

### 5. Stress Testing Scenarios

**File:** `golid/stress_test.go`

Tests system stability under extreme conditions:
- Thousands of concurrent signals and effects
- Rapid signal updates and DOM manipulations
- Memory pressure scenarios
- Long-running stability tests
- Error recovery testing

**Key Tests:**
- `TestSignalStress` - Massive signal operations
- `TestEffectStress` - Effect cascade stress
- `TestDOMStress` - DOM manipulation stress
- `TestMemoryStress` - Memory pressure testing
- `TestSystemStability` - Long-running stability
- `TestComprehensiveStress` - Full system stress

**Run Command:**
```bash
cd golid && env -i PATH="$PATH" HOME="$HOME" GOOS=js GOARCH=wasm go test -v -run="TestSignalStress"
```

## Running Tests

### Prerequisites

1. **Go 1.21+** installed
2. **wasmbrowsertest** installed:
   ```bash
   go install github.com/agnivade/wasmbrowsertest@latest
   ln -s $(which wasmbrowsertest) $(go env GOPATH)/bin/go_js_wasm_exec
   ```
3. **Chrome browser** for headless testing

### Using the Test Runner Script

The comprehensive test runner provides automated execution of all test categories:

```bash
# Run all tests with default configuration
./test_runner.sh

# Run with verbose output
VERBOSE=true ./test_runner.sh

# Run with performance tests enabled
PERFORMANCE_TESTS=true ./test_runner.sh

# Run with stress tests enabled
STRESS_TESTS=true ./test_runner.sh

# Run comprehensive stress tests
COMPREHENSIVE_STRESS=true STRESS_TESTS=true ./test_runner.sh

# Run with coverage reporting
COVERAGE=true ./test_runner.sh
```

### Manual Test Execution

#### Core Tests (Non-WASM)
```bash
cd golid
go test -tags="!js,!wasm" -v -race ./...
```

#### Browser Tests (WASM)
```bash
cd golid
env -i PATH="$PATH" HOME="$HOME" GOOS=js GOARCH=wasm go test -v -timeout=300s
```

#### Specific Test Categories
```bash
# Signal tests
cd golid && env -i PATH="$PATH" HOME="$HOME" GOOS=js GOARCH=wasm go test -v -run="TestSignal"

# Performance tests
cd golid && env -i PATH="$PATH" HOME="$HOME" GOOS=js GOARCH=wasm go test -v -run="Performance"

# Memory leak tests
cd golid && env -i PATH="$PATH" HOME="$HOME" GOOS=js GOARCH=wasm go test -v -run="MemoryLeak"

# Stress tests
cd golid && env -i PATH="$PATH" HOME="$HOME" GOOS=js GOARCH=wasm go test -v -run="Stress"
```

#### Benchmarks
```bash
cd golid
env -i PATH="$PATH" HOME="$HOME" GOOS=js GOARCH=wasm go test -bench=. -benchmem -timeout=600s
```

## Performance Targets

The testing infrastructure validates these performance targets from the implementation roadmap:

### Signal System Performance
- **Signal Update Time:** < 5μs (10x improvement over baseline)
- **Effect Execution:** < 10μs (efficient reactive updates)
- **Memo Computation:** < 25μs (including caching overhead)

### DOM Performance
- **DOM Update Batch:** < 10ms (10x improvement over baseline)
- **Attribute Updates:** < 5ms per batch
- **Event Handling:** < 1ms latency

### Memory Efficiency
- **Memory per Signal:** < 200B (5x reduction vs virtual DOM)
- **Effect Memory:** < 400B per effect
- **DOM Binding Memory:** < 300B per binding

### Concurrency
- **Max Concurrent Effects:** > 10,000 (100x improvement)
- **Cascade Depth:** < 10 levels (bounded to prevent infinite loops)
- **Concurrent Updates:** Support for 1000+ simultaneous operations

### System Stability
- **Memory Leak Threshold:** < 1MB growth over 1000 operations
- **Error Recovery:** 100% recovery from effect errors
- **Long-running Stability:** 30+ seconds continuous operation

## CI/CD Integration

### GitHub Actions Workflow

The project includes a comprehensive GitHub Actions workflow (`.github/workflows/test.yml`) that runs:

1. **Validation:** Code formatting, go vet, module verification
2. **Core Tests:** Non-WASM reactive system tests with race detection
3. **Browser Tests:** WASM tests with real browser integration
4. **Performance Tests:** Regression testing against performance targets
5. **Memory Leak Tests:** Automated leak detection
6. **Stress Tests:** System stability under load (main branch only)
7. **Comprehensive Suite:** Full test execution with reporting

### Trigger Conditions

- **Push/PR:** Core and browser tests
- **Main Branch:** All tests including stress tests
- **Scheduled (Daily):** Performance regression monitoring
- **Commit Messages:**
  - `[perf]` - Triggers performance tests
  - `[stress]` - Triggers stress tests

### Performance Monitoring

Daily automated performance monitoring tracks:
- Performance trend analysis
- Regression detection
- Memory usage patterns
- System stability metrics

## Troubleshooting

### Common Issues

#### wasmbrowsertest Not Found
```bash
# Install wasmbrowsertest
go install github.com/agnivade/wasmbrowsertest@latest

# Create symlink
ln -s $(which wasmbrowsertest) $(go env GOPATH)/bin/go_js_wasm_exec

# Verify installation
which go_js_wasm_exec
```

#### Browser Tests Failing
```bash
# Check Chrome installation
google-chrome --version

# Run with verbose output
VERBOSE=true ./test_runner.sh

# Check browser console (if running interactively)
# Tests run in headless mode by default
```

#### Memory Leak False Positives
```bash
# Run garbage collection before tests
go clean -cache

# Increase memory thresholds for development
# Edit memory_leak_test.go thresholds if needed
```

#### Performance Test Failures
```bash
# Check system load
top

# Run performance tests in isolation
cd golid && env -i PATH="$PATH" HOME="$HOME" GOOS=js GOARCH=wasm go test -v -run="TestSignalUpdatePerformance"

# Adjust performance targets if hardware differs significantly
```

### Debug Mode

Enable debug output for detailed test execution:

```bash
# Enable verbose logging
export VERBOSE=true

# Enable Go test verbose output
cd golid && env -i PATH="$PATH" HOME="$HOME" GOOS=js GOARCH=wasm go test -v -run="TestName"

# Enable browser console output (for interactive debugging)
# Modify test to remove headless mode temporarily
```

### Test Environment

Ensure consistent test environment:

```bash
# Clean environment
env -i PATH="$PATH" HOME="$HOME" GOOS=js GOARCH=wasm go test

# Minimal PATH to avoid conflicts
export PATH="/usr/bin:/bin:$(go env GOPATH)/bin"

# Clear Go module cache if needed
go clean -modcache
```

## Contributing

When adding new tests:

1. **Follow naming conventions:** `Test[Component][Feature]`
2. **Use test utilities:** Leverage `test_utils/` for common functionality
3. **Add performance validation:** Include performance assertions for new features
4. **Document test purpose:** Clear comments explaining test objectives
5. **Update this guide:** Add new test categories and instructions

### Test Categories Guidelines

- **Unit Tests:** Test individual functions/components in isolation
- **Integration Tests:** Test component interactions and system behavior
- **Browser Tests:** Test DOM integration and browser API usage
- **Performance Tests:** Validate performance targets and detect regressions
- **Memory Tests:** Detect leaks and validate cleanup
- **Stress Tests:** Test system limits and stability

### Performance Test Guidelines

- **Establish baselines:** Measure current performance before optimization
- **Use realistic scenarios:** Test with representative data and usage patterns
- **Account for variance:** Allow reasonable performance variance in CI
- **Document targets:** Clear performance targets with justification
- **Monitor trends:** Track performance over time, not just absolute values

This comprehensive testing infrastructure ensures the Golid reactive system meets its performance targets while maintaining reliability and stability across all supported environments.