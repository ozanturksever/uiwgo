// leak_detector.go
// Advanced leak detection with pattern recognition and intelligent analysis

package golid

import (
	"context"
	"fmt"
	"runtime"
	"sort"
	"sync"
	"sync/atomic"
	"time"
)

// ------------------------------------
// 🔍 Advanced Leak Detector Core
// ------------------------------------

var leakDetectorIdCounter uint64

// LeakDetector provides advanced memory leak detection with pattern recognition
type LeakDetector struct {
	id            uint64
	memoryManager *MemoryManager

	// Detection data
	allocations      map[uint64]*Allocation
	patterns         *PatternAnalyzer
	violations       []*LeakViolation
	suspiciousAllocs map[uint64]*SuspiciousAllocation

	// Configuration
	config LeakDetectionConfig

	// State
	enabled   bool
	analyzing bool

	// Statistics
	stats LeakDetectionStats

	// Background detection
	ctx             context.Context
	cancel          context.CancelFunc
	detectionTicker *time.Ticker
	stopDetection   chan bool

	// Synchronization
	mutex sync.RWMutex
}

// LeakDetectionConfig provides configuration for leak detection
type LeakDetectionConfig struct {
	CheckInterval         time.Duration `json:"check_interval"`
	ThresholdCount        int           `json:"threshold_count"`
	ThresholdMemory       int64         `json:"threshold_memory"`
	SuspiciousAge         time.Duration `json:"suspicious_age"`
	PatternAnalysisDepth  int           `json:"pattern_analysis_depth"`
	EnablePatternLearning bool          `json:"enable_pattern_learning"`
	EnableStackAnalysis   bool          `json:"enable_stack_analysis"`
	EnableTrendAnalysis   bool          `json:"enable_trend_analysis"`
	AlertThreshold        int           `json:"alert_threshold"`
	FalsePositiveFilter   bool          `json:"false_positive_filter"`
}

// LeakDetectionStats tracks leak detection statistics
type LeakDetectionStats struct {
	TotalChecks           uint64        `json:"total_checks"`
	LeaksDetected         uint64        `json:"leaks_detected"`
	FalsePositives        uint64        `json:"false_positives"`
	PatternsIdentified    uint64        `json:"patterns_identified"`
	SuspiciousAllocations uint64        `json:"suspicious_allocations"`
	AverageCheckTime      time.Duration `json:"average_check_time"`
	LastCheckTime         time.Time     `json:"last_check_time"`
	StartTime             time.Time     `json:"start_time"`
	mutex                 sync.RWMutex
}

// SuspiciousAllocation represents an allocation that might be leaking
type SuspiciousAllocation struct {
	Allocation     *Allocation            `json:"allocation"`
	SuspicionLevel SuspicionLevel         `json:"suspicion_level"`
	Reasons        []SuspicionReason      `json:"reasons"`
	FirstDetected  time.Time              `json:"first_detected"`
	LastSeen       time.Time              `json:"last_seen"`
	CheckCount     int                    `json:"check_count"`
	PatternMatches []string               `json:"pattern_matches"`
	StackAnalysis  *StackAnalysis         `json:"stack_analysis,omitempty"`
	TrendAnalysis  *TrendAnalysis         `json:"trend_analysis,omitempty"`
	Metadata       map[string]interface{} `json:"metadata"`
}

// SuspicionLevel defines levels of suspicion for potential leaks
type SuspicionLevel int

const (
	SuspicionLow SuspicionLevel = iota
	SuspicionMedium
	SuspicionHigh
	SuspicionCritical
)

// String returns string representation of suspicion level
func (s SuspicionLevel) String() string {
	switch s {
	case SuspicionLow:
		return "Low"
	case SuspicionMedium:
		return "Medium"
	case SuspicionHigh:
		return "High"
	case SuspicionCritical:
		return "Critical"
	default:
		return "Unknown"
	}
}

// SuspicionReason defines reasons for suspicion
type SuspicionReason string

const (
	ReasonAge               SuspicionReason = "old_allocation"
	ReasonHighAccessCount   SuspicionReason = "high_access_count"
	ReasonNoRecentAccess    SuspicionReason = "no_recent_access"
	ReasonPatternMatch      SuspicionReason = "pattern_match"
	ReasonStackSimilarity   SuspicionReason = "stack_similarity"
	ReasonOwnerDisposed     SuspicionReason = "owner_disposed"
	ReasonThresholdExceeded SuspicionReason = "threshold_exceeded"
	ReasonTrendAnalysis     SuspicionReason = "trend_analysis"
)

