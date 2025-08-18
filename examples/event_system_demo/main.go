//go:build js && wasm

// Event System Demo
// Demonstrates the new robust event system with delegation, subscription management, and automatic cleanup

package main

import (
	"fmt"
	"syscall/js"
	"time"

	"app/golid"

	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

func main() {
	fmt.Println("🎯 Golid Event System Demo")
	fmt.Println("===========================")

	// Create the main application
	app := CreateEventSystemDemo()

	// Mount to DOM
	golid.Mount("app", app)

	// Keep the application running
	select {}
}

func CreateEventSystemDemo() Node {
	// Create signals for demo state
	clickCount, setClickCount := golid.CreateSignal(0)
	inputValue, setInputValue := golid.CreateSignal("")
	mousePosition, setMousePosition := golid.CreateSignal("0, 0")
	eventLog, setEventLog := golid.CreateSignal([]string{})

	// Helper function to add to event log
	addToLog := func(message string) {
		currentLog := eventLog()
		newLog := append(currentLog, fmt.Sprintf("[%s] %s", time.Now().Format("15:04:05"), message))
		if len(newLog) > 10 {
			newLog = newLog[len(newLog)-10:] // Keep only last 10 entries
		}
		setEventLog(newLog)
	}

	return Div(
		Class("event-demo"),
		Style(`
			font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
			max-width: 1200px;
			margin: 0 auto;
			padding: 20px;
			background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
			min-height: 100vh;
		`),

		H1(
			Text("🎯 Golid Event System Demo"),
			Style("color: white; text-align: center; margin-bottom: 30px;"),
		),

		// Performance Stats Section
		createStatsSection(),

		// Basic Event Handling Demo
		createBasicEventsSection(clickCount, setClickCount, addToLog),

		// Input Handling Demo
		createInputSection(inputValue, setInputValue, addToLog),

		// Mouse Events Demo
		createMouseEventsSection(mousePosition, setMousePosition, addToLog),

		// Advanced Features Demo
		createAdvancedFeaturesSection(addToLog),

		// Event Log
		createEventLogSection(eventLog),

		// Event System Controls
		createControlsSection(addToLog),
	)
}

func createStatsSection() Node {
	return Div(
		Class("stats-section"),
		Style(`
			background: rgba(255, 255, 255, 0.1);
			backdrop-filter: blur(10px);
			border-radius: 15px;
			padding: 20px;
			margin-bottom: 20px;
			color: white;
		`),

		H2(Text("📊 Event System Performance")),

		Div(
			ID("stats-display"),
			Style("font-family: monospace; font-size: 14px;"),
			Text("Loading stats..."),
		),

		// Update stats every second
		Script(Text(`
			setInterval(() => {
				if (window.golid && window.golid.GetEventSystemStats) {
					const stats = window.golid.GetEventSystemStats();
					const display = document.getElementById('stats-display');
					if (display) {
						display.innerHTML = '<pre>' + JSON.stringify(stats, null, 2) + '</pre>';
					}
				}
			}, 1000);
		`)),
	)
}

func createBasicEventsSection(clickCount func() int, setClickCount func(int), addToLog func(string)) Node {
	return Div(
		Class("basic-events"),
		Style(`
			background: rgba(255, 255, 255, 0.1);
			backdrop-filter: blur(10px);
			border-radius: 15px;
			padding: 20px;
			margin-bottom: 20px;
			color: white;
		`),

		H2(Text("🖱️ Basic Event Handling")),

		P(Text("Click counter demonstrates event delegation and automatic cleanup:")),

		Div(
			Style("display: flex; gap: 15px; align-items: center; margin: 15px 0;"),

			Button(
				golid.OnClickV2(func() {
					newCount := clickCount() + 1
					setClickCount(newCount)
					addToLog(fmt.Sprintf("Button clicked! Count: %d", newCount))
				}),
				Style(`
					background: #4CAF50;
					color: white;
					border: none;
					padding: 12px 24px;
					border-radius: 8px;
					cursor: pointer;
					font-size: 16px;
					transition: all 0.3s ease;
				`),
				Text("Click Me!"),
			),

			Div(
				Style(`
					background: rgba(0, 0, 0, 0.2);
					padding: 10px 20px;
					border-radius: 8px;
					font-size: 18px;
					font-weight: bold;
				`),
				golid.TextSignal(golid.CreateMemo(func() string {
					return fmt.Sprintf("Clicks: %d", clickCount())
				}, nil)),
			),
		),

		// Multiple buttons to test delegation
		Div(
			Style("display: flex; gap: 10px; flex-wrap: wrap;"),

			Button(
				golid.OnClickV2(func() {
					setClickCount(0)
					addToLog("Counter reset!")
				}),
				Style(`
					background: #f44336;
					color: white;
					border: none;
					padding: 8px 16px;
					border-radius: 6px;
					cursor: pointer;
				`),
				Text("Reset"),
			),

			Button(
				golid.OnClickV2(func() {
					newCount := clickCount() + 10
					setClickCount(newCount)
					addToLog(fmt.Sprintf("Added 10! Count: %d", newCount))
				}),
				Style(`
					background: #2196F3;
					color: white;
					border: none;
					padding: 8px 16px;
					border-radius: 6px;
					cursor: pointer;
				`),
				Text("+10"),
			),

			Button(
				golid.OnClickV2(func() {
					newCount := clickCount() - 5
					if newCount < 0 {
						newCount = 0
					}
					setClickCount(newCount)
					addToLog(fmt.Sprintf("Subtracted 5! Count: %d", newCount))
				}),
				Style(`
					background: #FF9800;
					color: white;
					border: none;
					padding: 8px 16px;
					border-radius: 6px;
					cursor: pointer;
				`),
				Text("-5"),
			),
		),
	)
}

func createInputSection(inputValue func() string, setInputValue func(string), addToLog func(string)) Node {
	return Div(
		Class("input-section"),
		Style(`
			background: rgba(255, 255, 255, 0.1);
			backdrop-filter: blur(10px);
			border-radius: 15px;
			padding: 20px;
			margin-bottom: 20px;
			color: white;
		`),

		H2(Text("⌨️ Input Event Handling")),

		P(Text("Input events with debouncing and reactive updates:")),

		Div(
			Style("margin: 15px 0;"),

			Label(
				Text("Type something: "),
				Style("display: block; margin-bottom: 8px; font-weight: bold;"),
			),

			Input(
				Type("text"),
				Placeholder("Start typing..."),
				golid.OnInputV2(func(value string) {
					setInputValue(value)
					addToLog(fmt.Sprintf("Input changed: '%s'", value))
				}),
				Style(`
					width: 100%;
					padding: 12px;
					border: 2px solid rgba(255, 255, 255, 0.3);
					border-radius: 8px;
					background: rgba(255, 255, 255, 0.1);
					color: white;
					font-size: 16px;
					backdrop-filter: blur(5px);
				`),
			),
		),

		Div(
			Style(`
				background: rgba(0, 0, 0, 0.2);
				padding: 15px;
				border-radius: 8px;
				margin-top: 15px;
			`),

			P(
				Style("margin: 0 0 10px 0; font-weight: bold;"),
				Text("Live Preview:"),
			),

			Div(
				Style(`
					font-size: 18px;
					font-family: monospace;
					background: rgba(255, 255, 255, 0.1);
					padding: 10px;
					border-radius: 6px;
					min-height: 24px;
				`),
				golid.TextSignal(golid.CreateMemo(func() string {
					value := inputValue()
					if value == "" {
						return "(empty)"
					}
					return fmt.Sprintf("'%s' (%d characters)", value, len(value))
				}, nil)),
			),
		),
	)
}

func createMouseEventsSection(mousePosition func() string, setMousePosition func(string), addToLog func(string)) Node {
	return Div(
		Class("mouse-section"),
		Style(`
			background: rgba(255, 255, 255, 0.1);
			backdrop-filter: blur(10px);
			border-radius: 15px;
			padding: 20px;
			margin-bottom: 20px;
			color: white;
		`),

		H2(Text("🖱️ Mouse Event Handling")),

		P(Text("Mouse events with throttling for performance:")),

		Div(
			golid.OnEventThrottled("mousemove", func(e js.Value) {
				x := e.Get("clientX").Int()
				y := e.Get("clientY").Int()
				setMousePosition(fmt.Sprintf("%d, %d", x, y))
			}, 50), // Throttle to 20fps

			golid.OnMouseEnterV2(func() {
				addToLog("Mouse entered tracking area")
			}),

			golid.OnMouseLeaveV2(func() {
				addToLog("Mouse left tracking area")
			}),

			Style(`
				background: rgba(255, 255, 255, 0.1);
				border: 2px dashed rgba(255, 255, 255, 0.5);
				border-radius: 12px;
				padding: 40px;
				text-align: center;
				cursor: crosshair;
				transition: all 0.3s ease;
				margin: 15px 0;
			`),

			H3(
				Text("Mouse Tracking Area"),
				Style("margin: 0 0 15px 0;"),
			),

			P(
				Text("Move your mouse in this area"),
				Style("margin: 0 0 15px 0; opacity: 0.8;"),
			),

			Div(
				Style(`
					background: rgba(0, 0, 0, 0.3);
					padding: 15px;
					border-radius: 8px;
					font-family: monospace;
					font-size: 18px;
				`),

				Text("Position: "),
				golid.TextSignal(mousePosition),
			),
		),
	)
}

func createAdvancedFeaturesSection(addToLog func(string)) Node {
	return Div(
		Class("advanced-section"),
		Style(`
			background: rgba(255, 255, 255, 0.1);
			backdrop-filter: blur(10px);
			border-radius: 15px;
			padding: 20px;
			margin-bottom: 20px;
			color: white;
		`),

		H2(Text("⚡ Advanced Event Features")),

		P(Text("Debouncing, throttling, and custom event options:")),

		Div(
			Style("display: grid; grid-template-columns: repeat(auto-fit, minmax(250px, 1fr)); gap: 15px; margin: 15px 0;"),

			// Debounced input
			Div(
				Style(`
					background: rgba(255, 255, 255, 0.1);
					padding: 15px;
					border-radius: 8px;
				`),

				H4(Text("Debounced Input"), Style("margin: 0 0 10px 0;")),

				Input(
					Type("text"),
					Placeholder("Debounced (500ms)"),
					golid.OnEventDebounced("input", func(e js.Value) {
						value := e.Get("target").Get("value").String()
						addToLog(fmt.Sprintf("Debounced input: '%s'", value))
					}, 500),
					Style(`
						width: 100%;
						padding: 8px;
						border: 1px solid rgba(255, 255, 255, 0.3);
						border-radius: 4px;
						background: rgba(255, 255, 255, 0.1);
						color: white;
					`),
				),
			),

			// Throttled button
			Div(
				Style(`
					background: rgba(255, 255, 255, 0.1);
					padding: 15px;
					border-radius: 8px;
				`),

				H4(Text("Throttled Button"), Style("margin: 0 0 10px 0;")),

				Button(
					golid.OnEventThrottled("click", func(e js.Value) {
						addToLog("Throttled click executed")
					}, 1000),
					Style(`
						width: 100%;
						background: #9C27B0;
						color: white;
						border: none;
						padding: 10px;
						border-radius: 4px;
						cursor: pointer;
					`),
					Text("Click rapidly (throttled)"),
				),
			),

			// Custom event options
			Div(
				Style(`
					background: rgba(255, 255, 255, 0.1);
					padding: 15px;
					border-radius: 8px;
				`),

				H4(Text("Custom Options"), Style("margin: 0 0 10px 0;")),

				Button(
					golid.OnEventWithOptions("click", func(e js.Value) {
						addToLog("High priority event executed")
					}, golid.EventOptions{
						Priority: golid.UserBlocking,
						Once:     false,
						Delegate: true,
					}),
					Style(`
						width: 100%;
						background: #E91E63;
						color: white;
						border: none;
						padding: 10px;
						border-radius: 4px;
						cursor: pointer;
					`),
					Text("High Priority Event"),
				),
			),
		),
	)
}

func createEventLogSection(eventLog func() []string) Node {
	return Div(
		Class("event-log"),
		Style(`
			background: rgba(0, 0, 0, 0.3);
			backdrop-filter: blur(10px);
			border-radius: 15px;
			padding: 20px;
			margin-bottom: 20px;
			color: white;
		`),

		H2(Text("📝 Event Log")),

		Div(
			Style(`
				background: rgba(0, 0, 0, 0.5);
				border-radius: 8px;
				padding: 15px;
				font-family: monospace;
				font-size: 14px;
				max-height: 300px;
				overflow-y: auto;
				border: 1px solid rgba(255, 255, 255, 0.2);
			`),

			golid.ForEachSignal(golid.CreateMemo(func() []string {
				log := eventLog()
				if len(log) == 0 {
					return []string{"No events yet..."}
				}
				return log
			}, nil), func(entry string) Node {
				return Div(
					Style("padding: 2px 0; border-bottom: 1px solid rgba(255, 255, 255, 0.1);"),
					Text(entry),
				)
			}),
		),
	)
}

func createControlsSection(addToLog func(string)) Node {
	return Div(
		Class("controls-section"),
		Style(`
			background: rgba(255, 255, 255, 0.1);
			backdrop-filter: blur(10px);
			border-radius: 15px;
			padding: 20px;
			color: white;
		`),

		H2(Text("🎛️ Event System Controls")),

		Div(
			Style("display: flex; gap: 15px; flex-wrap: wrap;"),

			Button(
				golid.OnClickV2(func() {
					stats := golid.GetEventSystemStats()
					addToLog(fmt.Sprintf("Event system stats: %d subscriptions", stats["subscriptions"]))
				}),
				Style(`
					background: #607D8B;
					color: white;
					border: none;
					padding: 12px 20px;
					border-radius: 8px;
					cursor: pointer;
				`),
				Text("Get Stats"),
			),

			Button(
				golid.OnClickV2(func() {
					// Simulate memory cleanup
					addToLog("Manual cleanup triggered")
					// Note: In a real app, you might call golid.CleanupEventSystem()
				}),
				Style(`
					background: #795548;
					color: white;
					border: none;
					padding: 12px 20px;
					border-radius: 8px;
					cursor: pointer;
				`),
				Text("Cleanup Events"),
			),

			Button(
				golid.OnClickV2(func() {
					addToLog("Performance test started...")
					// Create and immediately cleanup many event handlers
					start := time.Now()
					for i := 0; i < 100; i++ {
						cleanup := golid.Subscribe(js.Global().Get("document"), "click", func(e js.Value) {
							// Do nothing
						})
						cleanup() // Immediate cleanup
					}
					duration := time.Since(start)
					addToLog(fmt.Sprintf("Created/cleaned 100 handlers in %v", duration))
				}),
				Style(`
					background: #FF5722;
					color: white;
					border: none;
					padding: 12px 20px;
					border-radius: 8px;
					cursor: pointer;
				`),
				Text("Performance Test"),
			),
		),
	)
}
