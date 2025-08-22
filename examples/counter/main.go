//go:build js && wasm

package main

import (
	"fmt"
	"syscall/js"

	comps "github.com/ozanturksever/uiwgo/comps"
	reactivity "github.com/ozanturksever/uiwgo/reactivity"

	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

func main() {
	comps.Mount("app", func() Node { return CounterApp() })
	// Prevent exit
	select {}
}

func CounterApp() Node {
	count := reactivity.CreateSignal(0)
	double := reactivity.CreateMemo(func() int { return count.Get() * 2 })

	// Effect logging to console
	reactivity.CreateEffect(func() {
		fmt.Println("Count changed:", count.Get())
	})

	// Expose functions to window; we use onclick attributes to call them
	incFn := js.FuncOf(func(this js.Value, args []js.Value) any { count.Set(count.Get() + 1); return nil })
	decFn := js.FuncOf(func(this js.Value, args []js.Value) any { count.Set(count.Get() - 1); return nil })
	resetFn := js.FuncOf(func(this js.Value, args []js.Value) any { count.Set(0); return nil })
	js.Global().Set("incrementCounter", incFn)
	js.Global().Set("decrementCounter", decFn)
	js.Global().Set("resetCounter", resetFn)

	// Note: In a more advanced version, we'd register cleanup to release funcs.
	// For this MVP demo, functions live for the page lifetime.

	return Div(
		Style("font-family: Arial, sans-serif; max-width: 600px; margin: 50px auto; padding: 20px; background-color: #f5f5f5; min-height: 100vh;"),
		Div(
			Style("background: white; padding: 30px; border-radius: 10px; box-shadow: 0 2px 10px rgba(0,0,0,0.1); text-align: center;"),
			H1(Text("Counter Example")),
			P(Text("A simple counter demonstrating uiwgo's reactive signals")),

			Div(
				Style("font-size: 2em; font-weight: bold; color: #333; margin: 20px 0; padding: 20px; background-color: #f8f9fa; border-radius: 8px; border: 2px solid #e9ecef;"),
				comps.BindText(func() string { return fmt.Sprintf("Count: %d (double: %d)", count.Get(), double.Get()) }),
			),

			Div(
				Button(
					Style("font-size: 1.2em; padding: 10px 20px; margin: 0 10px; border: none; border-radius: 5px; cursor: pointer; background-color: #28a745; color: white; transition: background-color 0.2s;"),
					Text("+ Increment"),
					Attr("onclick", "window.incrementCounter()"),
				),
				Button(
					Style("font-size: 1.2em; padding: 10px 20px; margin: 0 10px; border: none; border-radius: 5px; cursor: pointer; background-color: #dc3545; color: white; transition: background-color 0.2s;"),
					Text("- Decrement"),
					Attr("onclick", "window.decrementCounter()"),
				),
				Br(),
				Button(
					Style("font-size: 1.2em; padding: 10px 20px; margin: 10px; border: none; border-radius: 5px; cursor: pointer; background-color: #6c757d; color: white; transition: background-color 0.2s;"),
					Text("Reset"),
					Attr("onclick", "window.resetCounter()"),
				),
			),

			Div(
				Style("margin-top: 20px; color: #666; font-style: italic;"),
				comps.BindText(func() string {
					c := count.Get()
					if c == 0 {
						return "Counter is at zero"
					} else if c > 0 {
						return fmt.Sprintf("Counter is positive (+%d)", c)
					}
					return fmt.Sprintf("Counter is negative (%d)", c)
				}),
			),
		),
	)
}
