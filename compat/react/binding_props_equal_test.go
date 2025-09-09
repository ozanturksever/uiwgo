//go:build js && wasm

package react

import (
    "syscall/js"
    "testing"
)

func TestPropsEqual_Primitives(t *testing.T) {
    b := &ComponentBinding{}

    a := map[string]interface{}{"s": "x", "i": 3, "b": true, "f": 2.5}
    c := map[string]interface{}{"s": "x", "i": 3, "b": true, "f": 2.5}

    if !b.propsEqual(a, c) {
        t.Fatalf("expected primitive props to be equal")
    }

    // type-aware: string "1" should not equal number 1
    d := map[string]interface{}{"v": "1"}
    e := map[string]interface{}{"v": 1}
    if b.propsEqual(d, e) {
        t.Fatalf("expected string vs int not equal")
    }

    // float vs int equal by numeric value
    f1 := map[string]interface{}{"n": 1}
    f2 := map[string]interface{}{"n": 1.0}
    if !b.propsEqual(f1, f2) {
        t.Fatalf("expected numeric 1 and 1.0 to be equal")
    }
}

func TestPropsEqual_JSValue(t *testing.T) {
    b := &ComponentBinding{}

    obj := js.Global().Get("Object").New()
    obj.Set("k", 1)

    same := map[string]interface{}{"node": obj}
    copy := map[string]interface{}{"node": obj}
    if !b.propsEqual(same, copy) {
        t.Fatalf("expected same js.Value reference to be equal")
    }

    other := js.Global().Get("Object").New()
    a := map[string]interface{}{"node": obj}
    c := map[string]interface{}{"node": other}
    if b.propsEqual(a, c) {
        t.Fatalf("expected different js.Value objects to be not equal")
    }
}

func TestPropsEqual_Nested(t *testing.T) {
    b := &ComponentBinding{}

    a := map[string]interface{}{
        "user": map[string]interface{}{
            "name": "Ada",
            "tags": []interface{}{"x", 2, true},
        },
    }
    c := map[string]interface{}{
        "user": map[string]interface{}{
            "name": "Ada",
            "tags": []interface{}{"x", 2, true},
        },
    }
    if !b.propsEqual(a, c) {
        t.Fatalf("expected nested equal")
    }

    d := map[string]interface{}{
        "user": map[string]interface{}{
            "name": "Ada",
            "tags": []interface{}{"x", 2, false}, // differs
        },
    }
    if b.propsEqual(a, d) {
        t.Fatalf("expected nested not equal when element differs")
    }
}
