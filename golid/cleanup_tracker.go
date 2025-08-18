// cleanup_tracker.go
// Comprehensive cleanup tracking and verification system

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
// 🧹 Cleanup Tracker Core
// ------------------------------------

var cleanupTrackerIdCounter uint64

// CleanupTracker tracks all cleanup operations and verifies completion
type CleanupTracker struct {
	id            uint64
	memoryManager *MemoryManager

	// Cleanup tracking
	cleanupOperations map[uint64]*CleanupOperation
	pendingCleanups   map[uint64]*CleanupOperation
	failedCleanups    map[uint64]*CleanupOperation

	// Configuration
	config CleanupConfig

	// State
	enabled          bool
	verificationMode bool

	// Statistics
	stats CleanupStats

	// Synchronization
	mutex sync.RWMutex

	// Background verification
	ctx                context.Context
	cancel             context.CancelFunc
	verificationTicker *time.Ticker
	stopVerification   chan bool
}

// CleanupConfig provides configuration for cleanup tracking
type CleanupConfig struct {
	VerificationInterval   time.Duration `json:"verification_interval"`
	VerificationTimeout    time.Duration `json:"verification_timeout"`
	MaxRetryAttempts       int           `json:"max_retry_attempts"`
	RetryBackoffMultiplier float64       `json:"retry_backoff_multiplier"`
	FailureThreshold       int           `json:"failure_threshold"`
	EnableDetailedLogging  bool          `json:"enable_detailed_logging"`
	EnableStackTraces      bool          `json:"enable_stack_traces"`
	HistoryRetention       time.Duration `json:"history_retention"`
}

// CleanupStats tracks cleanup operation statistics
type CleanupStats struct {
	TotalOperations      uint64        `json:"total_operations"`
	SuccessfulOperations uint64        `json:"successful_operations"`
	FailedOperations     uint64        `json:"failed_operations"`
	PendingOperations    uint64        `json:"pending_operations"`
	RetryOperations      uint64        `json:"retry_operations"`
	AverageCleanupTime   time.Duration `json:"average_cleanup_time"`
	MaxCleanupTime       time.Duration `json:"max_cleanup_time"`
	MinCleanupTime       time.Duration `json:"min_cleanup_time"`
	VerificationFailures uint64        `json:"verification_failures"`
	TimeoutFailures      uint64        `json:"timeout_failures"`
	StartTime            time.Time     `json:"start_time"`
	LastUpdate           time.Time     `json:"last_update"`
	mutex                sync.RWMutex
}

// CleanupOperation represents a tracked cleanup operation
type CleanupOperation struct {
	ID           uint64       `json:"id"`
	ResourceID   uint64       `json:"resource_id"`
	ResourceType string       `json:"resource_type"`
	CleanupFunc  func() error `json:"-"`
	VerifyFunc   func() bool  `json:"-"`

	// Timing
	StartTime time.Time     `json:"start_time"`
	EndTime   time.Time     `json:"end_time"`
	Duration  time.Duration `json:"duration"`
	Timeout   time.Duration `json:"timeout"`

	// Status
	Status     CleanupStatus `json:"status"`
	Error      error         `json:"error,omitempty"`
	RetryCount int           `json:"retry_count"`
	MaxRetries int           `json:"max_retries"`

	// Context
	Owner      *Owner                 `json:"-"`
	Metadata   map[string]interface{} `json:"metadata"`
	StackTrace []uintptr              `json:"-"`

	// Verification
	VerificationTime time.Time `json:"verification_time"`
	Verified         bool      `json:"verified"`

	mutex sync.RWMutex
}

// CleanupStatus represents the status of a cleanup operation
type CleanupStatus int

const (
	CleanupPending CleanupStatus = iota
	CleanupRunning
	CleanupCompleted
	CleanupFailed
	CleanupTimedOut
	CleanupRetrying
	CleanupVerified
	CleanupVerificationFailed
)

// String returns string representation of cleanup status
func (s CleanupStatus) String() string {
	switch s {
	case CleanupPending:
		return "Pending"
	case CleanupRunning:
		return "Running"
	case CleanupCompleted:
		return "Completed"
	case CleanupFailed:
		return "Failed"
	case CleanupTimedOut:
		return "TimedOut"
	case CleanupRetrying:
		return "Retrying"
	case CleanupVerified:
		return "Verified"
	case CleanupVerificationFailed:
		return "VerificationFailed"
	default:
		return "Unknown"
	}
}

// ------------------------------------
// 🏗️ Cleanup Tracker Creation
// ------------------------------------

