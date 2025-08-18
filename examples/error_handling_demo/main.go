// main.go
// Comprehensive examples demonstrating error handling patterns in Golid

package main

import (
	"fmt"
	"time"

	"app/golid"
	"maragu.dev/gomponents"
	"maragu.dev/gomponents/html"
)

func main() {
	fmt.Println("🛡️ Golid Error Handling and Recovery Demo")
	fmt.Println("==========================================")

	// Demo 1: Basic Error Boundary
	fmt.Println("\n1. Basic Error Boundary Demo")
	basicErrorBoundaryDemo()

	// Demo 2: Fallback Signals
	fmt.Println("\n2. Fallback Signals Demo")
	fallbackSignalsDemo()

	// Demo 3: Circuit Breaker Pattern
	fmt.Println("\n3. Circuit Breaker Demo")
	circuitBreakerDemo()

	// Demo 4: Graceful Degradation
	fmt.Println("\n4. Graceful Degradation Demo")
	gracefulDegradationDemo()

	// Demo 5: Progressive Enhancement
	fmt.Println("\n5. Progressive Enhancement Demo")
	progressiveEnhancementDemo()

	// Demo 6: Error Recovery Strategies
	fmt.Println("\n6. Error Recovery Strategies Demo")
	errorRecoveryDemo()

	// Demo 7: Safe Signal Operations
	fmt.Println("\n7. Safe Signal Operations Demo")
	safeSignalDemo()

	fmt.Println("\n✅ All error handling demos completed!")
}

// ------------------------------------
// 🛡️ Basic Error Boundary Demo
// ------------------------------------

func basicErrorBoundaryDemo() {
	// Create an error boundary with fallback
	boundary := golid.CreateErrorBoundaryEnhanced(
		func(err error, info *golid.ErrorInfo) gomponents.Node {
			return html.Div(
				html.Class("error-fallback"),
				html.H3(gomponents.Text("Something went wrong!")),
				html.P(gomponents.Textf("Error: %v", err)),
				html.P(gomponents.Textf("Component: %s", info.Component)),
				html.P(gomponents.Textf("Time: %s", info.Timestamp.Format(time.RFC3339))),
			)
		},
		golid.ErrorBoundaryEnhancedOptions{
			MaxRetries: 3,
			RetryDelay: time.Second,
			OnError: func(err error, info *golid.ErrorInfo) {
				fmt.Printf("❌ Error caught by boundary: %v\n", err)
			},
			OnRecovery: func(info *golid.ErrorInfo) {
				fmt.Printf("✅ Boundary recovered from error\n")
			},
		},
	)

	// Simulate a function that might panic
	riskyFunction := func() {
		fmt.Println("   Executing risky function...")
		panic("Simulated error in component")
	}

	// Execute within error boundary
	err := boundary.Catch(riskyFunction)
	if err != nil {
		fmt.Printf("   Error handled by boundary: %v\n", err)
	}

	// Show boundary state
	fmt.Printf("   Boundary state: %s\n", boundary.GetState().String())
	fmt.Printf("   Retry count: %d\n", boundary.GetRetryCount())

	// Demonstrate recovery
	fmt.Println("   Attempting recovery...")
	if recoveryErr := boundary.Recover(); recoveryErr != nil {
		fmt.Printf("   Recovery failed: %v\n", recoveryErr)
	} else {
		fmt.Printf("   Recovery successful!\n")
	}
}

// ------------------------------------
// 🔄 Fallback Signals Demo
// ------------------------------------

func fallbackSignalsDemo() {
	// Create a primary data source that might fail
	primarySource := func() string {
		// Simulate intermittent failures
		if time.Now().Unix()%3 == 0 {
			panic("Primary data source unavailable")
		}
		return "Primary data: Current time is " + time.Now().Format("15:04:05")
	}

	// Create a reliable fallback source
	fallbackSource := func() string {
		return "Fallback data: Service temporarily unavailable"
	}

	// Create fallback signal
	fallbackSignal := golid.CreateFallbackSignal(
		primarySource,
		fallbackSource,
		golid.FallbackOptions[string]{
			MaxErrors: 2,
			Timeout:   time.Second,
			Validator: func(data string) bool {
				return len(data) > 0
			},
		},
	)

	// Test fallback signal multiple times
	for i := 0; i < 5; i++ {
		fmt.Printf("   Attempt %d: ", i+1)
		data := fallbackSignal.Get()
		state := fallbackSignal.GetState()
		fmt.Printf("%s (State: %s)\n", data, state.String())
		time.Sleep(100 * time.Millisecond)
	}
}

