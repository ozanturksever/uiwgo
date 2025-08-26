//go:build js && wasm

package comps

import (
	g "maragu.dev/gomponents"
	h "maragu.dev/gomponents/html"
)

// Example functional components demonstrating the functional-based model

// ExampleComponent is a functional component that takes a title prop
func ExampleComponent(title string) g.Node {
	return h.Div(g.Text("test:"), g.Text(title), OtherComponent())
}

// OtherComponent is a simple functional component
func OtherComponent() g.Node {
	return h.Div(g.Text("from other comp"))
}
