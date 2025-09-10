//go:build js && wasm

package main

import (
	"fmt"
	"time"

	"github.com/ozanturksever/logutil"
	"github.com/ozanturksever/uiwgo/action"
	"github.com/ozanturksever/uiwgo/comps"
	"github.com/ozanturksever/uiwgo/dom"
	"github.com/ozanturksever/uiwgo/reactivity"
	g "maragu.dev/gomponents"
	"maragu.dev/gomponents/html"
)

// Define action types
var (
	IncrementAction    = action.DefineAction[int]("counter.increment")
	DecrementAction    = action.DefineAction[int]("counter.decrement")
	ErrorAction        = action.DefineAction[string]("demo.error")
	AnalyticsAction    = action.DefineAction[string]("analytics.event")
	ToggleLoggerAction = action.DefineAction[bool]("observability.toggle_logger")
)

func main() {
	// Create a bus instance
	bus := action.New()

	// Set up observability features
	setupObservability(bus)

	// Mount the component
	comps.Mount("app", func() g.Node {
		return ObservabilityDemoComponent(bus)
	})

	// Prevent exit
	select {}
}

func setupObservability(bus action.Bus) {
	// Enable debug ring buffer with size 10
	action.EnableDebugRingBuffer(bus, 10)

	// Set up enhanced error handler
	bus.OnError(func(ctx action.Context, err error, recovered any) {
		logutil.Logf("🚨 ERROR: %v (recovered: %v) TraceID: %s", err, recovered, ctx.TraceID)
	})

	// Set up analytics tap with filter
	analyticsEvents := reactivity.CreateSignal([]string{})
	analyticsTap := action.NewAnalyticsTap(bus, func(event action.AnalyticsEvent) {
		current := analyticsEvents.Get()
		entry := fmt.Sprintf("[%s] %s (TraceID: %s)",
			event.Timestamp.Format("15:04:05"),
			event.ActionType,
			event.TraceID)

		// Keep only last 5 events
		if len(current) >= 5 {
			current = current[1:]
		}
		analyticsEvents.Set(append(current, entry))
	}, action.WithAnalyticsFilter(func(act any) bool {
		// Only track certain action types for analytics
		if actionData, ok := act.(action.Action[string]); ok {
			return actionData.Type == IncrementAction.Name ||
				actionData.Type == DecrementAction.Name ||
				actionData.Type == AnalyticsAction.Name
		}
		return false
	}))

	// Store analytics tap for cleanup (in real app you'd dispose this on unmount)
	_ = analyticsTap
}

