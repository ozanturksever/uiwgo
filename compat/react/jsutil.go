//go:build js && wasm

package react

import (
    "fmt"
    "reflect"
    "syscall/js"
)

// MapToJSObject converts a Go map[string]interface{} into a native JavaScript Object recursively,
// preserving js.Value and js.Func values and converting slices to JS arrays.
func MapToJSObject(m map[string]interface{}) js.Value {
    obj := js.Global().Get("Object").New()
    for k, v := range m {
        obj.Set(k, toJS(v))
    }
    return obj
}

// toJS converts Go values to something acceptable for js.Value.Set/Call:
// - js.Value and js.Func are passed through
// - maps become JS objects
// - slices/arrays become JS arrays
// - primitives are passed directly
// - nil becomes null
// - unsupported structs and complex types fall back to string representation
func toJS(v interface{}) interface{} {
    if v == nil {
        return js.Null()
    }

    switch x := v.(type) {
    case js.Value:
        return x
    case js.Func:
        return x
    case string, bool, int, int8, int16, int32, int64,
        uint, uint8, uint16, uint32, uint64,
        float32, float64:
        return x
    case map[string]interface{}:
        return MapToJSObject(x)
    case []interface{}:
        arr := js.Global().Get("Array").New(len(x))
        for i, el := range x {
            arr.SetIndex(i, toJS(el))
        }
        return arr
    }

    // Handle typed slices/arrays using reflection
    rv := reflect.ValueOf(v)
    switch rv.Kind() {
    case reflect.Slice, reflect.Array:
        l := rv.Len()
        arr := js.Global().Get("Array").New(l)
        for i := 0; i < l; i++ {
            arr.SetIndex(i, toJS(rv.Index(i).Interface()))
        }
        return arr
    case reflect.Map:
        // Only support map with string keys; others are stringified
        if rv.Type().Key().Kind() == reflect.String {
            obj := js.Global().Get("Object").New()
            iter := rv.MapRange()
            for iter.Next() {
                k := iter.Key().String()
                obj.Set(k, toJS(iter.Value().Interface()))
            }
            return obj
        }
        // Fallback string for unsupported map key kinds
        return fmt.Sprint(v)
    case reflect.Pointer:
        if rv.IsNil() {
            return js.Null()
        }
        return toJS(rv.Elem().Interface())
    case reflect.Struct:
        // Not supported: convert to string representation to avoid panics
        return fmt.Sprint(v)
    default:
        // Best-effort fallback; ValueOf will panic on unsupported types, so avoid it
        return fmt.Sprint(v)
    }
}
