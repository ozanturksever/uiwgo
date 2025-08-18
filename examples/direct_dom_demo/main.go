// main.go
// Example demonstrating direct DOM manipulation with fine-grained reactivity

//go:build js && wasm

package main

import (
	"fmt"
	"strconv"
	"syscall/js"
	"time"

	"app/golid"

	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

func main() {
	// Initialize the reactive system
	golid.Run()
}

func init() {
	// Set up the demo when the page loads
	js.Global().Set("addEventListener", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		if len(args) > 0 && args[0].Get("type").String() == "DOMContentLoaded" {
			setupDirectDOMDemo()
		}
		return nil
	}))

	// Also set up immediately if DOM is already loaded
	if js.Global().Get("document").Get("readyState").String() == "complete" {
		setupDirectDOMDemo()
	}
}

func setupDirectDOMDemo() {
	// Create root for scoped reactivity
	_, cleanup := golid.CreateRoot(func() interface{} {
		setupCounterDemo()
		setupFormDemo()
		setupListDemo()
		setupPerformanceDemo()
		return nil
	})

	// Register cleanup (in a real app, you'd call this on page unload)
	_ = cleanup
}

// ------------------------------------
// 🔢 Counter Demo - Basic Reactivity
// ------------------------------------

func setupCounterDemo() {
	// Create reactive signals
	count, setCount := golid.CreateSignal(0)

	// Render initial HTML
	counterHTML := Div(
		H2(Text("Direct DOM Counter Demo")),
		P(Text("Count: "), Span(ID("count-display"))),
		Button(ID("increment-btn"), Text("Increment")),
		Button(ID("decrement-btn"), Text("Decrement")),
		Button(ID("reset-btn"), Text("Reset")),
	)

	golid.RenderTo(counterHTML, js.Global().Get("document").Get("body"))

	// Bind reactive text to count display
	countDisplay := js.Global().Get("document").Call("getElementById", "count-display")
	golid.BindTextReactive(countDisplay, func() string {
		return fmt.Sprintf("%d", count())
	})

	// Bind event handlers
	incrementBtn := js.Global().Get("document").Call("getElementById", "increment-btn")
	golid.BindEventReactive(incrementBtn, "click", func(event js.Value) {
		setCount(count() + 1)
	})

	decrementBtn := js.Global().Get("document").Call("getElementById", "decrement-btn")
	golid.BindEventReactive(decrementBtn, "click", func(event js.Value) {
		setCount(count() - 1)
	})

	resetBtn := js.Global().Get("document").Call("getElementById", "reset-btn")
	golid.BindEventReactive(resetBtn, "click", func(event js.Value) {
		setCount(0)
	})

	// Bind reactive class for styling
	golid.BindClassReactive(countDisplay, "positive", func() bool {
		return count() > 0
	})

	golid.BindClassReactive(countDisplay, "negative", func() bool {
		return count() < 0
	})
}

// ------------------------------------
// 📝 Form Demo - Two-Way Binding
// ------------------------------------

func setupFormDemo() {
	// Create reactive signals for form state
	name, setName := golid.CreateSignal("")
	email, setEmail := golid.CreateSignal("")
	isValid := golid.CreateMemo(func() bool {
		return len(name()) > 0 && len(email()) > 3
	}, nil)

	// Render form HTML
	formHTML := Div(
		H2(Text("Reactive Form Demo")),
		Div(
			Label(Text("Name: ")),
			Input(ID("name-input"), Type("text"), Placeholder("Enter your name")),
		),
		Div(
			Label(Text("Email: ")),
			Input(ID("email-input"), Type("email"), Placeholder("Enter your email")),
		),
		Button(ID("submit-btn"), Text("Submit")),
		Div(ID("form-output")),
	)

	golid.RenderTo(formHTML, js.Global().Get("document").Get("body"))

	// Set up two-way binding for name input
	nameInput := js.Global().Get("document").Call("getElementById", "name-input")
	golid.BindFormInput(nameInput, name, setName)

	// Set up two-way binding for email input
	emailInput := js.Global().Get("document").Call("getElementById", "email-input")
	golid.BindFormInput(emailInput, email, setEmail)

	// Bind submit button state
	submitBtn := js.Global().Get("document").Call("getElementById", "submit-btn")
	golid.BindAttributeReactive(submitBtn, "disabled", func() string {
		if isValid() {
			return ""
		}
		return "disabled"
	})

	// Bind form output
	formOutput := js.Global().Get("document").Call("getElementById", "form-output")
	golid.BindTextReactive(formOutput, func() string {
		if isValid() {
			return fmt.Sprintf("Hello %s! Your email is %s", name(), email())
		}
		return "Please fill in all fields"
	})

	// Handle form submission
	golid.BindEventReactive(submitBtn, "click", func(event js.Value) {
		if isValid() {
			js.Global().Call("alert", fmt.Sprintf("Form submitted!\nName: %s\nEmail: %s", name(), email()))
		}
	})
}

// ------------------------------------
// 📋 List Demo - Dynamic Lists
// ------------------------------------

