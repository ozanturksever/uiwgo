package mockdom

import (
	"testing"
	"github.com/ozanturksever/uiwgo/bridge"
)

// TestMockJSValue tests the MockJSValue implementation
func TestMockJSValue(t *testing.T) {
	mock := NewMockJSValue("test")
	
	if mock.Type() != bridge.StubTypeString {
		t.Errorf("Expected type StubTypeString, got %v", mock.Type())
	}
	
	if mock.String() != "test" {
		t.Errorf("Expected string 'test', got %s", mock.String())
	}
	
	if mock.Int() != 0 {
		t.Errorf("Expected int 0, got %d", mock.Int())
	}
	
	if mock.Float() != 0.0 {
		t.Errorf("Expected float 0.0, got %f", mock.Float())
	}
	
	if mock.Bool() {
		t.Error("Expected bool false, got true")
	}
	
	if mock.IsNull() {
		t.Error("Expected IsNull false, got true")
	}
	
	if mock.IsUndefined() {
		t.Error("Expected IsUndefined false, got true")
	}
	
	if mock.Length() != 4 {
		t.Errorf("Expected length 4, got %d", mock.Length())
	}
	
	// Test method calls
	result := mock.Call("toString")
	if result == nil {
		t.Error("Expected Call to return a value")
	}
	
	prop := mock.Get("length")
	if prop == nil {
		t.Error("Expected Get to return a value")
	}
	
	mock.Set("test", "value")
	// No error expected
	
	index := mock.Index(0)
	if index == nil {
		t.Error("Expected Index to return a value")
	}
}

// TestMockJSBridge tests the MockJSBridge implementation
func TestMockJSBridge(t *testing.T) {
	mock := &MockJSBridge{}
	
	global := mock.Global()
	if global == nil {
		t.Error("Expected Global to return a value")
	}
	
	undefined := mock.Undefined()
	if undefined == nil {
		t.Error("Expected Undefined to return a value")
	}
	
	null := mock.Null()
	if null == nil {
		t.Error("Expected Null to return a value")
	}
	
	valueOf := mock.ValueOf("test")
	if valueOf == nil {
		t.Error("Expected ValueOf to return a value")
	}
	
	funcOf := mock.FuncOf(func(this bridge.JSValue, args []bridge.JSValue) interface{} {
		return "result"
	})
	if funcOf == nil {
		t.Error("Expected FuncOf to return a value")
	}
	
	// Test the function
	result := funcOf.Call("call")
	if result == nil {
		t.Error("Expected function to return a value")
	}
}

// TestMockDOMEvent tests the MockDOMEvent implementation
func TestMockDOMEvent(t *testing.T) {
	target := NewMockDOMElement("button")
	mock := NewMockDOMEvent("click", target)
	
	if mock.Type() != "click" {
		t.Errorf("Expected type 'click', got %s", mock.Type())
	}
	
	if mock.Target() == nil {
		t.Error("Expected target to be set")
	}
	
	mock.PreventDefault()
	if !mock.IsDefaultPrevented() {
		t.Error("Expected default prevented to be true")
	}
	
	mock.StopPropagation()
	if !mock.IsPropagationStopped() {
		t.Error("Expected propagation stopped to be true")
	}
}

// TestMockDOMStyle tests the MockDOMStyle implementation
func TestMockDOMStyle(t *testing.T) {
	mock := NewMockDOMStyle()
	
	// Test setting and getting properties
	mock.SetProperty("color", "red")
	if mock.GetPropertyValue("color") != "red" {
		t.Errorf("Expected color 'red', got %s", mock.GetPropertyValue("color"))
	}
	
	mock.RemoveProperty("color")
	if mock.GetPropertyValue("color") != "" {
		t.Errorf("Expected empty color after removal, got %s", mock.GetPropertyValue("color"))
	}
	
	// Test Get and Set methods
	mock.Set("font-size", "14px")
	if mock.Get("font-size") != "14px" {
		t.Errorf("Expected font-size '14px', got %s", mock.Get("font-size"))
	}
	
	mock.Remove("font-size")
	if mock.Get("font-size") != "" {
		t.Errorf("Expected empty font-size after removal, got %s", mock.Get("font-size"))
	}
}

