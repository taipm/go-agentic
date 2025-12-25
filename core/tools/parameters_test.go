package tools

import (
	"strings"
	"testing"
)

// TestParameterExtractorRequired tests required parameter extraction
func TestParameterExtractorRequired(t *testing.T) {
	tests := []struct {
		name      string
		args      map[string]interface{}
		extract   func(*ParameterExtractor) (string, int, bool)
		wantErr   bool
		errSubstr string
	}{
		{
			name: "all_required_valid",
			args: map[string]interface{}{
				"name":   "John",
				"age":    30,
				"active": true,
			},
			extract: func(pe *ParameterExtractor) (string, int, bool) {
				return pe.RequireString("name"), pe.RequireInt("age"), pe.RequireBool("active")
			},
			wantErr: false,
		},
		{
			name: "missing_required_string",
			args: map[string]interface{}{
				"age":    30,
				"active": true,
			},
			extract: func(pe *ParameterExtractor) (string, int, bool) {
				return pe.RequireString("name"), pe.RequireInt("age"), pe.RequireBool("active")
			},
			wantErr:   true,
			errSubstr: "name",
		},
		{
			name: "type_mismatch_int",
			args: map[string]interface{}{
				"name":   "John",
				"age":    "not_an_int", // string instead of int - should fail
				"active": true,
			},
			extract: func(pe *ParameterExtractor) (string, int, bool) {
				return pe.RequireString("name"), pe.RequireInt("age"), pe.RequireBool("active")
			},
			wantErr:   true,
			errSubstr: "age",
		},
		{
			name: "multiple_errors",
			args: map[string]interface{}{
				"age": "not_an_int",
			},
			extract: func(pe *ParameterExtractor) (string, int, bool) {
				return pe.RequireString("name"), pe.RequireInt("age"), pe.RequireBool("active")
			},
			wantErr:   true,
			errSubstr: "name", // Should have multiple errors including missing name
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pe := NewParameterExtractor(tt.args)
			name, age, active := tt.extract(pe)
			err := pe.Errors()

			if (err != nil) != tt.wantErr {
				t.Errorf("Errors() = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && err != nil {
				if !strings.Contains(err.Error(), tt.errSubstr) {
					t.Errorf("Error %q does not contain %q", err.Error(), tt.errSubstr)
				}
			}

			// When no error, verify values are correct
			if !tt.wantErr {
				if name != "John" {
					t.Errorf("Expected name=John, got %q", name)
				}
				if age != 30 {
					t.Errorf("Expected age=30, got %d", age)
				}
				if !active {
					t.Errorf("Expected active=true, got %v", active)
				}
			}
		})
	}
}

// TestParameterExtractorOptional tests optional parameter extraction
func TestParameterExtractorOptional(t *testing.T) {
	tests := []struct {
		name     string
		args     map[string]interface{}
		extract  func(*ParameterExtractor) (string, int)
		expected string
		expInt   int
	}{
		{
			name: "optional_provided",
			args: map[string]interface{}{
				"name": "John",
				"age":  30,
			},
			extract: func(pe *ParameterExtractor) (string, int) {
				return pe.OptionalString("name", "default"), pe.OptionalInt("age", 0)
			},
			expected: "John",
			expInt:   30,
		},
		{
			name: "optional_missing",
			args: map[string]interface{}{},
			extract: func(pe *ParameterExtractor) (string, int) {
				return pe.OptionalString("name", "default"), pe.OptionalInt("age", 25)
			},
			expected: "default",
			expInt:   25,
		},
		{
			name: "optional_coercion_failure",
			args: map[string]interface{}{
				"age": "not_an_int",
			},
			extract: func(pe *ParameterExtractor) (string, int) {
				return pe.OptionalString("name", "default"), pe.OptionalInt("age", 25)
			},
			expected: "default",
			expInt:   25, // Falls back to default
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pe := NewParameterExtractor(tt.args)
			name, age := tt.extract(pe)

			if name != tt.expected {
				t.Errorf("Expected name=%q, got %q", tt.expected, name)
			}
			if age != tt.expInt {
				t.Errorf("Expected age=%d, got %d", tt.expInt, age)
			}

			// Optional should not accumulate errors
			if err := pe.Errors(); err != nil {
				t.Errorf("Unexpected error from optional extraction: %v", err)
			}
		})
	}
}

