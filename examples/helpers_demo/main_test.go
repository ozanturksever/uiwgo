//go:build !js && !wasm

package main

import (
	"context"
	"testing"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/ozanturksever/uiwgo/internal/testhelpers"
)

func TestHelpersDemo_PageRender(t *testing.T) {
	server := testhelpers.NewViteServer("helpers_demo", "localhost:0")
	if err := server.Start(); err != nil {
		t.Fatalf("Failed to start dev server: %v", err)
	}
	defer server.Stop()

	chromedpCtx := testhelpers.MustNewChromedpContext(testhelpers.DefaultConfig())
	defer chromedpCtx.Cancel()

	err := chromedp.Run(chromedpCtx.Ctx,
		chromedp.Navigate(server.URL()),
		chromedp.WaitVisible("body"),
		chromedp.Sleep(2*time.Second),
		chromedp.WaitVisible(`h1`, chromedp.ByQuery),
		chromedp.WaitVisible(`.demo-section`, chromedp.ByQuery),
	)
	if err != nil {
		t.Fatalf("Page render test failed: %v", err)
	}

	// Verify main title
	var title string
	err = chromedp.Run(chromedpCtx.Ctx,
		chromedp.Text(`h1`, &title, chromedp.ByQuery),
	)
	if err != nil {
		t.Fatalf("Failed to get title: %v", err)
	}

	if title != "🛠️ UIwGo Helpers Demo" {
		t.Errorf("Expected title '🛠️ UIwGo Helpers Demo', got '%s'", title)
	}

	// Verify all demo sections are present
	demoSections := []string{
		"Show Helper - Conditional Rendering",
		"For Helper - List Rendering with Keys",
		"Index Helper - Index-based Reconciliation",
		"Switch/Match Helper - Conditional Branching",
		"Dynamic Helper - Dynamic Component Rendering",
		"Fragment Helper - Grouping Without Wrapper",
		"Portal Helper - Render to Different Location",
		"Memo Helper - Memoized Rendering",
		"Lazy Helper - Lazy Loading",
		"ErrorBoundary Helper - Error Handling",
	}

	for _, section := range demoSections {
		err = chromedp.Run(chromedpCtx.Ctx,
			chromedp.WaitVisible(`//h3[contains(text(), "`+section+`")]`, chromedp.BySearch),
		)
		if err != nil {
			t.Errorf("Demo section '%s' not found: %v", section, err)
		}
	}
}

func TestHelpersDemo_ShowHelper(t *testing.T) {
	server := testhelpers.NewViteServer("helpers_demo", "localhost:0")
	if err := server.Start(); err != nil {
		t.Fatalf("Failed to start dev server: %v", err)
	}
	defer server.Stop()

	chromedpCtx := testhelpers.MustNewChromedpContext(testhelpers.DefaultConfig())
	defer chromedpCtx.Cancel()

	err := chromedp.Run(chromedpCtx.Ctx,
		chromedp.Navigate(server.URL()),
		chromedp.WaitVisible("body"),
		chromedp.Sleep(2*time.Second),
		chromedp.WaitVisible(`//h3[contains(text(), "Show Helper")]`, chromedp.BySearch),
	)
	if err != nil {
		t.Fatalf("Failed to load page: %v", err)
	}

	// Initially, conditional content should be visible
	err = chromedp.Run(chromedpCtx.Ctx,
		chromedp.WaitVisible(`.conditional-content`, chromedp.ByQuery),
	)
	if err != nil {
		t.Fatalf("Conditional content should be initially visible: %v", err)
	}

	// Click toggle button to hide content
	err = chromedp.Run(chromedpCtx.Ctx,
		chromedp.Click(`//button[contains(text(), "Toggle Visibility")]`, chromedp.BySearch),
		chromedp.Sleep(100*time.Millisecond),
	)
	if err != nil {
		t.Fatalf("Failed to click toggle button: %v", err)
	}

	// Content should now be hidden
	ctx, cancel := context.WithTimeout(chromedpCtx.Ctx, 2*time.Second)
	defer cancel()
	err = chromedp.Run(ctx,
		chromedp.WaitNotPresent(`.conditional-content`, chromedp.ByQuery),
	)
	if err != nil {
		t.Fatalf("Conditional content should be hidden after toggle: %v", err)
	}

	// Click toggle again to show content
	err = chromedp.Run(chromedpCtx.Ctx,
		chromedp.Click(`//button[contains(text(), "Toggle Visibility")]`, chromedp.BySearch),
		chromedp.Sleep(100*time.Millisecond),
		chromedp.WaitVisible(`.conditional-content`, chromedp.ByQuery),
	)
	if err != nil {
		t.Fatalf("Conditional content should be visible after second toggle: %v", err)
	}
}

