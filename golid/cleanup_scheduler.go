// cleanup_scheduler.go
// Coordinated cleanup scheduling and execution system

package golid

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"sync/atomic"
	"time"
)

// ------------------------------------
// 📅 Cleanup Scheduler Core
// ------------------------------------

var cleanupSchedulerIdCounter uint64

// CleanupScheduler provides coordinated cleanup scheduling and execution
type CleanupScheduler struct {
	id            uint64
	memoryManager *MemoryManager

	// Scheduling queues
	immediateQueue []*ScheduledCleanup
	delayedQueue   []*ScheduledCleanup
	periodicQueue  []*ScheduledCleanup

	// Execution management
	workers     []*CleanupWorker
	workQueue   chan *ScheduledCleanup
	resultQueue chan *CleanupResult

	// Configuration
	config SchedulerConfig

	// State
	enabled bool
	running bool
	paused  bool

	// Statistics
	stats CleanupSchedulerStats

	// Background scheduling
	ctx            context.Context
	cancel         context.CancelFunc
	scheduleTicker *time.Ticker
	wg             sync.WaitGroup // Track background goroutines
	stopScheduling chan bool

	// Synchronization
	mutex       sync.RWMutex
	workerMutex sync.RWMutex
}

// SchedulerConfig provides configuration for cleanup scheduling
type SchedulerConfig struct {
	WorkerCount           int           `json:"worker_count"`
	QueueSize             int           `json:"queue_size"`
	ScheduleInterval      time.Duration `json:"schedule_interval"`
	MaxConcurrentCleanups int           `json:"max_concurrent_cleanups"`
	CleanupTimeout        time.Duration `json:"cleanup_timeout"`
	RetryAttempts         int           `json:"retry_attempts"`
	RetryDelay            time.Duration `json:"retry_delay"`
	PriorityLevels        int           `json:"priority_levels"`
	EnableBatching        bool          `json:"enable_batching"`
	BatchSize             int           `json:"batch_size"`
	EnablePrioritization  bool          `json:"enable_prioritization"`
}

// CleanupSchedulerStats tracks cleanup scheduler statistics
type CleanupSchedulerStats struct {
	TotalScheduled       uint64        `json:"total_scheduled"`
	TotalExecuted        uint64        `json:"total_executed"`
	TotalSuccessful      uint64        `json:"total_successful"`
	TotalFailed          uint64        `json:"total_failed"`
	TotalRetried         uint64        `json:"total_retried"`
	QueuedImmediate      uint64        `json:"queued_immediate"`
	QueuedDelayed        uint64        `json:"queued_delayed"`
	QueuedPeriodic       uint64        `json:"queued_periodic"`
	ActiveWorkers        uint64        `json:"active_workers"`
	AverageExecutionTime time.Duration `json:"average_execution_time"`
	MaxExecutionTime     time.Duration `json:"max_execution_time"`
	ThroughputPerSecond  float64       `json:"throughput_per_second"`
	StartTime            time.Time     `json:"start_time"`
	LastScheduleTime     time.Time     `json:"last_schedule_time"`
	mutex                sync.RWMutex
}

// ScheduledCleanup represents a scheduled cleanup operation
type ScheduledCleanup struct {
	ID           uint64       `json:"id"`
	ResourceID   uint64       `json:"resource_id"`
	ResourceType string       `json:"resource_type"`
	CleanupFunc  func() error `json:"-"`
	VerifyFunc   func() bool  `json:"-"`

	// Scheduling
	ScheduleType ScheduleType    `json:"schedule_type"`
	Priority     CleanupPriority `json:"priority"`
	ScheduledAt  time.Time       `json:"scheduled_at"`
	ExecuteAt    time.Time       `json:"execute_at"`
	Delay        time.Duration   `json:"delay"`
	Interval     time.Duration   `json:"interval"`

	// Execution
	Status       CleanupExecutionStatus `json:"status"`
	AttemptCount int                    `json:"attempt_count"`
	MaxAttempts  int                    `json:"max_attempts"`
	LastAttempt  time.Time              `json:"last_attempt"`
	NextRetry    time.Time              `json:"next_retry"`

	// Context
	Owner        *Owner                 `json:"-"`
	Dependencies []uint64               `json:"dependencies"`
	Dependents   []uint64               `json:"dependents"`
	Metadata     map[string]interface{} `json:"metadata"`

	// Results
	ExecutionTime time.Duration `json:"execution_time"`
	Error         error         `json:"error,omitempty"`

	mutex sync.RWMutex
}

