package tools

import (
	"testing"
)

// TestCoerceToString tests the CoerceToString function
func TestCoerceToString(t *testing.T) {
	tests := []struct {
		name      string
		input     interface{}
		expected  string
		wantError bool
	}{
		// String inputs
		{name: "string_basic", input: "hello", expected: "hello", wantError: false},
		{name: "string_empty", input: "", expected: "", wantError: false},
		{name: "string_with_spaces", input: "  hello world  ", expected: "  hello world  ", wantError: false},

		// Float64 inputs
		{name: "float64_integer_value", input: 42.0, expected: "42", wantError: false},
		{name: "float64_decimal_value", input: 3.14, expected: "3.14", wantError: false},
		{name: "float64_zero", input: 0.0, expected: "0", wantError: false},
		{name: "float64_negative", input: -5.5, expected: "-5.5", wantError: false},

		// Int64 inputs
		{name: "int64_positive", input: int64(100), expected: "100", wantError: false},
		{name: "int64_zero", input: int64(0), expected: "0", wantError: false},
		{name: "int64_negative", input: int64(-50), expected: "-50", wantError: false},

		// Int inputs
		{name: "int_positive", input: 42, expected: "42", wantError: false},
		{name: "int_zero", input: 0, expected: "0", wantError: false},
		{name: "int_negative", input: -10, expected: "-10", wantError: false},

		// Int32 inputs
		{name: "int32_positive", input: int32(32), expected: "32", wantError: false},
		{name: "int32_zero", input: int32(0), expected: "0", wantError: false},

		// Bool inputs
		{name: "bool_true", input: true, expected: "true", wantError: false},
		{name: "bool_false", input: false, expected: "false", wantError: false},

		// Nil input
		{name: "nil_input", input: nil, expected: "", wantError: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := CoerceToString(tt.input)
			if (err != nil) != tt.wantError {
				t.Errorf("CoerceToString() error = %v, wantError %v", err, tt.wantError)
				return
			}
			if result != tt.expected {
				t.Errorf("CoerceToString() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// TestCoerceToInt tests the CoerceToInt function
func TestCoerceToInt(t *testing.T) {
	tests := []struct {
		name      string
		input     interface{}
		expected  int
		wantError bool
	}{
		// Float64 inputs
		{name: "float64_integer_value", input: 42.0, expected: 42, wantError: false},
		{name: "float64_decimal_value", input: 3.14, expected: 3, wantError: false},
		{name: "float64_zero", input: 0.0, expected: 0, wantError: false},

		// Int64 inputs
		{name: "int64_positive", input: int64(100), expected: 100, wantError: false},
		{name: "int64_zero", input: int64(0), expected: 0, wantError: false},

		// Int inputs
		{name: "int_positive", input: 42, expected: 42, wantError: false},
		{name: "int_zero", input: 0, expected: 0, wantError: false},

		// String inputs
		{name: "string_valid", input: "123", expected: 123, wantError: false},
		{name: "string_zero", input: "0", expected: 0, wantError: false},
		{name: "string_negative", input: "-50", expected: -50, wantError: false},
		{name: "string_invalid", input: "not_a_number", expected: 0, wantError: true},

		// Nil input
		{name: "nil_input", input: nil, expected: 0, wantError: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := CoerceToInt(tt.input)
			if (err != nil) != tt.wantError {
				t.Errorf("CoerceToInt() error = %v, wantError %v", err, tt.wantError)
				return
			}
			if result != tt.expected {
				t.Errorf("CoerceToInt() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// TestCoerceToBool tests the CoerceToBool function
func TestCoerceToBool(t *testing.T) {
	tests := []struct {
		name      string
		input     interface{}
		expected  bool
		wantError bool
	}{
		// Bool inputs
		{name: "bool_true", input: true, expected: true, wantError: false},
		{name: "bool_false", input: false, expected: false, wantError: false},

		// String inputs
		{name: "string_true", input: "true", expected: true, wantError: false},
		{name: "string_yes", input: "yes", expected: true, wantError: false},
		{name: "string_1", input: "1", expected: true, wantError: false},
		{name: "string_false", input: "false", expected: false, wantError: false},
		{name: "string_no", input: "no", expected: false, wantError: false},
		{name: "string_0", input: "0", expected: false, wantError: false},
		{name: "string_invalid", input: "maybe", expected: false, wantError: true},

		// Numeric inputs
		{name: "int_nonzero", input: 1, expected: true, wantError: false},
		{name: "int_zero", input: 0, expected: false, wantError: false},
		{name: "float64_nonzero", input: 1.5, expected: true, wantError: false},
		{name: "float64_zero", input: 0.0, expected: false, wantError: false},

		// Nil input
		{name: "nil_input", input: nil, expected: false, wantError: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := CoerceToBool(tt.input)
			if (err != nil) != tt.wantError {
				t.Errorf("CoerceToBool() error = %v, wantError %v", err, tt.wantError)
				return
			}
			if result != tt.expected {
				t.Errorf("CoerceToBool() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// TestCoerceToFloat tests the CoerceToFloat function
func TestCoerceToFloat(t *testing.T) {
	tests := []struct {
		name      string
		input     interface{}
		expected  float64
		wantError bool
	}{
		// Float64 inputs
		{name: "float64_basic", input: 3.14, expected: 3.14, wantError: false},
		{name: "float64_zero", input: 0.0, expected: 0.0, wantError: false},

		// Int inputs
		{name: "int_basic", input: 42, expected: 42.0, wantError: false},
		{name: "int_zero", input: 0, expected: 0.0, wantError: false},

		// String inputs
		{name: "string_float", input: "3.14", expected: 3.14, wantError: false},
		{name: "string_int", input: "42", expected: 42.0, wantError: false},
		{name: "string_invalid", input: "not_a_number", expected: 0, wantError: true},

		// Nil input
		{name: "nil_input", input: nil, expected: 0, wantError: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := CoerceToFloat(tt.input)
			if (err != nil) != tt.wantError {
				t.Errorf("CoerceToFloat() error = %v, wantError %v", err, tt.wantError)
				return
			}
			if result != tt.expected {
				t.Errorf("CoerceToFloat() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// TestMustGetString tests the MustGetString function
func TestMustGetString(t *testing.T) {
	tests := []struct {
		name      string
		args      map[string]interface{}
		key       string
		expected  string
		wantError bool
	}{
		{name: "exists_string", args: map[string]interface{}{"name": "Alice"}, key: "name", expected: "Alice", wantError: false},
		{name: "exists_float", args: map[string]interface{}{"age": 25.0}, key: "age", expected: "25", wantError: false},
		{name: "missing_required", args: map[string]interface{}{}, key: "name", expected: "", wantError: true},
		{name: "nil_value", args: map[string]interface{}{"name": nil}, key: "name", expected: "", wantError: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := MustGetString(tt.args, tt.key)
			if (err != nil) != tt.wantError {
				t.Errorf("MustGetString() error = %v, wantError %v", err, tt.wantError)
				return
			}
			if result != tt.expected {
				t.Errorf("MustGetString() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// TestMustGetInt tests the MustGetInt function
func TestMustGetInt(t *testing.T) {
	tests := []struct {
		name      string
		args      map[string]interface{}
		key       string
		expected  int
		wantError bool
	}{
		{name: "exists_float", args: map[string]interface{}{"count": 42.0}, key: "count", expected: 42, wantError: false},
		{name: "exists_int", args: map[string]interface{}{"count": 10}, key: "count", expected: 10, wantError: false},
		{name: "missing_required", args: map[string]interface{}{}, key: "count", expected: 0, wantError: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := MustGetInt(tt.args, tt.key)
			if (err != nil) != tt.wantError {
				t.Errorf("MustGetInt() error = %v, wantError %v", err, tt.wantError)
				return
			}
			if result != tt.expected {
				t.Errorf("MustGetInt() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// TestMustGetBool tests the MustGetBool function
func TestMustGetBool(t *testing.T) {
	tests := []struct {
		name      string
		args      map[string]interface{}
		key       string
		expected  bool
		wantError bool
	}{
		{name: "exists_bool", args: map[string]interface{}{"active": true}, key: "active", expected: true, wantError: false},
		{name: "exists_string_true", args: map[string]interface{}{"active": "yes"}, key: "active", expected: true, wantError: false},
		{name: "missing_required", args: map[string]interface{}{}, key: "active", expected: false, wantError: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := MustGetBool(tt.args, tt.key)
			if (err != nil) != tt.wantError {
				t.Errorf("MustGetBool() error = %v, wantError %v", err, tt.wantError)
				return
			}
			if result != tt.expected {
				t.Errorf("MustGetBool() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// TestOptionalGetString tests the OptionalGetString function
func TestOptionalGetString(t *testing.T) {
	tests := []struct {
		name      string
		args      map[string]interface{}
		key       string
		defaultVal string
		expected  string
	}{
		{name: "exists", args: map[string]interface{}{"name": "Bob"}, key: "name", defaultVal: "default", expected: "Bob"},
		{name: "missing", args: map[string]interface{}{}, key: "name", defaultVal: "default", expected: "default"},
		{name: "coercion_works", args: map[string]interface{}{"name": 123.0}, key: "name", defaultVal: "default", expected: "123"},
		{name: "coercion_fails_uses_default", args: map[string]interface{}{"name": nil}, key: "name", defaultVal: "default", expected: "default"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := OptionalGetString(tt.args, tt.key, tt.defaultVal)
			if result != tt.expected {
				t.Errorf("OptionalGetString() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// TestOptionalGetInt tests the OptionalGetInt function
func TestOptionalGetInt(t *testing.T) {
	tests := []struct {
		name      string
		args      map[string]interface{}
		key       string
		defaultVal int
		expected  int
	}{
		{name: "exists", args: map[string]interface{}{"count": 10.0}, key: "count", defaultVal: 0, expected: 10},
		{name: "missing", args: map[string]interface{}{}, key: "count", defaultVal: 99, expected: 99},
		{name: "coercion_fails_uses_default", args: map[string]interface{}{"count": nil}, key: "count", defaultVal: 5, expected: 5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := OptionalGetInt(tt.args, tt.key, tt.defaultVal)
			if result != tt.expected {
				t.Errorf("OptionalGetInt() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// TestOptionalGetBool tests the OptionalGetBool function
func TestOptionalGetBool(t *testing.T) {
	tests := []struct {
		name      string
		args      map[string]interface{}
		key       string
		defaultVal bool
		expected  bool
	}{
		{name: "exists_true", args: map[string]interface{}{"active": true}, key: "active", defaultVal: false, expected: true},
		{name: "exists_false", args: map[string]interface{}{"active": false}, key: "active", defaultVal: true, expected: false},
		{name: "missing", args: map[string]interface{}{}, key: "active", defaultVal: true, expected: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := OptionalGetBool(tt.args, tt.key, tt.defaultVal)
			if result != tt.expected {
				t.Errorf("OptionalGetBool() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// TestOptionalGetFloat tests the OptionalGetFloat function
func TestOptionalGetFloat(t *testing.T) {
	tests := []struct {
		name      string
		args      map[string]interface{}
		key       string
		defaultVal float64
		expected  float64
	}{
		{name: "exists", args: map[string]interface{}{"price": 3.14}, key: "price", defaultVal: 0.0, expected: 3.14},
		{name: "missing", args: map[string]interface{}{}, key: "price", defaultVal: 9.99, expected: 9.99},
		{name: "coercion_fails_uses_default", args: map[string]interface{}{"price": nil}, key: "price", defaultVal: 1.5, expected: 1.5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := OptionalGetFloat(tt.args, tt.key, tt.defaultVal)
			if result != tt.expected {
				t.Errorf("OptionalGetFloat() = %v, want %v", result, tt.expected)
			}
		})
	}
}
