//go:build js && wasm

package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/ozanturksever/logutil"
	"github.com/ozanturksever/uiwgo/comps"
	"github.com/ozanturksever/uiwgo/dom"
	"github.com/ozanturksever/uiwgo/reactivity"
	g "maragu.dev/gomponents"
	h "maragu.dev/gomponents/html"
)

type FormStep struct {
	ID          string
	Title       string
	Description string
	IsValid     func() bool
	Component   func() g.Node
}

type MultiStepFormState struct {
	currentStep reactivity.Signal[int]
	steps       []FormStep

	// Form data
	personalInfo reactivity.Signal[PersonalInfo]
	contactInfo  reactivity.Signal[ContactInfo]
	preferences  reactivity.Signal[Preferences]

	// Validation
	errors       reactivity.Signal[map[string]string]
	isSubmitting reactivity.Signal[bool]
	submitted    reactivity.Signal[bool]
}

type PersonalInfo struct {
	FirstName string
	LastName  string
	BirthDate string
	Gender    string
}

type ContactInfo struct {
	Email   string
	Phone   string
	Address string
	City    string
	Country string
}

type Preferences struct {
	Newsletter    bool
	Notifications bool
	Theme         string
	Language      string
}

func NewMultiStepFormState() *MultiStepFormState {
	state := &MultiStepFormState{
		currentStep:  reactivity.CreateSignal(0),
		personalInfo: reactivity.CreateSignal(PersonalInfo{}),
		contactInfo:  reactivity.CreateSignal(ContactInfo{}),
		preferences:  reactivity.CreateSignal(Preferences{Theme: "light", Language: "en"}),
		errors:       reactivity.CreateSignal(make(map[string]string)),
		isSubmitting: reactivity.CreateSignal(false),
		submitted:    reactivity.CreateSignal(false),
	}

	state.steps = []FormStep{
		{
			ID:          "personal",
			Title:       "Personal Information",
			Description: "Please provide your basic personal details",
			IsValid:     state.validatePersonalInfo,
			Component:   state.renderPersonalInfoStep,
		},
		{
			ID:          "contact",
			Title:       "Contact Information",
			Description: "How can we reach you?",
			IsValid:     state.validateContactInfo,
			Component:   state.renderContactInfoStep,
		},
		{
			ID:          "preferences",
			Title:       "Preferences",
			Description: "Customize your experience",
			IsValid:     state.validatePreferences,
			Component:   state.renderPreferencesStep,
		},
		{
			ID:          "review",
			Title:       "Review & Submit",
			Description: "Please review your information before submitting",
			IsValid:     func() bool { return true },
			Component:   state.renderReviewStep,
		},
	}

	return state
}

func (mfs *MultiStepFormState) validatePersonalInfo() bool {
	info := mfs.personalInfo.Get()
	errors := make(map[string]string)

	if strings.TrimSpace(info.FirstName) == "" {
		errors["firstName"] = "First name is required"
	}

	if strings.TrimSpace(info.LastName) == "" {
		errors["lastName"] = "Last name is required"
	}

	if info.BirthDate == "" {
		errors["birthDate"] = "Birth date is required"
	}

	mfs.errors.Set(errors)
	return len(errors) == 0
}

func (mfs *MultiStepFormState) validateContactInfo() bool {
	info := mfs.contactInfo.Get()
	errors := make(map[string]string)

	if !strings.Contains(info.Email, "@") {
		errors["email"] = "Valid email is required"
	}

	if len(info.Phone) < 10 {
		errors["phone"] = "Valid phone number is required"
	}

	if strings.TrimSpace(info.Address) == "" {
		errors["address"] = "Address is required"
	}

	mfs.errors.Set(errors)
	return len(errors) == 0
}

func (mfs *MultiStepFormState) validatePreferences() bool {
	// Preferences are optional, so always valid
	mfs.errors.Set(make(map[string]string))
	return true
}

func (mfs *MultiStepFormState) render() g.Node {
	return h.Div(
		h.Class("multi-step-form"),

		// Form header
		h.Div(
			h.Class("form-header"),
			h.H1(g.Text("Multi-Step Form")),
			h.P(g.Text("Complete your registration in 4 easy steps")),
		),

		// Step indicator
		mfs.renderStepIndicator(),

		// Form content
		h.Div(
			h.Class("form-content"),
			comps.Show(comps.ShowProps{
				When: reactivity.CreateMemo(func() bool {
					return !mfs.submitted.Get()
				}),
				Children: mfs.renderCurrentStep(),
			}),
			comps.Show(comps.ShowProps{
				When:     mfs.submitted,
				Children: mfs.renderSuccessMessage(),
			}),
		),

		// Navigation buttons
		comps.Show(comps.ShowProps{
			When: reactivity.CreateMemo(func() bool {
				return !mfs.submitted.Get()
			}),
			Children: mfs.renderNavigation(),
		}),
	)
}

