//go:build js && wasm

// event_subscriptions.go
// Subscription tracking and automatic cleanup implementation

package golid

import (
	"sync"
	"sync/atomic"
	"syscall/js"
	"time"
)

// ------------------------------------
// 🎯 Event Batcher Types
// ------------------------------------

// EventBatcher manages batched event processing for performance
type EventBatcher struct {
	queue        chan *BatchedEvent
	processing   bool
	batchSize    int
	flushTimeout time.Duration
	scheduler    *Scheduler
	metrics      *BatchMetrics
	mutex        sync.RWMutex
	disposed     bool
}

// BatchedEvent represents an event to be processed in a batch
type BatchedEvent struct {
	handler   func()
	priority  Priority
	timestamp time.Time
}

// BatchMetrics tracks batching performance
type BatchMetrics struct {
	eventsQueued        uint64
	batchesProcessed    uint64
	averageBatchSize    float64
	totalProcessingTime time.Duration
	mutex               sync.RWMutex
}

// ------------------------------------
// 🎯 Custom Event Bus Types
// ------------------------------------

// CustomEventBus manages application-level custom events
type CustomEventBus struct {
	listeners map[string]map[uint64]*CustomEventListener
	mutex     sync.RWMutex
	disposed  bool
}

// CustomEventListener represents a custom event listener
type CustomEventListener struct {
	id      uint64
	handler func(interface{})
	once    bool
	owner   *Owner
	created time.Time
}

// ------------------------------------
// 🎯 Event Metrics Implementation
// ------------------------------------

// NewEventMetrics creates a new event metrics tracker
func NewEventMetrics() *EventMetrics {
	return &EventMetrics{
		eventCounts: make(map[string]uint64),
	}
}

// incrementTotal increments the total subscription count
func (m *EventMetrics) incrementTotal() {
	atomic.AddUint64(&m.totalSubscriptions, 1)
	atomic.AddUint64(&m.activeSubscriptions, 1)

	// Update peak if necessary
	current := atomic.LoadUint64(&m.activeSubscriptions)
	for {
		peak := atomic.LoadUint64(&m.peakSubscriptions)
		if current <= peak || atomic.CompareAndSwapUint64(&m.peakSubscriptions, peak, current) {
			break
		}
	}
}

// incrementDelegated increments the delegated events count
func (m *EventMetrics) incrementDelegated() {
	atomic.AddUint64(&m.delegatedEvents, 1)
}

// incrementDirect increments the direct events count
func (m *EventMetrics) incrementDirect() {
	atomic.AddUint64(&m.directEvents, 1)
}

// incrementCleanup increments the cleanup operations count
func (m *EventMetrics) incrementCleanup() {
	atomic.AddUint64(&m.cleanupOperations, 1)
	atomic.AddUint64(&m.activeSubscriptions, ^uint64(0)) // Decrement
}

// addCleanup adds multiple cleanup operations
func (m *EventMetrics) addCleanup(count uint64) {
	atomic.AddUint64(&m.cleanupOperations, count)
	// Subtract from active subscriptions
	for i := uint64(0); i < count; i++ {
		atomic.AddUint64(&m.activeSubscriptions, ^uint64(0)) // Decrement
	}
}

// GetStats returns current metrics statistics
func (m *EventMetrics) GetStats() map[string]interface{} {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	return map[string]interface{}{
		"totalSubscriptions":  atomic.LoadUint64(&m.totalSubscriptions),
		"activeSubscriptions": atomic.LoadUint64(&m.activeSubscriptions),
		"delegatedEvents":     atomic.LoadUint64(&m.delegatedEvents),
		"directEvents":        atomic.LoadUint64(&m.directEvents),
		"cleanupOperations":   atomic.LoadUint64(&m.cleanupOperations),
		"memoryLeaksDetected": atomic.LoadUint64(&m.memoryLeaksDetected),
		"peakSubscriptions":   atomic.LoadUint64(&m.peakSubscriptions),
		"averageResponseTime": m.averageResponseTime,
		"eventCounts":         m.eventCounts,
	}
}

// ------------------------------------
// 🏭 Event Batcher Implementation
// ------------------------------------

// NewEventBatcher creates a new event batcher
func NewEventBatcher() *EventBatcher {
	batcher := &EventBatcher{
		queue:        make(chan *BatchedEvent, 1000),
		batchSize:    50,
		flushTimeout: 16 * time.Millisecond, // ~60fps
		scheduler:    getScheduler(),
		metrics:      NewBatchMetrics(),
	}

	go batcher.processBatches()
	return batcher
}

