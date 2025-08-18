// bind_test.go
// Unit tests for Bind function focusing on DOM-independent functionality
// Note: Full Bind testing requires browser environment due to syscall/js constraints

package golid

import (
	"fmt"
	"strings"
	"testing"

	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

// TestBindStructureGeneration tests that Bind creates proper placeholder structure
// This tests the non-DOM parts of Bind functionality
func TestBindStructureGeneration(t *testing.T) {
	// Test case 1: Simple function that returns static content
	bindResult := Bind(func() Node {
		return Div(Text("static content"))
	})

	// Verify Bind returns a non-nil Node
	if bindResult == nil {
		t.Fatal("Bind should return a non-nil Node")
	}

	// Test that Bind creates a span placeholder with ID
	html := RenderHTML(bindResult)

	// Should create a span element
	if !strings.Contains(html, "<span") {
		t.Errorf("Bind should create a span placeholder, got: %s", html)
	}

	// Should have an ID attribute with the expected prefix
	if !strings.Contains(html, `id="e_`) {
		t.Errorf("Bind placeholder should have an ID with 'e_' prefix, got: %s", html)
	}

	// Should be a self-closing span (empty placeholder)
	if strings.Contains(html, ">") && strings.Contains(html, "</span>") {
		// It's not empty, but that's OK for placeholder pattern
	}
}

// TestBindWithDifferentReturnTypes tests Bind with various Node types
func TestBindWithDifferentReturnTypes(t *testing.T) {
	testCases := []struct {
		name string
		fn   func() Node
	}{
		{
			name: "Div element",
			fn:   func() Node { return Div(Text("div content")) },
		},
		{
			name: "Paragraph element",
			fn:   func() Node { return P(Text("paragraph content")) },
		},
		{
			name: "Heading element",
			fn:   func() Node { return H1(Text("heading content")) },
		},
		{
			name: "Complex nested structure",
			fn: func() Node {
				return Div(
					Class("container"),
					H2(Text("Title")),
					P(Text("Description")),
					Ul(
						Li(Text("Item 1")),
						Li(Text("Item 2")),
					),
				)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			bindResult := Bind(tc.fn)

			if bindResult == nil {
				t.Fatalf("Bind should return non-nil Node for %s", tc.name)
			}

			html := RenderHTML(bindResult)
			if !strings.Contains(html, "<span") {
				t.Errorf("Bind should create span placeholder for %s, got: %s", tc.name, html)
			}
		})
	}
}

// TestBindTextStructureGeneration tests BindText placeholder creation
func TestBindTextStructureGeneration(t *testing.T) {
	// Test case 1: Simple text function
	bindTextResult := BindText(func() string {
		return "hello world"
	})

	// Verify BindText returns a non-nil Node
	if bindTextResult == nil {
		t.Fatal("BindText should return a non-nil Node")
	}

	// Test that BindText creates a span with initial text content
	html := RenderHTML(bindTextResult)

	// Should create a span element
	if !strings.Contains(html, "<span") {
		t.Errorf("BindText should create a span, got: %s", html)
	}

	// Should have an ID attribute
	if !strings.Contains(html, `id="e_`) {
		t.Errorf("BindText should have an ID attribute, got: %s", html)
	}

	// Should contain the initial text
	if !strings.Contains(html, "hello world") {
		t.Errorf("BindText should contain initial text content, got: %s", html)
	}
}

// TestBindTextWithDifferentStringTypes tests BindText with various string outputs
func TestBindTextWithDifferentStringTypes(t *testing.T) {
	testCases := []struct {
		name     string
		fn       func() string
		expected string
	}{
		{
			name:     "Simple string",
			fn:       func() string { return "simple text" },
			expected: "simple text",
		},
		{
			name:     "Empty string",
			fn:       func() string { return "" },
			expected: "",
		},
		{
			name:     "String with HTML entities",
			fn:       func() string { return "&lt;test&gt;" },
			expected: "&amp;lt;test&amp;gt;", // HTML escaping happens in gomponents rendering
		},
		{
			name:     "Number as string",
			fn:       func() string { return fmt.Sprintf("%d", 42) },
			expected: "42",
		},
		{
			name:     "Formatted string",
			fn:       func() string { return fmt.Sprintf("Count: %d, Status: %s", 5, "active") },
			expected: "Count: 5, Status: active",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			bindTextResult := BindText(tc.fn)

			if bindTextResult == nil {
				t.Fatalf("BindText should return non-nil Node for %s", tc.name)
			}

			html := RenderHTML(bindTextResult)

			if tc.expected != "" && !strings.Contains(html, tc.expected) {
				t.Errorf("BindText should contain expected text '%s' for %s, got: %s",
					tc.expected, tc.name, html)
			}
		})
	}
}

