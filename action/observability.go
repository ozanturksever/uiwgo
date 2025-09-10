package action

import (
	"sync"
	"time"

	"github.com/ozanturksever/logutil"
)

// DevLogEntry represents a development logger entry
type DevLogEntry struct {
	ActionType      string
	TraceID         string
	Source          string
	SubscriberCount int
	Duration        time.Duration
	Error           error
	Timestamp       time.Time
}

// DebugRingBufferEntry represents an entry in the debug ring buffer
type DebugRingBufferEntry struct {
	ActionType string
	Payload    any
	TraceID    string
	Source     string
	Timestamp  time.Time
	Meta       map[string]any
}

// AnalyticsEvent represents an analytics event
type AnalyticsEvent struct {
	ActionType string
	TraceID    string
	Source     string
	Timestamp  time.Time
	Meta       map[string]any
}

// AnalyticsOption represents configuration for analytics tap
type AnalyticsOption interface {
	applyAnalytics(*analyticsOptions)
}

type analyticsOptions struct {
	filter func(any) bool
}

// Enhanced error handler signature for observability
type ErrorHandler func(ctx Context, err error, recovered any)

// DevLogger provides configurable development logging for action dispatches
type DevLogger struct {
	enabled bool
	handler func(DevLogEntry)
	mu      sync.RWMutex
}

// DebugRingBuffer maintains a fixed-size circular buffer of actions per type
type DebugRingBuffer struct {
	buffers map[string]*actionBuffer
	size    int
	mu      sync.RWMutex
}

// actionBuffer is a circular buffer for a specific action type
type actionBuffer struct {
	entries []DebugRingBufferEntry
	index   int
	full    bool
	mu      sync.RWMutex
}

// AnalyticsTap provides filtered observation of all actions for analytics
type AnalyticsTap struct {
	subscription Subscription
	handler      func(AnalyticsEvent)
	filter       func(any) bool
}

// observabilityManager handles all observability features for a bus
type observabilityManager struct {
	devLogger       *DevLogger
	debugBuffer     *DebugRingBuffer
	enhancedOnError ErrorHandler
	mu              sync.RWMutex
}

// Global observability managers per bus instance
var (
	busObservability = make(map[*busImpl]*observabilityManager)
	busObsMutex      sync.RWMutex
)

// getObservabilityManager gets or creates an observability manager for a bus
func getObservabilityManager(bus *busImpl) *observabilityManager {
	busObsMutex.RLock()
	if obs, exists := busObservability[bus]; exists {
		busObsMutex.RUnlock()
		return obs
	}
	busObsMutex.RUnlock()

	busObsMutex.Lock()
	defer busObsMutex.Unlock()

	// Double-check after acquiring write lock
	if obs, exists := busObservability[bus]; exists {
		return obs
	}

	obs := &observabilityManager{
		devLogger:   &DevLogger{},
		debugBuffer: &DebugRingBuffer{buffers: make(map[string]*actionBuffer)},
	}
	busObservability[bus] = obs
	return obs
}

// EnableDevLogger enables development logging for a bus
func EnableDevLogger(bus Bus, logger func(DevLogEntry)) {
	if busImpl, ok := bus.(*busImpl); ok {
		obs := getObservabilityManager(busImpl)
		obs.devLogger.mu.Lock()
		defer obs.devLogger.mu.Unlock()
		obs.devLogger.enabled = true
		obs.devLogger.handler = logger
	}
}

// DisableDevLogger disables development logging for a bus
func DisableDevLogger(bus Bus) {
	if busImpl, ok := bus.(*busImpl); ok {
		obs := getObservabilityManager(busImpl)
		obs.devLogger.mu.Lock()
		defer obs.devLogger.mu.Unlock()
		obs.devLogger.enabled = false
		obs.devLogger.handler = nil
	}
}

// EnableDebugRingBuffer enables debug ring buffer for a bus with specified size
func EnableDebugRingBuffer(bus Bus, size int) {
	if busImpl, ok := bus.(*busImpl); ok {
		obs := getObservabilityManager(busImpl)
		obs.debugBuffer.mu.Lock()
		defer obs.debugBuffer.mu.Unlock()
		obs.debugBuffer.size = size
		// Clear existing buffers to apply new size
		obs.debugBuffer.buffers = make(map[string]*actionBuffer)
	}
}

