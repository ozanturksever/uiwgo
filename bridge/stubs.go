//go:build !js || !wasm

package bridge

import (
	"errors"
)

var (
	// ErrNotSupported is returned when operation is not supported on this platform
	ErrNotSupported = errors.New("operation not supported on non-js/wasm platform")
)

// StubType represents a JavaScript type for non-js builds
type StubType int

const (
	StubTypeUndefined StubType = iota
	StubTypeNull
	StubTypeBoolean
	StubTypeNumber
	StubTypeString
	StubTypeSymbol
	StubTypeObject
	StubTypeFunction
)

// StubValue represents a JavaScript value for non-js builds
type StubValue struct{}

// StubJSValue implements JSValue for non-js builds
type StubJSValue struct {
	value interface{}
}

func (s *StubJSValue) Type() StubType                                     { return StubTypeUndefined }
func (s *StubJSValue) String() string                                     { return "" }
func (s *StubJSValue) Int() int                                           { return 0 }
func (s *StubJSValue) Float() float64                                     { return 0 }
func (s *StubJSValue) Bool() bool                                         { return false }
func (s *StubJSValue) IsNull() bool                                       { return true }
func (s *StubJSValue) IsUndefined() bool                                  { return true }
func (s *StubJSValue) Get(key string) JSValue                            { return &StubJSValue{} }
func (s *StubJSValue) Set(key string, value interface{})                 {}
func (s *StubJSValue) Call(method string, args ...interface{}) JSValue   { return &StubJSValue{} }
func (s *StubJSValue) Index(i int) JSValue                               { return &StubJSValue{} }
func (s *StubJSValue) Length() int                                        { return 0 }
func (s *StubJSValue) Raw() StubValue                                     { return StubValue{} }

// StubJSBridge implements JSBridge for non-js builds
type StubJSBridge struct{}

func (s *StubJSBridge) Global() JSValue                                                 { return &StubJSValue{} }
func (s *StubJSBridge) Undefined() JSValue                                              { return &StubJSValue{} }
func (s *StubJSBridge) Null() JSValue                                                   { return &StubJSValue{} }
func (s *StubJSBridge) ValueOf(x interface{}) JSValue                                   { return &StubJSValue{} }
func (s *StubJSBridge) FuncOf(fn func(this JSValue, args []JSValue) interface{}) JSValue { return &StubJSValue{} }

// StubDOMEvent implements DOMEvent for non-js builds
type StubDOMEvent struct{}

func (s *StubDOMEvent) Type() string              { return "" }
func (s *StubDOMEvent) Target() DOMElement        { return &StubDOMElement{} }
func (s *StubDOMEvent) PreventDefault()           {}
func (s *StubDOMEvent) StopPropagation()          {}

// StubDOMStyle implements DOMStyle for non-js builds
type StubDOMStyle struct{}

func (s *StubDOMStyle) Get(property string) string        { return "" }
func (s *StubDOMStyle) Set(property, value string)        {}
func (s *StubDOMStyle) Remove(property string)            {}

// StubDOMElement implements DOMElement for non-js builds
type StubDOMElement struct{}

func (s *StubDOMElement) ID() string                                         { return "" }
func (s *StubDOMElement) SetID(id string)                                   {}
func (s *StubDOMElement) TagName() string                                   { return "" }
func (s *StubDOMElement) ClassName() string                                 { return "" }
func (s *StubDOMElement) SetClassName(className string)                     {}
func (s *StubDOMElement) GetAttribute(name string) string                   { return "" }
func (s *StubDOMElement) SetAttribute(name, value string)                   {}
func (s *StubDOMElement) RemoveAttribute(name string)                       {}
func (s *StubDOMElement) AddClass(className string)                         {}
func (s *StubDOMElement) RemoveClass(className string)                      {}
func (s *StubDOMElement) ToggleClass(className string)                      {}
func (s *StubDOMElement) HasClass(className string) bool                    { return false }
func (s *StubDOMElement) QuerySelector(selector string) DOMElement          { return &StubDOMElement{} }
func (s *StubDOMElement) QuerySelectorAll(selector string) []DOMElement     { return []DOMElement{} }
func (s *StubDOMElement) AddEventListener(eventType string, listener func(DOMEvent)) {}
func (s *StubDOMElement) RemoveEventListener(eventType string, listener func(DOMEvent)) {}
func (s *StubDOMElement) Click()                                             {}
func (s *StubDOMElement) Focus()                                             {}
func (s *StubDOMElement) Blur()                                              {}
func (s *StubDOMElement) InnerHTML() string                                 { return "" }
func (s *StubDOMElement) SetInnerHTML(html string)                          {}
func (s *StubDOMElement) TextContent() string                               { return "" }
func (s *StubDOMElement) SetTextContent(text string)                        {}
func (s *StubDOMElement) Value() string                                      { return "" }
func (s *StubDOMElement) SetValue(value string)                             {}
func (s *StubDOMElement) Style() DOMStyle                                   { return &StubDOMStyle{} }
func (s *StubDOMElement) Parent() DOMElement                                { return &StubDOMElement{} }
func (s *StubDOMElement) Children() []DOMElement                            { return []DOMElement{} }
func (s *StubDOMElement) AppendChild(child DOMElement)                      {}
func (s *StubDOMElement) RemoveChild(child DOMElement)                      {}
func (s *StubDOMElement) Remove()                                            {}
func (s *StubDOMElement) Clone(deep bool) DOMElement                        { return &StubDOMElement{} }
func (s *StubDOMElement) IsVisible() bool                                    { return false }
func (s *StubDOMElement) Raw() interface{}                                   { return nil }

