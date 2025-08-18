// error_handling.go
// Error handling and recovery mechanisms for the reactivity system

package golid

import (
	"fmt"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

// ------------------------------------
// 🚨 Error Handling Types
// ------------------------------------

// ErrorBoundary provides error isolation and recovery for reactive computations
type ErrorBoundary struct {
	owner     *Owner
	fallback  func(error) interface{}
	onError   func(error, *ErrorInfo)
	recovered bool
	mutex     sync.RWMutex
}

// ErrorInfo contains detailed information about an error
type ErrorInfo struct {
	Error     error
	Component string
	Stack     []uintptr
	Props     map[string]interface{}
	Timestamp time.Time
	Context   map[string]interface{}
}

// RecoveryStrategy defines how to handle different types of errors
type RecoveryStrategy int

const (
	Retry RecoveryStrategy = iota
	Fallback
	Propagate
	Ignore
)

// ErrorHandler manages error recovery strategies
type ErrorHandler struct {
	strategies map[string]RecoveryStrategy
	retries    map[string]int
	maxRetries int
	mutex      sync.RWMutex
}

// ------------------------------------
// 🛡️ Error Boundary Implementation
// ------------------------------------

// CreateErrorBoundary creates a new error boundary
func CreateErrorBoundary(fallback func(error) interface{}) *ErrorBoundary {
	return &ErrorBoundary{
		owner:     getCurrentOwner(),
		fallback:  fallback,
		onError:   nil,
		recovered: false,
	}
}

// Catch executes a function within the error boundary
func (e *ErrorBoundary) Catch(fn func()) (err error) {
	defer func() {
		if r := recover(); r != nil {
			e.mutex.Lock()
			e.recovered = true
			e.mutex.Unlock()

			// Convert panic to error
			switch v := r.(type) {
			case error:
				err = v
			case string:
				err = fmt.Errorf("panic: %s", v)
			default:
				err = fmt.Errorf("panic: %v", v)
			}

			// Create error info
			info := &ErrorInfo{
				Error:     err,
				Component: "unknown",
				Stack:     getStackTrace(),
				Props:     make(map[string]interface{}),
				Timestamp: time.Now(),
				Context:   make(map[string]interface{}),
			}

			// Call error handler if provided
			if e.onError != nil {
				e.onError(err, info)
			}

			// Call fallback if provided
			if e.fallback != nil {
				e.fallback(err)
			}
		}
	}()

	fn()
	return nil
}

// Reset resets the error boundary state
func (e *ErrorBoundary) Reset() {
	e.mutex.Lock()
	defer e.mutex.Unlock()
	e.recovered = false
}

// OnError sets the error handler callback
func (e *ErrorBoundary) OnError(handler func(error, *ErrorInfo)) {
	e.mutex.Lock()
	defer e.mutex.Unlock()
	e.onError = handler
}

// IsRecovered returns whether the boundary has caught an error
func (e *ErrorBoundary) IsRecovered() bool {
	e.mutex.RLock()
	defer e.mutex.RUnlock()
	return e.recovered
}

// ------------------------------------
// 🔄 Error Recovery Strategies
// ------------------------------------

// NewErrorHandler creates a new error handler
func NewErrorHandler() *ErrorHandler {
	return &ErrorHandler{
		strategies: make(map[string]RecoveryStrategy),
		retries:    make(map[string]int),
		maxRetries: 3,
	}
}

// SetStrategy sets the recovery strategy for a specific error pattern
func (h *ErrorHandler) SetStrategy(pattern string, strategy RecoveryStrategy) {
	h.mutex.Lock()
	defer h.mutex.Unlock()
	h.strategies[pattern] = strategy
}

// SetMaxRetries sets the maximum number of retries
func (h *ErrorHandler) SetMaxRetries(max int) {
	h.mutex.Lock()
	defer h.mutex.Unlock()
	h.maxRetries = max
}

// Handle processes an error according to the configured strategy
func (h *ErrorHandler) Handle(err error, context string) bool {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	strategy, exists := h.strategies[context]
	if !exists {
		strategy = Propagate // Default strategy
	}

	switch strategy {
	case Retry:
		retryCount := h.retries[context]
		if retryCount < h.maxRetries {
			h.retries[context] = retryCount + 1
			return true // Indicate retry
		}
		// Max retries reached, fall through to propagate
		fallthrough

	case Propagate:
		return false // Propagate the error

	case Ignore:
		return true // Ignore the error

	case Fallback:
		// Fallback handling would be implemented by the caller
		return true
	}

	return false
}

// ResetRetries resets the retry count for a specific context
func (h *ErrorHandler) ResetRetries(context string) {
	h.mutex.Lock()
	defer h.mutex.Unlock()
	delete(h.retries, context)
}

// ------------------------------------
// 🔧 Safe Computation Execution
// ------------------------------------

// SafeCreateEffect creates an effect with error handling
func SafeCreateEffect(fn func(), owner *Owner, errorHandler *ErrorHandler) *Computation {
	boundary := CreateErrorBoundary(func(err error) interface{} {
		if errorHandler != nil {
			errorHandler.Handle(err, "effect")
		}
		return nil
	})

	return CreateEffect(func() {
		boundary.Catch(fn)
	}, owner)
}

// SafeCreateMemo creates a memo with error handling
func SafeCreateMemo[T any](fn func() T, defaultValue T, owner *Owner, errorHandler *ErrorHandler) func() T {
	boundary := CreateErrorBoundary(func(err error) interface{} {
		if errorHandler != nil {
			errorHandler.Handle(err, "memo")
		}
		return defaultValue
	})

	return CreateMemo(func() T {
		var result T
		err := boundary.Catch(func() {
			result = fn()
		})
		if err != nil {
			return defaultValue
		}
		return result
	}, owner)
}

// SafeCreateSignal creates a signal with validation
func SafeCreateSignal[T any](initial T, validator func(T) error, options ...SignalOptions[T]) (func() T, func(T) error) {
	getter, setter := CreateSignal(initial, options...)

	safeSetter := func(value T) error {
		if validator != nil {
			if err := validator(value); err != nil {
				return fmt.Errorf("signal validation failed: %w", err)
			}
		}
		setter(value)
		return nil
	}

	return getter, safeSetter
}

// ------------------------------------
// 🧹 Cleanup and Resource Management
// ------------------------------------

// SafeCleanup ensures cleanup functions don't panic
func SafeCleanup(cleanupFn func()) {
	defer func() {
		if r := recover(); r != nil {
			// Log the panic but don't propagate it during cleanup
			fmt.Printf("Warning: Panic during cleanup: %v\n", r)
		}
	}()
	cleanupFn()
}

// WithTimeout executes a function with a timeout
func WithTimeout[T any](fn func() T, timeout time.Duration, defaultValue T) T {
	result := make(chan T, 1)
	done := make(chan bool, 1)

	go func() {
		defer func() {
			if r := recover(); r != nil {
				// Handle panic in goroutine
				done <- true
			}
		}()
		result <- fn()
		done <- true
	}()

	select {
	case res := <-result:
		return res
	case <-time.After(timeout):
		return defaultValue
	case <-done:
		return defaultValue
	}
}

// ------------------------------------
// 🔍 Debugging and Diagnostics
// ------------------------------------

// getStackTrace captures the current stack trace
func getStackTrace() []uintptr {
	const depth = 32
	var pcs [depth]uintptr
	n := runtime.Callers(3, pcs[:])
	return pcs[0:n]
}

// FormatStackTrace formats a stack trace for display
func FormatStackTrace(stack []uintptr) []string {
	frames := runtime.CallersFrames(stack)
	var result []string

	for {
		frame, more := frames.Next()
		result = append(result, fmt.Sprintf("%s:%d %s", frame.File, frame.Line, frame.Function))
		if !more {
			break
		}
	}

	return result
}

// ReactivityDiagnostics provides diagnostic information about the reactive system
type ReactivityDiagnostics struct {
	SignalCount      int
	ComputationCount int
	OwnerCount       int
	SchedulerStats   SchedulerStats
	MemoryUsage      runtime.MemStats
}

// GetDiagnostics returns current system diagnostics
func GetDiagnostics() *ReactivityDiagnostics {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	return &ReactivityDiagnostics{
		SignalCount:      int(atomic.LoadUint64(&signalIdCounter)),
		ComputationCount: int(atomic.LoadUint64(&computationIdCounter)),
		OwnerCount:       int(atomic.LoadUint64(&ownerIdCounter)),
		SchedulerStats:   getScheduler().GetStats(),
		MemoryUsage:      memStats,
	}
}

// ------------------------------------
// 🧪 Testing Utilities
// ------------------------------------

// MockErrorBoundary creates a mock error boundary for testing
func MockErrorBoundary() *ErrorBoundary {
	return &ErrorBoundary{
		fallback: func(err error) interface{} {
			return fmt.Sprintf("Mock fallback: %v", err)
		},
		recovered: false,
	}
}

// SimulateError simulates an error for testing purposes
func SimulateError(message string) {
	panic(fmt.Errorf("simulated error: %s", message))
}

// ------------------------------------
// 🔄 Global Error Handler
// ------------------------------------

var globalErrorHandler = NewErrorHandler()

// SetGlobalErrorStrategy sets a global error strategy
func SetGlobalErrorStrategy(pattern string, strategy RecoveryStrategy) {
	globalErrorHandler.SetStrategy(pattern, strategy)
}

// HandleGlobalError handles an error using the global error handler
func HandleGlobalError(err error, context string) bool {
	return globalErrorHandler.Handle(err, context)
}
