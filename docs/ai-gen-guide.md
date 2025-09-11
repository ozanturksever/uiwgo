# AI Developer's Guide to Generating UI with the Go-WASM Framework

This guide outlines the core principles and patterns for generating idiomatic UI code using the Go-WASM framework. The goal is to build fully reactive, component-based user interfaces written entirely in Go.

## Core Philosophy

1.  **Declarative UI in Go:** The UI is a direct result of the state. You declare what the UI should look like for a given state, and the framework handles the rendering. All UI is defined using gomponents in Go code (`html.Div`, `html.Button`, etc.), not in separate template files.
2.  **State-Driven Reactivity:** All application state that can change must be wrapped in a `reactivity.Signal`. The UI automatically updates when these signals change. Logic should be driven by updating signals, not by manually manipulating the DOM.
3.  **Component-Based Architecture:** The UI is built by composing small, reusable functions called components. A component is simply a Go function that returns a `g.Node`.

## 1. State Management with Signals

The foundation of any application is its state.

### reactivity.CreateSignal[T]

Use signals for any piece of data that can change and should trigger a UI update.

-   **Creation:** `mySignal := reactivity.CreateSignal("initial value")`
-   **Reading:** `currentValue := mySignal.Get()`
-   **Writing:** `mySignal.Set("new value")` (This triggers the reactivity)

```go
// State for a simple counter
type CounterComponent struct {
    count reactivity.Signal[int]
}

func NewCounterComponent() *CounterComponent {
    return &CounterComponent{
        count: reactivity.CreateSignal(0),
    }
}
```

### reactivity.CreateMemo[T]

Use memos for **derived state**. A memo caches its result and only recomputes when one of its underlying signals changes.

```go
// A memo that derives a message from the count signal
doubledCount := reactivity.CreateMemo(func() string {
    return fmt.Sprintf("The count is %d, and double that is %d.", c.count.Get(), c.count.Get() * 2)
})

// Use it in the UI with comps.BindText
comps.BindText(doubledCount.Get)
```

### Advanced Reactivity Primitives

While `Signal` and `Memo` are the most common tools, the framework provides other primitives for more complex scenarios.

#### reactivity.CreateEffect

An effect is a function that re-runs whenever one of its dependent signals changes. Use it for side effects that don't directly render UI, such as logging, saving to local storage, or making API calls.

```go
// Log the count to the console whenever it changes
reactivity.CreateEffect(func() {
    logutil.Log("The new count is:", c.count.Get())
})
```

#### reactivity.CreateStore

A store is used for managing complex, nested state objects. It provides deep, proxy-based reactivity, meaning you can update nested fields and the UI will react accordingly.

```go
type User struct {
    FirstName string
    LastName  string
}

// Create a store for a user object
userStore := reactivity.CreateStore(User{FirstName: "John", LastName: "Doe"})

// Update a nested field
userStore.Set(func(u User) User {
    u.FirstName = "Jane"
    return u
})

// The UI will update automatically
H1(comps.BindText(func() string {
    return fmt.Sprintf("Hello, %s", userStore.Get().FirstName)
}))
```

#### reactivity.CreateResource

A resource is the ideal way to handle asynchronous operations, especially data fetching. It automatically manages loading and error states for you. A resource is created from a source signal (e.g., a user ID) and a fetcher function. The fetcher re-runs whenever the source signal changes.

```go
// 1. A signal that provides the input for the fetcher
userID := reactivity.CreateSignal(1)

// 2. The fetcher function that performs the async operation
func fetchUser(id int) (User, error) {
    // Simulate an API call
    time.Sleep(1 * time.Second)
    if id == 2 {
        return User{}, fmt.Errorf("user not found")
    }
    return User{ID: id, Name: fmt.Sprintf("User-%d", id)}, nil
}

// 3. Create the resource
userRes := reactivity.CreateResource(userID, fetchUser)

// 4. Render the UI based on the resource's state
func() g.Node {
    if userRes.Loading() {
        return P(Text("Loading..."))
    }
    if err := userRes.Error(); err != nil {
        return P(Text(fmt.Sprintf("Error: %v", err)))
    }
    return H1(Text(userRes.Data().Name))
}
```

## 2. Advanced State Management: The Action Bus

For simple components, updating a signal directly is fine. For larger applications, the **Action Bus** provides a centralized, decoupled way to manage state changes, following a CQRS-like pattern.

*(Existing content for Action Bus remains here...)*

## 3. Application Architecture with AppManager

For a structured application, the `appmanager` package provides a robust framework for managing lifecycle, state, and routing. It is the recommended entry point for most applications.

### Core Concepts

