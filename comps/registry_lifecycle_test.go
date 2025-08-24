//go:build js && wasm

package comps

import (
	"syscall/js"
	"testing"
	"time"

	"github.com/ozanturksever/uiwgo/reactivity"
	g "maragu.dev/gomponents"
)

// TestRegistryLifecycleAndMemory tests that registry entries are properly
// linked to scopes and cleaned up on disposal to prevent unbounded growth
func TestRegistryLifecycleAndMemory(t *testing.T) {
	// Skip if not in browser environment
	if js.Global().Get("document").IsUndefined() {
		t.Skip("Skipping browser-specific test")
	}

	// Create test container
	document := js.Global().Get("document")
	container := document.Call("createElement", "div")
	container.Set("id", "test-registry-lifecycle")
	document.Get("body").Call("appendChild", container)
	defer document.Get("body").Call("removeChild", container)

	// Record initial registry sizes
	initialTextSize := len(textRegistry)
	initialHTMLSize := len(htmlRegistry)
	initialShowSize := len(showRegistry)
	initialForSize := len(forRegistry)
	initialIndexSize := len(indexRegistry)
	initialSwitchSize := len(switchRegistry)
	initialDynamicSize := len(dynamicRegistry)

	// Create signals for testing
	textSignal := reactivity.CreateSignal("initial")
	showSignal := reactivity.CreateSignal(true)
	items := reactivity.CreateSignal([]string{"item1", "item2"})
	switchSignal := reactivity.CreateSignal("case1")

	// Mount component with various binders
	disposer := Mount("test-registry-lifecycle", func() Node {
		return g.El("div",
			// Text binder
			BindText(func() string {
				return "Text: " + textSignal.Get()
			}),
			// HTML binder
			BindHTML(func() g.Node {
				return g.El("span", g.Text("HTML: "+textSignal.Get()))
			}),
			// Show binder
			Show(ShowProps{
				When:     showSignal,
				Children: g.El("p", g.Text("Shown content")),
			}),
			// For binder
			For(ForProps[string]{
				Items: items,
				Key:   func(item string) string { return item },
				Children: func(item string, index int) g.Node {
					return g.El("li", g.Text(item))
				},
			}),
			// Index binder
			Index(IndexProps[string]{
				Items: items,
				Children: func(getItem func() string, index int) g.Node {
					return g.El("div", g.Text(getItem()))
				},
			}),
			// Switch binder
			Switch(SwitchProps{
				When: switchSignal,
				Children: []g.Node{
					Match(MatchProps{
						When:     "case1",
						Children: g.El("div", g.Text("Case 1")),
					}),
					Match(MatchProps{
						When:     "case2",
						Children: g.El("div", g.Text("Case 2")),
					}),
				},
				Fallback: g.El("div", g.Text("Default")),
			}),
			// Dynamic binder
			Dynamic(DynamicProps{
				Component: func() g.Node {
					return g.El("span", g.Text("Dynamic: "+textSignal.Get()))
				},
			}),
		)
	})

	// Verify registries have grown
	if len(textRegistry) <= initialTextSize {
		t.Error("Text registry should have grown after mount")
	}
	if len(htmlRegistry) <= initialHTMLSize {
		t.Error("HTML registry should have grown after mount")
	}
	if len(showRegistry) <= initialShowSize {
		t.Error("Show registry should have grown after mount")
	}
	if len(forRegistry) <= initialForSize {
		t.Error("For registry should have grown after mount")
	}
	if len(indexRegistry) <= initialIndexSize {
		t.Error("Index registry should have grown after mount")
	}
	if len(switchRegistry) <= initialSwitchSize {
		t.Error("Switch registry should have grown after mount")
	}
	if len(dynamicRegistry) <= initialDynamicSize {
		t.Error("Dynamic registry should have grown after mount")
	}

	// Trigger some updates to ensure binders are active
	textSignal.Set("updated")
	showSignal.Set(false)
	items.Set([]string{"item1", "item2", "item3"})
	switchSignal.Set("case2")
	time.Sleep(10 * time.Millisecond)

	// Verify all registry entries are for the correct container
	for id, binder := range textRegistry {
		if binder.container == "test-registry-lifecycle" {
			t.Logf("Text binder %s is correctly associated with container", id)
		}
	}
	for id, binder := range htmlRegistry {
		if binder.container == "test-registry-lifecycle" {
			t.Logf("HTML binder %s is correctly associated with container", id)
		}
	}
	for id, binder := range forRegistry {
		if binder.mountContainer == "test-registry-lifecycle" {
			t.Logf("For binder %s is correctly associated with container", id)
		}
	}

	// Call disposer to clean up
	disposer()

	// Verify registries are back to initial sizes
	if len(textRegistry) != initialTextSize {
		t.Errorf("Text registry not cleaned up: expected %d, got %d", initialTextSize, len(textRegistry))
	}
	if len(htmlRegistry) != initialHTMLSize {
		t.Errorf("HTML registry not cleaned up: expected %d, got %d", initialHTMLSize, len(htmlRegistry))
	}
	if len(showRegistry) != initialShowSize {
		t.Errorf("Show registry not cleaned up: expected %d, got %d", initialShowSize, len(showRegistry))
	}
	if len(forRegistry) != initialForSize {
		t.Errorf("For registry not cleaned up: expected %d, got %d", initialForSize, len(forRegistry))
	}
	if len(indexRegistry) != initialIndexSize {
		t.Errorf("Index registry not cleaned up: expected %d, got %d", initialIndexSize, len(indexRegistry))
	}
	if len(switchRegistry) != initialSwitchSize {
		t.Errorf("Switch registry not cleaned up: expected %d, got %d", initialSwitchSize, len(switchRegistry))
	}
	if len(dynamicRegistry) != initialDynamicSize {
		t.Errorf("Dynamic registry not cleaned up: expected %d, got %d", initialDynamicSize, len(dynamicRegistry))
	}

	// Verify no entries remain for the test container
	for id, binder := range textRegistry {
		if binder.container == "test-registry-lifecycle" {
			t.Errorf("Text binder %s still exists for disposed container", id)
		}
	}
	for id, binder := range htmlRegistry {
		if binder.container == "test-registry-lifecycle" {
			t.Errorf("HTML binder %s still exists for disposed container", id)
		}
	}
	for id, binder := range forRegistry {
		if binder.mountContainer == "test-registry-lifecycle" {
			t.Errorf("For binder %s still exists for disposed container", id)
		}
	}
}

