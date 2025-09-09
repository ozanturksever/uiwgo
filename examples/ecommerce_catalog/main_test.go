//go:build !js && !wasm

package main

import (
	"strings"
	"testing"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/ozanturksever/uiwgo/internal/testhelpers"
)

func TestEcommerceCatalog_InitialRender(t *testing.T) {
	server := testhelpers.NewViteServer("ecommerce_catalog", "localhost:0")
	if err := server.Start(); err != nil {
		t.Fatalf("Failed to start dev server: %v", err)
	}
	defer server.Stop()

	chromedpCtx := testhelpers.MustNewChromedpContext(testhelpers.DefaultConfig())
	defer chromedpCtx.Cancel()

	var title, headerText string
	err := chromedp.Run(chromedpCtx.Ctx,
		chromedp.Navigate(server.URL()),
		chromedp.WaitVisible("body", chromedp.ByQuery),
		chromedp.Sleep(500*time.Millisecond), // Wait for WASM to initialize
		chromedp.Title(&title),
		chromedp.Text("h1", &headerText),
	)
	if err != nil {
		t.Fatalf("Test failed: %v", err)
	}

	if title != "E-commerce Product Catalog - UIwGo Example" {
		t.Errorf("Expected title 'E-commerce Product Catalog - UIwGo Example', got '%s'", title)
	}

	if headerText != "Product Catalog" {
		t.Errorf("Expected header 'Product Catalog', got '%s'", headerText)
	}
}

func TestEcommerceCatalog_ProductsLoad(t *testing.T) {
	server := testhelpers.NewViteServer("ecommerce_catalog", "localhost:0")
	if err := server.Start(); err != nil {
		t.Fatalf("Failed to start dev server: %v", err)
	}
	defer server.Stop()

	chromedpCtx := testhelpers.MustNewChromedpContext(testhelpers.DefaultConfig())
	defer chromedpCtx.Cancel()

	err := chromedp.Run(chromedpCtx.Ctx,
		chromedp.Navigate(server.URL()),
		chromedp.WaitVisible("body", chromedp.ByQuery),
		chromedp.Sleep(500*time.Millisecond), // Wait for WASM to initialize
		// Wait for loading to complete and products to appear
		chromedp.WaitNotPresent(".loading-state", chromedp.ByQuery),
		chromedp.WaitVisible(".product-grid", chromedp.ByQuery),
	)
	if err != nil {
		t.Fatalf("Test failed: %v", err)
	}

	// Check that products are loaded
	var productCount int
	err = chromedp.Run(chromedpCtx.Ctx,
		chromedp.Evaluate(`document.querySelectorAll('.product-card').length`, &productCount),
	)
	if err != nil {
		t.Fatalf("Failed to count products: %v", err)
	}

	if productCount == 0 {
		t.Error("Expected products to be loaded, but found none")
	}

	// Verify sample products are present
	var hasWirelessHeadphones bool
	err = chromedp.Run(chromedpCtx.Ctx,
		chromedp.Evaluate(`document.body.textContent.includes('Wireless Headphones')`, &hasWirelessHeadphones),
	)
	if err != nil {
		t.Fatalf("Failed to check for sample product: %v", err)
	}

	if !hasWirelessHeadphones {
		t.Error("Expected to find 'Wireless Headphones' product")
	}
}