func (mfs *MultiStepFormState) renderStepIndicator() g.Node {
	return h.Div(
		h.Class("step-indicator"),

		comps.For(comps.ForProps[FormStep]{
			Items: reactivity.CreateSignal(mfs.steps),
			Key: func(step FormStep) string {
				return step.ID
			},
			Children: func(step FormStep, index int) g.Node {
				currentStep := mfs.currentStep.Get()
				isCompleted := index < currentStep

				return h.Div(
					h.Class("step"),
					g.Attr("style", func() string {
						if index < len(mfs.steps)-1 {
							return "position: relative;"
						}
						return ""
					}()),

					h.Div(
						h.Class("step-number"),
						g.If(isCompleted, g.Text("✓")),
						g.If(!isCompleted, g.Text(fmt.Sprintf("%d", index+1))),
					),
					h.Div(
						h.Class("step-title"),
						g.Text(step.Title),
					),
				)
			},
		}),
	)
}

func (mfs *MultiStepFormState) renderCurrentStep() g.Node {
	return comps.Switch(comps.SwitchProps{
		When: reactivity.CreateMemo(func() string {
			step := mfs.currentStep.Get()
			if step >= 0 && step < len(mfs.steps) {
				return mfs.steps[step].ID
			}
			return "unknown"
		}),
		Children: func() []g.Node {
			var children []g.Node
			for _, step := range mfs.steps {
				children = append(children, comps.Match(comps.MatchProps{
					When:     step.ID,
					Children: step.Component(),
				}))
			}
			return children
		}(),
	})
}

func (mfs *MultiStepFormState) renderPersonalInfoStep() g.Node {
	info := mfs.personalInfo.Get()
	errors := mfs.errors.Get()

	return h.Div(
		h.Class("form-step"),

		h.H2(g.Text("Personal Information")),
		h.P(g.Text("Please provide your basic personal details")),

		h.Div(
			h.Class("form-group"),
			h.Label(g.Text("First Name *")),
			h.Input(
				h.Type("text"),
				h.Value(info.FirstName),
				g.Attr("class", func() string {
					if errors["firstName"] != "" {
						return "error"
					}
					return ""
				}()),
				dom.OnInputInline(func(el dom.Element) {
					newInfo := info
					newInfo.FirstName = el.Underlying().Get("value").String()
					mfs.personalInfo.Set(newInfo)
				}),
			),
			comps.Show(comps.ShowProps{
				When: reactivity.CreateMemo(func() bool {
					return errors["firstName"] != ""
				}),
				Children: h.Div(
					h.Class("error-message"),
					g.Text(errors["firstName"]),
				),
			}),
		),

		h.Div(
			h.Class("form-group"),
			h.Label(g.Text("Last Name *")),
			h.Input(
				h.Type("text"),
				h.Value(info.LastName),
				g.Attr("class", func() string {
					if errors["lastName"] != "" {
						return "error"
					}
					return ""
				}()),
				dom.OnInputInline(func(el dom.Element) {
					newInfo := info
					newInfo.LastName = el.Underlying().Get("value").String()
					mfs.personalInfo.Set(newInfo)
				}),
			),
			comps.Show(comps.ShowProps{
				When: reactivity.CreateMemo(func() bool {
					return errors["lastName"] != ""
				}),
				Children: h.Div(
					h.Class("error-message"),
					g.Text(errors["lastName"]),
				),
			}),
		),

		h.Div(
			h.Class("form-group"),
			h.Label(g.Text("Birth Date *")),
			h.Input(
				h.Type("date"),
				h.Value(info.BirthDate),
				g.Attr("class", func() string {
					if errors["birthDate"] != "" {
						return "error"
					}
					return ""
				}()),
				dom.OnInputInline(func(el dom.Element) {
					newInfo := info
					newInfo.BirthDate = el.Underlying().Get("value").String()
					mfs.personalInfo.Set(newInfo)
				}),
			),
			comps.Show(comps.ShowProps{
				When: reactivity.CreateMemo(func() bool {
					return errors["birthDate"] != ""
				}),
				Children: h.Div(
					h.Class("error-message"),
					g.Text(errors["birthDate"]),
				),
			}),
		),

		h.Div(
			h.Class("form-group"),
			h.Label(g.Text("Gender")),
			h.Select(
				h.Value(info.Gender),
				dom.OnChangeInline(func(el dom.Element) {
					newInfo := info
					newInfo.Gender = el.Underlying().Get("value").String()
					mfs.personalInfo.Set(newInfo)
				}),
				h.Option(h.Value(""), g.Text("Select..."), g.Attr("disabled", "true")),
				h.Option(h.Value("male"), g.Text("Male")),
				h.Option(h.Value("female"), g.Text("Female")),
				h.Option(h.Value("other"), g.Text("Other")),
				h.Option(h.Value("prefer-not-to-say"), g.Text("Prefer not to say")),
			),
		),
	)
}

