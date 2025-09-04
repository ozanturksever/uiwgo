//go:build js && wasm

package main

import (
	"github.com/ozanturksever/uiwgo/bridge"
	"github.com/ozanturksever/uiwgo/logutil"
	"github.com/ozanturksever/uiwgo/router"
	"github.com/ozanturksever/uiwgo/wasm"
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

func main() {
	// Initialize WASM and bridge
	if err := wasm.QuickInit(); err != nil {
		logutil.Logf("Failed to initialize WASM: %v", err)
		return
	}
	
	// Initialize the bridge manager
	bridge.InitializeManager(bridge.NewRealManager())
	
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
	outlet := bridge.GetElementByID("app")
	if outlet == nil {
		logutil.Log("Could not find #app element")
		return
	}

	// Create router with the outlet element
	router.New(routes, outlet)

	// Prevent exit
	select {}
}
