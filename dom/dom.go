//go:build js && wasm

package dom

import (
	"fmt"
	"sync"
	"syscall/js"

	reactivity "github.com/ozanturksever/uiwgo/reactivity"
	"honnef.co/go/js/dom/v2"
)

// Document provides a wrapper around the global document
var Document = dom.GetWindow().Document()

// Window provides a wrapper around the global window
var Window = dom.GetWindow()

// JSFunctionManager manages JavaScript functions created on-the-fly
type JSFunctionManager struct {
	mu        sync.RWMutex
	functions map[string]js.Func
	counter   int
}

// Global function manager instance
var FunctionManager = &JSFunctionManager{
	functions: make(map[string]js.Func),
}

// CreateJSFunction creates a JavaScript function on-the-fly and registers it globally
// Returns the function name that can be used in event handlers
func (fm *JSFunctionManager) CreateJSFunction(name string, fn func(this js.Value, args []js.Value) any) string {
	fm.mu.Lock()
	defer fm.mu.Unlock()

	// If name is empty, generate one
	if name == "" {
		fm.counter++
		name = fmt.Sprintf("uiwgo_fn_%d", fm.counter)
	}

	// Clean up existing function if it exists
	if existing, exists := fm.functions[name]; exists {
		existing.Release()
	}

	// Create new JS function
	jsFunc := js.FuncOf(fn)
	fm.functions[name] = jsFunc

	// Register globally
	js.Global().Set(name, jsFunc)

	return name
}

// ReleaseJSFunction releases a JavaScript function and removes it from global scope
func (fm *JSFunctionManager) ReleaseJSFunction(name string) {
	fm.mu.Lock()
	defer fm.mu.Unlock()

	if jsFunc, exists := fm.functions[name]; exists {
		jsFunc.Release()
		delete(fm.functions, name)
		js.Global().Delete(name)
	}
}

// ReleaseAll releases all managed JavaScript functions
func (fm *JSFunctionManager) ReleaseAll() {
	fm.mu.Lock()
	defer fm.mu.Unlock()

	for name, jsFunc := range fm.functions {
		jsFunc.Release()
		js.Global().Delete(name)
	}
	fm.functions = make(map[string]js.Func)
}

// ElementBuilder provides a fluent interface for creating DOM elements
type ElementBuilder struct {
	element      dom.Element
	cleanupFuncs []func()
	scope        *reactivity.CleanupScope
}

// NewElement creates a new DOM element with the specified tag name
// If there's a current cleanup scope, the element will be associated with it
func NewElement(tagName string) *ElementBuilder {
	el := Document.CreateElement(tagName)
	scope := reactivity.NewCleanupScope(reactivity.GetCurrentCleanupScope())
	return &ElementBuilder{
		element:      el,
		cleanupFuncs: make([]func(), 0),
		scope:        scope,
	}
}

// SetAttribute sets an attribute on the element
func (eb *ElementBuilder) SetAttribute(name, value string) *ElementBuilder {
	eb.element.SetAttribute(name, value)
	return eb
}

// SetClass sets the class attribute
func (eb *ElementBuilder) SetClass(className string) *ElementBuilder {
	eb.element.SetAttribute("class", className)
	return eb
}

// SetID sets the id attribute
func (eb *ElementBuilder) SetID(id string) *ElementBuilder {
	eb.element.SetAttribute("id", id)
	return eb
}

// SetStyle sets the style attribute
func (eb *ElementBuilder) SetStyle(style string) *ElementBuilder {
	eb.element.SetAttribute("style", style)
	return eb
}

// SetText sets the text content of the element
func (eb *ElementBuilder) SetText(text string) *ElementBuilder {
	eb.element.SetTextContent(text)
	return eb
}

// SetHTML sets the innerHTML of the element
func (eb *ElementBuilder) SetHTML(html string) *ElementBuilder {
	eb.element.SetInnerHTML(html)
	return eb
}

// AppendChild appends a child element
func (eb *ElementBuilder) AppendChild(child dom.Node) *ElementBuilder {
	eb.element.AppendChild(child)
	return eb
}

// AppendElement appends another ElementBuilder's element as a child
func (eb *ElementBuilder) AppendElement(child *ElementBuilder) *ElementBuilder {
	eb.element.AppendChild(child.element)
	// Transfer cleanup functions to parent
	eb.cleanupFuncs = append(eb.cleanupFuncs, child.cleanupFuncs...)
	return eb
}

// OnClick adds a click event listener with automatic cleanup
func (eb *ElementBuilder) OnClick(handler func(event dom.Event)) *ElementBuilder {
	jsFunc := js.FuncOf(func(this js.Value, args []js.Value) any {
		event := dom.WrapEvent(args[0])
		handler(event)
		return nil
	})

	// Use addEventListener instead of inline attribute
	eb.element.Underlying().Call("addEventListener", "click", jsFunc)

	// Register cleanup with scope
	if eb.scope != nil {
		eb.scope.RegisterDisposer(func() {
			eb.element.Underlying().Call("removeEventListener", "click", jsFunc)
			jsFunc.Release()
		})
	}

	return eb
}

