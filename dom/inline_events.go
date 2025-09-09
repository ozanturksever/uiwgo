//go:build js && wasm

package dom

import (
	"regexp"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"syscall/js"

	"github.com/ozanturksever/logutil"
	reactivity "github.com/ozanturksever/uiwgo/reactivity"
	domv2 "honnef.co/go/js/dom/v2"
	g "maragu.dev/gomponents"
)

// Inline event registry and utilities

var (
	inlineIDCounter      uint64
	inlineClickHandlers  = map[string]func(Element){}
	inlineClickOnceHandlers = map[string]func(Element){}
	inlineInputHandlers  = map[string]func(Element){}
	inlineChangeHandlers = map[string]func(Element){}
	inlineKeydownHandlers = map[string]func(Element){}
	inlineKeyExpectations = map[string]string{} // id -> expected key (optional)
	inlineSubmitHandlers = map[string]func(Element, map[string]string){}
	inlineFormResetHandlers = map[string]func(Element){}
	inlineFormChangeHandlers = map[string]func(Element, map[string]string){}
	inlineBlurHandlers   = map[string]func(Element){}
	inlineFocusHandlers  = map[string]func(Element){}
	inlineFocusWithinHandlers = map[string]func(Element, bool){} // element, isFocusEntering
	inlineValidateHandlers = map[string]func(Element, string) bool{}
	inlineBlurValidateHandlers = map[string]func(Element, string) bool{}
	inlineDebouncedInputHandlers = map[string]func(Element){}
	inlineSearchHandlers = map[string]func(Element, string){}
	inlineDebounceTimers = map[string]js.Value{} // stores setTimeout IDs
	inlineTabHandlers = map[string]func(Element){}
	inlineShiftTabHandlers = map[string]func(Element){}
	inlineArrowKeyHandlers = map[string]func(Element, string){} // direction: "up", "down", "left", "right"
	inlineDragStartHandlers = map[string]func(Element, js.Value){} // element, dataTransfer
	inlineDropHandlers = map[string]func(Element, js.Value){} // element, dataTransfer
	inlineDragOverHandlers = map[string]func(Element, js.Value){} // element, dataTransfer
	inlineOutsideClickHandlers = map[string]func(Element){} // element
	inlineEscapeCloseHandlers = map[string]func(Element){} // element
	inlineFileSelectHandlers = map[string]func(Element, []js.Value){} // element, files
	inlineFileDropHandlers = map[string]func(Element, []js.Value){} // element, files
	inlineInitHandlers = map[string]func(Element){}
	inlineDestroyHandlers = map[string]func(Element){}
	inlineVisibleHandlers = map[string]func(Element){}
	inlineResizeHandlers = map[string]func(Element){}
	inlineHandlersMu     sync.RWMutex
)

func nextInlineID(prefix string) string {
	id := atomic.AddUint64(&inlineIDCounter, 1)
	return prefix + "-" + strconv.FormatUint(id, 36)
}

// OnClickInline attaches an inline click handler to the element being created in gomponents.
// It returns a gomponents attribute like data-uiwgo-onclick="<id>" that will be auto-bound post-mount.
func OnClickInline(handler func(el Element)) g.Node {
	id := nextInlineID("clk")
	inlineHandlersMu.Lock()
	inlineClickHandlers[id] = handler
	inlineHandlersMu.Unlock()
	return g.Attr("data-uiwgo-onclick", id)
}

// OnInputInline attaches an inline input handler (input event)
func OnInputInline(handler func(el Element)) g.Node {
	id := nextInlineID("inp")
	inlineHandlersMu.Lock()
	inlineInputHandlers[id] = handler
	inlineHandlersMu.Unlock()
	return g.Attr("data-uiwgo-oninput", id)
}

// OnChangeInline attaches an inline change handler (change event)
func OnChangeInline(handler func(el Element)) g.Node {
	id := nextInlineID("chg")
	inlineHandlersMu.Lock()
	inlineChangeHandlers[id] = handler
	inlineHandlersMu.Unlock()
	return g.Attr("data-uiwgo-onchange", id)
}

// OnKeyDownInline attaches an inline keydown handler; optional expectedKey filters by key value
func OnKeyDownInline(handler func(el Element), expectedKey string) g.Node {
	id := nextInlineID("key")
	inlineHandlersMu.Lock()
	inlineKeydownHandlers[id] = handler
	if expectedKey != "" {
		inlineKeyExpectations[id] = expectedKey
	}
	inlineHandlersMu.Unlock()
	return g.Attr("data-uiwgo-onkeydown", id)
}

// OnEnterInline convenience for keydown Enter; uses dedicated marker so multiple key handlers can coexist
func OnEnterInline(handler func(el Element)) g.Node {
	id := nextInlineID("ent")
	inlineHandlersMu.Lock()
	inlineKeydownHandlers[id] = handler
	inlineKeyExpectations[id] = "Enter"
	inlineHandlersMu.Unlock()
	return g.Attr("data-uiwgo-onenter", id)
}

// OnEscapeInline convenience for keydown Escape; uses dedicated marker
func OnEscapeInline(handler func(el Element)) g.Node {
	id := nextInlineID("esc")
	inlineHandlersMu.Lock()
	inlineKeydownHandlers[id] = handler
	inlineKeyExpectations[id] = "Escape"
	inlineHandlersMu.Unlock()
	return g.Attr("data-uiwgo-onescape", id)
}

// OnSubmitInline attaches an inline form submit handler with automatic form data serialization
func OnSubmitInline(handler func(el Element, formData map[string]string)) g.Node {
	id := nextInlineID("sub")
	inlineHandlersMu.Lock()
	inlineSubmitHandlers[id] = handler
	inlineHandlersMu.Unlock()
	return g.Attr("data-uiwgo-onsubmit", id)
}

// OnFormResetInline attaches an inline form reset handler
func OnFormResetInline(handler func(el Element)) g.Node {
	id := nextInlineID("rst")
	inlineHandlersMu.Lock()
	inlineFormResetHandlers[id] = handler
	inlineHandlersMu.Unlock()
	return g.Attr("data-uiwgo-onreset", id)
}

// OnFormChangeInline attaches an inline form change handler with automatic form data serialization
func OnFormChangeInline(handler func(el Element, formData map[string]string)) g.Node {
	id := nextInlineID("fch")
	inlineHandlersMu.Lock()
	inlineFormChangeHandlers[id] = handler
	inlineHandlersMu.Unlock()
	return g.Attr("data-uiwgo-onformchange", id)
}

// ValidationPattern represents common validation patterns
type ValidationPattern string

const (
	ValidationEmail    ValidationPattern = "email"
	ValidationURL      ValidationPattern = "url"
	ValidationPhone    ValidationPattern = "phone"
	ValidationRequired ValidationPattern = "required"
	ValidationMinLength ValidationPattern = "minlength"
	ValidationMaxLength ValidationPattern = "maxlength"
	ValidationRegex    ValidationPattern = "pattern"
	ValidationNumber   ValidationPattern = "number"
)

// OnValidateInline binds a validation handler with built-in patterns
func OnValidateInline(pattern ValidationPattern, options ...string) g.Node {
	id := nextInlineID("val")
	validator := createValidator(pattern, options...)
	inlineHandlersMu.Lock()
	inlineValidateHandlers[id] = validator
	inlineHandlersMu.Unlock()
	return g.Attr("data-uiwgo-validate", id)
}

// OnBlurValidateInline binds a validation handler that triggers on blur
func OnBlurValidateInline(pattern ValidationPattern, options ...string) g.Node {
	id := nextInlineID("blrval")
	validator := createValidator(pattern, options...)
	inlineHandlersMu.Lock()
	inlineBlurValidateHandlers[id] = validator
	inlineHandlersMu.Unlock()
	return g.Attr("data-uiwgo-blur-validate", id)
}

