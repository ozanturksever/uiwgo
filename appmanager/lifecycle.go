package appmanager

import (
	"fmt"
	"strings"

	"github.com/ozanturksever/logutil"
	"github.com/ozanturksever/uiwgo/reactivity"
)

// NewLifecycleManager constructs a LifecycleManager with default state
func NewLifecycleManager() *LifecycleManager {
	lm := &LifecycleManager{
		hooks: make(map[LifecycleEvent][]LifecycleHook),
		state: reactivity.CreateSignal[LifecycleState](LifecycleStateUninitialized),
	}

	// Simple log when state changes
	_ = reactivity.CreateSignal(0) // no-op to keep linter happy if effects are added later
	return lm
}

// AddHook registers a hook for the given event
func (lm *LifecycleManager) AddHook(event LifecycleEvent, hook LifecycleHook) {
	lm.hooks[event] = append(lm.hooks[event], hook)
}

// ExecuteHooks runs all hooks for an event and aggregates errors
func (lm *LifecycleManager) ExecuteHooks(event LifecycleEvent, ctx *LifecycleContext) error {
	hooks := lm.hooks[event]
	if len(hooks) == 0 {
		return nil
	}

	var errs []string
	for i, h := range hooks {
		if err := h(ctx); err != nil {
			msg := fmt.Sprintf("hook %d for event %s: %v", i, event, err)
			logutil.Logf("Lifecycle hook error: %s", msg)
			errs = append(errs, msg)
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("lifecycle hook failures: %s", strings.Join(errs, "; "))
	}
	return nil
}

// GetState returns the current lifecycle state
func (lm *LifecycleManager) GetState() LifecycleState {
	return lm.state.Get()
}

// setState updates the lifecycle state
func (lm *LifecycleManager) setState(s LifecycleState) { lm.state.Set(s) }
