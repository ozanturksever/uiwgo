package validators

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

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

// FieldsMatch validates that two fields have the same value
func FieldsMatch(field1, field2 string, message ...string) form.CrossFieldValidator {
	msg := fmt.Sprintf("Fields %s and %s must match", field1, field2)
	if len(message) > 0 {
		msg = message[0]
	}
	
	return func(values map[string]any) error {
		value1, exists1 := values[field1]
		value2, exists2 := values[field2]
		
		if !exists1 || !exists2 {
			return nil // Skip validation if fields don't exist
		}
		
		// Convert both values to strings for comparison
		str1 := fmt.Sprintf("%v", value1)
		str2 := fmt.Sprintf("%v", value2)
		
		if str1 != str2 {
			return errors.New(msg)
		}
		
		return nil
	}
}

// DateRange validates that a date field is within a specified range
func DateRange(startField, endField string, message ...string) form.CrossFieldValidator {
	msg := "End date must be after start date"
	if len(message) > 0 {
		msg = message[0]
	}
	
	return func(values map[string]any) error {
		startValue, startExists := values[startField]
		endValue, endExists := values[endField]
		
		if !startExists || !endExists {
			return nil // Skip validation if fields don't exist
		}
		
		// Convert to strings and parse as dates
		startStr, startOk := startValue.(string)
		endStr, endOk := endValue.(string)
		
		if !startOk || !endOk || startStr == "" || endStr == "" {
			return nil // Skip validation if not strings or empty
		}
		
		// Try parsing as RFC3339 date format (YYYY-MM-DD)
		startDate, err1 := time.Parse("2006-01-02", startStr)
		endDate, err2 := time.Parse("2006-01-02", endStr)
		
		if err1 != nil || err2 != nil {
			// Try parsing as datetime format
			startDate, err1 = time.Parse(time.RFC3339, startStr)
			endDate, err2 = time.Parse(time.RFC3339, endStr)
			
			if err1 != nil || err2 != nil {
				return nil // Skip validation if dates can't be parsed
			}
		}
		
		if !endDate.After(startDate) {
			return errors.New(msg)
		}
		
		return nil
	}
}

// NumericRange validates that a numeric field is within a specified range relative to another field
func NumericRange(minField, maxField string, message ...string) form.CrossFieldValidator {
	msg := "Maximum value must be greater than minimum value"
	if len(message) > 0 {
		msg = message[0]
	}
	
	return func(values map[string]any) error {
		minValue, minExists := values[minField]
		maxValue, maxExists := values[maxField]
		
		if !minExists || !maxExists {
			return nil // Skip validation if fields don't exist
		}
		
		// Convert to strings and parse as numbers
		minStr, minOk := minValue.(string)
		maxStr, maxOk := maxValue.(string)
		
		if !minOk || !maxOk || minStr == "" || maxStr == "" {
			return nil // Skip validation if not strings or empty
		}
		
		minNum, err1 := strconv.ParseFloat(minStr, 64)
		maxNum, err2 := strconv.ParseFloat(maxStr, 64)
		
		if err1 != nil || err2 != nil {
			return nil // Skip validation if numbers can't be parsed
		}
		
		if maxNum <= minNum {
			return errors.New(msg)
		}
		
		return nil
	}
}

// ConditionalRequired validates that a field is required when another field has a specific value
func ConditionalRequired(dependentField, triggerField, triggerValue string, message ...string) form.CrossFieldValidator {
	msg := fmt.Sprintf("This field is required when %s is %s", triggerField, triggerValue)
	if len(message) > 0 {
		msg = message[0]
	}
	
	return func(values map[string]any) error {
		dependentValue, dependentExists := values[dependentField]
		triggerVal, triggerExists := values[triggerField]
		
		if !triggerExists {
			return nil // Skip validation if trigger field doesn't exist
		}
		
		// Check if trigger field has the specified value
		triggerStr := fmt.Sprintf("%v", triggerVal)
		if triggerStr != triggerValue {
			return nil // Skip validation if trigger condition not met
		}
		
		// Now check if dependent field is required
		if !dependentExists || dependentValue == nil {
			return errors.New(msg)
		}
		
		if str, ok := dependentValue.(string); ok && strings.TrimSpace(str) == "" {
			return errors.New(msg)
		}
		
		return nil
	}
}

// AtLeastOneRequired validates that at least one of the specified fields has a value
func AtLeastOneRequired(fields []string, message ...string) form.CrossFieldValidator {
	msg := "At least one of these fields is required"
	if len(message) > 0 {
		msg = message[0]
	}
	
	return func(values map[string]any) error {
		for _, field := range fields {
			value, exists := values[field]
			if exists && value != nil {
				if str, ok := value.(string); ok {
					if strings.TrimSpace(str) != "" {
						return nil // Found a non-empty field
					}
				} else {
					return nil // Found a non-string value (considered valid)
				}
			}
		}
		
		return errors.New(msg)
	}
}

// MutuallyExclusive validates that only one of the specified fields has a value
func MutuallyExclusive(fields []string, message ...string) form.CrossFieldValidator {
	msg := "Only one of these fields can have a value"
	if len(message) > 0 {
		msg = message[0]
	}
	
	return func(values map[string]any) error {
		filledCount := 0
		
		for _, field := range fields {
			value, exists := values[field]
			if exists && value != nil {
				if str, ok := value.(string); ok {
					if strings.TrimSpace(str) != "" {
						filledCount++
					}
				} else {
					filledCount++
				}
			}
		}
		
		if filledCount > 1 {
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