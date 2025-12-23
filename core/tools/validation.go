package tools

import (
	"fmt"
	"strconv"
)

// ParameterSchema represents a tool's parameter schema
type ParameterSchema map[string]interface{}

// ExtractRequiredFields extracts required field names from tool parameters
func ExtractRequiredFields(params ParameterSchema) []string {
	var requiredFields []string
	if required, ok := params["required"].([]interface{}); ok {
		for _, r := range required {
			if rStr, ok := r.(string); ok {
				requiredFields = append(requiredFields, rStr)
			}
		}
	}
	return requiredFields
}

// ValidateFieldType validates a single field's type against schema
func ValidateFieldType(toolName, fieldName string, fieldValue interface{}, propSchema interface{}) error {
	propMap, ok := propSchema.(map[string]interface{})
	if !ok {
		return nil // Skip if schema is not a map
	}

	expectedType, ok := propMap["type"].(string)
	if !ok {
		return nil // Skip if type not specified
	}

	switch expectedType {
	case "string":
		// Allow numeric types to be coerced to string (common with text-parsed tool calls)
		switch fieldValue.(type) {
		case string:
			// Already a string, valid
		case float64, int, int64, int32:
			// Numeric types can be coerced to string - validation passes
			// Handler should do the actual conversion
		default:
			return fmt.Errorf("tool '%s': parameter '%s' should be string, got %T", toolName, fieldName, fieldValue)
		}
	case "number", "integer":
		switch fieldValue.(type) {
		case float64, int, int64:
			// Valid number types
		case string:
			// Try to parse string as number
			s := fieldValue.(string)
			if _, err := strconv.ParseFloat(s, 64); err != nil {
				if _, err := strconv.ParseInt(s, 10, 64); err != nil {
					return fmt.Errorf("tool '%s': parameter '%s' should be number, got string '%s'", toolName, fieldName, s)
				}
			}
		default:
			return fmt.Errorf("tool '%s': parameter '%s' should be number, got %T", toolName, fieldName, fieldValue)
		}
	case "boolean":
		if _, ok := fieldValue.(bool); !ok {
			return fmt.Errorf("tool '%s': parameter '%s' should be boolean, got %T", toolName, fieldName, fieldValue)
		}
	case "array":
		if _, ok := fieldValue.([]interface{}); !ok {
			return fmt.Errorf("tool '%s': parameter '%s' should be array, got %T", toolName, fieldName, fieldValue)
		}
	case "object":
		if _, ok := fieldValue.(map[string]interface{}); !ok {
			return fmt.Errorf("tool '%s': parameter '%s' should be object, got %T", toolName, fieldName, fieldValue)
		}
	}

	return nil
}

// ValidateArguments validates tool arguments against parameter schema
func ValidateArguments(toolName string, params ParameterSchema, args map[string]interface{}) error {
	if params == nil {
		return nil // No parameters defined, so any args are acceptable
	}

	// Get parameter schema
	properties, ok := params["properties"].(map[string]interface{})
	if !ok {
		return nil // No properties defined, skip validation
	}

	// Check required fields are present
	requiredFields := ExtractRequiredFields(params)
	for _, fieldName := range requiredFields {
		if _, exists := args[fieldName]; !exists {
			return fmt.Errorf("tool '%s': required parameter '%s' is missing", toolName, fieldName)
		}
	}

	// Validate parameter types
	for argName, argValue := range args {
		if propSchema, exists := properties[argName]; exists {
			if err := ValidateFieldType(toolName, argName, argValue, propSchema); err != nil {
				return err
			}
		}
	}

	return nil
}
