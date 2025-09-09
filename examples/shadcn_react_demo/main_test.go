//go:build !js && !wasm

package main

import (
	"strings"
	"testing"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/ozanturksever/uiwgo/internal/testhelpers"
)

func TestShadcnReactDemo_BasicRendering(t *testing.T) {
	server := testhelpers.NewViteServer("shadcn_react_demo", "localhost:0")
	if err := server.Start(); err != nil {
		t.Fatalf("Failed to start dev server: %v", err)
	}
	defer server.Stop()

	chromedpCtx := testhelpers.MustNewChromedpContext(testhelpers.ExtendedTimeoutConfig())
	defer chromedpCtx.Cancel()

	var pageTitle, greeting string

	err := chromedp.Run(chromedpCtx.Ctx,
		testhelpers.Actions.NavigateAndWaitForLoad(server.URL(), "body"),
		// Wait longer for WASM to load and React to render
		chromedp.Sleep(3*time.Second),
		// Check if the app container has content
		chromedp.WaitVisible(`#app`, chromedp.ByID),
		// Wait for the page to load and check basic elements
		chromedp.WaitVisible(`text("shadcn/ui + React Demo")`, chromedp.BySearch),
		chromedp.Text(`text("shadcn/ui + React Demo")`, &pageTitle, chromedp.BySearch),
		chromedp.WaitVisible(`text("Hello, World!")`, chromedp.BySearch),
		chromedp.Text(`text("Hello, World!")`, &greeting, chromedp.BySearch),
	)
	if err != nil {
		t.Fatalf("Test failed: %v", err)
	}

	t.Logf("Page title: '%s', Greeting: '%s'", pageTitle, greeting)

	if !containsText(pageTitle, "shadcn/ui + React Demo") {
		t.Errorf("Expected page title to contain 'shadcn/ui + React Demo', got '%s'", pageTitle)
	}

	if !containsText(greeting, "Hello, World!") {
		t.Errorf("Expected greeting to contain 'Hello, World!', got '%s'", greeting)
	}
}

func TestShadcnReactDemo_ButtonInteraction(t *testing.T) {
	server := testhelpers.NewViteServer("shadcn_react_demo", "localhost:0")
	if err := server.Start(); err != nil {
		t.Fatalf("Failed to start dev server: %v", err)
	}
	defer server.Stop()

	chromedpCtx := testhelpers.MustNewChromedpContext(testhelpers.ExtendedTimeoutConfig())
	defer chromedpCtx.Cancel()

	var initialCount, afterIncrement, afterDecrement, afterReset string

	err := chromedp.Run(chromedpCtx.Ctx,
		testhelpers.Actions.NavigateAndWaitForLoad(server.URL(), "body"),
		// Wait longer for WASM to load and React to render
		chromedp.Sleep(3*time.Second),
		// Check if the app container has content
		chromedp.WaitVisible(`#app`, chromedp.ByID),
		// Wait for the counter value element specifically
		chromedp.WaitVisible(`#counter-value`, chromedp.ByID),
		chromedp.Text(`#counter-value`, &initialCount, chromedp.ByID),
		// Click increment button
		chromedp.WaitVisible(`#increment-btn`, chromedp.ByID),
		testhelpers.Actions.ClickAndWait(`#increment-btn`, 200*time.Millisecond),
		chromedp.Text(`#counter-value`, &afterIncrement, chromedp.ByID),
		// Click decrement button
		chromedp.WaitVisible(`#decrement-btn`, chromedp.ByID),
		testhelpers.Actions.ClickAndWait(`#decrement-btn`, 200*time.Millisecond),
		chromedp.Text(`#counter-value`, &afterDecrement, chromedp.ByID),
		// Click reset button
		chromedp.WaitVisible(`#reset-btn`, chromedp.ByID),
		testhelpers.Actions.ClickAndWait(`#reset-btn`, 200*time.Millisecond),
		chromedp.Text(`#counter-value`, &afterReset, chromedp.ByID),
	)
	if err != nil {
		t.Fatalf("Test failed: %v", err)
	}

	t.Logf("Count progression: initial='%s', after increment='%s', after decrement='%s', after reset='%s'",
		initialCount, afterIncrement, afterDecrement, afterReset)

	if initialCount != "0" {
		t.Errorf("Expected initial count to be 0, got '%s'", initialCount)
	}

	if afterIncrement != "1" {
		t.Error("Count should have increased after clicking increment")
	}

	if afterDecrement != "0" {
		t.Error("Count should have decreased after clicking decrement")
	}

	if afterReset != "0" {
		t.Errorf("Count should be 0 after reset, got '%s'", afterReset)
	}
}

