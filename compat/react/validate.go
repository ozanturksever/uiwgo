//go:build js && wasm

package react

import (
    "errors"
    "fmt"
    "reflect"
    "syscall/js"
)

// validateProps walks the props map and returns an error if a value contains
// types that cannot be reasonably serialized or passed to JS (e.g., channels).
// We allow js.Value/js.Func passthrough, primitives, strings, maps with string keys,
// slices/arrays, pointers (validated by element), and nil.
func validateProps(m map[string]interface{}) error {
    if m == nil {
        return nil
    }
    for k, v := range m {
        if err := validateValue(v); err != nil {
            return fmt.Errorf("prop %q: %w", k, err)
        }
    }
    return nil
}

func validateValue(v interface{}) error {
    if v == nil {
        return nil
    }
    switch v.(type) {
    case js.Value, js.Func:
        return nil
    case string, bool, int, int8, int16, int32, int64,
        uint, uint8, uint16, uint32, uint64,
        float32, float64:
        return nil
    case map[string]interface{}:
        return validateProps(v.(map[string]interface{}))
    case []interface{}:
        for _, el := range v.([]interface{}) {
            if err := validateValue(el); err != nil {
                return err
            }
        }
        return nil
    }

    rv := reflect.ValueOf(v)
    switch rv.Kind() {
    case reflect.Slice, reflect.Array:
        l := rv.Len()
        for i := 0; i < l; i++ {
            if err := validateValue(rv.Index(i).Interface()); err != nil {
                return err
            }
        }
        return nil
    case reflect.Map:
        // Only support string keys
        if rv.Type().Key().Kind() != reflect.String {
            return errors.New("map with non-string keys is not supported")
        }
        iter := rv.MapRange()
        for iter.Next() {
            if err := validateValue(iter.Value().Interface()); err != nil {
                return err
            }
        }
        return nil
    case reflect.Pointer:
        if rv.IsNil() {
            return nil
        }
        return validateValue(rv.Elem().Interface())
    case reflect.Struct:
        // Structs would require explicit conversion; treat as unsupported for now
        return errors.New("struct value is not supported")
    case reflect.Chan, reflect.Func, reflect.UnsafePointer:
        return errors.New("unsupported value type")
    default:
        return nil
    }
}
