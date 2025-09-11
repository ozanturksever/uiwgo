package widgets

import (
	"github.com/ozanturksever/uiwgo/form"
	. "maragu.dev/gomponents"
	"maragu.dev/gomponents/html"
	"github.com/ozanturksever/uiwgo/dom"
)

// CheckboxOptions configures checkbox widget behavior
type CheckboxOptions struct {
	Label       string
	Class       string
	LabelClass  string
	Disabled    bool
	Required    bool
	Attributes  map[string]string
}

// Checkbox creates a checkbox input widget bound to form state
// The field value should be a boolean (true/false)
func Checkbox(state *form.State, fieldName string, opts CheckboxOptions) Node {
	value := state.GetFieldValue(fieldName)
	checked := false
	if value != nil {
		if boolVal, ok := value.(bool); ok {
			checked = boolVal
		}
	}

	// Apply default Tailwind styling if no custom class provided
	checkboxClass := opts.Class
	if checkboxClass == "" {
		checkboxClass = "h-4 w-4 text-blue-600 focus:ring-blue-500 border-gray-300 rounded"
	}
	
	// Create checkbox input with inline event handler
	checkboxInput := html.Input(
		html.Type("checkbox"),
		html.Name(fieldName),
		html.ID(fieldName),
		html.Class(checkboxClass),
		If(checked, html.Checked()),
		If(opts.Disabled, html.Disabled()),
		If(opts.Required, html.Required()),
		// Inline event handler for change events
		dom.OnChangeInline(func(el dom.Element) {
			checked := el.Underlying().Get("checked").Bool()
			if checked {
				state.SetFieldValue(fieldName, "true")
			} else {
				state.SetFieldValue(fieldName, "false")
			}
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
		checkboxInput = html.Input(
			html.Type("checkbox"),
			html.Name(fieldName),
			html.ID(fieldName),
			html.Class(checkboxClass),
			If(checked, html.Checked()),
			If(opts.Disabled, html.Disabled()),
			If(opts.Required, html.Required()),
			dom.OnChangeInline(func(el dom.Element) {
				checked := el.Underlying().Get("checked").Bool()
				if checked {
					state.SetFieldValue(fieldName, "true")
				} else {
					state.SetFieldValue(fieldName, "false")
				}
				state.ValidateField(fieldName)
			}),
			Group(attrs),
		)
	}

	// If no label is provided, return just the checkbox
	if opts.Label == "" {
		return checkboxInput
	}

	// Apply default label styling if no custom class provided
	labelClass := opts.LabelClass
	if labelClass == "" {
		labelClass = "flex items-center text-sm font-medium text-gray-700"
	}
	
	// Return checkbox with label
	return html.Div(
		html.Class("flex items-center space-x-2"),
		html.Label(
			html.For(fieldName),
			html.Class(labelClass),
			checkboxInput,
			html.Span(
				html.Class("ml-2"),
				Text(opts.Label),
			),
		),
	)
}

// CheckboxGroup creates a group of related checkboxes for multi-select
// The field value should be a slice of strings representing selected values
type CheckboxGroupOption struct {
	Value string
	Label string
}

type CheckboxGroupOptions struct {
	Options    []CheckboxGroupOption
	Class      string
	LabelClass string
	Disabled   bool
	Required   bool
}

func CheckboxGroup(state *form.State, fieldName string, opts CheckboxGroupOptions) Node {
	value := state.GetFieldValue(fieldName)
	selectedValues := make(map[string]bool)
	
	// Parse current selected values
	if value != nil {
		if slice, ok := value.([]string); ok {
			for _, v := range slice {
				selectedValues[v] = true
			}
		}
	}

	checkboxes := make([]Node, 0, len(opts.Options))
	
	for _, option := range opts.Options {
		checkboxID := fieldName + "_" + option.Value
		isChecked := selectedValues[option.Value]
		
		checkbox := html.Div(
			html.Class("checkbox-group-item"),
			html.Label(
				html.For(checkboxID),
				html.Class(opts.LabelClass),
				html.Input(
					html.Type("checkbox"),
					html.Name(fieldName+"[]"),
					html.ID(checkboxID),
					html.Value(option.Value),
					If(isChecked, html.Checked()),
				If(opts.Disabled, html.Disabled()),
				If(opts.Required, html.Required()),
					// Inline event handler for checkbox group changes
					dom.OnChangeInline(func(el dom.Element) {
						checkboxValue := el.Underlying().Get("value").String()
						isChecked := el.Underlying().Get("checked").Bool()
						
						// Get current selected values
						currentValue := state.GetFieldValue(fieldName)
						currentSelected := make([]string, 0)
						if currentValue != nil {
							if slice, ok := currentValue.([]string); ok {
								currentSelected = slice
							}
						}
						
						// Update selected values
						newSelected := make([]string, 0)
						for _, v := range currentSelected {
							if v != checkboxValue {
								newSelected = append(newSelected, v)
							}
						}
						if isChecked {
							newSelected = append(newSelected, checkboxValue)
						}
						
						state.SetFieldValue(fieldName, newSelected)
						// Trigger validation on change
						state.ValidateField(fieldName)
					}),
				),
				Text(" "+option.Label),
			),
		)
		
		checkboxes = append(checkboxes, checkbox)
	}

	return html.Div(
		html.Class("checkbox-group "+opts.Class),
		Group(checkboxes),
	)
}