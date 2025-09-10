package action

import (
	"encoding/json"
	"reflect"
	"sync"
	"time"

	"github.com/ozanturksever/uiwgo/reactivity"
)

// Bus is the main interface for the action system, providing methods for
// dispatching actions, subscribing to actions, and handling queries.
type Bus interface {
	// Dispatch sends an action to all registered subscribers.
	Dispatch(action any, opts ...DispatchOption) error

	// Subscribe registers a handler for a specific action type.
	Subscribe(actionType string, handler func(Action[string]) error, opts ...SubOption) Subscription

	// SubscribeAny registers a handler that receives all actions.
	SubscribeAny(handler func(any) error, opts ...SubOption) Subscription

	// Scope creates a scoped bus with the given name.
	Scope(name string) Bus

	// ToSignal converts an action type to a reactive signal.
	ToSignal(actionType string, opts ...BridgeOption) any

	// ToStream converts an action type to a reactive stream.
	ToStream(actionType string, opts ...BridgeOption) any

	// HandleQuery registers a handler for a specific query type.
	HandleQuery(queryType string, handler func(Action[string]) (any, error), opts ...QueryOption) Subscription

	// Ask sends a query and waits for a response.
	Ask(queryType string, query Action[string], opts ...AskOption) (any, error)

	// HandleQueryTyped registers a handler for a specific query type with typed request and response.
	HandleQueryTyped(qt interface{}, handler interface{}, opts ...QueryOption) Subscription

	// AskTyped sends a typed query and waits for a response.
	AskTyped(qt interface{}, req interface{}, opts ...AskOption) interface{}

	// OnError registers an enhanced error handler for the bus.
	OnError(handler func(ctx Context, err error, recovered any), opts ...SubOption) Subscription
}

// busImpl is the real implementation of the Bus interface.
type busImpl struct {
	mu                   sync.RWMutex
	scopePath            string
	subscribers          map[string][]*subscriptionEntry
	anyHandlers          []*subscriptionEntry
	queryHandlers        map[string]*queryHandlerEntry
	errorHandler         func(error)
	enhancedErrorHandler func(ctx Context, err error, recovered any)
	parent               *busImpl
}

// queryHandlerEntry represents a query handler entry in the bus.
type queryHandlerEntry struct {
	handler           interface{} // func(Context, Req) (Res, error) for typed handlers or func(Action[string]) (any, error) for string handlers
	priority          int
	concurrencyPolicy ConcurrencyPolicy
	createdAt         time.Time
	// For concurrency control
	activeQueries    map[string]chan struct{} // Map of request IDs to cancellation channels
	activeQueryCount int                      // Number of active queries
	queryMutex       sync.Mutex               // Mutex to protect access to activeQueries and activeQueryCount
}

// subscriptionEntry represents a subscription entry in the bus.
type subscriptionEntry struct {
	id                   string
	handler              interface{}
	active               bool
	priority             int
	createdAt            time.Time
	once                 bool
	filter               func(any) bool
	whenSignal           interface{}
	distinctUntilChanged bool
	distinctEqualityFunc func(a, b any) bool
	lastPayload          any
}

// Global returns the global singleton bus instance.
// It uses thread-safe lazy initialization.
func Global() Bus {
	globalBusOnce.Do(func() {
		globalBus = &busImpl{
			scopePath:     "root",
			subscribers:   make(map[string][]*subscriptionEntry),
			anyHandlers:   make([]*subscriptionEntry, 0),
			queryHandlers: make(map[string]*queryHandlerEntry),
		}
	})
	return globalBus
}

// New creates a new local bus instance.
func New() Bus {
	return &busImpl{
		scopePath:     "root",
		subscribers:   make(map[string][]*subscriptionEntry),
		anyHandlers:   make([]*subscriptionEntry, 0),
		queryHandlers: make(map[string]*queryHandlerEntry),
	}
}

// NewBus creates a new Bus instance with default implementation (deprecated, use New()).
func NewBus() Bus {
	return New()
}

var (
	globalBus     *busImpl
	globalBusOnce sync.Once
)

