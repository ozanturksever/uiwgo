//go:build !js || !wasm

package wasm

import (
	"errors"
	"time"

	"github.com/ozanturksever/uiwgo/logutil"
)

var (
	// ErrNotSupported is returned on non-WASM platforms
	ErrNotSupported = errors.New("WASM operations not supported on this platform")
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
	logutil.Log("WASM initialization not supported on this platform")
	return ErrNotSupported
}

// QuickInit initializes with default configuration
func QuickInit() error {
	logutil.Log("WASM quick initialization not supported on this platform")
	return ErrNotSupported
}

// InitAndThen initializes and then runs a callback function
func InitAndThen(then func() error, cfg InitConfig) error {
	logutil.Log("WASM InitAndThen not supported on this platform")
	return ErrNotSupported
}

// IsInitialized returns false on non-WASM platforms
func IsInitialized() bool {
	return false
}

// ResetInitialized is a no-op on non-WASM platforms
func ResetInitialized() {
	// No-op
}

// WaitForElement returns error on non-WASM platforms
func WaitForElement(selector string, timeout time.Duration) error {
	logutil.Log("WaitForElement not supported on this platform")
	return ErrNotSupported
}

// WaitForFunction returns error on non-WASM platforms
func WaitForFunction(fn func() bool, timeout time.Duration) error {
	logutil.Log("WaitForFunction not supported on this platform")
	return ErrNotSupported
}