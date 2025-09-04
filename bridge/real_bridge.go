//go:build js && wasm

package bridge

import (
	"errors"
	"fmt"
	"strings"
	"syscall/js"

	"honnef.co/go/js/dom/v2"
	"github.com/ozanturksever/uiwgo/logutil"
)

// RealJSValue wraps js.Value to implement JSValue interface
type RealJSValue struct {
	value js.Value
}

func NewRealJSValue(value js.Value) JSValue {
	return &RealJSValue{value: value}
}

func (r *RealJSValue) Type() js.Type                                    { return r.value.Type() }
func (r *RealJSValue) String() string                                   { return r.value.String() }
func (r *RealJSValue) Int() int                                         { return r.value.Int() }
func (r *RealJSValue) Float() float64                                   { return r.value.Float() }
func (r *RealJSValue) Bool() bool                                       { return r.value.Bool() }
func (r *RealJSValue) IsNull() bool                                     { return r.value.Type() == js.TypeNull }
func (r *RealJSValue) IsUndefined() bool                                { return r.value.Type() == js.TypeUndefined }
func (r *RealJSValue) Get(key string) JSValue                          { return NewRealJSValue(r.value.Get(key)) }
func (r *RealJSValue) Set(key string, value interface{})               { r.value.Set(key, value) }
func (r *RealJSValue) Call(method string, args ...interface{}) JSValue { return NewRealJSValue(r.value.Call(method, args...)) }
func (r *RealJSValue) Index(i int) JSValue                             { return NewRealJSValue(r.value.Index(i)) }
func (r *RealJSValue) Length() int                                      { return r.value.Length() }
func (r *RealJSValue) Raw() js.Value                                    { return r.value }

// RealJSBridge implements JSBridge using syscall/js
type RealJSBridge struct{}

func NewRealJSBridge() JSBridge {
	return &RealJSBridge{}
}

func (r *RealJSBridge) Global() JSValue {
	return NewRealJSValue(js.Global())
}

func (r *RealJSBridge) Undefined() JSValue {
	return NewRealJSValue(js.Undefined())
}

func (r *RealJSBridge) Null() JSValue {
	return NewRealJSValue(js.Null())
}

func (r *RealJSBridge) ValueOf(x interface{}) JSValue {
	return NewRealJSValue(js.ValueOf(x))
}

func (r *RealJSBridge) FuncOf(fn func(this JSValue, args []JSValue) interface{}) JSValue {
	// Convert the function to work with js.Value
	jsFn := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		// Convert js.Value args to JSValue args
		jsValueArgs := make([]JSValue, len(args))
		for i, arg := range args {
			jsValueArgs[i] = NewRealJSValue(arg)
		}
		return fn(NewRealJSValue(this), jsValueArgs)
	})
	return NewRealJSValue(jsFn.Value)
}

// RealDOMEvent wraps dom.Event to implement DOMEvent interface
type RealDOMEvent struct {
	event dom.Event
}

func NewRealDOMEvent(event dom.Event) DOMEvent {
	return &RealDOMEvent{event: event}
}

func (r *RealDOMEvent) Type() string {
	return r.event.Type()
}

func (r *RealDOMEvent) Target() DOMElement {
	if target := r.event.Target(); target != nil {
		if element, ok := target.(dom.Element); ok {
			return NewRealDOMElement(element)
		}
	}
	return nil
}

func (r *RealDOMEvent) PreventDefault() {
	r.event.PreventDefault()
}

func (r *RealDOMEvent) StopPropagation() {
	r.event.StopPropagation()
}

// RealDOMStyle wraps dom.CSSStyleDeclaration to implement DOMStyle interface
type RealDOMStyle struct {
	style dom.CSSStyleDeclaration
}

func NewRealDOMStyle(style dom.CSSStyleDeclaration) DOMStyle {
	return &RealDOMStyle{style: style}
}

func (r *RealDOMStyle) Get(property string) string {
	return r.style.GetPropertyValue(property)
}

func (r *RealDOMStyle) Set(property, value string) {
	r.style.SetProperty(property, value, "")
}

func (r *RealDOMStyle) Remove(property string) {
	r.style.RemoveProperty(property)
}

// RealDOMElement wraps dom.Element to implement DOMElement interface
type RealDOMElement struct {
	element dom.Element
}

func NewRealDOMElement(element dom.Element) DOMElement {
	if element == nil {
		return nil
	}
	return &RealDOMElement{element: element}
}

