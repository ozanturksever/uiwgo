// memory_profiler.go
// Development tools for memory usage analysis and profiling

package golid

import (
	"encoding/json"
	"fmt"
	"runtime"
	"sort"
	"sync"
	"time"
)

// ------------------------------------
// 🔬 Memory Profiler Core
// ------------------------------------

// MemoryProfiler provides development tools for memory analysis
type MemoryProfiler struct {
	memoryManager *MemoryManager
	enabled       bool
	profiling     bool

	// Profiling data
	snapshots []*MemorySnapshot
	profiles  []*AllocationProfile

	// Configuration
	config ProfilerConfig

	// Synchronization
	mutex sync.RWMutex
}

// ProfilerConfig provides configuration for memory profiling
type ProfilerConfig struct {
	MaxSnapshots      int           `json:"max_snapshots"`
	SnapshotInterval  time.Duration `json:"snapshot_interval"`
	EnableStackTraces bool          `json:"enable_stack_traces"`
	EnableTrends      bool          `json:"enable_trends"`
	ProfileDepth      int           `json:"profile_depth"`
	ReportFormat      string        `json:"report_format"`
}

// MemorySnapshot represents a point-in-time memory state
type MemorySnapshot struct {
	ID                uint64                    `json:"id"`
	Timestamp         time.Time                 `json:"timestamp"`
	TotalAllocations  uint64                    `json:"total_allocations"`
	ActiveAllocations uint64                    `json:"active_allocations"`
	TotalMemory       int64                     `json:"total_memory"`
	ActiveMemory      int64                     `json:"active_memory"`
	AllocationsByType map[AllocationType]uint64 `json:"allocations_by_type"`
	MemoryByType      map[AllocationType]int64  `json:"memory_by_type"`
	RuntimeStats      RuntimeMemoryStats        `json:"runtime_stats"`
	LeakViolations    int                       `json:"leak_violations"`
	CleanupStats      CleanupSnapshotStats      `json:"cleanup_stats"`
}

// RuntimeMemoryStats contains Go runtime memory statistics
type RuntimeMemoryStats struct {
	Alloc         uint64   `json:"alloc"`
	TotalAlloc    uint64   `json:"total_alloc"`
	Sys           uint64   `json:"sys"`
	Lookups       uint64   `json:"lookups"`
	Mallocs       uint64   `json:"mallocs"`
	Frees         uint64   `json:"frees"`
	HeapAlloc     uint64   `json:"heap_alloc"`
	HeapSys       uint64   `json:"heap_sys"`
	HeapIdle      uint64   `json:"heap_idle"`
	HeapInuse     uint64   `json:"heap_inuse"`
	HeapReleased  uint64   `json:"heap_released"`
	HeapObjects   uint64   `json:"heap_objects"`
	StackInuse    uint64   `json:"stack_inuse"`
	StackSys      uint64   `json:"stack_sys"`
	MSpanInuse    uint64   `json:"mspan_inuse"`
	MSpanSys      uint64   `json:"mspan_sys"`
	MCacheInuse   uint64   `json:"mcache_inuse"`
	MCacheSys     uint64   `json:"mcache_sys"`
	BuckHashSys   uint64   `json:"buck_hash_sys"`
	GCSys         uint64   `json:"gc_sys"`
	OtherSys      uint64   `json:"other_sys"`
	NextGC        uint64   `json:"next_gc"`
	LastGC        uint64   `json:"last_gc"`
	PauseTotalNs  uint64   `json:"pause_total_ns"`
	PauseNs       []uint64 `json:"pause_ns"`
	PauseEnd      []uint64 `json:"pause_end"`
	NumGC         uint32   `json:"num_gc"`
	NumForcedGC   uint32   `json:"num_forced_gc"`
	GCCPUFraction float64  `json:"gc_cpu_fraction"`
}

// CleanupSnapshotStats contains cleanup statistics for a snapshot
type CleanupSnapshotStats struct {
	TotalOperations    uint64 `json:"total_operations"`
	SuccessfulCleanups uint64 `json:"successful_cleanups"`
	FailedCleanups     uint64 `json:"failed_cleanups"`
	PendingCleanups    uint64 `json:"pending_cleanups"`
}

