//go:build js && wasm

package dom

import (
	"bytes"
	"fmt"
	"syscall/js"

	reactivity "github.com/ozanturksever/uiwgo/reactivity"
	"honnef.co/go/js/dom/v2"
	g "maragu.dev/gomponents"
)

// GomponentsRenderer renders gomponents nodes to DOM using dom/v2
type GomponentsRenderer struct {
	container        dom.Element
	reactiveElements []*ReactiveElement
}

// NewGomponentsRenderer creates a new renderer for the given container
func NewGomponentsRenderer(container dom.Element) *GomponentsRenderer {
	return &GomponentsRenderer{
		container:        container,
		reactiveElements: make([]*ReactiveElement, 0),
	}
}

// Render renders a gomponents node tree to the DOM
func (gr *GomponentsRenderer) Render(node g.Node) error {
	// First, render to HTML string (existing approach)
	var buf bytes.Buffer
	if err := node.Render(&buf); err != nil {
		return err
	}

	// Set the HTML content
	gr.container.SetInnerHTML(buf.String())

	return nil
}

// RenderReactive renders a gomponents node tree with reactive capabilities
func (gr *GomponentsRenderer) RenderReactive(node g.Node) error {
	// Render the static structure first
	if err := gr.Render(node); err != nil {
		return err
	}

	// Scan for reactive elements and enhance them
	gr.enhanceReactiveElements()

	return nil
}

// enhanceReactiveElements scans the rendered DOM for reactive markers and enhances them
func (gr *GomponentsRenderer) enhanceReactiveElements() {
	// Find elements with reactive data attributes
	textElements := gr.container.QuerySelectorAll("[data-uiwgo-txt]")
	for _, el := range textElements {
		gr.enhanceTextElement(el)
	}

	htmlElements := gr.container.QuerySelectorAll("[data-uiwgo-html]")
	for _, el := range htmlElements {
		gr.enhanceHTMLElement(el)
	}

	showElements := gr.container.QuerySelectorAll("[data-uiwgo-show]")
	for _, el := range showElements {
		gr.enhanceShowElement(el)
	}

	// Control flow components
	forElements := gr.container.QuerySelectorAll("[data-uiwgo-for]")
	for _, el := range forElements {
		gr.enhanceForElement(el)
	}

	indexElements := gr.container.QuerySelectorAll("[data-uiwgo-index]")
	for _, el := range indexElements {
		gr.enhanceIndexElement(el)
	}

	switchElements := gr.container.QuerySelectorAll("[data-uiwgo-switch]")
	for _, el := range switchElements {
		gr.enhanceSwitchElement(el)
	}

	dynamicElements := gr.container.QuerySelectorAll("[data-uiwgo-dynamic]")
	for _, el := range dynamicElements {
		gr.enhanceDynamicElement(el)
	}
}

// enhanceTextElement enhances a text element with reactive capabilities
func (gr *GomponentsRenderer) enhanceTextElement(el dom.Element) {
	id := el.GetAttribute("data-uiwgo-txt")
	if id == "" {
		return
	}

	// Create reactive element wrapper
	reactiveEl := WrapElement(el)
	gr.reactiveElements = append(gr.reactiveElements, reactiveEl)

	// Mark as enhanced to avoid duplicate processing
	el.SetAttribute("data-uiwgo-enhanced", "true")
}

// enhanceHTMLElement enhances an HTML element with reactive capabilities
func (gr *GomponentsRenderer) enhanceHTMLElement(el dom.Element) {
	id := el.GetAttribute("data-uiwgo-html")
	if id == "" {
		return
	}

	// Create reactive element wrapper
	reactiveEl := WrapElement(el)
	gr.reactiveElements = append(gr.reactiveElements, reactiveEl)

	// Mark as enhanced to avoid duplicate processing
	el.SetAttribute("data-uiwgo-enhanced", "true")
}

// enhanceShowElement enhances a show element with reactive capabilities
func (gr *GomponentsRenderer) enhanceShowElement(el dom.Element) {
	id := el.GetAttribute("data-uiwgo-show")
	if id == "" {
		return
	}

	// Create reactive element wrapper
	reactiveEl := WrapElement(el)
	gr.reactiveElements = append(gr.reactiveElements, reactiveEl)

	// Mark as enhanced to avoid duplicate processing
	el.SetAttribute("data-uiwgo-enhanced", "true")
}

