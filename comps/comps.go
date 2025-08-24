//go:build js && wasm

package comps

import (
	"bytes"
	"fmt"
	"syscall/js"

	g "maragu.dev/gomponents"
)

// Node is an alias for gomponents.Node for convenience.
type Node = g.Node

// ComponentFunc defines the signature for components.
type ComponentFunc[P any] func(props P) Node

var (
	// queue of functions to execute after Mount has attached DOM
	mountQueue []func()
)

// Mount renders a root component into a specific DOM element identified by its ID.
// It runs any OnMount functions and attaches reactive binders.
func Mount(elementID string, root func() Node) {
	doc := js.Global().Get("document")
	if doc.IsUndefined() || doc.IsNull() {
		panic("document is not available (not running in a browser)")
	}
	container := doc.Call("getElementById", elementID)
	if container.IsUndefined() || container.IsNull() {
		panic(fmt.Sprintf("Mount: element with id '%s' not found", elementID))
	}

	// Render gomponents Node to HTML string
	var buf bytes.Buffer
	if err := root().Render(&buf); err != nil {
		panic(err)
	}
	container.Set("innerHTML", buf.String())

	// Attach reactive binders (BindText, Show, etc.)
	attachBinders(container)

	// Run queued OnMount callbacks
	for _, f := range mountQueue {
		f()
	}
	mountQueue = nil
}

// enqueueOnMount stores a function to be executed after Mount completes DOM attachment.
func enqueueOnMount(fn func()) { mountQueue = append(mountQueue, fn) }