// AllocationProfile represents an allocation profiling session
type AllocationProfile struct {
	ID          uint64                `json:"id"`
	Name        string                `json:"name"`
	StartTime   time.Time             `json:"start_time"`
	EndTime     time.Time             `json:"end_time"`
	Duration    time.Duration         `json:"duration"`
	Allocations []*ProfiledAllocation `json:"allocations"`
	Summary     AllocationSummary     `json:"summary"`
	HotSpots    []*AllocationHotSpot  `json:"hot_spots"`
	Trends      *AllocationTrends     `json:"trends,omitempty"`
}

// ProfiledAllocation represents a single allocation in a profile
type ProfiledAllocation struct {
	ID         uint64                 `json:"id"`
	Type       AllocationType         `json:"type"`
	Size       int64                  `json:"size"`
	Timestamp  time.Time              `json:"timestamp"`
	StackTrace []string               `json:"stack_trace,omitempty"`
	Metadata   map[string]interface{} `json:"metadata"`
	LifetimeMs int64                  `json:"lifetime_ms"`
	Disposed   bool                   `json:"disposed"`
}

// AllocationSummary provides summary statistics for a profile
type AllocationSummary struct {
	TotalAllocations  uint64                    `json:"total_allocations"`
	TotalMemory       int64                     `json:"total_memory"`
	AverageSize       float64                   `json:"average_size"`
	MedianSize        int64                     `json:"median_size"`
	MaxSize           int64                     `json:"max_size"`
	MinSize           int64                     `json:"min_size"`
	AllocationsByType map[AllocationType]uint64 `json:"allocations_by_type"`
	MemoryByType      map[AllocationType]int64  `json:"memory_by_type"`
	AverageLifetime   time.Duration             `json:"average_lifetime"`
	LeakCandidates    uint64                    `json:"leak_candidates"`
}

// AllocationHotSpot represents a memory allocation hot spot
type AllocationHotSpot struct {
	Location    string          `json:"location"`
	Function    string          `json:"function"`
	File        string          `json:"file"`
	Line        int             `json:"line"`
	Count       uint64          `json:"count"`
	TotalMemory int64           `json:"total_memory"`
	AverageSize float64         `json:"average_size"`
	Type        AllocationType  `json:"type"`
	Severity    HotSpotSeverity `json:"severity"`
}

// HotSpotSeverity defines severity levels for allocation hot spots
type HotSpotSeverity int

const (
	HotSpotLow HotSpotSeverity = iota
	HotSpotMedium
	HotSpotHigh
	HotSpotCritical
)

// String returns string representation of hot spot severity
func (s HotSpotSeverity) String() string {
	switch s {
	case HotSpotLow:
		return "Low"
	case HotSpotMedium:
		return "Medium"
	case HotSpotHigh:
		return "High"
	case HotSpotCritical:
		return "Critical"
	default:
		return "Unknown"
	}
}

// AllocationTrends provides trend analysis for allocations
type AllocationTrends struct {
	AllocationRate    float64       `json:"allocation_rate"`    // Allocations per second
	MemoryGrowthRate  float64       `json:"memory_growth_rate"` // Bytes per second
	LeakTrend         string        `json:"leak_trend"`         // "increasing", "stable", "decreasing"
	PredictedPeakTime time.Duration `json:"predicted_peak_time"`
	Confidence        float64       `json:"confidence"`
}

// ------------------------------------
// 🏗️ Memory Profiler Creation
// ------------------------------------

// NewMemoryProfiler creates a new memory profiler
func NewMemoryProfiler(memoryManager *MemoryManager) *MemoryProfiler {
	return &MemoryProfiler{
		memoryManager: memoryManager,
		enabled:       true,
		profiling:     false,
		snapshots:     make([]*MemorySnapshot, 0),
		profiles:      make([]*AllocationProfile, 0),
		config:        DefaultProfilerConfig(),
	}
}

// DefaultProfilerConfig returns default profiler configuration
func DefaultProfilerConfig() ProfilerConfig {
	return ProfilerConfig{
		MaxSnapshots:      100,
		SnapshotInterval:  10 * time.Second,
		EnableStackTraces: true,
		EnableTrends:      true,
		ProfileDepth:      10,
		ReportFormat:      "json",
	}
}

// ------------------------------------
// 📸 Memory Snapshots
// ------------------------------------

