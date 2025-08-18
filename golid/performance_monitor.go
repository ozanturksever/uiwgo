// performance_monitor.go
// Production performance monitoring and optimization utilities

package golid

import (
	"encoding/json"
	"fmt"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

// ------------------------------------
// 📊 Performance Monitor
// ------------------------------------

// PerformanceMonitor provides real-time performance monitoring for production
type PerformanceMonitor struct {
	enabled        bool
	metrics        *PerformanceMetrics
	alerts         *PerformanceAlertManager
	profiler       *ReactivityProfiler
	optimizer      *RuntimeOptimizer
	mutex          sync.RWMutex
	stopChan       chan bool
	reportInterval time.Duration
	thresholds     PerformanceThresholds
}

// PerformanceMetrics tracks comprehensive system performance
type PerformanceMetrics struct {
	// Signal Performance
	SignalUpdates       uint64        `json:"signal_updates"`
	SignalUpdateLatency time.Duration `json:"signal_update_latency"`
	EffectExecutions    uint64        `json:"effect_executions"`
	EffectLatency       time.Duration `json:"effect_latency"`

	// DOM Performance
	DOMUpdates        uint64        `json:"dom_updates"`
	DOMUpdateLatency  time.Duration `json:"dom_update_latency"`
	BatchedOperations uint64        `json:"batched_operations"`

	// Memory Performance
	MemoryUsage  uint64 `json:"memory_usage"`
	SignalMemory uint64 `json:"signal_memory"`
	EffectMemory uint64 `json:"effect_memory"`
	GCPauses     uint64 `json:"gc_pauses"`

	// Scheduler Performance
	SchedulerQueue int `json:"scheduler_queue"`
	BatchDepth     int `json:"batch_depth"`
	CascadeDepth   int `json:"cascade_depth"`

	// Error Tracking
	ErrorCount    uint64 `json:"error_count"`
	RecoveryCount uint64 `json:"recovery_count"`

	// Timestamps
	LastUpdate time.Time `json:"last_update"`
	StartTime  time.Time `json:"start_time"`

	mutex sync.RWMutex
}

// PerformanceThresholds defines alert thresholds
type PerformanceThresholds struct {
	MaxSignalLatency   time.Duration `json:"max_signal_latency"`
	MaxDOMLatency      time.Duration `json:"max_dom_latency"`
	MaxMemoryPerSignal uint64        `json:"max_memory_per_signal"`
	MaxCascadeDepth    int           `json:"max_cascade_depth"`
	MaxErrorRate       float64       `json:"max_error_rate"`
}

// PerformanceAlertManager handles performance alerts (renamed to avoid conflicts)
type PerformanceAlertManager struct {
	handlers []PerformanceAlertHandler
	enabled  bool
	mutex    sync.RWMutex
}

type PerformanceAlertHandler func(alert PerformanceAlert)

type PerformanceAlert struct {
	Type      string      `json:"type"`
	Severity  string      `json:"severity"`
	Message   string      `json:"message"`
	Value     interface{} `json:"value"`
	Threshold interface{} `json:"threshold"`
	Timestamp time.Time   `json:"timestamp"`
}

// ReactivityProfiler provides detailed profiling
type ReactivityProfiler struct {
	enabled        bool
	signalProfiles map[uint64]*SignalProfile
	effectProfiles map[uint64]*EffectProfile
	mutex          sync.RWMutex
}

type SignalProfile struct {
	ID              uint64        `json:"id"`
	UpdateCount     uint64        `json:"update_count"`
	TotalLatency    time.Duration `json:"total_latency"`
	AverageLatency  time.Duration `json:"average_latency"`
	SubscriberCount int           `json:"subscriber_count"`
	LastUpdate      time.Time     `json:"last_update"`
}

type EffectProfile struct {
	ID              uint64        `json:"id"`
	ExecutionCount  uint64        `json:"execution_count"`
	TotalLatency    time.Duration `json:"total_latency"`
	AverageLatency  time.Duration `json:"average_latency"`
	DependencyCount int           `json:"dependency_count"`
	LastExecution   time.Time     `json:"last_execution"`
}

// RuntimeOptimizer provides automatic optimizations
type RuntimeOptimizer struct {
	enabled            bool
	gcOptimization     bool
	batchOptimization  bool
	memoryOptimization bool
	mutex              sync.RWMutex
}

// ------------------------------------
// 🏭 Global Performance Monitor
// ------------------------------------

var (
	globalPerformanceMonitor *PerformanceMonitor
	performanceMonitorOnce   sync.Once
)

// GetPerformanceMonitor returns the global performance monitor
func GetPerformanceMonitor() *PerformanceMonitor {
	performanceMonitorOnce.Do(func() {
		globalPerformanceMonitor = &PerformanceMonitor{
			enabled:        false,
			metrics:        NewPerformanceMetrics(),
			alerts:         NewPerformanceAlertManager(),
			profiler:       NewReactivityProfiler(),
			optimizer:      NewRuntimeOptimizer(),
			stopChan:       make(chan bool),
			reportInterval: 5 * time.Second,
			thresholds:     GetDefaultThresholds(),
		}
	})
	return globalPerformanceMonitor
}

// NewPerformanceMetrics creates new performance metrics
func NewPerformanceMetrics() *PerformanceMetrics {
	return &PerformanceMetrics{
		StartTime:  time.Now(),
		LastUpdate: time.Now(),
	}
}

// NewPerformanceAlertManager creates new performance alert manager
func NewPerformanceAlertManager() *PerformanceAlertManager {
	return &PerformanceAlertManager{
		handlers: make([]PerformanceAlertHandler, 0),
		enabled:  false,
	}
}

// NewReactivityProfiler creates new profiler
func NewReactivityProfiler() *ReactivityProfiler {
	return &ReactivityProfiler{
		enabled:        false,
		signalProfiles: make(map[uint64]*SignalProfile),
		effectProfiles: make(map[uint64]*EffectProfile),
	}
}

// NewRuntimeOptimizer creates new optimizer
func NewRuntimeOptimizer() *RuntimeOptimizer {
	return &RuntimeOptimizer{
		enabled:            false,
		gcOptimization:     true,
		batchOptimization:  true,
		memoryOptimization: true,
	}
}

// GetDefaultThresholds returns default performance thresholds
func GetDefaultThresholds() PerformanceThresholds {
	return PerformanceThresholds{
		MaxSignalLatency:   5 * time.Microsecond,
		MaxDOMLatency:      10 * time.Millisecond,
		MaxMemoryPerSignal: 200, // bytes
		MaxCascadeDepth:    10,
		MaxErrorRate:       0.01, // 1%
	}
}

// ------------------------------------
// 🎯 Performance Monitor API
// ------------------------------------

// Enable starts performance monitoring
func (pm *PerformanceMonitor) Enable() {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	if pm.enabled {
		return
	}

	pm.enabled = true
	pm.metrics.StartTime = time.Now()

	// Start monitoring goroutine
	go pm.monitorLoop()
}

// Disable stops performance monitoring
func (pm *PerformanceMonitor) Disable() {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	if !pm.enabled {
		return
	}

	pm.enabled = false
	close(pm.stopChan)
	pm.stopChan = make(chan bool)
}

// IsEnabled returns monitoring status
func (pm *PerformanceMonitor) IsEnabled() bool {
	pm.mutex.RLock()
	defer pm.mutex.RUnlock()
	return pm.enabled
}

// RecordSignalUpdate records signal update performance
func (pm *PerformanceMonitor) RecordSignalUpdate(signalID uint64, latency time.Duration) {
	if !pm.enabled {
		return
	}

	atomic.AddUint64(&pm.metrics.SignalUpdates, 1)
	pm.updateLatency(&pm.metrics.SignalUpdateLatency, latency)

	// Profile individual signal
	if pm.profiler.enabled {
		pm.profiler.RecordSignalUpdate(signalID, latency)
	}

	// Check thresholds
	if latency > pm.thresholds.MaxSignalLatency {
		pm.alerts.TriggerAlert(PerformanceAlert{
			Type:      "signal_latency",
			Severity:  "warning",
			Message:   fmt.Sprintf("Signal update latency %v exceeds threshold %v", latency, pm.thresholds.MaxSignalLatency),
			Value:     latency,
			Threshold: pm.thresholds.MaxSignalLatency,
			Timestamp: time.Now(),
		})
	}
}

// RecordEffectExecution records effect execution performance
func (pm *PerformanceMonitor) RecordEffectExecution(effectID uint64, latency time.Duration) {
	if !pm.enabled {
		return
	}

	atomic.AddUint64(&pm.metrics.EffectExecutions, 1)
	pm.updateLatency(&pm.metrics.EffectLatency, latency)

	// Profile individual effect
	if pm.profiler.enabled {
		pm.profiler.RecordEffectExecution(effectID, latency)
	}
}

// RecordDOMUpdate records DOM update performance
func (pm *PerformanceMonitor) RecordDOMUpdate(latency time.Duration, batched bool) {
	if !pm.enabled {
		return
	}

	atomic.AddUint64(&pm.metrics.DOMUpdates, 1)
	pm.updateLatency(&pm.metrics.DOMUpdateLatency, latency)

	if batched {
		atomic.AddUint64(&pm.metrics.BatchedOperations, 1)
	}

	// Check thresholds
	if latency > pm.thresholds.MaxDOMLatency {
		pm.alerts.TriggerAlert(PerformanceAlert{
			Type:      "dom_latency",
			Severity:  "warning",
			Message:   fmt.Sprintf("DOM update latency %v exceeds threshold %v", latency, pm.thresholds.MaxDOMLatency),
			Value:     latency,
			Threshold: pm.thresholds.MaxDOMLatency,
			Timestamp: time.Now(),
		})
	}
}

// RecordError records error occurrence
func (pm *PerformanceMonitor) RecordError() {
	if !pm.enabled {
		return
	}

	atomic.AddUint64(&pm.metrics.ErrorCount, 1)
}

// RecordRecovery records error recovery
func (pm *PerformanceMonitor) RecordRecovery() {
	if !pm.enabled {
		return
	}

	atomic.AddUint64(&pm.metrics.RecoveryCount, 1)
}

// UpdateMemoryStats updates memory statistics
func (pm *PerformanceMonitor) UpdateMemoryStats() {
	if !pm.enabled {
		return
	}

	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	pm.metrics.mutex.Lock()
	pm.metrics.MemoryUsage = m.Alloc
	pm.metrics.GCPauses = uint64(m.NumGC)
	pm.metrics.LastUpdate = time.Now()
	pm.metrics.mutex.Unlock()
}

// GetMetrics returns current performance metrics
func (pm *PerformanceMonitor) GetMetrics() PerformanceMetrics {
	pm.metrics.mutex.RLock()
	defer pm.metrics.mutex.RUnlock()

	// Update runtime stats
	pm.UpdateMemoryStats()

	return *pm.metrics
}

// GetMetricsJSON returns metrics as JSON
func (pm *PerformanceMonitor) GetMetricsJSON() ([]byte, error) {
	metrics := pm.GetMetrics()
	return json.MarshalIndent(metrics, "", "  ")
}

// ------------------------------------
// 🔧 Internal Methods
// ------------------------------------

// updateLatency updates running average latency
func (pm *PerformanceMonitor) updateLatency(current *time.Duration, newLatency time.Duration) {
	// Simple exponential moving average
	alpha := 0.1
	if *current == 0 {
		*current = newLatency
	} else {
		*current = time.Duration(float64(*current)*(1-alpha) + float64(newLatency)*alpha)
	}
}

// monitorLoop runs the monitoring loop
func (pm *PerformanceMonitor) monitorLoop() {
	ticker := time.NewTicker(pm.reportInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			pm.UpdateMemoryStats()
			pm.checkThresholds()

			if pm.optimizer.enabled {
				pm.optimizer.OptimizeRuntime(pm.metrics)
			}

		case <-pm.stopChan:
			return
		}
	}
}