// TestMockDOMElement tests the MockDOMElement implementation
func TestMockDOMElement(t *testing.T) {
	mock := NewMockDOMElement("div")
	mock.SetAttribute("id", "test-id")
	mock.SetAttribute("class", "test-class")
	mock.SetTextContent("Hello World")
	mock.SetInnerHTML("<span>Hello World</span>")
	
	if mock.TagName() != "div" {
		t.Errorf("Expected tag name 'div', got %s", mock.TagName())
	}
	
	if mock.GetAttribute("id") != "test-id" {
		t.Errorf("Expected ID 'test-id', got %s", mock.GetAttribute("id"))
	}
	
	mock.SetAttribute("id", "new-id")
	if mock.GetAttribute("id") != "new-id" {
		t.Errorf("Expected ID 'new-id', got %s", mock.GetAttribute("id"))
	}
	
	if mock.GetAttribute("class") != "test-class" {
		t.Errorf("Expected class name 'test-class', got %s", mock.GetAttribute("class"))
	}
	
	mock.SetAttribute("class", "new-class")
	if mock.GetAttribute("class") != "new-class" {
		t.Errorf("Expected class name 'new-class', got %s", mock.GetAttribute("class"))
	}
	
	if mock.TextContent() != "Hello World" {
		t.Errorf("Expected text content 'Hello World', got %s", mock.TextContent())
	}
	
	mock.SetTextContent("New Text")
	if mock.TextContent() != "New Text" {
		t.Errorf("Expected text content 'New Text', got %s", mock.TextContent())
	}
	
	if mock.InnerHTML() != "<span>Hello World</span>" {
		t.Errorf("Expected innerHTML '<span>Hello World</span>', got %s", mock.InnerHTML())
	}
	
	mock.SetInnerHTML("<p>New HTML</p>")
	if mock.InnerHTML() != "<p>New HTML</p>" {
		t.Errorf("Expected innerHTML '<p>New HTML</p>', got %s", mock.InnerHTML())
	}
	
	// Test attributes
	mock.SetAttribute("data-test", "value")
	if mock.GetAttribute("data-test") != "value" {
		t.Errorf("Expected attribute 'value', got %s", mock.GetAttribute("data-test"))
	}
	
	if !mock.HasAttribute("data-test") {
		t.Error("Expected HasAttribute to return true")
	}
	
	mock.RemoveAttribute("data-test")
	if mock.HasAttribute("data-test") {
		t.Error("Expected HasAttribute to return false after removal")
	}
	
	// Test style
	style := mock.Style()
	if style == nil {
		t.Error("Expected Style to return a value")
	}
	
	// Test event listeners
	listenerCalled := false
	listener := func(event bridge.DOMEvent) {
		listenerCalled = true
	}
	
	mock.AddEventListener("click", listener)
	// Note: EventListeners is not exposed, so we can't test the internal state
	
	mock.RemoveEventListener("click", listener)
	// Note: EventListeners is not exposed, so we can't test the internal state
	
	// Test dispatch event
	event := NewMockDOMEvent("click", mock)
	mock.AddEventListener("click", listener)
	mock.DispatchEvent(event)
	if !listenerCalled {
		t.Error("Expected event listener to be called")
	}
	
	// Test query selector
	child := NewMockDOMElement("span")
	child.SetAttribute("class", "child")
	mock.AppendChild(child)
	
	found := mock.QuerySelector(".child")
	if found == nil {
		t.Error("Expected QuerySelector to find element")
	}
	
	all := mock.QuerySelectorAll(".child")
	if len(all) != 1 {
		t.Errorf("Expected QuerySelectorAll to find 1 element, got %d", len(all))
	}
	
	// Test parent/child relationships
	// Note: Children is not exposed, so we test through other methods
	found2 := mock.QuerySelector(".child")
	if found2 == nil {
		t.Error("Expected child to still be found after second append")
	}
	
	mock.RemoveChild(child)
	found3 := mock.QuerySelector(".child")
	if found3 != nil {
		t.Error("Expected child to be removed")
	}
	
	// Test focus/blur
	mock.Focus()
	// Note: Focused is not exposed, so we can't test the internal state
	
	mock.Blur()
	// Note: Focused is not exposed, so we can't test the internal state
	
	// Test click
	mock.Click()
	// No error expected
}

