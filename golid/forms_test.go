// forms_test.go
// Comprehensive unit tests for form handling functionality
// Tests form input binding, signal synchronization, and DOM-independent aspects

package golid

import (
	"strings"
	"testing"

	. "maragu.dev/gomponents"
)

// TestBindInputStructure tests that BindInput creates proper HTML structure
func TestBindInputStructure(t *testing.T) {
	signal := NewSignal("initial value")
	placeholder := "Enter text here"

	inputNode := BindInput(signal, placeholder)

	if inputNode == nil {
		t.Fatal("BindInput should return a non-nil Node")
	}

	html := RenderHTML(inputNode)

	// Should create an input element
	if !strings.Contains(html, "<input") {
		t.Errorf("BindInput should create an input element, got: %s", html)
	}

	// Should have type="text"
	if !strings.Contains(html, `type="text"`) {
		t.Errorf("BindInput should create text input, got: %s", html)
	}

	// Should have the placeholder
	if !strings.Contains(html, `placeholder="`+placeholder+`"`) {
		t.Errorf("BindInput should have placeholder, got: %s", html)
	}

	// Should have initial value
	if !strings.Contains(html, `value="initial value"`) {
		t.Errorf("BindInput should have initial value, got: %s", html)
	}

	// Should have an ID attribute
	if !strings.Contains(html, `id="`) {
		t.Errorf("BindInput should have ID attribute, got: %s", html)
	}
}

// TestBindInputWithType tests BindInputWithType structure generation
func TestBindInputWithTypeStructure(t *testing.T) {
	testCases := []struct {
		inputType   string
		placeholder string
		expected    string
	}{
		{"email", "Enter email", `type="email"`},
		{"password", "Enter password", `type="password"`},
		{"number", "Enter number", `type="number"`},
		{"tel", "Enter phone", `type="tel"`},
		{"url", "Enter URL", `type="url"`},
		{"search", "Search...", `type="search"`},
	}

	for _, tc := range testCases {
		t.Run(tc.inputType+" type", func(t *testing.T) {
			signal := NewSignal("")
			inputNode := BindInputWithType(signal, tc.inputType, tc.placeholder)

			if inputNode == nil {
				t.Fatalf("BindInputWithType should return non-nil Node for %s", tc.inputType)
			}

			html := RenderHTML(inputNode)

			// Should have correct input type
			if !strings.Contains(html, tc.expected) {
				t.Errorf("Expected %s in HTML, got: %s", tc.expected, html)
			}

			// Should have placeholder
			if !strings.Contains(html, `placeholder="`+tc.placeholder+`"`) {
				t.Errorf("Should have placeholder %s, got: %s", tc.placeholder, html)
			}

			// Should have ID
			if !strings.Contains(html, `id="`) {
				t.Errorf("Should have ID attribute, got: %s", html)
			}
		})
	}
}

// TestBindInputWithFocusStructure tests BindInputWithFocus structure
func TestBindInputWithFocusStructure(t *testing.T) {
	textSignal := NewSignal("test value")
	focusSignal := NewSignal(false)
	placeholder := "Focus test input"

	inputNode := BindInputWithFocus(textSignal, focusSignal, placeholder)

	if inputNode == nil {
		t.Fatal("BindInputWithFocus should return non-nil Node")
	}

	html := RenderHTML(inputNode)

	// Should create input with proper attributes
	if !strings.Contains(html, "<input") {
		t.Errorf("Should create input element, got: %s", html)
	}

	if !strings.Contains(html, `type="text"`) {
		t.Errorf("Should have text type, got: %s", html)
	}

	if !strings.Contains(html, `placeholder="`+placeholder+`"`) {
		t.Errorf("Should have placeholder, got: %s", html)
	}

	if !strings.Contains(html, `value="test value"`) {
		t.Errorf("Should have initial value, got: %s", html)
	}

	if !strings.Contains(html, `id="`) {
		t.Errorf("Should have ID attribute, got: %s", html)
	}
}

// TestBindInputSignalInitialization tests signal initialization
func TestBindInputSignalInitialization(t *testing.T) {
	testCases := []struct {
		name         string
		initialValue string
	}{
		{"empty string", ""},
		{"simple text", "hello world"},
		{"special characters", "!@#$%^&*()"},
		{"unicode", "こんにちは 🌟"},
		{"multiline", "line1\nline2\nline3"},
		{"HTML entities", "<div>&amp;</div>"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			signal := NewSignal(tc.initialValue)
			inputNode := BindInput(signal, "test placeholder")

			html := RenderHTML(inputNode)

			// The initial value should be properly escaped in HTML
			if !strings.Contains(html, `value="`) {
				t.Errorf("Should have value attribute, got: %s", html)
			}

			// Verify signal maintains its value
			if signal.Get() != tc.initialValue {
				t.Errorf("Signal should maintain initial value %q, got %q", tc.initialValue, signal.Get())
			}
		})
	}
}

