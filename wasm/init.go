//go:build js && wasm

package wasm

import (
	"context"
	"errors"
	"time"

	"honnef.co/go/js/dom/v2"
	"github.com/ozanturksever/uiwgo/logutil"
)

var (
	// ErrTimeout is returned when initialization times out
	ErrTimeout = errors.New("wasm initialization timeout")
	// ErrDOMNotReady is returned when DOM is not ready
	ErrDOMNotReady = errors.New("DOM not ready")
	// ErrReadyCheckFailed is returned when custom ready check fails
	ErrReadyCheckFailed = errors.New("ready check failed")
	
	// initOnce ensures initialization only happens once
	initialized bool
)

// InitConfig holds configuration for WASM initialization
type InitConfig struct {
	// Timeout for the entire initialization process
	Timeout time.Duration
	// ReadySelector is an optional CSS selector to wait for visibility
	ReadySelector string
	// ReadyCheck is an optional custom readiness predicate
	ReadyCheck func() bool
	// RetryCount is the number of retries for ready checks
	RetryCount int
	// RetryInterval is the interval between retries
	RetryInterval time.Duration
}

// DefaultConfig returns a sensible default configuration
func DefaultConfig() InitConfig {
	return InitConfig{
		Timeout:       30 * time.Second,
		ReadySelector: "",
		ReadyCheck:    nil,
		RetryCount:    10,
		RetryInterval: 100 * time.Millisecond,
	}
}

// QuickConfig returns a fast configuration for simple cases
func QuickConfig() InitConfig {
	return InitConfig{
		Timeout:       5 * time.Second,
		ReadySelector: "",
		ReadyCheck:    nil,
		RetryCount:    5,
		RetryInterval: 50 * time.Millisecond,
	}
}

// Initialize waits for DOM readiness and optional custom checks
func Initialize(cfg InitConfig) error {
	logutil.Logf("Starting WASM initialization with timeout %v", cfg.Timeout)
	
	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), cfg.Timeout)
	defer cancel()
	
	// Step 1: Wait for DOM ready
	if err := waitForDOMReady(ctx); err != nil {
		return err
	}
	logutil.Log("DOM is ready")
	
	// Step 2: Wait for optional selector visibility
	if cfg.ReadySelector != "" {
		if err := waitForSelector(ctx, cfg.ReadySelector, cfg.RetryCount, cfg.RetryInterval); err != nil {
			return err
		}
		logutil.Logf("Selector %s is visible", cfg.ReadySelector)
	}
	
	// Step 3: Wait for optional custom ready check
	if cfg.ReadyCheck != nil {
		if err := waitForReadyCheck(ctx, cfg.ReadyCheck, cfg.RetryCount, cfg.RetryInterval); err != nil {
			return err
		}
		logutil.Log("Custom ready check passed")
	}
	
	initialized = true
	logutil.Log("WASM initialization completed successfully")
	return nil
}

// QuickInit initializes with default configuration
func QuickInit() error {
	return Initialize(DefaultConfig())
}

// InitAndThen initializes and then runs a callback function
func InitAndThen(then func() error, cfg InitConfig) error {
	if err := Initialize(cfg); err != nil {
		return err
	}
	
	if then != nil {
		logutil.Log("Running post-initialization callback")
		return then()
	}
	
	return nil
}

// IsInitialized returns true if WASM has been initialized
func IsInitialized() bool {
	return initialized
}

// ResetInitialized resets the initialization state (useful for testing)
func ResetInitialized() {
	initialized = false
}

// waitForDOMReady waits for the DOM to be ready
func waitForDOMReady(ctx context.Context) error {
	window := dom.GetWindow()
	document := window.Document()
	
	// Check if already ready
	if document.ReadyState() == "complete" || document.ReadyState() == "interactive" {
		return nil
	}
	
	// Wait for DOMContentLoaded or load event
	ready := make(chan bool, 1)
	errorCh := make(chan error, 1)
	
	// Listen for DOMContentLoaded
	document.AddEventListener("DOMContentLoaded", false, func(event dom.Event) {
		select {
		case ready <- true:
		default:
		}
	})
	
	// Also listen for load event as fallback
	window.AddEventListener("load", false, func(event dom.Event) {
		select {
		case ready <- true:
		default:
		}
	})
	
	// Wait for either ready signal or timeout
	select {
	case <-ready:
		return nil
	case <-ctx.Done():
		return ErrTimeout
	case err := <-errorCh:
		return err
	}
}

// waitForSelector waits for a CSS selector to be visible
func waitForSelector(ctx context.Context, selector string, retryCount int, retryInterval time.Duration) error {
	window := dom.GetWindow()
	document := window.Document()
	
	for i := 0; i < retryCount; i++ {
		// Check if context is cancelled
		select {
		case <-ctx.Done():
			return ErrTimeout
		default:
		}
		
		// Try to find the element
		element := document.QuerySelector(selector)
		if element != nil {
			// Check if element is visible
			style := element.Style()
			display := style.GetPropertyValue("display")
			visibility := style.GetPropertyValue("visibility")
			
			if display != "none" && visibility != "hidden" {
				return nil
			}
		}
		
		// Wait before retry
		time.Sleep(retryInterval)
	}
	
	return errors.New("selector not found or not visible: " + selector)
}

// waitForReadyCheck waits for a custom ready check to pass
func waitForReadyCheck(ctx context.Context, readyCheck func() bool, retryCount int, retryInterval time.Duration) error {
	for i := 0; i < retryCount; i++ {
		// Check if context is cancelled
		select {
		case <-ctx.Done():
			return ErrTimeout
		default:
		}
		
		// Run the ready check
		if readyCheck() {
			return nil
		}
		
		// Wait before retry
		time.Sleep(retryInterval)
	}
	
	return ErrReadyCheckFailed
}

// WaitForElement waits for an element to appear in the DOM
func WaitForElement(selector string, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	
	return waitForSelector(ctx, selector, 50, 100*time.Millisecond)
}

// WaitForFunction waits for a function to return true
func WaitForFunction(fn func() bool, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	
	return waitForReadyCheck(ctx, fn, 50, 100*time.Millisecond)
}