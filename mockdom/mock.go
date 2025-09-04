//go:build !js || !wasm

package mockdom

import (
	"fmt"
	"strings"
	"sync"

	"github.com/ozanturksever/uiwgo/bridge"
	"github.com/ozanturksever/uiwgo/logutil"
)

// MockJSValue implements bridge.JSValue for testing
type MockJSValue struct {
	mu         sync.RWMutex
	value      interface{}
	properties map[string]*MockJSValue
	methods    map[string]func(args ...*MockJSValue) *MockJSValue
	type_      bridge.StubType
}

// NewMockJSValue creates a new mock JS value
func NewMockJSValue(value interface{}) *MockJSValue {
	mock := &MockJSValue{
		value:      value,
		properties: make(map[string]*MockJSValue),
		methods:    make(map[string]func(args ...*MockJSValue) *MockJSValue),
	}
	
	// Set type based on value
	switch value.(type) {
	case nil:
		mock.type_ = bridge.StubTypeUndefined
	case bool:
		mock.type_ = bridge.StubTypeBoolean
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64:
		mock.type_ = bridge.StubTypeNumber
	case string:
		mock.type_ = bridge.StubTypeString
	default:
		mock.type_ = bridge.StubTypeObject
	}
	
	return mock
}

// Type returns the type of the JS value
func (m *MockJSValue) Type() bridge.StubType {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.type_
}

// IsUndefined returns true if the value is undefined
func (m *MockJSValue) IsUndefined() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.value == nil
}

// IsNull returns true if the value is null
func (m *MockJSValue) IsNull() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.value == nil
}

// Bool returns the boolean value
func (m *MockJSValue) Bool() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if b, ok := m.value.(bool); ok {
		return b
	}
	return false
}

// Int returns the integer value
func (m *MockJSValue) Int() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if i, ok := m.value.(int); ok {
		return i
	}
	return 0
}

// Float returns the float value
func (m *MockJSValue) Float() float64 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if f, ok := m.value.(float64); ok {
		return f
	}
	return 0.0
}

// String returns the string value
func (m *MockJSValue) String() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if s, ok := m.value.(string); ok {
		return s
	}
	return fmt.Sprintf("%v", m.value)
}

// Get retrieves a property
func (m *MockJSValue) Get(key string) bridge.JSValue {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if prop, exists := m.properties[key]; exists {
		return prop
	}
	return NewMockJSValue(nil)
}

// Set sets a property
func (m *MockJSValue) Set(key string, value interface{}) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if jsVal, ok := value.(bridge.JSValue); ok {
		m.properties[key] = jsVal.(*MockJSValue)
	} else {
		m.properties[key] = NewMockJSValue(value)
	}
}

// Call calls a method
func (m *MockJSValue) Call(method string, args ...interface{}) bridge.JSValue {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	if fn, exists := m.methods[method]; exists {
		mockArgs := make([]*MockJSValue, len(args))
		for i, arg := range args {
			if jsVal, ok := arg.(bridge.JSValue); ok {
				mockArgs[i] = jsVal.(*MockJSValue)
			} else {
				mockArgs[i] = NewMockJSValue(arg)
			}
		}
		return fn(mockArgs...)
	}
	
	logutil.Logf("Mock method %s called with %d args", method, len(args))
	return NewMockJSValue(nil)
}

// Index retrieves an array element by index
func (m *MockJSValue) Index(i int) bridge.JSValue {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	// If the value is a slice, try to access the element
	if slice, ok := m.value.([]interface{}); ok {
		if i >= 0 && i < len(slice) {
			return NewMockJSValue(slice[i])
		}
	}
	
	return NewMockJSValue(nil)
}

// Length returns the length property
func (m *MockJSValue) Length() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	// If the value is a slice, return its length
	if slice, ok := m.value.([]interface{}); ok {
		return len(slice)
	}
	
	// If the value is a string, return its length
	if str, ok := m.value.(string); ok {
		return len(str)
	}
	
	return 0
}