// LeakViolation represents a detected memory leak violation
type LeakViolation struct {
	ResourceType ResourceType           `json:"resource_type"`
	Count        int                    `json:"count"`
	Threshold    int                    `json:"threshold"`
	DetectedAt   time.Time              `json:"detected_at"`
	Severity     string                 `json:"severity"`
	Description  string                 `json:"description"`
	Metadata     map[string]interface{} `json:"metadata"`
}

// PatternAnalyzer analyzes allocation patterns to identify potential leaks
type PatternAnalyzer struct {
	patterns        map[string]*LeakPattern
	learningEnabled bool
	analysisDepth   int
	mutex           sync.RWMutex
}

// LeakPattern represents a detected leak pattern
type LeakPattern struct {
	ID              string                 `json:"id"`
	Name            string                 `json:"name"`
	Description     string                 `json:"description"`
	Confidence      float64                `json:"confidence"`
	Occurrences     int                    `json:"occurrences"`
	LastSeen        time.Time              `json:"last_seen"`
	Characteristics map[string]interface{} `json:"characteristics"`
	StackSignatures []string               `json:"stack_signatures"`
	TypePatterns    []AllocationType       `json:"type_patterns"`
	SizePatterns    []int64                `json:"size_patterns"`
	TimingPatterns  []time.Duration        `json:"timing_patterns"`
}

// StackAnalysis provides analysis of allocation stack traces
type StackAnalysis struct {
	Signature          string   `json:"signature"`
	CommonFrames       []string `json:"common_frames"`
	SimilarAllocations int      `json:"similar_allocations"`
	Confidence         float64  `json:"confidence"`
}

// TrendAnalysis provides trend analysis for allocations
type TrendAnalysis struct {
	GrowthRate   float64   `json:"growth_rate"`
	Trend        string    `json:"trend"`
	Prediction   string    `json:"prediction"`
	Confidence   float64   `json:"confidence"`
	DataPoints   int       `json:"data_points"`
	AnalysisTime time.Time `json:"analysis_time"`
}

// ------------------------------------
// 🏗️ Leak Detector Creation
// ------------------------------------

// NewLeakDetector creates a new advanced leak detector
func NewLeakDetector(memoryManager *MemoryManager) *LeakDetector {
	ctx, cancel := context.WithCancel(context.Background())

	config := DefaultLeakDetectionConfig()

	ld := &LeakDetector{
		id:               atomic.AddUint64(&leakDetectorIdCounter, 1),
		memoryManager:    memoryManager,
		allocations:      make(map[uint64]*Allocation),
		patterns:         NewPatternAnalyzer(config.PatternAnalysisDepth, config.EnablePatternLearning),
		violations:       make([]*LeakViolation, 0),
		suspiciousAllocs: make(map[uint64]*SuspiciousAllocation),
		config:           config,
		enabled:          true,
		analyzing:        true,
		ctx:              ctx,
		cancel:           cancel,
		detectionTicker:  time.NewTicker(config.CheckInterval),
		stopDetection:    make(chan bool, 1),
		stats: LeakDetectionStats{
			StartTime: time.Now(),
		},
	}

	// Start background detection
	go ld.backgroundDetection()

	return ld
}

// DefaultLeakDetectionConfig returns default leak detection configuration
func DefaultLeakDetectionConfig() LeakDetectionConfig {
	return LeakDetectionConfig{
		CheckInterval:         30 * time.Second,
		ThresholdCount:        100,
		ThresholdMemory:       50 * 1024 * 1024, // 50MB
		SuspiciousAge:         5 * time.Minute,
		PatternAnalysisDepth:  10,
		EnablePatternLearning: true,
		EnableStackAnalysis:   true,
		EnableTrendAnalysis:   true,
		AlertThreshold:        10,
		FalsePositiveFilter:   true,
	}
}

// NewPatternAnalyzer creates a new pattern analyzer
func NewPatternAnalyzer(depth int, learningEnabled bool) *PatternAnalyzer {
	return &PatternAnalyzer{
		patterns:        make(map[string]*LeakPattern),
		learningEnabled: learningEnabled,
		analysisDepth:   depth,
	}
}

