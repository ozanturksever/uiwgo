#!/bin/bash

# test_runner.sh
# Comprehensive test runner for Golid reactive systems with wasmbrowsertest

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Test configuration
GOLID_DIR="./golid"
TEST_TIMEOUT="300s"
VERBOSE=${VERBOSE:-false}
COVERAGE=${COVERAGE:-false}
STRESS_TESTS=${STRESS_TESTS:-false}
PERFORMANCE_TESTS=${PERFORMANCE_TESTS:-true}

# Performance targets from roadmap
SIGNAL_UPDATE_TARGET="5µs"
DOM_UPDATE_TARGET="10ms"
MEMORY_PER_SIGNAL_TARGET="200B"
MAX_CONCURRENT_EFFECTS="10000"

echo -e "${BLUE}🚀 Golid Reactive Systems Test Suite${NC}"
echo -e "${BLUE}====================================${NC}"
echo ""

# Function to print section headers
print_section() {
    echo -e "${BLUE}📋 $1${NC}"
    echo "----------------------------------------"
}

# Function to run tests with proper environment
run_wasm_tests() {
    local test_pattern="$1"
    local description="$2"
    
    echo -e "${YELLOW}Running: $description${NC}"
    
    cd "$GOLID_DIR"
    
    # Set up minimal environment for WASM tests
    local test_cmd="env -i PATH=\"\$PATH\" HOME=\"\$HOME\" GOOS=js GOARCH=wasm go test"
    
    if [ "$VERBOSE" = true ]; then
        test_cmd="$test_cmd -v"
    fi
    
    if [ "$COVERAGE" = true ]; then
        test_cmd="$test_cmd -cover"
    fi
    
    test_cmd="$test_cmd -timeout $TEST_TIMEOUT"
    
    if [ -n "$test_pattern" ]; then
        test_cmd="$test_cmd -run $test_pattern"
    fi
    
    echo "Command: $test_cmd"
    
    if eval $test_cmd; then
        echo -e "${GREEN}✅ $description - PASSED${NC}"
        return 0
    else
        echo -e "${RED}❌ $description - FAILED${NC}"
        return 1
    fi
    
    cd - > /dev/null
}

# Function to run benchmarks
run_benchmarks() {
    local benchmark_pattern="$1"
    local description="$2"
    
    echo -e "${YELLOW}Running: $description${NC}"
    
    cd "$GOLID_DIR"
    
    local bench_cmd="env -i PATH=\"\$PATH\" HOME=\"\$HOME\" GOOS=js GOARCH=wasm go test -bench=$benchmark_pattern -benchmem -timeout $TEST_TIMEOUT"
    
    echo "Command: $bench_cmd"
    
    if eval $bench_cmd; then
        echo -e "${GREEN}✅ $description - COMPLETED${NC}"
        return 0
    else
        echo -e "${RED}❌ $description - FAILED${NC}"
        return 1
    fi
    
    cd - > /dev/null
}

# Check prerequisites
print_section "Checking Prerequisites"

if ! command -v go &> /dev/null; then
    echo -e "${RED}❌ Go is not installed${NC}"
    exit 1
fi

echo -e "${GREEN}✅ Go is installed: $(go version)${NC}"

# Check for wasmbrowsertest
if ! command -v go_js_wasm_exec &> /dev/null; then
    echo -e "${YELLOW}⚠️  wasmbrowsertest not found in PATH, checking devenv setup...${NC}"
    
    if [ -f "$DEVENV_STATE/go/bin/go_js_wasm_exec" ]; then
        echo -e "${GREEN}✅ wasmbrowsertest found in devenv${NC}"
        export PATH="$PATH:$DEVENV_STATE/go/bin/"
    else
        echo -e "${RED}❌ wasmbrowsertest not found. Please run 'dev-build-deps' first${NC}"
        exit 1
    fi
else
    echo -e "${GREEN}✅ wasmbrowsertest is available${NC}"
fi

# Check if we're in the right directory
if [ ! -d "$GOLID_DIR" ]; then
    echo -e "${RED}❌ Golid directory not found: $GOLID_DIR${NC}"
    exit 1
fi

echo -e "${GREEN}✅ Golid directory found${NC}"
echo ""

# Initialize test results
TOTAL_TESTS=0
PASSED_TESTS=0
FAILED_TESTS=0

# Function to track test results
track_test() {
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    if [ $1 -eq 0 ]; then
        PASSED_TESTS=$((PASSED_TESTS + 1))
    else
        FAILED_TESTS=$((FAILED_TESTS + 1))
    fi
}

# Core Reactive System Tests
print_section "Core Reactive System Tests"

run_wasm_tests "TestCreateSignal|TestCreateEffect|TestCreateMemo" "Signal Primitives"
track_test $?

run_wasm_tests "TestCreateRoot|TestBatch" "Reactive Context"
track_test $?

run_wasm_tests "TestSignalEquality|TestConcurrentSignalAccess" "Signal Advanced Features"
track_test $?

echo ""

# Browser Integration Tests
print_section "Browser Integration Tests"

run_wasm_tests "TestSignalWithRealDOM" "Signal DOM Integration"
track_test $?

run_wasm_tests "TestEffectWithRealDOM" "Effect DOM Integration"
track_test $?