// Raw returns the underlying stub value
func (m *MockJSValue) Raw() bridge.StubValue {
	return bridge.StubValue{}
}

// SetMethod sets a mock method
func (m *MockJSValue) SetMethod(name string, fn func(args ...*MockJSValue) *MockJSValue) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.methods[name] = fn
}

// MockJSBridge implements bridge.JSBridge for testing
type MockJSBridge struct {
	mu      sync.RWMutex
	globals map[string]*MockJSValue
}

// FuncOf creates a function wrapper for Go functions
func (m *MockJSBridge) FuncOf(fn func(this bridge.JSValue, args []bridge.JSValue) interface{}) bridge.JSValue {
	mockFunc := NewMockJSValue(fn)
	mockFunc.SetMethod("call", func(args ...*MockJSValue) *MockJSValue {
		// Convert mock args to bridge.JSValue interface
		var bridgeArgs []bridge.JSValue
		if len(args) > 1 {
			bridgeArgs = make([]bridge.JSValue, len(args)-1) // Skip 'this' argument
			for i := 1; i < len(args); i++ {
				bridgeArgs[i-1] = args[i]
			}
		} else {
			bridgeArgs = []bridge.JSValue{}
		}
		
		// Call the original function with mock 'this' and args
		var thisArg bridge.JSValue
		if len(args) > 0 {
			thisArg = args[0]
		} else {
			thisArg = NewMockJSValue(nil)
		}
		
		result := fn(thisArg, bridgeArgs)
		return NewMockJSValue(result)
	})
	return mockFunc
}

// Null returns a null JS value
func (m *MockJSBridge) Null() bridge.JSValue {
	return NewMockJSValue(nil)
}

// Undefined returns an undefined JS value
func (m *MockJSBridge) Undefined() bridge.JSValue {
	return NewMockJSValue(nil)
}

// NewMockJSBridge creates a new mock JS bridge
func NewMockJSBridge() *MockJSBridge {
	return &MockJSBridge{
		globals: make(map[string]*MockJSValue),
	}
}

// Global returns the global object
func (m *MockJSBridge) Global() bridge.JSValue {
	return NewMockJSValue("global")
}

// GetGlobal returns a global JS value
func (m *MockJSBridge) GetGlobal(name string) bridge.JSValue {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if global, exists := m.globals[name]; exists {
		return global
	}
	return NewMockJSValue(nil)
}

// SetGlobal sets a global JS value
func (m *MockJSBridge) SetGlobal(name string, value interface{}) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if jsVal, ok := value.(bridge.JSValue); ok {
		m.globals[name] = jsVal.(*MockJSValue)
	} else {
		m.globals[name] = NewMockJSValue(value)
	}
}

// ValueOf converts a Go value to a JS value
func (m *MockJSBridge) ValueOf(value interface{}) bridge.JSValue {
	return NewMockJSValue(value)
}

// MockDOMElement implements bridge.DOMElement for testing
type MockDOMElement struct {
	mu         sync.RWMutex
	tagName    string
	id         string
	className  string
	textContent string
	innerHTML  string
	attributes map[string]string
	style      *MockDOMStyle
	children   []*MockDOMElement
	parent     *MockDOMElement
	eventListeners map[string][]func(bridge.DOMEvent)
}

// NewMockDOMElement creates a new mock DOM element
func NewMockDOMElement(tagName string) *MockDOMElement {
	return &MockDOMElement{
		tagName:        tagName,
		attributes:     make(map[string]string),
		style:          NewMockDOMStyle(),
		children:       make([]*MockDOMElement, 0),
		eventListeners: make(map[string][]func(bridge.DOMEvent)),
	}
}

// TagName returns the tag name
func (m *MockDOMElement) TagName() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.tagName
}

// ID returns the element ID
func (m *MockDOMElement) ID() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.id
}

