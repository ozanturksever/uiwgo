// dom_test.go
// Unit tests for DOM utilities and reactive binding functionality

package golid

import (
	"strings"
	"testing"

	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

// TestGenID tests the unique ID generation functionality
func TestGenID(t *testing.T) {
	// Test that GenID generates non-empty IDs
	id1 := GenID()
	if id1 == "" {
		t.Error("GenID should return a non-empty string")
	}

	// Test that GenID has the expected prefix
	if !strings.HasPrefix(id1, "e_") {
		t.Errorf("GenID should start with 'e_', got: %s", id1)
	}

	// Test that GenID generates unique IDs
	id2 := GenID()
	if id1 == id2 {
		t.Errorf("GenID should generate unique IDs, got duplicate: %s", id1)
	}

	// Test multiple generations for uniqueness
	ids := make(map[string]bool)
	for i := 0; i < 100; i++ {
		id := GenID()
		if ids[id] {
			t.Errorf("GenID generated duplicate ID: %s", id)
		}
		ids[id] = true
	}
}

// TestRenderHTML tests HTML rendering functionality
func TestRenderHTML(t *testing.T) {
	// Test simple element rendering
	node := Div(Text("hello"))
	html := RenderHTML(node)
	expected := "<div>hello</div>"
	if html != expected {
		t.Errorf("Expected %s, got %s", expected, html)
	}

	// Test element with attributes
	node = Div(Class("test-class"), Text("content"))
	html = RenderHTML(node)
	if !strings.Contains(html, `class="test-class"`) {
		t.Errorf("Expected class attribute in HTML: %s", html)
	}
	if !strings.Contains(html, "content") {
		t.Errorf("Expected text content in HTML: %s", html)
	}

	// Test nested elements
	node = Div(
		H1(Text("Title")),
		P(Text("Paragraph")),
	)
	html = RenderHTML(node)
	if !strings.Contains(html, "<h1>Title</h1>") {
		t.Errorf("Expected nested h1 element: %s", html)
	}
	if !strings.Contains(html, "<p>Paragraph</p>") {
		t.Errorf("Expected nested p element: %s", html)
	}
}

// TestBindComponentStructure tests that Bind creates proper component structure
func TestBindComponentStructure(t *testing.T) {
	// Test that Bind returns a Node
	bindNode := Bind(func() Node {
		return Div(Text("test content"))
	})

	if bindNode == nil {
		t.Error("Bind should return a non-nil Node")
	}

	// Test that Bind creates a span placeholder
	html := RenderHTML(bindNode)
	if !strings.Contains(html, "<span") {
		t.Errorf("Bind should create a span placeholder, got: %s", html)
	}

	// Test that the placeholder has an ID
	if !strings.Contains(html, `id="e_`) {
		t.Errorf("Bind placeholder should have an ID attribute, got: %s", html)
	}
}

// TestBindWithSignal tests Bind integration with Signal system (non-DOM parts)
func TestBindWithSignal(t *testing.T) {
	// Create a signal
	counter := NewSignal(0)

	// Test that Bind can access signal values in function
	bindNode := Bind(func() Node {
		count := counter.Get()
		return Div(Text(string(rune(count + 48)))) // Convert to character
	})

	// Test that the bind node is created
	if bindNode == nil {
		t.Error("Bind with signal should return a non-nil Node")
	}

	// Test HTML structure
	html := RenderHTML(bindNode)
	if !strings.Contains(html, "<span") {
		t.Errorf("Bind should create span placeholder: %s", html)
	}
}

// TestBindTextComponentStructure tests BindText component structure
func TestBindTextComponentStructure(t *testing.T) {
	// Test that BindText returns a Node
	bindTextNode := BindText(func() string {
		return "test text"
	})

	if bindTextNode == nil {
		t.Error("BindText should return a non-nil Node")
	}

	// Test that BindText creates a span with text
	html := RenderHTML(bindTextNode)
	if !strings.Contains(html, "<span") {
		t.Errorf("BindText should create a span, got: %s", html)
	}

	// Test that the span has an ID
	if !strings.Contains(html, `id="e_`) {
		t.Errorf("BindText span should have an ID attribute, got: %s", html)
	}

	// Test that initial text is rendered
	if !strings.Contains(html, "test text") {
		t.Errorf("BindText should render initial text, got: %s", html)
	}
}

// TestBindTextWithSignal tests BindText with Signal integration
func TestBindTextWithSignal(t *testing.T) {
	// Create a signal
	message := NewSignal("hello")

	// Test that BindText can access signal values
	bindTextNode := BindText(func() string {
		return message.Get()
	})

	// Test that the bind text node is created
	if bindTextNode == nil {
		t.Error("BindText with signal should return a non-nil Node")
	}

	// Test HTML structure includes initial value
	html := RenderHTML(bindTextNode)
	if !strings.Contains(html, "hello") {
		t.Errorf("BindText should render signal value: %s", html)
	}
}

// TestForEach tests the ForEach utility function
func TestForEach(t *testing.T) {
	// Test with string slice
	items := []string{"apple", "banana", "cherry"}
	node := ForEach(items, func(item string) Node {
		return Li(Text(item))
	})

	if node == nil {
		t.Error("ForEach should return a non-nil Node")
	}

	html := RenderHTML(node)

	// Check that all items are rendered
	for _, item := range items {
		if !strings.Contains(html, item) {
			t.Errorf("ForEach should render item %s, got: %s", item, html)
		}
	}

	// Check that list items are created
	itemCount := strings.Count(html, "<li>")
	if itemCount != len(items) {
		t.Errorf("Expected %d list items, got %d in: %s", len(items), itemCount, html)
	}
}

// TestForEachSignal tests the ForEachSignal utility function (structure only)
func TestForEachSignal(t *testing.T) {
	// Create a signal with slice data
	items := NewSignal([]string{"one", "two", "three"})

	// Test that ForEachSignal returns a Node
	node := ForEachSignal(items, func(item string) Node {
		return Li(Text(item))
	})

	if node == nil {
		t.Error("ForEachSignal should return a non-nil Node")
	}

	// Test that it creates a Bind wrapper (should be a span)
	html := RenderHTML(node)
	if !strings.Contains(html, "<span") {
		t.Errorf("ForEachSignal should create a Bind wrapper (span), got: %s", html)
	}
}

// TestSignalOperations tests basic Signal functionality used by Bind
func TestSignalOperations(t *testing.T) {
	// Test integer signal
	intSignal := NewSignal(42)
	if intSignal.Get() != 42 {
		t.Errorf("Expected 42, got %d", intSignal.Get())
	}

	intSignal.Set(100)
	if intSignal.Get() != 100 {
		t.Errorf("Expected 100, got %d", intSignal.Get())
	}

	// Test string signal
	strSignal := NewSignal("hello")
	if strSignal.Get() != "hello" {
		t.Errorf("Expected 'hello', got %s", strSignal.Get())
	}

	strSignal.Set("world")
	if strSignal.Get() != "world" {
		t.Errorf("Expected 'world', got %s", strSignal.Get())
	}

	// Test boolean signal
	boolSignal := NewSignal(true)
	if !boolSignal.Get() {
		t.Error("Expected true")
	}

	boolSignal.Set(false)
	if boolSignal.Get() {
		t.Error("Expected false")
	}
}

// TestComplexBindScenarios tests more complex Bind usage patterns
func TestComplexBindScenarios(t *testing.T) {
	// Test Bind with conditional rendering
	showDetails := NewSignal(false)

	bindNode := Bind(func() Node {
		if showDetails.Get() {
			return Div(
				Class("details"),
				H2(Text("Details")),
				P(Text("Some detailed information")),
			)
		}
		return P(Text("Click to show details"))
	})

	if bindNode == nil {
		t.Error("Complex Bind should return a non-nil Node")
	}

	// Test initial state (false)
	html := RenderHTML(bindNode)
	if !strings.Contains(html, "<span") {
		t.Errorf("Complex Bind should create span placeholder: %s", html)
	}

	// Test Bind with multiple signals
	firstName := NewSignal("John")
	lastName := NewSignal("Doe")

	bindNode = Bind(func() Node {
		return H1(Text(firstName.Get() + " " + lastName.Get()))
	})

	if bindNode == nil {
		t.Error("Multi-signal Bind should return a non-nil Node")
	}
}