// TakeSnapshot captures a memory snapshot
func (mp *MemoryProfiler) TakeSnapshot() *MemorySnapshot {
	if !mp.enabled {
		return nil
	}

	mp.mutex.Lock()
	defer mp.mutex.Unlock()

	snapshot := &MemorySnapshot{
		ID:        uint64(len(mp.snapshots) + 1),
		Timestamp: time.Now(),
	}

	// Collect memory manager statistics
	if mp.memoryManager != nil {
		stats := mp.memoryManager.GetMemoryStats()
		snapshot.TotalAllocations = stats.TotalAllocations
		snapshot.ActiveAllocations = stats.ActiveAllocations
		snapshot.TotalMemory = stats.TotalMemoryAllocated
		snapshot.ActiveMemory = stats.ActiveMemoryUsage
		snapshot.LeakViolations = int(stats.LeaksDetected)

		snapshot.CleanupStats = CleanupSnapshotStats{
			TotalOperations:    stats.CleanupOperations,
			SuccessfulCleanups: stats.SuccessfulCleanups,
			FailedCleanups:     stats.FailedCleanups,
		}

		// Collect allocations by type
		snapshot.AllocationsByType = make(map[AllocationType]uint64)
		snapshot.MemoryByType = make(map[AllocationType]int64)

		for allocType := AllocSignal; allocType <= AllocCustom; allocType++ {
			allocs := mp.memoryManager.GetAllocationsByType(allocType)
			snapshot.AllocationsByType[allocType] = uint64(len(allocs))

			var totalMemory int64
			for _, alloc := range allocs {
				totalMemory += alloc.Size
			}
			snapshot.MemoryByType[allocType] = totalMemory
		}
	}

	// Collect runtime statistics
	snapshot.RuntimeStats = mp.collectRuntimeStats()

	// Add to snapshots list
	mp.snapshots = append(mp.snapshots, snapshot)

	// Limit snapshots to max count
	if len(mp.snapshots) > mp.config.MaxSnapshots {
		mp.snapshots = mp.snapshots[1:]
	}

	return snapshot
}

// collectRuntimeStats collects Go runtime memory statistics
func (mp *MemoryProfiler) collectRuntimeStats() RuntimeMemoryStats {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	stats := RuntimeMemoryStats{
		Alloc:         m.Alloc,
		TotalAlloc:    m.TotalAlloc,
		Sys:           m.Sys,
		Lookups:       m.Lookups,
		Mallocs:       m.Mallocs,
		Frees:         m.Frees,
		HeapAlloc:     m.HeapAlloc,
		HeapSys:       m.HeapSys,
		HeapIdle:      m.HeapIdle,
		HeapInuse:     m.HeapInuse,
		HeapReleased:  m.HeapReleased,
		HeapObjects:   m.HeapObjects,
		StackInuse:    m.StackInuse,
		StackSys:      m.StackSys,
		MSpanInuse:    m.MSpanInuse,
		MSpanSys:      m.MSpanSys,
		MCacheInuse:   m.MCacheInuse,
		MCacheSys:     m.MCacheSys,
		BuckHashSys:   m.BuckHashSys,
		GCSys:         m.GCSys,
		OtherSys:      m.OtherSys,
		NextGC:        m.NextGC,
		LastGC:        m.LastGC,
		PauseTotalNs:  m.PauseTotalNs,
		NumGC:         m.NumGC,
		NumForcedGC:   m.NumForcedGC,
		GCCPUFraction: m.GCCPUFraction,
	}

	// Copy recent pause times
	pauseCount := len(m.PauseNs)
	if pauseCount > 10 {
		pauseCount = 10
	}
	stats.PauseNs = make([]uint64, pauseCount)
	stats.PauseEnd = make([]uint64, pauseCount)

	for i := 0; i < pauseCount; i++ {
		stats.PauseNs[i] = m.PauseNs[(m.NumGC+255-uint32(i))%256]
		stats.PauseEnd[i] = m.PauseEnd[(m.NumGC+255-uint32(i))%256]
	}

	return stats
}

// GetSnapshots returns all memory snapshots
func (mp *MemoryProfiler) GetSnapshots() []*MemorySnapshot {
	mp.mutex.RLock()
	defer mp.mutex.RUnlock()

	snapshots := make([]*MemorySnapshot, len(mp.snapshots))
	copy(snapshots, mp.snapshots)
	return snapshots
}

