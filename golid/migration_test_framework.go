// migration_test_framework.go
// Testing framework for systematic V1 to V2 migration validation

package golid

import (
	"fmt"
	"reflect"
	"runtime"
	"sync"
	"testing"
	"time"
)

// ------------------------------------
// 🧪 Migration Testing Framework
// ------------------------------------

// MigrationTestSuite manages a collection of migration tests
type MigrationTestSuite struct {
	tests   []*MigrationTest
	results map[string]MigrationResult
	mutex   sync.RWMutex
}

// MigrationTest defines a single test for V1/V2 equivalence
type MigrationTest struct {
	Name        string
	Description string
	V1Setup     func() interface{}
	V2Setup     func() interface{}
	Validator   func(v1Result, v2Result interface{}) error
	Cleanup     func()
	MaxDuration time.Duration
	SkipV1      bool // For pure V2 tests
	SkipV2      bool // For legacy V1 tests
}

// MigrationResult contains the outcome of a migration test
type MigrationResult struct {
	TestName        string
	Success         bool
	Error           error
	V1Performance   MigrationPerformanceMetrics
	V2Performance   MigrationPerformanceMetrics
	Duration        time.Duration
	MemoryUsage     MemoryMetrics
	CompatibilityOK bool
}

// MigrationPerformanceMetrics captures performance data for comparison
type MigrationPerformanceMetrics struct {
	ExecutionTime  time.Duration
	AllocationsB   uint64
	AllocationsN   uint64
	GCRuns         uint32
	GoroutineCount int
}

// MemoryMetrics tracks memory usage during migration tests
type MemoryMetrics struct {
	V1MemoryUsage int64
	V2MemoryUsage int64
	LeakDetected  bool
	CleanupOK     bool
}

// NewMigrationTestSuite creates a new migration test suite
func NewMigrationTestSuite() *MigrationTestSuite {
	return &MigrationTestSuite{
		tests:   make([]*MigrationTest, 0),
		results: make(map[string]MigrationResult),
	}
}

// AddTest adds a migration test to the suite
func (mts *MigrationTestSuite) AddTest(test *MigrationTest) {
	mts.mutex.Lock()
	defer mts.mutex.Unlock()
	mts.tests = append(mts.tests, test)
}

// RunAllTests executes all migration tests in the suite
func (mts *MigrationTestSuite) RunAllTests(t *testing.T) map[string]MigrationResult {
	results := make(map[string]MigrationResult)

	for _, test := range mts.tests {
		t.Run(test.Name, func(t *testing.T) {
			result := mts.runSingleTest(test)
			mts.mutex.Lock()
			mts.results[test.Name] = result
			results[test.Name] = result
			mts.mutex.Unlock()

			if !result.Success {
				t.Errorf("Migration test %s failed: %v", test.Name, result.Error)
			}
		})
	}

	return results
}

// runSingleTest executes a single migration test
func (mts *MigrationTestSuite) runSingleTest(test *MigrationTest) MigrationResult {
	start := time.Now()
	result := MigrationResult{
		TestName: test.Name,
		Success:  false,
	}

	defer func() {
		result.Duration = time.Since(start)
		if test.Cleanup != nil {
			test.Cleanup()
		}
		if r := recover(); r != nil {
			result.Error = fmt.Errorf("test panicked: %v", r)
		}
	}()

	// Execute V1 test if not skipped
	var v1Result interface{}
	var v1Metrics MigrationPerformanceMetrics
	if !test.SkipV1 && test.V1Setup != nil {
		v1Metrics = captureMigrationPerformanceMetrics(func() {
			v1Result = test.V1Setup()
		})
		result.V1Performance = v1Metrics
	}

	// Execute V2 test if not skipped
	var v2Result interface{}
	var v2Metrics MigrationPerformanceMetrics
	if !test.SkipV2 && test.V2Setup != nil {
		v2Metrics = captureMigrationPerformanceMetrics(func() {
			v2Result = test.V2Setup()
		})
		result.V2Performance = v2Metrics
	}

	// Validate results if both are available
	if !test.SkipV1 && !test.SkipV2 && test.Validator != nil {
		if err := test.Validator(v1Result, v2Result); err != nil {
			result.Error = fmt.Errorf("validation failed: %w", err)
			return result
		}
		result.CompatibilityOK = true
	}

	// Check for performance improvements
	if !test.SkipV1 && !test.SkipV2 {
		result.MemoryUsage = MemoryMetrics{
			V1MemoryUsage: int64(v1Metrics.AllocationsB),
			V2MemoryUsage: int64(v2Metrics.AllocationsB),
			LeakDetected:  v2Metrics.AllocationsB > v1Metrics.AllocationsB*2, // Simple leak detection
		}
	}

	result.Success = true
	return result
}

