//go:build js && wasm

package main

import (
	"encoding/json"
	"syscall/js"
	"time"

	"github.com/ozanturksever/logutil"
	"github.com/ozanturksever/uiwgo/comps"
	"github.com/ozanturksever/uiwgo/dom"
	"github.com/ozanturksever/uiwgo/form"
	"github.com/ozanturksever/uiwgo/form/validators"
	"github.com/ozanturksever/uiwgo/form/widgets"
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

func main() {
	// Define form schema with various field types and validators
	schema := []form.FieldDef{
		{
			Name:         "name",
			Label:        "Full Name",
			InitialValue: "",
			Validators:   []form.Validator{validators.Required()},
			Widget:       widgets.TextInput,
		},
		{
			Name:         "email",
			Label:        "Email Address",
			InitialValue: "",
			Validators:   []form.Validator{validators.Required(), validators.Email()},
			Widget:       widgets.EmailInput,
		},
		{
			Name:         "password",
			Label:        "Password",
			InitialValue: "",
			Validators:   []form.Validator{validators.Required(), validators.MinLength(8)},
			Widget:       widgets.PasswordInput,
		},
		{
			Name:         "confirm_password",
			Label:        "Confirm Password",
			InitialValue: "",
			Validators:   []form.Validator{validators.Required()},
			Widget:       widgets.PasswordInput,
		},
		{
			Name:         "bio",
			Label:        "Biography",
			InitialValue: "",
			Validators:   []form.Validator{validators.MaxLength(500)},
			Widget:       widgets.TextArea,
			WidgetAttrs:  []Node{Rows("4"), Placeholder("Tell us about yourself...")},
		},
	}

	// Create form state
	formState := form.NewFromSchema(schema)

	// Helper function to update form state debug display
	updateFormStateDebug := func() {
		go func() {
			// Small delay to ensure form updates are processed
			time.Sleep(50 * time.Millisecond)

			formEl := js.Global().Get("document").Call("getElementById", "registration-form")
			debugEl := js.Global().Get("document").Call("getElementById", "form-state-debug")

			if !formEl.IsNull() && !debugEl.IsNull() {
				formData := js.Global().Get("FormData").New(formEl)
				data := make(map[string]interface{})

				// Extract form data
				formData.Call("forEach", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
					value := args[0].String()
					key := args[1].String()
					data[key] = value
					return nil
				}))

				// Convert to JSON and display
				jsonBytes, _ := json.MarshalIndent(data, "", "  ")
				debugEl.Set("textContent", string(jsonBytes))
			}
		}()
	}

	// Create the main app component
	app := func() Node {
		return Div(
			Class("min-h-screen bg-gray-50 py-12 px-4 sm:px-6 lg:px-8"),
			Div(
				Class("max-w-md mx-auto bg-white rounded-lg shadow-md p-6"),
				H1(
					Class("text-2xl font-bold text-gray-900 mb-6 text-center"),
					Text("User Registration"),
				),

				// Form with custom layout using individual Field components
				form.FormFor(formState, form.ForOptions{
					OnSubmit: func(state *form.State) error {
						// This will be handled by inline event binding
						return nil
					},
					ValidateOnSubmit: true,
					Attributes: []Node{
						Class("space-y-4"),
						ID("registration-form"),
						// Add inline event handlers
						dom.OnFormChangeInline(func(el dom.Element, formData map[string]string) {
							logutil.Log("Form data changed:", formData)
							updateFormStateDebug()
						}),
						dom.OnFormResetInline(func(el dom.Element) {
							logutil.Log("Form reset")
							updateFormStateDebug()
						}),
						dom.OnSubmitInline(func(el dom.Element, formData map[string]string) {
							logutil.Log("Form submitted with data:", formData)

							// Validate passwords match
							password := formData["password"]
							confirmPassword := formData["confirm_password"]

							if password != confirmPassword {
								go func() {
									time.Sleep(100 * time.Millisecond)
									js.Global().Call("alert", "Passwords do not match!")
								}()
								return
							}

							// Show success message and reset form
							go func() {
								time.Sleep(100 * time.Millisecond)

								// Convert form data to JSON for display
								jsonBytes, _ := json.MarshalIndent(formData, "", "  ")
								message := "Registration successful!\n\n" + string(jsonBytes)
								js.Global().Call("alert", message)

								// Reset form after alert
								formEl := js.Global().Get("document").Call("getElementById", "registration-form")
								if !formEl.IsNull() {
									formEl.Call("reset")
									time.Sleep(100 * time.Millisecond)
									updateFormStateDebug()
								}
							}()
						}),
					},
				},
					// Custom field layout
					Div(
						Class("grid grid-cols-1 gap-4"),

						// Name field
						form.Field(formState, "name", form.FieldOptions{
							ShowLabel:  true,
							ShowError:  true,
							Attributes: []Node{Class("form-field-custom")},
						}),

						// Email field
						form.Field(formState, "email", form.FieldOptions{
							ShowLabel:  true,
							ShowError:  true,
							Attributes: []Node{Class("form-field-custom")},
						}),

						// Password fields in a row
						Div(
							Class("grid grid-cols-1 md:grid-cols-2 gap-4"),
							form.Field(formState, "password", form.FieldOptions{
								ShowLabel:  true,
								ShowError:  true,
								Attributes: []Node{Class("form-field-custom")},
							}),
							form.Field(formState, "confirm_password", form.FieldOptions{
								ShowLabel:  true,
								ShowError:  true,
								Attributes: []Node{Class("form-field-custom")},
							}),
						),

						// Bio field
						form.Field(formState, "bio", form.FieldOptions{
							ShowLabel:  true,
							ShowError:  true,
							Attributes: []Node{Class("form-field-custom")},
						}),
					),

					// Submit button
					Div(
						Class("mt-6"),
						Button(
							Type("submit"),
							Class("w-full bg-blue-600 text-white py-2 px-4 rounded-md hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2 transition-colors"),
							ID("submit-btn"),
							Text("Register"),
						),
					),
				),

				// Demo section showing form state
				Div(
					Class("mt-8 p-4 bg-gray-100 rounded-lg"),
					H3(
						Class("text-lg font-semibold text-gray-800 mb-2"),
						Text("Form State (Debug)"),
					),
					Pre(
						Class("text-sm text-gray-600 whitespace-pre-wrap"),
						ID("form-state-debug"),
						Text("Form values will appear here..."),
					),
				),
			),
		)
	}

	// Mount the app
	comps.Mount("app", app)

	// Initialize form state debug display
	go func() {
		time.Sleep(100 * time.Millisecond)
		updateFormStateDebug()
	}()

	// Keep the program alive to handle events
	select {}
}