-   **`appmanager.AppConfig`**: A struct to configure your application, including its ID, mount point, routes, and initial state.
-   **`appmanager.NewAppManager`**: Creates a new application instance from a config.
-   **`am.Initialize()`**: Sets up the application, including the router and state persistence.
-   **`am.Mount()`**: Renders the root component of your application.

### Example Usage

```go
// main.go
func main() {
    cfg := &appmanager.AppConfig{
        AppID:          "my-app",
        MountElementID: "app",
        EnableRouter:   true, // Automatically sets up the router
        Routes: []*router.RouteDefinition{
            router.Route("/", HomeComponent),
            router.Route("/about", AboutComponent),
        },
        InitialState: appmanager.AppState{
            UI:     appmanager.UIState{Theme: "light"},
            Custom: map[string]any{"version": "1.0"},
        },
    }

    am := appmanager.NewAppManager(cfg)

    if err := am.Initialize(context.Background()); err != nil {
        logutil.Logf("Failed to initialize: %v", err)
        return
    }

    // Mount the root component, which often includes the router outlet
    if err := am.Mount(RootComponent); err != nil {
        logutil.Logf("Failed to mount: %v", err)
        return
    }

    select {}
}

// RootComponent contains the main layout and the router outlet
func RootComponent() g.Node {
    return Div(
        // Header, nav, etc.
        Nav(
            router.A("/", Text("Home")),
            router.A("/about", Text("About")),
        ),
        // The router will render page components here
        Main(ID("router-outlet")),
    )
}
```

## 4. Building Components

*(Existing content for Building Components remains here, renumbered...)*

## 5. Core UI Generation Patterns

*(Existing content for Core UI Generation Patterns remains here, renumbered...)*

### View Switching and Routing

#### In-Component View Switching

Use `comps.Switch` to render different components based on a signal's value. This is ideal for tabs or multi-step forms within a single page.

#### Page-Level Routing with the `router` package

To build a multi-page application, use the `router` package. It maps URL paths to component functions and handles navigation.

**1. Define Routes:**
Create a slice of `router.RouteDefinition` structs. You can define static paths, dynamic parameters (`:id`), optional parameters (`:section?`), and wildcards (`*filepath`).

```go
routes := []*router.RouteDefinition{
    router.Route("/", HomeComponent),
    router.Route("/users/:id", UserProfileComponent),
    router.Route("/files/*filepath", FileBrowserComponent),
    // Nested routes
    router.Route("/admin", AdminLayoutComponent,
        router.Route("/", AdminDashboardComponent),
        router.Route("/settings", AdminSettingsComponent),
    ),
}
```

**2. Create a Router Instance:**
Instantiate the router with your routes and the DOM element that will serve as the rendering outlet.

```go
// Get the element where pages will be rendered
outlet := dom.GetWindow().Document().GetElementByID("app")

// Create the router
appRouter := router.New(routes, outlet)
```

**3. Navigate with the `<A>` component:**
Use `router.A` to create navigation links. It automatically handles history updates without a page reload.

```go
// Renders an <a> tag that navigates to /about
router.A("/about", Text("About Us"))
```

**4. Accessing Route Parameters:**
In your component, you can get the current route parameters from the router instance.

```go
func UserProfileComponent(props ...any) interface{} {
    params := appRouter.Params()
    userID := params["id"] // Access the :id parameter

    return H1(Text("Profile for user: " + userID))
}
```

## 6. Example: Todo App with Action Bus

This example refactors the Todo app to use the Action Bus for more structured state management.

```go
package main

import (
	"fmt"
	"strconv"
	"time"

	"github.com/ozanturksever/uiwgo/action"
	"github.com/ozanturksever/uiwgo/comps"
	"github.com/ozanturksever/uiwgo/dom"
	"github.com/ozanturksever/uiwgo/reactivity"
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

// --- Actions ---
var (
	AddTodoAction = action.DefineAction[string]("todo.add")
)

// --- State and Component ---
type Todo struct {
	ID        string
	Text      string
	Completed bool
}

type TodoApp struct {
	bus         action.Bus
	todos       reactivity.Signal[[]Todo]
	newTodoText reactivity.Signal[string]
}

func NewTodoApp(bus action.Bus) *TodoApp {
	app := &TodoApp{
		bus:         bus,
		todos:       reactivity.CreateSignal([]Todo{}),
		newTodoText: reactivity.CreateSignal(""),
	}

	// The handler now lives with the state, not in the component.
	action.OnAction(bus, AddTodoAction, func(ctx action.Context, text string) {
		if text != "" {
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

func (app *TodoApp) render() Node {
	return Div(
		H1(Text("Go-WASM Todo List (with Action Bus)")),

		Form(
			// The form now dispatches an action instead of modifying state directly.
			dom.OnSubmitInline(func(el dom.Element, data map[string]string) {
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
			comps.For(comps.ForProps[Todo]{
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
```