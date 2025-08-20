package main

import (
	"app/golid"
	"fmt"

	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

type Todo struct {
	ID        int
	Text      string
	Completed bool
}

func main() {
	_, cleanup := golid.CreateRoot(func() interface{} {
		golid.Render(TodoApp())
		return nil
	})
	defer cleanup()
	golid.Run()
}

func TodoApp() Node {
	// State
	todos, setTodos := golid.CreateSignal([]Todo{})
	newTodoText := golid.NewSignal("")
	nextID, setNextID := golid.CreateSignal(1)

	toggleTodo := func(id int) {
		currentTodos := todos()
		for i, todo := range currentTodos {
			if todo.ID == id {
				currentTodos[i].Completed = !currentTodos[i].Completed
				break
			}
		}
		setTodos(currentTodos)
	}

	deleteTodo := func(id int) {
		currentTodos := todos()
		var newTodos []Todo
		for _, todo := range currentTodos {
			if todo.ID != id {
				newTodos = append(newTodos, todo)
			}
		}
		setTodos(newTodos)
	}

	return Div(
		Style(`
			font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
			max-width: 600px;
			margin: 0 auto;
			padding: 20px;
			background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
			min-height: 100vh;
		`),

		H1(
			Text("My Awesome Todo List"),
			Style("color: white; text-align: center; margin-bottom: 30px;"),
		),

		// Add todo form
		Div(
			Style("display: flex; gap: 10px; margin-bottom: 20px;"),
			Input(
				Type("text"),
				Placeholder("Add a new todo..."),
				golid.BindInput(newTodoText, "Add a new todo..."),
				Style(`
					flex: 1;
					padding: 12px;
					border: none;
					border-radius: 8px;
					font-size: 16px;
					outline: none;
				`),
			),
			Button(
				Text("Add"),
				golid.OnClickV2(func() {
					text := newTodoText.Get()
					if text != "" {
						currentTodos := todos()
						id := nextID()
						newTodo := Todo{
							ID:        id,
							Text:      text,
							Completed: false,
						}
						setTodos(append(currentTodos, newTodo))
						newTodoText.Set("")
						setNextID(id + 1)
					}
				}),
				Style(`
					padding: 12px 24px;
					background: #4CAF50;
					color: white;
					border: none;
					border-radius: 8px;
					cursor: pointer;
					font-size: 16px;
				`),
			),
		),

		// Todo list
		Div(
			Style("background: rgba(255, 255, 255, 0.1); border-radius: 12px; padding: 20px;"),
			golid.Bind(func() Node {
				currentTodos := todos()
				if len(currentTodos) == 0 {
					return Div(
						Style("text-align: center; padding: 40px; color: rgba(255,255,255,0.7); font-style: italic;"),
						Text("No todos yet. Add one above!"),
					)
				}
				var todoNodes []Node
				for _, todo := range currentTodos {
					todoNodes = append(todoNodes, TodoItem(todo, toggleTodo, deleteTodo))
				}
				return Div(todoNodes...)
			}),
		),

		// Stats
		Div(
			Style("margin-top: 20px; text-align: center; color: white;"),
			P(
				golid.Bind(func() Node {
					todoList := todos()
					total := len(todoList)
					completed := 0
					for _, todo := range todoList {
						if todo.Completed {
							completed++
						}
					}
					return Text(fmt.Sprintf("Total: %d | Completed: %d | Remaining: %d", total, completed, total-completed))
				}),
			),
		),
	)
}

func TodoItem(todo Todo, toggleTodo func(int), deleteTodo func(int)) Node {
	return Div(
		Style(`
			display: flex;
			align-items: center;
			gap: 10px;
			padding: 12px;
			margin-bottom: 8px;
			background: rgba(255, 255, 255, 0.1);
			border-radius: 8px;
			color: white;
		`),
		Input(
			Type("checkbox"),
			func() Node {
				if todo.Completed {
					return Checked()
				}
				return nil
			}(),
			golid.OnClickV2(func() {
				toggleTodo(todo.ID)
			}),
			Style("margin-right: 10px;"),
		),
		Span(
			Text(todo.Text),
			Style(func() string {
				if todo.Completed {
					return "flex: 1; text-decoration: line-through; opacity: 0.6;"
				}
				return "flex: 1;"
			}()),
		),
		Button(
			Text("Delete"),
			golid.OnClickV2(func() {
				deleteTodo(todo.ID)
			}),
			Style(`
				padding: 6px 12px;
				background: #f44336;
				color: white;
				border: none;
				border-radius: 4px;
				cursor: pointer;
				font-size: 14px;
			`),
		),
	)
}
