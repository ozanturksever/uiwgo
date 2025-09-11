//go:build !js && !wasm

package main

import (
	"testing"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/ozanturksever/uiwgo/internal/testhelpers"
)

func TestFormDemo_PageRender(t *testing.T) {
	server := testhelpers.NewViteServer("form_demo", "localhost:0")
	if err := server.Start(); err != nil {
		t.Fatalf("Failed to start vite server: %v", err)
	}
	defer server.Stop()

	chromedpCtx := testhelpers.MustNewChromedpContext(testhelpers.DefaultConfig())
	defer chromedpCtx.Cancel()

	var title string
	err := chromedp.Run(chromedpCtx.Ctx,
		testhelpers.Actions.NavigateAndWaitForLoad(server.URL(), "#registration-form"),
		chromedp.Title(&title),
	)
	if err != nil {
		t.Fatalf("Failed to load page: %v", err)
	}

	if title != "Form Demo - UIWGO" {
		t.Errorf("Expected title 'Form Demo - UIWGO', got '%s'", title)
	}
}

func TestFormDemo_FormFieldsPresent(t *testing.T) {
	server := testhelpers.NewViteServer("form_demo", "localhost:0")
	if err := server.Start(); err != nil {
		t.Fatalf("Failed to start vite server: %v", err)
	}
	defer server.Stop()

	chromedpCtx := testhelpers.MustNewChromedpContext(testhelpers.DefaultConfig())
	defer chromedpCtx.Cancel()

	err := chromedp.Run(chromedpCtx.Ctx,
		testhelpers.Actions.NavigateAndWaitForLoad(server.URL(), "#registration-form"),
		
		// Check that all form fields are present
		chromedp.WaitVisible(`input[name="name"]`, chromedp.ByQuery),
		chromedp.WaitVisible(`input[name="email"]`, chromedp.ByQuery),
		chromedp.WaitVisible(`input[name="password"]`, chromedp.ByQuery),
		chromedp.WaitVisible(`input[name="confirm_password"]`, chromedp.ByQuery),
		chromedp.WaitVisible(`textarea[name="bio"]`, chromedp.ByQuery),
		chromedp.WaitVisible(`#submit-btn`, chromedp.ByID),
	)
	if err != nil {
		t.Fatalf("Form fields not found: %v", err)
	}
}

func TestFormDemo_FormInputAndValidation(t *testing.T) {
	server := testhelpers.NewViteServer("form_demo", "localhost:0")
	if err := server.Start(); err != nil {
		t.Fatalf("Failed to start vite server: %v", err)
	}
	defer server.Stop()

	chromedpCtx := testhelpers.MustNewChromedpContext(testhelpers.DefaultConfig())
	defer chromedpCtx.Cancel()

	err := chromedp.Run(chromedpCtx.Ctx,
		testhelpers.Actions.NavigateAndWaitForLoad(server.URL(), "#registration-form"),
		
		// Fill out the form with valid data
		testhelpers.Actions.SendKeysAndWait(`input[name="name"]`, "John Doe", 100*time.Millisecond),
		testhelpers.Actions.SendKeysAndWait(`input[name="email"]`, "john.doe@example.com", 100*time.Millisecond),
		testhelpers.Actions.SendKeysAndWait(`input[name="password"]`, "password123", 100*time.Millisecond),
		testhelpers.Actions.SendKeysAndWait(`input[name="confirm_password"]`, "password123", 100*time.Millisecond),
		testhelpers.Actions.SendKeysAndWait(`textarea[name="bio"]`, "I am a software developer.", 100*time.Millisecond),
		
		// Wait for form state to update
		chromedp.Sleep(500*time.Millisecond),
		
		// Check that form state debug shows the entered values
		chromedp.WaitVisible(`#form-state-debug`, chromedp.ByID),
	)
	if err != nil {
		t.Fatalf("Failed to fill form: %v", err)
	}

	// Verify form state contains the entered data
	var debugText string
	err = chromedp.Run(chromedpCtx.Ctx,
		chromedp.Text(`#form-state-debug`, &debugText, chromedp.ByID),
	)
	if err != nil {
		t.Fatalf("Failed to get debug text: %v", err)
	}

	// Check that the debug text contains our form data
	expectedValues := []string{"John Doe", "john.doe@example.com", "password123", "I am a software developer."}
	for _, value := range expectedValues {
		if !contains(debugText, value) {
			t.Errorf("Expected form state to contain '%s', but it didn't. Debug text: %s", value, debugText)
		}
	}
}

