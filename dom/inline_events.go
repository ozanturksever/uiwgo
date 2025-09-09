//go:build js && wasm

package dom

import (
	"strconv"
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
	inlineInputHandlers  = map[string]func(Element){}
	inlineChangeHandlers = map[string]func(Element){}
	inlineKeydownHandlers = map[string]func(Element){}
	inlineKeyExpectations = map[string]string{} // id -> expected key (optional)
	inlineHandlersMu     sync.RWMutex
)

func nextInlineID(prefix string) string {
	id := atomic.AddUint64(&inlineIDCounter, 1)
	return prefix + "_" + strconv.FormatUint(id, 36)
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
	})
}
