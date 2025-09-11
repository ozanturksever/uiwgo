# **AI Developer's Guide to Generating UI with the Go-WASM Framework**

This guide outlines the core principles and patterns for generating idiomatic UI code using the Go-WASM framework. The goal is to build fully reactive, component-based user interfaces written entirely in Go.

## **Core Philosophy**

1. **Declarative UI in Go:** The UI is a direct result of the state. You declare what the UI should look like for a given state, and the framework handles the rendering. All UI is defined using gomponents in Go code (html.Div, html.Button, etc.), not in separate template files.  
2. **State-Driven Reactivity:** All application state that can change must be wrapped in a reactivity.Signal. The UI automatically updates when these signals change. Logic should be driven by updating signals, not by manually manipulating the DOM.  
3. **Component-Based Architecture:** The UI is built by composing small, reusable functions called components. A component is simply a Go function that returns a g.Node.

## **1\. State Management with Signals**

The foundation of any application is its state.

### **reactivity.CreateSignal\[T\]**

Use signals for any piece of data that can change and should trigger a UI update.

* **Creation:** mySignal := reactivity.CreateSignal("initial value")  
* **Reading:** currentValue := mySignal.Get()  
* **Writing:** mySignal.Set("new value") (This triggers the reactivity)

// State for a simple counter  
type CounterComponent struct {  
    count reactivity.Signal\[int\]  
}

func NewCounterComponent() \*CounterComponent {  
    return \&CounterComponent{  
        count: reactivity.CreateSignal(0),  
    }  
}

### **reactivity.CreateMemo\[T\]**

Use memos for **derived state**. A memo caches its result and only recomputes when one of its underlying signals changes.

// A memo that derives a message from the count signal  
doubledCount := reactivity.CreateMemo(func() string {  
    return fmt.Sprintf("The count is %d, and double that is %d.", c.count.Get(), c.count.Get() \* 2\)  
})

// Use it in the UI with comps.BindText  
comps.BindText(doubledCount.Get)

## **2\. Advanced State Management: The Action Bus**

For simple components, updating a signal directly in an event handler is fine. For larger applications, this can lead to complex dependencies and "prop drilling" (passing callbacks through many layers of components).

The **Action Bus** provides a centralized, decoupled way to manage state changes. Components dispatch "actions" (events), and dedicated handlers listen for these actions to update the state. This follows a CQRS-like pattern.

### **Core Concepts**

* action.Bus: The central event bus for the application.  
* action.DefineAction\[T\]: Defines a type-safe action with a specific payload type T.  
* bus.Dispatch(action): Components use this to send an action to the bus.  
* action.OnAction(bus, action, handler): Subscribes a handler function to a specific action type. The handler contains the logic to update state signals.

### **How to Use the Action Bus**

Step 1: Define Your Actions  
Typically, actions are defined globally for your application.  
// actions/actions.go  
var (  
    IncrementAction \= action.DefineAction\[int\]("counter.increment")  
    SetUserAction   \= action.DefineAction\[string\]("user.set")  
)

Step 2: Create a Bus  
Instantiate a single bus at the root of your application.  
// main.go  
bus := action.New()

Step 3: Dispatch Actions from Components  
Instead of setting a signal directly, the component's event handler dispatches an action.  
// The component doesn't know \*how\* the count is incremented, only that it \*should\* be.  
Button(  
    Text("Increment"),  
    dom.OnClickInline(func(el dom.Element) {  
        bus.Dispatch(IncrementAction.New(1)) // Dispatch an action with a payload of 1  
    }),  
)

Step 4: Handle Actions and Update State  
In a central location (like your main function or a "store"), listen for actions and update the state.  
// main.go or app/store.go  
count := reactivity.CreateSignal(0)

// This handler listens for IncrementAction and updates the signal.  
action.OnAction(bus, IncrementAction, func(ctx action.Context, payload int) {  
    count.Set(count.Get() \+ payload)  
})

This pattern decouples your UI components from your business logic, making the application much easier to manage and test.

