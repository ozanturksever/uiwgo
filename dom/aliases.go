//go:build js && wasm

package dom

import (
	// Alias the external DOM package to avoid name collision with this package
	domv2 "honnef.co/go/js/dom/v2"
)

// Element is re-exported for consumers to use typed DOM elements in handlers.
// It aliases honnef.co/go/js/dom/v2 Element for safer, statically-typed interop.
type Element = domv2.Element

// Event is re-exported for consumers to use typed events in handlers.
type Event = domv2.Event
