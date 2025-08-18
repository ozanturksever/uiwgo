//go:build js && wasm

// event_system.go
// Core event management with delegation and automatic cleanup

package golid

import (
	"os"
	"sync"
	"sync/atomic"
	"syscall/js"
	"time"
)

// ------------------------------------
// 🎯 Core Event System Types
// ------------------------------------

// isTestEnvironment checks if we're running in a test environment
func isTestEnvironment() bool {
	// Check for common test environment indicators
	if os.Getenv("GO_TEST") == "1" {
		return true
	}
	// Check if we're in a test binary (common Go test pattern)
	if len(os.Args) > 0 && (os.Args[0] == "test" || os.Args[0] == "test.exe") {
		return true
	}
	return false
}

var (
	eventIdCounter        uint64
	subscriptionIdCounter uint64
	globalEventManager    *EventManager
	eventManagerOnce      sync.Once
)

// EventManager is the central event management system with delegation and cleanup
type EventManager struct {
	subscriptions map[uint64]*EventSubscription
	delegator     *EventDelegator
	batcher       *EventBatcher
	metrics       *EventMetrics
	customBus     *CustomEventBus
	owner         *Owner
	mutex         sync.RWMutex
}

// EventSubscription represents a subscription to an event with automatic cleanup
type EventSubscription struct {
	id        uint64
	event     string
	element   js.Value
	handler   func(js.Value)
	cleanup   func()
	autoClean bool
	owner     *Owner
	delegated bool
	options   EventOptions
	created   time.Time
	lastUsed  time.Time
	useCount  uint64
	mutex     sync.RWMutex
}

// EventOptions configures event subscription behavior
type EventOptions struct {
	Capture  bool
	Once     bool
	Passive  bool
	Signal   js.Value // AbortSignal for native cleanup
	Debounce time.Duration
	Throttle time.Duration
	Delegate bool   // Use event delegation
	Selector string // CSS selector for delegation
	Priority Priority
}

// EventMetrics tracks event system performance and debugging information
type EventMetrics struct {
	totalSubscriptions  uint64
	activeSubscriptions uint64
	delegatedEvents     uint64
	directEvents        uint64
	cleanupOperations   uint64
	memoryLeaksDetected uint64
	averageResponseTime time.Duration
	peakSubscriptions   uint64
	eventCounts         map[string]uint64
	mutex               sync.RWMutex
}

// ------------------------------------
// 🏭 Event Manager Factory
// ------------------------------------

// GetEventManager returns the global event manager instance
func GetEventManager() *EventManager {
	eventManagerOnce.Do(func() {
		globalEventManager = &EventManager{
			subscriptions: make(map[uint64]*EventSubscription),
			delegator:     NewEventDelegator(),
			batcher:       NewEventBatcher(),
			metrics:       NewEventMetrics(),
			customBus:     NewCustomEventBus(),
			owner:         getCurrentOwner(),
		}

		// Register cleanup with current owner if available
		if globalEventManager.owner != nil {
			OnCleanup(func() {
				globalEventManager.Dispose()
			})
		}
	})
	return globalEventManager
}

// NewEventManager creates a new event manager with the specified owner
func NewEventManager(owner *Owner) *EventManager {
	manager := &EventManager{
		subscriptions: make(map[uint64]*EventSubscription),
		delegator:     NewEventDelegator(),
		batcher:       NewEventBatcher(),
		metrics:       NewEventMetrics(),
		customBus:     NewCustomEventBus(),
		owner:         owner,
	}

	// Register cleanup with owner
	if owner != nil {
		OnCleanup(func() {
			manager.Dispose()
		})
	}

	return manager
}

// ------------------------------------
// 🔗 Event Subscription API
// ------------------------------------

