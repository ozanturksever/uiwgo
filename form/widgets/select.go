package widgets

import (
	. "maragu.dev/gomponents"
	"maragu.dev/gomponents/html"
	"github.com/ozanturksever/uiwgo/dom"
	"github.com/ozanturksever/uiwgo/form"
)

// SelectOption represents a single option in a select dropdown
type SelectOption struct {
	Value    string
	Label    string
	Disabled bool
}

// SelectOptions configures select dropdown behavior
type SelectOptions struct {
	Options     []SelectOption
	Placeholder string // Placeholder text for empty selection
	Class       string
	Disabled    bool
	Required    bool
	Multiple    bool // Whether to allow multiple selections
	Size        int  // Number of visible options (for multiple selects)
	Attributes  map[string]string
}

// SelectWidget creates a select dropdown widget bound to form state
// For single select: field value should be a string
// For multiple select: field value should be a slice of strings
func SelectWidget(state *form.State, fieldName string, opts SelectOptions) Node {
	value := state.GetFieldValue(fieldName)
	selectedValues := make(map[string]bool)
	
	// Parse current selected values
	if opts.Multiple {
		if value != nil {
			if slice, ok := value.([]string); ok {
				for _, v := range slice {
					selectedValues[v] = true
				}
			}
		}
	} else {
		if value != nil {
			if strVal, ok := value.(string); ok {
				selectedValues[strVal] = true
			}
		}
	}

	// Create options
	options := make([]Node, 0, len(opts.Options)+1)
	
	// Add placeholder option if provided and not multiple
	if opts.Placeholder != "" && !opts.Multiple {
		placeholderOption := html.Option(
			html.Value(""),
			html.Disabled(),
			If(len(selectedValues) == 0, html.Selected()),
			Text(opts.Placeholder),
		)
		options = append(options, placeholderOption)
	}
	
	// Add regular options
	for _, option := range opts.Options {
		isSelected := selectedValues[option.Value]
		
		optionNode := html.Option(
			html.Value(option.Value),
			If(isSelected, html.Selected()),
			If(option.Disabled, html.Disabled()),
			Text(option.Label),
		)
		
		options = append(options, optionNode)
	}

	// Create select element with inline event handler
	// Apply default Tailwind styling if no custom class provided
	className := opts.Class
	if className == "" {
		className = "w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500 transition-colors duration-200 bg-white"
	}
	
	selectElement := html.Select(
		html.Name(fieldName),
		html.ID(fieldName),
		html.Class(className),
		If(opts.Disabled, html.Disabled()),
		If(opts.Required, html.Required()),
		If(opts.Multiple, html.Multiple()),
		
		// Inline event handler for change events
		dom.OnChangeInline(func(el dom.Element) {
			if opts.Multiple {
				// Handle multiple selection
				selectedOptions := el.Underlying().Get("selectedOptions")
				length := selectedOptions.Get("length").Int()
				
				newValues := make([]string, 0, length)
				for i := 0; i < length; i++ {
					optionValue := selectedOptions.Index(i).Get("value").String()
					newValues = append(newValues, optionValue)
				}
				
				state.SetFieldValue(fieldName, newValues)
			} else {
				// Handle single selection
				selectedValue := el.Underlying().Get("value").String()
				state.SetFieldValue(fieldName, selectedValue)
			}
			
			// Trigger validation on change
			state.ValidateField(fieldName)
		}),
		Group(options),
	)

	// Add custom attributes if provided
	if opts.Attributes != nil {
		attrs := make([]Node, 0, len(opts.Attributes))
		for key, value := range opts.Attributes {
			attrs = append(attrs, Attr(key, value))
		}
		selectElement = html.Select(
			html.Name(fieldName),
			html.ID(fieldName),
			html.Class(className),
			If(opts.Disabled, html.Disabled()),
			If(opts.Required, html.Required()),
			If(opts.Multiple, html.Multiple()),
			dom.OnChangeInline(func(el dom.Element) {
				if opts.Multiple {
					selectedOptions := el.Underlying().Get("selectedOptions")
					length := selectedOptions.Get("length").Int()
					newValues := make([]string, 0, length)
					for i := 0; i < length; i++ {
						optionValue := selectedOptions.Index(i).Get("value").String()
						newValues = append(newValues, optionValue)
					}
					state.SetFieldValue(fieldName, newValues)
				} else {
					selectedValue := el.Underlying().Get("value").String()
					state.SetFieldValue(fieldName, selectedValue)
				}
				state.ValidateField(fieldName)
			}),
			Group(options),
			Group(attrs),
		)
	}

	return selectElement
}

// Datalist creates a datalist element for input autocomplete
// Used with text inputs to provide suggestions
type DatalistOptions struct {
	Options []SelectOption
	ID      string // Required ID for the datalist
}

func Datalist(opts DatalistOptions) Node {
	options := make([]Node, 0, len(opts.Options))
	
	for _, option := range opts.Options {
		optionNode := html.Option(
			html.Value(option.Value),

			If(option.Disabled, html.Disabled()),
		)
		
		options = append(options, optionNode)
	}

	return html.Div(
		html.ID(opts.ID),
		Group(options),
	)
}