// ObservabilityDemoComponent demonstrates enhanced observability features
func ObservabilityDemoComponent(bus action.Bus) g.Node {
	// Create signals for state
	count := reactivity.CreateSignal(0)
	devLoggerEnabled := reactivity.CreateSignal(false)
	logEntries := reactivity.CreateSignal([]string{})
	errorCount := reactivity.CreateSignal(0)

	// Use OnAction to register action handlers with lifecycle management
	action.OnAction(bus, IncrementAction, func(ctx action.Context, payload int) {
		count.Set(count.Get() + payload)
		logutil.Logf("✅ Increment executed: %d (TraceID: %s)", payload, ctx.TraceID)
	})

	action.OnAction(bus, DecrementAction, func(ctx action.Context, payload int) {
		count.Set(count.Get() - payload)
		logutil.Logf("✅ Decrement executed: %d (TraceID: %s)", payload, ctx.TraceID)
	})

	action.OnAction(bus, ErrorAction, func(ctx action.Context, payload string) {
		errorCount.Set(errorCount.Get() + 1)
		logutil.Logf("💥 About to panic with: %s (TraceID: %s)", payload, ctx.TraceID)
		panic("Demo error: " + payload)
	})

	action.OnAction(bus, ToggleLoggerAction, func(ctx action.Context, enabled bool) {
		devLoggerEnabled.Set(enabled)
		if enabled {
			// Enable dev logger
			action.EnableDevLogger(bus, func(entry action.DevLogEntry) {
				current := logEntries.Get()
				logLine := fmt.Sprintf("[%s] %s -> %d subscribers (%v)",
					entry.Timestamp.Format("15:04:05"),
					entry.ActionType,
					entry.SubscriberCount,
					entry.Duration)

				// Keep only last 8 log entries
				if len(current) >= 8 {
					current = current[1:]
				}
				logEntries.Set(append(current, logLine))
			})
			logutil.Log("🔍 Dev logger enabled")
		} else {
			action.DisableDevLogger(bus)
			logutil.Log("🔍 Dev logger disabled")
		}
	})

	return html.Div(
		g.Attr("style", "font-family: monospace; max-width: 800px; margin: 20px; line-height: 1.5;"),

		// Header
		html.H1(g.Text("🔍 Action System Observability Demo")),
		html.P(g.Text("This demo showcases E8: Observability, Errors, and Tracing features")),

		// Counter Section
		html.Div(
			g.Attr("style", "border: 1px solid #ddd; padding: 15px; margin: 10px 0; border-radius: 5px;"),
			html.H3(g.Text("🔢 Counter")),
			html.P(
				comps.BindText(func() string {
					return fmt.Sprintf("Count: %d", count.Get())
				}),
			),
			html.Button(
				html.ID("inc-btn"),
				g.Text("Increment"),
				g.Attr("style", "margin: 5px; padding: 8px 16px;"),
				dom.OnClick("inc-btn", func() {
					bus.Dispatch(action.Action[string]{
						Type:    IncrementAction.Name,
						Payload: "1",
						TraceID: fmt.Sprintf("trace-%d", time.Now().UnixNano()),
						Source:  "counter-ui",
						Meta:    map[string]any{"user": "demo", "component": "counter"},
					})
				}),
			),
			html.Button(
				html.ID("dec-btn"),
				g.Text("Decrement"),
				g.Attr("style", "margin: 5px; padding: 8px 16px;"),
				dom.OnClick("dec-btn", func() {
					bus.Dispatch(action.Action[string]{
						Type:    DecrementAction.Name,
						Payload: "1",
						TraceID: fmt.Sprintf("trace-%d", time.Now().UnixNano()),
						Source:  "counter-ui",
						Meta:    map[string]any{"user": "demo", "component": "counter"},
					})
				}),
			),
		),

		// Observability Controls
		html.Div(
			g.Attr("style", "border: 1px solid #ddd; padding: 15px; margin: 10px 0; border-radius: 5px;"),
			html.H3(g.Text("🔍 Observability Controls")),
			html.Label(
				html.Input(
					html.Type("checkbox"),
					html.ID("logger-toggle"),
					dom.OnClick("logger-toggle", func() {
						enabled := !devLoggerEnabled.Get()
						bus.Dispatch(action.Action[string]{
							Type:    ToggleLoggerAction.Name,
							Payload: fmt.Sprintf("%t", enabled),
							TraceID: fmt.Sprintf("trace-%d", time.Now().UnixNano()),
							Source:  "observability-ui",
						})
					}),
				),
				g.Text(" Enable Dev Logger"),
			),
			html.Br(),
			html.Button(
				html.ID("error-btn"),
				g.Text("Trigger Error"),
				g.Attr("style", "margin: 5px; padding: 8px 16px; background: #ff6b6b; color: white;"),
				dom.OnClick("error-btn", func() {
					bus.Dispatch(action.Action[string]{
						Type:    ErrorAction.Name,
						Payload: "Demo panic for observability testing",
						TraceID: fmt.Sprintf("trace-%d", time.Now().UnixNano()),
						Source:  "error-demo",
						Meta:    map[string]any{"intentional": true},
					})
				}),
			),
			html.Button(
				html.ID("analytics-btn"),
				g.Text("Send Analytics Event"),
				g.Attr("style", "margin: 5px; padding: 8px 16px; background: #4ecdc4; color: white;"),
				dom.OnClick("analytics-btn", func() {
					bus.Dispatch(action.Action[string]{
						Type:    AnalyticsAction.Name,
						Payload: "user_interaction",
						TraceID: fmt.Sprintf("trace-%d", time.Now().UnixNano()),
						Source:  "analytics-demo",
						Meta:    map[string]any{"event_type": "button_click"},
					})
				}),
			),
			html.Button(
				html.ID("debug-buffer-btn"),
				g.Text("Show Debug Buffer"),
				g.Attr("style", "margin: 5px; padding: 8px 16px; background: #9b59b6; color: white;"),
				dom.OnClick("debug-buffer-btn", func() {
					// Show debug buffer contents in console
					entries := action.GetDebugRingBufferEntries(bus, IncrementAction.Name)
					logutil.Logf("🔍 Debug buffer for %s (%d entries):", IncrementAction.Name, len(entries))
					for i, entry := range entries {
						logutil.Logf("  %d: %s at %s (TraceID: %s)", i, entry.Payload, entry.Timestamp.Format("15:04:05"), entry.TraceID)
					}
				}),
			),
		),

		// Statistics
		html.Div(
			g.Attr("style", "border: 1px solid #ddd; padding: 15px; margin: 10px 0; border-radius: 5px;"),
			html.H3(g.Text("📊 Statistics")),
			html.P(
				comps.BindText(func() string {
					return fmt.Sprintf("Errors recovered: %d", errorCount.Get())
				}),
			),
			html.P(
				comps.BindText(func() string {
					stats := action.GetObservabilityStats(bus)
					return fmt.Sprintf("Dev Logger: %t | Debug Buffer Size: %d | Error Handler: %t",
						stats.DevLoggerEnabled, stats.DebugBufferSize, stats.EnhancedErrorHandlerSet)
				}),
			),
		),

		// Dev Logger Output
		comps.Show(comps.ShowProps{
			When: devLoggerEnabled,
			Children: html.Div(
				g.Attr("style", "border: 1px solid #ddd; padding: 15px; margin: 10px 0; border-radius: 5px; background: #f8f9fa;"),
				html.H3(g.Text("📝 Dev Logger Output")),
				html.Div(
					g.Attr("style", "font-family: 'Courier New', monospace; font-size: 12px; background: #000; color: #0f0; padding: 10px; border-radius: 3px; max-height: 200px; overflow-y: auto;"),
					comps.For(comps.ForProps[string]{
						Items: logEntries,
						Children: func(entry string, index int) g.Node {
							return html.Div(g.Text(entry))
						},
					}),
				),
			),
		}),

		// Instructions
		html.Div(
			g.Attr("style", "border: 1px solid #ddd; padding: 15px; margin: 10px 0; border-radius: 5px; background: #f0f8ff;"),
			html.H3(g.Text("💡 Instructions")),
			html.Ul(
				html.Li(g.Text("Enable Dev Logger to see action dispatch details")),
				html.Li(g.Text("Use Increment/Decrement to generate traced actions")),
				html.Li(g.Text("Trigger Error to test enhanced error handling with context")),
				html.Li(g.Text("Send Analytics Event to see filtered analytics tap")),
				html.Li(g.Text("Show Debug Buffer to inspect action history in console")),
				html.Li(g.Text("Check browser console for detailed logs and trace information")),
			),
		),
	)
}
