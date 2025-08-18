// store_middleware.go
// Middleware system for action processing and store operations

package golid

import (
	"fmt"
	"log"
	"sync"
	"time"
)

// ------------------------------------
// 🔧 Built-in Store Middleware
// ------------------------------------

// LoggingMiddleware logs store operations for debugging
type LoggingMiddleware[T any] struct {
	name   string
	logger func(string, ...interface{})
}

// NewLoggingMiddleware creates a new logging middleware
func NewLoggingMiddleware[T any](name string) *LoggingMiddleware[T] {
	return &LoggingMiddleware[T]{
		name:   name,
		logger: log.Printf,
	}
}

// BeforeUpdate logs before store update
func (m *LoggingMiddleware[T]) BeforeUpdate(store *Store[T], oldValue, newValue T) T {
	m.logger("[%s] Store update: %v -> %v", m.name, oldValue, newValue)
	return newValue
}

// AfterUpdate logs after store update
func (m *LoggingMiddleware[T]) AfterUpdate(store *Store[T], oldValue, newValue T) {
	m.logger("[%s] Store updated successfully", m.name)
}

// ------------------------------------
// 🛡️ Validation Middleware
// ------------------------------------

// ValidationMiddleware validates store values before updates
type ValidationMiddleware[T any] struct {
	validator func(T) error
	name      string
}

// NewValidationMiddleware creates a new validation middleware
func NewValidationMiddleware[T any](name string, validator func(T) error) *ValidationMiddleware[T] {
	return &ValidationMiddleware[T]{
		validator: validator,
		name:      name,
	}
}

// BeforeUpdate validates the new value
func (m *ValidationMiddleware[T]) BeforeUpdate(store *Store[T], oldValue, newValue T) T {
	if err := m.validator(newValue); err != nil {
		panic(fmt.Sprintf("validation failed for store %s: %v", m.name, err))
	}
	return newValue
}

// AfterUpdate does nothing for validation middleware
func (m *ValidationMiddleware[T]) AfterUpdate(store *Store[T], oldValue, newValue T) {
	// No-op
}

// ------------------------------------
// 📊 Performance Middleware
// ------------------------------------

// PerformanceMiddleware tracks store operation performance
type PerformanceMiddleware[T any] struct {
	name      string
	startTime time.Time
	stats     *PerformanceStats
	mutex     sync.RWMutex
}

// PerformanceStats tracks performance metrics
type PerformanceStats struct {
	TotalOperations int64
	TotalDuration   time.Duration
	AverageDuration time.Duration
	MaxDuration     time.Duration
	MinDuration     time.Duration
}

// NewPerformanceMiddleware creates a new performance middleware
func NewPerformanceMiddleware[T any](name string) *PerformanceMiddleware[T] {
	return &PerformanceMiddleware[T]{
		name: name,
		stats: &PerformanceStats{
			MinDuration: time.Hour, // Initialize with high value
		},
	}
}

// BeforeUpdate records start time
func (m *PerformanceMiddleware[T]) BeforeUpdate(store *Store[T], oldValue, newValue T) T {
	m.startTime = time.Now()
	return newValue
}

// AfterUpdate records performance metrics
func (m *PerformanceMiddleware[T]) AfterUpdate(store *Store[T], oldValue, newValue T) {
	duration := time.Since(m.startTime)

	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.stats.TotalOperations++
	m.stats.TotalDuration += duration
	m.stats.AverageDuration = m.stats.TotalDuration / time.Duration(m.stats.TotalOperations)

	if duration > m.stats.MaxDuration {
		m.stats.MaxDuration = duration
	}

	if duration < m.stats.MinDuration {
		m.stats.MinDuration = duration
	}
}

// GetStats returns performance statistics
func (m *PerformanceMiddleware[T]) GetStats() PerformanceStats {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return *m.stats
}

// ------------------------------------
// 🔄 Transform Middleware
// ------------------------------------

// TransformMiddleware transforms values before store updates
type TransformMiddleware[T any] struct {
	transformer func(T) T
	name        string
}

// NewTransformMiddleware creates a new transform middleware
func NewTransformMiddleware[T any](name string, transformer func(T) T) *TransformMiddleware[T] {
	return &TransformMiddleware[T]{
		transformer: transformer,
		name:        name,
	}
}

// BeforeUpdate applies transformation
func (m *TransformMiddleware[T]) BeforeUpdate(store *Store[T], oldValue, newValue T) T {
	return m.transformer(newValue)
}

// AfterUpdate does nothing for transform middleware
func (m *TransformMiddleware[T]) AfterUpdate(store *Store[T], oldValue, newValue T) {
	// No-op
}

// ------------------------------------
// 🔧 Built-in Action Middleware
// ------------------------------------

// ActionLoggingMiddleware logs action operations
type ActionLoggingMiddleware[T, R any] struct {
	name   string
	logger func(string, ...interface{})
}

