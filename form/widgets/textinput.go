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
			dom.OnInputInline(func(el dom.Element) {
			// Update form state when input changes
			newValue := el.Underlying().Get("value").String()
			state.SetFieldValue(fieldName, newValue)
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
			dom.OnInputInline(func(el dom.Element) {
				// Update form state when input changes
				newValue := el.Underlying().Get("value").String()
				state.SetFieldValue(fieldName, newValue)
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
			dom.OnInputInline(func(el dom.Element) {
				// Update form state when input changes
				newValue := el.Underlying().Get("value").String()
				state.SetFieldValue(fieldName, newValue)
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
			dom.OnInputInline(func(el dom.Element) {
				// Update form state when input changes
				newValue := el.Underlying().Get("value").String()
				state.SetFieldValue(fieldName, newValue)
			}),
		}, attrs...)...,
	)
}