// ScheduleType defines types of cleanup scheduling
type ScheduleType int

const (
	ScheduleImmediate ScheduleType = iota
	ScheduleDelayed
	SchedulePeriodic
	ScheduleConditional
)

// String returns string representation of schedule type
func (s ScheduleType) String() string {
	switch s {
	case ScheduleImmediate:
		return "Immediate"
	case ScheduleDelayed:
		return "Delayed"
	case SchedulePeriodic:
		return "Periodic"
	case ScheduleConditional:
		return "Conditional"
	default:
		return "Unknown"
	}
}

// CleanupPriority defines priority levels for cleanup operations
type CleanupPriority int

const (
	PriorityLow CleanupPriority = iota
	PriorityNormal
	PriorityHigh
	PriorityCritical
)

// String returns string representation of cleanup priority
func (p CleanupPriority) String() string {
	switch p {
	case PriorityLow:
		return "Low"
	case PriorityNormal:
		return "Normal"
	case PriorityHigh:
		return "High"
	case PriorityCritical:
		return "Critical"
	default:
		return "Unknown"
	}
}

// CleanupExecutionStatus defines execution status of cleanup operations
type CleanupExecutionStatus int

const (
	StatusPending CleanupExecutionStatus = iota
	StatusQueued
	StatusExecuting
	StatusCompleted
	StatusFailed
	StatusRetrying
	StatusCancelled
)

// String returns string representation of execution status
func (s CleanupExecutionStatus) String() string {
	switch s {
	case StatusPending:
		return "Pending"
	case StatusQueued:
		return "Queued"
	case StatusExecuting:
		return "Executing"
	case StatusCompleted:
		return "Completed"
	case StatusFailed:
		return "Failed"
	case StatusRetrying:
		return "Retrying"
	case StatusCancelled:
		return "Cancelled"
	default:
		return "Unknown"
	}
}

// CleanupWorker executes cleanup operations
type CleanupWorker struct {
	id          uint64
	scheduler   *CleanupScheduler
	workQueue   chan *ScheduledCleanup
	resultQueue chan *CleanupResult
	active      bool
	ctx         context.Context
	cancel      context.CancelFunc
	mutex       sync.RWMutex
}

// CleanupResult represents the result of a cleanup execution
type CleanupResult struct {
	CleanupID     uint64        `json:"cleanup_id"`
	Success       bool          `json:"success"`
	Error         error         `json:"error,omitempty"`
	ExecutionTime time.Duration `json:"execution_time"`
	Verified      bool          `json:"verified"`
	Timestamp     time.Time     `json:"timestamp"`
}

// ------------------------------------
// 🏗️ Cleanup Scheduler Creation
// ------------------------------------

// NewCleanupScheduler creates a new cleanup scheduler
func NewCleanupScheduler(memoryManager *MemoryManager) *CleanupScheduler {
	ctx, cancel := context.WithCancel(context.Background())

	config := DefaultSchedulerConfig()

	cs := &CleanupScheduler{
		id:             atomic.AddUint64(&cleanupSchedulerIdCounter, 1),
		memoryManager:  memoryManager,
		immediateQueue: make([]*ScheduledCleanup, 0),
		delayedQueue:   make([]*ScheduledCleanup, 0),
		periodicQueue:  make([]*ScheduledCleanup, 0),
		workers:        make([]*CleanupWorker, 0),
		workQueue:      make(chan *ScheduledCleanup, config.QueueSize),
		resultQueue:    make(chan *CleanupResult, config.QueueSize),
		config:         config,
		enabled:        true,
		running:        false,
		paused:         false,
		ctx:            ctx,
		cancel:         cancel,
		scheduleTicker: time.NewTicker(config.ScheduleInterval),
		stopScheduling: make(chan bool, 1),
		stats: CleanupSchedulerStats{
			StartTime: time.Now(),
		},
	}

	// Create workers
	cs.createWorkers()

	// Start background scheduling
	cs.wg.Add(1)
	go cs.backgroundScheduling()

	// Start result processing
	cs.wg.Add(1)
	go cs.processResults()

	return cs
}

