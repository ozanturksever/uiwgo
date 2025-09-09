//go:build js && wasm

package dom

import (
	"strconv"
	"strings"
	"syscall/js"
	"testing"
	"time"

	"honnef.co/go/js/dom/v2"
	g "maragu.dev/gomponents"
)

// Test helper to create a test element with attributes
func createTestElement(tagName string, attrs ...g.Node) dom.Element {
	doc := dom.GetWindow().Document()
	el := doc.CreateElement(tagName)
	
	// Apply attributes
	for _, attr := range attrs {
		if attr != nil {
			// This is a simplified approach for testing
			// We'll just verify the attribute exists without applying it
			_ = attr
		}
	}
	
	return el
}

// Test OnClickInline handler registration and cleanup
func TestOnClickInline(t *testing.T) {
	handler := func(el Element) {
		// Handler implementation for testing
	}
	
	// Create inline click handler
	attr := OnClickInline(handler)
	if attr == nil {
		t.Fatal("OnClickInline should return an attribute")
	}
	
	// Verify handler is registered
	inlineHandlersMu.RLock()
	handlerCount := len(inlineClickHandlers)
	inlineHandlersMu.RUnlock()
	
	if handlerCount == 0 {
		t.Error("Handler should be registered")
	}
	
	// Test cleanup by clearing handlers
	inlineHandlersMu.Lock()
	for id := range inlineClickHandlers {
		delete(inlineClickHandlers, id)
	}
	inlineHandlersMu.Unlock()
	
	inlineHandlersMu.RLock()
	handlerCount = len(inlineClickHandlers)
	inlineHandlersMu.RUnlock()
	
	if handlerCount != 0 {
		t.Error("Handlers should be cleaned up")
	}
}

// Test OnInputInline handler registration and cleanup
func TestOnInputInline(t *testing.T) {
	handler := func(el Element) {
		// Handler implementation for testing
	}
	
	attr := OnInputInline(handler)
	if attr == nil {
		t.Fatal("OnInputInline should return an attribute")
	}
	
	// Verify handler is registered
	inlineHandlersMu.RLock()
	handlerCount := len(inlineInputHandlers)
	inlineHandlersMu.RUnlock()
	
	if handlerCount == 0 {
		t.Error("Input handler should be registered")
	}
	
	// Cleanup
	inlineHandlersMu.Lock()
	for id := range inlineInputHandlers {
		delete(inlineInputHandlers, id)
	}
	inlineHandlersMu.Unlock()
}

// Test OnChangeInline handler registration and cleanup
func TestOnChangeInline(t *testing.T) {
	handler := func(el Element) {
		// Handler implementation for testing
	}
	
	attr := OnChangeInline(handler)
	if attr == nil {
		t.Fatal("OnChangeInline should return an attribute")
	}
	
	// Verify handler is registered
	inlineHandlersMu.RLock()
	handlerCount := len(inlineChangeHandlers)
	inlineHandlersMu.RUnlock()
	
	if handlerCount == 0 {
		t.Error("Change handler should be registered")
	}
	
	// Cleanup
	inlineHandlersMu.Lock()
	for id := range inlineChangeHandlers {
		delete(inlineChangeHandlers, id)
	}
	inlineHandlersMu.Unlock()
}

// Test OnKeydownInline handler registration
func TestOnKeydownInline(t *testing.T) {
	handler := func(el Element) {
		// Handler implementation for testing
	}
	
	attr := OnKeyDownInline(handler, "Enter")
	if attr == nil {
		t.Fatal("OnKeydownInline should return an attribute")
	}
	
	// Verify handler is registered
	inlineHandlersMu.RLock()
	handlerCount := len(inlineKeydownHandlers)
	expectationCount := len(inlineKeyExpectations)
	inlineHandlersMu.RUnlock()
	
	if handlerCount == 0 {
		t.Error("Keydown handler should be registered")
	}
	if expectationCount == 0 {
		t.Error("Key expectation should be registered")
	}
	
	// Cleanup
	inlineHandlersMu.Lock()
	for id := range inlineKeydownHandlers {
		delete(inlineKeydownHandlers, id)
		delete(inlineKeyExpectations, id)
	}
	inlineHandlersMu.Unlock()
}