func TestShadcnReactDemo_InputFunctionality(t *testing.T) {
	server := testhelpers.NewViteServer("shadcn_react_demo", "localhost:0")
	if err := server.Start(); err != nil {
		t.Fatalf("Failed to start dev server: %v", err)
	}
	defer server.Stop()

	chromedpCtx := testhelpers.MustNewChromedpContext(testhelpers.ExtendedTimeoutConfig())
	defer chromedpCtx.Cancel()

	var initialGreeting, afterInput string
	testName := "Alice"

	err := chromedp.Run(chromedpCtx.Ctx,
		testhelpers.Actions.NavigateAndWaitForLoad(server.URL(), "body"),
		testhelpers.Actions.WaitForWASMInit(`text("Hello,")`, 500*time.Millisecond),
		chromedp.WaitVisible(`text("Hello,")`, chromedp.BySearch),
		chromedp.Text(`text("Hello,")/../..`, &initialGreeting, chromedp.BySearch),
		// Clear and type in the name input
		chromedp.WaitVisible(`input[placeholder="Enter your name"]`, chromedp.ByQuery),
		chromedp.Clear(`input[placeholder="Enter your name"]`, chromedp.ByQuery),
		testhelpers.Actions.SendKeysAndWait(`input[placeholder="Enter your name"]`, testName, 200*time.Millisecond),
		chromedp.Text(`text("Hello,")/../..`, &afterInput, chromedp.BySearch),
	)
	if err != nil {
		t.Fatalf("Test failed: %v", err)
	}

	t.Logf("Greeting changed from '%s' to '%s'", initialGreeting, afterInput)

	if !containsText(initialGreeting, "Hello, World!") {
		t.Errorf("Expected initial greeting to contain 'Hello, World!', got '%s'", initialGreeting)
	}

	if !containsText(afterInput, "Hello, "+testName+"!") {
		t.Errorf("Expected greeting to contain 'Hello, %s!', got '%s'", testName, afterInput)
	}
}

func TestShadcnReactDemo_TodoManagement(t *testing.T) {
	server := testhelpers.NewViteServer("shadcn_react_demo", "localhost:0")
	if err := server.Start(); err != nil {
		t.Fatalf("Failed to start dev server: %v", err)
	}
	defer server.Stop()

	chromedpCtx := testhelpers.MustNewChromedpContext(testhelpers.DefaultConfig())
	defer chromedpCtx.Cancel()

	todoText := "Test todo item"
	var emptyMessage, afterAdd string

	err := chromedp.Run(chromedpCtx.Ctx,
		testhelpers.Actions.NavigateAndWaitForLoad(server.URL(), "body"),
		testhelpers.Actions.WaitForWASMInit(`text("No todos yet")`, 2*time.Second),
		// Check empty state
		chromedp.WaitVisible(`text("No todos yet")`, chromedp.BySearch),
		chromedp.Text(`text("No todos yet")/../..`, &emptyMessage, chromedp.BySearch),
		// Add a todo
		chromedp.WaitVisible(`input[placeholder="Add a new todo"]`, chromedp.ByQuery),
		testhelpers.Actions.SendKeysAndWait(`input[placeholder="Add a new todo"]`, todoText, 100*time.Millisecond),
		testhelpers.Actions.ClickAndWait(`button:contains("Add")`, 200*time.Millisecond),
		// Check if todo was added
		chromedp.WaitVisible(`text("`+todoText+`")`, chromedp.BySearch),
		chromedp.Text(`text("`+todoText+`")/../..`, &afterAdd, chromedp.BySearch),
	)
	if err != nil {
		t.Fatalf("Test failed: %v", err)
	}

	t.Logf("Todo state: empty='%s', after add='%s'", emptyMessage, afterAdd)

	if !containsText(emptyMessage, "No todos yet") {
		t.Errorf("Expected empty state message, got '%s'", emptyMessage)
	}

	if !containsText(afterAdd, todoText) {
		t.Errorf("Expected todo text '%s' to be present, got '%s'", todoText, afterAdd)
	}
}