// DefaultSchedulerConfig returns default scheduler configuration
func DefaultSchedulerConfig() SchedulerConfig {
	return SchedulerConfig{
		WorkerCount:           4,
		QueueSize:             1000,
		ScheduleInterval:      time.Second,
		MaxConcurrentCleanups: 10,
		CleanupTimeout:        30 * time.Second,
		RetryAttempts:         3,
		RetryDelay:            time.Second,
		PriorityLevels:        4,
		EnableBatching:        true,
		BatchSize:             10,
		EnablePrioritization:  true,
	}
}

// createWorkers creates cleanup worker goroutines
func (cs *CleanupScheduler) createWorkers() {
	cs.workerMutex.Lock()
	defer cs.workerMutex.Unlock()

	for i := 0; i < cs.config.WorkerCount; i++ {
		worker := cs.createWorker()
		cs.workers = append(cs.workers, worker)
		go worker.run()
	}
}

// createWorker creates a single cleanup worker
func (cs *CleanupScheduler) createWorker() *CleanupWorker {
	ctx, cancel := context.WithCancel(cs.ctx)

	return &CleanupWorker{
		id:          atomic.AddUint64(&cleanupSchedulerIdCounter, 1),
		scheduler:   cs,
		workQueue:   cs.workQueue,
		resultQueue: cs.resultQueue,
		active:      true,
		ctx:         ctx,
		cancel:      cancel,
	}
}

// ------------------------------------
// 📅 Cleanup Scheduling
// ------------------------------------

// ScheduleImmediate schedules a cleanup for immediate execution
func (cs *CleanupScheduler) ScheduleImmediate(resourceID uint64, resourceType string, cleanupFunc func() error, verifyFunc func() bool, priority CleanupPriority, owner *Owner, metadata map[string]interface{}) *ScheduledCleanup {
	cleanup := &ScheduledCleanup{
		ID:           atomic.AddUint64(&cleanupSchedulerIdCounter, 1),
		ResourceID:   resourceID,
		ResourceType: resourceType,
		CleanupFunc:  cleanupFunc,
		VerifyFunc:   verifyFunc,
		ScheduleType: ScheduleImmediate,
		Priority:     priority,
		ScheduledAt:  time.Now(),
		ExecuteAt:    time.Now(),
		Status:       StatusPending,
		MaxAttempts:  cs.config.RetryAttempts,
		Owner:        owner,
		Dependencies: make([]uint64, 0),
		Dependents:   make([]uint64, 0),
		Metadata:     metadata,
	}

	cs.addToQueue(cleanup)
	return cleanup
}

// ScheduleDelayed schedules a cleanup for delayed execution
func (cs *CleanupScheduler) ScheduleDelayed(resourceID uint64, resourceType string, cleanupFunc func() error, verifyFunc func() bool, delay time.Duration, priority CleanupPriority, owner *Owner, metadata map[string]interface{}) *ScheduledCleanup {
	cleanup := &ScheduledCleanup{
		ID:           atomic.AddUint64(&cleanupSchedulerIdCounter, 1),
		ResourceID:   resourceID,
		ResourceType: resourceType,
		CleanupFunc:  cleanupFunc,
		VerifyFunc:   verifyFunc,
		ScheduleType: ScheduleDelayed,
		Priority:     priority,
		ScheduledAt:  time.Now(),
		ExecuteAt:    time.Now().Add(delay),
		Delay:        delay,
		Status:       StatusPending,
		MaxAttempts:  cs.config.RetryAttempts,
		Owner:        owner,
		Dependencies: make([]uint64, 0),
		Dependents:   make([]uint64, 0),
		Metadata:     metadata,
	}

	cs.addToQueue(cleanup)
	return cleanup
}

// SchedulePeriodic schedules a cleanup for periodic execution
func (cs *CleanupScheduler) SchedulePeriodic(resourceID uint64, resourceType string, cleanupFunc func() error, verifyFunc func() bool, interval time.Duration, priority CleanupPriority, owner *Owner, metadata map[string]interface{}) *ScheduledCleanup {
	cleanup := &ScheduledCleanup{
		ID:           atomic.AddUint64(&cleanupSchedulerIdCounter, 1),
		ResourceID:   resourceID,
		ResourceType: resourceType,
		CleanupFunc:  cleanupFunc,
		VerifyFunc:   verifyFunc,
		ScheduleType: SchedulePeriodic,
		Priority:     priority,
		ScheduledAt:  time.Now(),
		ExecuteAt:    time.Now().Add(interval),
		Interval:     interval,
		Status:       StatusPending,
		MaxAttempts:  cs.config.RetryAttempts,
		Owner:        owner,
		Dependencies: make([]uint64, 0),
		Dependents:   make([]uint64, 0),
		Metadata:     metadata,
	}

	cs.addToQueue(cleanup)
	return cleanup
}