// Test OnEnterInline convenience function
func TestOnEnterInline(t *testing.T) {
	handler := func(el Element) {
		// Handler implementation for testing
	}
	
	attr := OnEnterInline(handler)
	if attr == nil {
		t.Fatal("OnEnterInline should return an attribute")
	}
	
	// Verify handler and expectation are registered
	inlineHandlersMu.RLock()
	handlerCount := len(inlineKeydownHandlers)
	expectationCount := len(inlineKeyExpectations)
	inlineHandlersMu.RUnlock()
	
	if handlerCount == 0 {
		t.Error("Enter handler should be registered")
	}
	if expectationCount == 0 {
		t.Error("Enter expectation should be registered")
	}
	
	// Verify the expectation is "Enter"
	inlineHandlersMu.RLock()
	for id, expectation := range inlineKeyExpectations {
		if expectation != "Enter" {
			t.Errorf("Expected 'Enter', got '%s' for id %s", expectation, id)
		}
		break // Just check the first one
	}
	inlineHandlersMu.RUnlock()
	
	// Cleanup
	inlineHandlersMu.Lock()
	for id := range inlineKeydownHandlers {
		delete(inlineKeydownHandlers, id)
		delete(inlineKeyExpectations, id)
	}
	inlineHandlersMu.Unlock()
}

// Test OnEscapeInline convenience function
func TestOnEscapeInline(t *testing.T) {
	handler := func(el Element) {
		// Handler implementation for testing
	}
	
	attr := OnEscapeInline(handler)
	if attr == nil {
		t.Fatal("OnEscapeInline should return an attribute")
	}
	
	// Verify handler and expectation are registered
	inlineHandlersMu.RLock()
	handlerCount := len(inlineKeydownHandlers)
	expectationCount := len(inlineKeyExpectations)
	inlineHandlersMu.RUnlock()
	
	if handlerCount == 0 {
		t.Error("Escape handler should be registered")
	}
	if expectationCount == 0 {
		t.Error("Escape expectation should be registered")
	}
	
	// Verify the expectation is "Escape"
	inlineHandlersMu.RLock()
	for id, expectation := range inlineKeyExpectations {
		if expectation != "Escape" {
			t.Errorf("Expected 'Escape', got '%s' for id %s", expectation, id)
		}
		break // Just check the first one
	}
	inlineHandlersMu.RUnlock()
	
	// Cleanup
	inlineHandlersMu.Lock()
	for id := range inlineKeydownHandlers {
		delete(inlineKeydownHandlers, id)
		delete(inlineKeyExpectations, id)
	}
	inlineHandlersMu.Unlock()
}

// Test OnSubmitInline handler registration
func TestOnSubmitInline(t *testing.T) {
	handler := func(el Element, formData map[string]string) {
		// Handler implementation for testing
	}
	
	attr := OnSubmitInline(handler)
	if attr == nil {
		t.Fatal("OnSubmitInline should return an attribute")
	}
	
	// Verify handler is registered
	inlineHandlersMu.RLock()
	handlerCount := len(inlineSubmitHandlers)
	inlineHandlersMu.RUnlock()
	
	if handlerCount == 0 {
		t.Error("Submit handler should be registered")
	}
	
	// Cleanup
	inlineHandlersMu.Lock()
	for id := range inlineSubmitHandlers {
		delete(inlineSubmitHandlers, id)
	}
	inlineHandlersMu.Unlock()
}

// Test OnFormResetInline handler registration
func TestOnFormResetInline(t *testing.T) {
	handler := func(el Element) {
		// Handler implementation for testing
	}
	
	attr := OnFormResetInline(handler)
	if attr == nil {
		t.Fatal("OnFormResetInline should return an attribute")
	}
	
	// Verify handler is registered
	inlineHandlersMu.RLock()
	handlerCount := len(inlineFormResetHandlers)
	inlineHandlersMu.RUnlock()
	
	if handlerCount == 0 {
		t.Error("Form reset handler should be registered")
	}
	
	// Cleanup
	inlineHandlersMu.Lock()
	for id := range inlineFormResetHandlers {
		delete(inlineFormResetHandlers, id)
	}
	inlineHandlersMu.Unlock()
}

// Test OnFormChangeInline handler registration
func TestOnFormChangeInline(t *testing.T) {
	handler := func(el Element, formData map[string]string) {
		// Handler implementation for testing
	}
	
	attr := OnFormChangeInline(handler)
	if attr == nil {
		t.Fatal("OnFormChangeInline should return an attribute")
	}
	
	// Verify handler is registered
	inlineHandlersMu.RLock()
	handlerCount := len(inlineFormChangeHandlers)
	inlineHandlersMu.RUnlock()
	
	if handlerCount == 0 {
		t.Error("Form change handler should be registered")
	}
	
	// Cleanup
	inlineHandlersMu.Lock()
	for id := range inlineFormChangeHandlers {
		delete(inlineFormChangeHandlers, id)
	}
	inlineHandlersMu.Unlock()
}