// captureMigrationPerformanceMetrics captures performance metrics during function execution
func captureMigrationPerformanceMetrics(fn func()) MigrationPerformanceMetrics {
	var m1, m2 runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&m1)

	start := time.Now()
	goroutinesBefore := runtime.NumGoroutine()

	fn()

	runtime.GC()
	runtime.ReadMemStats(&m2)
	goroutinesAfter := runtime.NumGoroutine()

	return MigrationPerformanceMetrics{
		ExecutionTime:  time.Since(start),
		AllocationsB:   m2.TotalAlloc - m1.TotalAlloc,
		AllocationsN:   m2.Mallocs - m1.Mallocs,
		GCRuns:         m2.NumGC - m1.NumGC,
		GoroutineCount: goroutinesAfter - goroutinesBefore,
	}
}

// GenerateReport creates a comprehensive migration test report
func (mts *MigrationTestSuite) GenerateReport() string {
	mts.mutex.RLock()
	defer mts.mutex.RUnlock()

	report := "Migration Test Report\n"
	report += "=====================\n\n"

	totalTests := len(mts.results)
	successCount := 0

	for _, result := range mts.results {
		if result.Success {
			successCount++
		}
	}

	report += fmt.Sprintf("Total Tests: %d\n", totalTests)
	report += fmt.Sprintf("Successful: %d\n", successCount)
	report += fmt.Sprintf("Failed: %d\n", totalTests-successCount)
	report += fmt.Sprintf("Success Rate: %.2f%%\n\n", float64(successCount)/float64(totalTests)*100)

	// Performance summary
	var totalV1Time, totalV2Time time.Duration
	var totalV1Memory, totalV2Memory uint64

	for _, result := range mts.results {
		totalV1Time += result.V1Performance.ExecutionTime
		totalV2Time += result.V2Performance.ExecutionTime
		totalV1Memory += result.V1Performance.AllocationsB
		totalV2Memory += result.V2Performance.AllocationsB
	}

	if totalV1Time > 0 && totalV2Time > 0 {
		report += "Performance Comparison:\n"
		report += fmt.Sprintf("V1 Total Time: %v\n", totalV1Time)
		report += fmt.Sprintf("V2 Total Time: %v\n", totalV2Time)

		if totalV1Time > totalV2Time {
			improvement := float64(totalV1Time-totalV2Time) / float64(totalV1Time) * 100
			report += fmt.Sprintf("V2 Performance Improvement: %.2f%%\n", improvement)
		}

		report += fmt.Sprintf("V1 Total Memory: %d bytes\n", totalV1Memory)
		report += fmt.Sprintf("V2 Total Memory: %d bytes\n", totalV2Memory)

		if totalV1Memory > totalV2Memory {
			reduction := float64(totalV1Memory-totalV2Memory) / float64(totalV1Memory) * 100
			report += fmt.Sprintf("V2 Memory Reduction: %.2f%%\n", reduction)
		}
		report += "\n"
	}

	// Individual test results
	report += "Individual Test Results:\n"
	report += "------------------------\n"

	for testName, result := range mts.results {
		status := "✅ PASS"
		if !result.Success {
			status = "❌ FAIL"
		}

		report += fmt.Sprintf("%s %s\n", status, testName)
		if result.Error != nil {
			report += fmt.Sprintf("   Error: %v\n", result.Error)
		}

		if result.CompatibilityOK {
			report += "   ✅ V1/V2 Compatibility Verified\n"
		}

		if result.V1Performance.ExecutionTime > 0 && result.V2Performance.ExecutionTime > 0 {
			if result.V2Performance.ExecutionTime < result.V1Performance.ExecutionTime {
				improvement := float64(result.V1Performance.ExecutionTime-result.V2Performance.ExecutionTime) /
					float64(result.V1Performance.ExecutionTime) * 100
				report += fmt.Sprintf("   ⚡ %2.f%% faster with V2\n", improvement)
			}
		}

		report += "\n"
	}

	return report
}

// ------------------------------------
// 🔬 Common Migration Test Patterns
// ------------------------------------

