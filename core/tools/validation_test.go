package tools

import (
	"context"
	"testing"

	"github.com/taipm/go-agentic/core/common"
)

// TestValidateToolSchema tests the ValidateToolSchema function
func TestValidateToolSchema(t *testing.T) {
	tests := []struct {
		name      string
		tool      *common.Tool
		wantError bool
		errMsg    string
	}{
		{
			name:      "nil_tool",
			tool:      nil,
			wantError: true,
			errMsg:    "tool cannot be nil",
		},
		{
			name: "empty_name",
			tool: &common.Tool{
				Name:        "",
				Description: "Test tool",
				Func:        func(ctx context.Context, args map[string]interface{}) (string, error) { return "", nil },
			},
			wantError: true,
			errMsg:    "tool name cannot be empty",
		},
		{
			name: "empty_description",
			tool: &common.Tool{
				Name:        "TestTool",
				Description: "",
				Func:        func(ctx context.Context, args map[string]interface{}) (string, error) { return "", nil },
			},
			wantError: true,
			errMsg:    "description cannot be empty",
		},
		{
			name: "nil_handler",
			tool: &common.Tool{
				Name:        "TestTool",
				Description: "Test tool",
				Func:        nil,
			},
			wantError: true,
			errMsg:    "handler function cannot be nil",
		},
		{
			name: "valid_tool_no_parameters",
			tool: &common.Tool{
				Name:        "TestTool",
				Description: "Test tool",
				Func:        func(ctx context.Context, args map[string]interface{}) (string, error) { return "", nil },
			},
			wantError: false,
		},
		{
			name: "valid_tool_with_parameters",
			tool: &common.Tool{
				Name:        "TestTool",
				Description: "Test tool",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"name": map[string]interface{}{
							"type":        "string",
							"description": "User name",
						},
					},
					"required": []interface{}{"name"},
				},
				Func: func(ctx context.Context, args map[string]interface{}) (string, error) { return "", nil },
			},
			wantError: false,
		},
		{
			name: "invalid_parameters_type",
			tool: &common.Tool{
				Name:        "TestTool",
				Description: "Test tool",
				Parameters: map[string]interface{}{
					"type": "array",  // ← Wrong type!
				},
				Func: func(ctx context.Context, args map[string]interface{}) (string, error) { return "", nil },
			},
			wantError: true,
			errMsg:    "Parameters.type must be 'object'",
		},
		{
			name: "missing_parameters_type",
			tool: &common.Tool{
				Name:        "TestTool",
				Description: "Test tool",
				Parameters: map[string]interface{}{
					"properties": map[string]interface{}{},
					// ← Missing "type" field!
				},
				Func: func(ctx context.Context, args map[string]interface{}) (string, error) { return "", nil },
			},
			wantError: true,
			errMsg:    "Parameters.type field missing",
		},
		{
			name: "required_field_not_in_properties",
			tool: &common.Tool{
				Name:        "TestTool",
				Description: "Test tool",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"name": map[string]interface{}{
							"type": "string",
						},
					},
					"required": []interface{}{"age"},  // ← Field "age" doesn't exist in properties!
				},
				Func: func(ctx context.Context, args map[string]interface{}) (string, error) { return "", nil },
			},
			wantError: true,
			errMsg:    "required parameter \"age\" not found in properties",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateToolSchema(tt.tool)
			if (err != nil) != tt.wantError {
				t.Errorf("ValidateToolSchema() error = %v, wantError %v", err, tt.wantError)
				return
			}
			if tt.wantError && err != nil {
				if tt.errMsg != "" && !contains(err.Error(), tt.errMsg) {
					t.Errorf("ValidateToolSchema() error = %q, want %q", err.Error(), tt.errMsg)
				}
			}
		})
	}
}

// TestValidateToolCallArgs tests the ValidateToolCallArgs function
func TestValidateToolCallArgs(t *testing.T) {
	tests := []struct {
		name      string
		tool      *common.Tool
		args      map[string]interface{}
		wantError bool
		errMsg    string
	}{
		{
			name:      "nil_tool",
			tool:      nil,
			args:      map[string]interface{}{},
			wantError: false,  // Should not error
		},
		{
			name: "tool_no_schema",
			tool: &common.Tool{
				Name:        "TestTool",
				Description: "Test tool",
				Func:        func(ctx context.Context, args map[string]interface{}) (string, error) { return "", nil },
			},
			args:      map[string]interface{}{},
			wantError: false,  // Should not error
		},
		{
			name: "all_required_provided",
			tool: &common.Tool{
				Name:        "TestTool",
				Description: "Test tool",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"name": map[string]interface{}{"type": "string"},
						"age":  map[string]interface{}{"type": "integer"},
					},
					"required": []interface{}{"name", "age"},
				},
				Func: func(ctx context.Context, args map[string]interface{}) (string, error) { return "", nil },
			},
			args: map[string]interface{}{
				"name": "John",
				"age":  30,
			},
			wantError: false,
		},
		{
			name: "missing_required_parameter",
			tool: &common.Tool{
				Name:        "TestTool",
				Description: "Test tool",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"name": map[string]interface{}{"type": "string"},
						"age":  map[string]interface{}{"type": "integer"},
					},
					"required": []interface{}{"name", "age"},
				},
				Func: func(ctx context.Context, args map[string]interface{}) (string, error) { return "", nil },
			},
			args: map[string]interface{}{
				"name": "John",
				// ← Missing "age"!
			},
			wantError: true,
			errMsg:    "required parameter \"age\" is missing",
		},
		{
			name: "extra_parameters_ok",
			tool: &common.Tool{
				Name:        "TestTool",
				Description: "Test tool",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"name": map[string]interface{}{"type": "string"},
					},
					"required": []interface{}{"name"},
				},
				Func: func(ctx context.Context, args map[string]interface{}) (string, error) { return "", nil },
			},
			args: map[string]interface{}{
				"name":    "John",
				"email":   "john@example.com",  // ← Extra parameter, should be OK
				"country": "USA",               // ← Extra parameter, should be OK
			},
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateToolCallArgs(tt.tool, tt.args)
			if (err != nil) != tt.wantError {
				t.Errorf("ValidateToolCallArgs() error = %v, wantError %v", err, tt.wantError)
				return
			}
			if tt.wantError && err != nil {
				if tt.errMsg != "" && !contains(err.Error(), tt.errMsg) {
					t.Errorf("ValidateToolCallArgs() error = %q, want %q", err.Error(), tt.errMsg)
				}
			}
		})
	}
}