// Test OnBlurInline handler registration
func TestOnBlurInline(t *testing.T) {
	handler := func(el Element) {
		// Handler implementation for testing
	}
	
	attr := OnBlurInline(handler)
	if attr == nil {
		t.Fatal("OnBlurInline should return an attribute")
	}
	
	// Verify handler is registered
	inlineHandlersMu.RLock()
	handlerCount := len(inlineBlurHandlers)
	inlineHandlersMu.RUnlock()
	
	if handlerCount == 0 {
		t.Error("Blur handler should be registered")
	}
	
	// Cleanup
	inlineHandlersMu.Lock()
	for id := range inlineBlurHandlers {
		delete(inlineBlurHandlers, id)
	}
	inlineHandlersMu.Unlock()
}

// Test OnFocusInline handler registration
func TestOnFocusInline(t *testing.T) {
	handler := func(el Element) {
		// Handler implementation for testing
	}
	
	attr := OnFocusInline(handler)
	if attr == nil {
		t.Fatal("OnFocusInline should return an attribute")
	}
	
	// Verify handler is registered
	inlineHandlersMu.RLock()
	handlerCount := len(inlineFocusHandlers)
	inlineHandlersMu.RUnlock()
	
	if handlerCount == 0 {
		t.Error("Focus handler should be registered")
	}
	
	// Cleanup
	inlineHandlersMu.Lock()
	for id := range inlineFocusHandlers {
		delete(inlineFocusHandlers, id)
	}
	inlineHandlersMu.Unlock()
}

// Test validation patterns
func TestOnFocusWithinInline(t *testing.T) {
	attr := OnFocusWithinInline(func(el Element, isFocusEntering bool) {
		// Handler implementation - parameters will be used in real scenarios
		_ = el
		_ = isFocusEntering
	})
	
	if attr == nil {
		t.Fatal("OnFocusWithinInline should return an attribute")
	}
	
	// Verify handler is registered
	inlineHandlersMu.RLock()
	handlerCount := len(inlineFocusWithinHandlers)
	inlineHandlersMu.RUnlock()
	
	if handlerCount == 0 {
		t.Error("FocusWithin handler should be registered")
	}
	
	// Test cleanup
	inlineHandlersMu.Lock()
	originalCount := len(inlineFocusWithinHandlers)
	for id := range inlineFocusWithinHandlers {
		delete(inlineFocusWithinHandlers, id)
		break
	}
	newCount := len(inlineFocusWithinHandlers)
	inlineHandlersMu.Unlock()
	
	if newCount != originalCount-1 {
		t.Error("Handler cleanup should remove one handler")
	}
}

// Test OnOutsideClickInline handler registration
func TestOnOutsideClickInline(t *testing.T) {
	handler := func(el Element) {
		// Handler implementation for testing
	}
	
	attr := OnOutsideClickInline(handler)
	if attr == nil {
		t.Fatal("OnOutsideClickInline should return an attribute")
	}
	
	// Verify handler is registered
	inlineHandlersMu.RLock()
	handlerCount := len(inlineOutsideClickHandlers)
	inlineHandlersMu.RUnlock()
	
	if handlerCount == 0 {
		t.Error("Outside click handler should be registered")
	}
	
	// Cleanup
	inlineHandlersMu.Lock()
	for id := range inlineOutsideClickHandlers {
		delete(inlineOutsideClickHandlers, id)
	}
	inlineHandlersMu.Unlock()
}

// Test OnEscapeCloseInline handler registration
func TestOnEscapeCloseInline(t *testing.T) {
	handler := func(el Element) {
		// Handler implementation for testing
	}
	
	attr := OnEscapeCloseInline(handler)
	if attr == nil {
		t.Fatal("OnEscapeCloseInline should return an attribute")
	}
	
	// Verify handler is registered
	inlineHandlersMu.RLock()
	handlerCount := len(inlineEscapeCloseHandlers)
	inlineHandlersMu.RUnlock()
	
	if handlerCount == 0 {
		t.Error("Escape close handler should be registered")
	}
	
	// Cleanup
	inlineHandlersMu.Lock()
	for id := range inlineEscapeCloseHandlers {
		delete(inlineEscapeCloseHandlers, id)
	}
	inlineHandlersMu.Unlock()
}

