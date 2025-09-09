//go:build js && wasm

package react

import (
    "errors"
    "testing"
)

func TestGetBridge_NotInitialized_ReturnsSentinelError(t *testing.T) {
    // Ensure not initialized
    globalBridge = nil

    _, err := GetBridge()
    if err == nil {
        t.Fatalf("expected error when bridge not initialized")
    }
    if !errors.Is(err, ErrBridgeNotInitialized) {
        t.Fatalf("expected ErrBridgeNotInitialized, got %v", err)
    }
}

func TestBind_InvalidArgs_ReturnsSentinelErrors(t *testing.T) {
    // Missing component ID
    _, err := Bind(BindingOptions{})
    if !errors.Is(err, ErrComponentIDRequired) {
        t.Fatalf("expected ErrComponentIDRequired, got %v", err)
    }

    // Missing component name
    _, err = Bind(BindingOptions{ComponentID: "cid"})
    if !errors.Is(err, ErrComponentNameRequired) {
        t.Fatalf("expected ErrComponentNameRequired, got %v", err)
    }

    // Missing props function
    _, err = Bind(BindingOptions{ComponentID: "cid", ComponentName: "Name"})
    if !errors.Is(err, ErrPropsFuncRequired) {
        t.Fatalf("expected ErrPropsFuncRequired, got %v", err)
    }
}
