package main

import (
	"fmt"
	"syscall/js"
	"time"

	"github.com/ozanturksever/uiwgo/comps"
	"github.com/ozanturksever/uiwgo/reactivity"
	g "maragu.dev/gomponents"
)

type TestItem struct {
	ID   string
	Name string
}

func main() {
	// Create a test container
	doc := js.Global().Get("document")
	container := doc.Call("createElement", "div")
	container.Set("id", "test-container")
	doc.Get("body").Call("appendChild", container)

	// Create reactive signals
	itemsSignal := reactivity.CreateSignal([]TestItem{
		{ID: "reactive1", Name: "Reactive Item 1"},
	})

	showSignal := reactivity.CreateSignal(true)

	fmt.Println("Starting test...")

	// Mount the component
	disposer := comps.Mount(container.Get("id").String(), func() g.Node {
		fmt.Println("Rendering For component...")
		return comps.For(comps.ForProps[TestItem]{
			Items: itemsSignal,
			Key:   func(item TestItem) string { return item.ID },
			Children: func(item TestItem, index int) g.Node {
				fmt.Printf("Rendering child for item: %s\n", item.Name)
				return comps.Show(comps.ShowProps{
					When: showSignal,
					Children: g.El("div",
						g.Attr("class", "reactive-item"),
						g.Text(item.Name),
					),
				})
			},
		})
	})
	defer disposer()

	// Allow time for DOM updates
	time.Sleep(50 * time.Millisecond)

	// Check initial state
	reactiveItems := container.Call("querySelectorAll", ".reactive-item")
	fmt.Printf("Initial reactive items: %d\n", reactiveItems.Get("length").Int())
	fmt.Printf("Container innerHTML: %s\n", container.Get("innerHTML").String())

	// Hide items
	fmt.Println("Hiding items...")
	showSignal.Set(false)
	time.Sleep(50 * time.Millisecond)

	hiddenItems := container.Call("querySelectorAll", ".reactive-item")
	fmt.Printf("Hidden reactive items: %d\n", hiddenItems.Get("length").Int())
	fmt.Printf("Container innerHTML after hide: %s\n", container.Get("innerHTML").String())

	// Show items again
	fmt.Println("Showing items again...")
	showSignal.Set(true)
	time.Sleep(50 * time.Millisecond)

	visibleItems := container.Call("querySelectorAll", ".reactive-item")
	fmt.Printf("Visible reactive items: %d\n", visibleItems.Get("length").Int())
	fmt.Printf("Container innerHTML after show: %s\n", container.Get("innerHTML").String())

	fmt.Println("Test completed.")
}
