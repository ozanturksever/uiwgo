// memory_debug.go
// Debugging utilities and development tools for memory management

package golid

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"
)

// ------------------------------------
// 🔍 Memory Debugger
// ------------------------------------

// MemoryDebugger provides comprehensive debugging utilities for memory management
type MemoryDebugger struct {
	memoryManager *MemoryManager
	enabled       bool
	logLevel      DebugLogLevel
	logBuffer     []DebugLogEntry
	bufferMutex   sync.RWMutex
	maxLogEntries int
}

// DebugLogLevel represents the level of debug logging
type DebugLogLevel int

const (
	DebugLogOff DebugLogLevel = iota
	DebugLogError
	DebugLogWarn
	DebugLogInfo
	DebugLogDebug
	DebugLogTrace
)

// String returns the string representation of the debug log level
func (level DebugLogLevel) String() string {
	switch level {
	case DebugLogOff:
		return "OFF"
	case DebugLogError:
		return "ERROR"
	case DebugLogWarn:
		return "WARN"
	case DebugLogInfo:
		return "INFO"
	case DebugLogDebug:
		return "DEBUG"
	case DebugLogTrace:
		return "TRACE"
	default:
		return "UNKNOWN"
	}
}

// DebugLogEntry represents a single debug log entry
type DebugLogEntry struct {
	Timestamp    time.Time              `json:"timestamp"`
	Level        DebugLogLevel          `json:"level"`
	Category     string                 `json:"category"`
	Message      string                 `json:"message"`
	Data         map[string]interface{} `json:"data,omitempty"`
	StackTrace   string                 `json:"stack_trace,omitempty"`
	AllocationID uint64                 `json:"allocation_id,omitempty"`
}

// NewMemoryDebugger creates a new memory debugger
func NewMemoryDebugger(memoryManager *MemoryManager) *MemoryDebugger {
	return &MemoryDebugger{
		memoryManager: memoryManager,
		enabled:       false,
		logLevel:      DebugLogInfo,
		logBuffer:     make([]DebugLogEntry, 0),
		maxLogEntries: 10000,
	}
}

// SetEnabled enables or disables the memory debugger
func (md *MemoryDebugger) SetEnabled(enabled bool) {
	md.bufferMutex.Lock()
	defer md.bufferMutex.Unlock()
	md.enabled = enabled

	if enabled {
		md.Log(DebugLogInfo, "debugger", "Memory debugger enabled", nil)
	}
}

// SetLogLevel sets the debug log level
func (md *MemoryDebugger) SetLogLevel(level DebugLogLevel) {
	md.bufferMutex.Lock()
	defer md.bufferMutex.Unlock()
	md.logLevel = level

	md.Log(DebugLogInfo, "debugger", "Log level changed", map[string]interface{}{
		"new_level": level.String(),
	})
}

// Log adds a debug log entry
func (md *MemoryDebugger) Log(level DebugLogLevel, category, message string, data map[string]interface{}) {
	if !md.enabled || level > md.logLevel {
		return
	}

	md.bufferMutex.Lock()
	defer md.bufferMutex.Unlock()

	entry := DebugLogEntry{
		Timestamp: time.Now(),
		Level:     level,
		Category:  category,
		Message:   message,
		Data:      data,
	}

	// Add stack trace for errors and warnings
	if level <= DebugLogWarn {
		entry.StackTrace = captureStackTrace(3) // Skip this function and callers
	}

	md.logBuffer = append(md.logBuffer, entry)

	// Trim buffer if it exceeds max entries
	if len(md.logBuffer) > md.maxLogEntries {
		md.logBuffer = md.logBuffer[len(md.logBuffer)-md.maxLogEntries:]
	}
}

