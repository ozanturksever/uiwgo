//go:build js && wasm

package main

import (
	"fmt"

	comps "github.com/ozanturksever/uiwgo/comps"
	dom "github.com/ozanturksever/uiwgo/dom"
	reactivity "github.com/ozanturksever/uiwgo/reactivity"

	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

func main() {
	_ = comps.Mount("app", func() Node { return InlineEventsDemo() })
	select {}
}

func InlineEventsDemo() Node {
	count := reactivity.CreateSignal(0)
	name := reactivity.CreateSignal("")
	color := reactivity.CreateSignal("red")
	todos := reactivity.CreateSignal([]string{})
	currentInput := reactivity.CreateSignal("")

	addTodo := func(text string) {
		if text == "" { return }
		items := append(todos.Get(), text)
		todos.Set(items)
	}

	return Div(
		Style("font-family: Arial, sans-serif; max-width: 700px; margin: 30px auto; padding: 20px;"),
		H1(Text("Inline Events Demo")),

		// Counter (click handlers)
		Div(
			ID("counter"),
			Div(
				ID("count-display"),
				comps.BindText(func() string { return fmt.Sprintf("Count: %d", count.Get()) }),
			),
			Div(
				Button(ID("inc-btn"), Text("+"), dom.OnClickInline(func(el dom.Element){ count.Set(count.Get()+1) })),
				Button(ID("dec-btn"), Text("-"), dom.OnClickInline(func(el dom.Element){ count.Set(count.Get()-1) })),
				Button(ID("reset-btn"), Text("Reset"), dom.OnClickInline(func(el dom.Element){ count.Set(0) })),
			),
		),

		// Input (oninput)
		Div(
			ID("name"),
			Input(
				ID("name-input"),
				Type("text"),
				Placeholder("Your name"),
				// Update name reactively via inline input handler
				dom.OnInputInline(func(el dom.Element){
					v := el.Underlying().Get("value").String()
					name.Set(v)
				}),
			),
			P(ID("hello-output"), comps.BindText(func() string { return fmt.Sprintf("Hello, %s", name.Get()) })),
		),

		// Select (onchange)
		Div(
			ID("color"),
			Select(
				ID("color-select"),
				Option(Value("red"), Text("Red")),
				Option(Value("green"), Text("Green")),
				Option(Value("blue"), Text("Blue")),
				dom.OnChangeInline(func(el dom.Element){
					v := el.Underlying().Get("value").String()
					color.Set(v)
				}),
			),
			P(ID("color-output"), comps.BindText(func() string { return fmt.Sprintf("Color: %s", color.Get()) })),
		),

		// Keydown (Enter adds todo, Escape clears input)
		Div(
			ID("todos"),
			Input(
				ID("todo-input"),
				Type("text"),
				Placeholder("Add todo and press Enter"),
				dom.OnInputInline(func(el dom.Element){ currentInput.Set(el.Underlying().Get("value").String()) }),
				dom.OnEnterInline(func(el dom.Element){
					text := el.Underlying().Get("value").String()
					addTodo(text)
					// clear
					el.Underlying().Set("value", "")
					currentInput.Set("")
				}),
				dom.OnEscapeInline(func(el dom.Element){
					el.Underlying().Set("value", "")
					currentInput.Set("")
				}),
			),
			Ul(ID("todo-list"),
				comps.BindHTML(func() Node {
					lis := make([]Node, 0)
					for _, t := range todos.Get() {
						lis = append(lis, Li(Class("todo-item"), Text(t)))
					}
					return Group(lis)
				}),
			),
		),
	)
}