func TestFormDemo_FormSubmission(t *testing.T) {
	server := testhelpers.NewViteServer("form_demo", "localhost:0")
	if err := server.Start(); err != nil {
		t.Fatalf("Failed to start vite server: %v", err)
	}
	defer server.Stop()

	chromedpCtx := testhelpers.MustNewChromedpContext(testhelpers.VisibleConfig()) // Use visible for alert handling
	defer chromedpCtx.Cancel()

	err := chromedp.Run(chromedpCtx.Ctx,
		testhelpers.Actions.NavigateAndWaitForLoad(server.URL(), "#registration-form"),
		
		// Fill out the form with valid data
		testhelpers.Actions.SendKeysAndWait(`input[name="name"]`, "Jane Smith", 100*time.Millisecond),
		testhelpers.Actions.SendKeysAndWait(`input[name="email"]`, "jane.smith@example.com", 100*time.Millisecond),
		testhelpers.Actions.SendKeysAndWait(`input[name="password"]`, "securepass123", 100*time.Millisecond),
		testhelpers.Actions.SendKeysAndWait(`input[name="confirm_password"]`, "securepass123", 100*time.Millisecond),
		testhelpers.Actions.SendKeysAndWait(`textarea[name="bio"]`, "I love building web applications.", 100*time.Millisecond),
		
		// Submit the form
		testhelpers.Actions.ClickAndWait(`#submit-btn`, 100*time.Millisecond),
		
		// Wait for alert and accept it
		chromedp.Sleep(1*time.Second),
	)
	if err != nil {
		t.Fatalf("Failed to submit form: %v", err)
	}
}

func TestFormDemo_PasswordMismatchValidation(t *testing.T) {
	server := testhelpers.NewViteServer("form_demo", "localhost:0")
	if err := server.Start(); err != nil {
		t.Fatalf("Failed to start vite server: %v", err)
	}
	defer server.Stop()

	chromedpCtx := testhelpers.MustNewChromedpContext(testhelpers.VisibleConfig()) // Use visible for alert handling
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
		
		// Wait for alert and accept it
		chromedp.Sleep(1*time.Second),
	)
	if err != nil {
		t.Fatalf("Failed to test password mismatch: %v", err)
	}
}

func TestFormDemo_FormReset(t *testing.T) {
	server := testhelpers.NewViteServer("form_demo", "localhost:0")
	if err := server.Start(); err != nil {
		t.Fatalf("Failed to start vite server: %v", err)
	}
	defer server.Stop()

	chromedpCtx := testhelpers.MustNewChromedpContext(testhelpers.DefaultConfig())
	defer chromedpCtx.Cancel()

	err := chromedp.Run(chromedpCtx.Ctx,
		testhelpers.Actions.NavigateAndWaitForLoad(server.URL(), "#registration-form"),
		
		// Fill out some fields
		testhelpers.Actions.SendKeysAndWait(`input[name="name"]`, "Test Name", 100*time.Millisecond),
		testhelpers.Actions.SendKeysAndWait(`input[name="email"]`, "test@example.com", 100*time.Millisecond),
		
		// Wait for form state to update
		chromedp.Sleep(500*time.Millisecond),
	)
	if err != nil {
		t.Fatalf("Failed to fill form fields: %v", err)
	}

	// Verify that form state shows the entered data
	var debugTextBefore string
	err = chromedp.Run(chromedpCtx.Ctx,
		chromedp.Text(`#form-state-debug`, &debugTextBefore, chromedp.ByID),
	)
	if err != nil {
		t.Fatalf("Failed to get debug text before reset: %v", err)
	}

	if !contains(debugTextBefore, "Test Name") {
		t.Errorf("Expected form state to contain 'Test Name' before reset")
	}

	// Reset the form by executing JavaScript
	err = chromedp.Run(chromedpCtx.Ctx,
		chromedp.Evaluate(`document.getElementById('registration-form').reset()`, nil),
		chromedp.Sleep(500*time.Millisecond),
	)
	if err != nil {
		t.Fatalf("Failed to reset form: %v", err)
	}

	// Verify that form state is cleared
	var debugTextAfter string
	err = chromedp.Run(chromedpCtx.Ctx,
		chromedp.Text(`#form-state-debug`, &debugTextAfter, chromedp.ByID),
	)
	if err != nil {
		t.Fatalf("Failed to get debug text after reset: %v", err)
	}

	if contains(debugTextAfter, "Test Name") {
		t.Errorf("Expected form state to be cleared after reset, but still contains 'Test Name'")
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 || 
		(len(s) > len(substr) && findSubstring(s, substr)))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}