// TestRepeatedMountUnmountCycles tests that repeated mount/unmount cycles
// do not cause unbounded registry growth
func TestRepeatedMountUnmountCycles(t *testing.T) {
	// Skip if not in browser environment
	if js.Global().Get("document").IsUndefined() {
		t.Skip("Skipping browser-specific test")
	}

	// Create test container
	document := js.Global().Get("document")
	container := document.Call("createElement", "div")
	container.Set("id", "test-repeated-cycles")
	document.Get("body").Call("appendChild", container)
	defer document.Get("body").Call("removeChild", container)

	// Record initial registry sizes
	initialTextSize := len(textRegistry)
	initialHTMLSize := len(htmlRegistry)
	initialShowSize := len(showRegistry)
	initialForSize := len(forRegistry)

	// Perform multiple mount/unmount cycles
	const cycles = 10
	for i := 0; i < cycles; i++ {
		// Create fresh signals for each cycle
		textSignal := reactivity.CreateSignal("cycle")
		showSignal := reactivity.CreateSignal(true)
		items := reactivity.CreateSignal([]string{"a", "b"})

		// Mount component
		disposer := Mount("test-repeated-cycles", func() Node {
			return g.El("div",
				BindText(func() string {
					return textSignal.Get()
				}),
				BindHTML(func() g.Node {
					return g.El("span", g.Text(textSignal.Get()))
				}),
				Show(ShowProps{
					When:     showSignal,
					Children: g.El("p", g.Text("content")),
				}),
				For(ForProps[string]{
					Items: items,
					Key:   func(item string) string { return item },
					Children: func(item string, index int) g.Node {
						return g.El("li", g.Text(item))
					},
				}),
			)
		})

		// Trigger some updates
		textSignal.Set("updated")
		showSignal.Set(false)
		items.Set([]string{"a", "b", "c"})
		time.Sleep(5 * time.Millisecond)

		// Unmount
		disposer()

		// Check registry sizes haven't grown unbounded
		if len(textRegistry) > initialTextSize+5 {
			t.Errorf("Cycle %d: Text registry growing unbounded: %d entries", i, len(textRegistry))
		}
		if len(htmlRegistry) > initialHTMLSize+5 {
			t.Errorf("Cycle %d: HTML registry growing unbounded: %d entries", i, len(htmlRegistry))
		}
		if len(showRegistry) > initialShowSize+5 {
			t.Errorf("Cycle %d: Show registry growing unbounded: %d entries", i, len(showRegistry))
		}
		if len(forRegistry) > initialForSize+5 {
			t.Errorf("Cycle %d: For registry growing unbounded: %d entries", i, len(forRegistry))
		}
	}

	// Final verification: registries should be back to initial sizes
	if len(textRegistry) != initialTextSize {
		t.Errorf("Final text registry size: expected %d, got %d", initialTextSize, len(textRegistry))
	}
	if len(htmlRegistry) != initialHTMLSize {
		t.Errorf("Final HTML registry size: expected %d, got %d", initialHTMLSize, len(htmlRegistry))
	}
	if len(showRegistry) != initialShowSize {
		t.Errorf("Final show registry size: expected %d, got %d", initialShowSize, len(showRegistry))
	}
	if len(forRegistry) != initialForSize {
		t.Errorf("Final for registry size: expected %d, got %d", initialForSize, len(forRegistry))
	}
}