// addToQueue adds a cleanup to the appropriate queue
func (cs *CleanupScheduler) addToQueue(cleanup *ScheduledCleanup) {
	if !cs.enabled {
		return
	}

	cs.mutex.Lock()
	defer cs.mutex.Unlock()

	switch cleanup.ScheduleType {
	case ScheduleImmediate:
		cs.immediateQueue = append(cs.immediateQueue, cleanup)
		cs.updateStats(func(stats *CleanupSchedulerStats) {
			atomic.AddUint64(&stats.QueuedImmediate, 1)
		})
	case ScheduleDelayed:
		cs.delayedQueue = append(cs.delayedQueue, cleanup)
		cs.updateStats(func(stats *CleanupSchedulerStats) {
			atomic.AddUint64(&stats.QueuedDelayed, 1)
		})
	case SchedulePeriodic:
		cs.periodicQueue = append(cs.periodicQueue, cleanup)
		cs.updateStats(func(stats *CleanupSchedulerStats) {
			atomic.AddUint64(&stats.QueuedPeriodic, 1)
		})
	}

	cleanup.Status = StatusQueued

	cs.updateStats(func(stats *CleanupSchedulerStats) {
		atomic.AddUint64(&stats.TotalScheduled, 1)
	})
}

// ------------------------------------
// 🔄 Background Scheduling
// ------------------------------------

// backgroundScheduling runs continuous cleanup scheduling
func (cs *CleanupScheduler) backgroundScheduling() {
	defer cs.wg.Done()
	for {
		select {
		case <-cs.ctx.Done():
			return
		case <-cs.stopScheduling:
			return
		case <-cs.scheduleTicker.C:
			if cs.enabled && cs.running && !cs.paused {
				cs.processScheduledCleanups()
			}
		}
	}
}

// processScheduledCleanups processes all scheduled cleanups
func (cs *CleanupScheduler) processScheduledCleanups() {
	cs.mutex.Lock()
	defer cs.mutex.Unlock()

	now := time.Now()
	cs.updateStats(func(stats *CleanupSchedulerStats) {
		stats.LastScheduleTime = now
	})

	// Process immediate queue
	cs.processImmediateQueue()

	// Process delayed queue
	cs.processDelayedQueue(now)

	// Process periodic queue
	cs.processPeriodicQueue(now)
}

// processImmediateQueue processes the immediate cleanup queue
func (cs *CleanupScheduler) processImmediateQueue() {
	if len(cs.immediateQueue) == 0 {
		return
	}

	// Sort by priority if prioritization is enabled
	if cs.config.EnablePrioritization {
		sort.Slice(cs.immediateQueue, func(i, j int) bool {
			return cs.immediateQueue[i].Priority > cs.immediateQueue[j].Priority
		})
	}

	// Process cleanups
	var remaining []*ScheduledCleanup
	processed := 0

	for _, cleanup := range cs.immediateQueue {
		if processed >= cs.config.MaxConcurrentCleanups {
			remaining = append(remaining, cleanup)
			continue
		}

		if cs.executeCleanup(cleanup) {
			processed++
		} else {
			remaining = append(remaining, cleanup)
		}
	}

	cs.immediateQueue = remaining
}

// processDelayedQueue processes the delayed cleanup queue
func (cs *CleanupScheduler) processDelayedQueue(now time.Time) {
	if len(cs.delayedQueue) == 0 {
		return
	}

	var remaining []*ScheduledCleanup
	processed := 0

	for _, cleanup := range cs.delayedQueue {
		if now.Before(cleanup.ExecuteAt) {
			remaining = append(remaining, cleanup)
			continue
		}

		if processed >= cs.config.MaxConcurrentCleanups {
			remaining = append(remaining, cleanup)
			continue
		}

		if cs.executeCleanup(cleanup) {
			processed++
		} else {
			remaining = append(remaining, cleanup)
		}
	}

	cs.delayedQueue = remaining
}

