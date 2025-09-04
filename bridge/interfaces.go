package bridge

// Note: JSValue and JSBridge interfaces are defined in:
// - interfaces_wasm.go for js && wasm builds
// - interfaces_stub.go for !js || !wasm builds

// DOMEvent represents a DOM event
type DOMEvent interface {
	// Type returns the event type
	Type() string
	// Target returns the event target
	Target() DOMElement
	// PreventDefault prevents the default action
	PreventDefault()
	// StopPropagation stops event propagation
	StopPropagation()
}

// DOMElement represents a DOM element
type DOMElement interface {
	// ID returns the element ID
	ID() string
	// SetID sets the element ID
	SetID(id string)
	// TagName returns the tag name
	TagName() string
	// ClassName returns the class name
	ClassName() string
	// SetClassName sets the class name
	SetClassName(className string)
	// GetAttribute gets an attribute value
	GetAttribute(name string) string
	// SetAttribute sets an attribute value
	SetAttribute(name, value string)
	// RemoveAttribute removes an attribute
	RemoveAttribute(name string)
	// AddClass adds a CSS class
	AddClass(className string)
	// RemoveClass removes a CSS class
	RemoveClass(className string)
	// ToggleClass toggles a CSS class
	ToggleClass(className string)
	// HasClass checks if element has a CSS class
	HasClass(className string) bool
	// QuerySelector finds first descendant matching selector
	QuerySelector(selector string) DOMElement
	// QuerySelectorAll finds all descendants matching selector
	QuerySelectorAll(selector string) []DOMElement
	// AddEventListener adds an event listener
	AddEventListener(eventType string, listener func(DOMEvent))
	// RemoveEventListener removes an event listener
	RemoveEventListener(eventType string, listener func(DOMEvent))
	// Click triggers a click event
	Click()
	// Focus sets focus to the element
	Focus()
	// Blur removes focus from the element
	Blur()
	// InnerHTML returns the inner HTML
	InnerHTML() string
	// SetInnerHTML sets the inner HTML
	SetInnerHTML(html string)
	// TextContent returns the text content
	TextContent() string
	// SetTextContent sets the text content
	SetTextContent(text string)
	// Value returns the value (for form elements)
	Value() string
	// SetValue sets the value (for form elements)
	SetValue(value string)
	// Style returns the style object
	Style() DOMStyle
	// Parent returns the parent element
	Parent() DOMElement
	// Children returns child elements
	Children() []DOMElement
	// AppendChild appends a child element
	AppendChild(child DOMElement)
	// RemoveChild removes a child element
	RemoveChild(child DOMElement)
	// Remove removes the element from the DOM
	Remove()
	// Clone clones the element
	Clone(deep bool) DOMElement
	// IsVisible returns true if element is visible
	IsVisible() bool
	// Raw returns the underlying DOM element for advanced use
	Raw() interface{}
}

// DOMStyle represents element styles
type DOMStyle interface {
	// Get gets a style property
	Get(property string) string
	// Set sets a style property
	Set(property, value string)
	// Remove removes a style property
	Remove(property string)
}

// DOMDocument represents the document
type DOMDocument interface {
	// GetElementByID finds element by ID
	GetElementByID(id string) DOMElement
	// QuerySelector finds first element matching selector
	QuerySelector(selector string) DOMElement
	// QuerySelectorAll finds all elements matching selector
	QuerySelectorAll(selector string) []DOMElement
	// CreateElement creates a new element
	CreateElement(tagName string) DOMElement
	// CreateTextNode creates a text node
	CreateTextNode(text string) DOMElement
	// Body returns the body element
	Body() DOMElement
	// Head returns the head element
	Head() DOMElement
	// Title returns the document title
	Title() string
	// SetTitle sets the document title
	SetTitle(title string)
	// URL returns the document URL
	URL() string
	// ReadyState returns the document ready state
	ReadyState() string
	// AddEventListener adds a document event listener
	AddEventListener(eventType string, listener func(DOMEvent))
	// RemoveEventListener removes a document event listener
	RemoveEventListener(eventType string, listener func(DOMEvent))
}

// DOMBridge provides typed DOM access
type DOMBridge interface {
	// Document returns the document
	Document() DOMDocument
	// Window returns the window object as JSValue
	Window() JSValue
	// CreateEvent creates a new event
	CreateEvent(eventType string) DOMEvent
	// DispatchEvent dispatches an event
	DispatchEvent(target DOMElement, event DOMEvent)
}

// ComponentBridge provides generic component operations
type ComponentBridge interface {
	// InitializeComponent initializes a component by name with options
	InitializeComponent(componentName string, selector string, options map[string]any) error
	// DestroyComponent destroys a component instance
	DestroyComponent(selector string, componentType string) error
	// GetComponentInstance gets a component instance
	GetComponentInstance(selector, componentType string) (JSValue, error)
	// InitializeAll initializes multiple components
	InitializeAll(components []string) error
}

// Manager combines JS/DOM/Component bridges with utility conversions
type Manager interface {
	// JS returns the JavaScript bridge
	JS() JSBridge
	// DOM returns the DOM bridge
	DOM() DOMBridge
	// Component returns the component bridge
	Component() ComponentBridge
	// GoStringsToJSArray converts Go strings to JS array
	GoStringsToJSArray(strings []string) JSValue
	// GoMapToJSObject converts Go map to JS object
	GoMapToJSObject(m map[string]any) JSValue
	// JSArrayToGoStrings converts JS array to Go strings
	JSArrayToGoStrings(jsArray JSValue) []string
	// JSObjectToGoMap converts JS object to Go map
	JSObjectToGoMap(jsObj JSValue) map[string]any
}