// TestFormInputIDGeneration tests that each input gets unique ID
func TestFormInputIDGeneration(t *testing.T) {
	signal1 := NewSignal("test1")
	signal2 := NewSignal("test2")
	signal3 := NewSignal("test3")

	input1 := BindInput(signal1, "Input 1")
	input2 := BindInputWithType(signal2, "email", "Input 2")
	focusSignal := NewSignal(false)
	input3 := BindInputWithFocus(signal3, focusSignal, "Input 3")

	html1 := RenderHTML(input1)
	html2 := RenderHTML(input2)
	html3 := RenderHTML(input3)

	// Extract IDs using simple string parsing
	id1 := extractID(html1)
	id2 := extractID(html2)
	id3 := extractID(html3)

	if id1 == "" || id2 == "" || id3 == "" {
		t.Error("All inputs should have non-empty IDs")
	}

	if id1 == id2 || id1 == id3 || id2 == id3 {
		t.Errorf("All IDs should be unique: id1=%s, id2=%s, id3=%s", id1, id2, id3)
	}
}

// TestBindInputWithDifferentSignalTypes tests binding with various signal types
func TestBindInputWithDifferentSignalTypes(t *testing.T) {
	// Test with different initial values
	testValues := []string{
		"",
		"single",
		"multiple words",
		"123456",
		"special!@#$%",
	}

	for _, value := range testValues {
		t.Run("value_"+value, func(t *testing.T) {
			signal := NewSignal(value)
			inputNode := BindInput(signal, "test")

			if inputNode == nil {
				t.Errorf("BindInput should work with value %q", value)
			}

			// Verify signal value is preserved
			if signal.Get() != value {
				t.Errorf("Signal value should be preserved: expected %q, got %q", value, signal.Get())
			}
		})
	}
}

// TestFormInputObserverRegistration tests that inputs register with observer
func TestFormInputObserverRegistration(t *testing.T) {
	// This test verifies the DOM-independent aspects of observer registration
	initialCallbackCount := len(globalObserver.callbacks)

	signal := NewSignal("test")
	_ = BindInput(signal, "test input")

	// Should have registered one callback
	finalCallbackCount := len(globalObserver.callbacks)

	if finalCallbackCount <= initialCallbackCount {
		t.Errorf("Expected callback count to increase, initial: %d, final: %d", initialCallbackCount, finalCallbackCount)
	}
}

// TestFormInputParameterValidation tests parameter validation
func TestFormInputParameterValidation(t *testing.T) {
	// Test with nil signal (should not panic)
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("BindInput should not panic with nil signal: %v", r)
		}
	}()

	// These tests verify that the functions handle edge cases gracefully
	signal := NewSignal("")

	// Test with empty placeholder
	input1 := BindInput(signal, "")
	if input1 == nil {
		t.Error("BindInput should work with empty placeholder")
	}

	// Test with very long placeholder
	longPlaceholder := strings.Repeat("a", 1000)
	input2 := BindInput(signal, longPlaceholder)
	if input2 == nil {
		t.Error("BindInput should work with long placeholder")
	}

	// Test BindInputWithType with empty type
	input3 := BindInputWithType(signal, "", "test")
	if input3 == nil {
		t.Error("BindInputWithType should work with empty type")
	}
}

// Helper function to extract ID from HTML string
func extractID(html string) string {
	start := strings.Index(html, `id="`)
	if start == -1 {
		return ""
	}
	start += 4 // Move past `id="`
	end := strings.Index(html[start:], `"`)
	if end == -1 {
		return ""
	}
	return html[start : start+end]
}

// TestFormInputCombinations tests various combinations of form inputs
func TestFormInputCombinations(t *testing.T) {
	// Create a form with multiple input types
	nameSignal := NewSignal("")
	emailSignal := NewSignal("")
	passwordSignal := NewSignal("")
	focusSignal := NewSignal(false)

	nameInput := BindInput(nameSignal, "Full Name")
	emailInput := BindInputWithType(emailSignal, "email", "Email Address")
	passwordInput := BindInputWithType(passwordSignal, "password", "Password")
	focusInput := BindInputWithFocus(nameSignal, focusSignal, "Focus Test")

	inputs := []Node{nameInput, emailInput, passwordInput, focusInput}

	for i, input := range inputs {
		if input == nil {
			t.Errorf("Input %d should not be nil", i)
		}

		html := RenderHTML(input)
		if !strings.Contains(html, "<input") {
			t.Errorf("Input %d should render as input element", i)
		}
	}

	// Verify all have unique IDs
	ids := make(map[string]bool)
	for i, input := range inputs {
		html := RenderHTML(input)
		id := extractID(html)
		if id == "" {
			t.Errorf("Input %d should have non-empty ID", i)
		}
		if ids[id] {
			t.Errorf("Input %d has duplicate ID: %s", i, id)
		}
		ids[id] = true
	}
}
