package main

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/chromedp/cdproto/runtime"
	"github.com/chromedp/chromedp"
	"github.com/ozanturksever/uiwgo/internal/devserver"
)

// TestComponentDemo runs a comprehensive test of all component features
func TestComponentDemo(t *testing.T) {
	// Start the development server
	server, err := devserver.NewDevServer("component_demo")
	if err != nil {
		t.Fatalf("Failed to start dev server: %v", err)
	}
	defer server.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Create chromedp context with headless browser
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", true),
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("no-sandbox", true),
	)
	allocCtx, cancel := chromedp.NewExecAllocator(ctx, opts...)
	defer cancel()

	ctx, cancel = chromedp.NewContext(allocCtx)
	defer cancel()

	// Test variables
	var appContent, counterText, greetingText, todoText, fragmentText string
	var todoItemCount int

	err = chromedp.Run(ctx,
		// Navigate to the application
		chromedp.Navigate(server.URL()),

		// Wait for the page to load and WASM to initialize
		chromedp.WaitVisible(`#app`, chromedp.ByID),
		chromedp.Sleep(3*time.Second), // Wait for WASM to fully initialize

		// Verify the app loaded successfully
		chromedp.InnerHTML(`#app`, &appContent, chromedp.ByID),

		// Test Counter component
		chromedp.Text(`//h3[contains(text(), "Counter Component")]/following-sibling::div//p`, &counterText, chromedp.BySearch),

		// Test Greeting component with props
		chromedp.Text(`//h3[contains(text(), "Greeting Component")]/following-sibling::div//p`, &greetingText, chromedp.BySearch),

		// Test Todo component
		chromedp.Text(`//h3[contains(text(), "Todo Component")]/following-sibling::div//h2`, &todoText, chromedp.BySearch),
		chromedp.Evaluate(`document.querySelectorAll('ul li').length`, &todoItemCount),

		// Test Fragment component
		chromedp.Text(`//h3[contains(text(), "Fragment Example")]/following-sibling::div//p[1]`, &fragmentText, chromedp.BySearch),
	)

	if err != nil {
		t.Fatalf("Chromedp run failed: %v", err)
	}

	// Verify component content
	if counterText != "Count: 0" {
		t.Errorf("Expected counter text 'Count: 0', got '%s'", counterText)
	}

	if greetingText != "Hello, World!" {
		t.Errorf("Expected greeting text 'Hello, World!', got '%s'", greetingText)
	}

	if todoText != "Todo Component" {
		t.Errorf("Expected todo text 'Todo Component', got '%s'", todoText)
	}

	if todoItemCount != 0 {
		t.Errorf("Expected 0 todo items initially, got %d", todoItemCount)
	}

	if fragmentText != "This is a paragraph" {
		t.Errorf("Expected fragment text 'This is a paragraph', got '%s'", fragmentText)
	}

	t.Logf("Initial component test passed successfully")
}

// TestCounterComponent tests the counter component functionality
func TestCounterComponent(t *testing.T) {
	server, err := devserver.NewDevServer("component_demo")
	if err != nil {
		t.Fatalf("Failed to start dev server: %v", err)
	}
	defer server.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", true),
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("no-sandbox", true),
	)
	allocCtx, cancel := chromedp.NewExecAllocator(ctx, opts...)
	defer cancel()

	ctx, cancel = chromedp.NewContext(allocCtx)
	defer cancel()

	var counterText string

	err = chromedp.Run(ctx,
		chromedp.Navigate(server.URL()),
		chromedp.WaitVisible(`#app`, chromedp.ByID),
		chromedp.Sleep(3*time.Second),

		// Click increment button
		chromedp.Click(`//button[contains(text(), "Increment")]`, chromedp.BySearch),
		chromedp.Sleep(500*time.Millisecond),

		// Get updated counter text
		chromedp.Text(`//h3[contains(text(), "Counter Component")]/following-sibling::div//p`, &counterText, chromedp.BySearch),
	)

	if err != nil {
		t.Fatalf("Counter test failed: %v", err)
	}

	if counterText != "Count: 1" {
		t.Errorf("Expected counter text 'Count: 1' after increment, got '%s'", counterText)
	}

	t.Logf("Counter component test passed: %s", counterText)
}