// ------------------------------------
// ⚡ Circuit Breaker Demo
// ------------------------------------

func circuitBreakerDemo() {
	// Create circuit breaker
	cb := golid.NewCircuitBreaker(golid.CircuitBreakerOptions{
		MaxFailures:  3,
		ResetTimeout: 2 * time.Second,
		HalfOpenMax:  2,
		OnStateChange: func(state golid.CircuitBreakerState) {
			fmt.Printf("   🔄 Circuit breaker state changed to: %s\n", state.String())
		},
	})

	// Simulate a service that fails initially then recovers
	failureCount := 0
	unreliableService := func() (string, error) {
		failureCount++
		if failureCount <= 4 {
			return "", fmt.Errorf("service failure #%d", failureCount)
		}
		return "Service is working!", nil
	}

	// Test circuit breaker behavior
	for i := 0; i < 10; i++ {
		fmt.Printf("   Request %d: ", i+1)

		if !cb.CanExecute() {
			fmt.Println("❌ Request rejected by circuit breaker")
			continue
		}

		result, err := unreliableService()
		if err != nil {
			cb.RecordFailure()
			fmt.Printf("❌ Service failed: %v\n", err)
		} else {
			cb.RecordSuccess()
			fmt.Printf("✅ Service succeeded: %s\n", result)
		}

		time.Sleep(500 * time.Millisecond)
	}

	// Show final metrics
	metrics := cb.GetMetrics()
	fmt.Printf("   Final metrics - Total: %d, Success: %d, Failed: %d, Rejected: %d\n",
		metrics.TotalRequests(), metrics.SuccessRequests(), metrics.FailedRequests(), metrics.RejectedRequests())
}

// ------------------------------------
// 🎯 Graceful Degradation Demo
// ------------------------------------

func gracefulDegradationDemo() {
	// Register some features
	golid.RegisterFeature("animations", func() bool {
		// Simulate feature detection
		return time.Now().Unix()%2 == 0
	}, func() interface{} {
		return "Static content (animations disabled)"
	})

	golid.RegisterFeature("advanced-ui", func() bool {
		return false // Simulate unsupported feature
	}, func() interface{} {
		return "Basic UI fallback"
	})

	// Test feature availability
	features := []string{"animations", "advanced-ui", "non-existent"}
	for _, feature := range features {
		available := golid.IsFeatureAvailable(feature)
		fmt.Printf("   Feature '%s': %s\n", feature,
			map[bool]string{true: "✅ Available", false: "❌ Unavailable"}[available])

		if !available {
			if fallback := golid.GetFeatureFallback(feature); fallback != nil {
				fmt.Printf("     Fallback: %v\n", fallback)
			}
		}
	}

	// Test degraded mode
	fmt.Println("   Testing degraded mode...")
	golid.EnableDegradedMode(golid.ModerateDegradation)
	fmt.Printf("   Degraded mode enabled: %s\n", golid.GetDegradationLevel().String())

	// Test feature availability in degraded mode
	testFeatures := []string{"animations", "lazy-loading", "core-rendering"}
	for _, feature := range testFeatures {
		enabled := golid.IsFeatureEnabled(feature)
		fmt.Printf("     Feature '%s' in degraded mode: %s\n", feature,
			map[bool]string{true: "✅ Enabled", false: "❌ Disabled"}[enabled])
	}

	// Disable degraded mode
	golid.DisableDegradedMode()
	fmt.Println("   Degraded mode disabled")
}

// ------------------------------------
// 🚀 Progressive Enhancement Demo
// ------------------------------------