// SetID sets the element ID
func (m *MockDOMElement) SetID(id string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.id = id
	m.attributes["id"] = id
}

// ClassName returns the class name
func (m *MockDOMElement) ClassName() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.className
}

// SetClassName sets the class name
func (m *MockDOMElement) SetClassName(className string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.className = className
	m.attributes["class"] = className
}

// TextContent returns the text content
func (m *MockDOMElement) TextContent() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.textContent
}

// SetTextContent sets the text content
func (m *MockDOMElement) SetTextContent(content string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.textContent = content
}

// InnerHTML returns the inner HTML
func (m *MockDOMElement) InnerHTML() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.innerHTML
}

// SetInnerHTML sets the inner HTML
func (m *MockDOMElement) SetInnerHTML(html string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.innerHTML = html
}

// GetAttribute returns an attribute value
func (m *MockDOMElement) GetAttribute(name string) string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.attributes[name]
}

// SetAttribute sets an attribute value
func (m *MockDOMElement) SetAttribute(name, value string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.attributes[name] = value
	
	// Synchronize special attributes with fields
	if name == "class" {
		m.className = value
	} else if name == "id" {
		m.id = value
	}
}

// RemoveAttribute removes an attribute
func (m *MockDOMElement) RemoveAttribute(name string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.attributes, name)
	
	// Synchronize special attributes with fields
	if name == "class" {
		m.className = ""
	} else if name == "id" {
		m.id = ""
	}
}

// HasAttribute checks if an attribute exists
func (m *MockDOMElement) HasAttribute(name string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	_, exists := m.attributes[name]
	return exists
}

// DispatchEvent dispatches an event to this element
func (m *MockDOMElement) DispatchEvent(event bridge.DOMEvent) {
	m.mu.RLock()
	listeners := m.eventListeners[event.Type()]
	m.mu.RUnlock()
	
	for _, listener := range listeners {
		listener(event)
	}
}

// Style returns the element style
func (m *MockDOMElement) Style() bridge.DOMStyle {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.style
}

// QuerySelector finds a child element by selector
func (m *MockDOMElement) QuerySelector(selector string) bridge.DOMElement {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	// Simple selector implementation for testing
	for _, child := range m.children {
		if m.matchesSelector(child, selector) {
			return child
		}
		// Recursively search children
		if result := child.QuerySelector(selector); result != nil {
			return result
		}
	}
	return nil
}

// QuerySelectorAll finds all child elements by selector
func (m *MockDOMElement) QuerySelectorAll(selector string) []bridge.DOMElement {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	var results []bridge.DOMElement
	for _, child := range m.children {
		if m.matchesSelector(child, selector) {
			results = append(results, child)
		}
		// Recursively search children
		childResults := child.QuerySelectorAll(selector)
		results = append(results, childResults...)
	}
	return results
}

// AppendChild appends a child element

// AddEventListener adds an event listener
func (m *MockDOMElement) AddEventListener(eventType string, listener func(bridge.DOMEvent)) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.eventListeners[eventType] = append(m.eventListeners[eventType], listener)
}

// RemoveEventListener removes an event listener (simplified for testing)
func (m *MockDOMElement) RemoveEventListener(eventType string, listener func(bridge.DOMEvent)) {
	m.mu.Lock()
	defer m.mu.Unlock()
	// For testing, we'll just clear all listeners of this type
	delete(m.eventListeners, eventType)
}

// GetElementByID finds a child element by ID
func (m *MockDOMElement) GetElementByID(id string) bridge.DOMElement {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	for _, child := range m.children {
		if child.ID() == id {
			return child
		}
		// Recursively search children
		if result := child.GetElementByID(id); result != nil {
			return result
		}
	}
	return nil
}

