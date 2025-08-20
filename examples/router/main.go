package main

import (
	"app/golid"
	"fmt"

	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

func main() {
	// Initialize the router
	router := golid.NewRouter()
	golid.SetGlobalRouter(router)

	// Define routes
	router.AddRoute("/", func(params golid.RouteParams) Node {
		return HomePage()
	})

	router.AddRoute("/about", func(params golid.RouteParams) Node {
		return AboutPage()
	})

	router.AddRoute("/contact", func(params golid.RouteParams) Node {
		return ContactPage()
	})

	router.AddRoute("/user/:id", func(params golid.RouteParams) Node {
		userID := params["id"]
		return UserPage(userID)
	})

	router.AddRoute("/products/:category/:id", func(params golid.RouteParams) Node {
		category := params["category"]
		productID := params["id"]
		return ProductPage(category, productID)
	})

	// Render the router-based app
	golid.Render(RouterApp())
	golid.Run()
}

func RouterApp() Node {
	return Div(
		Style("font-family: Arial, sans-serif; min-height: 100vh; background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);"),
		NavigationBar(),
		golid.RouterOutlet(),
	)
}

func NavigationBar() Node {
	return Div(
		Style("background: rgba(255,255,255,0.95); backdrop-filter: blur(10px); padding: 20px; box-shadow: 0 2px 20px rgba(0,0,0,0.1); position: sticky; top: 0; z-index: 100;"),
		Div(
			Style("max-width: 1000px; margin: 0 auto; display: flex; align-items: center; justify-content: space-between;"),

			// Logo
			H1(
				Style("margin: 0; color: #333; font-size: 1.8em;"),
				Text("🚀 Router Demo"),
			),

			// Navigation Links
			Nav(
				Style("display: flex; gap: 20px; align-items: center;"),
				golid.RouterLink("/",
					Div(
						Style("padding: 10px 15px; color: #333; text-decoration: none; border-radius: 5px; transition: background-color 0.2s; cursor: pointer;"),
						Text("Home"),
					),
				),
				golid.RouterLink("/about",
					Div(
						Style("padding: 10px 15px; color: #333; text-decoration: none; border-radius: 5px; transition: background-color 0.2s; cursor: pointer;"),
						Text("About"),
					),
				),
				golid.RouterLink("/contact",
					Div(
						Style("padding: 10px 15px; color: #333; text-decoration: none; border-radius: 5px; transition: background-color 0.2s; cursor: pointer;"),
						Text("Contact"),
					),
				),
				golid.RouterLink("/user/123",
					Div(
						Style("padding: 10px 15px; color: #333; text-decoration: none; border-radius: 5px; transition: background-color 0.2s; cursor: pointer;"),
						Text("User Profile"),
					),
				),
				golid.RouterLink("/products/electronics/laptop-456",
					Div(
						Style("padding: 10px 15px; color: #333; text-decoration: none; border-radius: 5px; transition: background-color 0.2s; cursor: pointer;"),
						Text("Sample Product"),
					),
				),
			),
		),
	)
}

func HomePage() Node {
	return PageContainer(
		H1(Text("Welcome to Router Demo")),
		P(Text("This example demonstrates client-side routing with Golid framework.")),
		Div(
			Style("display: grid; grid-template-columns: repeat(auto-fit, minmax(250px, 1fr)); gap: 20px; margin-top: 30px;"),

			FeatureCard(
				"Static Routes",
				"Simple navigation between fixed pages like Home, About, and Contact.",
				"/about",
			),

			FeatureCard(
				"Dynamic Routes",
				"Routes with parameters like /user/:id to show user-specific content.",
				"/user/123",
			),

			FeatureCard(
				"Nested Parameters",
				"Multiple parameters in a single route like /products/:category/:id.",
				"/products/electronics/laptop-456",
			),

			FeatureCard(
				"Contact Form",
				"A contact page with form handling and navigation back to home.",
				"/contact",
			),
		),

		Div(
			Style("margin-top: 40px; padding: 20px; background-color: rgba(255,255,255,0.1); border-radius: 10px;"),
			H3(Text("Router Features Demonstrated:")),
			Ul(
				Li(Text("Client-side navigation with RouterLink")),
				Li(Text("Route parameters extraction")),
				Li(Text("RouterOutlet for rendering matched routes")),
				Li(Text("Multiple parameter routes")),
				Li(Text("Navigation state management")),
			),
		),
	)
}

func AboutPage() Node {
	return PageContainer(
		H1(Text("About This Demo")),
		P(Text("This routing demo showcases how to build single-page applications with Golid framework.")),

		Div(
			Style("display: flex; gap: 20px; margin-top: 30px;"),
			Div(
				Style("flex: 1; padding: 20px; background-color: rgba(255,255,255,0.1); border-radius: 10px;"),
				H3(Text("Key Features")),
				Ul(
					Li(Text("Declarative routing with route patterns")),
					Li(Text("Automatic parameter extraction")),
					Li(Text("History API integration")),
					Li(Text("Component-based page rendering")),
				),
			),
			Div(
				Style("flex: 1; padding: 20px; background-color: rgba(255,255,255,0.1); border-radius: 10px;"),
				H3(Text("Technical Details")),
				Ul(
					Li(Text("Built with Go and WebAssembly")),
					Li(Text("Reactive components with signals")),
					Li(Text("No JavaScript dependencies")),
					Li(Text("Type-safe route parameters")),
				),
			),
		),

		Div(
			Style("margin-top: 30px; text-align: center;"),
			golid.RouterLink("/",
				Button(
					Style("padding: 10px 20px; background-color: #007bff; color: white; border: none; border-radius: 5px; cursor: pointer;"),
					Text("← Back to Home"),
				),
			),
		),
	)
}

func ContactPage() Node {
	// Contact form state
	name := golid.NewSignal("")
	email := golid.NewSignal("")
	message := golid.NewSignal("")
	submitted := golid.NewSignal(false)

	return PageContainer(
		H1(Text("Contact Us")),
		P(Text("Get in touch with us using the form below.")),

		golid.Bind(func() Node {
			if submitted.Get() {
				return Div(
					Style("text-align: center; padding: 40px; background-color: rgba(40, 167, 69, 0.2); border-radius: 10px; margin-top: 20px;"),
					H3(Style("color: #28a745;"), Text("✅ Message Sent!")),
					P(Text("Thank you for your message. We'll get back to you soon.")),
					Button(
						Style("padding: 10px 20px; background-color: #007bff; color: white; border: none; border-radius: 5px; cursor: pointer; margin-top: 10px;"),
						Text("Send Another Message"),
						golid.OnClickV2(func() {
							submitted.Set(false)
							name.Set("")
							email.Set("")
							message.Set("")
						}),
					),
				)
			}

			return Div(
				Style("max-width: 500px; margin: 0 auto; background-color: rgba(255,255,255,0.1); padding: 30px; border-radius: 10px; margin-top: 20px;"),
				Div(
					Style("margin-bottom: 20px;"),
					Label(Style("display: block; margin-bottom: 5px; font-weight: bold;"), Text("Name:")),
					Input(
						Type("text"),
						Placeholder("Your name"),
						Style("width: 100%; padding: 10px; border: 2px solid #ddd; border-radius: 5px; font-size: 16px;"),
						golid.OnInputV2(func(val string) {
							name.Set(val)
						}),
					),
				),

				Div(
					Style("margin-bottom: 20px;"),
					Label(Style("display: block; margin-bottom: 5px; font-weight: bold;"), Text("Email:")),
					Input(
						Type("email"),
						Placeholder("your.email@example.com"),
						Style("width: 100%; padding: 10px; border: 2px solid #ddd; border-radius: 5px; font-size: 16px;"),
						golid.OnInputV2(func(val string) {
							email.Set(val)
						}),
					),
				),

				Div(
					Style("margin-bottom: 20px;"),
					Label(Style("display: block; margin-bottom: 5px; font-weight: bold;"), Text("Message:")),
					Textarea(
						Rows("5"),
						Placeholder("Your message here..."),
						Style("width: 100%; padding: 10px; border: 2px solid #ddd; border-radius: 5px; font-size: 16px; resize: vertical;"),
						golid.OnInputV2(func(val string) {
							message.Set(val)
						}),
					),
				),

				Button(
					Style("width: 100%; padding: 15px; background-color: #28a745; color: white; border: none; border-radius: 5px; cursor: pointer; font-size: 16px;"),
					Text("Send Message"),
					golid.OnClickV2(func() {
						if name.Get() != "" && email.Get() != "" && message.Get() != "" {
							submitted.Set(true)
						}
					}),
				),
			)
		}),
	)
}

func UserPage(userID string) Node {
	return PageContainer(
		H1(golid.BindText(func() string {
			return fmt.Sprintf("User Profile: %s", userID)
		})),
		P(Text("This page demonstrates route parameters extraction.")),

		Div(
			Style("background-color: rgba(255,255,255,0.1); padding: 20px; border-radius: 10px; margin-top: 20px;"),
			H3(Text("Route Information")),
			P(golid.BindText(func() string {
				return fmt.Sprintf("User ID from route: %s", userID)
			})),
			P(Text("This value was extracted from the URL path /user/:id")),
		),

		Div(
			Style("margin-top: 30px; display: flex; gap: 10px; justify-content: center;"),
			golid.RouterLink("/user/456",
				Button(
					Style("padding: 10px 20px; background-color: #17a2b8; color: white; border: none; border-radius: 5px; cursor: pointer;"),
					Text("View User 456"),
				),
			),
			golid.RouterLink("/user/789",
				Button(
					Style("padding: 10px 20px; background-color: #17a2b8; color: white; border: none; border-radius: 5px; cursor: pointer;"),
					Text("View User 789"),
				),
			),
			golid.RouterLink("/",
				Button(
					Style("padding: 10px 20px; background-color: #6c757d; color: white; border: none; border-radius: 5px; cursor: pointer;"),
					Text("← Back to Home"),
				),
			),
		),
	)
}

func ProductPage(category, productID string) Node {
	return PageContainer(
		H1(Text("Product Details")),
		P(Text("This page demonstrates multiple route parameters.")),

		Div(
			Style("background-color: rgba(255,255,255,0.1); padding: 20px; border-radius: 10px; margin-top: 20px;"),
			H3(Text("Product Information")),
			P(golid.BindText(func() string {
				return fmt.Sprintf("Category: %s", category)
			})),
			P(golid.BindText(func() string {
				return fmt.Sprintf("Product ID: %s", productID)
			})),
			P(Text("Both values were extracted from /products/:category/:id")),
		),

		Div(
			Style("margin-top: 30px; display: flex; gap: 10px; justify-content: center; flex-wrap: wrap;"),
			golid.RouterLink("/products/books/novel-123",
				Button(
					Style("padding: 10px 15px; background-color: #17a2b8; color: white; border: none; border-radius: 5px; cursor: pointer;"),
					Text("Books → Novel 123"),
				),
			),
			golid.RouterLink("/products/clothing/shirt-789",
				Button(
					Style("padding: 10px 15px; background-color: #17a2b8; color: white; border: none; border-radius: 5px; cursor: pointer;"),
					Text("Clothing → Shirt 789"),
				),
			),
			golid.RouterLink("/",
				Button(
					Style("padding: 10px 15px; background-color: #6c757d; color: white; border: none; border-radius: 5px; cursor: pointer;"),
					Text("← Back to Home"),
				),
			),
		),
	)
}

// Helper components
func PageContainer(children ...Node) Node {
	nodes := []Node{Style("max-width: 1000px; margin: 0 auto; padding: 40px 20px; color: white;")}
	nodes = append(nodes, children...)
	return Div(nodes...)
}

func FeatureCard(title, description, href string) Node {
	return golid.RouterLink(href,
		Div(
			Style("background: rgba(255,255,255,0.1); backdrop-filter: blur(10px); padding: 20px; border-radius: 10px; cursor: pointer; transition: transform 0.2s, background-color 0.2s; border: 1px solid rgba(255,255,255,0.2);"),
			H3(Style("margin-top: 0; color: white;"), Text(title)),
			P(Style("color: rgba(255,255,255,0.8); margin-bottom: 0;"), Text(description)),
		),
	)
}
