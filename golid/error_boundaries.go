// error_boundaries.go
// Enhanced error boundary components with reactive system integration

package golid

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"maragu.dev/gomponents"
)

// ------------------------------------
// 🛡️ Enhanced Error Boundary Types
// ------------------------------------

// ErrorBoundaryEnhanced provides component-level error isolation with reactive integration
type ErrorBoundaryEnhanced struct {
	id            uint64
	owner         *Owner
	parent        *ErrorBoundaryEnhanced
	children      []*ErrorBoundaryEnhanced
	fallback      func(error, *ErrorInfo) gomponents.Node
	onError       func(error, *ErrorInfo)
	onRecovery    func(*ErrorInfo)
	state         func() ErrorBoundaryState
	setState      func(ErrorBoundaryState)
	errorInfo     func() *ErrorInfo
	setErrorInfo  func(*ErrorInfo)
	retryCount    func() int
	setRetryCount func(int)
	maxRetries    int
	retryDelay    time.Duration
	isolationMode IsolationMode
	recovery      *ErrorRecoveryEnhanced
	metrics       *ErrorMetrics
	mutex         sync.RWMutex
}

// ErrorBoundaryState represents the current state of an error boundary
type ErrorBoundaryState int

const (
	BoundaryHealthy ErrorBoundaryState = iota
	BoundaryError
	BoundaryRecovering
	BoundaryFailed
)

// String returns the string representation of ErrorBoundaryState
func (s ErrorBoundaryState) String() string {
	switch s {
	case BoundaryHealthy:
		return "Healthy"
	case BoundaryError:
		return "Error"
	case BoundaryRecovering:
		return "Recovering"
	case BoundaryFailed:
		return "Failed"
	default:
		return "Unknown"
	}
}

// IsolationMode defines how errors are isolated within the boundary
type IsolationMode int

const (
	IsolateComponent IsolationMode = iota // Isolate at component level
	IsolateSubtree                        // Isolate entire subtree
	IsolateSignals                        // Isolate signal propagation
	IsolateEffects                        // Isolate effect execution
)

// ErrorRecoveryEnhanced manages recovery strategies and mechanisms
type ErrorRecoveryEnhanced struct {
	strategy       RecoveryStrategyType
	autoRetry      bool
	retryBackoff   BackoffStrategy
	circuitBreaker *CircuitBreaker
	fallbackChain  []func(error, *ErrorInfo) gomponents.Node
	mutex          sync.RWMutex
}

// RecoveryStrategyType defines recovery approach
type RecoveryStrategyType int

const (
	RetryStrategy RecoveryStrategyType = iota
	FallbackStrategy
	PropagateStrategy
	IgnoreStrategy
)

// BackoffStrategy defines retry backoff behavior
type BackoffStrategy int

const (
	LinearBackoff BackoffStrategy = iota
	ExponentialBackoff
	FixedBackoff
)

// ErrorMetrics tracks error boundary performance and statistics
type ErrorMetrics struct {
	totalErrors      uint64
	recoveredErrors  uint64
	failedRecoveries uint64
	avgRecoveryTime  time.Duration
	lastError        time.Time
	errorRate        float64
	mutex            sync.RWMutex
}

// ------------------------------------
// 🏗️ Error Boundary Creation
// ------------------------------------

var errorBoundaryEnhancedIdCounter uint64

// CreateErrorBoundaryEnhanced creates a new enhanced error boundary
func CreateErrorBoundaryEnhanced(fallback func(error, *ErrorInfo) gomponents.Node, options ...ErrorBoundaryEnhancedOptions) *ErrorBoundaryEnhanced {
	id := atomic.AddUint64(&errorBoundaryEnhancedIdCounter, 1)
	owner := getCurrentOwner()

	// Create reactive signals for boundary state using existing signal system
	stateSignal := NewSignal(BoundaryHealthy)
	errorInfoSignal := NewSignal[*ErrorInfo](nil)
	retryCountSignal := NewSignal(0)

	boundary := &ErrorBoundaryEnhanced{
		id:            id,
		owner:         owner,
		fallback:      fallback,
		state:         stateSignal.Get,
		setState:      stateSignal.Set,
		errorInfo:     errorInfoSignal.Get,
		setErrorInfo:  errorInfoSignal.Set,
		retryCount:    retryCountSignal.Get,
		setRetryCount: retryCountSignal.Set,
		maxRetries:    3,
		retryDelay:    time.Second,
		isolationMode: IsolateComponent,
		recovery: &ErrorRecoveryEnhanced{
			strategy:     RetryStrategy,
			autoRetry:    true,
			retryBackoff: ExponentialBackoff,
		},
		metrics: &ErrorMetrics{},
	}

	// Apply options
	if len(options) > 0 {
		opt := options[0]
		if opt.MaxRetries > 0 {
			boundary.maxRetries = opt.MaxRetries
		}
		if opt.RetryDelay > 0 {
			boundary.retryDelay = opt.RetryDelay
		}
		if opt.OnError != nil {
			boundary.onError = opt.OnError
		}
		if opt.OnRecovery != nil {
			boundary.onRecovery = opt.OnRecovery
		}
		if opt.IsolationMode != 0 {
			boundary.isolationMode = opt.IsolationMode
		}
		if opt.Recovery != nil {
			boundary.recovery = opt.Recovery
		}
	}

	// Initialize circuit breaker if not provided
	if boundary.recovery.circuitBreaker == nil {
		boundary.recovery.circuitBreaker = NewCircuitBreaker(CircuitBreakerOptions{
			MaxFailures:  5,
			ResetTimeout: 30 * time.Second,
			HalfOpenMax:  3,
		})
	}

	return boundary
}