func TestHelpersDemo_ForHelper(t *testing.T) {
	server := testhelpers.NewViteServer("helpers_demo", "localhost:0")
	if err := server.Start(); err != nil {
		t.Fatalf("Failed to start dev server: %v", err)
	}
	defer server.Stop()

	chromedpCtx := testhelpers.MustNewChromedpContext(testhelpers.DefaultConfig())
	defer chromedpCtx.Cancel()

	err := chromedp.Run(chromedpCtx.Ctx,
		chromedp.Navigate(server.URL()),
		chromedp.WaitVisible("body"),
		chromedp.Sleep(2*time.Second),
		chromedp.WaitVisible(`//h3[contains(text(), "For Helper")]`, chromedp.BySearch),
	)
	if err != nil {
		t.Fatalf("Failed to load page: %v", err)
	}

	// Count initial todo items
	var initialCount int
	err = chromedp.Run(chromedpCtx.Ctx,
		chromedp.Evaluate(`document.querySelectorAll('li').length`, &initialCount),
	)
	if err != nil {
		t.Fatalf("Failed to count initial todos: %v", err)
	}

	if initialCount < 3 {
		t.Errorf("Expected at least 3 initial todos, got %d", initialCount)
	}

	// Add a new todo
	err = chromedp.Run(chromedpCtx.Ctx,
		chromedp.Click(`//button[contains(text(), "Add Todo")]`, chromedp.BySearch),
		chromedp.Sleep(100*time.Millisecond),
	)
	if err != nil {
		t.Fatalf("Failed to click Add Todo button: %v", err)
	}

	// Verify new todo was added
	var newCount int
	err = chromedp.Run(chromedpCtx.Ctx,
		chromedp.Evaluate(`document.querySelectorAll('li').length`, &newCount),
	)
	if err != nil {
		t.Fatalf("Failed to count todos after adding: %v", err)
	}

	if newCount != initialCount+1 {
		t.Errorf("Expected %d todos after adding, got %d", initialCount+1, newCount)
	}

	// Test checkbox functionality
	err = chromedp.Run(chromedpCtx.Ctx,
		chromedp.Click(`input[type="checkbox"]`, chromedp.ByQuery),
		chromedp.Sleep(100*time.Millisecond),
	)
	if err != nil {
		t.Fatalf("Failed to click checkbox: %v", err)
	}
}

func TestHelpersDemo_IndexHelper(t *testing.T) {
	server := testhelpers.NewViteServer("helpers_demo", "localhost:0")
	if err := server.Start(); err != nil {
		t.Fatalf("Failed to start dev server: %v", err)
	}
	defer server.Stop()

	chromedpCtx := testhelpers.MustNewChromedpContext(testhelpers.DefaultConfig())
	defer chromedpCtx.Cancel()

	err := chromedp.Run(chromedpCtx.Ctx,
		chromedp.Navigate(server.URL()),
		chromedp.WaitVisible("body"),
		chromedp.Sleep(2*time.Second),
		chromedp.WaitVisible(`//h3[contains(text(), "Index Helper")]`, chromedp.BySearch),
	)
	if err != nil {
		t.Fatalf("Failed to load page: %v", err)
	}

	// Get initial order of numbers
	var initialText string
	err = chromedp.Run(chromedpCtx.Ctx,
		chromedp.Text(`//h3[contains(text(), "Index Helper")]/following-sibling::ul`, &initialText, chromedp.BySearch),
	)
	if err != nil {
		t.Fatalf("Failed to get initial numbers: %v", err)
	}

	// Click shuffle button
	err = chromedp.Run(chromedpCtx.Ctx,
		chromedp.Click(`//button[contains(text(), "Shuffle Numbers")]`, chromedp.BySearch),
		chromedp.Sleep(200*time.Millisecond),
	)
	if err != nil {
		t.Fatalf("Failed to click shuffle button: %v", err)
	}

	// Get new order of numbers
	var newText string
	err = chromedp.Run(chromedpCtx.Ctx,
		chromedp.Text(`//h3[contains(text(), "Index Helper")]/following-sibling::ul`, &newText, chromedp.BySearch),
	)
	if err != nil {
		t.Fatalf("Failed to get shuffled numbers: %v", err)
	}

	// Verify that the list still contains the same number of items
	var itemCount int
	err = chromedp.Run(chromedpCtx.Ctx,
		chromedp.Evaluate(`document.querySelectorAll('h3:contains("Index Helper") + ul li').length`, &itemCount),
	)
	if err == nil && itemCount == 5 {
		// Good, we have the expected number of items
	} else {
		// Fallback check
		err = chromedp.Run(chromedpCtx.Ctx,
			chromedp.WaitVisible(`//h3[contains(text(), "Index Helper")]/following-sibling::ul/li[5]`, chromedp.BySearch),
		)
		if err != nil {
			t.Fatalf("Expected 5 number items after shuffle: %v", err)
		}
	}
}

