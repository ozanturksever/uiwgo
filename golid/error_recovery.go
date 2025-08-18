// error_recovery.go
// Error recovery mechanisms and circuit breaker patterns

package golid

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

// ------------------------------------
// 🔄 Circuit Breaker Implementation
// ------------------------------------

// CircuitBreakerState represents the state of a circuit breaker
type CircuitBreakerState int

const (
	CircuitClosed CircuitBreakerState = iota
	CircuitOpen
	CircuitHalfOpen
)

// String returns the string representation of CircuitBreakerState
func (s CircuitBreakerState) String() string {
	switch s {
	case CircuitClosed:
		return "Closed"
	case CircuitOpen:
		return "Open"
	case CircuitHalfOpen:
		return "HalfOpen"
	default:
		return "Unknown"
	}
}

// CircuitBreaker implements the circuit breaker pattern for error recovery
type CircuitBreaker struct {
	maxFailures   int
	resetTimeout  time.Duration
	halfOpenMax   int
	state         CircuitBreakerState
	failures      int
	halfOpenCount int
	lastFailTime  time.Time
	mutex         sync.RWMutex
	onStateChange func(CircuitBreakerState)
	metrics       *CircuitBreakerMetrics
}

// CircuitBreakerOptions provides configuration for circuit breaker creation
type CircuitBreakerOptions struct {
	MaxFailures   int
	ResetTimeout  time.Duration
	HalfOpenMax   int
	OnStateChange func(CircuitBreakerState)
}

// CircuitBreakerMetrics tracks circuit breaker performance
type CircuitBreakerMetrics struct {
	totalRequests    uint64
	successRequests  uint64
	failedRequests   uint64
	rejectedRequests uint64
	stateChanges     uint64
	lastStateChange  time.Time
}

// Getter methods for CircuitBreakerMetrics
func (m *CircuitBreakerMetrics) TotalRequests() uint64 {
	return m.totalRequests
}

func (m *CircuitBreakerMetrics) SuccessRequests() uint64 {
	return m.successRequests
}

func (m *CircuitBreakerMetrics) FailedRequests() uint64 {
	return m.failedRequests
}

func (m *CircuitBreakerMetrics) RejectedRequests() uint64 {
	return m.rejectedRequests
}

func (m *CircuitBreakerMetrics) StateChanges() uint64 {
	return m.stateChanges
}

func (m *CircuitBreakerMetrics) LastStateChange() time.Time {
	return m.lastStateChange
}

// NewCircuitBreaker creates a new circuit breaker
func NewCircuitBreaker(options CircuitBreakerOptions) *CircuitBreaker {
	cb := &CircuitBreaker{
		maxFailures:   options.MaxFailures,
		resetTimeout:  options.ResetTimeout,
		halfOpenMax:   options.HalfOpenMax,
		state:         CircuitClosed,
		onStateChange: options.OnStateChange,
		metrics:       &CircuitBreakerMetrics{},
	}

	// Set defaults if not provided
	if cb.maxFailures <= 0 {
		cb.maxFailures = 5
	}
	if cb.resetTimeout <= 0 {
		cb.resetTimeout = 30 * time.Second
	}
	if cb.halfOpenMax <= 0 {
		cb.halfOpenMax = 3
	}

	return cb
}

// CanExecute checks if the circuit breaker allows execution
func (cb *CircuitBreaker) CanExecute() bool {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()

	atomic.AddUint64(&cb.metrics.totalRequests, 1)

	switch cb.state {
	case CircuitClosed:
		return true
	case CircuitOpen:
		if time.Since(cb.lastFailTime) >= cb.resetTimeout {
			cb.setState(CircuitHalfOpen)
			cb.halfOpenCount = 0
			return true
		}
		atomic.AddUint64(&cb.metrics.rejectedRequests, 1)
		return false
	case CircuitHalfOpen:
		if cb.halfOpenCount < cb.halfOpenMax {
			cb.halfOpenCount++
			return true
		}
		atomic.AddUint64(&cb.metrics.rejectedRequests, 1)
		return false
	default:
		return false
	}
}

// RecordSuccess records a successful execution
func (cb *CircuitBreaker) RecordSuccess() {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()

	atomic.AddUint64(&cb.metrics.successRequests, 1)

	switch cb.state {
	case CircuitHalfOpen:
		cb.failures = 0
		cb.setState(CircuitClosed)
	case CircuitClosed:
		cb.failures = 0
	}
}

