package main

import (
	"context"
	"testing"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/ozanturksever/uiwgo/internal/devserver"
)

// Helper function
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func TestRouterExample(t *testing.T) {
	// Create and start the development server
	server := devserver.NewServer("router", "localhost:0")
	if err := server.Start(); err != nil {
		t.Fatalf("Failed to start dev server: %v", err)
	}
	defer server.Stop()

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Create Chrome options for headless browser
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", true),
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("no-sandbox", true),
		chromedp.Flag("disable-dev-shm-usage", true),
	)

	// Create a new allocator context with the options
	allocCtx, cancel := chromedp.NewExecAllocator(ctx, opts...)
	defer cancel()

	// Create a new browser context
	browserCtx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()

	// Navigate to the router example
	var title string
	err := chromedp.Run(browserCtx,
		chromedp.Navigate(server.URL()),
		chromedp.Sleep(2*time.Second), // Wait for WASM to initialize
		chromedp.WaitVisible("div.container", chromedp.ByQuery),
		chromedp.Title(&title),
	)
	if err != nil {
		t.Fatalf("Failed to load router example: %v", err)
	}

	if title != "Router Example - oiwgo" {
		t.Errorf("Expected title 'Router Example - oiwgo', got '%s'", title)
	}
}

func TestRouterNavigation(t *testing.T) {
	// Create and start the development server
	server := devserver.NewServer("router", "localhost:0")
	if err := server.Start(); err != nil {
		t.Fatalf("Failed to start dev server: %v", err)
	}
	defer server.Stop()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Create Chrome options for headless browser
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", true),
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("no-sandbox", true),
		chromedp.Flag("disable-dev-shm-usage", true),
	)

	// Create a new allocator context with the options
	allocCtx, cancel := chromedp.NewExecAllocator(ctx, opts...)
	defer cancel()

	// Create a new browser context
	browserCtx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()

	var initialPath, aboutPath, consoleOutput, linkHTML, directCallPath string
	err := chromedp.Run(browserCtx,
		chromedp.Navigate(server.URL()),
		chromedp.Sleep(4*time.Second), // Wait for WASM to initialize
		chromedp.WaitVisible("div.container", chromedp.ByQuery),
		
		// Get initial breadcrumb
		chromedp.TextContent("div.breadcrumbs code", &initialPath, chromedp.ByQuery),
		
		// Check what the About link looks like
		chromedp.OuterHTML("nav a[href='/about']", &linkHTML, chromedp.ByQuery),
		
		// Check if uiwgo_navigate function exists
		chromedp.Evaluate(`typeof uiwgo_navigate`, &consoleOutput),
		
		// Try calling uiwgo_navigate directly instead of clicking
		chromedp.Evaluate(`uiwgo_navigate('/about', false); 'called'`, nil),
		chromedp.Sleep(1*time.Second), // Wait for direct call
		
		// Check breadcrumb after direct call
		chromedp.TextContent("div.breadcrumbs code", &directCallPath, chromedp.ByQuery),
		
		// Try to click About link
		chromedp.Click("nav a[href='/about']", chromedp.ByQuery),
		chromedp.Sleep(2*time.Second), // Wait for navigation
		
		// Check if breadcrumb changed
		chromedp.TextContent("div.breadcrumbs code", &aboutPath, chromedp.ByQuery),
	)
	
	if err != nil {
		t.Fatalf("Router navigation test failed: %v", err)
	}
	
	t.Logf("Initial path: %s", initialPath)
	t.Logf("Direct call path: %s", directCallPath)
	t.Logf("About path: %s", aboutPath)
	t.Logf("Link HTML: %s", linkHTML)
	t.Logf("Console: %s", consoleOutput)
	
	// Check if navigation worked
	if initialPath != "/" {
		t.Errorf("Expected initial path '/', got '%s'", initialPath)
	}
	
	if aboutPath != "/about" {
		t.Errorf("Expected about path '/about', got '%s' - navigation may not be working", aboutPath)
	}
}

func TestRouterBreadcrumbs(t *testing.T) {
	// Create and start the development server
	server := devserver.NewServer("router", "localhost:0")
	if err := server.Start(); err != nil {
		t.Fatalf("Failed to start dev server: %v", err)
	}
	defer server.Stop()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Create Chrome options for headless browser
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", true),
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("no-sandbox", true),
		chromedp.Flag("disable-dev-shm-usage", true),
	)

	// Create a new allocator context with the options
	allocCtx, cancel := chromedp.NewExecAllocator(ctx, opts...)
	defer cancel()

	// Create a new browser context
	browserCtx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()

	var breadcrumbText string
	err := chromedp.Run(browserCtx,
		chromedp.Navigate(server.URL()),
		chromedp.Sleep(4*time.Second), // Wait for WASM to initialize
		chromedp.WaitVisible("div.container", chromedp.ByQuery),
		
		// Navigate to users page and check breadcrumb
		chromedp.Click("nav a[href='/users']", chromedp.ByQuery),
		chromedp.Sleep(1*time.Second),
		chromedp.WaitVisible("div.breadcrumbs", chromedp.ByQuery),
		chromedp.TextContent("div.breadcrumbs code", &breadcrumbText, chromedp.ByQuery),
	)
	
	if err != nil {
		t.Fatalf("Breadcrumb test failed: %v", err)
	}
	
	if breadcrumbText != "/users" {
		t.Errorf("Expected breadcrumb '/users', got '%s'", breadcrumbText)
	}
}