func TestHelpersDemo_SwitchHelper(t *testing.T) {
	server := testhelpers.NewViteServer("helpers_demo", "localhost:0")
	if err := server.Start(); err != nil {
		t.Fatalf("Failed to start dev server: %v", err)
	}
	defer server.Stop()

	chromedpCtx := testhelpers.MustNewChromedpContext(testhelpers.DefaultConfig())
	defer chromedpCtx.Cancel()

	err := chromedp.Run(chromedpCtx.Ctx,
		chromedp.Navigate(server.URL()),
		chromedp.WaitVisible("body"),
		chromedp.Sleep(2*time.Second),
		chromedp.WaitVisible(`//h3[contains(text(), "Switch Helper")]`, chromedp.BySearch),
	)
	if err != nil {
		t.Fatalf("Failed to load page: %v", err)
	}

	// Test different status states
	statusTests := []struct {
		buttonText   string
		expectedText string
	}{
		{"Loading", "⏳ Loading..."},
		{"Success", "✅ Success! Operation completed."},
		{"Error", "❌ Error occurred! Please try again."},
		{"Unknown", "❓ Unknown status"},
	}

	for _, test := range statusTests {
		// Click the status button
		err = chromedp.Run(chromedpCtx.Ctx,
			chromedp.Click(`//button[contains(text(), "`+test.buttonText+`")]`, chromedp.BySearch),
			chromedp.Sleep(100*time.Millisecond),
		)
		if err != nil {
			t.Fatalf("Failed to click %s button: %v", test.buttonText, err)
		}

		// Verify the expected text appears
		err = chromedp.Run(chromedpCtx.Ctx,
			chromedp.WaitVisible(`//p[contains(text(), "`+test.expectedText+`")]`, chromedp.BySearch),
		)
		if err != nil {
			t.Errorf("Expected text '%s' not found for status '%s': %v", test.expectedText, test.buttonText, err)
		}
	}
}

func TestHelpersDemo_DynamicHelper(t *testing.T) {
	server := testhelpers.NewViteServer("helpers_demo", "localhost:0")
	if err := server.Start(); err != nil {
		t.Fatalf("Failed to start dev server: %v", err)
	}
	defer server.Stop()

	chromedpCtx := testhelpers.MustNewChromedpContext(testhelpers.DefaultConfig())
	defer chromedpCtx.Cancel()

	err := chromedp.Run(chromedpCtx.Ctx,
		chromedp.Navigate(server.URL()),
		chromedp.WaitVisible("body"),
		chromedp.Sleep(2*time.Second),
		chromedp.WaitVisible(`//h3[contains(text(), "Dynamic Helper")]`, chromedp.BySearch),
	)
	if err != nil {
		t.Fatalf("Failed to load page: %v", err)
	}

	// Test different view states
	viewTests := []struct {
		buttonText   string
		expectedText string
	}{
		{"Home View", "🏠 Home View - Welcome to the helpers demo!"},
		{"Profile View", "👤 Profile View - User settings and preferences"},
		{"Settings View", "⚙️ Settings View - Application configuration"},
	}

	for _, test := range viewTests {
		// Click the view button
		err = chromedp.Run(chromedpCtx.Ctx,
			chromedp.Click(`//button[contains(text(), "`+test.buttonText+`")]`, chromedp.BySearch),
			chromedp.Sleep(100*time.Millisecond),
		)
		if err != nil {
			t.Fatalf("Failed to click %s button: %v", test.buttonText, err)
		}

		// Verify the expected view content appears
		err = chromedp.Run(chromedpCtx.Ctx,
			chromedp.WaitVisible(`//p[contains(text(), "`+test.expectedText+`")]`, chromedp.BySearch),
		)
		if err != nil {
			t.Errorf("Expected view content '%s' not found for '%s': %v", test.expectedText, test.buttonText, err)
		}
	}
}

