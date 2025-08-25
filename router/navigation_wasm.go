//go:build js && wasm

package router

import (
	g "maragu.dev/gomponents"
	html "maragu.dev/gomponents/html"
)

// A creates a navigation link component for WASM builds that can be used inside
// gomponents component trees. It renders an <a> element with the provided href
// and children. Event binding for client-side navigation will be added later;
// for now, rendering support is sufficient for current tests.
func A(href string, children ...any) g.Node {
	// Convert variadic children to gomponents nodes, allowing simple string values too.
	nodes := make([]g.Node, 0, len(children)+1)
	nodes = append(nodes, html.Href(href))
	for _, ch := range children {
		switch v := ch.(type) {
		case g.Node:
			nodes = append(nodes, v)
		case string:
			nodes = append(nodes, g.Text(v))
		default:
			// Fallback to string representation
			nodes = append(nodes, g.Textf("%v", v))
		}
	}
	return html.A(nodes...)
}
