package main

import (
	"app/golid"
	"fmt"
	"time"

	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

func main() {
	golid.Render(LifecycleApp())
	golid.Run()
}

func LifecycleApp() Node {
	return Div(
		Style("font-family: Arial, sans-serif; max-width: 1000px; margin: 50px auto; padding: 20px; background-color: #f5f5f5; min-height: 100vh;"),
		Div(
			Style("background: white; padding: 30px; border-radius: 10px; box-shadow: 0 2px 10px rgba(0,0,0,0.1);"),
			H1(Style("text-align: center; color: #333; margin-bottom: 10px;"), Text("Component Lifecycle Example")),
			P(Style("text-align: center; color: #666; margin-bottom: 30px;"), Text("Demonstrates OnInit, OnMount, OnDismount hooks and cleanup patterns")),

			// Log area at the top
			LogArea(),

			// Main demo components
			Div(
				Style("display: grid; grid-template-columns: 1fr 1fr; gap: 20px; margin-top: 20px;"),

				// Left column - Dynamic Components
				Div(
					H2(Style("color: #333; margin-bottom: 15px;"), Text("Dynamic Components")),
					DynamicComponentDemo(),
				),

				// Right column - Cleanup Examples
				Div(
					H2(Style("color: #333; margin-bottom: 15px;"), Text("Cleanup Examples")),
					CleanupDemo(),
				),
			),

			// Bottom section - Always Mounted Component
			Div(
				Style("margin-top: 30px; padding: 20px; background-color: #f8f9fa; border-radius: 8px;"),
				H2(Style("color: #333; margin-bottom: 15px;"), Text("Always Mounted Component")),
				AlwaysMountedComponent().Render(),
			),
		),
	)
}

// Global log signal to track lifecycle events
var lifecycleLog = golid.NewSignal([]string{})

func addLog(message string) {
	currentLog := lifecycleLog.Get()
	timestamp := time.Now().Format("15:04:05.000")
	newEntry := fmt.Sprintf("[%s] %s", timestamp, message)

	// Keep only the last 15 entries
	if len(currentLog) >= 15 {
		currentLog = currentLog[1:]
	}

	lifecycleLog.Set(append(currentLog, newEntry))
	golid.Log("Lifecycle:", newEntry)
}

func LogArea() Node {
	return Div(
		Style("background-color: #1e1e1e; color: #00ff00; padding: 15px; border-radius: 5px; font-family: monospace; font-size: 12px; max-height: 200px; overflow-y: auto; margin-bottom: 20px;"),
		H3(Style("color: #00ff00; margin-top: 0;"), Text("📋 Lifecycle Events Log")),
		golid.Bind(func() Node {
			logs := lifecycleLog.Get()
			if len(logs) == 0 {
				return Div(
					Style("font-style: italic; color: #888;"),
					Text("No lifecycle events yet..."),
				)
			}

			var logNodes []Node
			for _, entry := range logs {
				logNodes = append(logNodes, Div(Text(entry)))
			}
			return Div(logNodes...)
		}),
	)
}

func DynamicComponentDemo() Node {
	showComponent1 := golid.NewSignal(false)
	showComponent2 := golid.NewSignal(false)
	componentCount := golid.NewSignal(0)

	// Create component instances once to prevent recreation loops
	timerComp := TimerComponent()
	counterComp := CounterComponent()

	return Div(
		Style("background-color: #fff3cd; padding: 20px; border-radius: 8px; border-left: 4px solid #ffc107;"),

		P(Text("Toggle components to see lifecycle hooks in action:")),

		// Control buttons
		Div(
			Style("display: flex; gap: 10px; margin-bottom: 20px; flex-wrap: wrap;"),
			Button(
				Style("padding: 8px 16px; background-color: #007bff; color: white; border: none; border-radius: 4px; cursor: pointer;"),
				golid.BindText(func() string {
					if showComponent1.Get() {
						return "Hide Timer Component"
					}
					return "Show Timer Component"
				}),
				golid.OnClickV2(func() {
					showComponent1.Set(!showComponent1.Get())
				}),
			),
			Button(
				Style("padding: 8px 16px; background-color: #28a745; color: white; border: none; border-radius: 4px; cursor: pointer;"),
				golid.BindText(func() string {
					if showComponent2.Get() {
						return "Hide Counter Component"
					}
					return "Show Counter Component"
				}),
				golid.OnClickV2(func() {
					showComponent2.Set(!showComponent2.Get())
				}),
			),
			Button(
				Style("padding: 8px 16px; background-color: #17a2b8; color: white; border: none; border-radius: 4px; cursor: pointer;"),
				Text("Create Temporary Component"),
				golid.OnClickV2(func() {
					count := componentCount.Get()
					componentCount.Set(count + 1)

					// This will create a component that self-destructs after 3 seconds
					go func() {
						time.Sleep(3 * time.Second)
						componentCount.Set(componentCount.Get() - 1)
					}()
				}),
			),
		),

		// Dynamic component areas - now using pre-created instances
		golid.Bind(func() Node {
			if showComponent1.Get() {
				return timerComp.Render()
			}
			return Div()
		}),

		golid.Bind(func() Node {
			if showComponent2.Get() {
				return counterComp.Render()
			}
			return Div()
		}),

		// Temporary components - these are meant to be recreated
		golid.Bind(func() Node {
			count := componentCount.Get()
			if count == 0 {
				return Div()
			}

			var components []Node
			for i := 0; i < count; i++ {
				components = append(components,
					TemporaryComponent(i+1).Render(),
				)
			}
			return Div(components...)
		}),
	)
}

func CleanupDemo() Node {
	// Create component instance once to prevent recreation loops
	resourceComp := ResourceManagerComponent()

	return Div(
		Style("background-color: #d1ecf1; padding: 20px; border-radius: 8px; border-left: 4px solid #17a2b8;"),

		P(Text("These components demonstrate proper cleanup:")),

		Ul(
			Li(Text("🕒 Timer components clear intervals on dismount")),
			Li(Text("📡 Event listeners are properly removed")),
			Li(Text("🧹 Resources are cleaned up automatically")),
			Li(Text("📊 State is properly reset")),
		),

		resourceComp.Render(),
	)
}

// Timer Component with cleanup
func TimerComponent() *golid.Component {
	elapsed := golid.NewSignal(0)
	var stopTimer chan bool

	return golid.NewComponent(func() Node {
		return Div(
			Style("margin: 10px 0; padding: 15px; background-color: white; border-radius: 5px; border: 1px solid #ddd;"),
			H4(Style("margin-top: 0; color: #007bff;"), Text("⏱️ Timer Component")),
			golid.BindText(func() string {
				return fmt.Sprintf("Elapsed: %d seconds", elapsed.Get())
			}),
		)
	}).OnInit(func() {
		addLog("🔄 TimerComponent: OnInit - Component initializing")
		stopTimer = make(chan bool)
	}).OnMount(func() {
		addLog("🟢 TimerComponent: OnMount - Starting timer")

		// Start timer using goroutine (simulating setInterval)
		go func() {
			ticker := time.NewTicker(1 * time.Second)
			defer ticker.Stop()

			for {
				select {
				case <-ticker.C:
					elapsed.Set(elapsed.Get() + 1)
				case <-stopTimer:
					addLog("🛑 TimerComponent: Goroutine stopping - Timer cleanup complete")
					return
				}
			}
		}()

	}).OnDismount(func() {
		addLog("🔴 TimerComponent: OnDismount - Cleaning up timer")
		if stopTimer != nil {
			close(stopTimer)
		}
		elapsed.Set(0) // Reset state
	})
}

// Counter Component with lifecycle logging
func CounterComponent() *golid.Component {
	count := golid.NewSignal(0)

	return golid.NewComponent(func() Node {
		return Div(
			Style("margin: 10px 0; padding: 15px; background-color: white; border-radius: 5px; border: 1px solid #ddd;"),
			H4(Style("margin-top: 0; color: #28a745;"), Text("🔢 Counter Component")),
			Div(
				Style("margin: 10px 0;"),
				golid.BindText(func() string {
					return fmt.Sprintf("Count: %d", count.Get())
				}),
			),
			Button(
				Style("padding: 5px 10px; background-color: #28a745; color: white; border: none; border-radius: 3px; cursor: pointer; margin-right: 5px;"),
				Text("+"),
				golid.OnClickV2(func() {
					count.Set(count.Get() + 1)
				}),
			),
			Button(
				Style("padding: 5px 10px; background-color: #dc3545; color: white; border: none; border-radius: 3px; cursor: pointer;"),
				Text("-"),
				golid.OnClickV2(func() {
					count.Set(count.Get() - 1)
				}),
			),
		)
	}).OnInit(func() {
		addLog("🔄 CounterComponent: OnInit - Initializing with count 0")
	}).OnMount(func() {
		addLog("🟢 CounterComponent: OnMount - Component mounted to DOM")
	}).OnDismount(func() {
		addLog("🔴 CounterComponent: OnDismount - Resetting counter state")
		count.Set(0)
	})
}

// Temporary Component that demonstrates short lifecycle
func TemporaryComponent(id int) *golid.Component {
	return golid.NewComponent(func() Node {
		return Div(
			Style("margin: 5px 0; padding: 10px; background-color: #fff3cd; border-radius: 4px; border: 1px solid #ffc107;"),
			Text(fmt.Sprintf("⏳ Temporary Component #%d (will disappear in 3s)", id)),
		)
	}).OnInit(func() {
		addLog(fmt.Sprintf("🔄 TempComponent#%d: OnInit - Created", id))
	}).OnMount(func() {
		addLog(fmt.Sprintf("🟢 TempComponent#%d: OnMount - Mounted", id))
	}).OnDismount(func() {
		addLog(fmt.Sprintf("🔴 TempComponent#%d: OnDismount - Self-destructed", id))
	})
}

// Always Mounted Component to show persistent lifecycle
func AlwaysMountedComponent() *golid.Component {
	mountTime := golid.NewSignal(time.Now())
	heartbeat := golid.NewSignal(0)
	var stopHeartbeat chan bool

	return golid.NewComponent(func() Node {
		return Div(
			Style("padding: 15px; background-color: #e8f5e8; border-radius: 5px; border: 1px solid #28a745;"),
			H4(Style("margin-top: 0; color: #28a745;"), Text("💚 Always Mounted Component")),
			P(golid.BindText(func() string {
				duration := time.Since(mountTime.Get())
				return fmt.Sprintf("Mounted for: %v", duration.Round(time.Second))
			})),
			P(golid.BindText(func() string {
				return fmt.Sprintf("Heartbeat: %d", heartbeat.Get())
			})),
			P(
				Style("font-size: 12px; color: #666; margin-bottom: 0;"),
				Text("This component demonstrates a component that stays mounted throughout the app lifecycle."),
			),
		)
	}).OnInit(func() {
		addLog("🔄 AlwaysMountedComponent: OnInit - Persistent component initializing")
		mountTime.Set(time.Now())
		stopHeartbeat = make(chan bool)

		// Start heartbeat
		go func() {
			ticker := time.NewTicker(2 * time.Second)
			defer ticker.Stop()

			for {
				select {
				case <-ticker.C:
					heartbeat.Set(heartbeat.Get() + 1)
				case <-stopHeartbeat:
					addLog("🛑 AlwaysMountedComponent: Heartbeat stopping")
					return
				}
			}
		}()

	}).OnMount(func() {
		addLog("🟢 AlwaysMountedComponent: OnMount - Persistent component mounted")
	}).OnDismount(func() {
		addLog("🔴 AlwaysMountedComponent: OnDismount - This should not happen in normal flow")
		if stopHeartbeat != nil {
			close(stopHeartbeat)
		}
	})
}

// Resource Manager Component demonstrating complex cleanup
func ResourceManagerComponent() *golid.Component {
	connectionStatus := golid.NewSignal("disconnected")
	messageCount := golid.NewSignal(0)
	var stopMessages chan bool

	return golid.NewComponent(func() Node {
		return Div(
			Style("margin-top: 15px; padding: 15px; background-color: white; border-radius: 5px; border: 1px solid #17a2b8;"),
			H4(Style("margin-top: 0; color: #17a2b8;"), Text("🔌 Resource Manager")),
			P(golid.BindText(func() string {
				return fmt.Sprintf("Connection: %s", connectionStatus.Get())
			})),
			P(golid.BindText(func() string {
				return fmt.Sprintf("Messages received: %d", messageCount.Get())
			})),
			P(
				Style("font-size: 12px; color: #666; margin-bottom: 0;"),
				Text("Simulates a component that manages external resources and connections."),
			),
		)
	}).OnInit(func() {
		addLog("🔄 ResourceManager: OnInit - Preparing resources")
		stopMessages = make(chan bool)
	}).OnMount(func() {
		addLog("🟢 ResourceManager: OnMount - Connecting to external services")
		connectionStatus.Set("connecting...")

		// Simulate connection process
		go func() {
			time.Sleep(1 * time.Second)
			connectionStatus.Set("connected")
			addLog("📡 ResourceManager: Connected to external service")

			// Simulate receiving messages
			ticker := time.NewTicker(3 * time.Second)
			defer ticker.Stop()

			for {
				select {
				case <-ticker.C:
					messageCount.Set(messageCount.Get() + 1)
				case <-stopMessages:
					addLog("🛑 ResourceManager: Message receiver stopping")
					return
				}
			}
		}()

	}).OnDismount(func() {
		addLog("🔴 ResourceManager: OnDismount - Cleaning up resources and connections")
		if stopMessages != nil {
			close(stopMessages)
		}
		connectionStatus.Set("disconnected")
		messageCount.Set(0)
	})
}
