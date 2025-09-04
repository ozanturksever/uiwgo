//go:build js && wasm

package bridge

import (
	"testing"
)

// TestRealBridgeInitialization tests the real bridge initialization in WASM environment
func TestRealBridgeInitialization(t *testing.T) {
	// Reset manager state
	ResetManager()
	
	// Create a real manager
	manager := NewRealManager()
	if manager == nil {
		t.Fatal("Expected NewRealManager to return a non-nil manager")
	}
	
	// Set the manager
	SetManager(manager)
	
	// Should now return the real manager
	if GetManager() != manager {
		t.Error("Expected GetManager to return the set manager")
	}
	
	// Test manager methods
	if manager.JS() == nil {
		t.Error("Expected JS() to return a non-nil bridge")
	}
	
	if manager.DOM() == nil {
		t.Error("Expected DOM() to return a non-nil bridge")
	}
	
	if manager.Component() == nil {
		t.Error("Expected Component() to return a non-nil bridge")
	}
}

// TestRealJSBridge tests the real JS bridge functionality
func TestRealJSBridge(t *testing.T) {
	
	// Create a real manager
	manager := NewRealManager()
	SetManager(manager)
	
	// Get JS bridge
	jsBridge := manager.JS()
	if jsBridge == nil {
		t.Fatal("Expected JSBridge to return a non-nil bridge")
	}
	
	// Test Global function
	globalVal := jsBridge.Global()
	if globalVal == nil {
		t.Error("Expected Global to return a non-nil value")
	}
	
	// Test accessing console from global
	console := globalVal.Get("console")
	if console == nil {
		t.Error("Expected to access console from global")
	}
	
	// Test ValueOf function
	val := jsBridge.ValueOf("test string")
	if val == nil {
		t.Error("Expected ValueOf to return a non-nil value")
	}
	if val.String() != "test string" {
		t.Errorf("Expected ValueOf string to be 'test string', got %s", val.String())
	}
	
	// Test ValueOf with different types
	intVal := jsBridge.ValueOf(42)
	if intVal.Int() != 42 {
		t.Errorf("Expected ValueOf int to be 42, got %d", intVal.Int())
	}
	
	boolVal := jsBridge.ValueOf(true)
	if !boolVal.Bool() {
		t.Error("Expected ValueOf bool to be true")
	}
	
	floatVal := jsBridge.ValueOf(3.14)
	if floatVal.Float() != 3.14 {
		t.Errorf("Expected ValueOf float to be 3.14, got %f", floatVal.Float())
	}
}

// TestRealDOMBridge tests the real DOM bridge functionality
func TestRealDOMBridge(t *testing.T) {
	// Create a real manager
	manager := NewRealManager()
	SetManager(manager)
	
	// Get DOM bridge
	domBridge := manager.DOM()
	if domBridge == nil {
		t.Fatal("Expected DOMBridge to return a non-nil bridge")
	}
	
	// Test Document
	doc := domBridge.Document()
	if doc == nil {
		t.Fatal("Expected GetDocument to return a non-nil document")
	}
	
	// Test document body
	body := doc.Body()
	if body == nil {
		t.Error("Expected Body to return a non-nil element")
	}
	
	// Test CreateElement
	element := doc.CreateElement("div")
	if element == nil {
		t.Fatal("Expected CreateElement to return a non-nil element")
	}
	
	// Test element properties
	if element.TagName() != "DIV" {
		t.Errorf("Expected TagName to be 'DIV', got %s", element.TagName())
	}
	
	// Test setting and getting ID
	element.SetID("test-id")
	if element.ID() != "test-id" {
		t.Errorf("Expected ID to be 'test-id', got %s", element.ID())
	}
	
	// Test setting and getting class name
	element.SetClassName("test-class")
	if element.ClassName() != "test-class" {
		t.Errorf("Expected ClassName to be 'test-class', got %s", element.ClassName())
	}
	
	// Test setting and getting text content
	element.SetTextContent("test content")
	if element.TextContent() != "test content" {
		t.Errorf("Expected TextContent to be 'test content', got %s", element.TextContent())
	}
	
	// Test setting and getting innerHTML
	element.SetInnerHTML("<span>test</span>")
	if element.InnerHTML() != "<span>test</span>" {
		t.Errorf("Expected InnerHTML to be '<span>test</span>', got %s", element.InnerHTML())
	}
	
	// Test attributes
	element.SetAttribute("data-test", "test-value")
	if element.GetAttribute("data-test") != "test-value" {
		t.Errorf("Expected attribute to be 'test-value', got %s", element.GetAttribute("data-test"))
	}
	
	// Test removing attributes
	element.RemoveAttribute("data-test")
	if element.GetAttribute("data-test") != "" {
		t.Error("Expected attribute to be removed")
	}
}

