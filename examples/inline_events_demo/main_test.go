//go:build !js && !wasm

package main

import (
	"testing"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/ozanturksever/uiwgo/internal/testhelpers"
)

func startServer(t *testing.T) *testhelpers.ViteServer {
	t.Helper()
	server := testhelpers.NewViteServer("inline_events_demo", "localhost:0")
	if err := server.Start(); err != nil {
		t.Fatalf("Failed to start dev server: %v", err)
	}
	return server
}

func TestInlineEventsDemo_ClickHandlers(t *testing.T) {
	server := startServer(t)
	defer server.Stop()

	ctx := testhelpers.MustNewChromedpContext(testhelpers.ExtendedTimeoutConfig())
	defer ctx.Cancel()

	var countText string
	err := chromedp.Run(ctx.Ctx,
		chromedp.Navigate(server.URL()),
		chromedp.WaitVisible(`#inc-btn`, chromedp.ByID),
		chromedp.Click(`#inc-btn`, chromedp.ByID),
		chromedp.Sleep(200*time.Millisecond),
		chromedp.Text(`#count-display`, &countText, chromedp.ByID),
	)
	if err != nil {
		t.Fatalf("Browser actions failed: %v", err)
	}
	if countText == "" || countText == "Count: 0" {
		t.Fatalf("Expected count to increase, got: %q", countText)
	}
}

func TestInlineEventsDemo_InputAndChangeHandlers(t *testing.T) {
	server := startServer(t)
	defer server.Stop()

	ctx := testhelpers.MustNewChromedpContext(testhelpers.ExtendedTimeoutConfig())
	defer ctx.Cancel()

	var hello string
	var colorText string
	err := chromedp.Run(ctx.Ctx,
		chromedp.Navigate(server.URL()),
		chromedp.WaitVisible(`#name-input`, chromedp.ByID),
		chromedp.SendKeys(`#name-input`, "Alice", chromedp.ByID),
		chromedp.Sleep(150*time.Millisecond),
		chromedp.Text(`#hello-output`, &hello, chromedp.ByID),

		// Change select value to green and fire change
		chromedp.Evaluate(`(function(){
		  const sel = document.getElementById('color-select');
		  sel.value = 'green';
		  sel.dispatchEvent(new Event('change', { bubbles: true }));
		  return true;
		})()`, nil),
		chromedp.Sleep(150*time.Millisecond),
		chromedp.Text(`#color-output`, &colorText, chromedp.ByID),
	)
	if err != nil {
		t.Fatalf("Browser actions failed: %v", err)
	}
	if hello != "Hello, Alice" {
		t.Fatalf("Expected greeting to update, got: %q", hello)
	}
	if colorText != "Color: green" {
		t.Fatalf("Expected color to update to green, got: %q", colorText)
	}
}

func TestInlineEventsDemo_KeydownEnterAndEscape(t *testing.T) {
	server := startServer(t)
	defer server.Stop()

	ctx := testhelpers.MustNewChromedpContext(testhelpers.ExtendedTimeoutConfig())
	defer ctx.Cancel()

	var todoCount int
	err := chromedp.Run(ctx.Ctx,
		chromedp.Navigate(server.URL()),
		chromedp.WaitVisible(`#todo-input`, chromedp.ByID),
		chromedp.SendKeys(`#todo-input`, "First", chromedp.ByID),
		// Press Enter
		chromedp.SendKeys(`#todo-input`, "\r", chromedp.ByID),
		chromedp.Sleep(200*time.Millisecond),
		chromedp.Evaluate(`document.querySelectorAll('#todo-list .todo-item').length`, &todoCount),
	)
	if err != nil {
		t.Fatalf("Browser actions failed: %v", err)
	}
	if todoCount != 1 {
		t.Fatalf("Expected 1 todo after Enter, got %d", todoCount)
	}

	// Type something and hit Escape - input should clear
	err = chromedp.Run(ctx.Ctx,
		chromedp.SendKeys(`#todo-input`, "temp", chromedp.ByID),
		chromedp.SendKeys(`#todo-input`, "\u001b", chromedp.ByID), // Escape key
		chromedp.Sleep(150*time.Millisecond),
	)
	if err != nil {
		t.Fatalf("Browser actions failed: %v", err)
	}
	var inputVal string
	err = chromedp.Run(ctx.Ctx,
		chromedp.Evaluate(`document.getElementById('todo-input').value`, &inputVal),
	)
	if err != nil {
		t.Fatalf("Eval failed: %v", err)
	}
	if inputVal != "" {
		t.Fatalf("Expected input to clear on Escape, got: %q", inputVal)
	}
}