func (r *RealDOMElement) ID() string {
	return r.element.ID()
}

func (r *RealDOMElement) SetID(id string) {
	r.element.SetID(id)
}

func (r *RealDOMElement) TagName() string {
	return r.element.TagName()
}

func (r *RealDOMElement) ClassName() string {
	return r.element.Class().String()
}

func (r *RealDOMElement) SetClassName(className string) {
	// Use Class() method which returns a TokenList
	r.element.Class().SetString(className)
}

func (r *RealDOMElement) GetAttribute(name string) string {
	return r.element.GetAttribute(name)
}

func (r *RealDOMElement) SetAttribute(name, value string) {
	r.element.SetAttribute(name, value)
}

func (r *RealDOMElement) RemoveAttribute(name string) {
	r.element.RemoveAttribute(name)
}

func (r *RealDOMElement) AddClass(className string) {
	r.element.Class().Add(className)
}

func (r *RealDOMElement) RemoveClass(className string) {
	r.element.Class().Remove(className)
}

func (r *RealDOMElement) ToggleClass(className string) {
	r.element.Class().Toggle(className)
}

func (r *RealDOMElement) HasClass(className string) bool {
	return r.element.Class().Contains(className)
}

func (r *RealDOMElement) QuerySelector(selector string) DOMElement {
	if element := r.element.QuerySelector(selector); element != nil {
		return NewRealDOMElement(element)
	}
	return nil
}

func (r *RealDOMElement) QuerySelectorAll(selector string) []DOMElement {
	elements := r.element.QuerySelectorAll(selector)
	result := make([]DOMElement, len(elements))
	for i, element := range elements {
		result[i] = NewRealDOMElement(element)
	}
	return result
}

func (r *RealDOMElement) AddEventListener(eventType string, listener func(DOMEvent)) {
	r.element.AddEventListener(eventType, false, func(event dom.Event) {
		listener(NewRealDOMEvent(event))
	})
}

func (r *RealDOMElement) RemoveEventListener(eventType string, listener func(DOMEvent)) {
	// Note: dom/v2 doesn't provide a direct way to remove specific listeners
	// This is a limitation of the current implementation
	logutil.Logf("Warning: RemoveEventListener not fully supported for %s", eventType)
}

func (r *RealDOMElement) Click() {
	if htmlElem, ok := r.element.(dom.HTMLElement); ok {
		htmlElem.Click()
	}
}

func (r *RealDOMElement) Focus() {
	if htmlElem, ok := r.element.(dom.HTMLElement); ok {
		htmlElem.Focus()
	}
}

func (r *RealDOMElement) Blur() {
	if htmlElem, ok := r.element.(dom.HTMLElement); ok {
		htmlElem.Blur()
	}
}

func (r *RealDOMElement) InnerHTML() string {
	return r.element.InnerHTML()
}

func (r *RealDOMElement) SetInnerHTML(html string) {
	r.element.SetInnerHTML(html)
}

func (r *RealDOMElement) TextContent() string {
	return r.element.TextContent()
}

func (r *RealDOMElement) SetTextContent(text string) {
	r.element.SetTextContent(text)
}

func (r *RealDOMElement) Value() string {
	if input, ok := r.element.(dom.HTMLInputElement); ok {
		return input.Value()
	}
	if textarea, ok := r.element.(dom.HTMLTextAreaElement); ok {
		return textarea.Value()
	}
	if selectEl, ok := r.element.(dom.HTMLSelectElement); ok {
		return selectEl.Value()
	}
	return r.element.GetAttribute("value")
}

func (r *RealDOMElement) SetValue(value string) {
	if input, ok := r.element.(dom.HTMLInputElement); ok {
		input.SetValue(value)
		return
	}
	if textarea, ok := r.element.(dom.HTMLTextAreaElement); ok {
		textarea.SetValue(value)
		return
	}
	if selectEl, ok := r.element.(dom.HTMLSelectElement); ok {
		selectEl.SetValue(value)
		return
	}
	r.element.SetAttribute("value", value)
}

func (r *RealDOMElement) Style() DOMStyle {
	if htmlElem, ok := r.element.(dom.HTMLElement); ok {
		return NewRealDOMStyle(*htmlElem.Style())
	}
	// Return a stub style for non-HTML elements
	return &RealDOMStyle{}
}

func (r *RealDOMElement) Parent() DOMElement {
	if parent := r.element.ParentElement(); parent != nil {
		return NewRealDOMElement(parent)
	}
	return nil
}