// NewActionLoggingMiddleware creates a new action logging middleware
func NewActionLoggingMiddleware[T, R any](name string) *ActionLoggingMiddleware[T, R] {
	return &ActionLoggingMiddleware[T, R]{
		name:   name,
		logger: log.Printf,
	}
}

// BeforeAction logs before action execution
func (m *ActionLoggingMiddleware[T, R]) BeforeAction(action *StoreAction[T, R], payload T) T {
	m.logger("[%s] Action executing with payload: %v", m.name, payload)
	return payload
}

// AfterAction logs after action execution
func (m *ActionLoggingMiddleware[T, R]) AfterAction(action *StoreAction[T, R], payload T, result R) R {
	m.logger("[%s] Action completed with result: %v", m.name, result)
	return result
}

// OnError logs action errors
func (m *ActionLoggingMiddleware[T, R]) OnError(action *StoreAction[T, R], payload T, err error) error {
	m.logger("[%s] Action error: %v", m.name, err)
	return err
}

// ------------------------------------
// 🛡️ Action Validation Middleware
// ------------------------------------

// ActionValidationMiddleware validates action payloads
type ActionValidationMiddleware[T, R any] struct {
	payloadValidator func(T) error
	resultValidator  func(R) error
	name             string
}

// NewActionValidationMiddleware creates a new action validation middleware
func NewActionValidationMiddleware[T, R any](name string, payloadValidator func(T) error, resultValidator func(R) error) *ActionValidationMiddleware[T, R] {
	return &ActionValidationMiddleware[T, R]{
		payloadValidator: payloadValidator,
		resultValidator:  resultValidator,
		name:             name,
	}
}

// BeforeAction validates the payload
func (m *ActionValidationMiddleware[T, R]) BeforeAction(action *StoreAction[T, R], payload T) T {
	if m.payloadValidator != nil {
		if err := m.payloadValidator(payload); err != nil {
			panic(fmt.Sprintf("payload validation failed for action %s: %v", m.name, err))
		}
	}
	return payload
}

// AfterAction validates the result
func (m *ActionValidationMiddleware[T, R]) AfterAction(action *StoreAction[T, R], payload T, result R) R {
	if m.resultValidator != nil {
		if err := m.resultValidator(result); err != nil {
			panic(fmt.Sprintf("result validation failed for action %s: %v", m.name, err))
		}
	}
	return result
}

// OnError does nothing for validation middleware
func (m *ActionValidationMiddleware[T, R]) OnError(action *StoreAction[T, R], payload T, err error) error {
	return err
}

// ------------------------------------
// 📊 Action Performance Middleware
// ------------------------------------

// ActionPerformanceMiddleware tracks action performance
type ActionPerformanceMiddleware[T, R any] struct {
	name      string
	startTime time.Time
	stats     *PerformanceStats
	mutex     sync.RWMutex
}

// NewActionPerformanceMiddleware creates a new action performance middleware
func NewActionPerformanceMiddleware[T, R any](name string) *ActionPerformanceMiddleware[T, R] {
	return &ActionPerformanceMiddleware[T, R]{
		name: name,
		stats: &PerformanceStats{
			MinDuration: time.Hour,
		},
	}
}

// BeforeAction records start time
func (m *ActionPerformanceMiddleware[T, R]) BeforeAction(action *StoreAction[T, R], payload T) T {
	m.startTime = time.Now()
	return payload
}

// AfterAction records performance metrics
func (m *ActionPerformanceMiddleware[T, R]) AfterAction(action *StoreAction[T, R], payload T, result R) R {
	duration := time.Since(m.startTime)

	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.stats.TotalOperations++
	m.stats.TotalDuration += duration
	m.stats.AverageDuration = m.stats.TotalDuration / time.Duration(m.stats.TotalOperations)

	if duration > m.stats.MaxDuration {
		m.stats.MaxDuration = duration
	}

	if duration < m.stats.MinDuration {
		m.stats.MinDuration = duration
	}

	return result
}

// OnError does nothing for performance middleware
func (m *ActionPerformanceMiddleware[T, R]) OnError(action *StoreAction[T, R], payload T, err error) error {
	return err
}

// GetStats returns performance statistics
func (m *ActionPerformanceMiddleware[T, R]) GetStats() PerformanceStats {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return *m.stats
}

// ------------------------------------
// 🔄 Async Action Middleware
// ------------------------------------

// AsyncActionLoggingMiddleware logs async action operations
type AsyncActionLoggingMiddleware[T, R any] struct {
	name   string
	logger func(string, ...interface{})
}

// NewAsyncActionLoggingMiddleware creates a new async action logging middleware
func NewAsyncActionLoggingMiddleware[T, R any](name string) *AsyncActionLoggingMiddleware[T, R] {
	return &AsyncActionLoggingMiddleware[T, R]{
		name:   name,
		logger: log.Printf,
	}
}

// BeforeAction logs before async action execution
func (m *AsyncActionLoggingMiddleware[T, R]) BeforeAction(action *AsyncStoreAction[T, R], payload T) T {
	m.logger("[%s] Async action executing with payload: %v", m.name, payload)
	return payload
}

