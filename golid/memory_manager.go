// memory_manager.go
// Central memory management and leak detection system for Golid framework

package golid

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

// ------------------------------------
// 🧠 Memory Manager Core
// ------------------------------------

var (
	globalMemoryManager    *MemoryManager
	memoryManagerOnce      sync.Once
	memoryManagerIdCounter uint64
	allocationIdCounter    uint64
)

// MemoryManager provides centralized memory management and leak detection
type MemoryManager struct {
	id               uint64
	allocations      map[uint64]*Allocation
	resourceRegistry *ResourceRegistry
	cleanupTracker   *CleanupTracker
	leakDetector     *LeakDetector
	resourceMonitor  *ResourceMonitor
	cleanupScheduler *CleanupScheduler

	// Configuration
	config MemoryConfig

	// State tracking
	enabled    bool
	monitoring bool
	debugMode  bool

	// Statistics
	stats MemoryStats

	// Synchronization
	mutex sync.RWMutex

	// Background operations
	ctx              context.Context
	cancel           context.CancelFunc
	backgroundTicker *time.Ticker
	stopBackground   chan bool
}

// MemoryConfig provides configuration for memory management
type MemoryConfig struct {
	// Leak detection settings
	LeakThreshold      int           `json:"leak_threshold"`
	LeakCheckInterval  time.Duration `json:"leak_check_interval"`
	LeakAlertThreshold int           `json:"leak_alert_threshold"`

	// Cleanup settings
	CleanupBatchSize     int           `json:"cleanup_batch_size"`
	CleanupTimeout       time.Duration `json:"cleanup_timeout"`
	CleanupRetryAttempts int           `json:"cleanup_retry_attempts"`

	// Monitoring settings
	MonitoringInterval time.Duration `json:"monitoring_interval"`
	HistoryRetention   time.Duration `json:"history_retention"`
	MetricsBufferSize  int           `json:"metrics_buffer_size"`

	// Performance settings
	MaxConcurrentCleanup int   `json:"max_concurrent_cleanup"`
	GCTriggerThreshold   int64 `json:"gc_trigger_threshold"`

	// Development settings
	EnableStackTraces     bool `json:"enable_stack_traces"`
	EnableDetailedLogging bool `json:"enable_detailed_logging"`
	EnableProfiling       bool `json:"enable_profiling"`
}

// MemoryStats tracks memory usage statistics
type MemoryStats struct {
	// Allocation tracking
	TotalAllocations     uint64 `json:"total_allocations"`
	ActiveAllocations    uint64 `json:"active_allocations"`
	PeakAllocations      uint64 `json:"peak_allocations"`
	TotalMemoryAllocated int64  `json:"total_memory_allocated"`
	ActiveMemoryUsage    int64  `json:"active_memory_usage"`
	PeakMemoryUsage      int64  `json:"peak_memory_usage"`

	// Leak detection
	LeaksDetected  uint64 `json:"leaks_detected"`
	LeakViolations uint64 `json:"leak_violations"`
	FalsePositives uint64 `json:"false_positives"`

	// Cleanup tracking
	CleanupOperations  uint64 `json:"cleanup_operations"`
	SuccessfulCleanups uint64 `json:"successful_cleanups"`
	FailedCleanups     uint64 `json:"failed_cleanups"`
	CleanupRetries     uint64 `json:"cleanup_retries"`

	// Performance metrics
	AverageCleanupTime time.Duration `json:"average_cleanup_time"`
	MaxCleanupTime     time.Duration `json:"max_cleanup_time"`
	GCTriggers         uint64        `json:"gc_triggers"`

	// Timestamps
	StartTime  time.Time `json:"start_time"`
	LastUpdate time.Time `json:"last_update"`

	mutex sync.RWMutex
}

// Allocation represents a tracked memory allocation
type Allocation struct {
	ID           uint64                 `json:"id"`
	Type         AllocationType         `json:"type"`
	Size         int64                  `json:"size"`
	Timestamp    time.Time              `json:"timestamp"`
	Owner        *Owner                 `json:"-"`
	ResourceID   uint64                 `json:"resource_id"`
	StackTrace   []uintptr              `json:"-"`
	Metadata     map[string]interface{} `json:"metadata"`
	CleanupFunc  func()                 `json:"-"`
	Disposed     bool                   `json:"disposed"`
	LastAccessed time.Time              `json:"last_accessed"`
	AccessCount  uint64                 `json:"access_count"`
	mutex        sync.RWMutex
}

