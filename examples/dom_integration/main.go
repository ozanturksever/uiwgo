//go:build js && wasm

package main

import (
	"fmt"
	"strconv"
	"syscall/js"
	"time"

	"github.com/ozanturksever/uiwgo/comps"
	"github.com/ozanturksever/uiwgo/dom"
	"github.com/ozanturksever/uiwgo/reactivity"
	domv2 "honnef.co/go/js/dom/v2"
	"maragu.dev/gomponents"
	"maragu.dev/gomponents/html"
)

//func main() {
//	// Wait for DOM to be ready
//	js.Global().Get("document").Call("addEventListener", "DOMContentLoaded", js.FuncOf(func(this js.Value, args []js.Value) any {
//		fmt.Println("DOM is ready!")
//		runApp()
//		return nil
//	}))
//
//	// Keep the program running
//	select {}
//}

func main() {
	runApp()

	select {}
}
func runApp() {
	// Create reactive signals
	counter := reactivity.CreateSignal(0)
	name := reactivity.CreateSignal("World")
	isVisible := reactivity.CreateSignal(true)
	todos := reactivity.CreateSignal([]string{"Learn Go", "Build WASM app", "Use dom/v2"})
	newTodo := reactivity.CreateSignal("")
	
	// Control flow examples
	items := reactivity.CreateSignal([]string{"Apple", "Banana", "Cherry"})
	selectedTab := reactivity.CreateSignal("home")
	currentComponent := reactivity.CreateSignal[func() gomponents.Node](nil)

	// Create the main app component using traditional gomponents
	app := html.Div(
		html.Class("container"),
		html.H1(gomponents.Text("DOM/v2 Integration Demo")),

		// Counter section
		html.Div(
			html.Class("section"),
			html.H2(gomponents.Text("Counter Example")),
			html.P(
				gomponents.Text("Count: "),
				comps.BindText(func() string {
					return strconv.Itoa(counter.Get())
				}),
			),
			html.Button(
				html.ID("increment-btn"),
				gomponents.Text("Increment"),
			),
			html.Button(
				html.ID("decrement-btn"),
				gomponents.Text("Decrement"),
			),
			html.Button(
				html.ID("reset-btn"),
				gomponents.Text("Reset"),
			),
		),

		// Name input section
		html.Div(
			html.Class("section"),
			html.H2(gomponents.Text("Name Input Example")),
			html.P(
				gomponents.Text("Hello, "),
				comps.BindText(func() string {
					return name.Get()
				}),
				gomponents.Text("!"),
			),
			html.Input(
				html.ID("name-input"),
				html.Type("text"),
				html.Placeholder("Enter your name"),
				html.Value(name.Get()),
			),
		),

		// Visibility toggle section
		html.Div(
			html.Class("section"),
			html.H2(gomponents.Text("Visibility Toggle Example")),
			html.Button(
				html.ID("toggle-btn"),
				gomponents.Text("Toggle Visibility"),
			),
			comps.Show(comps.ShowProps{
				When: isVisible,
				Children: html.P(
					html.Style("color: green; font-weight: bold;"),
					gomponents.Text("This text can be toggled!"),
				),
			}),
		),

		// Todo list section
		html.Div(
			html.Class("section"),
			html.H2(gomponents.Text("Todo List Example")),
			html.Div(
				html.Input(
					html.ID("todo-input"),
					html.Type("text"),
					html.Placeholder("Enter a new todo"),
					html.Value(newTodo.Get()),
				),
				html.Button(
					html.ID("add-todo-btn"),
					gomponents.Text("Add Todo"),
				),
			),
			html.Ul(
				html.ID("todo-list"),
				comps.BindHTML(func() gomponents.Node {
					todoItems := todos.Get()
					var items []gomponents.Node
					for i, todo := range todoItems {
						items = append(items, html.Li(
							html.DataAttr("index", strconv.Itoa(i)),
							html.Span(gomponents.Text(todo)),
							html.Button(
								html.Class("delete-todo"),
								html.DataAttr("index", strconv.Itoa(i)),
								gomponents.Text("Delete"),
							),
						))
					}
					return gomponents.Group(items)
				}),
			),
		),

		// Dynamic element creation section
		html.Div(
			html.Class("section"),
			html.H2(gomponents.Text("Dynamic Element Creation")),
			html.Button(
				html.ID("create-element-btn"),
				gomponents.Text("Create Dynamic Element"),
			),
			html.Div(
				html.ID("dynamic-container"),
			),
		),

		// Control Flow Components Examples
		html.Div(
			html.Class("section"),
			html.H2(gomponents.Text("Control Flow Components")),
			
			// For component example
			html.H3(gomponents.Text("For Component (Keyed List)")),
			html.P(gomponents.Text("Efficiently renders lists with keyed reconciliation:")),
			comps.For(comps.ForProps[string]{
				Items: items,
				Key: func(item string) string {
				return item // Use item as key
			},
				Children: func(item string, index int) gomponents.Node {
					return html.Div(
						html.Class("list-item"),
						html.Span(comps.BindText(func() string {
						return fmt.Sprintf("%d: %s", index, item)
					})),
						html.Button(
							html.Class("remove-item"),
							html.DataAttr("index", strconv.Itoa(index)),
							gomponents.Text("Remove"),
						),
					)
				},
			}),
			html.Button(
				html.ID("add-item-btn"),
				gomponents.Text("Add Item"),
			),
			html.Button(
				html.ID("shuffle-items-btn"),
				gomponents.Text("Shuffle Items"),
			),
			
			// Index component example
			html.H3(gomponents.Text("Index Component (Index-based)")),
			html.P(gomponents.Text("Renders lists with index-based reconciliation:")),
			comps.Index(comps.IndexProps[string]{
				Items: items,
				Children: func(getItem func() string, index int) gomponents.Node {
					return html.Div(
					html.Class("index-item"),
					comps.BindText(func() string {
						return fmt.Sprintf("Index %d: %s", index, getItem())
					}),
				)
				},
			}),
			
			// Switch/Match component example
			html.H3(gomponents.Text("Switch/Match Component")),
			html.P(gomponents.Text("Conditionally renders content based on a value:")),
			html.Div(
				html.Button(
					html.ID("tab-home"),
					gomponents.Text("Home"),
				),
				html.Button(
					html.ID("tab-about"),
					gomponents.Text("About"),
				),
				html.Button(
					html.ID("tab-contact"),
					gomponents.Text("Contact"),
				),
			),
			comps.Switch(comps.SwitchProps{
				When: selectedTab,
				Fallback: html.P(gomponents.Text("Unknown tab selected")),
				Children: []gomponents.Node{
					comps.Match(comps.MatchProps{
						When: "home",
						Children: html.Div(
							html.H4(gomponents.Text("Home Page")),
							html.P(gomponents.Text("Welcome to the home page!")),
						),
					}),
					comps.Match(comps.MatchProps{
						When: "about",
						Children: html.Div(
							html.H4(gomponents.Text("About Page")),
							html.P(gomponents.Text("Learn more about us here.")),
						),
					}),
					comps.Match(comps.MatchProps{
						When: "contact",
						Children: html.Div(
							html.H4(gomponents.Text("Contact Page")),
							html.P(gomponents.Text("Get in touch with us!")),
						),
					}),
				},
			}),
			
			// Dynamic component example
			html.H3(gomponents.Text("Dynamic Component")),
			html.P(gomponents.Text("Dynamically renders different components:")),
			html.Div(
				html.Button(
					html.ID("load-hello-comp"),
					gomponents.Text("Load Hello Component"),
				),
				html.Button(
					html.ID("load-counter-comp"),
					gomponents.Text("Load Counter Component"),
				),
				html.Button(
					html.ID("clear-comp"),
					gomponents.Text("Clear Component"),
				),
			),
			comps.Dynamic(comps.DynamicProps{
				Component: currentComponent,
			}),
		),
	)

	// Mount the app using traditional comps.Mount
	comps.Mount("app", func() comps.Node { return app })

	// Now enhance with dom/v2 event bindings
	enhanceWithDOMv2(counter, name, isVisible, todos, newTodo, items, selectedTab, currentComponent)

	// Expose some functions to global scope for testing
	exposeGlobalFunctions(counter, name, isVisible, todos, newTodo, items, selectedTab, currentComponent)
}

