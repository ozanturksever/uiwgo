//go:build !js && !wasm

package main

import (
	"testing"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/ozanturksever/uiwgo/internal/testhelpers"
)

func TestAppManagerDemo_Counter(t *testing.T) {
	// Start Vite dev server for this example
	server := testhelpers.NewViteServer("appmanager_demo", "localhost:0")
	if err := server.Start(); err != nil {
		t.Fatalf("Failed to start dev server: %v", err)
	}
	defer server.Stop()

	// Browser context
	chromedpCtx := testhelpers.MustNewChromedpContext(testhelpers.DefaultConfig())
	defer chromedpCtx.Cancel()

	var text string
	err := chromedp.Run(chromedpCtx.Ctx,
		// Navigate and wait for basic load and wasm initialization
		testhelpers.Actions.NavigateAndWaitForLoad(server.URL(), "#app"),
		chromedp.WaitVisible(`#counter-text`, chromedp.ByQuery),
		chromedp.Sleep(500*time.Millisecond),

		// Click Increment twice
		chromedp.Click(`#inc-btn`, chromedp.ByQuery),
		chromedp.Sleep(150*time.Millisecond),
		chromedp.Click(`#inc-btn`, chromedp.ByQuery),
		chromedp.Sleep(150*time.Millisecond),

		// Read text
		chromedp.Text(`#counter-text`, &text, chromedp.ByQuery),
	)
	if err != nil {
		t.Fatalf("chromedp run failed: %v", err)
	}
	if text != "Count: 2" {
		t.Fatalf("expected 'Count: 2', got %q", text)
	}

	// Decrement once and check
	err = chromedp.Run(chromedpCtx.Ctx,
		chromedp.Click(`#dec-btn`, chromedp.ByQuery),
		chromedp.Sleep(150*time.Millisecond),
		chromedp.Text(`#counter-text`, &text, chromedp.ByQuery),
	)
	if err != nil {
		t.Fatalf("chromedp run failed: %v", err)
	}
	if text != "Count: 1" {
		t.Fatalf("expected 'Count: 1', got %q", text)
	}

	// Reset and check
	err = chromedp.Run(chromedpCtx.Ctx,
		chromedp.Click(`#reset-btn`, chromedp.ByQuery),
		chromedp.Sleep(150*time.Millisecond),
		chromedp.Text(`#counter-text`, &text, chromedp.ByQuery),
	)
	if err != nil {
		t.Fatalf("chromedp run failed: %v", err)
	}
	if text != "Count: 0" {
		t.Fatalf("expected 'Count: 0', got %q", text)
	}
}

func TestAppManagerDemo_Routing(t *testing.T) {
	server := testhelpers.NewViteServer("appmanager_demo", "localhost:0")
	if err := server.Start(); err != nil {
		t.Fatalf("Failed to start dev server: %v", err)
	}
	defer server.Stop()

	chromedpCtx := testhelpers.MustNewChromedpContext(testhelpers.DefaultConfig())
	defer chromedpCtx.Cancel()

	err := chromedp.Run(chromedpCtx.Ctx,
		testhelpers.Actions.NavigateAndWaitForLoad(server.URL(), "#app"),
		chromedp.WaitVisible(`a[href="/about"]`, chromedp.ByQuery),
		chromedp.Click(`a[href="/about"]`, chromedp.ByQuery),
		chromedp.Sleep(300*time.Millisecond),
		chromedp.WaitVisible(`#about-page`, chromedp.ByQuery),
	)
	if err != nil {
		t.Fatalf("routing test failed: %v", err)
	}
}
