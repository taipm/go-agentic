package tools

import (
	"encoding/json"
	"fmt"
)

// FormatToolSuccess formats a successful tool result as JSON.
//
// Example output:
//
//	{
//	  "status": "success",
//	  "message": "User retrieved",
//	  "data": {"id": 123, "name": "John"}
//	}
func FormatToolSuccess(message string, data interface{}) string {
	result := map[string]interface{}{
		"status":  "success",
		"message": message,
	}

	if data != nil {
		result["data"] = data
	}

	return formatJSON(result)
}

// FormatToolError formats a tool error as JSON with optional hints.
//
// Parameters:
//   - toolName: Name of the tool that failed
//   - err: The error that occurred
//   - hints: Optional map of hints (e.g., {"param": "what went wrong"})
//
// Example output:
//
//	{
//	  "status": "error",
//	  "tool": "GetUser",
//	  "error": "user not found",
//	  "hints": {"user_id": "Must be a positive integer"}
//	}
func FormatToolError(toolName string, err error, hints map[string]string) string {
	result := map[string]interface{}{
		"status": "error",
		"tool":   toolName,
	}

	if err != nil {
		result["error"] = err.Error()
	} else {
		result["error"] = "unknown error"
	}

	if len(hints) > 0 {
		result["hints"] = hints
	}

	return formatJSON(result)
}

// FormatValidationError formats a validation error with the arguments that caused it.
// Useful for debugging parameter issues.
//
// Example output:
//
//	{
//	  "status": "error",
//	  "type": "validation",
//	  "tool": "GetUser",
//	  "error": "parameter \"user_id\": required parameter missing",
//	  "received": {"name": "John"}
//	}
func FormatValidationError(toolName string, err error, receivedArgs map[string]interface{}) string {
	result := map[string]interface{}{
		"status": "error",
		"type":   "validation",
		"tool":   toolName,
	}

	if err != nil {
		result["error"] = err.Error()
	} else {
		result["error"] = "validation failed"
	}

	if len(receivedArgs) > 0 {
		// Only include non-empty received args for clarity
		result["received"] = receivedArgs
	}

	return formatJSON(result)
}

// FormatJSON converts any value to a pretty-printed JSON string.
// Useful for wrapping existing result data.
func FormatJSON(v interface{}) string {
	return formatJSON(v)
}

// FormatText returns plain text result wrapped in a simple JSON structure.
//
// Example output:
//
//	{
//	  "status": "success",
//	  "message": "",
//	  "data": "plain text result here"
//	}
func FormatText(text string) string {
	return FormatToolSuccess("", text)
}

// FormatMixed returns text with additional structured data.
// Useful when you want to return both human-readable text and machine-readable data.
//
// Example output:
//
//	{
//	  "status": "success",
//	  "text": "5 results found",
//	  "data": [{"id": 1}, {"id": 2}, ...]
//	}
func FormatMixed(text string, data interface{}) string {
	result := map[string]interface{}{
		"status": "success",
		"text":   text,
		"data":   data,
	}
	return formatJSON(result)
}

// formatJSON is a helper that marshals any value to pretty-printed JSON.
// Panics should never occur in normal usage, but we use MarshalIndent
// for readability of debug output.
func formatJSON(v interface{}) string {
	jsonBytes, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		// Fallback: return error message as JSON
		fallback := map[string]interface{}{
			"status": "error",
			"error":  fmt.Sprintf("failed to format result: %v", err),
		}
		if fb, err := json.Marshal(fallback); err == nil {
			return string(fb)
		}
		// Last resort: plain string
		return fmt.Sprintf(`{"status":"error","error":"%v"}`, err)
	}
	return string(jsonBytes)
}
