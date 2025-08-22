package main

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/ozanturksever/uiwgo/internal/devserver"
)

func TestCounterApp(t *testing.T) {
	// Create and start the development server
	server := devserver.NewServer("counter", "localhost:0")
	if err := server.Start(); err != nil {
		t.Fatalf("Failed to start dev server: %v", err)
	}
	defer server.Stop()

	// Create a context with timeout for the entire test
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Create Chrome options for visible browser (not headless)
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", false), // Make browser visible
		chromedp.Flag("disable-gpu", false),
		chromedp.Flag("no-sandbox", true),
		chromedp.Flag("disable-dev-shm-usage", true),
	)

	// Create a new allocator context with the options
	allocCtx, cancel := chromedp.NewExecAllocator(ctx, opts...)
	defer cancel()

	// Create a new browser context
	browserCtx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()

	// Navigate to the test server and perform the test
	var countText string
	err := chromedp.Run(browserCtx,
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