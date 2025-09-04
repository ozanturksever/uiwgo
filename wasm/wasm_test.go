//go:build !js || !wasm

package wasm

import (
	"testing"
	"time"
)

// TestDefaultConfig tests the default configuration
func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	
	if cfg.Timeout != 30*time.Second {
		t.Errorf("Expected timeout to be 30s, got %v", cfg.Timeout)
	}
	
	if cfg.ReadySelector != "" {
		t.Errorf("Expected empty ready selector, got %s", cfg.ReadySelector)
	}
	
	if cfg.ReadyCheck != nil {
		t.Error("Expected ready check to be nil")
	}
	
	if cfg.RetryCount != 10 {
		t.Errorf("Expected retry count to be 10, got %d", cfg.RetryCount)
	}
	
	if cfg.RetryInterval != 100*time.Millisecond {
		t.Errorf("Expected retry interval to be 100ms, got %v", cfg.RetryInterval)
	}
}

// TestQuickConfig tests the quick configuration
func TestQuickConfig(t *testing.T) {
	cfg := QuickConfig()
	
	if cfg.Timeout != 5*time.Second {
		t.Errorf("Expected timeout to be 5s, got %v", cfg.Timeout)
	}
	
	if cfg.ReadySelector != "" {
		t.Errorf("Expected empty ready selector, got %s", cfg.ReadySelector)
	}
	
	if cfg.ReadyCheck != nil {
		t.Error("Expected ready check to be nil")
	}
	
	if cfg.RetryCount != 5 {
		t.Errorf("Expected retry count to be 5, got %d", cfg.RetryCount)
	}
	
	if cfg.RetryInterval != 50*time.Millisecond {
		t.Errorf("Expected retry interval to be 50ms, got %v", cfg.RetryInterval)
	}
}

// TestInitializeOnNonWASM tests initialization on non-WASM platforms
func TestInitializeOnNonWASM(t *testing.T) {
	// Reset initialization state
	ResetInitialized()
	
	// Should not be initialized initially
	if IsInitialized() {
		t.Error("Expected IsInitialized to be false initially")
	}
	
	// Initialize should return error on non-WASM
	err := Initialize(DefaultConfig())
	if err != ErrNotSupported {
		t.Errorf("Expected ErrNotSupported, got %v", err)
	}
	
	// Should still not be initialized
	if IsInitialized() {
		t.Error("Expected IsInitialized to be false after failed initialization")
	}
}

// TestQuickInitOnNonWASM tests quick initialization on non-WASM platforms
func TestQuickInitOnNonWASM(t *testing.T) {
	// Reset initialization state
	ResetInitialized()
	
	// QuickInit should return error on non-WASM
	err := QuickInit()
	if err != ErrNotSupported {
		t.Errorf("Expected ErrNotSupported, got %v", err)
	}
	
	// Should not be initialized
	if IsInitialized() {
		t.Error("Expected IsInitialized to be false after failed quick init")
	}
}

// TestInitAndThenOnNonWASM tests InitAndThen on non-WASM platforms
func TestInitAndThenOnNonWASM(t *testing.T) {
	// Reset initialization state
	ResetInitialized()
	
	callbackCalled := false
	callback := func() error {
		callbackCalled = true
		return nil
	}
	
	// InitAndThen should return error on non-WASM
	err := InitAndThen(callback, DefaultConfig())
	if err != ErrNotSupported {
		t.Errorf("Expected ErrNotSupported, got %v", err)
	}
	
	// Callback should not be called
	if callbackCalled {
		t.Error("Expected callback not to be called on failed initialization")
	}
	
	// Should not be initialized
	if IsInitialized() {
		t.Error("Expected IsInitialized to be false after failed InitAndThen")
	}
}

// TestWaitForElementOnNonWASM tests WaitForElement on non-WASM platforms
func TestWaitForElementOnNonWASM(t *testing.T) {
	err := WaitForElement("#test", 1*time.Second)
	if err != ErrNotSupported {
		t.Errorf("Expected ErrNotSupported, got %v", err)
	}
}

// TestWaitForFunctionOnNonWASM tests WaitForFunction on non-WASM platforms
func TestWaitForFunctionOnNonWASM(t *testing.T) {
	fn := func() bool { return true }
	err := WaitForFunction(fn, 1*time.Second)
	if err != ErrNotSupported {
		t.Errorf("Expected ErrNotSupported, got %v", err)
	}
}

// TestResetInitialized tests the reset functionality
func TestResetInitialized(t *testing.T) {
	// Reset should work without error
	ResetInitialized()
	
	// Should not be initialized after reset
	if IsInitialized() {
		t.Error("Expected IsInitialized to be false after reset")
	}
}

// TestInitConfigCustomization tests custom configuration
func TestInitConfigCustomization(t *testing.T) {
	customTimeout := 15 * time.Second
	customSelector := "#app"
	customRetryCount := 20
	customRetryInterval := 200 * time.Millisecond
	
	cfg := InitConfig{
		Timeout:       customTimeout,
		ReadySelector: customSelector,
		ReadyCheck:    func() bool { return true },
		RetryCount:    customRetryCount,
		RetryInterval: customRetryInterval,
	}
	
	if cfg.Timeout != customTimeout {
		t.Errorf("Expected timeout to be %v, got %v", customTimeout, cfg.Timeout)
	}
	
	if cfg.ReadySelector != customSelector {
		t.Errorf("Expected ready selector to be %s, got %s", customSelector, cfg.ReadySelector)
	}
	
	if cfg.ReadyCheck == nil {
		t.Error("Expected ready check to be set")
	}
	
	if cfg.RetryCount != customRetryCount {
		t.Errorf("Expected retry count to be %d, got %d", customRetryCount, cfg.RetryCount)
	}
	
	if cfg.RetryInterval != customRetryInterval {
		t.Errorf("Expected retry interval to be %v, got %v", customRetryInterval, cfg.RetryInterval)
	}
	
	// Test the ready check function
	if !cfg.ReadyCheck() {
		t.Error("Expected ready check to return true")
	}
}

// TestErrorConstants tests that error constants are properly defined
func TestErrorConstants(t *testing.T) {
	if ErrNotSupported == nil {
		t.Error("Expected ErrNotSupported to be defined")
	}
	
	if ErrNotSupported.Error() == "" {
		t.Error("Expected ErrNotSupported to have a message")
	}
	
	// Test that the error message is descriptive
	expectedMsg := "WASM operations not supported on this platform"
	if ErrNotSupported.Error() != expectedMsg {
		t.Errorf("Expected error message '%s', got '%s'", expectedMsg, ErrNotSupported.Error())
	}
}