// ------------------------------------
// 🔍 Leak Detection Methods
// ------------------------------------

// CheckAllocation checks a specific allocation for potential leaks
func (ld *LeakDetector) CheckAllocation(alloc *Allocation) {
	if !ld.enabled || alloc == nil {
		return
	}

	ld.mutex.Lock()
	ld.allocations[alloc.ID] = alloc
	ld.mutex.Unlock()

	// Perform immediate analysis
	suspicious := ld.analyzeAllocation(alloc)
	if suspicious != nil {
		ld.mutex.Lock()
		ld.suspiciousAllocs[alloc.ID] = suspicious
		ld.mutex.Unlock()

		ld.updateStats(func(stats *LeakDetectionStats) {
			atomic.AddUint64(&stats.SuspiciousAllocations, 1)
		})
	}
}

// PerformLeakCheck performs a comprehensive leak detection check
func (ld *LeakDetector) PerformLeakCheck() {
	if !ld.enabled {
		return
	}

	start := time.Now()
	defer func() {
		duration := time.Since(start)
		ld.updateStats(func(stats *LeakDetectionStats) {
			atomic.AddUint64(&stats.TotalChecks, 1)
			stats.LastCheckTime = time.Now()

			// Update average check time
			totalChecks := atomic.LoadUint64(&stats.TotalChecks)
			if totalChecks > 0 {
				stats.AverageCheckTime = time.Duration(
					(int64(stats.AverageCheckTime)*int64(totalChecks-1) + int64(duration)) / int64(totalChecks),
				)
			}
		})
	}()

	// Get current allocations from memory manager
	if ld.memoryManager != nil {
		ld.syncAllocations()
	}

	// Analyze all allocations
	ld.analyzeAllocations()

	// Perform pattern analysis
	if ld.config.EnablePatternLearning {
		ld.patterns.AnalyzePatterns(ld.suspiciousAllocs)
	}

	// Check for violations
	violations := ld.checkViolations()
	if len(violations) > 0 {
		ld.mutex.Lock()
		ld.violations = append(ld.violations, violations...)
		ld.mutex.Unlock()

		ld.updateStats(func(stats *LeakDetectionStats) {
			atomic.AddUint64(&stats.LeaksDetected, uint64(len(violations)))
		})
	}
}

// syncAllocations synchronizes allocations with memory manager
func (ld *LeakDetector) syncAllocations() {
	if ld.memoryManager == nil {
		return
	}

	ld.mutex.Lock()
	defer ld.mutex.Unlock()

	// Clear old allocations
	ld.allocations = make(map[uint64]*Allocation)

	// Get all allocation types
	for allocType := AllocSignal; allocType <= AllocCustom; allocType++ {
		allocs := ld.memoryManager.GetAllocationsByType(allocType)
		for _, alloc := range allocs {
			ld.allocations[alloc.ID] = alloc
		}
	}
}

// analyzeAllocations analyzes all current allocations
func (ld *LeakDetector) analyzeAllocations() {
	ld.mutex.RLock()
	allocations := make([]*Allocation, 0, len(ld.allocations))
	for _, alloc := range ld.allocations {
		allocations = append(allocations, alloc)
	}
	ld.mutex.RUnlock()

	for _, alloc := range allocations {
		if suspicious := ld.analyzeAllocation(alloc); suspicious != nil {
			ld.mutex.Lock()
			ld.suspiciousAllocs[alloc.ID] = suspicious
			ld.mutex.Unlock()
		}
	}
}