// AllocationType defines types of memory allocations
type AllocationType int

const (
	AllocSignal AllocationType = iota
	AllocEffect
	AllocMemo
	AllocResource
	AllocDOMBinding
	AllocEventSubscription
	AllocComponent
	AllocCustom
)

// String returns string representation of allocation type
func (a AllocationType) String() string {
	switch a {
	case AllocSignal:
		return "Signal"
	case AllocEffect:
		return "Effect"
	case AllocMemo:
		return "Memo"
	case AllocResource:
		return "Resource"
	case AllocDOMBinding:
		return "DOMBinding"
	case AllocEventSubscription:
		return "EventSubscription"
	case AllocComponent:
		return "Component"
	case AllocCustom:
		return "Custom"
	default:
		return "Unknown"
	}
}

// ------------------------------------
// 🏗️ Memory Manager Creation
// ------------------------------------

// GetGlobalMemoryManager returns the global memory manager instance
func GetGlobalMemoryManager() *MemoryManager {
	memoryManagerOnce.Do(func() {
		globalMemoryManager = NewMemoryManager(DefaultMemoryConfig())
	})
	return globalMemoryManager
}

// NewMemoryManager creates a new memory manager instance
func NewMemoryManager(config MemoryConfig) *MemoryManager {
	ctx, cancel := context.WithCancel(context.Background())

	mm := &MemoryManager{
		id:               atomic.AddUint64(&memoryManagerIdCounter, 1),
		allocations:      make(map[uint64]*Allocation),
		config:           config,
		enabled:          true,
		monitoring:       true,
		debugMode:        config.EnableDetailedLogging,
		ctx:              ctx,
		cancel:           cancel,
		backgroundTicker: time.NewTicker(config.MonitoringInterval),
		stopBackground:   make(chan bool, 1),
		stats: MemoryStats{
			StartTime:  time.Now(),
			LastUpdate: time.Now(),
		},
	}

	// Initialize sub-components
	mm.resourceRegistry = NewResourceRegistry(mm)
	mm.cleanupTracker = NewCleanupTracker(mm)
	mm.leakDetector = NewLeakDetector(mm)
	mm.resourceMonitor = NewResourceMonitor(mm)
	mm.cleanupScheduler = NewCleanupScheduler(mm)

	// Start background monitoring
	go mm.backgroundMonitoring()

	return mm
}

// DefaultMemoryConfig returns default memory management configuration
func DefaultMemoryConfig() MemoryConfig {
	return MemoryConfig{
		LeakThreshold:         100,
		LeakCheckInterval:     30 * time.Second,
		LeakAlertThreshold:    500,
		CleanupBatchSize:      50,
		CleanupTimeout:        5 * time.Second,
		CleanupRetryAttempts:  3,
		MonitoringInterval:    10 * time.Second,
		HistoryRetention:      24 * time.Hour,
		MetricsBufferSize:     1000,
		MaxConcurrentCleanup:  10,
		GCTriggerThreshold:    100 * 1024 * 1024, // 100MB
		EnableStackTraces:     false,
		EnableDetailedLogging: false,
		EnableProfiling:       false,
	}
}

// ------------------------------------
// 🔍 Allocation Tracking
// ------------------------------------

