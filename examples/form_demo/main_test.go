//go:build !js && !wasm

package main

import (
	"testing"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/ozanturksever/uiwgo/internal/testhelpers"
)

func TestFormDemo_PasswordMismatchValidation(t *testing.T) {
	server := testhelpers.NewViteServer("form_demo", "localhost:0")
	if err := server.Start(); err != nil {
		t.Fatalf("Failed to start vite server: %v", err)
	}
	defer server.Stop()

	chromedpCtx := testhelpers.MustNewChromedpContext(testhelpers.DefaultConfig())
	defer chromedpCtx.Cancel()

	err := chromedp.Run(chromedpCtx.Ctx,
		testhelpers.Actions.NavigateAndWaitForLoad(server.URL(), "#registration-form"),
		// Fill out the form with mismatched passwords
		testhelpers.Actions.SendKeysAndWait(`input[name="name"]`, "Test User", 100*time.Millisecond),
		testhelpers.Actions.SendKeysAndWait(`input[name="email"]`, "test@example.com", 100*time.Millisecond),
		testhelpers.Actions.SendKeysAndWait(`input[name="password"]`, "password123", 100*time.Millisecond),
		testhelpers.Actions.SendKeysAndWait(`input[name="confirm_password"]`, "differentpassword", 100*time.Millisecond),
		// Submit the form
		testhelpers.Actions.ClickAndWait(`#submit-btn`, 100*time.Millisecond),
		// Check for the validation error message (it should appear somewhere on the page)
		chromedp.WaitVisible(`//*[contains(text(), "Passwords must match")]`, chromedp.BySearch),
	)
	if err != nil {
		t.Fatalf("Failed to test password mismatch: %v", err)
	}
}