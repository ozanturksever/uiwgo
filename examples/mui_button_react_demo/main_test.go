//go:build !js && !wasm

package main

import (
    "testing"
    "time"

    "github.com/chromedp/chromedp"
    "github.com/ozanturksever/uiwgo/internal/testhelpers"
)

func TestMUIButtonDemo_RenderAndClick(t *testing.T) {
    server := testhelpers.NewViteServer("mui_button_react_demo", "localhost:0")
    if err := server.Start(); err != nil {
        t.Fatalf("Failed to start dev server: %v", err)
    }
    defer server.Stop()

    chromedpCtx := testhelpers.MustNewChromedpContext(testhelpers.DefaultConfig())
    defer chromedpCtx.Cancel()

    var btnText string

    err := chromedp.Run(chromedpCtx.Ctx,
        testhelpers.Actions.NavigateAndWaitForLoad(server.URL(), "body"),
        chromedp.WaitVisible(`#mui-demo-button`, chromedp.ByID),
        chromedp.Text(`#mui-demo-button`, &btnText, chromedp.ByID),
        testhelpers.Actions.ClickAndWait(`#mui-demo-button`, 200*time.Millisecond),
    )
    if err != nil {
        t.Fatalf("Browser automation failed: %v", err)
    }

    if btnText != "Hello world" {
        t.Fatalf("Expected button text 'Hello world', got '%s'", btnText)
    }
}