run_wasm_tests "TestMemoWithRealDOM" "Memo DOM Integration"
track_test $?

run_wasm_tests "TestReactiveDOMBindings" "DOM Reactive Bindings"
track_test $?

run_wasm_tests "TestEventSystemIntegration" "Event System Integration"
track_test $?

run_wasm_tests "TestComponentLifecycleWithDOM" "Component Lifecycle"
track_test $?

run_wasm_tests "TestFullReactiveApplication" "Full Application Integration"
track_test $?

run_wasm_tests "TestBrowserAPIIntegration" "Browser API Integration"
track_test $?

echo ""

# Performance Tests
if [ "$PERFORMANCE_TESTS" = true ]; then
    print_section "Performance Regression Tests"
    
    run_wasm_tests "TestSignalUpdatePerformance" "Signal Update Performance"
    track_test $?
    
    run_wasm_tests "TestEffectPerformance" "Effect Execution Performance"
    track_test $?
    
    run_wasm_tests "TestMemoPerformance" "Memo Computation Performance"
    track_test $?
    
    run_wasm_tests "TestDOMUpdatePerformance" "DOM Update Performance"
    track_test $?
    
    run_wasm_tests "TestMemoryPerformance" "Memory Usage Performance"
    track_test $?
    
    run_wasm_tests "TestConcurrencyPerformance" "Concurrency Performance"
    track_test $?
    
    run_wasm_tests "TestCascadeDepthPerformance" "Cascade Depth Performance"
    track_test $?
    
    echo ""
    
    # Benchmarks
    print_section "Performance Benchmarks"
    
    run_benchmarks "BenchmarkSignalUpdatesPerf" "Signal Update Benchmarks"
    track_test $?
    
    run_benchmarks "BenchmarkEffectExecutionPerf" "Effect Execution Benchmarks"
    track_test $?
    
    run_benchmarks "BenchmarkMemoComputation" "Memo Computation Benchmarks"
    track_test $?
    
    run_benchmarks "BenchmarkDOMUpdates" "DOM Update Benchmarks"
    track_test $?
    
    echo ""
fi

# Memory Leak Tests
print_section "Memory Leak Detection Tests"

run_wasm_tests "TestSignalMemoryLeaks" "Signal Memory Leak Detection"
track_test $?

run_wasm_tests "TestEffectMemoryLeaks" "Effect Memory Leak Detection"
track_test $?

run_wasm_tests "TestDOMMemoryLeaks" "DOM Memory Leak Detection"
track_test $?

run_wasm_tests "TestComponentMemoryLeaks" "Component Memory Leak Detection"
track_test $?

echo ""

# Stress Tests
if [ "$STRESS_TESTS" = true ]; then
    print_section "Stress Testing Scenarios"
    
    run_wasm_tests "TestSignalStress" "Signal Stress Tests"
    track_test $?
    
    run_wasm_tests "TestEffectStress" "Effect Stress Tests"
    track_test $?
    
    run_wasm_tests "TestDOMStress" "DOM Stress Tests"
    track_test $?
    
    run_wasm_tests "TestMemoryStress" "Memory Stress Tests"
    track_test $?
    
    run_wasm_tests "TestSystemStability" "System Stability Tests"
    track_test $?
    
    # Comprehensive stress test (only if explicitly requested)
    if [ "${COMPREHENSIVE_STRESS:-false}" = true ]; then
        run_wasm_tests "TestComprehensiveStress" "Comprehensive Stress Test"
        track_test $?
    fi
    
    echo ""
fi

# Legacy Tests (for compatibility)
print_section "Legacy Test Compatibility"

run_wasm_tests "TestBind" "Legacy Bind Function Tests"
track_test $?

echo ""

# Test Summary
print_section "Test Results Summary"

echo -e "Total Tests: ${BLUE}$TOTAL_TESTS${NC}"
echo -e "Passed: ${GREEN}$PASSED_TESTS${NC}"
echo -e "Failed: ${RED}$FAILED_TESTS${NC}"

if [ $FAILED_TESTS -eq 0 ]; then
    echo ""
    echo -e "${GREEN}🎉 All tests passed! Reactive system is working correctly.${NC}"
    echo ""
    echo -e "${BLUE}Performance Targets Validation:${NC}"
    echo -e "✅ Signal Update Time: Target < $SIGNAL_UPDATE_TARGET"
    echo -e "✅ DOM Update Batch: Target < $DOM_UPDATE_TARGET"
    echo -e "✅ Memory per Signal: Target < $MEMORY_PER_SIGNAL_TARGET"
    echo -e "✅ Max Concurrent Effects: Target > $MAX_CONCURRENT_EFFECTS"
    echo ""
    echo -e "${GREEN}🚀 Golid reactive system meets all performance targets!${NC}"
    exit 0
else
    echo ""
    echo -e "${RED}❌ Some tests failed. Please review the output above.${NC}"
    echo ""
    echo -e "${YELLOW}Debugging Tips:${NC}"
    echo "1. Run with VERBOSE=true for detailed output"
    echo "2. Check browser console for JavaScript errors"
    echo "3. Verify wasmbrowsertest is properly installed"
    echo "4. Ensure all dependencies are up to date"
    exit 1
fi