// OnEvent adds a generic event listener with automatic cleanup
func (eb *ElementBuilder) OnEvent(eventType string, handler func(event dom.Event)) *ElementBuilder {
	jsFunc := js.FuncOf(func(this js.Value, args []js.Value) any {
		event := dom.WrapEvent(args[0])
		handler(event)
		return nil
	})

	eb.element.Underlying().Call("addEventListener", eventType, jsFunc)

	// Register cleanup with scope
	if eb.scope != nil {
		eb.scope.RegisterDisposer(func() {
			eb.element.Underlying().Call("removeEventListener", eventType, jsFunc)
			jsFunc.Release()
		})
	}

	return eb
}

// BindReactiveText binds reactive text content using a signal or computed function
func (eb *ElementBuilder) BindReactiveText(textFn func() string) *ElementBuilder {
	// Set initial text
	eb.element.SetTextContent(textFn())

	// Set the element's scope as current scope for the effect
	previous := reactivity.GetCurrentCleanupScope()
	reactivity.SetCurrentCleanupScope(eb.scope)
	
	// Create reactive effect - it will automatically register with the current scope
	reactivity.CreateEffect(func() {
		eb.element.SetTextContent(textFn())
	})
	
	// Restore previous scope
	reactivity.SetCurrentCleanupScope(previous)

	return eb
}

// BindReactiveHTML binds reactive HTML content
func (eb *ElementBuilder) BindReactiveHTML(htmlFn func() string) *ElementBuilder {
	// Set initial HTML
	eb.element.SetInnerHTML(htmlFn())

	// Set the element's scope as current scope for the effect
	previous := reactivity.GetCurrentCleanupScope()
	reactivity.SetCurrentCleanupScope(eb.scope)
	
	// Create reactive effect - it will automatically register with the current scope
	reactivity.CreateEffect(func() {
		eb.element.SetInnerHTML(htmlFn())
	})
	
	// Restore previous scope
	reactivity.SetCurrentCleanupScope(previous)

	return eb
}

// BindReactiveAttribute binds a reactive attribute value
func (eb *ElementBuilder) BindReactiveAttribute(attrName string, valueFn func() string) *ElementBuilder {
	// Set initial value
	eb.element.SetAttribute(attrName, valueFn())

	// Set the element's scope as current scope for the effect
	previous := reactivity.GetCurrentCleanupScope()
	reactivity.SetCurrentCleanupScope(eb.scope)
	
	// Create reactive effect - it will automatically register with the current scope
	reactivity.CreateEffect(func() {
		eb.element.SetAttribute(attrName, valueFn())
	})
	
	// Restore previous scope
	reactivity.SetCurrentCleanupScope(previous)

	return eb
}

// Build returns the built DOM element
func (eb *ElementBuilder) Build() dom.Element {
	// Register the element with its scope for automatic cleanup
	if eb.scope != nil {
		RegisterElementScope(eb.element.Underlying(), eb.scope)
	}
	return eb.element
}

// GetScope returns the cleanup scope associated with this element
func (eb *ElementBuilder) GetScope() *reactivity.CleanupScope {
	return eb.scope
}

// BuildWithCleanup returns the built DOM element and a cleanup function
// The cleanup function will dispose the element's scope and all associated resources
func (eb *ElementBuilder) BuildWithCleanup() (dom.Element, func()) {
	cleanup := func() {
		// Dispose the scope, which will handle all registered disposers
		if eb.scope != nil {
			eb.scope.Dispose()
		}
		// Also run legacy cleanup functions for backward compatibility
		for _, fn := range eb.cleanupFuncs {
			fn()
		}
	}
	return eb.element, cleanup
}

// Utility functions for common operations

// GetElementByID gets an element by its ID using dom/v2
func GetElementByID(id string) dom.Element {
	return Document.GetElementByID(id)
}

// QuerySelector performs a CSS selector query
func QuerySelector(selector string) dom.Element {
	return Document.QuerySelector(selector)
}

// QuerySelectorAll performs a CSS selector query returning all matches
func QuerySelectorAll(selector string) []dom.Element {
	return Document.QuerySelectorAll(selector)
}

// CreateTextNode creates a new text node
func CreateTextNode(text string) *dom.Text {
	return Document.CreateTextNode(text)
}

// CreateJSFunctionOnTheFly creates a JavaScript function and returns its name for immediate use
func CreateJSFunctionOnTheFly(fn func(this js.Value, args []js.Value) any) string {
	return FunctionManager.CreateJSFunction("", fn)
}

// CreateNamedJSFunction creates a JavaScript function with a specific name
func CreateNamedJSFunction(name string, fn func(this js.Value, args []js.Value) any) string {
	return FunctionManager.CreateJSFunction(name, fn)
}

// ReleaseJSFunction releases a JavaScript function by name
func ReleaseJSFunction(name string) {
	FunctionManager.ReleaseJSFunction(name)
}

// CleanupAllJSFunctions releases all managed JavaScript functions
func CleanupAllJSFunctions() {
	FunctionManager.ReleaseAll()
}
