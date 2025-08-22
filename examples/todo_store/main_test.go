//go:build !js && !wasm

package main

import (
	"context"
	"testing"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/ozanturksever/uiwgo/internal/devserver"
)

func TestTodoStoreApp(t *testing.T) {
	// Start the development server
	server := devserver.NewServer("todo_store", "localhost:0")
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

	// Navigate to the app and test todo store functionality
	var todoCount int
	err := chromedp.Run(ctx,
		// Navigate to the application
		chromedp.Navigate(server.URL()),

		// Wait for the page to load
		chromedp.WaitVisible(`#new-todo-input`, chromedp.ByID),

		// Add one todo
		chromedp.SendKeys(`#new-todo-input`, "Test Todo", chromedp.ByID),
		chromedp.Click(`#add-todo-btn`, chromedp.ByID),

		// Wait for todo to be added
		chromedp.Sleep(500*time.Millisecond),

		// Count todos
		chromedp.Evaluate(`document.querySelectorAll('.todo-item').length`, &todoCount),
	)

	if err != nil {
		t.Fatalf("Browser automation failed: %v", err)
	}

	// Verify the results
	if todoCount != 1 {
		t.Errorf("Expected 1 todo item, got: %d", todoCount)
	}

	t.Logf("Test passed! Todo count: %d", todoCount)
}