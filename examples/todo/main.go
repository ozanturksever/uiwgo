package main

import (
	"app/golid"
	"fmt"
	"time"

	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

type Todo struct {
	ID        int
	Text      string
	Completed bool
	CreatedAt time.Time
}

func main() {
	golid.Render(TodoApp())
	golid.Run()
}

func TodoApp() Node {
	return Div(
		Style("font-family: Arial, sans-serif; max-width: 800px; margin: 50px auto; padding: 20px; background-color: #f5f5f5; min-height: 100vh;"),
		Div(
			Style("background: white; padding: 30px; border-radius: 10px; box-shadow: 0 2px 10px rgba(0,0,0,0.1);"),
			H1(Style("text-align: center; color: #333; margin-bottom: 30px;"), Text("Todo List Example")),
			P(Style("text-align: center; color: #666; margin-bottom: 30px;"), Text("Demonstrates list rendering, form handling, and reactive updates")),
			TodoComponent(),
		),
	)
}

func TodoComponent() Node {
	// Signals for managing state
	todos := golid.NewSignal([]Todo{
		{ID: 1, Text: "Learn Golid framework", Completed: false, CreatedAt: time.Now().Add(-time.Hour)},
		{ID: 2, Text: "Build a todo app", Completed: false, CreatedAt: time.Now().Add(-30 * time.Minute)},
		{ID: 3, Text: "Deploy to production", Completed: false, CreatedAt: time.Now()},
	})
	newTodoText := golid.NewSignal("")
	nextID := golid.NewSignal(4)

	return Div(
		// Add new todo form
		Div(
			Style("margin-bottom: 30px; padding: 20px; background-color: #f8f9fa; border-radius: 8px;"),
			H3(Style("margin-top: 0; color: #333;"), Text("Add New Todo")),
			Div(
				Style("display: flex; gap: 10px; align-items: center;"),
				Input(
					Type("text"),
					Placeholder("Enter new todo..."),
					Style("flex: 1; padding: 10px; border: 2px solid #ddd; border-radius: 5px; font-size: 16px;"),
					golid.OnInput(func(val string) {
						newTodoText.Set(val)
					}),
				),
				Button(
					Style("padding: 10px 20px; background-color: #007bff; color: white; border: none; border-radius: 5px; cursor: pointer; font-size: 16px;"),
					Text("Add Todo"),
					golid.OnClick(func() {
						text := newTodoText.Get()
						if text != "" {
							currentTodos := todos.Get()
							newTodo := Todo{
								ID:        nextID.Get(),
								Text:      text,
								Completed: false,
								CreatedAt: time.Now(),
							}
							todos.Set(append(currentTodos, newTodo))
							nextID.Set(nextID.Get() + 1)
							newTodoText.Set("")
							// Clear the input field
							//golid.NodeFromID("new-todo-input").Set("value", "")
						}
					}),
				),
			),
		),

		// Todo statistics
		Div(
			Style("margin-bottom: 20px; padding: 15px; background-color: #e7f3ff; border-radius: 8px;"),
			golid.BindText(func() string {
				currentTodos := todos.Get()
				total := len(currentTodos)
				completed := 0
				for _, todo := range currentTodos {
					if todo.Completed {
						completed++
					}
				}
				return fmt.Sprintf("📊 Total: %d | ✅ Completed: %d | ⏳ Remaining: %d", total, completed, total-completed)
			}),
		),

		// Todo list
		Div(
			H3(Style("color: #333; margin-bottom: 15px;"), Text("Todo Items")),
			golid.Bind(func() Node {
				currentTodos := todos.Get()
				if len(currentTodos) == 0 {
					return Div(
						Style("text-align: center; padding: 40px; color: #666; font-style: italic;"),
						Text("No todos yet. Add one above!"),
					)
				}
				return Div(
					golid.ForEach(currentTodos, func(todo Todo) Node {
						return TodoItem(todo, todos)
					}),
				)
			}),
		),

		// Clear completed button
		Div(
			Style("margin-top: 20px; text-align: center;"),
			Button(
				Style("padding: 10px 20px; background-color: #dc3545; color: white; border: none; border-radius: 5px; cursor: pointer;"),
				Text("Clear Completed"),
				golid.OnClick(func() {
					currentTodos := todos.Get()
					var remainingTodos []Todo
					for _, todo := range currentTodos {
						if !todo.Completed {
							remainingTodos = append(remainingTodos, todo)
						}
					}
					todos.Set(remainingTodos)
				}),
			),
		),
	)
}

func TodoItem(todo Todo, todosSignal *golid.Signal[[]Todo]) Node {
	return Div(
		Style(fmt.Sprintf("display: flex; align-items: center; gap: 15px; padding: 15px; margin-bottom: 10px; background-color: %s; border-radius: 8px; border-left: 4px solid %s;",
			func() string {
				if todo.Completed {
					return "#f8f9fa"
				}
				return "white"
			}(),
			func() string {
				if todo.Completed {
					return "#28a745"
				}
				return "#007bff"
			}())),

		// Checkbox to toggle completion
		Input(
			Type("checkbox"),
			Style("transform: scale(1.2); cursor: pointer;"),
			func() Node {
				if todo.Completed {
					return Checked()
				}
				return nil
			}(),
			golid.OnClick(func() {
				currentTodos := todosSignal.Get()
				for i, t := range currentTodos {
					if t.ID == todo.ID {
						currentTodos[i].Completed = !currentTodos[i].Completed
						break
					}
				}
				todosSignal.Set(currentTodos)
			}),
		),

		// Todo text
		Div(
			Style(fmt.Sprintf("flex: 1; font-size: 16px; %s",
				func() string {
					if todo.Completed {
						return "text-decoration: line-through; color: #666;"
					}
					return "color: #333;"
				}())),
			Text(todo.Text),
		),

		// Created time
		Div(
			Style("font-size: 12px; color: #999;"),
			Text(todo.CreatedAt.Format("15:04")),
		),

		// Delete button
		Button(
			Style("padding: 5px 10px; background-color: #dc3545; color: white; border: none; border-radius: 3px; cursor: pointer; font-size: 12px;"),
			Text("Delete"),
			golid.OnClick(func() {
				currentTodos := todosSignal.Get()
				var newTodos []Todo
				for _, t := range currentTodos {
					if t.ID != todo.ID {
						newTodos = append(newTodos, t)
					}
				}
				todosSignal.Set(newTodos)
			}),
		),
	)
}