// enhanceForElement enhances a For element with reactive capabilities
func (gr *GomponentsRenderer) enhanceForElement(el dom.Element) {
	id := el.GetAttribute("data-uiwgo-for")
	if id == "" {
		return
	}

	// Create reactive element wrapper
	reactiveEl := WrapElement(el)
	gr.reactiveElements = append(gr.reactiveElements, reactiveEl)

	// Mark as enhanced to avoid duplicate processing
	el.SetAttribute("data-uiwgo-enhanced", "true")
}

// enhanceIndexElement enhances an Index element with reactive capabilities
func (gr *GomponentsRenderer) enhanceIndexElement(el dom.Element) {
	id := el.GetAttribute("data-uiwgo-index")
	if id == "" {
		return
	}

	// Create reactive element wrapper
	reactiveEl := WrapElement(el)
	gr.reactiveElements = append(gr.reactiveElements, reactiveEl)

	// Mark as enhanced to avoid duplicate processing
	el.SetAttribute("data-uiwgo-enhanced", "true")
}

// enhanceSwitchElement enhances a Switch element with reactive capabilities
func (gr *GomponentsRenderer) enhanceSwitchElement(el dom.Element) {
	id := el.GetAttribute("data-uiwgo-switch")
	if id == "" {
		return
	}

	// Create reactive element wrapper
	reactiveEl := WrapElement(el)
	gr.reactiveElements = append(gr.reactiveElements, reactiveEl)

	// Mark as enhanced to avoid duplicate processing
	el.SetAttribute("data-uiwgo-enhanced", "true")
}

// enhanceDynamicElement enhances a Dynamic element with reactive capabilities
func (gr *GomponentsRenderer) enhanceDynamicElement(el dom.Element) {
	id := el.GetAttribute("data-uiwgo-dynamic")
	if id == "" {
		return
	}

	// Create reactive element wrapper
	reactiveEl := WrapElement(el)
	gr.reactiveElements = append(gr.reactiveElements, reactiveEl)

	// Mark as enhanced to avoid duplicate processing
	el.SetAttribute("data-uiwgo-enhanced", "true")
}

// Cleanup cleans up all reactive elements managed by this renderer
func (gr *GomponentsRenderer) Cleanup() {
	for _, reactiveEl := range gr.reactiveElements {
		reactiveEl.Cleanup()
	}
	gr.reactiveElements = gr.reactiveElements[:0]
}

// Enhanced Mount function that uses dom/v2
func MountWithDOM(elementID string, root func() g.Node) (*GomponentsRenderer, error) {
	container := GetElementByID(elementID)
	if container == nil {
		return nil, fmt.Errorf("element with id '%s' not found", elementID)
	}

	renderer := NewGomponentsRenderer(container)
	if err := renderer.RenderReactive(root()); err != nil {
		return nil, err
	}

	return renderer, nil
}

// ReactiveNodeBuilder provides a fluent interface for building reactive DOM nodes
type ReactiveNodeBuilder struct {
	element  *ReactiveElement
	children []*ReactiveNodeBuilder
	scope    *reactivity.CleanupScope
}

// NewReactiveNode creates a new reactive node builder
func NewReactiveNode(tagName string) *ReactiveNodeBuilder {
	// Use current cleanup scope as parent, or create root scope if none exists
	parentScope := reactivity.GetCurrentCleanupScope()
	scope := reactivity.NewCleanupScope(parentScope)
	return &ReactiveNodeBuilder{
		element:  CreateReactiveElement(tagName),
		children: make([]*ReactiveNodeBuilder, 0),
		scope:    scope,
	}
}

// SetText sets the text content
func (rnb *ReactiveNodeBuilder) SetText(text string) *ReactiveNodeBuilder {
	rnb.element.Element().SetTextContent(text)
	return rnb
}

// BindText binds reactive text content
func (rnb *ReactiveNodeBuilder) BindText(textSignal reactivity.Signal[string]) *ReactiveNodeBuilder {
	// Temporarily set this node's scope as current
	prevScope := reactivity.GetCurrentCleanupScope()
	reactivity.SetCurrentCleanupScope(rnb.scope)
	defer reactivity.SetCurrentCleanupScope(prevScope)
	
	rnb.element.BindText(textSignal)
	return rnb
}