// analyzeAllocation analyzes a single allocation for suspicious behavior
func (ld *LeakDetector) analyzeAllocation(alloc *Allocation) *SuspiciousAllocation {
	if alloc == nil || alloc.Disposed {
		return nil
	}

	suspicious := &SuspiciousAllocation{
		Allocation:     alloc,
		SuspicionLevel: SuspicionLow,
		Reasons:        make([]SuspicionReason, 0),
		FirstDetected:  time.Now(),
		LastSeen:       time.Now(),
		CheckCount:     1,
		PatternMatches: make([]string, 0),
		Metadata:       make(map[string]interface{}),
	}

	// Check if already suspicious
	ld.mutex.RLock()
	if existing, exists := ld.suspiciousAllocs[alloc.ID]; exists {
		suspicious = existing
		suspicious.LastSeen = time.Now()
		suspicious.CheckCount++
	}
	ld.mutex.RUnlock()

	// Age analysis
	age := time.Since(alloc.Timestamp)
	if age > ld.config.SuspiciousAge {
		suspicious.Reasons = append(suspicious.Reasons, ReasonAge)
		suspicious.SuspicionLevel = ld.increaseSuspicion(suspicious.SuspicionLevel)
	}

	// Access pattern analysis
	alloc.mutex.RLock()
	timeSinceLastAccess := time.Since(alloc.LastAccessed)
	accessCount := alloc.AccessCount
	alloc.mutex.RUnlock()

	if timeSinceLastAccess > ld.config.SuspiciousAge/2 {
		suspicious.Reasons = append(suspicious.Reasons, ReasonNoRecentAccess)
		suspicious.SuspicionLevel = ld.increaseSuspicion(suspicious.SuspicionLevel)
	}

	if accessCount > 1000 { // High access count might indicate a leak
		suspicious.Reasons = append(suspicious.Reasons, ReasonHighAccessCount)
		suspicious.SuspicionLevel = ld.increaseSuspicion(suspicious.SuspicionLevel)
	}

	// Owner analysis
	if alloc.Owner != nil {
		alloc.Owner.mutex.RLock()
		ownerDisposed := alloc.Owner.disposed
		alloc.Owner.mutex.RUnlock()

		if ownerDisposed {
			suspicious.Reasons = append(suspicious.Reasons, ReasonOwnerDisposed)
			suspicious.SuspicionLevel = SuspicionCritical
		}
	}

	// Stack analysis
	if ld.config.EnableStackAnalysis && len(alloc.StackTrace) > 0 {
		stackAnalysis := ld.analyzeStackTrace(alloc.StackTrace)
		if stackAnalysis != nil {
			suspicious.StackAnalysis = stackAnalysis
			if stackAnalysis.Confidence > 0.7 {
				suspicious.Reasons = append(suspicious.Reasons, ReasonStackSimilarity)
				suspicious.SuspicionLevel = ld.increaseSuspicion(suspicious.SuspicionLevel)
			}
		}
	}

	// Pattern matching
	patterns := ld.patterns.MatchPatterns(alloc)
	if len(patterns) > 0 {
		suspicious.PatternMatches = patterns
		suspicious.Reasons = append(suspicious.Reasons, ReasonPatternMatch)
		suspicious.SuspicionLevel = ld.increaseSuspicion(suspicious.SuspicionLevel)
	}

	// Trend analysis
	if ld.config.EnableTrendAnalysis {
		trendAnalysis := ld.analyzeTrend(alloc)
		if trendAnalysis != nil {
			suspicious.TrendAnalysis = trendAnalysis
			if trendAnalysis.Confidence > 0.6 && trendAnalysis.Trend == "increasing" {
				suspicious.Reasons = append(suspicious.Reasons, ReasonTrendAnalysis)
				suspicious.SuspicionLevel = ld.increaseSuspicion(suspicious.SuspicionLevel)
			}
		}
	}

	// Return only if suspicious
	if len(suspicious.Reasons) > 0 {
		return suspicious
	}

	return nil
}

// increaseSuspicion increases the suspicion level
func (ld *LeakDetector) increaseSuspicion(current SuspicionLevel) SuspicionLevel {
	switch current {
	case SuspicionLow:
		return SuspicionMedium
	case SuspicionMedium:
		return SuspicionHigh
	case SuspicionHigh:
		return SuspicionCritical
	default:
		return current
	}
}

// analyzeStackTrace analyzes stack trace for patterns
func (ld *LeakDetector) analyzeStackTrace(stackTrace []uintptr) *StackAnalysis {
	if len(stackTrace) == 0 {
		return nil
	}

	frames := runtime.CallersFrames(stackTrace)
	var frameStrings []string

	for {
		frame, more := frames.Next()
		frameStrings = append(frameStrings, fmt.Sprintf("%s:%d", frame.Function, frame.Line))
		if !more {
			break
		}
	}

	// Create signature from top frames
	signature := ""
	if len(frameStrings) > 0 {
		maxFrames := 5
		if len(frameStrings) < maxFrames {
			maxFrames = len(frameStrings)
		}
		for i := 0; i < maxFrames; i++ {
			if i > 0 {
				signature += "|"
			}
			signature += frameStrings[i]
		}
	}

	// Find similar allocations
	similarCount := ld.countSimilarStackTraces(signature)

	confidence := 0.0
	if similarCount > 1 {
		confidence = float64(similarCount) / 10.0 // Simple confidence calculation
		if confidence > 1.0 {
			confidence = 1.0
		}
	}

	return &StackAnalysis{
		Signature:          signature,
		CommonFrames:       frameStrings,
		SimilarAllocations: similarCount,
		Confidence:         confidence,
	}
}

