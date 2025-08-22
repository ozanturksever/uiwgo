//go:build !js && !wasm

package main

import (
	"context"
	"testing"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/ozanturksever/uiwgo/internal/devserver"
)

func TestResourceApp(t *testing.T) {
	// Start the development server
	server := devserver.NewServer("resource", "localhost:0")
	if err := server.Start(); err != nil {
		t.Fatalf("Failed to start dev server: %v", err)
	}
	defer server.Stop()

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Create chromedp context with visible browser for debugging
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", false),
		chromedp.Flag("disable-gpu", false),
		chromedp.Flag("no-sandbox", true),
	)
	allocCtx, cancel := chromedp.NewExecAllocator(ctx, opts...)
	defer cancel()

	ctx, cancel = chromedp.NewContext(allocCtx)
	defer cancel()

	// Navigate to the app and test resource loading
	var loadingText, user1Text, errorText string
	err := chromedp.Run(ctx,
		// Navigate to the application
		chromedp.Navigate(server.URL()),

		// Wait for the page to load
		chromedp.WaitVisible(`#load-user1-btn`, chromedp.ByID),

		// Test loading User 1 (success case)
		chromedp.Click(`#load-user1-btn`, chromedp.ByID),

		// Wait for loading state and capture loading text
		chromedp.Sleep(100*time.Millisecond), // Brief moment to catch loading state
		chromedp.Text(`#user-display`, &loadingText, chromedp.ByID),

		// Wait for the request to complete (900ms delay + processing time)
		chromedp.Sleep(1200*time.Millisecond),

		// Capture the successful result
		chromedp.Text(`#user-display`, &user1Text, chromedp.ByID),

		// Test loading User 2 (error case)
		chromedp.Click(`#load-user2-btn`, chromedp.ByID),

		// Wait for the error to appear
		chromedp.Sleep(1200*time.Millisecond),

		// Capture the error result
		chromedp.Text(`#user-display`, &errorText, chromedp.ByID),
	)

	if err != nil {
		t.Fatalf("Browser automation failed: %v", err)
	}

	// Verify the results
	if loadingText != "Loading..." {
		t.Errorf("Expected loading text to be 'Loading...', got: %s", loadingText)
	}

	if user1Text != "User #1\n\nName: User-1" {
		t.Errorf("Expected user1 text to contain user info, got: %s", user1Text)
	}

	if errorText != "Error: user 2 not found" {
		t.Errorf("Expected error text to contain 'user 2 not found', got: %s", errorText)
	}

	t.Logf("Test passed! Loading: %s, User1: %s, Error: %s", loadingText, user1Text, errorText)
}