// TestRealDOMStyle tests the real DOM style functionality
func TestRealDOMStyle(t *testing.T) {
	
	// Create a real manager
	manager := NewRealManager()
	SetManager(manager)
	
	// Get DOM bridge and create element
	domBridge := manager.DOM()
	doc := domBridge.Document()
	element := doc.CreateElement("div")
	
	// Get style
	style := element.Style()
	if style == nil {
		t.Fatal("Expected Style to return a non-nil style")
	}
	
	// Test setting and getting style properties
	style.Set("color", "red")
	if style.Get("color") != "red" {
		t.Errorf("Expected color to be 'red', got %s", style.Get("color"))
	}
	
	// Test removing style properties
	style.Remove("color")
	if style.Get("color") != "" {
		t.Error("Expected color property to be removed")
	}
}

// TestRealDOMEvent tests the real DOM event functionality
func TestRealDOMEvent(t *testing.T) {
	
	// Create a real manager
	manager := NewRealManager()
	SetManager(manager)
	
	// Get DOM bridge
	domBridge := manager.DOM()
	
	// Test CreateEvent
	event := domBridge.CreateEvent("Event")
	if event == nil {
		t.Fatal("Expected CreateEvent to return a non-nil event")
	}
	
	// Note: In a mock environment, we can't test all event properties
	// but we can test that the event object is created
}

// TestRealComponentBridge tests the real component bridge functionality
func TestRealComponentBridge(t *testing.T) {
	// Create a real manager
	manager := NewRealManager()
	SetManager(manager)
	
	// Get component bridge
	compBridge := manager.Component()
	if compBridge == nil {
		t.Fatal("Expected ComponentBridge to return a non-nil bridge")
	}
	
	// Test component initialization
	err := compBridge.InitializeComponent("test-component", "#test-selector", map[string]any{"data": "test-data"})
	if err != nil {
		t.Errorf("Expected no error from InitializeComponent, got %v", err)
	}
	
	// Test component destruction
	err = compBridge.DestroyComponent("#test-selector", "test-component")
	if err != nil {
		t.Errorf("Expected no error from DestroyComponent, got %v", err)
	}
}

// TestConvenienceFunctionsWithRealManager tests the convenience functions with real manager
func TestConvenienceFunctionsWithRealManager(t *testing.T) {
	// Reset manager state
	ResetManager()
	
	// Create and set a real manager
	manager := NewRealManager()
	SetManager(manager)
	
	// Test InitializeComponent
	err := InitializeComponent("test-component", "#test-selector", map[string]any{"data": "test-data"})
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	
	// Test DestroyComponent
	err = DestroyComponent("#test-selector", "test-component")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	
	// Test GetDocument
	doc, err := GetDocument()
	if err != nil {
		t.Errorf("Expected no error from GetDocument, got %v", err)
	}
	if doc == nil {
		t.Error("Expected GetDocument to return a document")
	}
	
	// Test QuerySelector
	element, err := QuerySelector("body")
	if err != nil {
		t.Errorf("Expected no error from QuerySelector, got %v", err)
	}
	if element == nil {
		t.Error("Expected QuerySelector to return an element")
	}
	
	// Test GetElementByID
	// First create an element with an ID
	document, err := GetDocument()
	if err != nil {
		t.Errorf("Expected no error from GetDocument, got %v", err)
	}
	testElement := document.CreateElement("div")
	testElement.SetID("test-element")
	document.Body().AppendChild(testElement)
	
	element, err = GetElementByID("test-element")
	if err != nil {
		t.Errorf("Expected no error from GetElementByID, got %v", err)
	}
	if element == nil {
		t.Error("Expected GetElementByID to return an element")
	}
	if element.ID() != "test-element" {
		t.Errorf("Expected element ID to be 'test-element', got %s", element.ID())
	}
}

// TestManagerErrorHandling tests error handling in manager operations
func TestManagerErrorHandling(t *testing.T) {
	// Reset manager state
	ResetManager()
	
	// Test operations without manager
	err := InitializeComponent("test-component", "#test-selector", map[string]any{"data": "test-data"})
	if err == nil {
		t.Error("Expected error when no manager is set")
	}
	
	err = DestroyComponent("#test-selector", "test-component")
	if err == nil {
		t.Error("Expected error when no manager is set")
	}
	
	doc, err := GetDocument()
	if err == nil {
		t.Error("Expected error when no manager is set")
	}
	if doc != nil {
		t.Error("Expected GetDocument to return nil when no manager is set")
	}
	
	element, err := QuerySelector("#test")
	if err == nil {
		t.Error("Expected error from QuerySelector when no manager is set")
	}
	if element != nil {
		t.Error("Expected QuerySelector to return nil when no manager is set")
	}
	
	element, err = GetElementByID("test")
	if err == nil {
		t.Error("Expected error from GetElementByID when no manager is set")
	}
	if element != nil {
		t.Error("Expected GetElementByID to return nil when no manager is set")
	}
}