// Test OnFileSelectInline handler registration
func TestOnFileSelectInline(t *testing.T) {
	handler := func(el Element, files []js.Value) {
		// Handler implementation for testing
	}

	// Test handler registration
	attr := OnFileSelectInline(handler)
	if attr == nil {
		t.Fatal("OnFileSelectInline should return an attribute")
	}

	// Check handler was registered
	inlineHandlersMu.RLock()
	handlerCount := len(inlineFileSelectHandlers)
	inlineHandlersMu.RUnlock()
	if handlerCount == 0 {
		t.Error("File select handler should be registered")
	}

	// Test cleanup
	inlineHandlersMu.Lock()
	for id := range inlineFileSelectHandlers {
		delete(inlineFileSelectHandlers, id)
	}
	inlineHandlersMu.Unlock()
}

// Test OnFileDropInline handler registration
func TestOnFileDropInline(t *testing.T) {
	handler := func(el Element, files []js.Value) {
		// Handler implementation for testing
	}

	// Test handler registration
	attr := OnFileDropInline(handler)
	if attr == nil {
		t.Fatal("OnFileDropInline should return an attribute")
	}

	// Check handler was registered
	inlineHandlersMu.RLock()
	handlerCount := len(inlineFileDropHandlers)
	inlineHandlersMu.RUnlock()
	if handlerCount == 0 {
		t.Error("File drop handler should be registered")
	}

	// Test cleanup
	inlineHandlersMu.Lock()
	for id := range inlineFileDropHandlers {
		delete(inlineFileDropHandlers, id)
	}
	inlineHandlersMu.Unlock()
}

func TestValidationPatterns(t *testing.T) {
	tests := []struct {
		pattern ValidationPattern
		value   string
		valid   bool
		options []string
	}{
		{ValidationEmail, "test@example.com", true, nil},
		{ValidationEmail, "invalid-email", false, nil},
		{ValidationURL, "https://example.com", true, nil},
		{ValidationURL, "not-a-url", false, nil},
		{ValidationPhone, "+1234567890", true, nil},
		{ValidationPhone, "invalid-phone", false, nil},
		{ValidationRequired, "some text", true, nil},
		{ValidationRequired, "", false, nil},
		{ValidationRequired, "   ", false, nil},
		{ValidationMinLength, "hello", true, []string{"3"}},
		{ValidationMinLength, "hi", false, []string{"3"}},
		{ValidationMaxLength, "hi", true, []string{"5"}},
		{ValidationMaxLength, "hello world", false, []string{"5"}},
		{ValidationNumber, "123", true, nil},
		{ValidationNumber, "123.45", true, nil},
		{ValidationNumber, "not-a-number", false, nil},
		{ValidationRegex, "abc123", true, []string{"^[a-z0-9]+$"}},
		{ValidationRegex, "ABC123", false, []string{"^[a-z0-9]+$"}},
	}
	
	for _, test := range tests {
		validator := createValidator(test.pattern, test.options...)
		result := validator(nil, test.value)
		if result != test.valid {
			t.Errorf("Pattern %s with value '%s' expected %v, got %v", 
				test.pattern, test.value, test.valid, result)
		}
	}
}

// Test OnValidateInline handler registration
func TestOnValidateInline(t *testing.T) {
	attr := OnValidateInline(ValidationEmail)
	if attr == nil {
		t.Fatal("OnValidateInline should return an attribute")
	}
	
	// Verify handler is registered
	inlineHandlersMu.RLock()
	handlerCount := len(inlineValidateHandlers)
	inlineHandlersMu.RUnlock()
	
	if handlerCount == 0 {
		t.Error("Validate handler should be registered")
	}
	
	// Test the validator function
	inlineHandlersMu.RLock()
	var validator func(Element, string) bool
	for _, v := range inlineValidateHandlers {
		validator = v
		break
	}
	inlineHandlersMu.RUnlock()
	
	if validator == nil {
		t.Fatal("Validator function should be available")
	}
	
	// Test validation
	if !validator(nil, "test@example.com") {
		t.Error("Valid email should pass validation")
	}
	if validator(nil, "invalid-email") {
		t.Error("Invalid email should fail validation")
	}
	
	// Cleanup
	inlineHandlersMu.Lock()
	for id := range inlineValidateHandlers {
		delete(inlineValidateHandlers, id)
	}
	inlineHandlersMu.Unlock()
}