## **3\. Building Components**

A component is a Go function that returns a g.Node. For stateful components, use a struct to hold the signals and define a render() method.

// A stateful component structure  
type GreeterComponent struct {  
    name reactivity.Signal\[string\]  
}

// The render method defines the UI  
func (c \*GreeterComponent) render() g.Node {  
    return Div(  
        Label(For("name-input"), Text("Enter your name:")),  
        Input(  
            ID("name-input"),  
            Type("text"),  
            // Event handlers update the signal  
            dom.OnInputInline(func(el dom.Element) {  
                c.name.Set(el.Underlying().Get("value").String())  
            }),  
        ),  
        H1(  
            // The UI reacts to signal changes  
            comps.BindText(func() string {  
                return fmt.Sprintf("Hello, %s\!", c.name.Get())  
            }),  
        ),  
    )  
}

## **4\. Core UI Generation Patterns**

### **Displaying Reactive Text and HTML**

* **Static Text:** Use g.Text("Hello").  
* **Reactive Text:** Use comps.BindText(func() string { ... }). The function will be re-run whenever a signal used inside it changes.

// Renders: \<p\>Count: 0\</p\>, and updates automatically  
P(comps.BindText(func() string {  
    return fmt.Sprintf("Count: %d", c.count.Get())  
}))

### **Handling User Input & Events**

This is the most critical pattern for interactivity. **Always use dom.On...Inline functions.** The Go callback function should almost always call .Set() on a signal or bus.Dispatch() for an action.

* dom.OnClickInline(func(el dom.Element) { ... })  
* dom.OnInputInline(func(el dom.Element) { ... })  
* dom.OnSubmitInline(func(el dom.Element, formData map\[string\]string) { ... })

### **Direct DOM Interaction: honnef.co/go/js/dom/v2 vs syscall/js**

When you need to interact with a DOM element within an event handler, the framework provides a dom.Element argument. This is a wrapper around the powerful honnef.co/go/js/dom/v2 library.

**Guideline: Strongly prefer using the dom.Element wrapper and its underlying honnef.co methods over raw syscall/js calls.**

* **Why?** It provides a type-safe, idiomatic Go API for the DOM, reducing the risk of runtime panics from typos in property names or incorrect type assertions. It makes the code cleaner and more predictable.  
* **Accessing the underlying object:** The raw js.Value can be accessed via el.Underlying() for properties not yet covered by the typed wrapper.

**Example: Getting an input's value**

// ✅ DO THIS: Use the provided dom.Element wrapper.  
dom.OnInputInline(func(el dom.Element) {  
    // Access the underlying js.Value and its properties in a safe way.  
    value := el.Underlying().Get("value").String()  
    c.name.Set(value)  
})

// ❌ AVOID THIS: Raw syscall/js is verbose and unsafe.  
import "syscall/js"

dom.OnClickInline(func(el dom.Element) {  
    // This is brittle\! A typo in "value" would compile but panic at runtime.  
    value := js.Global().Get("document").Call("getElementById", "my-input").Get("value").String()  
    c.name.Set(value)  
})

### **Conditional Rendering**

* **Toggling Elements:** Use comps.Show. Its children are only rendered and attached to the DOM when the When signal is true.  
* **Toggling Attributes:** Use g.If(condition, attribute). This is perfect for applying classes or attributes like disabled or checked.

// Show a message only when count is greater than 5  
comps.Show(comps.ShowProps{  
    When: reactivity.CreateMemo(func() bool { return c.count.Get() \> 5 }).Get,  
    Children: P(Text("The count is high\!")),  
})

// Apply an "active" class to a button if it's the current view  
Button(  
    Class("tab"),  
    g.If(c.currentView.Get() \== "profile", Class("active")),  
    Text("Profile"),  
)

### **Rendering Lists**

Always use comps.For for rendering lists from a slice signal. You **must** provide a Key function that returns a unique string for each item. This is essential for efficient updates.