func TestRouter404Page(t *testing.T) {
	// Create and start the development server
	server := devserver.NewServer("router", "localhost:0")
	if err := server.Start(); err != nil {
		t.Fatalf("Failed to start dev server: %v", err)
	}
	defer server.Stop()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Create Chrome options for headless browser
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", true),
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("no-sandbox", true),
		chromedp.Flag("disable-dev-shm-usage", true),
	)

	// Create a new allocator context with the options
	allocCtx, cancel := chromedp.NewExecAllocator(ctx, opts...)
	defer cancel()

	// Create a new browser context
	browserCtx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()

	var pageText string
	err := chromedp.Run(browserCtx,
		chromedp.Navigate(server.URL() + "/nonexistent-page"),
		chromedp.Sleep(4*time.Second), // Wait for WASM to initialize
		chromedp.WaitVisible("div.container", chromedp.ByQuery),
		chromedp.WaitVisible("div.error", chromedp.ByQuery),
		chromedp.TextContent("div.error h3", &pageText, chromedp.ByQuery),
	)
	
	if err != nil {
		t.Fatalf("404 page test failed: %v", err)
	}
	
	if pageText != "404 - Page Not Found" {
		t.Errorf("Expected '404 - Page Not Found', got '%s'", pageText)
	}
}

func TestRouterParameterExtraction(t *testing.T) {
	// Create and start the development server
	server := devserver.NewServer("router", "localhost:0")
	if err := server.Start(); err != nil {
		t.Fatalf("Failed to start dev server: %v", err)
	}
	defer server.Stop()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Create Chrome options for headless browser
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", true),
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("no-sandbox", true),
		chromedp.Flag("disable-dev-shm-usage", true),
	)

	// Create a new allocator context with the options
	allocCtx, cancel := chromedp.NewExecAllocator(ctx, opts...)
	defer cancel()

	// Create a new browser context
	browserCtx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()

	var userInfo string
	err := chromedp.Run(browserCtx,
		chromedp.Navigate(server.URL() + "/users/1"),
		chromedp.Sleep(4*time.Second), // Wait for WASM to initialize
		chromedp.WaitVisible("div.container", chromedp.ByQuery),
		chromedp.WaitVisible("div.card", chromedp.ByQuery),
		chromedp.TextContent("div.card p:first-of-type", &userInfo, chromedp.ByQuery),
	)
	
	if err != nil {
		t.Fatalf("Parameter extraction test failed: %v", err)
	}
	
	// Check if the user ID is displayed correctly
	if userInfo != "ID: 1" {
		t.Errorf("Expected 'ID: 1', got '%s'", userInfo)
	}
}

func TestRouterBackButton(t *testing.T) {
	// Create and start the development server
	server := devserver.NewServer("router", "localhost:0")
	if err := server.Start(); err != nil {
		t.Fatalf("Failed to start dev server: %v", err)
	}
	defer server.Stop()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Create Chrome options for headless browser
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", true),
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("no-sandbox", true),
		chromedp.Flag("disable-dev-shm-usage", true),
	)

	// Create a new allocator context with the options
	allocCtx, cancel := chromedp.NewExecAllocator(ctx, opts...)
	defer cancel()

	// Create a new browser context
	browserCtx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()

	var currentPath string
	err := chromedp.Run(browserCtx,
		chromedp.Navigate(server.URL()),
		chromedp.Sleep(4*time.Second), // Wait for WASM to initialize
		chromedp.WaitVisible("div.container", chromedp.ByQuery),
		
		// Navigate to about page
		chromedp.Click("nav a[href='/about']", chromedp.ByQuery),
		chromedp.WaitVisible("h3", chromedp.ByQuery),
		
		// Navigate to users page
		chromedp.Click("nav a[href='/users']", chromedp.ByQuery),
		chromedp.WaitVisible("div.user-list", chromedp.ByQuery),
		
		// Use browser back button
		chromedp.NavigateBack(),
		chromedp.WaitVisible("h3", chromedp.ByQuery),
		
		// Check current path in breadcrumbs
		chromedp.TextContent("div.breadcrumbs code", &currentPath, chromedp.ByQuery),
	)
	
	if err != nil {
		t.Fatalf("Back button test failed: %v", err)
	}
	
	if currentPath != "/about" {
		t.Errorf("Expected to be back on '/about', got '%s'", currentPath)
	}
}

func TestRouterLoadingAndContent(t *testing.T) {
	// Create and start the development server
	server := devserver.NewServer("router", "localhost:0")
	if err := server.Start(); err != nil {
		t.Fatalf("Failed to start dev server: %v", err)
	}
	defer server.Stop()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Create Chrome options for headless browser
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", true),
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("no-sandbox", true),
		chromedp.Flag("disable-dev-shm-usage", true),
	)

	// Create a new allocator context with the options
	allocCtx, cancel := chromedp.NewExecAllocator(ctx, opts...)
	defer cancel()

	// Create a new browser context
	browserCtx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()

	err := chromedp.Run(browserCtx,
		chromedp.Navigate(server.URL()),
		chromedp.Sleep(2*time.Second), // Wait for WASM to initialize
		
		// Wait for the app to load
		chromedp.WaitVisible("div.container", chromedp.ByQuery),
		
		// Verify content is loaded
		chromedp.WaitVisible("h1", chromedp.ByQuery),
		chromedp.WaitVisible("nav", chromedp.ByQuery),
		chromedp.WaitVisible("div.main", chromedp.ByQuery),
	)
	
	if err != nil {
		t.Fatalf("Loading and content test failed: %v", err)
	}
}