// RecordFailure records a failed execution
func (cb *CircuitBreaker) RecordFailure() {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()

	atomic.AddUint64(&cb.metrics.failedRequests, 1)
	cb.failures++
	cb.lastFailTime = time.Now()

	switch cb.state {
	case CircuitClosed:
		if cb.failures >= cb.maxFailures {
			cb.setState(CircuitOpen)
		}
	case CircuitHalfOpen:
		cb.setState(CircuitOpen)
	}
}

// GetState returns the current state of the circuit breaker
func (cb *CircuitBreaker) GetState() CircuitBreakerState {
	cb.mutex.RLock()
	defer cb.mutex.RUnlock()
	return cb.state
}

// setState changes the circuit breaker state and triggers callbacks
func (cb *CircuitBreaker) setState(newState CircuitBreakerState) {
	if cb.state != newState {
		cb.state = newState
		atomic.AddUint64(&cb.metrics.stateChanges, 1)
		cb.metrics.lastStateChange = time.Now()

		if cb.onStateChange != nil {
			go cb.onStateChange(newState)
		}
	}
}

// GetMetrics returns current circuit breaker metrics
func (cb *CircuitBreaker) GetMetrics() CircuitBreakerMetrics {
	return *cb.metrics
}

// Reset manually resets the circuit breaker to closed state
func (cb *CircuitBreaker) Reset() {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()

	cb.failures = 0
	cb.halfOpenCount = 0
	cb.setState(CircuitClosed)
}

// Dispose cleans up the circuit breaker resources
func (cb *CircuitBreaker) Dispose() {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()

	cb.onStateChange = nil
	cb.metrics = nil
}

// ------------------------------------
// 🔄 Retry Mechanisms
// ------------------------------------

// RetryPolicy defines retry behavior
type RetryPolicy struct {
	MaxAttempts   int
	BaseDelay     time.Duration
	MaxDelay      time.Duration
	BackoffFactor float64
	Jitter        bool
}

// DefaultRetryPolicy returns a sensible default retry policy
func DefaultRetryPolicy() RetryPolicy {
	return RetryPolicy{
		MaxAttempts:   3,
		BaseDelay:     100 * time.Millisecond,
		MaxDelay:      5 * time.Second,
		BackoffFactor: 2.0,
		Jitter:        true,
	}
}

// RetryWithPolicy executes a function with retry logic
func RetryWithPolicy[T any](fn func() (T, error), policy RetryPolicy) (T, error) {
	var result T
	var lastErr error

	for attempt := 0; attempt < policy.MaxAttempts; attempt++ {
		result, lastErr = fn()
		if lastErr == nil {
			return result, nil
		}

		// Don't delay after the last attempt
		if attempt < policy.MaxAttempts-1 {
			delay := calculateDelay(attempt, policy)
			time.Sleep(delay)
		}
	}

	return result, fmt.Errorf("retry failed after %d attempts: %w", policy.MaxAttempts, lastErr)
}

// calculateDelay calculates the delay for a retry attempt
func calculateDelay(attempt int, policy RetryPolicy) time.Duration {
	delay := float64(policy.BaseDelay) * powBackoff(policy.BackoffFactor, float64(attempt))

	if delay > float64(policy.MaxDelay) {
		delay = float64(policy.MaxDelay)
	}

	if policy.Jitter {
		// Add up to 10% jitter
		jitter := delay * 0.1 * (2.0*rand() - 1.0)
		delay += jitter
	}

	return time.Duration(delay)
}

// Simple power function for delay calculation
func powBackoff(base, exp float64) float64 {
	if exp == 0 {
		return 1
	}
	result := base
	for i := 1; i < int(exp); i++ {
		result *= base
	}
	return result
}

// Simple random function for jitter
func rand() float64 {
	// Simple linear congruential generator
	seed := time.Now().UnixNano()
	return float64((seed*1103515245+12345)&0x7fffffff) / float64(0x7fffffff)
}

// ------------------------------------
// 🛡️ Error Recovery Coordinator
// ------------------------------------

// ErrorRecoveryCoordinator manages multiple recovery strategies
type ErrorRecoveryCoordinator struct {
	strategies     map[string]*RecoveryStrategyImpl
	circuitBreaker *CircuitBreaker
	retryPolicy    RetryPolicy
	fallbackChain  []func(error) interface{}
	mutex          sync.RWMutex
}

// RecoveryStrategyImpl defines a specific recovery approach
type RecoveryStrategyImpl struct {
	Name           string
	Priority       int
	CanHandle      func(error) bool
	Recover        func(error) (interface{}, error)
	CircuitBreaker *CircuitBreaker
	RetryPolicy    *RetryPolicy
}

