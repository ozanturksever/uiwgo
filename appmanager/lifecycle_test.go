package appmanager

import (
    "testing"
)

func TestLifecycleManager_AddHook(t *testing.T) {
    lm := NewLifecycleManager()

    called := false
    hook := func(ctx *LifecycleContext) error {
        called = true
        return nil
    }

    lm.AddHook(EventBeforeInit, hook)

    err := lm.ExecuteHooks(EventBeforeInit, &LifecycleContext{Event: EventBeforeInit})

    if err != nil {
        t.Fatalf("Unexpected error: %v", err)
    }
    if !called {
        t.Fatalf("Expected hook to be called")
    }
}
