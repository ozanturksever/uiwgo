package main

import (
	"app/golid"
	"fmt"

	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

func main() {
	golid.Render(ConditionalApp())
	golid.Run()
}

func ConditionalApp() Node {
	return Div(
		Style("font-family: Arial, sans-serif; max-width: 800px; margin: 50px auto; padding: 20px; background-color: #f5f5f5; min-height: 100vh;"),
		Div(
			Style("background: white; padding: 30px; border-radius: 10px; box-shadow: 0 2px 10px rgba(0,0,0,0.1);"),
			H1(Style("text-align: center; color: #333;"), Text("Conditional Rendering Example")),
			P(Style("text-align: center; color: #666;"), Text("Demonstrates dynamic UI updates based on state")),
			ConditionalComponent(),
		),
	)
}

func ConditionalComponent() Node {
	// State signals
	showDetails := golid.NewSignal(false)
	userType := golid.NewSignal("guest")
	count := golid.NewSignal(0)

	return Div(
		// Basic show/hide toggle
		Div(
			Style("margin-bottom: 30px; padding: 20px; border: 1px solid #ddd; border-radius: 8px;"),
			H3(Text("Show/Hide Toggle")),
			Button(
				Style("padding: 10px 20px; background-color: #007bff; color: white; border: none; border-radius: 5px; cursor: pointer; margin-right: 10px;"),
				golid.BindText(func() string {
					if showDetails.Get() {
						return "Hide Details"
					}
					return "Show Details"
				}),
				golid.OnClick(func() {
					showDetails.Set(!showDetails.Get())
				}),
			),
			golid.Bind(func() Node {
				if showDetails.Get() {
					return Div(
						Style("margin-top: 15px; padding: 15px; background-color: #f8f9fa; border-radius: 5px;"),
						Text("🎉 Here are the details you requested!"),
					)
				}
				return Div()
			}),
		),

		// User type conditional rendering
		Div(
			Style("margin-bottom: 30px; padding: 20px; border: 1px solid #ddd; border-radius: 8px;"),
			H3(Text("User Type Conditional")),
			Div(
				Style("margin-bottom: 15px;"),
				Button(
					Style("padding: 8px 15px; background-color: #28a745; color: white; border: none; border-radius: 3px; cursor: pointer; margin-right: 10px;"),
					Text("Guest"),
					golid.OnClick(func() { userType.Set("guest") }),
				),
				Button(
					Style("padding: 8px 15px; background-color: #17a2b8; color: white; border: none; border-radius: 3px; cursor: pointer; margin-right: 10px;"),
					Text("User"),
					golid.OnClick(func() { userType.Set("user") }),
				),
				Button(
					Style("padding: 8px 15px; background-color: #dc3545; color: white; border: none; border-radius: 3px; cursor: pointer;"),
					Text("Admin"),
					golid.OnClick(func() { userType.Set("admin") }),
				),
			),
			golid.Bind(func() Node {
				switch userType.Get() {
				case "guest":
					return Div(
						Style("padding: 10px; background-color: #e7f3ff; border-radius: 5px;"),
						Text("👋 Welcome guest! Please sign in to access more features."),
					)
				case "user":
					return Div(
						Style("padding: 10px; background-color: #e8f5e8; border-radius: 5px;"),
						Text("😊 Hello user! You have access to basic features."),
					)
				case "admin":
					return Div(
						Style("padding: 10px; background-color: #ffe6e6; border-radius: 5px;"),
						Text("🔐 Admin panel access granted! You can manage everything."),
					)
				default:
					return Text("Unknown user type")
				}
			}),
		),

		// Count-based conditional styling
		Div(
			Style("margin-bottom: 30px; padding: 20px; border: 1px solid #ddd; border-radius: 8px;"),
			H3(Text("Dynamic Styling")),
			Div(
				Style("margin-bottom: 15px;"),
				Button(
					Style("padding: 8px 15px; background-color: #28a745; color: white; border: none; border-radius: 3px; cursor: pointer; margin-right: 10px;"),
					Text("+"),
					golid.OnClick(func() { count.Set(count.Get() + 1) }),
				),
				Button(
					Style("padding: 8px 15px; background-color: #dc3545; color: white; border: none; border-radius: 3px; cursor: pointer;"),
					Text("-"),
					golid.OnClick(func() { count.Set(count.Get() - 1) }),
				),
			),
			golid.Bind(func() Node {
				c := count.Get()
				var bgColor, textColor, message string

				if c > 5 {
					bgColor = "#28a745"
					textColor = "white"
					message = fmt.Sprintf("🚀 High count: %d", c)
				} else if c > 0 {
					bgColor = "#ffc107"
					textColor = "black"
					message = fmt.Sprintf("⚡ Medium count: %d", c)
				} else if c == 0 {
					bgColor = "#6c757d"
					textColor = "white"
					message = "⚪ Zero count"
				} else {
					bgColor = "#dc3545"
					textColor = "white"
					message = fmt.Sprintf("🔻 Negative count: %d", c)
				}

				return Div(
					Style(fmt.Sprintf("padding: 15px; background-color: %s; color: %s; border-radius: 5px; text-align: center; font-weight: bold;", bgColor, textColor)),
					Text(message),
				)
			}),
		),
	)
}