// GetElementsByTagName finds child elements by tag name
func (m *MockDOMElement) GetElementsByTagName(tagName string) []bridge.DOMElement {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	var results []bridge.DOMElement
	for _, child := range m.children {
		if strings.ToLower(child.TagName()) == strings.ToLower(tagName) {
			results = append(results, child)
		}
		// Recursively search children
		childResults := child.GetElementsByTagName(tagName)
		results = append(results, childResults...)
	}
	return results
}

// GetElementsByClassName finds child elements by class name
func (m *MockDOMElement) GetElementsByClassName(className string) []bridge.DOMElement {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	var results []bridge.DOMElement
	for _, child := range m.children {
		if child.HasClass(className) {
			results = append(results, child)
		}
		// Recursively search children
		childResults := child.GetElementsByClassName(className)
		results = append(results, childResults...)
	}
	return results
}

// AppendChild appends a child element
func (m *MockDOMElement) AppendChild(child bridge.DOMElement) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if mockChild, ok := child.(*MockDOMElement); ok {
		m.children = append(m.children, mockChild)
		mockChild.parent = m
	}
}

// RemoveChild removes a child element
func (m *MockDOMElement) RemoveChild(child bridge.DOMElement) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	if mockChild, ok := child.(*MockDOMElement); ok {
		for i, c := range m.children {
			if c == mockChild {
				m.children = append(m.children[:i], m.children[i+1:]...)
				break
			}
		}
	}
}

// Remove removes the element from the DOM
func (m *MockDOMElement) Remove() {
	if m.parent != nil {
		m.parent.RemoveChild(m)
	}
}

// Clone clones the element
func (m *MockDOMElement) Clone(deep bool) bridge.DOMElement {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	clone := NewMockDOMElement(m.tagName)
	clone.id = m.id
	clone.className = m.className
	clone.textContent = m.textContent
	clone.innerHTML = m.innerHTML
	
	// Copy attributes
	for k, v := range m.attributes {
		clone.attributes[k] = v
	}
	
	if deep {
		for _, child := range m.children {
			clone.AppendChild(child.Clone(true))
		}
	}
	
	return clone
}

// IsVisible returns true if element is visible
func (m *MockDOMElement) IsVisible() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	// For mock purposes, assume visible unless display is none
	display := m.style.GetPropertyValue("display")
	return display != "none"
}

// Raw returns the underlying DOM element
func (m *MockDOMElement) Raw() interface{} {
	return m
}

// Parent returns the parent element
func (m *MockDOMElement) Parent() bridge.DOMElement {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.parent
}

// Children returns child elements
func (m *MockDOMElement) Children() []bridge.DOMElement {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	children := make([]bridge.DOMElement, len(m.children))
	for i, child := range m.children {
		children[i] = child
	}
	return children
}

// Click triggers a click event
func (m *MockDOMElement) Click() {
	event := NewMockDOMEvent("click", m)
	m.TriggerEvent("click", event)
}

// Focus sets focus to the element
func (m *MockDOMElement) Focus() {
	// For mock purposes, just trigger focus event
	event := NewMockDOMEvent("focus", m)
	m.TriggerEvent("focus", event)
}

// Blur removes focus from the element
func (m *MockDOMElement) Blur() {
	// For mock purposes, just trigger blur event
	event := NewMockDOMEvent("blur", m)
	m.TriggerEvent("blur", event)
}

// Value returns the value (for form elements)
func (m *MockDOMElement) Value() string {
	return m.GetAttribute("value")
}

// SetValue sets the value (for form elements)
func (m *MockDOMElement) SetValue(value string) {
	m.SetAttribute("value", value)
}

// HasClass checks if element has a CSS class
func (m *MockDOMElement) HasClass(className string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	classes := strings.Fields(m.className)
	for _, class := range classes {
		if class == className {
			return true
		}
	}
	return false
}

// AddClass adds a CSS class
func (m *MockDOMElement) AddClass(className string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	if !m.hasClassUnsafe(className) {
		if m.className == "" {
			m.className = className
		} else {
			m.className += " " + className
		}
	}
}