func (mfs *MultiStepFormState) renderContactInfoStep() g.Node {
	info := mfs.contactInfo.Get()
	errors := mfs.errors.Get()

	return h.Div(
		h.Class("form-step"),

		h.H2(g.Text("Contact Information")),
		h.P(g.Text("How can we reach you?")),

		h.Div(
			h.Class("form-group"),
			h.Label(g.Text("Email *")),
			h.Input(
				h.Type("email"),
				h.Value(info.Email),
				g.Attr("class", func() string {
					if errors["email"] != "" {
						return "error"
					}
					return ""
				}()),
				dom.OnInputInline(func(el dom.Element) {
					newInfo := info
					newInfo.Email = el.Underlying().Get("value").String()
					mfs.contactInfo.Set(newInfo)
				}),
			),
			comps.Show(comps.ShowProps{
				When: reactivity.CreateMemo(func() bool {
					return errors["email"] != ""
				}),
				Children: h.Div(
					h.Class("error-message"),
					g.Text(errors["email"]),
				),
			}),
		),

		h.Div(
			h.Class("form-group"),
			h.Label(g.Text("Phone *")),
			h.Input(
				h.Type("tel"),
				h.Value(info.Phone),
				g.Attr("class", func() string {
					if errors["phone"] != "" {
						return "error"
					}
					return ""
				}()),
				dom.OnInputInline(func(el dom.Element) {
					newInfo := info
					newInfo.Phone = el.Underlying().Get("value").String()
					mfs.contactInfo.Set(newInfo)
				}),
			),
			comps.Show(comps.ShowProps{
				When: reactivity.CreateMemo(func() bool {
					return errors["phone"] != ""
				}),
				Children: h.Div(
					h.Class("error-message"),
					g.Text(errors["phone"]),
				),
			}),
		),

		h.Div(
			h.Class("form-group"),
			h.Label(g.Text("Address *")),
			h.Textarea(
				h.Value(info.Address),
				g.Attr("rows", "3"),
				g.Attr("class", func() string {
					if errors["address"] != "" {
						return "error"
					}
					return ""
				}()),
				dom.OnInputInline(func(el dom.Element) {
					newInfo := info
					newInfo.Address = el.Underlying().Get("value").String()
					mfs.contactInfo.Set(newInfo)
				}),
			),
			comps.Show(comps.ShowProps{
				When: reactivity.CreateMemo(func() bool {
					return errors["address"] != ""
				}),
				Children: h.Div(
					h.Class("error-message"),
					g.Text(errors["address"]),
				),
			}),
		),

		h.Div(
			h.Class("form-row"),
			h.Div(
				h.Class("form-group"),
				h.Style("flex: 1;"),
				h.Label(g.Text("City")),
				h.Input(
					h.Type("text"),
					h.Value(info.City),
					dom.OnInputInline(func(el dom.Element) {
						newInfo := info
						newInfo.City = el.Underlying().Get("value").String()
						mfs.contactInfo.Set(newInfo)
					}),
				),
			),

			h.Div(
				h.Class("form-group"),
				h.Style("flex: 1;"),
				h.Label(g.Text("Country")),
				h.Input(
					h.Type("text"),
					h.Value(info.Country),
					dom.OnInputInline(func(el dom.Element) {
						newInfo := info
						newInfo.Country = el.Underlying().Get("value").String()
						mfs.contactInfo.Set(newInfo)
					}),
				),
			),
		),
	)
}

