// template_compiler.go
// Template compilation and optimization for direct DOM manipulation

//go:build js && wasm

package golid

import (
	"fmt"
	"regexp"
	"strings"
	"sync"
	"sync/atomic"
	"syscall/js"
)

// ------------------------------------
// 🏗️ Template Types
// ------------------------------------

// DOMTemplate represents a compiled template with optimized DOM operations
type DOMTemplate struct {
	id       string
	html     string
	factory  func() js.Value
	bindings []BindingDescriptor
	static   bool
	cached   js.Value
	mutex    sync.RWMutex
}

// BindingDescriptor describes a reactive binding within a template
type BindingDescriptor struct {
	path     []int              // DOM tree path to element
	type_    BindingType        // Type of binding
	accessor func() interface{} // Function to get the value
	property string             // Property/attribute name
}

// BindingType defines the type of DOM binding
type BindingType int

const (
	TextBinding BindingType = iota
	AttributeBinding
	PropertyBinding
	EventBinding
	ClassBinding
	StyleBinding
)

// HydrationContext manages template hydration state
type HydrationContext struct {
	root     js.Value
	bindings map[string]*DOMBinding
	owner    *Owner
	mutex    sync.RWMutex
}

// Global template registry
var (
	templateRegistry    = make(map[string]*DOMTemplate)
	templateIdCounter   uint64
	templateRegistryMux sync.RWMutex
)

// ------------------------------------
// 🔧 Template Compilation
// ------------------------------------

// CompileTemplate compiles an HTML template for optimized rendering
func CompileTemplate(html string) *DOMTemplate {
	templateId := fmt.Sprintf("tpl_%d", atomic.AddUint64(&templateIdCounter, 1))

	template := &DOMTemplate{
		id:       templateId,
		html:     html,
		bindings: make([]BindingDescriptor, 0),
		static:   isStaticTemplate(html),
	}

	// Parse and optimize the template
	template.parseBindings()
	template.createFactory()

	// Register template
	templateRegistryMux.Lock()
	templateRegistry[templateId] = template
	templateRegistryMux.Unlock()

	return template
}

// CompileTemplateWithBindings compiles a template with explicit bindings
func CompileTemplateWithBindings(html string, bindings []BindingDescriptor) *DOMTemplate {
	template := CompileTemplate(html)
	template.bindings = bindings
	template.static = len(bindings) == 0
	return template
}

// ------------------------------------
// 🔍 Template Parsing
// ------------------------------------

// parseBindings extracts reactive bindings from template HTML
func (t *DOMTemplate) parseBindings() {
	// Parse interpolation expressions {{...}}
	t.parseTextBindings()

	// Parse attribute bindings [attr]="..."
	t.parseAttributeBindings()

	// Parse event bindings (click)="..."
	t.parseEventBindings()

	// Parse class bindings [class.name]="..."
	t.parseClassBindings()

	// Parse style bindings [style.prop]="..."
	t.parseStyleBindings()
}

// parseTextBindings finds text interpolation bindings
func (t *DOMTemplate) parseTextBindings() {
	re := regexp.MustCompile(`\{\{([^}]+)\}\}`)
	matches := re.FindAllStringSubmatch(t.html, -1)

	for _, match := range matches {
		if len(match) > 1 {
			_ = strings.TrimSpace(match[1]) // expression for future use
			binding := BindingDescriptor{
				type_:    TextBinding,
				property: "textContent",
				path:     t.findElementPath(match[0]),
			}
			t.bindings = append(t.bindings, binding)
		}
	}
}

// parseAttributeBindings finds attribute bindings
func (t *DOMTemplate) parseAttributeBindings() {
	re := regexp.MustCompile(`\[([^\]]+)\]="([^"]*)"`)
	matches := re.FindAllStringSubmatch(t.html, -1)

	for _, match := range matches {
		if len(match) > 2 {
			attr := strings.TrimSpace(match[1])
			expression := strings.TrimSpace(match[2])

			binding := BindingDescriptor{
				type_:    AttributeBinding,
				property: attr,
				path:     t.findElementPath(match[0]),
			}
			t.bindings = append(t.bindings, binding)
			_ = expression // TODO: Parse expression
		}
	}
}