// BindTextFunc binds reactive text content using a function
func (rnb *ReactiveNodeBuilder) BindTextFunc(textFn func() string) *ReactiveNodeBuilder {
	// Temporarily set this node's scope as current
	prevScope := reactivity.GetCurrentCleanupScope()
	reactivity.SetCurrentCleanupScope(rnb.scope)
	defer reactivity.SetCurrentCleanupScope(prevScope)
	
	rnb.element.BindTextFunc(textFn)
	return rnb
}

// SetAttribute sets an attribute
func (rnb *ReactiveNodeBuilder) SetAttribute(name, value string) *ReactiveNodeBuilder {
	rnb.element.Element().SetAttribute(name, value)
	return rnb
}

// BindAttribute binds a reactive attribute
func (rnb *ReactiveNodeBuilder) BindAttribute(name string, valueSignal reactivity.Signal[string]) *ReactiveNodeBuilder {
	// Temporarily set this node's scope as current
	prevScope := reactivity.GetCurrentCleanupScope()
	reactivity.SetCurrentCleanupScope(rnb.scope)
	defer reactivity.SetCurrentCleanupScope(prevScope)
	
	rnb.element.BindAttribute(name, valueSignal)
	return rnb
}

// BindAttributeFunc binds a reactive attribute using a function
func (rnb *ReactiveNodeBuilder) BindAttributeFunc(name string, valueFn func() string) *ReactiveNodeBuilder {
	// Temporarily set this node's scope as current
	prevScope := reactivity.GetCurrentCleanupScope()
	reactivity.SetCurrentCleanupScope(rnb.scope)
	defer reactivity.SetCurrentCleanupScope(prevScope)
	
	rnb.element.BindAttributeFunc(name, valueFn)
	return rnb
}

// SetClass sets the class attribute
func (rnb *ReactiveNodeBuilder) SetClass(className string) *ReactiveNodeBuilder {
	return rnb.SetAttribute("class", className)
}

// BindClass binds a reactive class
func (rnb *ReactiveNodeBuilder) BindClass(classSignal reactivity.Signal[string]) *ReactiveNodeBuilder {
	// Temporarily set this node's scope as current
	prevScope := reactivity.GetCurrentCleanupScope()
	reactivity.SetCurrentCleanupScope(rnb.scope)
	defer reactivity.SetCurrentCleanupScope(prevScope)
	
	rnb.element.BindClass(classSignal)
	return rnb
}

// BindClassFunc binds a reactive class using a function
func (rnb *ReactiveNodeBuilder) BindClassFunc(classFn func() string) *ReactiveNodeBuilder {
	// Temporarily set this node's scope as current
	prevScope := reactivity.GetCurrentCleanupScope()
	reactivity.SetCurrentCleanupScope(rnb.scope)
	defer reactivity.SetCurrentCleanupScope(prevScope)
	
	rnb.element.BindClassFunc(classFn)
	return rnb
}

// SetStyle sets the style attribute
func (rnb *ReactiveNodeBuilder) SetStyle(style string) *ReactiveNodeBuilder {
	return rnb.SetAttribute("style", style)
}

// BindStyle binds a reactive style
func (rnb *ReactiveNodeBuilder) BindStyle(styleSignal reactivity.Signal[string]) *ReactiveNodeBuilder {
	// Temporarily set this node's scope as current
	prevScope := reactivity.GetCurrentCleanupScope()
	reactivity.SetCurrentCleanupScope(rnb.scope)
	defer reactivity.SetCurrentCleanupScope(prevScope)
	
	rnb.element.BindStyle(styleSignal)
	return rnb
}

// BindStyleFunc binds a reactive style using a function
func (rnb *ReactiveNodeBuilder) BindStyleFunc(styleFn func() string) *ReactiveNodeBuilder {
	// Temporarily set this node's scope as current
	prevScope := reactivity.GetCurrentCleanupScope()
	reactivity.SetCurrentCleanupScope(rnb.scope)
	defer reactivity.SetCurrentCleanupScope(prevScope)
	
	rnb.element.BindStyleFunc(styleFn)
	return rnb
}

// OnClick adds a click event handler
func (rnb *ReactiveNodeBuilder) OnClick(handler func(event dom.Event)) *ReactiveNodeBuilder {
	rnb.element.OnClick(handler)
	return rnb
}

// OnEvent adds a generic event handler
func (rnb *ReactiveNodeBuilder) OnEvent(eventType string, handler func(event dom.Event)) *ReactiveNodeBuilder {
	rnb.element.OnEvent(eventType, handler)
	return rnb
}