func (mfs *MultiStepFormState) renderPreferencesStep() g.Node {
	prefs := mfs.preferences.Get()

	return h.Div(
		h.Class("form-step"),

		h.H2(g.Text("Preferences")),
		h.P(g.Text("Customize your experience")),

		h.Div(
			h.Class("form-group"),
			h.Label(
				h.Class("checkbox-group"),
				h.Input(
					h.Type("checkbox"),
					g.Attr("checked", fmt.Sprintf("%t", prefs.Newsletter)),
					dom.OnChangeInline(func(el dom.Element) {
						newPrefs := prefs
						newPrefs.Newsletter = el.Underlying().Get("checked").Bool()
						mfs.preferences.Set(newPrefs)
					}),
				),
				g.Text(" Subscribe to newsletter"),
			),
		),

		h.Div(
			h.Class("form-group"),
			h.Label(
				h.Class("checkbox-group"),
				h.Input(
					h.Type("checkbox"),
					g.Attr("checked", fmt.Sprintf("%t", prefs.Notifications)),
					dom.OnChangeInline(func(el dom.Element) {
						newPrefs := prefs
						newPrefs.Notifications = el.Underlying().Get("checked").Bool()
						mfs.preferences.Set(newPrefs)
					}),
				),
				g.Text(" Enable push notifications"),
			),
		),

		h.Div(
			h.Class("form-group"),
			h.Label(g.Text("Theme")),
			h.Select(
				h.Value(prefs.Theme),
				dom.OnChangeInline(func(el dom.Element) {
					newPrefs := prefs
					newPrefs.Theme = el.Underlying().Get("value").String()
					mfs.preferences.Set(newPrefs)
				}),
				h.Option(h.Value("light"), g.Text("Light")),
				h.Option(h.Value("dark"), g.Text("Dark")),
				h.Option(h.Value("auto"), g.Text("Auto")),
			),
		),

		h.Div(
			h.Class("form-group"),
			h.Label(g.Text("Language")),
			h.Select(
				h.Value(prefs.Language),
				dom.OnChangeInline(func(el dom.Element) {
					newPrefs := prefs
					newPrefs.Language = el.Underlying().Get("value").String()
					mfs.preferences.Set(newPrefs)
				}),
				h.Option(h.Value("en"), g.Text("English")),
				h.Option(h.Value("es"), g.Text("Spanish")),
				h.Option(h.Value("fr"), g.Text("French")),
				h.Option(h.Value("de"), g.Text("German")),
			),
		),
	)
}

func (mfs *MultiStepFormState) renderReviewStep() g.Node {
	personal := mfs.personalInfo.Get()
	contact := mfs.contactInfo.Get()
	prefs := mfs.preferences.Get()

	return h.Div(
		h.Class("form-step"),

		h.H2(g.Text("Review & Submit")),
		h.P(g.Text("Please review your information before submitting")),

		h.Div(
			h.Class("review-section"),
			h.H3(g.Text("Personal Information")),
			h.Div(
				h.Class("review-item"),
				h.Span(h.Class("review-label"), g.Text("Name:")),
				h.Span(h.Class("review-value"), g.Text(fmt.Sprintf("%s %s", personal.FirstName, personal.LastName))),
			),
			comps.Show(comps.ShowProps{
				When: reactivity.CreateMemo(func() bool {
					return personal.BirthDate != ""
				}),
				Children: h.Div(
					h.Class("review-item"),
					h.Span(h.Class("review-label"), g.Text("Birth Date:")),
					h.Span(h.Class("review-value"), g.Text(personal.BirthDate)),
				),
			}),
			comps.Show(comps.ShowProps{
				When: reactivity.CreateMemo(func() bool {
					return personal.Gender != ""
				}),
				Children: h.Div(
					h.Class("review-item"),
					h.Span(h.Class("review-label"), g.Text("Gender:")),
					h.Span(h.Class("review-value"), g.Text(strings.Title(personal.Gender))),
				),
			}),
		),

		h.Div(
			h.Class("review-section"),
			h.H3(g.Text("Contact Information")),
			h.Div(
				h.Class("review-item"),
				h.Span(h.Class("review-label"), g.Text("Email:")),
				h.Span(h.Class("review-value"), g.Text(contact.Email)),
			),
			h.Div(
				h.Class("review-item"),
				h.Span(h.Class("review-label"), g.Text("Phone:")),
				h.Span(h.Class("review-value"), g.Text(contact.Phone)),
			),
			comps.Show(comps.ShowProps{
				When: reactivity.CreateMemo(func() bool {
					return contact.Address != ""
				}),
				Children: h.Div(
					h.Class("review-item"),
					h.Span(h.Class("review-label"), g.Text("Address:")),
					h.Span(h.Class("review-value"), g.Text(contact.Address)),
				),
			}),
			comps.Show(comps.ShowProps{
				When: reactivity.CreateMemo(func() bool {
					return contact.City != "" || contact.Country != ""
				}),
				Children: h.Div(
					h.Class("review-item"),
					h.Span(h.Class("review-value"), g.Text(fmt.Sprintf("%s, %s", contact.City, contact.Country))),
				),
			}),
		),

		h.Div(
			h.Class("review-section"),
			h.H3(g.Text("Preferences")),
			h.Div(
				h.Class("review-item"),
				h.Span(h.Class("review-label"), g.Text("Newsletter:")),
				h.Span(h.Class("review-value"), g.Text(func() string {
					if prefs.Newsletter {
						return "Subscribed"
					}
					return "Not subscribed"
				}())),
			),
			h.Div(
				h.Class("review-item"),
				h.Span(h.Class("review-label"), g.Text("Notifications:")),
				h.Span(h.Class("review-value"), g.Text(func() string {
					if prefs.Notifications {
						return "Enabled"
					}
					return "Disabled"
				}())),
			),
			h.Div(
				h.Class("review-item"),
				h.Span(h.Class("review-label"), g.Text("Theme:")),
				h.Span(h.Class("review-value"), g.Text(strings.Title(prefs.Theme))),
			),
			h.Div(
				h.Class("review-item"),
				h.Span(h.Class("review-label"), g.Text("Language:")),
				h.Span(h.Class("review-value"), g.Text(strings.ToUpper(prefs.Language))),
			),
		),
	)
}