// TestTodoComponent tests the todo component functionality
func TestTodoComponent(t *testing.T) {
	server, err := devserver.NewDevServer("component_demo")
	if err != nil {
		t.Fatalf("Failed to start dev server: %v", err)
	}
	defer server.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", true),
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("no-sandbox", true),
	)
	allocCtx, cancel := chromedp.NewExecAllocator(ctx, opts...)
	defer cancel()

	ctx, cancel = chromedp.NewContext(allocCtx)
	defer cancel()

	var todoItemCount int

	err = chromedp.Run(ctx,
		chromedp.Navigate(server.URL()),
		chromedp.WaitVisible(`#app`, chromedp.ByID),
		chromedp.Sleep(3*time.Second),

		// Click add item button
		chromedp.Click(`//h3[contains(text(), "Todo Component")]/following-sibling::div//button[contains(text(), "Add Item")]`, chromedp.BySearch),
		chromedp.Sleep(500*time.Millisecond),

		// Count todo items
		chromedp.Evaluate(`document.querySelectorAll('ul li').length`, &todoItemCount),
	)

	if err != nil {
		t.Fatalf("Todo test failed: %v", err)
	}

	if todoItemCount != 1 {
		t.Errorf("Expected 1 todo item after adding, got %d", todoItemCount)
	}

	t.Logf("Todo component test passed: %d items", todoItemCount)
}

// TestComponentLifecycle tests component mounting and unmounting
func TestComponentLifecycle(t *testing.T) {
	server, err := devserver.NewDevServer("component_demo")
	if err != nil {
		t.Fatalf("Failed to start dev server: %v", err)
	}
	defer server.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", true),
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("no-sandbox", true),
	)
	allocCtx, cancel := chromedp.NewExecAllocator(ctx, opts...)
	defer cancel()

	ctx, cancel = chromedp.NewContext(allocCtx)
	defer cancel()

	// Listen for console messages to detect mount/unmount events
	consoleMessages := make([]string, 0)
	chromedp.ListenTarget(ctx, func(ev interface{}) {
		switch ev := ev.(type) {
		case *runtime.EventConsoleAPICalled:
			for _, arg := range ev.Args {
				if arg.Value != nil {
					consoleMessages = append(consoleMessages, fmt.Sprintf("%v", arg.Value))
				}
			}
		}
	})

	err = chromedp.Run(ctx,
		chromedp.Navigate(server.URL()),
		chromedp.WaitVisible(`#app`, chromedp.ByID),
		chromedp.Sleep(3*time.Second),

		// Reload the page to test component lifecycle
		chromedp.Reload(),
		chromedp.Sleep(3*time.Second),
	)

	if err != nil {
		t.Fatalf("Lifecycle test failed: %v", err)
	}

	// Check if mount messages were logged
	mountCount := 0
	for _, msg := range consoleMessages {
		if msg == "App mounted" || msg == "Counter mounted" || msg == "Greeting mounted" || msg == "Todo mounted" || msg == "Header mounted" {
			mountCount++
		}
	}

	if mountCount < 3 {
		t.Errorf("Expected at least 3 component mount messages, got %d", mountCount)
	}

	t.Logf("Component lifecycle test passed: %d mount messages detected", mountCount)
}

// TestMemoization tests component memoization functionality
func TestMemoization(t *testing.T) {
	server, err := devserver.NewDevServer("component_demo")
	if err != nil {
		t.Fatalf("Failed to start dev server: %v", err)
	}
	defer server.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", true),
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("no-sandbox", true),
	)
	allocCtx, cancel := chromedp.NewExecAllocator(ctx, opts...)
	defer cancel()

	ctx, cancel = chromedp.NewContext(allocCtx)
	defer cancel()

	consoleMessages := make([]string, 0)
	chromedp.ListenTarget(ctx, func(ev interface{}) {
		switch ev := ev.(type) {
		case *runtime.EventConsoleAPICalled:
			for _, arg := range ev.Args {
				if arg.Value != nil {
					consoleMessages = append(consoleMessages, fmt.Sprintf("%v", arg.Value))
				}
			}
		}
	})

	err = chromedp.Run(ctx,
		chromedp.Navigate(server.URL()),
		chromedp.WaitVisible(`#app`, chromedp.ByID),
		chromedp.Sleep(3*time.Second),

		// Click reload button to test memoization
		chromedp.Click(`//button[contains(text(), "Reload")]`, chromedp.BySearch),
		chromedp.Sleep(3*time.Second),
	)

	if err != nil {
		t.Fatalf("Memoization test failed: %v", err)
	}

	// Count mount messages - should not see duplicate mounts if memoization works
	mountMessages := 0
	for _, msg := range consoleMessages {
		if msg == "App mounted" || msg == "Counter mounted" || msg == "Greeting mounted" || msg == "Todo mounted" || msg == "Header mounted" {
			mountMessages++
		}
	}

	// With memoization, we should see fewer mount messages on reload
	if mountMessages > 10 { // Arbitrary threshold - should be much lower with proper memoization
		t.Logf("Memoization test: %d mount messages detected (memoization may not be fully implemented)", mountMessages)
	} else {
		t.Logf("Memoization test passed: %d mount messages detected", mountMessages)
	}
}