func TestHelpersDemo_PortalHelper(t *testing.T) {
	server := testhelpers.NewViteServer("helpers_demo", "localhost:0")
	if err := server.Start(); err != nil {
		t.Fatalf("Failed to start dev server: %v", err)
	}
	defer server.Stop()

	chromedpCtx := testhelpers.MustNewChromedpContext(testhelpers.DefaultConfig())
	defer chromedpCtx.Cancel()

	err := chromedp.Run(chromedpCtx.Ctx,
		chromedp.Navigate(server.URL()),
		chromedp.WaitVisible("body"),
		chromedp.Sleep(2*time.Second),
		chromedp.WaitVisible(`//h3[contains(text(), "Portal Helper")]`, chromedp.BySearch),
	)
	if err != nil {
		t.Fatalf("Failed to load page: %v", err)
	}

	// Click Open Modal button
	err = chromedp.Run(chromedpCtx.Ctx,
		chromedp.Click(`//button[contains(text(), "Open Modal")]`, chromedp.BySearch),
		chromedp.Sleep(200*time.Millisecond),
	)
	if err != nil {
		t.Fatalf("Failed to click Open Modal button: %v", err)
	}

	// Verify modal appears
	err = chromedp.Run(chromedpCtx.Ctx,
		chromedp.WaitVisible(`//h3[contains(text(), "Modal Title")]`, chromedp.BySearch),
	)
	if err != nil {
		t.Fatalf("Modal should be visible: %v", err)
	}

	// Verify modal content
	err = chromedp.Run(chromedpCtx.Ctx,
		chromedp.WaitVisible(`//p[contains(text(), "This modal is rendered using Portal helper")]`, chromedp.BySearch),
	)
	if err != nil {
		t.Fatalf("Modal content should be visible: %v", err)
	}

	// Click Close Modal button
	err = chromedp.Run(chromedpCtx.Ctx,
		chromedp.Click(`//button[contains(text(), "Close Modal")]`, chromedp.BySearch),
		chromedp.Sleep(200*time.Millisecond),
	)
	if err != nil {
		t.Fatalf("Failed to click Close Modal button: %v", err)
	}

	// Verify modal is hidden
	ctx, cancel := context.WithTimeout(chromedpCtx.Ctx, 2*time.Second)
	defer cancel()
	err = chromedp.Run(ctx,
		chromedp.WaitNotPresent(`//h3[contains(text(), "Modal Title")]`, chromedp.BySearch),
	)
	if err != nil {
		t.Fatalf("Modal should be hidden after close: %v", err)
	}
}