func enhanceWithDOMv2(counter reactivity.Signal[int], name reactivity.Signal[string], isVisible reactivity.Signal[bool], todos reactivity.Signal[[]string], newTodo reactivity.Signal[string], items reactivity.Signal[[]string], selectedTab reactivity.Signal[string], currentComponent reactivity.Signal[func() gomponents.Node]) {
	doc := domv2.GetWindow().Document()

	// Counter button events
	if incrementBtn := doc.GetElementByID("increment-btn"); incrementBtn != nil {
		dom.BindClickToCallback(incrementBtn, func() {
			counter.Set(counter.Get() + 1)
		})
	}

	if decrementBtn := doc.GetElementByID("decrement-btn"); decrementBtn != nil {
		dom.BindClickToCallback(decrementBtn, func() {
			counter.Set(counter.Get() - 1)
		})
	}

	if resetBtn := doc.GetElementByID("reset-btn"); resetBtn != nil {
		dom.BindClickToSignal(resetBtn, counter, 0)
	}

	// Name input event
	if nameInput := doc.GetElementByID("name-input"); nameInput != nil {
		dom.BindInputToSignal(nameInput, name)
	}

	// Visibility toggle event
	if toggleBtn := doc.GetElementByID("toggle-btn"); toggleBtn != nil {
		dom.BindClickToCallback(toggleBtn, func() {
			isVisible.Set(!isVisible.Get())
		})
	}

	// Todo input events
	if todoInput := doc.GetElementByID("todo-input"); todoInput != nil {
		dom.BindInputToSignal(todoInput, newTodo)

		// Add todo on Enter key
		dom.BindEnterKeyToCallback(todoInput, func() {
			addTodo(todos, newTodo)
		})
	}

	// Add todo button event
	if addTodoBtn := doc.GetElementByID("add-todo-btn"); addTodoBtn != nil {
		dom.BindClickToCallback(addTodoBtn, func() {
			addTodo(todos, newTodo)
		})
	}

	// Todo list delegation for delete buttons
	if todoList := doc.GetElementByID("todo-list"); todoList != nil {
		dom.DelegateEvent(todoList, "click", ".delete-todo", func(event domv2.Event, target domv2.Element) {
			if indexStr := target.GetAttribute("data-index"); indexStr != "" {
				if index, err := strconv.Atoi(indexStr); err == nil {
					deleteTodo(todos, index)
				}
			}
		})
	}

	// Dynamic element creation
	if createBtn := doc.GetElementByID("create-element-btn"); createBtn != nil {
		dom.BindClickToCallback(createBtn, func() {
			createDynamicElement()
		})
	}

	// Control flow examples event handlers
	
	// Add item button
	if addItemBtn := doc.GetElementByID("add-item-btn"); addItemBtn != nil {
		dom.BindClickToCallback(addItemBtn, func() {
			currentItems := items.Get()
			newItem := fmt.Sprintf("Item %d", len(currentItems)+1)
			updatedItems := append(currentItems, newItem)
			items.Set(updatedItems)
		})
	}

	// Shuffle items button
	if shuffleBtn := doc.GetElementByID("shuffle-items-btn"); shuffleBtn != nil {
		dom.BindClickToCallback(shuffleBtn, func() {
			currentItems := items.Get()
			shuffledItems := make([]string, len(currentItems))
			copy(shuffledItems, currentItems)
			// Simple shuffle algorithm
			for i := len(shuffledItems) - 1; i > 0; i-- {
				j := int(js.Global().Get("Math").Call("random").Float() * float64(i+1))
				shuffledItems[i], shuffledItems[j] = shuffledItems[j], shuffledItems[i]
			}
			items.Set(shuffledItems)
		})
	}

	// Remove item button delegation for For component
	if body := doc.QuerySelector("body"); body != nil {
		dom.DelegateEvent(body, "click", ".remove-item", func(event domv2.Event, target domv2.Element) {
			if indexStr := target.GetAttribute("data-index"); indexStr != "" {
				if index, err := strconv.Atoi(indexStr); err == nil {
					currentItems := items.Get()
					if index >= 0 && index < len(currentItems) {
						updatedItems := append(currentItems[:index], currentItems[index+1:]...)
						items.Set(updatedItems)
					}
				}
			}
		})
	}

	// Tab navigation buttons
	if homeTab := doc.GetElementByID("tab-home"); homeTab != nil {
		dom.BindClickToCallback(homeTab, func() {
			selectedTab.Set("home")
		})
	}
	if aboutTab := doc.GetElementByID("tab-about"); aboutTab != nil {
		dom.BindClickToCallback(aboutTab, func() {
			selectedTab.Set("about")
		})
	}
	if contactTab := doc.GetElementByID("tab-contact"); contactTab != nil {
		dom.BindClickToCallback(contactTab, func() {
			selectedTab.Set("contact")
		})
	}

	// Dynamic component buttons
	if helloBtn := doc.GetElementByID("load-hello-comp"); helloBtn != nil {
		dom.BindClickToCallback(helloBtn, func() {
			currentComponent.Set(func() gomponents.Node {
				return html.Div(
					html.Class("hello-component"),
					html.DataAttr("dynamic-component", "hello"),
					html.H4(gomponents.Text("Hello Component")),
					html.P(gomponents.Text("Hello from Dynamic Component!")),
				)
			})
		})
	}
	if counterBtn := doc.GetElementByID("load-counter-comp"); counterBtn != nil {
		dom.BindClickToCallback(counterBtn, func() {
			localCounter := reactivity.CreateSignal(0)
			currentComponent.Set(func() gomponents.Node {
				return html.Div(
					html.Class("counter-component"),
					html.DataAttr("dynamic-component", "counter"),
					html.H4(gomponents.Text("Dynamic Counter")),
					html.P(
						gomponents.Text("Count: "),
						comps.BindText(func() string {
							return strconv.Itoa(localCounter.Get())
						}),
					),
					html.Button(
						html.ID("dyn-counter-btn"),
						gomponents.Text("Increment"),
					),
				)
			})
		})
	}
	if clearBtn := doc.GetElementByID("clear-comp"); clearBtn != nil {
		dom.BindClickToCallback(clearBtn, func() {
			currentComponent.Set(nil)
		})
	}

	// Dynamic counter button delegation
	if body := doc.QuerySelector("body"); body != nil {
		dom.DelegateEvent(body, "click", "#dyn-counter-btn", func(event domv2.Event, target domv2.Element) {
			// Find the dynamic counter component and increment its local counter
			// This is a simplified approach - in a real app you'd want better state management
			js.Global().Get("console").Call("log", "Dynamic counter button clicked")
		})
	}
}

