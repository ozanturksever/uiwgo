package form

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

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

// SubmissionHandler defines a function that handles form submission
type SubmissionHandler func(ctx context.Context, values map[string]any) error

// SubmissionOptions configures form submission behavior
type SubmissionOptions struct {
	URL     string            // URL to submit to (for HTTP submissions)
	Method  string            // HTTP method (GET, POST, PUT, etc.)
	Headers map[string]string // Additional headers
	Handler SubmissionHandler // Custom submission handler
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
	
	// crossFieldValidators stores validators that operate across multiple fields
	crossFieldValidators []CrossFieldValidator
	
	// submissionOptions configures how the form is submitted
	submissionOptions *SubmissionOptions
	
	// isSubmitting tracks whether a submission is in progress
	isSubmitting reactivity.Signal[bool]
	
	// submissionError tracks submission-specific errors
	submissionError reactivity.Signal[error]
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

	state.isSubmitting = reactivity.CreateSignal[bool](false)
	state.submissionError = reactivity.CreateSignal[error](nil)
	
	return state
}

// AddCrossFieldValidator adds a cross-field validator to the form
func (s *State) AddCrossFieldValidator(validator CrossFieldValidator) {
	s.crossFieldValidators = append(s.crossFieldValidators, validator)
}

// SetSubmissionOptions configures how the form should be submitted
func (s *State) SetSubmissionOptions(options SubmissionOptions) {
	s.submissionOptions = &options
}

// IsSubmitting returns whether a submission is currently in progress
func (s *State) IsSubmitting() bool {
	return s.isSubmitting.Get()
}

// GetSubmissionError returns the current submission error
func (s *State) GetSubmissionError() error {
	return s.submissionError.Get()
}

// ValidateWithCrossField performs full validation including cross-field validators
func (s *State) ValidateWithCrossField() bool {
	// First run regular field validation
	isValid := s.Validate()
	
	// Then run cross-field validation
	values := s.Values()
	for _, validator := range s.crossFieldValidators {
		if err := validator(values); err != nil {
			s.SetGlobalError(err)
			isValid = false
			break
		}
	}
	
	return isValid
}

// Submit validates and submits the form
func (s *State) Submit(ctx context.Context) error {
	// Prevent multiple simultaneous submissions
	if s.IsSubmitting() {
		return errors.New("form submission already in progress")
	}
	
	// Clear previous submission errors
	s.submissionError.Set(nil)
	s.isSubmitting.Set(true)
	
	defer func() {
		s.isSubmitting.Set(false)
	}()
	
	// Validate the form
	if !s.ValidateWithCrossField() {
		err := errors.New("form validation failed")
		s.submissionError.Set(err)
		return err
	}
	
	// Get form values
	values := s.Values()
	
	// Submit using configured options
	if s.submissionOptions == nil {
		err := errors.New("no submission options configured")
		s.submissionError.Set(err)
		return err
	}
	
	var err error
	if s.submissionOptions.Handler != nil {
		// Use custom handler
		err = s.submissionOptions.Handler(ctx, values)
	} else if s.submissionOptions.URL != "" {
		// Use HTTP submission
		err = s.submitHTTP(ctx, values)
	} else {
		err = errors.New("no submission handler or URL configured")
	}
	
	if err != nil {
		s.submissionError.Set(err)
		return err
	}
	
	return nil
}

// submitHTTP handles HTTP form submission
func (s *State) submitHTTP(ctx context.Context, values map[string]any) error {
	// Convert values to JSON
	jsonData, err := json.Marshal(values)
	if err != nil {
		return fmt.Errorf("failed to marshal form data: %w", err)
	}
	
	// Create HTTP request
	method := s.submissionOptions.Method
	if method == "" {
		method = "POST"
	}
	
	req, err := http.NewRequestWithContext(ctx, method, s.submissionOptions.URL, strings.NewReader(string(jsonData)))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	
	// Set headers
	req.Header.Set("Content-Type", "application/json")
	for key, value := range s.submissionOptions.Headers {
		req.Header.Set(key, value)
	}
	
	// Send request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to submit form: %w", err)
	}
	defer resp.Body.Close()
	
	// Check response status
	if resp.StatusCode >= 400 {
		return fmt.Errorf("form submission failed with status %d", resp.StatusCode)
	}
	
	return nil
}

// Reset clears all form values and errors
func (s *State) Reset() {
	for _, field := range s.schema {
		s.SetFieldValue(field.Name, field.InitialValue)
		s.SetFieldError(field.Name, nil)
	}
	s.SetGlobalError(nil)
	s.submissionError.Set(nil)
}

// NewState creates a new form state with the given schema
func NewState(schema []FieldDef) *State {
	return NewFromSchema(schema)
}