// TestParameterExtractorWithTool tests tool context in errors
func TestParameterExtractorWithTool(t *testing.T) {
	args := map[string]interface{}{}
	pe := NewParameterExtractor(args).WithTool("GetUser")
	pe.RequireString("name")
	pe.RequireInt("id")

	err := pe.Errors()
	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	errStr := err.Error()
	if !strings.Contains(errStr, "GetUser") {
		t.Errorf("Error should contain tool name, got: %q", errStr)
	}
	if !strings.Contains(errStr, "name") || !strings.Contains(errStr, "id") {
		t.Errorf("Error should contain parameter names, got: %q", errStr)
	}
}

// TestParameterExtractorHasErrors tests HasErrors method
func TestParameterExtractorHasErrors(t *testing.T) {
	tests := []struct {
		name     string
		args     map[string]interface{}
		extract  func(*ParameterExtractor)
		wantHasErrors bool
	}{
		{
			name: "no_errors",
			args: map[string]interface{}{
				"name": "John",
			},
			extract: func(pe *ParameterExtractor) {
				pe.RequireString("name")
			},
			wantHasErrors: false,
		},
		{
			name: "with_errors",
			args: map[string]interface{}{},
			extract: func(pe *ParameterExtractor) {
				pe.RequireString("name")
			},
			wantHasErrors: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pe := NewParameterExtractor(tt.args)
			tt.extract(pe)

			if pe.HasErrors() != tt.wantHasErrors {
				t.Errorf("HasErrors() = %v, want %v", pe.HasErrors(), tt.wantHasErrors)
			}
		})
	}
}

// TestParameterExtractorFloatTypes tests float parameter extraction
func TestParameterExtractorFloatTypes(t *testing.T) {
	args := map[string]interface{}{
		"price": 19.99,
	}
	pe := NewParameterExtractor(args)
	price := pe.RequireFloat("price")

	if price != 19.99 {
		t.Errorf("Expected 19.99, got %f", price)
	}

	if pe.Errors() != nil {
		t.Errorf("Unexpected error: %v", pe.Errors())
	}
}

// TestParameterExtractorBoolCoercion tests bool coercion from various types
func TestParameterExtractorBoolCoercion(t *testing.T) {
	tests := []struct {
		name     string
		args     map[string]interface{}
		expected bool
	}{
		{
			name: "bool_true",
			args: map[string]interface{}{
				"flag": true,
			},
			expected: true,
		},
		{
			name: "bool_false",
			args: map[string]interface{}{
				"flag": false,
			},
			expected: false,
		},
		{
			name: "string_true",
			args: map[string]interface{}{
				"flag": "true",
			},
			expected: true,
		},
		{
			name: "string_false",
			args: map[string]interface{}{
				"flag": "false",
			},
			expected: false,
		},
		{
			name: "float_nonzero",
			args: map[string]interface{}{
				"flag": 1.0,
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pe := NewParameterExtractor(tt.args)
			result := pe.RequireBool("flag")

			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}

			if pe.Errors() != nil {
				t.Errorf("Unexpected error: %v", pe.Errors())
			}
		})
	}
}

// TestParameterExtractorAllValid tests AllValid helper
func TestParameterExtractorAllValid(t *testing.T) {
	tests := []struct {
		name     string
		args     map[string]interface{}
		extract  func(*ParameterExtractor)
		expected bool
	}{
		{
			name: "all_valid",
			args: map[string]interface{}{
				"name": "John",
			},
			extract: func(pe *ParameterExtractor) {
				pe.RequireString("name")
			},
			expected: true,
		},
		{
			name: "has_errors",
			args: map[string]interface{}{},
			extract: func(pe *ParameterExtractor) {
				pe.RequireString("name")
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pe := NewParameterExtractor(tt.args)
			tt.extract(pe)

			if pe.AllValid() != tt.expected {
				t.Errorf("AllValid() = %v, want %v", pe.AllValid(), tt.expected)
			}
		})
	}
}