// parseEventBindings finds event bindings
func (t *DOMTemplate) parseEventBindings() {
	re := regexp.MustCompile(`\(([^)]+)\)="([^"]*)"`)
	matches := re.FindAllStringSubmatch(t.html, -1)

	for _, match := range matches {
		if len(match) > 2 {
			event := strings.TrimSpace(match[1])
			handler := strings.TrimSpace(match[2])

			binding := BindingDescriptor{
				type_:    EventBinding,
				property: event,
				path:     t.findElementPath(match[0]),
			}
			t.bindings = append(t.bindings, binding)
			_ = handler // TODO: Parse handler
		}
	}
}

// parseClassBindings finds CSS class bindings
func (t *DOMTemplate) parseClassBindings() {
	re := regexp.MustCompile(`\[class\.([^\]]+)\]="([^"]*)"`)
	matches := re.FindAllStringSubmatch(t.html, -1)

	for _, match := range matches {
		if len(match) > 2 {
			className := strings.TrimSpace(match[1])
			expression := strings.TrimSpace(match[2])

			binding := BindingDescriptor{
				type_:    ClassBinding,
				property: className,
				path:     t.findElementPath(match[0]),
			}
			t.bindings = append(t.bindings, binding)
			_ = expression // TODO: Parse expression
		}
	}
}

// parseStyleBindings finds CSS style bindings
func (t *DOMTemplate) parseStyleBindings() {
	re := regexp.MustCompile(`\[style\.([^\]]+)\]="([^"]*)"`)
	matches := re.FindAllStringSubmatch(t.html, -1)

	for _, match := range matches {
		if len(match) > 2 {
			styleProp := strings.TrimSpace(match[1])
			expression := strings.TrimSpace(match[2])

			binding := BindingDescriptor{
				type_:    StyleBinding,
				property: styleProp,
				path:     t.findElementPath(match[0]),
			}
			t.bindings = append(t.bindings, binding)
			_ = expression // TODO: Parse expression
		}
	}
}

// findElementPath finds the DOM path to an element containing the binding
func (t *DOMTemplate) findElementPath(bindingText string) []int {
	// Simplified path finding - in a real implementation, this would
	// parse the HTML structure and return the path to the element
	return []int{0} // Root element for now
}

// ------------------------------------
// 🏭 Template Factory
// ------------------------------------

// createFactory creates an optimized DOM factory function
func (t *DOMTemplate) createFactory() {
	t.factory = func() js.Value {
		t.mutex.RLock()
		defer t.mutex.RUnlock()

		// For static templates, reuse cached DOM
		if t.static && t.cached.Truthy() {
			return t.cached.Call("cloneNode", true)
		}

		// Create DOM from HTML
		container := doc.Call("createElement", "div")
		container.Set("innerHTML", t.html)

		var result js.Value
		if container.Get("children").Get("length").Int() == 1 {
			result = container.Get("firstElementChild")
		} else {
			// Multiple root elements - return document fragment
			fragment := doc.Call("createDocumentFragment")
			children := container.Get("children")
			length := children.Get("length").Int()

			for i := 0; i < length; i++ {
				fragment.Call("appendChild", children.Index(i))
			}
			result = fragment
		}

		// Cache static templates
		if t.static {
			t.cached = result.Call("cloneNode", true)
		}

		return result
	}
}

// ------------------------------------
// 🔄 Template Instantiation
// ------------------------------------

// Clone creates a new DOM instance from the template
func (t *DOMTemplate) Clone() js.Value {
	if t.factory == nil {
		t.createFactory()
	}
	return t.factory()
}