// TestScopeBasedRegistryCleanup tests that registry entries are properly
// linked to scopes and cleaned up when scopes are disposed
func TestScopeBasedRegistryCleanup(t *testing.T) {
	// Skip if not in browser environment
	if js.Global().Get("document").IsUndefined() {
		t.Skip("Skipping browser-specific test")
	}

	// Create test container
	document := js.Global().Get("document")
	container := document.Call("createElement", "div")
	container.Set("id", "test-scope-cleanup")
	document.Get("body").Call("appendChild", container)
	defer document.Get("body").Call("removeChild", container)

	// Create a signal
	textSignal := reactivity.CreateSignal("test")

	// Mount component
	disposer := Mount("test-scope-cleanup", func() Node {
		return g.El("div",
			BindText(func() string {
				return textSignal.Get()
			}),
		)
	})

	// Find the binder entry for our container
	var binderID string
	var binder textBinder
	for id, b := range textRegistry {
		if b.container == "test-scope-cleanup" {
			binderID = id
			binder = b
			break
		}
	}

	if binderID == "" {
		t.Fatal("Could not find binder for test container")
	}

	// Verify the binder has an effect
	if binder.effect == nil {
		t.Error("Binder should have an effect")
	}

	// Verify the effect is active by updating the signal
	textSignal.Set("updated")
	time.Sleep(10 * time.Millisecond)

	// Verify the update worked
	if !stringContains(container.Get("innerHTML").String(), "updated") {
		t.Error("Effect should be active and update DOM")
	}

	// Call disposer
	disposer()

	// Verify the binder is removed from registry
	if _, exists := textRegistry[binderID]; exists {
		t.Error("Binder should be removed from registry after disposal")
	}

	// Verify the effect is disposed by trying to update signal
	textSignal.Set("should-not-update")
	time.Sleep(10 * time.Millisecond)

	// Container should be empty
	if container.Get("innerHTML").String() != "" {
		t.Error("Container should be empty after disposal")
	}
}

// Helper function to check if a string contains a substring
func stringContains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || 
		(len(s) > len(substr) && 
			(s[:len(substr)] == substr || 
			 s[len(s)-len(substr):] == substr || 
			 indexOfSubstring(s, substr) >= 0)))
}

func indexOfSubstring(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}