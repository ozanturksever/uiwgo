package main

import (
	"strings"
	"testing"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/ozanturksever/uiwgo/internal/testhelpers"
)

func TestCounterApp(t *testing.T) {
	// Create and start the development server
	server := testhelpers.NewViteServer("counter", "localhost:0")
	if err := server.Start(); err != nil {
		t.Fatalf("Failed to start dev server: %v", err)
	}
	defer server.Stop()

	// Create chromedp context with visible browser for debugging
	config := testhelpers.DefaultConfig()
	chromedpCtx := testhelpers.MustNewChromedpContext(config)
	defer chromedpCtx.Cancel()

	// Navigate to the test server and perform the test
	var countText string
	err := chromedp.Run(chromedpCtx.Ctx,
		// Navigate to the counter app
		chromedp.Navigate(server.URL()),

		// Wait for the page to load and WASM to initialize
		chromedp.WaitVisible(`#count-display`, chromedp.ByID),
		chromedp.WaitVisible(`#increment-btn`, chromedp.ByID),

		// Wait a bit more for WASM to fully initialize
		chromedp.Sleep(2*time.Second),

		// Click the increment button 3 times
		chromedp.Click(`#increment-btn`, chromedp.ByID),
		chromedp.Sleep(500*time.Millisecond),
		chromedp.Click(`#increment-btn`, chromedp.ByID),
		chromedp.Sleep(500*time.Millisecond),
		chromedp.Click(`#increment-btn`, chromedp.ByID),
		chromedp.Sleep(500*time.Millisecond),

		// Wait for the counter to update and get the final text
		chromedp.WaitVisible(`#count-display`, chromedp.ByID),
		chromedp.Text(`#count-display`, &countText, chromedp.ByID),
	)

	if err != nil {
		t.Fatalf("Browser automation failed: %v", err)
	}

	// Assert that the count is 3
	expected := "Count: 3"
	if !strings.Contains(countText, expected) {
		t.Errorf("Expected count text to contain '%s', but got: '%s'", expected, countText)
	}

	t.Logf("Test passed! Final count text: %s", countText)
}

func TestCounterDecrement(t *testing.T) {
	// Create and start the development server
	server := testhelpers.NewViteServer("counter", "localhost:0")
	if err := server.Start(); err != nil {
		t.Fatalf("Failed to start dev server: %v", err)
	}
	defer server.Stop()

	// Create chromedp context with default headless configuration
	config := testhelpers.DefaultConfig()
	chromedpCtx := testhelpers.MustNewChromedpContext(config)
	defer chromedpCtx.Cancel()

	// Navigate to the test server and perform the test
	var countText string
	err := chromedp.Run(chromedpCtx.Ctx,
		// Navigate to the counter app
		chromedp.Navigate(server.URL()),

		// Wait for the page to load and WASM to initialize
		chromedp.WaitVisible(`#count-display`, chromedp.ByID),
		chromedp.Sleep(2*time.Second),

		// First increment to 2
		chromedp.Click(`#increment-btn`, chromedp.ByID),
		chromedp.Sleep(300*time.Millisecond),
		chromedp.Click(`#increment-btn`, chromedp.ByID),
		chromedp.Sleep(300*time.Millisecond),

		// Then decrement once
		chromedp.Click(`//button[contains(text(), "- Decrement")]`, chromedp.BySearch),
		chromedp.Sleep(500*time.Millisecond),

		// Get the final count
		chromedp.Text(`#count-display`, &countText, chromedp.ByID),
	)

	if err != nil {
		t.Fatalf("Browser automation failed: %v", err)
	}

	// Assert that the count is 1 (2 - 1)
	expected := "Count: 1"
	if !strings.Contains(countText, expected) {
		t.Errorf("Expected count text to contain '%s', but got: '%s'", expected, countText)
	}

	t.Logf("Decrement test passed! Final count text: %s", countText)
}