func progressiveEnhancementDemo() {
	// Create progressive component
	progressive := golid.CreateProgressiveComponent(
		// Baseline implementation
		func() gomponents.Node {
			return html.Div(
				html.Class("baseline"),
				html.H3(gomponents.Text("Basic Content")),
				html.P(gomponents.Text("This is the baseline experience.")),
			)
		},
		// Enhanced implementation
		func() gomponents.Node {
			return html.Div(
				html.Class("enhanced"),
				html.H3(gomponents.Text("Enhanced Content")),
				html.P(gomponents.Text("This is the enhanced experience with animations!")),
				html.Div(
					html.Class("animation"),
					gomponents.Text("✨ Animated element"),
				),
			)
		},
		[]string{"animations", "advanced-ui"},
	)

	// Add fallback
	progressive.WithFallback(func() gomponents.Node {
		return html.Div(
			html.Class("fallback"),
			gomponents.Text("Content temporarily unavailable"),
		)
	})

	// Render component (would normally render to DOM)
	fmt.Println("   Progressive component created with baseline and enhanced versions")
	fmt.Println("   Component would render based on feature availability")

	// Test conditional rendering
	result := golid.ConditionalFeatureRender(
		"animations",
		func() gomponents.Node {
			return gomponents.Text("Animated content")
		},
		func() gomponents.Node {
			return gomponents.Text("Static content")
		},
	)

	fmt.Printf("   Conditional render result: %T\n", result)
}

// ------------------------------------
// 🔧 Error Recovery Strategies Demo
// ------------------------------------

func errorRecoveryDemo() {
	// Create error recovery coordinator
	coordinator := golid.NewErrorRecoveryCoordinator()

	// Register recovery strategies
	networkStrategy := &golid.RecoveryStrategyImpl{
		Name:     "network-errors",
		Priority: 1,
		CanHandle: func(err error) bool {
			return fmt.Sprintf("%v", err) == "network timeout"
		},
		Recover: func(err error) (interface{}, error) {
			fmt.Println("     🔄 Attempting network recovery...")
			return "Recovered network data", nil
		},
	}

	validationStrategy := &golid.RecoveryStrategyImpl{
		Name:     "validation-errors",
		Priority: 2,
		CanHandle: func(err error) bool {
			return fmt.Sprintf("%v", err) == "validation failed"
		},
		Recover: func(err error) (interface{}, error) {
			fmt.Println("     🔄 Attempting validation recovery...")
			return "Default valid data", nil
		},
	}

	coordinator.RegisterStrategy(networkStrategy)
	coordinator.RegisterStrategy(validationStrategy)

	// Add fallback chain
	coordinator.AddFallback(func(err error) interface{} {
		fmt.Printf("     🆘 Using final fallback for: %v\n", err)
		return "Emergency fallback data"
	})

	// Test recovery with different error types
	testErrors := []error{
		fmt.Errorf("network timeout"),
		fmt.Errorf("validation failed"),
		fmt.Errorf("unknown error"),
	}

	for i, testErr := range testErrors {
		fmt.Printf("   Test %d - Error: %v\n", i+1, testErr)
		result, err := coordinator.RecoverFromError(testErr)
		if err != nil {
			fmt.Printf("     ❌ Recovery failed: %v\n", err)
		} else {
			fmt.Printf("     ✅ Recovery successful: %v\n", result)
		}
	}
}

// ------------------------------------
// 🔒 Safe Signal Operations Demo
// ------------------------------------

func safeSignalDemo() {
	// Create a signal that might fail
	riskySignal := golid.NewSignal("initial value")

	// Simulate signal corruption
	go func() {
		time.Sleep(100 * time.Millisecond)
		// This would normally cause issues, but we'll handle it safely
		riskySignal.Set("corrupted value that might cause panic")
	}()

	// Safe signal access
	fmt.Println("   Testing safe signal access...")
	for i := 0; i < 3; i++ {
		safeValue := golid.SafeSignalAccess(riskySignal, "fallback value")
		fmt.Printf("   Safe access %d: %s\n", i+1, safeValue)
		time.Sleep(50 * time.Millisecond)
	}

	// Safe effect execution
	fmt.Println("   Testing safe effect execution...")
	golid.SafeEffectExecution(func() {
		fmt.Println("     Effect executing safely...")
		// This effect might panic, but it's protected
		if time.Now().Unix()%2 == 0 {
			panic("Effect panic simulation")
		}
		fmt.Println("     Effect completed successfully!")
	}, golid.ErrorBoundaryEnhancedOptions{
		MaxRetries: 2,
		OnError: func(err error, info *golid.ErrorInfo) {
			fmt.Printf("     Effect error handled: %v\n", err)
		},
	})

	// Test with error boundary wrapper
	result := golid.WithErrorBoundary(func() string {
		// This computation might fail
		if time.Now().Unix()%3 == 0 {
			panic("Computation failed")
		}
		return "Computation successful"
	}, "Computation fallback result")

	fmt.Printf("   Computation result: %s\n", result)
}
