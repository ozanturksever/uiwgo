package appmanager

import "github.com/ozanturksever/uiwgo/reactivity"

// AppStore wraps the generic reactivity store for AppState
// and provides a minimal API used by tests and manager.
type AppStore struct {
    state    reactivity.Store[AppState]
    setState func(...any)
}

func NewAppStore(initial AppState, _ string) *AppStore {
    st, set := reactivity.CreateStore(initial)
    return &AppStore{state: st, setState: set}
}

func (s *AppStore) Get() AppState { return s.state.Get() }

// Set updates nested state using path semantics from reactivity.CreateStore
func (s *AppStore) Set(pathAndValue ...any) { s.setState(pathAndValue...) }

// Replace replaces the entire AppState
func (s *AppStore) Replace(st AppState) { s.setState(st) }
