//go:build !js && !wasm

package main

import (
	"testing"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/ozanturksever/uiwgo/internal/testhelpers"
)

func TestMultiStepForm(t *testing.T) {
	server := testhelpers.NewViteServer("multi_step_form", "localhost:0")
	if err := server.Start(); err != nil {
		t.Fatalf("Failed to start dev server: %v", err)
	}
	defer server.Stop()

	chromedpCtx := testhelpers.MustNewChromedpContext(testhelpers.DefaultConfig())
	defer chromedpCtx.Cancel()

	err := chromedp.Run(chromedpCtx.Ctx,
		testhelpers.Actions.NavigateAndWaitForLoad(server.URL(), "body"),
		// Wait for WASM to initialize
		chromedp.Sleep(2*time.Second),
	)
	if err != nil {
		t.Fatalf("Failed to navigate to multi-step form: %v", err)
	}

	// Test that the form loads correctly
	var title string
	err = chromedp.Run(chromedpCtx.Ctx,
		chromedp.Text("h1", &title, chromedp.ByQuery),
	)
	if err != nil {
		t.Fatalf("Failed to get title: %v", err)
	}

	if title != "Multi-Step Form" {
		t.Errorf("Expected title 'Multi-Step Form', got '%s'", title)
	}

	// Test that step indicators are visible
	var stepIndicatorText string
	err = chromedp.Run(chromedpCtx.Ctx,
		chromedp.Text(".step-indicator", &stepIndicatorText, chromedp.ByQuery),
	)
	if err != nil {
		t.Fatalf("Failed to get step indicator: %v", err)
	}

	// Test navigation buttons
	var prevButtonDisabled string
	err = chromedp.Run(chromedpCtx.Ctx,
		chromedp.AttributeValue("button", "disabled", &prevButtonDisabled, nil, chromedp.ByQuery),
	)
	if err != nil {
		t.Fatalf("Failed to check previous button: %v", err)
	}

	if prevButtonDisabled == "" {
		t.Error("Previous button should be disabled on first step")
	}
}

func TestMultiStepFormNavigation(t *testing.T) {
	server := testhelpers.NewViteServer("multi_step_form", "localhost:0")
	if err := server.Start(); err != nil {
		t.Fatalf("Failed to start dev server: %v", err)
	}
	defer server.Stop()

	chromedpCtx := testhelpers.MustNewChromedpContext(testhelpers.DefaultConfig())
	defer chromedpCtx.Cancel()

	err := chromedp.Run(chromedpCtx.Ctx,
		testhelpers.Actions.NavigateAndWaitForLoad(server.URL(), "body"),
		// Wait for WASM to initialize
		chromedp.Sleep(2*time.Second),
	)
	if err != nil {
		t.Fatalf("Failed to navigate to multi-step form: %v", err)
	}

	// Fill in personal info step
	err = chromedp.Run(chromedpCtx.Ctx,
		chromedp.SendKeys("input[type='text']", "John", chromedp.ByQuery),
		chromedp.SendKeys("input[type='text']", "Doe", chromedp.ByQueryAll),
		chromedp.SendKeys("input[type='date']", "1990-01-01", chromedp.ByQuery),
	)
	if err != nil {
		t.Fatalf("Failed to fill personal info: %v", err)
	}

	// Click next button
	err = chromedp.Run(chromedpCtx.Ctx,
		chromedp.Click("button:not([disabled])", chromedp.ByQuery),
	)
	if err != nil {
		t.Fatalf("Failed to click next button: %v", err)
	}

	// Wait for transition
	time.Sleep(500 * time.Millisecond)

	// Verify we're on step 2 (contact info)
	var stepTitle string
	err = chromedp.Run(chromedpCtx.Ctx,
		chromedp.Text("h2", &stepTitle, chromedp.ByQuery),
	)
	if err != nil {
		t.Fatalf("Failed to get step title: %v", err)
	}

	if stepTitle != "Contact Information" {
		t.Errorf("Expected step title 'Contact Information', got '%s'", stepTitle)
	}
}