// countSimilarStackTraces counts allocations with similar stack traces
func (ld *LeakDetector) countSimilarStackTraces(signature string) int {
	count := 0

	ld.mutex.RLock()
	for _, suspicious := range ld.suspiciousAllocs {
		if suspicious.StackAnalysis != nil && suspicious.StackAnalysis.Signature == signature {
			count++
		}
	}
	ld.mutex.RUnlock()

	return count
}

// analyzeTrend analyzes allocation trends
func (ld *LeakDetector) analyzeTrend(alloc *Allocation) *TrendAnalysis {
	// Simple trend analysis based on access patterns
	alloc.mutex.RLock()
	accessCount := alloc.AccessCount
	age := time.Since(alloc.Timestamp)
	alloc.mutex.RUnlock()

	if age.Seconds() == 0 {
		return nil
	}

	accessRate := float64(accessCount) / age.Seconds()

	trend := "stable"
	prediction := "normal"
	confidence := 0.5

	if accessRate > 10 { // High access rate
		trend = "increasing"
		prediction = "potential_leak"
		confidence = 0.7
	} else if accessRate < 0.1 { // Very low access rate
		trend = "decreasing"
		prediction = "unused_resource"
		confidence = 0.6
	}

	return &TrendAnalysis{
		GrowthRate:   accessRate,
		Trend:        trend,
		Prediction:   prediction,
		Confidence:   confidence,
		DataPoints:   1, // Would be more in a real implementation
		AnalysisTime: time.Now(),
	}
}

// checkViolations checks for leak violations based on thresholds
func (ld *LeakDetector) checkViolations() []*LeakViolation {
	var violations []*LeakViolation

	// Count allocations by type
	typeCounts := make(map[AllocationType]int)
	totalMemory := int64(0)

	ld.mutex.RLock()
	for _, alloc := range ld.allocations {
		if !alloc.Disposed {
			typeCounts[alloc.Type]++
			totalMemory += alloc.Size
		}
	}
	ld.mutex.RUnlock()

	// Check count thresholds
	for allocType, count := range typeCounts {
		if count > ld.config.ThresholdCount {
			violation := &LeakViolation{
				ResourceType: ResourceType(allocType),
				Count:        count,
				Threshold:    ld.config.ThresholdCount,
				DetectedAt:   time.Now(),
			}
			violations = append(violations, violation)
		}
	}

	// Check memory threshold
	if totalMemory > ld.config.ThresholdMemory {
		violation := &LeakViolation{
			ResourceType: ResourceCustom, // Use custom for memory violations
			Count:        int(totalMemory),
			Threshold:    int(ld.config.ThresholdMemory),
			DetectedAt:   time.Now(),
		}
		violations = append(violations, violation)
	}

	return violations
}

// ------------------------------------
// 🧠 Pattern Analysis
// ------------------------------------

// AnalyzePatterns analyzes suspicious allocations for patterns
func (pa *PatternAnalyzer) AnalyzePatterns(suspiciousAllocs map[uint64]*SuspiciousAllocation) {
	if !pa.learningEnabled {
		return
	}

	pa.mutex.Lock()
	defer pa.mutex.Unlock()

	// Group allocations by characteristics
	typeGroups := make(map[AllocationType][]*SuspiciousAllocation)
	sizeGroups := make(map[int64][]*SuspiciousAllocation)

	for _, suspicious := range suspiciousAllocs {
		alloc := suspicious.Allocation
		typeGroups[alloc.Type] = append(typeGroups[alloc.Type], suspicious)

		// Group by size ranges
		sizeRange := (alloc.Size / 1024) * 1024 // Round to KB
		sizeGroups[sizeRange] = append(sizeGroups[sizeRange], suspicious)
	}

	// Analyze type patterns
	for allocType, group := range typeGroups {
		if len(group) >= 3 { // Minimum group size for pattern
			pa.createOrUpdatePattern(fmt.Sprintf("type_%s", allocType.String()), group)
		}
	}

	// Analyze size patterns
	for sizeRange, group := range sizeGroups {
		if len(group) >= 3 {
			pa.createOrUpdatePattern(fmt.Sprintf("size_%d", sizeRange), group)
		}
	}
}