// NewCleanupTracker creates a new cleanup tracker
func NewCleanupTracker(memoryManager *MemoryManager) *CleanupTracker {
	ctx, cancel := context.WithCancel(context.Background())

	config := DefaultCleanupConfig()

	ct := &CleanupTracker{
		id:                 atomic.AddUint64(&cleanupTrackerIdCounter, 1),
		memoryManager:      memoryManager,
		cleanupOperations:  make(map[uint64]*CleanupOperation),
		pendingCleanups:    make(map[uint64]*CleanupOperation),
		failedCleanups:     make(map[uint64]*CleanupOperation),
		config:             config,
		enabled:            true,
		verificationMode:   true,
		ctx:                ctx,
		cancel:             cancel,
		verificationTicker: time.NewTicker(config.VerificationInterval),
		stopVerification:   make(chan bool, 1),
		stats: CleanupStats{
			StartTime:      time.Now(),
			LastUpdate:     time.Now(),
			MinCleanupTime: time.Hour, // Initialize to high value
		},
	}

	// Start background verification
	go ct.backgroundVerification()

	return ct
}

// DefaultCleanupConfig returns default cleanup tracking configuration
func DefaultCleanupConfig() CleanupConfig {
	return CleanupConfig{
		VerificationInterval:   5 * time.Second,
		VerificationTimeout:    30 * time.Second,
		MaxRetryAttempts:       3,
		RetryBackoffMultiplier: 2.0,
		FailureThreshold:       10,
		EnableDetailedLogging:  false,
		EnableStackTraces:      false,
		HistoryRetention:       1 * time.Hour,
	}
}

// ------------------------------------
// 🔍 Cleanup Operation Tracking
// ------------------------------------

// TrackCleanup registers a new cleanup operation for tracking
func (ct *CleanupTracker) TrackCleanup(resourceID uint64, resourceType string, cleanupFunc func() error, verifyFunc func() bool, owner *Owner, metadata map[string]interface{}) *CleanupOperation {
	if !ct.enabled {
		return nil
	}

	operation := &CleanupOperation{
		ID:           atomic.AddUint64(&cleanupTrackerIdCounter, 1),
		ResourceID:   resourceID,
		ResourceType: resourceType,
		CleanupFunc:  cleanupFunc,
		VerifyFunc:   verifyFunc,
		StartTime:    time.Now(),
		Timeout:      ct.config.VerificationTimeout,
		Status:       CleanupPending,
		MaxRetries:   ct.config.MaxRetryAttempts,
		Owner:        owner,
		Metadata:     metadata,
	}

	// Capture stack trace if enabled
	if ct.config.EnableStackTraces {
		operation.StackTrace = make([]uintptr, 32)
		n := runtime.Callers(2, operation.StackTrace)
		operation.StackTrace = operation.StackTrace[:n]
	}

	ct.mutex.Lock()
	ct.cleanupOperations[operation.ID] = operation
	ct.pendingCleanups[operation.ID] = operation
	ct.mutex.Unlock()

	// Update statistics
	ct.updateStats(func(stats *CleanupStats) {
		atomic.AddUint64(&stats.TotalOperations, 1)
		atomic.AddUint64(&stats.PendingOperations, 1)
	})

	return operation
}

