package main

import (
	"fmt"

	"github.com/ozanturksever/uiwgo/comps"
	"github.com/ozanturksever/uiwgo/reactivity"
	g "maragu.dev/gomponents"
	h "maragu.dev/gomponents/html"
)

// Functional Counter component
func CounterComponent() g.Node {
	count := reactivity.CreateSignal(0)
	return h.Div(
		h.H2(g.Text("Counter Component")),
		h.P(g.Text(fmt.Sprintf("Count: %d", count.Get()))),
		h.P(g.Text("Interactive buttons require JavaScript integration")),
		g.Attr("id", "counter-component"),
	)
}

// Functional Greeting component with props
func GreetingComponent(name string) g.Node {
	return h.Div(
		h.H2(g.Text("Greeting Component")),
		h.P(g.Text(fmt.Sprintf("Hello, %s!", name))),
	)
}

// Functional Todo component
func TodoComponent() g.Node {
	return h.Div(
		h.H2(g.Text("Todo Component")),
		h.P(g.Text("Todo list functionality requires JavaScript integration")),
		h.Ul(
			h.Li(g.Text("Static todo item 1")),
			h.Li(g.Text("Static todo item 2")),
		),
		g.Attr("id", "todo-component"),
	)
}

// Functional Header component using Fragment
func HeaderComponent(title string) g.Node {
	return comps.Fragment(
		h.H1(g.Text(title)),
		h.Hr(),
	)
}

// Main functional app component
func AppComponent() g.Node {
	return h.Div(
		h.H1(g.Text("Component System Demo")),
		h.Hr(),

		// Header component using Fragment
		HeaderComponent("Welcome to Component Demo"),

		// Counter component
		h.H3(g.Text("1. Counter Component (Stateful)")),
		CounterComponent(),

		// Greeting component with props
		h.H3(g.Text("2. Greeting Component (With Props)")),
		GreetingComponent("World"),

		// Todo component
		h.H3(g.Text("3. Todo Component (List Management)")),
		TodoComponent(),

		// Fragment example
		h.H3(g.Text("4. Fragment Example")),
		comps.Fragment(
			h.P(g.Text("This is a paragraph")),
			h.P(g.Text("This is another paragraph")),
			h.P(g.Text("All rendered without a wrapper div")),
		),

		// Memo example
		h.H3(g.Text("5. Memoization Example")),
		h.Button(
			g.Attr("onclick", "location.reload()"),
			g.Text("Reload to test memoization"),
		),
		h.P(g.Text("Check console for mount/unmount logs to see memoization in action")),

		g.Attr("style", "font-family: Arial, sans-serif; padding: 20px;"),
	)
}

func main() {
	// Mount the app
	comps.Mount("app", func() g.Node {
		return AppComponent()
	})
}