// TrackAllocation registers a new memory allocation
func (mm *MemoryManager) TrackAllocation(allocType AllocationType, size int64, owner *Owner, metadata map[string]interface{}) *Allocation {
	if !mm.enabled {
		return nil
	}

	alloc := &Allocation{
		ID:           atomic.AddUint64(&allocationIdCounter, 1),
		Type:         allocType,
		Size:         size,
		Timestamp:    time.Now(),
		Owner:        owner,
		Metadata:     metadata,
		LastAccessed: time.Now(),
		AccessCount:  1,
	}

	// Capture stack trace if enabled
	if mm.config.EnableStackTraces {
		alloc.StackTrace = make([]uintptr, 32)
		n := runtime.Callers(2, alloc.StackTrace)
		alloc.StackTrace = alloc.StackTrace[:n]
	}

	mm.mutex.Lock()
	mm.allocations[alloc.ID] = alloc
	mm.mutex.Unlock()

	// Update statistics
	mm.updateStats(func(stats *MemoryStats) {
		atomic.AddUint64(&stats.TotalAllocations, 1)
		atomic.AddUint64(&stats.ActiveAllocations, 1)
		atomic.AddInt64(&stats.TotalMemoryAllocated, size)
		atomic.AddInt64(&stats.ActiveMemoryUsage, size)

		// Update peaks
		if stats.ActiveAllocations > stats.PeakAllocations {
			stats.PeakAllocations = stats.ActiveAllocations
		}
		if stats.ActiveMemoryUsage > stats.PeakMemoryUsage {
			stats.PeakMemoryUsage = stats.ActiveMemoryUsage
		}
	})

	// Register with resource registry
	if mm.resourceRegistry != nil {
		mm.resourceRegistry.RegisterAllocation(alloc)
	}

	// Check for potential leaks
	if mm.leakDetector != nil {
		go mm.leakDetector.CheckAllocation(alloc)
	}

	return alloc
}

// ReleaseAllocation marks an allocation as released
func (mm *MemoryManager) ReleaseAllocation(id uint64) bool {
	mm.mutex.Lock()
	alloc, exists := mm.allocations[id]
	if !exists {
		mm.mutex.Unlock()
		return false
	}

	if alloc.Disposed {
		mm.mutex.Unlock()
		return false
	}

	alloc.mutex.Lock()
	alloc.Disposed = true
	size := alloc.Size
	alloc.mutex.Unlock()

	delete(mm.allocations, id)
	mm.mutex.Unlock()

	// Update statistics
	mm.updateStats(func(stats *MemoryStats) {
		atomic.AddUint64(&stats.ActiveAllocations, ^uint64(0)) // Decrement
		atomic.AddInt64(&stats.ActiveMemoryUsage, -size)
	})

	// Notify cleanup tracker
	if mm.cleanupTracker != nil {
		mm.cleanupTracker.RecordCleanup(id, true, nil)
	}

	return true
}

// GetAllocation retrieves an allocation by ID
func (mm *MemoryManager) GetAllocation(id uint64) (*Allocation, bool) {
	mm.mutex.RLock()
	defer mm.mutex.RUnlock()

	alloc, exists := mm.allocations[id]
	if exists && !alloc.Disposed {
		alloc.mutex.Lock()
		alloc.LastAccessed = time.Now()
		atomic.AddUint64(&alloc.AccessCount, 1)
		alloc.mutex.Unlock()
	}

	return alloc, exists
}

// GetAllocationsByType returns all allocations of a specific type
func (mm *MemoryManager) GetAllocationsByType(allocType AllocationType) []*Allocation {
	mm.mutex.RLock()
	defer mm.mutex.RUnlock()

	var result []*Allocation
	for _, alloc := range mm.allocations {
		if alloc.Type == allocType && !alloc.Disposed {
			result = append(result, alloc)
		}
	}

	return result
}

// GetAllocationsByOwner returns all allocations owned by a specific owner
func (mm *MemoryManager) GetAllocationsByOwner(owner *Owner) []*Allocation {
	mm.mutex.RLock()
	defer mm.mutex.RUnlock()

	var result []*Allocation
	for _, alloc := range mm.allocations {
		if alloc.Owner == owner && !alloc.Disposed {
			result = append(result, alloc)
		}
	}

	return result
}

// ------------------------------------
// 🔧 Memory Management Operations
// ------------------------------------

// ForceGC triggers garbage collection if memory usage exceeds threshold
func (mm *MemoryManager) ForceGC() {
	if mm.stats.ActiveMemoryUsage > mm.config.GCTriggerThreshold {
		runtime.GC()
		mm.updateStats(func(stats *MemoryStats) {
			atomic.AddUint64(&stats.GCTriggers, 1)
		})
	}
}