func (r *RealDOMElement) Children() []DOMElement {
	// Use ChildNodes from BasicNode which is available on all elements
	childNodes := r.element.ChildNodes()
	result := make([]DOMElement, 0, len(childNodes))
	for _, child := range childNodes {
		if elem, ok := child.(dom.Element); ok {
			result = append(result, NewRealDOMElement(elem))
		}
	}
	return result
}

func (r *RealDOMElement) AppendChild(child DOMElement) {
	if realChild, ok := child.(*RealDOMElement); ok {
		r.element.AppendChild(realChild.element)
	}
}

func (r *RealDOMElement) RemoveChild(child DOMElement) {
	if realChild, ok := child.(*RealDOMElement); ok {
		r.element.RemoveChild(realChild.element)
	}
}

func (r *RealDOMElement) Remove() {
	r.element.Remove()
}

func (r *RealDOMElement) Clone(deep bool) DOMElement {
	cloned := r.element.CloneNode(deep)
	if element, ok := cloned.(dom.Element); ok {
		return NewRealDOMElement(element)
	}
	return nil
}

func (r *RealDOMElement) IsVisible() bool {
	// Simple visibility check - element exists and has non-zero dimensions
	if r.element == nil {
		return false
	}
	// Check if element has style display: none
	if htmlElem, ok := r.element.(dom.HTMLElement); ok {
		style := htmlElem.Style()
		return style != nil
	}
	return true
}

func (r *RealDOMElement) Raw() interface{} {
	return r.element
}

// RealDOMDocument wraps dom.Document to implement DOMDocument interface
type RealDOMDocument struct {
	document dom.Document
}

func NewRealDOMDocument(document dom.Document) DOMDocument {
	return &RealDOMDocument{document: document}
}

func (r *RealDOMDocument) GetElementByID(id string) DOMElement {
	if element := r.document.GetElementByID(id); element != nil {
		return NewRealDOMElement(element)
	}
	return nil
}

func (r *RealDOMDocument) QuerySelector(selector string) DOMElement {
	if element := r.document.QuerySelector(selector); element != nil {
		return NewRealDOMElement(element)
	}
	return nil
}

func (r *RealDOMDocument) QuerySelectorAll(selector string) []DOMElement {
	elements := r.document.QuerySelectorAll(selector)
	result := make([]DOMElement, len(elements))
	for i, element := range elements {
		result[i] = NewRealDOMElement(element)
	}
	return result
}

func (r *RealDOMDocument) CreateElement(tagName string) DOMElement {
	element := r.document.CreateElement(tagName)
	return NewRealDOMElement(element)
}

func (r *RealDOMDocument) CreateTextNode(text string) DOMElement {
	// Create text node using the underlying JS API
	jsDoc := r.document.Underlying()
	textNode := jsDoc.Call("createTextNode", text)
	// Wrap the JS value as an element (text nodes can be treated as nodes)
	element := dom.WrapElement(textNode)
	return NewRealDOMElement(element)
}

func (r *RealDOMDocument) Body() DOMElement {
	// Access body through the underlying JS object
	jsDoc := r.document.Underlying()
	body := jsDoc.Get("body")
	if !body.IsNull() {
		element := dom.WrapElement(body)
		return NewRealDOMElement(element)
	}
	return nil
}

func (r *RealDOMDocument) Head() DOMElement {
	// Access head through the underlying JS object
	jsDoc := r.document.Underlying()
	head := jsDoc.Get("head")
	if !head.IsNull() {
		element := dom.WrapElement(head)
		return NewRealDOMElement(element)
	}
	return nil
}

func (r *RealDOMDocument) Title() string {
	// Access title through the underlying JS object
	jsDoc := r.document.Underlying()
	return jsDoc.Get("title").String()
}

func (r *RealDOMDocument) SetTitle(title string) {
	// Set title through the underlying JS object
	jsDoc := r.document.Underlying()
	jsDoc.Set("title", title)
}

func (r *RealDOMDocument) URL() string {
	// Access URL through the underlying JS object
	jsDoc := r.document.Underlying()
	return jsDoc.Get("URL").String()
}

func (r *RealDOMDocument) ReadyState() string {
	// Access readyState through the underlying JS object
	jsDoc := r.document.Underlying()
	return jsDoc.Get("readyState").String()
}

func (r *RealDOMDocument) AddEventListener(eventType string, listener func(DOMEvent)) {
	r.document.AddEventListener(eventType, false, func(event dom.Event) {
		listener(NewRealDOMEvent(event))
	})
}