// ExecuteCleanup executes a cleanup operation with tracking
func (ct *CleanupTracker) ExecuteCleanup(operation *CleanupOperation) error {
	if operation == nil || operation.CleanupFunc == nil {
		return fmt.Errorf("invalid cleanup operation")
	}

	operation.mutex.Lock()
	if operation.Status != CleanupPending && operation.Status != CleanupRetrying {
		operation.mutex.Unlock()
		return fmt.Errorf("cleanup operation not in pending state: %s", operation.Status)
	}

	operation.Status = CleanupRunning
	operation.StartTime = time.Now()
	operation.mutex.Unlock()

	// Execute cleanup with timeout
	done := make(chan error, 1)
	go func() {
		defer func() {
			if r := recover(); r != nil {
				done <- fmt.Errorf("cleanup panicked: %v", r)
			}
		}()
		done <- operation.CleanupFunc()
	}()

	var err error
	select {
	case err = <-done:
		// Cleanup completed
	case <-time.After(operation.Timeout):
		err = fmt.Errorf("cleanup timed out after %v", operation.Timeout)
		ct.updateStats(func(stats *CleanupStats) {
			atomic.AddUint64(&stats.TimeoutFailures, 1)
		})
	}

	// Update operation status
	operation.mutex.Lock()
	operation.EndTime = time.Now()
	operation.Duration = operation.EndTime.Sub(operation.StartTime)
	operation.Error = err

	if err != nil {
		operation.Status = CleanupFailed
		ct.mutex.Lock()
		ct.failedCleanups[operation.ID] = operation
		delete(ct.pendingCleanups, operation.ID)
		ct.mutex.Unlock()

		ct.updateStats(func(stats *CleanupStats) {
			atomic.AddUint64(&stats.FailedOperations, 1)
			atomic.AddUint64(&stats.PendingOperations, ^uint64(0)) // Decrement
		})
	} else {
		operation.Status = CleanupCompleted
		ct.mutex.Lock()
		delete(ct.pendingCleanups, operation.ID)
		ct.mutex.Unlock()

		ct.updateStats(func(stats *CleanupStats) {
			atomic.AddUint64(&stats.SuccessfulOperations, 1)
			atomic.AddUint64(&stats.PendingOperations, ^uint64(0)) // Decrement

			// Update timing statistics
			if operation.Duration < stats.MinCleanupTime {
				stats.MinCleanupTime = operation.Duration
			}
			if operation.Duration > stats.MaxCleanupTime {
				stats.MaxCleanupTime = operation.Duration
			}

			// Update average (simple moving average)
			totalOps := atomic.LoadUint64(&stats.SuccessfulOperations)
			if totalOps > 0 {
				stats.AverageCleanupTime = time.Duration(
					(int64(stats.AverageCleanupTime)*int64(totalOps-1) + int64(operation.Duration)) / int64(totalOps),
				)
			}
		})
	}
	operation.mutex.Unlock()

	// Schedule verification if cleanup succeeded and verification function exists
	if err == nil && operation.VerifyFunc != nil && ct.verificationMode {
		go ct.scheduleVerification(operation)
	}

	return err
}

// RecordCleanup records the result of a cleanup operation
func (ct *CleanupTracker) RecordCleanup(resourceID uint64, success bool, err error) {
	ct.mutex.RLock()
	var operation *CleanupOperation
	for _, op := range ct.cleanupOperations {
		if op.ResourceID == resourceID {
			operation = op
			break
		}
	}
	ct.mutex.RUnlock()

	if operation == nil {
		// Create a minimal operation record for external cleanups
		operation = &CleanupOperation{
			ID:         atomic.AddUint64(&cleanupTrackerIdCounter, 1),
			ResourceID: resourceID,
			EndTime:    time.Now(),
			Error:      err,
		}

		if success {
			operation.Status = CleanupCompleted
		} else {
			operation.Status = CleanupFailed
		}

		ct.mutex.Lock()
		ct.cleanupOperations[operation.ID] = operation
		ct.mutex.Unlock()
	}

	// Update statistics
	if success {
		ct.updateStats(func(stats *CleanupStats) {
			atomic.AddUint64(&stats.SuccessfulOperations, 1)
		})
	} else {
		ct.updateStats(func(stats *CleanupStats) {
			atomic.AddUint64(&stats.FailedOperations, 1)
		})
	}
}

// ------------------------------------
// ✅ Cleanup Verification
// ------------------------------------

// scheduleVerification schedules verification for a completed cleanup
func (ct *CleanupTracker) scheduleVerification(operation *CleanupOperation) {
	// Wait a brief moment for cleanup effects to propagate
	time.Sleep(100 * time.Millisecond)

	ct.verifyCleanup(operation)
}

// verifyCleanup verifies that a cleanup operation was successful
func (ct *CleanupTracker) verifyCleanup(operation *CleanupOperation) {
	if operation == nil || operation.VerifyFunc == nil {
		return
	}

	operation.mutex.Lock()
	if operation.Status != CleanupCompleted {
		operation.mutex.Unlock()
		return
	}
	operation.mutex.Unlock()

	// Perform verification
	verified := false
	verificationStart := time.Now()

	defer func() {
		if r := recover(); r != nil {
			if ct.config.EnableDetailedLogging {
				fmt.Printf("🚨 Cleanup verification panicked for resource %d: %v\n", operation.ResourceID, r)
			}
		}
	}()

	verified = operation.VerifyFunc()

	operation.mutex.Lock()
	operation.VerificationTime = time.Now()
	operation.Verified = verified

	if verified {
		operation.Status = CleanupVerified
	} else {
		operation.Status = CleanupVerificationFailed
		ct.updateStats(func(stats *CleanupStats) {
			atomic.AddUint64(&stats.VerificationFailures, 1)
		})

		if ct.config.EnableDetailedLogging {
			fmt.Printf("🚨 Cleanup verification failed for resource %d (type: %s)\n",
				operation.ResourceID, operation.ResourceType)
		}
	}
	operation.mutex.Unlock()

	verificationDuration := time.Since(verificationStart)
	if ct.config.EnableDetailedLogging {
		fmt.Printf("✅ Verified cleanup for resource %d in %v (verified: %t)\n",
			operation.ResourceID, verificationDuration, verified)
	}
}