// GetSnapshotRange returns snapshots within a time range
func (mp *MemoryProfiler) GetSnapshotRange(start, end time.Time) []*MemorySnapshot {
	mp.mutex.RLock()
	defer mp.mutex.RUnlock()

	var result []*MemorySnapshot
	for _, snapshot := range mp.snapshots {
		if !snapshot.Timestamp.Before(start) && !snapshot.Timestamp.After(end) {
			result = append(result, snapshot)
		}
	}

	return result
}

// ------------------------------------
// 📊 Allocation Profiling
// ------------------------------------

// StartProfiling begins an allocation profiling session
func (mp *MemoryProfiler) StartProfiling(name string) *AllocationProfile {
	if !mp.enabled {
		return nil
	}

	mp.mutex.Lock()
	defer mp.mutex.Unlock()

	profile := &AllocationProfile{
		ID:          uint64(len(mp.profiles) + 1),
		Name:        name,
		StartTime:   time.Now(),
		Allocations: make([]*ProfiledAllocation, 0),
	}

	mp.profiles = append(mp.profiles, profile)
	mp.profiling = true

	return profile
}

// StopProfiling ends the current profiling session
func (mp *MemoryProfiler) StopProfiling() *AllocationProfile {
	mp.mutex.Lock()
	defer mp.mutex.Unlock()

	if !mp.profiling || len(mp.profiles) == 0 {
		return nil
	}

	profile := mp.profiles[len(mp.profiles)-1]
	profile.EndTime = time.Now()
	profile.Duration = profile.EndTime.Sub(profile.StartTime)

	// Collect current allocations
	if mp.memoryManager != nil {
		mp.collectProfileAllocations(profile)
	}

	// Generate summary and analysis
	mp.generateProfileSummary(profile)
	mp.findHotSpots(profile)

	if mp.config.EnableTrends {
		mp.analyzeTrends(profile)
	}

	mp.profiling = false
	return profile
}

// collectProfileAllocations collects allocations for profiling
func (mp *MemoryProfiler) collectProfileAllocations(profile *AllocationProfile) {
	for allocType := AllocSignal; allocType <= AllocCustom; allocType++ {
		allocs := mp.memoryManager.GetAllocationsByType(allocType)

		for _, alloc := range allocs {
			// Only include allocations created during profiling
			if alloc.Timestamp.After(profile.StartTime) {
				profiledAlloc := &ProfiledAllocation{
					ID:        alloc.ID,
					Type:      alloc.Type,
					Size:      alloc.Size,
					Timestamp: alloc.Timestamp,
					Metadata:  alloc.Metadata,
					Disposed:  alloc.Disposed,
				}

				// Calculate lifetime
				if alloc.Disposed {
					profiledAlloc.LifetimeMs = profile.EndTime.Sub(alloc.Timestamp).Milliseconds()
				} else {
					profiledAlloc.LifetimeMs = profile.EndTime.Sub(alloc.Timestamp).Milliseconds()
				}

				// Add stack trace if enabled and available
				if mp.config.EnableStackTraces && len(alloc.StackTrace) > 0 {
					profiledAlloc.StackTrace = mp.formatStackTrace(alloc.StackTrace)
				}

				profile.Allocations = append(profile.Allocations, profiledAlloc)
			}
		}
	}
}

// formatStackTrace formats a stack trace for display
func (mp *MemoryProfiler) formatStackTrace(stackTrace []uintptr) []string {
	frames := runtime.CallersFrames(stackTrace)
	var result []string

	count := 0
	for {
		frame, more := frames.Next()
		if count >= mp.config.ProfileDepth {
			break
		}

		result = append(result, fmt.Sprintf("%s:%d %s", frame.File, frame.Line, frame.Function))
		count++

		if !more {
			break
		}
	}

	return result
}