func (r *RealDOMDocument) RemoveEventListener(eventType string, listener func(DOMEvent)) {
	// Note: dom/v2 doesn't provide a direct way to remove specific listeners
	logutil.Logf("Warning: RemoveEventListener not fully supported for %s", eventType)
}

// RealDOMBridge implements DOMBridge using honnef.co/go/js/dom/v2
type RealDOMBridge struct {
	document DOMDocument
	window   JSValue
}

func NewRealDOMBridge() DOMBridge {
	return &RealDOMBridge{
		document: NewRealDOMDocument(dom.GetWindow().Document()),
		window:   NewRealJSValue(js.Global()),
	}
}

func (r *RealDOMBridge) Document() DOMDocument {
	return r.document
}

func (r *RealDOMBridge) Window() JSValue {
	return r.window
}

func (r *RealDOMBridge) CreateEvent(eventType string) DOMEvent {
	// Create event using the underlying JS API
	if doc, ok := r.document.(*RealDOMDocument); ok {
		jsDoc := doc.document.Underlying()
		event := jsDoc.Call("createEvent", eventType)
		if !event.IsNull() {
			// Wrap the JS event using dom.WrapEvent
			wrappedEvent := dom.WrapEvent(event)
			return NewRealDOMEvent(wrappedEvent)
		}
	}
	return nil
}

func (r *RealDOMBridge) DispatchEvent(target DOMElement, event DOMEvent) {
	if realTarget, ok := target.(*RealDOMElement); ok {
		if realEvent, ok := event.(*RealDOMEvent); ok {
			realTarget.element.DispatchEvent(realEvent.event)
		}
	}
}

// RealComponentBridge implements ComponentBridge for generic component operations
type RealComponentBridge struct {
	jsBridge JSBridge
}

func NewRealComponentBridge(jsBridge JSBridge) ComponentBridge {
	return &RealComponentBridge{jsBridge: jsBridge}
}

func (r *RealComponentBridge) InitializeComponent(componentName string, selector string, options map[string]any) error {
	if componentName == "" {
		return errors.New("component name cannot be empty")
	}
	if selector == "" {
		return errors.New("selector cannot be empty")
	}

	// Look up the component constructor in the global scope
	global := r.jsBridge.Global()
	componentConstructor := global.Get(componentName)
	
	if componentConstructor.IsUndefined() {
		return fmt.Errorf("component constructor '%s' not found in global scope", componentName)
	}

	// Convert options to JS object
	var jsOptions JSValue
	if options != nil {
		// Create a JS object from the map
		jsObj := r.jsBridge.ValueOf(map[string]interface{}{})
		for key, value := range options {
			jsObj.Set(key, value)
		}
		jsOptions = jsObj
	} else {
		jsOptions = r.jsBridge.ValueOf(map[string]interface{}{})
	}

	// Try to initialize the component
	// Different libraries may have different initialization patterns
	// We'll try common patterns:
	
	// Pattern 1: Constructor with selector and options
	result := componentConstructor.Call("new", selector, jsOptions)
	if !result.IsUndefined() {
		logutil.Logf("Initialized component %s with selector %s", componentName, selector)
		return nil
	}

	// Pattern 2: Static method like Component.init(selector, options)
	initMethod := componentConstructor.Get("init")
	if !initMethod.IsUndefined() {
		result = initMethod.Call("call", componentConstructor, selector, jsOptions)
		logutil.Logf("Initialized component %s with selector %s using init method", componentName, selector)
		return nil
	}

	// Pattern 3: Direct call as function
	result = componentConstructor.Call("call", global, selector, jsOptions)
	logutil.Logf("Initialized component %s with selector %s using direct call", componentName, selector)
	return nil
}

func (r *RealComponentBridge) DestroyComponent(selector string, componentType string) error {
	if selector == "" {
		return errors.New("selector cannot be empty")
	}
	if componentType == "" {
		return errors.New("component type cannot be empty")
	}

	// Look up the component constructor
	global := r.jsBridge.Global()
	componentConstructor := global.Get(componentType)
	
	if componentConstructor.IsUndefined() {
		return fmt.Errorf("component constructor '%s' not found", componentType)
	}

	// Try common destroy patterns
	// Pattern 1: Static destroy method
	destroyMethod := componentConstructor.Get("destroy")
	if !destroyMethod.IsUndefined() {
		destroyMethod.Call("call", componentConstructor, selector)
		logutil.Logf("Destroyed component %s with selector %s", componentType, selector)
		return nil
	}

	// Pattern 2: getInstance and then destroy
	getInstanceMethod := componentConstructor.Get("getInstance")
	if !getInstanceMethod.IsUndefined() {
		instance := getInstanceMethod.Call("call", componentConstructor, selector)
		if !instance.IsUndefined() {
			instanceDestroy := instance.Get("destroy")
			if !instanceDestroy.IsUndefined() {
				instanceDestroy.Call("call", instance)
				logutil.Logf("Destroyed component instance %s with selector %s", componentType, selector)
				return nil
			}
		}
	}

	logutil.Logf("Warning: Could not find destroy method for component %s", componentType)
	return nil
}

