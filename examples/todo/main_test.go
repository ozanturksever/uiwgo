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

func TestTodoRemoval(t *testing.T) {
	// Start the development server
	server := devserver.NewServer("todo", "localhost:0")
	if err := server.Start(); err != nil {
		t.Fatalf("Failed to start dev server: %v", err)
	}
	defer server.Stop()

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Create chromedp context
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", false),
		chromedp.Flag("disable-gpu", false),
		chromedp.Flag("no-sandbox", true),
	)
	allocCtx, cancel := chromedp.NewExecAllocator(ctx, opts...)
	defer cancel()

	ctx, cancel = chromedp.NewContext(allocCtx)
	defer cancel()

	var todoCount int
	err := chromedp.Run(ctx,
		// Navigate to the application
		chromedp.Navigate(server.URL()),

		// Wait for the page to load
		chromedp.WaitVisible(`#new-todo-input`, chromedp.ByID),

		// Add two todos
		chromedp.SendKeys(`#new-todo-input`, "First Todo", chromedp.ByID),
		chromedp.Click(`#add-todo-btn`, chromedp.ByID),
		chromedp.Sleep(200*time.Millisecond),

		chromedp.SendKeys(`#new-todo-input`, "Second Todo", chromedp.ByID),
		chromedp.Click(`#add-todo-btn`, chromedp.ByID),
		chromedp.Sleep(200*time.Millisecond),

		// Verify we have 2 todos
		chromedp.Evaluate(`document.querySelectorAll('.todo-item').length`, &todoCount),
	)

	if err != nil {
		t.Fatalf("Browser automation failed: %v", err)
	}

	if todoCount != 2 {
		t.Errorf("Expected 2 todos to be added, got: %d", todoCount)
	}

	// Remove the first todo
	err = chromedp.Run(ctx,
		// Wait for the destroy button to be visible
		chromedp.WaitVisible(`.todo-destroy`, chromedp.ByQuery),
		// Click the destroy button of the first todo using JavaScript
		chromedp.Evaluate(`document.querySelector('.todo-destroy').click()`, nil),
		chromedp.Sleep(200*time.Millisecond),

		// Count remaining todos
		chromedp.Evaluate(`document.querySelectorAll('.todo-item').length`, &todoCount),
	)

	if err != nil {
		t.Fatalf("Browser automation failed: %v", err)
	}

	if todoCount != 1 {
		t.Errorf("Expected 1 todo after removal, got: %d", todoCount)
	}

	t.Logf("Test passed! Todo removal working correctly")
}

func TestTodoMarking(t *testing.T) {
	// Start the development server
	server := devserver.NewServer("todo", "localhost:0")
	if err := server.Start(); err != nil {
		t.Fatalf("Failed to start dev server: %v", err)
	}
	defer server.Stop()

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Create chromedp context
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", false),
		chromedp.Flag("disable-gpu", false),
		chromedp.Flag("no-sandbox", true),
	)
	allocCtx, cancel := chromedp.NewExecAllocator(ctx, opts...)
	defer cancel()

	ctx, cancel = chromedp.NewContext(allocCtx)
	defer cancel()

	var isChecked bool
	err := chromedp.Run(ctx,
		// Navigate to the application
		chromedp.Navigate(server.URL()),

		// Wait for the page to load
		chromedp.WaitVisible(`#new-todo-input`, chromedp.ByID),

		// Add a todo
		chromedp.SendKeys(`#new-todo-input`, "Test Todo", chromedp.ByID),
		chromedp.Click(`#add-todo-btn`, chromedp.ByID),
		chromedp.Sleep(200*time.Millisecond),

		// Click the checkbox to mark as completed
		chromedp.Click(`.todo-toggle`, chromedp.ByQuery),
		chromedp.Sleep(200*time.Millisecond),

		// Check if the checkbox is checked
		chromedp.Evaluate(`document.querySelector('.todo-toggle').checked`, &isChecked),
	)

	if err != nil {
		t.Fatalf("Browser automation failed: %v", err)
	}

	if !isChecked {
		t.Errorf("Expected todo to be marked as completed")
	}

	// Test unmarking
	err = chromedp.Run(ctx,
		// Click the checkbox again to unmark
		chromedp.Click(`.todo-toggle`, chromedp.ByQuery),
		chromedp.Sleep(200*time.Millisecond),

		// Check if the checkbox is unchecked
		chromedp.Evaluate(`document.querySelector('.todo-toggle').checked`, &isChecked),
	)

	if err != nil {
		t.Fatalf("Browser automation failed: %v", err)
	}

	if isChecked {
		t.Errorf("Expected todo to be unmarked")
	}

	t.Logf("Test passed! Todo marking/unmarking working correctly")
}