// LogAllocation logs allocation-related events
func (md *MemoryDebugger) LogAllocation(level DebugLogLevel, message string, allocation *Allocation) {
	if allocation == nil {
		return
	}

	data := map[string]interface{}{
		"allocation_id":   allocation.ID,
		"allocation_type": allocation.Type.String(),
		"size":            allocation.Size,
		"owner":           allocation.Owner,
		"disposed":        allocation.Disposed,
	}

	if allocation.Metadata != nil {
		data["metadata"] = allocation.Metadata
	}

	entry := DebugLogEntry{
		Timestamp:    time.Now(),
		Level:        level,
		Category:     "allocation",
		Message:      message,
		Data:         data,
		AllocationID: allocation.ID,
	}

	if level <= DebugLogWarn {
		entry.StackTrace = captureStackTrace(3)
	}

	md.bufferMutex.Lock()
	defer md.bufferMutex.Unlock()

	md.logBuffer = append(md.logBuffer, entry)

	if len(md.logBuffer) > md.maxLogEntries {
		md.logBuffer = md.logBuffer[len(md.logBuffer)-md.maxLogEntries:]
	}
}

// GetLogs returns debug log entries with optional filtering
func (md *MemoryDebugger) GetLogs(filter DebugLogFilter) []DebugLogEntry {
	md.bufferMutex.RLock()
	defer md.bufferMutex.RUnlock()

	logs := make([]DebugLogEntry, 0)

	for _, entry := range md.logBuffer {
		if md.matchesFilter(entry, filter) {
			logs = append(logs, entry)
		}
	}

	// Sort by timestamp (newest first)
	sort.Slice(logs, func(i, j int) bool {
		return logs[i].Timestamp.After(logs[j].Timestamp)
	})

	// Apply limit
	if filter.Limit > 0 && len(logs) > filter.Limit {
		logs = logs[:filter.Limit]
	}

	return logs
}

// DebugLogFilter provides filtering options for debug logs
type DebugLogFilter struct {
	Level        *DebugLogLevel `json:"level,omitempty"`
	Category     string         `json:"category,omitempty"`
	AllocationID *uint64        `json:"allocation_id,omitempty"`
	Since        *time.Time     `json:"since,omitempty"`
	Until        *time.Time     `json:"until,omitempty"`
	Limit        int            `json:"limit,omitempty"`
}

// matchesFilter checks if a log entry matches the filter criteria
func (md *MemoryDebugger) matchesFilter(entry DebugLogEntry, filter DebugLogFilter) bool {
	if filter.Level != nil && entry.Level > *filter.Level {
		return false
	}

	if filter.Category != "" && entry.Category != filter.Category {
		return false
	}

	if filter.AllocationID != nil && entry.AllocationID != *filter.AllocationID {
		return false
	}

	if filter.Since != nil && entry.Timestamp.Before(*filter.Since) {
		return false
	}

	if filter.Until != nil && entry.Timestamp.After(*filter.Until) {
		return false
	}

	return true
}

// ClearLogs clears the debug log buffer
func (md *MemoryDebugger) ClearLogs() {
	md.bufferMutex.Lock()
	defer md.bufferMutex.Unlock()

	md.logBuffer = make([]DebugLogEntry, 0)
	md.Log(DebugLogInfo, "debugger", "Debug logs cleared", nil)
}

// ------------------------------------
// 📊 Memory Visualizer
// ------------------------------------

// MemoryVisualizer provides visualization utilities for memory data
type MemoryVisualizer struct {
	debugger *MemoryDebugger
}

// NewMemoryVisualizer creates a new memory visualizer
func NewMemoryVisualizer(debugger *MemoryDebugger) *MemoryVisualizer {
	return &MemoryVisualizer{
		debugger: debugger,
	}
}

// GenerateAllocationChart generates a text-based chart of allocations by type
func (mv *MemoryVisualizer) GenerateAllocationChart() string {
	if mv.debugger.memoryManager == nil {
		return "Memory manager not available"
	}

	// Get allocation counts by type
	typeCounts := make(map[AllocationType]int)

	for allocType := AllocSignal; allocType <= AllocCustom; allocType++ {
		allocs := mv.debugger.memoryManager.GetAllocationsByType(allocType)
		typeCounts[allocType] = len(allocs)
	}

	// Find max count for scaling
	maxCount := 0
	for _, count := range typeCounts {
		if count > maxCount {
			maxCount = count
		}
	}

	if maxCount == 0 {
		return "No allocations found"
	}

	// Generate chart
	var chart strings.Builder
	chart.WriteString("Allocation Distribution:\n")
	chart.WriteString("========================\n\n")

	for allocType := AllocSignal; allocType <= AllocCustom; allocType++ {
		count := typeCounts[allocType]
		percentage := float64(count) / float64(maxCount) * 100
		barLength := int(percentage / 5) // Scale to 20 chars max

		chart.WriteString(fmt.Sprintf("%-15s [%3d] ", allocType.String(), count))

		// Draw bar
		for i := 0; i < barLength; i++ {
			chart.WriteString("█")
		}

		chart.WriteString(fmt.Sprintf(" %.1f%%\n", percentage))
	}

	return chart.String()
}