// NewBatchMetrics creates a new batch metrics tracker
func NewBatchMetrics() *BatchMetrics {
	return &BatchMetrics{}
}

// Schedule adds an event to the batch queue
func (b *EventBatcher) Schedule(handler func(), priority Priority) {
	if b.disposed {
		return
	}

	event := &BatchedEvent{
		handler:   handler,
		priority:  priority,
		timestamp: time.Now(),
	}

	select {
	case b.queue <- event:
		atomic.AddUint64(&b.metrics.eventsQueued, 1)
	default:
		// Queue is full, execute immediately to prevent blocking
		handler()
	}
}

// processBatches processes events in batches
func (b *EventBatcher) processBatches() {
	ticker := time.NewTicker(b.flushTimeout)
	defer ticker.Stop()

	batch := make([]*BatchedEvent, 0, b.batchSize)

	for {
		select {
		case event := <-b.queue:
			if event == nil {
				return // Channel closed
			}

			batch = append(batch, event)

			// Process batch if it's full
			if len(batch) >= b.batchSize {
				b.processBatch(batch)
				batch = batch[:0] // Reset slice
			}

		case <-ticker.C:
			// Process batch on timeout
			if len(batch) > 0 {
				b.processBatch(batch)
				batch = batch[:0] // Reset slice
			}
		}
	}
}

// processBatch processes a batch of events
func (b *EventBatcher) processBatch(batch []*BatchedEvent) {
	if len(batch) == 0 {
		return
	}

	startTime := time.Now()

	// Sort by priority (higher priority first)
	b.sortBatchByPriority(batch)

	// Execute all handlers in the batch
	for _, event := range batch {
		event.handler()
	}

	// Update metrics
	processingTime := time.Since(startTime)
	b.metrics.mutex.Lock()
	atomic.AddUint64(&b.metrics.batchesProcessed, 1)
	b.metrics.totalProcessingTime += processingTime

	// Update average batch size
	batchCount := atomic.LoadUint64(&b.metrics.batchesProcessed)
	b.metrics.averageBatchSize = (b.metrics.averageBatchSize*float64(batchCount-1) + float64(len(batch))) / float64(batchCount)
	b.metrics.mutex.Unlock()
}

// sortBatchByPriority sorts events by priority
func (b *EventBatcher) sortBatchByPriority(batch []*BatchedEvent) {
	// Simple bubble sort for small batches
	n := len(batch)
	for i := 0; i < n-1; i++ {
		for j := 0; j < n-i-1; j++ {
			// Safe priority comparison - convert to int for comparison
			if int(batch[j].priority) > int(batch[j+1].priority) {
				batch[j], batch[j+1] = batch[j+1], batch[j]
			}
		}
	}
}

// Dispose cleans up the batcher
func (b *EventBatcher) Dispose() {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	if b.disposed {
		return
	}

	b.disposed = true
	close(b.queue)
}

// GetStats returns batcher statistics
func (b *EventBatcher) GetStats() map[string]interface{} {
	b.metrics.mutex.RLock()
	defer b.metrics.mutex.RUnlock()

	return map[string]interface{}{
		"eventsQueued":        atomic.LoadUint64(&b.metrics.eventsQueued),
		"batchesProcessed":    atomic.LoadUint64(&b.metrics.batchesProcessed),
		"averageBatchSize":    b.metrics.averageBatchSize,
		"totalProcessingTime": b.metrics.totalProcessingTime,
		"queueLength":         len(b.queue),
		"queueCapacity":       cap(b.queue),
	}
}

// ------------------------------------
// 🎯 Custom Event Bus Implementation
// ------------------------------------

// NewCustomEventBus creates a new custom event bus
func NewCustomEventBus() *CustomEventBus {
	return &CustomEventBus{
		listeners: make(map[string]map[uint64]*CustomEventListener),
	}
}

// On adds a listener for a custom event
func (bus *CustomEventBus) On(event string, handler func(interface{})) func() {
	return bus.subscribe(event, handler, false)
}

// Once adds a one-time listener for a custom event
func (bus *CustomEventBus) Once(event string, handler func(interface{})) func() {
	return bus.subscribe(event, handler, true)
}

