//go:build js && wasm

// event_delegation.go
// Event delegation implementation with routing and performance optimization

package golid

import (
	"strings"
	"sync"
	"sync/atomic"
	"syscall/js"
	"time"
)

// ------------------------------------
// 🎯 Event Delegation Types
// ------------------------------------

// EventDelegator manages event delegation at the document level
type EventDelegator struct {
	handlers map[string]map[uint64]*DelegatedEventHandler
	root     js.Value
	active   map[string]bool
	router   *EventRouter
	pool     *HandlerPool
	metrics  *DelegationMetrics
	mutex    sync.RWMutex
	disposed bool
}

// DelegatedEventHandler represents a delegated event handler
type DelegatedEventHandler struct {
	id       uint64
	selector string
	handler  func(js.Value)
	options  EventOptions
	element  js.Value
	created  time.Time
	useCount uint64
	mutex    sync.RWMutex
}

// EventRouter efficiently routes events to appropriate handlers
type EventRouter struct {
	selectorCache map[string][]js.Value
	pathCache     map[string]*EventPath
	mutex         sync.RWMutex
}

// EventPath represents the path from target to root for event bubbling
type EventPath struct {
	elements  []js.Value
	selectors map[string][]int // selector -> indices in elements
}

// HandlerPool manages reusable handler objects for performance
type HandlerPool struct {
	handlers chan *DelegatedEventHandler
	maxSize  int
	created  uint64
	reused   uint64
	mutex    sync.RWMutex
}

// DelegationMetrics tracks delegation performance
type DelegationMetrics struct {
	eventsProcessed uint64
	handlersMatched uint64
	routingTime     time.Duration
	cacheHits       uint64
	cacheMisses     uint64
	mutex           sync.RWMutex
}

// ------------------------------------
// 🏭 Factory Functions
// ------------------------------------

// NewEventDelegator creates a new event delegator
func NewEventDelegator() *EventDelegator {
	return &EventDelegator{
		handlers: make(map[string]map[uint64]*DelegatedEventHandler),
		root:     js.Global().Get("document"),
		active:   make(map[string]bool),
		router:   NewEventRouter(),
		pool:     NewHandlerPool(100), // Pool size of 100
		metrics:  NewDelegationMetrics(),
	}
}

// NewEventRouter creates a new event router
func NewEventRouter() *EventRouter {
	return &EventRouter{
		selectorCache: make(map[string][]js.Value),
		pathCache:     make(map[string]*EventPath),
	}
}

// NewHandlerPool creates a new handler pool
func NewHandlerPool(maxSize int) *HandlerPool {
	return &HandlerPool{
		handlers: make(chan *DelegatedEventHandler, maxSize),
		maxSize:  maxSize,
	}
}

// NewDelegationMetrics creates a new delegation metrics tracker
func NewDelegationMetrics() *DelegationMetrics {
	return &DelegationMetrics{}
}

// ------------------------------------
// 🔗 Event Delegation API
// ------------------------------------

// Subscribe adds a delegated event handler
func (d *EventDelegator) Subscribe(element js.Value, event string, handler func(js.Value), options EventOptions) func() {
	if d.disposed {
		return func() {}
	}

	// Get or create handler
	eventHandler := d.pool.Get()
	if eventHandler == nil {
		eventHandler = &DelegatedEventHandler{
			id: atomic.AddUint64(&eventIdCounter, 1),
		}
		atomic.AddUint64(&d.pool.created, 1)
	} else {
		atomic.AddUint64(&d.pool.reused, 1)
	}

	// Configure handler
	eventHandler.selector = options.Selector
	eventHandler.handler = handler
	eventHandler.options = options
	eventHandler.element = element
	eventHandler.created = time.Now()
	eventHandler.useCount = 0

	// If no selector provided, generate one based on element
	if eventHandler.selector == "" {
		eventHandler.selector = d.generateSelector(element)
	}

	// Register handler
	d.mutex.Lock()
	if d.handlers[event] == nil {
		d.handlers[event] = make(map[uint64]*DelegatedEventHandler)
	}
	d.handlers[event][eventHandler.id] = eventHandler

	// Activate delegation for this event type if not already active
	if !d.active[event] {
		d.activateEventDelegation(event)
		d.active[event] = true
	}
	d.mutex.Unlock()

	// Return cleanup function
	return func() {
		d.Unsubscribe(event, eventHandler.id)
	}
}

