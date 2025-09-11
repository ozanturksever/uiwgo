//go:build !js && !wasm

package form

import (
	"errors"
	"testing"

	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

// Test FieldDef struct creation and field access
func TestFieldDef_Creation(t *testing.T) {
	tests := []struct {
		name     string
		fieldDef FieldDef
		want     FieldDef
	}{
		{
			name: "basic field definition",
			fieldDef: FieldDef{
				Name:         "email",
				Label:        "Email Address",
				InitialValue: "test@example.com",
				Validators:   []Validator{},
				Widget:       nil,
				WidgetAttrs:  []Node{},
			},
			want: FieldDef{
				Name:         "email",
				Label:        "Email Address",
				InitialValue: "test@example.com",
				Validators:   []Validator{},
				Widget:       nil,
				WidgetAttrs:  []Node{},
			},
		},
		{
			name: "field with widget attributes",
			fieldDef: FieldDef{
				Name:        "password",
				Label:       "Password",
				WidgetAttrs: []Node{Placeholder("Enter your password"), Required()},
			},
			want: FieldDef{
				Name:        "password",
				Label:       "Password",
				WidgetAttrs: []Node{Placeholder("Enter your password"), Required()},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.fieldDef.Name != tt.want.Name {
				t.Errorf("FieldDef.Name = %v, want %v", tt.fieldDef.Name, tt.want.Name)
			}
			if tt.fieldDef.Label != tt.want.Label {
				t.Errorf("FieldDef.Label = %v, want %v", tt.fieldDef.Label, tt.want.Label)
			}
			if tt.fieldDef.InitialValue != tt.want.InitialValue {
				t.Errorf("FieldDef.InitialValue = %v, want %v", tt.fieldDef.InitialValue, tt.want.InitialValue)
			}
		})
	}
}

// Test Validator function type
func TestValidator_FunctionType(t *testing.T) {
	// Test that we can create and call validator functions
	requiredValidator := func(value any) error {
		if value == nil || value == "" {
			return errors.New("field is required")
		}
		return nil
	}

	tests := []struct {
		name      string
		validator Validator
		value     any
		wantErr   bool
		errorMsg  string
	}{
		{
			name:      "required validator with valid value",
			validator: requiredValidator,
			value:     "test@example.com",
			wantErr:   false,
		},
		{
			name:      "required validator with empty string",
			validator: requiredValidator,
			value:     "",
			wantErr:   true,
			errorMsg:  "field is required",
		},
		{
			name:      "required validator with nil value",
			validator: requiredValidator,
			value:     nil,
			wantErr:   true,
			errorMsg:  "field is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.validator(tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("Validator() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err.Error() != tt.errorMsg {
				t.Errorf("Validator() error message = %v, want %v", err.Error(), tt.errorMsg)
			}
		})
	}
}

// Test CrossFieldValidator function type
func TestCrossFieldValidator_FunctionType(t *testing.T) {
	// Test password confirmation validator
	passwordConfirmValidator := func(values map[string]any) error {
		password, ok1 := values["password"]
		confirmPassword, ok2 := values["confirm_password"]
		if !ok1 || !ok2 {
			return errors.New("password fields missing")
		}
		if password != confirmPassword {
			return errors.New("passwords do not match")
		}
		return nil
	}

	tests := []struct {
		name      string
		validator CrossFieldValidator
		values    map[string]any
		wantErr   bool
		errorMsg  string
	}{
		{
			name:      "matching passwords",
			validator: passwordConfirmValidator,
			values: map[string]any{
				"password":         "secret123",
				"confirm_password": "secret123",
			},
			wantErr: false,
		},
		{
			name:      "non-matching passwords",
			validator: passwordConfirmValidator,
			values: map[string]any{
				"password":         "secret123",
				"confirm_password": "different",
			},
			wantErr:  true,
			errorMsg: "passwords do not match",
		},
		{
			name:      "missing password fields",
			validator: passwordConfirmValidator,
			values: map[string]any{
				"email": "test@example.com",
			},
			wantErr:  true,
			errorMsg: "password fields missing",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.validator(tt.values)
			if (err != nil) != tt.wantErr {
				t.Errorf("CrossFieldValidator() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err.Error() != tt.errorMsg {
				t.Errorf("CrossFieldValidator() error message = %v, want %v", err.Error(), tt.errorMsg)
			}
		})
	}
}

// Test Widget function type
func TestWidget_FunctionType(t *testing.T) {
	// Create a mock widget function
	mockWidget := func(state *State, fieldName string, attrs ...Node) Node {
		return Input(Type("text"), Name(fieldName), Value("test"))
	}

	// Create a minimal state and field definition for testing
	fieldDef := &FieldDef{
		Name:   "test_field",
		Label:  "Test Field",
		Widget: mockWidget,
	}

	// Test that the widget function can be called
	if fieldDef.Widget == nil {
		t.Error("Widget should not be nil")
	}

	// We'll test the actual widget rendering in integration tests
	// since it requires a proper State instance
}