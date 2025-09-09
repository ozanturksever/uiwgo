//go:build !js && !wasm

package testhelpers

import (
	"context"
	"time"

	"github.com/chromedp/chromedp"
)

// ChromedpConfig holds configuration options for chromedp browser setup
type ChromedpConfig struct {
	// Headless determines if the browser runs in headless mode
	Headless bool
	// Timeout sets the context timeout for the entire test
	Timeout time.Duration
	// DisableGPU disables GPU acceleration
	DisableGPU bool
	// NoSandbox disables the sandbox
	NoSandbox bool
	// DisableDevShmUsage disables /dev/shm usage
	DisableDevShmUsage bool
	// AdditionalFlags allows adding custom Chrome flags
	AdditionalFlags []chromedp.ExecAllocatorOption
}

// DefaultConfig returns a sensible default configuration for chromedp tests
func DefaultConfig() ChromedpConfig {
	return ChromedpConfig{
		Headless:           true,
		Timeout:            5 * time.Second,
		DisableGPU:         true,
		NoSandbox:          true,
		DisableDevShmUsage: true,
	}
}

// VisibleConfig returns a configuration for visible browser testing (useful for debugging)
func VisibleConfig() ChromedpConfig {
	return ChromedpConfig{
		Headless:           false,
		Timeout:            5 * time.Second,
		DisableGPU:         false,
		NoSandbox:          true,
		DisableDevShmUsage: true,
	}
}

// ExtendedTimeoutConfig returns a configuration with longer timeout for complex tests
func ExtendedTimeoutConfig() ChromedpConfig {
	config := DefaultConfig()
	config.Timeout = 5 * time.Second
	return config
}

// ChromedpTestContext holds the context and cancel function for a chromedp test
type ChromedpTestContext struct {
	Ctx    context.Context
	Cancel context.CancelFunc
}

// NewChromedpContext creates a new chromedp context with the given configuration
// Returns a ChromedpTestContext that should be cleaned up with defer ctx.Cancel()
func NewChromedpContext(config ChromedpConfig) (*ChromedpTestContext, error) {
	// Create base context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), config.Timeout)

	// Build Chrome options
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", config.Headless),
		chromedp.Flag("disable-gpu", config.DisableGPU),
		chromedp.Flag("no-sandbox", config.NoSandbox),
	)

	// Add disable-dev-shm-usage if requested
	if config.DisableDevShmUsage {
		opts = append(opts, chromedp.Flag("disable-dev-shm-usage", true))
	}

	// Add any additional flags
	opts = append(opts, config.AdditionalFlags...)

	// Create allocator context
	allocCtx, allocCancel := chromedp.NewExecAllocator(ctx, opts...)

	// Create browser context
	browserCtx, browserCancel := chromedp.NewContext(allocCtx)

	// Create a combined cancel function that cleans up all contexts
	combinedCancel := func() {
		browserCancel()
		allocCancel()
		cancel()
	}

	return &ChromedpTestContext{
		Ctx:    browserCtx,
		Cancel: combinedCancel,
	}, nil
}

// MustNewChromedpContext is like NewChromedpContext but panics on error
// Useful for test setup where you want to fail fast
func MustNewChromedpContext(config ChromedpConfig) *ChromedpTestContext {
	ctx, err := NewChromedpContext(config)
	if err != nil {
		panic(err)
	}
	return ctx
}

// CommonTestActions provides common chromedp actions used across tests
type CommonTestActions struct{}

// WaitForWASMInit waits for WASM to initialize by waiting for a visible element and adding a delay
func (CommonTestActions) WaitForWASMInit(selector string, delay time.Duration) chromedp.Action {
	return chromedp.Tasks{
		chromedp.WaitVisible(selector, chromedp.ByQuery),
		chromedp.Sleep(delay),
	}
}

// NavigateAndWaitForLoad navigates to a URL and waits for the page to load
func (CommonTestActions) NavigateAndWaitForLoad(url, waitSelector string) chromedp.Action {
	return chromedp.Tasks{
		chromedp.Navigate(url),
		chromedp.WaitVisible(waitSelector, chromedp.ByQuery),
		chromedp.Sleep(500 * time.Millisecond), // Basic WASM init time
	}
}

// ClickAndWait clicks an element and waits for a specified duration
func (CommonTestActions) ClickAndWait(selector string, wait time.Duration) chromedp.Action {
	return chromedp.Tasks{
		chromedp.Click(selector, chromedp.ByQuery),
		chromedp.Sleep(wait),
	}
}

// SendKeysAndWait sends keys to an element and waits for a specified duration
func (CommonTestActions) SendKeysAndWait(selector, text string, wait time.Duration) chromedp.Action {
	return chromedp.Tasks{
		chromedp.SendKeys(selector, text, chromedp.ByQuery),
		chromedp.Sleep(wait),
	}
}

// Global instance for easy access to common actions
var Actions = CommonTestActions{}