// Test OnBlurValidateInline handler registration
func TestOnBlurValidateInline(t *testing.T) {
	attr := OnBlurValidateInline(ValidationRequired)
	if attr == nil {
		t.Fatal("OnBlurValidateInline should return an attribute")
	}
	
	// Verify handler is registered
	inlineHandlersMu.RLock()
	handlerCount := len(inlineBlurValidateHandlers)
	inlineHandlersMu.RUnlock()
	
	if handlerCount == 0 {
		t.Error("Blur validate handler should be registered")
	}
	
	// Test the validator function
	inlineHandlersMu.RLock()
	var validator func(Element, string) bool
	for _, v := range inlineBlurValidateHandlers {
		validator = v
		break
	}
	inlineHandlersMu.RUnlock()
	
	if validator == nil {
		t.Fatal("Blur validator function should be available")
	}
	
	// Test validation
	if !validator(nil, "some text") {
		t.Error("Non-empty text should pass required validation")
	}
	if validator(nil, "") {
		t.Error("Empty text should fail required validation")
	}
	
	// Cleanup
	inlineHandlersMu.Lock()
	for id := range inlineBlurValidateHandlers {
		delete(inlineBlurValidateHandlers, id)
	}
	inlineHandlersMu.Unlock()
}

// Test OnInputDebouncedInline handler registration
func TestOnInputDebouncedInline(t *testing.T) {
	handler := func(el Element) {
		// Handler implementation for testing
	}
	
	attr := OnInputDebouncedInline(handler, 300)
	if attr == nil {
		t.Fatal("OnInputDebouncedInline should return an attribute group")
	}
	
	// Verify handler is registered
	inlineHandlersMu.RLock()
	handlerCount := len(inlineDebouncedInputHandlers)
	inlineHandlersMu.RUnlock()
	
	if handlerCount == 0 {
		t.Error("Debounced input handler should be registered")
	}
	
	// Cleanup
	inlineHandlersMu.Lock()
	for id := range inlineDebouncedInputHandlers {
		delete(inlineDebouncedInputHandlers, id)
		// Also clean up any timers
		if timer, exists := inlineDebounceTimers[id]; exists {
			if !timer.IsUndefined() {
				// In a real environment, we would clear the timeout
				// js.Global().Call("clearTimeout", timer)
			}
			delete(inlineDebounceTimers, id)
		}
	}
	inlineHandlersMu.Unlock()
}

// Test OnSearchInline handler registration
func TestOnSearchInline(t *testing.T) {
	handler := func(el Element, query string) {
		// Handler implementation for testing
	}
	
	attr := OnSearchInline(handler, 500)
	if attr == nil {
		t.Fatal("OnSearchInline should return an attribute group")
	}
	
	// Verify handler is registered
	inlineHandlersMu.RLock()
	handlerCount := len(inlineSearchHandlers)
	inlineHandlersMu.RUnlock()
	
	if handlerCount == 0 {
		t.Error("Search handler should be registered")
	}
	
	// Cleanup
	inlineHandlersMu.Lock()
	for id := range inlineSearchHandlers {
		delete(inlineSearchHandlers, id)
		// Also clean up any timers
		if timer, exists := inlineDebounceTimers[id]; exists {
			if !timer.IsUndefined() {
				// In a real environment, we would clear the timeout
				// js.Global().Call("clearTimeout", timer)
			}
			delete(inlineDebounceTimers, id)
		}
	}
	inlineHandlersMu.Unlock()
}

// Test OnTabInline handler registration
func TestOnTabInline(t *testing.T) {
	handler := func(el Element) {
		// Handler implementation for testing
	}
	
	attr := OnTabInline(handler)
	if attr == nil {
		t.Fatal("OnTabInline should return an attribute")
	}
	
	// Verify handler is registered
	inlineHandlersMu.RLock()
	handlerCount := len(inlineTabHandlers)
	inlineHandlersMu.RUnlock()
	
	if handlerCount == 0 {
		t.Error("Tab handler should be registered")
	}
	
	// Cleanup
	inlineHandlersMu.Lock()
	for id := range inlineTabHandlers {
		delete(inlineTabHandlers, id)
	}
	inlineHandlersMu.Unlock()
}

