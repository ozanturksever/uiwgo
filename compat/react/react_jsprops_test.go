//go:build js && wasm

package react

import (
    "testing"
    "syscall/js"
)

func TestRender_WithJSFuncProp_DoesNotMarshalJSON(t *testing.T) {
    // Set up mock ReactCompat on window
    setupMockReactCompat()

    // Override renderComponent to validate props types on JS side
    cb := js.FuncOf(func(this js.Value, args []js.Value) interface{} { return nil })
    defer cb.Release()

    node := js.Global().Get("Object").New()
    node.Set("marker", 1)

    js.Global().Get("window").Get("ReactCompat").Set("renderComponent", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
        if len(args) < 2 {
            return nil
        }
        props := args[1]
        // onClick must be a function, node must be exactly the same object
        if props.Get("onClick").Type() != js.TypeFunction {
            return nil
        }
        if !props.Get("node").Equal(node) {
            return nil
        }
        return "ok-123"
    }))

    // Initialize global bridge
    if err := InitializeBridge(); err != nil {
        t.Fatalf("InitializeBridge failed: %v", err)
    }

    props := Props{
        "label": "Click",
        "onClick": cb, // must survive as function
        "node": node,  // must be passed through by reference
    }

    // Should succeed if props are converted without JSON marshalling
    if _, err := Render("Clickable", props, nil); err != nil {
        t.Fatalf("Render failed with JS interop props: %v", err)
    }
}

func TestUpdate_WithJSObjectProp_DoesNotMarshalJSON(t *testing.T) {
    setupMockReactCompat()

    // Override updateComponent to validate props types on JS side
    node := js.Global().Get("Object").New()
    node.Set("k", 7)

    js.Global().Get("window").Get("ReactCompat").Set("updateComponent", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
        if len(args) < 2 {
            return false
        }
        props := args[1]
        if !props.Get("node").Equal(node) {
            return false
        }
        return true
    }))

    if err := InitializeBridge(); err != nil {
        t.Fatalf("InitializeBridge failed: %v", err)
    }

    // First render a component (renderComponent can be simple)
    id, err := Render("Widget", Props{"x": 1}, nil)
    if err != nil {
        t.Fatalf("initial Render failed: %v", err)
    }

    // Update with a JS object in props; must not rely on JSON marshalling
    if err := Update(id, Props{"node": node}); err != nil {
        t.Fatalf("Update failed with js.Value prop: %v", err)
    }
}
