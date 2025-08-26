//go:build !js && !wasm

package main

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/ozanturksever/uiwgo/internal/testhelpers"
	"github.com/ozanturksever/uiwgo/internal/devserver"
)

// TestNavigation_BrowserHistoryPopstate tests that the router correctly handles
// browser history navigation via popstate events.
func TestNavigation_BrowserHistoryPopstate(t *testing.T) {
	// Start the development server with router_test example
	server := devserver.NewServer("router_test", "localhost:0")
	if err := server.Start(); err != nil {
		t.Fatalf("Failed to start dev server: %v", err)
	}
	defer server.Stop()

	// Create chromedp context
	chromedpCtx := testhelpers.MustNewChromedpContext(testhelpers.DefaultConfig())
	defer chromedpCtx.Cancel()

	// Variable to capture the current location from JavaScript
	var currentLocation string

	err := chromedp.Run(chromedpCtx.Ctx,
		// Navigate to the application
		chromedp.Navigate(server.URL()),

		// Wait for the page to load and the router to initialize
		chromedp.WaitVisible("body", chromedp.ByQuery),
		chromedp.Sleep(1*time.Second), // Give time for WASM to initialize

		// Use JavaScript to push a new state and then go back to trigger popstate
		chromedp.Evaluate(`
			// Push a new state to history
			window.history.pushState({}, "", "/test-route");
			// Go back to trigger popstate event
			window.history.back();
		`, nil),

		// Wait for the popstate event to be processed
		chromedp.Sleep(500*time.Millisecond),

		// Read the current location from the router's global state
		// This assumes we expose the router's location state to JavaScript
		chromedp.Evaluate(`
			(() => {
				// Access the router's current location through a global variable
				// This will be implemented by exposing the router state to JS
				if (window.__router && window.__router.location) {
					return window.__router.location.pathname;
				}
				return "no-router-found";
			})();
		`, &currentLocation),
	)

	if err != nil {
		t.Fatalf("Browser automation failed: %v", err)
	}

	// Verify that the router's location was updated after the popstate event
	// The initial location should be "/" and after going back from "/test-route",
	// it should return to "/"
	if currentLocation != "/" {
		t.Errorf("Expected router location to be '/' after popstate, got: %s", currentLocation)
	}

	t.Logf("Test passed! Router location after popstate: %s", currentLocation)
}

// TestRouterInitialRenderMountsComponent tests that the router performs an initial render
// of the matched component when the router is created.
func TestRouterInitialRenderMountsComponent(t *testing.T) {
	// Start the development server with router_test example
	server := devserver.NewServer("router_test", "localhost:0")
	if err := server.Start(); err != nil {
		t.Fatalf("Failed to start dev server: %v", err)
	}
	defer server.Stop()

	// Create chromedp context
	chromedpCtx := testhelpers.MustNewChromedpContext(testhelpers.DefaultConfig())
	defer chromedpCtx.Cancel()

	// Variable to capture the content of the root element
	var rootContent string

	err := chromedp.Run(chromedpCtx.Ctx,
		// Navigate to the application
		chromedp.Navigate(server.URL()),

		// Wait for the page to load and the router to initialize
		chromedp.WaitVisible("body", chromedp.ByQuery),
		chromedp.Sleep(2*time.Second), // Give time for WASM to initialize and render

		// Check that the app element contains the rendered component content
		chromedp.InnerHTML("#app", &rootContent, chromedp.ByQuery),
	)

	if err != nil {
		t.Fatalf("Browser automation failed: %v", err)
	}

	// Verify that the root element contains the expected content from the "/" route component
	expectedContent := "Home Page"
	if !strings.Contains(rootContent, expectedContent) {
		t.Errorf("Expected root element to contain '%s', but got: %s", expectedContent, rootContent)
	}

	t.Logf("Test passed! Root element content: %s", rootContent)
}

// TestRouterUpdatesViewOnRouteChange tests that the router updates the view when the route changes
// programmatically via router.Navigate.
func TestRouterUpdatesViewOnRouteChange(t *testing.T) {
	// Start the development server with router_test example
	server := devserver.NewServer("router_test", "localhost:0")
	if err := server.Start(); err != nil {
		t.Fatalf("Failed to start dev server: %v", err)
	}
	defer server.Stop()

	// Create chromedp context with longer timeout for this test
	config := testhelpers.DefaultConfig()
	config.Timeout = 60 * time.Second
	chromedpCtx := testhelpers.MustNewChromedpContext(config)
	defer chromedpCtx.Cancel()

	// Variable to capture the content of the root element before and after navigation
	var initialContent string
	var navigatedContent string

	err := chromedp.Run(chromedpCtx.Ctx,
		// Navigate to the application
		chromedp.Navigate(server.URL()),

		// Wait for the page to load and the initial content to be rendered
		chromedp.WaitVisible(`#app:not(:empty)`, chromedp.ByQuery),

		// Check that the app element contains the initial rendered component content ("Home Page")
		chromedp.InnerHTML("#app", &initialContent, chromedp.ByQuery),

		// Verify initial content is correct
		chromedp.ActionFunc(func(ctx context.Context) error {
			expectedContent := "Home Page"
			if !strings.Contains(initialContent, expectedContent) {
				t.Errorf("Expected initial root element to contain '%s', but got: %s", expectedContent, initialContent)
			}
			return nil
		}),

		// Use JavaScript to call router.Navigate("/test-route")
		chromedp.Evaluate(`
			// Access the global router instance and navigate to /test-route
			if (window.__router) {
				window.__router.Navigate("/test-route");
			} else {
				throw new Error("Router not found on window");
			}
		`, nil),

		// Wait for the navigation to complete and the new content to be rendered
		chromedp.WaitVisible(`#app:not(:empty)`, chromedp.ByQuery),

		// Check the updated content of the root element
		chromedp.InnerHTML("#app", &navigatedContent, chromedp.ByQuery),
	)

	if err != nil {
		t.Fatalf("Browser automation failed: %v", err)
	}

	// Verify that the root element now contains the content from the "/test-route" component
	expectedContent := "Test Route Page"
	if !strings.Contains(navigatedContent, expectedContent) {
		t.Errorf("Expected root element to contain '%s' after navigation, but got: %s", expectedContent, navigatedContent)
	}

	t.Logf("Test passed! Initial content: %s, Navigated content: %s", initialContent, navigatedContent)
}
