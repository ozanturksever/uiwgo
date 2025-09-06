//go:build js && wasm

package appmanager

import (
	"fmt"

	"github.com/ozanturksever/logutil"
	"github.com/ozanturksever/uiwgo/comps"
	"github.com/ozanturksever/uiwgo/router"
	dom "honnef.co/go/js/dom/v2"
	g "maragu.dev/gomponents"
)

// Mount renders the root component into MountElementID and initializes the router if enabled
func (am *AppManager) Mount(root func() g.Node) error {
	if !am.initialized.Get() {
		return fmt.Errorf("app manager not initialized")
	}

	// beforeMount hooks
	if err := am.lifecycle.ExecuteHooks(EventBeforeMount, &LifecycleContext{Event: EventBeforeMount, Manager: am}); err != nil {
		return fmt.Errorf("beforeMount hooks failed: %w", err)
	}

	// Mount component via comps
	am.disposer = comps.Mount(am.config.MountElementID, root)

	// Setup router after mount if enabled
	if am.config.EnableRouter && len(am.config.Routes) > 0 {
		// Prefer #router-outlet, fallback to mount element
		outlet := dom.GetWindow().Document().GetElementByID("router-outlet")
		if outlet == nil {
			outlet = dom.GetWindow().Document().GetElementByID(am.config.MountElementID)
		}
		logutil.Log("outlet:", outlet, "id:", am.config.MountElementID)
		if outlet != nil {
			am.router = router.New(am.config.Routes, outlet)
			// Wire router navigation callbacks to lifecycle hooks and store updates
			am.router.OnBeforeNavigate = func(path string, options router.NavigateOptions) {
				if err := am.lifecycle.ExecuteHooks(EventBeforeRoute, &LifecycleContext{Event: EventBeforeRoute, Manager: am, Data: map[string]any{"path": path}}); err != nil {
					logutil.Logf("beforeRoute hooks failed: %v", err)
				}
			}
			am.router.OnAfterNavigate = func(path string, options router.NavigateOptions) {
				// Update router state snapshot
				st := am.store.Get()
				st.Router.PreviousPath = st.Router.CurrentPath
				st.Router.CurrentPath = path
				am.store.Replace(st)
				if err := am.lifecycle.ExecuteHooks(EventAfterRoute, &LifecycleContext{Event: EventAfterRoute, Manager: am, Data: map[string]any{"path": path}}); err != nil {
					logutil.Logf("afterRoute hooks failed: %v", err)
				}
			}
		}
	}

	am.running.Set(true)
	am.lifecycle.setState(LifecycleStateRunning)

	// afterMount hooks
	if err := am.lifecycle.ExecuteHooks(EventAfterMount, &LifecycleContext{Event: EventAfterMount, Manager: am}); err != nil {
		logutil.Logf("afterMount hooks failed: %v", err)
	}
	return nil
}