func (r *RealComponentBridge) GetComponentInstance(selector, componentType string) (JSValue, error) {
	if selector == "" {
		return nil, errors.New("selector cannot be empty")
	}
	if componentType == "" {
		return nil, errors.New("component type cannot be empty")
	}

	// Look up the component constructor
	global := r.jsBridge.Global()
	componentConstructor := global.Get(componentType)
	
	if componentConstructor.IsUndefined() {
		return nil, fmt.Errorf("component constructor '%s' not found", componentType)
	}

	// Try to get instance using common patterns
	getInstanceMethod := componentConstructor.Get("getInstance")
	if !getInstanceMethod.IsUndefined() {
		instance := getInstanceMethod.Call("call", componentConstructor, selector)
		if !instance.IsUndefined() {
			return instance, nil
		}
	}

	return nil, fmt.Errorf("could not get instance of component %s with selector %s", componentType, selector)
}

func (r *RealComponentBridge) InitializeAll(components []string) error {
	var errors []string
	
	for _, componentName := range components {
		// Use a generic selector that targets elements with data attributes
		selector := fmt.Sprintf("[data-%s]", strings.ToLower(componentName))
		
		if err := r.InitializeComponent(componentName, selector, nil); err != nil {
			errorMsg := fmt.Sprintf("failed to initialize %s: %v", componentName, err)
			errors = append(errors, errorMsg)
			logutil.Logf("Error: %s", errorMsg)
		}
	}
	
	if len(errors) > 0 {
		return fmt.Errorf("failed to initialize some components: %s", strings.Join(errors, "; "))
	}
	
	return nil
}

// RealManager implements Manager combining all bridges
type RealManager struct {
	jsBridge        JSBridge
	domBridge       DOMBridge
	componentBridge ComponentBridge
}

// NewRealManager creates a new real manager for js/wasm builds
func NewRealManager() Manager {
	jsBridge := NewRealJSBridge()
	domBridge := NewRealDOMBridge()
	componentBridge := NewRealComponentBridge(jsBridge)
	
	return &RealManager{
		jsBridge:        jsBridge,
		domBridge:       domBridge,
		componentBridge: componentBridge,
	}
}

func (r *RealManager) JS() JSBridge {
	return r.jsBridge
}

func (r *RealManager) DOM() DOMBridge {
	return r.domBridge
}

func (r *RealManager) Component() ComponentBridge {
	return r.componentBridge
}

func (r *RealManager) GoStringsToJSArray(strings []string) JSValue {
	// Convert Go string slice to JS array
	jsArray := r.jsBridge.ValueOf([]interface{}{})
	for i, str := range strings {
		jsArray.Index(i).Set("value", str)
	}
	return jsArray
}

func (r *RealManager) GoMapToJSObject(m map[string]any) JSValue {
	// Convert Go map to JS object
	jsObj := r.jsBridge.ValueOf(map[string]interface{}{})
	for key, value := range m {
		jsObj.Set(key, value)
	}
	return jsObj
}

func (r *RealManager) JSArrayToGoStrings(jsArray JSValue) []string {
	if jsArray.IsUndefined() || jsArray.IsNull() {
		return []string{}
	}
	
	length := jsArray.Length()
	result := make([]string, length)
	for i := 0; i < length; i++ {
		result[i] = jsArray.Index(i).String()
	}
	return result
}

func (r *RealManager) JSObjectToGoMap(jsObj JSValue) map[string]any {
	if jsObj.IsUndefined() || jsObj.IsNull() {
		return map[string]any{}
	}
	
	// This is a simplified implementation
	// In a real scenario, you'd need to iterate over object keys
	// which requires more complex JS interop
	result := make(map[string]any)
	
	// For now, return empty map as this requires more complex implementation
	logutil.Log("Warning: JSObjectToGoMap is not fully implemented")
	return result
}