// checkThresholds checks all performance thresholds
func (pm *PerformanceMonitor) checkThresholds() {
	metrics := pm.GetMetrics()

	// Check cascade depth
	scheduler := getScheduler()
	if scheduler != nil {
		stats := scheduler.GetStats()
		if stats.QueueSize > pm.thresholds.MaxCascadeDepth {
			pm.alerts.TriggerAlert(PerformanceAlert{
				Type:      "cascade_depth",
				Severity:  "error",
				Message:   fmt.Sprintf("Queue size %d exceeds threshold %d", stats.QueueSize, pm.thresholds.MaxCascadeDepth),
				Value:     stats.QueueSize,
				Threshold: pm.thresholds.MaxCascadeDepth,
				Timestamp: time.Now(),
			})
		}
	}

	// Check error rate
	if metrics.SignalUpdates > 0 {
		errorRate := float64(metrics.ErrorCount) / float64(metrics.SignalUpdates)
		if errorRate > pm.thresholds.MaxErrorRate {
			pm.alerts.TriggerAlert(PerformanceAlert{
				Type:      "error_rate",
				Severity:  "error",
				Message:   fmt.Sprintf("Error rate %.2f%% exceeds threshold %.2f%%", errorRate*100, pm.thresholds.MaxErrorRate*100),
				Value:     errorRate,
				Threshold: pm.thresholds.MaxErrorRate,
				Timestamp: time.Now(),
			})
		}
	}
}

