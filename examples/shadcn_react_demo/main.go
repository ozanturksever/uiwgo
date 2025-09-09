//go:build js && wasm

package main

import (
	"fmt"
	"strconv"
	"syscall/js"

	"github.com/ozanturksever/logutil"
	"github.com/ozanturksever/uiwgo/compat/react"
	"github.com/ozanturksever/uiwgo/reactivity"
)

type Todo struct {
	ID        int    `json:"id"`
	Text      string `json:"text"`
	Completed bool   `json:"completed"`
}

// ShadcnDemoApp represents the main application component
type ShadcnDemoApp struct {
	counter     reactivity.Signal[int]
	theme       reactivity.Signal[string]
	todos       reactivity.Signal[[]Todo]
	newTodoText reactivity.Signal[string]
	bridge      *react.ReactBridge
	componentID react.ComponentID
	nextTodoID  int
}

// NewShadcnDemoApp creates a new ShadcnDemoApp instance
func NewShadcnDemoApp() (*ShadcnDemoApp, error) {
	bridge, err := react.GetBridge()
	if err != nil {
		return nil, fmt.Errorf("failed to get React bridge: %w", err)
	}

	return &ShadcnDemoApp{
		counter:     reactivity.CreateSignal(0),
		theme:       reactivity.CreateSignal("light"),
		todos:       reactivity.CreateSignal([]Todo{}),
		newTodoText: reactivity.CreateSignal(""),
		bridge:      bridge,
		nextTodoID:  1,
	}, nil
}

func main() {
	logutil.Log("Starting ShadcnDemoApp")

	// Initialize React bridge
	err := react.InitializeBridge()
	if err != nil {
		logutil.Logf("Failed to initialize React bridge: %v", err)
		return
	}

	app, err := NewShadcnDemoApp()
	if err != nil {
		logutil.Logf("Failed to create app: %v", err)
		return
	}

	err = app.Render()
	if err != nil {
		logutil.Logf("Failed to render app: %v", err)
		return
	}

	logutil.Log("ShadcnDemoApp started successfully")

	// Keep the program running
	select {}
}

// Render renders the application using React components
func (app *ShadcnDemoApp) Render() error {
	logutil.Log("Rendering ShadcnDemoApp with React")

	// Set up effects for reactive updates
	app.setupEffects()

	// Set up event handlers for Go-side logic
	app.setupEventHandlers()

	// Render the main React component
	props := react.Props{
		"counter":     app.counter.Get(),
		"theme":       app.theme.Get(),
		"todos":       app.todos.Get(),
		"newTodoText": app.newTodoText.Get(),
	}

	componentID, err := app.bridge.Render("ShadcnDemo", props, &react.RenderOptions{
		ContainerID: "app",
		Replace:     true,
	})
	if err != nil {
		return fmt.Errorf("failed to render React component: %w", err)
	}

	app.componentID = componentID
	logutil.Log("ShadcnDemoApp rendered successfully with React")
	return nil
}

// setupEffects sets up reactive effects for React component updates
func (app *ShadcnDemoApp) setupEffects() {
	// Counter effect - tracks counter signal
	reactivity.CreateEffect(func() {
		_ = app.counter.Get() // Read signal to track dependency
		app.updateReactComponent()
	})

	// Theme effect - tracks theme signal
	reactivity.CreateEffect(func() {
		_ = app.theme.Get() // Read signal to track dependency
		app.updateReactComponent()
		app.updateTheme()
	})

	// Todos effect - tracks todos signal
	reactivity.CreateEffect(func() {
		_ = app.todos.Get() // Read signal to track dependency
		app.updateReactComponent()
	})

	// New todo text effect - tracks newTodoText signal
	reactivity.CreateEffect(func() {
		_ = app.newTodoText.Get() // Read signal to track dependency
		app.updateReactComponent()
	})
}

// updateReactComponent updates the React component with current state
func (app *ShadcnDemoApp) updateReactComponent() {
	if app.componentID == "" {
		return // Component not yet rendered
	}

	props := react.Props{
		"counter":     app.counter.Get(),
		"theme":       app.theme.Get(),
		"todos":       app.todos.Get(),
		"newTodoText": app.newTodoText.Get(),
	}

	err := app.bridge.Update(app.componentID, props)
	if err != nil {
		logutil.Logf("Failed to update React component: %v", err)
	}
}

// updateTheme updates the theme using React bridge
func (app *ShadcnDemoApp) updateTheme() {
	theme := app.theme.Get()
	logutil.Logf("Updating theme to: %s", theme)

	// Use React bridge to set theme
	err := app.bridge.SetTheme(theme)
	if err != nil {
		logutil.Logf("Failed to set theme via React bridge: %v", err)
	}
}

// setupEventHandlers sets up global event handlers and exposes Go functions to React
func (app *ShadcnDemoApp) setupEventHandlers() {
	// Expose Go functions to global scope for React to call
	js.Global().Set("goIncrementCounter", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		app.incrementCounter()
		return nil
	}))

	js.Global().Set("goDecrementCounter", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		app.decrementCounter()
		return nil
	}))

	js.Global().Set("goToggleTheme", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		app.toggleTheme()
		return nil
	}))

	js.Global().Set("goAddTodo", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		app.addTodo()
		return nil
	}))

	js.Global().Set("goRemoveTodo", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		if len(args) > 0 {
			id := args[0].String()
			app.removeTodo(id)
		}
		return nil
	}))

	js.Global().Set("goUpdateNewTodoText", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		if len(args) > 0 {
			text := args[0].String()
			app.newTodoText.Set(text)
		}
		return nil
	}))
}

// incrementCounter increments the counter
func (app *ShadcnDemoApp) incrementCounter() {
	current := app.counter.Get()
	app.counter.Set(current + 1)
	logutil.Logf("Counter incremented to: %d", current+1)
}

// decrementCounter decrements the counter
func (app *ShadcnDemoApp) decrementCounter() {
	current := app.counter.Get()
	app.counter.Set(current - 1)
	logutil.Logf("Counter decremented to: %d", current-1)
}

// toggleTheme toggles between light and dark theme
func (app *ShadcnDemoApp) toggleTheme() {
	current := app.theme.Get()
	if current == "light" {
		app.theme.Set("dark")
	} else {
		app.theme.Set("light")
	}
	logutil.Logf("Theme toggled to: %s", app.theme.Get())
}

// addTodo adds a new todo
func (app *ShadcnDemoApp) addTodo() {
	text := app.newTodoText.Get()
	if text != "" {
		newTodo := Todo{
			ID:        app.nextTodoID,
			Text:      text,
			Completed: false,
		}
		app.nextTodoID++
		todos := app.todos.Get()
		app.todos.Set(append(todos, newTodo))
		app.newTodoText.Set("")
		logutil.Logf("Added todo: %s", text)
	}
}

// removeTodo removes a todo by ID
func (app *ShadcnDemoApp) removeTodo(idStr string) {
	id, err := strconv.Atoi(idStr)
	if err != nil {
		logutil.Logf("Invalid todo ID: %s", idStr)
		return
	}

	todos := app.todos.Get()
	for i, todo := range todos {
		if todo.ID == id {
			// Remove todo at index i
			newTodos := append(todos[:i], todos[i+1:]...)
			app.todos.Set(newTodos)
			logutil.Logf("Removed todo: %s", todo.Text)
			break
		}
	}
}

func parseInt(s string) int {
	if val, err := strconv.Atoi(s); err == nil {
		return val
	}
	return 0
}