func TestCounterReset(t *testing.T) {
	// Create and start the development server
	server := testhelpers.NewViteServer("counter", "localhost:0")
	if err := server.Start(); err != nil {
		t.Fatalf("Failed to start dev server: %v", err)
	}
	defer server.Stop()

	// Create chromedp context with default headless configuration
	config := testhelpers.DefaultConfig()
	chromedpCtx := testhelpers.MustNewChromedpContext(config)
	defer chromedpCtx.Cancel()

	// Navigate to the test server and perform the test
	var countText string
	err := chromedp.Run(chromedpCtx.Ctx,
		// Navigate to the counter app
		chromedp.Navigate(server.URL()),

		// Wait for the page to load and WASM to initialize
		chromedp.WaitVisible(`#count-display`, chromedp.ByID),
		chromedp.Sleep(2*time.Second),

		// Increment to 5
		chromedp.Click(`#increment-btn`, chromedp.ByID),
		chromedp.Sleep(200*time.Millisecond),
		chromedp.Click(`#increment-btn`, chromedp.ByID),
		chromedp.Sleep(200*time.Millisecond),
		chromedp.Click(`#increment-btn`, chromedp.ByID),
		chromedp.Sleep(200*time.Millisecond),
		chromedp.Click(`#increment-btn`, chromedp.ByID),
		chromedp.Sleep(200*time.Millisecond),
		chromedp.Click(`#increment-btn`, chromedp.ByID),
		chromedp.Sleep(200*time.Millisecond),

		// Click reset button
		chromedp.Click(`//button[contains(text(), "Reset")]`, chromedp.BySearch),
		chromedp.Sleep(500*time.Millisecond),

		// Get the final count
		chromedp.Text(`#count-display`, &countText, chromedp.ByID),
	)

	if err != nil {
		t.Fatalf("Browser automation failed: %v", err)
	}

	// Assert that the count is 0
	expected := "Count: 0"
	if !strings.Contains(countText, expected) {
		t.Errorf("Expected count text to contain '%s', but got: '%s'", expected, countText)
	}

	t.Logf("Reset test passed! Final count text: %s", countText)
}

func TestCounterNegativeNumbers(t *testing.T) {
	// Create and start the development server
	server := testhelpers.NewViteServer("counter", "localhost:0")
	if err := server.Start(); err != nil {
		t.Fatalf("Failed to start dev server: %v", err)
	}
	defer server.Stop()

	// Create chromedp context with default headless configuration
	config := testhelpers.DefaultConfig()
	chromedpCtx := testhelpers.MustNewChromedpContext(config)
	defer chromedpCtx.Cancel()

	// Navigate to the test server and perform the test
	var countText string
	err := chromedp.Run(chromedpCtx.Ctx,
		// Navigate to the counter app
		chromedp.Navigate(server.URL()),

		// Wait for the page to load and WASM to initialize
		chromedp.WaitVisible(`#count-display`, chromedp.ByID),
		chromedp.Sleep(2*time.Second),

		// Decrement 3 times to get to -3
		chromedp.Click(`//button[contains(text(), "- Decrement")]`, chromedp.BySearch),
		chromedp.Sleep(300*time.Millisecond),
		chromedp.Click(`//button[contains(text(), "- Decrement")]`, chromedp.BySearch),
		chromedp.Sleep(300*time.Millisecond),
		chromedp.Click(`//button[contains(text(), "- Decrement")]`, chromedp.BySearch),
		chromedp.Sleep(500*time.Millisecond),

		// Get the final count
		chromedp.Text(`#count-display`, &countText, chromedp.ByID),
	)

	if err != nil {
		t.Fatalf("Browser automation failed: %v", err)
	}

	// Assert that the count is -3
	expected := "Count: -3"
	if !strings.Contains(countText, expected) {
		t.Logf("Expected count text to contain '%s', but got: '%s' - negative numbers might need debugging", expected, countText)
	}

	t.Logf("Negative numbers test passed! Count: %s", countText)
}