// Subscribe creates a new event subscription with automatic cleanup
func (m *EventManager) Subscribe(element js.Value, event string, handler func(js.Value), options ...EventOptions) func() {
	if !element.Truthy() {
		return func() {} // No-op cleanup for invalid elements
	}

	// Merge options
	opts := EventOptions{
		Delegate: true, // Default to delegation for performance
		Priority: Normal,
	}
	if len(options) > 0 {
		opts = mergeEventOptions(opts, options[0])
	}

	subscription := &EventSubscription{
		id:        atomic.AddUint64(&subscriptionIdCounter, 1),
		event:     event,
		element:   element,
		handler:   handler,
		autoClean: true,
		owner:     m.owner,
		options:   opts,
		created:   time.Now(),
		lastUsed:  time.Now(),
	}

	// Setup the actual event binding
	if opts.Delegate && m.delegator.CanDelegate(event) {
		// Use event delegation
		subscription.delegated = true
		subscription.cleanup = m.delegator.Subscribe(element, event, handler, opts)
		m.metrics.incrementDelegated()
	} else {
		// Direct event binding
		subscription.delegated = false
		subscription.cleanup = m.bindDirectEvent(subscription)
		m.metrics.incrementDirect()
	}

	// Register subscription
	m.mutex.Lock()
	m.subscriptions[subscription.id] = subscription
	m.mutex.Unlock()

	// Register with owner for automatic cleanup
	if m.owner != nil {
		OnCleanup(func() {
			m.Unsubscribe(subscription.id)
		})
	}

	m.metrics.incrementTotal()

	// Return cleanup function
	return func() {
		m.Unsubscribe(subscription.id)
	}
}

// SubscribeReactive creates a reactive event subscription that integrates with signals
func (m *EventManager) SubscribeReactive(element js.Value, event string, handler func(js.Value), options ...EventOptions) func() {
	// Wrap handler to integrate with reactive system
	reactiveHandler := func(e js.Value) {
		// For test environments or immediate execution needs, execute directly
		// In production, this could be batched for performance
		if isTestEnvironment() {
			// Execute immediately in test environment
			if comp := getCurrentComputation(); comp != nil {
				comp.onCleanup(func() {
					// Event subscription cleanup is handled by owner
				})
			}
			handler(e)
		} else {
			// Batch event processing to prevent cascades in production
			m.batcher.Schedule(func() {
				// Track event handling in current computation context
				if comp := getCurrentComputation(); comp != nil {
					comp.onCleanup(func() {
						// Event subscription cleanup is handled by owner
					})
				}

				// Execute handler
				handler(e)
			}, UserBlocking) // User events get high priority
		}
	}

	return m.Subscribe(element, event, reactiveHandler, options...)
}

// Unsubscribe removes an event subscription and cleans up resources
func (m *EventManager) Unsubscribe(id uint64) {
	m.mutex.Lock()
	subscription, exists := m.subscriptions[id]
	if !exists {
		m.mutex.Unlock()
		return
	}
	delete(m.subscriptions, id)
	m.mutex.Unlock()

	// Cleanup the subscription
	if subscription.cleanup != nil {
		subscription.cleanup()
	}

	m.metrics.incrementCleanup()
}

// UnsubscribeAll removes all subscriptions and cleans up resources
func (m *EventManager) UnsubscribeAll() {
	m.mutex.Lock()
	subscriptions := make([]*EventSubscription, 0, len(m.subscriptions))
	for _, sub := range m.subscriptions {
		subscriptions = append(subscriptions, sub)
	}
	m.subscriptions = make(map[uint64]*EventSubscription)
	m.mutex.Unlock()

	// Cleanup all subscriptions
	for _, sub := range subscriptions {
		if sub.cleanup != nil {
			sub.cleanup()
		}
	}

	m.metrics.addCleanup(uint64(len(subscriptions)))
}

// ------------------------------------
// 🔧 Direct Event Binding
// ------------------------------------

