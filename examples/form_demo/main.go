//go:build js && wasm

package main

import (
	"github.com/ozanturksever/logutil"
	"github.com/ozanturksever/uiwgo/comps"
	"github.com/ozanturksever/uiwgo/form"
	"github.com/ozanturksever/uiwgo/form/validators"
	"github.com/ozanturksever/uiwgo/form/widgets"
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

func main() {
	// Define form schema with comprehensive field types
	schema := []form.FieldDef{
		{
			Name:         "name",
			Label:        "Full Name",
			InitialValue: "",
			Validators:   []form.Validator{validators.Required("Name is required"), validators.MinLength(2, "Name must be at least 2 characters")},
			Widget:       widgets.TextInput,
		},
		{
			Name:         "email",
			Label:        "Email Address",
			InitialValue: "",
			Validators:   []form.Validator{validators.Required("Email is required"), validators.Email("Please enter a valid email address")},
			Widget:       widgets.EmailInput,
		},
		{
			Name:         "age",
			Label:        "Age",
			InitialValue: "",
			Validators:   []form.Validator{validators.Required("Age is required")},
			Widget:       widgets.TextInput,
		},
		{
			Name:         "password",
			Label:        "Password",
			InitialValue: "",
			Validators:   []form.Validator{validators.Required("Password is required"), validators.MinLength(8, "Password must be at least 8 characters")},
			Widget:       widgets.PasswordInput,
		},
		{
			Name:         "bio",
			Label:        "Biography",
			InitialValue: "",
			Validators:   []form.Validator{validators.MaxLength(500, "Biography must be less than 500 characters")},
			Widget:       widgets.TextArea,
		},
		{
			Name:         "country",
			Label:        "Country",
			InitialValue: "",
			Validators:   []form.Validator{validators.Required("Please select a country")},
			Widget: func(state *form.State, fieldName string, attrs ...Node) Node {
				return widgets.SelectWidget(state, fieldName, widgets.SelectOptions{
					Options: []widgets.SelectOption{
						{Value: "", Label: "Select a country"},
						{Value: "us", Label: "United States"},
						{Value: "ca", Label: "Canada"},
						{Value: "uk", Label: "United Kingdom"},
						{Value: "de", Label: "Germany"},
						{Value: "fr", Label: "France"},
					},
				})
			},
		},
		{
			Name:         "newsletter",
			Label:        "Subscribe to newsletter",
			InitialValue: false,
			Validators:   []form.Validator{},
			Widget: func(state *form.State, fieldName string, attrs ...Node) Node {
				return widgets.Checkbox(state, fieldName, widgets.CheckboxOptions{
					Label: "Subscribe to newsletter",
				})
			},
		},
		{
			Name:         "interests",
			Label:        "Interests",
			InitialValue: []string{},
			Validators:   []form.Validator{},
			Widget: func(state *form.State, fieldName string, attrs ...Node) Node {
				return widgets.CheckboxGroup(state, fieldName, widgets.CheckboxGroupOptions{
					Options: []widgets.CheckboxGroupOption{
						{Value: "programming", Label: "Programming"},
						{Value: "design", Label: "Design"},
						{Value: "music", Label: "Music"},
						{Value: "sports", Label: "Sports"},
						{Value: "travel", Label: "Travel"},
					},
				})
			},
		},
		{
			Name:         "gender",
			Label:        "Gender",
			InitialValue: "",
			Validators:   []form.Validator{},
			Widget: func(state *form.State, fieldName string, attrs ...Node) Node {
				return widgets.RadioGroup(state, fieldName, widgets.RadioGroupOptions{
					Options: []widgets.RadioOption{
						{Value: "male", Label: "Male"},
						{Value: "female", Label: "Female"},
						{Value: "other", Label: "Other"},
						{Value: "prefer-not-to-say", Label: "Prefer not to say"},
					},
				})
			},
		},
	}

	// Create form state
	formState := form.NewFromSchema(schema)

	// Handle form submission
	handleSubmit := func(state *form.State) error {
		values := state.Values()
		logutil.Log("Form submitted with values:", values)
		return nil
	}

	// Create the form component
	formComponent := func() Node {
		return Div(
			Class("min-h-screen bg-gradient-to-br from-blue-50 to-indigo-100 py-12 px-4 sm:px-6 lg:px-8"),
			Div(
				Class("max-w-2xl mx-auto"),
				Div(
					Class("bg-white shadow-xl rounded-lg overflow-hidden"),
					// Header
					Div(
						Class("bg-gradient-to-r from-blue-600 to-indigo-600 px-6 py-8"),
						H1(
							Class("text-3xl font-bold text-white text-center"),
							Text("User Registration Form"),
						),
						P(
							Class("text-blue-100 text-center mt-2"),
							Text("Please fill out all required fields"),
						),
					),
					// Form Content
					Div(
						Class("px-6 py-8"),
						form.SimpleForm(formState, handleSubmit,
							Div(
								Class("space-y-6"),
								// Personal Information Section
								Div(
									Class("border-b border-gray-200 pb-6"),
									H2(
										Class("text-xl font-semibold text-gray-900 mb-4 flex items-center"),
										Span(
											Class("w-6 h-6 bg-blue-100 text-blue-600 rounded-full flex items-center justify-center text-sm font-bold mr-3"),
											Text("1"),
										),
										Text("Personal Information"),
									),
									Div(
										Class("grid grid-cols-1 md:grid-cols-2 gap-6"),
										Div(
											Class("space-y-1"),
											form.Field(formState, "name", form.FieldOptions{ShowLabel: true, ShowError: true}),
										),
										Div(
											Class("space-y-1"),
											form.Field(formState, "email", form.FieldOptions{ShowLabel: true, ShowError: true}),
										),
									),
									Div(
										Class("grid grid-cols-1 md:grid-cols-2 gap-6"),
										Div(
											Class("space-y-1"),
											form.Field(formState, "age", form.FieldOptions{ShowLabel: true, ShowError: true}),
										),
										Div(
											Class("space-y-1"),
											form.Field(formState, "password", form.FieldOptions{ShowLabel: true, ShowError: true}),
										),
									),
								),
								// Additional Information Section
								Div(
									Class("border-b border-gray-200 pb-6"),
									H2(
										Class("text-xl font-semibold text-gray-900 mb-4 flex items-center"),
										Span(
											Class("w-6 h-6 bg-blue-100 text-blue-600 rounded-full flex items-center justify-center text-sm font-bold mr-3"),
											Text("2"),
										),
										Text("Additional Information"),
									),
									Div(
										Class("space-y-6"),
										Div(
											Class("space-y-1"),
											form.Field(formState, "bio", form.FieldOptions{ShowLabel: true, ShowError: true}),
										),
										Div(
											Class("space-y-1"),
											form.Field(formState, "country", form.FieldOptions{ShowLabel: true, ShowError: true}),
										),
									),
								),
								// Preferences Section
								Div(
									H2(
										Class("text-xl font-semibold text-gray-900 mb-4 flex items-center"),
										Span(
											Class("w-6 h-6 bg-blue-100 text-blue-600 rounded-full flex items-center justify-center text-sm font-bold mr-3"),
											Text("3"),
										),
										Text("Preferences"),
									),
									Div(
										Class("space-y-6"),
										Div(
											Class("space-y-1"),
											form.Field(formState, "newsletter", form.FieldOptions{ShowLabel: true, ShowError: true}),
										),
										Div(
											Class("space-y-1"),
											form.Field(formState, "interests", form.FieldOptions{ShowLabel: true, ShowError: true}),
										),
										Div(
											Class("space-y-1"),
											form.Field(formState, "gender", form.FieldOptions{ShowLabel: true, ShowError: true}),
										),
									),
								),
							),
							// Submit Button
							Div(
								Class("pt-6 border-t border-gray-200"),
								Button(
									Type("submit"),
									Class("w-full bg-gradient-to-r from-blue-600 to-indigo-600 text-white py-3 px-6 rounded-lg font-semibold text-lg hover:from-blue-700 hover:to-indigo-700 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2 transform transition-all duration-200 hover:scale-105 shadow-lg"),
									Text("Create Account"),
								),
							),
						),
					),
				),
			),
		)
	}

	// Mount the component
	comps.Mount("app", formComponent)
	select {}
}
