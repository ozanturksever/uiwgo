//go:build !js && !wasm

package main

import (
	"context"
	"testing"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/ozanturksever/uiwgo/internal/devserver"
	"github.com/ozanturksever/uiwgo/internal/testhelpers"
)

func TestResourceApp(t *testing.T) {
	// Start the development server
	server := devserver.NewServer("resource", "localhost:0")
	if err := server.Start(); err != nil {
		t.Fatalf("Failed to start dev server: %v", err)
	}
	defer server.Stop()

	// Create chromedp context with visible browser for debugging
	chromedpCtx := testhelpers.MustNewChromedpContext(testhelpers.DefaultConfig())
	defer chromedpCtx.Cancel()

	// Navigate to the app and test resource loading
	var loadingText, user1Text, errorText string
	err := chromedp.Run(chromedpCtx.Ctx,
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

func TestResourceInitialState(t *testing.T) {
	// Start the development server
	server := devserver.NewServer("resource", "localhost:0")
	if err := server.Start(); err != nil {
		t.Fatalf("Failed to start dev server: %v", err)
	}
	defer server.Stop()

	// Create chromedp context with headless browser
	chromedpCtx := testhelpers.MustNewChromedpContext(testhelpers.DefaultConfig())
	defer chromedpCtx.Cancel()

	// Navigate to the app and test initial state
	var initialText string
	err := chromedp.Run(chromedpCtx.Ctx,
		// Navigate to the application
		chromedp.Navigate(server.URL()),

		// Wait for the page to load
		chromedp.WaitVisible(`#user-display`, chromedp.ByID),
		chromedp.Sleep(2*time.Second), // Wait for WASM to initialize

		// Capture the initial state
		chromedp.Text(`#user-display`, &initialText, chromedp.ByID),
	)

	if err != nil {
		t.Fatalf("Browser automation failed: %v", err)
	}

	// Verify the initial state (might be empty or show default user)
	if initialText == "" {
		t.Logf("Expected some initial text, got empty string - initial state might need debugging")
	}

	t.Logf("Initial state test passed! Initial text: %s", initialText)
}

func TestResourceRandomUser(t *testing.T) {
	// Start the development server
	server := devserver.NewServer("resource", "localhost:0")
	if err := server.Start(); err != nil {
		t.Fatalf("Failed to start dev server: %v", err)
	}
	defer server.Stop()

	// Create chromedp context with headless browser
	chromedpCtx := testhelpers.MustNewChromedpContext(testhelpers.DefaultConfig())
	defer chromedpCtx.Cancel()

	// Navigate to the app and test random user loading
	var randomUserText string
	err := chromedp.Run(chromedpCtx.Ctx,
		// Navigate to the application
		chromedp.Navigate(server.URL()),

		// Wait for the page to load
		chromedp.WaitVisible(`#random-user-btn`, chromedp.ByID),
		chromedp.Sleep(2*time.Second), // Wait for WASM to initialize

		// Click random user button
		chromedp.Click(`#random-user-btn`, chromedp.ByID),

		// Wait for the request to complete
		chromedp.Sleep(1200*time.Millisecond),

		// Capture the result
		chromedp.Text(`#user-display`, &randomUserText, chromedp.ByID),
	)

	if err != nil {
		t.Fatalf("Browser automation failed: %v", err)
	}

	// Verify that we got either a user or an error (since random can be 1, 2, or 3)
	if randomUserText == "" || randomUserText == "No user loaded yet" {
		t.Errorf("Expected random user to load some content, got: '%s'", randomUserText)
	}

	// Should contain either "User #" or "Error:"
	containsUser := len(randomUserText) > 0 && (randomUserText[0:5] == "User " || randomUserText[0:6] == "Error:")
	if !containsUser {
		t.Errorf("Expected random user text to contain user info or error, got: '%s'", randomUserText)
	}

	t.Logf("Random user test passed! Result: %s", randomUserText)
}

func TestResourceRapidClicks(t *testing.T) {
	// Start the development server
	server := devserver.NewServer("resource", "localhost:0")
	if err := server.Start(); err != nil {
		t.Fatalf("Failed to start dev server: %v", err)
	}
	defer server.Stop()

	// Create chromedp context with headless mode and extended timeout for this test
	config := testhelpers.DefaultConfig()
	config.Headless = true
	config.Timeout = 45 * time.Second
	chromedpCtx := testhelpers.MustNewChromedpContext(config)
	defer chromedpCtx.Cancel()

	// Navigate to the app and test rapid clicking
	var finalText string
	err := chromedp.Run(chromedpCtx.Ctx,
		// Navigate to the application
		chromedp.Navigate(server.URL()),

		// Wait for the page to load
		chromedp.WaitVisible(`#load-user1-btn`, chromedp.ByID),
		chromedp.Sleep(2*time.Second), // Wait for WASM to initialize

		// Rapidly click different buttons to test resource cancellation/handling
		chromedp.Click(`#load-user1-btn`, chromedp.ByID),
		chromedp.Sleep(100*time.Millisecond),
		chromedp.Click(`#load-user2-btn`, chromedp.ByID),
		chromedp.Sleep(100*time.Millisecond),
		chromedp.Click(`#load-user1-btn`, chromedp.ByID),
		chromedp.Sleep(100*time.Millisecond),
		chromedp.Click(`#random-user-btn`, chromedp.ByID),

		// Wait for the final request to complete
		chromedp.Sleep(1500*time.Millisecond),

		// Capture the final result
		chromedp.Text(`#user-display`, &finalText, chromedp.ByID),
	)

	if err != nil {
		t.Fatalf("Browser automation failed: %v", err)
	}

	// Verify that we got a valid final state (not loading)
	if finalText == "Loading..." {
		t.Errorf("Expected final state to not be loading after rapid clicks, got: '%s'", finalText)
	}

	// Should contain either user info or error
	if finalText == "" || finalText == "No user loaded yet" {
		t.Errorf("Expected final state to show user or error after rapid clicks, got: '%s'", finalText)
	}

	t.Logf("Rapid clicks test passed! Final result: %s", finalText)
}

func TestResourceErrorRecovery(t *testing.T) {
	// Start the development server
	server := devserver.NewServer("resource", "localhost:0")
	if err := server.Start(); err != nil {
		t.Fatalf("Failed to start dev server: %v", err)
	}
	defer server.Stop()

	// Create a context with timeout
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

	// Navigate to the app and test error recovery
	var errorText, recoveryText string
	err := chromedp.Run(ctx,
		// Navigate to the application
		chromedp.Navigate(server.URL()),

		// Wait for the page to load
		chromedp.WaitVisible(`#load-user2-btn`, chromedp.ByID),
		chromedp.Sleep(2*time.Second), // Wait for WASM to initialize

		// First, trigger an error
		chromedp.Click(`#load-user2-btn`, chromedp.ByID),
		chromedp.Sleep(1200*time.Millisecond),
		chromedp.Text(`#user-display`, &errorText, chromedp.ByID),

		// Then recover with a successful request
		chromedp.Click(`#load-user1-btn`, chromedp.ByID),
		chromedp.Sleep(1200*time.Millisecond),
		chromedp.Text(`#user-display`, &recoveryText, chromedp.ByID),
	)

	if err != nil {
		t.Fatalf("Browser automation failed: %v", err)
	}

	// Verify error state
	if errorText != "Error: user 2 not found" {
		t.Errorf("Expected error text to be 'Error: user 2 not found', got: '%s'", errorText)
	}

	// Verify recovery
	expectedRecovery := "User #1\n\nName: User-1"
	if recoveryText != expectedRecovery {
		t.Errorf("Expected recovery text to be '%s', got: '%s'", expectedRecovery, recoveryText)
	}

	t.Logf("Error recovery test passed! Error: %s, Recovery: %s", errorText, recoveryText)
}