// generateProfileSummary generates summary statistics for a profile
func (mp *MemoryProfiler) generateProfileSummary(profile *AllocationProfile) {
	summary := AllocationSummary{
		AllocationsByType: make(map[AllocationType]uint64),
		MemoryByType:      make(map[AllocationType]int64),
	}

	if len(profile.Allocations) == 0 {
		profile.Summary = summary
		return
	}

	var totalMemory int64
	var totalLifetime time.Duration
	var sizes []int64
	leakCandidates := uint64(0)

	for _, alloc := range profile.Allocations {
		summary.TotalAllocations++
		totalMemory += alloc.Size
		sizes = append(sizes, alloc.Size)

		summary.AllocationsByType[alloc.Type]++
		summary.MemoryByType[alloc.Type] += alloc.Size

		lifetime := time.Duration(alloc.LifetimeMs) * time.Millisecond
		totalLifetime += lifetime

		// Consider long-lived allocations as leak candidates
		if lifetime > 5*time.Minute && !alloc.Disposed {
			leakCandidates++
		}
	}

	summary.TotalMemory = totalMemory
	summary.AverageSize = float64(totalMemory) / float64(summary.TotalAllocations)
	summary.AverageLifetime = totalLifetime / time.Duration(summary.TotalAllocations)
	summary.LeakCandidates = leakCandidates

	// Calculate size statistics
	sort.Slice(sizes, func(i, j int) bool { return sizes[i] < sizes[j] })
	summary.MinSize = sizes[0]
	summary.MaxSize = sizes[len(sizes)-1]
	summary.MedianSize = sizes[len(sizes)/2]

	profile.Summary = summary
}

// findHotSpots identifies allocation hot spots
func (mp *MemoryProfiler) findHotSpots(profile *AllocationProfile) {
	locationStats := make(map[string]*AllocationHotSpot)

	for _, alloc := range profile.Allocations {
		if len(alloc.StackTrace) == 0 {
			continue
		}

		// Use the top stack frame as the location
		location := alloc.StackTrace[0]

		if hotSpot, exists := locationStats[location]; exists {
			hotSpot.Count++
			hotSpot.TotalMemory += alloc.Size
			hotSpot.AverageSize = float64(hotSpot.TotalMemory) / float64(hotSpot.Count)
		} else {
			// Parse location information
			function, file, line := mp.parseStackFrame(location)

			locationStats[location] = &AllocationHotSpot{
				Location:    location,
				Function:    function,
				File:        file,
				Line:        line,
				Count:       1,
				TotalMemory: alloc.Size,
				AverageSize: float64(alloc.Size),
				Type:        alloc.Type,
				Severity:    HotSpotLow,
			}
		}
	}

	// Convert to slice and sort by total memory
	var hotSpots []*AllocationHotSpot
	for _, hotSpot := range locationStats {
		// Determine severity based on count and memory usage
		if hotSpot.Count > 100 || hotSpot.TotalMemory > 1024*1024 {
			hotSpot.Severity = HotSpotCritical
		} else if hotSpot.Count > 50 || hotSpot.TotalMemory > 512*1024 {
			hotSpot.Severity = HotSpotHigh
		} else if hotSpot.Count > 20 || hotSpot.TotalMemory > 256*1024 {
			hotSpot.Severity = HotSpotMedium
		}

		hotSpots = append(hotSpots, hotSpot)
	}

	// Sort by total memory usage
	sort.Slice(hotSpots, func(i, j int) bool {
		return hotSpots[i].TotalMemory > hotSpots[j].TotalMemory
	})

	// Limit to top hot spots
	maxHotSpots := 20
	if len(hotSpots) > maxHotSpots {
		hotSpots = hotSpots[:maxHotSpots]
	}

	profile.HotSpots = hotSpots
}

// parseStackFrame parses a stack frame string
func (mp *MemoryProfiler) parseStackFrame(frame string) (function, file string, line int) {
	// Simple parsing - in a real implementation, this would be more robust
	// Format: "file:line function"
	// This is a simplified implementation
	function = "unknown"
	file = "unknown"
	line = 0

	// In a real implementation, you would parse the frame string properly
	// For now, just return defaults
	return function, file, line
}

