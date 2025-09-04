//go:build !js || !wasm

package bridge

import (
	"testing"
)

// TestManagerInitialization tests the global manager initialization
func TestManagerInitialization(t *testing.T) {
	// Reset manager state
	ResetManager()
	
	// Initially should be nil
	if GetManager() != nil {
		t.Error("Expected manager to be nil initially")
	}
	
	// Create a mock manager
	mockManager := &mockManager{}
	
	// Set the manager
	SetManager(mockManager)
	
	// Should now return the mock manager
	if GetManager() != mockManager {
		t.Error("Expected GetManager to return the set manager")
	}
	
	// Reset should clear it
	ResetManager()
	if GetManager() != nil {
		t.Error("Expected manager to be nil after reset")
	}
}

// TestManagerConvenienceFunctions tests the convenience wrapper functions
func TestManagerConvenienceFunctions(t *testing.T) {
	// Reset manager state
	ResetManager()
	
	// Create a mock manager
	mockManager := &mockManager{
		initializeComponentCalled: false,
		destroyComponentCalled:    false,
	}
	
	// Set the manager
	SetManager(mockManager)
	
	// Test InitializeComponent
	err := InitializeComponent("test-id", "test-component")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if !mockManager.initializeComponentCalled {
		t.Error("Expected InitializeComponent to be called on manager")
	}
	
	// Test DestroyComponent
	err = DestroyComponent("test-id")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if !mockManager.destroyComponentCalled {
		t.Error("Expected DestroyComponent to be called on manager")
	}
	
	// Test GetDocument
	doc := GetDocument()
	if doc == nil {
		t.Error("Expected GetDocument to return a document")
	}
	
	// Test QuerySelector
	element := QuerySelector("#test")
	if element == nil {
		t.Error("Expected QuerySelector to return an element")
	}
	
	// Test GetElementByID
	element = GetElementByID("test")
	if element == nil {
		t.Error("Expected GetElementByID to return an element")
	}
}

// TestManagerConvenienceFunctionsWithoutManager tests behavior when no manager is set
func TestManagerConvenienceFunctionsWithoutManager(t *testing.T) {
	// Reset manager state
	ResetManager()
	
	// Test InitializeComponent without manager
	err := InitializeComponent("test-id", "test-component")
	if err == nil {
		t.Error("Expected error when no manager is set")
	}
	
	// Test DestroyComponent without manager
	err = DestroyComponent("test-id")
	if err == nil {
		t.Error("Expected error when no manager is set")
	}
	
	// Test GetDocument without manager
	doc := GetDocument()
	if doc != nil {
		t.Error("Expected GetDocument to return nil when no manager is set")
	}
	
	// Test QuerySelector without manager
	element := QuerySelector("#test")
	if element != nil {
		t.Error("Expected QuerySelector to return nil when no manager is set")
	}
	
	// Test GetElementByID without manager
	element = GetElementByID("test")
	if element != nil {
		t.Error("Expected GetElementByID to return nil when no manager is set")
	}
}

// mockManager is a simple mock implementation for testing
type mockManager struct {
	initializeComponentCalled bool
	destroyComponentCalled    bool
}

func (m *mockManager) JSBridge() JSBridge {
	return &mockJSBridge{}
}

func (m *mockManager) DOMBridge() DOMBridge {
	return &mockDOMBridge{}
}

func (m *mockManager) ComponentBridge() ComponentBridge {
	return &mockComponentBridge{
		manager: m,
	}
}

func (m *mockManager) Initialize() error {
	return nil
}

func (m *mockManager) Cleanup() error {
	return nil
}

// mockJSBridge is a simple mock implementation
type mockJSBridge struct{}

func (m *mockJSBridge) Global(name string) JSValue {
	return &mockJSValue{}
}

func (m *mockJSBridge) ValueOf(value interface{}) JSValue {
	return &mockJSValue{}
}

// mockJSValue is a simple mock implementation
type mockJSValue struct{}

