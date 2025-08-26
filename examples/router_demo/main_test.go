//go:build !js && !wasm

package main

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/chromedp/cdproto/runtime"
	"github.com/chromedp/chromedp"
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

// TestRouterDemo_StaticRouteNavigation tests navigation to different static routes
func TestRouterDemo_StaticRouteNavigation(t *testing.T) {
	server := devserver.NewServer("router_demo", "localhost:0")
	if err := server.Start(); err != nil {
		t.Fatalf("Failed to start dev server: %v", err)
	}
	defer server.Stop()

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
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
		chromedp.WaitVisible("#root", chromedp.ByQuery),
		chromedp.WaitNotPresent(".loading-indicator", chromedp.ByQuery),
		chromedp.Sleep(2*time.Second),

		// Navigate to About
		chromedp.Click(`a[href="/about"]`, chromedp.ByQuery),
		chromedp.WaitVisible("h1", chromedp.ByQuery),
		chromedp.Text("h1", &aboutContent, chromedp.ByQuery),

		// Navigate to Users
		chromedp.Click(`a[href="/users"]`, chromedp.ByQuery),
		chromedp.WaitVisible("h1", chromedp.ByQuery),
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

// TestRouterDemo_DynamicRouteParameters tests dynamic segment routing
func TestRouterDemo_DynamicRouteParameters(t *testing.T) {
	server := devserver.NewServer("router_demo", "localhost:0")
	if err := server.Start(); err != nil {
		t.Fatalf("Failed to start dev server: %v", err)
	}
	defer server.Stop()

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
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

	var userContent, profileContent string

	err := chromedp.Run(ctx,
		chromedp.Navigate(server.URL()),
		chromedp.WaitVisible("#root", chromedp.ByQuery),
		chromedp.WaitNotPresent(".loading-indicator", chromedp.ByQuery),
		chromedp.Sleep(2*time.Second),

		// Navigate to user 123 via SPA pushState
		chromedp.Evaluate(`(function(){ history.pushState({}, '', '/users/123'); window.dispatchEvent(new PopStateEvent('popstate')); })()`, nil),
		chromedp.ActionFunc(func(ctx context.Context) error {
			deadline := time.Now().Add(8 * time.Second)
			var pathname string
			for time.Now().Before(deadline) {
				if err := chromedp.Evaluate(`window.location.pathname`, &pathname).Do(ctx); err == nil && pathname == "/users/123" {
					return nil
				}
				time.Sleep(100 * time.Millisecond)
			}
			return fmt.Errorf("timed out waiting for pathname '/users/123'")
		}),
		chromedp.Text("#app", &userContent, chromedp.ByQuery),

		// Navigate to user 456 extended profile via SPA history
		chromedp.Evaluate(`(function(){ history.pushState({}, '', '/users/456/profile'); window.dispatchEvent(new PopStateEvent('popstate')); })()`, nil),
		chromedp.ActionFunc(func(ctx context.Context) error {
			deadline := time.Now().Add(8 * time.Second)
			var pathname string
			for time.Now().Before(deadline) {
				if err := chromedp.Evaluate(`window.location.pathname`, &pathname).Do(ctx); err == nil && pathname == "/users/456/profile" {
					return nil
				}
				time.Sleep(100 * time.Millisecond)
			}
			return fmt.Errorf("timed out waiting for pathname '/users/456/profile'")
		}),
		chromedp.Text("#app", &profileContent, chromedp.ByQuery),
	)

	if err != nil {
		t.Fatalf("Browser automation failed: %v", err)
	}

	if !strings.Contains(userContent, "User Profile") || !strings.Contains(userContent, "User ID: 123") {
		t.Errorf("Expected user profile view for ID 123, got: %s", userContent)
	}

	if !strings.Contains(profileContent, "Extended User Profile") || !strings.Contains(profileContent, "User ID: 456") {
		t.Errorf("Expected extended profile view for ID 456, got: %s", profileContent)
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

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
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

	var fileContent string

	err := chromedp.Run(ctx,
		chromedp.Navigate(server.URL()),
		chromedp.WaitVisible("#root", chromedp.ByQuery),
		chromedp.WaitNotPresent(".loading-indicator", chromedp.ByQuery),
		chromedp.Sleep(2*time.Second),

		// Navigate to a file path via SPA pushState
		chromedp.Evaluate(`(function(){ history.pushState({}, '', '/files/docs/readme.txt'); window.dispatchEvent(new PopStateEvent('popstate')); })()`, nil),
		chromedp.ActionFunc(func(ctx context.Context) error {
			deadline := time.Now().Add(8 * time.Second)
			var pathname string
			for time.Now().Before(deadline) {
				if err := chromedp.Evaluate(`window.location.pathname`, &pathname).Do(ctx); err == nil && pathname == "/files/docs/readme.txt" {
					return nil
				}
				time.Sleep(100 * time.Millisecond)
			}
			return fmt.Errorf("timed out waiting for pathname '/files/docs/readme.txt'")
		}),
		chromedp.Text("#app", &fileContent, chromedp.ByQuery),
	)

	if err != nil {
		t.Fatalf("Browser automation failed: %v", err)
	}

	if !(strings.Contains(fileContent, "File Browser") && strings.Contains(fileContent, "File Path:") && strings.Contains(fileContent, "docs/readme.txt")) {
		t.Errorf("Expected file browser content to show path docs/readme.txt, got: %s", fileContent)
	}

	t.Logf("Test passed! Wildcard routes work correctly")
}

// TestRouterDemo_NestedRoutes tests nested route rendering and layout persistence
func TestRouterDemo_NestedRoutes(t *testing.T) {
	server := devserver.NewServer("router_demo", "localhost:0")
	if err := server.Start(); err != nil {
		t.Fatalf("Failed to start dev server: %v", err)
	}
	defer server.Stop()

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
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
		chromedp.WaitVisible("#root", chromedp.ByQuery),
		chromedp.WaitNotPresent(".loading-indicator", chromedp.ByQuery),
		chromedp.Sleep(2*time.Second),

		// Navigate to Admin via SPA pushState
		chromedp.Evaluate(`(function(){ history.pushState({}, '', '/admin'); window.dispatchEvent(new PopStateEvent('popstate')); })()`, nil),
		chromedp.ActionFunc(func(ctx context.Context) error {
			deadline := time.Now().Add(8 * time.Second)
			var pathname string
			for time.Now().Before(deadline) {
				if err := chromedp.Evaluate(`window.location.pathname`, &pathname).Do(ctx); err == nil && pathname == "/admin" {
					return nil
				}
				time.Sleep(100 * time.Millisecond)
			}
			return fmt.Errorf("timed out waiting for pathname '/admin'")
		}),
		chromedp.Text("#app", &adminContent, chromedp.ByQuery),

		// Navigate to Admin Settings (nested route) via SPA pushState
		chromedp.Evaluate(`(function(){ history.pushState({}, '', '/admin/settings'); window.dispatchEvent(new PopStateEvent('popstate')); })()`, nil),
		chromedp.ActionFunc(func(ctx context.Context) error {
			deadline := time.Now().Add(8 * time.Second)
			var pathname string
			for time.Now().Before(deadline) {
				if err := chromedp.Evaluate(`window.location.pathname`, &pathname).Do(ctx); err == nil && pathname == "/admin/settings" {
					return nil
				}
				time.Sleep(100 * time.Millisecond)
			}
			return fmt.Errorf("timed out waiting for pathname '/admin/settings'")
		}),
		chromedp.Text("#app", &adminSettingsContent, chromedp.ByQuery),
	)

	if err != nil {
		t.Fatalf("Browser automation failed: %v", err)
	}

	if !(strings.Contains(adminContent, "Admin Panel") && strings.Contains(adminContent, "Admin Dashboard")) {
		t.Errorf("Expected admin dashboard view under layout, got: %s", adminContent)
	}

	if !(strings.Contains(adminSettingsContent, "Admin Panel") && strings.Contains(adminSettingsContent, "Admin Settings")) {
		t.Errorf("Expected admin settings view under layout, got: %s", adminSettingsContent)
	}

	// Verify that the layout is still present in nested route
	if !strings.Contains(adminSettingsContent, "Admin Panel") {
		t.Errorf("Expected admin settings to still contain layout 'Admin Panel', got: %s", adminSettingsContent)
	}

	t.Logf("Test passed! Nested routes work correctly")
}

// TestRouterDemo_NotFoundRoute tests the catch-all 404 route
func TestRouterDemo_NotFoundRoute(t *testing.T) {
	server := devserver.NewServer("router_demo", "localhost:0")
	if err := server.Start(); err != nil {
		t.Fatalf("Failed to start dev server: %v", err)
	}
	defer server.Stop()

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
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

	// Listen to console events for debugging
	chromedp.ListenTarget(ctx, func(ev interface{}) {
		switch e := ev.(type) {
		case *runtime.EventConsoleAPICalled:
			for _, arg := range e.Args {
				s := arg.Value.String()
				t.Logf("console: %s", s)
			}
		}
	})

	var notFoundContent string

	err := chromedp.Run(ctx,
		chromedp.Navigate(server.URL()),
		chromedp.WaitVisible("#app", chromedp.ByQuery),
		chromedp.WaitNotPresent(".loading-indicator", chromedp.ByQuery),
		// Switch to the non-existent route via SPA navigation to keep asset URLs working
		chromedp.Evaluate(`(function(){ history.pushState({}, '', '/non-existent-route'); window.dispatchEvent(new PopStateEvent('popstate')); })()`, nil),
		chromedp.ActionFunc(func(ctx context.Context) error {
			deadline := time.Now().Add(8 * time.Second)
			var pathname string
			for time.Now().Before(deadline) {
				if err := chromedp.Evaluate(`window.location.pathname`, &pathname).Do(ctx); err == nil && pathname == "/non-existent-route" {
					return nil
				}
				time.Sleep(100 * time.Millisecond)
			}
			return fmt.Errorf("timed out waiting for pathname '/non-existent-route'")
		}),
		chromedp.ActionFunc(func(ctx context.Context) error {
			deadline := time.Now().Add(10 * time.Second)
			var h1 string
			for time.Now().Before(deadline) {
				if err := chromedp.Evaluate(`(function(){ const el=document.querySelector('#app h1'); return el? el.textContent : ''; })()`, &h1).Do(ctx); err == nil {
					if strings.Contains(h1, "404 - Page Not Found") {
						return nil
					}
				}
				time.Sleep(100 * time.Millisecond)
			}
			return fmt.Errorf("timed out waiting for 404 heading; last h1='%s'", h1)
		}),
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
	var hrefBefore, hrefAfter string
	var histBefore, histAfter, histAfterBack int64

	err := chromedp.Run(ctx,
		chromedp.Navigate(server.URL()),
		chromedp.WaitVisible("#root", chromedp.ByQuery),
		chromedp.WaitVisible("#app", chromedp.ByQuery),
		chromedp.WaitNotPresent(".loading-indicator", chromedp.ByQuery),
		chromedp.Sleep(1*time.Second),
		chromedp.Evaluate(`window.location.href`, &hrefBefore),
		chromedp.Evaluate(`window.history.length`, &histBefore),
		chromedp.WaitVisible("h1", chromedp.ByQuery),
		chromedp.Text("h1", &homeContent, chromedp.ByQuery),

		// Step 1: Use native pushState to /about and dispatch popstate to trigger router render
		chromedp.Evaluate(`(function(){ history.pushState({}, '', '/about'); window.dispatchEvent(new PopStateEvent('popstate')); })()`, nil),
		chromedp.ActionFunc(func(ctx context.Context) error {
			var pathname string
			deadline := time.Now().Add(8 * time.Second)
			for time.Now().Before(deadline) {
				if err := chromedp.Evaluate(`window.location.pathname`, &pathname).Do(ctx); err == nil && pathname == "/about" {
					return nil
				}
				time.Sleep(100 * time.Millisecond)
			}
			return fmt.Errorf("timed out waiting for pathname '/about'")
		}),
		chromedp.WaitVisible("h1", chromedp.ByQuery),
		chromedp.Text("h1", &aboutContent, chromedp.ByQuery),

		// Step 2: pushState back to home and dispatch popstate to render
		chromedp.Evaluate(`(function(){ history.pushState({}, '', '/'); window.dispatchEvent(new PopStateEvent('popstate')); })()`, nil),
		chromedp.ActionFunc(func(ctx context.Context) error {
			var pathname string
			deadline := time.Now().Add(8 * time.Second)
			for time.Now().Before(deadline) {
				if err := chromedp.Evaluate(`window.location.pathname`, &pathname).Do(ctx); err == nil && pathname == "/" {
					return nil
				}
				time.Sleep(100 * time.Millisecond)
			}
			return fmt.Errorf("timed out waiting for pathname '/' after pushState home")
		}),
		chromedp.Text("h1", &homeContent, chromedp.ByQuery),
		chromedp.Evaluate(`window.location.href`, &hrefAfter),
		chromedp.Evaluate(`window.history.length`, &histAfter),

		// Step 3: Use browser back; expect to return to About
		chromedp.Evaluate(`window.history.back()`, nil),
		chromedp.ActionFunc(func(ctx context.Context) error {
			var pathname string
			deadline := time.Now().Add(8 * time.Second)
			for time.Now().Before(deadline) {
				_ = chromedp.Evaluate(`window.history.length`, &histAfterBack).Do(ctx)
				if err := chromedp.Evaluate(`window.location.pathname`, &pathname).Do(ctx); err == nil && pathname == "/about" {
					return nil
				}
				time.Sleep(100 * time.Millisecond)
			}
			return fmt.Errorf("timed out waiting for window.location.pathname='/about' ; last history.length=%d", histAfterBack)
		}),
		chromedp.Text("h1", &backToHomeContent, chromedp.ByQuery),
	)

	if err != nil {
		t.Fatalf("Browser automation failed: %v (before=%s len=%d after=%s len=%d afterBackLen=%d)", err, hrefBefore, histBefore, hrefAfter, histAfter, histAfterBack)
	}

	if !strings.Contains(aboutContent, "About UIWGo Router") {
		t.Errorf("Expected about content after first navigation, got: %s", aboutContent)
	}

	if !strings.Contains(homeContent, "Router Demo - Home") {
		t.Errorf("Expected home content after navigating back home, got: %s", homeContent)
	}

	if !strings.Contains(backToHomeContent, "About UIWGo Router") {
		t.Errorf("Expected about content after browser back, got: %s", backToHomeContent)
	}

	t.Logf("Test passed! Browser history navigation works correctly")
}