// GenerateMemoryTimeline generates a timeline of memory events
func (mv *MemoryVisualizer) GenerateMemoryTimeline(duration time.Duration) string {
	since := time.Now().Add(-duration)
	filter := DebugLogFilter{
		Since: &since,
		Limit: 100,
	}

	logs := mv.debugger.GetLogs(filter)

	if len(logs) == 0 {
		return "No memory events in the specified duration"
	}

	var timeline strings.Builder
	timeline.WriteString(fmt.Sprintf("Memory Timeline (Last %v):\n", duration))
	timeline.WriteString("================================\n\n")

	for _, entry := range logs {
		timestamp := entry.Timestamp.Format("15:04:05.000")
		timeline.WriteString(fmt.Sprintf("[%s] %s/%s: %s\n",
			timestamp,
			entry.Level.String(),
			entry.Category,
			entry.Message,
		))

		if entry.AllocationID > 0 {
			timeline.WriteString(fmt.Sprintf("           Allocation ID: %d\n", entry.AllocationID))
		}
	}

	return timeline.String()
}

// GenerateLeakReport generates a detailed leak report
func (mv *MemoryVisualizer) GenerateLeakReport() string {
	if mv.debugger.memoryManager == nil || mv.debugger.memoryManager.leakDetector == nil {
		return "Leak detector not available"
	}

	violations := mv.debugger.memoryManager.leakDetector.GetViolations()
	suspicious := mv.debugger.memoryManager.leakDetector.GetSuspiciousAllocations()

	var report strings.Builder
	report.WriteString("Memory Leak Report:\n")
	report.WriteString("===================\n\n")

	// Violations
	report.WriteString(fmt.Sprintf("Confirmed Leaks: %d\n", len(violations)))
	if len(violations) > 0 {
		report.WriteString("------------------------\n")
		for i, violation := range violations {
			if i >= 10 { // Limit to first 10
				report.WriteString(fmt.Sprintf("... and %d more\n", len(violations)-10))
				break
			}

			report.WriteString(fmt.Sprintf("Leak #%d:\n", i+1))
			report.WriteString(fmt.Sprintf("  Type: %s\n", violation.ResourceType))
			report.WriteString(fmt.Sprintf("  Severity: %s\n", violation.Severity))
			report.WriteString(fmt.Sprintf("  Description: %s\n", violation.Description))
			report.WriteString(fmt.Sprintf("  Count: %d\n", violation.Count))
			report.WriteString(fmt.Sprintf("  Threshold: %d\n", violation.Threshold))
			report.WriteString("\n")
		}
	}

	// Suspicious allocations
	report.WriteString(fmt.Sprintf("\nSuspicious Allocations: %d\n", len(suspicious)))
	if len(suspicious) > 0 {
		report.WriteString("---------------------------\n")
		for i, suspiciousAlloc := range suspicious {
			if i >= 10 { // Limit to first 10
				report.WriteString(fmt.Sprintf("... and %d more\n", len(suspicious)-10))
				break
			}

			alloc := suspiciousAlloc.Allocation
			if alloc != nil {
				age := time.Since(alloc.Timestamp)
				report.WriteString(fmt.Sprintf("Allocation #%d:\n", alloc.ID))
				report.WriteString(fmt.Sprintf("  Type: %s\n", alloc.Type.String()))
				report.WriteString(fmt.Sprintf("  Size: %d bytes\n", alloc.Size))
				report.WriteString(fmt.Sprintf("  Age: %v\n", age))
				report.WriteString(fmt.Sprintf("  Suspicion Level: %s\n", suspiciousAlloc.SuspicionLevel.String()))
				report.WriteString("\n")
			}
		}
	}

	return report.String()
}

// ------------------------------------
// 🛠️ Memory Inspector
// ------------------------------------