// StubDOMDocument implements DOMDocument for non-js builds
type StubDOMDocument struct{}

func (s *StubDOMDocument) GetElementByID(id string) DOMElement                        { return &StubDOMElement{} }
func (s *StubDOMDocument) QuerySelector(selector string) DOMElement                   { return &StubDOMElement{} }
func (s *StubDOMDocument) QuerySelectorAll(selector string) []DOMElement              { return []DOMElement{} }
func (s *StubDOMDocument) CreateElement(tagName string) DOMElement                     { return &StubDOMElement{} }
func (s *StubDOMDocument) CreateTextNode(text string) DOMElement                       { return &StubDOMElement{} }
func (s *StubDOMDocument) Body() DOMElement                                            { return &StubDOMElement{} }
func (s *StubDOMDocument) Head() DOMElement                                            { return &StubDOMElement{} }
func (s *StubDOMDocument) Title() string                                               { return "" }
func (s *StubDOMDocument) SetTitle(title string)                                      {}
func (s *StubDOMDocument) URL() string                                                 { return "" }
func (s *StubDOMDocument) ReadyState() string                                         { return "" }
func (s *StubDOMDocument) AddEventListener(eventType string, listener func(DOMEvent)) {}
func (s *StubDOMDocument) RemoveEventListener(eventType string, listener func(DOMEvent)) {}

// StubDOMBridge implements DOMBridge for non-js builds
type StubDOMBridge struct{}

func (s *StubDOMBridge) Document() DOMDocument                                  { return &StubDOMDocument{} }
func (s *StubDOMBridge) Window() JSValue                                        { return &StubJSValue{} }
func (s *StubDOMBridge) CreateEvent(eventType string) DOMEvent                  { return &StubDOMEvent{} }
func (s *StubDOMBridge) DispatchEvent(target DOMElement, event DOMEvent)        {}

// StubComponentBridge implements ComponentBridge for non-js builds
type StubComponentBridge struct{}

func (s *StubComponentBridge) InitializeComponent(componentName string, selector string, options map[string]any) error {
	return ErrNotSupported
}

func (s *StubComponentBridge) DestroyComponent(selector string, componentType string) error {
	return ErrNotSupported
}

func (s *StubComponentBridge) GetComponentInstance(selector, componentType string) (JSValue, error) {
	return &StubJSValue{}, ErrNotSupported
}

func (s *StubComponentBridge) InitializeAll(components []string) error {
	return ErrNotSupported
}

// StubManager implements Manager for non-js builds
type StubManager struct {
	jsBridge        JSBridge
	domBridge       DOMBridge
	componentBridge ComponentBridge
}

// NewStubManager creates a new stub manager for non-js builds
func NewStubManager() Manager {
	return &StubManager{
		jsBridge:        &StubJSBridge{},
		domBridge:       &StubDOMBridge{},
		componentBridge: &StubComponentBridge{},
	}
}

func (s *StubManager) JS() JSBridge                                    { return s.jsBridge }
func (s *StubManager) DOM() DOMBridge                                  { return s.domBridge }
func (s *StubManager) Component() ComponentBridge                      { return s.componentBridge }
func (s *StubManager) GoStringsToJSArray(strings []string) JSValue     { return &StubJSValue{} }
func (s *StubManager) GoMapToJSObject(m map[string]any) JSValue         { return &StubJSValue{} }
func (s *StubManager) JSArrayToGoStrings(jsArray JSValue) []string      { return []string{} }
func (s *StubManager) JSObjectToGoMap(jsObj JSValue) map[string]any      { return map[string]any{} }