// CleanupStaleAllocations removes allocations that haven't been accessed recently
func (mm *MemoryManager) CleanupStaleAllocations(maxAge time.Duration) int {
	mm.mutex.Lock()
	defer mm.mutex.Unlock()

	now := time.Now()
	var staleIDs []uint64

	for id, alloc := range mm.allocations {
		alloc.mutex.RLock()
		if now.Sub(alloc.LastAccessed) > maxAge && !alloc.Disposed {
			staleIDs = append(staleIDs, id)
		}
		alloc.mutex.RUnlock()
	}

	// Clean up stale allocations
	for _, id := range staleIDs {
		if alloc, exists := mm.allocations[id]; exists {
			if alloc.CleanupFunc != nil {
				alloc.CleanupFunc()
			}
			mm.ReleaseAllocation(id)
		}
	}

	return len(staleIDs)
}

// GetMemoryStats returns current memory statistics
func (mm *MemoryManager) GetMemoryStats() MemoryStats {
	mm.stats.mutex.RLock()
	defer mm.stats.mutex.RUnlock()

	stats := mm.stats
	stats.LastUpdate = time.Now()
	return stats
}

// ------------------------------------
// 🔄 Background Monitoring
// ------------------------------------

// backgroundMonitoring runs continuous memory monitoring
func (mm *MemoryManager) backgroundMonitoring() {
	for {
		select {
		case <-mm.ctx.Done():
			return
		case <-mm.stopBackground:
			return
		case <-mm.backgroundTicker.C:
			mm.performBackgroundTasks()
		}
	}
}

// performBackgroundTasks executes periodic maintenance tasks
func (mm *MemoryManager) performBackgroundTasks() {
	if !mm.monitoring {
		return
	}

	// Update memory statistics
	mm.updateMemoryMetrics()

	// Check for memory leaks
	if mm.leakDetector != nil {
		mm.leakDetector.PerformLeakCheck()
	}

	// Clean up stale allocations
	staleCount := mm.CleanupStaleAllocations(mm.config.HistoryRetention)
	if staleCount > 0 && mm.debugMode {
		fmt.Printf("🧹 Cleaned up %d stale allocations\n", staleCount)
	}

	// Trigger GC if needed
	mm.ForceGC()

	// Update resource monitor
	if mm.resourceMonitor != nil {
		mm.resourceMonitor.UpdateMetrics()
	}
}

// updateMemoryMetrics updates runtime memory metrics
func (mm *MemoryManager) updateMemoryMetrics() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	mm.updateStats(func(stats *MemoryStats) {
		stats.LastUpdate = time.Now()
		// Additional runtime metrics could be added here
	})
}

// updateStats safely updates memory statistics
func (mm *MemoryManager) updateStats(fn func(*MemoryStats)) {
	mm.stats.mutex.Lock()
	defer mm.stats.mutex.Unlock()
	fn(&mm.stats)
}

// ------------------------------------
// 🛠️ Configuration and Control
// ------------------------------------

// SetEnabled enables or disables memory management
func (mm *MemoryManager) SetEnabled(enabled bool) {
	mm.mutex.Lock()
	defer mm.mutex.Unlock()
	mm.enabled = enabled
}

// SetMonitoring enables or disables background monitoring
func (mm *MemoryManager) SetMonitoring(monitoring bool) {
	mm.mutex.Lock()
	defer mm.mutex.Unlock()
	mm.monitoring = monitoring
}

// SetDebugMode enables or disables debug mode
func (mm *MemoryManager) SetDebugMode(debug bool) {
	mm.mutex.Lock()
	defer mm.mutex.Unlock()
	mm.debugMode = debug
}

// UpdateConfig updates the memory manager configuration
func (mm *MemoryManager) UpdateConfig(config MemoryConfig) {
	mm.mutex.Lock()
	defer mm.mutex.Unlock()

	mm.config = config
	mm.debugMode = config.EnableDetailedLogging

	// Update ticker interval if changed
	if mm.backgroundTicker != nil {
		mm.backgroundTicker.Stop()
		mm.backgroundTicker = time.NewTicker(config.MonitoringInterval)
	}
}

