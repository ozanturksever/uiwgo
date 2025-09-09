//go:build js && wasm

package main

import (
	"fmt"
	"time"

	"github.com/ozanturksever/logutil"
	comps "github.com/ozanturksever/uiwgo/comps"
	dom "github.com/ozanturksever/uiwgo/dom"
	reactivity "github.com/ozanturksever/uiwgo/reactivity"
	g "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

type Todo struct {
	ID   string
	Text string
	Done bool
}

type AppState struct {
	showConditional reactivity.Signal[bool]
	todos           reactivity.Signal[[]Todo]
	numbers         reactivity.Signal[[]int]
	status          reactivity.Signal[string]
	currentView     reactivity.Signal[func() g.Node]
	showModal       reactivity.Signal[bool]
	expensiveData   reactivity.Signal[[]string]
	lazyLoaded      reactivity.Signal[bool]
	hasError        reactivity.Signal[bool]
}

func NewAppState() *AppState {
	return &AppState{
		showConditional: reactivity.CreateSignal(true),
		todos: reactivity.CreateSignal([]Todo{
			{ID: "1", Text: "Learn UIwGo helpers", Done: false},
			{ID: "2", Text: "Build awesome apps", Done: true},
			{ID: "3", Text: "Share with community", Done: false},
		}),
		numbers: reactivity.CreateSignal([]int{10, 20, 30, 40, 50}),
		status: reactivity.CreateSignal("loading"),
		currentView: reactivity.CreateSignal(func() g.Node {
			return P(g.Text("🏠 Home View - Welcome to the helpers demo!"))
		}),
		showModal: reactivity.CreateSignal(false),
		expensiveData: reactivity.CreateSignal([]string{"Item 1", "Item 2", "Item 3"}),
		lazyLoaded: reactivity.CreateSignal(false),
		hasError: reactivity.CreateSignal(false),
	}
}

func (app *AppState) ShowDemo() g.Node {
	return Div(
		Class("demo-section"),
		H3(g.Text("Show Helper - Conditional Rendering")),
		Button(
			g.Text("Toggle Visibility"),
			dom.OnClickInline(func(el dom.Element) {
				app.showConditional.Set(!app.showConditional.Get())
			}),
		),
		comps.Show(comps.ShowProps{
			When: app.showConditional,
			Children: Div(
				Class("conditional-content"),
				Style("background: #e8f5e8; padding: 10px; margin: 10px 0; border-radius: 4px;"),
				P(g.Text("✅ This content is conditionally visible!")),
			),
		}),
	)
}

func (app *AppState) ForDemo() g.Node {
	return Div(
		Class("demo-section"),
		H3(g.Text("For Helper - List Rendering with Keys")),
		Button(
			g.Text("Add Todo"),
			dom.OnClickInline(func(el dom.Element) {
				current := app.todos.Get()
				newTodo := Todo{
					ID:   fmt.Sprintf("%d", time.Now().UnixNano()),
					Text: fmt.Sprintf("New todo #%d", len(current)+1),
					Done: false,
				}
				app.todos.Set(append(current, newTodo))
			}),
		),
		Ul(
			Style("list-style: none; padding: 0;"),
			comps.For(comps.ForProps[Todo]{
				Items: app.todos,
				Key: func(todo Todo) string { return todo.ID },
				Children: func(todo Todo, index int) g.Node {
					return Li(
						Style("background: #f5f5f5; margin: 5px 0; padding: 10px; border-radius: 4px; display: flex; align-items: center;"),
						Input(
							Type("checkbox"),
							// Conditional checked attribute handled by signal
							Style("margin-right: 10px;"),
							dom.OnChangeInline(func(el dom.Element) {
					current := app.todos.Get()
					for i, t := range current {
						if t.ID == todo.ID {
							current[i].Done = !current[i].Done
							break
						}
					}
					app.todos.Set(current)
				}),
						),
						Span(
							g.Text(fmt.Sprintf("%d. %s", index+1, todo.Text)),
							// Conditional styling handled by signal
						),
					)
				},
			}),
		),
	)
}

func (app *AppState) IndexDemo() g.Node {
	return Div(
		Class("demo-section"),
		H3(g.Text("Index Helper - Index-based Reconciliation")),
		Button(
			g.Text("Shuffle Numbers"),
			dom.OnClickInline(func(el dom.Element) {
				current := app.numbers.Get()
				// Simple shuffle
				for i := len(current) - 1; i > 0; i-- {
					j := int(time.Now().UnixNano()) % (i + 1)
					current[i], current[j] = current[j], current[i]
				}
				app.numbers.Set(current)
			}),
		),
		Ul(
			Style("list-style: none; padding: 0;"),
			comps.Index(comps.IndexProps[int]{
				Items: app.numbers,
				Children: func(getItem func() int, index int) g.Node {
					return Li(
						Style("background: #e8f4fd; margin: 5px 0; padding: 10px; border-radius: 4px;"),
						g.Text(fmt.Sprintf("Index %d: Value %d", index, getItem())),
					)
				},
			}),
		),
	)
}

func (app *AppState) SwitchDemo() g.Node {
	return Div(
		Class("demo-section"),
		H3(g.Text("Switch/Match Helper - Conditional Branching")),
		Div(
			Style("margin: 10px 0;"),
			Button(
				g.Text("Loading"),
				dom.OnClickInline(func(el dom.Element) {
					app.status.Set("loading")
				}),
			),
			Button(
				g.Text("Success"),
				Style("margin-left: 10px;"),
				dom.OnClickInline(func(el dom.Element) {
					app.status.Set("success")
				}),
			),
			Button(
				g.Text("Error"),
				Style("margin-left: 10px;"),
				dom.OnClickInline(func(el dom.Element) {
					app.status.Set("error")
				}),
			),
			Button(
				g.Text("Unknown"),
				Style("margin-left: 10px;"),
				dom.OnClickInline(func(el dom.Element) {
					app.status.Set("unknown")
				}),
			),
		),
		Div(
			Style("padding: 15px; border-radius: 4px; margin: 10px 0;"),
			comps.Switch(comps.SwitchProps{
				When: app.status,
				Fallback: P(
					Style("color: #666; font-style: italic;"),
					g.Text("❓ Unknown status"),
				),
				Children: []g.Node{
					comps.Match(comps.MatchProps{
						When: "loading",
						Children: P(
							Style("color: #0066cc;"),
							g.Text("⏳ Loading..."),
						),
					}),
					comps.Match(comps.MatchProps{
						When: "success",
						Children: P(
							Style("color: #00aa00;"),
							g.Text("✅ Success! Operation completed."),
						),
					}),
					comps.Match(comps.MatchProps{
						When: "error",
						Children: P(
							Style("color: #cc0000;"),
							g.Text("❌ Error occurred! Please try again."),
						),
					}),
				},
			}),
		),
	)
}

func (app *AppState) DynamicDemo() g.Node {
	return Div(
		Class("demo-section"),
		H3(g.Text("Dynamic Helper - Dynamic Component Rendering")),
		Div(
			Style("margin: 10px 0;"),
			Button(
				g.Text("Home View"),
				dom.OnClickInline(func(el dom.Element) {
					app.currentView.Set(func() g.Node {
						return P(
							Style("background: #e8f5e8; padding: 15px; border-radius: 4px;"),
							g.Text("🏠 Home View - Welcome to the helpers demo!"),
						)
					})
				}),
			),
			Button(
				g.Text("Profile View"),
				Style("margin-left: 10px;"),
				dom.OnClickInline(func(el dom.Element) {
					app.currentView.Set(func() g.Node {
						return P(
							Style("background: #fff3cd; padding: 15px; border-radius: 4px;"),
							g.Text("👤 Profile View - User settings and preferences"),
						)
					})
				}),
			),
			Button(
				g.Text("Settings View"),
				Style("margin-left: 10px;"),
				dom.OnClickInline(func(el dom.Element) {
					app.currentView.Set(func() g.Node {
						return P(
							Style("background: #f8d7da; padding: 15px; border-radius: 4px;"),
							g.Text("⚙️ Settings View - Application configuration"),
						)
					})
				}),
			),
		),
		Div(
			Style("margin: 10px 0;"),
			comps.Dynamic(comps.DynamicProps{
				Component: app.currentView,
			}),
		),
	)
}

func (app *AppState) FragmentDemo() g.Node {
	return Div(
		Class("demo-section"),
		H3(g.Text("Fragment Helper - Grouping Without Wrapper")),
		P(g.Text("The content below uses Fragment to group elements without a wrapper:")),
		Div(
			Style("border: 2px dashed #ccc; padding: 10px; margin: 10px 0;"),
			comps.Fragment(
				H4(g.Text("Fragment Content")),
				P(g.Text("This paragraph is inside a fragment.")),
				P(g.Text("This is another paragraph in the same fragment.")),
				Button(g.Text("Fragment Button")),
			),
		),
	)
}

func (app *AppState) PortalDemo() g.Node {
	return Div(
		Class("demo-section"),
		H3(g.Text("Portal Helper - Render to Different Location")),
		Button(
			g.Text("Open Modal"),
			dom.OnClickInline(func(el dom.Element) {
				app.showModal.Set(true)
			}),
		),
		comps.Show(comps.ShowProps{
			When: app.showModal,
			Children: comps.Portal("body", Div(
				Style("position: fixed; top: 0; left: 0; width: 100%; height: 100%; background: rgba(0,0,0,0.5); display: flex; align-items: center; justify-content: center; z-index: 1000;"),
				Div(
					Style("background: white; padding: 20px; border-radius: 8px; max-width: 400px; box-shadow: 0 4px 6px rgba(0,0,0,0.1);"),
					H3(g.Text("Modal Title")),
					P(g.Text("This modal is rendered using Portal helper. It's rendered directly to the body element, outside the normal component tree.")),
					Button(
						g.Text("Close Modal"),
						Style("background: #007bff; color: white; border: none; padding: 8px 16px; border-radius: 4px; cursor: pointer;"),
						dom.OnClickInline(func(el dom.Element) {
							app.showModal.Set(false)
						}),
					),
				),
			)),
		}),
	)
}

func (app *AppState) MemoDemo() g.Node {
	return Div(
		Class("demo-section"),
		H3(g.Text("Memo Helper - Memoized Rendering")),
		Button(
			g.Text("Add Item"),
			dom.OnClickInline(func(el dom.Element) {
				current := app.expensiveData.Get()
				newItem := fmt.Sprintf("Item %d", len(current)+1)
				app.expensiveData.Set(append(current, newItem))
				logutil.Log("Added item, memo will re-render")
			}),
		),
		Button(
				g.Text("Force Re-render (no memo change)"),
				Style("margin-left: 10px;"),
				dom.OnClickInline(func(el dom.Element) {
					// This won't cause memo to re-render since data hasn't changed
					logutil.Log("Button clicked, but memo won't re-render")
				}),
			),
		Div(
			Style("margin: 10px 0;"),
			comps.Memo(func() g.Node {
				logutil.Log("Memo component rendering - this should only log when data changes")
			data := app.expensiveData.Get()
			var items []g.Node
			for _, item := range data {
				items = append(items, Li(
					Style("background: #f0f8ff; margin: 2px 0; padding: 5px; border-radius: 3px;"),
					g.Text(item),
				))
			}
			return Div(
				P(g.Text(fmt.Sprintf("Memoized list (%d items):", len(data)))),
				Ul(append([]g.Node{Style("list-style: none; padding: 0;")}, items...)...),
			)
			}, app.expensiveData.Get()),
		),
	)
}

func (app *AppState) LazyDemo() g.Node {
	return Div(
		Class("demo-section"),
		H3(g.Text("Lazy Helper - Lazy Loading")),
		Button(
			g.Text("Load Lazy Component"),
			dom.OnClickInline(func(el dom.Element) {
				app.lazyLoaded.Set(true)
				logutil.Log("Lazy component will be loaded")
			}),
		),
		comps.Show(comps.ShowProps{
			When: app.lazyLoaded,
			Children: comps.Lazy(func() func() g.Node {
				logutil.Log("Lazy loader function called - simulating module loading")
				// Simulate loading delay
				return func() g.Node {
					logutil.Log("Lazy component rendered")
					return Div(
						Style("background: #e6f3ff; padding: 15px; border-radius: 4px; margin: 10px 0;"),
						H4(g.Text("🚀 Lazy Loaded Component")),
						P(g.Text("This component was loaded on demand using the Lazy helper.")),
						P(g.Text("In a real application, this could be loaded from a separate module or bundle.")),
					)
				}
			}),
		}),
	)
}

func RiskyComponent() g.Node {
	return Div(
		P(g.Text("This component might throw an error...")),
		Button(
			g.Text("Trigger Error"),
			Style("background: #dc3545; color: white; border: none; padding: 8px 16px; border-radius: 4px; cursor: pointer;"),
			dom.OnClickInline(func(el dom.Element) {
				// Simulate an error
				panic("Simulated error for ErrorBoundary demo")
			}),
		),
	)
}

func (app *AppState) ErrorBoundaryDemo() g.Node {
	return Div(
		Class("demo-section"),
		H3(g.Text("ErrorBoundary Helper - Error Handling")),
		Button(
			g.Text("Toggle Error Component"),
			dom.OnClickInline(func(el dom.Element) {
				app.hasError.Set(!app.hasError.Get())
			}),
		),
		comps.Show(comps.ShowProps{
			When: app.hasError,
			Children: comps.ErrorBoundary(comps.ErrorBoundaryProps{
				Fallback: func(err error) g.Node {
					return Div(
						Style("background: #f8d7da; border: 1px solid #f5c6cb; color: #721c24; padding: 15px; border-radius: 4px; margin: 10px 0;"),
						H4(g.Text("🚨 Error Boundary Caught an Error")),
						P(g.Text(fmt.Sprintf("Error: %s", err.Error()))),
						Button(
							g.Text("Retry"),
							Style("background: #007bff; color: white; border: none; padding: 8px 16px; border-radius: 4px; cursor: pointer;"),
							dom.OnClickInline(func(el dom.Element) {
								app.hasError.Set(false)
								logutil.Log("Retrying after error")
							}),
						),
					)
				},
				Children: RiskyComponent(),
			}),
		}),
	)
}

func HelpersDemo() g.Node {
	app := NewAppState()

	return Div(
		Style("font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; max-width: 1200px; margin: 0 auto; padding: 20px;"),
		H1(
			Style("color: #333; text-align: center; margin-bottom: 30px;"),
			g.Text("🛠️ UIwGo Helpers Demo"),
		),
		P(
			Style("text-align: center; color: #666; margin-bottom: 40px;"),
			g.Text("This demo showcases all the helper functions available in UIwGo for building dynamic, reactive user interfaces."),
		),

		// Add some global styles for demo sections
		Style(`
			.demo-section {
				margin: 30px 0;
				padding: 20px;
				border: 1px solid #e0e0e0;
				border-radius: 8px;
				background: #fafafa;
			}
			.demo-section h3 {
				margin-top: 0;
				color: #333;
				border-bottom: 2px solid #007bff;
				padding-bottom: 10px;
			}
			button {
				background: #007bff;
				color: white;
				border: none;
				padding: 8px 16px;
				border-radius: 4px;
				cursor: pointer;
				margin: 2px;
			}
			button:hover {
				background: #0056b3;
			}
		`),

		app.ShowDemo(),
		app.ForDemo(),
		app.IndexDemo(),
		app.SwitchDemo(),
		app.DynamicDemo(),
		app.FragmentDemo(),
		app.PortalDemo(),
		app.MemoDemo(),
		app.LazyDemo(),
		app.ErrorBoundaryDemo(),

		Div(
			Style("margin-top: 50px; padding: 20px; background: #e8f5e8; border-radius: 8px; text-align: center;"),
			H3(g.Text("🎉 Demo Complete!")),
			P(g.Text("You've seen all the UIwGo helper functions in action. Check the browser console for additional logging output.")),
		),
	)
}

func main() {
	logutil.Log("UIwGo Helpers Demo starting...")

	// Mount the application
	comps.Mount("app", HelpersDemo)

	logutil.Log("UIwGo Helpers Demo mounted successfully!")

	// Keep the program running
	select {}
}