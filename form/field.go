package form

import (
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

// FieldOptions configures how a field is rendered
type FieldOptions struct {
	Label      string
	ShowLabel  bool
	ShowError  bool
	Widget     Widget
	Attributes []Node
	LabelAttrs []Node
	ErrorAttrs []Node
}

// Field renders a complete form field with label, widget, and error display
func Field(state *State, fieldName string, options FieldOptions) Node {
	// Find field definition
	var fieldDef *FieldDef
	for _, field := range state.schema {
		if field.Name == fieldName {
			fieldDef = &field
			break
		}
	}
	
	if fieldDef == nil {
		return Div(Text("Field not found: " + fieldName))
	}
	
	// Use field label if not overridden in options
	labelText := options.Label
	if labelText == "" {
		labelText = fieldDef.Label
	}
	
	// Use field widget if not overridden in options
	widget := options.Widget
	if widget == nil {
		widget = fieldDef.Widget
	}
	
	// Build field container
	var elements []Node
	
	// Add label if enabled
	if options.ShowLabel && labelText != "" {
		labelAttrs := append([]Node{
			For(fieldName),
		}, options.LabelAttrs...)
		
		elements = append(elements, Label(
			append(labelAttrs, Text(labelText))...,
		))
	}
	
	// Add widget
		if widget != nil {
			widgetAttrs := append(fieldDef.WidgetAttrs, options.Attributes...)
			elements = append(elements, widget(state, fieldName, widgetAttrs...))
		}
	
	// Add error display if enabled
	if options.ShowError {
		errorAttrs := append([]Node{
			Class("field-error"),
		}, options.ErrorAttrs...)
		
		// Create error element that reactively shows/hides based on field error
		var errorContent []Node
		if state.GetFieldError(fieldName) != nil {
			errorContent = append(errorContent, Text(state.GetFieldError(fieldName).Error()))
		}
		errorElement := Div(
			append(errorAttrs, errorContent...)...,
		)
		
		elements = append(elements, errorElement)
	}
	
	return Div(
		append([]Node{Class("form-field")}, elements...)...,
	)
}

// DefaultFieldOptions returns sensible defaults for field rendering
func DefaultFieldOptions() FieldOptions {
	return FieldOptions{
		ShowLabel: true,
		ShowError: true,
	}
}

// LabelOnlyField renders just the label for a field
func LabelOnlyField(state *State, fieldName string, attrs ...Node) Node {
	// Find field definition
	var fieldDef *FieldDef
	for _, field := range state.schema {
		if field.Name == fieldName {
			fieldDef = &field
			break
		}
	}
	
	if fieldDef == nil {
		return Text("Field not found: " + fieldName)
	}
	
	return Label(
		append(append([]Node{
			For(fieldName),
		}, attrs...), Text(fieldDef.Label))...,
	)
}

// WidgetOnlyField renders just the widget for a field
func WidgetOnlyField(state *State, fieldName string, attrs ...Node) Node {
	// Find field definition
	var fieldDef *FieldDef
	for _, field := range state.schema {
		if field.Name == fieldName {
			fieldDef = &field
			break
		}
	}
	
	if fieldDef == nil {
		return Div(Text("Field not found: " + fieldName))
	}
	
	if fieldDef.Widget == nil {
		return Div(Text("No widget defined for field: " + fieldName))
	}
	
	widgetAttrs := append(fieldDef.WidgetAttrs, attrs...)
	return fieldDef.Widget(state, fieldName, widgetAttrs...)
}

// ErrorOnlyField renders just the error display for a field
func ErrorOnlyField(state *State, fieldName string, attrs ...Node) Node {
	errorAttrs := append([]Node{
		Class("field-error"),
	}, attrs...)
	
	var errorContent []Node
	if state.GetFieldError(fieldName) != nil {
		errorContent = append(errorContent, Text(state.GetFieldError(fieldName).Error()))
	}
	return Div(
		append(errorAttrs, errorContent...)...,
	)
}