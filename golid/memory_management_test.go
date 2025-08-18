// memory_management_test.go
// Comprehensive tests for the memory management system

package golid

import (
	"testing"
	"time"
)

// TestMemoryManagerIntegration tests the complete memory management system
func TestMemoryManagerIntegration(t *testing.T) {
	// Create memory manager
	mm := NewMemoryManager(DefaultMemoryConfig())
	defer mm.Dispose()

	// Test allocation tracking
	alloc := mm.TrackAllocation(AllocSignal, 1024, nil, map[string]interface{}{
		"test": "signal_allocation",
	})

	if alloc == nil {
		t.Fatal("Expected allocation to be tracked")
	}

	if alloc.Size != 1024 {
		t.Errorf("Expected allocation size 1024, got %d", alloc.Size)
	}

	if alloc.Type != AllocSignal {
		t.Errorf("Expected allocation type AllocSignal, got %v", alloc.Type)
	}

	// Test allocation retrieval
	retrieved, exists := mm.GetAllocation(alloc.ID)
	if !exists {
		t.Fatal("Expected allocation to exist")
	}

	if retrieved.ID != alloc.ID {
		t.Errorf("Expected allocation ID %d, got %d", alloc.ID, retrieved.ID)
	}

	// Test allocation release
	released := mm.ReleaseAllocation(alloc.ID)
	if !released {
		t.Error("Expected allocation to be released")
	}

	// Test stats
	stats := mm.GetMemoryStats()
	if stats.TotalAllocations == 0 {
		t.Error("Expected total allocations to be greater than 0")
	}
}

// TestCleanupTracker tests the cleanup tracking system
func TestCleanupTracker(t *testing.T) {
	mm := NewMemoryManager(DefaultMemoryConfig())
	defer mm.Dispose()

	tracker := mm.cleanupTracker

	// Test cleanup tracking
	cleaned := false
	operation := tracker.TrackCleanup(
		123,
		"test_resource",
		func() error {
			cleaned = true
			return nil
		},
		func() bool {
			return cleaned
		},
		nil,
		map[string]interface{}{
			"test": "cleanup_operation",
		},
	)

	if operation == nil {
		t.Fatal("Expected cleanup operation to be tracked")
	}

	// Test cleanup execution
	err := tracker.ExecuteCleanup(operation)
	if err != nil {
		t.Errorf("Expected cleanup to succeed, got error: %v", err)
	}

	if !cleaned {
		t.Error("Expected cleanup function to be called")
	}

	// Test cleanup stats
	stats := tracker.GetStats()
	if stats.TotalOperations == 0 {
		t.Error("Expected total operations to be greater than 0")
	}
}

// TestResourceMonitor tests the resource monitoring system
func TestResourceMonitor(t *testing.T) {
	mm := NewMemoryManager(DefaultMemoryConfig())
	defer mm.Dispose()

	monitor := mm.resourceMonitor

	// Create some allocations
	for i := 0; i < 5; i++ {
		mm.TrackAllocation(AllocSignal, int64(100*(i+1)), nil, nil)
	}

	// Update metrics
	monitor.UpdateMetrics()

	// Test metrics
	metrics := monitor.GetCurrentMetrics()
	if metrics.SignalCount == 0 {
		t.Error("Expected signal count to be greater than 0")
	}

	// Test report
	report := monitor.GetReport()
	if report == nil {
		t.Error("Expected report to be generated")
	}

	if report["monitor_id"] == nil {
		t.Error("Expected monitor ID in report")
	}
}

// TestLeakDetector tests the advanced leak detection system
func TestLeakDetector(t *testing.T) {
	mm := NewMemoryManager(DefaultMemoryConfig())
	defer mm.Dispose()

	detector := mm.leakDetector

	// Create allocations that might be considered leaks
	for i := 0; i < 150; i++ { // Exceed default threshold
		alloc := mm.TrackAllocation(AllocSignal, 1024, nil, nil)
		// Make allocation look old
		alloc.Timestamp = time.Now().Add(-10 * time.Minute)
	}

	// Perform leak check
	detector.PerformLeakCheck()

	// Check for violations
	violations := detector.GetViolations()
	if len(violations) == 0 {
		t.Error("Expected leak violations to be detected")
	}

	// Check suspicious allocations
	suspicious := detector.GetSuspiciousAllocations()
	if len(suspicious) == 0 {
		t.Error("Expected suspicious allocations to be detected")
	}
}

