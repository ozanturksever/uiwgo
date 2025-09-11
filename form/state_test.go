package form

import (
	"errors"
	"testing"

	"github.com/ozanturksever/uiwgo/reactivity"
)

func TestNewFromSchema(t *testing.T) {
	t.Run("creates state with reactive signals for each field", func(t *testing.T) {
		// Define a simple schema
		schema := []FieldDef{
			{Name: "username", Label: "Username"},
			{Name: "email", Label: "Email"},
			{Name: "age", Label: "Age"},
		}

		// Create state from schema
		state := NewFromSchema(schema)

		// Verify state is not nil
		if state == nil {
			t.Fatal("NewFromSchema returned nil")
		}

		// Verify field values are initialized
		if len(state.fieldValues) != 3 {
			t.Errorf("Expected 3 field values, got %d", len(state.fieldValues))
		}

		// Verify field errors are initialized
		if len(state.fieldErrors) != 3 {
			t.Errorf("Expected 3 field errors, got %d", len(state.fieldErrors))
		}

		// Verify specific fields exist
		for _, field := range schema {
			if _, exists := state.fieldValues[field.Name]; !exists {
				t.Errorf("Field value signal for %s not found", field.Name)
			}
			if _, exists := state.fieldErrors[field.Name]; !exists {
				t.Errorf("Field error signal for %s not found", field.Name)
			}
		}

		// Verify global error signal is initialized
		if state.globalError == nil {
			t.Error("Global error signal not initialized")
		}
	})

	t.Run("initializes field values to empty strings", func(t *testing.T) {
		schema := []FieldDef{
			{Name: "username", Label: "Username"},
			{Name: "email", Label: "Email"},
		}

		state := NewFromSchema(schema)

		// Check initial values
		usernameValue := state.fieldValues["username"].Get()
		if usernameValue != "" {
			t.Errorf("Expected empty string for username, got %v", usernameValue)
		}

		emailValue := state.fieldValues["email"].Get()
		if emailValue != "" {
			t.Errorf("Expected empty string for email, got %v", emailValue)
		}
	})

	t.Run("initializes field errors to nil", func(t *testing.T) {
		schema := []FieldDef{
			{Name: "username", Label: "Username"},
		}

		state := NewFromSchema(schema)

		// Check initial error state
		usernameError := state.fieldErrors["username"].Get()
		if usernameError != nil {
			t.Errorf("Expected nil error for username, got %v", usernameError)
		}
	})

	t.Run("initializes global error to nil", func(t *testing.T) {
		schema := []FieldDef{
			{Name: "username", Label: "Username"},
		}

		state := NewFromSchema(schema)

		// Check initial global error state
		globalError := state.globalError.Get()
		if globalError != nil {
			t.Errorf("Expected nil global error, got %v", globalError)
		}
	})
}

func TestState_GetValue(t *testing.T) {
	t.Run("returns current field value", func(t *testing.T) {
		schema := []FieldDef{{Name: "username", Label: "Username"}}
		state := NewFromSchema(schema)

		// Set a value
		state.fieldValues["username"].Set("testuser")

		// Get the value
		value := state.GetFieldValue("username")
		if value != "testuser" {
			t.Errorf("Expected 'testuser', got %v", value)
		}
	})

	t.Run("returns nil for non-existent field", func(t *testing.T) {
		schema := []FieldDef{{Name: "username", Label: "Username"}}
		state := NewFromSchema(schema)

		value := state.GetFieldValue("nonexistent")
		if value != nil {
			t.Errorf("Expected nil for non-existent field, got %v", value)
		}
	})
}