func (m *mockJSValue) Type() string                                { return "object" }
func (m *mockJSValue) IsUndefined() bool                           { return false }
func (m *mockJSValue) IsNull() bool                                { return false }
func (m *mockJSValue) Bool() bool                                  { return true }
func (m *mockJSValue) Int() int                                    { return 42 }
func (m *mockJSValue) Float() float64                              { return 42.0 }
func (m *mockJSValue) String() string                              { return "mock" }
func (m *mockJSValue) Get(key string) JSValue                      { return &mockJSValue{} }
func (m *mockJSValue) Set(key string, value interface{})           {}
func (m *mockJSValue) Call(method string, args ...interface{}) JSValue { return &mockJSValue{} }

// mockDOMBridge is a simple mock implementation
type mockDOMBridge struct{}

func (m *mockDOMBridge) GetDocument() DOMDocument {
	return &mockDOMDocument{}
}

func (m *mockDOMBridge) QuerySelector(selector string) DOMElement {
	return &mockDOMElement{}
}

func (m *mockDOMBridge) GetElementByID(id string) DOMElement {
	return &mockDOMElement{}
}

// mockDOMDocument is a simple mock implementation
type mockDOMDocument struct{}

func (m *mockDOMDocument) Body() DOMElement                                    { return &mockDOMElement{} }
func (m *mockDOMDocument) GetElementByID(id string) DOMElement                 { return &mockDOMElement{} }
func (m *mockDOMDocument) QuerySelector(selector string) DOMElement            { return &mockDOMElement{} }
func (m *mockDOMDocument) QuerySelectorAll(selector string) []DOMElement      { return []DOMElement{&mockDOMElement{}} }
func (m *mockDOMDocument) CreateElement(tagName string) DOMElement             { return &mockDOMElement{} }

// mockDOMElement is a simple mock implementation
type mockDOMElement struct{}

func (m *mockDOMElement) TagName() string                                      { return "div" }
func (m *mockDOMElement) ID() string                                           { return "test" }
func (m *mockDOMElement) SetID(id string)                                      {}
func (m *mockDOMElement) ClassName() string                                    { return "test-class" }
func (m *mockDOMElement) SetClassName(className string)                        {}
func (m *mockDOMElement) TextContent() string                                  { return "test content" }
func (m *mockDOMElement) SetTextContent(content string)                        {}
func (m *mockDOMElement) InnerHTML() string                                    { return "<span>test</span>" }
func (m *mockDOMElement) SetInnerHTML(html string)                             {}
func (m *mockDOMElement) GetAttribute(name string) string                      { return "test-attr" }
func (m *mockDOMElement) SetAttribute(name, value string)                      {}
func (m *mockDOMElement) RemoveAttribute(name string)                          {}
func (m *mockDOMElement) Style() DOMStyle                                      { return &mockDOMStyle{} }
func (m *mockDOMElement) QuerySelector(selector string) DOMElement             { return &mockDOMElement{} }
func (m *mockDOMElement) QuerySelectorAll(selector string) []DOMElement       { return []DOMElement{&mockDOMElement{}} }
func (m *mockDOMElement) AddEventListener(eventType string, listener func(DOMEvent)) {}
func (m *mockDOMElement) RemoveEventListener(eventType string, listener func(DOMEvent)) {}

// mockDOMStyle is a simple mock implementation
type mockDOMStyle struct{}

func (m *mockDOMStyle) GetPropertyValue(property string) string { return "test-value" }
func (m *mockDOMStyle) SetProperty(property, value string)      {}
func (m *mockDOMStyle) RemoveProperty(property string)          {}

// mockDOMEvent is a simple mock implementation
type mockDOMEvent struct{}

func (m *mockDOMEvent) Type() string                 { return "click" }
func (m *mockDOMEvent) Target() DOMElement           { return &mockDOMElement{} }
func (m *mockDOMEvent) CurrentTarget() DOMElement    { return &mockDOMElement{} }
func (m *mockDOMEvent) PreventDefault()              {}
func (m *mockDOMEvent) StopPropagation()             {}

// mockComponentBridge is a simple mock implementation
type mockComponentBridge struct {
	manager *mockManager
}

func (m *mockComponentBridge) InitializeComponent(id string, component interface{}) error {
	m.manager.initializeComponentCalled = true
	return nil
}

func (m *mockComponentBridge) DestroyComponent(id string) error {
	m.manager.destroyComponentCalled = true
	return nil
}