// createOrUpdatePattern creates or updates a leak pattern
func (pa *PatternAnalyzer) createOrUpdatePattern(id string, group []*SuspiciousAllocation) {
	pattern, exists := pa.patterns[id]
	if !exists {
		pattern = &LeakPattern{
			ID:              id,
			Name:            fmt.Sprintf("Pattern %s", id),
			Description:     fmt.Sprintf("Detected pattern for %s", id),
			Confidence:      0.5,
			Occurrences:     0,
			Characteristics: make(map[string]interface{}),
			StackSignatures: make([]string, 0),
			TypePatterns:    make([]AllocationType, 0),
			SizePatterns:    make([]int64, 0),
			TimingPatterns:  make([]time.Duration, 0),
		}
		pa.patterns[id] = pattern
	}

	pattern.Occurrences += len(group)
	pattern.LastSeen = time.Now()

	// Update confidence based on group size and consistency
	pattern.Confidence = float64(len(group)) / 10.0
	if pattern.Confidence > 1.0 {
		pattern.Confidence = 1.0
	}

	// Extract characteristics
	for _, suspicious := range group {
		alloc := suspicious.Allocation

		// Add type pattern
		found := false
		for _, existingType := range pattern.TypePatterns {
			if existingType == alloc.Type {
				found = true
				break
			}
		}
		if !found {
			pattern.TypePatterns = append(pattern.TypePatterns, alloc.Type)
		}

		// Add size pattern
		found = false
		for _, existingSize := range pattern.SizePatterns {
			if existingSize == alloc.Size {
				found = true
				break
			}
		}
		if !found {
			pattern.SizePatterns = append(pattern.SizePatterns, alloc.Size)
		}

		// Add stack signature if available
		if suspicious.StackAnalysis != nil {
			found = false
			for _, existingSig := range pattern.StackSignatures {
				if existingSig == suspicious.StackAnalysis.Signature {
					found = true
					break
				}
			}
			if !found {
				pattern.StackSignatures = append(pattern.StackSignatures, suspicious.StackAnalysis.Signature)
			}
		}
	}
}

// MatchPatterns matches an allocation against known patterns
func (pa *PatternAnalyzer) MatchPatterns(alloc *Allocation) []string {
	pa.mutex.RLock()
	defer pa.mutex.RUnlock()

	var matches []string

	for _, pattern := range pa.patterns {
		if pa.matchesPattern(alloc, pattern) {
			matches = append(matches, pattern.ID)
		}
	}

	return matches
}

// matchesPattern checks if an allocation matches a pattern
func (pa *PatternAnalyzer) matchesPattern(alloc *Allocation, pattern *LeakPattern) bool {
	// Check type pattern
	for _, typePattern := range pattern.TypePatterns {
		if alloc.Type == typePattern {
			return true
		}
	}

	// Check size pattern (within 10% tolerance)
	for _, sizePattern := range pattern.SizePatterns {
		tolerance := sizePattern / 10
		if alloc.Size >= sizePattern-tolerance && alloc.Size <= sizePattern+tolerance {
			return true
		}
	}

	return false
}

// ------------------------------------
// 🔄 Background Detection
// ------------------------------------

// backgroundDetection runs continuous leak detection
func (ld *LeakDetector) backgroundDetection() {
	for {
		select {
		case <-ld.ctx.Done():
			return
		case <-ld.stopDetection:
			return
		case <-ld.detectionTicker.C:
			ld.PerformLeakCheck()
		}
	}
}

// ------------------------------------
// 📊 Reporting and Statistics
// ------------------------------------