func TestHelpersDemo_MemoHelper(t *testing.T) {
	server := testhelpers.NewViteServer("helpers_demo", "localhost:0")
	if err := server.Start(); err != nil {
		t.Fatalf("Failed to start dev server: %v", err)
	}
	defer server.Stop()

	chromedpCtx := testhelpers.MustNewChromedpContext(testhelpers.DefaultConfig())
	defer chromedpCtx.Cancel()

	err := chromedp.Run(chromedpCtx.Ctx,
		chromedp.Navigate(server.URL()),
		chromedp.WaitVisible("body"),
		chromedp.Sleep(2*time.Second),
		chromedp.WaitVisible(`//h3[contains(text(), "Memo Helper")]`, chromedp.BySearch),
	)
	if err != nil {
		t.Fatalf("Failed to load page: %v", err)
	}

	// Count initial items in memo
	var initialText string
	err = chromedp.Run(chromedpCtx.Ctx,
		chromedp.Text(`//p[contains(text(), "Memoized list")]`, &initialText, chromedp.BySearch),
	)
	if err != nil {
		t.Fatalf("Failed to get initial memo text: %v", err)
	}

	// Add an item
	err = chromedp.Run(chromedpCtx.Ctx,
		chromedp.Click(`//button[contains(text(), "Add Item")]`, chromedp.BySearch),
		chromedp.Sleep(200*time.Millisecond),
	)
	if err != nil {
		t.Fatalf("Failed to click Add Item button: %v", err)
	}

	// Verify item count increased
	var newText string
	err = chromedp.Run(chromedpCtx.Ctx,
		chromedp.Text(`//p[contains(text(), "Memoized list")]`, &newText, chromedp.BySearch),
	)
	if err != nil {
		t.Fatalf("Failed to get updated memo text: %v", err)
	}

	if newText == initialText {
		t.Errorf("Memo should have updated after adding item")
	}

	// Test force re-render button (should not change memo)
	err = chromedp.Run(chromedpCtx.Ctx,
		chromedp.Click(`//button[contains(text(), "Force Re-render")]`, chromedp.BySearch),
		chromedp.Sleep(100*time.Millisecond),
	)
	if err != nil {
		t.Fatalf("Failed to click Force Re-render button: %v", err)
	}
}

func TestHelpersDemo_LazyHelper(t *testing.T) {
	server := testhelpers.NewViteServer("helpers_demo", "localhost:0")
	if err := server.Start(); err != nil {
		t.Fatalf("Failed to start dev server: %v", err)
	}
	defer server.Stop()

	chromedpCtx := testhelpers.MustNewChromedpContext(testhelpers.DefaultConfig())
	defer chromedpCtx.Cancel()

	err := chromedp.Run(chromedpCtx.Ctx,
		chromedp.Navigate(server.URL()),
		chromedp.WaitVisible("body"),
		chromedp.Sleep(2*time.Second),
		chromedp.WaitVisible(`//h3[contains(text(), "Lazy Helper")]`, chromedp.BySearch),
	)
	if err != nil {
		t.Fatalf("Failed to load page: %v", err)
	}

	// Initially, lazy component should not be loaded
	ctx, cancel := context.WithTimeout(chromedpCtx.Ctx, 1*time.Second)
	defer cancel()
	err = chromedp.Run(ctx,
		chromedp.WaitVisible(`//h4[contains(text(), "Lazy Loaded Component")]`, chromedp.BySearch),
	)
	if err == nil {
		t.Errorf("Lazy component should not be loaded initially")
	}

	// Click Load Lazy Component button
	err = chromedp.Run(chromedpCtx.Ctx,
		chromedp.Click(`//button[contains(text(), "Load Lazy Component")]`, chromedp.BySearch),
		chromedp.Sleep(200*time.Millisecond),
	)
	if err != nil {
		t.Fatalf("Failed to click Load Lazy Component button: %v", err)
	}

	// Now lazy component should be visible
	err = chromedp.Run(chromedpCtx.Ctx,
		chromedp.WaitVisible(`//h4[contains(text(), "🚀 Lazy Loaded Component")]`, chromedp.BySearch),
	)
	if err != nil {
		t.Fatalf("Lazy component should be loaded after button click: %v", err)
	}

	// Verify lazy component content
	err = chromedp.Run(chromedpCtx.Ctx,
		chromedp.WaitVisible(`//p[contains(text(), "This component was loaded on demand")]`, chromedp.BySearch),
	)
	if err != nil {
		t.Fatalf("Lazy component content should be visible: %v", err)
	}
}