func TestEcommerceCatalog_SearchFunctionality(t *testing.T) {
	server := testhelpers.NewViteServer("ecommerce_catalog", "localhost:0")
	if err := server.Start(); err != nil {
		t.Fatalf("Failed to start dev server: %v", err)
	}
	defer server.Stop()

	chromedpCtx := testhelpers.MustNewChromedpContext(testhelpers.DefaultConfig())
	defer chromedpCtx.Cancel()

	err := chromedp.Run(chromedpCtx.Ctx,
		chromedp.Navigate(server.URL()),
		chromedp.WaitVisible("body", chromedp.ByQuery),
		chromedp.Sleep(500*time.Millisecond), // Wait for WASM to initialize
		chromedp.WaitNotPresent(".loading-state", chromedp.ByQuery),
		chromedp.WaitVisible(".product-grid", chromedp.ByQuery),
	)
	if err != nil {
		t.Fatalf("Test failed: %v", err)
	}

	// Get initial product count
	var initialCount int
	err = chromedp.Run(chromedpCtx.Ctx,
		chromedp.Evaluate(`document.querySelectorAll('.product-card').length`, &initialCount),
	)
	if err != nil {
		t.Fatalf("Failed to get initial product count: %v", err)
	}

	// Search for "laptop"
	err = chromedp.Run(chromedpCtx.Ctx,
		chromedp.Clear(`input[placeholder="Search products..."]`),
		chromedp.SendKeys(`input[placeholder="Search products..."]`, "laptop"),
		// Wait a moment for the search to filter
		chromedp.Sleep(500*time.Millisecond),
	)
	if err != nil {
		t.Fatalf("Failed to search: %v", err)
	}

	// Check filtered results
	var filteredCount int
	err = chromedp.Run(chromedpCtx.Ctx,
		chromedp.Evaluate(`document.querySelectorAll('.product-card').length`, &filteredCount),
	)
	if err != nil {
		t.Fatalf("Failed to get filtered product count: %v", err)
	}

	if filteredCount >= initialCount {
		t.Errorf("Expected filtered count (%d) to be less than initial count (%d)", filteredCount, initialCount)
	}

	// Verify laptop is in results
	var hasLaptop bool
	err = chromedp.Run(chromedpCtx.Ctx,
		chromedp.Evaluate(`document.body.textContent.includes('Laptop')`, &hasLaptop),
	)
	if err != nil {
		t.Fatalf("Failed to check for laptop: %v", err)
	}

	if !hasLaptop {
		t.Error("Expected to find 'Laptop' in search results")
	}

	// Clear search and verify all products return
	err = chromedp.Run(chromedpCtx.Ctx,
		chromedp.Clear(`input[placeholder="Search products..."]`),
		chromedp.Sleep(500*time.Millisecond),
	)
	if err != nil {
		t.Fatalf("Failed to clear search: %v", err)
	}

	var clearedCount int
	err = chromedp.Run(chromedpCtx.Ctx,
		chromedp.Evaluate(`document.querySelectorAll('.product-card').length`, &clearedCount),
	)
	if err != nil {
		t.Fatalf("Failed to get cleared product count: %v", err)
	}

	if clearedCount != initialCount {
		t.Errorf("Expected cleared count (%d) to equal initial count (%d)", clearedCount, initialCount)
	}
}

func TestEcommerceCatalog_CategoryFilter(t *testing.T) {
	server := testhelpers.NewViteServer("ecommerce_catalog", "localhost:0")
	if err := server.Start(); err != nil {
		t.Fatalf("Failed to start dev server: %v", err)
	}
	defer server.Stop()

	chromedpCtx := testhelpers.MustNewChromedpContext(testhelpers.DefaultConfig())
	defer chromedpCtx.Cancel()

	err := chromedp.Run(chromedpCtx.Ctx,
		chromedp.Navigate(server.URL()),
		chromedp.WaitVisible("body", chromedp.ByQuery),
		chromedp.Sleep(500*time.Millisecond), // Wait for WASM to initialize
		chromedp.WaitNotPresent(".loading-state", chromedp.ByQuery),
		chromedp.WaitVisible(".product-grid", chromedp.ByQuery),
	)
	if err != nil {
		t.Fatalf("Test failed: %v", err)
	}

	// Select Electronics category
	err = chromedp.Run(chromedpCtx.Ctx,
		chromedp.SetValue(`select`, "Electronics", chromedp.ByQuery),
		chromedp.Sleep(500*time.Millisecond),
	)
	if err != nil {
		t.Fatalf("Failed to select category: %v", err)
	}

	// Verify only Electronics products are shown
	var electronicsOnly bool
	err = chromedp.Run(chromedpCtx.Ctx,
		chromedp.Evaluate(`
			const categories = Array.from(document.querySelectorAll('.category')).map(el => el.textContent);
			categories.every(cat => cat === 'Electronics')
		`, &electronicsOnly),
	)
	if err != nil {
		t.Fatalf("Failed to check category filter: %v", err)
	}

	if !electronicsOnly {
		t.Error("Expected only Electronics products to be shown")
	}
}

