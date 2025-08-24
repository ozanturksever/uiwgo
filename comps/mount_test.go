package comps

import (
	"fmt"
	"syscall/js"
	"testing"
	"time"

	"github.com/ozanturksever/uiwgo/reactivity"
	g "maragu.dev/gomponents"
)

// TestMountDisposer tests that Mount returns a disposer function that properly cleans up
func TestMountDisposer(t *testing.T) {
	// Skip if not in browser environment
	if js.Global().Get("document").IsUndefined() {
		t.Skip("Skipping browser-specific test")
	}

	// Create a test container
	document := js.Global().Get("document")
	container := document.Call("createElement", "div")
	container.Set("id", "test-container")
	document.Get("body").Call("appendChild", container)
	defer document.Get("body").Call("removeChild", container)

	// Create a signal for testing
	counter := reactivity.CreateSignal(0)

	// Mount the component (create it inside the mount function)
	disposer := Mount("test-container", func() Node {
		return g.El("div",
			g.El("p", BindText(func() string {
				return fmt.Sprintf("Count: %d", counter.Get())
			})),
		)
	})

	// Verify the component is mounted
	innerHTML := container.Get("innerHTML").String()
	if innerHTML == "" {
		t.Fatal("Component was not mounted")
	}
	t.Logf("Initial innerHTML: %s", innerHTML)

	// Check if text binder was created
	t.Logf("Text registry size: %d", len(textRegistry))
	for id, binder := range textRegistry {
		t.Logf("Text binder %s: container=%s", id, binder.container)
	}

	// Update the signal to trigger reactive update
	counter.Set(1)
	time.Sleep(10 * time.Millisecond) // Allow time for DOM updates

	// Verify the update occurred
	updatedHTML := container.Get("innerHTML").String()
	t.Logf("Updated innerHTML: %s", updatedHTML)
	if !contains(updatedHTML, "Count: 1") {
		t.Error("Reactive update did not occur")
	}

	// Count registry entries for this specific container
	textEntriesForContainer := 0
	htmlEntriesForContainer := 0
	for _, binder := range textRegistry {
		if binder.container == "test-container" {
			textEntriesForContainer++
		}
	}
	for _, binder := range htmlRegistry {
		if binder.container == "test-container" {
			htmlEntriesForContainer++
		}
	}

	// Call the disposer
	disposer()

	// Verify the container is cleared
	if container.Get("innerHTML").String() != "" {
		t.Error("Container was not cleared after disposal")
	}

	// Verify registry entries for this container are cleaned up
	textEntriesAfter := 0
	htmlEntriesAfter := 0
	for _, binder := range textRegistry {
		if binder.container == "test-container" {
			textEntriesAfter++
		}
	}
	for _, binder := range htmlRegistry {
		if binder.container == "test-container" {
			htmlEntriesAfter++
		}
	}

	if textEntriesAfter != 0 {
		t.Errorf("Text registry was not cleaned up: had %d entries for test-container, still has %d", textEntriesForContainer, textEntriesAfter)
	}
	if htmlEntriesAfter != 0 {
		t.Errorf("HTML registry was not cleaned up: had %d entries for test-container, still has %d", htmlEntriesForContainer, htmlEntriesAfter)
	}

	// Verify no effects are still active by updating the signal
	counter.Set(2)
	time.Sleep(10 * time.Millisecond)

	// Container should still be empty
	if container.Get("innerHTML").String() != "" {
		t.Error("Effects were not properly disposed - container updated after disposal")
	}
}

// TestMountRemountSameContainer tests that mounting to the same container after disposal works correctly
func TestMountRemountSameContainer(t *testing.T) {
	// Skip if not in browser environment
	if js.Global().Get("document").IsUndefined() {
		t.Skip("Skipping browser-specific test")
	}

	// Create a test container
	document := js.Global().Get("document")
	container := document.Call("createElement", "div")
	container.Set("id", "test-remount-container")
	document.Get("body").Call("appendChild", container)
	defer document.Get("body").Call("removeChild", container)

	// First mount
	counter1 := reactivity.CreateSignal(10)

	disposer1 := Mount("test-remount-container", func() Node {
		return g.El("div", BindText(func() string {
			return fmt.Sprintf("First: %d", counter1.Get())
		}))
	})

	// Verify first mount
	if !contains(container.Get("innerHTML").String(), "First: 10") {
		t.Error("First component was not mounted correctly")
	}

	// Dispose first mount
	disposer1()

	// Verify cleanup
	if container.Get("innerHTML").String() != "" {
		t.Error("Container was not cleared after first disposal")
	}

	// Second mount to same container
	counter2 := reactivity.CreateSignal(20)

	disposer2 := Mount("test-remount-container", func() Node {
		return g.El("div", BindText(func() string {
			return fmt.Sprintf("Second: %d", counter2.Get())
		}))
	})

	// Verify second mount
	if !contains(container.Get("innerHTML").String(), "Second: 20") {
		t.Error("Second component was not mounted correctly")
	}

	// Update second signal
	counter2.Set(21)
	time.Sleep(10 * time.Millisecond)

	// Verify second component is reactive
	if !contains(container.Get("innerHTML").String(), "Second: 21") {
		t.Error("Second component is not reactive")
	}

	// Update first signal (should have no effect)
	counter1.Set(11)
	time.Sleep(10 * time.Millisecond)

	// Should still show second component
	if !contains(container.Get("innerHTML").String(), "Second: 21") {
		t.Error("First component effects were not properly disposed")
	}

	// Clean up
	disposer2()
}

// TestShowComponentCleanup tests that Show components are properly cleaned up
func TestShowComponentCleanup(t *testing.T) {
	// Skip if not in browser environment
	if js.Global().Get("document").IsUndefined() {
		t.Skip("Skipping browser-specific test")
	}

	// Create a test container
	document := js.Global().Get("document")
	container := document.Call("createElement", "div")
	container.Set("id", "test-show-container")
	document.Get("body").Call("appendChild", container)
	defer document.Get("body").Call("removeChild", container)

	// Create a signal for show condition
	showSignal := reactivity.CreateSignal(true)

	// Mount the component with Show
	disposer := Mount("test-show-container", func() Node {
		return g.El("div",
			Show(ShowProps{
				When:     showSignal,
				Children: g.El("p", g.Text("Visible content")),
			}),
		)
	})

	// Verify the show component is visible
	if !contains(container.Get("innerHTML").String(), "Visible content") {
		t.Error("Show component content is not visible")
	}

	// Count initial registry entries
	initialShowCount := len(showRegistry)

	// Call the disposer
	disposer()

	// Verify show registry is cleaned up
	if len(showRegistry) >= initialShowCount {
		t.Error("Show registry was not cleaned up")
	}

	// Verify updating the signal has no effect after disposal
	showSignal.Set(false)
	time.Sleep(10 * time.Millisecond)

	// Container should be empty
	if container.Get("innerHTML").String() != "" {
		t.Error("Show component effects were not properly disposed")
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && findSubstring(s, substr)
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		match := true
		for j := 0; j < len(substr); j++ {
			if s[i+j] != substr[j] {
				match = false
				break
			}
		}
		if match {
			return true
		}
	}
	return false
}