//go:build js && wasm

package dom

import (
	"fmt"
	"sync"
	"syscall/js"

	reactivity "github.com/ozanturksever/uiwgo/reactivity"
	"honnef.co/go/js/dom/v2"
)

// ReactiveElement wraps a DOM element with reactive capabilities
type ReactiveElement struct {
	element dom.Element
	effects []reactivity.Effect
	cleanupFuncs []func()
	mu sync.RWMutex
}

// NewReactiveElement creates a new reactive element wrapper
func NewReactiveElement(element dom.Element) *ReactiveElement {
	return &ReactiveElement{
		element: element,
		effects: make([]reactivity.Effect, 0),
		cleanupFuncs: make([]func(), 0),
	}
}

// WrapElement wraps an existing DOM element with reactive capabilities
func WrapElement(element dom.Element) *ReactiveElement {
	return NewReactiveElement(element)
}

// Element returns the underlying DOM element
func (re *ReactiveElement) Element() dom.Element {
	return re.element
}

// BindText creates a reactive text binding
func (re *ReactiveElement) BindText(textSignal reactivity.Signal[string]) *ReactiveElement {
	re.mu.Lock()
	defer re.mu.Unlock()

	// Set initial value
	re.element.SetTextContent(textSignal.Get())

	// Create reactive effect
	effect := reactivity.CreateEffect(func() {
		re.element.SetTextContent(textSignal.Get())
	})

	re.effects = append(re.effects, effect)
	return re
}

// BindTextFunc creates a reactive text binding using a function
func (re *ReactiveElement) BindTextFunc(textFn func() string) *ReactiveElement {
	re.mu.Lock()
	defer re.mu.Unlock()

	// Set initial value
	re.element.SetTextContent(textFn())

	// Create reactive effect
	effect := reactivity.CreateEffect(func() {
		re.element.SetTextContent(textFn())
	})

	re.effects = append(re.effects, effect)
	return re
}

// BindHTML creates a reactive HTML binding
func (re *ReactiveElement) BindHTML(htmlSignal reactivity.Signal[string]) *ReactiveElement {
	re.mu.Lock()
	defer re.mu.Unlock()

	// Set initial value
	re.element.SetInnerHTML(htmlSignal.Get())

	// Create reactive effect
	effect := reactivity.CreateEffect(func() {
		re.element.SetInnerHTML(htmlSignal.Get())
	})

	re.effects = append(re.effects, effect)
	return re
}

// BindHTMLFunc creates a reactive HTML binding using a function
func (re *ReactiveElement) BindHTMLFunc(htmlFn func() string) *ReactiveElement {
	re.mu.Lock()
	defer re.mu.Unlock()

	// Set initial value
	re.element.SetInnerHTML(htmlFn())

	// Create reactive effect
	effect := reactivity.CreateEffect(func() {
		re.element.SetInnerHTML(htmlFn())
	})

	re.effects = append(re.effects, effect)
	return re
}

// BindAttribute creates a reactive attribute binding
func (re *ReactiveElement) BindAttribute(attrName string, valueSignal reactivity.Signal[string]) *ReactiveElement {
	re.mu.Lock()
	defer re.mu.Unlock()

	// Set initial value
	re.element.SetAttribute(attrName, valueSignal.Get())

	// Create reactive effect
	effect := reactivity.CreateEffect(func() {
		re.element.SetAttribute(attrName, valueSignal.Get())
	})

	re.effects = append(re.effects, effect)
	return re
}

// BindAttributeFunc creates a reactive attribute binding using a function
func (re *ReactiveElement) BindAttributeFunc(attrName string, valueFn func() string) *ReactiveElement {
	re.mu.Lock()
	defer re.mu.Unlock()

	// Set initial value
	re.element.SetAttribute(attrName, valueFn())

	// Create reactive effect
	effect := reactivity.CreateEffect(func() {
		re.element.SetAttribute(attrName, valueFn())
	})

	re.effects = append(re.effects, effect)
	return re
}

// BindClass creates a reactive class binding
func (re *ReactiveElement) BindClass(classSignal reactivity.Signal[string]) *ReactiveElement {
	return re.BindAttribute("class", classSignal)
}

// BindClassFunc creates a reactive class binding using a function
func (re *ReactiveElement) BindClassFunc(classFn func() string) *ReactiveElement {
	return re.BindAttributeFunc("class", classFn)
}

// BindStyle creates a reactive style binding
func (re *ReactiveElement) BindStyle(styleSignal reactivity.Signal[string]) *ReactiveElement {
	return re.BindAttribute("style", styleSignal)
}

// BindStyleFunc creates a reactive style binding using a function
func (re *ReactiveElement) BindStyleFunc(styleFn func() string) *ReactiveElement {
	return re.BindAttributeFunc("style", styleFn)
}

// BindVisibility creates a reactive visibility binding
func (re *ReactiveElement) BindVisibility(visibleSignal reactivity.Signal[bool]) *ReactiveElement {
	re.mu.Lock()
	defer re.mu.Unlock()

	// Set initial visibility
	if visibleSignal.Get() {
		re.element.SetAttribute("style", "display: block;")
	} else {
		re.element.SetAttribute("style", "display: none;")
	}

	// Create reactive effect
	effect := reactivity.CreateEffect(func() {
		if visibleSignal.Get() {
			re.element.SetAttribute("style", "display: block;")
		} else {
			re.element.SetAttribute("style", "display: none;")
		}
	})

	re.effects = append(re.effects, effect)
	return re
}