// GetDebugRingBufferEntries retrieves entries from the debug ring buffer for an action type
func GetDebugRingBufferEntries(bus Bus, actionType string) []DebugRingBufferEntry {
	if busImpl, ok := bus.(*busImpl); ok {
		obs := getObservabilityManager(busImpl)
		obs.debugBuffer.mu.RLock()
		defer obs.debugBuffer.mu.RUnlock()

		buffer, exists := obs.debugBuffer.buffers[actionType]
		if !exists {
			return []DebugRingBufferEntry{}
		}

		buffer.mu.RLock()
		defer buffer.mu.RUnlock()

		var result []DebugRingBufferEntry
		if buffer.full {
			// Buffer is full, read from current index to end, then from start to current index
			for i := buffer.index; i < len(buffer.entries); i++ {
				result = append(result, buffer.entries[i])
			}
			for i := 0; i < buffer.index; i++ {
				result = append(result, buffer.entries[i])
			}
		} else {
			// Buffer not full, read from start to current index
			for i := 0; i < buffer.index; i++ {
				result = append(result, buffer.entries[i])
			}
		}

		return result
	}
	return []DebugRingBufferEntry{}
}

// ClearDebugRingBuffer clears the debug ring buffer for an action type
func ClearDebugRingBuffer(bus Bus, actionType string) {
	if busImpl, ok := bus.(*busImpl); ok {
		obs := getObservabilityManager(busImpl)
		obs.debugBuffer.mu.Lock()
		defer obs.debugBuffer.mu.Unlock()

		if buffer, exists := obs.debugBuffer.buffers[actionType]; exists {
			buffer.mu.Lock()
			defer buffer.mu.Unlock()
			buffer.index = 0
			buffer.full = false
			// Clear entries
			for i := range buffer.entries {
				buffer.entries[i] = DebugRingBufferEntry{}
			}
		}
	}
}

// NewAnalyticsTap creates a new analytics tap for observing actions
func NewAnalyticsTap(bus Bus, handler func(AnalyticsEvent), opts ...AnalyticsOption) Subscription {
	// Apply options
	options := &analyticsOptions{}
	for _, opt := range opts {
		opt.applyAnalytics(options)
	}

	tap := &AnalyticsTap{
		handler: handler,
		filter:  options.filter,
	}

	// Subscribe to all actions
	tap.subscription = bus.SubscribeAny(func(action any) error {
		// Apply filter if provided
		if tap.filter != nil && !tap.filter(action) {
			return nil
		}

		// Extract event data
		event := AnalyticsEvent{
			Timestamp: time.Now(),
		}

		switch act := action.(type) {
		case Action[string]:
			event.ActionType = act.Type
			event.TraceID = act.TraceID
			event.Source = act.Source
			event.Meta = act.Meta
		case Action[any]:
			event.ActionType = act.Type
			event.TraceID = act.TraceID
			event.Source = act.Source
			event.Meta = act.Meta
		default:
			event.ActionType = "unknown"
		}

		// Call handler
		tap.handler(event)
		return nil
	})

	return tap
}

// Dispose disposes the analytics tap
func (at *AnalyticsTap) Dispose() error {
	if at.subscription != nil {
		return at.subscription.Dispose()
	}
	return nil
}

// IsActive returns whether the analytics tap is active
func (at *AnalyticsTap) IsActive() bool {
	if at.subscription != nil {
		return at.subscription.IsActive()
	}
	return false
}

// WithAnalyticsFilter sets a filter for analytics events
func WithAnalyticsFilter(filter func(any) bool) AnalyticsOption {
	return &analyticsFilterOption{filter: filter}
}

type analyticsFilterOption struct {
	filter func(any) bool
}

func (o *analyticsFilterOption) applyAnalytics(opts *analyticsOptions) {
	opts.filter = o.filter
}

// logDevEntry logs a development entry if dev logger is enabled
func (obs *observabilityManager) logDevEntry(actionType string, ctx Context, subscriberCount int, duration time.Duration, err error) {
	obs.devLogger.mu.RLock()
	defer obs.devLogger.mu.RUnlock()

	if obs.devLogger.enabled && obs.devLogger.handler != nil {
		entry := DevLogEntry{
			ActionType:      actionType,
			TraceID:         ctx.TraceID,
			Source:          ctx.Source,
			SubscriberCount: subscriberCount,
			Duration:        duration,
			Error:           err,
			Timestamp:       time.Now(),
		}
		obs.devLogger.handler(entry)
	}
}