// TestValidateToolMap tests the ValidateToolMap function
func TestValidateToolMap(t *testing.T) {
	tests := []struct {
		name      string
		toolsMap  map[string]*common.Tool
		wantError bool
		errMsg    string
	}{
		{
			name:      "empty_map",
			toolsMap:  make(map[string]*common.Tool),
			wantError: false,
		},
		{
			name: "single_valid_tool",
			toolsMap: map[string]*common.Tool{
				"GetStatus": {
					Name:        "GetStatus",
					Description: "Get status",
					Func:        func(ctx context.Context, args map[string]interface{}) (string, error) { return "", nil },
				},
			},
			wantError: false,
		},
		{
			name: "multiple_valid_tools",
			toolsMap: map[string]*common.Tool{
				"GetStatus": {
					Name:        "GetStatus",
					Description: "Get status",
					Func:        func(ctx context.Context, args map[string]interface{}) (string, error) { return "", nil },
				},
				"SetStatus": {
					Name:        "SetStatus",
					Description: "Set status",
					Func:        func(ctx context.Context, args map[string]interface{}) (string, error) { return "", nil },
				},
			},
			wantError: false,
		},
		{
			name: "key_mismatch_tool_name",
			toolsMap: map[string]*common.Tool{
				"GetStatus": {
					Name:        "FetchStatus",  // ← Key doesn't match Name!
					Description: "Get status",
					Func:        func(ctx context.Context, args map[string]interface{}) (string, error) { return "", nil },
				},
			},
			wantError: true,
			errMsg:    "does not match tool.Name",
		},
		{
			name: "invalid_tool_in_map",
			toolsMap: map[string]*common.Tool{
				"GetStatus": {
					Name:        "GetStatus",
					Description: "Get status",
					Func:        func(ctx context.Context, args map[string]interface{}) (string, error) { return "", nil },
				},
				"InvalidTool": {
					Name:        "InvalidTool",
					Description: "",  // ← Empty description!
					Func:        func(ctx context.Context, args map[string]interface{}) (string, error) { return "", nil },
				},
			},
			wantError: true,
			errMsg:    "description cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateToolMap(tt.toolsMap)
			if (err != nil) != tt.wantError {
				t.Errorf("ValidateToolMap() error = %v, wantError %v", err, tt.wantError)
				return
			}
			if tt.wantError && err != nil {
				if tt.errMsg != "" && !contains(err.Error(), tt.errMsg) {
					t.Errorf("ValidateToolMap() error = %q, want %q", err.Error(), tt.errMsg)
				}
			}
		})
	}
}

// TestValidateToolReferences tests the ValidateToolReferences function
func TestValidateToolReferences(t *testing.T) {
	toolsMap := map[string]*common.Tool{
		"GetStatus": {
			Name:        "GetStatus",
			Description: "Get status",
			Func:        func(ctx context.Context, args map[string]interface{}) (string, error) { return "", nil },
		},
		"SetStatus": {
			Name:        "SetStatus",
			Description: "Set status",
			Func:        func(ctx context.Context, args map[string]interface{}) (string, error) { return "", nil },
		},
	}

	tests := []struct {
		name       string
		toolsMap   map[string]*common.Tool
		toolNames  []string
		wantError  bool
		errMsg     string
	}{
		{
			name:      "empty_references",
			toolsMap:  toolsMap,
			toolNames: []string{},
			wantError: false,
		},
		{
			name:      "all_references_exist",
			toolsMap:  toolsMap,
			toolNames: []string{"GetStatus", "SetStatus"},
			wantError: false,
		},
		{
			name:      "single_reference_exists",
			toolsMap:  toolsMap,
			toolNames: []string{"GetStatus"},
			wantError: false,
		},
		{
			name:      "reference_not_found",
			toolsMap:  toolsMap,
			toolNames: []string{"GetStatus", "DeleteStatus"},  // ← DeleteStatus doesn't exist!
			wantError: true,
			errMsg:    "referenced tool \"DeleteStatus\" not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateToolReferences(tt.toolsMap, tt.toolNames)
			if (err != nil) != tt.wantError {
				t.Errorf("ValidateToolReferences() error = %v, wantError %v", err, tt.wantError)
				return
			}
			if tt.wantError && err != nil {
				if tt.errMsg != "" && !contains(err.Error(), tt.errMsg) {
					t.Errorf("ValidateToolReferences() error = %q, want %q", err.Error(), tt.errMsg)
				}
			}
		})
	}
}

// Helper function to check if error message contains expected string
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && len(substr) > 0 && isContain(s, substr))
}

func isContain(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