// processPeriodicQueue processes the periodic cleanup queue
func (cs *CleanupScheduler) processPeriodicQueue(now time.Time) {
	if len(cs.periodicQueue) == 0 {
		return
	}

	processed := 0

	for _, cleanup := range cs.periodicQueue {
		if now.Before(cleanup.ExecuteAt) {
			continue
		}

		if processed >= cs.config.MaxConcurrentCleanups {
			break
		}

		if cs.executeCleanup(cleanup) {
			processed++
			// Reschedule for next interval
			cleanup.ExecuteAt = now.Add(cleanup.Interval)
			cleanup.Status = StatusQueued
		}
	}
}

// executeCleanup attempts to execute a cleanup operation
func (cs *CleanupScheduler) executeCleanup(cleanup *ScheduledCleanup) bool {
	select {
	case cs.workQueue <- cleanup:
		cleanup.Status = StatusExecuting
		cs.updateStats(func(stats *CleanupSchedulerStats) {
			atomic.AddUint64(&stats.TotalExecuted, 1)
		})
		return true
	default:
		// Work queue is full
		return false
	}
}

// ------------------------------------
// 👷 Worker Management
// ------------------------------------

// run executes the worker's main loop
func (w *CleanupWorker) run() {
	for {
		select {
		case <-w.ctx.Done():
			return
		case cleanup := <-w.workQueue:
			if cleanup != nil {
				w.executeCleanup(cleanup)
			}
		}
	}
}

// executeCleanup executes a single cleanup operation
func (w *CleanupWorker) executeCleanup(cleanup *ScheduledCleanup) {
	w.mutex.Lock()
	w.active = true
	w.mutex.Unlock()

	defer func() {
		w.mutex.Lock()
		w.active = false
		w.mutex.Unlock()
	}()

	start := time.Now()
	var err error

	cleanup.mutex.Lock()
	cleanup.AttemptCount++
	cleanup.LastAttempt = start
	cleanup.mutex.Unlock()

	// Execute cleanup with timeout
	done := make(chan error, 1)
	go func() {
		defer func() {
			if r := recover(); r != nil {
				done <- fmt.Errorf("cleanup panicked: %v", r)
			}
		}()
		done <- cleanup.CleanupFunc()
	}()

	select {
	case err = <-done:
		// Cleanup completed
	case <-time.After(w.scheduler.config.CleanupTimeout):
		err = fmt.Errorf("cleanup timed out after %v", w.scheduler.config.CleanupTimeout)
	case <-w.ctx.Done():
		err = fmt.Errorf("cleanup cancelled")
	}

	executionTime := time.Since(start)

	cleanup.mutex.Lock()
	cleanup.ExecutionTime = executionTime
	cleanup.Error = err
	cleanup.mutex.Unlock()

	// Create result
	result := &CleanupResult{
		CleanupID:     cleanup.ID,
		Success:       err == nil,
		Error:         err,
		ExecutionTime: executionTime,
		Timestamp:     time.Now(),
	}

	// Verify cleanup if verification function is provided
	if err == nil && cleanup.VerifyFunc != nil {
		result.Verified = cleanup.VerifyFunc()
		if !result.Verified {
			result.Success = false
			result.Error = fmt.Errorf("cleanup verification failed")
		}
	}

	// Update cleanup status
	cleanup.mutex.Lock()
	if result.Success {
		cleanup.Status = StatusCompleted
	} else {
		if cleanup.AttemptCount < cleanup.MaxAttempts {
			cleanup.Status = StatusRetrying
			cleanup.NextRetry = time.Now().Add(w.scheduler.config.RetryDelay)
		} else {
			cleanup.Status = StatusFailed
		}
	}
	cleanup.mutex.Unlock()

	// Send result
	select {
	case w.resultQueue <- result:
	default:
		// Result queue is full, drop result
	}
}

// ------------------------------------
// 📊 Result Processing
// ------------------------------------

// processResults processes cleanup execution results
func (cs *CleanupScheduler) processResults() {
	defer cs.wg.Done()
	for {
		select {
		case <-cs.ctx.Done():
			return
		case result := <-cs.resultQueue:
			cs.handleResult(result)
		}
	}
}