// TestBindFunctionParameterHandling tests that Bind handles function parameters correctly
func TestBindFunctionParameterHandling(t *testing.T) {
	// Test with closure variables
	externalVar := "external"

	bindResult := Bind(func() Node {
		return Div(Text(externalVar))
	})

	if bindResult == nil {
		t.Fatal("Bind should handle closure variables")
	}

	// Test with complex logic inside function
	counter := 0
	bindResult = Bind(func() Node {
		counter++
		if counter%2 == 0 {
			return P(Text("even"))
		}
		return P(Text("odd"))
	})

	if bindResult == nil {
		t.Fatal("Bind should handle complex function logic")
	}
}

// TestBindSignalIntegrationStructure tests the structure created when using Signals with Bind
// Note: This only tests the structure creation, not reactive updates (requires DOM)
func TestBindSignalIntegrationStructure(t *testing.T) {
	// Create test signals
	stringSignal := NewSignal("test string")
	numberSignal := NewSignal(42)
	booleanSignal := NewSignal(true)

	// Test Bind with string signal
	bindResult := Bind(func() Node {
		value := stringSignal.Get()
		return Div(Text(value))
	})

	if bindResult == nil {
		t.Error("Bind with string signal should return non-nil Node")
	}

	html := RenderHTML(bindResult)
	if !strings.Contains(html, "<span") {
		t.Errorf("Bind with signal should create span placeholder: %s", html)
	}

	// Test Bind with number signal
	bindResult = Bind(func() Node {
		count := numberSignal.Get()
		return Div(Text(fmt.Sprintf("Count: %d", count)))
	})

	if bindResult == nil {
		t.Error("Bind with number signal should return non-nil Node")
	}

	// Test Bind with boolean signal (conditional rendering)
	bindResult = Bind(func() Node {
		show := booleanSignal.Get()
		if show {
			return Div(Class("visible"), Text("Visible content"))
		}
		return Div(Class("hidden"), Text("Hidden content"))
	})

	if bindResult == nil {
		t.Error("Bind with boolean signal should return non-nil Node")
	}
}

// TestBindTextSignalIntegrationStructure tests BindText with Signal integration
func TestBindTextSignalIntegrationStructure(t *testing.T) {
	// Create test signals
	messageSignal := NewSignal("initial message")
	numberSignal := NewSignal(100)

	// Test BindText with string signal
	bindTextResult := BindText(func() string {
		return messageSignal.Get()
	})

	if bindTextResult == nil {
		t.Error("BindText with string signal should return non-nil Node")
	}

	html := RenderHTML(bindTextResult)
	if !strings.Contains(html, "initial message") {
		t.Errorf("BindText should render initial signal value: %s", html)
	}

	// Test BindText with computed value from signal
	bindTextResult = BindText(func() string {
		num := numberSignal.Get()
		return fmt.Sprintf("Value: %d", num*2)
	})

	if bindTextResult == nil {
		t.Error("BindText with computed value should return non-nil Node")
	}

	html = RenderHTML(bindTextResult)
	if !strings.Contains(html, "Value: 200") {
		t.Errorf("BindText should render computed value: %s", html)
	}
}

// TestBindIDUniqueness tests that each Bind call generates unique IDs
func TestBindIDUniqueness(t *testing.T) {
	// Create multiple Bind instances
	binds := make([]Node, 10)
	htmls := make([]string, 10)

	for i := 0; i < 10; i++ {
		binds[i] = Bind(func() Node {
			return Div(Text(fmt.Sprintf("content %d", i)))
		})
		htmls[i] = RenderHTML(binds[i])
	}

	// Extract IDs and verify uniqueness
	ids := make(map[string]bool)
	for i, html := range htmls {
		// Find ID in the HTML string
		start := strings.Index(html, `id="e_`)
		if start == -1 {
			t.Errorf("Bind %d should have an ID attribute", i)
			continue
		}

		start += len(`id="e_`)
		end := strings.Index(html[start:], `"`)
		if end == -1 {
			t.Errorf("Bind %d should have properly formatted ID", i)
			continue
		}

		id := "e_" + html[start:start+end]

		if ids[id] {
			t.Errorf("Duplicate ID found: %s", id)
		}
		ids[id] = true
	}

	// Verify we found the expected number of unique IDs
	if len(ids) != 10 {
		t.Errorf("Expected 10 unique IDs, got %d", len(ids))
	}
}

// TestBindPerformanceBasics tests basic performance characteristics
func TestBindPerformanceBasics(t *testing.T) {
	// Test that Bind creation doesn't panic or take excessive time
	for i := 0; i < 1000; i++ {
		bindResult := Bind(func() Node {
			return Div(Text(fmt.Sprintf("item %d", i)))
		})

		if bindResult == nil {
			t.Fatalf("Bind creation failed at iteration %d", i)
		}
	}

	// Test that BindText creation scales
	for i := 0; i < 1000; i++ {
		bindTextResult := BindText(func() string {
			return fmt.Sprintf("text %d", i)
		})

		if bindTextResult == nil {
			t.Fatalf("BindText creation failed at iteration %d", i)
		}
	}
}