func setupListDemo() {
	// Create reactive signals for list state
	items, setItems := golid.CreateSignal([]string{"Item 1", "Item 2", "Item 3"})
	newItem, setNewItem := golid.CreateSignal("")

	// Render list HTML
	listHTML := Div(
		H2(Text("Reactive List Demo")),
		Div(
			Input(ID("new-item-input"), Type("text"), Placeholder("Enter new item")),
			Button(ID("add-item-btn"), Text("Add Item")),
		),
		Ul(ID("items-list")),
		P(ID("item-count")),
	)

	golid.RenderTo(listHTML, js.Global().Get("document").Get("body"))

	// Set up input binding
	newItemInput := js.Global().Get("document").Call("getElementById", "new-item-input")
	golid.BindFormInput(newItemInput, newItem, setNewItem)

	// Set up list rendering
	itemsList := js.Global().Get("document").Call("getElementById", "items-list")
	golid.ListRender(itemsList, items, func(item string) string {
		return item // Use item as key
	}, func(item string) js.Value {
		li := js.Global().Get("document").Call("createElement", "li")
		li.Set("textContent", item)

		// Add delete button
		deleteBtn := js.Global().Get("document").Call("createElement", "button")
		deleteBtn.Set("textContent", "Delete")
		deleteBtn.Call("addEventListener", "click", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			currentItems := items()
			newItems := make([]string, 0, len(currentItems))
			for _, i := range currentItems {
				if i != item {
					newItems = append(newItems, i)
				}
			}
			setItems(newItems)
			return nil
		}))

		li.Call("appendChild", deleteBtn)
		return li
	})

	// Bind item count
	itemCount := js.Global().Get("document").Call("getElementById", "item-count")
	golid.BindTextReactive(itemCount, func() string {
		return fmt.Sprintf("Total items: %d", len(items()))
	})

	// Handle adding new items
	addBtn := js.Global().Get("document").Call("getElementById", "add-item-btn")
	golid.BindEventReactive(addBtn, "click", func(event js.Value) {
		if newItem() != "" {
			currentItems := items()
			newItems := append(currentItems, newItem())
			setItems(newItems)
			setNewItem("")
		}
	})

	// Handle Enter key in input
	golid.BindEventReactive(newItemInput, "keypress", func(event js.Value) {
		if event.Get("key").String() == "Enter" {
			if newItem() != "" {
				currentItems := items()
				newItems := append(currentItems, newItem())
				setItems(newItems)
				setNewItem("")
			}
		}
	})
}

// ------------------------------------
// ⚡ Performance Demo - Stress Test
// ------------------------------------

func setupPerformanceDemo() {
	// Create signals for performance testing
	updateCount, setUpdateCount := golid.CreateSignal(0)
	isRunning, setIsRunning := golid.CreateSignal(false)

	// Render performance demo HTML
	perfHTML := Div(
		H2(Text("Performance Demo - Direct DOM vs Virtual DOM")),
		P(Text("This demo shows the performance benefits of direct DOM manipulation.")),
		Div(
			Button(ID("start-perf-btn"), Text("Start Performance Test")),
			Button(ID("stop-perf-btn"), Text("Stop Test")),
		),
		Div(ID("perf-stats")),
		Div(ID("perf-grid")),
	)

	golid.RenderTo(perfHTML, js.Global().Get("document").Get("body"))

	// Bind performance stats
	perfStats := js.Global().Get("document").Call("getElementById", "perf-stats")
	golid.BindTextReactive(perfStats, func() string {
		status := "Stopped"
		if isRunning() {
			status = "Running"
		}
		return fmt.Sprintf("Status: %s | Updates: %d", status, updateCount())
	})

	// Create performance grid
	perfGrid := js.Global().Get("document").Call("getElementById", "perf-grid")
	perfGrid.Get("style").Set("display", "grid")
	perfGrid.Get("style").Set("grid-template-columns", "repeat(10, 1fr)")
	perfGrid.Get("style").Set("gap", "2px")
	perfGrid.Get("style").Set("max-width", "500px")

	// Create 100 reactive elements for stress testing
	for i := 0; i < 100; i++ {
		cellId := fmt.Sprintf("perf-cell-%d", i)
		cell := js.Global().Get("document").Call("createElement", "div")
		cell.Set("id", cellId)
		cell.Get("style").Set("width", "40px")
		cell.Get("style").Set("height", "40px")
		cell.Get("style").Set("border", "1px solid #ccc")
		cell.Get("style").Set("display", "flex")
		cell.Get("style").Set("align-items", "center")
		cell.Get("style").Set("justify-content", "center")
		perfGrid.Call("appendChild", cell)

		// Bind reactive content to each cell
		cellIndex := i
		golid.BindTextReactive(cell, func() string {
			return strconv.Itoa((updateCount() + cellIndex) % 100)
		})

		// Bind reactive background color
		golid.BindStyleReactive(cell, "background-color", func() string {
			value := (updateCount() + cellIndex) % 100
			if value < 33 {
				return "#ffebee"
			} else if value < 66 {
				return "#e8f5e8"
			}
			return "#e3f2fd"
		})
	}

	// Performance test controls
	startBtn := js.Global().Get("document").Call("getElementById", "start-perf-btn")
	stopBtn := js.Global().Get("document").Call("getElementById", "stop-perf-btn")

	var ticker *time.Ticker
	var done chan bool

	golid.BindEventReactive(startBtn, "click", func(event js.Value) {
		if !isRunning() {
			setIsRunning(true)
			ticker = time.NewTicker(16 * time.Millisecond) // ~60fps
			done = make(chan bool)

			go func() {
				for {
					select {
					case <-ticker.C:
						setUpdateCount(updateCount() + 1)
					case <-done:
						return
					}
				}
			}()
		}
	})

	golid.BindEventReactive(stopBtn, "click", func(event js.Value) {
		if isRunning() {
			setIsRunning(false)
			if ticker != nil {
				ticker.Stop()
			}
			if done != nil {
				close(done)
			}
		}
	})

	// Bind button states
	golid.BindAttributeReactive(startBtn, "disabled", func() string {
		if isRunning() {
			return "disabled"
		}
		return ""
	})

	golid.BindAttributeReactive(stopBtn, "disabled", func() string {
		if !isRunning() {
			return "disabled"
		}
		return ""
	})
}