// ErrorBoundaryEnhancedOptions provides configuration for error boundary creation
type ErrorBoundaryEnhancedOptions struct {
	MaxRetries    int
	RetryDelay    time.Duration
	OnError       func(error, *ErrorInfo)
	OnRecovery    func(*ErrorInfo)
	IsolationMode IsolationMode
	Recovery      *ErrorRecoveryEnhanced
}

// ------------------------------------
// 🎯 Error Boundary Operations
// ------------------------------------

// Catch executes a function within the error boundary with reactive integration
func (eb *ErrorBoundaryEnhanced) Catch(fn func()) (err error) {
	// Check circuit breaker state
	if !eb.recovery.circuitBreaker.CanExecute() {
		return fmt.Errorf("circuit breaker is open")
	}

	defer func() {
		if r := recover(); r != nil {
			eb.handlePanic(r)
		}
	}()

	// Execute within isolated context based on isolation mode
	switch eb.isolationMode {
	case IsolateComponent:
		err = eb.catchComponent(fn)
	case IsolateSubtree:
		err = eb.catchSubtree(fn)
	case IsolateSignals:
		err = eb.catchSignals(fn)
	case IsolateEffects:
		err = eb.catchEffects(fn)
	default:
		err = eb.catchComponent(fn)
	}

	// Update circuit breaker based on result
	if err != nil {
		eb.recovery.circuitBreaker.RecordFailure()
		eb.updateMetrics(err)
	} else {
		eb.recovery.circuitBreaker.RecordSuccess()
	}

	return err
}

// catchComponent isolates errors at the component level
func (eb *ErrorBoundaryEnhanced) catchComponent(fn func()) error {
	var capturedError error

	// Execute with error capture
	func() {
		defer func() {
			if r := recover(); r != nil {
				capturedError = eb.convertPanicToError(r)
			}
		}()
		fn()
	}()

	return capturedError
}

// catchSubtree isolates errors for entire component subtree
func (eb *ErrorBoundaryEnhanced) catchSubtree(fn func()) error {
	// For now, same as component isolation
	// In a full implementation, this would suspend child computations
	return eb.catchComponent(fn)
}

// catchSignals isolates signal propagation during execution
func (eb *ErrorBoundaryEnhanced) catchSignals(fn func()) error {
	// For now, same as component isolation
	// In a full implementation, this would create signal isolation context
	return eb.catchComponent(fn)
}

// catchEffects isolates effect execution during function execution
func (eb *ErrorBoundaryEnhanced) catchEffects(fn func()) error {
	// For now, same as component isolation
	// In a full implementation, this would suspend effect scheduling
	return eb.catchComponent(fn)
}

// ------------------------------------
// 🔄 Error Recovery and Retry
// ------------------------------------

// Recover attempts to recover from an error state
func (eb *ErrorBoundaryEnhanced) Recover() error {
	eb.mutex.Lock()
	defer eb.mutex.Unlock()

	currentState := eb.state()
	if currentState != BoundaryError {
		return fmt.Errorf("boundary is not in error state")
	}

	// Set recovering state
	eb.setState(BoundaryRecovering)

	// Attempt recovery based on strategy
	switch eb.recovery.strategy {
	case RetryStrategy:
		return eb.retryRecovery()
	case FallbackStrategy:
		return eb.fallbackRecovery()
	case PropagateStrategy:
		return eb.propagateError()
	default:
		return eb.retryRecovery()
	}
}

