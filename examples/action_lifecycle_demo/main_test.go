//go:build !js && !wasm

package main

import (
	"strings"
	"testing"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/ozanturksever/uiwgo/internal/testhelpers"
)

// TestActionLifecycleDemo verifies that the action lifecycle demo works correctly
func TestActionLifecycleDemo(t *testing.T) {
	// Start the development server with action_lifecycle_demo example
	server := testhelpers.NewViteServer("action_lifecycle_demo", "localhost:0")
	if err := server.Start(); err != nil {
		t.Fatalf("Failed to start dev server: %v", err)
	}
	defer server.Stop()

	// Create chromedp context with extended timeout for WASM build
	chromedpCtx := testhelpers.MustNewChromedpContext(testhelpers.ExtendedTimeoutConfig())
	defer chromedpCtx.Cancel()

	// Variable to capture the count text
	var countText string

	err := chromedp.Run(chromedpCtx.Ctx,
		// Navigate to the application
		chromedp.Navigate(server.URL()),

		// Wait for the page to load and WASM to initialize
		chromedp.WaitVisible("body", chromedp.ByQuery),
		chromedp.Sleep(2*time.Second), // Give time for WASM to initialize

		// Check initial count
		chromedp.Text(`//p[contains(text(), 'Count:')]`, &countText, chromedp.BySearch),
	)
	if err != nil {
		t.Fatalf("Browser automation failed: %v", err)
	}

	// Verify initial count is 0
	if !strings.Contains(countText, "Count: 0") {
		t.Errorf("Expected initial count to be 0, but got: %s", countText)
	}

	// Test increment button
	err = chromedp.Run(chromedpCtx.Ctx,
		// Click increment button 3 times
		chromedp.Click("#inc-btn", chromedp.ByID),
		chromedp.Sleep(300*time.Millisecond),
		chromedp.Click("#inc-btn", chromedp.ByID),
		chromedp.Sleep(300*time.Millisecond),
		chromedp.Click("#inc-btn", chromedp.ByID),
		chromedp.Sleep(300*time.Millisecond),

		// Get updated count
		chromedp.Text(`//p[contains(text(), 'Count:')]`, &countText, chromedp.BySearch),
	)
	if err != nil {
		t.Fatalf("Failed to click increment or get count: %v", err)
	}

	// Verify count is now 3
	if !strings.Contains(countText, "Count: 3") {
		t.Errorf("Expected count to be 3 after 3 increments, but got: %s", countText)
	}

	// Test decrement button
	err = chromedp.Run(chromedpCtx.Ctx,
		// Click decrement button once
		chromedp.Click("#dec-btn", chromedp.ByID),
		chromedp.Sleep(300*time.Millisecond),

		// Get updated count
		chromedp.Text(`//p[contains(text(), 'Count:')]`, &countText, chromedp.BySearch),
	)
	if err != nil {
		t.Fatalf("Failed to click decrement or get count: %v", err)
	}

	// Verify count is now 2
	if !strings.Contains(countText, "Count: 2") {
		t.Errorf("Expected count to be 2 after decrement, but got: %s", countText)
	}

	t.Logf("Test passed! Lifecycle helpers work correctly. Final count: %s", countText)
}