// analyzeTrends analyzes allocation trends
func (mp *MemoryProfiler) analyzeTrends(profile *AllocationProfile) {
	if len(profile.Allocations) < 2 {
		return
	}

	trends := &AllocationTrends{}

	// Calculate allocation rate
	duration := profile.Duration.Seconds()
	if duration > 0 {
		trends.AllocationRate = float64(profile.Summary.TotalAllocations) / duration
		trends.MemoryGrowthRate = float64(profile.Summary.TotalMemory) / duration
	}

	// Analyze leak trend (simplified)
	if profile.Summary.LeakCandidates > profile.Summary.TotalAllocations/10 {
		trends.LeakTrend = "increasing"
		trends.Confidence = 0.8
	} else if profile.Summary.LeakCandidates > profile.Summary.TotalAllocations/20 {
		trends.LeakTrend = "stable"
		trends.Confidence = 0.6
	} else {
		trends.LeakTrend = "decreasing"
		trends.Confidence = 0.7
	}

	// Predict peak time (simplified)
	if trends.MemoryGrowthRate > 0 {
		// Assume we'll hit a 100MB limit
		remainingMemory := 100*1024*1024 - profile.Summary.TotalMemory
		if remainingMemory > 0 {
			trends.PredictedPeakTime = time.Duration(float64(remainingMemory)/trends.MemoryGrowthRate) * time.Second
		}
	}

	profile.Trends = trends
}

// ------------------------------------
// 📋 Reporting and Analysis
// ------------------------------------

// GenerateReport generates a comprehensive memory analysis report
func (mp *MemoryProfiler) GenerateReport() map[string]interface{} {
	mp.mutex.RLock()
	defer mp.mutex.RUnlock()

	report := map[string]interface{}{
		"profiler_info": map[string]interface{}{
			"enabled":        mp.enabled,
			"profiling":      mp.profiling,
			"config":         mp.config,
			"snapshot_count": len(mp.snapshots),
			"profile_count":  len(mp.profiles),
		},
	}

	// Add latest snapshot
	if len(mp.snapshots) > 0 {
		report["latest_snapshot"] = mp.snapshots[len(mp.snapshots)-1]
	}

	// Add snapshot trends
	if len(mp.snapshots) > 1 {
		report["snapshot_trends"] = mp.analyzeSnapshotTrends()
	}

	// Add profile summaries
	if len(mp.profiles) > 0 {
		profileSummaries := make([]map[string]interface{}, 0)
		for _, profile := range mp.profiles {
			profileSummaries = append(profileSummaries, map[string]interface{}{
				"id":              profile.ID,
				"name":            profile.Name,
				"duration":        profile.Duration,
				"summary":         profile.Summary,
				"hot_spots_count": len(profile.HotSpots),
			})
		}
		report["profiles"] = profileSummaries
	}

	// Add recommendations
	report["recommendations"] = mp.generateRecommendations()

	return report
}

// analyzeSnapshotTrends analyzes trends across snapshots
func (mp *MemoryProfiler) analyzeSnapshotTrends() map[string]interface{} {
	if len(mp.snapshots) < 2 {
		return nil
	}

	first := mp.snapshots[0]
	last := mp.snapshots[len(mp.snapshots)-1]
	duration := last.Timestamp.Sub(first.Timestamp)

	trends := map[string]interface{}{
		"duration_minutes": duration.Minutes(),
		"memory_growth": map[string]interface{}{
			"total_bytes":  last.TotalMemory - first.TotalMemory,
			"active_bytes": last.ActiveMemory - first.ActiveMemory,
			"growth_rate":  float64(last.ActiveMemory-first.ActiveMemory) / duration.Seconds(),
		},
		"allocation_growth": map[string]interface{}{
			"total_count":  last.TotalAllocations - first.TotalAllocations,
			"active_count": last.ActiveAllocations - first.ActiveAllocations,
			"growth_rate":  float64(last.TotalAllocations-first.TotalAllocations) / duration.Seconds(),
		},
	}

	// Calculate average values
	var totalMemory, activeMemory int64
	var totalAllocs, activeAllocs uint64

	for _, snapshot := range mp.snapshots {
		totalMemory += snapshot.TotalMemory
		activeMemory += snapshot.ActiveMemory
		totalAllocs += snapshot.TotalAllocations
		activeAllocs += snapshot.ActiveAllocations
	}

	count := int64(len(mp.snapshots))
	trends["averages"] = map[string]interface{}{
		"total_memory":       totalMemory / count,
		"active_memory":      activeMemory / count,
		"total_allocations":  totalAllocs / uint64(count),
		"active_allocations": activeAllocs / uint64(count),
	}

	return trends
}

