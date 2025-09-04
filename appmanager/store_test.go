package appmanager

import "testing"

func TestAppStore_Creation(t *testing.T) {
    initialState := AppState{
        User: map[string]any{"name": "test"},
    }

    store := NewAppStore(initialState, "test-key")

    if store == nil {
        t.Fatal("store is nil")
    }
    got := store.Get().User.(map[string]any)["name"]
    if got != "test" {
        t.Fatalf("expected user.name 'test', got %v", got)
    }
}