func TestEcommerceCatalog_ViewModeToggle(t *testing.T) {
	server := testhelpers.NewViteServer("ecommerce_catalog", "localhost:0")
	if err := server.Start(); err != nil {
		t.Fatalf("Failed to start dev server: %v", err)
	}
	defer server.Stop()

	chromedpCtx := testhelpers.MustNewChromedpContext(testhelpers.DefaultConfig())
	defer chromedpCtx.Cancel()

	err = chromedp.Run(chromedpCtx.Ctx,
		chromedp.Navigate(server.URL()),
		chromedp.WaitVisible("body", chromedp.ByQuery),
		chromedp.Sleep(500*time.Millisecond), // Wait for WASM to initialize
		chromedp.WaitNotPresent(".loading-state", chromedp.ByQuery),
		chromedp.WaitVisible(".product-grid", chromedp.ByQuery),
	)
	if err != nil {
		t.Fatalf("Test failed: %v", err)
	}

	// Initially should be in grid view
	var hasGrid bool
	err = chromedp.Run(chromedpCtx.Ctx,
		chromedp.Evaluate(`document.querySelector('.product-grid') !== null`, &hasGrid),
	)
	if err != nil {
		t.Fatalf("Failed to check grid view: %v", err)
	}

	if !hasGrid {
		t.Error("Expected to start in grid view")
	}

	// Click List button
	err = chromedp.Run(chromedpCtx.Ctx,
		chromedp.Click(`button:contains("List")`, chromedp.ByQuery),
		chromedp.Sleep(500*time.Millisecond),
	)
	if err != nil {
		t.Fatalf("Failed to click List button: %v", err)
	}

	// Should now be in list view
	var hasList bool
	err = chromedp.Run(chromedpCtx.Ctx,
		chromedp.Evaluate(`document.querySelector('.product-list') !== null`, &hasList),
	)
	if err != nil {
		t.Fatalf("Failed to check list view: %v", err)
	}

	if !hasList {
		t.Error("Expected to switch to list view")
	}

	// Click Grid button to switch back
	err = chromedp.Run(chromedpCtx.Ctx,
		chromedp.Click(`button:contains("Grid")`, chromedp.ByQuery),
		chromedp.Sleep(500*time.Millisecond),
	)
	if err != nil {
		t.Fatalf("Failed to click Grid button: %v", err)
	}

	// Should be back in grid view
	err = chromedp.Run(chromedpCtx.Ctx,
		chromedp.Evaluate(`document.querySelector('.product-grid') !== null`, &hasGrid),
	)
	if err != nil {
		t.Fatalf("Failed to check grid view return: %v", err)
	}

	if !hasGrid {
		t.Error("Expected to switch back to grid view")
	}
}

func TestEcommerceCatalog_SortFunctionality(t *testing.T) {
	server := testhelpers.NewViteServer("ecommerce_catalog", "localhost:0")
	if err := server.Start(); err != nil {
		t.Fatalf("Failed to start dev server: %v", err)
	}
	defer server.Stop()

	chromedpCtx := testhelpers.MustNewChromedpContext(testhelpers.DefaultConfig())
	defer chromedpCtx.Cancel()

	err := chromedp.Run(chromedpCtx.Ctx,
		chromedp.Navigate(server.URL()),
		chromedp.WaitVisible("body", chromedp.ByQuery),
		chromedp.Sleep(500*time.Millisecond), // Wait for WASM to initialize
		chromedp.WaitNotPresent(".loading-state", chromedp.ByQuery),
		chromedp.WaitVisible(".product-grid", chromedp.ByQuery),
	)
	if err != nil {
		t.Fatalf("Test failed: %v", err)
	}

	// Get initial product names
	var initialNames []string
	err = chromedp.Run(chromedpCtx.Ctx,
		chromedp.Evaluate(`
			Array.from(document.querySelectorAll('.product-info h3')).map(el => el.textContent)
		`, &initialNames),
	)
	if err != nil {
		t.Fatalf("Failed to get initial product names: %v", err)
	}

	// Sort by price
	err = chromedp.Run(chromedpCtx.Ctx,
		chromedp.SetValue(`select[value="name"]`, "price", chromedp.ByQuery),
		chromedp.Sleep(500*time.Millisecond),
	)
	if err != nil {
		t.Fatalf("Failed to sort by price: %v", err)
	}

	// Get sorted product names
	var sortedNames []string
	err = chromedp.Run(chromedpCtx.Ctx,
		chromedp.Evaluate(`
			Array.from(document.querySelectorAll('.product-info h3')).map(el => el.textContent)
		`, &sortedNames),
	)
	if err != nil {
		t.Fatalf("Failed to get sorted product names: %v", err)
	}

	// Verify order changed
	if len(initialNames) > 0 && len(sortedNames) > 0 {
		orderChanged := false
		for i := range initialNames {
			if i < len(sortedNames) && initialNames[i] != sortedNames[i] {
				orderChanged = true
				break
			}
		}
		if !orderChanged {
			t.Error("Expected product order to change when sorting by price")
		}
	}

	// Test sort direction toggle
	err = chromedp.Run(chromedpCtx.Ctx,
		chromedp.Click(`button:contains("Asc")`, chromedp.ByQuery),
		chromedp.Sleep(500*time.Millisecond),
	)
	if err != nil {
		t.Fatalf("Failed to toggle sort direction: %v", err)
	}

	// Verify button text changed to Desc
	var buttonText string
	err = chromedp.Run(chromedpCtx.Ctx,
		chromedp.Text(`button:contains("Desc")`, &buttonText),
	)
	if err != nil {
		t.Fatalf("Failed to check sort button text: %v", err)
	}

	if !strings.Contains(buttonText, "Desc") {
		t.Error("Expected sort button to show 'Desc' after toggle")
	}
}

