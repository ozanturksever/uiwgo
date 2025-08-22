//go:build js && wasm

package main

import (
	"fmt"
	"strings"
	"syscall/js"

	comps "github.com/ozanturksever/uiwgo/comps"
	reactivity "github.com/ozanturksever/uiwgo/reactivity"

	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

type Todo struct {
	ID        int
	Title     string
	Completed bool
}

func main() {
	comps.Mount("app", func() Node { return TodoApp() })
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
		comps.OnCleanup(func() { fmt.Println("[TodoApp] cleanup before stats recompute") })
		fmt.Printf("[TodoApp] todos=%d remaining=%d\n", len(todos.Get()), remaining.Get())
	})

	// Expose window functions for events
	addFn := js.FuncOf(func(this js.Value, args []js.Value) any {
		doc := js.Global().Get("document")
		v := doc.Call("getElementById", "newTodo").Get("value").String()
		title := strings.TrimSpace(v)
		if title == "" {
			return nil
		}
		list := append([]Todo{}, todos.Get()...)
		list = append(list, Todo{ID: nextID, Title: title})
		nextID++
		todos.Set(list)
		// clear input
		doc.Call("getElementById", "newTodo").Set("value", "")
		return nil
	})
	toggleFn := js.FuncOf(func(this js.Value, args []js.Value) any {
		if len(args) == 0 {
			return nil
		}
		id := args[0].Int()
		list := append([]Todo{}, todos.Get()...)
		for i := range list {
			if list[i].ID == id {
				list[i].Completed = !list[i].Completed
				break
			}
		}
		todos.Set(list)
		return nil
	})
	removeFn := js.FuncOf(func(this js.Value, args []js.Value) any {
		if len(args) == 0 {
			return nil
		}
		id := args[0].Int()
		src := todos.Get()
		list := make([]Todo, 0, len(src))
		for _, t := range src {
			if t.ID != id {
				list = append(list, t)
			}
		}
		todos.Set(list)
		return nil
	})
	clearCompletedFn := js.FuncOf(func(this js.Value, args []js.Value) any {
		src := todos.Get()
		list := make([]Todo, 0, len(src))
		for _, t := range src {
			if !t.Completed {
				list = append(list, t)
			}
		}
		todos.Set(list)
		return nil
	})
	js.Global().Set("addTodo", addFn)
	js.Global().Set("toggleTodo", toggleFn)
	js.Global().Set("removeTodo", removeFn)
	js.Global().Set("clearCompleted", clearCompletedFn)

	// Components
	return Div(
		Style("font-family: Arial, sans-serif; max-width: 700px; margin: 40px auto; padding: 20px; background-color: #f5f5f5; min-height: 100vh;"),
		Div(
			Style("background: white; padding: 30px; border-radius: 10px; box-shadow: 0 2px 10px rgba(0,0,0,0.1);"),
			AppHeader("TodoMVC (UiwGo)", "Demonstrates composition, reactive list, and lifecycle"),
			TodoInput(),
			comps.Show(comps.ShowProps{When: reactivity.CreateMemo(func() bool { return len(todos.Get()) == 0 }), Children: P(Style("color:#777; font-style: italic;"), Text("No todos yet. Add one!"))}),
			TodoList(todos),
			StatsFooter(remaining, hasCompleted),
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
			doc := js.Global().Get("document")
			doc.Call("getElementById", "newTodo").Call("focus")
		}),
		Input(Type("text"), ID("newTodo"), Placeholder("What needs to be done?"), Style("flex:1; padding: 10px; font-size: 1rem;")),
		Button(Text("Add"), Style("padding: 10px 16px;"), Attr("onclick", "window.addTodo()")),
	)
}

func TodoList(todos reactivity.Signal[[]Todo]) Node {
	return Div(
		comps.BindHTML(func() Node {
			items := make([]Node, 0)
			for _, t := range todos.Get() {
				checkbox := Input(Type("checkbox"), Attr("onclick", fmt.Sprintf("window.toggleTodo(%d)", t.ID)))
				if t.Completed {
					checkbox = Input(Type("checkbox"), Attr("onclick", fmt.Sprintf("window.toggleTodo(%d)", t.ID)), Attr("checked", "true"))
				}
				items = append(items,
					Li(
						Style("display:flex; align-items:center; gap:10px; padding: 6px 0;"),
						checkbox,
						Span(Text(t.Title), Style("flex:1;")),
						Button(Text("×"), Style("padding:4px 8px"), Attr("onclick", fmt.Sprintf("window.removeTodo(%d)", t.ID))),
					),
				)
			}
			return Ul(items...)
		}),
	)
}

func StatsFooter(remaining reactivity.Signal[int], hasCompleted reactivity.Signal[bool]) Node {
	return Div(
		Style("display:flex; align-items:center; justify-content: space-between; margin-top: 12px; color:#555;"),
		Div(
			comps.BindText(func() string { return fmt.Sprintf("%d items left", remaining.Get()) }),
		),
		comps.Show(comps.ShowProps{
			When:     hasCompleted,
			Children: Button(Text("Clear completed"), Attr("onclick", "window.clearCompleted()")),
		}),
	)
}
