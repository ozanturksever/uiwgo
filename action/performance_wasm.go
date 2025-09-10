//go:build js && wasm
// +build js,wasm

package action

import (
	"sync"
	"time"
)

// PerformanceConfig holds configuration for performance optimizations
// In WASM builds, most optimizations are disabled or simplified
type PerformanceConfig struct {
	// Object pools - disabled in WASM
	EnableObjectPooling bool
	ActionPoolSize      int
	ContextPoolSize     int
	SubscriberPoolSize  int

	// Batching - simplified in WASM
	EnableReactiveBatching bool
	BatchWindow            time.Duration
	BatchSize              int

	// Async scheduling - simplified in WASM
	EnableMicrotaskScheduler bool
	MicrotaskQueueSize       int
	WorkerPoolSize           int

	// Profiling - basic in WASM
	EnableProfiling     bool
	ProfilingLevel      ProfilingLevel
	MemoryTrackingLevel MemoryTrackingLevel
}

// ProfilingLevel defines the level of profiling detail
type ProfilingLevel int

const (
	ProfilingOff ProfilingLevel = iota
	ProfilingBasic
	ProfilingDetailed
	ProfilingVerbose
)

// MemoryTrackingLevel defines memory tracking detail level
type MemoryTrackingLevel int

const (
	MemoryTrackingOff MemoryTrackingLevel = iota
	MemoryTrackingBasic
	MemoryTrackingDetailed
)

// DefaultPerformanceConfig returns a default configuration optimized for WASM
func DefaultPerformanceConfig() PerformanceConfig {
	return PerformanceConfig{
		// Disable object pooling in WASM for simplicity
		EnableObjectPooling: false,
		ActionPoolSize:      0,
		ContextPoolSize:     0,
		SubscriberPoolSize:  0,

		// Simplified batching
		EnableReactiveBatching: false,
		BatchWindow:            10 * time.Millisecond,
		BatchSize:              50,

		// Simplified scheduling
		EnableMicrotaskScheduler: false,
		MicrotaskQueueSize:       100,
		WorkerPoolSize:           1,

		// Basic profiling only
		EnableProfiling:     false,
		ProfilingLevel:      ProfilingOff,
		MemoryTrackingLevel: MemoryTrackingOff,
	}
}

// performanceManager is a simplified version for WASM builds
type performanceManager struct {
	config PerformanceConfig
	batchProcessor *batchProcessor // Stub for compatibility
	mu     sync.RWMutex
}

// batchProcessor is a stub implementation for WASM
type batchProcessor struct {
	// Empty struct for WASM compatibility
}

// Global performance manager
var (
	globalPerfManager *performanceManager
	perfManagerOnce   sync.Once
)

// GetPerformanceManager returns the global performance manager
func GetPerformanceManager() *performanceManager {
	perfManagerOnce.Do(func() {
		globalPerfManager = NewPerformanceManager(DefaultPerformanceConfig())
	})
	return globalPerfManager
}

// NewPerformanceManager creates a new performance manager with the given config
func NewPerformanceManager(config PerformanceConfig) *performanceManager {
	return &performanceManager{
		config: config,
		batchProcessor: nil, // Disabled in WASM builds
	}
}

// Stub implementations for WASM compatibility

// EnablePerformanceOptimizations enables performance optimizations (no-op in WASM)
func EnablePerformanceOptimizations(config PerformanceConfig) {
	// No-op in WASM builds
}

// SetPerformanceConfig sets the performance configuration (no-op in WASM)
func SetPerformanceConfig(config PerformanceConfig) {
	pm := GetPerformanceManager()
	pm.mu.Lock()
	pm.config = config
	pm.mu.Unlock()
}

// GetPerformanceConfig returns the current performance configuration
func GetPerformanceConfig() PerformanceConfig {
	pm := GetPerformanceManager()
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	return pm.config
}

// dispatchMetrics is a simplified version for WASM
type dispatchMetrics struct {
	count           int64
	totalDuration   time.Duration
	avgDuration     time.Duration
	maxDuration     time.Duration
	minDuration     time.Duration
	totalAllocBytes int64
	totalAllocCount int64
	lastUpdated     time.Time
}

// GetDispatchMetrics returns dispatch metrics (simplified in WASM)
func GetDispatchMetrics(actionType string) *dispatchMetrics {
	// Return empty metrics in WASM builds
	return &dispatchMetrics{}
}

// ResetPerformanceMetrics resets all performance metrics (no-op in WASM)
func ResetPerformanceMetrics() {
	// No-op in WASM builds
}

// OptimizedDispatch performs an optimized dispatch (simplified in WASM)
func OptimizedDispatch(bus Bus, action any, opts ...DispatchOption) error {
	// In WASM, just use the regular dispatch
	return bus.Dispatch(action, opts...)
}