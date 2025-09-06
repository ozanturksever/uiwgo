//go:build js && wasm

package main

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/ozanturksever/logutil"
	comps "github.com/ozanturksever/uiwgo/comps"
	dom "github.com/ozanturksever/uiwgo/dom"
	reactivity "github.com/ozanturksever/uiwgo/reactivity"
	domv2 "honnef.co/go/js/dom/v2"

	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

type Todo struct {
	ID        int
	Title     string
	Completed bool
}

func main() {
	// Mount the app and get a disposer function
	disposer := comps.Mount("app", func() Node { return TodoApp() })
	_ = disposer // We don't use it in this example since the app runs indefinitely

	// Prevent exit
	select {}
}

func TodoApp() Node {
	// State
	todos := reactivity.CreateSignal([]Todo{})
	nextID := 1

	// Derived state
	remaining := reactivity.CreateMemo(func() int {
		cnt := 0
		for _, t := range todos.Get() {
			if !t.Completed {
				cnt++
			}
		}
		return cnt
	})
	hasCompleted := reactivity.CreateMemo(func() bool {
		for _, t := range todos.Get() {
			if t.Completed {
				return true
			}
		}
		return false
	})

	// Lifecycle: log stats on changes with cleanup
	reactivity.CreateEffect(func() {
		comps.OnCleanup(func() { logutil.Log("[TodoApp] cleanup before stats recompute") })
		logutil.Logf("[TodoApp] todos=%d remaining=%d", len(todos.Get()), remaining.Get())
	})

	// Helper functions for todo operations
	addTodo := func(title string) {
		title = strings.TrimSpace(title)
		if title == "" {
			return
		}
		list := append([]Todo{}, todos.Get()...)
		list = append(list, Todo{ID: nextID, Title: title})
		nextID++
		todos.Set(list)
	}

	toggleTodo := func(id int) {
		list := append([]Todo{}, todos.Get()...)
		for i := range list {
			if list[i].ID == id {
				list[i].Completed = !list[i].Completed
				break
			}
		}
		todos.Set(list)
	}

	deleteTodo := func(id int) {
		src := todos.Get()
		list := make([]Todo, 0, len(src))
		for _, t := range src {
			if t.ID != id {
				list = append(list, t)
			}
		}
		todos.Set(list)
	}

	clearCompleted := func() {
		src := todos.Get()
		list := make([]Todo, 0, len(src))
		for _, t := range src {
			if !t.Completed {
				list = append(list, t)
			}
		}
		todos.Set(list)
	}

	// Setup DOM event handlers after mount
	comps.OnMount(func() {
		// Add todo button event
		if addBtn := dom.GetElementByID("add-todo-btn"); addBtn != nil {
			dom.BindClickToCallback(addBtn, func() {
				if input := dom.GetElementByID("new-todo-input"); input != nil {
					value := input.Underlying().Get("value").String()
					addTodo(value)
					input.Underlying().Set("value", "")
				}
			})
		}

		// Enter key on input
		if input := dom.GetElementByID("new-todo-input"); input != nil {
			dom.BindEnterKeyToCallback(input, func() {
				value := input.Underlying().Get("value").String()
				addTodo(value)
				input.Underlying().Set("value", "")
			})
		}

		// Clear completed button event
		//if clearBtn := dom.GetElementByID("clear-completed-btn"); clearBtn != nil {
		//	fmt.Println("[TodoApp] clear completed button")
		//	dom.BindClickToCallback(clearBtn, clearCompleted)
		//}

		// Todo list delegation for toggle and remove buttons
		if todoList := dom.GetElementByID("todo-list"); todoList != nil {
			dom.DelegateEvent(todoList, "click", "[data-action='toggle']", func(e domv2.Event, target domv2.Element) {
				idStr := target.GetAttribute("data-id")
				if id, err := strconv.Atoi(idStr); err == nil {
					toggleTodo(id)
				}
			})

			dom.DelegateEvent(todoList, "click", "[data-action='destroy']", func(e domv2.Event, target domv2.Element) {
				idStr := target.GetAttribute("data-id")
				if id, err := strconv.Atoi(idStr); err == nil {
					deleteTodo(id)
				}
			})
		}
	})

	// Components
	return Div(
		Style("font-family: Arial, sans-serif; max-width: 700px; margin: 40px auto; padding: 20px; background-color: #f5f5f5; min-height: 100vh;"),
		Div(
			Style("background: white; padding: 30px; border-radius: 10px; box-shadow: 0 2px 10px rgba(0,0,0,0.1);"),
			AppHeader("TodoMVC (UiwGo)", "Demonstrates composition, reactive list, and lifecycle"),
			TodoInput(),
			comps.Show(comps.ShowProps{When: reactivity.CreateMemo(func() bool { return len(todos.Get()) == 0 }), Children: P(Style("color:#777; font-style: italic;"), Text("No todos yet. Add one!"))}),
			TodoList(todos),
			StatsFooter(remaining, hasCompleted, clearCompleted),
		),
	)
}

func AppHeader(title, subtitle string) Node {
	return Div(
		H1(Text(title)),
		P(Text(subtitle)),
	)
}

func TodoInput() Node {
	return Div(
		Style("display:flex; gap: 10px; margin: 10px 0;"),
		comps.OnMount(func() {
			if input := dom.GetElementByID("new-todo-input"); input != nil {
				input.Underlying().Call("focus")
			}
		}),
		Input(Type("text"), ID("new-todo-input"), Placeholder("What needs to be done?"), Style("flex:1; padding: 10px; font-size: 1rem;")),
		Button(ID("add-todo-btn"), Text("Add"), Style("padding: 10px 16px;")),
	)
}

func TodoList(todos reactivity.Signal[[]Todo]) Node {
	return Div(
		ID("todo-list"),
		comps.BindHTML(func() Node {
			items := make([]Node, 0)
			for _, t := range todos.Get() {
				checkbox := Input(Class("todo-toggle"), Type("checkbox"), Attr("data-id", fmt.Sprintf("%d", t.ID)), Attr("data-action", "toggle"))
				if t.Completed {
					checkbox = Input(Class("todo-toggle"), Type("checkbox"), Attr("data-id", fmt.Sprintf("%d", t.ID)), Attr("data-action", "toggle"), Attr("checked", "true"))
				}
				items = append(items,
					Li(
						Class("todo-item"),
						Style("display:flex; align-items:center; gap:10px; padding: 6px 0;"),
						checkbox,
						Span(Text(t.Title), Style("flex:1;")),
						Button(Class("todo-destroy"), Text("×"), Style("padding:4px 8px"), Attr("data-id", fmt.Sprintf("%d", t.ID)), Attr("data-action", "destroy")),
					),
				)
			}
			return Ul(items...)
		}),
	)
}

func StatsFooter(remaining reactivity.Signal[int], hasCompleted reactivity.Signal[bool], clearCompleted func()) Node {
	return Div(
		ID("stats-footer"),
		Style("display:flex; align-items:center; justify-content: space-between; margin-top: 12px; color:#555;"),
		Div(
			comps.BindHTML(func() Node {
				count := remaining.Get()
				itemText := "items"
				if count == 1 {
					itemText = "item"
				}
				return Text(fmt.Sprintf("%d %s left", count, itemText))
			}),
		),
		comps.Show(comps.ShowProps{When: hasCompleted, Children: Button(
			ID("clear-completed-btn"),
			Text("Clear completed"),
			dom.OnClick("clear-completed-btn", func() {
				logutil.Log("[TodoApp] clear completed button")
				clearCompleted()
			}),
		)}),
	)
}