func addTodo(todos reactivity.Signal[[]string], newTodo reactivity.Signal[string]) {
	todoText := newTodo.Get()
	if todoText != "" {
		currentTodos := todos.Get()
		updatedTodos := append(currentTodos, todoText)
		todos.Set(updatedTodos)
		newTodo.Set("")
	}
}

func deleteTodo(todos reactivity.Signal[[]string], index int) {
	currentTodos := todos.Get()
	if index >= 0 && index < len(currentTodos) {
		updatedTodos := make([]string, 0, len(currentTodos)-1)
		updatedTodos = append(updatedTodos, currentTodos[:index]...)
		updatedTodos = append(updatedTodos, currentTodos[index+1:]...)
		todos.Set(updatedTodos)
	}
}

func createDynamicElement() {
	doc := domv2.GetWindow().Document()
	container := doc.GetElementByID("dynamic-container")
	if container == nil {
		return
	}

	// Create a new element using dom/v2
	div := doc.CreateElement("div")
	timestamp := time.Now().Format("15:04:05")

	// Set initial content
	div.SetTextContent(fmt.Sprintf("Created at %s", timestamp))
	div.SetAttribute("style", "padding: 10px; margin: 5px; background: #f0f0f0; border: 1px solid #ccc; border-radius: 4px;")

	// Add a click handler that changes the background color
	colorToggle := reactivity.CreateSignal(false)
	div.AddEventListener("click", false, func(event domv2.Event) {
		colorToggle.Set(!colorToggle.Get())
		if colorToggle.Get() {
			div.SetAttribute("style", "padding: 10px; margin: 5px; background: #ffeb3b; border: 1px solid #ccc; border-radius: 4px;")
		} else {
			div.SetAttribute("style", "padding: 10px; margin: 5px; background: #f0f0f0; border: 1px solid #ccc; border-radius: 4px;")
		}
	})

	// Create a delete button
	deleteBtn := doc.CreateElement("button")
	deleteBtn.SetTextContent("Delete")
	deleteBtn.SetAttribute("style", "margin-left: 10px; padding: 5px 10px; background: #f44336; color: white; border: none; border-radius: 3px; cursor: pointer;")

	deleteBtn.AddEventListener("click", false, func(event domv2.Event) {
		// Remove from DOM
		container.RemoveChild(div)
	})

	// Append delete button to the div
	div.AppendChild(deleteBtn)

	// Append to container
	container.AppendChild(div)
}

