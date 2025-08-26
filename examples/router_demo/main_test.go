//go:build !js && !wasm

package main

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/chromedp/cdproto/runtime"
	"github.com/ozanturksever/uiwgo/internal/devserver"
)

// TestRouterDemo_HomePageRender tests that the home page renders correctly
func TestRouterDemo_HomePageRender(t *testing.T) {
	server := devserver.NewServer("router_demo", "localhost:0")
	if err := server.Start(); err != nil {
		t.Fatalf("Failed to start dev server: %v", err)
	}
	defer server.Stop()

	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
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

	var pageTitle, homeContent string

	err := chromedp.Run(ctx,
		chromedp.Navigate(server.URL()),
    chromedp.WaitVisible("#root", chromedp.ByQuery),
    chromedp.WaitNotPresent(".loading-indicator", chromedp.ByQuery),
		chromedp.Sleep(2*time.Second), // Give time for WASM to initialize
		chromedp.Title(&pageTitle),
		chromedp.Text("h1", &homeContent, chromedp.ByQuery),
	)

	if err != nil {
		t.Fatalf("Browser automation failed: %v", err)
	}

	if pageTitle != "UIWGo Router Demo" {
		t.Errorf("Expected page title 'UIWGo Router Demo', got: %s", pageTitle)
	}

	if !strings.Contains(homeContent, "Router Demo - Home") {
		t.Errorf("Expected home page content to contain 'Router Demo - Home', got: %s", homeContent)
	}

	t.Logf("Test passed! Home page rendered correctly with title: %s", pageTitle)
}

// TestRouterDemo_StaticRouteNavigation tests navigation to static routes
func TestRouterDemo_StaticRouteNavigation(t *testing.T) {
	server := devserver.NewServer("router_demo", "localhost:0")
	if err := server.Start(); err != nil {
		t.Fatalf("Failed to start dev server: %v", err)
	}
	defer server.Stop()

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

	var aboutContent, usersContent string

	err := chromedp.Run(ctx,
		chromedp.Navigate(server.URL()),
		chromedp.WaitVisible("body", chromedp.ByQuery),
		chromedp.Sleep(2*time.Second),
		
		// Navigate to About page
		chromedp.Click(`a[href="/about"]`, chromedp.ByQuery),
		chromedp.Sleep(500*time.Millisecond),
		chromedp.Text("h1", &aboutContent, chromedp.ByQuery),
		
		// Navigate to Users page
		chromedp.Click(`a[href="/users"]`, chromedp.ByQuery),
		chromedp.Sleep(500*time.Millisecond),
		chromedp.Text("h1", &usersContent, chromedp.ByQuery),
	)

	if err != nil {
		t.Fatalf("Browser automation failed: %v", err)
	}

	if !strings.Contains(aboutContent, "About UIWGo Router") {
		t.Errorf("Expected about page content to contain 'About UIWGo Router', got: %s", aboutContent)
	}

	if !strings.Contains(usersContent, "Users List") {
		t.Errorf("Expected users page content to contain 'Users List', got: %s", usersContent)
	}

	t.Logf("Test passed! Static route navigation works correctly")
}

// TestRouterDemo_DynamicRouteParameters tests dynamic route parameters
func TestRouterDemo_DynamicRouteParameters(t *testing.T) {
	server := devserver.NewServer("router_demo", "localhost:0")
	if err := server.Start(); err != nil {
		t.Fatalf("Failed to start dev server: %v", err)
	}
	defer server.Stop()

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

	var userProfileContent string

	err := chromedp.Run(ctx,
		chromedp.Navigate(server.URL()),
		chromedp.WaitVisible("body", chromedp.ByQuery),
		chromedp.Sleep(2*time.Second),
		
		// Navigate to user profile with ID 123
		chromedp.Click(`a[href="/users/123"]`, chromedp.ByQuery),
		chromedp.Sleep(500*time.Millisecond),
		chromedp.Text("#app", &userProfileContent, chromedp.ByQuery),
	)

	if err != nil {
		t.Fatalf("Browser automation failed: %v", err)
	}

	if !strings.Contains(userProfileContent, "User ID: 123") {
		t.Errorf("Expected user profile to contain 'User ID: 123', got: %s", userProfileContent)
	}

	if !strings.Contains(userProfileContent, "User Profile") {
		t.Errorf("Expected user profile to contain 'User Profile', got: %s", userProfileContent)
	}

	t.Logf("Test passed! Dynamic route parameters work correctly")
}