// NewErrorRecoveryCoordinator creates a new recovery coordinator
func NewErrorRecoveryCoordinator() *ErrorRecoveryCoordinator {
	return &ErrorRecoveryCoordinator{
		strategies: make(map[string]*RecoveryStrategyImpl),
		circuitBreaker: NewCircuitBreaker(CircuitBreakerOptions{
			MaxFailures:  5,
			ResetTimeout: 30 * time.Second,
			HalfOpenMax:  3,
		}),
		retryPolicy:   DefaultRetryPolicy(),
		fallbackChain: make([]func(error) interface{}, 0),
	}
}

// RegisterStrategy registers a recovery strategy
func (erc *ErrorRecoveryCoordinator) RegisterStrategy(strategy *RecoveryStrategyImpl) {
	erc.mutex.Lock()
	defer erc.mutex.Unlock()
	erc.strategies[strategy.Name] = strategy
}

// RecoverFromError attempts to recover from an error using registered strategies
func (erc *ErrorRecoveryCoordinator) RecoverFromError(err error) (interface{}, error) {
	erc.mutex.RLock()
	defer erc.mutex.RUnlock()

	// Try strategies in priority order
	for _, strategy := range erc.strategies {
		if strategy.CanHandle(err) {
			if strategy.CircuitBreaker != nil && !strategy.CircuitBreaker.CanExecute() {
				continue
			}

			result, recoveryErr := strategy.Recover(err)
			if recoveryErr == nil {
				if strategy.CircuitBreaker != nil {
					strategy.CircuitBreaker.RecordSuccess()
				}
				return result, nil
			}

			if strategy.CircuitBreaker != nil {
				strategy.CircuitBreaker.RecordFailure()
			}
		}
	}

	// Try fallback chain
	for _, fallback := range erc.fallbackChain {
		if result := fallback(err); result != nil {
			return result, nil
		}
	}

	return nil, fmt.Errorf("no recovery strategy could handle error: %w", err)
}

// AddFallback adds a fallback function to the chain
func (erc *ErrorRecoveryCoordinator) AddFallback(fallback func(error) interface{}) {
	erc.mutex.Lock()
	defer erc.mutex.Unlock()
	erc.fallbackChain = append(erc.fallbackChain, fallback)
}

// ------------------------------------
// 🔧 Error Recovery Utilities
// ------------------------------------

// WithErrorRecovery wraps a computation with error recovery
func WithErrorRecovery[T any](computation func() T, recovery func(error) T) func() T {
	return func() T {
		defer func() {
			if r := recover(); r != nil {
				var err error
				switch v := r.(type) {
				case error:
					err = v
				case string:
					err = fmt.Errorf("panic: %s", v)
				default:
					err = fmt.Errorf("panic: %v", v)
				}

				// Call recovery function
				recovery(err)
			}
		}()

		return computation()
	}
}

// SafeExecute executes a function with comprehensive error handling
func SafeExecute[T any](fn func() (T, error), options ...SafeExecuteOptions) (T, error) {
	var opts SafeExecuteOptions
	if len(options) > 0 {
		opts = options[0]
	}

	// Set defaults
	if opts.RetryPolicy == nil {
		defaultPolicy := DefaultRetryPolicy()
		opts.RetryPolicy = &defaultPolicy
	}
	if opts.Timeout == 0 {
		opts.Timeout = 30 * time.Second
	}

	// Execute with timeout
	resultChan := make(chan T, 1)
	errorChan := make(chan error, 1)

	go func() {
		result, err := RetryWithPolicy(fn, *opts.RetryPolicy)
		if err != nil {
			errorChan <- err
		} else {
			resultChan <- result
		}
	}()

	select {
	case result := <-resultChan:
		return result, nil
	case err := <-errorChan:
		if opts.Fallback != nil {
			if fallbackResult, ok := opts.Fallback(err).(T); ok {
				return fallbackResult, nil
			}
		}
		return *new(T), err
	case <-time.After(opts.Timeout):
		if opts.Fallback != nil {
			if fallbackResult, ok := opts.Fallback(fmt.Errorf("operation timed out")).(T); ok {
				return fallbackResult, nil
			}
		}
		return *new(T), fmt.Errorf("operation timed out after %v", opts.Timeout)
	}
}

// SafeExecuteOptions provides configuration for safe execution
type SafeExecuteOptions struct {
	RetryPolicy *RetryPolicy
	Timeout     time.Duration
	Fallback    func(error) interface{}
}