// AppendChild appends a child node
func (rnb *ReactiveNodeBuilder) AppendChild(child *ReactiveNodeBuilder) *ReactiveNodeBuilder {
	rnb.element.Element().AppendChild(child.element.Element())
	rnb.children = append(rnb.children, child)
	
	// Establish parent-child scope relationship
	if child.scope != nil && rnb.scope != nil {
		child.scope.SetParent(rnb.scope)
	}
	
	return rnb
}

// Build returns the built reactive element
func (rnb *ReactiveNodeBuilder) Build() *ReactiveElement {
	return rnb.element
}

// BuildElement returns the underlying DOM element
func (rnb *ReactiveNodeBuilder) BuildElement() dom.Element {
	return rnb.element.Element()
}

// Cleanup cleans up the node and all its children
func (rnb *ReactiveNodeBuilder) Cleanup() {
	// Dispose the scope (this will handle all cleanup automatically)
	if rnb.scope != nil {
		rnb.scope.Dispose()
	}

	// Clean up children and element for backward compatibility
	for _, child := range rnb.children {
		child.Cleanup()
	}
	rnb.element.Cleanup()
}

// GetScope returns the cleanup scope for this reactive node builder
func (rnb *ReactiveNodeBuilder) GetScope() *reactivity.CleanupScope {
	return rnb.scope
}

// Helper functions for common HTML elements

// Div creates a reactive div element
func Div() *ReactiveNodeBuilder {
	return NewReactiveNode("div")
}

// Span creates a reactive span element
func Span() *ReactiveNodeBuilder {
	return NewReactiveNode("span")
}

// Button creates a reactive button element
func Button() *ReactiveNodeBuilder {
	return NewReactiveNode("button")
}

// Input creates a reactive input element
func Input() *ReactiveNodeBuilder {
	return NewReactiveNode("input")
}

// P creates a reactive paragraph element
func P() *ReactiveNodeBuilder {
	return NewReactiveNode("p")
}

// H1 creates a reactive h1 element
func H1() *ReactiveNodeBuilder {
	return NewReactiveNode("h1")
}

// H2 creates a reactive h2 element
func H2() *ReactiveNodeBuilder {
	return NewReactiveNode("h2")
}

// H3 creates a reactive h3 element
func H3() *ReactiveNodeBuilder {
	return NewReactiveNode("h3")
}

// Ul creates a reactive ul element
func Ul() *ReactiveNodeBuilder {
	return NewReactiveNode("ul")
}

// Li creates a reactive li element
func Li() *ReactiveNodeBuilder {
	return NewReactiveNode("li")
}

// A creates a reactive anchor element
func A() *ReactiveNodeBuilder {
	return NewReactiveNode("a")
}

// Img creates a reactive img element
func Img() *ReactiveNodeBuilder {
	return NewReactiveNode("img")
}

// Form creates a reactive form element
func Form() *ReactiveNodeBuilder {
	return NewReactiveNode("form")
}

// Label creates a reactive label element
func Label() *ReactiveNodeBuilder {
	return NewReactiveNode("label")
}

// Textarea creates a reactive textarea element
func Textarea() *ReactiveNodeBuilder {
	return NewReactiveNode("textarea")
}

// Select creates a reactive select element
func Select() *ReactiveNodeBuilder {
	return NewReactiveNode("select")
}

// Option creates a reactive option element
func Option() *ReactiveNodeBuilder {
	return NewReactiveNode("option")
}

// Compatibility functions for existing code

// CreateJSFunctionForEvent creates a JS function specifically for event handling
func CreateJSFunctionForEvent(handler func(event dom.Event)) string {
	return CreateJSFunctionOnTheFly(func(this js.Value, args []js.Value) any {
		if len(args) > 0 {
			event := dom.WrapEvent(args[0])
			handler(event)
		}
		return nil
	})
}

// CreateJSFunctionForSignal creates a JS function that updates a signal
func CreateJSFunctionForSignal[T any](signal reactivity.Signal[T], value T) string {
	return CreateJSFunctionOnTheFly(func(this js.Value, args []js.Value) any {
		signal.Set(value)
		return nil
	})
}

// CreateJSFunctionForCallback creates a JS function that calls a Go callback
func CreateJSFunctionForCallback(callback func()) string {
	return CreateJSFunctionOnTheFly(func(this js.Value, args []js.Value) any {
		callback()
		return nil
	})
}