// TestRouterDemo_WildcardRoutes tests wildcard route matching
func TestRouterDemo_WildcardRoutes(t *testing.T) {
	server := devserver.NewServer("router_demo", "localhost:0")
	if err := server.Start(); err != nil {
		t.Fatalf("Failed to start dev server: %v", err)
	}
	defer server.Stop()

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

	var fileBrowserContent string

	err := chromedp.Run(ctx,
		chromedp.Navigate(server.URL()),
		chromedp.WaitVisible("body", chromedp.ByQuery),
		chromedp.Sleep(2*time.Second),
		
		// Navigate to file browser with wildcard path
		chromedp.Click(`a[href="/files/docs/readme.txt"]`, chromedp.ByQuery),
		chromedp.Sleep(500*time.Millisecond),
		chromedp.Text("#app", &fileBrowserContent, chromedp.ByQuery),
	)

	if err != nil {
		t.Fatalf("Browser automation failed: %v", err)
	}

	if !strings.Contains(fileBrowserContent, "docs/readme.txt") {
		t.Errorf("Expected file browser to contain 'docs/readme.txt', got: %s", fileBrowserContent)
	}

	if !strings.Contains(fileBrowserContent, "File Browser") {
		t.Errorf("Expected file browser to contain 'File Browser', got: %s", fileBrowserContent)
	}

	t.Logf("Test passed! Wildcard routes work correctly")
}

// TestRouterDemo_NestedRoutes tests nested route functionality
func TestRouterDemo_NestedRoutes(t *testing.T) {
	server := devserver.NewServer("router_demo", "localhost:0")
	if err := server.Start(); err != nil {
		t.Fatalf("Failed to start dev server: %v", err)
	}
	defer server.Stop()

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

	var adminContent, adminSettingsContent string

	err := chromedp.Run(ctx,
		chromedp.Navigate(server.URL()),
		chromedp.WaitVisible("body", chromedp.ByQuery),
		chromedp.Sleep(2*time.Second),
		
		// Navigate to admin panel
		chromedp.Click(`a[href="/admin"]`, chromedp.ByQuery),
		chromedp.Sleep(500*time.Millisecond),
		chromedp.Text("#app", &adminContent, chromedp.ByQuery),
		
		// Navigate to admin settings (nested route)
		chromedp.Click(`a[href="/admin/settings"]`, chromedp.ByQuery),
		chromedp.Sleep(500*time.Millisecond),
		chromedp.Text("#app", &adminSettingsContent, chromedp.ByQuery),
	)

	if err != nil {
		t.Fatalf("Browser automation failed: %v", err)
	}

	if !strings.Contains(adminContent, "Admin Panel") {
		t.Errorf("Expected admin page to contain 'Admin Panel', got: %s", adminContent)
	}

	if !strings.Contains(adminContent, "Admin Dashboard") {
		t.Errorf("Expected admin page to contain 'Admin Dashboard', got: %s", adminContent)
	}

	if !strings.Contains(adminSettingsContent, "Admin Settings") {
		t.Errorf("Expected admin settings to contain 'Admin Settings', got: %s", adminSettingsContent)
	}

	// Verify that the layout is still present in nested route
	if !strings.Contains(adminSettingsContent, "Admin Panel") {
		t.Errorf("Expected admin settings to still contain layout 'Admin Panel', got: %s", adminSettingsContent)
	}

	t.Logf("Test passed! Nested routes work correctly")
}

