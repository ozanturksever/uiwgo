package spec

import (
	"fmt"
	"github.com/ozanturksever/uiwgo"

	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

func main() {
	uiwgo.Render(CounterApp())
	uiwgo.Run()
}

func CounterApp() Node {
	return Div(
		Style("font-family: Arial, sans-serif; max-width: 600px; margin: 50px auto; padding: 20px; background-color: #f5f5f5; min-height: 100vh;"),
		Div(
			Style("background: white; padding: 30px; border-radius: 10px; box-shadow: 0 2px 10px rgba(0,0,0,0.1); text-align: center;"),
			AppHeader("Counter Example", "A simple counter demonstrating uiwgo's reactive signals"),
			CounterComponent(),
		),
	)
}

func CounterComponent() Node {
	// Create a reactive signal to hold the counter value
	count := uiwgo.NewSignal(0)

	return Div(
		// Lifecycle hooks
		uiwgo.OnMount(func() {
			fmt.Println("CounterComponent mounted")
		}),
		// Effect runs whenever accessed signals inside change (e.g., count)
		uiwgo.Effect(func() {
			fmt.Println("CounterComponent count changed:", count.Get())
		}),
		uiwgo.OnCleanup(func() {
			fmt.Println("CounterComponent unmounted")
		}),

		// Composition: child components for display and controls
		CounterDisplay(func() int { return count.Get() }),
		CounterControls(
			func() { count.Set(count.Get() + 1) },
			func() { count.Set(count.Get() - 1) },
			func() { count.Set(0) },
		),

		// Show additional info reactively
		Div(
			Style("margin-top: 20px; color: #666; font-style: italic;"),
			uiwgo.BindText(func() string {
				currentCount := count.Get()
				if currentCount == 0 {
					return "Counter is at zero"
				} else if currentCount > 0 {
					return fmt.Sprintf("Counter is positive (+%d)", currentCount)
				} else {
					return fmt.Sprintf("Counter is negative (%d)", currentCount)
				}
			}),
		),
	)
}

// Composed header component
func AppHeader(title, subtitle string) Node {
	return Div(
		H1(Text(title)),
		P(Text(subtitle)),
	)
}

// Display-only child component that renders the reactive count
func CounterDisplay(getCount func() int) Node {
	return Div(
		Style("font-size: 2em; font-weight: bold; color: #333; margin: 20px 0; padding: 20px; background-color: #f8f9fa; border-radius: 8px; border: 2px solid #e9ecef;"),
		uiwgo.BindText(func() string {
			return fmt.Sprintf("Count: %d", getCount())
		}),
	)
}

// Controls child component; parent injects behavior via callbacks
func CounterControls(onInc, onDec, onReset func()) Node {
	return Div(
		// Increment button
		Button(
			Style("font-size: 1.2em; padding: 10px 20px; margin: 0 10px; border: none; border-radius: 5px; cursor: pointer; background-color: #28a745; color: white; transition: background-color 0.2s;"),
			Text("+ Increment"),
			uiwgo.OnClick(onInc),
		),

		// Decrement button
		Button(
			Style("font-size: 1.2em; padding: 10px 20px; margin: 0 10px; border: none; border-radius: 5px; cursor: pointer; background-color: #dc3545; color: white; transition: background-color 0.2s;"),
			Text("- Decrement"),
			uiwgo.OnClick(onDec),
		),

		Br(),

		// Reset button
		Button(
			Style("font-size: 1.2em; padding: 10px 20px; margin: 10px; border: none; border-radius: 5px; cursor: pointer; background-color: #6c757d; color: white; transition: background-color 0.2s;"),
			Text("Reset"),
			uiwgo.OnClick(onReset),
		),
	)
}
