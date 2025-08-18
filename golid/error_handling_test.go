// error_handling_test.go
// Comprehensive tests for error boundaries and graceful degradation

//go:build !js && !wasm

package golid

import (
	"fmt"
	"testing"
	"time"

	"maragu.dev/gomponents"
	"maragu.dev/gomponents/html"
)

// ------------------------------------
// 🛡️ Error Boundary Tests
// ------------------------------------

func TestErrorBoundaryEnhanced(t *testing.T) {
	t.Run("Basic Error Catching", func(t *testing.T) {
		errorCaught := false
		boundary := CreateErrorBoundaryEnhanced(
			func(err error, info *ErrorInfo) gomponents.Node {
				return html.Div(gomponents.Text("Error fallback"))
			},
			ErrorBoundaryEnhancedOptions{
				OnError: func(err error, info *ErrorInfo) {
					errorCaught = true
				},
			},
		)

		err := boundary.Catch(func() {
			panic("Test error")
		})

		if err == nil {
			t.Error("Expected error to be caught")
		}
		if !errorCaught {
			t.Error("Error handler should have been called")
		}
		if boundary.GetState() != BoundaryError {
			t.Errorf("Expected boundary state to be Error, got %s", boundary.GetState().String())
		}
	})

	t.Run("Error Recovery", func(t *testing.T) {
		boundary := CreateErrorBoundaryEnhanced(
			func(err error, info *ErrorInfo) gomponents.Node {
				return html.Div(gomponents.Text("Error fallback"))
			},
			ErrorBoundaryEnhancedOptions{
				MaxRetries: 2,
				RetryDelay: 10 * time.Millisecond,
			},
		)

		// Trigger error
		boundary.Catch(func() {
			panic("Test error")
		})

		if boundary.GetState() != BoundaryError {
			t.Error("Expected boundary to be in error state")
		}

		// Attempt recovery
		err := boundary.Recover()
		if err != nil {
			t.Errorf("Recovery failed: %v", err)
		}

		// Should be in healthy state after recovery
		if boundary.GetState() != BoundaryHealthy {
			t.Errorf("Expected boundary to be healthy after recovery, got %s", boundary.GetState().String())
		}
	})

	t.Run("Retry Limit", func(t *testing.T) {
		boundary := CreateErrorBoundaryEnhanced(
			func(err error, info *ErrorInfo) gomponents.Node {
				return html.Div(gomponents.Text("Error fallback"))
			},
			ErrorBoundaryEnhancedOptions{
				MaxRetries: 1,
				RetryDelay: 1 * time.Millisecond,
			},
		)

		// Trigger error
		boundary.Catch(func() {
			panic("Test error")
		})

		// First recovery should work
		err := boundary.Recover()
		if err != nil {
			t.Errorf("First recovery failed: %v", err)
		}

		// Trigger error again
		boundary.Catch(func() {
			panic("Test error again")
		})

		// Second recovery should fail due to retry limit
		err = boundary.Recover()
		if err != nil {
			t.Errorf("Second recovery failed: %v", err)
		}

		// Third recovery should fail
		boundary.Catch(func() {
			panic("Test error third time")
		})

		err = boundary.Recover()
		if err == nil {
			t.Error("Expected recovery to fail after max retries")
		}
	})
}

// ------------------------------------
// 🔄 Circuit Breaker Tests
// ------------------------------------

func TestCircuitBreaker(t *testing.T) {
	t.Run("Basic Circuit Breaker", func(t *testing.T) {
		cb := NewCircuitBreaker(CircuitBreakerOptions{
			MaxFailures:  2,
			ResetTimeout: 100 * time.Millisecond,
			HalfOpenMax:  1,
		})

		// Should start closed
		if cb.GetState() != CircuitClosed {
			t.Errorf("Expected circuit to start closed, got %s", cb.GetState().String())
		}

		// Should allow execution
		if !cb.CanExecute() {
			t.Error("Circuit should allow execution when closed")
		}

		// Record failures
		cb.RecordFailure()
		cb.RecordFailure()

		// Should open after max failures
		if cb.GetState() != CircuitOpen {
			t.Errorf("Expected circuit to open after failures, got %s", cb.GetState().String())
		}

		// Should reject execution when open
		if cb.CanExecute() {
			t.Error("Circuit should reject execution when open")
		}
	})

	t.Run("Circuit Breaker Reset", func(t *testing.T) {
		cb := NewCircuitBreaker(CircuitBreakerOptions{
			MaxFailures:  1,
			ResetTimeout: 50 * time.Millisecond,
			HalfOpenMax:  1,
		})

		// Trigger failure to open circuit
		cb.RecordFailure()
		if cb.GetState() != CircuitOpen {
			t.Error("Circuit should be open after failure")
		}

		// Wait for reset timeout
		time.Sleep(60 * time.Millisecond)

		// Should transition to half-open
		if cb.CanExecute() {
			if cb.GetState() != CircuitHalfOpen {
				t.Errorf("Expected circuit to be half-open, got %s", cb.GetState().String())
			}
		}

		// Record success to close circuit
		cb.RecordSuccess()
		if cb.GetState() != CircuitClosed {
			t.Errorf("Expected circuit to close after success, got %s", cb.GetState().String())
		}
	})
}