// handleResult handles a single cleanup result
func (cs *CleanupScheduler) handleResult(result *CleanupResult) {
	// Guard against nil results from closed channels
	if result == nil {
		return
	}

	if result.Success {
		cs.updateStats(func(stats *CleanupSchedulerStats) {
			atomic.AddUint64(&stats.TotalSuccessful, 1)

			// Update execution time statistics
			if result.ExecutionTime > stats.MaxExecutionTime {
				stats.MaxExecutionTime = result.ExecutionTime
			}

			// Update average execution time
			total := atomic.LoadUint64(&stats.TotalSuccessful)
			if total > 0 {
				stats.AverageExecutionTime = time.Duration(
					(int64(stats.AverageExecutionTime)*int64(total-1) + int64(result.ExecutionTime)) / int64(total),
				)
			}
		})
	} else {
		cs.updateStats(func(stats *CleanupSchedulerStats) {
			atomic.AddUint64(&stats.TotalFailed, 1)
		})

		// Schedule retry if needed
		cs.scheduleRetry(result.CleanupID)
	}
}

// scheduleRetry schedules a retry for a failed cleanup
func (cs *CleanupScheduler) scheduleRetry(cleanupID uint64) {
	cs.mutex.Lock()
	defer cs.mutex.Unlock()

	// Find the cleanup in queues
	var cleanup *ScheduledCleanup

	// Check immediate queue
	for _, c := range cs.immediateQueue {
		if c.ID == cleanupID {
			cleanup = c
			break
		}
	}

	// Check delayed queue
	if cleanup == nil {
		for _, c := range cs.delayedQueue {
			if c.ID == cleanupID {
				cleanup = c
				break
			}
		}
	}

	if cleanup != nil && cleanup.Status == StatusRetrying {
		// Move to delayed queue for retry
		cleanup.ExecuteAt = cleanup.NextRetry
		cleanup.ScheduleType = ScheduleDelayed

		// Remove from current queue and add to delayed queue
		cs.removeFromQueue(cleanup)
		cs.delayedQueue = append(cs.delayedQueue, cleanup)

		cs.updateStats(func(stats *CleanupSchedulerStats) {
			atomic.AddUint64(&stats.TotalRetried, 1)
		})
	}
}

// removeFromQueue removes a cleanup from its current queue
func (cs *CleanupScheduler) removeFromQueue(target *ScheduledCleanup) {
	// Remove from immediate queue
	for i, cleanup := range cs.immediateQueue {
		if cleanup.ID == target.ID {
			cs.immediateQueue = append(cs.immediateQueue[:i], cs.immediateQueue[i+1:]...)
			return
		}
	}

	// Remove from delayed queue
	for i, cleanup := range cs.delayedQueue {
		if cleanup.ID == target.ID {
			cs.delayedQueue = append(cs.delayedQueue[:i], cs.delayedQueue[i+1:]...)
			return
		}
	}

	// Remove from periodic queue
	for i, cleanup := range cs.periodicQueue {
		if cleanup.ID == target.ID {
			cs.periodicQueue = append(cs.periodicQueue[:i], cs.periodicQueue[i+1:]...)
			return
		}
	}
}

// ------------------------------------
// 📊 Statistics and Reporting
// ------------------------------------

// GetStats returns current scheduler statistics
func (cs *CleanupScheduler) GetStats() CleanupSchedulerStats {
	cs.stats.mutex.RLock()
	defer cs.stats.mutex.RUnlock()

	stats := cs.stats

	// Calculate throughput
	elapsed := time.Since(stats.StartTime).Seconds()
	if elapsed > 0 {
		stats.ThroughputPerSecond = float64(stats.TotalExecuted) / elapsed
	}

	// Count active workers
	cs.workerMutex.RLock()
	activeCount := uint64(0)
	for _, worker := range cs.workers {
		worker.mutex.RLock()
		if worker.active {
			activeCount++
		}
		worker.mutex.RUnlock()
	}
	cs.workerMutex.RUnlock()
	stats.ActiveWorkers = activeCount

	return stats
}