// retryRecovery attempts to retry the failed operation
func (eb *ErrorBoundaryEnhanced) retryRecovery() error {
	currentRetries := eb.retryCount()
	if currentRetries >= eb.maxRetries {
		eb.setState(BoundaryFailed)
		return fmt.Errorf("max retries exceeded")
	}

	// Calculate backoff delay
	delay := eb.calculateBackoffDelay(currentRetries)
	time.Sleep(delay)

	// Increment retry count
	eb.setRetryCount(currentRetries + 1)

	// Reset to healthy state for retry
	eb.setState(BoundaryHealthy)
	eb.setErrorInfo(nil)

	// Trigger recovery callback
	if eb.onRecovery != nil {
		eb.onRecovery(eb.errorInfo())
	}

	return nil
}

// fallbackRecovery uses fallback rendering
func (eb *ErrorBoundaryEnhanced) fallbackRecovery() error {
	// Reset to healthy state with fallback active
	eb.setState(BoundaryHealthy)

	// Trigger recovery callback
	if eb.onRecovery != nil {
		eb.onRecovery(eb.errorInfo())
	}

	return nil
}

// propagateError propagates the error up the boundary chain
func (eb *ErrorBoundaryEnhanced) propagateError() error {
	if eb.parent != nil {
		return eb.parent.handleChildError(eb.errorInfo())
	}
	return fmt.Errorf("no parent boundary to propagate to")
}

// calculateBackoffDelay calculates the delay for retry based on backoff strategy
func (eb *ErrorBoundaryEnhanced) calculateBackoffDelay(retryCount int) time.Duration {
	switch eb.recovery.retryBackoff {
	case LinearBackoff:
		return eb.retryDelay * time.Duration(retryCount+1)
	case ExponentialBackoff:
		return eb.retryDelay * time.Duration(1<<uint(retryCount))
	case FixedBackoff:
		return eb.retryDelay
	default:
		return eb.retryDelay
	}
}

// ------------------------------------
// 🚨 Error Handling and Reporting
// ------------------------------------

// handlePanic converts a panic to an error and processes it
func (eb *ErrorBoundaryEnhanced) handlePanic(r interface{}) {
	err := eb.convertPanicToError(r)
	eb.handleError(err)
}

// convertPanicToError converts a panic value to an error
func (eb *ErrorBoundaryEnhanced) convertPanicToError(r interface{}) error {
	switch v := r.(type) {
	case error:
		return v
	case string:
		return fmt.Errorf("panic: %s", v)
	default:
		return fmt.Errorf("panic: %v", v)
	}
}

// handleError processes an error within the boundary
func (eb *ErrorBoundaryEnhanced) handleError(err error) {
	// Create error info
	info := &ErrorInfo{
		Error:     err,
		Component: fmt.Sprintf("boundary-%d", eb.id),
		Stack:     getStackTrace(),
		Props:     make(map[string]interface{}),
		Timestamp: time.Now(),
		Context:   eb.gatherErrorContext(),
	}

	// Update boundary state
	eb.setState(BoundaryError)
	eb.setErrorInfo(info)

	// Call error handler
	if eb.onError != nil {
		eb.onError(err, info)
	}

	// Update metrics
	eb.updateMetrics(err)

	// Attempt auto-recovery if enabled
	if eb.recovery.autoRetry {
		go func() {
			time.Sleep(100 * time.Millisecond) // Brief delay before retry
			eb.Recover()
		}()
	}
}

// handleChildError handles errors propagated from child boundaries
func (eb *ErrorBoundaryEnhanced) handleChildError(childError *ErrorInfo) error {
	// Create new error info for this boundary
	info := &ErrorInfo{
		Error:     childError.Error,
		Component: fmt.Sprintf("boundary-%d", eb.id),
		Stack:     getStackTrace(),
		Props:     make(map[string]interface{}),
		Timestamp: time.Now(),
		Context:   eb.gatherErrorContext(),
	}

	// Add child error context
	info.Context["childError"] = childError

	eb.handleError(childError.Error)
	return nil
}

// gatherErrorContext collects contextual information for error reporting
func (eb *ErrorBoundaryEnhanced) gatherErrorContext() map[string]interface{} {
	context := make(map[string]interface{})

	context["boundaryId"] = eb.id
	context["state"] = eb.state().String()
	context["retryCount"] = eb.retryCount()
	context["isolationMode"] = eb.isolationMode
	context["circuitBreakerState"] = eb.recovery.circuitBreaker.GetState()

	// Add owner context if available
	if eb.owner != nil {
		context["ownerId"] = eb.owner.id
		context["ownerDisposed"] = eb.owner.disposed
	}

	return context
}