// RemoveClass removes a CSS class
func (m *MockDOMElement) RemoveClass(className string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	classes := strings.Fields(m.className)
	var newClasses []string
	for _, class := range classes {
		if class != className {
			newClasses = append(newClasses, class)
		}
	}
	m.className = strings.Join(newClasses, " ")
}

// ToggleClass toggles a CSS class
func (m *MockDOMElement) ToggleClass(className string) {
	if m.HasClass(className) {
		m.RemoveClass(className)
	} else {
		m.AddClass(className)
	}
}

// hasClassUnsafe checks if element has a CSS class without locking (for internal use)
func (m *MockDOMElement) hasClassUnsafe(className string) bool {
	classes := strings.Fields(m.className)
	for _, class := range classes {
		if class == className {
			return true
		}
	}
	return false
}

// TriggerEvent triggers an event for testing
func (m *MockDOMElement) TriggerEvent(eventType string, event bridge.DOMEvent) {
	m.mu.RLock()
	listeners := m.eventListeners[eventType]
	m.mu.RUnlock()
	
	for _, listener := range listeners {
		listener(event)
	}
}

// matchesSelector checks if an element matches a simple selector
func (m *MockDOMElement) matchesSelector(element *MockDOMElement, selector string) bool {
	selector = strings.TrimSpace(selector)
	
	// Simple implementations for common selectors
	if strings.HasPrefix(selector, "#") {
		// ID selector
		return element.id == selector[1:]
	}
	if strings.HasPrefix(selector, ".") {
		// Class selector
		return strings.Contains(element.className, selector[1:])
	}
	// Tag selector
	return strings.ToLower(element.tagName) == strings.ToLower(selector)
}

// MockDOMStyle implements bridge.DOMStyle for testing
type MockDOMStyle struct {
	mu         sync.RWMutex
	properties map[string]string
}

// NewMockDOMStyle creates a new mock DOM style
func NewMockDOMStyle() *MockDOMStyle {
	return &MockDOMStyle{
		properties: make(map[string]string),
	}
}

// GetPropertyValue returns a CSS property value
func (m *MockDOMStyle) GetPropertyValue(property string) string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.properties[property]
}

// Get gets a CSS property value (bridge.DOMStyle interface)
func (m *MockDOMStyle) Get(property string) string {
	return m.GetPropertyValue(property)
}

// SetProperty sets a CSS property value
func (m *MockDOMStyle) SetProperty(property, value string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.properties[property] = value
}

// Set sets a CSS property (bridge.DOMStyle interface)
func (m *MockDOMStyle) Set(property, value string) {
	m.SetProperty(property, value)
}

// RemoveProperty removes a CSS property
func (m *MockDOMStyle) RemoveProperty(property string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.properties, property)
}

// Remove removes a CSS property (bridge.DOMStyle interface)
func (m *MockDOMStyle) Remove(property string) {
	m.RemoveProperty(property)
}

// MockDOMEvent implements bridge.DOMEvent for testing
type MockDOMEvent struct {
	mu           sync.RWMutex
	eventType    string
	target       bridge.DOMElement
	currentTarget bridge.DOMElement
	prevented    bool
	stopped      bool
}

// NewMockDOMEvent creates a new mock DOM event
func NewMockDOMEvent(eventType string, target bridge.DOMElement) *MockDOMEvent {
	return &MockDOMEvent{
		eventType:     eventType,
		target:        target,
		currentTarget: target,
	}
}

// Type returns the event type
func (m *MockDOMEvent) Type() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.eventType
}

// Target returns the event target
func (m *MockDOMEvent) Target() bridge.DOMElement {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.target
}

// CurrentTarget returns the current event target
func (m *MockDOMEvent) CurrentTarget() bridge.DOMElement {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.currentTarget
}