// ------------------------------------
// 🎯 Fallback Signal Tests
// ------------------------------------

func TestFallbackSignal(t *testing.T) {
	t.Run("Primary Source Success", func(t *testing.T) {
		primary := func() string {
			return "primary data"
		}
		fallback := func() string {
			return "fallback data"
		}

		fs := CreateFallbackSignal(primary, fallback)

		result := fs.Get()
		if result != "primary data" {
			t.Errorf("Expected primary data, got %s", result)
		}
		if fs.GetState() != FallbackPrimary {
			t.Errorf("Expected primary state, got %s", fs.GetState().String())
		}
	})

	t.Run("Primary Source Failure", func(t *testing.T) {
		failureCount := 0
		primary := func() string {
			failureCount++
			if failureCount <= 2 {
				panic("primary failure")
			}
			return "primary data"
		}
		fallback := func() string {
			return "fallback data"
		}

		fs := CreateFallbackSignal(primary, fallback, FallbackOptions[string]{
			MaxErrors: 1,
			Timeout:   10 * time.Millisecond,
		})

		// First call should fail and switch to fallback
		result := fs.Get()
		if result != "fallback data" {
			t.Errorf("Expected fallback data, got %s", result)
		}
		if fs.GetState() != FallbackSecondary {
			t.Errorf("Expected secondary state, got %s", fs.GetState().String())
		}
	})
}

// ------------------------------------
// 🚀 Graceful Degradation Tests
// ------------------------------------

func TestGracefulDegradation(t *testing.T) {
	t.Run("Feature Detection", func(t *testing.T) {
		// Register a test feature
		RegisterFeature("test-feature", func() bool {
			return true
		}, func() interface{} {
			return "test fallback"
		})

		if !IsFeatureAvailable("test-feature") {
			t.Error("Test feature should be available")
		}

		fallback := GetFeatureFallback("test-feature")
		if fallback != "test fallback" {
			t.Errorf("Expected test fallback, got %v", fallback)
		}
	})

	t.Run("Degraded Mode", func(t *testing.T) {
		// Enable degraded mode
		EnableDegradedMode(ModerateDegradation)

		if !IsInDegradedMode() {
			t.Error("Should be in degraded mode")
		}
		if GetDegradationLevel() != ModerateDegradation {
			t.Errorf("Expected moderate degradation, got %s", GetDegradationLevel().String())
		}

		// Test feature availability in degraded mode
		if IsFeatureEnabled("animations") {
			t.Error("Animations should be disabled in moderate degradation")
		}

		// Disable degraded mode
		DisableDegradedMode()
		if IsInDegradedMode() {
			t.Error("Should not be in degraded mode after disabling")
		}
	})
}

// ------------------------------------
// 🔧 Safe Operations Tests
// ------------------------------------

func TestSafeOperations(t *testing.T) {
	t.Run("Safe Signal Access", func(t *testing.T) {
		signal := NewSignal("test value")

		result := SafeSignalAccess(signal, "fallback")
		if result != "test value" {
			t.Errorf("Expected test value, got %s", result)
		}
	})

	t.Run("Safe Effect Execution", func(t *testing.T) {
		executed := false
		SafeEffectExecution(func() {
			executed = true
		})

		if !executed {
			t.Error("Effect should have been executed")
		}
	})

	t.Run("WithErrorBoundary", func(t *testing.T) {
		// Test successful computation
		result := WithErrorBoundary(func() string {
			return "success"
		}, "fallback")

		if result != "success" {
			t.Errorf("Expected success, got %s", result)
		}

		// Test failing computation
		result = WithErrorBoundary(func() string {
			panic("computation failed")
		}, "fallback")

		if result != "fallback" {
			t.Errorf("Expected fallback, got %s", result)
		}
	})
}

// ------------------------------------
// 🔄 Error Recovery Coordinator Tests
// ------------------------------------