// TestResourceRegistry tests the resource registry system
func TestResourceRegistry(t *testing.T) {
	mm := NewMemoryManager(DefaultMemoryConfig())
	defer mm.Dispose()

	registry := mm.resourceRegistry

	// Create and register allocations
	alloc1 := mm.TrackAllocation(AllocSignal, 1024, nil, nil)
	alloc2 := mm.TrackAllocation(AllocEffect, 2048, nil, nil)

	// Test retrieval by type
	signals := registry.GetAllocationsByType(AllocSignal)
	if len(signals) == 0 {
		t.Error("Expected signal allocations to be found")
	}

	effects := registry.GetAllocationsByType(AllocEffect)
	if len(effects) == 0 {
		t.Error("Expected effect allocations to be found")
	}

	// Test dependency tracking
	registry.AddDependency(alloc1.ID, alloc2.ID)
	deps := registry.GetDependencies(alloc1.ID)
	if len(deps) == 0 {
		t.Error("Expected dependencies to be tracked")
	}

	if deps[0] != alloc2.ID {
		t.Errorf("Expected dependency %d, got %d", alloc2.ID, deps[0])
	}
}

// TestCleanupScheduler tests the cleanup scheduling system
func TestCleanupScheduler(t *testing.T) {
	mm := NewMemoryManager(DefaultMemoryConfig())
	defer mm.Dispose()

	scheduler := mm.cleanupScheduler
	scheduler.Start()
	defer scheduler.Stop()

	// Test immediate scheduling
	executed := false
	cleanup := scheduler.ScheduleImmediate(
		123,
		"test_resource",
		func() error {
			executed = true
			return nil
		},
		func() bool {
			return executed
		},
		PriorityNormal,
		nil,
		nil,
	)

	if cleanup == nil {
		t.Fatal("Expected cleanup to be scheduled")
	}

	// Wait for execution
	time.Sleep(100 * time.Millisecond)

	// Check stats
	stats := scheduler.GetStats()
	if stats.TotalScheduled == 0 {
		t.Error("Expected scheduled operations to be greater than 0")
	}
}

// TestMemoryProfiler tests the memory profiling system
func TestMemoryProfiler(t *testing.T) {
	mm := NewMemoryManager(DefaultMemoryConfig())
	defer mm.Dispose()

	profiler := NewMemoryProfiler(mm)

	// Test snapshot
	snapshot := profiler.TakeSnapshot()
	if snapshot == nil {
		t.Fatal("Expected snapshot to be created")
	}

	if snapshot.ID == 0 {
		t.Error("Expected snapshot to have valid ID")
	}

	// Test profiling
	profile := profiler.StartProfiling("test_profile")
	if profile == nil {
		t.Fatal("Expected profiling to start")
	}

	// Create some allocations during profiling
	for i := 0; i < 10; i++ {
		mm.TrackAllocation(AllocSignal, int64(100*(i+1)), nil, nil)
	}

	// Stop profiling
	completedProfile := profiler.StopProfiling()
	if completedProfile == nil {
		t.Fatal("Expected profiling to complete")
	}

	if completedProfile.Summary.TotalAllocations == 0 {
		t.Error("Expected allocations to be captured in profile")
	}

	// Test report generation
	report := profiler.GenerateReport()
	if report == nil {
		t.Error("Expected report to be generated")
	}

	if report["profiler_info"] == nil {
		t.Error("Expected profiler info in report")
	}
}

