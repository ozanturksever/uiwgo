package form

import (
	"github.com/ozanturksever/uiwgo/reactivity"
	. "maragu.dev/gomponents"
)

// Validator is a function that validates a single field value.
// It returns an error if the value is invalid, nil otherwise.
type Validator func(value any) error

// CrossFieldValidator is a function that validates across multiple fields.
// It receives a map of all form values and returns an error if validation fails.
type CrossFieldValidator func(values map[string]any) error

// Widget is a function that renders a form field.
// It takes the form state, field name, and optional attributes and returns a gomponents Node.
type Widget func(state *State, fieldName string, attrs ...Node) Node

// FieldDef defines the structure and behavior of a single form field.
type FieldDef struct {
	// Name is the programmatic name of the field (e.g., "user_email")
	Name string
	
	// Label is the user-visible label for the field
	Label string
	
	// InitialValue is the default value for the field
	InitialValue any
	
	// Validators is a slice of per-field validation functions
	Validators []Validator
	
	// Widget is the function responsible for rendering the field's UI
	Widget Widget
	
	// WidgetAttrs are optional attributes to pass to the widget
	WidgetAttrs []Node
}

// State represents the current state of a form, including field values and errors.
type State struct {
	// schema holds the form schema definition
	schema []FieldDef
	
	// fieldValues holds reactive signals for each field's value
	fieldValues map[string]reactivity.Signal[any]
	
	// fieldErrors holds reactive signals for each field's error state
	fieldErrors map[string]reactivity.Signal[error]
	
	// globalError holds the reactive signal for form-wide errors
	globalError reactivity.Signal[error]
}

// Values returns a map of all current field values.
func (s *State) Values() map[string]any {
	values := make(map[string]any)
	for name, signal := range s.fieldValues {
		values[name] = signal.Get()
	}
	return values
}

// GetFieldValue returns the current value of a specific field.
func (s *State) GetFieldValue(fieldName string) any {
	if signal, exists := s.fieldValues[fieldName]; exists {
		return signal.Get()
	}
	return nil
}

// SetFieldValue sets the value of a specific field.
func (s *State) SetFieldValue(fieldName string, value any) {
	if signal, exists := s.fieldValues[fieldName]; exists {
		signal.Set(value)
	}
}

// GetFieldError returns the current error for a specific field.
func (s *State) GetFieldError(fieldName string) error {
	if signal, exists := s.fieldErrors[fieldName]; exists {
		return signal.Get()
	}
	return nil
}

// SetFieldError sets the error for a specific field.
func (s *State) SetFieldError(fieldName string, err error) {
	if signal, exists := s.fieldErrors[fieldName]; exists {
		signal.Set(err)
	}
}

// GetGlobalError returns the current global form error.
func (s *State) GetGlobalError() error {
	return s.globalError.Get()
}

// SetGlobalError sets the global form error
func (s *State) SetGlobalError(err error) {
	s.globalError.Set(err)
}

// ValidateField validates a single field using its validators
func (s *State) ValidateField(fieldName string) error {
	// Find the field definition
	var fieldDef *FieldDef
	for _, field := range s.schema {
		if field.Name == fieldName {
			fieldDef = &field
			break
		}
	}
	
	if fieldDef == nil {
		return nil // Field not found, no validation
	}
	
	// Get current field value
	value := s.GetFieldValue(fieldName)
	
	// Run field validators
	for _, validator := range fieldDef.Validators {
		if err := validator(value); err != nil {
			s.SetFieldError(fieldName, err)
			return err
		}
	}
	
	// Clear field error if validation passed
	s.SetFieldError(fieldName, nil)
	return nil
}

// Validate validates all fields and runs cross-field validation
func (s *State) Validate() bool {
	isValid := true
	
	// Validate all fields
	for _, field := range s.schema {
		if err := s.ValidateField(field.Name); err != nil {
			isValid = false
		}
	}
	
	// TODO: Run cross-field validators when implemented
	// This would require adding CrossFieldValidators to the schema
	
	return isValid
}

// GetSchema returns the form's field definitions.
func (s *State) GetSchema() []FieldDef {
	return s.schema
}

// GetFieldDef returns the field definition for a specific field name.
func (s *State) GetFieldDef(fieldName string) *FieldDef {
	for i := range s.schema {
		if s.schema[i].Name == fieldName {
			return &s.schema[i]
		}
	}
	return nil
}

// NewFromSchema creates a new form state from a schema definition.
// It initializes reactive signals for each field's value and error state.
func NewFromSchema(schema []FieldDef) *State {
	state := &State{
		schema: schema,
		fieldValues: make(map[string]reactivity.Signal[any]),
		fieldErrors: make(map[string]reactivity.Signal[error]),
		globalError: reactivity.CreateSignal[error](nil),
	}

	// Initialize signals for each field
	for _, field := range schema {
		state.fieldValues[field.Name] = reactivity.CreateSignal[any]("")
		state.fieldErrors[field.Name] = reactivity.CreateSignal[error](nil)
	}

	return state
}