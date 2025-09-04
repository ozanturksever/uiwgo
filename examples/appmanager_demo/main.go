//go:build js && wasm

package main

import (
	"context"
	"strconv"
	"time"

	"github.com/ozanturksever/uiwgo/appmanager"
	"github.com/ozanturksever/uiwgo/bridge"
	"github.com/ozanturksever/uiwgo/comps"
	"github.com/ozanturksever/uiwgo/dom"
	"github.com/ozanturksever/uiwgo/logutil"
	"github.com/ozanturksever/uiwgo/reactivity"
	"github.com/ozanturksever/uiwgo/router"
	"github.com/ozanturksever/uiwgo/wasm"
	g "maragu.dev/gomponents"
	h "maragu.dev/gomponents/html"
)

func main() {
	// Initialize WASM and bridge
	if err := wasm.QuickInit(); err != nil {
		logutil.Logf("Failed to initialize WASM: %v", err)
		return
	}
	bridge.InitializeManager(bridge.NewRealManager())

	// Build AppManager config
	cfg := &appmanager.AppConfig{
		AppID:             "appmanager-demo",
		MountElementID:    "app",
		EnableRouter:      true,
		EnablePersistence: false,
		Timeout:           20 * time.Second,
		Routes: []*router.RouteDefinition{
			router.Route("/", HomeComponent),
			router.Route("/about", AboutComponent),
			router.Route("/users/:id", UserComponent),
		},
		InitialState: appmanager.AppState{
			UI: appmanager.UIState{Theme: "light"},
			Custom: map[string]any{"counter": 0},
		},
	}

	am := appmanager.NewAppManager(cfg)

	// Hooks to toggle loading during navigation
	am.AddHook(appmanager.EventBeforeRoute, func(ctx *appmanager.LifecycleContext) error {
		st := am.GetState()
		st.UI.Loading = true
		am.SetState(st)
		return nil
	})
	am.AddHook(appmanager.EventAfterRoute, func(ctx *appmanager.LifecycleContext) error {
		st := am.GetState()
		st.UI.Loading = false
		am.SetState(st)
		return nil
	})

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	if err := am.Initialize(ctx); err != nil {
		logutil.Logf("Failed to Initialize AppManager: %v", err)
		return
	}

	// Mount root; keep a persistent counter UI outside router outlet
	if err := am.Mount(func() g.Node { return RootComponent() }); err != nil {
		logutil.Logf("Failed to mount: %v", err)
		return
	}

	// Cleanup on unload
	reactivity.RegisterCleanup(func() { am.Cleanup() })

	select {}
}

// RootComponent: header/nav, persistent counter UI, and router outlet
func RootComponent() g.Node {
	// Reactive counter state
	count := reactivity.CreateSignal(0)
	onInc := func() { count.Set(count.Get() + 1) }
	onDec := func() { count.Set(count.Get() - 1) }
	onReset := func() { count.Set(0) }

	return h.Div(
		h.ID("demo-root"),
		h.Class("min-h-screen bg-gradient-to-br from-slate-50 to-slate-100 text-slate-800"),
		// Page container
		h.Div(
			h.Class("max-w-5xl mx-auto p-6 space-y-6"),

			// Header / Nav
			h.Header(
				h.Class("bg-white/80 backdrop-blur border rounded-xl shadow-sm px-4 py-3"),
				h.Nav(
					h.Class("flex items-center justify-between gap-4"),
					// Brand
					h.Div(
						h.Class("flex items-center gap-2"),
						h.Span(h.Class("inline-flex h-8 w-8 items-center justify-center rounded-lg bg-indigo-600 text-white font-bold"), g.Text("AM")),
						h.Strong(h.Class("text-slate-900"), g.Text("App Manager Demo")),
					),
					// Links
					h.Div(
						h.Class("flex items-center gap-1"),
						h.A(h.Href("/"), h.Class("px-3 py-2 rounded-md text-slate-600 hover:text-slate-900 hover:bg-slate-100"), g.Text("Home")),
						h.A(h.Href("/about"), h.Class("px-3 py-2 rounded-md text-slate-600 hover:text-slate-900 hover:bg-slate-100"), g.Text("About")),
						h.A(h.Href("/users/123"), h.Class("px-3 py-2 rounded-md text-slate-600 hover:text-slate-900 hover:bg-slate-100"), g.Text("Profile")),
					),
				),
			),

			// Persistent Counter Card (tests rely on these IDs)
			h.Div(
				h.Class("bg-white border rounded-xl shadow p-4 flex flex-wrap items-center justify-between gap-4"),
				h.Div(
					h.Class("flex items-baseline gap-2"),
					h.Span(h.Class("text-slate-600"), g.Text("Count:")),
					h.Span(
						h.ID("counter-text"),
						h.Class("text-2xl font-semibold text-slate-900"),
						comps.BindText(func() string { return "Count: " + strconv.Itoa(count.Get()) }),
					),
				),
				h.Div(
					h.Class("flex items-center gap-2"),
					h.Button(h.ID("inc-btn"), h.Class("inline-flex items-center rounded-lg bg-indigo-600 text-white px-3 py-2 text-sm font-medium hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-indigo-500"), g.Text("Increment")),
					h.Button(h.ID("dec-btn"), h.Class("inline-flex items-center rounded-lg bg-slate-200 text-slate-800 px-3 py-2 text-sm font-medium hover:bg-slate-300 focus:outline-none focus:ring-2 focus:ring-slate-400"), g.Text("Decrement")),
					h.Button(h.ID("reset-btn"), h.Class("inline-flex items-center rounded-lg bg-rose-500 text-white px-3 py-2 text-sm font-medium hover:bg-rose-600 focus:outline-none focus:ring-2 focus:ring-rose-400"), g.Text("Reset")),
				),
			),

			// Delegate click handlers
			dom.OnClick("inc-btn", onInc),
			dom.OnClick("dec-btn", onDec),
			dom.OnClick("reset-btn", onReset),

			// Router outlet (router renders here)
			h.Main(h.ID("router-outlet"), h.Class("bg-white border rounded-xl shadow p-6 min-h-[200px]")),
		),
	)
}

// Route components
func HomeComponent(props ...any) interface{} {
	return h.Div(
		h.ID("home-page"),
		h.Class("space-y-2"),
		h.H1(h.Class("text-xl font-semibold text-slate-900"), g.Text("Welcome Home")),
		h.P(h.Class("text-slate-600"), g.Text("This is the home page rendered by the router.")),
	)
}

func AboutComponent(props ...any) interface{} {
	return h.Div(
		h.ID("about-page"),
		h.Class("space-y-2"),
		h.H1(h.Class("text-xl font-semibold text-slate-900"), g.Text("About Us")),
		h.P(h.Class("text-slate-600"), g.Text("A simple demo showing AppManager with routing and state.")),
	)
}

func UserComponent(props ...any) interface{} {
	return h.Div(
		h.ID("user-page"),
		h.Class("space-y-2"),
		h.H1(h.Class("text-xl font-semibold text-slate-900"), g.Text("User Profile")),
		h.P(h.Class("text-slate-600"), g.Text("Profile information would appear here.")),
	)
}
