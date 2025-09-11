package form

import (
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

// RenderOptions configures how the entire form is rendered
type RenderOptions struct {
	Layout        string // "vertical", "horizontal", "inline"
	ShowLabels    bool   // Whether to show field labels
	ShowErrors    bool   // Whether to show field errors
	SubmitButton  Node   // Custom submit button, if nil a default is used
	ResetButton   Node   // Custom reset button, if nil none is shown
	FormAttrs     []Node // Attributes for the form element
	ContainerAttrs []Node // Attributes for the form container
	FieldAttrs    []Node // Default attributes for field containers
}

// Render creates a complete form with all fields using the specified layout
func Render(state *State, options RenderOptions, onSubmit func(*State) error) Node {
	// Set defaults
	layout := options.Layout
	if layout == "" {
		layout = "vertical"
	}
	
	showLabels := options.ShowLabels
	showErrors := options.ShowErrors
	if !options.ShowLabels && !options.ShowErrors {
		// Default to showing both if not specified
		showLabels = true
		showErrors = true
	}
	
	// Build form fields
	var fields []Node
	
	for _, fieldDef := range state.schema {
		fieldOptions := FieldOptions{
			ShowLabel: showLabels,
			ShowError: showErrors,
			Attributes: options.FieldAttrs,
		}
		
		// Apply layout-specific styling
		switch layout {
		case "horizontal":
			fieldOptions.LabelAttrs = append(fieldOptions.LabelAttrs, Class("form-label-horizontal"))
			fieldOptions.Attributes = append(fieldOptions.Attributes, Class("form-field-horizontal"))
		case "inline":
			fieldOptions.LabelAttrs = append(fieldOptions.LabelAttrs, Class("form-label-inline"))
			fieldOptions.Attributes = append(fieldOptions.Attributes, Class("form-field-inline"))
		default: // vertical
			fieldOptions.LabelAttrs = append(fieldOptions.LabelAttrs, Class("form-label-vertical"))
			fieldOptions.Attributes = append(fieldOptions.Attributes, Class("form-field-vertical"))
		}
		
		fields = append(fields, Field(state, fieldDef.Name, fieldOptions))
	}
	
	// Add global error display
	var globalErrorContent []Node
	if state.GetGlobalError() != nil {
		globalErrorContent = append(globalErrorContent, Div(
			Class("alert alert-error"),
			Text(state.GetGlobalError().Error()),
		))
	}
	globalErrorElement := Div(
		append([]Node{Class("form-global-error")}, globalErrorContent...)...,
	)
	
	// Add buttons
	var buttons []Node
	
	// Submit button
	submitBtn := options.SubmitButton
	if submitBtn == nil {
		submitBtn = Button(
			Type("submit"),
			Class("btn btn-primary"),
			Text("Submit"),
		)
	}
	buttons = append(buttons, submitBtn)
	
	// Reset button
	if options.ResetButton != nil {
		buttons = append(buttons, options.ResetButton)
	}
	
	buttonContainer := Div(
		append([]Node{Class("form-buttons")}, buttons...)...,
	)
	
	// Build form content
	formContent := append([]Node{globalErrorElement}, fields...)
	formContent = append(formContent, buttonContainer)
	
	// Create form with submission handling
	forOptions := ForOptions{
		OnSubmit:         onSubmit,
		ValidateOnSubmit: true,
		Attributes:       options.FormAttrs,
	}
	
	form := FormFor(state, forOptions, formContent...)
	
	// Wrap in container if needed
	if len(options.ContainerAttrs) > 0 {
		containerAttrs := append([]Node{
			Class("form-container form-layout-" + layout),
		}, options.ContainerAttrs...)
		
		return Div(
			append(containerAttrs, form)...,
		)
	}
	
	return Div(
		Class("form-container form-layout-" + layout),
		form,
	)
}

// DefaultRenderOptions returns sensible defaults for form rendering
func DefaultRenderOptions() RenderOptions {
	return RenderOptions{
		Layout:     "vertical",
		ShowLabels: true,
		ShowErrors: true,
	}
}

// VerticalForm creates a form with vertical layout (default)
func VerticalForm(state *State, onSubmit func(*State) error) Node {
	options := DefaultRenderOptions()
	options.Layout = "vertical"
	return Render(state, options, onSubmit)
}

// HorizontalForm creates a form with horizontal layout (labels beside fields)
func HorizontalForm(state *State, onSubmit func(*State) error) Node {
	options := DefaultRenderOptions()
	options.Layout = "horizontal"
	return Render(state, options, onSubmit)
}

// InlineForm creates a form with inline layout (all fields in a row)
func InlineForm(state *State, onSubmit func(*State) error) Node {
	options := DefaultRenderOptions()
	options.Layout = "inline"
	return Render(state, options, onSubmit)
}

// QuickForm creates a simple vertical form with minimal configuration
func QuickForm(state *State, onSubmit func(*State) error, submitText string) Node {
	options := DefaultRenderOptions()
	
	if submitText != "" {
		options.SubmitButton = Button(
			Type("submit"),
			Class("btn btn-primary"),
			Text(submitText),
		)
	}
	
	return Render(state, options, onSubmit)
}