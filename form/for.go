package form

import (
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

// ForOptions configures the form submission behavior
type ForOptions struct {
	Method      string // "GET" or "POST", defaults to "POST"
	Action      string // Form action URL
	OnSubmit    func(*State) error // Custom submit handler
	ValidateOnSubmit bool // Whether to validate before submission, defaults to true
	Attributes  []Node // Additional form attributes
}

// FormFor creates a form element with submission handling and validation
func FormFor(state *State, options ForOptions, children ...Node) Node {
	// Set defaults
	method := options.Method
	if method == "" {
		method = "POST"
	}
	

	
	// Build form attributes
	formAttrs := []Node{
		Method(method),
	}
	
	if options.Action != "" {
		formAttrs = append(formAttrs, Action(options.Action))
	}
	
	// Add custom submit handler
	if options.OnSubmit != nil {
		// TODO: OnSubmit handler will be implemented with JS event binding
		// For now, we'll add a data attribute to identify the form
		formAttrs = append(formAttrs, Attr("data-has-submit-handler", "true"))
	}
	
	// Add additional attributes
	formAttrs = append(formAttrs, options.Attributes...)
	
	return Form(
		append(formAttrs, children...)...,
	)
}

// DefaultForOptions returns sensible defaults for form options
func DefaultForOptions() ForOptions {
	return ForOptions{
		Method:           "POST",
		ValidateOnSubmit: true,
	}
}

// SimpleForm creates a basic form with default options
func SimpleForm(state *State, onSubmit func(*State) error, children ...Node) Node {
	options := DefaultForOptions()
	options.OnSubmit = onSubmit
	
	return FormFor(state, options, children...)
}

// GetFormData extracts all field values from the form state as a map
func GetFormData(state *State) map[string]string {
	data := make(map[string]string)
	
	for _, field := range state.schema {
		value := state.GetFieldValue(field.Name)
		if strValue, ok := value.(string); ok {
			data[field.Name] = strValue
		} else {
			data[field.Name] = ""
		}
	}
	
	return data
}

// SetFormData sets multiple field values from a map
func SetFormData(state *State, data map[string]string) {
	for fieldName, value := range data {
		state.SetFieldValue(fieldName, value)
	}
}

// ClearForm resets all field values and errors
func ClearForm(state *State) {
	// Clear all field values
	for _, field := range state.schema {
		state.SetFieldValue(field.Name, "")
		state.SetFieldError(field.Name, nil)
	}
	
	// Clear global error
	state.SetGlobalError(nil)
}

// ResetForm resets all fields to their initial values and clears errors
func ResetForm(state *State) {
	// Reset all field values to initial values
	for _, field := range state.schema {
		state.SetFieldValue(field.Name, field.InitialValue)
		state.SetFieldError(field.Name, nil)
	}
	
	// Clear global error
	state.SetGlobalError(nil)
}