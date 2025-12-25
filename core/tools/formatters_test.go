package tools

import (
	"encoding/json"
	"strings"
	"testing"
)

// TestFormatToolSuccess tests success result formatting
func TestFormatToolSuccess(t *testing.T) {
	tests := []struct {
		name    string
		message string
		data    interface{}
		check   func(t *testing.T, result string)
	}{
		{
			name:    "success_with_data",
			message: "User retrieved",
			data: map[string]interface{}{
				"id":   123,
				"name": "John",
			},
			check: func(t *testing.T, result string) {
				var obj map[string]interface{}
				if err := json.Unmarshal([]byte(result), &obj); err != nil {
					t.Fatalf("Invalid JSON: %v", err)
				}
				if obj["status"] != "success" {
					t.Errorf("Expected status=success, got %v", obj["status"])
				}
				if obj["message"] != "User retrieved" {
					t.Errorf("Expected message, got %v", obj["message"])
				}
				if data, ok := obj["data"]; !ok || data == nil {
					t.Errorf("Expected data field")
				}
			},
		},
		{
			name:    "success_without_data",
			message: "Operation completed",
			data:    nil,
			check: func(t *testing.T, result string) {
				var obj map[string]interface{}
				if err := json.Unmarshal([]byte(result), &obj); err != nil {
					t.Fatalf("Invalid JSON: %v", err)
				}
				if obj["status"] != "success" {
					t.Errorf("Expected status=success, got %v", obj["status"])
				}
				if _, ok := obj["data"]; ok {
					t.Errorf("Expected no data field when nil")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatToolSuccess(tt.message, tt.data)
			tt.check(t, result)
		})
	}
}

// TestFormatToolError tests error formatting
func TestFormatToolError(t *testing.T) {
	tests := []struct {
		name     string
		toolName string
		err      error
		hints    map[string]string
		check    func(t *testing.T, result string)
	}{
		{
			name:     "error_with_hints",
			toolName: "GetUser",
			err:      newTestError("user not found"),
			hints: map[string]string{
				"user_id": "Must be a positive integer",
			},
			check: func(t *testing.T, result string) {
				var obj map[string]interface{}
				if err := json.Unmarshal([]byte(result), &obj); err != nil {
					t.Fatalf("Invalid JSON: %v", err)
				}
				if obj["status"] != "error" {
					t.Errorf("Expected status=error, got %v", obj["status"])
				}
				if obj["tool"] != "GetUser" {
					t.Errorf("Expected tool=GetUser, got %v", obj["tool"])
				}
				if !strings.Contains(obj["error"].(string), "user not found") {
					t.Errorf("Expected error message, got %v", obj["error"])
				}
				if _, ok := obj["hints"]; !ok {
					t.Errorf("Expected hints field")
				}
			},
		},
		{
			name:     "error_without_hints",
			toolName: "DeleteUser",
			err:      newTestError("permission denied"),
			hints:    nil,
			check: func(t *testing.T, result string) {
				var obj map[string]interface{}
				if err := json.Unmarshal([]byte(result), &obj); err != nil {
					t.Fatalf("Invalid JSON: %v", err)
				}
				if obj["status"] != "error" {
					t.Errorf("Expected status=error")
				}
				if _, ok := obj["hints"]; ok {
					t.Errorf("Expected no hints when empty")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatToolError(tt.toolName, tt.err, tt.hints)
			tt.check(t, result)
		})
	}
}

// TestFormatValidationError tests validation error formatting
func TestFormatValidationError(t *testing.T) {
	args := map[string]interface{}{
		"name": "John",
		"age":  "invalid",
	}

	result := FormatValidationError("GetUser", newTestError("age must be integer"), args)

	var obj map[string]interface{}
	if err := json.Unmarshal([]byte(result), &obj); err != nil {
		t.Fatalf("Invalid JSON: %v", err)
	}

	if obj["status"] != "error" {
		t.Errorf("Expected status=error, got %v", obj["status"])
	}
	if obj["type"] != "validation" {
		t.Errorf("Expected type=validation, got %v", obj["type"])
	}
	if obj["tool"] != "GetUser" {
		t.Errorf("Expected tool=GetUser, got %v", obj["tool"])
	}
	if _, ok := obj["received"]; !ok {
		t.Errorf("Expected received field")
	}
}

// TestFormatJSON tests basic JSON formatting
func TestFormatJSON(t *testing.T) {
	data := map[string]interface{}{
		"status": "ok",
		"count":  42,
	}

	result := FormatJSON(data)

	var obj map[string]interface{}
	if err := json.Unmarshal([]byte(result), &obj); err != nil {
		t.Fatalf("Invalid JSON: %v", err)
	}

	if obj["status"] != "ok" {
		t.Errorf("Expected status=ok, got %v", obj["status"])
	}
	if obj["count"] != float64(42) {
		t.Errorf("Expected count=42, got %v", obj["count"])
	}
}

// TestFormatText tests text formatting
func TestFormatText(t *testing.T) {
	result := FormatText("plain text result")

	var obj map[string]interface{}
	if err := json.Unmarshal([]byte(result), &obj); err != nil {
		t.Fatalf("Invalid JSON: %v", err)
	}

	if obj["status"] != "success" {
		t.Errorf("Expected status=success")
	}
	if obj["data"] != "plain text result" {
		t.Errorf("Expected data to be plain text")
	}
}

// TestFormatMixed tests mixed text + data formatting
func TestFormatMixed(t *testing.T) {
	data := map[string]interface{}{
		"item1": "value1",
		"item2": "value2",
	}

	result := FormatMixed("5 results found", data)

	var obj map[string]interface{}
	if err := json.Unmarshal([]byte(result), &obj); err != nil {
		t.Fatalf("Invalid JSON: %v", err)
	}

	if obj["status"] != "success" {
		t.Errorf("Expected status=success")
	}
	if obj["text"] != "5 results found" {
		t.Errorf("Expected text field")
	}
	if _, ok := obj["data"]; !ok {
		t.Errorf("Expected data field")
	}
}

// TestFormatErrorHandling tests formatting of unmarshal errors
func TestFormatErrorHandling(t *testing.T) {
	// Test with nil data (should handle gracefully)
	result := FormatToolSuccess("test", nil)

	var obj map[string]interface{}
	if err := json.Unmarshal([]byte(result), &obj); err != nil {
		t.Fatalf("Should handle nil data: %v", err)
	}

	if obj["status"] != "success" {
		t.Errorf("Expected valid success response")
	}
}

// Helper function for creating test errors
func newTestError(msg string) error {
	return &testError{msg: msg}
}

type testError struct {
	msg string
}

func (e *testError) Error() string {
	return e.msg
}