// ------------------------------------
// 🔄 Background Verification
// ------------------------------------

// backgroundVerification runs continuous cleanup verification
func (ct *CleanupTracker) backgroundVerification() {
	for {
		select {
		case <-ct.ctx.Done():
			return
		case <-ct.stopVerification:
			return
		case <-ct.verificationTicker.C:
			ct.performBackgroundVerification()
		}
	}
}

// performBackgroundVerification performs periodic verification tasks
func (ct *CleanupTracker) performBackgroundVerification() {
	if !ct.verificationMode {
		return
	}

	// Verify pending operations
	ct.verifyPendingOperations()

	// Retry failed operations
	ct.retryFailedOperations()

	// Clean up old operations
	ct.cleanupOldOperations()

	// Update statistics
	ct.updateStats(func(stats *CleanupStats) {
		stats.LastUpdate = time.Now()
	})
}

// verifyPendingOperations verifies operations that haven't been verified yet
func (ct *CleanupTracker) verifyPendingOperations() {
	ct.mutex.RLock()
	var toVerify []*CleanupOperation

	for _, operation := range ct.cleanupOperations {
		operation.mutex.RLock()
		if operation.Status == CleanupCompleted && !operation.Verified && operation.VerifyFunc != nil {
			toVerify = append(toVerify, operation)
		}
		operation.mutex.RUnlock()
	}
	ct.mutex.RUnlock()

	for _, operation := range toVerify {
		ct.verifyCleanup(operation)
	}
}

// retryFailedOperations retries failed cleanup operations
func (ct *CleanupTracker) retryFailedOperations() {
	ct.mutex.RLock()
	var toRetry []*CleanupOperation

	for _, operation := range ct.failedCleanups {
		operation.mutex.RLock()
		if operation.RetryCount < operation.MaxRetries &&
			time.Since(operation.EndTime) > ct.calculateRetryDelay(operation.RetryCount) {
			toRetry = append(toRetry, operation)
		}
		operation.mutex.RUnlock()
	}
	ct.mutex.RUnlock()

	for _, operation := range toRetry {
		ct.retryCleanupOperation(operation)
	}
}

// retryCleanupOperation retries a failed cleanup operation
func (ct *CleanupTracker) retryCleanupOperation(operation *CleanupOperation) {
	operation.mutex.Lock()
	operation.RetryCount++
	operation.Status = CleanupRetrying
	operation.mutex.Unlock()

	ct.updateStats(func(stats *CleanupStats) {
		atomic.AddUint64(&stats.RetryOperations, 1)
	})

	if ct.config.EnableDetailedLogging {
		fmt.Printf("🔄 Retrying cleanup for resource %d (attempt %d/%d)\n",
			operation.ResourceID, operation.RetryCount, operation.MaxRetries)
	}

	// Execute the retry
	go func() {
		err := ct.ExecuteCleanup(operation)
		if err != nil && ct.config.EnableDetailedLogging {
			fmt.Printf("🚨 Cleanup retry failed for resource %d: %v\n", operation.ResourceID, err)
		}
	}()
}

// calculateRetryDelay calculates the delay before retrying based on attempt count
func (ct *CleanupTracker) calculateRetryDelay(retryCount int) time.Duration {
	baseDelay := time.Second
	multiplier := ct.config.RetryBackoffMultiplier

	delay := float64(baseDelay) * pow(multiplier, float64(retryCount))
	return time.Duration(delay)
}

// pow calculates x^y for float64 (simple implementation)
func pow(x, y float64) float64 {
	if y == 0 {
		return 1
	}
	result := x
	for i := 1; i < int(y); i++ {
		result *= x
	}
	return result
}

// cleanupOldOperations removes old completed operations from memory
func (ct *CleanupTracker) cleanupOldOperations() {
	ct.mutex.Lock()
	defer ct.mutex.Unlock()

	cutoff := time.Now().Add(-ct.config.HistoryRetention)
	var toRemove []uint64

	for id, operation := range ct.cleanupOperations {
		operation.mutex.RLock()
		if (operation.Status == CleanupVerified || operation.Status == CleanupCompleted) &&
			operation.EndTime.Before(cutoff) {
			toRemove = append(toRemove, id)
		}
		operation.mutex.RUnlock()
	}

	for _, id := range toRemove {
		delete(ct.cleanupOperations, id)
		delete(ct.pendingCleanups, id)
		delete(ct.failedCleanups, id)
	}

	if len(toRemove) > 0 && ct.config.EnableDetailedLogging {
		fmt.Printf("🧹 Cleaned up %d old cleanup operations\n", len(toRemove))
	}
}