func TestShadcnReactDemo_ThemeToggle(t *testing.T) {
	server := testhelpers.NewViteServer("shadcn_react_demo", "localhost:0")
	if err := server.Start(); err != nil {
		t.Fatalf("Failed to start dev server: %v", err)
	}
	defer server.Stop()

	chromedpCtx := testhelpers.MustNewChromedpContext(testhelpers.ExtendedTimeoutConfig())
	defer chromedpCtx.Cancel()

	var initialTheme, afterToggle, htmlClass string

	err := chromedp.Run(chromedpCtx.Ctx,
		testhelpers.Actions.NavigateAndWaitForLoad(server.URL(), "body"),
		testhelpers.Actions.WaitForWASMInit(`text("Theme:")`, 2*time.Second),
		// Check initial theme
		chromedp.WaitVisible(`text("Theme:")`, chromedp.BySearch),
		chromedp.Text(`text("Theme:")/../..`, &initialTheme, chromedp.BySearch),
		// Toggle theme
		chromedp.WaitVisible(`button:contains("Dark"), button:contains("Light")`, chromedp.BySearch),
		testhelpers.Actions.ClickAndWait(`button:contains("Dark"), button:contains("Light")`, 200*time.Millisecond),
		chromedp.Text(`text("Theme:")/../..`, &afterToggle, chromedp.BySearch),
		// Check HTML class for theme
		chromedp.AttributeValue(`html`, "class", &htmlClass, nil),
	)
	if err != nil {
		t.Fatalf("Test failed: %v", err)
	}

	t.Logf("Theme state: initial='%s', after toggle='%s', html class='%s'", initialTheme, afterToggle, htmlClass)

	if !containsText(initialTheme, "Theme: light") {
		t.Errorf("Expected initial theme to be light, got '%s'", initialTheme)
	}

	if containsText(afterToggle, "Theme: light") {
		t.Errorf("Expected theme to change after toggle, but still got '%s'", afterToggle)
	}
}

func TestShadcnReactDemo_ComponentIntegration(t *testing.T) {
	server := testhelpers.NewViteServer("shadcn_react_demo", "localhost:0")
	if err := server.Start(); err != nil {
		t.Fatalf("Failed to start dev server: %v", err)
	}
	defer server.Stop()

	chromedpCtx := testhelpers.MustNewChromedpContext(testhelpers.ExtendedTimeoutConfig())
	defer chromedpCtx.Cancel()

	var cardTitle, badgeText string

	err := chromedp.Run(chromedpCtx.Ctx,
		testhelpers.Actions.NavigateAndWaitForLoad(server.URL(), "body"),
		testhelpers.Actions.WaitForWASMInit(`text("Card Demo")`, 2*time.Second),
		// Check for shadcn components
		chromedp.WaitVisible(`text("Card Demo")`, chromedp.BySearch),
		chromedp.Text(`text("Card Demo")`, &cardTitle, chromedp.BySearch),
		chromedp.WaitVisible(`text("New")`, chromedp.BySearch),
		chromedp.Text(`text("New")`, &badgeText, chromedp.BySearch),
	)
	if err != nil {
		t.Fatalf("Test failed: %v", err)
	}

	t.Logf("Card title: '%s', Badge text: '%s'", cardTitle, badgeText)

	if !containsText(cardTitle, "Card Demo") {
		t.Errorf("Expected card title 'Card Demo', got '%s'", cardTitle)
	}

	if !containsText(badgeText, "New") {
		t.Errorf("Expected badge text 'New', got '%s'", badgeText)
	}
}