// ------------------------------------
// 🚨 Alert Manager
// ------------------------------------

// AddHandler adds an alert handler
func (am *PerformanceAlertManager) AddHandler(handler PerformanceAlertHandler) {
	am.mutex.Lock()
	defer am.mutex.Unlock()
	am.handlers = append(am.handlers, handler)
}

// Enable enables alert processing
func (am *PerformanceAlertManager) Enable() {
	am.mutex.Lock()
	defer am.mutex.Unlock()
	am.enabled = true
}

// Disable disables alert processing
func (am *PerformanceAlertManager) Disable() {
	am.mutex.Lock()
	defer am.mutex.Unlock()
	am.enabled = false
}

// TriggerAlert triggers an alert
func (am *PerformanceAlertManager) TriggerAlert(alert PerformanceAlert) {
	am.mutex.RLock()
	defer am.mutex.RUnlock()

	if !am.enabled {
		return
	}

	for _, handler := range am.handlers {
		go handler(alert)
	}
}

// ------------------------------------
// 📈 Profiler Methods
// ------------------------------------

// Enable enables profiling
func (rp *ReactivityProfiler) Enable() {
	rp.mutex.Lock()
	defer rp.mutex.Unlock()
	rp.enabled = true
}

// Disable disables profiling
func (rp *ReactivityProfiler) Disable() {
	rp.mutex.Lock()
	defer rp.mutex.Unlock()
	rp.enabled = false
}