// createValidator creates a validation function based on pattern and options
func createValidator(pattern ValidationPattern, options ...string) func(Element, string) bool {
	return func(el Element, value string) bool {
		switch pattern {
		case ValidationRequired:
			return strings.TrimSpace(value) != ""
		case ValidationEmail:
			emailRegex := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
			matched, _ := regexp.MatchString(emailRegex, value)
			return matched
		case ValidationURL:
			urlRegex := `^https?://[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}(/.*)?$`
			matched, _ := regexp.MatchString(urlRegex, value)
			return matched
		case ValidationPhone:
			phoneRegex := `^[+]?[1-9]?[0-9]{7,15}$`
			matched, _ := regexp.MatchString(phoneRegex, value)
			return matched
		case ValidationMinLength:
			if len(options) > 0 {
				if minLen, err := strconv.Atoi(options[0]); err == nil {
					return len(value) >= minLen
				}
			}
			return true
		case ValidationMaxLength:
			if len(options) > 0 {
				if maxLen, err := strconv.Atoi(options[0]); err == nil {
					return len(value) <= maxLen
				}
			}
			return true
		case ValidationRegex:
			if len(options) > 0 {
				matched, _ := regexp.MatchString(options[0], value)
				return matched
			}
			return true
		case ValidationNumber:
			_, err := strconv.ParseFloat(value, 64)
			return err == nil
		default:
			return true
		}
	}
}

// OnBlurInline attaches an inline blur handler
func OnBlurInline(handler func(el Element)) g.Node {
	id := nextInlineID("blr")
	inlineHandlersMu.Lock()
	inlineBlurHandlers[id] = handler
	inlineHandlersMu.Unlock()
	return g.Attr("data-uiwgo-onblur", id)
}

// OnFocusInline attaches an inline focus handler
func OnFocusInline(handler func(el Element)) g.Node {
	id := nextInlineID("focus")
	inlineHandlersMu.Lock()
	inlineFocusHandlers[id] = handler
	inlineHandlersMu.Unlock()
	return g.Attr("data-uiwgo-onfocus", id)
}

// OnFocusWithinInline attaches a handler that fires when focus enters or leaves an element or its descendants
// The handler receives the element and a boolean indicating if focus is entering (true) or leaving (false)
func OnFocusWithinInline(handler func(el Element, isFocusEntering bool)) g.Node {
	id := nextInlineID("focuswithin")
	inlineHandlersMu.Lock()
	inlineFocusWithinHandlers[id] = handler
	inlineHandlersMu.Unlock()
	return g.Attr("data-uiwgo-onfocuswithin", id)
}

// OnInputDebouncedInline creates a debounced input handler that waits for a pause in typing
func OnInputDebouncedInline(handler func(el Element), delayMs int) g.Node {
	id := nextInlineID("debounced")
	inlineHandlersMu.Lock()
	inlineDebouncedInputHandlers[id] = handler
	inlineHandlersMu.Unlock()
	return g.Group([]g.Node{
		g.Attr("data-inline-debounced", id),
		g.Attr("data-debounce-delay", strconv.Itoa(delayMs)),
	})
}

// OnSearchInline creates a search input handler with debouncing and query extraction
func OnSearchInline(handler func(el Element, query string), delayMs int) g.Node {
	id := nextInlineID("search")
	inlineHandlersMu.Lock()
	inlineSearchHandlers[id] = handler
	inlineHandlersMu.Unlock()
	return g.Group([]g.Node{
		g.Attr("data-inline-search", id),
		g.Attr("data-search-delay", strconv.Itoa(delayMs)),
	})
}

// OnTabInline handles Tab key navigation for accessibility
func OnTabInline(handler func(el Element)) g.Node {
	id := nextInlineID("tab")
	inlineHandlersMu.Lock()
	inlineTabHandlers[id] = handler
	inlineHandlersMu.Unlock()

	return g.Attr("data-inline-tab", id)
}

// OnShiftTabInline handles Shift+Tab key navigation for accessibility
func OnShiftTabInline(handler func(el Element)) g.Node {
	id := nextInlineID("shifttab")
	inlineHandlersMu.Lock()
	inlineShiftTabHandlers[id] = handler
	inlineHandlersMu.Unlock()

	return g.Attr("data-inline-shifttab", id)
}

// OnArrowKeysInline handles arrow key navigation for accessibility
// The handler receives the element and direction ("up", "down", "left", "right")
func OnArrowKeysInline(handler func(el Element, direction string)) g.Node {
	id := nextInlineID("arrow")
	inlineHandlersMu.Lock()
	inlineArrowKeyHandlers[id] = handler
	inlineHandlersMu.Unlock()

	return g.Attr("data-inline-arrow", id)
}

// OnDragStartInline handles dragstart events for drag and drop functionality
// The handler receives the element and dataTransfer object
func OnDragStartInline(handler func(el Element, dataTransfer js.Value)) g.Node {
	id := nextInlineID("dragstart")
	inlineHandlersMu.Lock()
	inlineDragStartHandlers[id] = handler
	inlineHandlersMu.Unlock()

	return g.Attr("data-inline-dragstart", id)
}

// OnDropInline handles drop events for drag and drop functionality
// The handler receives the element and dataTransfer object
func OnDropInline(handler func(el Element, dataTransfer js.Value)) g.Node {
	id := nextInlineID("drop")
	inlineHandlersMu.Lock()
	inlineDropHandlers[id] = handler
	inlineHandlersMu.Unlock()

	return g.Attr("data-inline-drop", id)
}

// OnDragOverInline handles dragover events for drag and drop functionality
// The handler receives the element and dataTransfer object
// Note: You typically want to call event.preventDefault() in the handler to allow dropping
func OnDragOverInline(handler func(el Element, dataTransfer js.Value)) g.Node {
	id := nextInlineID("dragover")
	inlineHandlersMu.Lock()
	inlineDragOverHandlers[id] = handler
	inlineHandlersMu.Unlock()

	return g.Attr("data-inline-dragover", id)
}

// serializeFormData extracts form data from a form element and returns it as a map

// OnOutsideClickInline registers a handler that triggers when a click occurs outside the element
func OnOutsideClickInline(handler func(el Element)) g.Node {
	id := nextInlineID("outside-click")
	inlineHandlersMu.Lock()
	inlineOutsideClickHandlers[id] = handler
	inlineHandlersMu.Unlock()
	return g.Attr("data-uiwgo-onoutsideclick", id)
}

// OnEscapeCloseInline registers a handler that triggers when the Escape key is pressed
func OnEscapeCloseInline(handler func(el Element)) g.Node {
	id := nextInlineID("escapeclose")
	inlineHandlersMu.Lock()
	inlineEscapeCloseHandlers[id] = handler
	inlineHandlersMu.Unlock()
	return g.Attr("data-uiwgo-onescapeclose", id)
}

// OnFileSelectInline creates a file input handler that triggers when files are selected
func OnFileSelectInline(handler func(el Element, files []js.Value)) g.Node {
	id := nextInlineID("fileselect")
	inlineHandlersMu.Lock()
	inlineFileSelectHandlers[id] = handler
	inlineHandlersMu.Unlock()
	return g.Attr("data-uiwgo-onfileselect", id)
}

// OnFileDropInline creates a drop zone handler for file uploads
func OnFileDropInline(handler func(el Element, files []js.Value)) g.Node {
	id := nextInlineID("filedrop")
	inlineHandlersMu.Lock()
	inlineFileDropHandlers[id] = handler
	inlineHandlersMu.Unlock()
	return g.Attr("data-uiwgo-onfiledrop", id)
}

func serializeFormData(formEl Element) map[string]string {
	formData := make(map[string]string)
	if formEl == nil {
		return formData
	}

	// Get the underlying JS form element
	formJS := formEl.Underlying()
	if formJS.IsUndefined() || formJS.IsNull() {
		return formData
	}

	// Use FormData API to extract form values
	formDataJS := js.Global().Get("FormData").New(formJS)
	if formDataJS.IsUndefined() || formDataJS.IsNull() {
		return formData
	}

	// Iterate through FormData entries
	iterator := formDataJS.Call("entries")
	for {
		next := iterator.Call("next")
		if next.Get("done").Bool() {
			break
		}
		value := next.Get("value")
		if value.Length() >= 2 {
			key := value.Index(0).String()
			val := value.Index(1).String()
			formData[key] = val
		}
	}

	return formData
}