func TestEcommerceCatalog_OutOfStockFilter(t *testing.T) {
	server := testhelpers.NewViteServer("ecommerce_catalog", "localhost:0")
	if err := server.Start(); err != nil {
		t.Fatalf("Failed to start dev server: %v", err)
	}
	defer server.Stop()

	chromedpCtx := testhelpers.MustNewChromedpContext(testhelpers.DefaultConfig())
	defer chromedpCtx.Cancel()

	err := chromedp.Run(chromedpCtx.Ctx,
		chromedp.Navigate(server.URL()),
		chromedp.WaitVisible("body", chromedp.ByQuery),
		chromedp.Sleep(500*time.Millisecond), // Wait for WASM to initialize
		chromedp.WaitNotPresent(".loading-state", chromedp.ByQuery),
		chromedp.WaitVisible(".product-grid", chromedp.ByQuery),
	)
	if err != nil {
		t.Fatalf("Test failed: %v", err)
	}

	// Get initial product count
	var initialCount int
	err = chromedp.Run(chromedpCtx.Ctx,
		chromedp.Evaluate(`document.querySelectorAll('.product-card').length`, &initialCount),
	)
	if err != nil {
		t.Fatalf("Failed to get initial product count: %v", err)
	}

	// Uncheck "Show out of stock" checkbox
	err = chromedp.Run(chromedpCtx.Ctx,
		chromedp.Click(`input[type="checkbox"]`, chromedp.ByQuery),
		chromedp.Sleep(500*time.Millisecond),
	)
	if err != nil {
		t.Fatalf("Failed to uncheck out of stock filter: %v", err)
	}

	// Get filtered product count
	var filteredCount int
	err = chromedp.Run(chromedpCtx.Ctx,
		chromedp.Evaluate(`document.querySelectorAll('.product-card').length`, &filteredCount),
	)
	if err != nil {
		t.Fatalf("Failed to get filtered product count: %v", err)
	}

	if filteredCount >= initialCount {
		t.Errorf("Expected filtered count (%d) to be less than initial count (%d) when hiding out of stock", filteredCount, initialCount)
	}

	// Verify no "Out of Stock" labels are visible
	var hasOutOfStock bool
	err = chromedp.Run(chromedpCtx.Ctx,
		chromedp.Evaluate(`document.body.textContent.includes('Out of Stock')`, &hasOutOfStock),
	)
	if err != nil {
		t.Fatalf("Failed to check for out of stock labels: %v", err)
	}

	if hasOutOfStock {
		t.Error("Expected no 'Out of Stock' labels when filter is unchecked")
	}
}

func TestEcommerceCatalog_EmptyState(t *testing.T) {
	server := testhelpers.NewViteServer("ecommerce_catalog", "localhost:0")
	if err := server.Start(); err != nil {
		t.Fatalf("Failed to start dev server: %v", err)
	}
	defer server.Stop()

	chromedpCtx := testhelpers.MustNewChromedpContext(testhelpers.DefaultConfig())
	defer chromedpCtx.Cancel()

	err := chromedp.Run(chromedpCtx.Ctx,
		chromedp.Navigate(server.URL()),
		chromedp.WaitVisible("body", chromedp.ByQuery),
		chromedp.Sleep(500*time.Millisecond), // Wait for WASM to initialize
		chromedp.WaitNotPresent(".loading-state", chromedp.ByQuery),
		chromedp.WaitVisible(".product-grid", chromedp.ByQuery),
	)
	if err != nil {
		t.Fatalf("Test failed: %v", err)
	}

	// Search for something that doesn't exist
	err = chromedp.Run(chromedpCtx.Ctx,
		chromedp.Clear(`input[placeholder="Search products..."]`),
		chromedp.SendKeys(`input[placeholder="Search products..."]`, "nonexistentproduct"),
		chromedp.Sleep(500*time.Millisecond),
	)
	if err != nil {
		t.Fatalf("Failed to search for nonexistent product: %v", err)
	}

	// Verify empty state is shown
	var hasEmptyState bool
	err = chromedp.Run(chromedpCtx.Ctx,
		chromedp.Evaluate(`document.querySelector('.empty-state') !== null`, &hasEmptyState),
	)
	if err != nil {
		t.Fatalf("Failed to check for empty state: %v", err)
	}

	if !hasEmptyState {
		t.Error("Expected empty state to be shown when no products match search")
	}

	// Verify empty state message
	var emptyMessage string
	err = chromedp.Run(chromedpCtx.Ctx,
		chromedp.Text(".empty-state h3", &emptyMessage),
	)
	if err != nil {
		t.Fatalf("Failed to get empty state message: %v", err)
	}

	if emptyMessage != "No products found" {
		t.Errorf("Expected empty state message 'No products found', got '%s'", emptyMessage)
	}
}