// RecordSignalUpdate records signal update for profiling
func (rp *ReactivityProfiler) RecordSignalUpdate(signalID uint64, latency time.Duration) {
	rp.mutex.Lock()
	defer rp.mutex.Unlock()

	profile, exists := rp.signalProfiles[signalID]
	if !exists {
		profile = &SignalProfile{
			ID: signalID,
		}
		rp.signalProfiles[signalID] = profile
	}

	profile.UpdateCount++
	profile.TotalLatency += latency
	profile.AverageLatency = profile.TotalLatency / time.Duration(profile.UpdateCount)
	profile.LastUpdate = time.Now()
}

// RecordEffectExecution records effect execution for profiling
func (rp *ReactivityProfiler) RecordEffectExecution(effectID uint64, latency time.Duration) {
	rp.mutex.Lock()
	defer rp.mutex.Unlock()

	profile, exists := rp.effectProfiles[effectID]
	if !exists {
		profile = &EffectProfile{
			ID: effectID,
		}
		rp.effectProfiles[effectID] = profile
	}

	profile.ExecutionCount++
	profile.TotalLatency += latency
	profile.AverageLatency = profile.TotalLatency / time.Duration(profile.ExecutionCount)
	profile.LastExecution = time.Now()
}

// GetSignalProfiles returns all signal profiles
func (rp *ReactivityProfiler) GetSignalProfiles() map[uint64]*SignalProfile {
	rp.mutex.RLock()
	defer rp.mutex.RUnlock()

	profiles := make(map[uint64]*SignalProfile)
	for id, profile := range rp.signalProfiles {
		profileCopy := *profile
		profiles[id] = &profileCopy
	}
	return profiles
}