// ------------------------------------
// 🧹 Cleanup and Disposal
// ------------------------------------

// Dispose cleans up the memory manager and all its resources
func (mm *MemoryManager) Dispose() {
	mm.mutex.Lock()
	defer mm.mutex.Unlock()

	// Stop background monitoring
	if mm.cancel != nil {
		mm.cancel()
	}

	if mm.backgroundTicker != nil {
		mm.backgroundTicker.Stop()
	}

	select {
	case mm.stopBackground <- true:
	default:
	}

	// Dispose sub-components
	if mm.resourceRegistry != nil {
		mm.resourceRegistry.Dispose()
	}

	if mm.cleanupTracker != nil {
		mm.cleanupTracker.Dispose()
	}

	if mm.leakDetector != nil {
		mm.leakDetector.Dispose()
	}

	if mm.resourceMonitor != nil {
		mm.resourceMonitor.Dispose()
	}

	if mm.cleanupScheduler != nil {
		mm.cleanupScheduler.Dispose()
	}

	// Clean up all remaining allocations
	for id, alloc := range mm.allocations {
		if alloc.CleanupFunc != nil {
			alloc.CleanupFunc()
		}
		delete(mm.allocations, id)
	}

	mm.enabled = false
	mm.monitoring = false
}

// ------------------------------------
// 🔍 Debugging and Introspection
// ------------------------------------

// GetDetailedReport returns a detailed memory usage report
func (mm *MemoryManager) GetDetailedReport() map[string]interface{} {
	mm.mutex.RLock()
	defer mm.mutex.RUnlock()

	report := map[string]interface{}{
		"manager_id":       mm.id,
		"enabled":          mm.enabled,
		"monitoring":       mm.monitoring,
		"debug_mode":       mm.debugMode,
		"config":           mm.config,
		"stats":            mm.GetMemoryStats(),
		"allocation_count": len(mm.allocations),
	}

	// Add allocation breakdown by type
	typeBreakdown := make(map[string]int)
	sizeBreakdown := make(map[string]int64)

	for _, alloc := range mm.allocations {
		if !alloc.Disposed {
			typeName := alloc.Type.String()
			typeBreakdown[typeName]++
			sizeBreakdown[typeName] += alloc.Size
		}
	}

	report["allocations_by_type"] = typeBreakdown
	report["memory_by_type"] = sizeBreakdown

	// Add sub-component reports
	if mm.leakDetector != nil {
		report["leak_detector"] = mm.leakDetector.GetReport()
	}

	if mm.resourceMonitor != nil {
		report["resource_monitor"] = mm.resourceMonitor.GetReport()
	}

	if mm.cleanupTracker != nil {
		report["cleanup_tracker"] = mm.cleanupTracker.GetReport()
	}

	return report
}

// DumpAllocations returns detailed information about all active allocations
func (mm *MemoryManager) DumpAllocations() []map[string]interface{} {
	mm.mutex.RLock()
	defer mm.mutex.RUnlock()

	var allocations []map[string]interface{}

	for _, alloc := range mm.allocations {
		if !alloc.Disposed {
			allocInfo := map[string]interface{}{
				"id":            alloc.ID,
				"type":          alloc.Type.String(),
				"size":          alloc.Size,
				"timestamp":     alloc.Timestamp,
				"last_accessed": alloc.LastAccessed,
				"access_count":  alloc.AccessCount,
				"metadata":      alloc.Metadata,
			}

			if alloc.Owner != nil {
				allocInfo["owner_id"] = alloc.Owner.id
			}

			if mm.config.EnableStackTraces && len(alloc.StackTrace) > 0 {
				frames := runtime.CallersFrames(alloc.StackTrace)
				var stackInfo []string
				for {
					frame, more := frames.Next()
					stackInfo = append(stackInfo, fmt.Sprintf("%s:%d %s", frame.File, frame.Line, frame.Function))
					if !more {
						break
					}
				}
				allocInfo["stack_trace"] = stackInfo
			}

			allocations = append(allocations, allocInfo)
		}
	}

	return allocations
}