// recordDebugEntry records an entry in the debug ring buffer
func (obs *observabilityManager) recordDebugEntry(actionType string, action any, ctx Context) {
	obs.debugBuffer.mu.RLock()
	size := obs.debugBuffer.size
	obs.debugBuffer.mu.RUnlock()

	if size <= 0 {
		return // Debug buffer not enabled
	}

	obs.debugBuffer.mu.Lock()
	buffer, exists := obs.debugBuffer.buffers[actionType]
	if !exists {
		buffer = &actionBuffer{
			entries: make([]DebugRingBufferEntry, size),
		}
		obs.debugBuffer.buffers[actionType] = buffer
	}
	obs.debugBuffer.mu.Unlock()

	// Extract payload
	var payload any
	switch act := action.(type) {
	case Action[string]:
		payload = act.Payload
	case Action[any]:
		payload = act.Payload
	default:
		payload = action
	}

	entry := DebugRingBufferEntry{
		ActionType: actionType,
		Payload:    payload,
		TraceID:    ctx.TraceID,
		Source:     ctx.Source,
		Timestamp:  time.Now(),
		Meta:       ctx.Meta,
	}

	buffer.mu.Lock()
	defer buffer.mu.Unlock()

	buffer.entries[buffer.index] = entry
	buffer.index++

	if buffer.index >= len(buffer.entries) {
		buffer.index = 0
		buffer.full = true
	}
}

// callEnhancedErrorHandler calls the enhanced error handler if set
func (obs *observabilityManager) callEnhancedErrorHandler(ctx Context, err error, recovered any) {
	obs.mu.RLock()
	handler := obs.enhancedOnError
	obs.mu.RUnlock()

	if handler != nil {
		handler(ctx, err, recovered)
	}
}

// setEnhancedErrorHandler sets the enhanced error handler
func (obs *observabilityManager) setEnhancedErrorHandler(handler ErrorHandler) {
	obs.mu.Lock()
	defer obs.mu.Unlock()
	obs.enhancedOnError = handler
}

// instrumentDispatch instruments a dispatch with observability features
func instrumentDispatch(bus *busImpl, actionType string, action any, ctx Context, subscriberCount int, dispatchFunc func() error) error {
	obs := getObservabilityManager(bus)

	// Record in debug buffer
	obs.recordDebugEntry(actionType, action, ctx)

	// Measure dispatch duration
	start := time.Now()
	err := dispatchFunc()
	duration := time.Since(start)

	// Log development entry
	obs.logDevEntry(actionType, ctx, subscriberCount, duration, err)

	return err
}

// Enhanced error handling - this will be integrated into bus.go
func handleEnhancedError(bus *busImpl, ctx Context, err error, recovered any) {
	// Call enhanced error handler if set on the bus
	bus.mu.RLock()
	enhancedHandler := bus.enhancedErrorHandler
	legacyHandler := bus.errorHandler
	bus.mu.RUnlock()

	if enhancedHandler != nil {
		enhancedHandler(ctx, err, recovered)
	}

	// Also call legacy error handler if set
	if legacyHandler != nil {
		if dispatchErr, ok := err.(*dispatchError); ok {
			legacyHandler(dispatchErr)
		} else {
			legacyHandler(err)
		}
	}
}

// SetEnhancedErrorHandler sets an enhanced error handler for a bus
func SetEnhancedErrorHandler(bus Bus, handler ErrorHandler) {
	if busImpl, ok := bus.(*busImpl); ok {
		obs := getObservabilityManager(busImpl)
		obs.setEnhancedErrorHandler(handler)
	}
}

// Additional utility functions for better observability

// GetObservabilityStats returns observability statistics for a bus
func GetObservabilityStats(bus Bus) ObservabilityStats {
	stats := ObservabilityStats{}

	if busImpl, ok := bus.(*busImpl); ok {
		obs := getObservabilityManager(busImpl)

		obs.devLogger.mu.RLock()
		stats.DevLoggerEnabled = obs.devLogger.enabled
		obs.devLogger.mu.RUnlock()

		obs.debugBuffer.mu.RLock()
		stats.DebugBufferSize = obs.debugBuffer.size
		stats.DebugBufferActionTypes = len(obs.debugBuffer.buffers)
		obs.debugBuffer.mu.RUnlock()

		obs.mu.RLock()
		stats.EnhancedErrorHandlerSet = obs.enhancedOnError != nil
		obs.mu.RUnlock()
	}

	return stats
}

// ObservabilityStats provides statistics about observability features
type ObservabilityStats struct {
	DevLoggerEnabled        bool
	DebugBufferSize         int
	DebugBufferActionTypes  int
	EnhancedErrorHandlerSet bool
}

// LogObservabilityEvent logs a general observability event
func LogObservabilityEvent(level string, message string, meta map[string]any) {
	logutil.Logf("[OBSERVABILITY-%s] %s %v", level, message, meta)
}