// PreventDefault prevents the default action
func (m *MockDOMEvent) PreventDefault() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.prevented = true
}

// StopPropagation stops event propagation
func (m *MockDOMEvent) StopPropagation() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.stopped = true
}

// IsDefaultPrevented returns true if default was prevented
func (m *MockDOMEvent) IsDefaultPrevented() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.prevented
}

// IsPropagationStopped returns true if propagation was stopped
func (m *MockDOMEvent) IsPropagationStopped() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.stopped
}

// MockDOMDocument implements bridge.DOMDocument for testing
type MockDOMDocument struct {
	mu        sync.RWMutex
	body      *MockDOMElement
	elements  map[string]*MockDOMElement // ID -> Element mapping
	listeners map[string][]func(bridge.DOMEvent)
	title     string
	readyState string
}

// NewMockDOMDocument creates a new mock DOM document
func NewMockDOMDocument() *MockDOMDocument {
	body := NewMockDOMElement("body")
	return &MockDOMDocument{
		body:      body,
		elements:  make(map[string]*MockDOMElement),
		listeners: make(map[string][]func(bridge.DOMEvent)),
		title:     "",
		readyState: "complete",
	}
}

// Body returns the document body
func (m *MockDOMDocument) Body() bridge.DOMElement {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.body
}

// Head returns the document head
func (m *MockDOMDocument) Head() bridge.DOMElement {
	// For mock purposes, create a simple head element
	return NewMockDOMElement("head")
}

// ReadyState returns the document ready state
func (m *MockDOMDocument) ReadyState() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.readyState
}

// GetElementByID returns an element by ID
func (m *MockDOMDocument) GetElementByID(id string) bridge.DOMElement {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if element, exists := m.elements[id]; exists {
		return element
	}
	return nil
}

// QuerySelector finds an element by selector
func (m *MockDOMDocument) QuerySelector(selector string) bridge.DOMElement {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	// Handle special case for body selector
	if selector == "body" {
		return m.body
	}
	
	// Check if it's an ID selector
	if strings.HasPrefix(selector, "#") {
		id := strings.TrimPrefix(selector, "#")
		if elem, exists := m.elements[id]; exists {
			return elem
		}
	}
	
	// Delegate to body for other selectors
	return m.body.QuerySelector(selector)
}

// QuerySelectorAll finds all elements by selector
func (m *MockDOMDocument) QuerySelectorAll(selector string) []bridge.DOMElement {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	// Handle special case for body selector
	if selector == "body" {
		return []bridge.DOMElement{m.body}
	}
	
	// Check if it's an ID selector
	if strings.HasPrefix(selector, "#") {
		id := strings.TrimPrefix(selector, "#")
		if elem, exists := m.elements[id]; exists {
			return []bridge.DOMElement{elem}
		}
		return []bridge.DOMElement{}
	}
	
	// Delegate to body for other selectors
	return m.body.QuerySelectorAll(selector)
}

// CreateElement creates a new element
func (m *MockDOMDocument) CreateElement(tagName string) bridge.DOMElement {
	return NewMockDOMElement(tagName)
}

// AddElement adds an element to the document (for testing)
func (m *MockDOMDocument) AddElement(id string, element *MockDOMElement) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.elements[id] = element
	element.SetID(id)
}

// AddEventListener adds an event listener to the document
func (m *MockDOMDocument) AddEventListener(eventType string, listener func(bridge.DOMEvent)) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.listeners[eventType] = append(m.listeners[eventType], listener)
}

// RemoveEventListener removes an event listener from the document
func (m *MockDOMDocument) RemoveEventListener(eventType string, listener func(bridge.DOMEvent)) {
	m.mu.Lock()
	defer m.mu.Unlock()
	// Note: In a real implementation, you'd need to compare function pointers
	// For mock purposes, we'll just clear all listeners of this type
	delete(m.listeners, eventType)
}