type Todo struct {  
    ID   string  
    Text string  
}

todos := reactivity.CreateSignal(\[\]Todo{...})

// Render the list  
Ul(  
    comps.For(comps.ForProps\[Todo\]{  
        Items: todos,  
        Key: func(todo Todo) string { return todo.ID }, // Must be unique  
        Children: func(todo Todo, index int) g.Node {  
            // This function is called for each item  
            return Li(Text(todo.Text))  
        },  
    }),  
)

### **View Switching and Routing**

* **In-Component View Switching:** Use comps.Switch to render different components based on a signal's value. This is ideal for tabs or multi-step forms within a single page.  
* **Page-Level Routing:** To build a multi-page application, use the router package. Define routes that map URL paths to the main component function for that page.

## **5\. Example: Todo App with Action Bus**

This example refactors the Todo app to use the Action Bus for more structured state management.

package main

import (  
	"fmt"  
	"strconv"  
	"time"

	"\[github.com/ozanturksever/uiwgo/action\](https://github.com/ozanturksever/uiwgo/action)"  
	"\[github.com/ozanturksever/uiwgo/comps\](https://github.com/ozanturksever/uiwgo/comps)"  
	"\[github.com/ozanturksever/uiwgo/dom\](https://github.com/ozanturksever/uiwgo/dom)"  
	"\[github.com/ozanturksever/uiwgo/reactivity\](https://github.com/ozanturksever/uiwgo/reactivity)"  
	. "maragu.dev/gomponents"  
	. "maragu.dev/gomponents/html"  
)

// \--- Actions \---  
var (  
	AddTodoAction \= action.DefineAction\[string\]("todo.add")  
)

// \--- State and Component \---  
type Todo struct {  
	ID        string  
	Text      string  
	Completed bool  
}

type TodoApp struct {  
	bus         action.Bus  
	todos       reactivity.Signal\[\[\]Todo\]  
	newTodoText reactivity.Signal\[string\]  
}

func NewTodoApp(bus action.Bus) \*TodoApp {  
	app := \&TodoApp{  
		bus:         bus,  
		todos:       reactivity.CreateSignal(\[\]Todo{}),  
		newTodoText: reactivity.CreateSignal(""),  
	}

	// The handler now lives with the state, not in the component.  
	action.OnAction(bus, AddTodoAction, func(ctx action.Context, text string) {  
		if text \!= "" {  
			newTodo := Todo{  
				ID:   strconv.FormatInt(time.Now().UnixNano(), 10),  
				Text: text,  
			}  
			app.todos.Set(append(app.todos.Get(), newTodo))  
			app.newTodoText.Set("") // Clear input field  
		}  
	})

	return app  
}

func (app \*TodoApp) render() Node {  
	return Div(  
		H1(Text("Go-WASM Todo List (with Action Bus)")),

		Form(  
			// The form now dispatches an action instead of modifying state directly.  
			dom.OnSubmitInline(func(el dom.Element, data map\[string\]string) {  
				app.bus.Dispatch(AddTodoAction.New(app.newTodoText.Get()))  
			}),  
			Input(  
				Type("text"),  
				Attr("placeholder", "What needs to be done?"),  
				// Bind the input's value to our signal  
				Attr("value", app.newTodoText.Get()),  
				dom.OnInputInline(func(el dom.Element) {  
					app.newTodoText.Set(el.Underlying().Get("value").String())  
				}),  
			),  
			Button(Type("submit"), Text("Add Todo")),  
		),

		Ul(  
			comps.For(comps.ForProps\[Todo\]{  
				Items: app.todos,  
				Key:   func(item Todo) string { return item.ID },  
				Children: func(item Todo, index int) Node {  
					return Li(  
						Class("todo-item"),  
						g.If(item.Completed, Class("completed")),  
						Text(item.Text),  
					)  
				},  
			}),  
		),  
	)  
}

func main() {  
	bus := action.New()  
	app := NewTodoApp(bus)

	comps.Mount("app", app.render)  
	select {}  
}  
