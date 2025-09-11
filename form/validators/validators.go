package validators

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/ozanturksever/uiwgo/form"
)

// Required validates that a field has a non-empty value
func Required(message ...string) form.Validator {
	msg := "This field is required"
	if len(message) > 0 {
		msg = message[0]
	}
	
	return func(value any) error {
		if value == nil {
			return errors.New(msg)
		}
		
		str, ok := value.(string)
		if !ok {
			return errors.New(msg)
		}
		
		if strings.TrimSpace(str) == "" {
			return errors.New(msg)
		}
		
		return nil
	}
}

// MinLength validates that a string field has at least the specified length
func MinLength(minLen int, message ...string) form.Validator {
	msg := fmt.Sprintf("This field must be at least %d characters long", minLen)
	if len(message) > 0 {
		msg = message[0]
	}
	
	return func(value any) error {
		if value == nil {
			return nil // Allow nil values, use Required() for non-nil validation
		}
		
		str, ok := value.(string)
		if !ok {
			return nil // Skip validation for non-string values
		}
		
		if len(str) < minLen {
			return errors.New(msg)
		}
		
		return nil
	}
}

// MaxLength validates that a string field has at most the specified length
func MaxLength(maxLen int, message ...string) form.Validator {
	msg := fmt.Sprintf("This field must be at most %d characters long", maxLen)
	if len(message) > 0 {
		msg = message[0]
	}
	
	return func(value any) error {
		if value == nil {
			return nil
		}
		
		str, ok := value.(string)
		if !ok {
			return nil
		}
		
		if len(str) > maxLen {
			return errors.New(msg)
		}
		
		return nil
	}
}

// Email validates that a field contains a valid email address
func Email(message ...string) form.Validator {
	msg := "Please enter a valid email address"
	if len(message) > 0 {
		msg = message[0]
	}
	
	// Basic email regex pattern
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	
	return func(value any) error {
		if value == nil {
			return nil // Allow nil values, use Required() for non-nil validation
		}
		
		str, ok := value.(string)
		if !ok {
			return nil // Skip validation for non-string values
		}
		
		if str == "" {
			return nil // Allow empty strings, use Required() for non-empty validation
		}
		
		if !emailRegex.MatchString(str) {
			return errors.New(msg)
		}
		
		return nil
	}
}

// PasswordsMatch is a cross-field validator that ensures two password fields match
func PasswordsMatch(passwordField, confirmField string, message ...string) form.CrossFieldValidator {
	msg := "Passwords do not match"
	if len(message) > 0 {
		msg = message[0]
	}
	
	return func(values map[string]any) error {
		password, passwordExists := values[passwordField]
		confirm, confirmExists := values[confirmField]
		
		if !passwordExists || !confirmExists {
			return nil // Skip validation if fields don't exist
		}
		
		// Convert to strings for comparison
		passwordStr, passwordOk := password.(string)
		confirmStr, confirmOk := confirm.(string)
		
		if !passwordOk || !confirmOk {
			return nil // Skip validation if not strings
		}
		
		if passwordStr != confirmStr {
			return errors.New(msg)
		}
		
		return nil
	}
}