// generateRecommendations generates optimization recommendations
func (mp *MemoryProfiler) generateRecommendations() []string {
	var recommendations []string

	// Analyze latest snapshot
	if len(mp.snapshots) > 0 {
		latest := mp.snapshots[len(mp.snapshots)-1]

		// Check for high memory usage
		if latest.ActiveMemory > 100*1024*1024 { // 100MB
			recommendations = append(recommendations, "High memory usage detected. Consider implementing more aggressive cleanup strategies.")
		}

		// Check for leak violations
		if latest.LeakViolations > 0 {
			recommendations = append(recommendations, fmt.Sprintf("Memory leak violations detected (%d). Review resource cleanup patterns.", latest.LeakViolations))
		}

		// Check allocation distribution
		maxTypeCount := uint64(0)
		var dominantType AllocationType
		for allocType, count := range latest.AllocationsByType {
			if count > maxTypeCount {
				maxTypeCount = count
				dominantType = allocType
			}
		}

		if maxTypeCount > latest.ActiveAllocations/2 {
			recommendations = append(recommendations, fmt.Sprintf("Allocation pattern dominated by %s allocations. Consider pooling or caching strategies.", dominantType.String()))
		}
	}

	// Analyze profiles
	for _, profile := range mp.profiles {
		if profile.Summary.LeakCandidates > 0 {
			recommendations = append(recommendations, fmt.Sprintf("Profile '%s' shows %d potential leak candidates. Review long-lived allocations.", profile.Name, profile.Summary.LeakCandidates))
		}

		// Check for hot spots
		for _, hotSpot := range profile.HotSpots {
			if hotSpot.Severity >= HotSpotHigh {
				recommendations = append(recommendations, fmt.Sprintf("High-impact allocation hot spot detected at %s (%d allocations, %d bytes)", hotSpot.Function, hotSpot.Count, hotSpot.TotalMemory))
			}
		}
	}

	// General recommendations
	if len(recommendations) == 0 {
		recommendations = append(recommendations, "Memory usage appears healthy. Continue monitoring for optimal performance.")
	}

	return recommendations
}

// ExportProfile exports a profile to JSON format
func (mp *MemoryProfiler) ExportProfile(profileID uint64) ([]byte, error) {
	mp.mutex.RLock()
	defer mp.mutex.RUnlock()

	for _, profile := range mp.profiles {
		if profile.ID == profileID {
			return json.MarshalIndent(profile, "", "  ")
		}
	}

	return nil, fmt.Errorf("profile with ID %d not found", profileID)
}

// ExportSnapshots exports snapshots to JSON format
func (mp *MemoryProfiler) ExportSnapshots() ([]byte, error) {
	mp.mutex.RLock()
	defer mp.mutex.RUnlock()

	return json.MarshalIndent(mp.snapshots, "", "  ")
}

// ClearSnapshots clears all stored snapshots
func (mp *MemoryProfiler) ClearSnapshots() {
	mp.mutex.Lock()
	defer mp.mutex.Unlock()
	mp.snapshots = make([]*MemorySnapshot, 0)
}

// ClearProfiles clears all stored profiles
func (mp *MemoryProfiler) ClearProfiles() {
	mp.mutex.Lock()
	defer mp.mutex.Unlock()
	mp.profiles = make([]*AllocationProfile, 0)
}

// SetEnabled enables or disables the profiler
func (mp *MemoryProfiler) SetEnabled(enabled bool) {
	mp.mutex.Lock()
	defer mp.mutex.Unlock()
	mp.enabled = enabled
}

// UpdateConfig updates the profiler configuration
func (mp *MemoryProfiler) UpdateConfig(config ProfilerConfig) {
	mp.mutex.Lock()
	defer mp.mutex.Unlock()
	mp.config = config
}

// GetProfiles returns all allocation profiles
func (mp *MemoryProfiler) GetProfiles() []*AllocationProfile {
	mp.mutex.RLock()
	defer mp.mutex.RUnlock()

	profiles := make([]*AllocationProfile, len(mp.profiles))
	copy(profiles, mp.profiles)
	return profiles
}

// GetProfile returns a specific allocation profile
func (mp *MemoryProfiler) GetProfile(profileID uint64) *AllocationProfile {
	mp.mutex.RLock()
	defer mp.mutex.RUnlock()

	for _, profile := range mp.profiles {
		if profile.ID == profileID {
			return profile
		}
	}

	return nil
}