// CreateTextNode creates a text node
func (m *MockDOMDocument) CreateTextNode(text string) bridge.DOMElement {
	textNode := NewMockDOMElement("#text")
	textNode.SetTextContent(text)
	return textNode
}

// GetElementsByTagName finds elements by tag name
func (m *MockDOMDocument) GetElementsByTagName(tagName string) []bridge.DOMElement {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.body.QuerySelectorAll(tagName)
}

// GetElementsByClassName finds elements by class name
func (m *MockDOMDocument) GetElementsByClassName(className string) []bridge.DOMElement {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.body.QuerySelectorAll("." + className)
}

// Title returns the document title
func (m *MockDOMDocument) Title() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.title
}

// SetTitle sets the document title
func (m *MockDOMDocument) SetTitle(title string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.title = title
}

// SetReadyState sets the document ready state
func (m *MockDOMDocument) SetReadyState(state string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.readyState = state
}

// JSValue returns the underlying JS value
func (m *MockDOMDocument) JSValue() bridge.JSValue {
	return NewMockJSValue(m)
}

// URL returns the document URL
func (m *MockDOMDocument) URL() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return "http://localhost:3000"
}

// Raw returns the underlying document
func (m *MockDOMDocument) Raw() interface{} {
	return m
}

// MockDOMBridge implements bridge.DOMBridge for testing
type MockDOMBridge struct {
	mu       sync.RWMutex
	document *MockDOMDocument
}

// NewMockDOMBridge creates a new mock DOM bridge
func NewMockDOMBridge() *MockDOMBridge {
	return &MockDOMBridge{
		document: NewMockDOMDocument(),
	}
}

// GetDocument returns the document
func (m *MockDOMBridge) GetDocument() bridge.DOMDocument {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.document
}

// QuerySelector finds an element by selector
func (m *MockDOMBridge) QuerySelector(selector string) bridge.DOMElement {
	return m.document.QuerySelector(selector)
}

// GetElementByID returns an element by ID
func (m *MockDOMBridge) GetElementByID(id string) bridge.DOMElement {
	return m.document.GetElementByID(id)
}

// Window returns the window object as JSValue
func (m *MockDOMBridge) Window() bridge.JSValue {
	return NewMockJSValue(map[string]interface{}{
		"location": map[string]interface{}{
			"href": "http://localhost:3000",
		},
	})
}

// CreateEvent creates a new event
func (m *MockDOMBridge) CreateEvent(eventType string) bridge.DOMEvent {
	return NewMockDOMEvent(eventType, nil)
}

// DispatchEvent dispatches an event
func (m *MockDOMBridge) DispatchEvent(target bridge.DOMElement, event bridge.DOMEvent) {
	// For mock purposes, we'll just trigger the event on the target
	if mockTarget, ok := target.(*MockDOMElement); ok {
		mockTarget.TriggerEvent(event.Type(), event)
	}
}

// Document returns the document
func (m *MockDOMBridge) Document() bridge.DOMDocument {
	return m.document
}

// MockComponentBridge implements bridge.ComponentBridge for testing
type MockComponentBridge struct {
	mu         sync.RWMutex
	components map[string]interface{}
}

// NewMockComponentBridge creates a new mock component bridge
func NewMockComponentBridge() *MockComponentBridge {
	return &MockComponentBridge{
		components: make(map[string]interface{}),
	}
}

// InitializeComponent initializes a component
func (m *MockComponentBridge) InitializeComponent(selector string, componentType string, config map[string]any) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	key := selector + ":" + componentType
	m.components[key] = componentType
	logutil.Logf("Mock: Initialized component %s", selector)
	return nil
}

// DestroyComponent destroys a component
func (m *MockComponentBridge) DestroyComponent(id string, componentType string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	key := id + ":" + componentType
	delete(m.components, key)
	logutil.Logf("Mock: Destroyed component %s", id)
	return nil
}