// TestMemoryLeakDetection tests comprehensive leak detection
func TestMemoryLeakDetection(t *testing.T) {
	mm := NewMemoryManager(DefaultMemoryConfig())
	defer mm.Dispose()

	// Create a scenario that should trigger leak detection
	var allocations []*Allocation

	// Create many allocations
	for i := 0; i < 200; i++ {
		alloc := mm.TrackAllocation(AllocSignal, 1024, nil, map[string]interface{}{
			"index": i,
		})
		allocations = append(allocations, alloc)

		// Make some allocations look old and suspicious
		if i%10 == 0 {
			alloc.Timestamp = time.Now().Add(-15 * time.Minute)
		}
	}

	// Trigger leak detection
	mm.leakDetector.PerformLeakCheck()

	// Check results
	violations := mm.leakDetector.GetViolations()
	if len(violations) == 0 {
		t.Error("Expected leak violations to be detected with 200 allocations")
	}

	suspicious := mm.leakDetector.GetSuspiciousAllocations()
	if len(suspicious) == 0 {
		t.Error("Expected suspicious allocations to be detected")
	}

	// Test cleanup of some allocations
	for i := 0; i < 100; i++ {
		mm.ReleaseAllocation(allocations[i].ID)
	}

	// Check that stats are updated
	stats := mm.GetMemoryStats()
	if stats.ActiveAllocations >= 200 {
		t.Error("Expected active allocations to decrease after cleanup")
	}
}

// TestIntegratedCleanupFlow tests the complete cleanup flow
func TestIntegratedCleanupFlow(t *testing.T) {
	mm := NewMemoryManager(DefaultMemoryConfig())
	defer mm.Dispose()

	// Start scheduler
	mm.cleanupScheduler.Start()
	defer mm.cleanupScheduler.Stop()

	// Create allocations with cleanup
	var cleanedCount int
	var allocations []*Allocation

	for i := 0; i < 10; i++ {
		alloc := mm.TrackAllocation(AllocSignal, 1024, nil, nil)
		allocations = append(allocations, alloc)

		// Schedule cleanup
		mm.cleanupScheduler.ScheduleDelayed(
			alloc.ID,
			"test_signal",
			func() error {
				cleanedCount++
				mm.ReleaseAllocation(alloc.ID)
				return nil
			},
			func() bool {
				_, exists := mm.GetAllocation(alloc.ID)
				return !exists
			},
			50*time.Millisecond,
			PriorityNormal,
			nil,
			nil,
		)
	}

	// Wait for cleanups to execute
	time.Sleep(200 * time.Millisecond)

	// Check that cleanups were executed
	if cleanedCount == 0 {
		t.Error("Expected some cleanups to be executed")
	}

	// Check scheduler stats
	stats := mm.cleanupScheduler.GetStats()
	if stats.TotalScheduled == 0 {
		t.Error("Expected scheduled operations")
	}
}

// BenchmarkMemoryManager benchmarks memory manager performance
func BenchmarkMemoryManager(b *testing.B) {
	mm := NewMemoryManager(DefaultMemoryConfig())
	defer mm.Dispose()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		alloc := mm.TrackAllocation(AllocSignal, 1024, nil, nil)
		mm.ReleaseAllocation(alloc.ID)
	}
}

// BenchmarkLeakDetection benchmarks leak detection performance
func BenchmarkLeakDetection(b *testing.B) {
	mm := NewMemoryManager(DefaultMemoryConfig())
	defer mm.Dispose()

	// Create some allocations
	for i := 0; i < 1000; i++ {
		mm.TrackAllocation(AllocSignal, 1024, nil, nil)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		mm.leakDetector.PerformLeakCheck()
	}
}

// TestMemoryManagerReport tests comprehensive reporting
func TestMemoryManagerReport(t *testing.T) {
	mm := NewMemoryManager(DefaultMemoryConfig())
	defer mm.Dispose()

	// Create some activity
	for i := 0; i < 50; i++ {
		alloc := mm.TrackAllocation(AllocSignal, int64(100*(i+1)), nil, nil)
		if i%2 == 0 {
			mm.ReleaseAllocation(alloc.ID)
		}
	}

	// Generate comprehensive report
	report := mm.GetDetailedReport()

	// Validate report structure
	if report["manager_id"] == nil {
		t.Error("Expected manager ID in report")
	}

	if report["stats"] == nil {
		t.Error("Expected stats in report")
	}

	if report["allocations_by_type"] == nil {
		t.Error("Expected allocations by type in report")
	}

	if report["leak_detector"] == nil {
		t.Error("Expected leak detector report")
	}

	if report["resource_monitor"] == nil {
		t.Error("Expected resource monitor report")
	}

	if report["cleanup_tracker"] == nil {
		t.Error("Expected cleanup tracker report")
	}
}
