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

type TodoItem struct {
	ID        int
	Title     string
	Completed bool
}

type AppState struct {
	Todos []TodoItem
}

func main() {
	// Mount the app and get a disposer function
	disposer := comps.Mount("app", func() Node { return TodoStoreApp() })
	_ = disposer // We don't use it in this example since the app runs indefinitely
	
	// Prevent exit
	select {}
}

func TodoStoreApp() Node {
	// Create store
	store, setState := reactivity.CreateStore(AppState{Todos: []TodoItem{}})
	nextID := 1

	// Expose JS handlers
	addFn := js.FuncOf(func(this js.Value, args []js.Value) any {
		doc := js.Global().Get("document")
		v := strings.TrimSpace(doc.Call("getElementById", "new-todo-input").Get("value").String())
		if v == "" {
			return nil
		}
		// Build new slice
		// read snapshot once here (non-reactive)
		cur := store.Get().Todos
		list := make([]TodoItem, 0, len(cur)+1)
		list = append(list, cur...)
		list = append(list, TodoItem{ID: nextID, Title: v})
		nextID++
		setState("Todos", list)
		// clear
		doc.Call("getElementById", "new-todo-input").Set("value", "")
		return nil
	})
	toggleFn := js.FuncOf(func(this js.Value, args []js.Value) any {
		if len(args) == 0 {
			return nil
		}
		idx := args[0].Int()
		completed := reactivity.Adapt[bool](store.Select("Todos", idx, "Completed")).Get()
		setState("Todos", idx, "Completed", !completed)
		return nil
	})
	removeFn := js.FuncOf(func(this js.Value, args []js.Value) any {
		if len(args) == 0 {
			return nil
		}
		idx := args[0].Int()
		cur := store.Get().Todos
		if idx < 0 || idx >= len(cur) {
			return nil
		}
		list := append([]TodoItem{}, cur[:idx]...)
		list = append(list, cur[idx+1:]...)
		setState("Todos", list)
		return nil
	})
	clearCompletedFn := js.FuncOf(func(this js.Value, args []js.Value) any {
		cur := store.Get().Todos
		list := make([]TodoItem, 0, len(cur))
		for _, t := range cur {
			if !t.Completed {
				list = append(list, t)
			}
		}
		setState("Todos", list)
		return nil
	})
	js.Global().Set("addTodo", addFn)
	js.Global().Set("toggleTodo", toggleFn)
	js.Global().Set("removeTodo", removeFn)
	js.Global().Set("clearCompleted", clearCompletedFn)

	remaining := reactivity.CreateMemo(func() int {
		cnt := 0
		l := store.SelectLen("Todos").Get()
		for i := 0; i < l; i++ {
			if !reactivity.Adapt[bool](store.Select("Todos", i, "Completed")).Get() {
				cnt++
			}
		}
		return cnt
	})
	hasCompleted := reactivity.CreateMemo(func() bool {
		l := store.SelectLen("Todos").Get()
		for i := 0; i < l; i++ {
			if reactivity.Adapt[bool](store.Select("Todos", i, "Completed")).Get() {
				return true
			}
		}
		return false
	})

	return Div(
		Style("font-family: Arial, sans-serif; max-width: 700px; margin: 40px auto; padding: 20px; background-color: #f5f5f5; min-height: 100vh;"),
		Div(
			Style("background: white; padding: 30px; border-radius: 10px; box-shadow: 0 2px 10px rgba(0,0,0,0.1);"),
			H1(Text("TodoMVC (Store)")),
			P(Text("Demonstrates fine-grained store updates: only changed item re-renders")),
			TodoInput(),
			comps.BindHTML(func() Node {
				// Only depends on length; does not re-run when a single item's fields change
				l := store.SelectLen("Todos").Get()
				items := make([]Node, 0, l)
				for i := 0; i < l; i++ {
					items = append(items, TodoItemView(store, i))
				}
				allItems := append([]Node{ID("todo-list")}, items...)
				return Ul(allItems...)
			}),
			Div(
				Style("display:flex; align-items:center; justify-content: space-between; margin-top: 12px; color:#555;"),
				Div(comps.BindText(func() string { return fmt.Sprintf("%d items left", remaining.Get()) })),
				comps.Show(comps.ShowProps{When: hasCompleted, Children: Button(ID("clear-completed-btn"), Text("Clear completed"), Attr("onclick", "window.clearCompleted()"))}),
			),
		),
	)
}

func TodoInput() Node {
	return Div(
		Style("display:flex; gap: 10px; margin: 10px 0;"),
		Input(Type("text"), ID("new-todo-input"), Placeholder("What needs to be done?"), Style("flex:1; padding: 10px; font-size: 1rem;")),
		Button(ID("add-todo-btn"), Text("Add"), Style("padding: 10px 16px;"), Attr("onclick", "window.addTodo()")),
	)
}

func TodoItemView(store reactivity.Store[AppState], i int) Node {
	// Per-item rendering binder that depends only on this item's fields
	renders := 0
	return comps.BindHTMLAs("li", func() Node {
		id := reactivity.Adapt[int](store.Select("Todos", i, "ID")).Get()
		title := reactivity.Adapt[string](store.Select("Todos", i, "Title")).Get()
		completed := reactivity.Adapt[bool](store.Select("Todos", i, "Completed")).Get()
		renders++
		fmt.Printf("[Item %d] render count=%d completed=%v\n", id, renders, completed)

		checkbox := Input(Class("todo-toggle"), Type("checkbox"), Attr("onclick", fmt.Sprintf("window.toggleTodo(%d)", i)))
		if completed {
			checkbox = Input(Class("todo-toggle"), Type("checkbox"), Attr("onclick", fmt.Sprintf("window.toggleTodo(%d)", i)), Attr("checked", "true"))
		}

		return Group([]Node{
			checkbox,
			Span(Class("todo-label"), Text(title), Style("flex:1;")),
			Button(Class("todo-destroy"), Text("×"), Style("padding:4px 8px"), Attr("onclick", fmt.Sprintf("window.removeTodo(%d)", i))),
		})
	}, Class("todo-item"), Style("display:flex; align-items:center; gap:10px; padding: 6px 0;"))
}