// TestRouterDemo_NotFoundRoute tests the 404 catch-all route
func TestRouterDemo_NotFoundRoute(t *testing.T) {
	server := devserver.NewServer("router_demo", "localhost:0")
	if err := server.Start(); err != nil {
		t.Fatalf("Failed to start dev server: %v", err)
	}
	defer server.Stop()

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

	// Enable console logging
	listenCtx, listenCancel := context.WithCancel(ctx)
	defer listenCancel()
	chromedp.ListenTarget(listenCtx, func(ev interface{}) {
		if consoleEv, ok := ev.(*runtime.EventConsoleAPICalled); ok {
			for _, arg := range consoleEv.Args {
				t.Logf("Console: %s", arg.Value)
			}
		} else if exceptionEv, ok := ev.(*runtime.EventExceptionThrown); ok {
			t.Logf("Exception: %v", exceptionEv.ExceptionDetails.Exception.Description)
		}
	})

	var notFoundContent string

	err := chromedp.Run(ctx,
		// Load the home page to initialize WASM
		chromedp.Navigate(server.URL()),
		chromedp.WaitVisible("body", chromedp.ByQuery),
		chromedp.Sleep(2*time.Second),
		
		// Debug: Check what's available in window
		chromedp.Evaluate(`console.log('Window keys:', Object.keys(window)); console.log('Router:', window.__router);`, nil),
		
		// Poll for router availability with more detailed check
		chromedp.Poll(`(function() {
			console.log('Polling for router');
			console.log('__router exists:', !!window.__router);
			console.log('Type of __router:', typeof window.__router);
			if (window.__router) {
				console.log('Router found:', window.__router);
				console.log('Has Navigate:', 'Navigate' in window.__router);
				console.log('Navigate type:', typeof window.__router.Navigate);
				console.log('Navigate is function:', typeof window.__router.Navigate === 'function');
				if (window.__router.location) {
					console.log('Location pathname:', window.__router.location.pathname);
				}
				if (window.__router.Navigate) {
					console.log('Navigate function found');
					return true;
				} else {
					console.log('Navigate function missing');
					return false;
				}
			} else {
				console.log('Router not found');
				return false;
			}
		})()`, nil, chromedp.WithPollingTimeout(15*time.Second)),
		
		// Navigate to non-existent route using router
		chromedp.Evaluate(`window.__router.Navigate('/non-existent-route');`, nil),
		chromedp.Sleep(500*time.Millisecond),
		chromedp.Text("#app", &notFoundContent, chromedp.ByQuery),
	)

	if err != nil {
		t.Fatalf("Browser automation failed: %v", err)
	}

	if !strings.Contains(notFoundContent, "404 - Page Not Found") {
		t.Errorf("Expected 404 page to contain '404 - Page Not Found', got: %s", notFoundContent)
	}

	if !strings.Contains(notFoundContent, "/non-existent-route") {
		t.Errorf("Expected 404 page to show requested path '/non-existent-route', got: %s", notFoundContent)
	}

	t.Logf("Test passed! 404 catch-all route works correctly")
}

// TestRouterDemo_BrowserHistoryNavigation tests browser back/forward navigation
func TestRouterDemo_BrowserHistoryNavigation(t *testing.T) {
	server := devserver.NewServer("router_demo", "localhost:0")
	if err := server.Start(); err != nil {
		t.Fatalf("Failed to start dev server: %v", err)
	}
	defer server.Stop()

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

	var homeContent, aboutContent, backToHomeContent string

	err := chromedp.Run(ctx,
		chromedp.Navigate(server.URL()),
		chromedp.WaitVisible("body", chromedp.ByQuery),
		chromedp.Sleep(2*time.Second),
		chromedp.Text("h1", &homeContent, chromedp.ByQuery),
		
		// Navigate to About page
		chromedp.Click(`a[href="/about"]`, chromedp.ByQuery),
		chromedp.Sleep(500*time.Millisecond),
		chromedp.Text("h1", &aboutContent, chromedp.ByQuery),
		
		// Use browser back button
		chromedp.NavigateBack(),
		chromedp.Sleep(500*time.Millisecond),
		chromedp.Text("h1", &backToHomeContent, chromedp.ByQuery),
	)

	if err != nil {
		t.Fatalf("Browser automation failed: %v", err)
	}

	if !strings.Contains(homeContent, "Router Demo - Home") {
		t.Errorf("Expected initial home content, got: %s", homeContent)
	}

	if !strings.Contains(aboutContent, "About UIWGo Router") {
		t.Errorf("Expected about content after navigation, got: %s", aboutContent)
	}

	if !strings.Contains(backToHomeContent, "Router Demo - Home") {
		t.Errorf("Expected home content after browser back, got: %s", backToHomeContent)
	}

	t.Logf("Test passed! Browser history navigation works correctly")
}