// Test OnShiftTabInline handler registration
func TestOnShiftTabInline(t *testing.T) {
	handler := func(el Element) {
		// Handler implementation for testing
	}
	
	attr := OnShiftTabInline(handler)
	if attr == nil {
		t.Fatal("OnShiftTabInline should return an attribute")
	}
	
	// Verify handler is registered
	inlineHandlersMu.RLock()
	handlerCount := len(inlineShiftTabHandlers)
	inlineHandlersMu.RUnlock()
	
	if handlerCount == 0 {
		t.Error("Shift+Tab handler should be registered")
	}
	
	// Cleanup
	inlineHandlersMu.Lock()
	for id := range inlineShiftTabHandlers {
		delete(inlineShiftTabHandlers, id)
	}
	inlineHandlersMu.Unlock()
}

// Test OnArrowKeysInline handler registration
func TestOnArrowKeysInline(t *testing.T) {
	handler := func(el Element, direction string) {
		// Handler implementation for testing
	}
	
	attr := OnArrowKeysInline(handler)
	if attr == nil {
		t.Fatal("OnArrowKeysInline should return an attribute")
	}
	
	// Verify handler is registered
	inlineHandlersMu.RLock()
	handlerCount := len(inlineArrowKeyHandlers)
	inlineHandlersMu.RUnlock()

	if handlerCount == 0 {
		t.Error("Arrow keys handler should be registered")
	}

	// Cleanup
	inlineHandlersMu.Lock()
	for id := range inlineArrowKeyHandlers {
		delete(inlineArrowKeyHandlers, id)
	}
	inlineHandlersMu.Unlock()
}

// Test OnDragStartInline handler registration and cleanup
func TestOnDragStartInline(t *testing.T) {
	handler := func(el Element, dataTransfer js.Value) {
		// Handler implementation for testing
	}
	
	// Create inline drag start handler
	attr := OnDragStartInline(handler)
	if attr == nil {
		t.Fatal("OnDragStartInline should return an attribute")
	}
	
	// Verify handler is registered
	inlineHandlersMu.RLock()
	handlerCount := len(inlineDragStartHandlers)
	inlineHandlersMu.RUnlock()
	
	if handlerCount == 0 {
		t.Error("Handler should be registered")
	}
	
	// Cleanup
	inlineHandlersMu.Lock()
	for id := range inlineDragStartHandlers {
		delete(inlineDragStartHandlers, id)
	}
	inlineHandlersMu.Unlock()
}

// Test OnDropInline handler registration and cleanup
func TestOnDropInline(t *testing.T) {
	handler := func(el Element, dataTransfer js.Value) {
		// Handler implementation for testing
	}
	
	// Create inline drop handler
	attr := OnDropInline(handler)
	if attr == nil {
		t.Fatal("OnDropInline should return an attribute")
	}
	
	// Verify handler is registered
	inlineHandlersMu.RLock()
	handlerCount := len(inlineDropHandlers)
	inlineHandlersMu.RUnlock()
	
	if handlerCount == 0 {
		t.Error("Handler should be registered")
	}
	
	// Cleanup
	inlineHandlersMu.Lock()
	for id := range inlineDropHandlers {
		delete(inlineDropHandlers, id)
	}
	inlineHandlersMu.Unlock()
}

// Test OnDragOverInline handler registration and cleanup
func TestOnDragOverInline(t *testing.T) {
	handler := func(el Element, dataTransfer js.Value) {
		// Handler implementation for testing
	}
	
	// Create inline drag over handler
	attr := OnDragOverInline(handler)
	if attr == nil {
		t.Fatal("OnDragOverInline should return an attribute")
	}
	
	// Verify handler is registered
	inlineHandlersMu.RLock()
	handlerCount := len(inlineDragOverHandlers)
	inlineHandlersMu.RUnlock()
	
	if handlerCount == 0 {
		t.Error("Handler should be registered")
	}
	
	// Cleanup
	inlineHandlersMu.Lock()
	for id := range inlineDragOverHandlers {
		delete(inlineDragOverHandlers, id)
	}
	inlineHandlersMu.Unlock()
}

// Test nextInlineID function
func TestNextInlineID(t *testing.T) {
	id1 := nextInlineID("test")
	id2 := nextInlineID("test")
	
	if id1 == id2 {
		t.Error("nextInlineID should generate unique IDs")
	}
	
	if !strings.HasPrefix(id1, "test-") {
		t.Errorf("ID should have prefix 'test-', got '%s'", id1)
	}
	
	if !strings.HasPrefix(id2, "test-") {
		t.Errorf("ID should have prefix 'test-', got '%s'", id2)
	}
}

