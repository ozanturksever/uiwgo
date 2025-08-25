package main

import (
	"fmt"

	"github.com/ozanturksever/uiwgo/comps"
	"github.com/ozanturksever/uiwgo/reactivity"
	g "maragu.dev/gomponents"
	h "maragu.dev/gomponents/html"
)

// Simple component with state
type Counter struct {
	count reactivity.Signal[int]
}

func NewCounter() *Counter {
	return &Counter{
		count: reactivity.CreateSignal(0),
	}
}

func (c *Counter) OnMount() {
	fmt.Println("Counter mounted")
}

func (c *Counter) OnUnMount() {
	fmt.Println("Counter unmounted")
}

func (c *Counter) Render() g.Node {
	return h.Div(
		h.H2(g.Text("Counter Component")),
		h.P(g.Text(fmt.Sprintf("Count: %d", c.count.Get()))),
		h.P(g.Text("Interactive buttons require JavaScript integration")),
		g.Attr("id", fmt.Sprintf("counter-%p", c)),
	)
}

// Component with props
type GreetingProps struct {
	Name string
}

type Greeting struct{}

func (gr *Greeting) OnMount() {
	fmt.Println("Greeting mounted")
}

func (gr *Greeting) OnUnMount() {
	fmt.Println("Greeting unmounted")
}

func (gr *Greeting) Render(props GreetingProps) g.Node {
	return h.Div(
		h.H2(g.Text("Greeting Component")),
		h.P(g.Text(fmt.Sprintf("Hello, %s!", props.Name))),
	)
}

// Stateful component example
type Todo struct {
	items reactivity.Signal[[]string]
}

func NewTodo() *Todo {
	return &Todo{
		items: reactivity.CreateSignal([]string{}),
	}
}

func (t *Todo) OnMount() {
	fmt.Println("Todo mounted")
}

func (t *Todo) OnUnMount() {
	fmt.Println("Todo unmounted")
}

func (t *Todo) Render() g.Node {
	return h.Div(
		h.H2(g.Text("Todo Component")),
		h.P(g.Text("Todo list functionality requires JavaScript integration")),
		h.Ul(
			h.Li(g.Text("Static todo item 1")),
			h.Li(g.Text("Static todo item 2")),
		),
		g.Attr("id", fmt.Sprintf("todo-%p", t)),
	)
}

// Component using Fragment
type Header struct {
	Title string
}

func (hd *Header) OnMount() {
	fmt.Println("Header mounted")
}

func (hd *Header) OnUnMount() {
	fmt.Println("Header unmounted")
}

func (hd *Header) Render() g.Node {
	return comps.Fragment(
		h.H1(g.Text(hd.Title)),
		h.Hr(),
	)
}

// Main app component
type App struct{}

func (a *App) OnMount() {
	fmt.Println("App mounted")
}

func (a *App) OnUnMount() {
	fmt.Println("App unmounted")
}

func (a *App) Render() g.Node {
	counter := NewCounter()
	todo := NewTodo()

	return h.Div(
		h.H1(g.Text("Component System Demo")),
		h.Hr(),

		// Header component using Fragment
		comps.ComponentFactory(&Header{Title: "Welcome to Component Demo"}),

		// Counter component
		h.H3(g.Text("1. Counter Component (Stateful)")),
		comps.ComponentFactory(counter),

		// Greeting component with props
		h.H3(g.Text("2. Greeting Component (With Props)")),
		comps.ComponentFactoryWithProps(&Greeting{}, GreetingProps{Name: "World"}),

		// Todo component
		h.H3(g.Text("3. Todo Component (List Management)")),
		comps.ComponentFactory(todo),

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
		return comps.ComponentFactory(&App{})
	})
}
