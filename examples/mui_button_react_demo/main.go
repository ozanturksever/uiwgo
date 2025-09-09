//go:build js && wasm

package main

import (
	"github.com/ozanturksever/logutil"
)

func main() {
	// Minimal WASM entry for the example, keep runtime alive
	logutil.Log("[mui_button_react_demo] WASM initialized")
	select {}
}