// ------------------------------------
// 📊 Metrics and Monitoring
// ------------------------------------

// updateMetrics updates error boundary metrics
func (eb *ErrorBoundaryEnhanced) updateMetrics(err error) {
	eb.metrics.mutex.Lock()
	defer eb.metrics.mutex.Unlock()

	eb.metrics.totalErrors++
	eb.metrics.lastError = time.Now()

	// Calculate error rate (errors per minute)
	if eb.metrics.totalErrors > 1 {
		duration := time.Since(eb.metrics.lastError).Minutes()
		if duration > 0 {
			eb.metrics.errorRate = float64(eb.metrics.totalErrors) / duration
		}
	}
}

// GetMetrics returns current error boundary metrics
func (eb *ErrorBoundaryEnhanced) GetMetrics() ErrorMetrics {
	eb.metrics.mutex.RLock()
	defer eb.metrics.mutex.RUnlock()
	return *eb.metrics
}

// GetState returns the current boundary state
func (eb *ErrorBoundaryEnhanced) GetState() ErrorBoundaryState {
	return eb.state()
}

// GetRetryCount returns the current retry count
func (eb *ErrorBoundaryEnhanced) GetRetryCount() int {
	return eb.retryCount()
}

// ------------------------------------
// 🧹 Cleanup and Disposal
// ------------------------------------

// Dispose cleans up the error boundary and its resources
func (eb *ErrorBoundaryEnhanced) Dispose() {
	eb.mutex.Lock()
	defer eb.mutex.Unlock()

	// Dispose child boundaries
	for _, child := range eb.children {
		child.Dispose()
	}

	// Clear references
	eb.children = nil
	eb.parent = nil
	eb.fallback = nil
	eb.onError = nil
	eb.onRecovery = nil

	// Dispose circuit breaker
	if eb.recovery.circuitBreaker != nil {
		eb.recovery.circuitBreaker.Dispose()
	}
}

// ------------------------------------
// 🎨 Reactive Error Boundary Component
// ------------------------------------

// ErrorBoundaryComponent creates a reactive component with error boundary
func ErrorBoundaryComponent(
	fallback func(error, *ErrorInfo) gomponents.Node,
	children func() gomponents.Node,
	options ...ErrorBoundaryEnhancedOptions,
) gomponents.Node {
	boundary := CreateErrorBoundaryEnhanced(fallback, options...)

	// Create reactive component that responds to boundary state
	return CreateReactiveComponent(func() gomponents.Node {
		state := boundary.state()
		errorInfo := boundary.errorInfo()

		switch state {
		case BoundaryError, BoundaryFailed:
			if boundary.fallback != nil && errorInfo != nil {
				return boundary.fallback(errorInfo.Error, errorInfo)
			}
			return gomponents.Text("An error occurred")

		case BoundaryRecovering:
			return gomponents.Text("Recovering...")

		default:
			// Render children within error boundary
			var result gomponents.Node
			err := boundary.Catch(func() {
				result = children()
			})

			if err != nil {
				// Error occurred, fallback will be rendered on next update
				return gomponents.Text("Loading...")
			}

			return result
		}
	})
}

// CreateReactiveComponent creates a component that re-renders when dependencies change
func CreateReactiveComponent(render func() gomponents.Node) gomponents.Node {
	// This would integrate with the existing component system
	// For now, return the rendered result
	return render()
}

// ------------------------------------
// 🔧 Utility Functions
// ------------------------------------

// WithErrorBoundary wraps a computation with error boundary protection
func WithErrorBoundary[T any](computation func() T, fallback T, options ...ErrorBoundaryEnhancedOptions) T {
	boundary := CreateErrorBoundaryEnhanced(func(err error, info *ErrorInfo) gomponents.Node {
		return gomponents.Text(fmt.Sprintf("Error: %v", err))
	}, options...)

	var result T
	err := boundary.Catch(func() {
		result = computation()
	})

	if err != nil {
		return fallback
	}

	return result
}

// SafeSignalAccess provides safe access to signals with error boundaries
func SafeSignalAccess[T any](signal *Signal[T], fallback T) T {
	return WithErrorBoundary(func() T {
		return signal.Get()
	}, fallback)
}

// SafeEffectExecution executes an effect with error boundary protection
func SafeEffectExecution(fn func(), options ...ErrorBoundaryEnhancedOptions) {
	boundary := CreateErrorBoundaryEnhanced(func(err error, info *ErrorInfo) gomponents.Node {
		fmt.Printf("Effect error: %v\n", err)
		return gomponents.Text("Effect failed")
	}, options...)

	boundary.Catch(fn)
}
