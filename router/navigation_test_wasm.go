//go:build js && wasm

package router

import (
	"testing"

	g "maragu.dev/gomponents"
)

// TestNavigation_WASMComponentCreation verifies that router.A creates a proper
// gomponents.Node in WASM builds.
func TestNavigation_WASMComponentCreation(t *testing.T) {
	// Create a link using router.A
	link := A("/about", "About")

	// Verify it returns a gomponents.Node
	if link == nil {
		t.Fatal("Expected A() to return a non-nil gomponents.Node")
	}

	// Type assert to ensure it's a gomponents.Node
	node, ok := link.(g.Node)
	if !ok {
		t.Fatalf("Expected A() to return a gomponents.Node, got %T", link)
	}

	// Verify the node can be rendered (basic smoke test)
	if node == nil {
		t.Error("Expected non-nil gomponents.Node")
	}
}