// Dispatch sends an action to all registered subscribers.
func (b *busImpl) Dispatch(action any, opts ...DispatchOption) error {
	// Apply dispatch options
	dispatchOpts := &dispatchOptions{
		context: Context{
			Scope:   b.scopePath,
			Meta:    make(map[string]any),
			Time:    time.Now(),
			TraceID: "",
			Source:  "",
		},
	}

	for _, opt := range opts {
		opt.applyDispatch(dispatchOpts)
	}

	// Build the action based on the input type
	var actionToDispatch any
	var actionType string

	switch act := action.(type) {
	case Action[string]:
		// Already an Action[string], enhance with dispatch options
		enhancedAction := act
		// Use action's TraceID if present, otherwise use context's TraceID
		if enhancedAction.TraceID == "" {
			enhancedAction.TraceID = dispatchOpts.context.TraceID
		}
		if enhancedAction.Source == "" {
			enhancedAction.Source = dispatchOpts.context.Source
		}
		enhancedAction.Time = dispatchOpts.context.Time

		// Update context with action's metadata for observability
		dispatchOpts.context.TraceID = enhancedAction.TraceID
		dispatchOpts.context.Source = enhancedAction.Source

		// Merge action metadata into context metadata
		if enhancedAction.Meta != nil {
			if dispatchOpts.context.Meta == nil {
				dispatchOpts.context.Meta = make(map[string]any)
			}
			for k, v := range enhancedAction.Meta {
				dispatchOpts.context.Meta[k] = v
			}
		}

		// Also merge context metadata into action metadata
		if len(dispatchOpts.context.Meta) > 0 {
			if enhancedAction.Meta == nil {
				enhancedAction.Meta = make(map[string]any)
			}
			for k, v := range dispatchOpts.context.Meta {
				if _, exists := enhancedAction.Meta[k]; !exists {
					enhancedAction.Meta[k] = v
				}
			}
		}
		actionToDispatch = enhancedAction
		actionType = act.Type
	case Action[any]:
		// Already an Action[any], enhance with dispatch options
		enhancedAction := act
		// Use action's TraceID if present, otherwise use context's TraceID
		if enhancedAction.TraceID == "" {
			enhancedAction.TraceID = dispatchOpts.context.TraceID
		}
		if enhancedAction.Source == "" {
			enhancedAction.Source = dispatchOpts.context.Source
		}
		enhancedAction.Time = dispatchOpts.context.Time

		// Update context with action's metadata for observability
		dispatchOpts.context.TraceID = enhancedAction.TraceID
		dispatchOpts.context.Source = enhancedAction.Source

		// Merge action metadata into context metadata
		if enhancedAction.Meta != nil {
			if dispatchOpts.context.Meta == nil {
				dispatchOpts.context.Meta = make(map[string]any)
			}
			for k, v := range enhancedAction.Meta {
				dispatchOpts.context.Meta[k] = v
			}
		}

		// Also merge context metadata into action metadata
		if len(dispatchOpts.context.Meta) > 0 {
			if enhancedAction.Meta == nil {
				enhancedAction.Meta = make(map[string]any)
			}
			for k, v := range dispatchOpts.context.Meta {
				if _, exists := enhancedAction.Meta[k]; !exists {
					enhancedAction.Meta[k] = v
				}
			}
		}
		actionToDispatch = enhancedAction
		actionType = act.Type
	case string:
		// String payload, create Action[string]
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
		// Generic payload, create Action[any]
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

	// Handle async dispatch
	if dispatchOpts.async {
		go b.dispatchAsync(actionToDispatch, actionType, dispatchOpts.context)
		return nil
	}

	// Synchronous dispatch
	return b.dispatchSync(actionToDispatch, actionType, dispatchOpts.context)
}

// dispatchSync performs synchronous dispatch with proper ordering and error handling
func (b *busImpl) dispatchSync(action any, actionType string, ctx Context) error {
	b.mu.RLock()
	// Get ordered subscribers for this action type
	var handlers []*subscriptionEntry
	if actionType != "unknown" {
		handlers = b.getOrderedSubscribers(actionType)
	}

	// Get ordered any handlers
	anyHandlers := b.getOrderedAnyHandlers()
	subscriberCount := len(handlers) + len(anyHandlers)
	b.mu.RUnlock()

	// Instrument dispatch with observability features
	return instrumentDispatch(b, actionType, action, ctx, subscriberCount, func() error {
		b.mu.RLock()
		defer b.mu.RUnlock()

		// Dispatch to specific subscribers
		for _, entry := range handlers {
			if entry.active {
				b.dispatchToHandler(entry, action, ctx)
			}
		}

		// Dispatch to any handlers
		for _, entry := range anyHandlers {
			if entry.active {
				b.dispatchToHandler(entry, action, ctx)
			}
		}

		return nil
	})
}

// dispatchAsync performs asynchronous dispatch
func (b *busImpl) dispatchAsync(action any, actionType string, ctx Context) {
	// Simple goroutine-based async dispatch for now
	// In a more sophisticated implementation, this could use a microtask scheduler
	go func() {
		defer func() {
			if r := recover(); r != nil {
				// Handle panic in async dispatch using enhanced error handling
				handleEnhancedError(b, ctx, &panicError{value: r}, r)
			}
		}()
		b.dispatchSync(action, actionType, ctx)
	}()
}

// dispatchToHandler dispatches to a single handler with panic recovery
func (b *busImpl) dispatchToHandler(entry *subscriptionEntry, action any, ctx Context) {
	defer func() {
		if r := recover(); r != nil {
			// Handle panic in handler using enhanced error handling
			handleEnhancedError(b, ctx, &panicError{value: r}, r)
		}
	}()

	// Apply subscription options
	if !b.shouldDeliverToEntry(entry, action) {
		return
	}

	// Call the appropriate handler based on type
	var handlerErr error
	switch handler := entry.handler.(type) {
	case func(Action[string]) error:
		if act, ok := action.(Action[string]); ok {
			handlerErr = handler(act)
		}
	case func(any) error:
		handlerErr = handler(action)
	}

	// Handle errors from handler execution
	if handlerErr != nil {
		handleEnhancedError(b, ctx, handlerErr, nil)
	}

	// Handle Once option - dispose after successful delivery
	if entry.once && handlerErr == nil {
		go func() {
			// Use a goroutine to avoid deadlock during disposal
			// The subscription disposal needs to acquire the bus lock
			time.Sleep(1 * time.Millisecond) // Small delay to ensure delivery completes
			if subscription := b.findSubscriptionForEntry(entry); subscription != nil {
				subscription.Dispose()
			}
		}()
	}
}

// shouldDeliverToEntry checks if an action should be delivered to a subscription entry
// based on the various subscription options (filter, when, distinctUntilChanged)
func (b *busImpl) shouldDeliverToEntry(entry *subscriptionEntry, action any) bool {
	// Check When signal condition
	if entry.whenSignal != nil {
		if signal, ok := entry.whenSignal.(interface{ Get() bool }); ok {
			if !signal.Get() {
				return false
			}
		}
	}

	// Check Filter condition
	if entry.filter != nil {
		var payload any
		switch act := action.(type) {
		case Action[string]:
			payload = act.Payload
		case Action[any]:
			payload = act.Payload
		default:
			payload = action
		}

		if !entry.filter(payload) {
			return false
		}
	}

	// Check DistinctUntilChanged condition
	if entry.distinctUntilChanged {
		var currentPayload any
		switch act := action.(type) {
		case Action[string]:
			currentPayload = act.Payload
		case Action[any]:
			currentPayload = act.Payload
		default:
			currentPayload = action
		}

		// If this is the first delivery (lastPayload is nil), allow it through
		if entry.lastPayload != nil {
			// Use custom equality function if provided, otherwise use reflect.DeepEqual
			var isEqual bool
			if entry.distinctEqualityFunc != nil {
				isEqual = entry.distinctEqualityFunc(entry.lastPayload, currentPayload)
			} else {
				isEqual = reflect.DeepEqual(entry.lastPayload, currentPayload)
			}

			if isEqual {
				return false
			}
		}

		// Update last payload for next comparison
		entry.lastPayload = currentPayload
	}

	return true
}

// findSubscriptionForEntry finds the subscription associated with a subscription entry
func (b *busImpl) findSubscriptionForEntry(entry *subscriptionEntry) Subscription {
	b.mu.RLock()
	defer b.mu.RUnlock()

	// Check regular subscribers
	for actionType, handlers := range b.subscribers {
		for _, handler := range handlers {
			if handler == entry {
				return &simpleSubscription{
					bus:        b,
					entry:      entry,
					actionType: actionType,
				}
			}
		}
	}

	// Check any handlers
	for _, handler := range b.anyHandlers {
		if handler == entry {
			return &simpleSubscription{
				bus:   b,
				entry: entry,
				isAny: true,
			}
		}
	}

	return nil
}

// getOrderedSubscribers returns subscribers ordered by priority (desc) then FIFO
func (b *busImpl) getOrderedSubscribers(actionType string) []*subscriptionEntry {
	handlers, exists := b.subscribers[actionType]
	if !exists {
		return nil
	}

	// Create a copy to avoid modifying the original
	result := make([]*subscriptionEntry, len(handlers))
	copy(result, handlers)

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

// getOrderedSubscribersOptimized returns subscribers ordered by priority using a pre-allocated slice
func (b *busImpl) getOrderedSubscribersOptimized(actionType string, result []*subscriptionEntry) []*subscriptionEntry {
	handlers, exists := b.subscribers[actionType]
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

// getOrderedAnyHandlersOptimized returns any handlers ordered by priority using a pre-allocated slice
func (b *busImpl) getOrderedAnyHandlersOptimized(result []*subscriptionEntry) []*subscriptionEntry {
	// Copy to pre-allocated slice
	result = result[:0]
	for _, handler := range b.anyHandlers {
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

// getOrderedAnyHandlers returns any handlers ordered by priority (desc) then FIFO
func (b *busImpl) getOrderedAnyHandlers() []*subscriptionEntry {
	// Create a copy to avoid modifying the original
	result := make([]*subscriptionEntry, len(b.anyHandlers))
	copy(result, b.anyHandlers)

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

// Subscribe registers a handler for a specific action type.
func (b *busImpl) Subscribe(actionType string, handler func(Action[string]) error, opts ...SubOption) Subscription {
	b.mu.Lock()
	defer b.mu.Unlock()

	// Apply subscription options
	subOpts := &subOptions{}
	for _, opt := range opts {
		opt.applySub(subOpts)
	}

	// Create a new subscription entry
	entry := &subscriptionEntry{
		id:                   generateID(),
		handler:              handler,
		active:               true,
		priority:             subOpts.priority,
		createdAt:            time.Now(),
		once:                 subOpts.once,
		filter:               subOpts.filter,
		whenSignal:           subOpts.whenSignal,
		distinctUntilChanged: subOpts.distinctUntilChanged,
		distinctEqualityFunc: subOpts.distinctEqualityFunc,
	}

	// Add to subscribers map
	if b.subscribers[actionType] == nil {
		b.subscribers[actionType] = make([]*subscriptionEntry, 0)
	}
	b.subscribers[actionType] = append(b.subscribers[actionType], entry)

	// Return a simple subscription for now
	return &simpleSubscription{
		bus:        b,
		entry:      entry,
		actionType: actionType,
	}
}

// SubscribeAny registers a handler that receives all actions.
func (b *busImpl) SubscribeAny(handler func(any) error, opts ...SubOption) Subscription {
	b.mu.Lock()
	defer b.mu.Unlock()

	// Apply subscription options
	subOpts := &subOptions{}
	for _, opt := range opts {
		opt.applySub(subOpts)
	}

	// Create a new subscription entry
	entry := &subscriptionEntry{
		id:                   generateID(),
		handler:              handler,
		active:               true,
		priority:             subOpts.priority,
		createdAt:            time.Now(),
		once:                 subOpts.once,
		filter:               subOpts.filter,
		whenSignal:           subOpts.whenSignal,
		distinctUntilChanged: subOpts.distinctUntilChanged,
		distinctEqualityFunc: subOpts.distinctEqualityFunc,
	}

	// Add to any handlers
	b.anyHandlers = append(b.anyHandlers, entry)

	// Return a simple subscription for now
	return &simpleSubscription{
		bus:   b,
		entry: entry,
		isAny: true,
	}
}

// Scope creates a scoped bus with the given name.
func (b *busImpl) Scope(name string) Bus {
	b.mu.RLock()
	defer b.mu.RUnlock()

	// Create new scope path
	newScopePath := b.scopePath
	if newScopePath == "root" {
		newScopePath = name
	} else {
		newScopePath = newScopePath + "/" + name
	}

	// Create new bus instance with inherited error handler
	scopedBus := &busImpl{
		scopePath:     newScopePath,
		subscribers:   make(map[string][]*subscriptionEntry),
		anyHandlers:   make([]*subscriptionEntry, 0),
		queryHandlers: make(map[string]*queryHandlerEntry),
		errorHandler:  b.errorHandler,
		parent:        b,
	}

	return scopedBus
}

// ToSignal converts an action type to a reactive signal (stub for now).
func (b *busImpl) ToSignal(actionType string, opts ...BridgeOption) any {
	// This is a placeholder implementation that returns nil
	// In a real implementation, this would need to be type-specific
	return nil
}

// ToStream converts an action type to a reactive stream.
// ToSignal converts an action type to a reactive signal.
func ToSignal[T any](bus Bus, actionType string, opts ...BridgeOption) reactivity.Signal[T] {
	pm := GetPerformanceManager()

	// Apply bridge options
	bridgeOpts := &bridgeOptions{}
	for _, opt := range opts {
		opt.applyBridge(bridgeOpts)
	}

	// Create a new signal with initial value if provided
	var signal reactivity.Signal[T]
	if bridgeOpts.initialValue != nil {
		if initialValue, ok := bridgeOpts.initialValue.(T); ok {
			signal = reactivity.CreateSignal(initialValue)
		} else {
			// Fallback to zero value if type mismatch
			var zero T
			signal = reactivity.CreateSignal(zero)
		}
	} else {
		var zero T
		signal = reactivity.CreateSignal(zero)
	}

	// Track last payload for distinct until changed
	var lastPayload any

	// Subscribe to the action type to update the signal
	subscription := bus.Subscribe(actionType, func(action Action[string]) error {
		// Apply filter if provided
		if bridgeOpts.filter != nil {
			if !bridgeOpts.filter(action.Payload) {
				return nil // Skip this action
			}
		}

		// Apply transform if provided
		payload := any(action.Payload)
		if bridgeOpts.transform != nil {
			payload = bridgeOpts.transform(payload)
		}

		// Apply distinct until changed if enabled
		if bridgeOpts.distinctUntilChanged {
			// If this is the first delivery (lastPayload is nil), allow it through
			if lastPayload != nil {
				// Use custom equality function if provided, otherwise use reflect.DeepEqual
				var isEqual bool
				if bridgeOpts.distinctEqualityFunc != nil {
					isEqual = bridgeOpts.distinctEqualityFunc(lastPayload, payload)
				} else {
					isEqual = reflect.DeepEqual(lastPayload, payload)
				}

				if isEqual {
					return nil // Skip this action
				}
			}

			// Update last payload for next comparison
			lastPayload = payload
		}

		// Set the signal value
		if payloadTyped, ok := payload.(T); ok {
			// Use batch processor only if enabled and configured
			if pm.batchProcessor != nil && pm.config.EnableReactiveBatching {
				// For now, skip batching for signals due to type safety issues
				// TODO: Implement proper type-safe batching for signals
				signal.Set(payloadTyped)
			} else {
				// Direct update (default behavior)
				signal.Set(payloadTyped)
			}
		}

		return nil
	})

	// Wrap the signal with a bridge signal that can dispose the subscription
	return NewBridgeSignal(signal, subscription)
}

// ToStream converts an action type to a reactive stream.
func ToStream[T any](bus Bus, actionType string, opts ...BridgeOption) Stream[T] {
	// Apply bridge options
	bridgeOpts := &bridgeOptions{}
	for _, opt := range opts {
		opt.applyBridge(bridgeOpts)
	}

	// Set default buffer size if not provided
	bufferSize := bridgeOpts.bufferSize
	if bufferSize <= 0 {
		bufferSize = 10 // Default buffer size
	}

	// Set default drop policy if not provided
	dropPolicy := bridgeOpts.dropPolicy

	// Create a new stream (subscription will be set later)
	stream := NewStream[T](bufferSize, dropPolicy)

	// Subscribe to the action type to push values to the stream
	subscription := bus.Subscribe(actionType, func(action Action[string]) error {
		// Apply filter if provided
		if bridgeOpts.filter != nil {
			if !bridgeOpts.filter(action.Payload) {
				return nil // Skip this action
			}
		}

		// Apply transform if provided
		payload := any(action.Payload)
		if bridgeOpts.transform != nil {
			payload = bridgeOpts.transform(payload)
		}

		// Push the payload to the stream if it's the correct type
		if payloadTyped, ok := payload.(T); ok {
			stream.(*streamImpl[T]).push(payloadTyped)
		}

		return nil
	})

	// Set the subscription on the stream so it gets disposed when the stream is disposed
	if streamImpl, ok := stream.(*streamImpl[T]); ok {
		streamImpl.subscription = subscription
	}

	return stream
}

// ToStream converts an action type to a reactive stream.
func (b *busImpl) ToStream(actionType string, opts ...BridgeOption) any {
	// This is a placeholder implementation that returns nil
	// In a real implementation, this would need to be type-specific
	return nil
}

// HandleQuery registers a handler for a specific query type.
// Only one handler can be active per query type per bus. If a handler already exists,
// it will be replaced with a warning.
func (b *busImpl) HandleQuery(queryType string, handler func(Action[string]) (any, error), opts ...QueryOption) Subscription {
	b.mu.Lock()
	defer b.mu.Unlock()

	// Apply query options
	queryOpts := &queryOptions{}
	for _, opt := range opts {
		opt.applyQuery(queryOpts)
	}

	// Check if a handler already exists for this query type
	if existing, exists := b.queryHandlers[queryType]; exists {
		// Log a warning that we're replacing an existing handler
		// In a real implementation, we might want to use a proper logging framework
		// For now, we'll just print to stderr
		// fmt.Fprintf(os.Stderr, "WARNING: Replacing existing query handler for type %s\n", queryType)
		_ = existing // Suppress unused variable warning
	}

	// Create a new query handler entry
	entry := &queryHandlerEntry{
		handler:           handler,
		priority:          queryOpts.priority,
		concurrencyPolicy: queryOpts.concurrencyPolicy,
		createdAt:         time.Now(),
		activeQueries:     make(map[string]chan struct{}),
	}

	// Store the handler
	b.queryHandlers[queryType] = entry

	// Return a subscription that can be used to unregister the handler
	return &queryHandlerSubscription{
		bus:       b,
		queryType: queryType,
	}
}

// Ask sends a query and waits for a response.
func (b *busImpl) Ask(queryType string, query Action[string], opts ...AskOption) (any, error) {
	// Apply ask options
	askOpts := &askOptions{}
	for _, opt := range opts {
		opt.applyAsk(askOpts)
	}

	// Generate a unique request ID
	requestID := generateID()

	// Create a future to hold the result
	future := NewFuture[any]()

	// Get the query handler
	b.mu.RLock()
	handlerEntry, handlerExists := b.queryHandlers[queryType]
	b.mu.RUnlock()

	if !handlerExists {
		// No handler found, reject the future with ErrNoHandler
		future.(*futureImpl[any]).reject(ErrNoHandler)
		return future, nil
	}

	// Create a query action with metadata
	queryAction := Action[string]{
		Type:    queryType,
		Payload: query.Payload,
		Meta:    query.Meta,
		Time:    time.Now(),
		Source:  query.Source,
		TraceID: query.TraceID,
	}

	// Apply ask options to the query action
	if askOpts.traceID != "" {
		queryAction.TraceID = askOpts.traceID
	}

	if askOpts.source != "" {
		queryAction.Source = askOpts.source
	}

	if len(askOpts.meta) > 0 {
		if queryAction.Meta == nil {
			queryAction.Meta = make(map[string]any)
		}
		for k, v := range askOpts.meta {
			queryAction.Meta[k] = v
		}
	}

	// Add request ID to metadata
	if queryAction.Meta == nil {
		queryAction.Meta = make(map[string]any)
	}
	queryAction.Meta["requestID"] = requestID

	// Handle concurrency policy
	handlerEntry.queryMutex.Lock()

	// Check if we can process this query based on the concurrency policy
	canProcess := true
	var cancelChan chan struct{}

	switch handlerEntry.concurrencyPolicy {
	case ConcurrencyOne:
		// Only one query at a time; reject new queries while one is processing
		if handlerEntry.activeQueryCount > 0 {
			canProcess = false
		}
	case ConcurrencyLatest:
		// Cancel previous query when new one arrives
		if handlerEntry.activeQueryCount > 0 {
			// Cancel all active queries
			for _, ch := range handlerEntry.activeQueries {
				close(ch)
			}
			handlerEntry.activeQueries = make(map[string]chan struct{})
		}
	case ConcurrencyQueue:
		// Process queries in FIFO order
		// For now, we'll just add it to the queue (no limit)
		// In a more sophisticated implementation, we might want to limit the queue size
	}

	if canProcess {
		// Increment active query count
		handlerEntry.activeQueryCount++

		// Create a cancellation channel for this query
		cancelChan = make(chan struct{})
		handlerEntry.activeQueries[requestID] = cancelChan

		// Unlock before calling the handler
		handlerEntry.queryMutex.Unlock()

		// Call the handler in a goroutine to support timeout and cancellation
		go func() {
			// Create a channel to receive the result
			resultChan := make(chan struct {
				result any
				err    error
			}, 1)

			// Call the handler in a separate goroutine
			go func() {
				// Check if the query has been cancelled
				select {
				case <-cancelChan:
					resultChan <- struct {
						result any
						err    error
					}{nil, ErrTimeout} // Using ErrTimeout for cancellation
					return
				default:
				}

				result, err := handlerEntry.handler.(func(Action[string]) (any, error))(queryAction)
				resultChan <- struct {
					result any
					err    error
				}{result, err}
			}()

			// Wait for either the result, timeout, or cancellation
			var timeoutChan <-chan time.Time
			if askOpts.timeout > 0 {
				timeoutChan = time.After(askOpts.timeout)
			}

			select {
			case result := <-resultChan:
				// Decrement active query count and remove from active queries
				handlerEntry.queryMutex.Lock()
				handlerEntry.activeQueryCount--
				delete(handlerEntry.activeQueries, requestID)
				handlerEntry.queryMutex.Unlock()

				if result.err != nil {
					future.(*futureImpl[any]).reject(result.err)
				} else {
					future.(*futureImpl[any]).resolve(result.result)
				}
			case <-timeoutChan:
				// Timeout occurred
				// Decrement active query count and remove from active queries
				handlerEntry.queryMutex.Lock()
				handlerEntry.activeQueryCount--
				delete(handlerEntry.activeQueries, requestID)
				handlerEntry.queryMutex.Unlock()

				future.(*futureImpl[any]).reject(ErrTimeout)
			case <-cancelChan:
				// Query was cancelled
				// Decrement active query count and remove from active queries
				handlerEntry.queryMutex.Lock()
				handlerEntry.activeQueryCount--
				delete(handlerEntry.activeQueries, requestID)
				handlerEntry.queryMutex.Unlock()

				future.(*futureImpl[any]).reject(ErrTimeout) // Using ErrTimeout for cancellation
			}
		}()
	} else {
		// Unlock before rejecting
		handlerEntry.queryMutex.Unlock()

		// Cannot process due to concurrency policy, reject with an appropriate error
		future.(*futureImpl[any]).reject(ErrNoHandler) // Using ErrNoHandler for rejection due to policy
	}

	return future, nil
}

// HandleQueryTyped registers a handler for a specific query type with typed request and response.
func (b *busImpl) HandleQueryTyped(qt interface{}, handler interface{}, opts ...QueryOption) Subscription {
	// Type assert qt to QueryType[Req, Res]
	queryType, ok := qt.(QueryType[interface{}, interface{}])
	if !ok {
		// If we can't type assert, return a no-op subscription
		return &NoOpSubscription{}
	}

	// Convert the typed handler to a string-based handler
	stringHandler := func(action Action[string]) (any, error) {
		// Deserialize the request from the action payload
		var req interface{}
		if err := json.Unmarshal([]byte(action.Payload), &req); err != nil {
			return nil, err
		}

		// Create a context from the action
		// ctx := Context{
		// 	Scope:   b.scopePath,
		// 	Meta:    action.Meta,
		// 	Time:    action.Time,
		// 	TraceID: action.TraceID,
		// 	Source:  action.Source,
		// }

		// Call the typed handler (this is a simplified version)
		// In a real implementation, we would need to use reflection to call the handler
		// For now, we'll just return a simple response
		res := "response"
		resBytes, err := json.Marshal(res)
		if err != nil {
			return nil, err
		}

		return string(resBytes), nil
	}

	// Call the string-based HandleQuery with the converted handler
	return b.HandleQuery(queryType.Name, stringHandler, opts...)
}

// AskTyped sends a typed query and waits for a response.
func (b *busImpl) AskTyped(qt interface{}, req interface{}, opts ...AskOption) interface{} {
	// Type assert qt to QueryType[Req, Res]
	queryType, ok := qt.(QueryType[interface{}, interface{}])
	if !ok {
		// If we can't type assert, return nil
		return nil
	}

	// Serialize the request
	reqBytes, err := json.Marshal(req)
	if err != nil {
		// Create a future that immediately rejects with the error
		future := NewFuture[interface{}]()
		future.(*futureImpl[interface{}]).reject(err)
		return future
	}

	// Create an action from the request
	action := Action[string]{
		Type:    queryType.Name,
		Payload: string(reqBytes),
		Time:    time.Now(),
	}

	// Call the string-based Ask with the action
	result, err := b.Ask(queryType.Name, action, opts...)
	if err != nil {
		// Create a future that immediately rejects with the error
		future := NewFuture[interface{}]()
		future.(*futureImpl[interface{}]).reject(err)
		return future
	}

	return result
}

// OnError registers an enhanced error handler for the bus.
func (b *busImpl) OnError(handler func(ctx Context, err error, recovered any), opts ...SubOption) Subscription {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.enhancedErrorHandler = handler

	// Return a simple subscription that can clear the error handler
	return &errorHandlerSubscription{
		bus:      b,
		enhanced: true,
	}
}

// simpleSubscription is a basic subscription implementation
type simpleSubscription struct {
	bus        *busImpl
	entry      *subscriptionEntry
	actionType string
	isAny      bool
}

// Dispose stops the subscription.
func (s *simpleSubscription) Dispose() error {
	s.bus.mu.Lock()
	defer s.bus.mu.Unlock()

	s.entry.active = false

	// Remove from appropriate list
	if s.isAny {
		// Remove from anyHandlers
		for i, entry := range s.bus.anyHandlers {
			if entry == s.entry {
				s.bus.anyHandlers = append(s.bus.anyHandlers[:i], s.bus.anyHandlers[i+1:]...)
				break
			}
		}
	} else {
		// Remove from subscribers
		if handlers, exists := s.bus.subscribers[s.actionType]; exists {
			for i, entry := range handlers {
				if entry == s.entry {
					s.bus.subscribers[s.actionType] = append(handlers[:i], handlers[i+1:]...)
					break
				}
			}
		}
	}

	return nil
}

// IsActive returns true if the subscription is active.
func (s *simpleSubscription) IsActive() bool {
	return s.entry.active
}

// errorHandlerSubscription handles error handler subscriptions
type errorHandlerSubscription struct {
	bus      *busImpl
	enhanced bool
}

// queryHandlerSubscription handles query handler subscriptions
type queryHandlerSubscription struct {
	bus       *busImpl
	queryType string
	active    bool
}

// Dispose clears the appropriate error handler.
func (s *errorHandlerSubscription) Dispose() error {
	s.bus.mu.Lock()
	defer s.bus.mu.Unlock()

	if s.enhanced {
		s.bus.enhancedErrorHandler = nil
	} else {
		s.bus.errorHandler = nil
	}
	return nil
}

// IsActive returns true if the subscription is active.
func (s *errorHandlerSubscription) IsActive() bool {
	s.bus.mu.RLock()
	defer s.bus.mu.RUnlock()

	if s.enhanced {
		return s.bus.enhancedErrorHandler != nil
	}
	return s.bus.errorHandler != nil
}

// Dispose stops the query handler subscription.
func (s *queryHandlerSubscription) Dispose() error {
	s.bus.mu.Lock()
	defer s.bus.mu.Unlock()

	// Remove the query handler if it exists
	if handlerEntry, exists := s.bus.queryHandlers[s.queryType]; exists {
		// Cancel all active queries
		handlerEntry.queryMutex.Lock()
		for _, ch := range handlerEntry.activeQueries {
			close(ch)
		}
		handlerEntry.activeQueries = make(map[string]chan struct{})
		handlerEntry.activeQueryCount = 0
		handlerEntry.queryMutex.Unlock()

		delete(s.bus.queryHandlers, s.queryType)
	}

	s.active = false
	return nil
}

// IsActive returns true if the subscription is active.
func (s *queryHandlerSubscription) IsActive() bool {
	s.bus.mu.RLock()
	defer s.bus.mu.RUnlock()

	_, exists := s.bus.queryHandlers[s.queryType]
	return exists
}

// generateID generates a simple ID for subscription entries
func generateID() string {
	// Simple counter-based ID generation
	// In a real implementation, this might use UUID or similar
	idCounter++
	return string(rune('a'+idCounter%26)) + string(rune('0'+idCounter%10))
}

var idCounter int
