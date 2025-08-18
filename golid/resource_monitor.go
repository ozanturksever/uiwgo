// resource_monitor.go
// Real-time resource usage monitoring and reporting system

package golid

import (
	"context"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

// ------------------------------------
// 📊 Resource Monitor Core
// ------------------------------------

var resourceMonitorIdCounter uint64

// ResourceMonitor provides real-time monitoring of resource usage
type ResourceMonitor struct {
	id            uint64
	memoryManager *MemoryManager

	// Monitoring data
	metrics *ResourceMetrics
	history *MetricsHistory
	alerts  *AlertManager

	// Configuration
	config MonitorConfig

	// State
	enabled    bool
	collecting bool

	// Background monitoring
	ctx            context.Context
	cancel         context.CancelFunc
	monitorTicker  *time.Ticker
	stopMonitoring chan bool

	// Synchronization
	mutex sync.RWMutex
}

// MonitorConfig provides configuration for resource monitoring
type MonitorConfig struct {
	CollectionInterval    time.Duration   `json:"collection_interval"`
	HistoryRetention      time.Duration   `json:"history_retention"`
	HistoryBufferSize     int             `json:"history_buffer_size"`
	AlertThresholds       AlertThresholds `json:"alert_thresholds"`
	EnableDetailedMetrics bool            `json:"enable_detailed_metrics"`
	EnableAlerts          bool            `json:"enable_alerts"`
	EnableTrendAnalysis   bool            `json:"enable_trend_analysis"`
}

// AlertThresholds defines thresholds for various alerts
type AlertThresholds struct {
	MemoryUsage        int64   `json:"memory_usage"`         // Bytes
	AllocationRate     float64 `json:"allocation_rate"`      // Allocations per second
	CleanupFailureRate float64 `json:"cleanup_failure_rate"` // Percentage
	LeakDetectionRate  float64 `json:"leak_detection_rate"`  // Leaks per minute
	GCPressure         float64 `json:"gc_pressure"`          // GC frequency
	ResourceCount      int     `json:"resource_count"`       // Total resources
}

// ResourceMetrics contains current resource usage metrics
type ResourceMetrics struct {
	// Memory metrics
	TotalMemoryUsage  int64   `json:"total_memory_usage"`
	ActiveAllocations uint64  `json:"active_allocations"`
	AllocationRate    float64 `json:"allocation_rate"`
	DeallocationRate  float64 `json:"deallocation_rate"`

	// Resource metrics
	SignalCount            uint64 `json:"signal_count"`
	EffectCount            uint64 `json:"effect_count"`
	MemoCount              uint64 `json:"memo_count"`
	ResourceCount          uint64 `json:"resource_count"`
	DOMBindingCount        uint64 `json:"dom_binding_count"`
	EventSubscriptionCount uint64 `json:"event_subscription_count"`
	ComponentCount         uint64 `json:"component_count"`

	// Performance metrics
	AverageCleanupTime time.Duration `json:"average_cleanup_time"`
	CleanupSuccessRate float64       `json:"cleanup_success_rate"`
	LeakDetectionRate  float64       `json:"leak_detection_rate"`

	// System metrics
	GoroutineCount int       `json:"goroutine_count"`
	GCStats        GCMetrics `json:"gc_stats"`

	// Timestamps
	Timestamp          time.Time     `json:"timestamp"`
	CollectionDuration time.Duration `json:"collection_duration"`

	mutex sync.RWMutex
}

// GCMetrics contains garbage collection metrics
type GCMetrics struct {
	NumGC        uint32        `json:"num_gc"`
	PauseTotal   time.Duration `json:"pause_total"`
	PauseNs      []uint64      `json:"pause_ns"`
	LastGC       time.Time     `json:"last_gc"`
	NextGC       uint64        `json:"next_gc"`
	HeapAlloc    uint64        `json:"heap_alloc"`
	HeapSys      uint64        `json:"heap_sys"`
	HeapIdle     uint64        `json:"heap_idle"`
	HeapInuse    uint64        `json:"heap_inuse"`
	HeapReleased uint64        `json:"heap_released"`
	HeapObjects  uint64        `json:"heap_objects"`
}

// MetricsHistory stores historical metrics data
type MetricsHistory struct {
	entries      []*ResourceMetrics
	maxSize      int
	currentIndex int
	full         bool
	mutex        sync.RWMutex
}

// AlertManager handles resource monitoring alerts
type AlertManager struct {
	thresholds   AlertThresholds
	activeAlerts map[string]*Alert
	alertHistory []*Alert
	enabled      bool
	mutex        sync.RWMutex
}

// Alert represents a resource monitoring alert
type Alert struct {
	ID         string                 `json:"id"`
	Type       AlertType              `json:"type"`
	Severity   AlertSeverity          `json:"severity"`
	Message    string                 `json:"message"`
	Timestamp  time.Time              `json:"timestamp"`
	Value      interface{}            `json:"value"`
	Threshold  interface{}            `json:"threshold"`
	Metadata   map[string]interface{} `json:"metadata"`
	Resolved   bool                   `json:"resolved"`
	ResolvedAt time.Time              `json:"resolved_at,omitempty"`
}

// AlertType defines types of alerts
type AlertType string

const (
	AlertMemoryUsage    AlertType = "memory_usage"
	AlertAllocationRate AlertType = "allocation_rate"
	AlertCleanupFailure AlertType = "cleanup_failure"
	AlertLeakDetection  AlertType = "leak_detection"
	AlertGCPressure     AlertType = "gc_pressure"
	AlertResourceCount  AlertType = "resource_count"
)

// AlertSeverity defines alert severity levels
type AlertSeverity string

const (
	AlertInfo     AlertSeverity = "info"
	AlertWarning  AlertSeverity = "warning"
	AlertCritical AlertSeverity = "critical"
)

// ------------------------------------
// 🏗️ Resource Monitor Creation
// ------------------------------------

// NewResourceMonitor creates a new resource monitor
func NewResourceMonitor(memoryManager *MemoryManager) *ResourceMonitor {
	ctx, cancel := context.WithCancel(context.Background())

	config := DefaultMonitorConfig()

	rm := &ResourceMonitor{
		id:             atomic.AddUint64(&resourceMonitorIdCounter, 1),
		memoryManager:  memoryManager,
		metrics:        &ResourceMetrics{},
		history:        NewMetricsHistory(config.HistoryBufferSize),
		alerts:         NewAlertManager(config.AlertThresholds),
		config:         config,
		enabled:        true,
		collecting:     true,
		ctx:            ctx,
		cancel:         cancel,
		monitorTicker:  time.NewTicker(config.CollectionInterval),
		stopMonitoring: make(chan bool, 1),
	}

	// Start background monitoring
	go rm.backgroundMonitoring()

	return rm
}

// DefaultMonitorConfig returns default monitoring configuration
func DefaultMonitorConfig() MonitorConfig {
	return MonitorConfig{
		CollectionInterval:    5 * time.Second,
		HistoryRetention:      1 * time.Hour,
		HistoryBufferSize:     720, // 1 hour at 5-second intervals
		EnableDetailedMetrics: true,
		EnableAlerts:          true,
		EnableTrendAnalysis:   true,
		AlertThresholds: AlertThresholds{
			MemoryUsage:        100 * 1024 * 1024, // 100MB
			AllocationRate:     1000,              // 1000 allocations/sec
			CleanupFailureRate: 5.0,               // 5% failure rate
			LeakDetectionRate:  10.0,              // 10 leaks/minute
			GCPressure:         10.0,              // 10 GCs/minute
			ResourceCount:      10000,             // 10k resources
		},
	}
}

// NewMetricsHistory creates a new metrics history buffer
func NewMetricsHistory(maxSize int) *MetricsHistory {
	return &MetricsHistory{
		entries: make([]*ResourceMetrics, maxSize),
		maxSize: maxSize,
	}
}

// NewAlertManager creates a new alert manager
func NewAlertManager(thresholds AlertThresholds) *AlertManager {
	return &AlertManager{
		thresholds:   thresholds,
		activeAlerts: make(map[string]*Alert),
		alertHistory: make([]*Alert, 0),
		enabled:      true,
	}
}

// ------------------------------------
// 📊 Metrics Collection
// ------------------------------------

// UpdateMetrics collects and updates current resource metrics
func (rm *ResourceMonitor) UpdateMetrics() {
	if !rm.enabled || !rm.collecting {
		return
	}

	start := time.Now()
	metrics := rm.collectMetrics()
	metrics.Timestamp = start
	metrics.CollectionDuration = time.Since(start)

	rm.mutex.Lock()
	rm.metrics = metrics
	rm.mutex.Unlock()

	// Add to history
	rm.history.Add(metrics)

	// Check for alerts
	if rm.config.EnableAlerts {
		rm.alerts.CheckThresholds(metrics)
	}
}

// collectMetrics gathers current resource metrics
func (rm *ResourceMonitor) collectMetrics() *ResourceMetrics {
	metrics := &ResourceMetrics{}

	// Collect memory manager metrics
	if rm.memoryManager != nil {
		memStats := rm.memoryManager.GetMemoryStats()
		metrics.TotalMemoryUsage = memStats.ActiveMemoryUsage
		metrics.ActiveAllocations = memStats.ActiveAllocations

		// Calculate rates (simplified - would need historical data for accuracy)
		metrics.AllocationRate = float64(memStats.TotalAllocations) / time.Since(memStats.StartTime).Seconds()

		// Cleanup metrics
		if memStats.CleanupOperations > 0 {
			metrics.CleanupSuccessRate = float64(memStats.SuccessfulCleanups) / float64(memStats.CleanupOperations) * 100
		}
		metrics.AverageCleanupTime = memStats.AverageCleanupTime

		// Leak detection rate
		if memStats.LeaksDetected > 0 {
			metrics.LeakDetectionRate = float64(memStats.LeaksDetected) / time.Since(memStats.StartTime).Minutes()
		}

		// Collect allocation counts by type
		rm.collectAllocationCounts(metrics)
	}

	// Collect system metrics
	rm.collectSystemMetrics(metrics)

	return metrics
}

// collectAllocationCounts collects counts of different allocation types
func (rm *ResourceMonitor) collectAllocationCounts(metrics *ResourceMetrics) {
	if rm.memoryManager == nil {
		return
	}

	// Get allocations by type
	signalAllocs := rm.memoryManager.GetAllocationsByType(AllocSignal)
	effectAllocs := rm.memoryManager.GetAllocationsByType(AllocEffect)
	memoAllocs := rm.memoryManager.GetAllocationsByType(AllocMemo)
	resourceAllocs := rm.memoryManager.GetAllocationsByType(AllocResource)
	domAllocs := rm.memoryManager.GetAllocationsByType(AllocDOMBinding)
	eventAllocs := rm.memoryManager.GetAllocationsByType(AllocEventSubscription)
	componentAllocs := rm.memoryManager.GetAllocationsByType(AllocComponent)

	metrics.SignalCount = uint64(len(signalAllocs))
	metrics.EffectCount = uint64(len(effectAllocs))
	metrics.MemoCount = uint64(len(memoAllocs))
	metrics.ResourceCount = uint64(len(resourceAllocs))
	metrics.DOMBindingCount = uint64(len(domAllocs))
	metrics.EventSubscriptionCount = uint64(len(eventAllocs))
	metrics.ComponentCount = uint64(len(componentAllocs))
}

// collectSystemMetrics collects system-level metrics
func (rm *ResourceMonitor) collectSystemMetrics(metrics *ResourceMetrics) {
	// Goroutine count
	metrics.GoroutineCount = runtime.NumGoroutine()

	// GC metrics
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	metrics.GCStats = GCMetrics{
		NumGC:        memStats.NumGC,
		PauseTotal:   time.Duration(memStats.PauseTotalNs),
		LastGC:       time.Unix(0, int64(memStats.LastGC)),
		NextGC:       memStats.NextGC,
		HeapAlloc:    memStats.HeapAlloc,
		HeapSys:      memStats.HeapSys,
		HeapIdle:     memStats.HeapIdle,
		HeapInuse:    memStats.HeapInuse,
		HeapReleased: memStats.HeapReleased,
		HeapObjects:  memStats.HeapObjects,
	}

	// Recent pause times (last 10)
	pauseCount := len(memStats.PauseNs)
	if pauseCount > 10 {
		pauseCount = 10
	}
	metrics.GCStats.PauseNs = make([]uint64, pauseCount)
	for i := 0; i < pauseCount; i++ {
		metrics.GCStats.PauseNs[i] = memStats.PauseNs[(memStats.NumGC+255-uint32(i))%256]
	}
}

// ------------------------------------
// 📈 Metrics History Management
// ------------------------------------

// Add adds a metrics entry to the history
func (mh *MetricsHistory) Add(metrics *ResourceMetrics) {
	mh.mutex.Lock()
	defer mh.mutex.Unlock()

	mh.entries[mh.currentIndex] = metrics
	mh.currentIndex = (mh.currentIndex + 1) % mh.maxSize

	if mh.currentIndex == 0 {
		mh.full = true
	}
}

// GetRecent returns the most recent metrics entries
func (mh *MetricsHistory) GetRecent(count int) []*ResourceMetrics {
	mh.mutex.RLock()
	defer mh.mutex.RUnlock()

	if count <= 0 {
		return nil
	}

	size := mh.currentIndex
	if mh.full {
		size = mh.maxSize
	}

	if count > size {
		count = size
	}

	result := make([]*ResourceMetrics, count)
	for i := 0; i < count; i++ {
		index := (mh.currentIndex - 1 - i + mh.maxSize) % mh.maxSize
		result[i] = mh.entries[index]
	}

	return result
}

// GetRange returns metrics entries within a time range
func (mh *MetricsHistory) GetRange(start, end time.Time) []*ResourceMetrics {
	mh.mutex.RLock()
	defer mh.mutex.RUnlock()

	var result []*ResourceMetrics

	size := mh.currentIndex
	if mh.full {
		size = mh.maxSize
	}

	for i := 0; i < size; i++ {
		entry := mh.entries[i]
		if entry != nil && !entry.Timestamp.Before(start) && !entry.Timestamp.After(end) {
			result = append(result, entry)
		}
	}

	return result
}

// ------------------------------------
// 🚨 Alert Management
// ------------------------------------

// CheckThresholds checks metrics against alert thresholds
func (am *AlertManager) CheckThresholds(metrics *ResourceMetrics) {
	if !am.enabled {
		return
	}

	am.mutex.Lock()
	defer am.mutex.Unlock()

	// Check memory usage
	if metrics.TotalMemoryUsage > am.thresholds.MemoryUsage {
		am.triggerAlert(AlertMemoryUsage, AlertWarning,
			"Memory usage exceeded threshold",
			metrics.TotalMemoryUsage, am.thresholds.MemoryUsage)
	} else {
		am.resolveAlert(string(AlertMemoryUsage))
	}

	// Check allocation rate
	if metrics.AllocationRate > am.thresholds.AllocationRate {
		am.triggerAlert(AlertAllocationRate, AlertWarning,
			"Allocation rate exceeded threshold",
			metrics.AllocationRate, am.thresholds.AllocationRate)
	} else {
		am.resolveAlert(string(AlertAllocationRate))
	}

	// Check cleanup failure rate
	if 100-metrics.CleanupSuccessRate > am.thresholds.CleanupFailureRate {
		am.triggerAlert(AlertCleanupFailure, AlertCritical,
			"Cleanup failure rate exceeded threshold",
			100-metrics.CleanupSuccessRate, am.thresholds.CleanupFailureRate)
	} else {
		am.resolveAlert(string(AlertCleanupFailure))
	}

	// Check leak detection rate
	if metrics.LeakDetectionRate > am.thresholds.LeakDetectionRate {
		am.triggerAlert(AlertLeakDetection, AlertCritical,
			"Leak detection rate exceeded threshold",
			metrics.LeakDetectionRate, am.thresholds.LeakDetectionRate)
	} else {
		am.resolveAlert(string(AlertLeakDetection))
	}

	// Check total resource count
	totalResources := metrics.SignalCount + metrics.EffectCount + metrics.MemoCount +
		metrics.ResourceCount + metrics.DOMBindingCount + metrics.EventSubscriptionCount +
		metrics.ComponentCount

	if totalResources > uint64(am.thresholds.ResourceCount) {
		am.triggerAlert(AlertResourceCount, AlertWarning,
			"Total resource count exceeded threshold",
			totalResources, am.thresholds.ResourceCount)
	} else {
		am.resolveAlert(string(AlertResourceCount))
	}
}

// triggerAlert creates or updates an alert
func (am *AlertManager) triggerAlert(alertType AlertType, severity AlertSeverity, message string, value, threshold interface{}) {
	alertID := string(alertType)

	if existing, exists := am.activeAlerts[alertID]; exists {
		// Update existing alert
		existing.Value = value
		existing.Timestamp = time.Now()
		return
	}

	// Create new alert
	alert := &Alert{
		ID:        alertID,
		Type:      alertType,
		Severity:  severity,
		Message:   message,
		Timestamp: time.Now(),
		Value:     value,
		Threshold: threshold,
		Metadata:  make(map[string]interface{}),
	}

	am.activeAlerts[alertID] = alert
	am.alertHistory = append(am.alertHistory, alert)
}

// resolveAlert resolves an active alert
func (am *AlertManager) resolveAlert(alertID string) {
	if alert, exists := am.activeAlerts[alertID]; exists {
		alert.Resolved = true
		alert.ResolvedAt = time.Now()
		delete(am.activeAlerts, alertID)
	}
}

// GetActiveAlerts returns all active alerts
func (am *AlertManager) GetActiveAlerts() []*Alert {
	am.mutex.RLock()
	defer am.mutex.RUnlock()

	var alerts []*Alert
	for _, alert := range am.activeAlerts {
		alerts = append(alerts, alert)
	}

	return alerts
}

// GetAlertHistory returns alert history
func (am *AlertManager) GetAlertHistory() []*Alert {
	am.mutex.RLock()
	defer am.mutex.RUnlock()

	history := make([]*Alert, len(am.alertHistory))
	copy(history, am.alertHistory)
	return history
}

// ------------------------------------
// 🔄 Background Monitoring
// ------------------------------------

// backgroundMonitoring runs continuous resource monitoring
func (rm *ResourceMonitor) backgroundMonitoring() {
	for {
		select {
		case <-rm.ctx.Done():
			return
		case <-rm.stopMonitoring:
			return
		case <-rm.monitorTicker.C:
			rm.UpdateMetrics()
		}
	}
}

// ------------------------------------
// 📊 Reporting and Analysis
// ------------------------------------

// GetCurrentMetrics returns the current metrics
func (rm *ResourceMonitor) GetCurrentMetrics() *ResourceMetrics {
	rm.mutex.RLock()
	defer rm.mutex.RUnlock()

	if rm.metrics == nil {
		return &ResourceMetrics{}
	}

	// Return a copy to prevent concurrent access issues
	metrics := *rm.metrics
	return &metrics
}

// GetReport returns a comprehensive monitoring report
func (rm *ResourceMonitor) GetReport() map[string]interface{} {
	rm.mutex.RLock()
	defer rm.mutex.RUnlock()

	report := map[string]interface{}{
		"monitor_id":      rm.id,
		"enabled":         rm.enabled,
		"collecting":      rm.collecting,
		"config":          rm.config,
		"current_metrics": rm.GetCurrentMetrics(),
		"active_alerts":   rm.alerts.GetActiveAlerts(),
	}

	// Add recent metrics trend
	recent := rm.history.GetRecent(10)
	if len(recent) > 0 {
		report["recent_metrics"] = recent

		// Add trend analysis if enabled
		if rm.config.EnableTrendAnalysis {
			report["trends"] = rm.analyzeTrends(recent)
		}
	}

	return report
}

// analyzeTrends performs basic trend analysis on recent metrics
func (rm *ResourceMonitor) analyzeTrends(recent []*ResourceMetrics) map[string]interface{} {
	if len(recent) < 2 {
		return map[string]interface{}{"status": "insufficient_data"}
	}

	trends := map[string]interface{}{}

	// Memory usage trend
	first := recent[len(recent)-1]
	last := recent[0]

	memoryTrend := float64(last.TotalMemoryUsage-first.TotalMemoryUsage) / float64(first.TotalMemoryUsage) * 100
	trends["memory_usage_trend"] = memoryTrend

	// Allocation trend
	allocationTrend := float64(last.ActiveAllocations) - float64(first.ActiveAllocations)
	trends["allocation_trend"] = allocationTrend

	// Resource count trends
	firstTotal := first.SignalCount + first.EffectCount + first.MemoCount + first.ResourceCount
	lastTotal := last.SignalCount + last.EffectCount + last.MemoCount + last.ResourceCount
	resourceTrend := float64(lastTotal) - float64(firstTotal)
	trends["resource_count_trend"] = resourceTrend

	return trends
}

// GetHistoricalData returns historical metrics data
func (rm *ResourceMonitor) GetHistoricalData(duration time.Duration) []*ResourceMetrics {
	end := time.Now()
	start := end.Add(-duration)
	return rm.history.GetRange(start, end)
}

// ------------------------------------
// 🛠️ Configuration and Control
// ------------------------------------

// SetEnabled enables or disables resource monitoring
func (rm *ResourceMonitor) SetEnabled(enabled bool) {
	rm.mutex.Lock()
	defer rm.mutex.Unlock()
	rm.enabled = enabled
}

// SetCollecting enables or disables metrics collection
func (rm *ResourceMonitor) SetCollecting(collecting bool) {
	rm.mutex.Lock()
	defer rm.mutex.Unlock()
	rm.collecting = collecting
}

// UpdateConfig updates the monitor configuration
func (rm *ResourceMonitor) UpdateConfig(config MonitorConfig) {
	rm.mutex.Lock()
	defer rm.mutex.Unlock()

	rm.config = config

	// Update alert thresholds
	rm.alerts.mutex.Lock()
	rm.alerts.thresholds = config.AlertThresholds
	rm.alerts.enabled = config.EnableAlerts
	rm.alerts.mutex.Unlock()

	// Update ticker interval if changed
	if rm.monitorTicker != nil {
		rm.monitorTicker.Stop()
		rm.monitorTicker = time.NewTicker(config.CollectionInterval)
	}
}

// ------------------------------------
// 🧹 Disposal
// ------------------------------------

// Dispose cleans up the resource monitor and all its resources
func (rm *ResourceMonitor) Dispose() {
	rm.mutex.Lock()
	defer rm.mutex.Unlock()

	// Stop background monitoring
	if rm.cancel != nil {
		rm.cancel()
	}

	if rm.monitorTicker != nil {
		rm.monitorTicker.Stop()
	}

	select {
	case rm.stopMonitoring <- true:
	default:
	}

	// Clear data
	rm.metrics = nil
	rm.history = nil
	rm.alerts = nil

	rm.enabled = false
	rm.collecting = false
}