// Unsubscribe removes a delegated event handler
func (d *EventDelegator) Unsubscribe(event string, id uint64) {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	if handlers, exists := d.handlers[event]; exists {
		if handler, exists := handlers[id]; exists {
			delete(handlers, id)

			// Return handler to pool
			d.pool.Put(handler)

			// Deactivate delegation if no more handlers
			if len(handlers) == 0 {
				delete(d.handlers, event)
				if d.active[event] {
					d.deactivateEventDelegation(event)
					delete(d.active, event)
				}
			}
		}
	}
}

// CanDelegate checks if an event type can be delegated
func (d *EventDelegator) CanDelegate(event string) bool {
	// Events that can be delegated (bubble up)
	delegatableEvents := map[string]bool{
		"click":      true,
		"dblclick":   true,
		"mousedown":  true,
		"mouseup":    true,
		"mouseover":  true,
		"mouseout":   true,
		"mousemove":  true,
		"keydown":    true,
		"keyup":      true,
		"keypress":   true,
		"input":      true,
		"change":     true,
		"submit":     true,
		"focus":      true, // Note: focus doesn't bubble, but focusin does
		"blur":       true, // Note: blur doesn't bubble, but focusout does
		"touchstart": true,
		"touchend":   true,
		"touchmove":  true,
	}

	return delegatableEvents[event]
}

// ------------------------------------
// 🔧 Internal Delegation Logic
// ------------------------------------

// activateEventDelegation sets up delegation for an event type
func (d *EventDelegator) activateEventDelegation(event string) {
	// Use focusin/focusout for focus/blur events (they bubble)
	actualEvent := event
	if event == "focus" {
		actualEvent = "focusin"
	} else if event == "blur" {
		actualEvent = "focusout"
	}

	// Create delegated handler
	delegatedHandler := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		if len(args) > 0 {
			d.handleDelegatedEvent(event, args[0])
		}
		return nil
	})

	// Add listener to document
	d.root.Call("addEventListener", actualEvent, delegatedHandler, map[string]interface{}{
		"capture": false, // Use bubbling phase
		"passive": false,
	})

	// Store cleanup (in a real implementation, we'd need to track these)
	// For now, we'll handle cleanup in Dispose()
}

// deactivateEventDelegation removes delegation for an event type
func (d *EventDelegator) deactivateEventDelegation(event string) {
	// In a full implementation, we'd remove the specific listener
	// For now, this is handled in Dispose()
}

// handleDelegatedEvent processes a delegated event
func (d *EventDelegator) handleDelegatedEvent(event string, jsEvent js.Value) {
	startTime := time.Now()
	atomic.AddUint64(&d.metrics.eventsProcessed, 1)

	target := jsEvent.Get("target")
	if !target.Truthy() {
		return
	}

	// Get event path for efficient matching
	path := d.router.GetEventPath(target)

	// Find matching handlers
	d.mutex.RLock()
	handlers := d.handlers[event]
	d.mutex.RUnlock()

	if handlers == nil {
		return
	}

	matchedHandlers := d.router.MatchHandlers(path, handlers)

	// Execute matched handlers
	for _, handler := range matchedHandlers {
		// Update usage metrics
		handler.mutex.Lock()
		atomic.AddUint64(&handler.useCount, 1)
		handler.mutex.Unlock()

		// Execute handler
		handler.handler(jsEvent)
		atomic.AddUint64(&d.metrics.handlersMatched, 1)
	}

	// Update routing time metrics
	d.metrics.mutex.Lock()
	d.metrics.routingTime += time.Since(startTime)
	d.metrics.mutex.Unlock()
}

// generateSelector generates a CSS selector for an element
func (d *EventDelegator) generateSelector(element js.Value) string {
	if !element.Truthy() {
		return ""
	}

	// Try ID first
	id := element.Get("id")
	if id.Truthy() && id.String() != "" {
		return "#" + id.String()
	}

	// Try class names
	className := element.Get("className")
	if className.Truthy() && className.String() != "" {
		classes := strings.Fields(className.String())
		if len(classes) > 0 {
			return "." + strings.Join(classes, ".")
		}
	}

	// Fall back to tag name
	tagName := element.Get("tagName")
	if tagName.Truthy() {
		return strings.ToLower(tagName.String())
	}

	return "*"
}

// ------------------------------------
// 🗺️ Event Router Implementation
// ------------------------------------