// TestMockDOMDocument tests the MockDOMDocument implementation
func TestMockDOMDocument(t *testing.T) {
	mock := NewMockDOMDocument()
	
	// Test element creation
	elem := mock.CreateElement("div")
	if elem == nil {
		t.Error("Expected CreateElement to return an element")
	}
	
	if elem.TagName() != "div" {
		t.Errorf("Expected tag name 'div', got %s", elem.TagName())
	}
	
	// Test getElementById
	testElem := mock.CreateElement("div")
	testElem.SetID("test")
	mock.AddElement("test", testElem.(*MockDOMElement))
	// Also add to body so QuerySelectorAll can find it
	mock.Body().AppendChild(testElem)
	
	found := mock.GetElementByID("test")
	if found == nil {
		t.Error("Expected GetElementByID to find element")
	}
	
	if found.GetAttribute("id") != "test" {
		t.Errorf("Expected ID 'test', got %s", found.GetAttribute("id"))
	}
	
	notFound := mock.GetElementByID("nonexistent")
	if notFound != nil {
		t.Errorf("Expected GetElementByID to return nil for nonexistent element, got: %v", notFound)
	}
	
	// Test querySelector
	queryResult := mock.QuerySelector("#test")
	if queryResult == nil {
		t.Error("Expected QuerySelector to find element")
	}
	
	// Test querySelectorAll
	allResults := mock.QuerySelectorAll("div")
	if len(allResults) == 0 {
		t.Error("Expected QuerySelectorAll to find elements")
	}
	
	// Test body and head
	body := mock.Body()
	if body == nil {
		t.Error("Expected Body to return body element")
	}
	
	if body.TagName() != "body" {
		t.Errorf("Expected body tag name 'body', got %s", body.TagName())
	}
	
	head := mock.Head()
	if head == nil {
		t.Error("Expected Head to return head element")
	}
	
	if head.TagName() != "head" {
		t.Errorf("Expected head tag name 'head', got %s", head.TagName())
	}
	
	// Test title
	mock.SetTitle("Test Title")
	if mock.Title() != "Test Title" {
		t.Errorf("Expected title 'Test Title', got %s", mock.Title())
	}
	
	// Test ready state
	mock.SetReadyState("complete")
	if mock.ReadyState() != "complete" {
		t.Errorf("Expected ready state 'complete', got %s", mock.ReadyState())
	}
	
	jsValue := mock.JSValue()
	if jsValue == nil {
		t.Error("Expected JSValue to return a value")
	}
}

// TestMockDOMBridge tests the MockDOMBridge implementation
func TestMockDOMBridge(t *testing.T) {
	mock := NewMockDOMBridge()
	
	doc := mock.GetDocument()
	if doc == nil {
		t.Error("Expected GetDocument to return a document")
	}
	
	elem := mock.QuerySelector("body")
	if elem == nil {
		t.Error("Expected QuerySelector to find body element")
	}
	
	// Test GetElementByID
	testElem := NewMockDOMElement("div")
	testElem.SetAttribute("id", "test")
	mock.document.elements["test"] = testElem
	
	found := mock.GetElementByID("test")
	if found == nil {
		t.Error("Expected GetElementByID to find element")
	}
	
	if found.GetAttribute("id") != "test" {
		t.Errorf("Expected ID 'test', got %s", found.GetAttribute("id"))
	}
}