// GetEffectProfiles returns all effect profiles
func (rp *ReactivityProfiler) GetEffectProfiles() map[uint64]*EffectProfile {
	rp.mutex.RLock()
	defer rp.mutex.RUnlock()

	profiles := make(map[uint64]*EffectProfile)
	for id, profile := range rp.effectProfiles {
		profileCopy := *profile
		profiles[id] = &profileCopy
	}
	return profiles
}

// ------------------------------------
// ⚡ Runtime Optimizer
// ------------------------------------

// Enable enables runtime optimization
func (ro *RuntimeOptimizer) Enable() {
	ro.mutex.Lock()
	defer ro.mutex.Unlock()
	ro.enabled = true
}

// Disable disables runtime optimization
func (ro *RuntimeOptimizer) Disable() {
	ro.mutex.Lock()
	defer ro.mutex.Unlock()
	ro.enabled = false
}

// OptimizeRuntime performs runtime optimizations
func (ro *RuntimeOptimizer) OptimizeRuntime(metrics *PerformanceMetrics) {
	ro.mutex.RLock()
	defer ro.mutex.RUnlock()

	if !ro.enabled {
		return
	}

	// GC optimization
	if ro.gcOptimization {
		ro.optimizeGC(metrics)
	}

	// Batch optimization
	if ro.batchOptimization {
		ro.optimizeBatching(metrics)
	}

	// Memory optimization
	if ro.memoryOptimization {
		ro.optimizeMemory(metrics)
	}
}

// optimizeGC optimizes garbage collection
func (ro *RuntimeOptimizer) optimizeGC(metrics *PerformanceMetrics) {
	// Trigger GC if memory usage is high
	if metrics.MemoryUsage > 100*1024*1024 { // 100MB threshold
		runtime.GC()
	}
}

// optimizeBatching optimizes batching parameters
func (ro *RuntimeOptimizer) optimizeBatching(metrics *PerformanceMetrics) {
	scheduler := getScheduler()
	if scheduler == nil {
		return
	}

	// Adjust batch size based on performance
	if metrics.DOMUpdateLatency > 5*time.Millisecond {
		// Increase batch size to reduce overhead
		// Implementation would adjust scheduler parameters
	}
}

// optimizeMemory optimizes memory usage
func (ro *RuntimeOptimizer) optimizeMemory(metrics *PerformanceMetrics) {
	// Clean up unused signal profiles
	monitor := GetPerformanceMonitor()
	if monitor.profiler.enabled {
		// Implementation would clean old profiles
	}
}

// ------------------------------------
// 🎯 Global API
// ------------------------------------

// EnablePerformanceMonitoring enables global performance monitoring
func EnablePerformanceMonitoring() {
	GetPerformanceMonitor().Enable()
}

// DisablePerformanceMonitoring disables global performance monitoring
func DisablePerformanceMonitoring() {
	GetPerformanceMonitor().Disable()
}

// GetPerformanceMetrics returns current performance metrics
func GetPerformanceMetrics() PerformanceMetrics {
	return GetPerformanceMonitor().GetMetrics()
}

// GetPerformanceMetricsJSON returns metrics as JSON
func GetPerformanceMetricsJSON() ([]byte, error) {
	return GetPerformanceMonitor().GetMetricsJSON()
}

// AddPerformanceAlertHandler adds a performance alert handler
func AddPerformanceAlertHandler(handler PerformanceAlertHandler) {
	GetPerformanceMonitor().alerts.AddHandler(handler)
}

// EnablePerformanceProfiling enables detailed profiling
func EnablePerformanceProfiling() {
	GetPerformanceMonitor().profiler.Enable()
}

// DisablePerformanceProfiling disables detailed profiling
func DisablePerformanceProfiling() {
	GetPerformanceMonitor().profiler.Disable()
}

// EnableRuntimeOptimization enables automatic runtime optimization
func EnableRuntimeOptimization() {
	GetPerformanceMonitor().optimizer.Enable()
}

// DisableRuntimeOptimization disables automatic runtime optimization
func DisableRuntimeOptimization() {
	GetPerformanceMonitor().optimizer.Disable()
}