// AfterAction logs after async action execution
func (m *AsyncActionLoggingMiddleware[T, R]) AfterAction(action *AsyncStoreAction[T, R], payload T, result R) R {
	m.logger("[%s] Async action completed with result: %v", m.name, result)
	return result
}

// OnError logs async action errors
func (m *AsyncActionLoggingMiddleware[T, R]) OnError(action *AsyncStoreAction[T, R], payload T, err error) error {
	m.logger("[%s] Async action error: %v", m.name, err)
	return err
}

// ------------------------------------
// 🔧 Dispatcher Middleware
// ------------------------------------

// DispatcherLoggingMiddleware logs dispatcher operations
type DispatcherLoggingMiddleware struct {
	logger func(string, ...interface{})
}

// NewDispatcherLoggingMiddleware creates a new dispatcher logging middleware
func NewDispatcherLoggingMiddleware() *DispatcherLoggingMiddleware {
	return &DispatcherLoggingMiddleware{
		logger: log.Printf,
	}
}

// BeforeDispatch logs before action dispatch
func (m *DispatcherLoggingMiddleware) BeforeDispatch(actionName string, payload interface{}) interface{} {
	m.logger("[Dispatcher] Dispatching action '%s' with payload: %v", actionName, payload)
	return payload
}

// AfterDispatch logs after action dispatch
func (m *DispatcherLoggingMiddleware) AfterDispatch(actionName string, payload interface{}, result interface{}) {
	m.logger("[Dispatcher] Action '%s' completed with result: %v", actionName, result)
}

// OnError logs dispatcher errors
func (m *DispatcherLoggingMiddleware) OnError(actionName string, payload interface{}, err error) error {
	m.logger("[Dispatcher] Action '%s' failed: %v", actionName, err)
	return err
}

// ------------------------------------
// 📊 Dispatcher Performance Middleware
// ------------------------------------

// DispatcherPerformanceMiddleware tracks dispatcher performance
type DispatcherPerformanceMiddleware struct {
	startTime time.Time
	stats     map[string]*PerformanceStats
	mutex     sync.RWMutex
}

// NewDispatcherPerformanceMiddleware creates a new dispatcher performance middleware
func NewDispatcherPerformanceMiddleware() *DispatcherPerformanceMiddleware {
	return &DispatcherPerformanceMiddleware{
		stats: make(map[string]*PerformanceStats),
	}
}

// BeforeDispatch records start time
func (m *DispatcherPerformanceMiddleware) BeforeDispatch(actionName string, payload interface{}) interface{} {
	m.startTime = time.Now()

	m.mutex.Lock()
	if _, exists := m.stats[actionName]; !exists {
		m.stats[actionName] = &PerformanceStats{
			MinDuration: time.Hour,
		}
	}
	m.mutex.Unlock()

	return payload
}

// AfterDispatch records performance metrics
func (m *DispatcherPerformanceMiddleware) AfterDispatch(actionName string, payload interface{}, result interface{}) {
	duration := time.Since(m.startTime)

	m.mutex.Lock()
	defer m.mutex.Unlock()

	stats := m.stats[actionName]
	stats.TotalOperations++
	stats.TotalDuration += duration
	stats.AverageDuration = stats.TotalDuration / time.Duration(stats.TotalOperations)

	if duration > stats.MaxDuration {
		stats.MaxDuration = duration
	}

	if duration < stats.MinDuration {
		stats.MinDuration = duration
	}
}

// OnError does nothing for performance middleware
func (m *DispatcherPerformanceMiddleware) OnError(actionName string, payload interface{}, err error) error {
	return err
}

// GetStats returns performance statistics for all actions
func (m *DispatcherPerformanceMiddleware) GetStats() map[string]PerformanceStats {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	result := make(map[string]PerformanceStats)
	for name, stats := range m.stats {
		result[name] = *stats
	}
	return result
}

// ------------------------------------
// 🧪 Testing Middleware
// ------------------------------------

// TestMiddleware for testing purposes
type TestMiddleware[T any] struct {
	beforeCalled bool
	afterCalled  bool
	errorCalled  bool
	mutex        sync.RWMutex
}

// NewTestMiddleware creates a new test middleware
func NewTestMiddleware[T any]() *TestMiddleware[T] {
	return &TestMiddleware[T]{}
}

// BeforeUpdate marks before as called
func (m *TestMiddleware[T]) BeforeUpdate(store *Store[T], oldValue, newValue T) T {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.beforeCalled = true
	return newValue
}

// AfterUpdate marks after as called
func (m *TestMiddleware[T]) AfterUpdate(store *Store[T], oldValue, newValue T) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.afterCalled = true
}

// GetCallStatus returns call status for testing
func (m *TestMiddleware[T]) GetCallStatus() (before, after, error bool) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return m.beforeCalled, m.afterCalled, m.errorCalled
}

// Reset resets call status
func (m *TestMiddleware[T]) Reset() {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.beforeCalled = false
	m.afterCalled = false
	m.errorCalled = false
}