func TestMultiStepFormValidation(t *testing.T) {
	server := testhelpers.NewViteServer("multi_step_form", "localhost:0")
	if err := server.Start(); err != nil {
		t.Fatalf("Failed to start dev server: %v", err)
	}
	defer server.Stop()

	chromedpCtx := testhelpers.MustNewChromedpContext(testhelpers.DefaultConfig())
	defer chromedpCtx.Cancel()

	err := chromedp.Run(chromedpCtx.Ctx,
		testhelpers.Actions.NavigateAndWaitForLoad(server.URL(), "body"),
		// Wait for WASM to initialize
		chromedp.Sleep(2*time.Second),
	)
	if err != nil {
		t.Fatalf("Failed to navigate to multi-step form: %v", err)
	}

	// Try to proceed without filling required fields
	err = chromedp.Run(chromedpCtx.Ctx,
		chromedp.Click("button:not([disabled])", chromedp.ByQuery),
	)
	if err != nil {
		t.Fatalf("Failed to click next button: %v", err)
	}

	// Wait for validation
	time.Sleep(500 * time.Millisecond)

	// Check if error messages are displayed
	var errorMessage string
	err = chromedp.Run(chromedpCtx.Ctx,
		chromedp.Text(".error-message", &errorMessage, chromedp.ByQuery),
	)
	if err != nil {
		t.Fatalf("Failed to get error message: %v", err)
	}

	if errorMessage == "" {
		t.Error("Expected error message for missing required fields")
	}
}

func TestMultiStepFormSubmission(t *testing.T) {
	server := testhelpers.NewViteServer("multi_step_form", "localhost:0")
	if err := server.Start(); err != nil {
		t.Fatalf("Failed to start dev server: %v", err)
	}
	defer server.Stop()

	chromedpCtx := testhelpers.MustNewChromedpContext(testhelpers.DefaultConfig())
	defer chromedpCtx.Cancel()

	err := chromedp.Run(chromedpCtx.Ctx,
		testhelpers.Actions.NavigateAndWaitForLoad(server.URL(), "body"),
		// Wait for WASM to initialize
		chromedp.Sleep(2*time.Second),
	)
	if err != nil {
		t.Fatalf("Failed to navigate to multi-step form: %v", err)
	}

	// Fill personal info step
	err = chromedp.Run(chromedpCtx.Ctx,
		chromedp.SendKeys("input[type='text']", "John", chromedp.ByQuery),
		chromedp.SendKeys("input[type='text']", "Doe", chromedp.ByQueryAll),
		chromedp.SendKeys("input[type='date']", "1990-01-01", chromedp.ByQuery),
	)
	if err != nil {
		t.Fatalf("Failed to fill personal info: %v", err)
	}

	// Navigate through all steps
	for i := 0; i < 3; i++ {
		err = chromedp.Run(chromedpCtx.Ctx,
			chromedp.Click("button:not([disabled])", chromedp.ByQuery),
		)
		if err != nil {
			t.Fatalf("Failed to navigate to next step: %v", err)
		}
		time.Sleep(500 * time.Millisecond)
	}

	// Fill contact info
	err = chromedp.Run(chromedpCtx.Ctx,
		chromedp.SendKeys("input[type='email']", "john.doe@example.com", chromedp.ByQuery),
		chromedp.SendKeys("input[type='tel']", "+1234567890", chromedp.ByQueryAll),
		chromedp.SendKeys("textarea", "123 Main St", chromedp.ByQuery),
		chromedp.SendKeys("input[type='text']", "New York", chromedp.ByQueryAll),
		chromedp.SendKeys("input[type='text']", "USA", chromedp.ByQueryAll),
	)
	if err != nil {
		t.Fatalf("Failed to fill contact info: %v", err)
	}

	// Navigate to preferences
	err = chromedp.Run(chromedpCtx.Ctx,
		chromedp.Click("button:not([disabled])", chromedp.ByQuery),
	)
	if err != nil {
		t.Fatalf("Failed to navigate to preferences: %v", err)
	}
	time.Sleep(500 * time.Millisecond)

	// Navigate to review
	err = chromedp.Run(chromedpCtx.Ctx,
		chromedp.Click("button:not([disabled])", chromedp.ByQuery),
	)
	if err != nil {
		t.Fatalf("Failed to navigate to review: %v", err)
	}
	time.Sleep(500 * time.Millisecond)

	// Submit the form
	err = chromedp.Run(chromedpCtx.Ctx,
		chromedp.Click("button[type='submit']", chromedp.ByQuery),
	)
	if err != nil {
		t.Fatalf("Failed to submit form: %v", err)
	}

	// Wait for success message
	time.Sleep(1 * time.Second)

	// Check for success message
	var successMessage string
	err = chromedp.Run(chromedpCtx.Ctx,
		chromedp.Text(".success-message", &successMessage, chromedp.ByQuery),
	)
	if err != nil {
		t.Fatalf("Failed to get success message: %v", err)
	}

	if successMessage == "" {
		t.Error("Expected success message after form submission")
	}
}