func exposeGlobalFunctions(counter reactivity.Signal[int], name reactivity.Signal[string], isVisible reactivity.Signal[bool], todos reactivity.Signal[[]string], newTodo reactivity.Signal[string], items reactivity.Signal[[]string], selectedTab reactivity.Signal[string], currentComponent reactivity.Signal[func() gomponents.Node]) {
	// Expose counter functions
	dom.CreateNamedJSFunction("incrementCounter", func(this js.Value, args []js.Value) any {
		counter.Set(counter.Get() + 1)
		return nil
	})

	dom.CreateNamedJSFunction("decrementCounter", func(this js.Value, args []js.Value) any {
		counter.Set(counter.Get() - 1)
		return nil
	})

	dom.CreateNamedJSFunction("setCounter", func(this js.Value, args []js.Value) any {
		if len(args) > 0 {
			if value, err := strconv.Atoi(args[0].String()); err == nil {
				counter.Set(value)
			}
		}
		return nil
	})

	// Expose name functions
	dom.CreateNamedJSFunction("setName", func(this js.Value, args []js.Value) any {
		if len(args) > 0 {
			name.Set(args[0].String())
		}
		return nil
	})

	// Expose visibility functions
	dom.CreateNamedJSFunction("toggleVisibility", func(this js.Value, args []js.Value) any {
		isVisible.Set(!isVisible.Get())
		return nil
	})

	// Expose todo functions
	dom.CreateNamedJSFunction("addTodoFromJS", func(this js.Value, args []js.Value) any {
		if len(args) > 0 {
			addTodo(todos, reactivity.CreateSignal(args[0].String()))
		}
		return nil
	})

	dom.CreateNamedJSFunction("clearTodos", func(this js.Value, args []js.Value) any {
		todos.Set([]string{})
		return nil
	})

	// Expose cleanup function
	dom.CreateNamedJSFunction("cleanupAll", func(this js.Value, args []js.Value) any {
		dom.CleanupAllEvents()
		dom.CleanupAllJSFunctions()
		return nil
	})

	fmt.Println("Global functions exposed:")
	fmt.Println("- incrementCounter()")
	fmt.Println("- decrementCounter()")
	fmt.Println("- setCounter(value)")
	fmt.Println("- setName(name)")
	fmt.Println("- toggleVisibility()")
	fmt.Println("- addTodoFromJS(todo)")
	fmt.Println("- clearTodos()")
	fmt.Println("- cleanupAll()")
}