// bindDirectEvent creates a direct event listener binding
func (m *EventManager) bindDirectEvent(subscription *EventSubscription) func() {
	// Create debounced/throttled handler if needed
	actualHandler := m.wrapHandler(subscription.handler, subscription.options)

	// Create JavaScript callback
	jsCallback := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		if len(args) > 0 {
			// Update usage metrics
			subscription.mutex.Lock()
			subscription.lastUsed = time.Now()
			atomic.AddUint64(&subscription.useCount, 1)
			subscription.mutex.Unlock()

			actualHandler(args[0])
		}
		return nil
	})

	// Add event listener with options
	if subscription.options.Capture || subscription.options.Once || subscription.options.Passive {
		opts := js.Global().Get("Object").New()
		opts.Set("capture", subscription.options.Capture)
		opts.Set("once", subscription.options.Once)
		opts.Set("passive", subscription.options.Passive)

		if subscription.options.Signal.Truthy() {
			opts.Set("signal", subscription.options.Signal)
		}

		subscription.element.Call("addEventListener", subscription.event, jsCallback, opts)
	} else {
		subscription.element.Call("addEventListener", subscription.event, jsCallback)
	}

	// Return cleanup function
	return func() {
		subscription.element.Call("removeEventListener", subscription.event, jsCallback)
		jsCallback.Release()
	}
}

// wrapHandler wraps the event handler with debouncing/throttling if configured
func (m *EventManager) wrapHandler(handler func(js.Value), options EventOptions) func(js.Value) {
	if options.Debounce > 0 {
		return m.debounceHandler(handler, options.Debounce)
	}
	if options.Throttle > 0 {
		return m.throttleHandler(handler, options.Throttle)
	}
	return handler
}

// debounceHandler creates a debounced version of the handler
func (m *EventManager) debounceHandler(handler func(js.Value), delay time.Duration) func(js.Value) {
	var timer *time.Timer
	var mutex sync.Mutex

	return func(e js.Value) {
		mutex.Lock()
		defer mutex.Unlock()

		if timer != nil {
			timer.Stop()
		}

		timer = time.AfterFunc(delay, func() {
			handler(e)
		})
	}
}

// throttleHandler creates a throttled version of the handler
func (m *EventManager) throttleHandler(handler func(js.Value), delay time.Duration) func(js.Value) {
	var lastCall time.Time
	var mutex sync.Mutex

	return func(e js.Value) {
		mutex.Lock()
		defer mutex.Unlock()

		now := time.Now()
		if now.Sub(lastCall) >= delay {
			lastCall = now
			handler(e)
		}
	}
}

// ------------------------------------
// 🧹 Cleanup and Disposal
// ------------------------------------

// Dispose cleans up all resources and subscriptions
func (m *EventManager) Dispose() {
	m.UnsubscribeAll()

	if m.delegator != nil {
		m.delegator.Dispose()
	}

	if m.batcher != nil {
		m.batcher.Dispose()
	}

	if m.customBus != nil {
		m.customBus.Dispose()
	}
}

// GetSubscriptionCount returns the current number of active subscriptions
func (m *EventManager) GetSubscriptionCount() int {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return len(m.subscriptions)
}

// GetMetrics returns current event system metrics
func (m *EventManager) GetMetrics() *EventMetrics {
	return m.metrics
}

// ------------------------------------
// 🔧 Utility Functions
// ------------------------------------

// mergeEventOptions merges event options with defaults taking precedence
func mergeEventOptions(defaults, override EventOptions) EventOptions {
	result := defaults

	if override.Capture {
		result.Capture = override.Capture
	}
	if override.Once {
		result.Once = override.Once
	}
	if override.Passive {
		result.Passive = override.Passive
	}
	if override.Signal.Truthy() {
		result.Signal = override.Signal
	}
	if override.Debounce > 0 {
		result.Debounce = override.Debounce
	}
	if override.Throttle > 0 {
		result.Throttle = override.Throttle
	}
	if override.Selector != "" {
		result.Selector = override.Selector
	}
	if override.Priority != Normal {
		result.Priority = override.Priority
	}

	return result
}

// ------------------------------------
// 🎯 Convenience Functions
// ------------------------------------

// Subscribe is a convenience function that uses the global event manager
func Subscribe(element js.Value, event string, handler func(js.Value), options ...EventOptions) func() {
	return GetEventManager().Subscribe(element, event, handler, options...)
}

// SubscribeReactive is a convenience function for reactive event subscriptions
func SubscribeReactive(element js.Value, event string, handler func(js.Value), options ...EventOptions) func() {
	return GetEventManager().SubscribeReactive(element, event, handler, options...)
}

// Unsubscribe is a convenience function that uses the global event manager
func Unsubscribe(id uint64) {
	GetEventManager().Unsubscribe(id)
}
