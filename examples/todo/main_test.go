//go:build !js && !wasm

package main

import (
	"context"
	"testing"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/ozanturksever/uiwgo/internal/devserver"
)

func TestTodoApp(t *testing.T) {
	// Start the development server
	server := devserver.NewServer("todo", "localhost:0")
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

	// Navigate to the app and test todo functionality
	var todoCount int
	err := chromedp.Run(ctx,
		// Navigate to the application
		chromedp.Navigate(server.URL()),

		// Wait for the page to load
		chromedp.WaitVisible(`#new-todo-input`, chromedp.ByID),

		// Add a todo by typing and clicking the button
		chromedp.SendKeys(`#new-todo-input`, "Learn Go", chromedp.ByID),
		chromedp.Click(`#add-todo-btn`, chromedp.ByID),

		// Wait for todo to be added
		chromedp.Sleep(300*time.Millisecond),

		// Count todo items
		chromedp.Evaluate(`document.querySelectorAll('.todo-item').length`, &todoCount),
	)

	if err != nil {
		t.Fatalf("Browser automation failed: %v", err)
	}

	// Verify the results
	if todoCount != 1 {
		t.Errorf("Expected 1 todo to be added, got: %d", todoCount)
	}

	t.Logf("Test passed! Todo count: %d", todoCount)
}