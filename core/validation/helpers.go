// Package validation provides configuration validation rules.
package validation

import (
	"fmt"
	"time"

	"github.com/taipm/go-agentic/core/common"
)

// ValidateDuration validates that a duration is positive
func ValidateDuration(d time.Duration, fieldName string) error {
	if d <= 0 {
		return &common.ValidationError{
			Field:   fieldName,
			Message: fmt.Sprintf("duration must be positive, got %v", d),
		}
	}
	return nil
}

// ValidateInt validates an integer is within a range
func ValidateInt(val int, min, max int, fieldName string) error {
	if val < min || val > max {
		return &common.ValidationError{
			Field:   fieldName,
			Message: fmt.Sprintf("value must be between %d and %d, got %d", min, max, val),
		}
	}
	return nil
}

// ValidateFloatRange validates a float is within a range
func ValidateFloatRange(val float64, min, max float64, fieldName string) error {
	if val < min || val > max {
		return &common.ValidationError{
			Field:   fieldName,
			Message: fmt.Sprintf("value must be between %f and %f, got %f", min, max, val),
		}
	}
	return nil
}

// Contains checks if a string slice contains a value
func Contains(arr []string, s string) bool {
	for _, v := range arr {
		if v == s {
			return true
		}
	}
	return false
}

// ValidateRequiredFields checks that required fields are not empty
func ValidateRequiredFields(fields map[string]interface{}, required []string) error {
	for _, field := range required {
		if val, exists := fields[field]; !exists || val == "" {
			return &common.ValidationError{
				Field:   field,
				Message: "required field is missing or empty",
			}
		}
	}
	return nil
}

// FormatValidationError formats a validation error message
func FormatValidationError(err error) string {
	if valErr, ok := err.(*common.ValidationError); ok {
		return fmt.Sprintf("validation error in field '%s': %s", valErr.Field, valErr.Message)
	}
	return fmt.Sprintf("validation error: %v", err)
}