// MemoryInspector provides detailed inspection utilities
type MemoryInspector struct {
	debugger *MemoryDebugger
}

// NewMemoryInspector creates a new memory inspector
func NewMemoryInspector(debugger *MemoryDebugger) *MemoryInspector {
	return &MemoryInspector{
		debugger: debugger,
	}
}

// InspectAllocation provides detailed information about a specific allocation
func (mi *MemoryInspector) InspectAllocation(allocationID uint64) map[string]interface{} {
	if mi.debugger.memoryManager == nil {
		return map[string]interface{}{
			"error": "Memory manager not available",
		}
	}

	alloc, exists := mi.debugger.memoryManager.GetAllocation(allocationID)
	if !exists || alloc == nil {
		return map[string]interface{}{
			"error": "Allocation not found",
		}
	}

	age := time.Since(alloc.Timestamp)

	inspection := map[string]interface{}{
		"id":         alloc.ID,
		"type":       alloc.Type.String(),
		"size":       alloc.Size,
		"owner":      alloc.Owner,
		"created_at": alloc.Timestamp,
		"age":        age.String(),
		"disposed":   alloc.Disposed,
		"metadata":   alloc.Metadata,
	}

	if alloc.StackTrace != nil && len(alloc.StackTrace) > 0 {
		inspection["stack_trace"] = alloc.StackTrace
	}

	// Get related log entries
	filter := DebugLogFilter{
		AllocationID: &allocationID,
		Limit:        50,
	}
	logs := mi.debugger.GetLogs(filter)
	inspection["log_entries"] = len(logs)

	// Check if it's suspicious
	if mi.debugger.memoryManager.leakDetector != nil {
		suspicious := mi.debugger.memoryManager.leakDetector.GetSuspiciousAllocations()
		for _, suspiciousAlloc := range suspicious {
			if suspiciousAlloc.Allocation != nil && suspiciousAlloc.Allocation.ID == allocationID {
				inspection["suspicious"] = true
				inspection["suspicion_level"] = suspiciousAlloc.SuspicionLevel.String()
				inspection["suspicion_reasons"] = suspiciousAlloc.Reasons
				break
			}
		}
	}

	return inspection
}

// InspectOwner provides detailed information about allocations for a specific owner
func (mi *MemoryInspector) InspectOwner(owner *Owner) map[string]interface{} {
	if mi.debugger.memoryManager == nil {
		return map[string]interface{}{
			"error": "Memory manager not available",
		}
	}

	if owner == nil {
		return map[string]interface{}{
			"error": "Owner is nil",
		}
	}

	allocs := mi.debugger.memoryManager.GetAllocationsByOwner(owner)

	inspection := map[string]interface{}{
		"owner_id":             owner.id,
		"total_allocations":    len(allocs),
		"active_allocations":   0,
		"disposed_allocations": 0,
		"total_size":           int64(0),
		"active_size":          int64(0),
		"allocation_types":     make(map[string]int),
	}

	typeCounts := make(map[string]int)

	for _, alloc := range allocs {
		inspection["total_size"] = inspection["total_size"].(int64) + alloc.Size

		if alloc.Disposed {
			inspection["disposed_allocations"] = inspection["disposed_allocations"].(int) + 1
		} else {
			inspection["active_allocations"] = inspection["active_allocations"].(int) + 1
			inspection["active_size"] = inspection["active_size"].(int64) + alloc.Size
		}

		typeStr := alloc.Type.String()
		typeCounts[typeStr]++
	}

	inspection["allocation_types"] = typeCounts

	return inspection
}

// ------------------------------------
// 📈 Performance Analyzer
// ------------------------------------

// PerformanceAnalyzer analyzes memory management performance
type PerformanceAnalyzer struct {
	debugger *MemoryDebugger
}

// NewPerformanceAnalyzer creates a new performance analyzer
func NewPerformanceAnalyzer(debugger *MemoryDebugger) *PerformanceAnalyzer {
	return &PerformanceAnalyzer{
		debugger: debugger,
	}
}