// BindVisibilityFunc creates a reactive visibility binding using a function
func (re *ReactiveElement) BindVisibilityFunc(visibleFn func() bool) *ReactiveElement {
	re.mu.Lock()
	defer re.mu.Unlock()

	// Set initial visibility
	if visibleFn() {
		re.element.SetAttribute("style", "display: block;")
	} else {
		re.element.SetAttribute("style", "display: none;")
	}

	// Create reactive effect
	effect := reactivity.CreateEffect(func() {
		if visibleFn() {
			re.element.SetAttribute("style", "display: block;")
		} else {
			re.element.SetAttribute("style", "display: none;")
		}
	})

	re.effects = append(re.effects, effect)
	return re
}

// OnClick adds a click event handler with automatic cleanup
func (re *ReactiveElement) OnClick(handler func(event dom.Event)) *ReactiveElement {
	re.mu.Lock()
	defer re.mu.Unlock()

	funcName := CreateJSFunctionOnTheFly(func(this js.Value, args []js.Value) any {
		event := dom.WrapEvent(args[0])
		handler(event)
		return nil
	})

	re.element.SetAttribute("onclick", fmt.Sprintf("%s(event)", funcName))

	// Add cleanup function
	re.cleanupFuncs = append(re.cleanupFuncs, func() {
		ReleaseJSFunction(funcName)
	})

	return re
}

// OnEvent adds a generic event handler with automatic cleanup
func (re *ReactiveElement) OnEvent(eventType string, handler func(event dom.Event)) *ReactiveElement {
	re.mu.Lock()
	defer re.mu.Unlock()

	funcName := CreateJSFunctionOnTheFly(func(this js.Value, args []js.Value) any {
		event := dom.WrapEvent(args[0])
		handler(event)
		return nil
	})

	attrName := "on" + eventType
	re.element.SetAttribute(attrName, fmt.Sprintf("%s(event)", funcName))

	// Add cleanup function
	re.cleanupFuncs = append(re.cleanupFuncs, func() {
		re.element.RemoveAttribute(attrName)
		ReleaseJSFunction(funcName)
	})

	return re
}

// AddCleanup adds a custom cleanup function
func (re *ReactiveElement) AddCleanup(cleanup func()) *ReactiveElement {
	re.mu.Lock()
	defer re.mu.Unlock()

	re.cleanupFuncs = append(re.cleanupFuncs, cleanup)
	return re
}

// Cleanup disposes all effects and cleanup functions
func (re *ReactiveElement) Cleanup() {
	re.mu.Lock()
	defer re.mu.Unlock()

	// Dispose all effects
	for _, effect := range re.effects {
		effect.Dispose()
	}
	re.effects = re.effects[:0]

	// Run all cleanup functions
	for _, cleanup := range re.cleanupFuncs {
		cleanup()
	}
	re.cleanupFuncs = re.cleanupFuncs[:0]
}

// ReactiveElementManager manages multiple reactive elements
type ReactiveElementManager struct {
	elements []*ReactiveElement
	mu sync.RWMutex
}

// NewReactiveElementManager creates a new manager
func NewReactiveElementManager() *ReactiveElementManager {
	return &ReactiveElementManager{
		elements: make([]*ReactiveElement, 0),
	}
}

// Add adds a reactive element to the manager
func (rem *ReactiveElementManager) Add(element *ReactiveElement) {
	rem.mu.Lock()
	defer rem.mu.Unlock()

	rem.elements = append(rem.elements, element)
}

// CleanupAll cleans up all managed reactive elements
func (rem *ReactiveElementManager) CleanupAll() {
	rem.mu.Lock()
	defer rem.mu.Unlock()

	for _, element := range rem.elements {
		element.Cleanup()
	}
	rem.elements = rem.elements[:0]
}

// Global reactive element manager
var GlobalReactiveManager = NewReactiveElementManager()

// Helper functions for creating reactive elements

// CreateReactiveDiv creates a reactive div element
func CreateReactiveDiv() *ReactiveElement {
	el := Document.CreateElement("div")
	reactiveEl := NewReactiveElement(el)
	GlobalReactiveManager.Add(reactiveEl)
	return reactiveEl
}

// CreateReactiveSpan creates a reactive span element
func CreateReactiveSpan() *ReactiveElement {
	el := Document.CreateElement("span")
	reactiveEl := NewReactiveElement(el)
	GlobalReactiveManager.Add(reactiveEl)
	return reactiveEl
}

// CreateReactiveButton creates a reactive button element
func CreateReactiveButton() *ReactiveElement {
	el := Document.CreateElement("button")
	reactiveEl := NewReactiveElement(el)
	GlobalReactiveManager.Add(reactiveEl)
	return reactiveEl
}

// CreateReactiveInput creates a reactive input element
func CreateReactiveInput() *ReactiveElement {
	el := Document.CreateElement("input")
	reactiveEl := NewReactiveElement(el)
	GlobalReactiveManager.Add(reactiveEl)
	return reactiveEl
}

// CreateReactiveElement creates a reactive element with the specified tag
func CreateReactiveElement(tagName string) *ReactiveElement {
	el := Document.CreateElement(tagName)
	reactiveEl := NewReactiveElement(el)
	GlobalReactiveManager.Add(reactiveEl)
	return reactiveEl
}

// CleanupAllReactiveElements cleans up all globally managed reactive elements
func CleanupAllReactiveElements() {
	GlobalReactiveManager.CleanupAll()
}