func TestShadcnReactDemo_ImmediateReactivity(t *testing.T) {
	server := testhelpers.NewViteServer("shadcn_react_demo", "localhost:0")
	if err := server.Start(); err != nil {
		t.Fatalf("Failed to start dev server: %v", err)
	}
	defer server.Stop()

	chromedpCtx := testhelpers.MustNewChromedpContext(testhelpers.ExtendedTimeoutConfig())
	defer chromedpCtx.Cancel()

	var initialCount, afterIncrement, afterTodoAdd, afterInput string
	todoText := "Immediate reactivity test"
	testName := "ReactivityTest"

	err := chromedp.Run(chromedpCtx.Ctx,
		testhelpers.Actions.NavigateAndWaitForLoad(server.URL(), "body"),
		// Wait longer for WASM to load and React to render
		chromedp.Sleep(3*time.Second),
		// Check if the app container has content
		chromedp.WaitVisible(`#app`, chromedp.ByID),
		
		// Test 1: Counter reactivity works immediately
		chromedp.WaitVisible(`#counter-value`, chromedp.ByID),
		chromedp.Text(`#counter-value`, &initialCount, chromedp.ByID),
		chromedp.Click(`#increment-btn`, chromedp.ByID),
		// Wait briefly for reactivity to update
		chromedp.Sleep(100*time.Millisecond),
		chromedp.Text(`#counter-value`, &afterIncrement, chromedp.ByID),
		
		// Test 2: Todo addition works immediately
		chromedp.WaitVisible(`#new-todo-input`, chromedp.ByID),
		chromedp.SendKeys(`#new-todo-input`, todoText, chromedp.ByID),
		chromedp.Click(`#add-todo-btn`, chromedp.ByID),
		// Wait briefly for reactivity to update
		chromedp.Sleep(100*time.Millisecond),
		chromedp.WaitVisible(`text("`+todoText+`")`, chromedp.BySearch),
		chromedp.Text(`#todo-list`, &afterTodoAdd, chromedp.ByID),
		
		// Test 3: Input field reactivity works immediately
		chromedp.WaitVisible(`#name`, chromedp.ByID),
		chromedp.Clear(`#name`, chromedp.ByID),
		chromedp.SendKeys(`#name`, testName, chromedp.ByID),
		// Wait briefly for reactivity to update
		chromedp.Sleep(100*time.Millisecond),
		chromedp.WaitVisible(`text("Hello, `+testName+`!")`, chromedp.BySearch),
		chromedp.Text(`text("Hello, `+testName+`!")`, &afterInput, chromedp.BySearch),
	)
	if err != nil {
		t.Fatalf("Test failed: %v", err)
	}

	// Verify counter reactivity
	if initialCount == "" {
		t.Error("Initial count should not be empty")
	}
	if afterIncrement == "" {
		t.Error("Counter should update immediately after increment")
	}
	if initialCount == afterIncrement {
		t.Errorf("Counter should have changed from '%s' to '%s'", initialCount, afterIncrement)
	}

	// Verify todo reactivity
	if !strings.Contains(afterTodoAdd, todoText) {
		t.Errorf("Todo list should contain '%s' immediately after adding, got: %s", todoText, afterTodoAdd)
	}

	// Verify input reactivity
	if !strings.Contains(afterInput, testName) {
		t.Errorf("Greeting should contain '%s' immediately after input change, got: %s", testName, afterInput)
	}

	t.Logf("Reactivity test passed - Counter: %s->%s, Todo added: %v, Input updated: %v", 
		initialCount, afterIncrement, strings.Contains(afterTodoAdd, todoText), strings.Contains(afterInput, testName))
}

func containsText(text, substring string) bool {
	return strings.Contains(strings.ToLower(text), strings.ToLower(substring))
}

func findSubstring(text, substring string) bool {
	return strings.Contains(text, substring)
}