// TestMockComponentBridge tests the MockComponentBridge implementation
func TestMockComponentBridge(t *testing.T) {
	mock := NewMockComponentBridge()
	
	// Test component initialization
	selector := "#test-component"
	componentType := "test-component"
	config := map[string]any{"prop": "value"}
	
	err := mock.InitializeComponent(selector, componentType, config)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	
	if _, exists := mock.GetComponent(selector); !exists {
		t.Error("Expected component to be stored")
	}
	
	// Test component destruction
	err = mock.DestroyComponent(selector, componentType)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	
	if _, exists := mock.GetComponent(selector); exists {
		t.Error("Expected component to be removed")
	}
	
	// Test GetComponentInstance
	_, err = mock.GetComponentInstance(selector, componentType)
	if err == nil {
		t.Error("Expected error for non-existent component")
	}
	
	// Test InitializeAll
	err = mock.InitializeAll([]string{"comp1", "comp2"})
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}

// TestMockManager tests the MockManager implementation
func TestMockManager(t *testing.T) {
	mock := NewMockManager()
	
	jsBridge := mock.JS()
	if jsBridge == nil {
		t.Error("Expected JS to return a bridge")
	}
	
	domBridge := mock.DOM()
	if domBridge == nil {
		t.Error("Expected DOM to return a bridge")
	}
	
	componentBridge := mock.Component()
	if componentBridge == nil {
		t.Error("Expected Component to return a bridge")
	}
	
	// Test utility methods
	strings := []string{"a", "b", "c"}
	jsArray := mock.GoStringsToJSArray(strings)
	if jsArray == nil {
		t.Error("Expected GoStringsToJSArray to return a value")
	}
	
	goMap := map[string]any{"key": "value"}
	jsObj := mock.GoMapToJSObject(goMap)
	if jsObj == nil {
		t.Error("Expected GoMapToJSObject to return a value")
	}
	
	// Test conversion back
	convertedStrings := mock.JSArrayToGoStrings(jsArray)
	if len(convertedStrings) != len(strings) {
		t.Errorf("Expected %d strings, got %d", len(strings), len(convertedStrings))
	}
	
	convertedMap := mock.JSObjectToGoMap(jsObj)
	if len(convertedMap) != len(goMap) {
		t.Errorf("Expected %d map entries, got %d", len(goMap), len(convertedMap))
	}
}

// TestSetupMockEnvironment tests the setup and teardown functions
func TestSetupMockEnvironment(t *testing.T) {
	// Setup mock environment
	SetupMockEnvironment()
	
	// Verify that the global manager is set
	manager := bridge.GetManager()
	if manager == nil {
		t.Error("Expected global manager to be set after setup")
	}
	
	// Test that we can use the mock environment
	doc, err := bridge.GetDocument()
	if err != nil {
		t.Errorf("Expected no error from GetDocument, got %v", err)
	}
	if doc == nil {
		t.Error("Expected GetDocument to return a document")
	}
	
	elem, err := bridge.QuerySelector("body")
	if err != nil {
		t.Errorf("Expected no error from QuerySelector, got %v", err)
	}
	if elem == nil {
		t.Error("Expected QuerySelector to find body element")
	}
	
	// Teardown mock environment
	TeardownMockEnvironment()
	
	// Verify that the global manager is cleared
	manager = bridge.GetManager()
	if manager != nil {
		t.Error("Expected global manager to be cleared after teardown")
	}
}

// TestMockEnvironmentIsolation tests that mock environments are isolated
func TestMockEnvironmentIsolation(t *testing.T) {
	// Setup first environment
	SetupMockEnvironment()
	manager1 := bridge.GetManager()
	
	// Setup second environment (should replace first)
	SetupMockEnvironment()
	manager2 := bridge.GetManager()
	
	// Managers should be different instances
	if manager1 == manager2 {
		t.Error("Expected different manager instances")
	}
	
	// Teardown
	TeardownMockEnvironment()
}