// Hydrate applies reactive bindings to a DOM element
func (t *DOMTemplate) Hydrate(root js.Value) *HydrationContext {
	if !root.Truthy() {
		return nil
	}

	ctx := &HydrationContext{
		root:     root,
		bindings: make(map[string]*DOMBinding),
		owner:    getCurrentOwner(),
	}

	// Apply all bindings
	for i, binding := range t.bindings {
		element := t.findElementByPath(root, binding.path)
		if !element.Truthy() {
			continue
		}

		bindingId := fmt.Sprintf("%s_%d", t.id, i)

		switch binding.type_ {
		case TextBinding:
			if binding.accessor != nil {
				domBinding := BindTextReactive(element, func() string {
					if val := binding.accessor(); val != nil {
						return fmt.Sprintf("%v", val)
					}
					return ""
				})
				if domBinding != nil {
					ctx.bindings[bindingId] = domBinding
				}
			}

		case AttributeBinding:
			if binding.accessor != nil {
				domBinding := BindAttributeReactive(element, binding.property, func() string {
					if val := binding.accessor(); val != nil {
						return fmt.Sprintf("%v", val)
					}
					return ""
				})
				if domBinding != nil {
					ctx.bindings[bindingId] = domBinding
				}
			}

		case ClassBinding:
			if binding.accessor != nil {
				domBinding := BindClassReactive(element, binding.property, func() bool {
					if val := binding.accessor(); val != nil {
						if b, ok := val.(bool); ok {
							return b
						}
					}
					return false
				})
				if domBinding != nil {
					ctx.bindings[bindingId] = domBinding
				}
			}

		case StyleBinding:
			if binding.accessor != nil {
				domBinding := BindStyleReactive(element, binding.property, func() string {
					if val := binding.accessor(); val != nil {
						return fmt.Sprintf("%v", val)
					}
					return ""
				})
				if domBinding != nil {
					ctx.bindings[bindingId] = domBinding
				}
			}
		}
	}

	return ctx
}

// findElementByPath finds an element by its DOM path
func (t *DOMTemplate) findElementByPath(root js.Value, path []int) js.Value {
	current := root

	for _, index := range path {
		children := current.Get("children")
		if children.Get("length").Int() <= index {
			return js.Undefined()
		}
		current = children.Index(index)
	}

	return current
}

// ------------------------------------
// 🧹 Template Cleanup
// ------------------------------------

// Cleanup removes all bindings and cleans up resources
func (ctx *HydrationContext) Cleanup() {
	ctx.mutex.Lock()
	defer ctx.mutex.Unlock()

	for _, binding := range ctx.bindings {
		if binding.cleanup != nil {
			binding.cleanup()
		}
	}

	ctx.bindings = make(map[string]*DOMBinding)
}

// GetBindingCount returns the number of active bindings
func (ctx *HydrationContext) GetBindingCount() int {
	ctx.mutex.RLock()
	defer ctx.mutex.RUnlock()
	return len(ctx.bindings)
}

// ------------------------------------
// 🔧 Template Utilities
// ------------------------------------

// isStaticTemplate determines if a template contains no reactive bindings
func isStaticTemplate(html string) bool {
	// Check for interpolation expressions
	if strings.Contains(html, "{{") {
		return false
	}

	// Check for attribute bindings
	if strings.Contains(html, "[") && strings.Contains(html, "]") {
		return false
	}

	// Check for event bindings
	if strings.Contains(html, "(") && strings.Contains(html, ")") {
		return false
	}

	return true
}

// GetTemplate retrieves a template by ID
func GetTemplate(id string) *DOMTemplate {
	templateRegistryMux.RLock()
	defer templateRegistryMux.RUnlock()
	return templateRegistry[id]
}

// ClearTemplateRegistry clears all registered templates (for testing)
func ClearTemplateRegistry() {
	templateRegistryMux.Lock()
	defer templateRegistryMux.Unlock()
	templateRegistry = make(map[string]*DOMTemplate)
}

// GetTemplateStats returns statistics about the template system
func GetTemplateStats() map[string]interface{} {
	templateRegistryMux.RLock()
	defer templateRegistryMux.RUnlock()

	staticCount := 0
	dynamicCount := 0

	for _, template := range templateRegistry {
		if template.static {
			staticCount++
		} else {
			dynamicCount++
		}
	}

	return map[string]interface{}{
		"totalTemplates":   len(templateRegistry),
		"staticTemplates":  staticCount,
		"dynamicTemplates": dynamicCount,
	}
}