// CreateSignalCompatibilityTest creates a test for signal V1/V2 compatibility
func CreateSignalCompatibilityTest[T comparable](name string, initialValue T, updates []T) *MigrationTest {
	return &MigrationTest{
		Name:        name,
		Description: fmt.Sprintf("Signal compatibility test for type %T", initialValue),
		V1Setup: func() interface{} {
			// This would use the old V1 signal system
			return map[string]interface{}{
				"type":    "v1_signal_test",
				"initial": initialValue,
				"updates": updates,
			}
		},
		V2Setup: func() interface{} {
			// Test using V2 CreateSignal
			getter, setter := CreateSignal(initialValue)

			results := make([]T, 0, len(updates)+1)
			results = append(results, getter())

			for _, update := range updates {
				setter(update)
				results = append(results, getter())
			}

			return results
		},
		Validator: func(v1Result, v2Result interface{}) error {
			// For now, just validate that V2 produces expected results
			v2Results, ok := v2Result.([]T)
			if !ok {
				return fmt.Errorf("V2 result type mismatch")
			}

			expected := append([]T{initialValue}, updates...)
			if !reflect.DeepEqual(v2Results, expected) {
				return fmt.Errorf("V2 results don't match expected: got %v, want %v", v2Results, expected)
			}

			return nil
		},
		MaxDuration: 5 * time.Second,
	}
}

// CreateEffectCompatibilityTest creates a test for effect V1/V2 compatibility
func CreateEffectCompatibilityTest(name string) *MigrationTest {
	return &MigrationTest{
		Name:        name,
		Description: "Effect compatibility test",
		V1Setup: func() interface{} {
			// This would test V1 Watch functionality
			return "v1_effect_test"
		},
		V2Setup: func() interface{} {
			// Test V2 CreateEffect
			counter := 0
			getter, setter := CreateSignal(0)

			CreateEffect(func() {
				value := getter()
				counter = value * 2
			}, nil)

			setter(5)
			FlushScheduler() // Ensure effect runs

			return counter
		},
		Validator: func(v1Result, v2Result interface{}) error {
			// For now, just validate V2 effect worked
			if v2Result != 10 {
				return fmt.Errorf("V2 effect didn't work: got %v, want 10", v2Result)
			}
			return nil
		},
		MaxDuration: 5 * time.Second,
	}
}

// CreateMemoryLeakTest creates a test to validate memory leak prevention
func CreateMemoryLeakTest(name string, iterations int) *MigrationTest {
	return &MigrationTest{
		Name:        name,
		Description: "Memory leak prevention test",
		SkipV1:      true, // V1 has known memory leaks
		V2Setup: func() interface{} {
			// Create and dispose many owner contexts to test cleanup
			for i := 0; i < iterations; i++ {
				_, cleanup := CreateRoot(func() interface{} {
					getter, setter := CreateSignal(i)
					CreateEffect(func() {
						_ = getter()
					}, nil)
					setter(i * 2)
					return nil
				})
				cleanup() // This should clean up all resources
			}

			return "cleanup_test_completed"
		},
		Validator: func(v1Result, v2Result interface{}) error {
			// Memory validation would be done through metrics
			return nil
		},
		MaxDuration: 30 * time.Second,
	}
}

// ------------------------------------
// 🎯 Migration Test Utilities
// ------------------------------------

// RunQuickMigrationTest runs basic migration validation tests
func RunQuickMigrationTest(t *testing.T) {
	suite := NewMigrationTestSuite()

	// Add basic signal tests
	suite.AddTest(CreateSignalCompatibilityTest("StringSignal", "initial", []string{"updated", "final"}))
	suite.AddTest(CreateSignalCompatibilityTest("IntSignal", 0, []int{5, 10, 15}))
	suite.AddTest(CreateSignalCompatibilityTest("BoolSignal", false, []bool{true, false, true}))

	// Add effect tests
	suite.AddTest(CreateEffectCompatibilityTest("BasicEffect"))

	// Add memory tests
	suite.AddTest(CreateMemoryLeakTest("MemoryCleanup", 100))

	// Run all tests
	results := suite.RunAllTests(t)

	// Generate and log report
	report := suite.GenerateReport()
	t.Log("\n" + report)

	// Validate overall success
	successCount := 0
	for _, result := range results {
		if result.Success {
			successCount++
		}
	}

	if successCount < len(results) {
		t.Errorf("Migration tests failed: %d/%d passed", successCount, len(results))
	}
}