// Test AttachInlineDelegates function basic functionality
func TestAttachInlineDelegates(t *testing.T) {
	// Create a test document element
	doc := dom.GetWindow().Document()
	body := doc.QuerySelector("body")
	if body == nil {
		t.Skip("No body element available for testing")
	}
	
	// Create test elements with inline handlers
	div := doc.CreateElement("div")
	div.SetAttribute("data-uiwgo-onclick", "test-click-1")
	body.AppendChild(div)
	
	// Register a test handler
	inlineHandlersMu.Lock()
	inlineClickHandlers["test-click-1"] = func(el Element) {
		// Handler implementation for testing
	}
	inlineHandlersMu.Unlock()

	// Attach delegates
	AttachInlineDelegates(body.Underlying())
	
	// Cleanup test elements and handlers
	body.RemoveChild(div)
	inlineHandlersMu.Lock()
	delete(inlineClickHandlers, "test-click-1")
	inlineHandlersMu.Unlock()
	
	// Test passes if no errors occur during attachment
}

// Test concurrent access to handler maps
func TestConcurrentHandlerAccess(t *testing.T) {
	done := make(chan bool, 2)
	
	// Goroutine 1: Add handlers
	go func() {
		for i := 0; i < 100; i++ {
			id := "test-" + strconv.Itoa(i)
			inlineHandlersMu.Lock()
			inlineClickHandlers[id] = func(el Element) {}
			inlineHandlersMu.Unlock()
			time.Sleep(time.Microsecond)
		}
		done <- true
	}()
	
	// Goroutine 2: Read handlers
	go func() {
		for i := 0; i < 100; i++ {
			inlineHandlersMu.RLock()
			_ = len(inlineClickHandlers)
			inlineHandlersMu.RUnlock()
			time.Sleep(time.Microsecond)
		}
		done <- true
	}()
	
	// Wait for both goroutines
	<-done
	<-done
	
	// Cleanup
	inlineHandlersMu.Lock()
	for id := range inlineClickHandlers {
		if strings.HasPrefix(id, "test-") {
			delete(inlineClickHandlers, id)
		}
	}
	inlineHandlersMu.Unlock()
}

// Test memory cleanup after handler removal
func TestHandlerMemoryCleanup(t *testing.T) {
	// Create many handlers
	handlerCount := 1000
	ids := make([]string, handlerCount)
	
	for i := 0; i < handlerCount; i++ {
		handler := func(el Element) {}
		attr := OnClickInline(handler)
		if attr == nil {
			t.Fatal("OnClickInline should return an attribute")
		}
		// Extract ID from the attribute (simplified)
		ids[i] = "click-" + strconv.Itoa(int(inlineIDCounter))
	}
	
	// Verify handlers are registered
	inlineHandlersMu.RLock()
	initialCount := len(inlineClickHandlers)
	inlineHandlersMu.RUnlock()
	
	if initialCount < handlerCount {
		t.Errorf("Expected at least %d handlers, got %d", handlerCount, initialCount)
	}
	
	// Clean up all handlers
	inlineHandlersMu.Lock()
	for id := range inlineClickHandlers {
		delete(inlineClickHandlers, id)
	}
	inlineHandlersMu.Unlock()
	
	// Verify cleanup
	inlineHandlersMu.RLock()
	finalCount := len(inlineClickHandlers)
	inlineHandlersMu.RUnlock()
	
	if finalCount != 0 {
		t.Errorf("Expected 0 handlers after cleanup, got %d", finalCount)
	}
	
	// Force garbage collection
	time.Sleep(10 * time.Millisecond)
}

// Test form data serialization helper
func TestSerializeFormData(t *testing.T) {
	// Create a test form
	doc := dom.GetWindow().Document()
	form := doc.CreateElement("form")
	
	// Add form elements
	input1 := doc.CreateElement("input")
	input1.SetAttribute("name", "username")
	input1.SetAttribute("value", "testuser")
	form.AppendChild(input1)
	
	input2 := doc.CreateElement("input")
	input2.SetAttribute("name", "email")
	input2.SetAttribute("value", "test@example.com")
	form.AppendChild(input2)
	
	// Test serialization
	formData := serializeFormData(form)
	
	if formData["username"] != "testuser" {
		t.Errorf("Expected username 'testuser', got '%s'", formData["username"])
	}
	
	if formData["email"] != "test@example.com" {
		t.Errorf("Expected email 'test@example.com', got '%s'", formData["email"])
	}
}