// subscribe adds a listener with the specified options
func (bus *CustomEventBus) subscribe(event string, handler func(interface{}), once bool) func() {
	if bus.disposed {
		return func() {}
	}

	listener := &CustomEventListener{
		id:      atomic.AddUint64(&subscriptionIdCounter, 1),
		handler: handler,
		once:    once,
		owner:   getCurrentOwner(),
		created: time.Now(),
	}

	bus.mutex.Lock()
	if bus.listeners[event] == nil {
		bus.listeners[event] = make(map[uint64]*CustomEventListener)
	}
	bus.listeners[event][listener.id] = listener
	bus.mutex.Unlock()

	// Register cleanup with owner
	if listener.owner != nil {
		OnCleanup(func() {
			bus.Off(event, listener.id)
		})
	}

	// Return unsubscribe function
	return func() {
		bus.Off(event, listener.id)
	}
}

// Off removes a listener
func (bus *CustomEventBus) Off(event string, id uint64) {
	bus.mutex.Lock()
	defer bus.mutex.Unlock()

	if listeners, exists := bus.listeners[event]; exists {
		delete(listeners, id)
		if len(listeners) == 0 {
			delete(bus.listeners, event)
		}
	}
}

// Emit emits a custom event to all listeners
func (bus *CustomEventBus) Emit(event string, data interface{}) {
	bus.mutex.RLock()
	listeners := bus.listeners[event]
	if listeners == nil {
		bus.mutex.RUnlock()
		return
	}

	// Copy listeners to avoid holding lock during execution
	listenersCopy := make([]*CustomEventListener, 0, len(listeners))
	for _, listener := range listeners {
		listenersCopy = append(listenersCopy, listener)
	}
	bus.mutex.RUnlock()

	// Execute listeners
	toRemove := make([]uint64, 0)
	for _, listener := range listenersCopy {
		listener.handler(data)

		// Mark one-time listeners for removal
		if listener.once {
			toRemove = append(toRemove, listener.id)
		}
	}

	// Remove one-time listeners
	if len(toRemove) > 0 {
		bus.mutex.Lock()
		if listeners := bus.listeners[event]; listeners != nil {
			for _, id := range toRemove {
				delete(listeners, id)
			}
			if len(listeners) == 0 {
				delete(bus.listeners, event)
			}
		}
		bus.mutex.Unlock()
	}
}

// Dispose cleans up the event bus
func (bus *CustomEventBus) Dispose() {
	bus.mutex.Lock()
	defer bus.mutex.Unlock()

	if bus.disposed {
		return
	}

	bus.listeners = make(map[string]map[uint64]*CustomEventListener)
	bus.disposed = true
}

// GetStats returns event bus statistics
func (bus *CustomEventBus) GetStats() map[string]interface{} {
	bus.mutex.RLock()
	defer bus.mutex.RUnlock()

	totalListeners := 0
	eventTypes := len(bus.listeners)

	for _, listeners := range bus.listeners {
		totalListeners += len(listeners)
	}

	return map[string]interface{}{
		"eventTypes":     eventTypes,
		"totalListeners": totalListeners,
		"disposed":       bus.disposed,
	}
}

// ------------------------------------
// 🔧 Subscription Pool for Performance
// ------------------------------------

// SubscriptionPool manages reusable subscription objects
type SubscriptionPool struct {
	pool    chan *EventSubscription
	maxSize int
	created uint64
	reused  uint64
	mutex   sync.RWMutex
}

// NewSubscriptionPool creates a new subscription pool
func NewSubscriptionPool(maxSize int) *SubscriptionPool {
	return &SubscriptionPool{
		pool:    make(chan *EventSubscription, maxSize),
		maxSize: maxSize,
	}
}

// Get retrieves a subscription from the pool
func (p *SubscriptionPool) Get() *EventSubscription {
	select {
	case sub := <-p.pool:
		atomic.AddUint64(&p.reused, 1)
		return sub
	default:
		atomic.AddUint64(&p.created, 1)
		return &EventSubscription{}
	}
}

// Put returns a subscription to the pool
func (p *SubscriptionPool) Put(sub *EventSubscription) {
	if sub == nil {
		return
	}

	// Reset subscription state
	sub.handler = nil
	sub.cleanup = nil
	sub.element = js.Undefined()
	sub.owner = nil

	select {
	case p.pool <- sub:
		// Successfully returned to pool
	default:
		// Pool is full, let it be garbage collected
	}
}

// GetStats returns pool statistics
func (p *SubscriptionPool) GetStats() map[string]interface{} {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	return map[string]interface{}{
		"created":   atomic.LoadUint64(&p.created),
		"reused":    atomic.LoadUint64(&p.reused),
		"available": len(p.pool),
		"capacity":  p.maxSize,
	}
}