func TestClearMarkedTodos(t *testing.T) {
	// Start the development server
	server := devserver.NewServer("todo", "localhost:0")
	if err := server.Start(); err != nil {
		t.Fatalf("Failed to start dev server: %v", err)
	}
	defer server.Stop()

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Create chromedp context
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", false),
		chromedp.Flag("disable-gpu", false),
		chromedp.Flag("no-sandbox", true),
	)
	allocCtx, cancel := chromedp.NewExecAllocator(ctx, opts...)
	defer cancel()

	ctx, cancel = chromedp.NewContext(allocCtx)
	defer cancel()

	var todoCount int
	var clearBtnVisible bool
	err := chromedp.Run(ctx,
		// Navigate to the application
		chromedp.Navigate(server.URL()),

		// Wait for the page to load
		chromedp.WaitVisible(`#new-todo-input`, chromedp.ByID),

		// Add three todos
		chromedp.SendKeys(`#new-todo-input`, "Todo 1", chromedp.ByID),
		chromedp.Click(`#add-todo-btn`, chromedp.ByID),
		chromedp.Sleep(200*time.Millisecond),

		chromedp.SendKeys(`#new-todo-input`, "Todo 2", chromedp.ByID),
		chromedp.Click(`#add-todo-btn`, chromedp.ByID),
		chromedp.Sleep(200*time.Millisecond),

		chromedp.SendKeys(`#new-todo-input`, "Todo 3", chromedp.ByID),
		chromedp.Click(`#add-todo-btn`, chromedp.ByID),
		chromedp.Sleep(200*time.Millisecond),

		// Mark first two todos as completed
		chromedp.Evaluate(`document.querySelectorAll('.todo-toggle')[0].click()`, nil),
		chromedp.Sleep(200*time.Millisecond),
		chromedp.Evaluate(`document.querySelectorAll('.todo-toggle')[1].click()`, nil),
		chromedp.Sleep(200*time.Millisecond),

		// Check if clear completed button is visible
		chromedp.Evaluate(`document.querySelector('#clear-completed-btn') !== null`, &clearBtnVisible),
	)

	if err != nil {
		t.Fatalf("Browser automation failed: %v", err)
	}

	if !clearBtnVisible {
		t.Errorf("Expected clear completed button to be visible")
	}

	// Clear completed todos
	err = chromedp.Run(ctx,
		// Click clear completed button
		chromedp.Click(`#clear-completed-btn`, chromedp.ByID),
		chromedp.Sleep(200*time.Millisecond),

		// Count remaining todos
		chromedp.Evaluate(`document.querySelectorAll('.todo-item').length`, &todoCount),
	)

	if err != nil {
		t.Fatalf("Browser automation failed: %v", err)
	}

	if todoCount != 1 {
		t.Errorf("Expected 1 todo after clearing completed, got: %d", todoCount)
	}

	t.Logf("Test passed! Clear completed functionality working correctly")
}

func TestLeftItemsText(t *testing.T) {
	// Start the development server
	server := devserver.NewServer("todo", "localhost:0")
	if err := server.Start(); err != nil {
		t.Fatalf("Failed to start dev server: %v", err)
	}
	defer server.Stop()

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Create chromedp context
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", false),
		chromedp.Flag("disable-gpu", false),
		chromedp.Flag("no-sandbox", true),
	)
	allocCtx, cancel := chromedp.NewExecAllocator(ctx, opts...)
	defer cancel()

	ctx, cancel = chromedp.NewContext(allocCtx)
	defer cancel()

	var leftItemsText string
	err := chromedp.Run(ctx,
		// Navigate to the application
		chromedp.Navigate(server.URL()),

		// Wait for the page to load
		chromedp.WaitVisible(`#new-todo-input`, chromedp.ByID),

		// Initially should show 0 items left
		chromedp.WaitVisible(`#stats-footer`, chromedp.ByID),
		chromedp.Text(`#stats-footer > div:first-child`, &leftItemsText, chromedp.ByQuery),
	)

	if err != nil {
		t.Fatalf("Browser automation failed: %v", err)
	}

	if leftItemsText != "0 items left" {
		t.Errorf("Expected '0 items left', got: '%s'", leftItemsText)
	}

	// Add some todos and test the counter
	err = chromedp.Run(ctx,
		// Add two todos
		chromedp.SendKeys(`#new-todo-input`, "Todo 1", chromedp.ByID),
		chromedp.Click(`#add-todo-btn`, chromedp.ByID),
		chromedp.Sleep(200*time.Millisecond),

		chromedp.SendKeys(`#new-todo-input`, "Todo 2", chromedp.ByID),
		chromedp.Click(`#add-todo-btn`, chromedp.ByID),
		chromedp.Sleep(200*time.Millisecond),

		// Check items left text
		chromedp.Text(`#stats-footer > div:first-child`, &leftItemsText, chromedp.ByQuery),
	)

	if err != nil {
		t.Fatalf("Browser automation failed: %v", err)
	}

	if leftItemsText != "2 items left" {
		t.Errorf("Expected '2 items left', got: '%s'", leftItemsText)
	}

	// Mark one as completed and check again
	err = chromedp.Run(ctx,
		// Mark first todo as completed
		chromedp.Click(`.todo-toggle`, chromedp.ByQuery),
		chromedp.Sleep(200*time.Millisecond),

		// Check items left text
		chromedp.Text(`#stats-footer > div:first-child`, &leftItemsText, chromedp.ByQuery),
	)

	if err != nil {
		t.Fatalf("Browser automation failed: %v", err)
	}

	if leftItemsText != "1 item left" {
		t.Errorf("Expected '1 item left', got: '%s'", leftItemsText)
	}

	t.Logf("Test passed! Left items text working correctly")
}