func TestState_SetValue(t *testing.T) {
	t.Run("sets field value and triggers reactivity", func(t *testing.T) {
		schema := []FieldDef{{Name: "username", Label: "Username"}}
		state := NewFromSchema(schema)

		// Track if signal was triggered
		var triggered bool
		reactivity.CreateEffect(func() {
			state.fieldValues["username"].Get()
			triggered = true
		})

		// Reset trigger flag
		triggered = false

		// Set value
		state.SetFieldValue("username", "newuser")

		// Verify value was set
		value := state.GetFieldValue("username")
		if value != "newuser" {
			t.Errorf("Expected 'newuser', got %v", value)
		}

		// Verify reactivity was triggered
		if !triggered {
			t.Error("Expected effect to be triggered when value changed")
		}
	})

	t.Run("does nothing for non-existent field", func(t *testing.T) {
		schema := []FieldDef{{Name: "username", Label: "Username"}}
		state := NewFromSchema(schema)

		// This should not panic
		state.SetFieldValue("nonexistent", "value")
	})
}

func TestState_GetError(t *testing.T) {
	t.Run("returns current field error", func(t *testing.T) {
		schema := []FieldDef{{Name: "username", Label: "Username"}}
		state := NewFromSchema(schema)

		// Set an error
		testError := errors.New("validation failed")
		state.fieldErrors["username"].Set(testError)

		// Get the error
		err := state.GetFieldError("username")
		if err != testError {
			t.Errorf("Expected test error, got %v", err)
		}
	})

	t.Run("returns nil for non-existent field", func(t *testing.T) {
		schema := []FieldDef{{Name: "username", Label: "Username"}}
		state := NewFromSchema(schema)

		err := state.GetFieldError("nonexistent")
		if err != nil {
			t.Errorf("Expected nil for non-existent field, got %v", err)
		}
	})
}

func TestState_SetError(t *testing.T) {
	t.Run("sets field error and triggers reactivity", func(t *testing.T) {
		schema := []FieldDef{{Name: "username", Label: "Username"}}
		state := NewFromSchema(schema)

		// Track if signal was triggered
		var triggered bool
		reactivity.CreateEffect(func() {
			state.fieldErrors["username"].Get()
			triggered = true
		})

		// Reset trigger flag
		triggered = false

		// Set error
		testError := errors.New("validation failed")
		state.SetFieldError("username", testError)

		// Verify error was set
		err := state.GetFieldError("username")
		if err != testError {
			t.Errorf("Expected test error, got %v", err)
		}

		// Verify reactivity was triggered
		if !triggered {
			t.Error("Expected effect to be triggered when error changed")
		}
	})

	t.Run("does nothing for non-existent field", func(t *testing.T) {
		schema := []FieldDef{{Name: "username", Label: "Username"}}
		state := NewFromSchema(schema)

		// This should not panic
		testError := errors.New("test")
		state.SetFieldError("nonexistent", testError)
	})
}

func TestState_GetGlobalError(t *testing.T) {
	t.Run("returns current global error", func(t *testing.T) {
		schema := []FieldDef{{Name: "username", Label: "Username"}}
		state := NewFromSchema(schema)

		// Set global error
		testError := errors.New("form submission failed")
		state.SetGlobalError(testError)

		// Get the error
		err := state.GetGlobalError()
		if err != testError {
			t.Errorf("Expected test error, got %v", err)
		}
	})
}

func TestState_SetGlobalError(t *testing.T) {
	t.Run("sets global error and triggers reactivity", func(t *testing.T) {
		schema := []FieldDef{{Name: "username", Label: "Username"}}
		state := NewFromSchema(schema)

		// Track if signal was triggered
		var triggered bool
		reactivity.CreateEffect(func() {
			state.globalError.Get()
			triggered = true
		})

		// Reset trigger flag
		triggered = false

		// Set global error
		testError := errors.New("form submission failed")
		state.SetGlobalError(testError)

		// Verify error was set
		err := state.GetGlobalError()
		if err != testError {
			t.Errorf("Expected test error, got %v", err)
		}

		// Verify reactivity was triggered
		if !triggered {
			t.Error("Expected effect to be triggered when global error changed")
		}
	})
}