// ------------------------------------
// 📊 Statistics and Reporting
// ------------------------------------

// GetStats returns current cleanup statistics
func (ct *CleanupTracker) GetStats() CleanupStats {
	ct.stats.mutex.RLock()
	defer ct.stats.mutex.RUnlock()

	stats := ct.stats
	stats.LastUpdate = time.Now()
	return stats
}

// GetReport returns a detailed cleanup tracking report
func (ct *CleanupTracker) GetReport() map[string]interface{} {
	ct.mutex.RLock()
	defer ct.mutex.RUnlock()

	report := map[string]interface{}{
		"tracker_id":         ct.id,
		"enabled":            ct.enabled,
		"verification_mode":  ct.verificationMode,
		"config":             ct.config,
		"stats":              ct.GetStats(),
		"total_operations":   len(ct.cleanupOperations),
		"pending_operations": len(ct.pendingCleanups),
		"failed_operations":  len(ct.failedCleanups),
	}

	// Add operation breakdown by status
	statusBreakdown := make(map[string]int)
	typeBreakdown := make(map[string]int)

	for _, operation := range ct.cleanupOperations {
		operation.mutex.RLock()
		statusBreakdown[operation.Status.String()]++
		typeBreakdown[operation.ResourceType]++
		operation.mutex.RUnlock()
	}

	report["operations_by_status"] = statusBreakdown
	report["operations_by_type"] = typeBreakdown

	return report
}

// GetFailedOperations returns all failed cleanup operations
func (ct *CleanupTracker) GetFailedOperations() []*CleanupOperation {
	ct.mutex.RLock()
	defer ct.mutex.RUnlock()

	var failed []*CleanupOperation
	for _, operation := range ct.failedCleanups {
		failed = append(failed, operation)
	}

	return failed
}

// GetPendingOperations returns all pending cleanup operations
func (ct *CleanupTracker) GetPendingOperations() []*CleanupOperation {
	ct.mutex.RLock()
	defer ct.mutex.RUnlock()

	var pending []*CleanupOperation
	for _, operation := range ct.pendingCleanups {
		pending = append(pending, operation)
	}

	return pending
}

// ------------------------------------
// 🛠️ Configuration and Control
// ------------------------------------

// SetEnabled enables or disables cleanup tracking
func (ct *CleanupTracker) SetEnabled(enabled bool) {
	ct.mutex.Lock()
	defer ct.mutex.Unlock()
	ct.enabled = enabled
}

// SetVerificationMode enables or disables cleanup verification
func (ct *CleanupTracker) SetVerificationMode(enabled bool) {
	ct.mutex.Lock()
	defer ct.mutex.Unlock()
	ct.verificationMode = enabled
}

// UpdateConfig updates the cleanup tracker configuration
func (ct *CleanupTracker) UpdateConfig(config CleanupConfig) {
	ct.mutex.Lock()
	defer ct.mutex.Unlock()

	ct.config = config

	// Update ticker interval if changed
	if ct.verificationTicker != nil {
		ct.verificationTicker.Stop()
		ct.verificationTicker = time.NewTicker(config.VerificationInterval)
	}
}

// updateStats safely updates cleanup statistics
func (ct *CleanupTracker) updateStats(fn func(*CleanupStats)) {
	ct.stats.mutex.Lock()
	defer ct.stats.mutex.Unlock()
	fn(&ct.stats)
}

// ------------------------------------
// 🧹 Disposal
// ------------------------------------

// Dispose cleans up the cleanup tracker and all its resources
func (ct *CleanupTracker) Dispose() {
	ct.mutex.Lock()
	defer ct.mutex.Unlock()

	// Stop background verification
	if ct.cancel != nil {
		ct.cancel()
	}

	if ct.verificationTicker != nil {
		ct.verificationTicker.Stop()
	}

	select {
	case ct.stopVerification <- true:
	default:
	}

	// Clear all operations
	ct.cleanupOperations = make(map[uint64]*CleanupOperation)
	ct.pendingCleanups = make(map[uint64]*CleanupOperation)
	ct.failedCleanups = make(map[uint64]*CleanupOperation)

	ct.enabled = false
	ct.verificationMode = false
}