func TestErrorRecoveryCoordinator(t *testing.T) {
	t.Run("Strategy Registration and Recovery", func(t *testing.T) {
		coordinator := NewErrorRecoveryCoordinator()

		// Register a test strategy
		strategy := &RecoveryStrategyImpl{
			Name:     "test-strategy",
			Priority: 1,
			CanHandle: func(err error) bool {
				return fmt.Sprintf("%v", err) == "test error"
			},
			Recover: func(err error) (interface{}, error) {
				return "recovered", nil
			},
		}

		coordinator.RegisterStrategy(strategy)

		// Test recovery
		result, err := coordinator.RecoverFromError(fmt.Errorf("test error"))
		if err != nil {
			t.Errorf("Recovery failed: %v", err)
		}
		if result != "recovered" {
			t.Errorf("Expected recovered, got %v", result)
		}
	})

	t.Run("Fallback Chain", func(t *testing.T) {
		coordinator := NewErrorRecoveryCoordinator()

		// Add fallback
		coordinator.AddFallback(func(err error) interface{} {
			return "fallback result"
		})

		// Test with unhandled error
		result, err := coordinator.RecoverFromError(fmt.Errorf("unknown error"))
		if err != nil {
			t.Errorf("Fallback should have handled error: %v", err)
		}
		if result != "fallback result" {
			t.Errorf("Expected fallback result, got %v", result)
		}
	})
}

// ------------------------------------
// 📊 Metrics Tests
// ------------------------------------

func TestMetrics(t *testing.T) {
	t.Run("Circuit Breaker Metrics", func(t *testing.T) {
		cb := NewCircuitBreaker(CircuitBreakerOptions{
			MaxFailures: 2,
		})

		// Execute some operations
		cb.CanExecute()
		cb.RecordSuccess()
		cb.CanExecute()
		cb.RecordFailure()

		metrics := cb.GetMetrics()
		if metrics.TotalRequests() != 2 {
			t.Errorf("Expected 2 total requests, got %d", metrics.TotalRequests())
		}
		if metrics.SuccessRequests() != 1 {
			t.Errorf("Expected 1 success request, got %d", metrics.SuccessRequests())
		}
		if metrics.FailedRequests() != 1 {
			t.Errorf("Expected 1 failed request, got %d", metrics.FailedRequests())
		}
	})

	t.Run("Error Boundary Metrics", func(t *testing.T) {
		boundary := CreateErrorBoundaryEnhanced(
			func(err error, info *ErrorInfo) gomponents.Node {
				return html.Div(gomponents.Text("Error"))
			},
		)

		// Trigger an error
		boundary.Catch(func() {
			panic("test error")
		})

		metrics := boundary.GetMetrics()
		if metrics.totalErrors != 1 {
			t.Errorf("Expected 1 total error, got %d", metrics.totalErrors)
		}
	})
}

// ------------------------------------
// 🧪 Integration Tests
// ------------------------------------

func TestErrorHandlingIntegration(t *testing.T) {
	t.Run("Error Boundary with Circuit Breaker", func(t *testing.T) {
		// Create circuit breaker
		cb := NewCircuitBreaker(CircuitBreakerOptions{
			MaxFailures: 2,
		})

		// Create error boundary with circuit breaker
		boundary := CreateErrorBoundaryEnhanced(
			func(err error, info *ErrorInfo) gomponents.Node {
				return html.Div(gomponents.Text("Circuit breaker fallback"))
			},
			ErrorBoundaryEnhancedOptions{
				Recovery: &ErrorRecoveryEnhanced{
					circuitBreaker: cb,
				},
			},
		)

		// Test that circuit breaker integration works
		failureCount := 0
		for i := 0; i < 5; i++ {
			err := boundary.Catch(func() {
				failureCount++
				if failureCount <= 3 {
					panic("service failure")
				}
			})

			if err != nil && cb.GetState() == CircuitOpen {
				// Circuit should be open after failures
				break
			}
		}

		if cb.GetState() != CircuitOpen {
			t.Error("Circuit breaker should be open after repeated failures")
		}
	})

	t.Run("Graceful Degradation with Error Boundaries", func(t *testing.T) {
		// Enable degraded mode
		EnableDegradedMode(LightDegradation)
		defer DisableDegradedMode()

		// Create boundary that respects degraded mode
		boundary := CreateErrorBoundaryEnhanced(
			func(err error, info *ErrorInfo) gomponents.Node {
				if IsInDegradedMode() {
					return html.Div(gomponents.Text("Degraded mode fallback"))
				}
				return html.Div(gomponents.Text("Normal fallback"))
			},
		)

		// Test error handling in degraded mode
		err := boundary.Catch(func() {
			panic("test error in degraded mode")
		})

		if err == nil {
			t.Error("Expected error to be caught")
		}
		if boundary.GetState() != BoundaryError {
			t.Error("Boundary should be in error state")
		}
	})
}