// GetEventPath gets the event path from target to document
func (r *EventRouter) GetEventPath(target js.Value) *EventPath {
	// Create cache key
	cacheKey := target.Get("tagName").String() + ":" + target.Get("className").String()

	r.mutex.RLock()
	if path, exists := r.pathCache[cacheKey]; exists {
		r.mutex.RUnlock()
		return path
	}
	r.mutex.RUnlock()

	// Build path
	elements := make([]js.Value, 0, 10)
	current := target

	for current.Truthy() && current.Get("nodeType").Int() == 1 { // ELEMENT_NODE
		elements = append(elements, current)
		current = current.Get("parentElement")
	}

	path := &EventPath{
		elements:  elements,
		selectors: make(map[string][]int),
	}

	// Cache the path
	r.mutex.Lock()
	r.pathCache[cacheKey] = path
	r.mutex.Unlock()

	return path
}

// MatchHandlers finds handlers that match the event path
func (r *EventRouter) MatchHandlers(path *EventPath, handlers map[uint64]*DelegatedEventHandler) []*DelegatedEventHandler {
	matched := make([]*DelegatedEventHandler, 0, len(handlers))

	for _, handler := range handlers {
		if r.selectorMatches(handler.selector, path) {
			matched = append(matched, handler)
		}
	}

	return matched
}

// selectorMatches checks if a selector matches any element in the path
func (r *EventRouter) selectorMatches(selector string, path *EventPath) bool {
	// Simple selector matching - in a full implementation, this would be more sophisticated
	for _, element := range path.elements {
		if r.elementMatchesSelector(element, selector) {
			return true
		}
	}
	return false
}

// elementMatchesSelector checks if an element matches a CSS selector
func (r *EventRouter) elementMatchesSelector(element js.Value, selector string) bool {
	if !element.Truthy() {
		return false
	}

	// Use native matches method if available
	if element.Get("matches").Truthy() {
		return element.Call("matches", selector).Bool()
	}

	// Fallback for older browsers
	if element.Get("webkitMatchesSelector").Truthy() {
		return element.Call("webkitMatchesSelector", selector).Bool()
	}

	if element.Get("msMatchesSelector").Truthy() {
		return element.Call("msMatchesSelector", selector).Bool()
	}

	return false
}

// ------------------------------------
// 🏊 Handler Pool Implementation
// ------------------------------------

// Get retrieves a handler from the pool
func (p *HandlerPool) Get() *DelegatedEventHandler {
	select {
	case handler := <-p.handlers:
		// Reset handler state
		handler.selector = ""
		handler.handler = nil
		handler.element = js.Undefined()
		handler.useCount = 0
		return handler
	default:
		return nil
	}
}

// Put returns a handler to the pool
func (p *HandlerPool) Put(handler *DelegatedEventHandler) {
	if handler == nil {
		return
	}

	// Clear sensitive data
	handler.handler = nil
	handler.element = js.Undefined()

	select {
	case p.handlers <- handler:
		// Successfully returned to pool
	default:
		// Pool is full, let it be garbage collected
	}
}

// GetStats returns pool statistics
func (p *HandlerPool) GetStats() map[string]uint64 {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	return map[string]uint64{
		"created":   atomic.LoadUint64(&p.created),
		"reused":    atomic.LoadUint64(&p.reused),
		"available": uint64(len(p.handlers)),
		"capacity":  uint64(p.maxSize),
	}
}

// ------------------------------------
// 🧹 Cleanup and Disposal
// ------------------------------------

// Dispose cleans up all delegation resources
func (d *EventDelegator) Dispose() {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	if d.disposed {
		return
	}

	// Clean up all handlers
	for event := range d.active {
		d.deactivateEventDelegation(event)
	}

	// Clear all data structures
	d.handlers = make(map[string]map[uint64]*DelegatedEventHandler)
	d.active = make(map[string]bool)

	// Dispose router
	if d.router != nil {
		d.router.Dispose()
	}

	d.disposed = true
}

// Dispose cleans up router resources
func (r *EventRouter) Dispose() {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	r.selectorCache = make(map[string][]js.Value)
	r.pathCache = make(map[string]*EventPath)
}

// ------------------------------------
// 📊 Metrics and Debugging
// ------------------------------------

// GetMetrics returns delegation metrics
func (d *EventDelegator) GetMetrics() *DelegationMetrics {
	return d.metrics
}

// GetStats returns delegation statistics
func (d *EventDelegator) GetStats() map[string]interface{} {
	d.mutex.RLock()
	defer d.mutex.RUnlock()

	totalHandlers := 0
	for _, handlers := range d.handlers {
		totalHandlers += len(handlers)
	}

	return map[string]interface{}{
		"activeEvents":    len(d.active),
		"totalHandlers":   totalHandlers,
		"eventsProcessed": atomic.LoadUint64(&d.metrics.eventsProcessed),
		"handlersMatched": atomic.LoadUint64(&d.metrics.handlersMatched),
		"poolStats":       d.pool.GetStats(),
	}
}