// GetComponent gets a component by ID (checks all component types for this ID)
func (m *MockComponentBridge) GetComponent(id string) (interface{}, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	// Check if there's any component with this ID (regardless of type)
	for key, component := range m.components {
		if strings.HasPrefix(key, id+":") {
			return component, true
		}
	}
	return nil, false
}

// GetComponentInstance gets a component instance
func (m *MockComponentBridge) GetComponentInstance(selector, componentType string) (bridge.JSValue, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	key := selector + ":" + componentType
	if component, exists := m.components[key]; exists {
		return NewMockJSValue(component), nil
	}
	return nil, fmt.Errorf("component not found: %s", key)
}

// InitializeAll initializes all components
func (m *MockComponentBridge) InitializeAll(components []string) error {
	// For mock purposes, just return success
	return nil
}

// MockManager implements bridge.Manager for testing
type MockManager struct {
	mu            sync.RWMutex
	jsBridge      *MockJSBridge
	domBridge     *MockDOMBridge
	componentBridge *MockComponentBridge
}

// NewMockManager creates a new mock manager
func NewMockManager() *MockManager {
	return &MockManager{
		jsBridge:        NewMockJSBridge(),
		domBridge:       NewMockDOMBridge(),
		componentBridge: NewMockComponentBridge(),
	}
}

// JS returns the JS bridge (bridge.Manager interface)
func (m *MockManager) JS() bridge.JSBridge {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.jsBridge
}

// JSBridge returns the JS bridge
func (m *MockManager) JSBridge() bridge.JSBridge {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.jsBridge
}

// DOM returns the DOM bridge (bridge.Manager interface)
func (m *MockManager) DOM() bridge.DOMBridge {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.domBridge
}

// DOMBridge returns the DOM bridge
func (m *MockManager) DOMBridge() bridge.DOMBridge {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.domBridge
}

// Component returns the component bridge (bridge.Manager interface)
func (m *MockManager) Component() bridge.ComponentBridge {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.componentBridge
}

// ComponentBridge returns the component bridge
func (m *MockManager) ComponentBridge() bridge.ComponentBridge {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.componentBridge
}

// GoStringsToJSArray converts Go strings to JS array
func (m *MockManager) GoStringsToJSArray(strings []string) bridge.JSValue {
	return NewMockJSValue(strings)
}

// GoMapToJSObject converts Go map to JS object
func (m *MockManager) GoMapToJSObject(goMap map[string]any) bridge.JSValue {
	return NewMockJSValue(goMap)
}

// JSArrayToGoStrings converts JS array to Go strings
func (m *MockManager) JSArrayToGoStrings(jsArray bridge.JSValue) []string {
	if mockValue, ok := jsArray.(*MockJSValue); ok {
		if slice, ok := mockValue.value.([]string); ok {
			return slice
		}
	}
	return []string{}
}

// JSObjectToGoMap converts JS object to Go map
func (m *MockManager) JSObjectToGoMap(jsObj bridge.JSValue) map[string]any {
	if mockValue, ok := jsObj.(*MockJSValue); ok {
		if goMap, ok := mockValue.value.(map[string]any); ok {
			return goMap
		}
	}
	return map[string]any{}
}

// Initialize initializes the mock manager
func (m *MockManager) Initialize() error {
	logutil.Log("Mock manager initialized")
	return nil
}

// Cleanup cleans up the mock manager
func (m *MockManager) Cleanup() error {
	logutil.Log("Mock manager cleaned up")
	return nil
}

// SetupMockEnvironment sets up a complete mock environment for testing
func SetupMockEnvironment() *MockManager {
	manager := NewMockManager()
	bridge.SetManager(manager)
	logutil.Log("Mock environment set up")
	return manager
}

// TeardownMockEnvironment tears down the mock environment
func TeardownMockEnvironment() {
	bridge.ResetManager()
	logutil.Log("Mock environment torn down")
}