// AttachInlineDelegates scans under the provided root and installs delegated listeners
// for supported inline events. It registers cleanup with the current reactivity scope.
func AttachInlineDelegates(root js.Value) {
	// Helper to install a delegated listener with marker and registry handlers
	install := func(eventType, marker string, lookup func(id string) (func(Element), bool), collectIds func() []string) (installed bool, fn js.Func, ids []string) {
		// Check if any markers exist under root
		nodes := root.Call("querySelectorAll", marker)
		if !nodes.Truthy() || nodes.Get("length").Int() == 0 {
			return false, js.Func{}, nil
		}

		// Collect ids for cleanup scoping
		ids = collectIds()
		if len(ids) == 0 {
			// still install, dynamic elements might appear later
		}

		fn = js.FuncOf(func(this js.Value, args []js.Value) any {
			if len(args) == 0 {
				return nil
			}
			rawEvent := args[0]
			target := rawEvent.Get("target")
			if target.IsUndefined() || target.IsNull() {
				return nil
			}
			matched := target.Call("closest", marker)
			if matched.IsUndefined() || matched.IsNull() {
				return nil
			}
   attrName := marker[1 : len(marker)-1]
   id := matched.Call("getAttribute", attrName).String()
   if id == "" {
   	return nil
   }

   inlineHandlersMu.RLock()
   h, ok := lookup(id)
   inlineHandlersMu.RUnlock()
			if !ok {
				return nil
			}

			el := domv2.WrapElement(matched)
			if el == nil {
				return nil
			}
			defer func() {
				if r := recover(); r != nil {
					logutil.Logf("panic in inline handler for %s: %v", eventType, r)
				}
			}()
			h(el)
			return nil
		})

		root.Call("addEventListener", eventType, fn)
		return true, fn, ids
	}

	collect := func(attr string) []string {
		attrSel := "[" + attr + "]"
		nodes := root.Call("querySelectorAll", attrSel)
		ln := nodes.Get("length").Int()
		res := make([]string, 0, ln)
		for i := 0; i < ln; i++ {
			el := nodes.Call("item", i)
			id := el.Call("getAttribute", attr).String()
			if id != "" {
				res = append(res, id)
			}
		}
		return res
	}

	// Install for click
	clickInstalled, clickFn, clickIDs := install("click", "[data-uiwgo-onclick]", func(id string) (func(Element), bool) {
		return inlineClickHandlers[id], inlineClickHandlers[id] != nil
	}, func() []string { return collect("data-uiwgo-onclick") })

	// Install for input
	inputInstalled, inputFn, inputIDs := install("input", "[data-uiwgo-oninput]", func(id string) (func(Element), bool) {
		return inlineInputHandlers[id], inlineInputHandlers[id] != nil
	}, func() []string { return collect("data-uiwgo-oninput") })

	// Install for change
	changeInstalled, changeFn, changeIDs := install("change", "[data-uiwgo-onchange]", func(id string) (func(Element), bool) {
		return inlineChangeHandlers[id], inlineChangeHandlers[id] != nil
	}, func() []string { return collect("data-uiwgo-onchange") })

	// Install for keydown (with optional key expectation)
	keydownInstalled := false
	var keydownFn js.Func
	var keydownIDs []string
	{
		marker := "[data-uiwgo-onkeydown]"
		nodes := root.Call("querySelectorAll", marker)
		if nodes.Truthy() && nodes.Get("length").Int() > 0 {
			// Collect ids
			keydownIDs = collect("data-uiwgo-onkeydown")
			keydownFn = js.FuncOf(func(this js.Value, args []js.Value) any {
				if len(args) == 0 {
					return nil
				}
				rawEvent := args[0]
				key := rawEvent.Get("key").String()
				target := rawEvent.Get("target")
				if target.IsUndefined() || target.IsNull() { return nil }
				matched := target.Call("closest", marker)
				if matched.IsUndefined() || matched.IsNull() { return nil }
				id := matched.Call("getAttribute", "data-uiwgo-onkeydown").String()
				if id == "" { return nil }

				inlineHandlersMu.RLock()
				expected := inlineKeyExpectations[id]
				h := inlineKeydownHandlers[id]
				inlineHandlersMu.RUnlock()
				if h == nil { return nil }
				if expected != "" && key != expected { return nil }

				el := domv2.WrapElement(matched)
				if el == nil { return nil }
				defer func(){ if r := recover(); r != nil { logutil.Logf("panic in inline keydown: %v", r) } }()
				h(el)
				return nil
			})
			root.Call("addEventListener", "keydown", keydownFn)
			keydownInstalled = true
		}
	}

	// Additional keydown delegates for specific keys (Enter/Escape) with dedicated markers
 enterInstalled := false
	var enterFn js.Func
	var enterFnUp js.Func
	var enterIDs []string
	{
		marker := "[data-uiwgo-onenter]"
		nodes := root.Call("querySelectorAll", marker)
		if nodes.Truthy() && nodes.Get("length").Int() > 0 {
			enterIDs = collect("data-uiwgo-onenter")
			enterFn = js.FuncOf(func(this js.Value, args []js.Value) any {
				if len(args) == 0 { return nil }
				rawEvent := args[0]
				key := rawEvent.Get("key").String()
				if key != "Enter" { return nil }
				target := rawEvent.Get("target")
				if target.IsUndefined() || target.IsNull() { return nil }
				matched := target.Call("closest", marker)
				if matched.IsUndefined() || matched.IsNull() { return nil }
				id := matched.Call("getAttribute", "data-uiwgo-onenter").String()
				if id == "" { return nil }
				inlineHandlersMu.RLock()
				h := inlineKeydownHandlers[id]
				inlineHandlersMu.RUnlock()
				if h == nil { return nil }
				el := domv2.WrapElement(matched)
				if el == nil { return nil }
				defer func(){ if r := recover(); r != nil { logutil.Logf("panic in inline onenter: %v", r) } }()
				h(el)
				return nil
			})
			root.Call("addEventListener", "keydown", enterFn)
			// Also listen on keyup for robustness across drivers
			enterFnUp = js.FuncOf(func(this js.Value, args []js.Value) any {
				if len(args) == 0 { return nil }
				rawEvent := args[0]
				key := rawEvent.Get("key").String()
				if key != "Enter" { return nil }
				target := rawEvent.Get("target")
				if target.IsUndefined() || target.IsNull() { return nil }
				matched := target.Call("closest", marker)
				if matched.IsUndefined() || matched.IsNull() { return nil }
				id := matched.Call("getAttribute", "data-uiwgo-onenter").String()
				if id == "" { return nil }
				inlineHandlersMu.RLock()
				h := inlineKeydownHandlers[id]
				inlineHandlersMu.RUnlock()
				if h == nil { return nil }
				el := domv2.WrapElement(matched)
				if el == nil { return nil }
				defer func(){ if r := recover(); r != nil { logutil.Logf("panic in inline onenter (keyup): %v", r) } }()
				h(el)
				return nil
			})
			root.Call("addEventListener", "keyup", enterFnUp)
			enterInstalled = true
		}
	}

 escapeInstalled := false
	var escapeFn js.Func
	var escapeFnUp js.Func
	var escapeIDs []string
	{
		marker := "[data-uiwgo-onescape]"
		nodes := root.Call("querySelectorAll", marker)
		if nodes.Truthy() && nodes.Get("length").Int() > 0 {
			escapeIDs = collect("data-uiwgo-onescape")
			escapeFn = js.FuncOf(func(this js.Value, args []js.Value) any {
				if len(args) == 0 { return nil }
				rawEvent := args[0]
				key := rawEvent.Get("key").String()
				if key != "Escape" { return nil }
				target := rawEvent.Get("target")
				if target.IsUndefined() || target.IsNull() { return nil }
				matched := target.Call("closest", marker)
				if matched.IsUndefined() || matched.IsNull() { return nil }
				id := matched.Call("getAttribute", "data-uiwgo-onescape").String()
				if id == "" { return nil }
				inlineHandlersMu.RLock()
				h := inlineKeydownHandlers[id]
				inlineHandlersMu.RUnlock()
				if h == nil { return nil }
				el := domv2.WrapElement(matched)
				if el == nil { return nil }
				defer func(){ if r := recover(); r != nil { logutil.Logf("panic in inline onescape: %v", r) } }()
				h(el)
				return nil
			})
			root.Call("addEventListener", "keydown", escapeFn)
			// Also listen on keyup for robustness
			escapeFnUp = js.FuncOf(func(this js.Value, args []js.Value) any {
				if len(args) == 0 { return nil }
				rawEvent := args[0]
				key := rawEvent.Get("key").String()
				if key != "Escape" { return nil }
				target := rawEvent.Get("target")
				if target.IsUndefined() || target.IsNull() { return nil }
				matched := target.Call("closest", marker)
				if matched.IsUndefined() || matched.IsNull() { return nil }
				id := matched.Call("getAttribute", "data-uiwgo-onescape").String()
				if id == "" { return nil }
				inlineHandlersMu.RLock()
				h := inlineKeydownHandlers[id]
				inlineHandlersMu.RUnlock()
				if h == nil { return nil }
				el := domv2.WrapElement(matched)
				if el == nil { return nil }
				defer func(){ if r := recover(); r != nil { logutil.Logf("panic in inline onescape (keyup): %v", r) } }()
				h(el)
				return nil
			})
			root.Call("addEventListener", "keyup", escapeFnUp)
			escapeInstalled = true
		}
	}

	// Install for submit (with form data serialization)
	submitInstalled := false
	var submitFn js.Func
	var submitIDs []string
	{
		marker := "[data-uiwgo-onsubmit]"
		nodes := root.Call("querySelectorAll", marker)
		if nodes.Truthy() && nodes.Get("length").Int() > 0 {
			submitIDs = collect("data-uiwgo-onsubmit")
			submitFn = js.FuncOf(func(this js.Value, args []js.Value) any {
				if len(args) == 0 { return nil }
				rawEvent := args[0]
				rawEvent.Call("preventDefault") // Prevent default form submission
				target := rawEvent.Get("target")
				if target.IsUndefined() || target.IsNull() { return nil }
				matched := target.Call("closest", marker)
				if matched.IsUndefined() || matched.IsNull() { return nil }
				id := matched.Call("getAttribute", "data-uiwgo-onsubmit").String()
				if id == "" { return nil }
				inlineHandlersMu.RLock()
				h := inlineSubmitHandlers[id]
				inlineHandlersMu.RUnlock()
				if h == nil { return nil }
				el := domv2.WrapElement(matched)
				if el == nil { return nil }
				defer func(){ if r := recover(); r != nil { logutil.Logf("panic in inline submit: %v", r) } }()
				formData := serializeFormData(el)
				h(el, formData)
				return nil
			})
			root.Call("addEventListener", "submit", submitFn)
			submitInstalled = true
		}
	}

	// Install for reset
	resetInstalled := false
	var resetFn js.Func
	var resetIDs []string
	{
		marker := "[data-uiwgo-onreset]"
		nodes := root.Call("querySelectorAll", marker)
		if nodes.Truthy() && nodes.Get("length").Int() > 0 {
			resetIDs = collect("data-uiwgo-onreset")
			resetFn = js.FuncOf(func(this js.Value, args []js.Value) any {
				if len(args) == 0 { return nil }
				rawEvent := args[0]
				target := rawEvent.Get("target")
				if target.IsUndefined() || target.IsNull() { return nil }
				matched := target.Call("closest", marker)
				if matched.IsUndefined() || matched.IsNull() { return nil }
				id := matched.Call("getAttribute", "data-uiwgo-onreset").String()
				if id == "" { return nil }
				inlineHandlersMu.RLock()
				h := inlineFormResetHandlers[id]
				inlineHandlersMu.RUnlock()
				if h == nil { return nil }
				el := domv2.WrapElement(matched)
				if el == nil { return nil }
				defer func(){ if r := recover(); r != nil { logutil.Logf("panic in inline reset: %v", r) } }()
				h(el)
				return nil
			})
			root.Call("addEventListener", "reset", resetFn)
			resetInstalled = true
		}
	}

	// Install for blur
	blurInstalled, blurFn, blurIDs := install("blur", "[data-uiwgo-onblur]", func(id string) (func(Element), bool) {
		return inlineBlurHandlers[id], inlineBlurHandlers[id] != nil
	}, func() []string { return collect("data-uiwgo-onblur") })

	// Install for focus
	focusInstalled, focusFn, focusIDs := install("focus", "[data-uiwgo-onfocus]", func(id string) (func(Element), bool) {
		return inlineFocusHandlers[id], inlineFocusHandlers[id] != nil
	}, func() []string { return collect("data-uiwgo-onfocus") })

	// Install for focuswithin (using focusin/focusout events)
	focusWithinInstalled := false
	var focusInFn, focusOutFn js.Func
	var focusWithinIDs []string
	{
		marker := "[data-uiwgo-onfocuswithin]"
		nodes := root.Call("querySelectorAll", marker)
		if nodes.Truthy() && nodes.Get("length").Int() > 0 {
			focusWithinIDs = collect("data-uiwgo-onfocuswithin")
			// focusin event (focus entering)
			focusInFn = js.FuncOf(func(this js.Value, args []js.Value) any {
				if len(args) == 0 { return nil }
				rawEvent := args[0]
				target := rawEvent.Get("target")
				matched := target.Call("closest", marker)
				if !matched.Truthy() { return nil }
				id := matched.Call("getAttribute", "data-uiwgo-onfocuswithin").String()
				if id == "" { return nil }
				inlineHandlersMu.RLock()
				h := inlineFocusWithinHandlers[id]
				inlineHandlersMu.RUnlock()
				if h == nil { return nil }
				el := domv2.WrapElement(matched)
				if el == nil { return nil }
				defer func(){ if r := recover(); r != nil { logutil.Logf("panic in inline focuswithin (in): %v", r) } }()
				h(el, true) // focus entering
				return nil
			})
			// focusout event (focus leaving)
			focusOutFn = js.FuncOf(func(this js.Value, args []js.Value) any {
				if len(args) == 0 { return nil }
				rawEvent := args[0]
				target := rawEvent.Get("target")
				matched := target.Call("closest", marker)
				if !matched.Truthy() { return nil }
				id := matched.Call("getAttribute", "data-uiwgo-onfocuswithin").String()
				if id == "" { return nil }
				inlineHandlersMu.RLock()
				h := inlineFocusWithinHandlers[id]
				inlineHandlersMu.RUnlock()
				if h == nil { return nil }
				el := domv2.WrapElement(matched)
				if el == nil { return nil }
				defer func(){ if r := recover(); r != nil { logutil.Logf("panic in inline focuswithin (out): %v", r) } }()
				h(el, false) // focus leaving
				return nil
			})
			root.Call("addEventListener", "focusin", focusInFn)
			root.Call("addEventListener", "focusout", focusOutFn)
			focusWithinInstalled = true
		}
	}

	// Install for form change (delegated change on form elements with form data serialization)
	formChangeInstalled := false
	var formChangeFn js.Func
	var formChangeIDs []string
	{
		marker := "[data-uiwgo-onformchange]"
		nodes := root.Call("querySelectorAll", marker)
		if nodes.Truthy() && nodes.Get("length").Int() > 0 {
			formChangeIDs = collect("data-uiwgo-onformchange")
			formChangeFn = js.FuncOf(func(this js.Value, args []js.Value) any {
				if len(args) == 0 { return nil }
				rawEvent := args[0]
				target := rawEvent.Get("target")
				if target.IsUndefined() || target.IsNull() { return nil }
				// Find the form that contains this changed element
				form := target.Call("closest", "form")
				if form.IsUndefined() || form.IsNull() { return nil }
				// Check if this form has the form change marker
				matched := form.Call("closest", marker)
				if matched.IsUndefined() || matched.IsNull() { return nil }
				id := matched.Call("getAttribute", "data-uiwgo-onformchange").String()
				if id == "" { return nil }
				inlineHandlersMu.RLock()
				h := inlineFormChangeHandlers[id]
				inlineHandlersMu.RUnlock()
				if h == nil { return nil }
				el := domv2.WrapElement(matched)
				if el == nil { return nil }
				defer func(){ if r := recover(); r != nil { logutil.Logf("panic in inline form change: %v", r) } }()
				formData := serializeFormData(el)
				h(el, formData)
				return nil
			})
			root.Call("addEventListener", "change", formChangeFn)
			formChangeInstalled = true
		}
	}

	// Install for input validation
	validateInstalled := false
	var validateFn js.Func
	var validateIDs []string
	{
		marker := "[data-uiwgo-validate]"
		nodes := root.Call("querySelectorAll", marker)
		if nodes.Truthy() && nodes.Get("length").Int() > 0 {
			validateIDs = collect("data-uiwgo-validate")
			validateFn = js.FuncOf(func(this js.Value, args []js.Value) any {
				if len(args) == 0 { return nil }
				rawEvent := args[0]
				target := rawEvent.Get("target")
				if target.IsUndefined() || target.IsNull() { return nil }
				matched := target.Call("closest", marker)
				if matched.IsUndefined() || matched.IsNull() { return nil }
				id := matched.Call("getAttribute", "data-uiwgo-validate").String()
				if id == "" { return nil }
				inlineHandlersMu.RLock()
				h := inlineValidateHandlers[id]
				inlineHandlersMu.RUnlock()
				if h == nil { return nil }
				el := domv2.WrapElement(matched)
				if el == nil { return nil }
				value := target.Get("value").String()
				defer func(){ if r := recover(); r != nil { logutil.Logf("panic in inline validate: %v", r) } }()
				isValid := h(el, value)
				// Set validation state on element
				if isValid {
					matched.Call("removeAttribute", "data-invalid")
					matched.Call("setAttribute", "data-valid", "true")
				} else {
					matched.Call("removeAttribute", "data-valid")
					matched.Call("setAttribute", "data-invalid", "true")
				}
				return nil
			})
			root.Call("addEventListener", "input", validateFn)
			validateInstalled = true
		}
	}

	// Install for blur validation
	blurValidateInstalled := false
	var blurValidateFn js.Func
	var blurValidateIDs []string
	{
		marker := "[data-uiwgo-blur-validate]"
		nodes := root.Call("querySelectorAll", marker)
		if nodes.Truthy() && nodes.Get("length").Int() > 0 {
			blurValidateIDs = collect("data-uiwgo-blur-validate")
			blurValidateFn = js.FuncOf(func(this js.Value, args []js.Value) any {
				if len(args) == 0 { return nil }
				rawEvent := args[0]
				target := rawEvent.Get("target")
				if target.IsUndefined() || target.IsNull() { return nil }
				matched := target.Call("closest", marker)
				if matched.IsUndefined() || matched.IsNull() { return nil }
				id := matched.Call("getAttribute", "data-uiwgo-blur-validate").String()
				if id == "" { return nil }
				inlineHandlersMu.RLock()
				h := inlineBlurValidateHandlers[id]
				inlineHandlersMu.RUnlock()
				if h == nil { return nil }
				el := domv2.WrapElement(matched)
				if el == nil { return nil }
				value := target.Get("value").String()
				defer func(){ if r := recover(); r != nil { logutil.Logf("panic in inline blur validate: %v", r) } }()
				isValid := h(el, value)
				// Set validation state on element
				if isValid {
					matched.Call("removeAttribute", "data-invalid")
					matched.Call("setAttribute", "data-valid", "true")
				} else {
					matched.Call("removeAttribute", "data-valid")
					matched.Call("setAttribute", "data-invalid", "true")
				}
				return nil
			})
			root.Call("addEventListener", "blur", blurValidateFn)
			blurValidateInstalled = true
		}
	}

	// Install for debounced input
	debouncedInstalled := false
	var debouncedFn js.Func
	var debouncedIDs []string
	{
		marker := "[data-inline-debounced]"
		nodes := root.Call("querySelectorAll", marker)
		if nodes.Truthy() && nodes.Get("length").Int() > 0 {
			debouncedIDs = collect("data-inline-debounced")
			debouncedFn = js.FuncOf(func(this js.Value, args []js.Value) any {
				if len(args) == 0 { return nil }
				rawEvent := args[0]
				target := rawEvent.Get("target")
				if target.IsUndefined() || target.IsNull() { return nil }
				matched := target.Call("closest", marker)
				if matched.IsUndefined() || matched.IsNull() { return nil }
				id := matched.Call("getAttribute", "data-inline-debounced").String()
				if id == "" { return nil }
				delayStr := matched.Call("getAttribute", "data-debounce-delay").String()
				delay := 300 // default
				if delayStr != "" {
					if parsed, err := strconv.Atoi(delayStr); err == nil {
						delay = parsed
					}
				}
				inlineHandlersMu.RLock()
				h := inlineDebouncedInputHandlers[id]
				inlineHandlersMu.RUnlock()
				if h == nil { return nil }
				el := domv2.WrapElement(matched)
				if el == nil { return nil }
				// Clear existing timer
				inlineHandlersMu.Lock()
				if timer, exists := inlineDebounceTimers[id]; exists && !timer.IsUndefined() {
					js.Global().Call("clearTimeout", timer)
				}
				// Set new timer
				timer := js.Global().Call("setTimeout", js.FuncOf(func(this js.Value, args []js.Value) any {
					defer func(){ if r := recover(); r != nil { logutil.Logf("panic in inline debounced: %v", r) } }()
					h(el)
					inlineHandlersMu.Lock()
					delete(inlineDebounceTimers, id)
					inlineHandlersMu.Unlock()
					return nil
				}), delay)
				inlineDebounceTimers[id] = timer
				inlineHandlersMu.Unlock()
				return nil
			})
			root.Call("addEventListener", "input", debouncedFn)
			debouncedInstalled = true
		}
	}

	// Install for search input
	searchInstalled := false
	var searchFn js.Func
	var searchIDs []string
	{
		marker := "[data-inline-search]"
		nodes := root.Call("querySelectorAll", marker)
		if nodes.Truthy() && nodes.Get("length").Int() > 0 {
			searchIDs = collect("data-inline-search")
			searchFn = js.FuncOf(func(this js.Value, args []js.Value) any {
				if len(args) == 0 { return nil }
				rawEvent := args[0]
				target := rawEvent.Get("target")
				if target.IsUndefined() || target.IsNull() { return nil }
				matched := target.Call("closest", marker)
				if matched.IsUndefined() || matched.IsNull() { return nil }
				id := matched.Call("getAttribute", "data-inline-search").String()
				if id == "" { return nil }
				delayStr := matched.Call("getAttribute", "data-search-delay").String()
				delay := 300 // default
				if delayStr != "" {
					if parsed, err := strconv.Atoi(delayStr); err == nil {
						delay = parsed
					}
				}
				inlineHandlersMu.RLock()
				h := inlineSearchHandlers[id]
				inlineHandlersMu.RUnlock()
				if h == nil { return nil }
				el := domv2.WrapElement(matched)
				if el == nil { return nil }
				query := target.Get("value").String()
				// Clear existing timer
				inlineHandlersMu.Lock()
				if timer, exists := inlineDebounceTimers[id]; exists && !timer.IsUndefined() {
					js.Global().Call("clearTimeout", timer)
				}
				// Set new timer
				timer := js.Global().Call("setTimeout", js.FuncOf(func(this js.Value, args []js.Value) any {
					defer func(){ if r := recover(); r != nil { logutil.Logf("panic in inline search: %v", r) } }()
					h(el, query)
					inlineHandlersMu.Lock()
					delete(inlineDebounceTimers, id)
					inlineHandlersMu.Unlock()
					return nil
				}), delay)
				inlineDebounceTimers[id] = timer
				inlineHandlersMu.Unlock()
				return nil
		})
		root.Call("addEventListener", "input", searchFn)
		searchInstalled = true
	}
}

	// Install for Tab navigation
	tabInstalled := false
	var tabFn js.Func
	var tabIDs []string
	{
		marker := "[data-inline-tab]"
		nodes := root.Call("querySelectorAll", marker)
		if nodes.Truthy() && nodes.Get("length").Int() > 0 {
			tabIDs = collect("data-inline-tab")
			tabFn = js.FuncOf(func(this js.Value, args []js.Value) any {
				if len(args) == 0 { return nil }
				rawEvent := args[0]
				key := rawEvent.Get("key").String()
				if key != "Tab" { return nil }
				shiftKey := rawEvent.Get("shiftKey").Bool()
				if shiftKey { return nil } // Handle only Tab, not Shift+Tab
				target := rawEvent.Get("target")
				if target.IsUndefined() || target.IsNull() { return nil }
				matched := target.Call("closest", marker)
				if matched.IsUndefined() || matched.IsNull() { return nil }
				id := matched.Call("getAttribute", "data-inline-tab").String()
				if id == "" { return nil }
				inlineHandlersMu.RLock()
				h := inlineTabHandlers[id]
				inlineHandlersMu.RUnlock()
				if h == nil { return nil }
				el := domv2.WrapElement(matched)
				if el == nil { return nil }
				defer func(){ if r := recover(); r != nil { logutil.Logf("panic in inline tab: %v", r) } }()
				h(el)
				return nil
			})
			root.Call("addEventListener", "keydown", tabFn)
			tabInstalled = true
		}
	}

	// Install for Shift+Tab navigation
	shiftTabInstalled := false
	var shiftTabFn js.Func
	var shiftTabIDs []string
	{
		marker := "[data-inline-shifttab]"
		nodes := root.Call("querySelectorAll", marker)
		if nodes.Truthy() && nodes.Get("length").Int() > 0 {
			shiftTabIDs = collect("data-inline-shifttab")
			shiftTabFn = js.FuncOf(func(this js.Value, args []js.Value) any {
				if len(args) == 0 { return nil }
				rawEvent := args[0]
				key := rawEvent.Get("key").String()
				if key != "Tab" { return nil }
				shiftKey := rawEvent.Get("shiftKey").Bool()
				if !shiftKey { return nil } // Handle only Shift+Tab
				target := rawEvent.Get("target")
				if target.IsUndefined() || target.IsNull() { return nil }
				matched := target.Call("closest", marker)
				if matched.IsUndefined() || matched.IsNull() { return nil }
				id := matched.Call("getAttribute", "data-inline-shifttab").String()
				if id == "" { return nil }
				inlineHandlersMu.RLock()
				h := inlineShiftTabHandlers[id]
				inlineHandlersMu.RUnlock()
				if h == nil { return nil }
				el := domv2.WrapElement(matched)
				if el == nil { return nil }
				defer func(){ if r := recover(); r != nil { logutil.Logf("panic in inline shift+tab: %v", r) } }()
				h(el)
				return nil
			})
			root.Call("addEventListener", "keydown", shiftTabFn)
			shiftTabInstalled = true
		}
	}

	// Install for Arrow keys navigation
	arrowInstalled := false
	var arrowFn js.Func
	var arrowIDs []string
	{
		marker := "[data-inline-arrow]"
		nodes := root.Call("querySelectorAll", marker)
		if nodes.Truthy() && nodes.Get("length").Int() > 0 {
			arrowIDs = collect("data-inline-arrow")
			arrowFn = js.FuncOf(func(this js.Value, args []js.Value) any {
				if len(args) == 0 { return nil }
				rawEvent := args[0]
				key := rawEvent.Get("key").String()
				var direction string
				switch key {
				case "ArrowUp":
					direction = "up"
				case "ArrowDown":
					direction = "down"
				case "ArrowLeft":
					direction = "left"
				case "ArrowRight":
					direction = "right"
				default:
					return nil // Not an arrow key
				}
				target := rawEvent.Get("target")
				if target.IsUndefined() || target.IsNull() { return nil }
				matched := target.Call("closest", marker)
				if matched.IsUndefined() || matched.IsNull() { return nil }
				id := matched.Call("getAttribute", "data-inline-arrow").String()
				if id == "" { return nil }
				inlineHandlersMu.RLock()
				h := inlineArrowKeyHandlers[id]
				inlineHandlersMu.RUnlock()
				if h == nil { return nil }
				el := domv2.WrapElement(matched)
				if el == nil { return nil }
				defer func(){ if r := recover(); r != nil { logutil.Logf("panic in inline arrow: %v", r) } }()
				h(el, direction)
				return nil
			})
			root.Call("addEventListener", "keydown", arrowFn)
			arrowInstalled = true
		}
	}

	// Install for Drag Start events
	dragStartInstalled := false
	var dragStartFn js.Func
	var dragStartIDs []string
	{
		marker := "[data-inline-dragstart]"
		nodes := root.Call("querySelectorAll", marker)
		if nodes.Truthy() && nodes.Get("length").Int() > 0 {
			dragStartIDs = collect("data-inline-dragstart")
			dragStartFn = js.FuncOf(func(this js.Value, args []js.Value) any {
				if len(args) == 0 { return nil }
				rawEvent := args[0]
				target := rawEvent.Get("target")
				if target.IsUndefined() || target.IsNull() { return nil }
				matched := target.Call("closest", marker)
				if matched.IsUndefined() || matched.IsNull() { return nil }
				id := matched.Call("getAttribute", "data-inline-dragstart").String()
				if id == "" { return nil }
				inlineHandlersMu.RLock()
				h := inlineDragStartHandlers[id]
				inlineHandlersMu.RUnlock()
				if h == nil { return nil }
				el := domv2.WrapElement(matched)
				if el == nil { return nil }
				dataTransfer := rawEvent.Get("dataTransfer")
				defer func(){ if r := recover(); r != nil { logutil.Logf("panic in inline dragstart: %v", r) } }()
				h(el, dataTransfer)
				return nil
			})
			root.Call("addEventListener", "dragstart", dragStartFn)
			dragStartInstalled = true
		}
	}

	// Install for Drop events
	dropInstalled := false
	var dropFn js.Func
	var dropIDs []string
	{
		marker := "[data-inline-drop]"
		nodes := root.Call("querySelectorAll", marker)
		if nodes.Truthy() && nodes.Get("length").Int() > 0 {
			dropIDs = collect("data-inline-drop")
			dropFn = js.FuncOf(func(this js.Value, args []js.Value) any {
				if len(args) == 0 { return nil }
				rawEvent := args[0]
				target := rawEvent.Get("target")
				if target.IsUndefined() || target.IsNull() { return nil }
				matched := target.Call("closest", marker)
				if matched.IsUndefined() || matched.IsNull() { return nil }
				id := matched.Call("getAttribute", "data-inline-drop").String()
				if id == "" { return nil }
				inlineHandlersMu.RLock()
				h := inlineDropHandlers[id]
				inlineHandlersMu.RUnlock()
				if h == nil { return nil }
				el := domv2.WrapElement(matched)
				if el == nil { return nil }
				dataTransfer := rawEvent.Get("dataTransfer")
				defer func(){ if r := recover(); r != nil { logutil.Logf("panic in inline drop: %v", r) } }()
				h(el, dataTransfer)
				return nil
			})
			root.Call("addEventListener", "drop", dropFn)
			dropInstalled = true
		}
	}

	// Install for Drag Over events
	dragOverInstalled := false
	var dragOverFn js.Func
	var dragOverIDs []string
	{
		marker := "[data-inline-dragover]"
		nodes := root.Call("querySelectorAll", marker)
		if nodes.Truthy() && nodes.Get("length").Int() > 0 {
			dragOverIDs = collect("data-inline-dragover")
			dragOverFn = js.FuncOf(func(this js.Value, args []js.Value) any {
				if len(args) == 0 { return nil }
				rawEvent := args[0]
				target := rawEvent.Get("target")
				if target.IsUndefined() || target.IsNull() { return nil }
				matched := target.Call("closest", marker)
				if matched.IsUndefined() || matched.IsNull() { return nil }
				id := matched.Call("getAttribute", "data-inline-dragover").String()
				if id == "" { return nil }
				inlineHandlersMu.RLock()
				h := inlineDragOverHandlers[id]
				inlineHandlersMu.RUnlock()
				if h == nil { return nil }
				el := domv2.WrapElement(matched)
				if el == nil { return nil }
				dataTransfer := rawEvent.Get("dataTransfer")
				defer func(){ if r := recover(); r != nil { logutil.Logf("panic in inline dragover: %v", r) } }()
				h(el, dataTransfer)
				return nil
			})
			root.Call("addEventListener", "dragover", dragOverFn)
			dragOverInstalled = true
		}
	}

	// Install for Outside Click events
	outsideClickInstalled := false
	var outsideClickFn js.Func
	var outsideClickIDs []string
	{
		marker := "[data-uiwgo-onoutsideclick]"
		nodes := root.Call("querySelectorAll", marker)
		if nodes.Truthy() && nodes.Get("length").Int() > 0 {
			outsideClickIDs = collect("data-uiwgo-onoutsideclick")
			outsideClickFn = js.FuncOf(func(this js.Value, args []js.Value) any {
				if len(args) == 0 { return nil }
				rawEvent := args[0]
				target := rawEvent.Get("target")
				// Check if click is outside any element with outside click handler
				elements := root.Call("querySelectorAll", marker)
				for i := 0; i < elements.Get("length").Int(); i++ {
					element := elements.Index(i)
					// If target is not contained within this element, trigger outside click
					if !element.Call("contains", target).Bool() {
						id := element.Call("getAttribute", "data-uiwgo-onoutsideclick").String()
						if id == "" { continue }
						inlineHandlersMu.RLock()
						h := inlineOutsideClickHandlers[id]
						inlineHandlersMu.RUnlock()
						if h == nil { continue }
						el := domv2.WrapElement(element)
						if el == nil { continue }
						defer func(){ if r := recover(); r != nil { logutil.Logf("panic in inline outside click: %v", r) } }()
						h(el)
					}
				}
				return nil
			})
			root.Call("addEventListener", "click", outsideClickFn)
			outsideClickInstalled = true
		}
	}

	// Install for Escape Close events
	escapeCloseInstalled := false
	var escapeCloseFn js.Func
	var escapeCloseIDs []string
	{
		marker := "[data-uiwgo-onescapeclose]"
		nodes := root.Call("querySelectorAll", marker)
		if nodes.Truthy() && nodes.Get("length").Int() > 0 {
			escapeCloseIDs = collect("data-uiwgo-onescapeclose")
			escapeCloseFn = js.FuncOf(func(this js.Value, args []js.Value) any {
				if len(args) == 0 { return nil }
				rawEvent := args[0]
				// Check if Escape key was pressed
				if rawEvent.Get("key").String() != "Escape" { return nil }
				// Trigger all escape close handlers
				elements := root.Call("querySelectorAll", marker)
				for i := 0; i < elements.Get("length").Int(); i++ {
					element := elements.Index(i)
					id := element.Call("getAttribute", "data-uiwgo-onescapeclose").String()
					if id == "" { continue }
					inlineHandlersMu.RLock()
					h := inlineEscapeCloseHandlers[id]
					inlineHandlersMu.RUnlock()
					if h == nil { continue }
					el := domv2.WrapElement(element)
					if el == nil { continue }
					defer func(){ if r := recover(); r != nil { logutil.Logf("panic in inline escape close: %v", r) } }()
					h(el)
				}
				return nil
			})
			root.Call("addEventListener", "keydown", escapeCloseFn)
			escapeCloseInstalled = true
		}
	}

	// Install for File Select events
	fileSelectInstalled := false
	var fileSelectFn js.Func
	var fileSelectIDs []string
	{
		marker := "[data-uiwgo-onfileselect]"
		nodes := root.Call("querySelectorAll", marker)
		if nodes.Truthy() && nodes.Get("length").Int() > 0 {
			fileSelectIDs = collect("data-uiwgo-onfileselect")
			fileSelectFn = js.FuncOf(func(this js.Value, args []js.Value) any {
				if len(args) == 0 { return nil }
				rawEvent := args[0]
				target := rawEvent.Get("target")
				// Check if target has file select handler and files were selected
				id := target.Call("getAttribute", "data-uiwgo-onfileselect").String()
				if id == "" { return nil }
				files := target.Get("files")
				if !files.Truthy() || files.Get("length").Int() == 0 { return nil }
				// Convert FileList to []js.Value
				fileArray := make([]js.Value, files.Get("length").Int())
				for i := 0; i < len(fileArray); i++ {
					fileArray[i] = files.Index(i)
				}
				inlineHandlersMu.RLock()
				h := inlineFileSelectHandlers[id]
				inlineHandlersMu.RUnlock()
				if h == nil { return nil }
				el := domv2.WrapElement(target)
				if el == nil { return nil }
				defer func(){ if r := recover(); r != nil { logutil.Logf("panic in inline file select: %v", r) } }()
				h(el, fileArray)
				return nil
			})
			root.Call("addEventListener", "change", fileSelectFn)
			fileSelectInstalled = true
		}
	}

	// Install for File Drop events
	fileDropInstalled := false
	var fileDropFn js.Func
	var dragOverPreventFn js.Func
	var fileDropIDs []string
	{
		marker := "[data-uiwgo-onfiledrop]"
		nodes := root.Call("querySelectorAll", marker)
		if nodes.Truthy() && nodes.Get("length").Int() > 0 {
			fileDropIDs = collect("data-uiwgo-onfiledrop")
			fileDropFn = js.FuncOf(func(this js.Value, args []js.Value) any {
				if len(args) == 0 { return nil }
				rawEvent := args[0]
				rawEvent.Call("preventDefault") // Prevent default drop behavior
				target := rawEvent.Get("target")
				// Find the closest element with file drop handler
				var dropElement js.Value
				current := target
				for current.Truthy() && current.Get("nodeType").Int() == 1 {
					if current.Call("hasAttribute", "data-uiwgo-onfiledrop").Bool() {
						dropElement = current
						break
					}
					current = current.Get("parentElement")
				}
				if !dropElement.Truthy() { return nil }
				id := dropElement.Call("getAttribute", "data-uiwgo-onfiledrop").String()
				if id == "" { return nil }
				dataTransfer := rawEvent.Get("dataTransfer")
				if !dataTransfer.Truthy() { return nil }
				files := dataTransfer.Get("files")
				if !files.Truthy() || files.Get("length").Int() == 0 { return nil }
				// Convert FileList to []js.Value
				fileArray := make([]js.Value, files.Get("length").Int())
				for i := 0; i < len(fileArray); i++ {
					fileArray[i] = files.Index(i)
				}
				inlineHandlersMu.RLock()
				h := inlineFileDropHandlers[id]
				inlineHandlersMu.RUnlock()
				if h == nil { return nil }
				el := domv2.WrapElement(dropElement)
				if el == nil { return nil }
				defer func(){ if r := recover(); r != nil { logutil.Logf("panic in inline file drop: %v", r) } }()
				h(el, fileArray)
				return nil
			})
			root.Call("addEventListener", "drop", fileDropFn)
			// Also prevent default dragover to allow drop
			dragOverPreventFn = js.FuncOf(func(this js.Value, args []js.Value) any {
				if len(args) == 0 { return nil }
				rawEvent := args[0]
				target := rawEvent.Get("target")
				// Check if target or ancestor has file drop handler
				current := target
				for current.Truthy() && current.Get("nodeType").Int() == 1 {
					if current.Call("hasAttribute", "data-uiwgo-onfiledrop").Bool() {
						rawEvent.Call("preventDefault")
						break
					}
					current = current.Get("parentElement")
				}
				return nil
			})
			root.Call("addEventListener", "dragover", dragOverPreventFn)
			fileDropInstalled = true
		}
	}

	// Cleanup
	reactivity.OnCleanup(func() {
		if clickInstalled {
			root.Call("removeEventListener", "click", clickFn)
			clickFn.Release()
			inlineHandlersMu.Lock()
			for _, id := range clickIDs { delete(inlineClickHandlers, id) }
			inlineHandlersMu.Unlock()
		}
		if inputInstalled {
			root.Call("removeEventListener", "input", inputFn)
			inputFn.Release()
			inlineHandlersMu.Lock()
			for _, id := range inputIDs { delete(inlineInputHandlers, id) }
			inlineHandlersMu.Unlock()
		}
		if changeInstalled {
			root.Call("removeEventListener", "change", changeFn)
			changeFn.Release()
			inlineHandlersMu.Lock()
			for _, id := range changeIDs { delete(inlineChangeHandlers, id) }
			inlineHandlersMu.Unlock()
		}
		if keydownInstalled {
			root.Call("removeEventListener", "keydown", keydownFn)
			keydownFn.Release()
			inlineHandlersMu.Lock()
			for _, id := range keydownIDs {
				delete(inlineKeydownHandlers, id)
				delete(inlineKeyExpectations, id)
			}
			inlineHandlersMu.Unlock()
		}
		if enterInstalled {
			root.Call("removeEventListener", "keydown", enterFn)
			enterFn.Release()
			root.Call("removeEventListener", "keyup", enterFnUp)
			enterFnUp.Release()
			inlineHandlersMu.Lock()
			for _, id := range enterIDs {
				delete(inlineKeydownHandlers, id)
				delete(inlineKeyExpectations, id)
			}
			inlineHandlersMu.Unlock()
		}
		if escapeInstalled {
			root.Call("removeEventListener", "keydown", escapeFn)
			escapeFn.Release()
			root.Call("removeEventListener", "keyup", escapeFnUp)
			escapeFnUp.Release()
			inlineHandlersMu.Lock()
			for _, id := range escapeIDs {
				delete(inlineKeydownHandlers, id)
				delete(inlineKeyExpectations, id)
			}
			inlineHandlersMu.Unlock()
		}
		if submitInstalled {
			root.Call("removeEventListener", "submit", submitFn)
			submitFn.Release()
			inlineHandlersMu.Lock()
			for _, id := range submitIDs { delete(inlineSubmitHandlers, id) }
			inlineHandlersMu.Unlock()
		}
		if resetInstalled {
			root.Call("removeEventListener", "reset", resetFn)
			resetFn.Release()
			inlineHandlersMu.Lock()
			for _, id := range resetIDs { delete(inlineFormResetHandlers, id) }
			inlineHandlersMu.Unlock()
		}
		if blurInstalled {
			root.Call("removeEventListener", "blur", blurFn)
			blurFn.Release()
			inlineHandlersMu.Lock()
			for _, id := range blurIDs { delete(inlineBlurHandlers, id) }
			inlineHandlersMu.Unlock()
		}
		if focusInstalled {
			root.Call("removeEventListener", "focus", focusFn)
			focusFn.Release()
			inlineHandlersMu.Lock()
			for _, id := range focusIDs { delete(inlineFocusHandlers, id) }
			inlineHandlersMu.Unlock()
		}
		if focusWithinInstalled {
			root.Call("removeEventListener", "focusin", focusInFn)
			root.Call("removeEventListener", "focusout", focusOutFn)
			focusInFn.Release()
			focusOutFn.Release()
			inlineHandlersMu.Lock()
			for _, id := range focusWithinIDs { delete(inlineFocusWithinHandlers, id) }
			inlineHandlersMu.Unlock()
		}
		if formChangeInstalled {
			root.Call("removeEventListener", "change", formChangeFn)
			formChangeFn.Release()
			inlineHandlersMu.Lock()
			for _, id := range formChangeIDs { delete(inlineFormChangeHandlers, id) }
			inlineHandlersMu.Unlock()
		}
		if validateInstalled {
			root.Call("removeEventListener", "input", validateFn)
			validateFn.Release()
			inlineHandlersMu.Lock()
			for _, id := range validateIDs { delete(inlineValidateHandlers, id) }
			inlineHandlersMu.Unlock()
		}
		if blurValidateInstalled {
			root.Call("removeEventListener", "blur", blurValidateFn)
			blurValidateFn.Release()
			inlineHandlersMu.Lock()
			for _, id := range blurValidateIDs { delete(inlineBlurValidateHandlers, id) }
			inlineHandlersMu.Unlock()
		}
		if debouncedInstalled {
			root.Call("removeEventListener", "input", debouncedFn)
			debouncedFn.Release()
			inlineHandlersMu.Lock()
			for _, id := range debouncedIDs {
				if timer, exists := inlineDebounceTimers[id]; exists && !timer.IsUndefined() {
					js.Global().Call("clearTimeout", timer)
				}
				delete(inlineDebouncedInputHandlers, id)
				delete(inlineDebounceTimers, id)
			}
			inlineHandlersMu.Unlock()
		}
		if searchInstalled {
			root.Call("removeEventListener", "input", searchFn)
			searchFn.Release()
			inlineHandlersMu.Lock()
			for _, id := range searchIDs {
				if timer, exists := inlineDebounceTimers[id]; exists && !timer.IsUndefined() {
					js.Global().Call("clearTimeout", timer)
				}
				delete(inlineSearchHandlers, id)
				delete(inlineDebounceTimers, id)
			}
			inlineHandlersMu.Unlock()
		}
		if tabInstalled {
			root.Call("removeEventListener", "keydown", tabFn)
			tabFn.Release()
			inlineHandlersMu.Lock()
			for _, id := range tabIDs { delete(inlineTabHandlers, id) }
			inlineHandlersMu.Unlock()
		}
		if shiftTabInstalled {
			root.Call("removeEventListener", "keydown", shiftTabFn)
			shiftTabFn.Release()
			inlineHandlersMu.Lock()
			for _, id := range shiftTabIDs { delete(inlineShiftTabHandlers, id) }
			inlineHandlersMu.Unlock()
		}
		if arrowInstalled {
			root.Call("removeEventListener", "keydown", arrowFn)
			arrowFn.Release()
			inlineHandlersMu.Lock()
			for _, id := range arrowIDs { delete(inlineArrowKeyHandlers, id) }
			inlineHandlersMu.Unlock()
		}
		if dragStartInstalled {
			root.Call("removeEventListener", "dragstart", dragStartFn)
			dragStartFn.Release()
			inlineHandlersMu.Lock()
			for _, id := range dragStartIDs { delete(inlineDragStartHandlers, id) }
			inlineHandlersMu.Unlock()
		}
		if dropInstalled {
			root.Call("removeEventListener", "drop", dropFn)
			dropFn.Release()
			inlineHandlersMu.Lock()
			for _, id := range dropIDs { delete(inlineDropHandlers, id) }
			inlineHandlersMu.Unlock()
		}
		if dragOverInstalled {
			root.Call("removeEventListener", "dragover", dragOverFn)
			dragOverFn.Release()
			inlineHandlersMu.Lock()
			for _, id := range dragOverIDs { delete(inlineDragOverHandlers, id) }
			inlineHandlersMu.Unlock()
		}
		if outsideClickInstalled {
			root.Call("removeEventListener", "click", outsideClickFn)
			outsideClickFn.Release()
			inlineHandlersMu.Lock()
			for _, id := range outsideClickIDs { delete(inlineOutsideClickHandlers, id) }
			inlineHandlersMu.Unlock()
		}
		if escapeCloseInstalled {
			root.Call("removeEventListener", "keydown", escapeCloseFn)
			escapeCloseFn.Release()
			inlineHandlersMu.Lock()
			for _, id := range escapeCloseIDs { delete(inlineEscapeCloseHandlers, id) }
			inlineHandlersMu.Unlock()
		}
		if fileSelectInstalled {
			root.Call("removeEventListener", "change", fileSelectFn)
			fileSelectFn.Release()
			inlineHandlersMu.Lock()
			for _, id := range fileSelectIDs { delete(inlineFileSelectHandlers, id) }
			inlineHandlersMu.Unlock()
		}
		if fileDropInstalled {
			root.Call("removeEventListener", "drop", fileDropFn)
			root.Call("removeEventListener", "dragover", dragOverPreventFn)
			fileDropFn.Release()
			dragOverPreventFn.Release()
			inlineHandlersMu.Lock()
			for _, id := range fileDropIDs { delete(inlineFileDropHandlers, id) }
			inlineHandlersMu.Unlock()
		}
	})
}
