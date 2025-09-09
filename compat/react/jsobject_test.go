//go:build js && wasm

package react

import (
    "syscall/js"
    "testing"
)

func TestMapToJSObject_Primitives(t *testing.T) {
    props := map[string]interface{}{
        "s": "hello",
        "i": 42,
        "f": 3.14,
        "b": true,
        "n": nil,
    }

    obj := MapToJSObject(props)
    if obj.IsUndefined() || obj.IsNull() {
        t.Fatal("MapToJSObject returned undefined/null")
    }
    if got := obj.Get("s").String(); got != "hello" {
        t.Fatalf("s mismatch: got %q", got)
    }
    if got := obj.Get("i").Int(); got != 42 {
        t.Fatalf("i mismatch: got %d", got)
    }
    if got := obj.Get("f").Float(); got != 3.14 {
        t.Fatalf("f mismatch: got %v", got)
    }
    if got := obj.Get("b").Bool(); got != true {
        t.Fatalf("b mismatch: got %v", got)
    }
    if v := obj.Get("n"); !v.IsUndefined() && !v.IsNull() {
        t.Fatalf("n expected undefined or null, got %v", v)
    }
}

func TestMapToJSObject_Nested(t *testing.T) {
    props := map[string]interface{}{
        "user": map[string]interface{}{
            "name": "Ada",
            "age": 37,
            "meta": map[string]interface{}{
                "active": true,
            },
        },
        "items": []interface{}{"a", 2, true},
    }

    obj := MapToJSObject(props)

    user := obj.Get("user")
    if user.IsUndefined() || user.IsNull() {
        t.Fatal("user is undefined/null")
    }
    if got := user.Get("name").String(); got != "Ada" {
        t.Fatalf("user.name mismatch: %q", got)
    }
    if got := user.Get("age").Int(); got != 37 {
        t.Fatalf("user.age mismatch: %d", got)
    }
    if got := user.Get("meta").Get("active").Bool(); !got {
        t.Fatalf("user.meta.active mismatch: %v", got)
    }

    items := obj.Get("items")
    if l := items.Get("length").Int(); l != 3 {
        t.Fatalf("items length mismatch: %d", l)
    }
    if items.Index(0).String() != "a" || items.Index(1).Int() != 2 || items.Index(2).Bool() != true {
        t.Fatalf("items content mismatch: %v, %v, %v", items.Index(0), items.Index(1), items.Index(2))
    }
}

func TestMapToJSObject_JSValuePassthrough(t *testing.T) {
    node := js.Global().Get("Object").New()
    node.Set("k", 1)

    cb := js.FuncOf(func(this js.Value, args []js.Value) interface{} { return nil })
    defer cb.Release()

    props := map[string]interface{}{
        "node": node,
        "cb":   cb,
    }

    obj := MapToJSObject(props)

    if !obj.Get("node").Equal(node) {
        t.Fatalf("node not passed through correctly")
    }
    if obj.Get("cb").Type() != js.TypeFunction {
        t.Fatalf("cb is not a function in JS object")
    }
}

func TestMapToJSObject_TypedSlices(t *testing.T) {
    props := map[string]interface{}{
        "nums": []int{1, 2, 3},
        "flags": []bool{true, false},
        "names": []string{"x", "y"},
    }

    obj := MapToJSObject(props)

    nums := obj.Get("nums")
    if nums.Get("length").Int() != 3 || nums.Index(0).Int() != 1 || nums.Index(2).Int() != 3 {
        t.Fatalf("nums mismatch: %v", nums)
    }
    flags := obj.Get("flags")
    if flags.Get("length").Int() != 2 || flags.Index(0).Bool() != true || flags.Index(1).Bool() != false {
        t.Fatalf("flags mismatch: %v", flags)
    }
    names := obj.Get("names")
    if names.Get("length").Int() != 2 || names.Index(0).String() != "x" || names.Index(1).String() != "y" {
        t.Fatalf("names mismatch: %v", names)
    }
}