// GetReport returns a comprehensive leak detection report
func (ld *LeakDetector) GetReport() map[string]interface{} {
	ld.mutex.RLock()
	defer ld.mutex.RUnlock()

	report := map[string]interface{}{
		"detector_id":       ld.id,
		"enabled":           ld.enabled,
		"analyzing":         ld.analyzing,
		"config":            ld.config,
		"stats":             ld.GetStats(),
		"total_allocations": len(ld.allocations),
		"suspicious_count":  len(ld.suspiciousAllocs),
		"violations_count":  len(ld.violations),
	}

	// Add suspicious allocations by level
	suspicionBreakdown := make(map[string]int)
	for _, suspicious := range ld.suspiciousAllocs {
		suspicionBreakdown[suspicious.SuspicionLevel.String()]++
	}
	report["suspicion_breakdown"] = suspicionBreakdown

	// Add pattern information
	ld.patterns.mutex.RLock()
	patternInfo := make(map[string]interface{})
	for id, pattern := range ld.patterns.patterns {
		patternInfo[id] = map[string]interface{}{
			"confidence":  pattern.Confidence,
			"occurrences": pattern.Occurrences,
			"last_seen":   pattern.LastSeen,
		}
	}
	ld.patterns.mutex.RUnlock()
	report["patterns"] = patternInfo

	return report
}

// GetStats returns current leak detection statistics
func (ld *LeakDetector) GetStats() LeakDetectionStats {
	ld.stats.mutex.RLock()
	defer ld.stats.mutex.RUnlock()
	return ld.stats
}

// GetSuspiciousAllocations returns all suspicious allocations
func (ld *LeakDetector) GetSuspiciousAllocations() []*SuspiciousAllocation {
	ld.mutex.RLock()
	defer ld.mutex.RUnlock()

	var suspicious []*SuspiciousAllocation
	for _, alloc := range ld.suspiciousAllocs {
		suspicious = append(suspicious, alloc)
	}

	// Sort by suspicion level
	sort.Slice(suspicious, func(i, j int) bool {
		return suspicious[i].SuspicionLevel > suspicious[j].SuspicionLevel
	})

	return suspicious
}

// GetViolations returns all detected violations
func (ld *LeakDetector) GetViolations() []*LeakViolation {
	ld.mutex.RLock()
	defer ld.mutex.RUnlock()

	violations := make([]*LeakViolation, len(ld.violations))
	copy(violations, ld.violations)
	return violations
}

// ------------------------------------
// 🛠️ Configuration and Control
// ------------------------------------

// SetEnabled enables or disables leak detection
func (ld *LeakDetector) SetEnabled(enabled bool) {
	ld.mutex.Lock()
	defer ld.mutex.Unlock()
	ld.enabled = enabled
}

// SetAnalyzing enables or disables pattern analysis
func (ld *LeakDetector) SetAnalyzing(analyzing bool) {
	ld.mutex.Lock()
	defer ld.mutex.Unlock()
	ld.analyzing = analyzing
}

// UpdateConfig updates the leak detection configuration
func (ld *LeakDetector) UpdateConfig(config LeakDetectionConfig) {
	ld.mutex.Lock()
	defer ld.mutex.Unlock()

	ld.config = config

	// Update pattern analyzer settings
	ld.patterns.mutex.Lock()
	ld.patterns.learningEnabled = config.EnablePatternLearning
	ld.patterns.analysisDepth = config.PatternAnalysisDepth
	ld.patterns.mutex.Unlock()

	// Update ticker interval if changed
	if ld.detectionTicker != nil {
		ld.detectionTicker.Stop()
		ld.detectionTicker = time.NewTicker(config.CheckInterval)
	}
}

// updateStats safely updates leak detection statistics
func (ld *LeakDetector) updateStats(fn func(*LeakDetectionStats)) {
	ld.stats.mutex.Lock()
	defer ld.stats.mutex.Unlock()
	fn(&ld.stats)
}

// ------------------------------------
// 🧹 Disposal
// ------------------------------------

// Dispose cleans up the leak detector and all its resources
func (ld *LeakDetector) Dispose() {
	ld.mutex.Lock()
	defer ld.mutex.Unlock()

	// Stop background detection
	if ld.cancel != nil {
		ld.cancel()
	}

	if ld.detectionTicker != nil {
		ld.detectionTicker.Stop()
	}

	select {
	case ld.stopDetection <- true:
	default:
	}

	// Clear data
	ld.allocations = make(map[uint64]*Allocation)
	ld.violations = make([]*LeakViolation, 0)
	ld.suspiciousAllocs = make(map[uint64]*SuspiciousAllocation)

	// Clear patterns
	if ld.patterns != nil {
		ld.patterns.mutex.Lock()
		ld.patterns.patterns = make(map[string]*LeakPattern)
		ld.patterns.mutex.Unlock()
	}

	ld.enabled = false
	ld.analyzing = false
}
