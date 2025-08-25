//go:build js && wasm

package main

import (
	"github.com/ozanturksever/uiwgo/router"
	dom "honnef.co/go/js/dom/v2"
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

func main() {
	// Define simple routes for testing
	routes := []*router.RouteDefinition{
		router.Route("/", func(props ...any) interface{} {
			return Div(Text("Home Page"))
		}),
		router.Route("/test-route", func(props ...any) interface{} {
			return Div(Text("Test Route Page"))
		}),
	}

	// Get the app element to use as outlet
	window := dom.GetWindow()
	doc := window.Document()
	outlet := doc.GetElementByID("app")
	if outlet == nil {
		panic("Could not find #app element")
	}

	// Create router with the outlet element
	router.New(routes, outlet)

	// Prevent exit

	// Prevent exit
	select {}
}
