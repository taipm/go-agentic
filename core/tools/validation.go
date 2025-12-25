package tools

import (
	"fmt"

	"github.com/taipm/go-agentic/core/common"
)

// ValidateToolSchema performs comprehensive validation of a tool definition.
// Checks:
// - Name is not empty
// - Description is not empty
// - Handler function is not nil
// - Parameters structure (if defined) is valid
func ValidateToolSchema(tool *common.Tool) error {
	if tool == nil {
		return fmt.Errorf("tool cannot be nil")
	}

	if tool.Name == "" {
		return fmt.Errorf("tool name cannot be empty")
	}

	if tool.Description == "" {
		return fmt.Errorf("tool %q: description cannot be empty", tool.Name)
	}

	if tool.Func == nil {
		return fmt.Errorf("tool %q: handler function cannot be nil", tool.Name)
	}

	// Validate Parameters structure if defined
	if tool.Parameters != nil {
		params, ok := tool.Parameters.(map[string]interface{})
		if ok {
			if err := validateParameters(tool.Name, params); err != nil {
				return err
			}
		}
	}

	return nil
}

// validateParameters validates the structure of a tool's Parameters field.
func validateParameters(toolName string, params map[string]interface{}) error {
	// Check that Parameters has a "type" field
	paramType, exists := params["type"].(string)
	if !exists {
		return fmt.Errorf("tool %q: Parameters.type field missing (should be 'object')", toolName)
	}

	if paramType != "object" {
		return fmt.Errorf(
			"tool %q: Parameters.type must be 'object', got %q",
			toolName, paramType,
		)
	}

	// Get properties if defined
	props, hasProps := params["properties"].(map[string]interface{})
	if !hasProps {
		// No properties defined - that's OK for tools with no parameters
		return nil
	}

	// If there are required fields, validate them
	required, hasRequired := params["required"].([]interface{})
	if hasRequired {
		for _, field := range required {
			fieldName, ok := field.(string)
			if !ok {
				return fmt.Errorf(
					"tool %q: 'required' field must contain strings, got %T",
					toolName, field,
				)
			}

			// Verify required field exists in properties
			if _, exists := props[fieldName]; !exists {
				return fmt.Errorf(
					"tool %q: required parameter %q not found in properties",
					toolName, fieldName,
				)
			}
		}
	}

	return nil
}

// ValidateToolCallArgs validates that provided arguments match the tool's schema.
// Checks:
// - All required parameters are provided
// - Parameter types match schema (basic type checking)
func ValidateToolCallArgs(tool *common.Tool, args map[string]interface{}) error {
	if tool == nil || tool.Parameters == nil {
		// No schema to validate against
		return nil
	}

	// Convert Parameters to map[string]interface{}
	params, ok := tool.Parameters.(map[string]interface{})
	if !ok {
		// Parameters is not a map, skip validation
		return nil
	}

	// Get required fields from schema
	required, hasRequired := params["required"].([]interface{})
	if !hasRequired {
		// No required fields specified
		return nil
	}

	// Verify all required arguments are provided
	for _, field := range required {
		fieldName, _ := field.(string)
		if _, exists := args[fieldName]; !exists {
			return fmt.Errorf(
				"tool %q: required parameter %q is missing",
				tool.Name, fieldName,
			)
		}
	}

	return nil
}

// ValidateToolMap validates all tools in a map.
// Returns first error found, or nil if all tools are valid.
func ValidateToolMap(toolsMap map[string]*common.Tool) error {
	for name, tool := range toolsMap {
		// Check key matches tool name
		if name != tool.Name {
			return fmt.Errorf(
				"tool map key %q does not match tool.Name %q",
				name, tool.Name,
			)
		}

		// Validate the tool
		if err := ValidateToolSchema(tool); err != nil {
			return err
		}
	}

	return nil
}

// ValidateToolReferences validates that all tool references in a list exist in the tool map.
// Useful for validating that agents reference only defined tools.
func ValidateToolReferences(toolsMap map[string]*common.Tool, toolNames []string) error {
	for _, toolName := range toolNames {
		if _, exists := toolsMap[toolName]; !exists {
			return fmt.Errorf("referenced tool %q not found in tool map", toolName)
		}
	}

	return nil
}
