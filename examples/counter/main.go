package main

import (
	"app/golid"
	"fmt"

	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

func main() {
	fmt.Println("!!!")
	_, cleanup := golid.CreateRoot(func() interface{} {
		app := CounterApp()
		golid.Render(app)
		return nil
	})

	defer cleanup()
	golid.Run()
}

func CounterApp() Node {
	return Div(
		Style("font-family: Arial, sans-serif; max-width: 600px; margin: 50px auto; padding: 20px; background-color: #f5f5f5; min-height: 100vh;"),
		Div(
			Style("background: white; padding: 30px; border-radius: 10px; box-shadow: 0 2px 10px rgba(0,0,0,0.1); text-align: center;"),
			H1(Text("Counter Example")),
			P(Text("A simple counter demonstrating Golid's reactive signals")),
			CounterComponent(),
		),
	)
}

func CounterComponent() Node {
	// Create a reactive signal to hold the counter value
	count, setCount := golid.CreateSignal(0)

	return Div(
		// Display the current count value reactively
		Div(
			Style("font-size: 2em; font-weight: bold; color: #333; margin: 20px 0; padding: 20px; background-color: #f8f9fa; border-radius: 8px; border: 2px solid #e9ecef;"),
			golid.Bind(func() Node {
				return Text(fmt.Sprintf("Count: %d", count()))
			}),
		),

		// Increment button
		Button(
			Style("font-size: 1.2em; padding: 10px 20px; margin: 0 10px; border: none; border-radius: 5px; cursor: pointer; background-color: #28a745; color: white; transition: background-color 0.2s;"),
			Text("+ Increment"),
			golid.OnClickV2(func() {
				setCount(count() + 1)
			}),
		),

		// Decrement button
		Button(
			Style("font-size: 1.2em; padding: 10px 20px; margin: 0 10px; border: none; border-radius: 5px; cursor: pointer; background-color: #dc3545; color: white; transition: background-color 0.2s;"),
			Text("- Decrement"),
			golid.OnClickV2(func() {
				setCount(count() - 1)
			}),
		),

		Br(),

		// Reset button
		Button(
			Style("font-size: 1.2em; padding: 10px 20px; margin: 10px; border: none; border-radius: 5px; cursor: pointer; background-color: #6c757d; color: white; transition: background-color 0.2s;"),
			Text("Reset"),
			golid.OnClickV2(func() {
				setCount(0)
			}),
		),

		// Show additional info reactively
		Div(
			Style("margin-top: 20px; color: #666; font-style: italic;"),
			golid.Bind(func() Node {
				currentCount := count()
				if currentCount == 0 {
					return Text("Counter is at zero")
				} else if currentCount > 0 {
					return Text(fmt.Sprintf("Counter is positive (+%d)", currentCount))
				} else {
					return Text(fmt.Sprintf("Counter is negative (%d)", currentCount))
				}
			}),
		),
	)
}