// AnalyzePerformance generates a performance analysis report
func (pa *PerformanceAnalyzer) AnalyzePerformance() map[string]interface{} {
	if pa.debugger.memoryManager == nil {
		return map[string]interface{}{
			"error": "Memory manager not available",
		}
	}

	stats := pa.debugger.memoryManager.GetMemoryStats()

	analysis := map[string]interface{}{
		"timestamp":    time.Now(),
		"memory_stats": stats,
	}

	// Calculate performance metrics
	if stats.TotalAllocations > 0 {
		leakRate := float64(stats.LeaksDetected) / float64(stats.TotalAllocations) * 100
		analysis["leak_rate_percent"] = leakRate

		if stats.CleanupOperations > 0 {
			cleanupSuccessRate := float64(stats.SuccessfulCleanups) / float64(stats.CleanupOperations) * 100
			analysis["cleanup_success_rate_percent"] = cleanupSuccessRate
		}

		avgAllocationSize := float64(stats.TotalMemoryAllocated) / float64(stats.TotalAllocations)
		analysis["average_allocation_size"] = avgAllocationSize
	}

	// Performance recommendations
	recommendations := make([]string, 0)

	if stats.LeaksDetected > 0 {
		recommendations = append(recommendations, "Memory leaks detected - review cleanup patterns")
	}

	if stats.CleanupOperations > 0 {
		successRate := float64(stats.SuccessfulCleanups) / float64(stats.CleanupOperations) * 100
		if successRate < 95 {
			recommendations = append(recommendations, "Low cleanup success rate - review cleanup implementations")
		}
	}

	if stats.ActiveMemoryUsage > 100*1024*1024 { // 100MB
		recommendations = append(recommendations, "High memory usage - consider optimization")
	}

	analysis["recommendations"] = recommendations

	return analysis
}

// ------------------------------------
// 🔧 Debug Utilities
// ------------------------------------

// captureStackTrace captures the current stack trace
func captureStackTrace(skip int) string {
	// This is a simplified implementation
	// In a real implementation, you would use runtime.Stack() or similar
	return fmt.Sprintf("Stack trace captured (skip: %d)", skip)
}

// FormatJSON formats data as pretty JSON
func FormatJSON(data interface{}) string {
	bytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Sprintf("Error formatting JSON: %v", err)
	}
	return string(bytes)
}

var (
	globalMemoryDebugger *MemoryDebugger
	debuggerOnce         sync.Once
)

// GetGlobalMemoryDebugger returns the global memory debugger instance
func GetGlobalMemoryDebugger() *MemoryDebugger {
	debuggerOnce.Do(func() {
		memoryManager := GetGlobalMemoryManager()
		globalMemoryDebugger = NewMemoryDebugger(memoryManager)
	})
	return globalMemoryDebugger
}

// EnableGlobalMemoryDebugging enables global memory debugging
func EnableGlobalMemoryDebugging() {
	debugger := GetGlobalMemoryDebugger()
	debugger.SetEnabled(true)
	debugger.SetLogLevel(DebugLogDebug)
}

// DisableGlobalMemoryDebugging disables global memory debugging
func DisableGlobalMemoryDebugging() {
	debugger := GetGlobalMemoryDebugger()
	debugger.SetEnabled(false)
}

// GetMemoryDebugReport generates a comprehensive debug report
func GetMemoryDebugReport() string {
	debugger := GetGlobalMemoryDebugger()
	visualizer := NewMemoryVisualizer(debugger)

	var report strings.Builder

	report.WriteString("=== GOLID MEMORY DEBUG REPORT ===\n")
	report.WriteString(fmt.Sprintf("Generated at: %s\n\n", time.Now().Format(time.RFC3339)))

	// Allocation chart
	report.WriteString(visualizer.GenerateAllocationChart())
	report.WriteString("\n")

	// Memory timeline
	report.WriteString(visualizer.GenerateMemoryTimeline(time.Hour))
	report.WriteString("\n")

	// Leak report
	report.WriteString(visualizer.GenerateLeakReport())
	report.WriteString("\n")

	// Performance analysis
	analyzer := NewPerformanceAnalyzer(debugger)
	perfAnalysis := analyzer.AnalyzePerformance()
	report.WriteString("Performance Analysis:\n")
	report.WriteString("====================\n")
	report.WriteString(FormatJSON(perfAnalysis))
	report.WriteString("\n")

	return report.String()
}