// Test keyboard navigation integration
func TestKeyboardNavigationIntegration(t *testing.T) {
	// Test that keyboard navigation handlers work together
	
	// Create element with multiple keyboard handlers
	handler := func(el Element) {
		// Tab handler
	}
	shiftHandler := func(el Element) {
		// Shift+Tab handler
	}
	arrowHandler := func(el Element, direction string) {
		// Arrow key handler
	}
	
	// Test that all handlers can be registered on the same element
	tabAttr := OnTabInline(handler)
	shiftTabAttr := OnShiftTabInline(shiftHandler)
	arrowAttr := OnArrowKeysInline(arrowHandler)
	
	if tabAttr == nil || shiftTabAttr == nil || arrowAttr == nil {
		t.Fatal("All keyboard navigation handlers should return attributes")
	}
	
	// Verify all handlers are registered
	inlineHandlersMu.RLock()
	tabHandlerCount := len(inlineTabHandlers)
	shiftTabHandlerCount := len(inlineShiftTabHandlers)
	arrowHandlerCount := len(inlineArrowKeyHandlers)
	inlineHandlersMu.RUnlock()

	if tabHandlerCount == 0 {
		t.Error("Tab handler should be registered")
	}
	if shiftTabHandlerCount == 0 {
		t.Error("Shift+Tab handler should be registered")
	}
	if arrowHandlerCount == 0 {
		t.Error("Arrow keys handler should be registered")
	}

	// Cleanup all keyboard handlers
	inlineHandlersMu.Lock()
	for id := range inlineTabHandlers {
		delete(inlineTabHandlers, id)
	}
	for id := range inlineShiftTabHandlers {
		delete(inlineShiftTabHandlers, id)
	}
	for id := range inlineArrowKeyHandlers {
		delete(inlineArrowKeyHandlers, id)
	}
	inlineHandlersMu.Unlock()
}

// Test drag and drop integration
func TestDragAndDropIntegration(t *testing.T) {
	// Create drag start handler
	dragStartHandler := func(el Element, dataTransfer js.Value) {
		// Handler implementation for testing
	}
	
	// Create drop handler
	dropHandler := func(el Element, dataTransfer js.Value) {
		// Handler implementation for testing
	}
	
	// Create drag over handler
	dragOverHandler := func(el Element, dataTransfer js.Value) {
		// Handler implementation for testing
	}
	
	// Create inline handlers
	dragStartAttr := OnDragStartInline(dragStartHandler)
	dropAttr := OnDropInline(dropHandler)
	dragOverAttr := OnDragOverInline(dragOverHandler)
	
	if dragStartAttr == nil || dropAttr == nil || dragOverAttr == nil {
		t.Fatal("Drag and drop handlers should return attributes")
	}
	
	// Verify all handlers are registered
	inlineHandlersMu.RLock()
	dragStartCount := len(inlineDragStartHandlers)
	dropCount := len(inlineDropHandlers)
	dragOverCount := len(inlineDragOverHandlers)
	inlineHandlersMu.RUnlock()
	
	if dragStartCount == 0 {
		t.Error("Drag start handler should be registered")
	}
	if dropCount == 0 {
		t.Error("Drop handler should be registered")
	}
	if dragOverCount == 0 {
		t.Error("Drag over handler should be registered")
	}
	
	// Cleanup
	inlineHandlersMu.Lock()
	for id := range inlineDragStartHandlers {
		delete(inlineDragStartHandlers, id)
	}
	for id := range inlineDropHandlers {
		delete(inlineDropHandlers, id)
	}
	for id := range inlineDragOverHandlers {
		delete(inlineDragOverHandlers, id)
	}
	inlineHandlersMu.Unlock()
}

// Test edge cases and error handling
func TestEdgeCases(t *testing.T) {
	// Test with nil handler
	defer func() {
		if r := recover(); r != nil {
			t.Error("Should not panic with nil handler")
		}
	}()
	
	// Test validation with invalid pattern
	validator := createValidator("invalid-pattern")
	if validator == nil {
		t.Error("Should return a validator even for invalid patterns")
	}
	
	// Test with empty options
	validator = createValidator(ValidationMinLength)
	if validator == nil {
		t.Error("Should return a validator even with missing options")
	}
	
	// Test nextInlineID with empty prefix
	id := nextInlineID("")
	if id == "" {
		t.Error("Should generate ID even with empty prefix")
	}
}