// GetReport returns a comprehensive scheduler report
func (cs *CleanupScheduler) GetReport() map[string]interface{} {
	cs.mutex.RLock()
	defer cs.mutex.RUnlock()

	report := map[string]interface{}{
		"scheduler_id": cs.id,
		"enabled":      cs.enabled,
		"running":      cs.running,
		"paused":       cs.paused,
		"config":       cs.config,
		"stats":        cs.GetStats(),
		"queue_sizes": map[string]int{
			"immediate": len(cs.immediateQueue),
			"delayed":   len(cs.delayedQueue),
			"periodic":  len(cs.periodicQueue),
		},
		"worker_count": len(cs.workers),
	}

	return report
}

// updateStats safely updates scheduler statistics
func (cs *CleanupScheduler) updateStats(fn func(*CleanupSchedulerStats)) {
	cs.stats.mutex.Lock()
	defer cs.stats.mutex.Unlock()
	fn(&cs.stats)
}

// ------------------------------------
// 🛠️ Configuration and Control
// ------------------------------------

// Start starts the cleanup scheduler
func (cs *CleanupScheduler) Start() {
	cs.mutex.Lock()
	defer cs.mutex.Unlock()
	cs.running = true
}

// Stop stops the cleanup scheduler
func (cs *CleanupScheduler) Stop() {
	cs.mutex.Lock()
	defer cs.mutex.Unlock()
	cs.running = false
}

// Pause pauses the cleanup scheduler
func (cs *CleanupScheduler) Pause() {
	cs.mutex.Lock()
	defer cs.mutex.Unlock()
	cs.paused = true
}

// Resume resumes the cleanup scheduler
func (cs *CleanupScheduler) Resume() {
	cs.mutex.Lock()
	defer cs.mutex.Unlock()
	cs.paused = false
}

// SetEnabled enables or disables the scheduler
func (cs *CleanupScheduler) SetEnabled(enabled bool) {
	cs.mutex.Lock()
	defer cs.mutex.Unlock()
	cs.enabled = enabled
}

// UpdateConfig updates the scheduler configuration
func (cs *CleanupScheduler) UpdateConfig(config SchedulerConfig) {
	cs.mutex.Lock()
	defer cs.mutex.Unlock()

	cs.config = config

	// Update ticker interval if changed
	if cs.scheduleTicker != nil {
		cs.scheduleTicker.Stop()
		cs.scheduleTicker = time.NewTicker(config.ScheduleInterval)
	}

	// Adjust worker count if needed
	cs.adjustWorkerCount(config.WorkerCount)
}

// adjustWorkerCount adjusts the number of workers
func (cs *CleanupScheduler) adjustWorkerCount(targetCount int) {
	cs.workerMutex.Lock()
	defer cs.workerMutex.Unlock()

	currentCount := len(cs.workers)

	if targetCount > currentCount {
		// Add workers
		for i := currentCount; i < targetCount; i++ {
			worker := cs.createWorker()
			cs.workers = append(cs.workers, worker)
			go worker.run()
		}
	} else if targetCount < currentCount {
		// Remove workers
		for i := targetCount; i < currentCount; i++ {
			if cs.workers[i].cancel != nil {
				cs.workers[i].cancel()
			}
		}
		cs.workers = cs.workers[:targetCount]
	}
}

// ------------------------------------
// 🧹 Disposal
// ------------------------------------

// Dispose cleans up the cleanup scheduler and all its resources
func (cs *CleanupScheduler) Dispose() {
	cs.mutex.Lock()
	defer cs.mutex.Unlock()

	// Stop background scheduling
	if cs.cancel != nil {
		cs.cancel()
	}

	if cs.scheduleTicker != nil {
		cs.scheduleTicker.Stop()
	}

	select {
	case cs.stopScheduling <- true:
	default:
	}

	// Stop all workers
	cs.workerMutex.Lock()
	for _, worker := range cs.workers {
		if worker.cancel != nil {
			worker.cancel()
		}
	}
	cs.workers = make([]*CleanupWorker, 0)
	cs.workerMutex.Unlock()

	// Wait for background goroutines to finish
	cs.wg.Wait()

	// Clear queues
	cs.immediateQueue = make([]*ScheduledCleanup, 0)
	cs.delayedQueue = make([]*ScheduledCleanup, 0)
	cs.periodicQueue = make([]*ScheduledCleanup, 0)

	// Close channels after goroutines have finished
	close(cs.workQueue)
	close(cs.resultQueue)

	cs.enabled = false
	cs.running = false
}
