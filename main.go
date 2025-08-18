package main

import (
	"app/golid"
	"fmt"
	"strings"

	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

func main() {

	// ---Uncomment one of the following examples to run it ---

	//app := CounterComponent()
	//app := List1()
	//app := List2()
	// app := TextCopyDemo()
	// app := TextCopyDemo1()
	//app := TextCopyDemo2()
	// app := TextCopyDemo3()

	// BindInputWithType examples
	//app := InputTypesDemo()
	// app := NumberInputDemo()
	// app := EmailPasswordDemo()

	// BindInputWithFocus examples
	// app := FocusStateDemo()
	// app := ValidationDemo()
	app := DynamicStylingDemo()

	// Combined examples
	// app := ComprehensiveFormDemo()

	// 🎯 ALL EXAMPLES IN ONE PAGE
	//app := AllExamplesShowcase()

	golid.Render(app)
	golid.Run()
}

func CounterComponent() Node {
	// Observable (represents the state of the app)
	count := golid.NewSignal(0)

	return Div(
		Style("border: 1px solid orange; padding: 10px; margin: 10px;"),

		// Bind text Element to the reactive count signal (observable)
		golid.Bind(func() Node {
			return Div(Text(fmt.Sprintf("Count = %d", count.Get())))
		}),

		// [+] Button element
		Button(
			Style("margin: 3px;"),
			Text("+"),
			golid.OnClick(func() {
				count.Set(count.Get() + 1)
			}),
		),

		// [-] Button element
		Button(
			Style("margin: 3px;"),
			Text("-"),
			golid.OnClick(func() {
				count.Set(count.Get() - 1)
			}),
		),
	)
}

func List1() Node {
	messages := []string{"Hello", "World", "Golid"}

	return Div(
		H3(Text("Messages")),
		golid.ForEach(messages, func(msg string) Node {
			return Li(Text(msg))
		}),
	)
}

func List2() Node {
	messages := []string{"Hello", "World", "Golid"}

	return Div(
		H3(Text("Messages")),
		Ul( // Wrap list items in a <ul>
			golid.ForEach(messages, func(msg string) Node {
				return Li(
					Text(msg),
					golid.OnClick(func() {
						golid.Log("Clicked on:", msg)
					}),
				)
			}),
		),
	)
}

func TextCopyDemo() Node {
	// Internal state
	inputValue := golid.NewSignal("")
	copiedText := golid.NewSignal("")

	return Div(
		H3(Text("Text Copier")),

		// Input field with OnInput binding
		Input(
			Type("text"),
			Placeholder("Type something..."),
			golid.OnInput(func(val string) {
				inputValue.Set(val)
			}),
		),

		// Button that copies current input value to output signal
		Button(
			Style("margin-left: 10px;"),
			Text("Copy"),
			golid.OnClick(func() {
				copiedText.Set(inputValue.Get())
			}),
		),

		// Output label
		Div(
			Style("margin-top: 10px; font-weight: bold;"),
			Text("Copied text: "),
			golid.Bind(func() Node {
				return Text(copiedText.Get())
			}),
		),
	)
}

func TextCopyDemo1() Node {
	inputValue := golid.NewSignal("")

	return Div(
		H3(Text("Live Mirror")),

		// Input field
		Input(
			Type("text"),
			Placeholder("Start typing..."),
			golid.OnInput(func(val string) {
				inputValue.Set(val)
			}),
		),

		// Live-updating label
		Div(
			Style("margin-top: 10px; font-weight: bold;"),
			Text("You typed: "),
			golid.Bind(func() Node {
				return Text(inputValue.Get())
			}),
		),
	)
}

func TextCopyDemo2() Node {
	shared := golid.NewSignal("")

	return Div(
		H3(Text("Twin Mirror Inputs")),

		// Input 1 (reactively updates on shared signal change)
		golid.Bind(func() Node {
			return Input(
				Type("text"),
				Placeholder("Input 1"),
				Style("margin-left: 10px;"),
				Value(shared.Get()),
				golid.OnInput(func(val string) {
					if val != shared.Get() {
						shared.Set(val)
					}
				}),
			)
		}),

		// Input 2 (also reacts to signal change)
		golid.Bind(func() Node {
			return Input(
				Type("text"),
				Placeholder("Input 2"),
				Style("margin-left: 10px;"),
				Value(shared.Get()),
				golid.OnInput(func(val string) {
					if val != shared.Get() {
						shared.Set(val)
					}
				}),
			)
		}),

		// Display current value
		Div(
			Style("margin-top: 10px; font-style: italic;"),
			Text("Shared value: "),
			golid.Bind(func() Node {
				return Text(shared.Get())
			}),
		),
	)
}

func TextCopyDemo3() Node {
	shared := golid.NewSignal("")

	return Div(
		H3(Text("Twin Mirror Inputs")),

		// Input 1 (reactively updates on shared signal change)
		golid.BindInput(shared, "1"),
		golid.BindInput(shared, "2"),

		// Display current value
		Div(
			Style("margin-top: 10px; font-style: italic;"),
			Text("Shared value: "),
			golid.Bind(func() Node {
				return Text(shared.Get())
			}),
		),
	)
}

// ==========================================
// BindInputWithType Examples
// ==========================================

func InputTypesDemo() Node {
	textValue := golid.NewSignal("")
	emailValue := golid.NewSignal("")
	passwordValue := golid.NewSignal("")
	urlValue := golid.NewSignal("")
	searchValue := golid.NewSignal("")

	return Div(
		Style("max-width: 600px; margin: 20px; font-family: Arial, sans-serif;"),
		H2(Text("🔤 Input Types Demo")),

		// Text Input
		Div(
			Style("margin-bottom: 15px;"),
			Label(Style("display: block; font-weight: bold; margin-bottom: 5px;"), Text("Text Input:")),
			golid.BindInputWithType(textValue, "text", "Enter some text..."),
			Div(Style("font-size: 12px; color: #666; margin-top: 5px;"),
				Text("Value: "), golid.BindText(func() string { return textValue.Get() })),
		),

		// Email Input
		Div(
			Style("margin-bottom: 15px;"),
			Label(Style("display: block; font-weight: bold; margin-bottom: 5px;"), Text("Email Input:")),
			golid.BindInputWithType(emailValue, "email", "user@example.com"),
			Div(Style("font-size: 12px; color: #666; margin-top: 5px;"),
				Text("Email: "), golid.BindText(func() string { return emailValue.Get() })),
		),

		// Password Input
		Div(
			Style("margin-bottom: 15px;"),
			Label(Style("display: block; font-weight: bold; margin-bottom: 5px;"), Text("Password Input:")),
			golid.BindInputWithType(passwordValue, "password", "Enter password..."),
			Div(Style("font-size: 12px; color: #666; margin-top: 5px;"),
				Text("Length: "), golid.BindText(func() string { return fmt.Sprintf("%d characters", len(passwordValue.Get())) })),
		),

		// URL Input
		Div(
			Style("margin-bottom: 15px;"),
			Label(Style("display: block; font-weight: bold; margin-bottom: 5px;"), Text("URL Input:")),
			golid.BindInputWithType(urlValue, "url", "https://example.com"),
			Div(Style("font-size: 12px; color: #666; margin-top: 5px;"),
				Text("URL: "), golid.BindText(func() string { return urlValue.Get() })),
		),

		// Search Input
		Div(
			Style("margin-bottom: 15px;"),
			Label(Style("display: block; font-weight: bold; margin-bottom: 5px;"), Text("Search Input:")),
			golid.BindInputWithType(searchValue, "search", "Search something..."),
			Div(Style("font-size: 12px; color: #666; margin-top: 5px;"),
				Text("Search: "), golid.BindText(func() string { return searchValue.Get() })),
		),
	)
}

func NumberInputDemo() Node {
	numberValue := golid.NewSignal("0")
	rangeValue := golid.NewSignal("50")

	return Div(
		Style("max-width: 500px; margin: 20px; font-family: Arial, sans-serif;"),
		H2(Text("🔢 Number Input Demo")),

		// Number Input
		Div(
			Style("margin-bottom: 20px;"),
			Label(Style("display: block; font-weight: bold; margin-bottom: 5px;"), Text("Number Input:")),
			golid.BindInputWithType(numberValue, "number", "Enter a number..."),
			Div(Style("margin-top: 10px; padding: 10px; background: #f0f0f0; border-radius: 5px;"),
				Text("Value: "), golid.BindText(func() string { return numberValue.Get() }),
				Br(),
				Text("Doubled: "), golid.BindText(func() string {
					if val := numberValue.Get(); val != "" {
						if num := parseNumber(val); num != 0 {
							return fmt.Sprintf("%.2f", num*2)
						}
					}
					return "0"
				}),
			),
		),

		// Range Input
		Div(
			Style("margin-bottom: 20px;"),
			Label(Style("display: block; font-weight: bold; margin-bottom: 5px;"), Text("Range Slider:")),
			golid.BindInputWithType(rangeValue, "range", ""),
			Div(Style("margin-top: 10px; padding: 10px; background: #e8f4fd; border-radius: 5px;"),
				Text("Slider Value: "), golid.BindText(func() string { return rangeValue.Get() }),
				Br(),
				golid.Bind(func() Node {
					val := rangeValue.Get()
					if val == "" {
						val = "0"
					}
					width := val + "%"
					return Div(
						Style("margin-top: 5px; height: 20px; background: #ddd; border-radius: 10px; overflow: hidden;"),
						Div(Style("height: 100%; background: linear-gradient(90deg, #4CAF50, #2196F3); width: "+width+"; transition: width 0.3s;")),
					)
				}),
			),
		),
	)
}

func EmailPasswordDemo() Node {
	email := golid.NewSignal("")
	password := golid.NewSignal("")
	confirmPassword := golid.NewSignal("")

	return Div(
		Style("max-width: 400px; margin: 20px; font-family: Arial, sans-serif;"),
		H2(Text("📧 Login Form Demo")),

		Div(
			Style("padding: 20px; border: 1px solid #ddd; border-radius: 8px; background: #fafafa;"),

			// Email
			Div(
				Style("margin-bottom: 15px;"),
				Label(Style("display: block; font-weight: bold; margin-bottom: 5px;"), Text("Email:")),
				golid.BindInputWithType(email, "email", "your@email.com"),
			),

			// Password
			Div(
				Style("margin-bottom: 15px;"),
				Label(Style("display: block; font-weight: bold; margin-bottom: 5px;"), Text("Password:")),
				golid.BindInputWithType(password, "password", "Password"),
			),

			// Confirm Password
			Div(
				Style("margin-bottom: 20px;"),
				Label(Style("display: block; font-weight: bold; margin-bottom: 5px;"), Text("Confirm Password:")),
				golid.BindInputWithType(confirmPassword, "password", "Confirm Password"),
			),

			// Form Status
			golid.Bind(func() Node {
				emailVal := email.Get()
				passVal := password.Get()
				confirmVal := confirmPassword.Get()

				var status string
				var color string

				if emailVal == "" || passVal == "" {
					status = "⚠️  Please fill in all fields"
					color = "#ff9800"
				} else if len(passVal) < 6 {
					status = "❌ Password must be at least 6 characters"
					color = "#f44336"
				} else if passVal != confirmVal {
					status = "❌ Passwords don't match"
					color = "#f44336"
				} else {
					status = "✅ Form is valid!"
					color = "#4caf50"
				}

				return Div(
					Style("padding: 10px; border-radius: 5px; background: "+color+"22; color: "+color+"; font-weight: bold;"),
					Text(status),
				)
			}),
		),
	)
}

// ==========================================
// BindInputWithFocus Examples
// ==========================================

func FocusStateDemo() Node {
	inputValue := golid.NewSignal("")
	isFocused := golid.NewSignal(false)

	return Div(
		Style("max-width: 500px; margin: 20px; font-family: Arial, sans-serif;"),
		H2(Text("🎯 Focus State Demo")),

		Div(
			Style("margin-bottom: 20px;"),
			Label(Style("display: block; font-weight: bold; margin-bottom: 5px;"), Text("Focus-Aware Input:")),
			golid.BindInputWithFocus(inputValue, isFocused, "Click to focus..."),
		),

		// Focus Status Display
		golid.Bind(func() Node {
			focused := isFocused.Get()
			var status, color, bgColor string

			if focused {
				status = "🟢 Input is FOCUSED"
				color = "#4caf50"
				bgColor = "#e8f5e8"
			} else {
				status = "🔴 Input is NOT focused"
				color = "#666"
				bgColor = "#f5f5f5"
			}

			return Div(
				Style("padding: 15px; border-radius: 8px; background: "+bgColor+"; color: "+color+"; font-weight: bold; margin-bottom: 15px;"),
				Text(status),
			)
		}),

		// Value Display
		Div(
			Style("padding: 10px; background: #f0f0f0; border-radius: 5px;"),
			Text("Current value: "), golid.BindText(func() string {
				val := inputValue.Get()
				if val == "" {
					return "(empty)"
				}
				return "\"" + val + "\""
			}),
		),
	)
}

func ValidationDemo() Node {
	username := golid.NewSignal("")
	usernameFocus := golid.NewSignal(false)
	email := golid.NewSignal("")
	emailFocus := golid.NewSignal(false)

	return Div(
		Style("max-width: 500px; margin: 20px; font-family: Arial, sans-serif;"),
		H2(Text("✅ Validation Demo")),

		Div(
			Style("padding: 20px; border: 1px solid #ddd; border-radius: 8px;"),

			// Username field
			Div(
				Style("margin-bottom: 20px;"),
				Label(Style("display: block; font-weight: bold; margin-bottom: 5px;"), Text("Username:")),
				golid.BindInputWithFocus(username, usernameFocus, "Enter username..."),

				// Username validation
				golid.Bind(func() Node {
					val := username.Get()
					focused := usernameFocus.Get()

					if val == "" && !focused {
						return Div() // Empty when not focused and empty
					}

					var message, color string
					if len(val) < 3 && val != "" {
						message = "❌ Username must be at least 3 characters"
						color = "#f44336"
					} else if len(val) >= 3 {
						message = "✅ Username looks good!"
						color = "#4caf50"
					}

					if message != "" {
						return Div(
							Style("font-size: 12px; color: "+color+"; margin-top: 5px;"),
							Text(message),
						)
					}
					return Div()
				}),
			),

			// Email field
			Div(
				Style("margin-bottom: 20px;"),
				Label(Style("display: block; font-weight: bold; margin-bottom: 5px;"), Text("Email:")),
				golid.BindInputWithFocus(email, emailFocus, "Enter email..."),

				// Email validation
				golid.Bind(func() Node {
					val := email.Get()
					focused := emailFocus.Get()

					if val == "" && !focused {
						return Div()
					}

					var message, color string
					if val != "" && !strings.Contains(val, "@") {
						message = "❌ Please enter a valid email address"
						color = "#f44336"
					} else if strings.Contains(val, "@") && strings.Contains(val, ".") {
						message = "✅ Email format looks good!"
						color = "#4caf50"
					}

					if message != "" {
						return Div(
							Style("font-size: 12px; color: "+color+"; margin-top: 5px;"),
							Text(message),
						)
					}
					return Div()
				}),
			),
		),
	)
}

func DynamicStylingDemo() Node {
	input1 := golid.NewSignal("")
	focus1 := golid.NewSignal(false)
	input2 := golid.NewSignal("")
	focus2 := golid.NewSignal(false)
	input3 := golid.NewSignal("")
	focus3 := golid.NewSignal(false)

	return Div(
		Style("max-width: 600px; margin: 20px; font-family: Arial, sans-serif;"),
		H2(Text("🎨 Dynamic Styling Demo")),

		// Glowing border effect
		Div(
			Style("margin-bottom: 20px;"),
			Label(Style("display: block; font-weight: bold; margin-bottom: 5px;"), Text("Glowing Border:")),
			golid.Bind(func() Node {
				focused := focus1.Get()
				style := "padding: 12px; border-radius: 8px; font-size: 16px; width: 100%; box-sizing: border-box; transition: all 0.3s;"
				if focused {
					style += " border: 2px solid #2196F3; box-shadow: 0 0 10px rgba(33, 150, 243, 0.3);"
				} else {
					style += " border: 2px solid #ddd;"
				}
				return golid.BindInputWithFocus(input1, focus1, "Focus me for glow effect...")
			}),
		),

		// Color changing background
		Div(
			Style("margin-bottom: 20px;"),
			Label(Style("display: block; font-weight: bold; margin-bottom: 5px;"), Text("Color Changing:")),
			golid.Bind(func() Node {
				focused := focus2.Get()
				style := "padding: 12px; border-radius: 8px; font-size: 16px; width: 100%; box-sizing: border-box; transition: all 0.3s; border: 2px solid #ddd;"
				if focused {
					style += " background: linear-gradient(45deg, #ff9a9e 0%, #fecfef 50%, #fecfef 100%);"
				} else {
					style += " background: white;"
				}
				return golid.BindInputWithFocus(input2, focus2, "Focus for gradient background...")
			}),
		),

		// Scale effect
		Div(
			Style("margin-bottom: 20px;"),
			Label(Style("display: block; font-weight: bold; margin-bottom: 5px;"), Text("Scale Effect:")),
			golid.Bind(func() Node {
				focused := focus3.Get()
				style := "padding: 12px; border-radius: 8px; font-size: 16px; width: 100%; box-sizing: border-box; transition: all 0.3s; border: 2px solid #4caf50;"
				if focused {
					style += " transform: scale(1.05); box-shadow: 0 5px 15px rgba(76, 175, 80, 0.3);"
				}
				return golid.BindInputWithFocus(input3, focus3, "Focus me to scale up...")
			}),
		),

		// Status display
		Div(
			Style("padding: 15px; background: #f5f5f5; border-radius: 8px;"),
			H4(Text("Focus Status:")),
			Ul(
				Li(Text("Input 1: "), golid.BindText(func() string {
					if focus1.Get() {
						return "🟢 Focused"
					}
					return "⚪ Not focused"
				})),
				Li(Text("Input 2: "), golid.BindText(func() string {
					if focus2.Get() {
						return "🟢 Focused"
					}
					return "⚪ Not focused"
				})),
				Li(Text("Input 3: "), golid.BindText(func() string {
					if focus3.Get() {
						return "🟢 Focused"
					}
					return "⚪ Not focused"
				})),
			),
		),
	)
}

// ==========================================
// Combined Example
// ==========================================

func ComprehensiveFormDemo() Node {
	// Form data signals
	firstName := golid.NewSignal("")
	lastName := golid.NewSignal("")
	email := golid.NewSignal("")
	age := golid.NewSignal("18")
	website := golid.NewSignal("")

	// Focus signals
	firstNameFocus := golid.NewSignal(false)
	lastNameFocus := golid.NewSignal(false)
	emailFocus := golid.NewSignal(false)
	ageFocus := golid.NewSignal(false)
	websiteFocus := golid.NewSignal(false)

	return Div(
		Style("max-width: 600px; margin: 20px; font-family: Arial, sans-serif;"),
		H2(Text("📋 Comprehensive Form Demo")),

		Div(
			Style("padding: 25px; border: 1px solid #ddd; border-radius: 12px; background: white; box-shadow: 0 2px 10px rgba(0,0,0,0.1);"),

			// First Name
			Div(
				Style("margin-bottom: 20px;"),
				Label(Style("display: block; font-weight: bold; margin-bottom: 8px; color: #333;"), Text("First Name:")),
				golid.BindInputWithFocus(firstName, firstNameFocus, "Enter your first name..."),
				golid.Bind(func() Node {
					val := firstName.Get()
					focused := firstNameFocus.Get()
					if val == "" && focused {
						return Div(Style("font-size: 12px; color: #ff9800; margin-top: 5px;"), Text("⚠️ First name is required"))
					}
					if len(val) > 0 && len(val) < 2 {
						return Div(Style("font-size: 12px; color: #f44336; margin-top: 5px;"), Text("❌ Name too short"))
					}
					if len(val) >= 2 {
						return Div(Style("font-size: 12px; color: #4caf50; margin-top: 5px;"), Text("✅ Looks good!"))
					}
					return Div()
				}),
			),

			// Last Name
			Div(
				Style("margin-bottom: 20px;"),
				Label(Style("display: block; font-weight: bold; margin-bottom: 8px; color: #333;"), Text("Last Name:")),
				golid.BindInputWithFocus(lastName, lastNameFocus, "Enter your last name..."),
				golid.Bind(func() Node {
					val := lastName.Get()
					focused := lastNameFocus.Get()
					if val == "" && focused {
						return Div(Style("font-size: 12px; color: #ff9800; margin-top: 5px;"), Text("⚠️ Last name is required"))
					}
					if len(val) >= 2 {
						return Div(Style("font-size: 12px; color: #4caf50; margin-top: 5px;"), Text("✅ Looks good!"))
					}
					return Div()
				}),
			),

			// Email
			Div(
				Style("margin-bottom: 20px;"),
				Label(Style("display: block; font-weight: bold; margin-bottom: 8px; color: #333;"), Text("Email Address:")),
				Div(
					Style("width: 100%;"),
					golid.BindInputWithType(email, "email", "user@example.com"),
				),
				golid.Bind(func() Node {
					val := email.Get()
					if val == "" {
						return Div(Style("font-size: 12px; color: #999; margin-top: 5px;"), Text("Please enter your email"))
					}
					if val != "" && !strings.Contains(val, "@") {
						return Div(Style("font-size: 12px; color: #f44336; margin-top: 5px;"), Text("❌ Invalid email format"))
					}
					if strings.Contains(val, "@") && strings.Contains(val, ".") {
						return Div(Style("font-size: 12px; color: #4caf50; margin-top: 5px;"), Text("✅ Valid email!"))
					}
					return Div()
				}),
			),

			// Age
			Div(
				Style("margin-bottom: 20px;"),
				Label(Style("display: block; font-weight: bold; margin-bottom: 8px; color: #333;"), Text("Age:")),
				Div(
					Style("width: 100%;"),
					golid.BindInputWithType(age, "number", "18"),
				),
				golid.Bind(func() Node {
					val := age.Get()
					if val == "" {
						return Div(Style("font-size: 12px; color: #999; margin-top: 5px;"), Text("Please enter your age"))
					}
					if num := parseNumber(val); num > 0 {
						if num < 13 {
							return Div(Style("font-size: 12px; color: #f44336; margin-top: 5px;"), Text("❌ Must be at least 13 years old"))
						}
						if num > 120 {
							return Div(Style("font-size: 12px; color: #ff9800; margin-top: 5px;"), Text("⚠️ Please enter a valid age"))
						}
						return Div(Style("font-size: 12px; color: #4caf50; margin-top: 5px;"), Text("✅ Valid age"))
					}
					return Div()
				}),
			),

			// Website
			Div(
				Style("margin-bottom: 25px;"),
				Label(Style("display: block; font-weight: bold; margin-bottom: 8px; color: #333;"), Text("Website (Optional):")),
				Div(
					Style("width: 100%;"),
					golid.BindInputWithType(website, "url", "https://yourwebsite.com"),
				),
				golid.Bind(func() Node {
					val := website.Get()
					if val == "" {
						return Div(Style("font-size: 12px; color: #999; margin-top: 5px;"), Text("Optional - enter your website"))
					}
					if val != "" && !strings.HasPrefix(val, "http://") && !strings.HasPrefix(val, "https://") {
						return Div(Style("font-size: 12px; color: #ff9800; margin-top: 5px;"), Text("⚠️ URL should start with http:// or https://"))
					}
					if strings.HasPrefix(val, "http") && strings.Contains(val, ".") {
						return Div(Style("font-size: 12px; color: #4caf50; margin-top: 5px;"), Text("✅ Valid URL!"))
					}
					return Div()
				}),
			),

			// Form Summary
			Hr(),
			Div(
				Style("margin-top: 20px; padding: 15px; background: #f8f9fa; border-radius: 8px;"),
				H4(Style("margin-top: 0; color: #333;"), Text("Form Summary:")),
				golid.Bind(func() Node {
					return Div(
						Text("👤 Name: "), golid.BindText(func() string {
							first := firstName.Get()
							last := lastName.Get()
							if first == "" && last == "" {
								return "(not entered)"
							}
							return first + " " + last
						}),
						Br(),
						Text("📧 Email: "), golid.BindText(func() string {
							val := email.Get()
							if val == "" {
								return "(not entered)"
							}
							return val
						}),
						Br(),
						Text("🎂 Age: "), golid.BindText(func() string {
							val := age.Get()
							if val == "" {
								return "(not entered)"
							}
							return val + " years old"
						}),
						Br(),
						Text("🌐 Website: "), golid.BindText(func() string {
							val := website.Get()
							if val == "" {
								return "(none)"
							}
							return val
						}),
					)
				}),
			),

			// Focus Status
			Div(
				Style("margin-top: 15px; padding: 10px; background: #e3f2fd; border-radius: 5px;"),
				Text("👀 Currently focused: "), golid.BindText(func() string {
					if firstNameFocus.Get() {
						return "First Name"
					}
					if lastNameFocus.Get() {
						return "Last Name"
					}
					if emailFocus.Get() {
						return "Email"
					}
					if ageFocus.Get() {
						return "Age"
					}
					if websiteFocus.Get() {
						return "Website"
					}
					return "None"
				}),
			),
		),
	)
}

// Helper function to parse numbers
func parseNumber(s string) float64 {
	if s == "" {
		return 0
	}
	// Simple number parsing - in real app you'd use strconv.ParseFloat
	var result float64
	fmt.Sscanf(s, "%f", &result)
	return result
}

// ==========================================
// 🎯 COMPLETE SHOWCASE - ALL EXAMPLES
// ==========================================

func AllExamplesShowcase() Node {
	return Div(
		Style("margin: 0; padding: 0; font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif; background: linear-gradient(135deg, #667eea 0%, #764ba2 100%); min-height: 100vh;"),

		// Header
		Div(
			Style("background: rgba(255,255,255,0.95); backdrop-filter: blur(10px); padding: 20px 0; text-align: center; box-shadow: 0 2px 20px rgba(0,0,0,0.1); position: sticky; top: 0; z-index: 100;"),
			H1(Style("margin: 0; color: #333; font-size: 2.5em; font-weight: 300;"), Text("🚀 Golid Framework Showcase")),
			P(Style("margin: 10px 0 0 0; color: #666; font-size: 1.1em;"), Text("A comprehensive demo of all reactive UI features")),
		),

		// Main content container
		Div(
			Style("display: flex; max-width: 1400px; margin: 0 auto; gap: 30px; padding: 30px;"),

			// Navigation Sidebar
			Div(
				Style("flex: 0 0 250px; background: rgba(255,255,255,0.9); border-radius: 12px; padding: 20px; height: fit-content; position: sticky; top: 120px; box-shadow: 0 8px 32px rgba(0,0,0,0.1);"),
				H3(Style("margin-top: 0; color: #333; border-bottom: 2px solid #667eea; padding-bottom: 10px;"), Text("📋 Navigation")),

				// Navigation links (visual only since we can't do real navigation in this setup)
				Div(Style("margin-bottom: 15px;"),
					Strong(Style("color: #667eea; font-size: 14px;"), Text("🧮 BASICS"))),
				Ul(Style("list-style: none; padding-left: 0; margin: 5px 0 20px 0;"),
					Li(Style("margin: 8px 0; padding: 8px; background: #f8f9fa; border-radius: 6px; font-size: 13px;"), Text("• Counter Component")),
					Li(Style("margin: 8px 0; padding: 8px; background: #f8f9fa; border-radius: 6px; font-size: 13px;"), Text("• List Examples")),
				),

				Div(Style("margin-bottom: 15px;"),
					Strong(Style("color: #667eea; font-size: 14px;"), Text("📝 TEXT INPUTS"))),
				Ul(Style("list-style: none; padding-left: 0; margin: 5px 0 20px 0;"),
					Li(Style("margin: 8px 0; padding: 8px; background: #f8f9fa; border-radius: 6px; font-size: 13px;"), Text("• Text Copy Demos")),
					Li(Style("margin: 8px 0; padding: 8px; background: #f8f9fa; border-radius: 6px; font-size: 13px;"), Text("• Twin Mirror Inputs")),
				),

				Div(Style("margin-bottom: 15px;"),
					Strong(Style("color: #667eea; font-size: 14px;"), Text("🔤 INPUT TYPES"))),
				Ul(Style("list-style: none; padding-left: 0; margin: 5px 0 20px 0;"),
					Li(Style("margin: 8px 0; padding: 8px; background: #f8f9fa; border-radius: 6px; font-size: 13px;"), Text("• Various Input Types")),
					Li(Style("margin: 8px 0; padding: 8px; background: #f8f9fa; border-radius: 6px; font-size: 13px;"), Text("• Number & Range")),
					Li(Style("margin: 8px 0; padding: 8px; background: #f8f9fa; border-radius: 6px; font-size: 13px;"), Text("• Login Form")),
				),

				Div(Style("margin-bottom: 15px;"),
					Strong(Style("color: #667eea; font-size: 14px;"), Text("🎯 FOCUS FEATURES"))),
				Ul(Style("list-style: none; padding-left: 0; margin: 5px 0 20px 0;"),
					Li(Style("margin: 8px 0; padding: 8px; background: #f8f9fa; border-radius: 6px; font-size: 13px;"), Text("• Focus State Demo")),
					Li(Style("margin: 8px 0; padding: 8px; background: #f8f9fa; border-radius: 6px; font-size: 13px;"), Text("• Live Validation")),
					Li(Style("margin: 8px 0; padding: 8px; background: #f8f9fa; border-radius: 6px; font-size: 13px;"), Text("• Dynamic Styling")),
				),

				Div(Style("margin-bottom: 15px;"),
					Strong(Style("color: #667eea; font-size: 14px;"), Text("📋 COMPREHENSIVE"))),
				Ul(Style("list-style: none; padding-left: 0; margin: 5px 0 0 0;"),
					Li(Style("margin: 8px 0; padding: 8px; background: #f8f9fa; border-radius: 6px; font-size: 13px;"), Text("• Complete Form")),
				),
			),

			// Main content area
			Div(
				Style("flex: 1; display: flex; flex-direction: column; gap: 40px;"),

				// Basic Examples Section
				createSection("🧮 Basic Components", "Core reactive features and fundamental building blocks",
					Div(Style("display: grid; grid-template-columns: repeat(auto-fit, minmax(400px, 1fr)); gap: 25px;"),
						createCard("Counter Component", CounterComponent()),
						createCard("Static List", List1()),
						createCard("Interactive List", List2()),
					),
				),

				// Text Input Examples Section
				createSection("📝 Text Input Demos", "Various approaches to handling text input and reactivity",
					Div(Style("display: grid; grid-template-columns: repeat(auto-fit, minmax(450px, 1fr)); gap: 25px;"),
						createCard("Text Copier", TextCopyDemo()),
						createCard("Live Mirror", TextCopyDemo1()),
					),
					Div(Style("margin-top: 25px; display: grid; grid-template-columns: repeat(auto-fit, minmax(450px, 1fr)); gap: 25px;"),
						createCard("Twin Mirror (Manual)", TextCopyDemo2()),
						createCard("Twin Mirror (BindInput)", TextCopyDemo3()),
					),
				),

				// BindInputWithType Examples Section
				createSection("🔤 Input Types Demo", "Showcase of different HTML input types with reactive binding",
					Div(Style("display: grid; grid-template-columns: repeat(auto-fit, minmax(500px, 1fr)); gap: 25px;"),
						createCard("Various Input Types", InputTypesDemo()),
						createCard("Number & Range Inputs", NumberInputDemo()),
						createCard("Email & Password Form", EmailPasswordDemo()),
					),
				),

				// BindInputWithFocus Examples Section
				createSection("🎯 Focus State Features", "Advanced input handling with focus state tracking",
					Div(Style("display: grid; grid-template-columns: repeat(auto-fit, minmax(500px, 1fr)); gap: 25px;"),
						createCard("Focus State Demo", FocusStateDemo()),
						createCard("Live Validation", ValidationDemo()),
						createCard("Dynamic Styling", DynamicStylingDemo()),
					),
				),

				// Comprehensive Example Section
				createSection("📋 Comprehensive Example", "A complete form combining all features together",
					createCard("Complete Registration Form", ComprehensiveFormDemo()),
				),

				// Footer
				Div(
					Style("margin-top: 50px; padding: 30px; background: rgba(255,255,255,0.9); border-radius: 12px; text-align: center; box-shadow: 0 8px 32px rgba(0,0,0,0.1);"),
					H3(Style("margin-top: 0; color: #333;"), Text("🎉 End of Showcase")),
					P(Style("color: #666; line-height: 1.6; margin-bottom: 0;"),
						Text("This showcase demonstrates the powerful reactive capabilities of the Golid framework. "),
						Text("From basic components to advanced form handling with focus states and validation, "),
						Text("Golid provides a complete toolkit for building interactive web applications with Go and WASM.")),
				),
			),
		),
	)
}

// Helper function to create consistent section headers
func createSection(title, description string, content ...Node) Node {
	return Div(
		Style("background: rgba(255,255,255,0.95); border-radius: 16px; padding: 30px; box-shadow: 0 8px 32px rgba(0,0,0,0.1); backdrop-filter: blur(10px);"),
		H2(Style("margin-top: 0; margin-bottom: 10px; color: #333; font-size: 1.8em; font-weight: 600;"), Text(title)),
		P(Style("margin-bottom: 25px; color: #666; font-size: 1.1em; line-height: 1.5;"), Text(description)),
		Group(content),
	)
}

// Helper function to create consistent cards
func createCard(title string, content Node) Node {
	return Div(
		Style("background: #fafbfc; border: 1px solid #e1e5e9; border-radius: 12px; padding: 20px; box-shadow: 0 2px 8px rgba(0,0,0,0.05); transition: transform 0.2s ease, box-shadow 0.2s ease;"),
		H4(Style("margin-top: 0; margin-bottom: 15px; color: #495057; font-size: 1.2em; font-weight: 600; padding-bottom: 8px; border-bottom: 2px solid #e9ecef;"), Text(title)),
		content,
	)
}