func (mfs *MultiStepFormState) renderSuccessMessage() g.Node {
	return h.Div(
		h.Class("success-message"),
		h.H2(g.Text("Registration Complete!")),
		h.P(g.Text("Thank you for completing the registration process.")),
		h.P(g.Text("We've received your information and will process it shortly.")),
		h.Button(
			h.Class("nav-button"),
			g.Text("Start Over"),
			dom.OnClickInline(func(el dom.Element) {
				// Reset form
				mfs.currentStep.Set(0)
				mfs.personalInfo.Set(PersonalInfo{})
				mfs.contactInfo.Set(ContactInfo{})
				mfs.preferences.Set(Preferences{Theme: "light", Language: "en"})
				mfs.errors.Set(make(map[string]string))
				mfs.submitted.Set(false)
			}),
		),
	)
}

func (mfs *MultiStepFormState) renderNavigation() g.Node {
	currentStep := mfs.currentStep.Get()
	isFirstStep := currentStep == 0
	isLastStep := currentStep == len(mfs.steps)-1

	return h.Div(
		h.Class("form-navigation"),

		h.Button(
			h.Class("nav-button prev"),
			g.Attr("disabled", fmt.Sprintf("%t", isFirstStep)),
			g.Text("Previous"),
			dom.OnClickInline(func(el dom.Element) {
				if currentStep > 0 {
					mfs.currentStep.Set(currentStep - 1)
				}
			}),
		),

		h.Button(
			h.Class(func() string {
				if isLastStep {
					return "nav-button submit"
				}
				return "nav-button next"
			}()),
			g.Text(func() string {
				if isLastStep {
					return "Submit"
				}
				return "Next"
			}()),
			dom.OnClickInline(func(el dom.Element) {
				// Validate current step
				if !mfs.steps[currentStep].IsValid() {
					return
				}

				if isLastStep {
					// Submit form
					mfs.submitForm()
				} else {
					// Go to next step
					mfs.currentStep.Set(currentStep + 1)
				}
			}),
		),
	)
}

func (mfs *MultiStepFormState) submitForm() {
	mfs.isSubmitting.Set(true)

	// Simulate API call
	go func() {
		time.Sleep(2 * time.Second)

		// Log the submitted data
		personal := mfs.personalInfo.Get()
		contact := mfs.contactInfo.Get()
		prefs := mfs.preferences.Get()

		logutil.Logf("Form submitted successfully!")
		logutil.Logf("Personal Info: %+v", personal)
		logutil.Logf("Contact Info: %+v", contact)
		logutil.Logf("Preferences: %+v", prefs)

		mfs.isSubmitting.Set(false)
		mfs.submitted.Set(true)
	}()
}

func main() {
	// Create form state
	formState := NewMultiStepFormState()

	// Mount the app and get a disposer function
	disposer := comps.Mount("app", func() g.Node { return formState.render() })
	_ = disposer // We don't use it in this example since the app runs indefinitely

	// Prevent exit
	select {}
}
