package widgets

import (
	. "maragu.dev/gomponents"
	"maragu.dev/gomponents/html"
	"github.com/ozanturksever/uiwgo/dom"
	"github.com/ozanturksever/uiwgo/form"
)

// RadioOption represents a single radio button option
type RadioOption struct {
	Value string
	Label string
}

// RadioGroupOptions configures radio button group behavior
type RadioGroupOptions struct {
	Options    []RadioOption
	Class      string
	LabelClass string
	Disabled   bool
	Required   bool
	Inline     bool // Whether to display options inline or stacked
	Attributes map[string]string
}

// RadioGroup creates a group of radio buttons for single selection
// The field value should be a string representing the selected value
func RadioGroup(state *form.State, fieldName string, opts RadioGroupOptions) Node {
	value := state.GetFieldValue(fieldName)
	selectedValue := ""
	if value != nil {
		if strVal, ok := value.(string); ok {
			selectedValue = strVal
		}
	}

	radioButtons := make([]Node, 0, len(opts.Options))
	
	for i, option := range opts.Options {
		radioID := fieldName + "_" + option.Value
		isSelected := selectedValue == option.Value
		
		// Create radio input with inline event handler and default styling
		radioInput := html.Input(
			html.Type("radio"),
			html.Name(fieldName),
			html.ID(radioID),
			html.Value(option.Value),
			html.Class("h-4 w-4 text-blue-600 focus:ring-blue-500 border-gray-300"),
			If(isSelected, html.Checked()),
			If(opts.Disabled, html.Disabled()),
			If(opts.Required && i == 0, html.Required()), // Only first radio needs required attribute
			// Inline event handler for change events
			dom.OnChangeInline(func(el dom.Element) {
				selectedValue := el.Underlying().Get("value").String()
				state.SetFieldValue(fieldName, selectedValue)
				// Trigger validation on change
				state.ValidateField(fieldName)
			}),
		)





		// Apply default label styling if no custom class provided
		labelClass := opts.LabelClass
		if labelClass == "" {
			labelClass = "flex items-center text-sm font-medium text-gray-700"
		}
		
		// Wrap radio button with label
		radioWrapper := html.Div(
			html.Class("flex items-center"),
			html.Label(
				html.For(radioID),
				html.Class(labelClass),
				radioInput,
				html.Span(
					html.Class("ml-2"),
					Text(option.Label),
				),
			),
		)
		
		radioButtons = append(radioButtons, radioWrapper)
	}

	// Apply default container styling based on inline option
	containerClass := opts.Class
	if containerClass == "" {
		if opts.Inline {
			containerClass = "flex flex-wrap gap-4"
		} else {
			containerClass = "space-y-2"
		}
	}

	// Create container with custom attributes if provided
	containerAttrs := []Node{
		html.Class(containerClass),
		Group(radioButtons),
	}
	
	if opts.Attributes != nil {
		for key, value := range opts.Attributes {
			containerAttrs = append(containerAttrs, Attr(key, value))
		}
	}
	
	return html.Div(containerAttrs...)
}

// RadioButton creates a single radio button (useful for custom layouts)
type RadioButtonOptions struct {
	Value      string
	Label      string
	Class      string
	LabelClass string
	Disabled   bool
	Required   bool
	Attributes map[string]string
}

func RadioButton(state *form.State, fieldName string, opts RadioButtonOptions) Node {
	value := state.GetFieldValue(fieldName)
	isSelected := false
	if value != nil {
		if strVal, ok := value.(string); ok {
			isSelected = strVal == opts.Value
		}
	}

	radioID := fieldName + "_" + opts.Value
	
	// Apply default styling if no custom class provided
	radioClass := opts.Class
	if radioClass == "" {
		radioClass = "h-4 w-4 text-blue-600 focus:ring-blue-500 border-gray-300"
	}
	
	// Create radio input with inline event handler
	radioInput := html.Input(
		html.Type("radio"),
		html.Name(fieldName),
		html.ID(radioID),
		html.Value(opts.Value),
		html.Class(radioClass),
		If(isSelected, html.Checked()),
			If(opts.Disabled, html.Disabled()),
			If(opts.Required, html.Required()),
			// Inline event handler for change events
			dom.OnChangeInline(func(el dom.Element) {
				selectedValue := el.Underlying().Get("value").String()
				state.SetFieldValue(fieldName, selectedValue)
				// Trigger validation on change
				state.ValidateField(fieldName)
			}),
	)

	// Add custom attributes if provided
	if opts.Attributes != nil {
		attrs := make([]Node, 0, len(opts.Attributes))
		for key, value := range opts.Attributes {
			attrs = append(attrs, Attr(key, value))
		}
		radioInput = html.Input(
			html.Type("radio"),
			html.Name(fieldName),
			html.ID(fieldName+"_"+opts.Value),
			html.Value(opts.Value),
			html.Class(radioClass),
			If(isSelected, html.Checked()),
			If(opts.Disabled, html.Disabled()),
			If(opts.Required, html.Required()),
			dom.OnChangeInline(func(el dom.Element) {
				selectedValue := el.Underlying().Get("value").String()
				state.SetFieldValue(fieldName, selectedValue)
				state.ValidateField(fieldName)
			}),
			Group(attrs),
		)
	}

	// If no label is provided, return just the radio button
	if opts.Label == "" {
		return radioInput
	}

	// Apply default label styling if no custom class provided
	labelClass := opts.LabelClass
	if labelClass == "" {
		labelClass = "flex items-center text-sm font-medium text-gray-700"
	}
	
	// Return radio button with label
	return html.Label(
		html.For(radioID),
		html.Class(labelClass),
		radioInput,
		html.Span(
			html.Class("ml-2"),
			Text(opts.Label),
		),
	)
}