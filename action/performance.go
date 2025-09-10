//go:build !js && !wasm
// +build !js,!wasm

package action

import (
	"runtime"
	"sync"
	"time"

	"github.com/ozanturksever/uiwgo/reactivity"
)

// PerformanceConfig holds configuration for performance optimizations
type PerformanceConfig struct {
	// Object pools
	EnableObjectPooling bool
	ActionPoolSize      int
	ContextPoolSize     int
	SubscriberPoolSize  int

	// Batching
	EnableReactiveBatching bool
	BatchWindow            time.Duration
	BatchSize              int

	// Async scheduling
	EnableMicrotaskScheduler bool
	MicrotaskQueueSize       int
	WorkerPoolSize           int

	// Profiling
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

// DefaultPerformanceConfig returns a default performance configuration
func DefaultPerformanceConfig() PerformanceConfig {
	return PerformanceConfig{
		EnableObjectPooling:      true,
		ActionPoolSize:           1000,
		ContextPoolSize:          500,
		SubscriberPoolSize:       100,
		EnableReactiveBatching:   true,
		BatchWindow:              time.Microsecond * 16, // ~60fps
		BatchSize:                50,
		EnableMicrotaskScheduler: true,
		MicrotaskQueueSize:       10000,
		WorkerPoolSize:           4,
		EnableProfiling:          false,
		ProfilingLevel:           ProfilingOff,
		MemoryTrackingLevel:      MemoryTrackingOff,
	}
}

// Performance manager for the action system
type performanceManager struct {
	config PerformanceConfig

	// Object pools
	actionPool     *actionPool
	contextPool    *contextPool
	subscriberPool *subscriberPool

	// Batching
	batchProcessor *batchProcessor

	// Async scheduling
	microtaskScheduler *microtaskScheduler

	// Profiling
	profiler *actionProfiler

	mu sync.RWMutex
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

// NewPerformanceManager creates a new performance manager
func NewPerformanceManager(config PerformanceConfig) *performanceManager {
	pm := &performanceManager{
		config: config,
	}

	if config.EnableObjectPooling {
		pm.actionPool = newActionPool(config.ActionPoolSize)
		pm.contextPool = newContextPool(config.ContextPoolSize)
		pm.subscriberPool = newSubscriberPool(config.SubscriberPoolSize)
	}

	if config.EnableReactiveBatching {
		pm.batchProcessor = newBatchProcessor(config.BatchWindow, config.BatchSize)
	}

	if config.EnableMicrotaskScheduler {
		pm.microtaskScheduler = newMicrotaskScheduler(config.MicrotaskQueueSize, config.WorkerPoolSize)
	}

	if config.EnableProfiling {
		pm.profiler = newActionProfiler(config.ProfilingLevel, config.MemoryTrackingLevel)
	}

	return pm
}

// Action pool for reusing Action objects
type actionPool struct {
	stringActions sync.Pool
	anyActions    sync.Pool
	maxSize       int
	mu            sync.Mutex
	currentSize   int
}

func newActionPool(maxSize int) *actionPool {
	return &actionPool{
		stringActions: sync.Pool{
			New: func() interface{} {
				return &Action[string]{}
			},
		},
		anyActions: sync.Pool{
			New: func() interface{} {
				return &Action[any]{}
			},
		},
		maxSize: maxSize,
	}
}

func (ap *actionPool) getStringAction() *Action[string] {
	action := ap.stringActions.Get().(*Action[string])
	// Reset the action
	*action = Action[string]{}
	return action
}

func (ap *actionPool) putStringAction(action *Action[string]) {
	if action == nil {
		return
	}

	ap.mu.Lock()
	if ap.currentSize < ap.maxSize {
		ap.currentSize++
		ap.stringActions.Put(action)
	}
	ap.mu.Unlock()
}

func (ap *actionPool) getAnyAction() *Action[any] {
	action := ap.anyActions.Get().(*Action[any])
	// Reset the action
	*action = Action[any]{}
	return action
}

func (ap *actionPool) putAnyAction(action *Action[any]) {
	if action == nil {
		return
	}

	ap.mu.Lock()
	if ap.currentSize < ap.maxSize {
		ap.currentSize++
		ap.anyActions.Put(action)
	}
	ap.mu.Unlock()
}

// Context pool for reusing Context objects
type contextPool struct {
	pool    sync.Pool
	maxSize int
	mu      sync.Mutex
	size    int
}

func newContextPool(maxSize int) *contextPool {
	return &contextPool{
		pool: sync.Pool{
			New: func() interface{} {
				return &Context{}
			},
		},
		maxSize: maxSize,
	}
}

func (cp *contextPool) get() *Context {
	ctx := cp.pool.Get().(*Context)
	// Reset the context
	*ctx = Context{
		Meta: make(map[string]any),
	}
	return ctx
}

func (cp *contextPool) put(ctx *Context) {
	if ctx == nil {
		return
	}

	cp.mu.Lock()
	if cp.size < cp.maxSize {
		cp.size++
		cp.pool.Put(ctx)
	}
	cp.mu.Unlock()
}

// Subscriber pool for reusing subscriber slices
type subscriberPool struct {
	pool    sync.Pool
	maxSize int
}

func newSubscriberPool(maxSize int) *subscriberPool {
	return &subscriberPool{
		pool: sync.Pool{
			New: func() interface{} {
				return make([]*subscriptionEntry, 0, 16)
			},
		},
		maxSize: maxSize,
	}
}

func (sp *subscriberPool) get() []*subscriptionEntry {
	slice := sp.pool.Get().([]*subscriptionEntry)
	return slice[:0] // Reset length but keep capacity
}

func (sp *subscriberPool) put(slice []*subscriptionEntry) {
	if slice == nil || cap(slice) > sp.maxSize {
		return
	}
	sp.pool.Put(slice)
}

// Batch processor for reactive signal updates
type batchProcessor struct {
	window     time.Duration
	maxSize    int
	mu         sync.Mutex
	pending    map[reactivity.Signal[any]]any
	timer      *time.Timer
	processing bool
}

func newBatchProcessor(window time.Duration, maxSize int) *batchProcessor {
	return &batchProcessor{
		window:  window,
		maxSize: maxSize,
		pending: make(map[reactivity.Signal[any]]any),
	}
}

func (bp *batchProcessor) scheduleUpdate(signal reactivity.Signal[any], value any) {
	bp.mu.Lock()
	defer bp.mu.Unlock()

	bp.pending[signal] = value

	if !bp.processing {
		if bp.timer != nil {
			bp.timer.Stop()
		}

		if len(bp.pending) >= bp.maxSize {
			// Process immediately if batch is full
			bp.processBatch()
		} else {
			// Schedule processing
			bp.timer = time.AfterFunc(bp.window, func() {
				bp.mu.Lock()
				if !bp.processing {
					bp.processBatch()
				}
				bp.mu.Unlock()
			})
		}
	}
}

func (bp *batchProcessor) processBatch() {
	if bp.processing {
		return
	}

	bp.processing = true
	defer func() {
		bp.processing = false
	}()

	if len(bp.pending) == 0 {
		return
	}

	// Copy pending updates
	updates := make(map[reactivity.Signal[any]]any, len(bp.pending))
	for signal, value := range bp.pending {
		updates[signal] = value
	}

	// Clear pending
	for signal := range bp.pending {
		delete(bp.pending, signal)
	}

	// Apply updates outside of lock
	bp.mu.Unlock()
	for signal, value := range updates {
		signal.Set(value)
	}
	bp.mu.Lock()
}

// Microtask scheduler for efficient async operations
type microtaskScheduler struct {
	queue      chan func()
	workers    int
	workerPool []chan func()
	roundRobin int
	mu         sync.Mutex
	started    bool
	stopCh     chan struct{}
}

func newMicrotaskScheduler(queueSize, workers int) *microtaskScheduler {
	ms := &microtaskScheduler{
		queue:      make(chan func(), queueSize),
		workers:    workers,
		workerPool: make([]chan func(), workers),
		stopCh:     make(chan struct{}),
	}

	// Create worker channels
	for i := 0; i < workers; i++ {
		ms.workerPool[i] = make(chan func(), 100)
	}

	return ms
}

func (ms *microtaskScheduler) start() {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	if ms.started {
		return
	}

	ms.started = true

	// Start main dispatcher
	go ms.dispatcher()

	// Start workers
	for i := 0; i < ms.workers; i++ {
		go ms.worker(i)
	}
}

func (ms *microtaskScheduler) dispatcher() {
	for {
		select {
		case task := <-ms.queue:
			// Round-robin dispatch to workers
			ms.mu.Lock()
			workerCh := ms.workerPool[ms.roundRobin]
			ms.roundRobin = (ms.roundRobin + 1) % ms.workers
			ms.mu.Unlock()

			select {
			case workerCh <- task:
				// Task dispatched
			default:
				// Worker busy, execute immediately in dispatcher
				go task()
			}

		case <-ms.stopCh:
			return
		}
	}
}

func (ms *microtaskScheduler) worker(id int) {
	workerCh := ms.workerPool[id]
	for {
		select {
		case task := <-workerCh:
			func() {
				defer func() {
					if r := recover(); r != nil {
						// Log panic but don't crash worker
						runtime.GC() // Help with cleanup after panic
					}
				}()
				task()
			}()

		case <-ms.stopCh:
			return
		}
	}
}

func (ms *microtaskScheduler) schedule(task func()) {
	if !ms.started {
		ms.start()
	}

	select {
	case ms.queue <- task:
		// Task queued
	default:
		// Queue full, execute immediately
		go task()
	}
}

func (ms *microtaskScheduler) stop() {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	if !ms.started {
		return
	}

	close(ms.stopCh)
	ms.started = false
}

// Action profiler for development builds
type actionProfiler struct {
	level           ProfilingLevel
	memoryLevel     MemoryTrackingLevel
	mu              sync.RWMutex
	dispatchMetrics map[string]*dispatchMetrics
	enabled         bool
}

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

func newActionProfiler(level ProfilingLevel, memoryLevel MemoryTrackingLevel) *actionProfiler {
	return &actionProfiler{
		level:           level,
		memoryLevel:     memoryLevel,
		dispatchMetrics: make(map[string]*dispatchMetrics),
		enabled:         level != ProfilingOff,
	}
}

func (ap *actionProfiler) profileDispatch(actionType string, duration time.Duration, allocBytes, allocCount int64) {
	if !ap.enabled {
		return
	}

	ap.mu.Lock()
	defer ap.mu.Unlock()

	metrics, exists := ap.dispatchMetrics[actionType]
	if !exists {
		metrics = &dispatchMetrics{
			minDuration: duration,
			maxDuration: duration,
		}
		ap.dispatchMetrics[actionType] = metrics
	}

	metrics.count++
	metrics.totalDuration += duration
	metrics.avgDuration = time.Duration(int64(metrics.totalDuration) / metrics.count)

	if duration > metrics.maxDuration {
		metrics.maxDuration = duration
	}
	if duration < metrics.minDuration {
		metrics.minDuration = duration
	}

	if ap.memoryLevel != MemoryTrackingOff {
		metrics.totalAllocBytes += allocBytes
		metrics.totalAllocCount += allocCount
	}

	metrics.lastUpdated = time.Now()
}

func (ap *actionProfiler) getMetrics(actionType string) *dispatchMetrics {
	ap.mu.RLock()
	defer ap.mu.RUnlock()

	if metrics, exists := ap.dispatchMetrics[actionType]; exists {
		// Return a copy to avoid race conditions
		return &dispatchMetrics{
			count:           metrics.count,
			totalDuration:   metrics.totalDuration,
			avgDuration:     metrics.avgDuration,
			maxDuration:     metrics.maxDuration,
			minDuration:     metrics.minDuration,
			totalAllocBytes: metrics.totalAllocBytes,
			totalAllocCount: metrics.totalAllocCount,
			lastUpdated:     metrics.lastUpdated,
		}
	}

	return nil
}

// API functions for performance management

// EnablePerformanceOptimizations enables performance optimizations with the given config
func EnablePerformanceOptimizations(config PerformanceConfig) {
	perfManagerOnce.Do(func() {
		globalPerfManager = NewPerformanceManager(config)
	})
}

// SetPerformanceConfig updates the performance configuration
func SetPerformanceConfig(config PerformanceConfig) {
	pm := GetPerformanceManager()
	pm.mu.Lock()
	defer pm.mu.Unlock()
	pm.config = config
}

// GetPerformanceConfig returns the current performance configuration
func GetPerformanceConfig() PerformanceConfig {
	pm := GetPerformanceManager()
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	return pm.config
}

// GetDispatchMetrics returns profiling metrics for an action type
func GetDispatchMetrics(actionType string) *dispatchMetrics {
	pm := GetPerformanceManager()
	if pm.profiler != nil {
		return pm.profiler.getMetrics(actionType)
	}
	return nil
}

// ResetPerformanceMetrics resets all performance metrics
func ResetPerformanceMetrics() {
	pm := GetPerformanceManager()
	if pm.profiler != nil {
		pm.profiler.mu.Lock()
		pm.profiler.dispatchMetrics = make(map[string]*dispatchMetrics)
		pm.profiler.mu.Unlock()
	}
}

// OptimizedDispatch performs an optimized dispatch using performance features
// This is an opt-in optimization that can be used when performance is critical
func OptimizedDispatch(bus Bus, action any, opts ...DispatchOption) error {
	pm := GetPerformanceManager()

	busImpl, ok := bus.(*busImpl)
	if !ok {
		// Fallback to regular dispatch for non-busImpl types
		return bus.Dispatch(action, opts...)
	}

	// Apply dispatch options - use context pool if available
	var context Context
	if pm.contextPool != nil && pm.config.EnableObjectPooling {
		contextPtr := pm.contextPool.get()
		context = *contextPtr
		defer pm.contextPool.put(contextPtr)
		// Reset context fields
		context.Scope = busImpl.scopePath
		context.Time = time.Now()
		context.TraceID = ""
		context.Source = ""
	} else {
		context = Context{
			Scope:   busImpl.scopePath,
			Meta:    make(map[string]any),
			Time:    time.Now(),
			TraceID: "",
			Source:  "",
		}
	}

	dispatchOpts := &dispatchOptions{
		context: context,
	}

	for _, opt := range opts {
		opt.applyDispatch(dispatchOpts)
	}

	// Build the action based on the input type
	var actionToDispatch any
	var actionType string

	switch act := action.(type) {
	case Action[string]:
		actionToDispatch = enhanceAction(act, dispatchOpts.context)
		actionType = act.Type
	case Action[any]:
		actionToDispatch = enhanceActionAny(act, dispatchOpts.context)
		actionType = act.Type
	case string:
		actionToDispatch = Action[string]{
			Type:    act,
			Payload: act,
			Meta:    dispatchOpts.context.Meta,
			Time:    dispatchOpts.context.Time,
			Source:  dispatchOpts.context.Source,
			TraceID: dispatchOpts.context.TraceID,
		}
		actionType = act
	default:
		actionToDispatch = Action[any]{
			Type:    "unknown",
			Payload: act,
			Meta:    dispatchOpts.context.Meta,
			Time:    dispatchOpts.context.Time,
			Source:  dispatchOpts.context.Source,
			TraceID: dispatchOpts.context.TraceID,
		}
		actionType = "unknown"
	}

	// Handle async dispatch with microtask scheduler
	if dispatchOpts.async {
		return optimizedDispatchAsync(busImpl, pm, actionToDispatch, actionType, dispatchOpts.context)
	}

	// Synchronous dispatch
	return optimizedDispatchSync(busImpl, pm, actionToDispatch, actionType, dispatchOpts.context)
}

// enhanceAction enhances an Action[string] with dispatch context
func enhanceAction(act Action[string], ctx Context) Action[string] {
	enhancedAction := act
	if enhancedAction.TraceID == "" {
		enhancedAction.TraceID = ctx.TraceID
	}
	if enhancedAction.Source == "" {
		enhancedAction.Source = ctx.Source
	}
	enhancedAction.Time = ctx.Time

	// Merge metadata
	if enhancedAction.Meta == nil {
		enhancedAction.Meta = make(map[string]any)
	}
	for k, v := range ctx.Meta {
		if _, exists := enhancedAction.Meta[k]; !exists {
			enhancedAction.Meta[k] = v
		}
	}

	return enhancedAction
}

// enhanceActionAny enhances an Action[any] with dispatch context
func enhanceActionAny(act Action[any], ctx Context) Action[any] {
	enhancedAction := act
	if enhancedAction.TraceID == "" {
		enhancedAction.TraceID = ctx.TraceID
	}
	if enhancedAction.Source == "" {
		enhancedAction.Source = ctx.Source
	}
	enhancedAction.Time = ctx.Time

	// Merge metadata
	if enhancedAction.Meta == nil {
		enhancedAction.Meta = make(map[string]any)
	}
	for k, v := range ctx.Meta {
		if _, exists := enhancedAction.Meta[k]; !exists {
			enhancedAction.Meta[k] = v
		}
	}

	return enhancedAction
}


// optimizedDispatchSync performs optimized synchronous dispatch
func optimizedDispatchSync(bus *busImpl, pm *performanceManager, action any, actionType string, ctx Context) error {
	var start time.Time
	var memStatsBefore, memStatsAfter runtime.MemStats

	// Profile if enabled
	if pm.profiler != nil && pm.profiler.enabled {
		start = time.Now()
		if pm.profiler.memoryLevel != MemoryTrackingOff {
			runtime.ReadMemStats(&memStatsBefore)
		}
	}

	bus.mu.RLock()
	// Get ordered subscribers - use optimized version with pooling if available
	var handlers []*subscriptionEntry
	if actionType != "unknown" {
		if pm.subscriberPool != nil && pm.config.EnableObjectPooling {
			pooledSlice := pm.subscriberPool.get()
			handlers = getOrderedSubscribersOptimized(bus, actionType, pooledSlice)
			defer pm.subscriberPool.put(pooledSlice)
		} else {
			handlers = bus.getOrderedSubscribers(actionType)
		}
	}

	// Get ordered any handlers
	var anyHandlers []*subscriptionEntry
	if pm.subscriberPool != nil && pm.config.EnableObjectPooling {
		pooledSlice := pm.subscriberPool.get()
		anyHandlers = getOrderedAnyHandlersOptimized(bus, pooledSlice)
		defer pm.subscriberPool.put(pooledSlice)
	} else {
		anyHandlers = bus.getOrderedAnyHandlers()
	}

	subscriberCount := len(handlers) + len(anyHandlers)
	bus.mu.RUnlock()

	// Instrument dispatch with observability features
	err := instrumentDispatch(bus, actionType, action, ctx, subscriberCount, func() error {
		bus.mu.RLock()
		defer bus.mu.RUnlock()

		// Dispatch to specific subscribers
		for _, entry := range handlers {
			if entry.active {
				bus.dispatchToHandler(entry, action, ctx)
			}
		}

		// Dispatch to any handlers
		for _, entry := range anyHandlers {
			if entry.active {
				bus.dispatchToHandler(entry, action, ctx)
			}
		}

		return nil
	})

	// Profile results if enabled
	if pm.profiler != nil && pm.profiler.enabled {
		duration := time.Since(start)
		var allocBytes, allocCount int64

		if pm.profiler.memoryLevel != MemoryTrackingOff {
			runtime.ReadMemStats(&memStatsAfter)
			allocBytes = int64(memStatsAfter.TotalAlloc - memStatsBefore.TotalAlloc)
			allocCount = int64(memStatsAfter.Mallocs - memStatsBefore.Mallocs)
		}

		pm.profiler.profileDispatch(actionType, duration, allocBytes, allocCount)
	}

	return err
}

// optimizedDispatchAsync performs optimized asynchronous dispatch
func optimizedDispatchAsync(bus *busImpl, pm *performanceManager, action any, actionType string, ctx Context) error {
	task := func() {
		defer func() {
			if r := recover(); r != nil {
				handleEnhancedError(bus, ctx, &panicError{value: r}, r)
			}
		}()
		optimizedDispatchSync(bus, pm, action, actionType, ctx)
	}

	// Use microtask scheduler if available
	if pm.microtaskScheduler != nil && pm.config.EnableMicrotaskScheduler {
		pm.microtaskScheduler.schedule(task)
	} else {
		// Fallback to goroutine
		go task()
	}

	return nil
}

// Helper functions for optimized subscriber handling
func getOrderedSubscribersOptimized(bus *busImpl, actionType string, result []*subscriptionEntry) []*subscriptionEntry {
	handlers, exists := bus.subscribers[actionType]
	if !exists {
		return result[:0]
	}

	// Copy to pre-allocated slice
	result = result[:0]
	for _, handler := range handlers {
		result = append(result, handler)
	}

	// Sort by priority (descending) then by creation time (ascending)
	for i := 0; i < len(result)-1; i++ {
		for j := i + 1; j < len(result); j++ {
			if result[i].priority < result[j].priority ||
				(result[i].priority == result[j].priority && result[i].createdAt.After(result[j].createdAt)) {
				result[i], result[j] = result[j], result[i]
			}
		}
	}

	return result
}

func getOrderedAnyHandlersOptimized(bus *busImpl, result []*subscriptionEntry) []*subscriptionEntry {
	// Copy to pre-allocated slice
	result = result[:0]
	for _, handler := range bus.anyHandlers {
		result = append(result, handler)
	}

	// Sort by priority (descending) then by creation time (ascending)
	for i := 0; i < len(result)-1; i++ {
		for j := i + 1; j < len(result); j++ {
			if result[i].priority < result[j].priority ||
				(result[i].priority == result[j].priority && result[i].createdAt.After(result[j].createdAt)) {
				result[i], result[j] = result[j], result[i]
			}
		}
	}

	return result
}