func TestHelpersDemo_ErrorBoundaryHelper(t *testing.T) {
	server := testhelpers.NewViteServer("helpers_demo", "localhost:0")
	if err := server.Start(); err != nil {
		t.Fatalf("Failed to start dev server: %v", err)
	}
	defer server.Stop()

	chromedpCtx := testhelpers.MustNewChromedpContext(testhelpers.DefaultConfig())
	defer chromedpCtx.Cancel()

	err := chromedp.Run(chromedpCtx.Ctx,
		chromedp.Navigate(server.URL()),
		chromedp.WaitVisible("body"),
		chromedp.Sleep(2*time.Second),
		chromedp.WaitVisible(`//h3[contains(text(), "ErrorBoundary Helper")]`, chromedp.BySearch),
	)
	if err != nil {
		t.Fatalf("Failed to load page: %v", err)
	}

	// Click Toggle Error Component button to show the risky component
	err = chromedp.Run(chromedpCtx.Ctx,
		chromedp.Click(`//button[contains(text(), "Toggle Error Component")]`, chromedp.BySearch),
		chromedp.Sleep(200*time.Millisecond),
	)
	if err != nil {
		t.Fatalf("Failed to click Toggle Error Component button: %v", err)
	}

	// Verify risky component is shown
	err = chromedp.Run(chromedpCtx.Ctx,
		chromedp.WaitVisible(`//p[contains(text(), "This component might throw an error")]`, chromedp.BySearch),
	)
	if err != nil {
		t.Fatalf("Risky component should be visible: %v", err)
	}

	// Click Trigger Error button
	err = chromedp.Run(chromedpCtx.Ctx,
		chromedp.Click(`//button[contains(text(), "Trigger Error")]`, chromedp.BySearch),
		chromedp.Sleep(300*time.Millisecond),
	)
	if err != nil {
		t.Fatalf("Failed to click Trigger Error button: %v", err)
	}

	// Verify error boundary caught the error
	err = chromedp.Run(chromedpCtx.Ctx,
		chromedp.WaitVisible(`//h4[contains(text(), "🚨 Error Boundary Caught an Error")]`, chromedp.BySearch),
	)
	if err != nil {
		t.Fatalf("Error boundary should catch and display error: %v", err)
	}

	// Verify error message is shown
	err = chromedp.Run(chromedpCtx.Ctx,
		chromedp.WaitVisible(`//p[contains(text(), "Error: Simulated error for ErrorBoundary demo")]`, chromedp.BySearch),
	)
	if err != nil {
		t.Fatalf("Error message should be displayed: %v", err)
	}

	// Click Retry button
	err = chromedp.Run(chromedpCtx.Ctx,
		chromedp.Click(`//button[contains(text(), "Retry")]`, chromedp.BySearch),
		chromedp.Sleep(200*time.Millisecond),
	)
	if err != nil {
		t.Fatalf("Failed to click Retry button: %v", err)
	}

	// Verify error boundary is cleared and component is hidden
	ctx, cancel := context.WithTimeout(chromedpCtx.Ctx, 2*time.Second)
	defer cancel()
	err = chromedp.Run(ctx,
		chromedp.WaitNotPresent(`//h4[contains(text(), "🚨 Error Boundary Caught an Error")]`, chromedp.BySearch),
	)
	if err != nil {
		t.Fatalf("Error boundary should be cleared after retry: %v", err)
	}
}

func TestHelpersDemo_FragmentHelper(t *testing.T) {
	server := testhelpers.NewViteServer("helpers_demo", "localhost:0")
	if err := server.Start(); err != nil {
		t.Fatalf("Failed to start dev server: %v", err)
	}
	defer server.Stop()

	chromedpCtx := testhelpers.MustNewChromedpContext(testhelpers.DefaultConfig())
	defer chromedpCtx.Cancel()

	err := chromedp.Run(chromedpCtx.Ctx,
		chromedp.Navigate(server.URL()),
		chromedp.WaitVisible("body"),
		chromedp.Sleep(2*time.Second),
		chromedp.WaitVisible(`//h3[contains(text(), "Fragment Helper")]`, chromedp.BySearch),
	)
	if err != nil {
		t.Fatalf("Failed to load page: %v", err)
	}

	// Verify fragment content is present
	err = chromedp.Run(chromedpCtx.Ctx,
		chromedp.WaitVisible(`//h4[contains(text(), "Fragment Content")]`, chromedp.BySearch),
		chromedp.WaitVisible(`//p[contains(text(), "This paragraph is inside a fragment")]`, chromedp.BySearch),
		chromedp.WaitVisible(`//p[contains(text(), "This is another paragraph in the same fragment")]`, chromedp.BySearch),
		chromedp.WaitVisible(`//button[contains(text(), "Fragment Button")]`, chromedp.BySearch),
	)
	if err != nil {
		t.Fatalf("Fragment content should be visible: %v", err)
	}
}