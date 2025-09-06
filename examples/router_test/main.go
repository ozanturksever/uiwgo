//go:build js && wasm

package main

import (
	"github.com/ozanturksever/logutil"
	"github.com/ozanturksever/uiwgo/router"
	"github.com/ozanturksever/uiwgo/wasm"
	"honnef.co/go/js/dom/v2"
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

func main() {
	// Initialize WASM and bridge
	if err := wasm.QuickInit(); err != nil {
		logutil.Logf("Failed to initialize WASM: %v", err)
		return
	}

	// Define simple routes for testing
	routes := []*router.RouteDefinition{
		router.Route("/", func(props ...any) interface{} {
			return Div(Text("Home Page"), router.A("/test-route", Text("Test Route")))
		}),
		router.Route("/test-route", func(props ...any) interface{} {
			return Div(Text("Test Route Page"))
		}),
	}

	// Get the app element to use as outlet
	outlet := dom.GetWindow().Document().GetElementByID("app")
	if outlet == nil {
		logutil.Log("Could not find #app element")
		return
	}

	// Create router with the outlet element
	router.New(routes, outlet)

	// Prevent exit
	select {}
}
