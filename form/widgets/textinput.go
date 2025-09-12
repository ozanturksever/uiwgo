package widgets

import (
	"github.com/ozanturksever/uiwgo/dom"
	"github.com/ozanturksever/uiwgo/form"
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

// TextInput creates a text input widget bound to a form field
func TextInput(state *form.State, fieldName string, attrs ...Node) Node {
	value := state.GetFieldValue(fieldName)
	strValue := ""
	if value != nil {
		if s, ok := value.(string); ok {
			strValue = s
		}
	}
	
	return Input(
		append([]Node{
			Type("text"),
			Name(fieldName),
			ID(fieldName),
			Value(strValue),
			Class("w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm placeholder-gray-400 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500 transition-colors duration-200"),
			dom.OnInputInline(func(el dom.Element) {
			// Update form state when input changes
			newValue := el.Underlying().Get("value").String()
			state.SetFieldValue(fieldName, newValue)
			// Trigger validation for this field
			state.ValidateField(fieldName)
		}),
		}, attrs...)...,
	)
}

// PasswordInput creates a password input widget bound to a form field
func PasswordInput(state *form.State, fieldName string, attrs ...Node) Node {
	value := state.GetFieldValue(fieldName)
	strValue := ""
	if value != nil {
		if s, ok := value.(string); ok {
			strValue = s
		}
	}
	
	return Input(
		append([]Node{
			Type("password"),
			Name(fieldName),
			ID(fieldName),
			Value(strValue),
			Class("w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm placeholder-gray-400 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500 transition-colors duration-200"),
			dom.OnInputInline(func(el dom.Element) {
				// Update form state when input changes
				newValue := el.Underlying().Get("value").String()
				state.SetFieldValue(fieldName, newValue)
				// Trigger validation for this field
				state.ValidateField(fieldName)
			}),
		}, attrs...)...,
	)
}

// EmailInput creates an email input widget bound to a form field
func EmailInput(state *form.State, fieldName string, attrs ...Node) Node {
	value := state.GetFieldValue(fieldName)
	strValue := ""
	if value != nil {
		if s, ok := value.(string); ok {
			strValue = s
		}
	}
	
	return Input(
		append([]Node{
			Type("email"),
			Name(fieldName),
			ID(fieldName),
			Value(strValue),
			Class("w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm placeholder-gray-400 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500 transition-colors duration-200"),
			dom.OnInputInline(func(el dom.Element) {
			// Update form state when input changes
			newValue := el.Underlying().Get("value").String()
			state.SetFieldValue(fieldName, newValue)
			// Trigger validation for this field
			state.ValidateField(fieldName)
		}),
		}, attrs...)...,
	)
	}
	
	// TextArea creates a textarea widget bound to a form field
	func TextArea(state *form.State, fieldName string, attrs ...Node) Node {
		value := state.GetFieldValue(fieldName)
		strValue := ""
		if value != nil {
			if s, ok := value.(string); ok {
				strValue = s
			}
		}
		
		return Textarea(
			append([]Node{
				Name(fieldName),
				ID(fieldName),
				Text(strValue),
				Class("w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm placeholder-gray-400 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500 transition-colors duration-200 resize-vertical min-h-[100px]"),
				dom.OnInputInline(func(el dom.Element) {
					// Update form state when input changes
					newValue := el.Underlying().Get("value").String()
					state.SetFieldValue(fieldName, newValue)
					// Trigger validation for this field
					state.ValidateField(fieldName)
				}),
		}, attrs...)...,
	)
}